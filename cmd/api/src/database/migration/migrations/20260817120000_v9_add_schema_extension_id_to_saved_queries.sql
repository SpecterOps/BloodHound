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
