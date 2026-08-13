-- Copyright 2026 Specter Ops, Inc.
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

-- +goose Up
CREATE TABLE IF NOT EXISTS saml_consumed_identifiers
(
  sso_provider_id INTEGER                  NOT NULL REFERENCES sso_providers (id) ON DELETE CASCADE,
  idp_issuer      VARCHAR(1024)            NOT NULL,
  identifier_type TEXT                     NOT NULL CHECK (identifier_type IN ('response', 'assertion')),
  identifier      VARCHAR(255)             NOT NULL,
  expires_at      TIMESTAMP WITH TIME ZONE NOT NULL,
  consumed_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

  PRIMARY KEY (sso_provider_id, idp_issuer, identifier_type, identifier)
);

CREATE INDEX IF NOT EXISTS idx_saml_consumed_identifiers_expires_at
  ON saml_consumed_identifiers (expires_at);

-- +goose Down
DROP TABLE IF EXISTS saml_consumed_identifiers;
