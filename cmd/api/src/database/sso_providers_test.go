// Copyright 2024 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_CreateAndGetSSOProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully create an SSO provider (SAML)", func(t *testing.T) {
		result, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang", model.SessionAuthProviderSAML, model.SSOProviderConfig{})
		require.NoError(t, err)

		assert.Equal(t, "Bloodhound Gang", result.Name)
		assert.Equal(t, "bloodhound-gang", result.Slug)
		assert.Equal(t, model.SessionAuthProviderSAML, result.Type)
		assert.NotEmpty(t, result.ID)
	})

	t.Run("successfully created an SSO provider with config values (SAML)", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		result, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang2", model.SessionAuthProviderSAML, config)
		require.NoError(t, err)

		assert.Equal(t, "Bloodhound Gang2", result.Name)
		assert.Equal(t, "bloodhound-gang2", result.Slug)
		assert.Equal(t, model.SessionAuthProviderSAML, result.Type)
		assert.Equal(t, true, result.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), result.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, result.Config.AutoProvision.RoleProvision)
		assert.NotEmpty(t, result.ID)
	})

	t.Run("successfully create an SSO provider (OIDC)", func(t *testing.T) {
		result, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang3", model.SessionAuthProviderOIDC, model.SSOProviderConfig{})
		require.NoError(t, err)

		assert.Equal(t, "Bloodhound Gang3", result.Name)
		assert.Equal(t, "bloodhound-gang3", result.Slug)
		assert.Equal(t, model.SessionAuthProviderOIDC, result.Type)
		assert.NotEmpty(t, result.ID)
	})

	t.Run("successfully created an SSO provider with config values (OIDC)", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		result, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang4", model.SessionAuthProviderOIDC, config)
		require.NoError(t, err)

		assert.Equal(t, "Bloodhound Gang4", result.Name)
		assert.Equal(t, "bloodhound-gang4", result.Slug)
		assert.Equal(t, model.SessionAuthProviderOIDC, result.Type)
		assert.Equal(t, true, result.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), result.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, result.Config.AutoProvision.RoleProvision)
		assert.NotEmpty(t, result.ID)
	})
}

func TestBloodhoundDB_DeleteSSOProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully delete an SSO provider associated with a SAML provider", func(t *testing.T) {
		samlProvider, err := dbInst.CreateSAMLIdentityProvider(testCtx, model.SAMLProvider{Name: "test"}, model.SSOProviderConfig{})
		require.NoError(t, err)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: samlProvider.SSOProviderID,
			PrincipalName: userPrincipal,
		})
		require.NoError(t, err)

		err = dbInst.DeleteSSOProvider(testCtx, int(samlProvider.SSOProviderID.Int32))
		require.NoError(t, err)

		user, err = dbInst.GetUser(testCtx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, null.NewInt32(0, false), user.SSOProviderID)
	})

	t.Run("successfully delete an SSO provider associated with a SAML provider with config values", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		samlProvider, err := dbInst.CreateSAMLIdentityProvider(testCtx, model.SAMLProvider{Name: "test2"}, config)
		require.NoError(t, err)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: samlProvider.SSOProviderID,
			PrincipalName: user2Principal,
			EmailAddress:  null.StringFrom(user2Principal),
		})
		require.NoError(t, err)

		err = dbInst.DeleteSSOProvider(testCtx, int(samlProvider.SSOProviderID.Int32))
		require.NoError(t, err)

		user, err = dbInst.GetUser(testCtx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, null.NewInt32(0, false), user.SSOProviderID)
	})

	t.Run("successfully delete an SSO provider associated with an OIDC provider", func(t *testing.T) {
		oidcProvider, err := dbInst.CreateOIDCProvider(testCtx, "test3", "test3", "test3", model.SSOProviderConfig{})
		require.NoError(t, err)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: null.Int32From(int32(oidcProvider.SSOProviderID)),
			PrincipalName: user3Principal,
			EmailAddress:  null.StringFrom(user3Principal),
		})
		require.NoError(t, err)

		err = dbInst.DeleteSSOProvider(testCtx, oidcProvider.SSOProviderID)
		require.NoError(t, err)

		user, err = dbInst.GetUser(testCtx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, null.NewInt32(0, false), user.SSOProviderID)
	})

	t.Run("successfully delete an SSO provider associated with an OIDC provider with config values", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		oidcProvider, err := dbInst.CreateOIDCProvider(testCtx, "test4", "test4", "test4", config)
		require.NoError(t, err)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: null.Int32From(int32(oidcProvider.SSOProviderID)),
			PrincipalName: user4Principal,
			EmailAddress:  null.StringFrom(user4Principal),
		})
		require.NoError(t, err)

		err = dbInst.DeleteSSOProvider(testCtx, oidcProvider.SSOProviderID)
		require.NoError(t, err)

		user, err = dbInst.GetUser(testCtx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, null.NewInt32(0, false), user.SSOProviderID)
	})
}

func TestBloodhoundDB_GetAllSSOProviders(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully list SSO providers with and without sorting", func(t *testing.T) {
		// Create SSO providers
		provider1, err := dbInst.CreateSSOProvider(testCtx, "First Provider", model.SessionAuthProviderSAML, model.SSOProviderConfig{})
		require.NoError(t, err)

		provider2, err := dbInst.CreateSSOProvider(testCtx, "Second Provider", model.SessionAuthProviderOIDC, model.SSOProviderConfig{})
		require.NoError(t, err)

		// Enable the OIDC feature flag
		oidcFlag, err := dbInst.GetFlagByKey(testCtx, appcfg.FeatureOIDCSupport)
		require.NoError(t, err)
		oidcFlag.Enabled = true

		err = dbInst.SetFlag(testCtx, oidcFlag)
		require.NoError(t, err)

		// Test default ordering (by created_at)
		providers, err := dbInst.GetAllSSOProviders(testCtx, "", model.SQLFilter{})
		require.NoError(t, err)
		require.Len(t, providers, 2)
		assert.Equal(t, provider1.ID, providers[0].ID)
		assert.Equal(t, provider2.ID, providers[1].ID)

		// Test ordering by name descending
		providers, err = dbInst.GetAllSSOProviders(testCtx, "name desc", model.SQLFilter{})
		require.NoError(t, err)
		require.Len(t, providers, 2)
		assert.Equal(t, provider2.ID, providers[0].ID)
		assert.Equal(t, provider1.ID, providers[1].ID)

		// Test filtering by name
		sqlFilter := model.SQLFilter{
			SQLString: "name = ?",
			Params:    []interface{}{"First Provider"},
		}
		providers, err = dbInst.GetAllSSOProviders(testCtx, "", sqlFilter)
		require.NoError(t, err)
		require.Len(t, providers, 1)
		assert.Equal(t, provider1.ID, providers[0].ID)
	})

	// This test fails individually, but passes when ran together with the other tests
	t.Run("successfully list SSO providers with and without sorting (with configs)", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		// Create SSO providers
		provider3, err := dbInst.CreateSSOProvider(testCtx, "Third Provider", model.SessionAuthProviderSAML, config)
		require.NoError(t, err)
		assert.Equal(t, true, provider3.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), provider3.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, provider3.Config.AutoProvision.RoleProvision)

		provider4, err := dbInst.CreateSSOProvider(testCtx, "Fourth Provider", model.SessionAuthProviderOIDC, config)
		require.NoError(t, err)
		assert.Equal(t, true, provider4.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), provider4.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, provider4.Config.AutoProvision.RoleProvision)

		// Enable the OIDC feature flag
		oidcFlag, err := dbInst.GetFlagByKey(testCtx, appcfg.FeatureOIDCSupport)
		require.NoError(t, err)
		oidcFlag.Enabled = true

		err = dbInst.SetFlag(testCtx, oidcFlag)
		require.NoError(t, err)

		// Test default ordering (by created_at)
		providers, err := dbInst.GetAllSSOProviders(testCtx, "", model.SQLFilter{})
		require.NoError(t, err)
		require.Len(t, providers, 4)
		assert.Equal(t, provider3.ID, providers[2].ID)
		assert.Equal(t, provider4.ID, providers[3].ID)

		// Test ordering by name descending
		providers, err = dbInst.GetAllSSOProviders(testCtx, "name desc", model.SQLFilter{})
		require.NoError(t, err)
		require.Len(t, providers, 4)
		assert.Equal(t, provider3.ID, providers[0].ID)
		assert.Equal(t, provider4.ID, providers[2].ID)

		// Test filtering by name
		sqlFilter := model.SQLFilter{
			SQLString: "name = ?",
			Params:    []interface{}{"Third Provider"},
		}
		providers, err = dbInst.GetAllSSOProviders(testCtx, "", sqlFilter)
		require.NoError(t, err)
		require.Len(t, providers, 1)
		assert.Equal(t, provider3.ID, providers[0].ID)
	})
}

func TestBloodhoundDB_GetSSOProviderBySlug(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		config  = model.SSOProviderConfig{}
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully get sso provider by slug (OIDC)", func(t *testing.T) {
		newProvider, err := dbInst.CreateOIDCProvider(testCtx, "Gotham Net", "https://test.localhost.com/auth", "gotham-net", config)
		require.Nil(t, err)

		provider, err := dbInst.GetSSOProviderBySlug(testCtx, "gotham-net")
		require.Nil(t, err)
		require.EqualValues(t, newProvider.SSOProviderID, provider.ID)
		require.NotNil(t, provider.OIDCProvider)
		require.Equal(t, newProvider.ClientID, provider.OIDCProvider.ClientID)
		require.Equal(t, newProvider.Issuer, provider.OIDCProvider.Issuer)
	})

	t.Run("successfully get sso provider by slug (OIDC) with configs", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		newProvider, err := dbInst.CreateOIDCProvider(testCtx, "Gotham Net2", "https://test.localhost.com/auth", "gotham-net2", config)
		require.Nil(t, err)

		provider, err := dbInst.GetSSOProviderBySlug(testCtx, "gotham-net2")
		require.Nil(t, err)
		require.EqualValues(t, newProvider.SSOProviderID, provider.ID)
		require.NotNil(t, provider.OIDCProvider)
		require.Equal(t, newProvider.ClientID, provider.OIDCProvider.ClientID)
		require.Equal(t, newProvider.Issuer, provider.OIDCProvider.Issuer)
		assert.Equal(t, true, provider.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), provider.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, provider.Config.AutoProvision.RoleProvision)
	})
}

func TestBloodhoundDB_GetSSOProviderUsers(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully list SSO provider users (SAML)", func(t *testing.T) {
		provider, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang", model.SessionAuthProviderSAML, model.SSOProviderConfig{})
		require.NoError(t, err)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: null.Int32From(provider.ID),
			PrincipalName: userPrincipal,
			EmailAddress:  null.StringFrom(userPrincipal),
		})
		require.NoError(t, err)

		returnedUsers, err := dbInst.GetSSOProviderUsers(testCtx, int(provider.ID))
		require.NoError(t, err)

		require.Len(t, returnedUsers, 1)
		assert.Equal(t, user.ID, returnedUsers[0].ID)
	})

	t.Run("successfully list SSO provider users (SAML) with configs", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		provider, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang2", model.SessionAuthProviderSAML, config)
		require.NoError(t, err)
		assert.Equal(t, true, provider.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), provider.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, provider.Config.AutoProvision.RoleProvision)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: null.Int32From(provider.ID),
			PrincipalName: user2Principal,
			EmailAddress:  null.StringFrom(user2Principal),
		})
		require.NoError(t, err)

		returnedUsers, err := dbInst.GetSSOProviderUsers(testCtx, int(provider.ID))
		require.NoError(t, err)

		require.Len(t, returnedUsers, 1)
		assert.Equal(t, user.ID, returnedUsers[0].ID)
	})

	t.Run("successfully list SSO provider users (OIDC)", func(t *testing.T) {
		provider, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang3", model.SessionAuthProviderOIDC, model.SSOProviderConfig{})
		require.NoError(t, err)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: null.Int32From(provider.ID),
			PrincipalName: user3Principal,
			EmailAddress:  null.StringFrom(user3Principal),
		})
		require.NoError(t, err)

		returnedUsers, err := dbInst.GetSSOProviderUsers(testCtx, int(provider.ID))
		require.NoError(t, err)

		require.Len(t, returnedUsers, 1)
		assert.Equal(t, user.ID, returnedUsers[0].ID)
	})

	t.Run("successfully list SSO provider users (OIDC) with configs", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}

		provider, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang4", model.SessionAuthProviderOIDC, config)
		require.NoError(t, err)
		assert.Equal(t, true, provider.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), provider.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, provider.Config.AutoProvision.RoleProvision)

		user, err := dbInst.CreateUser(testCtx, model.User{
			SSOProviderID: null.Int32From(provider.ID),
			PrincipalName: user4Principal,
			EmailAddress:  null.StringFrom(user4Principal),
		})
		require.NoError(t, err)

		returnedUsers, err := dbInst.GetSSOProviderUsers(testCtx, int(provider.ID))
		require.NoError(t, err)

		require.Len(t, returnedUsers, 1)
		assert.Equal(t, user.ID, returnedUsers[0].ID)
	})
}

func TestBloodhoundDB_GetSSOProviderById(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully get sso provider by id (SAML)", func(t *testing.T) {
		newSamlProvider, err := dbInst.CreateSAMLIdentityProvider(testCtx, model.SAMLProvider{
			Name:        "someName",
			DisplayName: "someName",
		}, model.SSOProviderConfig{})
		require.NoError(t, err)

		provider, err := dbInst.GetSSOProviderById(testCtx, newSamlProvider.SSOProviderID.Int32)
		require.NoError(t, err)

		require.EqualValues(t, newSamlProvider.SSOProviderID.Int32, provider.ID)
		require.NotNil(t, provider.SAMLProvider)
	})

	t.Run("successfully get sso provider by id with config values (SAML)", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}
		newSamlProvider, err := dbInst.CreateSAMLIdentityProvider(testCtx, model.SAMLProvider{
			Name:        "someName2",
			DisplayName: "someName2",
		}, config)
		require.NoError(t, err)

		provider, err := dbInst.GetSSOProviderById(testCtx, newSamlProvider.SSOProviderID.Int32)
		require.NoError(t, err)
		assert.Equal(t, true, provider.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), provider.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, provider.Config.AutoProvision.RoleProvision)

		require.EqualValues(t, newSamlProvider.SSOProviderID.Int32, provider.ID)
		require.NotNil(t, provider.SAMLProvider)
	})

	t.Run("successfully get sso provider by id (OIDC)", func(t *testing.T) {
		oidcProvider := model.OIDCProvider{
			ClientID: "bloodhound",
			Issuer:   "https://localhost/auth",
		}

		newOIDCProvider, err := dbInst.CreateOIDCProvider(testCtx, "test", oidcProvider.Issuer, oidcProvider.ClientID, model.SSOProviderConfig{})
		require.Nil(t, err)

		provider, err := dbInst.GetSSOProviderById(testCtx, int32(newOIDCProvider.SSOProviderID))
		require.NoError(t, err)

		require.EqualValues(t, int32(newOIDCProvider.SSOProviderID), provider.ID)
		require.NotNil(t, provider.OIDCProvider)
	})

	t.Run("successfully get sso provider by id with config values (OIDC)", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRoleId: 3,
				RoleProvision: true,
			},
		}
		oidcProvider := model.OIDCProvider{
			ClientID: "bloodhound2",
			Issuer:   "https://localhost/auth",
		}

		newOIDCProvider, err := dbInst.CreateOIDCProvider(testCtx, "test2", oidcProvider.Issuer, oidcProvider.ClientID, config)
		require.Nil(t, err)

		provider, err := dbInst.GetSSOProviderById(testCtx, int32(newOIDCProvider.SSOProviderID))
		require.NoError(t, err)
		assert.Equal(t, true, provider.Config.AutoProvision.Enabled)
		assert.Equal(t, int32(3), provider.Config.AutoProvision.DefaultRoleId)
		assert.Equal(t, true, provider.Config.AutoProvision.RoleProvision)

		require.EqualValues(t, int32(newOIDCProvider.SSOProviderID), provider.ID)
		require.NotNil(t, provider.OIDCProvider)
	})
}
