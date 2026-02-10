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
package v2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	uuid2 "github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/stretchr/testify/require"

	schemamocks "github.com/specterops/bloodhound/cmd/api/src/api/v2/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_OpenGraphSchemaIngest(t *testing.T) {

	var (
		mockCtrl             = gomock.NewController(t)
		mockOpenGraphService = schemamocks.NewMockOpenGraphSchemaService(mockCtrl)
		userId, err          = uuid2.NewV4()

		graphExtension = v2.GraphExtensionPayload{
			GraphSchemaExtension: v2.GraphSchemaExtensionPayload{
				Name:        "Test_Extension",
				DisplayName: "Test Extension",
				Version:     "1.0.0",
				Namespace:   "TEST",
			},
			/*GraphSchemaProperties: []v2.GraphSchemaPropertiesPayload{
				{
					Name:        "Property_1",
					DisplayName: "Property 1",
					DataType:    "string",
					Description: "a property",
				},
			},*/
			GraphSchemaRelationshipKinds: []v2.GraphSchemaRelationshipKindsPayload{
				{
					Name:          "TEST_GraphSchemaEdgeKind_1",
					Description:   "GraphSchemaRelationshipKind 1",
					IsTraversable: true,
				},
			},
			GraphSchemaNodeKinds: []v2.GraphSchemaNodeKindsPayload{
				{
					Name:          "TEST_GraphSchemaNodeKind_1",
					DisplayName:   "GraphSchemaNodeKind 1",
					Description:   "a graph schema node",
					IsDisplayKind: true,
					Icon:          "User",
					IconColor:     "blue",
				},
			},
			GraphEnvironments: []v2.EnvironmentPayload{
				{
					EnvironmentKind: "TEST_EnvironmentInput",
					SourceKind:      "Source_Kind_1",
					PrincipalKinds:  []string{"User"},
				},
			},
			GraphRelationshipFindings: []v2.RelationshipFindingsPayload{
				{
					Name:             "TEST_Finding_1",
					DisplayName:      "Finding 1",
					SourceKind:       "Source_Kind_1",
					RelationshipKind: "TEST_GraphSchemaEdgeKind_1",
					EnvironmentKind:  "TEST_EnvironmentInput",
					Remediation: v2.RemediationPayload{
						ShortDescription: "remediation for Finding_1",
						LongDescription:  "a remediation for Finding 1",
						ShortRemediation: "do x",
						LongRemediation:  "do x but better",
					},
				},
			},
		}
		serviceGraphExtension = model.GraphExtensionInput{
			ExtensionInput: model.ExtensionInput{
				Name:        "Test_Extension",
				DisplayName: "Test Extension",
				Version:     "1.0.0",
				Namespace:   "TEST",
			},
			PropertiesInput: model.PropertiesInput{},
			/*PropertiesInput: model.PropertiesInput{
				{
					Name:        "Property_1",
					DisplayName: "Property 1",
					DataType:    "string",
					Description: "a property",
				},
			},*/
			RelationshipKindsInput: model.RelationshipsInput{
				{
					Name:          "TEST_GraphSchemaEdgeKind_1",
					Description:   "GraphSchemaRelationshipKind 1",
					IsTraversable: true,
				},
			},
			NodeKindsInput: model.NodesInput{
				{
					Name:          "TEST_GraphSchemaNodeKind_1",
					DisplayName:   "GraphSchemaNodeKind 1",
					Description:   "a graph schema node",
					IsDisplayKind: true,
					Icon:          "User",
					IconColor:     "blue",
				},
			},
			EnvironmentsInput: []model.EnvironmentInput{
				{
					EnvironmentKindName: "TEST_EnvironmentInput",
					SourceKindName:      "Source_Kind_1",
					PrincipalKinds:      []string{"User"},
				},
			},
			RelationshipFindingsInput: []model.RelationshipFindingInput{
				{
					Name:                 "TEST_Finding_1",
					DisplayName:          "Finding 1",
					SourceKindName:       "Source_Kind_1",
					RelationshipKindName: "TEST_GraphSchemaEdgeKind_1",
					EnvironmentKindName:  "TEST_EnvironmentInput",
					RemediationInput: model.RemediationInput{
						ShortDescription: "remediation for Finding_1",
						LongDescription:  "a remediation for Finding 1",
						ShortRemediation: "do x",
						LongRemediation:  "do x but better",
					},
				},
			},
		}
	)
	defer mockCtrl.Finish()
	require.NoError(t, err)

	type fields struct {
		setupOpenGraphServiceMock func(t *testing.T, repository *schemamocks.MockOpenGraphSchemaService)
	}
	type args struct {
		buildRequest func() *http.Request
	}
	type want struct {
		responseCode int
		err          error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "fail - no user",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *schemamocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequest(http.MethodPut, "/api/v2/extensions", nil)
					require.NoError(t, err)
					return req
				},
			},
			want: want{
				responseCode: http.StatusUnauthorized,
				err:          fmt.Errorf("Code: 401 - errors: No associated user found"),
			},
		},
		{
			name: "fail - user is not admin",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *schemamocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPut, "/api/v2/extensions", nil)
					require.NoError(t, err)
					return req
				},
			},
			want: want{
				responseCode: http.StatusForbidden,
				err:          fmt.Errorf("Code: 403 - errors: user does not have sufficient permissions to create or update an extension"),
			},
		},
		{
			name: "fail - open graph extension payload cannot be empty",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *schemamocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut, "/api/v2/extensions", nil)
					require.NoError(t, err)
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: open graph extension payload cannot be empty"),
			},
		},
		{
			name: "fail - invalid content type",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *schemamocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					var body []byte
					body, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut, "/api/v2/extensions", bytes.NewReader(body))
					require.NoError(t, err)
					req.Header.Set("content-type", "invalid")
					return req
				},
			},
			want: want{
				responseCode: http.StatusUnsupportedMediaType,
				err:          fmt.Errorf("Code: 415 - errors: invalid content-type: [invalid]; Content type must be application/json"),
			},
		},
		{
			name: "fail - unable to decode graph schema",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *schemamocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader([]byte("awfawf")))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: unable to decode graph extension payload: invalid character 'a' looking for beginning of value"),
			},
		},
		{
			name: "fail - UpsertOpenGraphExtension - generic error",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, fmt.Errorf("generic error"))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusInternalServerError,
				err:          fmt.Errorf("Code: 500 - errors: an internal error has occurred that is preventing the service from servicing this request"),
			},
		},
		{
			name: "fail - UpsertOpenGraphExtension - validation error",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, fmt.Errorf("%w: some_error", model.ErrGraphExtensionValidation))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: %w: some_error", model.ErrGraphExtensionValidation),
			},
		},
		{
			name: "fail - UpsertOpenGraphExtension - cannot modify a built-in graph extension",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false,
						fmt.Errorf("Error upserting graph extension: %w", model.ErrGraphExtensionBuiltIn))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: Error upserting graph extension: %w", model.ErrGraphExtensionBuiltIn),
			},
		},
		{
			name: "fail - unable to refresh graph db kinds",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, fmt.Errorf("%w: graph_db error", model.ErrGraphDBRefreshKinds))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusInternalServerError,
				err:          fmt.Errorf("Code: 500 - errors: an internal error has occurred that is preventing the service from servicing this request"),
			},
		},
		{
			name: "success - updated graph extension",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(true, nil)
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusOK,
			},
		},
		{
			name: "success - inserted new graph extension",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, nil)
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				response = httptest.NewRecorder()
				request  = tt.args.buildRequest()
			)
			tt.fields.setupOpenGraphServiceMock(t, mockOpenGraphService)

			s := v2.Resources{
				OpenGraphSchemaService: mockOpenGraphService,
			}

			s.OpenGraphSchemaIngest(response, request)
			require.Equal(t, tt.want.responseCode, response.Code) // If success, only a 200 or 201 header response is made
			if tt.want.responseCode != http.StatusOK && tt.want.responseCode != http.StatusCreated {
				// Failure responses
				var errWrapper api.ErrorWrapper
				err := json.Unmarshal(response.Body.Bytes(), &errWrapper)
				require.NoError(t, err)
				require.EqualErrorf(t, errWrapper, tt.want.err.Error(), "Unexpected error: %v", errWrapper.Error())
			}
		})
	}
}

func TestResources_ListExtensions(t *testing.T) {
	t.Parallel()
	type mock struct {
		mockOpenGraphSchemaService *schemamocks.MockOpenGraphSchemaService
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
			name: "Error: error retrieving graph schema extensions",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockOpenGraphSchemaService.EXPECT().ListExtensions(gomock.Any()).Return(model.GraphSchemaExtensions{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"error listing graph schema extensions: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockOpenGraphSchemaService.EXPECT().ListExtensions(gomock.Any()).Return(model.GraphSchemaExtensions{
					{
						Serial: model.Serial{
							ID: 1,
						},
						DisplayName: "Display 1",
						Version:     "v1.0.0",
					},
					{
						Serial: model.Serial{
							ID: 2,
						},
						DisplayName: "Display 2",
						Version:     "v2.0.0",
					},
					{
						Serial: model.Serial{
							ID: 3,
						},
						DisplayName: "Display 3",
						Version:     "v3.0.0",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data": {"extensions":[{"id":"1", "name":"Display 1", "version":"v1.0.0"}, {"id":"2", "name":"Display 2", "version":"v2.0.0"}, {"id":"3", "name":"Display 3", "version":"v3.0.0"}]}}`,
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockOpenGraphSchemaService: schemamocks.NewMockOpenGraphSchemaService(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				OpenGraphSchemaService: mocks.mockOpenGraphSchemaService,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/extensions", resources.ListExtensions).Methods(http.MethodGet)

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_DeleteExtension(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockOpenGraphSchemaService *schemamocks.MockOpenGraphSchemaService
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
			name: "Error: error getting user from context",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions/id",
					},
					Method: http.MethodDelete,
				}

				return request
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusUnauthorized,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"No associated user found"}],"http_status":401,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: user does not have sufficient permissions",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions/id",
					},
					Method: http.MethodDelete,
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner:   model.User{},
						Session: model.UserSession{},
					},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"user does not have sufficient permissions to delete an extension"}],"http_status":403,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: invalid extension id",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions/id",
					},
					Method: http.MethodDelete,
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner: model.User{
							Roles: model.Roles{
								{
									Name: auth.RoleAdministrator,
									Permissions: model.Permissions{
										auth.Permissions().AuthManageSelf,
									},
								},
							},
						},
						Session: model.UserSession{},
					},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"id is malformed"}],"http_status":400,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: extension id not found in database",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions/1",
					},
					Method: http.MethodDelete,
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner: model.User{
							Roles: model.Roles{
								{
									Name: auth.RoleAdministrator,
									Permissions: model.Permissions{
										auth.Permissions().AuthManageSelf,
									},
								},
							},
						},
						Session: model.UserSession{},
					},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockOpenGraphSchemaService.EXPECT().DeleteExtension(gomock.Any(), int32(1)).Return(database.ErrNotFound)

			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message": "no extension found matching extension id: 1"}],"http_status":404,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: failed to delete extension that is built-in",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions/1",
					},
					Method: http.MethodDelete,
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner: model.User{
							Roles: model.Roles{
								{
									Name: auth.RoleAdministrator,
									Permissions: model.Permissions{
										auth.Permissions().AuthManageSelf,
									},
								},
							},
						},
						Session: model.UserSession{},
					},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockOpenGraphSchemaService.EXPECT().DeleteExtension(gomock.Any(), int32(1)).Return(model.ErrGraphExtensionBuiltIn)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message": "built-in extensions cannot be deleted"}],"http_status":400,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Error: failed to delete extension by id",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions/1",
					},
					Method: http.MethodDelete,
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner: model.User{
							Roles: model.Roles{
								{
									Name: auth.RoleAdministrator,
									Permissions: model.Permissions{
										auth.Permissions().AuthManageSelf,
									},
								},
							},
						},
						Session: model.UserSession{},
					},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockOpenGraphSchemaService.EXPECT().DeleteExtension(gomock.Any(), int32(1)).Return(errors.New("error"))

			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message": "an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
			},
		},
		{
			name: "Success",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						Path: "/api/v2/extensions/1",
					},
					Method: http.MethodDelete,
				}

				requestCtx := ctx.Context{
					RequestID: "id",
					AuthCtx: auth.Context{
						Owner: model.User{
							Roles: model.Roles{
								{
									Name: auth.RoleAdministrator,
									Permissions: model.Permissions{
										auth.Permissions().AuthManageSelf,
									},
								},
							},
						},
						Session: model.UserSession{},
					},
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockOpenGraphSchemaService.EXPECT().DeleteExtension(gomock.Any(), int32(1)).Return(nil)

			},
			expected: expected{
				responseCode:   http.StatusNoContent,
				responseHeader: http.Header{},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockOpenGraphSchemaService: schemamocks.NewMockOpenGraphSchemaService(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				OpenGraphSchemaService: mocks.mockOpenGraphSchemaService,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/extensions/{%s}", api.URIPathVariableExtensionID), resources.DeleteExtension).Methods(request.Method)

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if status != http.StatusNoContent {
				assert.JSONEq(t, testCase.expected.responseBody, body)
			} else {
				assert.Equal(t, testCase.expected.responseBody, body)
			}
		})
	}
}
