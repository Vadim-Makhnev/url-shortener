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

	cmd := r.redis.Get(ctx, shortCode)
	if cmd.Err() != nil {
		return "", fmt.Errorf("redis: can't get value by key: %s", shortCode)
	}

	return cmd.Val(), nil
}
