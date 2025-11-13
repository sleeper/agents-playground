package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/example/agents-playground/internal/config"
	"github.com/example/agents-playground/internal/storage/sqlite"
)

func TestRouterServesIndexHTML(t *testing.T) {
	store, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	router := NewRouter(config.Config{}, store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, "text/html; charset=utf-8", resp.Header().Get("Content-Type"))
	require.Contains(t, resp.Body.String(), "Local Notion")
}

func TestRouterHandlesFaviconRequest(t *testing.T) {
	store, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	router := NewRouter(config.Config{}, store)

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusNoContent, resp.Code)
	require.Empty(t, resp.Body.Len())
}
