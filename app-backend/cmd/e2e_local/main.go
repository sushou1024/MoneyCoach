package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type authErrorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

type authEmailRegisterStartResponse struct {
	Sent bool   `json:"sent"`
	Code string `json:"code"`
}

type uploadBatchCreateRequest struct {
	Purpose        string             `json:"purpose"`
	ImageCount     int                `json:"image_count"`
	Images         []uploadBatchImage `json:"images"`
	DeviceTimezone string             `json:"device_timezone,omitempty"`
}

type uploadBatchImage struct {
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

type imageFile struct {
	Path     string
	FileName string
	MimeType string
	Size     int64
}

func main() {
	var (
		baseURL   = flag.String("base-url", "http://localhost:8080", "Backend base URL")
		email     = flag.String("email", "jackcpku@gmail.com", "Email used for registration")
		password  = flag.String("password", "MoneyCoach-E2E-Password-01!", "Email account password")
		portfolio = flag.String("portfolio", "portfolio1", "Testdata portfolio name")
		timeout   = flag.Duration("timeout", 15*time.Minute, "Overall timeout for OCR/report readiness")
	)
	flag.Parse()

	base := strings.TrimRight(strings.TrimSpace(*baseURL), "/")
	if base == "" {
		fail(errors.New("base URL is required"))
	}

	ctx := context.Background()
	client := &http.Client{Timeout: 60 * time.Second}

	logf("Starting email auth for %s...", *email)
	status, body, err := requestJSON(ctx, client, http.MethodPost, base+"/v1/auth/email/register/start", "", map[string]string{
		"email": *email,
	}, http.StatusConflict)
	if err != nil {
		fail(err)
	}

	shouldLogin := false
	if status == http.StatusConflict {
		var payload authErrorEnvelope
		if err := json.Unmarshal(body, &payload); err != nil {
			fail(fmt.Errorf("decode register conflict payload: %w", err))
		}
		if strings.TrimSpace(payload.Error.Code) != "EMAIL_ALREADY_EXISTS" {
			fail(fmt.Errorf("unexpected register-start conflict: %s", strings.TrimSpace(string(body))))
		}
		shouldLogin = true
	} else {
		var startResp authEmailRegisterStartResponse
		if err := json.Unmarshal(body, &startResp); err != nil {
			fail(fmt.Errorf("decode register-start response: %w", err))
		}
		code := strings.TrimSpace(startResp.Code)
		if code == "" {
			code = prompt("Enter verification code sent to " + *email + ": ")
		}
		if code == "" {
			fail(errors.New("verification code is required"))
		}

		status, body, err = requestJSON(ctx, client, http.MethodPost, base+"/v1/auth/email/register", "", map[string]string{
			"email":    *email,
			"password": *password,
			"code":     code,
		}, http.StatusConflict)
		if err != nil {
			fail(err)
		}
		if status == http.StatusConflict {
			var payload authErrorEnvelope
			if err := json.Unmarshal(body, &payload); err != nil {
				fail(fmt.Errorf("decode register conflict payload: %w", err))
			}
			if strings.TrimSpace(payload.Error.Code) != "EMAIL_ALREADY_EXISTS" {
				fail(fmt.Errorf("unexpected register conflict: %s", strings.TrimSpace(string(body))))
			}
			shouldLogin = true
		}
	}

	if shouldLogin {
		status, body, err = requestJSON(ctx, client, http.MethodPost, base+"/v1/auth/email/login", "", map[string]string{
			"email":    *email,
			"password": *password,
		})
		if err != nil {
			fail(err)
		}
		if status != http.StatusOK {
			fail(fmt.Errorf("unexpected login status: %d", status))
		}
	}

	var auth authResponse
	if err := json.Unmarshal(body, &auth); err != nil {
		fail(err)
	}

	if auth.AccessToken == "" {
		fail(errors.New("missing access token after authentication"))
	}

	logf("Authenticated user_id=%s", auth.UserID)

	if err := setEnglishProfile(ctx, client, base, auth.AccessToken); err != nil {
		fail(err)
	}

	images, err := loadTestImages(*portfolio)
	if err != nil {
		fail(err)
	}

	createReq := uploadBatchCreateRequest{
		Purpose:    "holdings",
		ImageCount: len(images),
		Images:     make([]uploadBatchImage, 0, len(images)),
	}
	for _, img := range images {
		createReq.Images = append(createReq.Images, uploadBatchImage{
			FileName:  img.FileName,
			MimeType:  img.MimeType,
			SizeBytes: img.Size,
		})
	}

	var createResp uploadBatchCreateResponse
	if _, body, err := requestJSON(ctx, client, http.MethodPost, base+"/v1/upload-batches", auth.AccessToken, createReq); err != nil {
		fail(err)
	} else if err := json.Unmarshal(body, &createResp); err != nil {
		fail(err)
	}

	if createResp.UploadBatchID == "" || len(createResp.ImageUploads) != len(images) {
		fail(errors.New("unexpected upload batch response"))
	}

	logf("Upload batch created: id=%s status=%s expires_at=%s", createResp.UploadBatchID, createResp.Status, createResp.ExpiresAt)

	logf("Uploading %d images...", len(images))
	for i, upload := range createResp.ImageUploads {
		logf("Uploading image_id=%s file=%s", upload.ImageID, images[i].FileName)
		if err := uploadImage(ctx, client, upload.UploadURL, images[i], upload.Headers); err != nil {
			fail(err)
		}
	}

	imageIDs := make([]string, 0, len(createResp.ImageUploads))
	for _, upload := range createResp.ImageUploads {
		imageIDs = append(imageIDs, upload.ImageID)
	}

	if _, _, err := requestJSON(ctx, client, http.MethodPost, base+"/v1/upload-batches/"+createResp.UploadBatchID+"/complete", auth.AccessToken, map[string]any{
		"image_ids": imageIDs,
	}); err != nil {
		fail(err)
	}

	logf("Upload batch marked complete.")
	logf("Waiting for OCR + preview generation...")
	calculationID, err := waitForCalculation(ctx, client, base, auth.AccessToken, createResp.UploadBatchID, *timeout)
	if err != nil {
		fail(err)
	}

	if err := waitForPreview(ctx, client, base, auth.AccessToken, calculationID, *timeout); err != nil {
		fail(err)
	}

	logf("E2E flow complete. calculation_id=%s", calculationID)
}

func prompt(msg string) string {
	fmt.Printf("%s %s", time.Now().Format(time.RFC3339), msg)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func setEnglishProfile(ctx context.Context, client *http.Client, baseURL, token string) error {
	_, _, err := requestJSON(ctx, client, http.MethodPatch, baseURL+"/v1/users/me", token, map[string]string{
		"language": "en",
	})
	if err == nil {
		logf("Profile updated: language=en")
	}
	return err
}

func loadTestImages(portfolio string) ([]imageFile, error) {
	if strings.TrimSpace(portfolio) == "" {
		return nil, errors.New("portfolio is required")
	}
	glob := filepath.Join("testdata", "portfolios", portfolio, "images", "*")
	paths, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("glob testdata images: %w", err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no images found for %s", portfolio)
	}
	sort.Strings(paths)

	images := make([]imageFile, 0, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}
		if info.IsDir() {
			continue
		}
		images = append(images, imageFile{
			Path:     path,
			FileName: filepath.Base(path),
			MimeType: mimeTypeForPath(path),
			Size:     info.Size(),
		})
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no images found for %s", portfolio)
	}
	logf("Loaded %d images from testdata/%s", len(images), portfolio)
	return images, nil
}

func mimeTypeForPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return "image/png"
	}
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func uploadImage(ctx context.Context, client *http.Client, uploadURL string, image imageFile, headers map[string]string) error {
	file, err := os.Open(image.Path)
	if err != nil {
		return fmt.Errorf("open %s: %w", image.Path, err)
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, file)
	if err != nil {
		return err
	}
	req.ContentLength = image.Size
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", image.MimeType)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	logf("Upload PUT %s -> %d", uploadURL, resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func waitForCalculation(ctx context.Context, client *http.Client, baseURL, token, batchID string, timeout time.Duration) (string, error) {
	start := time.Now()
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		status, body, err := requestJSON(ctx, client, http.MethodGet, baseURL+"/v1/upload-batches/"+batchID, token, nil, http.StatusTooManyRequests)
		if err != nil {
			return "", err
		}
		if status == http.StatusTooManyRequests {
			logf("Rate limited while polling upload batch (attempt %d, elapsed %s). Backing off...", attempt, time.Since(start).Truncate(time.Second))
			time.Sleep(6 * time.Second)
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", err
		}
		statusValue, _ := payload["status"].(string)
		logf("Upload batch status=%s (attempt %d, elapsed %s)", statusValue, attempt, time.Since(start).Truncate(time.Second))
		switch statusValue {
		case "processing":
			time.Sleep(2 * time.Second)
		case "needs_review":
			logf("Upload batch needs review; submitting empty review payload.")
			if err := submitReview(ctx, client, baseURL, token, batchID); err != nil {
				return "", err
			}
			time.Sleep(2 * time.Second)
		case "completed":
			if calcID, ok := payload["calculation_id"].(string); ok && calcID != "" {
				logf("Upload batch completed. calculation_id=%s", calcID)
				return calcID, nil
			}
			return "", errors.New("missing calculation_id in completed response")
		case "failed":
			return "", fmt.Errorf("upload batch failed: %v", payload["error_code"])
		default:
			time.Sleep(2 * time.Second)
		}
	}
	return "", errors.New("timed out waiting for upload batch completion")
}

func submitReview(ctx context.Context, client *http.Client, baseURL, token, batchID string) error {
	_, _, err := requestJSON(ctx, client, http.MethodPost, baseURL+"/v1/upload-batches/"+batchID+"/review", token, map[string]any{
		"platform_overrides":  []any{},
		"resolutions":         []any{},
		"edits":               []any{},
		"duplicate_overrides": []any{},
	})
	return err
}

func waitForPreview(ctx context.Context, client *http.Client, baseURL, token, calcID string, timeout time.Duration) error {
	start := time.Now()
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		status, body, err := requestJSON(ctx, client, http.MethodGet, baseURL+"/v1/reports/preview/"+calcID, token, nil, http.StatusTooManyRequests)
		if err != nil {
			return err
		}
		if status == http.StatusTooManyRequests {
			logf("Rate limited while polling preview (attempt %d, elapsed %s). Backing off...", attempt, time.Since(start).Truncate(time.Second))
			time.Sleep(6 * time.Second)
			continue
		}
		if status == http.StatusAccepted {
			logf("Preview not ready yet (attempt %d, elapsed %s)", attempt, time.Since(start).Truncate(time.Second))
			time.Sleep(2 * time.Second)
			continue
		}
		if status == http.StatusOK {
			var preview map[string]any
			if err := json.Unmarshal(body, &preview); err != nil {
				return err
			}
			if fixed, ok := preview["fixed_metrics"].(map[string]any); ok {
				logf("Preview health_score=%v health_status=%v", fixed["health_score"], fixed["health_status"])
			}
			return nil
		}
		return fmt.Errorf("unexpected preview status %d", status)
	}
	return errors.New("timed out waiting for preview report")
}

func requestJSON(ctx context.Context, client *http.Client, method, url, token string, payload any, allowedStatuses ...int) (int, []byte, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	logf("HTTP %s %s -> %d (%s)", method, url, resp.StatusCode, time.Since(start).Truncate(time.Millisecond))
	if isAllowedStatus(resp.StatusCode, allowedStatuses...) {
		return resp.StatusCode, raw, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, raw, fmt.Errorf("request failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return resp.StatusCode, raw, nil
}

func isAllowedStatus(status int, allowedStatuses ...int) bool {
	if status >= 200 && status < 300 {
		return true
	}
	for _, allowed := range allowedStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

func logf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	fmt.Printf("%s %s", time.Now().Format(time.RFC3339), message)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, time.Now().Format(time.RFC3339), "Error:", err)
	os.Exit(1)
}
