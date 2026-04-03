package app

import "sort"

type assetPlanContext struct {
	Holding                 portfolioHolding
	AssetWeightPct          float64
	Series                  []ohlcPoint
	Closes                  []float64
	MA20                    float64
	MA50                    float64
	MA200                   float64
	HasMA20                 bool
	HasMA50                 bool
	HasMA200                bool
	Volatility30dDaily      float64
	Volatility30dAnnualized float64
	TrendState              string
	TrendStrength           string
}

type s16Candidate struct {
	Context       assetPlanContext
	NetEdgePct    float64
	Futures       futuresPremiumIndex
	FuturesSymbol string
}

func buildAssetPlanContexts(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, nonCashValue float64) []assetPlanContext {
	if nonCashValue <= 0 {
		return nil
	}
	contexts := make([]assetPlanContext, 0)
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
		weight := holding.ValueUSD / nonCashValue
		series := seriesByAssetKey[holding.AssetKey]
		closes := extractCloses(series)
		ctx := assetPlanContext{Holding: holding, AssetWeightPct: weight, Series: series, Closes: closes}
		if value, ok := simpleMovingAverage(closes, 20); ok {
			if rateToUSD := quoteFXRateToUSD(holding); rateToUSD > 0 && quoteCurrencyForHolding(holding) != "USD" {
				value *= rateToUSD
			}
			ctx.MA20 = value
			ctx.HasMA20 = true
		}
		if value, ok := simpleMovingAverage(closes, 50); ok {
			if rateToUSD := quoteFXRateToUSD(holding); rateToUSD > 0 && quoteCurrencyForHolding(holding) != "USD" {
				value *= rateToUSD
			}
			ctx.MA50 = value
			ctx.HasMA50 = true
		}
		if value, ok := simpleMovingAverage(closes, 200); ok {
			if rateToUSD := quoteFXRateToUSD(holding); rateToUSD > 0 && quoteCurrencyForHolding(holding) != "USD" {
				value *= rateToUSD
			}
			ctx.MA200 = value
			ctx.HasMA200 = true
		}
		if vol, ok := assetVolatilityDaily(closes); ok {
			ctx.Volatility30dDaily = vol
			ctx.Volatility30dAnnualized = vol * annualizationFactorForAsset(holding.AssetType)
		}
		if ctx.HasMA20 && ctx.HasMA50 && ctx.HasMA200 {
			ctx.TrendState, ctx.TrendStrength = computeTrendState(holding.CurrentPrice, ctx.MA20, ctx.MA50, ctx.MA200)
		}
		contexts = append(contexts, ctx)
	}

	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].AssetWeightPct > contexts[j].AssetWeightPct
	})

	return contexts
}
