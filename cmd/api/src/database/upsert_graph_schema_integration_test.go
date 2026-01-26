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
	"context"
	"strconv"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertGraphSchemaNodeKinds(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		err error

		testExtensionName = "Test_Extension_Upsert_Test"
		testExtension     = model.GraphSchemaExtension{
			Name:        testExtensionName,
			Version:     "1.0.0",
			DisplayName: "Test Extension",
			Namespace:   "TESTUPSERT",
		}
		newNodeKind1 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_1",
			DisplayName:   "Test Node Kind 1",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}

		newNodeKind2 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}

		newNodeKind3 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newNodeKind4 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}

		existingNodeKind1 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_1",
			DisplayName:   "Test Node Kind 1",
			Description:   "Test Node Kind 1",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}

		existingNodeKind2 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "Test Node Kind 2",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingNodeKind3 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "Test Node Kind 3",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}

		existingNodeKind4 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "Test Node Kind 4",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}

		updateNodeKind4 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Update_Node_Kind_4",
			DisplayName:   "Node Kind 4",
			Description:   "Node Kind 4",
			IsDisplayKind: true,
			Icon:          "Desktop",
			IconColor:     "orange",
		}

		// Used for creating an existing graph schema
		existingNodeKinds = model.GraphSchemaNodeKinds{existingNodeKind1, existingNodeKind2, existingNodeKind3, existingNodeKind4}

		// Used in both args and want, in doing so the SchemaExtensionId will be propagated back to the want
		// so the value, this allows for equality comparisons of the SchemaExtensionId rather than just checking
		// if it exists
		newNodeKinds    = model.GraphSchemaNodeKinds{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4}
		updateNodeKinds = model.GraphSchemaNodeKinds{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4,
			existingNodeKind1, updateNodeKind4}
	)

	type fields struct {
		setup    func(*testing.T) (int32, model.GraphSchemaNodeKinds)
		teardown func(*testing.T, int32)
	}

	type args struct {
		ctx                  context.Context
		graphSchemaNodeKinds model.GraphSchemaNodeKinds
	}
	tests := []struct {
		name                     string
		fields                   fields
		args                     args
		wantErr                  error
		wantGraphSchemaNodeKinds model.GraphSchemaNodeKinds
	}{
		{
			name: "success - create schema new node kinds",
			fields: fields{
				setup: func(t *testing.T) (int32, model.GraphSchemaNodeKinds) {
					t.Helper()
					var createdExtension model.GraphSchemaExtension
					createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version, testExtension.Namespace)
					require.NoError(t, err)

					return createdExtension.ID, model.GraphSchemaNodeKinds{}
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx:                  testSuite.Context,
				graphSchemaNodeKinds: newNodeKinds,
			},
			wantErr:                  nil,
			wantGraphSchemaNodeKinds: newNodeKinds,
		},
		{
			name: "success - update schema node kinds",
			fields: fields{
				setup: func(t *testing.T) (int32, model.GraphSchemaNodeKinds) {
					t.Helper()
					// Create a graph schema and ensure it exists
					var (
						gotExistingGraphExtension model.GraphExtension
						createdExtension          model.GraphSchemaExtension
					)

					createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version, testExtension.Namespace)
					require.NoError(t, err)

					for _, nodeKind := range existingNodeKinds {
						_, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind.Name,
							createdExtension.ID, nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind,
							nodeKind.Icon, nodeKind.IconColor)
						require.NoError(t, err)
					}

					gotExistingGraphExtension, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.NoError(t, err)
					return createdExtension.ID, gotExistingGraphExtension.GraphSchemaNodeKinds
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx:                  testSuite.Context,
				graphSchemaNodeKinds: updateNodeKinds,
			},
			wantErr:                  nil,
			wantGraphSchemaNodeKinds: updateNodeKinds,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				returnedGraphSchemaNodeKinds model.GraphSchemaNodeKinds
				gotGraphSchemaNodeKinds      model.GraphSchemaNodeKinds
			)
			extensionId, existingGraphSchemaNodeKinds := tt.fields.setup(t)

			for idx := range tt.args.graphSchemaNodeKinds {
				tt.args.graphSchemaNodeKinds[idx].SchemaExtensionId = extensionId
			}

			if returnedGraphSchemaNodeKinds, err = testSuite.BHDatabase.UpsertGraphSchemaNodeKinds(tt.args.ctx, tt.args.graphSchemaNodeKinds, existingGraphSchemaNodeKinds); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			} else {

				require.NoError(t, err)

				// Retrieve and compare upserted graph schema
				gotGraphSchemaNodeKinds, _, err = testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
					model.Filters{"schema_extension_id": []model.Filter{{
						Operator:    model.Equals,
						Value:       strconv.FormatInt(int64(extensionId), 10),
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 0)
				require.NoError(t, err)

				compareMapOfGraphSchemaNodeKinds(t, gotGraphSchemaNodeKinds.ToMapKeyedOnName(), tt.wantGraphSchemaNodeKinds.ToMapKeyedOnName())
				require.Equal(t, returnedGraphSchemaNodeKinds, gotGraphSchemaNodeKinds) // ensure what's returned matches what is retrieved
			}
			tt.fields.teardown(t, extensionId)
		})
	}
}

func TestBloodhoundDB_UpsertGraphSchemaEdgeKinds(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		err error

		testExtensionName = "Test_Extension_Upsert_Test"
		testExtension     = model.GraphSchemaExtension{
			Name:        testExtensionName,
			Version:     "1.0.0",
			DisplayName: "Test Extension",
			Namespace:   "TESTUPSERT",
		}

		// Used for creating an existing graph schema

		// Used in both args and want, in doing so the SchemaExtensionId will be propagated back to the want
		// so the value, this allows for equality comparisons of the SchemaExtensionId rather than just checking
		// if it exists
		newEdgeKind1 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		newEdgeKind2 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		newEdgeKind3 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		newEdgeKind4 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}

		existingEdgeKind1 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		existingEdgeKind2 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		existingEdgeKind3 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		existingEdgeKind4 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		updateEdgeKind4 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Update_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}

		existingEdgeKinds = model.GraphSchemaEdgeKinds{existingEdgeKind1, existingEdgeKind2, existingEdgeKind3, existingEdgeKind4}

		newEdgeKinds    = model.GraphSchemaEdgeKinds{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4}
		updateEdgeKinds = model.GraphSchemaEdgeKinds{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4,
			existingEdgeKind1, updateEdgeKind4}
	)

	type fields struct {
		setup    func(*testing.T) (int32, model.GraphSchemaEdgeKinds)
		teardown func(*testing.T, int32)
	}

	type args struct {
		ctx                  context.Context
		graphSchemaEdgeKinds model.GraphSchemaEdgeKinds
	}
	tests := []struct {
		name                     string
		fields                   fields
		args                     args
		wantErr                  error
		wantGraphSchemaNodeKinds model.GraphSchemaEdgeKinds
	}{
		{
			name: "success - create new schema edge kinds",
			fields: fields{
				setup: func(t *testing.T) (int32, model.GraphSchemaEdgeKinds) {
					t.Helper()
					var createdExtension model.GraphSchemaExtension
					createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version, testExtension.Namespace)
					require.NoError(t, err)

					return createdExtension.ID, model.GraphSchemaEdgeKinds{}
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx:                  testSuite.Context,
				graphSchemaEdgeKinds: newEdgeKinds,
			},
			wantErr:                  nil,
			wantGraphSchemaNodeKinds: newEdgeKinds,
		},
		{
			name: "success - update schema edge kinds",
			fields: fields{
				setup: func(t *testing.T) (int32, model.GraphSchemaEdgeKinds) {
					t.Helper()
					// Create a graph schema and ensure it exists
					var (
						gotExistingGraphExtension model.GraphExtension
						createdExtension          model.GraphSchemaExtension
					)

					createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version, testExtension.Namespace)
					require.NoError(t, err)

					for _, edgeKind := range existingEdgeKinds {
						_, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind.Name,
							createdExtension.ID, edgeKind.Description, edgeKind.IsTraversable)
						require.NoError(t, err)
					}

					gotExistingGraphExtension, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.NoError(t, err)
					return gotExistingGraphExtension.GraphSchemaExtension.ID, gotExistingGraphExtension.GraphSchemaEdgeKinds
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx:                  testSuite.Context,
				graphSchemaEdgeKinds: updateEdgeKinds,
			},
			wantErr:                  nil,
			wantGraphSchemaNodeKinds: updateEdgeKinds,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				returnedGraphSchemaEdgeKinds model.GraphSchemaEdgeKinds
				gotGraphSchemaEdgeKinds      model.GraphSchemaEdgeKinds
			)
			extensionId, existingGraphSchemaEdgeKinds := tt.fields.setup(t)

			for idx := range tt.args.graphSchemaEdgeKinds {
				tt.args.graphSchemaEdgeKinds[idx].SchemaExtensionId = extensionId
			}

			if returnedGraphSchemaEdgeKinds, err = testSuite.BHDatabase.UpsertGraphSchemaEdgeKinds(tt.args.ctx,
				tt.args.graphSchemaEdgeKinds, existingGraphSchemaEdgeKinds); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)

				// Retrieve and compare upserted graph schema
				gotGraphSchemaEdgeKinds, _, err = testSuite.BHDatabase.GetGraphSchemaEdgeKinds(testSuite.Context,
					model.Filters{"schema_extension_id": []model.Filter{{
						Operator:    model.Equals,
						Value:       strconv.FormatInt(int64(extensionId), 10),
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 0)
				require.NoError(t, err)

				compareMapOfGraphSchemaEdgeKinds(t, gotGraphSchemaEdgeKinds.ToMapKeyedOnName(), tt.wantGraphSchemaNodeKinds.ToMapKeyedOnName())
				require.Equal(t, returnedGraphSchemaEdgeKinds, gotGraphSchemaEdgeKinds) // ensure what's returned matches what is retrieved
			}
			tt.fields.teardown(t, extensionId)
		})
	}
}

func TestBloodhoundDB_UpsertGraphSchemaProperties(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		err error

		testExtensionName = "Test_Extension_Upsert_Test"
		testExtension     = model.GraphSchemaExtension{
			Name:        testExtensionName,
			Version:     "1.0.0",
			DisplayName: "Test Extension",
			Namespace:   "TESTUPSERT",
		}

		// Used for creating an existing graph schema

		// Used in both args and want, in doing so the SchemaExtensionId will be propagated back to the want
		// so the value, this allows for equality comparisons of the SchemaExtensionId rather than just checking
		// if it exists
		newProperty1 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_1",
			DisplayName: "Test Property 1",
			DataType:    "string",
			Description: "Test Property 1",
		}

		newProperty2 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_2",
			DisplayName: "Test Property 2",
			DataType:    "string",
			Description: "Test Property 2",
		}

		newProperty3 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_3",
			DisplayName: "Test Property 3",
			DataType:    "string",
			Description: "Test Property 3",
		}

		newProperty4 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_4",
			DisplayName: "Test Property 4",
			DataType:    "string",
			Description: "Test Property 4",
		}

		existingProperty1 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_1",
			DisplayName: "Property 1",
			DataType:    "string",
			Description: "Property 1",
		}

		existingProperty2 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_2",
			DisplayName: "Property 2",
			DataType:    "string",
			Description: "Property 2",
		}

		existingProperty3 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_3",
			DisplayName: "Property 3",
			DataType:    "string",
			Description: "Property 3",
		}

		existingProperty4 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_4",
			DisplayName: "Property 4",
			DataType:    "string",
			Description: "Property 4",
		}

		updateProperty4 = model.GraphSchemaProperty{
			Name:        "Upsert_Update_Property_4",
			DisplayName: "Property 4",
			DataType:    "string",
			Description: "Property 4",
		}

		existingProperties = model.GraphSchemaProperties{existingProperty1, existingProperty2, existingProperty3, existingProperty4}
		newProperties      = model.GraphSchemaProperties{newProperty1, newProperty2, newProperty3, newProperty4}
		updateProperties   = model.GraphSchemaProperties{newProperty1, newProperty2, newProperty3,
			newProperty4, existingProperty1, updateProperty4}
	)

	type fields struct {
		setup    func(*testing.T) (int32, model.GraphSchemaProperties)
		teardown func(*testing.T, int32)
	}

	type args struct {
		ctx                   context.Context
		graphSchemaProperties model.GraphSchemaProperties
	}
	tests := []struct {
		name                     string
		fields                   fields
		args                     args
		wantErr                  error
		wantGraphSchemaNodeKinds model.GraphSchemaProperties
	}{
		{
			name: "success - create new properties",
			fields: fields{
				setup: func(t *testing.T) (int32, model.GraphSchemaProperties) {
					t.Helper()
					var createdExtension model.GraphSchemaExtension
					createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version, testExtension.Namespace)
					require.NoError(t, err)

					return createdExtension.ID, model.GraphSchemaProperties{}
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx:                   testSuite.Context,
				graphSchemaProperties: newProperties,
			},
			wantGraphSchemaNodeKinds: newProperties,
		},
		{
			name: "success - update schema properties",
			fields: fields{
				setup: func(t *testing.T) (int32, model.GraphSchemaProperties) {
					t.Helper()
					// Create a graph schema and ensure it exists
					var (
						gotExistingGraphExtension model.GraphExtension
						createdExtension          model.GraphSchemaExtension
					)

					createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version, testExtension.Namespace)
					require.NoError(t, err)

					for _, property := range existingProperties {
						_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, createdExtension.ID, property.Name,
							property.DisplayName, property.Description, property.Description)
						require.NoError(t, err)
					}

					gotExistingGraphExtension, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.NoError(t, err)
					return gotExistingGraphExtension.GraphSchemaExtension.ID, gotExistingGraphExtension.GraphSchemaProperties
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx:                   testSuite.Context,
				graphSchemaProperties: updateProperties,
			},
			wantGraphSchemaNodeKinds: updateProperties,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				returnedGraphSchemaEdgeKinds model.GraphSchemaProperties
				gotGraphSchemaProperties     model.GraphSchemaProperties
			)
			extensionId, existingGraphSchemaProperties := tt.fields.setup(t)

			for idx := range tt.args.graphSchemaProperties {
				tt.args.graphSchemaProperties[idx].SchemaExtensionId = extensionId
			}

			if returnedGraphSchemaEdgeKinds, err = testSuite.BHDatabase.UpsertGraphSchemaProperties(tt.args.ctx,
				tt.args.graphSchemaProperties, existingGraphSchemaProperties); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {

				require.NoError(t, err)

				// Retrieve and compare upserted graph schema
				gotGraphSchemaProperties, _, err = testSuite.BHDatabase.GetGraphSchemaProperties(testSuite.Context,
					model.Filters{"schema_extension_id": []model.Filter{{
						Operator:    model.Equals,
						Value:       strconv.FormatInt(int64(extensionId), 10),
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 0)
				require.NoError(t, err)

				compareMapOfGraphSchemaProperties(t, gotGraphSchemaProperties.ToMapKeyedOnName(),
					tt.wantGraphSchemaNodeKinds.ToMapKeyedOnName())
				require.Equal(t, returnedGraphSchemaEdgeKinds, gotGraphSchemaProperties) // ensure what's returned matches what is retrieved
			}
			tt.fields.teardown(t, extensionId)
		})
	}
}

func compareGraphSchemaExtension(t *testing.T, got, want model.GraphSchemaExtension) {
	t.Helper()
	require.Greaterf(t, got.ID, int32(0), "GraphSchemaExtension - ID mismatch - got: %v", got.ID)
	require.Equalf(t, want.Name, got.Name, "GraphSchemaExtension - name mismatch - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.DisplayName, got.DisplayName, "GraphSchemaExtension - display_name mismatch - got %v, want %v", got.DisplayName, want.DisplayName)
	require.Equalf(t, want.Version, got.Version, "GraphSchemaExtension - version mismatch - got %v, want %v", got.Version, want.Version)
	require.Equalf(t, want.IsBuiltin, got.IsBuiltin, "GraphSchemaExtension - is_built mismatch - got %t, want %t", got.IsBuiltin, want.IsBuiltin)
	require.Equalf(t, want.Namespace, got.Namespace, "GraphSchemaExtension - namespace mismatch - got %v, want %v", got.Namespace, want.Namespace)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "GraphSchemaExtension - created_at mismatch - got: %s", got.CreatedAt.String())
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "GraphSchemaExtension - updated_at mismatch - got: %s", got.UpdatedAt.String())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaExtension - deleted_at is not null")
}

// compareMapOfGraphSchemaNodeKinds - compares two maps of model.GraphSchemaNodeKinds, use this if comparing
// model.GraphSchemaNodeKinds and ordering does not matter. Since the upsert extension function can insert
// nodes in any order, the other compareGraphSchemaNodeKinds test func cannot be used.
func compareMapOfGraphSchemaNodeKinds(t *testing.T, got, want map[string]model.GraphSchemaNodeKind) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaNodeKinds")
	for k, nodeKind := range got {
		if _, ok := want[k]; ok {
			compareGraphSchemaNodeKind(t, nodeKind, want[k])
		} else {
			require.FailNow(t, "node kind not found in want map", "nodeKind: %v", nodeKind.Name)
		}
	}
}

// compareMapOfGraphSchemaEdgeKinds - compares two maps of model.GraphSchemaEdgeKinds, use this if comparing
// model.GraphSchemaEdgeKinds and ordering does not matter. Since the upsert extension function can insert
// nodes in any order, the other compareGraphSchemaEdgeKinds test func cannot be used.
func compareMapOfGraphSchemaEdgeKinds(t *testing.T, got, want map[string]model.GraphSchemaEdgeKind) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaEdgeKinds")
	for k, edgeKind := range got {
		if _, ok := want[k]; ok {
			compareGraphSchemaEdgeKind(t, edgeKind, want[k])
		} else {
			require.FailNow(t, "edge kind not found in want map", "edgeKind: %v", edgeKind.Name)
		}
	}
}

// compareMapOfGraphSchemaProperties - compares two maps of model.GraphSchemaProperty, use this if comparing
// model.GraphSchemaProperties and ordering does not matter. Since the upsert extension function can insert
// nodes in any order, the other compareGraphSchemaProperties test func cannot be used.
func compareMapOfGraphSchemaProperties(t *testing.T, got, want map[string]model.GraphSchemaProperty) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaProperties")
	for k, property := range got {
		if _, ok := want[k]; ok {
			compareGraphSchemaProperty(t, property, want[k])
		} else {
			require.FailNow(t, "property not found in want map", "property: %v", property.Name)
		}
	}
}
