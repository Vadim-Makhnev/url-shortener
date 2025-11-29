package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Vadim-Makhnev/url-shortener/internal/handler"
	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
	"github.com/Vadim-Makhnev/url-shortener/internal/service"
	"github.com/Vadim-Makhnev/url-shortener/internal/test/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestMain_WithTestDB(t *testing.T) {
	connections := testhelper.NewTestDatabaseConnections(t)
	defer connections.Close()

	testhelper.CleanupTestDatabase(t, connections, "urls")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postgres := repository.NewRepositoryPostgres(logger, connections.Postgres)
	redis := repository.NewRedisRepository(connections.Redis)
	urlService := service.NewService(postgres, redis, logger)
	urlHandler := handler.NewHandler(urlService)

	app := application{
		handler: urlHandler,
	}

	router := app.routes()

	t.Run("CreateShortURL", func(t *testing.T) {
		requestBody := map[string]string{"url": "https://example.com"}
		jsonData, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response map[string]any
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "short_url")
		assert.Contains(t, response, "original_url")
		assert.Contains(t, response, "created_at")
	})

	t.Run("RedirectURL", func(t *testing.T) {
		createReq := map[string]string{"url": "https://google.com"}
		jsonData, _ := json.Marshal(createReq)

		req := httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		var createResponse map[string]any
		json.Unmarshal(rr.Body.Bytes(), &createResponse)

		shortURL := createResponse["short_url"].(string)
		shortCode := strings.TrimPrefix(shortURL, "/")

		req = httptest.NewRequest("GET", "/"+shortCode, nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code == http.StatusMovedPermanently || rr.Code == http.StatusFound {
			assert.Equal(t, "https://google.com", rr.Header().Get("Location"))
		} else {
			t.Errorf("Expected redirect, got status %d", rr.Code)
		}
	})

	t.Run("GetAllURLs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/urls", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "OK", rr.Body.String())
	})
}
