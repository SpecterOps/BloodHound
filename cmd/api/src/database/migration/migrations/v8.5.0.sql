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

-- upsert_schema_edge_kind - atomically upserts an edge kind into both the DAWGS kind and schema_edge_kinds tables.
-- This function addresses the edge case where both a schema_edge_kind and schema_node_kind can point to same DAWGS kind.
CREATE OR REPLACE FUNCTION upsert_schema_edge_kind(edge_kind_name TEXT, edge_kind_schema_extension_id INT,
                                                   edge_kind_description TEXT, edge_kind_is_traversable BOOLEAN)
    RETURNS TABLE (
        schema_edge_kind_id INT,
        return_schema_extension_id INT,
        return_name TEXT,
        return_description TEXT,
        return_is_traversable BOOL,
        return_created_at TIMESTAMP WITH TIME ZONE,
        return_updated_at TIMESTAMP WITH TIME ZONE,
        return_deleted_at TIMESTAMP WITH TIME ZONE
    )
as $$
BEGIN

    LOCK TABLE schema_node_kinds, schema_edge_kinds IN EXCLUSIVE MODE; -- DAWGS Kind table is append only so no need to lock

    IF (
       SELECT EXISTS (
                     SELECT 1
                     FROM schema_node_kinds snk
                     JOIN kind k ON snk.kind_id = k.id
                     WHERE k.name = edge_kind_name)) THEN
        RAISE EXCEPTION 'duplicate key value violates unique constraint "%", kind already declared in the schema_node_kinds table', edge_kind_name;
    END IF;

    RETURN QUERY
        WITH dawgs_kinds AS
            ( INSERT INTO kind (name) VALUES (edge_kind_name) ON CONFLICT (name) DO UPDATE SET name = edge_kind_name RETURNING id, name)
        INSERT INTO schema_edge_kinds (kind_id, schema_extension_id, description, is_traversable)
    SELECT id,
           edge_kind_schema_extension_id,
           edge_kind_description,
           edge_kind_is_traversable
    FROM dawgs_kinds
    RETURNING id, schema_extension_id, edge_kind_name, description, is_traversable, created_at, updated_at, deleted_at;
END
$$ LANGUAGE plpgsql;

-- upsert_schema_node_kind - atomically upserts a node kind into both the DAWGS kind and schema_node_kind tables.
-- This function addresses the edge case where both a schema_edge_kind and schema_node_kind can point to same DAWGS kind.
CREATE OR REPLACE FUNCTION upsert_schema_node_kind(node_kind_name TEXT, node_kind_schema_extension_id INT,
                                                   node_kind_display_name TEXT, node_kind_description TEXT,
                                                   node_kind_is_display_kind BOOLEAN, node_kind_icon TEXT,
                                                   node_kind_icon_color TEXT)
    RETURNS TABLE (
        schema_node_kind_id INT,
        return_schema_extension_id INT,
        return_name TEXT,
        return_display_name TEXT,
        return_description TEXT,
        return_is_display_kind bool,
        return_icon TEXT,
        return_icon_color TEXT,
        return_created_at TIMESTAMP WITH TIME ZONE,
        return_updated_at TIMESTAMP WITH TIME ZONE,
        return_deleted_at TIMESTAMP WITH TIME ZONE
    )
as $$
BEGIN

    LOCK TABLE schema_node_kinds, schema_edge_kinds IN EXCLUSIVE MODE; -- DAWGS Kind table is append only so no need to lock

    IF (
       SELECT EXISTS (
                     SELECT 1
                     FROM schema_edge_kinds sek
                     JOIN kind k ON sek.kind_id = k.id
                     WHERE k.name = node_kind_name)) THEN
        RAISE EXCEPTION 'duplicate key value violates unique constraint "%", kind already declared in the schema_edge_kinds table', node_kind_name;
    END IF;

    RETURN QUERY
    WITH dawgs_kinds
             AS ( INSERT INTO kind (name) VALUES (node_kind_name) ON CONFLICT (name) DO UPDATE SET name = node_kind_name RETURNING id, name)
    INSERT
    INTO schema_node_kinds (kind_id, schema_extension_id, display_name, description, is_display_kind, icon,
                            icon_color)
    SELECT dk.id,
           node_kind_schema_extension_id,
           node_kind_display_name,
           node_kind_description,
           node_kind_is_display_kind,
           node_kind_icon,
           node_kind_icon_color
    FROM dawgs_kinds dk
    RETURNING id, schema_extension_id, node_kind_name, display_name, description, is_display_kind,
        icon, icon_color, created_at, updated_at, deleted_at;
END
$$ LANGUAGE plpgsql;