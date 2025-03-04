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

drop function if exists create_pathspace_tables();
drop function if exists swap_root_front();
drop function if exists swap_terminal_front();
drop function if exists bidirectional_asp_harness(root_primer text, root_recursive text, terminal_primer text,
                                                  terminal_recursive text, max_depth int4);

--
create or replace function public.create_pathspace_tables()
  returns void as
$$
begin
  create temporary table root_front
  (
    root_id   int8   not null,
    next_id   int8   not null,
    depth     int4   not null,
    satisfied bool   not null,
    is_cycle  bool   not null,
    path      int8[] not null,
    primary key (path)
  ) on commit drop;

  create temporary table terminal_front
  (
    root_id   int8   not null,
    next_id   int8   not null,
    depth     int4   not null,
    satisfied bool   not null,
    is_cycle  bool   not null,
    path      int8[] not null,
    primary key (path)
  ) on commit drop;

  create temporary table next_front
  (
    root_id   int8   not null,
    next_id   int8   not null,
    depth     int4   not null,
    satisfied bool   not null,
    is_cycle  bool   not null,
    path      int8[] not null,
    primary key (path)
  ) on commit drop;

  create index root_front_next_id_index on root_front using btree (next_id);
  create index root_front_satisfied_index on root_front using btree (satisfied);
  create index root_front_is_cycle_index on root_front using btree (is_cycle);

  create index terminal_front_next_id_index on terminal_front using btree (next_id);
  create index terminal_front_satisfied_index on terminal_front using btree (satisfied);
  create index terminal_front_is_cycle_index on terminal_front using btree (is_cycle);

  create index next_front_next_id_index on next_front using btree (next_id);
  create index next_front_satisfied_index on next_front using btree (satisfied);
  create index next_front_is_cycle_index on next_front using btree (is_cycle);
end;
$$
  language plpgsql
  volatile
  strict;


create or replace function public.swap_root_front()
  returns void as
$$
begin
  alter table root_front
    rename to root_front_old;
  alter table next_front
    rename to root_front;
  alter table root_front_old
    rename to next_front;

  truncate table next_front;

  delete
  from root_front r
  where r.is_cycle
     or not exists(select 1 from edge e where e.end_id = r.next_id);

  return;
end;
$$
  language plpgsql
  volatile
  strict;

create or replace function public.swap_terminal_front()
  returns void as
$$
begin
  alter table terminal_front
    rename to terminal_front_old;
  alter table next_front
    rename to terminal_front;
  alter table terminal_front_old
    rename to next_front;

  truncate table next_front;

  delete
  from terminal_front r
  where r.is_cycle
     or not exists(select 1 from edge e where e.start_id = r.next_id);

  return;
end;
$$
  language plpgsql
  volatile
  strict;

create or replace function public.bidirectional_asp_harness(root_primer text, root_recursive text, terminal_primer text,
                                                            terminal_recursive text, max_depth int4)
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
  root_front_depth     int4 := 0;
  terminal_front_depth int4 := 0;
begin
  raise notice 'bidirectional_asp_harness start';

  -- Define two tables to represent pathspace of the recursive expansion. These are temporary and as such are unlogged.
  perform create_pathspace_tables();

  -- Populate the root front first with its primer query
  root_front_depth = root_front_depth + 1;
  execute root_primer;

  raise notice 'Expansion step % - Available Root Paths % - Num satisfied: %', root_front_depth + terminal_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

  if exists(select 1 from next_front r where r.satisfied) then
    -- Return all satisfied paths from the next front
    return query select * from next_front r where r.satisfied;
    return;
  end if;

  -- Swap the next_front table into the root_front
  perform swap_root_front();

  -- Populate the terminal front next with its primer query
  terminal_front_depth = terminal_front_depth + 1;
  execute terminal_primer;

  raise notice 'Expansion step % - Available Terminal Paths % - Num satisfied: %', root_front_depth + terminal_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

  -- Check to see if the two fronts meet somewhere in the middle
  if exists(select 1
            from next_front t
            where t.satisfied
            union
            select 1
            from next_front t
                   join root_front r on r.next_id = t.next_id) then
    -- Zip the path arrays together treating the matches as satisfied
    return query select r.root_id, t.root_id, r.depth + t.depth, true, false, r.path || t.path
                 from next_front t
                        join root_front r on r.next_id = t.next_id;
    return;
  end if;

  -- Swap the next_front table into the terminal_front
  perform swap_terminal_front();

  -- Iterate until one of the following conditions:
  -- * Traversal has reached the max allowed depth
  while root_front_depth + terminal_front_depth < max_depth and
        exists(select 1 from root_front union select 1 from terminal_front)
    loop
      if (select count(*) from root_front) < (select count(*) from terminal_front) then
        -- Populate the next front with the recursive root front query
        root_front_depth = root_front_depth + 1;
        execute root_recursive;

        raise notice 'Expansion step % - Available Root Paths % - Num satisfied: %', root_front_depth + terminal_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

        -- Check to see if the root front is satisfied
        if exists(select 1
                  from next_front r
                  where r.satisfied) then
          -- Return all satisfied paths from the next front
          return query select * from next_front r where r.satisfied;
          exit;
        end if;

        -- Swap the next_front table into the root_front
        perform swap_root_front();
      else
        -- Populate the next front with the recursive terminal front query
        terminal_front_depth = terminal_front_depth + 1;
        execute terminal_recursive;

        raise notice 'Expansion step % - Available Terminal Paths % - Num satisfied: %', root_front_depth + terminal_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

        -- Check to see if the two fronts meet somewhere in the middle
        if exists(select 1
                  from next_front t
                  where t.satisfied
                  union
                  select 1
                  from next_front t
                         join root_front r on r.next_id = t.next_id) then
          -- Zip the path arrays together treating the matches as satisfied
          return query select r.root_id, t.root_id, r.depth + t.depth, true, false, r.path || t.path
                       from next_front t
                              join root_front r on r.next_id = t.next_id;
          exit;
        end if;

        -- Swap the next_front table into the terminal_front
        perform swap_terminal_front();
      end if;
    end loop;
  return;
end;
$$
  language plpgsql volatile
                   strict;
