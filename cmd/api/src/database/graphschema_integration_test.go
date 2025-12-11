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
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		ext1 = model.GraphSchemaExtension{
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
		actualExtension1, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version)
		require.NoError(t, err)
		require.Equal(t, ext1.Name, actualExtension1.Name)
		require.Equal(t, ext1.DisplayName, actualExtension1.DisplayName)
		require.Equal(t, ext1.Version, actualExtension1.Version)

	})

	t.Run("fail - duplicate schema extension name", func(t *testing.T) {
		_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionName)

	})

	t.Run("success - create another graph schema extension", func(t *testing.T) {
		actualExtension2, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext2.Name, ext2.DisplayName, ext2.Version)
		require.NoError(t, err)
		require.Equal(t, ext2.Name, actualExtension2.Name)
		require.Equal(t, ext2.DisplayName, actualExtension2.DisplayName)
		require.Equal(t, ext2.Version, actualExtension2.Version)

	})

	t.Run("success - get a graph schema extension by its ID", func(t *testing.T) {
		got, err := testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, actualExtension1.ID)
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
		_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, 1234)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("success - update graph schema extension", func(t *testing.T) {
		updatedName := "updated name"
		actualExtension1.Name = updatedName
		updatedExtension, err := testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, actualExtension1)
		require.NoError(t, err)
		require.Equal(t, updatedName, updatedExtension.Name)
	})

	t.Run("fail - graph schema extension not found", func(t *testing.T) {
		nonExistentExtension := actualExtension1
		nonExistentExtension.ID = 1234

		_, err = testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, nonExistentExtension)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("fail - duplicate graph schema extension name", func(t *testing.T) {
		actualExtension1.Name = actualExtension2.Name
		_, err = testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, actualExtension1)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionName)
	})

	t.Run("success - delete graph schema extension", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, actualExtension1.ID)
		require.NoError(t, err)
	})

	t.Run("fail - graph schema extension not found", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, 1234)
		require.Error(t, err)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

func TestDatabase_GetGraphSchemaExtensions(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		ext1 = model.GraphSchemaExtension{
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

	_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version)
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext2.Name, ext2.DisplayName, ext2.Version)
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext3.Name, ext3.DisplayName, ext3.Version)
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext4.Name, ext4.DisplayName, ext4.Version)
	require.NoError(t, err)

	t.Run("successfully returns an array of extensions, no filtering or sorting", func(t *testing.T) {
		extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
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

		extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, filter, model.Sort{}, 0, 0)
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

		extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, filter, model.Sort{}, 0, 0)
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
		extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, filter, model.Sort{}, 0, 0)
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
		extensions, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, filter, model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}}, 0, 0)
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
		extensions, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, filter, model.Sort{{Column: "display_name", Direction: model.DescendingSortDirection}}, 0, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, "bob", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, no filtering or sorting, with skip", func(t *testing.T) {
		extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, model.Filters{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Len(t, extensions, 2)
		require.Equal(t, 4, total)
		require.Equal(t, "charlie", extensions[0].Name)
	})

	t.Run("successfully returns an array of extensions, no filtering or sorting, with limit", func(t *testing.T) {
		extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
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
		_, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, filter, model.Sort{}, 0, 0)
		require.Error(t, err)
		require.Equal(t, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)", err.Error())
	})
}

func TestDatabase_GraphSchemaNodeKind_CRUD(t *testing.T) {
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
		compareGraphSchemaNodeKind(t, gotNodeKind1, nodeKind1)
	})
	// Expected success - create a second model.GraphSchemaNodeKind
	t.Run("success - create a schema node kind 2", func(t *testing.T) {
		gotNodeKind2, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
		require.NoError(t, err)
		compareGraphSchemaNodeKind(t, gotNodeKind2, nodeKind2)
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
		compareGraphSchemaNodeKind(t, gotNodeKind1, nodeKind1)
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
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{nodeKind1, nodeKind2, nodeKind3, nodeKind4})
	})
	// Expected success - return schema node kinds whose name is Test_Kind_3
	t.Run("success - return node schema kinds using a filter", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: "Test_Kind_3", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, nodeKinds, 1)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{nodeKind3})
	})

	// Expected success - return schema node kinds fuzzy filtering on description
	t.Run("success - return schema node kinds using a fuzzy filterer", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Test Kind ", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{nodeKind3, nodeKind4})
	})
	// Expected success - return schema node kinds fuzzy filtering on description and sort ascending on description
	t.Run("success - return schema node kinds using a fuzzy filterer and an ascending sort column", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Test Kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.AscendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{nodeKind3, nodeKind4})
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
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{nodeKind4, nodeKind3})
	})
	// Expected success - return schema node kinds, no filtering or sorting, with skip
	t.Run("success - return schema node kinds using skip, no filtering or sorting", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{nodeKind3, nodeKind4})
	})
	// Expected success - return schema node kinds, no filtering or sorting, with limit
	t.Run("success - return schema node kinds using limit, no filtering or sorting", func(t *testing.T) {
		nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, nodeKinds, 2)
		compareGraphSchemaNodeKinds(t, nodeKinds, model.GraphSchemaNodeKinds{nodeKind1, nodeKind2})
	})
	// Expected fail - return error for filtering on non-existent column
	t.Run("fail - return error for filtering on non-existent column", func(t *testing.T) {
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

func TestDatabase_GraphSchemaProperties_CRUD(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		ext1 = model.GraphSchemaExtension{
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

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version)
	require.NoError(t, err)

	extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext2.Name, ext2.DisplayName, ext2.Version)
	require.NoError(t, err)

	var (
		extension1Property1 = model.GraphSchemaProperty{
			SchemaExtensionId: extension.ID,
			Name:              "ext_prop_1",
			DisplayName:       "Extension Property 1",
			DataType:          "string",
			Description:       "Extremely fun and exciting extension property",
		}
		extension1Property2 = model.GraphSchemaProperty{
			SchemaExtensionId: extension.ID,
			Name:              "ext_prop_2",
			DisplayName:       "Extension Property 2",
			DataType:          "integer",
			Description:       "Mediocre and average extension property",
		}
		extension2Property1 = model.GraphSchemaProperty{
			SchemaExtensionId: extension2.ID,
			Name:              "ext_prop_1",
			DisplayName:       "Extension Property 1",
			DataType:          "string",
			Description:       "Extremely boring and lame extension property 1",
		}
		extension2Property2 = model.GraphSchemaProperty{
			SchemaExtensionId: extension2.ID,
			Name:              "ext_prop_2",
			DisplayName:       "Extension Property 2",
			DataType:          "array",
			Description:       "Extremely boring and lame extension property 2",
		}
		updateProperty1Want = model.GraphSchemaProperty{
			SchemaExtensionId: extension2.ID,
			Name:              "ext_prop_3",
			Description:       "Extremely boring and lame extension property",
			DataType:          "integer",
			DisplayName:       "Extension Property 3",
		}

		gotExtension1Property1 = model.GraphSchemaProperty{}
		gotExtension1Property2 = model.GraphSchemaProperty{}
		gotExtension2Property1 = model.GraphSchemaProperty{}
	)

	// CREATE

	// Expected success - Create extension1Property1
	t.Run("success - create schema extension1Property1", func(t *testing.T) {
		gotExtension1Property1, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
		require.NoError(t, err)
		compareGraphSchemaProperty(t, gotExtension1Property1, extension1Property1)
	})
	// Expected fail - Ensure name uniqueness across same extension
	t.Run("fail - create schema extension1Property2", func(t *testing.T) {
		_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
		require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionPropertyName)
	})
	// Expected success - Ensure name can be duplicate across different extensions
	t.Run("success - create schema extension2Property1", func(t *testing.T) {
		gotExtension2Property1, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property1.SchemaExtensionId, extension2Property1.Name, extension2Property1.DisplayName, extension2Property1.DataType, extension2Property1.Description)
		require.NoError(t, err)
		compareGraphSchemaProperty(t, gotExtension2Property1, extension2Property1)
	})
	// Expected success - create extension property 2
	t.Run("success - create schema extension1Property2", func(t *testing.T) {
		gotExtension1Property2, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
		require.NoError(t, err)
		compareGraphSchemaProperty(t, gotExtension1Property2, extension1Property2)
	})

	// GET

	// Expected success - get schema ext1prop1
	t.Run("success - get schema extension1Property1", func(t *testing.T) {
		gotExtension1Property1, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, gotExtension1Property1.ID)
		require.NoError(t, err)
		compareGraphSchemaProperty(t, gotExtension1Property1, extension1Property1)
	})
	// Expected fail - return error for non-existent property
	t.Run("fail - return error for non-existent", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, 1234)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	// GET with pagination and filtering

	// setup
	_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
	require.NoError(t, err)

	// Expected success - return all schema properties
	t.Run("success - return all schema properties, no filter or sorting", func(t *testing.T) {
		schemaProperties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, schemaProperties, 4)
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension1Property1, extension2Property1, extension1Property2, extension2Property2})
	})
	// Expected success - return schema properties whose data type is an array
	t.Run("success - return properties using a filter", func(t *testing.T) {
		schemaProperties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context,
			model.Filters{"data_type": []model.Filter{{Operator: model.Equals, Value: "array", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, schemaProperties, 1)
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension2Property2})
	})

	// Expected success - return schema properties fuzzy filtering on description
	t.Run("success - return schema properties using a fuzzy filterer", func(t *testing.T) {
		schemaProperties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Extremely boring and lame extension property ", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, schemaProperties, 2)
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension2Property1, extension2Property2})
	})
	// Expected success - return schema properties fuzzy filtering on description and sort ascending on description
	t.Run("success - return schema properties using a fuzzy filterer and an ascending sort column", func(t *testing.T) {
		schemaProperties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Extremely boring and lame extension property ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.AscendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, schemaProperties, 2)
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension2Property1, extension2Property2})
	})
	// Expected success - return schema properties fuzzy filtering on description and sort descending on description
	t.Run("success - return schema properties using a fuzzy filterer and a descending sort column", func(t *testing.T) {
		schemaProperties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "Extremely boring and lame extension property ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.DescendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, schemaProperties, 2)
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension2Property1, extension2Property2})
	})
	// Expected success - return schema properties, no filtering or sorting, with skip
	t.Run("success - return schema properties using skip, no filtering or sorting", func(t *testing.T) {
		schemaProperties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, model.Filters{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, schemaProperties, 2)
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension1Property2, extension2Property2})
	})
	// Expected success - return schema properties, no filtering or sorting, with limit
	t.Run("success - return schema properties using limit, no filtering or sorting", func(t *testing.T) {
		schemaProperties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, schemaProperties, 2)
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension1Property1, extension2Property1})
	})
	// Expected fail - return error for filtering on non-existent column
	t.Run("fail - return error for filtering on non-existent column", func(t *testing.T) {
		_, _, err = testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context,
			model.Filters{"nonexistentcolumn": []model.Filter{{Operator: model.Equals, Value: "blah", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")
	})

	// UPDATE

	// Expected success - update ext1prop1 to updateProperty1Want
	t.Run("success - update schema extension1Property1 to", func(t *testing.T) {
		updatedExtensionProperty, err := testSuite.BHDatabase.UpdateGraphSchemaProperty(testSuite.Context, model.GraphSchemaProperty{
			Serial: model.Serial{
				ID: gotExtension1Property1.ID,
			},
			SchemaExtensionId: extension2.ID,
			Name:              updateProperty1Want.Name,
			DisplayName:       updateProperty1Want.DisplayName,
			DataType:          updateProperty1Want.DataType,
			Description:       updateProperty1Want.Description,
		})
		require.NoError(t, err)
		compareGraphSchemaProperty(t, updatedExtensionProperty, updateProperty1Want)
		require.NotEqual(t, updateProperty1Want.UpdatedAt, updatedExtensionProperty.UpdatedAt)
	})
	// Expected fail - return error if update causes name collision
	t.Run("fail - return error if update causes name collision", func(t *testing.T) {
		_, err := testSuite.BHDatabase.UpdateGraphSchemaProperty(testSuite.Context, model.GraphSchemaProperty{
			Serial: model.Serial{
				ID: gotExtension1Property1.ID,
			},
			SchemaExtensionId: extension.ID,
			Name:              extension1Property2.Name,
		})
		require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionPropertyName)
	})

	// DELETE

	// Expected success - delete first property
	t.Run("success - delete schema extension1Property1", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaProperty(testSuite.Context, gotExtension1Property1.ID)
		require.NoError(t, err)
	})
	// Expected fail - retrieve first property that was just deleted
	t.Run("fail - delete schema extension1Property2", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, gotExtension1Property1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

}

func TestDatabase_GraphSchemaEdgeKind_CRUD(t *testing.T) {
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
		edgeKind3 = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_3",
			Description:       "test edge kind 3",
			IsTraversable:     false,
		}
		edgeKind4 = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_4",
			Description:       "test edge kind 4",
			IsTraversable:     false,
		}
		updateWant = model.GraphSchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_345",
			Description:       "test edge kind",
			IsTraversable:     false,
		}

		gotEdgeKind1 = model.GraphSchemaEdgeKind{}
		gotEdgeKind2 = model.GraphSchemaEdgeKind{}
	)

	// CREATE

	// Expected success - create one model.GraphSchemaEdgeKind
	t.Run("success - create a schema edge kind #1", func(t *testing.T) {
		gotEdgeKind1, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind1, edgeKind1)
	})
	// Expected success - create a second model.GraphSchemaEdgeKind
	t.Run("success - create a schema edge kind #2", func(t *testing.T) {
		gotEdgeKind2, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind2, edgeKind2)
	})
	// Expected fail - return error indicating non unique name
	t.Run("fail - create schema edge kind does not have a unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
		require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
	})

	// GET

	// Expected success - get first model.GraphSchemaEdgeKind
	t.Run("success - get a schema edge kind #1", func(t *testing.T) {
		gotEdgeKind1, err = testSuite.BHDatabase.GetGraphSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind1, edgeKind1)
	})
	// Expected fail - return error for if an edge kind that does not exist
	t.Run("fail - get an edge kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetGraphSchemaEdgeKindById(testSuite.Context, 23423235)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	// GET With pagination / filtering

	// setup
	_, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind3.Name, edgeKind3.SchemaExtensionId, edgeKind3.Description, edgeKind3.IsTraversable)
	require.NoError(t, err)
	_, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind4.Name, edgeKind4.SchemaExtensionId, edgeKind4.Description, edgeKind4.IsTraversable)
	require.NoError(t, err)

	// Expected success - return all schema edge kinds
	t.Run("success - return edge schema kinds, no filter or sorting", func(t *testing.T) {
		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, edgeKinds, 4)
		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaEdgeKinds{edgeKind1, edgeKind2, edgeKind3, edgeKind4})
	})
	// Expected success - return schema edge kinds whose name is Test_Kind_3
	t.Run("success - return edge schema kinds using a filter", func(t *testing.T) {
		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context,
			model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: "test_edge_kind_3", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, edgeKinds, 1)
		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaEdgeKinds{edgeKind3})
	})

	// Expected success - return schema edge kinds fuzzy filtering on description
	t.Run("success - return schema edge kinds using a fuzzy filterer", func(t *testing.T) {
		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test edge kind ", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, edgeKinds, 2)
		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaEdgeKinds{edgeKind3, edgeKind4})
	})
	// Expected success - return schema edge kinds fuzzy filtering on description and sort ascending on description
	t.Run("success - return schema edge kinds using a fuzzy filterer and an ascending sort column", func(t *testing.T) {
		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test edge kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.AscendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, edgeKinds, 2)
		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaEdgeKinds{edgeKind3, edgeKind4})
	})
	// Expected success - return schema edge kinds fuzzy filtering on description and sort descending on description
	t.Run("success - return schema edge kinds using a fuzzy filterer and a descending sort column", func(t *testing.T) {
		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test edge kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.DescendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, edgeKinds, 2)
		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaEdgeKinds{edgeKind4, edgeKind3})
	})
	// Expected success - return schema edge kinds, no filtering or sorting, with skip
	t.Run("success - return schema edge kinds using skip, no filtering or sorting", func(t *testing.T) {
		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, edgeKinds, 2)
		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaEdgeKinds{edgeKind3, edgeKind4})
	})
	// Expected success - return schema edge kinds, no filtering or sorting, with limit
	t.Run("success - return schema edge kinds using limit, no filtering or sorting", func(t *testing.T) {
		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, edgeKinds, 2)
		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaEdgeKinds{edgeKind1, edgeKind2})
	})
	// Expected fail - return error for filtering on non-existent column
	t.Run("fail - return error for filtering on non-existent column", func(t *testing.T) {
		_, _, err = testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context,
			model.Filters{"nonexistentcolumn": []model.Filter{{Operator: model.Equals, Value: "blah", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")
	})

	// UPDATE

	// Expected success - update edgeKind1 to updateWant
	t.Run("success - update edgeKind1 to updateWant", func(t *testing.T) {
		updateWant.ID = gotEdgeKind1.ID
		gotEdgeKind3, err := testSuite.BHDatabase.UpdateGraphSchemaEdgeKind(testSuite.Context, updateWant)
		require.NoError(t, err)
		compareGraphSchemaEdgeKind(t, gotEdgeKind3, updateWant)
	})
	// Expected fail - return an error if update violates table constraints (update first edge kind to match the second)
	t.Run("fail - update schema edge kind does not have a unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaEdgeKind(testSuite.Context, model.GraphSchemaEdgeKind{Serial: model.Serial{ID: gotEdgeKind1.ID}, Name: edgeKind2.Name, SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
	})
	// Expected fail - return an error if trying to update an edge_kind that does not exist
	t.Run("fail - update an edge kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaEdgeKind(testSuite.Context, model.GraphSchemaEdgeKind{Serial: model.Serial{ID: 1124123}, Name: edgeKind2.Name, SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	// DELETE

	// Expected success - delete edge kind 1
	t.Run("success - delete edge kind 1", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaEdgeKind(testSuite.Context, gotEdgeKind1.ID)
		require.NoError(t, err)
	})
	// Expected fail - return an error if trying to delete an edge_kind that does not exist
	t.Run("fail - delete an edge kind that does not exist", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaEdgeKind(testSuite.Context, 1231231)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

// compareGraphSchemaNodeKinds - compares the returned list of model.GraphSchemaNodeKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaNodeKinds(t *testing.T, got, want model.GraphSchemaNodeKinds) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaNodeKinds")
	for i, schemaNodeKind := range got {
		compareGraphSchemaNodeKind(t, schemaNodeKind, want[i])
	}
}

func compareGraphSchemaNodeKind(t *testing.T, got, want model.GraphSchemaNodeKind) {
	t.Helper()
	// We cant predictably know the want id prior to running parallel tests as other tests may already be using this table.
	require.GreaterOrEqualf(t, got.ID, int32(1), "GraphSchemaNodeKinds - ID is invalid")
	require.Equalf(t, want.Name, got.Name, "GraphSchemaNodeKind(%v) - name mismatch", got.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaNodeKind(%v) - extension_id mismatch", got.SchemaExtensionId)
	require.Equalf(t, want.DisplayName, got.DisplayName, "GraphSchemaNodeKind(%v) - display_name mismatch", got.DisplayName)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaNodeKind(%v) - description mismatch", got.Description)
	require.Equalf(t, want.IsDisplayKind, got.IsDisplayKind, "GraphSchemaNodeKind(%v) - is_display_kind mismatch", got.IsDisplayKind)
	require.Equalf(t, want.Icon, got.Icon, "GraphSchemaNodeKind(%v) - icon mismatch", got.Icon)
	require.Equalf(t, want.IconColor, got.IconColor, "GraphSchemaNodeKind(%v) - icon_color mismatch", got.IconColor)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "GraphSchemaNodeKind(%v) - created_at is zero", got.CreatedAt.IsZero())
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "GraphSchemaNodeKind(%v) - updated_at is zero", got.UpdatedAt.IsZero())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaNodeKind(%v) - deleted_at is not null", got.DeletedAt.Valid)
}

// compareGraphSchemaProperties - compares the returned list of model.GraphSchemaProperties with the expected results.
// // Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaProperties(t *testing.T, got, want model.GraphSchemaProperties) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaProperties")
	for i, schemaProperty := range got {
		compareGraphSchemaProperty(t, schemaProperty, want[i])
	}
}

func compareGraphSchemaProperty(t *testing.T, got, want model.GraphSchemaProperty) {
	t.Helper()
	// We cant predictably know the want id prior to running parallel tests as other tests may already be using this table.
	require.GreaterOrEqualf(t, got.ID, int32(1), "GraphSchemaProperty - ID is invalid")
	require.Equalf(t, want.Name, got.Name, "GraphSchemaProperty - name mismatch - got: %v, want: %v", got.Name, want.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaProperty - schema_extension_id mismatch - got: %v, want: %v", got.SchemaExtensionId, want.SchemaExtensionId)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaProperty - description mismatch - got: %v, want: %v", got.Description, want.Description)
	require.Equalf(t, want.DisplayName, got.DisplayName, "GraphSchemaProperty - display_name mismatch - got: %v, want: %v", got.DisplayName, want.DisplayName)
	require.Equalf(t, want.DataType, got.DataType, "GraphSchemaProperty - data_type mismatch - got: %v, want: %v", got.DataType, want.DataType)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "GraphSchemaProperty - created_at is zero")
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "GraphSchemaProperty - updated_at is zero")
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaProperty - deleted_at is null")

}

// compareGraphSchemaEdgeKinds - compares the returned list of model.GraphSchemaEdgeKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaEdgeKinds(t *testing.T, got, want model.GraphSchemaEdgeKinds) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaEdgeKinds")
	for i, schemaEdgeKind := range got {
		compareGraphSchemaEdgeKind(t, schemaEdgeKind, want[i])
	}
}

func compareGraphSchemaEdgeKind(t *testing.T, got, want model.GraphSchemaEdgeKind) {
	t.Helper()
	// We cant predictably know the want id prior to running parallel tests as other tests may already be using this table.
	require.GreaterOrEqualf(t, got.ID, int32(1), "GraphSchemaEdgeKind - ID is invalid")
	require.Equalf(t, want.Name, got.Name, "GraphSchemaEdgeKind - name mismatch - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaEdgeKind - description mismatch - got %v, want %v", got.Description, want.Description)
	require.Equalf(t, want.IsTraversable, got.IsTraversable, "GraphSchemaEdgeKind - IsTraversable mismatch - got %t, want %t", got.IsTraversable, want.IsTraversable)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaEdgeKind - SchemaExtensionId mismatch - got %d, want %d", got.SchemaExtensionId, want.SchemaExtensionId)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "GraphSchemaEdgeKind(%v) - created_at is zero", got.CreatedAt.IsZero())
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "GraphSchemaEdgeKind(%v) - updated_at is zero", got.UpdatedAt.IsZero())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaEdgeKind(%v) - deleted_at is not null", got.DeletedAt.Valid)
}
