package app

import "strings"

func applyCashDeltas(byKey map[string]*portfolioHolding, deltas []transactionDelta) []string {
	warnings := make([]string, 0)
	for _, delta := range deltas {
		if delta.SkipCash {
			continue
		}
		currency := strings.ToUpper(strings.TrimSpace(delta.Currency))
		if currency == "" {
			continue
		}
		cashDelta, ok := cashDeltaForTrade(delta, currency)
		if !ok {
			warnings = append(warnings, "Skipped cash adjustment due to missing currency or price.")
			continue
		}
		if !applyCashDelta(byKey, currency, cashDelta) {
			warnings = append(warnings, "No matching cash balance for "+currency+"; skipped cash adjustment.")
		}
	}
	return warnings
}

func cashDeltaForTrade(delta transactionDelta, currency string) (float64, bool) {
	if delta.Amount == 0 {
		return 0, false
	}
	price := delta.PriceNative
	if price <= 0 {
		if currency == "USD" || stablecoinSet()[currency] {
			price = delta.PriceUSD
		}
	}
	if price <= 0 {
		return 0, false
	}
	notional := abs(delta.Amount) * price
	fees := cashFees(delta, currency)
	if delta.Amount > 0 {
		return -(notional + fees), true
	}
	return notional - fees, true
}

func cashFees(delta transactionDelta, currency string) float64 {
	fees := 0.0
	if delta.FeesNative > 0 && strings.EqualFold(delta.FeesCurrency, currency) {
		fees = delta.FeesNative
	} else if delta.FeesUSD > 0 && (currency == "USD" || stablecoinSet()[currency]) {
		fees = delta.FeesUSD
	}
	return fees
}

func applyCashDelta(holdings map[string]*portfolioHolding, currency string, delta float64) bool {
	for _, holding := range holdings {
		if !isCashHolding(*holding, currency) {
			continue
		}
		if delta < 0 && holding.Amount+delta < 0 {
			return false
		}
		holding.Amount += delta
		return true
	}
	return false
}

func isCashHolding(holding portfolioHolding, currency string) bool {
	if holding.AssetType == "forex" && strings.EqualFold(holding.Symbol, currency) {
		return true
	}
	if holding.AssetType == "crypto" && holding.BalanceType == "stablecoin" && strings.EqualFold(holding.Symbol, currency) {
		return true
	}
	return false
}
