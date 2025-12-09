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

func TestDatabase_GraphSchemaExtensions_CRUD(t *testing.T) {
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
		actualExtension1 = model.GraphSchemaExtension{}
		actualExtension2 = model.GraphSchemaExtension{}
		err              error
	)

	t.Run("success - create a graph schema extension", func(t *testing.T) {
		actualExtension1, err = suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext1.Name, ext1.DisplayName, ext1.Version)
		require.NoError(t, err)
		require.Equal(t, ext1.Name, actualExtension1.Name)
		require.Equal(t, ext1.DisplayName, actualExtension1.DisplayName)
		require.Equal(t, ext1.Version, actualExtension1.Version)

	})

	t.Run("fail - duplicate schema extension name", func(t *testing.T) {
		_, err := suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext1.Name, ext1.DisplayName, ext1.Version)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionName)

	})

	t.Run("success - create another graph schema extension", func(t *testing.T) {
		actualExtension2, err = suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext2.Name, ext2.DisplayName, ext2.Version)
		require.NoError(t, err)
		require.Equal(t, ext2.Name, actualExtension2.Name)
		require.Equal(t, ext2.DisplayName, actualExtension2.DisplayName)
		require.Equal(t, ext2.Version, actualExtension2.Version)

	})

	t.Run("success - get a graph schema extension by its ID", func(t *testing.T) {
		got, err := suite.BHDatabase.GetGraphSchemaExtensionById(testCtx, actualExtension1.ID)
		require.NoError(t, err)
		require.Equal(t, actualExtension1.Name, got.Name)
		require.Equal(t, actualExtension1.DisplayName, got.DisplayName)
		require.Equal(t, actualExtension1.Version, got.Version)
		require.Equal(t, false, got.IsBuiltin)
		require.Equal(t, false, got.CreatedAt.IsZero())
		require.Equal(t, false, got.UpdatedAt.IsZero())
		require.Equal(t, false, got.DeletedAt.Valid)
	})

	t.Run("fail - graph schema extension does not exist", func(t *testing.T) {
		_, err = suite.BHDatabase.GetGraphSchemaExtensionById(testCtx, 1234)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("success - update graph schema extension", func(t *testing.T) {
		updatedName := "updated name"
		actualExtension1.Name = updatedName
		updatedExtension, err := suite.BHDatabase.UpdateGraphSchemaExtension(testCtx, actualExtension1)
		require.NoError(t, err)
		require.Equal(t, updatedName, updatedExtension.Name)
	})

	t.Run("fail - graph schema extension not found", func(t *testing.T) {
		nonExistentExtension := actualExtension1
		nonExistentExtension.ID = 1234

		_, err = suite.BHDatabase.UpdateGraphSchemaExtension(testCtx, nonExistentExtension)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("fail - duplicate graph schema extension name", func(t *testing.T) {
		actualExtension1.Name = actualExtension2.Name
		_, err = suite.BHDatabase.UpdateGraphSchemaExtension(testCtx, actualExtension1)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionName)
	})

	t.Run("success - delete graph schema extension", func(t *testing.T) {
		err = suite.BHDatabase.DeleteGraphSchemaExtension(testCtx, actualExtension1.ID)
		require.NoError(t, err)
	})

	t.Run("fail - graph schema extension not found", func(t *testing.T) {
		err = suite.BHDatabase.DeleteGraphSchemaExtension(testCtx, 1234)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

func TestDatabase_GetGraphSchemaExtensions(t *testing.T) {
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
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, model.Filters{}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 4)
		require.Equal(t, 4, total)
	})

	t.Run("successfully returns an array of extensions, with filtering", func(t *testing.T) {
		var filter = make(model.Filters, 1)
		filter["name"] = []model.Filter{
			{
				Operator: model.Equals,
				Value:    "david",
			},
		}

		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, filter, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 1)
		require.Equal(t, 1, total)
	})

	t.Run("successfully returns an array of extensions, with multiple filters", func(t *testing.T) {
		var filter = make(model.Filters, 1)
		filter["name"] = []model.Filter{
			{
				Operator: model.Equals,
				Value:    "david",
			},
		}
		filter["display_name"] = []model.Filter{
			{
				Operator: model.Equals,
				Value:    "yet another extension",
			},
		}

		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, filter, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 1)
		require.Equal(t, 1, total)
	})

	t.Run("successfully returns an array of extensions, with fuzzy filtering", func(t *testing.T) {
		var filter = make(model.Filters, 1)
		filter["display_name"] = []model.Filter{
			{
				Operator: model.ApproximatelyEquals,
				Value:    "test extension",
			},
		}
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, filter, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, 2, total)
	})

	t.Run("successfully returns an array of extensions, with fuzzy filtering and sort ascending", func(t *testing.T) {
		var filter = make(model.Filters, 1)
		filter["display_name"] = []model.Filter{
			{
				Operator: model.ApproximatelyEquals,
				Value:    "test extension",
			},
		}
		extensions, _, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, filter, model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, "adam", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, with fuzzy filtering and sort descending", func(t *testing.T) {
		var filter = make(model.Filters, 1)
		filter["display_name"] = []model.Filter{
			{
				Operator: model.ApproximatelyEquals,
				Value:    "test extension",
			},
		}
		extensions, _, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, filter, model.Sort{{Column: "display_name", Direction: model.DescendingSortDirection}}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, "bob", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, no filtering or sorting, with skip", func(t *testing.T) {
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, model.Filters{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, 4, total)
		require.Equal(t, "charlie", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, no filtering or sorting, with limit", func(t *testing.T) {
		extensions, total, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, model.Filters{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, 4, total)
		require.Equal(t, "bob", extensions[1].Name)
	})

	t.Run("returns an error with bogus filtering", func(t *testing.T) {
		var filter = make(model.Filters, 1)
		filter["nonexistentcolumn"] = []model.Filter{
			{
				Operator: model.Equals,
				Value:    "david",
			},
		}
		_, _, err := suite.BHDatabase.GetGraphSchemaExtensions(testCtx, filter, model.Sort{}, 0, 0)
		require.Error(t, err)
		require.Equal(t, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)", err.Error())
	})
}

func TestDatabase_SchemaNodeKind_CRUD(t *testing.T) {
	t.Parallel()

	testSuite := setupIntegrationTestSuite(t)

	defer teardownIntegrationTestSuite(t, &testSuite)

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0")
	require.NoError(t, err)
	var (
		nodeKind1 = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_1",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_1",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		nodeKind2 = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_2",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_2",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		nodeKind3 = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_3",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_3",
			Description:       "Test Kind description 1", // used in fuzzy test
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		nodeKind4 = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_4",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_4",
			Description:       "Test Kind description 2", // used in fuzzy test
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}

		want = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_1",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_1",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		want2 = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_2",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_2",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		want3 = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_3",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_3",
			Description:       "Test Kind description 1", // used in fuzzy test
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		want4 = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_4",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_4",
			Description:       "Test Kind description 2", // used in fuzzy test
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		updateWant = model.GraphSchemaNodeKind{
			Name:              "Test_Kind_345",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_345",
			Description:       "a test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}

		gotNodeKind1 = model.GraphSchemaNodeKind{}
		gotNodeKind2 = model.GraphSchemaNodeKind{}
	)

	// CREATE

	// Expected success - create one model.GraphSchemaNodeKind
	t.Run("success - create a schema node kind 1", func(t *testing.T) {
		gotNodeKind1, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
		require.NoError(t, err)
		compareGraphSchemaNodeKind(t, gotNodeKind1, want)
	})
	// Expected success - create a second model.GraphSchemaNodeKind
	t.Run("success - create a schema node kind 2", func(t *testing.T) {
		gotNodeKind2, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
		require.NoError(t, err)
		compareGraphSchemaNodeKind(t, gotNodeKind2, want2)
	})
	// Expected fail - return error indicating non unique name
	t.Run("fail - create schema node kind does not have unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
		require.ErrorIs(t, err, database.ErrDuplicateSchemaNodeKindName)
	})

	// GET

	// Expected success - get the first model.GraphSchemaNodeKind
	t.Run("success - get schema node kind 1", func(t *testing.T) {
		gotNodeKind1, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, gotNodeKind1.ID)
		require.NoError(t, err)
		compareGraphSchemaNodeKind(t, gotNodeKind1, want)
	})
	// Expected fail - return an error if trying to return a node_kind that does not exist
	t.Run("fail - get a node kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, 1112412)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	// GET With pagination / filtering

	// setup
	_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
	require.NoError(t, err)
	_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind4.Name, nodeKind4.SchemaExtensionId, nodeKind4.DisplayName, nodeKind4.Description, nodeKind4.IsDisplayKind, nodeKind4.Icon, nodeKind4.IconColor)
	require.NoError(t, err)

	// Expected success - return all schema node kinds
	t.Run("success - return node schema kinds, no filter or sorting", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, nodeKinds, 4)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{want, want2, want3, want4})
	})
	// Expected success - return schema node kinds whose name is Test_Kind_3
	t.Run("success - return node schema kinds using a filter", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: "Test_Kind_3", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, nodeKinds, 1)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{want3})
	})

	// Expected success - return schema node kinds fuzzy filtering on description
	t.Run("success - return schema node kinds using a fuzzy filterer", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Test Kind ", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{want3, want4})
	})
	// Expected success - return schema node kinds fuzzy filtering on description and sort ascending on description
	t.Run("success - return schema node kinds using a fuzzy filterer and an ascending sort column", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Test Kind ", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{want3, want4})
	})
	// Expected success - return schema node kinds fuzzy filtering on description and sort descending on description
	t.Run("success - return schema node kinds using a fuzzy filterer and a descending sort column", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Test Kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.DescendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{want4, want3})
	})
	// Expected success - return schema node kinds, no filtering or sorting, with skip
	t.Run("success - return schema node kinds using skip, no filtering or sorting", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{want3, want4})
	})
	// Expected success - return schema node kinds, no filtering or sorting, with limit
	t.Run("success - return schema node kinds using limit, no filtering or and sorting", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{want, want2})
	})
	// Expected fail - return error for filtering on non-existent column
	t.Run("fail - return schema node kinds using a fuzzy filterer", func(t *testing.T) {
		_, _, err = testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"nonexistentcolumn": []model.Filter{{Operator: model.Equals, Value: "blah", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")
	})

	// UPDATE

	// Expected success - update schema node kind 1 to want 3
	t.Run("success - update schema node kind 1 to want 3", func(t *testing.T) {
		updateWant.ID = gotNodeKind1.ID
		gotUpdateNodeKind3, err := testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, updateWant)
		require.NoError(t, err)
		compareGraphSchemaNodeKind(t, gotUpdateNodeKind3, updateWant)
	})
	// Expected fail - return an error if update violates table constraints (updating the first kind to match the second)
	t.Run("fail - update schema node kind does not have unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, model.GraphSchemaNodeKind{Serial: model.Serial{ID: gotNodeKind1.ID}, Name: "Test_Kind_2", SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrDuplicateSchemaNodeKindName)
	})
	// Expected fail - return an error if trying to update a node_kind that does not exist
	t.Run("fail - update a node kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, model.GraphSchemaNodeKind{Serial: model.Serial{ID: 123123}, Name: "TEST_KIND_NOT_DUPLICATE", SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	// DELETE

	// Expected success - delete node kind 1
	t.Run("success - delete node kind 1", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaNodeKind(testSuite.Context, gotNodeKind1.ID)
		require.NoError(t, err)
	})
	// Expected fail - return an error if trying to delete a node_kind that does not exist
	t.Run("fail - delete a node kind that does not exist", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaNodeKind(testSuite.Context, gotNodeKind1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

func TestDatabase_SchemaProperties_CRUD(t *testing.T) {
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

	extension, err := suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext1.Name, ext1.DisplayName, ext1.Version)
	require.NoError(t, err)

	extension2, err := suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext2.Name, ext2.DisplayName, ext2.Version)
	require.NoError(t, err)

	var (
		extProp1 = model.GraphSchemaProperty{
			SchemaExtensionID: extension.ID,
			Name:              "ext_prop_1",
			DisplayName:       "Extension Property 1",
			DataType:          "string",
			Description:       "Extremely fun and exciting extension property",
		}
		extProp2 = model.GraphSchemaProperty{
			SchemaExtensionID: extension.ID,
			Name:              "ext_prop_2",
			DisplayName:       "Extension Property 2",
			DataType:          "integer",
			Description:       "Mediocre and average extension property",
		}
		extProp3 = model.GraphSchemaProperty{
			Name:        "ext_prop_3",
			DisplayName: "Extension Property 3",
			DataType:    "array",
			Description: "Extremely boring and lame extension property",
		}
	)

	extensionProperty1, err := suite.BHDatabase.CreateGraphSchemaProperty(testCtx, extProp1.SchemaExtensionID, extProp1.Name, extProp1.DisplayName, extProp1.DataType, extProp1.Description)
	require.NoError(t, err)
	require.Equal(t, extProp1.SchemaExtensionID, extensionProperty1.SchemaExtensionID)
	require.Equal(t, extProp1.Name, extensionProperty1.Name)
	require.Equal(t, extProp1.DisplayName, extensionProperty1.DisplayName)
	require.Equal(t, extProp1.DataType, extensionProperty1.DataType)
	require.Equal(t, extProp1.Description, extensionProperty1.Description)

	// Ensure name uniqueness across same extension
	_, err = suite.BHDatabase.CreateGraphSchemaProperty(testCtx, extProp1.SchemaExtensionID, extProp1.Name, extProp1.DisplayName, extProp1.DataType, extProp1.Description)
	require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionPropertyName)

	// Ensure name can be duplicate across different extensions
	_, err = suite.BHDatabase.CreateGraphSchemaProperty(testCtx, extension2.ID, extProp1.Name, extProp1.DisplayName, extProp1.DataType, extProp1.Description)
	require.NoError(t, err)

	extensionProperty2, err := suite.BHDatabase.CreateGraphSchemaProperty(testCtx, extProp2.SchemaExtensionID, extProp2.Name, extProp2.DisplayName, extProp2.DataType, extProp2.Description)
	require.NoError(t, err)
	require.Equal(t, extProp2.SchemaExtensionID, extensionProperty2.SchemaExtensionID)
	require.Equal(t, extProp2.Name, extensionProperty2.Name)
	require.Equal(t, extProp2.DisplayName, extensionProperty2.DisplayName)
	require.Equal(t, extProp2.DataType, extensionProperty2.DataType)
	require.Equal(t, extProp2.Description, extensionProperty2.Description)

	got, err := suite.BHDatabase.GetGraphSchemaPropertyById(testCtx, extensionProperty1.ID)
	require.NoError(t, err)
	require.Equal(t, extensionProperty1.Name, got.Name)
	require.Equal(t, extensionProperty1.DisplayName, got.DisplayName)
	require.Equal(t, extensionProperty1.DataType, got.DataType)
	require.Equal(t, extensionProperty1.Description, got.Description)
	require.Equal(t, false, got.CreatedAt.IsZero())
	require.Equal(t, false, got.UpdatedAt.IsZero())
	require.Equal(t, false, got.DeletedAt.Valid)

	_, err = suite.BHDatabase.GetGraphSchemaPropertyById(testCtx, 1234)
	require.ErrorIs(t, err, database.ErrNotFound)

	_, err = suite.BHDatabase.CreateGraphSchemaProperty(testCtx, extProp3.SchemaExtensionID, extProp3.Name, extProp3.DisplayName, extProp3.DataType, extProp3.Description)
	require.Error(t, err)

	updateProperty := extensionProperty2
	updateProperty.Name = "ext_prop_2"
	updateProperty.DisplayName = "Extension Property 2"
	updateProperty.DataType = "integer"
	updateProperty.Description = "Extremely boring and lame extension property"

	updatedExtensionProperty, err := suite.BHDatabase.UpdateGraphSchemaProperty(testCtx, updateProperty)
	require.NoError(t, err)
	require.Equal(t, updateProperty.ID, updatedExtensionProperty.ID)
	require.Equal(t, updateProperty.SchemaExtensionID, updatedExtensionProperty.SchemaExtensionID)
	require.Equal(t, updateProperty.Name, updatedExtensionProperty.Name)
	require.Equal(t, updateProperty.DisplayName, updatedExtensionProperty.DisplayName)
	require.Equal(t, updateProperty.DataType, updatedExtensionProperty.DataType)
	require.Equal(t, updateProperty.Description, updatedExtensionProperty.Description)
	require.NotEqual(t, updateProperty.UpdatedAt, updatedExtensionProperty.UpdatedAt)

	err = suite.BHDatabase.DeleteGraphSchemaProperty(testCtx, updatedExtensionProperty.ID)
	require.NoError(t, err)

	_, err = suite.BHDatabase.GetGraphSchemaPropertyById(testCtx, updatedExtensionProperty.ID)
	require.ErrorIs(t, err, database.ErrNotFound)
}

func TestDatabase_SchemaEdgeKind_CRUD(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)
	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_edge_kinds", "test_extension", "1.0.0")
	require.NoError(t, err)

	var (
		edgeKind1 = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_1",
			Description:       "test edge kind",
			IsTraversable:     false,
		}
		edgeKind2 = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_2",
			Description:       "test edge kind",
			IsTraversable:     true,
		}

		want1 = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_1",
			Description:       "test edge kind",
			IsTraversable:     false,
		}
		want2 = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_2",
			Description:       "test edge kind",
			IsTraversable:     true,
		}
		want3 = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_3",
			Description:       "test edge kind",
			IsTraversable:     false,
		}

		gotEdgeKind1 = model.GraphSchemaEdgeKind{}
		gotEdgeKind2 = model.GraphSchemaEdgeKind{}
	)

	// Expected success - create one model.GraphSchemaEdgeKind
	t.Run("success - create a schema edge kind #1", func(t *testing.T) {
		gotEdgeKind1, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind1, want1)
	})
	// Expected success - create a second model.GraphSchemaEdgeKind
	t.Run("success - create a schema edge kind #2", func(t *testing.T) {
		gotEdgeKind2, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind2, want2)
	})
	// Expected success - get first model.GraphSchemaEdgeKind
	t.Run("success - get a schema edge kind #1", func(t *testing.T) {
		gotEdgeKind1, err = testSuite.BHDatabase.GetGraphSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind1, want1)
	})
	// Expected fail - return error indicating non unique name
	t.Run("fail - create schema edge kind does not have a unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
		require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
	})
	// Expected success - update edgeKind1 to want3
	t.Run("success - update edgeKind1 to want3", func(t *testing.T) {
		want3.ID = gotEdgeKind1.ID
		gotEdgeKind3, err := testSuite.BHDatabase.UpdateGraphSchemaEdgeKind(testSuite.Context, want3)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind3, want3)
	})
	// Expected fail - return an error if update violates table constraints (update first edge kind to match the second)
	t.Run("fail - update schema edge kind does not have a unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaEdgeKind(testSuite.Context, model.GraphSchemaEdgeKind{Serial: model.Serial{ID: gotEdgeKind1.ID}, Name: edgeKind2.Name, SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
	})
	// Expected success - delete edge kind 1
	t.Run("success - delete edge kind 1", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaEdgeKind(testSuite.Context, gotEdgeKind1.ID)
		require.NoError(t, err)
	})
	// Expected fail - return error for if an edge kind that does not exist
	t.Run("fail - get an edge kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetGraphSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
	// Expected fail - return an error if trying to delete an edge_kind that does not exist (edgeKind1 was already deleted)
	t.Run("fail - delete an edge kind that does not exist", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaEdgeKind(testSuite.Context, gotEdgeKind1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
	// Expected fail - return an error if trying to update an edge_kind that does not exist
	t.Run("fail - update an edge kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaEdgeKind(testSuite.Context, model.GraphSchemaEdgeKind{Serial: model.Serial{ID: 1124123}, Name: edgeKind2.Name, SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

// compareGraphSchemaNodeKinds - compares the returned list of model.GraphSchemaNodeKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaNodeKinds(t *testing.T, got, want model.GraphSchemaNodeKinds) {
	for i, schemaNodeKind := range got {
		compareGraphSchemaNodeKind(t, schemaNodeKind, want[i])
	}
}

func compareGraphSchemaNodeKind(t *testing.T, got, want model.GraphSchemaNodeKind) {
	t.Helper()
	// We cant predictably know the want id prior to running parallel tests as other tests may already be using this table.
	require.Equalf(t, want.Name, got.Name, "CreateSchemaNodeKind(%v) - name mismatch", got.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "CreateSchemaNodeKind(%v) - extension_id mismatch", got.SchemaExtensionId)
	require.Equalf(t, want.DisplayName, got.DisplayName, "CreateSchemaNodeKind(%v) - display_name mismatch", got.DisplayName)
	require.Equalf(t, want.Description, got.Description, "CreateSchemaNodeKind(%v) - description mismatch", got.Description)
	require.Equalf(t, want.IsDisplayKind, got.IsDisplayKind, "CreateSchemaNodeKind(%v) - is_display_kind mismatch", got.IsDisplayKind)
	require.Equalf(t, want.Icon, got.Icon, "CreateSchemaNodeKind(%v) - icon mismatch", got.Icon)
	require.Equalf(t, want.IconColor, got.IconColor, "CreateSchemaNodeKind(%v) - icon_color mismatch", got.IconColor)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "CreateSchemaNodeKind(%v) - created_at is zero", got.CreatedAt.IsZero())
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "CreateSchemaNodeKind(%v) - updated_at is zero", got.UpdatedAt.IsZero())
	require.Equalf(t, false, got.DeletedAt.Valid, "CreateSchemaNodeKind(%v) - deleted_at is not null", got.DeletedAt.Valid)
}

func compareGraphSchemaEdgeKind(t *testing.T, got, want model.GraphSchemaEdgeKind) {
	t.Helper()
	// We cant predictably know the want id prior to running parallel tests as other tests may already be using this table.
	require.Equalf(t, want.Name, got.Name, "CreateSchemaEdgeKind - name - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.Description, got.Description, "CreateSchemaEdgeKind - description - got %v, want %v", got.Description, want.Description)
	require.Equalf(t, want.IsTraversable, got.IsTraversable, "CreateSchemaEdgeKind - IsTraversable - got %v, want %t", got.IsTraversable, want.IsTraversable)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "CreateSchemaEdgeKind - SchemaExtensionId - got %d, want %d", got.SchemaExtensionId, want.SchemaExtensionId)
}
