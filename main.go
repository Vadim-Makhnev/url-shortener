package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Vadim-Makhnev/url-shortener/internal/config"
	"github.com/Vadim-Makhnev/url-shortener/internal/handler"
	"github.com/Vadim-Makhnev/url-shortener/internal/metrics"
	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
	"github.com/Vadim-Makhnev/url-shortener/internal/service"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf(".env file not loaded: %v", err)
	}

	metrics.InitMetrics()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	dbConfig := config.NewDatabaseConfig()
	dbConnections, err := config.NewDatabaseConnections(dbConfig)
	if err != nil {
		log.Fatalf("initialize database connections: %v", err)
	}
	defer dbConnections.Close()

	postgres := repository.NewRepositoryPostgres(
		logger,
		dbConnections.Postgres,
	)

	redis := repository.NewRedisRepository(dbConnections.Redis)

	urlService := service.NewService(postgres, redis, logger)

	urlHandler := handler.NewHandler(urlService)

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/shorten", urlHandler.ShortenURL).Methods("POST")
	api.HandleFunc("/urls", urlHandler.GetURLs).Methods("GET")

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/{shortCode}", urlHandler.RedirectURL).Methods("GET")

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
