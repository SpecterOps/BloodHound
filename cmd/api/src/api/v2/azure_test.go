// Copyright 2025 Specter Ops, Inc.
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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	mocks_db "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	mocks_graph "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	azure_schema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"

	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/util/size"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_GetAZRelatedEntities(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase   *mocks_db.MockDatabase
		mockGraphDB    *graphmocks.MockDatabase
		mockGraphQuery *mocks_graph.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: missing query parameter object ID - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id&related_entity_type=",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"query parameter object_id is required"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: invalid type - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "type=bad&object_id=id&related_entity_type=list",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"query parameter \"type\" is malformed: invalid return type requested for related entities"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: malformed query parameter skip - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=list&skip=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"query parameter \"skip\" is malformed: error converting skip value true to int: strconv.Atoi: parsing \"true\": invalid syntax"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: malformed query parameter limit - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=list&skip=1&limit=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"query parameter \"limit\" is malformed: error converting limit value true to int: strconv.Atoi: parsing \"true\": invalid syntax"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: graphRelatedEntityType database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(v2.ErrParameterSkip)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error fetching related entity type inbound-control: invalid skip parameter"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetPrimaryDisplayKindsError",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(nil, errors.New("database error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: graphRelatedEntityType - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid skip parameter - 400",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(v2.ErrParameterSkip)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"invalid skip: 0"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: related entity type not found - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(v2.ErrParameterRelatedEntityType)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"no matching related entity list type for descendent-users"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: graph query memory limit - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(ops.ErrGraphQueryMemoryLimit)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"calculating the request results exceeded memory limitations due to the volume of objects involved"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: ReadTransaction database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an unknown error occurred during the request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetPrimaryDisplayKindsError",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=inbound-control&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(nil, errors.New("database error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: listRelatedEntityType - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=inbound-control&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":0,"limit":1,"skip":0,"data":[]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=inbound-control&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":0,"limit":1,"skip":0,"data":[]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=inbound-control&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				props := graph.AsProperties(map[string]any{
					"tenantid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", azure_schema.Role).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{azure_schema.Entity, azure_schema.Role},
					Properties: props,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":0,"limit":1,"skip":0,"data":[]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=inbound-control&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", azure_schema.Role).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{azure_schema.Entity, azure_schema.Role},
					Properties: graph.AsProperties(map[string]any{
						"tenantid": "12345",
					}),
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
				mockGraphDB:    graphmocks.NewMockDatabase(ctrl),
				mockGraphQuery: mocks_graph.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			bheCtx := bhctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				Graph:      mocks.mockGraphDB,
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/azure/{entity_type}", resources.GetAZEntity).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_GetAZEntityInformation(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase   *mocks_db.MockDatabase
		mockGraphDB    *graphmocks.MockDatabase
		mockGraphQuery *mocks_graph.MockGraph
	}
	type args struct {
		entityType string
	}
	type want struct {
		res any
		err error
	}
	type testData struct {
		name       string
		args       args
		setupMocks func(t *testing.T, mock *mock)
		want       want
	}

	tt := []testData{
		{
			name: "Error: entityTypeBase",
			args: args{
				entityType: "az-base",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeBase",
			args: args{
				entityType: "az-base",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.BaseDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, OutboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeUsers",
			args: args{
				entityType: "users",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeUsers",
			args: args{
				entityType: "users",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.UserDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, GroupMembership: 0, Roles: 0, ExecutionPrivileges: 0, OutboundObjectControl: 0, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeGroups",
			args: args{
				entityType: "groups",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeGroups",
			args: args{
				entityType: "groups",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.GroupDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Roles: 0, GroupMembers: 0, GroupMembership: 0, OutboundObjectControl: 0, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeTenants",
			args: args{
				entityType: "tenants",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeTenants",
			args: args{
				entityType: "tenants",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.TenantDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeManagementGroups",
			args: args{
				entityType: "management-groups",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeManagementGroups",
			args: args{
				entityType: "management-groups",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ManagementGroupDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeSubscriptions",
			args: args{
				entityType: "subscriptions",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeSubscriptions",
			args: args{
				entityType: "subscriptions",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.SubscriptionDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeResourceGroups",
			args: args{
				entityType: "resource-groups",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeResourceGroups",
			args: args{
				entityType: "resource-groups",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ResourceGroupDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeVMs",
			args: args{
				entityType: "vms",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeVMs",
			args: args{
				entityType: "vms",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.VMDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundExecutionPrivileges: 0, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeManagedClusters",
			args: args{
				entityType: "managed-clusters",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeManagedClusters",
			args: args{
				entityType: "managed-clusters",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ManagedClusterDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeContainerRegistries",
			args: args{
				entityType: "container-registries",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeContainerRegistries",
			args: args{
				entityType: "container-registries",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ContainerRegistryDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeWebApps",
			args: args{
				entityType: "web-apps",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeWebApps",
			args: args{
				entityType: "web-apps",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.WebAppDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeLogicApps",
			args: args{
				entityType: "logic-apps",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeLogicApps",
			args: args{
				entityType: "logic-apps",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.LogicAppDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeAutomationAccounts",
			args: args{
				entityType: "automation-accounts",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeAutomationAccounts",
			args: args{
				entityType: "automation-accounts",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.AutomationAccountDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeKeyVaults",
			args: args{
				entityType: "key-vaults",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeKeyVaults",
			args: args{
				entityType: "key-vaults",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.KeyVaultDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Readers: azure.KeyVaultReaderCounts{KeyReaders: 0, CertificateReaders: 0, SecretReaders: 0, AllReaders: 0}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeDevices",
			args: args{
				entityType: "devices",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeDevices",
			args: args{
				entityType: "devices",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.DeviceDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundExecutionPrivileges: 0, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeApplications",
			args: args{
				entityType: "applications",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeApplications",
			args: args{
				entityType: "applications",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ApplicationDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeVMScaleSets",
			args: args{
				entityType: "vm-scale-sets",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeVMScaleSets",
			args: args{
				entityType: "vm-scale-sets",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.VMScaleSetDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeServicePrincipals",
			args: args{
				entityType: "service-principals",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeServicePrincipals",
			args: args{
				entityType: "service-principals",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ServicePrincipalDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Roles: 0, InboundObjectControl: 0, OutboundObjectControl: 0, InboundAbusableAppRoleAssignments: 0, OutboundAbusableAppRoleAssignments: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeRoles",
			args: args{
				entityType: "roles",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeRoles",
			args: args{
				entityType: "roles",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.RoleDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, ActiveAssignments: 0, PIMAssignments: 0},
				err: nil,
			},
		},
		{
			name: "Error: entityTypeFunctionApps",
			args: args{
				entityType: "function-apps",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			want: want{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			name: "Success: entityTypeFunctionApps",
			args: args{
				entityType: "function-apps",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mocks.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.FunctionAppDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0},
				err: nil,
			},
		},
		{
			name: "Error: unknown azure entity",
			args: args{
				entityType: "unknown",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
			want: want{
				res: nil,
				err: errors.New("unknown azure entity unknown"),
			},
		},
		{
			name: "Error: GetPrimaryDisplayKindsError",
			args: args{
				entityType: "base",
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(nil, errors.New("database error"))
			},
			want: want{
				res: nil,
				err: errors.New("error fetching primary display kinds: database error"),
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
				mockGraphDB:    graphmocks.NewMockDatabase(ctrl),
				mockGraphQuery: mocks_graph.NewMockGraph(ctrl),
			}

			testCase.setupMocks(t, mocks)

			res, err := v2.GetAZEntityInformation(context.Background(), mocks.mockDatabase, mocks.mockGraphDB, testCase.args.entityType, "id", false)

			if err != nil && testCase.want.err != nil {
				require.Equal(t, testCase.want.err, err)
			} else {
				require.Equal(t, testCase.want.res, res)
			}
		})
	}
}

func TestManagementResource_GetAZEntity(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase   *mocks_db.MockDatabase
		mockGraphDB    *graphmocks.MockDatabase
		mockGraphQuery *mocks_graph.MockGraph
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: missing parameter object ID - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=&related_entity_type=bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"query parameter object_id is required"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: related entity type not found - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&related_entity_type=bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"no matching related entity list type for bad"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid query parameter counts - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&counts=bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: database error no results found - Not Found",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   `{"errors":[{"context":"","message":"not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetAZEntityInformation unknown azure entity - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/unknown",
						RawQuery: "object_id=id&counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetPrimaryDisplayKindsError",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(nil, errors.New("database error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"db error: error fetching primary display kinds: database error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: GetAZEntity - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"isOwnedObject":false, "isTierZero":false, "kind":"","props":null,"active_assignments":0,"approvers":0, "kinds":null, "pim_assignments":0}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"isOwnedObject":false, "isTierZero":false, "kind":"","props":null,"active_assignments":0,"approvers":0, "kinds":null, "pim_assignments":0}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)

				props := graph.AsProperties(map[string]any{
					"tenantid": "12345",
				})
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", azure_schema.Role).Return(&graph.Node{
					ID:         graph.ID(16),
					Kinds:      graph.Kinds{azure_schema.Entity, azure_schema.Role},
					Properties: props,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"isOwnedObject":false, "isTierZero":false, "kind":"","props":null,"active_assignments":0,"approvers":0, "kinds":null, "pim_assignments":0}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/roles",
						RawQuery: "object_id=id&counts=true",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", azure_schema.Role).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{azure_schema.Entity, azure_schema.Role},
					Properties: graph.AsProperties(map[string]any{
						"tenantid": "12345",
					}),
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase:   mocks_db.NewMockDatabase(ctrl),
				mockGraphDB:    graphmocks.NewMockDatabase(ctrl),
				mockGraphQuery: mocks_graph.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			bheCtx := bhctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				Graph:      mocks.mockGraphDB,
				GraphQuery: mocks.mockGraphQuery,
				DB:         mocks.mockDatabase,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/azure/{entity_type}", resources.GetAZEntity).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

const (
	etacAnchorObjectID  = "sp-tenant-a"
	etacAllowedTenantID = "tenant-a-id"
	etacAllowedObjectID = "user-tenant-a-allowed"
	etacAllowedName     = "ALLOWED USER TENANT A"
	etacForeignTenantID = "tenant-b-id"
	etacForeignObjectID = "user-tenant-b"
	etacForeignName     = "USER TENANT B"
)

// TestResources_GetAZEntity_ETACFiltersRelatedEntities verifies that GetAZEntity filters related entities from inaccessible tenants
// even when the anchor node belongs to an allowed tenant.
// This and the following ETAC test functions use related_entity_type=outbound-control to cover the ETAC logic shared by all related entity types.
func TestResources_GetAZEntity_ETACFiltersRelatedEntities(t *testing.T) {
	t.Parallel()

	type testData struct {
		name           string
		rawQuery       string
		isListResponse bool
	}

	tt := []testData{
		{
			name:           "type=graph must not expose foreign tenant nodes",
			rawQuery:       "object_id=" + etacAnchorObjectID + "&related_entity_type=outbound-control&type=graph",
			isListResponse: false,
		},
		{
			name:           "type=list must not expose foreign tenant nodes",
			rawQuery:       "object_id=" + etacAnchorObjectID + "&related_entity_type=outbound-control&type=list&skip=0&limit=100",
			isListResponse: true,
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl           = gomock.NewController(t)
				mockDatabase   = mocks_db.NewMockDatabase(ctrl)
				mockGraphDB    = graphmocks.NewMockDatabase(ctrl)
				mockGraphQuery = mocks_graph.NewMockGraph(ctrl)
				requestWithCtx = etacRestrictedRequest(testCase.rawQuery)
				response       = httptest.NewRecorder()
				router         = mux.NewRouter()
			)

			mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), etacAnchorObjectID, azure_schema.ServicePrincipal).Return(graph.NewNode(graph.ID(1), graph.AsProperties(graph.PropertyMap{
				azure_schema.TenantID: etacAllowedTenantID,
			}), azure_schema.Entity, azure_schema.ServicePrincipal), nil)

			setupAZMixedTenantOutboundControlTraversal(t, ctrl, mockGraphDB)

			resources := v2.Resources{
				Graph:      mockGraphDB,
				GraphQuery: mockGraphQuery,
				DB:         mockDatabase,
				DogTags:    etacEnabledDogTags(),
			}

			router.HandleFunc("/api/v2/azure/{entity_type}", resources.GetAZEntity).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, _, body := test.ProcessResponse(t, response)

			require.Equal(t, http.StatusOK, status)
			assertNoForeignTenantData(t, body)
			assertAllowedTenantDataPresent(t, body)

			if !testCase.isListResponse {
				assert.Contains(t, body, "** Hidden Object **", "graph responses redact inaccessible nodes rather than dropping them")
			} else {
				assert.NotContains(t, body, "** Hidden Object **", "list responses drop inaccessible nodes rather than redacting them")
				assert.Contains(t, body, `"count":1`, "count must reflect the filtered/allowed nodes, not the unfiltered traversal")

			}
		})
	}
}

// TestResources_GetAZRelatedEntities_ETACFiltersForeignTenants verifies that GetAZRelatedEntities filters related entities
// from inaccessible tenants.
func TestResources_GetAZRelatedEntities_ETACFiltersForeignTenants(t *testing.T) {
	t.Parallel()

	type testData struct {
		name           string
		rawQuery       string
		isListResponse bool
	}

	tt := []testData{
		{
			name:           "type=graph must not expose foreign tenant nodes",
			rawQuery:       "object_id=" + etacAnchorObjectID + "&related_entity_type=outbound-control&type=graph",
			isListResponse: false,
		},
		{
			name:           "type=list must not expose foreign tenant nodes",
			rawQuery:       "object_id=" + etacAnchorObjectID + "&related_entity_type=outbound-control&type=list&skip=0&limit=100",
			isListResponse: true,
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl           = gomock.NewController(t)
				mockDatabase   = mocks_db.NewMockDatabase(ctrl)
				mockGraphDB    = graphmocks.NewMockDatabase(ctrl)
				requestWithCtx = etacRestrictedRequest(testCase.rawQuery)
				response       = httptest.NewRecorder()
			)

			mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())

			setupAZMixedTenantOutboundControlTraversal(t, ctrl, mockGraphDB)

			resources := v2.Resources{
				Graph:   mockGraphDB,
				DB:      mockDatabase,
				DogTags: etacEnabledDogTags(),
			}

			// we need to extract the user's allowlist
			user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(requestWithCtx).AuthCtx)
			require.True(t, isUser)
			allowList := v2.ExtractEnvironmentIDsFromUser(&user)

			resources.GetAZRelatedEntities(requestWithCtx.Context(), response, requestWithCtx, etacAnchorObjectID, azure_schema.ServicePrincipal, allowList)

			status, _, body := test.ProcessResponse(t, response)

			require.Equal(t, http.StatusOK, status)
			assertNoForeignTenantData(t, body)
			assertAllowedTenantDataPresent(t, body)

			if !testCase.isListResponse {
				assert.Contains(t, body, "** Hidden Object **", "graph responses redact inaccessible nodes rather than dropping them")
			} else {
				assert.NotContains(t, body, "** Hidden Object **", "list responses drop inaccessible nodes rather than redacting them")
				assert.Contains(t, body, `"count":1`, "count must reflect the filtered/allowed nodes, not the unfiltered traversal")

			}
		})
	}
}

// TestResources_GetAZEntity_ETACPermissions verifies that users with a restricted environment allowlist are subject to filtering.
// Responses are unfiltered when the feature flag is off or the user has AllEnvironments.
func TestResources_GetAZEntity_ETACPermissions(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name                string
		user                model.User
		etacEnabled         bool
		expectForeignTenant bool
	}{
		{
			name:        "ETAC enabled with restricted environments allowlist hides foreign-tenant nodes",
			user:        etacRestrictedUser(),
			etacEnabled: true,
		},
		{
			name: "ETAC disabled with empty environment allowlist returns nodes from both tenants",
			user: model.User{
				AllEnvironments:                  false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{},
			},
			etacEnabled:         false,
			expectForeignTenant: true,
		},
		{
			name: "ETAC enabled for AllEnvironments user with empty environment allowlist returns nodes from both tenants",
			user: model.User{
				AllEnvironments:                  true,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{},
			},
			etacEnabled:         true,
			expectForeignTenant: true,
		},
	}

	for _, testCase := range testCases {
		for _, responseType := range []string{"graph", "list"} {
			t.Run(testCase.name+"/type="+responseType, func(t *testing.T) {
				t.Parallel()

				var (
					ctrl           = gomock.NewController(t)
					mockDatabase   = mocks_db.NewMockDatabase(ctrl)
					mockGraphDB    = graphmocks.NewMockDatabase(ctrl)
					mockGraphQuery = mocks_graph.NewMockGraph(ctrl)
					request        = httptest.NewRequest(http.MethodGet, "/api/v2/azure/service-principals?object_id="+etacAnchorObjectID+"&related_entity_type=outbound-control&type="+responseType+"&skip=0&limit=100", nil)
					bheCtx         = bhctx.Context{
						AuthCtx: auth.Context{Owner: testCase.user},
					}
					requestWithCtx = request.WithContext(bheCtx.ConstructGoContext())
					response       = httptest.NewRecorder()
					router         = mux.NewRouter()
					resources      = v2.Resources{
						Graph:      mockGraphDB,
						GraphQuery: mockGraphQuery,
						DB:         mockDatabase,
						DogTags: dogtags.NewTestService(dogtags.TestOverrides{
							Bools: map[dogtags.BoolDogTag]bool{
								dogtags.ETAC_ENABLED: testCase.etacEnabled,
							},
						}),
					}
				)

				mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				if testCase.etacEnabled && !testCase.user.AllEnvironments {
					mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), etacAnchorObjectID, azure_schema.ServicePrincipal).Return(graph.NewNode(graph.ID(1), graph.AsProperties(graph.PropertyMap{
						azure_schema.TenantID: etacAllowedTenantID,
					}), azure_schema.Entity, azure_schema.ServicePrincipal), nil)
				}

				setupAZMixedTenantOutboundControlTraversal(t, ctrl, mockGraphDB)

				router.HandleFunc("/api/v2/azure/{entity_type}", resources.GetAZEntity).Methods(requestWithCtx.Method)
				router.ServeHTTP(response, requestWithCtx)

				status, _, body := test.ProcessResponse(t, response)

				require.Equal(t, http.StatusOK, status)
				assertAllowedTenantDataPresent(t, body)
				if testCase.expectForeignTenant {
					assert.Contains(t, body, etacForeignObjectID, "object ID should remain visible -- ETAC filtering does not apply to this user")
					assert.Contains(t, body, etacForeignName, "node name should remain visible -- ETAC filtering does not apply to this user")

					if responseType == "graph" {
						assert.NotContains(t, body, "** Hidden Object **", "no hidden placeholders -- ETAC filtering does not apply to this user")
					}

					if responseType == "list" {
						assert.Contains(t, body, `"count":2`, "count must reflect the entire traversal -- ETAC filtering does not apply to this user")
					}
				} else {
					assertNoForeignTenantData(t, body)

					if responseType == "graph" {
						assert.Contains(t, body, "** Hidden Object **", "graph responses redact inaccessible nodes rather than dropping them")
					} else {
						assert.NotContains(t, body, "** Hidden Object **", "list responses drop inaccessible nodes rather than redacting them")
					}

					if responseType == "list" {
						assert.Contains(t, body, `"count":1`, "count must reflect the filtered/allowed nodes, not the unfiltered traversal")
					}
				}
			})
		}
	}
}

// etacRestrictedUser is a user whose ETAC allowlist contains only the anchor node's tenant.
func etacRestrictedUser() model.User {
	return model.User{
		AllEnvironments: false,
		EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
			{EnvironmentID: etacAllowedTenantID},
		},
	}
}

// setupAZMixedTenantOutboundControlTraversal mocks an outbound-control traversal in which the anchor Service Principal
// controls one node in the user's allowed tenant and one node in a foreign tenant.
func setupAZMixedTenantOutboundControlTraversal(t *testing.T, ctrl *gomock.Controller, mockGraphDB *graphmocks.MockDatabase) {
	t.Helper()

	var (
		mockTx        = graphmocks.NewMockTransaction(ctrl)
		mockNodeQuery = graphmocks.NewMockNodeQuery(ctrl)
		mockRelQuery  = graphmocks.NewMockRelationshipQuery(ctrl)

		anchorNode = graph.NewNode(graph.ID(1), graph.AsProperties(graph.PropertyMap{
			common.ObjectID:       etacAnchorObjectID,
			common.Name:           "SP TENANT A",
			azure_schema.TenantID: etacAllowedTenantID,
		}), azure_schema.Entity, azure_schema.ServicePrincipal)

		foreignNode = graph.NewNode(graph.ID(2), graph.AsProperties(graph.PropertyMap{
			common.ObjectID:       etacForeignObjectID,
			common.Name:           etacForeignName,
			azure_schema.TenantID: etacForeignTenantID,
		}), azure_schema.Entity, azure_schema.User)

		allowedNode = graph.NewNode(graph.ID(3), graph.AsProperties(graph.PropertyMap{
			common.ObjectID:       etacAllowedObjectID,
			common.Name:           etacAllowedName,
			azure_schema.TenantID: etacAllowedTenantID,
		}), azure_schema.Entity, azure_schema.User)

		foreignEdge = graph.NewRelationship(graph.ID(100), anchorNode.ID, foreignNode.ID, graph.NewProperties(), azure_schema.Owns)
		allowedEdge = graph.NewRelationship(graph.ID(101), anchorNode.ID, allowedNode.ID, graph.NewProperties(), azure_schema.Owns)
	)

	// directionalCursor mocks a graph.Cursor that yields the given results and closes
	directionalCursor := func(results ...graph.DirectionalResult) graph.Cursor[graph.DirectionalResult] {
		cursor := graphmocks.NewMockCursor[graph.DirectionalResult](ctrl)
		channel := make(chan graph.DirectionalResult, len(results))

		for _, result := range results {
			channel <- result
		}
		close(channel)

		cursor.EXPECT().Chan().Return(channel).AnyTimes()
		cursor.EXPECT().Error().Return(nil).AnyTimes()

		return cursor
	}

	mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, delegate func(tx graph.Transaction) error, _ ...graph.TransactionOption) error {
		return delegate(mockTx)
	})

	mockTx.EXPECT().GraphQueryMemoryLimit().Return(size.Size(0)).AnyTimes()
	mockTx.EXPECT().Nodes().Return(mockNodeQuery).AnyTimes()
	mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockNodeQuery).AnyTimes()
	mockNodeQuery.EXPECT().First().Return(anchorNode, nil)
	mockTx.EXPECT().Relationships().Return(mockRelQuery).AnyTimes()
	mockRelQuery.EXPECT().Filterf(gomock.Any()).Return(mockRelQuery).AnyTimes()

	// the anchor node traverses to the allowed and foreign nodes, which then terminate their paths
	mockRelQuery.EXPECT().FetchDirection(gomock.Any(), gomock.Any()).DoAndReturn(func(_ graph.Direction, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
		return delegate(directionalCursor(
			graph.NewDirectionalResult(graph.DirectionOutbound, allowedEdge, allowedNode),
			graph.NewDirectionalResult(graph.DirectionOutbound, foreignEdge, foreignNode),
		))
	})
	mockRelQuery.EXPECT().FetchDirection(gomock.Any(), gomock.Any()).DoAndReturn(func(_ graph.Direction, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
		return delegate(directionalCursor())
	}).AnyTimes()
}

// etacRestrictedRequest builds a request with the auth context of an ETAC-restricted user.
func etacRestrictedRequest(rawQuery string) *http.Request {
	request := &http.Request{
		URL: &url.URL{
			Path:     "/api/v2/azure/service-principals",
			RawQuery: rawQuery,
		},
		Method: http.MethodGet,
	}

	bheCtx := bhctx.Context{
		AuthCtx: auth.Context{
			PermissionOverrides: auth.PermissionOverrides{},
			Owner:               etacRestrictedUser(),
			Session:             model.UserSession{},
		},
	}

	return request.WithContext(bheCtx.ConstructGoContext())
}

func etacEnabledDogTags() dogtags.Service {
	return dogtags.NewTestService(dogtags.TestOverrides{
		Bools: map[dogtags.BoolDogTag]bool{
			dogtags.ETAC_ENABLED: true,
		},
	})
}

// assertNoForeignTenantData checks that the response omits the foreign tenantid, objectid, and name
func assertNoForeignTenantData(t *testing.T, body string) {
	t.Helper()

	assert.NotContains(t, body, etacForeignTenantID, "foreign tenantid must not be present in the response")
	assert.NotContains(t, body, etacForeignObjectID, "foreign tenant objectid must not be present in the response")
	assert.NotContains(t, body, etacForeignName, "foreign tenant node name must not be present in the response")
}

// assertAllowedTenantDataPresent checks that the response still includes the related node from the allowed tenant
func assertAllowedTenantDataPresent(t *testing.T, body string) {
	t.Helper()

	assert.Contains(t, body, etacAllowedObjectID, "allowed tenant objectid must be present in the response")
	assert.Contains(t, body, etacAllowedName, "allowed tenant node name must be present in the response")
}
