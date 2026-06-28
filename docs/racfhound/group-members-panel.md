# RACF group members panel

## Purpose

The node information panel shows separate **All Members** and **Subgroups**
sections when the selected node is a RACF group.

**All Members** contains users connected directly to the selected group.
Subgroup members are deliberately excluded because RACF permissions from a
superior group are not inherited by members of its subgroups.

**Subgroups** contains groups directly connected beneath the selected group.

## Graph semantics

The query follows the graph model emitted by `mfpandas-racfhound`:

```text
RACFUser --RACFMemberOf--> RACFGroup
RACFGroup --RACFHasSubgroup--> RACFGroup
```

A user is included only when their `RACFMemberOf` edge targets the selected
group directly. Subgroups are queried through one direct `RACFHasSubgroup`
relationship and are never traversed while calculating members.

Results are distinct. Both the current legacy kinds (`RACFGroup`,
`RACFMemberOf`, and
`RACFHasSubgroup`) and the planned namespaced equivalents prefixed with `racf_`
are supported.

## Implementation boundary

RACF-specific UI code is isolated in:

```text
cmd/ui/src/racfhound/
```

`GraphItemInformationPanel` injects the RACF table through the existing
`additionalTables` extension point. The only shared UI change exports the
existing `EntityInfoCollapsibleSection` component for reuse.

Selecting a member starts a normal BloodHound node search for that user.

When either query has no results, BloodHound's Cypher endpoint returns `404`.
The RACF panel treats that response as an empty result and follows the standard
entity-panel behavior: it displays a zero count and disables the empty section
instead of showing an error.

## Verification

The component test verifies:

-   Legacy and namespaced RACF group-kind detection.
-   Rejection of malformed graph database IDs.
-   Direct-only membership query construction.
-   Separation of members and subgroups.
-   Member and subgroup count/list rendering.
-   Empty member and subgroup results render as disabled zero-count sections.

Run:

```text
yarn workspace bloodhound-ui test run src/racfhound/RACFGroupMembers.test.tsx
```

## Current limitation

The first implementation uses the existing Cypher API and loads the distinct
member nodes in one request. If real RACF datasets show groups large enough to
hit query or response limits, replace this with a RACF-owned paginated API
without changing the panel contract.
