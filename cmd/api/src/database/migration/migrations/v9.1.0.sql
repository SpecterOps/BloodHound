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
-- Add OpenGraph Phase 2 feature flag


-- Add OpenGraph permissions to permissions table
INSERT INTO permissions(created_at, updated_at, authority, name)
VALUES (
        current_timestamp,
        current_timestamp,
        'opengraph',
        'Read'
       ),
       (
        current_timestamp,
        current_timestamp,
        'opengraph',
        'Write'
       )
ON CONFLICT DO NOTHING;

-- Add OpenGraph Read permissions to specific roles

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
ON (p.authority, p.name) = ('opengraph', 'Read')
WHERE r.name IN ('Administrator', 'User', 'Read-Only', 'Power User', 'Auditor')
ON CONFLICT DO NOTHING;

-- Add OpenGraph Write permissions to Admin role
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
ON (p.authority, p.name) = ('opengraph', 'Write')
WHERE r.name IN ('Administrator')
ON CONFLICT DO NOTHING;

-- Update the 'auth_tokens' table adding created_by column
-- As of current, we don't have a way to backfill the data, so we are leaving this field optional for now
ALTER TABLE auth_tokens
  ADD COLUMN IF NOT EXISTS created_by text;

-- Create fk_auth_tokens_created_by to auth_tokens referencing users.id if it doesn't exist
DO $$
  BEGIN
    IF NOT EXISTS (
      SELECT 1
      FROM pg_constraint
      WHERE conname = 'fk_auth_tokens_created_by'
    ) THEN
      ALTER TABLE auth_tokens
        ADD CONSTRAINT fk_auth_tokens_created_by
          FOREIGN KEY (created_by)
            REFERENCES users(id);
    END IF;
  END $$;

-- Add Findings Table feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'findings_table',
        'Findings Table',
        'Enables a new table view for findings.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Add Events Table 
CREATE TABLE IF NOT EXISTS events (
  id UUID PRIMARY KEY NOT NULL,
  type TEXT NOT NULL,
  message TEXT NOT NULL,
  data JSONB NOT NULL,
  created_at timestamp with time zone DEFAULT current_timestamp,
  processed_at timestamp with time zone DEFAULT NULL
);
