# RACF user groups panel

## Purpose

The node information panel shows a **Groups** section when the selected node is
a RACF user. It lists groups connected directly to that user.

## Graph semantics

The query follows outgoing membership relationships:

```text
RACFUser --RACFMemberOf--> RACFGroup
```

Only one direct relationship is followed. The panel does not recurse through
superior or subordinate groups and does not imply permission inheritance.
Results are distinct.

Both the current legacy kinds (`RACFUser`, `RACFGroup`, and `RACFMemberOf`) and
the planned namespaced equivalents prefixed with `racf_` are supported.

## Behavior

-   The section count is the number of directly connected groups.
-   Clicking a group starts a normal BloodHound node search for that group.
-   No-result Cypher responses render as a disabled zero-count section rather
    than an error.

## Implementation boundary

The implementation reuses the RACF-related-node component under:

```text
cmd/ui/src/racfhound/
```

`GraphItemInformationPanel` injects the table through the existing
`additionalTables` extension point.

## Verification

The component tests cover:

-   Legacy and namespaced RACF user-kind detection.
-   Direct-only group query construction.
-   Group count and list rendering.
-   Empty group results.
-   Rejection of malformed graph database IDs.

Run:

```text
yarn workspace bloodhound-ui test run src/racfhound/RACFGroupMembers.test.tsx
```
