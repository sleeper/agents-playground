package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/agents-playground/internal/config"
	"github.com/example/agents-playground/internal/storage/sqlite"
)

// HealthHandler returns liveness info.
func HealthHandler(store *sqlite.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(r.Context()); err != nil {
			respondJSON(w, http.StatusServiceUnavailable, Envelope{Errors: []APIError{{Message: err.Error()}}})
			return
		}
		respondJSON(w, http.StatusOK, Envelope{Data: map[string]any{"status": "ok", "time": time.Now().UTC()}})
	}
}

// MetricsHandler exposes simple metrics placeholder.
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte("# TYPE platform_requests_total counter\n"))
		_, _ = w.Write([]byte("platform_requests_total 0\n"))
	}
}

// ConfigHandler exposes runtime configuration snapshot.
func ConfigHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cfg)
	}
}
