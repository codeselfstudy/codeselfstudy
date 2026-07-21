-- +goose Up
CREATE TABLE emails (
  id            INTEGER PRIMARY KEY,
  message_id    TEXT NOT NULL UNIQUE,
  from_addr     TEXT NOT NULL DEFAULT '',
  to_addr       TEXT NOT NULL DEFAULT '',
  subject       TEXT NOT NULL DEFAULT '',
  sent_at       TEXT,
  received_at   TEXT NOT NULL,
  body_text     TEXT NOT NULL,
  status        TEXT NOT NULL DEFAULT 'received',
  extract_error TEXT
);

CREATE TABLE digests (
  id         INTEGER PRIMARY KEY,
  claimed_at TEXT NOT NULL,
  posted_at  TEXT,
  status     TEXT NOT NULL DEFAULT 'claimed',
  deal_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE deals (
  id                  INTEGER PRIMARY KEY,
  email_id            INTEGER NOT NULL REFERENCES emails(id),
  dedupe_key          TEXT NOT NULL UNIQUE,
  source              TEXT NOT NULL,
  title               TEXT NOT NULL,
  url                 TEXT,
  price               TEXT,
  ends_at             TEXT,
  description         TEXT,
  first_seen_at       TEXT NOT NULL,
  last_seen_at        TEXT NOT NULL,
  seen_count          INTEGER NOT NULL DEFAULT 1,
  posted_in_digest_id INTEGER REFERENCES digests(id)
);

CREATE INDEX deals_unposted ON deals(posted_in_digest_id) WHERE posted_in_digest_id IS NULL;

-- +goose Down
DROP INDEX deals_unposted;
DROP TABLE deals;
DROP TABLE digests;
DROP TABLE emails;
