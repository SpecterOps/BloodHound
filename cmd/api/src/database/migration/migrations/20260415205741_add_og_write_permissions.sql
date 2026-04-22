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
-- Add OpenGraph Write permissions to Admin role
-- +goose Up
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
ON (p.authority, p.name) = ('opengraph', 'Write')
WHERE r.name IN ('Administrator')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM roles_permissions
WHERE role_id IN (
    SELECT id FROM roles 
    WHERE name IN ('Administrator')
)
AND permission_id = (
    SELECT id FROM permissions 
    WHERE authority = 'opengraph' AND name = 'Write'
);
