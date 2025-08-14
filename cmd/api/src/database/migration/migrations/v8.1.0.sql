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

-- Add denormalized property columns to asset_group_tag_selector_nodes table
ALTER TABLE asset_group_tag_selector_nodes
  ADD COLUMN IF NOT EXISTS node_primary_kind TEXT,
  ADD COLUMN IF NOT EXISTS node_environment_id TEXT,
  ADD COLUMN IF NOT EXISTS node_object_id TEXT,
  ADD COLUMN IF NOT EXISTS node_name TEXT;

-- Add indexes for the above new columns added to asset_group_tag_selector_nodes table
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_primary_kind ON asset_group_tag_selector_nodes USING btree (node_primary_kind);
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_environment_id ON asset_group_tag_selector_nodes USING btree (node_environment_id);
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_object_id ON asset_group_tag_selector_nodes USING btree (node_object_id);
CREATE INDEX IF NOT EXISTS idx_agt_selector_nodes_name ON asset_group_tag_selector_nodes USING btree (node_name);

-- File Ingest Details
ALTER TABLE ingest_tasks ADD COLUMN IF NOT EXISTS provided_file_name text NOT NULL DEFAULT '';
CREATE TABLE IF NOT EXISTS completed_tasks (
    id BIGSERIAL PRIMARY KEY,
    ingest_job_id BIGINT NOT NULL REFERENCES ingest_jobs(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    parent_file_name TEXT NOT NULL,
    errors TEXT[] NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE INDEX IF NOT EXISTS idx_completed_tasks_ingest_job_id ON completed_tasks USING btree (ingest_job_id);
