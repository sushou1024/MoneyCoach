package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (m *marketClient) binanceKlines(ctx context.Context, symbol, interval string, limit int) ([]ohlcPoint, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" || limit <= 0 {
		return nil, nil
	}
	if m == nil {
		return nil, fmt.Errorf("market client unavailable")
	}
	query := candlestickQuery{
		Source:    candlestickSourceBinance,
		AssetType: "crypto",
		AssetKey:  binanceAssetKey(symbol),
		Symbol:    symbol,
		Interval:  interval,
		Currency:  binanceQuoteCurrency(symbol),
	}
	db := marketDB(m)
	if db == nil {
		return m.fetchBinanceKlinesRange(ctx, symbol, interval, limit, time.Time{}, time.Time{})
	}

	existing, err := loadLatestCandlesticks(ctx, db, query, limit)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if len(existing) >= limit && binanceSeriesFresh(existing, interval, now) {
		return candlestickRowsToPoints(existing), nil
	}

	fetchStart := binanceFetchStart(existing, interval, now, limit)
	points, err := m.fetchBinanceKlinesRange(ctx, symbol, interval, limit, fetchStart, now)
	if err != nil {
		return nil, err
	}
	rows := binanceCandlesticksFromPoints(query, points)
	if err := insertCandlesticks(ctx, db, rows); err != nil {
		return nil, err
	}
	merged := mergeCandlestickRows(existing, rows)
	merged = trimCandlestickRows(merged, limit)
	return candlestickRowsToPoints(merged), nil
}

func (m *marketClient) fetchBinanceKlinesRange(ctx context.Context, symbol, interval string, limit int, start, end time.Time) ([]ohlcPoint, error) {
	query := url.Values{}
	query.Set("symbol", symbol)
	query.Set("interval", interval)
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if !start.IsZero() {
		query.Set("startTime", strconv.FormatInt(start.UnixMilli(), 10))
	}
	if !end.IsZero() {
		query.Set("endTime", strconv.FormatInt(end.UnixMilli(), 10))
	}
	endpoint := fmt.Sprintf("%s/api/v3/klines?%s", strings.TrimRight(m.cfg.BinanceAPIBaseURL, "/"), query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance status %d", resp.StatusCode)
	}
	var raw [][]any
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}
	points := make([]ohlcPoint, 0, len(raw))
	for _, entry := range raw {
		point, ok := parseBinanceKline(entry)
		if !ok {
			continue
		}
		points = append(points, point)
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp < points[j].Timestamp })
	return points, nil
}

func binanceAssetKey(symbol string) string {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return ""
	}
	return "crypto:binance:" + symbol
}

func binanceQuoteCurrency(symbol string) string {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return ""
	}
	quotes := []string{"USDT", "USDC", "BUSD", "USD"}
	for _, quote := range quotes {
		if strings.HasSuffix(symbol, quote) {
			return quote
		}
	}
	return ""
}

func binanceCandlesticksFromPoints(query candlestickQuery, points []ohlcPoint) []MarketCandlestick {
	if len(points) == 0 {
		return nil
	}
	rows := make([]MarketCandlestick, 0, len(points))
	for _, point := range points {
		rows = append(rows, MarketCandlestick{
			Source:    query.Source,
			AssetType: query.AssetType,
			AssetKey:  query.AssetKey,
			Symbol:    query.Symbol,
			Interval:  query.Interval,
			Timestamp: time.Unix(point.Timestamp, 0).UTC(),
			Open:      point.Open,
			High:      point.High,
			Low:       point.Low,
			Close:     point.Close,
			Volume:    point.Volume,
			Currency:  query.Currency,
		})
	}
	return rows
}

func binanceSeriesFresh(rows []MarketCandlestick, interval string, now time.Time) bool {
	if len(rows) == 0 {
		return false
	}
	step, ok := intervalDuration(interval)
	if !ok || step == 0 {
		return false
	}
	latest := rows[len(rows)-1].Timestamp
	return latest.After(now.Add(-step))
}

func binanceFetchStart(rows []MarketCandlestick, interval string, now time.Time, limit int) time.Time {
	step, ok := intervalDuration(interval)
	if !ok || step == 0 {
		return time.Time{}
	}
	targetStart := now.Add(-time.Duration(limit) * step)
	if len(rows) == 0 {
		return targetStart
	}
	latest := rows[len(rows)-1].Timestamp
	if latest.After(targetStart) {
		return latest.Add(time.Millisecond)
	}
	return targetStart
}
