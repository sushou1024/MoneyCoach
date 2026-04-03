package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	coinGeckoBaseURL   = "https://pro-api.coingecko.com/api/v3"
	marketstackBaseURL = "https://api.marketstack.com/v2"
	oerBaseURL         = "https://openexchangerates.org/api"
)

type marketClient struct {
	cfg    Config
	client *http.Client
	redis  *redisStore
	cache  *marketCacheStore
	logger *log.Logger
}

func newMarketClient(cfg Config, client *http.Client, redis *redisStore, cache *marketCacheStore, logger *log.Logger) *marketClient {
	return &marketClient{cfg: cfg, client: client, redis: redis, cache: cache, logger: logger}
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
	Image                    string  `json:"image"`
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

type oerLatestResponse struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

func (m *marketClient) coinGeckoList(ctx context.Context) ([]coinGeckoCoinListEntry, error) {
	cacheKey := "cache:coingecko:list"
	var cached []coinGeckoCoinListEntry
	if ok, err := m.redis.getJSON(ctx, cacheKey, &cached); err == nil && ok {
		return cached, nil
	}

	query := url.Values{}
	query.Set("include_platform", "false")
	endpoint := fmt.Sprintf("%s/coins/list?%s", coinGeckoBaseURL, query.Encode())
	var resp []coinGeckoCoinListEntry
	if err := m.getJSON(ctx, endpoint, map[string]string{"x-cg-pro-api-key": m.cfg.CoinGeckoAPIKey}, &resp); err != nil {
		return nil, err
	}
	_ = m.redis.setJSON(ctx, cacheKey, resp, 24*time.Hour)
	return resp, nil
}

func (m *marketClient) coinGeckoSimplePrice(ctx context.Context, ids []string) (coinGeckoSimplePriceResponse, error) {
	if len(ids) == 0 {
		return coinGeckoSimplePriceResponse{}, nil
	}
	sort.Strings(ids)
	query := url.Values{}
	query.Set("ids", strings.Join(ids, ","))
	query.Set("vs_currencies", "usd")
	query.Set("include_market_cap", "true")
	query.Set("include_24hr_vol", "true")
	query.Set("include_24hr_change", "true")
	query.Set("include_last_updated_at", "true")
	endpoint := fmt.Sprintf("%s/simple/price?%s", coinGeckoBaseURL, query.Encode())
	var resp coinGeckoSimplePriceResponse
	if err := m.getJSON(ctx, endpoint, map[string]string{"x-cg-pro-api-key": m.cfg.CoinGeckoAPIKey}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *marketClient) coinGeckoMarkets(ctx context.Context, ids []string) ([]coinGeckoMarketsItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := url.Values{}
	query.Set("vs_currency", "usd")
	query.Set("ids", strings.Join(ids, ","))
	query.Set("price_change_percentage", "24h")
	endpoint := fmt.Sprintf("%s/coins/markets?%s", coinGeckoBaseURL, query.Encode())
	var resp []coinGeckoMarketsItem
	if err := m.getJSON(ctx, endpoint, map[string]string{"x-cg-pro-api-key": m.cfg.CoinGeckoAPIKey}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *marketClient) coinGeckoLogos(ctx context.Context, ids []string) (map[string]string, error) {
	ids = uniqueSortedStrings(ids)
	if len(ids) == 0 {
		return map[string]string{}, nil
	}
	cacheKey := "cache:coingecko:logos:" + hashStrings(ids)
	if m.redis != nil {
		var cached map[string]string
		if ok, err := m.redis.getJSON(ctx, cacheKey, &cached); ok && err == nil {
			return cached, nil
		}
	}

	items, err := m.coinGeckoMarkets(ctx, ids)
	if err != nil {
		return nil, err
	}
	logos := make(map[string]string, len(items))
	for _, item := range items {
		if item.ID == "" || item.Image == "" {
			continue
		}
		logos[item.ID] = item.Image
	}
	if m.redis != nil {
		_ = m.redis.setJSON(ctx, cacheKey, logos, 12*time.Hour)
	}
	return logos, nil
}

func (m *marketClient) coinGeckoTopMarkets(ctx context.Context, limit int) ([]coinGeckoMarketsItem, error) {
	if limit <= 0 {
		return nil, nil
	}
	cacheKey := fmt.Sprintf("cache:coingecko:markets:top:%d", limit)
	var cached []coinGeckoMarketsItem
	if ok, err := m.redis.getJSON(ctx, cacheKey, &cached); ok && err == nil {
		return cached, nil
	}
	query := url.Values{}
	query.Set("vs_currency", "usd")
	query.Set("order", "market_cap_desc")
	query.Set("per_page", strconv.Itoa(limit))
	query.Set("page", "1")
	query.Set("sparkline", "false")
	endpoint := fmt.Sprintf("%s/coins/markets?%s", coinGeckoBaseURL, query.Encode())
	var resp []coinGeckoMarketsItem
	if err := m.getJSON(ctx, endpoint, map[string]string{"x-cg-pro-api-key": m.cfg.CoinGeckoAPIKey}, &resp); err != nil {
		return nil, err
	}
	_ = m.redis.setJSON(ctx, cacheKey, resp, 15*time.Minute)
	return resp, nil
}

func (m *marketClient) coinGeckoMarketChartRange(ctx context.Context, coinID string, from, to time.Time) (coinGeckoMarketChartRangeResponse, error) {
	query := url.Values{}
	query.Set("vs_currency", "usd")
	query.Set("from", strconv.FormatInt(from.Unix(), 10))
	query.Set("to", strconv.FormatInt(to.Unix(), 10))
	endpoint := fmt.Sprintf("%s/coins/%s/market_chart/range?%s", coinGeckoBaseURL, url.PathEscape(coinID), query.Encode())
	var resp coinGeckoMarketChartRangeResponse
	if err := m.getJSON(ctx, endpoint, map[string]string{"x-cg-pro-api-key": m.cfg.CoinGeckoAPIKey}, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (m *marketClient) marketstackTicker(ctx context.Context, symbol string) (marketstackTickerResponse, error) {
	if override, ok := marketstackTickerOverride(symbol); ok {
		return override, nil
	}
	cacheKey := cacheKeyMarketstackTicker(symbol)
	if m.cache != nil {
		var cached marketstackTickerResponse
		if status, err := m.cache.getJSON(ctx, cacheKindMarketstackTick, cacheKey, &cached); err == nil {
			if status == cacheReadHit {
				return cached, nil
			}
			if status == cacheReadNegative {
				return marketstackTickerResponse{}, errMarketDataCacheNegative
			}
		}
	}

	endpoint := fmt.Sprintf("%s/tickers/%s?access_key=%s", marketstackBaseURL, url.PathEscape(symbol), url.QueryEscape(m.cfg.MarketstackAccessKey))
	start := time.Now()
	if m.logger != nil {
		m.logger.Printf("marketstack request type=ticker symbol=%s url=%s", symbol, redactAccessKey(endpoint))
	}
	var resp marketstackTickerResponse
	if err := m.getJSON(ctx, endpoint, nil, &resp); err != nil {
		if m.cache != nil {
			_ = m.cache.setNegative(ctx, cacheKindMarketstackTick, cacheKey, cacheStatusError, cacheTTLNegativeError)
		}
		if m.logger != nil {
			m.logger.Printf("marketstack response type=ticker symbol=%s status=error duration=%s err=%v", symbol, time.Since(start), err)
		}
		return resp, err
	}
	if m.logger != nil {
		m.logger.Printf("marketstack response type=ticker symbol=%s status=ok duration=%s mic=%s", symbol, time.Since(start), resp.StockExchange.MIC)
	}
	if m.cache != nil {
		_ = m.cache.setJSON(ctx, cacheKindMarketstackTick, cacheKey, resp, cacheTTLMarketstackTick)
	}
	return resp, nil
}

// marketstackTickerOverride bypasses /v2/tickers for symbols where Marketstack
// returns incorrect company/exchange metadata (e.g., CRCL resolves to an OTC name).
// We still use Marketstack for pricing, but we override identity data here to keep
// asset keys and reports aligned with the correct exchange.
func marketstackTickerOverride(symbol string) (marketstackTickerResponse, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return marketstackTickerResponse{}, false
	}
	override, ok := marketstackTickerOverrides[normalized]
	if !ok {
		return marketstackTickerResponse{}, false
	}
	if override.Symbol == "" {
		override.Symbol = normalized
	}
	return override, true
}

// NOTE: Add overrides sparingly with verified public listings.
var marketstackTickerOverrides = map[string]marketstackTickerResponse{
	"CRCL": {
		Name:   "Circle Internet Group Inc - Class A",
		Symbol: "CRCL",
		StockExchange: marketstackStockExchange{
			Name: "New York Stock Exchange",
			MIC:  "XNYS",
		},
	},
}

func (m *marketClient) marketstackEOD(ctx context.Context, symbols []string, latest bool) (marketstackEODResponse, error) {
	query := url.Values{}
	query.Set("access_key", m.cfg.MarketstackAccessKey)
	query.Set("symbols", strings.Join(symbols, ","))
	if len(symbols) > 0 {
		limit := len(symbols)
		if limit > 1000 {
			limit = 1000
		}
		query.Set("limit", strconv.Itoa(limit))
	}
	path := "/eod"
	if latest {
		path = "/eod/latest"
	}
	endpoint := fmt.Sprintf("%s%s?%s", marketstackBaseURL, path, query.Encode())
	start := time.Now()
	if m.logger != nil {
		m.logger.Printf("marketstack request type=eod latest=%t symbols=%d url=%s", latest, len(symbols), redactAccessKey(endpoint))
	}
	var resp marketstackEODResponse
	if err := m.getJSON(ctx, endpoint, nil, &resp); err != nil {
		if m.logger != nil {
			m.logger.Printf("marketstack response type=eod latest=%t symbols=%d status=error duration=%s err=%v", latest, len(symbols), time.Since(start), err)
		}
		return resp, err
	}
	if m.logger != nil {
		m.logger.Printf("marketstack response type=eod latest=%t symbols=%d status=ok duration=%s items=%d", latest, len(symbols), time.Since(start), len(resp.Data))
	}
	return resp, nil
}

func (m *marketClient) marketstackEODRange(ctx context.Context, symbols []string, start, end time.Time) (marketstackEODResponse, error) {
	query := url.Values{}
	query.Set("access_key", m.cfg.MarketstackAccessKey)
	query.Set("symbols", strings.Join(symbols, ","))
	query.Set("limit", "1000")
	if !start.IsZero() {
		query.Set("date_from", start.Format("2006-01-02"))
	}
	if !end.IsZero() {
		query.Set("date_to", end.Format("2006-01-02"))
	}
	endpoint := fmt.Sprintf("%s/eod?%s", marketstackBaseURL, query.Encode())
	callStart := time.Now()
	if m.logger != nil {
		m.logger.Printf("marketstack request type=eod_range symbols=%d url=%s", len(symbols), redactAccessKey(endpoint))
	}
	var resp marketstackEODResponse
	if err := m.getJSON(ctx, endpoint, nil, &resp); err != nil {
		if m.logger != nil {
			m.logger.Printf("marketstack response type=eod_range symbols=%d status=error duration=%s err=%v", len(symbols), time.Since(callStart), err)
		}
		return resp, err
	}
	if m.logger != nil {
		m.logger.Printf("marketstack response type=eod_range symbols=%d status=ok duration=%s items=%d", len(symbols), time.Since(callStart), len(resp.Data))
	}
	return resp, nil
}

func redactAccessKey(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	values := parsed.Query()
	if values.Has("access_key") {
		values.Set("access_key", "REDACTED")
		parsed.RawQuery = values.Encode()
	}
	return parsed.String()
}

func (m *marketClient) openExchangeLatest(ctx context.Context) (oerLatestResponse, error) {
	if m == nil {
		return oerLatestResponse{}, fmt.Errorf("market client unavailable")
	}
	db := marketDB(m)
	if db != nil {
		cached, ok, err := loadFXRatesForDate(ctx, db, time.Now().UTC(), "USD")
		if err != nil {
			return oerLatestResponse{}, err
		}
		if ok {
			return cached, nil
		}
	} else if m.redis != nil {
		cacheKey := "cache:oer:latest"
		var cached oerLatestResponse
		if ok, err := m.redis.getJSON(ctx, cacheKey, &cached); ok && err == nil {
			return cached, nil
		}
	}

	query := url.Values{}
	query.Set("app_id", m.cfg.OpenExchangeAppID)
	endpoint := fmt.Sprintf("%s/latest.json?%s", oerBaseURL, query.Encode())
	var resp oerLatestResponse
	if err := m.getJSON(ctx, endpoint, nil, &resp); err != nil {
		return resp, err
	}
	if db != nil {
		_ = storeFXRates(ctx, db, resp, fxRateSourceOpenExchange)
	} else if m.redis != nil {
		_ = m.redis.setJSON(ctx, "cache:oer:latest", resp, time.Hour)
	}
	return resp, nil
}

func (m *marketClient) openExchangeCurrencies(ctx context.Context) (map[string]string, error) {
	cacheKey := "cache:oer:currencies"
	var cached map[string]string
	if ok, err := m.redis.getJSON(ctx, cacheKey, &cached); ok && err == nil {
		return cached, nil
	}
	query := url.Values{}
	query.Set("app_id", m.cfg.OpenExchangeAppID)
	endpoint := fmt.Sprintf("%s/currencies.json?%s", oerBaseURL, query.Encode())
	var resp map[string]string
	if err := m.getJSON(ctx, endpoint, nil, &resp); err != nil {
		return nil, err
	}
	_ = m.redis.setJSON(ctx, cacheKey, resp, 7*24*time.Hour)
	return resp, nil
}

func (m *marketClient) getJSON(ctx context.Context, endpoint string, headers map[string]string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		return err
	}
	return nil
}
