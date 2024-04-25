-- CASE: match ()-[r]->() return r
with n0 as (select n0.* from node n0),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id)
select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as r
from e0;

-- CASE: match (n), ()-[r]->() return n, r
with n0 as (select n0.* from node n0),
     n1 as (select n1.* from node n1),
     e0 as (select e0.*
            from edge e0,
                 n1
            where n1.id = e0.start_id),
     n2 as (select n2.*
            from node n2,
                 e0
            where n2.id = e0.end_id)
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n,
       (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as r
from n0,
     e0;

-- case: match ()-[r]->(), ()-[e]->() return r, e
with n0 as (select n0.* from node n0),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id),
     n2 as (select n2.* from node n2),
     e1 as (select e1.*
            from edge e1,
                 n2
            where n2.id = e1.start_id),
     n3 as (select n3.*
            from node n3,
                 e1
            where n3.id = e1.end_id)
select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as r,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e
from e0,
     e1;

-- case: match ()-[r]->() where r.value = 42 return r
with n0 as (select n0.* from node n0),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id
              and e0.properties -> 'value' = 42),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id)
select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as r
from e0;

-- case: match (n)-[r]->() where n.name = '123' return n, r
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = '123'),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id)
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as n,
       (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as r
from n0,
     e0;

-- case: match (s)-[r]->(e) where s.name = '123' and e.name = '321' return s, r, e
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = '123'),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.start_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.end_id
              and n1.properties -> 'name' = '321')
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as s,
       (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as r,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as e
from n0,
     e0,
     n1;


-- 	TODO: Unary expression `not f.bool_field` is going to require type casting to bool
-- case: match (f), (s)-[r]->(e) where not f.bool_field and s.name = '123' and e.name = '321' return f, s, r, e
with n0 as (select n0.* from node n0 where not n0.properties -> 'bool_field'),
     n1 as (select n1.* from node n1 where n1.properties -> 'name' = '123'),
     e0 as (select e0.*
            from edge e0,
                 n1
            where n1.id = e0.start_id),
     n2 as (select n2.*
            from node n2,
                 e0
            where n2.id = e0.end_id
              and n2.properties -> 'name' = '321')
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite                        as f,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as s,
       (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as r,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as e
from n0,
     n1,
     e0,
     n2;

-- case: match ()-[e0]->(n)<-[e1]-() return e0, n, e1
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
select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1
from e0,
     n1,
     e1;

-- case: match ()-[e0]->(n)-[e1]->() return e0, n, e1
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
select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1
from e0,
     n1,
     e1;

-- case: match ()<-[e0]-(n)<-[e1]-() return e0, n, e1
with n0 as (select n0.* from node n0),
     e0 as (select e0.*
            from edge e0,
                 n0
            where n0.id = e0.end_id),
     n1 as (select n1.*
            from node n1,
                 e0
            where n1.id = e0.start_id),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.end_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.start_id)
select (e0.id, e0.start_id, e0.end_id, e0.kind_id, e0.properties)::edgecomposite as e0,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1
from e0,
     n1,
     e1;
