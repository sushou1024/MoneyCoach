package app

import (
	"fmt"
	"strings"
)

const maxPlansPerPortfolio = 3

func buildLockedPlans(profile userProfile, holdings []portfolioHolding, metrics portfolioMetrics, seriesByAssetKey map[string][]ohlcPoint, futuresByAssetKey map[string]futuresPremiumIndex, portfolioSnapshotID, deviceTimezone string) []lockedPlan {
	riskLevel := strings.ToLower(profile.RiskTolerance)
	if riskLevel == "" {
		riskLevel = "moderate"
	}

	contexts := buildAssetPlanContexts(holdings, seriesByAssetKey, metrics.NonCashPricedValueUSD)
	plans := make([]lockedPlan, 0, maxPlansPerPortfolio)
	selected := make(map[string]struct{})

	addPlan := func(plan lockedPlan) {
		if len(plans) >= maxPlansPerPortfolio {
			return
		}
		plans = append(plans, plan)
		if plan.AssetKey != "" {
			selected[plan.AssetKey] = struct{}{}
		}
	}

	assetPlansEnabled := metrics.NonCashPricedValueUSD > 0
	if assetPlansEnabled && len(plans) < maxPlansPerPortfolio {
		if candidate, ok := selectS16Candidate(contexts, metrics, futuresByAssetKey, selected); ok {
			addPlan(buildS16Plan(candidate, metrics.IdleCashUSD))
		} else if candidate, ok := selectS04Candidate(contexts, selected); ok {
			addPlan(buildS04Plan(riskLevel, candidate))
		}
	}

	if assetPlansEnabled && len(plans) < maxPlansPerPortfolio {
		if candidate, ok := selectS02Candidate(contexts, metrics, riskLevel, selected); ok {
			if plan, ok := buildS02Plan(riskLevel, metrics, candidate); ok {
				addPlan(plan)
			}
		}
	}

	if assetPlansEnabled && len(plans) < maxPlansPerPortfolio {
		if candidate, ok := selectS03Candidate(contexts, selected); ok {
			if plan, ok := buildS03Plan(riskLevel, candidate); ok {
				addPlan(plan)
			}
		}
	}

	if assetPlansEnabled && len(plans) < maxPlansPerPortfolio {
		if candidate, ok := selectS09Candidate(contexts, metrics, riskLevel, selected); ok {
			addPlan(buildS09Plan(riskLevel, metrics, candidate))
		}
	}

	if assetPlansEnabled && len(plans) < maxPlansPerPortfolio {
		if candidate, ok := selectS18Candidate(contexts, selected); ok {
			if plan, ok := buildS18Plan(candidate); ok {
				addPlan(plan)
			}
		}
	}

	if assetPlansEnabled && len(plans) < maxPlansPerPortfolio {
		if plan, ok := buildS22Plan(metrics, contexts, portfolioSnapshotID); ok {
			addPlan(plan)
		}
	}

	if len(plans) < maxPlansPerPortfolio {
		if plan, ok := buildS05Plan(profile, metrics, contexts, deviceTimezone); ok {
			addPlan(plan)
		}
	}

	if assetPlansEnabled && len(plans) < maxPlansPerPortfolio {
		if candidate, ok := selectS01Candidate(contexts, selected); ok {
			addPlan(buildS01Plan(profile, riskLevel, metrics, candidate))
		}
	}

	for i := range plans {
		plans[i].PlanID = fmt.Sprintf("plan_%02d", i+1)
	}

	return plans
}
