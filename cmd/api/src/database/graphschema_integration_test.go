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
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
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
			Namespace:   "test_namespace_1",
		}
		ext2 = model.GraphSchemaExtension{
			Name:        "test_name2",
			DisplayName: "test extension name 2",
			Version:     "1.0.0",
			Namespace:   "test_namespace_2",
		}
		actualExtension1 = model.GraphSchemaExtension{}
		actualExtension2 = model.GraphSchemaExtension{}
		err              error
	)

	t.Run("success - create a graph schema extension", func(t *testing.T) {
		actualExtension1, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
		require.NoError(t, err)
		require.Equal(t, ext1.Name, actualExtension1.Name)
		require.Equal(t, ext1.DisplayName, actualExtension1.DisplayName)
		require.Equal(t, ext1.Version, actualExtension1.Version)

	})

	t.Run("fail - duplicate schema extension name", func(t *testing.T) {
		_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
		require.Error(t, err)
		require.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionName)

	})

	t.Run("success - create another graph schema extension", func(t *testing.T) {
		actualExtension2, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext2.Name, ext2.DisplayName, ext2.Version, ext2.Namespace)
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
		require.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionName)
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
			Namespace:   "test_namespace_1",
		}
		ext2 = model.GraphSchemaExtension{
			Name:        "bob",
			DisplayName: "test extension name 2",
			Version:     "2.0.0",
			Namespace:   "test_namespace_2",
		}
		ext3 = model.GraphSchemaExtension{
			Name:        "charlie",
			DisplayName: "another extension",
			Version:     "3.0.0",
			Namespace:   "test_namespace_3",
		}
		ext4 = model.GraphSchemaExtension{
			Name:        "david",
			DisplayName: "yet another extension",
			Version:     "4.0.0",
			Namespace:   "test_namespace_4",
		}
	)

	_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext2.Name, ext2.DisplayName, ext2.Version, ext2.Namespace)
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext3.Name, ext3.DisplayName, ext3.Version, ext3.Namespace)
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext4.Name, ext4.DisplayName, ext4.Version, ext4.Namespace)
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

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "test_namespace_1")
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
		require.ErrorIs(t, err, model.ErrDuplicateSchemaNodeKindName)
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
		_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, 112)
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

	// Expected success - update schema node kind 1 to want 3, the name should NOT be updated
	t.Run("success - update schema node kind 1 to want 3", func(t *testing.T) {
		updateWant.ID = gotNodeKind1.ID
		gotUpdateNodeKind3, err := testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, updateWant)
		require.NoError(t, err)
		compareGraphSchemaNodeKind(t, gotUpdateNodeKind3, model.GraphSchemaNodeKind{
			Serial: model.Serial{
				Basic: model.Basic{
					CreatedAt: updateWant.CreatedAt,
					UpdatedAt: updateWant.UpdatedAt,
				},
			},
			Name:              nodeKind1.Name,
			SchemaExtensionId: updateWant.SchemaExtensionId,
			DisplayName:       updateWant.DisplayName,
			Description:       updateWant.Description,
			IsDisplayKind:     updateWant.IsDisplayKind,
			Icon:              updateWant.Icon,
			IconColor:         updateWant.IconColor,
		})
	})
	// Expected fail - return an error if trying to update a node_kind that does not exist
	t.Run("fail - update a node kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, model.GraphSchemaNodeKind{Serial: model.Serial{ID: 1223}, Name: "TEST_KIND_NOT_DUPLICATE", SchemaExtensionId: extension.ID})
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
			Namespace:   "test_namespace_1",
		}
		ext2 = model.GraphSchemaExtension{
			Name:        "test_name2",
			DisplayName: "test extension name 2",
			Version:     "1.0.0",
			Namespace:   "test_namespace_2",
		}
	)

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
	require.NoError(t, err)

	extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext2.Name, ext2.DisplayName, ext2.Version, ext2.Namespace)
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
		require.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionPropertyName)
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
		compareGraphSchemaProperties(t, schemaProperties, model.GraphSchemaProperties{extension2Property2, extension2Property1})
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
		require.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionPropertyName)
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

func TestDatabase_GraphSchemaRelationshipKind_CRUD(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)
	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_relationship_kinds", "test_extension", "1.0.0", "test_namespace_1")
	require.NoError(t, err)

	var (
		relationshipKind1 = model.GraphSchemaRelationshipKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_relationship_kind_1",
			Description:       "test relationship kind",
			IsTraversable:     false,
		}
		relationshipKind2 = model.GraphSchemaRelationshipKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_relationship_kind_2",
			Description:       "test relationship kind",
			IsTraversable:     true,
		}
		relationshipKind3 = model.GraphSchemaRelationshipKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_relationship_kind_3",
			Description:       "test relationship kind 3",
			IsTraversable:     false,
		}
		relationshipKind4 = model.GraphSchemaRelationshipKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_relationship_kind_4",
			Description:       "test relationship kind 4",
			IsTraversable:     false,
		}
		updateWant = model.GraphSchemaRelationshipKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_relationship_kind_345",
			Description:       "test relationship kind",
			IsTraversable:     false,
		}

		gotRelationshipKind1 = model.GraphSchemaRelationshipKind{}
		gotRelationshipKind2 = model.GraphSchemaRelationshipKind{}
	)

	// CREATE

	// Expected success - create one model.GraphSchemaRelationshipKind
	t.Run("success - create a schema relationship kind #1", func(t *testing.T) {
		gotRelationshipKind1, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, relationshipKind1.Name, relationshipKind1.SchemaExtensionId, relationshipKind1.Description, relationshipKind1.IsTraversable)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKind(t, gotRelationshipKind1, relationshipKind1)
	})
	// Expected success - create a second model.GraphSchemaRelationshipKind
	t.Run("success - create a schema relationship kind #2", func(t *testing.T) {
		gotRelationshipKind2, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, relationshipKind2.Name, relationshipKind2.SchemaExtensionId, relationshipKind2.Description, relationshipKind2.IsTraversable)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKind(t, gotRelationshipKind2, relationshipKind2)
	})
	// Expected fail - return error indicating non unique name
	t.Run("fail - create schema relationship kind does not have a unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, relationshipKind2.Name, relationshipKind2.SchemaExtensionId, relationshipKind2.Description, relationshipKind2.IsTraversable)
		require.ErrorIs(t, err, model.ErrDuplicateSchemaRelationshipKindName)
	})

	// GET

	// Expected success - get first model.GraphSchemaRelationshipKind
	t.Run("success - get a schema relationship kind #1", func(t *testing.T) {
		gotRelationshipKind1, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, gotRelationshipKind1.ID)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKind(t, gotRelationshipKind1, relationshipKind1)
	})
	// Expected fail - return error for if an relationship kind that does not exist
	t.Run("fail - get an relationship kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, 235)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	// GET With pagination / filtering

	// setup
	_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, relationshipKind3.Name, relationshipKind3.SchemaExtensionId, relationshipKind3.Description, relationshipKind3.IsTraversable)
	require.NoError(t, err)
	_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, relationshipKind4.Name, relationshipKind4.SchemaExtensionId, relationshipKind4.Description, relationshipKind4.IsTraversable)
	require.NoError(t, err)

	// Expected success - return all schema relationship kinds
	t.Run("success - return relationship schema kinds, no filter or sorting", func(t *testing.T) {
		relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, relationshipKinds, 4)
		compareGraphSchemaRelationshipKinds(t, relationshipKinds, model.GraphSchemaRelationshipKinds{relationshipKind1, relationshipKind2, relationshipKind3, relationshipKind4})
	})
	// Expected success - return schema relationship kinds whose name is Test_Kind_3
	t.Run("success - return relationship schema kinds using a filter", func(t *testing.T) {
		relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
			model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: "test_relationship_kind_3", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, relationshipKinds, 1)
		compareGraphSchemaRelationshipKinds(t, relationshipKinds, model.GraphSchemaRelationshipKinds{relationshipKind3})
	})

	// Expected success - return schema relationship kinds fuzzy filtering on description
	t.Run("success - return schema relationship kinds using a fuzzy filterer", func(t *testing.T) {
		relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test relationship kind ", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, relationshipKinds, 2)
		compareGraphSchemaRelationshipKinds(t, relationshipKinds, model.GraphSchemaRelationshipKinds{relationshipKind3, relationshipKind4})
	})
	// Expected success - return schema relationship kinds fuzzy filtering on description and sort ascending on description
	t.Run("success - return schema relationship kinds using a fuzzy filterer and an ascending sort column", func(t *testing.T) {
		relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test relationship kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.AscendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, relationshipKinds, 2)
		compareGraphSchemaRelationshipKinds(t, relationshipKinds, model.GraphSchemaRelationshipKinds{relationshipKind3, relationshipKind4})
	})
	// Expected success - return schema relationship kinds fuzzy filtering on description and sort descending on description
	t.Run("success - return schema relationship kinds using a fuzzy filterer and a descending sort column", func(t *testing.T) {
		relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test relationship kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
				Direction: model.DescendingSortDirection,
				Column:    "description",
			}}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, relationshipKinds, 2)
		compareGraphSchemaRelationshipKinds(t, relationshipKinds, model.GraphSchemaRelationshipKinds{relationshipKind4, relationshipKind3})
	})
	// Expected success - return schema relationship kinds, no filtering or sorting, with skip
	t.Run("success - return schema relationship kinds using skip, no filtering or sorting", func(t *testing.T) {
		relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, model.Filters{}, model.Sort{}, 2, 0)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, relationshipKinds, 2)
		compareGraphSchemaRelationshipKinds(t, relationshipKinds, model.GraphSchemaRelationshipKinds{relationshipKind3, relationshipKind4})
	})
	// Expected success - return schema relationship kinds, no filtering or sorting, with limit
	t.Run("success - return schema relationship kinds using limit, no filtering or sorting", func(t *testing.T) {
		relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, relationshipKinds, 2)
		compareGraphSchemaRelationshipKinds(t, relationshipKinds, model.GraphSchemaRelationshipKinds{relationshipKind1, relationshipKind2})
	})
	// Expected fail - return error for filtering on non-existent column
	t.Run("fail - return error for filtering on non-existent column", func(t *testing.T) {
		_, _, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
			model.Filters{"nonexistentcolumn": []model.Filter{{Operator: model.Equals, Value: "blah", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")
	})

	// UPDATE

	// Expected success - update relationshipKind1 to updateWant, the name should NOT be updated
	t.Run("success - update relationshipKind1 to updateWant", func(t *testing.T) {
		updateWant.ID = gotRelationshipKind1.ID
		gotRelationshipKind3, err := testSuite.BHDatabase.UpdateGraphSchemaRelationshipKind(testSuite.Context, updateWant)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKind(t, gotRelationshipKind3, model.GraphSchemaRelationshipKind{
			Serial: model.Serial{
				Basic: model.Basic{
					CreatedAt: updateWant.CreatedAt,
					UpdatedAt: updateWant.UpdatedAt,
				},
			},
			SchemaExtensionId: updateWant.SchemaExtensionId,
			Name:              relationshipKind1.Name,
			Description:       updateWant.Description,
			IsTraversable:     updateWant.IsTraversable,
		})
	})
	// Expected fail - return an error if trying to update an relationship_kind that does not exist
	t.Run("fail - update an relationship kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateGraphSchemaRelationshipKind(testSuite.Context, model.GraphSchemaRelationshipKind{Serial: model.Serial{ID: 1123}, Name: relationshipKind2.Name, SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	// DELETE

	// Expected success - delete relationship kind 1
	t.Run("success - delete relationship kind 1", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaRelationshipKind(testSuite.Context, gotRelationshipKind1.ID)
		require.NoError(t, err)
	})
	// Expected fail - return an error if trying to delete an gotRelationship_kind that does not exist
	t.Run("fail - delete an relationship kind that does not exist", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteGraphSchemaRelationshipKind(testSuite.Context, 1231)
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

// compareGraphSchemaRelationshipKinds - compares the returned list of model.GraphSchemaRelationshipKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaRelationshipKinds(t *testing.T, got, want model.GraphSchemaRelationshipKinds) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaRelationshipKinds")
	for i, schemaRelationshipKind := range got {
		compareGraphSchemaRelationshipKind(t, schemaRelationshipKind, want[i])
	}
}

func compareGraphSchemaRelationshipKind(t *testing.T, got, want model.GraphSchemaRelationshipKind) {
	t.Helper()
	// We cant predictably know the want id prior to running parallel tests as other tests may already be using this table.
	require.GreaterOrEqualf(t, got.ID, int32(1), "GraphSchemaRelationshipKind - ID is invalid")
	require.Equalf(t, want.Name, got.Name, "GraphSchemaRelationshipKind - name mismatch - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaRelationshipKind - description mismatch - got %v, want %v", got.Description, want.Description)
	require.Equalf(t, want.IsTraversable, got.IsTraversable, "GraphSchemaRelationshipKind - IsTraversable mismatch - got %t, want %t", got.IsTraversable, want.IsTraversable)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaRelationshipKind - SchemaExtensionId mismatch - got %d, want %d", got.SchemaExtensionId, want.SchemaExtensionId)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "GraphSchemaRelationshipKind(%v) - created_at is zero", got.CreatedAt.IsZero())
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "GraphSchemaRelationshipKind(%v) - updated_at is zero", got.UpdatedAt.IsZero())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaRelationshipKind(%v) - deleted_at is not null", got.DeletedAt.Valid)
}

func TestDatabase_GraphSchemaRelationshipKindWithSchemaName_Get(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)
	extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "test_namespace_1")
	require.NoError(t, err)
	extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "test_namespace_2")
	require.NoError(t, err)

	relationshipKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_relationship_kind_1", extensionA.ID, "test relationship kind 1", false)
	require.NoError(t, err)

	relationshipKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_relationship_kind_2", extensionA.ID, "test relationship kind 2", true)
	require.NoError(t, err)

	relationshipKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_relationship_kind_3", extensionB.ID, "test relationship kind 3", false)
	require.NoError(t, err)

	relationshipKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_relationship_kind_4", extensionB.ID, "test relationship kind 4", true)
	require.NoError(t, err)
	var (
		want1 = model.GraphSchemaRelationshipKindWithNamedSchema{
			ID:            relationshipKind1.ID,
			SchemaName:    extensionA.Name,
			Name:          relationshipKind1.Name,
			Description:   relationshipKind1.Description,
			IsTraversable: relationshipKind1.IsTraversable,
		}
		want2 = model.GraphSchemaRelationshipKindWithNamedSchema{
			ID:            relationshipKind2.ID,
			SchemaName:    extensionA.Name,
			Name:          relationshipKind2.Name,
			Description:   relationshipKind2.Description,
			IsTraversable: relationshipKind2.IsTraversable,
		}
		want3 = model.GraphSchemaRelationshipKindWithNamedSchema{
			ID:            relationshipKind3.ID,
			SchemaName:    extensionB.Name,
			Name:          relationshipKind3.Name,
			Description:   relationshipKind3.Description,
			IsTraversable: relationshipKind3.IsTraversable,
		}
		want4 = model.GraphSchemaRelationshipKindWithNamedSchema{
			ID:            relationshipKind4.ID,
			SchemaName:    extensionB.Name,
			Name:          relationshipKind4.Name,
			Description:   relationshipKind4.Description,
			IsTraversable: relationshipKind4.IsTraversable,
		}
	)

	t.Run("success - get a schema relationship kind with named schema, no filters", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want1, want2, want3, want4}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema, filter for schema name", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.Equals, Value: extensionA.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want1, want2}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema, filter for multiple schema names", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.Equals, Value: extensionA.Name, SetOperator: model.FilterOr}, {Operator: model.Equals, Value: extensionB.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want1, want2, want3, want4}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema, filter for fuzzy match schema names", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test", SetOperator: model.FilterOr}, {Operator: model.Equals, Value: extensionB.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want1, want2, want3, want4}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema, filter for is_traversable", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"is_traversable": []model.Filter{{Operator: model.Equals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want2, want4}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema, filter for schema name and is_traversable", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.Equals, Value: extensionA.Name, SetOperator: model.FilterAnd}}, "is_traversable": []model.Filter{{Operator: model.Equals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want2}, actual)

	})

	t.Run("success - get a schema relationship kind with named schema, filter for not equals schema name", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.NotEquals, Value: extensionA.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want3, want4}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema, filter for is not traversable", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"is_traversable": []model.Filter{{Operator: model.NotEquals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want1, want3}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema, sort by relationship name descending", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{model.SortItem{Column: "name", Direction: model.DescendingSortDirection}}, 0, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want4, want3, want2, want1}, actual)

	})

	t.Run("success - get a schema relationship kind with named schema using skip, no filtering or sorting", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{}, 1, 0)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want2, want3, want4}, actual)
	})

	t.Run("success - get a schema relationship kind with named schema using limit, no filtering or sorting", func(t *testing.T) {
		actual, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
		require.NoError(t, err)
		compareGraphSchemaRelationshipKindsWithSchemaName(t, model.GraphSchemaRelationshipKindsWithNamedSchema{want1, want2}, actual)
	})

	t.Run("fail - error building sql filter", func(t *testing.T) {
		_, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"is_traversable": []model.Filter{{Operator: "invalid", Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 2)
		require.EqualError(t, err, "invalid operator specified")
	})

	t.Run("fail - error building sql sort", func(t *testing.T) {
		_, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{model.SortItem{Column: "name", Direction: model.InvalidSortDirection}}, 0, 2)
		require.ErrorIs(t, err, database.ErrInvalidSortDirection)
	})

	t.Run("fail - attempt to filter non-existent column", func(t *testing.T) {
		_, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, model.Filters{"invalid": []model.Filter{{Operator: model.Equals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 2)
		require.EqualError(t, err, "ERROR: column \"invalid\" does not exist (SQLSTATE 42703)")
	})
}

// compareGraphSchemaRelationshipKindsWithSchemaName - compares the returned model.GraphSchemaRelationshipKindsWithNamedSchema with the expected results.
func compareGraphSchemaRelationshipKindsWithSchemaName(t *testing.T, expected, actual model.GraphSchemaRelationshipKindsWithNamedSchema) {
	t.Helper()
	require.Equalf(t, len(expected), len(actual), "length mismatch of GraphSchemaRelationshipKindsWithNamedSchema")
	for i, schemaRelationshipKind := range actual {
		compareGraphSchemaRelationshipKindWithNamedSchema(t, expected[i], schemaRelationshipKind)
	}
}

func compareGraphSchemaRelationshipKindWithNamedSchema(t *testing.T, expected, actual model.GraphSchemaRelationshipKindWithNamedSchema) {
	t.Helper()
	require.Equalf(t, expected.Name, actual.Name, "GraphSchemaRelationshipKindWithNamedSchema - name - got %v, want %v", actual.Name, expected.Name)
	require.Equalf(t, expected.Description, actual.Description, "GraphSchemaRelationshipKindWithNamedSchema - description - got %v, want %v", actual.Description, expected.Description)
	require.Equalf(t, expected.IsTraversable, actual.IsTraversable, "GraphSchemaRelationshipKindWithNamedSchema - IsTraversable - got %v, want %t", actual.IsTraversable, expected.IsTraversable)
	require.Equalf(t, expected.SchemaName, actual.SchemaName, "GraphSchemaRelationshipKindWithNamedSchema - SchemaName - got %v, want %v", actual.SchemaName, expected.SchemaName)
}

func TestCreateSchemaEnvironment(t *testing.T) {
	type args struct {
		extensionId, environmentKindId, sourceKindId int32
	}
	type want struct {
		res model.SchemaEnvironment
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: schema environment created",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				// Create Schema Extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "Extension1", "DisplayName", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				extensionId:       1,
				environmentKindId: 1,
				sourceKindId:      1,
			},
			want: want{
				res: model.SchemaEnvironment{
					Serial: model.Serial{
						ID: 1,
					},
					SchemaExtensionId: 1,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, testCase.args.extensionId, testCase.args.environmentKindId, testCase.args.sourceKindId)
			if testCase.want.err != nil {
				assert.EqualError(t, err, testCase.want.err.Error())
			} else {
				// Zero out date fields before comparison
				got.CreatedAt = time.Time{}
				got.UpdatedAt = time.Time{}
				got.DeletedAt = sql.NullTime{}

				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSchemaEnvironments(t *testing.T) {
	var (
		defaultSchemaExtensionID = int32(1)
	)
	type want struct {
		res []model.SchemaEnvironment
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: no schema environments",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			want: want{
				res: []model.SchemaEnvironment{},
			},
		},
		{
			name: "Success: single schema environment",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				// Create Schema Extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "Extension1", "DisplayName", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)
				// Create Environments
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, defaultSchemaExtensionID, int32(1), int32(1))
				require.NoError(t, err)

				return testSuite
			},
			want: want{
				res: []model.SchemaEnvironment{
					{
						Serial: model.Serial{
							ID: 1,
						},
						SchemaExtensionId:          1,
						SchemaExtensionDisplayName: "DisplayName",
						EnvironmentKindId:          1,
						EnvironmentKindName:        "Tag_Tier_Zero",
						SourceKindId:               1,
					},
				},
			},
		},
		{
			name: "Success: multiple schema environments",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				// Create Schema Extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "Extension1", "DisplayName", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)
				// Create Environments
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, defaultSchemaExtensionID, int32(1), int32(1))
				require.NoError(t, err)
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, defaultSchemaExtensionID, int32(2), int32(2))
				require.NoError(t, err)

				return testSuite
			},
			want: want{
				res: []model.SchemaEnvironment{
					{
						Serial: model.Serial{
							ID: 1,
						},
						SchemaExtensionId:          1,
						SchemaExtensionDisplayName: "DisplayName",
						EnvironmentKindId:          1,
						EnvironmentKindName:        "Tag_Tier_Zero",
						SourceKindId:               1,
					},
					{
						Serial: model.Serial{
							ID: 2,
							Basic: model.Basic{
								CreatedAt: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
								DeletedAt: sql.NullTime{Time: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: false},
							},
						},
						SchemaExtensionId:          1,
						SchemaExtensionDisplayName: "DisplayName",
						EnvironmentKindId:          2,
						EnvironmentKindName:        "Tag_Owned",
						SourceKindId:               2,
					},
				},
			},
		},
		{
			name: "Success: schema environment with empty display name",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				// Create Schema Extension with empty display name
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "Extension1", "", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)
				// Create Environment
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, defaultSchemaExtensionID, int32(1), int32(1))
				require.NoError(t, err)

				return testSuite
			},
			want: want{
				res: []model.SchemaEnvironment{
					{
						Serial: model.Serial{
							ID: 1,
						},
						SchemaExtensionId:          1,
						SchemaExtensionDisplayName: "",
						EnvironmentKindId:          1,
						EnvironmentKindName:        "Tag_Tier_Zero",
						SourceKindId:               1,
					},
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.GetEnvironments(testSuite.Context)
			if testCase.want.err != nil {
				assert.EqualError(t, err, testCase.want.err.Error())
			} else {
				// Zero out date fields before comparison
				for i := range got {
					got[i].CreatedAt = time.Time{}
					got[i].UpdatedAt = time.Time{}
					got[i].DeletedAt = sql.NullTime{}
				}
				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabase_GetSchemaEnvironmentByGraphSchemaExtensionId(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		environmentNodeKind1 = model.GraphSchemaNodeKind{
			Name:        "GES_Environment 1",
			DisplayName: "Environment 1",
			Description: "an environment kind",
		}
		principalNodeKind1 = model.GraphSchemaNodeKind{
			Name:        "GES_Principal 1",
			DisplayName: "Principal 1",
			Description: "a principal kind",
		}
		environmentNodeKind2 = model.GraphSchemaNodeKind{
			Name:        "GES_Environment_2",
			DisplayName: "Environment 2",
			Description: "an environment kind",
		}
		principalNodeKind2 = model.GraphSchemaNodeKind{
			Name:        "GES_Principal_2",
			DisplayName: "Principal 2",
			Description: "a principal kind",
		}
	)
	type want struct {
		err error
	}
	tests := []struct {
		name     string
		setup    func(t *testing.T) (int32, []model.SchemaEnvironment)
		teardown func(t *testing.T, extensionId int32)
		want     want
	}{
		{
			name: "success - no environments found",
			setup: func(t *testing.T) (int32, []model.SchemaEnvironment) {
				t.Helper()
				return 123, []model.SchemaEnvironment{}
			},
			teardown: func(t *testing.T, extensionId int32) {},
			want: want{
				err: nil,
			},
		},
		{
			name: "success - get schema environment by id",
			setup: func(t *testing.T) (int32, []model.SchemaEnvironment) {
				t.Helper()
				var (
					err                         error
					createdExtension            model.GraphSchemaExtension
					createdEnvironmentDAWGSKind model.Kind
					createdEnvironment          model.SchemaEnvironment
					createdSchemaEnvironments   = make([]model.SchemaEnvironment, 0)
				)

				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					"GetEnvironmentsByExtensionIdTest", "DisplayName", "v1.0.0", "GES")
				require.NoError(t, err)

				// Create environmentNodeKind so its added to the DAWGS kind table
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdEnvironmentDAWGSKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, environmentNodeKind1.Name)
				require.NoError(t, err)

				// Register source kind
				kindType := graph.StringKind(environmentNodeKind1.Name)
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(kindType)
				require.NoError(t, err)

				sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, environmentNodeKind1.Name)
				require.NoError(t, err)

				createdEnvironment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					createdEnvironmentDAWGSKind.ID, int32(sourceKind.ID))
				require.NoError(t, err)
				createdSchemaEnvironments = append(createdSchemaEnvironments, createdEnvironment)

				return createdExtension.ID, createdSchemaEnvironments
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				var err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)
			},
		},
		{
			name: "success - get multiple schema environment by id",
			setup: func(t *testing.T) (int32, []model.SchemaEnvironment) {
				t.Helper()
				var (
					err                         error
					createdExtension            model.GraphSchemaExtension
					createdEnvironmentDAWGSKind model.Kind
					createdEnvironment          model.SchemaEnvironment
					createdSchemaEnvironments   = make([]model.SchemaEnvironment, 0)
				)

				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					"GetEnvironmentsByExtensionIdTest", "DisplayName", "v1.0.0", "GES")
				require.NoError(t, err)

				// Create environmentNodeKind so its added to the DAWGS kind table
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdEnvironmentDAWGSKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, environmentNodeKind1.Name)
				require.NoError(t, err)

				// Register source kind
				sourceKindType := graph.StringKind(principalNodeKind1.Name)
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(sourceKindType)
				require.NoError(t, err)

				sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, sourceKindType.String())
				require.NoError(t, err)

				createdEnvironment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					createdEnvironmentDAWGSKind.ID, int32(sourceKind.ID))
				require.NoError(t, err)
				createdSchemaEnvironments = append(createdSchemaEnvironments, createdEnvironment)

				// repeat for another environment
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context,
					environmentNodeKind2.Name, createdExtension.ID, environmentNodeKind2.DisplayName, environmentNodeKind2.Description,
					environmentNodeKind2.IsDisplayKind, environmentNodeKind2.Icon, environmentNodeKind2.IconColor)
				require.NoError(t, err)
				createdEnvironmentDAWGSKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, environmentNodeKind2.Name)
				require.NoError(t, err)

				// Register source kind
				sourceKindType = graph.StringKind(principalNodeKind2.Name)
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(sourceKindType)
				require.NoError(t, err)

				sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, sourceKindType.String())
				require.NoError(t, err)

				createdEnvironment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					createdEnvironmentDAWGSKind.ID, int32(sourceKind.ID))
				require.NoError(t, err)
				createdSchemaEnvironments = append(createdSchemaEnvironments, createdEnvironment)

				return createdExtension.ID, createdSchemaEnvironments
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				var err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdExtensionId, createdEnvironments := tt.setup(t)

			if got, err := testSuite.BHDatabase.GetEnvironmentsByExtensionId(testSuite.Context,
				createdExtensionId); tt.want.err != nil {
				require.EqualError(t, err, tt.want.err.Error())
			} else {
				require.NoError(t, err)
				compareSchemaEnvironments(t, got, createdEnvironments)
				tt.teardown(t, createdExtensionId)
			}
		})
	}
}

func compareSchemaEnvironments(t *testing.T, got, want []model.SchemaEnvironment) {
	t.Helper()
	require.Equalf(t, len(got), len(want), "number of schema environments are not equal")
	if len(want) == 0 {
		require.Equalf(t, want, got, "mismatched findings")
		return
	}
	for i := range got {
		compareSchemaEnvironment(t, got[i], want[i])
	}
}

func compareSchemaEnvironment(t *testing.T, got, want model.SchemaEnvironment) {
	t.Helper()

	require.Greaterf(t, got.ID, int32(0), "SchemaEnvironment - id should not be 0: %v", got.ID)
	require.Equalf(t, got.SchemaExtensionId, want.SchemaExtensionId, "SchemaEnvironment - schema_extension_id mismatch - got: %v, want: %v", got.EnvironmentKindId, want.EnvironmentKindId)
	require.Equalf(t, got.EnvironmentKindId, want.EnvironmentKindId, "SchemaEnvironment - environment_kind mismatch - got: %v, want: %v", got.EnvironmentKindId, want.EnvironmentKindId)
	require.Equalf(t, got.SourceKindId, want.SourceKindId, "SchemaEnvironment - source_kind mismatch - got: %v, want: %v", got.SourceKindId, want.SourceKindId)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "SchemaEnvironment - created_at is zero: %v", got.CreatedAt.IsZero())
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "SchemaEnvironment - updated_at is zero: %v", got.UpdatedAt.IsZero())
	require.Equalf(t, false, got.DeletedAt.Valid, "SchemaEnvironment - deleted_at is not null: %v", got.DeletedAt.Valid)
}

func TestGetSchemaEnvironmentById(t *testing.T) {
	var (
		defaultSchemaExtensionID = int32(1)
	)
	type args struct {
		environmentId int32
	}
	type want struct {
		res model.SchemaEnvironment
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: get schema environment by id",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				// Create Schema Extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "Extension1", "DisplayName", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)
				// Create Environment
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, defaultSchemaExtensionID, int32(1), int32(1))
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				environmentId: 1,
			},
			want: want{
				res: model.SchemaEnvironment{
					Serial: model.Serial{
						ID: 1,
					},
					SchemaExtensionId: 1,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				},
			},
		},
		{
			name: "Fail: schema environment not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				environmentId: 9999,
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, testCase.args.environmentId)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				// Zero out date fields before comparison
				got.CreatedAt = time.Time{}
				got.UpdatedAt = time.Time{}
				got.DeletedAt = sql.NullTime{}

				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteSchemaEnvironment(t *testing.T) {
	var (
		defaultSchemaExtensionID = int32(1)
	)
	type args struct {
		environmentId int32
	}
	type want struct {
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: delete schema environment",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				// Create Schema Extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "Extension1", "DisplayName", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)
				// Create Environment
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, defaultSchemaExtensionID, int32(1), int32(1))
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				environmentId: 1,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "Fail: schema environment not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				environmentId: 9999,
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			err := testSuite.BHDatabase.DeleteEnvironment(testSuite.Context, testCase.args.environmentId)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.NoError(t, err)

				// Verify deletion by trying to get the environment
				_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, testCase.args.environmentId)
				assert.ErrorIs(t, err, database.ErrNotFound)
			}
		})
	}
}

func TestTransaction_SchemaEnvironment(t *testing.T) {
	t.Run("Success: multiple schema environments commit together", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Create extension first (outside transaction)
		_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "TransactionTestExt", "Transaction Test", "v1.0.0", "test_namespace_1")
		require.NoError(t, err)

		// Create two environments in a single transaction
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			_, err := tx.CreateEnvironment(testSuite.Context, 1, 1, 1)
			if err != nil {
				return err
			}
			_, err = tx.CreateEnvironment(testSuite.Context, 1, 2, 2)
			return err
		})
		require.NoError(t, err)

		// Verify both environments were created
		env1, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, 1)
		require.NoError(t, err)
		assert.Equal(t, int32(1), env1.EnvironmentKindId)

		env2, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, 2)
		require.NoError(t, err)
		assert.Equal(t, int32(2), env2.EnvironmentKindId)
	})

	t.Run("Rollback: error causes all schema environment operations to rollback", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Create extension first (outside transaction)
		_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "TransactionRollbackExt", "Transaction Rollback Test", "v1.0.0", "test_namespace_2")
		require.NoError(t, err)

		// Create one environment, then fail - should rollback
		expectedErr := fmt.Errorf("intentional error to trigger rollback")
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			_, err := tx.CreateEnvironment(testSuite.Context, 1, 1, 1)
			if err != nil {
				return err
			}
			return expectedErr
		})
		require.ErrorIs(t, err, expectedErr)

		// Verify the environment was NOT created (rolled back)
		_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, 1)
		assert.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("Rollback: database constraint error causes rollback", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Create extension first (outside transaction)
		_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "TransactionDbErrorExt", "Transaction DB Error Test", "v1.0.0", "test_namespace_3")
		require.NoError(t, err)

		// Create one environment, then try to create a duplicate - should rollback both
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			_, err := tx.CreateEnvironment(testSuite.Context, 1, 1, 1)
			if err != nil {
				return err
			}
			// Try to create duplicate (same environment_kind_id + source_kind_id) - will fail
			_, err = tx.CreateEnvironment(testSuite.Context, 1, 1, 1)
			return err
		})
		require.Error(t, err)

		// Verify the first environment was NOT created (rolled back due to second failure)
		_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, 1)
		assert.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("Success: create and delete in same transaction", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		// Create extension first (outside transaction)
		_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "TransactionCreateDeleteExt", "Transaction Create Delete Test", "v1.0.0", "test_namespace_4")
		require.NoError(t, err)

		// Create and delete in same transaction
		err = testSuite.BHDatabase.Transaction(testSuite.Context, func(tx *database.BloodhoundDB) error {
			env, err := tx.CreateEnvironment(testSuite.Context, 1, 1, 1)
			if err != nil {
				return err
			}
			return tx.DeleteEnvironment(testSuite.Context, env.ID)
		})
		require.NoError(t, err)

		// Verify the environment does not exist
		_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, 1)
		assert.ErrorIs(t, err, database.ErrNotFound)
	})
}

func TestCreateSchemaRelationshipFinding(t *testing.T) {
	type args struct {
		extensionId        int32
		relationshipKindId int32
		environmentId      int32
		name               string
		displayName        string
	}
	type want struct {
		res model.SchemaRelationshipFinding
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: schema relationship finding created",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "FindingExtension", "Finding Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				extensionId:        1,
				relationshipKindId: 1,
				environmentId:      1,
				name:               "T0TestFinding",
				displayName:        "Test Finding Display Name",
			},
			want: want{
				res: model.SchemaRelationshipFinding{
					ID:                 1,
					SchemaExtensionId:  1,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "T0TestFinding",
					DisplayName:        "Test Finding Display Name",
				},
			},
		},
		{
			name: "Fail: duplicate finding name",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "FindingExtension2", "Finding Extension 2", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "DuplicateName", "Display Name")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				extensionId:        1,
				relationshipKindId: 1,
				environmentId:      1,
				name:               "DuplicateName",
				displayName:        "Another Display Name",
			},
			want: want{
				err: model.ErrDuplicateSchemaRelationshipFindingName,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(
				testSuite.Context,
				testCase.args.extensionId,
				testCase.args.relationshipKindId,
				testCase.args.environmentId,
				testCase.args.name,
				testCase.args.displayName,
			)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				got.CreatedAt = time.Time{}
				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSchemaRelationshipFindingById(t *testing.T) {
	type args struct {
		findingId int32
	}
	type want struct {
		res model.SchemaRelationshipFinding
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: get schema relationship finding by id",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "GetFindingExt", "Get Finding Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "GetByIdFinding", "Get By ID Finding")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				findingId: 1,
			},
			want: want{
				res: model.SchemaRelationshipFinding{
					ID:                 1,
					SchemaExtensionId:  1,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "GetByIdFinding",
					DisplayName:        "Get By ID Finding",
				},
			},
		},
		{
			name: "Fail: schema relationship finding not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				findingId: 9999,
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, testCase.args.findingId)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				got.CreatedAt = time.Time{}
				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabase_GetSchemaRelationshipsBySchemaExtensionId(t *testing.T) {
	var (
		testSuite = setupIntegrationTestSuite(t)
		extension = model.GraphSchemaExtension{
			Name:        "TestGraphSchemaExtension",
			DisplayName: "Test Graph Schema Extension",
			Version:     "1.0.0",
			Namespace:   "TGSE",
		}
		environmentNodeKind1 = model.GraphSchemaNodeKind{
			Name:        "TGSE_Environment 1",
			DisplayName: "Environment 1",
			Description: "an environment kind",
		}
		environmentNodeKind2 = model.GraphSchemaNodeKind{
			Name:        "TGSE_Environment_2",
			DisplayName: "Environment 2",
			Description: "an environment kind",
		}
		sourceKind1 = model.GraphSchemaNodeKind{
			Name:        "Source_Kind_1",
			DisplayName: "Source Kind 1",
			Description: "a source kind",
		}
		edgeKind1 = model.GraphSchemaRelationshipKind{
			Name:          "TGSE_Edge_1",
			Description:   "an edge kind",
			IsTraversable: true,
		}
		edgeKind2 = model.GraphSchemaRelationshipKind{
			Name:          "TGSE_Edge_2",
			Description:   "an edge kind",
			IsTraversable: true,
		}
		finding1 = model.SchemaRelationshipFinding{
			Name:        "Finding_1",
			DisplayName: "Finding 1",
		}
		finding2 = model.SchemaRelationshipFinding{
			Name:        "Finding_2",
			DisplayName: "Finding 2",
		}
	)
	defer teardownIntegrationTestSuite(t, &testSuite)

	tests := []struct {
		name     string
		setup    func(t *testing.T) (int32, []model.SchemaRelationshipFinding)
		teardown func(t *testing.T, extensionId int32)
		wantErr  error
	}{
		{
			name: "success - no findings for extension",
			setup: func(t *testing.T) (int32, []model.SchemaRelationshipFinding) {
				return 0, []model.SchemaRelationshipFinding{}
			},
			teardown: func(t *testing.T, extensionId int32) {},
			wantErr:  nil,
		},
		{
			name: "success - retrieve one finding",
			setup: func(t *testing.T) (int32, []model.SchemaRelationshipFinding) {
				t.Helper()
				var (
					createdExtension       model.GraphSchemaExtension
					err                    error
					createdFindings        []model.SchemaRelationshipFinding
					envKind, edgeKind      model.Kind
					createdEnvironmentNode model.GraphSchemaNodeKind
					createdSourceKindNode  model.GraphSchemaNodeKind
					sourceKind             database.SourceKind
					createdEnvironment     model.SchemaEnvironment
					createdEdgeKind        model.GraphSchemaRelationshipKind
					createdFinding         model.SchemaRelationshipFinding
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				// create env and source node kinds
				createdEnvironmentNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdSourceKindNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err)

				// retrieve DAWGS env kind
				envKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode.Name)
				require.NoError(t, err)

				// register and retrieve source kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(sourceKind1.Name))
				require.NoError(t, err)
				sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err)

				// create environment
				createdEnvironment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					envKind.ID, int32(sourceKind.ID))
				require.NoError(t, err)

				// create and retrieve finding edge kind
				createdEdgeKind, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)
				edgeKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind.Name)
				require.NoError(t, err)

				// create findings
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, edgeKind.ID, createdEnvironment.ID, finding1.Name, finding1.DisplayName)
				require.NoError(t, err)

				createdFindings = append(createdFindings, createdFinding)
				return createdExtension.ID, createdFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
			wantErr: nil,
		},
		{
			name: "success - retrieve multiple findings",
			setup: func(t *testing.T) (int32, []model.SchemaRelationshipFinding) {
				t.Helper()
				var (
					createdExtension                                 model.GraphSchemaExtension
					err                                              error
					createdFindings                                  []model.SchemaRelationshipFinding
					dawgsEnvKind1, dawgsEdgeKind1                    model.Kind
					dawgsEnvKind2, dawgsEdgeKind2                    model.Kind
					createdEnvironmentNode1, createdEnvironmentNode2 model.GraphSchemaNodeKind
					createdSourceKindNode                            model.GraphSchemaNodeKind
					sourceKind                                       database.SourceKind
					createdEnvironment1, createdEnvironment2         model.SchemaEnvironment
					createdEdgeKind                                  model.GraphSchemaRelationshipKind
					createdFinding                                   model.SchemaRelationshipFinding
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				// create env and source node kinds
				createdEnvironmentNode1, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdEnvironmentNode2, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind2.Name,
					createdExtension.ID, environmentNodeKind2.DisplayName, environmentNodeKind2.Description,
					environmentNodeKind2.IsDisplayKind, environmentNodeKind2.Icon, environmentNodeKind2.IconColor)
				require.NoError(t, err)
				createdSourceKindNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err)

				// retrieve DAWGS env kind
				dawgsEnvKind1, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode1.Name)
				require.NoError(t, err)
				dawgsEnvKind2, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode2.Name)
				require.NoError(t, err)

				// register and retrieve source kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(sourceKind1.Name))
				require.NoError(t, err)
				sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err)

				// create environment
				createdEnvironment1, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind1.ID, int32(sourceKind.ID))
				require.NoError(t, err)
				createdEnvironment2, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind2.ID, int32(sourceKind.ID))
				require.NoError(t, err)

				// create and retrieve finding edge kind
				createdEdgeKind, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind1, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind.Name)
				require.NoError(t, err)
				createdEdgeKind, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name,
					createdExtension.ID, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind2, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind.Name)
				require.NoError(t, err)

				// create and append findings
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind1.ID, createdEnvironment1.ID, finding1.Name, finding1.DisplayName)
				require.NoError(t, err)

				createdFindings = append(createdFindings, createdFinding)

				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind2.ID, createdEnvironment2.ID, finding2.Name, finding2.DisplayName)
				require.NoError(t, err)

				createdFindings = append(createdFindings, createdFinding)

				return createdExtension.ID, createdFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdExtensionId, createdSchemaRelationshipFindings := tt.setup(t)
			if got, err := testSuite.BHDatabase.GetSchemaRelationshipFindingsBySchemaExtensionId(testSuite.Context,
				createdExtensionId); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				compareSchemaRelationshipFindings(t, got, createdSchemaRelationshipFindings)
				tt.teardown(t, createdExtensionId)
			}
		})
	}
}

func compareSchemaRelationshipFindings(t *testing.T, got, want []model.SchemaRelationshipFinding) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "mismatched length of findings")
	if len(want) == 0 {
		require.Equalf(t, want, got, "mismatched findings")
		return
	}
	for i := range got {
		compareSchemaRelationshipFinding(t, got[i], want[i])
	}
}

func compareSchemaRelationshipFinding(t *testing.T, got, want model.SchemaRelationshipFinding) {
	t.Helper()
	require.Greater(t, got.ID, int32(0))
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "SchemaRelationshipFinding - schema_extension_id mismatch")
	require.Equalf(t, want.EnvironmentId, got.EnvironmentId, "SchemaRelationshipFinding - environment_id mismatch")
	require.Equalf(t, want.RelationshipKindId, got.RelationshipKindId, "SchemaRelationshipFinding - relationship_kind_id mismatch")
	require.Equalf(t, want.Name, got.Name, "SchemaRelationshipFinding - name mismatch")
	require.Equalf(t, want.DisplayName, got.DisplayName, "SchemaRelationshipFinding - display name mismatch")
	require.Equalf(t, false, got.CreatedAt.IsZero(), "SchemaRelationshipFinding - created_at is zero")
}

func TestDeleteSchemaRelationshipFinding(t *testing.T) {
	type args struct {
		findingId int32
	}
	type want struct {
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: delete schema relationship finding",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "DeleteFindingExt", "Delete Finding Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "DeleteFinding", "Delete Finding")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				findingId: 1,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "Fail: schema relationship finding not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				findingId: 9999,
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			err := testSuite.BHDatabase.DeleteSchemaRelationshipFinding(testSuite.Context, testCase.args.findingId)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.NoError(t, err)
				_, err = testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, testCase.args.findingId)
				assert.ErrorIs(t, err, database.ErrNotFound)
			}
		})
	}
}

func TestCreateRemediation(t *testing.T) {
	type args struct {
		findingId        int32
		shortDescription string
		longDescription  string
		shortRemediation string
		longRemediation  string
	}
	type want struct {
		res model.Remediation
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: remediations created",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "RemediationExt", "Remediation Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "RemediationFinding", "Remediation Finding")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				findingId:        1,
				shortDescription: "Short desc",
				longDescription:  "Long desc",
				shortRemediation: "Short fix",
				longRemediation:  "Long fix",
			},
			want: want{
				res: model.Remediation{
					FindingID:        1,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.CreateRemediation(
				testSuite.Context,
				testCase.args.findingId,
				testCase.args.shortDescription,
				testCase.args.longDescription,
				testCase.args.shortRemediation,
				testCase.args.longRemediation,
			)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRemediationByFindingId(t *testing.T) {
	type args struct {
		findingId int32
	}
	type want struct {
		res model.Remediation
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: get remediations by finding id",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "GetRemediationExt", "Get Remediation Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "GetRemediationFinding", "Get Remediation Finding")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, 1, "Short", "Long", "Short fix", "Long fix")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				findingId: 1,
			},
			want: want{
				res: model.Remediation{
					FindingID:        1,
					ShortDescription: "Short",
					LongDescription:  "Long",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				},
			},
		},
		{
			name: "Fail: remediations not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				findingId: 9999,
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, testCase.args.findingId)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateRemediation(t *testing.T) {
	type args struct {
		findingId        int32
		shortDescription string
		longDescription  string
		shortRemediation string
		longRemediation  string
	}
	type want struct {
		res model.Remediation
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: update existing remediations",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "UpdateRemediationExt", "Update Remediation Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "UpdateRemediationFinding", "Update Remediation Finding")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, 1, "Original short", "Original long", "Original short fix", "Original long fix")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				findingId:        1,
				shortDescription: "Updated short",
				longDescription:  "Updated long",
				shortRemediation: "Updated short fix",
				longRemediation:  "Updated long fix",
			},
			want: want{
				res: model.Remediation{
					FindingID:        1,
					ShortDescription: "Updated short",
					LongDescription:  "Updated long",
					ShortRemediation: "Updated short fix",
					LongRemediation:  "Updated long fix",
				},
			},
		},
		{
			name: "Success: upsert creates new remediations",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "UpsertRemediationExt", "Upsert Remediation Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "UpsertRemediationFinding", "Upsert Remediation Finding")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				findingId:        1,
				shortDescription: "New short",
				longDescription:  "New long",
				shortRemediation: "New short fix",
				longRemediation:  "New long fix",
			},
			want: want{
				res: model.Remediation{
					FindingID:        1,
					ShortDescription: "New short",
					LongDescription:  "New long",
					ShortRemediation: "New short fix",
					LongRemediation:  "New long fix",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			got, err := testSuite.BHDatabase.UpdateRemediation(
				testSuite.Context,
				testCase.args.findingId,
				testCase.args.shortDescription,
				testCase.args.longDescription,
				testCase.args.shortRemediation,
				testCase.args.longRemediation,
			)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.Equal(t, testCase.want.res, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteRemediation(t *testing.T) {
	type args struct {
		findingId int32
	}
	type want struct {
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: delete remediations",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "DeleteRemediationExt", "Delete Remediation Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, 1, 1, 1, "DeleteRemediationFinding", "Delete Remediation Finding")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, 1, "Short", "Long", "Short fix", "Long fix")
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				findingId: 1,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "Fail: remediations not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				findingId: 9999,
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			err := testSuite.BHDatabase.DeleteRemediation(testSuite.Context, testCase.args.findingId)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.NoError(t, err)
				_, err = testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, testCase.args.findingId)
				assert.ErrorIs(t, err, database.ErrNotFound)
			}
		})
	}
}

func TestCreateSchemaEnvironmentPrincipalKind(t *testing.T) {
	type args struct {
		environmentId int32
		principalKind int32
	}
	type want struct {
		res model.SchemaEnvironmentPrincipalKind
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Error: duplicate principal kind",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "EnvPrincipalKindExt", "Env Principal Kind Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, 1, 1)
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				environmentId: 1,
				principalKind: 1,
			},
			want: want{
				res: model.SchemaEnvironmentPrincipalKind{},
				err: model.ErrDuplicatePrincipalKind,
			},
		},
		{
			name: "Success: schema environment principal kind created",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "EnvPrincipalKindExt", "Env Principal Kind Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				environmentId: 1,
				principalKind: 1,
			},
			want: want{
				res: model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: 1,
					PrincipalKind: 1,
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			result, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, testCase.args.environmentId, testCase.args.principalKind)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want.res.EnvironmentId, result.EnvironmentId)
				assert.Equal(t, testCase.want.res.PrincipalKind, result.PrincipalKind)
			}
		})
	}
}

func TestGetSchemaEnvironmentPrincipalKindsByEnvironmentId(t *testing.T) {
	type args struct {
		environmentId int32
	}
	type want struct {
		count int
		err   error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: get schema environment principal kinds by environment id",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "GetEnvPrincipalKindExt", "Get Env Principal Kind Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, 1, 2)
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				environmentId: 1,
			},
			want: want{
				count: 2,
			},
		},
		{
			name: "Success: returns empty slice when no principal kinds exist",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				environmentId: 9999,
			},
			want: want{
				count: 0,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			result, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, testCase.args.environmentId)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, testCase.want.count)
			}
		})
	}
}

func TestDeleteSchemaEnvironmentPrincipalKind(t *testing.T) {
	type args struct {
		environmentId int32
		principalKind int32
	}
	type want struct {
		err error
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		args  args
		want  want
	}{
		{
			name: "Success: delete schema environment principal kind",
			setup: func() IntegrationTestSuite {
				t.Helper()
				testSuite := setupIntegrationTestSuite(t)

				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "DeleteEnvPrincipalKindExt", "Delete Env Principal Kind Extension", "v1.0.0", "test_namespace_1")
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, 1, 1, 1)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, 1, 1)
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				environmentId: 1,
				principalKind: 1,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "Fail: schema environment principal kind not found",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				environmentId: 9999,
				principalKind: 9999,
			},
			want: want{
				err: database.ErrNotFound,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			err := testSuite.BHDatabase.DeletePrincipalKind(testSuite.Context, testCase.args.environmentId, testCase.args.principalKind)
			if testCase.want.err != nil {
				assert.ErrorIs(t, err, testCase.want.err)
			} else {
				assert.NoError(t, err)
				result, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, testCase.args.environmentId)
				assert.NoError(t, err)
				assert.Len(t, result, 0)
			}
		})
	}
}

func TestDeleteSchemaExtension_CascadeDeletesAllDependents(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "CascadeTestExtension", "Cascade Test Extension", "v1.0.0", "test_namespace_1")
	require.NoError(t, err)

	nodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, "CascadeTestNodeKind", extension.ID, "Cascade Test Node Kind", "Test description", false, "fa-test", "#000000")
	require.NoError(t, err)

	property, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension.ID, "cascade_test_property", "Cascade Test Property", "string", "Test description")
	require.NoError(t, err)

	relationshipKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "CascadeTestrelationshipKind", extension.ID, "Test description", true)
	require.NoError(t, err)

	environment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, extension.ID, nodeKind.ID, nodeKind.ID)
	require.NoError(t, err)

	relationshipFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, extension.ID, relationshipKind.ID, environment.ID, "CascadeTestFinding", "Cascade Test Finding")
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, relationshipFinding.ID, "Short desc", "Long desc", "Short remediation", "Long remediation")
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, environment.ID, nodeKind.ID)
	require.NoError(t, err)

	err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extension.ID)
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, nodeKind.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	_, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, property.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	_, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, relationshipKind.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, environment.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	_, err = testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, relationshipFinding.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	_, err = testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, relationshipFinding.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	principalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, environment.ID)
	assert.NoError(t, err)
	assert.Len(t, principalKinds, 0)
}
