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

-- case: match (n)-[*..]->(e) return n, e
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
select s0.n0 as n, s0.n1 as e
from s0;

-- case: match (n)-[*..]->(e:NodeKind1) where n.name = 'n1' return e
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on n0.properties ->> 'name' = 'n1' and n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
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
            where ex0.satisfied)
select s0.n1 as e
from s0;

-- todo: cypher expects the right hand binding for `r` to already be a list of relationships which seems strange
-- case: match (n)-[r*..]->(e:NodeKind1) where n.name = 'n1' and r.prop = 'a' return e
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on n0.properties ->> 'name' = 'n1' and n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where e0.properties ->> 'prop' = 'a'
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle
                                                                                         and e0.properties ->> 'prop' = 'a')
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
            where ex0.satisfied)
select s0.n1 as e
from s0;

-- case: match (n)-[*..]->(e:NodeKind1) where n.name = 'n2' return n
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on n0.properties ->> 'name' = 'n2' and n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
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
            where ex0.satisfied)
select s0.n0 as n
from s0;

-- case: match (n)-[*..]->(e:NodeKind1)-[]->(l) where n.name = 'n1' return l
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on n0.properties ->> 'name' = 'n1' and n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [1]::int2[],
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
select s1.n2 as l
from s1;

-- case: match (n)-[*..]->(e)-[:EdgeKind1|EdgeKind2]->()-[*..]->(l) where n.name = 'n1' and e.name = 'n2' return l
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.properties ->> 'name' = 'n2',
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on n0.properties ->> 'name' = 'n1' and n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.properties ->> 'name' = 'n2',
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
            where e1.kind_id = any (array [3, 4]::int2[])
              and (s0.n1).id = e1.start_id),
     s2 as (with recursive ex1(root_id, next_id, depth, satisfied, is_cycle, path) as (select e2.start_id,
                                                                                              e2.end_id,
                                                                                              1,
                                                                                              false,
                                                                                              e2.start_id = e2.end_id,
                                                                                              array [e2.id]
                                                                                       from s1
                                                                                              join edge e2 on (s1.n2).id = e2.start_id
                                                                                              join node n3 on n3.id = e2.end_id
                                                                                       union
                                                                                       select ex1.root_id,
                                                                                              e2.end_id,
                                                                                              ex1.depth + 1,
                                                                                              false,
                                                                                              e2.id = any (ex1.path),
                                                                                              ex1.path || e2.id
                                                                                       from ex1
                                                                                              join edge e2 on e2.start_id = ex1.next_id
                                                                                              join node n3 on n3.id = e2.end_id
                                                                                       where ex1.depth < 10
                                                                                         and not ex1.is_cycle)
            select s1.e0                                              as e0,
                   s1.e1                                              as e1,
                   s1.ep0                                             as ep0,
                   s1.n0                                              as n0,
                   s1.n1                                              as n1,
                   s1.n2                                              as n2,
                   (select array_agg((e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite)
                    from edge e2
                    where e2.id = any (ex1.path))                     as e2,
                   ex1.path                                           as ep1,
                   (n3.id, n3.kind_ids, n3.properties)::nodecomposite as n3
            from s1,
                 ex1
                   join edge e2 on e2.id = any (ex1.path)
                   join node n2 on n2.id = ex1.root_id
                   join node n3 on e2.id = ex1.path[array_length(ex1.path, 1)::int] and n3.id = e2.end_id)
select s2.n3 as l
from s2;

-- case: match p = (:NodeKind1)-[:EdgeKind1*1..]->(n:NodeKind2) where 'admin_tier_0' in split(n.system_tags, ' ') return p limit 1000
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&)
                                                                                              array [2]::int2[] and
                                                                                              'admin_tier_0' = any
                                                                                              (string_to_array(n1.properties ->> 'system_tags', ' ')::text[]),
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on
                                                                                         n0.kind_ids operator (pg_catalog.&&)
                                                                                         array [1]::int2[] and
                                                                                         n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where e0.kind_id = any (array [3]::int2[])
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&)
                                                                                              array [2]::int2[] and
                                                                                              'admin_tier_0' = any
                                                                                              (string_to_array(n1.properties ->> 'system_tags', ' ')::text[]),
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle
                                                                                         and e0.kind_id = any (array [3]::int2[]))
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
            where ex0.satisfied)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0
limit 1000;

-- case: match p = (s:NodeKind1)-[*..]->(e:NodeKind2) where s <> e return p
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on
                                                                                         n0.kind_ids operator (pg_catalog.&&)
                                                                                         array [1]::int2[] and
                                                                                         n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
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
            where ex0.satisfied
              and n0.id <> n1.id)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0;

-- case: match p = (g:NodeKind1)-[:EdgeKind1|EdgeKind2*]->(target:NodeKind1) where g.objectid ends with '1234' and target.objectid ends with '4567' return p
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&)
                                                                                              array [1]::int2[] and
                                                                                              n1.properties ->>
                                                                                              'objectid' like '%4567',
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on
                                                                                         n0.kind_ids operator (pg_catalog.&&)
                                                                                         array [1]::int2[] and
                                                                                         n0.properties ->>
                                                                                         'objectid' like '%1234' and
                                                                                         n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where e0.kind_id = any (array [3, 4]::int2[])
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&)
                                                                                              array [1]::int2[] and
                                                                                              n1.properties ->>
                                                                                              'objectid' like '%4567',
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle
                                                                                         and e0.kind_id = any (array [3, 4]::int2[]))
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
            where ex0.satisfied)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0;

-- case: match p = (:NodeKind1)<-[:EdgeKind1|EdgeKind2*..]-() return p limit 10
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              false,
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on
                                                                                         n0.kind_ids operator (pg_catalog.&&)
                                                                                         array [1]::int2[] and
                                                                                         n0.id = e0.end_id
                                                                                              join node n1 on n1.id = e0.start_id
                                                                                       where e0.kind_id = any (array [3, 4]::int2[])
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              false,
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.start_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle
                                                                                         and e0.kind_id = any (array [3, 4]::int2[]))
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                   join edge e0 on e0.id = any (ex0.path)
                   join node n0 on n0.id = ex0.root_id
                   join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int] and n1.id = e0.start_id)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0
limit 10;


-- case: match p = (:NodeKind1)<-[:EdgeKind1|EdgeKind2*..]-(:NodeKind2)<-[:EdgeKind1|EdgeKind2*..]-(:NodeKind1) return p limit 10
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on
                                                                                         n0.kind_ids operator (pg_catalog.&&)
                                                                                         array [1]::int2[] and
                                                                                         n0.id = e0.end_id
                                                                                              join node n1 on n1.id = e0.start_id
                                                                                       where e0.kind_id = any (array [3, 4]::int2[])
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.start_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle
                                                                                         and e0.kind_id = any (array [3, 4]::int2[]))
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                   join edge e0 on e0.id = any (ex0.path)
                   join node n0 on n0.id = ex0.root_id
                   join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int] and n1.id = e0.start_id
            where ex0.satisfied),
     s1 as (with recursive ex1(root_id, next_id, depth, satisfied, is_cycle, path) as (select e1.start_id,
                                                                                              e1.end_id,
                                                                                              1,
                                                                                              false,
                                                                                              e1.start_id = e1.end_id,
                                                                                              array [e1.id]
                                                                                       from s0
                                                                                              join edge e1 on
                                                                                         e1.kind_id = any
                                                                                         (array [3, 4]::int2[]) and
                                                                                         (s0.n1).id = e1.end_id
                                                                                              join node n2 on n2.id = e1.start_id
                                                                                       union
                                                                                       select ex1.root_id,
                                                                                              e1.end_id,
                                                                                              ex1.depth + 1,
                                                                                              n2.kind_ids operator (pg_catalog.&&) array [1]::int2[],
                                                                                              e1.id = any (ex1.path),
                                                                                              ex1.path || e1.id
                                                                                       from ex1
                                                                                              join edge e1 on e1.start_id = ex1.next_id
                                                                                              join node n2 on n2.id = e1.start_id
                                                                                       where ex1.depth < 10
                                                                                         and not ex1.is_cycle)
            select s0.e0                                              as e0,
                   s0.ep0                                             as ep0,
                   s0.n0                                              as n0,
                   s0.n1                                              as n1,
                   (select array_agg((e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite)
                    from edge e1
                    where e1.id = any (ex1.path))                     as e1,
                   ex1.path                                           as ep1,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite as n2
            from s0,
                 ex1
                   join edge e1 on e1.id = any (ex1.path)
                   join node n1 on n1.id = ex1.root_id
                   join node n2 on e1.id = ex1.path[array_length(ex1.path, 1)::int] and n2.id = e1.start_id)
select edges_to_path(variadic s1.ep1 || s1.ep0)::pathcomposite as p
from s1
limit 10;

-- case: match p = (n:NodeKind1)-[:EdgeKind1|EdgeKind2*1..2]->(r:NodeKind2) where r.name =~ '(?i)Global Administrator.*|User Administrator.*|Cloud Application Administrator.*|Authentication Policy Administrator.*|Exchange Administrator.*|Helpdesk Administrator.*|Privileged Authentication Administrator.*' return p limit 10
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&)
                                                                                              array [2]::int2[] and
                                                                                              n1.properties ->> 'name' ~
                                                                                              '(?i)Global Administrator.*|User Administrator.*|Cloud Application Administrator.*|Authentication Policy Administrator.*|Exchange Administrator.*|Helpdesk Administrator.*|Privileged Authentication Administrator.*',
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from edge e0
                                                                                              join node n0 on
                                                                                         n0.kind_ids operator (pg_catalog.&&)
                                                                                         array [1]::int2[] and
                                                                                         n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where e0.kind_id = any (array [3, 4]::int2[])
                                                                                       union
                                                                                       select ex0.root_id,
                                                                                              e0.end_id,
                                                                                              ex0.depth + 1,
                                                                                              n1.kind_ids operator (pg_catalog.&&)
                                                                                              array [2]::int2[] and
                                                                                              n1.properties ->> 'name' ~
                                                                                              '(?i)Global Administrator.*|User Administrator.*|Cloud Application Administrator.*|Authentication Policy Administrator.*|Exchange Administrator.*|Helpdesk Administrator.*|Privileged Authentication Administrator.*',
                                                                                              e0.id = any (ex0.path),
                                                                                              ex0.path || e0.id
                                                                                       from ex0
                                                                                              join edge e0 on e0.start_id = ex0.next_id
                                                                                              join node n1 on n1.id = e0.end_id
                                                                                       where ex0.depth < 10
                                                                                         and not ex0.is_cycle
                                                                                         and e0.kind_id = any (array [3, 4]::int2[]))
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
            where ex0.satisfied)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0
limit 10;
