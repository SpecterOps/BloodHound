# openCypher to PgSQL 16 Translation

## Renaming

All variable bindings in an openCypher query are renamed to simplify reordering, decomposition and optimization
operations.

```cypher
match (n) return labels(n)

TODO: match (n)-[r:MemberOf]->()
```

The above openCypher query has the named variable `n` renamed to `n0` in the below SQL translation:

```postgresql
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
select (s0.n0).kind_ids
from s0;
```

## Scope Management

Translation requires some amount of query planning and result dependency negotiation. This translation driver organizes
openCypher query plans as an incrementally named variable in the SQL translation as `sN` where `N` is the scope's
frame.

Consider the following openCypher query:

```cypher
match (n), ()-[r]->() return n, r
```

The above openCypher query, when translated is broken up into two distinct planned elements. First, the plan executes a
select for `match (n)` with `n` renamed to `n0` and saves the projection under the result alias `s0` such that `s0.n0`
contains the result of `match (n) return n`:

```postgresql
s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0)
```

The second planned element executes a select and join for the `match ()-[r]->()` with `r` renamed to `e0`. The current
implement also eagerly binds all anonymous pattern elements in a given pattern part. The declarations are then saved
as a projection under the result alias `s1`. In addition to joining `n1`, `e0`, `n2` to the result scope, the previous
result `s0.n0` is also copied as `s1.n0`:

```postgresql
 s1 as (select s0.n0                                                                     as n0,
               (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1,
               (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
               (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
        from s0,
             edge e0
                 join node n1 on n1.id = e0.start_id
                 join node n2 on n2.id = e0.end_id)
```

Once the query reaches the final select statement, the following namespace is available authored in openCypher:
`match (n0), (n1)-[e0]->(n2)`. The resulting projection may then access them user the result alias `s1`:

```postgresql
select s1.n0 as n, s1.e0 as r
from s1;
```

The resulting translation being equal to: `return n, r`. The translation strives to preserve user preference with regard
to projection aliases. In the case of the original openCypher query, the result set the querying client expects is:
`(n), [r]` where `n` is a node and `r` is an edge.

The complete translation is:

```postgresql
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0),
     s1 as (select s0.n0                                                                     as n0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n2
            from s0,
                 edge e0
                     join node n1 on n1.id = e0.start_id
                     join node n2 on n2.id = e0.end_id)
select s1.n0 as n, s1.e0 as r
from s1;
```

## Criteria Decomposition

openCypher formally specifies bound elements of a query in the pattern component: `match (s)-[r]->(e) return s, r, e`
and allows referencing of the components in openCpyher expressions:
`match (s)-[r]->(e) where s.name = '123' return s, r, e`.

The translator isolates filtering criteria for bound elements in a data structure named a `constraint`. Constraints
are collected and organized by the bound elements they reference. This allows the translator to reduce, deconstruct and
reorder filtering criteria.

For example, consider the following openCypher query:

`match (s)-[r]->(e) where s.name = '123' and e:Target and not r.property return s, r, e`.

The above openCypher query would first undergo renaming:

`match (n0)-[e0]->(n1) where n0.name = '123' and n1:Target and not e0.property return n0 as s, e0 as r, n1 as e`.

After renaming the translator will isolate the following constraints:

* `(n0) where n0.name = '123'`
* `[e0] where not e0.property`
* `(n1) where n1:Target`

After renaming the translator, will construct the required SQL AST nodes and assigns constraints to their associated
select operations. This is done by comparing the available scope declarations and the requirements of the constraints:

```postgresql
-- match (s)-[r]->(e) where s.name = '123' and e:NodeKind1 and not r.property return s, r, e
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n0,
                   (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n1
            from edge e0
                     join node n0 on n0.properties ->> 'name' = '123' and n0.id = e0.start_id
                     join node n1 on n1.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n1.id = e0.end_id
            where not (e0.properties -> 'property')::bool)
select s0.n0 as s, s0.e0 as r, s0.n1 as e
from s0;
```

### Dependency Satisfaction

Query filter expressions may involve operations that can not be isolated to a single bound element. Take for example
the following openCypher query:
`match (s:NodeKind1), (e:NodeKind2) where s.selected or s.tid = e.tid and e.enabled return s, e`

The above openCypher query would first undergo renaming:

`match (n0:NodeKind1), (n1:NodeKind2) where n0.selected or n0.tid = n1.tid and n1.enabled return n0 as s, n1 as e`

After renaming the translator will isolate the following constraints:

* `(n0) where n0:NodeKind1`
* `(n1) where n1:NodeKind2`
* `(n0), (n1) where n0.selected or n0.tid = n1.tid and n1.enabled`

The comparison `n0.selected or n0.tid = n1.tid` can not be decomposed further and must have both `n0` and `n1` projected
in scope before it can be satisfied. When the query plan first selects for `n0` it can only satisfy the constraint
`(n0) where n0:NodeKind1`:

```postgresql
s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.kind_ids operator (pg_catalog.&&) array [1]::int2[])
```

The second query will project `n1` and make it available. This then allows the planner to satisfy both remaining
constraints: `(n1) where n1:NodeKind2` and `(n0), (n1) where n0.selected or n0.tid = n1.tid`:

```postgresql
s1 as (select s0.n0 as n0, (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from s0,
                 node n1
            where n1.kind_ids operator (pg_catalog.&&) array [2]::int2[] and ((s0.n0).properties -> 'selected')::bool
               or (s0.n0).properties -> 'tid' = n1.properties -> 'tid' and (n1.properties -> 'enabled')::bool)
```

The complete translation is:

```postgresql
-- match (s:NodeKind1), (e:NodeKind2) where s.selected or s.tid = e.tid and e.enabled return s, e
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
```

## Expansions

* `packages/go/cypher/models/pgsql/translate/expansion.go`
* `packages/go/cypher/models/pgsql/translate/translation_cases/pattern_expansion.sql`

The translator represents pattern expansion as
a [recursive CTE](https://www.postgresql.org/docs/current/queries-with.html#QUERIES-WITH-RECURSIVE) to offload as much
of the traversal work to the database. Currently, all expansions are hard limited to an expansion depth of 5 steps.

Consider the following openCypher query: `match (n)-[*..]->(e) return n, e`.

To perform the recursive traversal, the translator first defines a record shape for the CTE to represent the expansion's
pathspace:

```postgresql
with s0 as (with recursive ex0(root_id, next_id, depth, satisfied, is_cycle, path)
```

| Column      | type    | Usage                                                                                  |
|-------------|---------|----------------------------------------------------------------------------------------|
| `root_id`   | Int8    | Node that the path originated from. Simplifies referencing the root node of each path. |
| `next_id`   | Int8    | Next node to expand to.                                                                |
| `depth`     | Int     | Depth of the current traversal.                                                        |
| `satisfied` | Boolean | True if the expansion is satisfied.                                                    |
| `is_cycle`  | Boolean | True if the expansion is a cycle.                                                      |
| `path`      | Int8[]  | Array of edges in order of traversal.                                                  |

The translator then formats two queries. First is the primer query that populates the initial pathspace of the
expansion:

```postgresql
select e0.start_id,
       e0.end_id,
       1,
       false,
       e0.start_id = e0.end_id,
       array [e0.id]
from edge e0
         join node n0 on n0.id = e0.start_id
         join node n1 on n1.id = e0.end_id
```

This query is then unioned to a recursive query. Unlike the primer query, the recursive query also includes the
temporary table defined by the pathspace record shape `ex0`:

```postgresql
select ex0.root_id,
       e0.end_id,
       ex0.depth + 1,
       false,
       e0.id = any (ex0.path),
       ex0.path || e0.id
from ex0
         join edge e0 on e0.start_id = ex0.next_id
         join node n1 on n1.id = e0.end_id
where ex0.depth < 5
  and not ex0.is_cycle
```

The resulting output of the result alias `s0` is then populated from the pathspace table:

```postgresql
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
       (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
        from edge e0
        where e0.id = any (ex0.path))                     as e0,
       ex0.path                                           as ep0,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
from ex0
         join edge e0 on e0.id = any (ex0.path)
         join node n0 on n0.id = ex0.root_id
         join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int4] and n1.id = e0.end_id
```

This binds the following additional components to query scope:

| Component | Type            | Usage                                            |
|-----------|-----------------|--------------------------------------------------|
| `n0`      | nodecomposite   | Root node of each recursively expanded path.     |
| `e0`      | edgecomposite[] | edgecomposite array of edges                     |
| `ep0`     | int4[]          | Int4 array of edge identifiers                   |
| `n1`      | nodecomposite   | Terminal node of each recursively expanded path. |

The translator then authors the final projection `return n, e`:

```postgresql
select s0.n0 as n, s0.n1 as e
from s0;
```

The complete translation is:

```postgresql
-- match (n)-[*..]->(e) return n, e
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
                                                                                       where ex0.depth < 5
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
                     join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int4] and n1.id = e0.end_id)
select s0.n0 as n, s0.n1 as e
from s0;
```

### All Shortest Paths

* `packages/go/cypher/models/pgsql/translate/expansion.go`
* `packages/go/cypher/models/pgsql/translate/translation_cases/shortest_paths.sql`

Recursive harness in PL/pgSQL with non-parameterized primer and recursive queries.

All shortest paths is a search function that will return the first path found and all other paths of the same depth.
Naively, this early termination can be authored as a condition in the temporary table of a recursive expansion CTE
however recursive CTEs may not perform certain operations on the temporary table allocated for the CTE
statement: https://www.postgresql.org/docs/current/queries-with.html#QUERIES-WITH-RECURSIVE.

This means a different construct is required to perform early termination of a shortest-path traversal. For all shortest
paths queries, a PL/PgSQL function named `asp_harness` exists to execute recursive expansion with early termination.

The goal of this effort, even with the caveats of the method pursued, is to prevent the need for unnecessary round trips
to the database by offloading as much of the expansion logic to the database as possible. This may result in a sizeable
query that is less efficient than its constituent components if reordered, but it allows the implementation to avoid
hundreds of thousands or millions of round trips to execute a traversal expansion.

#### `asp_harness`

```postgresql
-- All shortest path traversal harness.
create or replace function public.asp_harness(primer_query text, recursive_query text, max_depth int4)
    returns table
            (
                root_id   int4,
                next_id   int4,
                depth     int4,
                satisfied bool,
                is_cycle  bool,
                path      int4[]
            )
as
$$
declare
    depth int4 := 1;
begin
    -- Define two tables to represent pathspace of the recursive expansion. These are temporary and as such are unlogged.
    create temporary table pathspace
    (
        root_id   int4   not null,
        next_id   int4   not null,
        depth     int4   not null,
        satisfied bool   not null,
        is_cycle  bool   not null,
        path      int4[] not null,
        primary key (path)
    ) on commit drop;

    create temporary table next_pathspace
    (
        root_id   int4   not null,
        next_id   int4   not null,
        depth     int4   not null,
        satisfied bool   not null,
        is_cycle  bool   not null,
        path      int4[] not null,
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

    -- Populate initial pathspace with the primer query
    execute primer_query;

    -- Iterate until either we reach a depth limit or if there are satisfied paths
    while depth < max_depth and not exists(select 1 from next_pathspace np where np.satisfied)
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

            -- Remove any non-satisfied terminals and cycles from pathspace to prune any terminal, non-satisfied paths
            delete
            from pathspace p
            where p.is_cycle
               or not exists(select 1 from edge e where e.start_id = p.next_id);

            -- Increase the current depth and execute the recursive query
            depth := depth + 1;
            execute recursive_query;
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
```

#### Expansion using the ASP Harness

Take for consideration the following openCypher query: `match p = allShortestPaths((s:NodeKind1)-[*..]->()) return p`

The table record shape of the `asp_harness` function matches the recursive CTE traversal's record shape, meaning the
result alias for this expansion can be dropped into a chain of expansions with minimal result set processing. As such
the result alias record shape for an all shortest paths query is declared no differently than a recursive CTE traversal:

```postgresql
with ex0(root_id, next_id, depth, satisfied, is_cycle, path)
```

The harness function is then called. The first two parameters are expected to be `Text` values that contain reified
statements that represent the primer and recursive queries for the traversal expansion. In the test case being explored,
these parameter expectations are encoded in the test case header:

```postgresql
-- pgsql_params: {"pi0":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select e0.start_id, e0.end_id, 1, exists (select 1 from edge e0 where n1.id = e0.start_id), e0.start_id = e0.end_id, array [e0.id] from edge e0 join node n0 on n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.id = e0.start_id join node n1 on n1.id = e0.end_id;", "pi1":"insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path) select ex0.root_id, e0.end_id, ex0.depth + 1, exists (select 1 from edge e0 where n1.id = e0.start_id), e0.id = any (ex0.path), ex0.path || e0.id from pathspace ex0 join edge e0 on e0.start_id = ex0.next_id join node n1 on n1.id = e0.end_id where ex0.depth < 5 and not ex0.is_cycle;"}
```

For this translation the primer query will be formated as the `pi0` parameter:

```postgresql
insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path)
select e0.start_id,
       e0.end_id,
       1,
       exists (select 1 from edge e0 where n1.id = e0.start_id),
       e0.start_id = e0.end_id,
       array [e0.id]
from edge e0
         join node n0 on n0.kind_ids operator (pg_catalog.&&) array [1]::int2[] and n0.id = e0.start_id
         join node n1 on n1.id = e0.end_id;
```

The translation then authors the recursive query formatted as the `pi1` parameter:

```postgresql
insert into next_pathspace (root_id, next_id, depth, satisfied, is_cycle, path)
select ex0.root_id,
       e0.end_id,
       ex0.depth + 1,
       exists (select 1 from edge e0 where n1.id = e0.start_id),
       e0.id = any (ex0.path),
       ex0.path || e0.id
from pathspace ex0
         join edge e0 on e0.start_id = ex0.next_id
         join node n1 on n1.id = e0.end_id
where ex0.depth < 5
  and not ex0.is_cycle;
```

The resulting output of the result alias `s0` is then populated from the pathspace table:

```postgresql
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
       (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
        from edge e0
        where e0.id = any (ex0.path))                     as e0,
       ex0.path                                           as ep0,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
from ex0
         join edge e0 on e0.id = any (ex0.path)
         join node n0 on n0.id = ex0.root_id
         join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int4] and n1.id = e0.end_id
```

This binds the following additional components to query scope:

| Component | Type            | Usage                                            |
|-----------|-----------------|--------------------------------------------------|
| `n0`      | nodecomposite   | Root node of each recursively expanded path.     |
| `e0`      | edgecomposite[] | edgecomposite array of edges                     |
| `ep0`     | int4[]          | Int4 array of edge identifiers                   |
| `n1`      | nodecomposite   | Terminal node of each recursively expanded path. |

Lastly, since this query is to return the traversed paths, the result variable `ep0` is passed to a function
`edges_to_path` to create the expected return record shape:

```postgresql
select edges_to_path(variadic ep0)::pathcomposite as p
from s0;
```

The complete translation is:

```postgresql
-- match p = allShortestPaths((s:NodeKind1)-[*..]->()) return p
with s0 as (with ex0(root_id, next_id, depth, satisfied, is_cycle, path)
                     as (select * from asp_harness(@pi0::text, @pi1::text, 5))
            select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0,
                   (select array_agg((e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite)
                    from edge e0
                    where e0.id = any (ex0.path))                     as e0,
                   ex0.path                                           as ep0,
                   (n1.id, n1.kind_ids, n1.properties)::nodecomposite as n1
            from ex0
                     join edge e0 on e0.id = any (ex0.path)
                     join node n0 on n0.id = ex0.root_id
                     join node n1 on e0.id = ex0.path[array_length(ex0.path, 1)::int4] and n1.id = e0.end_id)
select edges_to_path(variadic ep0)::pathcomposite as p
from s0;
```
