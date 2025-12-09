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

-- OpenGraph schema_node_kinds -  stores node kinds for open graph extensions
CREATE TABLE IF NOT EXISTS schema_node_kinds (
    id SERIAL PRIMARY KEY ,
    schema_extension_id INT NOT NULL REFERENCES schema_extensions (id) ON DELETE CASCADE, -- indicates which extension this node kind belongs to
    name TEXT UNIQUE NOT NULL, -- unique is required by the DAWGS kind table
    display_name TEXT NOT NULL, -- can be different from name but usually isn't other than Base/Entity
    description TEXT NOT NULL, -- human-readable description of the kind
    is_display_kind BOOL NOT NULL DEFAULT FALSE,
    icon TEXT NOT NULL, -- font-awesome icon
    icon_color TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX idx_graph_schema_node_kinds_extensions_id ON schema_node_kinds (schema_extension_id);

-- OpenGraph schema properties
CREATE TABLE IF NOT EXISTS schema_properties (
    id SERIAL NOT NULL,
    schema_extension_id INT NOT NULL,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    data_type TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    CONSTRAINT fk_schema_extensions_schema_properties FOREIGN KEY (schema_extension_id) REFERENCES schema_extensions(id) ON DELETE CASCADE,
    UNIQUE (schema_extension_id, name)
);

CREATE INDEX idx_schema_properties_schema_extensions_id on schema_properties (schema_extension_id);
-- OpenGraph schema_edge_kinds - store edge kinds for open graph extensions
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

-- OpenGraph schema_environments - stores environment mappings.
CREATE TABLE IF NOT EXISTS schema_environments (
    id SERIAL,
    schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id),
    environment_kind_id INTEGER NOT NULL REFERENCES kind(id),
    source_kind_id INTEGER NOT NULL REFERENCES kind(id),
    PRIMARY KEY (id),
    UNIQUE(environment_kind_id,source_kind_id)
);

-- OpenGraph schema_relationship_findings - Individual findings. ie T0WriteOwner, T0ADCSESC1, T0DCSync
CREATE TABLE IF NOT EXISTS schema_relationship_findings (
    id SERIAL,
    schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id) ON DELETE CASCADE,
    relationship_kind_id INTEGER NOT NULL REFERENCES kind(id),
    environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    PRIMARY KEY(id),
    UNIQUE(name)
);

-- OpenGraph schema_remediations - Remediation content table with FK to findings
CREATE TABLE IF NOT EXISTS schema_remediations (
    id SERIAL,
    finding_id INTEGER NOT NULL REFERENCES schema_relationship_findings(id) ON DELETE CASCADE,
    short_description TEXT,
    long_description TEXT,
    short_remediation TEXT,
    long_remediation TEXT,
    PRIMARY KEY(id),
    UNIQUE(finding_id)
);

-- OpenGraph schema_environments_principal_kinds - Environment to principal mappings
CREATE TABLE IF NOT EXISTS schema_environments_principal_kinds (
    id SERIAL,
    environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
    principal_kind INTEGER NOT NULL REFERENCES kind(id),
    PRIMARY KEY(id),
    UNIQUE(principal_kind)
);
