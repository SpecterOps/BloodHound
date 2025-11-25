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

func TestDatabase_CreateAndGetSchemaEdgeKinds(t *testing.T) {
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
	)

	// Expected success - create one model.SchemaEdgeKind
	gotEdgeKind1, err := testSuite.BHDatabase.CreateSchemaEdgeKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
	require.NoError(t, err)
	compareSchemaEdgeKind(t, gotEdgeKind1, want1)
	// Expected success - create a second model.SchemaEdgeKind
	gotEdgeKind2, err := testSuite.BHDatabase.CreateSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
	require.NoError(t, err)
	compareSchemaEdgeKind(t, gotEdgeKind2, want2)
	// Expected success - get first model.SchemaEdgeKind
	gotEdgeKind1, err = testSuite.BHDatabase.GetSchemaEdgeKindById(testSuite.Context, gotEdgeKind1.ID)
	require.NoError(t, err)
	compareSchemaEdgeKind(t, gotEdgeKind1, want1)
	// Expected fail - return error for record that does not exist
	_, err = testSuite.BHDatabase.GetSchemaEdgeKindById(testSuite.Context, 312412213)
	require.EqualError(t, err, "entity not found")
	// Expected fail - return error indicating non unique name
	_, err = testSuite.BHDatabase.CreateSchemaEdgeKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
	require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
}

func compareSchemaEdgeKind(t *testing.T, got, want model.SchemaEdgeKind) {
	t.Helper()
	require.Equalf(t, want.Name, got.Name, "CreateSchemaEdgeKind - name - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.Description, got.Description, "CreateSchemaEdgeKind - description - got %v, want %v", got.Description, want.Description)
	require.Equalf(t, want.IsTraversable, got.IsTraversable, "CreateSchemaEdgeKind - IsTraversable - got %v, want %t", got.IsTraversable, want.IsTraversable)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "CreateSchemaEdgeKind - SchemaExtensionId - got %d, want %d", got.SchemaExtensionId, want.SchemaExtensionId)
}
