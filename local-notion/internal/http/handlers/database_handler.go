package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/agents-playground/internal/domain"
	"github.com/example/agents-playground/internal/storage/sqlite"
)

// DatabaseHandler manages database endpoints.
type DatabaseHandler struct {
	store *sqlite.Store
}

// NewDatabaseHandler creates a database handler.
func NewDatabaseHandler(store *sqlite.Store) *DatabaseHandler {
	return &DatabaseHandler{store: store}
}

// CreateDatabaseRequest payload.
type CreateDatabaseRequest struct {
	Slug        string                    `json:"slug"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Icon        *string                   `json:"icon"`
	CoverImage  *string                   `json:"cover_image_id"`
	Properties  []DatabasePropertyRequest `json:"properties"`
	Views       []DatabaseViewRequest     `json:"views"`
}

type DatabasePropertyRequest struct {
	Name       string              `json:"name"`
	Slug       string              `json:"slug"`
	Type       domain.PropertyType `json:"type"`
	Config     map[string]any      `json:"config"`
	IsRequired bool                `json:"is_required"`
	Default    any                 `json:"default"`
	OrderIndex int                 `json:"order_index"`
}

type DatabaseViewRequest struct {
	Name          string            `json:"name"`
	Type          domain.ViewType   `json:"type"`
	Filters       map[string]any    `json:"filters"`
	Sorts         []domain.ViewSort `json:"sorts"`
	Grouping      map[string]any    `json:"grouping"`
	Display       []string          `json:"display_properties"`
	LayoutOptions map[string]any    `json:"layout_options"`
}

// CreateDatabase handles POST /api/databases.
func (h *DatabaseHandler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req CreateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Envelope{Errors: []APIError{{Message: "invalid request body"}}})
		return
	}
	input := sqlite.CreateDatabaseInput{
		Slug:        req.Slug,
		Title:       req.Title,
		Description: req.Description,
		Icon:        req.Icon,
		CoverImage:  req.CoverImage,
	}
	for _, prop := range req.Properties {
		input.Properties = append(input.Properties, sqlite.DatabasePropertyInput{
			Name:       prop.Name,
			Slug:       prop.Slug,
			Type:       prop.Type,
			Config:     prop.Config,
			IsRequired: prop.IsRequired,
			Default:    prop.Default,
			OrderIndex: prop.OrderIndex,
		})
	}
	for _, view := range req.Views {
		input.Views = append(input.Views, sqlite.DatabaseViewInput{
			Name:          view.Name,
			Type:          view.Type,
			Filters:       view.Filters,
			Sorts:         view.Sorts,
			Grouping:      view.Grouping,
			Display:       view.Display,
			LayoutOptions: view.LayoutOptions,
		})
	}
	database, err := h.store.CreateDatabase(r.Context(), input)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Envelope{Errors: []APIError{{Message: err.Error()}}})
		return
	}
	respondJSON(w, http.StatusCreated, Envelope{Data: database})
}

// GetDatabase handles GET /api/databases/{id}.
func (h *DatabaseHandler) GetDatabase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	database, err := h.store.GetDatabase(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Envelope{Errors: []APIError{{Message: err.Error()}}})
		return
	}
	if database == nil {
		respondJSON(w, http.StatusNotFound, Envelope{Errors: []APIError{{Message: "database not found"}}})
		return
	}
	respondJSON(w, http.StatusOK, Envelope{Data: database})
}

// CreateItemRequest handles POST /api/databases/{id}/items.
type CreateItemRequest struct {
	Page struct {
		Slug    string   `json:"slug"`
		Title   string   `json:"title"`
		Summary string   `json:"summary"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	} `json:"page"`
	Position int            `json:"position"`
	Values   map[string]any `json:"values"`
}

// CreateItem creates a new database item.
func (h *DatabaseHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Envelope{Errors: []APIError{{Message: "invalid request body"}}})
		return
	}
	item, err := h.store.CreateDatabaseItem(r.Context(), sqlite.CreateDatabaseItemInput{
		DatabaseID: id,
		Page: sqlite.CreatePageInput{
			Slug:    req.Page.Slug,
			Title:   req.Page.Title,
			Summary: req.Page.Summary,
			Content: req.Page.Content,
			Tags:    req.Page.Tags,
		},
		Position: req.Position,
		Values:   req.Values,
	})
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Envelope{Errors: []APIError{{Message: err.Error()}}})
		return
	}
	respondJSON(w, http.StatusCreated, Envelope{Data: item})
}

// ListViewItems handles GET /api/databases/{id}/views/{viewID}/items.
func (h *DatabaseHandler) ListViewItems(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	viewID := chi.URLParam(r, "viewID")
	items, err := h.store.ListViewItems(r.Context(), id, viewID)
	if err != nil {
		if errors.Is(err, sqlite.ErrViewNotFound) {
			respondJSON(w, http.StatusNotFound, Envelope{Errors: []APIError{{Message: err.Error()}}})
			return
		}
		respondJSON(w, http.StatusInternalServerError, Envelope{Errors: []APIError{{Message: err.Error()}}})
		return
	}
	respondJSON(w, http.StatusOK, Envelope{Data: items})
}
