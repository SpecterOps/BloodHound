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

-- Add the zone variant of the display name to schema_findings.
-- For built-in findings this is populated from tx-title.md by schemagen.
-- display_name is already the canonical UI title: schemagen now writes the full title
-- from title.md into that column instead of the short CUE token.
ALTER TABLE schema_findings ADD COLUMN IF NOT EXISTS zone_display_name TEXT;

-- +goose Down

ALTER TABLE schema_findings DROP COLUMN IF EXISTS zone_display_name;
