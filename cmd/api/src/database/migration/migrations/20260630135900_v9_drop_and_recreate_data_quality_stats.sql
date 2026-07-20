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

-- Rename the old data_quality_stats table that is not currently used anywhere but still exists in prod
ALTER TABLE IF EXISTS data_quality_stats RENAME TO data_quality_stats_old;
ALTER SEQUENCE IF EXISTS data_quality_stats_id_seq RENAME TO data_quality_stats_old_id_seq;
ALTER INDEX IF EXISTS data_quality_stats_pkey RENAME TO data_quality_stats_old_pkey;
ALTER INDEX IF EXISTS idx_data_quality_stats_deleted_at RENAME TO idx_data_quality_stats_old_deleted_at;
ALTER INDEX IF EXISTS idx_data_quality_stats_created_at RENAME TO idx_data_quality_stats_old_created_at;


CREATE TABLE IF NOT EXISTS data_quality_stats (
    id SERIAL PRIMARY KEY,
    run_id TEXT NOT NULL,
    schema_extension_id INTEGER NOT NULL,
    schema_environment_kind_id INTEGER NOT NULL REFERENCES kind(id),
    environment_id TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL DEFAULT 0,
    kind_id INTEGER DEFAULT NULL REFERENCES kind(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
); 

CREATE INDEX IF NOT EXISTS idx_data_quality_stats_created_at ON data_quality_stats (created_at);

-- +goose Down
DROP TABLE IF EXISTS data_quality_stats;
ALTER TABLE IF EXISTS data_quality_stats_old RENAME TO data_quality_stats;
ALTER SEQUENCE IF EXISTS data_quality_stats_old_id_seq RENAME TO data_quality_stats_id_seq;
ALTER INDEX IF EXISTS data_quality_stats_old_pkey RENAME TO data_quality_stats_pkey;
ALTER INDEX IF EXISTS idx_data_quality_stats_old_deleted_at RENAME TO idx_data_quality_stats_deleted_at;
ALTER INDEX IF EXISTS idx_data_quality_stats_old_created_at RENAME TO idx_data_quality_stats_created_at;
