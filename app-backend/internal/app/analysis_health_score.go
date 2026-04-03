package app

import "math"

func computeHealthScoreBaseline(metrics portfolioMetrics) int {
	baseline := 60.0
	if metrics.PricedCoveragePct < 0.80 {
		baseline -= (0.80 - metrics.PricedCoveragePct) * 60
	}
	if metrics.CashPct > 0.15 {
		baseline -= (metrics.CashPct - 0.15) * 100
	}
	if metrics.TopAssetPct > 0.40 {
		baseline -= (metrics.TopAssetPct - 0.40) * 100
	}
	if metrics.Volatility30dAnnualized > 0.70 {
		baseline -= (metrics.Volatility30dAnnualized - 0.70) * 50
	}
	if metrics.MaxDrawdown90d > 0.35 {
		baseline -= (metrics.MaxDrawdown90d - 0.35) * 80
	}
	if metrics.AvgPairwiseCorr > 0.70 {
		baseline -= (metrics.AvgPairwiseCorr - 0.70) * 60
	}
	baseline = math.Min(baseline, 90)
	if metrics.PricedCoveragePct < 0.60 {
		baseline = math.Min(baseline, 45)
	}
	if metrics.NonCashPricedValueUSD < 250 {
		baseline = math.Min(baseline, 49)
	}
	return int(math.Round(baseline))
}
