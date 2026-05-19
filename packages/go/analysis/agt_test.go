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
		require.NoError(t, bhDB.InsertSelectorNode(
			testCtx,
			tag.ID,
			selector.ID,
			selectedNodeId,
			model.AssetGroupCertificationManual,
			null.StringFrom(model.AssetGroupActorBloodHound),
			model.AssetGroupSelectorNodeSourceSeed,
			ad.User.String(),
			"",
			"REGRESSION-MIXED-OBJECT-ID-0",
			"regression-mixed-node-0",
		))

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
}
