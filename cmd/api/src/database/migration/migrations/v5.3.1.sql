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

CREATE INDEX IF NOT EXISTS idx_asset_group_collections_asset_group_id ON asset_group_collections USING btree (asset_group_id);
CREATE INDEX IF NOT EXISTS idx_asset_group_collections_created_at ON asset_group_collections USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_asset_group_collections_updated_at ON asset_group_collections USING btree (updated_at);

CREATE INDEX IF NOT EXISTS idx_asset_group_collection_entries_asset_group_collection_id ON asset_group_collection_entries USING btree (asset_group_collection_id);
CREATE INDEX IF NOT EXISTS idx_asset_group_collection_entries_created_at ON asset_group_collection_entries USING btree (created_at);
CREATE INDEX IF NOT EXISTS idx_asset_group_collection_entries_updated_at ON asset_group_collection_entries USING btree (updated_at);

TRUNCATE asset_group_collection_entries;
