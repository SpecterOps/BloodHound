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
-- Add Alerts permissions to permissions table
INSERT INTO permissions(created_at, updated_at, authority, name)
VALUES (
           current_timestamp,
           current_timestamp,
           'alerts',
           'Read'
       ),
       (
           current_timestamp,
           current_timestamp,
           'alerts',
           'Manage'
       )
ON CONFLICT DO NOTHING;

-- Add Alerts Read permission to Administrator and Auditor roles
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
       JOIN permissions p
            ON (p.authority, p.name) = ('alerts', 'Read')
WHERE r.name IN ('Administrator', 'Auditor')
ON CONFLICT DO NOTHING;

-- Add Alerts Manage permission to Administrator role
INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
       JOIN permissions p
            ON (p.authority, p.name) = ('alerts', 'Manage')
WHERE r.name = 'Administrator'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM roles_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE (authority, name) IN (('alerts', 'Read'), ('alerts', 'Manage'))
);

DELETE FROM permissions
WHERE (authority, name) IN (('alerts', 'Read'), ('alerts', 'Manage'));
