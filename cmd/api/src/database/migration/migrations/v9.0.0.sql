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
