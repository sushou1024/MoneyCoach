package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	cacheKindMarketstackTick = "marketstack_ticker"
	cacheKindCandlestickTail = "candlestick_tail"
)

const (
	cacheTTLMarketstackTick = 24 * time.Hour
	cacheTTLCandlestickTail = 6 * time.Hour
)

const (
	cacheStatusHit   = "hit"
	cacheStatusEmpty = "empty"
	cacheStatusError = "error"
)

const (
	cacheTTLNegativeError = 10 * time.Minute
)

type cacheReadStatus int

const (
	cacheReadMiss cacheReadStatus = iota
	cacheReadHit
	cacheReadNegative
)

var errMarketDataCacheNegative = errors.New("market data cache negative")

type marketCacheStore struct {
	db *gorm.DB
}

func newMarketCacheStore(db *gorm.DB) *marketCacheStore {
	if db == nil {
		return nil
	}
	return &marketCacheStore{db: db}
}

func (c *marketCacheStore) getJSON(ctx context.Context, kind, key string, out any) (cacheReadStatus, error) {
	if c == nil || out == nil {
		return cacheReadMiss, nil
	}
	var entry MarketDataCache
	now := time.Now().UTC()
	err := c.db.WithContext(ctx).
		Where("cache_kind = ? AND cache_key = ? AND expires_at > ?", kind, key, now).
		First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return cacheReadMiss, nil
		}
		return cacheReadMiss, err
	}
	status := strings.TrimSpace(entry.CacheStatus)
	if status == "" {
		status = cacheStatusHit
	}
	if status != cacheStatusHit {
		return cacheReadNegative, nil
	}
	if len(entry.Payload) == 0 {
		return cacheReadNegative, nil
	}
	if err := json.Unmarshal(entry.Payload, out); err != nil {
		return cacheReadMiss, err
	}
	return cacheReadHit, nil
}

func (c *marketCacheStore) setJSON(ctx context.Context, kind, key string, value any, ttl time.Duration) error {
	if c == nil || ttl <= 0 {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	entry := MarketDataCache{
		CacheKind:   kind,
		CacheKey:    key,
		CacheStatus: cacheStatusHit,
		Payload:     datatypes.JSON(raw),
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return c.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cache_kind"}, {Name: "cache_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"cache_status", "payload", "expires_at", "updated_at"}),
	}).Create(&entry).Error
}

func (c *marketCacheStore) setNegative(ctx context.Context, kind, key, status string, ttl time.Duration) error {
	if c == nil || ttl <= 0 {
		return nil
	}
	status = strings.TrimSpace(status)
	if status != cacheStatusEmpty && status != cacheStatusError {
		status = cacheStatusEmpty
	}
	now := time.Now().UTC()
	entry := MarketDataCache{
		CacheKind:   kind,
		CacheKey:    key,
		CacheStatus: status,
		Payload:     nil,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return c.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cache_kind"}, {Name: "cache_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"cache_status", "payload", "expires_at", "updated_at"}),
	}).Create(&entry).Error
}

func cacheKey(parts ...string) string {
	return strings.Join(parts, "|")
}

func cacheKeyMarketstackTicker(symbol string) string {
	return cacheKey("marketstack_ticker", strings.ToUpper(strings.TrimSpace(symbol)))
}
