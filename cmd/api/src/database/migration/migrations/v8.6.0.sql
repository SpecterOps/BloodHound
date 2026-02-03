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

ALTER TABLE IF EXISTS schema_environments
    DROP CONSTRAINT IF EXISTS schema_environments_source_kind_id_fkey;

ALTER TABLE IF EXISTS schema_environments
    ADD CONSTRAINT schema_environments_source_kind_id_fkey
    FOREIGN KEY (source_kind_id) REFERENCES source_kinds(id);


-- OpenGraph Findings feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'opengraph_findings',
    'OpenGraph Findings',
    'Enable OpenGraph Findings',
    false,
    false)
ON CONFLICT DO NOTHING;

-- Add API Tokens parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('auth.api_tokens',
        'API Tokens',
        'This configuration parameter enables/disables authorization through API Tokens',
        '{"enabled":true}',
        current_timestamp,
        current_timestamp)
  ON CONFLICT DO NOTHING;

-- Add Timeouts parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('api.timeout_limit',
        'Query Timeout Limit',
        'This configuration parameter enables/disables a timeout limit for API Requests',
        '{"enabled":true}',
        current_timestamp,
        current_timestamp)
  ON CONFLICT DO NOTHING;

-- Update Scheduled Analysis description 
UPDATE parameters SET description = 'This configuration parameter allows setting a schedule for analysis. When enabled, analysis will run when the scheduled time arrives or when manually requested' WHERE key = 'analysis.scheduled';

-- Add Namespace column to schema_extensions
ALTER TABLE schema_extensions
    ADD COLUMN IF NOT EXISTS namespace TEXT;

UPDATE schema_extensions SET namespace = LEFT(name, 3)
WHERE namespace IS NULL OR namespace = '';

ALTER TABLE schema_extensions
    ALTER COLUMN namespace SET NOT NULL;

DO $$
    BEGIN
        IF NOT EXISTS (
                      SELECT 1
                      FROM pg_constraint
                      WHERE conname = 'schema_extensions_namespace_unique'
        ) THEN
            ALTER TABLE schema_extensions
                ADD CONSTRAINT schema_extensions_namespace_unique UNIQUE (namespace);
        END IF;
    END$$;

ALTER TABLE IF EXISTS schema_edge_kinds RENAME TO schema_relationship_kinds;

-- Remove ETAC from feature flags since it has moved to DogTags
DELETE FROM feature_flags WHERE key = 'environment_targeted_access_control';