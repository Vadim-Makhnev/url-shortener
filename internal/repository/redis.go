package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultCacheTTL = 24 * time.Hour
)

type RedisRepository struct {
	redis *redis.Client
}

func NewRedisRepository(redis *redis.Client) *RedisRepository {
	return &RedisRepository{
		redis: redis,
	}
}

func (r *RedisRepository) Set(ctx context.Context, shortCode, originalURL string) error {
	if err := r.redis.Set(ctx, shortCode, originalURL, defaultCacheTTL).Err(); err != nil {
		return fmt.Errorf("redis: failed to set key %s: %w", shortCode, err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, shortCode string) (string, error) {

	val, err := r.redis.Get(ctx, shortCode).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("redis: key not found: %s", shortCode)
		}
		return "", fmt.Errorf("redis: can't get value by key %s: %w", shortCode, err)
	}

	return val, nil
}
