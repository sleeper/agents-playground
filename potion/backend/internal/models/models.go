package models

import "time"

// Page represents a markdown page that can optionally have a parent page.
type Page struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ParentID  *string   `json:"parentId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PageLink represents a reference between two pages.
type PageLink struct {
	ID        string    `json:"id"`
	SourceID  string    `json:"sourceId"`
	TargetID  string    `json:"targetId"`
	CreatedAt time.Time `json:"createdAt"`
}

// Database represents a lightweight Notion-style database definition.
type Database struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	View        string            `json:"view"`
	Schema      map[string]string `json:"schema"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// DatabaseEntry represents a row/item inside a database.
type DatabaseEntry struct {
	ID         string                 `json:"id"`
	DatabaseID string                 `json:"databaseId"`
	Title      string                 `json:"title"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}
