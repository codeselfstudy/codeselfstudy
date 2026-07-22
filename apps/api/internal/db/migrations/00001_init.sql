-- +goose Up
-- Placeholder so the embedded migrations/ directory has at least one file
-- (//go:embed migrations/*.sql in migrate.go fails the build otherwise).
-- Remove or rename this when the first real migration lands.
SELECT 1;

-- +goose Down
SELECT 1;
