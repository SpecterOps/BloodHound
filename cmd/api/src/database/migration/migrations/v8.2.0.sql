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

CREATE INDEX IF NOT EXISTS idx_agt_history_actor ON asset_group_history USING btree (actor);
CREATE INDEX IF NOT EXISTS idx_agt_history_action ON asset_group_history USING btree (action);
CREATE INDEX IF NOT EXISTS idx_agt_history_target ON asset_group_history USING btree (target);
CREATE INDEX IF NOT EXISTS idx_agt_history_email ON asset_group_history USING btree (email);
CREATE INDEX IF NOT EXISTS idx_agt_history_env_id ON asset_group_history USING btree (environment_id);
