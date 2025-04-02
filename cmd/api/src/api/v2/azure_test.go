package v2_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/dawgs/graph"
	graphmocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/ops"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_GetAZRelatedEntities(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDB         *graphmocks.MockDatabase
		mockGraphQuery *mocks.MockGraph
	}
	type args struct {
		context  context.Context
		objectID string
	}
	type expected struct {
		responseBody   any
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		args             args
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: empty relatedEntityType - Bad Request",
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph&skip=0&limit=1&related_entity_type=inbound-control",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=descendent-users&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "related_entity_type=inbound-control&skip=0&limit=1",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {
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
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDB:         graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resouces := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				Graph:      mocks.mockDB,
			}

			response := httptest.NewRecorder()

			resouces.GetAZRelatedEntities(testCase.args.context, response, request, testCase.args.objectID)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := processResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			require.Equal(t, testCase.expected.responseBody, body)
		})
	}
}

func TestManagementResource_GetAZEntity(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDB         *graphmocks.MockDatabase
		mockGraphQuery *mocks.MockGraph
	}
	type args struct {
		context  context.Context
		objectID string
	}
	type expected struct {
		responseBody   any
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		args             args
		buildRequest     func() *http.Request
		emulateWithMocks func(t *testing.T, mock *mock, req *http.Request)
		expected         expected
	}

	tt := []testData{
		{
			name: "Error: missing parameter object ID - Bad Request",
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
			args: args{
				context:  context.Background(),
				objectID: "id",
			},
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
				mockGraphQuery: mocks.NewMockGraph(ctrl),
				mockDB:         graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resouces := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
				Graph:      mocks.mockDB,
			}

			response := httptest.NewRecorder()

			resouces.GetAZEntity(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := processResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			require.Equal(t, testCase.expected.responseBody, body)
		})
	}
}
