-- case: match (s) return s
with n1 as (select n1.* from node n1)
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite as s
from n1;

-- case: match (s) where s.name = '1234' return s
with n1 as (select n1.* from node n1 where n1.properties -> 'name' = '1234')
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite as s
from n1;

-- case: match (s), (e) where s.name = '1234' return s
with n1 as (select n1.* from node n1 where n1.properties -> 'name' = '1234'),
     n2 as (select n2.* from node n2)
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite as s
from n1;

-- case: match (s:A), (e:B) where s.name = e.name return s, e
with n1 as (select n1.* from node n1 where n1.kind_ids operator (pg_catalog.&&) array []::int2[]),
     n2 as (select n2.*
            from node n2,
                 n1
            where n1.properties -> 'name' = n2.properties -> 'name'
              and n2.kind_ids operator (pg_catalog.&&) array []::int2[])
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite as s,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite as e
from n1,
     n2;

-- case: match (s), (e) where s.name = '1234' and e.other = 1234 return s
with n1 as (select n1.* from node n1 where n1.properties -> 'name' = '1234'),
     n2 as (select n2.* from node n2 where n2.properties -> 'other' = 1234)
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite as s
from n1;

-- case: match (n), (k) where n.name = '1234' and k.name = '1234' match (e) where e.name = n.name return k, e
with n1 as (select n1.* from node n1 where n1.properties -> 'name' = '1234'),
     n2 as (select n2.* from node n2 where n2.properties -> 'name' = '1234'),
     n3 as (select n3.*
            from node n3,
                 n1
            where n3.properties -> 'name' = n1.properties -> 'name')
select (n2.id, n2.kind_ids, n2.properties)::nodecomposite as k,
       (n3.id, n3.kind_ids, n3.properties)::nodecomposite as e
from n2,
     n3;
