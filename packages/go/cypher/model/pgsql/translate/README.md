# Translation

# Graph Query Intermediate Representation

## Intent

The GQIR is a simplification of the openCypher query model. The model is designed to allow for a single-walk translation
of a given openCypher query. The GQIR is used to normalize and reorganize query components to serialize them for
encoding into different end-query models.

### Pattern Rewriting

The example below demonstrates reading out label constraints and reorganizing them to simplify translation to PgSQL:

```text
--- Query starts as
match (n:Label) where n:Label2 return n;

--- Reordered when translated to GQIR
match (n1) where n1:Label1 and n1:Label2 return n1 as n;
```

### Identifier Rebinding

To simplify translation, all implicit pattern operands are given explicit identifiers. Existing identifiers are aliased
with new identifiers to avoid collisions.

```text
--- Query starts as
match (s:Label) return s;

--- Identifiers rewritten when translated to GQIR
match (n1) where n1:Label return n1
```

# Stages

## Regular Query

## Single-Part Query

### Pattern Translation

### Match Translation

### Where Translation

### Projection Translation

## Multi-Part Query
