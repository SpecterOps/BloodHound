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

func TestBloodhoundDB_SchemaNodeKind_CRUD(t *testing.T) {
	t.Parallel()

	testSuite := setupIntegrationTestSuite(t)

	defer teardownIntegrationTestSuite(t, &testSuite)

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0")
	require.NoError(t, err)
	var (
		nodeKind1 = model.SchemaNodeKind{
			Name:              "Test_Kind_1",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_1",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		nodeKind2 = model.SchemaNodeKind{
			Name:              "Test_Kind_2",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_2",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}

		want = model.SchemaNodeKind{
			Name:              "Test_Kind_1",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_1",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		want2 = model.SchemaNodeKind{
			Name:              "Test_Kind_2",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_2",
			Description:       "A test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}
		want3 = model.SchemaNodeKind{
			Name:              "Test_Kind_3",
			SchemaExtensionId: extension.ID,
			DisplayName:       "Test_Kind_3",
			Description:       "a test kind",
			IsDisplayKind:     false,
			Icon:              "test_icon",
			IconColor:         "blue",
		}

		gotNodeKind1 = model.SchemaNodeKind{}
		gotNodeKind2 = model.SchemaNodeKind{}
	)

	// Expected success - create one model.SchemaNodeKind
	t.Run("success - create a schema node kind 1", func(t *testing.T) {
		gotNodeKind1, err = testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
		require.NoError(t, err)
		compareSchemaNodeKind(t, gotNodeKind1, want)
	})
	// Expected success - create a second model.SchemaNodeKind
	t.Run("success - create a schema node kind 2", func(t *testing.T) {
		gotNodeKind2, err = testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
		require.NoError(t, err)
		compareSchemaNodeKind(t, gotNodeKind2, want2)
	})
	// Expected success - get the first model.SchemaNodeKind
	t.Run("success - get schema node kind 1", func(t *testing.T) {
		gotNodeKind1, err = testSuite.BHDatabase.GetSchemaNodeKindByID(testSuite.Context, gotNodeKind1.ID)
		require.NoError(t, err)
		compareSchemaNodeKind(t, gotNodeKind1, want)
	})
	// Expected fail - return error indicating non unique name
	t.Run("fail - create schema node kind does not have unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
		require.ErrorIs(t, err, database.ErrDuplicateSchemaNodeKindName)
	})
	// Expected success - Update node kind 1 to want 3
	t.Run("success - update schema node kind 1 to want 3", func(t *testing.T) {
		want3.ID = gotNodeKind1.ID
		gotUpdateNodeKind3, err := testSuite.BHDatabase.UpdateSchemaNodeKindById(testSuite.Context, want3)
		require.NoError(t, err)
		compareSchemaNodeKind(t, gotUpdateNodeKind3, want3)
	})
	// Expected fail - return an error if update violates table constraints (updating the first kind to match the second)
	t.Run("fail - update schema node kind does not have unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateSchemaNodeKindById(testSuite.Context, model.SchemaNodeKind{Serial: model.Serial{ID: gotNodeKind1.ID}, Name: "Test_Kind_2", SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrDuplicateSchemaNodeKindName)
	})
	// Expected success - delete node kind 1
	t.Run("success - delete node kind 1", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteSchemaNodeKindById(testSuite.Context, gotNodeKind1.ID)
		require.NoError(t, err)
	})
	// Expected fail - return an error if trying to return a node_kind that does not exist
	t.Run("fail - get a node kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetSchemaNodeKindByID(testSuite.Context, gotNodeKind1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
	// Expected fail - return an error if trying to delete a node_kind that does not exist
	t.Run("fail - delete a node kind that does not exist", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteSchemaNodeKindById(testSuite.Context, gotNodeKind1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
	// Expected fail - return an error if trying to update a node_kind that does not exist
	t.Run("fail - update a node kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateSchemaNodeKindById(testSuite.Context, model.SchemaNodeKind{Serial: model.Serial{ID: 123123}, Name: "TEST_KIND_NOT_DUPLICATE", SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

func compareSchemaNodeKind(t *testing.T, got, want model.SchemaNodeKind) {
	t.Helper()
	// We cant predictably know the want id prior to running parallel tests as other tests may already be using this table.
	require.Equalf(t, want.Name, got.Name, "CreateSchemaNodeKind(%v) - name mismatch", got.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "CreateSchemaNodeKind(%v) - extension_id mismatch", got.SchemaExtensionId)
	require.Equalf(t, want.DisplayName, got.DisplayName, "CreateSchemaNodeKind(%v) - display_name mismatch", got.DisplayName)
	require.Equalf(t, want.Description, got.Description, "CreateSchemaNodeKind(%v) - description mismatch", got.Description)
	require.Equalf(t, want.IsDisplayKind, got.IsDisplayKind, "CreateSchemaNodeKind(%v) - is_display_kind mismatch", got.IsDisplayKind)
	require.Equalf(t, want.Icon, got.Icon, "CreateSchemaNodeKind(%v) - icon mismatch", got.Icon)
	require.Equalf(t, want.IconColor, got.IconColor, "CreateSchemaNodeKind(%v) - icon_color mismatch", got.IconColor)
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
		edgeKind1 = model.SchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_1",
			Description:       "test edge kind",
			IsTraversable:     false,
		}
		edgeKind2 = model.SchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_2",
			Description:       "test edge kind",
			IsTraversable:     true,
		}

		want1 = model.SchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_1",
			Description:       "test edge kind",
			IsTraversable:     false,
		}
		want2 = model.SchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_2",
			Description:       "test edge kind",
			IsTraversable:     true,
		}
		want3 = model.SchemaEdgeKind{
			Serial:            model.Serial{},
			SchemaExtensionId: extension.ID,
			Name:              "test_edge_kind_3",
			Description:       "test edge kind",
			IsTraversable:     false,
		}

		gotEdgeKind1 = model.SchemaEdgeKind{}
		gotEdgeKind2 = model.SchemaEdgeKind{}
	)

	// Expected success - create one model.SchemaEdgeKind
	t.Run("success - create a schema edge kind #1", func(t *testing.T) {
		gotEdgeKind1, err = testSuite.BHDatabase.CreateSchemaEdgeKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
		require.NoError(t, err)
		compareSchemaEdgeKind(t, gotEdgeKind1, want1)
	})
	// Expected success - create a second model.SchemaEdgeKind
	t.Run("success - create a schema edge kind #2", func(t *testing.T) {
		gotEdgeKind2, err = testSuite.BHDatabase.CreateSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
		require.NoError(t, err)
		compareSchemaEdgeKind(t, gotEdgeKind2, want2)
	})
	// Expected success - get first model.SchemaEdgeKind
	t.Run("success - get a schema edge kind #1", func(t *testing.T) {
		gotEdgeKind1, err = testSuite.BHDatabase.GetSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
		require.NoError(t, err)
		compareSchemaEdgeKind(t, gotEdgeKind1, want1)
	})
	// Expected fail - return error indicating non unique name
	t.Run("fail - create schema edge kind does not have a unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.CreateSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
		require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
	})
	// Expected success - update edgeKind1 to want3
	t.Run("success - update edgeKind1 to want3", func(t *testing.T) {
		want3.ID = gotEdgeKind1.ID
		gotEdgeKind3, err := testSuite.BHDatabase.UpdateSchemaEdgeKindByID(testSuite.Context, want3)
		require.NoError(t, err)
		compareSchemaEdgeKind(t, gotEdgeKind3, want3)
	})
	// Expected fail - return an error if update violates table constraints (update first edge kind to match the second)
	t.Run("fail - update schema edge kind does not have a unique name", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateSchemaEdgeKindByID(testSuite.Context, model.SchemaEdgeKind{Serial: model.Serial{ID: gotEdgeKind1.ID}, Name: edgeKind2.Name, SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
	})
	// Expected success - delete edge kind 1
	t.Run("success - delete edge kind 1", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
		require.NoError(t, err)
	})
	// Expected fail - return error for if an edge kind that does not exist
	t.Run("fail - get an edge kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.GetSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
	// Expected fail - return an error if trying to delete an edge_kind that does not exist (edgeKind1 was already deleted)
	t.Run("fail - delete an edge kind that does not exist", func(t *testing.T) {
		err = testSuite.BHDatabase.DeleteSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
	// Expected fail - return an error if trying to update an edge_kind that does not exist
	t.Run("fail - update an edge kind that does not exist", func(t *testing.T) {
		_, err = testSuite.BHDatabase.UpdateSchemaEdgeKindByID(testSuite.Context, model.SchemaEdgeKind{Serial: model.Serial{ID: 1124123}, Name: edgeKind2.Name, SchemaExtensionId: extension.ID})
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

func compareSchemaEdgeKind(t *testing.T, got, want model.SchemaEdgeKind) {
	t.Helper()
	require.Equalf(t, want.Name, got.Name, "CreateSchemaEdgeKind - name - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.Description, got.Description, "CreateSchemaEdgeKind - description - got %v, want %v", got.Description, want.Description)
	require.Equalf(t, want.IsTraversable, got.IsTraversable, "CreateSchemaEdgeKind - IsTraversable - got %v, want %t", got.IsTraversable, want.IsTraversable)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "CreateSchemaEdgeKind - SchemaExtensionId - got %d, want %d", got.SchemaExtensionId, want.SchemaExtensionId)
}
