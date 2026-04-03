package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	testUserID                = "usr_test"
	maxPlansPerPortfolio      = 3
	defaultSupportLookback    = 90
	defaultVolatilityLookback = 30
	minPortfolioWeight        = 0.01
	rsiPeriod                 = 14
	bollingerPeriod           = 20
	bollingerStdDev           = 2.0
	binanceKlinesInterval     = "4h"
	binanceKlinesLimit        = 200
)

type userProfile struct {
	Name           string
	RiskTolerance  string
	RiskPreference string
	PainPoints     []string
	Experience     string
	Style          string
	Markets        []string
	Timezone       string
}

type portfolioHolding struct {
	SymbolRaw           string
	Symbol              string
	AssetType           string
	AssetKey            string
	CoinGeckoID         string
	ExchangeMIC         string
	Amount              float64
	ValueUSD            float64
	ValueFromScreenshot *float64
	DisplayCurrency     *string
	PricingSource       string
	ValuationStatus     string
	BalanceType         string
	CurrentPrice        float64
	AvgPrice            *float64
	PNLPercent          *float64
	CostBasisStatus     string
}

type portfolioMetrics struct {
	NetWorthUSD             float64
	PricedValueUSD          float64
	IdleCashUSD             float64
	CashPct                 float64
	TopAssetPct             float64
	PricedCoveragePct       float64
	MetricsIncomplete       bool
	Volatility30dDaily      float64
	Volatility30dAnnualized float64
	MaxDrawdown90d          float64
	AvgPairwiseCorr         float64
	HealthScoreBaseline     int
	VolatilityScoreBaseline int
	CryptoWeight            float64
}

type previewReportPayload struct {
	MetaData             metaDataPayload        `json:"meta_data"`
	ValuationAsOf        string                 `json:"valuation_as_of"`
	MarketDataSnapshotID string                 `json:"market_data_snapshot_id"`
	UserProfile          userProfilePayload     `json:"user_profile"`
	Portfolio            portfolioPayload       `json:"portfolio"`
	ComputedMetrics      computedMetricsPayload `json:"computed_metrics"`
	FixedMetrics         previewFixedMetrics    `json:"fixed_metrics"`
	NetWorthDisplay      float64                `json:"net_worth_display"`
	BaseCurrency         string                 `json:"base_currency"`
	BaseFXRateToUSD      float64                `json:"base_fx_rate_to_usd"`
}

type metaDataPayload struct {
	CalculationID string `json:"calculation_id"`
}

type userProfilePayload struct {
	RiskTolerance  string   `json:"risk_tolerance"`
	RiskPreference string   `json:"risk_preference,omitempty"`
	PainPoints     []string `json:"pain_points"`
	Experience     string   `json:"experience"`
	Style          string   `json:"style"`
	Markets        []string `json:"markets"`
}

type portfolioPayload struct {
	NetWorthUSD float64                   `json:"net_worth_usd"`
	Holdings    []portfolioHoldingPayload `json:"holdings"`
}

type portfolioHoldingPayload struct {
	AssetKey  string  `json:"asset_key"`
	Symbol    string  `json:"symbol"`
	AssetType string  `json:"asset_type"`
	Amount    float64 `json:"amount"`
	ValueUSD  float64 `json:"value_usd"`
}

type computedMetricsPayload struct {
	NetWorthUSD             float64 `json:"net_worth_usd"`
	CashPct                 float64 `json:"cash_pct"`
	TopAssetPct             float64 `json:"top_asset_pct"`
	Volatility30dAnnualized float64 `json:"volatility_30d_annualized"`
	MaxDrawdown90d          float64 `json:"max_drawdown_90d"`
	AvgPairwiseCorr         float64 `json:"avg_pairwise_corr"`
	HealthScoreBaseline     int     `json:"health_score_baseline"`
	VolatilityScoreBaseline int     `json:"volatility_score_baseline"`
	PricedCoveragePct       float64 `json:"priced_coverage_pct"`
	MetricsIncomplete       bool    `json:"metrics_incomplete"`
}

type portfolioFactsPayload struct {
	NetWorthUSD             float64 `json:"net_worth_usd"`
	CashPct                 float64 `json:"cash_pct"`
	TopAssetPct             float64 `json:"top_asset_pct"`
	Volatility30dAnnualized float64 `json:"volatility_30d_annualized"`
	MaxDrawdown90d          float64 `json:"max_drawdown_90d"`
	AvgPairwiseCorr         float64 `json:"avg_pairwise_corr"`
	PricedCoveragePct       float64 `json:"priced_coverage_pct"`
	MetricsIncomplete       bool    `json:"metrics_incomplete"`
}

type paidReportPayload struct {
	UserPortfolio   portfolioPayload      `json:"user_portfolio"`
	UserProfile     userProfilePayload    `json:"user_profile"`
	PreviousTeaser  previewPromptOutput   `json:"previous_teaser"`
	LockedPlans     []lockedPlan          `json:"locked_plans"`
	PortfolioFacts  portfolioFactsPayload `json:"portfolio_facts"`
	FixedMetrics    previewFixedMetrics   `json:"fixed_metrics"`
	NetWorthDisplay float64               `json:"net_worth_display"`
	BaseCurrency    string                `json:"base_currency"`
	BaseFXRateToUSD float64               `json:"base_fx_rate_to_usd"`
}

type lockedPlan struct {
	PlanID       string         `json:"plan_id"`
	StrategyID   string         `json:"strategy_id"`
	AssetType    string         `json:"asset_type"`
	Symbol       string         `json:"symbol"`
	AssetKey     string         `json:"asset_key"`
	LinkedRiskID string         `json:"linked_risk_id"`
	Parameters   map[string]any `json:"parameters,omitempty"`
}

type ohlcPoint struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
}

type pricePoint struct {
	Timestamp int64
	Value     float64
}

type returnPoint struct {
	Timestamp int64
	Value     float64
}

type marketstackPriceQuote struct {
	Close         float64
	PriceCurrency string
	Symbol        string
	Exchange      string
}

type insightItem struct {
	ID              string
	Type            string
	Asset           string
	AssetKey        string
	Timeframe       string
	Severity        string
	TriggerReason   string
	TriggerKey      string
	StrategyID      string
	PlanID          string
	SuggestedAction string
	CreatedAt       time.Time
	ExpiresAt       time.Time
}

type indicatorSnapshot struct {
	Asset     string
	AssetKey  string
	AssetType string
	Interval  string
	Source    string
	RSI       float64
	UpperBand float64
	LowerBand float64
	LastClose float64
}

type indicatorSeries struct {
	Points   []ohlcPoint
	Interval string
	Source   string
}

func TestPortfolioFlowEndToEnd(t *testing.T) {
	geminiKey := requireEnv(t, "GEMINI_API_KEY")
	coingeckoKey := requireEnv(t, "COINGECKO_PRO_API_KEY")

	geminiClient := &http.Client{Timeout: 60 * time.Second}
	apiClient := &http.Client{Timeout: 30 * time.Second}

	portfolios := []struct {
		Name       string
		ImagePaths []string
	}{
		{Name: "portfolio1", ImagePaths: portfolioImagePaths(t, "portfolio1")},
		{Name: "portfolio2", ImagePaths: portfolioImagePaths(t, "portfolio2")},
		{Name: "portfolio3", ImagePaths: portfolioImagePaths(t, "portfolio3")},
		{Name: "portfolio4", ImagePaths: portfolioImagePaths(t, "portfolio4")},
	}

	profiles := []userProfile{
		{
			Name:           "aggressive_swing_speculator",
			RiskTolerance:  "Aggressive",
			RiskPreference: "Speculator",
			PainPoints:     []string{"Bagholder", "FOMO"},
			Experience:     "Intermediate",
			Style:          "Swing Trading",
			Markets:        []string{"Crypto"},
			Timezone:       "America/New_York",
		},
		{
			Name:           "conservative_longterm_yield",
			RiskTolerance:  "Conservative",
			RiskPreference: "Yield Seeker",
			PainPoints:     []string{"Seeking Stable Yield"},
			Experience:     "Beginner",
			Style:          "Long-Term",
			Markets:        []string{"Crypto"},
			Timezone:       "Europe/London",
		},
		{
			Name:           "moderate_day_speculator",
			RiskTolerance:  "Moderate",
			RiskPreference: "Speculator",
			PainPoints:     []string{"Messy Portfolio"},
			Experience:     "Expert",
			Style:          "Day Trading",
			Markets:        []string{"Crypto"},
			Timezone:       "Asia/Singapore",
		},
	}

	outputLanguages := []string{
		"English",
		"Simplified Chinese",
	}

	for _, portfolio := range portfolios {
		portfolio := portfolio
		t.Run(portfolio.Name, func(t *testing.T) {
			debugLogf(t, "E2E portfolio=%s images=%v", portfolio.Name, portfolio.ImagePaths)
			ocr := runOCR(t, geminiClient, geminiKey, portfolio.ImagePaths)
			debugLogOCRImages(t, ocr.Images)

			assets, platformGuess := collectOCRAssets(t, ocr.Images)
			if len(assets) == 0 {
				t.Fatal("expected OCR assets, got empty")
			}

			holdings, seriesByAssetKey, metrics := buildPortfolioState(t, apiClient, coingeckoKey, platformGuess, assets)
			if len(holdings) == 0 {
				t.Fatal("expected holdings after normalization, got empty")
			}
			if metrics.NetWorthUSD <= 0 {
				t.Fatalf("expected positive net worth, got %v", metrics.NetWorthUSD)
			}

			debugLogJSON(t, "E2E holdings post-filter (1% threshold)", holdings)
			debugLogJSON(t, "E2E metrics", metrics)

			for profileIdx, profile := range profiles {
				profile := profile
				for langIdx, outputLanguage := range outputLanguages {
					outputLanguage := outputLanguage
					scenarioID := fmt.Sprintf("%s/%s/%s", portfolio.Name, profile.Name, slugify(outputLanguage))
					scenarioLabel := fmt.Sprintf("%s | %s | %s", portfolio.Name, profile.Name, outputLanguage)
					t.Run(scenarioID, func(t *testing.T) {
						snapshotTime := time.Now().UTC().Truncate(time.Second)
						valuationAsOf := snapshotTime.Format(time.RFC3339)
						marketDataSnapshotID := fmt.Sprintf("snap_%d_%d", snapshotTime.UnixNano(), langIdx)
						calculationID := fmt.Sprintf("calc_e2e_%s_%s_%s_%d", portfolio.Name, profile.Name, slugify(outputLanguage), profileIdx)

						previewInput := buildPreviewPayload(calculationID, valuationAsOf, marketDataSnapshotID, profile, holdings, metrics)
						previewRequest := buildGeminiRequest(t, applyOutputLanguage(PreviewReportPrompt, outputLanguage), previewInput, 0.4, geminiMaxOutputTokens)

						previewResp := callPreviewWithRetries(t, geminiClient, geminiKey, previewRequest, scenarioID, scenarioLabel, previewInput.ComputedMetrics)
						debugLogJSON(t, "E2E preview output "+scenarioID, previewResp)
						debugLogPreviewText(t, scenarioLabel, previewResp)
						assertPreviewResponse(t, previewResp, previewInput)

						lockedPlans := buildLockedPlans(t, profile, holdings, metrics, seriesByAssetKey)
						lockedPlans = assignLinkedRiskIDs(lockedPlans, previewResp.IdentifiedRisks)
						debugLogJSON(t, "E2E locked plans "+scenarioID, lockedPlans)

						paidInput := buildPaidPayload(profile, holdings, previewResp, lockedPlans, metrics)
						paidRequest := buildGeminiRequest(t, applyOutputLanguage(PaidReportPrompt, outputLanguage), paidInput, 0.4, geminiMaxOutputTokens)

						var paidResp paidPromptOutput
						callGeminiJSON(t, geminiClient, geminiKey, paidRequest, &paidResp, "paid-e2e-"+scenarioID)
						normalizePaidStatus(&paidResp)
						mergePaidPlanParameters(&paidResp, lockedPlans)
						debugLogJSON(t, "E2E paid output "+scenarioID, paidResp)
						debugLogPaidText(t, scenarioLabel, paidResp)
						assertPaidResponse(t, paidResp, previewResp, lockedPlans)

						insights, indicators := buildInsights(t, apiClient, holdings, lockedPlans, seriesByAssetKey)
						debugLogInsights(t, scenarioLabel, indicators, insights)
						assertInsights(t, holdings, lockedPlans, indicators, insights)
					})
				}
			}
		})
	}
}

func runOCR(t *testing.T, client *http.Client, apiKey string, imagePaths []string) ocrPromptResponse {
	t.Helper()

	if len(imagePaths) == 0 {
		t.Fatal("expected at least one image path")
	}

	parts := []geminiPart{{Text: buildOCRInput(imagePaths)}}
	for _, imagePath := range imagePaths {
		imageBytes := mustReadFile(t, imagePath)
		encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{
				MimeType: imageMimeType(imagePath),
				Data:     encodedImage,
			},
		})
	}

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: OCRPortfolioPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: parts,
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.0,
			MaxOutputTokens:  geminiMaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	var parsed ocrPromptResponse
	callGeminiJSON(t, client, apiKey, request, &parsed, "OCR-e2e")
	return parsed
}

func buildOCRInput(imagePaths []string) string {
	lines := make([]string, 0, len(imagePaths)+2)
	lines = append(lines, "Input images:")
	for idx, imagePath := range imagePaths {
		lines = append(lines, fmt.Sprintf("- image_id: img_%d (%s)", idx+1, filepath.Base(imagePath)))
	}
	lines = append(lines, "Return strict JSON per the system prompt.")
	return strings.Join(lines, "\n")
}

func slugify(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return "unknown"
	}
	return strings.Join(strings.Fields(trimmed), "_")
}

func collectOCRAssets(t *testing.T, images []ocrImage) ([]ocrAsset, string) {
	t.Helper()

	var assets []ocrAsset
	platformGuess := ""
	for _, image := range images {
		if image.Status != "success" {
			continue
		}
		if platformGuess == "" {
			platformGuess = image.PlatformGuess
		}
		assets = append(assets, image.Assets...)
	}
	return assets, platformGuess
}

func debugLogPreviewText(t *testing.T, label string, preview previewPromptOutput) {
	t.Helper()

	debugLogf(t, "Preview text (%s)", label)
	debugLogf(t, "  Health score: %d (%s) | Volatility score: %d", preview.FixedMetrics.HealthScore, preview.FixedMetrics.HealthStatus, preview.FixedMetrics.VolatilityScore)
	for _, risk := range preview.IdentifiedRisks {
		debugLogf(t, "  %s %s/%s: %s", risk.RiskID, risk.Type, risk.Severity, risk.TeaserText)
	}
	debugLogf(t, "  Projection: %s | %s", preview.LockedProjection.PotentialUpside, preview.LockedProjection.CTA)
}

func debugLogOCRImages(t *testing.T, images []ocrImage) {
	t.Helper()

	for _, image := range images {
		errorReason := "null"
		if image.ErrorReason != nil {
			errorReason = *image.ErrorReason
		}
		debugLogf(t, "OCR image %s status=%s platform=%s error_reason=%s assets=%d", image.ImageID, image.Status, image.PlatformGuess, errorReason, len(image.Assets))
		for _, asset := range image.Assets {
			symbol := "null"
			if asset.Symbol != nil {
				symbol = *asset.Symbol
			}
			displayCurrency := "null"
			if asset.DisplayCurrency != nil {
				displayCurrency = *asset.DisplayCurrency
			}
			value := formatOptionalFloat(asset.ValueFromScreenshot, 2)
			avgPrice := formatOptionalFloat(asset.AvgPrice, priceDecimals(asset.AssetType))
			pnlPercent := formatOptionalFloat(asset.PNLPercent, 2)
			amount := formatFloat(asset.Amount, amountDecimals(asset.AssetType))
			debugLogf(t, "  asset symbol_raw=%s symbol=%s type=%s amount=%s value=%s %s avg_price=%s pnl_percent=%s", asset.SymbolRaw, symbol, asset.AssetType, amount, value, displayCurrency, avgPrice, pnlPercent)
		}
	}
}

func debugLogPaidText(t *testing.T, label string, paid paidPromptOutput) {
	t.Helper()

	debugLogf(t, "Paid text (%s)", label)
	debugLogf(t, "  Health score: %d (%s) | Volatility: %d (%s)", paid.ReportHeader.HealthScore.Value, paid.ReportHeader.HealthScore.Status, paid.ReportHeader.VolatilityDashboard.Value, paid.ReportHeader.VolatilityDashboard.Status)
	for _, insight := range paid.RiskInsights {
		debugLogf(t, "  %s %s/%s: %s", insight.RiskID, insight.Type, insight.Severity, insight.Message)
	}
	for _, plan := range paid.OptimizationPlan {
		debugLogf(t, "  Plan %s %s %s: %s | %s", plan.PlanID, plan.StrategyID, plan.Symbol, plan.Rationale, plan.ExpectedOutcome)
	}
	debugLogf(t, "  Verdict: %s", paid.TheVerdict.ConstructiveComment)
}

func debugLogInsights(t *testing.T, label string, indicators []indicatorSnapshot, insights []insightItem) {
	t.Helper()

	debugLogf(t, "Insights indicators (%s)", label)
	if len(indicators) == 0 {
		debugLogf(t, "  (none)")
	}
	for _, indicator := range indicators {
		debugLogf(
			t,
			"  %s %s %s close=%s rsi=%s bollinger=[%s,%s]",
			indicator.Asset,
			indicator.Interval,
			indicator.Source,
			formatFloat(indicator.LastClose, priceDecimals(indicator.AssetType)),
			formatFloat(indicator.RSI, 2),
			formatFloat(indicator.LowerBand, priceDecimals(indicator.AssetType)),
			formatFloat(indicator.UpperBand, priceDecimals(indicator.AssetType)),
		)
	}

	debugLogf(t, "Insights signals (%s)", label)
	if len(insights) == 0 {
		debugLogf(t, "  (none)")
	}
	for _, item := range insights {
		debugLogf(
			t,
			"  [%s] %s %s %s/%s: %s | %s | trigger_key=%s plan=%s",
			item.Type,
			item.Asset,
			item.AssetKey,
			item.StrategyID,
			item.Severity,
			item.TriggerReason,
			item.SuggestedAction,
			item.TriggerKey,
			item.PlanID,
		)
	}
}

func formatOptionalFloat(value *float64, decimals int) string {
	if value == nil {
		return "null"
	}
	return formatFloat(*value, decimals)
}

func formatFloat(value float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, roundTo(value, decimals))
}

func buildPortfolioState(t *testing.T, client *http.Client, apiKey, platformGuess string, assets []ocrAsset) ([]portfolioHolding, map[string][]ohlcPoint, portfolioMetrics) {
	t.Helper()

	coinList := fetchCoinGeckoList(t, client, apiKey)
	holdings, oerRates := resolveHoldings(t, client, apiKey, platformGuess, assets, coinList)
	holdings = aggregateHoldings(holdings)
	debugLogJSON(t, "E2E holdings pre-filter", holdings)
	holdings = filterLowValueHoldings(holdings)

	seriesByAssetKey := fetchPriceSeries(t, client, apiKey, holdings, oerRates)
	metrics := computePortfolioMetrics(holdings, seriesByAssetKey)
	return holdings, seriesByAssetKey, metrics
}

func fetchCoinGeckoList(t *testing.T, client *http.Client, apiKey string) []coinGeckoCoinListEntry {
	t.Helper()
	query := url.Values{}
	query.Set("include_platform", "false")

	var resp []coinGeckoCoinListEntry
	coinGeckoGet(t, client, apiKey, "/coins/list", query, &resp)
	return resp
}

func resolveHoldings(t *testing.T, client *http.Client, apiKey, platformGuess string, assets []ocrAsset, coinList []coinGeckoCoinListEntry) ([]portfolioHolding, map[string]float64) {
	t.Helper()

	symbolToIDs := make(map[string][]string)
	for _, entry := range coinList {
		if entry.Symbol == "" || entry.ID == "" {
			continue
		}
		key := strings.ToLower(entry.Symbol)
		symbolToIDs[key] = append(symbolToIDs[key], entry.ID)
	}

	platformCategory := platformGuessToCategory(platformGuess)
	stablecoins := stablecoinSet()

	var holdings []portfolioHolding
	for _, asset := range assets {
		if asset.SymbolRaw == "" || asset.Amount == 0 {
			continue
		}

		symbol := ""
		if asset.Symbol != nil && strings.TrimSpace(*asset.Symbol) != "" {
			symbol = *asset.Symbol
		} else if alias, ok := aliasSymbol(asset.SymbolRaw); ok {
			symbol = alias
		} else {
			symbol = asset.SymbolRaw
		}
		symbol = normalizeSymbol(symbol)
		if symbol == "" {
			continue
		}

		assetType := normalizeAssetType(asset.AssetType)

		balanceType := "unknown"
		if platformCategory == "crypto_exchange" || platformCategory == "wallet" {
			if stablecoins[symbol] {
				balanceType = "stablecoin"
			}
			if symbol == "USD" {
				assetType = "forex"
				balanceType = "fiat_cash"
			}
		}

		coinGeckoID := ""
		if assetType == "crypto" {
			coinGeckoID = resolveCoinGeckoID(t, client, apiKey, symbol, symbolToIDs)
		}

		assetKey := ""
		switch assetType {
		case "crypto":
			if coinGeckoID != "" {
				assetKey = "crypto:cg:" + coinGeckoID
			} else {
				symbolRawNormalized := normalizeSymbol(asset.SymbolRaw)
				assetKey = manualAssetKey(testUserID, symbolRawNormalized, platformGuess)
			}
		case "forex":
			assetKey = "forex:fx:" + symbol
		case "stock":
			assetKey = ""
		default:
			symbolRawNormalized := normalizeSymbol(asset.SymbolRaw)
			assetKey = manualAssetKey(testUserID, symbolRawNormalized, platformGuess)
		}

		holding := portfolioHolding{
			SymbolRaw:           asset.SymbolRaw,
			Symbol:              symbol,
			AssetType:           assetType,
			AssetKey:            assetKey,
			CoinGeckoID:         coinGeckoID,
			Amount:              asset.Amount,
			ValueFromScreenshot: asset.ValueFromScreenshot,
			DisplayCurrency:     asset.DisplayCurrency,
			BalanceType:         balanceType,
		}

		if asset.AvgPrice != nil {
			value := *asset.AvgPrice
			holding.AvgPrice = &value
		}
		if asset.PNLPercent != nil {
			value := *asset.PNLPercent / 100.0
			holding.PNLPercent = &value
		}

		holdings = append(holdings, holding)
	}

	stockTickers := resolveMarketstackTickers(t, client, holdings)
	for i := range holdings {
		if holdings[i].AssetType != "stock" {
			continue
		}
		symbolKey := strings.ToUpper(holdings[i].Symbol)
		if ticker, ok := stockTickers[symbolKey]; ok {
			mic := strings.ToUpper(strings.TrimSpace(ticker.StockExchange.MIC))
			holdings[i].ExchangeMIC = mic
			if mic != "" {
				holdings[i].AssetKey = stockAssetKey(mic, holdings[i].Symbol)
			}
		}
		if holdings[i].AssetKey == "" {
			holdings[i].AssetKey = manualAssetKey(testUserID, holdings[i].SymbolRaw, platformGuess)
		}
	}

	priceMap := fetchCoinGeckoPrices(t, client, apiKey, holdings)
	stockPrices := fetchMarketstackPrices(t, client, holdings)
	oerRates := fetchOERRatesIfNeeded(t, client, holdings, stockPrices)

	for i := range holdings {
		applyPricing(&holdings[i], priceMap, stockPrices, oerRates)
		applyCostBasis(&holdings[i])
	}

	return holdings, oerRates
}

func fetchCoinGeckoPrices(t *testing.T, client *http.Client, apiKey string, holdings []portfolioHolding) map[string]coinGeckoSimplePrice {
	t.Helper()

	ids := make([]string, 0, len(holdings))
	seen := make(map[string]struct{})
	for _, holding := range holdings {
		if holding.CoinGeckoID == "" {
			continue
		}
		if _, ok := seen[holding.CoinGeckoID]; ok {
			continue
		}
		seen[holding.CoinGeckoID] = struct{}{}
		ids = append(ids, holding.CoinGeckoID)
	}
	if len(ids) == 0 {
		return map[string]coinGeckoSimplePrice{}
	}

	query := url.Values{}
	query.Set("ids", strings.Join(ids, ","))
	query.Set("vs_currencies", "usd")
	query.Set("include_last_updated_at", "true")

	var resp coinGeckoSimplePriceResponse
	coinGeckoGet(t, client, apiKey, "/simple/price", query, &resp)
	return resp
}

func stockSymbols(holdings []portfolioHolding) []string {
	seen := make(map[string]struct{})
	symbols := make([]string, 0, len(holdings))
	for _, holding := range holdings {
		if holding.AssetType != "stock" {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(holding.Symbol))
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	return symbols
}

func resolveMarketstackTickers(t *testing.T, client *http.Client, holdings []portfolioHolding) map[string]marketstackTickerResponse {
	t.Helper()

	symbols := stockSymbols(holdings)
	if len(symbols) == 0 {
		return nil
	}

	accessKey := requireEnv(t, "MARKETSTACK_ACCESS_KEY")
	resolved := make(map[string]marketstackTickerResponse, len(symbols))
	for _, symbol := range symbols {
		ticker, ok := fetchMarketstackTicker(t, client, accessKey, symbol)
		if !ok {
			continue
		}
		resolved[strings.ToUpper(symbol)] = ticker
	}
	return resolved
}

func fetchMarketstackTicker(t *testing.T, client *http.Client, accessKey, symbol string) (marketstackTickerResponse, bool) {
	t.Helper()

	if override, ok := marketstackTickerOverride(symbol); ok {
		return override, true
	}

	endpoint := fmt.Sprintf("%s/tickers/%s?access_key=%s", marketstackBaseURL, url.PathEscape(symbol), url.QueryEscape(accessKey))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("build marketstack ticker request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("marketstack ticker request failed: %v", err)
	}
	defer resp.Body.Close()
	debugLogf(t, "HTTP %s %s -> %d (%s)", req.Method, sanitizeURL(req.URL), resp.StatusCode, time.Since(start))

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		debugLogf(t, "Marketstack ticker lookup failed for %s: %s", symbol, strings.TrimSpace(string(body)))
		return marketstackTickerResponse{}, false
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("marketstack ticker lookup %s status %d: %s", symbol, resp.StatusCode, string(body))
	}

	var parsed marketstackTickerResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&parsed); err != nil {
		t.Fatalf("decode marketstack ticker %s: %v", symbol, err)
	}
	if parsed.Symbol == "" {
		return marketstackTickerResponse{}, false
	}
	return parsed, true
}

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

func fetchMarketstackPrices(t *testing.T, client *http.Client, holdings []portfolioHolding) map[string]marketstackPriceQuote {
	t.Helper()

	symbols := stockSymbols(holdings)
	if len(symbols) == 0 {
		return nil
	}

	accessKey := requireEnv(t, "MARKETSTACK_ACCESS_KEY")
	quotesBySymbol := make(map[string]marketstackPriceQuote, len(symbols))
	for _, symbol := range symbols {
		query := url.Values{}
		query.Set("access_key", accessKey)
		query.Set("symbols", symbol)
		query.Set("limit", "1")

		endpoint := fmt.Sprintf("%s/eod/latest?%s", marketstackBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build marketstack latest request: %v", err)
		}
		req.Header.Set("Accept", "application/json")

		var resp marketstackEODResponse
		getJSON(t, client, req, &resp)
		if len(resp.Data) == 0 {
			debugLogf(t, "Marketstack latest missing for %s", symbol)
			continue
		}
		bar := resp.Data[0]
		priceCurrency := strings.ToUpper(strings.TrimSpace(bar.PriceCurrency))
		if priceCurrency == "" {
			priceCurrency = "USD"
		}
		quotesBySymbol[strings.ToUpper(symbol)] = marketstackPriceQuote{
			Close:         bar.Close,
			PriceCurrency: priceCurrency,
			Symbol:        bar.Symbol,
			Exchange:      bar.Exchange,
		}
	}

	prices := make(map[string]marketstackPriceQuote)
	for _, holding := range holdings {
		if holding.AssetType != "stock" {
			continue
		}
		symbolKey := strings.ToUpper(holding.Symbol)
		if quote, ok := quotesBySymbol[symbolKey]; ok {
			prices[holding.AssetKey] = quote
		}
	}

	return prices
}

func fetchOERRatesIfNeeded(t *testing.T, client *http.Client, holdings []portfolioHolding, stockPrices map[string]marketstackPriceQuote) map[string]float64 {
	t.Helper()

	needRates := false
	for _, holding := range holdings {
		if holding.AssetType == "forex" && !strings.EqualFold(holding.Symbol, "USD") {
			needRates = true
			break
		}
		if holding.ValueFromScreenshot == nil || holding.DisplayCurrency == nil {
			continue
		}
		if strings.EqualFold(*holding.DisplayCurrency, "USD") {
			continue
		}
		needRates = true
		break
	}

	if !needRates {
		for _, quote := range stockPrices {
			if quote.PriceCurrency == "" || strings.EqualFold(quote.PriceCurrency, "USD") {
				continue
			}
			needRates = true
			break
		}
	}

	if !needRates {
		return nil
	}

	appID := requireEnv(t, "OPEN_EXCHANGE_APP_ID")
	query := url.Values{}
	query.Set("app_id", appID)

	endpoint := fmt.Sprintf("%s/latest.json?%s", oerBaseURL, query.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("build OER request: %v", err)
	}

	var resp oerLatestResponse
	getJSON(t, client, req, &resp)
	return resp.Rates
}

func applyPricing(holding *portfolioHolding, priceMap map[string]coinGeckoSimplePrice, stockPrices map[string]marketstackPriceQuote, oerRates map[string]float64) {
	if holding.AssetType == "crypto" && holding.BalanceType == "stablecoin" && strings.HasPrefix(holding.AssetKey, "crypto:cg:") {
		holding.CurrentPrice = roundTo(1, 8)
		holding.ValueUSD = roundTo(holding.Amount, 2)
		holding.ValuationStatus = "priced"
		holding.PricingSource = "COINGECKO"
		return
	}

	if holding.AssetType == "crypto" && holding.CoinGeckoID != "" {
		if price, ok := priceMap[holding.CoinGeckoID]; ok && price.USD > 0 {
			holding.CurrentPrice = roundTo(price.USD, 8)
			holding.ValueUSD = roundTo(holding.CurrentPrice*holding.Amount, 2)
			holding.ValuationStatus = "priced"
			holding.PricingSource = "COINGECKO"
			return
		}
	}

	if holding.AssetType == "stock" {
		if quote, ok := stockPrices[holding.AssetKey]; ok && quote.Close > 0 {
			price := quote.Close
			currency := strings.ToUpper(strings.TrimSpace(quote.PriceCurrency))
			if currency == "" {
				currency = "USD"
			}
			if currency != "USD" {
				if oerRates != nil {
					if rate, ok := oerRates[currency]; ok && rate > 0 {
						price = price / rate
					} else {
						price = 0
					}
				} else {
					price = 0
				}
			}
			if price > 0 {
				holding.CurrentPrice = roundTo(price, 8)
				holding.ValueUSD = roundTo(holding.CurrentPrice*holding.Amount, 2)
				holding.ValuationStatus = "priced"
				holding.PricingSource = "MARKETSTACK"
				return
			}
		}
	}

	if holding.AssetType == "forex" {
		if holding.Symbol == "USD" {
			holding.CurrentPrice = 1
			holding.ValueUSD = roundTo(holding.Amount, 2)
			holding.ValuationStatus = "priced"
			holding.PricingSource = "OER"
			return
		}
		if oerRates != nil {
			if rate, ok := oerRates[strings.ToUpper(holding.Symbol)]; ok && rate > 0 {
				holding.CurrentPrice = roundTo(1/rate, 8)
				holding.ValueUSD = roundTo(holding.Amount*holding.CurrentPrice, 2)
				holding.ValuationStatus = "priced"
				holding.PricingSource = "OER"
				return
			}
		}
	}

	if holding.ValueFromScreenshot != nil {
		valueUSD := *holding.ValueFromScreenshot
		currency := "USD"
		if holding.DisplayCurrency != nil {
			currency = strings.ToUpper(strings.TrimSpace(*holding.DisplayCurrency))
		}
		if currency != "" && currency != "USD" && oerRates != nil {
			if rate, ok := oerRates[currency]; ok && rate > 0 {
				valueUSD = valueUSD / rate
			}
		}
		holding.ValueUSD = roundTo(valueUSD, 2)
		holding.ValuationStatus = "user_provided"
		holding.PricingSource = "USER_PROVIDED"
		return
	}

	holding.ValuationStatus = "unpriced"
}

func applyCostBasis(holding *portfolioHolding) {
	if holding.ValuationStatus != "priced" || holding.CurrentPrice <= 0 {
		holding.CostBasisStatus = "unknown"
		return
	}

	if holding.AvgPrice != nil && *holding.AvgPrice > 0 {
		pnl := (holding.CurrentPrice - *holding.AvgPrice) / *holding.AvgPrice
		holding.PNLPercent = &pnl
		holding.CostBasisStatus = "provided"
		return
	}

	if holding.PNLPercent != nil {
		denom := 1 + *holding.PNLPercent
		if denom > 0 {
			avg := holding.CurrentPrice / denom
			avg = roundTo(avg, 8)
			holding.AvgPrice = &avg
			holding.CostBasisStatus = "provided"
			return
		}
	}

	holding.CostBasisStatus = "unknown"
}

func aggregateHoldings(holdings []portfolioHolding) []portfolioHolding {
	byKey := make(map[string]*portfolioHolding)
	for _, holding := range holdings {
		key := holding.AssetKey
		if key == "" {
			continue
		}
		if existing, ok := byKey[key]; ok {
			combinedAmount := existing.Amount + holding.Amount
			if combinedAmount > 0 {
				if existing.AvgPrice != nil && holding.AvgPrice != nil {
					weighted := (*existing.AvgPrice*existing.Amount + *holding.AvgPrice*holding.Amount) / combinedAmount
					weighted = roundTo(weighted, 8)
					existing.AvgPrice = &weighted
				} else if existing.AvgPrice == nil && holding.AvgPrice != nil {
					value := *holding.AvgPrice
					existing.AvgPrice = &value
				}
			}
			existing.Amount = combinedAmount
			existing.ValueUSD = roundTo(existing.ValueUSD+holding.ValueUSD, 2)
			if existing.ValuationStatus == "unpriced" && holding.ValuationStatus != "unpriced" {
				existing.ValuationStatus = holding.ValuationStatus
				existing.PricingSource = holding.PricingSource
				existing.CurrentPrice = holding.CurrentPrice
				existing.CoinGeckoID = holding.CoinGeckoID
			}
			existing.CostBasisStatus = mergeCostBasis(existing.CostBasisStatus, holding.CostBasisStatus)
			continue
		}

		copy := holding
		byKey[key] = &copy
	}

	result := make([]portfolioHolding, 0, len(byKey))
	for _, holding := range byKey {
		if holding.AvgPrice != nil && holding.CurrentPrice > 0 {
			pnl := (holding.CurrentPrice - *holding.AvgPrice) / *holding.AvgPrice
			holding.PNLPercent = &pnl
		}
		result = append(result, *holding)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ValueUSD > result[j].ValueUSD
	})
	return result
}

func filterLowValueHoldings(holdings []portfolioHolding) []portfolioHolding {
	total := 0.0
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" && holding.ValuationStatus != "user_provided" {
			continue
		}
		total += holding.ValueUSD
	}
	if total <= 0 {
		return holdings
	}

	threshold := total * minPortfolioWeight
	filtered := make([]portfolioHolding, 0, len(holdings))
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" && holding.ValuationStatus != "user_provided" {
			filtered = append(filtered, holding)
			continue
		}
		if holding.ValueUSD >= threshold {
			filtered = append(filtered, holding)
		}
	}
	return filtered
}

func mergeCostBasis(a, b string) string {
	if a == "provided" || b == "provided" {
		return "provided"
	}
	return "unknown"
}

func fetchPriceSeries(t *testing.T, client *http.Client, apiKey string, holdings []portfolioHolding, oerRates map[string]float64) map[string][]ohlcPoint {
	t.Helper()

	seriesByID := make(map[string][]ohlcPoint)
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" || holding.AssetType != "crypto" || holding.CoinGeckoID == "" {
			continue
		}
		if _, ok := seriesByID[holding.CoinGeckoID]; ok {
			continue
		}
		points := fetchCoinGeckoOHLC(t, client, apiKey, holding.CoinGeckoID, defaultSupportLookback)
		seriesByID[holding.CoinGeckoID] = points
	}

	seriesByAssetKey := make(map[string][]ohlcPoint)
	for _, holding := range holdings {
		if holding.CoinGeckoID == "" {
			continue
		}
		if points, ok := seriesByID[holding.CoinGeckoID]; ok {
			seriesByAssetKey[holding.AssetKey] = points
		}
	}

	stockSeries := fetchMarketstackSeries(t, client, holdings, oerRates)
	for assetKey, points := range stockSeries {
		seriesByAssetKey[assetKey] = points
	}

	return seriesByAssetKey
}

func fetchMarketstackSeries(t *testing.T, client *http.Client, holdings []portfolioHolding, oerRates map[string]float64) map[string][]ohlcPoint {
	t.Helper()

	symbols := stockSymbols(holdings)
	if len(symbols) == 0 {
		return nil
	}

	accessKey := requireEnv(t, "MARKETSTACK_ACCESS_KEY")
	seriesBySymbol := make(map[string][]ohlcPoint, len(symbols))

	for _, symbol := range symbols {
		query := url.Values{}
		query.Set("access_key", accessKey)
		query.Set("symbols", symbol)
		query.Set("limit", fmt.Sprintf("%d", defaultSupportLookback))

		endpoint := fmt.Sprintf("%s/eod?%s", marketstackBaseURL, query.Encode())
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			t.Fatalf("build marketstack eod request: %v", err)
		}
		req.Header.Set("Accept", "application/json")

		var resp marketstackEODResponse
		getJSON(t, client, req, &resp)
		if len(resp.Data) == 0 {
			debugLogf(t, "Marketstack EOD empty for %s", symbol)
			continue
		}

		points := make([]ohlcPoint, 0, len(resp.Data))
		currency := ""
		rate := 1.0
		for _, bar := range resp.Data {
			if currency == "" {
				currency = strings.ToUpper(strings.TrimSpace(bar.PriceCurrency))
				if currency == "" {
					currency = "USD"
				}
				if currency != "USD" {
					if oerRates == nil {
						debugLogf(t, "Missing OER rates for %s EOD currency %s", symbol, currency)
						points = nil
						break
					}
					var ok bool
					rate, ok = oerRates[currency]
					if !ok || rate <= 0 {
						debugLogf(t, "Missing OER rate for %s currency %s", symbol, currency)
						points = nil
						break
					}
				}
			}
			parsed, ok := parseMarketstackTime(bar.Date)
			if !ok {
				continue
			}

			points = append(points, ohlcPoint{
				Timestamp: parsed.UnixMilli(),
				Open:      bar.Open / rate,
				High:      bar.High / rate,
				Low:       bar.Low / rate,
				Close:     bar.Close / rate,
			})
		}

		if len(points) == 0 {
			continue
		}
		sort.Slice(points, func(i, j int) bool {
			return points[i].Timestamp < points[j].Timestamp
		})
		if len(points) > defaultSupportLookback {
			points = points[len(points)-defaultSupportLookback:]
		}
		seriesBySymbol[strings.ToUpper(symbol)] = points
	}

	seriesByAssetKey := make(map[string][]ohlcPoint)
	for _, holding := range holdings {
		if holding.AssetType != "stock" {
			continue
		}
		symbolKey := strings.ToUpper(holding.Symbol)
		if points, ok := seriesBySymbol[symbolKey]; ok {
			seriesByAssetKey[holding.AssetKey] = points
		}
	}

	return seriesByAssetKey
}

func parseMarketstackTime(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-0700",
		"2006-01-02",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func fetchCoinGeckoOHLC(t *testing.T, client *http.Client, apiKey, coinID string, days int) []ohlcPoint {
	t.Helper()
	query := url.Values{}
	query.Set("vs_currency", "usd")
	now := time.Now().UTC()
	query.Set("from", strconv.FormatInt(now.Add(-time.Duration(days)*24*time.Hour).Unix(), 10))
	query.Set("to", strconv.FormatInt(now.Unix(), 10))

	var resp coinGeckoMarketChartRangeResponse
	coinGeckoGet(t, client, apiKey, fmt.Sprintf("/coins/%s/market_chart/range", coinID), query, &resp)
	if len(resp.Prices) == 0 {
		debugLogf(t, "CoinGecko market chart range empty for %s", coinID)
		return nil
	}
	daily, err := deriveDailyOHLCV(resp.Prices, resp.TotalVolumes)
	if err != nil {
		t.Fatalf("derive daily OHLCV for %s: %v", coinID, err)
	}
	points := make([]ohlcPoint, 0, len(daily))
	for _, day := range daily {
		points = append(points, ohlcPoint{
			Timestamp: day.DayStart.UnixMilli(),
			Open:      day.Open,
			High:      day.High,
			Low:       day.Low,
			Close:     day.Close,
		})
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp < points[j].Timestamp
	})
	if len(points) > days {
		points = points[len(points)-days:]
	}
	return points
}

func computePortfolioMetrics(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) portfolioMetrics {
	netWorth := 0.0
	pricedValue := 0.0
	idleCash := 0.0
	topAssetValue := 0.0

	for _, holding := range holdings {
		if holding.ValuationStatus == "priced" || holding.ValuationStatus == "user_provided" {
			netWorth += holding.ValueUSD
			if holding.ValueUSD > topAssetValue {
				topAssetValue = holding.ValueUSD
			}
		}
		if holding.ValuationStatus == "priced" {
			pricedValue += holding.ValueUSD
			if holding.BalanceType == "fiat_cash" || holding.BalanceType == "stablecoin" {
				idleCash += holding.ValueUSD
			}
		}
	}

	netWorth = roundTo(netWorth, 2)
	pricedValue = roundTo(pricedValue, 2)
	idleCash = roundTo(idleCash, 2)

	metrics := portfolioMetrics{
		NetWorthUSD:    netWorth,
		PricedValueUSD: pricedValue,
		IdleCashUSD:    idleCash,
	}

	if netWorth > 0 {
		metrics.CashPct = idleCash / netWorth
		metrics.TopAssetPct = topAssetValue / netWorth
		metrics.PricedCoveragePct = pricedValue / netWorth
	}
	metrics.MetricsIncomplete = metrics.PricedCoveragePct < 0.60

	metrics.CryptoWeight = computeCryptoWeight(holdings, pricedValue)
	annualization := math.Sqrt(252)
	if metrics.CryptoWeight >= 0.50 {
		annualization = math.Sqrt(365)
	}

	eligibleHoldings := filterHoldingsForMetrics(holdings, seriesByAssetKey)
	portfolioSeries := buildPortfolioSeries(eligibleHoldings, seriesByAssetKey)
	if len(portfolioSeries) >= 2 {
		returns := logReturns(portfolioSeries)
		metrics.Volatility30dDaily = stddev(lastNReturns(returns, defaultVolatilityLookback))
		metrics.Volatility30dAnnualized = metrics.Volatility30dDaily * annualization
		metrics.MaxDrawdown90d = maxDrawdown(portfolioSeries)
		metrics.AvgPairwiseCorr = avgPairwiseCorrelation(eligibleHoldings, seriesByAssetKey)
	} else {
		metrics.Volatility30dDaily = 0.04
		metrics.Volatility30dAnnualized = 0.04 * annualization
		metrics.MaxDrawdown90d = 0.10
		metrics.AvgPairwiseCorr = 0.30
	}

	metrics.HealthScoreBaseline = computeHealthScoreBaseline(metrics)
	metrics.VolatilityScoreBaseline = int(math.Round(clamp(metrics.Volatility30dAnnualized*100, 0, 100)))
	return metrics
}

func computeCryptoWeight(holdings []portfolioHolding, pricedValue float64) float64 {
	if pricedValue == 0 {
		return 0
	}
	cryptoValue := 0.0
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType == "crypto" {
			cryptoValue += holding.ValueUSD
		}
	}
	return cryptoValue / pricedValue
}

func filterHoldingsForMetrics(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) []portfolioHolding {
	eligible := make([]portfolioHolding, 0, len(holdings))
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		points := seriesByAssetKey[holding.AssetKey]
		if len(points) < 20 {
			continue
		}
		eligible = append(eligible, holding)
	}
	return eligible
}

func buildPortfolioSeries(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) []pricePoint {
	if len(holdings) == 0 {
		return nil
	}

	timestampCounts := make(map[int64]int)
	for _, holding := range holdings {
		points := seriesByAssetKey[holding.AssetKey]
		for _, point := range points {
			timestampCounts[point.Timestamp]++
		}
	}

	commonTimestamps := make([]int64, 0, len(timestampCounts))
	for ts, count := range timestampCounts {
		if count == len(holdings) {
			commonTimestamps = append(commonTimestamps, ts)
		}
	}
	sort.Slice(commonTimestamps, func(i, j int) bool {
		return commonTimestamps[i] < commonTimestamps[j]
	})

	if len(commonTimestamps) > defaultSupportLookback {
		commonTimestamps = commonTimestamps[len(commonTimestamps)-defaultSupportLookback:]
	}

	series := make([]pricePoint, 0, len(commonTimestamps))
	for _, ts := range commonTimestamps {
		value := 0.0
		for _, holding := range holdings {
			points := seriesByAssetKey[holding.AssetKey]
			closeValue := findClose(points, ts)
			if closeValue <= 0 {
				value = 0
				break
			}
			value += holding.Amount * closeValue
		}
		if value > 0 {
			series = append(series, pricePoint{Timestamp: ts, Value: value})
		}
	}
	return series
}

func findClose(points []ohlcPoint, timestamp int64) float64 {
	for _, point := range points {
		if point.Timestamp == timestamp {
			return point.Close
		}
	}
	return 0
}

func logReturns(series []pricePoint) []returnPoint {
	if len(series) < 2 {
		return nil
	}
	returns := make([]returnPoint, 0, len(series)-1)
	for i := 1; i < len(series); i++ {
		prev := series[i-1]
		curr := series[i]
		if prev.Value <= 0 || curr.Value <= 0 {
			continue
		}
		value := math.Log(curr.Value / prev.Value)
		returns = append(returns, returnPoint{Timestamp: curr.Timestamp, Value: value})
	}
	return returns
}

func lastNReturns(returns []returnPoint, n int) []float64 {
	if len(returns) == 0 {
		return nil
	}
	if len(returns) > n {
		returns = returns[len(returns)-n:]
	}
	values := make([]float64, 0, len(returns))
	for _, point := range returns {
		values = append(values, point.Value)
	}
	return values
}

func stddev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	variance := 0.0
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func maxDrawdown(series []pricePoint) float64 {
	if len(series) == 0 {
		return 0
	}
	peak := series[0].Value
	maxDrawdown := 0.0
	for _, point := range series[1:] {
		if point.Value > peak {
			peak = point.Value
		}
		if peak > 0 {
			drawdown := (peak - point.Value) / peak
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}
	}
	return maxDrawdown
}

func avgPairwiseCorrelation(holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) float64 {
	top := topHoldingsByValue(holdings, 5)
	if len(top) < 2 {
		return 0.30
	}

	returnsByKey := make(map[string]map[int64]float64)
	for _, holding := range top {
		points := seriesByAssetKey[holding.AssetKey]
		returnsByKey[holding.AssetKey] = returnsByTimestamp(points)
	}

	sum := 0.0
	count := 0
	for i := 0; i < len(top); i++ {
		for j := i + 1; j < len(top); j++ {
			r1 := returnsByKey[top[i].AssetKey]
			r2 := returnsByKey[top[j].AssetKey]
			xs, ys := overlappingReturns(r1, r2)
			if len(xs) < 20 {
				continue
			}
			sum += pearsonCorrelation(xs, ys)
			count++
		}
	}

	if count == 0 {
		return 0.30
	}
	return sum / float64(count)
}

func returnsByTimestamp(points []ohlcPoint) map[int64]float64 {
	series := make([]pricePoint, 0, len(points))
	for _, point := range points {
		series = append(series, pricePoint{Timestamp: point.Timestamp, Value: point.Close})
	}
	sort.Slice(series, func(i, j int) bool {
		return series[i].Timestamp < series[j].Timestamp
	})

	result := make(map[int64]float64)
	for i := 1; i < len(series); i++ {
		prev := series[i-1]
		curr := series[i]
		if prev.Value <= 0 || curr.Value <= 0 {
			continue
		}
		result[curr.Timestamp] = math.Log(curr.Value / prev.Value)
	}
	return result
}

func overlappingReturns(a, b map[int64]float64) ([]float64, []float64) {
	if len(a) == 0 || len(b) == 0 {
		return nil, nil
	}

	xs := make([]float64, 0, len(a))
	ys := make([]float64, 0, len(a))
	for ts, val := range a {
		if other, ok := b[ts]; ok {
			xs = append(xs, val)
			ys = append(ys, other)
		}
	}
	return xs, ys
}

func pearsonCorrelation(xs, ys []float64) float64 {
	if len(xs) == 0 || len(xs) != len(ys) {
		return 0
	}
	meanX := mean(xs)
	meanY := mean(ys)
	var num, denomX, denomY float64
	for i := range xs {
		dx := xs[i] - meanX
		dy := ys[i] - meanY
		num += dx * dy
		denomX += dx * dx
		denomY += dy * dy
	}
	if denomX == 0 || denomY == 0 {
		return 0
	}
	return num / math.Sqrt(denomX*denomY)
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func topHoldingsByValue(holdings []portfolioHolding, n int) []portfolioHolding {
	candidates := make([]portfolioHolding, 0, len(holdings))
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		candidates = append(candidates, holding)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ValueUSD > candidates[j].ValueUSD
	})
	if len(candidates) > n {
		candidates = candidates[:n]
	}
	return candidates
}

func computeHealthScoreBaseline(metrics portfolioMetrics) int {
	concentrationPenalty := clamp((metrics.TopAssetPct-0.20)*100, 0, 40)
	cashPenalty := clamp((0.05-metrics.CashPct)*200, 0, 20)
	volPenalty := clamp(metrics.Volatility30dAnnualized*100*0.4, 0, 25)
	drawdownPenalty := clamp(metrics.MaxDrawdown90d*100*0.4, 0, 25)
	corrPenalty := clamp((metrics.AvgPairwiseCorr-0.30)*50, 0, 15)

	score := 100 - (concentrationPenalty + cashPenalty + volPenalty + drawdownPenalty + corrPenalty)
	score = clamp(score, 0, 100)
	if metrics.TopAssetPct >= 0.70 {
		score = math.Min(score, 45)
	} else if metrics.TopAssetPct >= 0.50 {
		score = math.Min(score, 49)
	}
	return int(math.Round(score))
}

func buildPreviewPayload(calculationID, valuationAsOf, marketDataSnapshotID string, profile userProfile, holdings []portfolioHolding, metrics portfolioMetrics) previewReportPayload {
	fixed := previewFixedMetrics{
		NetWorthUSD:     metrics.NetWorthUSD,
		HealthScore:     metrics.HealthScoreBaseline,
		HealthStatus:    healthStatusFromScore(metrics.HealthScoreBaseline),
		VolatilityScore: metrics.VolatilityScoreBaseline,
	}
	return previewReportPayload{
		MetaData: metaDataPayload{
			CalculationID: calculationID,
		},
		ValuationAsOf:        valuationAsOf,
		MarketDataSnapshotID: marketDataSnapshotID,
		UserProfile: userProfilePayload{
			RiskTolerance:  profile.RiskTolerance,
			RiskPreference: profile.RiskPreference,
			PainPoints:     profile.PainPoints,
			Experience:     profile.Experience,
			Style:          profile.Style,
			Markets:        profile.Markets,
		},
		Portfolio: portfolioPayload{
			NetWorthUSD: metrics.NetWorthUSD,
			Holdings:    buildPortfolioHoldingsPayload(holdings),
		},
		ComputedMetrics: computedMetricsPayload{
			NetWorthUSD:             metrics.NetWorthUSD,
			CashPct:                 roundTo(metrics.CashPct, 4),
			TopAssetPct:             roundTo(metrics.TopAssetPct, 4),
			Volatility30dAnnualized: roundTo(metrics.Volatility30dAnnualized, 4),
			MaxDrawdown90d:          roundTo(metrics.MaxDrawdown90d, 4),
			AvgPairwiseCorr:         roundTo(metrics.AvgPairwiseCorr, 4),
			HealthScoreBaseline:     metrics.HealthScoreBaseline,
			VolatilityScoreBaseline: metrics.VolatilityScoreBaseline,
			PricedCoveragePct:       roundTo(metrics.PricedCoveragePct, 4),
			MetricsIncomplete:       metrics.MetricsIncomplete,
		},
		FixedMetrics:    fixed,
		NetWorthDisplay: metrics.NetWorthUSD,
		BaseCurrency:    "USD",
		BaseFXRateToUSD: 1,
	}
}

func buildPortfolioHoldingsPayload(holdings []portfolioHolding) []portfolioHoldingPayload {
	result := make([]portfolioHoldingPayload, 0, len(holdings))
	for _, holding := range holdings {
		if holding.ValuationStatus == "unpriced" {
			continue
		}
		result = append(result, portfolioHoldingPayload{
			AssetKey:  holding.AssetKey,
			Symbol:    holding.Symbol,
			AssetType: holding.AssetType,
			Amount:    roundTo(holding.Amount, amountDecimals(holding.AssetType)),
			ValueUSD:  roundTo(holding.ValueUSD, 2),
		})
	}
	return result
}

func buildPaidPayload(profile userProfile, holdings []portfolioHolding, preview previewPromptOutput, plans []lockedPlan, metrics portfolioMetrics) paidReportPayload {
	return paidReportPayload{
		UserPortfolio: portfolioPayload{
			NetWorthUSD: metrics.NetWorthUSD,
			Holdings:    buildPortfolioHoldingsPayload(holdings),
		},
		UserProfile: userProfilePayload{
			RiskTolerance:  profile.RiskTolerance,
			RiskPreference: profile.RiskPreference,
			Experience:     profile.Experience,
			Style:          profile.Style,
			Markets:        profile.Markets,
		},
		PreviousTeaser: preview,
		LockedPlans:    plans,
		PortfolioFacts: portfolioFactsPayload{
			NetWorthUSD:             metrics.NetWorthUSD,
			CashPct:                 roundTo(metrics.CashPct, 4),
			TopAssetPct:             roundTo(metrics.TopAssetPct, 4),
			Volatility30dAnnualized: roundTo(metrics.Volatility30dAnnualized, 4),
			MaxDrawdown90d:          roundTo(metrics.MaxDrawdown90d, 4),
			AvgPairwiseCorr:         roundTo(metrics.AvgPairwiseCorr, 4),
			PricedCoveragePct:       roundTo(metrics.PricedCoveragePct, 4),
			MetricsIncomplete:       metrics.MetricsIncomplete,
		},
		FixedMetrics:    preview.FixedMetrics,
		NetWorthDisplay: preview.NetWorthDisplay,
		BaseCurrency:    preview.BaseCurrency,
		BaseFXRateToUSD: preview.BaseFXRateToUSD,
	}
}

func buildGeminiRequest(t *testing.T, systemPrompt string, payload any, temperature float64, maxTokens int) geminiRequest {
	t.Helper()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal gemini payload: %v", err)
	}
	return geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: string(payloadBytes)}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      temperature,
			MaxOutputTokens:  maxTokens,
			ResponseMimeType: "application/json",
		},
	}
}

func callPreviewWithRetries(t *testing.T, client *http.Client, apiKey string, request geminiRequest, scenarioID, scenarioLabel string, metrics computedMetricsPayload) previewPromptOutput {
	t.Helper()

	const maxAttempts = 3
	var parsed previewPromptOutput
	var violations []string
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		callGeminiJSON(t, client, apiKey, request, &parsed, "preview-e2e-"+scenarioID)
		normalizePreviewStatus(&parsed)
		violations = collectRiskViolations(parsed.IdentifiedRisks, metrics.MetricsIncomplete)
		if len(violations) == 0 {
			debugLogf(t, "Preview risk validation ok (%s) attempt=%d", scenarioLabel, attempt)
			return parsed
		}
		debugLogf(t, "Preview risk validation failed (%s) attempt=%d violations=%s", scenarioLabel, attempt, strings.Join(violations, "; "))
		if attempt < maxAttempts {
			retryPrompt := fmt.Sprintf("%s\n\n### Validation Fix\nYour previous response violated: %s. Regenerate the full JSON and fix all issues.", request.SystemInstruction.Parts[0].Text, strings.Join(violations, "; "))
			request.SystemInstruction = &geminiSystemInstruction{Parts: []geminiPart{{Text: retryPrompt}}}
		}
	}
	return parsed
}

func assertPreviewResponse(t *testing.T, preview previewPromptOutput, input previewReportPayload) {
	t.Helper()

	if preview.MetaData.CalculationID == "" {
		t.Fatal("preview response missing calculation_id")
	}
	if preview.MetaData.CalculationID != input.MetaData.CalculationID {
		t.Fatalf("calculation_id mismatch: got %q want %q", preview.MetaData.CalculationID, input.MetaData.CalculationID)
	}
	if preview.ValuationAsOf == "" || preview.ValuationAsOf != input.ValuationAsOf {
		t.Fatalf("valuation_as_of mismatch: got %q want %q", preview.ValuationAsOf, input.ValuationAsOf)
	}
	if preview.MarketDataSnapshotID == "" || preview.MarketDataSnapshotID != input.MarketDataSnapshotID {
		t.Fatalf("market_data_snapshot_id mismatch: got %q want %q", preview.MarketDataSnapshotID, input.MarketDataSnapshotID)
	}
	if !floatEquals(preview.FixedMetrics.NetWorthUSD, input.FixedMetrics.NetWorthUSD) {
		t.Fatalf("preview net_worth_usd mismatch: got %v want %v", preview.FixedMetrics.NetWorthUSD, input.FixedMetrics.NetWorthUSD)
	}
	if preview.FixedMetrics.HealthScore != input.FixedMetrics.HealthScore {
		t.Fatalf("health_score mismatch: got %d want %d", preview.FixedMetrics.HealthScore, input.FixedMetrics.HealthScore)
	}
	if preview.FixedMetrics.HealthStatus != input.FixedMetrics.HealthStatus {
		t.Fatalf("health_status mismatch: got %q want %q", preview.FixedMetrics.HealthStatus, input.FixedMetrics.HealthStatus)
	}
	if preview.FixedMetrics.VolatilityScore != input.FixedMetrics.VolatilityScore {
		t.Fatalf("volatility_score mismatch: got %d want %d", preview.FixedMetrics.VolatilityScore, input.FixedMetrics.VolatilityScore)
	}
	if !floatEquals(preview.NetWorthDisplay, input.NetWorthDisplay) {
		t.Fatalf("net_worth_display mismatch: got %v want %v", preview.NetWorthDisplay, input.NetWorthDisplay)
	}
	if preview.BaseCurrency != input.BaseCurrency {
		t.Fatalf("base_currency mismatch: got %q want %q", preview.BaseCurrency, input.BaseCurrency)
	}
	if !floatEquals(preview.BaseFXRateToUSD, input.BaseFXRateToUSD) {
		t.Fatalf("base_fx_rate_to_usd mismatch: got %v want %v", preview.BaseFXRateToUSD, input.BaseFXRateToUSD)
	}

	if preview.FixedMetrics.HealthStatus == "" {
		t.Fatalf("health_status missing: %#v", preview.FixedMetrics)
	}
	if len(preview.IdentifiedRisks) != 3 {
		t.Fatalf("expected 3 identified_risks, got %d", len(preview.IdentifiedRisks))
	}
	validateRisks(t, preview.IdentifiedRisks, input.ComputedMetrics.MetricsIncomplete)

	if preview.LockedProjection.PotentialUpside == "" || preview.LockedProjection.CTA == "" {
		t.Fatalf("locked_projection missing fields: %#v", preview.LockedProjection)
	}
}

func assertPaidResponse(t *testing.T, paid paidPromptOutput, preview previewPromptOutput, plans []lockedPlan) {
	t.Helper()

	if paid.MetaData.CalculationID != preview.MetaData.CalculationID {
		t.Fatalf("paid calculation_id mismatch: got %q want %q", paid.MetaData.CalculationID, preview.MetaData.CalculationID)
	}
	if paid.ValuationAsOf == "" || paid.ValuationAsOf != preview.ValuationAsOf {
		t.Fatalf("paid valuation_as_of mismatch: got %q want %q", paid.ValuationAsOf, preview.ValuationAsOf)
	}
	if paid.MarketDataSnapshotID == "" || paid.MarketDataSnapshotID != preview.MarketDataSnapshotID {
		t.Fatalf("paid market_data_snapshot_id mismatch: got %q want %q", paid.MarketDataSnapshotID, preview.MarketDataSnapshotID)
	}
	if !floatEquals(paid.NetWorthDisplay, preview.NetWorthDisplay) {
		t.Fatalf("paid net_worth_display mismatch: got %v want %v", paid.NetWorthDisplay, preview.NetWorthDisplay)
	}
	if paid.BaseCurrency != preview.BaseCurrency {
		t.Fatalf("paid base_currency mismatch: got %q want %q", paid.BaseCurrency, preview.BaseCurrency)
	}
	if !floatEquals(paid.BaseFXRateToUSD, preview.BaseFXRateToUSD) {
		t.Fatalf("paid base_fx_rate_to_usd mismatch: got %v want %v", paid.BaseFXRateToUSD, preview.BaseFXRateToUSD)
	}

	if paid.ReportHeader.HealthScore.Value != preview.FixedMetrics.HealthScore {
		t.Fatalf("paid health_score mismatch: got %d want %d", paid.ReportHeader.HealthScore.Value, preview.FixedMetrics.HealthScore)
	}
	if paid.ReportHeader.VolatilityDashboard.Value != preview.FixedMetrics.VolatilityScore {
		t.Fatalf("paid volatility_score mismatch: got %d want %d", paid.ReportHeader.VolatilityDashboard.Value, preview.FixedMetrics.VolatilityScore)
	}
	expectedHealthStatus := scoreStatusFromScore(paid.ReportHeader.HealthScore.Value)
	if paid.ReportHeader.HealthScore.Status != expectedHealthStatus {
		t.Fatalf("paid health_score status mismatch: got %q want %q", paid.ReportHeader.HealthScore.Status, expectedHealthStatus)
	}
	expectedVolStatus := volatilityStatusFromScore(paid.ReportHeader.VolatilityDashboard.Value)
	if paid.ReportHeader.VolatilityDashboard.Status != expectedVolStatus {
		t.Fatalf("paid volatility_dashboard status mismatch: got %q want %q", paid.ReportHeader.VolatilityDashboard.Status, expectedVolStatus)
	}
	if len(paid.RiskInsights) != len(preview.IdentifiedRisks) {
		t.Fatalf("risk_insights count mismatch: got %d want %d", len(paid.RiskInsights), len(preview.IdentifiedRisks))
	}

	previewRisks := make(map[string]previewRisk, len(preview.IdentifiedRisks))
	for _, risk := range preview.IdentifiedRisks {
		previewRisks[risk.RiskID] = risk
	}
	for _, insight := range paid.RiskInsights {
		previewRisk, ok := previewRisks[insight.RiskID]
		if !ok {
			t.Fatalf("unexpected risk_id in paid report: %q", insight.RiskID)
		}
		if insight.Type != previewRisk.Type || insight.Severity != previewRisk.Severity {
			t.Fatalf("risk mismatch for %s: got %s/%s want %s/%s", insight.RiskID, insight.Type, insight.Severity, previewRisk.Type, previewRisk.Severity)
		}
		if insight.Message == "" {
			t.Fatalf("risk_insight message empty for %s", insight.RiskID)
		}
	}

	if len(paid.OptimizationPlan) != len(plans) {
		t.Fatalf("optimization_plan count mismatch: got %d want %d", len(paid.OptimizationPlan), len(plans))
	}

	expectedPlans := make(map[string]lockedPlan, len(plans))
	for _, plan := range plans {
		expectedPlans[plan.PlanID] = plan
	}
	for _, plan := range paid.OptimizationPlan {
		expected, ok := expectedPlans[plan.PlanID]
		if !ok {
			t.Fatalf("unexpected plan_id %q", plan.PlanID)
		}
		if len(expected.Parameters) > 0 && len(plan.Parameters) == 0 {
			t.Fatalf("missing parameters for plan %s", plan.PlanID)
		}
		if len(expected.Parameters) > 0 {
			if !jsonDeepEqual(plan.Parameters, expected.Parameters) {
				t.Fatalf("parameters mismatch for plan %s", plan.PlanID)
			}
		}
		if plan.LinkedRiskID != expected.LinkedRiskID {
			t.Fatalf("linked_risk_id mismatch for plan %s: got %s want %s", plan.PlanID, plan.LinkedRiskID, expected.LinkedRiskID)
		}
		if plan.ExecutionSummary == "" || plan.Rationale == "" || plan.ExpectedOutcome == "" {
			t.Fatalf("missing execution_summary, rationale, or expected_outcome for plan %s", plan.PlanID)
		}
		delete(expectedPlans, plan.PlanID)
	}
	if len(expectedPlans) != 0 {
		t.Fatalf("missing optimization plans: %v", mapKeys(expectedPlans))
	}

	if paid.TheVerdict.ConstructiveComment == "" {
		t.Fatalf("missing constructive_comment: %#v", paid.TheVerdict)
	}
}

func buildInsights(t *testing.T, client *http.Client, holdings []portfolioHolding, plans []lockedPlan, seriesByAssetKey map[string][]ohlcPoint) ([]insightItem, []indicatorSnapshot) {
	t.Helper()

	now := time.Now().UTC()
	priceByAssetKey := make(map[string]float64, len(holdings))
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" || holding.CurrentPrice <= 0 {
			continue
		}
		priceByAssetKey[holding.AssetKey] = holding.CurrentPrice
	}

	portfolioWatch := buildPortfolioWatchSignals(plans, seriesByAssetKey, priceByAssetKey, now)
	actionAlert := buildActionAlertSignals(plans, now)
	marketAlpha, indicators := buildMarketAlphaSignals(t, client, holdings, seriesByAssetKey, now)

	insights := append(portfolioWatch, actionAlert...)
	insights = append(insights, marketAlpha...)
	sortInsights(insights)

	return insights, indicators
}

func assertInsights(t *testing.T, holdings []portfolioHolding, plans []lockedPlan, indicators []indicatorSnapshot, insights []insightItem) {
	t.Helper()

	if hasIndicatorEligibleHoldings(holdings) && len(indicators) == 0 {
		t.Fatal("expected market alpha indicators for eligible holdings, got none")
	}
	if hasHoldingSymbol(holdings, "BTC") && !hasIndicatorSourceForAsset(indicators, "BTC", "binance") {
		t.Fatal("expected Binance 4h indicator series for BTC")
	}
	if hasStockHoldings(holdings) && !hasIndicatorSource(indicators, "marketstack_eod") {
		t.Fatal("expected Marketstack indicator series for stock holdings")
	}

	planIDs := make(map[string]struct{}, len(plans))
	planAssetKeys := make(map[string]struct{}, len(plans))
	for _, plan := range plans {
		planIDs[plan.PlanID] = struct{}{}
		planAssetKeys[plan.AssetKey] = struct{}{}
	}

	holdingAssetKeys := make(map[string]struct{}, len(holdings))
	for _, holding := range holdings {
		holdingAssetKeys[holding.AssetKey] = struct{}{}
	}

	for _, item := range insights {
		if item.Type != "portfolio_watch" && item.Type != "market_alpha" && item.Type != "action_alert" {
			t.Fatalf("unexpected insight type: %s", item.Type)
		}
		if item.Type == "market_alpha" {
			if item.Timeframe != "4h" && item.Timeframe != "1d" {
				t.Fatalf("market_alpha missing or invalid timeframe: %#v", item)
			}
		} else if item.Timeframe != "" {
			t.Fatalf("non-market_alpha should not include timeframe: %#v", item)
		}
		if item.TriggerReason == "" || item.TriggerKey == "" {
			t.Fatalf("insight missing trigger fields: %#v", item)
		}
		if item.Severity == "" {
			t.Fatalf("insight missing severity: %#v", item)
		}
		if item.Type == "market_alpha" && item.StrategyID != "" {
			t.Fatalf("market_alpha should not include strategy_id: %#v", item)
		}
		if item.Type == "portfolio_watch" && item.StrategyID == "" {
			t.Fatalf("portfolio_watch missing strategy_id: %#v", item)
		}
		if item.Type == "action_alert" && item.PlanID == "" {
			t.Fatalf("action_alert missing plan_id: %#v", item)
		}
		if item.PlanID != "" {
			if _, ok := planIDs[item.PlanID]; !ok {
				t.Fatalf("insight plan_id not in locked plans: %#v", item)
			}
		}
		if item.AssetKey != "" {
			if _, ok := holdingAssetKeys[item.AssetKey]; !ok {
				if _, ok := planAssetKeys[item.AssetKey]; !ok {
					t.Fatalf("insight asset_key not in holdings/plans: %#v", item)
				}
			}
		}
	}
}

func buildPortfolioWatchSignals(plans []lockedPlan, seriesByAssetKey map[string][]ohlcPoint, priceByAssetKey map[string]float64, now time.Time) []insightItem {
	items := make([]insightItem, 0)

	for _, plan := range plans {
		currentPrice, ok := priceByAssetKey[plan.AssetKey]
		if !ok || currentPrice <= 0 {
			continue
		}
		switch plan.StrategyID {
		case "S01":
			stopLossPrice, ok := getFloatParam(plan.Parameters, "stop_loss_price")
			if !ok {
				continue
			}
			if currentPrice <= stopLossPrice*1.005 {
				items = append(items, insightItem{
					ID:              insightID(plan, "stop_loss"),
					Type:            "portfolio_watch",
					Asset:           plan.Symbol,
					AssetKey:        plan.AssetKey,
					Severity:        "High",
					TriggerReason:   fmt.Sprintf("Price touched stop-loss %s", formatFloat(stopLossPrice, priceDecimals(plan.AssetType))),
					TriggerKey:      fmt.Sprintf("S01:%s:stop_loss", strings.ToUpper(plan.Symbol)),
					StrategyID:      plan.StrategyID,
					PlanID:          plan.PlanID,
					SuggestedAction: "Consider executing stop-loss to limit further drawdown.",
					CreatedAt:       now,
					ExpiresAt:       now.Add(7 * 24 * time.Hour),
				})
			}
		case "S02":
			orderPrice, amountUSD, ok := firstSafetyOrder(plan.Parameters)
			if !ok || orderPrice <= 0 {
				continue
			}
			if currentPrice <= orderPrice {
				items = append(items, insightItem{
					ID:              insightID(plan, "safety_order"),
					Type:            "portfolio_watch",
					Asset:           plan.Symbol,
					AssetKey:        plan.AssetKey,
					Severity:        "Medium",
					TriggerReason:   fmt.Sprintf("Price reached safety order %s", formatFloat(orderPrice, priceDecimals(plan.AssetType))),
					TriggerKey:      fmt.Sprintf("S02:%s:safety_order", strings.ToUpper(plan.Symbol)),
					StrategyID:      plan.StrategyID,
					PlanID:          plan.PlanID,
					SuggestedAction: fmt.Sprintf("Suggested buy: $%s", formatFloat(amountUSD, 2)),
					CreatedAt:       now,
					ExpiresAt:       now.Add(7 * 24 * time.Hour),
				})
			}
		case "S03":
			activationPrice, ok := getFloatParam(plan.Parameters, "activation_price")
			if !ok {
				continue
			}
			callbackRate, ok := getFloatParam(plan.Parameters, "callback_rate")
			if !ok || callbackRate <= 0 {
				continue
			}
			series, ok := seriesByAssetKey[plan.AssetKey]
			if !ok || len(series) == 0 {
				continue
			}
			peak := maxClose(series)
			if peak < activationPrice || peak <= 0 {
				continue
			}
			drawdown := (peak - currentPrice) / peak
			if drawdown >= callbackRate {
				items = append(items, insightItem{
					ID:              insightID(plan, "trailing_stop"),
					Type:            "portfolio_watch",
					Asset:           plan.Symbol,
					AssetKey:        plan.AssetKey,
					Severity:        "Medium",
					TriggerReason:   fmt.Sprintf("Trailing stop drawdown %.1f%% from peak %s", drawdown*100, formatFloat(peak, priceDecimals(plan.AssetType))),
					TriggerKey:      fmt.Sprintf("S03:%s:trailing_stop", strings.ToUpper(plan.Symbol)),
					StrategyID:      plan.StrategyID,
					PlanID:          plan.PlanID,
					SuggestedAction: "Consider taking profit per trailing stop rules.",
					CreatedAt:       now,
					ExpiresAt:       now.Add(7 * 24 * time.Hour),
				})
			}
		case "S04":
			layers := planLayers(plan.Parameters)
			for _, layer := range layers {
				targetPrice, ok := getFloatParam(layer, "target_price")
				if !ok || targetPrice <= 0 {
					continue
				}
				if currentPrice < targetPrice {
					continue
				}
				layerName := getStringParam(layer, "layer_name")
				sellPct, _ := getFloatParam(layer, "sell_percentage")
				items = append(items, insightItem{
					ID:              insightID(plan, layerName),
					Type:            "portfolio_watch",
					Asset:           plan.Symbol,
					AssetKey:        plan.AssetKey,
					Severity:        "High",
					TriggerReason:   fmt.Sprintf("Price reached take-profit %s", formatFloat(targetPrice, priceDecimals(plan.AssetType))),
					TriggerKey:      fmt.Sprintf("S04:%s:%s", strings.ToUpper(plan.Symbol), slugify(layerName)),
					StrategyID:      plan.StrategyID,
					PlanID:          plan.PlanID,
					SuggestedAction: fmt.Sprintf("Sell %.0f%% as %s take-profit.", sellPct*100, layerName),
					CreatedAt:       now,
					ExpiresAt:       now.Add(7 * 24 * time.Hour),
				})
			}
		}
	}

	return items
}

func buildActionAlertSignals(plans []lockedPlan, now time.Time) []insightItem {
	items := make([]insightItem, 0)

	const executionWindow = 24 * time.Hour
	for _, plan := range plans {
		if plan.StrategyID != "S05" {
			continue
		}
		nextExecution := getStringParam(plan.Parameters, "next_execution_at")
		if nextExecution == "" {
			continue
		}
		scheduled, err := time.Parse(time.RFC3339, nextExecution)
		if err != nil {
			continue
		}
		if now.Before(scheduled) || now.After(scheduled.Add(executionWindow)) {
			continue
		}
		amount, ok := getFloatParam(plan.Parameters, "amount")
		if !ok {
			continue
		}
		items = append(items, insightItem{
			ID:              insightID(plan, scheduled.Format("20060102")),
			Type:            "action_alert",
			Asset:           plan.Symbol,
			AssetKey:        plan.AssetKey,
			Severity:        "Medium",
			TriggerReason:   fmt.Sprintf("DCA scheduled at %s", scheduled.Format(time.RFC3339)),
			TriggerKey:      fmt.Sprintf("S05:%s:%s", strings.ToUpper(plan.Symbol), scheduled.Format("20060102")),
			StrategyID:      plan.StrategyID,
			PlanID:          plan.PlanID,
			SuggestedAction: fmt.Sprintf("Suggested buy: $%s", formatFloat(amount, 2)),
			CreatedAt:       now,
			ExpiresAt:       scheduled.Add(executionWindow),
		})
	}

	return items
}

func buildMarketAlphaSignals(t *testing.T, client *http.Client, holdings []portfolioHolding, seriesByAssetKey map[string][]ohlcPoint, now time.Time) ([]insightItem, []indicatorSnapshot) {
	t.Helper()

	items := make([]insightItem, 0)
	indicators := make([]indicatorSnapshot, 0)
	minCloses := bollingerPeriod
	if rsiPeriod+1 > minCloses {
		minCloses = rsiPeriod + 1
	}

	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		if holding.BalanceType == "stablecoin" || holding.BalanceType == "fiat_cash" {
			continue
		}
		series, ok := indicatorSeriesForHolding(t, client, holding, seriesByAssetKey)
		if !ok || len(series.Points) < minCloses {
			continue
		}
		closes := extractCloses(series.Points)
		rsi, ok := computeRSI(closes, rsiPeriod)
		if !ok {
			continue
		}
		upper, lower, ok := computeBollinger(closes, bollingerPeriod, bollingerStdDev)
		if !ok {
			continue
		}
		lastClose := closes[len(closes)-1]
		indicators = append(indicators, indicatorSnapshot{
			Asset:     holding.Symbol,
			AssetKey:  holding.AssetKey,
			AssetType: holding.AssetType,
			Interval:  series.Interval,
			Source:    series.Source,
			RSI:       roundTo(rsi, 2),
			UpperBand: roundTo(upper, priceDecimals(holding.AssetType)),
			LowerBand: roundTo(lower, priceDecimals(holding.AssetType)),
			LastClose: roundTo(lastClose, priceDecimals(holding.AssetType)),
		})
		if rsi < 30 && lastClose <= lower {
			reason := fmt.Sprintf("RSI(14) < 30 on %s chart, touching lower Bollinger band", series.Interval)
			items = append(items, insightItem{
				ID:              insightIDFromHolding(holding, "rsi_bollinger"),
				Type:            "market_alpha",
				Asset:           holding.Symbol,
				AssetKey:        holding.AssetKey,
				Timeframe:       series.Interval,
				Severity:        "Medium",
				TriggerReason:   reason,
				TriggerKey:      fmt.Sprintf("market_alpha:%s:rsi_bollinger_%s", strings.ToLower(holding.Symbol), series.Interval),
				SuggestedAction: "Watch for rebound confirmation.",
				CreatedAt:       now,
				ExpiresAt:       now.Add(24 * time.Hour),
			})
		}
	}

	return items, indicators
}

func indicatorSeriesForHolding(t *testing.T, client *http.Client, holding portfolioHolding, seriesByAssetKey map[string][]ohlcPoint) (indicatorSeries, bool) {
	t.Helper()

	switch holding.AssetType {
	case "crypto":
		symbol := strings.ToUpper(strings.TrimSpace(holding.Symbol))
		if symbol == "" {
			return indicatorSeries{}, false
		}
		series, ok := fetchBinanceKlines(t, client, symbol+"USDT", binanceKlinesInterval, binanceKlinesLimit)
		if ok {
			return indicatorSeries{Points: series, Interval: binanceKlinesInterval, Source: "binance"}, true
		}
		if fallback, ok := seriesByAssetKey[holding.AssetKey]; ok && len(fallback) > 0 {
			return indicatorSeries{Points: fallback, Interval: "1d", Source: "coingecko_daily"}, true
		}
	case "stock":
		if fallback, ok := seriesByAssetKey[holding.AssetKey]; ok && len(fallback) > 0 {
			return indicatorSeries{Points: fallback, Interval: "1d", Source: "marketstack_eod"}, true
		}
	}
	return indicatorSeries{}, false
}

func fetchBinanceKlines(t *testing.T, client *http.Client, symbol, interval string, limit int) ([]ohlcPoint, bool) {
	t.Helper()

	query := url.Values{}
	query.Set("symbol", symbol)
	query.Set("interval", interval)
	query.Set("limit", strconv.Itoa(limit))

	endpoint := fmt.Sprintf("https://api.binance.com/api/v3/klines?%s", query.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("build binance klines request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("binance klines request failed: %v", err)
	}
	defer resp.Body.Close()
	debugLogf(t, "HTTP %s %s -> %d (%s)", req.Method, sanitizeURL(req.URL), resp.StatusCode, time.Since(start))

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		debugLogf(t, "Binance klines unavailable for %s: %s", symbol, strings.TrimSpace(string(body)))
		return nil, false
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("binance klines %s status %d: %s", symbol, resp.StatusCode, string(body))
	}

	var raw [][]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode binance klines %s: %v", symbol, err)
	}

	points := make([]ohlcPoint, 0, len(raw))
	for _, row := range raw {
		if len(row) < 5 {
			continue
		}
		ts, ok := parseBinanceInt(row[0])
		if !ok {
			continue
		}
		open, ok := parseBinanceFloat(row[1])
		if !ok {
			continue
		}
		high, ok := parseBinanceFloat(row[2])
		if !ok {
			continue
		}
		low, ok := parseBinanceFloat(row[3])
		if !ok {
			continue
		}
		closeValue, ok := parseBinanceFloat(row[4])
		if !ok {
			continue
		}
		points = append(points, ohlcPoint{
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closeValue,
		})
	}

	if len(points) == 0 {
		return nil, false
	}
	return points, true
}

func parseBinanceFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case json.Number:
		parsed, err := v.Float64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func parseBinanceInt(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case json.Number:
		parsed, err := v.Int64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func computeRSI(closes []float64, period int) (float64, bool) {
	if len(closes) < period+1 {
		return 0, false
	}
	start := len(closes) - period
	if start <= 0 {
		start = 1
	}
	gain := 0.0
	loss := 0.0
	for i := start; i < len(closes); i++ {
		delta := closes[i] - closes[i-1]
		if delta > 0 {
			gain += delta
		} else {
			loss -= delta
		}
	}
	avgGain := gain / float64(period)
	avgLoss := loss / float64(period)
	if avgLoss == 0 {
		return 100, true
	}
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi, true
}

func computeBollinger(closes []float64, period int, stdDev float64) (float64, float64, bool) {
	if len(closes) < period {
		return 0, 0, false
	}
	start := len(closes) - period
	window := closes[start:]
	sum := 0.0
	for _, value := range window {
		sum += value
	}
	mean := sum / float64(period)
	variance := 0.0
	for _, value := range window {
		diff := value - mean
		variance += diff * diff
	}
	std := math.Sqrt(variance / float64(period))
	upper := mean + stdDev*std
	lower := mean - stdDev*std
	return upper, lower, true
}

func extractCloses(points []ohlcPoint) []float64 {
	closes := make([]float64, 0, len(points))
	for _, point := range points {
		closes = append(closes, point.Close)
	}
	return closes
}

func maxClose(points []ohlcPoint) float64 {
	max := 0.0
	for _, point := range points {
		if point.Close > max {
			max = point.Close
		}
	}
	return max
}

func firstSafetyOrder(params map[string]any) (float64, float64, bool) {
	raw, ok := params["safety_orders"]
	if !ok {
		return 0, 0, false
	}
	switch orders := raw.(type) {
	case []map[string]any:
		if len(orders) == 0 {
			return 0, 0, false
		}
		price, okPrice := getFloatParam(orders[0], "price")
		amount, okAmount := getFloatParam(orders[0], "amount_usd")
		return price, amount, okPrice && okAmount
	case []any:
		if len(orders) == 0 {
			return 0, 0, false
		}
		if first, ok := orders[0].(map[string]any); ok {
			price, okPrice := getFloatParam(first, "price")
			amount, okAmount := getFloatParam(first, "amount_usd")
			return price, amount, okPrice && okAmount
		}
	}
	return 0, 0, false
}

func planLayers(params map[string]any) []map[string]any {
	raw, ok := params["layers"]
	if !ok {
		return nil
	}
	switch layers := raw.(type) {
	case []map[string]any:
		return layers
	case []any:
		parsed := make([]map[string]any, 0, len(layers))
		for _, item := range layers {
			if layer, ok := item.(map[string]any); ok {
				parsed = append(parsed, layer)
			}
		}
		return parsed
	default:
		return nil
	}
}

func getFloatParam(params map[string]any, key string) (float64, bool) {
	value, ok := params[key]
	if !ok {
		return 0, false
	}
	return parseBinanceFloat(value)
}

func getStringParam(params map[string]any, key string) string {
	value, ok := params[key]
	if !ok {
		return ""
	}
	if parsed, ok := value.(string); ok {
		return parsed
	}
	return ""
}

func insightID(plan lockedPlan, suffix string) string {
	return fmt.Sprintf("ins_%s_%s_%s", strings.ToLower(plan.StrategyID), strings.ToLower(plan.Symbol), slugify(suffix))
}

func insightIDFromHolding(holding portfolioHolding, suffix string) string {
	return fmt.Sprintf("ins_%s_%s", strings.ToLower(holding.Symbol), slugify(suffix))
}

func hasIndicatorEligibleHoldings(holdings []portfolioHolding) bool {
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		if holding.BalanceType == "stablecoin" || holding.BalanceType == "fiat_cash" {
			continue
		}
		return true
	}
	return false
}

func hasStockHoldings(holdings []portfolioHolding) bool {
	for _, holding := range holdings {
		if holding.ValuationStatus == "priced" && holding.AssetType == "stock" {
			return true
		}
	}
	return false
}

func hasHoldingSymbol(holdings []portfolioHolding, symbol string) bool {
	for _, holding := range holdings {
		if strings.EqualFold(holding.Symbol, symbol) {
			return true
		}
	}
	return false
}

func hasIndicatorSource(indicators []indicatorSnapshot, source string) bool {
	for _, indicator := range indicators {
		if indicator.Source == source {
			return true
		}
	}
	return false
}

func hasIndicatorSourceForAsset(indicators []indicatorSnapshot, symbol, source string) bool {
	for _, indicator := range indicators {
		if strings.EqualFold(indicator.Asset, symbol) && indicator.Source == source {
			return true
		}
	}
	return false
}

func sortInsights(items []insightItem) {
	sort.SliceStable(items, func(i, j int) bool {
		typeRank := insightTypeRank(items[i].Type)
		typeRankOther := insightTypeRank(items[j].Type)
		if typeRank != typeRankOther {
			return typeRank < typeRankOther
		}
		severityRank := insightSeverityRank(items[i].Severity)
		severityRankOther := insightSeverityRank(items[j].Severity)
		if severityRank != severityRankOther {
			return severityRank < severityRankOther
		}
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].TriggerKey < items[j].TriggerKey
	})
}

func insightTypeRank(value string) int {
	switch value {
	case "portfolio_watch":
		return 0
	case "action_alert":
		return 1
	case "market_alpha":
		return 2
	default:
		return 3
	}
}

func insightSeverityRank(value string) int {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func mergePaidPlanParameters(paid *paidPromptOutput, plans []lockedPlan) {
	if paid == nil || len(plans) == 0 {
		return
	}

	byID := make(map[string]lockedPlan, len(plans))
	for _, plan := range plans {
		byID[plan.PlanID] = plan
	}
	for i := range paid.OptimizationPlan {
		if locked, ok := byID[paid.OptimizationPlan[i].PlanID]; ok {
			paid.OptimizationPlan[i].StrategyID = locked.StrategyID
			paid.OptimizationPlan[i].AssetType = locked.AssetType
			paid.OptimizationPlan[i].Symbol = locked.Symbol
			paid.OptimizationPlan[i].AssetKey = locked.AssetKey
			paid.OptimizationPlan[i].LinkedRiskID = locked.LinkedRiskID
			paid.OptimizationPlan[i].Parameters = locked.Parameters
		}
	}
}

func jsonDeepEqual(a, b any) bool {
	encodedA, err := json.Marshal(a)
	if err != nil {
		return false
	}
	encodedB, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(encodedA) == string(encodedB)
}

func validateRisks(t *testing.T, risks []previewRisk, metricsIncomplete bool) {
	t.Helper()
	violations := collectRiskViolations(risks, metricsIncomplete)
	if len(violations) == 0 {
		return
	}
	t.Fatalf("risk validation failed: %s", strings.Join(violations, "; "))
}

func collectRiskViolations(risks []previewRisk, metricsIncomplete bool) []string {
	var violations []string
	allowedTypes := map[string]struct{}{
		"Liquidity Risk":           {},
		"Concentration Risk":       {},
		"Volatility Risk":          {},
		"Correlation Risk":         {},
		"Drawdown Risk":            {},
		"Inefficient Capital Risk": {},
	}
	allowedSeverity := map[string]struct{}{
		"Low":      {},
		"Medium":   {},
		"High":     {},
		"Critical": {},
	}

	ids := make(map[string]struct{}, len(risks))
	types := make(map[string]struct{}, len(risks))
	teasers := make(map[string]struct{}, len(risks))
	for _, risk := range risks {
		id := strings.TrimSpace(risk.RiskID)
		if id == "" {
			violations = append(violations, "missing risk_id")
		} else {
			if id != "risk_01" && id != "risk_02" && id != "risk_03" {
				violations = append(violations, fmt.Sprintf("unexpected risk_id %s", id))
			}
			if _, ok := ids[id]; ok {
				violations = append(violations, fmt.Sprintf("duplicate risk_id %s", id))
			}
			ids[id] = struct{}{}
		}

		typ := strings.TrimSpace(risk.Type)
		if _, ok := allowedTypes[typ]; !ok {
			violations = append(violations, fmt.Sprintf("unexpected risk type %s", typ))
		} else {
			if _, ok := types[typ]; ok {
				violations = append(violations, fmt.Sprintf("duplicate risk type %s", typ))
			}
			types[typ] = struct{}{}
		}

		if _, ok := allowedSeverity[risk.Severity]; !ok {
			violations = append(violations, fmt.Sprintf("unexpected risk severity %s", risk.Severity))
		}

		teaser := strings.TrimSpace(risk.TeaserText)
		if teaser == "" {
			violations = append(violations, fmt.Sprintf("empty teaser_text for %s", id))
		} else {
			key := strings.ToLower(teaser)
			if _, ok := teasers[key]; ok {
				violations = append(violations, "duplicate teaser_text")
			}
			teasers[key] = struct{}{}
		}
	}

	if len(ids) != 3 {
		violations = append(violations, fmt.Sprintf("expected 3 distinct risk_ids, got %d", len(ids)))
	}
	if _, ok := ids["risk_01"]; !ok {
		violations = append(violations, "missing risk_01 in identified_risks")
	}
	if _, ok := ids["risk_02"]; !ok {
		violations = append(violations, "missing risk_02 in identified_risks")
	}
	if _, ok := ids["risk_03"]; !ok {
		violations = append(violations, "missing risk_03 in identified_risks")
	}

	if metricsIncomplete {
		for _, risk := range risks {
			if risk.RiskID == "risk_03" && risk.Severity == "Low" {
				violations = append(violations, "risk_03 severity must be at least Medium when metrics_incomplete=true")
				break
			}
		}
	}

	return violations
}

func buildLockedPlans(t *testing.T, profile userProfile, holdings []portfolioHolding, metrics portfolioMetrics, seriesByAssetKey map[string][]ohlcPoint) []lockedPlan {
	t.Helper()

	riskLevel := strings.ToLower(profile.RiskTolerance)
	if riskLevel == "" {
		riskLevel = "moderate"
	}

	eligible := eligibleHoldingsForPlans(holdings)
	plans := make([]lockedPlan, 0, maxPlansPerPortfolio)
	selected := make(map[string]struct{})

	if metrics.IdleCashUSD >= 20 {
		if candidate := selectMostNegativePNL(eligible); candidate != nil && candidate.CostBasisStatus == "provided" && candidate.PNLPercent != nil && *candidate.PNLPercent <= -0.10 {
			if plan, ok := buildS02Plan(profile, *candidate, metrics); ok {
				plans = append(plans, plan)
				selected[candidate.AssetKey] = struct{}{}
			}
		}
	}

	if len(plans) < maxPlansPerPortfolio {
		if candidate := selectBestProfit(eligible, 0.30); candidate != nil {
			if plan, ok := buildS04Plan(profile, *candidate); ok {
				plans = append(plans, plan)
				selected[candidate.AssetKey] = struct{}{}
			}
		} else if candidate := selectBestProfit(eligible, minProfitPct(riskLevel)); candidate != nil {
			plans = append(plans, buildS03Plan(riskLevel, *candidate))
			selected[candidate.AssetKey] = struct{}{}
		}
	}

	if len(plans) < maxPlansPerPortfolio && metrics.IdleCashUSD >= 50 {
		plan := buildS05Plan(profile, holdings, metrics)
		plans = append(plans, plan)
		selected[plan.AssetKey] = struct{}{}
	}

	if len(plans) < maxPlansPerPortfolio {
		if candidate := selectHighestValue(eligible, selected); candidate != nil {
			plans = append(plans, buildS01Plan(profile, *candidate, metrics, seriesByAssetKey))
		}
	}

	for i := range plans {
		plans[i].PlanID = fmt.Sprintf("plan_%02d", i+1)
	}

	return plans
}

func eligibleHoldingsForPlans(holdings []portfolioHolding) []portfolioHolding {
	eligible := make([]portfolioHolding, 0, len(holdings))
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		eligible = append(eligible, holding)
	}
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].ValueUSD > eligible[j].ValueUSD
	})
	return eligible
}

func selectMostNegativePNL(holdings []portfolioHolding) *portfolioHolding {
	var selected *portfolioHolding
	for i := range holdings {
		if holdings[i].PNLPercent == nil {
			continue
		}
		if selected == nil || *holdings[i].PNLPercent < *selected.PNLPercent {
			selected = &holdings[i]
		}
	}
	return selected
}

func selectBestProfit(holdings []portfolioHolding, threshold float64) *portfolioHolding {
	var selected *portfolioHolding
	for i := range holdings {
		if holdings[i].PNLPercent == nil {
			continue
		}
		if *holdings[i].PNLPercent < threshold {
			continue
		}
		if selected == nil || *holdings[i].PNLPercent > *selected.PNLPercent {
			selected = &holdings[i]
		}
	}
	return selected
}

func selectHighestValue(holdings []portfolioHolding, excluded map[string]struct{}) *portfolioHolding {
	for i := range holdings {
		if _, ok := excluded[holdings[i].AssetKey]; ok {
			continue
		}
		return &holdings[i]
	}
	return nil
}

func buildS01Plan(profile userProfile, holding portfolioHolding, metrics portfolioMetrics, seriesByAssetKey map[string][]ohlcPoint) lockedPlan {
	riskLevel := strings.ToLower(profile.RiskTolerance)
	if riskLevel == "" {
		riskLevel = "moderate"
	}

	baseSL := baseStopLossPct(riskLevel)
	volAdj := 1.0
	if metrics.Volatility30dDaily > 0.06 {
		volAdj = 1.5
	} else if metrics.Volatility30dDaily > 0.04 {
		volAdj = 1.2
	}

	experienceAdj := 1.0
	switch strings.ToLower(profile.Experience) {
	case "beginner":
		experienceAdj = 0.8
	case "expert":
		experienceAdj = 1.2
	}

	lossAdj := 1.0
	if holding.PNLPercent != nil && *holding.PNLPercent < -0.10 {
		lossAdj = 1.2
	}

	stopLossPct := clamp(baseSL*volAdj*experienceAdj*lossAdj, 0.03, 0.15)
	stopLossPrice := holding.CurrentPrice * (1 - stopLossPct)

	if holding.AvgPrice != nil {
		if holding.PNLPercent != nil && *holding.PNLPercent < 0 {
			stopFromCost := *holding.AvgPrice * (1 - stopLossPct)
			if stopFromCost < stopLossPrice {
				stopLossPrice = stopFromCost
			}
		} else {
			stopLossPrice = *holding.AvgPrice * (1 - stopLossPct)
		}
	}

	adjusted := false
	params := map[string]any{
		"stop_loss_pct":   roundTo(stopLossPct, 4),
		"stop_loss_price": roundTo(stopLossPrice, priceDecimals(holding.AssetType)),
	}

	if series, ok := seriesByAssetKey[holding.AssetKey]; ok {
		if support, ok := closestSupportBelow(series, stopLossPrice); ok {
			if stopLossPrice > support*0.95 {
				stopLossPrice = support * 0.98
				stopLossPrice = roundTo(stopLossPrice, priceDecimals(holding.AssetType))
				params["stop_loss_price"] = stopLossPrice
				params["adjustment_reason"] = "Adjusted to support level"
				params["support_level"] = roundTo(support, priceDecimals(holding.AssetType))
				adjusted = true
			}
		}
	}

	if !adjusted {
		params["adjustment_reason"] = "Risk/volatility-based"
	}

	return lockedPlan{
		StrategyID: "S01",
		AssetType:  holding.AssetType,
		Symbol:     holding.Symbol,
		AssetKey:   holding.AssetKey,
		Parameters: params,
	}
}

func buildS02Plan(profile userProfile, holding portfolioHolding, metrics portfolioMetrics) (lockedPlan, bool) {
	riskLevel := strings.ToLower(profile.RiskTolerance)
	if riskLevel == "" {
		riskLevel = "moderate"
	}

	stepPct, numOrders, scale := s02Config(riskLevel)
	seriesSum := 0.0
	for i := 0; i < numOrders; i++ {
		seriesSum += math.Pow(scale, float64(i))
	}

	maxAffordable := metrics.IdleCashUSD / seriesSum
	if maxAffordable < 20 {
		return lockedPlan{}, false
	}
	baseOrderUSD := clamp(metrics.IdleCashUSD*0.05, 20, math.Min(2000, maxAffordable))

	safetyOrders := make([]map[string]any, 0, numOrders)
	totalUSD := 0.0
	totalAsset := 0.0
	for i := 1; i <= numOrders; i++ {
		price := holding.CurrentPrice * (1 - stepPct*float64(i))
		amountUSD := baseOrderUSD * math.Pow(scale, float64(i-1))
		amountAsset := amountUSD / price
		safetyOrders = append(safetyOrders, map[string]any{
			"price":        roundTo(price, priceDecimals(holding.AssetType)),
			"amount_usd":   roundTo(amountUSD, 2),
			"amount_asset": roundTo(amountAsset, amountDecimals(holding.AssetType)),
		})
		totalUSD += amountUSD
		totalAsset += amountAsset
	}

	targetAvg := 0.0
	if holding.AvgPrice != nil && holding.Amount > 0 {
		targetAvg = (*holding.AvgPrice*holding.Amount + totalUSD) / (holding.Amount + totalAsset)
	}

	params := map[string]any{
		"safety_orders":    safetyOrders,
		"scale":            roundTo(scale, 2),
		"target_avg_price": roundTo(targetAvg, priceDecimals(holding.AssetType)),
	}

	return lockedPlan{
		StrategyID: "S02",
		AssetType:  holding.AssetType,
		Symbol:     holding.Symbol,
		AssetKey:   holding.AssetKey,
		Parameters: params,
	}, true
}

func buildS03Plan(riskLevel string, holding portfolioHolding) lockedPlan {
	minProfit := minProfitPct(riskLevel)
	callback := callbackRate(riskLevel)

	activation := holding.CurrentPrice
	if holding.AvgPrice != nil {
		activation = math.Max(holding.CurrentPrice, *holding.AvgPrice*(1+minProfit))
		if holding.PNLPercent != nil && *holding.PNLPercent >= minProfit {
			activation = holding.CurrentPrice
		}
	}

	params := map[string]any{
		"activation_price": roundTo(activation, priceDecimals(holding.AssetType)),
		"callback_rate":    roundTo(callback, 3),
	}

	return lockedPlan{
		StrategyID: "S03",
		AssetType:  holding.AssetType,
		Symbol:     holding.Symbol,
		AssetKey:   holding.AssetKey,
		Parameters: params,
	}
}

func buildS04Plan(profile userProfile, holding portfolioHolding) (lockedPlan, bool) {
	if holding.AvgPrice == nil {
		return lockedPlan{}, false
	}

	riskLevel := strings.ToLower(profile.RiskTolerance)
	if riskLevel == "" {
		riskLevel = "moderate"
	}

	var layerConfigs []struct {
		Name       string
		SellPct    float64
		Multiplier float64
	}
	switch riskLevel {
	case "conservative":
		layerConfigs = []struct {
			Name       string
			SellPct    float64
			Multiplier float64
		}{
			{"Layer 1", 0.40, 1.30},
			{"Layer 2", 0.35, 1.50},
			{"Layer 3", 0.25, 1.80},
		}
	case "aggressive":
		layerConfigs = []struct {
			Name       string
			SellPct    float64
			Multiplier float64
		}{
			{"Layer 1", 0.20, 1.50},
			{"Layer 2", 0.30, 2.00},
			{"Layer 3", 0.50, 3.00},
		}
	default:
		layerConfigs = []struct {
			Name       string
			SellPct    float64
			Multiplier float64
		}{
			{"Layer 1", 0.30, 1.40},
			{"Layer 2", 0.40, 1.70},
			{"Layer 3", 0.30, 2.00},
		}
	}

	layers := make([]map[string]any, 0, len(layerConfigs))
	for _, cfg := range layerConfigs {
		targetPrice := *holding.AvgPrice * cfg.Multiplier
		if holding.CurrentPrice >= targetPrice {
			targetPrice = holding.CurrentPrice * 1.05
		}
		sellAmount := holding.Amount * cfg.SellPct
		expectedProfit := (targetPrice - *holding.AvgPrice) * sellAmount
		layers = append(layers, map[string]any{
			"layer_name":          cfg.Name,
			"sell_percentage":     roundTo(cfg.SellPct, 2),
			"sell_amount":         roundTo(sellAmount, amountDecimals(holding.AssetType)),
			"target_price":        roundTo(targetPrice, priceDecimals(holding.AssetType)),
			"expected_profit_usd": roundTo(expectedProfit, 2),
		})
	}

	return lockedPlan{
		StrategyID: "S04",
		AssetType:  holding.AssetType,
		Symbol:     holding.Symbol,
		AssetKey:   holding.AssetKey,
		Parameters: map[string]any{"layers": layers},
	}, true
}

func buildS05Plan(profile userProfile, holdings []portfolioHolding, metrics portfolioMetrics) lockedPlan {
	baseAmount := math.Min(metrics.IdleCashUSD*0.10, metrics.NetWorthUSD*0.02)
	riskAdjustment := 1.0
	switch strings.ToLower(profile.RiskPreference) {
	case "yield seeker":
		riskAdjustment = 0.75
	case "speculator":
		riskAdjustment = 1.25
	}

	amount := clamp(baseAmount*riskAdjustment, 20, 2000)
	frequency := "weekly"
	if strings.EqualFold(profile.Style, "Day Trading") || strings.EqualFold(profile.Style, "Scalping") {
		frequency = "daily"
		amount = amount * 0.5
	}

	nextExecution := nextExecutionAt(profile.Timezone, frequency)

	target := pickS05Target(profile, holdings)
	return lockedPlan{
		StrategyID: "S05",
		AssetType:  target.AssetType,
		Symbol:     target.Symbol,
		AssetKey:   target.AssetKey,
		Parameters: map[string]any{
			"amount":            roundTo(amount, 2),
			"frequency":         frequency,
			"next_execution_at": nextExecution,
		},
	}
}

func pickS05Target(profile userProfile, holdings []portfolioHolding) portfolioHolding {
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.AssetType != "crypto" && holding.AssetType != "stock" {
			continue
		}
		if holding.BalanceType == "fiat_cash" || holding.BalanceType == "stablecoin" {
			continue
		}
		return holding
	}

	if marketsHas(profile.Markets, "Crypto") {
		return portfolioHolding{
			AssetType: "crypto",
			Symbol:    "BTC",
			AssetKey:  "crypto:cg:bitcoin",
		}
	}
	return portfolioHolding{
		AssetType: "stock",
		Symbol:    "SPY",
		AssetKey:  "stock:mic:XNYS:SPY",
	}
}

func assignLinkedRiskIDs(plans []lockedPlan, risks []previewRisk) []lockedPlan {
	if len(plans) == 0 || len(risks) == 0 {
		return plans
	}

	riskByType := make(map[string]string)
	for _, risk := range risks {
		riskByType[risk.Type] = risk.RiskID
	}

	assigned := make(map[string]struct{})
	for i := range plans {
		switch plans[i].StrategyID {
		case "S05":
			if id := riskByType["Liquidity Risk"]; id != "" {
				plans[i].LinkedRiskID = id
				assigned[id] = struct{}{}
				continue
			}
		case "S01", "S02":
			if id := riskByType["Drawdown Risk"]; id != "" {
				plans[i].LinkedRiskID = id
				assigned[id] = struct{}{}
				continue
			}
		case "S03", "S04":
			if id := riskByType["Volatility Risk"]; id != "" {
				plans[i].LinkedRiskID = id
				assigned[id] = struct{}{}
				continue
			}
		}
	}

	unassigned := make([]string, 0, len(risks))
	for _, risk := range risks {
		if _, ok := assigned[risk.RiskID]; !ok {
			unassigned = append(unassigned, risk.RiskID)
		}
	}

	index := 0
	for i := range plans {
		if plans[i].LinkedRiskID != "" {
			continue
		}
		if len(unassigned) == 0 {
			plans[i].LinkedRiskID = risks[0].RiskID
			continue
		}
		plans[i].LinkedRiskID = unassigned[index%len(unassigned)]
		index++
	}

	return plans
}

func closestSupportBelow(points []ohlcPoint, target float64) (float64, bool) {
	if len(points) < 7 {
		return 0, false
	}

	lows := make([]float64, len(points))
	for i, point := range points {
		lows[i] = point.Low
	}

	supports := make([]float64, 0)
	for i := 3; i < len(points)-3; i++ {
		low := lows[i]
		if low <= minSlice(lows[i-3:i]) && low <= minSlice(lows[i+1:i+4]) {
			supports = append(supports, low)
		}
	}

	closest := 0.0
	for _, support := range supports {
		if support >= target {
			continue
		}
		if closest == 0 || target-support < target-closest {
			closest = support
		}
	}

	if closest == 0 {
		return 0, false
	}
	return closest, true
}

func minSlice(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}

func nextExecutionAt(timezone, frequency string) string {
	location := time.UTC
	if timezone != "" {
		if loc, err := time.LoadLocation(timezone); err == nil {
			location = loc
		}
	}

	now := time.Now().In(location)
	scheduled := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, location)
	if !now.Before(scheduled) {
		if frequency == "weekly" {
			scheduled = scheduled.AddDate(0, 0, 7)
		} else {
			scheduled = scheduled.AddDate(0, 0, 1)
		}
	}
	return scheduled.UTC().Format(time.RFC3339)
}

func minProfitPct(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.08
	case "aggressive":
		return 0.18
	default:
		return 0.12
	}
}

func callbackRate(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.03
	case "aggressive":
		return 0.08
	default:
		return 0.05
	}
}

func baseStopLossPct(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.05
	case "aggressive":
		return 0.12
	default:
		return 0.08
	}
}

func s02Config(riskLevel string) (float64, int, float64) {
	switch riskLevel {
	case "conservative":
		return 0.05, 2, 1.3
	case "aggressive":
		return 0.10, 4, 2.0
	default:
		return 0.07, 3, 1.6
	}
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

func platformGuessToCategory(guess string) string {
	switch strings.ToLower(strings.TrimSpace(guess)) {
	case "okx", "binance", "coinbase", "kraken", "bybit", "kucoin":
		return "crypto_exchange"
	case "metamask", "trust wallet", "coinbase wallet":
		return "wallet"
	default:
		return "unknown"
	}
}

func stablecoinSet() map[string]bool {
	return map[string]bool{
		"USDT":  true,
		"USDC":  true,
		"DAI":   true,
		"TUSD":  true,
		"BUSD":  true,
		"FDUSD": true,
		"USDP":  true,
		"FRAX":  true,
	}
}

func aliasSymbol(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "bitcoin":
		return "BTC", true
	case "ethereum", "ether":
		return "ETH", true
	case "tether":
		return "USDT", true
	case "usd coin":
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
	normalized := strings.ToLower(strings.TrimSpace(assetType))
	switch normalized {
	case "", "crypto":
		return "crypto"
	case "equity", "stock":
		return "stock"
	case "forex", "fx", "fiat", "fiat_cash", "cash":
		return "forex"
	default:
		return normalized
	}
}

func normalizeSymbol(symbol string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(symbol))
	var builder strings.Builder
	for _, r := range trimmed {
		if r >= 'A' && r <= 'Z' {
			builder.WriteRune(r)
		}
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func manualAssetKey(userID, symbolRaw, platformGuess string) string {
	data := fmt.Sprintf("%s|%s", symbolRaw, platformGuess)
	sum := sha256.Sum256([]byte(data))
	return fmt.Sprintf("manual:%s:%x", userID, sum[:6])
}

func stockAssetKey(exchangeMIC, symbol string) string {
	return fmt.Sprintf("stock:mic:%s:%s", exchangeMIC, symbol)
}

func resolveCoinGeckoID(t *testing.T, client *http.Client, apiKey, symbol string, symbolToIDs map[string][]string) string {
	t.Helper()

	candidates := symbolToIDs[strings.ToLower(symbol)]
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	query := url.Values{}
	query.Set("vs_currency", "usd")
	query.Set("ids", strings.Join(candidates, ","))

	var resp []coinGeckoMarketsItem
	coinGeckoGet(t, client, apiKey, "/coins/markets", query, &resp)
	if len(resp) == 0 {
		return ""
	}

	bestID := ""
	bestCap := 0.0
	tied := false
	for _, item := range resp {
		if item.MarketCap <= 0 {
			continue
		}
		if item.MarketCap > bestCap {
			bestCap = item.MarketCap
			bestID = item.ID
			tied = false
		} else if item.MarketCap == bestCap && item.MarketCap != 0 {
			tied = true
		}
	}
	if bestID == "" || tied {
		return ""
	}
	return bestID
}

func marketsHas(markets []string, value string) bool {
	for _, market := range markets {
		if strings.EqualFold(strings.TrimSpace(market), value) {
			return true
		}
	}
	return false
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

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
