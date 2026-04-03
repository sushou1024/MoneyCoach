package app

func computeTrendState(price, ma20, ma50, ma200 float64) (string, string) {
	if price > ma20 && ma20 > ma50 && ma50 > ma200 {
		return "strong_up", "strong"
	}
	if price > ma50 && ma50 > ma200 {
		return "up", "medium"
	}
	if price < ma20 && ma20 < ma50 && ma50 < ma200 {
		return "strong_down", "strong"
	}
	if price < ma50 && ma50 < ma200 {
		return "down", "medium"
	}
	return "neutral", "weak"
}
