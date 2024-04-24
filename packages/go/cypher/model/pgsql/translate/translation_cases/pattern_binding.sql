-- case: match p = ()-[]->() return p
with n0 as (select n0.* from node n0),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id)
select (array [(n0.id, n0.kind_ids, n0.properties)::nodecomposite, (n1.id, n1.kind_ids, n1.properties)::nodecomposite],
        array [(e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite])::pathcomposite as p
from n0,
     e0,
     n1
where n0.id = e0.start_id
  and n1.id = e0.end_id;

-- case: match p = ()-[]->()-[]->() return p
with n0 as (select n0.* from node n0),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id)
select (array [(n0.id, n0.kind_ids, n0.properties)::nodecomposite, (n1.id, n1.kind_ids, n1.properties)::nodecomposite, (n2.id, n2.kind_ids, n2.properties)::nodecomposite],
        array [(e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite, (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite])::pathcomposite as p
from n0,
     e0,
     n1,
     e1,
     n2
where n1.id = e1.start_id
  and n2.id = e1.end_id
  and n0.id = e0.start_id
  and n1.id = e0.end_id;

-- case: match p = ()-[]->()<-[]-() return p
with n0 as (select n0.* from node n0),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.end_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.start_id)
select (array [(n0.id, n0.kind_ids, n0.properties)::nodecomposite, (n1.id, n1.kind_ids, n1.properties)::nodecomposite, (n2.id, n2.kind_ids, n2.properties)::nodecomposite],
        array [(e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite, (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite])::pathcomposite as p
from n0,
     e0,
     n1,
     e1,
     n2
where n1.id = e1.end_id
  and n2.id = e1.start_id
  and n0.id = e0.start_id
  and n1.id = e0.end_id;

-- case: match p = (a)-[]->()<-[]-(f) where a.name = 'value' and f.is_target return p
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = 'value'),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.end_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.start_id
              and n2.properties -> 'is_target')
select (array [(n0.id, n0.kind_ids, n0.properties)::nodecomposite, (n1.id, n1.kind_ids, n1.properties)::nodecomposite, (n2.id, n2.kind_ids, n2.properties)::nodecomposite],
        array [(e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite, (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite])::pathcomposite as p
from n0,
     e0,
     n1,
     e1,
     n2
where n1.id = e1.end_id
  and n2.id = e1.start_id
  and n0.id = e0.start_id
  and n1.id = e0.end_id;
