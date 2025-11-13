CREATE TABLE IF NOT EXISTS pages (
    id TEXT PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    content TEXT,
    parent_page_id TEXT REFERENCES pages(id) ON DELETE SET NULL,
    cover_image_id TEXT,
    icon TEXT,
    tags TEXT,
    is_archived INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    storage_path TEXT NOT NULL,
    thumbnail_path TEXT,
    created_at DATETIME NOT NULL,
    uploaded_by TEXT
);

CREATE TABLE IF NOT EXISTS databases (
    id TEXT PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    cover_image_id TEXT,
    is_archived INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS database_properties (
    id TEXT PRIMARY KEY,
    database_id TEXT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    type TEXT NOT NULL,
    config TEXT,
    is_required INTEGER NOT NULL DEFAULT 0,
    default_value TEXT,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(database_id, slug)
);

CREATE TABLE IF NOT EXISTS database_views (
    id TEXT PRIMARY KEY,
    database_id TEXT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    filters TEXT,
    sorts TEXT,
    grouping TEXT,
    display_properties TEXT,
    layout_options TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS database_items (
    id TEXT PRIMARY KEY,
    database_id TEXT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0,
    is_archived INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS database_values (
    id TEXT PRIMARY KEY,
    database_item_id TEXT NOT NULL REFERENCES database_items(id) ON DELETE CASCADE,
    property_id TEXT NOT NULL REFERENCES database_properties(id) ON DELETE CASCADE,
    value TEXT,
    is_computed INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(database_item_id, property_id)
);

CREATE INDEX IF NOT EXISTS idx_pages_parent ON pages(parent_page_id);
CREATE INDEX IF NOT EXISTS idx_database_items_database ON database_items(database_id);
CREATE INDEX IF NOT EXISTS idx_database_values_property ON database_values(property_id);
