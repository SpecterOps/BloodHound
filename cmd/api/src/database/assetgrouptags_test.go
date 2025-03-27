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
	"time"

	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateAssetGroupTagSelector(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testActor       = "test_actor"
		testName        = "test selector name"
		testDescription = "test description"
		isDefault       = false
		allowDisable    = true
		autoCertify     = false
		testSeeds       = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectID5678"},
		}
	)

	selector, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, testActor, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)
	require.Equal(t, 1, selector.AssetGroupTagId)
	require.False(t, selector.CreatedAt.IsZero())
	require.Equal(t, testActor, selector.CreatedBy)
	require.False(t, selector.UpdatedAt.IsZero())
	require.Equal(t, testActor, selector.UpdatedBy)
	require.Empty(t, selector.DisabledAt)
	require.Empty(t, selector.DisabledBy)
	require.Equal(t, testName, selector.Name)
	require.Equal(t, testDescription, selector.Description)
	require.Equal(t, autoCertify, selector.AutoCertify)
	require.Equal(t, isDefault, selector.IsDefault)

	for idx, seed := range testSeeds {
		require.Equal(t, seed.Type, selector.Seeds[idx].Type)
		require.Equal(t, seed.Value, selector.Seeds[idx].Value)
	}

	history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, model.AssetGroupHistoryActionCreateSelector, history[0].Action)
}

func TestDatabase_CreateAssetGroupTag(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		tagType         = model.AssetGroupTagTypeTier
		testActor       = "test_actor"
		testName        = "test tag name"
		testDescription = "test tag description"
		position        = null.Int32{}
		requireCertify  = null.Bool{}
	)

	t.Run("successfully creates tag", func(t *testing.T) {
		tag, err := dbInst.CreateAssetGroupTag(testCtx, tagType, testActor, testName, testDescription, position, requireCertify)
		require.NoError(t, err)
		require.Equal(t, tagType, tag.Type)
		require.False(t, tag.CreatedAt.IsZero())
		require.Equal(t, testActor, tag.CreatedBy)
		require.False(t, tag.UpdatedAt.IsZero())
		require.Equal(t, testActor, tag.UpdatedBy)
		require.Empty(t, tag.DeletedAt)
		require.Empty(t, tag.DeletedBy)
		require.Equal(t, testName, tag.Name)
		require.Equal(t, testDescription, tag.Description)
		require.Equal(t, null.Int32{}, tag.Position)
		require.Equal(t, null.Bool{}, tag.RequireCertify)

		tag, err = dbInst.GetAssetGroupTag(testCtx, tag.ID)
		require.NoError(t, err)
		require.Equal(t, tagType, tag.Type)
		require.False(t, tag.CreatedAt.IsZero())
		require.Equal(t, testActor, tag.CreatedBy)
		require.False(t, tag.UpdatedAt.IsZero())
		require.Equal(t, testActor, tag.UpdatedBy)
		require.Empty(t, tag.DeletedAt)
		require.Empty(t, tag.DeletedBy)
		require.Equal(t, testName, tag.Name)
		require.Equal(t, testDescription, tag.Description)
		require.Equal(t, null.Int32{}, tag.Position)
		require.Equal(t, null.Bool{}, tag.RequireCertify)

		// verify history record was also created
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.Len(t, history, 1)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
	})

	t.Run("Non existant tag errors out", func(t *testing.T) {
		_, err := dbInst.GetAssetGroupTag(testCtx, 1234)
		require.Error(t, err)
	})

}

func TestDatabase_GetAssetGroupTagSelectors(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testID          = "test_id"
		testName        = "test selector name"
		testDescription = "test description"
		isDefault       = false
		allowDisable    = true
		autoCertify     = false
		testSeeds       = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectID5678"},
		}
		testSeeds2 = []model.SelectorSeed{
			{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
		}
		test2ID          = "test2_id"
		test2Name        = "test2 selector name"
		test2Description = "test2 description"
		test3ID          = "test3_id"
		test3Name        = "test3 selector name"
		test3Description = "test3 description"
		created_at       = time.Now()
	)

	selector, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, testID, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)
	selector2, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, test2ID, test2Name, test2Description, isDefault, allowDisable, autoCertify, testSeeds2)
	require.NoError(t, err)
	selector3, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, test3ID, test3Name, test3Description, isDefault, allowDisable, autoCertify, testSeeds2)
	require.NoError(t, err)

	t.Run("successfully returns an array of selectors, no filters", func(t *testing.T) {
		results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{}, model.SQLFilter{})
		require.NoError(t, err)

		expected := []model.AssetGroupTagSelector{selector, selector2}
		require.Equal(t, expected, results)
	})

	t.Run("successfully returns an array of seed selector filters", func(t *testing.T) {
		results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{}, model.SQLFilter{SQLString: "type = ?", Params: []any{2}})
		require.NoError(t, err)

		expected := []model.AssetGroupTagSelector{selector2}
		require.Equal(t, expected, results)
	})

	t.Run("successfully returns an array of selector filters and seed filters", func(t *testing.T) {
		results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{SQLString: "created_at >= ?", Params: []any{created_at}}, model.SQLFilter{SQLString: "type = ?", Params: []any{2}})
		require.NoError(t, err)

		expected := []model.AssetGroupTagSelector{selector2, selector3}
		require.Equal(t, len(expected), len(results))

		require.Equal(t, test2Name, results[0].Name)
		require.Equal(t, test3Name, results[1].Name)
		require.Equal(t, test2Description, results[0].Description)
		require.Equal(t, test3Description, results[1].Description)
		require.Equal(t, test2ID, results[0].CreatedBy)
		require.Equal(t, test3ID, results[1].CreatedBy)
		require.True(t, results[0].AllowDisable)
		require.True(t, results[1].AllowDisable)
		require.False(t, results[0].AutoCertify)
		require.False(t, results[1].AutoCertify)
		require.False(t, results[0].IsDefault)
		require.False(t, results[1].IsDefault)
		require.False(t, results[0].CreatedAt.IsZero())
		require.False(t, results[1].CreatedAt.IsZero())
		require.False(t, results[0].UpdatedAt.IsZero())
		require.False(t, results[1].UpdatedAt.IsZero())
	})

}
