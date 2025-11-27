package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	redis *redis.Client
}

func NewRedisRepository(redis *redis.Client) *RedisRepository {
	return &RedisRepository{
		redis: redis,
	}
}

func (r *RedisRepository) Set(shortCode, originalURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.redis.Set(ctx, shortCode, originalURL, time.Duration(24*time.Hour)).Err()
}

func (r *RedisRepository) Get(shortCode string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, err := r.redis.Get(ctx, shortCode).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("redis: key not found: %s", shortCode)
		}
		return "", fmt.Errorf("redis: can't get value by key %s: %w", shortCode, err)
	}

	return val, nil
}
