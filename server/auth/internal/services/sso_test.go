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

package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/auth/internal/services"
	"github.com/specterops/bloodhound/server/auth/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ListAuthProviders(t *testing.T) {
	var (
		ctx               = context.Background()
		filters           = params.Filters{"name": {{Field: "name", Operator: params.Equals, Value: "Example"}}}
		sortItems         = params.SortItems{{Field: "name", Direction: params.Descending}}
		expectedProviders = []services.SSOProvider{{ID: 42, Name: "Example", Type: services.ProviderOIDC}}
		databaseErr       = errors.New("database unavailable")
	)

	testCases := []struct {
		name              string
		databaseProviders []services.SSOProvider
		databaseErr       error
		expectedProviders []services.SSOProvider
		expectedErr       error
	}{
		{
			name:              "Returns the providers listed by the database",
			databaseProviders: expectedProviders,
			expectedProviders: expectedProviders,
		},
		{
			name:        "Returns the database error",
			databaseErr: databaseErr,
			expectedErr: databaseErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				service      = services.NewService(databaseMock)
			)

			databaseMock.EXPECT().ListAuthProviders(ctx, filters, sortItems).
				Return(testCase.databaseProviders, testCase.databaseErr)

			providers, err := service.ListAuthProviders(ctx, filters, sortItems)

			require.ErrorIs(t, err, testCase.expectedErr)
			assert.Equal(t, testCase.expectedProviders, providers)
		})
	}
}

func TestProviderType_String(t *testing.T) {
	testCases := []struct {
		name           string
		providerType   services.ProviderType
		expectedString string
	}{
		{
			name:           "Returns SAML for the SAML provider type",
			providerType:   services.ProviderSAML,
			expectedString: "SAML",
		},
		{
			name:           "Returns OIDC for the OIDC provider type",
			providerType:   services.ProviderOIDC,
			expectedString: "OIDC",
		},
		{
			name:           "Returns Unknown for the zero value",
			providerType:   0,
			expectedString: "Unknown",
		},
		{
			name:           "Returns Unknown for an unrecognized value",
			providerType:   99,
			expectedString: "Unknown",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedString, testCase.providerType.String())
		})
	}
}
