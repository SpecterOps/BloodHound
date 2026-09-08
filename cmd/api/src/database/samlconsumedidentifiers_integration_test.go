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
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateSAMLConsumedIdentifiers(t *testing.T) {
	const idpIssuer = "https://idp.example.com/12345678-90ab-cdef-1234-567890abcdef"

	tests := []struct {
		name string
		test func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		{
			name: "Success: consumes both identifiers (SAMLResponse and assertion) on first login",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				ssoProviderID := registerAndGetSSOProvider(t, testSuite, "provider-first-login")

				err := testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-1", "assertion-1", time.Now().Add(time.Hour))
				require.NoError(t, err)
			},
		},
		{
			name: "Error: rejects a full replay of both identifiers",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				ssoProviderID := registerAndGetSSOProvider(t, testSuite, "provider-full-replay")
				expiresAt := time.Now().Add(time.Hour)

				err := testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-2", "assertion-2", expiresAt)
				require.NoError(t, err)

				// replaying the exact same response/assertion pair is rejected
				err = testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-2", "assertion-2", expiresAt)
				assert.ErrorIs(t, err, database.ErrSAMLIdentifierAlreadyConsumed)
			},
		},
		{
			name: "Error: rejects and rolls back when only one identifier is a replay",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				ssoProviderID := registerAndGetSSOProvider(t, testSuite, "provider-partial-replay")
				expiresAt := time.Now().Add(time.Hour)

				err := testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-3", "assertion-3", expiresAt)
				require.NoError(t, err)

				// reuse the SAMLResponse ID with a new assertion ID: only one row is new
				err = testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-3", "assertion-3-new", expiresAt)
				assert.ErrorIs(t, err, database.ErrSAMLIdentifierAlreadyConsumed)

				// "assertion-3-new" was never persisted and can now be consumed alongside a new SAMLResponse ID
				err = testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-3-new", "assertion-3-new", expiresAt)
				require.NoError(t, err)
			},
		},
		{
			name: "Success: identical identifiers are allowed for a different sso provider",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				firstProviderID := registerAndGetSSOProvider(t, testSuite, "provider-scope-first")
				secondProviderID := registerAndGetSSOProvider(t, testSuite, "provider-scope-second")
				expiresAt := time.Now().Add(time.Hour)

				err := testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, firstProviderID, idpIssuer, "shared-response", "shared-assertion", expiresAt)
				require.NoError(t, err)

				// the composite primary key includes sso_provider_id, so the same SAMLResponse/assertion values under a
				// different provider are not a replay
				err = testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, secondProviderID, idpIssuer, "shared-response", "shared-assertion", expiresAt)
				require.NoError(t, err)
			},
		},
		{
			name: "Error: rejects identifiers for a non-existent sso provider",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				// intentionally skip creating a provider (foreign key) so the insert fails before any replay check
				const nonExistentSSOProviderID = int32(999999)

				err := testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, nonExistentSSOProviderID, idpIssuer, "response-4", "assertion-4", time.Now().Add(time.Hour))
				require.Error(t, err)
				// assert this failed on the foreign key, not as a replay rejection
				assert.NotErrorIs(t, err, database.ErrSAMLIdentifierAlreadyConsumed)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			testCase.test(t, testSuite)
		})
	}
}

func TestDatabase_SweepSAMLConsumedIdentifiers(t *testing.T) {
	const idpIssuer = "https://idp.example.com/12345678-90ab-cdef-1234-567890abcdef"

	tests := []struct {
		name string
		test func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		{
			name: "Success: no-op when the table is empty",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				err := testSuite.BHDatabase.SweepSAMLConsumedIdentifiers(testSuite.Context)
				require.NoError(t, err)

				assert.Empty(t, remainingSAMLConsumedIdentifiers(t, testSuite))
			},
		},
		{
			name: "Success: deletes expired records and keeps unexpired ones",
			test: func(t *testing.T, testSuite IntegrationTestSuite) {
				ssoProviderID := registerAndGetSSOProvider(t, testSuite, "1234")

				err := testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-expired", "assertion-expired", time.Now().Add(-time.Hour))
				require.NoError(t, err)

				err = testSuite.BHDatabase.CreateSAMLConsumedIdentifiers(testSuite.Context, ssoProviderID, idpIssuer, "response-live", "assertion-live", time.Now().Add(time.Hour))
				require.NoError(t, err)

				err = testSuite.BHDatabase.SweepSAMLConsumedIdentifiers(testSuite.Context)
				require.NoError(t, err)

				assert.ElementsMatch(t, []string{"response-live", "assertion-live"}, remainingSAMLConsumedIdentifiers(t, testSuite))
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			testCase.test(t, testSuite)
		})
	}
}

// registerAndGetSSOProvider creates an SSO provider that saml_consumed_identifiers can reference and returns its int32 ID.
func registerAndGetSSOProvider(t *testing.T, testSuite IntegrationTestSuite, name string) int32 {
	t.Helper()

	provider, err := testSuite.BHDatabase.CreateSSOProvider(testSuite.Context, name, model.SessionAuthProviderSAML, model.SSOProviderConfig{})
	require.NoError(t, err)

	return provider.ID
}

// remainingSAMLConsumedIdentifiers returns the identifier values currently stored in saml_consumed_identifiers.
func remainingSAMLConsumedIdentifiers(t *testing.T, testSuite IntegrationTestSuite) []string {
	t.Helper()

	var identifiers []string

	result := testSuite.DB.WithContext(testSuite.Context).Raw(`SELECT identifier FROM saml_consumed_identifiers`).Scan(&identifiers)
	require.NoError(t, result.Error)

	return identifiers
}
