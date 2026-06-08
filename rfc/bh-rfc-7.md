---
bh-rfc: 7
title: Audit Logging Middleware
authors: |
    [Holms, Alyx](aholms@specterops.io)
status: DRAFT
created: 2026-06-08
audiences: |
    BloodHound Engineering, Product
---

# Audit Logging Middleware

## 1. Overview

This RFC proposes moving audit logging from a database-layer concern to an HTTP middleware concern. Today, audit coverage is opt-in and per-call-site; the middleware approach makes coverage the default behavior of the HTTP layer, closes existing gaps on read endpoints, and establishes a bounded-growth storage strategy through time-based table partitioning.

## 2. Motivation & Goals

Today, every audited operation must be explicitly wrapped in `BloodhoundDB.AuditableTransaction` in the database layer. This creates two problems: coverage is opt-in and developer-dependent, and a request-scoped concern (who called what, from where, with what result) lives in database code. There is also no retention strategy for `audit_logs`, which grows unbounded.

-   **Coverage by default** - Every registered route is audited automatically; new endpoints require no extra instrumentation.
-   **Separation of concerns** - Auditing moves out of the database layer into the HTTP layer where it belongs.
-   **Consistent attribution** - Actor, source IP, request ID, and timing derive from one place (`bhctx` + `IdentityResolver`) rather than being re-derived at each call site.
-   **Bounded table growth** - Monthly `created_at` partitioning with `DROP PARTITION` retention closes an existing operational gap regardless of this proposal's other details.

## 3. Considerations

### 3.1 Impact on Existing Systems

#### 3.1.1 Action Data Contract

The most significant impact of this proposal is a **breaking change to the `Action` field** of `audit_logs`. Today, `Action` is one of approximately 43 typed string constants (e.g., `"CreateUser"`, `"ExportAllRisks"`). The proposed default changes this to a derived string of the form `"METHOD /path/template"` (e.g., `"POST /api/v2/users"`).

Any consumer filtering by action name — the audit log UI, customer SIEM integrations, or analytics — will be affected. The migration strategy is:

-   **Existing instrumented endpoints:** During the migration window, endpoints that already carry a typed `AuditLogAction` constant continue to emit it via the optional enrichment hook (see [Section 5.2](#52-optional-semantic-overlay)), so their database value does not change until they are explicitly migrated off the constant.
-   **New endpoints:** Default to the route-template form from day one.
-   **Full retirement of typed constants:** Requires Product sign-off on consumer impact before the deprecation schedule is set. This RFC records the direction; the transition schedule is a pending Product conversation.

#### 3.1.2 Coexistence with Existing Audit Logging

The middleware is registered globally from day one and begins emitting the new route-template format for **every** endpoint immediately. Existing audit logging (`AuditableTransaction` and out-of-band `AppendAuditLog` calls) is **left in place** rather than removed up front, so the two systems coexist during the transition:

-   Endpoints that already emit a typed `AuditLogAction` continue to do so through their existing path, so their database value is unchanged.
-   The middleware additionally records the generic route-template row for those same endpoints. This temporary overlap is the deliberate cost of turning coverage on everywhere immediately.
-   When an endpoint is refactored into the new `server/` slice architecture, its semantic action (e.g. `"CreateUser"`) is **re-homed into the handler layer** and emitted through the injected audit slice (see [Section 6](#6-audit-log-slice)), and the old `AuditableTransaction` call is removed in the same change.

The legacy audit-logging code is removed wholesale only at the end of the transition (see [Section 3.5](#35-implementation-plan)).

#### 3.1.3 Out-of-Band Audit Writes

Several audit writes today occur outside the normal HTTP handler path and are **explicitly out of scope** for the middleware. These will continue to write directly via `AppendAuditLog`:

-   **Login events** (`LoginAttempt`) — written by the authentication layer.
-   **Support account operations** (`CreateSupportUserSessionAttempt`, `InvalidateSupportUserSession`, `InvalidateAllSupportUserSessions`) — written from `lib/go/api/tools/support_account`, which bypasses the HTTP stack.
-   **Daemon-triggered actions** (`CreateAuthTokens`, `UpdateAuthTokens`) — written by scheduled daemons, not request handlers.

### 3.2 Security & Compliance

#### 3.2.1 Read Auditing

The middleware writes the full `intent` → `success`/`failure` pair for **all** audited endpoints, reads included. Reads frequently expose sensitive data, so a record of what was attempted and whether that attempt succeeded is as valuable for a GET as for a mutation. A failed or denied read is a security-relevant event — the intent row preserves evidence of the attempt even if the handler errors, panics, or the connection drops before a result row is written.

#### 3.2.2 Sensitive Data Capture

Request bodies must **not** be blanket-logged. The middleware records method, route template, route parameters, and handler-contributed fields only, with a redaction denylist for known secret fields. This matches the existing behavior where `AuditData()` deliberately omits credentials and tokens.

#### 3.2.3 Anonymous Actor Path

Today, `newAuditLog` errors when there is no auth context, meaning failed logins are not audited at the database layer. The middleware will support an **anonymous actor** path that attributes an unauthenticated request to its source IP rather than dropping the audit record.

### 3.3 Drawbacks & Alternatives

The following alternatives were considered and rejected:

-   **Keep audit logging in the database layer and just add coverage.** Cannot guarantee completeness; relies on every author remembering to route through `AuditableTransaction`. Also does nothing for read endpoints.
-   **Pure middleware with no enrichment hook.** Simplest to build, but permanently loses the model-level `Fields` that make today's records useful for investigations.
-   **Asynchronous-only, fire-and-forget auditing for everything.** Lowest request latency, but weakens the durability of the `intent` guarantee and risks silently dropping audit records on crash. The `intent` write stays synchronous.
-   **Single-row (result-only) auditing for reads.** Halves read volume, but loses the pre-execution record of what was attempted, so denied or failed reads of sensitive data could go unrecorded.
-   **Mandatory route → `AuditLogAction` mapping.** A hand-written map from every registered route to a typed constant has the same opt-in coverage hole as today and adds maintenance burden.
-   **Named routes (`mux.Route.Name`) as the action identifier.** Requires one `.Name("…")` call per route at registration — same cardinality benefit as method+template but still per-route maintenance.
-   **Reflected Go symbol names.** Zero maintenance, but closures and wrapper types obscure the symbol name, coupling the audit data contract to internal Go naming. Renames become data-contract breaks.
-   **`DELETE`-based retention.** Simpler to migrate to, but under the write volume of auditing every endpoint it accumulates dead tuples that autovacuum must continually reclaim. Partitioning with `DROP PARTITION` removes the bloat risk by construction.

### 3.4 Quality

#### 3.4.1 Latency

The `intent` write is kept **synchronous** — its whole point is pre-execution durability. The `result` write is offloaded to a **buffered worker** so it is off the request critical path. The shared `commit_id` lets the result be reconciled to its intent asynchronously without ordering risk.

#### 3.4.2 Volume

Auditing every endpoint with a full intent+result pair significantly increases row count compared to today's mutation-only coverage. This is addressed through monthly partitioning with `DROP PARTITION` retention (see [Section 7](#7-retention-and-table-stability)) rather than by weakening the audit record. Reducing volume at the source — primarily UI polling behavior — is a recommended follow-up and out of scope for this proposal.

#### 3.4.3 HTTP Status as Outcome Signal

`< 400` maps to `success`; `>= 400` maps to `failure`. This is documented as "the API's response to the caller," not "the business outcome." A `200` can mask a partial failure; async-accepted (`202`) operations have an unknown final outcome at response time. Endpoints needing finer outcomes can contribute detail via the enrichment hook.

### 3.5 Implementation Plan

1. Add a `created_at`-partitioned migration for `audit_logs` (monthly range partitions, PK becomes `(id, created_at)`) and a partition-lifecycle worker hooked into the existing GC daemon (mechanics in [Section 7.2](#72-migration-mechanics) and [Section 7.3](#73-partition-lifecycle)).
2. Build the **audit log slice** (`server/audit/`) and publish its `Recorder` on `modules.Deps` so it can be injected into the middleware and into other feature slices (scaffolding guidance in [Section 6.4](#64-scaffolding-guidance)).
3. Register `AuditMiddleware` globally so every endpoint is audited in the route-template format from day one. Add exclusion-list support for health, liveness, and metrics routes. Existing audit logging is left in place and runs alongside it.
4. As endpoints are refactored into the `server/` slice architecture, re-home their semantic action into the handler layer via the injected audit slice and remove the corresponding `AuditableTransaction` call in the same change.
5. Obtain Product sign-off on the `Action` contract change and agree the deprecation timeline for the legacy action strings.
6. Remove the legacy audit-logging code (`AuditableTransaction` and the remaining direct `AppendAuditLog` call sites) wholesale after a cool-down period or as part of v3 preparation, once consumers have migrated.

## 4. Middleware Behavior

### 4.1 Registration and Positioning

The middleware registers in the **post-routing** stack (`registration.RegisterFossGlobalMiddleware`), after `ContextMiddleware` (so `bhctx` is available) and positioned to observe the result of `AuthMiddleware`.

A small, explicit **exclusion list** covers routes with no audit value (health checks, liveness probes, metrics scrape endpoints). Excluded routes emit no rows.

### 4.2 Per-Request Behavior

For each non-excluded routed request, the middleware:

1. **Before the handler:** generates a `commit_id`, resolves the actor from `bhctx.AuthCtx` (falling back to anonymous actor + source IP when unauthenticated), and writes an `intent` row.
2. **Calls the handler** with a status-capturing `http.ResponseWriter`.
3. **After the handler:** maps the captured status code to an outcome, drains any handler-contributed context enrichment, and writes the result row reusing the same `commit_id`.

## 5. Action Derivation

### 5.1 Default: Method + Route Template

The middleware derives `Action` as `"METHOD /path/template"` using `routeTemplateFor(muxRouter, request)` — the same gorilla/mux helper already in production for Prometheus metrics in `middleware.go`. Because path parameters collapse to their template placeholder (e.g., `{user_id}` rather than `123`), cardinality is bounded by the number of registered routes, not by request volume.

The API version prefix (e.g., `/api/v2/`) is **preserved** so audit consumers can identify which API version was called and detect traffic migrating between versions. New endpoints are covered the instant they are registered — zero per-route configuration required.

### 5.2 Optional Semantic Overlay

Handlers that need a human-readable, refactor-stable name for a security-significant operation may **optionally** record one through the injected audit slice (see [Section 6](#6-audit-log-slice)), which attaches it to the request context. The middleware records the handler-supplied name when present, falling back to the derived form otherwise.

This is the mechanism for **preserving an existing action string across a refactor**: when an endpoint that previously emitted a typed `AuditLogAction` (e.g. `"CreateUser"`) is moved into the `server/` slice architecture, its handler records that same action through the audit slice, so the database value does not change. The existing `AuditLogAction` constants are not removed or deprecated by this RFC; they remain available for handlers that choose to keep using them.

### 5.3 Model-Level Fields

The enrichment hook also covers the model-level `Fields` that `AuditData()` provides today. Handlers may attach structured detail (affected entity IDs, before/after diffs) to `bhctx`; the middleware drains it into `Fields` on the result row. Generic coverage (actor, action, status, IP) is automatic everywhere; high-value endpoints keep the same rich detail they emit today.

## 6. Audit Log Slice

Both the middleware and individual handlers need to write audit rows. Rather than have each re-implement actor resolution, intent/result pairing, and persistence, this RFC introduces an **audit log slice** in the new `server/` architecture (`server/audit/`). It follows the standard four-layer slice pattern (`appdb → services → handlers`) and exposes a small service that is injected wherever audit rows are written.

### 6.1 Responsibilities

The slice is the single chokepoint for writing to `audit_logs`:

-   Owns `audit_logs` persistence in its `appdb` layer, including the partitioned-table writes described in [Section 7](#7-retention-and-table-stability).
-   Owns the domain `Entry` type in its `services` layer. The three database status values (`intent`, `success`, `failure`) are internal to the appdb layer; callers express outcome through which `Recorder` method they call.
-   Exposes a `Recorder` interface that hides `commit_id` generation and intent/result bookkeeping behind a small surface, so the middleware and handlers never construct `model.AuditLog` rows by hand.

### 6.2 Service Surface

The `Recorder` interface keeps call sites trivial (its placement is explained in [Section 6.4.2](#642-interface-placement)):

```go
// server/audit/services/services.go
package services

// Entry carries the descriptive fields of an audit record. The slice fills in
// commit_id and timestamp; callers supply the rest. Outcome (success vs. failure)
// is expressed by which Recorder method is called, not by a status field on Entry.
type Entry struct {
	Action          model.AuditLogAction
	ActorID         string
	ActorName       string
	ActorEmail      string
	RequestID       string
	SourceIpAddress string
	Fields          map[string]any
}

// Recorder is the consumer-facing surface for writing audit rows. The middleware
// calls Intent before the handler and Success or Failure after; handlers re-home
// semantic actions through the same Recorder when an endpoint is migrated.
type Recorder interface {
	// Intent writes the pre-execution row and returns the commit ID that links
	// it to its eventual result.
	Intent(ctx context.Context, entry Entry) (CommitID, error)
	// Success writes the post-execution row with a successful outcome,
	// reusing the commit ID returned by Intent.
	Success(ctx context.Context, commitID CommitID, entry Entry) error
	// Failure writes the post-execution row with a failed outcome,
	// reusing the commit ID returned by Intent.
	Failure(ctx context.Context, commitID CommitID, entry Entry) error
}
```

A pair of context helpers is exposed for handlers that prefer to contribute via the request context rather than holding a `Recorder` directly. The handler calls `Contribute`; the middleware drains it with `FromContext` on the way out (see the reference implementation in [Section 8](#8-reference-implementation)):

```go
// Contribute records a semantic action and/or model-level Fields on the request
// context. The audit middleware folds the contribution into the result row.
func Contribute(ctx context.Context, action model.AuditLogAction, fields map[string]any)

// FromContext returns the contribution a handler attached during the request, or
// nil if the handler contributed nothing. The middleware calls this after the
// handler returns to overlay the semantic action and Fields onto the result row.
func FromContext(ctx context.Context) *Contribution
```

### 6.3 Injection

The slice's service is constructed once and published on `modules.Deps`, so other slices receive it the same way they receive the router and pool:

```go
// server/modules/modules.go
type Deps struct {
	Router *router.Router
	Pool   *pgxpool.Pool
	Audit  services.Recorder // shared audit writer, injected into other slices
}

func Register(deps Deps) {
	// ... nil checks for Router, Pool, Audit ...
	analysis.Register(deps.Router, deps.Pool, deps.Audit)
}
```

The `AuditMiddleware` receives the same `Recorder` at construction time. Feature handlers receive it through their own `Register` and use it to emit a semantic action or attach model-level `Fields`, without importing the persistence layer. This is what makes preserving a legacy action across a refactor a one-line handler call (see [Section 5.2](#52-optional-semantic-overlay)).

### 6.4 Scaffolding Guidance

The slice is built with the standard four-layer pattern documented in [`bhce/server/README.md`](../server/README.md) ("Adding a new feature module") and the step order in [`bhce/server/implementation_checklist.md`](../server/implementation_checklist.md). It is an **infrastructure slice** rather than a typical read/return feature, so a few deliberate deviations from that pattern apply — they are called out below so an implementer reproduces them intentionally.

#### 6.4.1 Package Layout and Layer Ownership

```
server/audit/
├── audit.go        # Register(): builds appdb → services, starts the worker, returns the Recorder
├── appdb/          # audit_logs writes + partition DDL; defines its own pgxQuerier
├── services/       # Entry, Status, CommitID, Contribution, Recorder, Database interface, the worker
└── handlers/       # OPTIONAL — only if the read endpoint (Section 6.4.5) is migrated here
```

-   **`services`** owns the domain types (`Entry`, `CommitID`, `Contribution`), the public `Recorder` surface, and the consumer-defined `Database` interface (the only methods the service calls). The buffered result-writer worker lives here. `Status` is an internal constant set used only by `appdb` to write the correct `status` column value; it is not part of the public interface.
-   **`appdb`** owns all SQL via `sqlbuilder.PostgreSQL` and `pgx`, defines its package-local `pgxQuerier` (only the pgx methods it uses), maps rows with `db:` tags, and returns `services`-layer sentinels. Because this layer does not use GORM, it must set `created_at` explicitly on every insert (GORM's `autoCreateTime` is not in play) and let the `audit_logs_id_seq` default populate `id`.
-   **`Status`** mirrors the existing `model.AuditLogEntryStatus` values (`intent`, `success`, `failure`) so the database contract is unchanged.

#### 6.4.2 Interface Placement

Per the repo convention, interfaces are defined by the consumer. The wrinkle here is that `Recorder` has **multiple** consumers (the middleware and every feature slice), so it is published from `services` and carried on `modules.Deps` rather than redefined in each consumer. The persistence `Database` interface stays consumer-defined in `services` and is implemented by `appdb.Store`, exactly like other slices.

#### 6.4.3 Construction and Registration Order

Unlike a normal feature slice — whose `Register` only wires its chain and registers routes — the audit slice's `Register` must **return** the constructed `Recorder` (and start its worker) so it can be published on `Deps` *before* any consumer is wired. In `modules.Register` the order is therefore: construct the audit slice first, assign `deps.Audit`, construct `AuditMiddleware` with it, then register every other slice passing `deps.Audit` through. Document this ordering constraint in `audit.go`.

#### 6.4.4 Result Worker and Error Isolation

The `Intent` write is synchronous (its purpose is pre-execution durability). `Result` hands off to a buffered worker owned by `services` (a bounded channel drained by a goroutine started in `Register`, flushed on shutdown). Two invariants the implementer must hold:

-   A `Recorder` error — or a full/closed worker buffer — is **logged and swallowed**; it must never fail or delay the originating request. The middleware already treats both `Intent` and `Result` errors as non-fatal.
-   The buffer's overflow policy (block vs. drop-and-count) is an explicit decision; the recommended default is drop-with-metric so audit pressure cannot stall request handling.

#### 6.4.5 Read Path

The existing read endpoint (`GET /api/v2/audit` → `v2.Resources.ListAuditLogs` → `BloodhoundDB.ListAuditLogs`) keeps working unchanged because partitioning is transparent to `SELECT`, and its `created_at BETWEEN` predicate already enables partition pruning. Migrating that read endpoint into the slice's `handlers`/`appdb` layers is **optional and out of scope** for this RFC; if done, follow the standard slice steps for the read side.

#### 6.4.6 Mocks and Tests

Follow the testing conventions established in the existing slices under `bhce/server/`. Mock generation and test structure for the audit slice should mirror those patterns.

## 7. Retention and Table Stability

### 7.1 Partitioning

`audit_logs` is converted to a **range-partitioned table on `created_at`** as part of this change — not as a later increment. PostgreSQL requires the partition key to be part of the primary key, so the PK becomes `(id, created_at)`. Retention is enforced by **`DROP PARTITION`** — an instant metadata operation with no per-row deletes, no vacuum churn, and no index bloat — rather than a `DELETE` sweep.

**Partition granularity must not exceed the minimum supported retention period.** Dropping a partition removes an entire range boundary-to-boundary; if a partition spans a longer period than the minimum retention window, it is impossible to honor that minimum without keeping data that should have been dropped, or dropping data that should still be retained. For example, if the minimum retention period is one month, monthly partitions are the largest permissible granularity — finer (weekly, daily) partitions are safe but coarser ones are not. The exact minimum retention period is a product decision (see [Section 7.3](#73-partition-lifecycle)); the initial implementation uses monthly partitions because it is the lowest-common-denominator granularity relative to the candidate lower bounds under discussion.

The current `audit_logs` shape the migration starts from (see baseline `00000000000001_init.sql`): `id bigint` backed by `audit_logs_id_seq`, `created_at timestamptz` (currently **nullable**), the actor/request/IP columns, `fields jsonb`, `status varchar(15)` defaulting to `'intent'` with the `status_check` constraint, and `commit_id text`. The GORM model also declares indexes on `created_at`, `actor_id`, and `action`.

### 7.2 Migration Mechanics

This must be a **new** goose migration; the baseline `00000000000001_init.sql` is immutable (see [`bhce/AGENTS.md`](../AGENTS.md)). Place it at `cmd/api/src/database/migration/migrations/YYYYMMDDHHMMSS_audit_logs_partitioning.sql` with `-- +goose Up` / `-- +goose Down` sections; wrap any `DO $$ ... $$` / function blocks (which contain inner semicolons) in `-- +goose StatementBegin` / `-- +goose StatementEnd`.

The non-obvious constraint an implementer must design around: **PostgreSQL cannot convert a populated regular table into a partitioned table in place.** The recommended approach in the `Up` migration is therefore a rename-and-swap:

1.  Rename the existing table aside: `ALTER TABLE audit_logs RENAME TO audit_logs_legacy`.
2.  Create the new parent: `CREATE TABLE audit_logs (... , PRIMARY KEY (id, created_at)) PARTITION BY RANGE (created_at)`, carrying the same columns, the `status_check` constraint, and column defaults. `created_at` must be **`NOT NULL`** (the partition key cannot be null) with a default of `now()`.
3.  Re-point the sequence so inserts keep auto-assigning `id`: keep `audit_logs_id_seq` and set it as the `id` default / owned column on the new table.
4.  Declare the indexes (`created_at`, `actor_id`, `action`) on the **parent**; on PostgreSQL 11+ these propagate to every partition automatically.
5.  Pre-create partitions covering the legacy data's `created_at` range plus the current and next month. Decide explicitly between a **`DEFAULT` partition** (prevents inserts from ever failing on a missing range, but complicates attaching new bounded partitions later) and **strict pre-creation** (no default, lifecycle worker guarantees the next partition exists). This RFC recommends strict pre-creation backed by the worker in [Section 7.3](#73-partition-lifecycle).
6.  Backfill: `INSERT INTO audit_logs SELECT ... FROM audit_logs_legacy`, substituting a non-null `created_at` for any legacy null rows. For large tables, batch the copy.
7.  Drop `audit_logs_legacy` once the backfill is verified (may be deferred to a follow-up migration if a verification window is desired).

The `Down` section must reverse the swap (recreate the flat table, copy rows back, drop the partitioned table and its partitions). Note in the migration that `Down` is best-effort and that any rows written after `Up` into months beyond the original data are still preserved by the copy-back.

### 7.3 Partition Lifecycle

Partition maintenance hooks into the existing GC daemon's 24-hour ticker (`daemons/gc/data_pruning.go`), which already runs `SweepSessions` / `SweepAssetGroupCollections` on startup and on each tick. Add an audit maintenance step that runs in the same place and:

-   **Pre-creates** the next month's partition ahead of time so writes never hit a missing partition.
-   **Drops** partitions whose entire range is older than the configured retention period.

The retention period must be **configurable** within a bounded range. The exact lower and upper bounds are a **product decision that must be resolved before implementation** — they are not set by this RFC. A few constraints the product discussion should account for:

-   The **lower bound** determines the minimum partition granularity (see [Section 7.1](#71-partitioning)). Lowering the bound may require switching to finer partitions (e.g., weekly instead of monthly) and a corresponding migration.
-   The **upper bound** prevents unbounded growth from misconfiguration and should reflect the maximum period the platform is willing to guarantee storage for (storage, regulatory, and operational considerations all apply).
-   The **default value** must sit within [lower bound, upper bound] and should reflect a reasonable operational baseline. A default of 12 months is used as a placeholder for the reference implementation; it is subject to change once Product signs off.

This open question is tracked in the implementation plan ([Section 3.5](#35-implementation-plan), step 5).

Two wiring options for the maintenance call, with the trade-off stated so reviewers can choose:

-   **Add a method to `database.Database`** (e.g. `SweepAuditLogPartitions(ctx)`) for parity with the existing sweeps. Lowest-friction, but keeps partition DDL in the legacy `BloodhoundDB`.
-   **Expose a maintenance interface from the audit slice** and give the daemon that dependency. Keeps all `audit_logs` DDL inside the slice (consistent with [Section 6.1](#61-responsibilities)) at the cost of threading a new dependency into the daemon's constructor.

This RFC recommends the slice-owned maintenance interface so the partitioned table has a single owner, but either satisfies the requirement.

## 8. Reference Implementation

The sketch below is illustrative, not production-ready. It demonstrates how the middleware writes through the injected audit slice ([Section 6](#6-audit-log-slice)) using the existing `responseRecorder` pattern.

```go
// AuditMiddleware writes an intent row before the handler runs and a success/failure
// row after it completes through the injected audit slice, which owns commit_id pairing.
func AuditMiddleware(auditRecorder audit.Recorder, muxRouter *mux.Router, idResolver auth.IdentityResolver, exclusions ExclusionList) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			var (
				ctx            = request.Context()
				requestContext = bhctx.FromRequest(request)
				routeTemplate  = routeTemplateFor(muxRouter, request)
				recorder       = &responseRecorder{delegate: response}
			)

			if exclusions.Contains(routeTemplate) {
				next.ServeHTTP(response, request)
				return
			}

			// Default action: "METHOD /route/template" — zero maintenance, bounded
			// cardinality. The version prefix is preserved (e.g. "/api/v2/") so
			// consumers can distinguish calls across API versions.
			entry := audit.Entry{
				Action:          model.AuditLogAction(request.Method + " " + routeTemplate),
				RequestID:       requestContext.RequestID,
				SourceIpAddress: requestContext.RequestIP,
			}
			// Anonymous actors (e.g. failed login) are recorded by source IP rather than dropped.
			if identity, err := idResolver.GetIdentity(requestContext.AuthCtx); err == nil {
				entry.ActorID = identity.ID.String()
				entry.ActorName = identity.Name
				entry.ActorEmail = identity.Email
			}

			// Intent: synchronous, written for every audited endpoint (reads included)
			// so a denied or failed attempt is always recorded. The slice returns the
			// commit_id that links this intent to its result.
			commitID, err := auditRecorder.Intent(ctx, entry)
			if err != nil {
				slog.ErrorContext(ctx, "failed to write audit intent", slog.String("err", err.Error()))
			}

			next.ServeHTTP(recorder, request)

			// A handler may have re-homed a semantic action (e.g. "CreateUser") and/or
			// model-level Fields through the audit slice's context hook; fold them in.
			if contribution := audit.FromContext(ctx); contribution != nil {
				if contribution.Action != "" {
					entry.Action = contribution.Action
				}
				entry.Fields = contribution.Fields
			}

			// The slice offloads the result write to a buffered worker; see Section 3.4.1.
			// HTTP status -> outcome: <400 success, >=400 failure.
			if recorder.statusCode >= http.StatusBadRequest {
				if err := auditRecorder.Failure(ctx, commitID, entry); err != nil {
					slog.ErrorContext(ctx, "failed to write audit failure", slog.String("err", err.Error()))
				}
			} else {
				if err := auditRecorder.Success(ctx, commitID, entry); err != nil {
					slog.ErrorContext(ctx, "failed to write audit success", slog.String("err", err.Error()))
				}
			}
		})
	}
}
```
