package app

import "time"

func newMarketSnapshot(valuationAsOf time.Time) MarketDataSnapshot {
	return MarketDataSnapshot{
		ID:            newID("snap"),
		ValuationAsOf: valuationAsOf,
		BaseCurrency:  "USD",
		CreatedAt:     valuationAsOf,
	}
}
