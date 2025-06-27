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

-- Add analysis_enabled flag to asset_group_tags
ALTER TABLE asset_group_tags ADD COLUMN IF NOT EXISTS analysis_enabled BOOL;

-- Set analysis_enabled to true for tier zero and false for other tiers
UPDATE asset_group_tags SET analysis_enabled = position = 1 WHERE type = 1 AND analysis_enabled IS NULL;

-- Add EULA custom text
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('eula.custom_text', 'EULA Custom Text', 'This configuration parameter overrides the EULA agreement text with provided text.', '{"custom_text": ""}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Add Auth Session TTL Hours
INSERT INTO parameters (key, name, description, value, created_at, updated_at) VALUES ('auth.session_ttl_hours', 'Auth Session TTL Hours', 'This configuration parameter determines the length of time in hours a logged in session stays active before expiration.', '{"hours": 8}', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Retire the `auto_tag_t0_parent_objects` feature flag
UPDATE feature_flags SET enabled = true, user_updatable = false WHERE key = 'auto_tag_t0_parent_objects';

-- Add RO-DC default selector to Tier Zero
WITH src_data AS (
  SELECT * FROM (VALUES
-- START
('Read-Only DCs', false, true, E'MATCH (n:Computer)\nWHERE n.isReadOnlyDC = true\nRETURN n;', E'An attacker with control over a RODC computer object can compromise Tier Zero principals. The attacker can modify the msDS-RevealOnDemandGroup and msDS-NeverRevealGroup attributes of the RODC computer object such that the RODC can retrieve the credentials of a targeted Tier Zero principal. The attacker can obtain admin access to the OS of the RODC through the managedBy attribute, from where they can obtain the credentials of the RODC krbtgt account. With that, the attacker can create a RODC golden ticket for the target principal. This ticket can be converted to a real golden ticket as the target has been added to the msDS-RevealOnDemandGroup attribute and is not protected by the msDS-NeverRevealGroup attribute. Therefore, the RODC computer object is Tier Zero.')
-- END
  ) AS s (name, enabled, allow_disable, cypher, description)
), inserted_selectors AS (
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
  (SELECT id FROM asset_group_tags WHERE type = 1 and position = 1 LIMIT 1),
  current_timestamp,
  'SYSTEM',
  current_timestamp,
  'SYSTEM',
  CASE WHEN NOT d.enabled THEN current_timestamp ELSE NULL END,
  CASE WHEN NOT d.enabled THEN 'SYSTEM' ELSE NULL END,
  d.name,
  d.description,
  true,
  d.allow_disable,
  false
FROM src_data d WHERE NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = d.name)
  RETURNING id, name
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
  s.id,
  2,
  d.cypher
FROM inserted_selectors s JOIN src_data d ON d.name = s.name;
