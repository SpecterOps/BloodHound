# OpenGraph Extension Upload — Architecture

## Overview

[`UpsertOpenGraphExtension`](upsert_schema_extension.go) is the database-layer entry point for the schema
extension upload API. It performs a **true idempotent upsert**, reconciling every entity in the
uploaded payload against the current database state within a single transaction:

- **Matched rows** keep their IDs — a row is updated in place when its match key matches an entry
  in the payload; it is never deleted and recreated.
- **New rows** (in the payload but not the database) are created.
- **Stale rows** (in the database but not the payload) are deleted.

The HTTP handler and service layer are not involved in reconciliation. They perform validation and
call `UpsertOpenGraphExtension`, which returns `(bool, error)` — `true` if the extension already
existed, `false` if it was newly created.

---

## Why This Design

A single upload describes a complete schema extension: its node kinds, relationship kinds, their
kind info, environments, principal kinds, findings, and remediations. Those live in several
different tables — some introduced for OpenGraph, some pre-existing and shared with other features.
`UpsertOpenGraphExtension` deliberately absorbs the responsibility of managing all of them behind
one endpoint. That concentration is intentional, and it buys a few properties that are hard to get
any other way:

- **One transaction, one consistent outcome.** Because every table is written inside the same
  transaction, a partially-applied schema is impossible. Either the whole extension — across all of
  its tables — lands, or nothing does and the previous state is untouched. Splitting the work across
  per-table endpoints would reintroduce the interleaving and partial-failure problems this design
  exists to avoid.
- **The payload is the source of truth.** Clients upload the full desired state and let the server
  diff it, rather than issuing their own create/update/delete calls per table. Re-uploading the same
  document is a no-op, and the client never has to know which rows already exist or in what order
  the tables must be touched.
- **Ordering and foreign keys are handled centrally.** Cross-table dependencies (e.g. findings
  referencing environments, kind info hanging off a kind) require a specific reconciliation order.
  Owning every table in one place lets the orchestrator enforce that order and resolve foreign keys
  itself, instead of pushing that coupling onto callers.

The tradeoff is that this one endpoint accumulates broad responsibility over many new and
pre-existing tables, so changes here have wide reach. The generic [`reconcile`](reconcile.go)
primitive and the per-entity `reconcileConfig` factories exist to keep that surface manageable:
each table's specifics stay isolated in its own config, and the orchestrator stays a thin,
ordered chain. The [Adding a New Entity](#adding-a-new-entity-to-the-reconciliation-pipeline)
guide is the contract for extending this endpoint without eroding those guarantees.

---

## The `reconcile` Algorithm

[`reconcile`](reconcile.go) is a generic set-differencing function defined in `reconcile.go`. It is the
core primitive used for every entity type in the upload pipeline.

```go
func reconcile[TInput any, TExisting any, K comparable](
    ctx          context.Context,
    inputs       []TInput,
    existingRows []TExisting,
    config       reconcileConfig[TInput, TExisting, K],
) ([]TExisting, error)
```

**The caller is responsible for fetching `existingRows` upfront.** `reconcile` performs no
database reads of its own.

Given the two slices, the algorithm:

1. Builds `existingByKey map[K]TExisting` using `config.getExistingKey`
2. Builds `inputKeys map[K]bool` using `config.getInputKey`
3. **Delete pass** — for each item in `existingRows` whose key is absent from `inputKeys`, calls `config.delete`
4. **Create/Update pass** — for each input, calls `config.update` if its key is in `existingByKey`, otherwise `config.create`
5. Returns the full slice of created and updated items

Any callback error causes an immediate return.

### [`reconcileConfig`](reconcile.go)

```go
type reconcileConfig[TInput any, TExisting any, K comparable] struct {
    getInputKey    func(input TInput) K
    getExistingKey func(existing TExisting) K
    create         func(ctx context.Context, input TInput) (TExisting, error)
    update         func(ctx context.Context, existing TExisting, input TInput) (TExisting, error)
    delete         func(ctx context.Context, existing TExisting) error
}
```

Each config is constructed by a factory method on `BloodhoundDB`. Every factory accepts its
parent identifier(s) as parameters and closes over them in the `create` callback, so `reconcile`
itself has no knowledge of the parent scope. The identifiers each nested entity requires are:

- **Node kinds** — `extensionId`
- **Relationship kinds** — `extensionId`
- **Kind info** — the parent `kindID` plus exactly one of `nodeKindID` / `relationshipKindID`
  (the other is nil)
- **Environments** — `extensionId`
- **Principal kinds** — `environmentId`
- **Findings** — `extensionId`

Model types require no interface methods — key extraction is entirely handled by the
`getInputKey` / `getExistingKey` closures.

---

## Orchestrator — [`UpsertOpenGraphExtension`](upsert_schema_extension.go)

The orchestrator opens a transaction and processes each entity type in sequence using a single
`if/else if` chain. Each step fetches existing rows for that entity type, then calls `reconcile`
with the corresponding config factory.

```text
BEGIN TRANSACTION
│
├─ findOrCreateExtension
│     Matches on extension name (UNIQUE). Updates metadata if found, inserts if not.
│     Returns (extension, schemaExists, err). Built-in extensions are rejected here.
│
├─ fetch + reconcile node kinds       → upsertCustomIcons(reconciledNodeKinds)
│     Each create/update callback internally reconciles kind info entries.
│
├─ fetch + reconcile relationship kinds
│     Each create/update callback internally reconciles kind info entries.
│
├─ fetch + reconcile environments
│     Each create/update callback internally reconciles principal kinds.
│
├─ fetch + reconcile findings
│     Each create/update callback internally creates/updates the paired remediation.
│
COMMIT (or ROLLBACK on any error)
```

Environments must be reconciled before findings because the finding callbacks resolve
`environment_kind_name` to an environment row that must already exist.

---

## Entity Configurations

### Extension — [`findOrCreateExtension`](upsert_schema_extension.go)

Matched on the `name` column (`UNIQUE` constraint on `schema_extensions`). Delegates to one of
two helpers:

- [`updateExistingExtension`](upsert_schema_extension.go) — copies `display_name`, `version`, and `namespace` onto the
  existing row and calls `UpdateGraphSchemaExtension`, preserving the row's `ID`. Returns `schemaExists=true`.
- [`createNewExtension`](upsert_schema_extension.go) — calls `CreateGraphSchemaExtension` for the new row. Returns `schemaExists=false`.

The `schemaExists` bool drives the HTTP 200/201 response.

### Node Kinds

| Field | Value |
|---|---|
| Match key | `name` (string) |
| Fetch | [`GetGraphSchemaNodeKindsByExtensionId`](graphschema.go) |
| Create | [`CreateGraphSchemaNodeKind`](graphschema.go) |
| Update | Copies `DisplayName`, `Description`, `IsDisplayKind`, `Icon`, `IconColor` → [`UpdateGraphSchemaNodeKind`](graphschema.go) |
| Delete | [`DeleteGraphSchemaNodeKind`](graphschema.go) |

Both the create and update callbacks recursively call `reconcile` for the kind's info entries via
[`kindInfoReconcileConfig`](upsert_schema_extension.go) — see [Kind Info](#kind-info) below. On create, `existingRows` is an empty
slice; on update, existing entries are fetched via [`GetKindInfos`](graphschema.go) first. The config is
passed `nodeKindID` (with `relationshipKindID` nil).

The reconciled node kind list is passed to [`upsertCustomIcons`](upsert_schema_extension.go), which creates or updates
`custom_node_kinds` rows for any node kind marked `IsDisplayKind`. Icons and colors that are
absent from the input are preserved from the existing row.

### Relationship Kinds

| Field | Value |
|---|---|
| Match key | `name` (string) |
| Fetch | [`GetGraphSchemaRelationshipKindsByExtensionId`](graphschema.go) |
| Create | [`CreateGraphSchemaRelationshipKind`](graphschema.go) |
| Update | Copies `Description`, `IsTraversable` → [`UpdateGraphSchemaRelationshipKind`](graphschema.go) |
| Delete | [`DeleteGraphSchemaRelationshipKind`](graphschema.go) |

Both the create and update callbacks recursively call `reconcile` for the kind's info entries via
[`kindInfoReconcileConfig`](upsert_schema_extension.go) — see [Kind Info](#kind-info) below. On create, `existingRows` is an empty
slice; on update, existing entries are fetched via [`GetKindInfos`](graphschema.go) first. The config is
passed `relationshipKindID` (with `nodeKindID` nil).

### Kind Info

Kind info entries (`schema_kind_info`) are the entity-panel definitions attached to a node or
relationship kind. They are reconciled **inside** the node-kind and relationship-kind callbacks —
`reconcile` is called recursively with [`kindInfoReconcileConfig`](upsert_schema_extension.go). The same config factory
serves both parents; the caller passes exactly one of `nodeKindID` / `relationshipKindID` (the
other is nil), both closed over by the `create` callback alongside the parent `kindID`.

| Field | Value |
|---|---|
| Match key | `InfoKey` (string) |
| Fetch | [`GetKindInfos`](graphschema.go) (by `KindID`); empty slice on parent create |
| Create | [`CreateKindInfo`](graphschema.go) |
| Update | Copies `Title`, `Position`, `Content` → [`UpdateKindInfo`](graphschema.go) |
| Delete | [`DeleteKindInfo`](graphschema.go) |

When an entry's `Content` is empty, [`CreateKindInfo`](graphschema.go) / [`UpdateKindInfo`](graphschema.go) substitute
`defaultKindInfoContent` (`{"markdown":{"content":""}}`). Constraint violations are mapped by
[`checkKindInfoError`](graphschema.go) to `ErrKindInfoKindNotFound`, `ErrKindInfoDuplicatePosition`, or
`ErrKindInfoDuplicateInfoKey`.

### Environments

| Field | Value |
|---|---|
| Match key | `EnvironmentKindName` (string) — populated via JOIN in [`GetEnvironmentsByExtensionId`](graphschema.go) |
| Fetch | [`GetEnvironmentsByExtensionId`](graphschema.go) |
| Create callback | [`CreateEnvironmentWithPrincipalKinds`](upsert_schema_environment.go) |
| Update callback | [`UpdateEnvironmentWithPrincipalKinds`](upsert_schema_environment.go) |
| Delete | [`DeleteEnvironment`](graphschema.go) (CASCADE deletes principal kinds) |

The `create` and `update` callbacks each internally resolve FK names and then call `reconcile`
again for the environment's principal kinds — see [Principal Kinds](#principal-kinds) below.

### Principal Kinds

Principal kinds are reconciled inside the environment callbacks — `reconcile` is called
recursively with [`principalKindReconcileConfig`](upsert_schema_environment.go). The key type is `int32` (the kind ID), not a
string name, because principal kind rows have no mutable fields — a match is an identity match.
The `update` callback is a no-op.

[`CreateEnvironmentWithPrincipalKinds`](upsert_schema_environment.go) passes `nil` as `existingRows` because the environment is
newly created and has no existing principal kinds. [`UpdateEnvironmentWithPrincipalKinds`](upsert_schema_environment.go) fetches
existing principal kinds via [`GetPrincipalKindsByEnvironmentId`](graphschema.go) before calling `reconcile`.

### Findings

| Field | Value |
|---|---|
| Match key | `name` (string) |
| Fetch | [`GetSchemaFindingsByExtensionId`](graphschema.go) |
| Create callback | [`CreateFindingWithRemediation`](upsert_schema_finding.go) |
| Update callback | [`UpdateFindingWithRemediation`](upsert_schema_finding.go) |
| Delete | [`DeleteSchemaFinding`](graphschema.go) (CASCADE deletes the paired remediation) |

### Remediations

Remediations are managed directly inside the finding callbacks — they are not reconciled through
`reconcile`. Each finding has exactly one remediation keyed by `finding_id`.

- [`CreateFindingWithRemediation`](upsert_schema_finding.go) calls [`CreateRemediation`](upsert_schema_remediation.go) after creating the finding row.
- [`UpdateFindingWithRemediation`](upsert_schema_finding.go) calls [`UpdateRemediation`](upsert_schema_remediation.go) after updating the finding row.

Because finding IDs are stable across uploads (matched by name, not deleted and recreated), the
remediation row follows its finding's lifecycle: created when the finding is new, updated when
the finding is updated, and cascade-deleted when the finding is deleted.

### FK Resolution — Findings

[`resolveFindingFKs`](upsert_schema_finding.go) translates the string names in a `RelationshipFindingInput` to their
corresponding database IDs before the finding row is written:

- `RelationshipKindName` → [`GetKindByName`](kind.go) → `kind_id`
- `EnvironmentKindName` → [`GetEnvironmentKindName`](graphschema.go) → `environment_id`

`GetEnvironmentKindName` filters the environments table directly by kind name (via the JOIN
alias `k.name`). Kind names carry a unique constraint, so no intermediate kind lookup is needed.

[`applyFindingInput`](upsert_schema_finding.go) applies the resolved IDs and mutable fields onto an existing `SchemaFinding`
struct, returning it ready for [`UpdateSchemaFinding`](graphschema.go).

---

## Adding a New Entity to the Reconciliation Pipeline

New entity types plug into the pipeline by supplying a `reconcileConfig` and a single step in the
orchestrator chain. `reconcile` itself never changes — it is fully generic over the input type,
the row type, and the match key.

### 1. Define the model types

Add the desired-state input type (`TInput`) and the current-state row type (`TExisting`) to the
`model` package, and add the input slice to [`GraphExtensionInput`](../model/graphschema.go) so it is parsed from the
uploaded payload.

### 2. Add the database methods

On `BloodhoundDB`, add the four operations the config will delegate to:

- a **fetch** scoped to the parent (e.g. `GetMyEntitiesByExtensionId`) — the orchestrator calls
  this to supply `existingRows`;
- **create**, **update**, and **delete** methods for a single row.

Add each new method to the `Database` interface in [`db.go`](db.go) and regenerate the mock
(`MockDatabase` in `mocks/db.go`) via `just prepare-for-codereview`.

**Schema migration — cascade constraints are required.** In the same migration that creates the
new table, every foreign key pointing at a parent that the pipeline can delete (the parent's own
row, or any 1:1 dependent row from step 6) MUST declare `ON DELETE CASCADE`. Reconciliation
deletes stale parents by ID and relies on the database to remove their dependents; if the
constraint is missing, deleting a stale parent leaves the paired row orphaned. If you are attaching
to an existing table, verify the constraint already exists and add a migration to introduce it if
it does not.

### 3. Write the config factory

Add a factory method on `BloodhoundDB` that returns
`reconcileConfig[TInput, TExisting, K]`, following the existing factories in
[`upsert_schema_extension.go`](upsert_schema_extension.go):

- `getInputKey` / `getExistingKey` extract the match key `K`;
- `create` closes over the parent scope (e.g. `extensionId`);
- `update` copies mutable fields onto the existing row (preserving its `ID`) before writing;
- `delete` removes the row by `ID`.

+**Choosing the key `K`:** use the stable unique identity carried by both the input and existing row. Use `name` only when the entity defines `name` as that identity. Use another stable external key when available. Use a database ID only when the uploaded input carries that same ID. If no stable identity exists, define one before adding the entity.

### 4. Add one step to the orchestrator chain

Insert a `fetch + reconcile` pair into the `if/else if` chain in
[`UpsertOpenGraphExtension`](upsert_schema_extension.go):

```go
} else if existingMine, err := bloodhoundDBTransaction.GetMyEntitiesByExtensionId(ctx, extension.ID); err != nil {
    return schemaExists, fmt.Errorf("failed to fetch existing my entities: %w", err)
} else if _, err := reconcile(ctx, graphExtensionInput.MyEntitiesInput, existingMine, bloodhoundDBTransaction.myEntityReconcileConfig(extension.ID)); err != nil {
    return schemaExists, fmt.Errorf("failed to reconcile my entities: %w", err)
```

**Ordering matters.** Place the step *before* any entity whose callbacks resolve a foreign-key
name back to your rows — this is why environments are reconciled before findings (see
[Orchestrator](#orchestrator--upsertopengraphextension)). All steps share the same transaction,
so any error rolls back the whole upload.

### 5. Nested child entities

If the new entity owns child rows, reconcile them **recursively inside** its `create` and `update`
callbacks rather than adding another top-level step — mirror [Kind Info](#kind-info) and
[Principal Kinds](#principal-kinds):

- on **create**, pass an empty `existingRows` slice (`[]TChild{}`);
- on **update**, fetch the current children first, then call `reconcile`;
- have the child config factory close over the just-created/updated parent ID.

### 6. Non-reconciled paired rows

If the new entity owns exactly one dependent row keyed by its ID (a 1:1 relationship), manage it
directly inside the callbacks instead of through `reconcile` — mirror
[Remediations](#remediations). The dependent row follows its parent's lifecycle via
`ON DELETE CASCADE`, so its foreign key **must** declare that constraint in the schema migration
(see step 2). Without it, deleting a stale parent orphans the dependent row.

### 7. Foreign-key resolution

If the input references other entities by name, resolve those names to IDs inside the callback
before writing the row — mirror [`resolveFindingFKs`](upsert_schema_finding.go). The referenced entity must be reconciled
earlier in the chain (see step 4).

### 8. Tests

Add unit tests for the new callbacks and an integration test that exercises the full
create → update → delete diff for the entity through `UpsertOpenGraphExtension`, asserting that
row IDs are preserved across a re-upload and that stale rows are deleted. If the entity owns child
or 1:1 dependent rows, the integration test MUST also assert that deleting a stale parent removes
its dependents — verifying the `ON DELETE CASCADE` constraint from step 2 and guarding against
orphaned rows.

---

## Finding Subtypes

Subtypes (`schema_findings_subtypes`) are **not** written by the upload API. They are managed
exclusively by the CUE-generated `findings.sql` migration for built-in extensions. The migration
truncates the table and re-inserts subtype rows on every run.

The read path is fully implemented: [`GetSchemaFindings`](graphschema.go) does a `LEFT JOIN` and aggregates
subtypes into `SchemaFinding.Subtypes []string`. [`CreateSchemaFindingSubtype`](graphschema.go) exists on the
`Database` interface but has no call sites in production code.
