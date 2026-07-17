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
-- Use Raw Object ID feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (current_timestamp,
    current_timestamp,
    'use_raw_object_id',
    'Use Raw Object ID',
    'When enabled, ingestion preserves the original case of object IDs instead of uppercasing them, fixing case-collision issues for sources with case-sensitive identifiers. (BED-8954)',
    false,
    false)
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM feature_flags WHERE key = 'use_raw_object_id';
