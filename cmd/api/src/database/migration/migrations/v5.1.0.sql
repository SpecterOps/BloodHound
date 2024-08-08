-- Copyright 2023 Specter Ops, Inc.
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

-- Add saved queries permissions

-- Create new permissions saved query write and read
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('saved_queries', 'Read', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;
INSERT INTO permissions (authority, name, created_at, updated_at) VALUES ('saved_queries', 'Write', current_timestamp, current_timestamp) ON CONFLICT DO NOTHING;

-- Add saved query permissions to administrator and user roles
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'Administrator'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Write')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Read')) ON CONFLICT DO NOTHING;
INSERT INTO roles_permissions (role_id, permission_id) VALUES ((SELECT id FROM roles WHERE roles.name  = 'User'), (SELECT id FROM permissions WHERE permissions.authority  = 'saved_queries'  and permissions.name = 'Write')) ON CONFLICT DO NOTHING;
