package server

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"

    "github.com/example/potion-revamp/backend/internal/models"
    "github.com/example/potion-revamp/backend/internal/storage"
)

// Server exposes HTTP handlers for the Potion workspace.
type Server struct {
    store *storage.Store
}

// New constructs a server instance bound to the given store.
func New(store *storage.Store) *Server {
    return &Server{store: store}
}

// Router builds the HTTP router with all endpoints registered.
func (s *Server) Router() http.Handler {
    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(cors.AllowAll().Handler)

    r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    r.Route("/pages", func(r chi.Router) {
        r.Get("/", s.handleListPages)
        r.Post("/", s.handleCreatePage)
        r.Route("/{pageID}", func(r chi.Router) {
            r.Get("/", s.handleGetPage)
            r.Put("/", s.handleUpdatePage)
            r.Delete("/", s.handleDeletePage)
            r.Put("/blocks", s.handleReplaceBlocks)
        })
    })

    r.Route("/databases", func(r chi.Router) {
        r.Get("/", s.handleListDatabases)
        r.Post("/", s.handleCreateDatabase)
        r.Route("/{databaseID}", func(r chi.Router) {
            r.Get("/", s.handleGetDatabase)
            r.Post("/entries", s.handleCreateEntry)
            r.Route("/entries/{entryID}", func(r chi.Router) {
                r.Put("/", s.handleUpdateEntry)
            })
            r.Get("/views/{viewID}", s.handleResolveView)
        })
    })

    return r
}

func (s *Server) handleListPages(w http.ResponseWriter, r *http.Request) {
    pages, err := s.store.ListPages(r.Context())
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, pages)
}

func (s *Server) handleCreatePage(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Title    string  `json:"title"`
        Icon     string  `json:"icon"`
        ParentID *string `json:"parentId"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondErr(w, http.StatusBadRequest, err)
        return
    }
    page, err := s.store.CreatePage(r.Context(), req.Title, req.ParentID, req.Icon)
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusCreated, page)
}

func (s *Server) handleGetPage(w http.ResponseWriter, r *http.Request) {
    pageID := chi.URLParam(r, "pageID")
    data, err := s.store.GetPageWithBlocks(r.Context(), pageID)
    if err != nil {
        respondErr(w, http.StatusNotFound, err)
        return
    }
    respondJSON(w, http.StatusOK, data)
}

func (s *Server) handleUpdatePage(w http.ResponseWriter, r *http.Request) {
    pageID := chi.URLParam(r, "pageID")
    var req struct {
        Title    string  `json:"title"`
        Icon     string  `json:"icon"`
        ParentID *string `json:"parentId"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondErr(w, http.StatusBadRequest, err)
        return
    }
    page := models.Page{ID: pageID, Title: req.Title, Icon: req.Icon, ParentID: req.ParentID}
    if err := s.store.UpdatePage(r.Context(), page); err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    updated, err := s.store.GetPageWithBlocks(r.Context(), pageID)
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeletePage(w http.ResponseWriter, r *http.Request) {
    pageID := chi.URLParam(r, "pageID")
    if err := s.store.DeletePage(r.Context(), pageID); err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReplaceBlocks(w http.ResponseWriter, r *http.Request) {
    pageID := chi.URLParam(r, "pageID")
    var req struct {
        Blocks []models.Block `json:"blocks"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondErr(w, http.StatusBadRequest, err)
        return
    }
    if req.Blocks == nil {
        req.Blocks = []models.Block{}
    }
    persisted, err := s.store.ReplacePageBlocks(r.Context(), pageID, req.Blocks)
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, map[string]interface{}{"blocks": persisted})
}

func (s *Server) handleListDatabases(w http.ResponseWriter, r *http.Request) {
    items, err := s.store.ListDatabases(r.Context())
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateDatabase(w http.ResponseWriter, r *http.Request) {
    var req models.Database
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondErr(w, http.StatusBadRequest, err)
        return
    }
    created, err := s.store.CreateDatabase(r.Context(), req)
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetDatabase(w http.ResponseWriter, r *http.Request) {
    dbID := chi.URLParam(r, "databaseID")
    data, err := s.store.GetDatabase(r.Context(), dbID)
    if err != nil {
        respondErr(w, http.StatusNotFound, err)
        return
    }
    respondJSON(w, http.StatusOK, data)
}

func (s *Server) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
    dbID := chi.URLParam(r, "databaseID")
    var req struct {
        Title  string                 `json:"title"`
        Values map[string]interface{} `json:"values"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondErr(w, http.StatusBadRequest, err)
        return
    }
    if req.Values == nil {
        req.Values = map[string]interface{}{}
    }
    entry := models.DatabaseEntry{DatabaseID: dbID, Title: req.Title, Properties: req.Values}
    created, err := s.store.CreateDatabaseEntry(r.Context(), entry)
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusCreated, created)
}

func (s *Server) handleUpdateEntry(w http.ResponseWriter, r *http.Request) {
    dbID := chi.URLParam(r, "databaseID")
    entryID := chi.URLParam(r, "entryID")
    var req struct {
        Title  string                 `json:"title"`
        Values map[string]interface{} `json:"values"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondErr(w, http.StatusBadRequest, err)
        return
    }
    if req.Values == nil {
        req.Values = map[string]interface{}{}
    }
    entry := models.DatabaseEntry{ID: entryID, DatabaseID: dbID, Title: req.Title, Properties: req.Values}
    if err := s.store.UpdateDatabaseEntry(r.Context(), entry); err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    updatedEntries, err := s.store.ListDatabaseEntries(r.Context(), dbID)
    if err != nil {
        respondErr(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, map[string]interface{}{"entries": updatedEntries})
}

func (s *Server) handleResolveView(w http.ResponseWriter, r *http.Request) {
    dbID := chi.URLParam(r, "databaseID")
    viewID := chi.URLParam(r, "viewID")
    payload, err := s.store.ResolveViewEntries(r.Context(), dbID, viewID)
    if err != nil {
        respondErr(w, http.StatusNotFound, err)
        return
    }
    respondJSON(w, http.StatusOK, payload)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if payload == nil {
        return
    }
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        log.Printf("failed to encode response: %v", err)
    }
}

func respondErr(w http.ResponseWriter, status int, err error) {
    respondJSON(w, status, map[string]string{"error": err.Error()})
}
