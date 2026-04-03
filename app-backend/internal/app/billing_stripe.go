package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/subscription"
	"github.com/stripe/stripe-go/v84/webhook"
	"gorm.io/gorm"
)

type billingStripeSessionRequest struct {
	PlanID     string `json:"plan_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

func (s *Server) handleBillingStripeSession(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if err := s.ensureStripeConfig(); err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		return
	}
	var req billingStripeSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	req.PlanID = strings.TrimSpace(req.PlanID)
	req.SuccessURL = strings.TrimSpace(req.SuccessURL)
	req.CancelURL = strings.TrimSpace(req.CancelURL)
	if req.PlanID == "" || req.SuccessURL == "" || req.CancelURL == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "plan_id, success_url, and cancel_url are required", nil)
		return
	}
	if !isSafeRedirectURL(req.SuccessURL) || !isSafeRedirectURL(req.CancelURL) {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid redirect url", nil)
		return
	}
	plan, ok := s.billingPlanByProductID(req.PlanID)
	if !ok {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unknown plan_id", nil)
		return
	}
	if strings.TrimSpace(plan.StripePriceID) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "stripe price not configured", nil)
		return
	}

	stripe.Key = strings.TrimSpace(s.cfg.StripeSecretKey)
	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:        stripe.String(req.SuccessURL),
		CancelURL:         stripe.String(req.CancelURL),
		ClientReferenceID: stripe.String(userID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(plan.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"user_id": userID,
			"plan_id": plan.PlanID,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id": userID,
				"plan_id": plan.PlanID,
			},
		},
	}
	if idem := strings.TrimSpace(r.Header.Get("Idempotency-Key")); idem != "" {
		params.Params.IdempotencyKey = stripe.String(idem)
	}
	if email := s.loadUserEmail(r.Context(), userID); email != "" {
		params.CustomerEmail = stripe.String(email)
	}
	sess, err := session.New(params)
	if err != nil {
		s.logger.Printf("stripe session error: %v", err)
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to create stripe session", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"checkout_url": sess.URL})
}

func (s *Server) handleWebhookStripe(w http.ResponseWriter, r *http.Request) {
	if err := s.ensureStripeWebhookConfig(); err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to read payload", nil)
		return
	}
	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, strings.TrimSpace(s.cfg.StripeWebhookSecret))
	if err != nil {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		sigPresent := sigHeader != ""
		s.logger.Printf(
			"stripe webhook signature error: %v request_id=%s sig_present=%t payload_bytes=%d",
			err,
			requestID,
			sigPresent,
			len(payload),
		)
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid webhook signature", nil)
		return
	}

	var handleErr error
	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			handleErr = err
			break
		}
		handleErr = s.processStripeCheckoutSession(r.Context(), &sess)
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			handleErr = err
			break
		}
		handleErr = s.syncStripeSubscription(r.Context(), &sub, "", "")
	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			handleErr = err
			break
		}
		handleErr = s.recordStripeInvoicePayment(r.Context(), &invoice)
	default:
		s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	if handleErr != nil {
		s.logger.Printf("stripe webhook error type=%s: %v", event.Type, handleErr)
		s.writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "failed to process webhook", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) processStripeCheckoutSession(ctx context.Context, sess *stripe.CheckoutSession) error {
	userID := strings.TrimSpace(sess.ClientReferenceID)
	if userID == "" {
		userID = strings.TrimSpace(sess.Metadata["user_id"])
	}
	planID := strings.TrimSpace(sess.Metadata["plan_id"])
	sub := sess.Subscription
	var subID string
	if sub != nil {
		subID = sub.ID
	}
	var fetched *stripe.Subscription
	var err error
	if sub == nil && subID != "" {
		stripe.Key = strings.TrimSpace(s.cfg.StripeSecretKey)
		fetched, err = subscription.Get(subID, nil)
		if err != nil {
			return err
		}
		sub = fetched
	}
	if sub != nil && sub.Status == "" && sub.ID != "" {
		stripe.Key = strings.TrimSpace(s.cfg.StripeSecretKey)
		fetched, err = subscription.Get(subID, nil)
		if err != nil {
			return err
		}
		sub = fetched
	}
	if err := s.syncStripeSubscription(ctx, sub, userID, planID); err != nil {
		return err
	}
	if userID != "" && sess.AmountTotal > 0 {
		amount := float64(sess.AmountTotal) / 100.0
		currency := strings.ToUpper(string(sess.Currency))
		if currency == "" {
			currency = "USD"
		}
		if err := s.recordPayment(ctx, userID, "stripe", sess.ID, amount, currency, "succeeded"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) recordStripeInvoicePayment(ctx context.Context, invoice *stripe.Invoice) error {
	subID := invoiceSubscriptionID(invoice)
	if subID == "" {
		return nil
	}
	stripe.Key = strings.TrimSpace(s.cfg.StripeSecretKey)
	sub, err := subscription.Get(subID, nil)
	if err != nil {
		return err
	}
	userID := strings.TrimSpace(sub.Metadata["user_id"])
	if userID == "" {
		return nil
	}
	amount := float64(invoice.AmountPaid) / 100.0
	currency := strings.ToUpper(string(invoice.Currency))
	if currency == "" {
		currency = "USD"
	}
	providerTxID := strings.TrimSpace(invoice.ID)
	if providerTxID == "" {
		providerTxID = subID
	}
	return s.recordPayment(ctx, userID, "stripe", providerTxID, amount, currency, "succeeded")
}

func invoiceSubscriptionID(invoice *stripe.Invoice) string {
	if invoice == nil {
		return ""
	}
	if invoice.Parent != nil && invoice.Parent.SubscriptionDetails != nil && invoice.Parent.SubscriptionDetails.Subscription != nil {
		return strings.TrimSpace(invoice.Parent.SubscriptionDetails.Subscription.ID)
	}
	if invoice.Lines == nil {
		return ""
	}
	for _, line := range invoice.Lines.Data {
		if line == nil || line.Subscription == nil {
			continue
		}
		if subID := strings.TrimSpace(line.Subscription.ID); subID != "" {
			return subID
		}
	}
	return ""
}

func (s *Server) syncStripeSubscription(ctx context.Context, sub *stripe.Subscription, fallbackUserID, fallbackPlanID string) error {
	if sub == nil {
		return fmt.Errorf("missing subscription payload")
	}
	userID := strings.TrimSpace(sub.Metadata["user_id"])
	if userID == "" {
		userID = strings.TrimSpace(fallbackUserID)
	}
	if userID == "" {
		return fmt.Errorf("subscription missing user_id")
	}
	planID := strings.TrimSpace(sub.Metadata["plan_id"])
	if planID == "" {
		planID = strings.TrimSpace(fallbackPlanID)
	}
	if planID == "" && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		if plan, ok := s.billingPlanByProductID(sub.Items.Data[0].Price.ID); ok {
			planID = plan.PlanID
		} else {
			planID = sub.Items.Data[0].Price.ID
		}
	}
	now := time.Now().UTC()
	periodEnd := now
	maxEnd := int64(0)
	if sub.Items != nil {
		for _, item := range sub.Items.Data {
			if item == nil {
				continue
			}
			if item.CurrentPeriodEnd > maxEnd {
				maxEnd = item.CurrentPeriodEnd
			}
		}
	}
	if maxEnd > 0 {
		periodEnd = time.Unix(maxEnd, 0).UTC()
	}
	ent := Entitlement{
		UserID:           userID,
		Status:           stripeEntitlementStatus(sub.Status),
		Provider:         "stripe",
		PlanID:           planID,
		CurrentPeriodEnd: &periodEnd,
		LastVerifiedAt:   &now,
	}
	return s.upsertEntitlement(ctx, ent)
}

func (s *Server) ensureStripeConfig() error {
	if strings.TrimSpace(s.cfg.StripeSecretKey) == "" {
		return fmt.Errorf("endpoint disabled")
	}
	return nil
}

func (s *Server) ensureStripeWebhookConfig() error {
	if strings.TrimSpace(s.cfg.StripeWebhookSecret) == "" || strings.TrimSpace(s.cfg.StripeSecretKey) == "" {
		return fmt.Errorf("endpoint disabled")
	}
	return nil
}

func stripeEntitlementStatus(status stripe.SubscriptionStatus) string {
	switch status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return "active"
	case stripe.SubscriptionStatusPastDue, stripe.SubscriptionStatusUnpaid:
		return "grace"
	default:
		return "expired"
	}
}

func isSafeRedirectURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "moneycoach", "exp", "exp+moneycoach":
		return true
	default:
		return false
	}
}

func (s *Server) loadUserEmail(ctx context.Context, userID string) string {
	var user User
	if err := s.db.DB().WithContext(ctx).Select("email").First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ""
		}
		return ""
	}
	if user.Email == nil {
		return ""
	}
	return strings.TrimSpace(*user.Email)
}
