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

ALTER TABLE IF EXISTS audit_logs
  RENAME COLUMN source TO source_ip_address;

ALTER TABLE IF EXISTS audit_logs
  DROP CONSTRAINT IF EXISTS audit_logs_status_check,
  ADD CONSTRAINT status_check
  CHECK (status IN ('intent', 'success', 'failure')),
  ALTER COLUMN status SET DEFAULT 'intent',
  ALTER COLUMN source_ip_address TYPE TEXT,
  ADD COLUMN IF NOT EXISTS commit_id TEXT;

-- Add indices for scalability
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_email ON audit_logs USING btree (actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_logs_source_ip_address ON audit_logs USING btree (source_ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs USING btree (status);
UPDATE feature_flags SET enabled = false, user_updatable = false WHERE key = 'adcs';

-- Add clients read permission
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('clients', 'Read', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Grant administrator client read
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;

-- Swap user clients manage for clients read permission
DELETE FROM roles_permissions WHERE role_id = (SELECT id FROM roles WHERE roles.name  = 'User') AND permission_id = (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Manage');
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;

-- Fix read-only missing create token
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Read-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'CreateToken')) ON CONFLICT DO NOTHING;

-- Add role Power User
INSERT INTO roles (name, description,  created_at, updated_at) VALUES ('Power User', 'Can upload data, manage clients, and perform any action a User can', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Populate power user permissions
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'app'  and permissions.name = 'ReadAppConfig')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'app'  and permissions.name = 'WriteAppConfig')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'risks'  and permissions.name = 'GenerateReport')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'risks'  and permissions.name = 'ManageRisks')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'CreateToken')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'ManageSelf')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Manage')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Tasking')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'collection'  and permissions.name = 'ManageJobs')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  and permissions.name = 'Write')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Write')) ON CONFLICT DO NOTHING;
