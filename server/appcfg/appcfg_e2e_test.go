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
	"io"
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
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/specterops/bloodhound/server/modules"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// datapipeStatusResponseEnvelope is the full JSON envelope shape returned by
// the GET /api/v2/datapipe/status handler. All six documented fields are included.
type datapipeStatusResponseEnvelope struct {
	Data model.DatapipeStatusWrapper `json:"data"`
}

func noopRateLimit() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler { return next }
}

func testDogTags() dogtags.Service {
	return dogtags.NewTestService(dogtags.TestOverrides{})
}

func assertDatapipeStatusWireContract(t *testing.T, body []byte, expectedStatus model.DatapipeStatus) {
	t.Helper()

	var envelope struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	require.Len(t, envelope.Data, 6)

	statusJSON, found := envelope.Data["status"]
	require.True(t, found)
	var status model.DatapipeStatus
	require.NoError(t, json.Unmarshal(statusJSON, &status))
	assert.Equal(t, expectedStatus, status)

	for _, field := range []string{"updated_at", "last_complete_analysis_at", "last_analysis_run_at", "last_complete_optimize_at"} {
		timestampJSON, found := envelope.Data[field]
		require.True(t, found, "%s must be present", field)
		require.NotEqual(t, "null", string(timestampJSON), "%s must remain a date-time string", field)

		var timestamp time.Time
		require.NoError(t, json.Unmarshal(timestampJSON, &timestamp), "%s must be a valid date-time", field)
	}

	_, found = envelope.Data["next_scheduled_analysis_at"]
	require.True(t, found, "next_scheduled_analysis_at must be present")
}

func TestGetDatapipeStatus(t *testing.T) {
	var (
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInst *router.Router, db *database.BloodhoundDB) {
			// Register the appcfg module using the new architecture
			modules.Register(modules.Deps{
				Router:              routerInst,
				Pool:                db.Pool(),
				Graph:               &graph.DatabaseSwitch{},
				RateLimitMiddleware: noopRateLimit,
				DogTags:             testDogTags(),
			})
		})
		db     = harness.DB
		server = harness.Server
	)

	// Create a test user and get a valid JWT token for authentication
	var (
		user = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
		}
		token = servertest.MintJWT(t, ctx, db, harness.Auther, user)
	)

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/datapipe/status", nil)
		require.NoError(t, err)
		// Add authentication header
		req.Header.Set("Authorization", "Bearer "+token)
		return req
	}

	// Authentication tests - validate middleware is properly attached
	t.Run("returns 401 Unauthorized when no authentication token is provided", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/datapipe/status", nil)
		require.NoError(t, err)
		// No Authorization header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 401 Unauthorized when an invalid token is provided", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/datapipe/status", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token-that-is-not-valid")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 Bad Request when Bearer prefix is missing", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/datapipe/status", nil)
		require.NoError(t, err)
		// Token without "Bearer" prefix
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 200 OK with datapipe status in idle state", func(t *testing.T) {
		// Ensure datapipe is in idle state (default after migration)
		err := db.SetDatapipeStatus(ctx, model.DatapipeStatusIdle)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assertDatapipeStatusWireContract(t, body, model.DatapipeStatusIdle)
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

	t.Run("returns 200 OK with datapipe status in optimizing state", func(t *testing.T) {
		err := db.SetDatapipeStatus(ctx, model.DatapipeStatusOptimizing)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var envelope datapipeStatusResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))

		assert.Equal(t, model.DatapipeStatusOptimizing, envelope.Data.Status)
	})

	t.Run("returns 200 OK with updated timestamps after analysis and optimization", func(t *testing.T) {
		// Set analysis start time
		err := db.SetLastAnalysisStartTime(ctx)
		require.NoError(t, err)

		// Complete the analysis
		err = db.UpdateLastAnalysisCompleteTime(ctx)
		require.NoError(t, err)

		// Complete graph storage optimization
		err = db.SetLastGraphOptimizeTime(ctx)
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
		assert.False(t, envelope.Data.LastCompleteOptimizeAt.IsZero(), "last_complete_optimize_at should be set")
	})
}

// # Application Configuration Tests

// ## Testing legacy pre-onion implementation - routing setup

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

// newGetConfigsHandler wires the legacy GetApplicationConfigurations handler
// backed by the given database. This mirrors the pre-migration wiring so the
// e2e test locks in the existing HTTP contract before the endpoint is moved
// into the identity vertical slice. Only the necessary resources are loaded.
func newGetConfigsHandler(t *testing.T, db *database.BloodhoundDB, cfg config.Configuration, auther api.Authenticator) http.Handler {
	t.Helper()

	// Register the actual production routes using v2.Resources and NewV2API
	// This ensures we're testing the real route wiring, not a test-specific mock.
	// Most parameters can be nil since we're only testing the datapipe/status endpoint.

	var (
		resolver   = auth.NewIdentityResolver()
		authorizer = auth.NewAuthorizer(db)
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	// Set up JWT signing key
	cfg.Crypto.JWT.SetSigningKeyBytes([]byte("test-secret-key-that-is-at-least-32-bytes-long"))

	apiCache, err := cache.NewCache(cache.Config{MaxSize: 1000})
	require.NoError(t, err)

	// Register global middleware (auth, rate limiting, etc)
	registration.RegisterFossGlobalMiddleware(&routerInst, cfg, resolver, auther, db)
	resources := v2.NewResources(
		db,                          // rdms database.Database
		nil,                         // graphDB *graph.DatabaseSwitch (not needed)
		cfg,                         // cfg config.Configuration
		apiCache,                    // apiCache cache.Cache
		nil,                         // graphQuery queries.Graph (not needed)
		config.CollectorManifests{}, // collectorManifests
		authorizer,                  // authorizer
		auther,                      // authenticator
		upload.IngestSchema{},       // ingestSchema (needs non-nil struct)
		nil,                         // fileServiceResolver (not needed)
		nil,                         // dogtagsService (not needed)
		nil,                         // openGraphSchemaService (not needed)
		nil,                         // alerts.Publisher (not needed)
	)
	registration.NewV2API(resources, &routerInst)

	return routerInst.Handler()
}

// getConfigsResponseEnvelope is the JSON envelope shape returned by the
// GET /api/v2/configs handler.
type getConfigsResponseEnvelope struct {
	Data []appcfg.Parameter `json:"data"`
}

func TestGetAppConfigs(t *testing.T) {
	var (
		ctx     = context.Background()
		db      = setupAppcfgDB(t)
		cfg, _  = config.NewDefaultConfiguration()
		authExt = api.NewAuthExtensions(cfg, db)
		auther  = api.NewAuthenticator(cfg, db, authExt)
		handler = newGetConfigsHandler(t, db, cfg, auther)
		user    = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
			EULAAccepted:  true, // Required for permission checks to work
			Roles:         model.Roles{servertest.AdminRole(t, ctx, db)},
		}
		token = servertest.MintJWT(t, ctx, db, auther, user)
	)

	t.Run("returns 200 OK with all seeded configs", func(t *testing.T) {
		recorder := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/config", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope getConfigsResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))

		// Minimal assertions to avoid live data changes breaking tests
		assert.Equal(t, "auth.password_expiration_window", string(envelope.Data[0].Key))
		assert.Equal(t, "Local Auth Password Expiry Window", envelope.Data[0].Name)
		assert.NotEmpty(t, envelope.Data[0].ID)
		assert.NotEmpty(t, envelope.Data[0].CreatedAt)
		assert.Empty(t, envelope.Data[0].DeletedAt)
		assert.NotEmpty(t, envelope.Data[0].UpdatedAt)
		assert.NotEmpty(t, envelope.Data[0].Value)
	})

	t.Run("get specific config returns OK", func(t *testing.T) {
		recorder := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/config?parameter=eq%3Aauth.password_expiration_window", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var envelope getConfigsResponseEnvelope
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))

		// Minimal assertions to avoid live data changes breaking tests
		assert.Len(t, envelope.Data, 1)
		assert.Equal(t, "auth.password_expiration_window", string(envelope.Data[0].Key))
		assert.Equal(t, "Local Auth Password Expiry Window", envelope.Data[0].Name)
		assert.NotEmpty(t, envelope.Data[0].ID)
		assert.NotEmpty(t, envelope.Data[0].CreatedAt)
		assert.Empty(t, envelope.Data[0].DeletedAt)
		assert.NotEmpty(t, envelope.Data[0].UpdatedAt)
		assert.NotEmpty(t, envelope.Data[0].Value)
	})

	t.Run("returns 401 Unauthorized when an invalid token is provided", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v2/config", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-that-is-not-valid")

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)

		var envelope api.ErrorWrapper
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		assert.Equal(t, http.StatusUnauthorized, envelope.HTTPStatus)
		assert.NotEmpty(t, envelope.Errors)
	})

	t.Run("returns 401 Unauthorized when no auth token is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/config", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)

		var envelope api.ErrorWrapper
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		assert.Equal(t, http.StatusUnauthorized, envelope.HTTPStatus)
		assert.NotEmpty(t, envelope.Errors)
	})

	t.Run("returns 400 bad request with malformed search", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/config?parameter=notacomparator%3Ahelloworld", nil)
		recorder := httptest.NewRecorder()
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var envelope api.ErrorWrapper
		require.NoError(t, json.NewDecoder(recorder.Body).Decode(&envelope))
		assert.Equal(t, http.StatusBadRequest, envelope.HTTPStatus)
		assert.NotEmpty(t, envelope.Errors)
	})
}
