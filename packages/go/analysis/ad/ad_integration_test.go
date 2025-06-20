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
// +build serial_integration

package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/analysis/ad/internal/nodeprops"
	"github.com/specterops/bloodhound/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"

	adAnalysis "github.com/specterops/bloodhound/analysis/ad"

	"github.com/stretchr/testify/require"
)

type linkWellKnownGroupsTestCase struct {
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

func TestLinkWellKnownGroups(t *testing.T) {
	wellKnownGroups := []struct {
		sidSuffix      wellknown.SIDSuffix
		nodeNamePrefix wellknown.NodeNamePrefix
	}{
		{
			sidSuffix:      wellknown.DomainUsersSIDSuffix,
			nodeNamePrefix: wellknown.DomainUsersNodeNamePrefix,
		},
		{
			sidSuffix:      wellknown.AuthenticatedUsersSIDSuffix,
			nodeNamePrefix: wellknown.AuthenticatedUsersNodeNamePrefix,
		},
		{
			sidSuffix:      wellknown.EveryoneSIDSuffix,
			nodeNamePrefix: wellknown.EveryoneNodeNamePrefix,
		},
		{
			sidSuffix:      wellknown.DomainComputersSIDSuffix,
			nodeNamePrefix: wellknown.DomainComputerNodeNamePrefix,
		},
	}
	createdWellKnownGroups := make(map[wellknown.NodeNamePrefix]*graph.Node)

	testCases := []linkWellKnownGroupsTestCase{
		{
			name: "Verifies that linking all well-known groups succeeds when they already exist.",
			setupFunc: func(
				t *testing.T,
				ctx context.Context,
				graphDB graph.Database,
			) *graph.Node {
				// NOTE: Testing the scenario requires created the wellknown groups ahead of time asserting their
				// execution when asserting the scenario prior to asserting the expected outcome for
				// LinkWellKnownGroups
				createdCollectedDomainNode := createNode(
					t,
					ctx,
					graphDB,
					generateCollectedDomain(),
				)
				for _, wellKnownGroup := range wellKnownGroups {
					createdWellKnownGroup := createNode(
						t,
						ctx,
						graphDB,
						generateWellKnownGroup(
							t,
							createdCollectedDomainNode,
							wellKnownGroup.sidSuffix,
							wellKnownGroup.nodeNamePrefix,
						),
					)
					createdWellKnownGroups[wellKnownGroup.nodeNamePrefix] = createdWellKnownGroup
					assertNodeExists(
						t,
						ctx,
						graphDB,
						createdWellKnownGroup,
						query.StringEndsWith(
							query.NodeProperty(common.ObjectID.String()),
							wellKnownGroup.sidSuffix.String(),
						),
						query.StringStartsWith(
							query.NodeProperty(common.Name.String()),
							wellKnownGroup.nodeNamePrefix.String(),
						),
					)
				}
				return createdCollectedDomainNode
			},
			assertionFunc: func(
				t *testing.T,
				ctx context.Context,
				graphDB graph.Database,
				expectedNode *graph.Node,
			) {
				// assert that the relationships exist
				expectations := map[string][]*graph.Node{
					"domain user node is linked to authenticated users node": {
						createdWellKnownGroups[wellknown.DomainUsersNodeNamePrefix],
						createdWellKnownGroups[wellknown.AuthenticatedUsersNodeNamePrefix],
					},
					"domain computer node is linked to authenticated users node": {
						createdWellKnownGroups[wellknown.DomainComputerNodeNamePrefix],
						createdWellKnownGroups[wellknown.AuthenticatedUsersNodeNamePrefix],
					},
					"authenticated users node is linked to everyone node": {
						createdWellKnownGroups[wellknown.AuthenticatedUsersNodeNamePrefix],
						createdWellKnownGroups[wellknown.EveryoneNodeNamePrefix],
					},
				}
				for name, nodes := range expectations {
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

			// Run LinkWellKnownGroups
			err := adAnalysis.LinkWellKnownGroups(ctx, graphDB)
			require.NoError(t, err)

			tc.assertionFunc(t, ctx, graphDB, createdNode)
			require.NoError(t, err)
		})
	}
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

func generateCollectedDomain() *graph.Node {
	return &graph.Node{
		Kinds: graph.Kinds{
			ad.Domain,
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			common.Collected: true,
			common.Name:      "domain-test.com",
			ad.DomainSID:     "S-1-5-21-1004336348-1177238915-682003330-512",
			ad.DomainFQDN:    "domain-test.com",
		}),
	}
}

func generateWellKnownGroup(
	t *testing.T,
	domainNode *graph.Node,
	sidSuffix wellknown.SIDSuffix,
	nodeNamePrefix wellknown.NodeNamePrefix,
) *graph.Node {
	require.NotNil(t, domainNode)
	require.NotNil(t, sidSuffix)
	require.NotNil(t, nodeNamePrefix)

	domainSID, domainName, err := nodeprops.ReadDomainIDandNameAsString(domainNode)
	require.NoError(t, err)

	var wellKnownSID string
	switch sidSuffix.String() {
	case wellknown.DomainUsersSIDSuffix.String(), wellknown.DomainComputersSIDSuffix.String():
		wellKnownSID = sidSuffix.PrependPrefix(domainSID)
	default:
		wellKnownSID = sidSuffix.PrependPrefix(domainName)
	}

	return &graph.Node{
		Kinds: graph.Kinds{
			ad.Entity,
			ad.Group,
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
