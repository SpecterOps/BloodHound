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

-- Configuring all existing records of "SYSTEM" to "BloodHound" within the asset_group_tags, asset_group_tag_selectors, asset_group_tag_selector_nodes, and asset_group_history tables
UPDATE asset_group_tags
SET created_by = CASE WHEN created_by = 'SYSTEM' THEN 'BloodHound' ELSE created_by END,
    updated_by = CASE WHEN updated_by = 'SYSTEM' THEN 'BloodHound' ELSE updated_by END,
    deleted_by = CASE WHEN deleted_by = 'SYSTEM' THEN 'BloodHound' ELSE deleted_by END
WHERE created_by = 'SYSTEM' OR updated_by = 'SYSTEM' OR deleted_by = 'SYSTEM';

UPDATE asset_group_tag_selectors
SET created_by = CASE WHEN created_by = 'SYSTEM' THEN 'BloodHound' ELSE created_by END,
    updated_by = CASE WHEN updated_by = 'SYSTEM' THEN 'BloodHound' ELSE updated_by END,
    disabled_by = CASE WHEN disabled_by = 'SYSTEM' THEN 'BloodHound' ELSE disabled_by END
WHERE created_by = 'SYSTEM' OR updated_by = 'SYSTEM' OR disabled_by = 'SYSTEM';

UPDATE asset_group_tag_selector_nodes
SET certified_by = 'BloodHound'
WHERE certified_by = 'SYSTEM';

UPDATE asset_group_history
SET actor = 'BloodHound'
WHERE actor = 'SYSTEM';


-- Explicitly set glyph values for the default asset_group_tags
-- Find Tier Zero by position
UPDATE asset_group_tags SET glyph = 'gem' WHERE position = 1;
-- Find Owned by type
UPDATE asset_group_tags SET glyph = 'skull' WHERE type = 3;
