-- Copyright 2024 Specter Ops, Inc.
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

-- Drop column saml_provider_id from users table
ALTER TABLE ONLY users
DROP CONSTRAINT IF EXISTS fk_users_saml_provider;
ALTER TABLE ONLY users
DROP COLUMN IF EXISTS saml_provider_id;

-- Add root_uri_version and backfill existing saml providers to 1 or "/v1/login/saml/"
ALTER TABLE ONLY saml_providers
  ADD COLUMN IF NOT EXISTS root_uri_version INTEGER NOT NULL DEFAULT 1;

-- Update root_uri_version to default to 2 or "/v2/sso/" for newly created saml providers
ALTER TABLE ONLY saml_providers
  ALTER COLUMN root_uri_version SET DEFAULT 2;

-- Set the `updated_posture_page` feature flag to true
UPDATE feature_flags SET enabled = true WHERE key = 'updated_posture_page';
