package app

import (
	"context"
	"strings"
)

func fetchCoinGeckoPrices(ctx context.Context, market *marketClient, holdings []portfolioHolding) map[string]coinGeckoSimplePrice {
	ids := make([]string, 0)
	for _, holding := range holdings {
		if holding.CoinGeckoID == "" {
			continue
		}
		ids = append(ids, holding.CoinGeckoID)
	}
	prices, err := market.coinGeckoSimplePrice(ctx, ids)
	if err != nil {
		return map[string]coinGeckoSimplePrice{}
	}
	return prices
}

func fetchMarketstackPrices(ctx context.Context, market *marketClient, holdings []portfolioHolding) map[string]marketstackPriceQuote {
	symbols := make([]string, 0)
	for _, holding := range holdings {
		if holding.AssetType != "stock" || holding.Symbol == "" {
			continue
		}
		symbols = append(symbols, holding.Symbol)
	}
	if len(symbols) == 0 {
		return map[string]marketstackPriceQuote{}
	}
	resp, err := market.marketstackEOD(ctx, symbols, true)
	if err != nil {
		return map[string]marketstackPriceQuote{}
	}
	prices := make(map[string]marketstackPriceQuote)
	for _, item := range resp.Data {
		if item.Symbol == "" {
			continue
		}
		currency := marketstackPriceCurrency(item.PriceCurrency, item.Symbol, item.Exchange)
		prices[item.Symbol] = marketstackPriceQuote{
			Close:         item.Close,
			PriceCurrency: currency,
			Symbol:        item.Symbol,
			Exchange:      item.Exchange,
		}
	}
	return prices
}

func fetchOERRatesIfNeeded(ctx context.Context, market *marketClient, holdings []portfolioHolding, stockPrices map[string]marketstackPriceQuote) map[string]float64 {
	needs := false
	for _, holding := range holdings {
		if holding.DisplayCurrency != nil && *holding.DisplayCurrency != "" && strings.ToUpper(*holding.DisplayCurrency) != "USD" {
			needs = true
			break
		}
	}
	if !needs {
		for _, price := range stockPrices {
			if price.PriceCurrency != "" && price.PriceCurrency != "USD" {
				needs = true
				break
			}
		}
	}
	if !needs {
		return map[string]float64{}
	}
	resp, err := market.openExchangeLatest(ctx)
	if err != nil {
		return map[string]float64{}
	}
	return resp.Rates
}
