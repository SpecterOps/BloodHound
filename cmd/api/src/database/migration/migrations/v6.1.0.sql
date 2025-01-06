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

-- Add Scheduled Analysis Configs
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES ('analysis.scheduled',
        'Scheduled Analysis',
        'This configuration parameter allows setting a schedule for analysis. When enabled, analysis will only run when the scheduled time arrives',
        '{
          "enabled": false,
          "rrule": ""
        }',
        current_timestamp, current_timestamp)
ON CONFLICT DO NOTHING;

-- Add last analysis time to datapipe status so we can track scheduled analysis time properly
ALTER TABLE datapipe_status
    ADD COLUMN IF NOT EXISTS "last_analysis_run_at" TIMESTAMP with time zone;

-- SSO Provider
CREATE TABLE IF NOT EXISTS sso_providers
(
    id         SERIAL PRIMARY KEY,
    name       TEXT    NOT NULL,
    slug       TEXT    NOT NULL,
    type       INTEGER NOT NULL,

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

    UNIQUE (name),
    UNIQUE (slug)
);

-- OIDC Provider
CREATE TABLE IF NOT EXISTS oidc_providers
(
    id              SERIAL PRIMARY KEY,
    client_id       TEXT                                                    NOT NULL,
    issuer          TEXT                                                    NOT NULL,
    sso_provider_id INTEGER REFERENCES sso_providers (id) ON DELETE CASCADE NULL,

    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Create the reference from saml_providers to sso_providers
ALTER TABLE ONLY saml_providers
    ADD COLUMN IF NOT EXISTS sso_provider_id INTEGER NULL;
ALTER TABLE ONLY saml_providers
    DROP CONSTRAINT IF EXISTS fk_saml_provider_sso_provider;
ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT fk_saml_provider_sso_provider FOREIGN KEY (sso_provider_id) REFERENCES sso_providers (id) ON DELETE CASCADE;

-- Backfill our sso_providers table with the existing data from saml_providers
-- The hardcoded type is determined by the AuthProvider
-- See:
-- https://github.com/SpecterOps/BloodHound/blob/main/cmd/api/src/model/auth.go#L565
INSERT INTO sso_providers(name, slug, type) (SELECT name, lower(replace(name, ' ', '-')), 1
                                             FROM saml_providers
                                             WHERE sso_provider_id IS NULL)
ON CONFLICT DO NOTHING;

-- Backfill the references from the newly created sso_provider entries
UPDATE saml_providers
SET sso_provider_id = (SELECT id FROM sso_providers WHERE name = saml_providers.name)
WHERE saml_providers.sso_provider_id IS NULL;

-- Add the sso_provider to the users table
ALTER TABLE ONLY users
    ADD COLUMN IF NOT EXISTS sso_provider_id INTEGER NULL;
ALTER TABLE ONLY users
    DROP CONSTRAINT IF EXISTS fk_users_sso_provider;
ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_sso_provider FOREIGN KEY (sso_provider_id) REFERENCES sso_providers (id) ON DELETE SET NULL;

-- Backfill users with their proper sso_provider when they have a saml_provider_id
UPDATE users u
SET sso_provider_id = (SELECT sso.id
                       FROM saml_providers saml
                                JOIN sso_providers sso ON sso.id = saml.sso_provider_id
                       WHERE u.saml_provider_id = saml.id)
WHERE sso_provider_id IS NULL
  AND saml_provider_id IS NOT NULL;
