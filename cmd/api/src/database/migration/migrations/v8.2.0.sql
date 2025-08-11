
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

-- Grant Upload-Only user GraphDBWrite permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Upload-Only'),
        (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' and permissions.name = 'Write'))
ON CONFLICT DO NOTHING;

-- Create new User-Upload-Only with GraphDBWrite
INSERT INTO roles (name, description, created_at, updated_at) VALUES 
('User-Upload-Only', 'Used to ingest files manually', current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON ((r.name = 'User-Upload-Only' AND (p.authority, p.name) IN (
      ('graphdb', 'Write')
    )))
ON CONFLICT DO NOTHING;
