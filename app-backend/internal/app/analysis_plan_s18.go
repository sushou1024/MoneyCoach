package app

func buildS18Plan(ctx assetPlanContext) (lockedPlan, bool) {
	if ctx.TrendState == "" || ctx.TrendState == "neutral" {
		return lockedPlan{}, false
	}
	trendAction := "wait"
	switch ctx.TrendState {
	case "strong_up":
		trendAction = "hold_or_add"
	case "up":
		trendAction = "hold"
	case "down", "strong_down":
		trendAction = "reduce_exposure"
	}

	params := map[string]any{
		"trend_state":    ctx.TrendState,
		"trend_strength": ctx.TrendStrength,
		"trend_action":   trendAction,
		"ma_short":       20,
		"ma_medium":      50,
		"ma_long":        200,
		"current_price":  roundTo(ctx.Holding.CurrentPrice, priceDecimals(ctx.Holding.AssetType)),
		"ma_20":          roundTo(ctx.MA20, priceDecimals(ctx.Holding.AssetType)),
		"ma_50":          roundTo(ctx.MA50, priceDecimals(ctx.Holding.AssetType)),
		"ma_200":         roundTo(ctx.MA200, priceDecimals(ctx.Holding.AssetType)),
	}

	return lockedPlan{
		StrategyID: "S18",
		AssetType:  ctx.Holding.AssetType,
		Symbol:     ctx.Holding.Symbol,
		AssetKey:   ctx.Holding.AssetKey,
		Parameters: params,
	}, true
}
