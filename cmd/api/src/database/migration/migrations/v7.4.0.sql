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

ALTER TABLE asset_group_history
	ADD COLUMN IF NOT EXISTS email VARCHAR(330) DEFAULT NULL;

-- Populate email for existing records by looking up the email address from the users table
UPDATE asset_group_history
	SET email = (SELECT email_address FROM users WHERE asset_group_history.actor = users.id)
	WHERE email IS NULL AND actor != 'SYSTEM';

-- Add asset_group_tag_selector_nodes table
CREATE TABLE IF NOT EXISTS asset_group_tag_selector_nodes
(
	selector_id int NOT NULL,
	node_id bigint NOT NULL,
	certified int NOT NULL DEFAULT 0,
	certified_by text,
	source int,
	created_at timestamp with time zone,
	updated_at timestamp with time zone,
	CONSTRAINT fk_asset_group_tag_selectors_asset_group_tag_selector_nodes FOREIGN KEY (selector_id) REFERENCES asset_group_tag_selectors(id) ON DELETE CASCADE,
	PRIMARY KEY (selector_id, node_id)
);

-- Migrate existing Tier Zero selectors
WITH inserted_selector AS (
  INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
  SELECT (SELECT id FROM asset_group_tags WHERE name = 'Tier Zero'), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, false
  FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
  WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

-- Migrate existing Owned selectors
WITH inserted_kind AS (
  INSERT INTO kind (name)
  SELECT 'Tag_' || replace(name, ' ', '_') as name
  FROM asset_groups
  WHERE tag = 'owned'
  ON CONFLICT DO NOTHING
  RETURNING id, name
),
inserted_tag AS (
  INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by)
  SELECT ik.id, 3, ag.name, ag.name, current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM'
  FROM inserted_kind ik JOIN asset_groups ag ON ik.name = 'Tag_' || replace(ag.name, ' ', '_')
  RETURNING id, name
),
inserted_selector AS (
  INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
  SELECT (SELECT id from inserted_tag), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, false
  FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
  WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;
