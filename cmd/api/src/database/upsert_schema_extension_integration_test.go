// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertOpenGraphExtension(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		err error
		got bool

		testExtensionName = "Test_Extension_Upsert_Test"
		testExtension     = model.ExtensionInput{
			Name:        testExtensionName,
			Version:     "1.0.0",
			DisplayName: "Test Extension",
			Namespace:   "Upsert",
		}
		newNodeKind1 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_1",
			DisplayName:   "Test Node Kind 1",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind1 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		newProperty1 = model.PropertyInput{
			Name:        "Upsert_New_Test_Property_1",
			DisplayName: "Test Property 1",
			DataType:    "string",
			Description: "Test Property 1",
		}
		newNodeKind2 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind2 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		newProperty2 = model.PropertyInput{
			Name:        "Upsert_New_Test_Property_2",
			DisplayName: "Test Property 2",
			DataType:    "string",
			Description: "Test Property 2",
		}
		newNodeKind3 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind3 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		newProperty3 = model.PropertyInput{
			Name:        "Upsert_New_Test_Property_3",
			DisplayName: "Test Property 3",
			DataType:    "string",
			Description: "Test Property 3",
		}
		newNodeKind4 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "blue",
		}
		newEdgeKind4 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		newProperty4 = model.PropertyInput{
			Name:        "Upsert_New_Test_Property_4",
			DisplayName: "Test Property 4",
			DataType:    "string",
			Description: "Test Property 4",
		}
		newSourceNodeKind = model.NodeInput{
			Name:          "Upsert_New_Test_Source_Kind",
			DisplayName:   "Upsert New Test Source Kind",
			Description:   "a source kind",
			IsDisplayKind: false,
			Icon:          "source",
			IconColor:     "green",
		}
		newEnvironmentNodeKind1 = model.NodeInput{
			Name:          "Upsert_New_Test_Environment_Kind_1",
			DisplayName:   "Upsert New Test Environment Kind 1",
			Description:   "an environment kind",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		newEnvironmentNodeKind2 = model.NodeInput{
			Name:          "Upsert_New_Test_Environment_Kind_2",
			DisplayName:   "Upsert New Test Environment Kind 2",
			Description:   "an environment kind",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		newEnvironment1 = model.EnvironmentInput{
			EnvironmentKindName: newEnvironmentNodeKind1.Name,
			SourceKindName:      newSourceNodeKind.Name,
			PrincipalKinds:      []string{newNodeKind1.Name, newNodeKind2.Name},
		}
		newEnvironment2 = model.EnvironmentInput{
			EnvironmentKindName: newEnvironmentNodeKind2.Name,
			SourceKindName:      newSourceNodeKind.Name,
			PrincipalKinds:      []string{newNodeKind3.Name, newNodeKind4.Name},
		}
		newFinding1 = model.RelationshipFindingInput{
			Name:                 "Upsert_New_Finding_1",
			EnvironmentKindName:  newEnvironmentNodeKind1.Name,
			SourceKindName:       newSourceNodeKind.Name,
			DisplayName:          "Finding 1",
			RelationshipKindName: newEdgeKind1.Name,
			RemediationInput: model.RemediationInput{
				ShortDescription: "a remediation",
				LongDescription:  "a remediation but longer",
				ShortRemediation: "do x",
				LongRemediation:  "do x but also y",
			},
		}
		newFinding2 = model.RelationshipFindingInput{
			Name:                 "Upsert_New_Finding_2",
			EnvironmentKindName:  newEnvironmentNodeKind1.Name,
			SourceKindName:       newSourceNodeKind.Name,
			DisplayName:          "Finding 2",
			RelationshipKindName: newEdgeKind2.Name,
			RemediationInput: model.RemediationInput{
				ShortDescription: "a remediation",
				LongDescription:  "a remediation but longer",
				ShortRemediation: "do x",
				LongRemediation:  "do x but also y",
			},
		}
		newFinding3 = model.RelationshipFindingInput{
			Name:                 "Upsert_New_Finding_3",
			EnvironmentKindName:  newEnvironmentNodeKind2.Name,
			SourceKindName:       newSourceNodeKind.Name,
			DisplayName:          "Finding 3",
			RelationshipKindName: newEdgeKind3.Name,
			RemediationInput: model.RemediationInput{
				ShortDescription: "a remediation",
				LongDescription:  "a remediation but longer",
				ShortRemediation: "do x",
				LongRemediation:  "do x but also y",
			},
		}

		existingNodeKind1 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Node_Kind_1",
			DisplayName:   "Test Node Kind 1",
			Description:   "Test Node Kind 1",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind1 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		existingProperty1 = model.PropertyInput{
			Name:        "Upsert_Existing_Test_Property_1",
			DisplayName: "Property 1",
			DataType:    "string",
			Description: "Property 1",
		}
		existingNodeKind2 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "Test Node Kind 2",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind2 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		existingProperty2 = model.PropertyInput{
			Name:        "Upsert_Existing_Test_Property_2",
			DisplayName: "Property 2",
			DataType:    "string",
			Description: "Property 2",
		}
		existingNodeKind3 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "Test Node Kind 3",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind3 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		existingProperty3 = model.PropertyInput{
			Name:        "Upsert_Existing_Test_Property_3",
			DisplayName: "Property 3",
			DataType:    "string",
			Description: "Property 3",
		}
		existingNodeKind4 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "Test Node Kind 4",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "red",
		}
		existingEdgeKind4 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		existingProperty4 = model.PropertyInput{
			Name:        "Upsert_Existing_Test_Property_4",
			DisplayName: "Property 4",
			DataType:    "string",
			Description: "Property 4",
		}
		existingEnvironmentNodeKind1 = model.NodeInput{
			Name:          "Upsert_Existing_Environment_Kind_1",
			DisplayName:   "Environment Kind 1",
			Description:   "Environment Kind 1",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		existingEnvironmentNodeKind2 = model.NodeInput{
			Name:          "Upsert_Existing_Environment_Kind_2",
			DisplayName:   "Environment Kind 2",
			Description:   "Environment Kind 2",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "green",
		}
		existingSourceKind1 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Source_Kind_1",
			DisplayName:   "Upsert Existing Test Source Kind_1",
			Description:   "a source kind",
			IsDisplayKind: false,
			Icon:          "source",
			IconColor:     "yellow",
		}
		existingEnvironment1 = model.EnvironmentInput{
			EnvironmentKindName: existingEnvironmentNodeKind1.Name,
			SourceKindName:      existingSourceKind1.Name,
			PrincipalKinds:      []string{existingNodeKind1.Name, existingNodeKind2.Name},
		}
		existingEnvironment2 = model.EnvironmentInput{
			EnvironmentKindName: existingEnvironmentNodeKind2.Name,
			SourceKindName:      existingSourceKind1.Name,
			PrincipalKinds:      []string{existingNodeKind3.Name, existingNodeKind4.Name},
		}
		existingFinding1 = model.RelationshipFindingInput{
			Name:                 "Upsert_Existing_Finding_1",
			EnvironmentKindName:  existingEnvironmentNodeKind1.Name,
			SourceKindName:       existingSourceKind1.Name,
			RelationshipKindName: existingEdgeKind1.Name,
			DisplayName:          "Existing Finding 1",
			RemediationInput: model.RemediationInput{
				ShortDescription: "A short description",
				LongDescription:  "A long description",
				ShortRemediation: "A short remediation",
				LongRemediation:  "A long remediation",
			},
		}
		existingFinding2 = model.RelationshipFindingInput{
			Name:                 "Upsert_Existing_Finding_2",
			EnvironmentKindName:  existingEnvironmentNodeKind2.Name,
			SourceKindName:       existingSourceKind1.Name,
			RelationshipKindName: existingEdgeKind2.Name,
			DisplayName:          "Existing Finding 2",
			RemediationInput: model.RemediationInput{
				ShortDescription: "A short description",
				LongDescription:  "A long description",
				ShortRemediation: "A long remediation",
				LongRemediation:  "A long remediation",
			},
		}

		updateNodeKind4 = model.NodeInput{
			Name:          "Upsert_Update_Node_Kind_4",
			DisplayName:   "Node Kind 4",
			Description:   "Node Kind 4",
			IsDisplayKind: true,
			Icon:          "Desktop",
			IconColor:     "orange",
		}
		updateEdgeKind4 = model.RelationshipInput{
			Name:          "Upsert_Update_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		updateProperty4 = model.PropertyInput{
			Name:        "Upsert_Update_Property_4",
			DisplayName: "Property 4",
			DataType:    "string",
			Description: "Property 4",
		}
		updateEnvironment1 = model.EnvironmentInput{
			EnvironmentKindName: existingEnvironmentNodeKind1.Name,
			SourceKindName:      newSourceNodeKind.Name,
			PrincipalKinds:      []string{newNodeKind1.Name, existingNodeKind1.Name, updateNodeKind4.Name},
		}
		updateFinding1 = model.RelationshipFindingInput{
			Name:                 "Upsert_Update_Finding_1",
			EnvironmentKindName:  existingEnvironmentNodeKind1.Name,
			SourceKindName:       newSourceNodeKind.Name,
			RelationshipKindName: updateEdgeKind4.Name,
			DisplayName:          "Update Finding 1",
			RemediationInput: model.RemediationInput{
				ShortDescription: "A short description",
				LongDescription:  "A long description",
				ShortRemediation: "A short remediation",
				LongRemediation:  "A long remediation",
			},
		}

		// Used for creating an existing graph schema
		existingNodeKinds = model.NodesInput{existingNodeKind1, existingNodeKind2, existingNodeKind3, existingNodeKind4,
			existingEnvironmentNodeKind1, existingEnvironmentNodeKind2, existingSourceKind1}
		existingEdgeKinds    = model.RelationshipsInput{existingEdgeKind1, existingEdgeKind2, existingEdgeKind3, existingEdgeKind4}
		existingProperties   = model.PropertiesInput{existingProperty1, existingProperty2, existingProperty3, existingProperty4}
		existingEnvironments = model.EnvironmentsInput{existingEnvironment1, existingEnvironment2}
		existingFindings     = model.RelationshipFindingsInput{existingFinding1, existingFinding2}

		// Used in both args and want, in doing so the SchemaExtensionId will be propagated back to the want
		// value, this allows for equality comparisons of the SchemaExtensionId rather than just checking
		// if it exists
		newNodeKinds = model.NodesInput{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4, newEnvironmentNodeKind1,
			newEnvironmentNodeKind2, newSourceNodeKind}
		newEdgeKinds    = model.RelationshipsInput{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4}
		newProperties   = model.PropertiesInput{newProperty1, newProperty2, newProperty3, newProperty4}
		newEnvironments = model.EnvironmentsInput{newEnvironment1, newEnvironment2}
		newFindings     = model.RelationshipFindingsInput{newFinding1, newFinding2, newFinding3}

		updateProperties = model.PropertiesInput{newProperty1, newProperty2, newProperty3,
			newProperty4, existingProperty1, updateProperty4}
		updateEdgeKinds = model.RelationshipsInput{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4,
			existingEdgeKind1, updateEdgeKind4}
		updateNodeKinds = model.NodesInput{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4,
			existingNodeKind1, existingSourceKind1, newEnvironmentNodeKind1, updateNodeKind4, newSourceNodeKind}
		updateEnvironments = model.EnvironmentsInput{newEnvironment1, updateEnvironment1}
		updateFindings     = model.RelationshipFindingsInput{newFinding1, updateFinding1}
	)

	type fields struct {
		setup    func(t *testing.T) int32
		teardown func(t *testing.T, id []int32) // extensions to delete
	}

	type args struct {
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
			name: "fail - duplicate node kinds",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: testExtension,
					NodeKindsInput: model.NodesInput{
						{
							Name: "DuplicateKind",
						},
						{
							Name: "DuplicateKind",
						},
					},
				},
			},
			wantErr: model.ErrDuplicateSchemaNodeKindName,
		},
		{
			name: "fail - duplicate relationship kinds",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: testExtension,
					NodeKindsInput: newNodeKinds,
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "DuplicateKind",
						},
						{
							Name: "DuplicateKind",
						},
					},
				},
			},
			wantErr: model.ErrDuplicateSchemaRelationshipKindName,
		},
		{
			name: "fail - duplicate properties",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					NodeKindsInput:         newNodeKinds,
					RelationshipKindsInput: newEdgeKinds,
					PropertiesInput: model.PropertiesInput{
						{
							Name: "DuplicateProperty",
						},
						{
							Name: "DuplicateProperty",
						},
					},
				},
			},
			wantErr: model.ErrDuplicateGraphSchemaExtensionPropertyName,
		},
		{
			name: "fail - cannot modify a built-in extension",
			fields: fields{
				setup: func(t *testing.T) int32 {
					t.Helper()
					var builtInExtension = model.GraphSchemaExtension{
						Name:        "Upsert_BuiltIn_Extension",
						DisplayName: "Built-in Extension",
						Version:     "1.0.0",
						Namespace:   "TEST",
					}
					result := testSuite.DB.WithContext(testSuite.Context).Raw(fmt.Sprintf(`
						INSERT INTO %s (name, display_name, version, is_builtin, namespace, created_at, updated_at)
							  VALUES (?, ?, ?, TRUE, ?, NOW(), NOW())
							  RETURNING id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at`,
						builtInExtension.TableName()), builtInExtension.Name, builtInExtension.DisplayName,
						builtInExtension.Version, builtInExtension.Namespace).Scan(&builtInExtension)
					require.NoError(t, result.Error)
					return builtInExtension.ID
				},
				teardown: func(t *testing.T, ids []int32) {
					t.Helper()
					for _, id := range ids {
						err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:        "Upsert_BuiltIn_Extension",
						DisplayName: "Built-in Extension",
						Version:     "1.0.0",
						Namespace:   "TEST",
					},
				},
			},
			wantErr: model.ErrGraphExtensionBuiltIn,
		},
		{
			name: "fail - first environment has invalid environment kind",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, ids []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					PropertiesInput:        newProperties,
					RelationshipKindsInput: newEdgeKinds,
					NodeKindsInput:         newNodeKinds,
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "NonExistent",
							SourceKindName:      newSourceNodeKind.Name,
							PrincipalKinds:      newEnvironment1.PrincipalKinds,
						},
					},
				},
			},
			wantErr: fmt.Errorf("environment kind 'NonExistent' not found"),
		},
		{
			name: "fail - first environment has invalid principal kind",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, ids []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					PropertiesInput:        newProperties,
					RelationshipKindsInput: newEdgeKinds,
					NodeKindsInput:         newNodeKinds,
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: newEnvironment1.EnvironmentKindName,
							SourceKindName:      newSourceNodeKind.Name,
							PrincipalKinds:      []string{"unknownKind"},
						},
						newEnvironment2,
					},
				},
			},
			want:    false,
			wantErr: fmt.Errorf("principal kind 'unknownKind' not found"),
		},
		{
			name: "fail - second environment fails, first should rollback",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, ids []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					PropertiesInput:        newProperties,
					RelationshipKindsInput: newEdgeKinds,
					NodeKindsInput:         newNodeKinds,
					EnvironmentsInput: model.EnvironmentsInput{
						newEnvironment1,
						{
							EnvironmentKindName: "NonExistent2",
							SourceKindName:      newSourceNodeKind.Name,
							PrincipalKinds:      newEnvironment2.PrincipalKinds,
						},
					},
				},
			},
			wantErr: fmt.Errorf("environment kind 'NonExistent2' not found"),
		},
		{
			name: "fail - second environment has invalid principal kind",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, ids []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					PropertiesInput:        newProperties,
					RelationshipKindsInput: newEdgeKinds,
					NodeKindsInput:         newNodeKinds,
					EnvironmentsInput: model.EnvironmentsInput{
						newEnvironment1,
						{
							EnvironmentKindName: newEnvironment2.EnvironmentKindName,
							SourceKindName:      newSourceNodeKind.Name,
							PrincipalKinds:      []string{"unknownKind"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("principal kind 'unknownKind' not found"),
		},
		{
			name: "fail - failure in second environment's latter principal kinds",
			fields: fields{
				setup:    func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, ids []int32) {},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					PropertiesInput:        newProperties,
					RelationshipKindsInput: newEdgeKinds,
					NodeKindsInput:         newNodeKinds,
					EnvironmentsInput: model.EnvironmentsInput{
						newEnvironment1,
						{
							EnvironmentKindName: newEnvironment2.EnvironmentKindName,
							SourceKindName:      newSourceNodeKind.Name,
							PrincipalKinds:      []string{newNodeKind1.Name, "unknownKind"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("principal kind 'unknownKind' not found"),
		},
		{
			name: "success - create new OpenGraph extension without environments",
			fields: fields{
				setup: func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()

					for _, id := range extensionIds {
						err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					NodeKindsInput:         newNodeKinds,
					RelationshipKindsInput: newEdgeKinds,
					PropertiesInput:        newProperties,
				},
			},
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:         testExtension,
				NodeKindsInput:         newNodeKinds,
				RelationshipKindsInput: newEdgeKinds,
				PropertiesInput:        newProperties,
			},
		},
		{
			name: "success - create full OpenGraph extension",
			fields: fields{
				setup: func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()

					for _, id := range extensionIds {
						err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:            testExtension,
					NodeKindsInput:            newNodeKinds,
					RelationshipKindsInput:    newEdgeKinds,
					PropertiesInput:           newProperties,
					EnvironmentsInput:         newEnvironments,
					RelationshipFindingsInput: newFindings,
				},
			},
			wantErr: nil,
			want:    false,
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:            testExtension,
				NodeKindsInput:            newNodeKinds,
				RelationshipKindsInput:    newEdgeKinds,
				PropertiesInput:           newProperties,
				EnvironmentsInput:         newEnvironments,
				RelationshipFindingsInput: newFindings,
			},
		},
		{
			name: "success - update full OpenGraph extension",
			fields: fields{
				setup: func(t *testing.T) int32 {
					t.Helper()
					// Create a graph schema and ensure it exists
					var (
						createdExtension model.GraphSchemaExtension
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
						err = testSuite.BHDatabase.UpsertSchemaEnvironmentWithPrincipalKinds(testSuite.Context, createdExtension.ID,
							environment.EnvironmentKindName, environment.SourceKindName, environment.PrincipalKinds)
						require.NoError(t, err)
					}
					// Create findings and remediations

					for _, finding := range existingFindings {
						var (
							schemaFinding model.SchemaRelationshipFinding
						)

						schemaFinding, err = testSuite.BHDatabase.UpsertFinding(testSuite.Context, createdExtension.ID,
							finding.SourceKindName, finding.RelationshipKindName, finding.EnvironmentKindName,
							finding.Name, finding.DisplayName)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, schemaFinding.ID, finding.RemediationInput.ShortDescription,
							finding.RemediationInput.LongDescription, finding.RemediationInput.ShortRemediation,
							finding.RemediationInput.LongRemediation)
						require.NoError(t, err)
					}

					return 0 // schema extension records will be deleted during upsert so no need to return extension for deletion
				},
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()
					for _, extensionId := range extensionIds {
						err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
						require.NoError(t, err)
						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:            testExtension,
					PropertiesInput:           updateProperties,
					RelationshipKindsInput:    updateEdgeKinds,
					NodeKindsInput:            updateNodeKinds,
					EnvironmentsInput:         updateEnvironments,
					RelationshipFindingsInput: updateFindings,
				},
			},

			wantErr: nil,
			want:    true,
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:            testExtension,
				PropertiesInput:           updateProperties,
				RelationshipKindsInput:    updateEdgeKinds,
				NodeKindsInput:            updateNodeKinds,
				EnvironmentsInput:         updateEnvironments,
				RelationshipFindingsInput: updateFindings,
			},
		}, {
			name: "success - insert new OpenGraph extension with one already present", // not update, two different extensions
			fields: fields{
				setup: func(t *testing.T) int32 {
					t.Helper()
					// Create a graph schema and ensure it exists
					var (
						createdExtension model.GraphSchemaExtension
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
						err = testSuite.BHDatabase.UpsertSchemaEnvironmentWithPrincipalKinds(testSuite.Context, createdExtension.ID,
							environment.EnvironmentKindName, environment.SourceKindName, environment.PrincipalKinds)
						require.NoError(t, err)
					}
					// Create findings and remediations

					for _, finding := range existingFindings {
						var (
							schemaFinding model.SchemaRelationshipFinding
						)

						schemaFinding, err = testSuite.BHDatabase.UpsertFinding(testSuite.Context, createdExtension.ID,
							finding.SourceKindName, finding.RelationshipKindName, finding.EnvironmentKindName,
							finding.Name, finding.DisplayName)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, schemaFinding.ID, finding.RemediationInput.ShortDescription,
							finding.RemediationInput.LongDescription, finding.RemediationInput.ShortRemediation,
							finding.RemediationInput.LongRemediation)
						require.NoError(t, err)
					}

					return createdExtension.ID
				},
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()
					for _, id := range extensionIds {
						err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:        "Upsert_Test_Extension_2",
						DisplayName: "Upsert Test Extension 2",
						Version:     "v1.0.0",
						Namespace:   "TWO",
					},
					PropertiesInput:           newProperties,
					RelationshipKindsInput:    newEdgeKinds,
					NodeKindsInput:            newNodeKinds,
					EnvironmentsInput:         newEnvironments,
					RelationshipFindingsInput: newFindings,
				},
			},
			want: false,
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput: model.ExtensionInput{
					Name:        "Upsert_Test_Extension_2",
					DisplayName: "Upsert Test Extension 2",
					Version:     "v1.0.0",
					Namespace:   "TWO",
				},
				PropertiesInput:           newProperties,
				RelationshipKindsInput:    newEdgeKinds,
				NodeKindsInput:            newNodeKinds,
				EnvironmentsInput:         newEnvironments,
				RelationshipFindingsInput: newFindings,
			},
		},
		{
			name: "success - environment source kind auto-registers",
			fields: fields{
				setup: func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()
					for _, id := range extensionIds {
						err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					PropertiesInput:        newProperties,
					RelationshipKindsInput: newEdgeKinds,
					NodeKindsInput:         newNodeKinds,
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: newEnvironment1.EnvironmentKindName,
							SourceKindName:      "UnregisteredSourceKind",
							PrincipalKinds:      newEnvironment1.PrincipalKinds,
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                 newFinding1.Name,
							DisplayName:          newFinding1.DisplayName,
							SourceKindName:       "UnregisteredSourceKind",
							RelationshipKindName: newFinding1.RelationshipKindName,
							EnvironmentKindName:  newEnvironment1.EnvironmentKindName,
							RemediationInput:     newFinding1.RemediationInput,
						},
					},
				},
			},
			want: false,
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:         testExtension,
				PropertiesInput:        newProperties,
				RelationshipKindsInput: newEdgeKinds,
				NodeKindsInput:         newNodeKinds,
				EnvironmentsInput: model.EnvironmentsInput{
					{
						EnvironmentKindName: newEnvironment1.EnvironmentKindName,
						SourceKindName:      "UnregisteredSourceKind",
						PrincipalKinds:      newEnvironment1.PrincipalKinds,
					},
				},
				RelationshipFindingsInput: model.RelationshipFindingsInput{
					{
						Name:                 newFinding1.Name,
						DisplayName:          newFinding1.DisplayName,
						SourceKindName:       "UnregisteredSourceKind",
						RelationshipKindName: newFinding1.RelationshipKindName,
						EnvironmentKindName:  newEnvironment1.EnvironmentKindName,
						RemediationInput:     newFinding1.RemediationInput,
					},
				},
			},
		},
		{
			name: "success - multiple environments with different source kinds",
			fields: fields{
				setup: func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()
					for _, id := range extensionIds {
						err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
					PropertiesInput:        newProperties,
					RelationshipKindsInput: newEdgeKinds,
					NodeKindsInput:         newNodeKinds,
					EnvironmentsInput: model.EnvironmentsInput{newEnvironment1,
						{
							EnvironmentKindName: newEnvironment2.EnvironmentKindName,
							SourceKindName:      "UnregisteredSourceKind",
							PrincipalKinds:      newEnvironment2.PrincipalKinds,
						},
					},
				},
			},
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:         testExtension,
				PropertiesInput:        newProperties,
				RelationshipKindsInput: newEdgeKinds,
				NodeKindsInput:         newNodeKinds,
				EnvironmentsInput: model.EnvironmentsInput{newEnvironment1,
					{
						EnvironmentKindName: newEnvironment2.EnvironmentKindName,
						SourceKindName:      "UnregisteredSourceKind",
						PrincipalKinds:      newEnvironment2.PrincipalKinds,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				setupGraphExtensionId, retrievedGraphExtensionId int32
				extensionToDelete                                = make([]int32, 0)
			)
			setupGraphExtensionId = tt.fields.setup(t)
			if setupGraphExtensionId != 0 {
				extensionToDelete = append(extensionToDelete, setupGraphExtensionId)
			}

			if got, err = testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, tt.args.graphExtension); tt.wantErr != nil {
				var totalRecords int
				require.ErrorContainsf(t, err, tt.wantErr.Error(), "unexpected error upserting open graph extension")
				// built-in extension test will still exist in the DB
				if tt.name != "fail - cannot modify a built-in extension" {
					_, totalRecords, err = testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context,
						model.Filters{"name": []model.Filter{{ // check to see if extension exists
							Operator:    model.Equals,
							Value:       tt.args.graphExtension.ExtensionInput.Name,
							SetOperator: model.FilterAnd,
						}}}, model.Sort{}, 0, 1)
					require.NoErrorf(t, err, "rollback was not completed and extension still exists: %s", tt.args.graphExtension.ExtensionInput.Name)
					require.Equalf(t, 0, totalRecords, "rollback was not completed and extension still exists: %s", tt.args.graphExtension.ExtensionInput.Name)
				}
			} else {
				require.NoError(t, err)
				// was it updated or not
				require.Equalf(t, tt.want, got, "UpsertOpenGraphExtension(%+v)", tt.args.graphExtension)
				retrievedGraphExtensionId = getAndCompareGraphExtension(t, testSuite.Context, testSuite.BHDatabase, tt.args.graphExtension)
			}

			if retrievedGraphExtensionId != 0 && retrievedGraphExtensionId != setupGraphExtensionId {
				extensionToDelete = append(extensionToDelete, retrievedGraphExtensionId)
			}

			tt.fields.teardown(t, extensionToDelete)
		})
	}
}

func getAndCompareGraphExtension(t *testing.T, testContext context.Context, db *database.BloodhoundDB, want model.GraphExtensionInput) int32 {
	t.Helper()
	var gotGraphExtension model.GraphSchemaExtension

	extensions, totalRecords, err := db.GetGraphSchemaExtensions(testContext,
		model.Filters{"name": []model.Filter{{ // check to see if extension exists
			Operator:    model.Equals,
			Value:       want.ExtensionInput.Name,
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
		sourceKind                   database.SourceKind
		dawgsPrincipalKind           model.Kind
		dawgsFindingRelationshipKind model.Kind
		dawgsFindingEnvironmentKind  model.Kind
		gotSchemaRelationshipFinding []model.SchemaRelationshipFinding
		gotRemediation               model.Remediation
		findingEnvironment           model.SchemaEnvironment
	)

	// Test Node Kinds

	gotNodeKinds, _, err = db.GetGraphSchemaNodeKinds(testContext,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	require.Equalf(t, len(gotNodeKinds), len(want.NodeKindsInput), "node kind - count mismatch")
	for idx, gotNodeKind := range gotNodeKinds {
		require.Greaterf(t, gotNodeKind.ID, int32(0), "NodeKindsInput - ID is invalid")
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
	gotSchemaEnvironments, err = db.GetEnvironmentsByExtensionId(testContext,
		gotGraphExtension.ID)
	require.NoError(t, err)

	require.Equalf(t, len(want.EnvironmentsInput), len(gotSchemaEnvironments), "environments - count mismatch")
	for idx, gotEnvironment := range gotSchemaEnvironments {
		require.Greaterf(t, gotEnvironment.ID, int32(0), "EnvironmentInput - ID is invalid")
		require.Equalf(t, gotGraphExtension.ID, gotEnvironment.SchemaExtensionId, "EnvironmentInput - SchemaExtensionId is invalid")
		require.Equalf(t, want.EnvironmentsInput[idx].EnvironmentKindName, gotEnvironment.EnvironmentKindName, "EnvironmentInput - EnvironmentKindName mismatch")
		sourceKind, err = db.GetSourceKindById(testContext, int(gotEnvironment.SourceKindId))
		require.NoError(t, err)
		require.Equalf(t, want.EnvironmentsInput[idx].SourceKindName, sourceKind.Name.String(), "EnvironmentInput - EnvironmentKindName mismatch")
		gotPrincipalKinds, err = db.GetPrincipalKindsByEnvironmentId(testContext, gotEnvironment.ID)
		require.NoError(t, err)
		require.Equalf(t, len(want.EnvironmentsInput[idx].PrincipalKinds), len(gotPrincipalKinds), "PrincipalKinds - count mismatch")
		for _, gotPrincipalKind := range gotPrincipalKinds {
			require.Equalf(t, gotEnvironment.ID, gotPrincipalKind.EnvironmentId, "PrincipalKind - EnvironmentId is invalid")
			dawgsPrincipalKind, err = db.GetKindById(testContext, gotPrincipalKind.PrincipalKind)
			require.NoError(t, err)
			require.Containsf(t, want.EnvironmentsInput[idx].PrincipalKinds, dawgsPrincipalKind.Name, "PrincipalKind - Name mismatch")
		}
	}

	// Test Findings
	gotSchemaRelationshipFinding, err = db.GetSchemaRelationshipFindingsBySchemaExtensionId(testContext, gotGraphExtension.ID)
	require.NoError(t, err)

	require.Equalf(t, len(want.RelationshipFindingsInput), len(gotSchemaRelationshipFinding), "mismatched number of findings")
	for i, finding := range gotSchemaRelationshipFinding {
		// Finding
		require.Greater(t, finding.ID, int32(0))
		require.Equalf(t, gotGraphExtension.ID, finding.SchemaExtensionId, "RelationshipFindingInput - graph schema extension id should be greater than 0")

		dawgsFindingRelationshipKind, err = db.GetKindById(testContext, finding.RelationshipKindId)
		require.NoError(t, err)
		require.Equalf(t, want.RelationshipFindingsInput[i].RelationshipKindName, dawgsFindingRelationshipKind.Name, "RelationshipFindingInput - relationship kind name mismatch")

		findingEnvironment, err = db.GetEnvironmentById(testContext, finding.EnvironmentId)
		require.NoError(t, err)
		dawgsFindingEnvironmentKind, err = db.GetKindById(testContext, findingEnvironment.EnvironmentKindId)
		require.NoError(t, err)
		require.Equalf(t, want.RelationshipFindingsInput[i].EnvironmentKindName, dawgsFindingEnvironmentKind.Name, "RelationshipFindingInput - environment kind name mismatch")

		require.Equalf(t, want.RelationshipFindingsInput[i].Name, finding.Name, "RelationshipFindingInput - name mismatch")
		require.Equalf(t, want.RelationshipFindingsInput[i].DisplayName, finding.DisplayName, "RelationshipFindingInput - display name mismatch")

		// Remediation
		gotRemediation, err = db.GetRemediationByFindingId(testContext, finding.ID)
		require.NoError(t, err)

		require.Equalf(t, want.RelationshipFindingsInput[i].RemediationInput.ShortRemediation, gotRemediation.ShortRemediation, "Remediation - short_remediation mismatch")
		require.Equalf(t, want.RelationshipFindingsInput[i].RemediationInput.LongRemediation, gotRemediation.LongRemediation, "Remediation - long_remediation mismatch")
		require.Equalf(t, want.RelationshipFindingsInput[i].RemediationInput.ShortDescription, gotRemediation.ShortDescription, "Remediation - short_description mismatch")
		require.Equalf(t, want.RelationshipFindingsInput[i].RemediationInput.LongDescription, gotRemediation.LongDescription, "Remediation - long_description mismatch")

	}

	return gotGraphExtension.ID
}
