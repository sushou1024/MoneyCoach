package app

import (
	"context"
	"strings"
)

type assetResolver struct {
	market        *marketClient
	userID        string
	platformGuess string
	symbolToIDs   map[string][]string
	fxSymbols     map[string]string
	stockCache    map[string]marketstackTickerResponse
	stablecoins   map[string]bool
	resolutions   map[ambiguityKey]AmbiguityResolution
	strict        bool
}

func (r assetResolver) resolveAsset(ctx context.Context, asset ocrAssetInput) (resolvedAsset, bool) {
	if asset.Amount == 0 {
		return resolvedAsset{}, false
	}

	symbolRaw := strings.TrimSpace(asset.SymbolRaw)
	assetPlatform, platformCategory := resolvePlatformCategory(r.platformGuess, asset.PlatformGuess)
	symbol, aliasUsed := resolvedSymbol(asset, symbolRaw)
	normalized := normalizeSymbol(symbol)
	if normalized == "" {
		normalized = normalizeSymbol(symbolRaw)
	}
	defaultType := defaultAssetTypeForSymbol(normalized)

	var resolution AmbiguityResolution
	resolutionFound := false
	if r.resolutions != nil {
		if value, ok := r.resolutions[newAmbiguityKey(symbolRaw, platformCategory)]; ok {
			resolution = value
			resolutionFound = true
		}
	}

	assetType := normalizeAssetType(asset.AssetType)
	assetType, balanceType := applyBalanceType(platformCategory, symbolRaw, normalized, assetType, r.stablecoins)
	if defaultType != "" {
		assetType = defaultType
	}

	if normalized != "" {
		cachedMarketstackTicker(ctx, r.market, r.stockCache, normalized)
	}
	cryptoCandidate, stockCandidate, forexCandidate := symbolCandidates(normalized, r.symbolToIDs, r.stockCache, r.fxSymbols)

	candidateTypes := 0
	if cryptoCandidate {
		candidateTypes++
	}
	if stockCandidate {
		candidateTypes++
	}
	if forexCandidate {
		candidateTypes++
	}

	needsResolution := aliasUsed || candidateTypes > 1
	resolvedCoinID := ""
	if cryptoCandidate && len(r.symbolToIDs[normalized]) > 1 {
		resolvedCoinID = resolveCoinGeckoIDWithFallback(ctx, r.market, normalized, r.symbolToIDs, defaultType == "crypto")
		if resolvedCoinID == "" && defaultType != "crypto" {
			needsResolution = true
		}
	}
	if defaultType != "" {
		needsResolution = false
	}

	assetType = resolveAssetType(assetType, cryptoCandidate, forexCandidate, stockCandidate)

	holding := portfolioHolding{
		SymbolRaw:           symbolRaw,
		Symbol:              normalized,
		AssetType:           assetType,
		Amount:              asset.Amount,
		ValueFromScreenshot: asset.ValueFromScreenshot,
		ManualValueUSD:      asset.ManualValueUSD,
		DisplayCurrency:     asset.DisplayCurrency,
		AvgPrice:            asset.AvgPrice,
		AvgPriceSource:      strings.TrimSpace(asset.AvgPriceSource),
		PNLPercent:          asset.PNLPercent,
		BalanceType:         balanceType,
		CostBasisStatus:     "unknown",
	}

	if resolutionFound {
		holding.AssetType = resolution.AssetType
		holding.Symbol = resolution.Symbol
		holding.AssetKey = resolution.AssetKey
		holding.CoinGeckoID = derefString(resolution.CoinGeckoID)
		holding.ExchangeMIC = derefString(resolution.ExchangeMIC)
	} else {
		if r.strict && needsResolution {
			holding.AssetKey = manualAssetKey(r.userID, normalized, assetPlatform)
		} else {
			switch holding.AssetType {
			case "crypto":
				coinID := resolvedCoinID
				if coinID == "" {
					coinID = resolveCoinGeckoIDWithFallback(ctx, r.market, normalized, r.symbolToIDs, defaultType == "crypto")
				}
				if coinID != "" {
					holding.CoinGeckoID = coinID
					holding.AssetKey = "crypto:cg:" + coinID
				} else {
					holding.AssetKey = manualAssetKey(r.userID, normalized, assetPlatform)
				}
			case "stock":
				if ticker, ok := cachedMarketstackTicker(ctx, r.market, r.stockCache, normalized); ok {
					holding.ExchangeMIC = ticker.StockExchange.MIC
					holding.AssetKey = stockAssetKey(ticker.StockExchange.MIC, normalized)
				}
				if holding.AssetKey == "" {
					holding.AssetKey = manualAssetKey(r.userID, normalized, assetPlatform)
				}
			case "forex":
				if normalized != "" {
					holding.AssetKey = "forex:fx:" + normalized
				}
				if holding.AssetKey == "" {
					holding.AssetKey = manualAssetKey(r.userID, normalized, assetPlatform)
				}
			default:
				holding.AssetKey = manualAssetKey(r.userID, normalized, assetPlatform)
			}
		}
	}

	if holding.Symbol == "" {
		holding.Symbol = normalized
	}
	if holding.AvgPrice != nil && holding.AvgPriceSource == "" {
		holding.AvgPriceSource = "provided"
	}

	return resolvedAsset{
		AssetID: asset.AssetID,
		ImageID: asset.ImageID,
		Holding: holding,
	}, true
}
