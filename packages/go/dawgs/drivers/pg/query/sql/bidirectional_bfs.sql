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

drop function if exists public.create_unidirectional_pathspace_tables();
drop function if exists public.create_bidirectional_pathspace_tables();
drop function if exists public.swap_forward_front();
drop function if exists public.swap_backward_front();
drop function if exists public.unidirectional_asp_harness(forward_primer text, forward_recursive text, max_depth int4);
drop function if exists public.bidirectional_asp_harness(forward_primer text, forward_recursive text,
                                                         backward_primer text, backward_recursive text, max_depth int4);

--
create or replace function public.create_unidirectional_pathspace_tables()
  returns void as
$$
begin
  create temporary table forward_front
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

  create index forward_front_next_id_index on forward_front using btree (next_id);
  create index forward_front_satisfied_index on forward_front using btree (satisfied);
  create index forward_front_is_cycle_index on forward_front using btree (is_cycle);

  create index next_front_next_id_index on next_front using btree (next_id);
  create index next_front_satisfied_index on next_front using btree (satisfied);
  create index next_front_is_cycle_index on next_front using btree (is_cycle);
end;
$$
  language plpgsql
  volatile
  strict;

create or replace function public.create_bidirectional_pathspace_tables()
  returns void as
$$
begin
  perform create_unidirectional_pathspace_tables();

  create temporary table backward_front
  (
    root_id   int8   not null,
    next_id   int8   not null,
    depth     int4   not null,
    satisfied bool   not null,
    is_cycle  bool   not null,
    path      int8[] not null,
    primary key (path)
  ) on commit drop;

  create index backward_front_next_id_index on backward_front using btree (next_id);
  create index backward_front_satisfied_index on backward_front using btree (satisfied);
  create index backward_front_is_cycle_index on backward_front using btree (is_cycle);
end;
$$
  language plpgsql
  volatile
  strict;


create or replace function public.swap_forward_front()
  returns void as
$$
begin
  alter table forward_front
    rename to forward_front_old;
  alter table next_front
    rename to forward_front;
  alter table forward_front_old
    rename to next_front;

  truncate table next_front;

  delete
  from forward_front r
  where r.is_cycle
     or not exists(select 1 from edge e where e.end_id = r.next_id);

  return;
end;
$$
  language plpgsql
  volatile
  strict;

create or replace function public.swap_backward_front()
  returns void as
$$
begin
  alter table backward_front
    rename to backward_front_old;
  alter table next_front
    rename to backward_front;
  alter table backward_front_old
    rename to next_front;

  truncate table next_front;

  delete
  from backward_front r
  where r.is_cycle
     or not exists(select 1 from edge e where e.start_id = r.next_id);

  return;
end;
$$
  language plpgsql
  volatile
  strict;

create or replace function public.unidirectional_asp_harness(forward_primer text, forward_recursive text, max_depth int4)
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
  forward_front_depth int4 := 0;
begin
  raise notice 'unidirectional_asp_harness start';

  -- Defines two tables to represent pathspace of the recursive expansion
  perform create_unidirectional_pathspace_tables();

  -- Populate the root front first with its primer query
  forward_front_depth = forward_front_depth + 1;
  execute forward_primer;

  raise notice 'Expansion step % - Available Root Paths % - Num satisfied: %', forward_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

  if exists(select 1 from next_front r where r.satisfied) then
    -- Return all satisfied paths from the next front
    return query select * from next_front r where r.satisfied;
    return;
  end if;

  -- Swap the next_front table into the forward_front
  perform swap_forward_front();

  while forward_front_depth < max_depth and exists(select 1 from forward_front)
    loop
      -- Populate the next front with the recursive root front query
      forward_front_depth = forward_front_depth + 1;
      execute forward_recursive;

      raise notice 'Expansion step % - Available Root Paths % - Num satisfied: %', forward_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

      -- Check to see if the root front is satisfied
      if exists(select 1
                from next_front r
                where r.satisfied) then
        -- Return all satisfied paths from the next front
        return query select * from next_front r where r.satisfied;
        exit;
      end if;

      -- Swap the next_front table into the forward_front
      perform swap_forward_front();
    end loop;
  return;
end;
$$
  language plpgsql volatile
                   strict;

create or replace function public.bidirectional_asp_harness(forward_primer text, forward_recursive text,
                                                            backward_primer text,
                                                            backward_recursive text, max_depth int4)
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
  forward_front_depth  int4 := 0;
  backward_front_depth int4 := 0;
begin
  raise notice 'bidirectional_asp_harness start';

  -- Defines three tables to represent pathspace of the recursive expansion
  perform create_bidirectional_pathspace_tables();

  -- Populate the root front first with its primer query
  forward_front_depth = forward_front_depth + 1;
  execute forward_primer;

  raise notice 'Expansion step % - Available Root Paths % - Num satisfied: %', forward_front_depth + backward_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

  if exists(select 1 from next_front r where r.satisfied) then
    -- Return all satisfied paths from the next front
    return query select * from next_front r where r.satisfied;
    return;
  end if;

  -- Swap the next_front table into the forward_front
  perform swap_forward_front();

  -- Populate the terminal front next with its primer query
  backward_front_depth = backward_front_depth + 1;
  execute backward_primer;

  raise notice 'Expansion step % - Available Terminal Paths % - Num satisfied: %', forward_front_depth + backward_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

  -- Check to see if the two fronts meet somewhere in the middle
  if exists(select 1
            from next_front t
            where t.satisfied
            union
            select 1
            from next_front t
                   join forward_front r on r.next_id = t.next_id) then
    -- Zip the path arrays together treating the matches as satisfied
    return query select r.root_id, t.root_id, r.depth + t.depth, true, false, r.path || t.path
                 from next_front t
                        join forward_front r on r.next_id = t.next_id;
    return;
  end if;

  -- Swap the next_front table into the backward_front
  perform swap_backward_front();

  while forward_front_depth + backward_front_depth < max_depth and
        exists(select 1 from forward_front union select 1 from backward_front)
    loop
      if (select count(*) from forward_front) < (select count(*) from backward_front) then
        -- Populate the next front with the recursive root front query
        forward_front_depth = forward_front_depth + 1;
        execute forward_recursive;

        raise notice 'Expansion step % - Available Root Paths % - Num satisfied: %', forward_front_depth + backward_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

        -- Check to see if the root front is satisfied
        if exists(select 1
                  from next_front r
                  where r.satisfied) then
          -- Return all satisfied paths from the next front
          return query select * from next_front r where r.satisfied;
          exit;
        end if;

        -- Swap the next_front table into the forward_front
        perform swap_forward_front();
      else
        -- Populate the next front with the recursive terminal front query
        backward_front_depth = backward_front_depth + 1;
        execute backward_recursive;

        raise notice 'Expansion step % - Available Terminal Paths % - Num satisfied: %', forward_front_depth + backward_front_depth, (select count(*) from next_front), (select count(*) from next_front p where p.satisfied);

        -- Check to see if the two fronts meet somewhere in the middle
        if exists(select 1
                  from next_front t
                  where t.satisfied
                  union
                  select 1
                  from next_front t
                         join forward_front r on r.next_id = t.next_id) then
          -- Zip the path arrays together treating the matches as satisfied
          return query select r.root_id, t.root_id, r.depth + t.depth, true, false, r.path || t.path
                       from next_front t
                              join forward_front r on r.next_id = t.next_id;
          exit;
        end if;

        -- Swap the next_front table into the backward_front
        perform swap_backward_front();
      end if;
    end loop;
  return;
end;
$$
  language plpgsql volatile
                   strict;
