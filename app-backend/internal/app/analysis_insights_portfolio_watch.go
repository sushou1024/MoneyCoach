package app

import (
	"context"
	"time"
)

func buildPortfolioWatchSignals(ctx context.Context, market *marketClient, holdings []portfolioHolding, plans []lockedPlan, seriesByAssetKey map[string][]ohlcPoint, riskByID map[string]string, outputLanguage string, baseCurrency string, rateFromUSD float64, now time.Time) []insightItem {
	items := make([]insightItem, 0)
	holdingByAsset := holdingsByAssetKey(holdings)
	for _, plan := range plans {
		severity := severityFromRisk(plan.LinkedRiskID, riskByID)
		switch plan.StrategyID {
		case "S01":
			if item, ok := buildPortfolioWatchS01(ctx, market, plan, holdingByAsset, seriesByAssetKey, severity, outputLanguage, baseCurrency, rateFromUSD, now); ok {
				items = append(items, item)
			}
		case "S04":
			items = append(items, buildPortfolioWatchS04(ctx, market, plan, holdingByAsset, seriesByAssetKey, severity, outputLanguage, baseCurrency, rateFromUSD, now)...)
		}
	}
	return items
}
