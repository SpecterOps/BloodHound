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

-- insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select e.start_id, e.end_id, 1, e.end_id = 2579859, e.start_id = e.end_id, array [e.id] from edge e where e.start_id = 6746823;
-- insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select r.root_id, e.end_id, 1, e.end_id = 2579859, e.id = any (r.path), r.path || e.id from root_front r join edge e on e.start_id = r.next_id;
-- insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select e.end_id, e.start_id, 1, e.start_id = 6746823, e.start_id = e.end_id, array [e.id] from edge e where e.end_id = 2579859;
-- insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select t.root_id, e.start_id, 1, e.start_id = 6746823, e.id = any (t.path), e.id || t.path from terminal_front t  join edge e on e.end_id = t.next_id;

select *
from bidirectional_asp_harness(
  'insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select e.start_id, e.end_id, 1, e.end_id = 2579859, e.start_id = e.end_id, array [e.id] from edge e where e.start_id = 6746823;',
  'insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select r.root_id, e.end_id, 1, e.end_id = 2579859, e.id = any (r.path), r.path || e.id from root_front r join edge e on e.start_id = r.next_id;',
  'insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select e.end_id, e.start_id, 1, e.start_id = 6746823, e.start_id = e.end_id, array [e.id] from edge e where e.end_id = 2579859;',
  'insert into next_front (root_id, next_id, depth, satisfied, is_cycle, path) select t.root_id, e.start_id, 1, e.start_id = 6746823, e.id = any (t.path), e.id || t.path from terminal_front t  join edge e on e.end_id = t.next_id;',
  4);
