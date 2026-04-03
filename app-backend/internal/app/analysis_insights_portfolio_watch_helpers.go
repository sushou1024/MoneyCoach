package app

import (
	"context"
	"fmt"
	"time"
)

func buildPortfolioWatchS01(ctx context.Context, market *marketClient, plan lockedPlan, holdingByAsset map[string]*portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, severity, outputLanguage, baseCurrency string, rateFromUSD float64, now time.Time) (insightItem, bool) {
	stopLossPrice, ok := getFloatParam(plan.Parameters, "stop_loss_price")
	if !ok || stopLossPrice <= 0 {
		return insightItem{}, false
	}
	price, ok := latestPriceForPlanUSD(ctx, market, plan, holdingByAsset, seriesByAssetKey)
	if !ok || price.Price <= 0 {
		return insightItem{}, false
	}
	if price.Price > stopLossPrice*1.005 {
		return insightItem{}, false
	}
	currency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
	reason := insightCopy(outputLanguage, copyStopLossReason, formatMoneyDisplay(stopLossPrice, currency, planRateFromUSD, priceDecimals(plan.AssetType)))
	action := insightCopy(outputLanguage, copyStopLossAction)
	return insightItem{
		ID:              newID("ins"),
		Type:            insightTypePortfolioWatch,
		Asset:           plan.Symbol,
		AssetKey:        plan.AssetKey,
		Severity:        severity,
		TriggerReason:   reason,
		TriggerKey:      buildPlanTriggerKey(insightTypePortfolioWatch, plan.PlanID, plan.StrategyID, "stop_loss"),
		StrategyID:      plan.StrategyID,
		PlanID:          plan.PlanID,
		SuggestedAction: action,
		CreatedAt:       now,
		ExpiresAt:       now.Add(7 * 24 * time.Hour),
	}, true
}

func buildPortfolioWatchS04(ctx context.Context, market *marketClient, plan lockedPlan, holdingByAsset map[string]*portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, severity, outputLanguage, baseCurrency string, rateFromUSD float64, now time.Time) []insightItem {
	price, ok := latestPriceForPlanUSD(ctx, market, plan, holdingByAsset, seriesByAssetKey)
	if !ok || price.Price <= 0 {
		return nil
	}
	currency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
	layers := planLayers(plan.Parameters)
	items := make([]insightItem, 0, len(layers))
	for i, layer := range layers {
		targetPrice, ok := getFloatParam(layer, "target_price")
		if !ok || targetPrice <= 0 {
			continue
		}
		if price.Price < targetPrice*0.995 {
			continue
		}
		layerName := getStringParam(layer, "layer_name")
		sellPct, _ := getFloatParam(layer, "sell_percentage")
		reason := insightCopy(outputLanguage, copyTakeProfitReason, formatMoneyDisplay(targetPrice, currency, planRateFromUSD, priceDecimals(plan.AssetType)))
		action := insightCopy(outputLanguage, copyTakeProfitAction, formatFloat(sellPct*100, 0), layerName)
		items = append(items, insightItem{
			ID:              newID("ins"),
			Type:            insightTypePortfolioWatch,
			Asset:           plan.Symbol,
			AssetKey:        plan.AssetKey,
			Severity:        severity,
			TriggerReason:   reason,
			TriggerKey:      buildPlanTriggerKey(insightTypePortfolioWatch, plan.PlanID, plan.StrategyID, fmt.Sprintf("layer_%d", i+1)),
			StrategyID:      plan.StrategyID,
			PlanID:          plan.PlanID,
			SuggestedAction: action,
			CreatedAt:       now,
			ExpiresAt:       now.Add(7 * 24 * time.Hour),
		})
	}
	return items
}
