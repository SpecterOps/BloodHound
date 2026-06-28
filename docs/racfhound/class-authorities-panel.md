# RACF class authorities panel

## Purpose

The node information panel shows **Users With CLAUTH** when the selected node is
a RACF class. It provides the inverse view of the RACF user's **Class
Authorities** relationship.

## Graph semantics

The query follows incoming class-authority relationships:

```text
RACFUser --RACFClassAuth--> RACFClass
```

Only direct `RACFClassAuth` relationships are listed. Results are distinct.

## Behavior

-   The section count is the number of users with direct CLAUTH.
-   Clicking a user starts a normal BloodHound node search.
-   No-result responses render as a disabled zero-count section rather than an
    error.

## Verification

Tests cover class-kind detection, direct incoming query direction, user
rendering, empty results, and malformed database IDs.

Run:

```text
yarn workspace bloodhound-ui test run src/racfhound/RACFClassRelationships.test.tsx
```
