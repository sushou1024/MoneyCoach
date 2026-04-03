package app

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}
	sort.Strings(unique)
	return unique
}

func hashStrings(values []string) string {
	if len(values) == 0 {
		return ""
	}
	payload := strings.Join(values, ",")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:8])
}
