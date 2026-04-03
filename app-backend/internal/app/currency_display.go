package app

import "strings"

func normalizeDisplayCurrency(display *string) (string, bool) {
	if display == nil {
		return "", false
	}
	raw := strings.ToUpper(strings.TrimSpace(*display))
	if raw == "" {
		return "", false
	}
	switch raw {
	case "$", "USD", "US$", "USD$":
		return "USD", true
	case "HK$", "HKD", "HKD$":
		return "HKD", true
	case "A$", "AUD", "AUD$":
		return "AUD", true
	case "C$", "CAD", "CAD$":
		return "CAD", true
	case "S$", "SGD", "SGD$":
		return "SGD", true
	case "NZ$", "NZD", "NZD$":
		return "NZD", true
	case "EUR", "GBP", "JPY", "CNY", "RMB", "KRW":
		return raw, true
	}

	cleaned := strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r
		}
		return -1
	}, raw)
	switch cleaned {
	case "US":
		return "USD", true
	case "HK":
		return "HKD", true
	}
	if len(cleaned) == 3 {
		return cleaned, true
	}
	return "", false
}

func convertDisplayPriceToUSD(value float64, display *string, oerRates map[string]float64) (float64, bool) {
	if display == nil || strings.TrimSpace(*display) == "" {
		return value, true
	}
	currency := strings.ToUpper(strings.TrimSpace(*display))
	if currency == "" || currency == "USD" {
		return value, true
	}
	if stablecoinSet()[currency] {
		return value, true
	}
	if rate, ok := oerRates[currency]; ok && rate > 0 {
		return value / rate, true
	}
	return 0, false
}

func draftValueFromScreenshot(value *float64, display *string, oerRates map[string]float64) *float64 {
	if value == nil {
		return nil
	}
	currency, ok := normalizeDisplayCurrency(display)
	if !ok {
		return nil
	}
	if currency == "USD" || stablecoinSet()[currency] {
		v := *value
		return &v
	}
	if rate, ok := oerRates[currency]; ok && rate > 0 {
		usdRate := 1 / rate
		v := *value * usdRate
		return &v
	}
	return nil
}
