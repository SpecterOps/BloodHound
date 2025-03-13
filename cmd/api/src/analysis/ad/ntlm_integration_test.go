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
// +build integration

package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/cardinality"

	"github.com/specterops/bloodhound/analysis"
	ad2 "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"
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
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToADCS")
		_, _, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
		require.NoError(t, err)

		cache := ad2.NewADCSCache()
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.NoError(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.NoError(t, err)
		err = cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)
		require.NoError(t, err)

		for _, domain := range domains {
			innerDomain := domain
			computerCache, err := fetchComputerCache(db, innerDomain)
			require.NoError(t, err)

			err = ad2.PostCoerceAndRelayNTLMToADCS(cache, operation, authenticatedUsers, computerCache)
			require.NoError(t, err)
		}

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
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Composition Test - CoerceAndRelayNTLMToADCS")
		_, _, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
		require.NoError(t, err)

		cache := ad2.NewADCSCache()
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.NoError(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.NoError(t, err)
		err = cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)
		require.NoError(t, err)

		for _, domain := range domains {
			innerDomain := domain
			computerCache, err := fetchComputerCache(db, innerDomain)
			require.NoError(t, err)

			err = ad2.PostCoerceAndRelayNTLMToADCS(cache, operation, authenticatedUsers, computerCache)
			require.NoError(t, err)
		}

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
				composition, err := ad2.GetCoerceAndRelayNTLMtoADCSEdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 6, len(nodes))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.AuthenticatedUsersGroup))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.CertTemplate1))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.EnterpriseCA1))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.RootCA))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.Domain))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToADCS.NTAuthStore))
			}
			return nil
		})
	})

}

func TestPostNTLMRelaySMB(t *testing.T) {
	// TODO: Add some negative tests here
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	t.Run("NTLMCoerceAndRelayNTLMToSMB Success", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToSMB.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToSMB")

			groupExpansions, computers, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := ad2.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := ad2.FetchProtectedUsersMappedToDomains(testContext.Context(), db, groupExpansions)
			require.NoError(t, err)

			compositionCounter := analysis.NewCompositionCounter()

			for _, domain := range domains {
				innerDomain := domain

				err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
					for _, computer := range computers {
						innerComputer := computer
						domainSid, _ := innerDomain.Properties.Get(ad.DomainSID.String()).String()

						if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
							t.Fatalf("authenticated user not found for %s", domainSid)
						} else if protectedUsersForDomain, ok := protectedUsersCache[domainSid]; !ok {
							continue
						} else if ldapSigningForDomain, ok := ldapSigningCache[domainSid]; !ok {
							continue
						} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
							continue
						} else if err = ad2.PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserID, &compositionCounter); err != nil {
							t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
						}
					}
					return nil
				})
				require.NoError(t, err)
			}

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

						if start.ID == harness.NTLMCoerceAndRelayNTLMToSMB.Group2.ID {
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToSMB.Computer9.ID)
						} else if start.ID == harness.NTLMCoerceAndRelayNTLMToSMB.Group1.ID {
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToSMB.Computer2.ID)
						} else {
							require.FailNow(t, "unrecognized start node id")
						}
					}
				}
				return nil
			})
		})
	})

	t.Run("NTLMCoerceAndRelayNTLMToSMB Self Relay Does Not Create Edge", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToSMBSelfRelay.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM - CoerceAndRelayNTLMToSMB - Relay To Self")

			groupExpansions, computers, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := ad2.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := ad2.FetchProtectedUsersMappedToDomains(testContext.Context(), db, groupExpansions)
			require.NoError(t, err)

			compositionCounter := analysis.NewCompositionCounter()

			for _, domain := range domains {
				innerDomain := domain

				err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
					for _, computer := range computers {
						innerComputer := computer
						domainSid, _ := innerDomain.Properties.Get(ad.DomainSID.String()).String()

						if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
							t.Fatalf("authenticated user not found for %s", domainSid)
						} else if protectedUsersForDomain, ok := protectedUsersCache[domainSid]; !ok {
							continue
						} else if ldapSigningForDomain, ok := ldapSigningCache[domainSid]; !ok {
							continue
						} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
							continue
						} else if err = ad2.PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserID, &compositionCounter); err != nil {
							t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
						}
					}
					return nil
				})
				require.NoError(t, err)
			}

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
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Composition Test - CoerceAndRelayNTLMToSMB")

		groupExpansions, computers, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
		require.NoError(t, err)

		ldapSigningCache, err := ad2.FetchLDAPSigningCache(testContext.Context(), db)
		require.NoError(t, err)

		protectedUsersCache, err := ad2.FetchProtectedUsersMappedToDomains(testContext.Context(), db, groupExpansions)
		require.NoError(t, err)

		compositionCounter := analysis.NewCompositionCounter()

		for _, domain := range domains {
			innerDomain := domain

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, _ := innerDomain.Properties.Get(ad.DomainSID.String()).String()

					if authenticatedUserID, ok := authenticatedUsers[domainSid]; !ok {
						t.Fatalf("authenticated user not found for %s", domainSid)
					} else if protectedUsersForDomain, ok := protectedUsersCache[domainSid]; !ok {
						continue
					} else if ldapSigningForDomain, ok := ldapSigningCache[domainSid]; !ok {
						continue
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
						continue
					} else if err = ad2.PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserID, &compositionCounter); err != nil {
						t.Logf("failed post processing for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)
		}

		err = operation.Done()
		require.NoError(t, err)

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
				composition, err := ad2.GetCoerceAndRelayNTLMtoSMBComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 4, len(nodes))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Computer8))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Computer9))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Group7))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Group8))
			}
			return nil
		})

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
				composition, err := ad2.GetCoerceAndRelayNTLMtoSMBComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 2, len(nodes))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Computer1))
				require.True(t, nodes.Contains(harness.NTLMCoerceAndRelayNTLMToSMB.Computer2))
			}
			return nil
		})

	})
}

func fetchComputerCache(db graph.Database, domain *graph.Node) (map[string]cardinality.Duplex[uint64], error) {
	cache := make(map[string]cardinality.Duplex[uint64])
	if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return cache, err
	} else {
		cache[domainSid] = cardinality.NewBitmap64()
		return cache, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			return tx.Nodes().Filter(
				query.And(
					query.Kind(query.Node(), ad.Computer),
					query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
				),
			).FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
				for id := range cursor.Chan() {
					cache[domainSid].Add(id.Uint64())
				}

				return cursor.Error()
			})
		})
	}
}

func TestPostCoerceAndRelayNTLMToLDAP(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	t.Run("NTLMCoerceAndRelayNTLMToLDAP Success", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToLDAP.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAP")

			groupExpansions, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := ad2.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := ad2.FetchProtectedUsersMappedToDomains(testContext.Context(), db, groupExpansions)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
						continue
					} else if err = ad2.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserID, ldapSigningCache); err != nil {
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

						dcSet, err := ad2.GetVulnerableDomainControllersForRelayNTLMtoLDAP(context.Background(), db, result)
						require.NoError(t, err)

						if start.ID == harness.NTLMCoerceAndRelayNTLMToLDAP.Group1.ID {
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToLDAP.Computer2.ID)
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAP.Computer1.ID))
						} else if start.ID == harness.NTLMCoerceAndRelayNTLMToLDAP.Group5.ID {
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToLDAP.Computer7.ID)
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAP.Computer6.ID))
						} else {
							require.FailNow(t, "unrecognized start node id")
						}

					}
				}
				return nil
			})
		})
	})

	t.Run("NTLMCoerceAndRelayNTLMToLDAPS Success", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayNTLMToLDAPS.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAPS")

			groupExpansions, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := ad2.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := ad2.FetchProtectedUsersMappedToDomains(testContext.Context(), db, groupExpansions)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
						continue
					} else if err = ad2.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserID, ldapSigningCache); err != nil {
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

						dcSet, err := ad2.GetVulnerableDomainControllersForRelayNTLMtoLDAPS(context.Background(), db, result)
						require.NoError(t, err)

						if start.ID == harness.NTLMCoerceAndRelayNTLMToLDAPS.Group1.ID {
							if end.ID != harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer2.ID && end.ID != harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer5.ID {
								require.FailNow(t, "unrecognized end node associated with Group1")
							}
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer1.ID))
						} else if start.ID == harness.NTLMCoerceAndRelayNTLMToLDAPS.Group5.ID {
							assert.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer7.ID)
							assert.True(t, dcSet.ContainsID(harness.NTLMCoerceAndRelayNTLMToLDAPS.Computer6.ID))
						} else {
							require.FailNow(t, "unrecognized start node id")
						}
					}
				}
				return nil
			})
		})
	})

	t.Run("NTLMCoerceAndRelayNTLMToLDAPS Self Relay Does Not Create Edge", func(t *testing.T) {
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.NTLMCoerceAndRelayToLDAPSelfRelay.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAP - Self Relay")

			groupExpansions, computers, _, authenticatedUsers, err := fetchNTLMPrereqs(db)
			require.NoError(t, err)

			ldapSigningCache, err := ad2.FetchLDAPSigningCache(testContext.Context(), db)
			require.NoError(t, err)

			protectedUsersCache, err := ad2.FetchProtectedUsersMappedToDomains(testContext.Context(), db, groupExpansions)
			require.NoError(t, err)

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
						continue
					} else if err = ad2.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserID, ldapSigningCache); err != nil {
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
					t.Fatalf("error fetching NTLM to LDAPS edges in integration test; %v", err)
				} else {
					require.Len(t, results, 0)
				}
				return nil
			})
		})
	})
}

func fetchNTLMPrereqs(db graph.Database) (expansions impact.PathAggregator, computers []*graph.Node, domains []*graph.Node, authenticatedUsers map[string]graph.ID, err error) {
	cache := make(map[string]graph.ID)
	if expansions, err = ad2.ExpandAllRDPLocalGroups(context.Background(), db); err != nil {
		return nil, nil, nil, cache, err
	} else if computers, err = ad2.FetchNodesByKind(context.Background(), db, ad.Computer); err != nil {
		return nil, nil, nil, cache, err
	} else if err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		if cache, err = ad2.FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, nil, cache, err
	} else if domains, err = ad2.FetchNodesByKind(context.Background(), db, ad.Domain); err != nil {
		return nil, nil, nil, cache, err
	} else {
		return expansions, computers, domains, cache, nil
	}
}
