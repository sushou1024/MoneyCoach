package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

const intelligenceLookbackDays = 260

type intelligenceRegimeResponse struct {
	AsOf            string                      `json:"as_of"`
	Scope           string                      `json:"scope"`
	Regime          string                      `json:"regime"`
	TrendStrength   string                      `json:"trend_strength"`
	Metrics         intelligenceRegimeMetrics   `json:"metrics"`
	TrendBreadth    intelligenceTrendBreadth    `json:"trend_breadth"`
	Drivers         []intelligenceDriver        `json:"drivers"`
	PortfolioImpact []intelligenceSemanticItem  `json:"portfolio_impact"`
	Actions         []intelligenceSemanticItem  `json:"actions"`
	Leaders         []intelligenceAssetLeader   `json:"leaders"`
	Laggards        []intelligenceAssetLeader   `json:"laggards"`
	FeaturedAssets  []intelligenceFeaturedAsset `json:"featured_assets"`
}

type intelligenceRegimeMetrics struct {
	Alpha30d                float64 `json:"alpha_30d"`
	Volatility30dAnnualized float64 `json:"volatility_30d_annualized"`
	MaxDrawdown90d          float64 `json:"max_drawdown_90d"`
	AvgPairwiseCorr         float64 `json:"avg_pairwise_corr"`
	CashPct                 float64 `json:"cash_pct"`
	TopAssetPct             float64 `json:"top_asset_pct"`
	PricedCoveragePct       float64 `json:"priced_coverage_pct"`
}

type intelligenceTrendBreadth struct {
	UpCount       int     `json:"up_count"`
	DownCount     int     `json:"down_count"`
	NeutralCount  int     `json:"neutral_count"`
	WeightedScore float64 `json:"weighted_score"`
}

type intelligenceDriver struct {
	ID         string   `json:"id"`
	Kind       string   `json:"kind"`
	Tone       string   `json:"tone"`
	ValueText  string   `json:"value_text"`
	Value      *float64 `json:"value,omitempty"`
	UpCount    *int     `json:"up_count,omitempty"`
	DownCount  *int     `json:"down_count,omitempty"`
	TotalCount *int     `json:"total_count,omitempty"`
}

type intelligenceSemanticItem struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type intelligenceAssetLeader struct {
	AssetKey   string  `json:"asset_key"`
	Symbol     string  `json:"symbol"`
	AssetType  string  `json:"asset_type"`
	Name       *string `json:"name,omitempty"`
	LogoURL    *string `json:"logo_url,omitempty"`
	Change30d  float64 `json:"change_30d"`
	WeightPct  float64 `json:"weight_pct"`
	TrendState string  `json:"trend_state"`
}

type intelligenceFeaturedAsset struct {
	AssetKey             string  `json:"asset_key"`
	Symbol               string  `json:"symbol"`
	AssetType            string  `json:"asset_type"`
	Name                 *string `json:"name,omitempty"`
	LogoURL              *string `json:"logo_url,omitempty"`
	ActionBias           string  `json:"action_bias"`
	SummarySignal        string  `json:"summary_signal"`
	WeightPct            float64 `json:"weight_pct"`
	BetaToPortfolio      float64 `json:"beta_to_portfolio"`
	SignalCount          int     `json:"signal_count"`
	LatestSignalSeverity *string `json:"latest_signal_severity,omitempty"`
}

type intelligenceAssetBriefResponse struct {
	AsOf            string                       `json:"as_of"`
	AssetKey        string                       `json:"asset_key"`
	Symbol          string                       `json:"symbol"`
	AssetType       string                       `json:"asset_type"`
	Name            *string                      `json:"name,omitempty"`
	LogoURL         *string                      `json:"logo_url,omitempty"`
	ExchangeMIC     *string                      `json:"exchange_mic,omitempty"`
	QuoteCurrency   string                       `json:"quote_currency"`
	CurrentPrice    float64                      `json:"current_price"`
	PriceChange24h  float64                      `json:"price_change_24h"`
	PriceChange7d   float64                      `json:"price_change_7d"`
	PriceChange30d  float64                      `json:"price_change_30d"`
	ActionBias      string                       `json:"action_bias"`
	SummarySignal   string                       `json:"summary_signal"`
	EntryZone       intelligencePriceZone        `json:"entry_zone"`
	Invalidation    intelligenceInvalidation     `json:"invalidation"`
	Technicals      intelligenceAssetTechnicals  `json:"technicals"`
	PortfolioFit    intelligencePortfolioFit     `json:"portfolio_fit"`
	WhyNow          []intelligenceSemanticItem   `json:"why_now"`
	RelatedInsights []intelligenceRelatedInsight `json:"related_insights"`
	RelatedPlans    []intelligenceRelatedPlan    `json:"related_plans"`
}

type intelligencePriceZone struct {
	Low   float64 `json:"low"`
	High  float64 `json:"high"`
	Basis string  `json:"basis"`
}

type intelligenceInvalidation struct {
	Price  float64 `json:"price"`
	Reason string  `json:"reason"`
}

type intelligenceAssetTechnicals struct {
	RSI14          float64 `json:"rsi_14"`
	BollingerUpper float64 `json:"bollinger_upper"`
	BollingerLower float64 `json:"bollinger_lower"`
	MA20           float64 `json:"ma_20"`
	MA50           float64 `json:"ma_50"`
	MA200          float64 `json:"ma_200"`
	TrendState     string  `json:"trend_state"`
	TrendStrength  string  `json:"trend_strength"`
}

type intelligencePortfolioFit struct {
	IsHeld              bool    `json:"is_held"`
	WeightPct           float64 `json:"weight_pct"`
	BetaToPortfolio     float64 `json:"beta_to_portfolio"`
	Role                string  `json:"role"`
	ConcentrationImpact string  `json:"concentration_impact"`
	RiskFlag            string  `json:"risk_flag"`
}

type intelligenceRelatedInsight struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Severity      string  `json:"severity"`
	TriggerReason string  `json:"trigger_reason"`
	CreatedAt     string  `json:"created_at"`
	PlanID        *string `json:"plan_id,omitempty"`
	StrategyID    *string `json:"strategy_id,omitempty"`
}

type intelligenceRelatedPlan struct {
	CalculationID   string `json:"calculation_id"`
	PlanID          string `json:"plan_id"`
	StrategyID      string `json:"strategy_id"`
	Priority        string `json:"priority"`
	Rationale       string `json:"rationale"`
	ExpectedOutcome string `json:"expected_outcome"`
}

type intelligenceContext struct {
	Profile                 UserProfile
	Snapshot                PortfolioSnapshot
	Holdings                []portfolioHolding
	SeriesByAssetKey        map[string][]ohlcPoint
	Metrics                 portfolioMetrics
	ActiveInsights          []Insight
	LatestPaidCalculationID string
	LatestStrategyRows      []ReportStrategy
	RiskSeverityByID        map[string]string
}

type intelligenceAssetData struct {
	Holding         portfolioHolding
	Name            *string
	LogoURL         *string
	WeightPct       float64
	IsHeld          bool
	SignalCount     int
	LatestSeverity  *string
	Series          []ohlcPoint
	Closes          []float64
	CurrentPrice    float64
	Change24h       float64
	Change7d        float64
	Change30d       float64
	RSI14           float64
	HasRSI14        bool
	BollingerUpper  float64
	BollingerLower  float64
	HasBollinger    bool
	MA20            float64
	HasMA20         bool
	MA50            float64
	HasMA50         bool
	MA200           float64
	HasMA200        bool
	TrendState      string
	TrendStrength   string
	BetaToPortfolio float64
	ActionBias      string
	SummarySignal   string
	EntryZone       intelligencePriceZone
	Invalidation    intelligenceInvalidation
}

type assetSignalSummary struct {
	count          int
	latestSeverity string
}

func (s *Server) buildIntelligenceContext(ctx context.Context, userID string) (*intelligenceContext, error) {
	var user User
	if err := s.db.DB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	if user.ActivePortfolioSnapshot == nil {
		return nil, fmt.Errorf("active portfolio not found")
	}

	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	snapshot, holdings, err := s.loadSnapshotWithHoldings(ctx, *user.ActivePortfolioSnapshot)
	if err != nil {
		return nil, err
	}

	end := snapshot.ValuationAsOf.UTC()
	if end.IsZero() {
		end = time.Now().UTC()
	}
	start := end.AddDate(0, 0, -intelligenceLookbackDays)
	seriesByAssetKey := fetchPriceSeriesRange(ctx, s.market, holdings, start, end)
	metrics := computePortfolioMetrics(holdings, seriesByAssetKey)
	alpha30d := computeAlpha30d(ctx, s.market, holdings, metrics, seriesByAssetKey)
	_ = alpha30d

	now := time.Now().UTC()
	var insightRows []Insight
	if err := s.db.DB().WithContext(ctx).
		Where("user_id = ? AND status = ? AND expires_at > ?", userID, "active", now).
		Find(&insightRows).Error; err != nil {
		return nil, err
	}

	calcID, err := s.loadLatestPaidCalculationID(ctx, userID)
	if err != nil {
		return nil, err
	}
	strategyRows := []ReportStrategy{}
	riskSeverity := map[string]string{}
	if calcID != "" {
		strategyRows, err = s.loadReportStrategies(ctx, calcID)
		if err != nil {
			return nil, err
		}
		var riskRows []ReportRisk
		if err := s.db.DB().WithContext(ctx).Where("calculation_id = ?", calcID).Find(&riskRows).Error; err != nil {
			return nil, err
		}
		for _, row := range riskRows {
			riskSeverity[row.RiskID] = row.Severity
		}
	}

	return &intelligenceContext{
		Profile:                 profile,
		Snapshot:                snapshot,
		Holdings:                holdings,
		SeriesByAssetKey:        seriesByAssetKey,
		Metrics:                 metrics,
		ActiveInsights:          insightRows,
		LatestPaidCalculationID: calcID,
		LatestStrategyRows:      strategyRows,
		RiskSeverityByID:        riskSeverity,
	}, nil
}

func (s *Server) buildMarketRegime(ctx context.Context, userID string) (*intelligenceRegimeResponse, error) {
	ctxData, err := s.buildIntelligenceContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	contexts := buildAssetPlanContexts(ctxData.Holdings, ctxData.SeriesByAssetKey, ctxData.Metrics.NonCashPricedValueUSD)
	alpha30d := computeAlpha30d(ctx, s.market, ctxData.Holdings, ctxData.Metrics, ctxData.SeriesByAssetKey)
	breadth, weightedScore := buildRegimeBreadth(contexts)
	regime := classifyRegime(weightedScore)
	trendStrength := classifyTrendStrength(weightedScore)
	signalSummary := summarizeSignalsByAsset(ctxData.ActiveInsights)
	leaders, laggards := s.buildRegimeLeaders(ctx, contexts)
	featured := s.buildFeaturedAssets(ctx, ctxData, contexts, signalSummary)

	response := &intelligenceRegimeResponse{
		AsOf:          ctxData.Snapshot.ValuationAsOf.Format(time.RFC3339),
		Scope:         "active_portfolio",
		Regime:        regime,
		TrendStrength: trendStrength,
		Metrics: intelligenceRegimeMetrics{
			Alpha30d:                alpha30d,
			Volatility30dAnnualized: ctxData.Metrics.Volatility30dAnnualized,
			MaxDrawdown90d:          ctxData.Metrics.MaxDrawdown90d,
			AvgPairwiseCorr:         ctxData.Metrics.AvgPairwiseCorr,
			CashPct:                 ctxData.Metrics.CashPct,
			TopAssetPct:             ctxData.Metrics.TopAssetPct,
			PricedCoveragePct:       ctxData.Metrics.PricedCoveragePct,
		},
		TrendBreadth:    breadth,
		Drivers:         buildRegimeDrivers(breadth, ctxData.Metrics, alpha30d),
		PortfolioImpact: buildRegimePortfolioImpacts(regime, ctxData.Metrics),
		Actions:         buildRegimeActions(regime, ctxData.Metrics, breadth),
		Leaders:         leaders,
		Laggards:        laggards,
		FeaturedAssets:  featured,
	}
	return response, nil
}

func (s *Server) buildAssetBrief(ctx context.Context, userID, assetKey string) (*intelligenceAssetBriefResponse, error) {
	ctxData, err := s.buildIntelligenceContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	portfolioReturns := returnsMap(portfolioReturnsFromIntersection(buildEligibleReturnSeries(ctxData.Holdings, ctxData.SeriesByAssetKey, 20)))
	asset, err := s.resolveIntelligenceAsset(ctx, ctxData, assetKey, portfolioReturns)
	if err != nil {
		return nil, err
	}
	whyNow := buildAssetWhyNow(asset)
	relatedInsights := buildRelatedInsights(ctxData.ActiveInsights, asset.Holding.AssetKey)
	relatedPlans := buildRelatedPlans(ctxData.LatestPaidCalculationID, ctxData.LatestStrategyRows, ctxData.RiskSeverityByID, asset.Holding.AssetKey)

	brief := &intelligenceAssetBriefResponse{
		AsOf:           ctxData.Snapshot.ValuationAsOf.Format(time.RFC3339),
		AssetKey:       asset.Holding.AssetKey,
		Symbol:         asset.Holding.Symbol,
		AssetType:      asset.Holding.AssetType,
		Name:           asset.Name,
		LogoURL:        asset.LogoURL,
		ExchangeMIC:    nullableString(asset.Holding.ExchangeMIC),
		QuoteCurrency:  quoteCurrencyForHolding(asset.Holding),
		CurrentPrice:   roundTo(asset.CurrentPrice, priceDecimals(asset.Holding.AssetType)),
		PriceChange24h: roundTo(asset.Change24h, 4),
		PriceChange7d:  roundTo(asset.Change7d, 4),
		PriceChange30d: roundTo(asset.Change30d, 4),
		ActionBias:     asset.ActionBias,
		SummarySignal:  asset.SummarySignal,
		EntryZone:      asset.EntryZone,
		Invalidation:   asset.Invalidation,
		Technicals: intelligenceAssetTechnicals{
			RSI14:          roundTo(asset.RSI14, 1),
			BollingerUpper: roundTo(asset.BollingerUpper, priceDecimals(asset.Holding.AssetType)),
			BollingerLower: roundTo(asset.BollingerLower, priceDecimals(asset.Holding.AssetType)),
			MA20:           roundTo(asset.MA20, priceDecimals(asset.Holding.AssetType)),
			MA50:           roundTo(asset.MA50, priceDecimals(asset.Holding.AssetType)),
			MA200:          roundTo(asset.MA200, priceDecimals(asset.Holding.AssetType)),
			TrendState:     asset.TrendState,
			TrendStrength:  asset.TrendStrength,
		},
		PortfolioFit: intelligencePortfolioFit{
			IsHeld:              asset.IsHeld,
			WeightPct:           roundTo(asset.WeightPct, 4),
			BetaToPortfolio:     roundTo(asset.BetaToPortfolio, 2),
			Role:                classifyAssetRole(asset.IsHeld, asset.WeightPct, asset.BetaToPortfolio),
			ConcentrationImpact: classifyConcentrationImpact(asset.IsHeld, asset.WeightPct),
			RiskFlag:            classifyAssetRiskFlag(asset.IsHeld, asset.WeightPct, asset.BetaToPortfolio),
		},
		WhyNow:          whyNow,
		RelatedInsights: relatedInsights,
		RelatedPlans:    relatedPlans,
	}
	return brief, nil
}

func buildRegimeBreadth(contexts []assetPlanContext) (intelligenceTrendBreadth, float64) {
	breadth := intelligenceTrendBreadth{}
	if len(contexts) == 0 {
		return breadth, 0
	}
	weighted := 0.0
	weightTotal := 0.0
	for _, item := range contexts {
		score := trendStateScore(item.TrendState)
		weighted += item.AssetWeightPct * score
		weightTotal += item.AssetWeightPct
		switch item.TrendState {
		case "strong_up", "up":
			breadth.UpCount++
		case "strong_down", "down":
			breadth.DownCount++
		default:
			breadth.NeutralCount++
		}
	}
	if weightTotal > 0 {
		breadth.WeightedScore = roundTo(weighted/weightTotal, 2)
	}
	return breadth, breadth.WeightedScore
}

func buildRegimeDrivers(breadth intelligenceTrendBreadth, metrics portfolioMetrics, alpha30d float64) []intelligenceDriver {
	drivers := make([]intelligenceDriver, 0, 4)
	total := breadth.UpCount + breadth.DownCount + breadth.NeutralCount
	if total > 0 {
		upCount := breadth.UpCount
		downCount := breadth.DownCount
		totalCount := total
		drivers = append(drivers, intelligenceDriver{
			ID:         "trend_breadth",
			Kind:       "trend_breadth",
			Tone:       regimeToneFromScore(breadth.WeightedScore),
			ValueText:  fmt.Sprintf("%d/%d assets trending up", breadth.UpCount, total),
			UpCount:    &upCount,
			DownCount:  &downCount,
			TotalCount: &totalCount,
		})
	}
	alphaValue := alpha30d
	drivers = append(drivers, intelligenceDriver{
		ID:        "alpha_30d",
		Kind:      "alpha_30d",
		Tone:      alphaTone(alpha30d),
		ValueText: fmt.Sprintf("30d alpha %+.1f%%", alpha30d*100),
		Value:     &alphaValue,
	})
	vol := metrics.Volatility30dAnnualized
	drivers = append(drivers, intelligenceDriver{
		ID:        "volatility",
		Kind:      "volatility",
		Tone:      volatilityTone(vol),
		ValueText: fmt.Sprintf("Volatility %.0f/100", clamp(vol*100, 0, 100)),
		Value:     &vol,
	})
	corr := metrics.AvgPairwiseCorr
	drivers = append(drivers, intelligenceDriver{
		ID:        "concentration",
		Kind:      "concentration",
		Tone:      concentrationTone(metrics.TopAssetPct),
		ValueText: fmt.Sprintf("Top position %.0f%% of portfolio", metrics.TopAssetPct*100),
		Value:     &metrics.TopAssetPct,
	})
	cash := metrics.CashPct
	drivers = append(drivers, intelligenceDriver{
		ID:        "cash_buffer",
		Kind:      "cash_buffer",
		Tone:      cashTone(metrics.CashPct),
		ValueText: fmt.Sprintf("Cash buffer %.0f%%", metrics.CashPct*100),
		Value:     &cash,
	})
	if corr > 0 {
		drivers = append(drivers, intelligenceDriver{
			ID:        "correlation",
			Kind:      "correlation",
			Tone:      correlationTone(corr),
			ValueText: fmt.Sprintf("Average correlation %.2f", corr),
			Value:     &corr,
		})
	}
	if len(drivers) > 4 {
		drivers = drivers[:4]
	}
	return drivers
}

func buildRegimePortfolioImpacts(regime string, metrics portfolioMetrics) []intelligenceSemanticItem {
	items := make([]intelligenceSemanticItem, 0, 3)
	if metrics.TopAssetPct >= 0.25 {
		items = append(items, intelligenceSemanticItem{ID: "impact_concentration", Kind: "high_concentration"})
	}
	if metrics.AvgPairwiseCorr >= 0.65 {
		items = append(items, intelligenceSemanticItem{ID: "impact_correlation", Kind: "correlated_book"})
	}
	if metrics.CashPct >= 0.15 && metrics.CashPct <= 0.45 {
		items = append(items, intelligenceSemanticItem{ID: "impact_cash", Kind: "healthy_cash_buffer"})
	}
	if len(items) < 3 {
		switch regime {
		case "risk_on":
			items = append(items, intelligenceSemanticItem{ID: "impact_regime", Kind: "mixed_trend"})
		case "risk_off":
			items = append(items, intelligenceSemanticItem{ID: "impact_regime", Kind: "high_beta"})
		default:
			items = append(items, intelligenceSemanticItem{ID: "impact_regime", Kind: "mixed_trend"})
		}
	}
	return uniqueSemanticItems(items)
}

func buildRegimeActions(regime string, metrics portfolioMetrics, breadth intelligenceTrendBreadth) []intelligenceSemanticItem {
	actions := make([]intelligenceSemanticItem, 0, 3)
	if regime == "risk_on" {
		actions = append(actions, intelligenceSemanticItem{ID: "action_pullbacks", Kind: "buy_pullbacks"})
		if metrics.TopAssetPct >= 0.25 {
			actions = append(actions, intelligenceSemanticItem{ID: "action_trim", Kind: "reduce_concentration"})
		}
		if metrics.CashPct < 0.08 {
			actions = append(actions, intelligenceSemanticItem{ID: "action_cash", Kind: "keep_cash_ready"})
		} else {
			actions = append(actions, intelligenceSemanticItem{ID: "action_chasing", Kind: "avoid_chasing"})
		}
	} else if regime == "risk_off" {
		actions = append(actions,
			intelligenceSemanticItem{ID: "action_defense", Kind: "review_defense"},
			intelligenceSemanticItem{ID: "action_trim", Kind: "reduce_concentration"},
			intelligenceSemanticItem{ID: "action_cash", Kind: "keep_cash_ready"},
		)
	} else {
		actions = append(actions,
			intelligenceSemanticItem{ID: "action_wait", Kind: "avoid_chasing"},
			intelligenceSemanticItem{ID: "action_pullbacks", Kind: "buy_pullbacks"},
		)
		if breadth.DownCount > breadth.UpCount {
			actions = append(actions, intelligenceSemanticItem{ID: "action_defense", Kind: "review_defense"})
		}
	}
	return uniqueSemanticItems(actions)
}

func (s *Server) buildRegimeLeaders(ctx context.Context, contexts []assetPlanContext) ([]intelligenceAssetLeader, []intelligenceAssetLeader) {
	if len(contexts) == 0 {
		return nil, nil
	}
	seriesCandidates := make([]intelligenceAssetLeader, 0, len(contexts))
	holdings := make([]portfolioHolding, 0, len(contexts))
	for _, item := range contexts {
		holdings = append(holdings, item.Holding)
	}
	logos := s.resolveHoldingLogos(ctx, holdings)
	for _, item := range contexts {
		change30d, ok := computeSeriesReturn(item.Series, 30)
		if !ok {
			continue
		}
		seriesCandidates = append(seriesCandidates, intelligenceAssetLeader{
			AssetKey:   item.Holding.AssetKey,
			Symbol:     item.Holding.Symbol,
			AssetType:  item.Holding.AssetType,
			Name:       nil,
			LogoURL:    nullableString(logos[item.Holding.AssetKey]),
			Change30d:  roundTo(change30d, 4),
			WeightPct:  roundTo(item.AssetWeightPct, 4),
			TrendState: item.TrendState,
		})
	}
	for i := range seriesCandidates {
		name := s.lookupAssetName(ctx, seriesCandidates[i].AssetKey, seriesCandidates[i].AssetType, seriesCandidates[i].Symbol)
		seriesCandidates[i].Name = nullableString(name)
	}
	return selectDistinctTopMovers(seriesCandidates, 2)
}

func selectDistinctTopMovers(candidates []intelligenceAssetLeader, limit int) ([]intelligenceAssetLeader, []intelligenceAssetLeader) {
	if limit <= 0 || len(candidates) == 0 {
		return nil, nil
	}

	leaders := make([]intelligenceAssetLeader, 0, limit)
	laggards := make([]intelligenceAssetLeader, 0, limit)
	for _, candidate := range candidates {
		switch {
		case candidate.Change30d > 0:
			leaders = append(leaders, candidate)
		case candidate.Change30d < 0:
			laggards = append(laggards, candidate)
		}
	}

	sort.Slice(leaders, func(i, j int) bool {
		if leaders[i].Change30d != leaders[j].Change30d {
			return leaders[i].Change30d > leaders[j].Change30d
		}
		return leaders[i].Symbol < leaders[j].Symbol
	})
	sort.Slice(laggards, func(i, j int) bool {
		if laggards[i].Change30d != laggards[j].Change30d {
			return laggards[i].Change30d < laggards[j].Change30d
		}
		return laggards[i].Symbol < laggards[j].Symbol
	})

	if len(leaders) > limit {
		leaders = leaders[:limit]
	}
	if len(laggards) > limit {
		laggards = laggards[:limit]
	}
	return leaders, laggards
}

func (s *Server) buildFeaturedAssets(ctx context.Context, ctxData *intelligenceContext, contexts []assetPlanContext, signals map[string]assetSignalSummary) []intelligenceFeaturedAsset {
	if len(contexts) == 0 {
		return nil
	}
	portfolioReturns := returnsMap(portfolioReturnsFromIntersection(buildEligibleReturnSeries(ctxData.Holdings, ctxData.SeriesByAssetKey, 20)))
	assets := make([]intelligenceAssetData, 0, len(contexts))
	for _, item := range contexts {
		data, ok := s.buildIntelligenceAssetData(ctx, ctxData, item.Holding, true, portfolioReturns)
		if !ok {
			continue
		}
		if summary, ok := signals[item.Holding.AssetKey]; ok {
			data.SignalCount = summary.count
			if summary.latestSeverity != "" {
				data.LatestSeverity = nullableString(summary.latestSeverity)
			}
		}
		assets = append(assets, data)
	}
	sort.Slice(assets, func(i, j int) bool {
		leftScore := float64(assets[i].SignalCount)*10 + assets[i].WeightPct*100 + absFloat(assets[i].BetaToPortfolio)
		rightScore := float64(assets[j].SignalCount)*10 + assets[j].WeightPct*100 + absFloat(assets[j].BetaToPortfolio)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return assets[i].Holding.Symbol < assets[j].Holding.Symbol
	})
	if len(assets) > 3 {
		assets = assets[:3]
	}
	featured := make([]intelligenceFeaturedAsset, 0, len(assets))
	for _, item := range assets {
		featured = append(featured, intelligenceFeaturedAsset{
			AssetKey:             item.Holding.AssetKey,
			Symbol:               item.Holding.Symbol,
			AssetType:            item.Holding.AssetType,
			Name:                 item.Name,
			LogoURL:              item.LogoURL,
			ActionBias:           item.ActionBias,
			SummarySignal:        item.SummarySignal,
			WeightPct:            roundTo(item.WeightPct, 4),
			BetaToPortfolio:      roundTo(item.BetaToPortfolio, 2),
			SignalCount:          item.SignalCount,
			LatestSignalSeverity: item.LatestSeverity,
		})
	}
	return featured
}

func (s *Server) resolveIntelligenceAsset(ctx context.Context, ctxData *intelligenceContext, assetKey string, portfolioReturns map[int64]float64) (intelligenceAssetData, error) {
	for _, holding := range ctxData.Holdings {
		if holding.AssetKey != assetKey {
			continue
		}
		data, ok := s.buildIntelligenceAssetData(ctx, ctxData, holding, true, portfolioReturns)
		if !ok {
			break
		}
		return data, nil
	}

	stub, ok := s.buildExternalAssetHolding(ctx, assetKey, ctxData.ActiveInsights)
	if !ok {
		return intelligenceAssetData{}, fmt.Errorf("asset not found")
	}
	data, ok := s.buildIntelligenceAssetData(ctx, ctxData, stub, false, portfolioReturns)
	if !ok {
		return intelligenceAssetData{}, fmt.Errorf("asset not found")
	}
	return data, nil
}

func (s *Server) buildIntelligenceAssetData(ctx context.Context, ctxData *intelligenceContext, holding portfolioHolding, isHeld bool, portfolioReturns map[int64]float64) (intelligenceAssetData, bool) {
	series := ctxData.SeriesByAssetKey[holding.AssetKey]
	if len(series) == 0 {
		series = fetchPriceSeriesRange(ctx, s.market, []portfolioHolding{holding}, ctxData.Snapshot.ValuationAsOf.AddDate(0, 0, -intelligenceLookbackDays), ctxData.Snapshot.ValuationAsOf)[holding.AssetKey]
	}
	data, ok := buildIntelligenceCoreData(ctxData.Snapshot, holding, isHeld, series, portfolioReturns)
	if !ok {
		return intelligenceAssetData{}, false
	}
	data.Name = nullableString(s.lookupAssetName(ctx, holding.AssetKey, holding.AssetType, holding.Symbol))
	if logo := s.resolveHoldingLogos(ctx, []portfolioHolding{holding})[holding.AssetKey]; logo != "" {
		data.LogoURL = nullableString(logo)
	}
	return data, true
}

func buildIntelligenceCoreData(
	snapshot PortfolioSnapshot,
	holding portfolioHolding,
	isHeld bool,
	series []ohlcPoint,
	portfolioReturns map[int64]float64,
) (intelligenceAssetData, bool) {
	if len(series) < 20 {
		return intelligenceAssetData{}, false
	}
	closes := extractCloses(series)
	currentPrice := quotePriceForHolding(holding)
	if currentPrice <= 0 {
		currentPrice = closes[len(closes)-1]
	}
	data := intelligenceAssetData{
		Holding:      holding,
		IsHeld:       isHeld,
		Series:       series,
		Closes:       closes,
		CurrentPrice: currentPrice,
	}
	if isHeld && snapshot.NetWorthUSD > 0 {
		data.WeightPct = clamp(holding.ValueUSD/snapshot.NetWorthUSD, 0, 1)
	}
	if change, ok := computeSeriesReturn(series, 1); ok {
		data.Change24h = change
	}
	if change, ok := computeSeriesReturn(series, 7); ok {
		data.Change7d = change
	}
	if change, ok := computeSeriesReturn(series, 30); ok {
		data.Change30d = change
	}
	if value, ok := computeRSI(closes, rsiPeriod); ok {
		data.RSI14 = value
		data.HasRSI14 = true
	}
	if upper, lower, ok := computeBollinger(closes, bollingerPeriod, bollingerStdDev); ok {
		data.BollingerUpper = upper
		data.BollingerLower = lower
		data.HasBollinger = true
	}
	if value, ok := simpleMovingAverage(closes, 20); ok {
		data.MA20 = value
		data.HasMA20 = true
	}
	if value, ok := simpleMovingAverage(closes, 50); ok {
		data.MA50 = value
		data.HasMA50 = true
	}
	if value, ok := simpleMovingAverage(closes, 200); ok {
		data.MA200 = value
		data.HasMA200 = true
	}
	data.TrendState, data.TrendStrength = deriveTrendStateForAsset(data)
	data.BetaToPortfolio = betaToPortfolio(returnsByTimestampFromPoints(series), portfolioReturns)
	data.EntryZone, data.Invalidation = deriveEntryZone(data)
	data.ActionBias, data.SummarySignal = deriveAssetAction(data)
	return data, true
}

func (s *Server) buildExternalAssetHolding(ctx context.Context, assetKey string, insights []Insight) (portfolioHolding, bool) {
	assetType := intelligenceAssetTypeFromKey(assetKey)
	if assetType == "" {
		return portfolioHolding{}, false
	}
	symbol := ""
	for _, insight := range insights {
		if derefString(insight.AssetKey) == assetKey && strings.TrimSpace(insight.Asset) != "" {
			symbol = strings.TrimSpace(insight.Asset)
			break
		}
	}
	holding := portfolioHolding{AssetKey: assetKey, AssetType: assetType, Symbol: symbol, ValuationStatus: "priced"}
	switch assetType {
	case "crypto":
		holding.CoinGeckoID = strings.TrimPrefix(assetKey, "crypto:cg:")
		if holding.CoinGeckoID == "" {
			return portfolioHolding{}, false
		}
		if holding.Symbol == "" {
			list, err := s.market.coinGeckoList(ctx)
			if err == nil {
				for _, coin := range list {
					if coin.ID == holding.CoinGeckoID {
						holding.Symbol = strings.ToUpper(strings.TrimSpace(coin.Symbol))
						break
					}
				}
			}
		}
		applyQuoteMetadata(&holding, assetQuoteMetadata{
			QuoteCurrency: "USD",
			FXRateToUSD:   1,
		})
	case "stock":
		parts := strings.Split(assetKey, ":")
		if len(parts) >= 4 {
			holding.ExchangeMIC = parts[len(parts)-2]
			holding.Symbol = parts[len(parts)-1]
		}
		if holding.Symbol == "" {
			return portfolioHolding{}, false
		}
		applyQuoteMetadata(&holding, assetQuoteMetadata{
			QuoteCurrency: s.resolveStockQuoteCurrency(ctx, assetKey, holding.Symbol),
		})
	case "forex":
		parts := strings.Split(assetKey, ":")
		if len(parts) >= 3 {
			holding.Symbol = parts[len(parts)-1]
		}
		applyQuoteMetadata(&holding, assetQuoteMetadata{
			QuoteCurrency:     normalizeCurrency(holding.Symbol),
			CurrentPriceQuote: 1,
			FXRateToUSD:       holding.CurrentPrice,
		})
	}
	return holding, holding.Symbol != ""
}

func (s *Server) lookupAssetName(ctx context.Context, assetKey, assetType, symbol string) string {
	switch assetType {
	case "crypto":
		coinID := strings.TrimPrefix(assetKey, "crypto:cg:")
		if coinID == "" {
			return ""
		}
		list, err := s.market.coinGeckoList(ctx)
		if err != nil {
			return ""
		}
		for _, coin := range list {
			if coin.ID == coinID {
				return strings.TrimSpace(coin.Name)
			}
		}
	case "stock":
		ticker, err := s.market.marketstackTicker(ctx, symbol)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(ticker.Name)
	case "forex":
		currencies, err := s.market.openExchangeCurrencies(ctx)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(currencies[strings.ToUpper(strings.TrimSpace(symbol))])
	}
	return ""
}

func buildRelatedInsights(insights []Insight, assetKey string) []intelligenceRelatedInsight {
	items := make([]intelligenceRelatedInsight, 0)
	for _, insight := range insights {
		if derefString(insight.AssetKey) != assetKey {
			continue
		}
		items = append(items, intelligenceRelatedInsight{
			ID:            insight.ID,
			Type:          insight.Type,
			Severity:      insight.Severity,
			TriggerReason: insight.TriggerReason,
			CreatedAt:     insight.CreatedAt.Format(time.RFC3339),
			PlanID:        insight.PlanID,
			StrategyID:    insight.StrategyID,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	if len(items) > 3 {
		items = items[:3]
	}
	return items
}

func buildRelatedPlans(calculationID string, strategyRows []ReportStrategy, riskSeverity map[string]string, assetKey string) []intelligenceRelatedPlan {
	if calculationID == "" {
		return nil
	}
	items := make([]intelligenceRelatedPlan, 0)
	for _, row := range strategyRows {
		if row.AssetKey != assetKey {
			continue
		}
		items = append(items, intelligenceRelatedPlan{
			CalculationID:   calculationID,
			PlanID:          row.PlanID,
			StrategyID:      row.StrategyID,
			Priority:        planPriorityFromSeverity(riskSeverity[row.LinkedRiskID]),
			Rationale:       row.Rationale,
			ExpectedOutcome: row.ExpectedOutcome,
		})
	}
	return items
}

func summarizeSignalsByAsset(insights []Insight) map[string]assetSignalSummary {
	summary := make(map[string]assetSignalSummary)
	for _, insight := range insights {
		assetKey := derefString(insight.AssetKey)
		if assetKey == "" {
			continue
		}
		item := summary[assetKey]
		item.count++
		if item.latestSeverity == "" || insightSeverityRank(insight.Severity) < insightSeverityRank(item.latestSeverity) {
			item.latestSeverity = insight.Severity
		}
		summary[assetKey] = item
	}
	return summary
}

func deriveTrendStateForAsset(asset intelligenceAssetData) (string, string) {
	if asset.HasMA20 && asset.HasMA50 && asset.HasMA200 {
		return computeTrendState(asset.CurrentPrice, asset.MA20, asset.MA50, asset.MA200)
	}
	if asset.HasMA20 && asset.HasMA50 {
		if asset.CurrentPrice > asset.MA20 && asset.MA20 > asset.MA50 {
			return "up", "medium"
		}
		if asset.CurrentPrice < asset.MA20 && asset.MA20 < asset.MA50 {
			return "down", "medium"
		}
	}
	return "neutral", "weak"
}

func deriveEntryZone(asset intelligenceAssetData) (intelligencePriceZone, intelligenceInvalidation) {
	precision := priceDecimals(asset.Holding.AssetType)
	support := minSeriesLow(asset.Series, 20)
	entryLow := support
	entryHigh := support
	basis := "support_only"
	if asset.HasMA20 && asset.HasMA50 {
		entryLow = minFloat(asset.MA20, asset.MA50)
		entryHigh = maxFloat(asset.MA20, asset.MA50)
		basis = "support_and_ma20"
	}
	if entryLow <= 0 || entryHigh <= 0 {
		entryLow = asset.CurrentPrice * 0.97
		entryHigh = asset.CurrentPrice * 1.01
	}
	if entryLow > entryHigh {
		entryLow, entryHigh = entryHigh, entryLow
	}
	invalidationPrice := minFloat(support*0.98, entryLow*0.98)
	if invalidationPrice <= 0 {
		invalidationPrice = entryLow * 0.98
	}
	return intelligencePriceZone{
			Low:   roundTo(entryLow, precision),
			High:  roundTo(entryHigh, precision),
			Basis: basis,
		}, intelligenceInvalidation{
			Price:  roundTo(invalidationPrice, precision),
			Reason: "break_below_support",
		}
}

func deriveAssetAction(asset intelligenceAssetData) (string, string) {
	nearEntry := asset.CurrentPrice >= asset.EntryZone.Low*0.98 && asset.CurrentPrice <= asset.EntryZone.High*1.02
	if asset.TrendState == "strong_down" || asset.TrendState == "down" {
		if asset.IsHeld {
			return "reduce", "downtrend_risk"
		}
		return "wait", "downtrend_risk"
	}
	if asset.HasRSI14 && asset.RSI14 <= 35 && nearEntry {
		return "accumulate", "trend_up_pullback"
	}
	if asset.TrendState == "strong_up" || asset.TrendState == "up" {
		if asset.HasRSI14 && asset.RSI14 >= 65 {
			if asset.IsHeld {
				return "hold", "overextended_uptrend"
			}
			return "wait", "overextended_uptrend"
		}
		if nearEntry {
			return "accumulate", "trend_up_pullback"
		}
		if asset.IsHeld {
			return "hold", "trend_up_pullback"
		}
		return "wait", "trend_up_pullback"
	}
	return "wait", "neutral_range"
}

func buildAssetWhyNow(asset intelligenceAssetData) []intelligenceSemanticItem {
	items := make([]intelligenceSemanticItem, 0, 3)
	nearEntry := asset.CurrentPrice >= asset.EntryZone.Low*0.98 && asset.CurrentPrice <= asset.EntryZone.High*1.02
	if nearEntry {
		items = append(items, intelligenceSemanticItem{ID: "why_entry_zone", Kind: "near_entry_zone"})
	}
	if asset.TrendState == "up" || asset.TrendState == "strong_up" {
		items = append(items, intelligenceSemanticItem{ID: "why_trend_up", Kind: "above_ma20_ma50"})
	}
	if asset.TrendState == "down" || asset.TrendState == "strong_down" {
		items = append(items, intelligenceSemanticItem{ID: "why_trend_down", Kind: "below_ma50"})
	}
	if asset.HasRSI14 && asset.RSI14 <= 35 {
		items = append(items, intelligenceSemanticItem{ID: "why_rsi_low", Kind: "rsi_oversold"})
	}
	if asset.HasRSI14 && asset.RSI14 >= 70 {
		items = append(items, intelligenceSemanticItem{ID: "why_rsi_hot", Kind: "rsi_hot"})
	}
	if asset.IsHeld && asset.WeightPct >= 0.20 {
		items = append(items, intelligenceSemanticItem{ID: "why_overweight", Kind: "portfolio_overweight"})
	}
	if asset.IsHeld && asset.WeightPct > 0 && asset.WeightPct <= 0.05 {
		items = append(items, intelligenceSemanticItem{ID: "why_underweight", Kind: "portfolio_underweight"})
	}
	items = uniqueSemanticItems(items)
	if len(items) > 3 {
		items = items[:3]
	}
	return items
}

func computeSeriesReturn(series []ohlcPoint, lookback int) (float64, bool) {
	if len(series) < 2 {
		return 0, false
	}
	closes := extractCloses(series)
	if len(closes) < 2 {
		return 0, false
	}
	window := lookback
	if len(closes)-1 < window {
		window = len(closes) - 1
	}
	if window <= 0 {
		return 0, false
	}
	start := closes[len(closes)-1-window]
	end := closes[len(closes)-1]
	if start <= 0 || end <= 0 {
		return 0, false
	}
	return (end / start) - 1, true
}

func minSeriesLow(series []ohlcPoint, lookback int) float64 {
	if len(series) == 0 {
		return 0
	}
	start := 0
	if len(series) > lookback {
		start = len(series) - lookback
	}
	minValue := series[start].Low
	for _, point := range series[start:] {
		if point.Low < minValue {
			minValue = point.Low
		}
	}
	return minValue
}

func intelligenceAssetTypeFromKey(assetKey string) string {
	switch {
	case strings.HasPrefix(assetKey, "crypto:"):
		return "crypto"
	case strings.HasPrefix(assetKey, "stock:"):
		return "stock"
	case strings.HasPrefix(assetKey, "forex:"):
		return "forex"
	default:
		return ""
	}
}

func classifyRegime(weightedScore float64) string {
	switch {
	case weightedScore >= 0.45:
		return "risk_on"
	case weightedScore <= -0.45:
		return "risk_off"
	default:
		return "neutral"
	}
}

func classifyTrendStrength(weightedScore float64) string {
	value := absFloat(weightedScore)
	switch {
	case value >= 1.0:
		return "strong"
	case value >= 0.45:
		return "medium"
	default:
		return "weak"
	}
}

func trendStateScore(state string) float64 {
	switch state {
	case "strong_up":
		return 2
	case "up":
		return 1
	case "down":
		return -1
	case "strong_down":
		return -2
	default:
		return 0
	}
}

func regimeToneFromScore(value float64) string {
	if value >= 0.45 {
		return "positive"
	}
	if value <= -0.45 {
		return "caution"
	}
	return "neutral"
}

func alphaTone(value float64) string {
	if value >= 0.03 {
		return "positive"
	}
	if value <= -0.03 {
		return "caution"
	}
	return "neutral"
}

func volatilityTone(value float64) string {
	if value >= 0.55 {
		return "caution"
	}
	if value <= 0.25 {
		return "positive"
	}
	return "neutral"
}

func concentrationTone(value float64) string {
	if value >= 0.25 {
		return "caution"
	}
	if value <= 0.12 {
		return "positive"
	}
	return "neutral"
}

func cashTone(value float64) string {
	if value >= 0.10 && value <= 0.40 {
		return "positive"
	}
	if value < 0.05 || value > 0.50 {
		return "caution"
	}
	return "neutral"
}

func correlationTone(value float64) string {
	if value >= 0.70 {
		return "caution"
	}
	if value <= 0.35 {
		return "positive"
	}
	return "neutral"
}

func classifyAssetRole(isHeld bool, weightPct, beta float64) string {
	if !isHeld {
		return "watchlist"
	}
	if weightPct >= 0.20 {
		return "core"
	}
	if beta >= 1.2 {
		return "satellite"
	}
	return "tactical"
}

func classifyConcentrationImpact(isHeld bool, weightPct float64) string {
	if !isHeld {
		return "limited"
	}
	if weightPct >= 0.20 {
		return "high"
	}
	if weightPct >= 0.10 {
		return "moderate"
	}
	return "limited"
}

func classifyAssetRiskFlag(isHeld bool, weightPct, beta float64) string {
	if !isHeld {
		return "watchlist_only"
	}
	if weightPct >= 0.20 {
		return "high_concentration"
	}
	if beta >= 1.2 {
		return "high_beta"
	}
	return "balanced"
}

func uniqueSemanticItems(items []intelligenceSemanticItem) []intelligenceSemanticItem {
	seen := make(map[string]struct{}, len(items))
	result := make([]intelligenceSemanticItem, 0, len(items))
	for _, item := range items {
		if item.Kind == "" {
			continue
		}
		if _, ok := seen[item.Kind]; ok {
			continue
		}
		seen[item.Kind] = struct{}{}
		result = append(result, item)
	}
	return result
}

func minFloat(left, right float64) float64 {
	if left == 0 {
		return right
	}
	if right == 0 {
		return left
	}
	if left < right {
		return left
	}
	return right
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
