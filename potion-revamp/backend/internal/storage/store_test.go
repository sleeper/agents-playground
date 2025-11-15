package storage_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"

    "github.com/example/potion-revamp/backend/internal/models"
    "github.com/example/potion-revamp/backend/internal/storage"
)

func newTestStore(t *testing.T) *storage.Store {
    t.Helper()
    store, err := storage.NewStore(":memory:")
    require.NoError(t, err)
    t.Cleanup(func() {
        store.Close()
    })
    return store
}

func TestReplacePageBlocksTracksLinks(t *testing.T) {
    store := newTestStore(t)
    ctx := context.Background()

    pageA, err := store.CreatePage(ctx, "Home", nil, "")
    require.NoError(t, err)
    pageB, err := store.CreatePage(ctx, "Recipes", nil, "")
    require.NoError(t, err)

    blocks := []models.Block{
        {
            Type: models.BlockTypeMarkdown,
            Data: map[string]interface{}{"markdown": "Welcome to Potion", "linkedPageIds": []interface{}{pageB.ID}},
        },
        {
            Type: models.BlockTypePageLink,
            Data: map[string]interface{}{"targetPageId": pageB.ID, "alias": "Favorite meals"},
        },
    }

    _, err = store.ReplacePageBlocks(ctx, pageA.ID, blocks)
    require.NoError(t, err)

    // The backlinks for Recipes should include Home.
    recipes, err := store.GetPageWithBlocks(ctx, pageB.ID)
    require.NoError(t, err)
    require.Len(t, recipes.Backlinks, 1)
    require.Equal(t, pageA.ID, recipes.Backlinks[0].ID)

    // Blocks for Home should persist with generated IDs.
    home, err := store.GetPageWithBlocks(ctx, pageA.ID)
    require.NoError(t, err)
    require.Len(t, home.Blocks, 2)
    require.NotEmpty(t, home.Blocks[0].ID)
    require.Equal(t, pageA.ID, home.Blocks[0].PageID)
}

func TestDatabaseLifecycle(t *testing.T) {
    store := newTestStore(t)
    ctx := context.Background()

    dbModel := models.Database{
        Title: "Meal Planner",
        Properties: []models.DatabaseProperty{
            {Name: "Name", Type: models.PropertyTypeTitle},
            {Name: "Status", Type: models.PropertyTypeSelect, Options: map[string]interface{}{"options": []map[string]string{{"id": "todo", "name": "To Do"}, {"id": "done", "name": "Done"}}}},
        },
        Views: []models.DatabaseView{{Name: "Board", Type: models.ViewTypeKanban, Options: map[string]interface{}{"groupBy": "Status"}}},
    }

    created, err := store.CreateDatabase(ctx, dbModel)
    require.NoError(t, err)
    require.NotEmpty(t, created.ID)
    require.Len(t, created.Properties, 2)
    require.Len(t, created.Views, 1)

    // Map property names to IDs for test convenience.
    propIDs := map[string]string{}
    for _, prop := range created.Properties {
        propIDs[prop.Name] = prop.ID
    }

    entry, err := store.CreateDatabaseEntry(ctx, models.DatabaseEntry{
        DatabaseID: created.ID,
        Title:      "Chili",
        Properties: map[string]interface{}{
            propIDs["Status"]: map[string]interface{}{"id": "todo", "name": "To Do"},
        },
    })
    require.NoError(t, err)
    require.NotEmpty(t, entry.ID)

    entry.Properties[propIDs["Status"]] = map[string]interface{}{"id": "done", "name": "Done"}
    entry.Title = "Veggie Chili"
    require.NoError(t, store.UpdateDatabaseEntry(ctx, entry))

    resolved, err := store.ResolveViewEntries(ctx, created.ID, created.Views[0].ID)
    require.NoError(t, err)
    require.Equal(t, created.ID, resolved.Database.ID)
    require.Len(t, resolved.Entries, 1)
    require.Equal(t, "Veggie Chili", resolved.Entries[0].Title)
    statusVal := resolved.Entries[0].Properties[propIDs["Status"]].(map[string]interface{})
    require.Equal(t, "Done", statusVal["name"])
}
