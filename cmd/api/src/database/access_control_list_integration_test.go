// Copyright 2025 Specter Ops, Inc.
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
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_AccessControlList(t *testing.T) {
	t.Parallel()

	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	newUser, err := suite.BHDatabase.CreateUser(context.Background(), model.User{
		FirstName:       null.StringFrom("First"),
		LastName:        null.StringFrom("Last"),
		EmailAddress:    null.StringFrom(userPrincipal),
		PrincipalName:   userPrincipal,
		AllEnvironments: true,
	})

	require.NoError(t, err)

	t.Run("Updating ACL disables AllEnvironments", func(t *testing.T) {
		err := suite.BHDatabase.UpdateUser(suite.Context, model.User{
			Unique: model.Unique{
				ID: newUser.ID,
			},
			EnvironmentAccessControl: []model.EnvironmentAccess{
				{
					Environment: "12345",
				},
				{
					Environment: "54321"},
			},
		})
		require.NoError(t, err)
		updatedUser, err := suite.BHDatabase.GetUser(suite.Context, newUser.ID)
		require.NoError(t, err)
		assert.False(t, updatedUser.AllEnvironments)
	})

	t.Run("GetEnvironmentAccessListForUser", func(t *testing.T) {
		result, err := suite.BHDatabase.GetEnvironmentAccessListForUser(suite.Context, newUser)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("Deleting User Removes ACL", func(t *testing.T) {
		err := suite.BHDatabase.DeleteUser(suite.Context, newUser)
		require.NoError(t, err)

		result, err := suite.BHDatabase.GetEnvironmentAccessListForUser(suite.Context, newUser)
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}
