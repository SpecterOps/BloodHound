// Copyright 2024 Specter Ops, Inc.
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

package ad_test

import (
	"context"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/analysis"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/lab/arrows"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostNTLMRelayADCS(t *testing.T) {
	// TODO: Add some negative tests here
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToADCS.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToADCS")

		localGroupData, cache, err := FetchADCSPrereqs(db)
		require.NoError(t, err)

		ntlmCache, err := adAnalysis.NewNTLMCache(context.Background(), db, localGroupData)
		require.NoError(t, err)

		err = adAnalysis.PostCoerceAndRelayNTLMToADCS(context.Background(), operation, cache, ntlmCache)
		require.NoError(t, err)

		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToADCS)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {

				require.Len(t, results, 1)
				rel := results[0]

				start, end, err := ops.FetchRelationshipNodes(tx, rel)
				require.NoError(t, err)

				require.Equal(t, start.ID, harness.NTLMCoerceAndRelayNTLMToADCS.AuthenticatedUsersGroup.ID)
				require.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToADCS.Computer.ID)
			}
			return nil
		})
	})
}

func TestNTLMRelayToADCSComposition(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToADCS.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Composition Test - CoerceAndRelayNTLMToADCS")

		localGroupData, cache, err := FetchADCSPrereqs(db)
		require.NoError(t, err)

		ntlmCache, err := adAnalysis.NewNTLMCache(context.Background(), db, localGroupData)
		require.NoError(t, err)

		err = adAnalysis.PostCoerceAndRelayNTLMToADCS(context.Background(), operation, cache, ntlmCache)
		require.NoError(t, err)

		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToADCS),
						query.Equals(query.StartProperty(common.Name.String()), "Authenticated Users Group"),
					)
				}).First(); err != nil {

				t.Fatalf("error fetching NTLM to ADCS edge in integration test: %v", err)
			} else {
				composition, err := adAnalysis.GetCoerceAndRelayNTLMtoADCSEdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 7, len(nodes))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.Computer))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.CertTemplate1))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.EnterpriseCA1))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.RootCA))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.Domain))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.NTAuthStore))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.AuthenticatedUsersGroup))
			}
			return nil
		})
	})

}

func TestCoerceAndRelayNTLMToADCSTrust(t *testing.T) {
	var (
		ctx        = context.Background()
		graphDB    = newNTLMIntegrationGraph(t, ctx)
		testEdges  []arrows.Edge
		otherEdges []arrows.Edge
	)

	fixture, err := arrows.LoadGraphFromFile(integration.Harnesses, "harnesses/CoerceAndRelayNTLMToADCSTrust.json")
	require.NoError(t, err)

	// Split edges into test edges and the other edges
	for _, edge := range fixture.Relationships {
		if edge.Type == ad.CoerceAndRelayNTLMToADCS.String() {
			testEdges = append(testEdges, edge)
		} else {
			otherEdges = append(otherEdges, edge)
		}
	}
	require.NotEmpty(t, testEdges)
	fixture.Relationships = otherEdges

	err = arrows.WriteGraphToDatabase(graphDB, &fixture)
	require.NoError(t, err)

	localGroupData, err := adAnalysis.FetchLocalGroupData(ctx, graphDB)
	require.NoError(t, err)

	ntlmCache, err := adAnalysis.NewNTLMCache(ctx, graphDB, localGroupData)
	require.NoError(t, err)

	enterpriseCertAuthorities, err := adAnalysis.FetchNodesByKind(ctx, graphDB, ad.EnterpriseCA)
	require.NoError(t, err)

	certTemplates, err := adAnalysis.FetchNodesByKind(ctx, graphDB, ad.CertTemplate)
	require.NoError(t, err)

	adcsCache := adAnalysis.NewADCSCache()
	err = adcsCache.BuildCache(ctx, graphDB, enterpriseCertAuthorities, certTemplates)
	require.NoError(t, err)

	operation := post.NewPostRelationshipOperation(ctx, graphDB, "CoerceAndRelayNTLMToADCS Trust Post Processing")
	err = adAnalysis.PostCoerceAndRelayNTLMToADCS(adcsCache, operation, ntlmCache)
	require.NoError(t, err)
	require.NoError(t, operation.Done())

	err = graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if results, err := ops.FetchRelationshipIDs(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToADCS)
		})); err != nil {
			t.Fatalf("error fetching CoerceAndRelayNTLMToADCS edges in integration test; %v", err)
		} else {
			require.Equal(t, len(testEdges), len(results))
		}

		for _, testEdge := range testEdges {
			if fromNode, found := findNTLMFixtureNodeByID(fixture.Nodes, testEdge.FromID); !found {
				t.Fatalf("error finding source node with ID %s; %v", testEdge.FromID, err)
			} else if toNode, found := findNTLMFixtureNodeByID(fixture.Nodes, testEdge.ToID); !found {
				t.Fatalf("error finding destination node with ID %s; %v", testEdge.ToID, err)
			} else if fromGraphNodeId, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
				return query.Equals(query.NodeProperty(common.Name.String()), fromNode.Caption)
			})); err != nil || len(fromGraphNodeId) != 1 {
				t.Fatalf("error fetching node with name %s in integration test; %v", fromNode.Caption, err)
			} else if toGraphNodeId, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
				return query.Equals(query.NodeProperty(common.Name.String()), toNode.Caption)
			})); err != nil || len(toGraphNodeId) != 1 {
				t.Fatalf("error fetching node with name %s in integration test; %v", toNode.Caption, err)
			} else if edge, err := analysis.FetchEdgeByStartAndEnd(ctx, graphDB, fromGraphNodeId[0], toGraphNodeId[0], ad.CoerceAndRelayNTLMToADCS); err != nil {
				t.Fatalf("error fetching CoerceAndRelayNTLMToADCS edge from node %s (ID: %d) to node %s (ID: %d) in integration test; %v", fromNode.Caption, fromGraphNodeId[0], toNode.Caption, toGraphNodeId[0], err)
			} else {
				require.NotNil(t, edge)

				composition, err := adAnalysis.GetCoerceAndRelayNTLMtoADCSEdgeComposition(ctx, graphDB, edge)
				require.NoError(t, err)
				require.Greater(t, composition.AllNodes().Len(), 0)

				compositionEdgeCount := 0
				for _, path := range composition.Paths() {
					compositionEdgeCount += len(path.Edges)
				}
				require.Greater(t, compositionEdgeCount, 0)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("error in CoerceAndRelayNTLMToADCS integration test; %v", err)
	}
}

func newNTLMIntegrationGraph(t *testing.T, ctx context.Context) graph.Database {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	connConf := pgtestdb.Custom(t, integration.GetPostgresConfig(cfg), pgtestdb.NoopMigrator{})
	cfg.Database.Connection = connConf.URL()

	pool, err := pg.NewPool(cfg.Database)
	require.NoError(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		ConnectionString: cfg.Database.PostgreSQLConnectionString(),
		Pool:             pool,
	})
	require.NoError(t, err)
	require.NoError(t, migrations.NewGraphMigrator(graphDB).Migrate(ctx))
	require.NoError(t, graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema()))

	t.Cleanup(func() {
		if err := graphDB.Close(ctx); err != nil {
			t.Logf("failed to close graph database: %v", err)
		}
	})

	return graphDB
}

func findNTLMFixtureNodeByID(nodes []arrows.Node, id string) (*arrows.Node, bool) {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i], true
		}
	}
	return nil, false
}

func TestPostNTLMRelaySMB(t *testing.T) {
	t.Run("NTLMCoerceAndRelayNTLMToSMB Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToSMB.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToSMB")

			grouplocalGroupData, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)
			ntlmCache, err := adAnalysis.NewNTLMCache(context.Background(), db, grouplocalGroupData)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, _ := innerComputer.Properties.Get(ad.DomainSID.String()).String()

					if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
						t.Fatalf("authenticated user not found for %s", domainSid)
					} else if err = adAnalysis.PostCoerceAndRelayNTLMToSMB(tx, outC, ntlmCache, innerComputer, authenticatedUserID); err != nil {
						t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)

			err = operation.Done()
			require.NoError(t, err)

			// Test start node
			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB)
				})); err != nil {
					t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
				} else {
					require.Len(t, results, 2)

					for _, result := range results {
						start, end, err := ops.FetchRelationshipNodes(tx, result)
						require.NoError(t, err)

						switch start.ID {
						case harness.NTLMCoerceAndRelayNTLMToSMB.Group2.ID:
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToSMB.Computer9.ID)
							coercionTargets, err := adAnalysis.GetCoercionTargetsForCoerceAndRelayNTLMtoSMB(context.Background(), db, result)
							assert.NoError(t, err)
							assert.Contains(t, coercionTargets.IDs(), harness.NTLMCoerceAndRelayNTLMToSMB.Computer8.ID)
							composition, err := adAnalysis.GetEdgeCompositionPath(context.Background(), db, result)
							assert.NoError(t, err)
							nodes := composition.AllNodes().IDs()
							assert.Contains(t, nodes, harness.NTLMCoerceAndRelayNTLMToSMB.Group8.ID)
							assert.Contains(t, nodes, harness.NTLMCoerceAndRelayNTLMToSMB.Group7.ID)
							assert.Contains(t, nodes, harness.NTLMCoerceAndRelayNTLMToSMB.Computer8.ID)
						case harness.NTLMCoerceAndRelayNTLMToSMB.Group1.ID:
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToSMB.Computer2.ID)
							coercionTargets, err := adAnalysis.GetCoercionTargetsForCoerceAndRelayNTLMtoSMB(context.Background(), db, result)
							assert.NoError(t, err)
							assert.Contains(t, coercionTargets.IDs(), harness.NTLMCoerceAndRelayNTLMToSMB.Computer1.ID)
							composition, err := adAnalysis.GetEdgeCompositionPath(context.Background(), db, result)
							assert.NoError(t, err)
							nodes := composition.AllNodes().IDs()
							assert.Contains(t, nodes, harness.NTLMCoerceAndRelayNTLMToSMB.Computer1.ID)
							assert.Contains(t, nodes, harness.NTLMCoerceAndRelayNTLMToSMB.Computer2.ID)
						default:
							require.FailNow(t, "unrecognized start node id")
						}
					}
				}
				return nil
			})
		})
	})

	t.Run("NTLMCoerceAndRelayNTLMToSMB Self Relay Does Not Create Edge", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToSMBSelfRelay.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM - CoerceAndRelayNTLMToSMB - Relay To Self")

			grouplocalGroupData, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)
			ntlmCache, err := adAnalysis.NewNTLMCache(context.Background(), db, grouplocalGroupData)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer

					if !ntlmCache.UnprotectedComputersCache.Contains(innerComputer.ID.Uint64()) {
						continue
					}

					domainSid, _ := innerComputer.Properties.Get(ad.DomainSID.String()).String()

					if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
						t.Fatalf("authenticated user not found for %s", domainSid)
					} else if err = adAnalysis.PostCoerceAndRelayNTLMToSMB(tx, outC, ntlmCache, innerComputer, authenticatedUserID); err != nil {
						t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)

			err = operation.Done()
			require.NoError(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB)
				})); err != nil {
					t.Fatalf("error fetching NTLM to LDAPS edges in integration test; %v", err)
				} else {
					require.Len(t, results, 0)
				}
				return nil
			})

		})
	})
}

func TestNTLMRelayToSMBComposition(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToSMB.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Composition Test - CoerceAndRelayNTLMToSMB")

		grouplocalGroupData, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
		require.NoError(t, err)
		ntlmCache, err := adAnalysis.NewNTLMCache(context.Background(), db, grouplocalGroupData)
		require.NoError(t, err)

		err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			for _, computer := range computers {
				innerComputer := computer
				domainSid, _ := innerComputer.Properties.Get(ad.DomainSID.String()).String()

				if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
					t.Fatalf("authenticated user not found for %s", domainSid)
				} else if err = adAnalysis.PostCoerceAndRelayNTLMToSMB(tx, outC, ntlmCache, innerComputer, authenticatedUserID); err != nil {
					t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
				}
			}
			return nil
		})
		require.NoError(t, err)

		err = operation.Done()
		require.NoError(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB),
						query.Equals(query.StartProperty(common.Name.String()), "Group1"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching NTLM to SMB edge in integration test: %v", err)
			} else {
				relayTargets, err := adAnalysis.GetCoercionTargetsForCoerceAndRelayNTLMtoSMB(context.Background(), db, edge)
				require.NoError(t, err)

				require.Len(t, relayTargets, 1)
				require.True(t, relayTargets.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Computer1))
			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB),
						query.Equals(query.StartProperty(common.Name.String()), "Group2"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching NTLM to SMB edge in integration test: %v", err)
			} else {
				relayTargets, err := adAnalysis.GetCoercionTargetsForCoerceAndRelayNTLMtoSMB(context.Background(), db, edge)
				require.NoError(t, err)

				require.Len(t, relayTargets, 1)
				require.True(t, relayTargets.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Computer8))
			}
			return nil
		})
	})
}

func TestPostCoerceAndRelayNTLMToLDAP(t *testing.T) {
	t.Run("NTLMCoerceAndRelayNTLMToLDAP Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToLDAP.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAP")

			grouplocalGroupData, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := adAnalysis.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := adAnalysis.FetchProtectedUsersMappedToDomains(testContext.Context(), db, grouplocalGroupData)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, err := innerComputer.Properties.Get(ad.DomainSID.String()).String()
					require.NoError(t, err)

					if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
						t.Fatalf("authenticated user not found for %s", domainSid)
					} else if protectedUsersForDomain, ok := protectedUsersCache[domainSid]; !ok {
						continue
					} else if ldapSigningForDomain, ok := ldapSigningCache[domainSid]; !ok {
						continue
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsVulnerableFunctionalLevel {
						continue
					} else if restrictNtlm, _ := innerComputer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); restrictNtlm {
						continue
					} else if err = adAnalysis.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserID, ldapSigningCache); err != nil {
						t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToLDAP.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)

			err = operation.Done()
			require.NoError(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAP)
				})); err != nil {
					t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
				} else {
					require.Len(t, results, 2)

					for _, result := range results {
						start, end, err := ops.FetchRelationshipNodes(tx, result)
						require.NoError(t, err)

						dcSet, err := adAnalysis.GetVulnerableDomainControllersForRelayNTLMtoLDAP(context.Background(), db, result)
						require.NoError(t, err)

						switch start.ID {
						case harness.NTLMCoerceAndRelayNTLMToLDAP.Group1.ID:
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToLDAP.Computer2.ID)
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAP.Computer1.ID))
						case harness.NTLMCoerceAndRelayNTLMToLDAP.Group5.ID:
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToLDAP.Computer7.ID)
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAP.Computer6.ID))
						default:
							require.FailNow(t, "unrecognized start node id")
						}

					}
				}
				return nil
			})
		})
	})

	t.Run("NTLMCoerceAndRelayNTLMToLDAPS Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToLDAPS.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAPS")

			grouplocalGroupData, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := adAnalysis.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := adAnalysis.FetchProtectedUsersMappedToDomains(testContext.Context(), db, grouplocalGroupData)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, err := innerComputer.Properties.Get(ad.DomainSID.String()).String()
					require.NoError(t, err)

					if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
						t.Fatalf("authenticated user not found for %s", domainSid)
					} else if protectedUsersForDomain, ok := protectedUsersCache[domainSid]; !ok {
						continue
					} else if ldapSigningForDomain, ok := ldapSigningCache[domainSid]; !ok {
						continue
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsVulnerableFunctionalLevel {
						continue
					} else if err = adAnalysis.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserID, ldapSigningCache); err != nil {
						t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToLDAPS.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)

			err = operation.Done()
			require.NoError(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAPS)
				})); err != nil {
					t.Fatalf("error fetching NTLM to LDAPS edges in integration test; %v", err)
				} else {
					require.Len(t, results, 3)

					for _, result := range results {
						start, end, err := ops.FetchRelationshipNodes(tx, result)
						require.NoError(t, err)

						dcSet, err := adAnalysis.GetVulnerableDomainControllersForRelayNTLMtoLDAPS(context.Background(), db, result)
						require.NoError(t, err)

						switch start.ID {
						case harness.NTLMCoerceAndRelayNTLMToLDAPS.Group1.ID:
							if end.ID != harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer2.ID && end.ID != harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer5.ID {
								require.FailNow(t, "unrecognized end node associated with Group1")
							}
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer1.ID))
						case harness.NTLMCoerceAndRelayNTLMToLDAPS.Group5.ID:
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer7.ID)
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer6.ID))
						default:
							require.FailNow(t, "unrecognized start node id")
						}
					}
				}
				return nil
			})
		})
	})

	t.Run("NTLMCoerceAndRelayNTLMToLDAPS Self Relay Does Not Create Edge", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayToLDAPSSelfRelay.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAPS - Self Relay")

			grouplocalGroupData, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := adAnalysis.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := adAnalysis.FetchProtectedUsersMappedToDomains(testContext.Context(), db, grouplocalGroupData)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, err := innerComputer.Properties.Get(ad.DomainSID.String()).String()
					require.NoError(t, err)

					if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
						t.Fatalf("authenticated user not found for %s", domainSid)
					} else if protectedUsersForDomain, ok := protectedUsersCache[domainSid]; !ok {
						continue
					} else if ldapSigningForDomain, ok := ldapSigningCache[domainSid]; !ok {
						continue
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsVulnerableFunctionalLevel {
						continue
					} else if err = adAnalysis.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserID, ldapSigningCache); err != nil {
						t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToLDAPS.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)

			err = operation.Done()
			require.NoError(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAPS)
				})); err != nil {
					t.Fatalf("error fetching NTLM to LDAPS edges in integration test; %v", err)
				} else {
					require.Len(t, results, 0)
				}
				return nil
			})
		})
	})

	t.Run("NTLMCoerceAndRelayNTLMToLDAP Self Relay Does Not Create Edge", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayToLDAPSelfRelay.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := post.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAP - Self Relay")

			grouplocalGroupData, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := adAnalysis.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := adAnalysis.FetchProtectedUsersMappedToDomains(testContext.Context(), db, grouplocalGroupData)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, err := innerComputer.Properties.Get(ad.DomainSID.String()).String()
					require.NoError(t, err)

					if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
						t.Fatalf("authenticated user not found for %s", domainSid)
					} else if protectedUsersForDomain, ok := protectedUsersCache[domainSid]; !ok {
						continue
					} else if ldapSigningForDomain, ok := ldapSigningCache[domainSid]; !ok {
						continue
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsVulnerableFunctionalLevel {
						continue
					} else if err = adAnalysis.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserID, ldapSigningCache); err != nil {
						t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToLDAP.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)

			err = operation.Done()
			require.NoError(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAP)
				})); err != nil {
					t.Fatalf("error fetching NTLM to LDAP edges in integration test; %v", err)
				} else {
					require.Len(t, results, 0)
				}
				return nil
			})
		})
	})
}

func fetchNTLMPrereqs(db graph.Database) (localGroupData *adAnalysis.LocalGroupData, computers []*graph.Node, domains []*graph.Node, authenticatedUsers map[string]graph.ID, err error) {
	cache := make(map[string]graph.ID)
	if localGroupData, err = adAnalysis.FetchLocalGroupData(context.Background(), db); err != nil {
		return nil, nil, nil, cache, err
	} else if computers, err = adAnalysis.FetchNodesByKind(context.Background(), db, ad.Computer); err != nil {
		return nil, nil, nil, cache, err
	} else if err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		if cache, err = adAnalysis.FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, nil, cache, err
	} else if domains, err = adAnalysis.FetchNodesByKind(context.Background(), db, ad.Domain); err != nil {
		return nil, nil, nil, cache, err
	} else {
		return localGroupData, computers, domains, cache, nil
	}
}
