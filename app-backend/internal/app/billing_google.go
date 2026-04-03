package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const googlePlayScope = "https://www.googleapis.com/auth/androidpublisher"

type googleSubscriptionPurchase struct {
	ExpiryTimeMillis  string      `json:"expiryTimeMillis"`
	PaymentState      int         `json:"paymentState"`
	OrderID           string      `json:"orderId"`
	PriceAmountMicros json.Number `json:"priceAmountMicros"`
	PriceCurrencyCode string      `json:"priceCurrencyCode"`
}

func (s *Server) verifyGoogleSubscription(ctx context.Context, productID, purchaseToken string) (googleSubscriptionPurchase, error) {
	if strings.TrimSpace(productID) == "" {
		return googleSubscriptionPurchase{}, fmt.Errorf("product_id required for verification")
	}
	if strings.TrimSpace(purchaseToken) == "" {
		return googleSubscriptionPurchase{}, fmt.Errorf("purchase_token required for verification")
	}
	if err := s.ensureGooglePlayConfig(); err != nil {
		return googleSubscriptionPurchase{}, err
	}
	client, err := s.googlePlayClient(ctx)
	if err != nil {
		return googleSubscriptionPurchase{}, err
	}
	url := fmt.Sprintf("https://androidpublisher.googleapis.com/androidpublisher/v3/applications/%s/purchases/subscriptions/%s/tokens/%s",
		s.cfg.GooglePlayPackageName, productID, purchaseToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return googleSubscriptionPurchase{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return googleSubscriptionPurchase{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return googleSubscriptionPurchase{}, fmt.Errorf("google verify failed status=%d body=%s", resp.StatusCode, string(body))
	}
	var purchase googleSubscriptionPurchase
	if err := json.NewDecoder(resp.Body).Decode(&purchase); err != nil {
		return googleSubscriptionPurchase{}, err
	}
	return purchase, nil
}

func (s *Server) applyGooglePurchase(ctx context.Context, userID string, plan billingPlan, purchaseToken string, purchase googleSubscriptionPurchase) (entitlementResponse, error) {
	if err := s.claimExternalSubscription(ctx, "google", purchaseToken, userID, plan.PlanID); err != nil {
		return entitlementResponse{}, err
	}
	expiry, err := parseMillis(purchase.ExpiryTimeMillis)
	if err != nil {
		return entitlementResponse{}, err
	}
	now := time.Now().UTC()
	status := "expired"
	paid := purchase.PaymentState == 1 || purchase.PaymentState == 2
	if expiry.After(now) && paid {
		status = "active"
	}
	ent := Entitlement{
		UserID:           userID,
		Status:           status,
		Provider:         "google",
		PlanID:           plan.PlanID,
		CurrentPeriodEnd: &expiry,
		LastVerifiedAt:   &now,
	}
	if err := s.upsertEntitlement(ctx, ent); err != nil {
		return entitlementResponse{}, err
	}
	if paid {
		amount, currency := parsePriceAmount(purchase)
		providerTxID := strings.TrimSpace(purchase.OrderID)
		if providerTxID == "" {
			providerTxID = purchaseToken
		}
		if err := s.recordPayment(ctx, userID, "google", providerTxID, amount, currency, "succeeded"); err != nil {
			return entitlementResponse{}, err
		}
	}
	return entitlementResponse{
		Status:           status,
		Provider:         "google",
		PlanID:           plan.PlanID,
		CurrentPeriodEnd: &expiry,
	}, nil
}

func (s *Server) ensureGooglePlayConfig() error {
	missing := make([]string, 0)
	if strings.TrimSpace(s.cfg.GooglePlayPackageName) == "" {
		missing = append(missing, "GOOGLE_PLAY_PACKAGE_NAME")
	}
	if strings.TrimSpace(s.cfg.GooglePlayServiceAccountJSON) == "" {
		missing = append(missing, "GOOGLE_PLAY_SERVICE_ACCOUNT_JSON")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (s *Server) googlePlayClient(ctx context.Context) (*http.Client, error) {
	raw, err := parseGoogleServiceAccountJSON(s.cfg.GooglePlayServiceAccountJSON)
	if err != nil {
		return nil, err
	}
	creds, err := google.CredentialsFromJSON(ctx, raw, googlePlayScope)
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(ctx, creds.TokenSource), nil
}

func parseGoogleServiceAccountJSON(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("GOOGLE_PLAY_SERVICE_ACCOUNT_JSON missing")
	}
	if strings.HasPrefix(raw, "{") {
		return []byte(raw), nil
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid GOOGLE_PLAY_SERVICE_ACCOUNT_JSON encoding")
	}
	return decoded, nil
}

func parseMillis(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("expiry time missing")
	}
	millis, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, millis*int64(time.Millisecond)).UTC(), nil
}

func parsePriceAmount(purchase googleSubscriptionPurchase) (float64, string) {
	amount := 0.0
	if purchase.PriceAmountMicros != "" {
		if micros, err := purchase.PriceAmountMicros.Int64(); err == nil {
			amount = float64(micros) / 1_000_000
		}
	}
	currency := strings.ToUpper(strings.TrimSpace(purchase.PriceCurrencyCode))
	if currency == "" {
		currency = "USD"
	}
	return amount, currency
}

func (s *Server) upsertEntitlement(ctx context.Context, ent Entitlement) error {
	return s.db.DB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"status", "provider", "plan_id", "current_period_end", "last_verified_at",
		}),
	}).Create(&ent).Error
}

func (s *Server) recordPayment(ctx context.Context, userID, provider, providerTxID string, amount float64, currency string, status string) error {
	if strings.TrimSpace(providerTxID) == "" {
		return nil
	}
	var existing Payment
	if err := s.db.DB().WithContext(ctx).
		First(&existing, "provider = ? AND provider_tx_id = ?", provider, providerTxID).Error; err == nil {
		if existing.UserID != userID {
			return fmt.Errorf("%w", errPaymentConflict)
		}
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	payment := Payment{
		ID:           newID("pay"),
		UserID:       userID,
		Provider:     provider,
		ProviderTxID: providerTxID,
		Amount:       amount,
		Currency:     currency,
		Status:       status,
		CreatedAt:    time.Now().UTC(),
	}
	return s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(&payment).Error; err != nil {
			var claimed Payment
			if errLookup := tx.First(&claimed, "provider = ? AND provider_tx_id = ?", provider, providerTxID).Error; errLookup == nil {
				if claimed.UserID != userID {
					return fmt.Errorf("%w", errPaymentConflict)
				}
				return nil
			}
			return err
		}
		if status != "succeeded" || amount <= 0 {
			return nil
		}
		amountUSD, ok := s.amountToUSD(ctx, amount, currency)
		if !ok {
			return nil
		}
		return tx.Model(&User{}).Where("id = ?", userID).
			Update("total_paid_amount", gorm.Expr("total_paid_amount + ?", amountUSD)).Error
	})
}

func (s *Server) amountToUSD(ctx context.Context, amount float64, currency string) (float64, bool) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" || currency == "USD" {
		return amount, true
	}
	resp, err := s.market.openExchangeLatest(ctx)
	if err != nil {
		return 0, false
	}
	if rate, ok := resp.Rates[currency]; ok && rate > 0 {
		return amount / rate, true
	}
	return 0, false
}
