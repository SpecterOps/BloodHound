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

-- case: match (n) return labels(n)
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select (s0.n0).kind_ids from s0;

-- case: match (n) where ID(n) = 1 return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.id = 1)) select s0.n0 as n from s0;

-- case: match (n) where coalesce(n.name, '') = '1234' return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (coalesce(n0.properties ->> 'name', '')::text = '1234')) select s0.n0 as n from s0;

-- case: match (n) where n.name = '1234' return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = '1234')) select s0.n0 as n from s0;

-- case: match (n:NodeKind1 {name: "SOME NAME"}) return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.properties ->> 'name' = 'SOME NAME') select s0.n0 as n from s0;

-- case: match (n) where n.objectid in $p return n
-- cypher_params: {"p":["1","2","3"]}
-- pgsql_params:{"pi0":["1","2","3"]}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'objectid' = any (@pi0::text[]))) select s0.n0 as n from s0;

-- case: match (s) where s.name = $myParam return s
-- cypher_params: {"myParam":"123"}
-- pgsql_params:{"pi0":"123"}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = @pi0::text)) select s0.n0 as s from s0;

-- case: match (s) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as s from s0;

-- case: match (s) where s.prop = [1, 2, 3] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (jsonb_to_text_array(n0.properties -> 'prop')::int8[] = array [1, 2, 3]::int8[])) select s0.n0 as s from s0;

-- case: match (s) where (s:NodeKind1 or s:NodeKind2) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] or n0.kind_ids operator (pg_catalog.&&) array [2]::int2[]))) select s0.n0 as s from s0;

-- case: match (n:NodeKind1), (e) where n.name = e.name return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where ((s0.n0).properties -> 'name' = n1.properties -> 'name')) select s1.n0 as n from s1;

-- case: match (s), (e) where id(s) in e.captured_ids return s, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where ((s0.n0).id = any (jsonb_to_text_array(n1.properties -> 'captured_ids')::int8[]))) select s1.n0 as s, s1.n1 as e from s1;

-- case: match (s) where s:NodeKind1 and s:NodeKind2 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.kind_ids operator (pg_catalog.&&) array [2]::int2[])) select s0.n0 as s from s0;

-- case: match (s) where s.name = '1234' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = '1234')) select s0.n0 as s from s0;

-- case: match (s:NodeKind1), (e:NodeKind2) where s.selected or s.tid = e.tid and e.enabled return s, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where (((s0.n0).properties ->> 'selected')::bool or (s0.n0).properties -> 'tid' = n1.properties -> 'tid' and (n1.properties ->> 'enabled')::bool) and n1.kind_ids operator (pg_catalog.&&) array [2]::int2[]) select s1.n0 as s, s1.n1 as e from s1;

-- case: match (s) where s.value + 2 / 3 > 10 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'value')::int8 + 2 / 3 > 10)) select s0.n0 as s from s0;

-- case: match (s), (e) where s.name = 'n1' return s, e.name as othername
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = 'n1')), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1) select s1.n0 as s, (s1.n1).properties -> 'name' as othername from s1;

-- case: match (s) where s.name in ['option 1', 'option 2'] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = any (array ['option 1', 'option 2']::text[]))) select s0.n0 as s from s0;

-- case: match (s) where toLower(s.name) = '1234' return distinct s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (lower(n0.properties ->> 'name')::text = '1234')) select s0.n0 as s from s0;

-- case: match (s:NodeKind1), (e:NodeKind2) where s.name = e.name return s, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where ((s0.n0).properties -> 'name' = n1.properties -> 'name') and n1.kind_ids operator (pg_catalog.&&) array [2]::int2[]) select s1.n0 as s, s1.n1 as e from s1;

-- case: match (n) where n.system_tags is not null and not (n:NodeKind1 or n:NodeKind2) return id(n)
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ? 'system_tags' and not (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] or n0.kind_ids operator (pg_catalog.&&) array [2]::int2[]))) select (s0.n0).id from s0;

-- case: match (s), (e) where s.name = '1234' and e.other = 1234 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = '1234')), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where ((n1.properties ->> 'other')::int8 = 1234)) select s1.n0 as s from s1;

-- case: match (s), (e) where s.name = '1234' or e.other = 1234 return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where ((s0.n0).properties ->> 'name' = '1234' or (n1.properties ->> 'other')::int8 = 1234)) select s1.n0 as s from s1;

-- case: match (n), (k) where n.name = '1234' and k.name = '1234' match (e) where e.name = n.name return k, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = '1234')), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where (n1.properties ->> 'name' = '1234')), s2 as (select s1.n0 as n0, s1.n1 as n1, (n2.id, n2.kind_ids, n2.properties)::nodecomposite as n2 from s1, node n2 where (n2.properties -> 'name' = (s1.n0).properties -> 'name')) select s2.n1 as k, s2.n2 as e from s2;

-- case: match (n) return n skip 5 limit 10
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as n from s0 offset 5 limit 10;

-- case: match (s) return s order by s.name, s.other_prop desc
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as s from s0 order by (s0.n0).properties -> 'name', (s0.n0).properties -> 'other_prop' desc;

-- case: match (s) where s.created_at = localtime() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::time without time zone = localtime(6)::time without time zone)) select s0.n0 as s from s0;

-- case: match (s) where s.created_at = localtime('4:4:4') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::time without time zone = ('4:4:4')::time without time zone)) select s0.n0 as s from s0;

-- case: match (s) where s.created_at = date() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::date = current_date::date)) select s0.n0 as s from s0;

-- case: match (s) where s.created_at = date('2023-4-4') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::date = ('2023-4-4')::date)) select s0.n0 as s from s0;

-- case: match (s) where s.created_at = datetime() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::timestamp with time zone = now()::timestamp with time zone)) select s0.n0 as s from s0;

-- case: match (s) where s.created_at = datetime('2019-06-01T18:40:32.142+0100') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::timestamp with time zone = ('2019-06-01T18:40:32.142+0100')::timestamp with time zone)) select s0.n0 as s from s0;

-- case: match (s) where s.created_at = localdatetime() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::timestamp without time zone = localtimestamp(6)::timestamp without time zone)) select s0.n0 as s from s0;

-- case: match (s) where s.created_at = localdatetime('2019-06-01T18:40:32.142') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'created_at')::timestamp without time zone = ('2019-06-01T18:40:32.142')::timestamp without time zone)) select s0.n0 as s from s0;

-- case: match (s) where not (s.name = '123') return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (not (n0.properties ->> 'name' = '123'))) select s0.n0 as s from s0;

-- case: match (s) return s.value + 1
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select ((s0.n0).properties ->> 'value')::int8 + 1 from s0;

-- case: match (s) return (s.value + 1) / 3
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select (((s0.n0).properties ->> 'value')::int8 + 1) / 3 from s0;

-- case: match (s) where id(s) in [1, 2, 3, 4] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.id = any (array [1, 2, 3, 4]::int8[]))) select s0.n0 as s from s0;

-- case: match (s) where s.name in ['option 1', 'option 2'] return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' = any (array ['option 1', 'option 2']::text[]))) select s0.n0 as s from s0;

-- case: match (s) where s.created_at is null return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (not n0.properties ? 'created_at')) select s0.n0 as s from s0;

-- case: match (s) where s.created_at is not null return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ? 'created_at')) select s0.n0 as s from s0;

-- case: match (s) where s.name starts with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' like '123%')) select s0.n0 as s from s0;

-- case: match (s) where not s.name starts with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (not coalesce(n0.properties ->> 'name', '')::text like '123%')) select s0.n0 as s from s0;

-- case: match (s) where s.name contains '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' like '%123%')) select s0.n0 as s from s0;

-- case: match (s) where not s.name contains '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (not coalesce(n0.properties ->> 'name', '')::text like '%123%')) select s0.n0 as s from s0;

-- case: match (s) where s.name ends with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' like '%123')) select s0.n0 as s from s0;

-- case: match (s) where not s.name ends with '123' return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (not coalesce(n0.properties ->> 'name', '')::text like '%123')) select s0.n0 as s from s0;

-- case: match (s) where s.name starts with s.other return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' like ((n0.properties ->> 'other') || '%')::text)) select s0.n0 as s from s0;

-- case: match (s) where s.name contains s.other return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' like ('%' || (n0.properties ->> 'other') || '%')::text)) select s0.n0 as s from s0;

-- case: match (s) where s.name ends with s.other return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'name' like ('%' || (n0.properties ->> 'other'))::text)) select s0.n0 as s from s0;

-- case: match (n) where n:NodeKind1 and toLower(n.tenantid) contains 'myid' and n.system_tags contains 'tag' return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and lower(n0.properties ->> 'tenantid')::text like '%myid%' and n0.properties ->> 'system_tags' like '%tag%')) select s0.n0 as n from s0;

-- case: match (s) where not (s)-[]-() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as s from s0 where (not exists (select 1 from edge e0 where e0.start_id = (s0.n0).id or e0.end_id = (s0.n0).id));

-- case: match (s) where not (s)-[]->()-[]->() return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as s from s0 where (not (with s1 as (select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0, s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0 join edge e0 on (s0.n0).id = e0.start_id join node n1 on n1.id = e0.end_id), s2 as (select s1.e0 as e0, (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1, s1.n0 as n0, s1.n1 as n1, (n2.id, n2.kind_ids, n2.properties)::nodecomposite as n2 from s1 join edge e1 on (s1.n1).id = e1.start_id join node n2 on n2.id = e1.end_id) select count(*) > 0 from s2));

-- case: match (s) where not (s)-[{prop: 'a'}]-({name: 'n3'}) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as s from s0 where (not exists (select 1 from edge e0 where e0.start_id = (s0.n0).id or e0.end_id = (s0.n0).id)) and e0.properties ->> 'prop' = 'a' and n1.properties ->> 'name' = 'n3';

-- case: match (s) where not (s)<-[{prop: 'a'}]-({name: 'n3'}) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as s from s0 where (not (with s1 as (select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0, s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0 join edge e0 on (s0.n0).id = e0.end_id join node n1 on n1.id = e0.start_id where n1.properties ->> 'name' = 'n3' and e0.properties ->> 'prop' = 'a') select count(*) > 0 from s1));

-- case: match (n:NodeKind1) where n.distinguishedname = toUpper('admin') return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'distinguishedname' = upper('admin')::text) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where n.distinguishedname starts with toUpper('admin') return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'distinguishedname' like upper('admin')::text || '%') and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where n.distinguishedname contains toUpper('admin') return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'distinguishedname' like '%' || upper('admin')::text || '%') and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where n.distinguishedname ends with toUpper('admin') return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'distinguishedname' like '%' || upper('admin')::text) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (s) where not (s)-[{prop: 'a'}]->({name: 'n3'}) return s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select s0.n0 as s from s0 where (not (with s1 as (select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0, s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0 join edge e0 on (s0.n0).id = e0.start_id join node n1 on n1.id = e0.end_id where n1.properties ->> 'name' = 'n3' and e0.properties ->> 'prop' = 'a') select count(*) > 0 from s1));

-- case: match (s) where not (s)-[]-() return id(s)
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0) select (s0.n0).id from s0 where (not exists (select 1 from edge e0 where e0.start_id = (s0.n0).id or e0.end_id = (s0.n0).id));

-- case: match (n) where n.system_tags contains ($param) return n
-- pgsql_params:{"pi0":null}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'system_tags' like '%' || (@pi0)::text || '%')) select s0.n0 as n from s0;

-- case: match (n) where n.system_tags starts with (1) return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'system_tags' like (1)::text || '%')) select s0.n0 as n from s0;

-- case: match (n) where n.system_tags ends with ('text') return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'system_tags' like '%' || ('text')::text)) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where toString(n.functionallevel) in ['2008 R2','2012','2008','2003','2003 Interim','2000 Mixed/Native'] return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'functionallevel' = any (array ['2008 R2', '2012', '2008', '2003', '2003 Interim', '2000 Mixed/Native']::text[])) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where toInt(n.value) in [1, 2, 3, 4] return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'value')::int8 = any (array [1, 2, 3, 4]::int8[])) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (u:NodeKind1) where u.pwdlastset < (datetime().epochseconds - (365 * 86400)) and not u.pwdlastset IN [-1.0, 0.0] return u limit 100
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'pwdlastset')::numeric < (extract(epoch from now()::timestamp with time zone)::numeric - (365 * 86400)) and not (n0.properties ->> 'pwdlastset')::float8 = any (array [- 1, 0]::float8[])) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as u from s0 limit 100;

-- case: match (u:NodeKind1) where u.pwdlastset < (datetime().epochmillis - 86400000) and not u.pwdlastset IN [-1.0, 0.0] return u limit 100
-- case: match (n:NodeKind1) where size(n.array_value) > 0 return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (jsonb_array_length(n0.properties -> 'array_value')::int > 0) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n) where 1 in n.array return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (1 = any (jsonb_to_text_array(n0.properties -> 'array')::int8[]))) select s0.n0 as n from s0;

-- case: match (n) where $p in n.array or $f in n.array return n
-- cypher_params: {"f":"text","p":1}
-- pgsql_params:{"pi0":1,"pi1":"text"}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (@pi0::float8 = any (jsonb_to_text_array(n0.properties -> 'array')::float8[]) or @pi1::text = any (jsonb_to_text_array(n0.properties -> 'array')::text[]))) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where coalesce(n.system_tags, '') contains 'admin_tier_0' return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (coalesce(n0.properties ->> 'system_tags', '')::text like '%admin_tier_0%') and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where coalesce(n.a, n.b, 1) = 1 return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (coalesce((n0.properties ->> 'a')::int8, (n0.properties ->> 'b')::int8, 1)::int8 = 1) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where coalesce(n.a, n.b) = 1 return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (coalesce(n0.properties ->> 'a', n0.properties ->> 'b')::int8 = 1) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where 1 = coalesce(n.a, n.b) return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (1 = coalesce(n0.properties ->> 'a', n0.properties ->> 'b')::int8) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (u:NodeKind1) where u.hasspn = true and u.enabled = true and not '-502' ends with u.objectid and not coalesce(u.gmsa, false) = true and not coalesce(u.msa, false) = true return u limit 10
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'hasspn')::bool = true and (n0.properties ->> 'enabled')::bool = true and not '-502' like ('%' || (n0.properties ->> 'objectid'))::text and not coalesce((n0.properties ->> 'gmsa')::bool, false)::bool = true and not coalesce((n0.properties ->> 'msa')::bool, false)::bool = true) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as u from s0 limit 10;

-- case: match (n:NodeKind1) where coalesce(n.name, '') = coalesce(n.migrated_name, '') return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (coalesce(n0.properties ->> 'name', '')::text = coalesce(n0.properties ->> 'migrated_name', '')::text) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where '1' in n.array_prop + ['1', '2'] return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ('1' = any (jsonb_to_text_array(n0.properties -> 'array_prop')::text[] || array ['1', '2']::text[])) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (n:NodeKind1) where ['DES-CBC-CRC', 'DES-CBC-MD5', 'RC4-HMAC-MD5'] in n.arrayProperty return n
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (array ['DES-CBC-CRC', 'DES-CBC-MD5', 'RC4-HMAC-MD5']::text[] operator (pg_catalog.&&) jsonb_to_text_array(n0.properties -> 'arrayProperty')::text[]) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as n from s0;

-- case: match (u:NodeKind1) where 'DES-CBC-CRC' in u.arrayProperty or 'DES-CBC-MD5' in u.arrayProperty or 'RC4-HMAC-MD5' in u.arrayProperty return u
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ('DES-CBC-CRC' = any (jsonb_to_text_array(n0.properties -> 'arrayProperty')::text[]) or 'DES-CBC-MD5' = any (jsonb_to_text_array(n0.properties -> 'arrayProperty')::text[]) or 'RC4-HMAC-MD5' = any (jsonb_to_text_array(n0.properties -> 'arrayProperty')::text[])) and n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]) select s0.n0 as u from s0;

-- case: match (n:NodeKind1) match (m:NodeKind2) where m.distinguishedname = 'CN=ADMINSDHOLDER,CN=SYSTEM,' + n.distinguishedname return m
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where (n1.properties ->> 'distinguishedname' = 'CN=ADMINSDHOLDER,CN=SYSTEM,' || (s0.n0).properties ->> 'distinguishedname') and n1.kind_ids operator (pg_catalog.&&) array [2]::int2[]) select s1.n1 as m from s1;

-- case: match (n:NodeKind1) match (m:NodeKind2) where m.distinguishedname = n.distinguishedname + 'CN=ADMINSDHOLDER,CN=SYSTEM,' return m
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where (n1.properties ->> 'distinguishedname' = (s0.n0).properties ->> 'distinguishedname' || 'CN=ADMINSDHOLDER,CN=SYSTEM,') and n1.kind_ids operator (pg_catalog.&&) array [2]::int2[]) select s1.n1 as m from s1;

-- case: match (n:NodeKind1) match (m:NodeKind2) where m.distinguishedname = n.unknown + m.unknown return m
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where (n1.properties ->> 'distinguishedname' = (s0.n0).properties -> 'unknown' + n1.properties -> 'unknown') and n1.kind_ids operator (pg_catalog.&&) array [2]::int2[]) select s1.n1 as m from s1;

-- case: match (n:NodeKind1) match (m:NodeKind2) where m.distinguishedname = '1' + '2' return m
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]), s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1 from s0, node n1 where (n1.properties ->> 'distinguishedname' = '1' || '2') and n1.kind_ids operator (pg_catalog.&&) array [2]::int2[]) select s1.n1 as m from s1;

