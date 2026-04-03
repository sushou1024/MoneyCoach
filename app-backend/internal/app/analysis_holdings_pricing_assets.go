package app

import "strings"

func applyUSDValuation(holding *portfolioHolding, price float64, source string) {
	holding.CurrentPrice = price
	holding.ValueUSD = holding.Amount * price
	holding.ValuationStatus = "priced"
	holding.PricingSource = source
	if holding.QuoteCurrency == "" {
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     "USD",
			CurrentPriceQuote: price,
			FXRateToUSD:       1,
		})
	}
}

func applyStablecoinPricing(holding *portfolioHolding) bool {
	if holding.BalanceType != "stablecoin" {
		return false
	}
	applyUSDValuation(holding, 1, "COINGECKO")
	currency := normalizeCurrency(holding.Symbol)
	if currency == "" {
		currency = "USD"
	}
	applyQuoteMetadata(holding, assetQuoteMetadata{
		QuoteCurrency:     currency,
		CurrentPriceQuote: 1,
		FXRateToUSD:       1,
	})
	return true
}

func applyCryptoPricing(holding *portfolioHolding, priceMap map[string]coinGeckoSimplePrice) bool {
	if holding.AssetType != "crypto" || holding.CoinGeckoID == "" {
		return false
	}
	price, ok := priceMap[holding.CoinGeckoID]
	if !ok || price.USD <= 0 {
		return false
	}
	applyUSDValuation(holding, price.USD, "COINGECKO")
	applyQuoteMetadata(holding, assetQuoteMetadata{
		QuoteCurrency:     "USD",
		CurrentPriceQuote: price.USD,
		FXRateToUSD:       1,
	})
	return true
}

func applyStockPricing(holding *portfolioHolding, stockPrices map[string]marketstackPriceQuote, oerRates map[string]float64) bool {
	if holding.AssetType != "stock" || holding.Symbol == "" {
		return false
	}
	quote, ok := stockPrices[holding.Symbol]
	if !ok || quote.Close <= 0 {
		return false
	}
	price := quote.Close
	currency := marketstackPriceCurrency(quote.PriceCurrency, quote.Symbol, quote.Exchange)
	if currency == "" {
		currency = "USD"
	}
	fxRateToUSD := 1.0
	if currency != "USD" {
		rate, ok := currencyRateToUSD(currency, oerRates)
		if !ok || rate <= 0 {
			return false
		}
		fxRateToUSD = rate
	}
	usdPrice := price * fxRateToUSD
	applyUSDValuation(holding, usdPrice, "MARKETSTACK")
	applyQuoteMetadata(holding, assetQuoteMetadata{
		QuoteCurrency:     currency,
		CurrentPriceQuote: price,
		FXRateToUSD:       fxRateToUSD,
	})
	return true
}

func applyForexPricing(holding *portfolioHolding, oerRates map[string]float64) bool {
	if holding.AssetType != "forex" || holding.Symbol == "" {
		return false
	}
	symbol := strings.ToUpper(holding.Symbol)
	if symbol == "USD" {
		applyUSDValuation(holding, 1, "OER")
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     "USD",
			CurrentPriceQuote: 1,
			FXRateToUSD:       1,
		})
		return true
	}
	if rate, ok := oerRates[symbol]; ok && rate > 0 {
		usdRate := 1 / rate
		applyUSDValuation(holding, usdRate, "OER")
		applyQuoteMetadata(holding, assetQuoteMetadata{
			QuoteCurrency:     symbol,
			CurrentPriceQuote: 1,
			FXRateToUSD:       usdRate,
		})
		return true
	}
	return false
}
