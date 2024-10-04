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

-- SSO Provider
CREATE TABLE IF NOT EXISTS sso_providers
(
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    type       int  NOT NULL,

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

    UNIQUE (name),
    UNIQUE (slug)
);

-- OIDC Provider
CREATE TABLE IF NOT EXISTS oidc_providers
(
    id              SERIAL PRIMARY KEY,
    client_id       TEXT NOT NULL,
    issuer          TEXT NOT NULL,
    sso_provider_id INTEGER REFERENCES sso_providers (id) ON DELETE CASCADE NULL,

    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Create the reference from saml_providers to sso_providers
ALTER TABLE ONLY saml_providers
    ADD COLUMN IF NOT EXISTS sso_provider_id INTEGER NULL;
ALTER TABLE ONLY saml_providers
    ADD CONSTRAINT fk_saml_provider_sso_provider FOREIGN KEY (sso_provider_id) REFERENCES sso_providers (id) ON DELETE CASCADE;

