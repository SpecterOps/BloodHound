// Copyright 2023 Specter Ops, Inc.
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
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/packages/go/headers"

	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupUser() model.User {
	return model.User{
		FirstName:     null.String{NullString: sql.NullString{String: "John", Valid: true}},
		LastName:      null.String{NullString: sql.NullString{String: "Doe", Valid: true}},
		EmailAddress:  null.String{NullString: sql.NullString{String: "johndoe@gmail.com", Valid: true}},
		PrincipalName: "John",
		AuthTokens:    model.AuthTokens{},
	}
}

func setupUserCtx(user model.User) context.Context {
	return context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{
		AuthCtx: auth.Context{
			PermissionOverrides: auth.PermissionOverrides{},
			Owner:               user,
			Session:             model.UserSession{},
		},
	})
}

func TestResources_ListFileUploadJobs(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbmocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resources.ListIngestJobs).
		Run([]apitest.Case{
			apitest.NewSortingErrorCase(),
			apitest.NewColumnNotFilterableCase(),
			apitest.NewInvalidFilterPredicateCase("id"),
			{
				Name: "GetAllIngestJobsDatabaseError",
				Setup: func() {
					mockDB.EXPECT().GetAllIngestJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.AddQueryParam(input, "skip", "1")
					apitest.AddQueryParam(input, "limit", "2")
					apitest.AddQueryParam(input, "sort_by", "start_time")
					apitest.AddQueryParam(input, "user_id", "eq:123")
				},
				Setup: func() {
					mockDB.EXPECT().GetAllIngestJobs(gomock.Any(), 1, 2, "start_time", model.SQLFilter{SQLString: "user_id = 123"}).Return([]model.IngestJob{}, 0, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})

}

func TestResources_StartIngestJob(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *dbmocks.MockDatabase
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
			name: "Error: Database Error - 500",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/start"}, Method: http.MethodPost,
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
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().CreateIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{}, errors.New("db error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"", "message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		}, {
			name: "Error: Unauthorized - 401",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/start"}, Method: http.MethodPost,
				}

				requestCtx := ctx.Context{
					RequestID: "id",
				}

				return request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, requestCtx.WithRequestID("id")))
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
			},
			expected: expected{
				responseCode:   http.StatusUnauthorized,
				responseBody:   `{"errors":[{"context":"", "message":"authentication is invalid"}],"http_status":401,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		}, {
			name: "Success: Happy Path - 201",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/start"}, Method: http.MethodPost,
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
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().CreateIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{
					UserID:           uuid.NullUUID{UUID: uuid.FromStringOrNil("id"), Valid: true},
					UserEmailAddress: null.NewString("email@notreal.com", true),
					User: model.User{
						PrincipalName: "name",
					},
					Status:        model.JobStatusRunning,
					StatusMessage: "",
					StartTime:     time.Time{},
					EndTime:       time.Time{},
					LastIngest:    time.Time{},
					TotalFiles:    0,
					FailedFiles:   0,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusCreated,
				responseBody:   `{"data":{"created_at":"0001-01-01T00:00:00Z", "deleted_at":{"Time":"0001-01-01T00:00:00Z", "Valid":false}, "end_time":"0001-01-01T00:00:00Z", "failed_files":0, "id":0, "last_ingest":"0001-01-01T00:00:00Z", "partial_failed_files":0, "start_time":"0001-01-01T00:00:00Z", "status":1, "status_message":"", "total_files":0, "updated_at":"0001-01-01T00:00:00Z", "user_email_address": "email@notreal.com", "user_id":"00000000-0000-0000-0000-000000000000"}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: dbmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB: mocks.mockDatabase,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(request.URL.String(), resources.StartIngestJob).Methods(request.Method)

			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func TestResources_EndIngestJob(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *dbmocks.MockDatabase
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
			name: "Error: Invalid Job  - 400",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/invalid/end"}, Method: http.MethodPost,
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
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"", "message":"id is malformed"}], "http_status":400, "request_id":"id", "timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: Invalid Job Status - 400",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/123/end"}, Method: http.MethodPost,
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
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{
					UserID:           uuid.NullUUID{UUID: uuid.FromStringOrNil("id"), Valid: true},
					UserEmailAddress: null.NewString("email@notreal.com", true),
					User:             model.User{PrincipalName: "name"},
					Status:           model.JobStatusComplete,
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"", "message":"job must be in running status to end"}], "http_status":400, "request_id":"id", "timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: Update Database Error - 500",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/123/end"}, Method: http.MethodPost,
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
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{
					UserID:           uuid.NullUUID{UUID: uuid.FromStringOrNil("id"), Valid: true},
					UserEmailAddress: null.NewString("email@notreal.com", true),
					User:             model.User{PrincipalName: "name"},
					Status:           model.JobStatusRunning,
				}, nil)
				mock.mockDatabase.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(errors.New("random error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetIngestJob Database Error - 500",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/123/end"}, Method: http.MethodPost,
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
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{}, errors.New("db error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"id","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: Happy Path - 200",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{Path: "/api/v2/file-upload/123/end"}, Method: http.MethodPost,
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
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{
					UserID:           uuid.NullUUID{UUID: uuid.FromStringOrNil("id"), Valid: true},
					UserEmailAddress: null.NewString("email@notreal.com", true),
					User:             model.User{PrincipalName: "name"},
					Status:           model.JobStatusRunning,
				}, nil)
				mock.mockDatabase.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   "",
				responseHeader: http.Header{},
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: dbmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB: mocks.mockDatabase,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/file-upload/{%s}/end", v2.FileUploadJobIdPathParameterName), resources.EndIngestJob).Methods(request.Method)

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

func TestResources_ListAcceptedFileUploadTypes(t *testing.T) {
	bytes, err := json.Marshal(ingest.AllowedFileUploadTypes)
	if err != nil {
		t.Fatalf("Error marshalling obj: %v", err)
	}
	apitest.
		NewHarness(t, v2.Resources{}.ListAcceptedFileUploadTypes).
		Run([]apitest.Case{
			{
				Name: "Success",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
					apitest.BodyContains(output, string(bytes))
				},
			},
		})
}

func TestResources_ProcessIngestTask(t *testing.T) {
	type mock struct {
		mockDatabase *dbmocks.MockDatabase
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
			name: "Error: missing content_type request header - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Content type must be application/json or application/zip"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: invalid file_upload_job_id parameter - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/id",
					},
					Method: http.MethodPost,
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"id is malformed"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: GetFileUploadJobByID database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: error saving ingest file fileupload.ErrInvalidJSON - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Body:   io.NopCloser(bytes.NewBufferString("ingest")),
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"Error saving ingest file: file is not valid json"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: error saving ingest file - Internal Server Error",
			buildRequest: func() *http.Request {
				data := map[string]interface{}{"name": "example", "value": 123}
				jsonBytes, err := json.Marshal(data)
				if err != nil {
					t.Fatalf("error marshalling json necessary for test %v", err)
				}
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Body:   io.NopCloser(bytes.NewBuffer(jsonBytes)),
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, nil)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"Error saving ingest file: no valid meta tag or data tag found"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: CreateIngestTask database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))),
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(gomock.Any(), gomock.Any()).Return(model.IngestTask{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: TouchFileUploadJobLastIngest database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))),
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(gomock.Any(), gomock.Any()).Return(model.IngestTask{}, nil)
				mock.mockDatabase.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: file uploaded - Accepted Unknown Json File",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))),
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(gomock.Any(), gomock.Cond(func(x model.IngestTask) bool {
					return x.OriginalFileName == "UnknownFileName.json"
				})).Return(model.IngestTask{}, nil)
				mock.mockDatabase.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusAccepted,
				responseHeader: http.Header{},
			},
		},
		{
			name: "Success: file uploaded - Accepted Named Json File",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))),
					Header: http.Header{
						headers.ContentType.String(): []string{"application/json"},
						v2.FileUploadFileNameHeader:  []string{"Testing.json"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(gomock.Any(), gomock.Cond(func(x model.IngestTask) bool {
					return x.OriginalFileName == "Testing.json"
				})).Return(model.IngestTask{}, nil)
				mock.mockDatabase.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusAccepted,
				responseHeader: http.Header{},
			},
		},
		{
			name: "Success: file uploaded - Accepted Named Zip File",
			buildRequest: func() *http.Request {
				buf := new(bytes.Buffer)
				zipWriter := zip.NewWriter(buf)

				zipFile, err := zipWriter.Create("example.json")
				if err != nil {
					t.Fatalf("error creating zip file: %v", err)
				}

				_, err = zipFile.Write([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))
				if err != nil {
					t.Fatalf("error creating zip file: %v", err)
				}

				err = zipWriter.Close()
				if err != nil {
					t.Fatalf("error closing zip file: %v", err)
				}

				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/file-upload/1",
					},
					Method: http.MethodPost,
					Body:   io.NopCloser(buf),
					Header: http.Header{
						headers.ContentType.String(): []string{"application/zip"},
						v2.FileUploadFileNameHeader:  []string{"Testing.zip"},
					},
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetIngestJob(gomock.Any(), int64(1)).Return(model.IngestJob{Status: model.JobStatusRunning}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(gomock.Any(), gomock.Cond(func(x model.IngestTask) bool {
					return x.OriginalFileName == "Testing.zip"
				})).Return(model.IngestTask{}, nil)
				mock.mockDatabase.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusAccepted,
				responseHeader: http.Header{},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: dbmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB:     mocks.mockDatabase,
				Config: config.Configuration{},
			}

			err := os.Mkdir(resources.Config.TempDirectory(), 0755)
			if err != nil {
				if !errors.Is(err, os.ErrExist) {
					t.Fatalf("error creating directory required for test, %v", err)
				}

			}

			defer os.RemoveAll(resources.Config.TempDirectory())

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/file-upload/{%s}", v2.FileUploadJobIdPathParameterName), resources.ProcessIngestTask).Methods(request.Method)
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

func TestIsValidContentTypeForUpload(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		want   bool
	}{
		{
			name: "Empty Content-Type",
			header: http.Header{
				"nope": []string{""},
			},
			want: false,
		},
		{
			name: "Invalid Content-Type",
			header: http.Header{
				"Content-Type": []string{"invalid"},
			},
			want: false,
		},
		{
			name: "Invalid Content Type - invalid media type",
			header: http.Header{
				"Content-Type": []string{";", ""},
			},
			want: false,
		},
		{
			name: "Valid Content-Type",
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := v2.IsValidContentTypeForUpload(testCase.header)
			require.Equal(t, testCase.want, got)
		})
	}
}
