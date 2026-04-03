package app

import (
	"context"
	"sort"
	"time"
)

func selectDailyAlphaSignal(ctx context.Context, market *marketClient, holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, outputLanguage string) *paidInsightItem {
	items, _ := buildMarketAlphaSignals(ctx, market, holdings, seriesByAssetKey, outputLanguage, time.Now().UTC())
	if len(items) == 0 {
		return nil
	}
	sort.Slice(items, func(i, j int) bool {
		rankI := riskSeverityRank(items[i].Severity)
		rankJ := riskSeverityRank(items[j].Severity)
		if rankI != rankJ {
			return rankI < rankJ
		}
		if items[i].BetaToPortfolio != items[j].BetaToPortfolio {
			return items[i].BetaToPortfolio > items[j].BetaToPortfolio
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	best := items[0]
	return &paidInsightItem{
		Type:            best.Type,
		Asset:           best.Asset,
		AssetKey:        best.AssetKey,
		Timeframe:       best.Timeframe,
		Severity:        best.Severity,
		TriggerReason:   best.TriggerReason,
		TriggerKey:      best.TriggerKey,
		StrategyID:      best.StrategyID,
		PlanID:          best.PlanID,
		SuggestedAction: best.SuggestedAction,
		CreatedAt:       best.CreatedAt.Format(time.RFC3339),
		ExpiresAt:       best.ExpiresAt.Format(time.RFC3339),
	}
}

func sortOptimizationPlans(paid *paidReportPayload, plans []lockedPlan, risks []previewRisk) {
	if paid == nil {
		return
	}
	riskRank := make(map[string]int, len(risks))
	for _, risk := range risks {
		riskRank[risk.RiskID] = riskSeverityRank(risk.Severity)
	}
	planOrder := make(map[string]int, len(plans))
	for i, plan := range plans {
		planOrder[plan.PlanID] = i
	}
	sort.Slice(paid.OptimizationPlan, func(i, j int) bool {
		left := paid.OptimizationPlan[i]
		right := paid.OptimizationPlan[j]
		ri, ok := riskRank[left.LinkedRiskID]
		if !ok {
			ri = riskSeverityRank("Medium")
		}
		rj, ok := riskRank[right.LinkedRiskID]
		if !ok {
			rj = riskSeverityRank("Medium")
		}
		if ri != rj {
			return ri < rj
		}
		return planOrder[left.PlanID] < planOrder[right.PlanID]
	})
	paid.ActionableAdvice = append([]paidPlan(nil), paid.OptimizationPlan...)
}
