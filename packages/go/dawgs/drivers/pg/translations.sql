-- match (q) where a:User return a
with recursive expansion_1(root_id, next_id, depth, stop, is_cycle, path) as (
  select r.start_id, r.end_id, 1, false, r.start_id = r.end_id, array[r.id]
  from edge r
         join node a on a.id = r.start_id
  where a.kind_ids operator(pg_catalog.&&) array[23]::int2[]
)
select a.properties from expansion_1
  join node a on a.id = expansion_1.root_id
where not expansion_1.is_cycle;

-- match (a)-[*..]->(b)-[*..]->(t) where a:User and b:Computer and t:User return a, b, t
with recursive expansion_1(root_id, next_id, depth, stop, is_cycle, interstitial, path) as (
  select r.start_id, r.end_id, 1, false, r.start_id = r.end_id, array[]::int4[], array[r.id]
  from edge r
         join node a on a.id = r.start_id
  where a.kind_ids operator(pg_catalog.&&) array[23]::int2[]
union all
  select expansion_1.root_id, r.end_id, expansion_1.depth + 1, b.kind_ids operator(pg_catalog.&&) array[24]::int2[], r.id = any(expansion_1.path), array[r.end_id], expansion_1.path || r.id
  from expansion_1
         join edge r on r.start_id = expansion_1.next_id
         join node b on b.id = r.end_id
  where not expansion_1.is_cycle and not expansion_1.stop
),
expansion_2(root_id, next_id, depth, stop, is_cycle, interstitial, path) as (
  select expansion_1.root_id, expansion_1.next_id, expansion_1.depth, false, expansion_1.is_cycle, expansion_1.interstitial || 0, expansion_1.path
  from expansion_1
  where not expansion_1.is_cycle and expansion_1.stop
union all
  select expansion_2.root_id, r.end_id, expansion_2.depth + 1, t.kind_ids operator(pg_catalog.&&) array[23]::int2[], r.id = any(expansion_2.path), trim_array(expansion_2.interstitial, 1) || array[r.end_id], expansion_2.path || r.id
  from expansion_2
        join edge r on r.start_id = expansion_2.next_id
        join node t on t.id = r.end_id
  where not expansion_2.is_cycle and not expansion_2.stop
)
select a.properties, b.properties, t.properties from expansion_2
    join node a on a.id = expansion_2.root_id
    join node t on t.id = expansion_2.next_id
    join node b on b.id = expansion_2.interstitial[1]
where not expansion_2.is_cycle and expansion_2.stop;

-- match (a)-[*..]->(b) where a:User and b:Computer return a, b
with recursive expansion_1(root_id, next_id, depth, stop, is_cycle, path) as (
  select r.start_id, r.end_id, 1, false, r.start_id = r.end_id, array[r.id]
  from edge r
         join node a on a.id = r.start_id
  where a.kind_ids operator(pg_catalog.&&) array[23]::int2[]
  union all
  select expansion_1.root_id, r.end_id, expansion_1.depth + 1, b.kind_ids operator(pg_catalog.&&) array[24]::int2[], r.id = any(expansion_1.path), expansion_1.path || r.id
  from expansion_1
         join edge r on r.start_id = expansion_1.next_id
         join node b on b.id = r.end_id
  where not expansion_1.is_cycle and not expansion_1.stop
)
select a.properties, b.properties from expansion_1
       join node a on a.id = expansion_1.root_id
       join node b on b.id = expansion_1.next_id
where not expansion_1.is_cycle and expansion_1.stop;
