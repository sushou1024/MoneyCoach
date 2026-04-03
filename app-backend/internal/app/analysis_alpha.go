package app

import (
	"context"
	"math"
)

func computeAlpha30d(ctx context.Context, market *marketClient, holdings []portfolioHolding, metrics portfolioMetrics, seriesByAssetKey map[string][]ohlcPoint) float64 {
	eligible := buildEligibleReturnSeries(holdings, seriesByAssetKey, 20)
	portfolioReturns := portfolioReturnsFromIntersection(eligible)
	portfolioReturn, ok := computeReturnFromReturns(portfolioReturns)
	if !ok {
		return 0
	}

	hasCrypto := false
	hasStock := false
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType == "crypto" && holding.BalanceType != "stablecoin" {
			hasCrypto = true
		}
		if holding.AssetType == "stock" {
			hasStock = true
		}
	}

	btcReturn, btcOk := loadBenchmarkReturn(ctx, market, "crypto", "BTC", "bitcoin", seriesByAssetKey)
	spyReturn, spyOk := loadBenchmarkReturn(ctx, market, "stock", "SPY", "", seriesByAssetKey)

	if hasCrypto && !hasStock {
		if !btcOk {
			return 0
		}
		return portfolioReturn - btcReturn
	}
	if hasStock && !hasCrypto {
		if !spyOk {
			return 0
		}
		return portfolioReturn - spyReturn
	}
	if !btcOk || !spyOk {
		return 0
	}
	benchmark := metrics.CryptoWeight*btcReturn + (1-metrics.CryptoWeight)*spyReturn
	return portfolioReturn - benchmark
}

func loadBenchmarkReturn(ctx context.Context, market *marketClient, assetType, symbol, coinID string, seriesByAssetKey map[string][]ohlcPoint) (float64, bool) {
	switch assetType {
	case "crypto":
		if coinID == "" {
			return 0, false
		}
		points := fetchCoinGeckoOHLC(ctx, market, coinID, defaultSupportLookback)
		if len(points) == 0 {
			return 0, false
		}
		return computeReturnFromCloses(points)
	case "stock":
		if symbol == "" {
			return 0, false
		}
		points := fetchMarketstackSeries(ctx, market, symbol, "")
		if len(points) == 0 {
			return 0, false
		}
		return computeReturnFromCloses(points)
	default:
		return 0, false
	}
}

func computeReturnFromCloses(points []ohlcPoint) (float64, bool) {
	if len(points) < 20 {
		return 0, false
	}
	closes := extractCloses(points)
	if len(closes) < 2 {
		return 0, false
	}
	window := 30
	if len(closes)-1 < window {
		window = len(closes) - 1
	}
	if window <= 0 {
		return 0, false
	}
	start := closes[len(closes)-1-window]
	end := closes[len(closes)-1]
	if start <= 0 || end <= 0 {
		return 0, false
	}
	return (end / start) - 1, true
}

func computeReturnFromReturns(returns []returnPoint) (float64, bool) {
	if len(returns) < 20 {
		return 0, false
	}
	window := 30
	if len(returns) < window {
		window = len(returns)
	}
	sum := 0.0
	for _, ret := range returns[len(returns)-window:] {
		sum += ret.Value
	}
	return math.Exp(sum) - 1, true
}
