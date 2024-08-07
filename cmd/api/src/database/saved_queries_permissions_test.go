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
	uuid2 "github.com/gofrs/uuid"
	"github.com/google/uuid"
	"testing"

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSavedQueriesPermissions_SharingToUser(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	user2, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: user2Principal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user.ID, "Test Query", "MATCH(n) RETURN n", "An example Query")
	require.NoError(t, err)

	permissions, err := dbInst.CreateSavedQueryPermissionToUser(testCtx, query.ID, user2.ID)
	require.NoError(t, err)

	assert.Equal(t, database.NullUUID(user2.ID), permissions.SharedToUserID)
	assert.Equal(t, false, permissions.Public)
	assert.Equal(t, query.ID, permissions.QueryID)
}

func TestSavedQueriesPermissions_SharingToGlobal(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user.ID, "Test Query", "MATCH(n) RETURN n", "An example Query")
	require.NoError(t, err)

	permissions, err := dbInst.CreateSavedQueryPermissionToPublic(testCtx, query.ID)
	require.NoError(t, err)

	assert.Equal(t, true, permissions.Public)
	assert.Equal(t, query.ID, permissions.QueryID)
}

func TestSavedQueriesPermissions_CheckUserHasPermissionToSavedQuery(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	user2, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: user2Principal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToUser(testCtx, query.ID, user2.ID)
	require.NoError(t, err)

	result, err := dbInst.CheckUserHasPermissionToSavedQuery(testCtx, query.ID, user2.ID)
	require.NoError(t, err)
	assert.True(t, result)

	result, err = dbInst.CheckUserHasPermissionToSavedQuery(testCtx, query.ID, user.ID)
	require.NoError(t, err)
	assert.False(t, result)
}

func TestSavedQueriesPermissions_CreateSavedQueryPermissionsBatch(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	users := make([]model.User, 0)
	for i := 0; i < 5; i++ {
		user, err := dbInst.CreateUser(testCtx, model.User{
			PrincipalName: uuid.NewString(),
		})
		require.NoError(t, err)
		users = append(users, user)
	}

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	permissions := make([]model.SavedQueriesPermissions, 0)
	for _, user := range users {
		permissions = append(permissions, model.SavedQueriesPermissions{
			QueryID:        query.ID,
			Public:         false,
			SharedToUserID: database.NullUUID(user.ID),
		})
	}

	err = dbInst.CreateSavedQueryPermissionsBatch(testCtx, permissions)
	require.NoError(t, err)

	permissions, err = dbInst.GetPermissionsForSavedQuery(testCtx, query.ID)
	require.NoError(t, err)
	assert.Len(t, permissions, 5)
}

func TestSavedQueriesPermissions_CreateSavedQueryPermissionsBatchBadDataError(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	users := make([]model.User, 0)
	for i := 0; i < 5; i++ {
		user, err := dbInst.CreateUser(testCtx, model.User{
			PrincipalName: uuid.NewString(),
		})
		require.NoError(t, err)
		users = append(users, user)
	}

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	permissions := make([]model.SavedQueriesPermissions, 0)
	for _, user := range users {
		permissions = append(permissions, model.SavedQueriesPermissions{
			QueryID:        query.ID,
			Public:         false,
			SharedToUserID: database.NullUUID(user.ID),
		})
	}

	invalidUUID, _ := uuid2.NewV4()
	permissions[3].SharedToUserID = database.NullUUID(invalidUUID)

	err = dbInst.CreateSavedQueryPermissionsBatch(testCtx, permissions)
	require.Error(t, err)

	permissions, err = dbInst.GetPermissionsForSavedQuery(testCtx, query.ID)
	require.NoError(t, err)

	assert.Len(t, permissions, 0)
}

func TestSavedQueriesPermissions_GetScopeForSavedQueryPublic(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	user2, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: user2Principal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user2.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToPublic(testCtx, query.ID)
	require.NoError(t, err)

	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	require.Equal(t, database.SavedQueryScopeMap{
		database.SavedQueryScopePublic: true,
		database.SavedQueryScopeOwned:  false,
		database.SavedQueryScopeShared: false,
	}, scope)
}

func TestSavedQueriesPermissions_GetScopeForSavedQueryShared(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	user2, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: user2Principal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user2.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToUser(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	require.Equal(t, database.SavedQueryScopeMap{
		database.SavedQueryScopePublic: false,
		database.SavedQueryScopeOwned:  false,
		database.SavedQueryScopeShared: true,
	}, scope)
}

func TestSavedQueriesPermissions_GetScopeForSavedQueryOwned(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	user2, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: user2Principal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToUser(testCtx, query.ID, user2.ID)
	require.NoError(t, err)

	scope, err := dbInst.GetScopeForSavedQuery(testCtx, query.ID, user1.ID)
	require.NoError(t, err)

	require.Equal(t, database.SavedQueryScopeMap{
		database.SavedQueryScopePublic: false,
		database.SavedQueryScopeOwned:  true,
		database.SavedQueryScopeShared: false,
	}, scope)
}

func TestSavedQueriesPermissions_DeleteSavedQueryPermission(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToPublic(testCtx, query.ID)
	require.NoError(t, err)

	err = dbInst.DeleteSavedQueryPermission(testCtx, query.ID)
	require.NoError(t, err)

	permissions, err := dbInst.GetPermissionsForSavedQuery(testCtx, query.ID)
	require.NoError(t, err)
	assert.Len(t, permissions, 0)
}

func TestSavedQueriesPermissions_DeleteSavedQueryPermissionsForUser(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	user1, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: userPrincipal,
	})
	require.NoError(t, err)

	user2, err := dbInst.CreateUser(testCtx, model.User{
		PrincipalName: user2Principal,
	})
	require.NoError(t, err)

	query, err := dbInst.CreateSavedQuery(testCtx, user1.ID, "Test Query", "TESTING", "Example")
	require.NoError(t, err)

	_, err = dbInst.CreateSavedQueryPermissionToUser(testCtx, query.ID, user2.ID)
	require.NoError(t, err)

	err = dbInst.DeleteSavedQueryPermissionsForUser(testCtx, query.ID, user2.ID)
	require.NoError(t, err)

	hasPermission, err := dbInst.CheckUserHasPermissionToSavedQuery(testCtx, query.ID, user2.ID)
	require.NoError(t, err)
	require.False(t, hasPermission)
}
