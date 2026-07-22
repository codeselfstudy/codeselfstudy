# Development

Notes for working on the codebase day-to-day.

## Running the app

```bash
just dev
```

Boots two servers in parallel:

- `apps/web/` (Astro dev server) on `http://localhost:7001`.
- `apps/api/` (Go + Echo) on `http://localhost:8080`.

The two servers run independently — there is no dev proxy. The navbar island calls `/api/me` as a same-origin request, so under `just dev` it targets the Astro dev server on `:7001`, which has no API, and 404s. The authenticated API works only when the Go binary serves the site and `/api/*` from one origin: run `just build`, then start the Go server. Use that Go-served build to exercise production URL behavior (redirects, trailing-slash canonicalization) and the authenticated API end to end.

To run only one side:

```bash
just dev-web   # Astro only
just dev-api   # Go only — sources .env.local so /api/me mounts
```

## Updating dependencies

```bash
# JS dependencies (apps/web)
bun update --interactive

# Go dependencies (apps/api)
cd apps/api && go get -u ./... && go mod tidy
```

## Running tests

```bash
just test            # Go race tests + Vitest
just test-watch      # Vitest in watch mode (web only)
just test-coverage   # Vitest with coverage report
```

Backend changes require Go test coverage — that's a hard rule of the stack. Add tests under `apps/api/` next to the code under test (`*_test.go`).

## Linting and formatting

```bash
just format    # Prettier across the repo
just lint      # ESLint + Stylelint on the web side
```

`lefthook` runs Prettier on staged files in `pre-commit` and `just test` + lint in `pre-push`.

## See available recipes

```bash
just
```
