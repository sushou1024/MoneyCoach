package app

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *Server) applyTradeSlip(ctx context.Context, batch UploadBatch, image UploadImage, parsed tradeSlipOCRResponse) error {
	if len(parsed.Trades) == 0 {
		return s.markBatchFailed(ctx, batch.ID, "INVALID_IMAGE")
	}

	if tradeSlipHasUnsupportedKeywords(parsed.Trades) {
		return s.markBatchFailed(ctx, batch.ID, "INVALID_IMAGE")
	}

	oerResp, _ := s.market.openExchangeLatest(ctx)
	oerRates := oerResp.Rates
	if oerRates == nil {
		oerRates = map[string]float64{}
	}
	coinList, _ := s.market.coinGeckoList(ctx)
	resolutions := s.loadAmbiguityResolutions(ctx, batch.UserID)
	deltas := make([]transactionDelta, 0, len(parsed.Trades))
	warnings := make([]string, 0)

	type tradeInput struct {
		ID         string
		BaseSymbol string
		Currency   string
		Amount     float64
		Side       string
		Price      float64
		Fees       *tradeSlipFees
		RawSymbol  string
	}

	tradeInputs := make([]tradeInput, 0, len(parsed.Trades))
	assets := make([]ocrAssetInput, 0, len(parsed.Trades))
	for i, trade := range parsed.Trades {
		rawSymbol := strings.TrimSpace(trade.Symbol)
		amount := 0.0
		if trade.Amount != nil {
			amount = *trade.Amount
		}
		if rawSymbol == "" || amount == 0 {
			continue
		}
		currency := ""
		if trade.Currency != nil {
			currency = *trade.Currency
		}
		base, currency := tradePairFromSymbol(rawSymbol, currency)
		if base == "" {
			warnings = append(warnings, fmt.Sprintf("Skipped trade with empty symbol (%s).", rawSymbol))
			continue
		}
		price := 0.0
		if trade.Price != nil {
			price = *trade.Price
		}
		id := fmt.Sprintf("trade_%d", i)
		tradeInputs = append(tradeInputs, tradeInput{
			ID:         id,
			BaseSymbol: base,
			Currency:   currency,
			Amount:     amount,
			Side:       trade.Side,
			Price:      price,
			Fees:       trade.Fees,
			RawSymbol:  rawSymbol,
		})
		symbol := base
		assets = append(assets, ocrAssetInput{
			AssetID:   id,
			SymbolRaw: symbol,
			Symbol:    &symbol,
			AssetType: "",
			Amount:    abs(amount),
		})
	}

	resolvedAssets, _ := resolveAssets(ctx, s.market, batch.UserID, image.PlatformGuess, assets, coinList, resolutions, false)
	resolvedByID := make(map[string]portfolioHolding, len(resolvedAssets))
	for _, resolved := range resolvedAssets {
		if resolved.AssetID == "" {
			continue
		}
		resolvedByID[resolved.AssetID] = resolved.Holding
	}

	for _, input := range tradeInputs {
		asset, ok := resolvedByID[input.ID]
		if !ok || asset.AssetKey == "" {
			warnings = append(warnings, fmt.Sprintf("Unrecognized asset %s; skipped trade.", input.RawSymbol))
			continue
		}
		amount := input.Amount
		if strings.EqualFold(input.Side, "sell") {
			amount = -amount
		}
		currency := strings.ToUpper(strings.TrimSpace(input.Currency))
		priceNative := input.Price
		currencyOK := isTradeCurrencySupported(currency, oerRates)
		useMarket := priceNative <= 0 || !currencyOK
		priceUSD := priceNative
		if !useMarket && currency != "" && currency != "USD" && !stablecoinSet()[currency] {
			if rate, ok := oerRates[currency]; ok && rate > 0 {
				priceUSD = priceNative / rate
			} else {
				useMarket = true
			}
		}
		if useMarket {
			if asset.CurrentPrice <= 0 {
				warnings = append(warnings, fmt.Sprintf("Missing price for %s; skipped trade.", asset.Symbol))
				continue
			}
			priceUSD = asset.CurrentPrice
			priceNative = 0
		}
		delta := transactionDelta{
			Symbol:         asset.Symbol,
			AssetType:      asset.AssetType,
			AssetKey:       asset.AssetKey,
			Amount:         amount,
			PriceUSD:       priceUSD,
			PriceNative:    priceNative,
			Currency:       currency,
			AvgPriceSource: "",
			SkipCash:       useMarket,
		}
		if useMarket {
			delta.AvgPriceSource = "derived_from_market"
			if currency == "" {
				warnings = append(warnings, fmt.Sprintf("Missing currency for %s; used market price and skipped cash adjustment.", asset.Symbol))
			} else if !currencyOK {
				warnings = append(warnings, fmt.Sprintf("Unsupported currency %s for %s; used market price and skipped cash adjustment.", currency, asset.Symbol))
			} else {
				warnings = append(warnings, fmt.Sprintf("Missing price for %s; used market price and skipped cash adjustment.", asset.Symbol))
			}
		}

		if input.Fees != nil && input.Fees.Amount != nil && *input.Fees.Amount > 0 {
			feesCurrency := ""
			if input.Fees.Currency != nil {
				feesCurrency = *input.Fees.Currency
			}
			feesCurrency = strings.ToUpper(strings.TrimSpace(feesCurrency))
			if feesCurrency == "" {
				feesCurrency = currency
			}
			delta.FeesCurrency = feesCurrency
			if feesCurrency == currency && currencyOK {
				delta.FeesNative = *input.Fees.Amount
			}
			if feesCurrency == "USD" || stablecoinSet()[feesCurrency] {
				delta.FeesUSD = *input.Fees.Amount
			} else if rate, ok := oerRates[feesCurrency]; ok && rate > 0 {
				delta.FeesUSD = *input.Fees.Amount / rate
			}
		}

		deltas = append(deltas, delta)
	}

	if len(deltas) == 0 {
		return s.markBatchFailed(ctx, batch.ID, "INVALID_IMAGE")
	}

	_, _, deltaWarnings, err := s.applyDeltaToActive(ctx, batch.UserID, batch.ID, deltas)
	if err != nil {
		return err
	}

	warnings = append(warnings, deltaWarnings...)
	updates := map[string]any{
		"status":       "completed",
		"completed_at": time.Now().UTC(),
		"error_code":   nil,
	}
	if len(warnings) > 0 {
		updates["warnings"] = warnings
	}
	return s.db.DB().WithContext(ctx).Model(&UploadBatch{}).Where("id = ?", batch.ID).Updates(updates).Error
}

func tradeSlipHasUnsupportedKeywords(trades []tradeSlipTrade) bool {
	for _, trade := range trades {
		currency := ""
		if trade.Currency != nil {
			currency = *trade.Currency
		}
		fields := []string{trade.Symbol, currency}
		for _, field := range fields {
			if field == "" {
				continue
			}
			upper := strings.ToUpper(field)
			for _, keyword := range tradeSlipUnsupportedKeywords() {
				if strings.Contains(upper, keyword) {
					return true
				}
			}
		}
	}
	return false
}

func tradeSlipUnsupportedKeywords() []string {
	return []string{
		"FUTURES",
		"OPTION",
		"OPTIONS",
		"MARGIN",
		"PERP",
		"PERPETUAL",
		"LEVERAGE",
		"CROSS",
		"ISOLATED",
		"CONTRACT",
		"FUNDING",
		"POSITIONS",
		"合约",
		"永续",
		"资金费率",
		"保证金",
		"杠杆",
		"仓位",
		"逐仓",
		"全仓",
		"期权",
		"交割合约",
		"多单",
		"空单",
	}
}

func tradePairFromSymbol(symbolRaw string, currency string) (string, string) {
	base := strings.ToUpper(strings.TrimSpace(symbolRaw))
	quote := strings.ToUpper(strings.TrimSpace(currency))
	if base == "" {
		return "", quote
	}
	if quote != "" {
		return base, quote
	}
	if strings.Contains(base, "/") {
		parts := strings.SplitN(base, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	if strings.Contains(base, "-") {
		parts := strings.SplitN(base, "-", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	for _, suffix := range tradeSlipQuoteSuffixes() {
		if strings.HasSuffix(base, suffix) && len(base) > len(suffix) {
			return base[:len(base)-len(suffix)], suffix
		}
	}
	return base, ""
}

func tradeSlipQuoteSuffixes() []string {
	suffixes := []string{"USD"}
	for stable := range stablecoinSet() {
		suffixes = append(suffixes, stable)
	}
	return suffixes
}

func isTradeCurrencySupported(currency string, oerRates map[string]float64) bool {
	if currency == "" {
		return false
	}
	if currency == "USD" || stablecoinSet()[currency] {
		return true
	}
	rate, ok := oerRates[currency]
	return ok && rate > 0
}
