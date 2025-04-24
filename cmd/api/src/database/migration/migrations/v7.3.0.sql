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

-- This table is normally created by dawgs, as defined in schema_up.sql
-- We add it here to maintain a new FK to asset_group_tags below regardless
-- of graph driver selected. Any future changes to the schema should be reflected
-- in `schema_up.sql` as well
CREATE TABLE IF NOT EXISTS kind
(
  id   SMALLSERIAL,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

-- Add asset_group_tags table
CREATE TABLE IF NOT EXISTS asset_group_tags
(
    id SERIAL NOT NULL,
    type int NOT NULL,
    kind_id smallint,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamp with time zone,
    created_by text,
    updated_at timestamp with time zone,
    updated_by text,
    deleted_at timestamp with time zone,
    deleted_by text,
    position integer,
    require_certify boolean,
    PRIMARY KEY (id),
    CONSTRAINT fk_kind_asset_group_tags FOREIGN KEY (kind_id) REFERENCES kind(id)
);

-- Add partial unique index for name for asset_group_tags
CREATE UNIQUE INDEX IF NOT EXISTS agl_name_unique_index ON asset_group_tags (name)
    WHERE deleted_at IS NULL;

-- Create tier xero record
WITH inserted_kind AS (
INSERT INTO kind (name) VALUES ('Tag_Tier_Zero') ON CONFLICT DO NOTHING
  RETURNING id)
INSERT INTO asset_group_tags (name, type, kind_id, description, created_by, created_at, updated_by, updated_at, position, require_certify)
  VALUES ('Tier Zero', 1, (SELECT id FROM inserted_kind), 'Tier Zero', 'SYSTEM', current_timestamp, 'SYSTEM', current_timestamp, 1, FALSE)
  ON CONFLICT DO NOTHING;

-- Add asset_group_history tables
CREATE TABLE IF NOT EXISTS asset_group_history
(
    id BIGSERIAL NOT NULL,
    actor text NOT NULL,
    action text NOT NULL,
    target text,
    asset_group_tag_id int NOT NULL,
    environment_id text,
    note text,
    created_at timestamp with time zone,
    PRIMARY KEY (id),
    CONSTRAINT fk_asset_group_history_asset_group_tags FOREIGN KEY (asset_group_tag_id) REFERENCES asset_group_tags(id)
);


-- Add asset_group_tag_selectors table
CREATE TABLE IF NOT EXISTS asset_group_tag_selectors
(
    id SERIAL NOT NULL,
    asset_group_tag_id int,
    created_at timestamp with time zone,
    created_by text,
    updated_at timestamp with time zone,
    updated_by text,
    disabled_at timestamp with time zone,
    disabled_by text,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    is_default boolean NOT NULL DEFAULT FALSE,
    allow_disable boolean NOT NULL DEFAULT TRUE,
    auto_certify boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    CONSTRAINT fk_asset_group_tags_asset_group_selectors FOREIGN KEY (asset_group_tag_id) REFERENCES asset_group_tags(id) ON DELETE CASCADE
);

-- Add asset_group_tag_selector_seeds table
CREATE TABLE IF NOT EXISTS asset_group_tag_selector_seeds
(
    selector_id int NOT NULL,
    type int NOT NULL,
    value text NOT NULL,
    CONSTRAINT fk_asset_group_tag_selectors_asset_group_tag_selector_seeds FOREIGN KEY (selector_id) REFERENCES asset_group_tag_selectors(id) ON DELETE CASCADE
);

-- generic ingest
ALTER TABLE IF EXISTS file_upload_jobs RENAME TO ingest_jobs;
ALTER TABLE ingest_tasks ADD COLUMN IF NOT EXISTS is_generic BOOLEAN NOT NULL DEFAULT FALSE;

-- GA for ntlm post processing
UPDATE feature_flags SET user_updatable = false WHERE key = 'ntlm_post_processing';
UPDATE feature_flags SET enabled = true WHERE key = 'ntlm_post_processing';
