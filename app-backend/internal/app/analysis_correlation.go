package app

import (
	"sort"
	"strings"
)

func avgPairwiseCorrelation(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) float64 {
	eligible := make([]portfolioHolding, 0)
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
		if len(points) < 20 {
			continue
		}
		eligible = append(eligible, holding)
	}
	if len(eligible) < 2 {
		return 0
	}
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].ValueUSD > eligible[j].ValueUSD
	})
	if len(eligible) > 5 {
		eligible = eligible[:5]
	}

	returnsByAsset := make(map[string]map[int64]float64, len(eligible))
	for _, holding := range eligible {
		points := seriesByAssetKey[holding.AssetKey]
		if len(points) > defaultSupportLookback {
			points = points[len(points)-defaultSupportLookback:]
		}
		returnsByAsset[holding.AssetKey] = returnsByTimestampFromPoints(points)
	}

	pairs := 0
	sum := 0.0
	for i := 0; i < len(eligible); i++ {
		for j := i + 1; j < len(eligible); j++ {
			xs, ys := overlappingReturns(returnsByAsset[eligible[i].AssetKey], returnsByAsset[eligible[j].AssetKey])
			if len(xs) < 20 {
				continue
			}
			corr := pearsonCorrelation(xs, ys)
			sum += corr
			pairs++
		}
	}
	if pairs == 0 {
		return 0
	}
	return sum / float64(pairs)
}
