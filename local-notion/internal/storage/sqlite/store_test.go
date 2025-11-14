package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/example/agents-playground/internal/domain"
)

func TestStoreCreateAndGetPage(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	page, err := store.CreatePage(ctx, CreatePageInput{Slug: "welcome", Title: "Welcome", Summary: "intro", Content: "hello"})
	require.NoError(t, err)
	require.NotEmpty(t, page.ID)

	fetched, err := store.GetPage(ctx, page.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, page.Title, fetched.Title)
	require.Equal(t, page.Slug, fetched.Slug)
	require.Equal(t, page.Summary, fetched.Summary)
}

func TestStoreCreatePageWithLinks(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	anchor, err := store.CreatePage(ctx, CreatePageInput{Slug: "anchor", Title: "Anchor"})
	require.NoError(t, err)

	linked, err := store.CreatePage(ctx, CreatePageInput{Slug: "linked", Title: "Linked", LinkedPageIDs: []string{anchor.ID, anchor.ID, ""}})
	require.NoError(t, err)
	require.Equal(t, []string{anchor.ID}, linked.LinkedPageIDs)

	fetchedLinked, err := store.GetPage(ctx, linked.ID)
	require.NoError(t, err)
	require.Equal(t, []string{anchor.ID}, fetchedLinked.LinkedPageIDs)

	fetchedAnchor, err := store.GetPage(ctx, anchor.ID)
	require.NoError(t, err)
	require.Contains(t, fetchedAnchor.BacklinkedPageIDs, linked.ID)
}

func TestStoreListPages(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, err := store.CreatePage(ctx, CreatePageInput{Slug: "first", Title: "First"})
	require.NoError(t, err)
	_, err = store.CreatePage(ctx, CreatePageInput{Slug: "second", Title: "Second", ParentPageID: &first.ID})
	require.NoError(t, err)

	pages, err := store.ListPages(ctx)
	require.NoError(t, err)
	require.Len(t, pages, 2)
	require.Equal(t, "First", pages[0].Title)
	require.Equal(t, first.ID, pages[0].ID)
	require.NotNil(t, pages[1].ParentPageID)
}

func TestStoreCreateDatabaseWithPropertiesAndViews(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	icon := "📚"
	cover := "cover.png"
	db, err := store.CreateDatabase(ctx, CreateDatabaseInput{
		Slug:        "recipes",
		Title:       "Recipes",
		Description: "Family cookbook",
		Icon:        &icon,
		CoverImage:  &cover,
		Properties: []DatabasePropertyInput{
			{
				Name:       "Category",
				Slug:       "category",
				Type:       domain.PropertyTypeSelect,
				Config:     map[string]any{"options": []string{"Dinner", "Dessert"}},
				IsRequired: true,
				OrderIndex: 0,
			},
			{
				Name:       "Prep Time",
				Slug:       "prep_time",
				Type:       domain.PropertyTypeNumber,
				Config:     map[string]any{"unit": "minutes"},
				IsRequired: false,
				OrderIndex: 1,
			},
		},
		Views: []DatabaseViewInput{
			{
				Name:    "Table",
				Type:    domain.ViewTypeTable,
				Filters: map[string]any{"is_archived": false},
				Sorts:   []domain.ViewSort{{PropertyID: "prep_time", Direction: "asc"}},
				Display: []string{"category", "prep_time"},
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, db.ID)
	require.Len(t, db.Properties, 2)
	require.Len(t, db.Views, 1)

	fetched, err := store.GetDatabase(ctx, db.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, db.Slug, fetched.Slug)
	require.Len(t, fetched.Properties, 2)
	require.Equal(t, "category", fetched.Properties[0].Slug)
	require.Equal(t, domain.PropertyTypeSelect, fetched.Properties[0].Type)
	require.Len(t, fetched.Views, 1)
	require.Equal(t, domain.ViewTypeTable, fetched.Views[0].Type)
}

func TestStoreCreateDatabaseItem(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	db, err := store.CreateDatabase(ctx, CreateDatabaseInput{
		Slug:  "inventory",
		Title: "Inventory",
		Properties: []DatabasePropertyInput{
			{Name: "Name", Slug: "name", Type: domain.PropertyTypeText},
			{Name: "Quantity", Slug: "qty", Type: domain.PropertyTypeNumber},
		},
		Views: []DatabaseViewInput{{Name: "Table", Type: domain.ViewTypeTable}},
	})
	require.NoError(t, err)

	item, err := store.CreateDatabaseItem(ctx, CreateDatabaseItemInput{
		DatabaseID: db.ID,
		Page: CreatePageInput{
			Slug:    "hammer",
			Title:   "Hammer",
			Content: "Sturdy hammer",
		},
		Position: 0,
		Values: map[string]any{
			"name": "Hammer",
			"qty":  2,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, item.ID)
	require.Equal(t, "Hammer", item.Page.Title)
	require.Equal(t, 2.0, toNumber(item.PropertyMap["qty"].RawValue))

	items, err := store.ListViewItems(ctx, db.ID, db.Views[0].ID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Hammer", items[0].Page.Title)
	require.Equal(t, 2.0, toNumber(items[0].PropertyMap["qty"].RawValue))
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func toNumber(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case float32:
		return float64(n)
	default:
		return 0
	}
}

func TestOpenCreatesDirectoryForFileDSN(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "data", "app.db")
	dsn := fmt.Sprintf("file:%s?_fk=1", dbPath)

	store, err := Open(dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = store.Close()
	})

	dirInfo, err := os.Stat(filepath.Dir(dbPath))
	require.NoError(t, err)
	require.True(t, dirInfo.IsDir())

	_, err = os.Stat(dbPath)
	require.NoError(t, err)
}

func TestStoreListViewItemsViewNotFound(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	db, err := store.CreateDatabase(ctx, CreateDatabaseInput{
		Slug:       "inventory",
		Title:      "Inventory",
		Properties: []DatabasePropertyInput{{Name: "Name", Slug: "name", Type: domain.PropertyTypeText}},
		Views:      []DatabaseViewInput{{Name: "Table", Type: domain.ViewTypeTable}},
	})
	require.NoError(t, err)

	_, err = store.ListViewItems(ctx, db.ID, "missing")
	require.ErrorIs(t, err, ErrViewNotFound)
}

func TestStoreCreatePageValidation(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.CreatePage(ctx, CreatePageInput{Title: "No slug"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "slug and title")
}
