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
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"

	"github.com/specterops/bloodhound/src/utils/test"
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
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resources.ListFileUploadJobs).
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
					mockDB.EXPECT().GetAllIngestJobs(gomock.Any(), 1, 2, "start_time", model.SQLFilter{SQLString: "user_id = ?", Params: []any{"123"}}).Return([]model.IngestJob{}, 0, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})

}

func TestResources_StartFileUploadJob(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
		user      = setupUser()
		userCtx   = setupUserCtx(user)
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resources.StartFileUploadJob).
		Run([]apitest.Case{
			{
				Name: "Unauthorized",
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusUnauthorized)
				},
			},
			{
				Name: "DatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().CreateIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{}, errors.New("db error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetContext(input, userCtx)
				},
				Setup: func() {
					mockDB.EXPECT().CreateIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusCreated)
				},
			},
		})
}

func TestResources_EndFileUploadJob(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)
	defer mockCtrl.Finish()

	apitest.
		NewHarness(t, resources.EndFileUploadJob).
		Run([]apitest.Case{
			{
				Name: "InvalidJobID",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "invalid")
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
				},
			},
			{
				Name: "GetIngestJobDatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{}, errors.New("db error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "InvalidJobStatus",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{
						Status: model.JobStatusComplete,
					}, nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "job must be in running status")
				},
			},
			{
				Name: "UpdateIngestJobDatabaseError",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{
						Status: model.JobStatusRunning,
					}, nil)
					mockDB.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
				},
			},
			{
				Name: "Success",
				Input: func(input *apitest.Input) {
					apitest.SetURLVar(input, v2.FileUploadJobIdPathParameterName, "123")
				},
				Setup: func() {
					mockDB.EXPECT().GetIngestJob(gomock.Any(), gomock.Any()).Return(model.IngestJob{
						Status: model.JobStatusRunning,
					}, nil)
					mockDB.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).Return(nil)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusOK)
				},
			},
		})
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

func TestManagementResource_ProcessFileUpload(t *testing.T) {
	type mock struct {
		mockDatabase *mocks.MockDatabase
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
			name: "Error: missing content_type request header - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL: &url.URL{},
				}

				return request
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Content type must be application/json or application/zip"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: missing file_upload_job_id parameter - Bad Request",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				request.Header.Add("Content-type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"id is malformed."}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: GetFileUploadJobByID database error - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				request.Header.Add("Content-type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "1",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFileUploadJob(req.Context(), int64(1)).Return(model.FileUploadJob{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: GetFileUploadJobByID database error - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				}

				request.Header.Add("Content-type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "1",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFileUploadJob(req.Context(), int64(1)).Return(model.FileUploadJob{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: error saving ingest file fileupload.ErrInvalidJSON - Internal Server Error",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
					Body:   io.NopCloser(bytes.NewBufferString("ingest")),
				}

				request.Header.Add("Content-type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "1",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFileUploadJob(req.Context(), int64(1)).Return(model.FileUploadJob{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Error saving ingest file: file is not valid json"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
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

				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
					Body:   io.NopCloser(bytes.NewBuffer(jsonBytes)),
				}

				request.Header.Add("Content-type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "1",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFileUploadJob(req.Context(), int64(1)).Return(model.FileUploadJob{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"Error saving ingest file: no valid meta tag or data tag found"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: CreateIngestTask database error - Internal Server Error",
			buildRequest: func() *http.Request {

				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))),
				}

				request.Header.Set("Content-Type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "1",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFileUploadJob(req.Context(), int64(1)).Return(model.FileUploadJob{}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(req.Context(), gomock.Any()).Return(model.IngestTask{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Error: TouchFileUploadJobLastIngest database error - Internal Server Error",
			buildRequest: func() *http.Request {

				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))),
				}

				request.Header.Set("Content-Type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "1",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFileUploadJob(req.Context(), int64(1)).Return(model.FileUploadJob{}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(req.Context(), gomock.Any()).Return(model.IngestTask{}, nil)
				mock.mockDatabase.EXPECT().UpdateFileUploadJob(req.Context(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   []byte(`{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}, "Location": []string{"/"}},
			},
		},
		{
			name: "Success: file uploaded - Accepted",
			buildRequest: func() *http.Request {
				request := &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
					Body:   io.NopCloser(bytes.NewReader([]byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`))),
				}

				request.Header.Set("Content-Type", "application/json")

				param := map[string]string{
					"file_upload_job_id": "1",
				}

				return mux.SetURLVars(request, param)
			},
			emulateWithMocks: func(t *testing.T, mock *mock, req *http.Request) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFileUploadJob(req.Context(), int64(1)).Return(model.FileUploadJob{}, nil)
				mock.mockDatabase.EXPECT().CreateIngestTask(req.Context(), gomock.Any()).Return(model.IngestTask{}, nil)
				mock.mockDatabase.EXPECT().UpdateFileUploadJob(req.Context(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusAccepted,
				responseBody:   []byte(``),
				responseHeader: http.Header{"Location": []string{"/"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.emulateWithMocks(t, mocks, request)

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

			resources.ProcessFileUpload(response, request)
			mux.NewRouter().ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			require.Equal(t, testCase.expected.responseCode, status)
			require.Equal(t, testCase.expected.responseHeader, header)
			require.Equal(t, testCase.expected.responseBody, body)
		})
	}
}

func TestIsValidContentTypeForUpload(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			got := v2.IsValidContentTypeForUpload(testCase.header)
			require.Equal(t, testCase.want, got)
		})
	}
}
