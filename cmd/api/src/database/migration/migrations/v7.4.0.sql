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

ALTER TABLE asset_group_history
	ADD COLUMN IF NOT EXISTS email VARCHAR(330) DEFAULT NULL;

-- Populate email for existing records by looking up the email address from the users table
UPDATE asset_group_history
	SET email = (SELECT email_address FROM users WHERE asset_group_history.actor = users.id)
	WHERE email IS NULL AND actor != 'SYSTEM';

-- Add asset_group_tag_selector_nodes table
CREATE TABLE IF NOT EXISTS asset_group_tag_selector_nodes
(
	selector_id int NOT NULL,
	node_id bigint NOT NULL,
	certified int NOT NULL DEFAULT 0,
	certified_by text,
	source int,
	created_at timestamp with time zone,
	updated_at timestamp with time zone,
	CONSTRAINT fk_asset_group_tag_selectors_asset_group_tag_selector_nodes FOREIGN KEY (selector_id) REFERENCES asset_group_tag_selectors(id) ON DELETE CASCADE,
	PRIMARY KEY (selector_id, node_id)
	);

-- Add custom_node_kinds table
CREATE TABLE IF NOT EXISTS custom_node_kinds (
  id            SERIAL        PRIMARY KEY,
  kind_name     VARCHAR(256)  NOT NULL,
  config        JSONB         NOT NULL,

  created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

  unique(kind_name)
);
