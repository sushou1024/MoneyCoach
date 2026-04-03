package app

import "strings"

func buildS01Plan(profile userProfile, riskLevel string, metrics portfolioMetrics, ctx assetPlanContext) lockedPlan {
	baseSL := baseStopLossPct(riskLevel)
	volAdj := 1.0
	if metrics.Volatility30dDaily > 0.06 {
		volAdj = 1.5
	} else if metrics.Volatility30dDaily > 0.04 {
		volAdj = 1.2
	}

	experienceAdj := 1.0
	switch strings.ToLower(profile.Experience) {
	case "beginner":
		experienceAdj = 0.8
	case "expert":
		experienceAdj = 1.2
	}

	lossAdj := 1.0
	if ctx.Holding.PNLPercent != nil && *ctx.Holding.PNLPercent < -0.10 {
		lossAdj = 1.2
	}

	stopLossPct := clamp(baseSL*volAdj*experienceAdj*lossAdj, 0.03, 0.15)
	provisional := ctx.Holding.CurrentPrice * (1 - stopLossPct)
	if ctx.Holding.AvgPrice != nil {
		if ctx.Holding.PNLPercent != nil && *ctx.Holding.PNLPercent < 0 {
			stopFromCost := *ctx.Holding.AvgPrice * (1 - stopLossPct)
			if stopFromCost < provisional {
				provisional = stopFromCost
			}
		} else {
			provisional = *ctx.Holding.AvgPrice * (1 - stopLossPct)
		}
	}

	params := map[string]any{
		"stop_loss_pct":   roundTo(stopLossPct, 4),
		"stop_loss_price": roundTo(provisional, priceDecimals(ctx.Holding.AssetType)),
	}

	supportPoints := ctx.Series
	if len(supportPoints) > defaultSupportLookback {
		supportPoints = supportPoints[len(supportPoints)-defaultSupportLookback:]
	}
	if support, ok := closestSupportBelow(supportPoints, provisional); ok {
		if provisional > support*0.95 {
			adjusted := roundTo(support*0.98, priceDecimals(ctx.Holding.AssetType))
			params["stop_loss_price"] = adjusted
			params["adjustment_reason"] = "Adjusted to support level"
			params["support_level"] = roundTo(support, priceDecimals(ctx.Holding.AssetType))
		} else {
			params["adjustment_reason"] = "Risk/volatility-based"
		}
	} else {
		params["adjustment_reason"] = "Risk/volatility-based"
	}

	return lockedPlan{
		StrategyID: "S01",
		AssetType:  ctx.Holding.AssetType,
		Symbol:     ctx.Holding.Symbol,
		AssetKey:   ctx.Holding.AssetKey,
		Parameters: params,
	}
}
