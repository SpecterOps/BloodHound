-- Add custom_node_kinds table
CREATE TABLE IF NOT EXISTS custom_node_kinds (
    id SERIAL PRIMARY KEY,
    kind_id SMALLINT UNIQUE NOT NULL,
    config JSONB NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT fk_kind_id FOREIGN KEY (kind_id) REFERENCES kind(id) ON DELETE CASCADE
);
