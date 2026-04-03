package app

import (
	"context"
	"fmt"
	"time"
)

func buildActionAlertS16(ctx context.Context, market *marketClient, plan lockedPlan, planState *planStateData, holdingByAsset map[string]*portfolioHolding, futuresByAssetKey map[string]futuresPremiumIndex, seriesByAssetKey map[string][]ohlcPoint, severity, outputLanguage, baseCurrency string, rateFromUSD float64, now time.Time) (insightItem, bool) {
	if planState == nil {
		return insightItem{}, false
	}
	futures, ok := futuresByAssetKey[plan.AssetKey]
	if !ok || futures.MarkPrice <= 0 {
		return insightItem{}, false
	}
	price, ok := latestPriceForPlanUSD(ctx, market, plan, holdingByAsset, seriesByAssetKey)
	if !ok || price.Price <= 0 {
		return insightItem{}, false
	}
	if cooldown, ok := planState.getTime("cooldown_until"); ok && now.Before(cooldown) {
		return insightItem{}, false
	}
	triggerRate, ok := planState.getFloat("trigger_funding_rate")
	if !ok {
		triggerRate, _ = getFloatParam(plan.Parameters, "trigger_funding_rate")
	}
	basisLimit, ok := planState.getFloat("trigger_basis_pct_max")
	if !ok {
		basisLimit, _ = getFloatParam(plan.Parameters, "trigger_basis_pct_max")
	}
	if triggerRate <= 0 && basisLimit <= 0 {
		return insightItem{}, false
	}
	basisPct := 0.0
	if price.Price > 0 {
		basisPct = (futures.MarkPrice - price.Price) / price.Price
	}
	planState.set("last_funding_rate", futures.LastFundingRate)
	planState.set("last_basis_pct", basisPct)
	if futures.LastFundingRate < triggerRate || basisPct > basisLimit {
		return insightItem{}, false
	}
	nextFunding := futures.NextFundingTime
	thresholdID := "funding_" + nextFunding.Format(time.RFC3339)
	amountUSD, ok := planState.getFloat("hedge_notional_usd")
	if !ok {
		amountUSD, _ = getFloatParam(plan.Parameters, "hedge_notional_usd")
	}
	reason := insightCopy(outputLanguage, copyFundingReason, formatFloat(futures.LastFundingRate*100, 2), formatFloat(basisPct*100, 2))
	action := insightCopy(outputLanguage, copyFundingAction)
	var suggested *suggestedQuantity
	if amountUSD > 0 {
		currency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
		suggested = usdSuggestedQuantity(amountUSD, currency, planRateFromUSD)
	}
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
		ExpiresAt:         nextFunding,
	}, true
}

func buildActionAlertS18(plan lockedPlan, planState *planStateData, seriesByAssetKey map[string][]ohlcPoint, severity, outputLanguage string, now time.Time) (insightItem, bool) {
	if planState == nil {
		return insightItem{}, false
	}
	if nextCheck, ok := planState.getTime("next_check_at"); ok && now.Before(nextCheck) {
		return insightItem{}, false
	}
	series := seriesByAssetKey[plan.AssetKey]
	if len(series) < 200 {
		return insightItem{}, false
	}
	closes := extractCloses(series)
	ma20, ok20 := simpleMovingAverage(closes, 20)
	ma50, ok50 := simpleMovingAverage(closes, 50)
	ma200, ok200 := simpleMovingAverage(closes, 200)
	if !ok20 || !ok50 || !ok200 {
		return insightItem{}, false
	}
	latest := closes[len(closes)-1]
	newTrend, _ := computeTrendState(latest, ma20, ma50, ma200)
	if newTrend == "neutral" {
		planState.set("trend_state", newTrend)
		planState.set("last_trend_state", newTrend)
		planState.set("next_check_at", now.Add(24*time.Hour).Format(time.RFC3339))
		return insightItem{}, false
	}
	lastTrend := planState.getString("trend_state")
	if lastTrend == "" {
		lastTrend = newTrend
	}
	lastSignal, _ := planState.getTime("last_signal_at")
	if !lastSignal.IsZero() && now.Before(lastSignal.Add(24*time.Hour)) {
		planState.set("trend_state", newTrend)
		planState.set("next_check_at", now.Add(24*time.Hour).Format(time.RFC3339))
		return insightItem{}, false
	}
	if trendTransitionAllowed(lastTrend, newTrend) {
		trendLabel := localizedTrendState(outputLanguage, newTrend)
		reason := insightCopy(outputLanguage, copyTrendReason, trendLabel)
		action := insightCopy(outputLanguage, trendActionCopyKey(plan.Parameters))
		item := insightItem{
			ID:              newID("ins"),
			Type:            insightTypeActionAlert,
			Asset:           plan.Symbol,
			AssetKey:        plan.AssetKey,
			Severity:        severity,
			TriggerReason:   reason,
			TriggerKey:      buildPlanTriggerKey(insightTypeActionAlert, plan.PlanID, plan.StrategyID, fmt.Sprintf("trend_%s_%s", newTrend, now.Format("2006-01-02"))),
			StrategyID:      plan.StrategyID,
			PlanID:          plan.PlanID,
			SuggestedAction: action,
			CreatedAt:       now,
			ExpiresAt:       now.Add(24 * time.Hour),
		}
		planState.set("last_signal_at", now.Format(time.RFC3339))
		planState.set("trend_state", newTrend)
		planState.set("last_trend_state", lastTrend)
		planState.set("next_check_at", now.Add(24*time.Hour).Format(time.RFC3339))
		return item, true
	}
	planState.set("trend_state", newTrend)
	planState.set("last_trend_state", lastTrend)
	planState.set("next_check_at", now.Add(24*time.Hour).Format(time.RFC3339))
	return insightItem{}, false
}

func buildActionAlertS22(plan lockedPlan, planState *planStateData, holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, severity, outputLanguage string, now time.Time) (insightItem, bool) {
	if planState == nil {
		return insightItem{}, false
	}
	threshold, ok := planState.getFloat("rebalance_threshold_pct")
	if !ok {
		threshold, _ = getFloatParam(plan.Parameters, "rebalance_threshold_pct")
	}
	if threshold <= 0 {
		threshold = 0.05
	}
	nextRebalance, _ := planState.getTime("next_rebalance_at")
	maxDrift, trades := computeRebalanceTrades(plan.Parameters, holdings, seriesByAssetKey)
	triggered := maxDrift >= threshold || (!nextRebalance.IsZero() && !now.Before(nextRebalance))
	if !triggered {
		return insightItem{}, false
	}
	reason := insightCopy(outputLanguage, copyRebalanceReason, formatFloat(maxDrift*100, 1))
	action := insightCopy(outputLanguage, copyRebalanceAction)
	thresholdID := "rebalance_"
	if !nextRebalance.IsZero() {
		thresholdID += nextRebalance.Format(time.RFC3339)
	}
	suggested := &suggestedQuantity{Mode: "rebalance", Trades: trades}
	planState.set("pending_rebalance", true)
	planState.set("rebalance_trades", trades)
	return insightItem{
		ID:                newID("ins"),
		Type:              insightTypeActionAlert,
		Asset:             "PORTFOLIO",
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
