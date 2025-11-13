package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/example/agents-playground/internal/domain"
	"github.com/google/uuid"
)

// ErrViewNotFound is returned when a database view cannot be located for an item listing.
var ErrViewNotFound = errors.New("view not found")

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store wraps access to the SQLite database.
type Store struct {
	db *sql.DB
}

// Open initializes a SQLite store at the provided DSN.
func Open(dsn string) (*Store, error) {
	if strings.TrimSpace(dsn) == "" {
		dsn = ":memory:"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes underlying db.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Ping verifies database connectivity.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func applyMigrations(db *sql.DB) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		sqlBytes, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// CreatePageInput holds fields for a new page.
type CreatePageInput struct {
	Slug         string
	Title        string
	Summary      string
	Content      string
	ParentPageID *string
	Tags         []string
}

// CreatePage persists a new page.
func (s *Store) CreatePage(ctx context.Context, in CreatePageInput) (*domain.Page, error) {
	if in.Slug == "" || in.Title == "" {
		return nil, errors.New("slug and title are required")
	}
	now := time.Now().UTC()
	id := uuid.NewString()
	tagJSON, err := json.Marshal(in.Tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO pages(
            id, slug, title, summary, content, parent_page_id, tags, is_archived, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
    `, id, in.Slug, in.Title, in.Summary, in.Content, in.ParentPageID, string(tagJSON), now, now)
	if err != nil {
		return nil, fmt.Errorf("insert page: %w", err)
	}
	return &domain.Page{
		ID:           id,
		Slug:         in.Slug,
		Title:        in.Title,
		Summary:      in.Summary,
		Content:      in.Content,
		ParentPageID: in.ParentPageID,
		Tags:         in.Tags,
		IsArchived:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetPage retrieves a page by id.
func (s *Store) GetPage(ctx context.Context, id string) (*domain.Page, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, slug, title, summary, content, parent_page_id, cover_image_id, icon, tags, is_archived, created_at, updated_at FROM pages WHERE id = ?`, id)
	var page domain.Page
	var tags string
	var parent sql.NullString
	var cover sql.NullString
	var icon sql.NullString
	if err := row.Scan(&page.ID, &page.Slug, &page.Title, &page.Summary, &page.Content, &parent, &cover, &icon, &tags, &page.IsArchived, &page.CreatedAt, &page.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan page: %w", err)
	}
	if parent.Valid {
		page.ParentPageID = &parent.String
	}
	if cover.Valid {
		page.CoverImageID = &cover.String
	}
	if icon.Valid {
		page.Icon = &icon.String
	}
	if tags != "" {
		if err := json.Unmarshal([]byte(tags), &page.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags: %w", err)
		}
	}
	return &page, nil
}

// CreateDatabaseInput defines payload for new database.
type CreateDatabaseInput struct {
	Slug        string
	Title       string
	Description string
	Icon        *string
	CoverImage  *string
	Properties  []DatabasePropertyInput
	Views       []DatabaseViewInput
}

// DatabasePropertyInput describes a property definition.
type DatabasePropertyInput struct {
	Name       string
	Slug       string
	Type       domain.PropertyType
	Config     map[string]any
	IsRequired bool
	Default    any
	OrderIndex int
}

// DatabaseViewInput describes saved view configuration.
type DatabaseViewInput struct {
	Name          string
	Type          domain.ViewType
	Filters       map[string]any
	Sorts         []domain.ViewSort
	Grouping      map[string]any
	Display       []string
	LayoutOptions map[string]any
}

// CreateDatabase persists a database with properties and views.
func (s *Store) CreateDatabase(ctx context.Context, in CreateDatabaseInput) (*domain.Database, error) {
	if in.Slug == "" || in.Title == "" {
		return nil, errors.New("slug and title are required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	now := time.Now().UTC()
	dbID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `INSERT INTO databases(id, slug, title, description, icon, cover_image_id, is_archived, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		dbID, in.Slug, in.Title, in.Description, in.Icon, in.CoverImage, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert database: %w", err)
	}
	props := make([]domain.DatabaseProperty, 0, len(in.Properties))
	for _, propInput := range in.Properties {
		propID := uuid.NewString()
		cfg, err := json.Marshal(propInput.Config)
		if err != nil {
			return nil, fmt.Errorf("marshal property config: %w", err)
		}
		defVal, err := json.Marshal(propInput.Default)
		if err != nil {
			return nil, fmt.Errorf("marshal property default: %w", err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO database_properties(id, database_id, name, slug, type, config, is_required, default_value, order_index, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			propID, dbID, propInput.Name, propInput.Slug, string(propInput.Type), string(cfg), boolToInt(propInput.IsRequired), string(defVal), propInput.OrderIndex, now, now)
		if err != nil {
			return nil, fmt.Errorf("insert property: %w", err)
		}
		props = append(props, domain.DatabaseProperty{
			ID:         propID,
			DatabaseID: dbID,
			Name:       propInput.Name,
			Slug:       propInput.Slug,
			Type:       propInput.Type,
			Config:     propInput.Config,
			IsRequired: propInput.IsRequired,
			Default:    propInput.Default,
			OrderIndex: propInput.OrderIndex,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	views := make([]domain.DatabaseView, 0, len(in.Views))
	for _, viewInput := range in.Views {
		viewID := uuid.NewString()
		filters, err := json.Marshal(viewInput.Filters)
		if err != nil {
			return nil, fmt.Errorf("marshal view filters: %w", err)
		}
		sorts, err := json.Marshal(viewInput.Sorts)
		if err != nil {
			return nil, fmt.Errorf("marshal view sorts: %w", err)
		}
		grouping, err := json.Marshal(viewInput.Grouping)
		if err != nil {
			return nil, fmt.Errorf("marshal view grouping: %w", err)
		}
		display, err := json.Marshal(viewInput.Display)
		if err != nil {
			return nil, fmt.Errorf("marshal view display: %w", err)
		}
		layout, err := json.Marshal(viewInput.LayoutOptions)
		if err != nil {
			return nil, fmt.Errorf("marshal view layout: %w", err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO database_views(id, database_id, name, type, filters, sorts, grouping, display_properties, layout_options, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			viewID, dbID, viewInput.Name, string(viewInput.Type), string(filters), string(sorts), string(grouping), string(display), string(layout), now, now)
		if err != nil {
			return nil, fmt.Errorf("insert view: %w", err)
		}
		views = append(views, domain.DatabaseView{
			ID:            viewID,
			DatabaseID:    dbID,
			Name:          viewInput.Name,
			Type:          viewInput.Type,
			Filters:       viewInput.Filters,
			Sorts:         viewInput.Sorts,
			Grouping:      viewInput.Grouping,
			Display:       viewInput.Display,
			LayoutOptions: viewInput.LayoutOptions,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit db: %w", err)
	}
	return &domain.Database{
		ID:          dbID,
		Slug:        in.Slug,
		Title:       in.Title,
		Description: in.Description,
		Icon:        in.Icon,
		CoverImage:  in.CoverImage,
		IsArchived:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
		Properties:  props,
		Views:       views,
	}, nil
}

// GetDatabase fetches a database and eager loads properties and views.
func (s *Store) GetDatabase(ctx context.Context, id string) (*domain.Database, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, slug, title, description, icon, cover_image_id, is_archived, created_at, updated_at FROM databases WHERE id = ?`, id)
	var dbModel domain.Database
	var icon sql.NullString
	var cover sql.NullString
	if err := row.Scan(&dbModel.ID, &dbModel.Slug, &dbModel.Title, &dbModel.Description, &icon, &cover, &dbModel.IsArchived, &dbModel.CreatedAt, &dbModel.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan database: %w", err)
	}
	if icon.Valid {
		dbModel.Icon = &icon.String
	}
	if cover.Valid {
		dbModel.CoverImage = &cover.String
	}
	propsRows, err := s.db.QueryContext(ctx, `SELECT id, name, slug, type, config, is_required, default_value, order_index, created_at, updated_at FROM database_properties WHERE database_id = ? ORDER BY order_index ASC`, dbModel.ID)
	if err != nil {
		return nil, fmt.Errorf("query properties: %w", err)
	}
	defer propsRows.Close()
	for propsRows.Next() {
		var prop domain.DatabaseProperty
		var cfg, def string
		var isReq int
		if err := propsRows.Scan(&prop.ID, &prop.Name, &prop.Slug, &prop.Type, &cfg, &isReq, &def, &prop.OrderIndex, &prop.CreatedAt, &prop.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan property: %w", err)
		}
		prop.DatabaseID = dbModel.ID
		if cfg != "" {
			_ = json.Unmarshal([]byte(cfg), &prop.Config)
		}
		if def != "" {
			var raw any
			if err := json.Unmarshal([]byte(def), &raw); err == nil {
				prop.Default = raw
			}
		}
		prop.IsRequired = isReq == 1
		dbModel.Properties = append(dbModel.Properties, prop)
	}
	viewRows, err := s.db.QueryContext(ctx, `SELECT id, name, type, filters, sorts, grouping, display_properties, layout_options, created_at, updated_at FROM database_views WHERE database_id = ? ORDER BY created_at ASC`, dbModel.ID)
	if err != nil {
		return nil, fmt.Errorf("query views: %w", err)
	}
	defer viewRows.Close()
	for viewRows.Next() {
		var view domain.DatabaseView
		var filters, sorts, grouping, display, layout string
		if err := viewRows.Scan(&view.ID, &view.Name, &view.Type, &filters, &sorts, &grouping, &display, &layout, &view.CreatedAt, &view.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan view: %w", err)
		}
		view.DatabaseID = dbModel.ID
		if filters != "" {
			_ = json.Unmarshal([]byte(filters), &view.Filters)
		}
		if sorts != "" {
			_ = json.Unmarshal([]byte(sorts), &view.Sorts)
		}
		if grouping != "" {
			_ = json.Unmarshal([]byte(grouping), &view.Grouping)
		}
		if display != "" {
			_ = json.Unmarshal([]byte(display), &view.Display)
		}
		if layout != "" {
			_ = json.Unmarshal([]byte(layout), &view.LayoutOptions)
		}
		dbModel.Views = append(dbModel.Views, view)
	}
	return &dbModel, nil
}

// CreateDatabaseItemInput describes payload for items.
type CreateDatabaseItemInput struct {
	DatabaseID string
	Page       CreatePageInput
	Position   int
	Values     map[string]any // keyed by property slug
}

// CreateDatabaseItem persists a new item and associated page/values.
func (s *Store) CreateDatabaseItem(ctx context.Context, in CreateDatabaseItemInput) (*domain.DatabaseItem, error) {
	if in.DatabaseID == "" {
		return nil, errors.New("database id required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	now := time.Now().UTC()
	pageID := uuid.NewString()
	tagJSON, err := json.Marshal(in.Page.Tags)
	if err != nil {
		return nil, fmt.Errorf("marshal page tags: %w", err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO pages(id, slug, title, summary, content, parent_page_id, tags, is_archived, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		pageID, in.Page.Slug, in.Page.Title, in.Page.Summary, in.Page.Content, in.Page.ParentPageID, string(tagJSON), now, now)
	if err != nil {
		return nil, fmt.Errorf("insert item page: %w", err)
	}
	itemID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `INSERT INTO database_items(id, database_id, page_id, position, is_archived, created_at, updated_at) VALUES(?, ?, ?, ?, 0, ?, ?)`,
		itemID, in.DatabaseID, pageID, in.Position, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert database item: %w", err)
	}
	// load property map to map slug->id
	propRows, err := tx.QueryContext(ctx, `SELECT id, slug FROM database_properties WHERE database_id = ?`, in.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("query properties: %w", err)
	}
	propMap := make(map[string]string)
	for propRows.Next() {
		var id, slug string
		if err := propRows.Scan(&id, &slug); err != nil {
			propRows.Close()
			return nil, fmt.Errorf("scan property: %w", err)
		}
		propMap[slug] = id
	}
	propRows.Close()
	storedValues := make(map[string]domain.DatabaseValue)
	for slug, value := range in.Values {
		propID, ok := propMap[slug]
		if !ok {
			return nil, fmt.Errorf("unknown property slug %s", slug)
		}
		valueID := uuid.NewString()
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshal value %s: %w", slug, err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO database_values(id, database_item_id, property_id, value, is_computed, created_at, updated_at) VALUES(?, ?, ?, ?, 0, ?, ?)`,
			valueID, itemID, propID, string(raw), now, now)
		if err != nil {
			return nil, fmt.Errorf("insert value %s: %w", slug, err)
		}
		storedValues[slug] = domain.DatabaseValue{
			ID:         valueID,
			ItemID:     itemID,
			PropertyID: propID,
			RawValue:   value,
			IsComputed: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit item: %w", err)
	}
	return &domain.DatabaseItem{
		ID:         itemID,
		DatabaseID: in.DatabaseID,
		Page: domain.Page{
			ID:        pageID,
			Slug:      in.Page.Slug,
			Title:     in.Page.Title,
			Summary:   in.Page.Summary,
			Content:   in.Page.Content,
			Tags:      in.Page.Tags,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Position:    in.Position,
		IsArchived:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
		PropertyMap: storedValues,
	}, nil
}

// ListViewItems fetches items rendered for a view (basic filters).
func (s *Store) ListViewItems(ctx context.Context, databaseID, viewID string) ([]domain.DatabaseItem, error) {
	if databaseID == "" || viewID == "" {
		return nil, errors.New("database id and view id required")
	}
	var ok int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM database_views WHERE id = ? AND database_id = ?`, viewID, databaseID).Scan(&ok)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrViewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("verify view: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT di.id, di.page_id, di.position, di.is_archived, di.created_at, di.updated_at, p.slug, p.title, p.summary, p.content, p.tags FROM database_items di JOIN pages p ON di.page_id = p.id WHERE di.database_id = ? ORDER BY di.position ASC, di.created_at ASC`, databaseID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()
	var items []domain.DatabaseItem
	for rows.Next() {
		var item domain.DatabaseItem
		var page domain.Page
		var tags string
		if err := rows.Scan(&item.ID, &page.ID, &item.Position, &item.IsArchived, &item.CreatedAt, &item.UpdatedAt, &page.Slug, &page.Title, &page.Summary, &page.Content, &tags); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		if tags != "" {
			_ = json.Unmarshal([]byte(tags), &page.Tags)
		}
		page.CreatedAt = item.CreatedAt
		page.UpdatedAt = item.UpdatedAt
		item.DatabaseID = databaseID
		item.Page = page
		item.PropertyMap = make(map[string]domain.DatabaseValue)
		items = append(items, item)
	}
	// load property slugs for view display
	propRows, err := s.db.QueryContext(ctx, `SELECT dp.id, dp.slug FROM database_properties dp WHERE dp.database_id = ?`, databaseID)
	if err != nil {
		return nil, fmt.Errorf("query property map: %w", err)
	}
	propIDs := make(map[string]string)
	for propRows.Next() {
		var id, slug string
		if err := propRows.Scan(&id, &slug); err != nil {
			propRows.Close()
			return nil, fmt.Errorf("scan property map: %w", err)
		}
		propIDs[id] = slug
	}
	propRows.Close()
	for idx := range items {
		rows, err := s.db.QueryContext(ctx, `SELECT dv.id, dv.property_id, dv.value, dv.is_computed, dv.created_at, dv.updated_at FROM database_values dv WHERE dv.database_item_id = ?`, items[idx].ID)
		if err != nil {
			return nil, fmt.Errorf("query item values: %w", err)
		}
		for rows.Next() {
			var value domain.DatabaseValue
			var raw string
			var isComputed int
			if err := rows.Scan(&value.ID, &value.PropertyID, &raw, &isComputed, &value.CreatedAt, &value.UpdatedAt); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan value: %w", err)
			}
			if raw != "" {
				var parsed any
				if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
					value.RawValue = parsed
				}
			}
			value.ItemID = items[idx].ID
			value.IsComputed = isComputed == 1
			if slug, ok := propIDs[value.PropertyID]; ok {
				items[idx].PropertyMap[slug] = value
			}
		}
		rows.Close()
	}
	return items, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
