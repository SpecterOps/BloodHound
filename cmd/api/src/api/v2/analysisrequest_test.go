// Copyright 2024 Specter Ops, Inc.
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

package v2_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_GetAnalysisRequest(t *testing.T) {
	const (
		url = "api/v2/analysis/status"
	)

	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	t.Run("success getting analysis", func(t *testing.T) {
		analysisRequest := model.AnalysisRequest{
			RequestedAt: time.Now(),
			RequestedBy: "test",
			RequestType: model.AnalysisRequestType("test-type"),
		}

		mockDB.EXPECT().GetAnalysisRequest(gomock.Any()).Return(analysisRequest, nil)

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL(url).
			OnHandlerFunc(resources.GetAnalysisRequest).
			Require().
			ResponseJSONBody(analysisRequest).
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("error getting analysis", func(t *testing.T) {
		mockDB.EXPECT().GetAnalysisRequest(gomock.Any()).Return(model.AnalysisRequest{}, fmt.Errorf("an error"))

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL(url).
			OnHandlerFunc(resources.GetAnalysisRequest).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}

func TestResources_RequestAnalysis(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *dbMocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
	}

	tt := []testData{
		{
			name: "Error: error finding scheduled analysis parameter - Status Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/analysis",
					},
					Method: http.MethodPut,
				}

				param := map[string]string{
					"object_id": "id",
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner:   model.User{},
						Session: model.UserSession{},
					},
				}

				request = mux.SetURLVars(request, param)
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{}, database.ErrNotFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"resource not found"}],"http_status":404,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetConfigurationParameter database error - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/analysis",
					},
					Method: http.MethodPut,
				}

				param := map[string]string{
					"object_id": "id",
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner:   model.User{},
						Session: model.UserSession{},
					},
				}

				request = mux.SetURLVars(request, param)
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{}, context.DeadlineExceeded)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"request timed out"}],"http_status":500,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: scheduled analysis parameter enabled - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/analysis",
					},
					Method: http.MethodPut,
				}

				param := map[string]string{
					"object_id": "id",
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner:   model.User{},
						Session: model.UserSession{},
					},
				}

				request = mux.SetURLVars(request, param)
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{
					Key: "key",
					Value: types.JSONBObject{
						Object: &appcfg.ScheduledAnalysisParameter{Enabled: true, RRule: "rule"},
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"analysis is configured to run on a schedule, unable to run just in time"}],"http_status":400,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: RequestAnalysis database error - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/analysis",
					},
					Method: http.MethodPut,
				}

				param := map[string]string{
					"object_id": "id",
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner:   model.User{},
						Session: model.UserSession{},
					},
				}

				request = mux.SetURLVars(request, param)
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{
					Key: "key",
					Value: types.JSONBObject{
						Object: &appcfg.ScheduledAnalysisParameter{Enabled: false, RRule: "rule"},
					},
				}, nil)
				mock.mockDatabase.EXPECT().RequestAnalysis(gomock.Any(), "00000000-0000-0000-0000-000000000000").Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: analysis request accepted - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/analysis",
					},
					Method: http.MethodPut,
				}

				param := map[string]string{
					"object_id": "id",
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner:   model.User{},
						Session: model.UserSession{},
					},
				}

				request = mux.SetURLVars(request, param)
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{
					Key: "key",
					Value: types.JSONBObject{
						Object: &appcfg.ScheduledAnalysisParameter{Enabled: false, RRule: "rule"},
					},
				}, nil)
				mock.mockDatabase.EXPECT().RequestAnalysis(gomock.Any(), "00000000-0000-0000-0000-000000000000").Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusAccepted,
				responseHeader: http.Header{},
			},
		},
		{
			name: "Success: user - analysis request accepted - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/analysis",
					},
					Method: http.MethodPut,
				}

				param := map[string]string{
					"object_id": "id",
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner:   model.User{},
						Session: model.UserSession{},
					},
				}

				request = mux.SetURLVars(request, param)
				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{
					Key: "key",
					Value: types.JSONBObject{
						Object: &appcfg.ScheduledAnalysisParameter{Enabled: false, RRule: "rule"},
					},
				}, nil)
				mock.mockDatabase.EXPECT().RequestAnalysis(gomock.Any(), "00000000-0000-0000-0000-000000000000").Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusAccepted,
				responseHeader: http.Header{},
			},
		},
		{
			name: "Success: unknown user - analysis request accepted - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/analysis",
					},
					Method: http.MethodPut,
				}

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.ScheduledAnalysis).Return(appcfg.Parameter{
					Key: "key",
					Value: types.JSONBObject{
						Object: &appcfg.ScheduledAnalysisParameter{Enabled: false, RRule: "rule"},
					},
				}, nil)
				mock.mockDatabase.EXPECT().RequestAnalysis(gomock.Any(), "unknown-user").Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusAccepted,
				responseBody:   ``,
				responseHeader: http.Header{},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: dbMocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB: mocks.mockDatabase,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/analysis", resources.RequestAnalysis).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if body != "" {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}
