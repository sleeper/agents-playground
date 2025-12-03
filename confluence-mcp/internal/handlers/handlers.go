package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"confluence-mcp/internal/auth"
	"confluence-mcp/internal/confluence"
	"confluence-mcp/internal/policy"
)

// ContextKey identifies per-request values.
type ContextKey string

const subjectKey ContextKey = "subject"

// API wires dependencies and exposes HTTP handlers.
type API struct {
	Client      *confluence.Client
	Auth        *auth.Authenticator
	PolicyStore *policy.Store
}

// Router builds the chi router with authentication and core endpoints.
func (a *API) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Group(func(r chi.Router) {
		r.Use(a.authMiddleware)
		r.Get("/spaces", a.listSpaces)
		r.Get("/search", a.searchContent)
		r.Get("/pages/{id}", a.getPage)
	})

	return r
}

func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subject, err := a.Auth.SubjectFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		ctx = contextWithSubject(ctx, subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) listSpaces(w http.ResponseWriter, r *http.Request) {
	subject := subjectFromContext(r.Context())
	spaces := a.PolicyStore.AllowedSpaces(subject)
	writeJSON(w, http.StatusOK, map[string]any{"spaces": spaces})
}

func (a *API) searchContent(w http.ResponseWriter, r *http.Request) {
	subject := subjectFromContext(r.Context())
	query := r.URL.Query().Get("q")
	limit := 25
	results, err := a.Client.Search(r.Context(), query, a.PolicyStore.AllowedSpaces(subject), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (a *API) getPage(w http.ResponseWriter, r *http.Request) {
	subject := subjectFromContext(r.Context())
	pageID := chi.URLParam(r, "id")

	page, err := a.Client.GetPage(r.Context(), pageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Enforce space scoping
	allowedSpaces := a.PolicyStore.AllowedSpaces(subject)
	allowed := false
	for _, s := range allowedSpaces {
		if s == page.Space.Key {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "space not permitted", http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func contextWithSubject(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, subjectKey, subject)
}

func subjectFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(subjectKey).(string); ok {
		return v
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
