package app

import "math"

type targetWeight struct {
	AssetKey  string
	Symbol    string
	WeightPct float64
}

func targetWeightsFromParams(params map[string]any) []targetWeight {
	raw, ok := params["target_weights"]
	if !ok {
		return nil
	}
	entries, ok := raw.([]any)
	if !ok {
		if typed, ok := raw.([]map[string]any); ok {
			entries = make([]any, 0, len(typed))
			for _, entry := range typed {
				entries = append(entries, entry)
			}
		}
	}
	weights := make([]targetWeight, 0)
	for _, entry := range entries {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		assetKey := getStringParam(item, "asset_key")
		symbol := getStringParam(item, "symbol")
		weightPct, ok := getFloatParam(item, "weight_pct")
		if assetKey == "" || !ok || weightPct <= 0 {
			continue
		}
		weights = append(weights, targetWeight{AssetKey: assetKey, Symbol: symbol, WeightPct: weightPct})
	}
	return weights
}

func computeRebalanceTrades(params map[string]any, holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) (float64, []rebalanceTrade) {
	weights := targetWeightsFromParams(params)
	if len(weights) == 0 {
		return 0, nil
	}
	currentValues := make(map[string]float64)
	total := 0.0
	assetTypes := make(map[string]string)
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		if holding.BalanceType == "stablecoin" || holding.BalanceType == "fiat_cash" {
			continue
		}
		if holding.ValueUSD <= 0 {
			continue
		}
		currentValues[holding.AssetKey] = holding.ValueUSD
		assetTypes[holding.AssetKey] = holding.AssetType
		total += holding.ValueUSD
	}
	if total <= 0 {
		return 0, nil
	}

	maxDrift := 0.0
	trades := make([]rebalanceTrade, 0)
	for _, weight := range weights {
		currentValue := currentValues[weight.AssetKey]
		currentWeight := currentValue / total
		drift := math.Abs(currentWeight - weight.WeightPct)
		if drift > maxDrift {
			maxDrift = drift
		}
		targetValue := weight.WeightPct * total
		tradeUSD := targetValue - currentValue
		if math.Abs(tradeUSD) < 0.01 {
			continue
		}
		side := "sell"
		if tradeUSD > 0 {
			side = "buy"
		}
		amountUSD := roundTo(math.Abs(tradeUSD), 2)
		var amountAsset *float64
		price := latestPriceForAsset(weight.AssetKey, weight.Symbol, assetTypes[weight.AssetKey], seriesByAssetKey)
		if price > 0 {
			converted := roundTo(amountUSD/price, amountDecimals(assetTypes[weight.AssetKey]))
			if converted > 0 {
				amountAsset = &converted
			}
		}
		trades = append(trades, rebalanceTrade{
			AssetKey:    weight.AssetKey,
			Symbol:      weight.Symbol,
			Side:        side,
			AmountUSD:   amountUSD,
			AmountAsset: amountAsset,
		})
	}
	return maxDrift, trades
}

func latestPriceForAsset(assetKey, symbol, assetType string, seriesByAssetKey map[string][]ohlcPoint) float64 {
	if assetKey != "" {
		if series, ok := seriesByAssetKey[assetKey]; ok && len(series) > 0 {
			return series[len(series)-1].Close
		}
	}
	return 0
}
