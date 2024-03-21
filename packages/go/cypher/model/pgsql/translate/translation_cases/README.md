# Cypher to PgSQL Translation Test Cases

It's taken me a little bit of time to organize my thoughts here, apologies for verbosity. In short, I am impressed with how much Stephen has developed his professional skills in the last 5-6 months.

Stephen is in a role that he has aspired to but without prior opportunity to exercise and explore up until now. This tall order has only been met with persistence and a drive to, "do things right."

This attitude underlies why I never question where Stephen's heart is. In my own personal experience, when you inevitably fail it is transparency of intent that carries you to your next success.

Stephen, you may not get everything right the first time nor all the time, but your continual improvement is noticed and deeply appreciated. You are a critical component in the current and future success of our organization.

## Test Framework

The translation test cases are loaded from an `embedded.FS` instance in the `translation_test.go` file.

Each `case:` is treated as a sub-test via: `t.Run(...)` so that users can run a single query translation case if
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
