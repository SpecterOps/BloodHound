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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pzRule(ruleKey string, name string, description string, enabled bool, allowDisable bool, seedValue string) model.PZRuleInput {
	return model.PZRuleInput{
		ExtensionRuleId: ruleKey,
		Name:            name,
		Description:     description,
		Seeds:           []model.SelectorSeedInput{{Type: model.SelectorTypeCypher, Value: seedValue}},
		Enabled:         enabled,
		AllowDisable:    allowDisable,
	}
}

func upsertExtensionPZRules(t *testing.T, testSuite IntegrationTestSuite, extensionName string, pzRules ...model.PZRuleInput) int32 {
	t.Helper()

	input := model.GraphExtensionInput{
		ExtensionInput: model.ExtensionInput{
			Name:        extensionName,
			DisplayName: "Privilege Zone Rule Extension",
			Version:     "1.0.0",
			Namespace:   "PZR",
		},
		NodeKindsInput: model.NodesInput{{Name: "PZR_Node"}},
		PZRulesInput:   pzRules,
	}

	_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, input)
	require.NoError(t, err)

	extensions, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(
		testSuite.Context,
		model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: extensionName, SetOperator: model.FilterAnd}}},
		model.Sort{},
		0,
		1,
	)
	require.NoError(t, err)
	require.Len(t, extensions, 1)

	return extensions[0].ID
}

func assertExtensionPZRules(t *testing.T, testSuite IntegrationTestSuite, extensionID int32, expectedPZRules ...model.PZRuleInput) map[string]model.AssetGroupTagSelector {
	t.Helper()

	var (
		expectedPZRulesByKey   = make(map[string]model.PZRuleInput, len(expectedPZRules))
		selectorsByKey         = make(map[string]model.AssetGroupTagSelector, len(expectedPZRules))
		tierZeroAssetGroupTags model.AssetGroupTags
		err                    error
	)

	tierZeroAssetGroupTags, err = testSuite.BHDatabase.GetAssetGroupTags(testSuite.Context, model.SQLFilter{
		SQLString: "type = ? AND position = ?",
		Params:    []any{model.AssetGroupTagTypeTier, model.AssetGroupTierZeroPosition},
	})
	require.NoError(t, err)
	require.Len(t, tierZeroAssetGroupTags, 1)
	tierZeroAssetGroupTagID := tierZeroAssetGroupTags[0].ID

	for _, expectedPZRule := range expectedPZRules {
		expectedPZRulesByKey[expectedPZRule.ExtensionRuleId] = expectedPZRule
	}

	selectors, err := testSuite.BHDatabase.GetAssetGroupTagSelectorsByExtensionId(testSuite.Context, extensionID)
	require.NoError(t, err)
	require.Len(t, selectors, len(expectedPZRules))

	for _, selector := range selectors {
		require.True(t, selector.RuleKey.Valid)
		expectedPZRule, found := expectedPZRulesByKey[selector.RuleKey.String]
		require.Truef(t, found, "unexpected selector returned with rule key %q", selector.RuleKey.String)

		selector, err = testSuite.BHDatabase.GetAssetGroupTagSelectorBySelectorId(testSuite.Context, selector.ID)
		require.NoError(t, err)
		assert.Equal(t, tierZeroAssetGroupTagID, selector.AssetGroupTagId)
		assert.True(t, selector.ExtensionId.Valid)
		assert.Equal(t, extensionID, selector.ExtensionId.Int32)
		assert.Equal(t, model.AssetGroupActorOpenGraphExtensionManagement, selector.CreatedBy)
		assert.Equal(t, model.AssetGroupActorOpenGraphExtensionManagement, selector.UpdatedBy)
		assert.Equal(t, expectedPZRule.Name, selector.Name)
		assert.Equal(t, expectedPZRule.Description, selector.Description)
		assert.Equal(t, model.SelectorAutoCertifyMethodDisabled, selector.AutoCertify)
		assert.Equal(t, expectedPZRule.AllowDisable, selector.AllowDisable)
		assert.Equal(t, !expectedPZRule.Enabled, selector.DisabledAt.Valid)
		assert.Equal(t, !expectedPZRule.Enabled, selector.DisabledBy.Valid)
		if !expectedPZRule.Enabled {
			assert.Equal(t, model.AssetGroupActorOpenGraphExtensionManagement, selector.DisabledBy.String)
		}
		assert.Len(t, selector.Seeds, len(expectedPZRule.Seeds))
		for index, expectedSeed := range expectedPZRule.Seeds {
			assert.Equal(t, expectedSeed.Type, selector.Seeds[index].Type)
			assert.Equal(t, expectedSeed.Value, selector.Seeds[index].Value)
		}
		selectorsByKey[selector.RuleKey.String] = selector
	}

	return selectorsByKey
}

func TestBloodhoundDB_UpsertOpenGraphExtensionPZRules(t *testing.T) {
	type testSetupData struct {
		graphExtensionInput       model.GraphExtensionInput
		extensionID               int32
		expectedPZRules           model.PZRulesInput
		existingSelectorsByKey    map[string]model.AssetGroupTagSelector
		deletedAssetGroupSelector []int
	}

	var (
		baseGraphExtensionInput = model.GraphExtensionInput{
			ExtensionInput: model.ExtensionInput{
				DisplayName: "Privilege Zone Rule Extension",
				Version:     "1.0.0",
				Namespace:   "PZR",
			},
			NodeKindsInput: model.NodesInput{{Name: "PZR_Node"}},
		}
	)

	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	type testCase struct {
		name   string
		setup  func(t *testing.T, testSuite IntegrationTestSuite) testSetupData
		assert func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error)
	}

	tests := []testCase{
		{
			name: "success_create_privilege_zone_rules",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				var (
					expectedPZRules = model.PZRulesInput{
						pzRule("PZR_create_one", "Create One", "first created rule", true, true, "MATCH (n:CreateOne) RETURN n"),
						pzRule("PZR_create_two", "Create Two", "second created rule", false, false, "MATCH (n:CreateTwo) RETURN n"),
					}
					graphExtensionInput = baseGraphExtensionInput
				)

				graphExtensionInput.ExtensionInput.Name = "PZRuleReconcileCreate"
				graphExtensionInput.PZRulesInput = expectedPZRules

				return testSetupData{
					graphExtensionInput: graphExtensionInput,
					expectedPZRules:     expectedPZRules,
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)

				extensions, _, lookupErr := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context,
					model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: setupData.graphExtensionInput.ExtensionInput.Name, SetOperator: model.FilterAnd}}},
					model.Sort{},
					0,
					1,
				)
				require.NoError(t, lookupErr)
				require.Len(t, extensions, 1)
				setupData.extensionID = extensions[0].ID

				assertExtensionPZRules(t, testSuite, setupData.extensionID, setupData.expectedPZRules...)
			},
		},
		{
			name: "success_update_privilege_zone_rule_in_place",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				var (
					extensionName       = "PZRuleReconcileUpdate"
					initialPZRules      = model.PZRulesInput{pzRule("PZR_keep", "Keep", "initial rule", true, true, "MATCH (n:Keep) RETURN n")}
					expectedPZRules     = model.PZRulesInput{pzRule("PZR_keep", "Keep Updated", "updated rule", false, false, "MATCH (n:KeepUpdated) RETURN n")}
					graphExtensionInput = baseGraphExtensionInput
					extensionID         = upsertExtensionPZRules(t, testSuite, extensionName, initialPZRules...)
				)

				graphExtensionInput.ExtensionInput.Name = extensionName
				graphExtensionInput.PZRulesInput = expectedPZRules

				return testSetupData{
					graphExtensionInput:    graphExtensionInput,
					extensionID:            extensionID,
					expectedPZRules:        expectedPZRules,
					existingSelectorsByKey: assertExtensionPZRules(t, testSuite, extensionID, initialPZRules...),
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)

				selectorsByKey := assertExtensionPZRules(t, testSuite, setupData.extensionID, setupData.expectedPZRules...)
				assert.Equal(t, setupData.existingSelectorsByKey["PZR_keep"].ID, selectorsByKey["PZR_keep"].ID)
			},
		},
		{
			name: "success_delete_privilege_zone_rule_absent_from_input",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				var (
					extensionName  = "PZRuleReconcileDelete"
					initialPZRules = model.PZRulesInput{
						pzRule("PZR_keep", "Keep", "retained rule", true, true, "MATCH (n:Keep) RETURN n"),
						pzRule("PZR_drop", "Drop", "removed rule", true, true, "MATCH (n:Drop) RETURN n"),
					}
					expectedPZRules     = model.PZRulesInput{pzRule("PZR_keep", "Keep", "retained rule", true, true, "MATCH (n:Keep) RETURN n")}
					graphExtensionInput = baseGraphExtensionInput
					extensionID         = upsertExtensionPZRules(t, testSuite, extensionName, initialPZRules...)
					existingSelectors   = assertExtensionPZRules(t, testSuite, extensionID, initialPZRules...)
				)

				graphExtensionInput.ExtensionInput.Name = extensionName
				graphExtensionInput.PZRulesInput = expectedPZRules

				return testSetupData{
					graphExtensionInput:       graphExtensionInput,
					extensionID:               extensionID,
					expectedPZRules:           expectedPZRules,
					deletedAssetGroupSelector: []int{existingSelectors["PZR_drop"].ID},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				assertExtensionPZRules(t, testSuite, setupData.extensionID, setupData.expectedPZRules...)

				for _, selectorID := range setupData.deletedAssetGroupSelector {
					_, selectorErr := testSuite.BHDatabase.GetAssetGroupTagSelectorBySelectorId(testSuite.Context, selectorID)
					assert.ErrorIs(t, selectorErr, database.ErrNotFound)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setupData := testCase.setup(t, testSuite)
			t.Cleanup(func() {
				if setupData.extensionID != 0 {
					assert.NoError(t, testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, setupData.extensionID))
				}
			})

			updated, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, setupData.graphExtensionInput)
			testCase.assert(t, testSuite, &setupData, updated, err)
		})
	}
}
