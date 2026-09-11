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

//go:build serial_integration

package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/ad/internal/nodeprops"
	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/lab/arrows"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"

	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"

	"github.com/stretchr/testify/require"
)

type linkWellKnownNodesTestCase struct {
	name      string
	setupFunc func(
		t *testing.T,
		ctx context.Context,
		graphDB graph.Database,
	) *graph.Node
	assertionFunc func(
		t *testing.T,
		ctx context.Context,
		graphDB graph.Database,
		expectedNode *graph.Node,
	)
}

func TestLinkWellKnownNodes(t *testing.T) {
	wellKnownNodes := []struct {
		sidSuffix      wellknown.SIDSuffix
		nodeNamePrefix wellknown.NodeNamePrefix
		kind           graph.Kind
	}{
		{
			sidSuffix:      wellknown.DomainUsersSIDSuffix,
			nodeNamePrefix: wellknown.DomainUsersNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.AuthenticatedUsersSIDSuffix,
			nodeNamePrefix: wellknown.AuthenticatedUsersNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.EveryoneSIDSuffix,
			nodeNamePrefix: wellknown.EveryoneNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.DomainComputersSIDSuffix,
			nodeNamePrefix: wellknown.DomainComputerNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.GuestSIDSuffix,
			nodeNamePrefix: wellknown.GuestNodeNamePrefix,
			kind:           ad.User,
		},
		{
			sidSuffix:      wellknown.NetworkSIDSuffix,
			nodeNamePrefix: wellknown.NetworkNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.ThisOrganizationSIDSuffix,
			nodeNamePrefix: wellknown.ThisOrganizationNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.ThisOrganizationCertificateSIDSuffix,
			nodeNamePrefix: wellknown.ThisOrganizationCertificateNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.AuthenticationAuthorityAssertedIdentitySIDSuffix,
			nodeNamePrefix: wellknown.AuthenticationAuthorityAssertedIdentityNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.KeyTrustSIDSuffix,
			nodeNamePrefix: wellknown.KeyTrustNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.MFAKeyPropertySIDSuffix,
			nodeNamePrefix: wellknown.MFAKeyPropertyNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.NTLMAuthenticationSIDSuffix,
			nodeNamePrefix: wellknown.NTLMAuthenticationNodeNamePrefix,
			kind:           ad.Group,
		},
		{
			sidSuffix:      wellknown.SchannelAuthenticationSIDSuffix,
			nodeNamePrefix: wellknown.SchannelAuthenticationNodeNamePrefix,
			kind:           ad.Group,
		},
	}
	wellKnownNodesDomain1 := make(map[wellknown.NodeNamePrefix]*graph.Node)
	wellKnownNodesDomain2 := make(map[wellknown.NodeNamePrefix]*graph.Node)
	wellKnownNodesDomain3 := make(map[wellknown.NodeNamePrefix]*graph.Node)

	testCases := []linkWellKnownNodesTestCase{
		{
			name: "Verifies that linking all well-known groups succeeds when they already exist.",
			setupFunc: func(
				t *testing.T,
				ctx context.Context,
				graphDB graph.Database,
			) *graph.Node {
				// NOTE: Testing the scenario requires created the wellknown groups ahead of time asserting their
				// execution when asserting the scenario prior to asserting the expected outcome for
				// LinkWellKnownNodes
				domain1Node := createNode(
					t,
					ctx,
					graphDB,
					generateCollectedDomain("DOMAIN1.LOCAL", "S-1-5-21-1004336348-1177238915-682003330"),
				)
				domain2Node := createNode(
					t,
					ctx,
					graphDB,
					generateCollectedDomain("DOMAIN2.LOCAL", "S-1-5-21-1004336348-1177238915-682003331"),
				)
				domain3Node := createNode(
					t,
					ctx,
					graphDB,
					generateCollectedDomain("DOMAIN3.LOCAL", "S-1-5-21-1004336348-1177238915-682003332"),
				)
				for _, wellKnownNode := range wellKnownNodes {
					// Domain1
					createdWellKnownNode := createNode(
						t,
						ctx,
						graphDB,
						generateWellKnownNode(
							t,
							domain1Node,
							wellKnownNode.sidSuffix,
							wellKnownNode.nodeNamePrefix,
							wellKnownNode.kind,
						),
					)
					wellKnownNodesDomain1[wellKnownNode.nodeNamePrefix] = createdWellKnownNode
					assertNodeExists(
						t,
						ctx,
						graphDB,
						createdWellKnownNode,
						query.StringEndsWith(
							query.NodeProperty(common.ObjectID.String()),
							wellKnownNode.sidSuffix.String(),
						),
						query.StringStartsWith(
							query.NodeProperty(common.Name.String()),
							wellKnownNode.nodeNamePrefix.String(),
						),
					)

					// Domain2
					createdWellKnownNode2 := createNode(
						t,
						ctx,
						graphDB,
						generateWellKnownNode(
							t,
							domain2Node,
							wellKnownNode.sidSuffix,
							wellKnownNode.nodeNamePrefix,
							wellKnownNode.kind,
						),
					)
					wellKnownNodesDomain2[wellKnownNode.nodeNamePrefix] = createdWellKnownNode2
					assertNodeExists(
						t,
						ctx,
						graphDB,
						createdWellKnownNode2,
						query.StringEndsWith(
							query.NodeProperty(common.ObjectID.String()),
							wellKnownNode.sidSuffix.String(),
						),
						query.StringStartsWith(
							query.NodeProperty(common.Name.String()),
							wellKnownNode.nodeNamePrefix.String(),
						),
					)

					// Domain3
					createdWellKnownNode3 := createNode(
						t,
						ctx,
						graphDB,
						generateWellKnownNode(
							t,
							domain3Node,
							wellKnownNode.sidSuffix,
							wellKnownNode.nodeNamePrefix,
							wellKnownNode.kind,
						),
					)
					wellKnownNodesDomain3[wellKnownNode.nodeNamePrefix] = createdWellKnownNode3
					assertNodeExists(
						t,
						ctx,
						graphDB,
						createdWellKnownNode3,
						query.StringEndsWith(
							query.NodeProperty(common.ObjectID.String()),
							wellKnownNode.sidSuffix.String(),
						),
						query.StringStartsWith(
							query.NodeProperty(common.Name.String()),
							wellKnownNode.nodeNamePrefix.String(),
						),
					)
				}

				//Trust edges
				err := createEdge(t, ctx, graphDB, domain2Node, domain1Node, ad.SameForestTrust)
				require.NoError(t, err)
				err = createEdge(t, ctx, graphDB, domain3Node, domain1Node, ad.CrossForestTrust)
				require.NoError(t, err)

				return domain1Node
			},
			assertionFunc: func(
				t *testing.T,
				ctx context.Context,
				graphDB graph.Database,
				expectedNode *graph.Node,
			) {
				// assert that the relationships exist
				expectationsMemberOfDomain1 := map[string][]*graph.Node{
					"domain users node is linked to authenticated users node": {
						wellKnownNodesDomain1[wellknown.DomainUsersNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.AuthenticatedUsersNodeNamePrefix],
					},
					"domain computers node is linked to authenticated users node": {
						wellKnownNodesDomain1[wellknown.DomainComputerNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.AuthenticatedUsersNodeNamePrefix],
					},
					"authenticated users node is linked to everyone node": {
						wellKnownNodesDomain1[wellknown.AuthenticatedUsersNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
					},
					"guest node is linked to everyone node": {
						wellKnownNodesDomain1[wellknown.GuestNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
					},
				}
				expectationsClaimSpecialIdentityDomain1 := map[string][]*graph.Node{
					"everyone node is linked to network node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.NetworkNodeNamePrefix],
					},
					"everyone node is linked to this organization node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.ThisOrganizationNodeNamePrefix],
					},
					"everyone node is linked to this organization certificate node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.ThisOrganizationCertificateNodeNamePrefix],
					},
					"everyone node is linked to authentication authority asserted identity node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.AuthenticationAuthorityAssertedIdentityNodeNamePrefix],
					},
					"everyone node is linked to key trust node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.KeyTrustNodeNamePrefix],
					},
					"everyone node is linked to mfa key property node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.MFAKeyPropertyNodeNamePrefix],
					},
					"everyone node is linked to ntlm authentication node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.NTLMAuthenticationNodeNamePrefix],
					},
					"everyone node is linked to schannel authentication node": {
						wellKnownNodesDomain1[wellknown.EveryoneNodeNamePrefix],
						wellKnownNodesDomain1[wellknown.SchannelAuthenticationNodeNamePrefix],
					},
				}

				for name, nodes := range expectationsMemberOfDomain1 {
					var expectedRelationship *graph.Relationship
					err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
						rel, err := tx.Relationships().Filterf(func() graph.Criteria {
							return query.And(
								query.Equals(query.StartID(), nodes[0].ID),
								query.Equals(query.EndID(), nodes[1].ID),
								query.Kind(query.Relationship(), ad.MemberOf),
							)
						}).First()
						expectedRelationship = rel
						return err
					})
					require.NoError(t, err, name)
					require.NotNil(t, expectedRelationship, name)
				}
				for name, nodes := range expectationsClaimSpecialIdentityDomain1 {
					var expectedRelationship *graph.Relationship
					err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
						rel, err := tx.Relationships().Filterf(func() graph.Criteria {
							return query.And(
								query.Equals(query.StartID(), nodes[0].ID),
								query.Equals(query.EndID(), nodes[1].ID),
								query.Kind(query.Relationship(), ad.ClaimSpecialIdentity),
							)
						}).First()
						expectedRelationship = rel
						return err
					})
					require.NoError(t, err, name)
					require.NotNil(t, expectedRelationship, name)
				}
			},
		},
	}

	// Run tests for each database driver
	ctx := context.Background()
	// Initialize the database
	graphDB := integration.OpenGraphDB(t, graphschema.DefaultGraphSchema())
	defer graphDB.Close(ctx)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdNode := tc.setupFunc(
				t,
				ctx,
				graphDB,
			)
			require.NotNil(t, createdNode)

			// Run LinkWellKnownNodes
			err := adAnalysis.LinkWellKnownNodes(ctx, graphDB)
			require.NoError(t, err)

			tc.assertionFunc(t, ctx, graphDB, createdNode)
			require.NoError(t, err)
		})
	}
}

// This test ensures that only AD nodes are given the ad.Group kind when they have an object ID matching one of the well-known suffixes
func TestFixWellKnownNodeTypes(t *testing.T) {
	adNodeData := graph.Node{
		Kinds: graph.Kinds{
			ad.Entity,
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			common.ObjectID: wellknown.DomainAdminsGroupSIDSuffix.PrependPrefix("Test AD Node"),
		}),
	}
	azNodeData := graph.Node{
		Kinds: graph.Kinds{
			azure.Entity,
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			common.ObjectID: wellknown.DomainAdminsGroupSIDSuffix.PrependPrefix("Test AZ Node"),
		}),
	}
	ogNodeData := graph.Node{
		Kinds: graph.Kinds{
			graph.StringKind("OGBaseKind"),
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			common.ObjectID: wellknown.DomainAdminsGroupSIDSuffix.PrependPrefix("Test OG Node"),
		}),
	}

	ctx := context.Background()

	graphDB := integration.OpenGraphDB(t, graphschema.DefaultGraphSchema())
	defer graphDB.Close(ctx)

	var (
		adNode *graph.Node
		azNode *graph.Node
		ogNode *graph.Node
	)

	err := graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if adNode, err = tx.CreateNode(adNodeData.Properties, adNodeData.Kinds...); err != nil {
			return err
		} else if azNode, err = tx.CreateNode(azNodeData.Properties, azNodeData.Kinds...); err != nil {
			return err
		} else if ogNode, err = tx.CreateNode(ogNodeData.Properties, ogNodeData.Kinds...); err != nil {
			return err
		}
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, adNode)
	require.NotNil(t, azNode)
	require.NotNil(t, ogNode)

	err = adAnalysis.FixWellKnownNodeTypes(ctx, graphDB)
	require.NoError(t, err)

	var (
		adResult *graph.Node
		azResult *graph.Node
		ogResult *graph.Node
	)

	err = graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if adResult, err = tx.Nodes().Filter(query.Equals(query.NodeID(), adNode.ID)).First(); err != nil {
			return err
		} else if azResult, err = tx.Nodes().Filter(query.Equals(query.NodeID(), azNode.ID)).First(); err != nil {
			return err
		} else if ogResult, err = tx.Nodes().Filter(query.Equals(query.NodeID(), ogNode.ID)).First(); err != nil {
			return err
		}
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, adResult)
	require.NotNil(t, azResult)
	require.NotNil(t, ogResult)

	require.True(t, adResult.Kinds.ContainsOneOf(ad.Group))
	require.False(t, azResult.Kinds.ContainsOneOf(ad.Group))
	require.False(t, ogResult.Kinds.ContainsOneOf(ad.Group))
}

func fetchNode(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	cypherComparisons ...*cypher.Comparison,
) (
	*graph.Node,
	error,
) {
	t.Helper()

	var fetchedNode *graph.Node
	err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		filteredNode, err := tx.Nodes().Filterf(func() graph.Criteria {
			graphCriterias := make([]graph.Criteria, len(cypherComparisons))
			for i := range cypherComparisons {
				graphCriterias[i] = cypherComparisons[i]
			}
			return query.And(graphCriterias...)
		}).First()
		if err != nil {
			return err
		}
		fetchedNode = filteredNode

		return nil
	})
	return fetchedNode, err
}

func assertNodeExists(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	expectedNode *graph.Node,
	queryCriterias ...*cypher.Comparison,
) {
	t.Helper()

	domainSID, err := expectedNode.Properties.Get(ad.DomainSID.String()).String()
	require.NoError(t, err)
	domainFQDN, err := expectedNode.Properties.Get(ad.DomainFQDN.String()).String()
	require.NoError(t, err)
	nodeName, err := expectedNode.Properties.Get(common.Name.String()).String()
	require.NoError(t, err)

	defaultQueryCriterias := []*cypher.Comparison{
		query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
		query.Equals(query.NodeProperty(ad.DomainFQDN.String()), domainFQDN),
		query.Equals(query.NodeProperty(common.Name.String()), nodeName),
	}

	queryCriterias = append(queryCriterias, defaultQueryCriterias...)
	fetchedNode, err := fetchNode(t, ctx, graphDB, queryCriterias...)
	require.NoError(t, err)
	require.NotNil(t, fetchedNode)
}

func createNode(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	nodeToCreate *graph.Node,
) *graph.Node {
	t.Helper()
	require.NotNil(t, graphDB)
	require.NotNil(t, nodeToCreate)

	if ctx == nil {
		ctx = context.Background()
	}

	var createdNode *graph.Node
	var err error
	err = graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		createdNode, err = tx.CreateNode(nodeToCreate.Properties, nodeToCreate.Kinds...)
		if err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, createdNode)

	assertNodeExists(t, ctx, graphDB, createdNode)
	return createdNode
}

func createEdge(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	startNode *graph.Node,
	endNode *graph.Node,
	edgeKind graph.Kind,
) error {
	t.Helper()
	require.NotNil(t, graphDB)
	require.NotNil(t, startNode)
	require.NotNil(t, endNode)

	if ctx == nil {
		ctx = context.Background()
	}

	var createdEdge *graph.Relationship
	var err error
	err = graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		createdEdge, err = tx.CreateRelationshipByIDs(startNode.ID, endNode.ID, edgeKind, graph.NewProperties())
		if err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, createdEdge)

	return nil
}

func generateCollectedDomain(name, sid string) *graph.Node {
	return &graph.Node{
		Kinds: graph.Kinds{
			ad.Domain, ad.Entity,
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			common.Collected: true,
			common.Name:      name,
			ad.DomainSID:     sid,
			ad.DomainFQDN:    name,
		}),
	}
}

func generateWellKnownNode(
	t *testing.T,
	domainNode *graph.Node,
	sidSuffix wellknown.SIDSuffix,
	nodeNamePrefix wellknown.NodeNamePrefix,
	nodeKind graph.Kind,
) *graph.Node {
	require.NotNil(t, domainNode)
	require.NotNil(t, sidSuffix)
	require.NotNil(t, nodeNamePrefix)

	domainSID, domainName, err := nodeprops.ReadDomainIDandNameAsString(domainNode)
	require.NoError(t, err)

	var wellKnownSID string
	switch sidSuffix.String() {
	case wellknown.DomainUsersSIDSuffix.String(), wellknown.DomainComputersSIDSuffix.String(), wellknown.GuestSIDSuffix.String():
		wellKnownSID = sidSuffix.PrependPrefix(domainSID)
	default:
		wellKnownSID = sidSuffix.PrependPrefix(domainName)
	}

	return &graph.Node{
		Kinds: graph.Kinds{
			ad.Entity,
			nodeKind,
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			ad.DomainFQDN:    domainName,
			ad.DomainSID:     domainSID,
			common.Collected: true,
			common.Name:      nodeNamePrefix.AppendSuffix(domainName),
			common.ObjectID:  wellKnownSID,
		}),
	}
}

func TestFetchEnforcedGPOs(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		// Check the first user
		var (
			enforcedGPOs, err = adAnalysis.FetchEnforcedGPOs(tx, harness.GPOEnforcement.UserC, 0, 0)
		)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, enforcedGPOs.Len())

		// Check the second user
		enforcedGPOs, err = adAnalysis.FetchEnforcedGPOs(tx, harness.GPOEnforcement.UserB, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, enforcedGPOs.Len())
	})
}

func TestFetchEnforcedGPOsPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		// OU A blocks inheritance, but is contained by the domain GPLinked by both GPOs. We should see both GPOs in this path.
		path, err := adAnalysis.FetchEnforcedGPOsPaths(context.Background(), db, harness.GPOEnforcement.OrganizationalUnitA)
		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)

		// OU C is contained by OU A which blocks inheritance - so we should only see the enforced GPO in this path.
		path, err = adAnalysis.FetchEnforcedGPOsPaths(context.Background(), db, harness.GPOEnforcement.OrganizationalUnitC)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)

		// OU D is contained by OU B which does not block inheritance - so we should see both GPOs in this path.
		path, err = adAnalysis.FetchEnforcedGPOsPaths(context.Background(), db, harness.GPOEnforcement.OrganizationalUnitD)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitD.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)

		// User C is contained by OU C which is contained by OU A - OU A blocks inheritance it should only be affected by the enforced GPO.
		path, err = adAnalysis.FetchEnforcedGPOsPaths(context.Background(), db, harness.GPOEnforcement.UserC)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.UserC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)

		// User D is contained by OU D which is contained by OU B - none of them block inheritance so it should be affected by both GPOs.
		path, err = adAnalysis.FetchEnforcedGPOsPaths(context.Background(), db, harness.GPOEnforcement.UserD)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Equal(t, 6, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.UserD.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitD.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)
	})
}

func TestFetchGPOAffectedContainerPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		containers, err := adAnalysis.FetchGPOAffectedContainerPaths(tx, harness.GPOEnforcement.GPOEnforced)

		test.RequireNilErr(t, err)
		nodes := containers.AllNodes().IDs()
		require.Equal(t, 6, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitD.ID)

		containers, err = adAnalysis.FetchGPOAffectedContainerPaths(tx, harness.GPOEnforcement.GPOUnenforced)
		test.RequireNilErr(t, err)
		nodes = containers.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitD.ID)
	})
}

func TestCreateGPOAffectedIntermediariesListDelegateAffectedContainers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		containers, err := adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter)(tx, harness.GPOEnforcement.GPOEnforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 5, containers.Len())
		require.Equal(t, 4, containers.ContainingNodeKinds(ad.OU).Len())
		require.Equal(t, 1, containers.ContainingNodeKinds(ad.Domain).Len())

		containers, err = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter)(tx, harness.GPOEnforcement.GPOUnenforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 4, containers.Len())
		require.False(t, containers.Contains(harness.GPOEnforcement.OrganizationalUnitC))
		require.Equal(t, 3, containers.ContainingNodeKinds(ad.OU).Len())
		require.Equal(t, 1, containers.ContainingNodeKinds(ad.Domain).Len())
	})
}

func TestCreateGPOAffectedIntermediariesPathDelegateAffectedUsers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		users, err := adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.User)(tx, harness.GPOEnforcement.GPOEnforced)

		test.RequireNilErr(t, err)
		nodes := users.AllNodes().IDs()
		require.Equal(t, 10, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserD.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserA.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)

		users, err = adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.User)(tx, harness.GPOEnforcement.GPOUnenforced)
		test.RequireNilErr(t, err)
		nodes = users.AllNodes().IDs()
		require.Equal(t, 8, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)
		require.NotContains(t, nodes, harness.GPOEnforcement.UserC.ID)
		require.NotContains(t, nodes, harness.GPOEnforcement.OrganizationalUnitC.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserD.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserB.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.UserA.ID)
	})
}

func TestCreateGPOAffectedResultsListDelegateAffectedUsers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		users, err := adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter)(tx, harness.GPOEnforcement.GPOEnforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 4, users.Len())

		users, err = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter)(tx, harness.GPOEnforcement.GPOUnenforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 3, users.Len())
		require.Equal(t, 3, users.ContainingNodeKinds(ad.User).Len())
	})
}

func TestCreateGPOAffectedIntermediariesListDelegateTierZero(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		harness.GPOEnforcement.UserC.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)
		harness.GPOEnforcement.UserD.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)
		tx.UpdateNode(harness.GPOEnforcement.UserC)
		tx.UpdateNode(harness.GPOEnforcement.UserD)

		users, err := adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter)(tx, harness.GPOEnforcement.GPOEnforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, users.Len())

		users, err = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter)(tx, harness.GPOEnforcement.GPOUnenforced, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, users.Len())
	})
}

func TestFetchComputerSessionPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		sessions, err := adAnalysis.FetchComputerSessionPaths(tx, harness.Session.ComputerA)

		test.RequireNilErr(t, err)
		nodes := sessions.AllNodes().IDs()
		require.Equal(t, 2, len(nodes))
		require.Contains(t, nodes, harness.Session.ComputerA.ID)
		require.Contains(t, nodes, harness.Session.User.ID)
	})
}

func TestFetchComputerSessions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		sessions, err := adAnalysis.FetchComputerSessions(tx, harness.Session.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, sessions.Len())
	})
}

func TestFetchGroupSessionPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.FetchGroupSessionPaths(tx, harness.Session.GroupA)

		test.RequireNilErr(t, err)
		nodes := computers.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))

		nestedComputers, err := adAnalysis.FetchGroupSessionPaths(tx, harness.Session.GroupC)

		test.RequireNilErr(t, err)
		nestedNodes := nestedComputers.AllNodes().IDs()
		require.Equal(t, 5, len(nestedNodes))
	})
}

func TestFetchGroupSessions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.FetchGroupSessions(tx, harness.Session.GroupA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, computers.Len())
		require.Equal(t, 2, computers.ContainingNodeKinds(ad.Computer).Len())

		nestedComputers, err := adAnalysis.FetchGroupSessions(tx, harness.Session.GroupC, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, nestedComputers.Len())
		require.Equal(t, 2, nestedComputers.ContainingNodeKinds(ad.Computer).Len())
	})
}

func TestFetchUserSessionPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		sessions, err := adAnalysis.FetchUserSessionPaths(tx, harness.Session.User)

		test.RequireNilErr(t, err)
		nodes := sessions.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.Session.User.ID)
		require.Contains(t, nodes, harness.Session.ComputerA.ID)
		require.Contains(t, nodes, harness.Session.ComputerB.ID)
	})
}

func TestFetchUserSessions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.Session.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.FetchUserSessions(tx, harness.Session.User, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, computers.Len())
		require.Equal(t, 2, computers.ContainingNodeKinds(ad.Computer).Len())
	})
}

func TestCreateOutboundLocalGroupPathDelegateUser(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.UserB)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 3, len(nodes))
	})
}

func TestCreateOutboundLocalGroupListDelegateUser(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.UserB, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, computers.Slice()[0].ID)
	})
}

func TestCreateOutboundLocalGroupPathDelegateGroup(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.GroupA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))
	})
}

func TestCreateOutboundLocalGroupListDelegateGroup(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.GroupA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, computers.Slice()[0].ID)
	})
}

func TestCreateOutboundLocalGroupPathDelegateComputer(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerC.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupB.ID)
		require.Equal(t, 4, len(nodes))
	})
}

func TestCreateOutboundLocalGroupListDelegateComputer(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		computers, err := adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, computers.Len())
	})
}

func TestCreateInboundLocalGroupPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateInboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.UserA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 4, len(nodes))

		path, err = adAnalysis.CreateInboundLocalGroupPathDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerC)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.GroupB.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerC.ID)
		require.Equal(t, 3, len(nodes))
	})
}

func TestCreateInboundLocalGroupListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, admins.Len())
		require.Equal(t, 2, admins.ContainingNodeKinds(ad.User).Len())

		admins, err = adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerC, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, admins.Slice()[0].ID)

		admins, err = adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo)(tx, harness.LocalGroupSQL.ComputerB, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())
		require.Equal(t, harness.LocalGroupSQL.ComputerA.ID, admins.Slice()[0].ID)
	})
}

func TestCreateSQLAdminPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserC.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))

		path, err = adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserC)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserC.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))
	})
}

func TestCreateSQLAdminListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.CreateSQLAdminListDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())

		computers, err := adAnalysis.CreateSQLAdminListDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserC, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
	})
}

func TestCreateConstrainedDelegationPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserD.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))

		path, err = adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserD)
		test.RequireNilErr(t, err)
		nodes = path.AllNodes().IDs()
		require.Contains(t, nodes, harness.LocalGroupSQL.UserD.ID)
		require.Contains(t, nodes, harness.LocalGroupSQL.ComputerA.ID)
		require.Equal(t, 2, len(nodes))
	})
}

func TestCreateConstrainedDelegationListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.LocalGroupSQL.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionInbound)(tx, harness.LocalGroupSQL.ComputerA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, admins.Len())

		computers, err := adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionOutbound)(tx, harness.LocalGroupSQL.UserD, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, computers.Len())
	})
}

func TestFetchOutboundADEntityControlPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OutboundControl.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		path, err := adAnalysis.FetchOutboundADEntityControlPaths(context.Background(), db, harness.OutboundControl.Controller)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Equal(t, 7, len(nodes))
		require.Contains(t, nodes, harness.OutboundControl.Controller.ID)
		require.Contains(t, nodes, harness.OutboundControl.GroupA.ID)
		require.Contains(t, nodes, harness.OutboundControl.UserC.ID)
		require.Contains(t, nodes, harness.OutboundControl.GroupB.ID)
		require.Contains(t, nodes, harness.OutboundControl.GroupC.ID)
		require.Contains(t, nodes, harness.OutboundControl.ComputerA.ID)
		require.Contains(t, nodes, harness.OutboundControl.ComputerC.ID)
	})
}

func TestFetchOutboundADEntityControl(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OutboundControl.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		control, err := adAnalysis.FetchOutboundADEntityControl(context.Background(), db, harness.OutboundControl.Controller, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 4, control.Len())
		ids := control.IDs()

		require.Contains(t, ids, harness.OutboundControl.GroupB.ID)
		require.Contains(t, ids, harness.OutboundControl.UserC.ID)
		require.Contains(t, ids, harness.OutboundControl.ComputerA.ID)
		require.Contains(t, ids, harness.OutboundControl.ComputerC.ID)

		control, err = adAnalysis.FetchOutboundADEntityControl(context.Background(), db, harness.OutboundControl.ControllerB, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, control.Len())
	})
}

func TestFetchInboundADEntityControllerPaths(t *testing.T) {
	t.Run("User", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			path, err := adAnalysis.FetchInboundADEntityControllerPaths(context.Background(), db, harness.InboundControl.ControlledUser)
			test.RequireNilErr(t, err)

			nodes := path.AllNodes().IDs()
			require.Equal(t, 5, len(nodes))
			require.Contains(t, nodes, harness.InboundControl.ControlledUser.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupA.ID)
			require.Contains(t, nodes, harness.InboundControl.UserA.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupB.ID)
			require.Contains(t, nodes, harness.InboundControl.UserD.ID)
		})
	})

	t.Run("Group", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			path, err := adAnalysis.FetchInboundADEntityControllerPaths(context.Background(), db, harness.InboundControl.ControlledGroup)
			test.RequireNilErr(t, err)

			nodes := path.AllNodes().IDs()
			require.Equal(t, 7, len(nodes))
			require.Contains(t, nodes, harness.InboundControl.ControlledGroup.ID)
			require.Contains(t, nodes, harness.InboundControl.UserE.ID)
			require.Contains(t, nodes, harness.InboundControl.UserF.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupC.ID)
			require.Contains(t, nodes, harness.InboundControl.UserG.ID)
			require.Contains(t, nodes, harness.InboundControl.GroupD.ID)
			require.Contains(t, nodes, harness.InboundControl.UserH.ID)
		})
	})
}

func TestFetchInboundADEntityControllers(t *testing.T) {
	t.Run("User", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			control, err := adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledUser, 0, 0)
			test.RequireNilErr(t, err)
			require.Equal(t, 4, control.Len())

			ids := control.IDs()
			require.Contains(t, ids, harness.InboundControl.UserD.ID)
			require.Contains(t, ids, harness.InboundControl.GroupB.ID)
			require.Contains(t, ids, harness.InboundControl.UserA.ID)
			require.Contains(t, ids, harness.InboundControl.GroupA.ID)

			control, err = adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledUser, 0, 1)
			test.RequireNilErr(t, err)
			require.Equal(t, 1, control.Len())
		})
	})

	t.Run("Group", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.InboundControl.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			control, err := adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledGroup, 0, 0)
			test.RequireNilErr(t, err)
			require.Equal(t, 6, control.Len())

			ids := control.IDs()
			require.Contains(t, ids, harness.InboundControl.GroupC.ID)
			require.Contains(t, ids, harness.InboundControl.GroupD.ID)
			require.Contains(t, ids, harness.InboundControl.UserE.ID)
			require.Contains(t, ids, harness.InboundControl.UserF.ID)
			require.Contains(t, ids, harness.InboundControl.UserG.ID)
			require.Contains(t, ids, harness.InboundControl.UserH.ID)

			control, err = adAnalysis.FetchInboundADEntityControllers(context.Background(), db, harness.InboundControl.ControlledGroup, 0, 1)
			test.RequireNilErr(t, err)
			require.Equal(t, 1, control.Len())
		})
	})
}

func TestCreateOUContainedPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OUHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.CreateOUContainedPathDelegate(ad.User)(tx, harness.OUHarness.OUA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.OUHarness.OUA.ID)
		require.Contains(t, nodes, harness.OUHarness.UserA.ID)
		require.Contains(t, nodes, harness.OUHarness.OUC.ID)
		require.Contains(t, nodes, harness.OUHarness.UserB.ID)

		paths, err = adAnalysis.CreateOUContainedPathDelegate(ad.User)(tx, harness.OUHarness.OUB)
		test.RequireNilErr(t, err)
		nodes = paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.OUHarness.OUB.ID)
		require.Contains(t, nodes, harness.OUHarness.OUD.ID)
		require.Contains(t, nodes, harness.OUHarness.OUE.ID)
		require.Contains(t, nodes, harness.OUHarness.UserC.ID)
	})
}

func TestCreateOUContainedListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OUHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		contained, err := adAnalysis.CreateOUContainedListDelegate(ad.User)(tx, harness.OUHarness.OUA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, contained.Len())

		contained, err = adAnalysis.CreateOUContainedListDelegate(ad.User)(tx, harness.OUHarness.OUB, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 1, contained.Len())
	})
}

func TestFetchGroupMemberPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		path, err := adAnalysis.FetchGroupMemberPaths(tx, harness.MembershipHarness.GroupB)

		test.RequireNilErr(t, err)
		nodes := path.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.MembershipHarness.GroupB.ID)
		require.Contains(t, nodes, harness.MembershipHarness.UserA.ID)
		require.Contains(t, nodes, harness.MembershipHarness.ComputerA.ID)
	})
}

func TestFetchGroupMembers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		members, err := adAnalysis.FetchGroupMembers(context.Background(), db, harness.MembershipHarness.GroupC, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 5, members.Len())
		require.Equal(t, 2, members.ContainingNodeKinds(ad.Computer).Len())
		require.Equal(t, 2, members.ContainingNodeKinds(ad.Group).Len())
		require.Equal(t, 1, members.ContainingNodeKinds(ad.User).Len())
	})
}

func TestFetchEntityGroupMembershipPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchEntityGroupMembershipPaths(tx, harness.MembershipHarness.UserA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.MembershipHarness.UserA.ID)
		require.Contains(t, nodes, harness.MembershipHarness.GroupB.ID)
		require.Contains(t, nodes, harness.MembershipHarness.GroupA.ID)
	})
}

func TestFetchEntityGroupMembership(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.MembershipHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		membership, err := adAnalysis.FetchEntityGroupMembership(tx, harness.MembershipHarness.UserA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 3, membership.Len())
	})
}

func TestCreateForeignEntityMembershipListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		members, err := adAnalysis.CreateForeignEntityMembershipListDelegate(ad.Group)(tx, harness.ForeignHarness.LocalDomain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 1, members.Len())
		require.Equal(t, 1, members.ContainingNodeKinds(ad.Group).Len())

		members, err = adAnalysis.CreateForeignEntityMembershipListDelegate(ad.User)(tx, harness.ForeignHarness.LocalDomain, 0, 0)
		test.RequireNilErr(t, err)
		require.Equal(t, 2, members.Len())
		require.Equal(t, 2, members.ContainingNodeKinds(ad.User).Len())
	})
}

func TestFetchCollectedDomains(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		domains, err := adAnalysis.FetchCollectedDomains(tx)
		test.RequireNilErr(t, err)
		for _, domain := range domains {
			collected, err := domain.Properties.Get(common.Collected.String()).Bool()
			test.RequireNilErr(t, err)
			require.True(t, collected)
		}
		require.Equal(t, harness.NumCollectedActiveDirectoryDomains, domains.Len())
		require.NotContains(t, domains.IDs(), harness.TrustDCSync.DomainC.ID)
	})
}

func TestCreateDomainTrustPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionOutbound)(tx, harness.TrustDCSync.DomainA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.TrustDCSync.DomainA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainC.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainD.ID)

		paths, err = adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionInbound)(tx, harness.TrustDCSync.DomainA)

		test.RequireNilErr(t, err)
		nodes = paths.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.TrustDCSync.DomainA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.DomainD.ID)
	})
}

func TestCreateDomainTrustListDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		domains, err := adAnalysis.CreateDomainTrustListDelegate(graph.DirectionOutbound)(tx, harness.TrustDCSync.DomainA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 3, domains.Len())
		ids := domains.IDs()
		require.Contains(t, ids, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, ids, harness.TrustDCSync.DomainC.ID)
		require.Contains(t, ids, harness.TrustDCSync.DomainD.ID)

		domains, err = adAnalysis.CreateDomainTrustListDelegate(graph.DirectionInbound)(tx, harness.TrustDCSync.DomainA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, domains.Len())
		ids = domains.IDs()
		require.Contains(t, ids, harness.TrustDCSync.DomainB.ID)
		require.Contains(t, ids, harness.TrustDCSync.DomainD.ID)
	})
}

func TestFetchDCSyncers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		dcSyncers, err := adAnalysis.FetchDCSyncers(tx, harness.TrustDCSync.DomainA, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, dcSyncers.Len())

		nodes := dcSyncers.IDs()
		require.Contains(t, nodes, harness.TrustDCSync.UserB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.UserA.ID)
	})
}

func TestFetchDCSyncerPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchDCSyncerPaths(tx, harness.TrustDCSync.DomainA)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.TrustDCSync.DomainA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.UserA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.GroupA.ID)
		require.Contains(t, nodes, harness.TrustDCSync.GroupB.ID)
		require.Contains(t, nodes, harness.TrustDCSync.UserB.ID)
	})
}

func TestCreateForeignEntityMembershipPathDelegate(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.WriteTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.Group)(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 2, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGroup.ID)

		paths, err = adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.User)(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes = paths.AllNodes().IDs()
		require.Equal(t, 4, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserA.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserB.ID)
	})
}

func TestFetchForeignAdmins(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.FetchForeignAdmins(tx, harness.ForeignHarness.LocalDomain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, admins.Len())
		require.Equal(t, 2, admins.ContainingNodeKinds(ad.User).Len())
	})
}

func TestFetchForeignAdminPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchForeignAdminPaths(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 5, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.LocalComputer.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserA.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserB.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
	})
}

func TestFetchForeignGPOControllers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		admins, err := adAnalysis.FetchForeignGPOControllers(tx, harness.ForeignHarness.LocalDomain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, admins.Len())
		require.Equal(t, 1, admins.ContainingNodeKinds(ad.User).Len())
		require.Equal(t, 1, admins.ContainingNodeKinds(ad.Group).Len())
	})
}

func TestFetchForeignGPOControllerPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ForeignHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchForeignGPOControllerPaths(tx, harness.ForeignHarness.LocalDomain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.ForeignHarness.ForeignUserA.ID)
		require.Contains(t, nodes, harness.ForeignHarness.ForeignGroup.ID)
		require.Contains(t, nodes, harness.ForeignHarness.LocalGPO.ID)
	})
}

func TestFetchAllEnforcedGPOs(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		gpos, err := adAnalysis.FetchAllEnforcedGPOs(context.Background(), db, graph.NewNodeSet(harness.GPOEnforcement.OrganizationalUnitD))

		test.RequireNilErr(t, err)
		require.Equal(t, 2, gpos.Len())

		gpos, err = adAnalysis.FetchAllEnforcedGPOs(context.Background(), db, graph.NewNodeSet(harness.GPOEnforcement.OrganizationalUnitC))

		test.RequireNilErr(t, err)
		require.Equal(t, 1, gpos.Len())
	})
}

func TestFetchEntityLinkedGPOList(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		gpos, err := adAnalysis.FetchEntityLinkedGPOList(tx, harness.GPOEnforcement.Domain, 0, 0)

		test.RequireNilErr(t, err)
		require.Equal(t, 2, gpos.Len())
	})
}

func TestFetchEntityLinkedGPOPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GPOEnforcement.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := adAnalysis.FetchEntityLinkedGPOPaths(tx, harness.GPOEnforcement.Domain)

		test.RequireNilErr(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 3, len(nodes))
		require.Contains(t, nodes, harness.GPOEnforcement.Domain.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOUnenforced.ID)
		require.Contains(t, nodes, harness.GPOEnforcement.GPOEnforced.ID)
	})
}

func TestFetchLocalGroupCompleteness(t *testing.T) {
	var (
		testCtx = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		graphDB = testCtx.Graph.Database
	)

	fixture, err := arrows.LoadGraphFromFile(integration.Harnesses, "harnesses/completenessharness.json")
	require.NoError(t, err)

	err = arrows.WriteGraphToDatabase(graphDB, &fixture)
	require.NoError(t, err)

	err = graphDB.ReadTransaction(testCtx.Context(), func(tx graph.Transaction) error {
		// why does this function ask for a transaction type?
		completeness, err := adAnalysis.FetchLocalGroupCompleteness(tx, "DOMAIN123")
		require.NoError(t, err)
		assert.Equal(t, .5, completeness)
		return nil
	})
	assert.NoError(t, err)
}

func TestFetchUserSessionCompleteness(t *testing.T) {
	var (
		testCtx = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		graphDB = testCtx.Graph.Database
	)

	fixture, err := arrows.LoadGraphFromFile(integration.Harnesses, "harnesses/completenessharness.json")
	require.NoError(t, err)

	err = arrows.WriteGraphToDatabase(graphDB, &fixture)
	require.NoError(t, err)

	err = graphDB.ReadTransaction(testCtx.Context(), func(tx graph.Transaction) error {
		// why does this function ask for a transaction type?
		completeness, err := adAnalysis.FetchUserSessionCompleteness(tx, "DOMAIN123")
		require.NoError(t, err)
		assert.Equal(t, .5, completeness)
		return nil
	})
	assert.NoError(t, err)
}

func TestSyncLAPSPassword(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.SyncLAPSPasswordHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		if localGroupData, err := adAnalysis.FetchLocalGroupData(testContext.Context(), db); err != nil {
			t.Fatalf("error expanding groups in integration test; %v", err)
		} else if _, err := adAnalysis.PostSyncLAPSPassword(testContext.Context(), db, localGroupData); err != nil {
			t.Fatalf("error creating SyncLAPSPassword edges in integration test; %v", err)
		} else {
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.SyncLAPSPassword)
				})); err != nil {
					t.Fatalf("error fetching SyncLAPSPassword edges in integration test; %v", err)
				} else {
					require.Equal(t, 4, len(results))

					require.True(t, results.Contains(harness.SyncLAPSPasswordHarness.Group1))
					require.True(t, results.Contains(harness.SyncLAPSPasswordHarness.Group4))
					require.True(t, results.Contains(harness.SyncLAPSPasswordHarness.User3))
					require.True(t, results.Contains(harness.SyncLAPSPasswordHarness.User5))
				}
				return nil
			})
		}
	})
}

func TestDCSync(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.DCSyncHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		if localGroupData, err := adAnalysis.FetchLocalGroupData(testContext.Context(), db); err != nil {
			t.Fatalf("error expanding groups in integration test; %v", err)
		} else if _, err := adAnalysis.PostDCSync(testContext.Context(), db, localGroupData); err != nil {
			t.Fatalf("error creating DCSync edges in integration test; %v", err)
		} else {
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.DCSync)
				})); err != nil {
					t.Fatalf("error fetching DCSync edges in integration test; %v", err)
				} else {
					require.Equal(t, 3, len(results))

					require.True(t, results.Contains(harness.DCSyncHarness.User1))
					require.True(t, results.Contains(harness.DCSyncHarness.User2))
					require.True(t, results.Contains(harness.DCSyncHarness.Group3))

				}
				return nil
			})
		}
	})
}

func TestOwnsWriteOwnerPriorCollectorVersions(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OwnsWriteOwnerPriorCollectorVersions.Setup(testContext)
		// To verify in Neo4j: MATCH (n:Computer) MATCH (u:User) RETURN n, u
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		if localGroupData, err := adAnalysis.FetchLocalGroupData(testContext.Context(), db); err != nil {
			t.Fatalf("error expanding groups in integration test; %v", err)
		} else if _, err := adAnalysis.PostOwnsAndWriteOwner(testContext.Context(), db, localGroupData); err != nil {
			t.Fatalf("error creating Owns/WriteOwner edges in integration test; %v", err)
		} else {
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {

				// Owns: MATCH (a)-[r:Owns]->(b) RETURN a, r, b;
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.Owns),
						query.Kind(query.Start(), ad.Entity),
					)
				})); err != nil {
					t.Fatalf("error fetching Owns edges in integration test; %v", err)
				} else {
					require.Equal(t, 14, len(results))

					for _, rel := range results {
						if startNode, err := ops.FetchNode(tx, rel.StartID); err != nil {
							t.Fatalf("error fetching start node in integration test; %v", err)
						} else if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							t.Fatalf("error fetching target node in integration test; %v", err)
						} else {
							// Extract 'name' properties from startNode and targetNode
							startNodeName, okStart := startNode.Properties.Map["name"].(string)
							if !okStart {
								startNodeName = "<unknown>"
							}
							targetNodeName, okTarget := targetNode.Properties.Map["name"].(string)
							if !okTarget {
								targetNodeName = "<unknown>"
							}

							switch targetNode.ID {
							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User1_NoOwnerRights_OwnerIsLowPriv.ID:
								// Domain1_User101_Owner -[Owns]-> Domain1_User1_NoOwnerRights_OwnerIsLowPriv
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User101_Owner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_Computer2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_Computer2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_MSA2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_MSA2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_GMSA2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_GMSA2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_User2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_Computer3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_Computer3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_MSA3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_MSA3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_GMSA3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_GMSA3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_User3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer1_NoOwnerRights.ID:
								// Domain2_User1_Owner -[Owns]-> Domain2_Computer1_NoOwnerRights
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User1_Owner.ID, startNode.ID)

								//
								// Below here are the expected false positives present after post-processing data from the prior collector versions
								//

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User7_OnlyNonabusableOwnerRightsAndNoneInherited.ID:
								// Domain1_User101_Owner -[Owns]-> Domain1_User7_OnlyNonabusableOwnerRightsAndNoneInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User101_Owner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User8_OnlyNonabusableOwnerRightsInherited.ID:
								// Domain1_User101_Owner -[Owns]-> Domain1_User8_OnlyNonabusableOwnerRightsInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User101_Owner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer5_OnlyNonabusableOwnerRightsAndNoneInherited.ID:
								// Domain2_User1_Owner -[Owns]-> Domain2_Computer5_OnlyNonabusableOwnerRightsAndNoneInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User1_Owner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer6_OnlyNonabusableOwnerRightsInherited.ID:
								// Domain2_User1_Owner -[Owns]-> Domain2_Computer6_OnlyNonabusableOwnerRightsInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User1_Owner.ID, startNode.ID)

							default:
								t.Fatalf("unexpected edge in integration test: %s -[Owns]-> %s", startNodeName, targetNodeName)
							}
						}
					}
				}

				// WriteOwner: MATCH (a)-[r:WriteOwner]->(b) RETURN a, r, b;
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.WriteOwner),
						query.Kind(query.Start(), ad.Entity),
					)
				})); err != nil {
					t.Fatalf("error fetching WriteOwner edges in integration test; %v", err)
				} else {
					require.Equal(t, 12, len(results))

					for _, rel := range results {
						if startNode, err := ops.FetchNode(tx, rel.StartID); err != nil {
							t.Fatalf("error fetching start node in integration test; %v", err)
						} else if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							t.Fatalf("error fetching target node in integration test; %v", err)
						} else {
							// Extract 'name' properties from startNode and targetNode
							startNodeName, okStart := startNode.Properties.Map["name"].(string)
							if !okStart {
								startNodeName = "<unknown>"
							}
							targetNodeName, okTarget := targetNode.Properties.Map["name"].(string)
							if !okTarget {
								targetNodeName = "<unknown>"
							}

							switch targetNode.ID {
							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User1_NoOwnerRights_OwnerIsLowPriv.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User1_NoOwnerRights_OwnerIsLowPriv
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User4_AbusableOwnerRightsNoneInherited.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User4_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User7_OnlyNonabusableOwnerRightsAndNoneInherited.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User7_OnlyNonabusableOwnerRightsAndNoneInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer1_NoOwnerRights.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer1_NoOwnerRights
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User2_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer2_AbusableOwnerRightsNoneInherited.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer2_AbusableOwnerRightsNoneInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User2_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer5_OnlyNonabusableOwnerRightsAndNoneInherited.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer5_OnlyNonabusableOwnerRightsAndNoneInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User2_WriteOwner.ID, startNode.ID)

								//
								// Below here are the expected false positives present after post-processing data from the prior collector versions
								//

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User6_AbusableOwnerRightsOnlyNonabusableInherited.ID:
								// Domain1_User101_Owner -[WriteOwner]-> Domain1_User6_AbusableOwnerRightsOnlyNonabusableInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User8_OnlyNonabusableOwnerRightsInherited.ID:
								// Domain1_User101_Owner -[WriteOwner]-> Domain1_User8_OnlyNonabusableOwnerRightsInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer4_AbusableOwnerRightsOnlyNonabusableInherited.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer4_AbusableOwnerRightsOnlyNonabusableInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User2_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_Computer6_OnlyNonabusableOwnerRightsInherited.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer6_OnlyNonabusableOwnerRightsInherited
								require.Equal(t, harness.OwnsWriteOwnerPriorCollectorVersions.Domain2_User2_WriteOwner.ID, startNode.ID)

							default:
								t.Fatalf("unexpected edge in integration test: %s -[Owns]-> %s", startNodeName, targetNodeName)
							}
						}
					}
				}

				return nil
			})
		}
	})
}

func TestOwnsWriteOwner(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.OwnsWriteOwner.Setup(testContext)
		// To verify in Neo4j: MATCH (n:Computer) MATCH (u:User) RETURN n, u
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		if localGroupData, err := adAnalysis.FetchLocalGroupData(testContext.Context(), db); err != nil {
			t.Fatalf("error expanding groups in integration test; %v", err)
		} else if _, err := adAnalysis.PostOwnsAndWriteOwner(testContext.Context(), db, localGroupData); err != nil {
			t.Fatalf("error creating Owns/WriteOwner edges in integration test; %v", err)
		} else {
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {

				// Owns: MATCH (a)-[r:Owns]->(b) RETURN a, r, b;
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.Owns),
						query.Kind(query.Start(), ad.Entity),
					)
				})); err != nil {
					t.Fatalf("error fetching Owns edges in integration test; %v", err)
				} else {
					require.Equal(t, 10, len(results))

					for _, rel := range results {
						if startNode, err := ops.FetchNode(tx, rel.StartID); err != nil {
							t.Fatalf("error fetching start node in integration test; %v", err)
						} else if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							t.Fatalf("error fetching target node in integration test; %v", err)
						} else {
							// Extract 'name' properties from startNode and targetNode
							startNodeName, okStart := startNode.Properties.Map["name"].(string)
							if !okStart {
								startNodeName = "<unknown>"
							}
							targetNodeName, okTarget := targetNode.Properties.Map["name"].(string)
							if !okTarget {
								targetNodeName = "<unknown>"
							}

							switch targetNode.ID {
							case harness.OwnsWriteOwner.Domain1_User1_NoOwnerRights_OwnerIsLowPriv.ID:
								// Domain1_User101_Owner -[Owns]-> Domain1_User1_NoOwnerRights_OwnerIsLowPriv
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User101_Owner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_Computer2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_Computer2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_MSA2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_MSA2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_GMSA2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_GMSA2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_User2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User102_DomainAdmin -[Owns]-> Domain1_User2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User102_DomainAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_Computer3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_Computer3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_MSA3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_MSA3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_GMSA3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_GMSA3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_User3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User103_EnterpriseAdmin -[Owns]-> Domain1_User3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User103_EnterpriseAdmin.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain2_Computer1_NoOwnerRights.ID:
								// Domain2_User1_Owner -[Owns]-> Domain2_Computer1_NoOwnerRights
								require.Equal(t, harness.OwnsWriteOwner.Domain2_User1_Owner.ID, startNode.ID)

							default:
								t.Fatalf("unexpected edge in integration test: %s -[Owns]-> %s", startNodeName, targetNodeName)
							}
						}
					}
				}

				// WriteOwner: MATCH (a)-[r:WriteOwner]->(b) RETURN a, r, b;
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.WriteOwner),
						query.Kind(query.Start(), ad.Entity),
					)
				})); err != nil {
					t.Fatalf("error fetching WriteOwner edges in integration test; %v", err)
				} else {
					require.Equal(t, 8, len(results))

					for _, rel := range results {
						if startNode, err := ops.FetchNode(tx, rel.StartID); err != nil {
							t.Fatalf("error fetching start node in integration test; %v", err)
						} else if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							t.Fatalf("error fetching target node in integration test; %v", err)
						} else {
							// Extract 'name' properties from startNode and targetNode
							startNodeName, okStart := startNode.Properties.Map["name"].(string)
							if !okStart {
								startNodeName = "<unknown>"
							}
							targetNodeName, okTarget := targetNode.Properties.Map["name"].(string)
							if !okTarget {
								targetNodeName = "<unknown>"
							}

							switch targetNode.ID {
							case harness.OwnsWriteOwner.Domain1_User1_NoOwnerRights_OwnerIsLowPriv.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User1_NoOwnerRights_OwnerIsLowPriv
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_User2_NoOwnerRights_OwnerIsDA.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User2_NoOwnerRights_OwnerIsDA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_User3_NoOwnerRights_OwnerIsEA.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User3_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_User4_AbusableOwnerRightsNoneInherited.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User4_NoOwnerRights_OwnerIsEA
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain1_User7_OnlyNonabusableOwnerRightsAndNoneInherited.ID:
								// Domain1_User104_WriteOwner -[WriteOwner]-> Domain1_User7_OnlyNonabusableOwnerRightsAndNoneInherited
								require.Equal(t, harness.OwnsWriteOwner.Domain1_User104_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain2_Computer1_NoOwnerRights.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer1_NoOwnerRights
								require.Equal(t, harness.OwnsWriteOwner.Domain2_User2_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain2_Computer2_AbusableOwnerRightsNoneInherited.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer2_AbusableOwnerRightsNoneInherited
								require.Equal(t, harness.OwnsWriteOwner.Domain2_User2_WriteOwner.ID, startNode.ID)

							case harness.OwnsWriteOwner.Domain2_Computer5_OnlyNonabusableOwnerRightsAndNoneInherited.ID:
								// Domain2_User2_WriteOwner -[WriteOwner]-> Domain2_Computer5_OnlyNonabusableOwnerRightsAndNoneInherited
								require.Equal(t, harness.OwnsWriteOwner.Domain2_User2_WriteOwner.ID, startNode.ID)

							default:
								t.Fatalf("unexpected edge in integration test: %s -[Owns]-> %s", startNodeName, targetNodeName)
							}
						}
					}
				}

				return nil
			})
		}
	})
}

func TestHasTrustKeys(t *testing.T) {
	var (
		testCtx = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		graphDB = testCtx.Graph.Database
	)

	fixture, err := arrows.LoadGraphFromFile(integration.Harnesses, "harnesses/HasTrustKeysHarness.json")
	require.NoError(t, err)

	// Split edges into test edges and the other edges
	testEdges := []arrows.Edge{}
	otherEdges := []arrows.Edge{}
	for _, edge := range fixture.Relationships {
		if edge.Type == ad.HasTrustKeys.String() {
			testEdges = append(testEdges, edge)
		} else {
			otherEdges = append(otherEdges, edge)
		}
	}
	fixture.Relationships = otherEdges

	err = arrows.WriteGraphToDatabase(graphDB, &fixture)
	require.NoError(t, err)

	err = graphDB.ReadTransaction(testCtx.Context(), func(tx graph.Transaction) error {
		if _, err := adAnalysis.PostHasTrustKeys(testCtx.Context(), graphDB); err != nil {
			t.Fatalf("error creating HasTrustKeys edges in integration test; %v", err)
		} else {
			if err = graphDB.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationshipIDs(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.HasTrustKeys)
				})); err != nil {
					t.Fatalf("error fetching HasTrustKeys edges in integration test; %v", err)
				} else {
					require.Equal(t, len(testEdges), len(results))
				}

				for _, testEdge := range testEdges {
					if fromNode, found := findNodeByID(fixture.Nodes, testEdge.FromID); !found {
						t.Fatalf("error finding source node with ID %s; %v", testEdge.FromID, err)
					} else if toNode, found := findNodeByID(fixture.Nodes, testEdge.ToID); !found {
						t.Fatalf("error finding destination node with ID %s; %v", testEdge.ToID, err)
					} else if fromGraphNodeId, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
						return query.Equals(query.NodeProperty(common.Name.String()), fromNode.Caption)
					})); err != nil || len(fromGraphNodeId) != 1 {
						t.Fatalf("error fetching node with name %s in integration test; %v", fromNode.Caption, err)
					} else if toGraphNodeId, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
						return query.Equals(query.NodeProperty(common.Name.String()), toNode.Caption)
					})); err != nil || len(toGraphNodeId) != 1 {
						t.Fatalf("error fetching node with name %s in integration test; %v", toNode.Caption, err)
					} else if edge, err := analysis.FetchEdgeByStartAndEnd(testCtx.Context(), graphDB, fromGraphNodeId[0], toGraphNodeId[0], ad.HasTrustKeys); err != nil {
						t.Fatalf("error fetching HasTrustKeys edge from node %s (ID: %d) to node %s (ID: %d) in integration test; %v", fromNode.Caption, fromGraphNodeId[0], toNode.Caption, toGraphNodeId[0], err)
					} else {
						require.NotNil(t, edge)
					}
				}

				return nil
			}); err != nil {
				t.Fatalf("error in HasTrustKeys integration test; %v", err)
			}
		}
		assert.NoError(t, err)
		return nil
	})
}

func findNodeByID(nodes []arrows.Node, id string) (*arrows.Node, bool) {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i], true
		}
	}
	return nil, false
}

func TestProtectAdminGroups(t *testing.T) {
	var (
		testCtx = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		graphDB = testCtx.Graph.Database
	)

	fixture, err := arrows.LoadGraphFromFile(integration.Harnesses, "harnesses/ProtectAdminGroupsHarness.json")
	require.NoError(t, err)

	// Split edges into test edges and the other edges
	testEdges := []arrows.Edge{}
	otherEdges := []arrows.Edge{}
	for _, edge := range fixture.Relationships {
		if edge.Type == ad.ProtectAdminGroups.String() {
			testEdges = append(testEdges, edge)
		} else {
			otherEdges = append(otherEdges, edge)
		}
	}
	fixture.Relationships = otherEdges

	err = arrows.WriteGraphToDatabase(graphDB, &fixture)
	require.NoError(t, err)

	err = graphDB.ReadTransaction(testCtx.Context(), func(tx graph.Transaction) error {
		if _, err := adAnalysis.PostProtectAdminGroups(testCtx.Context(), graphDB); err != nil {
			t.Fatalf("error creating ProtectAdminGroups edges in integration test; %v", err)
		} else {
			if err = graphDB.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationshipIDs(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.ProtectAdminGroups)
				})); err != nil {
					t.Fatalf("error fetching ProtectAdminGroups edges in integration test; %v", err)
				} else {
					require.Equal(t, len(testEdges), len(results))
				}

				for _, testEdge := range testEdges {
					if fromNode, found := findNodeByID(fixture.Nodes, testEdge.FromID); !found {
						t.Fatalf("error finding source node with ID %s; %v", testEdge.FromID, err)
					} else if toNode, found := findNodeByID(fixture.Nodes, testEdge.ToID); !found {
						t.Fatalf("error finding destination node with ID %s; %v", testEdge.ToID, err)
					} else if fromGraphNodeId, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
						return query.Equals(query.NodeProperty(common.Name.String()), fromNode.Caption)
					})); err != nil || len(fromGraphNodeId) != 1 {
						t.Fatalf("error fetching node with name %s in integration test; %v", fromNode.Caption, err)
					} else if toGraphNodeId, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
						return query.Equals(query.NodeProperty(common.Name.String()), toNode.Caption)
					})); err != nil || len(toGraphNodeId) != 1 {
						t.Fatalf("error fetching node with name %s in integration test; %v", toNode.Caption, err)
					} else if edge, err := analysis.FetchEdgeByStartAndEnd(testCtx.Context(), graphDB, fromGraphNodeId[0], toGraphNodeId[0], ad.ProtectAdminGroups); err != nil {
						t.Fatalf("error fetching ProtectAdminGroups edge from node %s (ID: %d) to node %s (ID: %d) in integration test; %v", fromNode.Caption, fromGraphNodeId[0], toNode.Caption, toGraphNodeId[0], err)
					} else {
						require.NotNil(t, edge)
					}
				}

				return nil
			}); err != nil {
				t.Fatalf("error in ProtectAdminGroups integration test; %v", err)
			}
		}
		assert.NoError(t, err)
		return nil
	})
}
