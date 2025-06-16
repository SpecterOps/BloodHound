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
  
-- Add analysis_enabled flag to asset_group_tags
ALTER TABLE asset_group_tags ADD COLUMN IF NOT EXISTS analysis_enabled BOOL NOT NULL DEFAULT false;

-- Set analysis_enabled for tier zero
UPDATE asset_group_tags SET analysis_enabled = true WHERE type = 1 AND position = 1;
