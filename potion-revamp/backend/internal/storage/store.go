package storage

import (
    "context"
    "database/sql"
    "embed"
    "encoding/json"
    "errors"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/google/uuid"
    _ "modernc.org/sqlite"

    "github.com/example/potion-revamp/backend/internal/models"
)

//go:embed schema.sql
var schemaFS embed.FS

// Store wraps persistence helpers for Potion.
type Store struct {
    db *sql.DB
}

// NewStore opens the SQLite database at path and ensures schema.
func NewStore(path string) (*Store, error) {
    dsn := path
    if !strings.Contains(path, "=") {
        dsn = fmt.Sprintf("file:%s?_foreign_keys=on", path)
    }

    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(1)

    if err := applySchema(db); err != nil {
        db.Close()
        return nil, err
    }

    return &Store{db: db}, nil
}

func applySchema(db *sql.DB) error {
    schema, err := schemaFS.ReadFile("schema.sql")
    if err != nil {
        return err
    }
    _, err = db.Exec(string(schema))
    return err
}

// Close shuts down the underlying database connection.
func (s *Store) Close() error {
    return s.db.Close()
}

// CreatePage inserts a new page with optional parent.
func (s *Store) CreatePage(ctx context.Context, title string, parentID *string, icon string) (models.Page, error) {
    now := time.Now().UTC()
    id := uuid.NewString()
    _, err := s.db.ExecContext(ctx, `INSERT INTO pages(id, title, icon, parent_id, created_at, updated_at) VALUES(?,?,?,?,?,?)`,
        id, title, nullIfEmpty(icon), parentID, now, now)
    if err != nil {
        return models.Page{}, err
    }
    return models.Page{ID: id, Title: title, Icon: icon, ParentID: parentID, CreatedAt: now, UpdatedAt: now}, nil
}

// UpdatePage modifies page metadata.
func (s *Store) UpdatePage(ctx context.Context, page models.Page) error {
    page.UpdatedAt = time.Now().UTC()
    res, err := s.db.ExecContext(ctx, `UPDATE pages SET title = ?, icon = ?, parent_id = ?, updated_at = ? WHERE id = ?`,
        page.Title, nullIfEmpty(page.Icon), page.ParentID, page.UpdatedAt, page.ID)
    if err != nil {
        return err
    }
    rows, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

// DeletePage removes a page and cascading content.
func (s *Store) DeletePage(ctx context.Context, id string) error {
    _, err := s.db.ExecContext(ctx, `DELETE FROM pages WHERE id = ?`, id)
    return err
}

// GetPageWithBlocks returns page details, ordered blocks, and backlinks.
func (s *Store) GetPageWithBlocks(ctx context.Context, id string) (models.PageWithBlocks, error) {
    page, err := s.getPage(ctx, id)
    if err != nil {
        return models.PageWithBlocks{}, err
    }
    blocks, err := s.listBlocks(ctx, id)
    if err != nil {
        return models.PageWithBlocks{}, err
    }
    backlinks, err := s.listBacklinks(ctx, id)
    if err != nil {
        return models.PageWithBlocks{}, err
    }
    return models.PageWithBlocks{Page: page, Blocks: blocks, Backlinks: backlinks}, nil
}

// ListPages returns all pages ordered by updated timestamp.
func (s *Store) ListPages(ctx context.Context) ([]models.Page, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT id, title, icon, parent_id, created_at, updated_at FROM pages ORDER BY updated_at DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var pages []models.Page
    for rows.Next() {
        var p models.Page
        var parent sql.NullString
        var icon sql.NullString
        if err := rows.Scan(&p.ID, &p.Title, &icon, &parent, &p.CreatedAt, &p.UpdatedAt); err != nil {
            return nil, err
        }
        if icon.Valid {
            p.Icon = icon.String
        }
        if parent.Valid {
            val := parent.String
            p.ParentID = &val
        }
        pages = append(pages, p)
    }
    if pages == nil {
        pages = []models.Page{}
    }
    return pages, rows.Err()
}

func (s *Store) getPage(ctx context.Context, id string) (models.Page, error) {
    var p models.Page
    var parent sql.NullString
    var icon sql.NullString
    err := s.db.QueryRowContext(ctx, `SELECT id, title, icon, parent_id, created_at, updated_at FROM pages WHERE id = ?`, id).
        Scan(&p.ID, &p.Title, &icon, &parent, &p.CreatedAt, &p.UpdatedAt)
    if err != nil {
        return models.Page{}, err
    }
    if icon.Valid {
        p.Icon = icon.String
    }
    if parent.Valid {
        val := parent.String
        p.ParentID = &val
    }
    return p, nil
}

func (s *Store) listBlocks(ctx context.Context, pageID string) ([]models.Block, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT id, position, type, data FROM blocks WHERE page_id = ? ORDER BY position ASC`, pageID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var blocks []models.Block
    for rows.Next() {
        var blk models.Block
        var raw string
        if err := rows.Scan(&blk.ID, &blk.Position, &blk.Type, &raw); err != nil {
            return nil, err
        }
        blk.PageID = pageID
        if err := json.Unmarshal([]byte(raw), &blk.Data); err != nil {
            return nil, err
        }
        blocks = append(blocks, blk)
    }
    if blocks == nil {
        blocks = []models.Block{}
    }
    return blocks, rows.Err()
}

func (s *Store) listBacklinks(ctx context.Context, pageID string) ([]models.Page, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT p.id, p.title, p.icon, p.parent_id, p.created_at, p.updated_at FROM page_links l INNER JOIN pages p ON p.id = l.source_page_id WHERE l.target_page_id = ? ORDER BY p.updated_at DESC`, pageID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var pages []models.Page
    for rows.Next() {
        var p models.Page
        var parent sql.NullString
        var icon sql.NullString
        if err := rows.Scan(&p.ID, &p.Title, &icon, &parent, &p.CreatedAt, &p.UpdatedAt); err != nil {
            return nil, err
        }
        if icon.Valid {
            p.Icon = icon.String
        }
        if parent.Valid {
            val := parent.String
            p.ParentID = &val
        }
        pages = append(pages, p)
    }
    if pages == nil {
        pages = []models.Page{}
    }
    return pages, rows.Err()
}

// ReplacePageBlocks overwrites blocks for a page and refreshes link index.
func (s *Store) ReplacePageBlocks(ctx context.Context, pageID string, blocks []models.Block) ([]models.Block, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    committed := false
    defer func() {
        if !committed {
            tx.Rollback()
        }
    }()

    if _, err = tx.ExecContext(ctx, `DELETE FROM page_links WHERE source_page_id = ?`, pageID); err != nil {
        return nil, err
    }
    if _, err = tx.ExecContext(ctx, `DELETE FROM embedded_database_views WHERE block_id IN (SELECT id FROM blocks WHERE page_id = ?)`, pageID); err != nil {
        return nil, err
    }
    if _, err = tx.ExecContext(ctx, `DELETE FROM blocks WHERE page_id = ?`, pageID); err != nil {
        return nil, err
    }

    persisted := make([]models.Block, 0, len(blocks))

    for idx, blk := range blocks {
        if blk.ID == "" {
            blk.ID = uuid.NewString()
        }
        blk.PageID = pageID
        blk.Position = idx
        now := time.Now().UTC()
        payload, marshalErr := json.Marshal(blk.Data)
        if marshalErr != nil {
            err = marshalErr
            return nil, err
        }
        if _, err = tx.ExecContext(ctx, `INSERT INTO blocks(id, page_id, position, type, data, created_at, updated_at) VALUES(?,?,?,?,?,?,?)`,
            blk.ID, pageID, blk.Position, string(blk.Type), string(payload), now, now); err != nil {
            return nil, err
        }
        if blk.Type == models.BlockTypeDatabaseRef {
            if viewID, ok := blk.Data["viewId"].(string); ok && viewID != "" {
                if _, err = tx.ExecContext(ctx, `INSERT INTO embedded_database_views(block_id, view_id) VALUES(?,?)`, blk.ID, viewID); err != nil {
                    return nil, err
                }
            }
        }
        linkIDs := extractLinkedPages(blk)
        for _, target := range linkIDs {
            if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO page_links(source_page_id, target_page_id) VALUES(?,?)`, pageID, target); err != nil {
                return nil, err
            }
        }
        persisted = append(persisted, blk)
    }

    if err = tx.Commit(); err != nil {
        return nil, err
    }
    committed = true
    return persisted, nil
}

func extractLinkedPages(blk models.Block) []string {
    var results []string
    if blk.Type == models.BlockTypePageLink {
        if id, ok := blk.Data["targetPageId"].(string); ok && id != "" {
            results = append(results, id)
        }
    }
    if linksRaw, ok := blk.Data["linkedPageIds"].([]interface{}); ok {
        for _, raw := range linksRaw {
            if id, ok := raw.(string); ok && id != "" {
                results = append(results, id)
            }
        }
    } else if linksStr, ok := blk.Data["linkedPageIds"].([]string); ok {
        results = append(results, linksStr...)
    }
    return results
}

// ListDatabases returns stored databases with metadata only.
func (s *Store) ListDatabases(ctx context.Context) ([]models.Database, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT id, title, description, icon, created_at, updated_at FROM databases ORDER BY updated_at DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var items []models.Database
    for rows.Next() {
        var dbRow models.Database
        var description sql.NullString
        var icon sql.NullString
        if err := rows.Scan(&dbRow.ID, &dbRow.Title, &description, &icon, &dbRow.CreatedAt, &dbRow.UpdatedAt); err != nil {
            return nil, err
        }
        if description.Valid {
            dbRow.Description = description.String
        }
        if icon.Valid {
            dbRow.Icon = icon.String
        }
        items = append(items, dbRow)
    }
    if items == nil {
        items = []models.Database{}
    }
    return items, rows.Err()
}

// CreateDatabase creates database with properties and optional views.
func (s *Store) CreateDatabase(ctx context.Context, payload models.Database) (models.Database, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return models.Database{}, err
    }
    committed := false
    defer func() {
        if !committed {
            tx.Rollback()
        }
    }()

    now := time.Now().UTC()
    payload.ID = uuid.NewString()
    payload.CreatedAt = now
    payload.UpdatedAt = now

    if _, err = tx.ExecContext(ctx, `INSERT INTO databases(id, title, description, icon, created_at, updated_at) VALUES(?,?,?,?,?,?)`,
        payload.ID, payload.Title, nullIfEmpty(payload.Description), nullIfEmpty(payload.Icon), payload.CreatedAt, payload.UpdatedAt); err != nil {
        return models.Database{}, err
    }

    for i := range payload.Properties {
        prop := &payload.Properties[i]
        if prop.ID == "" {
            prop.ID = uuid.NewString()
        }
        prop.DatabaseID = payload.ID
        prop.Position = i
        optionsJSON, marshalErr := json.Marshal(prop.Options)
        if marshalErr != nil {
            return models.Database{}, marshalErr
        }
        if _, err = tx.ExecContext(ctx, `INSERT INTO database_properties(id, database_id, name, type, options, position) VALUES(?,?,?,?,?,?)`,
            prop.ID, payload.ID, prop.Name, string(prop.Type), string(optionsJSON), prop.Position); err != nil {
            return models.Database{}, err
        }
    }

    for i := range payload.Views {
        view := &payload.Views[i]
        if view.ID == "" {
            view.ID = uuid.NewString()
        }
        view.DatabaseID = payload.ID
        view.Position = i
        optionsJSON, marshalErr := json.Marshal(view.Options)
        if marshalErr != nil {
            return models.Database{}, marshalErr
        }
        if _, err = tx.ExecContext(ctx, `INSERT INTO database_views(id, database_id, name, type, options, position) VALUES(?,?,?,?,?,?)`,
            view.ID, payload.ID, view.Name, string(view.Type), string(optionsJSON), view.Position); err != nil {
            return models.Database{}, err
        }
    }

    if err = tx.Commit(); err != nil {
        return models.Database{}, err
    }
    committed = true
    return payload, nil
}

// GetDatabase retrieves database metadata with properties and views.
func (s *Store) GetDatabase(ctx context.Context, id string) (models.Database, error) {
    var dbRow models.Database
    var description sql.NullString
    var icon sql.NullString
    err := s.db.QueryRowContext(ctx, `SELECT id, title, description, icon, created_at, updated_at FROM databases WHERE id = ?`, id).
        Scan(&dbRow.ID, &dbRow.Title, &description, &icon, &dbRow.CreatedAt, &dbRow.UpdatedAt)
    if err != nil {
        return models.Database{}, err
    }
    if description.Valid {
        dbRow.Description = description.String
    }
    if icon.Valid {
        dbRow.Icon = icon.String
    }
    props, err := s.listDatabaseProperties(ctx, id)
    if err != nil {
        return models.Database{}, err
    }
    views, err := s.listDatabaseViews(ctx, id)
    if err != nil {
        return models.Database{}, err
    }
    dbRow.Properties = props
    dbRow.Views = views
    return dbRow, nil
}

func (s *Store) listDatabaseProperties(ctx context.Context, dbID string) ([]models.DatabaseProperty, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT id, name, type, options, position FROM database_properties WHERE database_id = ? ORDER BY position ASC`, dbID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var props []models.DatabaseProperty
    for rows.Next() {
        var prop models.DatabaseProperty
        var options sql.NullString
        if err := rows.Scan(&prop.ID, &prop.Name, &prop.Type, &options, &prop.Position); err != nil {
            return nil, err
        }
        prop.DatabaseID = dbID
        if options.Valid && options.String != "" {
            if err := json.Unmarshal([]byte(options.String), &prop.Options); err != nil {
                return nil, err
            }
        } else {
            prop.Options = map[string]interface{}{}
        }
        props = append(props, prop)
    }
    if props == nil {
        props = []models.DatabaseProperty{}
    }
    return props, rows.Err()
}

func (s *Store) listDatabaseViews(ctx context.Context, dbID string) ([]models.DatabaseView, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT id, name, type, options, position FROM database_views WHERE database_id = ? ORDER BY position ASC`, dbID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var views []models.DatabaseView
    for rows.Next() {
        var view models.DatabaseView
        var options sql.NullString
        if err := rows.Scan(&view.ID, &view.Name, &view.Type, &options, &view.Position); err != nil {
            return nil, err
        }
        view.DatabaseID = dbID
        if options.Valid && options.String != "" {
            if err := json.Unmarshal([]byte(options.String), &view.Options); err != nil {
                return nil, err
            }
        } else {
            view.Options = map[string]interface{}{}
        }
        views = append(views, view)
    }
    if views == nil {
        views = []models.DatabaseView{}
    }
    return views, rows.Err()
}

// CreateDatabaseEntry inserts new entry with property values.
func (s *Store) CreateDatabaseEntry(ctx context.Context, entry models.DatabaseEntry) (models.DatabaseEntry, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return models.DatabaseEntry{}, err
    }
    committed := false
    defer func() {
        if !committed {
            tx.Rollback()
        }
    }()
    entry.ID = uuid.NewString()
    entry.CreatedAt = time.Now().UTC()
    entry.UpdatedAt = entry.CreatedAt

    if entry.Properties == nil {
        entry.Properties = map[string]interface{}{}
    }

    if _, err = tx.ExecContext(ctx, `INSERT INTO database_entries(id, database_id, title, created_at, updated_at) VALUES(?,?,?,?,?)`,
        entry.ID, entry.DatabaseID, entry.Title, entry.CreatedAt, entry.UpdatedAt); err != nil {
        return models.DatabaseEntry{}, err
    }

    for propID, value := range entry.Properties {
        optionsJSON, marshalErr := json.Marshal(value)
        if marshalErr != nil {
            return models.DatabaseEntry{}, marshalErr
        }
        if _, err = tx.ExecContext(ctx, `INSERT INTO database_entry_properties(entry_id, property_id, value) VALUES(?,?,?)`, entry.ID, propID, string(optionsJSON)); err != nil {
            return models.DatabaseEntry{}, err
        }
    }

    if err = tx.Commit(); err != nil {
        return models.DatabaseEntry{}, err
    }
    committed = true
    return entry, nil
}

// UpdateDatabaseEntry updates an entry and its properties.
func (s *Store) UpdateDatabaseEntry(ctx context.Context, entry models.DatabaseEntry) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    committed := false
    defer func() {
        if !committed {
            tx.Rollback()
        }
    }()

    entry.UpdatedAt = time.Now().UTC()
    if _, err = tx.ExecContext(ctx, `UPDATE database_entries SET title = ?, updated_at = ? WHERE id = ?`, entry.Title, entry.UpdatedAt, entry.ID); err != nil {
        return err
    }
    if _, err = tx.ExecContext(ctx, `DELETE FROM database_entry_properties WHERE entry_id = ?`, entry.ID); err != nil {
        return err
    }
    for propID, value := range entry.Properties {
        optionsJSON, marshalErr := json.Marshal(value)
        if marshalErr != nil {
            return marshalErr
        }
        if _, err = tx.ExecContext(ctx, `INSERT INTO database_entry_properties(entry_id, property_id, value) VALUES(?,?,?)`, entry.ID, propID, string(optionsJSON)); err != nil {
            return err
        }
    }
    if err = tx.Commit(); err != nil {
        return err
    }
    committed = true
    return nil
}

// ListDatabaseEntries returns entries along with property payloads.
func (s *Store) ListDatabaseEntries(ctx context.Context, databaseID string) ([]models.DatabaseEntry, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT id, title, created_at, updated_at FROM database_entries WHERE database_id = ? ORDER BY updated_at DESC`, databaseID)
    if err != nil {
        return nil, err
    }
    entries := []models.DatabaseEntry{}
    for rows.Next() {
        var entry models.DatabaseEntry
        if err := rows.Scan(&entry.ID, &entry.Title, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
            rows.Close()
            return nil, err
        }
        entry.DatabaseID = databaseID
        entries = append(entries, entry)
    }
    if err := rows.Close(); err != nil {
        return nil, err
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    for i := range entries {
        props, propErr := s.fetchEntryProperties(ctx, entries[i].ID)
        if propErr != nil {
            return nil, propErr
        }
        entries[i].Properties = props
    }
    return entries, nil
}

func (s *Store) fetchEntryProperties(ctx context.Context, entryID string) (map[string]interface{}, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT property_id, value FROM database_entry_properties WHERE entry_id = ?`, entryID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    props := map[string]interface{}{}
    for rows.Next() {
        var propID string
        var raw string
        if err := rows.Scan(&propID, &raw); err != nil {
            return nil, err
        }
        var val interface{}
        if err := json.Unmarshal([]byte(raw), &val); err != nil {
            return nil, err
        }
        props[propID] = val
    }
    return props, rows.Err()
}

// GetEmbeddedViewMapping returns the embedded view for a block.
func (s *Store) GetEmbeddedViewMapping(ctx context.Context, blockID string) (string, error) {
    var viewID string
    err := s.db.QueryRowContext(ctx, `SELECT view_id FROM embedded_database_views WHERE block_id = ?`, blockID).Scan(&viewID)
    if err != nil {
        return "", err
    }
    return viewID, nil
}

// UpsertDatabaseView updates metadata for an existing view.
func (s *Store) UpsertDatabaseView(ctx context.Context, view models.DatabaseView) error {
    payload, err := json.Marshal(view.Options)
    if err != nil {
        return err
    }
    _, err = s.db.ExecContext(ctx, `INSERT INTO database_views(id, database_id, name, type, options, position) VALUES(?,?,?,?,?,?)
        ON CONFLICT(id) DO UPDATE SET name = excluded.name, type = excluded.type, options = excluded.options, position = excluded.position`,
        view.ID, view.DatabaseID, view.Name, string(view.Type), string(payload), view.Position)
    return err
}

// ResolveViewEntries returns entries filtered and sorted for a view.
func (s *Store) ResolveViewEntries(ctx context.Context, databaseID, viewID string) (models.DatabaseWithEntries, error) {
    dbMeta, err := s.GetDatabase(ctx, databaseID)
    if err != nil {
        return models.DatabaseWithEntries{}, err
    }
    var view models.DatabaseView
    found := false
    for _, v := range dbMeta.Views {
        if v.ID == viewID {
            view = v
            found = true
            break
        }
    }
    if !found {
        return models.DatabaseWithEntries{}, errors.New("view not found")
    }
    entries, err := s.ListDatabaseEntries(ctx, databaseID)
    if err != nil {
        return models.DatabaseWithEntries{}, err
    }
    // Provide lightweight post-processing for kanban grouping order.
    if view.Type == models.ViewTypeKanban {
        groupProp, _ := view.Options["groupBy"].(string)
        if groupProp != "" {
            sortEntriesBySelect(entries, groupProp)
        }
    }
    if view.Type == models.ViewTypeGallery {
        coverProp, _ := view.Options["coverProperty"].(string)
        ensureCoverDefaults(entries, coverProp)
    }
    return models.DatabaseWithEntries{Database: dbMeta, Entries: entries}, nil
}

func sortEntriesBySelect(entries []models.DatabaseEntry, propID string) {
    sort.SliceStable(entries, func(i, j int) bool {
        vi, _ := entries[i].Properties[propID].(map[string]interface{})
        vj, _ := entries[j].Properties[propID].(map[string]interface{})
        ti := ""
        tj := ""
        if vi != nil {
            if name, ok := vi["name"].(string); ok {
                ti = name
            }
        }
        if vj != nil {
            if name, ok := vj["name"].(string); ok {
                tj = name
            }
        }
        return ti < tj
    })
}

func ensureCoverDefaults(entries []models.DatabaseEntry, propID string) {
    if propID == "" {
        return
    }
    for idx := range entries {
        if _, ok := entries[idx].Properties[propID]; !ok {
            entries[idx].Properties[propID] = map[string]interface{}{"name": ""}
        }
    }
}

func nullIfEmpty(value string) interface{} {
    if strings.TrimSpace(value) == "" {
        return nil
    }
    return value
}
