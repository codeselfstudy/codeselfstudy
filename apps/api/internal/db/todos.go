package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Todo mirrors the row shape of apps/web/src/db/schema.ts.
type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

// Todos is a thin SQL repository for the todos table.
type Todos struct {
	DB *sql.DB
}

// List returns todos in newest-first order. The cap is a safety rail for
// the experiment — paginate properly once we have a real UI.
func (t *Todos) List(ctx context.Context, limit int) ([]Todo, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := t.DB.QueryContext(ctx,
		`SELECT id, title, created_at FROM todos ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("todos: list: %w", err)
	}
	defer rows.Close()

	var out []Todo
	for rows.Next() {
		var (
			td   Todo
			unix sql.NullInt64
		)
		if err := rows.Scan(&td.ID, &td.Title, &unix); err != nil {
			return nil, fmt.Errorf("todos: scan: %w", err)
		}
		if unix.Valid {
			td.CreatedAt = time.Unix(unix.Int64, 0).UTC()
		}
		out = append(out, td)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("todos: rows: %w", err)
	}
	return out, nil
}

// Create inserts a todo and returns the newly assigned id and server-side
// timestamp. Title is trimmed; an empty title is rejected.
func (t *Todos) Create(ctx context.Context, title string) (Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Todo{}, errors.New("todos: title is required")
	}

	res, err := t.DB.ExecContext(ctx,
		`INSERT INTO todos (title) VALUES (?)`,
		title,
	)
	if err != nil {
		return Todo{}, fmt.Errorf("todos: insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Todo{}, fmt.Errorf("todos: last insert id: %w", err)
	}

	var unix sql.NullInt64
	if err := t.DB.QueryRowContext(ctx,
		`SELECT created_at FROM todos WHERE id = ?`, id,
	).Scan(&unix); err != nil {
		return Todo{}, fmt.Errorf("todos: read created_at: %w", err)
	}

	td := Todo{ID: id, Title: title}
	if unix.Valid {
		td.CreatedAt = time.Unix(unix.Int64, 0).UTC()
	}
	return td, nil
}
