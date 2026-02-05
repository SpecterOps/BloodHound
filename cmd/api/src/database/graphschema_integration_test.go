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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Graph Schema Extensions may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_GraphSchemaExtensions_CRUD(t *testing.T) {
	var (
		ext1 = model.GraphSchemaExtension{
			Name:        "adam",
			DisplayName: "test extension name 1",
			Version:     "1.0.0",
			Namespace:   "Test",
		}
		ext2 = model.GraphSchemaExtension{
			Name:        "bob",
			DisplayName: "test extension name 2",
			Version:     "2.0.0",
			Namespace:   "Test2",
		}
		ext3 = model.GraphSchemaExtension{
			Name:        "charlie",
			DisplayName: "another extension",
			Version:     "3.0.0",
			Namespace:   "AA",
		}
		ext4 = model.GraphSchemaExtension{
			Name:        "david",
			DisplayName: "yet another extension",
			Version:     "4.0.0",
			Namespace:   "ZZ",
		}
	)

	// Helper function to create all test extensions
	createTestExtensions := func(testSuite IntegrationTestSuite) {
		for _, ext := range []model.GraphSchemaExtension{ext1, ext2, ext3, ext4} {
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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				assert.NoError(t, err, "unexpected error occurred when creating extension")
			},
		},
		{
			name: "Error: fail to create duplicate schema extension name",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create test extensions
				createTestExtensions(testSuite)

				// Insert graph extension that already exists
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				assert.EqualError(t, err, "duplicate graph schema extension name: adam")
			},
		},
		{
			name: "Error: fail to create duplicate schema extension namespace",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create test extensions
				createTestExtensions(testSuite)

				// Insert graph extension that already exists
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "different", ext1.DisplayName, ext1.Version, ext1.Namespace)
				assert.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionNamespace)
			},
		},
		// GetGraphSchemaExtensionById
		{
			name: "Success: retrieves graph extension by id",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "unexpected error occurred when creating extension")

				// Retrieve created extension by ID
				extension, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				assert.NoError(t, err, "unexpected error occurred when getting extension by id")

				// Assert extension has been created
				assert.Equal(t, extension.Name, ext1.Name)
			},
		},
		// GetGraphSchemaExtensions
		{
			name: "Success: returns extensions, no filter or sorting",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 4 new records were created in this test
				assert.Equal(t, 4, total-baselineCount, "expected 4 new extensions")

				// Validate all created extensions exist in the results
				assert.True(t, assertContainsExtension(extensions, ext1), "ext1 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext2), "ext2 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext3), "ext3 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext4), "ext4 should exist in results")
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 1 matching record
				assert.Equal(t, 1, total-baselineCount, "expected 1 extension matching the filter")

				// Should only contain ext4 (david)
				assert.True(t, assertContainsExtension(extensions, ext4), "ext4 should exist in results")

				// Should not contain the others
				assert.False(t, assertContainsExtension(extensions, ext1), "ext1 should not be in filtered results")
				assert.False(t, assertContainsExtension(extensions, ext2), "ext2 should not be in filtered results")
				assert.False(t, assertContainsExtension(extensions, ext3), "ext3 should not be in filtered results")
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 1 matching record
				assert.Equal(t, 1, total-baselineCount, "expected 1 extension matching both filters")

				// Should contain only ext4 (matches both filters)
				assert.True(t, assertContainsExtension(extensions, ext4), "ext4 should exist in results")
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")

				// Should contain ext1 & ext2 (matches fuzzy filter)
				assert.True(t, assertContainsExtension(extensions, ext1), "ext1 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext2), "ext2 should exist in results")
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
				sort:  model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved are sorted in ascending order by display name
				assert.Equal(t, extensions[0].DisplayName, "test extension name 1", "expected display name to be in ascending order")
				assert.Equal(t, extensions[1].DisplayName, "test extension name 2", "expected display name to be in ascending order")
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
				sort:  model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (ext1 & ext2) are sorted in ascending order by display name
				assert.Equal(t, extensions[0].DisplayName, "test extension name 1", "expected display name to be in ascending order")
				assert.Equal(t, extensions[1].DisplayName, "test extension name 2", "expected display name to be in ascending order")
			},
		},
		{
			name: "Success: returns extensions, with fuzzy filtering and sort descending",
			args: args{
				filters: model.Filters{
					"display_name": []model.Filter{
						{
							Operator: model.ApproximatelyEquals,
							Value:    "test extension",
						},
					},
				},
				sort:  model.Sort{{Column: "display_name", Direction: model.DescendingSortDirection}},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "unexpected error occurred when retrieving graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (ext1 & ext2) are sorted in descending order by display name
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
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

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
				t.Helper()

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
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
				sort:  model.Sort{},
				skip:  0,
				limit: 1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				// Create test extensions
				createTestExtensions(testSuite)

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "unexpected error occurred when creating first extension")

				// Create a second extension to test a duplicated namespace
				secondExtension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "name", "display", "v1.0.0", "duplicate")
				require.NoError(t, err, "unexpected error occurred when creating second extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "unexpected error occurred when creating first extension")

				// Create a second extension to test a duplicated namespace
				secondExtension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "duplicate", "display", "v1.0.0", "random")
				require.NoError(t, err, "unexpected error occurred when creating second extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Update in database
				_, err := testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, model.GraphSchemaExtension{Serial: model.Serial{ID: int32(5000)}})
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// DeleteGraphSchemaExtension
		{
			name: "Success: extension deleted",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "unexpected error occurred when creating extension")

				// Delete extension
				err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extension.ID)
				require.NoError(t, err, "unexpected error occurred when deleting extension")

				// Validate it's no longer there
				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Error: failed to delete extension that does not exist",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Delete extension
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, int32(5000))
				assert.ErrorIs(t, err, database.ErrNotFound)
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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				nodeKind := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_2",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_2",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
				assert.NoError(t, err, "unexpected error occurred when creating node kind")

				// Create same node again
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaNodeKindName)
			},
		},
		// GetGraphSchemaNodeKindById
		{
			name: "Success: get schema node kind by id",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				nodeKind := model.GraphSchemaNodeKind{
					Name:              "Test_Kind_1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "A test kind",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				createdNodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind")

				_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, createdNodeKind.ID)
				assert.NoError(t, err, "unexpected error occurred getting node kind by id")

			},
		},
		{
			name: "Error: fail to retrieve a node kind that does not exist",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, 112)
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetGraphSchemaNodeKinds
		{
			name: "Success: return node schema kinds, no filter or sorting",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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

				// Insert Node Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 2")

				// Get Node Kinds back
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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

				// Insert Node Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 2")

				// Get Node Kinds back
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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

				// Insert Node Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 3")

				// Get Node Kinds back
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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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

				// Insert Node Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 3")

				// Get Node Kinds back
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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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

				// Insert Node Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 3")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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

				// Insert Node Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 3")

				_, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving node kinds")

				// Assert 3 matching records
				assert.Equal(t, 3, total-baselineCount, "expected 3 node kinds")
			},
		},
		{
			name: "Success: returns schema node kinds, no filter or sorting, with limit",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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

				// Insert Node Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 3")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				nodeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no nodeKinds are returned
				assert.Len(t, nodeKinds, 0, "expected 0 node kinds returned due on error")
			},
		},
		// UpdateGraphSchemaNodeKind
		{
			name: "Success: update schema node kind",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "Alpha",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				// Insert Node Kind 1
				createdNodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind")

				updatedNodeKind1 := model.GraphSchemaNodeKind{
					Serial: model.Serial{
						ID: createdNodeKind.ID,
					},
					Name:              "Test Kind 1", // name should not be updated
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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				// Update Node Kind 1
				_, err = testSuite.BHDatabase.UpdateGraphSchemaNodeKind(testSuite.Context, model.GraphSchemaNodeKind{
					Name:              "does not exist",
					SchemaExtensionId: extension.ID,
				})
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// DeleteGraphSchemaNodeKind
		{
			name: "Success: deleted node kind",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				nodeKind1 := model.GraphSchemaNodeKind{
					Name:              "Test Kind 1",
					SchemaExtensionId: extension.ID,
					DisplayName:       "Test_Kind_1",
					Description:       "Beta",
					IsDisplayKind:     false,
					Icon:              "test_icon",
					IconColor:         "blue",
				}

				// Insert Node Kind 1
				insertedNodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name, nodeKind1.SchemaExtensionId, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind, nodeKind1.Icon, nodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Delete Node Kind 1
				err = testSuite.BHDatabase.DeleteGraphSchemaNodeKind(testSuite.Context, insertedNodeKind.ID)
				assert.NoError(t, err, "unexpected error occureed when deleting node kind 1")

				// Validate Node Kind no longer exists
				_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, insertedNodeKind.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete schema node kind that does not exist",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Insert Node Kind 1
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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				// Create property
				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.NoError(t, err, "unexpected error occurred when creating property")

				assertContainsProperty(t, got, property)

				// Create same property again
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.ErrorIs(t, err, model.ErrDuplicateGraphSchemaExtensionPropertyName)
			},
		},
		{
			name: "Success: creates same property (without name collision) on multiple extensions",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				// Create property
				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.NoError(t, err, "unexpected error occurred when creating property")

				// Validate expected property was created
				assertContainsProperty(t, got, property)

				// Create new schema extension
				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "2.0.0", "Test_2")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetGraphSchemaProperties
		{
			name: "Success: return properties, no filter or sorting",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting baseline count")

				// Create Extensions
				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating another extension")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating another extension")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating another extension")

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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating another extension")

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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating another extension")

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
				assert.Equal(t, 2, total, "expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 2, "expected 2 properties to be returned when fuzzy filtering by description")

				// Assert extensions retrieved (extension1Property1 & extension2Property2) are sorted in descending order by description
				assert.Equal(t, properties[0].Description, "Extremely boring and lame extension property Beta", "expected description to be in ascending order")
				assert.Equal(t, properties[1].Description, "Extremely boring and lame extension property Alpha", "expected description to be in ascending order")
			},
		},
		{
			name: "Success: returns properties, no filter or sorting, with skip",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting baseline count")

				// Create Extensions
				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating another extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting baseline count")

				// Create Extensions
				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating another extension")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
					Serial: model.Serial{
						ID: newProperty.ID,
					},
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
			name: "Error: failed to update property that does not exist",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				// Update Property
				_, err = testSuite.BHDatabase.UpdateGraphSchemaProperty(testSuite.Context, model.GraphSchemaProperty{
					Name:              "does not exist",
					SchemaExtensionId: extension.ID,
				})
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// DeleteGraphSchemaProperty
		{
			name: "Success: property deleted",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				// Create Relationship Kind
				createdEdgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind")

				assertContainsRelationshipKind(t, createdEdgeKind, edgeKind)

				// Create same Relationship Kind again
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaRelationshipKindName)
			},
		},
		// GetGraphSchemaRelationshipKinds
		{
			name: "Success: get relationship kinds, no filter or sorting",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting baseline count of relationship kinds")

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge kind",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				createdEdgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				// Create Relationship Kind 2
				createdEdgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge kind",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "Error creating relationship kind 1")

				// Create Relationship Kind 2
				createdEdgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err, "Error creating relationship kind 2")

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "expected to retrieve relationship kinds")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "random",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge kind",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				createdEdgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 1")

				// Create Relationship Kind 2
				createdEdgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 2")

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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge beta",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge alpha",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				createdEdgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 1")

				// Create Relationship Kind 2
				createdEdgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 2")

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error retrieving relationship kinds")

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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge beta",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge alpha",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				createdEdgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 1")

				// Create Relationship Kind 2
				createdEdgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 2")

				// Get Relationship Kinds
				relationshipKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error retrieving relationship kinds")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting baseline count")

				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge beta",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge alpha",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 1")

				// Create Relationship Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 2")

				// Get Relationship Kinds
				_, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error retrieving relationship kinds")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "expected 2 relationship kinds")
			},
		},
		{
			name: "Success: returns relationship kinds, no filter or sorting, with limit",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error getting baseline count")

				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind1 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge beta",
					IsTraversable:     false,
				}

				edgeKind2 := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_2",
					Description:       "test edge alpha",
					IsTraversable:     true,
				}

				// Create Relationship Kind 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 1")

				// Create Relationship Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err, "unexpected error creating relationship kind 2")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Relationship Kinds back
				relationshipKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no relationship kinds are returned
				assert.Len(t, relationshipKinds, 0, "expected 0 relationship kinds returned due on error")
			},
		},
		// GetGraphSchemaRelationshipKindById
		{
			name: "Success: get relationship kind by id",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind := model.GraphSchemaRelationshipKind{
					Serial:            model.Serial{},
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				// Create Relationship Kind 1
				createdEdgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				require.NoError(t, err, "unexpected error when creating relationship kind")

				// Validate we retrieved relationship kind by id
				retrievedEdgeKind, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, createdEdgeKind.ID)
				assert.NoError(t, err, "unexpected error getting relationship kind by id")

				assertContainsRelationshipKind(t, createdEdgeKind, retrievedEdgeKind)
			},
		},
		{
			name: "Error: fail to retrieve a relationship kind that does not exist",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, 5868986)
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// UpdateGraphSchemaRelationshipKind
		{
			name: "Success: update relationship kind",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind := model.GraphSchemaRelationshipKind{
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				// Create Relationship Kind
				createdEdgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				require.NoError(t, err, "unexpected error when creating relationship kind")

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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				// Update Relationship Kind
				_, err = testSuite.BHDatabase.UpdateGraphSchemaRelationshipKind(testSuite.Context, model.GraphSchemaRelationshipKind{
					Name:              "does not exist",
					SchemaExtensionId: extension.ID,
				})
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// DeleteGraphSchemaRelationshipKind
		{
			name: "Success: delete relationship kind",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				edgeKind := model.GraphSchemaRelationshipKind{
					SchemaExtensionId: extension.ID,
					Name:              "test_edge_kind_1",
					Description:       "test edge kind",
					IsTraversable:     false,
				}

				// Create Relationship Kind
				createdEdgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name, edgeKind.SchemaExtensionId, edgeKind.Description, edgeKind.IsTraversable)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind")

				// Delete Relationship Kind
				err = testSuite.BHDatabase.DeleteGraphSchemaRelationshipKind(testSuite.Context, createdEdgeKind.ID)
				assert.NoError(t, err, "unexpected error occurred when deleting relationship kind")

				// Attempt to retrieve deleted Relationship Kind
				_, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, createdEdgeKind.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		{
			name: "Error: failed to update relationship kind that does not exist",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

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
							Value:       "test_extension_schema_a", // Extension A Name
							SetOperator: model.FilterOr,
						},
					},
				},
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				edgeKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_3", extensionB.ID, "test edge kind 3", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 3")

				edgeKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_4", extensionB.ID, "test edge kind 4", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 4")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extensionB.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extensionB.Name,
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
							Value:       "test_extension_schema_a", // Extension A Name
							SetOperator: model.FilterOr,
						},
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_b", // Extension B Name
							SetOperator: model.FilterOr,
						},
					},
				},
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				edgeKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_3", extensionB.ID, "test edge kind 3", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 3")

				edgeKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_4", extensionB.ID, "test edge kind 4", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 4")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extensionB.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extensionB.Name,
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "different-name-should-not-match", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionB.ID, "test edge kind 2", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}

				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionB.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				edgeKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_3", extensionB.ID, "test edge kind 3", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 3")

				edgeKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_4", extensionB.ID, "test edge kind 4", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 4")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extensionB.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extensionB.Name,
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
							Value:       "test_extension_schema_a", // should match edge kind 1 & 2
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				edgeKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_3", extensionB.ID, "test edge kind 3", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 3")

				edgeKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_4", extensionB.ID, "test edge kind 4", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 4")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extensionB.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extensionB.Name,
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
							Value:       "test_extension_schema_a", // should not return extension A's edge kinds
							SetOperator: model.FilterOr,
						},
					},
				},
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				edgeKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_3", extensionB.ID, "test edge kind 3", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 3")

				edgeKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_4", extensionB.ID, "test edge kind 4", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 4")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extensionB.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extensionB.Name,
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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				edgeKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_3", extensionB.ID, "test edge kind 3", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 3")

				edgeKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_4", extensionB.ID, "test edge kind 4", true)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 4")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}
				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind2.Name,
					Description:   edgeKind2.Description,
					IsTraversable: edgeKind2.IsTraversable,
				}

				want3 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind3.ID,
					SchemaName:    extensionB.Name,
					Name:          edgeKind3.Name,
					Description:   edgeKind3.Description,
					IsTraversable: edgeKind3.IsTraversable,
				}
				want4 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind4.ID,
					SchemaName:    extensionB.Name,
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
							Value:       "test_extension_schema_a", // Extension A Name
							SetOperator: model.FilterOr,
						},
						{
							Operator:    model.Equals,
							Value:       "test_extension_schema_b", // Extension B Name
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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Create Relationship Kinds
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionB.ID, "test edge kind 2", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				// Assert edge kinds retrieved are sorted in descending order by name
				edgeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "unexpected error occurred when retrieving relationship kinds")

				assert.Equal(t, edgeKinds[0].Name, "test_edge_kind_2", "expected name to be in descending order")
				assert.Equal(t, edgeKinds[1].Name, "test_edge_kind_1", "expected name to be in descending order")
			},
		},
		{
			name: "Success: get a schema edge kind with named schema using skip, no filtering or sorting",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extensions
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
				require.NoError(t, err, "unexpected error occurred when creating extension B")

				// Get Baseline Count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKindsWithSchemaName(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "unexpected error occurred when getting baseline count")

				// Create Relationship Kinds
				edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionB.ID, "test edge kind 2", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

				want1 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind1.ID,
					SchemaName:    extensionA.Name,
					Name:          edgeKind1.Name,
					Description:   edgeKind1.Description,
					IsTraversable: edgeKind1.IsTraversable,
				}

				want2 := model.GraphSchemaRelationshipKindWithNamedSchema{
					ID:            edgeKind2.ID,
					SchemaName:    extensionB.Name,
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
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create Extension
				extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension A")

				// Create Relationship Kinds
				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 1")

				_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", false)
				require.NoError(t, err, "unexpected error occurred when creating relationship kind 2")

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

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
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

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
				sort:  model.Sort{},
				skip:  0,
				limit: 0,
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, args args) {
				t.Helper()

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
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
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
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				// Create new environment
				got, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

				assertContainsEnvironment(t, got, environment)

				// Create same environment again
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				// Assert error
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaEnvironment)
			},
		},
		// GetEnvironmentByKinds
		{
			name: "Success: get environment by kinds - kind id and source id are unique",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentByKinds(testSuite.Context, newEnvironment.EnvironmentKindId, newEnvironment.SourceKindId)
				assert.NoError(t, err, database.ErrNotFound)

				assertContainsEnvironment(t, retrievedEnvironment, environment)

			},
		},
		{
			name: "Error: fail to get environment by unknown kinds",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				environment := model.SchemaEnvironment{
					EnvironmentKindId: 20586,
					SourceKindId:      257958,
				}

				_, err := testSuite.BHDatabase.GetEnvironmentByKinds(testSuite.Context, environment.EnvironmentKindId, environment.SourceKindId)
				assert.EqualError(t, err, database.ErrNotFound.Error(), "expected entity not found")
			},
		},
		// GetEnvironmentById
		{
			name: "Success: get environment by id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				// Create Environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

				// Validate environment
				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, newEnvironment.ID)
				assert.NoError(t, err, "failed to get environment by id")

				assertContainsEnvironment(t, retrievedEnvironment, environment)

			},
		},
		{
			name: "Error: fail to retrieve environment by id that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetEnvironments
		{
			name: "Success: return environments",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Get Environments - baseline count
				baselineEnvironments, err := testSuite.BHDatabase.GetEnvironments(testSuite.Context)
				assert.NoError(t, err, "unexpected error occurred when retrieving environments for baseline count")

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment1 := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}
				environment2 := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          2,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               2,
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
				assert.Len(t, environments, len(baselineEnvironments)+2, "unexpected error occured while calculating number of environments returned")

				// Validate all created environments exist in the results
				assertContainsEnvironments(t, environments, environment1, environment2)
			},
		},
		// GetEnvironmentsByExtensionId
		{
			name: "Success: return environments by extension id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment1 := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}
				environment2 := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          2,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               2,
				}

				// Create Environment 1
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment1.SchemaExtensionId, environment1.EnvironmentKindId, environment1.SourceKindId)
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
		// DeleteEnvironment
		{
			name: "Success: environment deleted",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				// Create Environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

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
				t.Helper()

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

// Graph Schema Relationship Findings may contain dynamically pre-inserted data, meaning the database
// may already contain existing records. These tests should be written to account for said data.
func TestDatabase_Findings_CRUD(t *testing.T) {
	// Helper functions to assert on Findings
	assertContainsFindings := func(t *testing.T, got []model.SchemaRelationshipFinding, expected ...model.SchemaRelationshipFinding) {
		t.Helper()
		for _, want := range expected {
			found := false
			for _, finding := range got {
				if finding.SchemaExtensionId == want.SchemaExtensionId &&
					finding.RelationshipKindId == want.RelationshipKindId &&
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

	assertContainsFinding := func(t *testing.T, got model.SchemaRelationshipFinding, expected ...model.SchemaRelationshipFinding) {
		t.Helper()
		assertContainsFindings(t, []model.SchemaRelationshipFinding{got}, expected...)
	}

	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		// CreateSchemaRelationshipFinding
		{
			name: "Success: create a finding",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "unexpected error occurred when creating finding")

				// Validate created finding is as expected
				retrievedFinding, err := testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving finding")

				assertContainsFinding(t, retrievedFinding, finding)

			},
		},
		{
			name: "Error: fails to create duplicate finding",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "unexpected error occurred when creating finding")

				// Create same finding again
				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, newFinding.SchemaExtensionId, newFinding.RelationshipKindId, newFinding.EnvironmentId, newFinding.Name, newFinding.DisplayName)
				// Assert error
				assert.ErrorIs(t, err, model.ErrDuplicateSchemaRelationshipFindingName)
			},
		},
		// GetSchemaRelationshipFindingById
		{
			name: "Success: get finding by id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "unexpected error occurred when creating finding")

				// Validate finding is as expected
				retrievedFinding, err := testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when retrieving finding by id")

				assertContainsFinding(t, retrievedFinding, finding)
			},
		},
		{
			name: "Error: fail to retrieve finding by id that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetSchemaRelationshipFindingByName
		{
			name: "Success: get finding by name",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "unexpected error occurred when creating finding")

				// Validate finding is as expected
				retrievedFinding, err := testSuite.BHDatabase.GetSchemaRelationshipFindingByName(testSuite.Context, newFinding.Name)
				assert.NoError(t, err, "unexpected error occurred when retrieving finding by name")

				assertContainsFinding(t, retrievedFinding, finding)
			},
		},
		{
			name: "Error: fail to retrieve finding by name that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetSchemaRelationshipFindingByName(testSuite.Context, "doesnotexist")
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// DeleteSchemaRelationshipFinding
		{
			name: "Success: finding deleted",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "unexpected error occurred when creating finding")

				assertContainsFinding(t, newFinding, finding)

				// Delete Finding
				err = testSuite.BHDatabase.DeleteSchemaRelationshipFinding(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "unexpected error occurred when deleting finding")

				// Validate finding no longer exists
				_, err = testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, newFinding.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete finding that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Delete Finding
				err := testSuite.BHDatabase.DeleteSchemaRelationshipFinding(testSuite.Context, int32(10000))
				assert.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// GetSchemaRelationshipFindingsBySchemaExtensionId
		{
			name: "Success: no findings found",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Get Findings
				findings, err := testSuite.BHDatabase.GetSchemaRelationshipFindingsBySchemaExtensionId(testSuite.Context, int32(10000))
				assert.NoError(t, err, "unexpected error occurred when retrieving findings by schema extension id")

				assert.Len(t, findings, 0, "findings were expected to be 0 when extension id was not found")
			},
		},
		{
			name: "Success: retrieve multiple findings by schema extension id",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				extension := model.GraphSchemaExtension{
					Name:        "TestGraphSchemaExtension",
					DisplayName: "Test Graph Schema Extension",
					Version:     "1.0.0",
					Namespace:   "TGSE",
				}

				// Create extension
				createdExtension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environmentNodeKind1 := model.GraphSchemaNodeKind{
					Name:        "TGSE_Environment 1",
					DisplayName: "Environment 1",
					Description: "an environment kind",
				}

				sourceKind1 := model.GraphSchemaNodeKind{
					Name:        "Source_Kind_1",
					DisplayName: "Source Kind 1",
					Description: "a source kind",
				}
				edgeKind1 := model.GraphSchemaRelationshipKind{
					Name:          "TGSE_Edge_1",
					Description:   "an edge kind",
					IsTraversable: true,
				}

				// Create Environment Node Kind
				createdEnvironmentNode, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind 1")

				// Create Source Kind Node
				createdSourceKindNode, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err, "unexpected error occurred when creating node kind")

				// Retrieve DAWGS Environment Kind
				envKind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode.Name)
				require.NoError(t, err, "unexpected error occurred when getting kind by name")

				// Register Source Kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(sourceKind1.Name))
				require.NoError(t, err, "unexpected error occurred when registering source kind")

				// Retrieve Source Kind
				sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err, "unexpected error occurred when retreiving source kind")

				// Create Environment
				createdEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					envKind.ID, int32(sourceKind.ID))
				require.NoError(t, err, "unexpected error occurred when creating environment")

				// Create Finding Edge Kind
				createdEdgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err, "unexpected error occurred when creating edge kind")

				// Retrieve Finding Edge Kind
				edgeKind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind.Name)
				require.NoError(t, err, "unexpected error occurred when retrieving edge kind")

				// Assign Extension ID, Edge Kind, & Environment ID to Finding 1
				finding1 := model.SchemaRelationshipFinding{
					SchemaExtensionId:  createdExtension.ID,
					RelationshipKindId: edgeKind.ID,
					EnvironmentId:      createdEnvironment.ID,
					Name:               "Finding_1",
					DisplayName:        "Finding 1",
				}

				// Assign Extension ID, Edge Kind, & Environment ID to Finding 2
				finding2 := model.SchemaRelationshipFinding{
					SchemaExtensionId:  createdExtension.ID,
					RelationshipKindId: edgeKind.ID,
					EnvironmentId:      createdEnvironment.ID,
					Name:               "Finding_2",
					DisplayName:        "Finding 2",
				}

				// Create Finding 1
				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, edgeKind.ID, createdEnvironment.ID, finding1.Name, finding1.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding 1")

				// Create Finding 2
				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, edgeKind.ID, createdEnvironment.ID, finding2.Name, finding2.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding 2")

				// Get Findings by Extension ID
				findings, err := testSuite.BHDatabase.GetSchemaRelationshipFindingsBySchemaExtensionId(testSuite.Context, createdExtension.ID)
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

	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		// CreateRemediation
		{
			name: "Success: create a remediation",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
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
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
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
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
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
				t.Helper()

				_, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetRemediationByFindingName
		{
			name: "Success: get remediation by finding name",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				assert.NoError(t, err, "unexpected error occurred when creating remediation")

				// Validate created remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingName(testSuite.Context, newFinding.Name)
				assert.NoError(t, err, "unexpected error occurred when retrieving remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, remediation)
			},
		},
		{
			name: "Error: fail to retrieve remediation by finding name that does not exist",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetRemediationByFindingName(testSuite.Context, "namedoesnotexist")
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// UpdateRemediation
		{
			name: "Success: remediation updated",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding")

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
				t.Helper()

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
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create Environment
				environment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "unexpected error occurred when creating environment")

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      environment.ID,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "unexpected error occurred when creating finding")

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
				t.Helper()

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

	tests := []struct {
		name   string
		assert func(t *testing.T, testSuite IntegrationTestSuite)
	}{
		// CreatePrincipalKind
		{
			name: "Success: create principal kind",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

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
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}
				// Create new principal kind
				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				assert.NoError(t, err, "unexpected error occurred when creating principal kind")

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
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

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
				t.Helper()

				principalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, int32(5000))
				assert.NoError(t, err)
				assert.Len(t, principalKinds, 0)
			},
		},
		// DeletePrincipalKind
		{
			name: "Success: principal kind deleted",
			assert: func(t *testing.T, testSuite IntegrationTestSuite) {
				t.Helper()

				// Create Extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err, "unexpected error occurred when creating extension")

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "unexpected error occurred when creating environment")

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
				t.Helper()

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

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "CascadeTestExtension", "Cascade Test Extension", "v1.0.0", "CTE")
	require.NoError(t, err, "unexpected error occurred when creating extension")

	nodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, "CascadeTestNodeKind", extension.ID, "Cascade Test Node Kind", "Test description", false, "fa-test", "#000000")
	require.NoError(t, err, "unexpected error occurred when creating node kind")

	dawgsEnvKind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, "CascadeTestNodeKind")
	require.NoError(t, err, "unexpected error occurred when getting kind by name")

	err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("CascadeTestSourceKind"))
	require.NoError(t, err, "unexpected error occurred when registering source kind")

	sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, "CascadeTestSourceKind")
	require.NoError(t, err, "unexpected error occurred when getting source kind by name")

	property, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension.ID, "cascade_test_property", "Cascade Test Property", "string", "Test description")
	require.NoError(t, err, "unexpected error occurred when creating property")

	edgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "CascadeTestEdgeKind", extension.ID, "Test description", true)
	require.NoError(t, err, "unexpected error occurred when creating relationship kind")

	environment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, extension.ID, dawgsEnvKind.ID, int32(sourceKind.ID))
	require.NoError(t, err, "unexpected error occurred when creating environment")

	relationshipFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, extension.ID, edgeKind.ID, environment.ID, "CascadeTestFinding", "Cascade Test Finding")
	require.NoError(t, err, "unexpected error occurred when creating finding")

	_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, relationshipFinding.ID, "Short desc", "Long desc", "Short remediation", "Long remediation")
	require.NoError(t, err, "unexpected error occurred when creating remediation")

	_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, environment.ID, dawgsEnvKind.ID)
	require.NoError(t, err, "unexpected error occurred when creating principal kind")

	err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extension.ID)
	require.NoError(t, err, "unexpected error occurred when deleting extension")

	_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, nodeKind.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	_, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, property.ID)
	assert.ErrorIs(t, err, database.ErrNotFound)

	_, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, edgeKind.ID)
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
