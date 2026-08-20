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

package featureflags_test

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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/analysis"
	"github.com/specterops/bloodhound/server/featureflags/internal/appdb"
	"github.com/specterops/bloodhound/server/featureflags/internal/handlers"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// injectAuthMiddleware wraps the given handler, attaching a bhctx.Context that
// identifies the supplied user as the request owner. This stands in for the
// production auth middleware so feature-flag handlers can resolve a user
// without requiring the full auth stack.
func injectAuthMiddleware(handler http.HandlerFunc, userID uuid.UUID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bhCtx := &bhctx.Context{
			AuthCtx: auth.Context{Owner: model.User{Unique: model.Unique{ID: userID}}},
		}
		handler(w, bhctx.SetRequestContext(r, bhCtx))
	}
}

// featureFlagsEnvelope is the JSON envelope returned by GET /api/v2/features.
type featureFlagsEnvelope struct {
	Data handlers.FeatureFlagsView `json:"data"`
}

// featureFlagEnvelope is the JSON envelope returned by PUT /api/v2/features/{feature_id}/toggle.
type featureFlagEnvelope struct {
	Data handlers.FeatureFlagView `json:"data"`
}

// setupFeatureFlagsDB creates an isolated test database with all migrations applied.
// The database is automatically closed when the test ends.
func setupFeatureFlagsDB(t *testing.T) *database.BloodhoundDB {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getFeatureFlagsPostgresConfig(t), pgtestdb.NoopMigrator{})
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

// getFeatureFlagsPostgresConfig reads the integration test configuration from the
// environment and returns a pgtestdb.Config for the featureflags e2e tests.
func getFeatureFlagsPostgresConfig(t *testing.T) pgtestdb.Config {
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

// newGetAllFlagsHandler wires the pgx-backed featureflags stack from a BloodhoundDB
// and returns its GET handler.
func newGetAllFlagsHandler(db *database.BloodhoundDB) http.HandlerFunc {
	store := appdb.NewStore(db.Pool())
	svc := services.NewService(store, analysis.NewAnalysisRequestSubmitter(db.Pool()))
	return handlers.NewHandlersContainer(svc).GetAllFlags
}

// newToggleFlagHandler wires the PUT /api/v2/features/{feature_id}/toggle handler backed
// by a pgx-backed featureflags stack and wrapped with auth-injecting middleware.
func newToggleFlagHandler(db *database.BloodhoundDB, userID uuid.UUID) http.HandlerFunc {
	store := appdb.NewStore(db.Pool())
	svc := services.NewService(store, analysis.NewAnalysisRequestSubmitter(db.Pool()))
	return injectAuthMiddleware(handlers.NewHandlersContainer(svc).ToggleFlag, userID)
}

// seedFeatureFlag inserts a feature_flags row directly, bypassing the Store API
// which does not expose flag creation. Returns the inserted flag's ID.
// If the key already exists, the row is updated in place so built-in migrated flags can be reused by tests.
func seedFeatureFlag(t *testing.T, ctx context.Context, pool *pgxpool.Pool, key, name string, enabled, userUpdatable bool) int32 {
	t.Helper()

	var id int32
	err := pool.QueryRow(ctx, `
		INSERT INTO feature_flags (key, name, description, enabled, user_updatable, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (key) DO UPDATE
		SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			enabled = EXCLUDED.enabled,
			user_updatable = EXCLUDED.user_updatable,
			updated_at = NOW()
		RETURNING id
	`, key, name, "seeded flag", enabled, userUpdatable).Scan(&id)
	require.NoError(t, err)
	return id
}

func assertAnalysisRequest(t *testing.T, ctx context.Context, pool *pgxpool.Pool, expectedRequestedBy string, expectedStep int32) {
	t.Helper()

	var (
		requestedBy  string
		requestType  string
		analysisStep int32
	)

	err := pool.QueryRow(ctx, `
		SELECT requested_by, request_type, analysis_step
		FROM analysis_request_switch
		LIMIT 1
	`).Scan(&requestedBy, &requestType, &analysisStep)
	require.NoError(t, err)
	assert.Equal(t, expectedRequestedBy, requestedBy)
	assert.Equal(t, string(model.AnalysisRequestAnalysis), requestType)
	assert.Equal(t, expectedStep, analysisStep)
}

func assertNoAnalysisRequest(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM analysis_request_switch WHERE request_type = $1`, model.AnalysisRequestAnalysis).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestGetAllFlags(t *testing.T) {
	var (
		db        = setupFeatureFlagsDB(t)
		ctx       = context.Background()
		muxRouter = mux.NewRouter()
		server    = httptest.NewServer(muxRouter)
	)
	muxRouter.HandleFunc("/api/v2/features", newGetAllFlagsHandler(db)).Methods(http.MethodGet)
	t.Cleanup(server.Close)

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/features", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 200 OK with the seeded flags", func(t *testing.T) {
		seeded := seedFeatureFlag(t, ctx, db.Pool(), "e2e_get_all_flag", "E2E Get All Flag", true, true)

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var envelope featureFlagsEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))

		found := false
		for _, flag := range envelope.Data {
			if flag.ID == seeded {
				found = true
				assert.Equal(t, "e2e_get_all_flag", flag.Key)
				assert.Equal(t, "E2E Get All Flag", flag.Name)
				assert.True(t, flag.Enabled)
				assert.True(t, flag.UserUpdatable)
			}
		}
		assert.True(t, found, "seeded flag should be in the returned list")
	})
}

func TestToggleFlag(t *testing.T) {
	var (
		db     = setupFeatureFlagsDB(t)
		userID = uuid.Must(uuid.NewV4())
		ctx    = context.Background()
	)

	newToggleRequest := func(t *testing.T, server *httptest.Server, featureID int32) *http.Request {
		t.Helper()
		target := fmt.Sprintf("%s/api/v2/features/%d/toggle", server.URL, featureID)
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, nil)
		require.NoError(t, err)
		return req
	}

	newServer := func(t *testing.T) *httptest.Server {
		t.Helper()
		muxRouter := mux.NewRouter()
		muxRouter.HandleFunc(
			"/api/v2/features/{"+api.URIPathVariableFeatureID+"}/toggle",
			newToggleFlagHandler(db, userID),
		).Methods(http.MethodPut)
		server := httptest.NewServer(muxRouter)
		t.Cleanup(server.Close)
		return server
	}

	type testCase struct {
		name           string
		seedFlag       func(t *testing.T) int32
		expectedStatus int
		assertResponse func(t *testing.T, body *http.Response)
		assertDB       func(t *testing.T)
	}

	testCases := []testCase{
		{
			name: "Success: returns 200 OK and flips the flag when it is user-updatable",
			seedFlag: func(t *testing.T) int32 {
				t.Helper()
				return seedFeatureFlag(t, ctx, db.Pool(), "e2e_toggle_flag", "E2E Toggle Flag", false, true)
			},
			expectedStatus: http.StatusOK,
			assertResponse: func(t *testing.T, body *http.Response) {
				t.Helper()

				var envelope featureFlagEnvelope
				require.NoError(t, json.NewDecoder(body.Body).Decode(&envelope))
				assert.True(t, envelope.Data.Enabled, "Enabled should have flipped from false to true")
			},
			assertDB: func(t *testing.T) {
				t.Helper()
				assertNoAnalysisRequest(t, ctx, db.Pool())
			},
		},
		{
			name: "Success: requests no-post-processing analysis when findings prioritization is enabled",
			seedFlag: func(t *testing.T) int32 {
				t.Helper()
				seedFeatureFlag(t, ctx, db.Pool(), appcfg.FeatureVariableAnalysisMode, "Variable Analysis Mode", false, true)
				return seedFeatureFlag(t, ctx, db.Pool(), services.FeatureFindingsPrioritizationV0, "Findings Prioritization v0", false, true)
			},
			expectedStatus: http.StatusOK,
			assertDB: func(t *testing.T) {
				t.Helper()
				assertAnalysisRequest(t, ctx, db.Pool(), services.PrioritizationFlagRequestSource, int32(model.AnalysisStepsNoPostProcessing().Bits()))
			},
		},
		{
			name: "Success: requests no post-processing analysis when findings prioritization is enabled and variable analysis mode is enabled",
			seedFlag: func(t *testing.T) int32 {
				t.Helper()
				seedFeatureFlag(t, ctx, db.Pool(), appcfg.FeatureVariableAnalysisMode, "Variable Analysis Mode", true, true)
				return seedFeatureFlag(t, ctx, db.Pool(), services.FeatureFindingsPrioritizationV0, "Findings Prioritization v0", false, true)
			},
			expectedStatus: http.StatusOK,
			assertDB: func(t *testing.T) {
				t.Helper()
				assertAnalysisRequest(t, ctx, db.Pool(), services.PrioritizationFlagRequestSource, int32(model.AnalysisStepsNoPostProcessing().Bits()))
			},
		},
		{
			name: "Success: does not request analysis when findings prioritization is disabled",
			seedFlag: func(t *testing.T) int32 {
				t.Helper()
				return seedFeatureFlag(t, ctx, db.Pool(), services.FeatureFindingsPrioritizationV0, "Findings Prioritization v0", true, true)
			},
			expectedStatus: http.StatusOK,
			assertDB: func(t *testing.T) {
				t.Helper()
				assertNoAnalysisRequest(t, ctx, db.Pool())
			},
		},
		{
			name: "Error: returns 403 Forbidden when the flag is not user-updatable",
			seedFlag: func(t *testing.T) int32 {
				t.Helper()
				return seedFeatureFlag(t, ctx, db.Pool(), "e2e_locked_flag", "E2E Locked Flag", false, false)
			},
			expectedStatus: http.StatusForbidden,
			assertDB: func(t *testing.T) {
				t.Helper()
				assertNoAnalysisRequest(t, ctx, db.Pool())
			},
		},
		{
			name: "Error: returns 404 Not Found when the flag does not exist",
			seedFlag: func(t *testing.T) int32 {
				t.Helper()
				return 999999
			},
			expectedStatus: http.StatusNotFound,
			assertDB: func(t *testing.T) {
				t.Helper()
				assertNoAnalysisRequest(t, ctx, db.Pool())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.NoError(t, db.DeleteAnalysisRequest(ctx))

			server := newServer(t)
			featureID := testCase.seedFlag(t)

			resp, err := http.DefaultClient.Do(newToggleRequest(t, server, featureID))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, testCase.expectedStatus, resp.StatusCode)
			if testCase.assertResponse != nil {
				testCase.assertResponse(t, resp)
			}
			if testCase.assertDB != nil {
				testCase.assertDB(t)
			}
		})
	}
}
