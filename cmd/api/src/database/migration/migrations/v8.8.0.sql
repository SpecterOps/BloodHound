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
-- Drop the compound unique constraint on schema_environments (environment_kind_id, source_kind_id)
-- and add a unique constraint on just environment_kind_id
ALTER TABLE IF EXISTS schema_environments
    DROP CONSTRAINT IF EXISTS schema_environments_environment_kind_id_source_kind_id_key;

DO $$
    BEGIN
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'schema_environments_environment_kind_id_key'
        ) THEN
            ALTER TABLE schema_environments
                ADD CONSTRAINT schema_environments_environment_kind_id_key UNIQUE (environment_kind_id);
        END IF;
    END$$;

-- Drop list findings table, it was unused;
DROP TABLE IF EXISTS schema_list_findings;

-- Create schema_findings table
CREATE TABLE IF NOT EXISTS schema_findings (
  id SERIAL,
  type INTEGER NOT NULL,
  schema_extension_id INTEGER NOT NULL REFERENCES schema_extensions(id) ON DELETE CASCADE,
  environment_id INTEGER NOT NULL REFERENCES schema_environments(id) ON DELETE CASCADE,
  kind_id INTEGER NOT NULL REFERENCES kind(id),
  name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
  PRIMARY KEY(id),
  UNIQUE(name)
);
CREATE INDEX IF NOT EXISTS idx_schema_findings_extension_id ON schema_findings (schema_extension_id);
CREATE INDEX IF NOT EXISTS idx_schema_findings_environment_id ON schema_findings(environment_id);

-- Populate schema_findings from old schema_relationship_findings
DO $$
BEGIN
  IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_relationship_findings') THEN
    INSERT INTO schema_findings (id, type, schema_extension_id, kind_id, environment_id, name, display_name, created_at)
    SELECT id, 1, schema_extension_id, relationship_kind_id, environment_id, name, display_name, created_at
    FROM schema_relationship_findings ON CONFLICT (name) DO NOTHING;
  END IF;
END
$$;

-- Update schema_remediations to reference schema_findings instead of schema_relationship_findings
ALTER TABLE schema_remediations DROP CONSTRAINT IF EXISTS schema_remediations_finding_id_fkey;
ALTER TABLE schema_remediations ADD CONSTRAINT schema_remediations_finding_id_fkey FOREIGN KEY (finding_id) REFERENCES schema_findings(id) ON DELETE CASCADE;

-- Drop schema_relationship_findings
DROP TABLE IF EXISTS schema_relationship_findings;

-- Add subtypes table to schema_findings
CREATE TABLE IF NOT EXISTS schema_findings_subtypes (
  schema_finding_id INTEGER NOT NULL REFERENCES schema_findings(id) ON DELETE CASCADE,
  subtype TEXT NOT NULL,
  PRIMARY KEY(schema_finding_id, subtype)
);

-- Update the 'auth_tokens' table adding expiration column
ALTER TABLE auth_tokens
ADD COLUMN IF NOT EXISTS expires_at timestamp with time zone;

-- Add a column to `custom_node_kinds` to more easily correlate OpenGraph icons  
ALTER TABLE IF EXISTS custom_node_kinds 
    ADD COLUMN IF NOT EXISTS schema_node_kind_id INTEGER REFERENCES schema_node_kinds (id) ON DELETE SET NULL;

-- Add Posture PDF Export feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'posture_pdf_export',
        'Posture PDF Export',
        'Enables PDF export from Posture page.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Backfill the custom_node_kinds table with any missing icon definitions from the schema_node_kinds table
DO $$
DECLARE schema_node_kind_record RECORD;
BEGIN
  FOR schema_node_kind_record IN 
    SELECT 
      schema_node_kinds.id, 
      kind.name AS kind_name, 
      schema_node_kinds.icon, 
      schema_node_kinds.icon_color 
    FROM schema_node_kinds 
    JOIN kind ON schema_node_kinds.kind_id = kind.id 
    JOIN schema_extensions ON schema_node_kinds.schema_extension_id = schema_extensions.id 
    WHERE schema_node_kinds.icon IS NOT NULL 
      AND schema_node_kinds.icon != '' 
      AND schema_node_kinds.is_display_kind = true 
      AND schema_node_kinds.deleted_at IS NULL 
      AND schema_extensions.is_builtin = false
  LOOP
    IF NOT EXISTS (SELECT 1 
      FROM custom_node_kinds 
      WHERE schema_node_kind_id = schema_node_kind_record.id) THEN
        IF NOT EXISTS (SELECT 1 
          FROM custom_node_kinds 
          WHERE kind_name = schema_node_kind_record.kind_name) THEN
            INSERT INTO custom_node_kinds (kind_name, schema_node_kind_id, config) 
            VALUES (schema_node_kind_record.kind_name, schema_node_kind_record.id, jsonb_build_object('icon', jsonb_build_object('type', 'font-awesome', 'name', schema_node_kind_record.icon, 'color', schema_node_kind_record.icon_color)));
        ELSE
          UPDATE custom_node_kinds SET schema_node_kind_id = schema_node_kind_record.id, config = jsonb_build_object('icon', jsonb_build_object('type', 'font-awesome', 'name', schema_node_kind_record.icon, 'color', schema_node_kind_record.icon_color)), updated_at = NOW() WHERE kind_name = schema_node_kind_record.kind_name;
        END IF;
    END IF;
  END LOOP;
END
$$;
