CREATE TABLE IF NOT EXISTS page_links (
    source_page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    target_page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL,
    PRIMARY KEY (source_page_id, target_page_id)
);

CREATE INDEX IF NOT EXISTS idx_page_links_target ON page_links(target_page_id);
