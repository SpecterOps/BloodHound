-- +goose Up
CREATE TABLE IF NOT EXISTS data_quality_aggregations (
    id SERIAL PRIMARY KEY,
    run_id TEXT NOT NULL,
    schema_extension_id INTEGER NOT NULL,
    schema_environment_kind_id INTEGER NOT NULL REFERENCES kind(id),
    metric_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL DEFAULT 0,
    kind_id INTEGER DEFAULT NULL REFERENCES kind(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
    );

CREATE INDEX IF NOT EXISTS idx_data_quality_aggregations_run_id ON data_quality_aggregations USING btree (run_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_aggregations_created_at ON data_quality_aggregations USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_data_quality_aggregations_updated_at ON data_quality_aggregations USING btree (updated_at);

-- +goose Down
DROP TABLE IF EXISTS data_quality_aggregations;
