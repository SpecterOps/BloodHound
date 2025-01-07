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
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	userPrincipal  = "first.last@example.com"
	user2Principal = "first2.last2@example.com"
	user3Principal = "first3.last3@example.com"
	user4Principal = "first4.last4@example.com"

	roleToDelete = auth.RoleReadOnly
)

func initAndGetRoles(t *testing.T) (database.Database, model.Roles) {
	dbInst := integration.SetupDB(t)

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
	var (
		dbInst  = integration.SetupDB(t)
		testCtx = context.Background()
	)
	if installation, err := dbInst.CreateInstallation(testCtx); err != nil {
		t.Fatalf("Error creating installation: %v", err)
	} else if fetchedInstallation, err := dbInst.GetInstallation(testCtx); err != nil {
		t.Fatalf("Failed to fetch installation: %v", err)
	} else if installation.ID.String() != fetchedInstallation.ID.String() {
		t.Fatalf("Installation fetched does not match the initially created installation")
	}
}

func TestDatabase_InitializePermissions(t *testing.T) {
	dbInst := integration.SetupDB(t)

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
		} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateUser, "principal_name", newUser.PrincipalName); err != nil {
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
			} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionUpdateUser, "principal_name", newUser.PrincipalName); err != nil {
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
	} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionDeleteUser, "principal_name", users[1].PrincipalName); err != nil {
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

func TestDatabase_UpdateUserAuth(t *testing.T) {
	var (
		ctx          = context.Background()
		dbInst, user = initAndCreateUser(t)
		secret       = model.AuthSecret{
			UserID:       user.ID,
			Digest:       "digest",
			DigestMethod: "fake",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}
		samlProvider = model.SAMLProvider{
			Name:            "provider",
			DisplayName:     "provider name",
			IssuerURI:       "https://idp.example.com/idp.xml",
			SingleSignOnURI: "https://idp.example.com/sso",
		}
	)

	if newSecret, err := dbInst.CreateAuthSecret(ctx, secret); err != nil {
		t.Fatalf("Failed to create auth secret: %v", err)
	} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateAuthSecret, "secret_user_id", newSecret.UserID.String()); err != nil {
		t.Fatalf("Failed to validate CreateAuthSecret audit logs:\n%v", err)
	} else {
		if newSAMLProvider, err := dbInst.CreateSAMLIdentityProvider(ctx, samlProvider, model.SSOProviderConfig{}); err != nil {
			t.Fatalf("Failed to create SAML provider: %v", err)
		} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateSAMLIdentityProvider, "saml_name", newSAMLProvider.Name); err != nil {
			t.Fatalf("Failed to validate CreateSAMLIdentityProvider audit logs:\n%v", err)
		} else {
			user, err = dbInst.GetUser(ctx, user.ID)
			if err != nil {
				t.Fatalf("Failed looking up user by principal %s: %v", user.PrincipalName, err)
			}

			user.FirstName = null.StringFrom("friendly man")

			if err := dbInst.UpdateUser(ctx, user); err != nil {
				t.Fatalf("Failed to update user: %v", err)
			} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionUpdateUser, "principal_name", user.PrincipalName); err != nil {
				t.Fatalf("Failed to validate UpdateUser audit logs:\n%v", err)
			} else if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
				t.Fatalf("Failed looking up user by principal %s: %v", user.PrincipalName, err)
			} else if updatedUser.AuthSecret == nil {
				t.Fatalf("Failed to find authsecret for user %s", user.PrincipalName)
			} else if _, err := dbInst.GetAuthSecret(ctx, updatedUser.AuthSecret.ID); err != nil {
				t.Fatalf("Failed to get authsecret by id %d", updatedUser.AuthSecret.ID)
			}

			user.AuthSecret = nil
			user.SSOProviderID = newSAMLProvider.SSOProviderID

			if err := dbInst.UpdateUser(ctx, user); err != nil {
				t.Fatalf("Failed to update user: %v", err)
			} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionUpdateUser, "principal_name", user.PrincipalName); err != nil {
				t.Fatalf("Failed to validate UpdateUser audit logs:\n%v", err)
			} else if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
				t.Fatalf("Failed looking up user by principal %s: %v", user.PrincipalName, err)
			} else if updatedUser.AuthSecret != nil {
				t.Fatalf("Found authsecret for user %s but expected it to be removed", user.PrincipalName)
			} else if _, err := dbInst.GetAuthSecret(ctx, newSecret.ID); err == nil {
				t.Fatalf("Found authsecret for id %d but expected it to be removed", newSecret.ID)
			}
		}
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
	} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateAuthToken, "id", newToken.ID.String()); err != nil {
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
	} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateAuthSecret, "secret_user_id", newSecret.UserID.String()); err != nil {
		t.Fatalf("Failed to validate CreateAuthSecret audit logs:\n%v", err)
	} else if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if updatedUser.AuthSecret.ID != newSecret.ID {
		t.Fatalf("Expected auth secret for user %s to be %d but saw %d", userPrincipal, newSecret.ID, updatedUser.AuthSecret.ID)
	} else {
		newSecret.Digest = updatedDigest

		if err := dbInst.UpdateAuthSecret(ctx, newSecret); err != nil {
			t.Fatalf("Failed to update auth secret %d: %v", newSecret.ID, err)
		} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionUpdateAuthSecret, "secret_user_id", newSecret.UserID.String()); err != nil {
			t.Fatalf("Failed to validate UpdateAuthSecret audit logs:\n%v", err)
		} else if updatedSecret, err := dbInst.GetAuthSecret(ctx, newSecret.ID); err != nil {
			t.Fatalf("Failed to fetch updated auth secret: %v", err)
		} else if updatedSecret.Digest != updatedDigest {
			t.Fatalf("Expected updated auth secret digest to be %s but saw %s", updatedDigest, updatedSecret.Digest)
		}

		if err := dbInst.DeleteAuthSecret(ctx, newSecret); err != nil {
			t.Fatalf("Failed to delete auth token: %v", err)
		} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionDeleteAuthSecret, "secret_user_id", newSecret.UserID.String()); err != nil {
			t.Fatalf("Failed to validate DeleteAuthSecret audit logs:\n%v", err)
		}
	}

	if updatedUser, err := dbInst.GetUser(ctx, user.ID); err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	} else if updatedUser.AuthSecret != nil {
		t.Fatalf("Expected user %s to have no auth secret set", userPrincipal)
	}
}

func TestDatabase_CreateUpdateDeleteSSOProvider(t *testing.T) {
	t.Run("successfully CreateUpdateDeleteSAMLProvider", func(t *testing.T) {
		var (
			ctx             = context.Background()
			dbInst, user    = initAndCreateUser(t)
			samlProvider    model.SAMLProvider
			newSAMLProvider model.SAMLProvider
			updatedUser     model.User
			config          = model.SSOProviderConfig{}
			err             error
		)
		// Initialize the SAMLProvider without setting SSOProviderID
		samlProvider = model.SAMLProvider{
			Name:            "provider",
			DisplayName:     "provider name",
			IssuerURI:       "https://idp.example.com/idp.xml",
			SingleSignOnURI: "https://idp.example.com/sso",
		}

		if newSAMLProvider, err = dbInst.CreateSAMLIdentityProvider(ctx, samlProvider, config); err != nil {
			t.Fatalf("Failed to create SAML provider: %v", err)
		} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateSAMLIdentityProvider, "saml_name", newSAMLProvider.Name); err != nil {
			t.Fatalf("Failed to validate CreateSAMLIdentityProvider audit logs:\n%v", err)
		} else {
			user.SSOProviderID = newSAMLProvider.SSOProviderID
			if err = dbInst.UpdateUser(ctx, user); err != nil {
				t.Fatalf("Failed to update user: %v", err)
			} else if updatedUser, err = dbInst.GetUser(ctx, user.ID); err != nil {
				t.Fatalf("Failed to fetch updated user: %v", err)
			} else if updatedUser.SSOProvider == nil {
				t.Fatalf("Updated user does not have a SAMLProvider set when it should")
			} else if updatedUser.SSOProvider.ID != newSAMLProvider.SSOProviderID.Int32 {
				t.Fatalf("Updated user has SSOProvider ID %d when %v was expected", updatedUser.SSOProvider.ID, newSAMLProvider.SSOProviderID)
			} else if updatedUser.SSOProvider.SAMLProvider.IssuerURI != newSAMLProvider.IssuerURI {
				t.Fatalf("Updated user has SAMLProvider URL %s when %s was expected", updatedUser.SSOProvider.SAMLProvider.IssuerURI, newSAMLProvider.IssuerURI)
			} else {
				updatedSSOProvider := model.SSOProvider{
					Name: "updated provider",
					Type: model.SessionAuthProviderSAML,
					SAMLProvider: &model.SAMLProvider{
						Serial: model.Serial{
							ID: newSAMLProvider.ID,
						},
						Name:            "updated provider",
						DisplayName:     newSAMLProvider.DisplayName,
						IssuerURI:       newSAMLProvider.IssuerURI,
						SingleSignOnURI: newSAMLProvider.SingleSignOnURI,
						SSOProviderID:   newSAMLProvider.SSOProviderID,
					},
					Config: config,
				}

				if _, err = dbInst.UpdateSAMLIdentityProvider(ctx, updatedSSOProvider); err != nil {
					t.Fatalf("Failed to update SAML provider: %v", err)
				} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionUpdateSAMLIdentityProvider, "saml_name", "updated provider"); err != nil {
					t.Fatalf("Failed to validate UpdateSAMLIdentityProvider audit logs:\n%v", err)
				} else {
					user.SSOProviderID = null.Int32{}
					if err = dbInst.UpdateUser(ctx, user); err != nil {
						t.Fatalf("Failed to update user: %v", err)
					} else if err = dbInst.DeleteSSOProvider(ctx, int(newSAMLProvider.SSOProviderID.Int32)); err != nil {
						t.Fatalf("Failed to delete SAML provider: %v", err)
					} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionDeleteSSOIdentityProvider, "name", "provider"); err != nil {
						t.Fatalf("Failed to validate DeleteSAMLIdentityProvider audit logs:\n%v", err)
					}
				}
			}
		}
	})

	t.Run("successfully CreateUpdateDeleteOIDCProvider", func(t *testing.T) {
		var (
			testCtx      = context.Background()
			dbInst, user = initAndCreateUser(t)
			oidcProvider = model.OIDCProvider{
				ClientID: "bloodhound",
				Issuer:   "https://localhost/auth",
			}
			updatedUser model.User
			emptyConfig = model.SSOProviderConfig{}
			config      = model.SSOProviderConfig{
				AutoProvision: model.SSOProviderAutoProvisionConfig{
					Enabled:       true,
					DefaultRoleId: 3,
					RoleProvision: true,
				},
			}
		)

		if newOIDCProvider, err := dbInst.CreateOIDCProvider(testCtx, "test_oidc", oidcProvider.Issuer, oidcProvider.ClientID, emptyConfig); err != nil {
			t.Fatalf("Failed to create OIDC provider: %v", err)
		} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionCreateOIDCIdentityProvider, "client_id", "bloodhound"); err != nil {
			t.Fatalf("Failed to validate CreateOIDCIdentityProvider audit logs:\n%v", err)
		} else {
			user.SSOProviderID = null.Int32From(int32(newOIDCProvider.SSOProviderID))
			if err = dbInst.UpdateUser(testCtx, user); err != nil {
				t.Fatalf("Failed to update user: %v", err)
			} else if updatedUser, err = dbInst.GetUser(testCtx, user.ID); err != nil {
				t.Fatalf("Failed to fetch updated user: %v", err)
			} else if updatedUser.SSOProvider == nil {
				t.Fatalf("Updated user does not have a OIDCProvider set when it should")
			} else if updatedUser.SSOProvider.ID != int32(newOIDCProvider.SSOProviderID) {
				t.Fatalf("Updated user has SSOProvider ID %d when %v was expected", updatedUser.SSOProvider.ID, newOIDCProvider.ID)
			} else if updatedUser.SSOProvider.OIDCProvider.Issuer != newOIDCProvider.Issuer {
				t.Fatalf("Updated user has OIDCProvider Issuer %s when %s was expected", updatedUser.SSOProvider.OIDCProvider.Issuer, newOIDCProvider.Issuer)
			} else if updatedUser.SSOProvider.Config != emptyConfig {
				t.Fatalf("Updated user has Config %v when %v was expected", updatedUser.SSOProvider.Config, emptyConfig)
			} else {
				updatedSSOProvider := model.SSOProvider{
					Name: "updated provider",
					Type: model.SessionAuthProviderOIDC,
					OIDCProvider: &model.OIDCProvider{
						Serial: model.Serial{
							ID: newOIDCProvider.ID,
						},
						Issuer:        newOIDCProvider.Issuer,
						ClientID:      newOIDCProvider.ClientID,
						SSOProviderID: newOIDCProvider.SSOProviderID,
					},
					Config: config,
				}

				if _, err = dbInst.UpdateOIDCProvider(testCtx, updatedSSOProvider); err != nil {
					t.Fatalf("Failed to update OIDC provider: %v", err)
				} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionUpdateOIDCIdentityProvider, "client_id", "bloodhound"); err != nil {
					t.Fatalf("Failed to validate UpdateOIDCIdentityProvider audit logs:\n%v", err)
				} else {
					user.SSOProviderID = null.Int32{}
					if err = dbInst.UpdateUser(testCtx, user); err != nil {
						t.Fatalf("Failed to update user: %v", err)
					} else if err = dbInst.DeleteSSOProvider(testCtx, int(newOIDCProvider.SSOProviderID)); err != nil {
						t.Fatalf("Failed to delete OIDC provider: %v", err)
					} else if err = test.VerifyAuditLogs(dbInst, model.AuditLogActionDeleteSSOIdentityProvider, "name", "test_oidc"); err != nil {
						t.Fatalf("Failed to validate DeleteSSOIdentityProvider audit logs:\n%v", err)
					}
				}
			}
		}
	})
}

func TestDatabase_CreateUserSession(t *testing.T) {
	var (
		testCtx      = context.Background()
		dbInst, user = initAndCreateUser(t)
		userSession  = model.UserSession{
			User:      user,
			UserID:    user.ID,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
	)

	if newUserSession, err := dbInst.CreateUserSession(testCtx, userSession); err != nil {
		t.Fatalf("Failed to create new user session: %v", err)
	} else if newUserSession.Expired() {
		t.Fatalf("Expected user session to remain valid. Session expires at: %s", newUserSession.ExpiresAt)
	} else {
		assert.Equal(t, user, newUserSession.User)
	}

	// Test expiry
	userSession.ExpiresAt = time.Now().UTC().Add(-time.Hour)

	if newUserSession, err := dbInst.CreateUserSession(testCtx, userSession); err != nil {
		t.Fatalf("Failed to create new user session: %v", err)
	} else if !newUserSession.Expired() {
		t.Fatalf("Expected user session to be expired. Session expires at: %s", newUserSession.ExpiresAt)
	} else {
		assert.Equal(t, user, newUserSession.User)
	}
}

func TestDatabase_SetUserSessionFlag(t *testing.T) {
	var (
		testCtx      = context.Background()
		dbInst, user = initAndCreateUser(t)
		userSession  = model.UserSession{
			User:      user,
			UserID:    user.ID,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
	)

	newUserSession, err := dbInst.CreateUserSession(testCtx, userSession)
	assert.Nil(t, err)

	err = dbInst.SetUserSessionFlag(testCtx, &newUserSession, model.SessionFlagFedEULAAccepted, true)
	assert.Nil(t, err)

	dbSess, err := dbInst.GetUserSession(testCtx, newUserSession.ID)
	assert.Nil(t, err)
	assert.True(t, dbSess.Flags[string(model.SessionFlagFedEULAAccepted)])
}

func TestDatabase_GetUserSSOSession(t *testing.T) {
	t.Run("Successful GetUserSSOSession (SAML)", func(t *testing.T) {
		var (
			testCtx      = context.Background()
			dbInst, user = initAndCreateUser(t)
			samlProvider = model.SAMLProvider{
				Name:            "provider",
				DisplayName:     "provider name",
				IssuerURI:       "https://idp.example.com/idp.xml",
				SingleSignOnURI: "https://idp.example.com/sso",
			}
		)

		// Initialize the SAMLProvider without setting SSOProviderID
		newSAMLProvider, err := dbInst.CreateSAMLIdentityProvider(testCtx, samlProvider, model.SSOProviderConfig{})
		require.Nil(t, err)

		user.SSOProviderID = newSAMLProvider.SSOProviderID
		err = dbInst.UpdateUser(testCtx, user)
		require.Nil(t, err)

		userSession := model.UserSession{
			AuthProviderID:   newSAMLProvider.ID,
			AuthProviderType: model.SessionAuthProviderSAML,
			User:             user,
			UserID:           user.ID,
			ExpiresAt:        time.Now().UTC().Add(time.Hour),
		}

		newUserSession, err := dbInst.CreateUserSession(testCtx, userSession)
		require.Nil(t, err)

		dbSess, err := dbInst.GetUserSession(testCtx, newUserSession.ID)
		require.Nil(t, err)
		require.NotNil(t, dbSess.User.SSOProvider)
		require.NotNil(t, dbSess.User.SSOProvider.SAMLProvider)
	})

	t.Run("Successful GetUserSSOSession (OIDC)", func(t *testing.T) {
		var (
			testCtx      = context.Background()
			dbInst, user = initAndCreateUser(t)
			oidcProvider = model.OIDCProvider{
				ClientID: "bloodhound",
				Issuer:   "https://localhost/auth",
			}
		)

		// Initialize the OIDCProvider without setting SSOProviderID
		newOIDCProvider, err := dbInst.CreateOIDCProvider(testCtx, "test", oidcProvider.Issuer, oidcProvider.ClientID, model.SSOProviderConfig{})
		require.Nil(t, err)

		user.SSOProviderID = null.Int32From(int32(newOIDCProvider.SSOProviderID))
		err = dbInst.UpdateUser(testCtx, user)
		require.Nil(t, err)

		userSession := model.UserSession{
			AuthProviderID:   newOIDCProvider.ID,
			AuthProviderType: model.SessionAuthProviderOIDC,
			User:             user,
			UserID:           user.ID,
			ExpiresAt:        time.Now().UTC().Add(time.Hour),
		}

		newUserSession, err := dbInst.CreateUserSession(testCtx, userSession)
		require.Nil(t, err)

		dbSess, err := dbInst.GetUserSession(testCtx, newUserSession.ID)
		require.Nil(t, err)
		require.NotNil(t, dbSess.User.SSOProvider)
		require.NotNil(t, dbSess.User.SSOProvider.OIDCProvider)
	})
}
