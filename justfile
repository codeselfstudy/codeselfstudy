# Find all the license files
find_licenses:
	find . -name "LICENSE.md" -not -path "*/node_modules/*"

# Deploy to Fly.io using the local .bun-version
deploy:
	fly deploy --build-arg BUN_VERSION=$(cat .bun-version)
