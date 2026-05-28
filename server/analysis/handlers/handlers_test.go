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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/analysis/handlers"
	"github.com/specterops/bloodhound/server/analysis/handlers/mocks"
	"github.com/specterops/bloodhound/server/analysis/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAuthenticatedRequest returns an *http.Request whose context carries a
// bhctx.Context with the supplied user wired in as the auth Owner. This mirrors
// what the auth middleware does for real requests.
func newAuthenticatedRequest(t *testing.T, method, target string, userID uuid.UUID) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, target, nil)
	require.NoError(t, err)
	bhCtx := &bhctx.Context{
		AuthCtx: auth.Context{Owner: model.User{Unique: model.Unique{ID: userID}}},
	}
	return bhctx.SetRequestContext(req, bhCtx)
}

func TestHandlers_GetRequest(t *testing.T) {
	newRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, "/api/v2/analysis/status", nil)
		require.NoError(t, err)
		return req
	}

	var (
		unexpectedErr = errors.New("unexpected database failure")
		expected      = services.RequestedAnalysis{
			RequestedBy:         "test-user",
			RequestType:         services.RequestedAnalysisTypeAnalysis,
			RequestedAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DeleteSourceKinds:   []string{"AZBase"},
			DeleteRelationships: []string{"HasSession"},
		}
	)

	tests := []struct {
		name       string
		svcResult  services.RequestedAnalysis
		svcErr     error
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:       "returns 200 with the analysis request view on success",
			svcResult:  expected,
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RequestedAnalysisView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, expected.RequestedBy, envelope.Data.RequestedBy)
				assert.Equal(t, expected.RequestType, envelope.Data.RequestType)
				assert.Equal(t, expected.DeleteSourceKinds, envelope.Data.DeleteSourceKinds)
				assert.Equal(t, expected.DeleteRelationships, envelope.Data.DeleteRelationships)
			},
		},
		{
			name:       "returns 204 No Content when no request is pending",
			svcErr:     services.ErrNoPendingRequest,
			wantStatus: http.StatusNoContent,
			assertBody: func(t *testing.T, body []byte) {
				assert.Empty(t, body)
			},
		},
		{
			name:       "returns 500 on unexpected service errors",
			svcErr:     unexpectedErr,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				analysisMock = mocks.NewMockAnalysis(t)
				handlerSet   = handlers.NewHandlersContainer(analysisMock)
				recorder     = httptest.NewRecorder()
				request      = newRequest(t)
			)

			analysisMock.EXPECT().GetRequest(request.Context()).Return(tt.svcResult, tt.svcErr)

			handlerSet.GetRequest(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestHandlers_CreateRequest(t *testing.T) {
	var (
		userID         = uuid.Must(uuid.NewV4())
		userIDString   = userID.String()
		requestedAt    = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		serviceErr     = errors.New("db unavailable")
		createdResult  = services.RequestedAnalysis{RequestedBy: userIDString, RequestType: services.RequestedAnalysisTypeAnalysis, RequestedAt: requestedAt}
		existingResult = services.RequestedAnalysis{RequestedBy: "other-user", RequestType: services.RequestedAnalysisTypeAnalysis, RequestedAt: requestedAt}
	)

	decodeRequestedBy := func(t *testing.T, body []byte) string {
		t.Helper()
		var envelope struct {
			Data handlers.RequestedAnalysisView `json:"data"`
		}
		require.NoError(t, json.Unmarshal(body, &envelope))
		return envelope.Data.RequestedBy
	}

	// expect configures the mock for the service call the handler is expected
	// to make. nil means the handler should short-circuit before reaching the
	// service (mockery's strict mode then guarantees no call is made).
	tests := []struct {
		name             string
		authenticated    bool
		expect           func(m *mocks.MockAnalysis, ctx context.Context)
		wantStatus       int
		wantRequestedBy  string // empty means do not assert on body
		wantBodyContains string
	}{
		{
			name:          "returns 202 Accepted with the new request body when this call accepted it",
			authenticated: true,
			expect: func(m *mocks.MockAnalysis, ctx context.Context) {
				m.EXPECT().CreateRequest(ctx, userIDString).Return(createdResult, true, nil)
			},
			wantStatus:      http.StatusAccepted,
			wantRequestedBy: userIDString,
		},
		{
			name:          "returns 200 OK with the existing request body when one was already pending",
			authenticated: true,
			expect: func(m *mocks.MockAnalysis, ctx context.Context) {
				m.EXPECT().CreateRequest(ctx, userIDString).Return(existingResult, false, nil)
			},
			wantStatus:      http.StatusOK,
			wantRequestedBy: existingResult.RequestedBy,
		},
		{
			// The route middleware (RequirePermissions) guarantees an authenticated
			// caller, so this branch is only reachable if something upstream is broken.
			// The handler treats that as an internal error rather than a 401.
			name:          "returns 500 when the auth context has no user (defensive guard past middleware)",
			authenticated: false,
			wantStatus:    http.StatusInternalServerError,
		},
		{
			name:          "returns 500 on service error",
			authenticated: true,
			expect: func(m *mocks.MockAnalysis, ctx context.Context) {
				m.EXPECT().CreateRequest(ctx, userIDString).Return(services.RequestedAnalysis{}, false, serviceErr)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				analysisMock = mocks.NewMockAnalysis(t)
				handlerSet   = handlers.NewHandlersContainer(analysisMock)
				recorder     = httptest.NewRecorder()
				request      *http.Request
			)

			if tt.authenticated {
				request = newAuthenticatedRequest(t, http.MethodPut, "/api/v2/analysis", userID)
			} else {
				req, err := http.NewRequest(http.MethodPut, "/api/v2/analysis", nil)
				require.NoError(t, err)
				request = req
			}
			if tt.expect != nil {
				tt.expect(analysisMock, request.Context())
			}

			handlerSet.CreateRequest(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.wantRequestedBy != "" {
				assert.Equal(t, tt.wantRequestedBy, decodeRequestedBy(t, recorder.Body.Bytes()))
			}
			if tt.wantBodyContains != "" {
				assert.Contains(t, recorder.Body.String(), tt.wantBodyContains)
			}
		})
	}
}
