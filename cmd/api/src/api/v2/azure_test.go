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
	"github.com/specterops/bloodhound/dawgs/graph"
	graphmocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/ops"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_GetAZRelatedEntities(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDB *graphmocks.MockDatabase
	}
	type expected struct {
		responseBody   any
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: empty relatedEntityType - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"query parameter \"related_entity_type\" is malformed: missing required parameter"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: invalid type - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "type=invalid&related_entity_type=list",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"query parameter \"type\" is malformed: invalid return type requested for related entities"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?type=invalid&related_entity_type=list"}},
			},
		},
		{
			name: "Error: malformed query parameter skip - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=list&skip=true",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"query parameter \"skip\" is malformed: error converting skip value true to int: strconv.Atoi: parsing \"true\": invalid syntax"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?related_entity_type=list&skip=true"}},
			},
		},
		{
			name: "Error: malformed query parameter limit - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=list&skip=1&limit=true",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"query parameter \"limit\" is malformed: error converting limit value true to int: strconv.Atoi: parsing \"true\": invalid syntax"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?related_entity_type=list&skip=1&limit=true"}},
			},
		},
		{
			name: "Error: graphRelatedEntityType database error - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(v2.ErrParameterSkip)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"error fetching related entity type inbound-control: invalid skip parameter"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?type=graph&skip=0&limit=1&related_entity_type=inbound-control"}},
			},
		},
		{
			name: "Success: graphRelatedEntityType - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   []byte(`{}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?type=graph&skip=0&limit=1&related_entity_type=inbound-control"}},
			},
		},
		{
			name: "Error: invalid skip parameter - 400",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(v2.ErrParameterSkip)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"invalid skip: 0"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?related_entity_type=descendent-users&skip=0&limit=1"}},
			},
		},
		{
			name: "Error: related entity type not found - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(v2.ErrParameterRelatedEntityType)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   []byte(`{"errors":[{"context":"","message":"no matching related entity list type for descendent-users"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?related_entity_type=descendent-users&skip=0&limit=1"}},
			},
		},
		{
			name: "Error: graph query memory limit - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(ops.ErrGraphQueryMemoryLimit)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"calculating the request results exceeded memory limitations due to the volume of objects involved"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?related_entity_type=descendent-users&skip=0&limit=1"}},
			},
		},
		{
			name: "Error: ReadTransaction database error - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"an unknown error occurred during the request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?related_entity_type=descendent-users&skip=0&limit=1"}},
			},
		},
		{
			name: "Success: listRelatedEntityType - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=inbound-control&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   []byte(`{"count":0,"limit":1,"skip":0,"data":[]}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?related_entity_type=inbound-control&skip=0&limit=1"}},
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
			testCase.emulateWithMocks(t, mocks, request)

			resouces := v2.Resources{
				Graph: mocks.mockDB,
			}

			response := httptest.NewRecorder()

			resouces.GetAZRelatedEntities(context.Background(), response, request, "id")
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			require.Equal(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetAZEntityInformation(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDB         *graphmocks.MockDatabase
		mockGraphQuery *mocks.MockGraph
	}
	type args struct {
		entityType string
	}
	type want struct {
		res any
		err error
	}
	type testData struct {
		name             string
		args             args
		emulateWithMocks func(t *testing.T, mock *mock)
		want             want
	}

	tt := []testData{
		{
			name: "Error: entityTypeBase",
			args: args{
				entityType: "az-base",
			},
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {
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
			emulateWithMocks: func(t *testing.T, mocks *mock) {},
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

			testCase.emulateWithMocks(t, mocks)

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
		responseBody   any
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: missing parameter object ID - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"query parameter object_id is required"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: related entity type not found - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "object_id=id&related_entity_type=bad",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   []byte(`{"errors":[{"context":"","message":"no matching related entity list type for bad"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?object_id=id&related_entity_type=bad"}},
			},
		},
		{
			name: "Error: invalid query parameter counts - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "object_id=id&counts=bad",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?object_id=id&counts=bad"}},
			},
		},
		{
			name: "Error: database error no results found - Not Found",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "object_id=id&counts=true",
					},
				}

				param := map[string]string{
					"entity_type": "roles",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(graph.ErrNoResultsFound)
			},
			expected: expected{
				responseCode:   http.StatusNotFound,
				responseBody:   []byte(`{"errors":[{"context":"","message":"not found"}],"http_status":404,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?object_id=id&counts=true"}},
			},
		},
		{
			name: "Error: GetAZEntityInformation unknown azure entity - Internal Server Error", // TODO: Should this actually be a 500 or a 400?
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "object_id=id&counts=true",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"db error: unknown azure entity "}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?object_id=id&counts=true"}},
			},
		},
		{
			name: "Success: GetAZEntity - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "object_id=id&counts=true",
					},
				}

				param := map[string]string{
					"entity_type": "roles",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
				t.Helper()
				mocks.mockDB.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   []byte(`{"data":{"kind":"","props":null,"active_assignments":0,"pim_assignments":0}}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?object_id=id&counts=true"}},
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
			testCase.emulateWithMocks(t, mocks, request)

			resouces := v2.Resources{
				Graph: mocks.mockDB,
			}

			response := httptest.NewRecorder()

			resouces.GetAZEntity(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			require.Equal(t, testCase.expected.responseBody, body)
		})
	}
}
