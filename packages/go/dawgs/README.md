# DAWGS (Database Abstraction Wrapper for Graph Systems)

DAWGS is the package that contains all domain specific implementation for interacting with the PostgresSQL database. It also maintains backwards compatibility with Neo4j.

## Cardinality

Cardinality refers to the number of relationships between entities (tables/objects). Some common cardinality types are One-to-One (1:1), One-to-Many (1:N), and Many-to-Many (M:N).

HyperLogLog (HLL) is a probabilistic algorithm used for cardinality estimation, meaning it estimates the number of distinct elements in a large dataset. 

Note: 32 bit vs 64 bit - Refers to the size of the hash output.
- Questions
    - when is one used vs the other?
    - what does the bitmap look like?

## Drivers

Initialization functions are responsible for starting and configuring the database drivers. These drivers are responsible for establishing the connection pool and managing the schema of the database.
- [Postgres](drivers/pg/pg.go)
- [Neo4j](drivers/neo4j/neo4j.go)

## Graph

## GraphCache

## Ops

## Query

## Traversal

## Util

## Vendor Mocks

Notes - 
Cypher is Neo4j's declarative query language.
