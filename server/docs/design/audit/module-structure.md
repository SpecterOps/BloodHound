# Audit Module Implementation Structure

**Related RFC:** [bh-rfc-7.md](../../../../rfc/bh-rfc-7.md)
**Status:** Implementation Guidance
**Last Updated:** 2026-06-16

## Overview

This document provides detailed implementation guidance for the audit module structure referenced in RFC 7. The audit module follows the module isolation pattern established by `server/clients/` and `server/kennel/`, with all implementation details encapsulated in `internal/` packages.

---

## Package Layout

```
bhce/server/audit/
├── audit.go              # Public API: Service, Entry, Contribution, Maintainer, Register()
├── audit_test.go         # Module-level tests
└── internal/             # All implementation details - inaccessible to other modules
    ├── appdb/
    │   ├── store.go      # PostgreSQL adapter (pgx)
    │   ├── store_test.go
    │   └── partition_ddl.go  # Partition lifecycle SQL
    ├── services/
    │   ├── service.go    # Service implementation (Intent/Success/Failure)
    │   ├── service_test.go
    │   ├── worker.go     # Buffered result-writer goroutine
    │   ├── worker_test.go
    │   ├── types.go      # Internal types (Status constants, Database port)
    │   └── redaction.go  # Sensitive field redaction logic
    └── middleware/       # OPTIONAL: If middleware implementation lives in module
        └── middleware.go
```

---

## Layer Responsibilities

### `audit.go` (Public API)

**Exports only what external consumers need:**

- `Entry` struct - Audit record data
- `Contribution` struct - Handler-contributed enrichment
- `Service` interface - Write operations (Intent/Success/Failure), uses `uuid.UUID` for commit IDs
- `Maintainer` interface - Partition lifecycle operations
- `Register(pool) (Service, Maintainer, error)` - Module construction
- `Contribute(ctx, action, fields) context.Context` - Context helper
- `FromContext(ctx) *Contribution` - Context helper

**Does NOT export:**
- Internal types (`Status`, `Database` port)
- Implementation details (worker, store, SQL)

### `internal/services/`

**Owns domain logic:**

- Service implementation (`Intent`, `Success`, `Failure` methods)
- Consumer-defined `Database` port interface (methods the service calls)
- Buffered result-writer worker (goroutine, channel, flush on shutdown)
- Internal `Status` constants (`intent`, `success`, `failure`)
- Redaction logic for sensitive fields

**Key pattern:** The service defines the `Database` interface it needs, and `appdb.Store` implements it. This is the dependency inversion principle in action.

### `internal/appdb/`

**Owns SQL and PostgreSQL specifics:**

- PostgreSQL adapter via `pgx` (no GORM)
- Implements `services.Database` interface
- Partition DDL (create, drop, attach)
- Row mapping with `db:` struct tags
- Returns `services`-layer types and errors

**Important:** Because this layer does not use GORM, it must:
- Set `created_at` explicitly on every insert (`now()` default)
- Let `audit_logs_id_seq` auto-assign `id`
- Map PostgreSQL errors to `services`-layer sentinels

---

## Interface Placement

Per repo convention, **interfaces are defined by the consumer.**

**`Service` interface:** Exported from public `audit.go` because it has **multiple consumers**:
- Middleware (main consumer)
- BHCE module registry (`modules.Services.Audit`)
- Potentially other feature modules

**`Maintainer` interface:** Exported from public `audit.go` because the GC daemon consumes it:
- GC daemon receives `modules.Services.AuditMaintainer`

**`Database` interface:** Consumer-defined in `internal/services`, implemented by `internal/appdb.Store`:
- Same pattern as `server/clients/internal/services` defines ports for its store

---

## Construction and Registration Order

Unlike a normal feature module whose `Register()` only wires its internal chain and registers routes, the audit module's `Register()` must **return** constructed services.

**Startup flow:**

1. `modules.Register(deps)` is called by entrypoint
2. Registry calls `audit.Register(pool)` first
3. Audit module constructs store, service, worker; starts worker goroutine
4. Audit module returns `(Service, Maintainer, error)`
5. Registry returns `modules.Services{Audit: service, AuditMaintainer: maintainer}`
6. Entrypoint constructs `AuditMiddleware` with `services.Audit`
7. Entrypoint constructs `GCDaemon` with `services.AuditMaintainer`
8. Middleware and daemon are registered on router/scheduler

**Document this ordering constraint in `audit.go` comments.**

---

## Result Worker Implementation

### Purpose

Offload `Success` and `Failure` writes from the request critical path while keeping `Intent` synchronous.

### Design

- **Channel:** Bounded buffered channel (size: 1000 recommended)
- **Goroutine:** Single worker goroutine started in `Register()`, drained continuously
- **Shutdown:** Graceful flush on shutdown (close channel, drain remaining items)

### Error Handling

**Invariants:**

1. A `Service` error (write failure) is **logged and swallowed** - never fails the request
2. A full worker buffer triggers the **drop-with-metric** policy:
   - Result write is dropped (not queued)
   - Prometheus counter `audit_result_drops_total` is incremented
   - Alert fires when drop rate exceeds 1% over 5 minutes

**Rationale:** Audit pressure cannot stall request handling. Dropped results are monitored and alerted.

### Alternative (NOT recommended)

Block-with-timeout: If buffer is full, block the `Success`/`Failure` call for up to N seconds. If timeout expires, drop the write. This adds latency to requests during audit backpressure.

---

## Sensitive Data Redaction

### Redaction Denylist

Case-insensitive substring match on field names in handler-contributed `Fields`:

- `password`
- `secret`
- `token`
- `api_key`
- `apikey`
- `private_key`
- `privatekey`

### Implementation

```go
// internal/services/redaction.go

var sensitivePatternsLower = []string{
    "password", "secret", "token", "api_key", "apikey", "private_key", "privatekey",
}

func redactSensitiveFields(fields map[string]any) map[string]any {
    redacted := make(map[string]any, len(fields))
    for key, value := range fields {
        keyLower := strings.ToLower(key)
        isSensitive := false
        for _, pattern := range sensitivePatternsLower {
            if strings.Contains(keyLower, pattern) {
                isSensitive = true
                break
            }
        }
        if isSensitive {
            redacted[key] = "[REDACTED]"
        } else {
            redacted[key] = value
        }
    }
    return redacted
}
```

Called in `services.Service` before handing `Fields` to `appdb` for persistence.

---

## Mocks and Tests

### Mock Generation

Use `go.uber.org/mock/mockgen` for interface mocks:

```go
//go:generate go tool mockery
```

Add directive to `internal/services/service.go` and `internal/appdb/store.go`.

### Test Structure

- **`audit_test.go`:** Module-level integration tests (public API contract)
- **`internal/services/service_test.go`:** Service logic tests with mocked database
- **`internal/appdb/store_test.go`:** Database adapter tests with test PostgreSQL instance
- **`internal/services/worker_test.go`:** Worker behavior tests (buffer full, shutdown flush)

Follow patterns established in `server/clients/` and `server/kennel/`.

---

## References

- **Pattern examples:** `server/clients/`, `server/kennel/`
- **RFC 7:** `bhce/rfc/bh-rfc-7.md` (architecture decisions, motivation, alternatives)
- **Migration procedure:** `migration-procedure.md`
- **Middleware reference:** `middleware-reference.md`
