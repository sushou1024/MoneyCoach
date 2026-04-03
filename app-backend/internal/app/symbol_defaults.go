package app

var defaultCryptoSymbols = map[string]struct{}{
	"AAVE":  {},
	"ADA":   {},
	"ARB":   {},
	"ASTER": {},
	"AVAX":  {},
	"BCH":   {},
	"BGB":   {},
	"BNB":   {},
	"BTC":   {},
	"DAI":   {},
	"DOGE":  {},
	"DOT":   {},
	"ENA":   {},
	"ETC":   {},
	"ETH":   {},
	"HBAR":  {},
	"HYPE":  {},
	"ICP":   {},
	"LINK":  {},
	"LTC":   {},
	"MNT":   {},
	"NEAR":  {},
	"OKB":   {},
	"OP":    {},
	"PAXG":  {},
	"PEPE":  {},
	"SHIB":  {},
	"SOL":   {},
	"SUI":   {},
	"TON":   {},
	"TRX":   {},
	"UNI":   {},
	"USDC":  {},
	"USDE":  {},
	"USDT":  {},
	"WLFI":  {},
	"XLM":   {},
	"XMR":   {},
	"XRP":   {},
	"ZEC":   {},
}

var defaultStockSymbols = map[string]struct{}{
	"AAPL": {},
	"AMZN": {},
	"CRCL": {},
	"GOOG": {},
	"META": {},
	"NVDA": {},
	"TSLA": {},
}

func defaultAssetTypeForSymbol(symbol string) string {
	if _, ok := defaultCryptoSymbols[symbol]; ok {
		return "crypto"
	}
	if _, ok := defaultStockSymbols[symbol]; ok {
		return "stock"
	}
	return ""
}
