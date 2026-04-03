package app

import (
	"encoding/json"
	"strconv"
)

func parseBinanceKline(raw []any) (ohlcPoint, bool) {
	if len(raw) < 7 {
		return ohlcPoint{}, false
	}
	openTime, ok := parseBinanceInt(raw[0])
	if !ok {
		return ohlcPoint{}, false
	}
	closeTime, ok := parseBinanceInt(raw[6])
	if !ok {
		closeTime = openTime
	}
	open, ok := parseBinanceFloat(raw[1])
	if !ok {
		return ohlcPoint{}, false
	}
	high, ok := parseBinanceFloat(raw[2])
	if !ok {
		return ohlcPoint{}, false
	}
	low, ok := parseBinanceFloat(raw[3])
	if !ok {
		return ohlcPoint{}, false
	}
	close, ok := parseBinanceFloat(raw[4])
	if !ok {
		return ohlcPoint{}, false
	}
	volume := 0.0
	if value, ok := parseBinanceFloat(raw[5]); ok {
		volume = value
	}
	return ohlcPoint{Timestamp: closeTime / 1000, Open: open, High: high, Low: low, Close: close, Volume: volume}, true
}

func parseBinanceFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, err == nil
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func parseBinanceInt(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case json.Number:
		parsed, err := v.Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func binanceSymbolCandidates(symbol string) []string {
	return []string{symbol + "USDT", "1000" + symbol + "USDT"}
}
