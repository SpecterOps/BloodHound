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

package appcfg_test

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

	"github.com/gorilla/mux"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/registration"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// datapipeStatusResponseEnvelope is the full JSON envelope shape returned by
// the GET /api/v2/datapipe/status handler. All five documented fields are included.
type datapipeStatusResponseEnvelope struct {
	Data model.DatapipeStatusWrapper `json:"data"`
}

// setupAppcfgDB creates an isolated test database with all migrations applied.
// The database is automatically closed when the test ends.
func setupAppcfgDB(t *testing.T) *database.BloodhoundDB {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getAppcfgPostgresConfig(t), pgtestdb.NoopMigrator{})
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

// getAppcfgPostgresConfig reads the integration test configuration from the
// environment and returns a pgtestdb.Config for the appcfg e2e tests.
func getAppcfgPostgresConfig(t *testing.T) pgtestdb.Config {
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

// newDatapipeStatusHandler wires the old v2.Resources.GetDatapipeStatus handler.
func newDatapipeStatusHandler(db *database.BloodhoundDB) http.HandlerFunc {
	return v2.Resources{DB: db}.GetDatapipeStatus
}

// mintJWT creates a signed JWT token for the given user using the authenticator.
// This creates a proper session in the database and returns a valid token.
func mintJWT(t *testing.T, ctx context.Context, db *database.BloodhoundDB, auther api.Authenticator, user model.User) string {
	t.Helper()

	// Create an auth secret for the user so they can authenticate
	authSecret := model.AuthSecret{
		Digest:       "dummy-digest-for-e2e-test",
		DigestMethod: "argon2",
		ExpiresAt:    time.Now().Add(24 * time.Hour).UTC(),
	}
	user.AuthSecret = &authSecret

	dbUser, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	// Reload user to get the AuthSecret with its populated ID
	dbUser, err = db.GetUser(ctx, dbUser.ID)
	require.NoError(t, err)
	require.NotNil(t, dbUser.AuthSecret, "User should have an AuthSecret")

	token, err := auther.CreateSession(ctx, dbUser, *dbUser.AuthSecret)
	require.NoError(t, err)
	return token
}

func TestGetDatapipeStatus(t *testing.T) {
	var (
		db        = setupAppcfgDB(t)
		ctx       = context.Background()
		muxRouter = mux.NewRouter()
		server    = httptest.NewServer(muxRouter)
	)
	muxRouter.HandleFunc("/api/v2/datapipe/status", newDatapipeStatusHandler(db)).Methods(http.MethodGet)
	t.Cleanup(server.Close)

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/datapipe/status", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 200 OK with datapipe status in idle state", func(t *testing.T) {
		// Ensure datapipe is in idle state (default after migration)
		err := db.SetDatapipeStatus(ctx, model.DatapipeStatusIdle)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var envelope datapipeStatusResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))

		// Verify all contract fields are present and valid
		assert.Equal(t, model.DatapipeStatusIdle, envelope.Data.Status)
		assert.IsType(t, time.Time{}, envelope.Data.UpdatedAt)
		assert.IsType(t, time.Time{}, envelope.Data.LastCompleteAnalysisAt)
		assert.IsType(t, time.Time{}, envelope.Data.LastAnalysisRunAt)
		// next_scheduled_analysis_at is nullable (Enterprise-only field)
		assert.IsType(t, time.Time{}, envelope.Data.NextScheduledAnalysisAt.Time)
	})

	t.Run("returns 200 OK with datapipe status in ingesting state", func(t *testing.T) {
		// Set datapipe to ingesting state
		err := db.SetDatapipeStatus(ctx, model.DatapipeStatusIngesting)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var envelope datapipeStatusResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))

		assert.Equal(t, model.DatapipeStatusIngesting, envelope.Data.Status)
	})

	t.Run("returns 200 OK with datapipe status in analyzing state", func(t *testing.T) {
		// Set datapipe to analyzing state
		err := db.SetDatapipeStatus(ctx, model.DatapipeStatusAnalyzing)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var envelope datapipeStatusResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))

		assert.Equal(t, model.DatapipeStatusAnalyzing, envelope.Data.Status)
	})

	t.Run("returns 200 OK with updated timestamps after analysis", func(t *testing.T) {
		// Set analysis start time
		err := db.SetLastAnalysisStartTime(ctx)
		require.NoError(t, err)

		// Complete the analysis
		err = db.UpdateLastAnalysisCompleteTime(ctx)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var envelope datapipeStatusResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))

		// Timestamps should be non-zero after setting them
		assert.False(t, envelope.Data.LastAnalysisRunAt.IsZero(), "last_analysis_run_at should be set")
		assert.False(t, envelope.Data.LastCompleteAnalysisAt.IsZero(), "last_complete_analysis_at should be set")
	})
}

func TestGetDatapipeStatus_WithRouting(t *testing.T) {
	var (
		ctx        = context.Background()
		db         = setupAppcfgDB(t)
		cfg, _     = config.NewDefaultConfiguration()
		resolver   = auth.NewIdentityResolver()
		authExt    = api.NewAuthExtensions(cfg, db)
		auther     = api.NewAuthenticator(cfg, db, authExt)
		authorizer = auth.NewAuthorizer(db)
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	// Set up JWT signing key
	cfg.Crypto.JWT.SetSigningKeyBytes([]byte("test-secret-key-that-is-at-least-32-bytes-long"))

	// Register global middleware (auth, rate limiting, etc)
	registration.RegisterFossGlobalMiddleware(&routerInst, cfg, resolver, auther, db)

	// Register the actual production routes using v2.Resources and NewV2API
	// This ensures we're testing the real route wiring, not a test-specific mock.
	// Most parameters can be nil since we're only testing the datapipe/status endpoint.
	apiCache, err := cache.NewCache(cache.Config{MaxSize: 1000})
	require.NoError(t, err)

	resources := v2.NewResources(
		db,                          // rdms database.Database
		nil,                         // graphDB *graph.DatabaseSwitch (not needed for datapipe status)
		cfg,                         // cfg config.Configuration
		apiCache,                    // apiCache cache.Cache
		nil,                         // graphQuery queries.Graph (not needed for datapipe status)
		config.CollectorManifests{}, // collectorManifests
		authorizer,                  // authorizer
		auther,                      // authenticator
		upload.IngestSchema{},       // ingestSchema (needs non-nil struct)
		nil,                         // fileServiceResolver (not needed for datapipe status)
		nil,                         // dogtagsService (not needed for datapipe status)
		nil,                         // openGraphSchemaService (not needed for datapipe status)
	)
	registration.NewV2API(resources, &routerInst)

	handler := routerInst.Handler()

	t.Run("returns 401 Unauthorized when no auth token is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/datapipe/status", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var envelope api.ErrorWrapper
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&envelope))
		assert.Equal(t, http.StatusUnauthorized, envelope.HTTPStatus)
		assert.NotEmpty(t, envelope.Errors)
	})

	t.Run("returns 401 Unauthorized with invalid auth token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/datapipe/status", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 200 OK with valid auth token", func(t *testing.T) {
		// Create a test user
		user := model.User{
			PrincipalName: "test-user",
		}

		// Mint a valid JWT (creates user and session in database)
		token := mintJWT(t, ctx, db, auther, user)

		// Set datapipe status
		err := db.SetDatapipeStatus(ctx, model.DatapipeStatusIdle)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/datapipe/status", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var envelope datapipeStatusResponseEnvelope
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&envelope))
		assert.Equal(t, model.DatapipeStatusIdle, envelope.Data.Status)
	})
}
