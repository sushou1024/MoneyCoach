package app

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ambiguityCandidate struct {
	AssetType   string `json:"asset_type"`
	Symbol      string `json:"symbol"`
	Name        string `json:"name,omitempty"`
	ExchangeMIC string `json:"exchange_mic,omitempty"`
	AssetKey    string `json:"asset_key,omitempty"`
	CoinGeckoID string `json:"coingecko_id,omitempty"`
}

type ambiguityView struct {
	ImageID    string               `json:"image_id"`
	SymbolRaw  string               `json:"symbol_raw"`
	Candidates []ambiguityCandidate `json:"candidates"`
}

func (s *Server) loadAmbiguityResolutions(ctx context.Context, userID string) map[ambiguityKey]AmbiguityResolution {
	if strings.TrimSpace(userID) == "" {
		return nil
	}
	var rows []AmbiguityResolution
	if err := s.db.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil
	}
	out := make(map[ambiguityKey]AmbiguityResolution, len(rows))
	for _, row := range rows {
		key := ambiguityKey{SymbolRawNormalized: row.SymbolRawNormalized, PlatformCategory: row.PlatformCategory}
		out[key] = row
	}
	return out
}

func (s *Server) ensureAmbiguities(ctx context.Context, batch UploadBatch, images []UploadImage, assets []OCRAsset, language string) ([]ambiguityView, error) {
	var existing []OCRAmbiguity
	if err := s.db.DB().WithContext(ctx).Where("upload_batch_id = ?", batch.ID).Find(&existing).Error; err == nil && len(existing) > 0 {
		return s.mapAmbiguityViews(existing), nil
	}

	coinList, err := s.market.coinGeckoList(ctx)
	if err != nil {
		return nil, err
	}
	fxSymbols, err := s.market.openExchangeCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	symbolToCoins := make(map[string][]coinGeckoCoinListEntry)
	symbolToIDs := make(map[string][]string)
	for _, coin := range coinList {
		if coin.Symbol == "" {
			continue
		}
		symbol := strings.ToUpper(coin.Symbol)
		symbolToCoins[symbol] = append(symbolToCoins[symbol], coin)
		symbolToIDs[symbol] = append(symbolToIDs[symbol], coin.ID)
	}

	imagePlatform := make(map[string]string, len(images))
	for _, image := range images {
		imagePlatform[image.ID] = platformGuessToCategory(image.PlatformGuess)
	}
	resolutions := s.loadAmbiguityResolutions(ctx, batch.UserID)
	stockCache := make(map[string]marketstackTickerResponse)
	created := make([]OCRAmbiguity, 0)

	for _, asset := range assets {
		symbolRaw := strings.TrimSpace(asset.SymbolRaw)
		if symbolRaw == "" {
			continue
		}
		platformCategory := imagePlatform[asset.UploadImageID]
		if resolutions != nil {
			if _, ok := resolutions[newAmbiguityKey(symbolRaw, platformCategory)]; ok {
				continue
			}
		}

		symbol := ""
		aliasUsed := false
		if asset.Symbol != nil {
			symbol = strings.TrimSpace(*asset.Symbol)
		}
		if symbol == "" {
			if alias, ok := aliasSymbol(symbolRaw); ok {
				symbol = alias
				aliasUsed = true
			}
		}
		normalized := normalizeSymbol(symbol)
		if normalized == "" {
			normalized = normalizeSymbol(symbolRaw)
		}
		if defaultAssetTypeForSymbol(normalized) != "" {
			continue
		}

		candidates := make([]ambiguityCandidate, 0)
		if entries := symbolToCoins[normalized]; len(entries) > 0 {
			for _, entry := range entries {
				candidates = append(candidates, ambiguityCandidate{
					AssetType:   "crypto",
					Symbol:      normalized,
					Name:        entry.Name,
					AssetKey:    "crypto:cg:" + entry.ID,
					CoinGeckoID: entry.ID,
				})
			}
		}
		if normalized != "" {
			if _, ok := stockCache[normalized]; !ok {
				if ticker, err := s.market.marketstackTicker(ctx, normalized); err == nil {
					stockCache[normalized] = ticker
				}
			}
			if ticker, ok := stockCache[normalized]; ok && ticker.StockExchange.MIC != "" {
				name := strings.TrimSpace(ticker.Name)
				if localized, ok := hkNameForLanguage(ticker.Symbol, language); ok {
					name = localized
				}
				candidates = append(candidates, ambiguityCandidate{
					AssetType:   "stock",
					Symbol:      normalized,
					Name:        name,
					ExchangeMIC: ticker.StockExchange.MIC,
					AssetKey:    stockAssetKey(ticker.StockExchange.MIC, normalized),
				})
			}
		}
		if name, ok := fxSymbols[normalized]; ok {
			candidates = append(candidates, ambiguityCandidate{
				AssetType: "forex",
				Symbol:    normalized,
				Name:      name,
				AssetKey:  "forex:fx:" + normalized,
			})
		}

		needsReview := aliasUsed || ambiguousCandidateTypes(candidates)
		if !needsReview && len(symbolToCoins[normalized]) > 1 {
			if resolveCoinGeckoID(ctx, s.market, normalized, symbolToIDs) == "" {
				needsReview = true
			}
		}
		if !needsReview {
			continue
		}

		encoded, _ := json.Marshal(candidates)
		created = append(created, OCRAmbiguity{
			ID:            newID("amb"),
			UploadBatchID: batch.ID,
			UploadImageID: asset.UploadImageID,
			SymbolRaw:     symbolRaw,
			Candidates:    datatypes.JSON(encoded),
		})
	}

	if len(created) == 0 {
		return nil, nil
	}

	if err := s.db.withTx(ctx, func(tx *gorm.DB) error {
		for _, row := range created {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.mapAmbiguityViews(created), nil
}

func (s *Server) mapAmbiguityViews(rows []OCRAmbiguity) []ambiguityView {
	items := make([]ambiguityView, 0, len(rows))
	for _, row := range rows {
		var candidates []ambiguityCandidate
		_ = json.Unmarshal(row.Candidates, &candidates)
		items = append(items, ambiguityView{
			ImageID:    row.UploadImageID,
			SymbolRaw:  row.SymbolRaw,
			Candidates: candidates,
		})
	}
	return items
}

func ambiguousCandidateTypes(candidates []ambiguityCandidate) bool {
	if len(candidates) <= 1 {
		return false
	}
	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		if candidate.AssetType == "" {
			continue
		}
		seen[candidate.AssetType] = struct{}{}
	}
	return len(seen) > 1
}
