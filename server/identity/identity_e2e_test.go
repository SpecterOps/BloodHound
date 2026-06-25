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

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/peterldowns/pgtestdb"
	authapi "github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
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

// createTestUser inserts a minimal user into the database for use in tests.
func createTestUser(t *testing.T, ctx context.Context, db *database.BloodhoundDB, principalName string) model.User {
	t.Helper()
	user, err := db.CreateUser(ctx, model.User{
		PrincipalName:   principalName,
		EULAAccepted:    true,
		AllEnvironments: true,
	})
	require.NoError(t, err)
	return user
}

// newManagementResource wires a real ManagementResource backed by the given database.
func newManagementResource(t *testing.T, db *database.BloodhoundDB) authapi.ManagementResource {
	t.Helper()
	cfg, err := config.NewDefaultConfiguration()
	require.NoError(t, err)
	return authapi.NewManagementResource(cfg, db, auth.NewAuthorizer(db), nil, nil, dogtags.NewDefaultService())
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

// userResponseEnvelope is the JSON envelope shape returned by the
// GET /api/v2/bloodhound-users/{user_id} handler.
type userResponseEnvelope struct {
	Data model.User `json:"data"`
}

func TestGetPermission(t *testing.T) {
	var (
		db          = setupIdentityDB(t)
		ctx         = context.Background()
		resource    = newManagementResource(t, db)
		handler     = resource.GetPermission
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
		db       = setupIdentityDB(t)
		ctx      = context.Background()
		resource = newManagementResource(t, db)
		handler  = resource.GetRole
		roles    model.Roles
		err      error
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

func TestGetUser(t *testing.T) {
	var (
		db       = setupIdentityDB(t)
		ctx      = context.Background()
		user     = createTestUser(t, ctx, db, "get-user-principal")
		resource = newManagementResource(t, db)
		handler  = resource.GetUser
	)

	newRequest := func(t *testing.T, userID string) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v2/bloodhound-users/"+userID, nil)
		require.NoError(t, err)
		return mux.SetURLVars(req, map[string]string{"user_id": userID})
	}

	t.Run("returns 200 OK with the user for a valid ID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, user.ID.String()))

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope userResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		assert.Equal(t, user.ID, envelope.Data.ID)
		assert.Equal(t, user.PrincipalName, envelope.Data.PrincipalName)
	})

	t.Run("returns 404 Not Found when the user does not exist", func(t *testing.T) {
		nonExistentID := uuid.Must(uuid.NewV4())
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, nonExistentID.String()))
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})

	t.Run("returns 400 Bad Request for a malformed user ID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler(recorder, newRequest(t, "not-a-uuid"))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}
