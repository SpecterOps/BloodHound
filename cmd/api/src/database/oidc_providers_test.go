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
	"time"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

// This test fails individually, but passes when ran together with the other tests
func TestBloodhoundDB_CreateUpdateOIDCProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	defer dbInst.Close(testCtx)

	t.Run("successfully create and update an OIDC provider", func(t *testing.T) {
		provider, err := dbInst.CreateOIDCProvider(testCtx, "test", "https://test.localhost.com/auth", "bloodhound", model.SSOProviderConfig{})
		require.NoError(t, err)

		require.Equal(t, "https://test.localhost.com/auth", provider.Issuer)
		require.Equal(t, "bloodhound", provider.ClientID)
		require.EqualValues(t, 1, provider.ID)

		_, count, err := dbInst.ListAuditLogs(testCtx, time.Now().Add(time.Minute), time.Now().Add(-time.Minute), 0, 10, "", model.SQLFilter{})
		require.NoError(t, err)
		require.Equal(t, 4, count)

		updatedSSOProvider := model.SSOProvider{
			Name: "updated provider",
			Type: model.SessionAuthProviderOIDC,
			OIDCProvider: &model.OIDCProvider{
				Serial: model.Serial{
					ID: provider.ID,
				},
				ClientID:      "gotham-net",
				Issuer:        "https://gotham.net",
				SSOProviderID: provider.SSOProviderID,
			},
		}

		provider, err = dbInst.UpdateOIDCProvider(testCtx, updatedSSOProvider)
		require.NoError(t, err)

		require.Equal(t, updatedSSOProvider.OIDCProvider.Issuer, provider.Issuer)
		require.Equal(t, updatedSSOProvider.OIDCProvider.ClientID, provider.ClientID)
		require.EqualValues(t, updatedSSOProvider.OIDCProvider.ID, provider.ID)

		_, count, err = dbInst.ListAuditLogs(testCtx, time.Now().Add(time.Minute), time.Now().Add(-time.Minute), 0, 10, "", model.SQLFilter{})
		require.NoError(t, err)
		require.Equal(t, 8, count)
	})

	t.Run("successfully create and update an OIDC provider with config values", func(t *testing.T) {
		config := model.SSOProviderConfig{
			AutoProvision: model.SSOProviderAutoProvisionConfig{
				Enabled:       true,
				DefaultRole:   3,
				RoleProvision: true,
			},
		}

		provider, err := dbInst.CreateOIDCProvider(testCtx, "test2", "https://test.localhost.com/auth", "bloodhound", config)
		require.NoError(t, err)

		require.Equal(t, "https://test.localhost.com/auth", provider.Issuer)
		require.Equal(t, "bloodhound", provider.ClientID)
		require.EqualValues(t, 2, provider.ID)

		_, count, err := dbInst.ListAuditLogs(testCtx, time.Now().Add(time.Minute), time.Now().Add(-time.Minute), 0, 10, "", model.SQLFilter{})
		require.NoError(t, err)
		require.Equal(t, 12, count)

		updatedSSOProvider := model.SSOProvider{
			Name: "updated provider2",
			Type: model.SessionAuthProviderOIDC,
			OIDCProvider: &model.OIDCProvider{
				Serial: model.Serial{
					ID: provider.ID,
				},
				ClientID:      "gotham-net",
				Issuer:        "https://gotham.net",
				SSOProviderID: provider.SSOProviderID,
			},
			Config: model.SSOProviderConfig{
				AutoProvision: model.SSOProviderAutoProvisionConfig{
					Enabled:       true,
					DefaultRole:   2,
					RoleProvision: false,
				},
			},
		}

		provider, err = dbInst.UpdateOIDCProvider(testCtx, updatedSSOProvider)
		require.NoError(t, err)

		require.Equal(t, updatedSSOProvider.OIDCProvider.Issuer, provider.Issuer)
		require.Equal(t, updatedSSOProvider.OIDCProvider.ClientID, provider.ClientID)
		require.EqualValues(t, updatedSSOProvider.OIDCProvider.ID, provider.ID)

		_, count, err = dbInst.ListAuditLogs(testCtx, time.Now().Add(time.Minute), time.Now().Add(-time.Minute), 0, 10, "", model.SQLFilter{})
		require.NoError(t, err)
		require.Equal(t, 16, count)
	})
}
