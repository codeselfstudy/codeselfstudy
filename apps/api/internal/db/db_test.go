package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func newTestDB(t *testing.T) *Todos {
	t.Helper()
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if err := ApplySchema(d); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return &Todos{DB: d}
}

func TestOpenRejectsRemoteSchemes(t *testing.T) {
	cases := []string{
		"libsql://example.turso.io",
		"http://localhost:8080/db",
		"https://example.com",
		"wss://example.com",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			if _, err := Open(u); err == nil {
				t.Fatalf("expected error for %q, got nil", u)
			}
		})
	}
}

func TestOpenRejectsEmptyURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestOpenLocalFileWorks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	if err := d.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestTodosCreateAndList(t *testing.T) {
	repo := newTestDB(t)
	ctx := t.Context()

	a, err := repo.Create(ctx, "buy milk")
	if err != nil {
		t.Fatalf("create a: %v", err)
	}
	b, err := repo.Create(ctx, "  walk dog  ") // trimmed
	if err != nil {
		t.Fatalf("create b: %v", err)
	}

	if a.ID == 0 || b.ID == 0 {
		t.Fatalf("expected ids to be assigned, got a=%d b=%d", a.ID, b.ID)
	}
	if b.ID <= a.ID {
		t.Errorf("expected ids to increase: a=%d b=%d", a.ID, b.ID)
	}
	if b.Title != "walk dog" {
		t.Errorf("title not trimmed: %q", b.Title)
	}
	if a.CreatedAt.IsZero() || b.CreatedAt.IsZero() {
		t.Errorf("created_at should be populated: a=%v b=%v", a.CreatedAt, b.CreatedAt)
	}
	if time.Since(a.CreatedAt) > 5*time.Second {
		t.Errorf("created_at too old: %v", a.CreatedAt)
	}

	rows, err := repo.List(ctx, 0) // 0 → default limit
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: want 2 got %d", len(rows))
	}
	// Newest-first ordering.
	if rows[0].ID != b.ID || rows[1].ID != a.ID {
		t.Errorf("order: want [%d, %d] got [%d, %d]", b.ID, a.ID, rows[0].ID, rows[1].ID)
	}
}

func TestTodosCreateRejectsBlankTitle(t *testing.T) {
	repo := newTestDB(t)
	cases := []string{"", "   ", "\t\n"}
	for _, title := range cases {
		t.Run(title, func(t *testing.T) {
			if _, err := repo.Create(t.Context(), title); err == nil {
				t.Fatal("expected error for blank title")
			}
		})
	}
}

func TestTodosListEmpty(t *testing.T) {
	repo := newTestDB(t)
	rows, err := repo.List(t.Context(), 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows: want 0 got %d", len(rows))
	}
}

func TestTodosListRespectsLimit(t *testing.T) {
	repo := newTestDB(t)
	ctx := t.Context()
	for i := 0; i < 5; i++ {
		if _, err := repo.Create(ctx, "t"); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	rows, err := repo.List(ctx, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: want 2 got %d", len(rows))
	}
}

func TestApplySchemaIsIdempotent(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	if err := ApplySchema(d); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if err := ApplySchema(d); err != nil {
		t.Fatalf("second apply: %v", err)
	}

	// Quick sanity check: insert/select round-trip after re-apply.
	repo := &Todos{DB: d}
	if _, err := repo.Create(context.Background(), "ok"); err != nil {
		t.Fatalf("create after re-apply: %v", err)
	}
}
