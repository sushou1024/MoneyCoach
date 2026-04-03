package app

import (
	"context"
	"strings"
)

func indicatorSeriesForHolding(ctx context.Context, market *marketClient, holding portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) (indicatorSeries, bool) {
	if market == nil {
		return indicatorSeries{}, false
	}
	if holding.AssetType == "crypto" {
		symbol := strings.ToUpper(holding.Symbol)
		if symbol == "" {
			return indicatorSeries{}, false
		}
		for _, candidate := range binanceSymbolCandidates(symbol) {
			points, err := market.binanceKlines(ctx, candidate, binanceKlinesInterval, binanceKlinesLimit)
			if err != nil || len(points) == 0 {
				continue
			}
			return indicatorSeries{AssetKey: holding.AssetKey, Symbol: holding.Symbol, Interval: "4h", Source: "binance", Points: points}, true
		}

		if points, ok := seriesByAssetKey[holding.AssetKey]; ok && len(points) > 0 {
			return indicatorSeries{AssetKey: holding.AssetKey, Symbol: holding.Symbol, Interval: "1d", Source: "coingecko", Points: points}, true
		}

		coinID := strings.TrimPrefix(holding.AssetKey, "crypto:cg:")
		if coinID != "" {
			points := fetchCoinGeckoOHLC(ctx, market, coinID, defaultSupportLookback)
			if len(points) > 0 {
				return indicatorSeries{AssetKey: holding.AssetKey, Symbol: holding.Symbol, Interval: "1d", Source: "coingecko", Points: points}, true
			}
		}
	}

	if holding.AssetType == "stock" {
		if points, ok := seriesByAssetKey[holding.AssetKey]; ok && len(points) > 0 {
			return indicatorSeries{AssetKey: holding.AssetKey, Symbol: holding.Symbol, Interval: "1d", Source: "marketstack", Points: points}, true
		}
		points := fetchMarketstackSeries(ctx, market, holding.Symbol, holding.AssetKey)
		if len(points) > 0 {
			return indicatorSeries{AssetKey: holding.AssetKey, Symbol: holding.Symbol, Interval: "1d", Source: "marketstack", Points: points}, true
		}
	}

	return indicatorSeries{}, false
}
