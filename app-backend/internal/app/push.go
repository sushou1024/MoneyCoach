package app

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	pushDedupePrefix     = "push:sent"
	pushThrottlePrefix   = "push:throttle"
	pushDailyCountPrefix = "push:count"
)

func (s *Server) sendInsightPushNotifications(ctx context.Context, userID string, insights []Insight) error {
	if len(insights) == 0 {
		return nil
	}
	if !s.hasActiveEntitlement(ctx, userID) {
		return nil
	}
	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return err
	}
	prefs := parseNotificationPrefs(profile.NotificationPrefs)
	devices, err := s.loadPushDevices(ctx, userID)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return nil
	}

	calculationID, err := s.loadLatestPaidCalculationID(ctx, userID)
	if err != nil {
		s.logger.Printf("insights push calc lookup error user=%s err=%v", userID, err)
	}

	now := time.Now().UTC()
	for _, insight := range insights {
		if !notificationEnabled(prefs, insight.Type) {
			continue
		}
		ttl, ok, err := s.allowInsightPush(ctx, userID, insight, now)
		if err != nil {
			s.logger.Printf("insights push gating error user=%s insight=%s err=%v", userID, insight.ID, err)
			continue
		}
		if !ok {
			continue
		}
		title, body := insightPushText(insight)
		payload := buildInsightPushPayload(insight, title, body, calculationID)
		if err := s.pushToDevices(ctx, devices, insight, payload, ttl); err != nil {
			s.logger.Printf("insights push send error user=%s insight=%s err=%v", userID, insight.ID, err)
		}
	}
	return nil
}

func (s *Server) loadPushDevices(ctx context.Context, userID string) ([]DeviceToken, error) {
	var tokens []DeviceToken
	if err := s.db.DB().WithContext(ctx).
		Where("user_id = ? AND push_enabled = ? AND revoked_at IS NULL", userID, true).
		Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func notificationEnabled(prefs map[string]bool, insightType string) bool {
	key := ""
	switch insightType {
	case insightTypePortfolioWatch:
		key = "portfolio_alerts"
	case insightTypeMarketAlpha:
		key = "market_alpha"
	case insightTypeActionAlert:
		key = "action_alerts"
	}
	if key == "" {
		return false
	}
	return prefs[key]
}

func (s *Server) allowInsightPush(ctx context.Context, userID string, insight Insight, now time.Time) (time.Duration, bool, error) {
	ttl := insight.ExpiresAt.Sub(now)
	if ttl <= 0 {
		return 0, false, nil
	}
	if ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}
	dedupeKey := fmt.Sprintf("%s:%s:%s", pushDedupePrefix, userID, insightDedupeKey(insight))
	ok, err := s.redis.setNX(ctx, dedupeKey, "1", ttl)
	if err != nil {
		return 0, false, err
	}
	if !ok {
		return 0, false, nil
	}

	throttleKey, throttleTTL := insightThrottleKey(userID, insight)
	if throttleKey != "" {
		throttled, err := s.redis.setNX(ctx, throttleKey, "1", throttleTTL)
		if err != nil {
			_ = s.redis.del(ctx, dedupeKey)
			return 0, false, err
		}
		if !throttled {
			_ = s.redis.del(ctx, dedupeKey)
			return 0, false, nil
		}
	}

	if limit := insightDailyLimit(insight.Type); limit > 0 {
		dayKey := fmt.Sprintf("%s:%s:%s:%s", pushDailyCountPrefix, userID, insight.Type, now.Format("20060102"))
		count, err := s.redis.incrRate(ctx, dayKey, 24*time.Hour)
		if err != nil {
			_ = s.redis.del(ctx, dedupeKey)
			if throttleKey != "" {
				_ = s.redis.del(ctx, throttleKey)
			}
			return 0, false, err
		}
		if count > int64(limit) {
			_ = s.redis.del(ctx, dedupeKey)
			if throttleKey != "" {
				_ = s.redis.del(ctx, throttleKey)
			}
			return 0, false, nil
		}
	}

	return ttl, true, nil
}

func insightDedupeKey(insight Insight) string {
	if strings.TrimSpace(insight.TriggerKey) != "" {
		return insight.TriggerKey
	}
	return insight.ID
}

func insightThrottleKey(userID string, insight Insight) (string, time.Duration) {
	switch insight.Type {
	case insightTypeMarketAlpha:
		return fmt.Sprintf("%s:%s:market_alpha", pushThrottlePrefix, userID), 12 * time.Hour
	case insightTypePortfolioWatch:
		return fmt.Sprintf("%s:%s:portfolio_watch", pushThrottlePrefix, userID), 6 * time.Hour
	case insightTypeActionAlert:
		planID := strings.TrimSpace(derefString(insight.PlanID))
		if planID == "" {
			return "", 0
		}
		return fmt.Sprintf("%s:%s:plan:%s", pushThrottlePrefix, userID, planID), 24 * time.Hour
	default:
		return "", 0
	}
}

func insightDailyLimit(insightType string) int {
	switch insightType {
	case insightTypeMarketAlpha, insightTypePortfolioWatch:
		return 3
	default:
		return 0
	}
}

func insightPushText(insight Insight) (string, string) {
	title := "Money Coach"
	switch insight.Type {
	case insightTypePortfolioWatch:
		title = "Portfolio Alert"
	case insightTypeActionAlert:
		title = "Action Alert"
	case insightTypeMarketAlpha:
		title = "Market Alpha"
	}
	body := strings.TrimSpace(insight.TriggerReason)
	if body == "" {
		body = strings.TrimSpace(derefString(insight.SuggestedAction))
	}
	if body == "" {
		body = "New insight available."
	}
	return title, body
}

type pushPayload struct {
	Title string
	Body  string
	Data  map[string]string
}

func buildInsightPushPayload(insight Insight, title, body, calculationID string) pushPayload {
	data := map[string]string{
		"deep_link":  "moneycoach://insights/" + insight.ID,
		"insight_id": insight.ID,
		"type":       insight.Type,
		"asset":      insight.Asset,
	}
	if strings.TrimSpace(calculationID) != "" {
		data["calculation_id"] = calculationID
	}
	if value := strings.TrimSpace(derefString(insight.StrategyID)); value != "" {
		data["strategy_id"] = value
	}
	if value := strings.TrimSpace(derefString(insight.PlanID)); value != "" {
		data["plan_id"] = value
	}
	return pushPayload{Title: title, Body: body, Data: data}
}

func (s *Server) pushToDevices(ctx context.Context, devices []DeviceToken, insight Insight, payload pushPayload, ttl time.Duration) error {
	var lastErr error
	for _, device := range devices {
		if device.RevokedAt != nil || !device.PushEnabled {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(device.PushProvider)) {
		case "apns":
			if err := s.sendAPNS(ctx, device, insight, payload, ttl); err != nil {
				lastErr = err
			}
		case "fcm":
			if err := s.sendFCM(ctx, device, insight, payload, ttl); err != nil {
				lastErr = err
			}
		default:
			continue
		}
	}
	return lastErr
}

func (s *Server) sendAPNS(ctx context.Context, device DeviceToken, insight Insight, payload pushPayload, ttl time.Duration) error {
	if strings.TrimSpace(s.cfg.APNSPrivateKey) == "" || strings.TrimSpace(s.cfg.APNSKeyID) == "" || strings.TrimSpace(s.cfg.APNSTeamID) == "" || strings.TrimSpace(s.cfg.APNSBundleID) == "" {
		return fmt.Errorf("APNS configuration missing")
	}
	jwtToken, err := s.apnsJWT()
	if err != nil {
		return err
	}
	endpoint := "https://api.push.apple.com/3/device/" + device.DeviceToken
	if strings.EqualFold(device.Environment, "sandbox") {
		endpoint = "https://api.sandbox.push.apple.com/3/device/" + device.DeviceToken
	}

	body, err := json.Marshal(apnsPayload{
		APS: apnsAPS{
			Alert: apnsAlert{Title: payload.Title, Body: payload.Body},
		},
		DeepLink:      payload.Data["deep_link"],
		InsightID:     payload.Data["insight_id"],
		Type:          payload.Data["type"],
		StrategyID:    payload.Data["strategy_id"],
		Asset:         payload.Data["asset"],
		CalculationID: payload.Data["calculation_id"],
		PlanID:        payload.Data["plan_id"],
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "bearer "+jwtToken)
	req.Header.Set("apns-topic", s.cfg.APNSBundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", strconv.Itoa(apnsPriority(insight.Type)))
	req.Header.Set("apns-expiration", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
	req.Header.Set("apns-collapse-id", insightDedupeKey(insight))
	req.Header.Set("content-type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payloadBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	reason := parseAPNSError(payloadBytes)
	if resp.StatusCode == http.StatusGone || isBadAPNSToken(reason) {
		s.revokeDeviceToken(ctx, device.ID)
	}
	return fmt.Errorf("apns push failed status=%d reason=%s", resp.StatusCode, reason)
}

func (s *Server) sendFCM(ctx context.Context, device DeviceToken, insight Insight, payload pushPayload, ttl time.Duration) error {
	if strings.TrimSpace(s.cfg.FCMServerKey) == "" {
		return fmt.Errorf("FCM configuration missing")
	}
	priority := "normal"
	if insight.Type == insightTypePortfolioWatch || insight.Type == insightTypeActionAlert {
		priority = "high"
	}
	body := fcmRequest{
		To:          device.DeviceToken,
		Priority:    priority,
		TimeToLive:  int(ttl.Seconds()),
		CollapseKey: insightDedupeKey(insight),
		Notification: map[string]string{
			"title": payload.Title,
			"body":  payload.Body,
		},
		Data: payload.Data,
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "key="+s.cfg.FCMServerKey)
	req.Header.Set("content-type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payloadBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fcm push failed status=%d body=%s", resp.StatusCode, string(payloadBytes))
	}
	var result fcmResponse
	if err := json.Unmarshal(payloadBytes, &result); err != nil {
		return nil
	}
	if len(result.Results) > 0 {
		if isBadFCMToken(result.Results[0].Error) {
			s.revokeDeviceToken(ctx, device.ID)
		}
	}
	return nil
}

func (s *Server) revokeDeviceToken(ctx context.Context, deviceID string) {
	if strings.TrimSpace(deviceID) == "" {
		return
	}
	now := time.Now().UTC()
	if err := s.db.DB().WithContext(ctx).Model(&DeviceToken{}).Where("id = ?", deviceID).
		Updates(map[string]any{"revoked_at": &now, "updated_at": now}).Error; err != nil {
		s.logger.Printf("device revoke error id=%s err=%v", deviceID, err)
	}
}

func (s *Server) apnsJWT() (string, error) {
	key, err := s.apnsSigningKey()
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": s.cfg.APNSTeamID,
		"iat": time.Now().Unix(),
	})
	token.Header["kid"] = s.cfg.APNSKeyID
	return token.SignedString(key)
}

func (s *Server) apnsSigningKey() (*ecdsa.PrivateKey, error) {
	s.apnsKeyOnce.Do(func() {
		s.apnsKey, s.apnsKeyErr = parseAPNSPrivateKey(s.cfg.APNSPrivateKey)
	})
	return s.apnsKey, s.apnsKeyErr
}

func parseAPNSPrivateKey(raw string) (*ecdsa.PrivateKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("APNS_PRIVATE_KEY missing")
	}
	if !strings.Contains(raw, "BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid APNS_PRIVATE_KEY encoding")
		}
		raw = string(decoded)
	}
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("invalid APNS_PRIVATE_KEY PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("APNS key must be ECDSA")
	}
	return key, nil
}

type apnsPayload struct {
	APS           apnsAPS `json:"aps"`
	DeepLink      string  `json:"deep_link,omitempty"`
	InsightID     string  `json:"insight_id,omitempty"`
	Type          string  `json:"type,omitempty"`
	StrategyID    string  `json:"strategy_id,omitempty"`
	Asset         string  `json:"asset,omitempty"`
	CalculationID string  `json:"calculation_id,omitempty"`
	PlanID        string  `json:"plan_id,omitempty"`
}

type apnsAPS struct {
	Alert apnsAlert `json:"alert"`
}

type apnsAlert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmRequest struct {
	To           string            `json:"to"`
	Priority     string            `json:"priority,omitempty"`
	TimeToLive   int               `json:"time_to_live,omitempty"`
	CollapseKey  string            `json:"collapse_key,omitempty"`
	Notification map[string]string `json:"notification,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
}

type fcmResponse struct {
	Results []fcmResult `json:"results"`
}

type fcmResult struct {
	Error string `json:"error"`
}

func apnsPriority(insightType string) int {
	switch insightType {
	case insightTypePortfolioWatch, insightTypeActionAlert:
		return 10
	default:
		return 5
	}
}

func parseAPNSError(body []byte) string {
	type apnsError struct {
		Reason string `json:"reason"`
	}
	var parsed apnsError
	if err := json.Unmarshal(body, &parsed); err == nil {
		return parsed.Reason
	}
	return string(body)
}

func isBadAPNSToken(reason string) bool {
	switch reason {
	case "BadDeviceToken", "Unregistered", "DeviceTokenNotForTopic":
		return true
	default:
		return false
	}
}

func isBadFCMToken(reason string) bool {
	switch reason {
	case "NotRegistered", "InvalidRegistration":
		return true
	default:
		return false
	}
}
