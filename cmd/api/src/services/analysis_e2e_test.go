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

//go:build e2e

package services_test

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
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	appdbAnalysis "github.com/specterops/bloodhound/server/appdb/analysis"
	analysisHandlers "github.com/specterops/bloodhound/server/handlers/v2/analysis"
	analysisService "github.com/specterops/bloodhound/server/services/analysis"
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

// analysisStatusEnvelope is a minimal JSON envelope covering the fields
// returned by the GET /api/v2/analysis handler.
type analysisStatusEnvelope struct {
	Data struct {
		RequestedBy string `json:"requested_by"`
		RequestType string `json:"request_type"`
	} `json:"data"`
}

// runGetAnalysisStatusSuite exercises GET /api/v2/analysis against a handler.
// A fresh httptest.Server is started without auth middleware so the test
// focuses on handler behaviour.
func runGetAnalysisStatusSuite(t *testing.T, db *database.BloodhoundDB, handler http.HandlerFunc) {
	t.Helper()

	var (
		ctx       = context.Background()
		muxRouter = mux.NewRouter()
		server    = httptest.NewServer(muxRouter)
	)

	muxRouter.HandleFunc("/api/v2/analysis", handler).Methods(http.MethodGet)
	t.Cleanup(server.Close)

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 204 when no request is pending", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("returns 200 with requester details when a request is pending", func(t *testing.T) {
		require.NoError(t, db.RequestAnalysis(ctx, "test-user"))

		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var envelope analysisStatusEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
		assert.Equal(t, "test-user", envelope.Data.RequestedBy)
		assert.Equal(t, "analysis", envelope.Data.RequestType)
	})
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

// newAnalysisHandler wires the pgx-backed analysis stack from a BloodhoundDB
// and returns its GET handler ready to pass to runGetAnalysisStatusSuite.
func newAnalysisHandler(db *database.BloodhoundDB) http.HandlerFunc {
	store := appdbAnalysis.NewStore(db.Pool())
	svc := analysisService.NewService(store)
	return analysisHandlers.NewHandlersContainer(svc).GetRequest
}

func TestGetAnalysisStatus(t *testing.T) {
	t.Run("new handler", func(t *testing.T) {
		db := setupAnalysisDB(t)
		runGetAnalysisStatusSuite(t, db, newAnalysisHandler(db))
	})
}

// newCreateAnalysisHandler wires the pgx-backed analysis stack from a BloodhoundDB
// and returns its PUT handler (wrapped with auth-injecting middleware) ready to
// pass to runCreateAnalysisRequestSuite.
func newCreateAnalysisHandler(db *database.BloodhoundDB, userID uuid.UUID) http.HandlerFunc {
	store := appdbAnalysis.NewStore(db.Pool())
	svc := analysisService.NewService(store)
	return injectAuthMiddleware(analysisHandlers.NewHandlersContainer(svc).CreateRequest, userID)
}

// runCreateAnalysisRequestSuite exercises PUT /api/v2/analysis.
// It shares the same GET handler infrastructure to verify end-to-end state.
func runCreateAnalysisRequestSuite(t *testing.T, db *database.BloodhoundDB, userID uuid.UUID, putHandler http.HandlerFunc) {
	t.Helper()

	var (
		ctx       = context.Background()
		muxRouter = mux.NewRouter()
		server    = httptest.NewServer(muxRouter)
	)

	muxRouter.HandleFunc("/api/v2/analysis", putHandler).Methods(http.MethodPut)
	muxRouter.HandleFunc("/api/v2/analysis", newAnalysisHandler(db)).Methods(http.MethodGet)
	t.Cleanup(server.Close)

	newPutRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		return req
	}

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/analysis", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 202 Accepted and persists the new request on first PUT", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))

		resp, err := http.DefaultClient.Do(newPutRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		var envelope analysisStatusEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
		assert.Equal(t, userID.String(), envelope.Data.RequestedBy)
		assert.Equal(t, "analysis", envelope.Data.RequestType)
	})

	t.Run("returns 200 OK with the existing request on a second PUT", func(t *testing.T) {
		// Following the previous subtest, a request is already pending. A second
		// PUT must be idempotent and return the existing request with 200 OK.
		resp, err := http.DefaultClient.Do(newPutRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET reflects pending request after PUT", func(t *testing.T) {
		require.NoError(t, db.DeleteAnalysisRequest(ctx))

		resp, err := http.DefaultClient.Do(newPutRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		resp, err = http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var envelope analysisStatusEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
		assert.Equal(t, "analysis", envelope.Data.RequestType)
	})
}

func TestCreateAnalysisRequest(t *testing.T) {
	t.Run("new handler", func(t *testing.T) {
		var (
			db     = setupAnalysisDB(t)
			userID = uuid.Must(uuid.NewV4())
		)
		runCreateAnalysisRequestSuite(t, db, userID, newCreateAnalysisHandler(db, userID))
	})
}
