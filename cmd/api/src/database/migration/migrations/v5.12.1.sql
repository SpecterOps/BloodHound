ALTER TABLE IF EXISTS saved_queries
  ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';

CREATE TABLE IF NOT EXISTS saved_queries_permissions
(
  id             BIGINT PRIMARY KEY,
  shared_to_user_id        TEXT REFERENCES users (id),
  query_id BIGSERIAL REFERENCES saved_queries (id),
  global         BOOL DEFAULT FALSE
);
