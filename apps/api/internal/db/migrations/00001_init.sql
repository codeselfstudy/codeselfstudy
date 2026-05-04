-- +goose Up
CREATE TABLE todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    created_at INTEGER DEFAULT (unixepoch())
);

-- +goose Down
DROP TABLE todos;
