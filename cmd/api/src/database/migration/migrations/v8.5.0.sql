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

-- OpenGraph schema_node_kinds -  stores node kinds for open graph extensions. This FK's to the DAWGS kind table directly.
CREATE TABLE IF NOT EXISTS schema_node_kinds (
    id SERIAL PRIMARY KEY,
    schema_extension_id INT NOT NULL REFERENCES schema_extensions (id) ON DELETE CASCADE, -- indicates which extension this node kind belongs to
    kind_id SMALLINT NOT NULL UNIQUE REFERENCES kind (id) ON DELETE CASCADE,
    display_name TEXT NOT NULL, -- can be different from name but usually isn't other than Base/Entity
    description TEXT NOT NULL, -- human-readable description of the kind
    is_display_kind BOOL NOT NULL DEFAULT FALSE,
    icon TEXT NOT NULL, -- font-awesome icon
    icon_color TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_graph_schema_node_kinds_extensions_id ON schema_node_kinds (schema_extension_id);

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

CREATE INDEX IF NOT EXISTS idx_schema_properties_schema_extensions_id on schema_properties (schema_extension_id);

-- OpenGraph schema_edge_kinds - store edge kinds for open graph extensions. This FK's to the DAWGS kind table directly.
-- Renamed to schema_relationship_kinds
CREATE TABLE IF NOT EXISTS schema_edge_kinds (
    id SERIAL PRIMARY KEY,
    schema_extension_id INT NOT NULL REFERENCES schema_extensions (id) ON DELETE CASCADE, -- indicates which extension this edge kind belongs to
    kind_id SMALLINT NOT NULL UNIQUE REFERENCES kind (id) ON DELETE CASCADE,
    description TEXT NOT NULL, -- human-readable description of the edge-kind
    is_traversable BOOL NOT NULL DEFAULT FALSE, -- indicates whether the given edge-kind is traversable
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_schema_edge_kinds_extensions_id ON schema_edge_kinds (schema_extension_id);

-- OpenGraph schema_environments - stores environment mappings.
CREATE TABLE IF NOT EXISTS schema_environments (
    id SERIAL,
    schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id) ON DELETE CASCADE,
    environment_kind_id INTEGER NOT NULL REFERENCES kind(id),
    source_kind_id INTEGER NOT NULL REFERENCES kind(id),
    PRIMARY KEY (id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    UNIQUE(environment_kind_id,source_kind_id)
);

CREATE INDEX IF NOT EXISTS idx_schema_environments_extension_id ON schema_environments (schema_extension_id);

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

CREATE INDEX IF NOT EXISTS idx_schema_relationship_findings_extension_id ON schema_relationship_findings (schema_extension_id);
CREATE INDEX IF NOT EXISTS idx_schema_relationship_findings_environment_id ON schema_relationship_findings(environment_id);

-- OpenGraph remediation_content_type - ENUM type for remediation content categories, IF NOT EXISTS is not natively supported for type creation
DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'remediation_content_type') THEN
            CREATE TYPE  remediation_content_type AS ENUM (
                'short_description',
                'long_description',
                'short_remediation',
                'long_remediation'
                );
        END IF;
    END
$$;

-- OpenGraph schema_remediations - Normalized remediation content table with FK to findings
CREATE TABLE IF NOT EXISTS schema_remediations (
    finding_id INTEGER NOT NULL REFERENCES schema_relationship_findings(id) ON DELETE CASCADE,
    content_type remediation_content_type NOT NULL,
    content TEXT STORAGE MAIN,
    PRIMARY KEY(finding_id, content_type)
);

-- Index for filtering by content_type (single content type queries)
CREATE INDEX IF NOT EXISTS idx_schema_remediations_content_type ON schema_remediations(content_type);

-- OpenGraph schema_environments_principal_kinds - Environment to principal mappings
CREATE TABLE IF NOT EXISTS schema_environments_principal_kinds (
    environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
    principal_kind INTEGER NOT NULL REFERENCES kind(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    PRIMARY KEY(environment_id, principal_kind)
);

CREATE INDEX IF NOT EXISTS idx_schema_environments_principal_kinds_principal_kind ON schema_environments_principal_kinds (principal_kind);

-- Added to report warnings for opengraph files that attempt to create invalid relationships.
ALTER TABLE ingest_jobs
    ADD COLUMN IF NOT EXISTS partial_failed_files integer DEFAULT 0;

ALTER TABLE completed_tasks
    ADD COLUMN IF NOT EXISTS warnings TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE source_kinds
    ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT true NOT NULL;

UPDATE source_kinds SET active = true WHERE active is NULL;

-- Enables Citrix RDP support by default
UPDATE parameters
SET
    value = '{ "enabled": true }'
WHERE key = 'analysis.citrix_rdp_support';

-- Drop old ETAC table if the old and new table exist due to a failed v8.3.0 migration
DO
$$
    BEGIN
        IF EXISTS (SELECT
                   FROM pg_tables
                   WHERE schemaname = 'public'
                     AND tablename = 'environment_targeted_access_control')
        THEN
            DROP TABLE IF EXISTS environment_access_control;
        END IF;
    END
$$;

-- Feature flag for Client Bearer Token Auth
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'client_bearer_auth',
        'Client Bearer Auth',
        'Enable clients to be authenticated using bearer tokens.',
        false,
        false)
  ON CONFLICT DO NOTHING;

 -- Add AGT tuning parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('analysis.tagging', 'Analysis Tagging Configuration', 'This configuration parameter determines the limits used during the asset group tagging phase of analysis', '{"dawgs_worker_limit": 2, "expansion_worker_limit": 3, "selector_worker_limit": 7}', current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

-- upsert_kind checks to see if a kind exists in the kind table and inserts it if not.
-- A SELECT is used instead of an insert CTE with ON CONDITION DO as the latter will increment the kind's SERIAL id even
-- if the kind already exists.
CREATE OR REPLACE FUNCTION upsert_kind(node_kind_name TEXT) RETURNS kind AS $$
DECLARE
    kind_row kind%rowtype;
BEGIN
    LOCK kind;

    SELECT * INTO kind_row FROM kind WHERE kind.name = node_kind_name;

    IF kind_row.id IS NULL THEN
        INSERT INTO kind (name) VALUES (node_kind_name) RETURNING * INTO kind_row;
    END IF;

    RETURN kind_row;
END $$ LANGUAGE plpgsql;
