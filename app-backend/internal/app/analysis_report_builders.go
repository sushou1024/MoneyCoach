package app

func buildPreviewPayload(calculationID, valuationAsOf, marketDataSnapshotID string, profile userProfile, holdings []portfolioHolding, metrics portfolioMetrics, fixed previewFixedMetrics, netWorthDisplay float64, baseCurrency string, baseFXRateToUSD float64) previewPayload {
	return previewPayload{
		MetaData:             metaDataPayload{CalculationID: calculationID},
		ValuationAsOf:        valuationAsOf,
		MarketDataSnapshotID: marketDataSnapshotID,
		UserProfile:          buildUserProfilePayload(profile),
		Portfolio:            buildPortfolioPayload(holdings, metrics.NetWorthUSD),
		ComputedMetrics: computedMetricsPayload{
			NetWorthUSD:             metrics.NetWorthUSD,
			CashPct:                 metrics.CashPct,
			TopAssetPct:             metrics.TopAssetPct,
			Volatility30dAnnualized: metrics.Volatility30dAnnualized,
			MaxDrawdown90d:          metrics.MaxDrawdown90d,
			AvgPairwiseCorr:         metrics.AvgPairwiseCorr,
			HealthScoreBaseline:     metrics.HealthScoreBaseline,
			VolatilityScoreBaseline: metrics.VolatilityScoreBaseline,
			PricedCoveragePct:       metrics.PricedCoveragePct,
			MetricsIncomplete:       metrics.MetricsIncomplete,
		},
		FixedMetrics:    fixed,
		NetWorthDisplay: netWorthDisplay,
		BaseCurrency:    baseCurrency,
		BaseFXRateToUSD: baseFXRateToUSD,
	}
}

func buildPaidPayload(profile userProfile, holdings []portfolioHolding, preview previewReportPayload, plans []lockedPlan, metrics portfolioMetrics, fixed previewFixedMetrics, netWorthDisplay float64, baseCurrency string, baseFXRateToUSD float64) paidPayload {
	return paidPayload{
		UserPortfolio:  buildPortfolioPayload(holdings, metrics.NetWorthUSD),
		UserProfile:    buildUserProfilePayload(profile),
		PreviousTeaser: preview,
		LockedPlans:    plans,
		PortfolioFacts: portfolioFactsPayload{
			NetWorthUSD:             metrics.NetWorthUSD,
			CashPct:                 metrics.CashPct,
			TopAssetPct:             metrics.TopAssetPct,
			Volatility30dAnnualized: metrics.Volatility30dAnnualized,
			MaxDrawdown90d:          metrics.MaxDrawdown90d,
			AvgPairwiseCorr:         metrics.AvgPairwiseCorr,
			PricedCoveragePct:       metrics.PricedCoveragePct,
			MetricsIncomplete:       metrics.MetricsIncomplete,
		},
		FixedMetrics:    fixed,
		NetWorthDisplay: netWorthDisplay,
		BaseCurrency:    baseCurrency,
		BaseFXRateToUSD: baseFXRateToUSD,
	}
}

func buildDirectPaidPayload(calculationID, valuationAsOf, marketDataSnapshotID string, profile userProfile, holdings []portfolioHolding, metrics portfolioMetrics, risks []previewRisk, plans []lockedPlan, fixed previewFixedMetrics, netWorthDisplay float64, baseCurrency string, baseFXRateToUSD float64) directPaidPayload {
	return directPaidPayload{
		MetaData:             metaDataPayload{CalculationID: calculationID},
		ValuationAsOf:        valuationAsOf,
		MarketDataSnapshotID: marketDataSnapshotID,
		UserPortfolio:        buildPortfolioPayload(holdings, metrics.NetWorthUSD),
		UserProfile:          buildUserProfilePayload(profile),
		ComputedMetrics: computedMetricsPayload{
			NetWorthUSD:             metrics.NetWorthUSD,
			CashPct:                 metrics.CashPct,
			TopAssetPct:             metrics.TopAssetPct,
			Volatility30dAnnualized: metrics.Volatility30dAnnualized,
			MaxDrawdown90d:          metrics.MaxDrawdown90d,
			AvgPairwiseCorr:         metrics.AvgPairwiseCorr,
			HealthScoreBaseline:     metrics.HealthScoreBaseline,
			VolatilityScoreBaseline: metrics.VolatilityScoreBaseline,
			PricedCoveragePct:       metrics.PricedCoveragePct,
			MetricsIncomplete:       metrics.MetricsIncomplete,
		},
		IdentifiedRisks: risks,
		LockedPlans:     plans,
		PortfolioFacts: portfolioFactsPayload{
			NetWorthUSD:             metrics.NetWorthUSD,
			CashPct:                 metrics.CashPct,
			TopAssetPct:             metrics.TopAssetPct,
			Volatility30dAnnualized: metrics.Volatility30dAnnualized,
			MaxDrawdown90d:          metrics.MaxDrawdown90d,
			AvgPairwiseCorr:         metrics.AvgPairwiseCorr,
			PricedCoveragePct:       metrics.PricedCoveragePct,
			MetricsIncomplete:       metrics.MetricsIncomplete,
		},
		FixedMetrics:    fixed,
		NetWorthDisplay: netWorthDisplay,
		BaseCurrency:    baseCurrency,
		BaseFXRateToUSD: baseFXRateToUSD,
	}
}

func buildPortfolioPayload(holdings []portfolioHolding, netWorth float64) portfolioPayload {
	payload := portfolioPayload{NetWorthUSD: netWorth}
	payload.Holdings = make([]portfolioHoldingPayload, 0, len(holdings))
	for _, holding := range holdings {
		payload.Holdings = append(payload.Holdings, portfolioHoldingPayload{
			AssetKey:  holding.AssetKey,
			Symbol:    holding.Symbol,
			AssetType: holding.AssetType,
			Amount:    holding.Amount,
			ValueUSD:  holding.ValueUSD,
		})
	}
	return payload
}

func buildUserProfilePayload(profile userProfile) userProfilePayload {
	return userProfilePayload{
		RiskTolerance:  profile.RiskTolerance,
		RiskPreference: profile.RiskPreference,
		PainPoints:     profile.PainPoints,
		Experience:     profile.Experience,
		Style:          profile.Style,
		Markets:        profile.Markets,
	}
}

func previewFromPromptOutput(llm previewPromptOutput) previewReportPayload {
	return previewReportPayload{
		MetaData:             llm.MetaData,
		ValuationAsOf:        llm.ValuationAsOf,
		MarketDataSnapshotID: llm.MarketDataSnapshotID,
		FixedMetrics:         llm.FixedMetrics,
		NetWorthDisplay:      llm.NetWorthDisplay,
		BaseCurrency:         llm.BaseCurrency,
		BaseFXRateToUSD:      llm.BaseFXRateToUSD,
		IdentifiedRisks:      llm.IdentifiedRisks,
		LockedProjection:     llm.LockedProjection,
	}
}

func paidFromPromptOutput(llm paidPromptOutput) paidReportPayload {
	return paidReportPayload{
		MetaData:             llm.MetaData,
		ValuationAsOf:        llm.ValuationAsOf,
		MarketDataSnapshotID: llm.MarketDataSnapshotID,
		NetWorthDisplay:      llm.NetWorthDisplay,
		BaseCurrency:         llm.BaseCurrency,
		BaseFXRateToUSD:      llm.BaseFXRateToUSD,
		ReportHeader:         llm.ReportHeader,
		Charts:               llm.Charts,
		RiskInsights:         llm.RiskInsights,
		OptimizationPlan:     llm.OptimizationPlan,
		TheVerdict:           llm.TheVerdict,
	}
}

func normalizePreviewStatus(preview *previewReportPayload) {
	if preview == nil {
		return
	}
	preview.FixedMetrics.HealthStatus = healthStatusFromScore(preview.FixedMetrics.HealthScore)
}

func mergePaidPlanParameters(paid *paidReportPayload, plans []lockedPlan) {
	if paid == nil {
		return
	}
	planByID := make(map[string]lockedPlan)
	for _, plan := range plans {
		planByID[plan.PlanID] = plan
	}
	for i := range paid.OptimizationPlan {
		if plan, ok := planByID[paid.OptimizationPlan[i].PlanID]; ok {
			paid.OptimizationPlan[i].Parameters = plan.Parameters
			paid.OptimizationPlan[i].LinkedRiskID = plan.LinkedRiskID
			paid.OptimizationPlan[i].AssetKey = plan.AssetKey
			paid.OptimizationPlan[i].AssetType = plan.AssetType
			paid.OptimizationPlan[i].Symbol = plan.Symbol
			paid.OptimizationPlan[i].QuoteCurrency = plan.QuoteCurrency
		}
	}
	for i := range paid.ActionableAdvice {
		if plan, ok := planByID[paid.ActionableAdvice[i].PlanID]; ok {
			paid.ActionableAdvice[i].Parameters = plan.Parameters
			paid.ActionableAdvice[i].LinkedRiskID = plan.LinkedRiskID
			paid.ActionableAdvice[i].AssetKey = plan.AssetKey
			paid.ActionableAdvice[i].AssetType = plan.AssetType
			paid.ActionableAdvice[i].Symbol = plan.Symbol
			paid.ActionableAdvice[i].QuoteCurrency = plan.QuoteCurrency
		}
	}
}
