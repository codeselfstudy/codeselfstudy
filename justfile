# List scripts
default:
  @just --list

# Cleans out build files and things like .DS_Store files
clean:
  ./scripts/clean.sh

# Run tests
test:
  bun run test

# Run the dev server
dev:
  bun run dev

# Run the build process
build: clean
  bun run build

# Find all the license files
find_licenses:
  find . -name "LICENSE.md" -not -path "*/node_modules/*"

# Deploy to Fly.io using the local .bun-version
deploy:
  if [ -f .bun-version ]; then time fly deploy --build-arg BUN_VERSION="$(cat .bun-version)"; else echo "Error: .bun-version file not found in repository root. Please create it or specify BUN_VERSION another way." >&2; exit 1; fi

# SSH into a live Fly machine
ssh:
  fly ssh console

# Generate a database migration
db_generate:
  bun run --bun drizzle-kit generate

# Migrate the database
db_migrate:
  bun run --bun drizzle-kit migrate

# Open the database studio
db_studio:
  bun run --bun drizzle-kit studio

# Push schema changes directly to the database (prototyping)
db_push:
  bun run --bun drizzle-kit push

# Introspect the database to generate schema files
db_pull:
  bun run --bun drizzle-kit pull

mockups:
  @echo "\n\nServing mockups at http://localhost:5555\n\n"
  cd mockups && python3 -m http.server 5555
