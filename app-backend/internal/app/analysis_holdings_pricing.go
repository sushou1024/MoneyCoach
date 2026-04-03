package app

func applyPricing(holding *portfolioHolding, priceMap map[string]coinGeckoSimplePrice, stockPrices map[string]marketstackPriceQuote, oerRates map[string]float64) {
	if holding == nil {
		return
	}
	applyAvgPriceConversion(holding, oerRates)
	if applyStablecoinPricing(holding) {
		return
	}
	if applyCryptoPricing(holding, priceMap) {
		return
	}
	if applyStockPricing(holding, stockPrices, oerRates) {
		return
	}
	if applyForexPricing(holding, oerRates) {
		return
	}
	if applyManualValuePricing(holding) {
		return
	}
	if applyScreenshotPricing(holding, oerRates) {
		return
	}

	holding.ValuationStatus = "unpriced"
}

func applyAvgPriceConversion(holding *portfolioHolding, oerRates map[string]float64) {
	if holding.AvgPrice == nil {
		return
	}
	if holding.AvgPriceSource == "user_input" {
		return
	}
	if converted, ok := convertDisplayPriceToUSD(*holding.AvgPrice, holding.DisplayCurrency, oerRates); ok {
		holding.AvgPrice = &converted
		if holding.AvgPriceSource == "" {
			holding.AvgPriceSource = "provided"
		}
		return
	}
	holding.AvgPrice = nil
}

func applyCostBasis(holding *portfolioHolding) {
	if holding.AvgPriceSource == "derived_from_market" {
		holding.CostBasisStatus = "unknown"
		holding.PNLPercent = nil
		return
	}
	if holding.AvgPrice == nil || *holding.AvgPrice <= 0 {
		if holding.PNLPercent != nil && holding.CurrentPrice > 0 {
			if 1+*holding.PNLPercent != 0 {
				derived := holding.CurrentPrice / (1 + *holding.PNLPercent)
				holding.AvgPrice = &derived
				holding.AvgPriceSource = "derived_from_pnl_percent"
				holding.CostBasisStatus = "provided"
			}
		}
		return
	}
	if holding.AvgPriceSource == "" {
		holding.AvgPriceSource = "provided"
	}
	if holding.CurrentPrice > 0 {
		pnl := (holding.CurrentPrice - *holding.AvgPrice) / *holding.AvgPrice
		holding.PNLPercent = &pnl
	}
	switch holding.AvgPriceSource {
	case "provided", "user_input", "derived_from_pnl_percent":
		holding.CostBasisStatus = "provided"
	default:
		holding.CostBasisStatus = "unknown"
	}
}
