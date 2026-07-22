# Architecture

The repo is a Bun workspace with three apps: an Astro frontend, a Go + Echo backend, and a Cloudflare Worker that receives forwarded email. In production a single Go binary fronts the site and the JSON API — there is no JavaScript runtime on the server — while the Worker runs on Cloudflare.

## Top-level shape

```
codeselfstudy/
├── apps/
│   ├── web/                    Astro (SSG, React islands)
│   │   ├── src/                pages, layouts, components, lib
│   │   ├── public/             static assets
│   │   ├── test/               Vitest setup
│   │   └── dist/               prerendered HTML/CSS/JS (build output)
│   ├── api/                    Go + Echo backend
│   │   ├── main.go             Echo bootstrap, static serving, route wiring
│   │   ├── cmd/migrate/        thin goose CLI for ad-hoc migration runs
│   │   ├── internal/auth/      WorkOS JWKS verifier + Echo middleware
│   │   ├── internal/db/        SQLite / Turso access (driver by URL scheme)
│   │   ├── internal/db/migrations/   goose .sql migrations (embedded)
│   │   ├── internal/ingest/    /api/ingest + /api/admin/digest handlers, config
│   │   ├── internal/store/     emails/deals/digests persistence
│   │   ├── internal/mailparse/ internal/htmltext/  MIME → normalized text
│   │   ├── internal/extract/   Gemini deal extraction
│   │   ├── internal/digest/    Slack Block Kit digest + HTTP poster
│   │   └── static/             populated at build time from web's dist/
│   └── email_receiver/         Cloudflare Worker (TS): email → /api/ingest
├── manual/                     this mdBook
├── mockups/                    HTML mockups
├── Dockerfile                  multi-stage: Go build + distroless runtime
├── fly.toml                    256 MB shared-cpu-1x
└── justfile                    dev / build / test / deploy recipes
```

## Request flow in production

1. The Go binary listens on `:8080`.
2. `/healthz` → 204 (Fly health check).
3. `/api/*` → JSON handlers. `/api/me` is gated by the WorkOS JWKS middleware; `POST /api/ingest` and `POST /api/admin/digest` (the email pipeline) are gated by the `INGEST_TOKEN` bearer. Unknown `/api/*` paths return a JSON 404, never the static fallback.
4. `/ws` → reserved for the future WebSocket hub.
5. Anything else → static handler. It first checks the legacy redirect map (`apps/api/redirects.go`) and emits a 308 if the path matches. Then, for an extensionless path with no trailing slash that resolves to a directory index, it emits a 301 to the trailing-slash form (the site is built with `trailingSlash: "always"`). Otherwise it resolves `/foo` against `static/foo`, then `static/foo/index.html`, then `static/foo.html`. If nothing matches it serves `static/404.html` with a 404 status.

`http.FileServer` is **not** used because it commits a 404 to the wire before Echo's error handler can run. The custom handler in `apps/api/main.go` owns that path.

## Auth

- **Client:** the Astro navbar mounts a WorkOS AuthKit sign-in control (`@workos-inc/authkit-react`) as a browser-only island, so `Layout.astro` renders `Navbar` with `client:only="react"`. Signing in yields a WorkOS access token that `apiFetch` (`apps/web/src/lib/api.ts`) attaches as a Bearer credential to the gated API — the navbar reads `/api/me` to show the signed-in user.
- **Server:** `apps/api/internal/auth/jwks.go` fetches the WorkOS JWKS on startup and refreshes periodically. `apps/api/internal/auth/middleware.go` validates signature, issuer, and expiry on every protected request. Validated claims are stashed in the Echo context for handlers to read.

The Go-side auth is opt-in: if `WORKOS_CLIENT_ID` / `WORKOS_API_HOSTNAME` (or their `VITE_`-prefixed fallbacks) are missing, `/api/me` is not mounted. Static serving and `/healthz` keep working — useful for barebones smoke tests.

## Database

- Pure-Go drivers only, so the distroless image stays static: `modernc.org/sqlite` for local files and `:memory:`, and `tursodatabase/libsql-client-go` for remote Turso. `internal/db.Open` selects the driver by `DATABASE_URL` scheme (`libsql://` / `http(s)` / `ws(s)` → Turso; anything else → SQLite). The Turso auth token rides in the URL as `?authToken=…`.
- **Schema is owned by Go.** Migrations live in `apps/api/internal/db/migrations/` as goose `.sql` files, are embedded into the binary via `//go:embed`, and apply on every startup. Goose tracks state in its own table, so re-running is a no-op.
- For dev: `just db_migrate`, `just db_status`, `just db_down`, `just db_create <name>` wrap a small CLI in `apps/api/cmd/migrate`.

## Email → deals → Slack

An opt-in pipeline turns forwarded "deals" newsletters into a Slack digest:

1. `apps/email_receiver` is a Cloudflare Worker wired to Cloudflare Email Routing. On each message it POSTs the raw RFC822 bytes to the Go server's `POST /api/ingest` (bearer `INGEST_TOKEN`, `Content-Type: message/rfc822`). There is no archive mailbox: an oversize message is rejected with `setReject()`, and a persistent POST failure propagates so Cloudflare fails the delivery and the sender's MTA retries.
2. `internal/ingest` reads the body (≤25 MB), parses the MIME (`internal/mailparse` + `internal/htmltext`), stores the email (`internal/store`, idempotent on message-id), extracts deals with the Gemini API (`internal/extract`), upserts them (dedup by sender registrable-domain + normalized title), and — best effort — posts a Block Kit digest to Slack (`internal/digest`) at most once per `DIGEST_INTERVAL`. `POST /api/admin/digest` forces one.
3. The pipeline is enabled only when `DATABASE_URL` and `INGEST_TOKEN` are set; otherwise the server runs static-only. Extraction needs `GEMINI_API_KEY` (empty ⇒ `/api/ingest` returns 500). The Slack webhook is `SLACK_WEBHOOK_FOR_DEALS_CHANNEL`.

Secrets (`DATABASE_URL` with its authToken, `INGEST_TOKEN`, `GEMINI_API_KEY`, `SLACK_WEBHOOK_FOR_DEALS_CHANNEL`) are Fly secrets; `INGEST_TOKEN` must match the value set on the Worker with `wrangler secret put`. Non-secret `GEMINI_MODEL` / `DIGEST_INTERVAL` / `REPOST_AFTER` live in `fly.toml`'s `[env]` block.

> Note: Cloudflare does not document what a thrown `email()` handler does. The Worker relies on a thrown handler producing a retryable delivery failure (so a transient Go outage → sender re-delivers); confirm this on the first real deploy. The deterministic oversize case already uses the documented `setReject()`.

## Build and deploy

- `bun run build` (or `just build`) prerenders all routes to `apps/web/dist/`.
- `just build` mirrors that directory into `apps/api/static/` so `go run .` serves the same content locally.
- The Dockerfile copies the prebuilt `apps/web/dist` into the Go build stage. Fly's remote builder never runs `bun install`.
- `just deploy` chains `just build` and `fly deploy`.

See [Deployment](./deployment.md) for image and Fly details.
