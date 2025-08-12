
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

-- Change current Upload-Only role name to Client-Tasking
UPDATE roles
SET "name"='Client-Tasking'
WHERE roles.name = 'Upload-Only';

-- Create Upload-Only role with GraphDBIngest permission
INSERT INTO roles (name, description, created_at, updated_at) VALUES 
('Upload-Only', 'Used for users to ingest files manually', current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON ((r.name = 'Upload-Only' AND (p.authority, p.name) IN (
      ('graphdb', 'Ingest')
    )))
ON CONFLICT DO NOTHING;

-- Migrate users with Client-Tasking role to Upload-Only role
UPDATE users_roles
SET role_id=(select id from roles r where r."name" = 'Upload-Only')
WHERE role_id=(select id from roles r where r."name" ='Client-Tasking');