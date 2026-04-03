package app

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
