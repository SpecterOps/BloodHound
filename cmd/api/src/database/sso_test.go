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

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_CreateSSOProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully create an SSO provider", func(t *testing.T) {
		result, err := dbInst.CreateSSOProvider(testCtx, "name", "some_slug", model.SessionAuthProviderSAML)
		require.NoError(t, err)

		assert.Equal(t, "name", result.Name)
		assert.Equal(t, "some_slug", result.Slug)
		assert.Equal(t, model.SessionAuthProviderSAML, result.Type)
	})
}

func TestBloodhoundDB_ListSSOProviders(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully list SSO providers", func(t *testing.T) {
		provider1, err := dbInst.CreateSSOProvider(testCtx, "Provider One", "provider-one", model.SessionAuthProviderSAML)
		require.NoError(t, err)

		provider2, err := dbInst.CreateSSOProvider(testCtx, "Provider Two", "provider-two", model.SessionAuthProviderOIDC)
		require.NoError(t, err)

		providers, err := dbInst.GetAllSSOProviders(testCtx)
		require.NoError(t, err)

		assert.Len(t, providers, 2)

		providerMap := make(map[int32]model.SSOProvider)
		for _, p := range providers {
			providerMap[p.ID] = p
		}

		retrievedProvider1, exists := providerMap[provider1.ID]
		require.True(t, exists, "Provider One not found in the list")
		assert.Equal(t, provider1.Name, retrievedProvider1.Name)
		assert.Equal(t, provider1.Slug, retrievedProvider1.Slug)
		assert.Equal(t, provider1.Type, retrievedProvider1.Type)

		retrievedProvider2, exists := providerMap[provider2.ID]
		require.True(t, exists, "Provider Two not found in the list")
		assert.Equal(t, provider2.Name, retrievedProvider2.Name)
		assert.Equal(t, provider2.Slug, retrievedProvider2.Slug)
		assert.Equal(t, provider2.Type, retrievedProvider2.Type)
	})
}
