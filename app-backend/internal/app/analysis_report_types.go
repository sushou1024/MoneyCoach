package app

type previewPayload struct {
	MetaData             metaDataPayload        `json:"meta_data"`
	ValuationAsOf        string                 `json:"valuation_as_of"`
	MarketDataSnapshotID string                 `json:"market_data_snapshot_id"`
	UserProfile          userProfilePayload     `json:"user_profile"`
	Portfolio            portfolioPayload       `json:"portfolio"`
	ComputedMetrics      computedMetricsPayload `json:"computed_metrics"`
	FixedMetrics         previewFixedMetrics    `json:"fixed_metrics"`
	NetWorthDisplay      float64                `json:"net_worth_display"`
	BaseCurrency         string                 `json:"base_currency"`
	BaseFXRateToUSD      float64                `json:"base_fx_rate_to_usd"`
}

type metaDataPayload struct {
	CalculationID string `json:"calculation_id"`
}

type userProfilePayload struct {
	RiskTolerance  string   `json:"risk_tolerance"`
	RiskPreference string   `json:"risk_preference,omitempty"`
	PainPoints     []string `json:"pain_points"`
	Experience     string   `json:"experience"`
	Style          string   `json:"style"`
	Markets        []string `json:"markets"`
}

type portfolioPayload struct {
	NetWorthUSD float64                   `json:"net_worth_usd"`
	Holdings    []portfolioHoldingPayload `json:"holdings"`
}

type portfolioHoldingPayload struct {
	AssetKey  string  `json:"asset_key"`
	Symbol    string  `json:"symbol"`
	AssetType string  `json:"asset_type"`
	Amount    float64 `json:"amount"`
	ValueUSD  float64 `json:"value_usd"`
}

type computedMetricsPayload struct {
	NetWorthUSD             float64 `json:"net_worth_usd"`
	CashPct                 float64 `json:"cash_pct"`
	TopAssetPct             float64 `json:"top_asset_pct"`
	Volatility30dAnnualized float64 `json:"volatility_30d_annualized"`
	MaxDrawdown90d          float64 `json:"max_drawdown_90d"`
	AvgPairwiseCorr         float64 `json:"avg_pairwise_corr"`
	HealthScoreBaseline     int     `json:"health_score_baseline"`
	VolatilityScoreBaseline int     `json:"volatility_score_baseline"`
	PricedCoveragePct       float64 `json:"priced_coverage_pct"`
	MetricsIncomplete       bool    `json:"metrics_incomplete"`
}

type previewReportPayload struct {
	MetaData             metaData              `json:"meta_data"`
	ValuationAsOf        string                `json:"valuation_as_of"`
	MarketDataSnapshotID string                `json:"market_data_snapshot_id"`
	FixedMetrics         previewFixedMetrics   `json:"fixed_metrics"`
	NetWorthDisplay      float64               `json:"net_worth_display,omitempty"`
	BaseCurrency         string                `json:"base_currency,omitempty"`
	BaseFXRateToUSD      float64               `json:"base_fx_rate_to_usd,omitempty"`
	AssetAllocation      []assetAllocationItem `json:"asset_allocation,omitempty"`
	IdentifiedRisks      []previewRisk         `json:"identified_risks"`
	LockedProjection     lockedProjection      `json:"locked_projection"`
}

type previewPromptOutput struct {
	MetaData             metaData            `json:"meta_data"`
	ValuationAsOf        string              `json:"valuation_as_of"`
	MarketDataSnapshotID string              `json:"market_data_snapshot_id"`
	FixedMetrics         previewFixedMetrics `json:"fixed_metrics"`
	NetWorthDisplay      float64             `json:"net_worth_display"`
	BaseCurrency         string              `json:"base_currency"`
	BaseFXRateToUSD      float64             `json:"base_fx_rate_to_usd"`
	IdentifiedRisks      []previewRisk       `json:"identified_risks"`
	LockedProjection     lockedProjection    `json:"locked_projection"`
}

type metaData struct {
	CalculationID string `json:"calculation_id"`
}

type previewFixedMetrics struct {
	NetWorthUSD     float64 `json:"net_worth_usd"`
	HealthScore     int     `json:"health_score"`
	HealthStatus    string  `json:"health_status"`
	VolatilityScore int     `json:"volatility_score"`
}

type previewRisk struct {
	RiskID     string `json:"risk_id"`
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	TeaserText string `json:"teaser_text"`
}

type lockedProjection struct {
	PotentialUpside string `json:"potential_upside"`
	CTA             string `json:"cta"`
}

type paidPayload struct {
	UserPortfolio   portfolioPayload      `json:"user_portfolio"`
	UserProfile     userProfilePayload    `json:"user_profile"`
	PreviousTeaser  previewReportPayload  `json:"previous_teaser"`
	LockedPlans     []lockedPlan          `json:"locked_plans"`
	PortfolioFacts  portfolioFactsPayload `json:"portfolio_facts"`
	FixedMetrics    previewFixedMetrics   `json:"fixed_metrics"`
	NetWorthDisplay float64               `json:"net_worth_display"`
	BaseCurrency    string                `json:"base_currency"`
	BaseFXRateToUSD float64               `json:"base_fx_rate_to_usd"`
}

type directPaidPayload struct {
	MetaData             metaDataPayload        `json:"meta_data"`
	ValuationAsOf        string                 `json:"valuation_as_of"`
	MarketDataSnapshotID string                 `json:"market_data_snapshot_id"`
	UserPortfolio        portfolioPayload       `json:"user_portfolio"`
	UserProfile          userProfilePayload     `json:"user_profile"`
	ComputedMetrics      computedMetricsPayload `json:"computed_metrics"`
	IdentifiedRisks      []previewRisk          `json:"identified_risks"`
	LockedPlans          []lockedPlan           `json:"locked_plans"`
	PortfolioFacts       portfolioFactsPayload  `json:"portfolio_facts"`
	FixedMetrics         previewFixedMetrics    `json:"fixed_metrics"`
	NetWorthDisplay      float64                `json:"net_worth_display"`
	BaseCurrency         string                 `json:"base_currency"`
	BaseFXRateToUSD      float64                `json:"base_fx_rate_to_usd"`
}

type portfolioFactsPayload struct {
	NetWorthUSD             float64 `json:"net_worth_usd"`
	CashPct                 float64 `json:"cash_pct"`
	TopAssetPct             float64 `json:"top_asset_pct"`
	Volatility30dAnnualized float64 `json:"volatility_30d_annualized"`
	MaxDrawdown90d          float64 `json:"max_drawdown_90d"`
	AvgPairwiseCorr         float64 `json:"avg_pairwise_corr"`
	PricedCoveragePct       float64 `json:"priced_coverage_pct"`
	MetricsIncomplete       bool    `json:"metrics_incomplete"`
}

type paidReportPayload struct {
	MetaData             metaData              `json:"meta_data"`
	ValuationAsOf        string                `json:"valuation_as_of"`
	MarketDataSnapshotID string                `json:"market_data_snapshot_id"`
	NetWorthDisplay      float64               `json:"net_worth_display,omitempty"`
	BaseCurrency         string                `json:"base_currency,omitempty"`
	BaseFXRateToUSD      float64               `json:"base_fx_rate_to_usd,omitempty"`
	AssetAllocation      []assetAllocationItem `json:"asset_allocation,omitempty"`
	ReportHeader         paidReportHeader      `json:"report_header"`
	Charts               paidCharts            `json:"charts"`
	RiskInsights         []paidRisk            `json:"risk_insights"`
	OptimizationPlan     []paidPlan            `json:"optimization_plan"`
	TheVerdict           paidVerdict           `json:"the_verdict"`
	RiskSummary          string                `json:"risk_summary"`
	ExposureAnalysis     []paidRisk            `json:"exposure_analysis"`
	ActionableAdvice     []paidPlan            `json:"actionable_advice"`
	DailyAlphaSignal     *paidInsightItem      `json:"daily_alpha_signal"`
}

type paidPromptOutput struct {
	MetaData             metaData         `json:"meta_data"`
	ValuationAsOf        string           `json:"valuation_as_of"`
	MarketDataSnapshotID string           `json:"market_data_snapshot_id"`
	NetWorthDisplay      float64          `json:"net_worth_display"`
	BaseCurrency         string           `json:"base_currency"`
	BaseFXRateToUSD      float64          `json:"base_fx_rate_to_usd"`
	ReportHeader         paidReportHeader `json:"report_header"`
	Charts               paidCharts       `json:"charts"`
	RiskInsights         []paidRisk       `json:"risk_insights"`
	OptimizationPlan     []paidPlan       `json:"optimization_plan"`
	TheVerdict           paidVerdict      `json:"the_verdict"`
}

type paidReportHeader struct {
	HealthScore         scoredValue `json:"health_score"`
	VolatilityDashboard scoredValue `json:"volatility_dashboard"`
}

type scoredValue struct {
	Value  int    `json:"value"`
	Status string `json:"status"`
}

type paidCharts struct {
	RadarChart paidRadarChart `json:"radar_chart"`
}

type paidRadarChart struct {
	Liquidity       int `json:"liquidity"`
	Diversification int `json:"diversification"`
	Alpha           int `json:"alpha"`
	Drawdown        int `json:"drawdown"`
}

type paidRisk struct {
	RiskID   string `json:"risk_id"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type paidPlan struct {
	PlanID           string         `json:"plan_id"`
	StrategyID       string         `json:"strategy_id"`
	AssetType        string         `json:"asset_type"`
	Symbol           string         `json:"symbol"`
	AssetKey         string         `json:"asset_key"`
	QuoteCurrency    string         `json:"quote_currency,omitempty"`
	LinkedRiskID     string         `json:"linked_risk_id"`
	Priority         string         `json:"priority,omitempty"`
	Parameters       map[string]any `json:"parameters,omitempty"`
	ExecutionSummary string         `json:"execution_summary,omitempty"`
	Rationale        string         `json:"rationale"`
	ExpectedOutcome  string         `json:"expected_outcome"`
}

type paidVerdict struct {
	ConstructiveComment string `json:"constructive_comment,omitempty"`
}

type paidInsightItem struct {
	Type            string `json:"type"`
	Asset           string `json:"asset"`
	AssetKey        string `json:"asset_key,omitempty"`
	Timeframe       string `json:"timeframe,omitempty"`
	Severity        string `json:"severity"`
	TriggerReason   string `json:"trigger_reason"`
	TriggerKey      string `json:"trigger_key"`
	StrategyID      string `json:"strategy_id,omitempty"`
	PlanID          string `json:"plan_id,omitempty"`
	SuggestedAction string `json:"suggested_action,omitempty"`
	CreatedAt       string `json:"created_at"`
	ExpiresAt       string `json:"expires_at"`
}
