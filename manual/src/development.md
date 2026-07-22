# Development

Notes for working on the codebase day-to-day.

## Running the app

```bash
just dev
```

Boots two servers in parallel:

- `apps/web/` (Astro dev server) on `http://localhost:7001`.
- `apps/api/` (Go + Echo) on `http://localhost:8080`.

The two servers run independently in this phase — there is no dev proxy, since the static site does not call the API yet. Build the site (`just build`) and run the Go server to exercise the production URL behavior (redirects, trailing-slash canonicalization) end to end.

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
