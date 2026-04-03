package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type localStorageClient struct {
	rootDir string
	baseURL string
}

func newLocalStorageClient(cfg Config) (*localStorageClient, error) {
	rootDir := strings.TrimSpace(cfg.ObjectStorageLocalDir)
	if rootDir == "" {
		rootDir = "local-uploads"
	}
	if err := os.MkdirAll(rootDir, 0o750); err != nil {
		return nil, fmt.Errorf("create local storage dir: %w", err)
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.LocalStorageBaseURL), "/")
	if baseURL == "" {
		baseURL = "http://localhost:" + cfg.Port
	}
	return &localStorageClient{rootDir: rootDir, baseURL: baseURL}, nil
}

func (l *localStorageClient) presignPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(ttl)
	pathOnDisk, err := l.objectPath(key)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	if err := os.MkdirAll(filepath.Dir(pathOnDisk), 0o750); err != nil {
		return "", nil, time.Time{}, fmt.Errorf("create local upload dir: %w", err)
	}
	uploadURL := l.baseURL + "/v1/local-uploads/" + key
	return uploadURL, map[string]string{"Content-Type": contentType}, expiresAt, nil
}

func (l *localStorageClient) headObject(ctx context.Context, key string) error {
	pathOnDisk, err := l.objectPath(key)
	if err != nil {
		return err
	}
	info, err := os.Stat(pathOnDisk)
	if err != nil {
		return fmt.Errorf("stat object %s: %w", key, err)
	}
	if info.IsDir() {
		return fmt.Errorf("object %s is a directory", key)
	}
	return nil
}

func (l *localStorageClient) getObjectBytes(ctx context.Context, key string) ([]byte, error) {
	pathOnDisk, err := l.objectPath(key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(pathOnDisk)
	if err != nil {
		return nil, fmt.Errorf("read object %s: %w", key, err)
	}
	return data, nil
}

func (l *localStorageClient) putObject(ctx context.Context, key string, src io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	pathOnDisk, err := l.objectPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pathOnDisk), 0o750); err != nil {
		return fmt.Errorf("create local upload dir: %w", err)
	}
	file, err := os.OpenFile(pathOnDisk, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return fmt.Errorf("open local object: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, src); err != nil {
		return fmt.Errorf("write local object: %w", err)
	}
	return nil
}

func (l *localStorageClient) objectPath(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("empty storage key")
	}
	if strings.Contains(key, "..") {
		return "", errors.New("invalid storage key")
	}
	clean := path.Clean("/" + key)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || clean == "." {
		return "", errors.New("invalid storage key")
	}
	return filepath.Join(l.rootDir, filepath.FromSlash(clean)), nil
}
