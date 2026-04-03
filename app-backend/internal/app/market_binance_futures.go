package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type binanceFuturesPremiumIndexResponse struct {
	Symbol          string `json:"symbol"`
	MarkPrice       string `json:"markPrice"`
	LastFundingRate string `json:"lastFundingRate"`
	NextFundingTime int64  `json:"nextFundingTime"`
}

func (m *marketClient) binanceFuturesPremiumIndex(ctx context.Context, symbol string) (futuresPremiumIndex, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return futuresPremiumIndex{}, fmt.Errorf("missing symbol")
	}
	query := url.Values{}
	query.Set("symbol", symbol)
	endpoint := fmt.Sprintf("%s/fapi/v1/premiumIndex?%s", strings.TrimRight(m.cfg.BinanceFuturesBaseURL, "/"), query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return futuresPremiumIndex{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return futuresPremiumIndex{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return futuresPremiumIndex{}, fmt.Errorf("binance futures status %d", resp.StatusCode)
	}

	var raw binanceFuturesPremiumIndexResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&raw); err != nil {
		return futuresPremiumIndex{}, err
	}
	markPrice, err := strconv.ParseFloat(raw.MarkPrice, 64)
	if err != nil {
		return futuresPremiumIndex{}, err
	}
	fundingRate, err := strconv.ParseFloat(raw.LastFundingRate, 64)
	if err != nil {
		return futuresPremiumIndex{}, err
	}
	return futuresPremiumIndex{
		Symbol:          raw.Symbol,
		MarkPrice:       markPrice,
		LastFundingRate: fundingRate,
		NextFundingTime: time.UnixMilli(raw.NextFundingTime).UTC(),
	}, nil
}
