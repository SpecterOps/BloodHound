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
			require.NoError(t, err)
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
		assert func(testSuite IntegrationTestSuite, args args)
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				assert.NoError(t, err, "failed to create new extension")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create test extensions
				createTestExtensions(testSuite)

				// Insert graph extension that already exists
				_, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				assert.EqualError(t, err, "duplicate graph schema extension name: ERROR: duplicate key value violates unique constraint \"schema_extensions_name_key\" (SQLSTATE 23505)")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "failed to create new extension")

				// Retrieve created extension by ID
				extension, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				assert.NoError(t, err)

				// Assert extension has been created
				assert.Equal(t, extension.Name, ext1.Name)
			},
		},
		// GetGraphSchemaExtensions
		{
			name: "Success: returns slice of extensions, no filtering or sorting",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   0,
			},
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 4 new records were created in this test
				assert.Equal(t, 4, total-baselineCount, "Expected 4 new extensions")

				// Validate all created extensions exist in the results
				assert.True(t, assertContainsExtension(extensions, ext1), "ext1 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext2), "ext2 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext3), "ext3 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext4), "ext4 should exist in results")
			},
		},
		{
			name: "Success: returns slice of extensions, with filtering",
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 1 matching record
				assert.Equal(t, 1, total-baselineCount, "Expected 1 extension matching the filter")

				// Should only contain ext4 (david)
				assert.True(t, assertContainsExtension(extensions, ext4), "ext4 should exist in results")

				// Should not contain the others
				assert.False(t, assertContainsExtension(extensions, ext1), "ext1 should not be in filtered results")
				assert.False(t, assertContainsExtension(extensions, ext2), "ext2 should not be in filtered results")
				assert.False(t, assertContainsExtension(extensions, ext3), "ext3 should not be in filtered results")
			},
		},
		{
			name: "Success: returns slice of extensions, with multiple filters",
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 1 matching record
				assert.Equal(t, 1, total-baselineCount, "Expected 1 extension matching both filters")

				// Should contain only ext4 (matches both filters)
				assert.True(t, assertContainsExtension(extensions, ext4), "ext4 should exist in results")
			},
		},
		{
			name: "Success: returns slice of extensions, with fuzzy filtering",
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "Expected 2 extensions matching fuzzy filter")

				// Should contain ext1 & ext2 (matches fuzzy filter)
				assert.True(t, assertContainsExtension(extensions, ext1), "ext1 should exist in results")
				assert.True(t, assertContainsExtension(extensions, ext2), "ext2 should exist in results")
			},
		},
		{
			name: "Success: returns slice of extensions, with fuzzy filtering and sort ascending",
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "Expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved are sorted in ascending order by display name
				assert.Equal(t, extensions[0].DisplayName, "test extension name 1", "Expected display name to be in ascending order")
				assert.Equal(t, extensions[1].DisplayName, "test extension name 2", "Expected display name to be in ascending order")
			},
		},
		{
			name: "Success: returns slice of extensions, with fuzzy filtering and sort ascending",
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "Expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (ext1 & ext2) are sorted in ascending order by display name
				assert.Equal(t, extensions[0].DisplayName, "test extension name 1", "Expected display name to be in ascending order")
				assert.Equal(t, extensions[1].DisplayName, "test extension name 2", "Expected display name to be in ascending order")
			},
		},
		{
			name: "Success: returns slice of extensions, with fuzzy filtering and sort descending",
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "Expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (ext1 & ext2) are sorted in descending order by display name
				assert.Equal(t, extensions[0].DisplayName, "test extension name 2", "Expected display name to be in descending order")
				assert.Equal(t, extensions[1].DisplayName, "test extension name 1", "Expected display name to be in descending order")
			},
		},
		{
			name: "Success: returns slice of extensions, no filtering or sorting, with skip",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				_, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert 4 matching records
				assert.Equal(t, 4, total-baselineCount, "Expected 4 extensions")
			},
		},
		{
			name: "Success: returns slice of extensions, no filtering or sorting, with limit",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get baseline count
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				require.NoError(t, err, "baseline count required")

				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, total, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.NoError(t, err, "failed to retrieve graph schema extensions")

				// Assert total records returned includes the number of records pre-inserted + the number of records created in this test
				assert.Equal(t, baselineCount+4, total, "Expected all extension records (6) returned")
				// Assert 1 record matching limit
				assert.Len(t, extensions, 1, "Expected 1 extension returned due to limit")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				// Create test extensions
				createTestExtensions(testSuite)

				// Get extensions after inserting test data
				extensions, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context, args.filters, args.sort, args.skip, args.limit,
				)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no extensions are returned
				assert.Len(t, extensions, 0, "Expected 0 extensions returned due on error")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "failed to create new extension")

				newlyCreatedExtension, err := testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extension.ID)
				require.NoError(t, err, "failed to retrieve newly created extension")

				// Modify some fields (not is_builtin)
				newlyCreatedExtension.Name = "new name"
				newlyCreatedExtension.DisplayName = "new display name"
				newlyCreatedExtension.Version = "v5.0.0"
				newlyCreatedExtension.Namespace = "different namespace"

				// Update in database
				updatedExtension, err := testSuite.BHDatabase.UpdateGraphSchemaExtension(testSuite.Context, newlyCreatedExtension)
				assert.NoError(t, err, "failed to update extension")

				// Validate fields are updated
				assert.Equal(t, newlyCreatedExtension.Name, updatedExtension.Name)
				assert.Equal(t, newlyCreatedExtension.DisplayName, updatedExtension.DisplayName)
				assert.Equal(t, newlyCreatedExtension.Version, updatedExtension.Version)
				assert.Equal(t, newlyCreatedExtension.Namespace, updatedExtension.Namespace)
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
			assert: func(testSuite IntegrationTestSuite, args args) {
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Create new extension
				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, ext1.Name, ext1.DisplayName, ext1.Version, ext1.Namespace)
				require.NoError(t, err, "failed to create new extension")

				// Delete extension
				err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extension.ID)
				require.NoError(t, err, "failed to retrieve newly created extension")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
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
			testCase.assert(testSuite, testCase.args)
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
			assert.Truef(t, found, "Expected node kind %v not found", want.Name)
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
		assert func(testSuite IntegrationTestSuite, args args)
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				assert.NoError(t, err, "failed to create node kind")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				assert.NoError(t, err, "failed to create node kind")

				// Create same node again
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind.Name, nodeKind.SchemaExtensionId, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.IconColor)
				assert.ErrorIs(t, err, database.ErrDuplicateSchemaNodeKindName)
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind")

				_, err = testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, createdNodeKind.ID)
				assert.NoError(t, err, "failed to get node kind by id")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "failed to create node kind 2")

				// Get Node Kinds back
				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve node kinds")

				// Validate number of results
				assert.Equal(t, baselineCount+2, total, "Expected total node kinds to be equal to how many node kinds exist in the database")
				assert.Len(t, nodeKinds, baselineCount+2, "Expected all node kinds to be returned when no filtering/sorting")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "failed to create node kind 2")

				// Get Node Kinds back
				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve node kinds")

				// Validate number of results
				assert.Equal(t, 1, total, "Expected total node kinds to be equal to how many results are returned from database")
				assert.Len(t, nodeKinds, 1, "Expected 1 node kind to be returned when filtering by name")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "failed to create node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "failed to create node kind 3")

				// Get Node Kinds back
				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve node kinds")

				// Validate number of results
				assert.Equal(t, 2, total, "Expected total node kinds to be equal to how many results are returned from database")
				assert.Len(t, nodeKinds, 2, "Expected 2 node kinds to be returned when fuzzy filtering by name")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()
				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "failed to create node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "failed to create node kind 3")

				// Get Node Kinds back
				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve node kinds")

				// Validate number of results
				assert.Equal(t, 2, total-baselineCount, "Expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (nodeKind1 & nodeKind2) are sorted in ascending order by description
				assert.Equal(t, nodeKinds[0].Description, "Alpha", "Expected description to be in ascending order")
				assert.Equal(t, nodeKinds[1].Description, "Beta", "Expected description to be in ascending order")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "failed to create node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "failed to create node kind 3")

				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve node kinds")

				// Assert 2 matching records
				assert.Equal(t, 2, total-baselineCount, "Expected 2 extensions matching fuzzy filter")

				// Assert extensions retrieved (nodeKind1 & nodeKind2) are sorted in ascending order by description
				assert.Equal(t, nodeKinds[0].Description, "Beta", "Expected description to be in descending order")
				assert.Equal(t, nodeKinds[1].Description, "Alpha", "Expected description to be in descending order")
			},
		},
		{
			name: "Success: returns schema node kinds, no filtering or sorting, with skip",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "failed to create node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "failed to create node kind 3")

				_, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve node kinds")

				// Assert 3 matching records
				assert.Equal(t, 3, total-baselineCount, "Expected 3 node kinds")
			},
		},
		{
			name: "Success: returns schema node kinds, no filtering or sorting, with limit",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial graph schema node kinds prior to insert")

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Insert Node Kind 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind2.Name, nodeKind2.SchemaExtensionId, nodeKind2.DisplayName, nodeKind2.Description, nodeKind2.IsDisplayKind, nodeKind2.Icon, nodeKind2.IconColor)
				require.NoError(t, err, "failed to create node kind 2")

				// Insert Node Kind 3
				_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind3.Name, nodeKind3.SchemaExtensionId, nodeKind3.DisplayName, nodeKind3.Description, nodeKind3.IsDisplayKind, nodeKind3.Icon, nodeKind3.IconColor)
				require.NoError(t, err, "failed to create node kind 3")

				nodeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve node kinds")

				// Assert total records returned includes the number of records pre-inserted + the number of records created in this test
				assert.Equal(t, baselineCount+3, total, "Expected all node kind records (6) returned")
				// Assert 1 record matching limit
				assert.Len(t, nodeKinds, 1, "Expected 1 node kind returned due to limit")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				nodeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no nodeKinds are returned
				assert.Len(t, nodeKinds, 0, "Expected 0 node kinds returned due on error")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err)

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
				require.NoError(t, err)

				// Retrieve Node Kind 1
				nodeKindWithChanges, err := testSuite.BHDatabase.GetGraphSchemaNodeKindById(testSuite.Context, updatedNodeKind.ID)
				require.NoError(t, err)

				// Assert on updated fields
				assert.Equal(t, updatedNodeKind.DisplayName, nodeKindWithChanges.DisplayName)
				assert.Equal(t, updatedNodeKind.Description, nodeKindWithChanges.Description)
				assert.Equal(t, updatedNodeKind.IsDisplayKind, nodeKindWithChanges.IsDisplayKind)
				assert.Equal(t, updatedNodeKind.Icon, nodeKindWithChanges.Icon)
				assert.Equal(t, updatedNodeKind.IconColor, nodeKindWithChanges.IconColor)
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create node kind 1")

				// Delete Node Kind 1
				err = testSuite.BHDatabase.DeleteGraphSchemaNodeKind(testSuite.Context, insertedNodeKind.ID)
				assert.NoError(t, err, "failed to delete node kind 1")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
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
			testCase.assert(testSuite, testCase.args)
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
			assert.Truef(t, found, "Expected property %v not found", want.Name)
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
		assert func(testSuite IntegrationTestSuite, args args)
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.NoError(t, err, "failed to create property")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.NoError(t, err, "failed to create property")

				assertContainsProperty(t, got, property)

				// Create same property again
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.ErrorIs(t, err, database.ErrDuplicateGraphSchemaExtensionPropertyName)
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension1.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				got, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				assert.NoError(t, err, "failed to create property")

				// Validate expected property was created
				assertContainsProperty(t, got, property)

				// Create new schema extension
				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "2.0.0", "Test_2")
				require.NoError(t, err)

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely fun and exciting extension property",
				}

				newProperty, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "failed to create property")

				_, err = testSuite.BHDatabase.GetGraphSchemaPropertyById(testSuite.Context, newProperty.ID)
				assert.NoError(t, err, "failed to get property by id")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial properties prior to insert")

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "failed to create property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "failed to create property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve properties")

				// Validate number of results
				assert.Equal(t, baselineCount+3, total, "Expected total properties to be equal to how many properties exist in the database")
				assert.Len(t, properties, baselineCount+3, "Expected all properties to be returned when no filtering/sorting")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "failed to create property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "failed to create property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve properties")

				// Validate number of results
				assert.Equal(t, 1, total, "Expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 1, "Expected 1 property to be returned when filtering by description")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "failed to create property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "failed to create property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve properties")

				// Validate number of results
				assert.Equal(t, 2, total, "Expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 2, "Expected 2 properties to be returned when fuzzy filtering by description")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "failed to create property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "failed to create property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve properties")

				// Validate number of results
				assert.Equal(t, 2, total, "Expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 2, "Expected 2 properties to be returned when fuzzy filtering by description")

				// Assert extensions retrieved (extension1Property1 & extension2Property2) are sorted in ascending order by description
				assert.Equal(t, properties[0].Description, "Extremely boring and lame extension property Alpha", "Expected description to be in ascending order")
				assert.Equal(t, properties[1].Description, "Extremely boring and lame extension property Beta", "Expected description to be in ascending order")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "failed to create property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "failed to create property 1 for extension 2")

				// Get Properties back
				properties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve properties")

				// Validate number of results
				assert.Equal(t, 2, total, "Expected total properties to be equal to how many results are returned from database")
				assert.Len(t, properties, 2, "Expected 2 properties to be returned when fuzzy filtering by description")

				// Assert extensions retrieved (extension1Property1 & extension2Property2) are sorted in descending order by description
				assert.Equal(t, properties[0].Description, "Extremely boring and lame extension property Beta", "Expected description to be in ascending order")
				assert.Equal(t, properties[1].Description, "Extremely boring and lame extension property Alpha", "Expected description to be in ascending order")
			},
		},
		{
			name: "Success: returns properties, no filtering or sorting, with skip",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    1,
				limit:   0,
			},
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial properties prior to insert")

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "failed to create property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "failed to create property 1 for extension 2")

				// Get Properties back
				_, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve properties")

				// Assert 3 matching records
				assert.Equal(t, 3, total-baselineCount, "Expected 3 properties")
			},
		},
		{
			name: "Success: returns schema node kinds, no filtering or sorting, with limit",
			args: args{
				filters: model.Filters{},
				sort:    model.Sort{},
				skip:    0,
				limit:   1,
			},
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				_, baselineCount, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				require.NoError(t, err, "Error getting initial properties prior to insert")

				extension1, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

				extension2, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_2", "test_extension_2", "1.0.0", "Test2")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create property 1 for extension 1")

				// Create Prop 2 for Extension 1
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension1Property2.SchemaExtensionId, extension1Property2.Name, extension1Property2.DisplayName, extension1Property2.DataType, extension1Property2.Description)
				require.NoError(t, err, "failed to create property 2 for extension 1")

				// Create Prop 1 for Extension 2
				_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension2Property2.SchemaExtensionId, extension2Property2.Name, extension2Property2.DisplayName, extension2Property2.DataType, extension2Property2.Description)
				require.NoError(t, err, "failed to create property 1 for extension 2")

				// Get Properties back
				proerties, total, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.NoError(t, err, "failed to retrieve properties")

				// Assert 3 matching records
				assert.Equal(t, 3, total-baselineCount, "Expected 3 properties")

				// Assert total records returned includes the number of records pre-inserted + the number of records created in this test
				assert.Equal(t, baselineCount+3, total, "Expected all properties returned")
				// Assert 1 record matching limit
				assert.Len(t, proerties, 1, "Expected 1 property returned due to limit")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				// Get Properties back
				properties, _, err := testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context, args.filters, args.sort, args.skip, args.limit)
				assert.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")

				// Assert no properties are returned
				assert.Len(t, properties, 0, "Expected 0 properties returned due on error")
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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}

				// Create Property for Extension
				newProperty, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "failed to create property for extension")

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
				require.NoError(t, err)

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

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
			assert: func(testSuite IntegrationTestSuite, args args) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				property := model.GraphSchemaProperty{
					SchemaExtensionId: extension.ID,
					Name:              "ext_prop_1",
					DisplayName:       "Extension Property 1",
					DataType:          "string",
					Description:       "Extremely boring and lame extension property 1",
				}

				// Create Property for Extension
				newProperty, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, property.SchemaExtensionId, property.Name, property.DisplayName, property.DataType, property.Description)
				require.NoError(t, err, "failed to create property for extension")

				err = testSuite.BHDatabase.DeleteGraphSchemaProperty(testSuite.Context, newProperty.ID)
				assert.NoError(t, err, "failed to delete property for extension")

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
			assert: func(testSuite IntegrationTestSuite, args args) {
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
			testCase.assert(testSuite, testCase.args)
		})
	}
}

// func TestDatabase_GraphSchemaEdgeKind_CRUD(t *testing.T) {
// 	t.Parallel()
// 	testSuite := setupIntegrationTestSuite(t)
// 	defer teardownIntegrationTestSuite(t, &testSuite)
// 	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
// 		"test_extension_schema_edge_kinds", "test_extension", "1.0.0", "Test")
// 	require.NoError(t, err)

// 	var (
// 		edgeKind1 = model.GraphSchemaRelationshipKind{
// 			Serial:            model.Serial{},
// 			SchemaExtensionId: extension.ID,
// 			Name:              "test_edge_kind_1",
// 			Description:       "test edge kind",
// 			IsTraversable:     false,
// 		}
// 		edgeKind2 = model.GraphSchemaRelationshipKind{
// 			Serial:            model.Serial{},
// 			SchemaExtensionId: extension.ID,
// 			Name:              "test_edge_kind_2",
// 			Description:       "test edge kind",
// 			IsTraversable:     true,
// 		}
// 		edgeKind3 = model.GraphSchemaRelationshipKind{
// 			Serial:            model.Serial{},
// 			SchemaExtensionId: extension.ID,
// 			Name:              "test_edge_kind_3",
// 			Description:       "test edge kind 3",
// 			IsTraversable:     false,
// 		}
// 		edgeKind4 = model.GraphSchemaRelationshipKind{
// 			Serial:            model.Serial{},
// 			SchemaExtensionId: extension.ID,
// 			Name:              "test_edge_kind_4",
// 			Description:       "test edge kind 4",
// 			IsTraversable:     false,
// 		}
// 		updateWant = model.GraphSchemaRelationshipKind{
// 			Serial:            model.Serial{},
// 			SchemaExtensionId: extension.ID,
// 			Name:              "test_edge_kind_345",
// 			Description:       "test edge kind",
// 			IsTraversable:     false,
// 		}

// 		gotEdgeKind1 = model.GraphSchemaRelationshipKind{}
// 		gotEdgeKind2 = model.GraphSchemaRelationshipKind{}
// 	)

// 	// CREATE

// 	// Expected success - create one model.GraphSchemaRelationshipKind
// 	t.Run("success - create a schema edge kind #1", func(t *testing.T) {
// 		gotEdgeKind1, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name, edgeKind1.SchemaExtensionId, edgeKind1.Description, edgeKind1.IsTraversable)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKind(t, gotEdgeKind1, edgeKind1)
// 	})
// 	// Expected success - create a second model.GraphSchemaRelationshipKind
// 	t.Run("success - create a schema edge kind #2", func(t *testing.T) {
// 		gotEdgeKind2, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKind(t, gotEdgeKind2, edgeKind2)
// 	})
// 	// Expected fail - return error indicating non unique name
// 	t.Run("fail - create schema edge kind does not have a unique name", func(t *testing.T) {
// 		_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind2.Name, edgeKind2.SchemaExtensionId, edgeKind2.Description, edgeKind2.IsTraversable)
// 		require.ErrorIs(t, err, database.ErrDuplicateSchemaEdgeKindName)
// 	})

// 	// GET

// 	// Expected success - get first model.GraphSchemaRelationshipKind
// 	t.Run("success - get a schema edge kind #1", func(t *testing.T) {
// 		gotEdgeKind1, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, gotEdgeKind1.ID)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKind(t, gotEdgeKind1, edgeKind1)
// 	})
// 	// Expected fail - return error for if an edge kind that does not exist
// 	t.Run("fail - get an edge kind that does not exist", func(t *testing.T) {
// 		_, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKindById(testSuite.Context, 235)
// 		require.ErrorIs(t, err, database.ErrNotFound)
// 	})

// 	// GET With pagination / filtering

// 	// setup
// 	_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind3.Name, edgeKind3.SchemaExtensionId, edgeKind3.Description, edgeKind3.IsTraversable)
// 	require.NoError(t, err)
// 	_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind4.Name, edgeKind4.SchemaExtensionId, edgeKind4.Description, edgeKind4.IsTraversable)
// 	require.NoError(t, err)

// 	// Expected success - return all schema edge kinds
// 	t.Run("success - return edge schema kinds, no filter or sorting", func(t *testing.T) {
// 		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		require.Equal(t, 139, total) // Need to account for AD and Azure Relationship Kinds added by the extension migrator
// 		require.Len(t, edgeKinds, 139)
// 	})
// 	// Expected success - return schema edge kinds whose name is Test_Kind_3
// 	t.Run("success - return edge schema kinds using a filter", func(t *testing.T) {
// 		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
// 			model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: "test_edge_kind_3", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		require.Equal(t, 1, total)
// 		require.Len(t, edgeKinds, 1)
// 		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaRelationshipKinds{edgeKind3})
// 	})

// 	// Expected success - return schema edge kinds fuzzy filtering on description
// 	t.Run("success - return schema edge kinds using a fuzzy filterer", func(t *testing.T) {
// 		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
// 			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test edge kind ", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		require.Equal(t, 2, total)
// 		require.Len(t, edgeKinds, 2)
// 		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaRelationshipKinds{edgeKind3, edgeKind4})
// 	})
// 	// Expected success - return schema edge kinds fuzzy filtering on description and sort ascending on description
// 	t.Run("success - return schema edge kinds using a fuzzy filterer and an ascending sort column", func(t *testing.T) {
// 		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
// 			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test edge kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
// 				Direction: model.AscendingSortDirection,
// 				Column:    "description",
// 			}}, 0, 0)
// 		require.NoError(t, err)
// 		require.Equal(t, 2, total)
// 		require.Len(t, edgeKinds, 2)
// 		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaRelationshipKinds{edgeKind3, edgeKind4})
// 	})
// 	// Expected success - return schema edge kinds fuzzy filtering on description and sort descending on description
// 	t.Run("success - return schema edge kinds using a fuzzy filterer and a descending sort column", func(t *testing.T) {
// 		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
// 			model.Filters{"description": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test edge kind ", SetOperator: model.FilterAnd}}}, model.Sort{{
// 				Direction: model.DescendingSortDirection,
// 				Column:    "description",
// 			}}, 0, 0)
// 		require.NoError(t, err)
// 		require.Equal(t, 2, total)
// 		require.Len(t, edgeKinds, 2)
// 		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaRelationshipKinds{edgeKind4, edgeKind3})
// 	})
// 	// Expected success - return schema edge kinds, no filtering or sorting, with skip
// 	t.Run("success - return schema edge kinds using skip, no filtering or sorting", func(t *testing.T) {
// 		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, model.Filters{}, model.Sort{}, 2, 0)
// 		require.NoError(t, err)
// 		require.Equal(t, 4, total)
// 		require.Len(t, edgeKinds, 2)
// 		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaRelationshipKinds{edgeKind3, edgeKind4})
// 	})
// 	// Expected success - return schema edge kinds, no filtering or sorting, with limit
// 	t.Run("success - return schema edge kinds using limit, no filtering or sorting", func(t *testing.T) {
// 		edgeKinds, total, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
// 		require.NoError(t, err)
// 		require.Equal(t, 4, total)
// 		require.Len(t, edgeKinds, 2)
// 		compareGraphSchemaEdgeKinds(t, edgeKinds, model.GraphSchemaRelationshipKinds{edgeKind1, edgeKind2})
// 	})
// 	// Expected fail - return error for filtering on non-existent column
// 	t.Run("fail - return error for filtering on non-existent column", func(t *testing.T) {
// 		_, _, err = testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
// 			model.Filters{"nonexistentcolumn": []model.Filter{{Operator: model.Equals, Value: "blah", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
// 		require.EqualError(t, err, "ERROR: column \"nonexistentcolumn\" does not exist (SQLSTATE 42703)")
// 	})

// 	// UPDATE

// 	// Expected success - update edgeKind1 to updateWant, the name should NOT be updated
// 	t.Run("success - update edgeKind1 to updateWant", func(t *testing.T) {
// 		updateWant.ID = gotEdgeKind1.ID
// 		gotEdgeKind3, err := testSuite.BHDatabase.UpdateGraphSchemaRelationshipKind(testSuite.Context, updateWant)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKind(t, gotEdgeKind3, model.GraphSchemaRelationshipKind{
// 			Serial: model.Serial{
// 				Basic: model.Basic{
// 					CreatedAt: updateWant.CreatedAt,
// 					UpdatedAt: updateWant.UpdatedAt,
// 				},
// 			},
// 			SchemaExtensionId: updateWant.SchemaExtensionId,
// 			Name:              edgeKind1.Name,
// 			Description:       updateWant.Description,
// 			IsTraversable:     updateWant.IsTraversable,
// 		})
// 	})
// 	// Expected fail - return an error if trying to update an edge_kind that does not exist
// 	t.Run("fail - update an edge kind that does not exist", func(t *testing.T) {
// 		_, err = testSuite.BHDatabase.UpdateGraphSchemaRelationshipKind(testSuite.Context, model.GraphSchemaRelationshipKind{Serial: model.Serial{ID: 1123}, Name: edgeKind2.Name, SchemaExtensionId: extension.ID})
// 		require.ErrorIs(t, err, database.ErrNotFound)
// 	})

// 	// DELETE

// 	// Expected success - delete edge kind 1
// 	t.Run("success - delete edge kind 1", func(t *testing.T) {
// 		err = testSuite.BHDatabase.DeleteGraphSchemaRelationshipKind(testSuite.Context, gotEdgeKind1.ID)
// 		require.NoError(t, err)
// 	})
// 	// Expected fail - return an error if trying to delete an edge_kind that does not exist
// 	t.Run("fail - delete an edge kind that does not exist", func(t *testing.T) {
// 		err = testSuite.BHDatabase.DeleteGraphSchemaRelationshipKind(testSuite.Context, 1231)
// 		require.ErrorIs(t, err, database.ErrNotFound)
// 	})
// }

// // compareGraphSchemaNodeKinds - compares the returned list of model.GraphSchemaNodeKinds with the expected results.
// // Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
// func compareGraphSchemaNodeKinds(t *testing.T, got, want model.GraphSchemaNodeKinds) {
// 	t.Helper()
// 	require.Equalf(t, len(want), len(got), "length mismatch of NodeKindsInput")
// 	for i, schemaNodeKind := range got {
// 		compareGraphSchemaNodeKind(t, schemaNodeKind, want[i])
// 	}
// }

// compareGraphSchemaProperties - compares the returned list of model.GraphSchemaProperties with the expected results.
// // Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaProperties(t *testing.T, got, want model.GraphSchemaProperties) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of PropertiesInput")
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

// compareGraphSchemaEdgeKinds - compares the returned list of model.GraphSchemaRelationshipKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaEdgeKinds(t *testing.T, got, want model.GraphSchemaRelationshipKinds) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of RelationshipKindsInput")
	for i, schemaEdgeKind := range got {
		compareGraphSchemaEdgeKind(t, schemaEdgeKind, want[i])
	}
}

func compareGraphSchemaEdgeKind(t *testing.T, got, want model.GraphSchemaRelationshipKind) {
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

// func TestDatabase_GraphSchemaEdgeKindWithSchemaName_Get(t *testing.T) {
// 	t.Parallel()
// 	testSuite := setupIntegrationTestSuite(t)
// 	defer teardownIntegrationTestSuite(t, &testSuite)
// 	extensionA, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_a", "test_extension_a", "1.0.0", "Test")
// 	require.NoError(t, err)
// 	extensionB, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_schema_b", "test_extension_b", "1.0.0", "Test2")
// 	require.NoError(t, err)

// 	edgeKind1, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_1", extensionA.ID, "test edge kind 1", false)
// 	require.NoError(t, err)

// 	edgeKind2, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_2", extensionA.ID, "test edge kind 2", true)
// 	require.NoError(t, err)

// 	edgeKind3, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_3", extensionB.ID, "test edge kind 3", false)
// 	require.NoError(t, err)

// 	edgeKind4, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "test_edge_kind_4", extensionB.ID, "test edge kind 4", true)
// 	require.NoError(t, err)
// 	var (
// 		want1 = model.GraphSchemaEdgeKindWithNamedSchema{
// 			ID:            edgeKind1.ID,
// 			SchemaName:    extensionA.Name,
// 			Name:          edgeKind1.Name,
// 			Description:   edgeKind1.Description,
// 			IsTraversable: edgeKind1.IsTraversable,
// 		}
// 		want2 = model.GraphSchemaEdgeKindWithNamedSchema{
// 			ID:            edgeKind2.ID,
// 			SchemaName:    extensionA.Name,
// 			Name:          edgeKind2.Name,
// 			Description:   edgeKind2.Description,
// 			IsTraversable: edgeKind2.IsTraversable,
// 		}
// 		want3 = model.GraphSchemaEdgeKindWithNamedSchema{
// 			ID:            edgeKind3.ID,
// 			SchemaName:    extensionB.Name,
// 			Name:          edgeKind3.Name,
// 			Description:   edgeKind3.Description,
// 			IsTraversable: edgeKind3.IsTraversable,
// 		}
// 		want4 = model.GraphSchemaEdgeKindWithNamedSchema{
// 			ID:            edgeKind4.ID,
// 			SchemaName:    extensionB.Name,
// 			Name:          edgeKind4.Name,
// 			Description:   edgeKind4.Description,
// 			IsTraversable: edgeKind4.IsTraversable,
// 		}
// 	)

// 	t.Run("success - get a schema edge kind with named schema, no filters", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want1, want2, want3, want4}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema, filter for schema name", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.Equals, Value: extensionA.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want1, want2}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema, filter for multiple schema names", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.Equals, Value: extensionA.Name, SetOperator: model.FilterOr}, {Operator: model.Equals, Value: extensionB.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want1, want2, want3, want4}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema, filter for fuzzy match schema names", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.ApproximatelyEquals, Value: "test", SetOperator: model.FilterOr}, {Operator: model.Equals, Value: extensionB.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want1, want2, want3, want4}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema, filter for is_traversable", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"is_traversable": []model.Filter{{Operator: model.Equals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want2, want4}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema, filter for schema name and is_traversable", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.Equals, Value: extensionA.Name, SetOperator: model.FilterAnd}}, "is_traversable": []model.Filter{{Operator: model.Equals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want2}, actual)

// 	})

// 	t.Run("success - get a schema edge kind with named schema, filter for not equals schema name", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"schema.name": []model.Filter{{Operator: model.NotEquals, Value: extensionA.Name, SetOperator: model.FilterOr}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want3, want4}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema, filter for is not traversable", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"is_traversable": []model.Filter{{Operator: model.NotEquals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want1, want3}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema, sort by edge name descending", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{model.SortItem{Column: "name", Direction: model.DescendingSortDirection}}, 0, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want4, want3, want2, want1}, actual)

// 	})

// 	t.Run("success - get a schema edge kind with named schema using skip, no filtering or sorting", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{}, 1, 0)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want2, want3, want4}, actual)
// 	})

// 	t.Run("success - get a schema edge kind with named schema using limit, no filtering or sorting", func(t *testing.T) {
// 		actual, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{}, 0, 2)
// 		require.NoError(t, err)
// 		compareGraphSchemaEdgeKindsWithSchemaName(t, model.GraphSchemaEdgeKindsWithNamedSchema{want1, want2}, actual)
// 	})

// 	t.Run("fail - error building sql filter", func(t *testing.T) {
// 		_, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"is_traversable": []model.Filter{{Operator: "invalid", Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 2)
// 		require.EqualError(t, err, "invalid operator specified")
// 	})

// 	t.Run("fail - error building sql sort", func(t *testing.T) {
// 		_, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{}, model.Sort{model.SortItem{Column: "name", Direction: model.InvalidSortDirection}}, 0, 2)
// 		require.ErrorIs(t, err, database.ErrInvalidSortDirection)
// 	})

// 	t.Run("fail - attempt to filter non-existent column", func(t *testing.T) {
// 		_, _, err := testSuite.BHDatabase.GetGraphSchemaEdgeKindsWithSchemaName(testSuite.Context, model.Filters{"invalid": []model.Filter{{Operator: model.Equals, Value: "true", SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 2)
// 		require.EqualError(t, err, "ERROR: column \"invalid\" does not exist (SQLSTATE 42703)")
// 	})
// }

// compareGraphSchemaEdgeKindsWithSchemaName - compares the returned model.GraphSchemaEdgeKindsWithNamedSchema with the expected results.
func compareGraphSchemaEdgeKindsWithSchemaName(t *testing.T, expected, actual model.GraphSchemaEdgeKindsWithNamedSchema) {
	t.Helper()
	require.Equalf(t, len(expected), len(actual), "length mismatch of GraphSchemaEdgeKindsWithNamedSchema")
	for i, schemaEdgeKind := range actual {
		compareGraphSchemaEdgeKindWithNamedSchema(t, expected[i], schemaEdgeKind)
	}
}

func compareGraphSchemaEdgeKindWithNamedSchema(t *testing.T, expected, actual model.GraphSchemaEdgeKindWithNamedSchema) {
	t.Helper()
	require.Equalf(t, expected.Name, actual.Name, "GraphSchemaEdgeKindWithNamedSchema - name - got %v, want %v", actual.Name, expected.Name)
	require.Equalf(t, expected.Description, actual.Description, "GraphSchemaEdgeKindWithNamedSchema - description - got %v, want %v", actual.Description, expected.Description)
	require.Equalf(t, expected.IsTraversable, actual.IsTraversable, "GraphSchemaEdgeKindWithNamedSchema - IsTraversable - got %v, want %t", actual.IsTraversable, expected.IsTraversable)
	require.Equalf(t, expected.SchemaName, actual.SchemaName, "GraphSchemaEdgeKindWithNamedSchema - SchemaName - got %v, want %v", actual.SchemaName, expected.SchemaName)
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
			assert.Truef(t, found, "Expected environment (extension_id=%v, kind_id=%v, source_kind_id=%v) not found",
				want.SchemaExtensionId, want.EnvironmentKindId, want.SourceKindId)
		}
	}

	assertContainsEnvironment := func(t *testing.T, got model.SchemaEnvironment, expected ...model.SchemaEnvironment) {
		t.Helper()
		assertContainsEnvironments(t, []model.SchemaEnvironment{got}, expected...)
	}

	tests := []struct {
		name   string
		assert func(testSuite IntegrationTestSuite)
	}{
		// CreateEnvironment
		{
			name: "Success: create an environment",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "failed to create environment")

				// Validate created environment is as expected
				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, newEnvironment.ID)
				assert.NoError(t, err, "failed to retrieve environment")

				assertContainsEnvironment(t, retrievedEnvironment, environment)

			},
		},
		{
			name: "Error: fails to create duplicate environment",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				// Create new environment
				got, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "failed to create environment")

				assertContainsEnvironment(t, got, environment)

				// Create same environment again
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				// Assert error
				assert.ErrorIs(t, err, database.ErrDuplicateSchemaEnvironment)
			},
		},
		// GetEnvironmentByKinds
		{
			name: "Success: get environment by kinds - kind id and source id are unique",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "failed to create environment")

				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentByKinds(testSuite.Context, newEnvironment.EnvironmentKindId, newEnvironment.SourceKindId)
				assert.NoError(t, err, database.ErrNotFound)

				assertContainsEnvironment(t, retrievedEnvironment, environment)

			},
		},
		{
			name: "Error: fail to get environment by unknown kinds",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				environment := model.SchemaEnvironment{
					EnvironmentKindId: 20586,
					SourceKindId:      257958,
				}

				_, err := testSuite.BHDatabase.GetEnvironmentByKinds(testSuite.Context, environment.EnvironmentKindId, environment.SourceKindId)
				assert.EqualError(t, err, database.ErrNotFound.Error(), "Expected entity not found")
			},
		},
		// GetEnvironmentById
		{
			name: "Success: get environment by id",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "failed to create environment")

				retrievedEnvironment, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, newEnvironment.ID)
				assert.NoError(t, err, "failed to get environment by id")

				assertContainsEnvironment(t, retrievedEnvironment, environment)

			},
		},
		{
			name: "Error: fail to retrieve environment by id that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetEnvironments
		{
			name: "Success: return environments",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension_1", "test_extension_1", "1.0.0", "Test1")
				require.NoError(t, err)

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
				require.NoError(t, err, "failed to create environment 1")

				// Create Environment 2
				_, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment2.SchemaExtensionId, environment2.EnvironmentKindId, environment2.SourceKindId)
				require.NoError(t, err, "failed to create environment 2")

				// Get Environments back
				environments, err := testSuite.BHDatabase.GetEnvironmentsByExtensionId(testSuite.Context, extension.ID)
				assert.NoError(t, err, "failed to retrieve environments by extension id")

				// Validate number of results is 2 environments created in this test
				// The reason that we don't need to total the baseline number of environments
				// is because we are retrieving environments based on the extension ID created
				// in this test.
				assert.Len(t, environments, 2, "Expected total environments on the extension to be 2")

				// Validate all created environments exist in the results
				assertContainsEnvironments(t, environments, environment1, environment2)
			},
		},
		// DeleteEnvironment
		{
			name: "Success: environment deleted",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId:          extension.ID,
					SchemaExtensionDisplayName: "DisplayName",
					EnvironmentKindId:          1,
					EnvironmentKindName:        "Tag_Tier_Zero",
					SourceKindId:               1,
				}

				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				assert.NoError(t, err, "failed to create environment")

				assertContainsEnvironment(t, newEnvironment, environment)

				err = testSuite.BHDatabase.DeleteEnvironment(testSuite.Context, newEnvironment.ID)
				assert.NoError(t, err, "failed to delete environment for extension")

				// Validate environment no longer exists
				_, err = testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, newEnvironment.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete environment that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
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
			testCase.assert(testSuite)
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
			assert.Truef(t, found, "Expected finding %v not found", want.Name)
		}
	}

	assertContainsFinding := func(t *testing.T, got model.SchemaRelationshipFinding, expected ...model.SchemaRelationshipFinding) {
		t.Helper()
		assertContainsFindings(t, []model.SchemaRelationshipFinding{got}, expected...)
	}

	tests := []struct {
		name   string
		assert func(testSuite IntegrationTestSuite)
	}{
		// CreateSchemaRelationshipFinding
		{
			name: "Success: create an environment",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "failed to create finding")

				// Validate created finding is as expected
				retrievedFinding, err := testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "failed to retrieve environment")

				assertContainsFinding(t, retrievedFinding, finding)

			},
		},
		{
			name: "Error: fails to create duplicate finding",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "failed to create finding")

				// Create same finding again
				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, newFinding.SchemaExtensionId, newFinding.RelationshipKindId, newFinding.EnvironmentId, newFinding.Name, newFinding.DisplayName)
				// Assert error
				assert.ErrorIs(t, err, database.ErrDuplicateSchemaRelationshipFindingName)
			},
		},
		// GetSchemaRelationshipFindingById
		{
			name: "Success: get finding by id",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "failed to create finding")

				retrievedFinding, err := testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "failed to get finding by id")

				assertContainsFinding(t, retrievedFinding, finding)
			},
		},
		{
			name: "Error: fail to retrieve finding by id that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetSchemaRelationshipFindingByName
		{
			name: "Success: get finding by name",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "failed to create finding")

				retrievedFinding, err := testSuite.BHDatabase.GetSchemaRelationshipFindingByName(testSuite.Context, newFinding.Name)
				assert.NoError(t, err, "failed to get finding by name")

				assertContainsFinding(t, retrievedFinding, finding)
			},
		},
		{
			name: "Error: fail to retrieve finding by name that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetSchemaRelationshipFindingByName(testSuite.Context, "doesnotexist")
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// DeleteSchemaRelationshipFinding
		{
			name: "Success: finding deleted",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				assert.NoError(t, err, "failed to create finding")

				assertContainsFinding(t, newFinding, finding)

				// Delete Finding
				err = testSuite.BHDatabase.DeleteSchemaRelationshipFinding(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "failed to delete environment for extension")

				// Validate finding no longer exists
				_, err = testSuite.BHDatabase.GetSchemaRelationshipFindingById(testSuite.Context, newFinding.ID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete finding that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				// Delete Finding
				err := testSuite.BHDatabase.DeleteSchemaRelationshipFinding(testSuite.Context, int32(10000))
				assert.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		// GetSchemaRelationshipFindingsBySchemaExtensionId
		{
			name: "Error: failed to get findings that do not exist",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				// Get Findings
				_, err := testSuite.BHDatabase.GetSchemaRelationshipFindingsBySchemaExtensionId(testSuite.Context, int32(10000))
				assert.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Success: retrieve multiple findings by schema extension id",
			assert: func(testSuite IntegrationTestSuite) {
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
				require.NoError(t, err)

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

				// Create Environment
				createdEnvironmentNode, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)

				// Create Source Kind Node
				createdSourceKindNode, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err)

				// Retrieve DAWGS Environment Kind
				envKind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode.Name)
				require.NoError(t, err)

				// Register Source Kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(sourceKind1.Name))
				require.NoError(t, err)

				// Retrieve Source Kind
				sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err)

				// Create Environment
				createdEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					envKind.ID, int32(sourceKind.ID))
				require.NoError(t, err)

				// Create Finding Edge Kind
				createdEdgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)

				// Retrieve Finding Edge Kind
				edgeKind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind.Name)
				require.NoError(t, err)

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
				require.NoError(t, err)

				// Create Finding 2
				_, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, edgeKind.ID, createdEnvironment.ID, finding2.Name, finding2.DisplayName)
				require.NoError(t, err)

				// Get Findings by Extension ID
				findings, err := testSuite.BHDatabase.GetSchemaRelationshipFindingsBySchemaExtensionId(testSuite.Context, createdExtension.ID)
				assert.NoError(t, err)

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
			testCase.assert(testSuite)
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
			assert.Truef(t, found, "Expected remediation for finding_id %v not found", want.FindingID)
		}
	}

	assertContainsRemediation := func(t *testing.T, got model.Remediation, expected ...model.Remediation) {
		t.Helper()
		assertContainsRemediations(t, []model.Remediation{got}, expected...)
	}

	tests := []struct {
		name   string
		assert func(testSuite IntegrationTestSuite)
	}{
		// CreateRemediation
		{
			name: "Success: create a remediation",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "failed to create finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				assert.NoError(t, err, "failed to create remediation")

				// Validate created remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "failed to retrieve remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, remediation)
			},
		},
		{
			name: "Error: fails to create duplicate remediation",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "failed to create finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "failed to create remediation")

				// Create same remediation again
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				// Assert error
				assert.EqualError(t, err, "ERROR: duplicate key value violates unique constraint \"schema_remediations_pkey\" (SQLSTATE 23505)")
			},
		},
		// GetRemediationByFindingId
		{
			name: "Success: get remediation by finding id",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "failed to create finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "failed to create remediation")

				// Validate created remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, newFinding.ID)
				assert.NoError(t, err, "failed to retrieve remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, remediation)
			},
		},
		{
			name: "Error: fail to retrieve remediation by id that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, int32(5000))
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// GetRemediationByFindingName
		{
			name: "Success: get remediation by finding name",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "failed to create finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				assert.NoError(t, err, "failed to create remediation")

				// Validate created remediation is as expected
				retrievedRemediation, err := testSuite.BHDatabase.GetRemediationByFindingName(testSuite.Context, newFinding.Name)
				assert.NoError(t, err, "failed to retrieve remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, remediation)
			},
		},
		{
			name: "Error: fail to retrieve remediation by finding name that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				_, err := testSuite.BHDatabase.GetRemediationByFindingName(testSuite.Context, "namedoesnotexist")
				require.ErrorIs(t, err, database.ErrNotFound)
			},
		},
		// UpdateRemediation
		{
			name: "Success: remediation updated",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "failed to create finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				createdRemediation, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "failed to create remediation")

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
				assert.NoError(t, err, "failed to retrieve updated remediation by finding id")

				assertContainsRemediation(t, retrievedRemediation, updatedRemediation)
			},
		},
		{
			name: "Error: failed to update remediation that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
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
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				finding := model.SchemaRelationshipFinding{
					SchemaExtensionId:  extension.ID,
					RelationshipKindId: 1,
					EnvironmentId:      1,
					Name:               "finding",
					DisplayName:        "display name",
				}

				// Create new finding
				newFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, finding.SchemaExtensionId, finding.RelationshipKindId, finding.EnvironmentId, finding.Name, finding.DisplayName)
				require.NoError(t, err, "failed to create finding")

				remediation := model.Remediation{
					FindingID:        newFinding.ID,
					ShortDescription: "Short desc",
					LongDescription:  "Long desc",
					ShortRemediation: "Short fix",
					LongRemediation:  "Long fix",
				}

				// Create new remediation
				createdRemediation, err := testSuite.BHDatabase.CreateRemediation(testSuite.Context, remediation.FindingID, remediation.ShortDescription, remediation.LongDescription, remediation.ShortRemediation, remediation.LongRemediation)
				require.NoError(t, err, "failed to create remediation")

				assertContainsRemediation(t, createdRemediation, remediation)

				// Delete Remediation
				err = testSuite.BHDatabase.DeleteRemediation(testSuite.Context, remediation.FindingID)
				assert.NoError(t, err, "failed to delete remediation by finding id")

				// Validate remediation no longer exists
				_, err = testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, remediation.FindingID)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "Error: failed to delete remediation that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
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
			testCase.assert(testSuite)
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
			assert.Truef(t, found, "Expected principal kind (env_id=%v, kind=%v) not found",
				want.EnvironmentId, want.PrincipalKind)
		}
	}

	tests := []struct {
		name   string
		assert func(testSuite IntegrationTestSuite)
	}{
		// CreatePrincipalKind
		{
			name: "Success: create principal kind",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "failed to create environment")

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}
				// Create new principal kind
				newPrincipalKind, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				assert.NoError(t, err, "failed to create principal kind")

				// Validate created principalKind is as expected
				retrievedPrincipalKind, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				assert.NoError(t, err, "failed to retrieve principal kind by environment id")

				assertContainsPrincipalKinds(t, retrievedPrincipalKind, principalKind)
			},
		},
		{
			name: "Error: fails to create duplicate principal kind",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "failed to create environment")

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}
				// Create new principal kind
				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				assert.NoError(t, err, "failed to create principal kind")

				// Create same principal kind again
				_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				// Assert error
				assert.EqualError(t, err, "duplicate principal kind: ERROR: duplicate key value violates unique constraint \"schema_environments_principal_kinds_pkey\" (SQLSTATE 23505)")
			},
		},
		// GetPrincipalKindsByEnvironmentId
		{
			name: "Success: get principal kinds by environment id",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "failed to create environment")

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}
				// Create new principal kind
				newPrincipalKind, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				require.NoError(t, err, "failed to create principal kind")

				// Validate we are able to retrieve principal kinds back by environment id
				retrievedPrincipalKind, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				assert.NoError(t, err, "failed to retrieve principal kind by environment id")

				assertContainsPrincipalKinds(t, retrievedPrincipalKind, principalKind)
			},
		},
		{
			name: "Success: principal kinds should return empty if none are found",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				principalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, int32(5000))
				assert.NoError(t, err)
				assert.Len(t, principalKinds, 0)
			},
		},
		// DeletePrincipalKind
		{
			name: "Success: principal kind deleted",
			assert: func(testSuite IntegrationTestSuite) {
				t.Helper()

				extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "test_extension", "test_extension", "1.0.0", "Test")
				require.NoError(t, err)

				environment := model.SchemaEnvironment{
					SchemaExtensionId: extension.ID,
					EnvironmentKindId: 1,
					SourceKindId:      1,
				}

				// Create new environment
				newEnvironment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, environment.SchemaExtensionId, environment.EnvironmentKindId, environment.SourceKindId)
				require.NoError(t, err, "failed to create environment")

				principalKind := model.SchemaEnvironmentPrincipalKind{
					EnvironmentId: newEnvironment.ID,
					PrincipalKind: 1,
				}
				// Create new principal kind
				newPrincipalKind, err := testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, principalKind.EnvironmentId, principalKind.PrincipalKind)
				require.NoError(t, err, "failed to create principal kind")

				// Validate created principalKind is as expected
				_, err = testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				require.NoError(t, err, "failed to retrieve principal kind by environment id")

				// Delete Principal Kind
				err = testSuite.BHDatabase.DeletePrincipalKind(testSuite.Context, newPrincipalKind.EnvironmentId, newPrincipalKind.PrincipalKind)
				assert.NoError(t, err, "failed to delete principal kind")

				// Validate principal kind no longer exists
				// Principal kinds returns empty slice when not found instead of error
				foundPrincipalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, newPrincipalKind.EnvironmentId)
				assert.NoError(t, err)
				assert.Len(t, foundPrincipalKinds, 0)
			},
		},
		{
			name: "Error: failed to delete principal kind that does not exist",
			assert: func(testSuite IntegrationTestSuite) {
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
			testCase.assert(testSuite)
		})
	}
}

func TestDeleteSchemaExtension_CascadeDeletesAllDependents(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	extension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, "CascadeTestExtension", "Cascade Test Extension", "v1.0.0", "CTE")
	require.NoError(t, err)

	nodeKind, err := testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, "CascadeTestNodeKind", extension.ID, "Cascade Test Node Kind", "Test description", false, "fa-test", "#000000")
	require.NoError(t, err)
	dawgsEnvKind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, "CascadeTestNodeKind")
	require.NoError(t, err)
	err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("CascadeTestSourceKind"))
	require.NoError(t, err)
	sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, "CascadeTestSourceKind")
	require.NoError(t, err)

	property, err := testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, extension.ID, "cascade_test_property", "Cascade Test Property", "string", "Test description")
	require.NoError(t, err)

	edgeKind, err := testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, "CascadeTestEdgeKind", extension.ID, "Test description", true)
	require.NoError(t, err)

	environment, err := testSuite.BHDatabase.CreateEnvironment(testSuite.Context, extension.ID, dawgsEnvKind.ID, int32(sourceKind.ID))
	require.NoError(t, err)

	relationshipFinding, err := testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context, extension.ID, edgeKind.ID, environment.ID, "CascadeTestFinding", "Cascade Test Finding")
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, relationshipFinding.ID, "Short desc", "Long desc", "Short remediation", "Long remediation")
	require.NoError(t, err)

	_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, environment.ID, dawgsEnvKind.ID)
	require.NoError(t, err)

	err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extension.ID)
	require.NoError(t, err)

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
