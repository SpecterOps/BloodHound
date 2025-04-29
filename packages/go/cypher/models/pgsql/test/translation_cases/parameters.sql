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

-- case: match (n) where n.objectid ends with $p0 and not (n:NodeKind1 or n:NodeKind2) return n
-- pgsql_params:{"pi0":null}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'objectid' like '%' || @pi0 and not (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] or n0.kind_ids operator (pg_catalog.&&) array [2]::int2[]))) select s0.n0 as n from s0;

-- case: match (n) where n.objectid starts with $p0 and not (n:NodeKind1 or n:NodeKind2) return n
-- pgsql_params:{"pi0":null}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'objectid' like @pi0 || '%' and not (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] or n0.kind_ids operator (pg_catalog.&&) array [2]::int2[]))) select s0.n0 as n from s0;

-- case: match (n) where n.objectid contains $p0 and not (n:NodeKind1 or n:NodeKind2) return n
-- pgsql_params:{"pi0":null}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where (n0.properties ->> 'objectid' like '%' || @pi0 || '%' and not (n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] or n0.kind_ids operator (pg_catalog.&&) array [2]::int2[]))) select s0.n0 as n from s0;

