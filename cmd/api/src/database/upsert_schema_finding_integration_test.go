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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_UpsertFinding(t *testing.T) {
	type args struct {
		sourceKindName, relationshipKindName, environmentKind, name, displayName string
	}
	tests := []struct {
		name          string
		setupData     func(t *testing.T, db *database.BloodhoundDB) int32 // Returns extensionId
		args          args
		assert        func(t *testing.T, db *database.BloodhoundDB, extensionId int32)
		expectedError string
	}{
		{
			name: "Success: Update existing finding - delete and re-create",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt", "Test", "v1.0.0", "Namespace")
				require.NoError(t, err)

				env, err := db.CreateEnvironment(context.Background(), ext.ID, 1, 1)
				require.NoError(t, err)

				// Create finding
				_, err = db.CreateSchemaRelationshipFinding(context.Background(), ext.ID, 1, env.ID, "Finding Name", "Finding Display Name")
				require.NoError(t, err)

				return ext.ID
			},
			args: args{
				sourceKindName:       "Base",
				relationshipKindName: "Tag_Tier_Zero",
				environmentKind:      "Tag_Tier_Zero",
				// Name triggers upsert so this needs to match the finding's name that we want to update
				name:        "Finding Name",
				displayName: "Updated Display Name",
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
				t.Helper()

				finding, err := db.GetSchemaRelationshipFindingByName(context.Background(), "Finding Name")
				require.NoError(t, err)

				assert.Equal(t, extensionId, finding.SchemaExtensionId)
				assert.Equal(t, "Finding Name", finding.Name)
				assert.Equal(t, "Updated Display Name", finding.DisplayName)
			},
		},
		{
			name: "Success: Create finding when none exists",
			setupData: func(t *testing.T, db *database.BloodhoundDB) int32 {
				t.Helper()
				ext, err := db.CreateGraphSchemaExtension(context.Background(), "TestExt2", "Test2", "v1.0.0", "Namespace")
				require.NoError(t, err)

				_, err = db.CreateEnvironment(context.Background(), ext.ID, 1, 1)
				require.NoError(t, err)

				// No finding created since we're testing the creation workflow
				return ext.ID
			},
			args: args{
				sourceKindName:       "Base",
				relationshipKindName: "Tag_Tier_Zero",
				environmentKind:      "Tag_Tier_Zero",
				name:                 "Finding",
				displayName:          "Finding Display Name",
			},
			assert: func(t *testing.T, db *database.BloodhoundDB, extensionId int32) {
				t.Helper()

				finding, err := db.GetSchemaRelationshipFindingByName(context.Background(), "Finding")
				require.NoError(t, err)

				assert.Equal(t, extensionId, finding.SchemaExtensionId)
				assert.Equal(t, "Finding", finding.Name)
				assert.Equal(t, "Finding Display Name", finding.DisplayName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			extensionId := tt.setupData(t, testSuite.BHDatabase)

			var findingResponse model.SchemaRelationshipFinding
			// Wrap the call in a transaction
			err := testSuite.BHDatabase.Transaction(context.Background(), func(tx *database.BloodhoundDB) error {
				finding, err := tx.UpsertFinding(
					context.Background(),
					extensionId,
					tt.args.sourceKindName,
					tt.args.relationshipKindName,
					tt.args.environmentKind,
					tt.args.name,
					tt.args.displayName,
				)
				findingResponse = finding
				return err
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, findingResponse.ID, "Finding should have been created/updated")
			}

			if tt.assert != nil {
				tt.assert(t, testSuite.BHDatabase, extensionId)
			}
		})
	}
}

func TestBloodhoundDB_GetGraphFindingsBySchemaExtensionId(t *testing.T) {
	var (
		testSuite = setupIntegrationTestSuite(t)

		extension = model.GraphSchemaExtension{
			Name:        "TestGraphExtension",
			DisplayName: "Test Graph Extension",
			Version:     "1.0.0",
			Namespace:   "Namespace",
		}

		edgeKind1 = model.GraphSchemaEdgeKind{
			Name:          "EdgeKind1",
			Description:   "an edge kind",
			IsTraversable: true,
		}
		edgeKind2 = model.GraphSchemaEdgeKind{
			Name:          "EdgeKind2",
			Description:   "an edge kind",
			IsTraversable: true,
		}
		environmentNodeKind1 = model.GraphSchemaNodeKind{
			Name:        "EnvironmentKind1",
			DisplayName: "Environment Kind 1",
			Description: "an environment kind",
		}
		environmentNodeKind2 = model.GraphSchemaNodeKind{
			Name:        "EnvironmentKind2",
			DisplayName: "Environment Kind 2",
			Description: "an environment kind",
		}
		sourceKind1 = model.GraphSchemaNodeKind{
			Name:        "SourceKind1",
			DisplayName: "Source Kind 1",
			Description: "a source kind",
		}

		finding1 = model.SchemaRelationshipFinding{
			Name:        "Finding1",
			DisplayName: "Finding 1",
		}
		finding2 = model.SchemaRelationshipFinding{
			Name:        "Finding2",
			DisplayName: "Finding 2",
		}

		remediation1 = model.Remediation{
			ShortDescription: "a remediation",
			LongDescription:  "a remediation but more detailed",
			ShortRemediation: "do x",
			LongRemediation:  "do x but better",
		}
		remediation2 = model.Remediation{
			ShortDescription: "a remediation",
			LongDescription:  "a remediation but more detailed",
			ShortRemediation: "do y",
			LongRemediation:  "do y but better",
		}
	)
	defer teardownIntegrationTestSuite(t, &testSuite)

	tests := []struct {
		name     string
		setup    func(t *testing.T) (int32, model.GraphFindings)
		teardown func(t *testing.T, extensionId int32)
		wantErr  error
	}{
		{
			name:     "fail - no findings for extension id",
			setup:    func(t *testing.T) (int32, model.GraphFindings) { return 132, model.GraphFindings{} },
			teardown: func(t *testing.T, extensionId int32) {},
			wantErr:  database.ErrNotFound,
		},
		{
			name: "fail - edge kind does not exist",
			setup: func(t *testing.T) (int32, model.GraphFindings) {
				t.Helper()
				var (
					createdExtension model.GraphSchemaExtension
					err              error

					createdGraphFindings = make(model.GraphFindings, 0)
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				return createdExtension.ID, createdGraphFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
			wantErr: database.ErrNotFound,
		},
		{
			name: "fail - environment kind does not exist",
			setup: func(t *testing.T) (int32, model.GraphFindings) {
				t.Helper()
				var (
					createdExtension model.GraphSchemaExtension
					err              error

					createdGraphFindings = make(model.GraphFindings, 0)
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				// create and retrieve finding edge kind
				_, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)

				return createdExtension.ID, createdGraphFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
			wantErr: database.ErrNotFound,
		},
		{
			name: "success - get 1 finding with no remediation",
			setup: func(t *testing.T) (int32, model.GraphFindings) {
				var (
					createdExtension            model.GraphSchemaExtension
					err                         error
					dawgsEnvKind, dawgsEdgeKind model.Kind
					createdEnvironmentNode      model.GraphSchemaNodeKind
					createdSourceKindNode       model.GraphSchemaNodeKind
					sourceKind                  database.SourceKind
					createdEnvironment          model.SchemaEnvironment
					createdEdgeKind             model.GraphSchemaEdgeKind
					createdFinding              model.SchemaRelationshipFinding
					createdGraphFindings        = make(model.GraphFindings, 0)
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				// create env and source node kinds
				createdEnvironmentNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdSourceKindNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err)

				// retrieve DAWGS env kind
				dawgsEnvKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode.Name)
				require.NoError(t, err)

				// register and retrieve source kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(createdSourceKindNode.Name))
				require.NoError(t, err)
				sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err)

				// create environment
				createdEnvironment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind.ID, int32(sourceKind.ID))
				require.NoError(t, err)

				// create and retrieve finding edge kind
				createdEdgeKind, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind.Name)
				require.NoError(t, err)

				// create findings
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind.ID, createdEnvironment.ID, finding1.Name, finding1.DisplayName)
				require.NoError(t, err)

				createdGraphFindings = append(createdGraphFindings, model.GraphFinding{
					ID:                createdFinding.ID,
					Name:              createdFinding.Name,
					SchemaExtensionId: createdFinding.SchemaExtensionId,
					DisplayName:       createdFinding.DisplayName,
					SourceKind:        sourceKind.Name.String(),
					RelationshipKind:  createdEdgeKind.Name,
					EnvironmentKind:   createdEnvironmentNode.Name,
				})
				return createdExtension.ID, createdGraphFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "success - get 1 finding and remediation",
			setup: func(t *testing.T) (int32, model.GraphFindings) {
				var (
					createdExtension            model.GraphSchemaExtension
					err                         error
					dawgsEnvKind, dawgsEdgeKind model.Kind
					createdEnvironmentNode      model.GraphSchemaNodeKind
					createdSourceKindNode       model.GraphSchemaNodeKind
					sourceKind                  database.SourceKind
					createdEnvironment          model.SchemaEnvironment
					createdEdgeKind             model.GraphSchemaEdgeKind
					createdFinding              model.SchemaRelationshipFinding
					createdRemediation          model.Remediation

					createdGraphFindings = make(model.GraphFindings, 0)
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				// create env and source node kinds
				createdEnvironmentNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdSourceKindNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err)

				// retrieve DAWGS env kind
				dawgsEnvKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode.Name)
				require.NoError(t, err)

				// register and retrieve source kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(sourceKind1.Name))
				require.NoError(t, err)
				sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err)

				// create environment
				createdEnvironment, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind.ID, int32(sourceKind.ID))
				require.NoError(t, err)

				// create and retrieve finding edge kind
				createdEdgeKind, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind.Name)
				require.NoError(t, err)

				// create findings
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind.ID, createdEnvironment.ID, finding1.Name, finding1.DisplayName)
				require.NoError(t, err)

				createdRemediation, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, createdFinding.ID,
					remediation1.ShortDescription, remediation1.LongDescription, remediation1.ShortRemediation, remediation1.LongRemediation)

				createdGraphFindings = append(createdGraphFindings, model.GraphFinding{
					ID:                createdFinding.ID,
					Name:              createdFinding.Name,
					SchemaExtensionId: createdFinding.SchemaExtensionId,
					DisplayName:       createdFinding.DisplayName,
					SourceKind:        sourceKind.Name.String(),
					RelationshipKind:  createdEdgeKind.Name,
					EnvironmentKind:   createdEnvironmentNode.Name,
					Remediation:       createdRemediation,
				})
				return createdExtension.ID, createdGraphFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "success - get multiple findings",
			setup: func(t *testing.T) (int32, model.GraphFindings) {
				var (
					err                                              error
					createdExtension                                 model.GraphSchemaExtension
					createdSourceKindNode                            model.GraphSchemaNodeKind
					sourceKind                                       database.SourceKind
					createdEdgeKind1, createdEdgeKind2               model.GraphSchemaEdgeKind
					createdFinding                                   model.SchemaRelationshipFinding
					createdGraphFindings                             = make(model.GraphFindings, 0)
					dawgsEnvKind1, dawgsEnvKind2                     model.Kind
					dawgsEdgeKind1, dawgsEdgeKind2                   model.Kind
					createdEnvironmentNode1, createdEnvironmentNode2 model.GraphSchemaNodeKind
					createdEnvironment1, createdEnvironment2         model.SchemaEnvironment
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				// create env and source node kinds
				createdEnvironmentNode1, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdEnvironmentNode2, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind2.Name,
					createdExtension.ID, environmentNodeKind2.DisplayName, environmentNodeKind2.Description,
					environmentNodeKind2.IsDisplayKind, environmentNodeKind2.Icon, environmentNodeKind2.IconColor)
				require.NoError(t, err)
				createdSourceKindNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err)

				// retrieve DAWGS env kinds
				dawgsEnvKind1, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode1.Name)
				require.NoError(t, err)
				dawgsEnvKind2, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode2.Name)
				require.NoError(t, err)

				// register and retrieve source kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(sourceKind1.Name))
				require.NoError(t, err)
				sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err)

				// create environments
				createdEnvironment1, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind1.ID, int32(sourceKind.ID))
				require.NoError(t, err)
				createdEnvironment2, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind2.ID, int32(sourceKind.ID))
				require.NoError(t, err)

				// create and retrieve finding edge kinds (dawgs id)
				createdEdgeKind1, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind1, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind1.Name)
				require.NoError(t, err)
				createdEdgeKind2, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind2.Name,
					createdExtension.ID, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind2, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind2.Name)
				require.NoError(t, err)

				// create finding 1
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind1.ID, createdEnvironment1.ID, finding1.Name, finding1.DisplayName)
				require.NoError(t, err)

				createdGraphFindings = append(createdGraphFindings, model.GraphFinding{
					ID:                createdFinding.ID,
					Name:              createdFinding.Name,
					SchemaExtensionId: createdFinding.SchemaExtensionId,
					DisplayName:       createdFinding.DisplayName,
					SourceKind:        sourceKind.Name.String(),
					RelationshipKind:  createdEdgeKind1.Name,
					EnvironmentKind:   createdEnvironmentNode1.Name,
				})

				// create finding 2
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind2.ID, createdEnvironment2.ID, finding2.Name, finding2.DisplayName)
				require.NoError(t, err)

				createdGraphFindings = append(createdGraphFindings, model.GraphFinding{
					ID:                createdFinding.ID,
					Name:              createdFinding.Name,
					SchemaExtensionId: createdFinding.SchemaExtensionId,
					DisplayName:       createdFinding.DisplayName,
					SourceKind:        sourceKind.Name.String(),
					RelationshipKind:  createdEdgeKind2.Name,
					EnvironmentKind:   createdEnvironmentNode2.Name,
				})

				return createdExtension.ID, createdGraphFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
		{
			name: "success - get multiple findings with remediations",
			setup: func(t *testing.T) (int32, model.GraphFindings) {
				var (
					err                                              error
					createdExtension                                 model.GraphSchemaExtension
					createdSourceKindNode                            model.GraphSchemaNodeKind
					sourceKind                                       database.SourceKind
					createdEdgeKind1, createdEdgeKind2               model.GraphSchemaEdgeKind
					createdFinding                                   model.SchemaRelationshipFinding
					createdGraphFindings                             = make(model.GraphFindings, 0)
					dawgsEnvKind1, dawgsEnvKind2                     model.Kind
					dawgsEdgeKind1, dawgsEdgeKind2                   model.Kind
					createdEnvironmentNode1, createdEnvironmentNode2 model.GraphSchemaNodeKind
					createdEnvironment1, createdEnvironment2         model.SchemaEnvironment
					createdRemediation                               model.Remediation
				)
				// create extension
				createdExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context,
					extension.Name, extension.DisplayName, extension.Version, extension.Namespace)
				require.NoError(t, err)

				// create env and source node kinds
				createdEnvironmentNode1, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind1.Name,
					createdExtension.ID, environmentNodeKind1.DisplayName, environmentNodeKind1.Description,
					environmentNodeKind1.IsDisplayKind, environmentNodeKind1.Icon, environmentNodeKind1.IconColor)
				require.NoError(t, err)
				createdEnvironmentNode2, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, environmentNodeKind2.Name,
					createdExtension.ID, environmentNodeKind2.DisplayName, environmentNodeKind2.Description,
					environmentNodeKind2.IsDisplayKind, environmentNodeKind2.Icon, environmentNodeKind2.IconColor)
				require.NoError(t, err)
				createdSourceKindNode, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, sourceKind1.Name,
					createdExtension.ID, sourceKind1.DisplayName, sourceKind1.Description, sourceKind1.IsDisplayKind,
					sourceKind1.Icon, sourceKind1.IconColor)
				require.NoError(t, err)

				// retrieve DAWGS env kinds
				dawgsEnvKind1, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode1.Name)
				require.NoError(t, err)
				dawgsEnvKind2, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEnvironmentNode2.Name)
				require.NoError(t, err)

				// register and retrieve source kind
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind(sourceKind1.Name))
				require.NoError(t, err)
				sourceKind, err = testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, createdSourceKindNode.Name)
				require.NoError(t, err)

				// create environments
				createdEnvironment1, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind1.ID, int32(sourceKind.ID))
				require.NoError(t, err)
				createdEnvironment2, err = testSuite.BHDatabase.CreateEnvironment(testSuite.Context, createdExtension.ID,
					dawgsEnvKind2.ID, int32(sourceKind.ID))
				require.NoError(t, err)

				// create and retrieve finding edge kinds (dawgs id)
				createdEdgeKind1, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name,
					createdExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind1, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind1.Name)
				require.NoError(t, err)
				createdEdgeKind2, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind2.Name,
					createdExtension.ID, edgeKind2.Description, edgeKind2.IsTraversable)
				require.NoError(t, err)
				dawgsEdgeKind2, err = testSuite.BHDatabase.GetKindByName(testSuite.Context, createdEdgeKind2.Name)
				require.NoError(t, err)

				// create finding 1
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind1.ID, createdEnvironment1.ID, finding1.Name, finding1.DisplayName)
				require.NoError(t, err)

				createdRemediation, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, createdFinding.ID,
					remediation1.ShortDescription, remediation1.LongDescription, remediation1.ShortRemediation, remediation1.LongRemediation)

				createdGraphFindings = append(createdGraphFindings, model.GraphFinding{
					ID:                createdFinding.ID,
					Name:              createdFinding.Name,
					SchemaExtensionId: createdFinding.SchemaExtensionId,
					DisplayName:       createdFinding.DisplayName,
					SourceKind:        sourceKind.Name.String(),
					RelationshipKind:  createdEdgeKind1.Name,
					EnvironmentKind:   createdEnvironmentNode1.Name,
					Remediation:       createdRemediation,
				})

				// create finding 2
				createdFinding, err = testSuite.BHDatabase.CreateSchemaRelationshipFinding(testSuite.Context,
					createdExtension.ID, dawgsEdgeKind2.ID, createdEnvironment2.ID, finding2.Name, finding2.DisplayName)
				require.NoError(t, err)

				createdRemediation, err = testSuite.BHDatabase.CreateRemediation(testSuite.Context, createdFinding.ID,
					remediation2.ShortDescription, remediation2.LongDescription, remediation2.ShortRemediation, remediation2.LongRemediation)

				createdGraphFindings = append(createdGraphFindings, model.GraphFinding{
					ID:                createdFinding.ID,
					Name:              createdFinding.Name,
					SchemaExtensionId: createdFinding.SchemaExtensionId,
					DisplayName:       createdFinding.DisplayName,
					SourceKind:        sourceKind.Name.String(),
					RelationshipKind:  createdEdgeKind2.Name,
					EnvironmentKind:   createdEnvironmentNode2.Name,
					Remediation:       createdRemediation,
				})

				return createdExtension.ID, createdGraphFindings
			},
			teardown: func(t *testing.T, extensionId int32) {
				t.Helper()
				err := testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, extensionId)
				require.NoError(t, err)

				_, err = testSuite.BHDatabase.GetGraphSchemaExtensionById(testSuite.Context, extensionId)
				require.EqualError(t, err, database.ErrNotFound.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extensionId, createdFindings := tt.setup(t)

			if got, err := testSuite.BHDatabase.GetGraphFindingsBySchemaExtensionId(context.Background(), extensionId); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				compareGraphFindings(t, got, createdFindings)
			}
			tt.teardown(t, extensionId)
		})
	}
}

func compareGraphFindings(t *testing.T, got, want model.GraphFindings) {
	t.Helper()

	require.Equalf(t, len(want), len(got), "mismatched number of findings")
	for i := range got {
		// Finding
		require.Greater(t, got[i].ID, int32(0))
		require.Greaterf(t, got[i].SchemaExtensionId, int32(0), "GraphFinding - graph schema extension id should be greater than 0")
		require.Equalf(t, want[i].SourceKind, got[i].SourceKind, "GraphFinding - source_kind mismatch")
		require.Equalf(t, want[i].EnvironmentKind, got[i].EnvironmentKind, "GraphFinding - environment_id mismatch")
		require.Equalf(t, want[i].RelationshipKind, got[i].RelationshipKind, "GraphFinding - relationship_kind_id mismatch")
		require.Equalf(t, want[i].Name, got[i].Name, "GraphFinding - name mismatch")
		require.Equalf(t, want[i].DisplayName, got[i].DisplayName, "GraphFinding - display name mismatch")

		// Remediation
		compareRemediation(t, got[i].Remediation, want[i].Remediation)
	}
}

func compareRemediation(t *testing.T, got, want model.Remediation) {
	t.Helper()

	require.Equalf(t, want.FindingID, got.FindingID, "Remediation - Finding ID mismatch - got: %+v, want: %+v", got, want)
	require.Equalf(t, want.ShortRemediation, got.ShortRemediation, "Remediation - short_remediation mismatch")
	require.Equalf(t, want.LongRemediation, got.LongRemediation, "Remediation - long_remediation mismatch")
	require.Equalf(t, want.ShortDescription, got.ShortDescription, "Remediation - short_description mismatch")
	require.Equalf(t, want.LongDescription, got.LongDescription, "Remediation - long_description mismatch")
}
