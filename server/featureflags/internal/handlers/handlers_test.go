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

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/featureflags/internal/handlers"
	"github.com/specterops/bloodhound/server/featureflags/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
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

// withFeatureIDVar attaches the {feature_id} mux path variable to the supplied
// request so the handler under test can resolve it via mux.Vars.
func withFeatureIDVar(req *http.Request, featureID string) *http.Request {
	return mux.SetURLVars(req, map[string]string{api.URIPathVariableFeatureID: featureID})
}

func TestHandlers_GetAllFlags(t *testing.T) {
	var (
		unexpectedErr = errors.New("unexpected database failure")
		flags         = []services.FeatureFlag{
			{ID: 1, Key: services.FeatureOpenHoundSupport, Name: "OpenHound", Enabled: true},
			{ID: 2, Key: services.FeatureAlerts, Name: "Alerts", Enabled: false, UserUpdatable: true},
		}
	)

	tests := []struct {
		name       string
		svcResult  []services.FeatureFlag
		svcErr     error
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:       "returns 200 with the feature flags view on success",
			svcResult:  flags,
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.FeatureFlagsView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				require.Len(t, envelope.Data, 2)
				assert.Equal(t, int32(1), envelope.Data[0].ID)
				assert.Equal(t, services.FeatureOpenHoundSupport, envelope.Data[0].Key)
				assert.True(t, envelope.Data[1].UserUpdatable)
			},
		},
		{
			name:       "returns an empty list when the service returns no flags",
			svcResult:  []services.FeatureFlag{},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.FeatureFlagsView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Empty(t, envelope.Data)
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
				featureFlagMock = mocks.NewMockFeatureFlag(t)
				handlerSet      = handlers.NewHandlersContainer(featureFlagMock)
				recorder        = httptest.NewRecorder()
				request         = httptest.NewRequest(http.MethodGet, "/api/v2/features", nil)
			)

			featureFlagMock.EXPECT().GetAllFlags(request.Context()).Return(tt.svcResult, tt.svcErr)

			handlerSet.GetAllFlags(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestHandlers_ToggleFlag(t *testing.T) {
	var (
		userID       = uuid.Must(uuid.NewV4())
		userIDString = userID.String()
		toggled      = services.FeatureFlag{ID: 5, Key: services.FeatureAlerts, Name: "Alerts", Enabled: true, UserUpdatable: true}
		serviceErr   = errors.New("db unavailable")
	)

	tests := []struct {
		name          string
		featureID     string
		authenticated bool
		expect        func(m *mocks.MockFeatureFlag, ctx context.Context)
		wantStatus    int
		assertBody    func(t *testing.T, body []byte)
	}{
		{
			name:          "returns 200 OK and the flag view on success",
			featureID:     "5",
			authenticated: true,
			expect: func(m *mocks.MockFeatureFlag, ctx context.Context) {
				m.EXPECT().ToggleFlag(ctx, int32(5), userIDString).Return(toggled, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.FeatureFlagView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, int32(5), envelope.Data.ID)
				assert.True(t, envelope.Data.Enabled)
			},
		},
		{
			name:          "returns 500 when the request has no authenticated user",
			featureID:     "5",
			authenticated: false,
			wantStatus:    http.StatusInternalServerError,
		},
		{
			name:          "returns 400 when feature_id is not parseable as an int32",
			featureID:     "not-a-number",
			authenticated: true,
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "returns 404 when the service reports ErrNotFound",
			featureID:     "5",
			authenticated: true,
			expect: func(m *mocks.MockFeatureFlag, ctx context.Context) {
				m.EXPECT().ToggleFlag(ctx, int32(5), userIDString).Return(services.FeatureFlag{}, services.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "returns 403 when the service reports ErrNotUserUpdatable",
			featureID:     "5",
			authenticated: true,
			expect: func(m *mocks.MockFeatureFlag, ctx context.Context) {
				m.EXPECT().ToggleFlag(ctx, int32(5), userIDString).Return(services.FeatureFlag{}, services.ErrNotUserUpdatable)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:          "returns 500 on unexpected service errors",
			featureID:     "5",
			authenticated: true,
			expect: func(m *mocks.MockFeatureFlag, ctx context.Context) {
				m.EXPECT().ToggleFlag(ctx, int32(5), userIDString).Return(services.FeatureFlag{}, serviceErr)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				featureFlagMock = mocks.NewMockFeatureFlag(t)
				handlerSet      = handlers.NewHandlersContainer(featureFlagMock)
				recorder        = httptest.NewRecorder()
				request         *http.Request
			)

			if tt.authenticated {
				request = newAuthenticatedRequest(t, http.MethodPut, "/api/v2/features/"+tt.featureID+"/toggle", userID)
			} else {
				req, err := http.NewRequest(http.MethodPut, "/api/v2/features/"+tt.featureID+"/toggle", nil)
				require.NoError(t, err)
				request = req
			}
			request = withFeatureIDVar(request, tt.featureID)

			if tt.expect != nil {
				tt.expect(featureFlagMock, request.Context())
			}

			handlerSet.ToggleFlag(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestHandlers_IsEnabled(t *testing.T) {
	var (
		ctx        = context.Background()
		serviceErr = errors.New("connection refused")
	)

	tests := []struct {
		name        string
		expect      func(m *mocks.MockFeatureFlag)
		wantEnabled bool
		wantErr     error
	}{
		{
			name: "delegates to the service and returns enabled=true",
			expect: func(m *mocks.MockFeatureFlag) {
				m.EXPECT().IsEnabled(ctx, services.FeatureOpenHoundSupport).Return(true, nil)
			},
			wantEnabled: true,
		},
		{
			name: "propagates service errors",
			expect: func(m *mocks.MockFeatureFlag) {
				m.EXPECT().IsEnabled(ctx, services.FeatureOpenHoundSupport).Return(false, serviceErr)
			},
			wantErr: serviceErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				featureFlagMock = mocks.NewMockFeatureFlag(t)
				handlerSet      = handlers.NewHandlersContainer(featureFlagMock)
			)

			tt.expect(featureFlagMock)

			enabled, err := handlerSet.IsEnabled(ctx, services.FeatureOpenHoundSupport)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantEnabled, enabled)
		})
	}
}
