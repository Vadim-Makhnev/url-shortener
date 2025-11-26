package mocks

import (
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
)

type MockRepo struct {
	urls map[string]string
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		urls: make(map[string]string),
	}
}

func (m *MockRepo) CreateURL(shortCode, originalURL string) (*repository.URL, error) {
	m.urls[shortCode] = originalURL

	id := len(m.urls)

	url := &repository.URL{
		ID:          id,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now().UTC(),
	}
	return url, nil
}

func (m *MockRepo) GetURLByShortCode(shortCode string) (string, error) {
	val, ok := m.urls[shortCode]

	if !ok {
		return "", repository.ErrNotFound
	}

	return val, nil
}

func (m *MockRepo) GetAllURLS() ([]repository.URL, error) {
	var urls []repository.URL
	now := time.Now().UTC()

	id := 1
	for shortCode, originalURL := range m.urls {
		urls = append(urls, repository.URL{
			ID:          id,
			ShortCode:   shortCode,
			OriginalURL: originalURL,
			CreatedAt:   now,
		})
		id++
	}

	return urls, nil
}
