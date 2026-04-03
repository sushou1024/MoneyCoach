package app

import (
	"context"
	"strings"
	"time"
)

func (s *Server) buildPlanChartSeries(ctx context.Context, snapshot PortfolioSnapshot, plan ReportStrategy) []any {
	if snapshot.ID == "" {
		return []any{}
	}
	assetKey := strings.TrimSpace(plan.AssetKey)
	assetType := strings.TrimSpace(plan.AssetType)
	if assetType == "" {
		assetType = assetTypeFromAssetKey(assetKey)
	}
	asOf := snapshot.ValuationAsOf
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	if assetType == "portfolio" || strings.HasPrefix(assetKey, "portfolio:") || assetKey == "" {
		_, holdings, err := s.loadSnapshotWithHoldings(ctx, snapshot.ID)
		if err != nil {
			return []any{}
		}
		seriesByAssetKey := fetchPriceSeriesAsOf(ctx, s.market, holdings, asOf)
		eligible := buildEligibleReturnSeries(holdings, seriesByAssetKey, 20)
		portfolioReturns := portfolioReturnsFromIntersection(eligible)
		priceSeries := tailPricePoints(portfolioPriceSeries(portfolioReturns), defaultSupportLookback)
		return chartSeriesFromPrices(priceSeries)
	}

	if assetType == "forex" {
		return []any{}
	}

	points := s.fetchPlanAssetSeries(ctx, plan, asOf)
	if len(points) == 0 {
		return []any{}
	}
	if len(points) > defaultSupportLookback {
		points = points[len(points)-defaultSupportLookback:]
	}
	return chartSeriesFromOHLC(points, snapshot.ValuationAsOf)
}

func (s *Server) fetchPlanAssetSeries(ctx context.Context, plan ReportStrategy, asOf time.Time) []ohlcPoint {
	if s.market == nil {
		return nil
	}
	start, end := planChartRange(asOf)
	assetType := strings.TrimSpace(plan.AssetType)
	assetKey := strings.TrimSpace(plan.AssetKey)
	if assetType == "" {
		assetType = assetTypeFromAssetKey(assetKey)
	}
	switch assetType {
	case "crypto":
		coinID := strings.TrimPrefix(assetKey, "crypto:cg:")
		if coinID == "" {
			return nil
		}
		return fetchCoinGeckoOHLCRange(ctx, s.market, coinID, start, end)
	case "stock":
		symbol := strings.TrimSpace(plan.Symbol)
		if symbol == "" {
			return nil
		}
		return fetchMarketstackSeriesRange(ctx, s.market, symbol, assetKey, start, end)
	default:
		return nil
	}
}

func planChartRange(asOf time.Time) (time.Time, time.Time) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	end := asOf.UTC()
	start := end.AddDate(0, 0, -defaultSupportLookback)
	return start, end
}

func chartSeriesFromPrices(points []pricePoint) []any {
	out := make([]any, 0, len(points))
	for _, point := range points {
		out = append(out, []any{time.Unix(point.Timestamp, 0).UTC().Format("2006-01-02"), roundTo(point.Value, 4)})
	}
	return out
}

func chartSeriesFromOHLC(points []ohlcPoint, _ time.Time) []any {
	out := make([]any, 0, len(points))
	for _, point := range points {
		out = append(out, []any{time.Unix(point.Timestamp, 0).UTC().Format("2006-01-02"), roundTo(point.Close, 4)})
	}
	return out
}
