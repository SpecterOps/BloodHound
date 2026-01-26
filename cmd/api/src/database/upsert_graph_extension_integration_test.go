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
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertOpenGraphExtension(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		err error
		got bool

		testExtensionName = "Test_Extension_Upsert_Test"
		testExtension     = model.GraphSchemaExtension{
			Name:        testExtensionName,
			Version:     "1.0.0",
			DisplayName: "Test Extension",
			Namespace:   "Upsert",
		}
		newNodeKind1 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_1",
			DisplayName:   "Test Node Kind 1",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind1 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		newProperty1 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_1",
			DisplayName: "Test Property 1",
			DataType:    "string",
			Description: "Test Property 1",
		}
		newNodeKind2 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind2 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		newProperty2 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_2",
			DisplayName: "Test Property 2",
			DataType:    "string",
			Description: "Test Property 2",
		}
		newNodeKind3 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind3 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		newProperty3 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_3",
			DisplayName: "Test Property 3",
			DataType:    "string",
			Description: "Test Property 3",
		}
		newNodeKind4 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind4 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_New_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		newProperty4 = model.GraphSchemaProperty{
			Name:        "Upsert_New_Test_Property_4",
			DisplayName: "Test Property 4",
			DataType:    "string",
			Description: "Test Property 4",
		}
		newSourceNodeKind = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Source_Kind",
			DisplayName:   "Upsert New Test Source Kind",
			Description:   "a source kind",
			IsDisplayKind: false,
			Icon:          "source",
			IconColor:     "green",
		}
		newEnvironmentNodeKind1 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Environment_Kind_1",
			DisplayName:   "Upsert New Test Environment Kind 1",
			Description:   "an environment kind",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		newEnvironmentNodeKind2 = model.GraphSchemaNodeKind{
			Name:          "Upsert_New_Test_Environment_Kind_2",
			DisplayName:   "Upsert New Test Environment Kind 2",
			Description:   "an environment kind",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		newEnvironment1 = model.GraphEnvironment{
			EnvironmentKind: newEnvironmentNodeKind1.Name,
			SourceKind:      newSourceNodeKind.Name,
			PrincipalKinds:  []string{newNodeKind1.Name, newNodeKind2.Name},
		}
		newEnvironment2 = model.GraphEnvironment{
			EnvironmentKind: newEnvironmentNodeKind2.Name,
			SourceKind:      newSourceNodeKind.Name,
			PrincipalKinds:  []string{newNodeKind3.Name, newNodeKind4.Name},
		}

		existingNodeKind1 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_1",
			DisplayName:   "Test Node Kind 1",
			Description:   "Test Node Kind 1",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind1 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		existingProperty1 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_1",
			DisplayName: "Property 1",
			DataType:    "string",
			Description: "Property 1",
		}
		existingNodeKind2 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "Test Node Kind 2",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind2 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		existingProperty2 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_2",
			DisplayName: "Property 2",
			DataType:    "string",
			Description: "Property 2",
		}
		existingNodeKind3 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "Test Node Kind 3",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind3 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		existingProperty3 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_3",
			DisplayName: "Property 3",
			DataType:    "string",
			Description: "Property 3",
		}
		existingNodeKind4 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "Test Node Kind 4",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind4 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Existing_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		existingProperty4 = model.GraphSchemaProperty{
			Name:        "Upsert_Existing_Test_Property_4",
			DisplayName: "Property 4",
			DataType:    "string",
			Description: "Property 4",
		}
		existingEnvironmentNodeKind1 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Environment_Kind_1",
			DisplayName:   "Environment Kind 1",
			Description:   "Environment Kind 1",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		existingEnvironmentNodeKind2 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Environment_Kind_2",
			DisplayName:   "Environment Kind 2",
			Description:   "Environment Kind 2",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		existingSourceKind1 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Existing_Test_Source_Kind_1",
			DisplayName:   "Upsert Existing Test Source Kind_1",
			Description:   "a source kind",
			IsDisplayKind: false,
			Icon:          "source",
			IconColor:     "yellow ",
		}
		existingEnvironment1 = model.GraphEnvironment{
			EnvironmentKind: existingEnvironmentNodeKind1.Name,
			SourceKind:      existingSourceKind1.Name,
			PrincipalKinds:  []string{existingNodeKind1.Name, existingNodeKind2.Name},
		}
		existingEnvironment2 = model.GraphEnvironment{
			EnvironmentKind: existingEnvironmentNodeKind2.Name,
			SourceKind:      existingSourceKind1.Name,
			PrincipalKinds:  []string{existingNodeKind3.Name, existingNodeKind4.Name},
		}

		updateNodeKind4 = model.GraphSchemaNodeKind{
			Name:          "Upsert_Update_Node_Kind_4",
			DisplayName:   "Node Kind 4",
			Description:   "Node Kind 4",
			IsDisplayKind: true,
			Icon:          "Desktop",
			IconColor:     "orange",
		}
		updateEdgeKind4 = model.GraphSchemaEdgeKind{
			Name:          "Upsert_Update_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		updateProperty4 = model.GraphSchemaProperty{
			Name:        "Upsert_Update_Property_4",
			DisplayName: "Property 4",
			DataType:    "string",
			Description: "Property 4",
		}
		updateEnvironment1 = model.GraphEnvironment{
			EnvironmentKind: existingEnvironmentNodeKind1.Name,
			SourceKind:      newSourceNodeKind.Name,
			PrincipalKinds:  []string{newNodeKind1.Name, existingNodeKind1.Name, updateNodeKind4.Name},
		}

		// Used for creating an existing graph schema
		existingNodeKinds = model.GraphSchemaNodeKinds{existingNodeKind1, existingNodeKind2, existingNodeKind3, existingNodeKind4,
			existingEnvironmentNodeKind1, existingEnvironmentNodeKind2, existingSourceKind1}
		existingEdgeKinds    = model.GraphSchemaEdgeKinds{existingEdgeKind1, existingEdgeKind2, existingEdgeKind3, existingEdgeKind4}
		existingProperties   = model.GraphSchemaProperties{existingProperty1, existingProperty2, existingProperty3, existingProperty4}
		existingEnvironments = model.GraphEnvironments{existingEnvironment1, existingEnvironment2}

		// Used in both args and want, in doing so the SchemaExtensionId will be propagated back to the want
		// value, this allows for equality comparisons of the SchemaExtensionId rather than just checking
		// if it exists
		newNodeKinds = model.GraphSchemaNodeKinds{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4, newEnvironmentNodeKind1,
			newEnvironmentNodeKind2, newSourceNodeKind}
		newEdgeKinds    = model.GraphSchemaEdgeKinds{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4}
		newProperties   = model.GraphSchemaProperties{newProperty1, newProperty2, newProperty3, newProperty4}
		newEnvironments = model.GraphEnvironments{newEnvironment1, newEnvironment2}

		updateProperties = model.GraphSchemaProperties{newProperty1, newProperty2, newProperty3,
			newProperty4, existingProperty1, updateProperty4}
		updateEdgeKinds = model.GraphSchemaEdgeKinds{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4,
			existingEdgeKind1, updateEdgeKind4}
		updateNodeKinds = model.GraphSchemaNodeKinds{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4,
			existingNodeKind1, existingSourceKind1, newEnvironmentNodeKind1, updateNodeKind4, newSourceNodeKind}
		updateEnvironments = model.GraphEnvironments{newEnvironment1, updateEnvironment1}
	)

	type fields struct {
		setup    func(*testing.T) model.GraphExtension
		teardown func(*testing.T, int32)
	}

	type args struct {
		ctx            context.Context
		graphExtension model.GraphExtension
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		want            bool
		wantErr         error
		wantGraphSchema model.GraphExtension
	}{
		{
			name: "success - create new OpenGraph extension without environments",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtension { return model.GraphExtension{} },
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx: testSuite.Context,
				graphExtension: model.GraphExtension{
					GraphSchemaExtension:  testExtension,
					GraphSchemaNodeKinds:  newNodeKinds,
					GraphSchemaEdgeKinds:  newEdgeKinds,
					GraphSchemaProperties: newProperties,
				},
			},
			wantErr: nil,
			want:    false,
			wantGraphSchema: model.GraphExtension{
				GraphSchemaExtension:  testExtension,
				GraphSchemaNodeKinds:  newNodeKinds,
				GraphSchemaEdgeKinds:  newEdgeKinds,
				GraphSchemaProperties: newProperties,
			},
		},
		{
			name: "success - create new OpenGraph extension with environments",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtension { return model.GraphExtension{} },
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx: testSuite.Context,
				graphExtension: model.GraphExtension{
					GraphSchemaExtension:  testExtension,
					GraphSchemaNodeKinds:  newNodeKinds,
					GraphSchemaEdgeKinds:  newEdgeKinds,
					GraphSchemaProperties: newProperties,
					GraphEnvironments:     newEnvironments,
				},
			},
			wantErr: nil,
			want:    false,
			wantGraphSchema: model.GraphExtension{
				GraphSchemaExtension:  testExtension,
				GraphSchemaNodeKinds:  newNodeKinds,
				GraphSchemaEdgeKinds:  newEdgeKinds,
				GraphSchemaProperties: newProperties,
				GraphEnvironments:     newEnvironments,
			},
		},
		{
			name: "success - update OpenGraph extension",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtension {
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
					for _, edgeKind := range existingEdgeKinds {
						_, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind.Name,
							createdExtension.ID, edgeKind.Description, edgeKind.IsTraversable)
						require.NoError(t, err)
					}
					for _, property := range existingProperties {
						_, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context,
							createdExtension.ID, property.Name, property.DisplayName, property.DataType, property.Description)
						require.NoError(t, err)
					}
					for _, environment := range existingEnvironments {
						var (
							envKind           model.Kind
							sourceKind        database.SourceKind
							schemaEnvironment model.SchemaEnvironment
						)
						envKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, environment.EnvironmentKind)
						require.NoError(t, err)
						sourceKindType := graph.StringKind(environment.SourceKind)
						err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(sourceKindType)
						require.NoError(t, err)
						sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, sourceKindType.String())
						require.NoError(t, err)
						schemaEnvironment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
							int32(envKind.ID), int32(sourceKind.ID))
						require.NoError(t, err)

						// create each principal kind
						var principalKind model.Kind
						for _, principalNodeKindName := range environment.PrincipalKinds {
							principalKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, principalNodeKindName)
							require.NoError(t, err)
							_, err = testSuite.BHDatabase.CreatePrincipalKind(testSuite.Context, schemaEnvironment.ID,
								int32(principalKind.ID))
							require.NoError(t, err)
						}
					}

					gotExistingGraphExtension, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtension.Name)
					require.NoError(t, err)
					return gotExistingGraphExtension
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
				ctx: testSuite.Context,
				graphExtension: model.GraphExtension{ // remove objects 2, 3 and update object 4
					GraphSchemaExtension:  testExtension,
					GraphSchemaProperties: updateProperties,
					GraphSchemaEdgeKinds:  updateEdgeKinds,
					GraphSchemaNodeKinds:  updateNodeKinds,
					GraphEnvironments:     updateEnvironments,
				},
			},
			wantErr: nil,
			want:    true,
			wantGraphSchema: model.GraphExtension{ // remove objects 2, 3 and update object 4
				GraphSchemaExtension:  testExtension,
				GraphSchemaProperties: updateProperties,
				GraphSchemaEdgeKinds:  updateEdgeKinds,
				GraphSchemaNodeKinds:  updateNodeKinds,
				GraphEnvironments:     updateEnvironments,
			},
		},
		{
			name: "fail - cannot modify a built-in extension",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtension {
					t.Helper()
					var builtInExtension = model.GraphSchemaExtension{
						Name:        "Upsert_BuiltIn_Extension",
						DisplayName: "Built-in Extension",
						Version:     "1.0.0",
						Namespace:   "TEST",
					}
					result := testSuite.DB.WithContext(testSuite.Context).Raw(fmt.Sprintf(`
						INSERT INTO %s (name, display_name, version, is_builtin, created_at, updated_at)
							  VALUES (?, ?, ?, TRUE, NOW(), NOW())
							  RETURNING id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at`,
						builtInExtension.TableName()), builtInExtension.Name, builtInExtension.DisplayName, builtInExtension.Version).Scan(&builtInExtension)
					require.NoError(t, result.Error)
					return model.GraphExtension{GraphSchemaExtension: builtInExtension}
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, "Upsert_BuiltIn_Extension")
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx: testSuite.Context,
				graphExtension: model.GraphExtension{
					GraphSchemaExtension: model.GraphSchemaExtension{
						Name:        "Upsert_BuiltIn_Extension",
						DisplayName: "Built-in Extension",
						Version:     "1.0.0",
					},
				},
			},
			wantErr: model.ErrGraphExtensionBuiltIn,
		},
		/*
			{
				name: "Success - Source kind auto-registers",
				setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
					t.Helper()
					ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
					require.NoError(t, err)

					return ext.ID
				},
				args: args{
					environments: []database.EnvironmentInput{
						{
							EnvironmentKindName: "Tag_Tier_Zero",
							SourceKindName:      "NewSource",
							PrincipalKinds:      []string{"Tag_Tier_Zero"},
						},
					},
				},
				assert: func(t *testing.T, db *database.BloodhoundDB) {
					t.Helper()

					sourceKind, err := db.GetSourceKindByName(context.Background(), "NewSource")
					assert.NoError(t, err)
					assert.Equal(t, graph.StringKind("NewSource"), sourceKind.Name)

					environments, err := db.GetEnvironments(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, 1, len(environments))
					assert.Equal(t, int32(sourceKind.ID), environments[0].SourceKindId)

					principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), environments[0].ID)
					assert.NoError(t, err)
					assert.Equal(t, 1, len(principalKinds))
				},
			},
			{
				name: "Success - Multiple environments with different source kinds",
				setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
					t.Helper()
					ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
					require.NoError(t, err)

					return ext.ID
				},
				args: args{
					environments: []database.EnvironmentInput{
						{
							EnvironmentKindName: "Tag_Tier_Zero",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"Tag_Tier_Zero"},
						},
						{
							EnvironmentKindName: "Tag_Owned",
							SourceKindName:      "NewSource",
							PrincipalKinds:      []string{"Tag_Owned"},
						},
					},
				},
				assert: func(t *testing.T, db *database.BloodhoundDB) {
					t.Helper()

					// Verify NewSource was auto-registered
					sourceKind, err := db.GetSourceKindByName(context.Background(), "NewSource")
					assert.NoError(t, err)
					assert.Equal(t, graph.StringKind("NewSource"), sourceKind.Name)

					environments, err := db.GetEnvironments(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, 2, len(environments), "Should have two environments")

					for _, env := range environments {
						principalKinds, err := db.GetPrincipalKindsByEnvironmentId(context.Background(), env.ID)
						assert.NoError(t, err)
						assert.Equal(t, 1, len(principalKinds), "Each environment should have one principal kind")
					}
				},
			},
			{
				name: "fail - First environment has invalid environment kind",
				setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
					t.Helper()
					ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
					require.NoError(t, err)

					return ext.ID
				},
				args: args{
					environments: []database.EnvironmentInput{
						{
							EnvironmentKindName: "NonExistent",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{},
						},
					},
				},
				expectedError: "environment kind 'NonExistent' not found",
				assert: func(t *testing.T, db *database.BloodhoundDB) {
					t.Helper()

					// Verify transaction rolled back - no environment created
					environments, err := db.GetEnvironments(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, 0, len(environments), "No environment should exist after rollback")
				},
			},
			{
				name: "fail - First environment has invalid principal kind",
				setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
					t.Helper()
					ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
					require.NoError(t, err)

					return ext.ID
				},
				args: args{
					environments: []database.EnvironmentInput{
						{
							EnvironmentKindName: "Tag_Tier_Zero",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"NonExistent"},
						},
					},
				},
				expectedError: "principal kind 'NonExistent' not found",
				assert: func(t *testing.T, db *database.BloodhoundDB) {
					t.Helper()

					// Verify transaction rolled back - no environment created
					environments, err := db.GetEnvironments(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, 0, len(environments), "No environment should exist after rollback")
				},
			},
			{
				name: "fail - Second environment fails, first should rollback",
				setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
					t.Helper()
					ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
					require.NoError(t, err)

					return ext.ID
				},
				args: args{
					environments: []database.EnvironmentInput{
						{
							EnvironmentKindName: "Tag_Tier_Zero",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"Tag_Tier_Zero"},
						},
						{
							EnvironmentKindName: "NonExistent",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{},
						},
					},
				},
				expectedError: "environment kind 'NonExistent' not found",
				assert: func(t *testing.T, db *database.BloodhoundDB) {
					t.Helper()

					// Verify complete transaction rollback - no environments created
					environments, err := db.GetEnvironments(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, 0, len(environments), "No environments should exist after rollback")
				},
			},
			{
				name: "fail - Second environment has invalid principal kind",
				setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
					t.Helper()
					ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
					require.NoError(t, err)

					return ext.ID
				},
				args: args{
					environments: []database.EnvironmentInput{
						{
							EnvironmentKindName: "Tag_Tier_Zero",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"Tag_Tier_Zero"},
						},
						{
							EnvironmentKindName: "Tag_Owned",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"NonExistent"},
						},
					},
				},
				expectedError: "principal kind 'NonExistent' not found",
				assert: func(t *testing.T, db *database.BloodhoundDB) {
					t.Helper()

					// Verify complete transaction rollback - no environments created
					environments, err := db.GetEnvironments(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, 0, len(environments), "No environments should exist after rollback")
				},
			},
			{
				name: "fail - Partial failure in first environment's principal kinds",
				setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
					t.Helper()
					ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0")
					require.NoError(t, err)

					return ext.ID
				},
				args: args{
					environments: []database.EnvironmentInput{
						{
							EnvironmentKindName: "Tag_Tier_Zero",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"Tag_Owned", "NonExistent"},
						},
					},
				},
				expectedError: "fail - principal kind 'NonExistent' not found",
				assert: func(t *testing.T, db *database.BloodhoundDB) {
					t.Helper()

					// Verify transaction rolled back - no environment created
					environments, err := db.GetEnvironments(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, 0, len(environments), "No environment should exist after rollback")
				},
			},

		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotGraphExtension model.GraphExtension
			fmt.Printf("\n\nTestName: %s\n\n", tt.name)
			existingGraphExtension := tt.fields.setup(t)

			if got, err = testSuite.BHDatabase.UpsertOpenGraphExtension(tt.args.ctx, tt.args.graphExtension); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				// there won't be a gotGraphSchemaExtension to delete if we expect an error
				gotGraphExtension.GraphSchemaExtension = existingGraphExtension.GraphSchemaExtension
			} else {
				require.NoError(t, err)
				// was it updated or not
				require.Equalf(t, tt.want, got, "UpsertOpenGraphExtension(%v, %v)", tt.args.ctx, tt.args.graphExtension)

				// Retrieve and compare upserted graph schema
				gotGraphExtension, err = testSuite.BHDatabase.GetGraphExtensionByName(testSuite.Context, testExtensionName)
				require.NoError(t, err)
				// fmt.Printf("\n\n%+v\n", gotGraphExtension.GraphEnvironments)
				compareGraphExtension(t, gotGraphExtension, tt.wantGraphSchema)
			}
			tt.fields.teardown(t, gotGraphExtension.GraphSchemaExtension.ID)
		})
	}
}

// TODO: Add tests for findings/remediations
func TestBloodhoundDB_GetGraphExtensionByName(t *testing.T) {

	var (
		err            error
		testSuite      = setupIntegrationTestSuite(t)
		graphExtension = model.GraphExtension{
			GraphSchemaExtension: model.GraphSchemaExtension{
				Name:        "TestGetGraphExtensionByName",
				DisplayName: "TestGetGraphExtensionByName",
				Version:     "1.0.0",
				Namespace:   "TGGEBN",
			},
			GraphSchemaProperties: model.GraphSchemaProperties{
				{
					Name:        "TGGEBN_Property_1",
					DisplayName: "Property_1",
					DataType:    "string",
					Description: "a property",
				},
				{
					Name:        "TGGEBN_Property_2",
					DisplayName: "Property_2",
					DataType:    "integer",
					Description: "a property",
				},
			},
			GraphSchemaEdgeKinds: model.GraphSchemaEdgeKinds{
				{
					Name:          "TGGEBN_EdgeKind_1",
					Description:   "an edge kind",
					IsTraversable: true,
				},
				{
					Name:          "TGGEBN_EdgeKind_2",
					Description:   "an edge kind",
					IsTraversable: true,
				},
			},
			GraphSchemaNodeKinds: model.GraphSchemaNodeKinds{
				{
					Name:          "TGGEBN_NodeKind_1",
					DisplayName:   "Node_1",
					Description:   "a node kind",
					IsDisplayKind: true,
					Icon:          "User",
					IconColor:     "blue",
				},
				{
					Name:          "TGGEBN_NodeKind_2",
					DisplayName:   "Node_2",
					Description:   "a node kind",
					IsDisplayKind: true,
					Icon:          "Desktop",
					IconColor:     "green",
				},
				{
					Name:          "TGGEBN_NodeKind_3",
					DisplayName:   "Node_3",
					Description:   "a node kind",
					IsDisplayKind: true,
					Icon:          "pasta",
					IconColor:     "yellow",
				},
				{
					Name:          "TGGEBN_NodeKind_4",
					DisplayName:   "Node_4",
					Description:   "a node kind",
					IsDisplayKind: true,
					Icon:          "apple",
					IconColor:     "red",
				},
				{
					Name:        "TGGEBN_EnvironmentKind_1",
					DisplayName: "Environment_1",
					Description: "a node environment kind",
				},
				{
					Name:        "TGGEBN_EnvironmentKind_2",
					DisplayName: "Environment_2",
					Description: "a node environment kind",
				},
				{
					Name:        "SourceKind",
					DisplayName: "SourceKind",
					Description: "the node source kind for an extension",
				},
			},
			GraphEnvironments: model.GraphEnvironments{
				{
					EnvironmentKind: "TGGEBN_EnvironmentKind_1",
					SourceKind:      "SourceKind",
					PrincipalKinds:  []string{"TGGEBN_NodeKind_1", "TGGEBN_NodeKind_2"},
				},
				{
					EnvironmentKind: "TGGEBN_EnvironmentKind_2",
					SourceKind:      "SourceKind",
					PrincipalKinds:  []string{"TGGEBN_NodeKind_3", "TGGEBN_NodeKind_4"},
				},
			},
			GraphFindings: model.GraphFindings{
				{
					Name:             "TTGGEBN_GraphFinding_1",
					DisplayName:      "Finding_1",
					SourceKind:       "SourceKind",
					RelationshipKind: "TGGEBN_EdgeKind_1",
					EnvironmentKind:  "TGGEBN_EnvironmentKind_1",
					Remediation: model.Remediation{
						ShortDescription: "a remediation",
						LongDescription:  "a detailed remediation",
						ShortRemediation: "do x",
						LongRemediation:  "do x to fix",
					},
				},
				{
					Name:             "TTGGEBN_GraphFinding_2",
					DisplayName:      "Finding_2",
					SourceKind:       "SourceKind",
					RelationshipKind: "TGGEBN_EdgeKind_2",
					EnvironmentKind:  "TGGEBN_EnvironmentKind_2",
					Remediation: model.Remediation{
						ShortDescription: "a remediation",
						LongDescription:  "a detailed remediation",
						ShortRemediation: "do y",
						LongRemediation:  "do y to fix",
					},
				},
			},
		}
	)
	defer teardownIntegrationTestSuite(t, &testSuite)

	type fields struct {
		setup    func(t *testing.T, extension model.GraphExtension)
		teardown func(t *testing.T, extensionId int32)
	}
	type args struct {
		ctx                context.Context
		argsGraphExtension model.GraphExtension
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		wantErr            error
		wantGraphExtension model.GraphExtension
	}{
		{
			name: "fail - non existent extension",
			fields: fields{
				setup:    func(t *testing.T, extension model.GraphExtension) {},
				teardown: func(t *testing.T, extensionId int32) {},
			},
			args: args{
				ctx: testSuite.Context,
				argsGraphExtension: model.GraphExtension{
					GraphSchemaExtension: model.GraphSchemaExtension{
						Name: "TestGetGraphExtensionByName",
					},
				},
			},
			wantErr:            database.ErrNotFound,
			wantGraphExtension: model.GraphExtension{},
		},
		{
			// schema extension exists but there are no nodes, edges, properties, environments or findings linked to the extension
			name: "success - existing but empty GraphSchemaExtension",
			fields: fields{
				setup: func(t *testing.T, extension model.GraphExtension) {
					t.Helper()
					_, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, extension.GraphSchemaExtension.Name, extension.GraphSchemaExtension.DisplayName,
						extension.GraphSchemaExtension.Version, extension.GraphSchemaExtension.Namespace)
					require.NoError(t, err)
				},
				teardown: func(t *testing.T, extensionId int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
					require.NoError(t, err)
				},
			},
			args: args{
				ctx: context.Background(),
				argsGraphExtension: model.GraphExtension{
					GraphSchemaExtension: model.GraphSchemaExtension{
						Name:        "TestGetGraphExtensionByName",
						DisplayName: "TestGetGraphExtensionByName",
						Version:     "1.0.0",
						Namespace:   "TestGetGraphExtensionByName",
					},
				},
			},
			wantGraphExtension: model.GraphExtension{
				GraphSchemaExtension: model.GraphSchemaExtension{
					Name:        "TestGetGraphExtensionByName",
					DisplayName: "TestGetGraphExtensionByName",
					Version:     "1.0.0",
					Namespace:   "TestGetGraphExtensionByName",
				},
			},
		},
		{
			name: "success - get graph extension",
			fields: fields{
				setup: func(t *testing.T, extension model.GraphExtension) {
					t.Helper()
					_, err = testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, extension)
					require.NoError(t, err)
				},
				teardown: func(t *testing.T, extensionId int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
					require.NoError(t, err)
				},
			},
			args: args{
				ctx:                testSuite.Context,
				argsGraphExtension: graphExtension,
			},
			wantGraphExtension: graphExtension,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got model.GraphExtension
			tt.fields.setup(t, tt.args.argsGraphExtension)

			if got, err = testSuite.BHDatabase.GetGraphExtensionByName(tt.args.ctx, tt.args.argsGraphExtension.GraphSchemaExtension.Name); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				compareGraphExtension(t, got, tt.wantGraphExtension)
				tt.fields.teardown(t, got.GraphSchemaExtension.ID)
			}
		})
	}
}

func compareGraphExtension(t *testing.T, got, want model.GraphExtension) {
	t.Helper()
	compareGraphSchemaExtension(t, got.GraphSchemaExtension, want.GraphSchemaExtension)
	compareMapOfGraphSchemaNodeKinds(t, got.GraphSchemaNodeKinds.ToMapKeyedOnName(), want.GraphSchemaNodeKinds.ToMapKeyedOnName())
	compareMapOfGraphSchemaEdgeKinds(t, got.GraphSchemaEdgeKinds.ToMapKeyedOnName(), want.GraphSchemaEdgeKinds.ToMapKeyedOnName())
	compareMapOfGraphSchemaProperties(t, got.GraphSchemaProperties.ToMapKeyedOnName(), want.GraphSchemaProperties.ToMapKeyedOnName())
	// environments share EnvKind and SourceKind as primary key
	var (
		gotMap  = make(map[string]model.GraphEnvironment, 0)
		wantMap = make(map[string]model.GraphEnvironment, 0)
	)
	for _, env := range got.GraphEnvironments {
		gotMap[env.EnvironmentKind+env.SourceKind] = env
	}
	for _, env := range want.GraphEnvironments {
		wantMap[env.EnvironmentKind+env.SourceKind] = env
	}
	compareMapOfGraphEnvironments(t, gotMap, wantMap)

	// can't use the compareGraphFindings func as UpsertOpenGraphExtension wont update the remediation's finding id
	// back in the want struct, so there's no way to know ahead of time what the value will be
	require.Equalf(t, len(want.GraphFindings), len(got.GraphFindings), "mismatched number of findings")
	for i := range got.GraphFindings {
		// Finding
		require.Greater(t, got.GraphFindings[i].ID, int32(0))
		require.Greaterf(t, got.GraphFindings[i].SchemaExtensionId, int32(0), "GraphFinding - graph schema extension id should be greater than 0")
		require.Equalf(t, want.GraphFindings[i].SourceKind, got.GraphFindings[i].SourceKind, "GraphFinding - source_kind mismatch")
		require.Equalf(t, want.GraphFindings[i].EnvironmentKind, got.GraphFindings[i].EnvironmentKind, "GraphFinding - environment_id mismatch")
		require.Equalf(t, want.GraphFindings[i].RelationshipKind, got.GraphFindings[i].RelationshipKind, "GraphFinding - relationship_kind_id mismatch")
		require.Equalf(t, want.GraphFindings[i].Name, got.GraphFindings[i].Name, "GraphFinding - name mismatch")
		require.Equalf(t, want.GraphFindings[i].DisplayName, got.GraphFindings[i].DisplayName, "GraphFinding - display name mismatch")

		// Remediation
		require.Equalf(t, want.GraphFindings[i].Remediation.ShortRemediation, got.GraphFindings[i].Remediation.ShortRemediation, "Remediation - short_remediation mismatch")
		require.Equalf(t, want.GraphFindings[i].Remediation.LongRemediation, got.GraphFindings[i].Remediation.LongRemediation, "Remediation - long_remediation mismatch")
		require.Equalf(t, want.GraphFindings[i].Remediation.ShortDescription, got.GraphFindings[i].Remediation.ShortDescription, "Remediation - short_description mismatch")
		require.Equalf(t, want.GraphFindings[i].Remediation.LongDescription, got.GraphFindings[i].Remediation.LongDescription, "Remediation - long_description mismatch")
	}
}
