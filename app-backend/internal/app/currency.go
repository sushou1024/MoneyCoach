package app

import "strings"

func normalizeCurrency(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "USD"
	}
	return value
}

func currencyRateToUSD(currency string, rates map[string]float64) (float64, bool) {
	currency = normalizeCurrency(currency)
	if currency == "USD" || stablecoinSet()[currency] {
		return 1, true
	}
	if rate, ok := rates[currency]; ok && rate > 0 {
		return 1 / rate, true
	}
	return 0, false
}

func currencyRateFromUSD(currency string, rates map[string]float64) (float64, bool) {
	currency = normalizeCurrency(currency)
	if currency == "USD" || stablecoinSet()[currency] {
		return 1, true
	}
	if rate, ok := rates[currency]; ok && rate > 0 {
		return rate, true
	}
	return 0, false
}

func convertUSDToCurrency(usd float64, currency string, rates map[string]float64) (float64, float64, bool) {
	rateFromUSD, ok := currencyRateFromUSD(currency, rates)
	if !ok {
		return usd, 0, false
	}
	if rateFromUSD == 0 {
		return usd, 0, false
	}
	return usd * rateFromUSD, 1 / rateFromUSD, true
}

func convertCurrencyToUSD(value float64, currency string, rates map[string]float64) (float64, float64, bool) {
	rateToUSD, ok := currencyRateToUSD(currency, rates)
	if !ok {
		return value, 0, false
	}
	return value * rateToUSD, rateToUSD, true
}

func rateToUSDFromRateFromUSD(rateFromUSD float64) float64 {
	if rateFromUSD <= 0 {
		return 1
	}
	return 1 / rateFromUSD
}
