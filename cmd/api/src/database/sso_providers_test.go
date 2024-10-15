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

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
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

	t.Run("successfully create an SSO provider", func(t *testing.T) {
		result, err := dbInst.CreateSSOProvider(testCtx, "Bloodhound Gang", model.SessionAuthProviderSAML)
		require.NoError(t, err)

		assert.Equal(t, "Bloodhound Gang", result.Name)
		assert.Equal(t, "bloodhound-gang", result.Slug)
		assert.Equal(t, model.SessionAuthProviderSAML, result.Type)

		provider, err := dbInst.GetSSOProvider(testCtx, int(result.ID))
		require.NoError(t, err)
		assert.Equal(t, model.SessionAuthProviderSAML, provider.Type)
	})
}

func TestBloodhoundDB_DeleteSSOProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully delete an SSO provider", func(t *testing.T) {
		provider, err := dbInst.CreateSSOProvider(testCtx, "test", model.SessionAuthProviderSAML)
		require.NoError(t, err)

		err = dbInst.DeleteSSOProvider(testCtx, int(provider.ID))
		require.NoError(t, err)

		provider, err = dbInst.GetSSOProvider(testCtx, int(provider.ID))
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}
