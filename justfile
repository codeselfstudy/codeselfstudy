# List available recipes
default:
  @just --list

# Run the web (port 7001) and api (port 8080) dev servers together. The Vite
# dev server proxies /api and /ws to the Go server.
dev:
  just dev-api & just dev-web && wait

# Web (Astro) dev server only
dev-web:
  bun run --env-file=.env.local --filter web dev

# Go API dev server only. Loads .env.local into the shell so /api/me and
# /api/todos can wire themselves up — Go has no built-in .env support.
# Skipped silently when the file is absent (smoke-test mode).
dev-api:
  set -a; [ -f .env.local ] && . ./.env.local; set +a; cd apps/api && go run .

# Cloudflare Worker (email_receiver) dev server via wrangler
dev_worker:
  bun run --filter email_receiver dev

# Build the prerendered web app, then mirror dist into apps/api/static so the
# Go binary can serve it locally and so the Dockerfile picks it up directly.
build: clean
  @echo '\nBuilding the project...\n'
  bun run --env-file=.env.local --filter web build
  /bin/rm -rf apps/api/static
  /bin/cp -R apps/web/dist apps/api/static
  @echo '\nDone with building.\n'

# Run unit tests across the whole repo (Go race tests + Vitest + Worker)
test:
  cd apps/api && go test -race ./...
  bun run --filter web test
  bun run --filter email_receiver test

# Worker (email_receiver) tests only: tsc typecheck + bun test
test_worker:
  bun run --filter email_receiver test

# Run web tests in watch mode
test-watch:
  bun run --filter web test:watch

# Run web tests with coverage
test-coverage:
  bun run --filter web test:coverage

# Build the site, then smoke-test it end to end through the Go server
# (pages, redirects, trailing-slash canonicalization, 404, sitemap).
smoke_test: build
  ./scripts/smoke.sh

# Format the whole repo with prettier
format:
  bun run format

# Lint web (eslint + stylelint)
lint:
  bun run --filter web lint

# Build and deploy to Fly. We build locally so the remote builder doesn't
# re-run `bun install` (~10min on a cold cache).
deploy: build
  fly deploy

# Tail Fly logs for the deployed app
logs:
  fly logs

# Show Fly machine status (memory, region, state)
status:
  fly status

# SSH into a live Fly machine
ssh:
  fly ssh console

# Deploy the Cloudflare Worker (email_receiver) via wrangler
deploy_worker:
  bun run --filter email_receiver deploy

# Tail the deployed Cloudflare Worker's logs
tail_worker:
  cd apps/email_receiver && bunx wrangler tail

# Apply pending migrations against DATABASE_URL. The server also runs this
# on startup, so this is mainly for ad-hoc dev runs.
db_migrate:
  set -a; [ -f .env.local ] && . ./.env.local; set +a; cd apps/api && go run ./cmd/migrate up

# Show migration status (which versions have/haven't been applied)
db_status:
  set -a; [ -f .env.local ] && . ./.env.local; set +a; cd apps/api && go run ./cmd/migrate status

# Roll back the most recent migration. Dev only — never against shared data.
db_down:
  set -a; [ -f .env.local ] && . ./.env.local; set +a; cd apps/api && go run ./cmd/migrate down

# Scaffold a new migration in apps/api/internal/db/migrations/. Pass a snake_case name.
db_create name:
  cd apps/api && go run ./cmd/migrate create {{name}}

# Find all the license files
find_licenses:
  find . -name "LICENSE.md" -not -path "*/node_modules/*"

# View HTML mockups
mockups:
  @echo "\n\nServing mockups at http://localhost:5555\n\n"
  cd mockups && python3 -m http.server 5555

# View the manual in the browser
manual:
  cd manual && mdbook serve -p 8001 --open

# Remove build artifacts (web dist, api ./static, root node_modules cache)
clean:
  @echo '\nCleaning project...\n'
  ./scripts/clean.sh
  @echo '\nDone with cleaning.\n'
