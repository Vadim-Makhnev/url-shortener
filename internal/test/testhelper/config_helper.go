package testhelper

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/config"
)

func NewDatabaseConfig(t *testing.T) *config.DatabaseConfig {
	t.Helper()

	return &config.DatabaseConfig{
		Postgres: NewTestPostgresConfig(),
		Redis:    NewTestRedisConfig(),
	}
}

func NewTestPostgresConfig() *config.PostgresConfig {
	return &config.PostgresConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5438"),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "password"),
		DBName:   getEnv("TEST_DB_NAME", "url_shortener_test"),
		SSLMode:  getEnv("TEST_DB_SSL_MODE", "disable"),
	}
}

func NewTestRedisConfig() *config.RedisConfig {
	return &config.RedisConfig{
		Host:     getEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getEnv("TEST_REDIS_PORT", "6381"),
		Password: getEnv("TEST_REDIS_PASSWORD", ""),
		DB:       0,
	}
}

func NewTestDatabaseConnections(t *testing.T) *config.DatabaseConnections {
	t.Helper()

	testConfig := NewDatabaseConfig(t)
	connections, err := config.NewDatabaseConnections(testConfig)

	if err != nil {
		t.Fatalf("Failed to create test database connections: %v", err)
	}

	createTestSchema(t, connections.Postgres)

	return connections
}

func CleanupTestDatabase(t *testing.T, connections *config.DatabaseConnections, tables ...string) {
	query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(tables, ", "))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := connections.Postgres.Exec(query)
	if err != nil {
		t.Fatalf("Failed to cleanup postgres: %v", err)
	}

	err = connections.Redis.FlushDB(ctx).Err()
	if err != nil {
		t.Fatalf("Failed to cleanup redis: %v", err)
	}

}

func createTestSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		click_count INTEGER DEFAULT 0
	);
	
	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	`

	_, err := db.Exec(query)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}
}

func getEnv(key, defaultValue string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
