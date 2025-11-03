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

-- Add Audit Log permission and Auditor role 
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('audit_log', 'Read', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

INSERT INTO roles (name, description, created_at, updated_at) VALUES 
 ('Auditor', 'Can read data and audit logs', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON (
    (r.name = 'Auditor' AND (p.authority, p.name) IN (
        ('app', 'ReadAppConfig'),
        ('risks', 'GenerateReport'),
        ('audit_log', 'Read'),
        ('auth', 'CreateToken'),
        ('auth', 'ManageSelf'),
        ('auth', 'ReadUsers'),
        ('graphdb', 'Read'),
        ('saved_queries', 'Read'),
        ('clients', 'Read')
    ))
    OR 
    (r.name = 'Administrator' AND (p.authority, p.name) IN (
               ('audit_log', 'Read')
    ))    
) 
ON CONFLICT DO NOTHING;

-- Adding site specific columns to ad_data_quality_aggregations and ad_data_quality_stats tables
ALTER TABLE "ad_data_quality_aggregations" ADD COLUMN IF NOT EXISTS sites BIGINT DEFAULT 0;
ALTER TABLE "ad_data_quality_aggregations" ADD COLUMN IF NOT EXISTS siteservers BIGINT DEFAULT 0;
ALTER TABLE "ad_data_quality_aggregations" ADD COLUMN IF NOT EXISTS sitesubnets BIGINT DEFAULT 0;

ALTER TABLE "ad_data_quality_stats" ADD COLUMN IF NOT EXISTS sites BIGINT DEFAULT 0;
ALTER TABLE "ad_data_quality_stats" ADD COLUMN IF NOT EXISTS siteservers BIGINT DEFAULT 0;
ALTER TABLE "ad_data_quality_stats" ADD COLUMN IF NOT EXISTS sitesubnets BIGINT DEFAULT 0;


-- Add Sites default selector to Tier Zero
WITH src_data AS (
  SELECT * FROM (VALUES
-- START
('Sites', true, true, E'MATCH (n:Site) \nRETURN n;', E'Control over an Active Directory site may allow users to compromise all assets associated with the site through the application of Group Policy Objects. Since AD Sites contain at least a Domain Controller as a Site Server, this results in the potential compromise of at least one domain in the forest. Therefore, Active Directory Site objects are Tier Zero.')
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
  2
FROM src_data d WHERE NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = d.name)
  RETURNING id, name
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
  s.id,
  2,
  d.cypher
FROM inserted_selectors s JOIN src_data d ON d.name = s.name;
