# Cypher to PgSQL Translation Test Cases

## Test Framework

The translation test cases are loaded from an `embedded.FS` instance in the `translator_test.go` file.

Each `case:` is treated as a subtest via: `t.Run(...)` so that users can run a single query translation case if
desired.

## Format

A test case is presented to the test framework as a SQL query that is preceded by a comment with the `case:` tag. The
end of the query is determined by matching a closing ';' character.

```postgresql
-- This is a regular comment and will be ignored.
-- case: match (n) return n
with n1 as (select n1.* from node n1)
select *
from n1;

-- Matching for the `case:` tag is case-insensitive.
-- CASE: match (n) return n
with n1 as (select n1.* from node n1)
select * -- Comments may follow query content.
-- Comments may also interleave with queries.
from n1;
```

### Parameter Matching

Test cases may specify parameter matchers for further assertion:

```postgresql
-- case: match (s) where s.name = $myParam return s
-- cypher_params: {"myParam": "123"}
-- pgsql_params: {"pi0": "123"}
with s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0
            from node n0
            where n0.properties ->> 'name' = @pi0::text)
select s0.n0 as s
from s0;
```
