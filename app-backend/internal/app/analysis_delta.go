package app

type transactionDelta struct {
	Symbol         string
	AssetType      string
	AssetKey       string
	Amount         float64
	PriceUSD       float64
	PriceNative    float64
	Currency       string
	FeesUSD        float64
	FeesNative     float64
	FeesCurrency   string
	AvgPriceSource string
	SkipCash       bool
}

func applyDelta(holdings []portfolioHolding, deltas []transactionDelta) ([]portfolioHolding, []string) {
	byKey := make(map[string]*portfolioHolding)
	for i := range holdings {
		byKey[holdings[i].AssetKey] = &holdings[i]
	}

	warnings := applyAssetDeltas(byKey, deltas)
	warnings = append(warnings, applyCashDeltas(byKey, deltas)...)
	return collectDeltaHoldings(byKey), warnings
}
