-- +goose Up
CREATE TABLE IF NOT EXISTS graph_retention_policies (
    id BIGSERIAL PRIMARY KEY,
    ttl INTERVAL NOT NULL DEFAULT INTERVAL '7 days',
    enabled BOOLEAN NOT NULL DEFAULT true,
    schema_extension_id INTEGER REFERENCES schema_extensions (id) ON DELETE CASCADE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT schema_extension_id_idx UNIQUE NULLS NOT DISTINCT (schema_extension_id)
);

INSERT INTO graph_retention_policies (created_at, updated_at) 
    values (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE graph_retention_policies;
