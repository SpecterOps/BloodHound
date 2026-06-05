# API Refactor Implementation Checklist

Use this checklist when creating a new endpoint or migrating an existing one to the new `server/` folder structure. Each step maps to a section in [`bhce/server/README.md`](README.md).

The pattern is **test-first migration**: write an integration test that documents the current endpoint's contract, complete the migration, then confirm the same test still passes — ideally with only the handler factory line changing.

**Commit after each step**

---

## Step 1 – Scaffold the feature directory

See README ["Adding a new feature module"](README.md#adding-a-new-feature-module) → sub-step 1.

-   [ ] Create the directory tree under `server/<feature>/`:
    ```
    server/<feature>/
    ├── <feature>.go        # Register() entry point
    ├── appdb/
    │   └── appdb.go
    ├── handlers/
    │   ├── handlers.go
    │   └── views.go
    ├── routes/
    │   └── routes.go
    └── services/
        └── services.go
    ```
-   [ ] Add the license header to every new file (run `just generate` or copy from an existing file)

---

## Step 2 – Write an e2e integration test against the existing endpoint

Before touching any production code, write a test that covers the HTTP contract of the endpoint in its current form. The goal is a green baseline that will still pass after migration.

-   [ ] Add a `TestXxx` function to `bhce/server/<feature>/<feature>_e2e_test.go`
-   [ ] Wire the existing (old) handler directly using the `v2.Resources` or equivalent struct, e.g.:
    ```go
    func newMyHandler(db *database.BloodhoundDB) http.HandlerFunc {
        return v2.Resources{DB: db}.MyHandler
    }
    ```
-   [ ] Assert every status code and relevant JSON field the endpoint can return (happy path + error paths)
-   [ ] Confirm the test is green: `go test -tags integration ./server/<feature>/...`

---

## Step 3 – Implement persistence in `appdb/appdb.go`

See README ["Step 3"](README.md#3-implement-persistence-in-appdbappdbgo).

-   [ ] Define a package-local `pgxQuerier` interface with only the pgx methods this store calls
-   [ ] Define a package-local row struct with `db:` tags for `pgx.RowToStructByName`
-   [ ] Implement `Store` methods using `sqlbuilder.PostgreSQL` for all SQL construction
-   [ ] Return services-layer sentinels (not driver errors) so callers can use `errors.Is` without importing `appdb`
-   [ ] Write unit tests in `appdb/appdb_test.go` using `pgxmock`
-   [ ] Write integration tests in `appdb/appdb_integration_test.go` using `pgtestdb` with `//go:build integration`

---

## Step 4 – Define domain types and interfaces in `services/services.go`

See README ["Step 2"](README.md#2-define-domain-types-and-interfaces-in-servicesservicesgo).

-   [ ] Define the domain struct(s) that this feature owns
-   [ ] Define sentinel errors (`var ErrNotFound = errors.New(...)`)
-   [ ] Define the `Database` interface (only the methods this feature calls)
-   [ ] Implement the Service methods on the `Service` struct that coordinate domain logic and call the `Database` interface
-   [ ] Implement a stub for the `Database` methods (no logic, no structs) to satisfy the interface in the service
-   [ ] Add `//go:generate go tool mockery` at the top of the file
-   [ ] Write unit tests in `services/services_test.go` using `MockDatabase` (AI can be helpful here)

---

## Step 5 – Define JSON views in `handlers/views.go`

See README ["Step 4"](README.md#4-define-json-views-in-handlersviewsgo).

-   [ ] Create a `XxxView` struct with `json:` tags for every response shape
-   [ ] Add a `BuildXxxView(domain services.Xxx) XxxView` function that projects the domain type to the view
-   [ ] Implement `JSONView() ([]byte, error)` to satisfy `responses.JSONViewer`

---

## Step 6 – Define the handler interface and methods in `handlers/handlers.go`

See README ["Step 5"](README.md#5-define-the-handler-interface-and-methods-in-handlershandlersgo).

-   [ ] Define the consumer-side `MyFeature` interface (only the service methods the handlers call)
-   [ ] Add `//go:generate go tool mockery` at the top of the file
-   [ ] Implement each `http.HandlerFunc` on `Handlers`:
    -   Extract values from the request
    -   Call the service via the interface
    -   Write the response using the `responses` package
-   [ ] Add a package-level `handleXxxError(request, response, err)` helper that maps sentinel errors to HTTP status codes using `errors.Is` — keeps every handler method DRY and makes the error contract visible in one place
-   [ ] Write unit tests in `handlers/handlers_test.go` using `MockMyFeature`

---

## Step 7 – Register routes in `routes/routes.go`

See README ["Step 6"](README.md#6-register-routes-in-handlersroutesgo).

-   [ ] Call `routerInst.GET/PUT/DELETE(...)` with the correct path and `.RequirePermissions(...)`
-   [ ] Write a routes test in `routes/routes_test.go` that asserts every method+path is registered

---

## Step 8 – Wire all layers in `<feature>/<feature>.go`

See README ["Step 7"](README.md#7-wire-all-layers-in-myfeaturemyfeaturego).

-   [ ] Implement `Register(routerInst *router.Router, pool *pgxpool.Pool)` that chains `appdb → services → handlers → routes`

---

## Step 9 – Add to the module registry

See README ["Step 8"](README.md#8-add-to-the-module-registry).

-   [ ] Import the new feature package in `server/modules/modules.go`
-   [ ] Call `<feature>.Register(deps.Router, deps.Pool)` inside `modules.Register`
-   [ ] If the feature needs new infrastructure, add it to the `Deps` struct first

---

## Step 11 – Swap the e2e test to use the new handler

-   [ ] Update the handler factory in the e2e test to wire the new stack instead of the old handler:

    ```go
    // Before (old):
    func newMyHandler(db *database.BloodhoundDB) http.HandlerFunc {
        return v2.Resources{DB: db}.MyHandler
    }

    // After (new):
    func newMyHandler(db *database.BloodhoundDB) http.HandlerFunc {
        store := appdb.NewStore(db.Pool())
        svc   := services.NewService(store)
        return handlers.NewHandlersContainer(svc).MyMethod
    }
    ```

-   [ ] Confirm the e2e test is still green (no assertion changes): `go test -tags integration ./cmd/api/src/services/...`

---

## Step 11 – Remove the old handler and route registration

-   [ ] Delete the old handler method from `v2.Resources` (or wherever it lived)
-   [ ] Remove the old route registration line from `cmd/api/src/api/registration/v2.go` (or equivalent)
-   [ ] Remove any now-unused imports
-   [ ] Confirm nothing references the deleted symbol: `go build ./...`

---

## Step 12 – Prepare for code review

-   [ ] Run `just prepare-for-codereview` (runs tests, generates mocks, adds license headers, generates OpenAPI docs)
-   [ ] Open a PR with the title format: `<conventional-commit-tag>: <Title> <Jira tag>`
