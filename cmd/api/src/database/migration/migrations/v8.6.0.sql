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
-- Drop the column with the incorrect foreign key
ALTER TABLE IF EXISTS schema_environments
    DROP COLUMN IF EXISTS source_kind_id;

-- Add the column back with the correct foreign key reference
ALTER TABLE IF EXISTS schema_environments
    ADD COLUMN source_kind_id INTEGER NOT NULL REFERENCES source_kinds(id);
