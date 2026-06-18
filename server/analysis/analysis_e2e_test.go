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

package analysis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/registration"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/analysis/internal/appdb"
	"github.com/specterops/bloodhound/server/analysis/internal/handlers"
	"github.com/specterops/bloodhound/server/analysis/internal/services"
	"github.com/specterops/bloodhound/server/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// injectAuthMiddleware wraps the given handler, attaching a bhctx.Context that
// identifies the supplied user as the request owner. This stands in for the
// production auth middleware so the analysis handler can resolve a user without
// requiring the full auth stack.
func injectAuthMiddleware(handler http.HandlerFunc, userID uuid.UUID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bhCtx := &bhctx.Context{
			AuthCtx: auth.Context{Owner: model.User{Unique: model.Unique{ID: userID}}},
		}
		handler(w, bhctx.SetRequestContext(r, bhCtx))
	}
}

// analysisResponseEnvelope is the full JSON envelope shape returned by the GET
// and PUT /api/v2/analysis handlers. All seven documented fields are included so
// the tests can verify the complete contract rather than a subset.
type analysisResponseEnvelope struct {
	Data handlers.RequestedAnalysisView `json:"data"`
}

// setupAnalysisDB creates an isolated test database with all migrations applied.
// The database is automatically closed when the test ends.
func setupAnalysisDB(t *testing.T) *database.BloodhoundDB {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getAnalysisPostgresConfig(t), pgtestdb.NoopMigrator{})
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

// getAnalysisPostgresConfig reads the integration test configuration from the
// environment and returns a pgtestdb.Config for the analysis e2e tests.
func getAnalysisPostgresConfig(t *testing.T) pgtestdb.Config {
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

// newAnalysisGetHandler wires the pgx-backed analysis stack from a BloodhoundDB
// and returns its GET handler.
func newAnalysisGetHandler(db *database.BloodhoundDB) http.HandlerFunc {
	store := appdb.NewStore(db.Pool())
	svc := services.NewService(store)
	return handlers.NewHandlersContainer(svc).GetAnalysisRequest
}

// newCreateAnalysisHandler wires the pgx-backed analysis stack from a BloodhoundDB
// and returns its PUT handler wrapped with auth-injecting middleware.
func newCreateAnalysisHandler(db *database.BloodhoundDB, userID uuid.UUID) http.HandlerFunc {
	store := appdb.NewStore(db.Pool())
	svc := services.NewService(store)
	return injectAuthMiddleware(handlers.NewHandlersContainer(svc).CreateAnalysisRequest, userID)
}

// newCancelAnalysisHandler wires the DELETE /api/v2/analysis handler backed by
// a pgx-backed analysis stack and wrapped with auth-injecting middleware.
func newCancelAnalysisHandler(db *database.BloodhoundDB, userID uuid.UUID) http.HandlerFunc {
	store := appdb.NewStore(db.Pool())
	svc := services.NewService(store)
	return injectAuthMiddleware(handlers.NewHandlersContainer(svc).CancelAnalysisRequest, userID)
}

// mintJWT creates a signed JWT token for the given user using the authenticator.
// This creates a proper session in the database and returns a valid token.
// The user is granted all roles (and permissions) from the database for testing.
func mintJWT(t *testing.T, ctx context.Context, db *database.BloodhoundDB, auther api.Authenticator, user model.User) string {
	t.Helper()

	roles, err := db.GetAllRoles(ctx, "", model.SQLFilter{})
	require.NoError(t, err)

	authSecret := model.AuthSecret{
		Digest:       "dummy-digest-for-e2e-test",
		DigestMethod: "argon2",
		ExpiresAt:    time.Now().Add(24 * time.Hour).UTC(),
	}
	user.AuthSecret = &authSecret
	user.Roles = roles

	dbUser, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	dbUser, err = db.GetUser(ctx, dbUser.ID)
	require.NoError(t, err)
	require.NotNil(t, dbUser.AuthSecret, "User should have an AuthSecret")

	token, err := auther.CreateSession(ctx, dbUser, *dbUser.AuthSecret)
	require.NoError(t, err)
	return token
}

func TestGetAnalysisStatus(t *testing.T) {
	var (
		db         = setupAnalysisDB(t)
		ctx        = context.Background()
		cfg, _     = config.NewDefaultConfiguration()
		authExt    = api.NewAuthExtensions(cfg, db)
		auther     = api.NewAuthenticator(cfg, db, authExt)
		authorizer = auth.NewAuthorizer(db)
		resolver   = auth.NewIdentityResolver()
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	cfg.Crypto.JWT.SetSigningKeyBytes([]byte("test-secret-key-that-is-at-least-32-bytes-long"))

	registration.RegisterFossGlobalMiddleware(&routerInst, cfg, resolver, auther, db)

	modules.Register(modules.Deps{
		Router: &routerInst,
		Pool:   db.Pool(),
	})

	var (
		handler = routerInst.Handler()
		server  = httptest.NewServer(handler)
	)
	t.Cleanup(server.Close)

	var (
		user = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
		}
		token = mintJWT(t, ctx, db, auther, user)
	)

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/analysis/status", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		return req
	}

	// Authentication tests - validate middleware is properly attached
	t.Run("returns 401 Unauthorized when no authentication token is provided", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/analysis/status", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 401 Unauthorized when an invalid token is provided", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/analysis/status", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token-that-is-not-valid")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 Bad Request when Bearer prefix is missing", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/analysis/status", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 200 OK with zero-valued request when no request is pending", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var envelope analysisResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
		// Verify it's a zero-valued response
		assert.Empty(t, envelope.Data.RequestedBy)
		assert.Empty(t, envelope.Data.RequestType)
	})

	t.Run("returns 200 OK with all contract fields when a request is pending", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))
		require.NoError(t, db.RequestAnalysis(ctx, "test-user"))
		t.Cleanup(func() { _ = db.DeleteAnalysisRequest(ctx) })

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var envelope analysisResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
		assert.Equal(t, "test-user", envelope.Data.RequestedBy)
		assert.Equal(t, services.RequestedAnalysisTypeAnalysis, envelope.Data.RequestType)
		assert.NotZero(t, envelope.Data.RequestedAt)
		assert.False(t, envelope.Data.DeleteAllGraph)
		assert.False(t, envelope.Data.DeleteSourcelessGraph)
		assert.Empty(t, envelope.Data.DeleteSourceKinds)
		assert.Empty(t, envelope.Data.DeleteRelationships)
	})
}

func TestCreateAnalysisRequest(t *testing.T) {
	var (
		db         = setupAnalysisDB(t)
		ctx        = context.Background()
		cfg, _     = config.NewDefaultConfiguration()
		authExt    = api.NewAuthExtensions(cfg, db)
		auther     = api.NewAuthenticator(cfg, db, authExt)
		authorizer = auth.NewAuthorizer(db)
		resolver   = auth.NewIdentityResolver()
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	cfg.Crypto.JWT.SetSigningKeyBytes([]byte("test-secret-key-that-is-at-least-32-bytes-long"))

	registration.RegisterFossGlobalMiddleware(&routerInst, cfg, resolver, auther, db)

	modules.Register(modules.Deps{
		Router: &routerInst,
		Pool:   db.Pool(),
	})

	var (
		handler = routerInst.Handler()
		server  = httptest.NewServer(handler)
	)
	t.Cleanup(server.Close)

	var (
		user = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
		}
		token = mintJWT(t, ctx, db, auther, user)
	)

	newPutRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		return req
	}

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/analysis/status", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		return req
	}

	// Authentication tests - validate middleware is properly attached
	t.Run("returns 401 Unauthorized when no authentication token is provided for PUT", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 401 Unauthorized when an invalid token is provided for PUT", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token-that-is-not-valid")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 Bad Request when Bearer prefix is missing for PUT", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 202 Accepted when no request is pending", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))

		resp, err := http.DefaultClient.Do(newPutRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		// Handler returns 202 with no body, matching main behavior
		assert.Empty(t, resp.Header.Get("Content-Type"))
	})

	t.Run("returns 202 Accepted when a request is already pending", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))
		// Seed a known pending request attributed to a different requester
		require.NoError(t, db.RequestAnalysis(ctx, "prior-user"))
		t.Cleanup(func() { _ = db.DeleteAnalysisRequest(ctx) })

		resp, err := http.DefaultClient.Do(newPutRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		// Handler returns 202 with no body, matching main behavior
		assert.Empty(t, resp.Header.Get("Content-Type"))
	})

	t.Run("GET reflects the pending request created by PUT", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))
		t.Cleanup(func() { _ = db.DeleteAnalysisRequest(ctx) })

		resp, err := http.DefaultClient.Do(newPutRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		resp, err = http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var envelope analysisResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
		assert.NotEmpty(t, envelope.Data.RequestedBy)
		assert.Equal(t, services.RequestedAnalysisTypeAnalysis, envelope.Data.RequestType)
	})
}

func TestCancelAnalysisRequest(t *testing.T) {
	var (
		db         = setupAnalysisDB(t)
		ctx        = context.Background()
		cfg, _     = config.NewDefaultConfiguration()
		authExt    = api.NewAuthExtensions(cfg, db)
		auther     = api.NewAuthenticator(cfg, db, authExt)
		authorizer = auth.NewAuthorizer(db)
		resolver   = auth.NewIdentityResolver()
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	cfg.Crypto.JWT.SetSigningKeyBytes([]byte("test-secret-key-that-is-at-least-32-bytes-long"))

	registration.RegisterFossGlobalMiddleware(&routerInst, cfg, resolver, auther, db)

	modules.Register(modules.Deps{
		Router: &routerInst,
		Pool:   db.Pool(),
	})

	var (
		handler = routerInst.Handler()
		server  = httptest.NewServer(handler)
	)
	t.Cleanup(server.Close)

	var (
		user = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
		}
		token = mintJWT(t, ctx, db, auther, user)
	)

	newDeleteRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		return req
	}

	// Authentication tests - validate middleware is properly attached
	t.Run("returns 401 Unauthorized when no authentication token is provided for DELETE", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 401 Unauthorized when an invalid token is provided for DELETE", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token-that-is-not-valid")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 Bad Request when Bearer prefix is missing for DELETE", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 202 Accepted when an analysis request is pending", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))
		require.NoError(t, db.RequestAnalysis(ctx, "test-user"))

		resp, err := http.DefaultClient.Do(newDeleteRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("returns 404 Not Found when no request is pending", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))

		resp, err := http.DefaultClient.Do(newDeleteRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns 409 Conflict when a deletion request is pending", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))
		require.NoError(t, db.RequestCollectedGraphDataDeletion(ctx, model.AnalysisRequest{
			RequestedBy: "test-user",
			RequestType: model.AnalysisRequestDeletion,
		}))
		t.Cleanup(func() { _ = db.DeleteAnalysisRequest(ctx) })

		resp, err := http.DefaultClient.Do(newDeleteRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}
