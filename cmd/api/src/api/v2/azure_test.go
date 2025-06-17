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
	"github.com/specterops/bloodhound/analysis/azure"

	"github.com/specterops/dawgs/graph"
	graphmocks "github.com/specterops/dawgs/graph/mocks"
	"github.com/specterops/dawgs/ops"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_GetAZRelatedEntities(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDB *graphmocks.MockDatabase
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
						Path:     "/api/v2/azure/invalid",
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
						Path:     "/api/v2/azure/type",
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
						Path:     "/api/v2/azure/type",
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
						Path:     "/api/v2/azure/type",
						RawQuery: "object_id=id&type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(v2.ErrParameterSkip)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"error fetching related entity type inbound-control: invalid skip parameter"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: graphRelatedEntityType - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
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
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(v2.ErrParameterSkip)
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
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(v2.ErrParameterRelatedEntityType)
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
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(ops.ErrGraphQueryMemoryLimit)
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
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&related_entity_type=descendent-users&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an unknown error occurred during the request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: listRelatedEntityType - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&related_entity_type=inbound-control&skip=0&limit=1",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mocks *mock) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":0,"limit":1,"skip":0,"data":[]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDB: graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				Graph: mocks.mockDB,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/azure/{entity_type}", resources.GetAZEntity).Methods(request.Method)
			router.ServeHTTP(response, request)

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
		mockDB *graphmocks.MockDatabase
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.BaseDetails(azure.BaseDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, OutboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.UserDetails(azure.UserDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, GroupMembership: 0, Roles: 0, ExecutionPrivileges: 0, OutboundObjectControl: 0, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.GroupDetails(azure.GroupDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Roles: 0, GroupMembers: 0, GroupMembership: 0, OutboundObjectControl: 0, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.TenantDetails(azure.TenantDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ManagementGroupDetails(azure.ManagementGroupDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.SubscriptionDetails(azure.SubscriptionDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ResourceGroupDetails(azure.ResourceGroupDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Descendents: azure.Descendents{DescendentCounts: map[string]int(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.VMDetails(azure.VMDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundExecutionPrivileges: 0, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ManagedClusterDetails(azure.ManagedClusterDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ContainerRegistryDetails(azure.ContainerRegistryDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.WebAppDetails(azure.WebAppDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.LogicAppDetails(azure.LogicAppDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.AutomationAccountDetails(azure.AutomationAccountDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.KeyVaultDetails(azure.KeyVaultDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Readers: azure.KeyVaultReaderCounts{KeyReaders: 0, CertificateReaders: 0, SecretReaders: 0, AllReaders: 0}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.DeviceDetails(azure.DeviceDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundExecutionPrivileges: 0, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ApplicationDetails(azure.ApplicationDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.VMScaleSetDetails(azure.VMScaleSetDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.ServicePrincipalDetails(azure.ServicePrincipalDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, Roles: 0, InboundObjectControl: 0, OutboundObjectControl: 0, InboundAbusableAppRoleAssignments: 0, OutboundAbusableAppRoleAssignments: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.RoleDetails(azure.RoleDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, ActiveAssignments: 0, PIMAssignments: 0}),
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
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
				mocks.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				res: azure.FunctionAppDetails(azure.FunctionAppDetails{Node: azure.Node{Kind: "", Properties: map[string]interface{}(nil)}, InboundObjectControl: 0}),
				err: nil,
			},
		},
		{
			name: "Error: unknown azure entity",
			args: args{
				entityType: "unknown",
			},
			setupMocks: func(t *testing.T, mocks *mock) {},
			want: want{
				res: nil,
				err: errors.New("unknown azure entity unknown"),
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDB: graphmocks.NewMockDatabase(ctrl),
			}

			testCase.setupMocks(t, mocks)

			res, err := v2.GetAZEntityInformation(context.Background(), mocks.mockDB, testCase.args.entityType, "id", false)

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
		mockDB *graphmocks.MockDatabase
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
						Path:     "/api/v2/azure/{entity_type}",
						RawQuery: "object_id=id&related_entity_type=bad",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
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
				mock.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(graph.ErrNoResultsFound)
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
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"db error: unknown azure entity unknown"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
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
				mock.mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{"isOwnedObject":false, "isTierZero":false, "kind":"","props":null,"active_assignments":0,"approvers":0, "pim_assignments":0}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDB: graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				Graph: mocks.mockDB,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/azure/{entity_type}", resources.GetAZEntity).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
