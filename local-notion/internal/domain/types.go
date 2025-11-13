package domain

import "time"

// Page represents a free-form content page.
type Page struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Content      string    `json:"content"`
	ParentPageID *string   `json:"parent_page_id"`
	CoverImageID *string   `json:"cover_image_id"`
	Icon         *string   `json:"icon"`
	Tags         []string  `json:"tags"`
	IsArchived   bool      `json:"is_archived"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Database represents a structured collection of page-backed items.
type Database struct {
	ID          string             `json:"id"`
	Slug        string             `json:"slug"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Icon        *string            `json:"icon"`
	CoverImage  *string            `json:"cover_image_id"`
	IsArchived  bool               `json:"is_archived"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Properties  []DatabaseProperty `json:"properties"`
	Views       []DatabaseView     `json:"views"`
}

// DatabaseProperty defines a field for items in a database.
type DatabaseProperty struct {
	ID         string         `json:"id"`
	DatabaseID string         `json:"database_id"`
	Name       string         `json:"name"`
	Slug       string         `json:"slug"`
	Type       PropertyType   `json:"type"`
	Config     map[string]any `json:"config"`
	IsRequired bool           `json:"is_required"`
	Default    any            `json:"default"`
	OrderIndex int            `json:"order_index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// PropertyType enumerates supported property field kinds.
type PropertyType string

const (
	PropertyTypeText        PropertyType = "text"
	PropertyTypeNumber      PropertyType = "number"
	PropertyTypeSelect      PropertyType = "select"
	PropertyTypeMultiSelect PropertyType = "multi_select"
	PropertyTypeDate        PropertyType = "date"
	PropertyTypeCheckbox    PropertyType = "checkbox"
	PropertyTypeRelation    PropertyType = "relation"
	PropertyTypeURL         PropertyType = "url"
	PropertyTypeEmail       PropertyType = "email"
	PropertyTypePhone       PropertyType = "phone"
	PropertyTypeMedia       PropertyType = "media"
	PropertyTypeFormula     PropertyType = "formula"
	PropertyTypeRollup      PropertyType = "rollup"
)

// DatabaseView configures a saved layout for rendering items.
type DatabaseView struct {
	ID            string         `json:"id"`
	DatabaseID    string         `json:"database_id"`
	Name          string         `json:"name"`
	Type          ViewType       `json:"type"`
	Filters       map[string]any `json:"filters"`
	Sorts         []ViewSort     `json:"sorts"`
	Grouping      map[string]any `json:"grouping"`
	Display       []string       `json:"display_properties"`
	LayoutOptions map[string]any `json:"layout_options"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// ViewType enumerates supported database view renderers.
type ViewType string

const (
	ViewTypeTable    ViewType = "table"
	ViewTypeList     ViewType = "list"
	ViewTypeGallery  ViewType = "gallery"
	ViewTypeBoard    ViewType = "board"
	ViewTypeCalendar ViewType = "calendar"
	ViewTypeTimeline ViewType = "timeline"
)

// ViewSort configures ordering for a view.
type ViewSort struct {
	PropertyID string `json:"property_id"`
	Direction  string `json:"direction"`
}

// DatabaseItem is a page backed entry in a database.
type DatabaseItem struct {
	ID          string                   `json:"id"`
	DatabaseID  string                   `json:"database_id"`
	Page        Page                     `json:"page"`
	Position    int                      `json:"position"`
	IsArchived  bool                     `json:"is_archived"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	PropertyMap map[string]DatabaseValue `json:"properties"`
}

// DatabaseValue stores a property value for a database item.
type DatabaseValue struct {
	ID         string    `json:"id"`
	ItemID     string    `json:"database_item_id"`
	PropertyID string    `json:"property_id"`
	RawValue   any       `json:"value"`
	IsComputed bool      `json:"is_computed"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
