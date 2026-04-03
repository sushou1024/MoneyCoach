package app

import "math"

func returnsByTimestampFromPoints(points []ohlcPoint) map[int64]float64 {
	returns := make(map[int64]float64)
	for i := 1; i < len(points); i++ {
		prev := points[i-1].Close
		curr := points[i].Close
		if prev <= 0 || curr <= 0 {
			continue
		}
		returns[points[i].Timestamp] = math.Log(curr / prev)
	}
	return returns
}

func overlappingReturns(a, b map[int64]float64) ([]float64, []float64) {
	xs := make([]float64, 0)
	ys := make([]float64, 0)
	for ts, value := range a {
		if other, ok := b[ts]; ok {
			xs = append(xs, value)
			ys = append(ys, other)
		}
	}
	return xs, ys
}

func pearsonCorrelation(xs, ys []float64) float64 {
	if len(xs) != len(ys) || len(xs) < 2 {
		return 0
	}
	meanX := mean(xs)
	meanY := mean(ys)
	var sumXY, sumXX, sumYY float64
	for i := range xs {
		dx := xs[i] - meanX
		dy := ys[i] - meanY
		sumXY += dx * dy
		sumXX += dx * dx
		sumYY += dy * dy
	}
	if sumXX == 0 || sumYY == 0 {
		return 0
	}
	return sumXY / math.Sqrt(sumXX*sumYY)
}

func mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stddev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := mean(values)
	sum := 0.0
	for _, v := range values {
		delta := v - m
		sum += delta * delta
	}
	return math.Sqrt(sum / float64(len(values)-1))
}
