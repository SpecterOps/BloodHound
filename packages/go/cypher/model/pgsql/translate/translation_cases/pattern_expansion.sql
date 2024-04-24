-- case: match (n)-[*..]->(e:Kind) where n.name = '1234' return e
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = '1234'),
     e0 as (with recursive e0(root_id, next_id, depth, satisfied, is_cycle, path) as (select e0.start_id,
                                                                                             e0.end_id,
                                                                                             1,
                                                                                             false,
                                                                                             e0.start_id = e0.end_id,
                                                                                             array [e0.id]
                                                                                      from edge e0,
                                                                                           n0
                                                                                      where n0.id = e0.start_id
                                                                                      union all
                                                                                      select e0.root_id,
                                                                                             e0.end_id,
                                                                                             e0.depth + 1,
                                                                                             n0.kind_ids operator (pg_catalog.&&) array []::int2[],
                                                                                             e0.id = any (e0.path),
                                                                                             e0.path || e0.id
                                                                                      from e0,
                                                                                           edge e0
                                                                                             join node n0 on n0.id = e0.end_id
                                                                                      where not e0.is_cycle
                                                                                        and not e0.satisfied)
            select e0.*
            from e0)
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as e
from n0;
