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
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
