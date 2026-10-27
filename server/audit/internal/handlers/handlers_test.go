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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/audit/internal/handlers"
	"github.com/specterops/bloodhound/server/audit/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_ListAuditLogs(t *testing.T) {
	var (
		before       = time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
		after        = before.Add(-time.Hour)
		serviceError = errors.New("database unavailable")
	)

	tests := []struct {
		name             string
		url              string
		requestContext   *bhctx.Context
		setupMocks       func(audit *mocks.MockAudit)
		expectedCode     int
		expectedContains string
	}{
		{
			name:           "Success: returns paginated audit logs - 200",
			url:            "/api/v2/audit?before=" + before.Format(time.RFC3339Nano) + "&after=" + after.Format(time.RFC3339Nano),
			requestContext: &bhctx.Context{Skip: 2, Limit: 10},
			setupMocks: func(audit *mocks.MockAudit) {
				audit.EXPECT().ListAuditLogs(mock.Anything, services.ListQuery{
					Before: before, After: after, Skip: 2, Limit: 10,
					Sort: params.SortItems{{Field: "created_at", Direction: params.Descending}},
				}).Return([]services.Log{{ID: 7, Action: "CreateUser", Fields: map[string]any{}}}, 12, nil)
			},
			expectedCode: http.StatusOK, expectedContains: `"count":12`,
		},
		{
			name:           "Success: supports the legacy offset alias - 200",
			url:            "/api/v2/audit?before=" + before.Format(time.RFC3339Nano) + "&after=" + after.Format(time.RFC3339Nano) + "&offset=4",
			requestContext: &bhctx.Context{Limit: 10},
			setupMocks: func(audit *mocks.MockAudit) {
				audit.EXPECT().ListAuditLogs(mock.Anything, mock.MatchedBy(func(query services.ListQuery) bool {
					return query.Skip == 4 && query.Limit == 10
				})).Return([]services.Log{}, 0, nil)
			},
			expectedCode: http.StatusOK, expectedContains: `"skip":4`,
		},
		{
			name:           "Error: after is later than before - 400",
			url:            "/api/v2/audit?before=" + after.Format(time.RFC3339Nano) + "&after=" + before.Format(time.RFC3339Nano),
			requestContext: &bhctx.Context{Limit: 10},
			setupMocks:     func(audit *mocks.MockAudit) {},
			expectedCode:   http.StatusBadRequest, expectedContains: "must not be later",
		},
		{
			name:           "Error: service failure - 500",
			url:            "/api/v2/audit?before=" + before.Format(time.RFC3339Nano) + "&after=" + after.Format(time.RFC3339Nano),
			requestContext: &bhctx.Context{Limit: 10},
			setupMocks: func(audit *mocks.MockAudit) {
				audit.EXPECT().ListAuditLogs(mock.Anything, mock.Anything).Return(nil, 0, serviceError)
			},
			expectedCode: http.StatusInternalServerError, expectedContains: "internal error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			audit := mocks.NewMockAudit(t)
			testCase.setupMocks(audit)
			handlerSet := handlers.NewHandlersContainer(audit)
			request := httptest.NewRequest(http.MethodGet, testCase.url, nil)
			request = request.WithContext(bhctx.Set(context.Background(), testCase.requestContext))
			response := httptest.NewRecorder()

			handlerSet.ListAuditLogs(response, request)

			assert.Equal(t, testCase.expectedCode, response.Code)
			assert.Contains(t, response.Body.String(), testCase.expectedContains)
		})
	}
}
