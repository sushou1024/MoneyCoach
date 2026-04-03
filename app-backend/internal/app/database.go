package app

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openDatabase(ctx context.Context, cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := migrateDatabase(ctx, db); err != nil {
		return nil, err
	}
	return db.WithContext(ctx), nil
}

func migrateDatabase(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).AutoMigrate(
		&User{},
		&AuthIdentity{},
		&AuthSession{},
		&UserProfile{},
		&DeviceToken{},
		&UploadBatch{},
		&UploadImage{},
		&OCRAsset{},
		&OCRAmbiguity{},
		&AmbiguityResolution{},
		&UserAssetOverride{},
		&MarketDataSnapshot{},
		&MarketDataSnapshotItem{},
		&AssetCatalogCrypto{},
		&AssetCatalogStock{},
		&AssetCatalogFX{},
		&PortfolioSnapshot{},
		&PortfolioHolding{},
		&PortfolioTransaction{},
		&Calculation{},
		&ReportRisk{},
		&ReportStrategy{},
		&PlanState{},
		&Insight{},
		&InsightEvent{},
		&Entitlement{},
		&ExternalSubscription{},
		&Payment{},
		&WaitlistEntry{},
		&QuotaUsage{},
		&MarketDataCache{},
		&MarketCandlestick{},
		&FXDailyRate{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}
