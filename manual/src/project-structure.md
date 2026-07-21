# Project Structure

This is a Bun workspace with two apps. Top-level shape:

- `apps/web/`: Astro frontend (SSG, React islands). Prerenders to `apps/web/dist/`.
- `apps/api/`: Go + Echo backend. Owns the database schema and serves the prerendered site at runtime.
- `manual/`: this mdBook documentation.
- `mockups/`: HTML mockups, served via `just mockups`.

Inside `apps/web/src/`:

- `pages/`: file-based routes (Astro `.astro` pages).
- `layouts/`: shared page shell (`Layout.astro` — `<head>`, nav, footer).
- `components/`: reusable UI building blocks, including the React navbar island.
- `components/copyrighted/`: licensed page copy (see its `LICENSE.md`).
- `lib/`: shared utilities (`metadata`, `cn`).
- `styles/`: the global stylesheet.

Inside `apps/api/`:

- `main.go`: Echo bootstrap, static serving, route wiring.
- `cmd/migrate/`: small CLI around goose for ad-hoc migration runs.
- `internal/auth/`: WorkOS JWKS verifier + Echo middleware.
- `internal/db/`: SQLite access (`modernc.org/sqlite`).
- `internal/db/migrations/`: goose `.sql` migrations, embedded into the binary and applied on startup.
- `static/`: populated at build time from `apps/web/dist/`. Gitignored.

Generated files:

- `apps/web/.astro/` holds Astro-generated content and type files (from `astro sync` / `astro check`). Gitignored.
