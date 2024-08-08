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

-- case: match (n) return labels(n)
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select (s0.n0).kind_ids
from s0;

-- case: match (n) where ID(n) = 1 return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.id = 1)
select s0.n0 as n
from s0;

-- case: match (n) where n.objectid in $p return n
-- cypher_params: {"p": ["1", "2", "3"]}
-- pgsql_params: {"pi0": ["1", "2", "3"]}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'objectid' = any (@pi0::text[]))
select s0.n0 as n
from s0;

-- case: match (s) where s.name = $myParam return s
-- cypher_params: {"myParam": "123"}
-- pgsql_params: {"pi0": "123"}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = @pi0::text)
select s0.n0 as s
from s0;

-- case: match (s) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as s
from s0;

-- case: match (s) where s.prop = [1, 2, 3] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where jsonb_to_text_array(n0.properties -> 'prop')::int8[] = array [1, 2, 3]::int8[])
select s0.n0 as s
from s0;

-- case: match (s) where (s:NodeKind1 or s:NodeKind2) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] or
                   n0.kind_ids operator (pg_catalog.&&) array [2]::int2[]))
select s0.n0 as s
from s0;

-- case: match (n:NodeKind1), (e) where n.name = e.name return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where (s0.n0).properties -> 'name' = n1.properties -> 'name')
select s1.n0 as n
from s1;

-- case: match (s), (e) where id(s) in e.captured_ids return s, e
--
-- This is a little weird for us since JSONB arrays are basically type any[] which requires a special form of
-- type negotiation in PgSQL
--
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where (s0.n0).id = any (jsonb_to_text_array(n1.properties -> 'captured_ids')::int4[]))
select s1.n0 as s, s1.n1 as e
from s1;

-- case: match (s) where s:NodeKind1 and s:NodeKind2 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
              and n0.kind_ids operator (pg_catalog.&&) array [2]::int2[])
select s0.n0 as s
from s0;

-- case: match (s) where s.name = '1234' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = '1234')
select s0.n0 as s
from s0;

-- case: match (s:NodeKind1), (e:NodeKind2) where s.selected or s.tid = e.tid and e.enabled return s, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where n1.kind_ids operator (pg_catalog.&&) array [2]::int2[] and ((s0.n0).properties -> 'selected')::bool
               or (s0.n0).properties -> 'tid' = n1.properties -> 'tid' and (n1.properties -> 'enabled')::bool)
select s1.n0 as s, s1.n1 as e
from s1;

-- case: match (s) where s.value + 2 / 3 > 10 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties -> 'value')::int8 + 2 / 3 > 10)
select s0.n0 as s
from s0;

-- case: match (s), (e) where s.name = 'n1' return s, e.name as othername
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = 'n1'),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1)
select s1.n0 as s, (s1.n1).properties -> 'name' as othername
from s1;

-- case: match (s) where s.name in ['option 1', 'option 2'] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = any (array ['option 1', 'option 2']::text[]))
select s0.n0 as s
from s0;

-- case: match (s) where toLower(s.name) = '1234' return distinct s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where lower(n0.properties ->> 'name')::text = '1234')
select s0.n0 as s
from s0;

-- case: match (s:NodeKind1), (e:NodeKind2) where s.name = e.name return s, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where n1.kind_ids operator (pg_catalog.&&) array [2]::int2[]
              and (s0.n0).properties -> 'name' = n1.properties -> 'name')
select s1.n0 as s, s1.n1 as e
from s1;

-- case: match (n) where n.system_tags is not null and not (n:NodeKind1 or n:NodeKind2) return id(n)
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ? 'system_tags'
              and not (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] or
                       n0.kind_ids operator (pg_catalog.&&) array [2]::int2[]))
select (s0.n0).id
from s0;

-- case: match (s), (e) where s.name = '1234' and e.other = 1234 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = '1234'),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where (n1.properties -> 'other')::int8 = 1234)
select s1.n0 as s
from s1;

-- case: match (s), (e) where s.name = '1234' or e.other = 1234 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where (s0.n0).properties ->> 'name' = '1234'
               or (n1.properties -> 'other')::int8 = 1234)
select s1.n0 as s
from s1;

-- case: match (n), (k) where n.name = '1234' and k.name = '1234' match (e) where e.name = n.name return k, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = '1234'),
     s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where n1.properties ->> 'name' = '1234'),
     s2 as (select s1.n0 as n0, s1.n1 as n1, (n2.id, n2.kind_ids, n2.properties)::nodecomposite as n2
            from s1,
                 node n2
            where n2.properties -> 'name' = (s1.n0).properties -> 'name')
select s2.n1 as k, s2.n2 as e
from s2;

-- case: match (n) return n skip 5 limit 10
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as n
from s0
offset 5 limit 10;

-- case: match (s) return s order by s.name, s.other_prop desc
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as s
from s0
order by (s0.n0).properties -> 'name', (s0.n0).properties -> 'other_prop' desc;

-- case: match (s) where s.created_at = localtime() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::time without time zone = localtime(6)::time without time zone)
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at = localtime('12:12:12') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::time without time zone = ('12:12:12')::time without time zone)
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at = date() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::date = current_date::date)
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at = date('2023-12-12') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::date = ('2023-12-12')::date)
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at = datetime() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::timestamp with time zone = now()::timestamp with time zone)
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at = datetime('2019-06-01T18:40:32.142+0100') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::timestamp with time zone =
                  ('2019-06-01T18:40:32.142+0100')::timestamp with time zone)
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at = localdatetime() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::timestamp without time zone =
                  localtimestamp(6)::timestamp without time zone)
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at = localdatetime('2019-06-01T18:40:32.142') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where (n0.properties ->> 'created_at')::timestamp without time zone =
                  ('2019-06-01T18:40:32.142')::timestamp without time zone)
select s0.n0 as s
from s0;

-- case: match (s) where not (s.name = '123') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where not (n0.properties ->> 'name' = '123'))
select s0.n0 as s
from s0;

-- case: match (s) return s.value + 1
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select ((s0.n0).properties -> 'value')::int8 + 1
from s0;

-- case: match (s) return (s.value + 1) / 3
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select (((s0.n0).properties -> 'value')::int8 + 1) / 3
from s0;

-- case: match (s) where id(s) in [1, 2, 3, 4] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.id = any (array [1, 2, 3, 4]::int8[]))
select s0.n0 as s
from s0;

-- case: match (s) where s.name in ['option 1', 'option 2'] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = any (array ['option 1', 'option 2']::text[]))
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at is null return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where not n0.properties ? 'created_at')
select s0.n0 as s
from s0;

-- case: match (s) where s.created_at is not null return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ? 'created_at')
select s0.n0 as s
from s0;

-- case: match (s) where s.name starts with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' like '123%')
select s0.n0 as s
from s0;

-- case: match (s) where not s.name starts with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where not coalesce(n0.properties ->> 'name', '')::text like '123%')
select s0.n0 as s
from s0;

-- case: match (s) where s.name contains '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' like '%123%')
select s0.n0 as s
from s0;

-- case: match (s) where not s.name contains '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where not coalesce(n0.properties ->> 'name', '')::text like '%123%')
select s0.n0 as s
from s0;

-- case: match (s) where s.name ends with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' like '%123')
select s0.n0 as s
from s0;

-- case: match (s) where not s.name ends with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where not coalesce(n0.properties ->> 'name', '')::text like '%123')
select s0.n0 as s
from s0;

-- case: match (s) where s.name starts with s.other return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' like ((n0.properties ->> 'other') || '%')::text)
select s0.n0 as s
from s0;

-- case: match (s) where s.name contains s.other return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' like ('%' || (n0.properties ->> 'other') || '%')::text)
select s0.n0 as s
from s0;

-- case: match (s) where s.name ends with s.other return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' like ('%' || (n0.properties ->> 'other'))::text)
select s0.n0 as s
from s0;

-- case: match (n) where n:NodeKind1 and toLower(n.tenantid) contains 'myid' and n.system_tags contains 'tag' return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]
              and lower(n0.properties ->> 'tenantid')::text like '%myid%'
              and n0.properties ->> 'system_tags' like '%tag%')
select s0.n0 as n
from s0;

-- case: match (s) where not (s)-[]-() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as s
from s0
where not exists (select 1 from edge e0 where e0.start_id = (s0.n0).id or e0.end_id = (s0.n0).id);

-- case: match (s) where not (s)-[]->()-[]->() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as s
from s0
where not (with s1 as (select s0.n0                                                                     as n0,
                              (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                              (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
                       from s0,
                            edge e0
                              join node n0 on (s0.n0).id = e0.start_id
                              join node n1 on n1.id = e0.end_id),
                s2 as (select s1.e0                                                                     as e0,
                              s1.n0                                                                     as n0,
                              s1.n1                                                                     as n1,
                              (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                              (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
                       from s1,
                            edge e1
                              join node n2 on n2.id = e1.end_id
                       where (s1.n1).id = e1.start_id)
           select count(*) > 0
           from s2);


-- case: match (s) where not (s)-[{prop: 'a'}]-({name: 'n3'}) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as s
from s0
where not (with s1 as (select s0.n0                                                                     as n0,
                              (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                              (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
                       from edge e0
                              join node n0 on (s0.n0).id = e0.end_id
                              join node n1 on n1.properties ->> 'name' = 'n3' and n1.id = e0.start_id
                       where e0.properties ->> 'prop' = 'a'
                       union
                       distinct
                       select s0.n0                                                                     as n0,
                              (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                              (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
                       from edge e0
                              join node n0 on (s0.n0).id = e0.start_id
                              join node n1 on n1.properties ->> 'name' = 'n3' and n1.id = e0.end_id
                       where e0.properties ->> 'prop' = 'a')
           select count(*) > 0
           from s1);

-- case: match (s) where not (s)<-[{prop: 'a'}]-({name: 'n3'}) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as s
from s0
where not (with s1 as (select s0.n0                                                                     as n0,
                              (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                              (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
                       from s0,
                            edge e0
                              join node n0 on (s0.n0).id = e0.end_id
                              join node n1 on n1.properties ->> 'name' = 'n3' and n1.id = e0.start_id
                       where e0.properties ->> 'prop' = 'a')
           select count(*) > 0
           from s1);

-- case: match (s) where not (s)-[{prop: 'a'}]->({name: 'n3'}) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select s0.n0 as s
from s0
where not (with s1 as (select s0.n0                                                                     as n0,
                              (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                              (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
                       from s0,
                            edge e0
                              join node n0 on (s0.n0).id = e0.start_id
                              join node n1 on n1.properties ->> 'name' = 'n3' and n1.id = e0.end_id
                       where e0.properties ->> 'prop' = 'a')
           select count(*) > 0
           from s1);

-- case: match (s) where not (s)-[]-() return id(s)
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select (s0.n0).id
from s0
where not exists (select 1 from edge e0 where e0.start_id = (s0.n0).id or e0.end_id = (s0.n0).id);
