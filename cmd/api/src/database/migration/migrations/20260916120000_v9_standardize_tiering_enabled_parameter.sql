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
-- Standardize the analysis.tiering parameter to use the "enabled" key like the other configuration
-- parameters, replacing the legacy "multi_tier_analysis_enabled" key while preserving its value.
UPDATE parameters
SET value = (value - 'multi_tier_analysis_enabled')
    || jsonb_build_object('enabled', COALESCE((value ->> 'multi_tier_analysis_enabled')::boolean, false)),
    updated_at = current_timestamp
WHERE key = 'analysis.tiering'
  AND value ? 'multi_tier_analysis_enabled';

-- +goose Down
UPDATE parameters
SET value = (value - 'enabled')
    || jsonb_build_object('multi_tier_analysis_enabled', COALESCE((value ->> 'enabled')::boolean, false)),
    updated_at = current_timestamp
WHERE key = 'analysis.tiering'
  AND value ? 'enabled';
