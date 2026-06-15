-- Copyright 2026 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0

-- +goose Up
-- Generic metric tables support node and relationship counts today and other OG DQ metrics later.
CREATE TABLE IF NOT EXISTS data_quality_stats (
    id SERIAL PRIMARY KEY,
    run_id TEXT NOT NULL,
    schema_extension_id INTEGER NOT NULL,
    schema_environment_id INTEGER NOT NULL,
    environment_id TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL DEFAULT 0,
    schema_node_kind_id INTEGER DEFAULT NULL,
    schema_relationship_kind_id INTEGER DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_data_quality_stats_run_id ON data_quality_stats (run_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_schema_extension_id ON data_quality_stats (schema_extension_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_schema_environment_id ON data_quality_stats (schema_environment_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_environment_id ON data_quality_stats (environment_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_metric_type ON data_quality_stats (metric_type);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_metric_name ON data_quality_stats (metric_name);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_schema_node_kind_id ON data_quality_stats (schema_node_kind_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_schema_relationship_kind_id ON data_quality_stats (schema_relationship_kind_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_created_at ON data_quality_stats (created_at);

CREATE TABLE IF NOT EXISTS data_quality_stats_aggregation (
    id SERIAL PRIMARY KEY,
    run_id TEXT NOT NULL,
    schema_extension_id INTEGER NOT NULL,
    schema_environment_id INTEGER NOT NULL,
    metric_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL DEFAULT 0,
    schema_node_kind_id INTEGER DEFAULT NULL,
    schema_relationship_kind_id INTEGER DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_run_id ON data_quality_stats_aggregation (run_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_schema_extension_id ON data_quality_stats_aggregation (schema_extension_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_schema_environment_id ON data_quality_stats_aggregation (schema_environment_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_metric_type ON data_quality_stats_aggregation (metric_type);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_metric_name ON data_quality_stats_aggregation (metric_name);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_schema_node_kind_id ON data_quality_stats_aggregation (schema_node_kind_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_schema_relationship_kind_id ON data_quality_stats_aggregation (schema_relationship_kind_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_stats_aggregation_created_at ON data_quality_stats_aggregation (created_at);

-- +goose Down
DROP TABLE IF EXISTS data_quality_stats_aggregation;
DROP TABLE IF EXISTS data_quality_stats;
