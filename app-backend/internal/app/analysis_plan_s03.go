package app

func buildS03Plan(riskLevel string, ctx assetPlanContext) (lockedPlan, bool) {
	if ctx.Holding.CurrentPrice <= 0 || ctx.Holding.PNLPercent == nil {
		return lockedPlan{}, false
	}
	trendStrength := "weak"
	if ctx.HasMA20 && ctx.HasMA50 {
		if ctx.Holding.CurrentPrice > ctx.MA20 && ctx.MA20 > ctx.MA50 {
			trendStrength = "medium"
			if ctx.HasMA200 && ctx.MA50 > ctx.MA200 {
				trendStrength = "strong"
			}
		}
	}
	basePct := s03BaseStopPct(riskLevel)
	volAdj := 1.0
	if ctx.Volatility30dDaily > 0.05 {
		volAdj = 1.3
	} else if ctx.Volatility30dDaily < 0.02 {
		volAdj = 0.8
	}
	trendAdj := 0.8
	switch trendStrength {
	case "strong":
		trendAdj = 1.2
	case "medium":
		trendAdj = 1.0
	}
	profitAdj := 1.0
	if *ctx.Holding.PNLPercent > 0.50 {
		profitAdj = 1.3
	} else if *ctx.Holding.PNLPercent > 0.30 {
		profitAdj = 1.1
	}
	trailingStopPct := clamp(basePct*volAdj*trendAdj*profitAdj, 0.05, 0.25)

	activationPrice := ctx.Holding.CurrentPrice
	callbackRate := trailingStopPct
	initialStop := ctx.Holding.CurrentPrice * (1 - trailingStopPct)

	params := map[string]any{
		"trailing_stop_pct":           roundTo(trailingStopPct, 4),
		"activation_price":            roundTo(activationPrice, priceDecimals(ctx.Holding.AssetType)),
		"callback_rate":               roundTo(callbackRate, 4),
		"initial_trailing_stop_price": roundTo(initialStop, priceDecimals(ctx.Holding.AssetType)),
	}

	return lockedPlan{
		StrategyID: "S03",
		AssetType:  ctx.Holding.AssetType,
		Symbol:     ctx.Holding.Symbol,
		AssetKey:   ctx.Holding.AssetKey,
		Parameters: params,
	}, true
}

func s03BaseStopPct(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.10
	case "aggressive":
		return 0.07
	default:
		return 0.08
	}
}
