package app

import (
	"math"
	"strings"
	"time"
)

func buildS05Plan(profile userProfile, metrics portfolioMetrics, contexts []assetPlanContext, deviceTimezone string) (lockedPlan, bool) {
	if metrics.IdleCashUSD < 50 {
		return lockedPlan{}, false
	}
	if !marketsHas(profile.Markets, "Crypto") && !marketsHas(profile.Markets, "Stocks") {
		return lockedPlan{}, false
	}

	baseAmount := math.Min(metrics.IdleCashUSD*0.10, metrics.NonCashPricedValueUSD*0.02)
	riskAdjustment := 1.0
	switch strings.ToLower(profile.RiskPreference) {
	case "yield seeker":
		riskAdjustment = 0.75
	case "speculator":
		riskAdjustment = 1.25
	}
	amount := clamp(baseAmount*riskAdjustment, 20, 2000)

	frequency := "monthly"
	if metrics.Volatility30dDaily > 0.06 {
		frequency = "biweekly"
	}
	if strings.EqualFold(profile.Style, "Day Trading") || strings.EqualFold(profile.Style, "Scalping") {
		frequency = "weekly"
	}

	timezone := strings.TrimSpace(profile.Timezone)
	if timezone == "" {
		timezone = strings.TrimSpace(deviceTimezone)
	}
	if timezone == "" {
		timezone = "UTC"
	}
	nextExecution := nextExecutionAt(timezone, frequency, time.Now())

	target := pickS05Target(profile, contexts)
	return lockedPlan{
		StrategyID: "S05",
		AssetType:  target.AssetType,
		Symbol:     target.Symbol,
		AssetKey:   target.AssetKey,
		Parameters: map[string]any{
			"amount":            roundTo(amount, 2),
			"frequency":         frequency,
			"next_execution_at": nextExecution,
		},
	}, true
}

func pickS05Target(profile userProfile, contexts []assetPlanContext) portfolioHolding {
	for _, ctx := range contexts {
		return ctx.Holding
	}
	if marketsHas(profile.Markets, "Crypto") {
		return portfolioHolding{AssetType: "crypto", Symbol: "BTC", AssetKey: "crypto:cg:bitcoin"}
	}
	return portfolioHolding{AssetType: "stock", Symbol: "SPY", AssetKey: "stock:mic:XNYS:SPY"}
}

func marketsHas(markets []string, value string) bool {
	for _, entry := range markets {
		if strings.EqualFold(entry, value) {
			return true
		}
	}
	return false
}
