package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/example/agents-playground/internal/config"
	"github.com/example/agents-playground/internal/http/handlers"
	"github.com/example/agents-playground/internal/logging"
	"github.com/example/agents-playground/internal/storage/sqlite"
)

// NewRouter wires up HTTP routing for the platform API.
func NewRouter(cfg config.Config, store *sqlite.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(logging.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	pageHandler := handlers.NewPageHandler(store)
	databaseHandler := handlers.NewDatabaseHandler(store)

	r.Get("/", handlers.IndexHandler())
	r.Get("/favicon.ico", handlers.FaviconHandler())

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", handlers.HealthHandler(store))
		api.Get("/metrics", handlers.MetricsHandler())
		api.Get("/config", handlers.ConfigHandler(cfg))

		api.Route("/pages", func(pr chi.Router) {
			pr.Post("/", pageHandler.CreatePage)
			pr.Route("/{id}", func(r chi.Router) {
				r.Get("/", pageHandler.GetPage)
			})
		})

		api.Route("/databases", func(dr chi.Router) {
			dr.Post("/", databaseHandler.CreateDatabase)
			dr.Route("/{id}", func(r chi.Router) {
				r.Get("/", databaseHandler.GetDatabase)
				r.Post("/items", databaseHandler.CreateItem)
				r.Get("/views/{viewID}/items", databaseHandler.ListViewItems)
			})
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, handlers.Envelope{Errors: []handlers.APIError{{Message: "resource not found"}}})
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
