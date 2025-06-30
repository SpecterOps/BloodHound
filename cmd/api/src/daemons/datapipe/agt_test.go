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
// +build integration

package datapipe

import (
	"context"
	"testing"

	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestAGT_FetchNodesFromSeeds_Expansions(t *testing.T) {
	var (
		testContext = integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		harness     = integration.NewGPOAppliesToHarness(testContext)

		// OU3 of GPOAppliesTo harness offers the best view into each expansion approach
		seedObjectId, _ = harness.OU3.Properties.Get(common.ObjectID.String()).String()
		seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
	)

	t.Run("FetchNodesFromSeeds with no expansion", func(t *testing.T) {
		result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodNone, -1)
		require.Len(t, result, 1)
		require.Equal(t, result[harness.OU3.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
	})

	t.Run("FetchNodesFromSeeds with only child expansion", func(t *testing.T) {
		result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
		require.Len(t, result, 4)

		require.Equal(t, result[harness.OU3.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[harness.Group3.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		require.Equal(t, result[harness.User3.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		require.Equal(t, result[harness.Computer3.ID].Source, model.AssetGroupSelectorNodeSourceChild)
	})

	t.Run("FetchNodesFromSeeds with only parent expansion", func(t *testing.T) {
		result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents, -1)
		require.Len(t, result, 2)

		require.Equal(t, result[harness.OU3.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[harness.OU2.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})

	t.Run("FetchNodesFromSeeds with all expansions", func(t *testing.T) {
		result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodAll, -1)
		require.Len(t, result, 7)

		require.Equal(t, result[harness.OU3.ID].Source, model.AssetGroupSelectorNodeSourceSeed)

		require.Equal(t, result[harness.OU2.ID].Source, model.AssetGroupSelectorNodeSourceParent)
		require.Equal(t, result[harness.GPO2.ID].Source, model.AssetGroupSelectorNodeSourceParent)
		require.Equal(t, result[harness.GPO3.ID].Source, model.AssetGroupSelectorNodeSourceParent)

		require.Equal(t, result[harness.Group3.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		require.Equal(t, result[harness.User3.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		require.Equal(t, result[harness.Computer3.ID].Source, model.AssetGroupSelectorNodeSourceChild)
	})

	t.Run("FetchNodesFromSeeds with all expansions with limit for seeds only", func(t *testing.T) {
		result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodAll, 1)
		require.Len(t, result, 1)

		require.Equal(t, result[harness.OU3.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
	})
}

func TestAGT_FetchNodesFromSeeds_ChildExpansion(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves AD group members without limit", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.MembershipHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.MembershipHarness.GroupB.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)

			result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
			require.Len(t, result, 3)

			require.Equal(t, result[testContext.Harness.MembershipHarness.GroupB.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.MembershipHarness.UserA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.MembershipHarness.ComputerA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		})
	})

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves AZ group members without limit", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZGroupMembership.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.AZGroupMembership.Group.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)
			result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
			require.Len(t, result, 4)

			require.Equal(t, result[testContext.Harness.AZGroupMembership.Group.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.AZGroupMembership.UserA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.AZGroupMembership.UserB.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.AZGroupMembership.UserC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		})
	})

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves OU contained entities without limit", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.OUHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.OUHarness.OUA.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)
			result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
			require.Len(t, result, 4)

			require.Equal(t, result[testContext.Harness.OUHarness.OUA.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.OUHarness.OUC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.OUHarness.UserA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.OUHarness.UserB.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		})
	})

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves Contained entities without limit", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.GPOAppliesTo.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.GPOAppliesTo.Container1.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)
			result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
			require.Len(t, result, 4)

			require.Equal(t, result[testContext.Harness.GPOAppliesTo.Container1.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.GPOAppliesTo.Group1.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.GPOAppliesTo.User1.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.GPOAppliesTo.Computer1.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		})
	})

	t.Run("FetchNodesFromSeeds_ChildExpansion with limit", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.OUHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.OUHarness.OUA.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)
			result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, 2)
			require.Len(t, result, 2)

			require.Equal(t, result[testContext.Harness.OUHarness.OUA.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			for id, node := range result {
				if id != testContext.Harness.OUHarness.OUA.ID {
					require.Contains(t, []graph.ID{testContext.Harness.OUHarness.OUC.ID, testContext.Harness.OUHarness.UserA.ID, testContext.Harness.OUHarness.UserB.ID}, id)
					require.Equal(t, node.Source, model.AssetGroupSelectorNodeSourceChild)
				}
			}
		})
	})
}

func TestAGT_FetchNodesFromSeeds_ParentExpansion(t *testing.T) {
	var (
		testContext = integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		harness     = integration.NewGPOAppliesToHarness(testContext)

		user3ObjectId, _ = harness.User3.Properties.Get(common.ObjectID.String()).String()
		seeds            = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: user3ObjectId}}
	)
	t.Run("TestAGT_FetchNodesFromSeeds_ParentExpansion retrieves OUs from entities without limit", func(t *testing.T) {
		result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents, -1)
		require.Len(t, result, 5)

		require.Equal(t, result[harness.User3.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[harness.OU3.ID].Source, model.AssetGroupSelectorNodeSourceParent)
		require.Equal(t, result[harness.OU2.ID].Source, model.AssetGroupSelectorNodeSourceParent)
		require.Equal(t, result[harness.GPO2.ID].Source, model.AssetGroupSelectorNodeSourceParent)
		require.Equal(t, result[harness.GPO3.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})

	t.Run("TestAGT_FetchNodesFromSeeds_ParentExpansion with limit", func(t *testing.T) {
		result := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents, 2)
		require.Len(t, result, 2)

		require.Equal(t, result[harness.User3.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		for id, node := range result {
			if id != harness.User3.ID {
				require.Contains(t, []graph.ID{harness.OU1.ID, harness.OU2.ID}, id)
				require.Equal(t, node.Source, model.AssetGroupSelectorNodeSourceParent)
			}
		}
	})
}
