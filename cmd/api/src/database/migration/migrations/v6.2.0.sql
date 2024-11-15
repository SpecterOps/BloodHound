-- Copyright 2024 Specter Ops, Inc.
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

-- Add updated posture page feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'updated_posture_page',
        'Updated Posture Page',
        'Enables the updated version of the posture page in the UI application',
        false,
        false)
ON CONFLICT DO NOTHING;

INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('graphdb', 'Ingest', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Grant the Upload-Only user GraphDBIngest permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Upload-Only'),
        (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' and permissions.name = 'Ingest'))
ON CONFLICT DO NOTHING;

-- Grant the Power User user GraphDBIngest permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Power User'),
        (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' and permissions.name = 'Ingest'))
ON CONFLICT DO NOTHING;

-- Grant the Admininstrator user GraphDBIngest permissions
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ((SELECT id FROM roles WHERE roles.name = 'Administrator'),
        (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' and permissions.name = 'Ingest'))
ON CONFLICT DO NOTHING;

-- Remove the GraphDBWrite permission from the Upload-Only role
DELETE FROM roles_permissions 
WHERE role_id = (SELECT id FROM roles WHERE roles.name = 'Upload-Only')
AND permission_id = (SELECT id FROM permissions WHERE permissions.authority = 'graphdb' AND permissions.name = 'Write');

-- Set the user's saml_provider_id to null when an sso_provider or saml_provider is deleted
ALTER TABLE ONLY users
    DROP CONSTRAINT IF EXISTS fk_users_saml_provider;
ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_saml_provider FOREIGN KEY (saml_provider_id) REFERENCES saml_providers (id) ON DELETE SET NULL;

-- Backfill users with their proper sso_provider when they have a saml_provider_id
UPDATE users u
SET sso_provider_id = (SELECT sso.id
                       FROM saml_providers saml
                                JOIN sso_providers sso ON sso.id = saml.sso_provider_id
                       WHERE u.saml_provider_id = saml.id)
WHERE sso_provider_id IS NULL
  AND saml_provider_id IS NOT NULL;
