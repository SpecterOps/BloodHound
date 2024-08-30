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

truncate table kind;
truncate table edge;
truncate table node;

insert into kind (id, name)
values (1, 'NodeKind1'),
       (2, 'NodeKind2'),
       (11, 'EdgeKind1'),
       (12, 'EdgeKind2'),
       (13, 'EdgeKind3');

insert into node (id, graph_id, kind_ids, properties)
values (1, 1, array [1], '{"name": "n1"}'),
       (2, 1, array [1, 2], '{"name": "n2"}'),
       (3, 1, array [1, 2], '{"name": "n3"}'),
       (4, 1, array [2], '{"name": "n4"}'),
       (5, 1, array [2], '{"name": "n5"}'),
       (6, 1, array [1, 2], '{"name": "n6"}');

insert into edge (graph_id, start_id, end_id, kind_id, properties)
values (1, 1, 2, 11, '{"name": "e1", "prop": "a"}'),
       (1, 2, 3, 12, '{"name": "e2", "prop": "a"}'),
       (1, 3, 4, 12, '{"name": "e3", "prop": "a"}'),
       (1, 4, 5, 12, '{"name": "e4", "prop": "a"}'),
       (1, 2, 4, 11, '{"name": "e5", "prop": "a"}'),
       (1, 2, 4, 13, '{"name": "e6", "prop": "a"}'),
       (1, 3, 4, 11, '{"name": "e7"}');
