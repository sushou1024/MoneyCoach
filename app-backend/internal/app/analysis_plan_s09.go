package app

import "math"

func buildS09Plan(riskLevel string, metrics portfolioMetrics, ctx assetPlanContext) lockedPlan {
	profitStep := s09ProfitStepPct(riskLevel)
	pnl := 0.0
	if ctx.Holding.PNLPercent != nil {
		pnl = *ctx.Holding.PNLPercent
	}
	maxAdditions := int(math.Floor(pnl / profitStep))
	baseAddition := clamp(metrics.IdleCashUSD*0.10, 20, math.Min(2000, metrics.IdleCashUSD))

	additions := make([]map[string]any, 0, maxAdditions)
	for i := 1; i <= maxAdditions; i++ {
		additionAmount := roundTo(baseAddition*math.Pow(1.2, float64(i-1)), 2)
		additions = append(additions, map[string]any{
			"addition_number":     i,
			"trigger_profit_pct":  roundTo(profitStep*float64(i), 4),
			"addition_amount_usd": additionAmount,
		})
	}

	return lockedPlan{
		StrategyID: "S09",
		AssetType:  ctx.Holding.AssetType,
		Symbol:     ctx.Holding.Symbol,
		AssetKey:   ctx.Holding.AssetKey,
		Parameters: map[string]any{
			"profit_step_pct":   roundTo(profitStep, 4),
			"max_additions":     maxAdditions,
			"base_addition_usd": roundTo(baseAddition, 2),
			"additions":         additions,
		},
	}
}

func s09ProfitStepPct(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.15
	case "aggressive":
		return 0.05
	default:
		return 0.10
	}
}
