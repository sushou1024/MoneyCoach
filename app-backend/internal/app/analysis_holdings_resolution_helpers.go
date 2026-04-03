package app

import (
	"context"
	"strings"
)

func resolvedSymbol(asset ocrAssetInput, symbolRaw string) (symbol string, aliasUsed bool) {
	symbol = ""
	aliasUsed = false
	if asset.Symbol != nil {
		symbol = strings.TrimSpace(*asset.Symbol)
	}
	if symbol == "" {
		if alias, ok := aliasSymbol(symbolRaw); ok {
			symbol = alias
			aliasUsed = true
		}
	}
	return symbol, aliasUsed
}

func buildSymbolToIDs(coinList []coinGeckoCoinListEntry) map[string][]string {
	symbolToIDs := make(map[string][]string)
	for _, coin := range coinList {
		if coin.Symbol == "" {
			continue
		}
		symbol := strings.ToUpper(coin.Symbol)
		symbolToIDs[symbol] = append(symbolToIDs[symbol], coin.ID)
	}
	return symbolToIDs
}

func cachedMarketstackTicker(ctx context.Context, market *marketClient, stockCache map[string]marketstackTickerResponse, symbol string) (marketstackTickerResponse, bool) {
	if symbol == "" {
		return marketstackTickerResponse{}, false
	}
	if ticker, ok := stockCache[symbol]; ok {
		return ticker, ticker.StockExchange.MIC != ""
	}
	ticker, err := market.marketstackTicker(ctx, symbol)
	if err != nil {
		return marketstackTickerResponse{}, false
	}
	stockCache[symbol] = ticker
	return ticker, ticker.StockExchange.MIC != ""
}

func resolvePlatformCategory(platformGuess, assetPlatform string) (string, string) {
	defaultPlatform := strings.TrimSpace(platformGuess)
	defaultCategory := platformGuessToCategory(defaultPlatform)
	assetPlatform = strings.TrimSpace(assetPlatform)
	platformCategory := defaultCategory
	if assetPlatform != "" {
		platformCategory = platformGuessToCategory(assetPlatform)
	}
	if assetPlatform == "" {
		assetPlatform = defaultPlatform
	}
	return assetPlatform, platformCategory
}

func symbolCandidates(normalized string, symbolToIDs map[string][]string, stockCache map[string]marketstackTickerResponse, fxSymbols map[string]string) (bool, bool, bool) {
	cryptoCandidate := len(symbolToIDs[normalized]) > 0
	stockCandidate := false
	if normalized != "" {
		if ticker, ok := stockCache[normalized]; ok && ticker.StockExchange.MIC != "" {
			stockCandidate = true
		}
	}
	forexCandidate := normalized != "" && fxSymbols[normalized] != ""
	return cryptoCandidate, stockCandidate, forexCandidate
}

func resolveAssetType(assetType string, cryptoCandidate, forexCandidate, stockCandidate bool) string {
	if assetType != "" {
		return assetType
	}
	if cryptoCandidate {
		return "crypto"
	}
	if forexCandidate {
		return "forex"
	}
	if stockCandidate {
		return "stock"
	}
	return ""
}
