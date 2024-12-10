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

-- case: match p = allShortestPaths((s:NodeKind1)-[*..]->()) return p
-- cypher_params: {}
-- pgsql_params: {"pi0":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select e0.start_id, e0.end_id, 1, exists (select 1 from edge e0 where n1.id = e0.start_id), e0.start_id = e0.end_id, array [e0.id] from edge e0 join node n0 on n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.id = e0.start_id join node n1 on n1.id = e0.end_id;", "pi1":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select ex0.root_id, e0.end_id, ex0.depth + 1, exists (select 1 from edge e0 where n1.id = e0.start_id), e0.id = any (ex0.path), ex0.path || e0.id from pathspace ex0 join edge e0 on e0.start_id = ex0.next_id join node n1 on n1.id = e0.end_id where ex0.depth < 5 and not ex0.is_cycle;"}
with s0 as (with ex0(root_id, next_id, depth, satisfied, is_cycle, path)
                   as (select * from asp_harness(@pi0::text, @pi1::text, 5))
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                   join edge e0 on e0.id = any (ex0.path)
                   join node n0 on n0.id = ex0.root_id
                   join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int4] and n1.id = e0.end_id)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0;

-- case: match p = allShortestPaths((s:NodeKind1)-[*..]->(e)) where e.name = '123' return p
-- cypher_params: {}
-- pgsql_params: {"pi0":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select e0.start_id, e0.end_id, 1, n1.properties ->> 'name' = '123', e0.start_id = e0.end_id, array [e0.id] from edge e0 join node n0 on n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.id = e0.start_id join node n1 on n1.id = e0.end_id;", "pi1":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select ex0.root_id, e0.end_id, ex0.depth + 1, n1.properties ->> 'name' = '123', e0.id = any (ex0.path), ex0.path || e0.id from pathspace ex0 join edge e0 on e0.start_id = ex0.next_id join node n1 on n1.id = e0.end_id where ex0.depth < 5 and not ex0.is_cycle;"}
with s0 as (with ex0(root_id, next_id, depth, satisfied, is_cycle, path)
                   as (select * from asp_harness(@pi0::text, @pi1::text, 5))
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                   join edge e0 on e0.id = any (ex0.path)
                   join node n0 on n0.id = ex0.root_id
                   join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int4] and n1.id = e0.end_id
            where ex0.satisfied)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0;

-- case: match p=shortestPath((n:NodeKind1)-[:EdgeKind1*1..]->(m)) where 'admin_tier_0' in split(m.system_tags, ' ') and n.objectid ends with '-513' and n<>m return p limit 1000
-- cypher_params: {}
-- pgsql_params: {"pi0":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select e0.start_id, e0.end_id, 1, 'admin_tier_0' = any (string_to_array(n1.properties ->> 'system_tags', ' ')::text[]), e0.start_id = e0.end_id, array [e0.id] from edge e0 join node n0 on n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.properties ->> 'objectid' like '%-513' and n0.id = e0.start_id join node n1 on n1.id = e0.end_id where e0.kind_id = any (array [11]::int2[]);", "pi1":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select ex0.root_id, e0.end_id, ex0.depth + 1, 'admin_tier_0' = any (string_to_array(n1.properties ->> 'system_tags', ' ')::text[]), e0.id = any (ex0.path), ex0.path || e0.id from pathspace ex0 join edge e0 on e0.start_id = ex0.next_id join node n1 on n1.id = e0.end_id where ex0.depth < 5 and not ex0.is_cycle and e0.kind_id = any (array [11]::int2[]);"}
with s0 as (with ex0(root_id, next_id, depth, satisfied, is_cycle, path)
                   as (select * from asp_harness(@pi0::text, @pi1::text, 5))
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                   join edge e0 on e0.id = any (ex0.path)
                   join node n0 on n0.id = ex0.root_id
                   join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int4] and n1.id = e0.end_id
            where ex0.satisfied
              and n0.id <> n1.id)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0
limit 1000;
