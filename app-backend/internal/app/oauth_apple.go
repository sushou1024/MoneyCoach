package app

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	appleIssuer         = "https://appleid.apple.com"
	appleIDKeysURL      = "https://appleid.apple.com/auth/keys"
	appleStoreKeysURL   = "https://api.storekit.itunes.apple.com/in-app/v1/jwsPublicKeys"
	appleKeyCachePeriod = 24 * time.Hour
)

var errAppleAuthNotConfigured = errors.New("apple oauth not configured")

type appleIDTokenClaims struct {
	Subject       string
	Email         string
	EmailVerified bool
}

type appleJWKSet struct {
	Keys []appleJWK `json:"keys"`
}

type appleJWK struct {
	KeyID string `json:"kid"`
	Kty   string `json:"kty"`
	Alg   string `json:"alg"`
	Use   string `json:"use"`
	N     string `json:"n"`
	E     string `json:"e"`
	Crv   string `json:"crv"`
	X     string `json:"x"`
	Y     string `json:"y"`
}

type appleKeyCache struct {
	mu        sync.Mutex
	fetchedAt time.Time
	keys      map[string]crypto.PublicKey
}

func (s *Server) verifyAppleIDToken(ctx context.Context, token string) (appleIDTokenClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return appleIDTokenClaims{}, errors.New("id_token required")
	}
	audiences := s.appleAllowedAudiences()
	if len(audiences) == 0 {
		return appleIDTokenClaims{}, errAppleAuthNotConfigured
	}

	claims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	parsed, err := parser.ParseWithClaims(token, claims, func(jwtToken *jwt.Token) (any, error) {
		kid, _ := jwtToken.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing key id")
		}
		return s.applePublicKey(ctx, appleIDKeysURL, kid)
	})
	if err != nil || parsed == nil || !parsed.Valid {
		return appleIDTokenClaims{}, errors.New("invalid apple token")
	}

	issuer, _ := claims["iss"].(string)
	if issuer != appleIssuer {
		return appleIDTokenClaims{}, errors.New("invalid issuer")
	}
	if !audienceMatches(claims["aud"], audiences) {
		return appleIDTokenClaims{}, errors.New("invalid audience")
	}

	subject, _ := claims["sub"].(string)
	if subject == "" {
		return appleIDTokenClaims{}, errors.New("missing subject")
	}
	email, _ := claims["email"].(string)
	emailVerified := parseBoolClaim(claims["email_verified"])

	return appleIDTokenClaims{
		Subject:       subject,
		Email:         email,
		EmailVerified: emailVerified,
	}, nil
}

func (s *Server) appleAllowedAudiences() []string {
	if len(s.cfg.AppleAllowedClient) == 0 {
		return nil
	}
	out := make([]string, 0, len(s.cfg.AppleAllowedClient))
	for _, entry := range s.cfg.AppleAllowedClient {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func (s *Server) applePublicKey(ctx context.Context, url string, keyID string) (crypto.PublicKey, error) {
	cache := s.appleKeyCache(url)
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.keys != nil && time.Since(cache.fetchedAt) < appleKeyCachePeriod {
		if key, ok := cache.keys[keyID]; ok {
			return key, nil
		}
	}

	keys, err := s.fetchAppleKeys(ctx, url)
	if err != nil {
		return nil, err
	}
	cache.keys = keys
	cache.fetchedAt = time.Now().UTC()
	if key, ok := cache.keys[keyID]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("apple key not found: %s", keyID)
}

func (s *Server) appleKeyCache(url string) *appleKeyCache {
	switch url {
	case appleStoreKeysURL:
		return &s.appleStoreKeys
	default:
		return &s.appleIDKeys
	}
}

func (s *Server) fetchAppleKeys(ctx context.Context, url string) (map[string]crypto.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("apple keys fetch failed status=%d", resp.StatusCode)
	}
	var set appleJWKSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, err
	}
	keys := make(map[string]crypto.PublicKey, len(set.Keys))
	for _, key := range set.Keys {
		parsed, err := parseAppleJWK(key)
		if err != nil {
			continue
		}
		keys[key.KeyID] = parsed
	}
	if len(keys) == 0 {
		return nil, errors.New("no apple keys found")
	}
	return keys, nil
}

func parseAppleJWK(jwk appleJWK) (crypto.PublicKey, error) {
	switch strings.ToUpper(jwk.Kty) {
	case "RSA":
		n, err := decodeBase64URLBigInt(jwk.N)
		if err != nil {
			return nil, err
		}
		e, err := decodeBase64URLBigInt(jwk.E)
		if err != nil {
			return nil, err
		}
		return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
	case "EC":
		curve := elliptic.P256()
		if strings.EqualFold(jwk.Crv, "P-256") {
			curve = elliptic.P256()
		}
		x, err := decodeBase64URLBigInt(jwk.X)
		if err != nil {
			return nil, err
		}
		y, err := decodeBase64URLBigInt(jwk.Y)
		if err != nil {
			return nil, err
		}
		return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}
}

func decodeBase64URLBigInt(value string) (*big.Int, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(decoded), nil
}

func parseBoolClaim(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	case float64:
		return v != 0
	default:
		return false
	}
}

func audienceMatches(raw any, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	values := make([]string, 0)
	switch aud := raw.(type) {
	case string:
		values = append(values, aud)
	case []any:
		for _, entry := range aud {
			if str, ok := entry.(string); ok {
				values = append(values, str)
			}
		}
	case []string:
		values = append(values, aud...)
	}
	if len(values) == 0 {
		return false
	}
	for _, candidate := range values {
		for _, allowedValue := range allowed {
			if candidate == allowedValue {
				return true
			}
		}
	}
	return false
}
