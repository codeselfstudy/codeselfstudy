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
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/resolve"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/session"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/users"
)

const staticDir = "static"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// `server -migrate` (the Fly release_command) applies database migrations
	// and exits, so schema changes run once per deploy in a temporary machine
	// before the new version serves traffic — a migration failure aborts the
	// deploy instead of crash-looping the app.
	if migrateRequested(os.Args[1:]) {
		runMigrate(ctx)
		return
	}

	// Wire WorkOS auth when the env is set. In barebones dev (no env) the
	// static-serve + /healthz path still works — useful for smoke testing
	// the binary without a real WorkOS tenant.
	v := newVerifierFromEnv()
	if v != nil {
		if err := v.Start(ctx); err != nil {
			log.Fatalf("auth: %v", err)
		}
	} else {
		log.Printf("auth: WORKOS_CLIENT_ID / WORKOS_API_HOSTNAME not set; auth disabled (/api/me and /auth/* off)")
	}

	// The server-side session (sealed first-party cookie) needs the extra
	// server secrets on top of the verifier. When they're absent the site still
	// runs; /auth/* and the cookie-gated /api/me are simply off.
	sess := newSessionFromEnv(v)

	// One database handle, opened on DATABASE_URL alone, shared by accounts and
	// (when INGEST_TOKEN is also set) the email pipeline. Accounts must not
	// require the deals pipeline, so the DB is no longer welded to ingest.
	database := newDatabaseFromEnv(ctx)
	if database != nil {
		defer database.Close()
	}

	ing := newIngestFromEnv(ctx, database)
	usr := newUsersFromEnv(database, sess)

	e := newServer(staticDir, v, sess, ing, usr)

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

// newVerifierFromEnv reads WorkOS client config from the environment. Only the
// canonical server names are accepted. Earlier revisions also honoured the
// PUBLIC_/VITE_-prefixed aliases the browser bundle once used, but
// session.LoadConfig never did — so a deploy carrying only the aliases brought
// the verifier up while leaving the session off, and /auth/* 404'd while
// /api/me answered. The web app no longer reads any WorkOS var, so the aliases
// have no remaining purpose. Returns nil if either value is missing — the
// caller falls back to running without auth.
func newVerifierFromEnv() *auth.Verifier {
	clientID := os.Getenv("WORKOS_CLIENT_ID")
	hostname := os.Getenv("WORKOS_API_HOSTNAME")
	if clientID == "" || hostname == "" {
		return nil
	}
	v, err := auth.NewVerifier(clientID, hostname)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}
	return v
}

// newSessionFromEnv builds the server-side session Manager when the extra
// WorkOS secrets (WORKOS_API_KEY, WORKOS_COOKIE_PASSWORD, APP_BASE_URL) are set,
// reusing the JWKS verifier for access-token validation. Returns nil — leaving
// /auth/* and the cookie session off — when the verifier is absent or the config
// is incomplete, so barebones/static-only boots keep working.
func newSessionFromEnv(v *auth.Verifier) *session.Manager {
	if v == nil {
		return nil
	}
	cfg := session.LoadConfig(os.Getenv)
	if missing := cfg.Missing(); len(missing) > 0 {
		// The verifier came up, so auth is clearly intended — but sign-in is
		// about to be silently unavailable. Name the exact culprits and say
		// what breaks, loudly: a bare "not fully set" listing every candidate
		// var (all of which were in fact set) once cost a full debugging cycle
		// on a production 404.
		log.Printf("auth: WARNING: WorkOS verifier is configured but the server-side session is NOT; missing: %s", strings.Join(missing, ", "))
		log.Printf("auth: WARNING: /auth/login, /auth/callback and /auth/logout will return 404 and nobody can sign in")
		return nil
	}
	m, err := session.New(cfg, v)
	if err != nil {
		log.Fatalf("auth: session: %v", err)
	}
	return m
}

// newDatabaseFromEnv opens the shared database when DATABASE_URL is set,
// returning nil (static-only) otherwise. Accounts need only DATABASE_URL; the
// ingest pipeline layers INGEST_TOKEN on top. The returned handle is owned by the
// caller, which must Close it on shutdown. Migrate on boot only for a local
// SQLite database (dev ergonomics); a remote libsql/Turso database is migrated
// out of band by `server -migrate` (the Fly release_command), so a migration
// failure aborts the deploy rather than crash-looping the serving process.
func newDatabaseFromEnv(ctx context.Context) *sql.DB {
	if os.Getenv("DATABASE_URL") == "" {
		log.Printf("db: DATABASE_URL not set; accounts and /api/ingest disabled")
		return nil
	}
	dbURL := db.ResolveURL(os.Getenv("DATABASE_URL"), os.Getenv("TURSO_AUTH_TOKEN"))
	database, err := db.Open(dbURL)
	if err != nil {
		log.Fatalf("db: open: %s", db.RedactToken(err.Error(), dbURL))
	}
	if !db.IsRemote(dbURL) {
		if err := db.Migrate(ctx, database); err != nil {
			_ = database.Close()
			log.Fatalf("db: migrate: %s", db.RedactToken(err.Error(), dbURL))
		}
	}
	return database
}

// newIngestFromEnv builds the email-ingest handlers over the shared database. It
// returns nil when the pipeline is not configured (no DATABASE_URL / INGEST_TOKEN,
// or no database), so the server still runs static-only — mirroring the WorkOS
// graceful-degrade above. Fatal only on a genuinely broken setup (bad duration,
// bad Gemini client).
func newIngestFromEnv(ctx context.Context, database *sql.DB) *ingest.Handlers {
	cfg, err := ingest.Load(os.Getenv)
	if err != nil {
		log.Fatalf("ingest: %v", err)
	}
	if !cfg.Enabled() || database == nil {
		log.Printf("ingest: DATABASE_URL / INGEST_TOKEN not both set; /api/ingest disabled")
		return nil
	}

	var extractor extract.Extractor = extract.Disabled{}
	if cfg.GeminiAPIKey != "" {
		g, err := extract.NewGemini(ctx, cfg.GeminiAPIKey, cfg.GeminiModel, "")
		if err != nil {
			log.Fatalf("ingest: gemini: %v", err)
		}
		extractor = g
	} else {
		log.Printf("ingest: GEMINI_API_KEY not set; extraction disabled (/api/ingest will 500)")
	}

	poster := digest.NewHTTPPoster(cfg.SlackWebhookURL)
	h := ingest.New(cfg, store.New(database), extractor, poster)
	// Clean tracking redirects out of deal URLs before they are stored
	// (best-effort; see internal/resolve).
	h.Resolver = resolve.New()
	// The Gemini extractor doubles as the page-text enricher for deadlines the
	// page's structured data doesn't state, and as the digest condenser that
	// collapses each digest to per-source essentials; Disabled does neither, so
	// both stay off without an API key.
	if en, ok := extractor.(extract.PageEnricher); ok {
		h.Enricher = en
	}
	if co, ok := extractor.(digest.Condenser); ok {
		h.Condenser = co
	}
	return h
}

// newUsersFromEnv builds the account handlers over the shared database and, when
// a session Manager is present, wires session.OnLogin to upsert the user row at
// login (sending a brand-new user to /settings/?welcome=1). Returns nil when
// there is no database, so /api/me falls back to the session's cookie profile.
func newUsersFromEnv(database *sql.DB, sess *session.Manager) *users.Handlers {
	if database == nil {
		return nil
	}
	st := store.New(database)
	if sess != nil {
		sess.OnLogin = func(ctx context.Context, p session.Profile) (string, error) {
			_, isNew, err := users.Upsert(ctx, st, p)
			if err != nil {
				return "", err
			}
			if isNew {
				return "/settings/?welcome=1", nil
			}
			return "", nil
		}
	}
	return users.New(st, newAdminPosterFromEnv())
}

// newAdminPosterFromEnv builds the admin-channel Slack poster for deletion-request
// notifications, or nil when SLACK_WEBHOOK_FOR_ADMIN_CHANNEL is unset (the
// deletion request still records its durable row; only the ping is skipped).
func newAdminPosterFromEnv() digest.WebhookPoster {
	if os.Getenv("SLACK_WEBHOOK_FOR_ADMIN_CHANNEL") == "" {
		return nil
	}
	return digest.NewHTTPPoster(os.Getenv("SLACK_WEBHOOK_FOR_ADMIN_CHANNEL"))
}

// migrateRequested reports whether the process was asked to apply migrations and
// exit (server -migrate), used by the Fly release_command. It scans args rather
// than using the flag package so it stays correct however the container
// ENTRYPOINT composes with the release command.
func migrateRequested(args []string) bool {
	for _, a := range args {
		if a == "-migrate" || a == "--migrate" {
			return true
		}
	}
	return false
}

// runMigrate opens DATABASE_URL, applies the embedded migrations, and returns.
// Invoked via `server -migrate` from the Fly release_command; it needs only
// DATABASE_URL (not the rest of the ingest config). Any failure is fatal, which
// makes the release step — and therefore the deploy — fail loudly.
func runMigrate(ctx context.Context) {
	dbURL := db.ResolveURL(os.Getenv("DATABASE_URL"), os.Getenv("TURSO_AUTH_TOKEN"))
	if dbURL == "" {
		log.Fatalf("migrate: DATABASE_URL is empty")
	}
	database, err := db.Open(dbURL)
	if err != nil {
		log.Fatalf("migrate: db open: %s", db.RedactToken(err.Error(), dbURL))
	}
	defer func() { _ = database.Close() }()
	if err := db.Migrate(ctx, database); err != nil {
		log.Fatalf("migrate: %s", db.RedactToken(err.Error(), dbURL))
	}
	log.Printf("migrate: schema up to date")
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
// staticRoot. Auth wiring for /api/me depends on what's configured:
//   - sess != nil: the /auth/* routes are mounted and /api/me is gated by the
//     first-party cookie session (the production path).
//   - else v != nil: /api/me is gated by the legacy Bearer/JWKS middleware.
//   - else: /api/me falls through to the /api/* JSON 404.
//
// Split out from main so tests can build a server against fixtures.
func newServer(staticRoot string, v *auth.Verifier, sess *session.Manager, ing *ingest.Handlers, usr *users.Handlers) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Canonical host: 308-redirect every www.* request to the bare apex host,
	// preserving scheme, path and query (e.g. https://www.codeselfstudy.com/x/
	// -> https://codeselfstudy.com/x/). Runs in the Pre phase so it fires before
	// routing, for every path and method. Echo's NonWWWRedirect defaults to 301;
	// force 308 so the method/body survive, matching the statuses in
	// redirects.go. c.Scheme() reads X-Forwarded-Proto, so behind Fly's
	// force_https proxy the redirect target stays https.
	e.Pre(middleware.NonWWWRedirectWithConfig(middleware.RedirectConfig{
		Code: http.StatusPermanentRedirect,
	}))

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

	// Authenticated API routes go before the /api/* catchall so the concrete
	// paths win over the JSON-404 fallback. The cookie session (when set up)
	// owns /auth/* and /api/me; otherwise the legacy Bearer path gates /api/me.
	switch {
	case sess != nil:
		sess.Register(e)
		api := e.Group("/api", sess.Middleware())
		// With the account DB present the users handlers own /api/me (DB-backed,
		// with the username) plus the settings routes; without it, the session's
		// cookie-profile /api/me keeps the site working.
		if usr != nil {
			usr.Register(api)
		} else {
			api.Match(getOrHead, "/me", sess.HandleMe)
		}
	case v != nil:
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

		// Service worker scripts live at fixed, unhashed URLs, so any cached
		// copy would pin an outdated worker in browsers and at the CDN — and a
		// 404 can be negatively cached too, hiding a later kill-switch deploy.
		// Set the header by request path, before resolution, so both the file
		// and its 404 fallback are always revalidated. Hashed /_astro/* assets
		// are immutable and intentionally left cacheable.
		if isServiceWorkerScript(clean) {
			c.Response().Header().Set("Cache-Control", "no-cache")
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

// isServiceWorkerScript reports whether p is one of the root-scope service
// worker scripts shipped as cache kill switches (apps/web/public/sw.js and
// service-worker.js). They must be served with Cache-Control: no-cache —
// unlike the content-hashed assets under /_astro, their URLs never change, so
// any cached copy would keep an outdated worker alive.
func isServiceWorkerScript(p string) bool {
	return p == "/sw.js" || p == "/service-worker.js"
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
