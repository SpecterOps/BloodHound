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

-- Prepend all found duplicate emails on user table with user id in preparation for unique constraint
UPDATE users SET email_address = id || '-' || lower(email_address) where lower(email_address) in (SELECT distinct(lower(email_address)) FROM users GROUP BY lower(email_address) HAVING count(lower(email_address)) > 1);

-- Add unique constraint on user emails
ALTER TABLE IF EXISTS users
  DROP CONSTRAINT IF EXISTS users_email_address_key;
ALTER TABLE IF EXISTS users
  ADD CONSTRAINT users_email_address_key UNIQUE (email_address);

-- Add `back_button_support` feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'back_button_support',
        'Back Button Support',
        'Enable users to quickly navigate between views in a wider range of scenarios by utilizing the browser navigation buttons.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Add `tier_management_engine` feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'tier_management_engine',
        'Tier Management Engine',
        'Updates the managed assets selector engine and the asset management page.',
        false,
        false)
ON CONFLICT DO NOTHING;

-- Add `NTLM Post Processing` feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
        current_timestamp,
        'ntlm_post_processing',
        'NTLM Post Processing Support',
        'Enable the post processing of NTLM relay attack paths, this will enable the creation of CoerceAndRelayNTLMTo[LDAP, LDAPS, ADCS, SMB] edges.',
        false,
        true)
ON CONFLICT DO NOTHING;
