# Architecture

The repo is a Bun workspace with an Astro frontend and a Go + Echo backend. In production a single Go binary fronts both — there is no JavaScript runtime on the server.

## Top-level shape

```
codeselfstudy/
├── apps/
│   ├── web/                    Astro (SSG, React islands)
│   │   ├── src/                pages, layouts, components, lib
│   │   ├── public/             static assets
│   │   ├── test/               Vitest setup
│   │   └── dist/               prerendered HTML/CSS/JS (build output)
│   └── api/                    Go + Echo backend
│       ├── main.go             Echo bootstrap, static serving, route wiring
│       ├── cmd/migrate/        thin goose CLI for ad-hoc migration runs
│       ├── internal/auth/      WorkOS JWKS verifier + Echo middleware
│       ├── internal/db/        SQLite access (modernc.org/sqlite)
│       ├── internal/db/migrations/   goose .sql migrations (embedded)
│       └── static/             populated at build time from web's dist/
├── manual/                     this mdBook
├── mockups/                    HTML mockups
├── Dockerfile                  multi-stage: Go build + distroless runtime
├── fly.toml                    256 MB shared-cpu-1x
└── justfile                    dev / build / test / deploy recipes
```

## Request flow in production

1. The Go binary listens on `:8080`.
2. `/healthz` → 204 (Fly health check).
3. `/api/*` → JSON handlers, gated by the WorkOS JWKS middleware where applicable. Unknown `/api/*` paths return a JSON 404, never the static fallback.
4. `/ws` → reserved for the future WebSocket hub.
5. Anything else → static handler. It first checks the legacy redirect map (`apps/api/redirects.go`) and emits a 308 if the path matches. Then, for an extensionless path with no trailing slash that resolves to a directory index, it emits a 301 to the trailing-slash form (the site is built with `trailingSlash: "always"`). Otherwise it resolves `/foo` against `static/foo`, then `static/foo/index.html`, then `static/foo.html`. If nothing matches it serves `static/404.html` with a 404 status.

`http.FileServer` is **not** used because it commits a 404 to the wire before Echo's error handler can run. The custom handler in `apps/api/main.go` owns that path.

## Auth

- **Client:** deferred to the API phase. The Astro build currently ships no sign-in UI; the client-side WorkOS integration (and the bearer-token fetch wrapper) will return when the API surface is built out.
- **Server:** `apps/api/internal/auth/jwks.go` fetches the WorkOS JWKS on startup and refreshes periodically. `apps/api/internal/auth/middleware.go` validates signature, issuer, and expiry on every protected request. Validated claims are stashed in the Echo context for handlers to read.

The Go-side auth is opt-in: if `WORKOS_CLIENT_ID` / `WORKOS_API_HOSTNAME` (or their `VITE_`-prefixed fallbacks) are missing, `/api/me` is not mounted. Static serving and `/healthz` keep working — useful for barebones smoke tests.

## Database

- Pure-Go SQLite via `modernc.org/sqlite` (no CGO) so the distroless image stays static.
- **Schema is owned by Go.** Migrations live in `apps/api/internal/db/migrations/` as goose `.sql` files, are embedded into the binary via `//go:embed`, and apply on every startup. Goose tracks state in its own table, so re-running is a no-op.
- For dev: `just db_migrate`, `just db_status`, `just db_down`, `just db_create <name>` wrap a small CLI in `apps/api/cmd/migrate`.
- Remote Turso (`libsql://`) is a planned follow-up; the current `Open()` rejects libsql/http schemes up front so misconfigurations fail fast.

## Build and deploy

- `bun run build` (or `just build`) prerenders all routes to `apps/web/dist/`.
- `just build` mirrors that directory into `apps/api/static/` so `go run .` serves the same content locally.
- The Dockerfile copies the prebuilt `apps/web/dist` into the Go build stage. Fly's remote builder never runs `bun install`.
- `just deploy` chains `just build` and `fly deploy`.

See [Deployment](./deployment.md) for image and Fly details.
