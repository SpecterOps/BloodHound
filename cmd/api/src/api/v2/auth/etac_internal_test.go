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

package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

// noOpAuditLogger discards audit entries emitted by authorization middleware in tests.
type noOpAuditLogger struct{}

// AppendAuditLog implements auth.AuditLogger without persisting test audit entries.
func (s noOpAuditLogger) AppendAuditLog(ctx context.Context, entry model.AuditEntry) error {
	return nil
}

func TestHasValidRolesForETAC(t *testing.T) {
	rolesTemplate := auth.Roles()

	readOnly := model.Role{
		Name:        rolesTemplate[auth.RoleReadOnly].Name,
		Description: rolesTemplate[auth.RoleReadOnly].Description,
		Permissions: rolesTemplate[auth.RoleReadOnly].Permissions,
		Serial:      model.Serial{},
	}

	readOnlyNoAPIKeys := model.Role{
		Name:        rolesTemplate[auth.RoleReadOnlyNoAPIKeys].Name,
		Description: rolesTemplate[auth.RoleReadOnlyNoAPIKeys].Description,
		Permissions: rolesTemplate[auth.RoleReadOnlyNoAPIKeys].Permissions,
		Serial:      model.Serial{},
	}

	uploadOnly := model.Role{
		Name:        rolesTemplate[auth.RoleUploadOnly].Name,
		Description: rolesTemplate[auth.RoleUploadOnly].Description,
		Permissions: rolesTemplate[auth.RoleUploadOnly].Permissions,
		Serial:      model.Serial{},
	}

	auditor := model.Role{
		Name:        rolesTemplate[auth.RoleAuditor].Name,
		Description: rolesTemplate[auth.RoleAuditor].Description,
		Permissions: rolesTemplate[auth.RoleAuditor].Permissions,
		Serial:      model.Serial{},
	}

	user := model.Role{
		Name:        rolesTemplate[auth.RoleUser].Name,
		Description: rolesTemplate[auth.RoleUser].Description,
		Permissions: rolesTemplate[auth.RoleUser].Permissions,
		Serial:      model.Serial{},
	}

	powerUser := model.Role{
		Name:        rolesTemplate[auth.RolePowerUser].Name,
		Description: rolesTemplate[auth.RolePowerUser].Description,
		Permissions: rolesTemplate[auth.RolePowerUser].Permissions,
		Serial:      model.Serial{},
	}

	administrator := model.Role{
		Name:        rolesTemplate[auth.RoleAdministrator].Name,
		Description: rolesTemplate[auth.RoleAdministrator].Description,
		Permissions: rolesTemplate[auth.RoleAdministrator].Permissions,
		Serial:      model.Serial{},
	}

	require.True(t, hasValidRolesForETAC(model.Roles{readOnly, readOnlyNoAPIKeys, uploadOnly, user}))
	require.False(t, hasValidRolesForETAC(model.Roles{administrator, auditor, powerUser}))
	require.NotContains(t, readOnlyNoAPIKeys.Permissions, auth.Permissions().AuthCreateToken)
}

func TestNoAPIKeyReadOnlyRoleCannotCreateTokens(t *testing.T) {
	rolesTemplate := auth.Roles()
	authorizer := auth.NewAuthorizer(noOpAuditLogger{})

	for _, testCase := range []struct {
		name           string
		role           model.Role
		expectedStatus int
	}{
		{
			name: "Read-Only (No API Keys) is denied",
			role: model.Role{
				Name:        rolesTemplate[auth.RoleReadOnlyNoAPIKeys].Name,
				Description: rolesTemplate[auth.RoleReadOnlyNoAPIKeys].Description,
				Permissions: rolesTemplate[auth.RoleReadOnlyNoAPIKeys].Permissions,
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Read-Only remains allowed",
			role: model.Role{
				Name:        rolesTemplate[auth.RoleReadOnly].Name,
				Description: rolesTemplate[auth.RoleReadOnly].Description,
				Permissions: rolesTemplate[auth.RoleReadOnly].Permissions,
			},
			expectedStatus: http.StatusNoContent,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v2/tokens", nil)
			request = bhctx.SetRequestContext(request, &bhctx.Context{
				AuthCtx: auth.Context{
					Owner: model.User{Roles: model.Roles{testCase.role}},
				},
			})
			response := httptest.NewRecorder()

			middleware.PermissionsCheckAll(authorizer, auth.Permissions().AuthCreateToken)(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				response.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(response, request)

			require.Equal(t, testCase.expectedStatus, response.Code)
		})
	}
}
