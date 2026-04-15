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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertOpenGraphExtension(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		testExtensionName = "Test_Extension_Upsert_Test"
		testExtension     = model.ExtensionInput{
			Name:        testExtensionName,
			Version:     "1.0.0",
			DisplayName: "Test Extension",
			Namespace:   "Upsert",
		}
		testExtensionNoDisplayName = model.ExtensionInput{
			Name:      testExtensionName,
			Version:   "1.0.0",
			Namespace: "Upsert",
		}
		newNodeKind1 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_1",
			DisplayName:   "Test Node Kind 1",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "#2779F5",
		}
		newEdgeKind1 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		newNodeKind2 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "#2779F5",
		}
		newEdgeKind2 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		newNodeKind3 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "#2779F5",
		}
		newEdgeKind3 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		newNodeKind4 = model.NodeInput{
			Name:          "Upsert_New_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "#2779F5",
		}
		newEdgeKind4 = model.RelationshipInput{
			Name:          "Upsert_New_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		newSourceNodeKind = model.NodeInput{
			Name:          "Upsert_New_Test_Source_Kind",
			DisplayName:   "Upsert New Test Source Kind",
			Description:   "a source kind",
			IsDisplayKind: false,
			Icon:          "source",
			IconColor:     "#22b939",
		}
		newEnvironmentNodeKind1 = model.NodeInput{
			Name:          "Upsert_New_Test_Environment_Kind_1",
			DisplayName:   "Upsert New Test Environment Kind 1",
			Description:   "an environment kind",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "#22b939",
		}
		newEnvironmentNodeKind2 = model.NodeInput{
			Name:          "Upsert_New_Test_Environment_Kind_2",
			DisplayName:   "Upsert New Test Environment Kind 2",
			Description:   "an environment kind",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "#22b939",
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
			IconColor:     "#F52735",
		}
		existingEdgeKind1 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_1",
			Description:   "Test Edge Kind 1",
			IsTraversable: true,
		}
		existingNodeKind2 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Node_Kind_2",
			DisplayName:   "Test Node Kind 2",
			Description:   "Test Node Kind 2",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "#F52735",
		}
		existingEdgeKind2 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_2",
			Description:   "Test Edge Kind 2",
			IsTraversable: true,
		}
		existingNodeKind3 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Node_Kind_3",
			DisplayName:   "Test Node Kind 3",
			Description:   "Test Node Kind 3",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "#F52735",
		}
		existingEdgeKind3 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_3",
			Description:   "Test Edge Kind 3",
			IsTraversable: true,
		}
		existingNodeKind4 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Node_Kind_4",
			DisplayName:   "Test Node Kind 4",
			Description:   "Test Node Kind 4",
			IsDisplayKind: true,
			Icon:          "User",
			IconColor:     "#F52735",
		}
		existingEdgeKind4 = model.RelationshipInput{
			Name:          "Upsert_Existing_Test_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		existingEnvironmentNodeKind1 = model.NodeInput{
			Name:          "Upsert_Existing_Environment_Kind_1",
			DisplayName:   "Environment Kind 1",
			Description:   "Environment Kind 1",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "#22b939",
		}
		existingEnvironmentNodeKind2 = model.NodeInput{
			Name:          "Upsert_Existing_Environment_Kind_2",
			DisplayName:   "Environment Kind 2",
			Description:   "Environment Kind 2",
			IsDisplayKind: false,
			Icon:          "environment",
			IconColor:     "#22b939",
		}
		existingSourceKind1 = model.NodeInput{
			Name:          "Upsert_Existing_Test_Source_Kind_1",
			DisplayName:   "Upsert Existing Test Source Kind_1",
			Description:   "a source kind",
			IsDisplayKind: false,
			Icon:          "source",
			IconColor:     "#F5E027",
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
			IconColor:     "#F5A327",
		}
		updateEdgeKind4 = model.RelationshipInput{
			Name:          "Upsert_Update_Edge_Kind_4",
			Description:   "Test Edge Kind 4",
			IsTraversable: true,
		}
		updateEnvironment1 = model.EnvironmentInput{
			EnvironmentKindName: existingEnvironmentNodeKind1.Name,
			SourceKindName:      newSourceNodeKind.Name,
			PrincipalKinds:      []string{newNodeKind1.Name, existingNodeKind1.Name, updateNodeKind4.Name},
		}
		updateFinding1 = model.RelationshipFindingInput{
			Name:                 "Upsert_Update_Finding_1",
			EnvironmentKindName:  existingEnvironmentNodeKind1.Name,
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
		existingEnvironments = model.EnvironmentsInput{existingEnvironment1, existingEnvironment2}
		existingFindings     = model.RelationshipFindingsInput{existingFinding1, existingFinding2}

		// Used in both args and want, in doing so the SchemaExtensionId will be propagated back to the want
		// value, this allows for equality comparisons of the SchemaExtensionId rather than just checking
		// if it exists
		newNodeKinds = model.NodesInput{newNodeKind1, newNodeKind2, newNodeKind3, newNodeKind4, newEnvironmentNodeKind1,
			newEnvironmentNodeKind2, newSourceNodeKind}
		newEdgeKinds    = model.RelationshipsInput{newEdgeKind1, newEdgeKind2, newEdgeKind3, newEdgeKind4}
		newEnvironments = model.EnvironmentsInput{newEnvironment1, newEnvironment2}
		newFindings     = model.RelationshipFindingsInput{newFinding1, newFinding2, newFinding3}

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
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						assert.ErrorIs(t, err, model.ErrGraphExtensionBuiltIn)
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
			wantErr: fmt.Errorf("error retrieving environment kind 'NonExistent': entity not found"),
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
			wantErr: fmt.Errorf("error retrieving principal kind 'unknownKind': entity not found"),
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
			wantErr: fmt.Errorf("error retrieving environment kind 'NonExistent2': entity not found"),
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
			wantErr: fmt.Errorf("error retrieving principal kind 'unknownKind': entity not found"),
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
			wantErr: fmt.Errorf("error retrieving principal kind 'unknownKind': entity not found"),
		},
		{
			name: "success - create new OpenGraph extension without environments",
			fields: fields{
				setup: func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()

					for _, id := range extensionIds {
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
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
				},
			},
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput:         testExtension,
				NodeKindsInput:         newNodeKinds,
				RelationshipKindsInput: newEdgeKinds,
			},
		},
		{
			name: "success - create full OpenGraph extension",
			fields: fields{
				setup: func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()

					for _, id := range extensionIds {
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
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

					createdExtension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
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
					for _, environment := range existingEnvironments {
						_, err = testSuite.BHDatabase.CreateEnvironmentWithPrincipalKinds(testSuite.Context, createdExtension.ID, environment)
						require.NoError(t, err)
					}
					// Create findings and remediations

					for _, finding := range existingFindings {
						_, err = testSuite.BHDatabase.CreateFindingWithRemediation(testSuite.Context, createdExtension.ID, finding)
						require.NoError(t, err)
					}

					return 0 // schema extension records will be deleted during upsert so no need to return extension for deletion
				},
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()
					for _, extensionId := range extensionIds {
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
						require.NoError(t, err)
						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:            testExtension,
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

					createdExtension, err := testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
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
					for _, environment := range existingEnvironments {
						_, err = testSuite.BHDatabase.CreateEnvironmentWithPrincipalKinds(testSuite.Context, createdExtension.ID, environment)
						require.NoError(t, err)
					}
					// Create findings and remediations

					for _, finding := range existingFindings {
						_, err = testSuite.BHDatabase.CreateFindingWithRemediation(testSuite.Context, createdExtension.ID, finding)
						require.NoError(t, err)
					}

					return createdExtension.ID
				},
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()
					for _, id := range extensionIds {
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
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
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
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
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:         testExtension,
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
		{
			name: "success - name is used as displayname when displayname is not provided",
			fields: fields{
				setup: func(t *testing.T) int32 { return 0 },
				teardown: func(t *testing.T, extensionIds []int32) {
					t.Helper()
					for _, id := range extensionIds {
						err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, id)
						require.NoError(t, err)

						_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, id)
						require.Equal(t, database.ErrNotFound, err)
					}
				},
			},
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput:            testExtensionNoDisplayName,
					NodeKindsInput:            newNodeKinds,
					RelationshipKindsInput:    newEdgeKinds,
					EnvironmentsInput:         newEnvironments,
					RelationshipFindingsInput: newFindings,
				},
			},
			wantErr: nil,
			want:    false,
			wantGraphSchema: model.GraphExtensionInput{
				ExtensionInput: model.ExtensionInput{
					Name:        testExtensionNoDisplayName.Name,
					Version:     testExtensionNoDisplayName.Version,
					DisplayName: testExtensionNoDisplayName.Name,
					Namespace:   testExtensionNoDisplayName.Namespace,
				},
				NodeKindsInput:            newNodeKinds,
				RelationshipKindsInput:    newEdgeKinds,
				EnvironmentsInput:         newEnvironments,
				RelationshipFindingsInput: newFindings,
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

			if got, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, tt.args.graphExtension); tt.wantErr != nil {
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
	require.Equalf(t, want.ExtensionInput.GetDisplayName(), gotGraphExtension.DisplayName, "GraphSchemaExtensionInput - displayname mismatch")
	require.Equalf(t, want.ExtensionInput.Version, gotGraphExtension.Version, "GraphSchemaExtensionInput - version mismatch")
	require.Equalf(t, want.ExtensionInput.Namespace, gotGraphExtension.Namespace, "GraphSchemaExtensionInput - namespace mismatch")

	var (
		schemaIdFilter = model.Filter{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(gotGraphExtension.ID), 10),
			SetOperator: model.FilterAnd,
		}

		gotNodeKinds                  model.GraphSchemaNodeKinds
		gotRelationshipKinds          model.GraphSchemaRelationshipKinds
		gotSchemaEnvironments         []model.SchemaEnvironment
		gotPrincipalKinds             model.SchemaEnvironmentPrincipalKinds
		sourceKind                    model.Kind
		dawgsPrincipalKinds           []model.Kind
		dawgsFindingRelationshipKinds []model.Kind
		dawgsFindingEnvironmentKinds  []model.Kind
		gotSchemaRelationshipFinding  []model.SchemaFinding
		gotRemediation                model.Remediation
		findingEnvironment            model.SchemaEnvironment
	)

	// Test Node Kinds

	gotNodeKinds, _, err = db.GetGraphSchemaNodeKinds(testContext,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	require.Equalf(t, len(want.NodeKindsInput), len(gotNodeKinds), "node kind - count mismatch")
	wantNodeKindsByName := make(map[string]model.NodeInput, len(want.NodeKindsInput))
	for _, wantNodeKind := range want.NodeKindsInput {
		wantNodeKindsByName[wantNodeKind.Name] = wantNodeKind
	}
	for _, gotNodeKind := range gotNodeKinds {
		wantNodeKind, ok := wantNodeKindsByName[gotNodeKind.Name]
		require.Truef(t, ok, "GraphSchemaNodeKind(%v) - unexpected node kind returned", gotNodeKind.Name)
		require.Greaterf(t, gotNodeKind.ID, int32(0), "GraphSchemaNodeKind(%v) - ID is invalid", gotNodeKind.Name)
		require.Equalf(t, gotGraphExtension.ID, gotNodeKind.SchemaExtensionId, "GraphSchemaNodeKind(%v) - SchemaExtensionId is invalid", gotNodeKind.Name)
		require.Equalf(t, wantNodeKind.DisplayName, gotNodeKind.DisplayName, "GraphSchemaNodeKind(%v) - display_name mismatch", gotNodeKind.Name)
		require.Equalf(t, wantNodeKind.Description, gotNodeKind.Description, "GraphSchemaNodeKind(%v) - description mismatch", gotNodeKind.Name)
		require.Equalf(t, wantNodeKind.IsDisplayKind, gotNodeKind.IsDisplayKind, "GraphSchemaNodeKind(%v) - is_display_kind mismatch", gotNodeKind.Name)
		require.Equalf(t, wantNodeKind.Icon, gotNodeKind.Icon, "GraphSchemaNodeKind(%v) - icon mismatch", gotNodeKind.Name)
		require.Equalf(t, wantNodeKind.IconColor, gotNodeKind.IconColor, "GraphSchemaNodeKind(%v) - icon_color mismatch", gotNodeKind.Name)
		require.Falsef(t, gotNodeKind.CreatedAt.IsZero(), "GraphSchemaNodeKind(%v) - created_at is zero", gotNodeKind.Name)
		require.Falsef(t, gotNodeKind.UpdatedAt.IsZero(), "GraphSchemaNodeKind(%v) - updated_at is zero", gotNodeKind.Name)
		require.Falsef(t, gotNodeKind.DeletedAt.Valid, "GraphSchemaNodeKind(%v) - deleted_at is not null", gotNodeKind.Name)
	}

	// Test Custom Icons

	iconMap := make(map[string]model.CustomNodeKind)
	icons, err := db.GetCustomNodeKinds(testContext, nil)
	require.Nil(t, err)
	for _, icon := range icons {
		iconMap[icon.KindName] = icon
	}

	// confirm node icon definitions in the custom node kind table match the node definitions in the graph schema node kind table

	for _, gotNodeKind := range gotNodeKinds {
		if gotNodeKind.IsDisplayKind {
			// confirm display node kinds are in the icon map
			icon, ok := iconMap[gotNodeKind.Name]
			require.True(t, ok)
			require.Equal(t, gotNodeKind.Icon, icon.Config.Icon.Name)
			require.Equal(t, gotNodeKind.IconColor, icon.Config.Icon.Color)
		} else {
			// confirm non-display node kinds are not in the icon map
			_, ok := iconMap[gotNodeKind.Name]
			require.False(t, ok)
		}
	}
	// Test Relationship Kinds

	gotRelationshipKinds, _, err = db.GetGraphSchemaRelationshipKinds(testContext,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	require.Equalf(t, len(want.RelationshipKindsInput), len(gotRelationshipKinds), "relationship kind - count mismatch")
	wantRelKindsByName := make(map[string]model.RelationshipInput, len(want.RelationshipKindsInput))
	for _, wantRelKind := range want.RelationshipKindsInput {
		wantRelKindsByName[wantRelKind.Name] = wantRelKind
	}
	for _, gotRelationshipKind := range gotRelationshipKinds {
		wantRelKind, ok := wantRelKindsByName[gotRelationshipKind.Name]
		require.Truef(t, ok, "GraphSchemaRelationshipKind(%v) - unexpected relationship kind returned", gotRelationshipKind.Name)
		require.Greaterf(t, gotRelationshipKind.ID, int32(0), "GraphSchemaRelationshipKind(%v) - ID is invalid", gotRelationshipKind.Name)
		require.Equalf(t, gotGraphExtension.ID, gotRelationshipKind.SchemaExtensionId, "GraphSchemaRelationshipKind(%v) - SchemaExtensionId is invalid", gotRelationshipKind.Name)
		require.Equalf(t, wantRelKind.Description, gotRelationshipKind.Description, "GraphSchemaRelationshipKind(%v) - description mismatch", gotRelationshipKind.Name)
		require.Equalf(t, wantRelKind.IsTraversable, gotRelationshipKind.IsTraversable, "GraphSchemaRelationshipKind(%v) - is_traversable mismatch", gotRelationshipKind.Name)
	}

	// Test Environments
	gotSchemaEnvironments, err = db.GetEnvironmentsByExtensionId(testContext,
		gotGraphExtension.ID)
	require.NoError(t, err)

	require.Equalf(t, len(want.EnvironmentsInput), len(gotSchemaEnvironments), "environments - count mismatch")
	wantEnvironmentsByKindName := make(map[string]model.EnvironmentInput, len(want.EnvironmentsInput))
	for _, wantEnvironment := range want.EnvironmentsInput {
		wantEnvironmentsByKindName[wantEnvironment.EnvironmentKindName] = wantEnvironment
	}
	for _, gotEnvironment := range gotSchemaEnvironments {
		wantEnvironment, ok := wantEnvironmentsByKindName[gotEnvironment.EnvironmentKindName]
		require.Truef(t, ok, "SchemaEnvironment(%v) - unexpected environment returned", gotEnvironment.EnvironmentKindName)
		require.Greaterf(t, gotEnvironment.ID, int32(0), "SchemaEnvironment(%v) - ID is invalid", gotEnvironment.EnvironmentKindName)
		require.Equalf(t, gotGraphExtension.ID, gotEnvironment.SchemaExtensionId, "SchemaEnvironment(%v) - SchemaExtensionId is invalid", gotEnvironment.EnvironmentKindName)
		sourceKinds, err := db.GetKindsByIDs(testContext, gotEnvironment.SourceKindId)
		require.NoError(t, err)
		require.Len(t, sourceKinds, 1)
		sourceKind = sourceKinds[0]
		require.Equalf(t, wantEnvironment.SourceKindName, sourceKind.Name, "SchemaEnvironment(%v) - SourceKindName mismatch", gotEnvironment.EnvironmentKindName)
		gotPrincipalKinds, err = db.GetPrincipalKindsByEnvironmentId(testContext, gotEnvironment.ID)
		require.NoError(t, err)
		require.Equalf(t, len(wantEnvironment.PrincipalKinds), len(gotPrincipalKinds), "SchemaEnvironment(%v) - PrincipalKinds count mismatch", gotEnvironment.EnvironmentKindName)
		for _, gotPrincipalKind := range gotPrincipalKinds {
			require.Equalf(t, gotEnvironment.ID, gotPrincipalKind.EnvironmentId, "SchemaEnvironment(%v) - PrincipalKind EnvironmentId is invalid", gotEnvironment.EnvironmentKindName)
			dawgsPrincipalKinds, err = db.GetKindsByIDs(testContext, gotPrincipalKind.PrincipalKind)
			require.NoError(t, err)
			require.Len(t, dawgsPrincipalKinds, 1)
			require.Containsf(t, wantEnvironment.PrincipalKinds, dawgsPrincipalKinds[0].Name, "SchemaEnvironment(%v) - PrincipalKind name mismatch", gotEnvironment.EnvironmentKindName)
		}
	}

	// Test Findings
	gotSchemaRelationshipFinding, err = db.GetSchemaFindingsByExtensionId(testContext, gotGraphExtension.ID)
	require.NoError(t, err)

	require.Equalf(t, len(want.RelationshipFindingsInput), len(gotSchemaRelationshipFinding), "mismatched number of findings")
	wantFindingsByName := make(map[string]model.RelationshipFindingInput, len(want.RelationshipFindingsInput))
	for _, wantFinding := range want.RelationshipFindingsInput {
		wantFindingsByName[wantFinding.Name] = wantFinding
	}
	for _, finding := range gotSchemaRelationshipFinding {
		wantFinding, ok := wantFindingsByName[finding.Name]
		require.Truef(t, ok, "SchemaFinding(%v) - unexpected finding returned", finding.Name)

		// Finding
		require.Greaterf(t, finding.ID, int32(0), "SchemaFinding(%v) - ID is invalid", finding.Name)
		require.Equalf(t, gotGraphExtension.ID, finding.SchemaExtensionId, "SchemaFinding(%v) - SchemaExtensionId is invalid", finding.Name)

		dawgsFindingRelationshipKinds, err = db.GetKindsByIDs(testContext, finding.KindId)
		require.NoError(t, err)
		require.Len(t, dawgsFindingRelationshipKinds, 1)
		require.Equalf(t, wantFinding.RelationshipKindName, dawgsFindingRelationshipKinds[0].Name, "SchemaFinding(%v) - relationship kind name mismatch", finding.Name)

		findingEnvironment, err = db.GetEnvironmentById(testContext, finding.EnvironmentId)
		require.NoError(t, err)
		dawgsFindingEnvironmentKinds, err = db.GetKindsByIDs(testContext, findingEnvironment.EnvironmentKindId)
		require.NoError(t, err)
		require.Len(t, dawgsFindingEnvironmentKinds, 1)
		require.Equalf(t, wantFinding.EnvironmentKindName, dawgsFindingEnvironmentKinds[0].Name, "SchemaFinding(%v) - environment kind name mismatch", finding.Name)

		require.Equalf(t, wantFinding.DisplayName, finding.DisplayName, "SchemaFinding(%v) - display name mismatch", finding.Name)

		// Remediation
		gotRemediation, err = db.GetRemediationByFindingId(testContext, finding.ID)
		require.NoError(t, err)

		require.Equalf(t, wantFinding.RemediationInput.ShortRemediation, gotRemediation.ShortRemediation, "SchemaFinding(%v) - Remediation short_remediation mismatch", finding.Name)
		require.Equalf(t, wantFinding.RemediationInput.LongRemediation, gotRemediation.LongRemediation, "SchemaFinding(%v) - Remediation long_remediation mismatch", finding.Name)
		require.Equalf(t, wantFinding.RemediationInput.ShortDescription, gotRemediation.ShortDescription, "SchemaFinding(%v) - Remediation short_description mismatch", finding.Name)
		require.Equalf(t, wantFinding.RemediationInput.LongDescription, gotRemediation.LongDescription, "SchemaFinding(%v) - Remediation long_description mismatch", finding.Name)
	}

	return gotGraphExtension.ID
}
