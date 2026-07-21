package main

import (
	"context"
	"database/sql"
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
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/digest"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/extract"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/ingest"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
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

	ing, database := newIngestFromEnv(ctx)
	if database != nil {
		defer database.Close()
	}

	e := newServer(staticDir, v, ing)

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

// authTokenOf extracts the authToken query value from a libsql DATABASE_URL, or
// "" if absent. Used to redact the token from database errors before logging.
func authTokenOf(dbURL string) string {
	const key = "authToken="
	i := strings.Index(dbURL, key)
	if i < 0 {
		return ""
	}
	v := dbURL[i+len(key):]
	if amp := strings.IndexByte(v, '&'); amp >= 0 {
		v = v[:amp]
	}
	return v
}

// redactToken replaces every occurrence of token in s with "***". A libsql
// driver error can embed the full connection URL including ?authToken=<token>;
// redacting the opaque token keeps the useful cause (dial/connect/auth failure)
// in the log while keeping the secret out of it. A no-op when token is "".
func redactToken(s, token string) string {
	if token == "" {
		return s
	}
	return strings.ReplaceAll(s, token, "***")
}

// newIngestFromEnv builds the email-ingest handlers from the environment. It
// returns (nil, nil) when the pipeline is not configured (no DATABASE_URL /
// INGEST_TOKEN), so the server still runs static-only — mirroring the WorkOS
// graceful-degrade above. The returned *sql.DB (when non-nil) is owned by the
// caller, which must Close it on shutdown. Fatal only on a genuinely broken
// setup (bad duration, unreachable DB, bad Gemini client).
func newIngestFromEnv(ctx context.Context) (*ingest.Handlers, *sql.DB) {
	cfg, err := ingest.Load(os.Getenv)
	if err != nil {
		log.Fatalf("ingest: %v", err)
	}
	if !cfg.Enabled() {
		log.Printf("ingest: DATABASE_URL / INGEST_TOKEN not set; /api/ingest disabled")
		return nil, nil
	}

	tok := authTokenOf(cfg.DatabaseURL)
	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("ingest: db open: %s", redactToken(err.Error(), tok))
	}
	if err := db.Migrate(ctx, database); err != nil {
		_ = database.Close()
		log.Fatalf("ingest: db migrate: %s", redactToken(err.Error(), tok))
	}

	var extractor extract.Extractor = extract.Disabled{}
	if cfg.GeminiAPIKey != "" {
		g, err := extract.NewGemini(ctx, cfg.GeminiAPIKey, cfg.GeminiModel, "")
		if err != nil {
			_ = database.Close()
			log.Fatalf("ingest: gemini: %v", err)
		}
		extractor = g
	} else {
		log.Printf("ingest: GEMINI_API_KEY not set; extraction disabled (/api/ingest will 500)")
	}

	poster := digest.NewHTTPPoster(cfg.SlackWebhookURL)
	return ingest.New(cfg, store.New(database), extractor, poster), database
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

// newServer wires the routes and middleware for an Echo instance rooted at
// staticRoot. When v is non-nil, /api/me is mounted under WorkOS-validated
// auth; otherwise the route falls through to the /api/* JSON 404. Split out
// from main so tests can build a server against fixtures.
func newServer(staticRoot string, v *auth.Verifier, ing *ingest.Handlers) *echo.Echo {
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

	// Every read endpoint accepts both GET and HEAD. HEAD is GET-without-body
	// per RFC 9110; declaring only GET makes Echo return 405 to HEAD probes
	// from curl -I, monitoring agents, and load balancers, which is wrong.
	getOrHead := []string{http.MethodGet, http.MethodHead}

	e.Match(getOrHead, "/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// Authenticated API routes go before the /api/* catchall so the static
	// paths win over the JSON-404 fallback.
	if v != nil {
		api := e.Group("/api", auth.Middleware(v))
		api.Match(getOrHead, "/me", handleMe)
	}

	// Mount the email-ingest routes (POST /api/ingest, /api/admin/digest) when
	// the pipeline is configured — before the /api/* reservation so the concrete
	// POST paths win over the JSON-404 wildcard.
	if ing != nil {
		ing.Register(e)
	}

	// Reserve /api/* and /ws so unrecognized requests there reach Echo's
	// default 404 (JSON, no SSG fallback). Any() already covers HEAD.
	e.Any("/api/*", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})
	e.Any("/ws", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound)
	})

	e.Match(getOrHead, "/*", staticHandler(staticRoot))
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
		reqPath := c.Request().URL.Path
		clean := filepath.Clean("/" + reqPath)
		if strings.Contains(clean, "..") {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		rel := strings.TrimPrefix(clean, "/")

		// Legacy redirects run first — before trailing-slash canonicalization
		// and before static resolution. This ordering matters: `/index.html` is
		// both a real file in the static root and a legacy redirect to `/`, so
		// resolving first would serve it 200 instead of redirecting.
		if r, ok := findRedirect(clean); ok {
			target := r.to
			if q := c.Request().URL.RawQuery; q != "" {
				target += "?" + q
			}
			return c.Redirect(r.status, target)
		}

		// Canonicalize to the trailing-slash form. An extensionless path with no
		// trailing slash that resolves to a directory's index.html gets a 301 to
		// `path + "/"`, so every page has a single canonical URL (the site is
		// built with trailingSlash "always"). The query string is preserved.
		// Paths with an extension (assets) and unresolvable paths are left alone
		// — no redirected assets, and no redirect-then-404 double hop. We avoid
		// Echo's AddTrailingSlash middleware precisely because it would redirect
		// asset URLs too.
		if reqPath != "/" && !strings.HasSuffix(reqPath, "/") && filepath.Ext(reqPath) == "" {
			if dirIndexExists(staticRoot, rel) {
				// Build the target from the normalized `clean`, not the raw
				// request path: a request for `//about` must redirect to
				// `/about/`, not the protocol-relative `//about/` (which a
				// browser reads as the off-origin host `about`).
				target := clean + "/"
				if q := c.Request().URL.RawQuery; q != "" {
					target += "?" + q
				}
				return c.Redirect(http.StatusMovedPermanently, target)
			}
		}

		if path, ok := resolveStatic(staticRoot, rel); ok {
			return c.File(path)
		}

		if data, err := os.ReadFile(filepath.Join(staticRoot, "404.html")); err == nil {
			return c.HTMLBlob(http.StatusNotFound, data)
		}
		return echo.NewHTTPError(http.StatusNotFound)
	}
}

// dirIndexExists reports whether rel resolves to a directory index.html — the
// specific candidate that trailing-slash canonicalization applies to.
func dirIndexExists(staticRoot, rel string) bool {
	info, err := os.Stat(filepath.Join(staticRoot, rel, "index.html"))
	return err == nil && !info.IsDir()
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
