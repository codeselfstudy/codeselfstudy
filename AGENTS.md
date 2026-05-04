# Repository Guidelines

## Project Structure & Module Organization

This is a Bun workspace with two apps:

- `apps/web/` — TanStack Start (Vite + React 19 + TypeScript). Prerenders to `apps/web/.output/public/` at build time.
  - `src/routes/` — TanStack Router file-based routes.
  - `src/components/` — reusable UI components.
  - `src/content/` and `src/data/` — site content and data helpers.
  - `src/db/` — Drizzle schema and access for the web side. **Drizzle is the source of truth for the schema**; the Go backend reads/writes via plain SQL.
  - `src/lib/` — shared utilities, including `useApiFetch` for authenticated calls into the Go API.
  - `src/integrations/workos/` — WorkOS AuthKit provider; sign-in flow stays client-side.
  - `public/` — static assets passed through to the prerendered output.
  - `test/` — Vitest setup and MSW handlers.
- `apps/api/` — Go + Echo backend.
  - `main.go` — Echo bootstrap, static serving, route wiring.
  - `internal/auth/` — WorkOS JWKS verifier + Echo middleware (validates bearer tokens against WorkOS-issued JWTs).
  - `internal/db/` — SQLite access via `modernc.org/sqlite` (pure Go, no CGO). Mirrors the schema from `apps/web/src/db/schema.ts`.
  - `static/` — populated at build time by `just build` from `apps/web/.output/public/`. Gitignored.
- `manual/` — mdBook documentation. Write documentation for any changed code. Mermaid is available.
- `mockups/` — HTML mockups, served via `just mockups`.

## Build, Test, and Development Commands

Never commit code unless I specifically request it.

If I ask a question, do not automatically implement a solution. Just answer the question, and I will let you know if I would like you to implement it.

Use Bun. Do not use npm, pnpm, or yarn commands. Bun handles the entire JS pipeline (install, run, build) — Node.js is not required at any point.

From the repo root:

- `just dev` — runs the Go API (`:8080`) and the web dev server (`:7001`) together. The Vite dev server proxies `/api` and `/ws` to the Go server.
- `just dev-web` — web only.
- `just dev-api` — Go API only.
- `just build` — prerenders the web app and mirrors `.output/public/` into `apps/api/static/` so the Go binary serves it.
- `just test` — Go race tests + Vitest.
- `just deploy` — `just build` then `fly deploy`. The web is built locally and the prebuilt `dist` ships in the Docker build context (Fly's remote builder doesn't re-run `bun install`).
- `just db_generate|db_migrate|db_push|db_pull|db_studio` — Drizzle tasks (run inside `apps/web/`).
- `just find_licenses` — list embedded `LICENSE.md` files.

Inside `apps/web/`:

- `bun run dev|build|preview|test|test:watch|test:coverage|lint`.

Inside `apps/api/`:

- `go run .` — start the API. Requires `apps/api/static/` to exist (run `just build` first if you want it to serve the web pages).
- `go test -race ./...` — run all Go tests.
- `go vet ./...`.

Do NOT add the agent name (e.g. Claude, Generated with Claude Code, Co-Authored-By Claude) anywhere in commit messages, PR descriptions, or other Git/GitHub messages.

## Coding Style & Naming Conventions

- TypeScript + React (frontend), Go (backend). Module syntax in JS is ESM.
- Indentation is 2 spaces (follow existing files). Go follows `gofmt`.
- React components and hooks use `PascalCase`/`camelCase`.
- Routes follow the TanStack conventions in `apps/web/src/routes/`.
- Formatting and linting in JS are handled by `prettier`, `eslint`, and `stylelint`. CI runs them without `--fix`; local `lint` script auto-fixes.

## Testing Guidelines

- Web framework: Vitest (`apps/web/vitest.config.ts`). Tests live next to the module under test as `*.test.ts` / `*.test.tsx`.
- API framework: stdlib `testing` + `httptest`. All Go code requires test coverage — this is a hard requirement of the stack experiment.
- Run `just test` before opening a PR. CI runs the same plus `go vet`.
- Web coverage: `apps/web/src/lib/**`, `apps/web/src/data/**`, `apps/web/src/hooks/**`, and `apps/web/src/env.ts` enforce 100% (lines/functions/statements) per `vitest.config.ts`. Follow the same when adding code in those paths.

## Commit & Pull Request Guidelines

- Commit messages in history are short, sentence-case summaries (often with a period). Follow that style.
- PRs include: a concise summary, testing notes (`just test`), and screenshots for UI changes.
- Default merge mode is "Create a merge commit" — preserves per-PR history. Don't rebase the PR branch on the integration branch unless explicitly asked.

## Configuration & Environment

- Copy `env.local.example` to `.env.local` at the repo root and fill in the values.
- Required vars:
  - `VITE_WORKOS_CLIENT_ID`, `VITE_WORKOS_API_HOSTNAME` — used by the web client and re-used by the Go API as fallbacks.
  - `WORKOS_API_KEY` — server-only WorkOS key, validated by `apps/web/src/env.ts`.
  - `DATABASE_URL` — SQLite path or libsql URL.
- Optional vars: `TURSO_AUTH_TOKEN` (when `DATABASE_URL` is a remote libsql URL), `WORKOS_CLIENT_ID` / `WORKOS_API_HOSTNAME` (override the `VITE_`-prefixed values for the Go API).
- Client-side variables must be prefixed with `VITE_` to be exposed to the browser.
- Env propagation: the root `package.json` scripts and the `justfile` pass `--env-file=.env.local` into `bun run`, since Bun's `--filter` `cd`s into `apps/web/` and would otherwise miss the root file. `just dev-api` sources `.env.local` into the Go process the same way.
- `/api/me` is disabled if WorkOS env is missing; `/api/todos` is disabled if `DATABASE_URL` is missing. Static serving and `/healthz` always work — useful for barebones smoke tests.
- The schema lives in Drizzle (`apps/web/src/db/schema.ts`). The Go side mirrors it in `apps/api/internal/db/db.go` (`Schema` constant) for tests; production migrations stay JS-side via `just db_*`.
- Don't read files or directories ending in `.bak` or that are blocked by `.gitignore`.
