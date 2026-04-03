package app

import (
	"context"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleMarketDataOHLCV(w http.ResponseWriter, r *http.Request) {
	assetType := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("asset_type")))
	symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	assetKey := strings.TrimSpace(r.URL.Query().Get("asset_key"))
	interval := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("interval")))
	start, hasStart := parseDateQuery(r.URL.Query().Get("start"))
	end, hasEnd := parseDateQuery(r.URL.Query().Get("end"))
	if assetType == "" || (symbol == "" && assetKey == "") {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "asset_type and symbol/asset_key required", nil)
		return
	}
	if interval == "" {
		interval = "1d"
	}

	series := make([][]any, 0)
	quoteCurrency := "USD"
	switch assetType {
	case "crypto":
		coinID := ""
		if strings.HasPrefix(assetKey, "crypto:cg:") {
			coinID = strings.TrimPrefix(assetKey, "crypto:cg:")
		}
		if interval == "4h" {
			quoteCurrency = "USDT"
			if symbol == "" {
				s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "symbol required for 4h", nil)
				return
			}
			points, err := s.market.binanceKlines(r.Context(), symbol+"USDT", "4h", binanceKlinesLimit)
			if err != nil {
				s.writeError(w, http.StatusInternalServerError, "MARKET_ERROR", "binance klines failed", nil)
				return
			}
			for _, point := range points {
				if withinDateRange(point.Timestamp, start, end, hasStart, hasEnd) {
					series = append(series, []any{time.Unix(point.Timestamp, 0).UTC().Format(time.RFC3339), point.Open, point.High, point.Low, point.Close, point.Volume})
				}
			}
		} else {
			quoteCurrency = "USD"
			if coinID == "" && symbol != "" {
				coinID = resolveCoinGeckoIDFromSymbol(r.Context(), s.market, symbol)
			}
			if coinID == "" {
				s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "asset_key or resolvable symbol required for crypto daily", nil)
				return
			}
			if hasStart || hasEnd {
				startTime := start
				if !hasStart {
					startTime = time.Now().UTC().AddDate(0, 0, -defaultSupportLookback)
				}
				endTime := end
				if !hasEnd {
					endTime = time.Now().UTC()
				}
				points := fetchCoinGeckoOHLCRange(r.Context(), s.market, coinID, startTime, endTime)
				for _, point := range points {
					series = append(series, []any{time.Unix(point.Timestamp, 0).UTC().Format(time.RFC3339), point.Open, point.High, point.Low, point.Close, point.Volume})
				}
			} else {
				points := fetchCoinGeckoOHLC(r.Context(), s.market, coinID, defaultSupportLookback)
				for _, point := range points {
					series = append(series, []any{time.Unix(point.Timestamp, 0).UTC().Format(time.RFC3339), point.Open, point.High, point.Low, point.Close, point.Volume})
				}
			}
		}
	case "stock":
		if symbol == "" && strings.HasPrefix(assetKey, "stock:mic:") {
			parts := strings.Split(assetKey, ":")
			if len(parts) >= 4 {
				symbol = parts[len(parts)-1]
			}
		}
		if symbol == "" {
			s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "symbol required", nil)
			return
		}
		startTime := start
		endTime := end
		if !hasStart {
			startTime = time.Now().UTC().AddDate(0, 0, -defaultSupportLookback)
		}
		if !hasEnd {
			endTime = time.Now().UTC()
		}
		points := fetchMarketstackSeriesRange(r.Context(), s.market, symbol, assetKey, startTime, endTime)
		quoteCurrency = s.resolveStockQuoteCurrency(r.Context(), assetKey, symbol)
		for _, point := range points {
			if withinDateRange(point.Timestamp, start, end, hasStart, hasEnd) {
				series = append(series, []any{time.Unix(point.Timestamp, 0).UTC().Format(time.RFC3339), point.Open, point.High, point.Low, point.Close, point.Volume})
			}
		}
	default:
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unsupported asset_type", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"series": series, "quote_currency": quoteCurrency})
}

func parseDateQuery(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func withinDateRange(timestamp int64, start, end time.Time, hasStart, hasEnd bool) bool {
	pointTime := time.Unix(timestamp, 0).UTC()
	if hasStart && pointTime.Before(start) {
		return false
	}
	if hasEnd {
		endExclusive := end.Add(24 * time.Hour)
		if pointTime.After(endExclusive) || pointTime.Equal(endExclusive) {
			return false
		}
	}
	return true
}

func resolveCoinGeckoIDFromSymbol(ctx context.Context, market *marketClient, symbol string) string {
	coinList, err := market.coinGeckoList(ctx)
	if err != nil {
		return ""
	}
	symbolToIDs := make(map[string][]string)
	for _, coin := range coinList {
		if coin.Symbol == "" {
			continue
		}
		upper := strings.ToUpper(coin.Symbol)
		symbolToIDs[upper] = append(symbolToIDs[upper], coin.ID)
	}
	return resolveCoinGeckoID(ctx, market, strings.ToUpper(strings.TrimSpace(symbol)), symbolToIDs)
}
