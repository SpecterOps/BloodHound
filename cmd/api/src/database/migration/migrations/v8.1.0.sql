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

-- Environment Targeted Access Control Feature Flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'targeted_access_control',
        'Targeted Access Control',
        'Enable power users and admins to set targeted access controls on users',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Environment Targeted Access Control
CREATE TABLE IF NOT EXISTS environment_access_control (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    environment TEXT NOT NULL,
    created_at timestamp with time zone DEFAULT current_timestamp,
    updated_at timestamp with time zone,
    CONSTRAINT environment_access_control_user_env_key UNIQUE (user_id, environment),
    CONSTRAINT environment_not_blank CHECK (btrim(environment) <> '')
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS all_environments BOOL DEFAULT TRUE;

-- File Ingest Details
ALTER TABLE ingest_jobs ADD COLUMN IF NOT EXISTS task_info jsonb NOT NULL DEFAULT '{"completed_tasks": []}'::jsonb;
ALTER TABLE ingest_tasks ADD COLUMN IF NOT EXISTS provided_file_name text NOT NULL DEFAULT '';
