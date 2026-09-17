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

// newListRolesHandler wires the identity slice's ListRoles handler backed by
// the given database, wrapped in the same filter and sort middleware the route
// applies in production (see routes.Register). The middleware parses and
// validates the query parameters into the BloodHound context, so this exercises
// the full request path the handler relies on.
func newListRolesHandler(db *database.BloodhoundDB) http.HandlerFunc {
	var (
		handlerSet = newIdentityHandlers(db)
		roleList   = handlers.RoleListView{}
		// Sort wraps filter (sort runs first) to mirror the production route's
		// WithSort-before-WithFilters ordering (see routes.Register).
		handler = middleware.SortMiddleware(roleList)(
			middleware.FilterMiddleware(roleList)(http.HandlerFunc(handlerSet.ListRoles)),
		)
	)

	return handler.ServeHTTP
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

func TestGetPermission(t *testing.T) {
	var (
		db          = setupIdentityDB(t)
		ctx         = context.Background()
		handlerSet  = newIdentityHandlers(db)
		handler     = handlerSet.GetPermission
		permissions model.Permissions
		err         error
	)

	permissions, err = db.GetAllPermissions(ctx, "", model.SQLFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, permissions, "expected migrations to seed at least one permission")
	seededPermission := permissions[0]

	newRequest := func(t *testing.T, permissionID string) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v2/permissions/"+permissionID, nil)
		require.NoError(t, err)
		return mux.SetURLVars(req, map[string]string{"permission_id": permissionID})
	}

	t.Run("returns 200 OK with the permission for a valid ID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, fmt.Sprintf("%d", seededPermission.ID)))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope permissionResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		assert.Equal(t, seededPermission.ID, envelope.Data.ID)
		assert.Equal(t, seededPermission.Authority, envelope.Data.Authority)
		assert.Equal(t, seededPermission.Name, envelope.Data.Name)
		assert.True(t, seededPermission.CreatedAt.Equal(envelope.Data.CreatedAt), "created_at should match the seeded permission")
		assert.True(t, seededPermission.UpdatedAt.Equal(envelope.Data.UpdatedAt), "updated_at should match the seeded permission")
		assert.Equal(t, seededPermission.DeletedAt.Valid, envelope.Data.DeletedAt.Valid, "deleted_at validity should match the seeded permission")
	})

	t.Run("returns 404 Not Found when the permission does not exist", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, "99999999"))
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})

	t.Run("returns 400 Bad Request for a malformed permission ID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, "not-an-int"))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestGetRole(t *testing.T) {
	var (
		db         = setupIdentityDB(t)
		ctx        = context.Background()
		handlerSet = newIdentityHandlers(db)
		handler    = handlerSet.GetRole
		roles      model.Roles
		err        error
	)

	roles, err = db.GetAllRoles(ctx, "", model.SQLFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, roles, "expected migrations to seed at least one role")
	seededRole := roles[0]

	newRequest := func(t *testing.T, roleID string) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v2/roles/"+roleID, nil)
		require.NoError(t, err)
		return mux.SetURLVars(req, map[string]string{"role_id": roleID})
	}

	t.Run("returns 200 OK with the role for a valid ID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, fmt.Sprintf("%d", seededRole.ID)))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope roleResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		assert.Equal(t, seededRole.ID, envelope.Data.ID)
		assert.Equal(t, seededRole.Name, envelope.Data.Name)
		assert.NotEmpty(t, envelope.Data.Permissions, "expected the role to preload its permissions")
		assert.True(t, seededRole.CreatedAt.Equal(envelope.Data.CreatedAt), "created_at should match the seeded role")
		assert.True(t, seededRole.UpdatedAt.Equal(envelope.Data.UpdatedAt), "updated_at should match the seeded role")
		assert.Equal(t, seededRole.DeletedAt.Valid, envelope.Data.DeletedAt.Valid, "deleted_at validity should match the seeded role")
	})

	t.Run("returns 404 Not Found when the role does not exist", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, "99999999"))
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})

	t.Run("returns 400 Bad Request for a malformed role ID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, "not-an-int"))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestListRoles(t *testing.T) {
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

	newRequest := func(t *testing.T, query url.Values) *http.Request {
		t.Helper()
		reqCtx := context.WithValue(ctx, bhctx.ValueKey, &bhctx.Context{})
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "/api/v2/roles", nil)
		require.NoError(t, err)
		req.URL.RawQuery = query.Encode()
		return req
	}

	t.Run("returns 200 OK with all seeded roles", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, url.Values{}))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope listRolesResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		require.Len(t, envelope.Data.Roles, len(roles))
		assert.NotEmpty(t, envelope.Data.Roles[0].Permissions, "expected roles to preload their permissions")
	})

	t.Run("returns 200 OK with roles sorted by name ascending", func(t *testing.T) {
		query := url.Values{}
		query.Add("sort_by", "name")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope listRolesResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		require.Len(t, envelope.Data.Roles, len(roles))
		for i := 1; i < len(envelope.Data.Roles); i++ {
			assert.LessOrEqual(t, envelope.Data.Roles[i-1].Name, envelope.Data.Roles[i].Name, "roles should be sorted by name ascending")
		}
	})

	t.Run("returns 200 OK with roles filtered by name", func(t *testing.T) {
		query := url.Values{}
		query.Add("name", "eq:"+seededRole.Name)

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope listRolesResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		require.Len(t, envelope.Data.Roles, 1)
		assert.Equal(t, seededRole.Name, envelope.Data.Roles[0].Name)
	})

	t.Run("returns 400 Bad Request for a non-sortable column", func(t *testing.T) {
		query := url.Values{}
		query.Add("sort_by", "invalidColumn")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsNotSortable)
	})

	t.Run("returns 400 Bad Request for a non-filterable column", func(t *testing.T) {
		query := url.Values{}
		query.Add("foo", "eq:bar")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsColumnNotFilterable)
	})

	t.Run("returns 400 Bad Request for a malformed filter predicate", func(t *testing.T) {
		query := url.Values{}
		query.Add("name", "invalidPredicate:foo")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsBadQueryParameterFilters)
	})

	t.Run("returns 400 Bad Request for an unsupported filter predicate", func(t *testing.T) {
		query := url.Values{}
		query.Add("name", "gt:0")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	})

	t.Run("returns the sort error when both the sort and filter are invalid", func(t *testing.T) {
		query := url.Values{}
		query.Add("sort_by", "invalidColumn")
		query.Add("foo", "eq:bar")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsNotSortable)
		assert.NotContains(t, recorder.Body.String(), api.ErrorResponseDetailsColumnNotFilterable)
	})
}

func TestListUsers(t *testing.T) {
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

	t.Run("returns 200 OK with regular users and excludes support accounts", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, url.Values{}))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope listUsersResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))

		names := principalsIn(envelope)
		assert.Contains(t, names, userA)
		assert.Contains(t, names, userB)
		assert.NotContains(t, names, supportUser, "support accounts must be excluded from the list")
	})

	t.Run("returns 200 OK with users sorted by principal_name ascending", func(t *testing.T) {
		query := url.Values{}
		query.Add("sort_by", "principal_name")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope listUsersResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		for i := 1; i < len(envelope.Data.Users); i++ {
			assert.LessOrEqual(t, envelope.Data.Users[i-1].PrincipalName, envelope.Data.Users[i].PrincipalName, "users should be sorted by principal_name ascending")
		}
	})

	t.Run("returns 200 OK with users filtered by email_address", func(t *testing.T) {
		query := url.Values{}
		query.Add("email_address", "eq:"+userA+"@example.com")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope listUsersResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		require.Len(t, envelope.Data.Users, 1)
		assert.Equal(t, userA, envelope.Data.Users[0].PrincipalName)
	})

	t.Run("returns 400 Bad Request for a non-sortable column", func(t *testing.T) {
		query := url.Values{}
		query.Add("sort_by", "invalidColumn")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsNotSortable)
	})

	t.Run("returns 400 Bad Request when sorting by the phantom deleted_at column", func(t *testing.T) {
		query := url.Values{}
		query.Add("sort_by", "deleted_at")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsNotSortable)
	})

	t.Run("returns 400 Bad Request for a non-filterable column", func(t *testing.T) {
		query := url.Values{}
		query.Add("foo", "eq:bar")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsColumnNotFilterable)
	})

	t.Run("returns 400 Bad Request for a malformed filter predicate", func(t *testing.T) {
		query := url.Values{}
		query.Add("email_address", "invalidPredicate:foo")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsBadQueryParameterFilters)
	})

	t.Run("returns 400 Bad Request for an unsupported filter predicate", func(t *testing.T) {
		query := url.Values{}
		query.Add("first_name", "gt:0")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	})

	t.Run("returns the sort error when both the sort and filter are invalid", func(t *testing.T) {
		// Sort is validated before filters, matching the legacy handler, so an
		// invalid sort takes precedence over an invalid filter in the response.
		query := url.Values{}
		query.Add("sort_by", "invalidColumn")
		query.Add("foo", "eq:bar")

		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, query))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), api.ErrorResponseDetailsNotSortable)
		assert.NotContains(t, recorder.Body.String(), api.ErrorResponseDetailsColumnNotFilterable)
	})
}
