-- case: match (s:Kind1) return s
with n1 as (select n1.* from node n1 where n1.kind_ids operator (pg_catalog.&&) array []::int2[])
select (n1.id, n1.kind_ids, n1.properties)::nodecomposite as s
from n1;
