package app

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

func latestPriceForPlan(ctx context.Context, market *marketClient, plan lockedPlan, seriesByAssetKey map[string][]ohlcPoint) (priceSnapshot, bool) {
	if plan.AssetType == "portfolio" || plan.Symbol == "" {
		return priceSnapshot{}, false
	}
	candidate := portfolioHolding{AssetType: plan.AssetType, Symbol: plan.Symbol, AssetKey: plan.AssetKey}
	series, ok := indicatorSeriesForHolding(ctx, market, candidate, seriesByAssetKey)
	if !ok || len(series.Points) == 0 {
		return priceSnapshot{}, false
	}
	last := series.Points[len(series.Points)-1]
	return priceSnapshot{
		Price:     last.Close,
		Timeframe: series.Interval,
		Timestamp: time.Unix(last.Timestamp, 0).UTC(),
		Source:    series.Source,
	}, true
}

func latestPriceForPlanUSD(ctx context.Context, market *marketClient, plan lockedPlan, holdingByAsset map[string]*portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) (priceSnapshot, bool) {
	price, ok := latestPriceForPlan(ctx, market, plan, seriesByAssetKey)
	if !ok {
		return priceSnapshot{}, false
	}
	holding := holdingByAsset[plan.AssetKey]
	if holding == nil {
		return price, true
	}
	if fxRateToUSD := quoteFXRateToUSD(*holding); fxRateToUSD > 0 {
		price.Price *= fxRateToUSD
	}
	return price, true
}

func holdingsByAssetKey(holdings []portfolioHolding) map[string]*portfolioHolding {
	byKey := make(map[string]*portfolioHolding, len(holdings))
	for i := range holdings {
		if holdings[i].AssetKey == "" {
			continue
		}
		byKey[holdings[i].AssetKey] = &holdings[i]
	}
	return byKey
}

func planDisplayCurrency(plan lockedPlan, holdingByAsset map[string]*portfolioHolding, baseCurrency string, rateFromUSD float64) (string, float64) {
	if plan.AssetType == "portfolio" || strings.HasPrefix(plan.AssetKey, "portfolio:") {
		return baseCurrency, rateFromUSD
	}
	holding := holdingByAsset[plan.AssetKey]
	if holding == nil {
		return "USD", 1
	}
	currency := quoteCurrencyForHolding(*holding)
	fxRateToUSD := quoteFXRateToUSD(*holding)
	if currency == "" || fxRateToUSD <= 0 {
		return "USD", 1
	}
	return currency, 1 / fxRateToUSD
}

func buildInsightCTAPayload(item insightItem) map[string]any {
	payload := map[string]any{
		"target_screen": "SC18",
		"asset":         item.Asset,
	}
	if item.AssetKey != "" {
		payload["asset_key"] = item.AssetKey
	}
	if item.PlanID != "" {
		payload["plan_id"] = item.PlanID
	}
	if item.StrategyID != "" {
		payload["strategy_id"] = item.StrategyID
	}
	if item.Timeframe != "" {
		payload["timeframe"] = item.Timeframe
	}
	return payload
}

func severityFromRisk(riskID string, riskByID map[string]string) string {
	if riskID == "" {
		return "Medium"
	}
	severity := strings.TrimSpace(riskByID[riskID])
	if severity == "" {
		return "Medium"
	}
	return severity
}

func planStateIndex(planState *planStateData, key string) int {
	value, ok := planState.getFloat(key)
	if !ok {
		return 0
	}
	return int(math.Round(value))
}

func trendTransitionAllowed(previous, current string) bool {
	if previous == "" {
		return true
	}
	if previous == "neutral" && current != "neutral" {
		return true
	}
	if (previous == "up" || previous == "strong_up") && (current == "down" || current == "strong_down") {
		return true
	}
	if (previous == "down" || previous == "strong_down") && (current == "up" || current == "strong_up") {
		return true
	}
	return false
}

func buildPlanTriggerKey(insightType, planID, strategyID, thresholdID string) string {
	return fmt.Sprintf("%s:%s:%s:%s", insightType, planID, strategyID, thresholdID)
}

func sortInsights(items []insightItem) {
	sort.Slice(items, func(i, j int) bool {
		typeRank := insightTypeRank(items[i].Type) - insightTypeRank(items[j].Type)
		if typeRank != 0 {
			return typeRank < 0
		}
		severityRank := insightSeverityRank(items[i].Severity) - insightSeverityRank(items[j].Severity)
		if severityRank != 0 {
			return severityRank < 0
		}
		if items[i].Type == insightTypeMarketAlpha && items[j].Type == insightTypeMarketAlpha {
			if items[i].BetaToPortfolio != items[j].BetaToPortfolio {
				return items[i].BetaToPortfolio > items[j].BetaToPortfolio
			}
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}

func insightTypeRank(value string) int {
	switch value {
	case insightTypePortfolioWatch:
		return 0
	case insightTypeActionAlert:
		return 1
	case insightTypeMarketAlpha:
		return 2
	default:
		return 3
	}
}

func insightSeverityRank(value string) int {
	switch strings.ToLower(value) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func formatFloat(value float64, decimals int) string {
	format := "%.0f"
	if decimals > 0 {
		format = "%." + strconv.Itoa(decimals) + "f"
	}
	return fmt.Sprintf(format, value)
}

func formatMoneyDisplay(valueUSD float64, baseCurrency string, rateFromUSD float64, decimals int) string {
	if rateFromUSD <= 0 {
		rateFromUSD = 1
		baseCurrency = "USD"
	}
	baseCurrency = normalizeCurrency(baseCurrency)
	converted := valueUSD * rateFromUSD
	return fmt.Sprintf("%s %s", formatFloat(converted, decimals), baseCurrency)
}

func displayAmountFromUSD(amountUSD float64, baseCurrency string, rateFromUSD float64) (float64, string) {
	if rateFromUSD <= 0 {
		return roundTo(amountUSD, 2), "USD"
	}
	baseCurrency = normalizeCurrency(baseCurrency)
	return roundTo(amountUSD*rateFromUSD, 2), baseCurrency
}

func usdSuggestedQuantity(amountUSD float64, baseCurrency string, rateFromUSD float64) *suggestedQuantity {
	amountDisplay, displayCurrency := displayAmountFromUSD(amountUSD, baseCurrency, rateFromUSD)
	amount := amountUSD
	return &suggestedQuantity{
		Mode:            "usd",
		AmountUSD:       &amount,
		AmountDisplay:   &amountDisplay,
		DisplayCurrency: displayCurrency,
	}
}
