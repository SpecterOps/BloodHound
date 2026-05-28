# bhce/server

Go packages that implement the BloodHound Enterprise HTTP API and feature modules.

## Contents

- [Architecture diagrams](#architecture-diagrams)
- [Package structure](#package-structure)
- [The module system](#the-module-system)
- [Layer architecture](#layer-architecture)
- [Adding a new feature module](#adding-a-new-feature-module)
- [Interface design](#interface-design)
- [Testing](#testing)
- [Mock generation](#mock-generation)
- [Code standards](#code-standards)

## Architecture diagrams

LikeC4 source files live in [`docs/architecture/`](docs/architecture/). They follow the [C4 model](https://c4model.com/) and cover four levels of detail:

**To view the diagrams:**
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

| View | C4 Level | What it shows |
|------|----------|---------------|
| `index` | 1 – System Context | BHE and its users/external systems |
| `containers` | 2 – Containers | Deployable pieces inside BHE |
| `apiServerComponents` | 3 – Components | Go packages inside the API server |
| `analysisInternals` | 4 – Code | Layer architecture inside a feature module |

Render locally with the [LikeC4 CLI](https://likec4.dev/tooling/cli/):

```sh
npx likec4 serve bhce/server/docs/architecture/
```

## Package structure

```
server/
├── modules/        # Shared Deps container and module registry
├── jsonapiv2/
│   └── responses/  # Shared HTTP response helpers (envelopes, error wrappers)
├── docs/
│   └── architecture/   # LikeC4 (C4 model) source for the diagrams above
└── <feature>/      # One directory per vertical feature slice
    ├── <feature>.go        # Register entry point
    ├── appdb/              # Persistence layer (SQL via go-sqlbuilder + pgx)
    ├── handlers/           # HTTP layer (handlers, routes, JSON views)
    └── services/           # Business-logic layer (domain types, interfaces)
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
│  – Defines Analysis interface            │
│  – Auth, status codes, JSON marshalling  │
└────────────────┬─────────────────────────┘
                 │ calls via interface
                 ▼
┌──────────────────────────────────────────┐
│  services  (business-logic layer)        │
│  – Owns domain types                     │
│  – Defines Database interface            │
│  – Maps storage errors to domain errors  │
└────────────────┬─────────────────────────┘
                 │ calls via interface
                 ▼
┌──────────────────────────────────────────┐
│  appdb  (persistence layer)              │
│  – Builds SQL with go-sqlbuilder         │
│  – Executes via pgx                      │
│  – Maps driver errors to sentinels       │
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

    handlers.Register(routerInst, handlerSet)
}
```

Each layer receives only the layer below it. Layers never reach across or skip a boundary.


## Adding a new feature module

Follow these steps to add a new feature that fits the same pattern as `analysis`.

### 1. Create the package tree

```
server/myfeature/
├── myfeature.go         # Register entry point
├── appdb/
│   ├── appdb.go         # Store struct + methods
│   └── appdb_test.go    # Unit tests (pgxmock)
├── handlers/
│   ├── handlers.go      # Handlers struct + MyFeature interface
│   ├── handlers_test.go # Unit tests (httptest)
│   ├── routes.go        # Register(router, handlers)
│   └── views.go         # JSON view types
└── services/
    ├── services.go      # Service struct + domain types + Database interface
    └── services_test.go # Unit tests (mock)
```

### 2. Define domain types and interfaces in `services/services.go`

The services package owns domain types and sentinel errors. The `Database` interface lives here so the persistence layer depends on the consumer (Dependency Inversion).

```go
package services

type MyRecord struct { /* ... */ }
var ErrNotFound = errors.New("not found")

type Database interface {
    GetMyRecord(ctx context.Context, id string) (MyRecord, error)
}

type Service struct{ db Database }
func NewService(databaseInterface Database) *Service { return &Service{db: databaseInterface} }
```

### 3. Implement persistence in `appdb/appdb.go`

Define the minimal `pgxQuerier` interface the store needs (intentionally duplicated per package). Use `go-sqlbuilder` for all SQL construction.

```go
package appdb

type pgxQuerier interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
type Store struct{ db pgxQuerier }
func NewStore(db pgxQuerier) *Store { return &Store{db: db} }
```

### 4. Define the handler interface in `handlers/handlers.go`

The `MyFeature` interface is defined here (consumer side) to enable independent mock substitution in tests.

```go
package handlers

type MyFeature interface {
    GetMyRecord(ctx context.Context, id string) (services.MyRecord, error)
}
type Handlers struct{ feature MyFeature }
func NewHandlersContainer(feature MyFeature) *Handlers { return &Handlers{feature: feature} }
```

### 5. Register routes in `handlers/routes.go`

```go
func Register(routerInst *router.Router, handlers *Handlers) {
    permissions := auth.Permissions()
    routerInst.GET("/api/v2/myfeature/:id", handlers.GetMyRecord).
        RequirePermissions(permissions.AppReadApplicationConfiguration)
}
```

### 6. Wire all layers in `myfeature/myfeature.go`

The feature's `Register` accepts only the infrastructure it actually uses (here, the router and pgx pool):

```go
package myfeature

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/server/myfeature/appdb"
    "github.com/specterops/bloodhound/server/myfeature/handlers"
    "github.com/specterops/bloodhound/server/myfeature/services"
)

func Register(routerInst *router.Router, pool *pgxpool.Pool) {
    var (
        store      = appdb.NewStore(pool)
        svc        = services.NewService(store)
        handlerSet = handlers.NewHandlersContainer(svc)
    )

    handlers.Register(routerInst, handlerSet)
}
```

### 7. Add to the module registry

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

### 8. Add mock targets to `.mockery.yml` and regenerate

```yaml
packages:
  github.com/specterops/bloodhound/server/myfeature/services:
    interfaces:
      Database:
  github.com/specterops/bloodhound/server/myfeature/handlers:
    interfaces:
      MyFeature:
```

Then run `just generate` to produce the mock files.

## Interface design

Interfaces are **always defined by the consumer**, not the producer:

| Interface | Defined in | Implemented by | Purpose |
|-----------|-----------|----------------|---------|
| `handlers.Analysis` | `handlers/handlers.go` | `services.Service` | Allows handler tests to swap in `MockAnalysis` |
| `services.Database` | `services/services.go` | `appdb.Store` | Allows service tests to swap in `MockDatabase` |
| `appdb.pgxQuerier` | `appdb/appdb.go` | `*pgxpool.Pool` | Allows store tests to swap in `pgxmock` |

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

- Receiver functions on structs use `s` as the variable name.
- No named returns — all return variables declared inside the function body.
- Group `var` declarations in a `var ( ... )` block, hoisted to the top of the function.
- Use `any` instead of `interface{}`.
- Prefer descriptive variable names (`databaseInterface` over `db`).
- Test files testing only exported logic use the `_test` package suffix.
- Integration test files carry `//go:build integration` (or `serial_integration`).
