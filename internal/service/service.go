package service

import (
	"log/slog"
	"math/rand"

	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
)

type Repository interface {
	CreateURL(url *repository.URL) error
	GetURLByShortCode(shortCode string) (*repository.URL, error)
	GetAllURLS() ([]repository.URL, error)
	IncrementAccessCount(shortCode string) error
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

func (s *URLService) ShortenURL(originalURL string) (string, error) {
	shortCode := generateShortCode()

	url := &repository.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}

	if err := s.repo.CreateURL(url); err != nil {
		s.logger.Error("ShortenURL:", "error", err)
		return "", err
	}

	return shortCode, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	url, err := s.repo.GetURLByShortCode(shortCode)
	if err != nil {
		s.logger.Error("GetOriginalURL:", "error", err)
		return "", nil
	}

	go s.repo.IncrementAccessCount(shortCode)

	return url.OriginalURL, nil
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
