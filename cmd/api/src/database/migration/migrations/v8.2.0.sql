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
-- Add OpenGraph Phase 2 feature flag
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable)
VALUES (
           current_timestamp,
           current_timestamp,
           'open_graph_phase_2',
           'Open Graph Phase 2',
           'Open Graph Phase 2 features',
           false,
           false
       )
ON CONFLICT DO NOTHING;

-- Add Stale Client Updated Logic rework parameter
INSERT INTO parameters (key, name, description, value, created_at, updated_at)
VALUES (
         'pipeline.updated_stale_client',
        'Stale Client Updated Logic',
        'Is used to updated the logic used for if a job has become stale. With this enabled, rather than checking the last ingest time, the last checkin time of the client is checked to timeout the job.',
        '{"enabled": true}',
           current_timestamp,
           current_timestamp,
       )
  ON CONFLICT DO NOTHING;
