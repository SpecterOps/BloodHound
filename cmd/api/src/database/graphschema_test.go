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

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateAndGetGraphSchemaExtensions(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx = context.Background()
		ext1    = model.GraphSchemaExtension{
			Name:        "test_name1",
			DisplayName: "test extension name 1",
			Version:     "1.0.0",
		}
		ext2 = model.GraphSchemaExtension{
			Name:        "test_name2",
			DisplayName: "test extension name 2",
			Version:     "1.0.0",
		}
	)

	extension1, err := suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext1.Name, ext1.DisplayName, ext1.Version)
	require.NoError(t, err)
	require.Equal(t, ext1.Name, extension1.Name)
	require.Equal(t, ext1.DisplayName, extension1.DisplayName)
	require.Equal(t, ext1.Version, extension1.Version)

	_, err = suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext1.Name, ext1.DisplayName, ext1.Version)
	require.Error(t, err)
	require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionName)

	extension2, err := suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext2.Name, ext2.DisplayName, ext2.Version)
	require.NoError(t, err)
	require.Equal(t, ext2.Name, extension2.Name)
	require.Equal(t, ext2.DisplayName, extension2.DisplayName)
	require.Equal(t, ext2.Version, extension2.Version)

	got, err := suite.BHDatabase.GetGraphSchemaExtensionById(testCtx, extension1.ID)
	require.NoError(t, err)
	require.Equal(t, extension1.Name, got.Name)
	require.Equal(t, extension1.DisplayName, got.DisplayName)
	require.Equal(t, extension1.Version, got.Version)
	require.Equal(t, false, got.IsBuiltin)
	require.Equal(t, false, got.CreatedAt.IsZero())
	require.Equal(t, false, got.UpdatedAt.IsZero())
	require.Equal(t, false, got.DeletedAt.Valid)

	_, err = suite.BHDatabase.GetGraphSchemaExtensionById(testCtx, 1234)
	require.Error(t, err)
	require.Equal(t, "entity not found", err.Error())
}

func TestDatabase_GetGraphSchemaExtensionsFilteredAndPaginated(t *testing.T) {
	t.Parallel()
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx = context.Background()
		ext1    = model.GraphSchemaExtension{
			Name:        "adam",
			DisplayName: "test extension name 1",
			Version:     "1.0.0",
		}
		ext2 = model.GraphSchemaExtension{
			Name:        "bob",
			DisplayName: "test extension name 2",
			Version:     "2.0.0",
		}
		ext3 = model.GraphSchemaExtension{
			Name:        "charlie",
			DisplayName: "another extension",
			Version:     "3.0.0",
		}
		ext4 = model.GraphSchemaExtension{
			Name:        "david",
			DisplayName: "yet another extension",
			Version:     "4.0.0",
		}
	)

	_, err := suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext1.Name, ext1.DisplayName, ext1.Version)
	require.NoError(t, err)

	_, err = suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext2.Name, ext2.DisplayName, ext2.Version)
	require.NoError(t, err)

	_, err = suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext3.Name, ext3.DisplayName, ext3.Version)
	require.NoError(t, err)

	_, err = suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext4.Name, ext4.DisplayName, ext4.Version)
	require.NoError(t, err)

	t.Run("successfully returns an array of extensions, no filtering or sorting", func(t *testing.T) {
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensionsFilteredAndPaginated(testCtx, model.SQLFilter{}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 4)
		require.Equal(t, 4, total)
	})

	t.Run("successfully returns an array of extensions, with filtering", func(t *testing.T) {
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensionsFilteredAndPaginated(testCtx, model.SQLFilter{SQLString: "name = ?", Params: []any{"david"}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 1)
		require.Equal(t, 1, total)
	})

	t.Run("successfully returns an array of extensions, with fuzzy filtering", func(t *testing.T) {
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensionsFilteredAndPaginated(testCtx, model.SQLFilter{SQLString: "display_name ILIKE ?", Params: []any{"%test extension%"}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, 2, total)
	})

	t.Run("successfully returns an array of extensions, with fuzzy filtering and sort ascending", func(t *testing.T) {
		extensions, _, err := suite.BHDatabase.GetGraphSchemaExtensionsFilteredAndPaginated(testCtx, model.SQLFilter{SQLString: "display_name ILIKE ?", Params: []any{"%test extension%"}}, model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, "adam", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, with fuzzy filtering and sort descending", func(t *testing.T) {
		extensions, _, err := suite.BHDatabase.GetGraphSchemaExtensionsFilteredAndPaginated(testCtx, model.SQLFilter{SQLString: "display_name ILIKE ?", Params: []any{"%test extension%"}}, model.Sort{{Column: "display_name", Direction: model.DescendingSortDirection}}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, "bob", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, no filtering or sorting, with skip", func(t *testing.T) {
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensionsFilteredAndPaginated(testCtx, model.SQLFilter{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, 4, total)
		require.Equal(t, "charlie", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, no filtering or sorting, with limit", func(t *testing.T) {
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensionsFilteredAndPaginated(testCtx, model.SQLFilter{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, 4, total)
		require.Equal(t, "bob", extensions[1].Name)
	})
}
