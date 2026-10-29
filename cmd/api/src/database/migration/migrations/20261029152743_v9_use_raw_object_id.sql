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
UPDATE feature_flags
SET enabled = TRUE
WHERE
  key = 'use_raw_object_id'
  -- this exists only on the new installation
  AND NOT EXISTS (SELECT 1 FROM installations);

-- +goose Down
-- Intentionaly preserve the flag's current state. Disabling it after
-- ingestion may cause object ID collisions and incorrect relationship resolution
