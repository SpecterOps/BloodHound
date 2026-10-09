// Copyright 2026 Specter Ops, Inc.
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

package handlers_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	testutils "github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlers_GetPermission(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock mock)
		expected     expected
	}

	var (
		unexpectedErr      = errors.New("unexpected database failure")
		expectedPermission = services.Permission{ID: 7, Authority: "app", Name: "ManageProviders"}
	)

	tt := []testData{
		{
			name: "Success: permission view is returned - 200",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/7", nil)
			},
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().GetPermission(testifyMock.Anything, 7).Return(expectedPermission, nil)
			},
			expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.PermissionView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, expectedPermission.ID, envelope.Data.ID)
				assert.Equal(t, expectedPermission.Authority, envelope.Data.Authority)
				assert.Equal(t, expectedPermission.Name, envelope.Data.Name)
			}},
		},
		{
			name: "Error: permission ID is malformed - 400",
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/not-an-int", nil)
			},
			expected: expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name:         "Error: permission does not exist - 404",
			buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/7", nil) },
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().GetPermission(testifyMock.Anything, 7).Return(services.Permission{}, services.ErrNoPermissionFound)
			},
			expected: expected{responseCode: http.StatusNotFound, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name:         "Error: service fails - 500",
			buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/permissions/7", nil) },
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().GetPermission(testifyMock.Anything, 7).Return(services.Permission{}, unexpectedErr)
			},
			expected: expected{responseCode: http.StatusInternalServerError, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				muxRouter    = mux.NewRouter()
				recorder     = httptest.NewRecorder()
				request      = testCase.buildRequest()
			)
			muxRouter.HandleFunc("/api/v2/permissions/{permission_id}", handlerSet.GetPermission).Methods(http.MethodGet)

			if testCase.setupMocks != nil {
				testCase.setupMocks(mock{identity: identityMock})
			}

			muxRouter.ServeHTTP(recorder, request)

			status, header, body := testutils.ProcessResponse(t, recorder)
			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, []byte(body))
			}
		})
	}
}

func TestHandlers_GetRole(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock mock)
		expected     expected
	}

	var (
		unexpectedErr = errors.New("unexpected database failure")
		expectedRole  = services.Role{
			ID:          3,
			Name:        "Administrator",
			Description: "Can manage the application",
			Permissions: []services.Permission{{ID: 1, Authority: "app", Name: "ManageProviders"}},
		}
	)

	tt := []testData{
		{
			name:         "Success: role view is returned - 200",
			buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/roles/3", nil) },
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().GetRole(testifyMock.Anything, int32(3)).Return(expectedRole, nil)
			},
			expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RoleView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, expectedRole.ID, envelope.Data.ID)
				assert.Equal(t, expectedRole.Name, envelope.Data.Name)
				assert.Equal(t, expectedRole.Description, envelope.Data.Description)
				require.Len(t, envelope.Data.Permissions, 1)
				assert.Equal(t, expectedRole.Permissions[0].Name, envelope.Data.Permissions[0].Name)
			}},
		},
		{
			name:         "Error: role ID is malformed - 400",
			buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/roles/not-an-int", nil) },
			expected:     expected{responseCode: http.StatusBadRequest, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name:         "Error: role does not exist - 404",
			buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/roles/3", nil) },
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().GetRole(testifyMock.Anything, int32(3)).Return(services.Role{}, services.ErrNoRoleFound)
			},
			expected: expected{responseCode: http.StatusNotFound, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
		{
			name:         "Error: service fails - 500",
			buildRequest: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v2/roles/3", nil) },
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().GetRole(testifyMock.Anything, int32(3)).Return(services.Role{}, unexpectedErr)
			},
			expected: expected{responseCode: http.StatusInternalServerError, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				muxRouter    = mux.NewRouter()
				recorder     = httptest.NewRecorder()
				request      = testCase.buildRequest()
			)
			muxRouter.HandleFunc("/api/v2/roles/{role_id}", handlerSet.GetRole).Methods(http.MethodGet)

			if testCase.setupMocks != nil {
				testCase.setupMocks(mock{identity: identityMock})
			}

			muxRouter.ServeHTTP(recorder, request)

			status, header, body := testutils.ProcessResponse(t, recorder)
			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, []byte(body))
			}
		})
	}
}

func TestHandlers_ListRoles(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCode   int
		responseHeader http.Header
		assertBody     func(t *testing.T, body []byte)
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock mock)
		expected     expected
	}

	var (
		unexpectedErr = errors.New("unexpected database failure")
		filters       = params.Filters{
			"name": {{Field: "name", Operator: params.Equals, Value: "Administrator", IsStringData: true}},
		}
		sortItems     = params.SortItems{{Field: "name", Direction: params.Ascending}}
		expectedRoles = []services.Role{
			{
				ID:          3,
				Name:        "Administrator",
				Description: "Can manage the application",
				Permissions: []services.Permission{{ID: 1, Authority: "app", Name: "ManageProviders"}},
			},
		}
	)

	tt := []testData{
		{
			name: "Success: role list view is returned - 200",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
				return bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})
			},
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().ListRoles(testifyMock.Anything, filters, sortItems).Return(expectedRoles, nil)
			},
			expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RoleListView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				require.Len(t, envelope.Data.Roles, 1)
				assert.Equal(t, expectedRoles[0].ID, envelope.Data.Roles[0].ID)
				assert.Equal(t, expectedRoles[0].Name, envelope.Data.Roles[0].Name)
			}},
		},
		{
			name: "Success: no roles match - 200",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
				return bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})
			},
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().ListRoles(testifyMock.Anything, filters, sortItems).Return([]services.Role{}, nil)
			},
			expected: expected{responseCode: http.StatusOK, responseHeader: http.Header{"Content-Type": []string{"application/json"}}, assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RoleListView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Empty(t, envelope.Data.Roles)
			}},
		},
		{
			name: "Error: service fails - 500",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/roles", nil)
				return bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})
			},
			setupMocks: func(mock mock) {
				mock.identity.EXPECT().ListRoles(testifyMock.Anything, filters, sortItems).Return(nil, unexpectedErr)
			},
			expected: expected{responseCode: http.StatusInternalServerError, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				muxRouter    = mux.NewRouter()
				recorder     = httptest.NewRecorder()
				request      = testCase.buildRequest()
			)
			muxRouter.HandleFunc("/api/v2/roles", handlerSet.ListRoles).Methods(http.MethodGet)

			if testCase.setupMocks != nil {
				testCase.setupMocks(mock{identity: identityMock})
			}

			muxRouter.ServeHTTP(recorder, request)

			status, header, body := testutils.ProcessResponse(t, recorder)
			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if testCase.expected.assertBody != nil {
				testCase.expected.assertBody(t, []byte(body))
			}
		})
	}
}

func TestHandlers_ListPermissions(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCode   int
		responseBody   func(t *testing.T, body []byte)
		responseHeader http.Header
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock *mock)
		expected     expected
	}

	var (
		unexpectedErr = errors.New("unexpected database failure")
		filters       = params.Filters{
			"authority": {{Field: "authority", Operator: params.Equals, Value: "app", IsStringData: true}},
		}
		sortItems           = params.SortItems{{Field: "name", Direction: params.Ascending}}
		expectedPermissions = []services.Permission{{ID: 7, Authority: "app", Name: "ManageProviders"}}
	)

	tt := []testData{
		{
			name: "Success: permissions are returned - 200",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
				return bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})
			},
			setupMocks: func(mock *mock) {
				mock.identity.EXPECT().ListPermissions(testifyMock.Anything, filters, sortItems).Return(expectedPermissions, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody: func(t *testing.T, body []byte) {
					var envelope struct {
						Data handlers.PermissionListView `json:"data"`
					}
					require.NoError(t, json.Unmarshal(body, &envelope))
					require.Len(t, envelope.Data.Permissions, 1)
					assert.Equal(t, expectedPermissions[0].ID, envelope.Data.Permissions[0].ID)
					assert.Equal(t, expectedPermissions[0].Authority, envelope.Data.Permissions[0].Authority)
					assert.Equal(t, expectedPermissions[0].Name, envelope.Data.Permissions[0].Name)
				},
			},
		},
		{
			name: "Success: no permissions match - 200",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
				return bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})
			},
			setupMocks: func(mock *mock) {
				mock.identity.EXPECT().ListPermissions(testifyMock.Anything, filters, sortItems).Return([]services.Permission{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody: func(t *testing.T, body []byte) {
					var envelope struct {
						Data handlers.PermissionListView `json:"data"`
					}
					require.NoError(t, json.Unmarshal(body, &envelope))
					assert.Empty(t, envelope.Data.Permissions)
				},
			},
		},
		{
			name: "Error: service fails - 500",
			buildRequest: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
				return bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})
			},
			setupMocks: func(mock *mock) {
				mock.identity.EXPECT().ListPermissions(testifyMock.Anything, filters, sortItems).Return(nil, unexpectedErr)
			},
			expected: expected{responseCode: http.StatusInternalServerError, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				muxRouter    = mux.NewRouter()
				recorder     = httptest.NewRecorder()
			)

			muxRouter.HandleFunc("/api/v2/permissions", handlerSet.ListPermissions).Methods(http.MethodGet)
			testCase.setupMocks(&mock{identity: identityMock})

			muxRouter.ServeHTTP(recorder, testCase.buildRequest())

			status, header, body := testutils.ProcessResponse(t, recorder)
			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if testCase.expected.responseBody != nil {
				testCase.expected.responseBody(t, []byte(body))
			}
		})
	}
}

func TestPermissionListView_DeletedAtIsNotQueryable(t *testing.T) {
	var view = handlers.PermissionListView{}

	_, isFilterable := view.ValidFilters()["deleted_at"]
	assert.False(t, isFilterable)
	assert.False(t, view.IsSortable("deleted_at"))

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("the request should be rejected by filter middleware")
	})
	handler := middleware.FilterMiddleware(view)(next)
	request := httptest.NewRequest(http.MethodGet, "/api/v2/permissions?deleted_at=eq:null", nil)
	request = bhctx.SetRequestContext(request, &bhctx.Context{})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlers_ListUsers(t *testing.T) {
	type mock struct {
		identity *mocks.MockIdentity
	}

	type expected struct {
		responseCode   int
		responseBody   func(t *testing.T, body []byte)
		responseHeader http.Header
	}

	type testData struct {
		name         string
		buildRequest func() *http.Request
		setupMocks   func(mock *mock)
		expected     expected
	}

	var (
		unexpectedErr = errors.New("unexpected database failure")
		filters       = params.Filters{
			"email_address": {{Field: "email_address", Operator: params.Equals, Value: "ada@example.com", IsStringData: true}},
		}
		sortItems = params.SortItems{{Field: "last_login", Direction: params.Descending}}
		userID    = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
		lastLogin = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		populated = []services.User{
			{
				ID:              userID,
				SSOProviderID:   sql.NullInt32{Int32: 7, Valid: true},
				FirstName:       sql.NullString{String: "Ada", Valid: true},
				LastName:        sql.NullString{Valid: false},
				EmailAddress:    sql.NullString{String: "ada@example.com", Valid: true},
				PrincipalName:   "ada",
				LastLogin:       lastLogin,
				AllEnvironments: true,
				EULAAccepted:    true,
				Roles: []services.Role{
					{ID: 3, Name: "Administrator", Permissions: []services.Permission{{ID: 1, Authority: "app", Name: "ManageProviders"}}},
				},
				EnvironmentTargetedAccessControl: []services.EnvironmentAccessControl{
					{ID: 5, UserID: userID.String(), EnvironmentID: "env-1"},
				},
				AuthSecret: &services.AuthSecret{ID: 9, DigestMethod: "argon2", TOTPActivated: true},
			},
		}
		bare = []services.User{{ID: userID, PrincipalName: "guest"}}
	)

	buildRequest := func() *http.Request {
		request := httptest.NewRequest(http.MethodGet, "/api/v2/bloodhound-users", nil)
		return bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})
	}

	tt := []testData{
		{
			name:         "Success: legacy user JSON contract is reproduced - 200",
			buildRequest: buildRequest,
			setupMocks: func(mock *mock) {
				mock.identity.EXPECT().ListUsers(testifyMock.Anything, filters, sortItems).Return(populated, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody: func(t *testing.T, body []byte) {
					var envelope struct {
						Data struct {
							Users []json.RawMessage `json:"users"`
						} `json:"data"`
					}
					require.NoError(t, json.Unmarshal(body, &envelope))
					require.Len(t, envelope.Data.Users, 1)

					var fields map[string]json.RawMessage
					require.NoError(t, json.Unmarshal(envelope.Data.Users[0], &fields))

					// Nullable fields marshal to a bare value or null (legacy null.String/null.Int32 contract).
					assert.JSONEq(t, `7`, string(fields["sso_provider_id"]))
					assert.JSONEq(t, `"Ada"`, string(fields["first_name"]))
					assert.JSONEq(t, `null`, string(fields["last_name"]))
					assert.JSONEq(t, `"ada@example.com"`, string(fields["email_address"]))

					// AuthSecret retains its legacy capitalized key and omits secret material.
					require.Contains(t, fields, "AuthSecret")
					var authSecret map[string]json.RawMessage
					require.NoError(t, json.Unmarshal(fields["AuthSecret"], &authSecret))
					assert.JSONEq(t, `"argon2"`, string(authSecret["digest_method"]))
					assert.NotContains(t, authSecret, "digest")
					assert.NotContains(t, authSecret, "totp_secret")

					// deleted_at retains the sql.NullTime object shape used by the legacy model.
					assert.JSONEq(t, `{"Time":"0001-01-01T00:00:00Z","Valid":false}`, string(fields["deleted_at"]))

					// Associations are always present as arrays.
					assert.Contains(t, fields, "roles")
					assert.Contains(t, fields, "environment_targeted_access_control")
				},
			},
		},
		{
			name:         "Success: nullable fields and AuthSecret render null when unset - 200",
			buildRequest: buildRequest,
			setupMocks: func(mock *mock) {
				mock.identity.EXPECT().ListUsers(testifyMock.Anything, filters, sortItems).Return(bare, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody: func(t *testing.T, body []byte) {
					var envelope struct {
						Data struct {
							Users []json.RawMessage `json:"users"`
						} `json:"data"`
					}
					require.NoError(t, json.Unmarshal(body, &envelope))
					require.Len(t, envelope.Data.Users, 1)

					var fields map[string]json.RawMessage
					require.NoError(t, json.Unmarshal(envelope.Data.Users[0], &fields))
					assert.JSONEq(t, `null`, string(fields["sso_provider_id"]))
					assert.JSONEq(t, `null`, string(fields["first_name"]))
					assert.JSONEq(t, `null`, string(fields["AuthSecret"]))
				},
			},
		},
		{
			name:         "Success: no users match - 200",
			buildRequest: buildRequest,
			setupMocks: func(mock *mock) {
				mock.identity.EXPECT().ListUsers(testifyMock.Anything, filters, sortItems).Return([]services.User{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
				responseBody: func(t *testing.T, body []byte) {
					var envelope struct {
						Data handlers.UserListView `json:"data"`
					}
					require.NoError(t, json.Unmarshal(body, &envelope))
					assert.Empty(t, envelope.Data.Users)
				},
			},
		},
		{
			name:         "Error: service fails - 500",
			buildRequest: buildRequest,
			setupMocks: func(mock *mock) {
				mock.identity.EXPECT().ListUsers(testifyMock.Anything, filters, sortItems).Return(nil, unexpectedErr)
			},
			expected: expected{responseCode: http.StatusInternalServerError, responseHeader: http.Header{"Content-Type": []string{"application/json"}}},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				muxRouter    = mux.NewRouter()
				recorder     = httptest.NewRecorder()
			)

			muxRouter.HandleFunc("/api/v2/bloodhound-users", handlerSet.ListUsers).Methods(http.MethodGet)
			testCase.setupMocks(&mock{identity: identityMock})

			muxRouter.ServeHTTP(recorder, testCase.buildRequest())

			status, header, body := testutils.ProcessResponse(t, recorder)
			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			if testCase.expected.responseBody != nil {
				testCase.expected.responseBody(t, []byte(body))
			}
		})
	}
}

func TestUserListView_IsSortable(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  bool
	}{
		{name: "first_name is sortable", field: "first_name", want: true},
		{name: "last_name is sortable", field: "last_name", want: true},
		{name: "email_address is sortable", field: "email_address", want: true},
		{name: "principal_name is sortable", field: "principal_name", want: true},
		{name: "last_login is sortable", field: "last_login", want: true},
		{name: "created_at is sortable", field: "created_at", want: true},
		{name: "updated_at is sortable", field: "updated_at", want: true},
		{name: "deleted_at is not sortable", field: "deleted_at", want: false},
		{name: "unknown field is not sortable", field: "nope", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, handlers.UserListView{}.IsSortable(tt.field))
		})
	}
}

func TestUserListView_ValidFilters(t *testing.T) {
	var validFilters = handlers.UserListView{}.ValidFilters()

	tests := []struct {
		name         string
		field        string
		wantPresent  bool
		wantString   bool
		wantOperator params.FilterOperator
	}{
		{name: "first_name filterable as string data", field: "first_name", wantPresent: true, wantString: true, wantOperator: params.Equals},
		{name: "principal_name filterable as string data", field: "principal_name", wantPresent: true, wantString: true, wantOperator: params.NotEquals},
		{name: "id filterable as non-string data", field: "id", wantPresent: true, wantString: false, wantOperator: params.Equals},
		{name: "last_login supports range operators", field: "last_login", wantPresent: true, wantString: false, wantOperator: params.GreaterThan},
		{name: "created_at supports range operators", field: "created_at", wantPresent: true, wantString: false, wantOperator: params.LessThanOrEquals},
		{name: "deleted_at is not filterable", field: "deleted_at", wantPresent: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, present := validFilters[tt.field]
			require.Equal(t, tt.wantPresent, present)
			if !tt.wantPresent {
				return
			}
			assert.Equal(t, tt.wantString, field.IsStringData)
			assert.Contains(t, field.Operators, tt.wantOperator)
		})
	}
}
