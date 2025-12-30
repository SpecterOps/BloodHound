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
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestTransaction_Commit(t *testing.T) {
	suite := setupIntegrationTestSuite(t)

	var createdGroup model.AssetGroup
	err := suite.BHDatabase.Transaction(suite.Context, func(tx *database.BloodhoundDB) error {
		var err error
		createdGroup, err = tx.CreateAssetGroup(suite.Context, "test-transaction-group", "test-tx-tag", false)
		return err
	})

	require.NoError(t, err)
	require.NotZero(t, createdGroup.ID)

	fetchedGroup, err := suite.BHDatabase.GetAssetGroup(suite.Context, createdGroup.ID)
	require.NoError(t, err)
	require.Equal(t, "test-transaction-group", fetchedGroup.Name)
}

func TestTransaction_Rollback(t *testing.T) {
	suite := setupIntegrationTestSuite(t)

	rollbackErr := errors.New("intentional rollback")
	var attemptedGroupID int32

	err := suite.BHDatabase.Transaction(suite.Context, func(tx *database.BloodhoundDB) error {
		createdGroup, err := tx.CreateAssetGroup(suite.Context, "should-rollback-group", "rollback-tag", false)
		if err != nil {
			return err
		}
		attemptedGroupID = createdGroup.ID
		return rollbackErr
	})

	require.ErrorIs(t, err, rollbackErr)

	_, err = suite.BHDatabase.GetAssetGroup(suite.Context, attemptedGroupID)
	require.ErrorIs(t, err, database.ErrNotFound)
}

func TestTransaction_MultipleOperations(t *testing.T) {
	suite := setupIntegrationTestSuite(t)

	var group1, group2 model.AssetGroup
	err := suite.BHDatabase.Transaction(suite.Context, func(tx *database.BloodhoundDB) error {
		var err error
		group1, err = tx.CreateAssetGroup(suite.Context, "multi-op-group-1", "multi-tag-1", false)
		if err != nil {
			return err
		}

		group2, err = tx.CreateAssetGroup(suite.Context, "multi-op-group-2", "multi-tag-2", false)
		return err
	})

	require.NoError(t, err)

	fetchedGroup1, err := suite.BHDatabase.GetAssetGroup(suite.Context, group1.ID)
	require.NoError(t, err)
	require.Equal(t, "multi-op-group-1", fetchedGroup1.Name)

	fetchedGroup2, err := suite.BHDatabase.GetAssetGroup(suite.Context, group2.ID)
	require.NoError(t, err)
	require.Equal(t, "multi-op-group-2", fetchedGroup2.Name)
}

func TestTransaction_PartialRollback(t *testing.T) {
	suite := setupIntegrationTestSuite(t)

	outsideGroup, err := suite.BHDatabase.CreateAssetGroup(suite.Context, "outside-tx-group", "outside-tag", false)
	require.NoError(t, err)

	rollbackErr := errors.New("partial rollback")
	err = suite.BHDatabase.Transaction(suite.Context, func(tx *database.BloodhoundDB) error {
		_, err := tx.CreateAssetGroup(suite.Context, "inside-tx-group", "inside-tag", false)
		if err != nil {
			return err
		}
		return rollbackErr
	})

	require.ErrorIs(t, err, rollbackErr)

	fetchedOutside, err := suite.BHDatabase.GetAssetGroup(suite.Context, outsideGroup.ID)
	require.NoError(t, err)
	require.Equal(t, "outside-tx-group", fetchedOutside.Name)
}

func TestTransaction_ContextCancellation(t *testing.T) {
	suite := setupIntegrationTestSuite(t)

	ctx, cancel := context.WithCancel(suite.Context)

	err := suite.BHDatabase.Transaction(ctx, func(tx *database.BloodhoundDB) error {
		cancel()
		_, err := tx.CreateAssetGroup(ctx, "cancelled-group", "cancelled-tag", false)
		return err
	})

	require.Error(t, err)
}

func TestTransaction_NestedTransactions(t *testing.T) {
	suite := setupIntegrationTestSuite(t)

	var outerGroup, innerGroup model.AssetGroup
	err := suite.BHDatabase.Transaction(suite.Context, func(outerTx *database.BloodhoundDB) error {
		var err error
		outerGroup, err = outerTx.CreateAssetGroup(suite.Context, "nested-outer-group", "nested-outer-tag", false)
		if err != nil {
			return err
		}

		return outerTx.Transaction(suite.Context, func(innerTx *database.BloodhoundDB) error {
			innerGroup, err = innerTx.CreateAssetGroup(suite.Context, "nested-inner-group", "nested-inner-tag", false)
			return err
		})
	})

	require.NoError(t, err)

	fetchedOuter, err := suite.BHDatabase.GetAssetGroup(suite.Context, outerGroup.ID)
	require.NoError(t, err)
	require.Equal(t, "nested-outer-group", fetchedOuter.Name)

	fetchedInner, err := suite.BHDatabase.GetAssetGroup(suite.Context, innerGroup.ID)
	require.NoError(t, err)
	require.Equal(t, "nested-inner-group", fetchedInner.Name)
}

func TestTransaction_NestedRollback(t *testing.T) {
	suite := setupIntegrationTestSuite(t)

	var outerGroupID int32
	nestedErr := errors.New("nested transaction failure")

	err := suite.BHDatabase.Transaction(suite.Context, func(outerTx *database.BloodhoundDB) error {
		outerGroup, err := outerTx.CreateAssetGroup(suite.Context, "nested-rollback-outer", "nested-rb-outer-tag", false)
		if err != nil {
			return err
		}
		outerGroupID = outerGroup.ID

		innerErr := outerTx.Transaction(suite.Context, func(innerTx *database.BloodhoundDB) error {
			_, err := innerTx.CreateAssetGroup(suite.Context, "nested-rollback-inner", "nested-rb-inner-tag", false)
			if err != nil {
				return err
			}
			return nestedErr
		})

		return innerErr
	})

	require.ErrorIs(t, err, nestedErr)

	_, err = suite.BHDatabase.GetAssetGroup(suite.Context, outerGroupID)
	require.ErrorIs(t, err, database.ErrNotFound)
}
