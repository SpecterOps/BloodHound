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

CREATE TABLE IF NOT EXISTS schema_kind_info (
    id SERIAL PRIMARY KEY,
    kind_id INT NOT NULL REFERENCES kind (id),
    node_kind_id INT REFERENCES schema_node_kinds (id) ON DELETE CASCADE,
    relationship_kind_id INT REFERENCES schema_relationship_kinds (id) ON DELETE CASCADE,
    info_key TEXT NOT NULL,
    title TEXT NOT NULL,
    position INT NOT NULL,
    content JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    CONSTRAINT schema_kind_info_kind_origin
        CHECK (num_nonnulls(node_kind_id, relationship_kind_id) = 1),
    CONSTRAINT schema_kind_info_info_key_not_empty
        CHECK (btrim(info_key) <> ''),
    CONSTRAINT schema_kind_info_title_not_empty
        CHECK (btrim(title) <> ''),
    CONSTRAINT schema_kind_info_position_nonnegative_nonzero
        CHECK (position >= 0),
    CONSTRAINT schema_kind_info_content_is_object
        CHECK (jsonb_typeof(content) = 'object'),
    CONSTRAINT schema_kind_info_unique_kind_position -- ensure position is unique per kind
        UNIQUE (kind_id, position), 
    CONSTRAINT schema_kind_info_unique_kind_info_key -- ensure info_key is unique per kind
        UNIQUE (kind_id, info_key)
);

CREATE INDEX IF NOT EXISTS idx_schema_kind_info_kind_id
    ON schema_kind_info (kind_id);

-- +goose Down

DROP TABLE IF EXISTS schema_kind_info;
