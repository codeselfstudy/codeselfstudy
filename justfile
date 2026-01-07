# Find all the license files
find_licenses:
	find . -name "LICENSE.md" -not -path "*/node_modules/*"
