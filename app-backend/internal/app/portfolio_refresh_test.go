package app

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

type failRoundTripper struct{}

func (failRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("unexpected outbound http call during refresh test")
}

func TestRefreshActivePortfolioCreatesNewActiveSnapshot(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := db.AutoMigrate(
		&User{},
		&UserProfile{},
		&UserAssetOverride{},
		&MarketDataSnapshot{},
		&MarketDataSnapshotItem{},
		&PortfolioSnapshot{},
		&PortfolioHolding{},
	); err != nil {
		t.Fatalf("auto migrate refresh tables: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	userID := "user_refresh_test"
	oldSnapshotID := "pf_refresh_old"
	oldMarketSnapshotID := "snap_refresh_old"

	t.Cleanup(func() {
		var snapshotIDs []string
		var marketSnapshotIDs []string
		_ = db.WithContext(ctx).Model(&PortfolioSnapshot{}).Where("user_id = ?", userID).Pluck("id", &snapshotIDs).Error
		_ = db.WithContext(ctx).Model(&PortfolioSnapshot{}).Where("user_id = ?", userID).Pluck("market_data_snapshot_id", &marketSnapshotIDs).Error
		if len(snapshotIDs) > 0 {
			_ = db.WithContext(ctx).Where("portfolio_snapshot_id IN ?", snapshotIDs).Delete(&PortfolioHolding{}).Error
			_ = db.WithContext(ctx).Where("id IN ?", snapshotIDs).Delete(&PortfolioSnapshot{}).Error
		}
		if len(marketSnapshotIDs) > 0 {
			_ = db.WithContext(ctx).Where("market_data_snapshot_id IN ?", marketSnapshotIDs).Delete(&MarketDataSnapshotItem{}).Error
			_ = db.WithContext(ctx).Where("id IN ?", marketSnapshotIDs).Delete(&MarketDataSnapshot{}).Error
		}
		_ = db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserProfile{}).Error
		_ = db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserAssetOverride{}).Error
		_ = db.WithContext(ctx).Where("id = ?", userID).Delete(&User{}).Error
	})

	usdRates := oerLatestResponse{
		Base:      "USD",
		Timestamp: now.Unix(),
		Rates: map[string]float64{
			"USD": 1,
		},
	}
	if err := storeFXRates(ctx, db, usdRates, fxRateSourceOpenExchange); err != nil {
		t.Fatalf("store fx rates: %v", err)
	}

	if err := db.WithContext(ctx).Create(&User{
		ID:                      userID,
		ActivePortfolioSnapshot: &oldSnapshotID,
		CreatedAt:               now,
		UpdatedAt:               now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.WithContext(ctx).Create(&UserProfile{
		UserID:       userID,
		Language:     "en",
		Timezone:     "UTC",
		BaseCurrency: "USD",
	}).Error; err != nil {
		t.Fatalf("create user profile: %v", err)
	}
	if err := db.WithContext(ctx).Create(&MarketDataSnapshot{
		ID:            oldMarketSnapshotID,
		ValuationAsOf: now.Add(-2 * time.Hour),
		BaseCurrency:  "USD",
		CreatedAt:     now.Add(-2 * time.Hour),
	}).Error; err != nil {
		t.Fatalf("create old market snapshot: %v", err)
	}
	if err := db.WithContext(ctx).Create(&PortfolioSnapshot{
		ID:                   oldSnapshotID,
		UserID:               userID,
		SourceUploadBatchID:  "upload_refresh_seed",
		MarketDataSnapshotID: oldMarketSnapshotID,
		ValuationAsOf:        now.Add(-2 * time.Hour),
		NetWorthUSD:          100,
		BaseCurrency:         "USD",
		SnapshotType:         "upload",
		Status:               "active",
		CreatedAt:            now.Add(-2 * time.Hour),
	}).Error; err != nil {
		t.Fatalf("create old portfolio snapshot: %v", err)
	}
	if err := db.WithContext(ctx).Create(&PortfolioHolding{
		ID:                  "ph_refresh_old",
		PortfolioSnapshotID: oldSnapshotID,
		AssetType:           "crypto",
		Symbol:              "USDC",
		AssetKey:            "crypto:cg:usd-coin",
		Amount:              100,
		ValueUSD:            100,
		PricingSource:       "COINGECKO",
		ValuationStatus:     "priced",
		BalanceType:         "stablecoin",
	}).Error; err != nil {
		t.Fatalf("create old holding: %v", err)
	}

	server := &Server{
		db: &dbStore{db: db},
		market: &marketClient{
			cfg:    Config{},
			client: &http.Client{Transport: failRoundTripper{}},
			cache:  newMarketCacheStore(db),
			logger: log.New(io.Discard, "", 0),
		},
		logger: log.New(io.Discard, "", 0),
	}

	newSnapshotID, err := server.refreshActivePortfolio(ctx, userID)
	if err != nil {
		t.Fatalf("refreshActivePortfolio: %v", err)
	}
	if newSnapshotID == "" || newSnapshotID == oldSnapshotID {
		t.Fatalf("unexpected new snapshot id: %q", newSnapshotID)
	}

	var user User
	if err := db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.ActivePortfolioSnapshot == nil || *user.ActivePortfolioSnapshot != newSnapshotID {
		t.Fatalf("unexpected active snapshot pointer: %#v", user.ActivePortfolioSnapshot)
	}

	var oldSnapshot PortfolioSnapshot
	if err := db.WithContext(ctx).First(&oldSnapshot, "id = ?", oldSnapshotID).Error; err != nil {
		t.Fatalf("load old snapshot: %v", err)
	}
	if oldSnapshot.Status != "archived" {
		t.Fatalf("expected old snapshot archived, got %q", oldSnapshot.Status)
	}
	if oldSnapshot.ReplacedBySnapshotID == nil || *oldSnapshot.ReplacedBySnapshotID != newSnapshotID {
		t.Fatalf("expected replaced_by_snapshot_id=%q, got %#v", newSnapshotID, oldSnapshot.ReplacedBySnapshotID)
	}

	var newSnapshot PortfolioSnapshot
	if err := db.WithContext(ctx).First(&newSnapshot, "id = ?", newSnapshotID).Error; err != nil {
		t.Fatalf("load new snapshot: %v", err)
	}
	if newSnapshot.SnapshotType != "refresh" {
		t.Fatalf("expected snapshot_type refresh, got %q", newSnapshot.SnapshotType)
	}
	if newSnapshot.Status != "active" {
		t.Fatalf("expected new snapshot active, got %q", newSnapshot.Status)
	}
	if newSnapshot.SourceUploadBatchID != "upload_refresh_seed" {
		t.Fatalf("expected source upload batch to be preserved, got %q", newSnapshot.SourceUploadBatchID)
	}
	if newSnapshot.NetWorthUSD != 100 {
		t.Fatalf("expected net worth 100, got %v", newSnapshot.NetWorthUSD)
	}

	var holdings []PortfolioHolding
	if err := db.WithContext(ctx).Where("portfolio_snapshot_id = ?", newSnapshotID).Find(&holdings).Error; err != nil {
		t.Fatalf("load refreshed holdings: %v", err)
	}
	if len(holdings) != 1 {
		t.Fatalf("expected 1 refreshed holding, got %d", len(holdings))
	}
	if holdings[0].Symbol != "USDC" || holdings[0].BalanceType != "stablecoin" {
		t.Fatalf("unexpected refreshed holding: %#v", holdings[0])
	}
	if holdings[0].ValueUSD != 100 || holdings[0].ValuationStatus != "priced" {
		t.Fatalf("unexpected refreshed pricing: value=%v status=%q", holdings[0].ValueUSD, holdings[0].ValuationStatus)
	}

	var items []MarketDataSnapshotItem
	if err := db.WithContext(ctx).Where("market_data_snapshot_id = ?", newSnapshot.MarketDataSnapshotID).Find(&items).Error; err != nil {
		t.Fatalf("load refreshed market data items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 refreshed market data item, got %d", len(items))
	}
	if got := derefString(items[0].QuoteCurrency); got != "USDC" {
		t.Fatalf("expected quote currency USDC, got %q", got)
	}
	if items[0].PriceNative == nil || *items[0].PriceNative != 1 {
		t.Fatalf("expected native price 1, got %#v", items[0].PriceNative)
	}
	if items[0].FXRateToUSD == nil || *items[0].FXRateToUSD != 1 {
		t.Fatalf("expected fx rate to usd 1, got %#v", items[0].FXRateToUSD)
	}

	resp, err := server.buildPortfolioSnapshotResponse(ctx, userID, newSnapshotID)
	if err != nil {
		t.Fatalf("buildPortfolioSnapshotResponse: %v", err)
	}
	if resp.PortfolioSnapshotID != newSnapshotID {
		t.Fatalf("unexpected response snapshot id: %q", resp.PortfolioSnapshotID)
	}
	if len(resp.Holdings) != 1 || resp.Holdings[0].Symbol != "USDC" {
		t.Fatalf("unexpected response holdings: %#v", resp.Holdings)
	}
	if resp.Holdings[0].QuoteCurrency != "USDC" {
		t.Fatalf("expected response quote currency USDC, got %q", resp.Holdings[0].QuoteCurrency)
	}
	if resp.Holdings[0].CurrentPrice != 1 {
		t.Fatalf("expected response current price 1, got %v", resp.Holdings[0].CurrentPrice)
	}
	if resp.Holdings[0].ValueQuote != 100 {
		t.Fatalf("expected response quote value 100, got %v", resp.Holdings[0].ValueQuote)
	}
}

func TestBuildPortfolioSnapshotResponseUsesSnapshotValuationAsOfForMetrics(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := db.AutoMigrate(
		&User{},
		&UserProfile{},
		&MarketDataSnapshot{},
		&PortfolioSnapshot{},
		&PortfolioHolding{},
	); err != nil {
		t.Fatalf("auto migrate portfolio tables: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	valuationAsOf := dayStartUTCFromTime(now.AddDate(0, 0, -7))
	userID := "user_snapshot_anchor_test"
	snapshotID := "pf_snapshot_anchor"
	marketSnapshotID := "snap_snapshot_anchor"
	assetKey := "stock:mic:XNAS:TEST"
	exchangeMIC := "XNAS"

	t.Cleanup(func() {
		cleanupCandlesticks(ctx, db, assetKey)
		_ = db.WithContext(ctx).Where("portfolio_snapshot_id = ?", snapshotID).Delete(&PortfolioHolding{}).Error
		_ = db.WithContext(ctx).Where("id = ?", snapshotID).Delete(&PortfolioSnapshot{}).Error
		_ = db.WithContext(ctx).Where("id = ?", marketSnapshotID).Delete(&MarketDataSnapshot{}).Error
		_ = db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserProfile{}).Error
		_ = db.WithContext(ctx).Where("id = ?", userID).Delete(&User{}).Error
	})

	usdRates := oerLatestResponse{
		Base:      "USD",
		Timestamp: now.Unix(),
		Rates: map[string]float64{
			"USD": 1,
		},
	}
	if err := storeFXRates(ctx, db, usdRates, fxRateSourceOpenExchange); err != nil {
		t.Fatalf("store fx rates: %v", err)
	}

	if err := db.WithContext(ctx).Create(&User{
		ID:                      userID,
		ActivePortfolioSnapshot: &snapshotID,
		CreatedAt:               now,
		UpdatedAt:               now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.WithContext(ctx).Create(&UserProfile{
		UserID:       userID,
		Language:     "en",
		Timezone:     "UTC",
		BaseCurrency: "USD",
	}).Error; err != nil {
		t.Fatalf("create user profile: %v", err)
	}
	if err := db.WithContext(ctx).Create(&MarketDataSnapshot{
		ID:            marketSnapshotID,
		ValuationAsOf: valuationAsOf,
		BaseCurrency:  "USD",
		CreatedAt:     valuationAsOf,
	}).Error; err != nil {
		t.Fatalf("create market snapshot: %v", err)
	}
	if err := db.WithContext(ctx).Create(&PortfolioSnapshot{
		ID:                   snapshotID,
		UserID:               userID,
		SourceUploadBatchID:  "upload_anchor_seed",
		MarketDataSnapshotID: marketSnapshotID,
		ValuationAsOf:        valuationAsOf,
		NetWorthUSD:          1000,
		BaseCurrency:         "USD",
		SnapshotType:         "refresh",
		Status:               "active",
		CreatedAt:            valuationAsOf,
	}).Error; err != nil {
		t.Fatalf("create portfolio snapshot: %v", err)
	}
	if err := db.WithContext(ctx).Create(&PortfolioHolding{
		ID:                  "ph_snapshot_anchor",
		PortfolioSnapshotID: snapshotID,
		AssetType:           "stock",
		Symbol:              "TEST",
		AssetKey:            assetKey,
		ExchangeMIC:         &exchangeMIC,
		Amount:              10,
		ValueUSD:            1000,
		PricingSource:       "MARKETSTACK",
		ValuationStatus:     "priced",
	}).Error; err != nil {
		t.Fatalf("create holding: %v", err)
	}

	rows := make([]MarketCandlestick, 0, defaultSupportLookback+1)
	for offset := defaultSupportLookback; offset >= 0; offset-- {
		ts := valuationAsOf.AddDate(0, 0, -offset)
		closePrice := 100 + float64(defaultSupportLookback-offset)
		rows = append(rows, MarketCandlestick{
			Source:    candlestickSourceMarketstack,
			AssetKey:  assetKey,
			AssetType: "stock",
			Symbol:    "TEST",
			Interval:  candlestickIntervalDaily,
			Timestamp: ts,
			Open:      closePrice - 0.5,
			High:      closePrice + 1,
			Low:       closePrice - 1,
			Close:     closePrice,
			Volume:    1000,
			Currency:  "USD",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	if err := insertCandlesticks(ctx, db, rows); err != nil {
		t.Fatalf("insert candlesticks: %v", err)
	}

	server := &Server{
		db: &dbStore{db: db},
		market: &marketClient{
			cfg:    Config{},
			client: &http.Client{Transport: failRoundTripper{}},
			cache:  newMarketCacheStore(db),
			logger: log.New(io.Discard, "", 0),
		},
		logger: log.New(io.Discard, "", 0),
	}

	resp, err := server.buildPortfolioSnapshotResponse(ctx, userID, snapshotID)
	if err != nil {
		t.Fatalf("buildPortfolioSnapshotResponse: %v", err)
	}
	if resp.DashboardMetrics == nil {
		t.Fatal("expected dashboard metrics")
	}
	respAsOf, err := time.Parse(time.RFC3339, resp.DashboardMetrics.ValuationAsOf)
	if err != nil {
		t.Fatalf("parse dashboard valuation_as_of: %v", err)
	}
	if !respAsOf.Equal(valuationAsOf) {
		t.Fatalf("expected valuation_as_of %q, got %q", valuationAsOf.Format(time.RFC3339), resp.DashboardMetrics.ValuationAsOf)
	}
	if resp.DashboardMetrics.MetricsIncomplete {
		t.Fatal("expected metrics to be complete when snapshot-valued OHLCV is available")
	}
}
