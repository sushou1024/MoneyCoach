package app

import (
	"context"
	"fmt"
	"time"
)

func buildMarketAlphaSignals(ctx context.Context, market *marketClient, holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, outputLanguage string, now time.Time) ([]insightItem, []indicatorSnapshot) {
	items := make([]insightItem, 0)
	indicators := make([]indicatorSnapshot, 0)
	if market == nil {
		return items, indicators
	}

	portfolioEligible := buildEligibleReturnSeries(holdings, seriesByAssetKey, 20)
	portfolioReturns := portfolioReturnsFromIntersection(portfolioEligible)
	portfolioReturnMap := returnsMap(portfolioReturns)

	candidates := marketAlphaUniverse(ctx, market, holdings)
	oerResp, _ := market.openExchangeLatest(ctx)
	oerRates := oerResp.Rates
	if oerRates == nil {
		oerRates = map[string]float64{}
	}
	minCloses := bollingerPeriod
	if rsiPeriod+1 > minCloses {
		minCloses = rsiPeriod + 1
	}

	for _, candidate := range candidates {
		series, ok := indicatorSeriesForHolding(ctx, market, candidate, seriesByAssetKey)
		if !ok || len(series.Points) < minCloses {
			continue
		}
		closes := extractCloses(series.Points)
		rsi, ok := computeRSI(closes, rsiPeriod)
		if !ok {
			continue
		}
		upper, lower, ok := computeBollinger(closes, bollingerPeriod, bollingerStdDev)
		if !ok {
			continue
		}
		last := series.Points[len(series.Points)-1]
		lastClose := closes[len(closes)-1]
		indicators = append(indicators, indicatorSnapshot{
			Asset:     candidate.Symbol,
			AssetKey:  candidate.AssetKey,
			AssetType: candidate.AssetType,
			Interval:  series.Interval,
			Source:    series.Source,
			RSI:       rsi,
			UpperBand: upper,
			LowerBand: lower,
			Close:     lastClose,
			Timestamp: last.Timestamp,
		})

		if !(rsi <= 30 && lastClose <= lower) {
			continue
		}
		severity := "Low"
		if rsi <= 20 || lastClose <= lower*0.99 {
			severity = "High"
		} else if rsi <= 25 {
			severity = "Medium"
		}

		assetReturns := returnsByTimestamp(ctx, market, seriesByAssetKey, candidate)
		beta := betaToPortfolio(assetReturns, portfolioReturnMap)

		assetRef := candidate.AssetKey
		if assetRef == "" {
			assetRef = candidate.Symbol
		}
		closeTime := time.Unix(last.Timestamp, 0).UTC()
		reason := insightCopy(outputLanguage, copyMarketAlphaReason, formatFloat(rsi, 1), series.Interval)
		action := insightCopy(outputLanguage, copyMarketAlphaAction)
		items = append(items, insightItem{
			ID:              newID("ins"),
			Type:            insightTypeMarketAlpha,
			Asset:           candidate.Symbol,
			AssetKey:        candidate.AssetKey,
			Timeframe:       series.Interval,
			Severity:        severity,
			TriggerReason:   reason,
			TriggerKey:      fmt.Sprintf("market_alpha:%s:%s:%s:%s", assetRef, series.Interval, marketAlphaSignalType, closeTime.Format(time.RFC3339)),
			SuggestedAction: action,
			BetaToPortfolio: beta,
			CreatedAt:       now,
			ExpiresAt:       now.Add(24 * time.Hour),
		})
	}

	return items, indicators
}
