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
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES (
    'analysis.graph_storage_optimization',
    'When to run graph storage optimization',
    'This configuration parameter controls which pipeline stages trigger graph storage optimization (vacuum/analyze) on the graph database. Each stage can be independently enabled or disabled.',
    '{"after_boot": false, "after_analysis": false, "min_interval_seconds": 86400}',
    current_timestamp,
    current_timestamp
)
ON CONFLICT DO NOTHING;

ALTER TABLE datapipe_status
    ADD COLUMN IF NOT EXISTS last_complete_optimize_at timestamp with time zone;

-- +goose Down
ALTER TABLE datapipe_status
    DROP COLUMN IF EXISTS last_complete_optimize_at;

DELETE FROM parameters WHERE key = 'analysis.graph_storage_optimization';
