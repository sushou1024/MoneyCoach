package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

func fetchPriceSeries(ctx context.Context, market *marketClient, holdings []portfolioHolding) map[string][]ohlcPoint {
	return fetchPriceSeriesAsOf(ctx, market, holdings, time.Now().UTC())
}

func fetchPriceSeriesAsOf(ctx context.Context, market *marketClient, holdings []portfolioHolding, asOf time.Time) map[string][]ohlcPoint {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	end := asOf.UTC()
	start := end.AddDate(0, 0, -defaultSupportLookback)
	return fetchPriceSeriesRange(ctx, market, holdings, start, end)
}

func fetchPriceSeriesRange(ctx context.Context, market *marketClient, holdings []portfolioHolding, start, end time.Time) map[string][]ohlcPoint {
	seriesByAssetKey := make(map[string][]ohlcPoint)
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		switch holding.AssetType {
		case "crypto":
			if holding.CoinGeckoID == "" {
				continue
			}
			if _, ok := seriesByAssetKey[holding.AssetKey]; ok {
				continue
			}
			series := fetchCoinGeckoOHLCRange(ctx, market, holding.CoinGeckoID, start, end)
			if len(series) > 0 {
				seriesByAssetKey[holding.AssetKey] = series
			}
		case "stock":
			if holding.Symbol == "" {
				continue
			}
			if _, ok := seriesByAssetKey[holding.AssetKey]; ok {
				continue
			}
			series := fetchMarketstackSeriesRange(ctx, market, holding.Symbol, holding.AssetKey, start, end)
			if len(series) > 0 {
				seriesByAssetKey[holding.AssetKey] = series
			}
		}
	}

	return seriesByAssetKey
}

func fetchMarketstackSeries(ctx context.Context, market *marketClient, symbol, assetKey string) []ohlcPoint {
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -defaultSupportLookback)
	return fetchMarketstackSeriesRange(ctx, market, symbol, assetKey, start, end)
}

func fetchMarketstackSeriesRange(ctx context.Context, market *marketClient, symbol, assetKey string, start, end time.Time) []ohlcPoint {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" || market == nil {
		return nil
	}
	resolvedKey := resolveMarketstackAssetKey(ctx, market, symbol, assetKey)
	if resolvedKey == "" {
		return nil
	}
	query := candlestickQuery{
		Source:    candlestickSourceMarketstack,
		AssetType: "stock",
		AssetKey:  resolvedKey,
		Symbol:    symbol,
		Interval:  candlestickIntervalDaily,
	}
	points, err := fetchCandlestickSeries(ctx, marketDB(market), query, start, end, func(ctx context.Context, fetchStart, fetchEnd time.Time) ([]MarketCandlestick, error) {
		resp, err := market.marketstackEODRange(ctx, []string{symbol}, fetchStart, fetchEnd)
		if err != nil {
			return nil, err
		}
		rows := make([]MarketCandlestick, 0, len(resp.Data))
		for _, bar := range resp.Data {
			if bar.Symbol == "" {
				continue
			}
			parsed, ok := parseMarketstackTime(bar.Date)
			if !ok {
				continue
			}
			ts := dayStartUTCFromTime(parsed)
			rows = append(rows, MarketCandlestick{
				Source:    candlestickSourceMarketstack,
				AssetType: "stock",
				AssetKey:  resolvedKey,
				Symbol:    strings.ToUpper(strings.TrimSpace(bar.Symbol)),
				Interval:  candlestickIntervalDaily,
				Timestamp: ts,
				Open:      bar.Open,
				High:      bar.High,
				Low:       bar.Low,
				Close:     bar.Close,
				Volume:    bar.Volume,
				Currency:  marketstackPriceCurrency(bar.PriceCurrency, bar.Symbol, bar.Exchange),
			})
		}
		return rows, nil
	})
	if err != nil {
		if market.logger != nil {
			market.logger.Printf("marketstack candles error symbol=%s err=%v", symbol, err)
		}
		return nil
	}
	return points
}

func parseMarketstackTime(value string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05-0700",
		"2006-01-02",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func fetchCoinGeckoOHLC(ctx context.Context, market *marketClient, coinID string, days int) []ohlcPoint {
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -days)
	return fetchCoinGeckoOHLCRange(ctx, market, coinID, start, end)
}

func fetchCoinGeckoOHLCRange(ctx context.Context, market *marketClient, coinID string, start, end time.Time) []ohlcPoint {
	if start.IsZero() || end.IsZero() {
		end = time.Now().UTC()
		start = end.AddDate(0, 0, -defaultSupportLookback)
	}
	if market == nil || strings.TrimSpace(coinID) == "" {
		return nil
	}
	assetKey := "crypto:cg:" + strings.TrimSpace(coinID)
	query := candlestickQuery{
		Source:    candlestickSourceCoinGecko,
		AssetType: "crypto",
		AssetKey:  assetKey,
		Interval:  candlestickIntervalDaily,
		Currency:  "USD",
	}
	points, err := fetchCandlestickSeries(ctx, marketDB(market), query, start, end, func(ctx context.Context, fetchStart, fetchEnd time.Time) ([]MarketCandlestick, error) {
		resp, err := market.coinGeckoMarketChartRange(ctx, coinID, fetchStart, fetchEnd)
		if err != nil {
			return nil, err
		}
		if len(resp.Prices) == 0 {
			return nil, nil
		}
		daily, err := deriveDailyOHLCV(resp.Prices, resp.TotalVolumes)
		if err != nil {
			return nil, err
		}
		rows := make([]MarketCandlestick, 0, len(daily))
		for _, day := range daily {
			rows = append(rows, MarketCandlestick{
				Source:    candlestickSourceCoinGecko,
				AssetType: "crypto",
				AssetKey:  assetKey,
				Interval:  candlestickIntervalDaily,
				Timestamp: day.DayStart,
				Open:      day.Open,
				High:      day.High,
				Low:       day.Low,
				Close:     day.Close,
				Volume:    day.Volume,
				Currency:  "USD",
			})
		}
		return rows, nil
	})
	if err != nil {
		if market.logger != nil {
			market.logger.Printf("coingecko candles error coin_id=%s err=%v", coinID, err)
		}
		return nil
	}
	return points
}

type dailyOHLCV struct {
	DayStart time.Time
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
}

type dailyAccumulator struct {
	dailyOHLCV
	openSet    bool
	volumeTime int64
}

func deriveDailyOHLCV(prices [][]float64, volumes [][]float64) ([]dailyOHLCV, error) {
	if len(prices) == 0 {
		return nil, fmt.Errorf("no price data")
	}
	acc := map[string]*dailyAccumulator{}
	for _, entry := range prices {
		if len(entry) < 2 {
			return nil, fmt.Errorf("unexpected price entry")
		}
		tsMs := int64(entry[0])
		price := entry[1]
		dayStart := dayStartUTC(tsMs)
		key := dayStart.Format("2006-01-02")
		bucket := acc[key]
		if bucket == nil {
			bucket = &dailyAccumulator{dailyOHLCV: dailyOHLCV{DayStart: dayStart}}
			acc[key] = bucket
		}
		if !bucket.openSet {
			bucket.Open = price
			bucket.High = price
			bucket.Low = price
			bucket.Close = price
			bucket.openSet = true
			continue
		}
		if price > bucket.High {
			bucket.High = price
		}
		if price < bucket.Low {
			bucket.Low = price
		}
		bucket.Close = price
	}

	for _, entry := range volumes {
		if len(entry) < 2 {
			return nil, fmt.Errorf("unexpected volume entry")
		}
		tsMs := int64(entry[0])
		volume := entry[1]
		dayStart := dayStartUTC(tsMs)
		key := dayStart.Format("2006-01-02")
		bucket := acc[key]
		if bucket == nil {
			continue
		}
		if tsMs >= bucket.volumeTime {
			bucket.Volume = volume
			bucket.volumeTime = tsMs
		}
	}

	daily := make([]dailyOHLCV, 0, len(acc))
	for _, bucket := range acc {
		daily = append(daily, bucket.dailyOHLCV)
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].DayStart.Before(daily[j].DayStart) })
	return daily, nil
}

func dayStartUTC(tsMs int64) time.Time {
	ts := time.Unix(0, tsMs*int64(time.Millisecond)).UTC()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
}
