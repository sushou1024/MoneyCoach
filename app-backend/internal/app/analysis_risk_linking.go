package app

import "sort"

func assignLinkedRiskIDs(plans []lockedPlan, risks []previewRisk) []lockedPlan {
	if len(plans) == 0 || len(risks) == 0 {
		return plans
	}

	riskByType := make(map[string]string)
	for _, risk := range risks {
		riskByType[risk.Type] = risk.RiskID
	}

	for i := range plans {
		preferred := preferredRiskTypes(plans[i].StrategyID)
		for _, riskType := range preferred {
			if id := riskByType[riskType]; id != "" {
				plans[i].LinkedRiskID = id
				break
			}
		}
		if plans[i].LinkedRiskID == "" {
			plans[i].LinkedRiskID = fallbackRiskID(risks)
		}
	}
	return plans
}

func preferredRiskTypes(strategyID string) []string {
	switch strategyID {
	case "S05", "S09", "S16":
		return []string{"Inefficient Capital Risk"}
	case "S22":
		return []string{"Concentration Risk", "Correlation Risk"}
	case "S01", "S02", "S03", "S04", "S18":
		return []string{"Drawdown Risk", "Volatility Risk", "Correlation Risk"}
	default:
		return nil
	}
}

func fallbackRiskID(risks []previewRisk) string {
	if len(risks) == 0 {
		return ""
	}
	order := map[string]int{"risk_03": 0, "risk_02": 1, "risk_01": 2}
	sort.Slice(risks, func(i, j int) bool {
		return order[risks[i].RiskID] < order[risks[j].RiskID]
	})
	return risks[0].RiskID
}
