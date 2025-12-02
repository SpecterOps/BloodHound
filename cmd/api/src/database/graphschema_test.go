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

func TestBloodhoundDB_CreateAndGetExtensionSchemaNodeKind(t *testing.T) {
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
	)

	// Expected success - create one model.SchemaNodeKind
	gotNodeKind1, err := testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
	require.NoError(t, err)
	compareSchemaNodeKind(t, gotNodeKind1, want)
	// Expected success - create a second model.SchemaNodeKind
	gotNodeKind2, err := testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
	require.NoError(t, err)
	compareSchemaNodeKind(t, gotNodeKind2, want2)
	// Expected success - get the first model.SchemaNodeKind
	gotNodeKind1, err = testSuite.BHDatabase.GetSchemaNodeKindByID(testSuite.Context, gotNodeKind1.ID)
	require.NoError(t, err)
	compareSchemaNodeKind(t, gotNodeKind1, want)
	// Expected fail - return error for record that does not exist
	_, err = testSuite.BHDatabase.GetSchemaNodeKindByID(testSuite.Context, 21321)
	require.EqualError(t, err, "entity not found")
	// Expected fail - return error indicating non unique name
	_, err = testSuite.BHDatabase.CreateSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
	require.ErrorIs(t, err, database.ErrDuplicateSchemaNodeKindName)
}

func compareSchemaNodeKind(t *testing.T, got, want model.SchemaNodeKind) {
	t.Helper()
	// require.Equalf(t, want.ID, got.ID, "CreateSchemaNodeKind(%v) - id mismatch", got.ID) - We cant predictably know the want id beforehand in parallel tests as other tests may already be using this table.
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
	)

	extension, err := suite.BHDatabase.CreateGraphSchemaExtension(testCtx, ext1.Name, ext1.DisplayName, ext1.Version)
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

	_, err = suite.BHDatabase.CreateGraphSchemaProperty(testCtx, extProp1.SchemaExtensionID, extProp1.Name, extProp1.DisplayName, extProp1.DataType, extProp1.Description)
	require.Error(t, err)
	require.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionPropertyName)

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
	require.Error(t, err)
	require.Equal(t, "entity not found", err.Error())

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
	require.Equal(t, err, database.ErrNotFound)
}
