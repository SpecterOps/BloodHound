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

package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/dataquality/internal/handlers"
	"github.com/specterops/bloodhound/server/dataquality/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAggregations(t *testing.T) {
	t.Parallel()
	var (
		start       = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		end         = start.Add(time.Hour)
		extensionID = int32(42)
		query       = services.AggregationQuery{EnvironmentKindID: 10, ExtensionID: &extensionID, Start: start, End: end, Sort: params.SortItems{{Field: "updated_at", Direction: params.Descending}}, Skip: 3, Limit: 7}
		baseURL     = "/api/v2/data-quality-stats-aggregations?schema_environment_kind_id=10&schema_extension_id=42&start=2026-01-01T00:00:00Z&end=2026-01-01T01:00:00Z"
	)
	testCases := []struct {
		name      string
		url       string
		setupMock func(*mocks.MockDataQuality, *http.Request)
		status    int
	}{
		{name: "forwards query and builds envelope", url: baseURL, status: 200, setupMock: func(service *mocks.MockDataQuality, request *http.Request) {
			service.EXPECT().ListAggregations(request.Context(), query).Return([]services.Aggregation{{ID: 1, RunID: "run", SchemaExtensionID: 42, SchemaEnvironmentKindID: 10, MetricType: "count", MetricName: "nodes", MetricValue: 1.5, CreatedAt: start, UpdatedAt: end}}, 42, nil).Once()
		}},
		{name: "missing extension", url: baseURL, status: 404, setupMock: func(service *mocks.MockDataQuality, request *http.Request) {
			service.EXPECT().ListAggregations(request.Context(), query).Return(nil, 0, services.ErrExtensionNotFound).Once()
		}},
		{name: "database fails", url: baseURL, status: 500, setupMock: func(service *mocks.MockDataQuality, request *http.Request) {
			service.EXPECT().ListAggregations(request.Context(), query).Return(nil, 0, errors.New("private database failure")).Once()
		}},
		{name: "invalid query", url: baseURL, status: 400, setupMock: func(service *mocks.MockDataQuality, request *http.Request) {
			service.EXPECT().ListAggregations(request.Context(), query).Return(nil, 0, services.ErrInvalidQuery).Once()
		}},
		{name: "missing kind", url: "/", status: 400},
		{name: "invalid kind", url: "/?schema_environment_kind_id=abc", status: 400},
		{name: "kind overflow", url: "/?schema_environment_kind_id=2147483648", status: 400},
		{name: "extension overflow", url: "/?schema_environment_kind_id=10&schema_extension_id=2147483648", status: 400},
		{name: "invalid start", url: "/?schema_environment_kind_id=10&start=bad", status: 400},
		{name: "invalid end", url: "/?schema_environment_kind_id=10&end=bad", status: 400},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				service           = mocks.NewMockDataQuality(t)
				request           = httptest.NewRequest(http.MethodGet, testCase.url, nil)
				bloodhoundContext = &bhctx.Context{Sort: query.Sort, Skip: 3, Limit: 7}
				recorder          = httptest.NewRecorder()
			)
			request = request.WithContext(bhctx.Set(request.Context(), bloodhoundContext))
			if testCase.setupMock != nil {
				testCase.setupMock(service, request)
			}
			handlers.NewHandlersContainer(service).ListAggregations(recorder, request)
			require.Equal(t, testCase.status, recorder.Code)
			if testCase.status == 200 {
				var view handlers.AggregationsView
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &view))
				require.NotNil(t, view.Start)
				assert.Equal(t, start, *view.Start)
				require.NotNil(t, view.End)
				assert.Equal(t, end, *view.End)
				assert.Equal(t, 42, view.Count)
				assert.Equal(t, 7, view.Limit)
				assert.Equal(t, 3, view.Skip)
				require.Len(t, view.Data, 1)
				assert.Equal(t, 1.5, view.Data[0].MetricValue)
				assert.Nil(t, view.Data[0].KindID)
			} else {
				assert.Contains(t, recorder.Body.String(), `"http_status":`)
				assert.NotContains(t, recorder.Body.String(), "private database failure")
			}
		})
	}
}
