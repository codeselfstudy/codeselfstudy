package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
)

// End-to-end check that /api/todos GET/POST go through the auth middleware
// and round-trip via the in-memory DB.

func newTestTodos(t *testing.T) *db.Todos {
	t.Helper()
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.Migrate(t.Context(), conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &db.Todos{DB: conn}
}

func TestApiTodosCreateThenList(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	repo := newTestTodos(t)
	e := newServer(fixtureDir(t), v, repo)

	bearer := "Bearer " + f.sign(t)

	// Create
	body := strings.NewReader(`{"title":"buy milk"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/todos", body)
	req.Header.Set(echo.HeaderAuthorization, bearer)
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create status: want 201 got %d body=%s", rec.Code, rec.Body.String())
	}
	var created db.Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.Title != "buy milk" || created.ID == 0 {
		t.Fatalf("created: %+v", created)
	}

	// List
	listReq := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	listReq.Header.Set(echo.HeaderAuthorization, bearer)
	listRec := httptest.NewRecorder()
	e.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("list status: want 200 got %d body=%s", listRec.Code, listRec.Body.String())
	}
	var rows []db.Todo
	if err := json.Unmarshal(listRec.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != created.ID {
		t.Fatalf("rows: %+v", rows)
	}
}

func TestApiTodosRequireAuth(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	repo := newTestTodos(t)
	e := newServer(fixtureDir(t), v, repo)

	cases := []struct {
		method, path string
	}{
		{http.MethodGet, "/api/todos"},
		{http.MethodPost, "/api/todos"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(`{}`)))
			req.Header.Set(echo.HeaderContentType, "application/json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status: want 401 got %d", rec.Code)
			}
		})
	}
}

func TestApiTodosRejectsBlankTitle(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	repo := newTestTodos(t)
	e := newServer(fixtureDir(t), v, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(`{"title":"  "}`))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+f.sign(t))
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: want 400 got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApiTodosDisabledWithoutDB(t *testing.T) {
	// Auth wired up, but no DB → /api/todos should fall through to the
	// /api/* JSON-404 catchall, not 401.
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	e := newServer(fixtureDir(t), v, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+f.sign(t))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: want 404 got %d body=%s", rec.Code, rec.Body.String())
	}
}
