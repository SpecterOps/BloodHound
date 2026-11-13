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

package analysis

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// newSameHostExpansionFixture links computer <-> VM <-> device without a direct
// computer/device edge, so reaching all identities requires revisiting the queue.
// The group and containers exercise ordinary expansion before and after host
// discovery; the sibling VM must remain excluded when only its parent is added.
func newSameHostExpansionFixture(t *testing.T, suite IntegrationTestSuite) map[string]*graph.Node {
	t.Helper()
	var nodes = make(map[string]*graph.Node)
	require.NoError(t, suite.GraphDB.WriteTransaction(suite.Context, func(transaction graph.Transaction) error {
		for name, kinds := range map[string]graph.Kinds{
			"computer": {ad.Entity, ad.Computer}, "vm": {azure.Entity, azure.VM},
			"device": {azure.Entity, azure.Device}, "group": {ad.Entity, ad.Group},
			"ou": {ad.Entity, ad.OU}, "resource-group": {azure.Entity, azure.ResourceGroup},
			"subscription": {azure.Entity, azure.Subscription}, "sibling": {azure.Entity, azure.VM},
			"non-host": {ad.Entity, ad.User},
		} {
			var node, err = transaction.CreateNode(graph.AsProperties(graph.PropertyMap{
				common.ObjectID: "same-host-" + name, common.Name: name,
				ad.DomainSID: "S-1-5-21-100-200-300", azure.TenantID: "same-host-tenant",
			}), kinds...)
			if err != nil {
				return err
			}
			nodes[name] = node
		}
		for _, relationship := range []struct {
			from, to string
			kind     graph.Kind
		}{
			{"computer", "vm", common.SameHostAs}, {"vm", "computer", common.SameHostAs},
			{"vm", "device", common.SameHostAs}, {"device", "vm", common.SameHostAs},
			{"computer", "group", ad.MemberOf}, {"ou", "computer", ad.Contains},
			{"resource-group", "vm", azure.Contains}, {"subscription", "resource-group", azure.Contains},
			{"resource-group", "sibling", azure.Contains},
			// Even a malformed SameHostAs edge must not expand into a non-host kind.
			{"vm", "non-host", common.SameHostAs}, {"non-host", "vm", common.SameHostAs},
		} {
			if _, err := transaction.CreateRelationshipByIDs(nodes[relationship.from].ID, nodes[relationship.to].ID, relationship.kind, graph.NewProperties()); err != nil {
				return err
			}
		}
		return nil
	}))
	return nodes
}

// TestAGTSameHostExpansion verifies membership and seed/child/parent provenance
// from each host identity, including hosts discovered through group membership.
// It also checks that None and Parents retain their existing expansion scope.
func TestAGTSameHostExpansion(t *testing.T) {
	var (
		suite      = setupIntegrationTestSuite(t)
		nodes      = newSameHostExpansionFixture(t, suite)
		parameters = appcfg.GetAGTParameters(suite.Context, suite.BHDatabase)
	)
	t.Cleanup(func() { teardownIntegrationTestSuite(t, &suite) })
	for _, scenario := range []struct {
		name     string
		seeds    []string
		method   model.AssetGroupExpansionMethod
		children []string
		parents  []string
	}{
		{"computer reaches VM and device through a cycle", []string{"computer"}, model.AssetGroupExpansionMethodChildren, []string{"vm", "device"}, nil},
		{"VM reaches both platforms", []string{"vm"}, model.AssetGroupExpansionMethodChildren, []string{"computer", "device"}, nil},
		{"device reaches VM and computer", []string{"device"}, model.AssetGroupExpansionMethodChildren, []string{"vm", "computer"}, nil},
		{"group membership discovers linked hosts", []string{"group"}, model.AssetGroupExpansionMethodChildren, []string{"computer", "vm", "device"}, nil},
		{"full expansion includes cross-platform parents", []string{"computer"}, model.AssetGroupExpansionMethodAll, []string{"vm", "device"}, []string{"ou", "resource-group", "subscription"}},
		{"no expansion keeps only the seed", []string{"computer"}, model.AssetGroupExpansionMethodNone, nil, nil},
		{"parents-only does not follow host links", []string{"computer"}, model.AssetGroupExpansionMethodParents, nil, []string{"ou"}},
		{"overlapping seeds retain seed provenance", []string{"computer", "device"}, model.AssetGroupExpansionMethodChildren, []string{"vm"}, nil},
		{"non-host roots do not follow host links", []string{"non-host"}, model.AssetGroupExpansionMethodChildren, nil, nil},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var (
				seeds    []model.SelectorSeed
				expected = make(map[graph.ID]model.AssetGroupSelectorNodeSource)
			)
			for _, name := range scenario.seeds {
				seeds = append(seeds, createSelectorSeed(t, nodes[name]))
				expected[nodes[name].ID] = model.AssetGroupSelectorNodeSourceSeed
			}
			for _, name := range scenario.children {
				expected[nodes[name].ID] = model.AssetGroupSelectorNodeSourceChild
			}
			for _, name := range scenario.parents {
				expected[nodes[name].ID] = model.AssetGroupSelectorNodeSourceParent
			}
			var actual, errs = FetchNodesFromSeeds(suite.Context, parameters, suite.GraphDB, seeds, scenario.method, -1)
			require.Empty(t, errs)
			require.ElementsMatch(t, slices.Collect(maps.Keys(expected)), slices.Collect(maps.Keys(actual)))
			for nodeID, source := range expected {
				require.Equal(t, source, actual[nodeID].Source)
			}
		})
	}
	t.Run("preview limits terminate cycles", func(t *testing.T) {
		// Limits interrupt different stages of the cyclic traversal, but must
		// never discard the original seed or return more members than requested.
		for _, limit := range []int{1, 2, 3, 4} {
			var actual, errs = FetchNodesFromSeeds(suite.Context, parameters, suite.GraphDB, []model.SelectorSeed{createSelectorSeed(t, nodes["computer"])}, model.AssetGroupExpansionMethodAll, limit)
			require.Empty(t, errs)
			require.Len(t, actual, limit)
			require.Equal(t, model.AssetGroupSelectorNodeSourceSeed, actual[nodes["computer"].ID].Source)
		}
	})
}

// createSameHostSelector creates persisted selectors with the requested policy.
// Tiers require certification so tests can distinguish selection from tagging;
// labels and owned tags keep their usual non-certification-gated membership.
func createSameHostSelector(t *testing.T, suite IntegrationTestSuite, seed *graph.Node, name string, tagType model.AssetGroupTagType, position int32, autoCertify model.SelectorAutoCertifyMethod) (model.AssetGroupTag, model.AssetGroupTagSelector) {
	t.Helper()
	var (
		actor          = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("12345678-9012-4567-9012-456789012345")}}
		tierPosition   null.Int32
		requireCertify null.Bool
	)
	if tagType == model.AssetGroupTagTypeTier {
		tierPosition = null.Int32From(position)
		requireCertify = null.BoolFrom(true)
	}
	var tag, tagErr = suite.BHDatabase.CreateAssetGroupTag(suite.Context, tagType, actor, name, "", tierPosition, requireCertify, null.String{})
	require.NoError(t, tagErr)
	var selector, selectorErr = suite.BHDatabase.CreateAssetGroupTagSelector(suite.Context, actor, model.AssetGroupTagSelector{
		AssetGroupTagId: tag.ID, Name: name, AllowDisable: true, AutoCertify: autoCertify,
		Seeds: []model.SelectorSeed{createSelectorSeed(t, seed)},
	})
	require.NoError(t, selectorErr)
	require.NoError(t, assertGraphKinds(suite.Context, suite.GraphDB, graph.Kinds{tag.ToKind()}))
	return tag, selector
}

// applySameHostTag persists the tagging diff and reads actual graph membership,
// ensuring assertions cover tag additions/removals rather than only selection.
func applySameHostTag(t *testing.T, suite IntegrationTestSuite, tag model.AssetGroupTag) []graph.ID {
	t.Helper()
	var updates = make(map[uint64]*graph.Node)
	require.NoError(t, tagAssetGroupNodesForTag(suite.Context, suite.BHDatabase, suite.GraphDB, tag, cardinality.NewBitmap64(), updates))
	require.NoError(t, ops.UpdateNodes(suite.Context, suite.GraphDB, slices.Collect(maps.Values(updates)), AGTBatchNodeUpdateSize))
	var tagged []graph.ID
	require.NoError(t, suite.GraphDB.ReadTransaction(suite.Context, func(transaction graph.Transaction) error {
		var nodes, err = ops.FetchNodes(transaction.Nodes().Filter(query.Kind(query.Node(), tag.ToKind())))
		for _, node := range nodes {
			tagged = append(tagged, node.ID)
		}
		return err
	}))
	return tagged
}

// TestAGTSameHostCertificationAndRemoval checks that host expansion feeds the
// existing certification and reconciliation rules. Discovering another identity
// must not bypass tier certification or make its membership permanent.
func TestAGTSameHostCertificationAndRemoval(t *testing.T) {
	var (
		suite      = setupIntegrationTestSuite(t)
		nodes      = newSameHostExpansionFixture(t, suite)
		parameters = appcfg.GetAGTParameters(suite.Context, suite.BHDatabase)
	)
	t.Cleanup(func() { teardownIntegrationTestSuite(t, &suite) })
	var primaryKinds, err = suite.BHDatabase.GetPrimaryDisplayKinds(suite.Context)
	require.NoError(t, err)
	for _, scenario := range []struct {
		name          string
		tagType       model.AssetGroupTagType
		certification model.SelectorAutoCertifyMethod
		selected      int
		tagged        []string
	}{
		{"label all-members", model.AssetGroupTagTypeLabel, model.SelectorAutoCertifyMethodAllMembers, 3, []string{"computer", "vm", "device"}},
		{"zone all-members", model.AssetGroupTagTypeTier, model.SelectorAutoCertifyMethodAllMembers, 6, []string{"computer", "vm", "device", "ou", "resource-group", "subscription"}},
		{"zone seeds-only", model.AssetGroupTagTypeTier, model.SelectorAutoCertifyMethodSeedsOnly, 6, []string{"computer"}},
		{"zone certification disabled", model.AssetGroupTagTypeTier, model.SelectorAutoCertifyMethodDisabled, 6, nil},
		{"label certification disabled", model.AssetGroupTagTypeLabel, model.SelectorAutoCertifyMethodDisabled, 3, []string{"computer", "vm", "device"}},
		{"owned does not expand", model.AssetGroupTagTypeOwned, model.SelectorAutoCertifyMethodAllMembers, 1, []string{"computer"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var (
				tag, selector = createSameHostSelector(t, suite, nodes["computer"], scenario.name, scenario.tagType, 2, scenario.certification)
				expected      []graph.ID
			)
			runSelectNodes(t, suite.Context, suite.BHDatabase, suite.GraphDB, parameters, primaryKinds, selector, tag.GetExpansionMethod())
			var selected = requireSelectorNodesById(t, suite.Context, suite.BHDatabase, selector.ID)
			require.Len(t, selected, scenario.selected)
			if scenario.selected > 1 {
				require.Equal(t, model.AssetGroupSelectorNodeSourceChild, selected[nodes["device"].ID].Source)
				require.Equal(t, model.AssetGroupSelectorNodeSourceChild, selected[nodes["vm"].ID].Source)
				var expectedCertification = model.AssetGroupCertificationPending
				if scenario.certification == model.SelectorAutoCertifyMethodAllMembers {
					expectedCertification = model.AssetGroupCertificationAuto
				}
				require.Equal(t, expectedCertification, selected[nodes["vm"].ID].Certified)
				require.Equal(t, expectedCertification, selected[nodes["device"].ID].Certified)
			}
			for _, name := range scenario.tagged {
				expected = append(expected, nodes[name].ID)
			}
			require.ElementsMatch(t, expected, applySameHostTag(t, suite, tag))
		})
	}
	t.Run("manual and revoked certifications survive reselection", func(t *testing.T) {
		// Reselection with all-members auto-certification must preserve an
		// operator's manual approval and revocation on the discovered identities.
		var tag, selector = createSameHostSelector(t, suite, nodes["computer"], "manual host certifications", model.AssetGroupTagTypeTier, 2, model.SelectorAutoCertifyMethodAllMembers)
		runSelectNodes(t, suite.Context, suite.BHDatabase, suite.GraphDB, parameters, primaryKinds, selector, tag.GetExpansionMethod())
		var (
			selected = requireSelectorNodesById(t, suite.Context, suite.BHDatabase, selector.ID)
			vm       = selected[nodes["vm"].ID]
			device   = selected[nodes["device"].ID]
		)
		vm.Certified = model.AssetGroupCertificationManual
		device.Certified = model.AssetGroupCertificationRevoked
		require.NoError(t, suite.BHDatabase.UpdateSelectorNodes(suite.Context, []model.AssetGroupSelectorNode{vm, device}))
		runSelectNodes(t, suite.Context, suite.BHDatabase, suite.GraphDB, parameters, primaryKinds, selector, tag.GetExpansionMethod())
		selected = requireSelectorNodesById(t, suite.Context, suite.BHDatabase, selector.ID)
		require.Equal(t, model.AssetGroupCertificationManual, selected[vm.NodeId].Certified)
		require.Equal(t, model.AssetGroupCertificationRevoked, selected[device.NodeId].Certified)
		require.ElementsMatch(t, []graph.ID{nodes["computer"].ID, nodes["vm"].ID, nodes["ou"].ID, nodes["resource-group"].ID, nodes["subscription"].ID}, applySameHostTag(t, suite, tag))
	})
	t.Run("removed correlations remove peers and their parents", func(t *testing.T) {
		// Once host links disappear, the Azure identities and their containers
		// lose membership; the AD seed and its independently reached OU remain.
		var tag, selector = createSameHostSelector(t, suite, nodes["computer"], "stale host correlations", model.AssetGroupTagTypeTier, 2, model.SelectorAutoCertifyMethodAllMembers)
		runSelectNodes(t, suite.Context, suite.BHDatabase, suite.GraphDB, parameters, primaryKinds, selector, tag.GetExpansionMethod())
		require.Len(t, applySameHostTag(t, suite, tag), 6)
		require.NoError(t, suite.GraphDB.WriteTransaction(suite.Context, func(transaction graph.Transaction) error {
			return transaction.Relationships().Filter(query.Kind(query.Relationship(), common.SameHostAs)).Delete()
		}))
		runSelectNodes(t, suite.Context, suite.BHDatabase, suite.GraphDB, parameters, primaryKinds, selector, tag.GetExpansionMethod())
		require.Len(t, requireSelectorNodesById(t, suite.Context, suite.BHDatabase, selector.ID), 2)
		require.ElementsMatch(t, []graph.ID{nodes["computer"].ID, nodes["ou"].ID}, applySameHostTag(t, suite, tag))
	})
}

// TestAGTSameHostTierPrecedence selects the same host component from different
// identities in two tiers. Processing the higher-priority tier first must keep
// every overlapping host and parent out of the lower-priority tier.
func TestAGTSameHostTierPrecedence(t *testing.T) {
	var (
		suite      = setupIntegrationTestSuite(t)
		nodes      = newSameHostExpansionFixture(t, suite)
		parameters = appcfg.GetAGTParameters(suite.Context, suite.BHDatabase)
	)
	t.Cleanup(func() { teardownIntegrationTestSuite(t, &suite) })
	var primaryKinds, err = suite.BHDatabase.GetPrimaryDisplayKinds(suite.Context)
	require.NoError(t, err)
	var tiers []model.AssetGroupTag
	for index, seed := range []string{"computer", "device"} {
		var tag, selector = createSameHostSelector(t, suite, nodes[seed], fmt.Sprintf("host tier %d", index), model.AssetGroupTagTypeTier, int32(index+2), model.SelectorAutoCertifyMethodAllMembers)
		tiers = append(tiers, tag)
		runSelectNodes(t, suite.Context, suite.BHDatabase, suite.GraphDB, parameters, primaryKinds, selector, tag.GetExpansionMethod())
	}
	var (
		exclusions = cardinality.NewBitmap64()
		updates    = make(map[uint64]*graph.Node)
	)
	// Share exclusions across tiers, matching the production tagging loop.
	for _, tier := range tiers {
		require.NoError(t, tagAssetGroupNodesForTag(suite.Context, suite.BHDatabase, suite.GraphDB, tier, exclusions, updates))
	}
	require.NoError(t, ops.UpdateNodes(suite.Context, suite.GraphDB, slices.Collect(maps.Values(updates)), AGTBatchNodeUpdateSize))
	require.NoError(t, suite.GraphDB.ReadTransaction(suite.Context, func(transaction graph.Transaction) error {
		for _, name := range []string{"computer", "vm", "device", "ou", "resource-group", "subscription"} {
			var node, err = transaction.Nodes().Filter(query.Equals(query.NodeID(), nodes[name].ID)).First()
			if err != nil {
				return err
			}
			require.True(t, node.Kinds.ContainsOneOf(tiers[0].ToKind()), name)
			require.False(t, node.Kinds.ContainsOneOf(tiers[1].ToKind()), name)
		}
		return nil
	}))
}
