package v2_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	graphmocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_GetEdgeComposition(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraph *graphmocks.MockDatabase
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
			name: "Error: missing edge_type parameter - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Expected edge_type parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: missing source_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Expected source_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test"}},
			},
		},
		{
			name: "Error: missing target_node parameter - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&source_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Expected target_node parameter to be set."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&source_node=test"}},
			},
		},
		{
			name: "Error: edge_type is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&edge_type=test2&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Expected only one edge_type."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&edge_type=test2&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: source_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&source_node=test2&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Expected only one source_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&source_node=test2&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: target_node is more than 1 - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&target_node=test2&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Expected only one target_node."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&target_node=test2&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: invalid edge_type - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=test&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Invalid edge requested: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=test&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: invalid startID for source_node - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=Meta&source_node=test&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Invalid value for startID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=Meta&source_node=test&target_node=test"}},
			},
		},
		{
			name: "Error: invalid endID for targetNode - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=Meta&source_node=1&target_node=test",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Invalid value for endID: test"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=Meta&source_node=1&target_node=test"}},
			},
		},
		{
			name: "Error: database error fetching edge by start and end - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=Meta&source_node=1&target_node=2",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Could not find edge matching criteria: error"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=Meta&source_node=1&target_node=2"}},
			},
		},
		{
			name: "Error: database error getting edge composition path - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=Meta&source_node=1&target_node=2",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Error getting composition for edge: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=Meta&source_node=1&target_node=2"}},
			},
		},
		{
			name: "Success: retrieved edge composition - OK",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{
						RawQuery: "edge_type=Meta&source_node=1&target_node=2",
					},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
				mock.mockGraph.EXPECT().ReadTransaction(req.Context(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   []byte(`{"data":{"nodes":{},"edges":[]}}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/?edge_type=Meta&source_node=1&target_node=2"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraph: graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resources := v2.Resources{
				Graph: mocks.mockGraph,
			}

			response := httptest.NewRecorder()

			resources.GetEdgeComposition(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			require.Equal(t, testCase.expected.responseBody, body)
		})
	}
}
