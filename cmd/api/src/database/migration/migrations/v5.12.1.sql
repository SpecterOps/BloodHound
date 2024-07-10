ALTER TABLE IF EXISTS saved_queries
  ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';

CREATE TABLE IF NOT EXISTS shared_saved_queries
(
  id             BIGINT PRIMARY KEY,
  owner_id       TEXT REFERENCES users (id),
  user_id        TEXT REFERENCES users (id),
  saved_query_id TEXT REFERENCES saved_queries (id),
  global         BOOL DEFAULT FALSE
);
