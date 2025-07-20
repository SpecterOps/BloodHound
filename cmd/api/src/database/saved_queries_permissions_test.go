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

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
)

func TestSavedQueriesPermissions_CreateSavedQueryPermissionToPublic(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		user    = createUser(t, dbInst, userPrincipal)
		user2   = createUser(t, dbInst, user2Principal)
	)

	t.Run("Creates saved query permission to public", func(t *testing.T) {
		query, err := dbInst.CreateSavedQuery(testCtx, user.ID, "Test Query", "TESTING", "Example")
		require.NoError(t, err)

		_, err = dbInst.CreateSavedQueryPermissionToPublic(testCtx, query.ID)
		require.NoError(t, err)

		scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeOwned:  true,
			model.SavedQueryScopeShared: false,
		}, scope)
	})

	t.Run("Creates saved query permission to public while deleting previous user's shared query permission", func(t *testing.T) {
		query, err := dbInst.CreateSavedQuery(testCtx, user.ID, "Test Query2", "TESTING2", "Example2")
		require.NoError(t, err)

		_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user2.ID)
		require.NoError(t, err)

		scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: true,
		}, scope)

		_, err = dbInst.CreateSavedQueryPermissionToPublic(testCtx, query.ID)
		require.NoError(t, err)

		scope2, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: true,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: false,
		}, scope2)
	})
}

func TestSavedQueriesPermissions_CreateSavedQueryPermissionsToUsers(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		user1   = createUser(t, dbInst, userPrincipal)
		user2   = createUser(t, dbInst, user2Principal)
		user3   = createUser(t, dbInst, user3Principal)
		user4   = createUser(t, dbInst, user4Principal)
	)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user2.ID, user3.ID, user4.ID)
	require.NoError(t, err)

	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
	require.NoError(t, err)
	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: true,
	}, scope)

	scope2, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user3.ID)
	require.NoError(t, err)
	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: true,
	}, scope2)

	scope3, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user4.ID)
	require.NoError(t, err)
	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: true,
	}, scope3)
}

func TestSavedQueriesPermissions_CreateSavedQueryPermissionsBatchBadDataError(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		user1   = createUser(t, dbInst, userPrincipal)
		user2   = createUser(t, dbInst, user2Principal)
	)

	unknownUUID, _ := uuid.NewV4()

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user2.ID, unknownUUID)
	require.Error(t, err)

	// verify partial share doesn't happen
	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
	require.NoError(t, err)
	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: false,
	}, scope)

	scope2, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, unknownUUID)
	require.NoError(t, err)
	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: false,
	}, scope2)
}

func TestSavedQueriesPermissions_GetScopeForSavedQueryPublic(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		user1   = createUser(t, dbInst, userPrincipal)
		user2   = createUser(t, dbInst, user2Principal)
	)

	query, err := dbInst.CreateSavedQuery(testCtx, user2.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToPublic(testCtx, query.ID)
	require.NoError(t, err)

	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: true,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: false,
	}, scope)
}

func TestSavedQueriesPermissions_GetScopeForSavedQueryShared(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		user1   = createUser(t, dbInst, userPrincipal)
		user2   = createUser(t, dbInst, user2Principal)
	)

	query, err := dbInst.CreateSavedQuery(testCtx, user2.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: true,
	}, scope)
}

func TestSavedQueriesPermissions_GetScopeForSavedQueryOwned(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		user1   = createUser(t, dbInst, userPrincipal)
		user2   = createUser(t, dbInst, user2Principal)
	)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user2.ID)
	require.NoError(t, err)

	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	require.Equal(t, database.SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  true,
		model.SavedQueryScopeShared: false,
	}, scope)
}

func TestSavedQueriesPermissions_DeleteSavedQueryPermissionsForUsers(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
		user1   = createUser(t, dbInst, userPrincipal)
		user2   = createUser(t, dbInst, user2Principal)
		user3   = createUser(t, dbInst, user3Principal)
	)

	t.Run("Deletes saved query permissions for user(s)", func(t *testing.T) {
		query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
		require.NoError(t, err)

		_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user2.ID, user3.ID)
		require.NoError(t, err)

		scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: true,
		}, scope)

		scope2, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user3.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: true,
		}, scope2)

		err = dbInst.DeleteSavedQueryPermissionsForUsers(testCtx, query.ID, user2.ID)
		require.NoError(t, err)

		scope3, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: false,
		}, scope3)

		scope4, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user3.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: true,
		}, scope4)
	})

	t.Run("Deletes saved query permissions given no provided users", func(t *testing.T) {
		query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query2", "TESTING2", "Example2")
		require.NoError(t, err)

		_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user2.ID)
		require.NoError(t, err)

		scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: true,
		}, scope)

		err = dbInst.DeleteSavedQueryPermissionsForUsers(testCtx, query.ID)
		require.NoError(t, err)

		scope2, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user2.ID)
		require.NoError(t, err)
		require.Equal(t, database.SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: false,
		}, scope2)
	})
}

func TestSavedQueriesPermissions_IsSavedQueryPublic(t *testing.T) {
	var (
		testCtx       = context.Background()
		dbInst, user1 = initAndCreateUser(t)
	)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToPublic(testCtx, query.ID)
	require.NoError(t, err)

	isPublic, err := dbInst.IsSavedQueryPublic(testCtx, query.ID)
	require.NoError(t, err)
	assert.True(t, isPublic)
}

func TestSavedQueriesPermissions_IsSavedQuerySharedToUser(t *testing.T) {
	var (
		testCtx       = context.Background()
		dbInst, user1 = initAndCreateUser(t)
	)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	isShared, err := dbInst.IsSavedQuerySharedToUser(testCtx, query.ID, user1.ID)
	require.NoError(t, err)
	assert.True(t, isShared)
}

func TestSavedQueriesPermissions_GetSavedQueryPermissions(t *testing.T) {
	var (
		testCtx                       = context.Background()
		dbInst, user1                 = initAndCreateUser(t)
		user2                         = createUser(t, dbInst, user2Principal)
		expectedSavedQueryPermissions = []model.SavedQueriesPermissions{{
			SharedToUserID: uuid.NullUUID{
				UUID:  user2.ID,
				Valid: true,
			},
			QueryID: 1,
			Public:  false,
			BigSerial: model.BigSerial{
				ID: 1,
			},
		}}
	)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Test Description")
	require.NoError(t, err)
	_, err = dbInst.CreateSavedQueryPermissionsToUsers(testCtx, query.ID, user2.ID)
	require.NoError(t, err)
	actualSavedQueryPermissions, err := dbInst.GetSavedQueryPermissions(testCtx, query.ID)
	require.NoError(t, err)
	assert.Equal(t, len(expectedSavedQueryPermissions), len(actualSavedQueryPermissions))
	for idx, _ := range expectedSavedQueryPermissions {
		expectedSavedQueryPermissions[idx].CreatedAt = actualSavedQueryPermissions[idx].CreatedAt
		expectedSavedQueryPermissions[idx].UpdatedAt = actualSavedQueryPermissions[idx].UpdatedAt
		assert.Equal(t, expectedSavedQueryPermissions[idx], actualSavedQueryPermissions[idx])
	}
}
