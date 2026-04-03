package app

import "math"

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

func buildRadarChart(metrics portfolioMetrics, alpha30d float64) paidRadarChart {
	liquidity := clamp((metrics.CashPct/0.20)*100, 0, 100)
	diversification := clamp(100-math.Max(0, (metrics.TopAssetPct-0.20)*125), 0, 100)
	drawdown := clamp(100-metrics.MaxDrawdown90d*200, 0, 100)
	alpha := clamp(50+alpha30d*500, 0, 100)
	return paidRadarChart{
		Liquidity:       int(math.Round(liquidity)),
		Diversification: int(math.Round(diversification)),
		Alpha:           int(math.Round(alpha)),
		Drawdown:        int(math.Round(drawdown)),
	}
}
