package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type uploadBatchCreateRequest struct {
	Purpose        string               `json:"purpose"`
	ImageCount     int                  `json:"image_count"`
	Images         []uploadBatchImageIn `json:"images"`
	DeviceTimezone string               `json:"device_timezone"`
}

type uploadBatchImageIn struct {
	FileName  string `json:"file_name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type uploadBatchCreateResponse struct {
	UploadBatchID string              `json:"upload_batch_id"`
	Status        string              `json:"status"`
	ImageUploads  []uploadBatchUpload `json:"image_uploads"`
	ExpiresAt     string              `json:"expires_at"`
}

type uploadBatchUpload struct {
	ImageID   string            `json:"image_id"`
	UploadURL string            `json:"upload_url"`
	Headers   map[string]string `json:"headers"`
}

type uploadBatchCompleteRequest struct {
	ImageIDs       []string `json:"image_ids"`
	ClientChecksum string   `json:"client_checksum"`
}

type uploadBatchReviewRequest struct {
	PlatformOverrides  []platformOverrideRequest    `json:"platform_overrides"`
	Resolutions        []ambiguityResolutionRequest `json:"resolutions"`
	Edits              []ocrAssetEditRequest        `json:"edits"`
	DuplicateOverrides []duplicateOverrideRequest   `json:"duplicate_overrides"`
}

type platformOverrideRequest struct {
	ImageID       string `json:"image_id"`
	PlatformGuess string `json:"platform_guess"`
}

type ambiguityResolutionRequest struct {
	SymbolRaw   string `json:"symbol_raw"`
	AssetType   string `json:"asset_type"`
	Symbol      string `json:"symbol"`
	ExchangeMIC string `json:"exchange_mic"`
	AssetKey    string `json:"asset_key"`
}

type ocrAssetEditRequest struct {
	AssetID             string   `json:"asset_id"`
	Action              string   `json:"action"`
	Symbol              *string  `json:"symbol"`
	AssetType           *string  `json:"asset_type"`
	Amount              *float64 `json:"amount"`
	ValueFromScreenshot *float64 `json:"value_from_screenshot"`
	DisplayCurrency     *string  `json:"display_currency"`
	AvgPrice            *float64 `json:"avg_price"`
	ManualValueUSD      *float64 `json:"manual_value_usd"`
	ManualValueDisplay  *float64 `json:"manual_value_display"`
}

type duplicateOverrideRequest struct {
	ImageID string `json:"image_id"`
	Include bool   `json:"include"`
}

func (s *Server) handleUploadBatchCreate(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	var req uploadBatchCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	purpose := strings.ToLower(strings.TrimSpace(req.Purpose))
	if purpose != "holdings" && purpose != "trade_slip" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid purpose", nil)
		return
	}
	if req.ImageCount <= 0 || req.ImageCount != len(req.Images) {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "image_count mismatch", nil)
		return
	}
	if purpose == "holdings" && req.ImageCount > 15 {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "max 15 images", nil)
		return
	}
	if purpose == "trade_slip" && req.ImageCount != 1 {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "trade slip requires 1 image", nil)
		return
	}

	if purpose == "holdings" {
		allowed, nextReset, err := s.checkHoldingsQuota(r.Context(), userID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "QUOTA_ERROR", "quota check failed", nil)
			return
		}
		if !allowed {
			s.writeError(w, http.StatusTooManyRequests, "QUOTA_EXCEEDED", "daily quota exceeded", map[string]any{"next_reset_at": nextReset})
			return
		}
	}

	now := time.Now().UTC()
	batch := UploadBatch{
		ID:             newID("ub"),
		UserID:         userID,
		Purpose:        purpose,
		Status:         "pending_upload",
		ImageCount:     req.ImageCount,
		DeviceTimezone: nullableString(req.DeviceTimezone),
		CreatedAt:      now,
	}

	uploads := make([]uploadBatchUpload, 0, len(req.Images))
	images := make([]UploadImage, 0, len(req.Images))
	for _, image := range req.Images {
		imageID := newID("img")
		storageKey := buildStorageKey(userID, batch.ID, imageID, image.FileName)
		images = append(images, UploadImage{
			ID:            imageID,
			UploadBatchID: batch.ID,
			StorageKey:    storageKey,
			CreatedAt:     now,
		})
		url, headers, expires, err := s.storage.presignPut(r.Context(), storageKey, image.MimeType, 15*time.Minute)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "failed to generate upload url", nil)
			return
		}
		uploads = append(uploads, uploadBatchUpload{ImageID: imageID, UploadURL: url, Headers: headers})
		_ = expires
	}

	if err := s.db.withTx(r.Context(), func(tx *gorm.DB) error {
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}
		for _, image := range images {
			if err := tx.Create(&image).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "failed to create batch", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, uploadBatchCreateResponse{
		UploadBatchID: batch.ID,
		Status:        batch.Status,
		ImageUploads:  uploads,
		ExpiresAt:     now.Add(15 * time.Minute).Format(time.RFC3339),
	})
}

func (s *Server) handleUploadBatchComplete(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	batchID := chiURLParam(r, "upload_batch_id")
	var batch UploadBatch
	if err := s.db.DB().WithContext(r.Context()).First(&batch, "id = ? AND user_id = ?", batchID, userID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found", nil)
		return
	}
	var req uploadBatchCompleteRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if len(req.ImageIDs) == 0 {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "image_ids required", nil)
		return
	}
	var images []UploadImage
	if err := s.db.DB().WithContext(r.Context()).Where("upload_batch_id = ?", batch.ID).Find(&images).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "failed to load images", nil)
		return
	}
	imageSet := make(map[string]UploadImage)
	for _, image := range images {
		imageSet[image.ID] = image
	}
	for _, id := range req.ImageIDs {
		image, ok := imageSet[id]
		if !ok {
			s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "image_id not in batch", nil)
			return
		}
		if err := s.storage.headObject(r.Context(), image.StorageKey); err != nil {
			s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "image not uploaded", nil)
			return
		}
	}
	if strings.TrimSpace(req.ClientChecksum) != "" {
		checksum, err := s.computeClientChecksum(r.Context(), req.ImageIDs, imageSet)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "client_checksum verification failed", nil)
			return
		}
		if !strings.EqualFold(checksum, strings.TrimSpace(req.ClientChecksum)) {
			s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "client_checksum mismatch", nil)
			return
		}
	}

	if err := s.db.DB().WithContext(r.Context()).Model(&UploadBatch{}).Where("id = ?", batch.ID).Update("status", "processing").Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "failed to update batch", nil)
		return
	}

	jobType := jobOCRHoldings
	if batch.Purpose == "trade_slip" {
		jobType = jobOCRTradeSlip
	}
	_ = s.queue.enqueue(r.Context(), jobType, batch.ID)

	s.writeJSON(w, http.StatusOK, map[string]any{"status": "processing", "poll_after_ms": 1500})
}

func (s *Server) handleUploadBatchGet(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	batchID := chiURLParam(r, "upload_batch_id")
	var batch UploadBatch
	if err := s.db.DB().WithContext(r.Context()).First(&batch, "id = ? AND user_id = ?", batchID, userID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found", nil)
		return
	}

	switch batch.Status {
	case "processing":
		s.writeJSON(w, http.StatusOK, map[string]any{"status": batch.Status})
	case "needs_review":
		s.writeJSON(w, http.StatusOK, s.buildNeedsReviewResponse(r.Context(), batch))
	case "completed":
		snapshotType := "scan"
		if batch.Purpose == "trade_slip" {
			snapshotType = "delta"
		}
		snapshotID := s.findSnapshotForBatch(r.Context(), batch.ID, snapshotType)
		if batch.Purpose == "trade_slip" {
			transactions := s.findTransactionsForSnapshot(r.Context(), snapshotID)
			warnings := make([]string, 0, len(batch.Warnings))
			warnings = append(warnings, batch.Warnings...)
			s.writeJSON(w, http.StatusOK, map[string]any{
				"status":                "completed",
				"portfolio_snapshot_id": snapshotID,
				"transaction_ids":       transactions,
				"warnings":              warnings,
			})
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]any{
			"status":                "completed",
			"portfolio_snapshot_id": snapshotID,
			"calculation_id":        s.findCalculationForBatch(r.Context(), batch.ID),
		})
	case "failed":
		s.writeJSON(w, http.StatusOK, map[string]any{"status": "failed", "error_code": batch.ErrorCode})
	default:
		s.writeJSON(w, http.StatusOK, map[string]any{"status": batch.Status})
	}
}

func (s *Server) handleUploadBatchReview(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	batchID := chiURLParam(r, "upload_batch_id")
	var batch UploadBatch
	if err := s.db.DB().WithContext(r.Context()).First(&batch, "id = ? AND user_id = ?", batchID, userID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found", nil)
		return
	}
	var req uploadBatchReviewRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	imageCategories := make(map[string]string)
	var images []UploadImage
	if err := s.db.DB().WithContext(r.Context()).Where("upload_batch_id = ?", batch.ID).Find(&images).Error; err == nil {
		for _, image := range images {
			imageCategories[image.ID] = platformGuessToCategory(image.PlatformGuess)
		}
	}
	symbolCategories := make(map[string]string)
	var ambiguities []OCRAmbiguity
	if err := s.db.DB().WithContext(r.Context()).Where("upload_batch_id = ?", batch.ID).Find(&ambiguities).Error; err == nil {
		for _, amb := range ambiguities {
			if category, ok := imageCategories[amb.UploadImageID]; ok {
				symbolCategories[amb.SymbolRaw] = category
			}
		}
	}

	var baseCurrency string
	var baseRateToUSD float64
	var baseRates map[string]float64
	needsManualConversion := false
	for _, edit := range req.Edits {
		if edit.ManualValueDisplay != nil || edit.AvgPrice != nil {
			needsManualConversion = true
			break
		}
	}
	if needsManualConversion {
		profile, err := s.ensureUserProfile(r.Context(), userID)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "PROFILE_ERROR", "failed to load profile", nil)
			return
		}
		oerResp, err := s.market.openExchangeLatest(r.Context())
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "FX_ERROR", "failed to load FX rates", nil)
			return
		}
		baseRates = oerResp.Rates
		baseCurrency = normalizeCurrency(profile.BaseCurrency)
		if rateToUSD, ok := currencyRateToUSD(baseCurrency, baseRates); ok {
			baseRateToUSD = rateToUSD
		} else {
			baseCurrency = "USD"
			baseRateToUSD = 1
		}
	}

	if err := s.db.withTx(r.Context(), func(tx *gorm.DB) error {
		for _, override := range req.PlatformOverrides {
			if override.ImageID == "" {
				continue
			}
			if err := tx.Model(&UploadImage{}).Where("id = ? AND upload_batch_id = ?", override.ImageID, batch.ID).Update("platform_guess", override.PlatformGuess).Error; err != nil {
				return err
			}
		}
		for _, edit := range req.Edits {
			if edit.AssetID == "" {
				continue
			}
			if strings.EqualFold(edit.Action, "remove") {
				if err := tx.Where("id = ?", edit.AssetID).Delete(&OCRAsset{}).Error; err != nil {
					return err
				}
				continue
			}
			updates := map[string]any{}
			if edit.Symbol != nil {
				updates["symbol"] = edit.Symbol
			}
			if edit.AssetType != nil {
				updates["asset_type"] = *edit.AssetType
			}
			if edit.Amount != nil {
				updates["amount"] = *edit.Amount
			}
			if edit.ValueFromScreenshot != nil {
				updates["value_from_screenshot"] = edit.ValueFromScreenshot
			}
			if edit.DisplayCurrency != nil {
				updates["display_currency"] = edit.DisplayCurrency
			}
			if edit.AvgPrice != nil {
				value := *edit.AvgPrice
				if value < 0 {
					value = 0
				}
				avgPriceUSD := value
				if baseRateToUSD > 0 {
					avgPriceUSD = value * baseRateToUSD
				}
				updates["avg_price"] = avgPriceUSD
				updates["avg_price_source"] = "user_input"
			}
			if edit.ManualValueDisplay != nil {
				value := *edit.ManualValueDisplay
				if value < 0 {
					value = 0
				}
				manualUSD := value
				if baseRateToUSD > 0 {
					manualUSD = value * baseRateToUSD
				}
				updates["manual_value_usd"] = &manualUSD
			} else if edit.ManualValueUSD != nil {
				updates["manual_value_usd"] = edit.ManualValueUSD
			}
			if len(updates) > 0 {
				if err := tx.Model(&OCRAsset{}).Where("id = ?", edit.AssetID).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
		for _, resolution := range req.Resolutions {
			if resolution.SymbolRaw == "" || resolution.AssetKey == "" {
				continue
			}
			platformCategory := symbolCategories[resolution.SymbolRaw]
			if platformCategory == "" {
				platformCategory = "unknown"
			}
			row := AmbiguityResolution{
				ID:                  newID("amb"),
				UserID:              userID,
				SymbolRaw:           resolution.SymbolRaw,
				SymbolRawNormalized: normalizeSymbol(resolution.SymbolRaw),
				PlatformCategory:    platformCategory,
				AssetType:           resolution.AssetType,
				Symbol:              resolution.Symbol,
				AssetKey:            resolution.AssetKey,
				ExchangeMIC:         nullableString(resolution.ExchangeMIC),
				CreatedAt:           time.Now().UTC(),
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		for _, override := range req.DuplicateOverrides {
			if override.ImageID == "" {
				continue
			}
			if override.Include {
				if err := tx.Model(&UploadImage{}).Where("id = ?", override.ImageID).Updates(map[string]any{
					"is_duplicate":          false,
					"duplicate_of_image_id": nil,
					"duplicate_override":    true,
				}).Error; err != nil {
					return err
				}
			}
		}

		return tx.Model(&UploadBatch{}).Where("id = ?", batch.ID).Update("status", "processing").Error
	}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "REVIEW_ERROR", "failed to save review", nil)
		return
	}

	_ = s.queue.enqueue(r.Context(), jobNormalize, batch.ID)
	s.writeJSON(w, http.StatusOK, map[string]any{"status": "processing", "poll_after_ms": 1500})
}

type ocrAssetReviewItem struct {
	AssetID             string   `json:"asset_id"`
	ImageID             string   `json:"image_id"`
	SymbolRaw           string   `json:"symbol_raw"`
	Symbol              *string  `json:"symbol"`
	AssetType           string   `json:"asset_type"`
	ExchangeMIC         *string  `json:"exchange_mic,omitempty"`
	Name                *string  `json:"name,omitempty"`
	LogoURL             *string  `json:"logo_url,omitempty"`
	Amount              float64  `json:"amount"`
	ValueFromScreenshot *float64 `json:"value_from_screenshot,omitempty"`
	ManualValueUSD      *float64 `json:"manual_value_usd,omitempty"`
	ManualValueDisplay  *float64 `json:"manual_value_display,omitempty"`
	DisplayCurrency     *string  `json:"display_currency"`
	Confidence          float64  `json:"confidence"`
	ValueUSDDraft       *float64 `json:"value_usd_priced_draft,omitempty"`
	ValueDisplayDraft   *float64 `json:"value_display_draft,omitempty"`
	PriceAsOf           string   `json:"price_as_of"`
	AvgPrice            *float64 `json:"avg_price,omitempty"`
	AvgPriceDisplay     *float64 `json:"avg_price_display,omitempty"`
	PNLPercent          *float64 `json:"pnl_percent,omitempty"`
}

func (s *Server) buildNeedsReviewResponse(ctx context.Context, batch UploadBatch) map[string]any {
	logStep := func(step string, started time.Time) {
		s.logger.Printf("needs-review step=%s batch=%s duration=%s", step, batch.ID, time.Since(started))
	}
	stepStart := time.Now()
	var images []UploadImage
	_ = s.db.DB().WithContext(ctx).Where("upload_batch_id = ?", batch.ID).Find(&images).Error

	var assets []OCRAsset
	_ = s.db.DB().WithContext(ctx).Where("upload_image_id IN (?)", imageIDs(images)).Find(&assets).Error
	logStep("load_inputs", stepStart)

	stepStart = time.Now()
	ocrAssets := make([]ocrAssetReviewItem, 0, len(assets))
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
		})
	}
	logStep("build_inputs", stepStart)
	stepStart = time.Now()
	coinList, _ := s.market.coinGeckoList(ctx)
	resolutions := s.loadAmbiguityResolutions(ctx, batch.UserID)
	resolvedAssets, _ := resolveAssets(ctx, s.market, batch.UserID, firstPlatform(images), inputAssets, coinList, resolutions, false)
	logosByAssetKey := s.resolveHoldingLogos(ctx, extractHoldings(resolvedAssets))
	logStep("resolve_assets", stepStart)
	aggregated := aggregateHoldings(extractHoldings(resolvedAssets))
	stepStart = time.Now()
	priceMap := fetchCoinGeckoPrices(ctx, s.market, aggregated)
	stockPrices := fetchMarketstackPrices(ctx, s.market, aggregated)
	oerRates := fetchOERRatesIfNeeded(ctx, s.market, aggregated, stockPrices)
	profile, _ := s.ensureUserProfile(ctx, batch.UserID)
	baseCurrency := normalizeCurrency(profile.BaseCurrency)
	language := profile.Language
	if baseCurrency != "USD" && len(oerRates) == 0 {
		if oerResp, err := s.market.openExchangeLatest(ctx); err == nil {
			oerRates = oerResp.Rates
		}
	}
	baseCurrency, rateFromUSD := resolveDisplayCurrency(baseCurrency, oerRates)
	baseRateToUSD := rateToUSDFromRateFromUSD(rateFromUSD)
	_ = s.db.DB().WithContext(ctx).Model(&UploadBatch{}).Where("id = ?", batch.ID).Updates(map[string]any{
		"base_currency":       baseCurrency,
		"base_fx_rate_to_usd": baseRateToUSD,
	}).Error
	logStep("pricing", stepStart)
	priceByAssetKey := make(map[string]float64)
	stepStart = time.Now()
	for i := range aggregated {
		applyPricing(&aggregated[i], priceMap, stockPrices, oerRates)
		applyCostBasis(&aggregated[i])
		if aggregated[i].ValuationStatus == "priced" {
			priceByAssetKey[aggregated[i].AssetKey] = aggregated[i].CurrentPrice
		}
	}
	resolvedByAssetID := make(map[string]portfolioHolding)
	for _, asset := range resolvedAssets {
		if asset.AssetID == "" {
			continue
		}
		resolvedByAssetID[asset.AssetID] = asset.Holding
	}
	nameBySymbol := map[string]*string{}
	if s.market != nil {
		for _, asset := range resolvedAssets {
			holding := asset.Holding
			displaySymbol := hkDisplaySymbol(holding.Symbol)
			if displaySymbol == "" || !isHongKongStockSymbol(displaySymbol, holding.ExchangeMIC) {
				continue
			}
			if _, ok := nameBySymbol[displaySymbol]; ok {
				continue
			}
			if localized, ok := hkNameForLanguage(displaySymbol, language); ok {
				nameBySymbol[displaySymbol] = &localized
				continue
			}
			ticker, err := s.market.marketstackTicker(ctx, displaySymbol)
			if err != nil {
				nameBySymbol[displaySymbol] = nil
				continue
			}
			name := strings.TrimSpace(ticker.Name)
			if name == "" {
				nameBySymbol[displaySymbol] = nil
				continue
			}
			nameBySymbol[displaySymbol] = &name
		}
	}
	logStep("apply_pricing", stepStart)
	stepStart = time.Now()
	ambiguities, _ := s.ensureAmbiguities(ctx, batch, images, assets, language)
	if ambiguities == nil {
		ambiguities = []ambiguityView{}
	}
	logStep("ambiguities", stepStart)
	stepStart = time.Now()
	for _, asset := range assets {
		resolved := resolvedByAssetID[asset.ID]
		symbol := asset.Symbol
		assetType := asset.AssetType
		exchangeMIC := ""
		var logoURL *string
		if resolved.AssetKey != "" {
			assetType = resolved.AssetType
			if resolved.Symbol != "" {
				resolvedSymbol := resolved.Symbol
				symbol = &resolvedSymbol
			}
			exchangeMIC = resolved.ExchangeMIC
			if logo := logosByAssetKey[resolved.AssetKey]; logo != "" {
				logoCopy := logo
				logoURL = &logoCopy
			}
		}
		displaySymbol := ""
		if symbol != nil {
			displaySymbol = hkDisplaySymbol(*symbol)
		}
		if displaySymbol == "" && asset.SymbolRaw != "" {
			displaySymbol = hkDisplaySymbol(asset.SymbolRaw)
		}
		if displaySymbol != "" && symbol != nil && *symbol != displaySymbol {
			normalized := displaySymbol
			symbol = &normalized
		}
		if displaySymbol != "" && symbol == nil {
			normalized := displaySymbol
			symbol = &normalized
		}
		var name *string
		if displaySymbol != "" && isHongKongStockSymbol(displaySymbol, exchangeMIC) {
			name = nameBySymbol[displaySymbol]
		}
		var exchangeMICPtr *string
		if exchangeMIC != "" {
			exchangeMICPtr = &exchangeMIC
		}
		displayCurrency := asset.DisplayCurrency
		if displayCurrency == nil || strings.TrimSpace(*displayCurrency) == "" {
			displayCurrency = &baseCurrency
		}
		draftValue := asset.ManualValueUSD
		if draftValue == nil {
			if price, ok := priceByAssetKey[resolved.AssetKey]; ok && price > 0 {
				value := asset.Amount * price
				draftValue = &value
			} else {
				draftValue = draftValueFromScreenshot(asset.ValueFromScreenshot, displayCurrency, oerRates)
			}
		}
		var avgPriceUSD *float64
		if asset.AvgPrice != nil {
			if asset.AvgPriceSource != nil && strings.TrimSpace(*asset.AvgPriceSource) == "user_input" {
				avgPriceUSD = asset.AvgPrice
			} else if converted, ok := convertDisplayPriceToUSD(*asset.AvgPrice, displayCurrency, oerRates); ok {
				avgPriceUSD = &converted
			}
		}
		var avgPriceDisplay *float64
		if avgPriceUSD != nil {
			value := *avgPriceUSD * rateFromUSD
			avgPriceDisplay = &value
		}
		var draftValueDisplay *float64
		if draftValue != nil {
			value := *draftValue * rateFromUSD
			draftValueDisplay = &value
		}
		var manualValueDisplay *float64
		if asset.ManualValueUSD != nil {
			value := *asset.ManualValueUSD * rateFromUSD
			manualValueDisplay = &value
		}
		updates := map[string]any{
			"value_usd_priced": draftValue,
		}
		if asset.DisplayCurrency == nil || strings.TrimSpace(*asset.DisplayCurrency) == "" {
			updates["display_currency"] = baseCurrency
		}
		_ = s.db.DB().WithContext(ctx).Model(&OCRAsset{}).Where("id = ?", asset.ID).Updates(updates).Error
		ocrAssets = append(ocrAssets, ocrAssetReviewItem{
			AssetID:             asset.ID,
			ImageID:             asset.UploadImageID,
			SymbolRaw:           asset.SymbolRaw,
			Symbol:              symbol,
			AssetType:           assetType,
			ExchangeMIC:         exchangeMICPtr,
			Name:                name,
			LogoURL:             logoURL,
			Amount:              asset.Amount,
			ValueFromScreenshot: asset.ValueFromScreenshot,
			ManualValueUSD:      asset.ManualValueUSD,
			ManualValueDisplay:  manualValueDisplay,
			DisplayCurrency:     displayCurrency,
			Confidence:          asset.Confidence,
			ValueUSDDraft:       draftValue,
			ValueDisplayDraft:   draftValueDisplay,
			PriceAsOf:           priceAsOf().Format(time.RFC3339),
			AvgPrice:            avgPriceUSD,
			AvgPriceDisplay:     avgPriceDisplay,
			PNLPercent:          asset.PNLPercent,
		})
	}
	sort.SliceStable(ocrAssets, func(i, j int) bool {
		left := ocrAssets[i].ValueUSDDraft
		right := ocrAssets[j].ValueUSDDraft
		if left == nil && right == nil {
			if ocrAssets[i].SymbolRaw == ocrAssets[j].SymbolRaw {
				return ocrAssets[i].AssetID < ocrAssets[j].AssetID
			}
			return ocrAssets[i].SymbolRaw < ocrAssets[j].SymbolRaw
		}
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		if *left == *right {
			if ocrAssets[i].SymbolRaw == ocrAssets[j].SymbolRaw {
				return ocrAssets[i].AssetID < ocrAssets[j].AssetID
			}
			return ocrAssets[i].SymbolRaw < ocrAssets[j].SymbolRaw
		}
		return *left > *right
	})
	logStep("build_payload", stepStart)

	imageResponses := make([]map[string]any, 0, len(images))
	for _, image := range images {
		imageResponses = append(imageResponses, map[string]any{
			"image_id":              image.ID,
			"status":                image.Status,
			"platform_guess":        image.PlatformGuess,
			"error_reason":          image.ErrorReason,
			"is_duplicate":          image.IsDuplicate,
			"duplicate_of_image_id": image.DuplicateOfImage,
			"warnings":              []string(image.Warnings),
		})
	}

	success, ignored, unsupported := countOCRStatuses(toOCRImages(images))
	return map[string]any{
		"upload_batch_id":     batch.ID,
		"status":              "needs_review",
		"base_currency":       baseCurrency,
		"base_fx_rate_to_usd": baseRateToUSD,
		"images":              imageResponses,
		"ambiguities":         ambiguities,
		"ocr_assets":          ocrAssets,
		"summary": map[string]any{
			"success_images":     success,
			"ignored_images":     ignored,
			"unsupported_images": unsupported,
		},
	}
}

func (s *Server) findCalculationForBatch(ctx context.Context, batchID string) string {
	var snapshot PortfolioSnapshot
	if err := s.db.DB().WithContext(ctx).Where("source_upload_batch_id = ?", batchID).First(&snapshot).Error; err != nil {
		return ""
	}
	var calc Calculation
	if err := s.db.DB().WithContext(ctx).Where("portfolio_snapshot_id = ?", snapshot.ID).First(&calc).Error; err != nil {
		return ""
	}
	return calc.ID
}

func (s *Server) findSnapshotForBatch(ctx context.Context, batchID string, snapshotType string) string {
	var snapshot PortfolioSnapshot
	query := s.db.DB().WithContext(ctx).Where("source_upload_batch_id = ?", batchID)
	if snapshotType != "" {
		query = query.Where("snapshot_type = ?", snapshotType)
	}
	if err := query.Order("created_at desc").First(&snapshot).Error; err != nil {
		return ""
	}
	return snapshot.ID
}

func (s *Server) findTransactionsForSnapshot(ctx context.Context, snapshotID string) []string {
	if snapshotID == "" {
		return nil
	}
	var txs []PortfolioTransaction
	if err := s.db.DB().WithContext(ctx).Where("snapshot_id_after = ?", snapshotID).Find(&txs).Error; err != nil {
		return nil
	}
	ids := make([]string, 0, len(txs))
	for _, tx := range txs {
		ids = append(ids, tx.ID)
	}
	return ids
}

func buildStorageKey(userID, batchID, imageID, filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}
	return "uploads/" + userID + "/" + batchID + "/" + imageID + ext
}

func imageMimeType(key string) string {
	switch strings.ToLower(filepath.Ext(key)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func firstPlatform(images []UploadImage) string {
	for _, image := range images {
		if image.PlatformGuess != "" {
			return image.PlatformGuess
		}
	}
	return ""
}

func toOCRImages(images []UploadImage) []ocrImage {
	result := make([]ocrImage, 0, len(images))
	for _, image := range images {
		result = append(result, ocrImage{
			ImageID: image.ID,
			Status:  image.Status,
		})
	}
	return result
}

func (s *Server) computeClientChecksum(ctx context.Context, imageIDs []string, imageSet map[string]UploadImage) (string, error) {
	lines := make([]string, 0, len(imageIDs))
	for _, id := range imageIDs {
		image, ok := imageSet[id]
		if !ok {
			return "", fmt.Errorf("missing image %s", id)
		}
		bytes, err := s.storage.getObjectBytes(ctx, image.StorageKey)
		if err != nil {
			return "", err
		}
		sum := sha256.Sum256(bytes)
		line := fmt.Sprintf("%s:%s:%d", id, hex.EncodeToString(sum[:]), len(bytes))
		lines = append(lines, line)
	}
	manifest := strings.Join(lines, "\n")
	hash := sha256.Sum256([]byte(manifest))
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}
