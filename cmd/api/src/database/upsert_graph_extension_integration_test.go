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
	"strconv"
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
		newEdgeKind1 = model.GraphSchemaRelationshipKind{
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
		newEdgeKind2 = model.GraphSchemaRelationshipKind{
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
		newEdgeKind3 = model.GraphSchemaRelationshipKind{
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
		newEdgeKind4 = model.GraphSchemaRelationshipKind{
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
		newEnvironment1 = model.EnvironmentInput{
			EnvironmentKind: newEnvironmentNodeKind1.Name,
			SourceKind:      newSourceNodeKind.Name,
			PrincipalKinds:  []string{newNodeKind1.Name, newNodeKind2.Name},
		}
		newEnvironment2 = model.EnvironmentInput{
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
		existingEdgeKind1 = model.GraphSchemaRelationshipKind{
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
		existingEdgeKind2 = model.GraphSchemaRelationshipKind{
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
		existingEdgeKind3 = model.GraphSchemaRelationshipKind{
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
		existingEdgeKind4 = model.GraphSchemaRelationshipKind{
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
		existingEnvironment1 = model.EnvironmentInput{
			EnvironmentKind: existingEnvironmentNodeKind1.Name,
			SourceKind:      existingSourceKind1.Name,
			PrincipalKinds:  []string{existingNodeKind1.Name, existingNodeKind2.Name},
		}
		existingEnvironment2 = model.EnvironmentInput{
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
		updateEdgeKind4 = model.GraphSchemaRelationshipKind{
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
		updateEnvironment1 = model.EnvironmentInput{
			EnvironmentKind: existingEnvironmentNodeKind1.Name,
			SourceKind:      newSourceNodeKind.Name,
			PrincipalKinds:  []string{newNodeKind1.Name, existingNodeKind1.Name, updateNodeKind4.Name},
		}

		// Used for creating an existing graph schema
		existingNodeKinds = model.GraphSchemaNodeKinds{existingNodeKind1, existingNodeKind2, existingNodeKind3, existingNodeKind4,
			existingEnvironmentNodeKind1, existingEnvironmentNodeKind2, existingSourceKind1}
		existingEdgeKinds    = model.GraphSchemaRelationshipKinds{existingEdgeKind1, existingEdgeKind2, existingEdgeKind3, existingEdgeKind4}
		existingProperties   = model.GraphSchemaProperties{existingProperty1, existingProperty2, existingProperty3, existingProperty4}
		existingEnvironments = model.EnvironmentsInput{existingEnvironment1, existingEnvironment2}

		// Used in both args and want, in doing so the SchemaExtensionId will be propagated back to the want
		// value, this allows for equality comparisons of the SchemaExtensionId rather than just checking
		// if it exists
		newNodeKinds = model.GraphSchemaNodeKinds{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4, newEnvironmentNodeKind1,
			newEnvironmentNodeKind2, newSourceNodeKind}
		newEdgeKinds    = model.GraphSchemaRelationshipKinds{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4}
		newProperties   = model.GraphSchemaProperties{newProperty1, newProperty2, newProperty3, newProperty4}
		newEnvironments = model.EnvironmentsInput{newEnvironment1, newEnvironment2}

		updateProperties = model.GraphSchemaProperties{newProperty1, newProperty2, newProperty3,
			newProperty4, existingProperty1, updateProperty4}
		updateEdgeKinds = model.GraphSchemaRelationshipKinds{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4,
			existingEdgeKind1, updateEdgeKind4}
		updateNodeKinds = model.GraphSchemaNodeKinds{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4,
			existingNodeKind1, existingSourceKind1, newEnvironmentNodeKind1, updateNodeKind4, newSourceNodeKind}
		updateEnvironments = model.EnvironmentsInput{newEnvironment1, updateEnvironment1}
	)

	type fields struct {
		setup    func(*testing.T) model.GraphExtensionInput
		teardown func(*testing.T, int32)
	}

	type args struct {
		ctx            context.Context
		graphExtension model.GraphExtensionInput
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		want            bool
		wantErr         error
		wantGraphSchema model.GraphExtensionInput
	}{
		{
			name: "success - create new OpenGraph extension without environments",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtensionInput { return model.GraphExtensionInput{} },
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx: testSuite.Context,
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					NodeKindsInput:         newNodeKinds,
					RelationshipKindsInput: newEdgeKinds,
					PropertiesInput:        newProperties,
				},
			},
			wantErr: nil,
			want:    false,
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:         testExtension,
				NodeKindsInput:         newNodeKinds,
				RelationshipKindsInput: newEdgeKinds,
				PropertiesInput:        newProperties,
			},
		},
		{
			name: "success - create new OpenGraph extension with environments",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtensionInput { return model.GraphExtensionInput{} },
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx: testSuite.Context,
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					NodeKindsInput:         newNodeKinds,
					RelationshipKindsInput: newEdgeKinds,
					PropertiesInput:        newProperties,
					EnvironmentsInput:      newEnvironments,
				},
			},
			wantErr: nil,
			want:    false,
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:         testExtension,
				NodeKindsInput:         newNodeKinds,
				RelationshipKindsInput: newEdgeKinds,
				PropertiesInput:        newProperties,
				EnvironmentsInput:      newEnvironments,
			},
		},
		{
			name: "success - update OpenGraph extension",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtensionInput {
					t.Helper()
					// Create a graph schema and ensure it exists
					var (
						gotExistingGraphExtension model.GraphExtensionInput
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
						_, err = testSuite.BHDatabase.CreateGraphSchemaRelationshipKind(testSuite.Context, edgeKind.Name,
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

					_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx: testSuite.Context,
				graphExtension: model.GraphExtensionInput{ // remove objects 2, 3 and update object 4
					ExtensionInput:         testExtension,
					PropertiesInput:        updateProperties,
					RelationshipKindsInput: updateEdgeKinds,
					NodeKindsInput:         updateNodeKinds,
					EnvironmentsInput:      updateEnvironments,
				},
			},
			wantErr: nil,
			want:    true,
			wantGraphSchema: model.GraphExtensionInput{ // remove objects 2, 3 and update object 4
				ExtensionInput:         testExtension,
				PropertiesInput:        updateProperties,
				RelationshipKindsInput: updateEdgeKinds,
				NodeKindsInput:         updateNodeKinds,
				EnvironmentsInput:      updateEnvironments,
			},
		},
		{
			name: "fail - cannot modify a built-in extension",
			fields: fields{
				setup: func(t *testing.T) model.GraphExtensionInput {
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
					return model.GraphExtensionInput{ExtensionInput: builtInExtension}
				},
				teardown: func(t *testing.T, id int32) {
					t.Helper()
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx: testSuite.Context,
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.GraphSchemaExtension{
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
			var gotGraphExtension model.GraphSchemaExtension
			existingGraphExtension := tt.fields.setup(t)

			if got, err = testSuite.BHDatabase.UpsertOpenGraphExtension(tt.args.ctx, tt.args.graphExtension); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				// there won't be a gotGraphSchemaExtension to delete if we expect an error
				gotGraphExtension = existingGraphExtension
			} else {
				require.NoError(t, err)
				// was it updated or not
				require.Equalf(t, tt.want, got, "UpsertOpenGraphExtension(%v, %v)", tt.args.ctx, tt.args.graphExtension)

				getAndCompareGraphExtension(t, testSuite.Context, tt.args.graphExtension.ExtensionInput.Name, testSuite.BHDatabase, tt.args.graphExtension)
			}
			tt.fields.teardown(t, gotGraphExtension.ID)
		})
	}
}

func getAndCompareGraphExtension(t *testing.T, testContext context.Context, extensionName string, db *database.BloodhoundDB, want model.GraphExtensionInput) {
	t.Helper()
	var gotGraphExtension model.GraphSchemaExtension

	extensions, totalRecords, err := db.GetGraphSchemaExtensions(testContext,
		model.Filters{"name": []model.Filter{{ // check to see if extension exists
			Operator:    model.Equals,
			Value:       extensionName,
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 1)
	require.NoError(t, err)
	// Always expect to create the extension
	require.Equal(t, 1, totalRecords)
	gotGraphExtension = extensions[0]

	// Compare Extensions
	require.Equalf(t, want.ExtensionInput.Name, gotGraphExtension.Name, "GraphSchemaExtensionInput - name mismatch")
	require.Equalf(t, want.ExtensionInput.DisplayName, gotGraphExtension.DisplayName, "GraphSchemaExtensionInput - displayname mismatch")
	require.Equalf(t, want.ExtensionInput.Version, gotGraphExtension.Version, "GraphSchemaExtensionInput - version mismatch")
	require.Equalf(t, want.ExtensionInput.Namespace, gotGraphExtension.Namespace, "GraphSchemaExtensionInput - namespace mismatch")

	var (
		schemaIdFilter = model.Filter{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(gotGraphExtension.ID), 10),
			SetOperator: model.FilterAnd,
		}

		gotNodeKinds                 model.GraphSchemaNodeKinds
		gotRelationshipKinds         model.GraphSchemaRelationshipKinds
		gotProperties                model.GraphSchemaProperties
		gotSchemaEnvironments        []model.SchemaEnvironment
		gotPrincipalKinds            model.SchemaEnvironmentPrincipalKinds
		dawgsPrincipalKind           model.Kind
		gotSchemaRelationshipFinding []model.SchemaRelationshipFinding
		gotRemediation               model.Remediation
	)

	// Test Node Kinds

	gotNodeKinds, _, err = db.GetGraphSchemaNodeKinds(testContext,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	require.Equalf(t, len(gotNodeKinds), len(want.NodeKindsInput), "node kind - count mismatch")
	for idx, gotNodeKind := range gotNodeKinds {
		require.GreaterOrEqualf(t, gotNodeKind.ID, int32(0), "NodeKindsInput - ID is invalid")
		require.Equalf(t, gotGraphExtension.ID, gotNodeKinds[idx].SchemaExtensionId, "NodeKindsInput - SchemaExtensionId is invalid")
		require.Equalf(t, want.NodeKindsInput[idx].Name, gotNodeKind.Name, "GraphSchemaNodeKind(%v) - name mismatch", gotNodeKind.Name)
		require.Equalf(t, want.NodeKindsInput[idx].DisplayName, gotNodeKind.DisplayName, "GraphSchemaNodeKind(%v) - display_name mismatch", gotNodeKind.DisplayName)
		require.Equalf(t, want.NodeKindsInput[idx].Description, gotNodeKind.Description, "GraphSchemaNodeKind(%v) - description mismatch", gotNodeKind.Description)
		require.Equalf(t, want.NodeKindsInput[idx].IsDisplayKind, gotNodeKind.IsDisplayKind, "GraphSchemaNodeKind(%v) - is_display_kind mismatch", gotNodeKind.IsDisplayKind)
		require.Equalf(t, want.NodeKindsInput[idx].Icon, gotNodeKind.Icon, "GraphSchemaNodeKind(%v) - icon mismatch", gotNodeKind.Icon)
		require.Equalf(t, want.NodeKindsInput[idx].IconColor, gotNodeKind.IconColor, "GraphSchemaNodeKind(%v) - icon_color mismatch", gotNodeKind.IconColor)
		require.Equalf(t, false, gotNodeKind.CreatedAt.IsZero(), "GraphSchemaNodeKind(%v) - created_at is zero", gotNodeKind.CreatedAt.IsZero())
		require.Equalf(t, false, gotNodeKind.UpdatedAt.IsZero(), "GraphSchemaNodeKind(%v) - updated_at is zero", gotNodeKind.UpdatedAt.IsZero())
		require.Equalf(t, false, gotNodeKind.DeletedAt.Valid, "GraphSchemaNodeKind(%v) - deleted_at is not null", gotNodeKind.DeletedAt.Valid)
	}

	// Test Relationship Kinds

	gotRelationshipKinds, _, err = db.GetGraphSchemaRelationshipKinds(testContext,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	require.Equalf(t, len(gotRelationshipKinds), len(want.RelationshipKindsInput), "relationship kind - count mismatch")
	for idx, gotRelationshipKind := range gotRelationshipKinds {
		require.Greaterf(t, gotRelationshipKind.ID, int32(0), "RelationshipKindsInput - ID is invalid")
		require.Equalf(t, gotGraphExtension.ID, gotRelationshipKind.SchemaExtensionId, "RelationshipKindsInput - SchemaExtensionId is invalid")
		require.Equalf(t, want.RelationshipKindsInput[idx].Name, gotRelationshipKind.Name, "RelationshipKindsInput - Name mismatch")
		require.Equalf(t, want.RelationshipKindsInput[idx].Description, gotRelationshipKind.Description, "RelationshipKindsInput - description mismatch")
		require.Equalf(t, want.RelationshipKindsInput[idx].IsTraversable, gotRelationshipKind.IsTraversable, "RelationshipKindsInput - is_traversable mismatch")
	}

	// Test Properties

	gotProperties, _, err = db.GetGraphSchemaProperties(testContext,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	require.Equalf(t, len(gotProperties), len(want.PropertiesInput), "properties - count mismatch")
	for idx, gotProperty := range gotProperties {
		require.Greaterf(t, gotProperty.ID, int32(0), "PropertyInput - ID is invalid")
		require.Equalf(t, gotGraphExtension.ID, gotProperty.SchemaExtensionId, "PropertyInput - SchemaExtensionId is invalid")
		require.Equalf(t, want.PropertiesInput[idx].Name, gotProperty.Name, "PropertyInput - Name mismatch")
		require.Equalf(t, want.PropertiesInput[idx].Description, gotProperty.Description, "PropertyInput - description mismatch")
		require.Equalf(t, want.PropertiesInput[idx].DataType, gotProperty.DataType, "PropertyInput - DataType mismatch")
		require.Equalf(t, want.PropertiesInput[idx].DisplayName, gotProperty.DisplayName, "PropertyInput - display_name mismatch")
	}

	// Test Environments

	gotSchemaEnvironments, err = db.GetSchemaEnvironmentsByGraphSchemExtensionId(testContext,
		gotGraphExtension.ID)
	require.NoError(t, err)
	require.Equalf(t, len(want.EnvironmentsInput), len(gotSchemaEnvironments), "environments - count mismatch")
	for idx, gotEnvironment := range gotSchemaEnvironments {
		require.Greaterf(t, gotEnvironment.ID, int32(0), "EnvironmentInput - ID is invalid")
		require.Equalf(t, gotGraphExtension.ID, gotEnvironment.SchemaExtensionId, "EnvironmentInput - SchemaExtensionId is invalid")
		require.Equalf(t, want.EnvironmentsInput[idx].EnvironmentKind, gotEnvironment.EnvironmentKindName, "EnvironmentInput - EnvironmentKindName mismatch")
		require.Equalf(t, want.EnvironmentsInput[idx].SourceKind, gotEnvironment.SourceKindName, "EnvironmentInput - SourceKind is invalid")
		gotPrincipalKinds, err = db.GetPrincipalKindsByEnvironmentId(testContext, gotEnvironment.ID)
		require.NoError(t, err)
		for _, gotPrincipalKind := range gotPrincipalKinds {
			require.Equalf(t, len(want.EnvironmentsInput[idx].PrincipalKinds), len(gotPrincipalKinds), "PrincipalKinds - count mismatch")
			require.Equalf(t, gotEnvironment.ID, gotPrincipalKind.EnvironmentId, "PrincipalKind - EnvironmentId is invalid")
			dawgsPrincipalKind, err = db.GetKindById(testContext, gotPrincipalKind.PrincipalKind)
			require.NoError(t, err)
			require.Containsf(t, want.EnvironmentsInput[idx].PrincipalKinds, dawgsPrincipalKind.Name, "PrincipalKind - Name mismatch")
		}
	}

	// Test Findings

	gotSchemaRelationshipFinding, err = db.GetSchemaRelationshipFindingsBySchemaExtensionId(testContext, gotGraphExtension.ID)
	require.NoError(t, err)
	require.Equalf(t, len(want.FindingsInput), len(gotSchemaRelationshipFinding), "mismatched number of findings")
	for i, finding := range gotSchemaRelationshipFinding {
		// Finding
		require.Greater(t, finding.ID, int32(0))
		require.Equalf(t, gotGraphExtension.ID, finding.SchemaExtensionId, "FindingInput - graph schema extension id should be greater than 0")

		require.Equalf(t, want.FindingsInput[i].EnvironmentKind, finding.EnvironmentKind, "FindingInput - environment_id mismatch")
		require.Equalf(t, want.FindingsInput[i].RelationshipKind, finding.RelationshipKind, "FindingInput - relationship_kind_id mismatch")
		require.Equalf(t, want.FindingsInput[i].Name, finding.Name, "FindingInput - name mismatch")
		require.Equalf(t, want.FindingsInput[i].DisplayName, finding.DisplayName, "FindingInput - display name mismatch")

		// Remediation
		gotRemediation, err = db.GetRemediationByFindingId(testContext, finding.ID)
		require.NoError(t, err)

		require.Equalf(t, want.FindingsInput[i].Remediation.ShortRemediation, gotRemediation.ShortRemediation, "Remediation - short_remediation mismatch")
		require.Equalf(t, want.FindingsInput[i].Remediation.LongRemediation, gotRemediation.LongRemediation, "Remediation - long_remediation mismatch")
		require.Equalf(t, want.FindingsInput[i].Remediation.ShortDescription, gotRemediation.ShortDescription, "Remediation - short_description mismatch")
		require.Equalf(t, want.FindingsInput[i].Remediation.LongDescription, gotRemediation.LongDescription, "Remediation - long_description mismatch")
	}
}
