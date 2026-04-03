package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"testing"
	"time"
)

const (
	cmcBaseURL            = "https://pro-api.coinmarketcap.com"
	coingeckoBaseURL      = "https://pro-api.coingecko.com/api/v3"
	marketstackBaseURL    = "https://api.marketstack.com/v2"
	oerBaseURL            = "https://openexchangerates.org/api"
	binanceFuturesBaseURL = "https://fapi.binance.com"
)

// Post-MVP: Fear & Greed is optional; this test is not required for MVP.
func TestCMCFearGreedAPI(t *testing.T) {
	if !optionalAPITestsEnabled() {
		t.Skip("RUN_OPTIONAL_API_TESTS not set; CoinMarketCap Fear & Greed is post-MVP")
	}
	apiKey := requireEnv(t, "CMC_PRO_API_KEY")
	client := &http.Client{Timeout: 20 * time.Second}
	t.Run("Latest", func(t *testing.T) {
		var resp cmcFearGreedLatestResponse
		cmcGet(t, client, apiKey, "/v3/fear-and-greed/latest", url.Values{}, &resp)
		assertCMCStatusOK(t, resp.Status)
		if resp.Data.Value < 0 || resp.Data.ValueClassification == "" {
			t.Fatalf("unexpected fear & greed latest data: %#v", resp.Data)
		}
		if resp.Data.UpdateTime == "" {
			t.Fatalf("missing update_time in fear & greed latest: %#v", resp.Data)
		}
	})

	t.Run("Historical", func(t *testing.T) {
		query := url.Values{}
		query.Set("limit", "5")

		var resp cmcFearGreedHistoricalResponse
		cmcGet(t, client, apiKey, "/v3/fear-and-greed/historical", query, &resp)
		assertCMCStatusOK(t, resp.Status)
		if len(resp.Data) == 0 {
			t.Fatal("expected fear & greed historical data, got empty")
		}
		if resp.Data[0].Timestamp == "" || resp.Data[0].ValueClassification == "" {
			t.Fatalf("unexpected fear & greed historical entry: %#v", resp.Data[0])
		}
	})
}

func TestCoinGeckoAPI(t *testing.T) {
	apiKey := requireEnv(t, "COINGECKO_PRO_API_KEY")
	client := &http.Client{Timeout: 20 * time.Second}
	coinID := "bitcoin"
	now := time.Now().UTC()
	from := strconv.FormatInt(now.Add(-48*time.Hour).Unix(), 10)
	to := strconv.FormatInt(now.Unix(), 10)
	var rangeResp coinGeckoMarketChartRangeResponse

	t.Run("CoinsList", func(t *testing.T) {
		query := url.Values{}
		query.Set("include_platform", "false")

		var resp []coinGeckoCoinListEntry
		coinGeckoGet(t, client, apiKey, "/coins/list", query, &resp)
		if len(resp) == 0 {
			t.Fatal("expected coins list, got empty")
		}
		if !coinGeckoListContains(resp, coinID) {
			t.Fatalf("expected coin ID %q in list", coinID)
		}
	})

	t.Run("SimplePrice", func(t *testing.T) {
		query := url.Values{}
		query.Set("ids", coinID)
		query.Set("vs_currencies", "usd")
		query.Set("include_market_cap", "true")
		query.Set("include_24hr_vol", "true")
		query.Set("include_24hr_change", "true")
		query.Set("include_last_updated_at", "true")

		var resp coinGeckoSimplePriceResponse
		coinGeckoGet(t, client, apiKey, "/simple/price", query, &resp)
		price, ok := resp[coinID]
		if !ok {
			t.Fatalf("expected price for %q, got keys: %v", coinID, mapKeys(resp))
		}
		if price.USD <= 0 {
			t.Fatalf("unexpected USD price: %#v", price)
		}
		if price.LastUpdatedAt == 0 {
			t.Fatalf("missing last_updated_at: %#v", price)
		}
	})

	t.Run("MarketChartRange", func(t *testing.T) {
		query := url.Values{}
		query.Set("vs_currency", "usd")
		query.Set("from", from)
		query.Set("to", to)

		var resp coinGeckoMarketChartRangeResponse
		coinGeckoGet(t, client, apiKey, fmt.Sprintf("/coins/%s/market_chart/range", coinID), query, &resp)
		if len(resp.Prices) == 0 || len(resp.TotalVolumes) == 0 {
			t.Fatalf("expected market chart data, got %#v", resp)
		}
		if len(resp.Prices[0]) < 2 {
			t.Fatalf("unexpected market chart price entry: %#v", resp.Prices[0])
		}
		rangeResp = resp
	})

	t.Run("DerivedDailyOHLCV", func(t *testing.T) {
		if len(rangeResp.Prices) == 0 {
			t.Skip("market chart range data missing")
		}
		daily, err := deriveDailyOHLCV(rangeResp.Prices, rangeResp.TotalVolumes)
		if err != nil {
			t.Fatalf("derive daily OHLCV: %v", err)
		}
		if len(daily) < 2 {
			t.Fatalf("expected at least 2 daily candles, got %d", len(daily))
		}
		for _, day := range daily {
			if day.Open <= 0 || day.Close <= 0 || day.High <= 0 || day.Low <= 0 {
				t.Fatalf("unexpected daily OHLC values: %#v", day)
			}
			if day.High < day.Open || day.High < day.Close {
				t.Fatalf("daily high below open/close: %#v", day)
			}
			if day.Low > day.Open || day.Low > day.Close {
				t.Fatalf("daily low above open/close: %#v", day)
			}
			if day.Volume < 0 {
				t.Fatalf("unexpected daily volume: %#v", day)
			}
		}
	})

	t.Run("CoinsMarkets", func(t *testing.T) {
		query := url.Values{}
		query.Set("vs_currency", "usd")
		query.Set("ids", coinID)
		query.Set("price_change_percentage", "24h")

		var resp []coinGeckoMarketsItem
		coinGeckoGet(t, client, apiKey, "/coins/markets", query, &resp)
		if len(resp) == 0 {
			t.Fatal("expected coins markets data, got empty")
		}
		item := resp[0]
		if item.ID == "" || item.Symbol == "" || item.Name == "" {
			t.Fatalf("unexpected markets item: %#v", item)
		}
		if item.CurrentPrice <= 0 || item.MarketCap <= 0 || item.TotalVolume <= 0 {
			t.Fatalf("unexpected markets values: %#v", item)
		}
		if item.ATH <= 0 || item.ATL <= 0 {
			t.Fatalf("unexpected ATH/ATL values: %#v", item)
		}
	})

	t.Run("CoinData", func(t *testing.T) {
		if !optionalAPITestsEnabled() {
			t.Skip("RUN_OPTIONAL_API_TESTS not set; /coins/{id} is optional for MVP")
		}
		query := url.Values{}
		query.Set("localization", "false")
		query.Set("tickers", "false")
		query.Set("market_data", "false")
		query.Set("community_data", "false")
		query.Set("developer_data", "false")
		query.Set("sparkline", "false")

		var resp coinGeckoCoinData
		coinGeckoGet(t, client, apiKey, fmt.Sprintf("/coins/%s", coinID), query, &resp)
		if resp.ID == "" || resp.Symbol == "" || resp.Name == "" {
			t.Fatalf("unexpected coin data: %#v", resp)
		}
		if resp.Image.Thumb == "" || resp.Image.Small == "" || resp.Image.Large == "" {
			t.Fatalf("missing coin image data: %#v", resp.Image)
		}
	})
}

func TestBinanceSpotAPI(t *testing.T) {
	baseURL := requireEnv(t, "BINANCE_API_BASE_URL")
	client := &http.Client{Timeout: 20 * time.Second}

	t.Run("Klines", func(t *testing.T) {
		query := url.Values{}
		query.Set("symbol", "BTCUSDT")
		query.Set("interval", "4h")
		query.Set("limit", "1")

		endpoint := fmt.Sprintf("%s/api/v3/klines?%s", baseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp [][]any
		getJSON(t, client, req, &resp)
		if len(resp) == 0 {
			t.Fatal("expected klines data, got empty")
		}
		kline, err := parseBinanceKline(resp[0])
		if err != nil {
			t.Fatalf("parse kline: %v", err)
		}
		if kline.OpenTime == 0 || kline.CloseTime == 0 {
			t.Fatalf("unexpected kline times: %#v", kline)
		}
		if kline.Open == "" || kline.Close == "" {
			t.Fatalf("unexpected kline prices: %#v", kline)
		}
	})
}

func TestBinanceFuturesAPI(t *testing.T) {
	client := &http.Client{Timeout: 20 * time.Second}

	t.Run("PremiumIndex", func(t *testing.T) {
		query := url.Values{}
		query.Set("symbol", "BTCUSDT")

		endpoint := fmt.Sprintf("%s/fapi/v1/premiumIndex?%s", binanceFuturesBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp binanceFuturesPremiumIndex
		getJSON(t, client, req, &resp)
		if resp.Symbol == "" || resp.MarkPrice == "" {
			t.Fatalf("unexpected premium index response: %#v", resp)
		}
		markPrice, err := strconv.ParseFloat(resp.MarkPrice, 64)
		if err != nil || markPrice <= 0 {
			t.Fatalf("unexpected mark price: %q (err=%v)", resp.MarkPrice, err)
		}
		if resp.LastFundingRate == "" {
			t.Fatalf("missing last funding rate: %#v", resp)
		}
		if _, err := strconv.ParseFloat(resp.LastFundingRate, 64); err != nil {
			t.Fatalf("parse last funding rate %q: %v", resp.LastFundingRate, err)
		}
		if resp.NextFundingTime == 0 || resp.Time == 0 {
			t.Fatalf("unexpected premium index timestamps: %#v", resp)
		}
	})

	t.Run("FundingRate", func(t *testing.T) {
		if !optionalAPITestsEnabled() {
			t.Skip("RUN_OPTIONAL_API_TESTS not set; funding rate history is optional for MVP")
		}
		query := url.Values{}
		query.Set("symbol", "BTCUSDT")
		query.Set("limit", "1")

		endpoint := fmt.Sprintf("%s/fapi/v1/fundingRate?%s", binanceFuturesBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp []binanceFuturesFundingRate
		getJSON(t, client, req, &resp)
		if len(resp) == 0 {
			t.Fatal("expected funding rate data, got empty")
		}
		if resp[0].Symbol == "" || resp[0].FundingRate == "" {
			t.Fatalf("unexpected funding rate response: %#v", resp[0])
		}
		if _, err := strconv.ParseFloat(resp[0].FundingRate, 64); err != nil {
			t.Fatalf("parse funding rate %q: %v", resp[0].FundingRate, err)
		}
		if resp[0].MarkPrice != "" {
			markPrice, err := strconv.ParseFloat(resp[0].MarkPrice, 64)
			if err != nil || markPrice <= 0 {
				t.Fatalf("unexpected mark price: %q (err=%v)", resp[0].MarkPrice, err)
			}
		}
		if resp[0].FundingTime == 0 {
			t.Fatalf("unexpected funding time: %#v", resp[0])
		}
	})
}

func TestMarketstackAPI(t *testing.T) {
	accessKey := requireEnv(t, "MARKETSTACK_ACCESS_KEY")
	client := &http.Client{Timeout: 20 * time.Second}

	t.Run("Tickers", func(t *testing.T) {
		endpoint := fmt.Sprintf("%s/tickers/AAPL?access_key=%s", marketstackBaseURL, url.QueryEscape(accessKey))
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp marketstackTickerResponse
		getJSON(t, client, req, &resp)
		if resp.Symbol == "" || resp.Name == "" {
			t.Fatalf("unexpected ticker response: %#v", resp)
		}
		if resp.StockExchange.MIC == "" {
			t.Fatalf("missing stock exchange MIC: %#v", resp.StockExchange)
		}
	})

	t.Run("EOD", func(t *testing.T) {
		query := url.Values{}
		query.Set("access_key", accessKey)
		query.Set("symbols", "AAPL")
		query.Set("limit", "1")

		endpoint := fmt.Sprintf("%s/eod?%s", marketstackBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp marketstackEODResponse
		getJSON(t, client, req, &resp)
		if len(resp.Data) == 0 {
			t.Fatal("expected EOD data, got empty")
		}
		if resp.Data[0].Symbol == "" || resp.Data[0].Close == 0 {
			t.Fatalf("unexpected EOD entry: %#v", resp.Data[0])
		}
	})

	t.Run("EODLatest", func(t *testing.T) {
		query := url.Values{}
		query.Set("access_key", accessKey)
		query.Set("symbols", "AAPL")
		query.Set("limit", "1")

		endpoint := fmt.Sprintf("%s/eod/latest?%s", marketstackBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp marketstackEODResponse
		getJSON(t, client, req, &resp)
		if len(resp.Data) == 0 {
			t.Fatal("expected EOD latest data, got empty")
		}
		if resp.Data[0].Symbol == "" || resp.Data[0].Close == 0 {
			t.Fatalf("unexpected EOD latest entry: %#v", resp.Data[0])
		}
	})

	t.Run("Intraday", func(t *testing.T) {
		if !optionalAPITestsEnabled() {
			t.Skip("RUN_OPTIONAL_API_TESTS not set; Marketstack intraday is post-MVP")
		}
		query := url.Values{}
		query.Set("access_key", accessKey)
		query.Set("symbols", "AAPL")
		query.Set("interval", "1hour")
		query.Set("limit", "1")

		endpoint := fmt.Sprintf("%s/intraday?%s", marketstackBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp marketstackIntradayResponse
		getJSON(t, client, req, &resp)
		if len(resp.Data) == 0 {
			t.Fatal("expected intraday data, got empty")
		}
		if resp.Data[0].Symbol == "" || resp.Data[0].Close == 0 {
			t.Fatalf("unexpected intraday entry: %#v", resp.Data[0])
		}
	})

	t.Run("IntradayLatest", func(t *testing.T) {
		if !optionalAPITestsEnabled() {
			t.Skip("RUN_OPTIONAL_API_TESTS not set; Marketstack intraday is post-MVP")
		}
		query := url.Values{}
		query.Set("access_key", accessKey)
		query.Set("symbols", "AAPL")
		query.Set("interval", "1hour")

		endpoint := fmt.Sprintf("%s/intraday/latest?%s", marketstackBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp marketstackIntradayResponse
		getJSON(t, client, req, &resp)
		if len(resp.Data) == 0 {
			t.Fatal("expected intraday latest data, got empty")
		}
		if resp.Data[0].Symbol == "" || resp.Data[0].Close == 0 {
			t.Fatalf("unexpected intraday latest entry: %#v", resp.Data[0])
		}
	})
}

func TestOpenExchangeRatesAPI(t *testing.T) {
	appID := requireEnv(t, "OPEN_EXCHANGE_APP_ID")
	client := &http.Client{Timeout: 20 * time.Second}

	t.Run("Latest", func(t *testing.T) {
		query := url.Values{}
		query.Set("app_id", appID)
		query.Set("symbols", "USD,EUR")

		endpoint := fmt.Sprintf("%s/latest.json?%s", oerBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp oerLatestResponse
		getJSON(t, client, req, &resp)
		if resp.Base == "" || resp.Timestamp == 0 {
			t.Fatalf("unexpected latest response: %#v", resp)
		}
		if len(resp.Rates) == 0 {
			t.Fatal("expected rates, got empty")
		}
		if _, ok := resp.Rates["USD"]; !ok {
			t.Fatalf("expected USD rate, got keys: %v", mapKeys(resp.Rates))
		}
	})

	t.Run("Currencies", func(t *testing.T) {
		query := url.Values{}
		query.Set("app_id", appID)

		endpoint := fmt.Sprintf("%s/currencies.json?%s", oerBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build request: %v", err)
		}

		var resp map[string]string
		getJSON(t, client, req, &resp)
		if len(resp) == 0 {
			t.Fatal("expected currencies, got empty")
		}
		if resp["USD"] == "" {
			t.Fatalf("expected USD currency name, got: %#v", resp["USD"])
		}
	})
}

func getJSON(t *testing.T, client *http.Client, req *http.Request, out any) {
	t.Helper()
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	debugLogf(t, "HTTP %s %s -> %d (%s)", req.Method, sanitizeURL(req.URL), resp.StatusCode, time.Since(start))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func cmcGet(t *testing.T, client *http.Client, apiKey, path string, query url.Values, out any) {
	t.Helper()
	endpoint := fmt.Sprintf("%s%s?%s", cmcBaseURL, path, query.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)

	getJSON(t, client, req, out)
}

func coinGeckoGet(t *testing.T, client *http.Client, apiKey, path string, query url.Values, out any) {
	t.Helper()
	endpoint := fmt.Sprintf("%s%s?%s", coingeckoBaseURL, path, query.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-cg-pro-api-key", apiKey)

	getJSON(t, client, req, out)
}

func assertCMCStatusOK(t *testing.T, status cmcStatus) {
	t.Helper()
	if status.ErrorCode != 0 {
		t.Fatalf("CMC error: code=%d message=%q", status.ErrorCode, status.ErrorMessage)
	}
}

func mapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func coinGeckoListContains(entries []coinGeckoCoinListEntry, coinID string) bool {
	for _, entry := range entries {
		if entry.ID == coinID {
			return true
		}
	}
	return false
}

func parseBinanceKline(raw []any) (binanceKline, error) {
	if len(raw) < 11 {
		return binanceKline{}, fmt.Errorf("expected 11+ fields, got %d", len(raw))
	}
	openTime, err := toInt64(raw[0])
	if err != nil {
		return binanceKline{}, fmt.Errorf("open time: %w", err)
	}
	closeTime, err := toInt64(raw[6])
	if err != nil {
		return binanceKline{}, fmt.Errorf("close time: %w", err)
	}
	open, ok := raw[1].(string)
	if !ok {
		return binanceKline{}, fmt.Errorf("open price type %T", raw[1])
	}
	close, ok := raw[4].(string)
	if !ok {
		return binanceKline{}, fmt.Errorf("close price type %T", raw[4])
	}

	return binanceKline{
		OpenTime:  openTime,
		Open:      open,
		CloseTime: closeTime,
		Close:     close,
	}, nil
}

func deriveDailyOHLCV(prices [][]float64, volumes [][]float64) ([]dailyOHLCV, error) {
	if len(prices) == 0 {
		return nil, fmt.Errorf("no price data")
	}

	days := map[string]*dailyAccumulator{}
	for _, entry := range prices {
		if len(entry) < 2 {
			return nil, fmt.Errorf("unexpected price entry: %#v", entry)
		}
		tsMs := int64(entry[0])
		price := entry[1]
		dayStart := dayStartUTC(tsMs)
		key := dayStart.Format("2006-01-02")

		acc := days[key]
		if acc == nil {
			acc = &dailyAccumulator{
				dailyOHLCV: dailyOHLCV{DayStart: dayStart},
			}
			days[key] = acc
		}
		if !acc.openSet {
			acc.Open = price
			acc.High = price
			acc.Low = price
			acc.Close = price
			acc.openSet = true
			continue
		}
		if price > acc.High {
			acc.High = price
		}
		if price < acc.Low {
			acc.Low = price
		}
		acc.Close = price
	}

	for _, entry := range volumes {
		if len(entry) < 2 {
			return nil, fmt.Errorf("unexpected volume entry: %#v", entry)
		}
		tsMs := int64(entry[0])
		volume := entry[1]
		dayStart := dayStartUTC(tsMs)
		key := dayStart.Format("2006-01-02")
		acc := days[key]
		if acc == nil {
			continue
		}
		if tsMs >= acc.volumeTime {
			acc.Volume = volume
			acc.volumeTime = tsMs
		}
	}

	daily := make([]dailyOHLCV, 0, len(days))
	for _, acc := range days {
		daily = append(daily, acc.dailyOHLCV)
	}
	sort.Slice(daily, func(i, j int) bool {
		return daily[i].DayStart.Before(daily[j].DayStart)
	})
	return daily, nil
}

func dayStartUTC(tsMs int64) time.Time {
	ts := time.Unix(0, tsMs*int64(time.Millisecond)).UTC()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	default:
		return 0, fmt.Errorf("unexpected type %T", value)
	}
}

type cmcStatus struct {
	Timestamp    string       `json:"timestamp"`
	ErrorCode    cmcErrorCode `json:"error_code"`
	ErrorMessage string       `json:"error_message"`
	Elapsed      int          `json:"elapsed"`
	CreditCount  int          `json:"credit_count"`
	Notice       string       `json:"notice"`
}

type cmcErrorCode int

func (c *cmcErrorCode) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*c = 0
		return nil
	}

	var number int
	if err := json.Unmarshal(data, &number); err == nil {
		*c = cmcErrorCode(number)
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		if text == "" {
			*c = 0
			return nil
		}
		parsed, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("parse error_code %q: %w", text, err)
		}
		*c = cmcErrorCode(parsed)
		return nil
	}

	return fmt.Errorf("unexpected error_code: %s", string(data))
}

type cmcFearGreedLatestResponse struct {
	Data   cmcFearGreedLatestData `json:"data"`
	Status cmcStatus              `json:"status"`
}

type cmcFearGreedLatestData struct {
	Value               int    `json:"value"`
	ValueClassification string `json:"value_classification"`
	UpdateTime          string `json:"update_time"`
}

type cmcFearGreedHistoricalResponse struct {
	Data   []cmcFearGreedHistoricalEntry `json:"data"`
	Status cmcStatus                     `json:"status"`
}

type cmcFearGreedHistoricalEntry struct {
	Timestamp           string `json:"timestamp"`
	Value               int    `json:"value"`
	ValueClassification string `json:"value_classification"`
}

type coinGeckoCoinListEntry struct {
	ID        string            `json:"id"`
	Symbol    string            `json:"symbol"`
	Name      string            `json:"name"`
	Platforms map[string]string `json:"platforms,omitempty"`
}

type coinGeckoSimplePriceResponse map[string]coinGeckoSimplePrice

type coinGeckoSimplePrice struct {
	USD           float64 `json:"usd"`
	USDMarketCap  float64 `json:"usd_market_cap"`
	USD24hVol     float64 `json:"usd_24h_vol"`
	USD24hChange  float64 `json:"usd_24h_change"`
	LastUpdatedAt int64   `json:"last_updated_at"`
}

type coinGeckoMarketChartRangeResponse struct {
	Prices       [][]float64 `json:"prices"`
	MarketCaps   [][]float64 `json:"market_caps"`
	TotalVolumes [][]float64 `json:"total_volumes"`
}

type coinGeckoMarketsItem struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	TotalVolume              float64 `json:"total_volume"`
	High24h                  float64 `json:"high_24h"`
	Low24h                   float64 `json:"low_24h"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	ATH                      float64 `json:"ath"`
	ATL                      float64 `json:"atl"`
	LastUpdated              string  `json:"last_updated"`
}

type coinGeckoCoinData struct {
	ID     string         `json:"id"`
	Symbol string         `json:"symbol"`
	Name   string         `json:"name"`
	Image  coinGeckoImage `json:"image"`
	Links  coinGeckoLinks `json:"links"`
}

type coinGeckoImage struct {
	Thumb string `json:"thumb"`
	Small string `json:"small"`
	Large string `json:"large"`
}

type coinGeckoLinks struct {
	Homepage []string `json:"homepage"`
}

type binanceKline struct {
	OpenTime  int64
	Open      string
	CloseTime int64
	Close     string
}

type binanceFuturesPremiumIndex struct {
	Symbol               string `json:"symbol"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`
	EstimatedSettlePrice string `json:"estimatedSettlePrice"`
	LastFundingRate      string `json:"lastFundingRate"`
	InterestRate         string `json:"interestRate"`
	NextFundingTime      int64  `json:"nextFundingTime"`
	Time                 int64  `json:"time"`
}

type binanceFuturesFundingRate struct {
	Symbol      string `json:"symbol"`
	FundingRate string `json:"fundingRate"`
	FundingTime int64  `json:"fundingTime"`
	MarkPrice   string `json:"markPrice"`
}

type marketstackTickerResponse struct {
	Name          string                   `json:"name"`
	Symbol        string                   `json:"symbol"`
	Sector        string                   `json:"sector"`
	Industry      string                   `json:"industry"`
	StockExchange marketstackStockExchange `json:"stock_exchange"`
}

type marketstackStockExchange struct {
	Name string `json:"name"`
	MIC  string `json:"mic"`
}

type marketstackPagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

type marketstackEODResponse struct {
	Pagination marketstackPagination `json:"pagination"`
	Data       []marketstackEODBar   `json:"data"`
}

type marketstackEODBar struct {
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        float64 `json:"volume"`
	Symbol        string  `json:"symbol"`
	Exchange      string  `json:"exchange"`
	Date          string  `json:"date"`
	PriceCurrency string  `json:"price_currency"`
}

type marketstackIntradayResponse struct {
	Pagination marketstackPagination    `json:"pagination"`
	Data       []marketstackIntradayBar `json:"data"`
}

type marketstackIntradayBar struct {
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        float64 `json:"volume"`
	Symbol        string  `json:"symbol"`
	Exchange      string  `json:"exchange"`
	Date          string  `json:"date"`
	PriceCurrency string  `json:"price_currency"`
}

type oerLatestResponse struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

type dailyOHLCV struct {
	DayStart time.Time
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
}

type dailyAccumulator struct {
	dailyOHLCV
	openSet    bool
	volumeTime int64
}
