package service

import (
	"log/slog"
	"math/rand"

	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
)

type Repository interface {
	CreateURL(shortCode, originalURL string) (*repository.URL, error)
	GetURLByShortCode(shortCode string) (string, error)
	GetAllURLS() ([]repository.URL, error)
}

type URLService struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *URLService {
	return &URLService{
		repo:   repo,
		logger: logger,
	}
}

func (s *URLService) ShortenURL(originalURL string) (*repository.URL, error) {
	shortCode := generateShortCode()

	url, err := s.repo.CreateURL(shortCode, originalURL)
	if err != nil {
		s.logger.Error("ShortenURL:", "error", err)
		return nil, err
	}

	return url, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	url, err := s.repo.GetURLByShortCode(shortCode)
	if err != nil {
		s.logger.Error("GetOriginalURL:", "error", err)
		return "", err
	}

	return url, nil
}

func (s *URLService) GetAllURLS() ([]repository.URL, error) {
	return s.repo.GetAllURLS()
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
