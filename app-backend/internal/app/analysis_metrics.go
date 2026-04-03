package app

func computePortfolioMetrics(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) portfolioMetrics {
	metrics := portfolioMetrics{}
	metrics.NetWorthUSD = computeNetWorth(holdings)
	if metrics.NetWorthUSD <= 0 {
		return metrics
	}

	threshold := metrics.NetWorthUSD * minPortfolioWeight
	cashLike := 0.0
	pricedValue := 0.0
	nonCashPriced := 0.0
	idleCash := 0.0
	topValue := 0.0
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" && holding.ValuationStatus != "user_provided" {
			continue
		}
		if holding.ValuationStatus == "priced" {
			pricedValue += holding.ValueUSD
			if holding.AssetType == "crypto" || holding.AssetType == "stock" {
				if holding.BalanceType != "stablecoin" {
					nonCashPriced += holding.ValueUSD
				}
			}
			if isCashLike(holding) {
				idleCash += holding.ValueUSD
			}
		}
		if holding.ValueUSD < threshold {
			continue
		}
		if isCashLike(holding) {
			cashLike += holding.ValueUSD
		}
		if holding.ValueUSD > topValue {
			topValue = holding.ValueUSD
		}
	}
	metrics.PricedValueUSD = pricedValue
	metrics.NonCashPricedValueUSD = nonCashPriced
	metrics.CashPct = clamp(cashLike/metrics.NetWorthUSD, 0, 1)
	metrics.TopAssetPct = clamp(topValue/metrics.NetWorthUSD, 0, 1)
	metrics.PricedCoveragePct = clamp(pricedValue/metrics.NetWorthUSD, 0, 1)
	metrics.IdleCashUSD = idleCash
	metrics.CryptoWeight = computeCryptoWeight(holdings, nonCashPriced)

	annualizationFactor := annualizationFactorFromWeight(metrics.CryptoWeight)
	volDaily, volAnnual, drawdown, avgCorr, fallback := computeMarketMetrics(holdings, seriesByAssetKey, annualizationFactor)
	metrics.Volatility30dDaily = volDaily
	metrics.Volatility30dAnnualized = volAnnual
	metrics.MaxDrawdown90d = drawdown
	metrics.AvgPairwiseCorr = avgCorr

	metrics.MetricsIncomplete = metrics.PricedCoveragePct < 0.60 || fallback
	metrics.HealthScoreBaseline = computeHealthScoreBaseline(metrics)
	metrics.VolatilityScoreBaseline = int(clamp(metrics.Volatility30dAnnualized*100, 0, 100))
	return metrics
}
