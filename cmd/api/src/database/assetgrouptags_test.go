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
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateAssetGroupTagSelector(t *testing.T) {
	t.Parallel()
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testActor       = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
		testName        = "test selector name"
		testDescription = "test description"
		isDefault       = false
		allowDisable    = true
		autoCertify     = model.SelectorAutoCertifyMethodAllMembers
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

	history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}}, 0, 0)
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
		autoCertify     = model.SelectorAutoCertifyMethodAllMembers
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
		autoCertify       = model.SelectorAutoCertifyMethodDisabled
		updateAutoCert    = model.SelectorAutoCertifyMethodSeedsOnly
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
		testCtx                = context.Background()
		tagTypeTier            = model.AssetGroupTagTypeTier
		tagTypeLabel           = model.AssetGroupTagTypeLabel
		testActor              = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
		testName               = "test tag name"
		testName2              = "test2 tag name"
		testName3              = "test3 tag name"
		testName4              = "test4 tag name"
		testDescription        = "test tag description"
		position               = null.Int32{}
		requireCertifyTier     = null.BoolFrom(true)
		requireCertifyLabel    = null.Bool{}
		glyph                  = null.StringFrom("rocket")
		sortAscendingCreatedAt = model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}}
	)

	t.Run("successfully creates tier", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		tag, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeTier, testActor, testName, testDescription, position, requireCertifyTier, glyph)
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
		require.Equal(t, glyph, tag.Glyph)

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
		require.Equal(t, glyph, tag.Glyph)

		// verify history record was also created
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
		require.NoError(t, err)
		require.Len(t, history, 1)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
	})

	t.Run("successfully creates label", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		tag, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeLabel, testActor, testName2, testDescription, position, requireCertifyLabel, glyph)
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
		require.Equal(t, glyph, tag.Glyph)

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
		require.Equal(t, glyph, tag.Glyph)
	})

	t.Run("successfully creates and shifts tiers", func(t *testing.T) {
		var (
			position2 = null.Int32From(2)
			position3 = null.Int32From(3)
			dbInst    = integration.SetupDB(t)
		)

		tag, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeTier, testActor, testName3, testDescription, position2, requireCertifyTier, null.String{})
		require.NoError(t, err)

		tag, err = dbInst.GetAssetGroupTag(testCtx, tag.ID)
		require.NoError(t, err)
		require.Equal(t, position2, tag.Position)

		// Create another tag at an existing position, forcing the first tag down
		tag2, err := dbInst.CreateAssetGroupTag(testCtx, tagTypeTier, testActor, testName4, testDescription, position2, requireCertifyTier, null.String{})
		require.NoError(t, err)

		tag2, err = dbInst.GetAssetGroupTag(testCtx, tag2.ID)
		require.NoError(t, err)
		require.Equal(t, position2, tag2.Position)

		tag, err = dbInst.GetAssetGroupTag(testCtx, tag.ID)
		require.NoError(t, err)
		require.Equal(t, position3, tag.Position)

		// verify history record was also created and shifted
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
		require.NoError(t, err)
		require.Len(t, history, 3)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[1].Action)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[2].Action)
	})

	t.Run("Non existent tag errors out", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		_, err := dbInst.GetAssetGroupTag(testCtx, 1234)
		require.Error(t, err)
	})

	t.Run("Duplicate null glyph is accepted", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		_, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "t1 name", "", null.Int32{}, null.Bool{}, null.String{})
		require.NoError(t, err)

		_, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "t2 name", "", null.Int32{}, null.Bool{}, null.String{})
		require.NoError(t, err)
	})

	t.Run("Duplicate glyph errors out", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		_, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "t1 name", "", null.Int32{}, null.Bool{}, glyph)
		require.NoError(t, err)

		_, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "t2 name", "", null.Int32{}, null.Bool{}, glyph)
		require.Error(t, err)
		require.EqualError(t, err, "duplicate glyph: ERROR: duplicate key value violates unique constraint \"asset_group_tags_glyph_key\" (SQLSTATE 23505)")
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
			null.String{},
		)
		require.NoError(t, err)

		toUpdate := origTier
		toUpdate.Name = "updated name"
		toUpdate.Description = "updated desc"
		toUpdate.RequireCertify.SetValid(true)
		toUpdate.Glyph = null.StringFrom("rocket")

		updatedTier, err := dbInst.UpdateAssetGroupTag(testCtx, testActor, toUpdate)
		require.NoError(t, err)

		require.Equal(t, toUpdate.Name, updatedTier.Name)
		require.Equal(t, toUpdate.Description, updatedTier.Description)
		require.Equal(t, toUpdate.RequireCertify, updatedTier.RequireCertify)
		require.Equal(t, toUpdate.Glyph, updatedTier.Glyph)

		// verify history records were created
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}}, 0, 0)
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
			null.String{},
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
			null.String{},
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
			null.String{},
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
			null.String{},
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
			null.String{},
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
			null.String{},
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
			null.String{},
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
		testCtx                = context.Background()
		testActor              = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
		sortAscendingCreatedAt = model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}}
	)

	t.Run("shifts tier higher successfully", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		tag1, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift1", "", null.Int32From(2), null.Bool{}, null.String{})
		require.NoError(t, err)
		tag2, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift2", "", null.Int32From(3), null.Bool{}, null.String{})
		require.NoError(t, err)
		tag3, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift3", "", null.Int32From(4), null.Bool{}, null.String{})
		require.NoError(t, err)
		tag4, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "upshift4", "", null.Int32From(5), null.Bool{}, null.String{})
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
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
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

		tag1, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift1", "", null.Int32From(2), null.Bool{}, null.String{})
		require.NoError(t, err)
		tag2, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift2", "", null.Int32From(3), null.Bool{}, null.String{})
		require.NoError(t, err)
		tag3, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift3", "", null.Int32From(4), null.Bool{}, null.String{})
		require.NoError(t, err)
		tag4, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "downshift4", "", null.Int32From(5), null.Bool{}, null.String{})
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
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
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
		testCtx                = context.Background()
		userID, _              = uuid.NewV4()
		userID2, _             = uuid.NewV4()
		userID4, _             = uuid.NewV4()
		testUser               = model.User{Unique: model.Unique{ID: userID}}
		testUser2              = model.User{Unique: model.Unique{ID: userID2}}
		testUser4              = model.User{Unique: model.Unique{ID: userID4}}
		testName               = "test tag name"
		testName2              = "test tag name2"
		testName4              = "test tag name4"
		testDescription        = "test tag description"
		testDescription2       = "test tag description2"
		testDescription4       = "test tag description4"
		position               = null.Int32From(2)
		position2              = null.Int32From(3)
		requireCertifyTier     = null.BoolFrom(true)
		sortAscendingCreatedAt = model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}}
		glyph                  = null.String{}
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

		assetGroupTagTier, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testUser, testName, testDescription, position, requireCertifyTier, glyph)
		require.NoError(t, err)
		require.Equal(t, model.AssetGroupTagTypeTier, assetGroupTagTier.Type)
		require.Equal(t, testName, assetGroupTagTier.Name)
		require.Equal(t, testDescription, assetGroupTagTier.Description)

		assetGroupTagTier2, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testUser2, testName2, testDescription2, position2, requireCertifyTier, glyph)
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
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
		require.NoError(t, err)
		require.Len(t, history, 4)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[1].Action)
		require.Equal(t, model.AssetGroupHistoryActionDeleteTag, history[2].Action)
		require.Equal(t, model.AssetGroupHistoryActionUpdateTag, history[3].Action)
	})

	t.Run("successfully deletes asset group label", func(t *testing.T) {
		dbInst := integration.SetupDB(t)

		assetGroupTagLabel4, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, testUser4, testName4, testDescription4, null.Int32{}, null.Bool{}, glyph)
		require.NoError(t, err)
		require.Equal(t, model.AssetGroupTagTypeLabel, assetGroupTagLabel4.Type)
		require.Equal(t, testName4, assetGroupTagLabel4.Name)
		require.Equal(t, testDescription4, assetGroupTagLabel4.Description)

		err = dbInst.DeleteAssetGroupTag(testCtx, testUser, assetGroupTagLabel4)
		require.NoError(t, err)

		// verify history records were created
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
		require.NoError(t, err)
		require.Len(t, history, 2)
		require.Equal(t, model.AssetGroupHistoryActionCreateTag, history[0].Action)
		require.Equal(t, model.AssetGroupHistoryActionDeleteTag, history[1].Action)
	})

	t.Run("Non existent asset group tag errors out", func(t *testing.T) {
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
	label1, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 1", testDescription, null.Int32{}, null.Bool{}, null.String{})
	require.NoError(t, err)
	label2, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 2", testDescription, null.Int32{}, null.Bool{}, null.String{})
	require.NoError(t, err)
	tier1, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, model.User{}, "tier 1", testDescription, null.Int32From(2), null.BoolFrom(false), null.String{})
	require.NoError(t, err)
	tier2, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, model.User{}, "tier 2", testDescription, null.Int32From(3), null.BoolFrom(false), null.String{})
	require.NoError(t, err)

	t.Run("filtering for Label returns labels", func(t *testing.T) {
		ids := []int{
			label1.ID,
			label2.ID,
		}

		items, err := dbInst.GetAssetGroupTags(testCtx, model.SQLFilter{SQLString: "type = " + strconv.Itoa(int(model.AssetGroupTagTypeLabel))})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(items), 2)
		for _, itm := range items {
			if itm.CreatedBy == model.AssetGroupActorBloodHound {
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

		items, err := dbInst.GetAssetGroupTags(testCtx, model.SQLFilter{SQLString: "type = " + strconv.Itoa(int(model.AssetGroupTagTypeTier))})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(items), 2)
		for _, itm := range items {
			if itm.CreatedBy == model.AssetGroupActorBloodHound {
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
			if itm.CreatedBy == model.AssetGroupActorBloodHound {
				continue
			}
			require.Contains(t, types, itm.Type)
			require.Contains(t, ids, itm.ID)
		}
	})
}

func TestDatabase_GetAssetGroupTagSelectorCounts(t *testing.T) {
	var (
		dbInst       = integration.SetupDB(t)
		testCtx      = context.Background()
		user         = model.User{}
		disabledTime = null.TimeFrom(time.Date(2025, time.March, 25, 12, 0, 0, 0, time.UTC))
	)

	var (
		err            error
		label1, label2 model.AssetGroupTag
	)

	label1, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 1", "", null.Int32{}, null.Bool{}, null.String{})
	require.NoError(t, err)
	label2, err = dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label 2", "", null.Int32{}, null.Bool{}, null.String{})
	require.NoError(t, err)

	// label1 selectors
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "1", "", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "2", "", false, true, model.SelectorAutoCertifyMethodAllMembers, []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "3", "", false, false, model.SelectorAutoCertifyMethodSeedsOnly, []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "4", "", true, false, model.SelectorAutoCertifyMethodSeedsOnly, []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "5", "", true, false, model.SelectorAutoCertifyMethodSeedsOnly, []model.SelectorSeed{})
	require.NoError(t, err)
	selector6, err := dbInst.CreateAssetGroupTagSelector(testCtx, label1.ID, model.User{}, "6", "", true, true, model.SelectorAutoCertifyMethodSeedsOnly, []model.SelectorSeed{})
	require.NoError(t, err)
	selector6.DisabledAt = disabledTime
	selector6.DisabledBy = null.StringFrom(user.ID.String())
	_, err = dbInst.UpdateAssetGroupTagSelector(testCtx, user.ID.String(), "someEmail@whatever.com", selector6)
	require.NoError(t, err)

	// label2 selectors
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label2.ID, model.User{}, "7", "", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label2.ID, model.User{}, "8", "", false, false, model.SelectorAutoCertifyMethodAllMembers, []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label2.ID, model.User{}, "9", "", true, false, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, label2.ID, model.User{}, "10", "", true, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)

	t.Run("single asset group tag selector(s) count", func(t *testing.T) {
		selectorCounts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{label1.ID})
		counts := selectorCounts[label1.ID]
		require.NoError(t, err)

		require.Equal(t, 6, counts.Selectors)
		require.Equal(t, 3, counts.CustomSelectors)
		require.Equal(t, 2, counts.DefaultSelectors)
		require.Equal(t, 1, counts.DisabledSelectors)
	})

	t.Run("multiple asset group tag selectors count", func(t *testing.T) {
		selectorCounts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{label1.ID, label2.ID})
		counts1 := selectorCounts[label1.ID]
		counts2 := selectorCounts[label2.ID]
		require.NoError(t, err)

		// label1
		require.Equal(t, 6, counts1.Selectors)
		require.Equal(t, 3, counts1.CustomSelectors)
		require.Equal(t, 2, counts1.DefaultSelectors)
		require.Equal(t, 1, counts1.DisabledSelectors)

		// label2
		require.Equal(t, 4, counts2.Selectors)
		require.Equal(t, 2, counts2.CustomSelectors)
		require.Equal(t, 2, counts2.DefaultSelectors)
		require.Equal(t, 0, counts2.DisabledSelectors)
	})

	t.Run("single asset group tag value set for id with no results", func(t *testing.T) {
		nonexistentId := 1234
		selectorCounts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{nonexistentId})
		counts := selectorCounts[nonexistentId]

		require.NoError(t, err)
		require.Equal(t, 0, counts.Selectors)
		require.Equal(t, 0, counts.CustomSelectors)
		require.Equal(t, 0, counts.DefaultSelectors)
		require.Equal(t, 0, counts.DisabledSelectors)
	})

	t.Run("multiple asset group tag values set for id with no results", func(t *testing.T) {
		nonexistentId := 1234
		selectorCounts, err := dbInst.GetAssetGroupTagSelectorCounts(testCtx, []int{label1.ID, nonexistentId})
		counts := selectorCounts[label1.ID]

		require.NoError(t, err)
		require.Equal(t, 6, counts.Selectors)
		require.Equal(t, 3, counts.CustomSelectors)
		require.Equal(t, 2, counts.DefaultSelectors)
		require.Equal(t, 1, counts.DisabledSelectors)
	})
}

func TestDatabase_GetAssetGroupTagSelectorsBySelectorIdFilteredAndPaginated(t *testing.T) {
	var (
		dbInst        = integration.SetupDB(t)
		testCtx       = context.Background()
		isDefault     = false
		allowDisable  = true
		autoCertify   = model.SelectorAutoCertifyMethodAllMembers
		test1Selector = model.AssetGroupTagSelector{
			Name:            "test selector name",
			Description:     "test description",
			AssetGroupTagId: 1,
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			},
			AllowDisable: true,
			AutoCertify:  model.SelectorAutoCertifyMethodAllMembers,
		}
		test2Selector = model.AssetGroupTagSelector{
			Name:            "test2 selector name",
			Description:     "test2 description",
			AssetGroupTagId: 1,
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
			},
			AllowDisable: true,
			AutoCertify:  model.SelectorAutoCertifyMethodAllMembers,
		}
	)

	test_started_at := time.Now()
	_, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, test1Selector.Name, test1Selector.Description, isDefault, allowDisable, autoCertify, test1Selector.Seeds)
	require.NoError(t, err)
	created_at := time.Now()
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, test2Selector.Name, test2Selector.Description, isDefault, allowDisable, autoCertify, test2Selector.Seeds)
	require.NoError(t, err)

	t.Run("successfully returns an array of selectors, no filters", func(t *testing.T) {
		orig_results, _, err := dbInst.GetAssetGroupTagSelectorsByTagId(testCtx, 1)
		require.NoError(t, err)

		results := make(model.AssetGroupTagSelectors, 0, 2)
		for _, n := range orig_results {
			if n.CreatedBy != model.AssetGroupActorBloodHound {
				results = append(results, n)
			}
		}

		require.Len(t, results, 2)
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
		orig_results, _, err := dbInst.GetAssetGroupTagSelectorsByTagIdFilteredAndPaginated(testCtx, 1, model.SQLFilter{}, model.SQLFilter{SQLString: "type = 2"}, model.Sort{}, 0, 0)
		require.NoError(t, err)

		results := make(model.AssetGroupTagSelectors, 0, 1)
		for _, n := range orig_results {
			if n.CreatedBy != model.AssetGroupActorBloodHound {
				results = append(results, n)
			}
		}

		require.Len(t, results, 1)
		for idx, seed := range test2Selector.Seeds {
			require.Equal(t, seed.Type, results[0].Seeds[idx].Type)
			require.Equal(t, seed.Value, results[0].Seeds[idx].Value)
		}
	})

	t.Run("successfully returns an array of selector filters", func(t *testing.T) {
		results, _, err := dbInst.GetAssetGroupTagSelectorsByTagIdFilteredAndPaginated(testCtx, 1, model.SQLFilter{SQLString: "created_at >= ?", Params: []any{created_at}}, model.SQLFilter{}, model.Sort{}, 0, 0)
		require.NoError(t, err)

		require.Len(t, results, 1)
		require.True(t, results[0].CreatedAt.After(created_at))

	})

	t.Run("successfully returns an array using the skip param", func(t *testing.T) {
		results, count, err := dbInst.GetAssetGroupTagSelectorsByTagIdFilteredAndPaginated(testCtx, 1, model.SQLFilter{SQLString: "created_at >= ?", Params: []any{test_started_at}}, model.SQLFilter{}, model.Sort{}, 1, 0)
		require.NoError(t, err)

		require.Len(t, results, 1)
		require.Equal(t, 2, count)
		require.Equal(t, test2Selector.Name, results[0].Name)
	})

	t.Run("successfully returns an array sorted by name desc", func(t *testing.T) {
		results, _, err := dbInst.GetAssetGroupTagSelectorsByTagIdFilteredAndPaginated(testCtx, 1, model.SQLFilter{}, model.SQLFilter{}, model.Sort{{Direction: model.DescendingSortDirection, Column: "name"}}, 0, 0)
		require.NoError(t, err)

		require.Greater(t, len(results), 1)
		require.True(t, sort.SliceIsSorted(results, func(i, j int) bool {
			return strings.ReplaceAll(strings.ToLower(results[i].Name), " ", "") > strings.ReplaceAll(strings.ToLower(results[j].Name), " ", "")
		}))
	})

	t.Run("successfully returns an array using the limit param", func(t *testing.T) {
		results, count, err := dbInst.GetAssetGroupTagSelectorsByTagIdFilteredAndPaginated(testCtx, 1, model.SQLFilter{SQLString: "created_at >= ?", Params: []any{test_started_at}}, model.SQLFilter{}, model.Sort{}, 0, 1)
		require.NoError(t, err)

		require.Len(t, results, 1)
		require.Equal(t, 2, count)
		require.Equal(t, test1Selector.Name, results[0].Name)
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
		autoCertify     = model.SelectorAutoCertifyMethodSeedsOnly
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
	err = dbInst.InsertSelectorNode(testCtx, 1, testSelectorId, graph.ID(testNodeId), model.AssetGroupCertification(certified), certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)
	selectors, err := dbInst.GetSelectorsByMemberId(testCtx, testMemberId, test1Selector.AssetGroupTagId)
	require.NoError(t, err)
	require.Equal(t, testSelectorId, selectors[0].ID)
}
func TestDatabase_DeleteAssetGroupTagSelector(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testName        = "test selector name"
		testDescription = "test description"
		isDefault       = false
		allowDisable    = true
		autoCertify     = model.SelectorAutoCertifyMethodSeedsOnly
		testSeeds       = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectID5678"},
		}
		sortAscendingCreatedAt = model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}}
	)

	selector, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)

	history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
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
		history, _, err := dbInst.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, sortAscendingCreatedAt, 0, 0)
		require.NoError(t, err)
		require.Len(t, history, 2)
		require.Equal(t, model.AssetGroupHistoryActionDeleteSelector, history[1].Action)
	})
}

func TestDatabase_GetOrderedAssetGroupTagTiers(t *testing.T) {
	var (
		testCtx      = context.Background()
		dbInst, user = initAndCreateUser(t)
		tagToDelete  model.AssetGroupTag
	)

	// Create tiers
	for i := range 5 {
		tag, err := dbInst.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, user, fmt.Sprintf("tag %d", i), "", null.Int32From(int32(i+2)), null.Bool{}, null.String{})
		require.NoError(t, err)
		// Delete the fourth entry to ensure positions are changed
		if i == 3 {
			tagToDelete = tag
		}
	}

	// Delete a tier to ensure deleted tiers are skipped
	err := dbInst.DeleteAssetGroupTag(testCtx, user, tagToDelete)
	require.NoError(t, err)

	orderedTags, err := dbInst.GetOrderedAssetGroupTagTiers(testCtx)
	require.NoError(t, err)
	for i, tag := range orderedTags {
		assert.Equal(t, model.AssetGroupTagTypeTier, tag.Type)
		assert.True(t, tag.DeletedAt.IsZero())
		assert.EqualValues(t, i+1, tag.Position.ValueOrZero())
	}
}

func TestDatabase_GetAssetGroupTagSelectors(t *testing.T) {
	var (
		dbInst        = integration.SetupDB(t)
		testCtx       = context.Background()
		isDefault     = false
		allowDisable  = true
		autoCertify   = model.SelectorAutoCertifyMethodDisabled
		test1Selector = model.AssetGroupTagSelector{
			Name:        "test selector name",
			Description: "test description",
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			},
		}
		test2Selector = model.AssetGroupTagSelector{
			Name:        "test2 selector name",
			Description: "test2 description",
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeCypher, Value: "MATCH (n:User) RETURN n LIMIT 1;"},
			},
		}
	)

	_, err := dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, test1Selector.Name, test1Selector.Description, isDefault, allowDisable, autoCertify, test1Selector.Seeds)
	require.NoError(t, err)
	_, err = dbInst.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, test2Selector.Name, test2Selector.Description, isDefault, allowDisable, autoCertify, test2Selector.Seeds)
	require.NoError(t, err)

	t.Run("returns all selectors", func(t *testing.T) {
		items, err := dbInst.GetAssetGroupTagSelectors(testCtx, model.SQLFilter{}, 0)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(items), 2)
	})
}

func TestDatabase_UpdateCertificationBySelectorNode(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx              = context.Background()
		testNote             = null.StringFrom("test")
		testAssetGroupTagId1 = 1
		testSelectorId1      = 1
		testSelectorId2      = 2
		testMemberId1        = 1
		testNodeId1          = graph.ID(uint64(testMemberId1))
		testMemberId2        = 2
		testNodeId2          = graph.ID(uint64(testMemberId2))
		certifiedBy          = null.StringFrom("testy")
		source               = 1
		isDefault            = false
		allowDisable         = true
		autoCertify          = model.SelectorAutoCertifyMethodSeedsOnly
		test1Selector        = model.AssetGroupTagSelector{
			Name:        "test selector name",
			Description: "test description",
			Seeds: []model.SelectorSeed{
				{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			},
		}
		updateInputCertify = database.UpdateCertificationBySelectorNodeInput{AssetGroupTagId: testAssetGroupTagId1, SelectorId: testSelectorId1, CertifiedBy: certifiedBy, CertificationStatus: model.AssetGroupCertificationManual, NodeId: testNodeId1, NodeName: test1Selector.Name, Note: testNote}
		updateInputPending = database.UpdateCertificationBySelectorNodeInput{AssetGroupTagId: testAssetGroupTagId1, SelectorId: testSelectorId2, CertifiedBy: certifiedBy, CertificationStatus: model.AssetGroupCertificationPending, NodeId: testNodeId1, NodeName: test1Selector.Name, Note: testNote}
		updateInputNode2   = database.UpdateCertificationBySelectorNodeInput{AssetGroupTagId: testAssetGroupTagId1, SelectorId: testSelectorId1, CertifiedBy: certifiedBy, CertificationStatus: model.AssetGroupCertificationManual, NodeId: testNodeId2, NodeName: test1Selector.Name, Note: testNote}
		sort               = make(model.Sort, 0)
	)
	_, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, testAssetGroupTagId1, model.User{}, test1Selector.Name, test1Selector.Description, isDefault, allowDisable, autoCertify, test1Selector.Seeds)
	require.NoError(t, err)

	// insert selector nodes
	err = suite.BHDatabase.InsertSelectorNode(testCtx, testAssetGroupTagId1, testSelectorId1, testNodeId1, model.AssetGroupCertificationRevoked, certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)
	err = suite.BHDatabase.InsertSelectorNode(testCtx, testAssetGroupTagId1, testSelectorId1, testNodeId2, model.AssetGroupCertificationPending, certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)
	err = suite.BHDatabase.InsertSelectorNode(testCtx, testAssetGroupTagId1, testSelectorId2, testNodeId1, model.AssetGroupCertificationPending, certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)

	t.Run("updates certification by selector node, certified", func(t *testing.T) {
		inputs := []database.UpdateCertificationBySelectorNodeInput{updateInputCertify}
		sqlFilter := model.SQLFilter{SQLString: "AND node_id = ?", Params: []any{updateInputCertify.NodeId}}
		_, rowCountBefore, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		err = suite.BHDatabase.UpdateCertificationBySelectorNode(testCtx, inputs)
		require.NoError(t, err)
		// confirm selector was updated
		selectorNodes, count, err := suite.BHDatabase.GetSelectorNodesBySelectorIdsFilteredAndPaginated(testCtx, sqlFilter, sort, 0, 100, updateInputCertify.SelectorId)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Equal(t, model.AssetGroupCertificationManual, selectorNodes[0].Certified)

		//confirm only one history record was created
		_, rowCountAfter, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		require.Equal(t, rowCountBefore+1, rowCountAfter)
	})

	t.Run("updates certification by selector node, pending", func(t *testing.T) {
		inputs := []database.UpdateCertificationBySelectorNodeInput{updateInputPending}
		sqlFilter := model.SQLFilter{SQLString: "AND node_id = ?", Params: []any{updateInputPending.NodeId}}
		_, rowCountBefore, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		err = suite.BHDatabase.UpdateCertificationBySelectorNode(testCtx, inputs)
		require.NoError(t, err)
		// confirm selectors were updated
		selectorNodes, count, err := suite.BHDatabase.GetSelectorNodesBySelectorIdsFilteredAndPaginated(testCtx, sqlFilter, sort, 0, 100, updateInputPending.SelectorId)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.Equal(t, model.AssetGroupCertificationPending, selectorNodes[0].Certified)

		//confirm no history record were created
		_, rowCountAfter, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		require.Equal(t, rowCountBefore, rowCountAfter)
	})

	t.Run("updates certification by selector node, mix of pending and certified for single node ID", func(t *testing.T) {
		inputs := []database.UpdateCertificationBySelectorNodeInput{updateInputCertify, updateInputPending}
		sqlFilter := model.SQLFilter{SQLString: "AND node_id = ?", Params: []any{updateInputCertify.NodeId}}
		_, rowCountBefore, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		err = suite.BHDatabase.UpdateCertificationBySelectorNode(testCtx, inputs)
		require.NoError(t, err)
		// confirm selectors were updated
		selectorNodes, count, err := suite.BHDatabase.GetSelectorNodesBySelectorIdsFilteredAndPaginated(testCtx, sqlFilter, sort, 0, 100, updateInputCertify.SelectorId, updateInputPending.SelectorId)
		require.NoError(t, err)
		require.Equal(t, 2, count)
		require.Equal(t, model.AssetGroupCertificationManual, selectorNodes[0].Certified)
		require.Equal(t, model.AssetGroupCertificationPending, selectorNodes[1].Certified)

		//confirm only one history record was created
		_, rowCountAfter, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		require.Equal(t, rowCountBefore+1, rowCountAfter)
	})

	t.Run("updates certification by selector node, 2 nodes", func(t *testing.T) {
		inputs := []database.UpdateCertificationBySelectorNodeInput{updateInputCertify, updateInputNode2}
		_, rowCountBefore, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		err = suite.BHDatabase.UpdateCertificationBySelectorNode(testCtx, inputs)
		require.NoError(t, err)
		// confirm selectors were updated
		selectors, err := suite.BHDatabase.GetSelectorNodesBySelectorIds(testCtx, testSelectorId1)
		require.NoError(t, err)
		require.Equal(t, 2, len(selectors))
		require.Equal(t, model.AssetGroupCertificationManual, selectors[0].Certified)
		require.Equal(t, model.AssetGroupCertificationManual, selectors[1].Certified)

		//confirm two history records created
		_, rowCountAfter, err := suite.BHDatabase.GetAssetGroupHistoryRecords(testCtx, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		require.Equal(t, rowCountBefore+2, rowCountAfter)
	})
}

func TestDatabase_InsertSelectorNodes(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx   = context.Background()
		testActor = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
	)

	t.Run("returns nil for empty input", func(t *testing.T) {
		require.NoError(t, suite.BHDatabase.InsertSelectorNodes(testCtx, nil))
		require.NoError(t, suite.BHDatabase.InsertSelectorNodes(testCtx, []model.AssetGroupSelectorNode{}))
	})

	t.Run("inserts selector nodes and creates asset group history records for certified nodes", func(t *testing.T) {
		var (
			assetGroupTagId = 1
			selector        = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch insert selector nodes")
			pendingNodeId   = graph.ID(50001)
			autoNodeId      = graph.ID(50002)
			manualNodeId    = graph.ID(50003)
			revokedNodeId   = graph.ID(50004)

			pendingNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            pendingNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "pending-primary-kind",
				NodeEnvironmentId: "pending-environment",
				NodeObjectId:      "pending-object-id",
				NodeName:          "insert-history-pending-node",
			}
			autoNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            autoNodeId,
				Certified:         model.AssetGroupCertificationAuto,
				CertifiedBy:       null.StringFrom(model.AssetGroupActorBloodHound),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "auto-primary-kind",
				NodeEnvironmentId: "auto-environment",
				NodeObjectId:      "auto-object-id",
				NodeName:          "insert-history-auto-node",
			}
			manualNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            manualNodeId,
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       null.StringFrom("manual-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceChild,
				NodePrimaryKind:   "manual-primary-kind",
				NodeEnvironmentId: "manual-environment",
				NodeObjectId:      "manual-object-id",
				NodeName:          "insert-history-manual-node",
			}
			revokedNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            revokedNodeId,
				Certified:         model.AssetGroupCertificationRevoked,
				CertifiedBy:       null.StringFrom("revoked-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "revoked-primary-kind",
				NodeEnvironmentId: "revoked-environment",
				NodeObjectId:      "revoked-object-id",
				NodeName:          "insert-history-revoked-node",
			}
		)

		require.NoError(t, suite.BHDatabase.InsertSelectorNodes(testCtx, []model.AssetGroupSelectorNode{
			pendingNode,
			autoNode,
			manualNode,
			revokedNode,
		}))

		selectorNodesBySelectorId := requireSelectorNodesBySelectorAndNodeId(t, testCtx, suite.BHDatabase, selector.ID)
		require.Len(t, selectorNodesBySelectorId[selector.ID], 4)
		assertUpdateSelectorNodesTestNode(t, pendingNode, selectorNodesBySelectorId[selector.ID][pendingNodeId])
		assertUpdateSelectorNodesTestNode(t, autoNode, selectorNodesBySelectorId[selector.ID][autoNodeId])
		assertUpdateSelectorNodesTestNode(t, manualNode, selectorNodesBySelectorId[selector.ID][manualNodeId])
		assertUpdateSelectorNodesTestNode(t, revokedNode, selectorNodesBySelectorId[selector.ID][revokedNodeId])

		historyRecordsByTarget := requireUpdateSelectorNodesHistoryRecordsByTarget(t, testCtx, suite.BHDatabase, pendingNode.NodeName, autoNode.NodeName, manualNode.NodeName, revokedNode.NodeName)
		require.Len(t, historyRecordsByTarget, 3)
		assert.NotContains(t, historyRecordsByTarget, pendingNode.NodeName)
		assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeAuto, autoNode.NodeName, assetGroupTagId, historyRecordsByTarget[autoNode.NodeName])
		assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeManual, manualNode.NodeName, assetGroupTagId, historyRecordsByTarget[manualNode.NodeName])
		assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeRevoked, revokedNode.NodeName, assetGroupTagId, historyRecordsByTarget[revokedNode.NodeName])

		require.NoError(t, suite.BHDatabase.InsertSelectorNodes(testCtx, []model.AssetGroupSelectorNode{autoNode}))
		historyRecordsByTarget = requireUpdateSelectorNodesHistoryRecordsByTarget(t, testCtx, suite.BHDatabase, pendingNode.NodeName, autoNode.NodeName, manualNode.NodeName, revokedNode.NodeName)
		require.Len(t, historyRecordsByTarget, 3)
	})
}

func TestDatabase_SelectorNodesBatching(t *testing.T) {
	const selectorNodeBatchTestCount = 5001

	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx                   = context.Background()
		testActor                 = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
		assetGroupTagId           = 1
		selector                  = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "temporary selector nodes batching")
		firstBatchCertifiedIndex  = 0
		secondBatchCertifiedIndex = selectorNodeBatchTestCount - 1
		firstBatchBoundaryIndex   = selectorNodeBatchTestCount - 2
		nodes                     = make([]model.AssetGroupSelectorNode, 0, selectorNodeBatchTestCount)
		updatedNodes              = make([]model.AssetGroupSelectorNode, 0, selectorNodeBatchTestCount)
	)

	for nodeIndex := 0; nodeIndex < selectorNodeBatchTestCount; nodeIndex++ {
		var (
			certified          = model.AssetGroupCertificationPending
			certifiedBy        = null.String{}
			updatedCertified   = model.AssetGroupCertificationPending
			updatedCertifiedBy = null.String{}
			nodeId             = graph.ID(600000 + nodeIndex)
		)

		if nodeIndex == firstBatchCertifiedIndex || nodeIndex == secondBatchCertifiedIndex {
			certified = model.AssetGroupCertificationAuto
			certifiedBy = null.StringFrom(model.AssetGroupActorBloodHound)
			updatedCertified = certified
			updatedCertifiedBy = certifiedBy
		}

		if nodeIndex == firstBatchCertifiedIndex {
			updatedCertified = model.AssetGroupCertificationManual
			updatedCertifiedBy = null.StringFrom("temporary-manual-certifier")
		} else if nodeIndex == secondBatchCertifiedIndex {
			updatedCertified = model.AssetGroupCertificationRevoked
			updatedCertifiedBy = null.StringFrom("temporary-revoked-certifier")
		}

		nodes = append(nodes, model.AssetGroupSelectorNode{
			SelectorId:        selector.ID,
			NodeId:            nodeId,
			Certified:         certified,
			CertifiedBy:       certifiedBy,
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			NodePrimaryKind:   fmt.Sprintf("temporary-insert-kind-%d", nodeIndex),
			NodeEnvironmentId: fmt.Sprintf("temporary-insert-environment-%d", nodeIndex),
			NodeObjectId:      fmt.Sprintf("temporary-insert-object-%d", nodeIndex),
			NodeName:          fmt.Sprintf("temporary-insert-node-%d", nodeIndex),
		})

		updatedNodes = append(updatedNodes, model.AssetGroupSelectorNode{
			SelectorId:        selector.ID,
			NodeId:            nodeId,
			Certified:         updatedCertified,
			CertifiedBy:       updatedCertifiedBy,
			Source:            model.AssetGroupSelectorNodeSourceParent,
			NodePrimaryKind:   fmt.Sprintf("temporary-update-kind-%d", nodeIndex),
			NodeEnvironmentId: fmt.Sprintf("temporary-update-environment-%d", nodeIndex),
			NodeObjectId:      fmt.Sprintf("temporary-update-object-%d", nodeIndex),
			NodeName:          fmt.Sprintf("temporary-update-node-%d", nodeIndex),
		})
	}

	require.NoError(t, suite.BHDatabase.InsertSelectorNodes(testCtx, nodes))

	selectorNodesBySelectorId := requireSelectorNodesBySelectorAndNodeId(t, testCtx, suite.BHDatabase, selector.ID)
	require.Len(t, selectorNodesBySelectorId[selector.ID], selectorNodeBatchTestCount)
	assertUpdateSelectorNodesTestNode(t, nodes[firstBatchCertifiedIndex], selectorNodesBySelectorId[selector.ID][nodes[firstBatchCertifiedIndex].NodeId])
	assertUpdateSelectorNodesTestNode(t, nodes[firstBatchBoundaryIndex], selectorNodesBySelectorId[selector.ID][nodes[firstBatchBoundaryIndex].NodeId])
	assertUpdateSelectorNodesTestNode(t, nodes[secondBatchCertifiedIndex], selectorNodesBySelectorId[selector.ID][nodes[secondBatchCertifiedIndex].NodeId])

	historyRecordsByTarget := requireUpdateSelectorNodesHistoryRecordsByTarget(t, testCtx, suite.BHDatabase, nodes[firstBatchCertifiedIndex].NodeName, nodes[secondBatchCertifiedIndex].NodeName)
	require.Len(t, historyRecordsByTarget, 2)
	assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeAuto, nodes[firstBatchCertifiedIndex].NodeName, assetGroupTagId, historyRecordsByTarget[nodes[firstBatchCertifiedIndex].NodeName])
	assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeAuto, nodes[secondBatchCertifiedIndex].NodeName, assetGroupTagId, historyRecordsByTarget[nodes[secondBatchCertifiedIndex].NodeName])

	require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(testCtx, updatedNodes))

	selectorNodesBySelectorId = requireSelectorNodesBySelectorAndNodeId(t, testCtx, suite.BHDatabase, selector.ID)
	require.Len(t, selectorNodesBySelectorId[selector.ID], selectorNodeBatchTestCount)
	assertUpdateSelectorNodesTestNode(t, updatedNodes[firstBatchCertifiedIndex], selectorNodesBySelectorId[selector.ID][updatedNodes[firstBatchCertifiedIndex].NodeId])
	assertUpdateSelectorNodesTestNode(t, updatedNodes[firstBatchBoundaryIndex], selectorNodesBySelectorId[selector.ID][updatedNodes[firstBatchBoundaryIndex].NodeId])
	assertUpdateSelectorNodesTestNode(t, updatedNodes[secondBatchCertifiedIndex], selectorNodesBySelectorId[selector.ID][updatedNodes[secondBatchCertifiedIndex].NodeId])

	historyRecordsByTarget = requireUpdateSelectorNodesHistoryRecordsByTarget(t, testCtx, suite.BHDatabase, updatedNodes[firstBatchCertifiedIndex].NodeName, updatedNodes[secondBatchCertifiedIndex].NodeName)
	require.Len(t, historyRecordsByTarget, 2)
	assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeManual, updatedNodes[firstBatchCertifiedIndex].NodeName, assetGroupTagId, historyRecordsByTarget[updatedNodes[firstBatchCertifiedIndex].NodeName])
	assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeRevoked, updatedNodes[secondBatchCertifiedIndex].NodeName, assetGroupTagId, historyRecordsByTarget[updatedNodes[secondBatchCertifiedIndex].NodeName])
}

func TestDatabase_UpdateSelectorNodes(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx   = context.Background()
		testActor = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
	)

	t.Run("returns nil for empty input", func(t *testing.T) {
		require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(testCtx, nil))
		require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(testCtx, []model.AssetGroupSelectorNode{}))
	})

	t.Run("updates existing selector nodes by selector ID and node ID", func(t *testing.T) {
		var (
			assetGroupTagId = 1
			selectorOne     = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch update selector one")
			selectorTwo     = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch update selector two")
			sharedNodeId    = graph.ID(10001)
			otherNodeId     = graph.ID(10002)

			oldSelectorOneSharedNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorOne.ID,
				NodeId:            sharedNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "old-primary-kind-1",
				NodeEnvironmentId: "old-environment-1",
				NodeObjectId:      "old-object-id-1",
				NodeName:          "old-node-name-1",
			}
			oldSelectorOneOtherNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorOne.ID,
				NodeId:            otherNodeId,
				Certified:         model.AssetGroupCertificationRevoked,
				CertifiedBy:       null.StringFrom("old-certifier-2"),
				Source:            model.AssetGroupSelectorNodeSourceChild,
				NodePrimaryKind:   "old-primary-kind-2",
				NodeEnvironmentId: "old-environment-2",
				NodeObjectId:      "old-object-id-2",
				NodeName:          "old-node-name-2",
			}
			untouchedSelectorTwoSharedNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorTwo.ID,
				NodeId:            sharedNodeId,
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       null.StringFrom("untouched-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "untouched-primary-kind",
				NodeEnvironmentId: "untouched-environment",
				NodeObjectId:      "untouched-object-id",
				NodeName:          "untouched-node-name",
			}
			updatedSelectorOneSharedNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorOne.ID,
				NodeId:            sharedNodeId,
				Certified:         model.AssetGroupCertificationAuto,
				CertifiedBy:       null.StringFrom(model.AssetGroupActorBloodHound),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "new-primary-kind-1",
				NodeEnvironmentId: "new-environment-1",
				NodeObjectId:      "new-object-id-1",
				NodeName:          "new-node-name-1",
			}
			updatedSelectorOneOtherNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorOne.ID,
				NodeId:            otherNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "new-primary-kind-2",
				NodeEnvironmentId: "new-environment-2",
				NodeObjectId:      "new-object-id-2",
				NodeName:          "new-node-name-2",
			}
		)

		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, oldSelectorOneSharedNode)
		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, oldSelectorOneOtherNode)
		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, untouchedSelectorTwoSharedNode)

		require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(testCtx, []model.AssetGroupSelectorNode{
			updatedSelectorOneSharedNode,
			updatedSelectorOneOtherNode,
		}))

		selectorNodesBySelectorId := requireSelectorNodesBySelectorAndNodeId(t, testCtx, suite.BHDatabase, selectorOne.ID, selectorTwo.ID)
		assertUpdateSelectorNodesTestNode(t, updatedSelectorOneSharedNode, selectorNodesBySelectorId[selectorOne.ID][sharedNodeId])
		assertUpdateSelectorNodesTestNode(t, updatedSelectorOneOtherNode, selectorNodesBySelectorId[selectorOne.ID][otherNodeId])
		assertUpdateSelectorNodesTestNode(t, untouchedSelectorTwoSharedNode, selectorNodesBySelectorId[selectorTwo.ID][sharedNodeId])
	})

	t.Run("creates asset group history records when certifications change", func(t *testing.T) {
		var (
			assetGroupTagId = 1
			selector        = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch update history changes")
			autoNodeId      = graph.ID(30001)
			manualNodeId    = graph.ID(30002)
			revokedNodeId   = graph.ID(30003)

			oldAutoNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            autoNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "old-auto-primary-kind",
				NodeEnvironmentId: "old-auto-environment",
				NodeObjectId:      "old-auto-object-id",
				NodeName:          "old-history-auto-node",
			}
			oldManualNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            manualNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "old-manual-primary-kind",
				NodeEnvironmentId: "old-manual-environment",
				NodeObjectId:      "old-manual-object-id",
				NodeName:          "old-history-manual-node",
			}
			oldRevokedNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            revokedNodeId,
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       null.StringFrom("previous-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceChild,
				NodePrimaryKind:   "old-revoked-primary-kind",
				NodeEnvironmentId: "old-revoked-environment",
				NodeObjectId:      "old-revoked-object-id",
				NodeName:          "old-history-revoked-node",
			}
			updatedAutoNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            autoNodeId,
				Certified:         model.AssetGroupCertificationAuto,
				CertifiedBy:       null.StringFrom(model.AssetGroupActorBloodHound),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "new-auto-primary-kind",
				NodeEnvironmentId: "new-auto-environment",
				NodeObjectId:      "new-auto-object-id",
				NodeName:          "history-auto-node",
			}
			updatedManualNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            manualNodeId,
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       null.StringFrom("manual-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "new-manual-primary-kind",
				NodeEnvironmentId: "new-manual-environment",
				NodeObjectId:      "new-manual-object-id",
				NodeName:          "history-manual-node",
			}
			updatedRevokedNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            revokedNodeId,
				Certified:         model.AssetGroupCertificationRevoked,
				CertifiedBy:       null.StringFrom("revoked-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "new-revoked-primary-kind",
				NodeEnvironmentId: "new-revoked-environment",
				NodeObjectId:      "new-revoked-object-id",
				NodeName:          "history-revoked-node",
			}
		)

		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, oldAutoNode)
		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, oldManualNode)
		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, oldRevokedNode)

		require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(testCtx, []model.AssetGroupSelectorNode{
			updatedAutoNode,
			updatedManualNode,
			updatedRevokedNode,
		}))

		historyRecordsByTarget := requireUpdateSelectorNodesHistoryRecordsByTarget(t, testCtx, suite.BHDatabase, updatedAutoNode.NodeName, updatedManualNode.NodeName, updatedRevokedNode.NodeName)
		require.Len(t, historyRecordsByTarget, 3)
		assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeAuto, updatedAutoNode.NodeName, assetGroupTagId, historyRecordsByTarget[updatedAutoNode.NodeName])
		assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeManual, updatedManualNode.NodeName, assetGroupTagId, historyRecordsByTarget[updatedManualNode.NodeName])
		assertUpdateSelectorNodesHistoryRecord(t, model.AssetGroupHistoryActionCertifyNodeRevoked, updatedRevokedNode.NodeName, assetGroupTagId, historyRecordsByTarget[updatedRevokedNode.NodeName])
	})

	t.Run("does not create asset group history records when certifications are unchanged", func(t *testing.T) {
		var (
			assetGroupTagId = 1
			selector        = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch update history unchanged")
			unchangedNodeId = graph.ID(40001)
			missingNodeId   = graph.ID(40002)

			oldUnchangedNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            unchangedNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "old-unchanged-primary-kind",
				NodeEnvironmentId: "old-unchanged-environment",
				NodeObjectId:      "old-unchanged-object-id",
				NodeName:          "old-history-unchanged-node",
			}
			updatedUnchangedNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            unchangedNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "new-unchanged-primary-kind",
				NodeEnvironmentId: "new-unchanged-environment",
				NodeObjectId:      "new-unchanged-object-id",
				NodeName:          "history-unchanged-node",
			}
			missingNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            missingNodeId,
				Certified:         model.AssetGroupCertificationAuto,
				CertifiedBy:       null.StringFrom(model.AssetGroupActorBloodHound),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "missing-history-primary-kind",
				NodeEnvironmentId: "missing-history-environment",
				NodeObjectId:      "missing-history-object-id",
				NodeName:          "history-missing-node",
			}
		)

		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, oldUnchangedNode)

		require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(testCtx, []model.AssetGroupSelectorNode{
			updatedUnchangedNode,
			missingNode,
		}))

		historyRecordsByTarget := requireUpdateSelectorNodesHistoryRecordsByTarget(t, testCtx, suite.BHDatabase, updatedUnchangedNode.NodeName, missingNode.NodeName)
		require.Empty(t, historyRecordsByTarget)
	})

	t.Run("ignores selector nodes that do not already exist", func(t *testing.T) {
		var (
			assetGroupTagId = 1
			selector        = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch update missing row")
			existingNodeId  = graph.ID(20001)
			missingNodeId   = graph.ID(20002)

			oldExistingNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            existingNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "old-existing-primary-kind",
				NodeEnvironmentId: "old-existing-environment",
				NodeObjectId:      "old-existing-object-id",
				NodeName:          "old-existing-node-name",
			}
			updatedExistingNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            existingNodeId,
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       null.StringFrom("existing-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceChild,
				NodePrimaryKind:   "new-existing-primary-kind",
				NodeEnvironmentId: "new-existing-environment",
				NodeObjectId:      "new-existing-object-id",
				NodeName:          "new-existing-node-name",
			}
			missingNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            missingNodeId,
				Certified:         model.AssetGroupCertificationAuto,
				CertifiedBy:       null.StringFrom(model.AssetGroupActorBloodHound),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "missing-primary-kind",
				NodeEnvironmentId: "missing-environment",
				NodeObjectId:      "missing-object-id",
				NodeName:          "missing-node-name",
			}
		)

		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, oldExistingNode)

		require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(testCtx, []model.AssetGroupSelectorNode{
			updatedExistingNode,
			missingNode,
		}))

		selectorNodesBySelectorId := requireSelectorNodesBySelectorAndNodeId(t, testCtx, suite.BHDatabase, selector.ID)
		require.Len(t, selectorNodesBySelectorId[selector.ID], 1)
		assertUpdateSelectorNodesTestNode(t, updatedExistingNode, selectorNodesBySelectorId[selector.ID][existingNodeId])
		assert.NotContains(t, selectorNodesBySelectorId[selector.ID], missingNodeId)
	})
}

func TestDatabase_DeleteSelectorNodes(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx   = context.Background()
		testActor = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
	)

	t.Run("returns nil for empty input", func(t *testing.T) {
		require.NoError(t, suite.BHDatabase.DeleteSelectorNodes(testCtx, nil))
		require.NoError(t, suite.BHDatabase.DeleteSelectorNodes(testCtx, []model.AssetGroupSelectorNode{}))
	})

	t.Run("deletes requested selector nodes", func(t *testing.T) {
		var (
			assetGroupTagId = 1
			selector        = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch delete requested nodes")
			deletedNodeId   = graph.ID(70001)
			retainedNodeId  = graph.ID(70002)

			deletedNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            deletedNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "delete-requested-primary-kind",
				NodeEnvironmentId: "delete-requested-environment",
				NodeObjectId:      "delete-requested-object-id",
				NodeName:          "delete-requested-node",
			}
			retainedNode = model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            retainedNodeId,
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       null.StringFrom("retained-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceChild,
				NodePrimaryKind:   "delete-retained-primary-kind",
				NodeEnvironmentId: "delete-retained-environment",
				NodeObjectId:      "delete-retained-object-id",
				NodeName:          "delete-retained-node",
			}
		)

		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, deletedNode)
		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, retainedNode)

		require.NoError(t, suite.BHDatabase.DeleteSelectorNodes(testCtx, []model.AssetGroupSelectorNode{
			deletedNode,
		}))

		selectorNodesBySelectorId := requireSelectorNodesBySelectorAndNodeId(t, testCtx, suite.BHDatabase, selector.ID)
		require.Len(t, selectorNodesBySelectorId[selector.ID], 1)
		assert.NotContains(t, selectorNodesBySelectorId[selector.ID], deletedNodeId)
		assertUpdateSelectorNodesTestNode(t, retainedNode, selectorNodesBySelectorId[selector.ID][retainedNodeId])
	})

	t.Run("leaves selector nodes not passed in untouched", func(t *testing.T) {
		var (
			assetGroupTagId           = 1
			selectorOne               = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch delete untouched selector one")
			selectorTwo               = createUpdateSelectorNodesTestSelector(t, testCtx, suite.BHDatabase, testActor, "batch delete untouched selector two")
			sharedNodeId              = graph.ID(80001)
			selectorOneRetainedNodeId = graph.ID(80002)

			deletedSelectorOneSharedNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorOne.ID,
				NodeId:            sharedNodeId,
				Certified:         model.AssetGroupCertificationAuto,
				CertifiedBy:       null.StringFrom(model.AssetGroupActorBloodHound),
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "delete-selector-one-primary-kind",
				NodeEnvironmentId: "delete-selector-one-environment",
				NodeObjectId:      "delete-selector-one-object-id",
				NodeName:          "delete-selector-one-shared-node",
			}
			retainedSelectorOneNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorOne.ID,
				NodeId:            selectorOneRetainedNodeId,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceChild,
				NodePrimaryKind:   "delete-selector-one-retained-primary-kind",
				NodeEnvironmentId: "delete-selector-one-retained-environment",
				NodeObjectId:      "delete-selector-one-retained-object-id",
				NodeName:          "delete-selector-one-retained-node",
			}
			retainedSelectorTwoSharedNode = model.AssetGroupSelectorNode{
				SelectorId:        selectorTwo.ID,
				NodeId:            sharedNodeId,
				Certified:         model.AssetGroupCertificationRevoked,
				CertifiedBy:       null.StringFrom("selector-two-certifier"),
				Source:            model.AssetGroupSelectorNodeSourceParent,
				NodePrimaryKind:   "delete-selector-two-primary-kind",
				NodeEnvironmentId: "delete-selector-two-environment",
				NodeObjectId:      "delete-selector-two-object-id",
				NodeName:          "delete-selector-two-shared-node",
			}
		)

		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, deletedSelectorOneSharedNode)
		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, retainedSelectorOneNode)
		insertUpdateSelectorNodesTestNode(t, testCtx, suite.BHDatabase, assetGroupTagId, retainedSelectorTwoSharedNode)

		require.NoError(t, suite.BHDatabase.DeleteSelectorNodes(testCtx, []model.AssetGroupSelectorNode{
			deletedSelectorOneSharedNode,
		}))

		selectorNodesBySelectorId := requireSelectorNodesBySelectorAndNodeId(t, testCtx, suite.BHDatabase, selectorOne.ID, selectorTwo.ID)
		require.Len(t, selectorNodesBySelectorId[selectorOne.ID], 1)
		require.Len(t, selectorNodesBySelectorId[selectorTwo.ID], 1)
		assert.NotContains(t, selectorNodesBySelectorId[selectorOne.ID], sharedNodeId)
		assertUpdateSelectorNodesTestNode(t, retainedSelectorOneNode, selectorNodesBySelectorId[selectorOne.ID][selectorOneRetainedNodeId])
		assertUpdateSelectorNodesTestNode(t, retainedSelectorTwoSharedNode, selectorNodesBySelectorId[selectorTwo.ID][sharedNodeId])
	})
}

func createUpdateSelectorNodesTestSelector(t *testing.T, ctx context.Context, bloodhoundDatabase *database.BloodhoundDB, testActor model.User, name string) model.AssetGroupTagSelector {
	t.Helper()

	selector, err := bloodhoundDatabase.CreateAssetGroupTagSelector(ctx, 1, testActor, name, "test description", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{
		{Type: model.SelectorTypeObjectId, Value: name + " object id"},
	})
	require.NoError(t, err)

	return selector
}

func insertUpdateSelectorNodesTestNode(t *testing.T, ctx context.Context, bloodhoundDatabase *database.BloodhoundDB, assetGroupTagId int, node model.AssetGroupSelectorNode) {
	t.Helper()

	err := bloodhoundDatabase.InsertSelectorNode(
		ctx,
		assetGroupTagId,
		node.SelectorId,
		node.NodeId,
		node.Certified,
		node.CertifiedBy,
		node.Source,
		node.NodePrimaryKind,
		node.NodeEnvironmentId,
		node.NodeObjectId,
		node.NodeName,
	)
	require.NoError(t, err)
}

func requireSelectorNodesBySelectorAndNodeId(t *testing.T, ctx context.Context, bloodhoundDatabase *database.BloodhoundDB, selectorIds ...int) map[int]map[graph.ID]model.AssetGroupSelectorNode {
	t.Helper()

	var (
		selectorNodesBySelectorId = make(map[int]map[graph.ID]model.AssetGroupSelectorNode)
		selectorNodes, err        = bloodhoundDatabase.GetSelectorNodesBySelectorIds(ctx, selectorIds...)
	)
	require.NoError(t, err)

	for _, selectorNode := range selectorNodes {
		if _, ok := selectorNodesBySelectorId[selectorNode.SelectorId]; !ok {
			selectorNodesBySelectorId[selectorNode.SelectorId] = make(map[graph.ID]model.AssetGroupSelectorNode)
		}

		selectorNodesBySelectorId[selectorNode.SelectorId][selectorNode.NodeId] = selectorNode
	}

	return selectorNodesBySelectorId
}

func assertUpdateSelectorNodesTestNode(t *testing.T, expectedNode model.AssetGroupSelectorNode, actualNode model.AssetGroupSelectorNode) {
	t.Helper()

	assert.Equal(t, expectedNode.SelectorId, actualNode.SelectorId)
	assert.Equal(t, expectedNode.NodeId, actualNode.NodeId)
	assert.Equal(t, expectedNode.Certified, actualNode.Certified)
	assert.Equal(t, expectedNode.CertifiedBy, actualNode.CertifiedBy)
	assert.Equal(t, expectedNode.Source, actualNode.Source)
	assert.Equal(t, expectedNode.NodePrimaryKind, actualNode.NodePrimaryKind)
	assert.Equal(t, expectedNode.NodeEnvironmentId, actualNode.NodeEnvironmentId)
	assert.Equal(t, expectedNode.NodeObjectId, actualNode.NodeObjectId)
	assert.Equal(t, expectedNode.NodeName, actualNode.NodeName)
	assert.False(t, actualNode.UpdatedAt.IsZero())
}

func requireUpdateSelectorNodesHistoryRecordsByTarget(t *testing.T, ctx context.Context, bloodhoundDatabase *database.BloodhoundDB, targets ...string) map[string]model.AssetGroupHistory {
	t.Helper()

	var (
		historyRecordsByTarget = make(map[string]model.AssetGroupHistory)
		targetsByName          = make(map[string]struct{}, len(targets))
		historyRecords, _, err = bloodhoundDatabase.GetAssetGroupHistoryRecords(ctx, model.SQLFilter{}, model.Sort{{Column: "created_at", Direction: model.AscendingSortDirection}}, 0, 0)
	)
	require.NoError(t, err)

	for _, target := range targets {
		targetsByName[target] = struct{}{}
	}

	for _, historyRecord := range historyRecords {
		if _, ok := targetsByName[historyRecord.Target]; ok {
			require.NotContains(t, historyRecordsByTarget, historyRecord.Target)
			historyRecordsByTarget[historyRecord.Target] = historyRecord
		}
	}

	return historyRecordsByTarget
}

func assertUpdateSelectorNodesHistoryRecord(t *testing.T, expectedAction model.AssetGroupHistoryAction, expectedTarget string, expectedAssetGroupTagId int, actualRecord model.AssetGroupHistory) {
	t.Helper()

	assert.Equal(t, model.AssetGroupActorBloodHound, actualRecord.Actor)
	assert.Equal(t, "", actualRecord.Email.ValueOrZero())
	assert.Equal(t, expectedAction, actualRecord.Action)
	assert.Equal(t, expectedTarget, actualRecord.Target)
	assert.Equal(t, expectedAssetGroupTagId, actualRecord.AssetGroupTagId)
	assert.Equal(t, null.String{}, actualRecord.EnvironmentId)
	assert.Equal(t, null.String{}, actualRecord.Note)
	assert.False(t, actualRecord.CreatedAt.IsZero())
}

func TestDatabase_GetAssetGroupSelectorNodeExpandedOrderedByIdAndPosition(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx         = context.Background()
		testActor       = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
		testName        = "test selector name"
		testDescription = "test description"
		certifiedBy     = null.StringFrom("testy")
		testObjectId1   = 1
		testObjectId2   = 2
		testNodeId1     = uint64(testObjectId1)
		testNodeId2     = uint64(testObjectId2)
		source          = 1
		isDefault       = false
		allowDisable    = true
		autoCertify     = model.SelectorAutoCertifyMethodDisabled
		testSeeds       = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectID5678"},
		}
	)

	tierTag, err := suite.BHDatabase.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, model.User{}, "tier 1", testDescription, null.Int32From(2), null.BoolFrom(false), null.String{})
	require.NoError(t, err)

	labelTag, err := suite.BHDatabase.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, model.User{}, "label", testDescription, null.Int32{}, null.Bool{}, null.String{})
	require.NoError(t, err)

	selector, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, tierTag.ID, testActor, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)

	err = suite.BHDatabase.InsertSelectorNode(testCtx, tierTag.ID, selector.ID, graph.ID(testNodeId1), model.AssetGroupCertificationManual, certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)

	err = suite.BHDatabase.InsertSelectorNode(testCtx, tierTag.ID, selector.ID, graph.ID(testNodeId2), model.AssetGroupCertificationManual, certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)

	err = suite.BHDatabase.InsertSelectorNode(testCtx, tierTag.ID, selector.ID, graph.ID(testNodeId1), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)

	err = suite.BHDatabase.InsertSelectorNode(testCtx, labelTag.ID, selector.ID, graph.ID(testNodeId1), model.AssetGroupCertificationManual, certifiedBy, model.AssetGroupSelectorNodeSource(source), "", "", "", "")
	require.NoError(t, err)

	tests := map[string]struct {
		Input          []int
		ExpectedLength int
	}{
		"returns all selected nodes, ignores tags": {
			Input:          []int{testObjectId1, testObjectId2},
			ExpectedLength: 2,
		},
		"returns empty slice if no nodes found": {
			Input:          []int{3},
			ExpectedLength: 0,
		},
		"returns empty slice if no node IDs passed in": {
			Input:          nil,
			ExpectedLength: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got, err := suite.BHDatabase.GetAssetGroupSelectorNodeExpandedOrderedByIdAndPosition(testCtx, test.Input...); err != nil {
				t.Fatalf("error getting selector node expanded ignore autocertify: %v", err)
			} else if len(got) != test.ExpectedLength {
				t.Fatalf("got %d expected %d", len(got), test.ExpectedLength)
			}
		})
	}
}

func TestDatabase_GetAggregatedSelectorNodesCertification(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx         = context.Background()
		testActor       = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
		testCertifiedBy = "testy"
		certifiedBy     = null.StringFrom(testCertifiedBy)
		source          = 1
	)

	// Enable require certify on tier 0
	tier0, err := suite.BHDatabase.GetAssetGroupTag(testCtx, 1)
	require.NoError(t, err)
	tier0.RequireCertify = null.BoolFrom(true)
	tier0, err = suite.BHDatabase.UpdateAssetGroupTag(testCtx, model.User{}, tier0)
	require.NoError(t, err)

	// create a test zone at position 2
	tier1, err := suite.BHDatabase.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "Test T1 Zone", "Test zone description", null.Int32From(2), null.BoolFrom(true), null.String{})
	require.NoError(t, err)

	// create a test zone at position 3
	tier2, err := suite.BHDatabase.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "Test T2 Zone", "Test zone description", null.Int32From(3), null.BoolFrom(true), null.String{})
	require.NoError(t, err)

	// Zone 0 is added by the migration and is ID 1
	sel0, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, "Test T0 selector", "description", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)

	sel0_1, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, "Test T0 selector number 2", "description", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)

	sel0_2, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, "Test T0 selector number 3", "description", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)

	sel0_3, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, 1, model.User{}, "Test T0 selector number 4", "description", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)

	sel1, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, tier1.ID, model.User{}, "Test T1 selector", "description", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)

	sel2, err := suite.BHDatabase.CreateAssetGroupTagSelector(testCtx, tier2.ID, model.User{}, "Test T2 selector", "description", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{})
	require.NoError(t, err)

	t.Run("Multiple nodes - verify highest tier wins", func(t *testing.T) {
		testNodeId := uint64(1)

		// a selector for T0 selects this node
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(testNodeId), model.AssetGroupCertificationPending, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0")
		require.NoError(t, err)

		// a selector for T1 also selects this node
		err = suite.BHDatabase.InsertSelectorNode(testCtx, tier1.ID, sel1.ID, graph.ID(testNodeId), model.AssetGroupCertificationPending, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_1", "environment", "objid", "NodeSelectedByT1")
		require.NoError(t, err)

		// a selector for T2 also selects this node
		err = suite.BHDatabase.InsertSelectorNode(testCtx, tier2.ID, sel2.ID, graph.ID(testNodeId), model.AssetGroupCertificationPending, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_2", "environment", "objid", "NodeSelectedByT2")
		require.NoError(t, err)

		// filtering on the nodes from this test only
		nodeCertifications, count, err := suite.BHDatabase.GetAggregatedSelectorNodesCertification(testCtx, model.SQLFilter{SQLString: "node_id = 1"}, 0, 0)
		require.NoError(t, err)

		// there should only be a single node returned
		require.Equal(t, 1, count)
		// it should be the one associated with T0
		require.Equal(t, "NodeSelectedByT0", nodeCertifications[0].NodeName)
	})

	t.Run("Multiple nodes - verify correct hybrid output", func(t *testing.T) {
		testNodeId := uint64(30)

		// a selector for T0 selects this node
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(testNodeId), model.AssetGroupCertificationPending, null.StringFrom("Sel0_1_NoCertifier"), model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0")
		require.NoError(t, err)

		timeBeforeSel0_1_NodeInserted := time.Now().UTC()

		// another selector for T0 selects this node
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0_1.ID, graph.ID(testNodeId), model.AssetGroupCertificationManual, null.StringFrom("Sel0_1_ManualCertifier"), model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0")
		require.NoError(t, err)

		// a selector for T1 also selects this node
		err = suite.BHDatabase.InsertSelectorNode(testCtx, tier1.ID, sel1.ID, graph.ID(testNodeId), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_1", "environment", "objid", "NodeSelectedByT1")
		require.NoError(t, err)

		// a selector for T2 also selects this node
		err = suite.BHDatabase.InsertSelectorNode(testCtx, tier2.ID, sel2.ID, graph.ID(testNodeId), model.AssetGroupCertificationPending, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_2", "environment", "objid", "NodeSelectedByT2")
		require.NoError(t, err)

		// filtering on the nodes from this test only
		nodeCertifications, count, err := suite.BHDatabase.GetAggregatedSelectorNodesCertification(testCtx, model.SQLFilter{SQLString: "node_id = 30"}, 0, 0)
		require.NoError(t, err)

		// there should only be a single node returned
		require.Equal(t, 1, count)
		// it should have the highest certification value associated with T0
		require.Equal(t, model.AssetGroupCertificationManual, nodeCertifications[0].Certified)
		// it should be associated with T0
		require.Equal(t, 1, nodeCertifications[0].AssetGroupTagId)
		// it should have a timestamp of the first node inserted for this tier
		require.True(t, nodeCertifications[0].CreatedAt.Before(timeBeforeSel0_1_NodeInserted))
	})

	t.Run("Multiple nodes - verify highest certify wins", func(t *testing.T) {
		testNodeId := uint64(2)
		// this one has certification revoked
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(testNodeId), model.AssetGroupCertificationRevoked, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_CertRevoked")
		require.NoError(t, err)

		// no certification
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0_1.ID, graph.ID(testNodeId), model.AssetGroupCertificationPending, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_CertNone")
		require.NoError(t, err)

		// manual certification
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0_2.ID, graph.ID(testNodeId), model.AssetGroupCertificationManual, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_CertManual")
		require.NoError(t, err)

		// auto certification
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0_3.ID, graph.ID(testNodeId), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_CertAuto")
		require.NoError(t, err)

		// filtering on the nodes from this test only
		nodeCertifications, count, err := suite.BHDatabase.GetAggregatedSelectorNodesCertification(testCtx, model.SQLFilter{SQLString: "node_id = 2"}, 0, 0)
		require.NoError(t, err)

		// there should only be a single node returned
		require.Equal(t, 1, count)
		// it should be the one that had auto certify
		require.Equal(t, "NodeSelectedByT0_CertAuto", nodeCertifications[0].NodeName)
		require.Equal(t, model.AssetGroupCertificationAuto, nodeCertifications[0].Certified)
	})

	t.Run("Multiple nodes - verify earliest create date wins", func(t *testing.T) {
		testNodeId := uint64(3)
		// created first
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(testNodeId), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_First")
		require.NoError(t, err)

		// created second
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0_1.ID, graph.ID(testNodeId), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_Second")
		require.NoError(t, err)

		// created third
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0_2.ID, graph.ID(testNodeId), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_Third")
		require.NoError(t, err)

		// created fourth
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0_3.ID, graph.ID(testNodeId), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_Fourth")
		require.NoError(t, err)

		// filtering on the nodes from this test only
		nodeCertifications, count, err := suite.BHDatabase.GetAggregatedSelectorNodesCertification(testCtx, model.SQLFilter{SQLString: "node_id = 3"}, 0, 0)
		require.NoError(t, err)

		// there should only be a single node returned
		require.Equal(t, 1, count)
		// it should be the one created first
		require.Equal(t, "NodeSelectedByT0_First", nodeCertifications[0].NodeName)
	})

	t.Run("test pagination", func(t *testing.T) {
		// create some nodes, all selected by a t0 selector
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(uint64(10)), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_First")
		require.NoError(t, err)

		// created second
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(uint64(11)), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_Second")
		require.NoError(t, err)

		// created third
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(uint64(12)), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_Third")
		require.NoError(t, err)

		// created fourth
		err = suite.BHDatabase.InsertSelectorNode(testCtx, 1, sel0.ID, graph.ID(uint64(13)), model.AssetGroupCertificationAuto, certifiedBy, model.AssetGroupSelectorNodeSource(source), "kind_0", "environment", "objid", "NodeSelectedByT0_Fourth")
		require.NoError(t, err)

		// filtering on the nodes from this test only
		nodeCertifications, count, err := suite.BHDatabase.GetAggregatedSelectorNodesCertification(testCtx, model.SQLFilter{SQLString: "node_id BETWEEN 10 AND 13"}, 0, 0)
		require.NoError(t, err)

		// verify 4 out of 4 nodes returned
		require.Equal(t, 4, len(nodeCertifications))
		require.Equal(t, 4, count)

		// skip the first 2
		nodeCertifications, count, err = suite.BHDatabase.GetAggregatedSelectorNodesCertification(testCtx, model.SQLFilter{SQLString: "node_id BETWEEN 10 AND 13"}, 2, 0)
		require.NoError(t, err)

		// verify 2 out of 4 nodes returned
		require.Equal(t, 2, len(nodeCertifications))
		require.Equal(t, 4, count)
		require.Equal(t, "NodeSelectedByT0_Third", nodeCertifications[0].NodeName)

		// limit to 2
		nodeCertifications, count, err = suite.BHDatabase.GetAggregatedSelectorNodesCertification(testCtx, model.SQLFilter{SQLString: "node_id BETWEEN 10 AND 13"}, 0, 2)
		require.NoError(t, err)

		// verify 2 out of 4 nodes returned
		require.Equal(t, 2, len(nodeCertifications))
		require.Equal(t, 4, count)
		require.Equal(t, "NodeSelectedByT0_First", nodeCertifications[0].NodeName)
	})
}
