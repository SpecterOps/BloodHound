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

	"github.com/specterops/bloodhound/dawgs/graph"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestAGT_FetchNodesFromSeeds_Expansions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.SetupActiveDirectory()

	var (
		seedObjectId, _ = testContext.Harness.GPOEnforcement.OrganizationalUnitC.Properties.Get(common.ObjectID.String()).String()
		seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
	)

	t.Run("FetchNodesFromSeeds with no expansion", func(t *testing.T) {
		result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodNone)
		require.NoError(t, err)

		require.Len(t, result, 1)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
	})

	t.Run("FetchNodesFromSeeds with only child expansion", func(t *testing.T) {
		result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren)
		require.NoError(t, err)

		require.Len(t, result, 2)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.UserC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
	})

	t.Run("FetchNodesFromSeeds with only parent expansion", func(t *testing.T) {
		result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents)
		require.NoError(t, err)

		require.Len(t, result, 2)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitA.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})

	t.Run("FetchNodesFromSeeds with all expansions", func(t *testing.T) {
		result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodAll)
		require.NoError(t, err)

		require.Len(t, result, 3)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.UserC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitA.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})
}

func TestAGT_FetchNodesFromSeeds_ChildExpansion(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves AD group members", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.MembershipHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.MembershipHarness.GroupB.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)

			result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren)
			require.NoError(t, err)

			require.Len(t, result, 3)

			require.Equal(t, result[testContext.Harness.MembershipHarness.GroupB.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.MembershipHarness.UserA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.MembershipHarness.ComputerA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		})
	})

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves AZ group members", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZGroupMembership.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.AZGroupMembership.Group.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)
			result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren)
			require.NoError(t, err)

			require.Len(t, result, 4)

			require.Equal(t, result[testContext.Harness.AZGroupMembership.Group.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.AZGroupMembership.UserA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.AZGroupMembership.UserB.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.AZGroupMembership.UserC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		})
	})

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves OU contained entities", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.OUHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.OUHarness.OUA.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)
			result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren)
			require.NoError(t, err)

			require.Len(t, result, 4)

			require.Equal(t, result[testContext.Harness.OUHarness.OUA.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.OUHarness.OUC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.OUHarness.UserA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.OUHarness.UserB.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		})
	})
}

func TestAGT_FetchNodesFromSeeds_ParentExpansion(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.SetupActiveDirectory()

	var (
		ouCObjectId, _ = testContext.Harness.GPOEnforcement.OrganizationalUnitC.Properties.Get(common.ObjectID.String()).String()
		seeds          = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: ouCObjectId}}
	)

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves OUs from entities", func(t *testing.T) {
		result, err := FetchNodesFromSeeds(context.Background(), testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents)
		require.NoError(t, err)

		require.Len(t, result, 2)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitA.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})
}
