-- +goose Up
ALTER TABLE saved_queries
    ADD COLUMN IF NOT EXISTS schema_extension_id INTEGER REFERENCES schema_extensions (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_saved_queries_schema_extension_id ON saved_queries (schema_extension_id);

DROP INDEX IF EXISTS idx_saved_queries_composite_index;

CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_composite_index
    ON saved_queries USING btree (user_id, name, schema_extension_id) NULLS NOT DISTINCT;

-- +goose Down
DROP INDEX IF EXISTS idx_saved_queries_composite_index;

CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_composite_index
    ON saved_queries USING btree (user_id, name);

DROP INDEX IF EXISTS idx_saved_queries_schema_extension_id;

ALTER TABLE saved_queries
    DROP COLUMN IF EXISTS schema_extension_id;
