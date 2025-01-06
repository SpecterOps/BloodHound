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

-- Add unique constraint to roles table
ALTER TABLE IF EXISTS roles
  DROP CONSTRAINT IF EXISTS roles_name_key;
ALTER TABLE IF EXISTS roles
  ADD CONSTRAINT roles_name_key UNIQUE (name);

-- Add unique constraint to permissions table
ALTER TABLE IF EXISTS permissions
  DROP CONSTRAINT IF EXISTS permissions_authority_name_key;
ALTER TABLE IF EXISTS permissions
  ADD CONSTRAINT permissions_authority_name_key UNIQUE (authority, name);

-- Feature Flags
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'clear_graph_data', 'Clear Graph Data', 'Enables the ability to delete all nodes and edges from the graph database.', true, false) ON CONFLICT DO NOTHING;
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'risk_exposure_new_calculation', 'Use new tier zero risk exposure calculation', 'Enables the use of new tier zero risk exposure metatree metrics.', false, false) ON CONFLICT DO NOTHING;
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'fedramp_eula', 'FedRAMP EULA', 'Enables showing the FedRAMP EULA on every login. (Enterprise only)', false, false) ON CONFLICT DO NOTHING;
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'auto_tag_t0_parent_objects', 'Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero', 'Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.', true, true) ON CONFLICT DO NOTHING;

-- Note - order matters permissions and roles ops must come before roles permissions ops
-- Permissions
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('saved_queries', 'Read', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('saved_queries', 'Write', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('clients', 'Read', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('db', 'Wipe', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('graphdb', 'Mutate', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Roles
INSERT INTO roles (name, description,  created_at, updated_at) VALUES ('Power User', 'Can upload data, manage clients, and perform any action a User can', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Roles Permissions
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Write')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'db'  and permissions.name = 'Wipe')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  AND permissions.name = 'Mutate')) ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Write')) ON CONFLICT DO NOTHING;
-- Swap user clients manage for clients read permission
DELETE FROM roles_permissions WHERE role_id = (SELECT id FROM roles WHERE roles.name  = 'User') AND permission_id = (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Manage');
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'clients'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Read-Only'), (SELECT id FROM permissions WHERE permissions.authority  = 'auth'  and permissions.name = 'CreateToken')) ON CONFLICT DO NOTHING;

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
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Power User'), (SELECT id FROM permissions WHERE permissions.authority  = 'graphdb'  AND permissions.name = 'Mutate')) ON CONFLICT DO NOTHING;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_saved_queries_description ON saved_queries using gin(description gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_saved_queries_name ON saved_queries USING gin(name gin_trgm_ops);
