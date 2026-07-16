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
	"fmt"
	"strconv"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertExtensionDoesNotExist asserts that no extension with the given name exists in the DB.
func assertExtensionDoesNotExist(t *testing.T, testSuite IntegrationTestSuite, extensionName string) {
	t.Helper()
	_, totalRecords, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context,
		model.Filters{"name": []model.Filter{{
			Operator:    model.Equals,
			Value:       extensionName,
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 1)
	require.NoError(t, err)
	assert.Equalf(t, 0, totalRecords, "extension should not exist: %s", extensionName)
}

// assertCustomNodeKindPresent asserts that a custom_node_kinds row exists for the given kind name.
func assertCustomNodeKindPresent(t *testing.T, testSuite IntegrationTestSuite, kindName string) {
	t.Helper()
	icons, err := testSuite.BHDatabase.GetCustomNodeKinds(testSuite.Context)
	require.NoError(t, err)
	for _, icon := range icons {
		if icon.KindName == kindName {
			return
		}
	}
	assert.Failf(t, "custom node kind missing", "expected custom node kind %q to exist", kindName)
}

// assertCustomNodeKindAbsent asserts that no custom_node_kinds row exists for the given kind name.
func assertCustomNodeKindAbsent(t *testing.T, testSuite IntegrationTestSuite, kindName string) {
	t.Helper()
	icons, err := testSuite.BHDatabase.GetCustomNodeKinds(testSuite.Context)
	require.NoError(t, err)
	for _, icon := range icons {
		assert.NotEqualf(t, kindName, icon.KindName, "custom node kind %q should have been removed", kindName)
	}
}

func TestBloodhoundDB_UpsertOpenGraphExtension(t *testing.T) {
	t.Parallel()

	type testSetupData struct {
		input           model.GraphExtensionInput
		wantErrContains string
	}
	type testCase struct {
		name   string
		setup  func(t *testing.T, testSuite IntegrationTestSuite) testSetupData
		assert func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error)
	}

	tests := []testCase{
		{
			name: "error_-_duplicate_node_kinds",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{Name: "DupNodeKindExt", DisplayName: "Dup Node", Version: "1.0.0", Namespace: "DUP_NK"},
						NodeKindsInput: model.NodesInput{{Name: "DuplicateKind"}, {Name: "DuplicateKind"}},
					},
					wantErrContains: model.ErrDuplicateSchemaNodeKindName.Error(),
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_duplicate_relationship_kinds",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "DupRelKindExt", DisplayName: "Dup Rel", Version: "1.0.0", Namespace: "DUP_RK"},
						NodeKindsInput:         model.NodesInput{{Name: "DupRK_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000"}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "DuplicateKind"}, {Name: "DuplicateKind"}},
					},
					wantErrContains: model.ErrDuplicateSchemaRelationshipKindName.Error(),
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_cannot_modify_a_built_in_extension",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				var builtInExtension = model.GraphSchemaExtension{
					Name: "BuiltIn_Ext", DisplayName: "Built-in Extension", Version: "1.0.0", Namespace: "BUILTIN",
				}
				result := testSuite.DB.WithContext(testSuite.Context).Raw(fmt.Sprintf(`
					INSERT INTO %s (name, display_name, version, is_builtin, namespace, created_at, updated_at)
					VALUES (?, ?, ?, TRUE, ?, NOW(), NOW())
					RETURNING id, name, display_name, version, is_builtin, created_at, updated_at, deleted_at`,
					builtInExtension.TableName()), builtInExtension.Name, builtInExtension.DisplayName,
					builtInExtension.Version, builtInExtension.Namespace).Scan(&builtInExtension)
				require.NoError(t, result.Error)
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name: "BuiltIn_Ext", DisplayName: "Built-in Extension", Version: "1.0.0", Namespace: "BUILTIN",
						},
					},
					wantErrContains: model.ErrGraphExtensionBuiltIn.Error(),
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
			},
		},
		{
			name: "error_-_first_environment_has_invalid_environment_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "BadEnvKindExt", DisplayName: "Bad Env Kind", Version: "1.0.0", Namespace: "BAD_EK"},
						NodeKindsInput:         model.NodesInput{{Name: "BadEK_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000"}, {Name: "BadEK_EnvKind", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "BadEK_RK1", IsTraversable: true}},
						EnvironmentsInput:      model.EnvironmentsInput{{EnvironmentKindName: "NonExistent", SourceKindName: "BadEK_SrcKind", PrincipalKinds: []string{"BadEK_NK1"}}},
					},
					wantErrContains: "error retrieving environment kind 'NonExistent': entity not found",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_first_environment_has_invalid_principal_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "BadPK1Ext", DisplayName: "Bad PK1", Version: "1.0.0", Namespace: "BAD_PK1"},
						NodeKindsInput:         model.NodesInput{{Name: "BadPK1_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000"}, {Name: "BadPK1_EnvKind1", IsDisplayKind: false}, {Name: "BadPK1_EnvKind2", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "BadPK1_RK1", IsTraversable: true}},
						EnvironmentsInput: model.EnvironmentsInput{
							{EnvironmentKindName: "BadPK1_EnvKind1", SourceKindName: "BadPK1_SrcKind", PrincipalKinds: []string{"unknownKind"}},
							{EnvironmentKindName: "BadPK1_EnvKind2", SourceKindName: "BadPK1_SrcKind", PrincipalKinds: []string{"BadPK1_NK1"}},
						},
					},
					wantErrContains: "error retrieving principal kind 'unknownKind': entity not found",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_second_environment_has_invalid_environment_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "RollbackEnvExt", DisplayName: "Rollback Env", Version: "1.0.0", Namespace: "RB_ENV"},
						NodeKindsInput:         model.NodesInput{{Name: "RBEnv_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000"}, {Name: "RBEnv_NK2", DisplayName: "NK2", IsDisplayKind: true, Icon: "user", IconColor: "#000"}, {Name: "RBEnv_EnvKind1", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "RBEnv_RK1", IsTraversable: true}},
						EnvironmentsInput: model.EnvironmentsInput{
							{EnvironmentKindName: "RBEnv_EnvKind1", SourceKindName: "RBEnv_SrcKind", PrincipalKinds: []string{"RBEnv_NK1"}},
							{EnvironmentKindName: "NonExistent2", SourceKindName: "RBEnv_SrcKind", PrincipalKinds: []string{"RBEnv_NK2"}},
						},
					},
					wantErrContains: "error retrieving environment kind 'NonExistent2': entity not found",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_second_environment_has_invalid_principal_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "BadPK2Ext", DisplayName: "Bad PK2", Version: "1.0.0", Namespace: "BAD_PK2"},
						NodeKindsInput:         model.NodesInput{{Name: "BadPK2_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000"}, {Name: "BadPK2_EnvKind1", IsDisplayKind: false}, {Name: "BadPK2_EnvKind2", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "BadPK2_RK1", IsTraversable: true}},
						EnvironmentsInput: model.EnvironmentsInput{
							{EnvironmentKindName: "BadPK2_EnvKind1", SourceKindName: "BadPK2_SrcKind", PrincipalKinds: []string{"BadPK2_NK1"}},
							{EnvironmentKindName: "BadPK2_EnvKind2", SourceKindName: "BadPK2_SrcKind", PrincipalKinds: []string{"unknownKind"}},
						},
					},
					wantErrContains: "error retrieving principal kind 'unknownKind': entity not found",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_second_environments_has_unknown_latter_principal_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "BadPK2bExt", DisplayName: "Bad PK2b", Version: "1.0.0", Namespace: "BAD_PK2B"},
						NodeKindsInput:         model.NodesInput{{Name: "BadPK2b_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000"}, {Name: "BadPK2b_EnvKind1", IsDisplayKind: false}, {Name: "BadPK2b_EnvKind2", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "BadPK2b_RK1", IsTraversable: true}},
						EnvironmentsInput: model.EnvironmentsInput{
							{EnvironmentKindName: "BadPK2b_EnvKind1", SourceKindName: "BadPK2b_SrcKind", PrincipalKinds: []string{"BadPK2b_NK1"}},
							{EnvironmentKindName: "BadPK2b_EnvKind2", SourceKindName: "BadPK2b_SrcKind", PrincipalKinds: []string{"BadPK2b_NK1", "unknownKind"}},
						},
					},
					wantErrContains: "error retrieving principal kind 'unknownKind': entity not found",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_finding_has_invalid_relationship_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "BadFindingRKExt", DisplayName: "Bad Finding RK", Version: "1.0.0", Namespace: "BAD_FRK"},
						NodeKindsInput:         model.NodesInput{{Name: "BadFRK_NK1", DisplayName: "NK1", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}, {Name: "BadFRK_EnvKind1", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "BadFRK_RK1", Description: "rk", IsTraversable: true}},
						EnvironmentsInput:      model.EnvironmentsInput{{EnvironmentKindName: "BadFRK_EnvKind1", SourceKindName: "BadFRK_SrcKind", PrincipalKinds: []string{"BadFRK_NK1"}}},
						RelationshipFindingsInput: model.RelationshipFindingsInput{
							{Name: "BadFRK_Finding1", DisplayName: "Finding 1", RelationshipKindName: "NonExistentRelKind", EnvironmentKindName: "BadFRK_EnvKind1", RemediationInput: model.RemediationInput{ShortDescription: "sd", LongDescription: "ld", ShortRemediation: "sr", LongRemediation: "lr"}},
						},
					},
					wantErrContains: "error retrieving relationship kind 'NonExistentRelKind'",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_finding_has_invalid_environment_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "BadFindingEKExt", DisplayName: "Bad Finding EK", Version: "1.0.0", Namespace: "BAD_FEK"},
						NodeKindsInput:         model.NodesInput{{Name: "BadFEK_NK1", DisplayName: "NK1", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}, {Name: "BadFEK_EnvKind1", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "BadFEK_RK1", Description: "rk", IsTraversable: true}},
						EnvironmentsInput:      model.EnvironmentsInput{{EnvironmentKindName: "BadFEK_EnvKind1", SourceKindName: "BadFEK_SrcKind", PrincipalKinds: []string{"BadFEK_NK1"}}},
						RelationshipFindingsInput: model.RelationshipFindingsInput{
							{Name: "BadFEK_Finding1", DisplayName: "Finding 1", RelationshipKindName: "BadFEK_RK1", EnvironmentKindName: "NonExistentEnvKind", RemediationInput: model.RemediationInput{ShortDescription: "sd", LongDescription: "ld", ShortRemediation: "sr", LongRemediation: "lr"}},
						},
					},
					wantErrContains: "error retrieving environment kind 'NonExistentEnvKind'",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "success_-_create_new_opengraph_extension_without_environments",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "NoEnvExt", DisplayName: "No Env Extension", Version: "1.0.0", Namespace: "NO_ENV"},
						NodeKindsInput:         model.NodesInput{{Name: "NoEnv_NK1", DisplayName: "NK1", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "NoEnv_RK1", Description: "rk", IsTraversable: true}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
			},
		},
		{
			name: "success_-_create_full_opengraph_extension",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				nodeKinds := model.NodesInput{
					{
						Name:          "Full_NK1",
						DisplayName:   "NK1",
						Description:   "nk",
						IsDisplayKind: true,
						Icon:          "user",
						IconColor:     "#2779F5",
						Info: model.KindInfoInputs{
							{InfoKey: "overview", Title: "Overview", Position: 1, Content: []byte(`{"markdown": {"content": "# Node Kind 1 Overview"}}`)},
							{InfoKey: "details", Title: "Details", Position: 2, Content: []byte(`{"markdown": {"content": "Details about NK1"}}`)},
						},
					},
					{Name: "Full_NK2", DisplayName: "NK2", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"},
					{Name: "Full_EnvKind1", DisplayName: "Env Kind 1", Description: "env", IsDisplayKind: false},
				}
				edgeKinds := model.RelationshipsInput{
					{
						Name:          "Full_RK1",
						Description:   "rk",
						IsTraversable: true,
						Info: model.KindInfoInputs{
							{InfoKey: "usage", Title: "Usage", Position: 1, Content: []byte(`{"markdown": {"content": "# How to use this relationship"}}`)},
						},
					},
				}
				environments := model.EnvironmentsInput{
					{EnvironmentKindName: "Full_EnvKind1", SourceKindName: "Full_SrcKind", PrincipalKinds: []string{"Full_NK1", "Full_NK2"}},
				}
				findings := model.RelationshipFindingsInput{
					{Name: "Full_Finding1", DisplayName: "Finding 1", RelationshipKindName: "Full_RK1", EnvironmentKindName: "Full_EnvKind1", RemediationInput: model.RemediationInput{ShortDescription: "sd", LongDescription: "ld", ShortRemediation: "sr", LongRemediation: "lr"}},
				}
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:            model.ExtensionInput{Name: "FullCreateExt", DisplayName: "Full Create", Version: "1.0.0", Namespace: "FULL_CREATE"},
						NodeKindsInput:            nodeKinds,
						RelationshipKindsInput:    edgeKinds,
						EnvironmentsInput:         environments,
						RelationshipFindingsInput: findings,
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
			},
		},
		{
			name: "success_-_update_full_opengraph_extension",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				extensionInput := model.ExtensionInput{Name: "UpdateFullExt", DisplayName: "Update Full", Version: "1.0.0", Namespace: "UPD_FULL"}
				existingNodeKinds := model.NodesInput{
					{
						Name:          "UpdFull_ExNK1",
						DisplayName:   "ExNK1",
						Description:   "ex",
						IsDisplayKind: true,
						Icon:          "User",
						IconColor:     "#F52735",
						Info: model.KindInfoInputs{
							{InfoKey: "overview", Title: "Overview", Position: 1, Content: []byte(`{"markdown": {"content": "# Initial Overview"}}`)},
							{InfoKey: "details", Title: "Details", Position: 2, Content: []byte(`{"markdown": {"content": "Initial details"}}`)},
						},
					},
					{Name: "UpdFull_ExNK2", DisplayName: "ExNK2", Description: "ex", IsDisplayKind: true, Icon: "User", IconColor: "#F52735"},
					{Name: "UpdFull_ExEnvKind1", DisplayName: "Ex Env 1", Description: "env", IsDisplayKind: false},
				}
				existingEdgeKinds := model.RelationshipsInput{
					{
						Name:          "UpdFull_ExRK1",
						Description:   "ex rk",
						IsTraversable: true,
						Info: model.KindInfoInputs{
							{InfoKey: "usage", Title: "Usage", Position: 1, Content: []byte(`{"markdown": {"content": "# Initial usage"}}`)},
						},
					},
				}
				existingEnvs := model.EnvironmentsInput{
					{EnvironmentKindName: "UpdFull_ExEnvKind1", SourceKindName: "UpdFull_ExSrcKind", PrincipalKinds: []string{"UpdFull_ExNK1"}},
				}
				existingFindings := model.RelationshipFindingsInput{
					{Name: "UpdFull_ExFinding1", DisplayName: "Ex Finding 1", RelationshipKindName: "UpdFull_ExRK1", EnvironmentKindName: "UpdFull_ExEnvKind1", RemediationInput: model.RemediationInput{ShortDescription: "sd", LongDescription: "ld", ShortRemediation: "sr", LongRemediation: "lr"}},
				}
				_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, model.GraphExtensionInput{
					ExtensionInput:            extensionInput,
					NodeKindsInput:            existingNodeKinds,
					RelationshipKindsInput:    existingEdgeKinds,
					EnvironmentsInput:         existingEnvs,
					RelationshipFindingsInput: existingFindings,
				})
				require.NoError(t, err)

				// Build the update input: keeps some existing, adds new, drops stale
				// Also tests Info reconciliation: update overview, delete details, add examples
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: extensionInput,
						NodeKindsInput: model.NodesInput{
							{
								Name:          "UpdFull_ExNK1",
								DisplayName:   "ExNK1 Updated",
								Description:   "updated",
								IsDisplayKind: true,
								Icon:          "User",
								IconColor:     "#F52735",
								Info: model.KindInfoInputs{
									{InfoKey: "overview", Title: "Updated Overview", Position: 1, Content: []byte(`{"markdown": {"content": "# Updated Overview"}}`)},
									{InfoKey: "examples", Title: "Examples", Position: 3, Content: []byte(`{"markdown": {"content": "# New Examples"}}`)},
								},
							},
							{Name: "UpdFull_NewNK3", DisplayName: "NewNK3", Description: "new", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"},
							{Name: "UpdFull_ExEnvKind1", DisplayName: "Ex Env 1", Description: "env", IsDisplayKind: false},
						},
						RelationshipKindsInput: model.RelationshipsInput{
							{
								Name:          "UpdFull_ExRK1",
								Description:   "ex rk updated",
								IsTraversable: true,
								Info: model.KindInfoInputs{
									{InfoKey: "usage", Title: "Updated Usage", Position: 1, Content: []byte(`{"markdown": {"content": "# Updated usage"}}`)},
								},
							},
							{Name: "UpdFull_NewRK2", Description: "new rk", IsTraversable: true},
						},
						EnvironmentsInput: model.EnvironmentsInput{
							{EnvironmentKindName: "UpdFull_ExEnvKind1", SourceKindName: "UpdFull_NewSrcKind", PrincipalKinds: []string{"UpdFull_ExNK1", "UpdFull_NewNK3"}},
						},
						RelationshipFindingsInput: model.RelationshipFindingsInput{
							{Name: "UpdFull_NewFinding2", DisplayName: "New Finding 2", RelationshipKindName: "UpdFull_NewRK2", EnvironmentKindName: "UpdFull_ExEnvKind1", RemediationInput: model.RemediationInput{ShortDescription: "sd2", LongDescription: "ld2", ShortRemediation: "sr2", LongRemediation: "lr2"}},
						},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
			},
		},
		{
			name: "success_-_insert_new_opengraph_extension_with_one_already_present",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				// Create the first extension
				firstExtInput := model.ExtensionInput{Name: "TwoExt_First", DisplayName: "First Ext", Version: "1.0.0", Namespace: "TWO_FIRST"}
				firstNodeKinds := model.NodesInput{{Name: "TwoExt_NK1", DisplayName: "NK1", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}}
				firstEdgeKinds := model.RelationshipsInput{{Name: "TwoExt_RK1", Description: "rk", IsTraversable: true}}
				_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, model.GraphExtensionInput{
					ExtensionInput:         firstExtInput,
					NodeKindsInput:         firstNodeKinds,
					RelationshipKindsInput: firstEdgeKinds,
				})
				require.NoError(t, err)

				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "TwoExt_Second", DisplayName: "Second Ext", Version: "1.0.0", Namespace: "TWO_SECOND"},
						NodeKindsInput:         model.NodesInput{{Name: "TwoExt_NK2", DisplayName: "NK2", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "TwoExt_RK2", Description: "rk", IsTraversable: true}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
			},
		},
		{
			name: "success_-_environment_source_kind_auto_registers",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "AutoSrcExt", DisplayName: "Auto Src", Version: "1.0.0", Namespace: "AUTO_SRC"},
						NodeKindsInput:         model.NodesInput{{Name: "AutoSrc_NK1", DisplayName: "NK1", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}, {Name: "AutoSrc_EnvKind1", DisplayName: "Env Kind 1", IsDisplayKind: false}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "AutoSrc_RK1", Description: "rk", IsTraversable: true}},
						EnvironmentsInput:      model.EnvironmentsInput{{EnvironmentKindName: "AutoSrc_EnvKind1", SourceKindName: "UnregisteredSourceKind", PrincipalKinds: []string{"AutoSrc_NK1"}}},
						RelationshipFindingsInput: model.RelationshipFindingsInput{
							{Name: "AutoSrc_Finding1", DisplayName: "Finding 1", RelationshipKindName: "AutoSrc_RK1", EnvironmentKindName: "AutoSrc_EnvKind1", RemediationInput: model.RemediationInput{ShortDescription: "sd", LongDescription: "ld", ShortRemediation: "sr", LongRemediation: "lr"}},
						},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
			},
		},
		{
			name: "success_-_multiple_environments_with_different_source_kinds",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{Name: "MultiSrcExt", DisplayName: "Multi Src", Version: "1.0.0", Namespace: "MULTI_SRC"},
						NodeKindsInput: model.NodesInput{
							{Name: "MultiSrc_NK1", DisplayName: "NK1", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"},
							{Name: "MultiSrc_NK2", DisplayName: "NK2", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"},
							{Name: "MultiSrc_EnvKind1", DisplayName: "Env 1", IsDisplayKind: false},
							{Name: "MultiSrc_EnvKind2", DisplayName: "Env 2", IsDisplayKind: false},
						},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "MultiSrc_RK1", Description: "rk", IsTraversable: true}},
						EnvironmentsInput: model.EnvironmentsInput{
							{EnvironmentKindName: "MultiSrc_EnvKind1", SourceKindName: "MultiSrc_SrcKind1", PrincipalKinds: []string{"MultiSrc_NK1"}},
							{EnvironmentKindName: "MultiSrc_EnvKind2", SourceKindName: "UnregisteredSourceKind2", PrincipalKinds: []string{"MultiSrc_NK2"}},
						},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
			},
		},
		{
			name: "success_-_name_is_used_as_displayname_when_displayname_is_not_provided",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput:         model.ExtensionInput{Name: "NoDisplayNameExt", Version: "1.0.0", Namespace: "NO_DN"},
						NodeKindsInput:         model.NodesInput{{Name: "NoDN_NK1", DisplayName: "NK1", Description: "nk", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}},
						RelationshipKindsInput: model.RelationshipsInput{{Name: "NoDN_RK1", Description: "rk", IsTraversable: true}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
			},
		},
		{
			name: "success_-_preserves_existing_icon_name_and_color_when_not_provided_on_update",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				input := model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{Name: "PreserveIconExt", DisplayName: "Preserve Icon Test", Version: "v1.0.0", Namespace: "ICON_PRESERVE"},
					NodeKindsInput: model.NodesInput{{Name: "PrsIcn_NK1", DisplayName: "Node Kind", Description: "test", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}},
				}
				_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, input)
				require.NoError(t, err)
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{Name: "PreserveIconExt", DisplayName: "Preserve Icon Test", Version: "v1.0.0", Namespace: "ICON_PRESERVE"},
						NodeKindsInput: model.NodesInput{{Name: "PrsIcn_NK1", DisplayName: "Node Kind", Description: "test", IsDisplayKind: true, Icon: "", IconColor: ""}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)

				icons, err := testSuite.BHDatabase.GetCustomNodeKinds(testSuite.Context)
				require.NoError(t, err)
				var preserved *model.CustomNodeKind
				for _, icon := range icons {
					if icon.KindName == "PrsIcn_NK1" {
						iconCopy := icon
						preserved = &iconCopy
						break
					}
				}
				assert.NotNil(t, preserved, "custom node kind should still exist after update without icon fields")
				assert.Equal(t, "user", preserved.Config.Icon.Name, "icon name should be preserved from original upsert")
				assert.Equal(t, "#2779F5", preserved.Config.Icon.Color, "icon color should be preserved from original upsert")
			},
		},
		{
			name: "error_-_node_kind_has_duplicate_info_keys",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{Name: "DupInfoKeyExt", DisplayName: "Dup Info Key", Version: "1.0.0", Namespace: "DUP_IK"},
						NodeKindsInput: model.NodesInput{{
							Name: "DupIK_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000",
							Info: model.KindInfoInputs{
								{InfoKey: "duplicate", Title: "First", Position: 1, Content: []byte(`{"markdown": {"content": "first"}}`)},
								{InfoKey: "duplicate", Title: "Second", Position: 2, Content: []byte(`{"markdown": {"content": "second"}}`)},
							},
						}},
					},
					wantErrContains: "duplicate",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.Error(t, err)
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "error_-_node_kind_has_duplicate_info_positions",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{Name: "DupInfoPosExt", DisplayName: "Dup Info Pos", Version: "1.0.0", Namespace: "DUP_IP"},
						NodeKindsInput: model.NodesInput{{
							Name: "DupIP_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000",
							Info: model.KindInfoInputs{
								{InfoKey: "first", Title: "First", Position: 1, Content: []byte(`{"markdown": {"content": "first"}}`)},
								{InfoKey: "second", Title: "Second", Position: 1, Content: []byte(`{"markdown": {"content": "second"}}`)},
							},
						}},
					},
					wantErrContains: "duplicate",
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				assert.Error(t, err)
				assert.ErrorContains(t, err, setupData.wantErrContains)
				assertExtensionDoesNotExist(t, testSuite, setupData.input.ExtensionInput.Name)
			},
		},
		{
			name: "success_-_update_node_kind_removes_all_infos",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				input := model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{Name: "RemoveInfoExt", DisplayName: "Remove Info", Version: "1.0.0", Namespace: "REM_INFO"},
					NodeKindsInput: model.NodesInput{{
						Name: "RemInfo_NK1", DisplayName: "NK1", IsDisplayKind: true, Icon: "user", IconColor: "#000",
						Info: model.KindInfoInputs{
							{InfoKey: "overview", Title: "Overview", Position: 1, Content: []byte(`{"markdown": {"content": "overview"}}`)},
							{InfoKey: "details", Title: "Details", Position: 2, Content: []byte(`{"markdown": {"content": "details"}}`)},
						},
					}},
				}
				updated, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, input)
				require.NoError(t, err)
				require.False(t, updated)
				return testSetupData{input: input}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				setupData.input.NodeKindsInput[0].Info = model.KindInfoInputs{}
				updatedAgain, updateErr := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, setupData.input)
				require.NoError(t, updateErr)
				require.True(t, updatedAgain)
				extensions, _, getExtErr := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: setupData.input.ExtensionInput.Name}}}, model.Sort{}, 0, 1)
				require.NoError(t, getExtErr)
				nodeKinds, getNodeErr := testSuite.BHDatabase.GetGraphSchemaNodeKindsByExtensionId(testSuite.Context, extensions[0].ID)
				require.NoError(t, getNodeErr)
				infos, getInfoErr := testSuite.BHDatabase.GetKindInfos(testSuite.Context, nodeKinds[0].KindId)
				require.NoError(t, getInfoErr)
				assert.Empty(t, infos)
			},
		},
		{
			name: "success_-_display_node_kind_flipped_to_non_display_removes_custom_node_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				extensionInput := model.ExtensionInput{Name: "FlipDisplayExt", DisplayName: "Flip Display Node Kind to Non-Display NK", Version: "1.0.0", Namespace: "FLIP_DISP"}
				_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, model.GraphExtensionInput{
					ExtensionInput: extensionInput,
					NodeKindsInput: model.NodesInput{{Name: "StartedAsDispNodeKind", DisplayName: "Node Kind", Description: "test", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}},
				})
				require.NoError(t, err)
				// The custom node kind should exist before the flip.
				assertCustomNodeKindPresent(t, testSuite, "StartedAsDispNodeKind")

				// Re-upload the extension with the kind flipped to a non-display kind.
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: extensionInput,
						NodeKindsInput: model.NodesInput{{Name: "StartedAsDispNodeKind", DisplayName: "Node Kind", Description: "test", IsDisplayKind: false}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
				assertCustomNodeKindAbsent(t, testSuite, "StartedAsDispNodeKind")
			},
		},
		{
			name: "success_-_non_display_node_kind_flipped_to_display_creates_custom_node_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				extensionInput := model.ExtensionInput{Name: "FlipToDisplayExt", DisplayName: "Flip Non-Display Node Kind to Display NK", Version: "1.0.0", Namespace: "FLIP_TO_DISP"}
				_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, model.GraphExtensionInput{
					ExtensionInput: extensionInput,
					NodeKindsInput: model.NodesInput{{Name: "StartedAsNonDispNodeKind", DisplayName: "Node Kind", Description: "test", IsDisplayKind: false}},
				})
				require.NoError(t, err)
				// The custom node kind should not exist before the flip.
				assertCustomNodeKindAbsent(t, testSuite, "StartedAsNonDispNodeKind")

				// Re-upload the extension with the kind flipped to a display kind.
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: extensionInput,
						NodeKindsInput: model.NodesInput{{Name: "StartedAsNonDispNodeKind", DisplayName: "Node Kind", Description: "test", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
				assertCustomNodeKindPresent(t, testSuite, "StartedAsNonDispNodeKind")
			},
		},
		{
			name: "success_-_dropped_non_display_node_kind_is_stubbed_into_custom_node_kinds",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				extensionInput := model.ExtensionInput{Name: "DropNonDisplayExt", DisplayName: "Drop Non-Display Kind", Version: "1.0.0", Namespace: "DROP_ND"}
				_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, model.GraphExtensionInput{
					ExtensionInput: extensionInput,
					NodeKindsInput: model.NodesInput{
						{Name: "NonDisplayKindToKeep", DisplayName: "Non-Display Kind 1", Description: "test", IsDisplayKind: false},
						{Name: "NonDisplayKindToDrop", DisplayName: "Non-Display Kind 2", Description: "test", IsDisplayKind: false},
					},
				})
				require.NoError(t, err)
				// Non-display kinds are not tracked in custom_node_kinds while the schema is active.
				assertCustomNodeKindAbsent(t, testSuite, "NonDisplayKindToDrop")

				// Re-upload the extension with NonDisplayKindToDrop removed, triggering a reconciliation delete.
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: extensionInput,
						NodeKindsInput: model.NodesInput{{Name: "NonDisplayKindToKeep", DisplayName: "Non-Display Kind 1", Description: "test", IsDisplayKind: false}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
				// The dropped non-display kind should now be stubbed into custom_node_kinds.
				assertCustomNodeKindPresent(t, testSuite, "NonDisplayKindToDrop")
				kind, err := testSuite.BHDatabase.GetCustomNodeKind(testSuite.Context, "NonDisplayKindToDrop")
				require.NoError(t, err)
				assert.Nil(t, kind.SchemaNodeKindId, "stub should not reference the (now deleted) schema_node_kind row")
			},
		},
		{
			name: "success_-_deleted_schema_node_kind_nulls_schema_node_kind_id_on_custom_node_kind",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				extensionInput := model.ExtensionInput{Name: "DropKindExt", DisplayName: "Drop Kind", Version: "1.0.0", Namespace: "DROP_KIND"}
				_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, model.GraphExtensionInput{
					ExtensionInput: extensionInput,
					NodeKindsInput: model.NodesInput{
						{Name: "NodeKindToKeep", DisplayName: "Node Kind 1", Description: "test", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"},
						{Name: "NodeKindToRemove", DisplayName: "Node Kind 2", Description: "test", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"},
					},
				})
				require.NoError(t, err)
				// The custom node kind should exist before the kind is removed.
				assertCustomNodeKindPresent(t, testSuite, "NodeKindToRemove")

				// Re-upload the extension with NodeKindToRemove removed entirely.
				return testSetupData{
					input: model.GraphExtensionInput{
						ExtensionInput: extensionInput,
						NodeKindsInput: model.NodesInput{{Name: "NodeKindToKeep", DisplayName: "Node Kind 1", Description: "test", IsDisplayKind: true, Icon: "user", IconColor: "#2779F5"}},
					},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				assertGraphExtension(t, testSuite, setupData.input)
				// verify the schema_node_kind_id FK is nulled
				assertCustomNodeKindPresent(t, testSuite, "NodeKindToRemove")
				kind, err := testSuite.BHDatabase.GetCustomNodeKind(testSuite.Context, "NodeKindToRemove")
				require.NoError(t, err)
				assert.Nil(t, kind.SchemaNodeKindId)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testSuite := setupIntegrationTestSuite(t)
			defer teardownIntegrationTestSuite(t, &testSuite)

			setupData := testCase.setup(t, testSuite)
			updated, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, setupData.input)
			testCase.assert(t, testSuite, setupData, updated, err)
		})
	}
}

// assertGraphExtension retrieves and validates the full extension state against the expected input.
func assertGraphExtension(t *testing.T, testSuite IntegrationTestSuite, want model.GraphExtensionInput) {
	t.Helper()

	extensions, totalRecords, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context,
		model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: want.ExtensionInput.Name, SetOperator: model.FilterAnd}}},
		model.Sort{}, 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, totalRecords)
	gotExtension := extensions[0]

	assert.Equalf(t, want.ExtensionInput.Name, gotExtension.Name, "Extension - name mismatch")
	assert.Equalf(t, want.ExtensionInput.GetDisplayName(), gotExtension.DisplayName, "Extension - displayname mismatch")
	assert.Equalf(t, want.ExtensionInput.Version, gotExtension.Version, "Extension - version mismatch")
	assert.Equalf(t, want.ExtensionInput.Namespace, gotExtension.Namespace, "Extension - namespace mismatch")

	assertNodeKinds(t, testSuite, gotExtension.ID, want.NodeKindsInput)
	assertRelationshipKinds(t, testSuite, gotExtension.ID, want.RelationshipKindsInput)
	assertEnvironments(t, testSuite, gotExtension.ID, want.EnvironmentsInput)
	assertFindings(t, testSuite, gotExtension.ID, want.RelationshipFindingsInput)
}

func assertNodeKinds(t *testing.T, testSuite IntegrationTestSuite, extensionId int32, wantNodeKinds model.NodesInput) {
	t.Helper()

	schemaIdFilter := model.Filter{Operator: model.Equals, Value: strconv.FormatInt(int64(extensionId), 10), SetOperator: model.FilterAnd}
	gotNodeKinds, _, err := testSuite.BHDatabase.GetGraphSchemaNodeKinds(testSuite.Context,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	assert.Equalf(t, len(wantNodeKinds), len(gotNodeKinds), "node kind - count mismatch")

	wantByName := make(map[string]model.NodeInput, len(wantNodeKinds))
	for _, wantNK := range wantNodeKinds {
		wantByName[wantNK.Name] = wantNK
	}
	for _, gotNK := range gotNodeKinds {
		wantNK, ok := wantByName[gotNK.Name]
		assert.Truef(t, ok, "NodeKind(%v) - unexpected", gotNK.Name)
		assert.Greaterf(t, gotNK.ID, int32(0), "NodeKind(%v) - ID invalid", gotNK.Name)
		assert.NotZerof(t, gotNK.KindId, "NodeKind(%v) - KindId should be populated from database", gotNK.Name)
		assert.Equalf(t, extensionId, gotNK.SchemaExtensionId, "NodeKind(%v) - SchemaExtensionId mismatch", gotNK.Name)
		assert.Equalf(t, wantNK.DisplayName, gotNK.DisplayName, "NodeKind(%v) - display_name mismatch", gotNK.Name)
		assert.Equalf(t, wantNK.Description, gotNK.Description, "NodeKind(%v) - description mismatch", gotNK.Name)
		assert.Equalf(t, wantNK.IsDisplayKind, gotNK.IsDisplayKind, "NodeKind(%v) - is_display_kind mismatch", gotNK.Name)
		assert.Equalf(t, wantNK.Icon, gotNK.Icon, "NodeKind(%v) - icon mismatch", gotNK.Name)
		assert.Equalf(t, wantNK.IconColor, gotNK.IconColor, "NodeKind(%v) - icon_color mismatch", gotNK.Name)
		assert.Falsef(t, gotNK.CreatedAt.IsZero(), "NodeKind(%v) - created_at is zero", gotNK.Name)
		assert.Falsef(t, gotNK.UpdatedAt.IsZero(), "NodeKind(%v) - updated_at is zero", gotNK.Name)
		assert.Falsef(t, gotNK.DeletedAt.Valid, "NodeKind(%v) - deleted_at is not null", gotNK.Name)

		// Assert Info entries
		assertKindInfos(t, testSuite, gotNK.KindId, gotNK.Name, wantNK.Info)
	}

	// Custom icon assertions
	icons, err := testSuite.BHDatabase.GetCustomNodeKinds(testSuite.Context)
	require.NoError(t, err)
	iconMap := make(map[string]model.CustomNodeKind, len(icons))
	for _, icon := range icons {
		iconMap[icon.KindName] = icon
	}
	for _, gotNK := range gotNodeKinds {
		if gotNK.IsDisplayKind {
			icon, ok := iconMap[gotNK.Name]
			assert.Truef(t, ok, "NodeKind(%v) - missing custom icon", gotNK.Name)
			assert.Equalf(t, gotNK.Icon, icon.Config.Icon.Name, "NodeKind(%v) - icon name mismatch", gotNK.Name)
			assert.Equalf(t, gotNK.IconColor, icon.Config.Icon.Color, "NodeKind(%v) - icon color mismatch", gotNK.Name)
		} else {
			_, ok := iconMap[gotNK.Name]
			assert.Falsef(t, ok, "NodeKind(%v) - non-display kind should not have custom icon", gotNK.Name)
		}
	}
}

func assertRelationshipKinds(t *testing.T, testSuite IntegrationTestSuite, extensionId int32, wantRelKinds model.RelationshipsInput) {
	t.Helper()

	schemaIdFilter := model.Filter{Operator: model.Equals, Value: strconv.FormatInt(int64(extensionId), 10), SetOperator: model.FilterAnd}
	gotRelKinds, _, err := testSuite.BHDatabase.GetGraphSchemaRelationshipKinds(testSuite.Context,
		model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0)
	require.NoError(t, err)
	assert.Equalf(t, len(wantRelKinds), len(gotRelKinds), "relationship kind - count mismatch")

	wantByName := make(map[string]model.RelationshipInput, len(wantRelKinds))
	for _, wantRK := range wantRelKinds {
		wantByName[wantRK.Name] = wantRK
	}
	for _, gotRK := range gotRelKinds {
		wantRK, ok := wantByName[gotRK.Name]
		assert.Truef(t, ok, "RelKind(%v) - unexpected", gotRK.Name)
		assert.Greaterf(t, gotRK.ID, int32(0), "RelKind(%v) - ID invalid", gotRK.Name)
		assert.NotZerof(t, gotRK.KindId, "RelKind(%v) - KindId should be populated from database", gotRK.Name)
		assert.Equalf(t, extensionId, gotRK.SchemaExtensionId, "RelKind(%v) - SchemaExtensionId mismatch", gotRK.Name)
		assert.Equalf(t, wantRK.Description, gotRK.Description, "RelKind(%v) - description mismatch", gotRK.Name)
		assert.Equalf(t, wantRK.IsTraversable, gotRK.IsTraversable, "RelKind(%v) - is_traversable mismatch", gotRK.Name)

		// Assert Info entries
		assertKindInfos(t, testSuite, gotRK.KindId, gotRK.Name, wantRK.Info)
	}
}

func assertEnvironments(t *testing.T, testSuite IntegrationTestSuite, extensionId int32, wantEnvironments model.EnvironmentsInput) {
	t.Helper()

	gotEnvironments, err := testSuite.BHDatabase.GetEnvironmentsByExtensionId(testSuite.Context, extensionId)
	require.NoError(t, err)
	assert.Equalf(t, len(wantEnvironments), len(gotEnvironments), "environments - count mismatch")

	wantByKindName := make(map[string]model.EnvironmentInput, len(wantEnvironments))
	for _, wantEnv := range wantEnvironments {
		wantByKindName[wantEnv.EnvironmentKindName] = wantEnv
	}
	for _, gotEnv := range gotEnvironments {
		wantEnv, ok := wantByKindName[gotEnv.EnvironmentKindName]
		assert.Truef(t, ok, "Environment(%v) - unexpected", gotEnv.EnvironmentKindName)
		assert.Greaterf(t, gotEnv.ID, int32(0), "Environment(%v) - ID invalid", gotEnv.EnvironmentKindName)
		assert.Equalf(t, extensionId, gotEnv.SchemaExtensionId, "Environment(%v) - SchemaExtensionId mismatch", gotEnv.EnvironmentKindName)

		sourceKinds, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, gotEnv.SourceKindId)
		require.NoError(t, err)
		require.Len(t, sourceKinds, 1)
		assert.Equalf(t, wantEnv.SourceKindName, sourceKinds[0].Name, "Environment(%v) - SourceKindName mismatch", gotEnv.EnvironmentKindName)

		gotPrincipalKinds, err := testSuite.BHDatabase.GetPrincipalKindsByEnvironmentId(testSuite.Context, gotEnv.ID)
		require.NoError(t, err)
		assert.Equalf(t, len(wantEnv.PrincipalKinds), len(gotPrincipalKinds), "Environment(%v) - PrincipalKinds count mismatch", gotEnv.EnvironmentKindName)
		for _, gotPK := range gotPrincipalKinds {
			assert.Equalf(t, gotEnv.ID, gotPK.EnvironmentId, "Environment(%v) - PrincipalKind EnvironmentId mismatch", gotEnv.EnvironmentKindName)
			principalKinds, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, gotPK.PrincipalKind)
			require.NoError(t, err)
			require.Len(t, principalKinds, 1)
			assert.Containsf(t, wantEnv.PrincipalKinds, principalKinds[0].Name, "Environment(%v) - PrincipalKind name mismatch", gotEnv.EnvironmentKindName)
		}
	}
}

func assertKindInfos(t *testing.T, testSuite IntegrationTestSuite, kindId int32, kindName string, wantInfos model.KindInfoInputs) {
	t.Helper()

	gotInfos, err := testSuite.BHDatabase.GetKindInfos(testSuite.Context, kindId)
	require.NoError(t, err)
	assert.Equalf(t, len(wantInfos), len(gotInfos), "Kind(%v) - info count mismatch", kindName)

	wantByKey := make(map[string]model.KindInfoInput, len(wantInfos))
	for _, wantInfo := range wantInfos {
		wantByKey[wantInfo.InfoKey] = wantInfo
	}

	for _, gotInfo := range gotInfos {
		wantInfo, ok := wantByKey[gotInfo.InfoKey]
		assert.Truef(t, ok, "Kind(%v) Info(%v) - unexpected", kindName, gotInfo.InfoKey)
		assert.Greaterf(t, gotInfo.ID, int32(0), "Kind(%v) Info(%v) - ID invalid", kindName, gotInfo.InfoKey)
		assert.Equalf(t, kindId, gotInfo.KindID, "Kind(%v) Info(%v) - KindID mismatch", kindName, gotInfo.InfoKey)
		assert.Equalf(t, wantInfo.Title, gotInfo.Title, "Kind(%v) Info(%v) - title mismatch", kindName, gotInfo.InfoKey)
		assert.Equalf(t, wantInfo.Position, gotInfo.Position, "Kind(%v) Info(%v) - position mismatch", kindName, gotInfo.InfoKey)
		assert.JSONEqf(t, string(wantInfo.Content), string(gotInfo.Content), "Kind(%v) Info(%v) - content mismatch", kindName, gotInfo.InfoKey)
	}
}

func assertFindings(t *testing.T, testSuite IntegrationTestSuite, extensionId int32, wantFindings model.RelationshipFindingsInput) {
	t.Helper()

	gotFindings, err := testSuite.BHDatabase.GetSchemaFindingsByExtensionId(testSuite.Context, extensionId)
	require.NoError(t, err)
	assert.Equalf(t, len(wantFindings), len(gotFindings), "findings - count mismatch")

	wantByName := make(map[string]model.RelationshipFindingInput, len(wantFindings))
	for _, wantFinding := range wantFindings {
		wantByName[wantFinding.Name] = wantFinding
	}
	for _, gotFinding := range gotFindings {
		wantFinding, ok := wantByName[gotFinding.Name]
		assert.Truef(t, ok, "Finding(%v) - unexpected", gotFinding.Name)
		assert.Greaterf(t, gotFinding.ID, int32(0), "Finding(%v) - ID invalid", gotFinding.Name)
		assert.Equalf(t, extensionId, gotFinding.SchemaExtensionId, "Finding(%v) - SchemaExtensionId mismatch", gotFinding.Name)

		relKinds, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, gotFinding.KindId)
		require.NoError(t, err)
		require.Len(t, relKinds, 1)
		assert.Equalf(t, wantFinding.RelationshipKindName, relKinds[0].Name, "Finding(%v) - relationship kind mismatch", gotFinding.Name)

		findingEnv, err := testSuite.BHDatabase.GetEnvironmentById(testSuite.Context, gotFinding.EnvironmentId)
		require.NoError(t, err)
		envKinds, err := testSuite.BHDatabase.GetKindsByIDs(testSuite.Context, findingEnv.EnvironmentKindId)
		require.NoError(t, err)
		require.Len(t, envKinds, 1)
		assert.Equalf(t, wantFinding.EnvironmentKindName, envKinds[0].Name, "Finding(%v) - environment kind mismatch", gotFinding.Name)

		assert.Equalf(t, wantFinding.DisplayName, gotFinding.DisplayName, "Finding(%v) - display name mismatch", gotFinding.Name)

		gotRemediation, err := testSuite.BHDatabase.GetRemediationByFindingId(testSuite.Context, gotFinding.ID)
		require.NoError(t, err)
		assert.Equalf(t, wantFinding.RemediationInput.ShortRemediation, gotRemediation.ShortRemediation, "Finding(%v) - short_remediation mismatch", gotFinding.Name)
		assert.Equalf(t, wantFinding.RemediationInput.LongRemediation, gotRemediation.LongRemediation, "Finding(%v) - long_remediation mismatch", gotFinding.Name)
		assert.Equalf(t, wantFinding.RemediationInput.ShortDescription, gotRemediation.ShortDescription, "Finding(%v) - short_description mismatch", gotFinding.Name)
		assert.Equalf(t, wantFinding.RemediationInput.LongDescription, gotRemediation.LongDescription, "Finding(%v) - long_description mismatch", gotFinding.Name)
	}
}
