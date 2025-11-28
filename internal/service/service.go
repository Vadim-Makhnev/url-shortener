package service

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
)

var (
	defaultTimeout = 5 * time.Second
)

type URL struct {
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}

type RepositoryPostgres interface {
	CreateURL(shortCode, originalURL string) (*repository.URL, error)
	GetURLByShortCode(shortCode string) (string, error)
	GetAllURLS() ([]repository.URL, error)
}

type RepositoryRedis interface {
	Set(ctx context.Context, shortCode, originalURL string) error
	Get(ctx context.Context, shortCode string) (string, error)
}

type URLService struct {
	postgres RepositoryPostgres
	logger   *slog.Logger
	redis    RepositoryRedis
}

func NewService(repo RepositoryPostgres, redis RepositoryRedis, logger *slog.Logger) *URLService {
	return &URLService{
		postgres: repo,
		redis:    redis,
		logger:   logger,
	}
}

func (s *URLService) ShortenURL(originalURL string) (*URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	shortCode := generateShortCode()

	url, err := s.postgres.CreateURL(shortCode, originalURL)
	if err != nil {
		s.logger.Error("ShortenURL:", "error", err)
		return nil, err
	}

	err = s.redis.Set(ctx, shortCode, originalURL)
	if err != nil {
		return nil, err
	}

	domainURL := &URL{
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
	}

	return domainURL, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	val, err := s.redis.Get(ctx, shortCode)
	if err == nil {
		return val, nil
	}

	url, err := s.postgres.GetURLByShortCode(shortCode)
	if err != nil {
		s.logger.Error("GetOriginalURL:", "error", err)
		return "", err
	}

	return url, nil
}

func (s *URLService) GetAllURLS() ([]URL, error) {
	urls, err := s.postgres.GetAllURLS()
	if err != nil {
		s.logger.Error("GetAllURLS:", "error", err)
		return nil, err
	}

	var res []URL

	for _, url := range urls {
		res = append(res, URL{
			ShortCode:   url.ShortCode,
			OriginalURL: url.OriginalURL,
			CreatedAt:   url.CreatedAt,
		})
	}

	return res, nil
}

func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
