package v2_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"

	schemamocks "github.com/specterops/bloodhound/cmd/api/src/api/v2/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestResources_GetExtensions(t *testing.T) {
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
				mock.mockOpenGraphSchemaService.EXPECT().GetExtensions(gomock.Any()).Return([]v2.ExtensionInfo{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"errors":[{"context":"","message":"error getting graph schema extensions: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
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
				mock.mockOpenGraphSchemaService.EXPECT().GetExtensions(gomock.Any()).Return([]v2.ExtensionInfo{
					{
						Id: "1",
						Name: "Display 1",
						Version: "v1.0.0",
					},
					{
						Id: "2",
						Name: "Display 2",
						Version: "v2.0.0",
					},
					{
						Id: "3",
						Name: "Display 3",
						Version: "v3.0.0",
					},
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody:   `{"data":{"extensions":[{"id":"1", "name":"Display 1", "version":"v1.0.0"}, {"id":"2", "name":"Display 2", "version":"v2.0.0"}, {"id":"3", "name":"Display 3", "version":"v3.0.0"}]}}`,
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
			router.HandleFunc("/api/v2/extensions", resources.GetExtensions).Methods("GET")

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
