# cypher to sql translation

Herein lies the heinous machinations of a stark raving madman, who through sheer courage and brute force created a program that eats the cypher query language and generates equivalent SQL, most of the time.

This package contains an implementation of a compiler that translates a string of **cypher** input to a string of **sql** output. The compiler does many things that are out of scope for this document, such as parsing an input string of **cypher** that adheres to the openCypher grammar, and producing an output string of **sql** that targets Bloodhound's homegrown postgres schema for graph databases, which can be referenced here: <TODO: insert a reference to the pg schema>. At the highest level, the compiler has 3 stages:

1. cypher string -> cypher AST
2. **cypher AST -> SQL AST**
3. SQL AST -> SQL string

This document will mainly describe **stage 2** 😀, which is the translation from the intermediate representation of cypher to the intermediate representation of SQL. The document begins with several commonly used examples of graph queries that every Bloodhound user can expect to encounter. After the quick tour of the most salient examples, it will explore the compiler's implementation in more depth.

## read this first

Now is a good time to point out where the AST definitions live:

The cypher AST nodes are defined in `cypher/model/model.go`.

The SQL AST nodes are defined in `cypher/model/pgsql/model.go`.

Some conventions that this document will use when referring to any AST node:

1. when referring to a cypher AST node, the type of the node will be prefixed with `cypher`. As such, the top-level root node of the cypher AST is called `cypher.RegularQuery`
2. likewise, when referring to a sql AST node, the type of the node will be prefixed with `pg`, for postgres. The top-level root node of the sql AST is called `pg.Statement`.

Using this new convention for talking about ASTs, we can express that stage 2 of the compiler is, roughly speaking, a mapping of `cypher.RegularQuery` structs to `pg.Statement` structs.

## query examples

### 1. fetch all nodes in the graph

Fetching all nodes is one of the more straightforward queries against a graph database. In cypher land, this is accomplished with the following cypher query:

```cypher
MATCH (s) RETURN s
```

the compiler translates the cypher above to the following sql:

```sql
WITH s AS(
	SELECT * FROM node
)
SELECT * FROM s;
```

What is shown here is only the input and output strings in the cypher and sql query languages respectively, but under the hood, the compiler produces two intermediate representations (IRs) before it outputs the sql.

The two specific IRs created during the translation of this example are a `cypher.SinglePartQuery` node first, followed by a `pg.Query` node second. The code for this specific translation can be found here and can help illuminate how specific language features from cypher get mapped onto their sql counterparts: <TODO: insert a link to some code or mention the function that does the translation> . Technically, `cypher.SinglePartQuery` and `pg.Query` are not the root-level nodes of either AST, but these two structures encapsulate the most relevant language features for this example, which is a "read all nodes" type of query.

More specifically, a `cypher.SinglePartQuery` node is a parent container for a list of `cypher.ReadingClause` nodes and a `cypher.Return` node. **"Reading clauses"** in cypher are language constructs that match specified patterns and criteria in a graph, e.g. typically everything before the **return** keyword of a query. So for the "fetch all nodes" query of this example, the reading clause is simply this part:

```
match (s)
```

and it gets nicely stuffed into the `cypher.ReadingClause` struct contained by the `cypher.SinglePartQuery` struct. Likewise, everything after the **return** keyword of the example query, in this case:

```
return s
```

gets nicely stuffed into the `cypher.Return` struct of the `cypher.SinglePartQuery` struct.

<todo: describe how the sql of this example fits into the pg.Query structure. at that point, i think the example is finished>

## todo

todo: im thinking maybe we just need a few fleshed out examples that illustrate what happens, and include some visualizations of the cypher ast and pg ast? screenshots of the cypher query explainer could also be helpful to explain why some decisions were made, ie. the sql ast extractor component is needed because we are trying to mimic the behavior seen in cypher's query expaliner for complex "where" clauses

combined with inline comments in the code on the translator functions.

todo: is it possible to have a 2 column table that contains the mapping from cypher AST nodes -> sql AST nodes? may not be possible if there isn't a 1:1 mapping

todo: there are other pieces of translation like rewriting the cypher that we may want to consider, i.e. moving criteria from the node pattern to a where clause, i.e.:
match (n {some: criteria}) ----> match (n) where n.some = 'criteria'

### todo: explain how "where" conditions get translated

### todo: explain how projections get translated (ie. return clauses)

### todo: explain how kind filters get translated (ie. match (a:kindA) )

### todo: explain how traversal is accomplished with recursive CTEs? especially why CTEs are needed over joins on the edge table
