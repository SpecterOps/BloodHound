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

package dataquality_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/specterops/bloodhound/server/modules"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataQualityAggregationsContract(t *testing.T) {
	var (
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInst *router.Router, databaseInstance *database.BloodhoundDB) {
			modules.Register(modules.Deps{
				Router: routerInst, Pool: databaseInstance.Pool(), Graph: &graph.DatabaseSwitch{},
				DogTags: dogtags.NewTestService(dogtags.TestOverrides{}), ConfigUpdateSideEffects: databaseInstance,
				RateLimitMiddleware: func() mux.MiddlewareFunc { return middleware.DefaultRateLimitMiddleware(databaseInstance) },
			})
		})
		databaseInstance = harness.DB
		pool             = databaseInstance.Pool()
		createdAt        = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		kindID           int32
		extensionID      int32
	)
	flag, err := databaseInstance.GetFlagByKey(ctx, appcfg.FeatureOpenGraphExtensionManagement)
	require.NoError(t, err)
	flag.Enabled = true
	require.NoError(t, databaseInstance.SetFlag(ctx, flag))
	require.NoError(t, pool.QueryRow(ctx, "INSERT INTO kind (name) VALUES ('AggregationTestEnvironment') RETURNING id").Scan(&kindID))
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace) VALUES ('AggregationTest', 'Aggregation Test', '1.0.0', false, 'AGT') RETURNING id`).Scan(&extensionID))
	for index := range 3 {
		_, err = pool.Exec(ctx, `INSERT INTO data_quality_aggregations (run_id, schema_extension_id, schema_environment_kind_id, metric_type, metric_name, metric_value, kind_id, created_at, updated_at) VALUES ($1, $2, $3, 'count', 'nodes', $4, $5, $6, $6)`, fmt.Sprintf("run-%d", index), extensionID, kindID, index+10, kindID, createdAt.Add(time.Duration(index)*time.Hour))
		require.NoError(t, err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO data_quality_aggregations (run_id, schema_extension_id, schema_environment_kind_id, metric_type, metric_name, created_at, deleted_at) VALUES ('deleted', $1, $2, 'count', 'nodes', $3, $3)`, extensionID, kindID, createdAt)
	require.NoError(t, err)
	token := servertest.MintJWT(t, ctx, databaseInstance, harness.Auther, model.User{
		PrincipalName: "aggregation-admin@example.com", EULAAccepted: true,
		EmailAddress: null.StringFrom("aggregation-admin@example.com"), Roles: model.Roles{servertest.AdminRole(t, ctx, databaseInstance)},
	})
	noPermissionToken := servertest.MintJWT(t, ctx, databaseInstance, harness.Auther, model.User{PrincipalName: "aggregation-no-role@example.com", EULAAccepted: true})
	baseQuery := fmt.Sprintf("?schema_environment_kind_id=%d&start=2026-01-02T00:00:00Z&end=2026-01-02T02:00:00Z", kindID)
	testCases := []struct {
		name   string
		query  string
		token  string
		status int
		runs   []string
		count  int
		limit  int
		skip   int
	}{
		{name: "unauthenticated", query: baseQuery, status: 401},
		{name: "missing permission", query: baseQuery, token: noPermissionToken, status: 403},
		{name: "missing kind", token: token, status: 400},
		{name: "invalid kind", query: "?schema_environment_kind_id=abc", token: token, status: 400},
		{name: "kind overflow", query: "?schema_environment_kind_id=2147483648", token: token, status: 400},
		{name: "invalid extension", query: baseQuery + "&schema_extension_id=abc", token: token, status: 400},
		{name: "missing extension", query: baseQuery + "&schema_extension_id=2147483647", token: token, status: 404},
		{name: "invalid sort", query: baseQuery + "&sort_by=metric_name", token: token, status: 400},
		{name: "empty sort", query: baseQuery + "&sort_by=", token: token, status: 400},
		{name: "negative skip", query: baseQuery + "&skip=-1", token: token, status: 400},
		{name: "invalid skip", query: baseQuery + "&skip=abc", token: token, status: 400},
		{name: "negative limit", query: baseQuery + "&limit=-1", token: token, status: 400},
		{name: "invalid limit", query: baseQuery + "&limit=abc", token: token, status: 400},
		{name: "invalid start", query: fmt.Sprintf("?schema_environment_kind_id=%d&start=bad", kindID), token: token, status: 400},
		{name: "invalid end", query: fmt.Sprintf("?schema_environment_kind_id=%d&end=bad", kindID), token: token, status: 400},
		{name: "inclusive window and default sorting", query: baseQuery, token: token, status: 200, runs: []string{"run-2", "run-1", "run-0"}, count: 3, limit: 1000},
		{name: "extension filter and pagination", query: baseQuery + fmt.Sprintf("&schema_extension_id=%d&limit=1&skip=1&sort_by=created_at", extensionID), token: token, status: 200, runs: []string{"run-1"}, count: 3, limit: 1, skip: 1},
		{name: "zero limit is unlimited", query: baseQuery + "&limit=0&sort_by=-updated_at", token: token, status: 200, runs: []string{"run-2", "run-1", "run-0"}, count: 3},
		{name: "unknown kind returns null", query: "?schema_environment_kind_id=2147483647", token: token, status: 200, runs: []string{}, limit: 1000},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, harness.Server.URL+"/api/v2/data-quality-stats-aggregations"+testCase.query, nil)
			require.NoError(t, err)
			if testCase.token != "" {
				request.Header.Set("Authorization", "Bearer "+testCase.token)
			}
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, testCase.status, response.StatusCode)
			if testCase.status != http.StatusOK {
				return
			}
			var body struct {
				Data  []map[string]any `json:"data"`
				Start time.Time        `json:"start"`
				End   time.Time        `json:"end"`
				Count int              `json:"count"`
				Limit int              `json:"limit"`
				Skip  int              `json:"skip"`
			}
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			if len(testCase.runs) == 0 {
				assert.Nil(t, body.Data)
			} else {
				require.NotNil(t, body.Data)
			}
			assert.Equal(t, testCase.count, body.Count)
			assert.Equal(t, testCase.limit, body.Limit)
			assert.Equal(t, testCase.skip, body.Skip)
			require.Len(t, body.Data, len(testCase.runs))
			for index, run := range testCase.runs {
				row := body.Data[index]
				assert.Equal(t, run, row["run_id"])
				assert.Equal(t, float64(extensionID), row["extension_id"])
				assert.Equal(t, float64(kindID), row["environment_kind_id"])
				assert.Equal(t, "nodes", row["metric_name"])
				assert.Equal(t, "count", row["metric_type"])
				assert.Equal(t, float64(kindID), row["kind_id"])
				assert.Contains(t, row, "id")
				assert.Contains(t, row, "metric_value")
				assert.Contains(t, row, "created_at")
				assert.Contains(t, row, "updated_at")
				assert.Equal(t, map[string]any{"Time": "0001-01-01T00:00:00Z", "Valid": false}, row["deleted_at"])
			}
			if len(testCase.runs) > 0 {
				assert.Equal(t, createdAt, body.Start)
				assert.Equal(t, createdAt.Add(2*time.Hour), body.End)
			} else {
				assert.InDelta(t, (30 * 24 * time.Hour).Seconds(), body.End.Sub(body.Start).Seconds(), 1)
			}
		})
	}
	t.Run("disabled feature flag", func(t *testing.T) {
		flag.Enabled = false
		require.NoError(t, databaseInstance.SetFlag(ctx, flag))
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, harness.Server.URL+"/api/v2/data-quality-stats-aggregations"+baseQuery, nil)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+token)
		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		defer response.Body.Close()
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	})
}
