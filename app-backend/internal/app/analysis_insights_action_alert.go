package app

import (
	"context"
	"time"
)

func buildActionAlertSignals(ctx context.Context, market *marketClient, holdings []portfolioHolding, plans []lockedPlan, planStates map[string]*planStateData, seriesByAssetKey map[string][]ohlcPoint, futuresByAssetKey map[string]futuresPremiumIndex, riskByID map[string]string, outputLanguage string, baseCurrency string, rateFromUSD float64, now time.Time) []insightItem {
	items := make([]insightItem, 0)
	holdingByAsset := holdingsByAssetKey(holdings)
	for _, plan := range plans {
		planState := planStates[plan.PlanID]
		severity := severityFromRisk(plan.LinkedRiskID, riskByID)

		switch plan.StrategyID {
		case "S02":
			if item, ok := buildActionAlertS02(ctx, market, plan, planState, holdingByAsset, seriesByAssetKey, severity, outputLanguage, baseCurrency, rateFromUSD, now); ok {
				items = append(items, item)
			}
		case "S03":
			if item, ok := buildActionAlertS03(ctx, market, plan, planState, holdingByAsset, seriesByAssetKey, severity, outputLanguage, baseCurrency, rateFromUSD, now); ok {
				items = append(items, item)
			}
		case "S05":
			if item, ok := buildActionAlertS05(plan, planState, holdingByAsset, severity, outputLanguage, baseCurrency, rateFromUSD, now); ok {
				items = append(items, item)
			}
		case "S09":
			if item, ok := buildActionAlertS09(plan, planState, holdingByAsset, severity, outputLanguage, baseCurrency, rateFromUSD, now); ok {
				items = append(items, item)
			}
		case "S16":
			if item, ok := buildActionAlertS16(ctx, market, plan, planState, holdingByAsset, futuresByAssetKey, seriesByAssetKey, severity, outputLanguage, baseCurrency, rateFromUSD, now); ok {
				items = append(items, item)
			}
		case "S18":
			if item, ok := buildActionAlertS18(plan, planState, seriesByAssetKey, severity, outputLanguage, now); ok {
				items = append(items, item)
			}
		case "S22":
			if item, ok := buildActionAlertS22(plan, planState, holdings, seriesByAssetKey, severity, outputLanguage, now); ok {
				items = append(items, item)
			}
		}
	}
	return items
}
