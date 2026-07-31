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

package analysis

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAGT_FetchNodesFromSeeds_Expansions(t *testing.T) {
	var (
		testCtx       = context.Background()
		db            = integration.SetupDB(t)
		testContext   = integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		agtParameters = appcfg.GetAGTParameters(testCtx, db)
	)
	testContext.SetupActiveDirectory()

	var (
		seedObjectId, _ = testContext.Harness.GPOEnforcement.OrganizationalUnitC.Properties.Get(common.ObjectID.String()).String()
		seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
	)

	t.Run("FetchNodesFromSeeds with no expansion", func(t *testing.T) {
		result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodNone, -1)
		require.Empty(t, errs)
		require.Len(t, result, 1)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
	})

	t.Run("FetchNodesFromSeeds with only child expansion", func(t *testing.T) {
		result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
		require.Empty(t, errs)
		require.Len(t, result, 2)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.UserC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
	})

	t.Run("FetchNodesFromSeeds with only parent expansion", func(t *testing.T) {
		result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents, -1)
		require.Empty(t, errs)
		require.Len(t, result, 2)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitA.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})

	t.Run("FetchNodesFromSeeds with all expansions", func(t *testing.T) {
		result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodAll, -1)
		require.Empty(t, errs)
		require.Len(t, result, 3)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.UserC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitA.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})

	t.Run("FetchNodesFromSeeds with all expansions with limit for seeds only", func(t *testing.T) {
		result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodAll, 1)
		require.Empty(t, errs)
		require.Len(t, result, 1)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
	})
}

func TestAGT_FetchNodesFromSeeds_ChildExpansion(t *testing.T) {
	var (
		testCtx       = context.Background()
		db            = integration.SetupDB(t)
		testContext   = integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		agtParameters = appcfg.GetAGTParameters(testCtx, db)
	)

	t.Run("FetchNodesFromSeeds_ChildExpansion retrieves AD group members without limit", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.MembershipHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			var (
				seedObjectId, _ = testContext.Harness.MembershipHarness.GroupB.Properties.Get(common.ObjectID.String()).String()
				seeds           = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: seedObjectId}}
			)

			result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
			require.Empty(t, errs)
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
			result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
			require.Empty(t, errs)
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
			result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, -1)
			require.Empty(t, errs)
			require.Len(t, result, 4)

			require.Equal(t, result[testContext.Harness.OUHarness.OUA.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
			require.Equal(t, result[testContext.Harness.OUHarness.OUC.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.OUHarness.UserA.ID].Source, model.AssetGroupSelectorNodeSourceChild)
			require.Equal(t, result[testContext.Harness.OUHarness.UserB.ID].Source, model.AssetGroupSelectorNodeSourceChild)
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
			result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodChildren, 2)
			require.Empty(t, errs)
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
		testCtx       = context.Background()
		db            = integration.SetupDB(t)
		testContext   = integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		agtParameters = appcfg.GetAGTParameters(testCtx, db)
	)
	testContext.SetupActiveDirectory()

	var (
		userCObjectId, _ = testContext.Harness.GPOEnforcement.UserC.Properties.Get(common.ObjectID.String()).String()
		seeds            = []model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: userCObjectId}}
	)

	t.Run("TestAGT_FetchNodesFromSeeds_ParentExpansion retrieves OUs from entities without limit", func(t *testing.T) {
		result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents, -1)
		require.Empty(t, errs)
		require.Len(t, result, 3)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.UserC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID].Source, model.AssetGroupSelectorNodeSourceParent)
		require.Equal(t, result[testContext.Harness.GPOEnforcement.OrganizationalUnitA.ID].Source, model.AssetGroupSelectorNodeSourceParent)
	})

	t.Run("TestAGT_FetchNodesFromSeeds_ParentExpansion with limit", func(t *testing.T) {
		result, errs := FetchNodesFromSeeds(testCtx, agtParameters, testContext.Graph.Database, seeds, model.AssetGroupExpansionMethodParents, 2)
		require.Empty(t, errs)
		require.Len(t, result, 2)

		require.Equal(t, result[testContext.Harness.GPOEnforcement.UserC.ID].Source, model.AssetGroupSelectorNodeSourceSeed)
		for id, node := range result {
			if id != testContext.Harness.GPOEnforcement.UserC.ID {
				require.Contains(t, []graph.ID{testContext.Harness.GPOEnforcement.OrganizationalUnitA.ID, testContext.Harness.GPOEnforcement.OrganizationalUnitC.ID}, id)
				require.Equal(t, node.Source, model.AssetGroupSelectorNodeSourceParent)
			}
		}
	})
}

func TestContainsOnlyCypherSelectorErrors(t *testing.T) {
	var (
		cypherErrN        = &CypherSelectorError{CypherQuery: "MATCH (n) RETURN n", Err: errors.New("")}
		cypherErrM        = &CypherSelectorError{CypherQuery: "MATCH (m) RETURN m", Err: errors.New("")}
		wrappedCypherErr  = fmt.Errorf("selector failed: %w", cypherErrN)
		wrappedOtherErr   = fmt.Errorf("object failed: %w", errors.New(""))
		nonCypherErr      = errors.New("some other error")
		objectSelectorErr = errors.New("object selector failure")
	)

	t.Parallel()

	testCases := []struct {
		name     string
		errs     []error
		expected bool
	}{
		{
			name:     "returns false for empty slice",
			errs:     []error{},
			expected: false,
		},
		{
			name:     "returns true for multiple CypherSelectorErrors",
			errs:     []error{cypherErrN, cypherErrM},
			expected: true,
		},
		{
			name:     "returns false for single non-CypherSelectorError",
			errs:     []error{nonCypherErr},
			expected: false,
		},
		{
			name:     "returns false for mix of CypherSelectorError and other errors",
			errs:     []error{cypherErrN, objectSelectorErr},
			expected: false,
		},
		{
			name:     "returns true for wrapped CypherSelectorError",
			errs:     []error{wrappedCypherErr},
			expected: true,
		},
		{
			name:     "returns false when wrapped non-CypherSelectorError is mixed in",
			errs:     []error{wrappedCypherErr, wrappedOtherErr},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, testCase.expected, ContainsOnlyCypherSelectorErrors(testCase.errs))
		})
	}
}

// assertGraphKinds registers any dynamic kinds (e.g. asset group tag kinds)
// with the underlying pg driver's schema manager so that subsequent graph
// transactions referencing those kinds can be translated.
func assertGraphKinds(ctx context.Context, graphDB graph.Database, kinds graph.Kinds) error {
	driver, ok := graphDB.(*pg.Driver)
	if !ok {
		return fmt.Errorf("graph database is not a pg driver")
	}
	_, err := driver.KindMapper().AssertKinds(ctx, kinds)
	return err
}

// TestTagAssetGroupNodesForTag exercises tagAssetGroupNodesForTag end-to-end
// and guards against regressions
func TestTagAssetGroupNodesForTag(t *testing.T) {
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx   = suite.Context
		bhDB      = suite.BHDatabase
		graphDB   = suite.GraphDB
		testActor = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")}}
	)

	t.Run("removes the tag kind from every previously tagged node when no selector references them", func(t *testing.T) {
		pzNodeTagCounterVec.Reset()

		tag, err := bhDB.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, testActor, "regression label all", "", null.Int32{}, null.Bool{}, null.String{})
		require.NoError(t, err)

		_, err = bhDB.CreateAssetGroupTagSelector(testCtx, tag.ID, testActor, "regression selector all", "", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "REGRESSION-ALL-NO-MATCH"},
		})
		require.NoError(t, err)

		var (
			tagKind          = tag.ToKind()
			previouslyTagged []graph.ID
		)

		require.NoError(t, assertGraphKinds(testCtx, graphDB, graph.Kinds{tagKind}))

		const nodeCount = 5
		require.NoError(t, graphDB.WriteTransaction(testCtx, func(tx graph.Transaction) error {
			for i := 0; i < nodeCount; i++ {
				node, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     fmt.Sprintf("regression-all-node-%d", i),
					common.ObjectID: fmt.Sprintf("REGRESSION-ALL-OBJECT-ID-%d", i),
				}), ad.Entity, ad.User, tagKind)
				if err != nil {
					return err
				}
				previouslyTagged = append(previouslyTagged, node.ID)
			}
			return nil
		}))

		var (
			exclusionSet  = cardinality.NewBitmap64()
			nodesToUpdate = make(map[uint64]*graph.Node)
		)

		assert.NoError(t, tagAssetGroupNodesForTag(testCtx, bhDB, graphDB, tag, exclusionSet, nodesToUpdate))

		assert.Lenf(t, nodesToUpdate, nodeCount, "expected all %d previously tagged nodes to be queued for kind removal", nodeCount)
		for _, nodeId := range previouslyTagged {
			_, present := nodesToUpdate[nodeId.Uint64()]
			assert.Truef(t, present, "expected node %d to be queued for kind removal", nodeId)
		}

		assert.Equal(t, 0.0, testutil.ToFloat64(pzNodeTagCounterVec.With(prometheus.Labels{"action": "tag_added", "position": "label"})))
		assert.Equal(t, float64(nodeCount), testutil.ToFloat64(pzNodeTagCounterVec.With(prometheus.Labels{"action": "tag_removed", "position": "label"})))
	})

	t.Run("preserves the tag on selected nodes and removes it from no longer selected nodes", func(t *testing.T) {
		pzNodeTagCounterVec.Reset()
		tag, err := bhDB.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeLabel, testActor, "regression label mixed", "", null.Int32{}, null.Bool{}, null.String{})
		require.NoError(t, err)

		selector, err := bhDB.CreateAssetGroupTagSelector(testCtx, tag.ID, testActor, "regression selector mixed", "", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "REGRESSION-MIXED-OBJECT-ID-0"},
		})
		require.NoError(t, err)

		var (
			tagKind          = tag.ToKind()
			previouslyTagged []graph.ID
		)

		require.NoError(t, assertGraphKinds(testCtx, graphDB, graph.Kinds{tagKind}))

		const nodeCount = 5
		require.NoError(t, graphDB.WriteTransaction(testCtx, func(tx graph.Transaction) error {
			for i := 0; i < nodeCount; i++ {
				node, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     fmt.Sprintf("regression-mixed-node-%d", i),
					common.ObjectID: fmt.Sprintf("REGRESSION-MIXED-OBJECT-ID-%d", i),
				}), ad.Entity, ad.User, tagKind)
				if err != nil {
					return err
				}
				previouslyTagged = append(previouslyTagged, node.ID)
			}
			return nil
		}))

		// Insert a selector node row for the first graph node so that it remains "selected" for this tag.
		// The remaining nodes have no selector_node row and should each have their tag kind removed.
		selectedNodeId := previouslyTagged[0]
		insertSelectorNodes(t, testCtx, bhDB, model.AssetGroupSelectorNode{
			SelectorId:      selector.ID,
			NodeId:          selectedNodeId,
			Certified:       model.AssetGroupCertificationManual,
			CertifiedBy:     null.StringFrom(model.AssetGroupActorBloodHound),
			Source:          model.AssetGroupSelectorNodeSourceSeed,
			NodePrimaryKind: ad.User.String(),
			NodeObjectId:    "REGRESSION-MIXED-OBJECT-ID-0",
			NodeName:        "regression-mixed-node-0",
		})

		var (
			exclusionSet  = cardinality.NewBitmap64()
			nodesToUpdate = make(map[uint64]*graph.Node)
		)

		require.NoError(t, tagAssetGroupNodesForTag(testCtx, bhDB, graphDB, tag, exclusionSet, nodesToUpdate))

		expectedRemovals := previouslyTagged[1:]
		assert.Lenf(t, nodesToUpdate, len(expectedRemovals), "expected %d previously tagged but no longer selected nodes to be queued for kind removal", len(expectedRemovals))
		for _, nodeId := range expectedRemovals {
			_, present := nodesToUpdate[nodeId.Uint64()]
			assert.Truef(t, present, "expected node %d to be queued for kind removal", nodeId)
		}
		_, selectedEnqueued := nodesToUpdate[selectedNodeId.Uint64()]
		assert.Falsef(t, selectedEnqueued, "expected selected node %d to NOT be queued for kind removal", selectedNodeId)

		assert.Equal(t, 0.0, testutil.ToFloat64(pzNodeTagCounterVec.With(prometheus.Labels{"action": "tag_added", "position": "label"})))
		assert.Equal(t, float64(len(expectedRemovals)), testutil.ToFloat64(pzNodeTagCounterVec.With(prometheus.Labels{"action": "tag_removed", "position": "label"})))
	})

	t.Run("adds and removes node tags for a tier tag type, and verifies the position label is calculated correctly", func(t *testing.T) {
		pzNodeTagCounterVec.Reset()

		const (
			tierPosition = 2
			addedCount   = 3 // selected by a selector but none with the tag exist yet in the graph
			removedCount = 2 // already carrying the tag kind in the graph but no longer selected
		)

		tag, err := bhDB.CreateAssetGroupTag(testCtx, model.AssetGroupTagTypeTier, testActor, "regression tier", "", null.Int32From(tierPosition), null.Bool{}, null.String{})
		require.NoError(t, err)

		selector, err := bhDB.CreateAssetGroupTagSelector(testCtx, tag.ID, testActor, "regression tier selector", "", false, true, model.SelectorAutoCertifyMethodDisabled, []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "REGRESSION-TIER-NO-MATCH"},
		})
		require.NoError(t, err)

		tagKind := tag.ToKind()
		require.NoError(t, assertGraphKinds(testCtx, graphDB, graph.Kinds{tagKind}))

		// Create nodes that are selected but do not yet carry the tag kind — these drive tag_added.
		var addedNodeIds []graph.ID
		require.NoError(t, graphDB.WriteTransaction(testCtx, func(tx graph.Transaction) error {
			for i := 0; i < addedCount; i++ {
				node, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     fmt.Sprintf("regression-tier-added-%d", i),
					common.ObjectID: fmt.Sprintf("REGRESSION-TIER-ADDED-ID-%d", i),
				}), ad.Entity, ad.User)
				if err != nil {
					return err
				}
				addedNodeIds = append(addedNodeIds, node.ID)
			}
			return nil
		}))

		// add these nodes to the selector so they are "selected" for this tag and should have the tag kind added
		for i, nodeId := range addedNodeIds {
			insertSelectorNodes(t, testCtx, bhDB, model.AssetGroupSelectorNode{
				SelectorId:      selector.ID,
				NodeId:          nodeId,
				Certified:       model.AssetGroupCertificationManual,
				CertifiedBy:     null.StringFrom(model.AssetGroupActorBloodHound),
				Source:          model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind: ad.User.String(),
				NodeObjectId:    fmt.Sprintf("REGRESSION-TIER-ADDED-ID-%d", i),
				NodeName:        fmt.Sprintf("regression-tier-added-%d", i),
			})
		}

		// Create nodes that already carry the tag kind but have no selector node row — these drive tag_removed.
		var removedNodeIds []graph.ID
		require.NoError(t, graphDB.WriteTransaction(testCtx, func(tx graph.Transaction) error {
			for i := 0; i < removedCount; i++ {
				node, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
					common.Name:     fmt.Sprintf("regression-tier-removed-%d", i),
					common.ObjectID: fmt.Sprintf("REGRESSION-TIER-REMOVED-ID-%d", i),
				}), ad.Entity, ad.User, tagKind)
				if err != nil {
					return err
				}
				removedNodeIds = append(removedNodeIds, node.ID)
			}
			return nil
		}))

		var (
			exclusionSet  = cardinality.NewBitmap64()
			nodesToUpdate = make(map[uint64]*graph.Node)
		)

		require.NoError(t, tagAssetGroupNodesForTag(testCtx, bhDB, graphDB, tag, exclusionSet, nodesToUpdate))

		assert.Lenf(t, nodesToUpdate, addedCount+removedCount, "expected %d nodes queued for kind update", addedCount+removedCount)

		assert.Equal(t, float64(addedCount), testutil.ToFloat64(pzNodeTagCounterVec.With(prometheus.Labels{"action": "tag_added", "position": "2"})))
		assert.Equal(t, float64(removedCount), testutil.ToFloat64(pzNodeTagCounterVec.With(prometheus.Labels{"action": "tag_removed", "position": "2"})))
	})
}

func TestSelectNodes(t *testing.T) {
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		testCtx       = suite.Context
		bhDB          = suite.BHDatabase
		graphDB       = suite.GraphDB
		testActor     = model.User{Unique: model.Unique{ID: uuid.FromStringOrNil("12345678-9012-4567-9012-456789012345")}}
		agtParameters = appcfg.GetAGTParameters(testCtx, bhDB)
	)

	primaryDisplayKinds, err := bhDB.GetPrimaryDisplayKinds(testCtx)
	require.NoError(t, err)

	t.Run("auto certify disabled leaves seed node certification pending", func(t *testing.T) {
		name := "select-nodes-disabled"
		node := insertGraphNode(t, testCtx, graphDB, name, ad.User)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes disabled", model.SelectorAutoCertifyMethodDisabled, createSelectorSeed(t, node))

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodNone)

		selectorNode := requireSelectorNode(t, testCtx, bhDB, selector.ID, node.ID)
		assert.Equal(t, model.AssetGroupCertificationPending, selectorNode.Certified)
		assert.False(t, selectorNode.CertifiedBy.Valid)
		assert.Equal(t, model.AssetGroupSelectorNodeSourceSeed, selectorNode.Source)
		assertSelectorNodeProperties(t, primaryDisplayKinds, node, selectorNode)
	})

	t.Run("auto certify all leaves all nodes certified", func(t *testing.T) {
		name := "select-nodes-all-members"
		groupNode, childNode := insertChildExpansionScenario(t, testCtx, graphDB, name)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes all members", model.SelectorAutoCertifyMethodAllMembers, createSelectorSeed(t, groupNode))

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodChildren)

		selectorNodes := requireSelectorNodesById(t, testCtx, bhDB, selector.ID)
		require.Len(t, selectorNodes, 2)

		groupSelectorNode := selectorNodes[groupNode.ID]
		assert.Equal(t, model.AssetGroupCertificationAuto, groupSelectorNode.Certified)
		assert.Equal(t, model.AssetGroupActorBloodHound, groupSelectorNode.CertifiedBy.String)
		assert.Equal(t, model.AssetGroupSelectorNodeSourceSeed, groupSelectorNode.Source)
		assertSelectorNodeProperties(t, primaryDisplayKinds, groupNode, groupSelectorNode)

		childSelectorNode := selectorNodes[childNode.ID]
		assert.Equal(t, model.AssetGroupCertificationAuto, childSelectorNode.Certified)
		assert.Equal(t, model.AssetGroupActorBloodHound, childSelectorNode.CertifiedBy.String)
		assert.Equal(t, model.AssetGroupSelectorNodeSourceChild, childSelectorNode.Source)
		assertSelectorNodeProperties(t, primaryDisplayKinds, childNode, childSelectorNode)
	})

	t.Run("auto certify seed only leaves child selector node certification pending", func(t *testing.T) {
		name := "select-nodes-seeds-only-child"
		groupNode, childNode := insertChildExpansionScenario(t, testCtx, graphDB, name)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes seeds only child", model.SelectorAutoCertifyMethodSeedsOnly, createSelectorSeed(t, groupNode))

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodChildren)

		selectorNodes := requireSelectorNodesById(t, testCtx, bhDB, selector.ID)
		require.Len(t, selectorNodes, 2)

		groupSelectorNode := selectorNodes[groupNode.ID]
		assert.Equal(t, model.AssetGroupCertificationAuto, groupSelectorNode.Certified)
		assert.Equal(t, model.AssetGroupActorBloodHound, groupSelectorNode.CertifiedBy.String)
		assert.Equal(t, model.AssetGroupSelectorNodeSourceSeed, groupSelectorNode.Source)

		childSelectorNode := selectorNodes[childNode.ID]
		assert.Equal(t, model.AssetGroupCertificationPending, childSelectorNode.Certified)
		assert.False(t, childSelectorNode.CertifiedBy.Valid)
		assert.Equal(t, model.AssetGroupSelectorNodeSourceChild, childSelectorNode.Source)
	})

	t.Run("auto certify seed only leaves parent selector node certification pending", func(t *testing.T) {
		name := "select-nodes-seeds-only-parent"
		parentNode, baseNode := insertParentExpansionScenario(t, testCtx, graphDB, name)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes seeds only parent", model.SelectorAutoCertifyMethodSeedsOnly, createSelectorSeed(t, baseNode))

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodParents)

		selectorNodes := requireSelectorNodesById(t, testCtx, bhDB, selector.ID)
		require.Len(t, selectorNodes, 2)

		seedSelectorNode := selectorNodes[baseNode.ID]
		assert.Equal(t, model.AssetGroupCertificationAuto, seedSelectorNode.Certified)
		assert.Equal(t, model.AssetGroupActorBloodHound, seedSelectorNode.CertifiedBy.String)
		assert.Equal(t, model.AssetGroupSelectorNodeSourceSeed, seedSelectorNode.Source)

		parentSelectorNode := selectorNodes[parentNode.ID]
		assert.Equal(t, model.AssetGroupCertificationPending, parentSelectorNode.Certified)
		assert.False(t, parentSelectorNode.CertifiedBy.Valid)
		assert.Equal(t, model.AssetGroupSelectorNodeSourceParent, parentSelectorNode.Source)
	})

	t.Run("updates stale selected node properties", func(t *testing.T) {
		name := "select-nodes-stale-properties"
		node := insertGraphNode(t, testCtx, graphDB, name, ad.User)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes stale properties", model.SelectorAutoCertifyMethodDisabled, createSelectorSeed(t, node))

		insertSelectorNodes(t, testCtx, bhDB, model.AssetGroupSelectorNode{
			SelectorId:        selector.ID,
			NodeId:            node.ID,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			NodePrimaryKind:   "stale-kind",
			NodeEnvironmentId: "stale-env",
			NodeObjectId:      "stale-object-id",
			NodeName:          "stale-name",
		})

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodNone)

		selectorNode := requireSelectorNode(t, testCtx, bhDB, selector.ID, node.ID)
		assert.Equal(t, model.AssetGroupCertificationPending, selectorNode.Certified)
		assert.False(t, selectorNode.CertifiedBy.Valid)
		assertSelectorNodeProperties(t, primaryDisplayKinds, node, selectorNode)
	})

	t.Run("preserves manual and revoked certifications while updating stale properties", func(t *testing.T) {
		manualName := "select-nodes-manual-preserve"
		manualNode := insertGraphNode(t, testCtx, graphDB, manualName, ad.User)
		revokedName := "select-nodes-revoked-preserve"
		revokedNode := insertGraphNode(t, testCtx, graphDB, revokedName, ad.User)
		selector := insertTagAndTagSelector(
			t,
			testCtx,
			bhDB,
			testActor,
			"select nodes preserve protected certifications",
			model.SelectorAutoCertifyMethodAllMembers,
			createSelectorSeed(t, manualNode),
			createSelectorSeed(t, revokedNode),
		)

		manualCertifiedBy := null.StringFrom("manual@example.com")
		revokedCertifiedBy := null.StringFrom("revoked@example.com")
		insertSelectorNodes(t, testCtx, bhDB,
			model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            manualNode.ID,
				Certified:         model.AssetGroupCertificationManual,
				CertifiedBy:       manualCertifiedBy,
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "stale-kind",
				NodeEnvironmentId: "stale-env",
				NodeObjectId:      "stale-object-id",
				NodeName:          "stale-name",
			},
			model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            revokedNode.ID,
				Certified:         model.AssetGroupCertificationRevoked,
				CertifiedBy:       revokedCertifiedBy,
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   "stale-kind",
				NodeEnvironmentId: "stale-env",
				NodeObjectId:      "stale-object-id",
				NodeName:          "stale-name",
			},
		)

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodNone)

		manualSelectorNode := requireSelectorNode(t, testCtx, bhDB, selector.ID, manualNode.ID)
		assert.Equal(t, model.AssetGroupCertificationManual, manualSelectorNode.Certified)
		assert.Equal(t, manualCertifiedBy, manualSelectorNode.CertifiedBy)
		assertSelectorNodeProperties(t, primaryDisplayKinds, manualNode, manualSelectorNode)

		revokedSelectorNode := requireSelectorNode(t, testCtx, bhDB, selector.ID, revokedNode.ID)
		assert.Equal(t, model.AssetGroupCertificationRevoked, revokedSelectorNode.Certified)
		assert.Equal(t, revokedCertifiedBy, revokedSelectorNode.CertifiedBy)
		assertSelectorNodeProperties(t, primaryDisplayKinds, revokedNode, revokedSelectorNode)
	})

	t.Run("downgrades existing automatic certification to pending when auto certification is disabled", func(t *testing.T) {
		name := "select-nodes-auto-disabled"
		node := insertGraphNode(t, testCtx, graphDB, name, ad.User)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes auto disabled", model.SelectorAutoCertifyMethodDisabled, createSelectorSeed(t, node))

		insertSelectorNodes(t, testCtx, bhDB, model.AssetGroupSelectorNode{
			SelectorId:        selector.ID,
			NodeId:            node.ID,
			Certified:         model.AssetGroupCertificationAuto,
			CertifiedBy:       null.StringFrom(model.AssetGroupActorBloodHound),
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			NodePrimaryKind:   ad.User.String(),
			NodeEnvironmentId: suffixDomainSID(name),
			NodeObjectId:      suffixObjectID(name),
			NodeName:          suffixName(name),
		})

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodNone)

		selectorNode := requireSelectorNode(t, testCtx, bhDB, selector.ID, node.ID)
		assert.Equal(t, model.AssetGroupCertificationPending, selectorNode.Certified)
		assert.False(t, selectorNode.CertifiedBy.Valid)
		assertSelectorNodeProperties(t, primaryDisplayKinds, node, selectorNode)
	})

	t.Run("deletes old selected nodes that are no longer selected", func(t *testing.T) {
		selectedName := "select-nodes-selected"
		unselectedName := "select-nodes-unselected"
		selectedNode := insertGraphNode(t, testCtx, graphDB, selectedName, ad.User)
		unselectedNode := insertGraphNode(t, testCtx, graphDB, unselectedName, ad.User)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes delete unselected", model.SelectorAutoCertifyMethodDisabled, createSelectorSeed(t, selectedNode))

		insertSelectorNodes(t, testCtx, bhDB,
			model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            selectedNode.ID,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   ad.User.String(),
				NodeEnvironmentId: suffixDomainSID(selectedName),
				NodeObjectId:      suffixObjectID(selectedName),
				NodeName:          suffixName(selectedName),
			},
			model.AssetGroupSelectorNode{
				SelectorId:        selector.ID,
				NodeId:            unselectedNode.ID,
				Certified:         model.AssetGroupCertificationPending,
				CertifiedBy:       null.String{},
				Source:            model.AssetGroupSelectorNodeSourceSeed,
				NodePrimaryKind:   ad.User.String(),
				NodeEnvironmentId: suffixDomainSID(unselectedName),
				NodeObjectId:      suffixObjectID(unselectedName),
				NodeName:          suffixName(unselectedName),
			},
		)

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodNone)

		selectorNodes := requireSelectorNodesById(t, testCtx, bhDB, selector.ID)
		require.Len(t, selectorNodes, 1)
		assert.Contains(t, selectorNodes, selectedNode.ID)
		assert.NotContains(t, selectorNodes, unselectedNode.ID)
	})

	t.Run("deletes all old selected nodes when no seeds resolve", func(t *testing.T) {
		name := "select-nodes-missing-seed-old"
		oldNode := insertGraphNode(t, testCtx, graphDB, name, ad.User)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes missing seed", model.SelectorAutoCertifyMethodDisabled,
			model.SelectorSeed{
				Type:  model.SelectorTypeObjectId,
				Value: "lorem-ipsum",
			},
		)

		insertSelectorNodes(t, testCtx, bhDB, model.AssetGroupSelectorNode{
			SelectorId:        selector.ID,
			NodeId:            oldNode.ID,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceSeed,
			NodePrimaryKind:   ad.User.String(),
			NodeEnvironmentId: suffixDomainSID(name),
			NodeObjectId:      suffixObjectID(name),
			NodeName:          suffixName(name),
		})

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodNone)

		assert.Empty(t, requireSelectorNodesById(t, testCtx, bhDB, selector.ID))
	})

	t.Run("auto certify seeds only certifies seed when it was previously a child", func(t *testing.T) {
		name := "select-nodes-stale-source"
		node := insertGraphNode(t, testCtx, graphDB, name, ad.User)
		selector := insertTagAndTagSelector(t, testCtx, bhDB, testActor, "select nodes stale source", model.SelectorAutoCertifyMethodSeedsOnly, createSelectorSeed(t, node))

		insertSelectorNodes(t, testCtx, bhDB, model.AssetGroupSelectorNode{
			SelectorId:        selector.ID,
			NodeId:            node.ID,
			Certified:         model.AssetGroupCertificationPending,
			CertifiedBy:       null.String{},
			Source:            model.AssetGroupSelectorNodeSourceChild,
			NodePrimaryKind:   ad.User.String(),
			NodeEnvironmentId: suffixDomainSID(name),
			NodeObjectId:      suffixObjectID(name),
			NodeName:          suffixName(name),
		})

		runSelectNodes(t, testCtx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, model.AssetGroupExpansionMethodNone)

		selectorNode := requireSelectorNode(t, testCtx, bhDB, selector.ID, node.ID)
		assert.Equal(t, model.AssetGroupCertificationAuto, selectorNode.Certified)
		assert.Equal(t, model.AssetGroupActorBloodHound, selectorNode.CertifiedBy.String)
	})
}

func insertTagAndTagSelector(t *testing.T, ctx context.Context, bhDB interface {
	CreateAssetGroupTag(ctx context.Context, assetGroupTagType model.AssetGroupTagType, user model.User, name, description string, position null.Int32, requireCertify null.Bool, glyph null.String) (model.AssetGroupTag, error)
	CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, user model.User, name string, description string, isDefault bool, allowDisable bool, autoCertify model.SelectorAutoCertifyMethod, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error)
}, testActor model.User, name string, autoCertify model.SelectorAutoCertifyMethod, seeds ...model.SelectorSeed) model.AssetGroupTagSelector {
	t.Helper()

	tag, err := bhDB.CreateAssetGroupTag(ctx, model.AssetGroupTagTypeLabel, testActor, name+" tag", "", null.Int32{}, null.Bool{}, null.String{})
	require.NoError(t, err)

	selector, err := bhDB.CreateAssetGroupTagSelector(ctx, tag.ID, testActor, name, "", false, true, autoCertify, seeds)
	require.NoError(t, err)

	return selector
}

func insertGraphNode(t *testing.T, ctx context.Context, graphDB graph.Database, name string, kinds ...graph.Kind) *graph.Node {
	t.Helper()

	var node *graph.Node
	objectID := suffixObjectID(name)

	require.NoError(t, graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		createdNode, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     suffixName(name),
			common.ObjectID: objectID,
			ad.DomainSID:    suffixDomainSID(name),
		}), append([]graph.Kind{ad.Entity}, kinds...)...)
		if err != nil {
			return err
		}

		node = createdNode
		return nil
	}))

	return node
}

func insertChildExpansionScenario(t *testing.T, ctx context.Context, graphDB graph.Database, name string) (*graph.Node, *graph.Node) {
	t.Helper()

	var (
		groupNode *graph.Node
		childNode *graph.Node
	)

	require.NoError(t, graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if groupNode, err = tx.CreateNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     suffixName(name + "-group"),
			common.ObjectID: suffixObjectID(name + "-group"),
			ad.DomainSID:    suffixDomainSID(name),
		}), ad.Entity, ad.Group); err != nil {
			return err
		}

		if childNode, err = tx.CreateNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     suffixName(name + "-child"),
			common.ObjectID: suffixObjectID(name + "-child"),
			ad.DomainSID:    suffixDomainSID(name),
		}), ad.Entity, ad.User); err != nil {
			return err
		}

		_, err = tx.CreateRelationshipByIDs(childNode.ID, groupNode.ID, ad.MemberOf, graph.NewProperties())
		return err
	}))

	return groupNode, childNode
}

func insertParentExpansionScenario(t *testing.T, ctx context.Context, graphDB graph.Database, name string) (*graph.Node, *graph.Node) {
	t.Helper()

	var (
		parentNode *graph.Node
		childNode  *graph.Node
	)

	require.NoError(t, graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if parentNode, err = tx.CreateNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     suffixName(name + "-parent"),
			common.ObjectID: suffixObjectID(name + "-parent"),
			ad.DomainSID:    suffixDomainSID(name),
		}), ad.Entity, ad.OU); err != nil {
			return err
		}

		if childNode, err = tx.CreateNode(graph.AsProperties(graph.PropertyMap{
			common.Name:       suffixName(name + "-base"),
			common.ObjectID:   suffixObjectID(name + "-base"),
			ad.DomainSID:      suffixDomainSID(name),
			ad.IsACLProtected: false,
		}), ad.Entity, ad.User); err != nil {
			return err
		}

		_, err = tx.CreateRelationshipByIDs(parentNode.ID, childNode.ID, ad.Contains, graph.NewProperties())
		return err
	}))

	return parentNode, childNode
}

func runSelectNodes(t *testing.T, ctx context.Context, bhDB database.Database, graphDB graph.Database, agtParameters appcfg.AGTParameters, primaryDisplayKinds schema.PrimaryDisplayKinds, selector model.AssetGroupTagSelector, expansionMethod model.AssetGroupExpansionMethod) {
	t.Helper()

	require.Empty(t, SelectNodes(ctx, bhDB, graphDB, agtParameters, primaryDisplayKinds, selector, expansionMethod))
}

func insertSelectorNodes(t *testing.T, ctx context.Context, bhDB interface {
	InsertSelectorNodes(ctx context.Context, nodes []model.AssetGroupSelectorNode) error
}, nodes ...model.AssetGroupSelectorNode) {
	t.Helper()

	require.NoError(t, bhDB.InsertSelectorNodes(ctx, nodes))
}

func requireSelectorNode(t *testing.T, ctx context.Context, bhDB interface {
	GetSelectorNodesBySelectorIds(ctx context.Context, selectorIds ...int) ([]model.AssetGroupSelectorNode, error)
}, selectorId int, nodeId graph.ID) model.AssetGroupSelectorNode {
	t.Helper()

	selectorNodes := requireSelectorNodesById(t, ctx, bhDB, selectorId)
	selectorNode, ok := selectorNodes[nodeId]
	require.Truef(t, ok, "expected selector %d to include node %d", selectorId, nodeId)

	return selectorNode
}

func requireSelectorNodesById(t *testing.T, ctx context.Context, bhDB interface {
	GetSelectorNodesBySelectorIds(ctx context.Context, selectorIds ...int) ([]model.AssetGroupSelectorNode, error)
}, selectorId int) map[graph.ID]model.AssetGroupSelectorNode {
	t.Helper()

	selectorNodes, err := bhDB.GetSelectorNodesBySelectorIds(ctx, selectorId)
	require.NoError(t, err)

	selectorNodesById := make(map[graph.ID]model.AssetGroupSelectorNode, len(selectorNodes))
	for _, selectorNode := range selectorNodes {
		selectorNodesById[selectorNode.NodeId] = selectorNode
	}

	return selectorNodesById
}

func assertSelectorNodeProperties(t *testing.T, primaryDisplayKinds schema.PrimaryDisplayKinds, node *graph.Node, selectorNode model.AssetGroupSelectorNode) {
	t.Helper()

	primaryKind, displayName, objectId, envId := model.GetAssetGroupMemberProperties(primaryDisplayKinds, node)
	assert.Equal(t, primaryKind, selectorNode.NodePrimaryKind)
	assert.Equal(t, displayName, selectorNode.NodeName)
	assert.Equal(t, objectId, selectorNode.NodeObjectId)
	assert.Equal(t, envId, selectorNode.NodeEnvironmentId)
}

func createSelectorSeed(t *testing.T, node *graph.Node) model.SelectorSeed {
	t.Helper()

	objectId, err := node.Properties.Get(common.ObjectID.String()).String()
	require.NoError(t, err)

	return model.SelectorSeed{
		Type:  model.SelectorTypeObjectId,
		Value: objectId,
	}
}

func suffixName(str string) string {
	return str + "-name"
}

func suffixObjectID(str string) string {
	return str + "-object-id"
}

func suffixDomainSID(str string) string {
	return str + "-domain-sid"
}
