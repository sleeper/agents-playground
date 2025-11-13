package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/agents-playground/internal/storage/sqlite"
)

// PageHandler manages page endpoints.
type PageHandler struct {
	store *sqlite.Store
}

// NewPageHandler constructs handler.
func NewPageHandler(store *sqlite.Store) *PageHandler {
	return &PageHandler{store: store}
}

// CreatePageRequest is the payload for POST /api/pages.
type CreatePageRequest struct {
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	Content      string   `json:"content"`
	ParentPageID *string  `json:"parent_page_id"`
	Tags         []string `json:"tags"`
}

// CreatePage handles page creation.
func (h *PageHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	var req CreatePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Envelope{Errors: []APIError{{Message: "invalid request body"}}})
		return
	}
	page, err := h.store.CreatePage(r.Context(), sqlite.CreatePageInput{
		Slug:         req.Slug,
		Title:        req.Title,
		Summary:      req.Summary,
		Content:      req.Content,
		ParentPageID: req.ParentPageID,
		Tags:         req.Tags,
	})
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Envelope{Errors: []APIError{{Message: err.Error()}}})
		return
	}
	respondJSON(w, http.StatusCreated, Envelope{Data: page})
}

// GetPage returns a page by id.
func (h *PageHandler) GetPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	page, err := h.store.GetPage(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Envelope{Errors: []APIError{{Message: err.Error()}}})
		return
	}
	if page == nil {
		respondJSON(w, http.StatusNotFound, Envelope{Errors: []APIError{{Message: "page not found"}}})
		return
	}
	respondJSON(w, http.StatusOK, Envelope{Data: page})
}
