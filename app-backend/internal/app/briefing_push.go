package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Server) sendBriefingPushNotifications(ctx context.Context, userID string, briefings []Briefing) error {
	if len(briefings) == 0 {
		return nil
	}

	devices, err := s.loadPushDevices(ctx, userID)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return nil
	}

	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return err
	}
	prefs := parseNotificationPrefs(profile.NotificationPrefs)
	if enabled, exists := prefs["daily_briefing"]; exists && !enabled {
		return nil
	}

	var lastErr error
	for _, briefing := range briefings {
		payload := pushPayload{
			Title: briefing.Title,
			Body:  briefing.PushText,
			Data: map[string]string{
				"deep_link":    "moneycoach://briefings/" + briefing.ID,
				"briefing_id":  briefing.ID,
				"type":         "daily_briefing",
				"briefing_type": briefing.Type,
			},
		}
		if err := s.pushBriefingToDevices(ctx, devices, briefing, payload); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (s *Server) pushBriefingToDevices(ctx context.Context, devices []DeviceToken, briefing Briefing, payload pushPayload) error {
	ttl := 12 * time.Hour
	var lastErr error
	for _, device := range devices {
		if device.RevokedAt != nil || !device.PushEnabled {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(device.PushProvider)) {
		case "apns":
			if err := s.sendBriefingAPNS(ctx, device, briefing, payload, ttl); err != nil {
				lastErr = err
			}
		case "fcm":
			if err := s.sendBriefingFCM(ctx, device, briefing, payload, ttl); err != nil {
				lastErr = err
			}
		default:
			continue
		}
	}
	return lastErr
}

func (s *Server) sendBriefingAPNS(ctx context.Context, device DeviceToken, briefing Briefing, payload pushPayload, ttl time.Duration) error {
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

	body, err := json.Marshal(briefingAPNSPayload{
		APS: apnsAPS{
			Alert: apnsAlert{Title: payload.Title, Body: payload.Body},
		},
		DeepLink:    payload.Data["deep_link"],
		BriefingID:  payload.Data["briefing_id"],
		Type:        payload.Data["type"],
		BriefingType: payload.Data["briefing_type"],
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
	req.Header.Set("apns-priority", "5")
	req.Header.Set("apns-expiration", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
	req.Header.Set("apns-collapse-id", "briefing-"+briefing.BriefingDate)
	req.Header.Set("content-type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("apns briefing push failed status=%d", resp.StatusCode)
}

func (s *Server) sendBriefingFCM(ctx context.Context, device DeviceToken, briefing Briefing, payload pushPayload, ttl time.Duration) error {
	if strings.TrimSpace(s.cfg.FCMServerKey) == "" {
		return fmt.Errorf("FCM configuration missing")
	}
	body := fcmRequest{
		To:          device.DeviceToken,
		Priority:    "normal",
		TimeToLive:  int(ttl.Seconds()),
		CollapseKey: "briefing-" + briefing.BriefingDate,
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fcm briefing push failed status=%d", resp.StatusCode)
	}
	return nil
}

type briefingAPNSPayload struct {
	APS          apnsAPS `json:"aps"`
	DeepLink     string  `json:"deep_link,omitempty"`
	BriefingID   string  `json:"briefing_id,omitempty"`
	Type         string  `json:"type,omitempty"`
	BriefingType string  `json:"briefing_type,omitempty"`
}
