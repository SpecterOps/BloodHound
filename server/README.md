# bhce/server

Go packages that implement the BloodHound Enterprise HTTP API and feature modules.

## Contents

-   [Architecture diagrams](#architecture-diagrams)
-   [Package structure](#package-structure)
-   [The module system](#the-module-system)
-   [Layer architecture](#layer-architecture)
-   [Adding a new feature module](#adding-a-new-feature-module)
-   [Interface design](#interface-design)
-   [Testing](#testing)
-   [Mock generation](#mock-generation)
-   [Code standards](#code-standards)

## Architecture diagrams

LikeC4 source files live in [`docs/architecture/`](docs/architecture/). They follow the [C4 model](https://c4model.com/) and cover four levels of detail:

**To view the diagrams:**

> **Note**: LikeC4 requires Node.js ≥ 22. Run `node --version` to confirm before installing.

```bash
# Install LikeC4 (if not already installed)
npm install -g likec4

# Serve interactive diagrams locally
cd bhce/server/docs/architecture
likec4 serve

# Or export to PNG
likec4 export png -o ./diagrams
```

**Available views:**

1. **System Context** (`index`) – Who uses BHE and what external systems it connects to
2. **Containers** (`containers`) – Deployable units: web UI, API server, databases
3. **API Server Components** (`apiServerComponents`) – Go packages and feature modules
4. **Analysis Internals** (`analysisInternals`) – Four-layer architecture within a feature
5. **Type Imports** (`analysisTypeImports`) – Shows how handlers import services types, and appdb imports services errors (dependency inversion)
6. **GET Request Flow** (`analysisGetFlow`) – Complete trace of `GET /api/v2/analysis` through all layers
7. **PUT Request Flow** (`analysisPutFlow`) – Complete trace of `PUT /api/v2/analysis`, including idempotent insert
8. **Shared Database Access** (`sharedDatabaseAccess`) – How multiple features independently access the same tables
9. **Module Registration** (`moduleRegistrationFlow`) – Startup sequence and feature wireup

## Package structure

```
server/
├── modules/        # Shared Deps container and module registry
├── responses/       # Shared HTTP response helpers (envelopes, error wrappers)
├── docs/
│   └── architecture/   # LikeC4 (C4 model) source for the diagrams above
└── <feature>/      # One directory per vertical feature slice
    ├── <feature>.go        # Register entry point
    └── internal/           # Internal implementation packages
        ├── appdb/          # Persistence layer (SQL via go-sqlbuilder + pgx)
        ├── handlers/       # HTTP layer (handlers, JSON views)
        ├── routes/         # Route registration
        └── services/       # Business-logic layer (domain types, interfaces)
```

Each feature is a self-contained vertical slice. It owns every layer from HTTP to SQL; nothing bleeds across feature boundaries.

## The module system

At startup, both `lib/go/services/entrypoint.go` and `bhce/cmd/api/src/services/entrypoint.go` call:

```go
modules.Register(modules.Deps{
    Router: &routerInst,
    Pool:   connections.RDMS.Pool(),
})
```

`modules.Deps` is the shared dependency container; new cross-cutting infrastructure (graph database, filesystem, caches, etc.) is added to that struct so every feature module pulls from a single, consistent place.

`modules.Register` is the central dispatcher — it calls each feature module's `Register` function with the dependencies that module needs:

```go
// server/modules/modules.go
func Register(deps Deps) {
    analysis.Register(deps.Router, deps.Pool)
}
```

Adding a feature is a one-line change in `modules.go`: import the new package and add a call to its `Register` function.

## Layer architecture

Every feature module follows a strict four-layer dependency chain assembled bottom-up inside its `Register` function:

```
HTTP request
     │
     ▼
┌──────────────────────────────────────────┐
│  handlers  (HTTP layer)                  │
│  – Defines the feature's  interface      |
│  – Auth, status codes, JSON marshalling  │
└────────────────┬─────────────────────────┘
                 │ calls via interface
                 ▼
┌──────────────────────────────────────────┐
│  services  (business-logic layer)        │
│  – Owns domain types                     │
│  – Defines the Database interface        |                   │
│  – Maps storage errors to domain errors  │
└────────────────┬─────────────────────────┘
                 │ calls via interface
                 ▼
┌──────────────────────────────────────────┐
│  appdb  (persistence layer)              │
│  – Builds SQL with go-sqlbuilder         │
│  – Executes via pgx                      │
│  – Returns services-layer sentinels      │
└────────────────┬─────────────────────────┘
                 │ pgx pool
                 ▼
           PostgreSQL
```

The `Register` function wires the chain and registers routes. It takes only the infrastructure it directly needs from `modules.Deps` (the router and pgx pool today), making the dependency surface explicit:

```go
// server/analysis/analysis.go
func Register(routerInst *router.Router, pool *pgxpool.Pool) {
    var (
        store      = appdb.NewStore(pool)
        svc        = services.NewService(store)
        handlerSet = handlers.NewHandlersContainer(svc)
    )

    routes.Register(routerInst, handlerSet)
}
```

Each layer receives only the layer below it. Layers never reach across or skip a boundary.

## Adding a new feature module

To add a new endpoint or migrate an existing one to this architecture, follow the **[Implementation Checklist](implementation_checklist.md)**.

The checklist covers:

1. Scaffolding the directory structure
2. Writing e2e tests with production routing
3. Implementing the persistence layer (appdb)
4. Defining domain types and interfaces (services)
5. Defining JSON views (handlers)
6. Implementing HTTP handlers
7. Registering routes
8. Wiring all layers together
9. Adding to the module registry
10. Swapping e2e tests to the new handler
11. Removing old code (for migrations)
12. Preparing for code review

**The sections below provide detailed code examples and architectural context for each layer. Refer to them while following the checklist.**

---

### Package tree structure

```
server/myfeature/
├── myfeature.go         # Register entry point
└── internal/            # Internal implementation packages
    ├── appdb/
    │   ├── appdb.go         # Store struct + methods
    │   └── appdb_test.go    # Unit tests (pgxmock)
    ├── handlers/
    │   ├── handlers.go      # Handlers struct + MyFeature interface
    │   ├── handlers_test.go # Unit tests (httptest)
    │   └── views.go         # JSON view types
    ├── routes/
    │   ├── routes.go        # Register(router, handlers)
    │   └── routes_test.go   # Route registration tests
    └── services/
        ├── services.go      # Service struct + domain types + Database interface
        └── services_test.go # Unit tests (mock)
```

---

### Domain types and interfaces (`internal/services/services.go`)

The services package owns domain types and sentinel errors. The `Database` interface lives here so the persistence layer depends on the consumer (Dependency Inversion). Add `//go:generate go tool mockery` so the mock is regenerated by `just generate`:

```go
package services

//go:generate go tool mockery

type MyRecord struct { /* ... */ }

// Sentinel errors are defined here. The appdb layer returns these same errors
// so that handlers can use errors.Is() without importing appdb.
var ErrNotFound = errors.New("not found")

type Database interface {
    GetMyRecord(ctx context.Context, id string) (MyRecord, error)
}

type Service struct{ db Database }
func NewService(databaseInterface Database) *Service { return &Service{db: databaseInterface} }
```

---

### Persistence layer (`internal/appdb/appdb.go`)

Define the minimal `pgxQuerier` interface using only the pgx methods this store actually calls (each appdb package defines its own copy so the abstraction stays scoped to what is exercised here). Always use `sqlbuilder.PostgreSQL` to build queries, `db:` struct tags to map column names, and `pgx.CollectOneRow`/`pgx.RowToStructByName` to scan results. Return services-layer sentinels (not appdb-specific ones) so callers can use `errors.Is` without importing appdb:

```go
package appdb

// pgxQuerier lists only the pgx methods this package actually calls.
// Add Query and/or Exec depending on what operations the store performs.
type pgxQuerier interface {
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// myRecord is the package-local DB row type. db: tags drive pgx.RowToStructByName scanning.
type myRecord struct {
    ID   string `db:"id"`
    Name string `db:"name"`
}

// toMyRecord translates a raw DB row into the domain model.
func toMyRecord(row myRecord) services.MyRecord {
    return services.MyRecord{ID: row.ID, Name: row.Name}
}

type Store struct{ db pgxQuerier }
func NewStore(db pgxQuerier) *Store { return &Store{db: db} }

func (s *Store) GetMyRecord(ctx context.Context, id string) (services.MyRecord, error) {
    var (
        rows pgx.Rows
        row  myRecord
        err  error
    )

    b := sqlbuilder.PostgreSQL.NewSelectBuilder()
    b.Select("id", "name").From("my_table").Where(b.Equal("id", id))
    sqlQuery, args := b.Build()

    rows, err = s.db.Query(ctx, sqlQuery, args...)
    if err != nil {
        return services.MyRecord{}, err
    }

    row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[myRecord])
    if errors.Is(err, pgx.ErrNoRows) {
        return services.MyRecord{}, services.ErrNotFound
    }
    if err != nil {
        return services.MyRecord{}, fmt.Errorf("reading rows: %w", err)
    }
    return toMyRecord(row), nil
}
```

---

### JSON views (`internal/handlers/views.go`)

View types decouple the wire format from the domain model — the public API shape can evolve independently of internal domain changes, and vice versa. Each view struct has `json:` tags, a standalone `BuildXxxView` builder function that projects from the domain type, and a `JSONView()` method to satisfy `responses.JSONViewer`:

```go
package handlers

import (
    "encoding/json"
    "github.com/specterops/bloodhound/server/myfeature/internal/services"
)

// MyRecordView is the JSON shape returned by the handler. It is separate from
// services.MyRecord so the wire format can change without touching the domain model.
type MyRecordView struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// BuildMyRecordView projects a domain model into the view type.
func BuildMyRecordView(r services.MyRecord) MyRecordView {
    return MyRecordView{ID: r.ID, Name: r.Name}
}

// JSONView satisfies responses.JSONViewer.
func (s MyRecordView) JSONView() ([]byte, error) {
    return json.Marshal(s)
}
```

---

### HTTP handlers (`internal/handlers/handlers.go`)

The `MyFeature` interface is defined here (consumer side) to enable independent mock substitution in tests. Add `//go:generate go tool mockery` so the mock is regenerated by `just generate`. Each handler method reads from the request, calls the service, maps known sentinel errors to appropriate HTTP status codes, and uses the `responses` package to write the JSON envelope:

```go
package handlers

//go:generate go tool mockery

import (
    "context"
    "errors"
    "net/http"
    "github.com/specterops/bloodhound/server/myfeature/internal/services"
    "github.com/specterops/bloodhound/server/responses"
)

type MyFeature interface {
    GetMyRecord(ctx context.Context, id string) (services.MyRecord, error)
}
type Handlers struct{ feature MyFeature }
func NewHandlersContainer(feature MyFeature) *Handlers { return &Handlers{feature: feature} }

func (s Handlers) GetMyRecord(response http.ResponseWriter, request *http.Request) {
    var ctx = request.Context()

    record, err := s.feature.GetMyRecord(ctx, /* extract id from request */)
    if errors.Is(err, services.ErrNotFound) {
        responses.WriteError(ctx, http.StatusNotFound, "record not found", response)
        return
    }
    if err != nil {
        responses.WriteInternalServerError(ctx, err, response)
        return
    }
    responses.WriteBasic(ctx, BuildMyRecordView(record), http.StatusOK, response)
}
```

---

### Route registration (`internal/routes/routes.go`)

```go
package routes

import (
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/cmd/api/src/auth"
    "github.com/specterops/bloodhound/server/myfeature/internal/handlers"
)

func Register(routerInst *router.Router, handlers *handlers.Handlers) {
    permissions := auth.Permissions()
    routerInst.GET("/api/v2/myfeature/:id", handlers.GetMyRecord).
        RequirePermissions(permissions.AppReadApplicationConfiguration)
}
```

---

### Layer wireup (`myfeature/myfeature.go`)

The feature's `Register` accepts only the infrastructure it actually uses (here, the router and pgx pool):

```go
package myfeature

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/server/myfeature/internal/appdb"
    "github.com/specterops/bloodhound/server/myfeature/internal/handlers"
    "github.com/specterops/bloodhound/server/myfeature/internal/routes"
    "github.com/specterops/bloodhound/server/myfeature/internal/services"
)

func Register(routerInst *router.Router, pool *pgxpool.Pool) {
    var (
        store      = appdb.NewStore(pool)
        svc        = services.NewService(store)
        handlerSet = handlers.NewHandlersContainer(svc)
    )

    routes.Register(routerInst, handlerSet)
}
```

---

### Module registry (`server/modules/modules.go`)

In `server/modules/modules.go`, import the new package and call its `Register` from `modules.Register`:

```go
import (
    "github.com/specterops/bloodhound/server/analysis"
    "github.com/specterops/bloodhound/server/myfeature" // ← new
)

func Register(deps Deps) {
    analysis.Register(deps.Router, deps.Pool)
    myfeature.Register(deps.Router, deps.Pool) // ← new
}
```

If the new feature needs infrastructure that isn't on `Deps` yet (graph database, filesystem, caches, etc.), add the field to the `Deps` struct in `modules.go` and populate it from each entrypoint that calls `modules.Register`.

## Interface design

Interfaces are **always defined by the consumer**, not the producer:

| Interface           | Defined in             | Implemented by     | Purpose                                        |
| ------------------- | ---------------------- | ------------------ | ---------------------------------------------- |
| `handlers.Analysis` | `handlers/handlers.go` | `services.Service` | Allows handler tests to swap in `MockAnalysis` |
| `services.Database` | `services/services.go` | `appdb.Store`      | Allows service tests to swap in `MockDatabase` |
| `appdb.pgxQuerier`  | `appdb/appdb.go`       | `*pgxpool.Pool`    | Allows store tests to swap in `pgxmock`        |

## Testing

### Persistence layer (`appdb_test.go`)

Use [pgxmock](https://github.com/pashagolub/pgxmock) to mock the pgx pool. Assert exact SQL and argument values — use `pgxmock.AnyArg()` only when the value is genuinely non-deterministic at test time (e.g., `time.Now()`).

### Service layer (`services_test.go`)

Use the generated `MockDatabase`. Pass concrete argument values to mock expectations; avoid `mock.Anything`.

### Handler layer (`handlers_test.go`)

Use the generated `MockAnalysis`. Capture responses with `httptest.NewRecorder`. Pass `request.Context()` to mock expectations rather than `mock.Anything`.

### Integration tests (`appdb_integration_test.go`)

Carry the `//go:build integration` build tag and use [pgtestdb](https://github.com/peterldowns/pgtestdb) for an isolated PostgreSQL instance.

```sh
go test -C bhce -tags integration ./server/myfeature/appdb/...
```

## Mock generation

Mocks are generated by [mockery](https://vektra.github.io/mockery/) from `bhce/.mockery.yml`. After adding an interface, run:

```sh
cd bhce && just generate
```

Never edit generated mock files by hand.

## Code standards

See [`bhce/AGENTS.md`](../../AGENTS.md) for the full list. Key points:

-   Receiver functions on structs use `s` as the variable name.
-   No named returns — all return variables declared inside the function body.
-   Group `var` declarations in a `var ( ... )` block, hoisted to the top of the function.
-   Use `any` instead of `interface{}`.
-   Prefer descriptive variable names (`databaseInterface` over `db`).
-   Test files testing only exported logic use the `_test` package suffix.
-   Integration test files carry `//go:build integration` (or `serial_integration`).
