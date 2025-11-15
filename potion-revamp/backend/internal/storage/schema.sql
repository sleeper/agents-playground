PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS pages (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    icon TEXT,
    parent_id TEXT REFERENCES pages(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS blocks (
    id TEXT PRIMARY KEY,
    page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    type TEXT NOT NULL,
    data TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_blocks_page_position ON blocks(page_id, position);

CREATE TABLE IF NOT EXISTS page_links (
    source_page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    target_page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    PRIMARY KEY (source_page_id, target_page_id)
);

CREATE TABLE IF NOT EXISTS databases (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS database_properties (
    id TEXT PRIMARY KEY,
    database_id TEXT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    options TEXT,
    position INTEGER NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_database_properties_name ON database_properties(database_id, name);

CREATE TABLE IF NOT EXISTS database_views (
    id TEXT PRIMARY KEY,
    database_id TEXT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    options TEXT,
    position INTEGER NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_database_views_name ON database_views(database_id, name);

CREATE TABLE IF NOT EXISTS database_entries (
    id TEXT PRIMARY KEY,
    database_id TEXT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS database_entry_properties (
    entry_id TEXT NOT NULL REFERENCES database_entries(id) ON DELETE CASCADE,
    property_id TEXT NOT NULL REFERENCES database_properties(id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    PRIMARY KEY (entry_id, property_id)
);

CREATE TABLE IF NOT EXISTS embedded_database_views (
    block_id TEXT PRIMARY KEY REFERENCES blocks(id) ON DELETE CASCADE,
    view_id TEXT NOT NULL REFERENCES database_views(id) ON DELETE CASCADE
);
