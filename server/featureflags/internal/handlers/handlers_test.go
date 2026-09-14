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
	handlersmocks "github.com/specterops/bloodhound/server/featureflags/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
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
	t.Parallel()

	type mock struct {
		service *handlersmocks.MockFeatureFlag
	}
	type args struct {
		buildRequest func() *http.Request
	}
	type want struct {
		responseCode int
		assertBody   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	var (
		unexpectedErr = errors.New("unexpected database failure")
		flags         = []services.FeatureFlag{
			{ID: 1, Key: services.FeatureOpenHoundSupport, Name: "OpenHound", Enabled: true},
			{ID: 2, Key: services.FeatureAlerts, Name: "Alerts", Enabled: false, UserUpdatable: true},
		}
	)

	tests := []struct {
		name       string
		args       args
		setupMocks func(t *testing.T, m *mock)
		want       want
	}{
		{
			name: "Success: returns the feature flags view - 200",
			args: args{
				buildRequest: func() *http.Request {
					return httptest.NewRequest(http.MethodGet, "/api/v2/features", nil)
				},
			},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().GetAllFlags(testifyMock.Anything).Return(flags, nil)
			},
			want: want{
				responseCode: http.StatusOK,
				assertBody: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					var envelope struct {
						Data handlers.FeatureFlagsView `json:"data"`
					}
					require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
					require.Len(t, envelope.Data, 2)
					assert.Equal(t, int32(1), envelope.Data[0].ID)
					assert.Equal(t, services.FeatureOpenHoundSupport, envelope.Data[0].Key)
					assert.True(t, envelope.Data[1].UserUpdatable)
				},
			},
		},
		{
			name: "Success: returns an empty list when the service returns no flags - 200",
			args: args{
				buildRequest: func() *http.Request {
					return httptest.NewRequest(http.MethodGet, "/api/v2/features", nil)
				},
			},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().GetAllFlags(testifyMock.Anything).Return([]services.FeatureFlag{}, nil)
			},
			want: want{
				responseCode: http.StatusOK,
				assertBody: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					var envelope struct {
						Data handlers.FeatureFlagsView `json:"data"`
					}
					require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
					assert.Empty(t, envelope.Data)
				},
			},
		},
		{
			name: "Error: unexpected service error - 500",
			args: args{
				buildRequest: func() *http.Request {
					return httptest.NewRequest(http.MethodGet, "/api/v2/features", nil)
				},
			},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().GetAllFlags(testifyMock.Anything).Return(nil, unexpectedErr)
			},
			want: want{
				responseCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var (
				m = &mock{
					service: handlersmocks.NewMockFeatureFlag(t),
				}
				handler  = handlers.NewHandlersContainer(m.service)
				recorder = httptest.NewRecorder()
				request  = tt.args.buildRequest()
			)
			tt.setupMocks(t, m)
			handler.GetAllFlags(recorder, request)
			assert.Equal(t, tt.want.responseCode, recorder.Code)
			if tt.want.assertBody != nil {
				tt.want.assertBody(t, recorder)
			}
		})
	}
}

func TestHandlers_ToggleFlag(t *testing.T) {
	t.Parallel()

	type mock struct {
		service *handlersmocks.MockFeatureFlag
	}
	type args struct {
		buildRequest func() *http.Request
	}
	type want struct {
		responseCode int
		assertBody   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	var (
		userID     = uuid.Must(uuid.NewV4())
		toggled    = services.FeatureFlag{ID: 5, Key: services.FeatureAlerts, Name: "Alerts", Enabled: true, UserUpdatable: true}
		serviceErr = errors.New("db unavailable")
	)

	tests := []struct {
		name       string
		args       args
		setupMocks func(t *testing.T, m *mock)
		want       want
	}{
		{
			name: "Success: toggles the flag and returns the view - 200",
			args: args{
				buildRequest: func() *http.Request {
					return withFeatureIDVar(newAuthenticatedRequest(t, http.MethodPut, "/api/v2/features/5/toggle", userID), "5")
				},
			},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().ToggleFlag(testifyMock.Anything, int32(5)).Return(toggled, nil)
			},
			want: want{
				responseCode: http.StatusOK,
				assertBody: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					var envelope struct {
						Data handlers.FeatureFlagView `json:"data"`
					}
					require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
					assert.Equal(t, int32(5), envelope.Data.ID)
					assert.True(t, envelope.Data.Enabled)
				},
			},
		},
		{
			name: "Error: feature_id is not parseable as an int32 - 400",
			args: args{
				buildRequest: func() *http.Request {
					return withFeatureIDVar(newAuthenticatedRequest(t, http.MethodPut, "/api/v2/features/not-a-number/toggle", userID), "not-a-number")
				},
			},
			setupMocks: func(t *testing.T, m *mock) {},
			want: want{
				responseCode: http.StatusBadRequest,
			},
		},
		{
			name: "Error: service reports ErrNotFound - 404",
			args: args{
				buildRequest: func() *http.Request {
					return withFeatureIDVar(newAuthenticatedRequest(t, http.MethodPut, "/api/v2/features/5/toggle", userID), "5")
				},
			},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().ToggleFlag(testifyMock.Anything, int32(5)).Return(services.FeatureFlag{}, services.ErrNotFound)
			},
			want: want{
				responseCode: http.StatusNotFound,
			},
		},
		{
			name: "Error: service reports ErrNotUserUpdatable - 403",
			args: args{
				buildRequest: func() *http.Request {
					return withFeatureIDVar(newAuthenticatedRequest(t, http.MethodPut, "/api/v2/features/5/toggle", userID), "5")
				},
			},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().ToggleFlag(testifyMock.Anything, int32(5)).Return(services.FeatureFlag{}, services.ErrNotUserUpdatable)
			},
			want: want{
				responseCode: http.StatusForbidden,
			},
		},
		{
			name: "Error: unexpected service error - 500",
			args: args{
				buildRequest: func() *http.Request {
					return withFeatureIDVar(newAuthenticatedRequest(t, http.MethodPut, "/api/v2/features/5/toggle", userID), "5")
				},
			},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().ToggleFlag(testifyMock.Anything, int32(5)).Return(services.FeatureFlag{}, serviceErr)
			},
			want: want{
				responseCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var (
				m = &mock{
					service: handlersmocks.NewMockFeatureFlag(t),
				}
				handler  = handlers.NewHandlersContainer(m.service)
				recorder = httptest.NewRecorder()
				request  = tt.args.buildRequest()
			)
			tt.setupMocks(t, m)
			handler.ToggleFlag(recorder, request)
			assert.Equal(t, tt.want.responseCode, recorder.Code)
			if tt.want.assertBody != nil {
				tt.want.assertBody(t, recorder)
			}
		})
	}
}

func TestHandlers_IsEnabled(t *testing.T) {
	t.Parallel()

	type mock struct {
		service *handlersmocks.MockFeatureFlag
	}
	type args struct {
		ctx        context.Context
		featureKey string
	}
	type want struct {
		enabled bool
		err     error
	}

	var (
		ctx        = context.Background()
		serviceErr = errors.New("connection refused")
	)

	tests := []struct {
		name       string
		args       args
		setupMocks func(t *testing.T, m *mock)
		want       want
	}{
		{
			name: "Success: delegates to the service and returns enabled=true",
			args: args{ctx: ctx, featureKey: services.FeatureOpenHoundSupport},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().IsEnabled(ctx, services.FeatureOpenHoundSupport).Return(true, nil)
			},
			want: want{enabled: true},
		},
		{
			name: "Error: propagates service errors",
			args: args{ctx: ctx, featureKey: services.FeatureOpenHoundSupport},
			setupMocks: func(t *testing.T, m *mock) {
				m.service.EXPECT().IsEnabled(ctx, services.FeatureOpenHoundSupport).Return(false, serviceErr)
			},
			want: want{err: serviceErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var (
				m = &mock{
					service: handlersmocks.NewMockFeatureFlag(t),
				}
				handler = handlers.NewHandlersContainer(m.service)
			)
			tt.setupMocks(t, m)
			enabled, err := handler.IsEnabled(tt.args.ctx, tt.args.featureKey)
			if tt.want.err != nil {
				assert.ErrorIs(t, err, tt.want.err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want.enabled, enabled)
		})
	}
}
