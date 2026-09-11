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
-   When an endpoint is refactored into the new `server/` module architecture, its semantic action (e.g. `"CreateUser"`) is **re-homed into the handler layer** and emitted through the injected audit service (see [Section 6](#6-audit-module)), and the old `AuditableTransaction` call is removed in the same change.

**Dual-write tagging:** To distinguish between duplicate rows written during the migration window, the `audit_logs` table includes a `source` column (`VARCHAR(20)`, default `'middleware'`). Legacy writes are tagged with `'legacy'`, middleware writes with `'middleware'`. Consumers should prefer the semantic action over the route-template when both exist for the same `commit_id`. The migration backfills existing rows with `source = 'legacy'`.

The legacy audit-logging code is removed wholesale only at the end of the transition (see [Section 3.5](#35-implementation-plan)).

#### 3.1.3 Out-of-Band Audit Writes

Several audit writes today occur outside the normal HTTP handler path and are **explicitly out of scope** for the middleware. These will continue to write directly via `AppendAuditLog`:

-   **Login events** (`LoginAttempt`) — written by the authentication layer.
-   **Support account operations** (`CreateSupportUserSessionAttempt`, `InvalidateSupportUserSession`, `InvalidateAllSupportUserSessions`) — written from `lib/go/api/tools/support_account`, which bypasses the HTTP stack.
-   **Daemon-triggered actions** (`CreateAuthTokens`, `UpdateAuthTokens`) — written by scheduled daemons, not request handlers.

#### 3.1.4 Deprecation Timeline for Legacy Audit Formats

The transition from legacy audit logging to the middleware-based approach follows a **3-month deprecation window** aligned with the default retention period:

-   **Month 0 (middleware deployment):** The middleware is registered globally and begins emitting the new route-template format for all endpoints. Legacy audit logging (`AuditableTransaction` and direct `AppendAuditLog` calls in handlers) continues to operate in parallel, so endpoints emit both formats during the transition.
-   **Months 0–3 (migration window):** Endpoints are incrementally refactored into the `server/` module architecture. Each refactor re-homes its semantic action into the handler layer (via the audit service or context helpers) and removes the corresponding `AuditableTransaction` call. Both old and new audit formats coexist in the database during this period.
-   **Month 3 (deprecation):** Legacy audit-logging code (`AuditableTransaction`, handler-layer `AppendAuditLog` calls, and the typed `AuditLogAction` constants not re-homed into handlers) is marked deprecated. The audit log UI and any customer integrations must migrate to consume the new route-template format or the re-homed semantic actions.
-   **Month 3+ (removal):** After the 3-month window, the legacy audit-logging infrastructure is removed wholesale. Only out-of-band writes (login events, support account operations, daemon-triggered actions per [Section 3.1.3](#313-out-of-band-audit-writes)) continue using the audit service's direct write path.

**Non-API endpoint audit writes** — specifically, any audit writes that occur outside the standard HTTP request/response cycle and do not correspond to a registered API route — are subject to the same 3-month support window. After month 3, only the explicitly scoped out-of-band writes listed in Section 3.1.3 are preserved; all other non-API audit paths are removed.

**Consumer migration guidance:**

Consumers of audit logs (UI, SIEM integrations, analytics) must adapt to the new action format during the 3-month window:

-   **Audit Log UI:** Update filter/search logic to recognize both typed constants (e.g., `"CreateUser"`) and route-template format (e.g., `"POST /api/v2/users"`). During the migration window, prefer rows with `source = 'legacy'` for existing semantic actions. After month 3, rely on route-template format or handler-contributed semantic actions (distinguished by the absence of HTTP method prefix).
-   **SIEM integrations:** Update alert rules and parsers to match the new format. Example: an alert that filtered on `Action = "CreateUser"` should be updated to match either `Action = "CreateUser"` OR `Action = "POST /api/v2/users"` during the transition, then switch to the route-template pattern after month 3 (or use the re-homed semantic action if the endpoint is refactored).
-   **Analytics queries:** Add logic to handle both formats. The `source` column can be used to distinguish middleware writes from legacy writes when deduplication is needed.

### 3.2 Security & Compliance

#### 3.2.1 Read Auditing

The middleware writes the full `intent` → `success`/`failure` pair for **all** audited endpoints, reads included. Reads frequently expose sensitive data, so a record of what was attempted and whether that attempt succeeded is as valuable for a GET as for a mutation. A failed or denied read is a security-relevant event — the intent row preserves evidence of the attempt even if the handler errors, panics, or the connection drops before a result row is written.

#### 3.2.2 Sensitive Data Capture

Request bodies must **not** be blanket-logged. The middleware records method, route template, route parameters, and handler-contributed fields only, with a **redaction denylist** for known secret fields. This matches the existing behavior where `AuditData()` deliberately omits credentials and tokens.

**Redaction denylist** (case-insensitive substring match in field names):
- `password`
- `secret`
- `token`
- `api_key`
- `apikey`
- `private_key`
- `privatekey`

Any handler-contributed field whose name contains one of these substrings has its value replaced with `"[REDACTED]"` before being written to `Fields`. The denylist is maintained in a centralized constant within the audit module (`internal/services/redaction.go`) and is updated by the security team as new sensitive field patterns are identified.

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

Auditing every endpoint with a full intent+result pair increases row count compared to today's mutation-only coverage. The actual multiplier depends on current audit coverage (unknown) and the ratio of read vs. write operations in production traffic (unknown).

**Expected characteristics:**

-   Every audited endpoint writes 2 rows (intent + result) instead of 1 row for mutations only
-   Read-heavy workloads (e.g., UI polling, dashboard refreshes) will see higher row counts than write-heavy workloads
-   Exact storage requirements depend on deployment size, request volume, and field usage patterns

This is addressed through monthly partitioning with `DROP PARTITION` retention (see [Section 7](#7-retention-and-table-stability)) rather than by weakening the audit record. The partition-drop operation is near-instant (metadata-only), preventing table bloat.

Reducing volume at the source — primarily UI polling behavior — is a recommended follow-up and out of scope for this proposal. **Monitoring:** Track partition sizes and row counts via Prometheus metrics to establish baselines and validate retention settings against actual usage.

#### 3.4.3 HTTP Status as Outcome Signal

`< 400` maps to `success`; `>= 400` maps to `failure`. This is documented as "the API's response to the caller," not "the business outcome." A `200` can mask a partial failure; async-accepted (`202`) operations have an unknown final outcome at response time. Endpoints needing finer outcomes can contribute detail via the enrichment hook.

### 3.5 Implementation Plan

1. Add a `created_at`-partitioned migration for `audit_logs` (monthly range partitions, PK becomes `(id, created_at)`) and a partition-lifecycle worker hooked into the existing GC daemon (mechanics in [Section 7.2](#72-migration-mechanics) and [Section 7.3](#73-partition-lifecycle)). Configure retention bounds at 1–12 months with a 3-month default.
2. Build the **audit module** (`bhce/server/audit/`) following the module isolation pattern (see [Section 6.4](#64-module-structure)). The module exports `audit.Service`, `audit.Entry`, and context helpers from its public API; all implementation details live in `internal/` packages.
3. Create the **BHCE module registry** (`bhce/server/modules/modules.go`) that calls `audit.Register(pool)` and returns `modules.Services{Audit: auditService}` for injection into middleware and other consumers (see [Section 6.3](#63-module-registry-and-injection)).
4. Register `AuditMiddleware` globally by calling `modules.Register(deps)` in the entrypoint to obtain the audit service, then constructing the middleware with `services.Audit`. Configure the middleware to audit every endpoint in the route-template format from day one. Add exclusion-list support for the health check route (`/health`). Existing audit logging is left in place and runs alongside it.
5. As endpoints are refactored into feature modules, re-home their semantic action into the handler layer via `audit.Contribute()` or by accepting `audit.Service` as a parameter to the module's `Register()` function. Remove the corresponding `AuditableTransaction` call in the same change.
6. After the 3-month migration window (see [Section 3.1.4](#314-deprecation-timeline-for-legacy-audit-formats)), mark legacy audit-logging infrastructure as deprecated and notify consumers to migrate to the new format.
7. Remove the legacy audit-logging code (`AuditableTransaction` and handler-layer `AppendAuditLog` call sites) wholesale after the deprecation period, preserving only the out-of-band writes documented in [Section 3.1.3](#313-out-of-band-audit-writes).

## 4. Middleware Behavior

### 4.1 Registration and Positioning

The middleware registers in the **post-routing** stack, positioned to wrap the handler and have access to both `ContextMiddleware` (for `bhctx`) and `AuthMiddleware` (for actor identity).

**Execution order:**
1. Pre-routing: `ContextMiddleware` → `MetricsMiddleware` → `LoggingMiddleware`
2. Route matching
3. Post-routing (wraps handler): `PanicHandler` → `AuthMiddleware` → `AuditMiddleware` → **Handler**

The audit middleware must register AFTER the existing post-routing middleware so it runs inside the auth context but still wraps the handler.

**Concrete wiring** in the startup entrypoint (`lib/go/services/entrypoint.go` or registration package):

```go
// Register BHCE modules and obtain cross-cutting services
bhceDeps := modules.Deps{
	Router: &routerInst,
	Pool:   connections.RDMS.Pool(),
}
bhceServices, err := modules.Register(bhceDeps)
if err != nil {
	return fmt.Errorf("failed to register BHCE modules: %w", err)
}

// Construct and register audit middleware with the service
auditMiddleware := middleware.NewAuditMiddleware(
	bhceServices.Audit,
	routerInst.MuxRouter(),
	identityResolver,
	[]string{"/health"}, // exclusion list
)
routerInst.UsePostrouting(auditMiddleware)
```

A small, explicit **exclusion list** covers routes with no audit value:
- `/health` - Health check endpoint (returns 200 OK, no auth required)

Excluded routes emit no rows. Future additions to the exclusion list follow the criteria: (1) no auth context required, (2) high request volume (>10 req/sec), (3) no security value.

**Note:** The `/metrics` endpoint and other pprof/debug endpoints are served on a separate Tools API daemon (different port configured by `MetricsPort`) and are not part of the main API router, so they are not audited by this middleware.

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

Handlers that need a human-readable, refactor-stable name for a security-significant operation may **optionally** record one through the injected audit service (see [Section 6](#6-audit-module)), which attaches it to the request context via `audit.Contribute()`. The middleware records the handler-supplied name when present, falling back to the derived form otherwise.

This is the mechanism for **preserving an existing action string across a refactor**: when an endpoint that previously emitted a typed `AuditLogAction` (e.g. `"CreateUser"`) is moved into the `server/` module architecture, its handler records that same action through `audit.Contribute()`, so the database value does not change. The existing `AuditLogAction` constants are not removed or deprecated by this RFC; they remain available for handlers that choose to keep using them.

### 5.3 Model-Level Fields

The enrichment hook also covers the model-level `Fields` that `AuditData()` provides today. Handlers may attach structured detail (affected entity IDs, before/after diffs) to `bhctx`; the middleware drains it into `Fields` on the result row. Generic coverage (actor, action, status, IP) is automatic everywhere; high-value endpoints keep the same rich detail they emit today.

## 6. Audit Module

Both the middleware and individual handlers need to write audit rows. Rather than have each re-implement actor resolution, intent/result pairing, and persistence, this RFC introduces an **audit module** in the BHCE server architecture (`bhce/server/audit/`). It follows the module isolation pattern established by `server/clients/` and `server/kennel/`, using `internal/` packages for encapsulation and exposing a small public API for audit recording.

### 6.1 Responsibilities

The audit module is the single chokepoint for writing to `audit_logs`:

-   Owns `audit_logs` persistence in its `internal/appdb` layer, including the partitioned-table writes described in [Section 7](#7-retention-and-table-stability).
-   Owns the domain `Entry` type and exports it from the public API (`audit.go`). The three database status values (`intent`, `success`, `failure`) are internal to the appdb layer; callers express outcome through which `Service` method they call (`Intent`, `Success`, `Failure`).
-   Exposes a `Service` interface that hides `commit_id` generation and intent/result bookkeeping behind a small surface, so the middleware and handlers never construct `model.AuditLog` rows by hand.
-   Owns the buffered result-writer worker that offloads result writes from the request critical path.

### 6.2 Public API

The audit module exports a minimal public API from `audit.go`. All implementation details live in `internal/` packages and are inaccessible to other modules.

**Public types and interface:**

```go
// bhce/server/audit/audit.go
package audit

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// Entry carries the descriptive fields of an audit record. The service fills in
// commit_id and timestamp; callers supply the rest. Outcome (success vs. failure)
// is expressed by which Service method is called, not by a status field on Entry.
type Entry struct {
	Action          model.AuditLogAction
	ActorID         string
	ActorName       string
	ActorEmail      string
	RequestID       string
	SourceIpAddress string
	Fields          map[string]any
}

// Service is the public interface for writing audit rows. The middleware
// calls Intent before the handler and Success or Failure after; handlers re-home
// semantic actions through the same Service when an endpoint is migrated.
type Service interface {
	// Intent writes the pre-execution row and returns the commit ID (uuid.UUID)
	// that links it to its eventual result.
	Intent(ctx context.Context, entry Entry) (uuid.UUID, error)
	// Success writes the post-execution row with a successful outcome,
	// reusing the commit ID returned by Intent.
	Success(ctx context.Context, commitID uuid.UUID, entry Entry) error
	// Failure writes the post-execution row with a failed outcome,
	// reusing the commit ID returned by Intent.
	Failure(ctx context.Context, commitID uuid.UUID, entry Entry) error
}
```

**Register function:**

```go
// Maintainer provides partition lifecycle operations for the GC daemon.
type Maintainer interface {
	PreCreateNextPartition(ctx context.Context) error
	DropExpiredPartitions(ctx context.Context, retentionMonths int) error
}

// Register wires the audit service to its PostgreSQL store, starts the
// buffered result-writer worker, and returns both the service (for middleware
// and handlers) and the maintainer (for the GC daemon).
// This is called by the BHCE module registry during startup.
func Register(pool *pgxpool.Pool) (Service, Maintainer, error)
```

**Context helpers** for handlers that prefer to contribute via the request context rather than holding a `Service` directly:

```go
// Contribution holds the optional semantic action and model-level Fields a handler
// attached during the request. The middleware folds the contribution into the result row.
type Contribution struct {
	Action model.AuditLogAction
	Fields map[string]any
}

// Contribute attaches a semantic action and/or model-level Fields to the request
// context using context.WithValue and returns the updated context. Handlers must
// propagate the returned context via request.WithContext() to make the contribution
// visible to the middleware's post-handler FromContext call.
func Contribute(ctx context.Context, action model.AuditLogAction, fields map[string]any) context.Context

// FromContext extracts the contribution a handler attached during the request, or
// returns nil if the handler contributed nothing. The middleware calls this after the
// handler returns to overlay the semantic action and Fields onto the result row.
func FromContext(ctx context.Context) *Contribution
```

### 6.3 Module Registry and Injection

The audit module is registered by the BHCE module registry, which calls `audit.Register()` and returns the service for injection into middleware and other consumers.

**BHCE Module Registry** (`bhce/server/modules/modules.go`):

```go
package modules

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/audit"
)

// Deps carries the shared infrastructure that BHCE modules need.
// BHE's modules.Deps embeds this struct to inherit Router and Pool.
type Deps struct {
	Router *router.Router
	Pool   *pgxpool.Pool
}

// Services holds the constructed BHCE cross-cutting services that need to be
// injected into middleware or passed to other modules.
type Services struct {
	Audit           audit.Service
	AuditMaintainer audit.Maintainer
}

// Register wires up BHCE modules and returns services that middleware and
// other modules depend on.
func Register(deps Deps) (*Services, error) {
	if deps.Router == nil {
		panic("modules: Register requires a non-nil Router")
	}
	if deps.Pool == nil {
		panic("modules: Register requires a non-nil Pool")
	}

	// Build the audit service and maintainer
	auditService, auditMaintainer, err := audit.Register(deps.Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to register audit module: %w", err)
	}

	// Future: Register other BHCE-specific feature modules here
	// analysis.Register(deps.Router, deps.Pool, auditService)

	return &Services{
		Audit:           auditService,
		AuditMaintainer: auditMaintainer,
	}, nil
}
```

**Middleware injection** in the startup entrypoint:

```go
// In lib/go/services/entrypoint.go or equivalent registration code
bhceDeps := modules.Deps{
	Router: &routerInst,
	Pool:   connections.RDMS.Pool(),
}

// Register BHCE modules and obtain cross-cutting services
bhceServices, err := modules.Register(bhceDeps)
if err != nil {
	return fmt.Errorf("failed to register BHCE modules: %w", err)
}

// Construct and register audit middleware with the service
auditMiddleware := middleware.NewAuditMiddleware(
	bhceServices.Audit,
	routerInst.MuxRouter(),
	identityResolver,
	[]string{"/health"}, // exclusions
)
routerInst.UsePostrouting(auditMiddleware)

// BHE modules register separately, embedding the BHCE Deps
bhemodules.Register(bhemodules.Deps{
	Deps: bhceDeps, // Embeds Router and Pool
	// ... BHE-specific dependencies ...
})
```

Feature handlers that need to emit semantic actions or attach model-level `Fields` can receive `audit.Service` as a parameter to their module's `Register()` function, or use the context helpers (`audit.Contribute`, `audit.FromContext`). This is what makes preserving a legacy action across a refactor a one-line handler call (see [Section 5.2](#52-optional-semantic-overlay)).

### 6.4 Implementation Guidance

The audit module follows the isolation pattern established by `server/clients/` and `server/kennel/`, with all implementation details encapsulated in `internal/` packages.

**Detailed implementation guidance** including package layout, layer responsibilities, interface placement, registration order, worker implementation, error handling, redaction logic, and testing patterns is documented in:

📄 **[`bhce/server/docs/design/audit/module-structure.md`](../server/docs/design/audit/module-structure.md)**

**Key points:**

-   **Public API** (`audit.go`): Exports `Service`, `Maintainer`, `Entry`, `Contribution`, `Register()`, and context helpers
-   **Internal implementation**: `internal/appdb/` (PostgreSQL), `internal/services/` (domain logic, worker)
-   **Result worker**: Buffered channel with drop-with-metric overflow policy
-   **Sensitive data**: Redaction denylist applied to handler-contributed fields
-   **Testing**: Follow `server/clients/` and `server/kennel/` patterns

The existing read endpoint (`GET /api/v2/audit`) keeps working unchanged because partitioning is transparent to `SELECT`. Migrating that endpoint into the audit module is **optional and out of scope**.

## 7. Retention and Table Stability

### 7.1 Partitioning

`audit_logs` is converted to a **range-partitioned table on `created_at`** as part of this change — not as a later increment. PostgreSQL requires the partition key to be part of the primary key, so the PK becomes `(id, created_at)`. Retention is enforced by **`DROP PARTITION`** — an instant metadata operation with no per-row deletes, no vacuum churn, and no index bloat — rather than a `DELETE` sweep.

**Partition granularity must not exceed the minimum supported retention period.** Dropping a partition removes an entire range boundary-to-boundary; if a partition spans a longer period than the minimum retention window, it is impossible to honor that minimum without keeping data that should have been dropped, or dropping data that should still be retained. With a minimum retention period of 1 month (see [Section 7.3](#73-partition-lifecycle)), monthly partitions are the largest permissible granularity — finer (weekly, daily) partitions are safe but coarser ones are not. Monthly partitions align exactly with the month-only retention granularity requirement.

The current `audit_logs` shape the migration starts from (see baseline `00000000000001_init.sql`): `id bigint` backed by `audit_logs_id_seq`, `created_at timestamptz` (currently **nullable**), the actor/request/IP columns, `fields jsonb`, `status varchar(15)` defaulting to `'intent'` with the `status_check` constraint, and `commit_id text`. The GORM model also declares indexes on `created_at`, `actor_id`, and `action`.

### 7.2 Migration Mechanics

This must be a **new** goose migration; the baseline `00000000000001_init.sql` is immutable (see [`bhce/AGENTS.md`](../AGENTS.md)). The migration converts `audit_logs` from a regular table to a range-partitioned table.

**Key constraint:** PostgreSQL cannot convert a populated regular table into a partitioned table in place. The migration uses a **rename-and-swap** strategy: rename the existing table aside, create a new partitioned table with the same name, backfill data, then drop the old table.

**Detailed step-by-step SQL procedure** including partition pre-creation, backfill logic, null handling, and rollback strategy is documented in:

📄 **[`bhce/server/docs/design/audit/migration-procedure.md`](../server/docs/design/audit/migration-procedure.md)**

**Key decisions:**

-   **NULL `created_at` handling:** `COALESCE(created_at, '2020-01-01'::timestamptz)` places null rows in a legacy partition
-   **DEFAULT partition:** Created for operational safety (prevents insert failures on missing partitions)
-   **Dual-write tagging:** Backfill sets `source = 'legacy'` for all existing rows
-   **Primary key:** Changes to `(id, created_at)` (composite, partition key must be in PK)
-   **Indexes:** Declared on parent table, propagate to partitions automatically (PostgreSQL 11+)

### 7.3 Partition Lifecycle

Partition maintenance hooks into the existing GC daemon's 24-hour ticker (`daemons/gc/data_pruning.go`), which already runs `SweepSessions` / `SweepAssetGroupCollections` on startup and on each tick. Add an audit maintenance step that runs in the same place and:

-   **Pre-creates** the next month's partition ahead of time so writes never hit a missing partition.
-   **Drops** partitions whose entire range is older than the configured retention period.

The retention period must be **configurable** within a **bounded range of 1 to 12 months**, with a **default of 3 months**. This configuration enforces the following constraints:

-   The **lower bound of 1 month** establishes the minimum retention period and aligns with the monthly partition granularity (see [Section 7.1](#71-partitioning)). Retention periods shorter than one month are not supported.
-   The **upper bound of 12 months** prevents unbounded growth from misconfiguration and sets the maximum period the platform guarantees storage for. Retention periods longer than 12 months (including unbounded retention) are explicitly not supported.
-   The **default value of 3 months** provides a reasonable operational baseline that balances storage cost, audit coverage, and typical investigation windows. This default aligns with the deprecation timeline for legacy audit formats (see [Section 3.1.4](#314-deprecation-timeline-for-legacy-audit-formats)).
-   The retention configuration must be **expressable only in whole months**. Finer granularity (weeks, days) and coarser granularity (years) are not supported.

The configuration value is validated at daemon startup to ensure it falls within `[1, 12]` months. Out-of-range values cause the daemon to fail-fast with a clear error message rather than silently clamping or defaulting.

The audit module exposes a **maintenance interface** that the GC daemon calls to manage partition lifecycle. The BHCE module registry returns `modules.Services.AuditMaintainer`, which the entrypoint passes to the GC daemon:

**Maintenance interface** (exported from `bhce/server/audit/audit.go`):

```go
type Maintainer interface {
	PreCreateNextPartition(ctx context.Context) error
	DropExpiredPartitions(ctx context.Context, retentionMonths int) error
}
```

**GC daemon integration** (in startup entrypoint):

```go
// After modules.Register returns services
bhceServices, err := modules.Register(bhceDeps)
if err != nil {
	return err
}

// Pass AuditMaintainer to GC daemon constructor
gcDaemon := gc.NewDaemon(
	connections.RDMS,
	bhceServices.AuditMaintainer, // New dependency
	retentionConfig,
)
```

The GC daemon calls `PreCreateNextPartition()` and `DropExpiredPartitions(retentionMonths)` on its 24-hour ticker, alongside the existing `SweepSessions` and `SweepAssetGroupCollections` operations. This approach ensures the audit module fully owns its storage lifecycle (consistent with [Section 6.1](#61-responsibilities)) while keeping the partition DDL encapsulated within the module.

## 8. Reference Implementation

The audit middleware demonstrates how HTTP-layer auditing works with the injected audit service. **A complete reference implementation** including constructor, handler logic, response recorder, route template helper, and handler contribution examples is documented in:

📄 **[`bhce/server/docs/design/audit/middleware-reference.md`](../server/docs/design/audit/middleware-reference.md)**

**Key patterns:**

-   **Intent before handler:** Synchronous write, records pre-execution state
-   **Result after handler:** Asynchronous write via buffered worker, records post-execution state
-   **Error handling:** Log and swallow - audit failures never fail the request
-   **Status mapping:** `< 400` → success, `>= 400` → failure
-   **Context enrichment:** Handlers optionally contribute semantic actions/fields via `audit.Contribute()`
-   **Anonymous actors:** Missing auth context leaves actor fields empty (source IP captured)
-   **Route exclusions:** `/health` and other non-audit-worthy endpoints are excluded
-   **Registration timing:** Must be registered AFTER existing post-routing middleware to wrap inside `AuthMiddleware`

The middleware uses the existing `responseRecorder` pattern to capture HTTP status codes.
