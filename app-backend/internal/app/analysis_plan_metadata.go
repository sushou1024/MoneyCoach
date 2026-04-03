package app

import "strings"

func applyPlanMetadata(paid *paidReportPayload, risks []previewRisk) {
	if paid == nil {
		return
	}
	riskSeverity := make(map[string]string, len(risks))
	for _, risk := range risks {
		if strings.TrimSpace(risk.RiskID) == "" {
			continue
		}
		riskSeverity[risk.RiskID] = risk.Severity
	}

	apply := func(plans []paidPlan) {
		for i := range plans {
			plans[i].Priority = planPriorityFromSeverity(riskSeverity[plans[i].LinkedRiskID])
		}
	}

	apply(paid.OptimizationPlan)
	apply(paid.ActionableAdvice)
}

func planPriorityFromSeverity(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical", "high":
		return "High"
	case "medium":
		return "Medium"
	case "low":
		return "Low"
	default:
		return "Medium"
	}
}
