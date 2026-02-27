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
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestExtension(t *testing.T, testSuite IntegrationTestSuite, name, displayName, version, namespace string) model.GraphSchemaExtension {
	t.Helper()
	// Create Extension with input arguments
	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, name, displayName, version, namespace)
	require.NoError(t, err, "unexpected error occurred when creating extension")
	return extension
}

func createTestNodeKind(t *testing.T, testSuite IntegrationTestSuite, name string, extensionID int32, displayName, description string, isDisplayKind bool, icon, iconColor string) model.GraphSchemaNodeKind {
	t.Helper()
	// Create Node Kind with input arguments
	nodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, name, extensionID, displayName, description, isDisplayKind, icon, iconColor)
	require.NoError(t, err, "unexpected error occurred when creating node kind")
	return nodeKind
}

func registerAndGetSourceKind(t *testing.T, testSuite IntegrationTestSuite, name string) database.SourceKind {
	t.Helper()
	// Register source kind with input arguments
	err := testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(name))
	require.NoError(t, err, "unexpected error occurred when registering source kind")
	// Retrieve registered source kind
	sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, name)
	require.NoError(t, err, "unexpected error occurred when retrieving source kind")
	return sourceKind
}

func getKindByName(t *testing.T, testSuite IntegrationTestSuite, name string) model.Kind {
	t.Helper()
	// Create Kind by name from DAWGS Kind Table
	kind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, name)
	require.NoError(t, err, "unexpected error occurred when getting kind by name")
	return kind
}

func createTestEnvironment(t *testing.T, testSuite IntegrationTestSuite, extensionID int32, envKindID int32, sourceKindID int32) model.SchemaEnvironment {
	t.Helper()
	// Create New Environment with input arguments
	env, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, extensionID, envKindID, sourceKindID)
	require.NoError(t, err, "unexpected error occurred when creating environment")
	return env
}

func createTestRelationshipKind(t *testing.T, testSuite IntegrationTestSuite, name string, extensionID int32, description string, isTraversable bool) model.GraphSchemaRelationshipKind {
	t.Helper()
	// Create New Relationship Kind with input arguments
	edgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, name, extensionID, description, isTraversable)
	require.NoError(t, err, "unexpected error occurred when creating relationship kind")
	return edgeKind
}

func createTestFinding(t *testing.T, testSuite IntegrationTestSuite, finding model.SchemaFinding) model.SchemaFinding {
	t.Helper()
	// Create New Finding with input arguments
	finding, err := testSuite.BHDatabase.CreateSchemaFinding(testSuite.Context, finding.Type, finding.SchemaExtensionId, finding.KindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
	require.NoError(t, err, "unexpected error occurred when creating finding")
	return finding
}

// Graph Schema Extensions may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_GraphSchemaExtensions_CRUD(t *testing.T) {
	var (
		extension1 = model.GraphSchemaExtension{
			Name:        "adam",
			DisplayName: "test extension name 1",
			Version:     "1.0.0",
			Namespace:   "Test",
		}
		extension2 = model.GraphSchemaExtension{
			Name:        "bob",
			DisplayName: "test extension name 2",
			Version:     "2.0.0",
			Namespace:   "Test2",
		}
		extension3 = model.GraphSchemaExtension{
			Name:        "charlie",
			DisplayName: "another extension",
			Version:     "3.0.0",
			Namespace:   "AA",
		}
		extension4 = model.GraphSchemaExtension{
			Name:        "david",
			DisplayName: "yet another extension",
			Version:     "4.0.0",
			Namespace:   "ZZ",
		}
	)

	// Helper function to create all test extensions
	createTestExtensions := func(t *testing.T, testSuite IntegrationTestSuite) {
		t.Helper()

		for _, ext := range []model.GraphSchemaExtension{extension1, extension2, extension3, extension4} {
			_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(
				testSuite.Context, ext.Name, ext.DisplayName, ext.Version, ext.Namespace,
			)
			require.NoError(t, err, "unexpected error occurred when creating test extensions")
		}
	}

	// Helper function to check if an extension with matching fields exists
	assertContainsExtension := func(extensions model.GraphSchemaExtensions, expected model.GraphSchemaExtension) bool {
		for _, ext := range extensions {
			if ext.Name == expected.Name &&
				ext.DisplayName == expected.DisplayName &&
				ext.Version == expected.Version &&
				ext.Namespace == expected.Namespace {
				return true
			}
		}
		return false
	}

	type args struct {
		filters     model.Filters
		sort        model.Sort
		skip, limit int
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, testSuite IntegrationTestSuite, args args)
	}{
		// CreateGraphSchemaExtension
		{
			name: "Success: extension created",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				// Create new extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, extension1.Name, extension1.DisplayName, extension1.Version, extension1.Namespace)
				assert.NoError(t, err, "unexpected error occurred when creating extension")
			},
		},
		{
			name: "Error: fail to create duplicate schema extension name",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				createTestExtensions(t, testSuite)

				// Insert graph extension that already exists
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, extension1.Name, extension1.DisplayName, extension1.Version, extension1.Namespace)
				assert.EqualError(t, err, "duplicate graph schema extension name: adam")
			},
		},
		{
			name: "Error: fail to create duplicate schema extension namespace",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				createTestExtensions(t, testSuite)

				// Insert graph extension that already exists
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "different", extension1.DisplayName, extension1.Version, extension1.Namespace)
				assert.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionNamespace)
			},
		},
		// GetGraphSchemaExtensionById
		{
			name: "Success: retrieves graph extension by id",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, extension1.Name, extension1.DisplayName, extension1.Version, extension1.Namespace)

				retrieved, err := testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				assert.NoError(t, err, "unexpected error occurred when getting extension by id")

				// Assert extension has been created as expected
				assert.True(t, assertContainsExtension(model.GraphSchemaExtensions{extension}, retrieved), "extension1 should exist in results")
			},
		},
		// GetGraphSchemaExtensions
		{
			name: "Error: parseFiltersAndPagination",
			args: args{
				filters: model.Filters{
					"`": []model.Filter{
						{},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				_, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.EqualError(t, err, "invalid operator specified")
			},
		},
		{
			name: "Success: returns extensions, no filter or sorting",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 4 new records were created in this test
				assert.Equal(t, 4, total-baselineCount, "expected 4 new extensions")

				// Validate all created extensions exist in the results
				assert.True(t, assertContainsExtension(extensions, extension1), "extension1 should exist in results")
				assert.True(t, assertContainsExtension(extensions, extension2), "extension2 should exist in results")
				assert.True(t, assertContainsExtension(extensions, extension3), "extension3 should exist in results")
				assert.True(t, assertContainsExtension(extensions, extension4), "extension4 should exist in results")
			},
		},
		{
			name: "Success: returns extensions, with filtering",
			args: args{
				filters: model.Filters{
					"name": []model.Filter{
						{
							Operator: model.Equals,
							Value:    "david",
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 1 matching record
				assert.Equal(t, 1, total-baselineCount, "expected 1 extension matching the filter")

				// Should only contain extension4 (david)
				assert.True(t, assertContainsExtension(extensions, extension4), "extension4 should exist in results")

				// Should not contain the others
				assert.False(t, assertContainsExtension(extensions, extension1), "extension1 should not be in filtered results")
				assert.False(t, assertContainsExtension(extensions, extension2), "extension2 should not be in filtered results")
				assert.False(t, assertContainsExtension(extensions, extension3), "extension3 should not be in filtered results")
			},
		},
		{
			name: "Success: returns extensions, with multiple filters",
			args: args{
				filters: model.Filters{
					"name": []model.Filter{
						{
							Operator: model.Equals,
							Value:    "david",
						},
					},
					"display_name": []model.Filter{
						{
							Operator: model.Equals,
							Value:    "yet another extension",
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 1 matching record
				assert.Equal(t, 1, total-baselineCount, "expected 1 extension matching both filters")
				// Should contain only extension4 (matches both filters)
				assert.True(t, assertContainsExtension(extensions, extension4), "extension4 should exist in results")
			},
		},
		{
			name: "Success: returns extensions, with fuzzy filtering",
			args: args{
				filters: model.Filters{
					"display_name": []model.Filter{
						{
							Operator: model.ApproximatelyEquals,
							Value:    "test extension",
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")
				// Should contain extension1 & extension2 (matches fuzzy filter)
				assert.True(t, assertContainsExtension(extensions, extension1), "extension1 should exist in results")
				assert.True(t, assertContainsExtension(extensions, extension2), "extension2 should exist in results")
			},
		},
		{
			name: "Success: returns extensions, with fuzzy filtering and sort ascending",
			args: args{
				filters: model.Filters{
					"display_name": []model.Filter{
						{
							Operator: model.ApproximatelyEquals,
							Value:    "test extension",
						},
					},
				},
				sort: model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")
				// Assert extensions retrieved are sorted in ascending order by display name
				assert.Equal(t, extensions[0].DisplayName, "test extension name 1", "expected display name to be in ascending order")
				assert.Equal(t, extensions[1].DisplayName, "test extension name 2", "expected display name to be in ascending order")
			},
		},
		{
			name: "Success: returns extensions, with fuzzy filtering and sort descending",
			args: args{
				filters: model.Filters{
					"display_name": []model.Filter{{
						Operator: model.ApproximatelyEquals,
						Value:    "test extension",
					},
					},
				},
				sort: model.Sort{{Column: "display_name", Direction: model.DescendingSortDirection}},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")
				// Assert extensions retrieved (extension1 & extension2) are sorted in descending order by display name
				assert.Equal(t, extensions[0].DisplayName, "test extension name 2", "expected display name to be in descending order")
				assert.Equal(t, extensions[1].DisplayName, "test extension name 1", "expected display name to be in descending order")
			},
		},
		{
			name: "Success: returns extensions, no filter or sorting, with skip",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				// Get extensions after inserting test data
				_, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 4 matching records
				assert.Equal(t, 4, total-baselineCount, "expected 4 extensions")
			},
		},
		{
			name: "Success: returns extensions, no filter or sorting, with limit",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				createTestExtensions(t, testSuite)

				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")
				// Assert total records returned includes the number of records pre-inserted + the number of records created in this test
				assert.Equal(t, baselineCount+4, total, "expected all extension records (6) returned")
				// Assert 1 record matching limit
				assert.Len(t, extensions, 1, "expected 1 extension returned due to limit")
			},
		},
		{
			name: "Error: returns an error with bogus filtering",
			args: args{
				filters: model.Filters{
					"nonexistentcolumn": []model.Filter{
						{
							Operator: model.Equals,
							Value:    "david",
						},
					},
				},
				limit: 1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				createTestExtensions(t, testSuite)

				// Get extensions after inserting test data
				extensions, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no extensions are returned
				assert.Len(t, extensions, 0, "expected 0 extensions returned due on error")
			},
		},
		// UpdateGraphSchemaExtension
		{
			name: "Success: extension updated",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, extension1.Name, extension1.DisplayName, extension1.Version, extension1.Namespace)

				newlyCreatedExtension, err := testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving newly created extension")

				// Modify some fields (not is_builtin)
				newlyCreatedExtension.Name = "new name"
				newlyCreatedExtension.DisplayName = "new display name"
				newlyCreatedExtension.Version = "v5.0.0"
				newlyCreatedExtension.Namespace = "different namespace"

				// Update in database
				updatedExtension, err := testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, newlyCreatedExtension)
				assert.NoError(t, err, "unexpected error occurred when updating extension")

				// Validate fields are updated
				assert.Equal(t, newlyCreatedExtension.Name, updatedExtension.Name)
				assert.Equal(t, newlyCreatedExtension.DisplayName, updatedExtension.DisplayName)
				assert.Equal(t, newlyCreatedExtension.Version, updatedExtension.Version)
				assert.Equal(t, newlyCreatedExtension.Namespace, updatedExtension.Namespace)
			},
		},
		{
			name: "Error: failed to update with duplicate namespace",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Create new extension
				extension := createTestExtension(t, testSuite, extension1.Name, extension1.DisplayName, extension1.Version, extension1.Namespace)

				// Create a second extension to test a duplicated namespace
				secondExtension := createTestExtension(t, testSuite, "name", "display", "v1.0.0", "duplicate")

				newlyCreatedExtension, err := testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving newly created extension")

				// Modify some fields (not is_builtin)
				newlyCreatedExtension.Name = "new name"
				newlyCreatedExtension.DisplayName = "new display name"
				newlyCreatedExtension.Version = "v5.0.0"
				// Modify extension to the second extension's namespace. Since this is unique, it will cause an error.
				newlyCreatedExtension.Namespace = secondExtension.Namespace

				// Update in database
				_, err = testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, newlyCreatedExtension)
				assert.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionNamespace)
			},
		},
		{
			name: "Error: failed to update with duplicate name",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Create new extension
				extension := createTestExtension(t, testSuite, extension1.Name, extension1.DisplayName, extension1.Version, extension1.Namespace)

				// Create a second extension to test a duplicated name
				secondExtension := createTestExtension(t, testSuite, "duplicate", "display", "v1.0.0", "random")

				newlyCreatedExtension, err := testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving newly created extension")

				// Modify some fields (not is_builtin)
				// Modify extension to the second extension's name. Since this is unique, it will cause an error.
				newlyCreatedExtension.Name = secondExtension.Name
				newlyCreatedExtension.DisplayName = "new display name"
				newlyCreatedExtension.Version = "v5.0.0"
				newlyCreatedExtension.Namespace = "different namespace"

				// Update in database
				_, err = testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, newlyCreatedExtension)
				assert.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionName)
			},
		},
		{
			name: "Error: failed to update extension that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Update in database
				_, err := testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, model.GraphSchemaExtension{Serial: model.Serial{ID: int32(5000)}})
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// DeleteGraphSchemaExtension
		{
			name: "Success: extension deleted",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, extension1.Name, extension1.DisplayName, extension1.Version, extension1.Namespace)

				// Delete Extension
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extension.ID)
				require.NoError(t, err, "unexpected error occurred when deleting extension")

				// Validate it's no longer there
				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Error: failed to delete extension that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Delete extension
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, int32(5000))
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Success: Source Kind deactivated when Extension is deleted",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				createdExtension := createTestExtension(t, testSuite, "TestGraphSchemaExtension", "Test Graph Schema Extension", "1.0.0", "TGSE")
				createdEnvironmentNode := createTestNodeKind(t, testSuite, "TGSE_Environment 1", createdExtension.ID, "Environment 1", "an environment kind", false, "", "")
				createdSourceKindNode := createTestNodeKind(t, testSuite, "Source_Kind_1", createdExtension.ID, "Source Kind 1", "a source kind", false, "", "")

				envKind := getKindByName(t, testSuite, createdEnvironmentNode.Name)
				sourceKind := registerAndGetSourceKind(t, testSuite, createdSourceKindNode.Name)

				// Create Environment
				_, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID, envKind.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when creating environment")

				// Delete extension
				err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, createdExtension.ID)
				require.NoError(t, err, "unexpected error occurred when deleting extension")

				// Validate Source Kind has been deactivated (no longer able to retrieve)
				_, err = testSuite.BHDatabase.GetSourceKindByID(testSuite.Context, sourceKind.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Success: Multiple Source Kinds are deactivated when Extension is deleted",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				createdExtension := createTestExtension(t, testSuite, "TestGraphSchemaExtension", "Test Graph Schema Extension", "1.0.0", "TGSE")

				createdEnvironmentNode1 := createTestNodeKind(t, testSuite, "TGSE_Environment 1", createdExtension.ID, "Environment 1", "an environment kind", false, "", "")
				createdEnvironmentNode2 := createTestNodeKind(t, testSuite, "TGSE_Environment 2", createdExtension.ID, "Environment 2", "an environment kind", false, "", "")
				createdSourceKindNode1 := createTestNodeKind(t, testSuite, "Source_Kind_1", createdExtension.ID, "Source Kind 1", "a source kind", false, "", "")
				createdSourceKindNodeB := createTestNodeKind(t, testSuite, "Source_Kind_2", createdExtension.ID, "Source Kind 2", "a source kind", false, "", "")

				environmentKind1 := getKindByName(t, testSuite, createdEnvironmentNode1.Name)
				environmentKind2 := getKindByName(t, testSuite, createdEnvironmentNode2.Name)
				retrievedSourceKindA := registerAndGetSourceKind(t, testSuite, createdSourceKindNode1.Name)
				retrievedSourceKindB := registerAndGetSourceKind(t, testSuite, createdSourceKindNodeB.Name)

				// Create Environment 1 with Source Kind 1
				_, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID, environmentKind1.ID, int32(retrievedSourceKindA.ID))
				require.NoError(t, err, "unexpected error occurred when creating Environment 1")

				// Create Environment 2 with Source Kind 2
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID, environmentKind2.ID, int32(retrievedSourceKindB.ID))
				require.NoError(t, err, "unexpected error occurred when creating Environment 2")

				// Delete Extension
				err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, createdExtension.ID)
				require.NoError(t, err, "unexpected error occurred when deleting extension")

				// Validate Source Kind has been deactivated (no longer able to retrieve) for both Source Kinds A & B
				_, err = testSuite.BHDatabase.GetSourceKindByID(testSuite.Context, retrievedSourceKindA.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)

				_, err = testSuite.BHDatabase.GetSourceKindByID(testSuite.Context, retrievedSourceKindB.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Success: Source Kind is NOT deactivated when multiple extensions environments uses the same source kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				createdExtension1 := createTestExtension(t, testSuite, "TestGraphSchemaExtension1", "Test Graph Schema Extension 1", "1.0.0", "TGSE-A")
				createdExtension2 := createTestExtension(t, testSuite, "TestGraphSchemaExtension2", "Test Graph Schema Extension 2", "1.0.0", "TGSE-B")

				createdEnvironmentNode1 := createTestNodeKind(t, testSuite, "TGSE_Environment 1", createdExtension1.ID, "Environment 1", "an environment kind", false, "", "")
				createdEnvironmentNode2 := createTestNodeKind(t, testSuite, "TGSE_Environment 2", createdExtension2.ID, "Environment 2", "an environment kind", false, "", "")
				createdSourceKindNode := createTestNodeKind(t, testSuite, "Source_Kind_1", createdExtension1.ID, "Source Kind 1", "a source kind", false, "", "")

				environmentKind1 := getKindByName(t, testSuite, createdEnvironmentNode1.Name)
				environmentKind2 := getKindByName(t, testSuite, createdEnvironmentNode2.Name)
				sourceKind := registerAndGetSourceKind(t, testSuite, createdSourceKindNode.Name)

				// Create Environment for Extension 1 using same Source Kind
				_, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension1.ID, environmentKind1.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when creating Environment 1")

				// Create Environment for Extension 2 using same Source Kind
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension2.ID, environmentKind2.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when creating Environment 2")

				// Delete Extension 1
				err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, createdExtension1.ID)
				require.NoError(t, err, "unexpected error occurred when deleting exension")

				// Validate Source Kind has NOT been deactivated
				retrievedSourceKind, err := testSuite.BHDatabase.GetSourceKindByID(testSuite.Context, sourceKind.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving source kind by id")
				// Retrieved Source Kind should still exist and be same Source Kind retrieved above
				assert.Equal(t, retrievedSourceKind, sourceKind)
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite, testCase.args)
		})
	}
}

// Graph Schema Node Kinds may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_GraphSchemaNodeKind_CRUD(t *testing.T) {
	// Helper functions to assert on node kind fields
	assertContainsNodeKinds := func(t *testing.T, got model.GraphSchemaNodeKinds, expected ...model.GraphSchemaNodeKind) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, nk := range got {
				if nk.Name == want.Name &&
					nk.SchemaExtensionId == want.SchemaExtensionId &&
					nk.DisplayName == want.DisplayName &&
					nk.Description == want.Description &&
					nk.IsDisplayKind == want.IsDisplayKind &&
					nk.Icon == want.Icon &&
					nk.IconColor == want.IconColor {

					// Additional validations for the found item
					assert.GreaterOrEqualf(t, nk.ID, int32(1), "NodeKind %v - ID is invalid", nk.Name)
					assert.Falsef(t, nk.CreatedAt.IsZero(), "NodeKind %v - created_at is zero", nk.Name)
					assert.Falsef(t, nk.UpdatedAt.IsZero(), "NodeKind %v - updated_at is zero", nk.Name)
					assert.Falsef(t, nk.DeletedAt.Valid, "NodeKind %v - deleted_at should be null", nk.Name)

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected node kind %v not found", want.Name)
		}
	}

	assertContainsNodeKind := func(t *testing.T, got model.GraphSchemaNodeKind, expected ...model.GraphSchemaNodeKind) {
		t.Helper()
		assertContainsNodeKinds(t, model.GraphSchemaNodeKinds{got}, expected...)
	}

	assertDoesNotContainNodeKinds := func(t *testing.T, got model.GraphSchemaNodeKinds, expected ...model.GraphSchemaNodeKind) {
		t.Helper()
		for _, want := range expected {
			for _, nk := range got {
				if nk.Name == want.Name &&
					nk.SchemaExtensionId == want.SchemaExtensionId &&
					nk.DisplayName == want.DisplayName &&
					nk.Description == want.Description &&
					nk.IsDisplayKind == want.IsDisplayKind &&
					nk.Icon == want.Icon &&
					nk.IconColor == want.IconColor {

					assert.Failf(t, "Unexpected node kind found", "Node kind %v should not be present", want.Name)
				}
			}
		}
	}

	type args struct {
		filters     model.Filters
		sort        model.Sort
		skip, limit int
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, testSuite IntegrationTestSuite, args args)
	}{
		// CreateGraphSchemaNodeKind
		{
			name: "Success: create a schema node kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				got, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
				assert.NoError(t, err, "unexpected error occurred when creating node kind")

				assertContainsNodeKind(t, got, nodeKind)
			},
		},
		{
			name: "Error: fails to create schema node kind that does not have a unique name",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)

				// Create same node again
				_, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaNodeKindName)
			},
		},
		// GetGraphSchemaNodeKindById
		{
			name: "Success: get schema node kind by id",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createdNodeKind := createTestNodeKind(t, testSuite, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)

				retrievedNodeKind, err := testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, createdNodeKind.ID)
				assert.NoError(t, err, "unexpected error occurred getting node kind by id")

				assertContainsNodeKind(t, retrievedNodeKind, nodeKind)
			},
		},
		{
			name: "Error: fail to retrieve a node kind that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, err := testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, 112)
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetGraphSchemaNodeKinds
		{
			name: "Error: parseFiltersAndPagination",
			args: args{
				filters: model.Filters{"`": []model.Filter{{}}},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, _, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "invalid operator specified")
			},
		},
		{
			name: "Success: return node schema kinds, no filter or sorting",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}
				nodeKind2 := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				createTestNodeKind(t, testSuite, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)

				// Get Node Kinds
				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Validate number of results
				assert.Equal(t, baselineCount+2, total, "expected total node kinds to be equal to how many node kinds exist in the database")
				assert.Len(t, nodeKinds, baselineCount+2, "expected all node kinds to be returned when no filter/sorting")

				// Validate all created nodeKinds exist in the results
				assertContainsNodeKinds(t, nodeKinds, nodeKind1, nodeKind2)
			},
		},
		{
			name: "Success: return schema node kinds using a filter",
			args: args{
				filters: model.Filters{"name": []model.Filter{
					{
						Operator:    model.Equals,
						Value:       "Test_Kind_2",
						SetOperator: model.FilterAnd,
					},
				},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind1 := model.GraphSchemaNodeKind{Name: "Test_Kind_1", SchemaExtensionId: extension.ID, DisplayName: "Test_Kind_1", Description: "A test kind", Icon: "test_icon", IconColor: "blue"}
				nodeKind2 := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "A test kind",
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				createTestNodeKind(t, testSuite, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)

				// Get Node Kinds
				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Validate number of results
				assert.Equal(t, 1, total, "expected total node kinds to be equal to how many results are returned from database")
				assert.Len(t, nodeKinds, 1, "expected 1 node kind to be returned when filtering by name")

				// Validate expected nodeKinds exist in the results
				assertContainsNodeKinds(t, nodeKinds, nodeKind2)
				assertDoesNotContainNodeKinds(t, nodeKinds, nodeKind1)
			},
		},
		{
			name: "Success: returns schema node kinds, with fuzzy filtering",
			args: args{
				filters: model.Filters{"name": []model.Filter{
					{
						Operator:    model.ApproximatelyEquals,
						Value:       "Test Kind ",
						SetOperator: model.FilterAnd,
					},
				},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}
				nodeKind2 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				nodeKind3 := model.GraphSchemaNodeKind{
					Name:              "Test Does Not Match Kind",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_3",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				createTestNodeKind(t, testSuite, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				createTestNodeKind(t, testSuite, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)

				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Validate number of results
				assert.Equal(t, 2, total, "expected total node kinds to be equal to how many results are returned from database")
				assert.Len(t, nodeKinds, 2, "expected 2 node kinds to be returned when fuzzy filtering by name")

				// Validate node kind is filtered by input argument
				assertContainsNodeKinds(t, nodeKinds, nodeKind2)
				assertDoesNotContainNodeKinds(t, nodeKinds, nodeKind3)
			},
		},
		{
			name: "Success: returns schema node kinds, with fuzzy filtering and sort ascending on description",
			args: args{
				filters: model.Filters{"name": []model.Filter{
					{
						Operator:    model.ApproximatelyEquals,
						Value:       "Test Kind ",
						SetOperator: model.FilterAnd,
					},
				},
				},
				sort: model.Sort{
					{
						Direction: model.AscendingSortDirection,
						Column:    "description",
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "Beta",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}
				nodeKind2 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "Alpha",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				nodeKind3 := model.GraphSchemaNodeKind{
					Name:              "Test Does Not Match Kind",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_3",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				createTestNodeKind(t, testSuite, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				createTestNodeKind(t, testSuite, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)

				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Validate number of results
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (nodeKind1 & nodeKind2) are sorted in ascending order by description
				assert.Equal(t, nodeKinds[0].Description, "Alpha", "expected description to be in ascending order")
				assert.Equal(t, nodeKinds[1].Description, "Beta", "expected description to be in ascending order")
			},
		},
		{
			name: "Success: returns schema node kinds, with fuzzy filtering and sort descending on description",
			args: args{
				filters: model.Filters{"name": []model.Filter{
					{
						Operator:    model.ApproximatelyEquals,
						Value:       "Test Kind ",
						SetOperator: model.FilterAnd,
					},
				},
				},
				sort: model.Sort{
					{
						Direction: model.DescendingSortDirection,
						Column:    "description",
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "Beta",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}
				nodeKind2 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "Alpha",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				nodeKind3 := model.GraphSchemaNodeKind{
					Name:              "Test Does Not Match Kind",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_3",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				createTestNodeKind(t, testSuite, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				createTestNodeKind(t, testSuite, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)

				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (nodeKind1 & nodeKind2) are sorted in ascending order by description
				assert.Equal(t, nodeKinds[0].Description, "Beta", "expected description to be in descending order")
				assert.Equal(t, nodeKinds[1].Description, "Alpha", "expected description to be in descending order")
			},
		},
		{
			name: "Success: returns schema node kinds, no filter or sorting, with skip",
			args: args{skip: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "Beta",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}
				nodeKind2 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "Alpha",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				nodeKind3 := model.GraphSchemaNodeKind{
					Name:              "Test Does Not Match Kind",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_3",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				createTestNodeKind(t, testSuite, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				createTestNodeKind(t, testSuite, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)

				_, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Assert 3 matching records
				assert.Equal(t, 3, total-baselineCount, "expected 3 node kinds")
			},
		},
		{
			name: "Success: returns schema node kinds, no filter or sorting, with limit",
			args: args{limit: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "Beta",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}
				nodeKind2 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "Alpha",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				nodeKind3 := model.GraphSchemaNodeKind{
					Name:              "Test Does Not Match Kind",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_3",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createTestNodeKind(t, testSuite, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				createTestNodeKind(t, testSuite, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				createTestNodeKind(t, testSuite, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)

				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Assert total records returned includes the number of records pre-inserted + the number of records created in this test
				assert.Equal(t, baselineCount+3, total, "expected all node kind records (6) returned")
				// Assert 1 record matching limit
				assert.Len(t, nodeKinds, 1, "expected 1 node kind returned due to limit")
			},
		},
		{
			name: "Error: returns an error with bogus filtering",
			args: args{
				filters: model.Filters{
					"nonexistentcolumn": []model.Filter{
						{
							Operator: model.Equals,
							Value:    "david",
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				nodeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no nodeKinds are returned
				assert.Len(t, nodeKinds, 0, "expected 0 node kinds returned due on error")
			},
		},
		// UpdateGraphSchemaNodeKind
		{
			name: "Success: update schema node kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
				createdNodeKind := createTestNodeKind(t, testSuite, "Test Kind 1", extension.ID, "Test_Kind_1", "Alpha", false, "test_icon", "blue")

				updatedNodeKind1 := model.GraphSchemaNodeKind{
					Serial:            model.Serial{ID: createdNodeKind.ID},
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Display Name",
					Description:       "Beta",
					IsDisplayKind:     false,
					Icon:              "test_icon_color",
					IconColor:         "green",
				}

				// Update Node Kind 1
				updatedNodeKind, err := testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, updatedNodeKind1)
				require.NoError(t, err, "unexpected error occurred when updating node kind")

				// Retrieve Node Kind 1
				nodeKindWithChanges, err := testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, updatedNodeKind.ID)
				require.NoError(t, err, "unexpected error occurred when retrieving node kind")

				// Validate updated fields
				assertContainsNodeKind(t, updatedNodeKind, nodeKindWithChanges)
			},
		},
		{
			name: "Error: failed to update schema node kind that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				// Update Node Kind 1
				_, err := testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, model.GraphSchemaNodeKind{
					Name:              "does not exist",
					SchemaExtensionId: extension.ID,
				})
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// DeleteGraphSchemaNodeKind
		{
			name: "Success: deleted node kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
				insertedNodeKind := createTestNodeKind(t, testSuite, "Test Kind 1", extension.ID, "Test_Kind_1", "Beta", false, "test_icon", "blue")

				err := testSuite.BHDatabase.DeleteGraphSchemaNodeKind(testSuite.Context, insertedNodeKind.ID)
				assert.NoError(t, err, "unexpected error occurred when deleting node kind 1")

				// Validate Node Kind no longer exists
				_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, insertedNodeKind.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete schema node kind that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				err := testSuite.BHDatabase.DeleteGraphSchemaNodeKind(testSuite.Context, int32(10000))
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite, testCase.args)
		})
	}
}

// Graph Schema Properties may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_GraphSchemaProperties_CRUD(t *testing.T) {
	// Helper functions to assert on properties
	assertContainsProperties := func(t *testing.T, got model.GraphSchemaProperties, expected ...model.GraphSchemaProperty) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, prop := range got {
				if prop.Name == want.Name &&
					prop.SchemaExtensionId == want.SchemaExtensionId &&
					prop.DisplayName == want.DisplayName &&
					prop.DataType == want.DataType &&
					prop.Description == want.Description {

					// Additional validations for the found item
					assert.GreaterOrEqualf(t, prop.ID, int32(1), "Property %v - ID is invalid", prop.Name)
					assert.Falsef(t, prop.CreatedAt.IsZero(), "Property %v - created_at is zero", prop.Name)
					assert.Falsef(t, prop.UpdatedAt.IsZero(), "Property %v - updated_at is zero", prop.Name)
					assert.Falsef(t, prop.DeletedAt.Valid, "Property %v - deleted_at should be null", prop.Name)

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected property %v not found", want.Name)
		}
	}

	assertContainsProperty := func(t *testing.T, got model.GraphSchemaProperty, expected ...model.GraphSchemaProperty) {
		t.Helper()
		assertContainsProperties(t, model.GraphSchemaProperties{got}, expected...)
	}

	assertDoesNotContainProperties := func(t *testing.T, got model.GraphSchemaProperties, expected ...model.GraphSchemaProperty) {
		t.Helper()
		for _, want := range expected {
			for _, prop := range got {
				if prop.Name == want.Name &&
					prop.SchemaExtensionId == want.SchemaExtensionId &&
					prop.DisplayName == want.DisplayName &&
					prop.DataType == want.DataType &&
					prop.Description == want.Description {

					assert.Failf(t, "Unexpected property found", "Property %v should not be present", want.Name)
				}
			}
		}
	}

	type args struct {
		filters     model.Filters
		sort        model.Sort
		skip, limit int
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, testSuite IntegrationTestSuite, args args)
	}{
		// CreateGraphSchemaProperty
		{
			name: "Success: create a property",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.NoError(t, err, "unexpected error occurred when creating property")

				assertContainsProperty(t, got, property)
			},
		},
		{
			name: "Error: fails to create duplicate property, name must be unique per extension",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				// Create property
				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "unexpected error occurred when creating property")
				assertContainsProperty(t, got, property)

				// Create same property again
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionPropertyName)
			},
		},
		{
			name: "Success: creates same property (without name collision) on multiple extensions",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				// Create property
				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "unexpected error occurred when creating property")

				// Validate expected property was created
				assertContainsProperty(t, got, property)

				// Create new schema extension
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "2.0.0", "Test_2")

				// Modify to point at different extension
				property.SchemaExtensionId = extension2.ID

				// Create same property again on new extension which is expected to be successful
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.NoError(t, err, "error creating schema property but no error was expected")
			},
		},
		// GetGraphSchemaPropertyById
		{
			name: "Success: get property by id",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				// Create new property
				newProperty, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "unexpected error occurred when creating property")

				// Validate property
				retrievedProperty, err := testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, newProperty.ID)
				assert.NoError(t, err, "failed to get property by id")

				assertContainsProperty(t, retrievedProperty, newProperty)
			},
		},
		{
			name: "Error: fail to retrieve property that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, err := testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetGraphSchemaProperties
		{
			name: "Error: parseFiltersAndPagination",
			args: args{
				filters: model.Filters{"`": []model.Filter{{}}},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, _, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "invalid operator specified")
			},
		},
		{
			name: "Success: return properties, no filter or sorting",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting baseline count")

				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test2")

				extension1Property1 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}
				extension1Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "integer",
					Description:       "Mediocre and average extension property",
				}
				extension2Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension2.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "array",
					Description:       "Extremely boring and lame extension property 2",
				}

				// Create Prop 1 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving properties")

				// Validate number of results
				assert.Equal(t, baselineCount+3, total, "expected total properties to be equal to how many properties exist in the database")
				assert.Len(t, properties, baselineCount+3, "expected all properties to be returned when no filter/sorting")

				// Validate all created properties exist in the results
				assertContainsProperties(t, properties, extension1Property1, extension1Property2, extension2Property2)
			},
		},
		{
			name: "Success: return properties using a filter",
			args: args{
				filters: model.Filters{"data_type": []model.Filter{
					{
						Operator:    model.Equals,
						Value:       "array",
						SetOperator: model.FilterAnd,
					},
				},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test2")

				extension1Property1 := model.GraphSchemaProperty{SchemaExtensionId: extension1.ID, Name: "ext_prop_1", DisplayName: "Extension Property 1", DataType: "string", Description: "Extremely boring and lame extension property 1"}
				extension1Property2 := model.GraphSchemaProperty{SchemaExtensionId: extension1.ID, Name: "ext_prop_2", DisplayName: "Extension Property 2", DataType: "integer", Description: "Mediocre and average extension property"}
				extension2Property2 := model.GraphSchemaProperty{SchemaExtensionId: extension2.ID, Name: "ext_prop_2", DisplayName: "Extension Property 2", DataType: "array", Description: "Extremely boring and lame extension property 2"}

				// Create Prop 1 for Extension 1
				_, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving properties")

				// Validate number of results
				assert.Equal(t, 1, total, "expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 1, "expected 1 property to be returned when filtering by description")

				// Validate expected properties exist in the results
				assertContainsProperties(t, properties, extension2Property2)
				assertDoesNotContainProperties(t, properties, extension1Property1, extension1Property2)
			},
		},
		{
			name: "Success: returns properties, with fuzzy filtering",
			args: args{
				filters: model.Filters{"description": []model.Filter{
					{
						Operator:    model.ApproximatelyEquals,
						Value:       "Extremely boring and lame extension property ",
						SetOperator: model.FilterAnd,
					},
				},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test2")

				extension1Property1 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}
				extension1Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "integer",
					Description:       "Mediocre and average extension property",
				}
				extension2Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension2.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "array",
					Description:       "Extremely boring and lame extension property 2",
				}

				// Create Prop 1 for Extension 1
				_, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving properties")

				// Validate number of results
				assert.Equal(t, 2, total, "expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 2, "expected 2 properties to be returned when fuzzy filtering by description")

				// Validate expected properties exist in the results
				assertContainsProperties(t, properties, extension1Property1, extension2Property2)
				assertDoesNotContainProperties(t, properties, extension1Property2)
			},
		},
		{
			name: "Success: returns properties, with fuzzy filtering and sort ascending on description",
			args: args{
				filters: model.Filters{"description": []model.Filter{
					{
						Operator:    model.ApproximatelyEquals,
						Value:       "Extremely boring and lame extension property ",
						SetOperator: model.FilterAnd,
					},
				},
				},
				sort: model.Sort{
					{
						Direction: model.AscendingSortDirection,
						Column:    "description",
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test2")

				extension1Property1 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property Beta",
				}
				extension1Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "integer",
					Description:       "Mediocre and average extension property",
				}
				extension2Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension2.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "array",
					Description:       "Extremely boring and lame extension property Alpha",
				}

				// Create Prop 1 for Extension 1
				_, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving properties")

				// Validate number of results
				assert.Equal(t, 2, total, "expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 2, "expected 2 properties to be returned when fuzzy filtering by description")

				// Assert extensions retrieved (extension1Property1 & extension2Property2) are sorted in ascending order by description
				assert.Equal(t, properties[0].Description, "Extremely boring and lame extension property Alpha", "expected description to be in ascending order")
				assert.Equal(t, properties[1].Description, "Extremely boring and lame extension property Beta", "expected description to be in ascending order")
			},
		},
		{
			name: "Success: returns properties, with fuzzy filtering and sort descending on description",
			args: args{
				filters: model.Filters{"description": []model.Filter{
					{
						Operator:    model.ApproximatelyEquals,
						Value:       "Extremely boring and lame extension property ",
						SetOperator: model.FilterAnd,
					},
				},
				},
				sort: model.Sort{
					{
						Direction: model.DescendingSortDirection,
						Column:    "description",
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test2")

				extension1Property1 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property Beta",
				}
				extension1Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "integer",
					Description:       "Mediocre and average extension property",
				}
				extension2Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension2.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "array",
					Description:       "Extremely boring and lame extension property Alpha",
				}

				// Create Prop 1 for Extension 1
				_, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving properties")

				// Validate number of results
				assert.Equal(t, 2, total, "expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 2, "expected 2 properties to be returned when fuzzy filtering by description")

				// Assert extensions retrieved (extension1Property1 & extension2Property2) are sorted in descending order by description
				assert.Equal(t, properties[0].Description, "Extremely boring and lame extension property Beta", "expected description to be in descending order")
				assert.Equal(t, properties[1].Description, "Extremely boring and lame extension property Alpha", "expected description to be in descending order")
			},
		},
		{
			name: "Success: returns properties, no filter or sorting, with skip",
			args: args{skip: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting baseline count")

				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test2")

				extension1Property1 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}
				extension1Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "integer",
					Description:       "Mediocre and average extension property",
				}
				extension2Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension2.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "array",
					Description:       "Extremely boring and lame extension property 2",
				}

				// Create Prop 1 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 2")

				// Get Properties back
				_, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving properties")

				// Assert 3 matching records
				assert.Equal(t, 3, total-baselineCount, "expected 3 properties")
			},
		},
		{
			name: "Success: returns schema node kinds, no filter or sorting, with limit",
			args: args{limit: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting baseline count")

				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test2")

				extension1Property1 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}
				extension1Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "integer",
					Description:       "Mediocre and average extension property",
				}
				extension2Property2 := model.GraphSchemaProperty{
					SchemaExtensionId: extension2.ID,
					Name:              "ext_prop_2",
					DisplayName:       "Extension Property 2",
					DataType:          "array",
					Description:       "Extremely boring and lame extension property 2",
				}

				// Create Prop 1 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property1.SchemaExtensionId, extension1Property1.Name, extension1Property1.DisplayName, extension1Property1.DataType, extension1Property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving properties")

				// Assert 3 matching records
				assert.Equal(t, 3, total-baselineCount, "expected 3 properties")

				// Assert total records returned includes the number of records pre-inserted + the number of records created in this test
				assert.Equal(t, baselineCount+3, total, "expected all properties returned")
				// Assert 1 record matching limit
				assert.Len(t, properties, 1, "expected 1 property returned due to limit")
			},
		},
		{
			name: "Error: returns an error with bogus filtering",
			args: args{
				filters: model.Filters{
					"nonexistentcolumn": []model.Filter{
						{
							Operator: model.Equals,
							Value:    "david",
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Get Properties back
				properties, _, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no properties are returned
				assert.Len(t, properties, 0, "expected 0 properties returned due on error")
			},
		},
		// UpdateGraphSchemaProperty
		{
			name: "Success: update property",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}

				// Create Property for Extension
				newProperty, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "unexpected error occurred when creating property for extension")

				updatedProperty := model.GraphSchemaProperty{
					Serial:            model.Serial{ID: newProperty.ID},
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1_update",
					DisplayName:       "Extension Property 1 Updated",
					DataType:          "string updated",
					Description:       "Extremely boring and lame extension property 1 updated",
				}

				// Update Property
				updatedProperty, err = testSuite.BHDatabase.UpdateGraphSchemaProperty(testSuite.Context, updatedProperty)
				require.NoError(t, err, "failed to update property for extension")

				// Retrieve Property
				propWithChanges, err := testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, updatedProperty.ID)
				require.NoError(t, err, "unexpected error occurred when getting property by id")

				// Assert on updated fields
				assert.Equal(t, updatedProperty.Name, propWithChanges.Name)
				assert.Equal(t, updatedProperty.DisplayName, propWithChanges.DisplayName)
				assert.Equal(t, updatedProperty.DataType, propWithChanges.DataType)
				assert.Equal(t, updatedProperty.Description, propWithChanges.Description)
			},
		},
		{
			name: "Error: duplicate property name",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				property1 := model.GraphSchemaProperty{SchemaExtensionId: extension.ID, Name: "ext_prop_1", DisplayName: "Extension Property 1", DataType: "string", Description: "Extremely boring and lame extension property 1"}
				property2 := model.GraphSchemaProperty{SchemaExtensionId: extension.ID, Name: "ext_prop_2", DisplayName: "Extension Property 2", DataType: "string", Description: "Extremely boring and lame extension property 2"}

				// Create Property 1 for Extension
				createdProperty1, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property1.SchemaExtensionId, property1.Name, property1.DisplayName, property1.DataType, property1.Description)
				require.NoError(t, err, "unexpected error occurred when creating property for extension")

				// Create Property 2 for Extension
				createdProperty2, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property2.SchemaExtensionId, property2.Name, property2.DisplayName, property2.DataType, property2.Description)
				require.NoError(t, err, "unexpected error occurred when creating property for extension")

				// Update Property 2 with name conflict
				createdProperty2.Name = createdProperty1.Name

				// Attempt to Update Property w/ name conflict
				_, err = testSuite.BHDatabase.UpdateGraphSchemaProperty(testSuite.Context, createdProperty2)
				assert.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionPropertyName)
			},
		},
		{
			name: "Error: failed to update property that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				// Update Property
				_, err := testSuite.BHDatabase.UpdateGraphSchemaProperty(testSuite.Context, model.GraphSchemaProperty{
					Name:              "does not exist",
					SchemaExtensionId: extension.ID,
				})
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// DeleteGraphSchemaProperty
		{
			name: "Success: property deleted",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}

				// Create Property for Extension
				newProperty, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "unexpected error occurred when creating property for extension")

				err = testSuite.BHDatabase.DeleteGraphSchemaProperty(testSuite.Context, newProperty.ID)
				assert.NoError(t, err, "unexpected error occurred when deleting property for extension")

				// Validate Property no longer exists
				_, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, newProperty.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete property that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Delete Property
				err := testSuite.BHDatabase.DeleteGraphSchemaProperty(testSuite.Context, int32(10000))
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite, testCase.args)
		})
	}
}

func TestDatabase_GraphSchemaRelationshipKind_CRUD(t *testing.T) {
	assertContainsRelationshipKinds := func(t *testing.T, got model.GraphSchemaRelationshipKinds, expected ...model.GraphSchemaRelationshipKind) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, rk := range got {
				if rk.Name == want.Name &&
					rk.Description == want.Description &&
					rk.IsTraversable == want.IsTraversable &&
					rk.SchemaExtensionId == want.SchemaExtensionId {

					// Additional validations for the found item
					assert.Greater(t, rk.ID, int32(0), "RelationshipKind %v - ID is invalid", rk.Name)

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected relationship kind %v not found", want.Name)
		}
	}

	assertContainsRelationshipKind := func(t *testing.T, got model.GraphSchemaRelationshipKind, expected ...model.GraphSchemaRelationshipKind) {
		t.Helper()
		assertContainsRelationshipKinds(t, []model.GraphSchemaRelationshipKind{got}, expected...)
	}

	assertDoesNotContainRelationshipKinds := func(t *testing.T, got []model.GraphSchemaRelationshipKind, expected ...model.GraphSchemaRelationshipKind) {
		t.Helper()
		for _, want := range expected {
			for _, rk := range got {
				if rk.Name == want.Name &&
					rk.Description == want.Description &&
					rk.IsTraversable == want.IsTraversable &&
					rk.SchemaExtensionId == want.SchemaExtensionId {

					assert.Failf(t, "Unexpected relationship kind found", "Relationship kind %v should not be present", want.Name)
				}
			}
		}
	}

	type args struct {
		filters     model.Filters
		sort        model.Sort
		skip, limit int
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, testSuite IntegrationTestSuite, args args)
	}{
		// CreateGraphSchemaRelationshipKind
		{
			name: "Success: create a schema relationship kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				edgeKind := model.GraphSchemaRelationshipKind{
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				// Create Relationship Kind
				createdEdgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				assert.NoError(t, err)

				assertContainsRelationshipKind(t, createdEdgeKind, edgeKind)
			},
		},
		{
			name: "Error: fails to create duplicate relationship kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				edgeKind := model.GraphSchemaRelationshipKind{
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				createdEdgeKind := createTestRelationshipKind(t, testSuite, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				assertContainsRelationshipKind(t, createdEdgeKind, edgeKind)

				// Create same Relationship Kind 1gain
				_, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaRelationshipKindName)
			},
		},
		// GetGraphSchemaRelationshipKinds
		{
			name: "Error: parseFiltersAndPagination",
			args: args{
				filters: model.Filters{"`": []model.Filter{{}}},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				_, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "invalid operator specified")
			},
		},
		{
			name: "Success: get relationship kinds, no filter or sorting",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting baseline count of relationship kinds")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge kind",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				createdEdgeKind1 := createTestRelationshipKind(t, testSuite, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)

				// Create Relationship Kind 2
				createdEdgeKind2 := createTestRelationshipKind(t, testSuite, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				// Assert only on newly created relationship kinds
				assert.Equal(t, 2, total-baselineCount, "expected 2 new relationship kinds")

				// Validate they exist and are as expected
				assertContainsRelationshipKinds(t, relationshipKinds, createdEdgeKind1, createdEdgeKind2)
			},
		},
		{
			name: "Success: returns relationship kinds, with filtering",
			args: args{
				filters: model.Filters{
					"name": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "test_edge_kind_2",
							SetOperator: model.FilterOr,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "test edge kind", false)
				createdEdgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension.ID, "test edge kind", true)

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				// Assert filtered relationship kinds
				assert.Equal(t, 1, total, "expected 1 filtered relationship kinds")

				// Validate they exist and are as expected
				assertContainsRelationshipKinds(t, relationshipKinds, createdEdgeKind2)
			},
		},
		{
			name: "Success: returns relationship kinds, with fuzzy filtering",
			args: args{
				filters: model.Filters{"description": []model.Filter{
					{
						Operator:    model.ApproximatelyEquals,
						Value:       "test edge",
						SetOperator: model.FilterAnd,
					},
				},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createdEdgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "random", false)
				createdEdgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension.ID, "test edge kind", true)

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error retrieving relationship kinds")

				// Assert filtered relationship kinds
				assert.Equal(t, 1, total, "expected 1 filtered relationship kinds")

				// Validate they exist and are as expected
				assertContainsRelationshipKinds(t, relationshipKinds, createdEdgeKind2)
				assertDoesNotContainRelationshipKinds(t, relationshipKinds, createdEdgeKind1)
			},
		},
		{
			name: "Success: returns relationship kinds, with fuzzy filtering and sort ascending on description",
			args: args{
				filters: model.Filters{
					"description": []model.Filter{
						{
							Operator:    model.ApproximatelyEquals,
							Value:       "test edge",
							SetOperator: model.FilterAnd,
						},
					},
				},
				sort: model.Sort{
					{
						Direction: model.AscendingSortDirection,
						Column:    "description",
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createdEdgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "test edge beta", false)
				createdEdgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension.ID, "test edge alpha", true)

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				// Assert filtered relationship kinds
				assert.Equal(t, 2, total, "expected 2 filtered relationship kinds")

				// Validate they exist
				assertContainsRelationshipKinds(t, relationshipKinds, createdEdgeKind1, createdEdgeKind2)

				// Assert relationship kinds retrieved are sorted in ascending order by description
				assert.Equal(t, relationshipKinds[0].Description, "test edge alpha", "expected description to be in ascending order")
				assert.Equal(t, relationshipKinds[1].Description, "test edge beta", "expected description to be in ascending order")
			},
		},
		{
			name: "Success: returns relationship kinds, with fuzzy filtering and sort descending on description",
			args: args{
				filters: model.Filters{
					"description": []model.Filter{
						{
							Operator:    model.ApproximatelyEquals,
							Value:       "test edge",
							SetOperator: model.FilterAnd,
						},
					},
				},
				sort: model.Sort{
					{
						Direction: model.DescendingSortDirection,
						Column:    "description",
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createdEdgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "test edge beta", false)
				createdEdgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension.ID, "test edge alpha", true)

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				// Assert filtered relationship kinds
				assert.Equal(t, 2, total, "expected 2 filtered relationship kinds")

				// Validate they exist
				assertContainsRelationshipKinds(t, relationshipKinds, createdEdgeKind1, createdEdgeKind2)

				// Assert relationship kinds retrieved are sorted in descending order by description
				assert.Equal(t, relationshipKinds[0].Description, "test edge beta", "expected description to be in descending order")
				assert.Equal(t, relationshipKinds[1].Description, "test edge alpha", "expected description to be in descending order")
			},
		},
		{
			name: "Success: returns relationship kinds, no filter or sorting, with skip",
			args: args{skip: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting baseline count")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "test edge beta", false)
				createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension.ID, "test edge alpha", true)

				// Get Relationship Kinds
				_, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error retrieving relationship kinds")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 relationship kinds")
			},
		},
		{
			name: "Success: returns relationship kinds, no filter or sorting, with limit",
			args: args{limit: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error getting baseline count")

				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "test edge beta", false)
				createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension.ID, "test edge alpha", true)

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error retrieving relationship kinds")

				// Assert total records returned includes the number of records pre-inserted + the number of records created in this test
				assert.Equal(t, baselineCount+2, total, "expected all relationship kinds returned")

				// Assert 1 record matching limit
				assert.Len(t, relationshipKinds, 1, "expected 1 relationship kind returned due to limit")
			},
		},
		{
			name: "Error: returns an error with bogus filtering",
			args: args{
				filters: model.Filters{
					"nonexistentcolumn": []model.Filter{
						{
							Operator: model.Equals,
							Value:    "david",
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Get Relationship Kinds
				relationshipKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no relationship kinds are returned
				assert.Len(t, relationshipKinds, 0, "expected 0 relationship kinds returned due on error")
			},
		},
		// GetGraphSchemaRelationshipKindById
		{
			name: "Success: get relationship kind by id",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				edgeKind := model.GraphSchemaRelationshipKind{
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				createdEdgeKind := createTestRelationshipKind(t, testSuite, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)

				// Validate we retrieved relationship kind by id
				retrievedEdgeKind, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, createdEdgeKind.ID)
				assert.NoError(t, err, "unexpected error getting relationship kind by id")

				assertContainsRelationshipKind(t, retrievedEdgeKind, createdEdgeKind)
			},
		},
		{
			name: "Error: fail to retrieve a relationship kind that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, 5868986)
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// UpdateGraphSchemaRelationshipKind
		{
			name: "Success: update relationship kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createdEdgeKind := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "test edge kind", false)

				updateEdgeKind := model.GraphSchemaRelationshipKind{
					Serial: model.Serial{
						ID: createdEdgeKind.ID,
					},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1", // name must remain unique
					Description:       "updated test edge kind",
					IsTraversable:     false,
				}

				// Update Relationship Kind
				updatedKind, err := testSuite.BHDatabase.UpdateGraphSchemaRelationshipKind(testSuite.Context, updateEdgeKind)
				assert.NoError(t, err, "unexpected error occurred when updating relationship kind")

				// Retrieve Relationship Kind
				edgeKindWithChanges, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, updatedKind.ID)
				assert.NoError(t, err, "unexpected error occurred retrieving updated relationship kind")

				// Validate updated fields
				assertContainsRelationshipKind(t, updateEdgeKind, edgeKindWithChanges)
			},
		},
		{
			name: "Error: failed to update relationship kind that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				// Update Relationship Kind
				_, err := testSuite.BHDatabase.UpdateGraphSchemaRelationshipKind(testSuite.Context, model.GraphSchemaRelationshipKind{
					Name:              "does not exist",
					SchemaExtensionId: extension.ID,
				})
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// DeleteGraphSchemaRelationshipKind
		{
			name: "Success: delete relationship kind",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")

				createdEdgeKind := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension.ID, "test edge kind", false)

				// Delete Relationship Kind
				err := testSuite.BHDatabase.DeleteGraphSchemaRelationshipKind(testSuite.Context, createdEdgeKind.ID)
				assert.NoError(t, err, "unexpected error occurred when deleting relationship kind")

				// Attempt to retrieve deleted Relationship Kind
				_, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, createdEdgeKind.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Error: failed to delete relationship kind that does not exist",
			args: args{},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Delete Relationship Kind
				err := testSuite.BHDatabase.DeleteGraphSchemaRelationshipKind(testSuite.Context, int32(38758765))
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite, testCase.args)
		})
	}
}

func TestDatabase_GetGraphSchemaRelationshipKindsWithSchemaName(t *testing.T) {
	assertContainsRelationshipKinds := func(t *testing.T, got model.GraphSchemaRelationshipKindsWithNamedSchema, expected ...model.GraphSchemaRelationshipKindWithNamedSchema) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, rk := range got {
				if rk.Name == want.Name &&
					rk.Description == want.Description &&
					rk.IsTraversable == want.IsTraversable &&
					rk.SchemaName == want.SchemaName {

					// Additional validations for the found item
					assert.Greater(t, rk.ID, int32(0), "RelationshipKind %v - ID is invalid", rk.Name)

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected relationship kind %v not found", want.Name)
		}
	}

	assertDoesNotContainRelationshipKinds := func(t *testing.T, got []model.GraphSchemaRelationshipKindWithNamedSchema, expected ...model.GraphSchemaRelationshipKindWithNamedSchema) {
		t.Helper()
		for _, want := range expected {
			for _, rk := range got {
				if rk.Name == want.Name &&
					rk.Description == want.Description &&
					rk.IsTraversable == want.IsTraversable &&
					rk.SchemaName == want.SchemaName {

					assert.Failf(t, "Unexpected relationship kind found", "Relationship kind %v should not be present", want.Name)
				}
			}
		}
	}

	type args struct {
		filters     model.Filters
		sort        model.Sort
		skip, limit int
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, testSuite IntegrationTestSuite, args args)
	}{
		{
			name: "Success: get a schema edge kind with named schema, filter for schema name",
			args: args{
				filters: model.Filters{
					"schema.name": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_1", // Extension 1 Name
							SetOperator: model.FilterOr,
						},
					},
				},
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension1.ID, "test edge kind 2", true)
				edgeKind3 := createTestRelationshipKind(t, testSuite, "test_edge_kind_3", extension2.ID, "test edge kind 3", false)
				edgeKind4 := createTestRelationshipKind(t, testSuite, "test_edge_kind_4", extension2.ID, "test edge kind 4", true)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind4.Name,
					Description:   edgeKind4.Description,
					IsTraversable: edgeKind4.IsTraversable,
				}

				// Validate edge kinds are as expected
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assertContainsRelationshipKinds(t, edgeKinds, want1, want2)
				assertDoesNotContainRelationshipKinds(t, edgeKinds, want3, want4)
			},
		},
		{
			name: "Success: get a schema edge kind with named schema, filter for multiple schema names",
			args: args{
				filters: model.Filters{
					"schema.name": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_1",
							SetOperator: model.FilterOr,
						},
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_2",
							SetOperator: model.FilterOr,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension1.ID, "test edge kind 2", true)
				edgeKind3 := createTestRelationshipKind(t, testSuite, "test_edge_kind_3", extension2.ID, "test edge kind 3", false)
				edgeKind4 := createTestRelationshipKind(t, testSuite, "test_edge_kind_4", extension2.ID, "test edge kind 4", true)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind4.Name,
					Description:   edgeKind4.Description,
					IsTraversable: edgeKind4.IsTraversable,
				}

				// Validate edge kinds are as expected
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assertContainsRelationshipKinds(t, edgeKinds, want1, want2, want3, want4)
			},
		},
		{
			name: "Success: get a schema edge kind with named schema, filter for fuzzy match schema names",
			args: args{
				filters: model.Filters{
					"schema.name": []model.Filter{
						{
							Operator:    model.ApproximatelyEquals,
							Value:       "test", // should match extension 1, but not extension 2
							SetOperator: model.FilterOr,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "different-name-should-not-match", "test_extension_2", "1.0.0", "Test2")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension2.ID, "test edge kind 2", false)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{ID: edgeKind1.ID, SchemaName: extension1.Name, Name: edgeKind1.Name, Description: edgeKind1.Description, IsTraversable: edgeKind1.IsTraversable}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{ID: edgeKind2.ID, SchemaName: extension2.Name, Name: edgeKind2.Name, Description: edgeKind2.Description, IsTraversable: edgeKind2.IsTraversable}

				// Validate edge kinds are as expected
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assertContainsRelationshipKinds(t, edgeKinds, want1)
				assertDoesNotContainRelationshipKinds(t, edgeKinds, want2)
			},
		},
		{
			name: "Success: get a schema edge kind with named schema, filter for is_traversable",
			args: args{
				filters: model.Filters{
					"is_traversable": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "true", // should match edge kind 2 & 4
							SetOperator: model.FilterAnd,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension1.ID, "test edge kind 2", true)
				edgeKind3 := createTestRelationshipKind(t, testSuite, "test_edge_kind_3", extension2.ID, "test edge kind 3", false)
				edgeKind4 := createTestRelationshipKind(t, testSuite, "test_edge_kind_4", extension2.ID, "test edge kind 4", true)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind4.Name,
					Description:   edgeKind4.Description,
					IsTraversable: edgeKind4.IsTraversable,
				}

				// Validate edge kinds are as expected
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assertContainsRelationshipKinds(t, edgeKinds, want2, want4)
				assertDoesNotContainRelationshipKinds(t, edgeKinds, want1, want3)
			},
		},
		{
			name: "Success: get a schema edge kind with named schema, filter for schema name and is_traversable",
			args: args{
				filters: model.Filters{
					"schema.name": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_1", // should match edge kind 1 & 2
							SetOperator: model.FilterAnd,
						},
					},
					"is_traversable": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "true", // should match edge kind 2
							SetOperator: model.FilterAnd,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension1.ID, "test edge kind 2", true)
				edgeKind3 := createTestRelationshipKind(t, testSuite, "test_edge_kind_3", extension2.ID, "test edge kind 3", false)
				edgeKind4 := createTestRelationshipKind(t, testSuite, "test_edge_kind_4", extension2.ID, "test edge kind 4", true)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind4.Name,
					Description:   edgeKind4.Description,
					IsTraversable: edgeKind4.IsTraversable,
				}

				// Validate edge kinds are as expected
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assertContainsRelationshipKinds(t, edgeKinds, want2)
				assertDoesNotContainRelationshipKinds(t, edgeKinds, want1, want3, want4)
			},
		},
		{
			name: "Success: get a schema edge kind with named schema, filter for not equals schema name",
			args: args{
				filters: model.Filters{
					"schema.name": []model.Filter{
						{
							Operator:    model.NotEquals,
							Value:       "test_extension_schema_1", // should not return extension A's edge kinds
							SetOperator: model.FilterOr,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension1.ID, "test edge kind 2", true)
				edgeKind3 := createTestRelationshipKind(t, testSuite, "test_edge_kind_3", extension2.ID, "test edge kind 3", false)
				edgeKind4 := createTestRelationshipKind(t, testSuite, "test_edge_kind_4", extension2.ID, "test edge kind 4", true)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind4.Name,
					Description:   edgeKind4.Description,
					IsTraversable: edgeKind4.IsTraversable,
				}

				// Validate edge kinds are as expected
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assertContainsRelationshipKinds(t, edgeKinds, want3, want4)
				assertDoesNotContainRelationshipKinds(t, edgeKinds, want1, want2)
			},
		},
		{
			name: "Success: get a schema edge kind, filter for is not traversable",
			args: args{
				filters: model.Filters{
					"is_traversable": []model.Filter{
						{
							Operator:    model.NotEquals,
							Value:       "true", // should return edge kinds 1 and 3
							SetOperator: model.FilterAnd,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension1.ID, "test edge kind 2", true)
				edgeKind3 := createTestRelationshipKind(t, testSuite, "test_edge_kind_3", extension2.ID, "test edge kind 3", false)
				edgeKind4 := createTestRelationshipKind(t, testSuite, "test_edge_kind_4", extension2.ID, "test edge kind 4", true)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind4.Name,
					Description:   edgeKind4.Description,
					IsTraversable: edgeKind4.IsTraversable,
				}

				// Validate edge kinds are as expected
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assertContainsRelationshipKinds(t, edgeKinds, want1, want3)
				assertDoesNotContainRelationshipKinds(t, edgeKinds, want2, want4)
			},
		},
		{
			name: "Success: get a schema edge kind with named schema, sort by edge name descending",
			args: args{
				filters: model.Filters{
					// Filtering by extensions created in this test to not account for preliminary db data
					"schema.name": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_1", // Extension 1 Name
							SetOperator: model.FilterOr,
						},
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_2", // Extension 2 Name
							SetOperator: model.FilterOr,
						},
					},
				},
				sort: model.Sort{
					model.SortItem{
						Column:    "name",
						Direction: model.DescendingSortDirection,
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension2.ID, "test edge kind 2", false)

				// Assert edge kinds retrieved are sorted in descending order by name
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assert.Equal(t, edgeKinds[0].Name, "test_edge_kind_2", "expected name to be in descending order")
				assert.Equal(t, edgeKinds[1].Name, "test_edge_kind_1", "expected name to be in descending order")
			},
		},
		{
			name: "Success: get a schema edge kind with named schema using skip, no filtering or sorting",
			args: args{skip: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")
				extension2 := createTestExtension(t, testSuite, "test_extension_schema_2", "test_extension_2", "1.0.0", "Test2")

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, 0, 0)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				edgeKind1 := createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				edgeKind2 := createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension2.ID, "test edge kind 2", false)

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extension1.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}

				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extension2.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				// Assert total is as expected
				edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				// We expect both will be found because data already exists in the database
				assert.Equal(t, 2, total-baselineCount, "expected 2 edge kinds")
				assertContainsRelationshipKinds(t, edgeKinds, want1, want2)
			},
		},
		{
			name: "Success: get a schema edge kind with named schema using limit, no filter or sorting",
			args: args{limit: 1},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				extension1 := createTestExtension(t, testSuite, "test_extension_schema_1", "test_extension_1", "1.0.0", "Test")

				createTestRelationshipKind(t, testSuite, "test_edge_kind_1", extension1.ID, "test edge kind 1", false)
				createTestRelationshipKind(t, testSuite, "test_edge_kind_2", extension1.ID, "test edge kind 2", false)

				// Assert only 1 edge kind is returned
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				// Assert 1 record matching limit
				assert.Len(t, edgeKinds, 1, "expected 1 relationship kind returned due to limit")
			},
		},
		{
			name: "Error: failed to build sql query",
			args: args{
				filters: model.Filters{
					"is_traversable": []model.Filter{
						{
							Operator:    "invalid",
							Value:       "true",
							SetOperator: model.FilterAnd,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "invalid operator specified")
			},
		},
		{
			name: "Error: error building sql sort",
			args: args{
				filters: model.Filters{},
				sort: model.Sort{
					model.SortItem{
						Column:    "name",
						Direction: model.InvalidSortDirection,
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "invalid sort direction")
			},
		},
		{
			name: "Error: failed to filter non-existent column",
			args: args{
				filters: model.Filters{
					"invalid": []model.Filter{
						{
							Operator:    "invalid",
							Value:       "true",
							SetOperator: model.FilterAnd,
						},
					},
				},
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				_, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "invalid operator specified")
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite, testCase.args)
		})
	}
}

// Graph Schema Environments may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_Environments_CRUD(t *testing.T) {
	// Helper functions to assert on environments
	assertContainsEnvironments := func(t *testing.T, got []model.SchemaEnvironment, expected ...model.SchemaEnvironment) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, env := range got {
				if env.SchemaExtensionId == want.SchemaExtensionId &&
					env.EnvironmentKindId == want.EnvironmentKindId &&
					env.SourceKindId == want.SourceKindId {

					// Additional validations for the found item
					assert.GreaterOrEqualf(t, env.ID, int32(1), "Environment - ID is invalid")
					assert.Falsef(t, env.CreatedAt.IsZero(), "Environment - created_at is zero")
					assert.Falsef(t, env.UpdatedAt.IsZero(), "Environment - updated_at is zero")
					assert.Falsef(t, env.DeletedAt.Valid, "Environment - deleted_at should be null")

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected environment (extension_id=%v, kind_id=%v, source_kind_id=%v) not found",
				want.SchemaExtensionId, want.EnvironmentKindId, want.SourceKindId)
		}
	}

	assertContainsEnvironment := func(t *testing.T, got model.SchemaEnvironment, expected ...model.SchemaEnvironment) {
		t.Helper()
		assertContainsEnvironments(t, []model.SchemaEnvironment{got}, expected...)
	}

	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		// CreateEnvironment
		{
			name: "Success: create an environment",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
				nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				envKind := getKindByName(t, testSuite, nodeKind.Name)
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: envKind.ID,
					SourceKindId:      int32(sourceKind.ID),
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				// Validate created environment is as expected
				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, newEnvironment.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving environment by id")

				assertContainsEnvironment(t, retrievedEnvironment, environment)
			},
		},
		{
			name: "Error: fails to create duplicate environment",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
				nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				envKind := getKindByName(t, testSuite, nodeKind.Name)
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: envKind.ID,
					SourceKindId:      int32(sourceKind.ID),
				}

				// Create new environment
				got, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

				assertContainsEnvironment(t, got, environment)

				// Create same Environment 1gain and assert error
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaEnvironment)
			},
		},
		// GetEnvironmentsFiltered
		{
			name: "Error: invalid filter",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				environments, err := testSuite.BHDatabase.GetEnvironmentsFiltered(testSuite.Context, model.Filters{
					"invalid_filter": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       "id",
							SetOperator: model.FilterAnd,
						},
					},
				})
				assert.EqualError(t, err, "ERROR: column \"invalid_filter\" does not exist (SQLSTATE 42703)")

				// Assert no environments are returned
				assert.Len(t, environments, 0, "expected 0 environments returned due on error")
			},
		},
		{
			name: "Success: filter environments by extension id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test_A")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test_B")
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
				nodeKind1 := createTestNodeKind(t, testSuite, "nodeKind1", extension1.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				nodeKind2 := createTestNodeKind(t, testSuite, "nodeKind2", extension2.ID, "Node Kind 2", "Test description", false, "fa-test", "#000000")
				environmentKind1 := getKindByName(t, testSuite, nodeKind1.Name)
				environmentKind2 := getKindByName(t, testSuite, nodeKind2.Name)

				environment1 := model.SchemaEnvironment{SchemaExtensionId: extension1.ID, EnvironmentKindId: environmentKind1.ID, SourceKindId: int32(sourceKind.ID)}
				environment2 := model.SchemaEnvironment{SchemaExtensionId: extension2.ID, EnvironmentKindId: environmentKind2.ID, SourceKindId: int32(sourceKind.ID)}

				_, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment1.SchemaExtensionId, environment1.EnvironmentKindId, environment1.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating Environment 1")
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment2.SchemaExtensionId, environment2.EnvironmentKindId, environment2.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating Environment 2")

				retrievedEnvironments, err := testSuite.BHDatabase.GetEnvironmentsFiltered(testSuite.Context, model.Filters{
					"schema_extension_id": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       fmt.Sprint(extension1.ID),
							SetOperator: model.FilterAnd,
						},
					},
				})
				assert.NoError(t, err)

				assertContainsEnvironments(t, retrievedEnvironments, environment1)
			},
		},
		{
			name: "Success: filter environments by extension display name",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension1 := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test_A")
				extension2 := createTestExtension(t, testSuite, "test_extension_2", "test_extension_2", "1.0.0", "Test_B")
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
				nodeKind1 := createTestNodeKind(t, testSuite, "nodeKind1", extension1.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				nodeKind2 := createTestNodeKind(t, testSuite, "nodeKind2", extension2.ID, "Node Kind B", "Test description", false, "fa-test", "#000000")
				environmentKind1 := getKindByName(t, testSuite, nodeKind1.Name)
				environmentKind2 := getKindByName(t, testSuite, nodeKind2.Name)

				environment1 := model.SchemaEnvironment{SchemaExtensionId: extension1.ID, EnvironmentKindId: environmentKind1.ID, SourceKindId: int32(sourceKind.ID)}
				environment2 := model.SchemaEnvironment{SchemaExtensionId: extension2.ID, EnvironmentKindId: environmentKind2.ID, SourceKindId: int32(sourceKind.ID)}

				_, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment1.SchemaExtensionId, environment1.EnvironmentKindId, environment1.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating Environment 1")
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment2.SchemaExtensionId, environment2.EnvironmentKindId, environment2.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating Environment 2")

				retrievedEnvironments, err := testSuite.BHDatabase.GetEnvironmentsFiltered(testSuite.Context, model.Filters{
					"display_name": []model.Filter{
						{
							Operator:    model.Equals,
							Value:       fmt.Sprint(extension2.DisplayName),
							SetOperator: model.FilterAnd,
						},
					},
				})
				assert.NoError(t, err)
				assertContainsEnvironments(t, retrievedEnvironments, environment2)
			},
		},
		{
			name: "Success: get environment by environment kind id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
				nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				envKind := getKindByName(t, testSuite, nodeKind.Name)

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: envKind.ID,
					SourceKindId:      int32(sourceKind.ID),
				}

				newEnvironment := createTestEnvironment(t, testSuite, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)

				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentByEnvironmentKindId(testSuite.Context, newEnvironment.EnvironmentKindId)
				assert.NoError(t, err)

				assertContainsEnvironment(t, retrievedEnvironment, environment)
			},
		},
		{
			name: "Error: fail to get environment by unknown kinds",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				_, err := testSuite.BHDatabase.GetEnvironmentByEnvironmentKindId(testSuite.Context, 20586)
				assert.EqualError(t, err, database.ErrNotFound.Error(), "expected entity not found")
			},
		},
		// GetEnvironmentById
		{
			name: "Success: get environment by id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
				nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				envKind := getKindByName(t, testSuite, nodeKind.Name)
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: envKind.ID,
					SourceKindId:      int32(sourceKind.ID),
				}

				newEnvironment := createTestEnvironment(t, testSuite, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)

				// Validate environment
				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, newEnvironment.ID)
				assert.NoError(t, err, "failed to get environment by id")

				assertContainsEnvironment(t, retrievedEnvironment, environment)
			},
		},
		{
			name: "Error: fail to retrieve environment by id that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				_, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetEnvironments
		{
			name: "Success: return environments",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				baselineEnvironments, err := testSuite.BHDatabase.GetEnvironments(testSuite.Context)
				require.NoError(t, err, "unexpected error occurred when retrieving environments for baseline count")

				extension := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				nodeKind1 := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				nodeKind2 := createTestNodeKind(t, testSuite, "nodeKind2", extension.ID, "Node Kind B", "Test description", false, "fa-test", "#000000")
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
				environmentKind1 := getKindByName(t, testSuite, nodeKind1.Name)
				environmentKind2 := getKindByName(t, testSuite, nodeKind2.Name)

				environment1 := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: environmentKind1.ID,
					SourceKindId:      int32(sourceKind.ID),
				}
				environment2 := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: environmentKind2.ID,
					SourceKindId:      int32(sourceKind.ID),
				}

				// Create Environment 1
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment1.SchemaExtensionId, environment1.EnvironmentKindId, environment1.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment 1")

				// Create Environment 2
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment2.SchemaExtensionId, environment2.EnvironmentKindId, environment2.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment 2")

				// Get Environments back
				environments, err := testSuite.BHDatabase.GetEnvironments(testSuite.Context)
				assert.NoError(t, err, "unexpected error occurred when retrieving environments by extension id")

				// Validate number of results is 2 environments created in this test +
				// the baseline count of environments (number of environments that existed
				// prior to creating environments in this test).
				assert.Len(t, environments, len(baselineEnvironments)+2, "unexpected error occurred while calculating number of environments returned")

				// Validate all created environments exist in the results
				assertContainsEnvironments(t, environments, environment1, environment2)
			},
		},
		// GetEnvironmentsByExtensionId
		{
			name: "Success: return environments by extension id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension := createTestExtension(t, testSuite, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				nodeKind1 := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				nodeKind2 := createTestNodeKind(t, testSuite, "nodeKind2", extension.ID, "Node Kind B", "Test description", false, "fa-test", "#000000")
				sourceKindA := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
				sourceKindB := registerAndGetSourceKind(t, testSuite, "Source_Kind_2")
				environmentKind1 := getKindByName(t, testSuite, nodeKind1.Name)
				environmentKind2 := getKindByName(t, testSuite, nodeKind2.Name)

				environment1 := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: environmentKind1.ID,
					SourceKindId:      int32(sourceKindA.ID),
				}
				environment2 := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: environmentKind2.ID,
					SourceKindId:      int32(sourceKindB.ID),
				}

				// Create Environment 1
				_, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment1.SchemaExtensionId, environment1.EnvironmentKindId, environment1.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment 1")

				// Create Environment 2
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment2.SchemaExtensionId, environment2.EnvironmentKindId, environment2.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment 2")

				// Get Environments back
				environments, err := testSuite.BHDatabase.GetEnvironmentsByExtensionId(testSuite.Context, extension.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving environments by extension id")

				// Validate number of results is 2 environments created in this test
				// The reason that we don't need to total the baseline number of environments
				// is because we are retrieving environments based on the extension ID created
				// in this test.
				assert.Len(t, environments, 2, "expected total environments on the extension to be 2")

				// Validate all created environments exist in the results
				assertContainsEnvironments(t, environments, environment1, environment2)
			},
		},
		{
			name: "Success: environment deleted",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
				nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
				envKind := getKindByName(t, testSuite, nodeKind.Name)
				sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: envKind.ID,
					SourceKindId:      int32(sourceKind.ID),
				}

				// Create Environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")
				assertContainsEnvironment(t, newEnvironment, environment)

				// Delete Environment
				err = testSuite.BHDatabase.DeleteEnvironment(testSuite.Context, newEnvironment.ID)
				assert.NoError(t, err, "unexpected error occurred when deleting environment for extension")

				// Validate environment no longer exists
				_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, newEnvironment.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete environment that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				// Delete Environment
				err := testSuite.BHDatabase.DeleteEnvironment(testSuite.Context, int32(10000))
				assert.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite)
		})
	}
}

// Graph Schema Findings may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_Findings_CRUD(t *testing.T) {
	// Helper functions to assert on Findings
	assertContainsFindings := func(t *testing.T, got []model.SchemaFinding, expected ...model.SchemaFinding) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, finding := range got {
				if finding.SchemaExtensionId == want.SchemaExtensionId &&
					finding.Type == want.Type &&
					finding.KindId == want.KindId &&
					finding.EnvironmentId == want.EnvironmentId &&
					finding.Name == want.Name &&
					finding.DisplayName == want.DisplayName {

					// Additional validations for the found item
					assert.GreaterOrEqualf(t, finding.ID, int32(1), "Finding %v - ID is invalid", finding.Name)
					assert.Falsef(t, finding.CreatedAt.IsZero(), "Finding %v - created_at is zero", finding.Name)

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected finding %v not found", want.Name)
		}
	}

	assertContainsFinding := func(t *testing.T, got model.SchemaFinding, expected ...model.SchemaFinding) {
		t.Helper()
		assertContainsFindings(t, []model.SchemaFinding{got}, expected...)
	}

	setupFindingDeps := func(t *testing.T, testSuite IntegrationTestSuite) (model.GraphSchemaExtension, model.SchemaEnvironment) {
		t.Helper()
		extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
		nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
		envKind := getKindByName(t, testSuite, nodeKind.Name)
		sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
		environment := createTestEnvironment(t, testSuite, extension.ID, envKind.ID, int32(sourceKind.ID))
		return extension, environment
	}

	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		// CreateSchemaFinding
		{
			name: "Success: create a finding",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {

				extension, environment := setupFindingDeps(t, testSuite)

				finding := model.SchemaFinding{
					SchemaExtensionId: extension.ID,
					KindId:            1,
					Type:              model.SchemaFindingTypeRelationship,
					EnvironmentId:     environment.ID,
					Name:              "finding",
					DisplayName:       "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaFinding(testSuite.Context, model.SchemaFindingTypeRelationship, finding.SchemaExtensionId, finding.KindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "unexpected error occurred when creating finding")

				// Validate created finding is as expected
				retrievedFinding, err := testSuite.BHDatabase.GetSchemaFindingById(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving finding")

				assertContainsFinding(t, retrievedFinding, finding)
			},
		},
		{
			name: "Error: fails to create duplicate finding",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension, environment := setupFindingDeps(t, testSuite)

				finding := model.SchemaFinding{
					SchemaExtensionId: extension.ID,
					KindId:            1,
					EnvironmentId:     environment.ID,
					Name:              "finding",
					DisplayName:       "display name",
				}

				newFinding := createTestFinding(t, testSuite, finding)

				// Create same finding again and assert error
				_, err := testSuite.BHDatabase.CreateSchemaFinding(testSuite.Context, model.SchemaFindingTypeRelationship, newFinding.SchemaExtensionId, newFinding.KindId, newFinding.EnvironmentId, newFinding.Name, newFinding.DisplayName)
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaFindingName)
			},
		},
		// GetSchemaFindingById
		{
			name: "Success: get finding by id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension, environment := setupFindingDeps(t, testSuite)

				finding := model.SchemaFinding{
					SchemaExtensionId: extension.ID,
					Type:              model.SchemaFindingTypeRelationship,
					KindId:            1,
					EnvironmentId:     environment.ID,
					Name:              "finding",
					DisplayName:       "display name",
				}

				newFinding := createTestFinding(t, testSuite, finding)

				// Validate finding is as expected
				retrievedFinding, err := testSuite.BHDatabase.GetSchemaFindingById(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving finding by id")

				assertContainsFinding(t, retrievedFinding, finding)
			},
		},
		{
			name: "Error: fail to retrieve finding by id that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {

				_, err := testSuite.BHDatabase.GetSchemaFindingById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetSchemaFindingByName
		{
			name: "Success: get finding by name",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension, environment := setupFindingDeps(t, testSuite)

				finding := model.SchemaFinding{
					SchemaExtensionId: extension.ID,
					Type:              model.SchemaFindingTypeRelationship,
					KindId:            1,
					EnvironmentId:     environment.ID,
					Name:              "finding",
					DisplayName:       "display name",
				}

				newFinding := createTestFinding(t, testSuite, finding)

				// Validate finding is as expected
				retrievedFinding, err := testSuite.BHDatabase.GetSchemaFindingByName(testSuite.Context, newFinding.Name)
				assert.NoError(t, err, "unexpected error occurred when retrieving finding by name")

				assertContainsFinding(t, retrievedFinding, finding)
			},
		},
		{
			name: "Error: fail to retrieve finding by name that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				_, err := testSuite.BHDatabase.GetSchemaFindingByName(testSuite.Context, "doesnotexist")
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// DeleteSchemaFinding
		{
			name: "Success: finding deleted",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				extension, environment := setupFindingDeps(t, testSuite)

				finding := model.SchemaFinding{
					SchemaExtensionId: extension.ID,
					Type:              model.SchemaFindingTypeRelationship,
					KindId:            1,
					EnvironmentId:     environment.ID,
					Name:              "finding",
					DisplayName:       "display name",
				}

				newFinding := createTestFinding(t, testSuite, finding)
				assertContainsFinding(t, newFinding, finding)

				// Delete Finding
				err := testSuite.BHDatabase.DeleteSchemaFinding(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when deleting finding")

				// Validate finding no longer exists
				_, err = testSuite.BHDatabase.GetSchemaFindingById(testSuite.Context, newFinding.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete finding that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				// Delete Finding
				err := testSuite.BHDatabase.DeleteSchemaFinding(testSuite.Context, int32(10000))
				assert.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// GetSchemaFindingsBySchemaExtensionId
		{
			name: "Success: no findings found",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				// Get Findings
				findings, err := testSuite.BHDatabase.GetSchemaFindingsBySchemaExtensionId(testSuite.Context, int32(10000))
				assert.NoError(t, err, "unexpected error occurred when retrieving findings by schema extension id")

				assert.Len(t, findings, 0, "findings were expected to be 0 when extension id was not found")
			},
		},
		{
			name: "Success: retrieve multiple findings by schema extension id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				createdExtension := createTestExtension(t, testSuite, "TestGraphSchemaExtension", "Test Graph Schema Extension", "1.0.0", "TGSE")
				createdEnvironmentNode := createTestNodeKind(t, testSuite, "TGSE_Environment 1", createdExtension.ID, "Environment 1", "an environment kind", false, "", "")
				createdSourceKindNode := createTestNodeKind(t, testSuite, "Source_Kind_1", createdExtension.ID, "Source Kind 1", "a source kind", false, "", "")
				envKind := getKindByName(t, testSuite, createdEnvironmentNode.Name)
				sourceKind := registerAndGetSourceKind(t, testSuite, createdSourceKindNode.Name)
				createdEnvironment := createTestEnvironment(t, testSuite, createdExtension.ID, envKind.ID, int32(sourceKind.ID))

				createdEdgeKind := createTestRelationshipKind(t, testSuite, "TGSE_Edge_1", createdExtension.ID, "an edge kind", true)
				edgeKind := getKindByName(t, testSuite, createdEdgeKind.Name)

				// Assign Extension ID, Edge Kind, & Environment ID to Finding 1
				finding1 := model.SchemaFinding{
					SchemaExtensionId: createdExtension.ID,
					KindId:            edgeKind.ID,
					EnvironmentId:     createdEnvironment.ID,
					Name:              "Finding_1",
					DisplayName:       "Finding 1",
				}

				// Assign Extension ID, Edge Kind, & Environment ID to Finding 2
				finding2 := model.SchemaFinding{
					SchemaExtensionId: createdExtension.ID,
					KindId:            edgeKind.ID,
					EnvironmentId:     createdEnvironment.ID,
					Name:              "Finding_2",
					DisplayName:       "Finding 2",
				}

				createTestFinding(t, testSuite, finding1)
				createTestFinding(t, testSuite, finding2)

				// Get Findings by Extension ID
				findings, err := testSuite.BHDatabase.GetSchemaFindingsBySchemaExtensionId(testSuite.Context, createdExtension.ID)
				assert.NoError(t, err, "unexpected error occurred when getting findings by extension id")

				// Validate both findings exist on extension
				assertContainsFindings(t, findings, finding1, finding2)
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite)
		})
	}
}

// Graph Schema Remediations may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_Remediations_CRUD(t *testing.T) {
	// Helper functions to assert on Findings
	assertContainsRemediations := func(t *testing.T, got []model.Remediation, expected ...model.Remediation) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, rem := range got {
				if rem.FindingID == want.FindingID &&
					rem.ShortDescription == want.ShortDescription &&
					rem.LongDescription == want.LongDescription &&
					rem.ShortRemediation == want.ShortRemediation &&
					rem.LongRemediation == want.LongRemediation {

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected remediation for finding_id %v not found", want.FindingID)
		}
	}

	assertContainsRemediation := func(t *testing.T, got model.Remediation, expected ...model.Remediation) {
		t.Helper()
		assertContainsRemediations(t, []model.Remediation{got}, expected...)
	}

	setupRemediationDeps := func(t *testing.T, testSuite IntegrationTestSuite) model.SchemaFinding {
		t.Helper()
		extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
		nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
		envKind := getKindByName(t, testSuite, nodeKind.Name)
		sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
		environment := createTestEnvironment(t, testSuite, extension.ID, envKind.ID, int32(sourceKind.ID))
		return createTestFinding(t, testSuite, model.SchemaFinding{
			SchemaExtensionId: extension.ID,
			KindId:            nodeKind.ID,
			Type:              model.SchemaFindingTypeRelationship,
			EnvironmentId:     environment.ID,
			Name:              "finding",
			DisplayName:       "display name",
		})
	}

	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		// CreateRemediation
		{
			name: "Success: create a remediation",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newFinding := setupRemediationDeps(t, testSuite)

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				assert.NoError(t, err, "unexpected error occurred when creating remediation")

				// Validate created remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, remediation)
			},
		},
		{
			name: "Error: fails to create duplicate remediation",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newFinding := setupRemediationDeps(t, testSuite)

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "unexpected error occurred when creating remediation")

				// Create same remediation again
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				// Assert error
				assert.EqualError(t, err, "ERROR: duplicate key value violates unique constraint \"schema_remediations_pkey\" (SQLSTATE 23505)")
			},
		},
		// GetRemediationByFindingId
		{
			name: "Success: get remediation by finding id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newFinding := setupRemediationDeps(t, testSuite)

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "unexpected error occurred when creating remediation")

				// Validate created remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, remediation)
			},
		},
		{
			name: "Error: fail to retrieve remediation by id that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				_, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetRemediationByFindingName
		{
			name: "Success: get remediation by finding name",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newFinding := setupRemediationDeps(t, testSuite)

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "unexpected error occurred when creating remediation")

				// Validate created remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingName(testSuite.Context, newFinding.Name)
				assert.NoError(t, err, "unexpected error occurred when retrieving remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, remediation)
			},
		},
		{
			name: "Error: fail to retrieve remediation by finding name that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				_, err := testSuite.BHDatabase.GetRemediationByFindingName(testSuite.Context, "namedoesnotexist")
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// UpdateRemediation
		{
			name: "Success: remediation updated",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newFinding := setupRemediationDeps(t, testSuite)

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				createdRemediation, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "unexpected error occurred when creating remediation")

				updatedRemediation := model.Remediation{
					FindingID:        createdRemediation.FindingID,
					ShortDescription: "Updated short desc",
					LongDescription:  "Updated long desc",
					ShortRemediation: "Updated short fix",
					LongRemediation:  "Updated long fix",
				}

				_, err = testSuite.BHDatabase.UpdateRemediation(testSuite.Context, updatedRemediation.FindingID, updatedRemediation.ShortDescription, updatedRemediation.LongDescription, updatedRemediation.ShortRemediation, updatedRemediation.LongRemediation)
				assert.NoError(t, err, "failed to updated remediation")

				// Validate updated remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, updatedRemediation.FindingID)
				assert.NoError(t, err, "unexpected error occurred when retrieving updated remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, updatedRemediation)
			},
		},
		{
			name: "Error: failed to update remediation that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				remediation := model.Remediation{
					FindingID:        1498659768, // finding id that does not existg
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Update Remediation
				_, err := testSuite.BHDatabase.UpdateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				assert.EqualError(t, err, "ERROR: insert or update on table \"schema_remediations\" violates foreign key constraint \"schema_remediations_finding_id_fkey\" (SQLSTATE 23503)")
			},
		},
		// DeleteRemediation
		{
			name: "Success: remediation deleted",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newFinding := setupRemediationDeps(t, testSuite)

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				createdRemediation, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "unexpected error occurred when creating remediation")

				assertContainsRemediation(t, createdRemediation, remediation)

				// Delete Remediation
				err = testSuite.BHDatabase.DeleteRemediation(testSuite.Context, remediation.FindingID)
				assert.NoError(t, err, "unexpected error occurred when deleting remediation by finding id")

				// Validate remediation no longer exists
				_, err = testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, remediation.FindingID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete remediation that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				// Delete Remediation
				err := testSuite.BHDatabase.DeleteRemediation(testSuite.Context, int32(10000))
				assert.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite)
		})
	}
}

// Graph Schema Principal Kinds may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_PrincipalKinds_CRUD(t *testing.T) {
	// Helper functions to assert on principal kinds
	assertContainsPrincipalKinds := func(t *testing.T, got []model.SchemaEnvironmentPrincipalKind, expected ...model.SchemaEnvironmentPrincipalKind) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, pk := range got {
				if pk.EnvironmentId == want.EnvironmentId &&
					pk.PrincipalKind == want.PrincipalKind {

					// Additional validations for the found item
					assert.Falsef(t, pk.CreatedAt.IsZero(), "PrincipalKind (env_id=%v, kind=%v) - created_at is zero",
						pk.EnvironmentId, pk.PrincipalKind)

					found = true
					break
				}
			}
			assert.Truef(t, found, "expected principal kind (env_id=%v, kind=%v) not found",
				want.EnvironmentId, want.PrincipalKind)
		}
	}

	setupPrincipalKindDeps := func(t *testing.T, testSuite IntegrationTestSuite) model.SchemaEnvironment {
		t.Helper()
		extension := createTestExtension(t, testSuite, "test_extension", "test_extension", "1.0.0", "Test")
		nodeKind := createTestNodeKind(t, testSuite, "nodeKind1", extension.ID, "Node Kind 1", "Test description", false, "fa-test", "#000000")
		envKind := getKindByName(t, testSuite, nodeKind.Name)
		sourceKind := registerAndGetSourceKind(t, testSuite, "Source_Kind_1")
		return createTestEnvironment(t, testSuite, extension.ID, envKind.ID, int32(sourceKind.ID))
	}

	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		// CreatePrincipalKind
		{
			name: "Success: create principal kind",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newEnvironment := setupPrincipalKindDeps(t, testSuite)

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}
				// Create new principal kind
				newPrincipalKind, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				assert.NoError(t, err, "unexpected error occurred when creating principal kind")

				// Validate created principalKind is as expected
				retrievedPrincipalKind, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				assert.NoError(t, err, "unexpected error occurred when retrieving principal kind by environment id")

				assertContainsPrincipalKinds(t, retrievedPrincipalKind, principalKind)
			},
		},
		{
			name: "Error: fails to create duplicate principal kind",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newEnvironment := setupPrincipalKindDeps(t, testSuite)

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}

				// Create new principal kind
				_, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				require.NoError(t, err, "unexpected error occurred when creating principal kind")

				// Create same principal kind again
				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				// Assert error
				assert.ErrorIs(t, err, model.ErrDuplicatePrincipalKind)
			},
		},
		// GetPrincipalKindsByEnvironmentId
		{
			name: "Success: get principal kinds by environment id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newEnvironment := setupPrincipalKindDeps(t, testSuite)

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}

				// Create new principal kind
				newPrincipalKind, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				require.NoError(t, err, "unexpected error occurred when creating principal kind")

				// Validate we are able to retrieve principal kinds back by environment id
				retrievedPrincipalKind, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				assert.NoError(t, err, "unexpected error occurred when retrieving principal kind by environment id")

				assertContainsPrincipalKinds(t, retrievedPrincipalKind, principalKind)
			},
		},
		{
			name: "Success: principal kinds should return empty if none are found",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				principalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, int32(5000))
				assert.NoError(t, err)
				assert.Len(t, principalKinds, 0)
			},
		},
		// DeletePrincipalKind
		{
			name: "Success: principal kind deleted",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				newEnvironment := setupPrincipalKindDeps(t, testSuite)

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}
				// Create new principal kind
				newPrincipalKind, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				require.NoError(t, err, "unexpected error occurred when creating principal kind")

				// Validate created principalKind is as expected
				_, err = testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				require.NoError(t, err, "unexpected error occurred when retrieving principal kind by environment id")

				// Delete Principal Kind
				err = testSuite.BHDatabase.DeletePrincipalKind(testSuite.Context, newPrincipalKind.EnvironmentId, newPrincipalKind.PrincipalKind)
				assert.NoError(t, err, "unexpected error occurred when deleting principal kind")

				// Validate principal kind no longer exists
				// Principal kinds returns empty slice when not found instead of error
				foundPrincipalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				assert.NoError(t, err)
				assert.Len(t, foundPrincipalKinds, 0)
			},
		},
		{
			name: "Error: failed to delete principal kind that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				// Delete Principal Kind
				err := testSuite.BHDatabase.DeletePrincipalKind(testSuite.Context, int32(10000), int32(18858))
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			// Run test assertions
			testCase.assert(t, testSuite)
		})
	}
}

func TestDeleteSchemaExtension_CascadeDeletesAllDependents(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	extension := createTestExtension(t, testSuite, "CascadeTestExtension", "Cascade Test Extension", "v1.0.0", "CTE")
	nodeKind := createTestNodeKind(t, testSuite, "CascadeTestNodeKind", extension.ID, "Cascade Test Node Kind", "Test description", false, "fa-test", "#000000")
	dawgsEnvKind := getKindByName(t, testSuite, "CascadeTestNodeKind")
	sourceKind := registerAndGetSourceKind(t, testSuite, "CascadeTestSourceKind")
	property, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension.ID, "cascade_test_property", "Cascade Test Property", "string", "Test description")
	require.NoError(t, err, "unexpected error occurred when creating property")

	edgeKind := createTestRelationshipKind(t, testSuite, "CascadeTestEdgeKind", extension.ID, "Test description", true)
	environment := createTestEnvironment(t, testSuite, extension.ID, dawgsEnvKind.ID, int32(sourceKind.ID))
	relationshipFinding := createTestFinding(t, testSuite, model.SchemaFinding{
		SchemaExtensionId: extension.ID,
		Type:              model.SchemaFindingTypeRelationship,
		KindId:            edgeKind.ID,
		EnvironmentId:     environment.ID,
		Name:              "CascadeTestFinding",
		DisplayName:       "Cascade Test Finding",
	})

	_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, relationshipFinding.ID, "Short desc", "Long desc", "Short remediation", "Long remediation")
	require.NoError(t, err, "unexpected error occurred when creating remediation")

	_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, environment.ID, dawgsEnvKind.ID)
	require.NoError(t, err, "unexpected error occurred when creating principal kind")

	// Delete the extension - should cascade to all dependents
	err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extension.ID)
	require.NoError(t, err, "unexpected error occurred when deleting extension")

	_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, nodeKind.ID)
	assert.ErrorIs(t, err, database.ErrNotFound, "node kind should have been cascade deleted")

	_, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, property.ID)
	assert.ErrorIs(t, err, database.ErrNotFound, "property should have been cascade deleted")

	_, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, edgeKind.ID)
	assert.ErrorIs(t, err, database.ErrNotFound, "relationship kind should have been cascade deleted")

	_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, environment.ID)
	assert.ErrorIs(t, err, database.ErrNotFound, "environment should have been cascade deleted")

	_, err = testSuite.BHDatabase.GetSchemaFindingById(testSuite.Context, relationshipFinding.ID)
	assert.ErrorIs(t, err, database.ErrNotFound, "finding should have been cascade deleted")

	_, err = testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, relationshipFinding.ID)
	assert.ErrorIs(t, err, database.ErrNotFound, "remediation should have been cascade deleted")

	principalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, environment.ID)
	assert.NoError(t, err)
	assert.Len(t, principalKinds, 0, "principal kinds should have been cascade deleted")

	// Validate Source Kind has been de-activated
	_, err = testSuite.BHDatabase.GetSourceKindByID(testSuite.Context, sourceKind.ID)
	assert.ErrorIs(t, err, database.ErrNotFound, "source kind should have been deactivated")
}
