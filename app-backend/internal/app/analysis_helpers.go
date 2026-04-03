package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"strings"
	"time"
)

var hkSymbolPattern = regexp.MustCompile(`(?i)(\d{1,5})\s*\.?\s*HK\b`)

func platformGuessToCategory(guess string) string {
	switch strings.ToLower(strings.TrimSpace(guess)) {
	case "binance", "okx", "bybit", "coinbase", "kraken", "kucoin":
		return "crypto_exchange"
	case "metamask", "trust wallet", "ledger live":
		return "wallet"
	case "futu", "ibkr", "fidelity", "robinhood", "charles schwab":
		return "broker_bank"
	default:
		return "unknown"
	}
}

func stablecoinSet() map[string]bool {
	return map[string]bool{
		"USDT":  true,
		"USDC":  true,
		"USDS":  true,
		"USDE":  true,
		"DAI":   true,
		"PYUSD": true,
		"USD1":  true,
		"FDUSD": true,
	}
}

func containsCashLabel(symbolRaw string) bool {
	upper := strings.ToUpper(strings.TrimSpace(symbolRaw))
	if upper == "" {
		return false
	}
	if upper == "USD" || upper == "US DOLLAR" {
		return true
	}
	if strings.Contains(upper, "BUYING POWER") {
		return true
	}
	if strings.Contains(upper, "AVAILABLE CASH") || strings.Contains(upper, "SETTLED CASH") {
		return true
	}
	return strings.Contains(upper, "CASH")
}

func aliasSymbol(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "bitcoin", "比特币":
		return "BTC", true
	case "ethereum", "ether", "以太坊":
		return "ETH", true
	case "tether", "泰达币":
		return "USDT", true
	case "usd coin", "美元币":
		return "USDC", true
	case "binance coin", "bnb":
		return "BNB", true
	case "solana":
		return "SOL", true
	case "xrp":
		return "XRP", true
	case "cardano":
		return "ADA", true
	case "dogecoin":
		return "DOGE", true
	default:
		return "", false
	}
}

func normalizeAssetType(assetType string) string {
	switch strings.ToLower(strings.TrimSpace(assetType)) {
	case "crypto", "cryptocurrency":
		return "crypto"
	case "stock", "equity":
		return "stock"
	case "forex", "fx", "fiat":
		return "forex"
	default:
		return ""
	}
}

func normalizeSymbol(symbol string) string {
	upper := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized, ok := normalizeHKSymbol(upper); ok {
		return normalized
	}
	replacer := strings.NewReplacer(" ", "", ".", "", "-", "", "_", "")
	return replacer.Replace(upper)
}

func normalizeHKSymbol(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	matches := hkSymbolPattern.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return "", false
	}
	digits := strings.TrimLeft(matches[len(matches)-1][1], "0")
	if digits == "" {
		return "", false
	}
	return digits + ".HK", true
}

func hkDisplaySymbol(symbol string) string {
	trimmed := strings.TrimSpace(symbol)
	if trimmed == "" {
		return ""
	}
	if normalized, ok := normalizeHKSymbol(trimmed); ok {
		return normalized
	}
	return trimmed
}

func isHongKongStockSymbol(symbol, exchangeMIC string) bool {
	if _, ok := normalizeHKSymbol(symbol); ok {
		return true
	}
	return strings.ToUpper(strings.TrimSpace(exchangeMIC)) == "XHKG"
}

func marketstackPriceCurrency(priceCurrency, symbol, exchange string) string {
	currency := strings.ToUpper(strings.TrimSpace(priceCurrency))
	if currency != "" {
		return currency
	}
	if isHongKongStockSymbol(symbol, exchange) {
		return "HKD"
	}
	return ""
}

func manualAssetKey(userID, symbolRaw, platformGuess string) string {
	normalized := normalizeSymbol(symbolRaw)
	payload := normalized + "|" + strings.TrimSpace(platformGuess)
	hash := sha256.Sum256([]byte(payload))
	return "manual:" + userID + ":" + hex.EncodeToString(hash[:6])
}

func stockAssetKey(exchangeMIC, symbol string) string {
	return "stock:mic:" + exchangeMIC + ":" + symbol
}

func resolveMarketstackAssetKey(ctx context.Context, market *marketClient, symbol, assetKey string) string {
	key := strings.TrimSpace(assetKey)
	if key != "" {
		return key
	}
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return ""
	}
	if market != nil {
		ticker, err := market.marketstackTicker(ctx, symbol)
		if err == nil && ticker.StockExchange.MIC != "" {
			return stockAssetKey(strings.ToUpper(strings.TrimSpace(ticker.StockExchange.MIC)), symbol)
		}
	}
	return stockAssetKey("UNKNOWN", symbol)
}

func hydrateHoldingIdentifiers(holding *portfolioHolding) {
	if holding == nil || holding.AssetKey == "" {
		return
	}
	if holding.CoinGeckoID == "" && strings.HasPrefix(holding.AssetKey, "crypto:cg:") {
		holding.CoinGeckoID = strings.TrimPrefix(holding.AssetKey, "crypto:cg:")
	}
}

func roundTo(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func priceDecimals(assetType string) int {
	if assetType == "crypto" {
		return 8
	}
	return 2
}

func amountDecimals(assetType string) int {
	if assetType == "crypto" {
		return 8
	}
	return 4
}

func baseStopLossPct(riskLevel string) float64 {
	switch strings.ToLower(strings.TrimSpace(riskLevel)) {
	case "conservative":
		return 0.05
	case "aggressive":
		return 0.12
	default:
		return 0.08
	}
}

func nextExecutionAt(timezone, frequency string, now time.Time) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	localNow := now.In(loc)
	var next time.Time
	switch frequency {
	case "weekly", "biweekly":
		next = nextFridayAtNine(localNow, loc)
	case "monthly":
		next = nextMonthlyFridayAtNine(localNow, loc)
	default:
		next = nextFridayAtNine(localNow, loc)
	}
	return next.UTC().Format(time.RFC3339)
}

func nextFridayAtNine(now time.Time, loc *time.Location) time.Time {
	base := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, loc)
	daysUntil := (int(time.Friday) - int(now.Weekday()) + 7) % 7
	if daysUntil == 0 && now.After(base) {
		daysUntil = 7
	}
	return base.AddDate(0, 0, daysUntil)
}

func nextMonthlyFridayAtNine(now time.Time, loc *time.Location) time.Time {
	first := time.Date(now.Year(), now.Month(), 1, 9, 0, 0, 0, loc)
	offset := (int(time.Friday) - int(first.Weekday()) + 7) % 7
	candidate := first.AddDate(0, 0, offset)
	if candidate.Before(now) || candidate.Equal(now) {
		nextMonth := first.AddDate(0, 1, 0)
		offset = (int(time.Friday) - int(nextMonth.Weekday()) + 7) % 7
		candidate = nextMonth.AddDate(0, 0, offset)
	}
	return candidate
}
