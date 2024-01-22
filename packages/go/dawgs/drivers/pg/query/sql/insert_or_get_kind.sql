-- Copyright 2023 Specter Ops, Inc.
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

-- Creates a new kind definition if it does not exist and returns the resulting ID. If the
-- kind already exists then the kind's assigned ID is returned.
with
  existing as (
    select id from kind where kind.name = @name
  ),
  inserted as (
    insert into kind (name) values (@name) on conflict (name) do nothing returning id
  )
select * from existing
union
select * from inserted;
