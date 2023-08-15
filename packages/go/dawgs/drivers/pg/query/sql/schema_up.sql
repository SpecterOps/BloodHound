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
  id   serial,
  name varchar(256) not null,
  primary key (id),
  unique (name)
);

-- The kind table contains name to ID mappings for graph kinds. Storage of these types is necessary to maintain search
-- capability of a database without the origin application that generated it.
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
      id         integer,
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
  id         serial      not null,
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
      id         integer,
      start_id   integer,
      end_id     integer,
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
  id         serial   not null,
  graph_id   integer  not null,
  start_id   integer  not null,
  end_id     integer  not null,
  kind_id    smallint not null,
  properties jsonb    not null,

  primary key (id, graph_id),
  foreign key (graph_id) references graph (id) on delete cascade
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
  language plpgsql;

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

-- Graph helper functions
create or replace function public.kinds(target anyelement) returns text[] as
$$
begin
  if pg_typeof(target) = 'node'::regtype then
    return (select array_agg(k.name) from kind k where k.id = any (target.kind_ids));
  elsif pg_typeof(target) = 'edge'::regtype then
    return (select array_agg(k.name) from kind k where k.id = target.kind_id);
  elsif pg_typeof(target) = 'int[]'::regtype then
    return (select array_agg(k.name) from kind k where k.id = any (target::int2[]));
  elsif pg_typeof(target) = 'int'::regtype then
    return (select array_agg(k.name) from kind k where k.id = target::int2);
  elsif pg_typeof(target) = 'int2[]'::regtype then
    return (select array_agg(k.name) from kind k where k.id = any (target));
  elsif pg_typeof(target) = 'int2'::regtype then
    return (select array_agg(k.name) from kind k where k.id = target);
  else
    raise exception 'Invalid argument type: %', pg_typeof(target) using hint = 'Type must be either node, edge, int[], int, int2[] or int2';
  end if;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create or replace function public.has_kind(target anyelement, variadic kind_name_in text[]) returns bool as
$$
begin
  if pg_typeof(target) = 'node'::regtype then
    return exists(select 1
                  where target.kind_ids operator (pg_catalog.&&)
                        (select array_agg(id) from kind k where k.name = any (kind_name_in)));
  elsif pg_typeof(target) = 'edge'::regtype then
    return exists(select 1
                  where target.kind_id in (select id from kind k where k.name = any (kind_name_in)));
  else
    raise exception 'Invalid argument type: %', pg_typeof(target) using hint = 'Type must be either node or edge';
  end if;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create
  or replace function public.get_node(id_in int4)
  returns setof node as
$$
select *
from node n
where n.id = id_in;
$$
  language sql immutable
               parallel safe
               strict;

create
  or replace function public.node_prop(target anyelement, property_name text)
  returns jsonb as
$$
begin
  if pg_typeof(target) = 'node'::regtype then
    return target.properties -> property_name;
  elsif pg_typeof(target) = 'int4'::regtype then
    return (select n.properties -> property_name from node n where n.id = target limit 1);
  else
    raise exception 'Invalid argument type: %', pg_typeof(target) using hint = 'Type must be either node or edge';
  end if;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;


create or replace function public.mt_get_root(owner_object_id text) returns setof node as
$$
select *
from node n
where has_kind(n, 'Meta')
  and n.properties ->> 'system_tags' like '%admin_tier_0%'
  and n.properties ->> 'owner_objectid' = owner_object_id;
$$
  language sql immutable
               parallel safe
               strict;

-- All shortest path traversal functions and schema

create or replace function public._format_asp_where_clause(root_criteria text, where_clause text) returns text as
$$
declare
  formatted_query text := '';
begin
  if length(root_criteria) > 0 then
    if length(where_clause) > 0 then
      formatted_query := ' where ' || root_criteria || ' and ' || where_clause;
    else
      formatted_query := ' where ' || root_criteria;
    end if;
  elsif length(where_clause) > 0 then
    formatted_query := ' where ' || where_clause;
  end if;

  return formatted_query;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create or replace function public._format_asp_query(terminal_criteria text, cycle_criteria text,
                                                    traversal_criteria text default '',
                                                    root_criteria text default '',
                                                    bind_pathspace bool default true,
                                                    bind_start bool default false,
                                                    bind_end bool default false) returns text as
$$
declare
  formatted_query text := 'insert into pathspace_next (path, next, is_terminal, is_cycle) ';
begin
  if bind_pathspace then
    formatted_query :=
      formatted_query || 'select p.path || r.id, r.end_id, ' || terminal_criteria || ', ' || cycle_criteria ||
      ' from edge r join pathspace_current p on p.next = r.start_id';
  else
    formatted_query := formatted_query || 'select array [r.id]::int4[], r.end_id, ' || terminal_criteria || ', ' ||
                       cycle_criteria ||
                       ' from edge r';
  end if;

  if bind_start then
    formatted_query := formatted_query || ' join node s on s.id = r.start_id';
  end if;

  if bind_end then
    formatted_query := formatted_query || ' join node e on e.id = r.end_id ';
  end if;

  formatted_query := formatted_query || _format_asp_where_clause(root_criteria, traversal_criteria) || ';';

  raise notice '_format_asp_query -> %', formatted_query;
  return formatted_query;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create or replace function public._all_shortest_paths(root_criteria text,
                                                      traversal_criteria text,
                                                      terminal_criteria text,
                                                      max_depth int4)
  returns table
          (
            path int4[]
          )
as
$$
declare
  has_root_criteria         bool := length(root_criteria) > 0;
  has_traversal_criteria    bool := length(traversal_criteria) > 0;
  has_terminal_criteria     bool := length(terminal_criteria) > 0;

  -- Make sure to take into account if queries will need the start or end node of edges bound by a join
  bind_root_node            bool := has_root_criteria and root_criteria like '%s.%';
  bind_terminal_node        bool := has_terminal_criteria and terminal_criteria like '%e.%';
  bind_traversal_start_node bool := has_traversal_criteria and traversal_criteria like '%s.%';
  bind_traversal_end_node   bool := has_traversal_criteria and traversal_criteria like '%e.%';
  depth                     int4 := 1;
begin
  -- Create two unlogged (no WAL writes) temporary tables (invisible to other sessions) for storing traversal
  -- fronts during path expansion.
  create temporary table pathspace_current
  (
    path        int4[] not null,
    next        int4   not null,
    is_terminal bool   not null,
    is_cycle    bool   not null,
    primary key (path)
  ) on commit drop;

  create temporary table pathspace_next
  (
    path        int4[] not null,
    next        int4   not null,
    is_terminal bool   not null,
    is_cycle    bool   not null,
    primary key (path)
  ) on commit drop;

  -- Create an index on the next node ID to accelerate joins
  create index if not exists pathspace_current_next_index on pathspace_current using btree (next);
  create index if not exists pathspace_next_next_index on pathspace_next using btree (next);

  -- Create an index on the is_terminal boolean to accelerate aggregation and selection
  create index if not exists pathspace_current_terminal_index on pathspace_current using btree (is_terminal);
  create index if not exists pathspace_next_terminal_index on pathspace_next using btree (is_terminal);

  -- Initial expansion to acquire the first traversal front
  execute _format_asp_query(terminal_criteria := terminal_criteria,
                            cycle_criteria := 'r.start_id = r.end_id',
                            bind_pathspace := false,
                            bind_start := bind_traversal_start_node or bind_root_node,
                            bind_end := bind_traversal_end_node or bind_terminal_node,
                            root_criteria := root_criteria,
                            traversal_criteria := traversal_criteria);

  -- Copy from the next pathspace table to the current pathspace table. Any non-terminal cycles are omitted as
  -- part of this copy to prune visited branches.
  insert into pathspace_current select * from pathspace_next p where not p.is_cycle or p.is_terminal;

  -- Truncate the next pathspace table to clear it
  truncate pathspace_next;

  -- Loop until either the current depth exceeds the max allowed depth or if any of the paths are terminal
  while depth < max_depth and (select count(*) from pathspace_current p where p.is_terminal) = 0
    loop
      -- Increase the depth counter as we're expanding a new front
      depth := depth + 1;

      -- Perform the next pathspace expansion
      execute _format_asp_query(terminal_criteria := terminal_criteria,
                                cycle_criteria := 'r.id = any (p.path)',
                                bind_pathspace := true,
                                bind_start := bind_traversal_start_node,
                                bind_end := bind_terminal_node or bind_traversal_end_node,
                                traversal_criteria := traversal_criteria);

      -- Truncate the old pathspace table to clear it
      truncate pathspace_current;

      -- Copy from the next pathspace table to the current pathspace table. Any non-terminal cycles are omitted as
      -- part of this copy to prune visited branches.
      insert into pathspace_current select * from pathspace_next p where not p.is_cycle or p.is_terminal;

      -- Truncate the next pathspace table to clear it
      truncate pathspace_next;
    end loop;

  -- Return the raw path (set of edge IDs) for each path found in pathspace
  return query select p.path
               from pathspace_current p
               -- Select only terminal paths
               where p.is_terminal;

  -- Close the set
  return;
end;
$$
  language plpgsql volatile
                   strict;

create or replace function public.edges_to_path(path variadic int4[]) returns pathComposite as
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

create or replace function public.all_shortest_paths(root_criteria text,
                                                     traversal_criteria text,
                                                     terminal_criteria text,
                                                     max_depth int4)
  returns pathComposite
as
$$
declare
  paths pathcomposite;
begin
  select array_agg(distinct (n.id, n.kind_ids, n.properties)::nodeComposite)::nodeComposite[],
         array_agg(distinct
                   (r.id, r.start_id, r.end_id, r.kind_id, r.properties)::edgeComposite)::edgeComposite[]
  into paths
  from _all_shortest_paths(root_criteria, traversal_criteria, terminal_criteria,
                           max_depth) as t
         join edge r on r.id = any (t.path)
         join node n on n.id = r.start_id or n.id = r.end_id;

  return paths;
end;
$$
  language plpgsql
  immutable
  strict;

-- Generic traversal functions and schema
do
$$
  begin
    create type _traversal_step as
    (
      root_criteria      text,
      traversal_criteria text,
      terminal_criteria  text,
      max_depth          integer
    );
  exception
    when duplicate_object then null;
  end
$$;

create or replace function public.traversal_step(root_criteria text default '',
                                                 traversal_criteria text default '',
                                                 terminal_criteria text default '',
                                                 max_depth integer default 0)
  returns _traversal_step as
$$
begin
  return (root_criteria, traversal_criteria, terminal_criteria, max_depth)::_traversal_step;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create or replace function public._format_traversal_continuation_termination(terminal_criteria text,
                                                                             bind_traversal_start_node bool,
                                                                             bind_traversal_end_node bool)
  returns text as
$$
declare
  formatted_query text := 'update pathspace_current p set terminal = true from edge r';
  where_clause    text := ' where not p.terminal and p.exhausted and r.start_id = p.path[array_length(p.path, 1)]';
begin
  if bind_traversal_start_node then
    formatted_query := formatted_query || ', node s';
    where_clause := where_clause || ' and s.id = r.start_id';
  end if;

  if bind_traversal_end_node then
    formatted_query := formatted_query || ', node e';
    where_clause := where_clause || ' and e.id = r.end_id';
  end if;

  return formatted_query || where_clause || ' and ' || terminal_criteria;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create or replace function public._format_traversal_query(traversal_criteria text,
                                                          terminal_criteria text,
                                                          mark_terminal bool,
                                                          bind_traversal_start_node bool,
                                                          bind_traversal_end_node bool)
  returns text as
$$
declare
  formatted_query text := 'with inserts as (insert into pathspace_next (path, next, terminal, exhausted, rejected) ';
begin
  formatted_query := formatted_query || 'select $1 || r.id, r.end_id, ';

  if length(terminal_criteria) > 0 then
    formatted_query := formatted_query || terminal_criteria || ', ';
  else
    formatted_query := formatted_query || mark_terminal || ', ';
  end if;

  formatted_query :=
    formatted_query || 'false, r.id = any ($1) from edge r ';

  if bind_traversal_start_node then
    formatted_query := formatted_query || ' join node s on s.id = r.start_id';
  end if;

  if bind_traversal_end_node then
    formatted_query := formatted_query || ' join node e on e.id = r.end_id ';
  end if;

  formatted_query := formatted_query || ' where r.start_id = $2';

  if length(traversal_criteria) > 0 then
    formatted_query := formatted_query || ' and ' || traversal_criteria;
  end if;

  return formatted_query || ' returning true) select count(*) from inserts;';
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create or replace function public._format_traversal_initial_query(root_criteria text,
                                                                  terminal_criteria text,
                                                                  mark_terminal bool,
                                                                  traversal_criteria text,
                                                                  bind_root_node bool,
                                                                  bind_terminal_node bool) returns text as
$$
declare
  formatted_query text := 'insert into pathspace_current (path, next, terminal, exhausted, rejected) ';
begin
  formatted_query := formatted_query || 'select array [r.id]::int4[], r.end_id, ';

  if length(terminal_criteria) > 0 then
    formatted_query := formatted_query || terminal_criteria || ', ';
  else
    formatted_query := formatted_query || mark_terminal || ', ';
  end if;

  formatted_query := formatted_query || 'false, r.start_id = r.end_id from edge r';

  if bind_root_node then
    formatted_query := formatted_query || ' join node s on s.id = r.start_id';
  end if;

  if bind_terminal_node then
    formatted_query := formatted_query || ' join node e on e.id = r.end_id ';
  end if;

  if length(root_criteria) > 0 then
    if length(traversal_criteria) > 0 then
      formatted_query := formatted_query || ' where ' || root_criteria || ' and ' || traversal_criteria;
    else
      formatted_query := formatted_query || ' where ' || root_criteria;
    end if;
  elsif length(traversal_criteria) > 0 then
    formatted_query := formatted_query || ' where ' || traversal_criteria;
  end if;

  return formatted_query;
end;
$$
  language plpgsql immutable
                   parallel safe
                   strict;

create or replace function public.expand_traversal_step(step _traversal_step,
                                                        continuation bool default false,
                                                        last_continuation bool default false)
  returns void as
$$
declare
  incomplete_path           record;
  num_expansions            int8;
  has_root_criteria         bool := length(step.root_criteria) > 0;
  has_traversal_criteria    bool := length(step.traversal_criteria) > 0;
  has_terminal_criteria     bool := length(step.terminal_criteria) > 0;

  -- Make sure to take into account if queries will need the start or end node of edges bound by a join
  bind_root_node            bool := has_root_criteria and step.root_criteria like '%s.%';
  bind_terminal_node        bool := has_terminal_criteria and step.terminal_criteria like '%e.%';
  bind_traversal_start_node bool := has_traversal_criteria and step.traversal_criteria like '%s.%';
  bind_traversal_end_node   bool := has_traversal_criteria and step.traversal_criteria like '%e.%';
  depth                     int4 := 0;
begin
  if not continuation then
    raise notice 'Starting a new traversal';

    -- Increase the depth counter as we're expanding a new front
    depth := depth + 1;

    -- Perform the initial expansion to acquire the first traversal front
    execute _format_traversal_initial_query(
      root_criteria := step.root_criteria,
      bind_root_node := bind_root_node or bind_traversal_start_node,
      terminal_criteria := step.terminal_criteria,
      mark_terminal := last_continuation and not has_terminal_criteria and depth = step.max_depth,
      bind_terminal_node := bind_terminal_node or bind_traversal_end_node,
      traversal_criteria := step.traversal_criteria);

    -- Copy from the next pathspace table to the current pathspace table and omit rejected segments.
    insert into pathspace_current select * from pathspace_next p where not p.rejected or p.terminal;

    -- Truncate the next pathspace table to clear it
    truncate pathspace_next;
  else
    raise notice 'Continuing traversal';

    if last_continuation then
      raise notice 'This is the last continuation.';
    end if;

    if has_terminal_criteria then
      -- If this is a traversal continuation then we must validate any exhausted paths that may also be terminal
      execute _format_traversal_continuation_termination(
        terminal_criteria := step.terminal_criteria,
        bind_traversal_start_node := bind_traversal_start_node,
        bind_traversal_end_node := bind_terminal_node or bind_traversal_end_node);

      -- Dump any paths that are not terminal but exhausted as this will prune pathspace to only paths that are
      -- eligible for further expansion
      delete from pathspace_current p where not p.terminal and p.exhausted;
    else
      -- Mark all non-terminal paths as no longer exhausted as this is a continuation
      update pathspace_current p set exhausted = false where not p.terminal;
    end if;
  end if;

  raise notice 'Current pathspace:';

  for incomplete_path in select * from pathspace_current p
    loop
      raise notice 'path: % - terminal: %, exhausted: %, rejected: %', incomplete_path.path, incomplete_path.terminal, incomplete_path.exhausted, incomplete_path.rejected;
    end loop;

  -- Loop until either the current depth exceeds the max allowed depth or if any of the paths are terminal
  while depth < step.max_depth and
        exists(select true from pathspace_current p where not p.terminal and not p.exhausted)
    loop
      raise notice 'Incomplete paths:';

      for incomplete_path in select * from pathspace_current p where not p.terminal and not p.exhausted
        loop
          raise notice 'path: % - terminal: %, exhausted: %, rejected: %', incomplete_path.path, incomplete_path.terminal, incomplete_path.exhausted, incomplete_path.rejected;
        end loop;

      -- Increase the depth counter as we're expanding a new front
      depth := depth + 1;

      -- Copy all terminal segments
      insert into pathspace_next select * from pathspace_current p where p.terminal;

      -- Expand all non-terminal, unexhausted segments
      for incomplete_path in select * from pathspace_current p where not p.terminal and not p.exhausted
        loop
          -- Expand the next front for this segment
          execute _format_traversal_query(
            traversal_criteria := step.traversal_criteria,
            terminal_criteria := step.terminal_criteria,
            mark_terminal := last_continuation and not has_terminal_criteria and depth = step.max_depth,
            bind_traversal_start_node := bind_traversal_start_node,
            bind_traversal_end_node := bind_terminal_node or bind_traversal_end_node)
            into num_expansions
            using incomplete_path.path, incomplete_path.next;

          if num_expansions = 0 then
            -- If there were no more expansions for this segment, insert into the next pathspace it as
            -- exhausted. The terminal status of the segment may be set to true if this is the last
            -- traversal continuation and there is no terminal criteria set.
            insert into pathspace_next (path, next, terminal, exhausted, rejected)
            values (incomplete_path.path, 0, last_continuation and not has_terminal_criteria, true,
                    false);
          end if;
        end loop;

      -- Truncate the old pathspace table to clear it
      truncate pathspace_current;

      -- Copy from the next pathspace into the current pathspace
      insert into pathspace_current select * from pathspace_next p where not p.rejected;

      -- Truncate the next pathspace table to clear it
      truncate pathspace_next;
    end loop;

  raise notice 'Step pathspace:';
  for incomplete_path in select * from pathspace_current p
    loop
      raise notice 'path: % - terminal: %, exhausted: %, rejected: %', incomplete_path.path, incomplete_path.terminal, incomplete_path.exhausted, incomplete_path.rejected;
    end loop;

  return;
end;
$$
  language plpgsql volatile
                   strict;

create or replace function public.traverse(steps variadic _traversal_step[])
  returns table
          (
            path int4[][]
          )
as
$$
declare
  step_idx  int4 = 0;
  next_step _traversal_step;
begin
  -- Create two unlogged (no WAL writes) temporary tables (invisible to other sessions) for storing traversal
  -- fronts during path expansion.
  create temporary table pathspace_current
  (
    path      int4[] not null,
    next      int4   not null,
    terminal  bool   not null,
    exhausted bool   not null,
    rejected  bool   not null,
    primary key (path)
  ) on commit drop;

  create temporary table pathspace_next
  (
    path      int4[] not null,
    next      int4   not null,
    terminal  bool   not null,
    exhausted bool   not null,
    rejected  bool   not null,
    primary key (path)
  ) on commit drop;

  -- Iterate through the traversal steps
  foreach next_step in array steps
    loop
      step_idx := step_idx + 1;

      raise notice 'Array length: % - Is last continuation: %', array_length(steps, 1), step_idx = array_length(steps, 1);

      perform expand_traversal_step(
        step := next_step,
        continuation := step_idx > 1,
        last_continuation := step_idx = array_length(steps, 1));
    end loop;

  -- Return the paths
  return query select p.path from pathspace_current p where p.terminal;

  -- Close the set
  return;
end;
$$
  language plpgsql volatile
                   strict;
