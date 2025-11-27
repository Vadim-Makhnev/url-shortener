package service

import (
	"log/slog"
	"math/rand"

	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
)

type RepositoryPostgres interface {
	CreateURL(shortCode, originalURL string) (*repository.URL, error)
	GetURLByShortCode(shortCode string) (string, error)
	GetAllURLS() ([]repository.URL, error)
}

type RepositoryRedis interface {
	Set(shortCode, originalURL string) error
	Get(shortCode string) (string, error)
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

func (s *URLService) ShortenURL(originalURL string) (*repository.URL, error) {
	shortCode := generateShortCode()

	url, err := s.postgres.CreateURL(shortCode, originalURL)
	if err != nil {
		s.logger.Error("ShortenURL:", "error", err)
		return nil, err
	}

	err = s.redis.Set(shortCode, originalURL)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	val, err := s.redis.Get(shortCode)
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

func (s *URLService) GetAllURLS() ([]repository.URL, error) {
	return s.postgres.GetAllURLS()
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
