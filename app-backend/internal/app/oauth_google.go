package app

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/idtoken"
)

var errGoogleAuthNotConfigured = errors.New("google oauth not configured")

type googleIDTokenClaims struct {
	Subject       string
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (s *Server) verifyGoogleIDToken(ctx context.Context, token string) (googleIDTokenClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return googleIDTokenClaims{}, errors.New("id_token required")
	}

	audiences := s.googleAllowedAudiences()
	if len(audiences) == 0 {
		return googleIDTokenClaims{}, errGoogleAuthNotConfigured
	}

	var lastErr error
	for _, audience := range audiences {
		payload, err := idtoken.Validate(ctx, token, audience)
		if err != nil {
			lastErr = err
			continue
		}

		claims, err := readGoogleClaims(payload.Claims)
		if err != nil {
			return googleIDTokenClaims{}, err
		}
		claims.Subject = payload.Subject
		if claims.Subject == "" {
			return googleIDTokenClaims{}, errors.New("missing subject")
		}
		return claims, nil
	}

	if lastErr != nil {
		return googleIDTokenClaims{}, lastErr
	}
	return googleIDTokenClaims{}, errors.New("no valid google audiences configured")
}

func (s *Server) googleAllowedAudiences() []string {
	if len(s.cfg.GoogleAllowedClient) == 0 {
		return nil
	}
	out := make([]string, 0, len(s.cfg.GoogleAllowedClient))
	for _, entry := range s.cfg.GoogleAllowedClient {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func readGoogleClaims(raw map[string]any) (googleIDTokenClaims, error) {
	var claims googleIDTokenClaims
	if raw == nil {
		return claims, errors.New("missing claims")
	}
	if value, ok := raw["email"].(string); ok {
		claims.Email = value
	}
	switch value := raw["email_verified"].(type) {
	case bool:
		claims.EmailVerified = value
	case string:
		claims.EmailVerified = strings.EqualFold(value, "true")
	case float64:
		claims.EmailVerified = value != 0
	}
	return claims, nil
}
