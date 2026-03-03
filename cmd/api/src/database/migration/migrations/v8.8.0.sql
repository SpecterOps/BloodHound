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
