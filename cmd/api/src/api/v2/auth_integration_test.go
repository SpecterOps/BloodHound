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

//go:build serial_integration
// +build serial_integration

package v2_test

import (
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/stretchr/testify/require"
)

const (
	otherUser   = "other@example.com"
	otherSecret = "secretTERCES123***"
	nonAdmin    = "not@dmin.com"
)

func Test_PermissionHandling(t *testing.T) {
	var (
		testCtx       = integration.NewFOSSContext(t)
		newUser       = testCtx.CreateUser(otherUser, otherUser, auth.RoleReadOnly)
		newUserToken  = testCtx.CreateAuthToken(newUser.ID, "TestToken")
		newUserClient = testCtx.NewAPIClientWithToken(newUserToken)

		_, versionErr   = newUserClient.Version()
		_, listUsersErr = newUserClient.ListUsers()
	)

	require.Nil(t, versionErr, "Version endpoint should allow requests from users with no permissions: %v", versionErr)
	require.NotNil(t, listUsersErr, "Users endpoint should not allow requests from users with no permissions: %v", listUsersErr)
}

func Test_AuthRolesMatchInternalDefinitions(t *testing.T) {
	var (
		testCtx     = integration.NewFOSSContext(t)
		actualRoles = testCtx.ListRoles()
	)

	require.Equal(t, len(auth.Roles()), len(actualRoles))

	for _, expectedRole := range auth.Roles() {
		if _, hasRole := actualRoles.FindByName(expectedRole.Name); !hasRole {
			t.Fatalf("Unable to find expected role %s", expectedRole.Name)
		}
	}
}

func Test_UserManagement(t *testing.T) {
	var (
		testCtx = integration.NewFOSSContext(t)
		newUser = testCtx.CreateUser(otherUser, otherUser, auth.RoleReadOnly)
	)

	t.Run("Delete User", func(t *testing.T) {
		// Expect to start with two users - admin and the new user
		testCtx.AssertUserCount(2)

		// Delete the user and expect to see one user remaining on the next count assertion
		testCtx.DeleteUser(newUser.ID)
		testCtx.AssertUserCount(1)

		// Exit reassigning a new, live user to the new user variable
		newUser = testCtx.CreateUser(otherUser, otherUser, auth.RoleReadOnly)
		testCtx.AssertUserCount(2)
	})

	t.Run("Set User Secret", func(t *testing.T) {
		var (
			newUserToken  = testCtx.CreateAuthToken(newUser.ID, "Set User Secret")
			newUserClient = testCtx.NewAPIClientWithToken(newUserToken)
			badSecretErr  = newUserClient.SetUserSecret(newUser.ID, "badsecret", false)
		)

		require.NotNil(t, badSecretErr, "Expected an error when using a bad user secret")

		testCtx.SetUserSecret(newUser.ID, integration.AdminUpdatedSecret, true)

		loginResponse, err := newUserClient.LoginSecret(newUser.EmailAddress.String, integration.AdminUpdatedSecret)
		require.Nilf(t, err, "Unexpected error encountered while logging in with updated secret: %v", err)
		require.True(t, loginResponse.AuthExpired, "Expected user auth to be expired after secret reset with needsPasswordSet enabled")

		testCtx.SetUserSecret(newUser.ID, otherSecret, false)

		loginResponse, err = newUserClient.LoginSecret(newUser.EmailAddress.String, otherSecret)
		require.Nilf(t, err, "Unexpected error encountered while logging in with updated secret: %v", err)
		require.False(t, loginResponse.AuthExpired, "Expected user auth to not be expired after secret reset")

		// Clean up the token
		testCtx.DeleteAuthToken(newUser.ID, newUserToken.ID)
	})

	t.Run("Create tokens for other user as an admin", func(t *testing.T) {
		testCtx.CreateAuthToken(newUser.ID, "TestToken1")
		testCtx.CreateAuthToken(newUser.ID, "TestToken2")
		testCtx.CreateAuthToken(newUser.ID, "TestToken3")

		actualTokens := testCtx.ListUserTokens(newUser.ID)

		require.Equalf(t, 3, len(actualTokens), "Expected 3 tokens but found %d", len(actualTokens))

		for _, actualToken := range actualTokens {
			// All tokens must strip their key after creation
			require.Equal(t, "", actualToken.Key)

			testCtx.DeleteAuthToken(newUser.ID, actualToken.ID)
		}

		actualTokens = testCtx.ListUserTokens(newUser.ID)
		require.Equal(t, 0, len(actualTokens), "Expected no remaining auth tokens for user %s but found %d", newUser.ID.String(), len(actualTokens))
	})
}

func Test_NonAdminFunctionality(t *testing.T) {
	var (
		testCtx        = integration.NewFOSSContext(t)
		newUser        = testCtx.CreateUser(otherUser, otherUser, auth.RoleReadOnly)
		nonAdminUser   = testCtx.CreateUser(nonAdmin, nonAdmin, auth.RoleUser)
		nonAdminToken  = testCtx.CreateAuthToken(nonAdminUser.ID, "NonAdmin Token")
		nonAdminClient = testCtx.NewAPIClientWithToken(nonAdminToken)
	)

	t.Run("Creating and listing tokens as a nonadmin is successful", func(t *testing.T) {
		newToken, err := nonAdminClient.CreateUserToken(nonAdminUser.ID, "NonAdmin Token 2")
		require.Nil(t, err, "Expected a successful response, got: %v", err)

		nonAdminActualTokens, err := nonAdminClient.ListAuthTokens()
		if err != nil {
			t.Logf("Encountered error checking auth tokens: %s", err)
			t.Fail()
		}

		require.Equalf(t, 2, len(nonAdminActualTokens.Tokens), "Expected 2 tokens but found %d", len(nonAdminActualTokens.Tokens))
		require.Equalf(t, nonAdminActualTokens.Tokens[0].UserID.UUID, nonAdminUser.ID, "Non-admins should not return tokens for other users: %v", nonAdminActualTokens)
		require.Equalf(t, nonAdminActualTokens.Tokens[1].UserID.UUID, nonAdminUser.ID, "Non-admins should not return tokens for other users: %v", nonAdminActualTokens)

		for _, actualToken := range nonAdminActualTokens.Tokens {
			// All tokens must strip their key after creation
			require.Equal(t, "", actualToken.Key)

			// Cleanup the token we created so it doesnt affect other tests
			if actualToken.ID == newToken.ID {
				testCtx.DeleteAuthToken(nonAdminToken.ID, actualToken.ID)
			}
		}

		actualTokens := testCtx.ListUserTokens(nonAdminUser.ID)
		require.Equal(t, 1, len(actualTokens), "Expected no new auth tokens for user %s but found %d", nonAdminUser.ID.String(), len(actualTokens))
	})

	t.Run("Creating tokens for other user as nonadmin should fail", func(t *testing.T) {
		_, err := nonAdminClient.CreateUserToken(newUser.ID, "NewUser Token")
		wrapper, ok := err.(api.ErrorWrapper)
		if !ok {
			t.Logf("Expected a valid API Error response. Failed unmarshalling to api.ErrorWrapper: %v", err)
			t.Fail()
		}
		require.Equal(t, http.StatusForbidden, wrapper.HTTPStatus)

		nonAdminActualTokens, err := nonAdminClient.ListAuthTokens()
		if err != nil {
			t.Logf("Encountered error checking auth tokens: %s", err)
			t.Fail()
		}

		require.Equalf(t, 1, len(nonAdminActualTokens.Tokens), "Expected 1 tokens but found %d", len(nonAdminActualTokens.Tokens))
		require.Equalf(t, nonAdminActualTokens.Tokens[0].UserID.UUID, nonAdminUser.ID, "Non-admins should not return tokens for other users: %v", nonAdminActualTokens.Tokens)
	})

	t.Run("Deleting tokens as nonadmin is successful", func(t *testing.T) {
		tokenToDelete, err := nonAdminClient.CreateUserToken(nonAdminUser.ID, "nonAdminUser token to delete")
		if err != nil {
			t.Logf("Encountered error creating auth token: %s", err)
			t.Fail()
		}

		err = nonAdminClient.DeleteUserToken(nonAdminUser.ID, tokenToDelete.ID)
		require.Nilf(t, err, "Received unexpected error when deleting token: %s", err)

		// Ensure the token gets cleaned up if there was an error above
		if err != nil {
			// Use the admin session to cleanup the newly created token
			testCtx.DeleteAuthToken(nonAdminUser.ID, tokenToDelete.ID)
		}
	})

	t.Run("Deleting tokens for other user as nonadmin should fail", func(t *testing.T) {
		tokenToDelete := testCtx.CreateAuthToken(newUser.ID, "newUser token to delete")

		err := nonAdminClient.DeleteUserToken(newUser.ID, tokenToDelete.ID)
		require.Equalf(t, errors.Error("API returned a 404 error"), err, "Expected to receive a 404 error when attempting to delete a token, got: %s", err)

		// Use the admin session to cleanup the newly created token
		testCtx.DeleteAuthToken(newUser.ID, tokenToDelete.ID)
	})
}
