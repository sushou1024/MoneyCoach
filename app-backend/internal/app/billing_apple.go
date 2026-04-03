package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	appleVerifyReceiptURL        = "https://buy.itunes.apple.com/verifyReceipt"
	appleVerifyReceiptSandboxURL = "https://sandbox.itunes.apple.com/verifyReceipt"
)

type appleReceiptResponse struct {
	Status             int                       `json:"status"`
	LatestReceiptInfo  []appleReceiptInfo        `json:"latest_receipt_info"`
	Receipt            *appleReceipt             `json:"receipt"`
	PendingRenewalInfo []applePendingRenewalInfo `json:"pending_renewal_info"`
}

type appleReceipt struct {
	BundleID string             `json:"bundle_id"`
	InApp    []appleReceiptInfo `json:"in_app"`
}

type appleReceiptInfo struct {
	ProductID             string `json:"product_id"`
	TransactionID         string `json:"transaction_id"`
	OriginalTransactionID string `json:"original_transaction_id"`
	ExpiresDateMS         string `json:"expires_date_ms"`
	PurchaseDateMS        string `json:"purchase_date_ms"`
	WebOrderLineItemID    string `json:"web_order_line_item_id"`
	SubscriptionGroupID   string `json:"subscription_group_identifier"`
	CancellationDateMS    string `json:"cancellation_date_ms"`
	CancellationReason    string `json:"cancellation_reason"`
	IsInIntroOfferPeriod  string `json:"is_in_intro_offer_period"`
	IsTrialPeriod         string `json:"is_trial_period"`
}

type applePendingRenewalInfo struct {
	ProductID                string `json:"product_id"`
	OriginalTransactionID    string `json:"original_transaction_id"`
	IsInBillingRetryPeriod   string `json:"is_in_billing_retry_period"`
	GracePeriodExpiresDateMS string `json:"grace_period_expires_date_ms"`
	AutoRenewStatus          string `json:"auto_renew_status"`
	AutoRenewProductID       string `json:"auto_renew_product_id"`
}

type appleWebhookRequest struct {
	SignedPayload string `json:"signedPayload"`
}

type appleServerNotification struct {
	NotificationType string                      `json:"notificationType"`
	Subtype          string                      `json:"subtype"`
	Data             appleServerNotificationData `json:"data"`
}

type appleServerNotificationData struct {
	BundleID              string `json:"bundleId"`
	Environment           string `json:"environment"`
	SignedTransactionInfo string `json:"signedTransactionInfo"`
	SignedRenewalInfo     string `json:"signedRenewalInfo"`
}

type appleTransactionInfo struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	ProductID             string `json:"productId"`
	ExpiresDate           string `json:"expiresDate"`
	RevocationDate        string `json:"revocationDate"`
	RevocationReason      string `json:"revocationReason"`
	PurchaseDate          string `json:"purchaseDate"`
	BundleID              string `json:"bundleId"`
}

type appleRenewalInfo struct {
	OriginalTransactionID  string `json:"originalTransactionId"`
	ProductID              string `json:"productId"`
	IsInBillingRetryPeriod string `json:"isInBillingRetryPeriod"`
	GracePeriodExpiresDate string `json:"gracePeriodExpiresDate"`
}

func (s *Server) ensureAppleIAPConfig() error {
	missing := make([]string, 0)
	if strings.TrimSpace(s.cfg.AppStoreSharedSecret) == "" {
		missing = append(missing, "APPSTORE_SHARED_SECRET")
	}
	if strings.TrimSpace(s.cfg.AppleIapProductIDWeekly) == "" || strings.TrimSpace(s.cfg.AppleIapProductIDYearly) == "" {
		missing = append(missing, "APPLE_IAP_PRODUCT_ID_WEEKLY", "APPLE_IAP_PRODUCT_ID_YEARLY")
	}
	if strings.TrimSpace(s.cfg.AppleIOSBundleID) == "" {
		missing = append(missing, "APPLE_IOS_BUNDLE_ID")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (s *Server) verifyAppleReceipt(ctx context.Context, receiptData, productID string) (appleReceiptInfo, *applePendingRenewalInfo, error) {
	if err := s.ensureAppleIAPConfig(); err != nil {
		return appleReceiptInfo{}, nil, err
	}
	resp, err := s.verifyAppleReceiptWithURL(ctx, appleVerifyReceiptURL, receiptData)
	if err != nil {
		return appleReceiptInfo{}, nil, err
	}
	if resp.Status == 21007 {
		resp, err = s.verifyAppleReceiptWithURL(ctx, appleVerifyReceiptSandboxURL, receiptData)
		if err != nil {
			return appleReceiptInfo{}, nil, err
		}
	}
	if resp.Status != 0 {
		return appleReceiptInfo{}, nil, fmt.Errorf("apple receipt status=%d", resp.Status)
	}
	if resp.Receipt != nil {
		if bundleID := strings.TrimSpace(resp.Receipt.BundleID); bundleID != "" {
			expected := strings.TrimSpace(s.cfg.AppleIOSBundleID)
			if expected != "" && bundleID != expected {
				return appleReceiptInfo{}, nil, fmt.Errorf("bundle id mismatch: %s", bundleID)
			}
		}
	}

	transactions := resp.LatestReceiptInfo
	if len(transactions) == 0 && resp.Receipt != nil {
		transactions = resp.Receipt.InApp
	}
	receipt, ok := selectAppleReceipt(transactions, productID)
	if !ok {
		return appleReceiptInfo{}, nil, fmt.Errorf("receipt missing product_id %s", productID)
	}
	renewal := selectAppleRenewal(resp.PendingRenewalInfo, receipt)
	return receipt, renewal, nil
}

func (s *Server) verifyAppleReceiptWithURL(ctx context.Context, url string, receiptData string) (appleReceiptResponse, error) {
	payload := map[string]any{
		"receipt-data":             receiptData,
		"password":                 strings.TrimSpace(s.cfg.AppStoreSharedSecret),
		"exclude-old-transactions": true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return appleReceiptResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return appleReceiptResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return appleReceiptResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return appleReceiptResponse{}, fmt.Errorf("apple receipt verify failed status=%d", resp.StatusCode)
	}
	var parsed appleReceiptResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return appleReceiptResponse{}, err
	}
	return parsed, nil
}

func selectAppleReceipt(entries []appleReceiptInfo, productID string) (appleReceiptInfo, bool) {
	var selected appleReceiptInfo
	var selectedTime time.Time
	for _, entry := range entries {
		if productID != "" && entry.ProductID != productID {
			continue
		}
		candidate := parseAppleReceiptTime(entry)
		if candidate.After(selectedTime) {
			selected = entry
			selectedTime = candidate
		}
	}
	if selected.ProductID == "" {
		return appleReceiptInfo{}, false
	}
	return selected, true
}

func parseAppleReceiptTime(entry appleReceiptInfo) time.Time {
	if entry.ExpiresDateMS != "" {
		if parsed, err := parseMillis(entry.ExpiresDateMS); err == nil {
			return parsed
		}
	}
	if entry.PurchaseDateMS != "" {
		if parsed, err := parseMillis(entry.PurchaseDateMS); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func selectAppleRenewal(entries []applePendingRenewalInfo, receipt appleReceiptInfo) *applePendingRenewalInfo {
	for _, entry := range entries {
		if entry.ProductID != "" && receipt.ProductID != "" && entry.ProductID != receipt.ProductID {
			continue
		}
		if entry.OriginalTransactionID != "" && receipt.OriginalTransactionID != "" && entry.OriginalTransactionID != receipt.OriginalTransactionID {
			continue
		}
		renewal := entry
		return &renewal
	}
	return nil
}

func (s *Server) applyAppleReceipt(ctx context.Context, userID string, plan billingPlan, receipt appleReceiptInfo, renewal *applePendingRenewalInfo) (entitlementResponse, error) {
	expiry := parseAppleReceiptTime(receipt)
	now := time.Now().UTC()
	if expiry.IsZero() {
		expiry = now
	}
	canceled := strings.TrimSpace(receipt.CancellationDateMS) != ""
	if canceled {
		cancellation := strings.TrimSpace(receipt.CancellationDateMS)
		if parsed, err := parseMillis(cancellation); err == nil {
			expiry = parsed
		} else {
			expiry = now
		}
	}
	status := "expired"
	if expiry.After(now) {
		status = "active"
	}
	if canceled {
		status = "expired"
	}
	if renewal != nil && !canceled {
		retry := strings.TrimSpace(renewal.IsInBillingRetryPeriod)
		isRetry := retry == "1" || strings.EqualFold(retry, "true")
		if isRetry {
			if renewal.GracePeriodExpiresDateMS != "" {
				if grace, err := parseMillis(renewal.GracePeriodExpiresDateMS); err == nil && grace.After(now) {
					status = "grace"
					expiry = grace
				}
			}
		}
	}
	externalID := strings.TrimSpace(receipt.OriginalTransactionID)
	if externalID == "" {
		externalID = strings.TrimSpace(receipt.TransactionID)
	}
	if err := s.claimExternalSubscription(ctx, "apple", externalID, userID, plan.PlanID); err != nil {
		return entitlementResponse{}, err
	}

	ent := Entitlement{
		UserID:           userID,
		Status:           status,
		Provider:         "apple",
		PlanID:           plan.PlanID,
		CurrentPeriodEnd: &expiry,
		LastVerifiedAt:   &now,
	}
	if err := s.upsertEntitlement(ctx, ent); err != nil {
		return entitlementResponse{}, err
	}

	providerTxID := strings.TrimSpace(receipt.OriginalTransactionID)
	if txID := strings.TrimSpace(receipt.TransactionID); txID != "" {
		providerTxID = txID
	}
	if err := s.recordPayment(ctx, userID, "apple", providerTxID, 0, "USD", "succeeded"); err != nil {
		return entitlementResponse{}, err
	}
	return entitlementResponse{
		Status:           status,
		Provider:         "apple",
		PlanID:           plan.PlanID,
		CurrentPeriodEnd: &expiry,
	}, nil
}

func (s *Server) handleWebhookApple(w http.ResponseWriter, r *http.Request) {
	var req appleWebhookRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	payload := strings.TrimSpace(req.SignedPayload)
	if payload == "" {
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	notification, err := s.parseAppleSignedPayload(r.Context(), payload)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if notification.Data.SignedTransactionInfo == "" {
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	transaction, err := s.parseAppleTransaction(r.Context(), notification.Data.SignedTransactionInfo)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if transaction.BundleID != "" && strings.TrimSpace(s.cfg.AppleIOSBundleID) != "" {
		if transaction.BundleID != s.cfg.AppleIOSBundleID {
			s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
	}
	plan, ok := s.billingPlanByProductID(transaction.ProductID)
	if !ok {
		plan = billingPlan{PlanID: strings.TrimSpace(transaction.ProductID), AppleProductID: transaction.ProductID}
	}
	externalID := strings.TrimSpace(transaction.OriginalTransactionID)
	if externalID == "" {
		externalID = strings.TrimSpace(transaction.TransactionID)
	}
	if externalID == "" {
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	var sub ExternalSubscription
	if err := s.db.DB().WithContext(r.Context()).
		First(&sub, "provider = ? AND external_id = ?", "apple", externalID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Printf("apple webhook: subscription not found for id=%s", externalID)
			s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to resolve subscription", nil)
		return
	}
	if !ok && strings.TrimSpace(sub.PlanID) != "" {
		plan.PlanID = strings.TrimSpace(sub.PlanID)
	}
	status, expiry := appleEntitlementStatus(r.Context(), s, transaction, notification.Data.SignedRenewalInfo)
	now := time.Now().UTC()
	ent := Entitlement{
		UserID:           sub.UserID,
		Status:           status,
		Provider:         "apple",
		PlanID:           plan.PlanID,
		CurrentPeriodEnd: &expiry,
		LastVerifiedAt:   &now,
	}
	if err := s.upsertEntitlement(r.Context(), ent); err != nil {
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to update entitlement", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) parseAppleSignedPayload(ctx context.Context, payload string) (appleServerNotification, error) {
	var out appleServerNotification
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"ES256"}))
	parsed, err := parser.ParseWithClaims(payload, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing key id")
		}
		return s.applePublicKey(ctx, appleStoreKeysURL, kid)
	})
	if err != nil || parsed == nil || !parsed.Valid {
		return out, errors.New("invalid signed payload")
	}
	raw, err := json.Marshal(claims)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

func (s *Server) parseAppleTransaction(ctx context.Context, payload string) (appleTransactionInfo, error) {
	var out appleTransactionInfo
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"ES256"}))
	parsed, err := parser.ParseWithClaims(payload, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing key id")
		}
		return s.applePublicKey(ctx, appleStoreKeysURL, kid)
	})
	if err != nil || parsed == nil || !parsed.Valid {
		return out, errors.New("invalid transaction payload")
	}
	raw, err := json.Marshal(claims)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

func appleEntitlementStatus(ctx context.Context, s *Server, tx appleTransactionInfo, signedRenewal string) (string, time.Time) {
	expiry := time.Now().UTC()
	if strings.TrimSpace(tx.RevocationDate) != "" {
		if revokedAt, err := parseMillis(tx.RevocationDate); err == nil {
			return "expired", revokedAt
		}
		return "expired", expiry
	}
	if strings.TrimSpace(tx.ExpiresDate) != "" {
		if parsed, err := parseMillis(tx.ExpiresDate); err == nil {
			expiry = parsed
		}
	}
	status := "expired"
	if expiry.After(time.Now().UTC()) {
		status = "active"
	}
	if signedRenewal == "" {
		return status, expiry
	}
	renewal, err := s.parseAppleRenewalInfo(ctx, signedRenewal)
	if err != nil {
		return status, expiry
	}
	retry := strings.TrimSpace(renewal.IsInBillingRetryPeriod)
	isRetry := retry == "1" || strings.EqualFold(retry, "true")
	if isRetry && strings.TrimSpace(renewal.GracePeriodExpiresDate) != "" {
		if grace, err := parseMillis(renewal.GracePeriodExpiresDate); err == nil {
			if grace.After(time.Now().UTC()) {
				return "grace", grace
			}
		}
	}
	return status, expiry
}

func (s *Server) parseAppleRenewalInfo(ctx context.Context, payload string) (appleRenewalInfo, error) {
	var out appleRenewalInfo
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"ES256"}))
	parsed, err := parser.ParseWithClaims(payload, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing key id")
		}
		return s.applePublicKey(ctx, appleStoreKeysURL, kid)
	})
	if err != nil || parsed == nil || !parsed.Valid {
		return out, errors.New("invalid renewal payload")
	}
	raw, err := json.Marshal(claims)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}
