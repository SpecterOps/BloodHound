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
		autoCertify     = null.BoolFrom(false)
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

func TestDatabase_GetAssetGroupTagSelectorBySelectorId(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testActor       = "test_actor"
		testName        = "test selector name"
		testDescription = "test description"
		isDefault       = false
		allowDisable    = true
		autoCertify     = null.BoolFrom(false)
		testSeeds       = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectID5678"},
		}
	)

	selector, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, testActor, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)

	// test the read by ID function
	readBackSelector, err := dbInst.GetAssetGroupTagSelectorBySelectorId(testCtx, selector.ID)
	require.NoError(t, err)
	require.Equal(t, 1, readBackSelector.AssetGroupTagId)
	require.False(t, readBackSelector.CreatedAt.IsZero())
	require.Equal(t, testActor, readBackSelector.CreatedBy)
	require.False(t, readBackSelector.UpdatedAt.IsZero())
	require.Equal(t, testActor, readBackSelector.UpdatedBy)
	require.Empty(t, readBackSelector.DisabledAt)
	require.Empty(t, readBackSelector.DisabledBy)
	require.Equal(t, testName, readBackSelector.Name)
	require.Equal(t, testDescription, readBackSelector.Description)
	require.Equal(t, autoCertify, readBackSelector.AutoCertify)
	require.Equal(t, isDefault, readBackSelector.IsDefault)
	for idx, seed := range testSeeds {
		require.Equal(t, seed.Type, readBackSelector.Seeds[idx].Type)
		require.Equal(t, seed.Value, readBackSelector.Seeds[idx].Value)
	}
}

func TestDatabase_UpdateAssetGroupTagSelector(t *testing.T) {
	var (
		dbInst            = integration.SetupDB(t)
		testCtx           = context.Background()
		testActor         = "test_actor"
		updateActor       = "updated actor"
		testName          = "test selector name"
		updateName        = "updated name"
		testDescription   = "test description"
		updateDescription = "updated description"
		isDefault         = false
		allowDisable      = true
		autoCertify       = null.BoolFrom(false)
		updateAutoCert    = null.BoolFrom(true)
		disabledTime      = null.TimeFrom(time.Date(2025, time.March, 25, 12, 0, 0, 0, time.UTC))
		testSeeds         = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectID5678"},
		}
		updateSeeds = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectIDUpdated1"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectIDUpdated2"},
		}
	)

	selector, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, testActor, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)

	selector.UpdatedBy = updateActor
	selector.Name = updateName
	selector.Description = updateDescription
	selector.DisabledAt = disabledTime
	selector.DisabledBy = null.StringFrom(updateActor)
	selector.AutoCertify = updateAutoCert
	selector.Seeds = updateSeeds

	// call the update function
	_, err = dbInst.UpdateAssetGroupTagSelector(testCtx, selector)
	require.NoError(t, err)

	readBackSelector, err := dbInst.GetAssetGroupTagSelectorBySelectorId(testCtx, selector.ID)
	require.NoError(t, err)
	require.Equal(t, selector.AssetGroupTagId, readBackSelector.AssetGroupTagId)      // should be unchanged
	require.False(t, readBackSelector.CreatedAt.IsZero())                             // should be unchanged
	require.Equal(t, testActor, readBackSelector.CreatedBy)                           // should be unchanged
	require.False(t, readBackSelector.UpdatedAt.IsZero())                             // should be updated
	require.Equal(t, updateActor, readBackSelector.UpdatedBy)                         // should be updated
	require.Equal(t, disabledTime.Time.UTC(), readBackSelector.DisabledAt.Time.UTC()) // should be updated
	require.Equal(t, null.StringFrom(updateActor), readBackSelector.DisabledBy)       // should be updated
	require.Equal(t, updateName, readBackSelector.Name)
	require.Equal(t, updateDescription, readBackSelector.Description)
	require.Equal(t, updateAutoCert, readBackSelector.AutoCertify)
	require.Equal(t, isDefault, readBackSelector.IsDefault)
	for idx, seed := range updateSeeds {
		require.Equal(t, seed.Type, readBackSelector.Seeds[idx].Type)
		require.Equal(t, seed.Value, readBackSelector.Seeds[idx].Value)
	}
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
		dbInst        = integration.SetupDB(t)
		testCtx       = context.Background()
		isDefault     = false
		allowDisable  = true
		autoCertify   = null.BoolFrom(false)
		test1Selector = model.AssetGroupTagSelector{
			Name:            "test selector name",
			Description:     "test description",
			AssetGroupTagId: 1,
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			},
			AllowDisable: true,
			AutoCertify:  autoCertify,
		}
		test2Selector = model.AssetGroupTagSelector{
			Name:            "test2 selector name",
			Description:     "test2 description",
			AssetGroupTagId: 1,
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
			},
			AllowDisable: true,
			AutoCertify:  autoCertify,
		}
	)

	_, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, "id", test1Selector.Name, test1Selector.Description, isDefault, allowDisable, autoCertify, test1Selector.Seeds)
	require.NoError(t, err)
	created_at := time.Now()
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, 1, "id2", test2Selector.Name, test2Selector.Description, isDefault, allowDisable, autoCertify, test2Selector.Seeds)
	require.NoError(t, err)

	t.Run("successfully returns an array of selectors, no filters", func(t *testing.T) {
		results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{}, model.SQLFilter{})
		require.NoError(t, err)

		require.Equal(t, 2, len(results))
		require.Equal(t, test1Selector.Name, results[0].Name)
		require.Equal(t, test2Selector.Name, results[1].Name)
		require.Equal(t, test1Selector.Description, results[0].Description)
		require.Equal(t, test2Selector.Description, results[1].Description)
		require.Equal(t, test1Selector.AssetGroupTagId, results[0].AssetGroupTagId)
		require.Equal(t, test2Selector.AssetGroupTagId, results[1].AssetGroupTagId)
		require.Equal(t, test1Selector.IsDefault, results[0].IsDefault)
		require.Equal(t, test2Selector.IsDefault, results[1].IsDefault)
		require.Equal(t, test1Selector.AllowDisable, results[0].AllowDisable)
		require.Equal(t, test2Selector.AllowDisable, results[1].AllowDisable)
		require.Equal(t, test1Selector.AutoCertify, results[0].AutoCertify)
		require.Equal(t, test2Selector.AutoCertify, results[1].AutoCertify)

		for idx, seed := range test1Selector.Seeds {
			require.Equal(t, seed.Type, results[0].Seeds[idx].Type)
			require.Equal(t, seed.Value, results[0].Seeds[idx].Value)
		}

		for idx, seed := range test2Selector.Seeds {
			require.Equal(t, seed.Type, results[1].Seeds[idx].Type)
			require.Equal(t, seed.Value, results[1].Seeds[idx].Value)
		}

	})

	t.Run("successfully returns an array of seed selector filters", func(t *testing.T) {
		results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{}, model.SQLFilter{SQLString: "type = ?", Params: []any{2}})
		require.NoError(t, err)

		require.Equal(t, 1, len(results))
		for idx, seed := range test2Selector.Seeds {
			require.Equal(t, seed.Type, results[0].Seeds[idx].Type)
			require.Equal(t, seed.Value, results[0].Seeds[idx].Value)
		}
	})

	t.Run("successfully returns an array of selector filters", func(t *testing.T) {
		results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{SQLString: "created_at >= ?", Params: []any{created_at}}, model.SQLFilter{})
		require.NoError(t, err)

		require.Equal(t, 1, len(results))
		require.True(t, results[0].CreatedAt.After(created_at))

	})

}
