# List available recipes
default:
  @just --list

# Run the web (port 7001) and api (port 8080) dev servers together. The Vite
# dev server proxies /api and /ws to the Go server.
dev:
  just dev-api & just dev-web && wait

# Web (TanStack Start) dev server only
dev-web:
  bun run --env-file=.env.local --filter web dev

# Go API dev server only
dev-api:
  cd apps/api && go run .

# Build the prerendered web app, then mirror dist into apps/api/static so the
# Go binary can serve it locally and so the Dockerfile picks it up directly.
build: clean
  @echo '\nBuilding the project...\n'
  bun run --env-file=.env.local --filter web build
  /bin/rm -rf apps/api/static
  /bin/cp -R apps/web/.output/public apps/api/static
  @echo '\nDone with building.\n'

# Run unit tests across the whole repo (Go race tests + Vitest)
test:
  cd apps/api && go test -race ./...
  bun run --filter web test

# Run web tests in watch mode
test-watch:
  bun run --filter web test:watch

# Run web tests with coverage
test-coverage:
  bun run --filter web test:coverage

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

# Generate a database migration (Drizzle, web workspace)
db_generate:
  cd apps/web && bun run --bun drizzle-kit generate

# Migrate the database
db_migrate:
  cd apps/web && bun run --bun drizzle-kit migrate

# Open the Drizzle studio
db_studio:
  cd apps/web && bun run --bun drizzle-kit studio

# Push schema changes directly (prototyping)
db_push:
  cd apps/web && bun run --bun drizzle-kit push

# Introspect the database to generate schema files
db_pull:
  cd apps/web && bun run --bun drizzle-kit pull

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

# Remove build artifacts (web .output/.tanstack, api ./static, root node_modules cache)
clean:
  @echo '\nCleaning project...\n'
  ./scripts/clean.sh
  @echo '\nDone with cleaning.\n'
