package app

import "math"

func selectS16Candidate(contexts []assetPlanContext, metrics portfolioMetrics, futuresByAssetKey map[string]futuresPremiumIndex, excluded map[string]struct{}) (s16Candidate, bool) {
	if metrics.IdleCashUSD < 100 {
		return s16Candidate{}, false
	}
	var selected *s16Candidate
	for _, ctx := range contexts {
		if ctx.Holding.AssetType != "crypto" {
			continue
		}
		if _, ok := excluded[ctx.Holding.AssetKey]; ok {
			continue
		}
		futures, ok := futuresByAssetKey[ctx.Holding.AssetKey]
		if !ok || futures.MarkPrice <= 0 || ctx.Holding.CurrentPrice <= 0 {
			continue
		}
		basis := (futures.MarkPrice - ctx.Holding.CurrentPrice) / ctx.Holding.CurrentPrice
		netEdge := futures.LastFundingRate*(24.0/8.0) + math.Max(0, -basis) - 0.002
		if netEdge < 0.005 {
			continue
		}
		candidate := s16Candidate{Context: ctx, NetEdgePct: netEdge, Futures: futures, FuturesSymbol: futures.Symbol}
		if selected == nil {
			selected = &candidate
			continue
		}
		if netEdge > selected.NetEdgePct {
			selected = &candidate
			continue
		}
		if netEdge == selected.NetEdgePct {
			if ctx.AssetWeightPct > selected.Context.AssetWeightPct {
				selected = &candidate
				continue
			}
			if ctx.AssetWeightPct == selected.Context.AssetWeightPct && ctx.Holding.AssetKey < selected.Context.Holding.AssetKey {
				selected = &candidate
			}
		}
	}
	if selected == nil {
		return s16Candidate{}, false
	}
	return *selected, true
}

func selectS04Candidate(contexts []assetPlanContext, excluded map[string]struct{}) (assetPlanContext, bool) {
	var selected *assetPlanContext
	for i := range contexts {
		ctx := contexts[i]
		if _, ok := excluded[ctx.Holding.AssetKey]; ok {
			continue
		}
		if ctx.Holding.AvgPrice == nil || ctx.Holding.PNLPercent == nil {
			continue
		}
		if *ctx.Holding.PNLPercent < 0.30 {
			continue
		}
		if selected == nil {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent > *selected.Holding.PNLPercent {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent == *selected.Holding.PNLPercent {
			if ctx.AssetWeightPct > selected.AssetWeightPct {
				selected = &ctx
				continue
			}
			if ctx.AssetWeightPct == selected.AssetWeightPct && ctx.Holding.AssetKey < selected.Holding.AssetKey {
				selected = &ctx
			}
		}
	}
	if selected == nil {
		return assetPlanContext{}, false
	}
	return *selected, true
}

func selectS02Candidate(contexts []assetPlanContext, metrics portfolioMetrics, riskLevel string, excluded map[string]struct{}) (assetPlanContext, bool) {
	if riskLevel == "conservative" {
		return assetPlanContext{}, false
	}
	var selected *assetPlanContext
	for i := range contexts {
		ctx := contexts[i]
		if _, ok := excluded[ctx.Holding.AssetKey]; ok {
			continue
		}
		if ctx.Holding.PNLPercent == nil || ctx.Holding.AvgPrice == nil {
			continue
		}
		if *ctx.Holding.PNLPercent > -0.20 {
			continue
		}
		if ctx.Holding.CostBasisStatus != "provided" {
			continue
		}
		if ctx.AssetWeightPct >= 0.60 {
			continue
		}
		if !ctx.HasMA50 || !ctx.HasMA200 || ctx.MA50 < ctx.MA200 {
			continue
		}
		currentInvestment := ctx.Holding.Amount * ctx.Holding.CurrentPrice
		if currentInvestment <= 0 {
			continue
		}
		if metrics.IdleCashUSD < currentInvestment*0.30 {
			continue
		}
		if len(ctx.Closes) < 200 {
			continue
		}
		if selected == nil {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent < *selected.Holding.PNLPercent {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent == *selected.Holding.PNLPercent {
			if ctx.AssetWeightPct > selected.AssetWeightPct {
				selected = &ctx
				continue
			}
			if ctx.AssetWeightPct == selected.AssetWeightPct && ctx.Holding.AssetKey < selected.Holding.AssetKey {
				selected = &ctx
			}
		}
	}
	if selected == nil {
		return assetPlanContext{}, false
	}
	return *selected, true
}

func selectS03Candidate(contexts []assetPlanContext, excluded map[string]struct{}) (assetPlanContext, bool) {
	var selected *assetPlanContext
	for i := range contexts {
		ctx := contexts[i]
		if _, ok := excluded[ctx.Holding.AssetKey]; ok {
			continue
		}
		if ctx.Holding.PNLPercent == nil || ctx.Holding.AvgPrice == nil {
			continue
		}
		if *ctx.Holding.PNLPercent < 0.15 {
			continue
		}
		if ctx.Holding.CostBasisStatus != "provided" {
			continue
		}
		if !ctx.HasMA20 || !ctx.HasMA50 {
			continue
		}
		if ctx.Holding.CurrentPrice <= ctx.MA20 || ctx.MA20 <= ctx.MA50 {
			continue
		}
		if len(ctx.Closes) < 50 {
			continue
		}
		if selected == nil {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent > *selected.Holding.PNLPercent {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent == *selected.Holding.PNLPercent {
			if ctx.AssetWeightPct > selected.AssetWeightPct {
				selected = &ctx
				continue
			}
			if ctx.AssetWeightPct == selected.AssetWeightPct && ctx.Holding.AssetKey < selected.Holding.AssetKey {
				selected = &ctx
			}
		}
	}
	if selected == nil {
		return assetPlanContext{}, false
	}
	return *selected, true
}

func selectS09Candidate(contexts []assetPlanContext, metrics portfolioMetrics, riskLevel string, excluded map[string]struct{}) (assetPlanContext, bool) {
	if metrics.IdleCashUSD < 20 {
		return assetPlanContext{}, false
	}
	profitStep := s09ProfitStepPct(riskLevel)
	var selected *assetPlanContext
	for i := range contexts {
		ctx := contexts[i]
		if _, ok := excluded[ctx.Holding.AssetKey]; ok {
			continue
		}
		if ctx.Holding.PNLPercent == nil {
			continue
		}
		if *ctx.Holding.PNLPercent < profitStep {
			continue
		}
		if selected == nil {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent > *selected.Holding.PNLPercent {
			selected = &ctx
			continue
		}
		if *ctx.Holding.PNLPercent == *selected.Holding.PNLPercent {
			if ctx.AssetWeightPct > selected.AssetWeightPct {
				selected = &ctx
				continue
			}
			if ctx.AssetWeightPct == selected.AssetWeightPct && ctx.Holding.AssetKey < selected.Holding.AssetKey {
				selected = &ctx
			}
		}
	}
	if selected == nil {
		return assetPlanContext{}, false
	}
	return *selected, true
}

func selectS18Candidate(contexts []assetPlanContext, excluded map[string]struct{}) (assetPlanContext, bool) {
	var selected *assetPlanContext
	for i := range contexts {
		ctx := contexts[i]
		if _, ok := excluded[ctx.Holding.AssetKey]; ok {
			continue
		}
		if ctx.TrendState == "" || ctx.TrendState == "neutral" {
			continue
		}
		if len(ctx.Closes) < 200 {
			continue
		}
		if selected == nil {
			selected = &ctx
			continue
		}
		if trendRank(ctx.TrendState) > trendRank(selected.TrendState) {
			selected = &ctx
			continue
		}
		if trendRank(ctx.TrendState) == trendRank(selected.TrendState) {
			if ctx.AssetWeightPct > selected.AssetWeightPct {
				selected = &ctx
				continue
			}
			if ctx.AssetWeightPct == selected.AssetWeightPct && ctx.Holding.AssetKey < selected.Holding.AssetKey {
				selected = &ctx
			}
		}
	}
	if selected == nil {
		return assetPlanContext{}, false
	}
	return *selected, true
}

func selectS01Candidate(contexts []assetPlanContext, excluded map[string]struct{}) (assetPlanContext, bool) {
	if len(contexts) == 0 {
		return assetPlanContext{}, false
	}
	limit := 3
	if len(contexts) < limit {
		limit = len(contexts)
	}
	for i := 0; i < limit; i++ {
		ctx := contexts[i]
		if _, ok := excluded[ctx.Holding.AssetKey]; ok {
			continue
		}
		return ctx, true
	}
	return assetPlanContext{}, false
}

func trendRank(state string) int {
	switch state {
	case "strong_down":
		return 4
	case "down":
		return 3
	case "strong_up":
		return 2
	case "up":
		return 1
	default:
		return 0
	}
}
