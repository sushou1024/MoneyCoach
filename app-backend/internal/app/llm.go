package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	geminiModel           = "gemini-3-flash-preview"
	geminiMaxOutputTokens = 65536
	geminiMaxAttempts     = 3
	geminiRetryBaseDelay  = 2 * time.Second
)

type geminiClient struct {
	apiKey string
	client *http.Client
}

func newGeminiClient(apiKey string, client *http.Client) *geminiClient {
	return &geminiClient{apiKey: apiKey, client: client}
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

type geminiCallResult struct {
	Text         string
	StatusCode   int
	FinishReason string
}

func (g *geminiClient) callGemini(ctx context.Context, request geminiRequest) (geminiCallResult, error) {
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		geminiModel,
		url.QueryEscape(g.apiKey),
	)

	payload, err := json.Marshal(request)
	if err != nil {
		return geminiCallResult{}, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < geminiMaxAttempts; attempt++ {
		httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return geminiCallResult{}, fmt.Errorf("build request: %w", err)
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRequest.Header.Set("Accept", "application/json")

		resp, err := g.client.Do(httpRequest)
		if err != nil {
			lastErr = err
			if attempt < geminiMaxAttempts-1 {
				select {
				case <-time.After(geminiRetryBaseDelay * time.Duration(1<<attempt)):
					continue
				case <-ctx.Done():
					return geminiCallResult{}, ctx.Err()
				}
			}
			return geminiCallResult{}, fmt.Errorf("gemini request failed: %w", err)
		}

		result := geminiCallResult{StatusCode: resp.StatusCode}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("gemini status %d: %s", resp.StatusCode, string(body))
			if attempt < geminiMaxAttempts-1 && isRetryableGeminiStatus(resp.StatusCode) {
				select {
				case <-time.After(geminiRetryBaseDelay * time.Duration(1<<attempt)):
					continue
				case <-ctx.Done():
					return geminiCallResult{}, ctx.Err()
				}
			}
			return geminiCallResult{}, lastErr
		}

		var parsed geminiResponse
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&parsed); err != nil {
			_ = resp.Body.Close()
			return geminiCallResult{}, fmt.Errorf("decode response: %w", err)
		}
		_ = resp.Body.Close()

		text := extractGeminiText(parsed)
		if strings.TrimSpace(text) == "" {
			return geminiCallResult{}, fmt.Errorf("gemini response empty")
		}

		result.Text = strings.TrimSpace(text)
		result.FinishReason = extractGeminiFinishReason(parsed)
		return result, nil
	}

	return geminiCallResult{}, fmt.Errorf("gemini request failed after retries: %v", lastErr)
}

func (g *geminiClient) callGeminiJSON(ctx context.Context, request geminiRequest, out any) (string, error) {
	result, err := g.callGemini(ctx, request)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal([]byte(result.Text), out); err == nil {
		return result.Text, nil
	}

	lastErr := fmt.Errorf("parse response JSON: %w", err)
	lastResponse := result.Text
	for attempt := 1; attempt <= 2; attempt++ {
		retryRequest := geminiRetryRequest(request, lastResponse)
		result, err = g.callGemini(ctx, retryRequest)
		if err != nil {
			lastErr = err
			continue
		}
		lastResponse = result.Text
		if err := json.Unmarshal([]byte(lastResponse), out); err == nil {
			return lastResponse, nil
		} else {
			lastErr = err
		}
	}
	return "", lastErr
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

func extractGeminiText(response geminiResponse) string {
	if len(response.Candidates) == 0 {
		return ""
	}

	candidate := response.Candidates[0]
	var builder strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			builder.WriteString(part.Text)
		}
	}
	return builder.String()
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
