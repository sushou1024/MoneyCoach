package app

import (
	"context"
	"time"
)

const (
	insightTypePortfolioWatch = "portfolio_watch"
	insightTypeActionAlert    = "action_alert"
	insightTypeMarketAlpha    = "market_alpha"
	marketAlphaSignalType     = "oversold"
)

var defaultStockWatchlist = []string{"SPY", "QQQ", "AAPL", "MSFT", "AMZN", "NVDA", "GOOGL", "META", "TSLA"}

type priceSnapshot struct {
	Price     float64
	Timeframe string
	Timestamp time.Time
	Source    string
}

func buildInsights(ctx context.Context, market *marketClient, holdings []portfolioHolding, plans []lockedPlan, planStates map[string]*planStateData, seriesByAssetKey map[string][]ohlcPoint, futuresByAssetKey map[string]futuresPremiumIndex, riskByID map[string]string, outputLanguage string, baseCurrency string, rateFromUSD float64, now time.Time) ([]insightItem, []indicatorSnapshot) {
	portfolioWatch := buildPortfolioWatchSignals(ctx, market, holdings, plans, seriesByAssetKey, riskByID, outputLanguage, baseCurrency, rateFromUSD, now)
	actionAlerts := buildActionAlertSignals(ctx, market, holdings, plans, planStates, seriesByAssetKey, futuresByAssetKey, riskByID, outputLanguage, baseCurrency, rateFromUSD, now)
	marketAlpha, indicators := buildMarketAlphaSignals(ctx, market, holdings, seriesByAssetKey, outputLanguage, now)

	items := append(append(portfolioWatch, actionAlerts...), marketAlpha...)
	sortInsights(items)
	for i := range items {
		items[i].CTAPayload = buildInsightCTAPayload(items[i])
	}
	return items, indicators
}
