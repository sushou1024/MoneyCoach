package app

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds all runtime configuration derived from env vars.
type Config struct {
	Port                         string
	DatabaseURL                  string
	RedisURL                     string
	ObjectStorageMode            string
	ObjectStorageLocalDir        string
	LocalStorageBaseURL          string
	ObjectStorageBucket          string
	ObjectStorageRegion          string
	LogosXHKGBaseURL             string
	LogosNasdaqBaseURL           string
	LogosNyseBaseURL             string
	JWTSigningSecret             string
	GeminiAPIKey                 string
	ResendAPIKey                 string
	ResendFromEmail              string
	CoinGeckoAPIKey              string
	CMCProAPIKey                 string
	MarketstackAccessKey         string
	OpenExchangeAppID            string
	BinanceAPIBaseURL            string
	BinanceFuturesBaseURL        string
	GoogleAllowedClient          []string
	AppleAllowedClient           []string
	AppleIOSBundleID             string
	APNSKeyID                    string
	APNSTeamID                   string
	APNSPrivateKey               string
	APNSBundleID                 string
	FCMServerKey                 string
	EmailOTPMode                 string
	GooglePlayPackageName        string
	GooglePlayServiceAccountJSON string
	GooglePlayWeeklyProductID    string
	GooglePlayAnnualProductID    string
	GooglePubSubAudience         string
	BillingDevMode               bool
	StripeSecretKey              string
	StripeWebhookSecret          string
	StripePriceIDWeekly          string
	StripePriceIDYearly          string
	AppStoreSharedSecret         string
	AppleIapProductIDWeekly      string
	AppleIapProductIDYearly      string
	APICorsAllowedOrigins        []string
	EnableDangerousResetRoutes   bool
	ResetAppSecret               string
}

// LoadConfig reads required env vars and returns a config struct.
func LoadConfig() (Config, error) {
	storageMode := strings.ToLower(strings.TrimSpace(os.Getenv("OBJECT_STORAGE_MODE")))
	if storageMode == "" {
		storageMode = "s3"
	}

	cfg := Config{
		Port:                         getenvDefault("PORT", "8080"),
		DatabaseURL:                  os.Getenv("DATABASE_URL"),
		RedisURL:                     os.Getenv("REDIS_URL"),
		ObjectStorageMode:            storageMode,
		ObjectStorageLocalDir:        os.Getenv("OBJECT_STORAGE_LOCAL_DIR"),
		LocalStorageBaseURL:          os.Getenv("LOCAL_STORAGE_BASE_URL"),
		ObjectStorageBucket:          os.Getenv("OBJECT_STORAGE_BUCKET"),
		ObjectStorageRegion:          os.Getenv("OBJECT_STORAGE_REGION"),
		LogosXHKGBaseURL:             os.Getenv("LOGOS_XHKG_BASE_URL"),
		LogosNasdaqBaseURL:           os.Getenv("LOGOS_NASDAQ_BASE_URL"),
		LogosNyseBaseURL:             os.Getenv("LOGOS_NYSE_BASE_URL"),
		JWTSigningSecret:             os.Getenv("JWT_SIGNING_SECRET"),
		GeminiAPIKey:                 os.Getenv("GEMINI_API_KEY"),
		ResendAPIKey:                 os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:              os.Getenv("RESEND_FROM_EMAIL"),
		CoinGeckoAPIKey:              os.Getenv("COINGECKO_PRO_API_KEY"),
		CMCProAPIKey:                 os.Getenv("CMC_PRO_API_KEY"),
		MarketstackAccessKey:         os.Getenv("MARKETSTACK_ACCESS_KEY"),
		OpenExchangeAppID:            os.Getenv("OPEN_EXCHANGE_APP_ID"),
		BinanceAPIBaseURL:            os.Getenv("BINANCE_API_BASE_URL"),
		BinanceFuturesBaseURL:        getenvDefault("BINANCE_FUTURES_BASE_URL", "https://fapi.binance.com"),
		GoogleAllowedClient:          splitCSV(os.Getenv("GOOGLE_ALLOWED_CLIENT_IDS")),
		AppleAllowedClient:           splitCSV(os.Getenv("APPLE_IOS_BUNDLE_ID")),
		AppleIOSBundleID:             os.Getenv("APPLE_IOS_BUNDLE_ID"),
		APNSKeyID:                    os.Getenv("APNS_KEY_ID"),
		APNSTeamID:                   os.Getenv("APNS_TEAM_ID"),
		APNSPrivateKey:               os.Getenv("APNS_PRIVATE_KEY"),
		APNSBundleID:                 os.Getenv("APNS_BUNDLE_ID"),
		FCMServerKey:                 os.Getenv("FCM_SERVER_KEY"),
		EmailOTPMode:                 os.Getenv("EMAIL_OTP_MODE"),
		GooglePlayPackageName:        os.Getenv("GOOGLE_PLAY_PACKAGE_NAME"),
		GooglePlayServiceAccountJSON: getenvDefault("GOOGLE_PLAY_SERVICE_ACCOUNT_JSON", os.Getenv("GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64")),
		GooglePlayWeeklyProductID:    os.Getenv("GOOGLE_PLAY_PRODUCT_WEEKLY"),
		GooglePlayAnnualProductID:    os.Getenv("GOOGLE_PLAY_PRODUCT_ANNUAL"),
		GooglePubSubAudience:         os.Getenv("GOOGLE_PUBSUB_AUDIENCE"),
		BillingDevMode:               parseBoolEnv("BILLING_DEV_MODE"),
		StripeSecretKey:              os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:          os.Getenv("STRIPE_WEBHOOK_SECRET"),
		StripePriceIDWeekly:          os.Getenv("STRIPE_PRICE_ID_WEEKLY"),
		StripePriceIDYearly:          os.Getenv("STRIPE_PRICE_ID_YEARLY"),
		AppStoreSharedSecret:         os.Getenv("APPSTORE_SHARED_SECRET"),
		AppleIapProductIDWeekly:      os.Getenv("APPLE_IAP_PRODUCT_ID_WEEKLY"),
		AppleIapProductIDYearly:      os.Getenv("APPLE_IAP_PRODUCT_ID_YEARLY"),
		APICorsAllowedOrigins:        splitCSV(os.Getenv("API_CORS_ALLOWED_ORIGINS")),
		EnableDangerousResetRoutes:   parseBoolEnv("ENABLE_DANGEROUS_RESET_ROUTES"),
		ResetAppSecret:               os.Getenv("RESET_APP_SECRET"),
	}
	cfg.DatabaseURL = normalizeDatabaseURL(cfg.DatabaseURL)

	if cfg.ObjectStorageMode != "s3" && cfg.ObjectStorageMode != "local" {
		return cfg, fmt.Errorf("invalid OBJECT_STORAGE_MODE %q", cfg.ObjectStorageMode)
	}

	if cfg.ObjectStorageMode == "local" {
		if strings.TrimSpace(cfg.ObjectStorageLocalDir) == "" {
			cfg.ObjectStorageLocalDir = "local-uploads"
		}
		if strings.TrimSpace(cfg.LocalStorageBaseURL) == "" {
			cfg.LocalStorageBaseURL = "http://localhost:" + cfg.Port
		}
	}

	var missing []string
	required := []struct {
		name  string
		value string
	}{
		{"DATABASE_URL", cfg.DatabaseURL},
		{"REDIS_URL", cfg.RedisURL},
		{"JWT_SIGNING_SECRET", cfg.JWTSigningSecret},
		{"GEMINI_API_KEY", cfg.GeminiAPIKey},
		{"RESEND_API_KEY", cfg.ResendAPIKey},
		{"RESEND_FROM_EMAIL", cfg.ResendFromEmail},
		{"COINGECKO_PRO_API_KEY", cfg.CoinGeckoAPIKey},
		{"MARKETSTACK_ACCESS_KEY", cfg.MarketstackAccessKey},
		{"OPEN_EXCHANGE_APP_ID", cfg.OpenExchangeAppID},
		{"BINANCE_API_BASE_URL", cfg.BinanceAPIBaseURL},
		{"LOGOS_XHKG_BASE_URL", cfg.LogosXHKGBaseURL},
		{"LOGOS_NASDAQ_BASE_URL", cfg.LogosNasdaqBaseURL},
		{"LOGOS_NYSE_BASE_URL", cfg.LogosNyseBaseURL},
	}
	if cfg.ObjectStorageMode == "s3" {
		required = append(required, struct {
			name  string
			value string
		}{"OBJECT_STORAGE_BUCKET", cfg.ObjectStorageBucket})
		required = append(required, struct {
			name  string
			value string
		}{"OBJECT_STORAGE_REGION", cfg.ObjectStorageRegion})
	}

	for _, req := range required {
		if strings.TrimSpace(req.value) == "" {
			missing = append(missing, req.name)
		}
	}
	if len(missing) > 0 {
		return cfg, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	if cfg.EnableDangerousResetRoutes && strings.TrimSpace(cfg.ResetAppSecret) == "" {
		return cfg, fmt.Errorf("missing required env vars: RESET_APP_SECRET")
	}

	return cfg, nil
}

func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func normalizeDatabaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if !strings.HasPrefix(raw, "postgres://") && !strings.HasPrefix(raw, "postgresql://") {
		return raw
	}
	if _, err := url.Parse(raw); err == nil {
		return raw
	}
	schemeIdx := strings.Index(raw, "://")
	if schemeIdx == -1 {
		return raw
	}
	scheme := raw[:schemeIdx+3]
	rest := raw[schemeIdx+3:]
	atIdx := strings.Index(rest, "@")
	if atIdx == -1 {
		return raw
	}
	userInfo := rest[:atIdx]
	hostAndPath := rest[atIdx+1:]
	user, pass, ok := strings.Cut(userInfo, ":")
	if !ok || strings.TrimSpace(pass) == "" {
		return raw
	}
	encodedPass := url.QueryEscape(pass)
	return scheme + user + ":" + encodedPass + "@" + hostAndPath
}

func parseBoolEnv(key string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
