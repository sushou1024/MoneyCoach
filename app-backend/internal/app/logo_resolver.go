package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type logoResolver struct {
	hkBaseURL     string
	hkMappingURL  string
	nasdaqBaseURL string
	nyseBaseURL   string
	client        *http.Client
	logger        *log.Logger

	mu          sync.RWMutex
	symbolToURL map[string]string
	lastFetched time.Time
	lastAttempt time.Time
	lastErr     error
	fetchTTL    time.Duration
	retryTTL    time.Duration
}

type xhkgLogoMapping struct {
	SymbolToFile map[string]string `json:"symbol_to_file"`
}

func newLogoResolver(cfg Config, client *http.Client, logger *log.Logger) *logoResolver {
	hkBaseURL := normalizeLogoBaseURL(cfg.LogosXHKGBaseURL)
	mappingURL := ""
	if hkBaseURL != "" {
		mappingURL = hkBaseURL + "logo-mapping.json"
	}
	nasdaqBaseURL := normalizeLogoBaseURL(cfg.LogosNasdaqBaseURL)
	nyseBaseURL := normalizeLogoBaseURL(cfg.LogosNyseBaseURL)
	if logger == nil {
		logger = log.Default()
	}
	return &logoResolver{
		hkBaseURL:     hkBaseURL,
		hkMappingURL:  mappingURL,
		nasdaqBaseURL: nasdaqBaseURL,
		nyseBaseURL:   nyseBaseURL,
		client:        client,
		logger:        logger,
		symbolToURL:   map[string]string{},
		fetchTTL:      12 * time.Hour,
		retryTTL:      5 * time.Minute,
	}
}

func (r *logoResolver) hkLogoURL(ctx context.Context, displaySymbol string) string {
	if r == nil {
		return ""
	}
	hkexSymbol, ok := hkexSymbolFromDisplay(displaySymbol)
	if !ok {
		return ""
	}
	if err := r.ensureMapping(ctx); err != nil {
		return ""
	}
	r.mu.RLock()
	file := r.symbolToURL[hkexSymbol]
	r.mu.RUnlock()
	if file == "" {
		return ""
	}
	if r.hkBaseURL == "" {
		return ""
	}
	return r.hkBaseURL + file
}

func (r *logoResolver) stockLogoURL(exchangeMIC, symbol string) string {
	if r == nil {
		return ""
	}
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return ""
	}
	exchange := strings.ToUpper(strings.TrimSpace(exchangeMIC))
	var baseURL string
	var prefix string
	switch exchange {
	case "XNAS":
		baseURL = r.nasdaqBaseURL
		prefix = "NASDAQ-"
	case "XNYS":
		baseURL = r.nyseBaseURL
		prefix = "NYSE-"
	default:
		return ""
	}
	if baseURL == "" {
		return ""
	}
	upperSymbol := strings.ToUpper(symbol)
	upperSymbol = strings.TrimSuffix(upperSymbol, ".SVG")
	if prefix != "" && !strings.HasPrefix(upperSymbol, prefix) {
		upperSymbol = prefix + upperSymbol
	}
	return baseURL + upperSymbol + ".svg"
}

func hkexSymbolFromDisplay(symbol string) (string, bool) {
	normalized, ok := normalizeHKSymbol(symbol)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSuffix(strings.ToUpper(normalized), ".HK")
	if trimmed == "" {
		return "", false
	}
	return "HKEX-" + trimmed, true
}

func (r *logoResolver) ensureMapping(ctx context.Context) error {
	if r.hkMappingURL == "" {
		return errors.New("logos mapping url not configured")
	}
	now := time.Now()

	r.mu.RLock()
	if len(r.symbolToURL) > 0 && now.Sub(r.lastFetched) < r.fetchTTL {
		r.mu.RUnlock()
		return nil
	}
	if now.Sub(r.lastAttempt) < r.retryTTL {
		err := r.lastErr
		r.mu.RUnlock()
		if err == nil {
			err = errors.New("logos mapping not ready")
		}
		return err
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.symbolToURL) > 0 && now.Sub(r.lastFetched) < r.fetchTTL {
		return nil
	}
	if now.Sub(r.lastAttempt) < r.retryTTL {
		if r.lastErr == nil {
			return errors.New("logos mapping not ready")
		}
		return r.lastErr
	}

	r.lastAttempt = now
	mapping, err := r.fetchMapping(ctx)
	if err != nil {
		r.lastErr = err
		r.logger.Printf("logos mapping fetch failed: %v", err)
		return err
	}
	r.symbolToURL = mapping
	r.lastFetched = now
	r.lastErr = nil
	return nil
}

func (r *logoResolver) fetchMapping(ctx context.Context) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.hkMappingURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("logos mapping status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var parsed xhkgLogoMapping
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if parsed.SymbolToFile == nil {
		return nil, errors.New("logos mapping missing symbol_to_file")
	}

	mapped := make(map[string]string, len(parsed.SymbolToFile))
	for symbol, file := range parsed.SymbolToFile {
		file = strings.TrimSpace(file)
		if symbol == "" || file == "" {
			continue
		}
		mapped[strings.TrimSpace(symbol)] = file
	}
	return mapped, nil
}

func normalizeLogoBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return ""
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base
}
