package models

import "time"

// Page represents a single document in the workspace.
type Page struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Icon      string    `json:"icon,omitempty"`
    ParentID  *string   `json:"parentId,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// BlockType enumerates the supported block types.
type BlockType string

const (
    BlockTypeMarkdown    BlockType = "markdown"
    BlockTypeHeading     BlockType = "heading"
    BlockTypePageLink    BlockType = "pageLink"
    BlockTypeDatabaseRef BlockType = "databaseView"
)

// Block represents an ordered building unit inside a page.
type Block struct {
    ID       string                 `json:"id"`
    PageID   string                 `json:"pageId"`
    Position int                    `json:"position"`
    Type     BlockType              `json:"type"`
    Data     map[string]interface{} `json:"data"`
}

// Database describes a structured collection of entries.
type Database struct {
    ID          string            `json:"id"`
    Title       string            `json:"title"`
    Description string            `json:"description,omitempty"`
    Icon        string            `json:"icon,omitempty"`
    CreatedAt   time.Time         `json:"createdAt"`
    UpdatedAt   time.Time         `json:"updatedAt"`
    Properties  []DatabaseProperty `json:"properties"`
    Views       []DatabaseView     `json:"views"`
}

// DatabaseProperty describes a column inside a database.
type DatabaseProperty struct {
    ID         string                 `json:"id"`
    DatabaseID string                 `json:"databaseId"`
    Name       string                 `json:"name"`
    Type       PropertyType           `json:"type"`
    Options    map[string]interface{} `json:"options,omitempty"`
    Position   int                    `json:"position"`
}

// PropertyType enumerates supported property types.
type PropertyType string

const (
    PropertyTypeTitle      PropertyType = "title"
    PropertyTypeText       PropertyType = "text"
    PropertyTypeNumber     PropertyType = "number"
    PropertyTypeSelect     PropertyType = "select"
    PropertyTypeMultiSelect PropertyType = "multi_select"
    PropertyTypeDate       PropertyType = "date"
    PropertyTypeCheckbox   PropertyType = "checkbox"
    PropertyTypeRelation   PropertyType = "relation"
)

// DatabaseEntry represents a row in a database.
type DatabaseEntry struct {
    ID         string                 `json:"id"`
    DatabaseID string                 `json:"databaseId"`
    Title      string                 `json:"title"`
    Properties map[string]interface{} `json:"properties"`
    CreatedAt  time.Time              `json:"createdAt"`
    UpdatedAt  time.Time              `json:"updatedAt"`
}

// DatabaseView captures a saved presentation of a database.
type DatabaseView struct {
    ID         string                 `json:"id"`
    DatabaseID string                 `json:"databaseId"`
    Name       string                 `json:"name"`
    Type       ViewType               `json:"type"`
    Options    map[string]interface{} `json:"options,omitempty"`
    Position   int                    `json:"position"`
}

// ViewType enumerates supported database view renderings.
type ViewType string

const (
    ViewTypeTable  ViewType = "table"
    ViewTypeKanban ViewType = "kanban"
    ViewTypeGallery ViewType = "gallery"
)

// PageLink models a resolved link between pages.
type PageLink struct {
    SourcePageID string `json:"sourcePageId"`
    TargetPageID string `json:"targetPageId"`
}

// PageWithBlocks bundles a page and its blocks for transport.
type PageWithBlocks struct {
    Page   Page    `json:"page"`
    Blocks []Block `json:"blocks"`
    Backlinks []Page `json:"backlinks"`
}

// DatabaseWithEntries includes entries for a given view rendering.
type DatabaseWithEntries struct {
    Database Database      `json:"database"`
    Entries  []DatabaseEntry `json:"entries"`
}
