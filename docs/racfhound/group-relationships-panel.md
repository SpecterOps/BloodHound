# RACF group relationships panel

## Purpose

RACF groups can be principals on SURROGAT access lists. The group information
panel therefore includes:

```text
Outbound Relationships
  Can Submit As
```

## Graph semantics

`mfpandas-racfhound` preserves the access-list principal type when emitting:

```text
RACFGroup --RACFSurrogateFor--> target user
```

**Can Submit As** follows this direct outgoing relationship. It lists the user
identities that members acting through the group's authority may submit work
as.

The relationship list is direct-only. Transitive SURROGAT chains belong in path
analysis. An inbound section is not added to RACF groups because a SURROGAT
profile targets a user identity.

## Compatibility and empty results

The query uses `RACFSurrogateFor`. Empty results render as a disabled zero-count
parent section.

## Verification

Tests cover query direction, non-recursive behavior, target-user rendering,
empty results, and malformed database IDs.

Run:

```text
yarn workspace bloodhound-ui test run src/racfhound/RACFGroupRelationships.test.tsx
```
