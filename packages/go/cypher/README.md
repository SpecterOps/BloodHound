# BloodHound Cypher Parser and Query Model

This project contains a golang parser implementation that for the [openCypher](https://opencypher.org/) project. The
primary goal of this implementation is to be lightweight and compatible with cypher dialect used by
[Neo4j](https://neo4j.com/).

Several features may not be fully supported while this implementation continues to mature. Eventual support for most
legacy openCypher features is planned.

* [Generating the ANTLR Grammar](#generating-the-antlr-grammar)
* [Regenerating the ANTLR Parser Implementation](#regenerating-the-antlr-parser-implementation)
* [Structure](#structure)
    * [Packages](#packages)
        * [analyzer](#analyzer)
        * [frontend](#frontend)
        * [grammar](#grammar)
        * [model](#model)
        * [parser](#parser)
        * [test](#test)
* [Language Features](#language-features)
    * [Query Cost Model](#query-cost-model)
        * [Caveats](#caveats)
        * [Cost Metrics](#cost-metrics)
    * [Filtered Model Features](#filtered-model-features)
    * [Unsupported Model Features](#unsupported-model-features)

## Generating the ANTLR Grammar

openCypher Repository Version Hash: `46223b9e8215814af333b7142850fea99a3949a7`

The [openCypher repository](https://github.com/opencypher/openCypher/tree/46223b9e8215814af333b7142850fea99a3949a7/grammar)
contains instructions on how to generate output artifacts from the grammar XML sources. Below is an example on how to
generate the ANTLR grammar from the checked out version hash:

```bash
./tools/grammar/src/main/shell/launch.sh Antlr4 --INCLUDE_LEGACY=true cypher.xml > grammar/generated/Cypher.g4
```

## Regenerating the ANTLR Parser Implementation

[Download the latest version of ANTLR](https://www.antlr.org/download.html) first and ensure that you have `java`
installed. The command below can be used to generate a new version of the openCypher parser backend.

```bash
java -jar ./antlr-4.13.0-complete.jar -Dlanguage=Go -o parser grammar/Cypher.g4
```

## Structure

### Packages

#### analyzer

This package contains query analysis and rewriting tools built for the openCypher query model.

#### frontend

This package contains the frontend implementation of the BloodHound cypher parser. This includes the AST visitor
implementations as well as formatting utilities for the query model.

#### grammar

This folder is not a golang package. Instead, it exists to contain the openCypher grammar that was used to generate the
`parser` package.

#### model

This package contains the query model that the openCypher grammar AST is first translated to. This model is not
intended to be a 1:1 representation of the openCypher grammar. It attempts to walk a fine line between accuracy of the
formal language model of openCypher and developer usability.

#### parser

This package is generated from the ANTLR grammar in the `grammar` folder.

#### test

This package contains testing tools for the parser along with a corpus of test cases for validating parser
functionality.

## Language Features

### Query Cost Model

The `analyzer` package implements a naive cost-based query complexity measure that uses simple heuristics to help users
determine if a query may result in poor database performance due to resource cost. The analyzer walks the openCypher
logical model and not the syntax model, making writing additional analysis components for query complexity easier.

#### Caveats

The cost model is built without insight into the source graph's statistics. Without a measure of connectivity or other
aspects, the cost model is based purely on query planning dynamics and therefore may inaccurately predict the cost of
certain operations. The cost model is intended to be used as a best-effort estimation along with stringent transactional
time-boxing to preserve runtime stability and performance of the database serving the graph.

The analyzer currently does not track variable aliasing and may produce incorrect cost metrics for queries that use
inline projections (e.g. `with a as b`). Support for variable aliasing is a planned feature.

#### Cost Metrics

The analyzer currently supports the following heuristics for cost modeling:

* [Tracking Indexed Lookups for Node Labels](#tracking-indexed-lookups-for-node-labels)
* [Function and Function-Like Expressions](#function-and-function-like-expressions)
* [Quantifier and Filter Expressions](#quantifier-and-filter-expressions)
* [Sorting](#sorting)
* [Rendering Multiple Projections](#rendering-multiple-projections)
* [Comparison Operators](#comparison-operators)
* [Node Patterns](#node-patterns)
* [Relationship Patterns](#relationship-patterns)

##### Tracking Indexed Lookups for Node Labels

The analyzer will add weight to a query's cost if a node lookup is not bound to a label. Additionally, the analyzer
supports narrowing of a node's label later in the query and will adjust query cost accordingly.

* Node Lookup without Label

```cypher
// Higher cost
match (n) return n;
```

* Node Lookup with Label

```cypher
// Lower cost
match (n:Label1) return n;
```

* Node Lookup with Label in Where Clause

```cypher
// Lower cost
match (n) where n:Label1 or n:Label2 return n;
```

##### Function and Function-Like Expressions

The analyzer will add weight to a query's cost based on use of function or function-like invocations.

* Using Collect

```cypher
// Higher cost
match (p:Person)-[:ACTED_IN]->(m:Movie) return p.name, collect(m.title)
```

* Inspecting a Relationship Label

```cypher
// Higher cost
match (p:Person)-[r]->() where not type(r) in ["RLabel1", "RLabel2"] return p
```

##### Quantifier and Filter Expressions

The analyzer will add additional weight to a query's cost if a quantifier expression is used.

* Filtering All Nodes in a Matched Pattern

```cypher
// Higher cost
match p = (:Person)-[:WorksFor]->(:Company) where none(n in nodes(p) where n.excluded) return p
```

##### Sorting

The analyzer will add additional weight to queries that sort projections.

* Ordering Results by Multiple Dimensions

```cypher
// Higher cost
match (p:Person) return p order by p.year asc, p.name desc
```

##### Rendering Multiple Projections

The analyzer will add compounding weight to queries that utilize more than one projection, inline or otherwise.

* Using `with` to Join Distinct Patterns

```cypher
// Higher cost
match (p:Person)-[:WorksFor]->(:Company {name: "thebigone.com"}) with p as s match p = (s)-[:WorksFor]->(:Company {name: "theotherone.net"}) return p
```

##### Comparison Operators

Regular expression comparisons may have unknown performance characteristics and are weighted higher in the cost model.

* Using Regular Expression Matching with a Backreference

```cypher
// Higher cost
match (p:Person) where p.favorite_quote =~ "(['\"])(.*?)\\1" return p
```

##### Node Patterns

The analyzer will add a compounding weight for each node pattern matched on.

* Single Node Lookup

```cypher
// Low cost
match (p:Person) where p.name = "Jim" return p
```

* Multiple Node Lookup

```cypher
// Higher cost
match (p1:Person), (p2:Person) where p1.name = "Jim" and p2.name = "Evil Jim" return p1, p2
```

* Multiple Node Lookup Without Labels

```cypher
// Highest cost
match (p1), (p2) where p1.name = "Jim" and p2.name = "Evil Jim" return p1, p2
```

##### Relationship Patterns

The analyzer will cost relationship patterns in multiple ways to attempt to account for path direction and expansion.

* Single Depth Pattern Lookup

```cypher
// Low cost
match p1 = (p2:Person)-[:WorksFor]->(:Company) where p2.name = "Jim" return p1
```

* Bidirectional Single Depth Pattern Lookup

```cypher
// Higher cost
match p1 = (p2:Person)<-[:WorksFor]->(:Company) where p2.name = "Jim" return p1
```

* Using Expansion with a Range Literal

```cypher
// Higher cost
match p1 = (p2:Person)-[:Knows *1]->(:Person) where p2.name = "Jim" return p1
```

* Using Expansion with a Greedy Range Literal

```cypher
// Higher cost
match p1 = (p2:Person)-[:Knows *1..]->(:Person) where p2.name = "Jim" return p1
```

* Using Chained Pattern Expansion with a Greedy Range Literal

```cypher
// Highest cost
match p1 = (:Company)<-[:WorksFor]-(p2:Person)-[:Knows *1..]->(:Person)-[:WorksFor]->(:Company) where p2.name = "Jim" return p1
```

* Shortest Path Lookup

```cypher
// Low cost
match p1 = shortestPath((p2:Person)-[:Knows *1..]->(p3:Person)) where p2.name = "Jim" and p3.name = "Evil Jim" return p1
```

* All Shortest Paths Lookup

```cypher
// Higher cost
match p1 = allShortestPaths((p2:Person)-[:Knows *1..]->(p3:Person)) where p2.name = "Jim" and p3.name = "Evil Jim" return p1
```

### Filtered Model Features

The `frontend` package contains a helpful AST filter for sanitizing potentially dangerous openCypher language
constructs when dealing with user provided input.

The following constructs are filtered by default:

* [Procedure Calls](#procedure-calls)
* [Merge Clauses](#merge-clauses)
* [Delete Clauses](#delete-clauses)
* [User Supplied Parameters](#user-supplied-parameters)
* [Create Clauses](#create-clauses)
* [Property Mutation](#property-mutation)

#### Procedure Calls

```cypher
CALL com.package.path.functionName() YIELD ResultField as A;
```

#### Merge Clauses

```cypher
MERGE (StartLabel {Name: 'A'})-[E:EdgeLabel]->(EndLabel {Target: true}) RETURN E;
```

#### Delete Clauses

```cypher
MATCH (V:Label) DELETE V;
MATCH (V:Label) DETACH DELETE V;

MATCH ()-[E:EdgeLabel]->() DELETE E;
```

#### User Supplied Parameters

```cypher
CREATE (V:Label $MyParameter) RETURN V;
```

#### Create Clauses

```cypher
CREATE (V:Label {Name: 'my name'});
```

#### Property Mutation

```cypher
MATCH (V:Label) SET V.Name = 'new name';
MATCH (V:Label) REMOVE V.Name;

MATCH ()-[E:EdgeLabel]->() SET E.Name = 'new name';
MATCH ()-[E:EdgeLabel]->() REMOVE E.Name;
```

### Unsupported Model Features

* [Index Commands](#index-commands)
* [Unique Constraint Commands](#unique-constraint-commands)
* [Node Property Existence Constraints](#node-property-existence-constraints)
* [Relationship Property Existence Constraints](#relationship-property-existence-constraints)
* [Load CSV Commands](#load-csv-commands)
* [For Each Clauses](#for-each-clauses)
* [Periodic Commit Hints](#periodic-commit-hints)
* [Union Clauses](#union-clauses)
* [Profile Clauses](#profile-clauses)
* [Explain Clauses](#explain-clauses)
* [Start Clauses](#start-clauses)
* [Reduce Clauses](#reduce-clauses)
* [Case Clauses](#case-clauses)
* [Existential Subquery](#existential-subquery)
* [Legacy List Expressions](#legacy-list-expressions)
* [Legacy Parameter Expansion](#legacy-parameter-expansion)

#### Index Commands

```cypher
CREATE INDEX ON :Label(Property);
DROP INDEX :Label(Property);
```

#### Unique Constraint Commands

```cypher
CREATE CONSTRAINT ON (V:Label) ASSERT V.Property IS UNIQUE;
DROP CONSTRAINT ON (V:Label) ASSERT V.Property IS UNIQUE;
```

#### Node Property Existence Constraints

```cypher
CREATE CONSTRAINT ON (V:Label) ASSERT EXISTS (V.Property);
DROP CONSTRAINT ON (V:Label) ASSERT EXISTS (V.Property);
```

#### Relationship Property Existence Constraints

```cypher
CREATE CONSTRAINT ON ()-[E:Label]->() ASSERT EXISTS (E.Must);
DROP CONSTRAINT ON ()-[E:Label]->() ASSERT EXISTS (E.Must);
```

#### Load CSV Commands

```cypher
LOAD CSV WITH HEADERS FROM 'file:///data.csv';
```

#### For Each Clauses

```cypher
MATCH P = (VS)-[*]->(VE)
WHERE VS.Name = 'A' AND VE.Name = 'D'
FOREACH (V IN nodes(P) | SET V.Marked = true)
```

#### Periodic Commit Hints

```cypher
USING PERIODIC COMMIT 5000 LOAD CSV WITH HEADERS FROM 'file:///data.csv';
```

#### Union Clauses

```cypher
MATCH (V1:LabelA)
RETURN V1.Name AS v1Name
UNION ALL
MATCH (V2:LabelB)
RETURN V2.Name AS v2Name
```

#### Profile Clauses

```cypher
PROFILE MATCH (V:Label) WHERE V.Name = 'indexed name' RETURN V;
```

#### Explain Clauses

```cypher
EXPLAIN MATCH (V:Label) WHERE V.Name = 'indexed name' RETURN V;
```

#### Start Clauses

```cypher
START V = node(*)
MATCH P = (T:Label)-[*]->(V)
WHERE T.Target = true
RETURN P
```

#### Reduce Clauses

```cypher
START V = node(*)
MATCH P = (T:Label)-[*]->(V)
WHERE T.Target = true
RETURN REDUCE(sum = '', X in P | sum + X.Weight)
```

#### Case Clauses

```cypher
MATCH (V)
RETURN CASE V.eyes
  WHEN 'blue'  THEN 1
  WHEN 'brown' THEN 2
  ELSE 3
END AS result
```

#### Existential Subquery

```cypher
MATCH (V:Label)
WHERE EXISTS {MATCH (V)-[E:EdgeLabel]->(T:TargetLabel) WHERE T.Property = 'target' return V}
RETURN V
```

#### Legacy List Expressions

```cypher
MATCH (VE:Excluded)    
WITH COLLECT(VE) AS excluded
MATCH (VI:Included) 
WITH excluded, COLLECT(VI) AS included
WITH FILTER(VI IN included WHERE NOT VI IN excluded) as resultList
UNWIND resultList as results
RETURN results.property
```

#### Legacy Parameter Expansion

```cypher
MATCH (V)
SET V.Name = {parameter}
```
