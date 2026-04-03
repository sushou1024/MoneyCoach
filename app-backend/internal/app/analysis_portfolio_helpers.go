package app

import "math"

func isCashLike(holding portfolioHolding) bool {
	if holding.BalanceType == "fiat_cash" || holding.BalanceType == "stablecoin" {
		return true
	}
	return holding.AssetType == "forex"
}

func computeCryptoWeight(holdings []portfolioHolding, nonCashValue float64) float64 {
	if nonCashValue <= 0 {
		return 0
	}
	crypto := 0.0
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" || holding.BalanceType == "stablecoin" {
			continue
		}
		crypto += holding.ValueUSD
	}
	return clamp(crypto/nonCashValue, 0, 1)
}

func annualizationFactorFromWeight(cryptoWeight float64) float64 {
	if cryptoWeight >= 0.50 {
		return math.Sqrt(365)
	}
	return math.Sqrt(252)
}
