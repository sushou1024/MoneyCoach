package app

func buildS04Plan(riskLevel string, ctx assetPlanContext) lockedPlan {
	if ctx.Holding.AvgPrice == nil {
		return lockedPlan{}
	}

	var layerConfigs []struct {
		Name       string
		SellPct    float64
		Multiplier float64
	}
	switch riskLevel {
	case "conservative":
		layerConfigs = []struct {
			Name       string
			SellPct    float64
			Multiplier float64
		}{{"Layer 1", 0.40, 1.30}, {"Layer 2", 0.35, 1.50}, {"Layer 3", 0.25, 1.80}}
	case "aggressive":
		layerConfigs = []struct {
			Name       string
			SellPct    float64
			Multiplier float64
		}{{"Layer 1", 0.20, 1.50}, {"Layer 2", 0.30, 2.00}, {"Layer 3", 0.50, 3.00}}
	default:
		layerConfigs = []struct {
			Name       string
			SellPct    float64
			Multiplier float64
		}{{"Layer 1", 0.30, 1.40}, {"Layer 2", 0.40, 1.70}, {"Layer 3", 0.30, 2.00}}
	}

	layers := make([]map[string]any, 0, len(layerConfigs))
	for _, cfg := range layerConfigs {
		targetPrice := *ctx.Holding.AvgPrice * cfg.Multiplier
		if ctx.Holding.CurrentPrice >= targetPrice {
			targetPrice = ctx.Holding.CurrentPrice * 1.05
		}
		sellAmount := ctx.Holding.Amount * cfg.SellPct
		expectedProfit := (targetPrice - *ctx.Holding.AvgPrice) * sellAmount
		layers = append(layers, map[string]any{
			"layer_name":          cfg.Name,
			"sell_percentage":     roundTo(cfg.SellPct, 2),
			"sell_amount":         roundTo(sellAmount, amountDecimals(ctx.Holding.AssetType)),
			"target_price":        roundTo(targetPrice, priceDecimals(ctx.Holding.AssetType)),
			"expected_profit_usd": roundTo(expectedProfit, 2),
		})
	}

	return lockedPlan{
		StrategyID: "S04",
		AssetType:  ctx.Holding.AssetType,
		Symbol:     ctx.Holding.Symbol,
		AssetKey:   ctx.Holding.AssetKey,
		Parameters: map[string]any{"layers": layers},
	}
}
