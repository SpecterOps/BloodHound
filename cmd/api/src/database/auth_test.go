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

	if roles, err := dbInst.GetAllRoles("", model.SQLFilter{}); err != nil {
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

	if permissions, err := dbInst.GetAllPermissions("", model.SQLFilter{}); err != nil {
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

func TestDatabase_UpdateRole(t *testing.T) {
	dbInst, roles := initAndGetRoles(t)

	if role, found := roles.FindByName(auth.RoleReadOnly); !found {
		t.Fatal("Unable to find role")
	} else if allPermissions, err := dbInst.GetAllPermissions("", model.SQLFilter{}); err != nil {
		t.Fatalf("Failed fetching all permissions: %v", err)
	} else {
		role.Permissions = allPermissions

		if err := dbInst.UpdateRole(role); err != nil {
			t.Fatalf("Failed updating role %s: %v", role.Name, err)
		}

		if updatedRole, err := dbInst.GetRole(role.ID); err != nil {
			t.Fatalf("Failed fetching updated role %s: %v", role.Name, err)
		} else {
			for _, permission := range role.Permissions {
				found := false

				for _, updatedPermission := range updatedRole.Permissions {
					if permission.Equals(updatedPermission) {
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("Updated role %s missing expected permission %s", role.Name, permission)
				}
			}
		}
	}
}

func TestDatabase_CreateGetUser(t *testing.T) {
	var (
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
	)

	for _, user := range users {
		if _, err := dbInst.CreateUser(context.Background(), user); err != nil {
			t.Fatalf("Error creating user: %v", err)
		} else if newUser, err := dbInst.LookupUser(user.PrincipalName); err != nil {
			t.Fatalf("Failed looking up user by principal %s: %v", user.PrincipalName, err)
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

			newUser.Roles = newUser.Roles.RemoveByName(roleToDelete)

			if err := dbInst.UpdateUser(context.Background(), newUser); err != nil {
				t.Fatalf("Failed to update user: %v", err)
			}

			if updatedUser, err := dbInst.LookupUser(user.PrincipalName); err != nil {
				t.Fatalf("Failed looking up user by principal %s: %v", user.PrincipalName, err)
			} else if _, found := updatedUser.Roles.FindByName(roleToDelete); found {
				t.Fatalf("Found role %s on user %s but expected it to be removed", roleToDelete, user.PrincipalName)
			}
		}
	}

	if usersResponse, err := dbInst.GetAllUsers("first_name", model.SQLFilter{}); err != nil {
		t.Fatalf("Error getting users: %v", err)
	} else if usersResponse[0].FirstName.String != "First" {
		t.Fatalf("ListUsers returned incorrectly sorted data")
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
	} else if updatedUser, err := dbInst.GetUser(user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if len(updatedUser.AuthTokens) != 1 {
		t.Fatalf("Expected 1 auth token for user %s but saw only %d", userPrincipal, len(updatedUser.AuthTokens))
	} else if !newToken.Name.Valid {
		t.Fatalf("Expected auth token to have valid name")
	} else if newToken.Name.String != expectedName {
		t.Fatalf("Expected auth token to have name %s but saw %v", expectedName, newToken.Name.String)
	} else if err := dbInst.DeleteAuthToken(ctx, newToken); err != nil {
		t.Fatalf("Failed to delete auth token: %v", err)
	}

	if updatedUser, err := dbInst.GetUser(user.ID); err != nil {
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
	} else if updatedUser, err := dbInst.GetUser(user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if updatedUser.AuthSecret.ID != newSecret.ID {
		t.Fatalf("Expected auth secret for user %s to be %d but saw %d", userPrincipal, newSecret.ID, updatedUser.AuthSecret.ID)
	} else {
		newSecret.Digest = updatedDigest

		if err := dbInst.UpdateAuthSecret(ctx, newSecret); err != nil {
			t.Fatalf("Failed to update auth secret %d: %v", newSecret.ID, err)
		} else if updatedSecret, err := dbInst.GetAuthSecret(newSecret.ID); err != nil {
			t.Fatalf("Failed to fetch updated auth secret: %v", err)
		} else if updatedSecret.Digest != updatedDigest {
			t.Fatalf("Expected updated auth secret digest to be %s but saw %s", updatedDigest, updatedSecret.Digest)
		}

		if err := dbInst.DeleteAuthSecret(ctx, newSecret); err != nil {
			t.Fatalf("Failed to delete auth token: %v", err)
		}
	}

	if updatedUser, err := dbInst.GetUser(user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if updatedUser.AuthSecret != nil {
		t.Fatalf("Expected user %s to have no auth secret set", userPrincipal)
	}
}

func TestDatabase_CreateSAMLProvider(t *testing.T) {
	dbInst, user := initAndCreateUser(t)

	samlProvider := model.SAMLProvider{
		Name:            "provider",
		DisplayName:     "provider name",
		IssuerURI:       "https://idp.example.com/idp.xml",
		SingleSignOnURI: "https://idp.example.com/sso",
	}

	if newSAMLProvider, err := dbInst.CreateSAMLIdentityProvider(context.Background(), samlProvider); err != nil {
		t.Fatalf("Failed to create SAML provider: %v", err)
	} else {
		user.SAMLProviderID = null.Int32From(newSAMLProvider.ID)

		if err := dbInst.UpdateUser(context.Background(), user); err != nil {
			t.Fatalf("Failed to update user: %v", err)
		} else if updatedUser, err := dbInst.GetUser(user.ID); err != nil {
			t.Fatalf("Failed to fetch updated user: %v", err)
		} else if updatedUser.SAMLProvider == nil {
			t.Fatalf("Updated user does not have a SAMLProvider set when it should")
		} else if updatedUser.SAMLProvider.ID != newSAMLProvider.ID {
			t.Fatalf("Updated user has SAMLProvider ID %d when %d was expected", updatedUser.SAMLProvider.ID, newSAMLProvider.ID)
		} else if updatedUser.SAMLProvider.IssuerURI != newSAMLProvider.IssuerURI {
			t.Fatalf("Updated user has SAMLProvider URL %s when %s was expected", updatedUser.SAMLProvider.IssuerURI, newSAMLProvider.IssuerURI)
		}
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
