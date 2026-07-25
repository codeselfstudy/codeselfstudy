# Project Structure

This is a Bun workspace with three apps. Top-level shape:

- `apps/web/`: Astro frontend (SSG, React islands). Prerenders to `apps/web/dist/`.
- `apps/api/`: Go + Echo backend. Owns the database schema and serves the prerendered site at runtime.
- `apps/email_receiver/`: Cloudflare Worker (TypeScript) that receives forwarded "deals" email and POSTs it to the Go server's `/api/ingest`.
- `manual/`: this mdBook documentation.
- `mockups/`: HTML mockups, served via `just mockups`.

Inside `apps/web/src/`:

- `pages/`: file-based routes (Astro `.astro` pages).
- `layouts/`: shared page shell (`Layout.astro` — `<head>`, nav, footer).
- `components/`: reusable UI building blocks, including the React navbar island.
- `components/auth/`: the navbar's `SignInButton` and `UserMenu`.
- `components/settings/`: the settings-page form island.
- `components/ui/`: small shadcn-style primitives (button, dropdown, sheet).
- `components/copyrighted/`: licensed page copy (see its `LICENSE.md`).
- `lib/`: shared utilities (`metadata`, `utils` with `cn`, `authHint`, `minify`).
- `styles/`: the global stylesheet.

Inside `apps/api/`:

- `main.go`: Echo bootstrap, static serving, route wiring.
- `cmd/migrate/`: small CLI around goose for ad-hoc migration runs.
- `internal/session/`: server-side WorkOS sign-in (`/auth/*`), sealed cookie session, session middleware.
- `internal/users/`: account rows and the session-gated `/api/me` + settings handlers.
- `internal/auth/`: WorkOS JWKS verifier + Echo middleware (legacy Bearer fallback).
- `internal/db/`: SQLite / Turso access (driver chosen by `DATABASE_URL` scheme).
- `internal/db/migrations/`: goose `.sql` migrations, embedded into the binary and applied by the deploy's release command (on boot for a local SQLite DB).
- `internal/ingest/`: the `/api/ingest` + `/api/admin/digest` handlers, bearer-token middleware, and pipeline config.
- `internal/store/`: emails/deals/digests persistence and the once-per-interval digest claim.
- `internal/mailparse/`, `internal/htmltext/`: raw MIME → normalized text.
- `internal/extract/`: Gemini deal extraction.
- `internal/digest/`: Slack Block Kit rendering + HTTP poster.
- `static/`: populated at build time from `apps/web/dist/`. Gitignored.

Inside `apps/email_receiver/` (Cloudflare Worker):

- `src/index.ts`: the Cloudflare `email()` handler (runtime wiring).
- `src/lib.ts`: `handleEmail` — POSTs the raw email to `/api/ingest` with retry (no archive forward).
- `wrangler.jsonc`, `tsconfig.json`, `eslint.config.js`, `bunfig.toml`: per-app config.

Generated files:

- `apps/web/.astro/` holds Astro-generated content and type files (from `astro sync` / `astro check`). Gitignored.
