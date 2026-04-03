package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStore struct {
	client *redis.Client
}

func newRedisStore(cfg Config) (*redisStore, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &redisStore{client: client}, nil
}

type idempotencyRecord struct {
	Hash       string            `json:"hash"`
	StatusCode int               `json:"status_code"`
	Body       json.RawMessage   `json:"body"`
	Headers    map[string]string `json:"headers"`
}

func (store *redisStore) getIdempotency(ctx context.Context, key string) (*idempotencyRecord, error) {
	value, err := store.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var record idempotencyRecord
	if err := json.Unmarshal([]byte(value), &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (store *redisStore) setIdempotency(ctx context.Context, key string, record idempotencyRecord, ttl time.Duration) error {
	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return store.client.Set(ctx, key, encoded, ttl).Err()
}

func (store *redisStore) incrRate(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := store.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := store.client.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (store *redisStore) getJSON(ctx context.Context, key string, dest any) (bool, error) {
	value, err := store.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false, err
	}
	return true, nil
}

func (store *redisStore) setJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.client.Set(ctx, key, encoded, ttl).Err()
}

func (store *redisStore) del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return store.client.Del(ctx, keys...).Err()
}

func (store *redisStore) setNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	return store.client.SetNX(ctx, key, value, ttl).Result()
}

func (store *redisStore) flushAll(ctx context.Context) error {
	return store.client.FlushDB(ctx).Err()
}
