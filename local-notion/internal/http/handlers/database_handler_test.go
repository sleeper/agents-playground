package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/example/agents-playground/internal/domain"
	"github.com/example/agents-playground/internal/storage/sqlite"
)

func TestDatabaseHandlerListViewItemsNotFound(t *testing.T) {
	store := newTestSQLiteStore(t)
	handler := NewDatabaseHandler(store)

	db, err := store.CreateDatabase(context.Background(), sqlite.CreateDatabaseInput{
		Slug:       "inventory",
		Title:      "Inventory",
		Properties: []sqlite.DatabasePropertyInput{{Name: "Name", Slug: "name", Type: domain.PropertyTypeText}},
		Views:      []sqlite.DatabaseViewInput{{Name: "Default", Type: domain.ViewTypeTable}},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/databases/"+db.ID+"/views/missing/items", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", db.ID)
	rctx.URLParams.Add("viewID", "missing")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.ListViewItems(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNotFound, res.StatusCode)

	var env responseEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&env))
	require.Len(t, env.Errors, 1)
	require.Contains(t, env.Errors[0].Message, "view not found")
}
