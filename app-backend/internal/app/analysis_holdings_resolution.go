package app

import (
	"context"
	"sort"
)

func applyBalanceType(platformCategory, symbolRaw, normalized, assetType string, stablecoins map[string]bool) (string, string) {
	balanceType := "unknown"
	if stablecoins[normalized] {
		balanceType = "stablecoin"
	}
	if platformCategory != "unknown" {
		if containsCashLabel(symbolRaw) {
			balanceType = "fiat_cash"
		}
		if (platformCategory == "crypto_exchange" || platformCategory == "wallet") && normalized == "USD" {
			assetType = "forex"
			balanceType = "fiat_cash"
		}
	}
	return assetType, balanceType
}

func resolveCoinGeckoID(ctx context.Context, market *marketClient, symbol string, symbolToIDs map[string][]string) string {
	return resolveCoinGeckoIDWithFallback(ctx, market, symbol, symbolToIDs, false)
}

func resolveCoinGeckoIDWithFallback(ctx context.Context, market *marketClient, symbol string, symbolToIDs map[string][]string, force bool) string {
	ids := symbolToIDs[symbol]
	if len(ids) == 0 {
		return ""
	}
	if len(ids) == 1 {
		return ids[0]
	}
	items, err := market.coinGeckoMarkets(ctx, ids)
	if err != nil || len(items) == 0 {
		if force {
			idsCopy := append([]string(nil), ids...)
			sort.Strings(idsCopy)
			return idsCopy[0]
		}
		return ""
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].MarketCap == items[j].MarketCap {
			return items[i].ID < items[j].ID
		}
		return items[i].MarketCap > items[j].MarketCap
	})
	if items[0].MarketCap == 0 {
		if force {
			return items[0].ID
		}
		return ""
	}
	if !force && len(items) > 1 && items[0].MarketCap == items[1].MarketCap {
		return ""
	}
	return items[0].ID
}
