package app

import (
	"math"
	"strings"
)

func buildS22Plan(metrics portfolioMetrics, contexts []assetPlanContext, portfolioSnapshotID string) (lockedPlan, bool) {
	if metrics.PricedCoveragePct < 0.80 {
		return lockedPlan{}, false
	}
	eligible := make([]assetPlanContext, 0)
	for _, ctx := range contexts {
		if ctx.Volatility30dAnnualized <= 0 {
			continue
		}
		eligible = append(eligible, ctx)
	}
	if len(eligible) < 3 {
		return lockedPlan{}, false
	}

	volFloor := 0.05
	invSum := 0.0
	for _, ctx := range eligible {
		invSum += 1 / math.Max(ctx.Volatility30dAnnualized, volFloor)
	}
	if invSum <= 0 {
		return lockedPlan{}, false
	}

	weights := make([]map[string]any, 0, len(eligible))
	for _, ctx := range eligible {
		weight := (1 / math.Max(ctx.Volatility30dAnnualized, volFloor)) / invSum
		weights = append(weights, map[string]any{
			"asset_key":                 ctx.Holding.AssetKey,
			"symbol":                    ctx.Holding.Symbol,
			"weight_pct":                roundTo(weight, 4),
			"volatility_30d_annualized": roundTo(ctx.Volatility30dAnnualized, 4),
		})
	}

	params := map[string]any{
		"target_weights":          weights,
		"vol_floor":               volFloor,
		"rebalance_threshold_pct": 0.05,
		"rebalance_frequency":     "monthly",
	}

	assetKey := "portfolio:"
	if strings.TrimSpace(portfolioSnapshotID) != "" {
		assetKey += portfolioSnapshotID
	}
	return lockedPlan{
		StrategyID: "S22",
		AssetType:  "portfolio",
		Symbol:     "PORTFOLIO",
		AssetKey:   assetKey,
		Parameters: params,
	}, true
}
