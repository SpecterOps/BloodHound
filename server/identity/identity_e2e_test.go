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

//go:build integration

package identity_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/appdb"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIdentityDB creates an isolated test database with all migrations applied.
// The database is automatically closed when the test ends.
func setupIdentityDB(t *testing.T) *database.BloodhoundDB {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getIdentityPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, db.Migrate(ctx))

	t.Cleanup(func() { db.Close(ctx) })

	return db
}

// getIdentityPostgresConfig reads the integration test configuration from the
// environment and returns a pgtestdb.Config for the identity e2e tests.
func getIdentityPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for entry := range strings.FieldsSeq(cfg.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

// newIdentityHandlers wires the identity store -> service -> handlers chain
// backed by the given database.
func newIdentityHandlers(db *database.BloodhoundDB) *handlers.Handlers {
	var (
		store      = appdb.NewStore(db.Pool())
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)
	return handlerSet
}

// permissionResponseEnvelope is the JSON envelope shape returned by the
// GET /api/v2/permissions/{permission_id} handler.
type permissionResponseEnvelope struct {
	Data model.Permission `json:"data"`
}

// roleResponseEnvelope is the JSON envelope shape returned by the
// GET /api/v2/roles/{role_id} handler.
type roleResponseEnvelope struct {
	Data model.Role `json:"data"`
}

// listRolesResponseEnvelope is the JSON envelope shape returned by the
// GET /api/v2/roles handler.
type listRolesResponseEnvelope struct {
	Data struct {
		Roles model.Roles `json:"roles"`
	} `json:"data"`
}

// listPermissionsResponseEnvelope is the JSON envelope shape returned by the
// GET /api/v2/permissions handler.
type listPermissionsResponseEnvelope struct {
	Data struct {
		Permissions model.Permissions `json:"permissions"`
	} `json:"data"`
}

// listUsersResponseEnvelope is the JSON envelope shape returned by the
// GET /api/v2/bloodhound-users handler.
type listUsersResponseEnvelope struct {
	Data struct {
		Users []struct {
			PrincipalName string  `json:"principal_name"`
			EmailAddress  *string `json:"email_address"`
			FirstName     *string `json:"first_name"`
		} `json:"users"`
	} `json:"data"`
}

// newListRolesHandler wires the identity slice's ListRoles handler backed by
// the given database, wrapped in the same filter and sort middleware the route
// applies in production (see routes.Register). The middleware parses and
// validates the query parameters into the BloodHound context, so this exercises
// the full request path the handler relies on.
func newListRolesHandler(db *database.BloodhoundDB) http.HandlerFunc {
	var (
		handlerSet = newIdentityHandlers(db)
		roleList   = handlers.RoleListView{}
		handler    = middleware.FilterMiddleware(roleList)(
			middleware.SortMiddleware(roleList)(http.HandlerFunc(handlerSet.ListRoles)),
		)
	)

	return handler.ServeHTTP
}

// newListPermissionsHandler wires the identity slice's ListPermissions handler
// through the production filter and sort middleware.
func newListPermissionsHandler(db *database.BloodhoundDB) http.HandlerFunc {
	var (
		handlerSet     = newIdentityHandlers(db)
		permissionList = handlers.PermissionListView{}
		handler        = middleware.FilterMiddleware(permissionList)(
			middleware.SortMiddleware(permissionList)(http.HandlerFunc(handlerSet.ListPermissions)),
		)
	)

	return handler.ServeHTTP
}

// newListUsersHandler wires the identity slice's ListUsers handler through the
// production sort and filter middleware. Sort is the outer middleware so that an
// invalid sort is rejected before filter validation, matching the legacy handler.
func newListUsersHandler(db *database.BloodhoundDB) http.HandlerFunc {
	var (
		handlerSet = newIdentityHandlers(db)
		userList   = handlers.UserListView{}
		handler    = middleware.SortMiddleware(userList)(
			middleware.FilterMiddleware(userList)(http.HandlerFunc(handlerSet.ListUsers)),
		)
	)

	return handler.ServeHTTP
}

// seedUser inserts a user row directly via the pool with every non-defaulted
// column populated so the strict pgx scanners in ListUsers do not encounter
// NULLs. It returns the seeded principal name for convenience.
func seedUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, principalName string, supportAccount bool) string {
	t.Helper()

	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, principal_name, first_name, last_name, email_address, last_login, is_disabled, all_environments, eula_accepted, support_account, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, now(), false, true, false, $5, now(), now())`,
		principalName, principalName+"-first", principalName+"-last", principalName+"@example.com", supportAccount,
	)
	require.NoError(t, err)

	return principalName
}

func TestIdentity_GetPermission(t *testing.T) {
	type mock struct {
		handler http.Handler
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		expected     expected
	}

	var (
		db          = setupIdentityDB(t)
		ctx         = context.Background()
		handlerSet  = newIdentityHandlers(db)
		muxRouter   = mux.NewRouter()
		permissions []services.Permission
		err         error
	)

	permissions, err = appdb.NewStore(db.Pool()).ListPermissions(ctx, params.Filters{}, params.SortItems{})
	require.NoError(t, err)
	require.NotEmpty(t, permissions, "expected migrations to seed at least one permission")
	seededPermission := permissions[0]
	muxRouter.HandleFunc("/api/v2/permissions/{permission_id}", handlerSet.GetPermission).Methods(http.MethodGet)
	testMock := mock{handler: muxRouter}

	tt := []testData{
		{
			name: "Success: seeded permission is returned - 200",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/"+fmt.Sprintf("%d", seededPermission.ID), nil)
			},
			expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
				var envelope permissionResponseEnvelope
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, seededPermission.ID, envelope.Data.ID)
				assert.Equal(t, seededPermission.Authority, envelope.Data.Authority)
				assert.Equal(t, seededPermission.Name, envelope.Data.Name)
				assert.True(t, seededPermission.CreatedAt.Equal(envelope.Data.CreatedAt), "created_at should match the seeded permission")
				assert.True(t, seededPermission.UpdatedAt.Equal(envelope.Data.UpdatedAt), "updated_at should match the seeded permission")
				assert.Equal(t, seededPermission.DeletedAt.Valid, envelope.Data.DeletedAt.Valid, "deleted_at validity should match the seeded permission")
			}},
		},
		{name: "Error: permission does not exist - 404", buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/99999999", nil) }, expected: expected{responseCode: http.StatusNotFound, responseHeader: http.Header{"Content-Type": []string{"application/json"}}}},
		{name: "Error: permission ID is malformed - 400", buildRequest: func() *http.Request {
			return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/not-an-int", nil)
		}, expected: expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			testMock.handler.ServeHTTP(recorder, testCase.buildRequest())

			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
			assert.Equal(t, testCase.expected.responseHeader, recorder.Result().Header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestIdentity_GetRole(t *testing.T) {
	type mock struct {
		handler http.Handler
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		expected     expected
	}

	var (
		db         = setupIdentityDB(t)
		ctx        = context.Background()
		handlerSet = newIdentityHandlers(db)
		muxRouter  = mux.NewRouter()
		roles      model.Roles
		err        error
	)

	roles, err = db.GetAllRoles(ctx, "", model.SQLFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, roles, "expected migrations to seed at least one role")
	seededRole := roles[0]
	muxRouter.HandleFunc("/api/v2/roles/{role_id}", handlerSet.GetRole).Methods(http.MethodGet)
	testMock := mock{handler: muxRouter}

	tt := []testData{
		{
			name: "Success: seeded role is returned - 200",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/roles/"+fmt.Sprintf("%d", seededRole.ID), nil)
			},
			expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
				var envelope roleResponseEnvelope
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, seededRole.ID, envelope.Data.ID)
				assert.Equal(t, seededRole.Name, envelope.Data.Name)
				assert.NotEmpty(t, envelope.Data.Permissions, "expected the role to preload its permissions")
				assert.True(t, seededRole.CreatedAt.Equal(envelope.Data.CreatedAt), "created_at should match the seeded role")
				assert.True(t, seededRole.UpdatedAt.Equal(envelope.Data.UpdatedAt), "updated_at should match the seeded role")
				assert.Equal(t, seededRole.DeletedAt.Valid, envelope.Data.DeletedAt.Valid, "deleted_at validity should match the seeded role")
			}},
		},
		{name: "Error: role does not exist - 404", buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/roles/99999999", nil) }, expected: expected{responseCode: http.StatusNotFound, responseHeader: http.Header{"Content-Type": []string{"application/json"}}}},
		{name: "Error: role ID is malformed - 400", buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/roles/not-an-int", nil) }, expected: expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			testMock.handler.ServeHTTP(recorder, testCase.buildRequest())

			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
			assert.Equal(t, testCase.expected.responseHeader, recorder.Result().Header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestIdentity_ListRoles(t *testing.T) {
	type mock struct {
		handler http.Handler
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func(t *testing.T) *http.Request
		expected     expected
	}

	var (
		db      = setupIdentityDB(t)
		ctx     = context.Background()
		handler = newListRolesHandler(db)
		roles   model.Roles
		err     error
	)

	roles, err = db.GetAllRoles(ctx, "name", model.SQLFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, roles, "expected migrations to seed at least one role")
	seededRole := roles[0]
	testMock := mock{handler: http.HandlerFunc(handler)}

	newRequest := func(t *testing.T, query url.Values) *http.Request {
		t.Helper()
		reqCtx := context.WithValue(ctx, bhctx.ValueKey, &bhctx.Context{})
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "/api/v2/roles", nil)
		require.NoError(t, err)
		req.URL.RawQuery = query.Encode()
		return req
	}

	tt := []testData{
		{name: "Success: all seeded roles are returned - 200", buildRequest: func(t *testing.T) *http.Request { return newRequest(t, url.Values{}) }, expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
			var envelope listRolesResponseEnvelope
			require.NoError(t, json.Unmarshal(body, &envelope))
			require.Len(t, envelope.Data.Roles, len(roles))
			assert.NotEmpty(t, envelope.Data.Roles[0].Permissions, "expected roles to preload their permissions")
		}}},
		{name: "Success: roles are sorted by name - 200", buildRequest: func(t *testing.T) *http.Request { return newRequest(t, url.Values{"sort_by": {"name"}}) }, expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
			var envelope listRolesResponseEnvelope
			require.NoError(t, json.Unmarshal(body, &envelope))
			require.Len(t, envelope.Data.Roles, len(roles))
			for i := 1; i < len(envelope.Data.Roles); i++ {
				assert.LessOrEqual(t, envelope.Data.Roles[i-1].Name, envelope.Data.Roles[i].Name, "roles should be sorted by name ascending")
			}
		}}},
		{name: "Success: roles are filtered by name - 200", buildRequest: func(t *testing.T) *http.Request { return newRequest(t, url.Values{"name": {"eq:" + seededRole.Name}}) }, expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
			var envelope listRolesResponseEnvelope
			require.NoError(t, json.Unmarshal(body, &envelope))
			require.Len(t, envelope.Data.Roles, 1)
			assert.Equal(t, seededRole.Name, envelope.Data.Roles[0].Name)
		}}},
		{name: "Error: sort column is not supported - 400", buildRequest: func(t *testing.T) *http.Request { return newRequest(t, url.Values{"sort_by": {"invalidColumn"}}) }, expected: expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) { assert.Contains(t, string(body), api.ErrorResponseDetailsNotSortable) }}},
		{name: "Error: filter column is not supported - 400", buildRequest: func(t *testing.T) *http.Request { return newRequest(t, url.Values{"foo": {"eq:bar"}}) }, expected: expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
			assert.Contains(t, string(body), api.ErrorResponseDetailsColumnNotFilterable)
		}}},
		{name: "Error: filter predicate is malformed - 400", buildRequest: func(t *testing.T) *http.Request { return newRequest(t, url.Values{"name": {"invalidPredicate:foo"}}) }, expected: expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
			assert.Contains(t, string(body), api.ErrorResponseDetailsBadQueryParameterFilters)
		}}},
		{name: "Error: filter predicate is not supported - 400", buildRequest: func(t *testing.T) *http.Request { return newRequest(t, url.Values{"name": {"gt:0"}}) }, expected: expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
			assert.Contains(t, string(body), api.ErrorResponseDetailsFilterPredicateNotSupported)
		}}},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			testMock.handler.ServeHTTP(recorder, testCase.buildRequest(t))
			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
			assert.Equal(t, testCase.expected.responseHeader, recorder.Result().Header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestIdentity_ListPermissions(t *testing.T) {
	type mock struct {
		handler http.Handler
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func(t *testing.T) *http.Request
		expected     expected
	}

	var (
		db          = setupIdentityDB(t)
		ctx         = context.Background()
		handler     = newListPermissionsHandler(db)
		permissions []services.Permission
		err         error
	)

	permissions, err = appdb.NewStore(db.Pool()).ListPermissions(ctx, params.Filters{}, params.SortItems{{Field: "name", Direction: params.Ascending}})
	require.NoError(t, err)
	require.NotEmpty(t, permissions, "expected migrations to seed at least one permission")
	seededPermission := permissions[0]
	testMock := mock{handler: http.HandlerFunc(handler)}

	newRequest := func(t *testing.T, query url.Values) *http.Request {
		t.Helper()
		reqCtx := context.WithValue(ctx, bhctx.ValueKey, &bhctx.Context{})
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "/api/v2/permissions", nil)
		require.NoError(t, err)
		req.URL.RawQuery = query.Encode()
		return req
	}

	tt := []testData{
		{
			name: "Success: all seeded permissions are returned - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					var envelope listPermissionsResponseEnvelope
					require.NoError(t, json.Unmarshal(body, &envelope))
					assert.Len(t, envelope.Data.Permissions, len(permissions))
				},
			},
		},
		{
			name: "Success: permissions are sorted by name - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"sort_by": {"name"}})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					var envelope listPermissionsResponseEnvelope
					require.NoError(t, json.Unmarshal(body, &envelope))
					require.Len(t, envelope.Data.Permissions, len(permissions))
					for i := 1; i < len(envelope.Data.Permissions); i++ {
						assert.LessOrEqual(t, envelope.Data.Permissions[i-1].Name, envelope.Data.Permissions[i].Name, "permissions should be sorted by name ascending")
					}
				},
			},
		},
		{
			name: "Success: permissions are filtered by authority - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"authority": {"eq:" + seededPermission.Authority}})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					var envelope listPermissionsResponseEnvelope
					require.NoError(t, json.Unmarshal(body, &envelope))
					require.NotEmpty(t, envelope.Data.Permissions)
					for _, permission := range envelope.Data.Permissions {
						assert.Equal(t, seededPermission.Authority, permission.Authority)
					}
				},
			},
		},
		{
			name: "Success: no permissions match - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"name": {"eq:does-not-exist"}})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					var envelope listPermissionsResponseEnvelope
					require.NoError(t, json.Unmarshal(body, &envelope))
					assert.Empty(t, envelope.Data.Permissions)
				},
			},
		},
		{
			name: "Error: sort column is not supported - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"sort_by": {"invalidColumn"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsNotSortable)
				},
			},
		},
		{
			name: "Error: filter column is not supported - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"foo": {"eq:bar"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsColumnNotFilterable)
				},
			},
		},
		{
			name: "Error: filter predicate is malformed - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"name": {"invalidPredicate:foo"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
		},
		{
			name: "Error: filter predicate is not supported - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"authority": {"gt:app"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsFilterPredicateNotSupported)
				},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			testMock.handler.ServeHTTP(recorder, testCase.buildRequest(t))

			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
			assert.Equal(t, testCase.expected.responseHeader, recorder.Result().Header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestIdentity_ListUsers(t *testing.T) {
	type mock struct {
		handler http.Handler
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func(t *testing.T) *http.Request
		expected     expected
	}

	var (
		db      = setupIdentityDB(t)
		ctx     = context.Background()
		handler = newListUsersHandler(db)
	)

	// Seed two regular users and one support account. The support account must
	// never appear in the response, mirroring the legacy support_account = false
	// filter the migrated handler applies.
	userA := seedUser(t, ctx, db.Pool(), "user-a", false)
	userB := seedUser(t, ctx, db.Pool(), "user-b", false)
	supportUser := seedUser(t, ctx, db.Pool(), "support", true)

	testMock := mock{handler: http.HandlerFunc(handler)}

	newRequest := func(t *testing.T, query url.Values) *http.Request {
		t.Helper()
		reqCtx := context.WithValue(ctx, bhctx.ValueKey, &bhctx.Context{})
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "/api/v2/bloodhound-users", nil)
		require.NoError(t, err)
		req.URL.RawQuery = query.Encode()
		return req
	}

	principalsIn := func(envelope listUsersResponseEnvelope) []string {
		names := make([]string, 0, len(envelope.Data.Users))
		for _, user := range envelope.Data.Users {
			names = append(names, user.PrincipalName)
		}
		return names
	}

	tt := []testData{
		{
			name: "Success: regular users are returned and support accounts excluded - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					var envelope listUsersResponseEnvelope
					require.NoError(t, json.Unmarshal(body, &envelope))

					names := principalsIn(envelope)
					assert.Contains(t, names, userA)
					assert.Contains(t, names, userB)
					assert.NotContains(t, names, supportUser, "support accounts must be excluded from the list")
				},
			},
		},
		{
			name: "Success: users are sorted by principal_name - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"sort_by": {"principal_name"}})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					var envelope listUsersResponseEnvelope
					require.NoError(t, json.Unmarshal(body, &envelope))
					for i := 1; i < len(envelope.Data.Users); i++ {
						assert.LessOrEqual(t, envelope.Data.Users[i-1].PrincipalName, envelope.Data.Users[i].PrincipalName, "users should be sorted by principal_name ascending")
					}
				},
			},
		},
		{
			name: "Success: users are filtered by email_address - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"email_address": {"eq:" + userA + "@example.com"}})
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					var envelope listUsersResponseEnvelope
					require.NoError(t, json.Unmarshal(body, &envelope))
					require.Len(t, envelope.Data.Users, 1)
					assert.Equal(t, userA, envelope.Data.Users[0].PrincipalName)
				},
			},
		},
		{
			name: "Error: sort column is not supported - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"sort_by": {"invalidColumn"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsNotSortable)
				},
			},
		},
		{
			name: "Error: deleted_at sort column is not supported - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"sort_by": {"deleted_at"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsNotSortable)
				},
			},
		},
		{
			name: "Error: filter column is not supported - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"foo": {"eq:bar"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsColumnNotFilterable)
				},
			},
		},
		{
			name: "Error: filter predicate is malformed - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"email_address": {"invalidPredicate:foo"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsBadQueryParameterFilters)
				},
			},
		},
		{
			name: "Error: filter predicate is not supported - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"first_name": {"gt:0"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsFilterPredicateNotSupported)
				},
			},
		},
		{
			name: "Error: sort error takes precedence over filter error - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return newRequest(t, url.Values{"sort_by": {"invalidColumn"}, "foo": {"eq:bar"}})
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				assertBody: func(t *testing.T, body []byte) {
					assert.Contains(t, string(body), api.ErrorResponseDetailsNotSortable)
					assert.NotContains(t, string(body), api.ErrorResponseDetailsColumnNotFilterable)
				},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			testMock.handler.ServeHTTP(recorder, testCase.buildRequest(t))

			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
			assert.Equal(t, testCase.expected.responseHeader, recorder.Result().Header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}
