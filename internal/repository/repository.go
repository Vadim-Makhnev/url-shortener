package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound = errors.New("url not found")
)

type URL struct {
	ID          int       `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int       `json:"access_count"`
}

type URLRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewRepository() (*URLRepository, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres driver: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &URLRepository{
		db:    db,
		redis: redisClient,
	}, nil
}

func (r *URLRepository) CreateURL(url *URL) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO urls (short_code, original_url) VALUES ($1, $2)
			RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, url.ShortCode, url.OriginalURL).Scan(&url.ID, &url.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: CreateURL: %w", err)
	}
	return nil
}

func (r *URLRepository) GetURLByShortCode(shortCode string) (*URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cachedURL, err := r.redis.Get(ctx, shortCode).Result()
	if err == nil {
		return &URL{
			ShortCode:   shortCode,
			OriginalURL: cachedURL,
		}, nil
	}

	var url URL
	query := `SELECT id, short_code, original_url, created_at, access_count
			FROM urls WHERE short_code = $1`

	err = r.db.QueryRowContext(ctx, query, shortCode).Scan(&url.ID, &url.ShortCode, &url.OriginalURL, &url.CreatedAt, &url.AccessCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: GetURLByShortCode: %w", err)
	}

	err = r.redis.Set(ctx, shortCode, url.OriginalURL, 24*time.Hour).Err()
	if err != nil {
		log.Printf("redis caching: %v", err)
	}

	return &url, nil
}

func (r *URLRepository) GetAllURLS() ([]URL, error) {
	query := `SELECT id, short_code, original_url, created_at, access_count FROM urls ORDER BY
	created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("repository: GetAllURLS: %w", err)
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var url URL
		if err := rows.Scan(&url.ID, &url.ShortCode, &url.OriginalURL, &url.CreatedAt, &url.AccessCount); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (r *URLRepository) IncrementAccessCount(shortCode string) error {
	query := `UPDATE urls SET access_count = access_count + 1 WHERE short_code = $1`
	_, err := r.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("repository: IncrementAccessCount: %w", err)
	}

	return nil
}
