package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type resendClient struct {
	apiKey string
	from   string
	client *http.Client
}

func newResendClient(apiKey, from string, client *http.Client) *resendClient {
	return &resendClient{apiKey: apiKey, from: from, client: client}
}

func (r *resendClient) sendVerificationCode(ctx context.Context, email, code string) error {
	payload := map[string]any{
		"from":    r.from,
		"to":      []string{email},
		"subject": "Your Money Coach verification code",
		"text":    fmt.Sprintf("Your verification code is %s. It expires in 10 minutes.", code),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend status %d", resp.StatusCode)
	}
	return nil
}
