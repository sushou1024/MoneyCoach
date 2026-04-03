package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type storageService interface {
	presignPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error)
	headObject(ctx context.Context, key string) error
	getObjectBytes(ctx context.Context, key string) ([]byte, error)
}

type s3StorageClient struct {
	bucket  string
	region  string
	client  *s3.Client
	presign *s3.PresignClient
}

func newStorageClient(ctx context.Context, cfg Config) (storageService, error) {
	if cfg.ObjectStorageMode == "local" {
		return newLocalStorageClient(cfg)
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.ObjectStorageRegion))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg)
	return &s3StorageClient{
		bucket:  cfg.ObjectStorageBucket,
		region:  cfg.ObjectStorageRegion,
		client:  client,
		presign: s3.NewPresignClient(client),
	}, nil
}

func (s *s3StorageClient) presignPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(ttl)
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}
	resp, err := s.presign.PresignPutObject(ctx, input, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", nil, time.Time{}, fmt.Errorf("presign put: %w", err)
	}
	return resp.URL, map[string]string{"Content-Type": contentType}, expiresAt, nil
}

func (s *s3StorageClient) headObject(ctx context.Context, key string) error {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("head object %s: %w", key, err)
	}
	return nil
}

func (s *s3StorageClient) getObjectBytes(ctx context.Context, key string) ([]byte, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read object %s: %w", key, err)
	}
	return data, nil
}
