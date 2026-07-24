-- +goose Up
CREATE TABLE users (
  id                  INTEGER PRIMARY KEY,
  workos_id           TEXT NOT NULL UNIQUE,
  email               TEXT NOT NULL,
  username            TEXT NOT NULL,
  timezone            TEXT NOT NULL DEFAULT '',
  avatar_url          TEXT NOT NULL DEFAULT '',
  created_at          TEXT NOT NULL,
  updated_at          TEXT NOT NULL,
  username_changed_at TEXT
);

-- Case-insensitive uniqueness: "JaneDoe" and "janedoe" are the same handle, but
-- the casing the user typed is preserved for display. NOCASE folds ASCII only,
-- which is exactly the allowed username charset. Queries that look a username up
-- (rather than relying on the insert/update conflict) MUST write
-- `WHERE username = ? COLLATE NOCASE`, or SQLite won't use this index.
CREATE UNIQUE INDEX users_username_unique ON users(username COLLATE NOCASE);
CREATE INDEX users_email ON users(email);

CREATE TABLE account_deletion_requests (
  id           INTEGER PRIMARY KEY,
  user_id      INTEGER NOT NULL REFERENCES users(id),
  requested_at TEXT NOT NULL,
  reason       TEXT NOT NULL DEFAULT '',
  status       TEXT NOT NULL DEFAULT 'pending',
  handled_at   TEXT
);

-- One open request per user, enforced in the schema so CreateDeletionRequest is
-- idempotent even under a double-click race (the store's check-then-insert is not
-- atomic). Also serves the admin's "who wants out" query — SQLite can scan this
-- partial index for `WHERE status = 'pending'` — and stays tiny no matter how many
-- historical requests accumulate.
CREATE UNIQUE INDEX deletion_requests_pending ON account_deletion_requests(user_id) WHERE status = 'pending';

-- +goose Down
DROP INDEX deletion_requests_pending;
DROP TABLE account_deletion_requests;
DROP INDEX users_email;
DROP INDEX users_username_unique;
DROP TABLE users;
