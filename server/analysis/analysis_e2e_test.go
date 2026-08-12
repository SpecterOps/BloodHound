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
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/analysis"
	"github.com/specterops/bloodhound/server/analysis/internal/handlers"
	"github.com/specterops/bloodhound/server/analysis/internal/services"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// analysisResponseEnvelope is the full JSON envelope shape returned by the GET
// and PUT /api/v2/analysis handlers. All seven documented fields are included so
// the tests can verify the complete contract rather than a subset.
type analysisResponseEnvelope struct {
	Data handlers.RequestedAnalysisView `json:"data"`
}

func TestGetAnalysisStatus(t *testing.T) {
	var (
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInst *router.Router, db *database.BloodhoundDB) {
			analysis.Register(routerInst, db.Pool())
		})
		db     = harness.DB
		server = harness.Server
		user   = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
			EULAAccepted:  true, // Required for permission checks to work
			Roles:         model.Roles{servertest.AdminRole(t, ctx, db)},
		}
		token = servertest.MintJWT(t, ctx, db, harness.Auther, user)
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
		require.NoError(t, db.RequestAnalysis(ctx, "test-user", model.AnalysisModeFull))
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
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInst *router.Router, db *database.BloodhoundDB) {
			analysis.Register(routerInst, db.Pool())
		})
		db     = harness.DB
		server = harness.Server
		user   = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
			EULAAccepted:  true, // Required for permission checks to work
			Roles:         model.Roles{servertest.AdminRole(t, ctx, db)},
		}
		token = servertest.MintJWT(t, ctx, db, harness.Auther, user)
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
		require.NoError(t, db.RequestAnalysis(ctx, "prior-user", model.AnalysisModeFull))
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
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInst *router.Router, db *database.BloodhoundDB) {
			analysis.Register(routerInst, db.Pool())
		})
		db     = harness.DB
		server = harness.Server
		user   = model.User{
			PrincipalName: "test-user@example.com",
			EmailAddress:  null.StringFrom("test-user@example.com"),
			EULAAccepted:  true, // Required for permission checks to work
			Roles:         model.Roles{servertest.AdminRole(t, ctx, db)},
		}
		token = servertest.MintJWT(t, ctx, db, harness.Auther, user)
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
		require.NoError(t, db.RequestAnalysis(ctx, "test-user", model.AnalysisModeFull))

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
