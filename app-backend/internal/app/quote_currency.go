package app

import (
	"context"
	"strings"
)

type assetQuoteMetadata struct {
	QuoteCurrency     string
	CurrentPriceQuote float64
	FXRateToUSD       float64
}

func applyQuoteMetadata(holding *portfolioHolding, meta assetQuoteMetadata) {
	if holding == nil {
		return
	}
	currency := normalizeCurrency(meta.QuoteCurrency)
	if currency == "" {
		currency = "USD"
	}
	fxRateToUSD := meta.FXRateToUSD
	if fxRateToUSD <= 0 {
		if currency == "USD" || stablecoinSet()[currency] {
			fxRateToUSD = 1
		}
	}
	priceQuote := meta.CurrentPriceQuote
	if priceQuote <= 0 && holding.CurrentPrice > 0 && fxRateToUSD > 0 {
		priceQuote = holding.CurrentPrice / fxRateToUSD
	}
	holding.QuoteCurrency = currency
	holding.FXRateToUSD = fxRateToUSD
	holding.CurrentPriceQuote = priceQuote
}

func quoteCurrencyForHolding(holding portfolioHolding) string {
	currency := normalizeCurrency(holding.QuoteCurrency)
	if currency != "" {
		return currency
	}
	switch {
	case holding.AssetType == "forex" && holding.Symbol != "":
		return normalizeCurrency(holding.Symbol)
	case holding.AssetType == "crypto" && holding.BalanceType == "stablecoin" && holding.Symbol != "":
		return normalizeCurrency(holding.Symbol)
	default:
		return "USD"
	}
}

func quoteFXRateToUSD(holding portfolioHolding) float64 {
	if holding.FXRateToUSD > 0 {
		return holding.FXRateToUSD
	}
	currency := quoteCurrencyForHolding(holding)
	if currency == "USD" || stablecoinSet()[currency] {
		return 1
	}
	return 0
}

func quotePriceForHolding(holding portfolioHolding) float64 {
	if holding.CurrentPriceQuote > 0 {
		return holding.CurrentPriceQuote
	}
	rateToUSD := quoteFXRateToUSD(holding)
	if rateToUSD > 0 && holding.CurrentPrice > 0 {
		return holding.CurrentPrice / rateToUSD
	}
	if holding.AssetType == "forex" && quoteCurrencyForHolding(holding) != "" {
		return 1
	}
	return holding.CurrentPrice
}

func quoteValueForHolding(holding portfolioHolding) float64 {
	if holding.Amount > 0 {
		if price := quotePriceForHolding(holding); price > 0 {
			return holding.Amount * price
		}
	}
	rateToUSD := quoteFXRateToUSD(holding)
	if rateToUSD > 0 && holding.ValueUSD > 0 {
		return holding.ValueUSD / rateToUSD
	}
	return holding.ValueUSD
}

func quoteAvgPriceForHolding(holding portfolioHolding) *float64 {
	if holding.AvgPrice == nil {
		return nil
	}
	rateToUSD := quoteFXRateToUSD(holding)
	if rateToUSD <= 0 {
		value := *holding.AvgPrice
		return &value
	}
	value := *holding.AvgPrice / rateToUSD
	return &value
}

func holdingQuoteMetadataFromSnapshotItem(item MarketDataSnapshotItem) assetQuoteMetadata {
	meta := assetQuoteMetadata{
		QuoteCurrency: normalizeCurrency(derefString(item.QuoteCurrency)),
	}
	if item.PriceNative != nil {
		meta.CurrentPriceQuote = *item.PriceNative
	}
	if item.FXRateToUSD != nil {
		meta.FXRateToUSD = *item.FXRateToUSD
	}
	if meta.QuoteCurrency == "" {
		meta.QuoteCurrency = "USD"
	}
	if meta.FXRateToUSD <= 0 && (meta.QuoteCurrency == "USD" || stablecoinSet()[meta.QuoteCurrency]) {
		meta.FXRateToUSD = 1
	}
	return meta
}

func assetQuoteMetadataForHolding(holding portfolioHolding) assetQuoteMetadata {
	meta := assetQuoteMetadata{
		QuoteCurrency: normalizeCurrency(holding.QuoteCurrency),
		FXRateToUSD:   holding.FXRateToUSD,
	}
	if meta.QuoteCurrency == "" {
		meta.QuoteCurrency = quoteCurrencyForHolding(holding)
	}
	if meta.FXRateToUSD <= 0 {
		meta.FXRateToUSD = quoteFXRateToUSD(holding)
	}
	meta.CurrentPriceQuote = quotePriceForHolding(holding)
	return meta
}

func (s *Server) enrichHoldingsWithQuoteMetadata(ctx context.Context, snapshot PortfolioSnapshot, holdings []portfolioHolding) []portfolioHolding {
	if len(holdings) == 0 {
		return holdings
	}

	itemsByAssetKey := s.loadSnapshotQuoteItems(ctx, snapshot.MarketDataSnapshotID)
	stockCurrencies := s.loadStockQuoteCurrencies(ctx, holdings)
	var (
		oerRates map[string]float64
		oerReady bool
	)
	loadOERRates := func() map[string]float64 {
		if oerReady {
			return oerRates
		}
		oerReady = true
		if s == nil || s.market == nil {
			oerRates = map[string]float64{}
			return oerRates
		}
		resp, err := s.market.openExchangeLatest(ctx)
		if err != nil || resp.Rates == nil {
			oerRates = map[string]float64{}
			return oerRates
		}
		oerRates = resp.Rates
		return oerRates
	}

	for i := range holdings {
		if item, ok := itemsByAssetKey[holdings[i].AssetKey]; ok {
			applyQuoteMetadata(&holdings[i], holdingQuoteMetadataFromSnapshotItem(item))
		}
		if holdings[i].QuoteCurrency == "" || holdings[i].CurrentPriceQuote <= 0 || quoteFXRateToUSD(holdings[i]) <= 0 {
			s.hydrateHoldingQuoteFallback(ctx, &holdings[i], stockCurrencies, loadOERRates)
		}
	}
	return holdings
}

func (s *Server) loadSnapshotQuoteItems(ctx context.Context, marketSnapshotID string) map[string]MarketDataSnapshotItem {
	if marketSnapshotID == "" {
		return nil
	}
	var rows []MarketDataSnapshotItem
	if err := s.db.DB().WithContext(ctx).
		Where("market_data_snapshot_id = ?", marketSnapshotID).
		Find(&rows).Error; err != nil {
		return nil
	}
	items := make(map[string]MarketDataSnapshotItem, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.AssetKey) == "" {
			continue
		}
		items[row.AssetKey] = row
	}
	return items
}

func (s *Server) loadStockQuoteCurrencies(ctx context.Context, holdings []portfolioHolding) map[string]string {
	byAssetKey := make(map[string]string)
	for _, holding := range holdings {
		if holding.AssetType != "stock" || holding.Symbol == "" {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(holding.Symbol))
		exchangeMIC := strings.ToUpper(strings.TrimSpace(holding.ExchangeMIC))
		var row AssetCatalogStock
		query := s.db.DB().WithContext(ctx).Model(&AssetCatalogStock{}).Where("symbol = ?", symbol)
		if exchangeMIC != "" && exchangeMIC != "UNKNOWN" {
			query = query.Where("exchange_mic = ?", exchangeMIC)
		}
		if err := query.First(&row).Error; err == nil {
			currency := normalizeCurrency(row.Currency)
			if currency != "" {
				byAssetKey[holding.AssetKey] = currency
				continue
			}
		}
		if exchangeMIC != "" {
			byAssetKey[holding.AssetKey] = marketstackPriceCurrency("", symbol, exchangeMIC)
		}
	}
	return byAssetKey
}

func (s *Server) hydrateHoldingQuoteFallback(_ context.Context, holding *portfolioHolding, stockCurrencies map[string]string, loadOERRates func() map[string]float64) {
	if holding == nil {
		return
	}

	switch {
	case holding.AssetType == "crypto" && holding.BalanceType == "stablecoin":
		currency := normalizeCurrency(holding.Symbol)
		if currency == "" {
			currency = "USD"
		}
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     currency,
			CurrentPriceQuote: 1,
			FXRateToUSD:       1,
		})
		return
	case holding.AssetType == "crypto":
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     "USD",
			CurrentPriceQuote: holding.CurrentPrice,
			FXRateToUSD:       1,
		})
		return
	case holding.AssetType == "forex":
		currency := normalizeCurrency(holding.Symbol)
		if currency == "" {
			currency = "USD"
		}
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     currency,
			CurrentPriceQuote: 1,
			FXRateToUSD:       holding.CurrentPrice,
		})
		return
	case holding.AssetType == "stock":
		currency := stockCurrencies[holding.AssetKey]
		if currency == "" {
			currency = marketstackPriceCurrency("", holding.Symbol, holding.ExchangeMIC)
		}
		if currency == "" {
			currency = "USD"
		}
		fxRateToUSD := 1.0
		if currency != "USD" {
			rates := loadOERRates()
			if rate, ok := currencyRateToUSD(currency, rates); ok && rate > 0 {
				fxRateToUSD = rate
			} else {
				fxRateToUSD = 0
			}
		}
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency: currency,
			FXRateToUSD:   fxRateToUSD,
		})
		return
	default:
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     "USD",
			CurrentPriceQuote: holding.CurrentPrice,
			FXRateToUSD:       1,
		})
	}
}

func buildMarketDataSnapshotItem(marketSnapshotID string, holding portfolioHolding) *MarketDataSnapshotItem {
	if holding.ValuationStatus != "priced" && holding.ValuationStatus != "user_provided" {
		return nil
	}
	priceUSD := holding.CurrentPrice
	if priceUSD <= 0 && holding.Amount > 0 {
		priceUSD = holding.ValueUSD / holding.Amount
	}
	if priceUSD <= 0 && holding.ValueUSD <= 0 {
		return nil
	}
	item := &MarketDataSnapshotItem{
		ID:                   newID("snap_item"),
		MarketDataSnapshotID: marketSnapshotID,
		AssetType:            holding.AssetType,
		Symbol:               holding.Symbol,
		AssetKey:             holding.AssetKey,
		CoinGeckoID:          nullableString(holding.CoinGeckoID),
		ExchangeMIC:          nullableString(holding.ExchangeMIC),
		PriceUSD:             priceUSD,
		PriceSource:          holding.PricingSource,
	}
	meta := assetQuoteMetadataForHolding(holding)
	if meta.CurrentPriceQuote > 0 {
		item.PriceNative = &meta.CurrentPriceQuote
	}
	if meta.QuoteCurrency != "" {
		currency := meta.QuoteCurrency
		item.QuoteCurrency = &currency
	}
	if meta.FXRateToUSD > 0 {
		rate := meta.FXRateToUSD
		item.FXRateToUSD = &rate
	}
	return item
}

func exchangeMICFromAssetKey(assetKey string) string {
	parts := strings.Split(strings.TrimSpace(assetKey), ":")
	if len(parts) < 4 || parts[0] != "stock" || parts[1] != "mic" {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(parts[2]))
}

func (s *Server) resolveStockQuoteCurrency(ctx context.Context, assetKey, symbol string) string {
	assetKey = strings.TrimSpace(assetKey)
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if assetKey != "" {
		var candle MarketCandlestick
		if err := s.db.DB().WithContext(ctx).
			Where("asset_key = ? AND asset_type = ?", assetKey, "stock").
			Order("timestamp desc").
			First(&candle).Error; err == nil {
			if currency := normalizeCurrency(candle.Currency); currency != "" {
				return currency
			}
		}
	}

	exchangeMIC := exchangeMICFromAssetKey(assetKey)
	if symbol != "" {
		var row AssetCatalogStock
		query := s.db.DB().WithContext(ctx).Model(&AssetCatalogStock{}).Where("symbol = ?", symbol)
		if exchangeMIC != "" && exchangeMIC != "UNKNOWN" {
			query = query.Where("exchange_mic = ?", exchangeMIC)
		}
		if err := query.First(&row).Error; err == nil {
			if currency := normalizeCurrency(row.Currency); currency != "" {
				return currency
			}
		}
	}

	currency := marketstackPriceCurrency("", symbol, exchangeMIC)
	if currency != "" {
		return currency
	}
	return "USD"
}
