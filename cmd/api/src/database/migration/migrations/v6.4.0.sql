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

-- Delete the `updated_posture_page` feature flag
DELETE FROM feature_flags WHERE key = 'updated_posture_page';

-- Add new config column in sso_providers table
ALTER TABLE IF EXISTS sso_providers ADD COLUMN IF NOT EXISTS config jsonb;

-- Update sso_providers table by backfilling existing sso providers' new config column with default values
UPDATE sso_providers set config = '{"auto_provision": {"enabled": false, "default_role_id": 0, "role_provision": false}}';
