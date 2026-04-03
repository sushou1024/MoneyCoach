package app

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFetchCandlestickSeriesIncremental(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	assetKey := "test:cg:" + ulid.Make().String()
	query := candlestickQuery{
		Source:    candlestickSourceCoinGecko,
		AssetType: "crypto",
		AssetKey:  assetKey,
		Interval:  candlestickIntervalDaily,
		Currency:  "USD",
	}
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	existing := MarketCandlestick{
		Source:    query.Source,
		AssetType: query.AssetType,
		AssetKey:  assetKey,
		Interval:  query.Interval,
		Timestamp: start,
		Open:      100,
		High:      110,
		Low:       95,
		Close:     105,
		Volume:    1000,
		Currency:  query.Currency,
	}
	if err := insertCandlesticks(ctx, db, []MarketCandlestick{existing}); err != nil {
		t.Fatalf("insert seed candlestick: %v", err)
	}
	t.Cleanup(func() {
		cleanupCandlesticks(ctx, db, assetKey)
	})

	var calls []candlestickRange
	fetcher := func(_ context.Context, fetchStart, fetchEnd time.Time) ([]MarketCandlestick, error) {
		t.Logf("fetcher called start=%s end=%s", fetchStart.Format(time.RFC3339), fetchEnd.Format(time.RFC3339))
		calls = append(calls, candlestickRange{Start: fetchStart, End: fetchEnd})
		rows := []MarketCandlestick{
			{
				Source:    query.Source,
				AssetType: query.AssetType,
				AssetKey:  assetKey,
				Interval:  query.Interval,
				Timestamp: start.Add(24 * time.Hour),
				Open:      105,
				High:      115,
				Low:       102,
				Close:     112,
				Volume:    900,
				Currency:  query.Currency,
			},
			{
				Source:    query.Source,
				AssetType: query.AssetType,
				AssetKey:  assetKey,
				Interval:  query.Interval,
				Timestamp: end,
				Open:      112,
				High:      120,
				Low:       108,
				Close:     118,
				Volume:    800,
				Currency:  query.Currency,
			},
		}
		return rows, nil
	}

	points, err := fetchCandlestickSeries(ctx, db, query, start, end, fetcher)
	if err != nil {
		t.Fatalf("fetchCandlestickSeries: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 fetch call, got %d", len(calls))
	}
	expectedStart := start.Add(24 * time.Hour)
	if !calls[0].Start.Equal(expectedStart) || !calls[0].End.Equal(end) {
		t.Fatalf("unexpected fetch range: %+v", calls[0])
	}
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}
	t.Logf("points: %#v", points)

	var count int64
	if err := db.WithContext(ctx).Model(&MarketCandlestick{}).Where("asset_key = ?", assetKey).Count(&count).Error; err != nil {
		t.Fatalf("count candlesticks: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 rows, got %d", count)
	}
}

func TestFetchCandlestickSeriesNoFetchWhenSufficient(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	assetKey := "test:cg:" + ulid.Make().String()
	query := candlestickQuery{
		Source:    candlestickSourceCoinGecko,
		AssetType: "crypto",
		AssetKey:  assetKey,
		Interval:  candlestickIntervalDaily,
		Currency:  "USD",
	}
	start := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	rows := []MarketCandlestick{
		{
			Source:    query.Source,
			AssetType: query.AssetType,
			AssetKey:  assetKey,
			Interval:  query.Interval,
			Timestamp: start,
			Open:      50,
			High:      55,
			Low:       45,
			Close:     52,
			Volume:    700,
			Currency:  query.Currency,
		},
		{
			Source:    query.Source,
			AssetType: query.AssetType,
			AssetKey:  assetKey,
			Interval:  query.Interval,
			Timestamp: start.Add(24 * time.Hour),
			Open:      52,
			High:      60,
			Low:       50,
			Close:     58,
			Volume:    650,
			Currency:  query.Currency,
		},
		{
			Source:    query.Source,
			AssetType: query.AssetType,
			AssetKey:  assetKey,
			Interval:  query.Interval,
			Timestamp: start.Add(48 * time.Hour),
			Open:      58,
			High:      62,
			Low:       55,
			Close:     60,
			Volume:    600,
			Currency:  query.Currency,
		},
	}
	if err := insertCandlesticks(ctx, db, rows); err != nil {
		t.Fatalf("insert seed candlesticks: %v", err)
	}
	t.Cleanup(func() {
		cleanupCandlesticks(ctx, db, assetKey)
	})

	calls := 0
	fetcher := func(_ context.Context, fetchStart, fetchEnd time.Time) ([]MarketCandlestick, error) {
		calls++
		t.Logf("unexpected fetcher call start=%s end=%s", fetchStart.Format(time.RFC3339), fetchEnd.Format(time.RFC3339))
		return nil, nil
	}
	end := start.Add(48 * time.Hour)
	points, err := fetchCandlestickSeries(ctx, db, query, start, end, fetcher)
	if err != nil {
		t.Fatalf("fetchCandlestickSeries: %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected no fetch calls, got %d", calls)
	}
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}
	t.Logf("points: %#v", points)
}

func TestFXDailyRatePersistence(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	resp := oerLatestResponse{
		Base:      "USD",
		Timestamp: now.Unix(),
		Rates: map[string]float64{
			"EUR": 0.92,
			"JPY": 150.5,
		},
	}
	if err := storeFXRates(ctx, db, resp, fxRateSourceOpenExchange); err != nil {
		t.Fatalf("storeFXRates: %v", err)
	}
	t.Cleanup(func() {
		cleanupFXRates(ctx, db, now, "USD")
	})

	loaded, ok, err := loadFXRatesForDate(ctx, db, now, "USD")
	if err != nil {
		t.Fatalf("loadFXRatesForDate: %v", err)
	}
	if !ok {
		t.Fatal("expected FX rates to be loaded")
	}
	if len(loaded.Rates) != len(resp.Rates) {
		t.Fatalf("unexpected rate count: %d", len(loaded.Rates))
	}
	if loaded.Rates["EUR"] != resp.Rates["EUR"] {
		t.Fatalf("unexpected EUR rate: %v", loaded.Rates["EUR"])
	}
	t.Logf("loaded rates: %#v", loaded.Rates)
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if dsn == "" {
		t.Fatalf("missing TEST_DATABASE_URL for database tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&MarketCandlestick{}, &FXDailyRate{}); err != nil {
		t.Fatalf("auto migrate test tables: %v", err)
	}
	t.Logf("connected test database")
	return db
}

func cleanupCandlesticks(ctx context.Context, db *gorm.DB, assetKey string) {
	if db == nil || assetKey == "" {
		return
	}
	_ = db.WithContext(ctx).Where("asset_key = ?", assetKey).Delete(&MarketCandlestick{}).Error
}

func cleanupFXRates(ctx context.Context, db *gorm.DB, date time.Time, baseCurrency string) {
	if db == nil {
		return
	}
	rateDate := dayStartUTCFromTime(date)
	base := strings.ToUpper(strings.TrimSpace(baseCurrency))
	if base == "" {
		base = "USD"
	}
	_ = db.WithContext(ctx).Where("rate_date = ? AND base_currency = ?", rateDate, base).Delete(&FXDailyRate{}).Error
}
