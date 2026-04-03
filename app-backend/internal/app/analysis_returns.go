package app

import (
	"context"
	"strings"
)

func returnsByTimestamp(ctx context.Context, market *marketClient, seriesByAssetKey map[string][]ohlcPoint, holding portfolioHolding) map[int64]float64 {
	points := seriesByAssetKey[holding.AssetKey]
	if len(points) == 0 {
		switch holding.AssetType {
		case "crypto":
			coinID := strings.TrimPrefix(holding.AssetKey, "crypto:cg:")
			if coinID != "" {
				points = fetchCoinGeckoOHLC(ctx, market, coinID, defaultSupportLookback)
			}
		case "stock":
			points = fetchMarketstackSeries(ctx, market, holding.Symbol, holding.AssetKey)
		}
	}
	if len(points) < 2 {
		return nil
	}
	return returnsByTimestampFromPoints(points)
}

func betaToPortfolio(assetReturns map[int64]float64, portfolioReturns map[int64]float64) float64 {
	if len(assetReturns) == 0 || len(portfolioReturns) == 0 {
		return 0
	}
	valuesX := make([]float64, 0)
	valuesY := make([]float64, 0)
	for ts, pr := range portfolioReturns {
		if ar, ok := assetReturns[ts]; ok {
			valuesX = append(valuesX, ar)
			valuesY = append(valuesY, pr)
		}
	}
	if len(valuesX) < 20 {
		return 0
	}
	meanX := mean(valuesX)
	meanY := mean(valuesY)
	cov := 0.0
	varY := 0.0
	for i := range valuesX {
		x := valuesX[i] - meanX
		y := valuesY[i] - meanY
		cov += x * y
		varY += y * y
	}
	if varY == 0 {
		return 0
	}
	return cov / varY
}
