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
│   │   ├── internal/session/   server-side WorkOS sign-in + cookie session
│   │   ├── internal/users/     account rows, /api/me + settings handlers
│   │   ├── internal/auth/      WorkOS JWKS verifier + Echo middleware (fallback)
│   │   ├── internal/db/        SQLite / Turso access (driver by URL scheme)
│   │   ├── internal/db/migrations/   goose .sql migrations (embedded)
│   │   ├── internal/ingest/    /api/ingest + /api/admin/digest handlers, config
│   │   ├── internal/store/     emails/deals/digests persistence
│   │   ├── internal/mailparse/ internal/htmltext/  MIME → normalized text
│   │   ├── internal/extract/   Gemini deal extraction
│   │   ├── internal/resolve/   deal-URL cleanup (tracking redirects → canonical)
│   │   ├── internal/expiry/    expiration date + price from a deal page's structured data
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
3. `/auth/*` → the server-side sign-in flow (login redirect, WorkOS code exchange, logout). `/api/*` → JSON handlers. `/api/me` and the settings routes are gated by the session-cookie middleware; `POST /api/ingest` and `POST /api/admin/digest` (the email pipeline) are gated by the `INGEST_TOKEN` bearer. Unknown `/api/*` paths return a JSON 404, never the static fallback.
4. `/ws` → reserved for the future WebSocket hub.
5. Anything else → static handler. It first checks the legacy redirect map (`apps/api/redirects.go`) and emits a 308 if the path matches. Then, for an extensionless path with no trailing slash that resolves to a directory index, it emits a 301 to the trailing-slash form (the site is built with `trailingSlash: "always"`). Otherwise it resolves `/foo` against `static/foo`, then `static/foo/index.html`, then `static/foo.html`. If nothing matches it serves `static/404.html` with a 404 status.

`http.FileServer` is **not** used because it commits a 404 to the wire before Echo's error handler can run. The custom handler in `apps/api/main.go` owns that path.

## Auth

Sign-in is entirely server-side — the browser bundle reads no WorkOS values.

- **Server:** `apps/api/internal/session` owns `GET /auth/login` (redirect to WorkOS's hosted sign-in), `GET /auth/callback` (code exchange via `WORKOS_API_KEY`), and `GET /auth/logout`. The session lives in a sealed first-party cookie (encrypted with `WORKOS_COOKIE_PASSWORD`); a session middleware gates `/api/*`. When `DATABASE_URL` is set, `apps/api/internal/users` owns the account routes on top of that session: `GET /api/me` (DB-backed, with the username), `GET`/`PATCH /api/settings`, and `POST /api/settings/delete-request`. Without the DB, the session's cookie-profile `/api/me` keeps the site working.
- **Client:** the navbar island renders `SignInButton` (which navigates to `/auth/login`) or `UserMenu` from `apps/web/src/components/auth/`. To paint the right control on the first frame it reads the JavaScript-readable `css_auth` hint cookie that the server sets beside the session cookie, plus a localStorage-cached username/avatar (`apps/web/src/lib/authHint.ts`), then confirms against `/api/me` (a same-origin fetch that sends the cookie).
- **Fallback:** the older Bearer/JWKS path (`apps/api/internal/auth`: JWKS fetch + refresh, signature/expiry validation) still exists and gates `/api/me` when the session config is absent but a verifier is configured.

The Go-side auth is opt-in: sign-in needs all five WorkOS variables (`WORKOS_CLIENT_ID`, `WORKOS_API_HOSTNAME`, `WORKOS_API_KEY`, `WORKOS_COOKIE_PASSWORD`, `APP_BASE_URL`) — with any of them missing the `/auth/*` routes stay off and the server logs which one is absent. Static serving and `/healthz` keep working — useful for barebones smoke tests.

## Database

- Pure-Go drivers only, so the distroless image stays static: `modernc.org/sqlite` for local files and `:memory:`, and `tursodatabase/libsql-client-go` for remote Turso. `internal/db.Open` selects the driver by `DATABASE_URL` scheme (`libsql://` / `http(s)` / `ws(s)` → Turso; anything else → SQLite). The Turso auth token rides in the URL as `?authToken=…`.
- **Schema is owned by Go.** Migrations live in `apps/api/internal/db/migrations/` as goose `.sql` files and are embedded into the binary via `//go:embed`. A local SQLite database is migrated on boot; a remote Turso database is migrated out of band by `server -migrate`, which Fly runs as the `[deploy] release_command` once per deploy before the new version serves traffic (see [Deployment](./deployment.md#migrations)). Goose tracks state in its own table, so re-running is a no-op.
- For dev: `just db_migrate`, `just db_status`, `just db_down`, `just db_create <name>` wrap a small CLI in `apps/api/cmd/migrate`.

## Email → deals → Slack

An opt-in pipeline turns forwarded "deals" newsletters into a Slack digest:

1. `apps/email_receiver` is a Cloudflare Worker wired to Cloudflare Email Routing. On each message it POSTs the raw RFC822 bytes to the Go server's `POST /api/ingest` (bearer `INGEST_TOKEN`, `Content-Type: message/rfc822`). There is no archive mailbox: an oversize message is rejected with `setReject()`, and a persistent POST failure propagates so Cloudflare fails the delivery and the sender's MTA retries.
2. `internal/ingest` reads the body (≤25 MB), parses the MIME (`internal/mailparse` + `internal/htmltext`), stores the email (`internal/store`, idempotent on message-id), extracts deals with the Gemini API (`internal/extract`), resolves each deal URL to its clean destination (`internal/resolve` follows tracking redirects — HTTP ones and the meta-refresh/JavaScript interstitials some trackers use — and strips tracking params, best-effort with an SSRF guard; a failure keeps the extracted URL), fills a missing expiration or price from the destination page's structured data (`internal/expiry` reads JSON-LD `priceValidUntil` / `availabilityEnds` and offer `price` / `priceCurrency`; an email-stated value is never overwritten), upserts them (dedup by sender registrable-domain + normalized title), and — best effort — posts a Block Kit digest to Slack (`internal/digest`) at most once per `DIGEST_INTERVAL`. `POST /api/admin/digest` forces one (see [Forcing a Slack digest](./deployment.md#forcing-a-slack-digest)).
3. Mail from an address in `APPROVED_FORWARDING_EMAILS` (comma-separated) skips that wait: its digest posts immediately, flushing every queued deal. A sender matches on either the `From:` header or the Worker's `X-Envelope-From` header, so both a hand-composed forward and an auto-forward rule work. Forcing skips only the interval check — the stale-claim guard still holds, so it stays race-safe. The ingest response carries `"forced":true` when the allowlist matched.
4. The pipeline is enabled only when `DATABASE_URL` and `INGEST_TOKEN` are set; otherwise the server runs static-only. Extraction needs `GEMINI_API_KEY` (empty ⇒ `/api/ingest` returns 500). The Slack webhook is `SLACK_WEBHOOK_FOR_DEALS_CHANNEL`.

Secrets (`DATABASE_URL` with its authToken, `INGEST_TOKEN`, `GEMINI_API_KEY`, `SLACK_WEBHOOK_FOR_DEALS_CHANNEL`, `APPROVED_FORWARDING_EMAILS`) are Fly secrets; `INGEST_TOKEN` must match the value set on the Worker with `wrangler secret put`. Non-secret `GEMINI_MODEL` / `DIGEST_INTERVAL` / `REPOST_AFTER` live in `fly.toml`'s `[env]` block.

> Note: Cloudflare does not document what a thrown `email()` handler does. The Worker relies on a thrown handler producing a retryable delivery failure (so a transient Go outage → sender re-delivers); confirm this on the first real deploy. The deterministic oversize case already uses the documented `setReject()`.

## Build and deploy

- `bun run build` (or `just build`) prerenders all routes to `apps/web/dist/`.
- `just build` mirrors that directory into `apps/api/static/` so `go run .` serves the same content locally.
- The Dockerfile copies the prebuilt `apps/web/dist` into the Go build stage. Fly's remote builder never runs `bun install`.
- `just deploy` chains `just build` and `fly deploy`.

See [Deployment](./deployment.md) for image and Fly details.
