package app

func computeMarketMetrics(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, annualizationFactor float64) (float64, float64, float64, float64, bool) {
	eligible := buildEligibleReturnSeries(holdings, seriesByAssetKey, 20)
	if len(eligible) == 0 {
		return fallbackMarketMetrics(annualizationFactor)
	}

	portfolioReturns := portfolioReturnsFromIntersection(eligible)
	if len(portfolioReturns) < 2 {
		return fallbackMarketMetrics(annualizationFactor)
	}

	lastReturns := lastNReturns(portfolioReturns, defaultVolatilityLookback)
	volDaily := 0.0
	if len(lastReturns) > 1 {
		volDaily = stddev(lastReturns)
	}
	volAnnual := volDaily * annualizationFactor

	priceSeries := portfolioPriceSeries(portfolioReturns)
	drawdown := maxDrawdown(tailPricePoints(priceSeries, defaultSupportLookback))
	avgCorr := avgPairwiseCorrelation(holdings, seriesByAssetKey)
	return volDaily, volAnnual, drawdown, avgCorr, false
}

func fallbackMarketMetrics(annualizationFactor float64) (float64, float64, float64, float64, bool) {
	volDaily := 0.04
	volAnnual := volDaily * annualizationFactor
	return volDaily, volAnnual, 0.10, 0.30, true
}
