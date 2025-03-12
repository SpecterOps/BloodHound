-- Copyright 2023 Specter Ops, Inc.
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

-- DAWGS Property Graph Partitioned Layout for PostgreSQL

-- Notes on TOAST:
--
-- Graph entity properties are stored in a JSONB column at the end of the row. There is a soft-limit of 2KiB for rows in
-- a PostgreSQL database page. The database will compress this value in an attempt not to exceed this limit. Once a
-- compressed value reaches the absolute limit of what the database can do to either compact it or give it more of the
-- 8 KiB page size limit, the database evicts the value to an associated TOAST (The Oversized-Attribute Storage Technique)
-- table and creates a reference to the entry to be joined upon fetch of the row.
--
-- TOAST comes with certain performance caveats that can affect access time anywhere from a factor 3 to 6 times. It is
-- in the best interest of the database user that the properties of a graph entity never exceed this limit in large
-- graphs.

-- We need the tri-gram extension to create a GIN text-search index. The goal here isn't full-text search, in which
-- case ts_vector and its ilk would be more suited. This particular selection was made to support accelerated lookups
-- for "contains", "starts with" and, "ends with" comparison operations.
create extension if not exists pg_trgm;

-- We need the intarray extension for extended integer array operations like unions. This is useful for managing kind
-- arrays for nodes.
create extension if not exists intarray;

-- This is an optional but useful extension for validating performance of queries
-- create extension if not exists pg_stat_statements;
--
-- create or replace function public.query_perf()
--   returns table
--           (
--             query              text,
--             calls              int,
--             total_time         numeric,
--             mean_time          numeric,
--             percent_total_time numeric
--           )
-- as
-- $$
-- select query                                                                      as query,
--        calls                                                                      as calls,
--        round(total_exec_time::numeric, 2)                                         as total_time,
--        round(mean_exec_time::numeric, 2)                                          as mean_time,
--        round((100 * total_exec_time / sum(total_exec_time) over ()):: numeric, 2) as percent_total_time
-- from pg_stat_statements
-- order by total_exec_time desc
-- limit 25
-- $$
--   language sql
--   immutable
--   parallel safe
--   strict;

-- Table definitions

-- The graph table contains name to ID mappings for graphs contained within the database. Each graph ID should have
-- corresponding table partitions for the node and edge tables.
create table if not exists graph
(
  id   bigserial,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

-- The kind table contains name to ID mappings for graph kinds. Storage of these types is necessary to maintain search
-- capability of a database without the origin application that generated it. 
-- To aid in testing tiering/labels, the kind table is duplicated in the stepwise migration files. Any schema upates here
-- should be reflected in a stepwise migration file as well. 
create table if not exists kind
(
  id   smallserial,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

-- Node composite type
do
$$
  begin
    create type nodeComposite as
    (
      id         bigint,
      kind_ids   smallint[8],
      properties jsonb
    );
  exception
    when duplicate_object then null;
  end
$$;

-- The node table is a partitioned table view that partitions over the graph ID that each node belongs to. Nodes may
-- contain a disjunction of up to 8 kinds for creating clique subsets without requiring edges.
create table if not exists node
(
  id         bigserial   not null,
  graph_id   integer     not null,
  kind_ids   smallint[8] not null,
  properties jsonb       not null,

  primary key (id, graph_id),
  foreign key (graph_id) references graph (id) on delete cascade
) partition by list (graph_id);

-- The storage strategy chosen for the properties JSONB column informs the database of the user's preference to resort
-- to creating a TOAST table entry only after there is no other possible way to inline the row attribute in the current
-- page.
alter table node
  alter column properties set storage main;

-- Index on the graph ID of each node.
create index if not exists node_graph_id_index on node using btree (graph_id);

-- Index node kind IDs so that lookups by kind is accelerated.
create index if not exists node_kind_ids_index on node using gin (kind_ids);

-- Edge composite type
do
$$
  begin
    create type edgeComposite as
    (
      id         bigint,
      start_id   bigint,
      end_id     bigint,
      kind_id    smallint,
      properties jsonb
    );
  exception
    when duplicate_object then null;
  end
$$;

-- The edge table is a partitioned table view that partitions over the graph ID that each edge belongs to.
create table if not exists edge
(
  id         bigserial not null,
  graph_id   integer   not null,
  start_id   bigint    not null,
  end_id     bigint    not null,
  kind_id    smallint  not null,
  properties jsonb     not null,

  primary key (id, graph_id),
  foreign key (graph_id) references graph (id) on delete cascade,

  unique (graph_id, start_id, end_id, kind_id)
) partition by list (graph_id);

-- delete_node_edges is a trigger and associated plpgsql function to cascade delete edges when attached nodes are
-- deleted. While this could be done with a foreign key relationship, it would scope the cascade delete to individual
-- node partitions and therefore require the graph_id value of each node as part of the delete statement.
create or replace function delete_node_edges() returns trigger as
$$
begin
  delete from edge where start_id = OLD.id or end_id = OLD.id;
  return null;
end
$$
  language plpgsql
  volatile
  strict;

-- Drop and create the delete_node_edges trigger for the delete_node_edges() plpgsql function. See the function comment
-- for more information.
drop trigger if exists delete_node_edges on node;
create trigger delete_node_edges
  after delete
  on node
  for each row
execute procedure delete_node_edges();


-- The storage strategy chosen for the properties JSONB column informs the database of the user's preference to resort
-- to creating a TOAST table entry only after there is no other possible way to inline the row attribute in the current
-- page.
alter table edge
  alter column properties set storage main;

-- Index on the graph ID of each edge.
create index if not exists edge_graph_id_index on edge using btree (graph_id);

-- Index on the start vertex of each edge.
create index if not exists edge_start_id_index on edge using btree (start_id);

-- Index on the start vertex of each edge.
create index if not exists edge_end_id_index on edge using btree (end_id);

-- Index on the kind of each edge.
create index if not exists edge_kind_index on edge using btree (kind_id);

-- Path composite type
do
$$
  begin
    create type pathComposite as
    (
      nodes nodeComposite[],
      edges edgeComposite[]
    );
  exception
    when duplicate_object then null;
  end
$$;

-- Database helper functions
create or replace function public.lock_details()
  returns table
          (
            datname      text,
            locktype     text,
            relation     text,
            lock_mode    text,
            txid         xid,
            virtual_txid text,
            pid          integer,
            tx_granted   bool,
            client_addr  text,
            client_port  integer,
            elapsed_time interval
          )
as
$$
select db.datname              as datname,
       locktype                as locktype,
       relation::regclass      as relation,
       mode                    as lock_mode,
       transactionid           as txid,
       virtualtransaction      as virtual_txid,
       l.pid                   as pid,
       granted                 as tx_granted,
       psa.client_addr         as client_addr,
       psa.client_port         as client_port,
       now() - psa.query_start as elapsed_time
from pg_catalog.pg_locks l
       left join pg_catalog.pg_database db on db.oid = l.database
       left join pg_catalog.pg_stat_activity psa on l.pid = psa.pid
where not l.pid = pg_backend_pid();
$$
  language sql
  immutable
  parallel safe
  strict;

create or replace function public.table_sizes()
  returns table
          (
            oid          int,
            table_schema text,
            table_name   text,
            total_bytes  numeric,
            total_size   text,
            index_size   text,
            toast_size   text,
            table_size   text
          )
as
$$
select oid                         as oid,
       table_schema                as table_schema,
       table_name                  as table_name,
       total_bytes                 as total_bytes,
       pg_size_pretty(total_bytes) as total_size,
       pg_size_pretty(index_bytes) as index_size,
       pg_size_pretty(toast_bytes) as toast_size,
       pg_size_pretty(table_bytes) as table_size
from (select *, total_bytes - index_bytes - coalesce(toast_bytes, 0) as table_bytes
      from (select c.oid                                 as oid,
                   nspname                               as table_schema,
                   relname                               as table_name,
                   c.reltuples                           as row_estimate,
                   pg_total_relation_size(c.oid)         as total_bytes,
                   pg_indexes_size(c.oid)                as index_bytes,
                   pg_total_relation_size(reltoastrelid) as toast_bytes
            from pg_class c
                   left join pg_namespace n on n.oid = c.relnamespace
            where relkind = 'r') a) a
order by total_bytes desc;
$$
  language sql
  immutable
  parallel safe
  strict;

create or replace function public.index_utilization()
  returns table
          (
            table_name    text,
            idx_scans     int,
            seq_scans     int,
            index_usage   int,
            rows_in_table int
          )
as
$$
select relname                                table_name,
       idx_scan                               index_scan,
       seq_scan                               table_scan,
       100 * idx_scan / (seq_scan + idx_scan) index_usage,
       n_live_tup                             rows_in_table
from pg_stat_user_tables
where seq_scan + idx_scan > 0
order by index_usage desc
limit 25;
$$
  language sql
  immutable
  parallel safe
  strict;

create or replace function public.jsonb_to_text_array(target jsonb)
  returns text[]
as
$$
begin
  if target != 'null'::jsonb then
    return array(select jsonb_array_elements_text(target));
  else
    return array[]::text[];
  end if;
end
$$
  language plpgsql
  immutable
  parallel safe
  strict;

-- All shortest path traversal harness.
create or replace function public.asp_harness(primer_query text, recursive_query text, max_depth int4)
  -- | Column      | type    | Usage                                                                                  |
  -- |-------------|---------|----------------------------------------------------------------------------------------|
  -- | `root_id`   | Int8    | Node that the path originated from. Simplifies referencing the root node of each path. |
  -- | `next_id`   | Int8    | Next node to expand to.                                                                |
  -- | `depth`     | Int4    | Depth of the current traversal.                                                        |
  -- | `satisfied` | Boolean | True if the expansion is satisfied.                                                    |
  -- | `is_cycle`  | Boolean | True if the expansion is a cycle.                                                      |
  -- | `path`      | Int8[]  | Array of edges in order of traversal.                                                  |
  returns table
          (
            root_id   int8,
            next_id   int8,
            depth     int4,
            satisfied bool,
            is_cycle  bool,
            path      int8[]
          )
as
$$
declare
  depth int4 := 1;
begin
  -- Define two tables to represent pathspace of the recursive expansion. These are temporary and as such are unlogged.
  create temporary table pathspace
  (
    root_id   int8   not null,
    next_id   int8   not null,
    depth     int4   not null,
    satisfied bool   not null,
    is_cycle  bool   not null,
    path      int8[] not null,
    primary key (path)
  ) on commit drop;

  create temporary table next_pathspace
  (
    root_id   int8   not null,
    next_id   int8   not null,
    depth     int4   not null,
    satisfied bool   not null,
    is_cycle  bool   not null,
    path      int8[] not null,
    primary key (path)
  ) on commit drop;

  -- Creating these indexes should speed up certain operations during recursive expansion. Benchmarking should be done
  -- to validate assumptions here.
  create index pathspace_next_id_index on pathspace using btree (next_id);
  create index pathspace_satisfied_index on pathspace using btree (satisfied);
  create index pathspace_is_cycle_index on pathspace using btree (is_cycle);

  create index next_pathspace_next_id_index on next_pathspace using btree (next_id);
  create index next_pathspace_satisfied_index on next_pathspace using btree (satisfied);
  create index next_pathspace_is_cycle_index on next_pathspace using btree (is_cycle);

  raise notice 'Expansion start';

  -- Populate initial pathspace with the primer query - this is the depth 1 traversal expansion
  execute primer_query;

  raise notice 'Expansion step % - Available Paths % - Num satisfied: %', depth, (select count(*) from next_pathspace), (select count(*) from next_pathspace p where p.satisfied);

  -- Iterate until one of the following conditions:
  -- * Traversal has reached the max allowed depth
  -- * Pathspace is exhausted (no further expansion possible)
  -- * A terminal node (and therefore one of the shortest paths) has been found
  while depth < max_depth and
        exists(select 1 from next_pathspace) and
        not exists(select 1 from next_pathspace np where np.satisfied)
    loop
      -- Rename tables to swap in the next pathspace as the current pathspace for the next traversal step
      alter table pathspace
        rename to pathspace_old;
      alter table next_pathspace
        rename to pathspace;
      alter table pathspace_old
        rename to next_pathspace;

      -- Clear the next pathspace scratch
      truncate table next_pathspace;

      -- Remove any non-satisfied terminals and cycles from pathspace
      raise notice 'Available Paths Before Delete % - Num satisfied: %', (select count(*) from pathspace), (select count(*) from pathspace p where p.satisfied);

      delete
      from pathspace p
      where p.is_cycle
         or not exists(select 1 from edge e where e.end_id = p.next_id);

      raise notice 'Available Paths After Delete % - Num satisfied: %', (select count(*) from pathspace), (select count(*) from pathspace p where p.satisfied);

      -- Increase the current depth and execute the recursive query
      depth := depth + 1;
      execute recursive_query;

      raise notice 'Expansion step % - Available Paths % - Num satisfied: %', depth, (select count(*) from next_pathspace), (select count(*) from next_pathspace p where p.satisfied);
    end loop;

  -- Return all satisfied paths from the next pathspace table
  return query select *
               from next_pathspace np
               where np.satisfied;

  -- Close the result set
  return;
end;
$$
  language plpgsql volatile
                   strict;

create or replace function public.edges_to_path(path variadic int8[]) returns pathComposite as
$$
select row (array_agg(distinct (n.id, n.kind_ids, n.properties)::nodeComposite)::nodeComposite[],
         array_agg(distinct (r.id, r.start_id, r.end_id, r.kind_id, r.properties)::edgeComposite)::edgeComposite[])::pathComposite
from edge r
       join node n on n.id = r.start_id or n.id = r.end_id
where r.id = any (path);
$$
  language sql
  immutable
  parallel safe
  strict;
