package main

func healthStatusFromScore(score int) string {
	switch {
	case score >= 90:
		return "Excellent"
	case score >= 70:
		return "Stable"
	case score >= 50:
		return "Warning"
	default:
		return "Critical"
	}
}

func scoreStatusFromScore(score int) string {
	switch {
	case score >= 70:
		return "Green"
	case score >= 50:
		return "Yellow"
	default:
		return "Red"
	}
}

func volatilityStatusFromScore(score int) string {
	switch {
	case score <= 39:
		return "Green"
	case score <= 59:
		return "Yellow"
	default:
		return "Red"
	}
}

func normalizePreviewStatus(preview *previewPromptOutput) {
	if preview == nil {
		return
	}
	preview.FixedMetrics.HealthStatus = healthStatusFromScore(preview.FixedMetrics.HealthScore)
}

func normalizePaidStatus(paid *paidPromptOutput) {
	if paid == nil {
		return
	}
	paid.ReportHeader.HealthScore.Status = scoreStatusFromScore(paid.ReportHeader.HealthScore.Value)
	paid.ReportHeader.VolatilityDashboard.Status = volatilityStatusFromScore(paid.ReportHeader.VolatilityDashboard.Value)
}
