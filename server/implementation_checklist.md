# Implementation Checklist

Use this checklist when creating a new feature endpoint or migrating an existing endpoint to the `server/` module architecture.

The pattern is **test-first migration**: write an integration test that documents the current endpoint's contract, complete the migration, then confirm the same test still passes.

**Commit after each step.**

For architectural background, terminology, and detailed code examples, see [`bhce/server/README.md`](README.md).

---

## Step 1 – Scaffold the feature directory

Create the directory tree under `server/<feature>/`:

```
server/<feature>/
├── <feature>.go             # Register() entry point
└── internal/                # Internal implementation packages
    ├── appdb/
    │   ├── appdb.go             # Store struct + methods
    │   └── appdb_test.go        # Unit tests (pgxmock)
    ├── handlers/
    │   ├── handlers.go          # Handlers struct + MyFeature interface
    │   ├── handlers_test.go     # Unit tests (httptest)
    │   └── views.go             # JSON view types
    ├── routes/
    │   ├── routes.go            # Register(router, handlers)
    │   └── routes_test.go       # Route registration tests
    └── services/
        ├── services.go          # Service struct + domain types + Database interface
        └── services_test.go     # Unit tests (mock)
```

**Checklist:**

-   [ ] Create the directory structure
-   [ ] Add the license header to every new file (run `just generate` or copy from an existing file)

---

## Step 2 – Write an e2e integration test against the existing endpoint

Before touching any production code, write a test that covers the HTTP contract of the endpoint in its current form. The goal is a green baseline that will still pass after migration.

**For new endpoints:**
Wire a minimal stub handler that returns a fixed response matching the planned contract.

**For existing endpoints (migration):**
Wire the existing handler from `v2.Resources` or equivalent.

**Checklist:**

-   [ ] Add `<feature>_e2e_test.go` with build tag `//go:build integration`
-   [ ] Wire the existing (old) handler using production routing:
    ```go
    // Use production route registration, not manual mux setup
    var (
        cfg        = config.Configuration{}
        authorizer = auth.NewAuthorizer(db)
        routerInst = router.NewRouter(cfg, authorizer, "")
    )

    // Register global middleware (required for auth to work)
    registration.RegisterFossGlobalMiddleware(&routerInst, cfg, resolver, auther, db)

    // Register old v2 routes using v2.Resources
    resources := v2.NewResources(db, /* ... */)
    registration.NewV2API(resources, &routerInst)

    handler := routerInst.Handler()
    server  := httptest.NewServer(handler)
    ```
-   [ ] Create authenticated requests with real JWT tokens using `api.Authenticator.CreateSession`
-   [ ] Assert every status code and relevant JSON field the endpoint can return (happy path + error paths)
-   [ ] Test authentication requirements (401 when unauthenticated)
-   [ ] Confirm the test is green: `go test -tags integration ./server/<feature>/...`

**Why use production routing in tests:**
This ensures the test catches wiring errors in route registration, middleware application, and authentication requirements — not just handler logic.

---

## Step 3 – Implement persistence in `internal/appdb/appdb.go`

The persistence layer uses [go-sqlbuilder](https://github.com/huandu/go-sqlbuilder) for query construction and [pgx v5](https://github.com/jackc/pgx) for execution.

**Pattern:**

```go
package appdb

import (
    "context"
    "errors"
    "fmt"

    "github.com/huandu/go-sqlbuilder"
    "github.com/jackc/pgx/v5"
    "github.com/specterops/bloodhound/server/<feature>/internal/services"
)

// pgxQuerier lists only the pgx methods this package actually calls.
// Each appdb package defines its own copy scoped to what it exercises.
type pgxQuerier interface {
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    // Add Exec if the store performs writes
}

// myRecord is the package-local DB row type.
// db: tags map column names for pgx.RowToStructByName scanning.
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
        return services.MyRecord{}, services.ErrNotFound  // ← Return services sentinel
    }
    if err != nil {
        return services.MyRecord{}, fmt.Errorf("reading rows: %w", err)
    }
    return toMyRecord(row), nil
}
```

**Checklist:**

-   [ ] Define a package-local `pgxQuerier` interface with only the pgx methods this store calls
-   [ ] Define package-local row structs with `db:` tags for `pgx.RowToStructByName`
-   [ ] Define `toXxx` translator functions that convert row structs to domain types
-   [ ] Implement `Store` methods using `sqlbuilder.PostgreSQL` for all SQL construction
-   [ ] Use `pgx.CollectOneRow` + `pgx.RowToStructByName` for single-row queries
-   [ ] Use `pgx.CollectRows` + `pgx.RowToStructByName` for multi-row queries
-   [ ] Return services-layer sentinels (e.g., `services.ErrNotFound`) not driver errors, so callers can use `errors.Is` without importing `appdb`
-   [ ] Write unit tests in `internal/appdb/appdb_test.go` using [pgxmock](https://github.com/pashagolub/pgxmock)
    -   Assert exact SQL and arguments (use `pgxmock.AnyArg()` only for non-deterministic values like `time.Now()`)
-   [ ] Write integration tests in `internal/appdb/appdb_integration_test.go` using [pgtestdb](https://github.com/peterldowns/pgtestdb) with `//go:build integration`

---

## Step 4 – Define domain types and interfaces in `internal/services/services.go`

The services package owns domain types and sentinel errors. The `Database` interface lives here (not in appdb) to enable Dependency Inversion — the persistence layer depends on the consumer.

**Pattern:**

```go
package services

import (
    "context"
    "errors"
)

//go:generate go tool mockery

type MyRecord struct {
    ID   string
    Name string
}

// Sentinel errors are defined here.
// The appdb layer returns these same errors so handlers can use errors.Is()
// without importing appdb.
var (
    ErrNotFound = errors.New("not found")
)

// Database defines the persistence operations required by this service.
// Defined here (consumer side) to enable dependency inversion.
type Database interface {
    GetMyRecord(ctx context.Context, id string) (MyRecord, error)
}

type Service struct {
    db Database
}

func NewService(databaseInterface Database) *Service {
    return &Service{db: databaseInterface}
}

func (s *Service) GetMyRecord(ctx context.Context, id string) (MyRecord, error) {
    return s.db.GetMyRecord(ctx, id)
}
```

**Checklist:**

-   [ ] Define the domain struct(s) that this feature owns
-   [ ] Define sentinel errors (`var ErrNotFound = errors.New(...)`)
-   [ ] Define the `Database` interface with only the methods this feature calls
-   [ ] Add `//go:generate go tool mockery` at the top of the file
-   [ ] Implement the `Service` struct and `NewService` constructor
-   [ ] Implement Service methods that coordinate domain logic and call the `Database` interface
-   [ ] Write unit tests in `internal/services/services_test.go` using the generated `MockDatabase`
    -   Use concrete argument values in mock expectations (avoid `mock.Anything`)

---

## Step 5 – Define JSON views in `internal/handlers/views.go`

View types decouple the wire format from the domain model — the public API shape can evolve independently of internal domain changes, and vice versa.

**Pattern:**

```go
package handlers

import (
    "encoding/json"
    "github.com/specterops/bloodhound/server/<feature>/internal/services"
)

// MyRecordView is the JSON shape returned by the handler.
// It is separate from services.MyRecord so the wire format can change
// without touching the domain model.
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

**Checklist:**

-   [ ] Create a `XxxView` struct with `json:` tags for every response shape
-   [ ] Add a `BuildXxxView(domain services.Xxx) XxxView` function that projects the domain type to the view
-   [ ] Implement `JSONView() ([]byte, error)` on each view type to satisfy `responses.JSONViewer`

---

## Step 6 – Define the handler interface and methods in `internal/handlers/handlers.go`

The handler interface is defined here (consumer side) to enable independent mock substitution in tests.

**Pattern:**

```go
package handlers

//go:generate go tool mockery

import (
    "context"
    "errors"
    "net/http"

    "github.com/specterops/bloodhound/server/<feature>/internal/services"
    "github.com/specterops/bloodhound/server/responses"
)

// MyFeature is the service interface required by these handlers.
// Defined here (consumer side) to allow tests to swap in MockMyFeature.
type MyFeature interface {
    GetMyRecord(ctx context.Context, id string) (services.MyRecord, error)
}

type Handlers struct {
    feature MyFeature
}

func NewHandlersContainer(feature MyFeature) *Handlers {
    return &Handlers{feature: feature}
}

func (s Handlers) GetMyRecord(response http.ResponseWriter, request *http.Request) {
    var ctx = request.Context()

    // Extract ID from request (e.g., path parameter)
    id := /* ... */

    record, err := s.feature.GetMyRecord(ctx, id)
    if err != nil {
        handleMyFeatureError(ctx, err, response)
        return
    }

    responses.WriteBasic(ctx, BuildMyRecordView(record), http.StatusOK, response)
}

// handleMyFeatureError maps service-layer sentinels to HTTP status codes.
// Centralizes the error contract in one place.
func handleMyFeatureError(ctx context.Context, err error, response http.ResponseWriter) {
    if errors.Is(err, services.ErrNotFound) {
        responses.WriteError(ctx, http.StatusNotFound, "resource not found", response)
        return
    }
    if errors.Is(err, context.DeadlineExceeded) {
        responses.WriteError(ctx, http.StatusInternalServerError, "request timed out", response)
        return
    }
    responses.WriteInternalServerError(ctx, err, response)
}
```

**Checklist:**

-   [ ] Define the consumer-side `MyFeature` interface (only the service methods the handlers call)
-   [ ] Add `//go:generate go tool mockery` at the top of the file
-   [ ] Implement the `Handlers` struct and `NewHandlersContainer` constructor
-   [ ] Implement each `http.HandlerFunc` on `Handlers`:
    -   Extract values from the request
    -   Call the service via the interface
    -   Write the response using `responses.WriteBasic` or `responses.WriteError`
-   [ ] Add a package-level `handleXxxError(ctx, err, response)` helper that maps sentinel errors to HTTP status codes using `errors.Is`
-   [ ] Write unit tests in `internal/handlers/handlers_test.go` using the generated `MockMyFeature`
    -   Use `httptest.NewRecorder()` to capture responses
    -   Pass `request.Context()` to mock expectations (not `mock.Anything`)

---

## Step 7 – Register routes in `internal/routes/routes.go`

**Pattern:**

```go
package routes

import (
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/cmd/api/src/auth"
    "github.com/specterops/bloodhound/server/<feature>/internal/handlers"
)

// Register attaches the <feature> endpoints to the given router instance.
func Register(routerInst *router.Router, handlers *handlers.Handlers) {
    var permissions = auth.Permissions()

    routerInst.GET("/api/v2/<feature>/:id", handlers.GetMyRecord).
        RequirePermissions(permissions.AppReadApplicationConfiguration)
}
```

**Checklist:**

-   [ ] Call `routerInst.GET/PUT/DELETE(...)` with the correct path
-   [ ] Attach `.RequirePermissions(...)` or `.RequireAuth()` to enforce authorization
-   [ ] Write a route registration test in `internal/routes/routes_test.go` that:
    -   Asserts every method+path is registered using `muxRouter.Match`
    -   Dispatches unauthenticated requests and asserts 401 Unauthorized (validates middleware is attached)

---

## Step 8 – Wire all layers in `<feature>/<feature>.go`

The `Register` function is the single place where the feature module assembles its dependency chain and attaches routes. It takes only the infrastructure it directly needs from `modules.Deps`.

**Pattern:**

```go
package <feature>

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/server/<feature>/internal/appdb"
    "github.com/specterops/bloodhound/server/<feature>/internal/handlers"
    "github.com/specterops/bloodhound/server/<feature>/internal/routes"
    "github.com/specterops/bloodhound/server/<feature>/internal/services"
)

// Register builds the <feature> store → service → handler chain and attaches
// the <feature> routes to the provided router. It is called from the modules
// registry and receives only the infrastructure it directly needs.
func Register(routerInst *router.Router, pool *pgxpool.Pool) {
    var (
        store      = appdb.NewStore(pool)
        svc        = services.NewService(store)
        handlerSet = handlers.NewHandlersContainer(svc)
    )

    routes.Register(routerInst, handlerSet)
}
```

**Checklist:**

-   [ ] Import the internal packages: `internal/appdb`, `internal/handlers`, `internal/routes`, `internal/services`
-   [ ] Implement `Register(routerInst *router.Router, pool *pgxpool.Pool)` that chains `appdb → services → handlers → routes`
-   [ ] If the feature needs infrastructure beyond the router and pgx pool, accept additional parameters (then add them to `modules.Deps` in Step 9)

---

## Step 9 – Add to the module registry

Adding a feature to the system is a one-line change in the central registry.

**Pattern:**

```go
// server/modules/modules.go
import (
    "github.com/specterops/bloodhound/server/analysis"
    "github.com/specterops/bloodhound/server/<feature>"  // ← new
)

func Register(deps Deps) {
    if deps.Router == nil {
        panic("modules: Register requires a non-nil Router")
    }
    if deps.Pool == nil {
        panic("modules: Register requires a non-nil Pool")
    }

    analysis.Register(deps.Router, deps.Pool)
    <feature>.Register(deps.Router, deps.Pool)  // ← new
}
```

**If the feature needs new infrastructure:**

1. Add the field to the `Deps` struct in `server/modules/modules.go`:
    ```go
    type Deps struct {
        Router    *router.Router
        Pool      *pgxpool.Pool
        GraphDB   graph.Database  // ← new
    }
    ```
2. Populate it in both entrypoints that call `modules.Register`:
    - `lib/go/services/entrypoint.go` (open-source)
    - `bhce/cmd/api/src/services/entrypoint.go` (enterprise)
3. Accept it in your feature's `Register` signature and pass it through

**Checklist:**

-   [ ] Import the new feature package in `server/modules/modules.go`
-   [ ] Call `<feature>.Register(deps.Router, deps.Pool)` inside `modules.Register`
-   [ ] If the feature needs new infrastructure not in `Deps`, add it to the struct and populate it in both entrypoints

---

## Step 10 – Swap the e2e test to use the new handler stack

Now that the new module is registered, update the e2e test to use it instead of the old v2.Resources handler.

**Pattern:**

```go
// Before (old):
resources := v2.NewResources(db, /* ... */)
registration.NewV2API(resources, &routerInst)

// After (new):
modules.Register(modules.Deps{
    Router: &routerInst,
    Pool:   db.Pool(),
})
```

The test setup remains the same (production routing + global middleware + authentication), but the route registration switches from `v2.NewResources` to `modules.Register`.

**Checklist:**

-   [ ] Replace the old route registration (`registration.NewV2API(resources, ...)`) with `modules.Register(modules.Deps{...})`
-   [ ] Remove the old `v2.NewResources` setup (no longer needed)
-   [ ] Confirm the e2e test is still green with zero assertion changes: `go test -tags integration ./server/<feature>/...`
-   [ ] Verify authentication tests still pass (401 when unauthenticated, 200 when authenticated)

**What this validates:**

-   The new module is correctly wired into the global registry
-   Routes are registered with the correct paths and methods
-   Middleware (auth, rate limiting) is properly attached
-   The full request flow works end-to-end

---

## Step 11 – Remove the old handler and route registration

**For migrations only** (skip if this is a new endpoint):

-   [ ] Delete the old handler method from `v2.Resources` (or wherever it lived)
-   [ ] Remove the old route registration line from `cmd/api/src/api/registration/v2.go` (or equivalent)
-   [ ] Delete any now-unused helper functions or test stubs
-   [ ] Remove any now-unused imports
-   [ ] Confirm nothing references the deleted symbol: `go build ./...`

---

## Step 12 – Prepare for code review

`just prepare-for-codereview` runs the full suite of checks that CI will run. It must pass before opening a PR.

**What it does:**

-   Runs `just deps` (installs dependencies)
-   Runs `just modsync` (tidies Go modules)
-   Runs `just generate` (generates mocks, adds license headers, runs goimports, formats code, bundles OpenAPI docs)
-   Runs `just license` (checks all files have license headers)
-   Runs `just analysis` (runs golangci-lint and eslint)
-   Runs `just show` (validates repository is clean)

**Checklist:**

-   [ ] Run `just prepare-for-codereview` and fix any issues it reports
-   [ ] Confirm all tests pass: `go test ./server/<feature>/...` and `go test -tags integration ./server/<feature>/...`
-   [ ] Open a PR with the title format: `<conventional-commit-tag>: <Title> <Jira tag>`
    -   Example: `feat: add datapipe status endpoint BED-8715`

---

## Additional resources

-   **Architecture diagrams**: See [`bhce/server/README.md`](README.md#architecture-diagrams) for LikeC4 diagrams covering the C4 model views
-   **Code standards**: See [`bhce/AGENTS.md`](../../AGENTS.md) for Golang code standards and PR instructions
-   **Testing patterns**: See [`bhce/server/README.md`](README.md#testing) for details on pgxmock, mockery, and pgtestdb
-   **Mock generation**: See [`bhce/server/README.md`](README.md#mock-generation) for mockery configuration
