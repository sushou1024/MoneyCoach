package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Server) processOCRHoldings(ctx context.Context, batchID string) error {
	var batch UploadBatch
	if err := s.db.DB().WithContext(ctx).First(&batch, "id = ?", batchID).Error; err != nil {
		return err
	}
	if batch.Status != "processing" && batch.Status != "pending_upload" {
		return nil
	}

	var images []UploadImage
	if err := s.db.DB().WithContext(ctx).Where("upload_batch_id = ?", batch.ID).Find(&images).Error; err != nil {
		return err
	}
	if len(images) == 0 {
		return fmt.Errorf("no images for batch %s", batch.ID)
	}

	s.logger.Printf("ocr holdings start batch=%s images=%d", batch.ID, len(images))

	parts := []geminiPart{{Text: buildOCRInput(images)}}
	phashByID := make(map[string]string)
	for _, image := range images {
		bytes, err := s.storage.getObjectBytes(ctx, image.StorageKey)
		if err != nil {
			return err
		}
		if hash, err := computePHash(bytes); err == nil && hash != "" {
			phashByID[image.ID] = hash
		}
		encoded := base64.StdEncoding.EncodeToString(bytes)
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{MimeType: imageMimeType(image.StorageKey), Data: encoded},
		})
	}

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: s.prompts.OCRPortfolio}}},
		Contents:          []geminiContent{{Role: "user", Parts: parts}},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.0,
			MaxOutputTokens:  geminiMaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	promptHash := hashPrompt(s.prompts.OCRPortfolio)

	var parsed ocrPromptResponse
	callStart := time.Now()
	raw, err := s.gemini.callGeminiJSON(ctx, request, &parsed)
	if err != nil {
		s.logger.Printf("ocr holdings gemini error batch=%s duration=%s err=%v", batch.ID, time.Since(callStart), err)
		return s.markOCRFailure(ctx, &batch, images, "EXTRACTION_FAILURE", err.Error())
	}
	s.logger.Printf("ocr holdings gemini done batch=%s duration=%s", batch.ID, time.Since(callStart))

	imageByID := make(map[string]ocrImage)
	for _, image := range parsed.Images {
		imageByID[image.ImageID] = image
	}
	warningsByID := buildImageWarnings(images, imageByID, phashByID)

	if err := s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Where("upload_image_id IN (?)", imageIDs(images)).Delete(&OCRAsset{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&UploadBatch{}).Where("id = ?", batch.ID).Updates(map[string]any{
			"ocr_prompt_hash":      promptHash,
			"ocr_model_output_raw": raw,
			"ocr_parse_error":      nil,
		}).Error; err != nil {
			return err
		}

		for _, image := range images {
			imageResult, ok := imageByID[image.ID]
			if !ok {
				reason := "PARSE_ERROR"
				imageResult = ocrImage{
					Status:        "ignored_invalid",
					ErrorReason:   &reason,
					PlatformGuess: image.PlatformGuess,
					Assets:        []ocrAsset{},
				}
			}
			status := normalizeOCRStatus(imageResult.Status)
			if status == "" {
				status = "ignored_invalid"
			}
			update := map[string]any{
				"status":         status,
				"error_reason":   imageResult.ErrorReason,
				"platform_guess": strings.TrimSpace(imageResult.PlatformGuess),
			}
			fingerprint := computeFingerprintV0(imageResult.PlatformGuess, imageResult.Assets)
			if fingerprint != "" {
				update["fingerprint_v0"] = fingerprint
			}
			if hash := phashByID[image.ID]; hash != "" {
				update["phash"] = hash
			}
			if warnings, ok := warningsByID[image.ID]; ok {
				update["warnings"] = warnings
			} else {
				update["warnings"] = []string{}
			}
			if err := tx.Model(&UploadImage{}).Where("id = ?", image.ID).Updates(update).Error; err != nil {
				return err
			}

			if status != "success" {
				continue
			}
			for _, asset := range imageResult.Assets {
				confidence := computeConfidence(asset)
				var avgPriceSource *string
				if asset.AvgPrice != nil {
					source := "provided"
					avgPriceSource = &source
				}
				ocrAsset := OCRAsset{
					ID:                  newID("ocr"),
					UploadImageID:       image.ID,
					SymbolRaw:           strings.TrimSpace(asset.SymbolRaw),
					Symbol:              asset.Symbol,
					AssetType:           normalizeAssetType(asset.AssetType),
					Amount:              asset.Amount,
					ValueFromScreenshot: asset.ValueFromScreenshot,
					DisplayCurrency:     asset.DisplayCurrency,
					Confidence:          confidence,
					AvgPrice:            asset.AvgPrice,
					AvgPriceSource:      avgPriceSource,
					PNLPercent:          asset.PNLPercent,
				}
				if err := tx.Create(&ocrAsset).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	successImages, _, _ := countOCRStatuses(parsed.Images)
	if successImages == 0 {
		errorCode := mapBatchError(parsed.Images)
		s.logger.Printf("ocr holdings done batch=%s status=failed error_code=%s", batch.ID, errorCode)
		return s.markBatchFailed(ctx, batch.ID, errorCode)
	}

	s.logger.Printf("ocr holdings done batch=%s status=needs_review success_images=%d", batch.ID, successImages)
	return s.updateBatchStatus(ctx, batch.ID, "needs_review", "")
}

func (s *Server) processOCRTradeSlip(ctx context.Context, batchID string) error {
	var batch UploadBatch
	if err := s.db.DB().WithContext(ctx).First(&batch, "id = ?", batchID).Error; err != nil {
		return err
	}
	if batch.Status != "processing" && batch.Status != "pending_upload" {
		return nil
	}

	var images []UploadImage
	if err := s.db.DB().WithContext(ctx).Where("upload_batch_id = ?", batch.ID).Find(&images).Error; err != nil {
		return err
	}
	if len(images) == 0 {
		return fmt.Errorf("no images for batch %s", batch.ID)
	}
	s.logger.Printf("ocr trade slip start batch=%s", batch.ID)
	image := images[0]
	bytes, err := s.storage.getObjectBytes(ctx, image.StorageKey)
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: s.prompts.TradeSlipOCR}}},
		Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: fmt.Sprintf("Input trade slip image:\n- image_id: %s\nReturn strict JSON per the system prompt.", image.ID)}, {InlineData: &geminiInlineData{MimeType: imageMimeType(image.StorageKey), Data: encoded}}}}},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.0,
			MaxOutputTokens:  geminiMaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	promptHash := hashPrompt(s.prompts.TradeSlipOCR)

	var parsed tradeSlipOCRResponse
	callStart := time.Now()
	raw, err := s.gemini.callGeminiJSON(ctx, request, &parsed)
	if err != nil {
		s.logger.Printf("ocr trade slip gemini error batch=%s duration=%s err=%v", batch.ID, time.Since(callStart), err)
		return s.markBatchFailed(ctx, batch.ID, "EXTRACTION_FAILURE")
	}
	s.logger.Printf("ocr trade slip gemini done batch=%s duration=%s trades=%d", batch.ID, time.Since(callStart), len(parsed.Trades))

	_ = s.db.DB().WithContext(ctx).Model(&UploadBatch{}).Where("id = ?", batch.ID).Updates(map[string]any{
		"ocr_prompt_hash":      promptHash,
		"ocr_model_output_raw": raw,
		"ocr_parse_error":      nil,
	}).Error

	if len(parsed.Trades) == 0 {
		return s.markBatchFailed(ctx, batch.ID, "INVALID_IMAGE")
	}
	return s.applyTradeSlip(ctx, batch, image, parsed)
}

func (s *Server) processNormalization(ctx context.Context, batchID string) error {
	var batch UploadBatch
	if err := s.db.DB().WithContext(ctx).First(&batch, "id = ?", batchID).Error; err != nil {
		return err
	}
	if batch.Status != "processing" {
		return nil
	}

	start := time.Now()
	logStep := func(step string, started time.Time) {
		s.logger.Printf("normalize step=%s batch=%s duration=%s", step, batch.ID, time.Since(started))
	}

	var images []UploadImage
	stepStart := time.Now()
	if err := s.db.DB().WithContext(ctx).Where("upload_batch_id = ?", batch.ID).Find(&images).Error; err != nil {
		return err
	}
	if len(images) == 0 {
		return fmt.Errorf("no images for batch %s", batch.ID)
	}

	var assets []OCRAsset
	if err := s.db.DB().WithContext(ctx).Where("upload_image_id IN (?)", imageIDs(images)).Find(&assets).Error; err != nil {
		return err
	}
	logStep("load_inputs", stepStart)
	s.logger.Printf("normalize start batch=%s images=%d assets=%d", batch.ID, len(images), len(assets))

	stepStart = time.Now()
	inputAssets := make([]ocrAssetInput, 0, len(assets))
	imagePlatform := make(map[string]string, len(images))
	for _, image := range images {
		imagePlatform[image.ID] = image.PlatformGuess
	}
	for _, asset := range assets {
		inputAssets = append(inputAssets, ocrAssetInput{
			AssetID:             asset.ID,
			ImageID:             asset.UploadImageID,
			PlatformGuess:       imagePlatform[asset.UploadImageID],
			SymbolRaw:           asset.SymbolRaw,
			Symbol:              asset.Symbol,
			AssetType:           asset.AssetType,
			Amount:              asset.Amount,
			ValueFromScreenshot: asset.ValueFromScreenshot,
			ManualValueUSD:      asset.ManualValueUSD,
			DisplayCurrency:     asset.DisplayCurrency,
			PNLPercent:          asset.PNLPercent,
			AvgPrice:            asset.AvgPrice,
			AvgPriceSource:      derefString(asset.AvgPriceSource),
		})
	}
	logStep("build_inputs", stepStart)

	stepStart = time.Now()
	coinList, err := s.market.coinGeckoList(ctx)
	if err != nil {
		return err
	}
	logStep("coingecko_list", stepStart)

	platformGuess := images[0].PlatformGuess
	resolutions := s.loadAmbiguityResolutions(ctx, batch.UserID)
	stepStart = time.Now()
	resolvedAssets, err := resolveAssets(ctx, s.market, batch.UserID, platformGuess, inputAssets, coinList, resolutions, true)
	if err != nil {
		return err
	}
	logStep("resolve_assets", stepStart)

	stepStart = time.Now()
	assetsByImage := make(map[string][]resolvedAsset)
	for _, asset := range resolvedAssets {
		if asset.ImageID == "" {
			continue
		}
		assetsByImage[asset.ImageID] = append(assetsByImage[asset.ImageID], asset)
	}
	fingerprintByImage := make(map[string]string)
	for _, image := range images {
		fingerprintByImage[image.ID] = computeFingerprintV1(image.PlatformGuess, assetsByImage[image.ID])
	}

	duplicateOf := make(map[string]string)
	duplicateSet := make(map[string]bool)
	sortedImages := append([]UploadImage(nil), images...)
	sort.Slice(sortedImages, func(i, j int) bool {
		if sortedImages[i].CreatedAt.Equal(sortedImages[j].CreatedAt) {
			return sortedImages[i].ID < sortedImages[j].ID
		}
		return sortedImages[i].CreatedAt.Before(sortedImages[j].CreatedAt)
	})
	seenFingerprint := make(map[string]string)
	for _, image := range sortedImages {
		fingerprint := fingerprintByImage[image.ID]
		if fingerprint == "" {
			continue
		}
		if firstID, ok := seenFingerprint[fingerprint]; ok {
			if image.DuplicateOverride {
				continue
			}
			duplicateSet[image.ID] = true
			duplicateOf[image.ID] = firstID
			continue
		}
		seenFingerprint[fingerprint] = image.ID
	}
	logStep("dedupe", stepStart)

	stepStart = time.Now()
	effectiveAssets := make([]resolvedAsset, 0, len(resolvedAssets))
	for _, asset := range resolvedAssets {
		if duplicateSet[asset.ImageID] {
			continue
		}
		effectiveAssets = append(effectiveAssets, asset)
	}

	holdings := aggregateHoldings(extractHoldings(effectiveAssets))
	priceMap := fetchCoinGeckoPrices(ctx, s.market, holdings)
	stockPrices := fetchMarketstackPrices(ctx, s.market, holdings)
	oerRates := fetchOERRatesIfNeeded(ctx, s.market, holdings, stockPrices)
	overrides := s.loadUserAssetOverrides(ctx, batch.UserID)
	for i := range holdings {
		applyPricing(&holdings[i], priceMap, stockPrices, oerRates)
		applyUserOverride(&holdings[i], overrides)
		applyCostBasis(&holdings[i])
	}
	logStep("pricing", stepStart)

	stepStart = time.Now()
	filteredHoldings, threshold, dropped := filterLowValueHoldings(holdings, minPortfolioWeight)
	if dropped > 0 {
		s.logger.Printf("normalize low-value filter batch=%s dropped=%d threshold=%.2f", batch.ID, dropped, threshold)
	}
	holdings = filteredHoldings
	logStep("low_value_filter", stepStart)

	valuationAsOf := priceAsOf()
	stepStart = time.Now()
	oerResp, _ := s.market.openExchangeLatest(ctx)
	seriesByAssetKey := fetchPriceSeriesAsOf(ctx, s.market, holdings, valuationAsOf)
	logStep("price_series", stepStart)
	stepStart = time.Now()
	metrics := computePortfolioMetrics(holdings, seriesByAssetKey)
	logStep("metrics", stepStart)

	stepStart = time.Now()
	profile, err := s.ensureUserProfile(ctx, batch.UserID)
	if err != nil {
		return err
	}
	logStep("load_profile", stepStart)

	stepStart = time.Now()
	futuresByAssetKey := fetchFuturesSnapshot(ctx, s.market, holdings)
	logStep("futures_snapshot", stepStart)

	providerPayload := map[string]any{
		"coingecko_ids":       mapKeys(priceMap),
		"marketstack_symbols": mapKeys(stockPrices),
		"open_exchange_rates": map[string]any{
			"base":      oerResp.Base,
			"timestamp": oerResp.Timestamp,
		},
		"binance_futures_symbols": futuresSymbols(futuresByAssetKey),
	}
	marketSnapshot := newMarketSnapshot(valuationAsOf)
	marketSnapshot.ProviderPayload = marshalJSON(providerPayload)

	baseCurrency, rateFromUSD := resolveDisplayCurrency(profile.BaseCurrency, oerResp.Rates)
	baseRateToUSD := rateToUSDFromRateFromUSD(rateFromUSD)

	portfolioSnapshot := PortfolioSnapshot{
		ID:                   newID("pf"),
		UserID:               batch.UserID,
		SourceUploadBatchID:  batch.ID,
		MarketDataSnapshotID: marketSnapshot.ID,
		ValuationAsOf:        valuationAsOf,
		NetWorthUSD:          metrics.NetWorthUSD,
		BaseCurrency:         baseCurrency,
		BaseFXRateToUSD:      &baseRateToUSD,
		SnapshotType:         "scan",
		Status:               "active",
		CreatedAt:            valuationAsOf,
	}

	calculationID := newID("calc")
	calculation := Calculation{
		ID:                  calculationID,
		PortfolioSnapshotID: portfolioSnapshot.ID,
		StatusPreview:       "processing",
		StatusPaid:          "not_started",
		HealthScore:         metrics.HealthScoreBaseline,
		VolatilityScore:     metrics.VolatilityScoreBaseline,
		HealthStatus:        healthStatusFromScore(metrics.HealthScoreBaseline),
		MetricsIncomplete:   metrics.MetricsIncomplete,
		PricedCoveragePct:   metrics.PricedCoveragePct,
		ModelVersionPreview: geminiModel,
		PromptHashPreview:   hashPrompt(s.prompts.PreviewReport),
		CreatedAt:           time.Now().UTC(),
	}

	stepStart = time.Now()
	if err := s.db.withTx(ctx, func(tx *gorm.DB) error {
		for _, asset := range resolvedAssets {
			if asset.AssetID == "" {
				continue
			}
			update := map[string]any{
				"symbol":       nullableString(asset.Holding.Symbol),
				"asset_type":   asset.Holding.AssetType,
				"asset_key":    nullableString(asset.Holding.AssetKey),
				"coingecko_id": nullableString(asset.Holding.CoinGeckoID),
				"exchange_mic": nullableString(asset.Holding.ExchangeMIC),
			}
			if err := tx.Model(&OCRAsset{}).Where("id = ?", asset.AssetID).Updates(update).Error; err != nil {
				return err
			}
		}
		for _, image := range images {
			update := map[string]any{
				"fingerprint_v1": fingerprintByImage[image.ID],
				"is_duplicate":   duplicateSet[image.ID],
			}
			if duplicateOfID, ok := duplicateOf[image.ID]; ok {
				update["duplicate_of_image_id"] = duplicateOfID
			} else {
				update["duplicate_of_image_id"] = nil
			}
			if err := tx.Model(&UploadImage{}).Where("id = ?", image.ID).Updates(update).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&marketSnapshot).Error; err != nil {
			return err
		}
		if err := tx.Create(&portfolioSnapshot).Error; err != nil {
			return err
		}
		for _, holding := range holdings {
			holdingRow := PortfolioHolding{
				ID:                  newID("ph"),
				PortfolioSnapshotID: portfolioSnapshot.ID,
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
			if err := tx.Create(&holdingRow).Error; err != nil {
				return err
			}
		}
		for _, holding := range holdings {
			item := buildMarketDataSnapshotItem(marketSnapshot.ID, holding)
			if item == nil {
				continue
			}
			if futures, ok := futuresByAssetKey[holding.AssetKey]; ok {
				item.RawPayload = marshalJSON(futures)
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&calculation).Error; err != nil {
			return err
		}

		if err := tx.Model(&User{}).Where("id = ?", batch.UserID).Update("active_portfolio_snapshot_id", portfolioSnapshot.ID).Error; err != nil {
			return err
		}

		return tx.Model(&UploadBatch{}).Where("id = ?", batch.ID).Updates(map[string]any{
			"status":       "completed",
			"completed_at": time.Now().UTC(),
		}).Error
	}); err != nil {
		return err
	}
	logStep("db_persist", stepStart)

	if s.hasActiveEntitlement(ctx, batch.UserID) {
		if err := s.db.DB().WithContext(ctx).
			Model(&Calculation{}).
			Where("calculation_id = ?", calculationID).
			Update("status_paid", "processing").Error; err != nil {
			return err
		}
		_ = s.queue.enqueue(ctx, jobPaidReport, calculationID)
		s.logger.Printf("normalize done batch=%s snapshot=%s calculation_id=%s paid_queued=true duration=%s", batch.ID, portfolioSnapshot.ID, calculationID, time.Since(start))
		return nil
	}

	_ = s.queue.enqueue(ctx, jobPreviewReport, calculationID)
	s.logger.Printf("normalize done batch=%s snapshot=%s calculation_id=%s paid_queued=false duration=%s", batch.ID, portfolioSnapshot.ID, calculationID, time.Since(start))
	return nil
}

func (s *Server) processPreviewReport(ctx context.Context, calculationID string) error {
	start := time.Now()
	logStep := func(step string, started time.Time) {
		s.logger.Printf("preview report step=%s calculation_id=%s duration=%s", step, calculationID, time.Since(started))
	}
	var calculation Calculation
	stepStart := time.Now()
	if err := s.db.DB().WithContext(ctx).First(&calculation, "calculation_id = ?", calculationID).Error; err != nil {
		return err
	}
	var snapshot PortfolioSnapshot
	if err := s.db.DB().WithContext(ctx).First(&snapshot, "id = ?", calculation.PortfolioSnapshotID).Error; err != nil {
		return err
	}
	var holdingsRows []PortfolioHolding
	if err := s.db.DB().WithContext(ctx).Where("portfolio_snapshot_id = ?", snapshot.ID).Find(&holdingsRows).Error; err != nil {
		return err
	}
	logStep("load_snapshot", stepStart)

	s.logger.Printf("preview report start calculation_id=%s holdings=%d", calculationID, len(holdingsRows))

	stepStart = time.Now()
	holdings := mapHoldingsFromRows(holdingsRows)
	oerResp, _ := s.market.openExchangeLatest(ctx)
	asOf := snapshot.ValuationAsOf
	if asOf.IsZero() {
		asOf = priceAsOf()
	}
	seriesByAssetKey := fetchPriceSeriesAsOf(ctx, s.market, holdings, asOf)
	logStep("price_series", stepStart)
	stepStart = time.Now()
	metrics := computePortfolioMetrics(holdings, seriesByAssetKey)
	logStep("metrics", stepStart)

	stepStart = time.Now()
	profile, err := s.ensureUserProfile(ctx, snapshot.UserID)
	if err != nil {
		return err
	}
	logStep("load_profile", stepStart)

	stepStart = time.Now()
	userProfile := userProfile{
		RiskTolerance:  profile.RiskLevel,
		RiskPreference: profile.RiskPreference,
		PainPoints:     []string(profile.PainPoints),
		Experience:     profile.Experience,
		Style:          profile.Style,
		Markets:        []string(profile.Markets),
		Timezone:       profile.Timezone,
	}
	logStep("build_profile_payload", stepStart)

	valuationAsOf := snapshot.ValuationAsOf.Format(time.RFC3339)
	baseCurrency, rateFromUSD := resolveDisplayCurrency(profile.BaseCurrency, oerResp.Rates)
	baseRateToUSD := rateToUSDFromRateFromUSD(rateFromUSD)
	netWorthDisplay := metrics.NetWorthUSD * rateFromUSD
	fixedMetrics := previewFixedMetrics{
		NetWorthUSD:     metrics.NetWorthUSD,
		HealthScore:     metrics.HealthScoreBaseline,
		HealthStatus:    healthStatusFromScore(metrics.HealthScoreBaseline),
		VolatilityScore: metrics.VolatilityScoreBaseline,
	}
	stepStart = time.Now()
	previewInput := buildPreviewPayload(calculation.ID, valuationAsOf, snapshot.MarketDataSnapshotID, userProfile, holdings, metrics, fixedMetrics, netWorthDisplay, baseCurrency, baseRateToUSD)
	logStep("build_payload", stepStart)

	outputLanguage := resolveOutputLanguage(profile.Language, "")
	prompt := applyOutputLanguage(s.prompts.PreviewReport, outputLanguage)

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: prompt}}},
		Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: mustJSON(previewInput)}}}},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.4,
			MaxOutputTokens:  geminiMaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	var previewLLM previewPromptOutput
	var violations []string
	// Empirical note: app-backend/preview_retry_audit.log shows all previews passed on attempt=1
	// across 4 portfolios x 3 profiles x 2 languages x 3 runs; retry is a safety net we can remove later.
	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = s.callGeminiJSONWithRetry(ctx, request, &previewLLM, "preview report", calculationID)
		if err != nil {
			return s.db.DB().WithContext(ctx).Model(&Calculation{}).Where("calculation_id = ?", calculationID).Update("status_preview", "failed").Error
		}
		violations = validatePreviewRisks(previewLLM.IdentifiedRisks, metrics)
		if len(violations) == 0 {
			break
		}
		if attempt < maxAttempts-1 {
			s.logger.Printf("preview report validation failed calculation_id=%s attempt=%d violations=%v", calculationID, attempt+1, violations)
			retryPrompt := fmt.Sprintf("%s\n\n### Validation Fix\nYour previous response violated: %s. Regenerate the full JSON and fix all issues.", prompt, strings.Join(violations, "; "))
			request.SystemInstruction = &geminiSystemInstruction{Parts: []geminiPart{{Text: retryPrompt}}}
			continue
		}
	}
	if len(violations) > 0 {
		s.logger.Printf("preview report validation failed calculation_id=%s violations=%v", calculationID, violations)
		return s.db.DB().WithContext(ctx).Model(&Calculation{}).Where("calculation_id = ?", calculationID).Update("status_preview", "failed").Error
	}

	preview := previewFromPromptOutput(previewLLM)
	preview.MetaData.CalculationID = calculation.ID
	preview.ValuationAsOf = valuationAsOf
	preview.MarketDataSnapshotID = snapshot.MarketDataSnapshotID
	preview.FixedMetrics = fixedMetrics
	stepStart = time.Now()
	normalizePreviewStatus(&preview)
	logStep("merge_risks", stepStart)
	stepStart = time.Now()
	preview.BaseCurrency = baseCurrency
	preview.BaseFXRateToUSD = baseRateToUSD
	preview.NetWorthDisplay = netWorthDisplay
	preview.AssetAllocation = buildAssetAllocation(holdings, metrics.NetWorthUSD, baseCurrency, rateFromUSD)
	logStep("display_fields", stepStart)

	payload := marshalJSON(preview)
	stepStart = time.Now()
	err = s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&Calculation{}).Where("calculation_id = ?", calculationID).Updates(map[string]any{
			"status_preview":      "ready",
			"health_score":        preview.FixedMetrics.HealthScore,
			"volatility_score":    preview.FixedMetrics.VolatilityScore,
			"health_status":       preview.FixedMetrics.HealthStatus,
			"preview_payload":     payload,
			"metrics_incomplete":  metrics.MetricsIncomplete,
			"priced_coverage_pct": metrics.PricedCoveragePct,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("calculation_id = ?", calculationID).Delete(&ReportRisk{}).Error; err != nil {
			return err
		}
		for _, risk := range preview.IdentifiedRisks {
			teaser := risk.TeaserText
			row := ReportRisk{
				ID:            newID("risk"),
				CalculationID: calculationID,
				RiskID:        risk.RiskID,
				Type:          risk.Type,
				Severity:      risk.Severity,
				TeaserText:    &teaser,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	logStep("persist", stepStart)
	if err == nil {
		s.logger.Printf("preview report done calculation_id=%s duration=%s", calculationID, time.Since(start))
	}
	return err
}

func (s *Server) processPaidReport(ctx context.Context, calculationID string) error {
	start := time.Now()
	logStep := func(step string, started time.Time) {
		s.logger.Printf("paid report step=%s calculation_id=%s duration=%s", step, calculationID, time.Since(started))
	}
	var calculation Calculation
	stepStart := time.Now()
	if err := s.db.DB().WithContext(ctx).First(&calculation, "calculation_id = ?", calculationID).Error; err != nil {
		return err
	}

	var snapshot PortfolioSnapshot
	if err := s.db.DB().WithContext(ctx).First(&snapshot, "id = ?", calculation.PortfolioSnapshotID).Error; err != nil {
		return err
	}
	var holdingsRows []PortfolioHolding
	if err := s.db.DB().WithContext(ctx).Where("portfolio_snapshot_id = ?", snapshot.ID).Find(&holdingsRows).Error; err != nil {
		return err
	}
	logStep("load_snapshot", stepStart)
	s.logger.Printf("paid report start calculation_id=%s holdings=%d", calculationID, len(holdingsRows))
	stepStart = time.Now()
	holdings := mapHoldingsFromRows(holdingsRows)
	oerResp, _ := s.market.openExchangeLatest(ctx)
	asOf := snapshot.ValuationAsOf
	if asOf.IsZero() {
		asOf = priceAsOf()
	}
	seriesByAssetKey := fetchPriceSeriesAsOf(ctx, s.market, holdings, asOf)
	logStep("price_series", stepStart)
	stepStart = time.Now()
	metrics := computePortfolioMetrics(holdings, seriesByAssetKey)
	logStep("metrics", stepStart)

	stepStart = time.Now()
	profile, err := s.ensureUserProfile(ctx, snapshot.UserID)
	if err != nil {
		return err
	}
	logStep("load_profile", stepStart)
	userProfile := userProfile{
		RiskTolerance:  profile.RiskLevel,
		RiskPreference: profile.RiskPreference,
		PainPoints:     []string(profile.PainPoints),
		Experience:     profile.Experience,
		Style:          profile.Style,
		Markets:        []string(profile.Markets),
		Timezone:       profile.Timezone,
	}
	outputLanguage := resolveOutputLanguage(profile.Language, "")
	valuationAsOf := snapshot.ValuationAsOf.Format(time.RFC3339)
	deviceTimezone := s.loadBatchTimezone(ctx, snapshot.SourceUploadBatchID)
	baseCurrency, rateFromUSD := resolveDisplayCurrency(profile.BaseCurrency, oerResp.Rates)
	baseRateToUSD := rateToUSDFromRateFromUSD(rateFromUSD)
	netWorthDisplay := metrics.NetWorthUSD * rateFromUSD
	fixedMetrics := previewFixedMetrics{
		NetWorthUSD:     metrics.NetWorthUSD,
		HealthScore:     metrics.HealthScoreBaseline,
		HealthStatus:    healthStatusFromScore(metrics.HealthScoreBaseline),
		VolatilityScore: metrics.VolatilityScoreBaseline,
	}
	stepStart = time.Now()
	futuresByAssetKey := fetchFuturesSnapshot(ctx, s.market, holdings)
	logStep("futures_snapshot", stepStart)
	stepStart = time.Now()
	plansUSD := buildLockedPlans(userProfile, holdings, metrics, seriesByAssetKey, futuresByAssetKey, snapshot.ID, deviceTimezone)
	planStates := buildInitialPlanStates(snapshot.UserID, plansUSD, futuresByAssetKey, snapshot.ValuationAsOf)
	logStep("build_plans", stepStart)
	holdingByAsset := holdingsByAssetKey(holdings)

	var (
		paidLLM         paidPromptOutput
		paid            paidReportPayload
		riskSeed        []previewRisk
		prompt          string
		promptHash      string
		plansDisplay    []lockedPlan
		healthScore     int
		volatilityScore int
	)

	if calculation.StatusPreview == "ready" && len(calculation.PreviewPayload) > 0 {
		var preview previewReportPayload
		stepStart = time.Now()
		if err := json.Unmarshal(calculation.PreviewPayload, &preview); err != nil {
			return err
		}
		logStep("load_preview_payload", stepStart)
		stepStart = time.Now()
		plansUSD = assignLinkedRiskIDs(plansUSD, preview.IdentifiedRisks)
		logStep("merge_risks", stepStart)
		riskSeed = preview.IdentifiedRisks
		stepStart = time.Now()
		plansDisplay = convertLockedPlansToDisplayByPlan(plansUSD, holdingByAsset, baseCurrency, rateFromUSD)
		paidInput := buildPaidPayload(userProfile, holdings, preview, plansDisplay, metrics, fixedMetrics, netWorthDisplay, baseCurrency, baseRateToUSD)
		logStep("build_payload", stepStart)
		prompt = applyOutputLanguage(s.prompts.PaidReport, outputLanguage)
		promptHash = hashPrompt(s.prompts.PaidReport)
		request := geminiRequest{
			SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: prompt}}},
			Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: mustJSON(paidInput)}}}},
			GenerationConfig: geminiGenerationConfig{
				Temperature:      0.4,
				MaxOutputTokens:  geminiMaxOutputTokens,
				ResponseMimeType: "application/json",
			},
		}
		err = s.callGeminiJSONWithRetry(ctx, request, &paidLLM, "paid report", calculationID)
		if err != nil {
			return s.db.DB().WithContext(ctx).Model(&Calculation{}).Where("calculation_id = ?", calculationID).Update("status_paid", "failed").Error
		}
		healthScore = fixedMetrics.HealthScore
		volatilityScore = fixedMetrics.VolatilityScore
	} else {
		computedRisks := buildIdentifiedRisks(metrics)
		plansUSD = assignLinkedRiskIDs(plansUSD, computedRisks)
		riskSeed = computedRisks
		stepStart = time.Now()
		plansDisplay = convertLockedPlansToDisplayByPlan(plansUSD, holdingByAsset, baseCurrency, rateFromUSD)
		paidInput := buildDirectPaidPayload(calculation.ID, valuationAsOf, snapshot.MarketDataSnapshotID, userProfile, holdings, metrics, riskSeed, plansDisplay, fixedMetrics, netWorthDisplay, baseCurrency, baseRateToUSD)
		logStep("build_payload", stepStart)
		prompt = applyOutputLanguage(s.prompts.PaidReportDirect, outputLanguage)
		promptHash = hashPrompt(s.prompts.PaidReportDirect)
		request := geminiRequest{
			SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: prompt}}},
			Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: mustJSON(paidInput)}}}},
			GenerationConfig: geminiGenerationConfig{
				Temperature:      0.4,
				MaxOutputTokens:  geminiMaxOutputTokens,
				ResponseMimeType: "application/json",
			},
		}
		err = s.callGeminiJSONWithRetry(ctx, request, &paidLLM, "paid report", calculationID)
		if err != nil {
			return s.db.DB().WithContext(ctx).Model(&Calculation{}).Where("calculation_id = ?", calculationID).Update("status_paid", "failed").Error
		}
		healthScore = fixedMetrics.HealthScore
		volatilityScore = fixedMetrics.VolatilityScore
	}

	paid = paidFromPromptOutput(paidLLM)
	paid.MetaData.CalculationID = calculation.ID
	paid.ValuationAsOf = valuationAsOf
	paid.MarketDataSnapshotID = snapshot.MarketDataSnapshotID
	paid.ReportHeader.HealthScore.Value = healthScore
	paid.ReportHeader.VolatilityDashboard.Value = volatilityScore
	paid.ReportHeader.HealthScore.Status = scoreStatusFromScore(healthScore)
	paid.ReportHeader.VolatilityDashboard.Status = volatilityStatusFromScore(volatilityScore)
	stepStart = time.Now()
	mergePaidPlanParameters(&paid, plansDisplay)
	mergePaidRisks(&paid, riskSeed)
	applyPlanMetadata(&paid, riskSeed)
	paid.ActionableAdvice = append([]paidPlan(nil), paid.OptimizationPlan...)
	logStep("merge_plans_risks", stepStart)
	stepStart = time.Now()
	paid.BaseCurrency = baseCurrency
	paid.BaseFXRateToUSD = baseRateToUSD
	paid.NetWorthDisplay = netWorthDisplay
	paid.AssetAllocation = buildAssetAllocation(holdings, metrics.NetWorthUSD, baseCurrency, rateFromUSD)
	logStep("display_fields", stepStart)
	stepStart = time.Now()
	paid.Charts.RadarChart = buildRadarChart(metrics, computeAlpha30d(ctx, s.market, holdings, metrics, seriesByAssetKey))
	logStep("radar_chart", stepStart)
	stepStart = time.Now()
	paid.DailyAlphaSignal = selectDailyAlphaSignal(ctx, s.market, holdings, seriesByAssetKey, outputLanguage)
	logStep("daily_alpha", stepStart)
	stepStart = time.Now()
	paid.RiskSummary = paid.TheVerdict.ConstructiveComment
	sortOptimizationPlans(&paid, plansUSD, riskSeed)
	logStep("sort_plans", stepStart)

	payload := marshalJSON(paid)
	stepStart = time.Now()
	err = s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&Calculation{}).Where("calculation_id = ?", calculationID).Updates(map[string]any{
			"status_paid":         "ready",
			"paid_payload":        payload,
			"paid_at":             time.Now().UTC(),
			"model_version_paid":  geminiModel,
			"prompt_hash_paid":    promptHash,
			"health_score":        healthScore,
			"volatility_score":    volatilityScore,
			"health_status":       healthStatusFromScore(healthScore),
			"metrics_incomplete":  metrics.MetricsIncomplete,
			"priced_coverage_pct": metrics.PricedCoveragePct,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("calculation_id = ?", calculationID).Delete(&ReportStrategy{}).Error; err != nil {
			return err
		}
		if err := tx.Where("calculation_id = ?", calculationID).Delete(&ReportRisk{}).Error; err != nil {
			return err
		}
		teaserByID := make(map[string]string)
		for _, risk := range riskSeed {
			if risk.TeaserText != "" {
				teaserByID[risk.RiskID] = risk.TeaserText
			}
		}
		for _, risk := range paid.RiskInsights {
			message := risk.Message
			teaser := teaserByID[risk.RiskID]
			row := ReportRisk{
				ID:            newID("risk"),
				CalculationID: calculationID,
				RiskID:        risk.RiskID,
				Type:          risk.Type,
				Severity:      risk.Severity,
				TeaserText:    nullableString(teaser),
				Message:       &message,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		rationaleByPlan := make(map[string]paidPlan)
		for _, plan := range paid.OptimizationPlan {
			rationaleByPlan[plan.PlanID] = plan
		}
		for _, plan := range plansUSD {
			merged := rationaleByPlan[plan.PlanID]
			row := ReportStrategy{
				ID:              newID("plan"),
				CalculationID:   calculationID,
				PlanID:          plan.PlanID,
				StrategyID:      plan.StrategyID,
				AssetType:       plan.AssetType,
				Symbol:          plan.Symbol,
				AssetKey:        plan.AssetKey,
				LinkedRiskID:    plan.LinkedRiskID,
				Parameters:      marshalJSON(plan.Parameters),
				Rationale:       merged.Rationale,
				ExpectedOutcome: merged.ExpectedOutcome,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("user_id = ?", snapshot.UserID).Delete(&PlanState{}).Error; err != nil {
			return err
		}
		for _, state := range planStates {
			if err := tx.Create(&state).Error; err != nil {
				return err
			}
		}
		return nil
	})
	logStep("persist", stepStart)
	if err == nil {
		s.logger.Printf("paid report done calculation_id=%s duration=%s", calculationID, time.Since(start))
	}
	return err
}

func (s *Server) processInsightsRefresh(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	if !s.hasActiveEntitlement(ctx, userID) {
		return nil
	}

	var user User
	if err := s.db.DB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return err
	}
	if user.ActivePortfolioSnapshot == nil {
		return nil
	}

	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return err
	}
	outputLanguage := resolveOutputLanguage(profile.Language, "")

	snapshot, holdings, err := s.loadSnapshotWithHoldings(ctx, *user.ActivePortfolioSnapshot)
	if err != nil {
		return err
	}

	_, plans, riskByID, err := s.loadLockedPlansForSnapshot(ctx, snapshot.ID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if len(plans) == 0 {
		return s.db.DB().WithContext(ctx).Model(&UserProfile{}).Where("user_id = ?", userID).Update("insights_refreshed_at", now).Error
	}

	var planRows []PlanState
	if err := s.db.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&planRows).Error; err != nil {
		return err
	}
	planStates := decodePlanStates(planRows)
	if len(planStates) == 0 {
		futuresByAssetKey := fetchFuturesSnapshot(ctx, s.market, holdings)
		initialStates := buildInitialPlanStates(userID, plans, futuresByAssetKey, snapshot.ValuationAsOf)
		if len(initialStates) > 0 {
			if err := s.db.DB().WithContext(ctx).Create(&initialStates).Error; err != nil {
				return err
			}
			planStates = decodePlanStates(initialStates)
		}
	}

	oerResp, _ := s.market.openExchangeLatest(ctx)
	seriesByAssetKey := fetchPriceSeries(ctx, s.market, holdings)
	futuresByAssetKey := fetchFuturesSnapshot(ctx, s.market, holdings)
	baseCurrency, rateFromUSD := resolveDisplayCurrency(profile.BaseCurrency, oerResp.Rates)
	items, _ := buildInsights(ctx, s.market, holdings, plans, planStates, seriesByAssetKey, futuresByAssetKey, riskByID, outputLanguage, baseCurrency, rateFromUSD, now)

	if err := s.db.DB().WithContext(ctx).Model(&Insight{}).
		Where("user_id = ? AND status = ? AND expires_at <= ?", userID, "active", now).
		Update("status", "expired").Error; err != nil {
		return err
	}

	var existing []Insight
	if err := s.db.DB().WithContext(ctx).Where("user_id = ? AND expires_at > ?", userID, now).Find(&existing).Error; err != nil {
		return err
	}
	existingKeys := make(map[string]struct{}, len(existing))
	for _, row := range existing {
		existingKeys[row.Type+"|"+row.TriggerKey] = struct{}{}
	}

	newInsights := make([]Insight, 0)
	for _, item := range items {
		key := item.Type + "|" + item.TriggerKey
		if _, ok := existingKeys[key]; ok {
			continue
		}
		newInsights = append(newInsights, buildInsightRecord(userID, item))
	}

	err = s.db.withTx(ctx, func(tx *gorm.DB) error {
		for _, planState := range planStates {
			if !planState.Updated {
				continue
			}
			if err := tx.Model(&PlanState{}).Where("id = ?", planState.ID).Updates(map[string]any{
				"state_json": marshalJSON(planState.State),
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		}

		for _, insight := range newInsights {
			if err := tx.Create(&insight).Error; err != nil {
				return err
			}
		}

		return tx.Model(&UserProfile{}).Where("user_id = ?", userID).Update("insights_refreshed_at", now).Error
	})
	if err != nil {
		return err
	}
	if len(newInsights) > 0 {
		if err := s.sendInsightPushNotifications(ctx, userID, newInsights); err != nil {
			s.logger.Printf("insights push error user=%s err=%v", userID, err)
		}
	}
	return nil
}

func (s *Server) markOCRFailure(ctx context.Context, batch *UploadBatch, images []UploadImage, code string, reason string) error {
	return s.markBatchFailed(ctx, batch.ID, code)
}

func (s *Server) markBatchFailed(ctx context.Context, batchID string, code string) error {
	return s.db.DB().WithContext(ctx).Model(&UploadBatch{}).Where("id = ?", batchID).Updates(map[string]any{
		"status":       "failed",
		"error_code":   code,
		"completed_at": time.Now().UTC(),
	}).Error
}

func (s *Server) updateBatchStatus(ctx context.Context, batchID string, status string, errorCode string) error {
	updates := map[string]any{"status": status}
	if errorCode != "" {
		updates["error_code"] = errorCode
	}
	return s.db.DB().WithContext(ctx).Model(&UploadBatch{}).Where("id = ?", batchID).Updates(updates).Error
}

func imageIDs(images []UploadImage) []string {
	ids := make([]string, 0, len(images))
	for _, image := range images {
		ids = append(ids, image.ID)
	}
	return ids
}

func buildOCRInput(images []UploadImage) string {
	var builder strings.Builder
	builder.WriteString("Input images:\n")
	for _, image := range images {
		builder.WriteString("- image_id: ")
		builder.WriteString(image.ID)
		builder.WriteString("\n")
	}
	builder.WriteString("Return strict JSON per the system prompt.")
	return builder.String()
}

func computeConfidence(asset ocrAsset) float64 {
	if strings.TrimSpace(asset.SymbolRaw) == "" || asset.Amount == 0 {
		return 0.4
	}
	return 0.9
}

func computeFingerprintV0(platformGuess string, assets []ocrAsset) string {
	if len(assets) == 0 {
		return ""
	}
	entries := make([]string, 0, len(assets))
	for _, asset := range assets {
		key := normalizeSymbol(asset.SymbolRaw)
		entries = append(entries, fmt.Sprintf("%s:%f", key, asset.Amount))
	}
	sort.Strings(entries)
	payload := platformGuess + "|" + strings.Join(entries, "|")
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

func computeFingerprintV1(platformGuess string, assets []resolvedAsset) string {
	if len(assets) == 0 {
		return ""
	}
	entries := make([]string, 0, len(assets))
	for _, asset := range assets {
		if asset.Holding.AssetKey == "" {
			continue
		}
		entries = append(entries, fmt.Sprintf("%s:%f", asset.Holding.AssetKey, asset.Holding.Amount))
	}
	if len(entries) == 0 {
		return ""
	}
	sort.Strings(entries)
	payload := platformGuess + "|" + strings.Join(entries, "|")
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

func normalizeOCRStatus(status string) string {
	status = strings.TrimSpace(status)
	switch status {
	case "success", "ignored_invalid", "ignored_unsupported", "ignored_blurry":
		return status
	default:
		return ""
	}
}

func countOCRStatuses(images []ocrImage) (int, int, int) {
	success := 0
	ignored := 0
	unsupported := 0
	for _, image := range images {
		switch image.Status {
		case "success":
			success++
		case "ignored_unsupported":
			ignored++
			unsupported++
		case "ignored_invalid", "ignored_blurry":
			ignored++
		}
	}
	return success, ignored, unsupported
}

func mapBatchError(images []ocrImage) string {
	anyParse := false
	allUnsupported := true
	allInvalid := true
	for _, image := range images {
		if image.ErrorReason != nil && *image.ErrorReason == "PARSE_ERROR" {
			anyParse = true
		}
		if image.Status != "ignored_unsupported" {
			allUnsupported = false
		}
		if image.Status != "ignored_invalid" && image.Status != "ignored_blurry" {
			allInvalid = false
		}
	}
	if anyParse {
		return "EXTRACTION_FAILURE"
	}
	if allUnsupported {
		return "UNSUPPORTED_ASSET_VIEW"
	}
	if allInvalid {
		return "INVALID_IMAGE"
	}
	return "EXTRACTION_FAILURE"
}

func hashPrompt(prompt string) string {
	hash := sha256.Sum256([]byte(prompt))
	return hex.EncodeToString(hash[:])
}

func mustJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func clampScore(value int, baseline int) int {
	min := baseline - 5
	max := baseline + 5
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func nullableString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	return &trimmed
}

func (s *Server) callGeminiJSONWithRetry(ctx context.Context, request geminiRequest, out any, label, calculationID string) error {
	var lastErr error
	for attempt := 0; attempt < geminiMaxAttempts; attempt++ {
		callStart := time.Now()
		_, err := s.gemini.callGeminiJSON(ctx, request, out)
		if err == nil {
			s.logger.Printf("%s gemini done calculation_id=%s duration=%s", label, calculationID, time.Since(callStart))
			return nil
		}
		lastErr = err
		s.logger.Printf("%s gemini error calculation_id=%s attempt=%d duration=%s err=%v", label, calculationID, attempt+1, time.Since(callStart), err)
		if attempt < geminiMaxAttempts-1 {
			delay := geminiJobRetryDelay(attempt)
			s.logger.Printf("%s gemini retry calculation_id=%s attempt=%d wait=%s", label, calculationID, attempt+1, delay)
			if err := sleepWithContext(ctx, delay); err != nil {
				return err
			}
		}
	}
	return lastErr
}

func geminiJobRetryDelay(attempt int) time.Duration {
	delay := geminiRetryBaseDelay * 5 * time.Duration(1<<attempt)
	if delay > time.Minute {
		return time.Minute
	}
	return delay
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
