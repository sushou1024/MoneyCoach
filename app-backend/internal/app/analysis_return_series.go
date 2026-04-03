package app

import (
	"math"
	"sort"
	"strings"
)

type assetReturnSeries struct {
	AssetKey           string
	Holding            portfolioHolding
	Weight             float64
	ReturnsByTimestamp map[int64]float64
	Returns            []returnPoint
}

func buildEligibleReturnSeries(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, minPoints int) []assetReturnSeries {
	eligible := make([]assetReturnSeries, 0)
	total := 0.0
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		if strings.HasPrefix(holding.AssetKey, "manual:") {
			continue
		}
		points := seriesByAssetKey[holding.AssetKey]
		if len(points) < minPoints {
			continue
		}
		total += holding.ValueUSD
		eligible = append(eligible, assetReturnSeries{
			AssetKey:           holding.AssetKey,
			Holding:            holding,
			ReturnsByTimestamp: returnsByTimestampFromPoints(points),
		})
	}

	if total <= 0 {
		return nil
	}
	for i := range eligible {
		eligible[i].Weight = eligible[i].Holding.ValueUSD / total
		eligible[i].Returns = returnPointsFromMap(eligible[i].ReturnsByTimestamp)
	}
	return eligible
}

func portfolioReturnsFromIntersection(series []assetReturnSeries) []returnPoint {
	if len(series) == 0 {
		return nil
	}
	counts := make(map[int64]int)
	for _, asset := range series {
		for ts := range asset.ReturnsByTimestamp {
			counts[ts]++
		}
	}
	intersect := make([]int64, 0)
	for ts, count := range counts {
		if count == len(series) {
			intersect = append(intersect, ts)
		}
	}
	sort.Slice(intersect, func(i, j int) bool { return intersect[i] < intersect[j] })
	if len(intersect) == 0 {
		return nil
	}
	returns := make([]returnPoint, 0, len(intersect))
	for _, ts := range intersect {
		value := 0.0
		for _, asset := range series {
			value += asset.Weight * asset.ReturnsByTimestamp[ts]
		}
		returns = append(returns, returnPoint{Timestamp: ts, Value: value})
	}
	return returns
}

func returnPointsFromMap(values map[int64]float64) []returnPoint {
	points := make([]returnPoint, 0, len(values))
	for ts, value := range values {
		points = append(points, returnPoint{Timestamp: ts, Value: value})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp < points[j].Timestamp })
	return points
}

func returnsMap(points []returnPoint) map[int64]float64 {
	byTimestamp := make(map[int64]float64, len(points))
	for _, point := range points {
		byTimestamp[point.Timestamp] = point.Value
	}
	return byTimestamp
}

func portfolioPriceSeries(returns []returnPoint) []pricePoint {
	if len(returns) == 0 {
		return nil
	}
	value := 1.0
	series := make([]pricePoint, 0, len(returns))
	for _, ret := range returns {
		value *= math.Exp(ret.Value)
		series = append(series, pricePoint{Timestamp: ret.Timestamp, Value: value})
	}
	return series
}

func tailPricePoints(series []pricePoint, n int) []pricePoint {
	if len(series) <= n {
		return series
	}
	return series[len(series)-n:]
}

func lastNReturns(returns []returnPoint, n int) []float64 {
	if len(returns) <= n {
		values := make([]float64, 0, len(returns))
		for _, ret := range returns {
			values = append(values, ret.Value)
		}
		return values
	}
	values := make([]float64, 0, n)
	for _, ret := range returns[len(returns)-n:] {
		values = append(values, ret.Value)
	}
	return values
}

func maxDrawdown(series []pricePoint) float64 {
	if len(series) == 0 {
		return 0
	}
	peak := series[0].Value
	maxDD := 0.0
	for _, point := range series {
		if point.Value > peak {
			peak = point.Value
		}
		drawdown := (peak - point.Value) / peak
		if drawdown > maxDD {
			maxDD = drawdown
		}
	}
	return maxDD
}
