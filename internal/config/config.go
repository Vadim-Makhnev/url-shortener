package config

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type DatabaseConfig struct {
	Postgres *PostgresConfig
	Redis    *RedisConfig
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Postgres: NewPostgresConfig(),
		Redis:    NewRedisConfig(),
	}
}

type DatabaseConnections struct {
	Postgres *sql.DB
	Redis    *redis.Client
}

func NewDatabaseConnections(config *DatabaseConfig) (*DatabaseConnections, error) {
	postgresDB, err := NewPostgresDB(config.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	redisClient, err := NewRedisClient(config.Redis)
	if err != nil {
		postgresDB.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}

	return &DatabaseConnections{
		Postgres: postgresDB,
		Redis:    redisClient,
	}, nil
}

func (dc *DatabaseConnections) Close() error {
	var errs []error

	if err := dc.Postgres.Close(); err != nil {
		errs = append(errs, fmt.Errorf("postgres close: %w", err))
	}

	if err := dc.Redis.Close(); err != nil {
		errs = append(errs, fmt.Errorf("redis close: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("database connections close errors: %v", errs)
	}

	return nil
}
