-- case: match (s:Kind1) return s
with n0 as (select n0.* from node n0 where n0.kind_ids operator (pg_catalog.&&) array []::int2[])
select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as s
from n0;
