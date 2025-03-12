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

CREATE TABLE IF NOT EXISTS kind
(
  id   SMALLSERIAL,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

-- Add asset_group_tiers table
CREATE TABLE IF NOT EXISTS asset_group_tiers
(
    id SERIAL NOT NULL,
    position integer NOT NULL,
    allow_certify boolean,
    PRIMARY KEY (id)
);

-- Add asset_group_labels table
CREATE TABLE IF NOT EXISTS asset_group_labels
(
    id SERIAL NOT NULL,
    asset_group_tier_id int,
    kind_id smallint,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone,
    created_by text,
    updated_at timestamp with time zone,
    updated_by text,
    deleted_at timestamp with time zone,
    deleted_by text,
    PRIMARY KEY (id),
    CONSTRAINT fk_asset_group_tiers_asset_group_labels FOREIGN KEY (asset_group_tier_id) REFERENCES asset_group_tiers(id),
    CONSTRAINT fk_kind_asset_group_labels FOREIGN KEY (kind_id) REFERENCES kind(id)
);

-- Add partial unique index for name for asset_group_labels
CREATE UNIQUE INDEX IF NOT EXISTS agl_name_unique_index ON asset_group_labels (name)
    WHERE deleted_at IS NULL;

-- Create tier xero record
INSERT INTO kind (name) VALUES ('TierZero') ON CONFLICT DO NOTHING;
INSERT INTO asset_group_tiers (id, position, allow_certify) VALUES (1, 0, false) ON CONFLICT DO NOTHING;
INSERT INTO asset_group_labels (name, asset_group_tier_id, kind_id, description, created_by, created_at, updated_by, updated_at)
    VALUES ('Tier Zero', 1, (SELECT id FROM kind WHERE name = 'TierZero'), 'Tier Zero', 'SYSTEM', current_timestamp, 'SYSTEM', current_timestamp)
    ON CONFLICT DO NOTHING;

-- Add asset_group_history tables
CREATE TABLE IF NOT EXISTS asset_group_history
(
    id BIGSERIAL NOT NULL,
    actor text NOT NULL,
    action text NOT NULL,
    target text,
    asset_group_label_id int NOT NULL,
    environment_id text,
    note text,
    created_at timestamp with time zone,
    PRIMARY KEY (id),
    CONSTRAINT fk_asset_group_history_asset_group_labels FOREIGN KEY (asset_group_label_id) REFERENCES asset_group_labels(id)
);


-- Add asset_group_label_selectors table
CREATE TABLE IF NOT EXISTS asset_group_label_selectors
(
    id SERIAL NOT NULL,
    asset_group_label_id int,
    created_at timestamp with time zone,
    created_by text,
    updated_at timestamp with time zone,
    updated_by text,
    disabled_at timestamp with time zone,
    disabled_by text,
    name text NOT NULL,
    description text,
    is_default boolean,
    allow_disable boolean,
    auto_certify boolean,
    PRIMARY KEY (id),
    CONSTRAINT fk_asset_group_labels_asset_group_selectors FOREIGN KEY (asset_group_label_id) REFERENCES asset_group_labels(id) ON DELETE CASCADE
);

-- Add asset_group_label_selector_seeds table
CREATE TABLE IF NOT EXISTS asset_group_label_selector_seeds
(
    selector_id int,
    type int,
    value text,
    CONSTRAINT fk_asset_group_label_selectors_asset_group_label_selector_seeds FOREIGN KEY (selector_id) REFERENCES asset_group_label_selectors(id) ON DELETE CASCADE
);
