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

	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_CreateOIDCProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully create an OIDC provider", func(t *testing.T) {
		provider, err := dbInst.CreateOIDCProvider(testCtx, "test", "https://test.localhost.com/auth", "bloodhound")
		require.NoError(t, err)

		assert.Equal(t, "test", provider.Name)
		assert.Equal(t, "https://test.localhost.com/auth", provider.Issuer)
		assert.Equal(t, "bloodhound", provider.ClientID)
	})
}

func TestBloodhoundDB_GetAllOIDCProviders(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully get all OIDC provider", func(t *testing.T) {
		firstProvider, err := dbInst.CreateOIDCProvider(testCtx, "test", "https://test.localhost.com/auth", "bloodhound")
		secondProvider, err := dbInst.CreateOIDCProvider(testCtx, "test_2", "https://test.localhost.com/auth", "another_client")
		require.NoError(t, err)

		providers, err := dbInst.GetAllOIDCProviders(testCtx)
		require.NoError(t, err)

		require.Len(t, providers, 2)
		assert.Equal(t, firstProvider, providers[0])
		assert.Equal(t, secondProvider, providers[1])
	})
}
