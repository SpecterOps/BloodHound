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

//go:build integration
// +build integration

package database_test

import (
	"context"
	"github.com/specterops/bloodhound/src/utils/test"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
)

const (
	userPrincipal  = "first.last@example.com"
	user2Principal = "first2.last2@example.com"
	roleToDelete   = auth.RoleReadOnly
)

func initAndGetRoles(t *testing.T) (database.Database, model.Roles) {
	dbInst := integration.OpenDatabase(t)
	if err := integration.Prepare(dbInst); err != nil {
		t.Fatalf("Failed preparing DB: %v", err)
	}

	if roles, err := dbInst.GetAllRoles(context.Background(), "", model.SQLFilter{}); err != nil {
		t.Fatalf("Error fetching roles: %v", err)
	} else {
		return dbInst, roles
	}

	return nil, nil
}

func initAndCreateUser(t *testing.T) (database.Database, model.User) {
	var (
		dbInst, roles = initAndGetRoles(t)
		user          = model.User{
			Roles:         roles,
			FirstName:     null.StringFrom("First"),
			LastName:      null.StringFrom("Last"),
			EmailAddress:  null.StringFrom(userPrincipal),
			PrincipalName: userPrincipal,
		}
	)

	if newUser, err := dbInst.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("Error creating user: %v", err)
	} else {
		return dbInst, newUser
	}

	return nil, model.User{}
}

func TestDatabase_Installation(t *testing.T) {
	dbInst := integration.OpenDatabase(t)
	if err := integration.Prepare(dbInst); err != nil {
		t.Fatalf("Failed preparing DB: %v", err)
	}

	if installation, err := dbInst.CreateInstallation(); err != nil {
		t.Fatalf("Error creating installation: %v", err)
	} else if fetchedInstallation, err := dbInst.GetInstallation(); err != nil {
		t.Fatalf("Failed to fetch installation: %v", err)
	} else if installation.ID.String() != fetchedInstallation.ID.String() {
		t.Fatalf("Installation fetched does not match the initially created installation")
	}
}

func TestDatabase_InitializePermissions(t *testing.T) {
	dbInst := integration.OpenDatabase(t)
	if err := integration.Prepare(dbInst); err != nil {
		t.Fatalf("Failed preparing DB: %v", err)
	}

	if permissions, err := dbInst.GetAllPermissions(context.Background(), "", model.SQLFilter{}); err != nil {
		t.Fatalf("Error fetching permissions: %v", err)
	} else {
		templates := auth.Permissions().All()

		for _, permissionTemplate := range templates {
			found := false

			for _, permission := range permissions {
				if permission.Equals(permissionTemplate) {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("Missing permission %s", permissionTemplate)
			}
		}
	}
}

func TestDatabase_InitializeRoles(t *testing.T) {
	var (
		_, roles  = initAndGetRoles(t)
		templates = auth.Roles()
	)

	for _, roleTemplate := range templates {
		found := false

		for _, role := range roles {
			if role.Name == roleTemplate.Name {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Missing role %s", roleTemplate.Name)
		}
	}
}

func TestDatabase_CreateGetDeleteUser(t *testing.T) {
	var (
		ctx           = context.Background()
		dbInst, roles = initAndGetRoles(t)

		users = model.Users{

			{
				Roles:         roles,
				FirstName:     null.StringFrom("First"),
				LastName:      null.StringFrom("Last"),
				EmailAddress:  null.StringFrom(userPrincipal),
				PrincipalName: userPrincipal,
			},
			{
				Roles:         roles,
				FirstName:     null.StringFrom("First2"),
				LastName:      null.StringFrom("Last2"),
				EmailAddress:  null.StringFrom(user2Principal),
				PrincipalName: user2Principal,
			},
		}

		createdUsers []model.User
	)

	for _, user := range users {
		if _, err := dbInst.CreateUser(ctx, user); err != nil {
			t.Fatalf("Error creating user: %v", err)
		} else if newUser, err := dbInst.LookupUser(ctx, user.PrincipalName); err != nil {
			t.Fatalf("Failed looking up user by principal %s: %v", user.PrincipalName, err)
		} else if err = test.VerifyAuditLogs(dbInst, "CreateUser", "principal_name", newUser.PrincipalName); err != nil {
			t.Fatalf("Failed to validate CreateUser audit logs:\n%v", err)
		} else {
			for _, role := range roles {
				found := false

				for _, userRole := range newUser.Roles {
					if userRole.Name == role.Name {
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("Missing role %s", role.Name)
				}
			}
			createdUsers = append(createdUsers, newUser)

			newUser.Roles = newUser.Roles.RemoveByName(roleToDelete)

			if err := dbInst.UpdateUser(ctx, newUser); err != nil {
				t.Fatalf("Failed to update user: %v", err)
			} else if err = test.VerifyAuditLogs(dbInst, "UpdateUser", "principal_name", newUser.PrincipalName); err != nil {
				t.Fatalf("Failed to validate UpdateUser audit logs:\n%v", err)
			} else if updatedUser, err := dbInst.LookupUser(ctx, user.PrincipalName); err != nil {
				t.Fatalf("Failed looking up user by principal %s: %v", user.PrincipalName, err)
			} else if _, found := updatedUser.Roles.FindByName(roleToDelete); found {
				t.Fatalf("Found role %s on user %s but expected it to be removed", roleToDelete, user.PrincipalName)
			}
		}
	}

	if err := dbInst.DeleteUser(ctx, createdUsers[1]); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, "DeleteUser", "principal_name", users[1].PrincipalName); err != nil {
		t.Fatalf("Failed to validate Deleteuser audit logs:\n%v", err)
	}

	if usersResponse, err := dbInst.GetAllUsers(ctx, "first_name", model.SQLFilter{}); err != nil {
		t.Fatalf("Error getting users: %v", err)
	} else if usersResponse[0].FirstName.String != "First" {
		t.Fatalf("ListUsers returned incorrectly sorted data")
	} else if len(usersResponse) > 1 {
		t.Fatalf("User '%s' exists but should have been deleted. Response: %v", createdUsers[1].PrincipalName, usersResponse)
	}
}

func TestDatabase_CreateGetDeleteAuthToken(t *testing.T) {
	var (
		ctx          = context.Background()
		dbInst, user = initAndCreateUser(t)
		expectedName = "test"
		token        = model.AuthToken{
			UserID:     database.NullUUID(user.ID),
			Key:        "key",
			HmacMethod: "fake",
			Name:       null.StringFrom(expectedName),
		}
	)

	if newToken, err := dbInst.CreateAuthToken(ctx, token); err != nil {
		t.Fatalf("Failed to create auth token: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, "CreateAuthToken", "id", newToken.ID.String()); err != nil {
		t.Fatalf("Failed to validate CreateAuthToken audit logs:\n%v", err)
	} else if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if len(updatedUser.AuthTokens) != 1 {
		t.Fatalf("Expected 1 auth token for user %s but saw only %d", userPrincipal, len(updatedUser.AuthTokens))
	} else if !newToken.Name.Valid {
		t.Fatalf("Expected auth token to have valid name")
	} else if newToken.Name.String != expectedName {
		t.Fatalf("Expected auth token to have name %s but saw %v", expectedName, newToken.Name.String)
	} else if err = dbInst.DeleteAuthToken(ctx, newToken); err != nil {
		t.Fatalf("Failed to delete auth token: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, "DeleteAuthToken", "id", newToken.ID.String()); err != nil {
		t.Fatalf("Failed to validate DeleteAuthToken audit logs:\n%v", err)
	}

	if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if len(updatedUser.AuthTokens) != 0 {
		t.Fatalf("Expected 0 auth tokens for user %s but saw %d", userPrincipal, len(updatedUser.AuthTokens))
	}
}

func TestDatabase_CreateGetDeleteAuthSecret(t *testing.T) {
	const updatedDigest = "updated"

	var (
		ctx          = context.Background()
		dbInst, user = initAndCreateUser(t)
		secret       = model.AuthSecret{
			UserID:       user.ID,
			Digest:       "digest",
			DigestMethod: "fake",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}
	)

	if newSecret, err := dbInst.CreateAuthSecret(ctx, secret); err != nil {
		t.Fatalf("Failed to create auth secret: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, "CreateAuthSecret", "secret_user_id", newSecret.UserID.String()); err != nil {
		t.Fatalf("Failed to validate CreateAuthSecret audit logs:\n%v", err)
	} else if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if updatedUser.AuthSecret.ID != newSecret.ID {
		t.Fatalf("Expected auth secret for user %s to be %d but saw %d", userPrincipal, newSecret.ID, updatedUser.AuthSecret.ID)
	} else {
		newSecret.Digest = updatedDigest

		if err := dbInst.UpdateAuthSecret(ctx, newSecret); err != nil {
			t.Fatalf("Failed to update auth secret %d: %v", newSecret.ID, err)
		} else if err = test.VerifyAuditLogs(dbInst, "UpdateAuthSecret", "secret_user_id", newSecret.UserID.String()); err != nil {
			t.Fatalf("Failed to validate UpdateAuthSecret audit logs:\n%v", err)
		} else if updatedSecret, err := dbInst.GetAuthSecret(ctx, newSecret.ID); err != nil {
			t.Fatalf("Failed to fetch updated auth secret: %v", err)
		} else if updatedSecret.Digest != updatedDigest {
			t.Fatalf("Expected updated auth secret digest to be %s but saw %s", updatedDigest, updatedSecret.Digest)
		}

		if err := dbInst.DeleteAuthSecret(ctx, newSecret); err != nil {
			t.Fatalf("Failed to delete auth token: %v", err)
		} else if err = test.VerifyAuditLogs(dbInst, "DeleteAuthSecret", "secret_user_id", newSecret.UserID.String()); err != nil {
			t.Fatalf("Failed to validate DeleteAuthSecret audit logs:\n%v", err)
		}
	}

	if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if updatedUser.AuthSecret != nil {
		t.Fatalf("Expected user %s to have no auth secret set", userPrincipal)
	}
}

func TestDatabase_CreateUpdateDeleteSAMLProvider(t *testing.T) {
	var (
		ctx          = context.Background()
		dbInst, user = initAndCreateUser(t)

		samlProvider = model.SAMLProvider{
			Name:            "provider",
			DisplayName:     "provider name",
			IssuerURI:       "https://idp.example.com/idp.xml",
			SingleSignOnURI: "https://idp.example.com/sso",
		}

		newSAMLProvider model.SAMLProvider
		err             error
	)

	if newSAMLProvider, err = dbInst.CreateSAMLIdentityProvider(ctx, samlProvider); err != nil {
		t.Fatalf("Failed to create SAML provider: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, "CreateSAMLIdentityProvider", "saml_name", newSAMLProvider.Name); err != nil {
		t.Fatalf("Failed to validate CreateSAMLIdentityProvider audit logs:\n%v", err)
	} else {
		user.SAMLProviderID = null.Int32From(newSAMLProvider.ID)

		if err := dbInst.UpdateUser(ctx, user); err != nil {
			t.Fatalf("Failed to update user: %v", err)
		} else if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
			t.Fatalf("Failed to fetch updated user: %v", err)
		} else if updatedUser.SAMLProvider == nil {
			t.Fatalf("Updated user does not have a SAMLProvider set when it should")
		} else if updatedUser.SAMLProvider.ID != newSAMLProvider.ID {
			t.Fatalf("Updated user has SAMLProvider ID %d when %d was expected", updatedUser.SAMLProvider.ID, newSAMLProvider.ID)
		} else if updatedUser.SAMLProvider.IssuerURI != newSAMLProvider.IssuerURI {
			t.Fatalf("Updated user has SAMLProvider URL %s when %s was expected", updatedUser.SAMLProvider.IssuerURI, newSAMLProvider.IssuerURI)
		}
	}

	updatedSAMLProvider := model.SAMLProvider{
		Serial: model.Serial{
			ID: newSAMLProvider.ID,
		},
		Name:            "updated provider",
		DisplayName:     newSAMLProvider.DisplayName,
		IssuerURI:       newSAMLProvider.IssuerURI,
		SingleSignOnURI: newSAMLProvider.SingleSignOnURI,
	}
	if err := dbInst.UpdateSAMLIdentityProvider(ctx, updatedSAMLProvider); err != nil {
		t.Fatalf("Failed to update SAML provider: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, "UpdateSAMLIdentityProvider", "saml_name", "updated provider"); err != nil {
		t.Fatalf("Failed to validate UpdateSAMLIdentityProvider audit logs:\n%v", err)
	}

	user.SAMLProviderID = null.Int32{}
	if err := dbInst.UpdateUser(ctx, user); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	} else if err := dbInst.DeleteSAMLProvider(ctx, newSAMLProvider); err != nil {
		t.Fatalf("Failed to delete SAML provider: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, "DeleteSAMLIdentityProvider", "saml_name", "provider"); err != nil {
		t.Fatalf("Failed to validate DeleteSAMLIdentityProvider audit logs:\n%v", err)
	}
}

func TestDatabase_CreateUserSession(t *testing.T) {
	var (
		dbInst, user = initAndCreateUser(t)
		userSession  = model.UserSession{
			User:      user,
			UserID:    user.ID,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
	)

	if newUserSession, err := dbInst.CreateUserSession(userSession); err != nil {
		t.Fatalf("Failed to create new user session: %v", err)
	} else if newUserSession.Expired() {
		t.Fatalf("Expected user session to remain valid. Session expires at: %s", newUserSession.ExpiresAt)
	} else {
		assert.Equal(t, user, newUserSession.User)
	}

	// Test expiry
	userSession.ExpiresAt = time.Now().UTC().Add(-time.Hour)

	if newUserSession, err := dbInst.CreateUserSession(userSession); err != nil {
		t.Fatalf("Failed to create new user session: %v", err)
	} else if !newUserSession.Expired() {
		t.Fatalf("Expected user session to be expired. Session expires at: %s", newUserSession.ExpiresAt)
	} else {
		assert.Equal(t, user, newUserSession.User)
	}
}
