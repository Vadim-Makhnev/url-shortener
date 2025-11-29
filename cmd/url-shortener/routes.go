package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/shorten", app.handler.ShortenURL).Methods("POST")
	api.HandleFunc("/urls", app.handler.GetURLs).Methods("GET")

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/{shortCode}", app.handler.RedirectURL).Methods("GET")

	return r
}
