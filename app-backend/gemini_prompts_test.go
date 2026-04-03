package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode"
)

const (
	geminiModel           = "gemini-3-flash-preview"
	geminiMaxOutputTokens = 65536 // 2048 truncates Gemini JSON output (MAX_TOKENS), causing incomplete OCR and report payloads.
	geminiMaxAttempts     = 3
	geminiRetryBaseDelay  = 2 * time.Second
)

func TestGeminiOCRPrompt(t *testing.T) {
	apiKey := requireEnv(t, "GEMINI_API_KEY")
	client := &http.Client{Timeout: 40 * time.Second}

	imagePaths := portfolioImagePaths(t, "portfolio1")
	imagePath := imagePaths[0]
	imageBytes := mustReadFile(t, imagePath)
	encodedImage := base64.StdEncoding.EncodeToString(imageBytes)

	request := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: OCRPortfolioPrompt}},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: "Input images:\n- image_id: img_1\nReturn strict JSON per the system prompt."},
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

	var parsed ocrPromptResponse
	callGeminiJSON(t, client, apiKey, request, &parsed, "OCR")
	debugLogJSON(t, "OCR output", parsed)
	if len(parsed.Images) == 0 {
		t.Fatal("expected OCR images, got empty")
	}

	image := findOCRImage(parsed.Images, "img_1")
	if image == nil {
		t.Fatalf("expected image_id img_1, got: %#v", parsed.Images)
	}
	debugLogf(t, "OCR image_id=%s platform=%s assets=%d", image.ImageID, image.PlatformGuess, len(image.Assets))
	if image.Status != "success" {
		t.Fatalf("expected status success, got %q", image.Status)
	}
	if image.PlatformGuess == "" {
		t.Fatalf("expected platform_guess, got empty: %#v", image)
	}
	if len(image.Assets) == 0 {
		t.Fatalf("expected assets for success image, got empty: %#v", image)
	}
	debugLogJSON(t, "OCR assets", image.Assets)
	for _, asset := range image.Assets {
		if asset.SymbolRaw == "" || asset.AssetType == "" || asset.Amount == 0 {
			t.Fatalf("unexpected asset fields: %#v", asset)
		}
	}
}

func TestGeminiPreviewPrompt(t *testing.T) {
	apiKey := requireEnv(t, "GEMINI_API_KEY")
	client := &http.Client{Timeout: 40 * time.Second}

	inputPath := filepath.Join("testdata", "prompts", "preview_input.json")
	inputBytes := mustReadFile(t, inputPath)

	var input previewInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		t.Fatalf("parse preview input JSON: %v", err)
	}

	tests := []struct {
		name           string
		outputLanguage string
		expectHan      bool
	}{
		{name: "english", outputLanguage: "English", expectHan: false},
		{name: "chinese", outputLanguage: "Simplified Chinese", expectHan: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			request := geminiRequest{
				SystemInstruction: &geminiSystemInstruction{
					Parts: []geminiPart{{Text: applyOutputLanguage(PreviewReportPrompt, test.outputLanguage)}},
				},
				Contents: []geminiContent{
					{
						Role:  "user",
						Parts: []geminiPart{{Text: string(inputBytes)}},
					},
				},
				GenerationConfig: geminiGenerationConfig{
					Temperature:      0.4,
					MaxOutputTokens:  geminiMaxOutputTokens,
					ResponseMimeType: "application/json",
				},
			}

			var parsed previewPromptOutput
			callGeminiJSON(t, client, apiKey, request, &parsed, "preview-"+test.name)
			normalizePreviewStatus(&parsed)
			debugLogJSON(t, "Preview output "+test.name, parsed)
			if parsed.MetaData.CalculationID != input.MetaData.CalculationID {
				t.Fatalf("calculation_id mismatch: got %q want %q", parsed.MetaData.CalculationID, input.MetaData.CalculationID)
			}
			if parsed.ValuationAsOf == "" || parsed.ValuationAsOf != input.ValuationAsOf {
				t.Fatalf("valuation_as_of mismatch: got %q want %q", parsed.ValuationAsOf, input.ValuationAsOf)
			}
			if parsed.MarketDataSnapshotID == "" || parsed.MarketDataSnapshotID != input.MarketDataSnapshotID {
				t.Fatalf("market_data_snapshot_id mismatch: got %q want %q", parsed.MarketDataSnapshotID, input.MarketDataSnapshotID)
			}
			debugLogf(t, "Preview health_score=%d volatility_score=%d risks=%d", parsed.FixedMetrics.HealthScore, parsed.FixedMetrics.VolatilityScore, len(parsed.IdentifiedRisks))
			if !floatEquals(parsed.FixedMetrics.NetWorthUSD, input.FixedMetrics.NetWorthUSD) {
				t.Fatalf("net_worth_usd mismatch: got %v want %v", parsed.FixedMetrics.NetWorthUSD, input.FixedMetrics.NetWorthUSD)
			}
			if parsed.FixedMetrics.HealthScore != input.FixedMetrics.HealthScore {
				t.Fatalf("health_score mismatch: got %d want %d", parsed.FixedMetrics.HealthScore, input.FixedMetrics.HealthScore)
			}
			if parsed.FixedMetrics.HealthStatus != input.FixedMetrics.HealthStatus {
				t.Fatalf("health_status mismatch: got %q want %q", parsed.FixedMetrics.HealthStatus, input.FixedMetrics.HealthStatus)
			}
			if parsed.FixedMetrics.VolatilityScore != input.FixedMetrics.VolatilityScore {
				t.Fatalf("volatility_score mismatch: got %d want %d", parsed.FixedMetrics.VolatilityScore, input.FixedMetrics.VolatilityScore)
			}
			if !floatEquals(parsed.NetWorthDisplay, input.NetWorthDisplay) {
				t.Fatalf("net_worth_display mismatch: got %v want %v", parsed.NetWorthDisplay, input.NetWorthDisplay)
			}
			if parsed.BaseCurrency != input.BaseCurrency {
				t.Fatalf("base_currency mismatch: got %q want %q", parsed.BaseCurrency, input.BaseCurrency)
			}
			if !floatEquals(parsed.BaseFXRateToUSD, input.BaseFXRateToUSD) {
				t.Fatalf("base_fx_rate_to_usd mismatch: got %v want %v", parsed.BaseFXRateToUSD, input.BaseFXRateToUSD)
			}
			if len(parsed.IdentifiedRisks) != 3 {
				t.Fatalf("expected 3 identified risks, got %d", len(parsed.IdentifiedRisks))
			}
			assertUniqueRiskTeasers(t, parsed.IdentifiedRisks)
			if parsed.LockedProjection.PotentialUpside == "" || parsed.LockedProjection.CTA == "" {
				t.Fatalf("expected locked_projection fields, got %#v", parsed.LockedProjection)
			}
			assertPreviewLanguage(t, parsed, test.expectHan)
		})
	}
}

func TestGeminiPaidPrompt(t *testing.T) {
	apiKey := requireEnv(t, "GEMINI_API_KEY")
	client := &http.Client{Timeout: 40 * time.Second}

	inputPath := filepath.Join("testdata", "prompts", "paid_input.json")
	inputBytes := mustReadFile(t, inputPath)

	var input paidInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		t.Fatalf("parse paid input JSON: %v", err)
	}

	tests := []struct {
		name           string
		outputLanguage string
		expectHan      bool
	}{
		{name: "english", outputLanguage: "English", expectHan: false},
		{name: "chinese", outputLanguage: "Simplified Chinese", expectHan: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			request := geminiRequest{
				SystemInstruction: &geminiSystemInstruction{
					Parts: []geminiPart{{Text: applyOutputLanguage(PaidReportPrompt, test.outputLanguage)}},
				},
				Contents: []geminiContent{
					{
						Role:  "user",
						Parts: []geminiPart{{Text: string(inputBytes)}},
					},
				},
				GenerationConfig: geminiGenerationConfig{
					Temperature:      0.4,
					MaxOutputTokens:  geminiMaxOutputTokens,
					ResponseMimeType: "application/json",
				},
			}

			var parsed paidPromptOutput
			callGeminiJSON(t, client, apiKey, request, &parsed, "paid-"+test.name)
			normalizePaidStatus(&parsed)
			debugLogJSON(t, "Paid output "+test.name, parsed)
			if parsed.MetaData.CalculationID != input.PreviousTeaser.MetaData.CalculationID {
				t.Fatalf("paid calculation_id mismatch: got %q want %q", parsed.MetaData.CalculationID, input.PreviousTeaser.MetaData.CalculationID)
			}
			if parsed.ValuationAsOf == "" || parsed.ValuationAsOf != input.PreviousTeaser.ValuationAsOf {
				t.Fatalf("paid valuation_as_of mismatch: got %q want %q", parsed.ValuationAsOf, input.PreviousTeaser.ValuationAsOf)
			}
			if parsed.MarketDataSnapshotID == "" || parsed.MarketDataSnapshotID != input.PreviousTeaser.MarketDataSnapshotID {
				t.Fatalf("paid market_data_snapshot_id mismatch: got %q want %q", parsed.MarketDataSnapshotID, input.PreviousTeaser.MarketDataSnapshotID)
			}
			if parsed.ReportHeader.HealthScore.Value != input.PreviousTeaser.FixedMetrics.HealthScore {
				t.Fatalf("health_score mismatch: got %d want %d", parsed.ReportHeader.HealthScore.Value, input.PreviousTeaser.FixedMetrics.HealthScore)
			}
			if parsed.ReportHeader.VolatilityDashboard.Value != input.PreviousTeaser.FixedMetrics.VolatilityScore {
				t.Fatalf("volatility_score mismatch: got %d want %d", parsed.ReportHeader.VolatilityDashboard.Value, input.PreviousTeaser.FixedMetrics.VolatilityScore)
			}
			if !floatEquals(parsed.NetWorthDisplay, input.NetWorthDisplay) {
				t.Fatalf("net_worth_display mismatch: got %v want %v", parsed.NetWorthDisplay, input.NetWorthDisplay)
			}
			if parsed.BaseCurrency != input.BaseCurrency {
				t.Fatalf("base_currency mismatch: got %q want %q", parsed.BaseCurrency, input.BaseCurrency)
			}
			if !floatEquals(parsed.BaseFXRateToUSD, input.BaseFXRateToUSD) {
				t.Fatalf("base_fx_rate_to_usd mismatch: got %v want %v", parsed.BaseFXRateToUSD, input.BaseFXRateToUSD)
			}
			expectedHealthStatus := scoreStatusFromScore(parsed.ReportHeader.HealthScore.Value)
			if parsed.ReportHeader.HealthScore.Status != expectedHealthStatus {
				t.Fatalf("health_score status mismatch: got %q want %q", parsed.ReportHeader.HealthScore.Status, expectedHealthStatus)
			}
			expectedVolStatus := volatilityStatusFromScore(parsed.ReportHeader.VolatilityDashboard.Value)
			if parsed.ReportHeader.VolatilityDashboard.Status != expectedVolStatus {
				t.Fatalf("volatility_dashboard status mismatch: got %q want %q", parsed.ReportHeader.VolatilityDashboard.Status, expectedVolStatus)
			}
			debugLogf(t, "Paid health_score=%d risk_insights=%d plans=%d", parsed.ReportHeader.HealthScore.Value, len(parsed.RiskInsights), len(parsed.OptimizationPlan))

			expectedRisks := make(map[string]struct{}, len(input.PreviousTeaser.IdentifiedRisks))
			for _, risk := range input.PreviousTeaser.IdentifiedRisks {
				expectedRisks[risk.RiskID] = struct{}{}
			}
			if len(parsed.RiskInsights) != 3 {
				t.Fatalf("expected 3 risk_insights, got %d", len(parsed.RiskInsights))
			}
			for _, risk := range parsed.RiskInsights {
				if _, ok := expectedRisks[risk.RiskID]; !ok {
					t.Fatalf("unexpected risk_id %q", risk.RiskID)
				}
			}

			expectedPlans := make(map[string]struct{}, len(input.LockedPlans))
			for _, plan := range input.LockedPlans {
				expectedPlans[plan.PlanID] = struct{}{}
			}
			for _, plan := range parsed.OptimizationPlan {
				if _, ok := expectedPlans[plan.PlanID]; !ok {
					t.Fatalf("unexpected plan_id %q", plan.PlanID)
				}
				delete(expectedPlans, plan.PlanID)
				if _, ok := expectedRisks[plan.LinkedRiskID]; !ok {
					t.Fatalf("optimization_plan linked_risk_id not in preview risks: %q", plan.LinkedRiskID)
				}
				if plan.Rationale == "" || plan.ExpectedOutcome == "" {
					t.Fatalf("expected rationale and expected_outcome: %#v", plan)
				}
			}
			if len(expectedPlans) != 0 {
				t.Fatalf("missing optimization plans: %v", mapKeys(expectedPlans))
			}
			if parsed.TheVerdict.ConstructiveComment == "" {
				t.Fatalf("expected constructive_comment, got %#v", parsed.TheVerdict)
			}
			assertPaidLanguage(t, parsed, test.expectHan)
		})
	}
}

func assertUniqueRiskTeasers(t *testing.T, risks []previewRisk) {
	t.Helper()
	seen := make(map[string]struct{}, len(risks))
	for _, risk := range risks {
		teaser := strings.TrimSpace(risk.TeaserText)
		if teaser == "" {
			t.Fatalf("empty teaser for risk_id=%s type=%s", risk.RiskID, risk.Type)
		}
		key := strings.ToLower(teaser)
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate teaser detected: %q", teaser)
		}
		seen[key] = struct{}{}
	}
}

type geminiCallResult struct {
	Text          string
	Duration      time.Duration
	StatusCode    int
	ContentLength int64
	FinishReason  string
	HeaderSummary string
}

func callGemini(t *testing.T, client *http.Client, apiKey string, request geminiRequest) geminiCallResult {
	t.Helper()

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		geminiModel,
		url.QueryEscape(apiKey),
	)

	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var lastErr error
	for attempt := 0; attempt < geminiMaxAttempts; attempt++ {
		httpRequest, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("build request: %v", err)
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRequest.Header.Set("Accept", "application/json")

		start := time.Now()
		resp, err := client.Do(httpRequest)
		if err != nil {
			lastErr = err
			if attempt < geminiMaxAttempts-1 {
				debugLogf(t, "Gemini request retrying after error: %v", err)
				time.Sleep(geminiRetryBaseDelay * time.Duration(1<<attempt))
				continue
			}
			t.Fatalf("gemini request failed after %d attempts: %v", geminiMaxAttempts, err)
		}

		result := geminiCallResult{
			Duration:      time.Since(start),
			StatusCode:    resp.StatusCode,
			ContentLength: resp.ContentLength,
			HeaderSummary: summarizeGeminiHeaders(resp.Header),
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("gemini status %d: %s", resp.StatusCode, string(body))
			if attempt < geminiMaxAttempts-1 && isRetryableGeminiStatus(resp.StatusCode) {
				debugLogf(t, "Gemini request retrying after status %d", resp.StatusCode)
				time.Sleep(geminiRetryBaseDelay * time.Duration(1<<attempt))
				continue
			}
			t.Fatalf("gemini status %d: %s", resp.StatusCode, string(body))
		}

		var parsed geminiResponse
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&parsed); err != nil {
			_ = resp.Body.Close()
			t.Fatalf("decode gemini response: %v", err)
		}
		_ = resp.Body.Close()

		text := extractGeminiText(t, parsed)
		result.Text = strings.TrimSpace(text)
		result.FinishReason = extractGeminiFinishReason(parsed)
		return result
	}

	t.Fatalf("gemini request failed after %d attempts: %v", geminiMaxAttempts, lastErr)
	return geminiCallResult{}
}

func callGeminiJSON(t *testing.T, client *http.Client, apiKey string, request geminiRequest, out any, label string) string {
	t.Helper()
	result := callGemini(t, client, apiKey, request)
	debugLogf(t, "Gemini %s attempt=0 duration=%s bytes=%d status=%d content_length=%d finish_reason=%s headers=%s", label, result.Duration, len(result.Text), result.StatusCode, result.ContentLength, result.FinishReason, result.HeaderSummary)
	if err := json.Unmarshal([]byte(result.Text), out); err == nil {
		return result.Text
	} else {
		lastErr := err
		lastResponse := result.Text
		for attempt := 1; attempt <= 2; attempt++ {
			retryRequest := geminiRetryRequest(request, lastResponse)
			result = callGemini(t, client, apiKey, retryRequest)
			debugLogf(t, "Gemini %s attempt=%d duration=%s bytes=%d status=%d content_length=%d finish_reason=%s headers=%s", label, attempt, result.Duration, len(result.Text), result.StatusCode, result.ContentLength, result.FinishReason, result.HeaderSummary)
			lastResponse = result.Text
			if err := json.Unmarshal([]byte(lastResponse), out); err == nil {
				return lastResponse
			} else {
				lastErr = err
			}
		}
		t.Fatalf("parse %s response JSON: %v\nraw: %s", label, lastErr, lastResponse)
	}
	return ""
}

func geminiRetryRequest(request geminiRequest, previousResponse string) geminiRequest {
	retry := request
	retry.Contents = append(append([]geminiContent(nil), request.Contents...), geminiContent{
		Role: "user",
		Parts: []geminiPart{{
			Text: fmt.Sprintf("The previous response was invalid JSON. Return corrected JSON only. Previous response:\n%s", previousResponse),
		}},
	})
	return retry
}

func isRetryableGeminiStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func summarizeGeminiHeaders(header http.Header) string {
	entries := []struct {
		Key   string
		Label string
	}{
		{Key: "Content-Type", Label: "content_type"},
		{Key: "X-Request-Id", Label: "x_request_id"},
		{Key: "X-Goog-Request-Id", Label: "x_goog_request_id"},
		{Key: "X-Goog-Trace-Id", Label: "x_goog_trace_id"},
		{Key: "X-Goog-Generation-Id", Label: "x_goog_generation_id"},
	}

	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		if value := strings.TrimSpace(header.Get(entry.Key)); value != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", entry.Label, value))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

func extractGeminiText(t *testing.T, response geminiResponse) string {
	t.Helper()
	if len(response.Candidates) == 0 {
		t.Fatal("gemini response missing candidates")
	}

	candidate := response.Candidates[0]
	var builder strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			builder.WriteString(part.Text)
		}
	}
	text := builder.String()
	if strings.TrimSpace(text) == "" {
		t.Fatalf("gemini response missing text parts: %#v", candidate.Content.Parts)
	}
	return text
}

func extractGeminiFinishReason(response geminiResponse) string {
	if len(response.Candidates) == 0 {
		return "missing_candidates"
	}
	reason := strings.TrimSpace(response.Candidates[0].FinishReason)
	if reason == "" {
		return "unknown"
	}
	return reason
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return data
}

func floatEquals(a, b float64) bool {
	const epsilon = 0.0001
	if a > b {
		return a-b < epsilon
	}
	return b-a < epsilon
}

func withinRange(value, baseline, delta int) bool {
	return value >= baseline-delta && value <= baseline+delta
}

func findOCRImage(images []ocrImage, imageID string) *ocrImage {
	for i := range images {
		if images[i].ImageID == imageID {
			return &images[i]
		}
	}
	return nil
}

func assertPreviewLanguage(t *testing.T, preview previewPromptOutput, expectHan bool) {
	t.Helper()

	for i, risk := range preview.IdentifiedRisks {
		assertLocalizedField(t, fmt.Sprintf("identified_risks[%d].teaser_text", i), risk.TeaserText, expectHan)
	}
	assertLocalizedField(t, "locked_projection.potential_upside", preview.LockedProjection.PotentialUpside, expectHan)
	assertLocalizedField(t, "locked_projection.cta", preview.LockedProjection.CTA, expectHan)
}

func assertPaidLanguage(t *testing.T, paid paidPromptOutput, expectHan bool) {
	t.Helper()

	for i, risk := range paid.RiskInsights {
		assertLocalizedField(t, fmt.Sprintf("risk_insights[%d].message", i), risk.Message, expectHan)
	}
	for i, plan := range paid.OptimizationPlan {
		if strings.TrimSpace(plan.ExecutionSummary) == "" {
			t.Fatalf("expected optimization_plan[%d].execution_summary, got empty", i)
		}
		assertLocalizedField(t, fmt.Sprintf("optimization_plan[%d].execution_summary", i), plan.ExecutionSummary, expectHan)
		assertLocalizedField(t, fmt.Sprintf("optimization_plan[%d].rationale", i), plan.Rationale, expectHan)
		assertLocalizedField(t, fmt.Sprintf("optimization_plan[%d].expected_outcome", i), plan.ExpectedOutcome, expectHan)
	}
	assertLocalizedField(t, "the_verdict.constructive_comment", paid.TheVerdict.ConstructiveComment, expectHan)
}

func assertLocalizedField(t *testing.T, name, value string, expectHan bool) {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		t.Fatalf("expected %s text, got empty", name)
	}
	hasHan := containsHan(value)
	if expectHan && !hasHan {
		t.Fatalf("expected %s to contain Chinese characters: %q", name, value)
	}
	if !expectHan && hasHan {
		t.Fatalf("expected %s to be English-only: %q", name, value)
	}
}

func containsHan(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"systemInstruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
	GenerationConfig  geminiGenerationConfig   `json:"generationConfig,omitempty"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature,omitempty"`
	MaxOutputTokens  int     `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type ocrPromptResponse struct {
	Images []ocrImage `json:"images"`
}

type ocrImage struct {
	ImageID       string     `json:"image_id"`
	Status        string     `json:"status"`
	ErrorReason   *string    `json:"error_reason"`
	PlatformGuess string     `json:"platform_guess"`
	Assets        []ocrAsset `json:"assets"`
}

type ocrAsset struct {
	SymbolRaw           string   `json:"symbol_raw"`
	Symbol              *string  `json:"symbol"`
	AssetType           string   `json:"asset_type"`
	Amount              float64  `json:"amount"`
	ValueFromScreenshot *float64 `json:"value_from_screenshot"`
	DisplayCurrency     *string  `json:"display_currency"`
	PNLPercent          *float64 `json:"pnl_percent"`
	AvgPrice            *float64 `json:"avg_price"`
}

type previewInput struct {
	MetaData struct {
		CalculationID string `json:"calculation_id"`
	} `json:"meta_data"`
	ValuationAsOf        string `json:"valuation_as_of"`
	MarketDataSnapshotID string `json:"market_data_snapshot_id"`
	ComputedMetrics      struct {
		NetWorthUSD             float64 `json:"net_worth_usd"`
		TopAssetPct             float64 `json:"top_asset_pct"`
		HealthScoreBaseline     int     `json:"health_score_baseline"`
		VolatilityScoreBaseline int     `json:"volatility_score_baseline"`
	} `json:"computed_metrics"`
	FixedMetrics    previewFixedMetrics `json:"fixed_metrics"`
	NetWorthDisplay float64             `json:"net_worth_display"`
	BaseCurrency    string              `json:"base_currency"`
	BaseFXRateToUSD float64             `json:"base_fx_rate_to_usd"`
}

type previewPromptOutput struct {
	MetaData             metaData            `json:"meta_data"`
	ValuationAsOf        string              `json:"valuation_as_of"`
	MarketDataSnapshotID string              `json:"market_data_snapshot_id"`
	FixedMetrics         previewFixedMetrics `json:"fixed_metrics"`
	NetWorthDisplay      float64             `json:"net_worth_display"`
	BaseCurrency         string              `json:"base_currency"`
	BaseFXRateToUSD      float64             `json:"base_fx_rate_to_usd"`
	IdentifiedRisks      []previewRisk       `json:"identified_risks"`
	LockedProjection     lockedProjection    `json:"locked_projection"`
}

type metaData struct {
	CalculationID string `json:"calculation_id"`
}

type previewFixedMetrics struct {
	NetWorthUSD     float64 `json:"net_worth_usd"`
	HealthScore     int     `json:"health_score"`
	HealthStatus    string  `json:"health_status"`
	VolatilityScore int     `json:"volatility_score"`
}

type previewRisk struct {
	RiskID     string `json:"risk_id"`
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	TeaserText string `json:"teaser_text"`
}

type lockedProjection struct {
	PotentialUpside string `json:"potential_upside"`
	CTA             string `json:"cta"`
}

type paidInput struct {
	PreviousTeaser struct {
		FixedMetrics         previewFixedMetrics `json:"fixed_metrics"`
		ValuationAsOf        string              `json:"valuation_as_of"`
		MarketDataSnapshotID string              `json:"market_data_snapshot_id"`
		NetWorthDisplay      float64             `json:"net_worth_display"`
		BaseCurrency         string              `json:"base_currency"`
		BaseFXRateToUSD      float64             `json:"base_fx_rate_to_usd"`
		IdentifiedRisks      []previewRiskInput  `json:"identified_risks"`
		MetaData             metaData            `json:"meta_data"`
	} `json:"previous_teaser"`
	LockedPlans     []lockedPlanInput   `json:"locked_plans"`
	FixedMetrics    previewFixedMetrics `json:"fixed_metrics"`
	NetWorthDisplay float64             `json:"net_worth_display"`
	BaseCurrency    string              `json:"base_currency"`
	BaseFXRateToUSD float64             `json:"base_fx_rate_to_usd"`
}

type previewRiskInput struct {
	RiskID string `json:"risk_id"`
}

type lockedPlanInput struct {
	PlanID string `json:"plan_id"`
}

type paidPromptOutput struct {
	MetaData             metaData         `json:"meta_data"`
	ValuationAsOf        string           `json:"valuation_as_of"`
	MarketDataSnapshotID string           `json:"market_data_snapshot_id"`
	NetWorthDisplay      float64          `json:"net_worth_display"`
	BaseCurrency         string           `json:"base_currency"`
	BaseFXRateToUSD      float64          `json:"base_fx_rate_to_usd"`
	ReportHeader         paidReportHeader `json:"report_header"`
	Charts               paidCharts       `json:"charts"`
	RiskInsights         []paidRisk       `json:"risk_insights"`
	OptimizationPlan     []paidPlan       `json:"optimization_plan"`
	TheVerdict           paidVerdict      `json:"the_verdict"`
}

type paidReportHeader struct {
	HealthScore         scoredValue `json:"health_score"`
	VolatilityDashboard scoredValue `json:"volatility_dashboard"`
}

type scoredValue struct {
	Value  int    `json:"value"`
	Status string `json:"status"`
}

type paidCharts struct {
	RadarChart paidRadarChart `json:"radar_chart"`
}

type paidRadarChart struct {
	Liquidity       int `json:"liquidity"`
	Diversification int `json:"diversification"`
	Alpha           int `json:"alpha"`
	Drawdown        int `json:"drawdown"`
}

type paidRisk struct {
	RiskID   string `json:"risk_id"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type paidPlan struct {
	PlanID           string         `json:"plan_id"`
	StrategyID       string         `json:"strategy_id"`
	AssetType        string         `json:"asset_type"`
	Symbol           string         `json:"symbol"`
	AssetKey         string         `json:"asset_key"`
	LinkedRiskID     string         `json:"linked_risk_id"`
	Parameters       map[string]any `json:"parameters,omitempty"`
	ExecutionSummary string         `json:"execution_summary,omitempty"`
	Rationale        string         `json:"rationale"`
	ExpectedOutcome  string         `json:"expected_outcome"`
}

type paidVerdict struct {
	ConstructiveComment string `json:"constructive_comment"`
}
