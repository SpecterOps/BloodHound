-- Copyright 2025 Specter Ops, Inc.
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

-- OpenGraph Search feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp, 
    'opengraph_search',
    'OpenGraph Search',
    'Enable OpenGraph Search', 
    false,
    false)
ON CONFLICT DO NOTHING;


-- OpenGraph graph schema - extensions (collectors)
CREATE TABLE IF NOT EXISTS schema_extensions (
    id SERIAL NOT NULL,
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    version TEXT NOT NULL,
    is_builtin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS schema_edge_kinds (
    id SERIAL NOT NULL,
    schema_extension_id INT NOT NULL REFERENCES schema_extensions (id) ON DELETE CASCADE, -- indicates which extension this edge kind belongs to
    name TEXT UNIQUE NOT NULL, -- unique is required by the DAWGS kind table, cypher only allows alphanumeric characters and underscores
    description TEXT NOT NULL, -- human-readable description of the edge-kind
    is_traversable BOOL NOT NULL DEFAULT FALSE, -- indicates whether the given edge-kind is traversable
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_schema_edge_kinds_extensions_id ON schema_edge_kinds (schema_extension_id);