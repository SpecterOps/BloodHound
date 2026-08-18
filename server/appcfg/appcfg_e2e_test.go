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
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
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
