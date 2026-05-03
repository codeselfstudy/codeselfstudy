package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/auth"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
)

const staticDir = "static"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wire WorkOS auth when the env is set. In barebones dev (no env) the
	// static-serve + /healthz path still works — useful for smoke testing
	// the binary without a real WorkOS tenant.
	v := newVerifierFromEnv()
	if v != nil {
		if err := v.Start(ctx); err != nil {
			log.Fatalf("auth: %v", err)
		}
	} else {
		log.Printf("auth: WORKOS_CLIENT_ID / WORKOS_API_HOSTNAME not set; /api/me disabled")
	}

	// Open the database when DATABASE_URL is set. Same opt-in shape as auth.
	var todos *db.Todos
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		conn, err := db.Open(dbURL)
		if err != nil {
			log.Fatalf("db: %v", err)
		}
		defer conn.Close()
		todos = &db.Todos{DB: conn}
	} else {
		log.Printf("db: DATABASE_URL not set; /api/todos disabled")
	}

	e := newServer(staticDir, v, todos)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
	}
}

// newVerifierFromEnv reads WorkOS client config from the environment. The
// web app already validates the same pair as VITE_WORKOS_CLIENT_ID /
// VITE_WORKOS_API_HOSTNAME (a Vite prefix convention; the OS-level vars are
// the same values), so production deploys reuse them. Returns nil if either
// is missing — the caller falls back to running without auth.
func newVerifierFromEnv() *auth.Verifier {
	clientID := firstNonEmpty(os.Getenv("WORKOS_CLIENT_ID"), os.Getenv("VITE_WORKOS_CLIENT_ID"))
	hostname := firstNonEmpty(os.Getenv("WORKOS_API_HOSTNAME"), os.Getenv("VITE_WORKOS_API_HOSTNAME"))
	if clientID == "" || hostname == "" {
		return nil
	}
	v, err := auth.NewVerifier(clientID, hostname)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}
	return v
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// handleMe returns the validated claims for the authenticated user. Light
// shape on purpose — tighten the response to whatever the frontend ends up
// needing.
func handleMe(c echo.Context) error {
	tok := auth.Claims(c)
	if tok == nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	resp := map[string]any{
		"sub": tok.Subject(),
		"iss": tok.Issuer(),
		"exp": tok.Expiration().Unix(),
	}
	if email, ok := tok.Get("email"); ok {
		resp["email"] = email
	}
	return c.JSON(http.StatusOK, resp)
}

func handleListTodos(repo *db.Todos) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := repo.List(c.Request().Context(), 100)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if rows == nil {
			rows = []db.Todo{}
		}
		return c.JSON(http.StatusOK, rows)
	}
}

func handleCreateTodo(repo *db.Todos) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body struct {
			Title string `json:"title"`
		}
		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		td, err := repo.Create(c.Request().Context(), body.Title)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusCreated, td)
	}
}

// newServer wires the routes and middleware for an Echo instance rooted at
// staticRoot. When v is non-nil, /api/me is mounted under WorkOS-validated
// auth; when todos is non-nil, /api/todos GET/POST are mounted under the
// same auth. Either or both may be nil — anything left out falls through to
// the /api/* JSON 404. Split out from main so tests can build a server
// against fixtures.
func newServer(staticRoot string, v *auth.Verifier, todos *db.Todos) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	// Gzip wraps the ResponseWriter, which breaks the connection hijack a
	// future WebSocket upgrade will need. Skip it on /ws so Phase 2's choices
	// don't quietly break Phase 4's hub.
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/ws"
		},
	}))

	e.GET("/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// Authenticated API routes go before the /api/* catchall so the static
	// paths win over the JSON-404 fallback.
	if v != nil {
		api := e.Group("/api", auth.Middleware(v))
		api.GET("/me", handleMe)
		if todos != nil {
			api.GET("/todos", handleListTodos(todos))
			api.POST("/todos", handleCreateTodo(todos))
		}
	}

	// Reserve /api/* and /ws so unrecognized requests there reach Echo's
	// default 404 (JSON, no SSG fallback).
	e.Any("/api/*", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	e.Any("/ws", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})

	e.GET("/*", staticHandler(staticRoot))
	return e
}

// staticHandler serves prerendered files out of staticRoot. Unlike
// http.FileServer it owns the 404 path: if a request can't be resolved to a
// file (or its index.html), it serves staticRoot/404.html with a 404 status —
// matching what `bun run build` ships under apps/web/.output/public/. That
// keeps SSG-style not-found pages working without abusing Echo's error
// handler (which never fires once FileServer writes to the wire).
func staticHandler(staticRoot string) echo.HandlerFunc {
	return func(c echo.Context) error {
		clean := filepath.Clean("/" + c.Request().URL.Path)
		if strings.Contains(clean, "..") {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		rel := strings.TrimPrefix(clean, "/")

		if path, ok := resolveStatic(staticRoot, rel); ok {
			return c.File(path)
		}

		if data, err := os.ReadFile(filepath.Join(staticRoot, "404.html")); err == nil {
			return c.HTMLBlob(http.StatusNotFound, data)
		}
		return echo.NewHTTPError(http.StatusNotFound)
	}
}

// resolveStatic walks the same fallbacks browsers expect from a prerendered
// site: an exact file, a directory's index.html, or `<path>.html`.
func resolveStatic(staticRoot, rel string) (string, bool) {
	candidates := []string{
		filepath.Join(staticRoot, rel),
		filepath.Join(staticRoot, rel, "index.html"),
		filepath.Join(staticRoot, rel+".html"),
	}
	for _, p := range candidates {
		info, err := os.Stat(p)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return "", false
		}
		if info.IsDir() {
			continue
		}
		return p, true
	}
	return "", false
}
