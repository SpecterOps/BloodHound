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

//go:build integration

package database_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

// Exercise complete application operations against the actual migrated schema.
func TestDatabase_UserRoleUniquenessCompatibility(t *testing.T) {
	var (
		testSuite         = setupIntegrationTestSuite(t)
		roles             model.Roles
		administratorRole model.Role
		userRole          model.Role
		user              model.User
		found             bool
		err               error
	)
	t.Cleanup(func() { teardownIntegrationTestSuite(t, &testSuite) })

	roles, err = testSuite.BHDatabase.GetAllRoles(testSuite.Context, "", model.SQLFilter{})
	require.NoError(t, err)
	administratorRole, found = roles.FindByName(auth.RoleAdministrator)
	require.True(t, found)
	userRole, found = roles.FindByName(auth.RoleUser)
	require.True(t, found)

	user, err = testSuite.BHDatabase.CreateUser(testSuite.Context, model.User{
		PrincipalName: "role-uniqueness@example.com",
		Roles:         model.Roles{administratorRole},
	})
	require.NoError(t, err)

	// Prove the constraint is active before exercising the application operation.
	err = testSuite.DB.Exec(
		"INSERT INTO users_roles (user_id, role_id) VALUES (?, ?)", user.ID, userRole.ID,
	).Error
	require.ErrorContains(t, err, "users_roles_user_id_key")

	for _, testCase := range []struct {
		name      string
		roles     model.Roles
		firstName string
	}{
		{name: "replace administrator with user", roles: model.Roles{userRole}},
		{name: "save unchanged role", roles: model.Roles{userRole}},
		{name: "update profile retaining role", roles: model.Roles{userRole}, firstName: "Updated"},
		{name: "remove role", roles: model.Roles{}},
		{name: "assign role to roleless user", roles: model.Roles{administratorRole}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var persistedRoleIDs []int32

			user.Roles = testCase.roles
			if testCase.firstName != "" {
				user.FirstName.SetValid(testCase.firstName)
			}
			require.NoError(t, testSuite.BHDatabase.UpdateUser(testSuite.Context, user))
			require.NoError(t, testSuite.DB.Table("users_roles").Where("user_id = ?", user.ID).
				Pluck("role_id", &persistedRoleIDs).Error)
			require.ElementsMatch(t, testCase.roles.IDs(), persistedRoleIDs)

			if testCase.firstName != "" {
				persistedUser, err := testSuite.BHDatabase.GetUser(testSuite.Context, user.ID)
				require.NoError(t, err)
				require.Equal(t, testCase.firstName, persistedUser.FirstName.String)
			}
		})
	}
}
