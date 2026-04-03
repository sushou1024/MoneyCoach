package main

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"
)

type assetCommandResponse struct {
	Intent   string                `json:"intent"`
	Payloads []assetCommandPayload `json:"payloads"`
}

type assetCommandPayload struct {
	TargetAsset   *assetCommandTarget  `json:"target_asset"`
	FundingSource *assetCommandFunding `json:"funding_source"`
	PricePerUnit  *float64             `json:"price_per_unit"`
}

type assetCommandTarget struct {
	Ticker string   `json:"ticker"`
	Amount *float64 `json:"amount"`
	Action string   `json:"action"`
}

type assetCommandFunding struct {
	Ticker     *string  `json:"ticker"`
	Amount     *float64 `json:"amount"`
	IsExplicit bool     `json:"is_explicit"`
}

type tradeSlipOCRResponse struct {
	ImageID string           `json:"image_id"`
	Trades  []tradeSlipTrade `json:"trades"`
}

type tradeSlipTrade struct {
	Side       string         `json:"side"`
	Symbol     string         `json:"symbol"`
	Amount     *float64       `json:"amount"`
	Price      *float64       `json:"price"`
	Currency   *string        `json:"currency"`
	ExecutedAt *string        `json:"executed_at"`
	Fees       *tradeSlipFees `json:"fees"`
}

type tradeSlipFees struct {
	Amount   *float64 `json:"amount"`
	Currency *string  `json:"currency"`
}

func TestAssetsTabUpdates(t *testing.T) {
	geminiKey := requireEnv(t, "GEMINI_API_KEY")
	coingeckoKey := requireEnv(t, "COINGECKO_PRO_API_KEY")

	geminiClient := &http.Client{Timeout: 60 * time.Second}
	apiClient := &http.Client{Timeout: 30 * time.Second}

	t.Run("TradeSlipPortfolio1", func(t *testing.T) {
		holdings, platformGuess, coinList := buildHoldingsForPortfolio(t, geminiClient, apiClient, geminiKey, coingeckoKey, "portfolio1")
		tradeImages := portfolioTradeImagePaths(t, "portfolio1")
		multiTradeSeen := false

		for _, tradeImage := range tradeImages {
			tradeSlip := runTradeSlipOCR(t, geminiClient, geminiKey, tradeImage)
			debugLogJSON(t, "Trade slip OCR output", tradeSlip)
			if tradeSlip.ImageID == "" {
				t.Fatal("trade slip missing image_id")
			}
			if len(tradeSlip.Trades) == 0 {
				t.Fatal("trade slip missing trades")
			}
			if len(tradeSlip.Trades) > 1 {
				multiTradeSeen = true
			}

			tradeSlip.Trades = applyTradeSlipEdits(tradeSlip.Trades)

			deltaAssets := tradesToAssets(tradeSlip.Trades)
			if len(deltaAssets) == 0 {
				t.Fatal("expected delta assets from trade slip")
			}

			deltaHoldings, _ := resolveHoldings(t, apiClient, coingeckoKey, platformGuess, deltaAssets, coinList)
			before := holdings
			updated := aggregateHoldings(append(before, deltaHoldings...))
			assertHoldingsDelta(t, before, updated, tradeDeltaBySymbol(tradeSlip.Trades))
			holdings = updated
		}

		if !multiTradeSeen {
			t.Fatal("expected at least one trade slip with multiple trades")
		}
	})

	t.Run("AssetCommandPortfolio2", func(t *testing.T) {
		holdings, platformGuess, coinList := buildHoldingsForPortfolio(t, geminiClient, apiClient, geminiKey, coingeckoKey, "portfolio2")
		ethBefore := holdingAmount(holdings, "ETH")

		tradeTextPaths := portfolioTradeTextPaths(t, "portfolio2")
		commandText := strings.TrimSpace(string(mustReadFile(t, tradeTextPaths[0])))
		command := runAssetCommandParser(t, geminiClient, geminiKey, commandText)
		debugLogJSON(t, "Asset command output", command)
		if command.Intent != "UPDATE_ASSET" {
			t.Fatalf("unexpected intent: %s", command.Intent)
		}
		payloads := command.Payloads
		if len(payloads) == 0 {
			t.Fatal("asset command missing payload")
		}
		payload := payloads[0]
		if payload.TargetAsset == nil {
			t.Fatal("asset command missing target_asset")
		}
		if !strings.EqualFold(payload.TargetAsset.Ticker, "ETH") {
			t.Fatalf("expected ETH, got %q", payload.TargetAsset.Ticker)
		}
		if payload.TargetAsset.Amount == nil || !floatEquals(*payload.TargetAsset.Amount, 0.1) {
			t.Fatalf("expected amount 0.1, got %v", payload.TargetAsset.Amount)
		}
		if !strings.EqualFold(payload.TargetAsset.Action, "ADD") {
			t.Fatalf("expected action ADD, got %q", payload.TargetAsset.Action)
		}
		if payload.PricePerUnit != nil {
			t.Fatalf("expected price_per_unit null, got %v", *payload.PricePerUnit)
		}

		deltaAssets := commandToAssets(command)
		if len(deltaAssets) == 0 {
			t.Fatal("expected delta assets from asset command")
		}

		deltaHoldings, _ := resolveHoldings(t, apiClient, coingeckoKey, platformGuess, deltaAssets, coinList)
		updated := aggregateHoldings(append(holdings, deltaHoldings...))
		ethAfter := holdingAmount(updated, "ETH")
		if ethAfter <= ethBefore {
			t.Fatalf("expected ETH to increase after command: before=%v after=%v", ethBefore, ethAfter)
		}
	})
}

func TestAssetCommandSimplePhrases(t *testing.T) {
	geminiKey := requireEnv(t, "GEMINI_API_KEY")
	coingeckoKey := requireEnv(t, "COINGECKO_PRO_API_KEY")
	_ = requireEnv(t, "MARKETSTACK_ACCESS_KEY")

	geminiClient := &http.Client{Timeout: 60 * time.Second}
	apiClient := &http.Client{Timeout: 30 * time.Second}
	coinList := fetchCoinGeckoList(t, apiClient, coingeckoKey)

	cases := []struct {
		name          string
		commandText   string
		expectSymbol  string
		expectAmount  float64
		expectType    string
		expectPricing string
		minPrice      float64
		maxPrice      float64
	}{
		{
			name:          "BoughtBTC",
			commandText:   "Bought 1 BTC",
			expectSymbol:  "BTC",
			expectAmount:  1,
			expectType:    "crypto",
			expectPricing: "COINGECKO",
			minPrice:      30000,
			maxPrice:      300000,
		},
		{
			name:          "BoughtETH",
			commandText:   "Bought 10 ETH",
			expectSymbol:  "ETH",
			expectAmount:  10,
			expectType:    "crypto",
			expectPricing: "COINGECKO",
			minPrice:      1000,
			maxPrice:      10000,
		},
		{
			name:          "BoughtAAPL",
			commandText:   "Bought 10 AAPL",
			expectSymbol:  "AAPL",
			expectAmount:  10,
			expectType:    "stock",
			expectPricing: "MARKETSTACK",
			minPrice:      100,
			maxPrice:      1000,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			command := runAssetCommandParser(t, geminiClient, geminiKey, tc.commandText)
			if command.Intent != "UPDATE_ASSET" {
				t.Fatalf("unexpected intent: %s", command.Intent)
			}
			payloads := command.Payloads
			if len(payloads) == 0 {
				t.Fatal("asset command missing payload")
			}
			payload := payloads[0]
			if payload.TargetAsset == nil {
				t.Fatal("asset command missing target_asset")
			}
			if !strings.EqualFold(payload.TargetAsset.Ticker, tc.expectSymbol) {
				t.Fatalf("expected %s, got %q", tc.expectSymbol, payload.TargetAsset.Ticker)
			}
			if payload.TargetAsset.Amount == nil || !floatEquals(*payload.TargetAsset.Amount, tc.expectAmount) {
				t.Fatalf("expected amount %v, got %v", tc.expectAmount, payload.TargetAsset.Amount)
			}
			if !strings.EqualFold(payload.TargetAsset.Action, "ADD") {
				t.Fatalf("expected action ADD, got %q", payload.TargetAsset.Action)
			}
			if payload.PricePerUnit != nil {
				t.Fatalf("expected price_per_unit null, got %v", *payload.PricePerUnit)
			}

			symbol := strings.ToUpper(strings.TrimSpace(payload.TargetAsset.Ticker))
			amount := *payload.TargetAsset.Amount
			assetType := tc.expectType
			platformGuess := "binance"
			if assetType == "stock" {
				platformGuess = "fidelity"
			}

			assets := []ocrAsset{{
				SymbolRaw: symbol,
				Symbol:    &symbol,
				AssetType: assetType,
				Amount:    amount,
			}}

			holdings, _ := resolveHoldings(t, apiClient, coingeckoKey, platformGuess, assets, coinList)
			holdings = aggregateHoldings(holdings)
			found := false
			for _, holding := range holdings {
				if !strings.EqualFold(holding.Symbol, symbol) {
					continue
				}
				found = true
				if holding.Amount < amount {
					t.Fatalf("expected amount >= %v, got %v", amount, holding.Amount)
				}
				if holding.ValueUSD <= 0 {
					t.Fatalf("expected priced value for %s, got %v", symbol, holding.ValueUSD)
				}
				if holding.Amount <= 0 {
					t.Fatalf("expected positive amount for %s, got %v", symbol, holding.Amount)
				}
				unitPrice := holding.ValueUSD / holding.Amount
				if unitPrice < tc.minPrice || unitPrice > tc.maxPrice {
					t.Fatalf("unexpected price for %s: got %v, want [%v, %v]", symbol, unitPrice, tc.minPrice, tc.maxPrice)
				}
				if holding.ValuationStatus != "priced" {
					t.Fatalf("expected priced status for %s, got %s", symbol, holding.ValuationStatus)
				}
				if holding.PricingSource != tc.expectPricing {
					t.Fatalf("expected pricing source %s, got %s", tc.expectPricing, holding.PricingSource)
				}
			}
			if !found {
				t.Fatalf("holding for %s not found", symbol)
			}
		})
	}
}

func buildHoldingsForPortfolio(t *testing.T, geminiClient, apiClient *http.Client, geminiKey, coingeckoKey, portfolio string) ([]portfolioHolding, string, []coinGeckoCoinListEntry) {
	t.Helper()

	ocr := runOCR(t, geminiClient, geminiKey, portfolioImagePaths(t, portfolio))
	assets, platformGuess := collectOCRAssets(t, ocr.Images)
	if len(assets) == 0 {
		t.Fatal("expected OCR assets, got empty")
	}

	coinList := fetchCoinGeckoList(t, apiClient, coingeckoKey)
	holdings, _ := resolveHoldings(t, apiClient, coingeckoKey, platformGuess, assets, coinList)
	holdings = aggregateHoldings(holdings)
	return holdings, platformGuess, coinList
}

func runTradeSlipOCR(t *testing.T, client *http.Client, apiKey, imagePath string) tradeSlipOCRResponse {
	t.Helper()

	imageBytes := mustReadFile(t, imagePath)
	encodedImage := base64.StdEncoding.EncodeToString(imageBytes)

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: TradeSlipOCRPrompt}},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: "Input trade slip image:\n- image_id: img_1\nReturn strict JSON per the system prompt."},
					{InlineData: &geminiInlineData{MimeType: imageMimeType(imagePath), Data: encodedImage}},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.0,
			MaxOutputTokens:  geminiMaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	var parsed tradeSlipOCRResponse
	callGeminiJSON(t, client, apiKey, request, &parsed, "trade-slip-ocr")
	return parsed
}

func applyTradeSlipEdits(trades []tradeSlipTrade) []tradeSlipTrade {
	edited := make([]tradeSlipTrade, len(trades))
	for i, trade := range trades {
		edited[i] = trade
		if trade.Amount != nil {
			amount := roundTo(*trade.Amount, 6)
			edited[i].Amount = &amount
		}
		if trade.Price != nil {
			price := roundTo(*trade.Price, 2)
			edited[i].Price = &price
		}
		if trade.Currency != nil {
			currency := strings.ToUpper(strings.TrimSpace(*trade.Currency))
			if currency != "" {
				edited[i].Currency = &currency
			}
		}
	}
	return edited
}

func runAssetCommandParser(t *testing.T, client *http.Client, apiKey, text string) assetCommandResponse {
	t.Helper()

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: AssetCommandPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: text}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.0,
			MaxOutputTokens:  geminiMaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	var parsed assetCommandResponse
	callGeminiJSON(t, client, apiKey, request, &parsed, "asset-command")
	return parsed
}

func tradesToAssets(trades []tradeSlipTrade) []ocrAsset {
	assets := make([]ocrAsset, 0, len(trades))
	for _, trade := range trades {
		if trade.Symbol == "" || trade.Amount == nil || *trade.Amount == 0 {
			continue
		}
		amount := *trade.Amount
		if strings.EqualFold(trade.Side, "sell") {
			amount = -amount
		}
		assets = append(assets, ocrAsset{
			SymbolRaw: strings.ToUpper(strings.TrimSpace(trade.Symbol)),
			AssetType: "crypto",
			Amount:    amount,
		})
	}
	return assets
}

func commandToAssets(command assetCommandResponse) []ocrAsset {
	payloads := command.Payloads
	assets := make([]ocrAsset, 0, len(payloads))
	for _, payload := range payloads {
		if payload.TargetAsset == nil {
			continue
		}
		target := payload.TargetAsset
		if target.Ticker == "" || target.Amount == nil || *target.Amount == 0 {
			continue
		}
		amount := *target.Amount
		if strings.EqualFold(target.Action, "REMOVE") {
			amount = -amount
		}
		var avgPrice *float64
		if payload.PricePerUnit != nil {
			value := *payload.PricePerUnit
			avgPrice = &value
		}
		assets = append(assets, ocrAsset{
			SymbolRaw: strings.ToUpper(strings.TrimSpace(target.Ticker)),
			AssetType: "crypto",
			Amount:    amount,
			AvgPrice:  avgPrice,
		})
	}
	return assets
}

func holdingAmount(holdings []portfolioHolding, symbol string) float64 {
	for _, holding := range holdings {
		if strings.EqualFold(holding.Symbol, symbol) {
			return holding.Amount
		}
	}
	return 0
}

func tradeDeltaBySymbol(trades []tradeSlipTrade) map[string]float64 {
	deltas := make(map[string]float64)
	for _, trade := range trades {
		symbol := strings.ToUpper(strings.TrimSpace(trade.Symbol))
		if symbol == "" || trade.Amount == nil || *trade.Amount == 0 {
			continue
		}
		amount := *trade.Amount
		if strings.EqualFold(trade.Side, "sell") {
			amount = -amount
		}
		deltas[symbol] += amount
	}
	return deltas
}

func assertHoldingsDelta(t *testing.T, before, after []portfolioHolding, deltas map[string]float64) {
	t.Helper()

	for symbol, delta := range deltas {
		beforeAmount := holdingAmount(before, symbol)
		afterAmount := holdingAmount(after, symbol)
		if !floatEquals(afterAmount-beforeAmount, delta) {
			t.Fatalf("unexpected delta for %s: got %v want %v", symbol, afterAmount-beforeAmount, delta)
		}
	}
}
