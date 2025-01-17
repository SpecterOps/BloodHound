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

-- case: match p = ()-[]->() return p
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.id = e0.start_id
                   join node n1 on n1.id = e0.end_id)
select edges_to_path(variadic array [(s0.e0).id]::int8[])::pathcomposite as p
from s0;

-- case: match p = ()-[r1]->()-[r2]->(e) return e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.id = e0.start_id
                   join node n1 on n1.id = e0.end_id),
     s1 as (select s0.e0                                                                     as e0,
                   s0.n0                                                                     as n0,
                   s0.n1                                                                     as n1,
                   (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
            from s0,
                 edge e1
                   join node n2 on n2.id = e1.end_id
            where (s0.n1).id = e1.start_id)
select s1.n2 as e
from s1;


-- case: match ()-[r1]->()-[r2]->()-[]->() where r1.name = 'a' and r2.name = 'b' return r1
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.id = e0.start_id
                   join node n1 on n1.id = e0.end_id
            where e0.properties ->> 'name' = 'a'),
     s1 as (select s0.e0                                                                     as e0,
                   s0.n0                                                                     as n0,
                   s0.n1                                                                     as n1,
                   (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
            from s0,
                 edge e1
                   join node n2 on n2.id = e1.end_id
            where e1.properties ->> 'name' = 'b'
              and (s0.n1).id = e1.start_id),
     s2 as (select s1.e0                                                                     as e0,
                   s1.e1                                                                     as e1,
                   s1.n0                                                                     as n0,
                   s1.n1                                                                     as n1,
                   s1.n2                                                                     as n2,
                   (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite as e2,
                   (n3.id, n3.kind_ids, n3.properties)::nodecomposite                        as n3
            from s1,
                 edge e2
                   join node n3 on n3.id = e2.end_id
            where (s1.n2).id = e2.start_id)
select s2.e0 as r1
from s2;

-- case: match p = (a)-[]->()<-[]-(f) where a.name = 'value' and f.is_target return p
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.properties ->> 'name' = 'value' and n0.id = e0.start_id
                   join node n1 on n1.id = e0.end_id),
     s1 as (select s0.e0                                                                     as e0,
                   s0.n0                                                                     as n0,
                   s0.n1                                                                     as n1,
                   (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
            from s0,
                 edge e1
                   join node n2 on (n2.properties ->> 'is_target')::bool and n2.id = e1.start_id
            where (s0.n1).id = e1.end_id)
select edges_to_path(variadic array [(s1.e0).id, (s1.e1).id]::int8[])::pathcomposite as p
from s1;

-- case: match p = ()-[*..]->() return p limit 1
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              false,
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              false,
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle)
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                   join edge e0 on e0.id = any (ex0.path)
                   join node n0 on n0.id = ex0.root_id
                   join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int] and n1.id = e0.end_id)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0
limit 1;

-- case: match p = (s)-[*..]->(i)-[]->() where id(s) = 1 and i.name = 'n3' return p limit 1
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.properties ->> 'name' = 'n3',
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on n0.id = 1 and n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.properties ->> 'name' = 'n3',
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle)
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                   join edge e0 on e0.id = any (ex0.path)
                   join node n0 on n0.id = ex0.root_id
                   join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int] and n1.id = e0.end_id
            where ex0.satisfied),
     s1 as (select s0.e0                                                                     as e0,
                   s0.ep0                                                                    as ep0,
                   s0.n0                                                                     as n0,
                   s0.n1                                                                     as n1,
                   (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
            from s0,
                 edge e1
                   join node n2 on n2.id = e1.end_id
            where (s0.n1).id = e1.start_id)
select edges_to_path(variadic array [(s1.e1).id]::int8[] || s1.ep0)::pathcomposite as p
from s1
limit 1;

-- case: match p = ()-[e:EdgeKind1]->()-[:EdgeKind1*..]->() return e, p
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.id = e0.start_id
                   join node n1 on n1.id = e0.end_id
            where e0.kind_id = any (array [3]::int2[])),
     s1 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e1.start_id,
                                                                                              e1.end_id,
                                                                                              1,
                                                                                              false,
                                                                                              e1.start_id = e1.end_id,
                                                                                              array [e1.id]
                                                                                       from s0
                                                                                              join edge e1
                                                                                                   on e1.kind_id = any (array [3]::int2[]) and (s0.n1).id = e1.start_id
                                                                                              join node n2 on n2.id = e1.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e1.end_id,
                                                                                              ex0.depth + 1,
                                                                                              false,
                                                                                              e1.id = any (ex0.path),
                                                                                              ex0.path || e1.id
                                                                                       from ex0
                                                                                              join edge e1 on e1.start_id = ex0.next_id
                                                                                              join node n2 on n2.id = e1.end_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle)
            select s0.e0                                              as e0,
                   s0.n0                                              as n0,
                   s0.n1                                              as n1,
                   (select array_agg((e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite)
                    from edge e1
                    where e1.id = any (ex0.path))                     as e1,
                   ex0.path                                           as ep0,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite as n2
            from s0,
                 ex0
                   join edge e1 on e1.id = any (ex0.path)
                   join node n1 on n1.id = ex0.root_id
                   join node n2 on e1.id = ex0.path[array_length(ex0.path, 1)::int] and n2.id = e1.end_id)
select s1.e0 as e, edges_to_path(variadic array [(s1.e0).id]::int8[] || s1.ep0)::pathcomposite as p
from s1;

-- case: match p = (m:NodeKind1)-[:EdgeKind1]->(c:NodeKind2) where m.objectid ends with "-513" and not toUpper(c.operatingsystem) contains "SERVER" return p limit 1000
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and
                                   n0.properties ->> 'objectid' like '%-513' and n0.id = e0.start_id
                   join node n1 on n1.kind_ids operator (pg_catalog.&&) array [2]::int2[] and
                                   not upper(n1.properties ->> 'operatingsystem')::text like '%SERVER%' and
                                   n1.id = e0.end_id
            where e0.kind_id = any (array [3]::int2[]))
select edges_to_path(variadic array [(s0.e0).id]::int8[])::pathcomposite as p
from s0
limit 1000;

-- case: match p = (:NodeKind1)-[:EdgeKind1|EdgeKind2]->(e:NodeKind2)-[:EdgeKind2]->(:NodeKind1) where 'a' in e.values or 'b' in e.values or size(e.values) = 0 return p
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.id = e0.start_id
                   join node n1 on n1.kind_ids operator (pg_catalog.&&) array [2]::int2[] and
                                   'a' = any (jsonb_to_text_array(n1.properties -> 'values')::text[]) or
                                   'b' = any (jsonb_to_text_array(n1.properties -> 'values')::text[]) or
                                   jsonb_array_length(n1.properties -> 'values')::int = 0 and n1.id = e0.end_id
            where e0.kind_id = any (array [3, 4]::int2[])),
     s1 as (select s0.e0                                                                     as e0,
                   s0.n0                                                                     as n0,
                   s0.n1                                                                     as n1,
                   (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
            from s0,
                 edge e1
                   join node n2 on n2.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n2.id = e1.end_id
            where e1.kind_id = any (array [4]::int2[])
              and (s0.n1).id = e1.start_id)
select edges_to_path(variadic array [(s1.e0).id, (s1.e1).id]::int8[])::pathcomposite as p
from s1;

-- todo: the case below covers untyped array literals but has not yet been fixed
-- case: match p = (:NodeKind1)-[:EdgeKind1|EdgeKind2]->(e:NodeKind2)-[:EdgeKind2]->(:NodeKind1) where (e.a = [] or 'a' in e.a) and (e.b = 0 or e.b = 1) return p
