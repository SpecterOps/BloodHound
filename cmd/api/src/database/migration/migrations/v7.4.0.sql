-- Add custom_node_kinds table
CREATE TABLE IF NOT EXISTS custom_node_kinds (
    id SERIAL PRIMARY KEY,
    kind_name varchar(256) UNIQUE NOT NULL,
    config JSONB NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
