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
	"log"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateAssetGroupTagSelector(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testActor       = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
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
	require.Equal(t, testActor.ID.String(), selector.CreatedBy)
	require.False(t, selector.UpdatedAt.IsZero())
	require.Equal(t, testActor.ID.String(), selector.UpdatedBy)
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
		testActor       = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
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
	require.Equal(t, testActor.ID.String(), readBackSelector.CreatedBy)
	require.False(t, readBackSelector.UpdatedAt.IsZero())
	require.Equal(t, testActor.ID.String(), readBackSelector.UpdatedBy)
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
		testActor         = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("11111111-9012-4567-9012-456789012345")}}
		updateActor       = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("22222222-9012-4567-9012-456789012345")}}
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

	selector.Name = updateName
	selector.Description = updateDescription
	selector.DisabledAt = disabledTime
	selector.DisabledBy = null.StringFrom(updateActor.ID.String())
	selector.AutoCertify = updateAutoCert
	selector.Seeds = updateSeeds

	// call the update function
	_, err = dbInst.UpdateAssetGroupTagSelector(testCtx, updateActor.ID.String(), updateActor.EmailAddress.ValueOrZero(), selector)
	require.NoError(t, err)

	readBackSelector, err := dbInst.GetAssetGroupTagSelectorBySelectorId(testCtx, selector.ID)
	require.NoError(t, err)
	require.Equal(t, selector.AssetGroupTagId, readBackSelector.AssetGroupTagId)            // should be unchanged
	require.False(t, readBackSelector.CreatedAt.IsZero())                                   // should be unchanged
	require.Equal(t, testActor.ID.String(), readBackSelector.CreatedBy)                     // should be unchanged
	require.False(t, readBackSelector.UpdatedAt.IsZero())                                   // should be updated
	require.Equal(t, updateActor.ID.String(), readBackSelector.UpdatedBy)                   // should be updated
	require.Equal(t, disabledTime.Time.UTC(), readBackSelector.DisabledAt.Time.UTC())       // should be updated
	require.Equal(t, null.StringFrom(updateActor.ID.String()), readBackSelector.DisabledBy) // should be updated
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
		testCtx             = context.Background()
		tagTypeTier         = model.AssetGroupTagTypeTier
		tagTypeLabel        = model.AssetGroupTagTypeLabel
		testActor           = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
		testName            = "test tag name"
		testName2           = "test2 tag name"
		testName3           = "test3 tag name"
		testName4           = "test4 tag name"
		testDescription     = "test tag description"
		position            = null.Int32{}
		requireCertifyTier  = null.BoolFrom(true)
		requireCertifyLabel = null.Bool{}
	)

	t.Run("successfully creates tier", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		tag, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeTier, testActor, testName, testDescription, position, requireCertifyTier)
		require.NoError(t, err)
		require.Equal(t, tagTypeTier, tag.Type)
		require.False(t, tag.CreatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.CreatedBy)
		require.False(t, tag.UpdatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.UpdatedBy)
		require.Empty(t, tag.DeletedAt)
		require.Empty(t, tag.DeletedBy)
		require.Equal(t, testName, tag.Name)
		require.Equal(t, testDescription, tag.Description)
		require.Equal(t, null.Int32From(2), tag.Position)
		require.Equal(t, null.BoolFrom(true), tag.RequireCertify)
		require.Equal(t, null.BoolFrom(false), tag.AnalysisEnabled)

		tag, err = dbInst.GetAssetGroupTag(testCtx, tag.ID)
		require.NoError(t, err)
		require.Equal(t, tagTypeTier, tag.Type)
		require.False(t, tag.CreatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.CreatedBy)
		require.False(t, tag.UpdatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.UpdatedBy)
		require.Empty(t, tag.DeletedAt)
		require.Empty(t, tag.DeletedBy)
		require.Equal(t, testName, tag.Name)
		require.Equal(t, testDescription, tag.Description)
		require.Equal(t, null.Int32From(2), tag.Position)
		require.Equal(t, null.BoolFrom(true), tag.RequireCertify)
		require.Equal(t, null.BoolFrom(false), tag.AnalysisEnabled)

		// verify history record was also created
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.Len(t, history, 1)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
	})

	t.Run("successfully creates label", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		tag, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeLabel, testActor, testName2, testDescription, position, requireCertifyLabel)
		require.NoError(t, err)
		require.Equal(t, tagTypeLabel, tag.Type)
		require.False(t, tag.CreatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.CreatedBy)
		require.False(t, tag.UpdatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.UpdatedBy)
		require.Empty(t, tag.DeletedAt)
		require.Empty(t, tag.DeletedBy)
		require.Equal(t, testName2, tag.Name)
		require.Equal(t, testDescription, tag.Description)
		require.Equal(t, null.Int32{}, tag.Position)
		require.Equal(t, null.Bool{}, tag.RequireCertify)
		require.Equal(t, null.Bool{}, tag.AnalysisEnabled)

		tag, err = dbInst.GetAssetGroupTag(testCtx, tag.ID)
		require.NoError(t, err)
		require.Equal(t, tagTypeLabel, tag.Type)
		require.False(t, tag.CreatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.CreatedBy)
		require.False(t, tag.UpdatedAt.IsZero())
		require.Equal(t, testActor.ID.String(), tag.UpdatedBy)
		require.Empty(t, tag.DeletedAt)
		require.Empty(t, tag.DeletedBy)
		require.Equal(t, testName2, tag.Name)
		require.Equal(t, testDescription, tag.Description)
		require.Equal(t, null.Int32{}, tag.Position)
		require.Equal(t, null.Bool{}, tag.RequireCertify)
		require.Equal(t, null.Bool{}, tag.AnalysisEnabled)
	})

	t.Run("successfully creates and shifts tiers", func(t *testing.T) {
		var (
			position2 = null.Int32From(2)
			position3 = null.Int32From(3)
			dbInst    = integration.SetupDB(t)
		)

		tag, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeTier, testActor, testName3, testDescription, position2, requireCertifyTier)
		require.NoError(t, err)

		tag, err = dbInst.GetAssetGroupTag(testCtx, tag.ID)
		require.NoError(t, err)
		require.Equal(t, position2, tag.Position)

		// Create another tag at an existing position, forcing the first tag down
		tag2, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeTier, testActor, testName4, testDescription, position2, requireCertifyTier)
		require.NoError(t, err)

		tag2, err = dbInst.GetAssetGroupTag(testCtx, tag2.ID)
		require.NoError(t, err)
		require.Equal(t, position2, tag2.Position)

		tag, err = dbInst.GetAssetGroupTag(testCtx, tag.ID)
		require.NoError(t, err)
		require.Equal(t, position3, tag.Position)

		// verify history record was also created and shifted
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.Len(t, history, 3)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[1].Action)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[2].Action)
	})

	t.Run("Non existant tag errors out", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		_, err := dbInst.GetAssetGroupTag(testCtx, 1234)
		require.Error(t, err)
	})

}

func TestDatabase_UpdateAssetGroupTag(t *testing.T) {
	var (
		testCtx   = context.Background()
		testActor = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
	)

	t.Run("updates regular fields successfully", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		origTier, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeTier,
			testActor,
			"test tier orig",
			"orig desc",
			null.Int32From(2),
			null.BoolFrom(false),
		)
		require.NoError(t, err)

		toUpdate := origTier
		toUpdate.Name = "updated name"
		toUpdate.Description = "updated desc"
		toUpdate.RequireCertify.SetValid(true)

		updatedTier, err := dbInst.UpdateAssetGroupTag(testCtx, testActor, toUpdate)
		require.NoError(t, err)

		require.Equal(t, toUpdate.Name, updatedTier.Name)
		require.Equal(t, toUpdate.Description, updatedTier.Description)
		require.Equal(t, toUpdate.RequireCertify, updatedTier.RequireCertify)

		// verify history records were created
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.Len(t, history, 2)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[1].Action)
	})

	t.Run("returns error for duplicate name", func(t *testing.T) {
		dbInst := integration.SetupDB(t)
		_, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeTier,
			testActor,
			"unique test",
			"",
			null.Int32From(2),
			null.BoolFrom(false),
		)
		require.NoError(t, err)
		origTier, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeTier,
			testActor,
			"orig",
			"",
			null.Int32From(3),
			null.BoolFrom(false),
		)
		require.NoError(t, err)

		toUpdate := origTier
		toUpdate.Name = "unique test"

		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, toUpdate)
		require.ErrorIs(t, err, database.ErrDuplicateAGName)
	})

	t.Run("requires position set for tier", func(t *testing.T) {
		dbInst := integration.SetupDB(t)
		tier, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeTier,
			testActor,
			"tier",
			"",
			null.Int32From(2),
			null.BoolFrom(false),
		)
		require.NoError(t, err)

		tier.Position = null.Int32{}

		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, tier)
		require.ErrorContains(t, err, "position is required for an existing tier")
	})

	t.Run("position is invalid for non-tier", func(t *testing.T) {
		dbInst := integration.SetupDB(t)
		tag, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeLabel,
			testActor,
			"tag1",
			"",
			null.Int32{},
			null.Bool{},
		)
		require.NoError(t, err)

		tag.Position = null.Int32From(2)

		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, tag)
		require.ErrorContains(t, err, "position, require_certify, and analysis_enabled are limited to tiers only")
	})

	t.Run("require_certify is invalid for non-tier", func(t *testing.T) {
		dbInst := integration.SetupDB(t)
		tag, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeLabel,
			testActor,
			"tag2",
			"",
			null.Int32{},
			null.Bool{},
		)
		require.NoError(t, err)

		tag.RequireCertify = null.BoolFrom(false)

		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, tag)
		require.ErrorContains(t, err, "position, require_certify, and analysis_enabled are limited to tiers only")
	})

	t.Run("blocks updating to position 1", func(t *testing.T) {
		dbInst := integration.SetupDB(t)
		tag, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeTier,
			testActor,
			"tagpos1",
			"",
			null.Int32From(2),
			null.Bool{},
		)
		require.NoError(t, err)

		tag.Position = null.Int32From(1)

		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, tag)
		require.ErrorIs(t, err, database.ErrPositionOutOfRange)
	})

	t.Run("blocks updating to invalid positions", func(t *testing.T) {
		dbInst := integration.SetupDB(t)
		tag, err := dbInst.CreateAssetGroupTag(
			testCtx,
			model.AssetGroupTagTypeTier,
			testActor,
			"tagposinv",
			"",
			null.Int32From(2),
			null.Bool{},
		)
		require.NoError(t, err)

		tag.Position = null.Int32From(50)
		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, tag)
		require.ErrorIs(t, err, database.ErrPositionOutOfRange)

		tag.Position = null.Int32From(0)
		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, tag)
		require.ErrorIs(t, err, database.ErrPositionOutOfRange)

		tag.Position = null.Int32From(-1)
		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, tag)
		require.ErrorIs(t, err, database.ErrPositionOutOfRange)
	})
}

func TestDatabase_UpdateAssetGroupTag_shifting(t *testing.T) {
	var (
		testCtx   = context.Background()
		testActor = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
	)

	t.Run("shifts tier higher successfully", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		tag1, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift1", "", null.Int32From(2), null.Bool{})
		require.NoError(t, err)
		tag2, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift2", "", null.Int32From(3), null.Bool{})
		require.NoError(t, err)
		tag3, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift3", "", null.Int32From(4), null.Bool{})
		require.NoError(t, err)
		tag4, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift4", "", null.Int32From(5), null.Bool{})
		require.NoError(t, err)

		toUpdate := tag3
		toUpdate.Position = null.Int32From(3)
		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, toUpdate)
		require.NoError(t, err)

		// load new positions
		updated1, err := dbInst.GetAssetGroupTag(testCtx, tag1.ID)
		require.NoError(t, err)
		updated2, err := dbInst.GetAssetGroupTag(testCtx, tag2.ID)
		require.NoError(t, err)
		updated3, err := dbInst.GetAssetGroupTag(testCtx, tag3.ID)
		require.NoError(t, err)
		updated4, err := dbInst.GetAssetGroupTag(testCtx, tag4.ID)
		require.NoError(t, err)
		log.Printf("shiftup: before: %v", []int32{tag1.Position.ValueOrZero(), tag2.Position.ValueOrZero(), tag3.Position.ValueOrZero(), tag4.Position.ValueOrZero()})
		log.Printf("shiftup: after: %v", []int32{updated1.Position.ValueOrZero(), updated2.Position.ValueOrZero(), updated3.Position.ValueOrZero(), updated4.Position.ValueOrZero()})

		// verify positions
		require.Greater(t,
			updated3.Position.ValueOrZero(),
			updated1.Position.ValueOrZero(),
		)
		require.Greater(t,
			updated2.Position.ValueOrZero(),
			updated3.Position.ValueOrZero(),
		)
		require.Greater(t,
			updated4.Position.ValueOrZero(),
			updated2.Position.ValueOrZero(),
		)

		// verify history records were created
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(history), 2)
		n := len(history) - 2
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[n+0].Action)
		require.Equal(t, history[n+0].AssetGroupTagId, tag2.ID)
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[n+1].Action)
		require.Equal(t, history[n+1].AssetGroupTagId, tag3.ID)
	})

	t.Run("shifts tier lower successfully", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		tag1, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift1", "", null.Int32From(2), null.Bool{})
		require.NoError(t, err)
		tag2, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift2", "", null.Int32From(3), null.Bool{})
		require.NoError(t, err)
		tag3, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift3", "", null.Int32From(4), null.Bool{})
		require.NoError(t, err)
		tag4, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift4", "", null.Int32From(5), null.Bool{})
		require.NoError(t, err)

		toUpdate := tag1
		toUpdate.Position = null.Int32From(3)
		_, err = dbInst.UpdateAssetGroupTag(testCtx, testActor, toUpdate)
		require.NoError(t, err)

		// load new positions
		updated1, err := dbInst.GetAssetGroupTag(testCtx, tag1.ID)
		require.NoError(t, err)
		updated2, err := dbInst.GetAssetGroupTag(testCtx, tag2.ID)
		require.NoError(t, err)
		updated3, err := dbInst.GetAssetGroupTag(testCtx, tag3.ID)
		require.NoError(t, err)
		updated4, err := dbInst.GetAssetGroupTag(testCtx, tag4.ID)
		require.NoError(t, err)
		log.Printf("shiftdown: before: %v", []int32{tag1.Position.ValueOrZero(), tag2.Position.ValueOrZero(), tag3.Position.ValueOrZero(), tag4.Position.ValueOrZero()})
		log.Printf("shiftdown: after: %v", []int32{updated1.Position.ValueOrZero(), updated2.Position.ValueOrZero(), updated3.Position.ValueOrZero(), updated4.Position.ValueOrZero()})

		// verify positions
		require.Greater(t,
			updated1.Position.ValueOrZero(),
			updated2.Position.ValueOrZero(),
		)
		require.Greater(t,
			updated3.Position.ValueOrZero(),
			updated1.Position.ValueOrZero(),
		)
		require.Greater(t,
			updated4.Position.ValueOrZero(),
			updated3.Position.ValueOrZero(),
		)

		// verify history records were created
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(history), 2)
		n := len(history) - 2
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[n+0].Action)
		require.Equal(t, history[n+0].AssetGroupTagId, tag2.ID)
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[n+1].Action)
		require.Equal(t, history[n+1].AssetGroupTagId, tag1.ID)
	})
}

func TestDatabase_DeleteAssetGroupTag(t *testing.T) {
	var (
		testCtx            = context.Background()
		userID, _          = uuid.NewV4()
		userID2, _         = uuid.NewV4()
		userID4, _         = uuid.NewV4()
		testUser           = model.User{Unique: model.Unique{ID: userID}}
		testUser2          = model.User{Unique: model.Unique{ID: userID2}}
		testUser4          = model.User{Unique: model.Unique{ID: userID4}}
		testName           = "test tag name"
		testName2          = "test tag name2"
		testName4          = "test tag name4"
		testDescription    = "test tag description"
		testDescription2   = "test tag description2"
		testDescription4   = "test tag description4"
		position           = null.Int32From(2)
		position2          = null.Int32From(3)
		requireCertifyTier = null.BoolFrom(true)
	)

	getTagOrder := func(orderedTags []model.AssetGroupTag) []int {
		var order []int
		for _, tag := range orderedTags {
			order = append(order, tag.ID)
		}
		return order
	}

	t.Run("successfully deletes asset group tag tier", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		assetGroupTagTier, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testUser, testName, testDescription, position, requireCertifyTier)
		require.NoError(t, err)
		require.Equal(t, model.AssetGroupTagTypeTier, assetGroupTagTier.Type)
		require.Equal(t, testName, assetGroupTagTier.Name)
		require.Equal(t, testDescription, assetGroupTagTier.Description)

		assetGroupTagTier2, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testUser2, testName2, testDescription2, position2, requireCertifyTier)
		require.NoError(t, err)
		require.Equal(t, model.AssetGroupTagTypeTier, assetGroupTagTier2.Type)
		require.Equal(t, testName2, assetGroupTagTier2.Name)
		require.Equal(t, testDescription2, assetGroupTagTier2.Description)

		orderedTagsBefore, err := dbInst.GetOrderedAssetGroupTagTiers(testCtx)
		require.NoError(t, err)
		require.Equal(t, []int{1, assetGroupTagTier.ID, assetGroupTagTier2.ID}, getTagOrder(orderedTagsBefore))

		err = dbInst.DeleteAssetGroupTag(testCtx, testUser, assetGroupTagTier)
		require.NoError(t, err)

		orderedTagsAfter, err := dbInst.GetOrderedAssetGroupTagTiers(testCtx)
		require.NoError(t, err)
		require.Equal(t, []int{1, assetGroupTagTier2.ID}, getTagOrder(orderedTagsAfter))

		// verify history records were created
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.Len(t, history, 4)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[1].Action)
		require.Equal(t, model.AssetGroupHistoryActionDeleteTag, history[2].Action)
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[3].Action)
	})

	t.Run("successfully deletes asset group label", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		assetGroupTagLabel4, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, testUser4, testName4, testDescription4, null.Int32{}, null.Bool{})
		require.NoError(t, err)
		require.Equal(t, model.AssetGroupTagTypeLabel, assetGroupTagLabel4.Type)
		require.Equal(t, testName4, assetGroupTagLabel4.Name)
		require.Equal(t, testDescription4, assetGroupTagLabel4.Description)

		err = dbInst.DeleteAssetGroupTag(testCtx, testUser, assetGroupTagLabel4)
		require.NoError(t, err)

		// verify history records were created
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.Len(t, history, 2)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
		require.Equal(t, model.AssetGroupHistoryActionDeleteTag, history[1].Action)
	})

	t.Run("Non existant asset group tag errors out", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		_, err := dbInst.GetAssetGroupTag(testCtx, 1234)
		require.Error(t, err)
	})
}

func TestDatabase_GetAssetGroupTags(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testDescription = "test tag description"
	)

	var (
		err            error
		label1, label2 model.AssetGroupTag
		tier1, tier2   model.AssetGroupTag
	)
	label1, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 1", testDescription, null.Int32{}, null.Bool{})
	require.NoError(t, err)
	label2, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 2", testDescription, null.Int32{}, null.Bool{})
	require.NoError(t, err)
	tier1, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, model.User{}, "tier 1", testDescription, null.Int32From(2), null.BoolFrom(false))
	require.NoError(t, err)
	tier2, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, model.User{}, "tier 2", testDescription, null.Int32From(3), null.BoolFrom(false))
	require.NoError(t, err)

	t.Run("filtering for Label returns labels", func(t *testing.T) {
		ids := []int{
			label1.ID,
			label2.ID,
		}

		items, err := dbInst.GetAssetGroupTags(testCtx, model.SQLFilter{SQLString: "type = ?", Params: []any{model.AssetGroupTagTypeLabel}})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(items), 2)
		for _, itm := range items {
			if itm.CreatedBy == model.AssetGroupActorSystem {
				continue
			}
			require.Equal(t, itm.Type, model.AssetGroupTagTypeLabel)
			require.Contains(t, ids, itm.ID)
		}
	})

	t.Run("filtering for Tier returns tiers", func(t *testing.T) {
		ids := []int{
			tier1.ID,
			tier2.ID,
		}

		items, err := dbInst.GetAssetGroupTags(testCtx, model.SQLFilter{SQLString: "type = ?", Params: []any{model.AssetGroupTagTypeTier}})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(items), 2)
		for _, itm := range items {
			if itm.CreatedBy == model.AssetGroupActorSystem {
				continue
			}
			require.Equal(t, itm.Type, model.AssetGroupTagTypeTier)
			require.Contains(t, ids, itm.ID)
		}
	})

	t.Run("no filter returns everything", func(t *testing.T) {
		ids := []int{
			label1.ID,
			label2.ID,
			tier1.ID,
			tier2.ID,
		}
		types := []model.AssetGroupTagType{
			model.AssetGroupTagTypeLabel,
			model.AssetGroupTagTypeTier,
		}

		items, err := dbInst.GetAssetGroupTags(testCtx, model.SQLFilter{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(items), 4)
		for _, itm := range items {
			if itm.CreatedBy == model.AssetGroupActorSystem {
				continue
			}
			require.Contains(t, types, itm.Type)
			require.Contains(t, ids, itm.ID)
		}
	})
}

func TestDatabase_GetAssetGroupTagSelectorCounts(t *testing.T) {
	var (
		dbInst  = integration.SetupDB(t)
		testCtx = context.Background()
	)

	var (
		err            error
		label1, label2 model.AssetGroupTag
	)

	label1, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 1", "", null.Int32{}, null.Bool{})
	require.NoError(t, err)
	label2, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 2", "", null.Int32{}, null.Bool{})
	require.NoError(t, err)

	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "", "", false, true, null.BoolFrom(false), []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "", "", false, true, null.BoolFrom(false), []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "", "", false, true, null.BoolFrom(false), []model.SelectorSeed{})
	require.NoError(t, err)

	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label2.ID, model.User{}, "", "", false, true, null.BoolFrom(false), []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label2.ID, model.User{}, "", "", false, true, null.BoolFrom(false), []model.SelectorSeed{})
	require.NoError(t, err)

	t.Run("single item count", func(t *testing.T) {
		counts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{label1.ID})
		require.NoError(t, err)
		require.Equal(t, 1, len(counts))
		require.Equal(t, 3, counts[label1.ID])
	})

	t.Run("multi item count", func(t *testing.T) {
		counts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{label1.ID, label2.ID})
		require.NoError(t, err)
		require.Equal(t, 2, len(counts))
		require.Equal(t, 3, counts[label1.ID])
		require.Equal(t, 2, counts[label2.ID])
	})

	t.Run("single value set for id with no results", func(t *testing.T) {
		nonexistentId := 1234
		counts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{nonexistentId})
		require.NoError(t, err)
		require.Equal(t, 1, len(counts))
		require.Equal(t, 0, counts[nonexistentId])
	})

	t.Run("multi value set for id with no results", func(t *testing.T) {
		nonexistentId := 1234
		counts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{label1.ID, nonexistentId})
		require.NoError(t, err)
		require.Equal(t, 2, len(counts))
		require.Equal(t, 3, counts[label1.ID])
		require.Equal(t, 0, counts[nonexistentId])
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

	_, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, test1Selector.Name, test1Selector.Description, isDefault, allowDisable, autoCertify, test1Selector.Seeds)
	require.NoError(t, err)
	created_at := time.Now()
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, test2Selector.Name, test2Selector.Description, isDefault, allowDisable, autoCertify, test2Selector.Seeds)
	require.NoError(t, err)

	t.Run("successfully returns an array of selectors, no filters", func(t *testing.T) {
		orig_results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{}, model.SQLFilter{})
		require.NoError(t, err)

		results := make(model.AssetGroupTagSelectors, 0, 2)
		for _, n := range orig_results {
			if n.CreatedBy != model.AssetGroupActorSystem {
				results = append(results, n)
			}
		}

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
		orig_results, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1, model.SQLFilter{}, model.SQLFilter{SQLString: "type = ?", Params: []any{2}})
		require.NoError(t, err)

		results := make(model.AssetGroupTagSelectors, 0, 1)
		for _, n := range orig_results {
			if n.CreatedBy != model.AssetGroupActorSystem {
				results = append(results, n)
			}
		}

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

func TestDatabase_GetSelectorsByMemberId(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testSelectorId  = 1
		testNodeId      = uint64(1)
		certified       = 1
		testCertifiedBy = "testy"
		certifiedBy     = null.StringFrom(testCertifiedBy)
		source          = 1
		testMemberId    = 1
		isDefault       = false
		allowDisable    = true
		autoCertify     = null.BoolFrom(false)
		test1Selector   = model.AssetGroupTagSelector{
			Name:            "test selector name",
			Description:     "test description",
			AssetGroupTagId: 1,
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			},
		}
	)
	_, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, test1Selector.Name, test1Selector.Description, isDefault, allowDisable, autoCertify, test1Selector.Seeds)
	require.NoError(t, err)
	err = dbInst.InsertSelectorNode(testCtx, testSelectorId, graph.ID(testNodeId), model.AssetGroupCertification(certified), certifiedBy, model.AssetGroupSelectorNodeSource(source))
	require.NoError(t, err)
	selectors, err := dbInst.GetSelectorsByMemberId(testCtx, testMemberId, test1Selector.AssetGroupTagId)
	require.NoError(t, err)
	require.Equal(t, testSelectorId, selectors[0].AssetGroupTagId)
}
func TestDatabase_DeleteAssetGroupTagSelector(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
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

	selector, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)

	history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, model.AssetGroupHistoryActionCreateSelector, history[0].Action)

	t.Run("successfully deletes tag", func(t *testing.T) {
		err := dbInst.DeleteAssetGroupTagSelector(testCtx, model.User{}, selector)
		require.NoError(t, err)

		// verify selector is gone
		_, err = dbInst.GetAssetGroupTagSelectorBySelectorId(testCtx, selector.ID)
		require.EqualError(t, err, "entity not found")

		// verify a history record was created for the delete action
		history, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
		require.NoError(t, err)
		require.Len(t, history, 2)
		require.Equal(t, model.AssetGroupHistoryActionDeleteSelector, history[1].Action)
	})
}
