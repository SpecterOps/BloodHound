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

package appdb_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/auth/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedAuthProvider writes the persistence schema directly so the read tests do
// not depend on the legacy management APIs or their feature flag checks.
func seedAuthProvider(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name, slug string, providerType services.ProviderType, timestamp time.Time, withDetails bool) services.SSOProvider {
	t.Helper()

	var provider = services.SSOProvider{
		Name: name,
		Slug: slug,
		Type: providerType,
		Config: services.ProviderConfig{AutoProvision: services.AutoProvisionConfig{
			Enabled: true, DefaultRoleID: 3, RoleProvision: true,
		}},
	}

	err := pool.QueryRow(ctx,
		`INSERT INTO sso_providers (name, slug, type, config, created_at, updated_at)
 VALUES ($1, $2, $3, $4, $5, $5) RETURNING id`,
		name, slug, int(providerType),
		`{"auto_provision":{"enabled":true,"default_role_id":3,"role_provision":true}}`, timestamp,
	).Scan(&provider.ID)
	require.NoError(t, err)

	if !withDetails {
		return provider
	}

	switch providerType {
	case services.ProviderOIDC:
		provider.OIDCProvider = &services.OIDCProvider{
			ClientID:           "client-" + slug,
			Issuer:             "https://issuer.example.com/" + slug,
			SSOProviderID:      int(provider.ID),
			ProviderTimestamps: services.ProviderTimestamps{CreatedAt: timestamp, UpdatedAt: timestamp},
		}
		err = pool.QueryRow(ctx,
			`INSERT INTO oidc_providers (client_id, issuer, sso_provider_id, created_at, updated_at)
 VALUES ($1, $2, $3, $4, $4) RETURNING id`,
			provider.OIDCProvider.ClientID, provider.OIDCProvider.Issuer, provider.ID, timestamp,
		).Scan(&provider.OIDCProvider.ID)
		require.NoError(t, err)
	case services.ProviderSAML:
		provider.SAMLProvider = &services.SAMLProvider{
			Name:                       "legacy-" + slug,
			DisplayName:                name + " display",
			IssuerURI:                  "https://saml.example.com/issuer",
			SingleSignOnURI:            "https://saml.example.com/login",
			RootURIVersion:             1,
			PrincipalAttributeMappings: []string{"email", "principal_name"},
			SSOProviderID:              &provider.ID,
			ProviderTimestamps:         services.ProviderTimestamps{CreatedAt: timestamp, UpdatedAt: timestamp},
		}
		err = pool.QueryRow(ctx,
			`INSERT INTO saml_providers (name, display_name, issuer_uri, single_sign_on_uri, root_uri_version, ous, sso_provider_id, created_at, updated_at)
 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8) RETURNING id`,
			provider.SAMLProvider.Name, provider.SAMLProvider.DisplayName, provider.SAMLProvider.IssuerURI,
			provider.SAMLProvider.SingleSignOnURI, provider.SAMLProvider.RootURIVersion,
			provider.SAMLProvider.PrincipalAttributeMappings, provider.ID, timestamp,
		).Scan(&provider.SAMLProvider.ID)
		require.NoError(t, err)
	}

	return provider
}

func TestStore_ListAuthProviders_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStoreAndPool(t)
		timestamp   = time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	)

	// Enable both kinds before testing the empty and populated results.
	result, err := pool.Exec(ctx, `UPDATE feature_flags SET enabled = true WHERE key = 'oidc_support'`)
	require.NoError(t, err)
	require.EqualValues(t, 1, result.RowsAffected())

	t.Run("empty listing preserves nil", func(t *testing.T) {
		providers, err := store.ListAuthProviders(ctx, nil, nil)
		require.NoError(t, err)
		assert.Nil(t, providers)
	})

	var (
		saml           = seedAuthProvider(t, ctx, pool, "Alpha SAML", "alpha-saml", services.ProviderSAML, timestamp, true)
		oidc           = seedAuthProvider(t, ctx, pool, "Zulu O'IDC", "zulu-oidc", services.ProviderOIDC, timestamp.Add(time.Hour), true)
		otherOIDC      = seedAuthProvider(t, ctx, pool, "Beta OIDC", "beta-oidc", services.ProviderOIDC, timestamp.Add(2*time.Hour), true)
		missingDetails = seedAuthProvider(t, ctx, pool, "Missing Details", "missing-details", services.ProviderOIDC, timestamp.Add(3*time.Hour), false)
		testCases      = []struct {
			name      string
			filters   params.Filters
			sortItems params.SortItems
			expected  []services.SSOProvider
		}{
			{
				name:     "default creation order and complete details",
				expected: []services.SSOProvider{saml, oidc, otherOIDC, missingDetails},
			},
			{
				name:      "ascending name",
				sortItems: params.SortItems{{Field: "name", Direction: params.Ascending}},
				expected:  []services.SSOProvider{saml, otherOIDC, missingDetails, oidc},
			},
			{
				name:      "descending name",
				sortItems: params.SortItems{{Field: "name", Direction: params.Descending}},
				expected:  []services.SSOProvider{oidc, missingDetails, otherOIDC, saml},
			},
			{
				name: "multiple sort columns",
				sortItems: params.SortItems{
					{Field: "type", Direction: params.Descending},
					{Field: "name", Direction: params.Ascending},
				},
				expected: []services.SSOProvider{otherOIDC, missingDetails, oidc, saml},
			},
			{
				name:     "quoted string equality",
				filters:  params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: oidc.Name}}},
				expected: []services.SSOProvider{oidc},
			},
			{
				name:     "case insensitive substring",
				filters:  params.Filters{"name": {{Field: "name", Operator: params.ApproximatelyEquals, Value: "sAmL"}}},
				expected: []services.SSOProvider{saml},
			},
			{
				name:     "numeric type",
				filters:  params.Filters{"type": {{Field: "type", Operator: params.Equals, Value: "2"}}},
				expected: []services.SSOProvider{oidc, otherOIDC, missingDetails},
			},
			{
				name:     "numeric ID",
				filters:  params.Filters{"id": {{Field: "id", Operator: params.Equals, Value: strconv.FormatInt(int64(oidc.ID), 10)}}},
				expected: []services.SSOProvider{oidc},
			},
			{
				name:     "timestamp comparison",
				filters:  params.Filters{"created_at": {{Field: "created_at", Operator: params.GreaterThanOrEquals, Value: timestamp.Add(2 * time.Hour).Format(time.RFC3339Nano)}}},
				expected: []services.SSOProvider{otherOIDC, missingDetails},
			},
			{
				name: "combined fields and repeated bounds",
				filters: params.Filters{
					"type": {{Field: "type", Operator: params.Equals, Value: "2"}},
					"created_at": {
						{Field: "created_at", Operator: params.GreaterThanOrEquals, Value: timestamp.Add(time.Hour).Format(time.RFC3339Nano)},
						{Field: "created_at", Operator: params.LessThan, Value: timestamp.Add(3 * time.Hour).Format(time.RFC3339Nano)},
					},
				},
				sortItems: params.SortItems{{Field: "name", Direction: params.Ascending}},
				expected:  []services.SSOProvider{otherOIDC, oidc},
			},
			{
				name:    "filter values cannot inject SQL",
				filters: params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "' OR TRUE --"}}},
			},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			providers, err := store.ListAuthProviders(ctx, testCase.filters, testCase.sortItems)
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, providers)
		})
	}

	t.Run("null config and missing details", func(t *testing.T) {
		_, err := pool.Exec(ctx, `UPDATE sso_providers SET config = NULL WHERE id = $1`, missingDetails.ID)
		require.NoError(t, err)

		providers, err := store.ListAuthProviders(ctx, params.Filters{"slug": {{Field: "slug", Operator: params.Equals, Value: missingDetails.Slug}}}, nil)
		require.NoError(t, err)

		missingDetails.Config = services.ProviderConfig{}
		assert.Equal(t, []services.SSOProvider{missingDetails}, providers)
	})

	t.Run("OIDC feature flag disabled returns only SAML", func(t *testing.T) {
		_, err := pool.Exec(ctx, `UPDATE feature_flags SET enabled = false WHERE key = 'oidc_support'`)
		require.NoError(t, err)

		providers, err := store.ListAuthProviders(ctx, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, []services.SSOProvider{saml}, providers)
	})
}
