package app

import "context"

func resolveHoldings(ctx context.Context, market *marketClient, userID, platformGuess string, assets []ocrAssetInput, coinList []coinGeckoCoinListEntry, resolutions map[ambiguityKey]AmbiguityResolution) ([]portfolioHolding, map[string]float64, error) {
	resolved, err := resolveAssets(ctx, market, userID, platformGuess, assets, coinList, resolutions, false)
	if err != nil {
		return nil, nil, err
	}
	holdings := aggregateHoldings(extractHoldings(resolved))
	priceMap := fetchCoinGeckoPrices(ctx, market, holdings)
	stockPrices := fetchMarketstackPrices(ctx, market, holdings)
	oerRates := fetchOERRatesIfNeeded(ctx, market, holdings, stockPrices)
	priceByAssetKey := make(map[string]float64)

	for i := range holdings {
		applyPricing(&holdings[i], priceMap, stockPrices, oerRates)
		applyCostBasis(&holdings[i])
		if holdings[i].ValuationStatus == "priced" {
			priceByAssetKey[holdings[i].AssetKey] = holdings[i].CurrentPrice
		}
	}

	return holdings, priceByAssetKey, nil
}

func resolveAssets(ctx context.Context, market *marketClient, userID, platformGuess string, assets []ocrAssetInput, coinList []coinGeckoCoinListEntry, resolutions map[ambiguityKey]AmbiguityResolution, strict bool) ([]resolvedAsset, error) {
	stablecoins := stablecoinSet()
	symbolToIDs := buildSymbolToIDs(coinList)

	fxSymbols, err := market.openExchangeCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	stockCache := make(map[string]marketstackTickerResponse)
	resolver := assetResolver{
		market:        market,
		userID:        userID,
		platformGuess: platformGuess,
		symbolToIDs:   symbolToIDs,
		fxSymbols:     fxSymbols,
		stockCache:    stockCache,
		stablecoins:   stablecoins,
		resolutions:   resolutions,
		strict:        strict,
	}
	holdings := make([]resolvedAsset, 0, len(assets))
	for _, asset := range assets {
		resolved, ok := resolver.resolveAsset(ctx, asset)
		if !ok {
			continue
		}
		holdings = append(holdings, resolved)
	}

	return holdings, nil
}
