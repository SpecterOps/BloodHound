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
		result, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang", model.SessionAuthProviderSAML)
		require.NoError(t, err)

		assert.Equal(t, "Bloodhound Gang", result.Name)
		assert.Equal(t, "bloodhound-gang", result.Slug)
		assert.Equal(t, model.SessionAuthProviderSAML, result.Type)
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
		provider1, err := dbInst.CreateSSOProvider(testCtx, "First Provider", model.SessionAuthProviderSAML)
		require.NoError(t, err)

		provider2, err := dbInst.CreateSSOProvider(testCtx, "Second Provider", model.SessionAuthProviderOIDC)
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
}
