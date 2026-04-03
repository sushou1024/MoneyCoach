package app

import "fmt"

func extractHoldings(resolved []resolvedAsset) []portfolioHolding {
	holdings := make([]portfolioHolding, 0, len(resolved))
	for _, asset := range resolved {
		holdings = append(holdings, asset.Holding)
	}
	return holdings
}

func aggregateHoldings(holdings []portfolioHolding) []portfolioHolding {
	if len(holdings) == 0 {
		return holdings
	}

	byKey := make(map[string]*portfolioHolding)
	order := make([]string, 0, len(holdings))
	for _, holding := range holdings {
		key := holding.AssetKey
		if key == "" {
			key = fmt.Sprintf("manual:%s", normalizeSymbol(holding.SymbolRaw))
		}
		if existing, ok := byKey[key]; ok {
			existing.Amount += holding.Amount
			existing.ValueUSD += holding.ValueUSD
			if existing.ValueFromScreenshot == nil && holding.ValueFromScreenshot != nil {
				existing.ValueFromScreenshot = holding.ValueFromScreenshot
			}
			if existing.ManualValueUSD == nil && holding.ManualValueUSD != nil {
				existing.ManualValueUSD = holding.ManualValueUSD
			}
			if existing.DisplayCurrency == nil && holding.DisplayCurrency != nil {
				existing.DisplayCurrency = holding.DisplayCurrency
			}
			if existing.AvgPrice == nil && holding.AvgPrice != nil {
				existing.AvgPrice = holding.AvgPrice
			}
			if existing.AvgPriceSource == "" && holding.AvgPriceSource != "" {
				existing.AvgPriceSource = holding.AvgPriceSource
			}
			if existing.PNLPercent == nil && holding.PNLPercent != nil {
				existing.PNLPercent = holding.PNLPercent
			}
			if existing.ValuationStatus == "unpriced" && holding.ValuationStatus != "" {
				existing.ValuationStatus = holding.ValuationStatus
			}
			continue
		}
		clone := holding
		byKey[key] = &clone
		order = append(order, key)
	}

	merged := make([]portfolioHolding, 0, len(order))
	for _, key := range order {
		merged = append(merged, *byKey[key])
	}
	return merged
}
