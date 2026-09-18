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
create table if not exists source_kinds (
  id smallserial,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);
INSERT INTO source_kinds (name)
VALUES ('Base'),
  ('AZBase') ON CONFLICT (name) DO NOTHING;
ALTER TABLE analysis_request_switch
ADD COLUMN IF NOT EXISTS delete_all_graph boolean DEFAULT false,
  ADD COLUMN IF NOT EXISTS delete_sourceless_graph boolean DEFAULT false,
  ADD COLUMN IF NOT EXISTS delete_source_kinds text [] DEFAULT ARRAY []::text [];
-- Remove the ReadAppConfig / WriteAppConfig from power users role
DELETE FROM roles_permissions
WHERE role_id = (
    SELECT id
    FROM roles
    WHERE roles.name = 'Power User'
  )
  AND permission_id IN (
    SELECT id
    FROM permissions
    WHERE permissions.authority = 'app'
      AND permissions.name IN ('WriteAppConfig')
  );
-- Add name index to asset_group_tag_selectors table for search
CREATE INDEX IF NOT EXISTS idx_asset_group_tag_selectors_name ON asset_group_tag_selectors USING btree (name);
-- if the explore_table_view flag doesnt exist, create explore_table_view feature flag, disable it, and make it non user-updatable.
-- this same query exists in 7.5.0, however it was merged after the release was cut so any tenants created before 7.5.0 will be missing this flag.
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'explore_table_view',
        'Explore Table View',
        'Adds a layout option to the Explore page that will display all nodes in a table view. It also will automatically display the table when a cypher query returned only nodes.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- enable explore_table_view feature flag
UPDATE feature_flags
SET enabled = true
WHERE key = 'explore_table_view';


-- Add Incoming Forest Trust Builders selector
WITH s AS (
  INSERT INTO asset_group_tag_selectors (
      asset_group_tag_id,
      created_at,
      created_by,
      updated_at,
      updated_by,
      disabled_at,
      disabled_by,
      name,
      description,
      is_default,
      allow_disable,
      auto_certify
    )
  SELECT
      (
        SELECT id
        FROM asset_group_tags
        WHERE name = 'Tier Zero'
      ),
      current_timestamp,
      'SYSTEM',
      current_timestamp,
      'SYSTEM',
      current_timestamp,
      'SYSTEM',
      'Incoming Forest Trust Builders',
      E'Members of this group can create incoming trusts that allow TGT delegation which can lead to compromise of the forest.',
      true,
      true,
      false
  WHERE NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = 'Incoming Forest Trust Builders')
  RETURNING id
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
	s.id,
	2,
	E'MATCH (n:Group) \nWHERE n.objectid ENDS WITH ''-557''\nRETURN n;'
FROM s;