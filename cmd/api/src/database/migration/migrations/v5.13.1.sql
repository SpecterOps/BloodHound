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

ALTER TABLE IF EXISTS saved_queries
  ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';

CREATE TABLE IF NOT EXISTS saved_queries_permissions
(
  id                BIGSERIAL PRIMARY KEY,
  shared_to_user_id TEXT REFERENCES users (id) ON DELETE CASCADE DEFAULT NULL,
  query_id          BIGSERIAL REFERENCES saved_queries (id) ON DELETE CASCADE  NOT NULL,
  public            BOOL                                         DEFAULT FALSE NOT NULL,
  created_at        TIMESTAMP WITH TIME ZONE                     DEFAULT now(),
  updated_at        TIMESTAMP WITH TIME ZONE                     DEFAULT now(),
  UNIQUE (shared_to_user_id, query_id)
);
