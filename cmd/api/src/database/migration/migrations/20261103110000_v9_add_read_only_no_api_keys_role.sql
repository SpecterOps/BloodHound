-- Copyright 2026 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
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
INSERT INTO roles (name, description, created_at, updated_at)
VALUES ('Read-Only (No API Keys)', 'Can read data without API key management', current_timestamp, current_timestamp)
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON (p.authority, p.name) IN (
    ('app', 'ReadAppConfig'),
    ('risks', 'GenerateReport'),
    ('auth', 'ManageSelf'),
    ('auth', 'ReadUsersMinimal'),
    ('graphdb', 'Read'),
    ('saved_queries', 'Read'),
    ('opengraph', 'Read')
)
WHERE r.name = 'Read-Only (No API Keys)'
ON CONFLICT DO NOTHING;

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM users_roles
        JOIN roles ON roles.id = users_roles.role_id
        WHERE roles.name = 'Read-Only (No API Keys)'
    ) THEN
        RAISE EXCEPTION 'cannot remove Read-Only (No API Keys) while it is assigned to users';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM sso_providers
        JOIN roles ON roles.id::text = sso_providers.config->'auto_provision'->>'default_role_id'
        WHERE roles.name = 'Read-Only (No API Keys)'
    ) THEN
        RAISE EXCEPTION 'cannot remove Read-Only (No API Keys) while it is an SSO default role';
    END IF;
END $$;
-- +goose StatementEnd

DELETE FROM roles_permissions
WHERE role_id = (SELECT id FROM roles WHERE name = 'Read-Only (No API Keys)');

DELETE FROM roles WHERE name = 'Read-Only (No API Keys)';
