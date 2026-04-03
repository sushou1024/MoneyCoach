package app

import (
	"context"
	"strings"
)

func marketAlphaUniverse(ctx context.Context, market *marketClient, holdings []portfolioHolding) []portfolioHolding {
	candidates := make([]portfolioHolding, 0)
	seen := make(map[string]struct{})

	for _, holding := range holdings {
		if !eligibleMarketAlphaHolding(holding) {
			continue
		}
		if holding.AssetKey == "" {
			continue
		}
		if _, ok := seen[holding.AssetKey]; ok {
			continue
		}
		seen[holding.AssetKey] = struct{}{}
		candidates = append(candidates, holding)
	}

	markets, err := market.coinGeckoTopMarkets(ctx, 50)
	if err == nil {
		stablecoins := stablecoinSet()
		for _, item := range markets {
			symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
			if symbol == "" || stablecoins[symbol] {
				continue
			}
			assetKey := "crypto:cg:" + item.ID
			if _, ok := seen[assetKey]; ok {
				continue
			}
			seen[assetKey] = struct{}{}
			candidates = append(candidates, portfolioHolding{
				Symbol:          symbol,
				AssetType:       "crypto",
				AssetKey:        assetKey,
				ValuationStatus: "priced",
			})
		}
	}

	for _, symbol := range defaultStockWatchlist {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		ticker, err := market.marketstackTicker(ctx, symbol)
		if err != nil || ticker.StockExchange.MIC == "" {
			continue
		}
		assetKey := stockAssetKey(ticker.StockExchange.MIC, ticker.Symbol)
		if _, ok := seen[assetKey]; ok {
			continue
		}
		seen[assetKey] = struct{}{}
		candidates = append(candidates, portfolioHolding{
			Symbol:          ticker.Symbol,
			AssetType:       "stock",
			AssetKey:        assetKey,
			ExchangeMIC:     ticker.StockExchange.MIC,
			ValuationStatus: "priced",
		})
	}

	return candidates
}

func eligibleMarketAlphaHolding(holding portfolioHolding) bool {
	if holding.ValuationStatus != "priced" {
		return false
	}
	if strings.HasPrefix(holding.AssetKey, "manual:") {
		return false
	}
	if holding.AssetType != "crypto" && holding.AssetType != "stock" {
		return false
	}
	if holding.BalanceType == "stablecoin" || holding.BalanceType == "fiat_cash" {
		return false
	}
	if stablecoinSet()[strings.ToUpper(holding.Symbol)] {
		return false
	}
	return true
}
