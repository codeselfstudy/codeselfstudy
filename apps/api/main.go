package main

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const staticDir = "static"

func main() {
	e := newServer(staticDir)

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

// newServer wires the routes and middleware for an Echo instance rooted at
// staticRoot. Split out from main so tests can build a server against a
// fixture directory.
func newServer(staticRoot string) *echo.Echo {
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
