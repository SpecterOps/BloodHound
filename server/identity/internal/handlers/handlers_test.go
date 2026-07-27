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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
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
