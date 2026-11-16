// Copyright 2026 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestDatabase_GetAssetGroupTagBySystemKey(t *testing.T) {
	var (
		suite = setupIntegrationTestSuite(t)
	)
	defer teardownIntegrationTestSuite(t, &suite)

	tag, err := suite.BHDatabase.CreateAssetGroupTag(suite.Context, model.AssetGroupTagTypeLabel, model.User{}, "System label", "", null.Int32{}, null.Bool{}, null.String{})
	require.NoError(t, err)
	require.False(t, tag.SystemKey.Valid)
	require.NoError(t, suite.DB.Model(&model.AssetGroupTag{}).Where("id = ?", tag.ID).Update("system_key", "test_system_label").Error)

	systemTag, err := suite.BHDatabase.GetAssetGroupTagBySystemKey(suite.Context, "test_system_label")
	require.NoError(t, err)
	require.Equal(t, tag.ID, systemTag.ID)
	require.Equal(t, "test_system_label", systemTag.SystemKey.ValueOrZero())
	tagByID, err := suite.BHDatabase.GetAssetGroupTag(suite.Context, tag.ID)
	require.NoError(t, err)
	require.Equal(t, systemTag, tagByID)
	tags, err := suite.BHDatabase.GetAssetGroupTags(suite.Context, model.SQLFilter{SQLString: "id = ?", Params: []any{tag.ID}})
	require.NoError(t, err)
	require.Equal(t, model.AssetGroupTags{systemTag}, tags)

	systemTag.Name = "Renamed system label"
	_, err = suite.BHDatabase.UpdateAssetGroupTag(suite.Context, model.User{}, systemTag)
	require.NoError(t, err)
	renamedTag, err := suite.BHDatabase.GetAssetGroupTagBySystemKey(suite.Context, "test_system_label")
	require.NoError(t, err)
	require.Equal(t, tag.ID, renamedTag.ID)
	require.Equal(t, "Renamed system label", renamedTag.Name)

	_, err = suite.BHDatabase.GetAssetGroupTagBySystemKey(suite.Context, "missing")
	require.ErrorIs(t, err, database.ErrNotFound)
	_, err = suite.BHDatabase.GetAssetGroupTagBySystemKey(suite.Context, "")
	require.ErrorIs(t, err, database.ErrNotFound)

	canceledContext, cancel := context.WithCancel(suite.Context)
	cancel()
	_, err = suite.BHDatabase.GetAssetGroupTagBySystemKey(canceledContext, "test_system_label")
	require.ErrorIs(t, err, context.Canceled)

	require.NoError(t, suite.BHDatabase.DeleteAssetGroupTag(suite.Context, model.User{}, renamedTag))
	_, err = suite.BHDatabase.GetAssetGroupTagBySystemKey(suite.Context, "test_system_label")
	require.ErrorIs(t, err, database.ErrNotFound)
}

func TestDatabase_SystemKeyInTierQueries(t *testing.T) {
	var (
		suite = setupIntegrationTestSuite(t)
	)
	defer teardownIntegrationTestSuite(t, &suite)

	tiers, err := suite.BHDatabase.GetOrderedAssetGroupTagTiers(suite.Context)
	require.NoError(t, err)
	require.NotEmpty(t, tiers)
	require.NoError(t, suite.DB.Model(&model.AssetGroupTag{}).Where("id = ?", tiers[0].ID).Update("system_key", "test_system_tier").Error)

	tiers, err = suite.BHDatabase.GetOrderedAssetGroupTagTiers(suite.Context)
	require.NoError(t, err)
	require.Equal(t, "test_system_tier", tiers[0].SystemKey.ValueOrZero())
	selectedTags, err := suite.BHDatabase.GetAssetGroupTagForSelection(suite.Context)
	require.NoError(t, err)
	for _, selectedTag := range selectedTags {
		if selectedTag.ID == tiers[0].ID {
			require.Equal(t, "test_system_tier", selectedTag.SystemKey.ValueOrZero())
			return
		}
	}
	t.Fatal("Tier Zero was not returned for selection")
}
