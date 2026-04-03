package app

import (
	"fmt"
	"net/http"
	"strings"
)

type assetLookupCandidate struct {
	Symbol      string `json:"symbol"`
	AssetType   string `json:"asset_type"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	AssetKey    string `json:"asset_key"`
	CoinGeckoID string `json:"coingecko_id,omitempty"`
	ExchangeMIC string `json:"exchange_mic,omitempty"`
}

type assetCommandRequest struct {
	Text string `json:"text"`
}

type assetCommandItem struct {
	ID      string
	Symbol  string
	Payload assetCommandPayload
}

func assetCommandAmount(target *assetCommandTarget) float64 {
	if target == nil || target.Amount == nil {
		return 0
	}
	return *target.Amount
}

func (s *Server) handleAssetsLookup(w http.ResponseWriter, r *http.Request) {
	symbolRaw := strings.TrimSpace(r.URL.Query().Get("symbol_raw"))
	assetType := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("asset_type")))
	if symbolRaw == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "symbol_raw required", nil)
		return
	}

	candidates := make([]assetLookupCandidate, 0)
	normalized := normalizeSymbol(symbolRaw)
	language := ""
	if userID := userIDFromContext(r.Context()); userID != "" {
		if profile, err := s.ensureUserProfile(r.Context(), userID); err == nil {
			language = profile.Language
		}
	}

	if assetType == "" || assetType == "crypto" {
		list, err := s.market.coinGeckoList(r.Context())
		if err == nil {
			for _, coin := range list {
				if strings.EqualFold(coin.Symbol, normalized) {
					candidates = append(candidates, assetLookupCandidate{
						Symbol:      strings.ToUpper(coin.Symbol),
						AssetType:   "crypto",
						Name:        coin.Name,
						Source:      "coingecko",
						AssetKey:    "crypto:cg:" + coin.ID,
						CoinGeckoID: coin.ID,
					})
				}
			}
		}
	}

	if assetType == "" || assetType == "stock" {
		if normalized != "" {
			if ticker, err := s.market.marketstackTicker(r.Context(), normalized); err == nil {
				name := strings.TrimSpace(ticker.Name)
				if localized, ok := hkNameForLanguage(ticker.Symbol, language); ok {
					name = localized
				}
				candidates = append(candidates, assetLookupCandidate{
					Symbol:      ticker.Symbol,
					AssetType:   "stock",
					Name:        name,
					Source:      "marketstack",
					AssetKey:    stockAssetKey(ticker.StockExchange.MIC, ticker.Symbol),
					ExchangeMIC: ticker.StockExchange.MIC,
				})
			}
		}
	}

	if assetType == "" || assetType == "forex" {
		if fx, err := s.market.openExchangeCurrencies(r.Context()); err == nil {
			if name, ok := fx[normalized]; ok {
				candidates = append(candidates, assetLookupCandidate{
					Symbol:    normalized,
					AssetType: "forex",
					Name:      name,
					Source:    "openexchangerates",
					AssetKey:  "forex:fx:" + normalized,
				})
			}
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"items": candidates})
}

func (s *Server) handleAssetsCommand(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	var req assetCommandRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "text required", nil)
		return
	}

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: s.prompts.AssetCommand}}},
		Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: req.Text}}}},
		GenerationConfig:  geminiGenerationConfig{Temperature: 0.0, MaxOutputTokens: geminiMaxOutputTokens, ResponseMimeType: "application/json"},
	}

	var parsed assetCommandResponse
	_, err := s.gemini.callGeminiJSON(r.Context(), request, &parsed)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "COMMAND_ERROR", "failed to parse command", nil)
		return
	}
	payloads := parsed.Payloads
	if parsed.Intent != "UPDATE_ASSET" || len(payloads) == 0 {
		s.writeJSON(w, http.StatusOK, map[string]any{"status": "ignored", "toast": "Only asset updates allowed here. See Insights for market news."})
		return
	}

	items := make([]assetCommandItem, 0, len(payloads))
	for i, payload := range payloads {
		if payload.TargetAsset == nil {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(payload.TargetAsset.Ticker))
		if symbol == "" {
			continue
		}
		items = append(items, assetCommandItem{
			ID:      fmt.Sprintf("cmd_%d", i),
			Symbol:  symbol,
			Payload: payload,
		})
	}
	if len(items) == 0 {
		s.writeJSON(w, http.StatusOK, map[string]any{"status": "ignored", "toast": "Please specify an amount."})
		return
	}

	coinList, _ := s.market.coinGeckoList(r.Context())
	assets := make([]ocrAssetInput, 0, len(items))
	for _, item := range items {
		amountForResolution := assetCommandAmount(item.Payload.TargetAsset)
		if amountForResolution == 0 {
			amountForResolution = 1
		}
		symbol := item.Symbol
		assets = append(assets, ocrAssetInput{
			AssetID:   item.ID,
			SymbolRaw: symbol,
			Symbol:    &symbol,
			AssetType: "",
			Amount:    amountForResolution,
		})
	}
	resolvedAssets, err := resolveAssets(r.Context(), s.market, userID, "", assets, coinList, nil, false)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "COMMAND_ERROR", "failed to resolve assets", nil)
		return
	}
	if len(resolvedAssets) == 0 {
		s.writeError(w, http.StatusBadRequest, "INVALID_ASSET", "unable to resolve asset", nil)
		return
	}

	resolvedByID := make(map[string]portfolioHolding, len(resolvedAssets))
	for _, resolved := range resolvedAssets {
		if resolved.AssetID == "" {
			continue
		}
		resolvedByID[resolved.AssetID] = resolved.Holding
	}

	holdings := aggregateHoldings(extractHoldings(resolvedAssets))
	priceMap := fetchCoinGeckoPrices(r.Context(), s.market, holdings)
	stockPrices := fetchMarketstackPrices(r.Context(), s.market, holdings)
	oerRates := fetchOERRatesIfNeeded(r.Context(), s.market, holdings, stockPrices)
	for _, item := range items {
		funding := item.Payload.FundingSource
		if funding == nil || !funding.IsExplicit || funding.Ticker == nil {
			continue
		}
		currency := strings.ToUpper(strings.TrimSpace(*funding.Ticker))
		if currency == "" || currency == "USD" || stablecoinSet()[currency] {
			continue
		}
		if _, ok := oerRates[currency]; ok {
			continue
		}
		if oerResp, err := s.market.openExchangeLatest(r.Context()); err == nil {
			oerRates = oerResp.Rates
		}
		break
	}
	priceByAssetKey := make(map[string]float64)
	priceBySymbol := make(map[string]float64)
	for i := range holdings {
		applyPricing(&holdings[i], priceMap, stockPrices, oerRates)
		applyCostBasis(&holdings[i])
		if holdings[i].ValuationStatus != "priced" || holdings[i].CurrentPrice <= 0 {
			continue
		}
		priceByAssetKey[holdings[i].AssetKey] = holdings[i].CurrentPrice
		symbol := strings.ToUpper(strings.TrimSpace(holdings[i].Symbol))
		if symbol != "" {
			if _, ok := priceBySymbol[symbol]; !ok {
				priceBySymbol[symbol] = holdings[i].CurrentPrice
			}
		}
	}

	warnings := make([]string, 0)
	fundingUnits := make(map[string]float64)
	resolveFundingUnit := func(currency string) float64 {
		if currency == "" {
			return 0
		}
		if currency == "USD" || stablecoinSet()[currency] {
			return 1
		}
		if cached, ok := fundingUnits[currency]; ok {
			return cached
		}
		if rate, ok := oerRates[currency]; ok && rate > 0 {
			unit := 1 / rate
			fundingUnits[currency] = unit
			return unit
		}
		if unit, ok := priceBySymbol[currency]; ok && unit > 0 {
			fundingUnits[currency] = unit
			return unit
		}
		fundingAsset := currency
		fundingAssets := []ocrAssetInput{{SymbolRaw: fundingAsset, Symbol: &fundingAsset, AssetType: "", Amount: 1}}
		fundingResolved, _, _ := resolveHoldings(r.Context(), s.market, userID, "", fundingAssets, coinList, nil)
		if len(fundingResolved) > 0 && fundingResolved[0].CurrentPrice > 0 {
			unit := fundingResolved[0].CurrentPrice
			fundingUnits[currency] = unit
			return unit
		}
		fundingUnits[currency] = 0
		return 0
	}

	deltas := make([]transactionDelta, 0, len(items))
	for _, item := range items {
		target := item.Payload.TargetAsset
		if target == nil {
			continue
		}
		resolved, ok := resolvedByID[item.ID]
		if !ok || resolved.AssetKey == "" {
			s.writeError(w, http.StatusBadRequest, "INVALID_ASSET", "unable to resolve asset", nil)
			return
		}
		priceUSD := 0.0
		priceNative := 0.0
		funding := item.Payload.FundingSource
		explicitFunding := funding != nil && funding.IsExplicit
		fundingCurrency := ""
		if funding != nil && funding.Ticker != nil {
			fundingCurrency = strings.ToUpper(strings.TrimSpace(*funding.Ticker))
		}
		if fundingCurrency == "" {
			explicitFunding = false
		}
		currencyUnitUSD := 0.0
		if explicitFunding {
			currencyUnitUSD = resolveFundingUnit(fundingCurrency)
			if currencyUnitUSD == 0 {
				warnings = append(warnings, fmt.Sprintf("Unsupported funding currency %s; skipped cash adjustment.", fundingCurrency))
				explicitFunding = false
				fundingCurrency = ""
			}
		}

		if item.Payload.PricePerUnit != nil {
			if explicitFunding {
				priceNative = *item.Payload.PricePerUnit
				priceUSD = priceNative * currencyUnitUSD
			} else {
				priceUSD = *item.Payload.PricePerUnit
				priceNative = priceUSD
			}
		} else if unit, ok := priceByAssetKey[resolved.AssetKey]; ok && unit > 0 {
			priceUSD = unit
			if explicitFunding && currencyUnitUSD > 0 {
				priceNative = priceUSD / currencyUnitUSD
			} else {
				priceNative = priceUSD
			}
		}
		if priceUSD <= 0 {
			s.writeError(w, http.StatusBadRequest, "INVALID_PRICE", "unable to determine price", nil)
			return
		}

		targetAmount := assetCommandAmount(target)
		fundingAmount := 0.0
		if explicitFunding && funding != nil && funding.Amount != nil {
			fundingAmount = *funding.Amount
		}
		if targetAmount == 0 && explicitFunding && fundingAmount > 0 && priceNative > 0 {
			targetAmount = fundingAmount / priceNative
		}
		if targetAmount == 0 {
			warnings = append(warnings, fmt.Sprintf("Missing amount for %s; skipped.", item.Symbol))
			continue
		}

		amount := targetAmount
		if strings.EqualFold(target.Action, "REMOVE") {
			amount = -targetAmount
		}
		currency := ""
		if explicitFunding {
			currency = fundingCurrency
		}

		deltas = append(deltas, transactionDelta{
			Symbol:      resolved.Symbol,
			AssetType:   resolved.AssetType,
			AssetKey:    resolved.AssetKey,
			Amount:      amount,
			PriceUSD:    priceUSD,
			PriceNative: priceNative,
			Currency:    currency,
		})
	}
	if len(deltas) == 0 {
		s.writeJSON(w, http.StatusOK, map[string]any{"status": "ignored", "toast": "Please specify an amount."})
		return
	}

	newSnapshotID, txIDs, deltaWarnings, err := s.applyDeltaToActive(r.Context(), userID, "", deltas)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "COMMAND_ERROR", "failed to apply update", nil)
		return
	}

	warnings = append(warnings, deltaWarnings...)
	toast := "✅ Asset updated."
	if len(warnings) > 0 {
		toast = "⚠️ Asset updated with warnings."
	}

	txID := ""
	if len(txIDs) > 0 {
		txID = txIDs[0]
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":                "completed",
		"transaction_id":        txID,
		"transaction_ids":       txIDs,
		"portfolio_snapshot_id": newSnapshotID,
		"parsed":                parsed,
		"toast":                 toast,
		"warnings":              warnings,
	})
}
