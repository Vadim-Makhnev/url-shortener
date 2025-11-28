package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/metrics"
	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type URLService interface {
	ShortenURL(originalURL string) (*repository.URL, error)
	GetOriginalURL(shortCode string) (string, error)
	GetAllURLS() ([]repository.URL, error)
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type URLResponse struct {
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type URLHandler struct {
	service URLService
}

func NewHandler(service URLService) *URLHandler {
	return &URLHandler{service: service}
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	metrics.URLShortenRequests.Inc()
	timer := prometheus.NewTimer(metrics.RequestDuration)
	defer timer.ObserveDuration()

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	url, err := h.service.ShortenURL(req.URL)
	if err != nil {
		http.Error(w, "failed to shorten URL", http.StatusInternalServerError)
		return
	}

	var res URLResponse

	res.ShortURL = addScheme(url.ShortCode)
	res.CreatedAt = url.CreatedAt
	res.OriginalURL = url.OriginalURL

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	metrics.URLRedirectRequests.Inc()
	timer := prometheus.NewTimer(metrics.RequestDuration)
	defer timer.ObserveDuration()

	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	originalURL, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	metrics.URLAccessCount.WithLabelValues(shortCode).Inc()

	redirectURL := addScheme(originalURL)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *URLHandler) GetURLs(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.GetAllURLS()
	if err != nil {
		http.Error(w, "Failed to get URLs", http.StatusInternalServerError)
		return
	}

	var urls []URLResponse

	for _, url := range list {
		urls = append(urls, URLResponse{
			ShortURL:    addScheme(url.ShortCode),
			OriginalURL: url.OriginalURL,
			CreatedAt:   url.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}

func addScheme(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	return "https://" + url
}
