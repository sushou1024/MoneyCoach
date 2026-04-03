package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func portfolioImagePaths(t *testing.T, portfolio string) []string {
	t.Helper()

	dir := filepath.Join("testdata", "portfolios", portfolio, "images")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read portfolio images dir %s: %v", dir, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp":
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatalf("no images found in %s", dir)
	}
	return paths
}

func portfolioTradeImagePaths(t *testing.T, portfolio string) []string {
	t.Helper()

	dir := filepath.Join("testdata", "portfolios", portfolio, "trades")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read portfolio trades dir %s: %v", dir, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp":
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatalf("no trade images found in %s", dir)
	}
	return paths
}

func portfolioTradeTextPaths(t *testing.T, portfolio string) []string {
	t.Helper()

	dir := filepath.Join("testdata", "portfolios", portfolio, "trades")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read portfolio trades dir %s: %v", dir, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".txt") {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatalf("no trade text files found in %s", dir)
	}
	return paths
}

func imageMimeType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}
