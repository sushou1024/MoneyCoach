package app

import "strings"

func applyAssetDeltas(byKey map[string]*portfolioHolding, deltas []transactionDelta) []string {
	warnings := make([]string, 0)
	for _, delta := range deltas {
		if delta.AssetKey == "" {
			continue
		}
		holding := ensureHoldingForDelta(byKey, delta)
		if delta.Amount == 0 {
			continue
		}
		if delta.Amount > 0 {
			applyBuyDelta(holding, delta)
			continue
		}
		if warning, ok := applySellDelta(holding, delta); ok {
			warnings = append(warnings, warning)
		}
	}
	return warnings
}

func ensureHoldingForDelta(byKey map[string]*portfolioHolding, delta transactionDelta) *portfolioHolding {
	holding := byKey[delta.AssetKey]
	if holding != nil {
		return holding
	}
	holding = &portfolioHolding{
		Symbol:    delta.Symbol,
		AssetType: delta.AssetType,
		AssetKey:  delta.AssetKey,
	}
	if delta.AssetType == "crypto" && stablecoinSet()[strings.ToUpper(strings.TrimSpace(delta.Symbol))] {
		holding.BalanceType = "stablecoin"
	}
	byKey[delta.AssetKey] = holding
	return holding
}

func applyBuyDelta(holding *portfolioHolding, delta transactionDelta) {
	if holding.AvgPriceSource == "user_input" {
		holding.Amount += delta.Amount
		return
	}
	if delta.PriceUSD > 0 {
		currentCost := 0.0
		currentAmount := holding.Amount
		if holding.AvgPrice != nil {
			currentCost = *holding.AvgPrice * currentAmount
		}
		newCost := delta.Amount * delta.PriceUSD
		newAmount := currentAmount + delta.Amount
		if newAmount > 0 {
			avg := (currentCost + newCost) / newAmount
			holding.AvgPrice = &avg
			if delta.AvgPriceSource != "" {
				holding.AvgPriceSource = delta.AvgPriceSource
			} else if holding.AvgPriceSource == "" {
				holding.AvgPriceSource = "trade"
			}
			if holding.AvgPriceSource == "derived_from_market" {
				holding.CostBasisStatus = "unknown"
			} else {
				holding.CostBasisStatus = "provided"
			}
		}
	}
	holding.Amount += delta.Amount
}

func applySellDelta(holding *portfolioHolding, delta transactionDelta) (string, bool) {
	warning := ""
	if holding.Amount+delta.Amount < 0 {
		warning = "Sell exceeds current amount; clamped to zero."
		delta.Amount = -holding.Amount
	}
	holding.Amount += delta.Amount
	if holding.Amount == 0 {
		holding.AvgPrice = nil
		holding.AvgPriceSource = ""
		holding.PNLPercent = nil
		holding.CostBasisStatus = "unknown"
	}
	if warning != "" {
		return warning, true
	}
	return "", false
}

func collectDeltaHoldings(byKey map[string]*portfolioHolding) []portfolioHolding {
	updated := make([]portfolioHolding, 0, len(byKey))
	for _, holding := range byKey {
		if holding.Amount == 0 {
			continue
		}
		updated = append(updated, *holding)
	}
	return updated
}
