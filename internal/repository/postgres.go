package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type URL struct {
	ID          int       `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type URLRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepositoryPostgres(logger *slog.Logger, postgres *sql.DB) *URLRepository {

	return &URLRepository{
		db:     postgres,
		logger: logger,
	}
}

func (r *URLRepository) CreateURL(shortCode, originalURL string) (*URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int
	var createdAt time.Time

	query := `INSERT INTO urls (short_code, original_url) VALUES ($1, $2)
			RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, shortCode, originalURL).Scan(&id, &createdAt)
	if err != nil {
		r.logger.Error("CreateURL", "original_url", originalURL, "short_code", shortCode, "id", id, "error", err)
		return nil, fmt.Errorf("repository: CreateURL: %w", err)
	}
	return &URL{
		ID:          id,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   createdAt,
	}, nil
}

func (r *URLRepository) GetURLByShortCode(shortCode string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var originalURL string
	query := `SELECT original_url FROM urls 
			WHERE short_code = $1`

	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		r.logger.Error("GetURLByshortCode", "short_code", shortCode, "error", err)
		return "", fmt.Errorf("repository: GetURLByShortCode: %w", err)
	}

	return originalURL, nil
}

func (r *URLRepository) GetAllURLS() ([]URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, short_code, original_url, created_at FROM urls ORDER BY
	created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("GetAllURLS", "error", err)
		return nil, fmt.Errorf("repository: GetAllURLS: %w", err)
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var url URL
		if err := rows.Scan(&url.ID, &url.ShortCode, &url.OriginalURL, &url.CreatedAt); err != nil {
			r.logger.Error("GetAllURLS scan", "error", err)
			return nil, err
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("GetAllURLS rows", "error", err)
		return nil, err
	}

	return urls, nil
}
