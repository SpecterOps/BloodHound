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
create table if not exists source_kinds
(
  id   smallserial,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

INSERT INTO source_kinds (name)
VALUES 
  ('Base'),
  ('AZBase')
ON CONFLICT (name) DO NOTHING;

ALTER TABLE analysis_request_switch
ADD COLUMN IF NOT EXISTS delete_all_graph boolean DEFAULT false,
ADD COLUMN IF NOT EXISTS delete_all_open_graph boolean DEFAULT false,
ADD COLUMN IF NOT EXISTS delete_source_kinds text[] DEFAULT ARRAY[]::text[];

