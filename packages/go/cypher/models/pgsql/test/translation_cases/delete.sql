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

-- case: match (s:NodeKind1) detach delete s
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[]),
     s1 as (delete from node n1 using s0 where (s0.n0).id = n1.id)
select 1;

-- case: match ()-[r:EdgeKind1]->() delete r
with s0 as (select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.id = e0.start_id
                   join node n1 on n1.id = e0.end_id
            where e0.kind_id = any (array [3]::int2[])),
     s1 as (delete from edge e1 using s0 where (s0.e0).id = e1.id)
select 1;

-- case: match ()-[]->()-[r:EdgeKind1]->() delete r
with s0 as (select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                   join node n0 on n0.id = e0.start_id
                   join node n1 on n1.id = e0.end_id),
     s1 as (select s0.e0                                                                     as e0,
                   (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
                   s0.n0                                                                     as n0,
                   s0.n1                                                                     as n1,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
            from s0,
                 edge e1
                   join node n2 on n2.id = e1.end_id
            where e1.kind_id = any (array [3]::int2[])
              and (s0.n1).id = e1.start_id),
     s2 as (delete from edge e2 using s1 where (s1.e1).id = e2.id)
select 1;
