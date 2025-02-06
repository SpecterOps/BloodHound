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

-- case: with '1' as target match (n:NodeKind1) where n.value = target return n
with s0 as (select '1' as i0),
     s1 as (select s0.i0 as i0, (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from s0,
                 node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
              and n0.properties ->> 'value' = s0.i0)
select s1.n0 as n
from s1;

-- case: match (n:NodeKind1) where n.value = 1 with n match (b) where id(b) = id(n) return b
with s0 as (with s1 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
                        from node n0
                        where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
                          and (n0.properties ->> 'value')::int8 = 1)
            select s1.n0 as n0
            from s1),
     s2 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where n1.id = (s0.n0).id)
select s2.n1 as b
from s2;

-- case: match (n:NodeKind1) where n.value = 1 with n match (f) where f.name = 'me' with f match (b) where id(b) = id(f) return b
with s0 as (with s1 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
                        from node n0
                        where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
                          and (n0.properties ->> 'value')::int8 = 1)
            select s1.n0 as n0
            from s1),
     s2 as (with s3 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
                        from node n1
                        where n1.properties ->> 'name' = 'me')
            select s3.n1 as n1
            from s3),
     s4 as (select s2.n1 as n1, (n2.id, n2.kind_ids, n2.properties)::nodecomposite as n2
            from s2,
                 node n2
            where n2.id = (s2.n1).id)
select s4.n2 as b
from s4;

-- case: match (n:NodeKind1)-[:EdgeKind1*1..]->(:NodeKind2)-[:EdgeKind2]->(m:NodeKind1) where (n:NodeKind1 or n:NodeKind2) and n.enabled = true with m, collect(distinct(n)) as p where size(p) >= 10 return m
with s0 as (with s1 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                                          e0.end_id,
                                                                                                          1,
                                                                                                          n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
                                                                                                          e0.start_id = e0.end_id,
                                                                                                          array [e0.id]
                                                                                                   from edge e0
                                                                                                          join node n0
                                                                                                               on
                                                                                                                 n0.kind_ids operator (pg_catalog.&&)
                                                                                                                 array [1]::int2[] and
                                                                                                                 (n0.kind_ids operator (pg_catalog.&&)
                                                                                                                  array [1]::int2[] or
                                                                                                                  n0.kind_ids operator (pg_catalog.&&)
                                                                                                                  array [2]::int2[]) and
                                                                                                                 (n0.properties ->> 'enabled')::bool =
                                                                                                                 true and
                                                                                                                 n0.id =
                                                                                                                 e0.start_id
                                                                                                          join node n1 on n1.id = e0.end_id
                                                                                                   where e0.kind_id = any (array [3]::int2[])
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
                        where ex0.satisfied),
                 s2 as (select s1.e0                                                                     as e0,
                               s1.ep0                                                                    as ep0,
                               s1.n0                                                                     as n0,
                               s1.n1                                                                     as n1,
                               (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                               (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
                        from s1,
                             edge e1
                               join node n2
                                    on n2.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n2.id = e1.end_id
                        where e1.kind_id = any (array [4]::int2[])
                          and (s1.n1).id = e1.start_id)
            select s2.n2 as n2, array_agg(distinct (n0))::nodecomposite[] as i0
            from s2
            group by n2)
select s0.n2 as m
from s0
where array_length(s0.i0, 1)::int >= 10;

-- case: match (n:NodeKind1)-[:EdgeKind1*1..]->(:NodeKind2)-[:EdgeKind2]->(m:NodeKind1) where (n:NodeKind1 or n:NodeKind2) and n.enabled = true with m, count(distinct(n)) as p where p >= 10 return m
with s0 as (with s1 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                                          e0.end_id,
                                                                                                          1,
                                                                                                          n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
                                                                                                          e0.start_id = e0.end_id,
                                                                                                          array [e0.id]
                                                                                                   from edge e0
                                                                                                          join node n0
                                                                                                               on
                                                                                                                 n0.kind_ids operator (pg_catalog.&&)
                                                                                                                 array [1]::int2[] and
                                                                                                                 (n0.kind_ids operator (pg_catalog.&&)
                                                                                                                  array [1]::int2[] or
                                                                                                                  n0.kind_ids operator (pg_catalog.&&)
                                                                                                                  array [2]::int2[]) and
                                                                                                                 (n0.properties ->> 'enabled')::bool =
                                                                                                                 true and
                                                                                                                 n0.id =
                                                                                                                 e0.start_id
                                                                                                          join node n1 on n1.id = e0.end_id
                                                                                                   where e0.kind_id = any (array [3]::int2[])
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
                        where ex0.satisfied),
                 s2 as (select s1.e0                                                                     as e0,
                               s1.ep0                                                                    as ep0,
                               s1.n0                                                                     as n0,
                               s1.n1                                                                     as n1,
                               (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                               (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
                        from s1,
                             edge e1
                               join node n2
                                    on n2.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n2.id = e1.end_id
                        where e1.kind_id = any (array [4]::int2[])
                          and (s1.n1).id = e1.start_id)
            select s2.n2 as n2, count((n0))::int8 as i0
            from s2
            group by n2)
select s0.n2 as m
from s0
where s0.i0 >= 10;

-- case: with 365 as max_days match (n:NodeKind1) where n.pwdlastset < (datetime().epochseconds - (max_days * 86400)) and not n.pwdlastset IN [-1.0, 0.0] return n limit 100
with s0 as (select 365 as i0),
     s1 as (select s0.i0 as i0, (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from s0,
                 node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
              and not (n0.properties ->> 'pwdlastset')::float8 = any (array [- 1, 0]::float8[])
              and (n0.properties ->> 'pwdlastset')::numeric <
                  (extract(epoch from now()::timestamp with time zone)::numeric - (s0.i0 * 86400)))
select s1.n0 as n
from s1
limit 100;

-- case: match (n:NodeKind1) where n.hasspn = true and n.enabled = true and not n.objectid ends with '-502' and not coalesce(n.gmsa, false) = true and not coalesce(n.msa, false) = true match (n)-[:EdgeKind1|EdgeKind2*1..]->(c:NodeKind2) with distinct n, count(c) as adminCount return n order by adminCount desc limit 100
with s0 as (with s1 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
                        from node n0
                        where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
                          and (n0.properties ->> 'hasspn')::bool = true
                          and (n0.properties ->> 'enabled')::bool = true
                          and not coalesce(n0.properties ->> 'objectid', '')::text like '%-502'
                          and not coalesce((n0.properties ->> 'gmsa')::bool, false)::bool = true
                          and not coalesce((n0.properties ->> 'msa')::bool, false)::bool = true),
                 s2 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                                          e0.end_id,
                                                                                                          1,
                                                                                                          n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
                                                                                                          e0.start_id = e0.end_id,
                                                                                                          array [e0.id]
                                                                                                   from s1
                                                                                                          join edge e0 on e0.start_id = (s1.n0).id
                                                                                                          join node n0 on n0.id = e0.start_id
                                                                                                          join node n1 on n1.id = e0.end_id
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
            select s2.n0 as n0, count(n1)::int8 as i0
            from s2
            group by n0)
select s0.n0 as n
from s0
order by i0 desc
limit 100;

-- case: match (n:NodeKind1) where n.objectid = 'S-1-5-21-1260426776-3623580948-1897206385-23225' match p = (n)-[:EdgeKind1|EdgeKind2*1..]->(c:NodeKind2) return p
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
              and n0.properties ->> 'objectid' = 'S-1-5-21-1260426776-3623580948-1897206385-23225'),
     s1 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                              e0.end_id,
                                                                                              1,
                                                                                              n1.kind_ids operator (pg_catalog.&&) array [2]::int2[],
                                                                                              e0.start_id = e0.end_id,
                                                                                              array [e0.id]
                                                                                       from s0
                                                                                              join edge e0 on e0.start_id = (s0.n0).id
                                                                                              join node n0 on n0.id = e0.start_id
                                                                                              join node n1 on n1.id = e0.end_id
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
from s1;

-- todo: match (dc)-[r:EdgeKind1*0..]->(g:NodeKind1) where g.objectid ends with '-516' with collect(dc) as exclude match p = (c:NodeKind2)-[n:EdgeKind2]->(u:NodeKind2)-[:EdgeKind2*1..]->(g:NodeKind1) where g.objectid ends with '-512' and not c in exclude return p limit 100
;
