package v2_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/queries/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestManagementResource_GetAZRelatedEntities(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
	}
	type args struct {
		context          context.Context
		objectID         string
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

				param := map[string]string{
					"object_id": "id",
				}

				return mux.SetURLVars(request, param)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"query parameter \"related_entity_type\" is malformed: missing required parameter"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
			emulateWithMocks: func(t *testing.T, mocks *mock, req *http.Request) {},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

			resouces := v2.Resources{
				GraphQuery: mocks.mockGraphQuery,
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
