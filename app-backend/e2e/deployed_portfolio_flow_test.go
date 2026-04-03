package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf16"
)

type e2eConfig struct {
	baseURL           string
	enableEntitlement bool
	deviceTimezone    string
	pollInterval      time.Duration
	stageTimeout      time.Duration
}

type uploadBatchCreateRequest struct {
	Purpose        string               `json:"purpose"`
	ImageCount     int                  `json:"image_count"`
	Images         []uploadBatchImageIn `json:"images"`
	DeviceTimezone string               `json:"device_timezone"`
}

type uploadBatchImageIn struct {
	FileName  string `json:"file_name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type uploadBatchCreateResponse struct {
	UploadBatchID string              `json:"upload_batch_id"`
	Status        string              `json:"status"`
	ImageUploads  []uploadBatchUpload `json:"image_uploads"`
	ExpiresAt     string              `json:"expires_at"`
}

type uploadBatchUpload struct {
	ImageID   string            `json:"image_id"`
	UploadURL string            `json:"upload_url"`
	Headers   map[string]string `json:"headers"`
}

type uploadBatchCompleteRequest struct {
	ImageIDs       []string `json:"image_ids"`
	ClientChecksum string   `json:"client_checksum"`
}

type uploadBatchReviewRequest struct {
	PlatformOverrides  []any `json:"platform_overrides"`
	Resolutions        []any `json:"resolutions"`
	Edits              []any `json:"edits"`
	DuplicateOverrides []any `json:"duplicate_overrides"`
}

type authEmailRegisterStartRequest struct {
	Email string `json:"email"`
}

type authEmailRegisterStartResponse struct {
	Sent bool   `json:"sent"`
	Code string `json:"code"`
}

type authEmailRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type authEmailLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type stageTimings struct {
	Portfolio string
	OCR       time.Duration
	Normalize time.Duration
	Preview   time.Duration
	Paid      time.Duration
	Total     time.Duration
}

type apiClient struct {
	baseURL      string
	token        string
	refreshToken string
	client       *http.Client
}

var (
	runIDOnce   sync.Once
	cachedRunID string
)

func TestDeployedPortfolioFlow(t *testing.T) {
	cfg := loadE2EConfig(t)
	runID := e2eRunID()

	portfolios := []struct {
		name       string
		imagePaths []string
	}{
		{
			name: "portfolio1",
			imagePaths: []string{
				portfolioImage("portfolio1", "images", "bybit.png"),
				portfolioImage("portfolio1", "images", "okx.png"),
			},
		},
		{
			name: "portfolio2",
			imagePaths: []string{
				portfolioImage("portfolio2", "images", "okx.png"),
			},
		},
		{
			name: "portfolio3",
			imagePaths: []string{
				portfolioImage("portfolio3", "images", "futu.png"),
			},
		},
		{
			name: "portfolio4",
			imagePaths: []string{
				portfolioImage("portfolio4", "images", "bybit.png"),
				portfolioImage("portfolio4", "images", "futu.png"),
				portfolioImage("portfolio4", "images", "okx.png"),
			},
		},
		{
			name: "portfolio5",
			imagePaths: []string{
				portfolioImage("portfolio5", "images", "hk.png"),
			},
		},
	}
	currencies := []string{"USD", "EUR", "CNY"}

	for _, currency := range currencies {
		for _, portfolio := range portfolios {
			currency := currency
			portfolio := portfolio
			t.Run(fmt.Sprintf("%s-%s", portfolio.name, strings.ToLower(currency)), func(t *testing.T) {
				email := fmt.Sprintf("e2e-%s-%s-%s@example.com", portfolio.name, strings.ToLower(currency), runID)
				client, err := loginWithEmail(cfg.baseURL, email)
				if err != nil {
					t.Fatalf("login %s: %v", email, err)
				}
				if err := clearDevEntitlement(client); err != nil {
					t.Fatalf("clear dev entitlement %s: %v", email, err)
				}
				if entitlement, err := fetchEntitlement(client); err == nil {
					auditLogJSON(t, "entitlement_after_clear", entitlement)
				}
				if err := updateBaseCurrency(client, currency); err != nil {
					t.Fatalf("set base currency %s: %v", currency, err)
				}
				t.Logf("base_currency set=%s email=%s", currency, email)

				previewTiming, calcID, err := runPreviewFlow(t, client, cfg, portfolio.name, portfolio.imagePaths, currency)
				if err != nil {
					t.Fatalf("preview flow %s: %v", portfolio.name, err)
				}
				if err := requirePaidEntitlement(client, calcID); err != nil {
					t.Fatalf("paywall %s: %v", portfolio.name, err)
				}
				if cfg.enableEntitlement {
					if err := client.postJSON("/v1/billing/dev/entitlement", map[string]any{}, nil); err != nil {
						t.Fatalf("enable entitlement %s: %v", email, err)
					}
					if entitlement, err := fetchEntitlement(client); err == nil {
						auditLogJSON(t, "entitlement_after_enable", entitlement)
					}
				}
				paidTiming, err := runPaidReport(t, client, cfg, portfolio.name, calcID, currency)
				if err != nil {
					t.Fatalf("paid flow %s: %v", portfolio.name, err)
				}

				directTiming, _, err := runDirectPaidFlow(t, client, cfg, portfolio.name, portfolio.imagePaths, currency)
				if err != nil {
					t.Fatalf("direct paid flow %s: %v", portfolio.name, err)
				}

				t.Logf("flow=preview portfolio=%s currency=%s ocr=%s normalize=%s preview=%s total=%s", portfolio.name, currency, previewTiming.OCR, previewTiming.Normalize, previewTiming.Preview, previewTiming.Total)
				t.Logf("flow=paid-after-upgrade portfolio=%s currency=%s paid=%s total=%s", portfolio.name, currency, paidTiming.Paid, paidTiming.Total)
				t.Logf("flow=direct-paid portfolio=%s currency=%s ocr=%s normalize=%s paid=%s total=%s", portfolio.name, currency, directTiming.OCR, directTiming.Normalize, directTiming.Paid, directTiming.Total)
			})
		}
	}
}

func TestDeployedPortfolioCurrencySwitch(t *testing.T) {
	cfg := loadE2EConfig(t)
	runID := e2eRunID()
	portfolio := struct {
		name       string
		imagePaths []string
	}{
		name: "portfolio1",
		imagePaths: []string{
			portfolioImage("portfolio1", "images", "bybit.png"),
			portfolioImage("portfolio1", "images", "okx.png"),
		},
	}
	email := fmt.Sprintf("e2e-currency-switch-%s@example.com", runID)
	client, err := loginWithEmail(cfg.baseURL, email)
	if err != nil {
		t.Fatalf("login %s: %v", email, err)
	}
	if err := clearDevEntitlement(client); err != nil {
		t.Fatalf("clear dev entitlement %s: %v", email, err)
	}
	if entitlement, err := fetchEntitlement(client); err == nil {
		auditLogJSON(t, "entitlement_after_clear", entitlement)
	}

	baseA := "USD"
	baseB := "EUR"
	if err := updateBaseCurrency(client, baseA); err != nil {
		t.Fatalf("set base currency %s: %v", baseA, err)
	}
	t.Logf("base_currency set=%s email=%s", baseA, email)

	_, calcID1, err := runPreviewFlow(t, client, cfg, portfolio.name, portfolio.imagePaths, baseA)
	if err != nil {
		t.Fatalf("preview flow %s (%s): %v", portfolio.name, baseA, err)
	}
	previewPayloadA, err := fetchPreviewReport(client, calcID1)
	if err != nil {
		t.Fatalf("fetch preview report (%s): %v", baseA, err)
	}
	auditLogJSON(t, "preview_report_before_switch", previewPayloadA)
	if err := assertReportCurrency(t, previewPayloadA, baseA, "preview-original"); err != nil {
		t.Fatalf("preview report currency (%s): %v", baseA, err)
	}
	netWorthDisplayA, ok := previewPayloadA["net_worth_display"].(float64)
	if !ok {
		t.Fatalf("missing net_worth_display in preview (%s)", baseA)
	}

	if err := updateBaseCurrency(client, baseB); err != nil {
		t.Fatalf("set base currency %s: %v", baseB, err)
	}
	t.Logf("base_currency switched=%s email=%s", baseB, email)

	_, calcID2, err := runPreviewFlow(t, client, cfg, portfolio.name, portfolio.imagePaths, baseB)
	if err != nil {
		t.Fatalf("preview flow %s (%s): %v", portfolio.name, baseB, err)
	}
	t.Logf("preview calc_id new=%s old=%s", calcID2, calcID1)

	previewPayloadA2, err := fetchPreviewReport(client, calcID1)
	if err != nil {
		t.Fatalf("fetch preview report after switch (%s): %v", baseA, err)
	}
	auditLogJSON(t, "preview_report_after_switch", previewPayloadA2)
	if err := assertReportCurrency(t, previewPayloadA2, baseA, "preview-after-switch"); err != nil {
		t.Fatalf("old preview currency after switch: %v", err)
	}
	netWorthDisplayA2, ok := previewPayloadA2["net_worth_display"].(float64)
	if !ok {
		t.Fatalf("missing net_worth_display in preview after switch (%s)", baseA)
	}
	if !floatEquals(netWorthDisplayA, netWorthDisplayA2) {
		t.Fatalf("net_worth_display changed after switch: before=%v after=%v", netWorthDisplayA, netWorthDisplayA2)
	}
}

func loadE2EConfig(t *testing.T) e2eConfig {
	t.Helper()
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("E2E_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	enableEntitlement := parseBoolEnv("E2E_ENABLE_DEV_ENTITLEMENT", true)
	deviceTimezone := strings.TrimSpace(os.Getenv("E2E_DEVICE_TIMEZONE"))
	if deviceTimezone == "" {
		deviceTimezone = "UTC"
	}
	pollInterval := parseDurationEnv("E2E_POLL_INTERVAL_MS", 1500*time.Millisecond)
	stageTimeout := parseDurationEnv("E2E_STAGE_TIMEOUT_MS", 10*time.Minute)

	return e2eConfig{
		baseURL:           baseURL,
		enableEntitlement: enableEntitlement,
		deviceTimezone:    deviceTimezone,
		pollInterval:      pollInterval,
		stageTimeout:      stageTimeout,
	}
}

func parseBoolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return time.Duration(value) * time.Millisecond
}

func e2eRunID() string {
	runIDOnce.Do(func() {
		runID := strings.TrimSpace(os.Getenv("E2E_RUN_ID"))
		if runID == "" {
			runID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		cachedRunID = runID
	})
	return cachedRunID
}

func portfolioImage(portfolio, group, name string) string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(file))
	return filepath.Join(root, "testdata", "portfolios", portfolio, group, name)
}

func runPreviewFlow(t *testing.T, client *apiClient, cfg e2eConfig, name string, imagePaths []string, baseCurrency string) (stageTimings, string, error) {
	t.Helper()
	start := time.Now()

	calcID, ocrDuration, normalizeDuration, err := runUploadAndNormalize(t, client, cfg, imagePaths, baseCurrency)
	if err != nil {
		return stageTimings{}, "", err
	}
	t.Logf("normalize complete calculation_id=%s portfolio=%s", calcID, name)

	var previewDuration time.Duration
	previewStart := time.Now()
	if err := pollPreview(t, client, calcID, cfg.pollInterval, cfg.stageTimeout); err != nil {
		return stageTimings{}, "", err
	}
	previewDuration = time.Since(previewStart)
	previewPayload, err := fetchPreviewReport(client, calcID)
	if err != nil {
		return stageTimings{}, "", err
	}
	auditLogJSON(t, "preview_report", previewPayload)
	if err := assertReportCurrency(t, previewPayload, baseCurrency, "preview"); err != nil {
		return stageTimings{}, "", err
	}

	total := time.Since(start)
	return stageTimings{
		Portfolio: name,
		OCR:       ocrDuration,
		Normalize: normalizeDuration,
		Preview:   previewDuration,
		Total:     total,
	}, calcID, nil
}

func runPaidReport(t *testing.T, client *apiClient, cfg e2eConfig, name, calculationID string, baseCurrency string) (stageTimings, error) {
	t.Helper()
	start := time.Now()
	paidStart := time.Now()
	var paidStartResp map[string]any
	if err := client.postJSON(fmt.Sprintf("/v1/reports/%s/paid", calculationID), nil, &paidStartResp); err != nil {
		return stageTimings{}, err
	}
	auditLogJSON(t, "paid_start_response", paidStartResp)
	if err := pollPaidReport(t, client, calculationID, cfg.pollInterval, cfg.stageTimeout); err != nil {
		return stageTimings{}, err
	}
	paidPayload, err := fetchPaidReport(client, calculationID)
	if err != nil {
		return stageTimings{}, err
	}
	auditLogJSON(t, "paid_report", paidPayload)
	if err := assertReportCurrency(t, paidPayload, baseCurrency, "paid"); err != nil {
		return stageTimings{}, err
	}
	paidDuration := time.Since(paidStart)
	return stageTimings{
		Portfolio: name,
		Paid:      paidDuration,
		Total:     time.Since(start),
	}, nil
}

func runDirectPaidFlow(t *testing.T, client *apiClient, cfg e2eConfig, name string, imagePaths []string, baseCurrency string) (stageTimings, string, error) {
	t.Helper()
	start := time.Now()
	calcID, ocrDuration, normalizeDuration, err := runUploadAndNormalize(t, client, cfg, imagePaths, baseCurrency)
	if err != nil {
		return stageTimings{}, "", err
	}
	t.Logf("normalize complete calculation_id=%s portfolio=%s direct=true", calcID, name)
	paidStart := time.Now()
	var paidStartResp map[string]any
	if err := client.postJSON(fmt.Sprintf("/v1/reports/%s/paid", calcID), nil, &paidStartResp); err != nil {
		return stageTimings{}, "", err
	}
	auditLogJSON(t, "direct_paid_start_response", paidStartResp)
	if err := pollPaidReport(t, client, calcID, cfg.pollInterval, cfg.stageTimeout); err != nil {
		return stageTimings{}, "", err
	}
	paidPayload, err := fetchPaidReport(client, calcID)
	if err != nil {
		return stageTimings{}, "", err
	}
	auditLogJSON(t, "direct_paid_report", paidPayload)
	if err := assertReportCurrency(t, paidPayload, baseCurrency, "direct-paid"); err != nil {
		return stageTimings{}, "", err
	}
	paidDuration := time.Since(paidStart)
	return stageTimings{
		Portfolio: name,
		OCR:       ocrDuration,
		Normalize: normalizeDuration,
		Paid:      paidDuration,
		Total:     time.Since(start),
	}, calcID, nil
}

func runUploadAndNormalize(t *testing.T, client *apiClient, cfg e2eConfig, imagePaths []string, baseCurrency string) (string, time.Duration, time.Duration, error) {
	t.Helper()
	uploadReq, err := buildUploadBatchRequest(imagePaths, cfg.deviceTimezone)
	if err != nil {
		return "", 0, 0, err
	}

	var uploadResp uploadBatchCreateResponse
	if err := client.postJSON("/v1/upload-batches", uploadReq, &uploadResp); err != nil {
		return "", 0, 0, err
	}
	t.Logf("upload batch created id=%s status=%s images=%d", uploadResp.UploadBatchID, uploadResp.Status, len(uploadResp.ImageUploads))
	if len(uploadResp.ImageUploads) != len(imagePaths) {
		return "", 0, 0, fmt.Errorf("upload batch images mismatch: got %d want %d", len(uploadResp.ImageUploads), len(imagePaths))
	}

	imageIDs := make([]string, 0, len(uploadResp.ImageUploads))
	for i, upload := range uploadResp.ImageUploads {
		imageIDs = append(imageIDs, upload.ImageID)
		if err := uploadFile(upload, imagePaths[i]); err != nil {
			return "", 0, 0, err
		}
	}

	ocrStart := time.Now()
	completeReq := uploadBatchCompleteRequest{ImageIDs: imageIDs}
	if err := client.postJSON(fmt.Sprintf("/v1/upload-batches/%s/complete", uploadResp.UploadBatchID), completeReq, nil); err != nil {
		return "", 0, 0, err
	}

	status, payload, err := pollUploadBatch(t, client, uploadResp.UploadBatchID, cfg.pollInterval, cfg.stageTimeout, "needs_review")
	if err != nil {
		return "", 0, 0, err
	}
	if status != "needs_review" {
		return "", 0, 0, fmt.Errorf("unexpected OCR status: %s", status)
	}
	auditLogJSON(t, "upload_needs_review", payload)
	if baseCurrency != "" {
		if err := assertNeedsReviewCurrency(t, payload, baseCurrency); err != nil {
			return "", 0, 0, err
		}
	}
	ocrDuration := time.Since(ocrStart)

	normalizeStart := time.Now()
	reviewReq := uploadBatchReviewRequest{
		PlatformOverrides:  []any{},
		Resolutions:        []any{},
		Edits:              []any{},
		DuplicateOverrides: []any{},
	}
	if err := client.postJSON(fmt.Sprintf("/v1/upload-batches/%s/review", uploadResp.UploadBatchID), reviewReq, nil); err != nil {
		return "", 0, 0, err
	}

	status, payload, err = pollUploadBatch(t, client, uploadResp.UploadBatchID, cfg.pollInterval, cfg.stageTimeout, "completed")
	if err != nil {
		return "", 0, 0, err
	}
	if status != "completed" {
		return "", 0, 0, fmt.Errorf("unexpected normalize status: %s", status)
	}
	auditLogJSON(t, "upload_completed", payload)
	calcID, ok := payload["calculation_id"].(string)
	if !ok || strings.TrimSpace(calcID) == "" {
		return "", 0, 0, errors.New("missing calculation_id in completed response")
	}
	normalizeDuration := time.Since(normalizeStart)
	if baseCurrency != "" {
		if err := assertPortfolioCurrency(t, client, baseCurrency); err != nil {
			return "", 0, 0, err
		}
	}

	return calcID, ocrDuration, normalizeDuration, nil
}

func updateBaseCurrency(client *apiClient, currency string) error {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return nil
	}
	return client.patchJSON("/v1/users/me", map[string]any{
		"base_currency": currency,
	}, nil)
}

func fetchEntitlement(client *apiClient) (map[string]any, error) {
	var payload map[string]any
	if err := client.getJSON("/v1/billing/entitlement", &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func clearDevEntitlement(client *apiClient) error {
	err := client.deleteJSON("/v1/billing/dev/entitlement", nil)
	if err != nil {
		if parseAPIError(err) == "NOT_FOUND" {
			return nil
		}
		return err
	}
	return nil
}

func fetchActivePortfolio(client *apiClient) (map[string]any, error) {
	var payload map[string]any
	if err := client.getJSON("/v1/portfolio/active", &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func fetchPreviewReport(client *apiClient, calculationID string) (map[string]any, error) {
	var payload map[string]any
	if err := client.getJSON(fmt.Sprintf("/v1/reports/preview/%s", calculationID), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func fetchPaidReport(client *apiClient, calculationID string) (map[string]any, error) {
	var payload map[string]any
	if err := client.getJSON(fmt.Sprintf("/v1/reports/%s", calculationID), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func assertNeedsReviewCurrency(t *testing.T, payload map[string]any, expectedCurrency string) error {
	t.Helper()
	got, _ := payload["base_currency"].(string)
	baseFX := payload["base_fx_rate_to_usd"]
	t.Logf("ocr base_currency=%v base_fx_rate_to_usd=%v", got, baseFX)
	if strings.TrimSpace(expectedCurrency) == "" {
		return nil
	}
	if strings.TrimSpace(got) == "" {
		return fmt.Errorf("missing base_currency in OCR response")
	}
	if !strings.EqualFold(got, expectedCurrency) {
		return fmt.Errorf("ocr base_currency mismatch: got %q want %q", got, expectedCurrency)
	}
	return nil
}

func assertPortfolioCurrency(t *testing.T, client *apiClient, expectedCurrency string) error {
	t.Helper()
	payload, err := fetchActivePortfolio(client)
	if err != nil {
		return err
	}
	auditLogJSON(t, "active_portfolio", payload)
	metrics, ok := payload["dashboard_metrics"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing dashboard_metrics in portfolio response")
	}
	got, _ := metrics["base_currency"].(string)
	t.Logf("portfolio base_currency=%v net_worth_usd=%v net_worth_display=%v base_fx_rate_to_usd=%v", got, metrics["net_worth_usd"], metrics["net_worth_display"], metrics["base_fx_rate_to_usd"])
	if strings.TrimSpace(expectedCurrency) == "" {
		return nil
	}
	if strings.TrimSpace(got) == "" {
		return fmt.Errorf("missing base_currency in portfolio metrics")
	}
	if !strings.EqualFold(got, expectedCurrency) {
		return fmt.Errorf("portfolio base_currency mismatch: got %q want %q", got, expectedCurrency)
	}
	return nil
}

func assertReportCurrency(t *testing.T, payload map[string]any, expectedCurrency, label string) error {
	t.Helper()
	got, _ := payload["base_currency"].(string)
	t.Logf("%s report base_currency=%v net_worth_display=%v base_fx_rate_to_usd=%v", label, got, payload["net_worth_display"], payload["base_fx_rate_to_usd"])
	if strings.TrimSpace(expectedCurrency) == "" {
		return nil
	}
	if strings.TrimSpace(got) == "" {
		return fmt.Errorf("missing base_currency in %s report", label)
	}
	if !strings.EqualFold(got, expectedCurrency) {
		return fmt.Errorf("%s report base_currency mismatch: got %q want %q", label, got, expectedCurrency)
	}
	return nil
}

func requirePaidEntitlement(client *apiClient, calculationID string) error {
	err := client.postJSON(fmt.Sprintf("/v1/reports/%s/paid", calculationID), nil, nil)
	if err == nil {
		return fmt.Errorf("expected entitlement error, got nil")
	}
	code := parseAPIError(err)
	if code != "ENTITLEMENT_REQUIRED" {
		return fmt.Errorf("expected ENTITLEMENT_REQUIRED, got %q", code)
	}
	return nil
}

func buildUploadBatchRequest(paths []string, timezone string) (*uploadBatchCreateRequest, error) {
	images := make([]uploadBatchImageIn, 0, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat image %s: %w", path, err)
		}
		mime := mimeTypeForFile(path)
		images = append(images, uploadBatchImageIn{
			FileName:  filepath.Base(path),
			MimeType:  mime,
			SizeBytes: info.Size(),
		})
	}
	return &uploadBatchCreateRequest{
		Purpose:        "holdings",
		ImageCount:     len(images),
		Images:         images,
		DeviceTimezone: timezone,
	}, nil
}

func uploadFile(upload uploadBatchUpload, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read image %s: %w", path, err)
	}
	req, err := http.NewRequest(http.MethodPut, upload.UploadURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build upload request: %w", err)
	}
	hasContentType := false
	for key, value := range upload.Headers {
		req.Header.Set(key, value)
		if strings.EqualFold(key, "Content-Type") {
			hasContentType = true
		}
	}
	if !hasContentType {
		req.Header.Set("Content-Type", mimeTypeForFile(path))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func pollUploadBatch(t *testing.T, client *apiClient, batchID string, interval, timeout time.Duration, expect string) (string, map[string]any, error) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	attempt := 0
	for {
		attempt++
		if time.Now().After(deadline) {
			return "", nil, fmt.Errorf("timeout waiting for upload batch %s", expect)
		}
		var payload map[string]any
		if err := client.getJSON(fmt.Sprintf("/v1/upload-batches/%s", batchID), &payload); err != nil {
			return "", nil, err
		}
		status, _ := payload["status"].(string)
		if status == expect {
			return status, payload, nil
		}
		if status == "failed" {
			auditLogJSON(t, fmt.Sprintf("upload_batch_failed_%s", batchID), payload)
			return status, payload, fmt.Errorf("upload batch failed: %v", payload["error_code"])
		}
		if attempt%10 == 0 {
			t.Logf("upload batch poll id=%s status=%s expect=%s attempt=%d", batchID, status, expect, attempt)
		}
		time.Sleep(interval)
	}
}

func pollPreview(t *testing.T, client *apiClient, calculationID string, interval, timeout time.Duration) error {
	t.Helper()
	deadline := time.Now().Add(timeout)
	attempt := 0
	for {
		attempt++
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for preview report")
		}
		var payload map[string]any
		err := client.getJSON(fmt.Sprintf("/v1/reports/preview/%s", calculationID), &payload)
		if err == nil {
			if _, ok := payload["meta_data"]; ok {
				return nil
			}
			if code := errorCodeFromPayload(payload); code == "NOT_READY" {
				if attempt%20 == 0 {
					t.Logf("preview poll not ready calculation_id=%s attempt=%d", calculationID, attempt)
				}
				time.Sleep(interval)
				continue
			}
		}
		if apiErr := parseAPIError(err); apiErr == "NOT_READY" {
			if attempt%20 == 0 {
				t.Logf("preview poll not ready calculation_id=%s attempt=%d", calculationID, attempt)
			}
			time.Sleep(interval)
			continue
		}
		if err != nil {
			return err
		}
		if attempt%20 == 0 {
			t.Logf("preview poll pending calculation_id=%s attempt=%d", calculationID, attempt)
		}
		time.Sleep(interval)
	}
}

func pollPaidReport(t *testing.T, client *apiClient, calculationID string, interval, timeout time.Duration) error {
	t.Helper()
	deadline := time.Now().Add(timeout)
	attempt := 0
	for {
		attempt++
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for paid report")
		}
		var payload map[string]any
		err := client.getJSON(fmt.Sprintf("/v1/reports/%s", calculationID), &payload)
		if err == nil {
			if isPaidPayload(payload) {
				return nil
			}
			if code := errorCodeFromPayload(payload); code == "NOT_READY" {
				if attempt%20 == 0 {
					t.Logf("paid poll not ready calculation_id=%s attempt=%d", calculationID, attempt)
				}
				time.Sleep(interval)
				continue
			}
			time.Sleep(interval)
			continue
		}
		if apiErr := parseAPIError(err); apiErr == "NOT_READY" {
			if attempt%20 == 0 {
				t.Logf("paid poll not ready calculation_id=%s attempt=%d", calculationID, attempt)
			}
			time.Sleep(interval)
			continue
		}
		return err
	}
}

func isPaidPayload(payload map[string]any) bool {
	if payload == nil {
		return false
	}
	if _, ok := payload["report_header"]; ok {
		if _, ok := payload["charts"]; ok {
			return true
		}
	}
	if _, ok := payload["risk_insights"]; ok {
		return true
	}
	return false
}

func parseAPIError(err error) string {
	var apiErr apiError
	if err != nil && errors.As(err, &apiErr) {
		return apiErr.code
	}
	return ""
}

type apiError struct {
	status int
	code   string
	body   string
}

func (e apiError) Error() string {
	if e.code != "" {
		return fmt.Sprintf("api error status=%d code=%s body=%s", e.status, e.code, e.body)
	}
	return fmt.Sprintf("api error status=%d body=%s", e.status, e.body)
}

func newAPIClient(baseURL, token string) *apiClient {
	return &apiClient{
		baseURL:      baseURL,
		token:        token,
		refreshToken: "",
		client:       &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *apiClient) getJSON(path string, out any) error {
	return c.requestJSON(http.MethodGet, path, nil, out)
}

func (c *apiClient) postJSON(path string, body any, out any) error {
	return c.requestJSON(http.MethodPost, path, body, out)
}

func (c *apiClient) patchJSON(path string, body any, out any) error {
	return c.requestJSON(http.MethodPatch, path, body, out)
}

func (c *apiClient) deleteJSON(path string, out any) error {
	return c.requestJSON(http.MethodDelete, path, nil, out)
}

func (c *apiClient) requestJSON(method, path string, body any, out any) error {
	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		if err := c.doRequest(method, path, body, out); err != nil {
			lastErr = err
			var apiErr apiError
			if errors.As(err, &apiErr) {
				if apiErr.status == http.StatusUnauthorized && strings.TrimSpace(c.refreshToken) != "" {
					if err := c.refreshAccessToken(); err != nil {
						return err
					}
					continue
				}
				if apiErr.code == "RATE_LIMITED" || apiErr.status == http.StatusTooManyRequests {
					delay := rateLimitBackoff(attempt)
					fmt.Printf("rate limited method=%s path=%s attempt=%d wait=%s\n", method, path, attempt+1, delay)
					time.Sleep(delay)
					continue
				}
			}
			return err
		}
		return nil
	}
	return lastErr
}

func (c *apiClient) doRequest(method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	c.applyHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req, out)
}

func (c *apiClient) refreshAccessToken() error {
	refresh := strings.TrimSpace(c.refreshToken)
	if refresh == "" {
		return errors.New("refresh token missing")
	}
	authClient := newAPIClient(c.baseURL, "")
	var resp struct {
		AccessToken string `json:"access_token"`
	}
	if err := authClient.postJSON("/v1/auth/refresh", map[string]any{"refresh_token": refresh}, &resp); err != nil {
		return err
	}
	token := strings.TrimSpace(resp.AccessToken)
	if token == "" {
		return errors.New("access token missing")
	}
	c.token = token
	return nil
}

func (c *apiClient) applyHeaders(req *http.Request) {
	if strings.TrimSpace(c.token) != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")
}

func (c *apiClient) do(req *http.Request, out any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiError{
			status: resp.StatusCode,
			code:   extractErrorCode(body),
			body:   strings.TrimSpace(string(body)),
		}
	}
	if out == nil {
		return nil
	}
	if len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w body=%s", err, strings.TrimSpace(string(body)))
	}
	return nil
}

func extractErrorCode(body []byte) string {
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return payload.Error.Code
}

func errorCodeFromPayload(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	raw, ok := payload["error"].(map[string]any)
	if !ok {
		return ""
	}
	code, _ := raw["code"].(string)
	return code
}

func loginWithEmail(baseURL, email string) (*apiClient, error) {
	password := "MoneyCoach-E2E-Password-01!"
	authClient := newAPIClient(baseURL, "")
	var authResp authResponse

	var startResp authEmailRegisterStartResponse
	err := authClient.postJSON("/v1/auth/email/register/start", authEmailRegisterStartRequest{Email: email}, &startResp)
	needsLogin := false
	if err != nil {
		var apiErr apiError
		if !errors.As(err, &apiErr) {
			return nil, err
		}
		if apiErr.status == http.StatusConflict && apiErr.code == "EMAIL_ALREADY_EXISTS" {
			needsLogin = true
		} else {
			return nil, err
		}
	}
	if !needsLogin {
		code := strings.TrimSpace(startResp.Code)
		if code == "" {
			return nil, errors.New("otp code missing (ensure EMAIL_OTP_MODE=debug or local)")
		}
		err = authClient.postJSON(
			"/v1/auth/email/register",
			authEmailRegisterRequest{Email: email, Password: password, Code: code},
			&authResp,
		)
		if err != nil {
			var apiErr apiError
			if !errors.As(err, &apiErr) {
				return nil, err
			}
			if apiErr.status == http.StatusConflict && apiErr.code == "EMAIL_ALREADY_EXISTS" {
				needsLogin = true
			} else {
				return nil, err
			}
		}
	}
	if needsLogin {
		if err := authClient.postJSON("/v1/auth/email/login", authEmailLoginRequest{Email: email, Password: password}, &authResp); err != nil {
			return nil, err
		}
	}

	token := strings.TrimSpace(authResp.AccessToken)
	if token == "" {
		return nil, errors.New("access token missing")
	}
	refresh := strings.TrimSpace(authResp.RefreshToken)
	if refresh == "" {
		return nil, errors.New("refresh token missing")
	}
	client := newAPIClient(baseURL, token)
	client.refreshToken = refresh
	return client, nil
}

func mimeTypeForFile(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func floatEquals(a, b float64) bool {
	const epsilon = 0.0001
	if a > b {
		return a-b < epsilon
	}
	return b-a < epsilon
}

func rateLimitBackoff(attempt int) time.Duration {
	delay := 5 * time.Second * time.Duration(1<<attempt)
	if delay > 70*time.Second {
		return 70 * time.Second
	}
	return delay
}

func auditLogJSON(t *testing.T, label string, payload any) {
	t.Helper()
	if payload == nil {
		t.Logf("audit %s: <nil>", label)
		return
	}
	encoded, err := marshalAuditJSON(payload)
	if err != nil {
		t.Logf("audit %s marshal error: %v", label, err)
		return
	}
	t.Logf("audit %s:\n%s", label, string(encoded))
	if err := writeAuditFile(t, label, encoded); err != nil {
		t.Logf("audit %s write error: %v", label, err)
	}
}

func marshalAuditJSON(payload any) ([]byte, error) {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}
	return asciiJSON(raw), nil
}

func asciiJSON(raw []byte) []byte {
	var builder strings.Builder
	builder.Grow(len(raw))
	for _, r := range string(raw) {
		if r <= 0x7f {
			builder.WriteRune(r)
			continue
		}
		for _, u := range utf16.Encode([]rune{r}) {
			fmt.Fprintf(&builder, "\\u%04x", u)
		}
	}
	return []byte(builder.String())
}

func writeAuditFile(t *testing.T, label string, content []byte) error {
	t.Helper()
	dir := "test-logs"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s-%s-%s.json", sanitizeFileComponent(t.Name()), sanitizeFileComponent(label), sanitizeFileComponent(e2eRunID()))
	path := filepath.Join(dir, fileName)
	if len(content) == 0 || content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	return os.WriteFile(path, content, 0o644)
}

func sanitizeFileComponent(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	builder.Grow(len(value))
	lastUnderscore := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	output := strings.Trim(builder.String(), "_")
	if output == "" {
		return "unknown"
	}
	return output
}
