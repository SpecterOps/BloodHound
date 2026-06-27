# RACF user relationships panel

## Purpose

The RACF user information panel groups direct control and authority
relationships by graph direction while retaining RACF-specific child labels.

```text
Outbound Relationships
  Can Submit As
  Class Authorities

Inbound Relationships
  Can Be Submitted As By
```

## SURROGAT semantics

`mfpandas-racfhound` emits a direct SURROGAT relationship as:

```text
source principal --RACFSurrogateFor--> target user
```

The source principal has READ access to the target's `USERID.SUBMIT` profile.
Accordingly:

-   **Can Submit As** follows outgoing `RACFSurrogateFor` relationships.
-   **Can Be Submitted As By** follows incoming `RACFSurrogateFor`
    relationships.

An incoming principal can be a RACF user or group. The UI preserves the node
kind returned by BloodHound so both render and navigate correctly.

These sections show direct relationships only. Transitive SURROGAT chains
belong in path analysis rather than the entity relationship list.

## CLAUTH semantics

`mfpandas-racfhound` emits class authority as:

```text
RACFUser --RACFClassAuth--> RACFClass
```

**Class Authorities** follows this outgoing relationship and lists the classes
for which the user has CLAUTH.

## Compatibility and empty results

Queries support both current legacy edge kinds and planned `racf_`-namespaced
edge kinds. No-result `404` responses from the Cypher endpoint are treated as
empty lists. Empty parent and child sections use the standard disabled
zero-count entity-panel behavior.

## Implementation boundary

RACF-specific relationship queries and components are under:

```text
cmd/ui/src/racfhound/
```

The parent relationship sections aggregate child counts and reuse cached query
results when expanded.

## Verification

The tests cover:

-   Outgoing SURROGAT direction and rendering.
-   Incoming SURROGAT direction with a group principal.
-   Outgoing CLAUTH direction and rendering.
-   Direct-only, non-recursive query construction.
-   Empty relationship groups.
-   Rejection of malformed graph database IDs.

Run:

```text
yarn workspace bloodhound-ui test run src/racfhound/RACFUserRelationships.test.tsx
```
