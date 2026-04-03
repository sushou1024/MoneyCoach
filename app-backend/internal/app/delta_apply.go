package app

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Server) applyDeltaToActive(ctx context.Context, userID string, sourceBatchID string, deltas []transactionDelta) (string, []string, []string, error) {
	var user User
	if err := s.db.DB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return "", nil, nil, err
	}
	if user.ActivePortfolioSnapshot == nil {
		return "", nil, nil, errNotFound
	}

	snapshot, holdings, err := s.loadSnapshotWithHoldings(ctx, *user.ActivePortfolioSnapshot)
	if err != nil {
		return "", nil, nil, err
	}

	updated, warnings := applyDelta(holdings, deltas)
	overrides := s.loadUserAssetOverrides(ctx, userID)
	updated = refreshHoldingsPricing(ctx, s.market, updated, overrides)
	filteredHoldings, threshold, dropped := filterLowValueHoldings(updated, minPortfolioWeight)
	if dropped > 0 {
		s.logger.Printf("delta low-value filter user=%s dropped=%d threshold=%.2f", userID, dropped, threshold)
	}
	updated = filteredHoldings

	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return "", nil, nil, err
	}
	oerResp, _ := s.market.openExchangeLatest(ctx)
	baseCurrency, rateFromUSD := resolveDisplayCurrency(profile.BaseCurrency, oerResp.Rates)
	baseRateToUSD := rateToUSDFromRateFromUSD(rateFromUSD)

	valuationAsOf := priceAsOf()
	marketSnapshot := newMarketSnapshot(valuationAsOf)
	sourceID := snapshot.SourceUploadBatchID
	if sourceBatchID != "" {
		sourceID = sourceBatchID
	}
	newSnapshot := PortfolioSnapshot{
		ID:                   newID("pf"),
		UserID:               userID,
		SourceUploadBatchID:  sourceID,
		MarketDataSnapshotID: marketSnapshot.ID,
		ValuationAsOf:        valuationAsOf,
		NetWorthUSD:          computeNetWorth(updated),
		BaseCurrency:         baseCurrency,
		BaseFXRateToUSD:      &baseRateToUSD,
		SnapshotType:         "delta",
		Status:               "active",
		CreatedAt:            valuationAsOf,
	}

	transactionIDs := make([]string, 0, len(deltas))
	now := time.Now().UTC()
	planStateUpdates := s.buildPlanStateUpdates(ctx, userID, sourceBatchID, deltas, now)

	err = s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(&marketSnapshot).Error; err != nil {
			return err
		}
		if err := tx.Create(&newSnapshot).Error; err != nil {
			return err
		}
		for _, holding := range updated {
			row := PortfolioHolding{
				ID:                  newID("ph"),
				PortfolioSnapshotID: newSnapshot.ID,
				AssetType:           holding.AssetType,
				Symbol:              holding.Symbol,
				AssetKey:            holding.AssetKey,
				CoinGeckoID:         nullableString(holding.CoinGeckoID),
				ExchangeMIC:         nullableString(holding.ExchangeMIC),
				Amount:              holding.Amount,
				ValueFromScreenshot: holding.ValueFromScreenshot,
				ValueUSD:            holding.ValueUSD,
				PricingSource:       holding.PricingSource,
				ValuationStatus:     holding.ValuationStatus,
				CurrencyConverted:   holding.CurrencyConverted,
				CostBasisStatus:     holding.CostBasisStatus,
				BalanceType:         holding.BalanceType,
				AvgPrice:            holding.AvgPrice,
				AvgPriceSource:      nullableString(holding.AvgPriceSource),
				PNLPercent:          holding.PNLPercent,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			item := buildMarketDataSnapshotItem(marketSnapshot.ID, holding)
			if item == nil {
				continue
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		for _, delta := range deltas {
			if delta.AssetKey == "" || delta.Amount == 0 {
				continue
			}
			transactionID := newID("tx")
			fees := delta.FeesUSD
			if fees <= 0 && delta.FeesNative > 0 {
				currency := strings.ToUpper(strings.TrimSpace(delta.Currency))
				if currency == "USD" || stablecoinSet()[currency] {
					fees = delta.FeesNative
				}
			}
			var feesPtr *float64
			if fees > 0 {
				feesPtr = &fees
			}
			row := PortfolioTransaction{
				ID:               transactionID,
				UserID:           userID,
				SnapshotIDBefore: snapshot.ID,
				SnapshotIDAfter:  newSnapshot.ID,
				Symbol:           delta.Symbol,
				AssetType:        delta.AssetType,
				AssetKey:         nullableString(delta.AssetKey),
				Side:             sideFromAmount(delta.Amount),
				Amount:           abs(delta.Amount),
				Price:            delta.PriceUSD,
				Currency:         delta.Currency,
				ExecutedAt:       &now,
				Fees:             feesPtr,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			transactionIDs = append(transactionIDs, transactionID)
		}
		if err := tx.Model(&PortfolioSnapshot{}).Where("id = ?", snapshot.ID).Updates(map[string]any{"status": "archived", "replaced_by_snapshot_id": newSnapshot.ID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", userID).Update("active_portfolio_snapshot_id", newSnapshot.ID).Error; err != nil {
			return err
		}
		if len(planStateUpdates) > 0 {
			for _, update := range planStateUpdates {
				if err := tx.Model(&PlanState{}).Where("id = ?", update.ID).Updates(map[string]any{
					"state_json": marshalJSON(update.State),
					"updated_at": now,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", nil, warnings, err
	}
	return newSnapshot.ID, transactionIDs, warnings, nil
}

func (s *Server) refreshActivePortfolio(ctx context.Context, userID string) (string, error) {
	var user User
	if err := s.db.DB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return "", err
	}
	if user.ActivePortfolioSnapshot == nil {
		return "", errNotFound
	}

	snapshot, holdings, err := s.loadSnapshotWithHoldings(ctx, *user.ActivePortfolioSnapshot)
	if err != nil {
		return "", err
	}

	overrides := s.loadUserAssetOverrides(ctx, userID)
	refreshed := refreshHoldingsPricing(ctx, s.market, holdings, overrides)
	filteredHoldings, threshold, dropped := filterLowValueHoldings(refreshed, minPortfolioWeight)
	if dropped > 0 {
		s.logger.Printf("refresh low-value filter user=%s dropped=%d threshold=%.2f", userID, dropped, threshold)
	}
	refreshed = filteredHoldings

	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return "", err
	}
	oerResp, _ := s.market.openExchangeLatest(ctx)
	baseCurrency, rateFromUSD := resolveDisplayCurrency(profile.BaseCurrency, oerResp.Rates)
	baseRateToUSD := rateToUSDFromRateFromUSD(rateFromUSD)

	valuationAsOf := priceAsOf()
	marketSnapshot := newMarketSnapshot(valuationAsOf)
	newSnapshot := PortfolioSnapshot{
		ID:                   newID("pf"),
		UserID:               userID,
		SourceUploadBatchID:  snapshot.SourceUploadBatchID,
		MarketDataSnapshotID: marketSnapshot.ID,
		ValuationAsOf:        valuationAsOf,
		NetWorthUSD:          computeNetWorth(refreshed),
		BaseCurrency:         baseCurrency,
		BaseFXRateToUSD:      &baseRateToUSD,
		SnapshotType:         "refresh",
		Status:               "active",
		CreatedAt:            valuationAsOf,
	}

	err = s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(&marketSnapshot).Error; err != nil {
			return err
		}
		if err := tx.Create(&newSnapshot).Error; err != nil {
			return err
		}
		for _, holding := range refreshed {
			row := PortfolioHolding{
				ID:                  newID("ph"),
				PortfolioSnapshotID: newSnapshot.ID,
				AssetType:           holding.AssetType,
				Symbol:              holding.Symbol,
				AssetKey:            holding.AssetKey,
				CoinGeckoID:         nullableString(holding.CoinGeckoID),
				ExchangeMIC:         nullableString(holding.ExchangeMIC),
				Amount:              holding.Amount,
				ValueFromScreenshot: holding.ValueFromScreenshot,
				ValueUSD:            holding.ValueUSD,
				PricingSource:       holding.PricingSource,
				ValuationStatus:     holding.ValuationStatus,
				CurrencyConverted:   holding.CurrencyConverted,
				CostBasisStatus:     holding.CostBasisStatus,
				BalanceType:         holding.BalanceType,
				AvgPrice:            holding.AvgPrice,
				AvgPriceSource:      nullableString(holding.AvgPriceSource),
				PNLPercent:          holding.PNLPercent,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			item := buildMarketDataSnapshotItem(marketSnapshot.ID, holding)
			if item == nil {
				continue
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&PortfolioSnapshot{}).
			Where("id = ?", snapshot.ID).
			Updates(map[string]any{"status": "archived", "replaced_by_snapshot_id": newSnapshot.ID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", userID).Update("active_portfolio_snapshot_id", newSnapshot.ID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return newSnapshot.ID, nil
}

func refreshHoldingsPricing(ctx context.Context, market *marketClient, holdings []portfolioHolding, overrides map[string]UserAssetOverride) []portfolioHolding {
	for i := range holdings {
		hydrateHoldingIdentifiers(&holdings[i])
	}
	priceMap := fetchCoinGeckoPrices(ctx, market, holdings)
	stockPrices := fetchMarketstackPrices(ctx, market, holdings)
	oerRates := fetchOERRatesIfNeeded(ctx, market, holdings, stockPrices)
	for i := range holdings {
		applyPricing(&holdings[i], priceMap, stockPrices, oerRates)
		applyUserOverride(&holdings[i], overrides)
		applyCostBasis(&holdings[i])
	}
	return holdings
}

func sideFromAmount(amount float64) string {
	if amount >= 0 {
		return "buy"
	}
	return "sell"
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
