package app

func applyManualValuePricing(holding *portfolioHolding) bool {
	if holding.ManualValueUSD == nil || *holding.ManualValueUSD <= 0 {
		return false
	}
	holding.ValueUSD = *holding.ManualValueUSD
	holding.ValuationStatus = "user_provided"
	holding.PricingSource = "USER_PROVIDED"
	currency, ok := normalizeDisplayCurrency(holding.DisplayCurrency)
	if !ok {
		currency = quoteCurrencyForHolding(*holding)
	}
	fxRateToUSD := 1.0
	if currency != "USD" && !stablecoinSet()[currency] {
		fxRateToUSD = 0
	}
	currentPriceQuote := 0.0
	if holding.Amount > 0 && fxRateToUSD > 0 {
		currentPriceQuote = (*holding.ManualValueUSD / fxRateToUSD) / holding.Amount
	}
	applyQuoteMetadata(holding, assetQuoteMetadata{
		QuoteCurrency:     currency,
		CurrentPriceQuote: currentPriceQuote,
		FXRateToUSD:       fxRateToUSD,
	})
	return true
}

func applyScreenshotPricing(holding *portfolioHolding, oerRates map[string]float64) bool {
	if holding.ValueFromScreenshot == nil {
		return false
	}
	currency, ok := normalizeDisplayCurrency(holding.DisplayCurrency)
	if !ok {
		holding.ValuationStatus = "unpriced"
		return true
	}
	if currency == "USD" {
		holding.ValueUSD = *holding.ValueFromScreenshot
		holding.ValuationStatus = "user_provided"
		holding.PricingSource = "USER_PROVIDED"
		currentPriceQuote := 0.0
		if holding.Amount > 0 {
			currentPriceQuote = *holding.ValueFromScreenshot / holding.Amount
		}
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     "USD",
			CurrentPriceQuote: currentPriceQuote,
			FXRateToUSD:       1,
		})
		return true
	}
	if rate, ok := oerRates[currency]; ok && rate > 0 {
		usdRate := 1 / rate
		holding.ValueUSD = *holding.ValueFromScreenshot * usdRate
		holding.ValuationStatus = "user_provided"
		holding.PricingSource = "USER_PROVIDED"
		holding.CurrencyConverted = true
		currentPriceQuote := 0.0
		if holding.Amount > 0 {
			currentPriceQuote = *holding.ValueFromScreenshot / holding.Amount
		}
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     currency,
			CurrentPriceQuote: currentPriceQuote,
			FXRateToUSD:       usdRate,
		})
		return true
	}
	if stablecoinSet()[currency] {
		holding.ValueUSD = *holding.ValueFromScreenshot
		holding.ValuationStatus = "user_provided"
		holding.PricingSource = "USER_PROVIDED"
		currentPriceQuote := 0.0
		if holding.Amount > 0 {
			currentPriceQuote = *holding.ValueFromScreenshot / holding.Amount
		}
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     currency,
			CurrentPriceQuote: currentPriceQuote,
			FXRateToUSD:       1,
		})
		return true
	}
	return false
}
