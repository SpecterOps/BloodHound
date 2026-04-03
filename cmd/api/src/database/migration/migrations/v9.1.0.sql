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

-- Create anonymize translation table for storing original-to-anonymized name mappings
CREATE TABLE IF NOT EXISTS anonymize_translation_entries (
    id              BIGSERIAL PRIMARY KEY,
    node_graph_id   BIGINT NOT NULL,
    property_key    TEXT NOT NULL,
    original_value  TEXT NOT NULL,
    anonymized_value TEXT NOT NULL,
    object_type     TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_anonymize_translation_node_prop ON anonymize_translation_entries (node_graph_id, property_key);
CREATE INDEX IF NOT EXISTS idx_anonymize_translation_original ON anonymize_translation_entries (original_value);
CREATE INDEX IF NOT EXISTS idx_anonymize_translation_anonymized ON anonymize_translation_entries (anonymized_value);
