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
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRequestWithVars(t *testing.T, target string, vars map[string]string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, target, nil)
	require.NoError(t, err)
	return mux.SetURLVars(req, vars)
}

func TestHandlers_GetPermission(t *testing.T) {
	var (
		unexpectedErr = errors.New("unexpected database failure")
		expected      = services.Permission{ID: 7, Authority: "app", Name: "ManageProviders"}
	)

	tests := []struct {
		name       string
		rawID      string
		expect     func(m *mocks.MockIdentity, ctx context.Context)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:  "returns 200 with the permission view on success",
			rawID: "7",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().GetPermission(ctx, 7).Return(expected, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.PermissionView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, expected.ID, envelope.Data.ID)
				assert.Equal(t, expected.Authority, envelope.Data.Authority)
				assert.Equal(t, expected.Name, envelope.Data.Name)
			},
		},
		{
			name:       "returns 400 for a malformed permission ID",
			rawID:      "not-an-int",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "returns 404 when the permission does not exist",
			rawID: "7",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().GetPermission(ctx, 7).Return(services.Permission{}, services.ErrNoPermissionFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "returns 500 on unexpected service error",
			rawID: "7",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().GetPermission(ctx, 7).Return(services.Permission{}, unexpectedErr)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				recorder     = httptest.NewRecorder()
				request      = newRequestWithVars(t, "/api/v2/permissions/"+tt.rawID, map[string]string{"permission_id": tt.rawID})
			)

			if tt.expect != nil {
				tt.expect(identityMock, request.Context())
			}

			handlerSet.GetPermission(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestHandlers_GetRole(t *testing.T) {
	var (
		unexpectedErr = errors.New("unexpected database failure")
		expected      = services.Role{
			ID:          3,
			Name:        "Administrator",
			Description: "Can manage the application",
			Permissions: []services.Permission{{ID: 1, Authority: "app", Name: "ManageProviders"}},
		}
	)

	tests := []struct {
		name       string
		rawID      string
		expect     func(m *mocks.MockIdentity, ctx context.Context)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:  "returns 200 with the role view on success",
			rawID: "3",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().GetRole(ctx, int32(3)).Return(expected, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RoleView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, expected.ID, envelope.Data.ID)
				assert.Equal(t, expected.Name, envelope.Data.Name)
				assert.Equal(t, expected.Description, envelope.Data.Description)
				require.Len(t, envelope.Data.Permissions, 1)
				assert.Equal(t, expected.Permissions[0].Name, envelope.Data.Permissions[0].Name)
			},
		},
		{
			name:       "returns 400 for a malformed role ID",
			rawID:      "not-an-int",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "returns 404 when the role does not exist",
			rawID: "3",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().GetRole(ctx, int32(3)).Return(services.Role{}, services.ErrNoRoleFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "returns 500 on unexpected service error",
			rawID: "3",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().GetRole(ctx, int32(3)).Return(services.Role{}, unexpectedErr)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				recorder     = httptest.NewRecorder()
				request      = newRequestWithVars(t, "/api/v2/roles/"+tt.rawID, map[string]string{"role_id": tt.rawID})
			)

			if tt.expect != nil {
				tt.expect(identityMock, request.Context())
			}

			handlerSet.GetRole(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestHandlers_ListRoles(t *testing.T) {
	var (
		unexpectedErr = errors.New("unexpected database failure")
		filters       = params.Filters{
			"name": {{Field: "name", Operator: params.Equals, Value: "Administrator", IsStringData: true}},
		}
		sortItems = params.SortItems{{Field: "name", Direction: params.Ascending}}
		expected  = []services.Role{
			{
				ID:          3,
				Name:        "Administrator",
				Description: "Can manage the application",
				Permissions: []services.Permission{{ID: 1, Authority: "app", Name: "ManageProviders"}},
			},
		}
	)

	tests := []struct {
		name       string
		expect     func(m *mocks.MockIdentity, ctx context.Context)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name: "returns 200 with the role list view on success",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().ListRoles(ctx, filters, sortItems).Return(expected, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RoleListView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				require.Len(t, envelope.Data.Roles, 1)
				assert.Equal(t, expected[0].ID, envelope.Data.Roles[0].ID)
				assert.Equal(t, expected[0].Name, envelope.Data.Roles[0].Name)
			},
		},
		{
			name: "returns 200 with an empty roles array when no roles match",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().ListRoles(ctx, filters, sortItems).Return([]services.Role{}, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RoleListView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Empty(t, envelope.Data.Roles)
			},
		},
		{
			name: "returns 500 on unexpected service error",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().ListRoles(ctx, filters, sortItems).Return(nil, unexpectedErr)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				recorder     = httptest.NewRecorder()
				request      = newRequestWithVars(t, "/api/v2/roles", nil)
			)

			request = bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})

			if tt.expect != nil {
				tt.expect(identityMock, request.Context())
			}

			handlerSet.ListRoles(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestHandlers_ListUsers(t *testing.T) {
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

	tests := []struct {
		name       string
		expect     func(m *mocks.MockIdentity, ctx context.Context)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name: "returns 200 reproducing the legacy user JSON contract",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().ListUsers(ctx, filters, sortItems).Return(populated, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
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
		{
			name: "returns null nullable fields and AuthSecret when unset",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().ListUsers(ctx, filters, sortItems).Return(bare, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
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
		{
			name: "returns 200 with an empty users array when no users match",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().ListUsers(ctx, filters, sortItems).Return([]services.User{}, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.UserListView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Empty(t, envelope.Data.Users)
			},
		},
		{
			name: "returns 500 on unexpected service error",
			expect: func(m *mocks.MockIdentity, ctx context.Context) {
				m.EXPECT().ListUsers(ctx, filters, sortItems).Return(nil, unexpectedErr)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				identityMock = mocks.NewMockIdentity(t)
				handlerSet   = handlers.NewHandlersContainer(identityMock)
				recorder     = httptest.NewRecorder()
				request      = newRequestWithVars(t, "/api/v2/bloodhound-users", nil)
			)

			request = bhctx.SetRequestContext(request, &bhctx.Context{Filters: filters, Sort: sortItems})

			if tt.expect != nil {
				tt.expect(identityMock, request.Context())
			}

			handlerSet.ListUsers(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
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
