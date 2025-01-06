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

-- Add PG dual ingest FF
INSERT INTO feature_flags (created_at, updated_at, key, name, description, enabled, user_updatable) VALUES (current_timestamp, current_timestamp, 'pg_migration_dual_ingest', 'PostgreSQL Migration Dual Ingest', 'Enables dual ingest pathing for both Neo4j and PostgreSQL.', false, false) ON CONFLICT DO NOTHING;
