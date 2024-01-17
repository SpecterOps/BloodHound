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

package integration

import (
	"github.com/specterops/bloodhound/src/model"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func (s *Context) ListAuthTokens() model.AuthTokens {
	listTokensResponse, err := s.AdminClient().ListAuthTokens()
	require.Nilf(s.TestCtrl, err, "Failed listing auth tokens: %v", err)

	return listTokensResponse.Tokens
}

func (s *Context) ListUserTokens(userID uuid.UUID) model.AuthTokens {
	listTokensResponse, err := s.AdminClient().ListUserTokens(userID)
	require.Nilf(s.TestCtrl, err, "Failed listing user %s tokens: %v", userID.String(), err)

	return listTokensResponse.Tokens
}

func (s *Context) DeleteAuthToken(userID, tokenID uuid.UUID) {
	err := s.AdminClient().DeleteUserToken(userID, tokenID)
	require.Nilf(s.TestCtrl, err, "Failed to delete auth token %s for user %s: %v", tokenID.String(), userID.String(), err)
}

func (s *Context) ListRoles() model.Roles {
	listRolesResponse, err := s.AdminClient().ListRoles()
	require.Nilf(s.TestCtrl, err, "Failed listing BloodHound auth roles: %v", err)

	return listRolesResponse.Roles
}

func (s *Context) GetRolesByName(roleNames ...string) model.Roles {
	var (
		roles      = s.ListRoles()
		foundRoles = make(model.Roles, len(roleNames))
	)

	for idx, roleName := range roleNames {
		role, hasRole := roles.FindByName(roleName)

		require.Truef(s.TestCtrl, hasRole, "Unable to find role: %s", roleName)
		foundRoles[idx] = role
	}

	return foundRoles
}

func (s *Context) SetUserRole(userID uuid.UUID, roleName string) {
	err := s.AdminClient().UserAddRole(userID, s.GetRolesByName(roleName)[0].ID)
	require.Nilf(s.TestCtrl, err, "Failed to set role for user %s: %v", userID.String(), err)
}

func (s *Context) RemoveUserRole(userID uuid.UUID, roleName string) {
	err := s.AdminClient().UserRemoveRole(userID, s.GetRolesByName(roleName)[0].ID)
	require.Nilf(s.TestCtrl, err, "Failed to remove role for user %s: %v", userID.String(), err)
}

func (s *Context) ListUsers() model.Users {
	listUsersResponse, err := s.AdminClient().ListUsers()
	require.Nilf(s.TestCtrl, err, "Failed to list users: %v", err)

	return listUsersResponse.Users
}

func (s *Context) AssertUserCount(expectedNumUsers int) {
	users := s.ListUsers()
	require.Equalf(s.TestCtrl, len(users), expectedNumUsers, "Expected %d users but found %d")
}
func (s *Context) SetUserSecret(userID uuid.UUID, secret string, needsPasswordReset bool) {
	err := s.AdminClient().SetUserSecret(userID, secret, needsPasswordReset)
	require.Nilf(s.TestCtrl, err, "Failed to set secert auth for user %s: %v", userID.String(), err)
}

func (s *Context) CreateUser(name, emailAddress, roleName string) model.User {
	var (
		roles        = s.GetRolesByName(roleName)
		newUser, err = s.AdminClient().CreateUser(name, emailAddress, roles.IDs())
	)

	require.Nilf(s.TestCtrl, err, "Failed to create user: %v", err)
	return newUser
}

func (s *Context) DeleteUser(userID uuid.UUID) {
	err := s.AdminClient().DeleteUser(userID)
	require.Nilf(s.TestCtrl, err, "Failed to delete user %s: %v", userID.String(), err)
}

func (s *Context) CreateAuthToken(userID uuid.UUID, tokenName string) model.AuthToken {
	newToken, err := s.AdminClient().CreateUserToken(userID, tokenName)

	require.Nilf(s.TestCtrl, err, "Failed to create auth token for user %s: %v", userID.String(), err)
	return newToken
}
