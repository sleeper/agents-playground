package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/example/agents-playground/internal/storage/sqlite"
)

type responseEnvelope struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func TestPageHandlerCreatePageSuccess(t *testing.T) {
	store := newTestSQLiteStore(t)
	handler := NewPageHandler(store)

	body := map[string]any{
		"slug":            "welcome",
		"title":           "Welcome",
		"summary":         "intro",
		"content":         "hello",
		"linked_page_ids": []string{},
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pages", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	handler.CreatePage(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	var env responseEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&env))
	require.Len(t, env.Errors, 0)
	require.NotEmpty(t, env.Data)
}

func TestPageHandlerCreatePageInvalidJSON(t *testing.T) {
	store := newTestSQLiteStore(t)
	handler := NewPageHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/pages", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	handler.CreatePage(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var env responseEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&env))
	require.Len(t, env.Errors, 1)
	require.Contains(t, env.Errors[0].Message, "invalid request body")
}

func TestPageHandlerListPages(t *testing.T) {
	store := newTestSQLiteStore(t)
	handler := NewPageHandler(store)

	_, err := store.CreatePage(context.Background(), sqlite.CreatePageInput{Slug: "alpha", Title: "Alpha"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/pages", nil)
	rec := httptest.NewRecorder()

	handler.ListPages(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var env responseEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&env))
	require.Len(t, env.Errors, 0)

	var pages []map[string]any
	require.NoError(t, json.Unmarshal(env.Data, &pages))
	require.Len(t, pages, 1)
	require.Equal(t, "alpha", pages[0]["slug"])
}

func newTestSQLiteStore(t *testing.T) *sqlite.Store {
	t.Helper()
	store, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}
