package app

func buildS16Plan(candidate s16Candidate, idleCashUSD float64) lockedPlan {
	spot := candidate.Context.Holding.CurrentPrice
	basisPct := 0.0
	if spot > 0 {
		basisPct = (candidate.Futures.MarkPrice - spot) / spot
	}
	heldHours := 24.0
	feePct := 0.002
	hedgeNotional := clamp(idleCashUSD*0.20, 100, 5000)
	spotAmount := 0.0
	if spot > 0 {
		spotAmount = hedgeNotional / spot
	}

	params := map[string]any{
		"funding_rate_8h":       roundTo(candidate.Futures.LastFundingRate, 6),
		"spot_price":            roundTo(spot, priceDecimals(candidate.Context.Holding.AssetType)),
		"mark_price":            roundTo(candidate.Futures.MarkPrice, priceDecimals(candidate.Context.Holding.AssetType)),
		"basis_pct":             roundTo(basisPct, 6),
		"holding_period_hours":  heldHours,
		"fee_pct":               feePct,
		"hedge_notional_usd":    roundTo(hedgeNotional, 2),
		"spot_amount":           roundTo(spotAmount, amountDecimals(candidate.Context.Holding.AssetType)),
		"futures_symbol":        candidate.FuturesSymbol,
		"trigger_funding_rate":  0.002,
		"trigger_basis_pct_max": 0.002,
	}

	return lockedPlan{
		StrategyID: "S16",
		AssetType:  candidate.Context.Holding.AssetType,
		Symbol:     candidate.Context.Holding.Symbol,
		AssetKey:   candidate.Context.Holding.AssetKey,
		Parameters: params,
	}
}
