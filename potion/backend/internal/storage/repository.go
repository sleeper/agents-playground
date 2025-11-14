package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/example/potion/internal/models"
	_ "modernc.org/sqlite"
)

// Repository provides CRUD operations for potion entities backed by SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository opens or creates the SQLite database at the given path and ensures the schema exists.
func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	repo := &Repository{db: db}
	if err := repo.ensureSchema(); err != nil {
		return nil, err
	}
	return repo, nil
}

// Close closes the underlying database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) ensureSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS pages (
            id TEXT PRIMARY KEY,
            title TEXT NOT NULL,
            content TEXT NOT NULL,
            parent_id TEXT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS page_links (
            id TEXT PRIMARY KEY,
            source_id TEXT NOT NULL,
            target_id TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL,
            FOREIGN KEY(source_id) REFERENCES pages(id) ON DELETE CASCADE,
            FOREIGN KEY(target_id) REFERENCES pages(id) ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS databases (
            id TEXT PRIMARY KEY,
            title TEXT NOT NULL,
            description TEXT NOT NULL,
            view TEXT NOT NULL,
            schema_json TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS database_entries (
            id TEXT PRIMARY KEY,
            database_id TEXT NOT NULL,
            title TEXT NOT NULL,
            properties_json TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL,
            FOREIGN KEY(database_id) REFERENCES databases(id) ON DELETE CASCADE
        );`,
	}

	for _, stmt := range stmts {
		if _, err := r.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// CreatePage stores a new page with markdown content.
func (r *Repository) CreatePage(ctx context.Context, title, content string, parentID *string) (models.Page, error) {
	now := time.Now().UTC()
	id := uuid.NewString()
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO pages (id, title, content, parent_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, title, content, parentID, now, now,
	); err != nil {
		return models.Page{}, err
	}
	return models.Page{
		ID:        id,
		Title:     title,
		Content:   content,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetPage retrieves a page by ID.
func (r *Repository) GetPage(ctx context.Context, id string) (models.Page, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, content, parent_id, created_at, updated_at FROM pages WHERE id = ?`, id,
	)
	var page models.Page
	var parentID sql.NullString
	if err := row.Scan(&page.ID, &page.Title, &page.Content, &parentID, &page.CreatedAt, &page.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Page{}, err
		}
		return models.Page{}, err
	}
	if parentID.Valid {
		page.ParentID = &parentID.String
	}
	return page, nil
}

// ListPages returns all pages ordered by update time desc.
func (r *Repository) ListPages(ctx context.Context) ([]models.Page, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, content, parent_id, created_at, updated_at FROM pages ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var page models.Page
		var parentID sql.NullString
		if err := rows.Scan(&page.ID, &page.Title, &page.Content, &parentID, &page.CreatedAt, &page.UpdatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			page.ParentID = &parentID.String
		}
		pages = append(pages, page)
	}
	return pages, rows.Err()
}

// UpdatePage updates metadata or content for a page.
func (r *Repository) UpdatePage(ctx context.Context, id, title, content string, parentID *string) (models.Page, error) {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`UPDATE pages SET title = ?, content = ?, parent_id = ?, updated_at = ? WHERE id = ?`,
		title, content, parentID, now, id,
	)
	if err != nil {
		return models.Page{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return models.Page{}, err
	}
	if affected == 0 {
		return models.Page{}, sql.ErrNoRows
	}
	page, err := r.GetPage(ctx, id)
	if err != nil {
		return models.Page{}, err
	}
	return page, nil
}

// CreatePageLink stores a reference from one page to another.
func (r *Repository) CreatePageLink(ctx context.Context, sourceID, targetID string) (models.PageLink, error) {
	now := time.Now().UTC()
	id := uuid.NewString()
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO page_links (id, source_id, target_id, created_at) VALUES (?, ?, ?, ?)`,
		id, sourceID, targetID, now,
	); err != nil {
		return models.PageLink{}, err
	}
	return models.PageLink{ID: id, SourceID: sourceID, TargetID: targetID, CreatedAt: now}, nil
}

// ListPageLinks returns all outgoing links for a page.
func (r *Repository) ListPageLinks(ctx context.Context, pageID string) ([]models.PageLink, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, source_id, target_id, created_at FROM page_links WHERE source_id = ? ORDER BY created_at DESC`,
		pageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []models.PageLink
	for rows.Next() {
		var link models.PageLink
		if err := rows.Scan(&link.ID, &link.SourceID, &link.TargetID, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// CreateDatabase stores a new database definition.
func (r *Repository) CreateDatabase(ctx context.Context, title, description, view string, schema map[string]string) (models.Database, error) {
	now := time.Now().UTC()
	id := uuid.NewString()
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return models.Database{}, err
	}
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO databases (id, title, description, view, schema_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, title, description, view, string(schemaJSON), now, now,
	); err != nil {
		return models.Database{}, err
	}
	return models.Database{
		ID:          id,
		Title:       title,
		Description: description,
		View:        view,
		Schema:      schema,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetDatabase retrieves a database definition.
func (r *Repository) GetDatabase(ctx context.Context, id string) (models.Database, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, view, schema_json, created_at, updated_at FROM databases WHERE id = ?`, id,
	)
	var dbModel models.Database
	var schemaJSON string
	if err := row.Scan(&dbModel.ID, &dbModel.Title, &dbModel.Description, &dbModel.View, &schemaJSON, &dbModel.CreatedAt, &dbModel.UpdatedAt); err != nil {
		return models.Database{}, err
	}
	if err := json.Unmarshal([]byte(schemaJSON), &dbModel.Schema); err != nil {
		return models.Database{}, err
	}
	if dbModel.Schema == nil {
		dbModel.Schema = map[string]string{}
	}
	return dbModel, nil
}

// ListDatabases returns all databases ordered by update time desc.
func (r *Repository) ListDatabases(ctx context.Context) ([]models.Database, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, description, view, schema_json, created_at, updated_at FROM databases ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []models.Database
	for rows.Next() {
		var dbModel models.Database
		var schemaJSON string
		if err := rows.Scan(&dbModel.ID, &dbModel.Title, &dbModel.Description, &dbModel.View, &schemaJSON, &dbModel.CreatedAt, &dbModel.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(schemaJSON), &dbModel.Schema); err != nil {
			return nil, err
		}
		if dbModel.Schema == nil {
			dbModel.Schema = map[string]string{}
		}
		dbs = append(dbs, dbModel)
	}
	return dbs, rows.Err()
}

// CreateDatabaseEntry stores a new entry for a database.
func (r *Repository) CreateDatabaseEntry(ctx context.Context, databaseID, title string, properties map[string]interface{}) (models.DatabaseEntry, error) {
	now := time.Now().UTC()
	id := uuid.NewString()
	propsJSON, err := json.Marshal(properties)
	if err != nil {
		return models.DatabaseEntry{}, err
	}
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO database_entries (id, database_id, title, properties_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, databaseID, title, string(propsJSON), now, now,
	); err != nil {
		return models.DatabaseEntry{}, err
	}
	return models.DatabaseEntry{
		ID:         id,
		DatabaseID: databaseID,
		Title:      title,
		Properties: properties,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// ListDatabaseEntries returns all entries for a given database ordered by creation time desc.
func (r *Repository) ListDatabaseEntries(ctx context.Context, databaseID string) ([]models.DatabaseEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, database_id, title, properties_json, created_at, updated_at FROM database_entries WHERE database_id = ? ORDER BY created_at DESC`,
		databaseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.DatabaseEntry
	for rows.Next() {
		var entry models.DatabaseEntry
		var propsJSON string
		if err := rows.Scan(&entry.ID, &entry.DatabaseID, &entry.Title, &propsJSON, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(propsJSON), &entry.Properties); err != nil {
			return nil, err
		}
		if entry.Properties == nil {
			entry.Properties = map[string]interface{}{}
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
