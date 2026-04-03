package app

import (
	"context"
	"fmt"
	"time"
)

func buildActionAlertS02(ctx context.Context, market *marketClient, plan lockedPlan, planState *planStateData, holdingByAsset map[string]*portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, severity, outputLanguage, baseCurrency string, rateFromUSD float64, now time.Time) (insightItem, bool) {
	if planState == nil {
		return insightItem{}, false
	}
	price, ok := latestPriceForPlanUSD(ctx, market, plan, holdingByAsset, seriesByAssetKey)
	if !ok || price.Price <= 0 {
		return insightItem{}, false
	}
	triggerPrice, ok := planState.getFloat("next_safety_order_price")
	if !ok || triggerPrice <= 0 {
		return insightItem{}, false
	}
	amountUSD, ok := planState.getFloat("next_safety_order_amount_usd")
	if !ok || amountUSD <= 0 {
		return insightItem{}, false
	}
	if price.Price > triggerPrice*1.005 {
		return insightItem{}, false
	}
	index := planStateIndex(planState, "next_safety_order_index")
	thresholdID := fmt.Sprintf("safety_%d", index+1)
	currency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
	reason := insightCopy(outputLanguage, copySafetyOrderReason, formatMoneyDisplay(triggerPrice, currency, planRateFromUSD, priceDecimals(plan.AssetType)))
	action := insightCopy(outputLanguage, copySafetyOrderAction, formatMoneyDisplay(amountUSD, currency, planRateFromUSD, 2))
	suggested := usdSuggestedQuantity(amountUSD, currency, planRateFromUSD)
	return insightItem{
		ID:                newID("ins"),
		Type:              insightTypeActionAlert,
		Asset:             plan.Symbol,
		AssetKey:          plan.AssetKey,
		Severity:          severity,
		TriggerReason:     reason,
		TriggerKey:        buildPlanTriggerKey(insightTypeActionAlert, plan.PlanID, plan.StrategyID, thresholdID),
		StrategyID:        plan.StrategyID,
		PlanID:            plan.PlanID,
		SuggestedAction:   action,
		SuggestedQuantity: suggested,
		CreatedAt:         now,
		ExpiresAt:         now.Add(24 * time.Hour),
	}, true
}

func buildActionAlertS03(ctx context.Context, market *marketClient, plan lockedPlan, planState *planStateData, holdingByAsset map[string]*portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, severity, outputLanguage, baseCurrency string, rateFromUSD float64, now time.Time) (insightItem, bool) {
	if planState == nil {
		return insightItem{}, false
	}
	if lastSignal, ok := planState.getTime("last_signal_at"); ok && !lastSignal.IsZero() {
		return insightItem{}, false
	}
	price, ok := latestPriceForPlanUSD(ctx, market, plan, holdingByAsset, seriesByAssetKey)
	if !ok || price.Price <= 0 {
		return insightItem{}, false
	}
	callbackRate, _ := getFloatParam(plan.Parameters, "callback_rate")
	peak := price.Price
	if value, ok := planState.getFloat("peak_price_since_activation"); ok && value > 0 {
		peak = value
	}
	trailingStop := 0.0
	if value, ok := planState.getFloat("trailing_stop_price"); ok && value > 0 {
		trailingStop = value
	}
	if callbackRate > 0 && price.Price > peak {
		peak = price.Price
		planState.set("peak_price_since_activation", peak)
		trailingStop = peak * (1 - callbackRate)
		planState.set("trailing_stop_price", trailingStop)
	}
	if trailingStop <= 0 {
		return insightItem{}, false
	}
	if price.Price > trailingStop {
		return insightItem{}, false
	}
	drawdown := 0.0
	if peak > 0 {
		drawdown = (peak - price.Price) / peak
	}
	currency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
	reason := insightCopy(outputLanguage, copyTrailingReason, formatFloat(drawdown*100, 1), formatMoneyDisplay(peak, currency, planRateFromUSD, priceDecimals(plan.AssetType)))
	action := insightCopy(outputLanguage, copyTrailingAction)
	holding := holdingByAsset[plan.AssetKey]
	if holding == nil || holding.Amount <= 0 {
		return insightItem{}, false
	}
	amount := holding.Amount
	return insightItem{
		ID:                newID("ins"),
		Type:              insightTypeActionAlert,
		Asset:             plan.Symbol,
		AssetKey:          plan.AssetKey,
		Severity:          severity,
		TriggerReason:     reason,
		TriggerKey:        buildPlanTriggerKey(insightTypeActionAlert, plan.PlanID, plan.StrategyID, "trailing_stop"),
		StrategyID:        plan.StrategyID,
		PlanID:            plan.PlanID,
		SuggestedAction:   action,
		SuggestedQuantity: &suggestedQuantity{Mode: "asset", AmountAsset: &amount, Symbol: plan.Symbol, AssetKey: plan.AssetKey},
		CreatedAt:         now,
		ExpiresAt:         now.Add(24 * time.Hour),
	}, true
}

func buildActionAlertS05(plan lockedPlan, planState *planStateData, holdingByAsset map[string]*portfolioHolding, severity, outputLanguage, baseCurrency string, rateFromUSD float64, now time.Time) (insightItem, bool) {
	if planState == nil {
		return insightItem{}, false
	}
	nextExecution, ok := planState.getTime("next_execution_at")
	if !ok {
		return insightItem{}, false
	}
	if now.Before(nextExecution) || now.After(nextExecution.Add(24*time.Hour)) {
		return insightItem{}, false
	}
	amount, ok := getFloatParam(plan.Parameters, "amount")
	if !ok || amount <= 0 {
		return insightItem{}, false
	}
	currency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
	reason := insightCopy(outputLanguage, copyDCAReason, nextExecution.Format(time.RFC3339))
	action := insightCopy(outputLanguage, copyDCAAction, formatMoneyDisplay(amount, currency, planRateFromUSD, 2))
	suggested := usdSuggestedQuantity(amount, currency, planRateFromUSD)
	return insightItem{
		ID:                newID("ins"),
		Type:              insightTypeActionAlert,
		Asset:             plan.Symbol,
		AssetKey:          plan.AssetKey,
		Severity:          severity,
		TriggerReason:     reason,
		TriggerKey:        buildPlanTriggerKey(insightTypeActionAlert, plan.PlanID, plan.StrategyID, "execution_"+nextExecution.Format(time.RFC3339)),
		StrategyID:        plan.StrategyID,
		PlanID:            plan.PlanID,
		SuggestedAction:   action,
		SuggestedQuantity: suggested,
		CreatedAt:         now,
		ExpiresAt:         nextExecution.Add(24 * time.Hour),
	}, true
}

func buildActionAlertS09(plan lockedPlan, planState *planStateData, holdingByAsset map[string]*portfolioHolding, severity, outputLanguage, baseCurrency string, rateFromUSD float64, now time.Time) (insightItem, bool) {
	if planState == nil {
		return insightItem{}, false
	}
	holding := holdingByAsset[plan.AssetKey]
	if holding == nil || holding.PNLPercent == nil {
		return insightItem{}, false
	}
	triggerPct, ok := planState.getFloat("next_trigger_profit_pct")
	if !ok || triggerPct <= 0 {
		return insightItem{}, false
	}
	if *holding.PNLPercent < triggerPct {
		return insightItem{}, false
	}
	amountUSD, ok := planState.getFloat("next_addition_amount_usd")
	if !ok || amountUSD <= 0 {
		return insightItem{}, false
	}
	index := planStateIndex(planState, "next_addition_index")
	thresholdID := fmt.Sprintf("addition_%d", index+1)
	reason := insightCopy(outputLanguage, copyAdditionReason, formatFloat(triggerPct*100, 1))
	currency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
	action := insightCopy(outputLanguage, copyAdditionAction, formatMoneyDisplay(amountUSD, currency, planRateFromUSD, 2))
	suggested := usdSuggestedQuantity(amountUSD, currency, planRateFromUSD)
	return insightItem{
		ID:                newID("ins"),
		Type:              insightTypeActionAlert,
		Asset:             plan.Symbol,
		AssetKey:          plan.AssetKey,
		Severity:          severity,
		TriggerReason:     reason,
		TriggerKey:        buildPlanTriggerKey(insightTypeActionAlert, plan.PlanID, plan.StrategyID, thresholdID),
		StrategyID:        plan.StrategyID,
		PlanID:            plan.PlanID,
		SuggestedAction:   action,
		SuggestedQuantity: suggested,
		CreatedAt:         now,
		ExpiresAt:         now.Add(24 * time.Hour),
	}, true
}
