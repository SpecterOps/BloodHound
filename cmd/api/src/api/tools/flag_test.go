// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/tools"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestToolContainer_ToggleFlag(t *testing.T) {
	t.Parallel()

	type mock struct {
		database *mocks.MockDatabase
	}

	type testCase struct {
		name       string
		request    func() *http.Request
		setupMocks func(t *testing.T, mock *mock)
		assert     func(t *testing.T, response *httptest.ResponseRecorder)
	}

	testCases := []testCase{
		{
			name: "Error: malformed feature id returns bad request",
			request: func() *http.Request {
				return newToggleFlagRequestWithFeatureID("bad")
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusBadRequest, response.Code)
			},
		},
		{
			name: "Error: get flag failure returns database error response",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(appcfg.FeatureFlag{}, errors.New("get flag failed"))
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusInternalServerError, response.Code)
			},
		},
		{
			name: "Error: set flag failure returns database error response",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				featureFlag := appcfg.FeatureFlag{
					Key:     appcfg.FeatureFindingsPrioritizationV0,
					Enabled: false,
				}

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(featureFlag, nil)
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).Return(errors.New("set flag failed"))
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusInternalServerError, response.Code)
			},
		},
		{
			name: "Error: analysis request failure returns database error response and rollsback ff",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				featureFlag := appcfg.FeatureFlag{
					Key:     appcfg.FeatureFindingsPrioritizationV0,
					Enabled: false,
				}

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(featureFlag, nil)
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.True(t, updatedFeatureFlag.Enabled, "expected persisted feature flag to be enabled")
						return nil
					},
				)
				mock.database.EXPECT().RequestAnalysis(gomock.Any(), appcfg.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(errors.New("request analysis failed"))
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.False(t, updatedFeatureFlag.Enabled, "expected persisted feature flag to be rolled back after analysis request failure")
						return nil
					},
				)
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusInternalServerError, response.Code)
			},
		},
		{
			name: "Error: rollback failure returns database error response",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				featureFlag := appcfg.FeatureFlag{
					Key:     appcfg.FeatureFindingsPrioritizationV0,
					Enabled: false,
				}

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(featureFlag, nil)
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.True(t, updatedFeatureFlag.Enabled, "expected persisted feature flag to be enabled")
						return nil
					},
				)
				mock.database.EXPECT().RequestAnalysis(gomock.Any(), appcfg.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(errors.New("request analysis failed"))
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.False(t, updatedFeatureFlag.Enabled, "expected rollback to restore the original disabled state")
						return errors.New("rollback failed")
					},
				)
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, http.StatusInternalServerError, response.Code)
			},
		},
		{
			name: "Success: enabling findings prioritization requests analysis and returns enabled response",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				featureFlag := appcfg.FeatureFlag{
					Key:     appcfg.FeatureFindingsPrioritizationV0,
					Enabled: false,
				}

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(featureFlag, nil)
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.True(t, updatedFeatureFlag.Enabled, "expected persisted feature flag to be enabled")
						return nil
					},
				)
				mock.database.EXPECT().RequestAnalysis(gomock.Any(), appcfg.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(nil)
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assertToggleFlagResponseEnabled(t, response, http.StatusOK, true)
			},
		},
		{
			name: "Success: enabling non-user-updatable findings prioritization still requests analysis and returns enabled response",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				featureFlag := appcfg.FeatureFlag{
					Key:           appcfg.FeatureFindingsPrioritizationV0,
					Enabled:       false,
					UserUpdatable: false,
				}

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(featureFlag, nil)
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.True(t, updatedFeatureFlag.Enabled, "expected persisted feature flag to be enabled")
						require.False(t, updatedFeatureFlag.UserUpdatable, "expected tools path to preserve non-user-updatable flags")
						return nil
					},
				)
				mock.database.EXPECT().RequestAnalysis(gomock.Any(), appcfg.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing).Return(nil)
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assertToggleFlagResponseEnabled(t, response, http.StatusOK, true)
			},
		},
		{
			name: "Success: disabling findings prioritization does not request analysis and returns disabled response",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				featureFlag := appcfg.FeatureFlag{
					Key:     appcfg.FeatureFindingsPrioritizationV0,
					Enabled: true,
				}

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(featureFlag, nil)
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.False(t, updatedFeatureFlag.Enabled, "expected persisted feature flag to be disabled")
						return nil
					},
				)
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assertToggleFlagResponseEnabled(t, response, http.StatusOK, false)
			},
		},
		{
			name: "Success: enabling non-prioritization flag does not request analysis and returns enabled response",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				featureFlag := appcfg.FeatureFlag{
					Key:     "some_other_flag",
					Enabled: false,
				}

				mock.database.EXPECT().GetFlag(gomock.Any(), int32(1)).Return(featureFlag, nil)
				mock.database.EXPECT().SetFlag(gomock.Any(), gomock.AssignableToTypeOf(appcfg.FeatureFlag{})).DoAndReturn(
					func(_ any, updatedFeatureFlag appcfg.FeatureFlag) error {
						require.True(t, updatedFeatureFlag.Enabled, "expected persisted feature flag to be enabled")
						return nil
					},
				)
			},
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				t.Helper()
				assertToggleFlagResponseEnabled(t, response, http.StatusOK, true)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mock := &mock{
				database: mocks.NewMockDatabase(mockController),
			}

			testCase.setupMocks(t, mock)

			toolContainer := tools.NewToolContainer(mock.database)
			request := newToggleFlagRequestWithFeatureID("1")
			if testCase.request != nil {
				request = testCase.request()
			}
			response := httptest.NewRecorder()

			toolContainer.ToggleFlag(response, request)
			testCase.assert(t, response)
		})
	}
}

func assertToggleFlagResponseEnabled(t *testing.T, response *httptest.ResponseRecorder, expectedStatusCode int, expectedEnabled bool) {
	t.Helper()

	assert.Equal(t, expectedStatusCode, response.Code)

	var apiResponse api.BasicResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &apiResponse))

	var actualResponse tools.ToggleFlagResponse
	require.NoError(t, json.Unmarshal(apiResponse.Data, &actualResponse))
	assert.Equal(t, expectedEnabled, actualResponse.Enabled)
}

func newToggleFlagRequestWithFeatureID(featureID string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/flags/1/toggle", nil)
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(tools.URIPathVariableFeatureID, featureID)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}
