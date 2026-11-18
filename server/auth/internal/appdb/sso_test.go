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

package appdb_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v6"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/auth/internal/appdb"
	"github.com/specterops/bloodhound/server/auth/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const providerQuery = `SELECT sso.id, sso.name, sso.type, sso.slug, sso.config, ` +
	`to_jsonb(oidc) AS oidc_provider, to_jsonb(saml) AS saml_provider FROM ` + appdb.TableSSOProviders + ` AS sso ` +
	`LEFT JOIN ` + appdb.TableOIDCProviders + ` AS oidc ON oidc.sso_provider_id = sso.id ` +
	`LEFT JOIN ` + appdb.TableSAMLProviders + ` AS saml ON saml.sso_provider_id = sso.id ` +
	`WHERE (sso.type = 1 OR (SELECT enabled FROM feature_flags WHERE key = 'oidc_support'))`

func providerColumns() []string {
	return []string{"id", "name", "type", "slug", "config", "oidc_provider", "saml_provider"}
}

func TestStore_ListAuthProviders_SQL(t *testing.T) {
	var (
		ctx      = context.Background()
		queryErr = errors.New("query failed")
	)

	testCases := []struct {
		name          string
		filters       params.Filters
		sortItems     params.SortItems
		setupMock     func(pool pgxmock.PgxPoolIface)
		expectedCount int
		expectedErr   error
		expectStrErr  string
		assertResult  func(t *testing.T, providers []services.SSOProvider)
	}{
		{
			name: "Uses the default sort and decodes both provider kinds",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				rows := pool.NewRows(providerColumns()).
					AddRow(int32(1), "OIDC", 2, "oidc", []byte(`{"auto_provision":{"enabled":true,"default_role_id":3}}`), []byte(`{"id":10,"client_id":"client","issuer":"issuer","sso_provider_id":1}`), nil).
					AddRow(int32(2), "SAML", 1, "saml", nil, nil, []byte(`{"id":20,"name":"saml","issuer_uri":"issuer","ous":["email"],"sso_provider_id":2}`))
				pool.ExpectQuery(providerQuery + " ORDER BY sso.created_at").WithArgs().WillReturnRows(rows)
			},
			expectedCount: 2,
			assertResult: func(t *testing.T, providers []services.SSOProvider) {
				require.Len(t, providers, 2)
				require.NotNil(t, providers[0].OIDCProvider)
				require.NotNil(t, providers[1].SAMLProvider)
				assert.True(t, providers[0].Config.AutoProvision.Enabled)
				assert.Equal(t, "client", providers[0].OIDCProvider.ClientID)
				assert.Equal(t, int32(10), providers[0].OIDCProvider.ID)
				assert.Equal(t, []string{"email"}, providers[1].SAMLProvider.PrincipalAttributeMappings)
				assert.Equal(t, "issuer", providers[1].SAMLProvider.IssuerURI)
			},
		},
		{
			name: "Returns an error for malformed OIDC provider JSON",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery + " ORDER BY sso.created_at").WithArgs().
					WillReturnRows(pool.NewRows(providerColumns()).
						AddRow(int32(1), "OIDC", 2, "oidc", []byte(`{}`), []byte(`{`), nil))
			},
			expectStrErr: "decoding OIDC provider",
			assertResult: func(t *testing.T, providers []services.SSOProvider) {
				assert.Nil(t, providers)
			},
		},
		{
			name: "Returns an error for malformed SAML provider JSON",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery + " ORDER BY sso.created_at").WithArgs().
					WillReturnRows(pool.NewRows(providerColumns()).
						AddRow(int32(2), "SAML", 1, "saml", []byte(`{}`), nil, []byte(`{`)))
			},
			expectStrErr: "decoding SAML provider",
			assertResult: func(t *testing.T, providers []services.SSOProvider) {
				assert.Nil(t, providers)
			},
		},
		{
			name: "Returns an error for malformed SSO config JSON",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery + " ORDER BY sso.created_at").WithArgs().
					WillReturnRows(pool.NewRows(providerColumns()).
						AddRow(int32(1), "OIDC", 2, "oidc", []byte(`{`), []byte(`{"client_id":"client"}`), nil))
			},
			expectStrErr: "decoding SSO config",
			assertResult: func(t *testing.T, providers []services.SSOProvider) {
				assert.Nil(t, providers)
			},
		},
		{
			name:      "Applies filter and sort arguments",
			filters:   params.Filters{"name": {{Field: "name", Operator: params.ApproximatelyEquals, Value: "O'IDC"}}},
			sortItems: params.SortItems{{Field: "name", Direction: params.Descending}},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery + " AND sso.name ILIKE $1 ORDER BY sso.name DESC").
					WithArgs("%O'IDC%").
					WillReturnRows(pool.NewRows(providerColumns()))
			},
		},
		{
			name: "Combines repeated filters with AND by default",
			filters: params.Filters{
				"name": {
					{Field: "name", Operator: params.NotEquals, Value: "Alpha"},
					{Field: "name", Operator: params.NotEquals, Value: "Beta"},
				},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery+" AND sso.name <> $1 AND sso.name <> $2 ORDER BY sso.created_at").
					WithArgs("Alpha", "Beta").
					WillReturnRows(pool.NewRows(providerColumns()))
			},
		},
		{
			name: "Groups OR filters and combines separate fields with AND",
			filters: params.Filters{
				"name": {
					{Field: "name", Operator: params.Equals, Value: "Alpha", SetOperator: params.FilterOr},
					{Field: "name", Operator: params.Equals, Value: "Beta", SetOperator: params.FilterOr},
				},
				"type": {{Field: "type", Operator: params.Equals, Value: "2"}},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery+" AND (sso.name = $1 OR sso.name = $2) AND sso.type = $3 ORDER BY sso.created_at").
					WithArgs("Alpha", "Beta", int32(2)).
					WillReturnRows(pool.NewRows(providerColumns()))
			},
		},
		{
			name: "Converts typed values and orders fields deterministically",
			filters: params.Filters{
				"type":       {{Field: "type", Operator: params.Equals, Value: "2"}},
				"created_at": {{Field: "created_at", Operator: params.GreaterThanOrEquals, Value: "2026-01-01T00:00:00Z"}},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery+" AND sso.created_at >= $1 AND sso.type = $2 ORDER BY sso.created_at").
					WithArgs(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), int32(2)).
					WillReturnRows(pool.NewRows(providerColumns()))
			},
		},
		{
			name:        "Rejects unsupported filter operators before querying",
			filters:     params.Filters{"name": {{Field: "name", Operator: "invalid", Value: "Alpha"}}},
			expectedErr: services.ErrInvalidProviderQuery,
		},
		{
			name: "Wraps the list query error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery + " ORDER BY sso.created_at").WithArgs().WillReturnError(queryErr)
			},
			expectStrErr: queryErr.Error(),
		},
		{
			name: "Wraps the row reading error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(providerQuery + " ORDER BY sso.created_at").WithArgs().
					WillReturnRows(pool.NewRows(providerColumns()).AddRow(1, "OIDC", 2, "oidc", nil, nil, nil).RowError(0, queryErr))
			},
			expectStrErr: queryErr.Error(),
		},
		{
			name:        "Rejects non-integer ID filters before querying",
			filters:     params.Filters{"id": {{Field: "id", Operator: params.Equals, Value: "invalid"}}},
			expectedErr: services.ErrInvalidProviderQuery,
		},
		{
			name:        "Rejects non-RFC3339 timestamp filters before querying",
			filters:     params.Filters{"created_at": {{Field: "created_at", Operator: params.Equals, Value: "invalid"}}},
			expectedErr: services.ErrInvalidProviderQuery,
		},
		{
			name:        "Rejects unknown filter fields before querying",
			filters:     params.Filters{"details": {{Field: "details", Operator: params.Equals, Value: "invalid"}}},
			expectedErr: services.ErrInvalidProviderQuery,
		},
		{
			name:        "Rejects unsafe sort fields before querying",
			sortItems:   params.SortItems{{Field: "id; DROP TABLE sso_providers", Direction: params.Ascending}},
			expectedErr: services.ErrInvalidProviderQuery,
		},
		{
			name:        "Rejects unsupported sort directions before querying",
			sortItems:   params.SortItems{{Field: "name", Direction: "sideways"}},
			expectedErr: services.ErrInvalidProviderQuery,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store, pool := newTestStore(t)

			if testCase.setupMock != nil {
				testCase.setupMock(pool)
			}

			providers, err := store.ListAuthProviders(ctx, testCase.filters, testCase.sortItems)

			switch {
			case testCase.expectedErr != nil:
				assert.ErrorIs(t, err, testCase.expectedErr)
			case testCase.expectStrErr != "":
				require.Error(t, err)
				assert.ErrorContains(t, err, testCase.expectStrErr)
			default:
				require.NoError(t, err)
			}

			assert.Len(t, providers, testCase.expectedCount)

			if testCase.assertResult != nil {
				testCase.assertResult(t, providers)
			}

			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
