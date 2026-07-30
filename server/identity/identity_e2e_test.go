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
	"bytes"
	"context"
	"encoding/base64"
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
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2auth "github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
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

// authTokenResponseEnvelope is the JSON envelope documented for
// POST /api/v2/tokens.
type authTokenResponseEnvelope struct {
	Data model.AuthToken `json:"data"`
}

type errorResponseEnvelope struct {
	HTTPStatus int `json:"http_status"`
	Errors     []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func assertErrorResponse(t *testing.T, response *http.Response, expectedStatus int, expectedMessage string) {
	t.Helper()

	assert.Equal(t, expectedStatus, response.StatusCode)
	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))

	var envelope errorResponseEnvelope
	require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))
	assert.Equal(t, expectedStatus, envelope.HTTPStatus)
	require.Len(t, envelope.Errors, 1)
	assert.Equal(t, expectedMessage, envelope.Errors[0].Message)
}

// newCreateAuthTokenHandler wires the existing CreateAuthToken handler to the
// integration database. The injected auth context stands in for the production
// authentication middleware.
func newCreateAuthTokenHandler(db *database.BloodhoundDB, user model.User, permissionOverrides auth.PermissionOverrides) http.HandlerFunc {
	var (
		authorizer    = auth.NewAuthorizer(db)
		authenticator = api.NewAuthenticator(config.Configuration{}, db, nil)
		resource      = v2auth.NewManagementResource(config.Configuration{}, db, authorizer, authenticator, nil, nil)
	)

	return func(response http.ResponseWriter, request *http.Request) {
		authContext := &bhctx.Context{
			AuthCtx: auth.Context{
				Owner:               user,
				PermissionOverrides: permissionOverrides,
			},
		}
		resource.CreateAuthToken(response, bhctx.SetRequestContext(request, authContext))
	}
}

func createIdentityTestUser(t *testing.T, db *database.BloodhoundDB, principalName string) model.User {
	t.Helper()

	user, err := db.CreateUser(context.Background(), model.User{
		FirstName:     null.StringFrom("Identity"),
		LastName:      null.StringFrom("Test"),
		EmailAddress:  null.StringFrom(principalName),
		PrincipalName: principalName,
	})
	require.NoError(t, err)

	return user
}

func setAPITokensEnabled(t *testing.T, db *database.BloodhoundDB, enabled bool) {
	t.Helper()

	value, err := types.NewJSONBObject(appcfg.APITokensParameter{Enabled: enabled})
	require.NoError(t, err)
	require.NoError(t, db.SetConfigurationParameter(context.Background(), appcfg.Parameter{
		Key:   appcfg.APITokens,
		Value: value,
	}))
}

func TestCreateAuthToken(t *testing.T) {
	var (
		db                = setupIdentityDB(t)
		ctx               = context.Background()
		authenticatedUser = createIdentityTestUser(t, db, "create-token-user@example.com")
		targetUser        = createIdentityTestUser(t, db, "create-token-target@example.com")
		muxRouter         = mux.NewRouter()
		server            = httptest.NewServer(muxRouter)
	)
	muxRouter.HandleFunc(
		"/api/v2/tokens",
		newCreateAuthTokenHandler(db, authenticatedUser, auth.PermissionOverrides{}),
	).Methods(http.MethodPost)
	t.Cleanup(server.Close)

	newRequest := func(t *testing.T, body string) *http.Request {
		t.Helper()

		request, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			server.URL+"/api/v2/tokens",
			bytes.NewBufferString(body),
		)
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		return request
	}

	t.Run("returns 200 OK and persists a token for the authenticated user", func(t *testing.T) {
		response, err := http.DefaultClient.Do(newRequest(t, `{"token_name":"automation token"}`))
		require.NoError(t, err)
		defer response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "application/json", response.Header.Get("Content-Type"))

		var envelope authTokenResponseEnvelope
		require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))

		assert.NotEqual(t, uuid.Nil, envelope.Data.ID)
		assert.True(t, envelope.Data.UserID.Valid)
		assert.Equal(t, authenticatedUser.ID, envelope.Data.UserID.UUID)
		assert.True(t, envelope.Data.Name.Valid)
		assert.Equal(t, "automation token", envelope.Data.Name.String)
		assert.Equal(t, auth.HMAC_SHA2_256, envelope.Data.HmacMethod)
		assert.NotZero(t, envelope.Data.LastAccess)
		assert.True(t, envelope.Data.CreatedBy.Valid)
		assert.Equal(t, authenticatedUser.ID, envelope.Data.CreatedBy.UUID)
		assert.NotZero(t, envelope.Data.CreatedAt)
		assert.NotZero(t, envelope.Data.UpdatedAt)
		assert.False(t, envelope.Data.DeletedAt.Valid)

		decodedKey, err := base64.StdEncoding.DecodeString(envelope.Data.Key)
		require.NoError(t, err)
		assert.Len(t, decodedKey, 40)

		persistedToken, err := db.GetUserToken(ctx, authenticatedUser.ID, envelope.Data.ID)
		require.NoError(t, err)
		assert.Equal(t, envelope.Data.ID, persistedToken.ID)
		assert.Equal(t, envelope.Data.Key, persistedToken.Key)
		assert.Equal(t, envelope.Data.Name, persistedToken.Name)
	})

	t.Run("returns 400 Bad Request for malformed JSON", func(t *testing.T) {
		response, err := http.DefaultClient.Do(newRequest(t, `{"token_name":`))
		require.NoError(t, err)
		defer response.Body.Close()

		assertErrorResponse(t, response, http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError)
	})

	t.Run("returns 403 Forbidden when creating a token for another user", func(t *testing.T) {
		body := fmt.Sprintf(`{"token_name":"forbidden token","user_id":%q}`, targetUser.ID)
		response, err := http.DefaultClient.Do(newRequest(t, body))
		require.NoError(t, err)
		defer response.Body.Close()

		assertErrorResponse(t, response, http.StatusForbidden, "missing permission to create tokens for other users")
	})

	t.Run("returns 500 Internal Server Error for an invalid target user ID", func(t *testing.T) {
		var (
			adminRouter = mux.NewRouter()
			adminServer = httptest.NewServer(adminRouter)
			overrides   = auth.PermissionOverrides{
				Enabled:     true,
				Permissions: model.Permissions{auth.Permissions().AuthManageUsers},
			}
		)
		adminRouter.HandleFunc(
			"/api/v2/tokens",
			newCreateAuthTokenHandler(db, authenticatedUser, overrides),
		).Methods(http.MethodPost)
		t.Cleanup(adminServer.Close)

		request, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			adminServer.URL+"/api/v2/tokens",
			bytes.NewBufferString(`{"token_name":"invalid owner","user_id":"not-a-uuid"}`),
		)
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		defer response.Body.Close()

		assertErrorResponse(t, response, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError)
	})

	t.Run("returns 403 Forbidden when API token creation is disabled", func(t *testing.T) {
		setAPITokensEnabled(t, db, false)
		t.Cleanup(func() {
			setAPITokensEnabled(t, db, true)
		})

		response, err := http.DefaultClient.Do(newRequest(t, `{"token_name":"disabled token"}`))
		require.NoError(t, err)
		defer response.Body.Close()

		assertErrorResponse(t, response, http.StatusForbidden, "API key creation is disabled")
	})
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
