-- +goose Up
CREATE TABLE
    resources (
        id INTEGER PRIMARY KEY,
        path TEXT NOT NULL UNIQUE,
        is_dir INTEGER NOT NULL CHECK (is_dir IN (0, 1)),
        created_at INTEGER NOT NULL DEFAULT (unixepoch ())
    ) STRICT;

CREATE TABLE
    resource_tags (
        resource_id INTEGER NOT NULL,
        tag TEXT NOT NULL COLLATE NOCASE CHECK (length (tag) > 0),
        PRIMARY KEY (resource_id, tag),
        FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE CASCADE
    ) STRICT;

CREATE INDEX idx_resource_tags_tag ON resource_tags (tag);

-- +goose Down
DROP TABLE resource_tags;

DROP TABLE resources;