// Copyright 2025 Specter Ops, Inc.
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
package auth_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
)

func TestResources_ListUsersMinimal(t *testing.T) {
	t.Parallel()

	user1Id, err := uuid.NewV4()
	require.NoError(t, err)
	user2Id, err := uuid.NewV4()
	require.NoError(t, err)

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}

	type expected struct {
		responseCode   int
		responseBody   string
		responseHeader http.Header
	}

	type fields struct {
		setupMocks func(t *testing.T, mock *mock)
	}

	type args struct {
		buildRequest func() *http.Request
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		expect expected
	}{
		{
			name: "fail - empty sort column",
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v2/bloodhound-users/minimal", nil)
					require.NoError(t, err)
					query := req.URL.Query()
					query.Add("sort_by", "")
					req.URL.RawQuery = query.Encode()
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"the specified column cannot be sorted because it is empty: "}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "fail - invalid sort column",
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v2/bloodhound-users/minimal", nil)
					require.NoError(t, err)
					query := req.URL.Query()
					query.Add("sort_by", "awfawf")
					req.URL.RawQuery = query.Encode()
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"the specified column cannot be sorted: awfawf"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "fail - parse query parameter filter error",
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v2/bloodhound-users/minimal", nil)
					require.NoError(t, err)
					query := req.URL.Query()
					query.Add("id", "afwa:3")
					req.URL.RawQuery = query.Encode()
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameter filters specified"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "fail - filter not supported for column",
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v2/bloodhound-users/minimal", nil)
					require.NoError(t, err)
					query := req.URL.Query()
					query.Add("awfaw", "eq:3")
					req.URL.RawQuery = query.Encode()
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"the specified column cannot be filtered: awfaw"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "fail - filter predicate not supported for specified column",
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v2/bloodhound-users/minimal", nil)
					require.NoError(t, err)
					query := req.URL.Query()
					query.Add("id", "lte:3")
					req.URL.RawQuery = query.Encode()
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"the specified filter predicate is not supported for this column: id lte"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "fail - DB error",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					mock.mockDatabase.EXPECT().GetAllActiveUsers(gomock.Any(), "id", model.SQLFilter{}).Return(nil, fmt.Errorf("db error"))
				},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v2/bloodhound-users/minimal", nil)
					require.NoError(t, err)
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "success",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mock) {
					mock.mockDatabase.EXPECT().GetAllActiveUsers(gomock.Any(), "id", model.SQLFilter{}).Return(model.Users{
						{
							FirstName: null.String{
								NullString: sql.NullString{
									String: "Test",
									Valid:  true,
								},
							},
							LastName: null.String{
								NullString: sql.NullString{
									String: "User1",
									Valid:  true,
								},
							},
							EmailAddress: null.String{
								NullString: sql.NullString{
									String: "testuser1@email.com", // should not show up in response
									Valid:  true,
								},
							},
							PrincipalName: "TestUser1",
							Unique: model.Unique{
								ID: user1Id,
							},
						},
						{
							FirstName: null.String{
								NullString: sql.NullString{
									String: "Test",
									Valid:  true,
								},
							},
							LastName: null.String{
								NullString: sql.NullString{
									String: "User2",
									Valid:  true,
								},
							},
							EmailAddress: null.String{
								NullString: sql.NullString{
									String: "testuser2@email.com", // should not show up in response
									Valid:  true,
								},
							},
							PrincipalName: "TestUser2",
							Unique: model.Unique{
								ID: user2Id,
							},
						},
					}, nil)
				},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v2/bloodhound-users/minimal", nil)
					require.NoError(t, err)
					return req
				},
			},
			expect: expected{
				responseCode:   http.StatusOK,
				responseBody:   fmt.Sprintf(`{"data":{"users":[{"first_name":"Test", "id":"%s", "last_name":"User1", "principal_name":"TestUser1"}, {"first_name":"Test", "id":"%s", "last_name":"User2", "principal_name":"TestUser2"}]}}`, user1Id, user2Id),
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			resource, mockDB := apitest.NewAuthManagementResource(mockCtrl)

			defer mockCtrl.Finish()

			if tt.fields.setupMocks != nil {
				tt.fields.setupMocks(t, &mock{mockDB})
			}
			response := httptest.NewRecorder()
			request := tt.args.buildRequest()
			resource.ListActiveUsersMinimal(response, request)
			statusCode, header, body := test.ProcessResponse(t, response)
			assert.Equal(t, tt.expect.responseCode, statusCode)
			assert.Equal(t, tt.expect.responseHeader, header)
			assert.JSONEq(t, tt.expect.responseBody, body)
		})
	}
}
