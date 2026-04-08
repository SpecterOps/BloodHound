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


-- Add support_account flag to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS support_account BOOL DEFAULT FALSE;

-- Rename opengraph_collector_platform_support feature flag to openhound_support
UPDATE feature_flags
SET key         = 'openhound_support',
    name        = 'OpenHound Support',
    description = 'Enable creation and communication with OpenHound platform'
WHERE key = 'opengraph_collector_platform_support';

-- Create Read Jobs permissions
INSERT INTO permissions (authority, name, created_at, updated_at)
VALUES ('collection', 'ReadJobs', current_timestamp, current_timestamp)
  ON CONFLICT DO NOTHING;

-- Add CollectionReadJobs permission to Administrator and Auditor
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON (p.authority, p.name) = ('collection', 'ReadJobs')
WHERE r.name IN ('Auditor', 'Administrator')
  ON CONFLICT DO NOTHING;

-- Remove unused permission
DELETE FROM roles_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE authority='auth' and name = 'ManageAppConfig');

DELETE FROM permissions
WHERE authority='auth' and name = 'ManageAppConfig';


DO $$
  BEGIN
    IF EXISTS (
      SELECT 1
      FROM information_schema.table_constraints tc
          JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
      WHERE tc.table_name = 'schema_environments'
        AND tc.constraint_name = 'schema_environments_source_kind_id_fkey'
        AND ccu.table_name = 'source_kinds'
    )
   THEN
      ALTER TABLE schema_environments DROP CONSTRAINT schema_environments_source_kind_id_fkey;

      UPDATE schema_environments SET source_kind_id = (
        SELECT sk.kind_id FROM source_kinds sk WHERE sk.id = schema_environments.source_kind_id
      );

      ALTER TABLE schema_environments ADD CONSTRAINT schema_environments_source_kind_id_fkey FOREIGN KEY (source_kind_id) REFERENCES kind(id);

      DELETE FROM source_kinds where active = false;
      ALTER TABLE source_kinds DROP COLUMN IF EXISTS active;
  END IF;
END $$;

CREATE OR REPLACE FUNCTION upsert_source_kind(source_kind_name TEXT) RETURNS source_kinds AS $$
DECLARE
  kind_row kind%rowtype;
  source_kind_row source_kinds%rowtype;
BEGIN
    -- Use advisory lock to serialize calls with the same source kind name
    PERFORM pg_advisory_xact_lock(hashtext(source_kind_name));

    SELECT * INTO kind_row FROM upsert_kind(source_kind_name);

    -- Then, try to find existing source_kind by kind_id
    SELECT * INTO source_kind_row FROM source_kinds WHERE kind_id = kind_row.id;

    IF source_kind_row.id IS NULL THEN
            INSERT INTO source_kinds (kind_id) VALUES (kind_row.id) RETURNING * INTO source_kind_row;
    END IF;

RETURN source_kind_row;
END $$
LANGUAGE plpgsql;

-- Add API Key Expiration Support feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'api_key_expiration_support',
        'API Key Expiration Support',
        'Enables API Key Expiration configuration options',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Add API Tokens Expiration Parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('auth.api_token_expiration',
        'API Token Expiration',
        'This configuration parameter enables/disables created API tokens to expire after the set number of days.',
        '{"enabled":false, "expiration_period":90}',
        current_timestamp,
        current_timestamp)
ON CONFLICT DO NOTHING;

-- Add column to allow deletion of relationships by kind
ALTER TABLE analysis_request_switch ADD COLUMN IF NOT EXISTS delete_relationships text [] DEFAULT ARRAY []::text [];

-- Auditors to have all environments access
UPDATE users
SET all_environments = true
WHERE id IN (
  SELECT u.id
  FROM users u
  JOIN users_roles ur ON ur.user_id = u.id
  JOIN roles r ON ur.role_id = r.id
  WHERE r.name = 'Auditor'
);

-- Make opengraph_extension_management user updatable
UPDATE feature_flags
SET user_updatable = true
WHERE key = 'opengraph_extension_management';