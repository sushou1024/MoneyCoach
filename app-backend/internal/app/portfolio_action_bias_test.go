package app

import (
	"context"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
)

func TestPortfolioHoldingActionBiasMatchesAssetBrief(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := db.AutoMigrate(
		&User{},
		&UserProfile{},
		&MarketDataSnapshot{},
		&MarketDataSnapshotItem{},
		&PortfolioSnapshot{},
		&PortfolioHolding{},
		&MarketCandlestick{},
		&FXDailyRate{},
	); err != nil {
		t.Fatalf("auto migrate action bias tables: %v", err)
	}

	asOf := dayStartUTCFromTime(time.Now().UTC())
	suffix := ulid.Make().String()
	userID := "user_action_bias_alignment_" + suffix
	snapshotID := "pf_action_bias_alignment_" + suffix
	marketSnapshotID := "snap_action_bias_alignment_" + suffix
	assetKey := "crypto:cg:binancecoin"
	coinGeckoID := "binancecoin"
	price := 241.0

	t.Cleanup(func() {
		_ = db.WithContext(ctx).Where("source = ? AND asset_key = ?", candlestickSourceCoinGecko, assetKey).Delete(&MarketCandlestick{}).Error
		_ = db.WithContext(ctx).Where("market_data_snapshot_id = ?", marketSnapshotID).Delete(&MarketDataSnapshotItem{}).Error
		_ = db.WithContext(ctx).Where("id = ?", marketSnapshotID).Delete(&MarketDataSnapshot{}).Error
		_ = db.WithContext(ctx).Where("portfolio_snapshot_id = ?", snapshotID).Delete(&PortfolioHolding{}).Error
		_ = db.WithContext(ctx).Where("id = ?", snapshotID).Delete(&PortfolioSnapshot{}).Error
		_ = db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserProfile{}).Error
		_ = db.WithContext(ctx).Where("id = ?", userID).Delete(&User{}).Error
	})

	_ = db.WithContext(ctx).Where("source = ? AND asset_key = ?", candlestickSourceCoinGecko, assetKey).Delete(&MarketCandlestick{}).Error

	usdRates := oerLatestResponse{
		Base:      "USD",
		Timestamp: asOf.Unix(),
		Rates: map[string]float64{
			"USD": 1,
		},
	}
	if err := storeFXRates(ctx, db, usdRates, fxRateSourceOpenExchange); err != nil {
		t.Fatalf("store fx rates: %v", err)
	}

	redisStore, err := newRedisStore(Config{RedisURL: "redis://localhost:6379/0"})
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	if err := redisStore.setJSON(ctx, "cache:coingecko:list", []coinGeckoCoinListEntry{{
		ID:     coinGeckoID,
		Symbol: "bnb",
		Name:   "BNB",
	}}, 24*time.Hour); err != nil {
		t.Fatalf("seed coingecko list cache: %v", err)
	}

	if err := db.WithContext(ctx).Create(&User{
		ID:                      userID,
		ActivePortfolioSnapshot: &snapshotID,
		CreatedAt:               asOf,
		UpdatedAt:               asOf,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.WithContext(ctx).Create(&UserProfile{
		UserID:       userID,
		Language:     "en",
		Timezone:     "UTC",
		BaseCurrency: "USD",
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := db.WithContext(ctx).Create(&MarketDataSnapshot{
		ID:            marketSnapshotID,
		ValuationAsOf: asOf,
		BaseCurrency:  "USD",
		CreatedAt:     asOf,
	}).Error; err != nil {
		t.Fatalf("create market snapshot: %v", err)
	}
	if err := db.WithContext(ctx).Create(&MarketDataSnapshotItem{
		ID:                   "snap_item_action_bias_alignment_" + suffix,
		MarketDataSnapshotID: marketSnapshotID,
		AssetType:            "crypto",
		Symbol:               "BNB",
		AssetKey:             assetKey,
		CoinGeckoID:          &coinGeckoID,
		PriceUSD:             price,
		PriceNative:          &price,
		QuoteCurrency:        strPtr("USD"),
		FXRateToUSD:          floatPtr(1),
		PriceSource:          "coingecko",
	}).Error; err != nil {
		t.Fatalf("create market snapshot item: %v", err)
	}
	if err := db.WithContext(ctx).Create(&PortfolioSnapshot{
		ID:                   snapshotID,
		UserID:               userID,
		MarketDataSnapshotID: marketSnapshotID,
		ValuationAsOf:        asOf,
		NetWorthUSD:          price * 10,
		BaseCurrency:         "USD",
		SnapshotType:         "upload",
		Status:               "active",
		CreatedAt:            asOf,
	}).Error; err != nil {
		t.Fatalf("create portfolio snapshot: %v", err)
	}
	if err := db.WithContext(ctx).Create(&PortfolioHolding{
		ID:                  "ph_action_bias_alignment_" + suffix,
		PortfolioSnapshotID: snapshotID,
		AssetType:           "crypto",
		Symbol:              "BNB",
		AssetKey:            assetKey,
		CoinGeckoID:         &coinGeckoID,
		Amount:              10,
		ValueUSD:            price * 10,
		PricingSource:       "COINGECKO",
		ValuationStatus:     "priced",
		BalanceType:         "coin",
	}).Error; err != nil {
		t.Fatalf("create holding: %v", err)
	}

	candles := make([]MarketCandlestick, 0, intelligenceLookbackDays+1)
	for i := 0; i <= intelligenceLookbackDays; i++ {
		ts := asOf.AddDate(0, 0, -(intelligenceLookbackDays - i))
		closePrice := 500.0 - float64(i)
		candles = append(candles, MarketCandlestick{
			Source:    candlestickSourceCoinGecko,
			AssetKey:  assetKey,
			AssetType: "crypto",
			Symbol:    "BNB",
			Interval:  candlestickIntervalDaily,
			Timestamp: ts,
			Open:      closePrice + 2,
			High:      closePrice + 4,
			Low:       closePrice - 4,
			Close:     closePrice,
			Volume:    1000 + float64(i),
			Currency:  "USD",
			CreatedAt: asOf,
			UpdatedAt: asOf,
		})
	}
	if err := db.WithContext(ctx).Create(&candles).Error; err != nil {
		t.Fatalf("create candlesticks: %v", err)
	}

	server := &Server{
		db: &dbStore{db: db},
		market: &marketClient{
			cfg:    Config{RedisURL: "redis://localhost:6379/0"},
			client: &http.Client{Transport: failRoundTripper{}},
			redis:  redisStore,
			cache:  newMarketCacheStore(db),
			logger: log.New(io.Discard, "", 0),
		},
		logger: log.New(io.Discard, "", 0),
	}

	resp, err := server.buildPortfolioSnapshotResponse(ctx, userID, snapshotID)
	if err != nil {
		t.Fatalf("buildPortfolioSnapshotResponse: %v", err)
	}
	if len(resp.Holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(resp.Holdings))
	}
	if resp.Holdings[0].ActionBias == nil {
		t.Fatalf("expected action_bias on holding response, got nil")
	}

	brief, err := server.buildAssetBrief(ctx, userID, assetKey)
	if err != nil {
		t.Fatalf("buildAssetBrief: %v", err)
	}
	if got, want := *resp.Holdings[0].ActionBias, brief.ActionBias; got != want {
		t.Fatalf("expected assets action_bias %q to match asset brief %q", got, want)
	}
	if brief.ActionBias != "reduce" {
		t.Fatalf("expected downtrend asset to resolve to reduce, got %q", brief.ActionBias)
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func strPtr(value string) *string {
	return &value
}
