package main

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"
)

const debugTestEnvVar = "DEBUG_TEST_OUTPUT"

func debugEnvEnabled() bool {
	value, ok := os.LookupEnv(debugTestEnvVar)
	if !ok {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" || normalized == "0" || normalized == "false" || normalized == "no" {
		return false
	}
	return true
}

func verboseLogsEnabled() bool {
	return testing.Verbose() || debugEnvEnabled()
}

func debugLogf(t *testing.T, format string, args ...any) {
	t.Helper()
	if !verboseLogsEnabled() {
		return
	}
	t.Logf(format, args...)
}

func debugLogJSON(t *testing.T, label string, value any) {
	t.Helper()
	if !debugEnvEnabled() {
		return
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Logf("%s JSON marshal failed: %v", label, err)
		return
	}
	t.Logf("%s parsed JSON:\n%s", label, pretty)
}

func sanitizeURL(raw *url.URL) string {
	if raw == nil {
		return ""
	}
	sanitized := *raw
	sanitized.User = nil

	if raw.RawQuery == "" {
		return sanitized.String()
	}

	values := raw.Query()
	for key := range values {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "key") || strings.Contains(lower, "token") || lower == "app_id" {
			values.Set(key, "REDACTED")
		}
	}
	sanitized.RawQuery = values.Encode()
	return sanitized.String()
}
