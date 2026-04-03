package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type billingReceiptAndroidRequest struct {
	PurchaseToken string `json:"purchase_token"`
	ProductID     string `json:"product_id"`
}

type billingReceiptIOSRequest struct {
	ReceiptData   string `json:"receipt_data"`
	ProductID     string `json:"product_id"`
	TransactionID string `json:"transaction_id"`
}

type billingDevEntitlementRequest struct {
	Status string `json:"status"`
	PlanID string `json:"plan_id"`
}

type googleWebhookRequest struct {
	Message googleWebhookMessage `json:"message"`
}

type googleWebhookMessage struct {
	Data string `json:"data"`
}

type googleWebhookPayload struct {
	PackageName              string                          `json:"packageName"`
	SubscriptionNotification *googleSubscriptionNotification `json:"subscriptionNotification"`
}

type googleSubscriptionNotification struct {
	NotificationType int    `json:"notificationType"`
	PurchaseToken    string `json:"purchaseToken"`
	SubscriptionID   string `json:"subscriptionId"`
}

func (s *Server) handleBillingPlans(w http.ResponseWriter, r *http.Request) {
	plans := s.billingPlans()
	items := make([]map[string]any, 0, len(plans))
	for _, plan := range plans {
		products := map[string]string{}
		if plan.AppleProductID != "" {
			products["apple"] = plan.AppleProductID
		}
		if plan.GoogleProductID != "" {
			products["google"] = plan.GoogleProductID
		}
		if plan.StripePriceID != "" {
			products["stripe"] = plan.StripePriceID
		}
		items = append(items, map[string]any{
			"plan_id":     plan.PlanID,
			"name":        plan.Name,
			"interval":    plan.Interval,
			"price":       plan.Price,
			"currency":    plan.Currency,
			"product_ids": products,
		})
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"plans": items})
}

func (s *Server) handleBillingEntitlement(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	entitlement := s.loadEntitlement(r.Context(), userID)
	s.writeJSON(w, http.StatusOK, entitlement)
}

func (s *Server) handleBillingDevEntitlement(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if !s.cfg.BillingDevMode {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint disabled", nil)
		return
	}
	var req billingDevEntitlementRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "expired" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "status must be active or expired", nil)
		return
	}
	planID := strings.TrimSpace(req.PlanID)
	if planID == "" {
		planID = "dev_pro"
	}

	now := time.Now().UTC()
	var periodEnd *time.Time
	if status == "active" {
		next := now.AddDate(0, 1, 0)
		periodEnd = &next
	} else {
		periodEnd = &now
	}

	var ent Entitlement
	err := s.db.DB().WithContext(r.Context()).First(&ent, "user_id = ?", userID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to load entitlement", nil)
		return
	}
	ent.UserID = userID
	ent.Status = status
	ent.Provider = "dev"
	ent.PlanID = planID
	ent.CurrentPeriodEnd = periodEnd
	ent.LastVerifiedAt = &now
	if err := s.db.DB().WithContext(r.Context()).Save(&ent).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to update entitlement", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, s.loadEntitlement(r.Context(), userID))
}

func (s *Server) handleBillingDevEntitlementDelete(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if !s.cfg.BillingDevMode {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint disabled", nil)
		return
	}
	now := time.Now().UTC()
	ent := Entitlement{
		UserID:           userID,
		Status:           "expired",
		Provider:         "dev",
		PlanID:           "dev_pro",
		CurrentPeriodEnd: &now,
		LastVerifiedAt:   &now,
	}
	if err := s.db.DB().WithContext(r.Context()).Save(&ent).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to update entitlement", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, s.loadEntitlement(r.Context(), userID))
}

func (s *Server) handleBillingReceiptAndroid(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	var req billingReceiptAndroidRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if strings.TrimSpace(req.PurchaseToken) == "" || strings.TrimSpace(req.ProductID) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "purchase_token and product_id required", nil)
		return
	}
	plan, ok := s.billingPlanByProductID(req.ProductID)
	if !ok {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unknown product_id", nil)
		return
	}
	productID := strings.TrimSpace(plan.GoogleVerifyID)
	if productID == "" {
		productID = strings.TrimSpace(plan.GoogleProductID)
	}
	if productID == "" {
		productID = req.ProductID
	}

	purchase, err := s.verifyGoogleSubscription(r.Context(), productID, req.PurchaseToken)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "BILLING_ERROR", err.Error(), nil)
		return
	}
	entitlement, err := s.applyGooglePurchase(r.Context(), userID, plan, req.PurchaseToken, purchase)
	if err != nil {
		if errors.Is(err, errSubscriptionClaimed) || errors.Is(err, errPaymentConflict) {
			s.writeError(w, http.StatusConflict, "BILLING_CONFLICT", "subscription already in use", nil)
			return
		}
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to update entitlement", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, entitlement)
}

func (s *Server) handleBillingReceiptIOS(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	var req billingReceiptIOSRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	req.ReceiptData = strings.TrimSpace(req.ReceiptData)
	req.ProductID = strings.TrimSpace(req.ProductID)
	req.TransactionID = strings.TrimSpace(req.TransactionID)
	if req.ReceiptData == "" || req.ProductID == "" || req.TransactionID == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "receipt_data, product_id, and transaction_id are required", nil)
		return
	}
	plan, ok := s.billingPlanByProductID(req.ProductID)
	if !ok {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unknown product_id", nil)
		return
	}
	receipt, renewal, err := s.verifyAppleReceipt(r.Context(), req.ReceiptData, req.ProductID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "BILLING_ERROR", err.Error(), nil)
		return
	}
	if req.TransactionID != "" {
		match := req.TransactionID == receipt.TransactionID || req.TransactionID == receipt.OriginalTransactionID
		if !match {
			s.writeError(w, http.StatusBadRequest, "BILLING_ERROR", "transaction_id does not match receipt", nil)
			return
		}
	}
	entitlement, err := s.applyAppleReceipt(r.Context(), userID, plan, receipt, renewal)
	if err != nil {
		if errors.Is(err, errSubscriptionClaimed) || errors.Is(err, errPaymentConflict) {
			s.writeError(w, http.StatusConflict, "BILLING_CONFLICT", "subscription already in use", nil)
			return
		}
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to update entitlement", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, entitlement)
}

func (s *Server) handleWebhookGoogle(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(s.cfg.GooglePubSubAudience) == "" {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint disabled", nil)
		return
	}
	if err := s.verifyGoogleWebhookAuth(r.Context(), r.Header.Get("Authorization")); err != nil {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid webhook token", nil)
		return
	}
	var req googleWebhookRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if strings.TrimSpace(req.Message.Data) == "" {
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(req.Message.Data)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid webhook payload", nil)
		return
	}
	var payload googleWebhookPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid webhook payload", nil)
		return
	}
	if payload.SubscriptionNotification == nil {
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	if s.cfg.GooglePlayPackageName != "" && payload.PackageName != "" && payload.PackageName != s.cfg.GooglePlayPackageName {
		s.logger.Printf("google webhook package mismatch: %s", payload.PackageName)
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	productID := strings.TrimSpace(payload.SubscriptionNotification.SubscriptionID)
	purchaseToken := strings.TrimSpace(payload.SubscriptionNotification.PurchaseToken)
	if productID == "" || purchaseToken == "" {
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	plan, ok := s.billingPlanByProductID(productID)
	if !ok {
		plan = billingPlan{PlanID: productID, GoogleProductID: productID}
	}
	var sub ExternalSubscription
	if err := s.db.DB().WithContext(r.Context()).
		First(&sub, "provider = ? AND external_id = ?", "google", purchaseToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Printf("google webhook: subscription not found for token=%s", purchaseToken)
			s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to resolve subscription", nil)
		return
	}
	if !ok && strings.TrimSpace(sub.PlanID) != "" {
		plan.PlanID = strings.TrimSpace(sub.PlanID)
	}

	productID = strings.TrimSpace(plan.GoogleVerifyID)
	if productID == "" {
		productID = strings.TrimSpace(plan.GoogleProductID)
	}
	if productID == "" {
		productID = strings.TrimSpace(payload.SubscriptionNotification.SubscriptionID)
	}
	purchase, err := s.verifyGoogleSubscription(r.Context(), productID, purchaseToken)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", err.Error(), nil)
		return
	}
	if _, err := s.applyGooglePurchase(r.Context(), sub.UserID, plan, purchaseToken, purchase); err != nil {
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to update entitlement", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) verifyGoogleWebhookAuth(ctx context.Context, header string) error {
	header = strings.TrimSpace(header)
	if header == "" {
		return errors.New("missing authorization")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return errors.New("invalid authorization")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return errors.New("invalid authorization")
	}
	if _, err := idtoken.Validate(ctx, token, strings.TrimSpace(s.cfg.GooglePubSubAudience)); err != nil {
		return err
	}
	return nil
}
