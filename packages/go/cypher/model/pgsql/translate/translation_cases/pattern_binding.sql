-- case: match p = ()-[]->() return p
with n1 as (select n1.* from node n1),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id)
select (array [(n1.id, n1.kind_ids, n1.properties)::nodecomposite, (n2.id, n2.kind_ids, n2.properties)::nodecomposite],
        array [(e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite])::pathcomposite as p
from n1,
     e1,
     n2
where n1.id = e1.start_id
  and n2.id = e1.end_id;

-- case: match p = ()-[]->()-[]->() return p
with n1 as (select n1.* from node n1),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id),
     e2 as (select e2.*
            from edge e2,
                 n2
            where n2.id = e2.start_id),
     n3 as (select n3.*
            from node n3,
                 e2
            where n3.id = e2.end_id)
select (array [(n1.id, n1.kind_ids, n1.properties)::nodecomposite, (n2.id, n2.kind_ids, n2.properties)::nodecomposite, (n3.id, n3.kind_ids, n3.properties)::nodecomposite],
        array [(e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite, (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite])::pathcomposite as p
from n1,
     e1,
     n2,
     e2,
     n3
where n2.id = e2.start_id
  and n3.id = e2.end_id
  and n1.id = e1.start_id
  and n2.id = e1.end_id;

-- case: match p = ()-[]->()<-[]-() return p
with n1 as (select n1.* from node n1),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id),
     e2 as (select e2.*
            from edge e2,
                 n2
            where n2.id = e2.end_id),
     n3 as (select n3.*
            from node n3,
                 e2
            where n3.id = e2.start_id)
select (array [(n1.id, n1.kind_ids, n1.properties)::nodecomposite, (n2.id, n2.kind_ids, n2.properties)::nodecomposite, (n3.id, n3.kind_ids, n3.properties)::nodecomposite],
        array [(e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite, (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite])::pathcomposite as p
from n1,
     e1,
     n2,
     e2,
     n3
where n2.id = e2.end_id
  and n3.id = e2.start_id
  and n1.id = e1.start_id
  and n2.id = e1.end_id;

-- case: match p = (a)-[]->()<-[]-(f) where a.name = 'value' and f.is_target return p
with n1 as (select n1.* from node n1 where n1.properties -> 'name' = 'value'),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id),
     e2 as (select e2.*
            from edge e2,
                 n2
            where n2.id = e2.end_id),
     n3 as (select n3.*
            from node n3,
                 e2
            where n3.id = e2.start_id
              and n3.properties -> 'is_target')
select (array [(n1.id, n1.kind_ids, n1.properties)::nodecomposite, (n2.id, n2.kind_ids, n2.properties)::nodecomposite, (n3.id, n3.kind_ids, n3.properties)::nodecomposite],
        array [(e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite, (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite])::pathcomposite as p
from n1,
     e1,
     n2,
     e2,
     n3
where n2.id = e2.end_id
  and n3.id = e2.start_id
  and n1.id = e1.start_id
  and n2.id = e1.end_id;
