-- CASE: match ()-[r]->() return r
with n1 as (select n1.* from node n1),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id)
select (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as r
from e1;

-- CASE: match (n), ()-[r]->() return n, r
with n1 as (select n1.* from node n1),
     n2 as (select n2.* from node n2),
     e1 as (select e1.*
            from edge e1,
                 n2
            where n2.id = e1.start_id),
     n3 as (select n3.*
            from node n3,
                 e1
            where n3.id = e1.end_id)
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as r
from n1,
     e1;

-- case: match ()-[r]->(), ()-[e]->() return r, e
with n1 as (select n1.* from node n1),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id),
     n3 as (select n3.* from node n3),
     e2 as (select e2.*
            from edge e2,
                 n3
            where n3.id = e2.start_id),
     n4 as (select n4.*
            from node n4,
                 e2
            where n4.id = e2.end_id)
select (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as r,
       (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite as e
from e1,
     e2;

-- case: match ()-[r]->() where r.value = 42 return r
with n1 as (select n1.* from node n1),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id
              and e1.properties -> 'value' = 42),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id)
select (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as r
from e1;

-- case: match (n)-[r]->() where n.name = '123' return n, r
with n1 as (select n1.* from node n1 where n1.properties -> 'name' = '123'),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id)
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as n,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as r
from n1,
     e1;

-- case: match (s)-[r]->(e) where s.name = '123' and e.name = '321' return s, r, e
with n1 as (select n1.* from node n1 where n1.properties -> 'name' = '123'),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.start_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.end_id
              and n2.properties -> 'name' = '321')
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as s,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as r,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as e
from n1,
     e1,
     n2;


-- 	TODO: Unary expression `not f.bool_field` is going to require type casting to bool
-- case: match (f), (s)-[r]->(e) where not f.bool_field and s.name = '123' and e.name = '321' return f, s, r, e
with n1 as (select n1.* from node n1 where not n1.properties -> 'bool_field'),
     n2 as (select n2.* from node n2 where n2.properties -> 'name' = '123'),
     e1 as (select e1.*
            from edge e1,
                 n2
            where n2.id = e1.start_id),
     n3 as (select n3.*
            from node n3,
                 e1
            where n3.id = e1.end_id
              and n3.properties -> 'name' = '321')
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite                        as f,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as s,
       (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as r,
       (n3.id, n3.kind_ids, n3.properties)::nodecomposite                        as e
from n1,
     n2,
     e1,
     n3;

-- case: match ()-[e1]->(n)<-[e2]-() return e1, n, e2
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
select (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n,
       (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite as e2
from e1,
     n2,
     e2;

-- case: match ()-[e1]->(n)-[e2]->() return e1, n, e2
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
select (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n,
       (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite as e2
from e1,
     n2,
     e2;

-- case: match ()<-[e1]-(n)<-[e2]-() return e1, n, e2
with n1 as (select n1.* from node n1),
     e1 as (select e1.*
            from edge e1,
                 n1
            where n1.id = e1.end_id),
     n2 as (select n2.*
            from node n2,
                 e1
            where n2.id = e1.start_id),
     e2 as (select e2.*
            from edge e2,
                 n2
            where n2.id = e2.end_id),
     n3 as (select n3.*
            from node n3,
                 e2
            where n3.id = e2.start_id)
select (e1.id, e1.start_id, e1.end_id, e1.kind_id, e1.properties)::edgecomposite as e1,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite                        as n,
       (e2.id, e2.start_id, e2.end_id, e2.kind_id, e2.properties)::edgecomposite as e2
from e1,
     n2,
     e2;
