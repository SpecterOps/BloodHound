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

-- drop any keys known to exist
ALTER TABLE IF EXISTS ONLY asset_group_selectors
  DROP CONSTRAINT IF EXISTS asset_group_selectors_name_key;

ALTER TABLE IF EXISTS ONLY asset_group_selectors
  DROP CONSTRAINT IF EXISTS asset_group_selectors_unique_name;

ALTER TABLE IF EXISTS ONLY asset_group_selectors
  DROP CONSTRAINT IF EXISTS idx_asset_group_selectors_name;

ALTER TABLE IF EXISTS ONLY asset_group_selectors
  DROP CONSTRAINT IF EXISTS idx_asset_group_selectors_deleted_at;

ALTER TABLE IF EXISTS ONLY asset_group_selectors
  DROP CONSTRAINT IF EXISTS asset_group_selectors_name_unique;

ALTER TABLE IF EXISTS ONLY asset_group_selectors
  DROP CONSTRAINT IF EXISTS asset_group_selectors_name_assetgroupid_key;

-- create the key we care about
ALTER TABLE IF EXISTS ONLY asset_group_selectors
  ADD CONSTRAINT asset_group_selectors_name_assetgroupid_key UNIQUE (name, asset_group_id);
