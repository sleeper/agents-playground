package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/example/potion/internal/storage"
)

// ServerConfig holds configuration for the HTTP server.
type ServerConfig struct {
	DBPath string
}

// Server is the main HTTP server for the Potion backend.
type Server struct {
	repo *storage.Repository
}

// NewServer constructs a server using the provided configuration.
func NewServer(cfg ServerConfig) (*Server, error) {
	repo, err := storage.NewRepository(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	return &Server{repo: repo}, nil
}

// Close releases resources used by the server.
func (s *Server) Close() error {
	return s.repo.Close()
}

// Router builds the HTTP router with all API routes.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/pages", func(r chi.Router) {
		r.Get("/", s.handleListPages)
		r.Post("/", s.handleCreatePage)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handleGetPage)
			r.Put("/", s.handleUpdatePage)
			r.Get("/links", s.handleListPageLinks)
		})
	})

	r.Route("/links", func(r chi.Router) {
		r.Post("/", s.handleCreateLink)
	})

	r.Route("/databases", func(r chi.Router) {
		r.Get("/", s.handleListDatabases)
		r.Post("/", s.handleCreateDatabase)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handleGetDatabase)
			r.Get("/entries", s.handleListDatabaseEntries)
			r.Post("/entries", s.handleCreateDatabaseEntry)
		})
	})

	return r
}

// Run starts the HTTP server at the provided address.
func (s *Server) Run(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Router(),
	}
	return srv.ListenAndServe()
}

func (s *Server) handleCreatePage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string  `json:"title"`
		Content  string  `json:"content"`
		ParentID *string `json:"parentId"`
	}
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}

	page, err := s.repo.CreatePage(r.Context(), req.Title, req.Content, req.ParentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, page)
}

func (s *Server) handleGetPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	page, err := s.repo.GetPage(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) handleListPages(w http.ResponseWriter, r *http.Request) {
	pages, err := s.repo.ListPages(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, pages)
}

func (s *Server) handleUpdatePage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Title    string  `json:"title"`
		Content  string  `json:"content"`
		ParentID *string `json:"parentId"`
	}
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	page, err := s.repo.UpdatePage(r.Context(), id, req.Title, req.Content, req.ParentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceID string `json:"sourceId"`
		TargetID string `json:"targetId"`
	}
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.SourceID == "" || req.TargetID == "" {
		writeError(w, http.StatusBadRequest, errors.New("sourceId and targetId are required"))
		return
	}

	link, err := s.repo.CreatePageLink(r.Context(), req.SourceID, req.TargetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func (s *Server) handleListPageLinks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	links, err := s.repo.ListPageLinks(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, links)
}

func (s *Server) handleCreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string            `json:"title"`
		Description string            `json:"description"`
		View        string            `json:"view"`
		Schema      map[string]string `json:"schema"`
	}
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	if req.View == "" {
		req.View = "table"
	}
	if req.Schema == nil {
		req.Schema = map[string]string{}
	}

	dbModel, err := s.repo.CreateDatabase(r.Context(), req.Title, req.Description, req.View, req.Schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, dbModel)
}

func (s *Server) handleGetDatabase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	dbModel, err := s.repo.GetDatabase(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dbModel)
}

func (s *Server) handleListDatabases(w http.ResponseWriter, r *http.Request) {
	dbs, err := s.repo.ListDatabases(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, dbs)
}

func (s *Server) handleCreateDatabaseEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Title      string                 `json:"title"`
		Properties map[string]interface{} `json:"properties"`
	}
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	if req.Properties == nil {
		req.Properties = map[string]interface{}{}
	}

	entry, err := s.repo.CreateDatabaseEntry(r.Context(), id, req.Title, req.Properties)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (s *Server) handleListDatabaseEntries(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entries, err := s.repo.ListDatabaseEntries(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func decodeJSON(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
