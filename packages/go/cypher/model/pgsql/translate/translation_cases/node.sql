-- case: match (s) return s
with n0 as (select n0.* from node n0)
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as s
from n0;

-- case: match (s) where s.name = '1234' return s
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = '1234')
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as s
from n0;

-- case: match (s), (e) where s.name = '1234' return s
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = '1234'),
     n1 as (select n1.* from node n1)
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as s
from n0;

-- case: match (s:A), (e:B) where s.name = e.name return s, e
with n0 as (select n0.* from node n0 where n0.kind_ids operator (pg_catalog.&&) array []::int2[]),
     n1 as (select n1.*
            from node n1,
                 n0
            where n0.properties -> 'name' = n1.properties -> 'name'
              and n1.kind_ids operator (pg_catalog.&&) array []::int2[])
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as s,
       (n1.id, n1.kind_ids, n1.properties)::nodecomposite as e
from n0,
     n1;

-- case: match (s), (e) where s.name = '1234' and e.other = 1234 return s
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = '1234'),
     n1 as (select n1.* from node n1 where n1.properties -> 'other' = 1234)
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as s
from n0;

-- case: match (n), (k) where n.name = '1234' and k.name = '1234' match (e) where e.name = n.name return k, e
with n0 as (select n0.* from node n0 where n0.properties -> 'name' = '1234'),
     n1 as (select n1.* from node n1 where n1.properties -> 'name' = '1234'),
     n2 as (select n2.*
            from node n2,
                 n0
            where n2.properties -> 'name' = n0.properties -> 'name')
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite as k,
       (n2.id, n2.kind_ids, n2.properties)::nodecomposite as e
from n1,
     n2;
