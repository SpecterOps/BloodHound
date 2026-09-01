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
    ADD COLUMN IF NOT EXISTS schema_extension_id INTEGER REFERENCES schema_extensions (id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS query_key TEXT
        CONSTRAINT chk_saved_queries_extension_shape
        CHECK (
            (schema_extension_id IS NULL AND query_key IS NULL AND user_id <> '00000000-0000-0000-0000-000000000000')
            OR
            (schema_extension_id IS NOT NULL AND query_key IS NOT NULL AND user_id = '00000000-0000-0000-0000-000000000000')
        ),
    ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_saved_queries_schema_extension_id
    ON saved_queries (schema_extension_id) WHERE schema_extension_id IS NOT NULL;

DROP INDEX IF EXISTS idx_saved_queries_composite_index;

CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_user_id_name
    ON saved_queries (user_id, name) WHERE schema_extension_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_schema_extension_id_name
    ON saved_queries (schema_extension_id, name) WHERE schema_extension_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_extension_query_key
    ON saved_queries (schema_extension_id, query_key) WHERE schema_extension_id IS NOT NULL;


-- +goose Down

DELETE FROM saved_queries WHERE schema_extension_id IS NOT NULL;

DROP INDEX IF EXISTS idx_saved_queries_user_id_name;
DROP INDEX IF EXISTS idx_saved_queries_schema_extension_id_name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_queries_composite_index
    ON saved_queries USING btree (user_id, name);

DROP INDEX IF EXISTS idx_saved_queries_schema_extension_id;

DROP INDEX IF EXISTS idx_saved_queries_extension_query_key;

ALTER TABLE saved_queries
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS query_key,
    DROP COLUMN IF EXISTS schema_extension_id;
