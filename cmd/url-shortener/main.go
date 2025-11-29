package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/config"
	"github.com/Vadim-Makhnev/url-shortener/internal/handler"
	"github.com/Vadim-Makhnev/url-shortener/internal/metrics"
	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
	"github.com/Vadim-Makhnev/url-shortener/internal/service"
	"github.com/joho/godotenv"
)

type application struct {
	port    string
	handler *handler.URLHandler
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf(".env file not loaded: %v", err)
	}

	addr := flag.String("addr", ":"+os.Getenv("PORT"), "HTTP network address")
	flag.Parse()

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

	app := application{
		handler: urlHandler,
	}

	srv := &http.Server{
		Handler:      app.routes(),
		Addr:         *addr,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	log.Fatal(err)
}
