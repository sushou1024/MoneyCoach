package app

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	candlestickSourceCoinGecko   = "coingecko"
	candlestickSourceMarketstack = "marketstack"
	candlestickSourceBinance     = "binance"

	candlestickIntervalDaily = "1d"
)

const fxRateSourceOpenExchange = "openexchangerates"

const candlestickStaleTailWindow = 72 * time.Hour

type candlestickQuery struct {
	Source    string
	AssetType string
	AssetKey  string
	Symbol    string
	Interval  string
	Currency  string
}

type candlestickRange struct {
	Start time.Time
	End   time.Time
}

type candlestickFetchFunc func(ctx context.Context, start, end time.Time) ([]MarketCandlestick, error)

func marketDB(market *marketClient) *gorm.DB {
	if market == nil || market.cache == nil {
		return nil
	}
	return market.cache.db
}

func normalizeCandlestickRange(start, end time.Time, interval string) (time.Time, time.Time) {
	start = start.UTC()
	end = end.UTC()
	if interval == candlestickIntervalDaily {
		start = dayStartUTCFromTime(start)
		end = dayStartUTCFromTime(end)
	}
	return start, end
}

func dayStartUTCFromTime(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func intervalDuration(interval string) (time.Duration, bool) {
	switch interval {
	case "1d":
		return 24 * time.Hour, true
	case "4h":
		return 4 * time.Hour, true
	default:
		return 0, false
	}
}

func candlestickTailCacheKey(query candlestickQuery, end time.Time) string {
	end = end.UTC()
	return cacheKey(query.Source, query.AssetKey, query.Interval, query.Currency, end.Format("2006-01-02"))
}

func isCandlestickTailRange(existing []MarketCandlestick, r candlestickRange, end time.Time) bool {
	if len(existing) == 0 || end.IsZero() {
		return false
	}
	return r.End.Equal(end)
}

func shouldSkipCandlestickTail(ctx context.Context, cache *marketCacheStore, query candlestickQuery, end time.Time) bool {
	if end.IsZero() {
		return false
	}
	now := time.Now().UTC()
	if end.After(now) {
		return true
	}
	if now.Sub(end) <= candlestickStaleTailWindow {
		return true
	}
	if cache == nil {
		return false
	}
	key := candlestickTailCacheKey(query, end)
	var marker struct{}
	status, err := cache.getJSON(ctx, cacheKindCandlestickTail, key, &marker)
	return err == nil && status == cacheReadNegative
}

func fetchCandlestickSeries(ctx context.Context, db *gorm.DB, query candlestickQuery, start, end time.Time, fetcher candlestickFetchFunc) ([]ohlcPoint, error) {
	start, end = normalizeCandlestickRange(start, end, query.Interval)
	if start.IsZero() || end.IsZero() {
		return nil, nil
	}
	if db == nil {
		rows, err := fetcher(ctx, start, end)
		if err != nil {
			return nil, err
		}
		return candlestickRowsToPoints(rows), nil
	}

	existing, err := loadCandlesticks(ctx, db, query, start, end)
	if err != nil {
		return nil, err
	}

	cache := newMarketCacheStore(db)
	missing := missingCandlestickRanges(existing, start, end, query.Interval)
	if len(existing) > 0 && len(missing) > 0 {
		filtered := missing[:0]
		for _, r := range missing {
			if isCandlestickTailRange(existing, r, end) && shouldSkipCandlestickTail(ctx, cache, query, end) {
				continue
			}
			filtered = append(filtered, r)
		}
		missing = filtered
	}
	fetched := make([]MarketCandlestick, 0)
	for _, r := range missing {
		if r.Start.After(r.End) {
			continue
		}
		rows, err := fetcher(ctx, r.Start, r.End)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			if isCandlestickTailRange(existing, r, end) && cache != nil {
				_ = cache.setNegative(ctx, cacheKindCandlestickTail, candlestickTailCacheKey(query, end), cacheStatusEmpty, cacheTTLCandlestickTail)
			}
			continue
		}
		fetched = append(fetched, rows...)
	}
	if len(fetched) > 0 {
		if err := insertCandlesticks(ctx, db, fetched); err != nil {
			return nil, err
		}
	}

	merged := mergeCandlestickRows(existing, fetched)
	return candlestickRowsToPoints(merged), nil
}

func loadCandlesticks(ctx context.Context, db *gorm.DB, query candlestickQuery, start, end time.Time) ([]MarketCandlestick, error) {
	if db == nil {
		return nil, nil
	}
	var rows []MarketCandlestick
	err := db.WithContext(ctx).
		Where("source = ? AND asset_key = ? AND interval = ? AND timestamp >= ? AND timestamp <= ?", query.Source, query.AssetKey, query.Interval, start, end).
		Order("timestamp asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func loadLatestCandlesticks(ctx context.Context, db *gorm.DB, query candlestickQuery, limit int) ([]MarketCandlestick, error) {
	if db == nil || limit <= 0 {
		return nil, nil
	}
	var rows []MarketCandlestick
	err := db.WithContext(ctx).
		Where("source = ? AND asset_key = ? AND interval = ?", query.Source, query.AssetKey, query.Interval).
		Order("timestamp desc").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Timestamp.Before(rows[j].Timestamp) })
	return rows, nil
}

func missingCandlestickRanges(rows []MarketCandlestick, start, end time.Time, interval string) []candlestickRange {
	if start.After(end) {
		return nil
	}
	if len(rows) == 0 {
		return []candlestickRange{{Start: start, End: end}}
	}
	earliest := rows[0].Timestamp
	latest := rows[len(rows)-1].Timestamp
	step, ok := intervalDuration(interval)
	if !ok || step == 0 {
		step = time.Nanosecond
	}
	ranges := make([]candlestickRange, 0, 2)
	if earliest.After(start) {
		ranges = append(ranges, candlestickRange{Start: start, End: earliest.Add(-step)})
	}
	if latest.Before(end) {
		ranges = append(ranges, candlestickRange{Start: latest.Add(step), End: end})
	}
	return ranges
}

func insertCandlesticks(ctx context.Context, db *gorm.DB, rows []MarketCandlestick) error {
	if db == nil || len(rows) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for i := range rows {
		if rows[i].CreatedAt.IsZero() {
			rows[i].CreatedAt = now
		}
		if rows[i].UpdatedAt.IsZero() {
			rows[i].UpdatedAt = now
		}
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "source"},
			{Name: "asset_key"},
			{Name: "interval"},
			{Name: "timestamp"},
		},
		DoNothing: true,
	}).Create(&rows).Error
}

func mergeCandlestickRows(existing, fetched []MarketCandlestick) []MarketCandlestick {
	if len(fetched) == 0 {
		return existing
	}
	byTimestamp := make(map[int64]MarketCandlestick, len(existing)+len(fetched))
	for _, row := range existing {
		byTimestamp[row.Timestamp.Unix()] = row
	}
	for _, row := range fetched {
		byTimestamp[row.Timestamp.Unix()] = row
	}
	merged := make([]MarketCandlestick, 0, len(byTimestamp))
	for _, row := range byTimestamp {
		merged = append(merged, row)
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Timestamp.Before(merged[j].Timestamp) })
	return merged
}

func trimCandlestickRows(rows []MarketCandlestick, limit int) []MarketCandlestick {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}
	return rows[len(rows)-limit:]
}

func candlestickRowsToPoints(rows []MarketCandlestick) []ohlcPoint {
	points := make([]ohlcPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, ohlcPoint{
			Timestamp: row.Timestamp.Unix(),
			Open:      row.Open,
			High:      row.High,
			Low:       row.Low,
			Close:     row.Close,
			Volume:    row.Volume,
		})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp < points[j].Timestamp })
	return points
}

func loadFXRatesForDate(ctx context.Context, db *gorm.DB, date time.Time, baseCurrency string) (oerLatestResponse, bool, error) {
	if db == nil {
		return oerLatestResponse{}, false, nil
	}
	base := strings.ToUpper(strings.TrimSpace(baseCurrency))
	if base == "" {
		base = "USD"
	}
	rateDate := dayStartUTCFromTime(date)
	var row FXDailyRate
	if err := db.WithContext(ctx).Where("rate_date = ? AND base_currency = ?", rateDate, base).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return oerLatestResponse{}, false, nil
		}
		return oerLatestResponse{}, false, err
	}
	var rates map[string]float64
	if len(row.Rates) > 0 {
		if err := json.Unmarshal(row.Rates, &rates); err != nil {
			return oerLatestResponse{}, false, err
		}
	}
	return oerLatestResponse{Base: row.BaseCurrency, Timestamp: row.Timestamp, Rates: rates}, true, nil
}

func storeFXRates(ctx context.Context, db *gorm.DB, resp oerLatestResponse, source string) error {
	if db == nil {
		return nil
	}
	if len(resp.Rates) == 0 {
		return nil
	}
	base := strings.ToUpper(strings.TrimSpace(resp.Base))
	if base == "" {
		base = "USD"
	}
	timestamp := resp.Timestamp
	if timestamp == 0 {
		timestamp = time.Now().UTC().Unix()
	}
	rateDate := dayStartUTCFromTime(time.Unix(timestamp, 0))
	raw, err := json.Marshal(resp.Rates)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	row := FXDailyRate{
		RateDate:     rateDate,
		BaseCurrency: base,
		Rates:        datatypes.JSON(raw),
		Source:       source,
		Timestamp:    timestamp,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "rate_date"}, {Name: "base_currency"}},
		DoNothing: true,
	}).Create(&row).Error
}
