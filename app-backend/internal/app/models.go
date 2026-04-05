package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type User struct {
	ID                      string    `gorm:"column:id;primaryKey"`
	Email                   *string   `gorm:"column:email"`
	TotalPaidAmount         float64   `gorm:"column:total_paid_amount;type:numeric"`
	ActivePortfolioSnapshot *string   `gorm:"column:active_portfolio_snapshot_id"`
	CreatedAt               time.Time `gorm:"column:created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }

type AuthIdentity struct {
	ID             string    `gorm:"column:id;primaryKey"`
	UserID         string    `gorm:"column:user_id;index"`
	Provider       string    `gorm:"column:provider"`
	ProviderUserID string    `gorm:"column:provider_user_id"`
	Email          *string   `gorm:"column:email"`
	PasswordHash   *string   `gorm:"column:password_hash"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (AuthIdentity) TableName() string { return "auth_identities" }

type AuthSession struct {
	ID               string     `gorm:"column:id;primaryKey"`
	UserID           string     `gorm:"column:user_id;index"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (AuthSession) TableName() string { return "auth_sessions" }

type UserProfile struct {
	UserID            string         `gorm:"column:user_id;primaryKey"`
	Markets           pq.StringArray `gorm:"column:markets;type:text[]"`
	Experience        string         `gorm:"column:experience"`
	Style             string         `gorm:"column:style"`
	PainPoints        pq.StringArray `gorm:"column:pain_points;type:text[]"`
	RiskPreference    string         `gorm:"column:risk_preference"`
	RiskLevel         string         `gorm:"column:risk_level"`
	Language          string         `gorm:"column:language"`
	Timezone          string         `gorm:"column:timezone"`
	BaseCurrency      string         `gorm:"column:base_currency"`
	InsightsRefreshed *time.Time     `gorm:"column:insights_refreshed_at"`
	NotificationPrefs datatypes.JSON `gorm:"column:notification_prefs;type:jsonb"`
}

func (UserProfile) TableName() string { return "user_profiles" }

type DeviceToken struct {
	ID             string     `gorm:"column:id;primaryKey"`
	UserID         string     `gorm:"column:user_id;index"`
	Platform       string     `gorm:"column:platform"`
	PushProvider   string     `gorm:"column:push_provider"`
	DeviceToken    string     `gorm:"column:device_token"`
	ClientDeviceID *string    `gorm:"column:client_device_id"`
	Environment    string     `gorm:"column:environment"`
	AppVersion     string     `gorm:"column:app_version"`
	OSVersion      string     `gorm:"column:os_version"`
	Locale         string     `gorm:"column:locale"`
	Timezone       string     `gorm:"column:timezone"`
	PushEnabled    bool       `gorm:"column:push_enabled"`
	LastSeenAt     time.Time  `gorm:"column:last_seen_at"`
	RevokedAt      *time.Time `gorm:"column:revoked_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (DeviceToken) TableName() string { return "device_tokens" }

type UploadBatch struct {
	ID              string         `gorm:"column:id;primaryKey"`
	UserID          string         `gorm:"column:user_id;index"`
	Purpose         string         `gorm:"column:purpose"`
	Status          string         `gorm:"column:status"`
	ImageCount      int            `gorm:"column:image_count"`
	DeviceTimezone  *string        `gorm:"column:device_timezone"`
	BaseCurrency    string         `gorm:"column:base_currency"`
	BaseFXRateToUSD *float64       `gorm:"column:base_fx_rate_to_usd;type:numeric"`
	OCRPromptHash   *string        `gorm:"column:ocr_prompt_hash"`
	OCRModelOutput  *string        `gorm:"column:ocr_model_output_raw"`
	OCRParseError   *string        `gorm:"column:ocr_parse_error"`
	OCRRetryCount   int            `gorm:"column:ocr_retry_count"`
	ErrorCode       *string        `gorm:"column:error_code"`
	Warnings        pq.StringArray `gorm:"column:warnings;type:text[]"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	CompletedAt     *time.Time     `gorm:"column:completed_at"`
}

func (UploadBatch) TableName() string { return "upload_batches" }

type UploadImage struct {
	ID                string         `gorm:"column:id;primaryKey"`
	UploadBatchID     string         `gorm:"column:upload_batch_id;index"`
	StorageKey        string         `gorm:"column:storage_key"`
	Status            string         `gorm:"column:status"`
	ErrorReason       *string        `gorm:"column:error_reason"`
	PlatformGuess     string         `gorm:"column:platform_guess"`
	FingerprintV0     *string        `gorm:"column:fingerprint_v0"`
	FingerprintV1     *string        `gorm:"column:fingerprint_v1"`
	PHash             *string        `gorm:"column:phash"`
	IsDuplicate       bool           `gorm:"column:is_duplicate"`
	DuplicateOfImage  *string        `gorm:"column:duplicate_of_image_id"`
	DuplicateOverride bool           `gorm:"column:duplicate_override"`
	Warnings          pq.StringArray `gorm:"column:warnings;type:text[]"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
}

func (UploadImage) TableName() string { return "upload_images" }

type OCRAsset struct {
	ID                  string   `gorm:"column:id;primaryKey"`
	UploadImageID       string   `gorm:"column:upload_image_id;index"`
	SymbolRaw           string   `gorm:"column:symbol_raw"`
	Symbol              *string  `gorm:"column:symbol"`
	AssetType           string   `gorm:"column:asset_type"`
	AssetKey            *string  `gorm:"column:asset_key"`
	CoinGeckoID         *string  `gorm:"column:coingecko_id"`
	ExchangeMIC         *string  `gorm:"column:exchange_mic"`
	Amount              float64  `gorm:"column:amount;type:numeric"`
	ValueFromScreenshot *float64 `gorm:"column:value_from_screenshot;type:numeric"`
	ValueUSD            *float64 `gorm:"column:value_usd_priced;type:numeric"`
	ManualValueUSD      *float64 `gorm:"column:manual_value_usd;type:numeric"`
	DisplayCurrency     *string  `gorm:"column:display_currency"`
	Confidence          float64  `gorm:"column:confidence;type:numeric"`
	AvgPrice            *float64 `gorm:"column:avg_price;type:numeric"`
	AvgPriceSource      *string  `gorm:"column:avg_price_source"`
	PNLPercent          *float64 `gorm:"column:pnl_percent;type:numeric"`
}

func (OCRAsset) TableName() string { return "ocr_assets" }

type OCRAmbiguity struct {
	ID            string         `gorm:"column:id;primaryKey"`
	UploadBatchID string         `gorm:"column:upload_batch_id;index"`
	UploadImageID string         `gorm:"column:upload_image_id;index"`
	SymbolRaw     string         `gorm:"column:symbol_raw"`
	Candidates    datatypes.JSON `gorm:"column:candidates_json;type:jsonb"`
}

func (OCRAmbiguity) TableName() string { return "ocr_ambiguities" }

type AmbiguityResolution struct {
	ID                  string    `gorm:"column:id;primaryKey"`
	UserID              string    `gorm:"column:user_id;index"`
	SymbolRaw           string    `gorm:"column:symbol_raw"`
	SymbolRawNormalized string    `gorm:"column:symbol_raw_normalized"`
	PlatformCategory    string    `gorm:"column:platform_category"`
	AssetType           string    `gorm:"column:asset_type"`
	Symbol              string    `gorm:"column:symbol"`
	AssetKey            string    `gorm:"column:asset_key"`
	CoinGeckoID         *string   `gorm:"column:coingecko_id"`
	ExchangeMIC         *string   `gorm:"column:exchange_mic"`
	CreatedAt           time.Time `gorm:"column:created_at"`
}

func (AmbiguityResolution) TableName() string { return "ambiguity_resolutions" }

type UserAssetOverride struct {
	ID             string    `gorm:"column:id;primaryKey"`
	UserID         string    `gorm:"column:user_id;index"`
	AssetKey       string    `gorm:"column:asset_key"`
	AvgPrice       float64   `gorm:"column:avg_price;type:numeric"`
	AvgPriceSource string    `gorm:"column:avg_price_source"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (UserAssetOverride) TableName() string { return "user_asset_overrides" }

type MarketDataSnapshot struct {
	ID              string         `gorm:"column:id;primaryKey"`
	ValuationAsOf   time.Time      `gorm:"column:valuation_as_of"`
	BaseCurrency    string         `gorm:"column:base_currency"`
	ProviderPayload datatypes.JSON `gorm:"column:provider_payloads;type:jsonb"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
}

func (MarketDataSnapshot) TableName() string { return "market_data_snapshots" }

type MarketDataSnapshotItem struct {
	ID                   string         `gorm:"column:id;primaryKey"`
	MarketDataSnapshotID string         `gorm:"column:market_data_snapshot_id;index"`
	AssetType            string         `gorm:"column:asset_type"`
	Symbol               string         `gorm:"column:symbol"`
	AssetKey             string         `gorm:"column:asset_key"`
	CoinGeckoID          *string        `gorm:"column:coingecko_id"`
	ExchangeMIC          *string        `gorm:"column:exchange_mic"`
	PriceUSD             float64        `gorm:"column:price_usd;type:numeric"`
	PriceNative          *float64       `gorm:"column:price_native;type:numeric"`
	QuoteCurrency        *string        `gorm:"column:quote_currency"`
	FXRateToUSD          *float64       `gorm:"column:fx_rate_to_usd;type:numeric"`
	PriceSource          string         `gorm:"column:price_source"`
	RawPayload           datatypes.JSON `gorm:"column:raw_payload;type:jsonb"`
}

func (MarketDataSnapshotItem) TableName() string { return "market_data_snapshot_items" }

type AssetCatalogCrypto struct {
	CoinGeckoID string `gorm:"column:coingecko_id;primaryKey"`
	Symbol      string `gorm:"column:symbol"`
	Name        string `gorm:"column:name"`
	Slug        string `gorm:"column:slug"`
	IsActive    bool   `gorm:"column:is_active"`
}

func (AssetCatalogCrypto) TableName() string { return "asset_catalog_crypto" }

type AssetCatalogStock struct {
	TickerKey   string `gorm:"column:ticker_key;primaryKey"`
	Symbol      string `gorm:"column:symbol"`
	ExchangeMIC string `gorm:"column:exchange_mic"`
	Name        string `gorm:"column:name"`
	Currency    string `gorm:"column:currency"`
}

func (AssetCatalogStock) TableName() string { return "asset_catalog_stock" }

type AssetCatalogFX struct {
	Symbol string `gorm:"column:symbol;primaryKey"`
	Name   string `gorm:"column:name"`
}

func (AssetCatalogFX) TableName() string { return "asset_catalog_fx" }

type PortfolioSnapshot struct {
	ID                   string    `gorm:"column:id;primaryKey"`
	UserID               string    `gorm:"column:user_id;index"`
	SourceUploadBatchID  string    `gorm:"column:source_upload_batch_id"`
	MarketDataSnapshotID string    `gorm:"column:market_data_snapshot_id"`
	ValuationAsOf        time.Time `gorm:"column:valuation_as_of"`
	NetWorthUSD          float64   `gorm:"column:net_worth_usd;type:numeric"`
	BaseCurrency         string    `gorm:"column:base_currency"`
	BaseFXRateToUSD      *float64  `gorm:"column:base_fx_rate_to_usd;type:numeric"`
	SnapshotType         string    `gorm:"column:snapshot_type"`
	Status               string    `gorm:"column:status"`
	ReplacedBySnapshotID *string   `gorm:"column:replaced_by_snapshot_id"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (PortfolioSnapshot) TableName() string { return "portfolio_snapshots" }

type PortfolioHolding struct {
	ID                  string         `gorm:"column:id;primaryKey"`
	PortfolioSnapshotID string         `gorm:"column:portfolio_snapshot_id;index"`
	AssetType           string         `gorm:"column:asset_type"`
	Symbol              string         `gorm:"column:symbol"`
	AssetKey            string         `gorm:"column:asset_key"`
	CoinGeckoID         *string        `gorm:"column:coingecko_id"`
	ExchangeMIC         *string        `gorm:"column:exchange_mic"`
	Amount              float64        `gorm:"column:amount;type:numeric"`
	ValueFromScreenshot *float64       `gorm:"column:value_from_screenshot;type:numeric"`
	ValueUSD            float64        `gorm:"column:value_usd_priced;type:numeric"`
	PricingSource       string         `gorm:"column:pricing_source"`
	ValuationStatus     string         `gorm:"column:valuation_status"`
	CurrencyConverted   bool           `gorm:"column:currency_converted"`
	CostBasisStatus     string         `gorm:"column:cost_basis_status"`
	BalanceType         string         `gorm:"column:balance_type"`
	AvgPrice            *float64       `gorm:"column:avg_price;type:numeric"`
	AvgPriceSource      *string        `gorm:"column:avg_price_source"`
	PNLPercent          *float64       `gorm:"column:pnl_percent;type:numeric"`
	Sources             pq.StringArray `gorm:"column:sources;type:text[]"`
}

func (PortfolioHolding) TableName() string { return "portfolio_holdings" }

type PortfolioTransaction struct {
	ID               string     `gorm:"column:id;primaryKey"`
	UserID           string     `gorm:"column:user_id;index"`
	SnapshotIDBefore string     `gorm:"column:snapshot_id_before"`
	SnapshotIDAfter  string     `gorm:"column:snapshot_id_after"`
	Symbol           string     `gorm:"column:symbol"`
	AssetType        string     `gorm:"column:asset_type"`
	AssetKey         *string    `gorm:"column:asset_key"`
	Side             string     `gorm:"column:side"`
	Amount           float64    `gorm:"column:amount;type:numeric"`
	Price            float64    `gorm:"column:price;type:numeric"`
	Currency         string     `gorm:"column:currency"`
	ExecutedAt       *time.Time `gorm:"column:executed_at"`
	Fees             *float64   `gorm:"column:fees;type:numeric"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (PortfolioTransaction) TableName() string { return "portfolio_transactions" }

type Calculation struct {
	ID                  string         `gorm:"column:calculation_id;primaryKey"`
	PortfolioSnapshotID string         `gorm:"column:portfolio_snapshot_id"`
	StatusPreview       string         `gorm:"column:status_preview"`
	StatusPaid          string         `gorm:"column:status_paid"`
	HealthScore         int            `gorm:"column:health_score"`
	VolatilityScore     int            `gorm:"column:volatility_score"`
	HealthStatus        string         `gorm:"column:health_status"`
	MetricsIncomplete   bool           `gorm:"column:metrics_incomplete"`
	PricedCoveragePct   float64        `gorm:"column:priced_coverage_pct;type:numeric"`
	ModelVersionPreview string         `gorm:"column:model_version_preview"`
	ModelVersionPaid    *string        `gorm:"column:model_version_paid"`
	PromptHashPreview   string         `gorm:"column:prompt_hash_preview"`
	PromptHashPaid      *string        `gorm:"column:prompt_hash_paid"`
	PreviewPayload      datatypes.JSON `gorm:"column:preview_payload;type:jsonb"`
	PaidPayload         datatypes.JSON `gorm:"column:paid_payload;type:jsonb"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	PaidAt              *time.Time     `gorm:"column:paid_at"`
}

func (Calculation) TableName() string { return "calculations" }

type ReportRisk struct {
	ID            string  `gorm:"column:id;primaryKey"`
	CalculationID string  `gorm:"column:calculation_id;index"`
	RiskID        string  `gorm:"column:risk_id"`
	Type          string  `gorm:"column:type"`
	Severity      string  `gorm:"column:severity"`
	TeaserText    *string `gorm:"column:teaser_text"`
	Message       *string `gorm:"column:message"`
}

func (ReportRisk) TableName() string { return "report_risks" }

type ReportStrategy struct {
	ID              string         `gorm:"column:id;primaryKey"`
	CalculationID   string         `gorm:"column:calculation_id;index"`
	PlanID          string         `gorm:"column:plan_id"`
	StrategyID      string         `gorm:"column:strategy_id"`
	AssetType       string         `gorm:"column:asset_type"`
	Symbol          string         `gorm:"column:symbol"`
	AssetKey        string         `gorm:"column:asset_key"`
	LinkedRiskID    string         `gorm:"column:linked_risk_id"`
	Parameters      datatypes.JSON `gorm:"column:parameters;type:jsonb"`
	Rationale       string         `gorm:"column:rationale"`
	ExpectedOutcome string         `gorm:"column:expected_outcome"`
}

func (ReportStrategy) TableName() string { return "report_strategies" }

type PlanState struct {
	ID         string         `gorm:"column:id;primaryKey"`
	UserID     string         `gorm:"column:user_id;index"`
	PlanID     string         `gorm:"column:plan_id"`
	StrategyID string         `gorm:"column:strategy_id"`
	AssetKey   string         `gorm:"column:asset_key"`
	State      datatypes.JSON `gorm:"column:state_json;type:jsonb"`
	UpdatedAt  time.Time      `gorm:"column:updated_at"`
}

func (PlanState) TableName() string { return "plan_states" }

type Insight struct {
	ID                string         `gorm:"column:id;primaryKey"`
	UserID            string         `gorm:"column:user_id;index"`
	Type              string         `gorm:"column:type"`
	Asset             string         `gorm:"column:asset"`
	AssetKey          *string        `gorm:"column:asset_key"`
	Timeframe         *string        `gorm:"column:timeframe"`
	Severity          string         `gorm:"column:severity"`
	TriggerKey        string         `gorm:"column:trigger_key"`
	TriggerReason     string         `gorm:"column:trigger_reason"`
	StrategyID        *string        `gorm:"column:strategy_id"`
	PlanID            *string        `gorm:"column:plan_id"`
	SuggestedAction   *string        `gorm:"column:suggested_action"`
	SuggestedQuantity datatypes.JSON `gorm:"column:suggested_quantity;type:jsonb"`
	CTAPayload        datatypes.JSON `gorm:"column:cta_payload;type:jsonb"`
	BetaToPortfolio   *float64       `gorm:"column:beta_to_portfolio"`
	Status            string         `gorm:"column:status"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	ExpiresAt         time.Time      `gorm:"column:expires_at"`
}

func (Insight) TableName() string { return "insights" }

type InsightEvent struct {
	ID        string         `gorm:"column:id;primaryKey"`
	InsightID string         `gorm:"column:insight_id;index"`
	EventType string         `gorm:"column:event_type"`
	Metadata  datatypes.JSON `gorm:"column:metadata;type:jsonb"`
	CreatedAt time.Time      `gorm:"column:created_at"`
}

func (InsightEvent) TableName() string { return "insight_events" }

type Entitlement struct {
	UserID           string     `gorm:"column:user_id;primaryKey"`
	Status           string     `gorm:"column:status"`
	Provider         string     `gorm:"column:provider"`
	PlanID           string     `gorm:"column:plan_id"`
	CurrentPeriodEnd *time.Time `gorm:"column:current_period_end"`
	LastVerifiedAt   *time.Time `gorm:"column:last_verified_at"`
}

func (Entitlement) TableName() string { return "entitlements" }

type ExternalSubscription struct {
	ID         string    `gorm:"column:id;primaryKey"`
	Provider   string    `gorm:"column:provider;index:idx_external_subscription_unique,unique"`
	ExternalID string    `gorm:"column:external_id;index:idx_external_subscription_unique,unique"`
	UserID     string    `gorm:"column:user_id;index"`
	PlanID     string    `gorm:"column:plan_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (ExternalSubscription) TableName() string { return "external_subscriptions" }

type Payment struct {
	ID           string    `gorm:"column:id;primaryKey"`
	UserID       string    `gorm:"column:user_id;index"`
	Provider     string    `gorm:"column:provider;index:idx_payment_unique,unique"`
	ProviderTxID string    `gorm:"column:provider_tx_id;index:idx_payment_unique,unique"`
	Amount       float64   `gorm:"column:amount;type:numeric"`
	Currency     string    `gorm:"column:currency"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (Payment) TableName() string { return "payments" }

type WaitlistEntry struct {
	ID         string    `gorm:"column:id;primaryKey"`
	UserID     string    `gorm:"column:user_id;index"`
	StrategyID string    `gorm:"column:strategy_id"`
	Rank       int       `gorm:"column:rank"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (WaitlistEntry) TableName() string { return "waitlist_entries" }

type QuotaUsage struct {
	ID               string    `gorm:"column:id;primaryKey"`
	UserID           string    `gorm:"column:user_id;index"`
	UsageDay         time.Time `gorm:"column:usage_day"`
	TimezoneUsed     string    `gorm:"column:timezone_used"`
	WindowStartedUTC time.Time `gorm:"column:window_started_at_utc"`
	HoldingsCount    int       `gorm:"column:holdings_batches_count"`
}

func (QuotaUsage) TableName() string { return "quota_usage" }

type MarketDataCache struct {
	CacheKind   string         `gorm:"column:cache_kind;primaryKey"`
	CacheKey    string         `gorm:"column:cache_key;primaryKey"`
	CacheStatus string         `gorm:"column:cache_status"`
	Payload     datatypes.JSON `gorm:"column:payload;type:jsonb"`
	ExpiresAt   time.Time      `gorm:"column:expires_at;index"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (MarketDataCache) TableName() string { return "market_data_cache" }

type MarketCandlestick struct {
	Source    string    `gorm:"column:source;primaryKey"`
	AssetKey  string    `gorm:"column:asset_key;primaryKey;index"`
	AssetType string    `gorm:"column:asset_type;index"`
	Symbol    string    `gorm:"column:symbol;index"`
	Interval  string    `gorm:"column:interval;primaryKey;index"`
	Timestamp time.Time `gorm:"column:timestamp;primaryKey;index"`
	Open      float64   `gorm:"column:open;type:numeric"`
	High      float64   `gorm:"column:high;type:numeric"`
	Low       float64   `gorm:"column:low;type:numeric"`
	Close     float64   `gorm:"column:close;type:numeric"`
	Volume    float64   `gorm:"column:volume;type:numeric"`
	Currency  string    `gorm:"column:currency"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (MarketCandlestick) TableName() string { return "market_candlesticks" }

type FXDailyRate struct {
	RateDate     time.Time      `gorm:"column:rate_date;primaryKey"`
	BaseCurrency string         `gorm:"column:base_currency;primaryKey"`
	Rates        datatypes.JSON `gorm:"column:rates;type:jsonb"`
	Source       string         `gorm:"column:source"`
	Timestamp    int64          `gorm:"column:timestamp"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

func (FXDailyRate) TableName() string { return "fx_daily_rates" }

type Briefing struct {
	ID           string    `gorm:"column:id;primaryKey"`
	UserID       string    `gorm:"column:user_id;index"`
	Type         string    `gorm:"column:type"`
	Priority     int       `gorm:"column:priority"`
	Title        string    `gorm:"column:title"`
	Body         string    `gorm:"column:body"`
	PushText     string    `gorm:"column:push_text"`
	BriefingDate string    `gorm:"column:briefing_date;index"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (Briefing) TableName() string { return "briefings" }
