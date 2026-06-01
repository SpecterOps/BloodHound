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
-- +goose Up
INSERT INTO permissions (authority, name, created_at, updated_at)
VALUES ('auth',
        'ReadUsersMinimal',
        current_timestamp,
        current_timestamp)
ON CONFLICT DO NOTHING;

DELETE FROM roles_permissions
WHERE permission_id = (SELECT id 
                        FROM permissions 
                        WHERE authority = 'auth' AND name = 'ReadUsers')
    AND role_id IN (SELECT id 
                    FROM roles 
                    WHERE name IN ('User', 'Read-Only', 'Power User'));

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
    ON (
      (r.name IN ('User', 'Read-Only', 'Power User') AND (p.authority, p.name) IN (
            ('auth', 'ReadUsersMinimal')
      ))
)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles_permissions r
JOIN permissions f
    ON f.id = r.permission_id
JOIN permissions p
    ON (p.authority, p.name) = ('auth', 'ReadUsersMinimal')
WHERE (f.authority, f.name) = ('auth', 'ReadUsers')
ON CONFLICT DO NOTHING;

-- +goose Down
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON (p.authority, p.name) = ('auth', 'ReadUsers')
WHERE r.name IN ('User', 'Read-Only', 'Power User')
ON CONFLICT DO NOTHING;

DELETE FROM roles_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE authority = 'auth' AND name = 'ReadUsersMinimal');

DELETE FROM permissions WHERE authority = 'auth' AND name = 'ReadUsersMinimal';
