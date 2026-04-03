package app

import "time"

type ocrAssetInput struct {
	AssetID             string
	ImageID             string
	PlatformGuess       string
	SymbolRaw           string
	Symbol              *string
	AssetType           string
	Amount              float64
	ValueFromScreenshot *float64
	ManualValueUSD      *float64
	DisplayCurrency     *string
	PNLPercent          *float64
	AvgPrice            *float64
	AvgPriceSource      string
}

type resolvedAsset struct {
	AssetID string
	ImageID string
	Holding portfolioHolding
}

type portfolioHolding struct {
	SymbolRaw           string
	Symbol              string
	AssetType           string
	AssetKey            string
	CoinGeckoID         string
	ExchangeMIC         string
	Amount              float64
	ValueUSD            float64
	ValueFromScreenshot *float64
	ManualValueUSD      *float64
	DisplayCurrency     *string
	PricingSource       string
	ValuationStatus     string
	BalanceType         string
	CurrentPrice        float64
	CurrentPriceQuote   float64
	QuoteCurrency       string
	FXRateToUSD         float64
	AvgPrice            *float64
	AvgPriceSource      string
	PNLPercent          *float64
	CostBasisStatus     string
	CurrencyConverted   bool
}

type portfolioMetrics struct {
	NetWorthUSD             float64
	PricedValueUSD          float64
	NonCashPricedValueUSD   float64
	IdleCashUSD             float64
	CashPct                 float64
	TopAssetPct             float64
	PricedCoveragePct       float64
	MetricsIncomplete       bool
	Volatility30dDaily      float64
	Volatility30dAnnualized float64
	MaxDrawdown90d          float64
	AvgPairwiseCorr         float64
	HealthScoreBaseline     int
	VolatilityScoreBaseline int
	CryptoWeight            float64
}

type lockedPlan struct {
	PlanID        string         `json:"plan_id"`
	StrategyID    string         `json:"strategy_id"`
	AssetType     string         `json:"asset_type"`
	Symbol        string         `json:"symbol"`
	AssetKey      string         `json:"asset_key"`
	QuoteCurrency string         `json:"quote_currency,omitempty"`
	LinkedRiskID  string         `json:"linked_risk_id"`
	Parameters    map[string]any `json:"parameters,omitempty"`
}

type insightItem struct {
	ID                string
	Type              string
	Asset             string
	AssetKey          string
	Timeframe         string
	Severity          string
	TriggerReason     string
	TriggerKey        string
	StrategyID        string
	PlanID            string
	SuggestedAction   string
	SuggestedQuantity *suggestedQuantity
	CTAPayload        map[string]any
	BetaToPortfolio   float64
	CreatedAt         time.Time
	ExpiresAt         time.Time
}

type suggestedQuantity struct {
	Mode            string           `json:"mode"`
	AmountUSD       *float64         `json:"amount_usd,omitempty"`
	AmountDisplay   *float64         `json:"amount_display,omitempty"`
	AmountAsset     *float64         `json:"amount_asset,omitempty"`
	Symbol          string           `json:"symbol,omitempty"`
	AssetKey        string           `json:"asset_key,omitempty"`
	Trades          []rebalanceTrade `json:"trades,omitempty"`
	DisplayCurrency string           `json:"display_currency,omitempty"`
}

type rebalanceTrade struct {
	AssetKey    string   `json:"asset_key"`
	Symbol      string   `json:"symbol"`
	Side        string   `json:"side"`
	AmountUSD   float64  `json:"amount_usd"`
	AmountAsset *float64 `json:"amount_asset,omitempty"`
}

type ohlcPoint struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type pricePoint struct {
	Timestamp int64
	Value     float64
}

type returnPoint struct {
	Timestamp int64
	Value     float64
}

type marketstackPriceQuote struct {
	Close         float64
	PriceCurrency string
	Symbol        string
	Exchange      string
}

// insightIndicatorSeries carries indicator inputs and metadata.
type indicatorSeries struct {
	AssetKey string
	Symbol   string
	Interval string
	Source   string
	Points   []ohlcPoint
}

type indicatorSnapshot struct {
	Asset     string
	AssetKey  string
	AssetType string
	Interval  string
	Source    string
	RSI       float64
	UpperBand float64
	LowerBand float64
	Close     float64
	Timestamp int64
}

type futuresPremiumIndex struct {
	Symbol          string
	MarkPrice       float64
	LastFundingRate float64
	NextFundingTime time.Time
}
