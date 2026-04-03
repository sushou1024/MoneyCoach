package app

import "math"

func simpleMovingAverage(closes []float64, window int) (float64, bool) {
	if len(closes) < window || window <= 0 {
		return 0, false
	}
	sum := 0.0
	for _, value := range closes[len(closes)-window:] {
		sum += value
	}
	return sum / float64(window), true
}

func assetVolatilityDaily(closes []float64) (float64, bool) {
	if len(closes) < 21 {
		return 0, false
	}
	returns := logReturnsFromCloses(closes)
	last := tailValues(returns, defaultVolatilityLookback)
	if len(last) < 2 {
		return 0, false
	}
	return stddev(last), true
}

func logReturnsFromCloses(closes []float64) []float64 {
	if len(closes) < 2 {
		return nil
	}
	returns := make([]float64, 0, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		prev := closes[i-1]
		curr := closes[i]
		if prev <= 0 || curr <= 0 {
			continue
		}
		returns = append(returns, math.Log(curr/prev))
	}
	return returns
}

func tailValues(values []float64, n int) []float64 {
	if n <= 0 || len(values) <= n {
		out := make([]float64, len(values))
		copy(out, values)
		return out
	}
	out := make([]float64, n)
	copy(out, values[len(values)-n:])
	return out
}

func annualizationFactorForAsset(assetType string) float64 {
	if assetType == "crypto" {
		return math.Sqrt(365)
	}
	return math.Sqrt(252)
}

func computeRSI(closes []float64, period int) (float64, bool) {
	if len(closes) <= period {
		return 0, false
	}
	gain := 0.0
	loss := 0.0
	for i := 1; i <= period; i++ {
		delta := closes[i] - closes[i-1]
		if delta >= 0 {
			gain += delta
		} else {
			loss -= delta
		}
	}
	avgGain := gain / float64(period)
	avgLoss := loss / float64(period)
	for i := period + 1; i < len(closes); i++ {
		delta := closes[i] - closes[i-1]
		if delta >= 0 {
			avgGain = (avgGain*float64(period-1) + delta) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) - delta) / float64(period)
		}
	}
	if avgLoss == 0 {
		return 100, true
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs)), true
}

func computeBollinger(closes []float64, period int, stdDev float64) (float64, float64, bool) {
	if len(closes) < period {
		return 0, 0, false
	}
	window := closes[len(closes)-period:]
	avg := mean(window)
	variance := 0.0
	for _, v := range window {
		delta := v - avg
		variance += delta * delta
	}
	std := math.Sqrt(variance / float64(period))
	return avg + stdDev*std, avg - stdDev*std, true
}

func extractCloses(points []ohlcPoint) []float64 {
	closes := make([]float64, 0, len(points))
	for _, point := range points {
		closes = append(closes, point.Close)
	}
	return closes
}
