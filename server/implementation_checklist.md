# Implementation Checklist

Use this checklist when creating a new feature endpoint or migrating an existing endpoint to the `server/` module architecture.

The pattern is **test-first migration**: write an integration test that documents the current endpoint's contract, complete the migration, then confirm the same test still passes.

**Commit after each step.**

For architectural background and terminology, see [`bhce/server/README.md`](README.md).
For code patterns and examples, see the [Code Patterns Reference](#code-patterns-reference) section at the end of this document.

---

## Step 1 – Scaffold the feature directory

-   [ ] Create the directory tree under `server/<feature>/` (see structure below)
-   [ ] Add the license header to every new file (run `just generate` or copy from an existing file)

<details>
<summary>📁 Expected directory structure</summary>

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
</details>

---

## Step 2 – Write an e2e integration test against the existing endpoint

Before touching any production code, write a test that covers the HTTP contract of the endpoint in its current form. The goal is a green baseline that will still pass after migration.

-   [ ] Add `<feature>_e2e_test.go` with build tag `//go:build integration`
-   [ ] Wire the existing (old) handler using **production routing** (not manual mux setup)
-   [ ] Use `registration.RegisterFossGlobalMiddleware` to set up auth/rate limiting middleware
-   [ ] For migrations: use `v2.NewResources` + `registration.NewV2API` to register old routes
-   [ ] For new endpoints: create a minimal stub handler
-   [ ] Create authenticated requests with real JWT tokens using `api.Authenticator.CreateSession`
-   [ ] Assert every status code and relevant JSON field the endpoint can return (happy path + error paths)
-   [ ] Test authentication requirements (401 when unauthenticated)
-   [ ] Confirm the test is green: `go test -tags integration ./server/<feature>/...`

> **Why use production routing in tests:** This ensures the test catches wiring errors in route registration, middleware application, and authentication requirements — not just handler logic.

📖 See [E2E Test Pattern](#e2e-test-pattern) for code example.

---

## Step 3 – Implement persistence in `internal/appdb/appdb.go`

The persistence layer uses [go-sqlbuilder](https://github.com/huandu/go-sqlbuilder) for query construction and [pgx v5](https://github.com/jackc/pgx) for execution.

-   [ ] Define a package-local `pgxQuerier` interface with only the pgx methods this store calls
-   [ ] Define package-local row structs with `db:` tags for `pgx.RowToStructByName`
-   [ ] Define `toXxx` translator functions that convert row structs to domain types
-   [ ] Implement `Store` methods using `sqlbuilder.PostgreSQL` for all SQL construction
-   [ ] Use `pgx.CollectOneRow` + `pgx.RowToStructByName` for single-row queries
-   [ ] Use `pgx.CollectRows` + `pgx.RowToStructByName` for multi-row queries
-   [ ] Return services-layer sentinels (e.g., `services.ErrNotFound`) not driver errors
-   [ ] Write unit tests in `appdb_test.go` using [pgxmock](https://github.com/pashagolub/pgxmock)
-   [ ] Write integration tests in `appdb_integration_test.go` using [pgtestdb](https://github.com/peterldowns/pgtestdb) with `//go:build integration`

📖 See [Persistence Layer Pattern](#persistence-layer-pattern) for code example.

---

## Step 4 – Define domain types and interfaces in `internal/services/services.go`

The services package owns domain types and sentinel errors. The `Database` interface lives here (not in appdb) to enable Dependency Inversion.

-   [ ] Define the domain struct(s) that this feature owns
-   [ ] Define sentinel errors (`var ErrNotFound = errors.New(...)`)
-   [ ] Define the `Database` interface with only the methods this feature calls
-   [ ] Add `//go:generate go tool mockery` at the top of the file
-   [ ] Implement the `Service` struct and `NewService` constructor
-   [ ] Implement Service methods that coordinate domain logic and call the `Database` interface
-   [ ] Write unit tests in `services_test.go` using the generated `MockDatabase` (use concrete argument values, avoid `mock.Anything`)

📖 See [Services Layer Pattern](#services-layer-pattern) for code example.

---

## Step 5 – Define JSON views in `internal/handlers/views.go`

View types decouple the wire format from the domain model.

-   [ ] Create a `XxxView` struct with `json:` tags for every response shape
-   [ ] Add a `BuildXxxView(domain services.Xxx) XxxView` function that projects the domain type to the view
-   [ ] Implement `JSONView() ([]byte, error)` on each view type to satisfy `responses.JSONViewer`

📖 See [JSON Views Pattern](#json-views-pattern) for code example.

---

## Step 6 – Define the handler interface and methods in `internal/handlers/handlers.go`

The handler interface is defined here (consumer side) to enable independent mock substitution in tests.

-   [ ] Define the consumer-side `MyFeature` interface (only the service methods the handlers call)
-   [ ] Add `//go:generate go tool mockery` at the top of the file
-   [ ] Implement the `Handlers` struct and `NewHandlersContainer` constructor
-   [ ] Implement each `http.HandlerFunc` on `Handlers` (extract request values, call service, write response)
-   [ ] Add a package-level `handleXxxError(ctx, err, response)` helper that maps sentinel errors to HTTP status codes using `errors.Is`
-   [ ] Write unit tests in `handlers_test.go` using the generated `MockMyFeature`
    -   Use `httptest.NewRecorder()` to capture responses
    -   Pass `request.Context()` to mock expectations (not `mock.Anything`)

📖 See [Handlers Layer Pattern](#handlers-layer-pattern) for code example.

---

## Step 7 – Register routes in `internal/routes/routes.go`

-   [ ] Call `routerInst.GET/PUT/DELETE(...)` with the correct path
-   [ ] Attach `.RequirePermissions(...)` or `.RequireAuth()` to enforce authorization
-   [ ] Write a route registration test in `routes_test.go` that:
    -   Asserts every method+path is registered using `muxRouter.Match`
    -   Dispatches unauthenticated requests and asserts 401 Unauthorized (validates middleware is attached)

📖 See [Routes Pattern](#routes-pattern) for code example.

---

## Step 8 – Wire all layers in `<feature>/<feature>.go`

The `Register` function is the single place where the feature module assembles its dependency chain.

-   [ ] Import the internal packages: `internal/appdb`, `internal/handlers`, `internal/routes`, `internal/services`
-   [ ] Implement `Register(routerInst *router.Router, pool *pgxpool.Pool)` that chains `appdb → services → handlers → routes`
-   [ ] If the feature needs infrastructure beyond the router and pgx pool, accept additional parameters (then add them to `modules.Deps` in Step 9)

📖 See [Wireup Pattern](#wireup-pattern) for code example.

---

## Step 9 – Add to the module registry

Adding a feature to the system is a one-line change in the central registry.

-   [ ] Import the new feature package in `server/modules/modules.go`
-   [ ] Call `<feature>.Register(deps.Router, deps.Pool)` inside `modules.Register`
-   [ ] If the feature needs new infrastructure not in `Deps`, add it to the struct and populate it in both entrypoints (`lib/go/services/entrypoint.go` and `bhce/cmd/api/src/services/entrypoint.go`)

📖 See [Module Registry Pattern](#module-registry-pattern) for code example.

---

## Step 10 – Swap the e2e test to use the new handler stack

Now that the new module is registered, update the e2e test to use it instead of the old v2.Resources handler.

-   [ ] Replace the old route registration (`registration.NewV2API(resources, ...)`) with `modules.Register(modules.Deps{...})`
-   [ ] Remove the old `v2.NewResources` setup (no longer needed)
-   [ ] Confirm the e2e test is still green with zero assertion changes: `go test -tags integration ./server/<feature>/...`
-   [ ] Verify authentication tests still pass (401 when unauthenticated, 200 when authenticated)

> **What this validates:** The new module is correctly wired into the global registry, routes are registered with the correct paths and methods, middleware is properly attached, and the full request flow works end-to-end.

📖 See [E2E Test Swap Pattern](#e2e-test-swap-pattern) for code example.

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

-   [ ] Run `just prepare-for-codereview` and fix any issues it reports
-   [ ] Confirm all tests pass: `go test ./server/<feature>/...` and `go test -tags integration ./server/<feature>/...`
-   [ ] Open a PR with the title format: `<conventional-commit-tag>: <Title> <Jira tag>`
    -   Example: `feat: add datapipe status endpoint BED-8715`

> **What `just prepare-for-codereview` does:** Runs `just deps`, `just modsync`, `just generate` (generates mocks, adds license headers, runs goimports, formats code, bundles OpenAPI docs), `just license`, `just analysis`, and `just show`.

---

## Additional Resources

-   **Architecture diagrams**: See [`bhce/server/README.md`](README.md#architecture-diagrams) for LikeC4 diagrams covering the C4 model views
-   **Code standards**: See [`bhce/AGENTS.md`](../../AGENTS.md) for Golang code standards and PR instructions
-   **Testing patterns**: See [`bhce/server/README.md`](README.md#testing) for details on pgxmock, mockery, and pgtestdb
-   **Mock generation**: See [`bhce/server/README.md`](README.md#mock-generation) for mockery configuration

---

# Code Patterns Reference

This section contains complete code examples for each step in the checklist. Use these as templates when implementing your feature.

## E2E Test Pattern

**File:** `server/<feature>/<feature>_e2e_test.go`

```go
//go:build integration

package appcfg_test

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/specterops/bloodhound/cmd/api/src/api"
    "github.com/specterops/bloodhound/cmd/api/src/api/registration"
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
    "github.com/specterops/bloodhound/cmd/api/src/auth"
    "github.com/specterops/bloodhound/cmd/api/src/config"
    "github.com/specterops/bloodhound/src/test/integration"
    "github.com/stretchr/testify/require"
)

func TestGetMyEndpoint_WithRouting(t *testing.T) {
    var (
        db         = integration.OpenDatabase(t)
        cfg        = config.Configuration{}
        authorizer = auth.NewAuthorizer(db)
        routerInst = router.NewRouter(cfg, authorizer, "")
    )

    // Register global middleware (required for auth to work)
    registration.RegisterFossGlobalMiddleware(&routerInst, cfg, nil, nil, db)

    // Register old v2 routes using v2.Resources
    resources := v2.NewResources(db, /* ... */)
    registration.NewV2API(resources, &routerInst)

    // Start test server with production routing
    handler := routerInst.Handler()
    server := httptest.NewServer(handler)
    defer server.Close()

    // Create authenticated session
    authenticator := api.NewAuthenticator(cfg.CryptoConfiguration(), db)
    user := integration.NewUser(t, db, "test-user", "Test User", auth.RoleAdministrator)
    token, err := authenticator.CreateSession(context.Background(), user)
    require.NoError(t, err)

    t.Run("returns 401 when unauthenticated", func(t *testing.T) {
        req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v2/my-endpoint", nil)
        require.NoError(t, err)

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()

        require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })

    t.Run("returns 200 with valid data", func(t *testing.T) {
        req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v2/my-endpoint", nil)
        require.NoError(t, err)
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()

        require.Equal(t, http.StatusOK, resp.StatusCode)

        var result MyResponseView
        err = json.NewDecoder(resp.Body).Decode(&result)
        require.NoError(t, err)
        require.NotEmpty(t, result.ID)
    })
}
```

## Persistence Layer Pattern

**File:** `server/<feature>/internal/appdb/appdb.go`

```go
// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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

// Store performs <feature> persistence operations directly against a PostgreSQL
// connection. Callers receive appdb-level sentinels rather than raw driver errors.
type Store struct {
    db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx connection pool.
func NewStore(db pgxQuerier) *Store {
    return &Store{db: db}
}

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

**File:** `server/<feature>/internal/appdb/appdb_test.go`

```go
package appdb_test

import (
    "context"
    "testing"

    "github.com/pashagolub/pgxmock/v3"
    "github.com/specterops/bloodhound/server/<feature>/internal/appdb"
    "github.com/stretchr/testify/require"
)

func TestStore_GetMyRecord(t *testing.T) {
    mock, err := pgxmock.NewPool()
    require.NoError(t, err)
    defer mock.Close()

    store := appdb.NewStore(mock)

    t.Run("returns record when found", func(t *testing.T) {
        rows := pgxmock.NewRows([]string{"id", "name"}).
            AddRow("abc123", "test-name")

        mock.ExpectQuery(`SELECT id, name FROM my_table WHERE id = \$1`).
            WithArgs("abc123").
            WillReturnRows(rows)

        record, err := store.GetMyRecord(context.Background(), "abc123")
        require.NoError(t, err)
        require.Equal(t, "abc123", record.ID)
        require.Equal(t, "test-name", record.Name)
        require.NoError(t, mock.ExpectationsWereMet())
    })
}
```

## Services Layer Pattern

**File:** `server/<feature>/internal/services/services.go`

```go
// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package services

//go:generate go tool mockery

import (
    "context"
    "errors"
)

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

// Database describes the persistence capabilities the <feature> Service requires.
// Implementations are expected to translate driver-specific not-found errors into
// appdb-level sentinels so that the Service can map them to its own failure-mode errors.
type Database interface {
    GetMyRecord(ctx context.Context, id string) (MyRecord, error)
}

// Service implements the <feature> use cases on top of a Database implementation.
type Service struct {
    db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
    return &Service{db: databaseInterface}
}

func (s *Service) GetMyRecord(ctx context.Context, id string) (MyRecord, error) {
    return s.db.GetMyRecord(ctx, id)
}
```

**File:** `server/<feature>/internal/services/services_test.go`

```go
package services_test

import (
    "context"
    "testing"

    "github.com/specterops/bloodhound/server/<feature>/internal/services"
    "github.com/specterops/bloodhound/server/<feature>/internal/services/mocks"
    "github.com/stretchr/testify/require"
)

func TestService_GetMyRecord(t *testing.T) {
    mockDB := mocks.NewMockDatabase(t)
    svc := services.NewService(mockDB)
    ctx := context.Background()

    expectedRecord := services.MyRecord{ID: "abc123", Name: "test"}

    mockDB.EXPECT().
        GetMyRecord(ctx, "abc123").
        Return(expectedRecord, nil)

    record, err := svc.GetMyRecord(ctx, "abc123")
    require.NoError(t, err)
    require.Equal(t, expectedRecord, record)
}
```

## JSON Views Pattern

**File:** `server/<feature>/internal/handlers/views.go`

```go
// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
    "encoding/json"
    "github.com/specterops/bloodhound/server/<feature>/internal/services"
)

// MyRecordView is the JSON shape returned by the <feature> handlers.
// It is decoupled from services.MyRecord so the wire format can evolve
// independently of the domain model.
type MyRecordView struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// BuildMyRecordView projects a services.MyRecord into the view type the
// handlers return in their JSON envelope.
func BuildMyRecordView(r services.MyRecord) MyRecordView {
    return MyRecordView{ID: r.ID, Name: r.Name}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s MyRecordView) JSONView() ([]byte, error) {
    return json.Marshal(s)
}
```

## Handlers Layer Pattern

**File:** `server/<feature>/internal/handlers/handlers.go`

```go
// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package handlers

//go:generate go tool mockery

import (
    "context"
    "errors"
    "net/http"

    "github.com/specterops/bloodhound/packages/go/responses"
    "github.com/specterops/bloodhound/server/<feature>/internal/services"
)

// MyFeature defines the <feature> service boundary for the <feature> handlers package.
// Defined here (consumer side) to allow tests to swap in MockMyFeature.
type MyFeature interface {
    GetMyRecord(ctx context.Context, id string) (services.MyRecord, error)
}

// Handlers is a dependency injection container for <feature> handlers
type Handlers struct {
    feature MyFeature
}

// NewHandlersContainer initializes the Handlers dependency injection container
func NewHandlersContainer(feature MyFeature) *Handlers {
    return &Handlers{
        feature: feature,
    }
}

func (s Handlers) GetMyRecord(response http.ResponseWriter, request *http.Request) {
    var ctx = request.Context()

    // Extract ID from request (e.g., path parameter via gorilla/mux)
    // id := mux.Vars(request)["id"]

    record, err := s.feature.GetMyRecord(ctx, "example-id")
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

**File:** `server/<feature>/internal/handlers/handlers_test.go`

```go
package handlers_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/specterops/bloodhound/server/<feature>/internal/handlers"
    "github.com/specterops/bloodhound/server/<feature>/internal/handlers/mocks"
    "github.com/specterops/bloodhound/server/<feature>/internal/services"
    "github.com/stretchr/testify/require"
)

func TestHandlers_GetMyRecord(t *testing.T) {
    mockFeature := mocks.NewMockMyFeature(t)
    h := handlers.NewHandlersContainer(mockFeature)

    t.Run("returns 200 with record", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/api/v2/my-record", nil)
        rec := httptest.NewRecorder()

        expectedRecord := services.MyRecord{ID: "abc123", Name: "test"}
        mockFeature.EXPECT().
            GetMyRecord(req.Context(), "example-id").
            Return(expectedRecord, nil)

        h.GetMyRecord(rec, req)

        require.Equal(t, http.StatusOK, rec.Code)
    })

    t.Run("returns 404 when not found", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/api/v2/my-record", nil)
        rec := httptest.NewRecorder()

        mockFeature.EXPECT().
            GetMyRecord(req.Context(), "example-id").
            Return(services.MyRecord{}, services.ErrNotFound)

        h.GetMyRecord(rec, req)

        require.Equal(t, http.StatusNotFound, rec.Code)
    })
}
```

## Routes Pattern

**File:** `server/<feature>/internal/routes/routes.go`

```go
// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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

**File:** `server/<feature>/internal/routes/routes_test.go`

```go
package routes_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gorilla/mux"
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/cmd/api/src/auth"
    "github.com/specterops/bloodhound/cmd/api/src/config"
    "github.com/specterops/bloodhound/server/<feature>/internal/handlers"
    "github.com/specterops/bloodhound/server/<feature>/internal/handlers/mocks"
    "github.com/specterops/bloodhound/server/<feature>/internal/routes"
    "github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
    var (
        cfg        = config.Configuration{}
        authorizer = auth.NewAuthorizer(nil)
        routerInst = router.NewRouter(cfg, authorizer, "")
        mockSvc    = mocks.NewMockMyFeature(t)
        h          = handlers.NewHandlersContainer(mockSvc)
    )

    routes.Register(&routerInst, h)

    t.Run("GET /api/v2/<feature>/:id is registered", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/api/v2/<feature>/123", nil)
        var routeMatch mux.RouteMatch

        matched := routerInst.Match(req, &routeMatch)
        require.True(t, matched, "Route should be registered")
    })

    t.Run("returns 401 when unauthenticated", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/api/v2/<feature>/123", nil)
        rec := httptest.NewRecorder()

        routerInst.ServeHTTP(rec, req)

        require.Equal(t, http.StatusUnauthorized, rec.Code)
    })
}
```

## Wireup Pattern

**File:** `server/<feature>/<feature>.go`

```go
// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package <feature> is the wireup module for the <feature> feature. It is the
// single place where the <feature> store, service, handlers and routes are
// composed; the layered subpackages themselves remain unaware of each other.
package <feature>

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/server/<feature>/internal/appdb"
    "github.com/specterops/bloodhound/server/<feature>/internal/handlers"
    "github.com/specterops/bloodhound/server/<feature>/internal/routes"
    "github.com/specterops/bloodhound/server/<feature>/internal/services"
)

// Register builds the <feature> store -> service -> handler chain and attaches
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

## Module Registry Pattern

**File:** `server/modules/modules.go`

```go
package modules

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/specterops/bloodhound/cmd/api/src/api/router"
    "github.com/specterops/bloodhound/server/analysis"
    "github.com/specterops/bloodhound/server/<feature>"  // ← new
)

// Deps holds the infrastructure dependencies required by all modules.
type Deps struct {
    Router *router.Router
    Pool   *pgxpool.Pool
}

// Register wires all server modules into the application.
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

## E2E Test Swap Pattern

**Before (old):**

```go
// Register old v2 routes using v2.Resources
resources := v2.NewResources(db, /* ... */)
registration.NewV2API(resources, &routerInst)
```

**After (new):**

```go
// Register all modules including the new feature
modules.Register(modules.Deps{
    Router: &routerInst,
    Pool:   db.Pool(),
})
```

The test setup remains the same (production routing + global middleware + authentication), but the route registration switches from `v2.NewResources` to `modules.Register`.
