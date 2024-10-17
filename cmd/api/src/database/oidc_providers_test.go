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

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_GetOIDCProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully get an OIDC provider", func(t *testing.T) {
		provider, err := dbInst.CreateOIDCProvider(testCtx, "test", "https://test.localhost.com/auth", "bloodhound")
		require.NoError(t, err)

		fetchedProvider, err := dbInst.GetOIDCProvider(testCtx, int(provider.ID))
		require.NoError(t, err)
		assert.Equal(t, fetchedProvider.Issuer, provider.Issuer)
		assert.Equal(t, fetchedProvider.ClientID, provider.ClientID)
	})

	t.Run("error OIDC provider not found", func(t *testing.T) {
		_, err := dbInst.GetOIDCProvider(testCtx, 1234)
		require.Error(t, err)
		assert.ErrorIs(t, err, database.ErrNotFound)
	})
}

func TestBloodhoundDB_CreateOIDCProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully create an OIDC provider", func(t *testing.T) {
		provider, err := dbInst.CreateOIDCProvider(testCtx, "test", "https://test.localhost.com/auth", "bloodhound")
		require.NoError(t, err)

		assert.Equal(t, "https://test.localhost.com/auth", provider.Issuer)
		assert.Equal(t, "bloodhound", provider.ClientID)

		_, count, err := dbInst.ListAuditLogs(testCtx, time.Now().Add(-time.Minute), time.Now().Add(time.Minute), 0, 10, "", model.SQLFilter{})
		require.NoError(t, err)
		assert.Equal(t, 4, count)
	})
}
