package app

func computeNetWorth(holdings []portfolioHolding) float64 {
	netWorth := 0.0
	for _, holding := range holdings {
		if holding.ValuationStatus == "priced" || holding.ValuationStatus == "user_provided" {
			netWorth += holding.ValueUSD
		}
	}
	return netWorth
}

func filterLowValueHoldings(holdings []portfolioHolding, minWeight float64) ([]portfolioHolding, float64, int) {
	if minWeight <= 0 {
		return holdings, 0, 0
	}
	total := computeNetWorth(holdings)
	if total <= 0 {
		return holdings, 0, 0
	}
	threshold := total * minWeight
	filtered := make([]portfolioHolding, 0, len(holdings))
	dropped := 0
	for _, holding := range holdings {
		if holding.ValuationStatus == "priced" || holding.ValuationStatus == "user_provided" {
			if holding.ValueUSD < threshold {
				dropped++
				continue
			}
		}
		filtered = append(filtered, holding)
	}
	return filtered, threshold, dropped
}
