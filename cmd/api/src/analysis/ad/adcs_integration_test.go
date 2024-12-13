// Copyright 2023 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/analysis"
	ad2 "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/graphschema"

	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test/integration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestADCSESC1(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ADCSESC1Harness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC1")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC1(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC1)
			})); err != nil {
				t.Fatalf("error fetching esc1 edges in integration test; %v", err)
			} else {
				assert.Equal(t, 8, len(results))

				//Domain 1 ESC1 edges created
				require.True(t, results.Contains(harness.ADCSESC1Harness.User13))
				require.True(t, results.Contains(harness.ADCSESC1Harness.User11))
				require.True(t, results.Contains(harness.ADCSESC1Harness.Group13))

				//Domain 2 ESC1 edges created
				require.True(t, results.Contains(harness.ADCSESC1Harness.Group22))

				//Domain 3 ESC1 edges created
				require.True(t, results.Contains(harness.ADCSESC1Harness.Group32))

				//Domain 4 ESC1 edges created
				require.True(t, results.Contains(harness.ADCSESC1Harness.Group42))
				require.True(t, results.Contains(harness.ADCSESC1Harness.Group43))

				//Group47 has Enroll on EnterpriseCA1 that has a valid chain to Domain1, a domain that is marked as not collected
				require.True(t, results.Contains(harness.ADCSESC1Harness.Group47))

			}

			// Domains 1 and 4 have multiple ADCSESC1 edges so we will just test edge comp of Domains 2 and 3

			// Domain 2 Edge Composition
			if edge, err := tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.ADCSESC1),
					query.Equals(query.EndID(), harness.ADCSESC1Harness.Domain2.ID),
				)
			}).First(); err != nil {
				t.Fatalf("error fetching esc1 domain 2 edges in integration test; %v", err)
			} else {
				comp, err := ad2.GetADCSESC1EdgeComposition(context.Background(), db, edge)
				assert.Nil(t, err)

				domain2Nodes := comp.AllNodes()
				assert.Len(t, domain2Nodes, 7)
				require.True(t, domain2Nodes.Contains(harness.ADCSESC1Harness.Group22))
				require.True(t, domain2Nodes.Contains(harness.ADCSESC1Harness.CertTemplate2))
				require.True(t, domain2Nodes.Contains(harness.ADCSESC1Harness.EnterpriseCA21))
				require.True(t, domain2Nodes.Contains(harness.ADCSESC1Harness.EnterpriseCA23))
				require.True(t, domain2Nodes.Contains(harness.ADCSESC1Harness.AuthStore2))
				require.True(t, domain2Nodes.Contains(harness.ADCSESC1Harness.RootCA2))
				require.True(t, domain2Nodes.Contains(harness.ADCSESC1Harness.Domain2))
			}

			// Domain 3 Edge Composition
			if edge, err := tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.ADCSESC1),
					query.Equals(query.EndID(), harness.ADCSESC1Harness.Domain3.ID),
				)
			}).First(); err != nil {
				t.Fatalf("error fetching esc1 edges in integration test; %v", err)
			} else {
				comp, err := ad2.GetADCSESC1EdgeComposition(context.Background(), db, edge)
				assert.Nil(t, err)

				domain3Nodes := comp.AllNodes()
				assert.Len(t, domain3Nodes, 6)
				require.True(t, domain3Nodes.Contains(harness.ADCSESC1Harness.Group32))
				require.True(t, domain3Nodes.Contains(harness.ADCSESC1Harness.CertTemplate3))
				require.True(t, domain3Nodes.Contains(harness.ADCSESC1Harness.EnterpriseCA31))
				require.True(t, domain3Nodes.Contains(harness.ADCSESC1Harness.RootCA3))
				require.True(t, domain3Nodes.Contains(harness.ADCSESC1Harness.AuthStore3))
				require.True(t, domain3Nodes.Contains(harness.ADCSESC1Harness.Domain3))
			}

			return nil
		})
		assert.Nil(t, err)
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ADCSESC1HarnessAuthUsers.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC1 Authenticated Users")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC1(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC1)
			})); err != nil {
				t.Fatalf("error fetching esc1 edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ADCSESC1HarnessAuthUsers.Group1))
			}

			return nil
		})
		assert.Nil(t, err)
	})
}

func TestGoldenCert(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ADCSGoldenCertHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - Golden Cert")

		_, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostGoldenCert(ctx, tx, outC, innerDomain, innerEnterpriseCA); err != nil {
							t.Logf("failed post processing for %s: %v", ad.GoldenCert.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.GoldenCert)
			})); err != nil {
				t.Fatalf("error fetching golden cert edges in integration test; %v", err)
			} else {
				assert.True(t, len(results) == 2)

				//Positive Cases
				assert.True(t, results.Contains(harness.ADCSGoldenCertHarness.Computer1))
				assert.True(t, results.Contains(harness.ADCSGoldenCertHarness.Computer3))

				//Negative Cases
				assert.False(t, results.Contains(harness.ADCSGoldenCertHarness.Computer21))
				assert.False(t, results.Contains(harness.ADCSGoldenCertHarness.Computer22))
				assert.False(t, results.Contains(harness.ADCSGoldenCertHarness.Computer23))
			}
			return nil
		})
	})

}

func TestIssuedSignedBy(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.IssuedSignedByHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - IssuedSignedBy")

		if rootCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.RootCA); err != nil {
			t.Logf("failed fetching rootCA nodes: %v", err)
		} else if enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA); err != nil {
			t.Logf("failed fetching enterpriseCA nodes: %v", err)
		} else if aiaCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.AIACA); err != nil {
			t.Logf("failed fetching AIACA nodes: %v", err)
		} else if err := ad2.PostIssuedSignedBy(operation, enterpriseCertAuthorities, rootCertAuthorities, aiaCertAuthorities); err != nil {
			t.Logf("failed post processing for %s: %v", ad.IssuedSignedBy.String(), err)
		}

		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.IssuedSignedBy)
			})); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edges in integration test; %v", err)
			} else {
				assert.Equal(t, 12, len(results))
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.EnterpriseCA2.ID, harness.IssuedSignedByHarness.EnterpriseCA1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (1) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.EnterpriseCA2.ID, harness.IssuedSignedByHarness.AIACA2_1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (2) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.EnterpriseCA1.ID, harness.IssuedSignedByHarness.RootCA2.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (3) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.EnterpriseCA1.ID, harness.IssuedSignedByHarness.AIACA1_2.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (4) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.RootCA2.ID, harness.IssuedSignedByHarness.RootCA1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (5) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.RootCA2.ID, harness.IssuedSignedByHarness.AIACA1_1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (6) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.AIACA2_2.ID, harness.IssuedSignedByHarness.EnterpriseCA1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (7) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.AIACA2_2.ID, harness.IssuedSignedByHarness.AIACA2_1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (8) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.AIACA2_1.ID, harness.IssuedSignedByHarness.RootCA2.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (9) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.AIACA2_1.ID, harness.IssuedSignedByHarness.AIACA1_2.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (10) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.AIACA1_2.ID, harness.IssuedSignedByHarness.RootCA1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (11) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.IssuedSignedByHarness.AIACA1_2.ID, harness.IssuedSignedByHarness.AIACA1_1.ID, ad.IssuedSignedBy); err != nil {
				t.Fatalf("error fetching IssuedSignedBy edge (12) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			return nil
		})
	})
}

func TestEnterpriseCAFor(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.EnterpriseCAForHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - EnterpriseCAFor")

		if enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA); err != nil {
			t.Logf("failed fetching enterpriseCA nodes: %v", err)
		} else if err := ad2.PostEnterpriseCAFor(operation, enterpriseCertAuthorities); err != nil {
			t.Logf("failed post processing for %s: %v", ad.EnterpriseCAFor.String(), err)
		}

		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.EnterpriseCAFor)
			})); err != nil {
				t.Fatalf("error fetching EnterpriseCAFor edges in integration test; %v", err)
			} else {
				assert.Equal(t, 3, len(results))
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.EnterpriseCAForHarness.EnterpriseCA1.ID, harness.EnterpriseCAForHarness.RootCA1.ID, ad.EnterpriseCAFor); err != nil {
				t.Fatalf("error fetching EnterpriseCAFor edge (1) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.EnterpriseCAForHarness.EnterpriseCA1.ID, harness.EnterpriseCAForHarness.AIACA1_1.ID, ad.EnterpriseCAFor); err != nil {
				t.Fatalf("error fetching EnterpriseCAFor edge (1) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.EnterpriseCAForHarness.EnterpriseCA2.ID, harness.EnterpriseCAForHarness.AIACA2_1.ID, ad.EnterpriseCAFor); err != nil {
				t.Fatalf("error fetching EnterpriseCAFor edge (1) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			return nil
		})
	})
}

func TestTrustedForNTAuth(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			harness.TrustedForNTAuthHarness.Setup(testContext)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			// post `TrustedForNTAuth` edges
			operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - TrustedForNTAuth")

			if err := ad2.PostTrustedForNTAuth(context.Background(), db, operation); err != nil {
				t.Logf("failed post processing for %s: %v", ad.TrustedForNTAuth.String(), err)
			}

			operation.Done()

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				if results, err := ops.FetchStartNodes(
					tx.Relationships().Filterf(func() graph.Criteria {
						return query.Kind(query.Relationship(), ad.TrustedForNTAuth)
					}),
				); err != nil {
					t.Fatalf("error fetching TrustedForNTAuth relationships; %v", err)
				} else {
					assert.True(t, len(results) == 2)

					//Positive Cases
					assert.True(t, results.Contains(harness.TrustedForNTAuthHarness.EnterpriseCA1))
					assert.True(t, results.Contains(harness.TrustedForNTAuthHarness.EnterpriseCA2))

					//Negative Cases
					assert.False(t, results.Contains(harness.TrustedForNTAuthHarness.EnterpriseCA3))
				}
				return nil
			})
		})
}

func TestEnrollOnBehalfOf(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.EnrollOnBehalfOfHarness1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		v1Templates := make([]*graph.Node, 0)
		v2Templates := make([]*graph.Node, 0)

		for _, template := range certTemplates {
			if version, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				continue
			} else if version == 1 {
				v1Templates = append(v1Templates, template)
			} else if version >= 2 {
				v2Templates = append(v2Templates, template)
			}
		}

		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			results, err := ad2.EnrollOnBehalfOfVersionOne(tx, v1Templates, certTemplates, harness.EnrollOnBehalfOfHarness1.Domain1)
			require.Nil(t, err)

			require.Len(t, results, 3)

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarness1.CertTemplate11.ID,
				ToID:   harness.EnrollOnBehalfOfHarness1.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarness1.CertTemplate13.ID,
				ToID:   harness.EnrollOnBehalfOfHarness1.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarness1.CertTemplate12.ID,
				ToID:   harness.EnrollOnBehalfOfHarness1.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			results, err := ad2.EnrollOnBehalfOfVersionTwo(tx, v2Templates, certTemplates, harness.EnrollOnBehalfOfHarness1.Domain1)
			require.Nil(t, err)

			require.Len(t, results, 0)

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.EnrollOnBehalfOfHarness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		v1Templates := make([]*graph.Node, 0)
		v2Templates := make([]*graph.Node, 0)

		for _, template := range certTemplates {
			if version, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				continue
			} else if version == 1 {
				v1Templates = append(v1Templates, template)
			} else if version >= 2 {
				v2Templates = append(v2Templates, template)
			}
		}

		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			results, err := ad2.EnrollOnBehalfOfVersionOne(tx, v1Templates, certTemplates, harness.EnrollOnBehalfOfHarness2.Domain2)
			require.Nil(t, err)

			require.Len(t, results, 0)
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			results, err := ad2.EnrollOnBehalfOfVersionTwo(tx, v2Templates, certTemplates, harness.EnrollOnBehalfOfHarness2.Domain2)
			require.Nil(t, err)

			require.Len(t, results, 1)
			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarness2.CertTemplate21.ID,
				ToID:   harness.EnrollOnBehalfOfHarness2.CertTemplate23.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.EnrollOnBehalfOfHarness3.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - EnrollOnBehalfOf 3")

		_, enterpriseCertAuthorities, certTemplates, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		if err := ad2.PostEnrollOnBehalfOf(domains, enterpriseCertAuthorities, certTemplates, cache, operation); err != nil {
			t.Logf("failed post processing for %s: %v", ad.EnrollOnBehalfOf.String(), err)
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if startNodes, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.EnrollOnBehalfOf)
			})); err != nil {
				t.Fatalf("error fetching EnrollOnBehalfOf edges in integration test; %v", err)
			} else if endNodes, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.EnrollOnBehalfOf)
			})); err != nil {
				t.Fatalf("error fetching EnrollOnBehalfOf edges in integration test; %v", err)
			} else {
				require.Len(t, startNodes, 2)
				require.True(t, startNodes.Contains(harness.EnrollOnBehalfOfHarness3.CertTemplate11))
				require.True(t, startNodes.Contains(harness.EnrollOnBehalfOfHarness3.CertTemplate12))

				require.Len(t, endNodes, 2)
				require.True(t, startNodes.Contains(harness.EnrollOnBehalfOfHarness3.CertTemplate12))
				require.True(t, startNodes.Contains(harness.EnrollOnBehalfOfHarness3.CertTemplate12))
			}

			return nil
		})
	})
}

func TestADCSESC3(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC3Harness1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC3")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC3(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC3)
			})); err != nil {
				t.Fatalf("error fetching esc3 edges in integration test; %v", err)
			} else {
				assert.Equal(t, 3, len(results))

				require.True(t, results.Contains(harness.ESC3Harness1.Computer1))
				require.True(t, results.Contains(harness.ESC3Harness1.Group2))
				require.True(t, results.Contains(harness.ESC3Harness1.User1))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC3Harness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC3")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC3(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC3)
			})); err != nil {
				t.Fatalf("error fetching esc3 edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC3Harness2.User1))
			}

			if edge, err := tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC3)
			}).First(); err != nil {
				t.Fatalf("error fetching esc3 edges in integration test; %v", err)
			} else {
				comp, err := ad2.GetADCSESC3EdgeComposition(context.Background(), db, edge)
				assert.Nil(t, err)
				assert.Equal(t, 8, len(comp.AllNodes()))
				assert.False(t, comp.AllNodes().Contains(harness.ESC3Harness2.User2))
				// CT3 requires DNS name meaning user3 -> domain is not a valid ESC3
				assert.False(t, comp.AllNodes().Contains(harness.ESC3Harness2.User3))
				// Enroller does not have DelegatedEnrollmentAgent edge on CT4
				assert.False(t, comp.AllNodes().Contains(harness.ESC3Harness2.CertTemplate4))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC3Harness3.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC3")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC3(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC3)
			})); err != nil {
				t.Fatalf("error fetching esc3 edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC3Harness3.Group1))
			}

			if edge, err := tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC3)
			}).First(); err != nil {
				t.Fatalf("error fetching esc3 edges in integration test; %v", err)
			} else {
				comp, err := ad2.GetADCSESC3EdgeComposition(context.Background(), db, edge)
				assert.Nil(t, err)
				assert.Equal(t, 7, len(comp.AllNodes()))
				assert.False(t, comp.AllNodes().Contains(harness.ESC3Harness3.User2))
			}
			return nil
		})
	})
}

func TestADCSESC4(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC4Template1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC4 template 1")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC4(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC4.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC4)
			})); err != nil {
				t.Fatalf("error fetching esc4 edges in integration test; %v", err)
			} else {
				require.Equal(t, 14, len(results))

				require.True(t, results.Contains(harness.ESC4Template1.Group11))
				require.True(t, results.Contains(harness.ESC4Template1.Group12))
				require.True(t, results.Contains(harness.ESC4Template1.Group13))
				require.True(t, results.Contains(harness.ESC4Template1.Group14))
				require.True(t, results.Contains(harness.ESC4Template1.Group15))

				require.True(t, results.Contains(harness.ESC4Template1.Group21))
				require.True(t, results.Contains(harness.ESC4Template1.Group22))

				require.True(t, results.Contains(harness.ESC4Template1.Group31))
				require.True(t, results.Contains(harness.ESC4Template1.Group32))

				require.True(t, results.Contains(harness.ESC4Template1.Group41))
				require.True(t, results.Contains(harness.ESC4Template1.Group42))
				require.True(t, results.Contains(harness.ESC4Template1.Group43))
				require.True(t, results.Contains(harness.ESC4Template1.Group44))
				require.True(t, results.Contains(harness.ESC4Template1.Group45))

			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC4Template2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC4 template 2")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC4(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC4.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC4)
			})); err != nil {
				t.Fatalf("error fetching esc4 edges in integration test; %v", err)
			} else {
				require.Equal(t, 18, len(results))

				require.True(t, results.Contains(harness.ESC4Template2.Group11))
				require.True(t, results.Contains(harness.ESC4Template2.Group12))

				require.True(t, results.Contains(harness.ESC4Template2.Group21))
				require.True(t, results.Contains(harness.ESC4Template2.Group22))
				require.True(t, results.Contains(harness.ESC4Template2.Group25))

				require.True(t, results.Contains(harness.ESC4Template2.Group31))
				require.True(t, results.Contains(harness.ESC4Template2.Group32))
				require.True(t, results.Contains(harness.ESC4Template2.Group33))
				require.True(t, results.Contains(harness.ESC4Template2.Group35))

				require.True(t, results.Contains(harness.ESC4Template2.Group41))
				require.True(t, results.Contains(harness.ESC4Template2.Group42))
				require.True(t, results.Contains(harness.ESC4Template2.Group44))
				require.True(t, results.Contains(harness.ESC4Template2.Group45))

				require.True(t, results.Contains(harness.ESC4Template2.Group51))
				require.True(t, results.Contains(harness.ESC4Template2.Group52))
				require.True(t, results.Contains(harness.ESC4Template2.Group53))
				require.True(t, results.Contains(harness.ESC4Template2.Group54))
				require.True(t, results.Contains(harness.ESC4Template2.Group55))

			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC4Template3.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC4 template 3")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC4(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC4.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC4)
			})); err != nil {
				t.Fatalf("error fetching esc4 edges in integration test; %v", err)
			} else {
				require.Equal(t, 4, len(results))

				require.True(t, results.Contains(harness.ESC4Template3.Group11))
				require.True(t, results.Contains(harness.ESC4Template3.Group17))
				require.True(t, results.Contains(harness.ESC4Template3.Group18))
				require.True(t, results.Contains(harness.ESC4Template3.Group19))
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC4Template4.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC4 template 4")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC4(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC4.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC4)
			})); err != nil {
				t.Fatalf("error fetching esc4 edges in integration test; %v", err)
			} else {
				require.Equal(t, 5, len(results))

				require.True(t, results.Contains(harness.ESC4Template4.Group1))
				require.True(t, results.Contains(harness.ESC4Template4.Computer1))
				require.True(t, results.Contains(harness.ESC4Template4.User1))
				require.True(t, results.Contains(harness.ESC4Template4.Group12))
				require.True(t, results.Contains(harness.ESC4Template4.Group13))
			}

			return nil
		})

	})

}

func TestADCSESC4Composition(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC4Template1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC4 template 1")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC4(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC4.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		// first scenario: composition reveals that principal `Group11` has esc4 on the domain via enrollment on ECA and GenericAll on `CertTemplate1`
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC4),
						query.Equals(query.StartProperty(common.Name.String()), "Group11"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching esc4 edge in integration test: %v", err)
			} else {
				composition, err := ad2.GetADCSESC4EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 7, len(nodes))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group11))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group0))
				require.True(t, nodes.Contains(harness.ESC4Template1.CertTemplate1))
				require.True(t, nodes.Contains(harness.ESC4Template1.EnterpriseCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.RootCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.NTAuthStore))
				require.True(t, nodes.Contains(harness.ESC4Template1.Domain))
			}
			return nil
		})

		// second scenario: composition reveals that principal `Group12` has esc4 on the domain via enrollment on ECA, and GenericWrite on `CertTemplate1`
		db.ReadTransaction(
			context.Background(),
			func(tx graph.Transaction) error {
				if edge, err := tx.Relationships().Filterf(
					func() graph.Criteria {
						return query.And(
							query.Kind(query.Relationship(), ad.ADCSESC4),
							query.Equals(query.StartProperty(common.Name.String()), "Group12"),
						)
					}).First(); err != nil {
					t.Fatalf("error fetching esc4 edge in integration test: %v", err)
				} else {
					composition, err := ad2.GetADCSESC4EdgeComposition(context.Background(), db, edge)
					require.Nil(t, err)

					nodes := composition.AllNodes()

					require.Equal(t, 7, len(nodes))
					require.True(t, nodes.Contains(harness.ESC4Template1.Group12))
					require.True(t, nodes.Contains(harness.ESC4Template1.Group0))
					require.True(t, nodes.Contains(harness.ESC4Template1.CertTemplate1))
					require.True(t, nodes.Contains(harness.ESC4Template1.EnterpriseCA))
					require.True(t, nodes.Contains(harness.ESC4Template1.RootCA))
					require.True(t, nodes.Contains(harness.ESC4Template1.NTAuthStore))
					require.True(t, nodes.Contains(harness.ESC4Template1.Domain))

					// assert that `GenericWrite` and `Enroll` edges exist between group12 and cert template
					for _, p := range composition.Paths() {
						for _, rel := range p.Edges {
							if rel.Kind.Is(ad.GenericWrite) || rel.Kind.Is(ad.Enroll) && rel.StartID == harness.ESC4Template1.Group12.ID {
								require.True(t, rel.StartID == harness.ESC4Template1.Group12.ID)
								require.True(t, rel.EndID == harness.ESC4Template1.CertTemplate1.ID)
							}

						}
					}
				}

				return nil
			})

		// third scenario: composition reveals that principal `Group13` has esc4 on the domain via enrollment on ECA, and `Enroll` and WritePKINameFlag` on `CertTemplate1`
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC4),
						query.Equals(query.StartProperty(common.Name.String()), "Group13"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching esc4 edge in integration test: %v", err)
			} else {
				composition, err := ad2.GetADCSESC4EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 7, len(nodes))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group13))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group0))
				require.True(t, nodes.Contains(harness.ESC4Template1.CertTemplate1))
				require.True(t, nodes.Contains(harness.ESC4Template1.EnterpriseCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.RootCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.NTAuthStore))
				require.True(t, nodes.Contains(harness.ESC4Template1.Domain))

				// assert that group13 has outbound edges `WritePKINameFlag` and `Enroll` to CertTemplate1
				for _, p := range composition.Paths() {
					for _, rel := range p.Edges {
						if (rel.Kind.Is(ad.WritePKINameFlag) || rel.Kind.Is(ad.Enroll)) && rel.StartID == harness.ESC4Template1.Group13.ID {
							require.True(t, rel.StartID == harness.ESC4Template1.Group13.ID)
							require.True(t, rel.EndID == harness.ESC4Template1.CertTemplate1.ID)
						}

					}
				}
			}

			return nil
		})

		// fourth scenario: composition reveals that principal `Group14` has esc4 on the domain via enrollment on ECA, and `Enroll` and WritePKINameFlag` on `CertTemplate1`
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC4),
						query.Equals(query.StartProperty(common.Name.String()), "Group14"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching esc4 edge in integration test: %v", err)
			} else {
				composition, err := ad2.GetADCSESC4EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 7, len(nodes))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group14))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group0))
				require.True(t, nodes.Contains(harness.ESC4Template1.CertTemplate1))
				require.True(t, nodes.Contains(harness.ESC4Template1.EnterpriseCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.RootCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.NTAuthStore))
				require.True(t, nodes.Contains(harness.ESC4Template1.Domain))

				// assert that group14 has outbound edges `WritePKIEnrollmentFlag` and `Enroll` to CertTemplate1
				for _, p := range composition.Paths() {
					for _, rel := range p.Edges {
						if (rel.Kind.Is(ad.WritePKIEnrollmentFlag) || rel.Kind.Is(ad.Enroll)) && rel.StartID == harness.ESC4Template1.Group14.ID {
							require.True(t, rel.StartID == harness.ESC4Template1.Group14.ID)
							require.True(t, rel.EndID == harness.ESC4Template1.CertTemplate1.ID)
						}

					}
				}
			}

			return nil
		})

		// fifth scenario: composition reveals that principal `Group15` has esc4 on the domain via enrollment on ECA, and `Enroll`, `WritePKINameFlag`, and `WritePKIEnrollmentFlag` on `CertTemplate1`
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC4),
						query.Equals(query.StartProperty(common.Name.String()), "Group15"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching esc4 edge in integration test: %v", err)
			} else {
				composition, err := ad2.GetADCSESC4EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				nodes := composition.AllNodes()

				require.Equal(t, 7, len(nodes))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group15))
				require.True(t, nodes.Contains(harness.ESC4Template1.Group0))
				require.True(t, nodes.Contains(harness.ESC4Template1.CertTemplate1))
				require.True(t, nodes.Contains(harness.ESC4Template1.EnterpriseCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.RootCA))
				require.True(t, nodes.Contains(harness.ESC4Template1.NTAuthStore))
				require.True(t, nodes.Contains(harness.ESC4Template1.Domain))

				// assert that group15 has 3 outbound edges: `WritePKIEnrollmentFlag`, `WritePKINameFlag`, and `Enroll` to CertTemplate1
				for _, p := range composition.Paths() {
					for _, rel := range p.Edges {
						if (rel.Kind.Is(ad.WritePKIEnrollmentFlag) || rel.Kind.Is(ad.WritePKINameFlag) || rel.Kind.Is(ad.Enroll)) && rel.StartID == harness.ESC4Template1.Group15.ID {
							require.True(t, rel.StartID == harness.ESC4Template1.Group15.ID)
							require.True(t, rel.EndID == harness.ESC4Template1.CertTemplate1.ID)
						}

					}
				}
			}

			return nil
		})
	})
}

func TestADCSESC9a(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aPrincipalHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 6, len(results))

				assert.True(t, results.Contains(harness.ESC9aPrincipalHarness.Group1))
				assert.True(t, results.Contains(harness.ESC9aPrincipalHarness.Group2))
				assert.True(t, results.Contains(harness.ESC9aPrincipalHarness.Group3))
				assert.True(t, results.Contains(harness.ESC9aPrincipalHarness.Group4))
				assert.True(t, results.Contains(harness.ESC9aPrincipalHarness.Group5))
				assert.True(t, results.Contains(harness.ESC9aPrincipalHarness.User2))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarness1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 3, len(results))

				assert.True(t, results.Contains(harness.ESC9aHarness1.Group1))
				assert.True(t, results.Contains(harness.ESC9aHarness1.Group2))
				assert.True(t, results.Contains(harness.ESC9aHarness1.Group3))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 4, len(results))

				assert.True(t, results.Contains(harness.ESC9aHarness2.User5))
				assert.True(t, results.Contains(harness.ESC9aHarness2.Computer5))
				assert.True(t, results.Contains(harness.ESC9aHarness2.Group5))
				assert.True(t, results.Contains(harness.ESC9aHarness2.Group6))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 4, len(results))

				assert.True(t, results.Contains(harness.ESC9aHarness2.User5))
				assert.True(t, results.Contains(harness.ESC9aHarness2.Computer5))
				assert.True(t, results.Contains(harness.ESC9aHarness2.Group5))
				assert.True(t, results.Contains(harness.ESC9aHarness2.Group6))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarnessVictim.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 2, len(results))

				assert.True(t, results.Contains(harness.ESC9aHarnessVictim.Group1))
				assert.True(t, results.Contains(harness.ESC9aHarnessVictim.Group2))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarnessECA.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))

				assert.True(t, results.Contains(harness.ESC9aHarnessECA.Group1))
			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))
				edge := results[0]

				if edgeComp, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc9: %v", err)
				} else {
					nodes := edgeComp.AllNodes().Slice()
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.Group1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.Domain1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.User1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.CertTemplate1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.EnterpriseCA1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.NTAuthStore1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.RootCA1)

					assert.Equal(t, len(nodes), 8)
				}
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarnessDC1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				require.Equal(t, 3, len(results))

				require.True(t, results.Contains(harness.ESC9aHarnessDC1.Group0))
				require.True(t, results.Contains(harness.ESC9aHarnessDC1.Group1))
				require.True(t, results.Contains(harness.ESC9aHarnessDC1.Group2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarnessDC2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC9aHarnessDC2.Group0))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aHarnessAuthUsers.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9A Authenticated Users")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9a)
			})); err != nil {
				t.Fatalf("error fetching esc9a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 3, len(results))

				require.True(t, results.Contains(harness.ESC9aHarnessAuthUsers.Group1))
				require.True(t, results.Contains(harness.ESC9aHarnessAuthUsers.Group2))
				require.True(t, results.Contains(harness.ESC9aHarnessAuthUsers.Group4))
			}

			return nil
		})
		assert.Nil(t, err)
	})
}

func TestADCSESC9b(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bPrincipalHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				assert.Equal(t, 6, len(results))

				require.True(t, results.Contains(harness.ESC9bPrincipalHarness.Group1))
				require.True(t, results.Contains(harness.ESC9bPrincipalHarness.Group2))
				require.True(t, results.Contains(harness.ESC9bPrincipalHarness.Group3))
				require.True(t, results.Contains(harness.ESC9bPrincipalHarness.Group4))
				require.True(t, results.Contains(harness.ESC9bPrincipalHarness.Group5))
				require.True(t, results.Contains(harness.ESC9bPrincipalHarness.Computer2))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bHarness1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				assert.Equal(t, 2, len(results))

				assert.True(t, results.Contains(harness.ESC9bHarness1.Group1))
				assert.True(t, results.Contains(harness.ESC9bHarness1.Group2))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bHarness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				assert.Equal(t, 2, len(results))

				assert.True(t, results.Contains(harness.ESC9bHarness2.User5))
				assert.True(t, results.Contains(harness.ESC9bHarness2.Computer5))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bHarnessVictim.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				assert.Equal(t, 2, len(results))

				assert.True(t, results.Contains(harness.ESC9bHarnessVictim.Group1))
				assert.True(t, results.Contains(harness.ESC9bHarnessVictim.Group2))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bHarnessECA.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))

				assert.True(t, results.Contains(harness.ESC9bHarnessECA.Group1))
			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))
				edge := results[0]

				if composition, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc9: %v", err)
				} else {
					names := []string{}
					for _, node := range composition.AllNodes() {
						name, _ := node.Properties.Get(common.Name.String()).String()
						names = append(names, name)
					}
					require.Equal(t, 8, len(composition.AllNodes()))
					require.Contains(t, names, "Group1")
					require.Contains(t, names, "Domain1")
					require.Contains(t, names, "DC1")
					require.Contains(t, names, "Computer1")
					require.Contains(t, names, "CertTemplate1")
					require.Contains(t, names, "EnterpriseCA1")
					require.Contains(t, names, "NTAuthStore1")
					require.Contains(t, names, "RootCA1")
				}
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bHarnessDC1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				require.Equal(t, 3, len(results))

				require.True(t, results.Contains(harness.ESC9bHarnessDC1.Group0))
				require.True(t, results.Contains(harness.ESC9bHarnessDC1.Group1))
				require.True(t, results.Contains(harness.ESC9bHarnessDC1.Group2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bHarnessDC2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC9b)
			})); err != nil {
				t.Fatalf("error fetching esc9b edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC9bHarnessDC2.Group0))
			}
			return nil
		})
	})
}

func TestADCSESC6a(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6aHarnessPrincipalEdges.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6a)
			})); err != nil {
				t.Fatalf("error fetching esc6a edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC6aHarnessPrincipalEdges.Group1))
				require.True(t, results.Contains(harness.ESC6aHarnessPrincipalEdges.Group2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6aHarnessECA.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6a)
			})); err != nil {
				t.Fatalf("error fetching esc6a edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))
				name, _ := results.Pick().Properties.Get(common.Name.String()).String()
				require.Equal(t, name, "Group0")
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6aHarnessTemplate1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6a)
			})); err != nil {
				t.Fatalf("error fetching esc6a edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))
				names := []string{}
				for _, result := range results.Slice() {
					name, _ := result.Properties.Get(common.Name.String()).String()
					names = append(names, name)
				}
				require.Contains(t, names, "Group1", "Group2")
			}

			if edge, err := tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.ADCSESC6a),
					query.Equals(query.StartProperty(common.Name.String()), "Group1"),
				)
			}).First(); err != nil {
				t.Fatalf("error fetching esc6a edges in integration test; %v", err)
			} else {
				composition, err := ad2.GetADCSESC6EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)
				names := []string{}
				for _, node := range composition.AllNodes() {
					name, _ := node.Properties.Get(common.Name.String()).String()
					names = append(names, name)
				}
				require.Equal(t, 7, len(composition.AllNodes()))
				require.Contains(t, names, "Group1")
				require.Contains(t, names, "Group0")
				require.Contains(t, names, "CertTemplate1")
				require.Contains(t, names, "EnterpriseCA")
				require.Contains(t, names, "RootCA")
				require.Contains(t, names, "NTAuthStore")
				require.Contains(t, names, "Domain")
			}

			if edge, err := tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.ADCSESC6a),
					query.Equals(query.StartProperty(common.Name.String()), "Group2"),
				)
			}).First(); err != nil {
				t.Fatalf("error fetching esc6a edges in integration test; %v", err)
			} else {
				composition, err := ad2.GetADCSESC6EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)
				names := []string{}
				for _, node := range composition.AllNodes() {
					name, _ := node.Properties.Get(common.Name.String()).String()
					names = append(names, name)
				}
				require.Equal(t, 7, len(composition.AllNodes()))
				require.Contains(t, names, "Group2")
				require.Contains(t, names, "Group0")
				require.Contains(t, names, "CertTemplate2")
				require.Contains(t, names, "EnterpriseCA")
				require.Contains(t, names, "RootCA")
				require.Contains(t, names, "NTAuthStore")
				require.Contains(t, names, "Domain")
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6aHarnessTemplate2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6a)
			})); err != nil {
				t.Fatalf("error fetching esc6a edges in integration test; %v", err)
			} else {
				names := []string{}
				for _, result := range results.Slice() {
					name, _ := result.Properties.Get(common.Name.String()).String()
					names = append(names, name)
				}
				require.Equal(t, 7, len(results))
				require.NotContains(t, names, "User2")
				require.NotContains(t, names, "User3")
			}
			return nil
		})
	})
}

func TestADCSESC6b(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6bTemplate1Harness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6b template 1")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6b)
			})); err != nil {
				t.Fatalf("error fetching esc6b edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC6bTemplate1Harness.Group1))
				require.True(t, results.Contains(harness.ESC6bTemplate1Harness.Group1))

				require.False(t, results.Contains(harness.ESC6bTemplate1Harness.Group3))
				require.False(t, results.Contains(harness.ESC6bTemplate1Harness.Group4))
				require.False(t, results.Contains(harness.ESC6bTemplate1Harness.Group5))

			}

			// run edge composition against the group1 node which has an outbound esc 6b edge
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC6b),
						query.Equals(query.StartProperty(common.Name.String()), "Group1"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching esc6b edge in integration test: %v", err)
			} else {
				composition, err := ad2.GetADCSESC6EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				require.Equal(t, 8, len(composition.AllNodes()))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.Group0))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.Group1))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.CertTemplate1))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.EnterpriseCA))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.RootCA))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.NTAuthStore))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.DC))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.Domain))

			}

			// run edge composition against the group2 node which has an outbound esc 6b edge
			if edge, err := tx.Relationships().Filterf(
				func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC6b),
						query.Equals(query.StartProperty(common.Name.String()), "Group2"),
					)
				}).First(); err != nil {
				t.Fatalf("error fetching esc6b edge in integration test: %v", err)
			} else {
				composition, err := ad2.GetADCSESC6EdgeComposition(context.Background(), db, edge)
				require.Nil(t, err)

				require.Equal(t, 8, len(composition.AllNodes()))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.Group0))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.Group2))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.CertTemplate2))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.EnterpriseCA))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.RootCA))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.NTAuthStore))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.DC))
				require.True(t, composition.AllNodes().Contains(harness.ESC6bTemplate1Harness.Domain))

			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6bECAHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6b eca")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(
				tx.Relationships().Filterf(func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.ADCSESC6b)
				})); err != nil {
				t.Fatalf("error fetching esc6b edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC6bECAHarness.Group0))
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6bPrincipalEdgesHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6b principal edges")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6b)
			})); err != nil {
				t.Fatalf("error fetching esc6b edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC6bPrincipalEdgesHarness.Group1))
				require.True(t, results.Contains(harness.ESC6bPrincipalEdgesHarness.Group2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6bTemplate2Harness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6b)
			})); err != nil {
				t.Fatalf("error fetching esc6b edges in integration test; %v", err)
			} else {
				require.Equal(t, 7, len(results))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.User1))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Computer1))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group1))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Computer2))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group2))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Computer3))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group3))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6bHarnessDC1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6b)
			})); err != nil {
				t.Fatalf("error fetching esc6b edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC6bHarnessDC1.Group0))
				require.True(t, results.Contains(harness.ESC6bHarnessDC1.Group1))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC6bHarnessDC2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC6b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6b)
			})); err != nil {
				t.Fatalf("error fetching esc6b edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC6bHarnessDC2.Group0))
			}
			return nil
		})
	})
}

func FetchADCSPrereqs(db graph.Database) (impact.PathAggregator, []*graph.Node, []*graph.Node, []*graph.Node, ad2.ADCSCache, error) {
	if expansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db); err != nil {
		return nil, nil, nil, nil, ad2.ADCSCache{}, err
	} else if eca, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA); err != nil {
		return nil, nil, nil, nil, ad2.ADCSCache{}, err
	} else if certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate); err != nil {
		return nil, nil, nil, nil, ad2.ADCSCache{}, err
	} else if domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain); err != nil {
		return nil, nil, nil, nil, ad2.ADCSCache{}, err
	} else {
		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, eca, certTemplates, domains)
		return expansions, eca, certTemplates, domains, cache, nil
	}
}

func TestADCSESC10a(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10aPrincipalHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				require.Equal(t, 6, len(results))

				require.True(t, results.Contains(harness.ESC10aPrincipalHarness.Group1))
				require.True(t, results.Contains(harness.ESC10aPrincipalHarness.Group2))
				require.True(t, results.Contains(harness.ESC10aPrincipalHarness.Group3))
				require.True(t, results.Contains(harness.ESC10aPrincipalHarness.Group4))
				require.True(t, results.Contains(harness.ESC10aPrincipalHarness.Group5))
				require.True(t, results.Contains(harness.ESC10aPrincipalHarness.User2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10aHarness1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				require.Equal(t, 3, len(results))

				require.True(t, results.Contains(harness.ESC10aHarness1.Group1))
				require.True(t, results.Contains(harness.ESC10aHarness1.Group2))
				require.True(t, results.Contains(harness.ESC10aHarness1.Group3))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10aHarness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				require.Equal(t, 4, len(results))

				require.True(t, results.Contains(harness.ESC10aHarness2.Group6))
				require.True(t, results.Contains(harness.ESC10aHarness2.Group5))
				require.True(t, results.Contains(harness.ESC10aHarness2.Computer5))
				require.True(t, results.Contains(harness.ESC10aHarness2.User5))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10aHarnessECA.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC10aHarnessECA.Group1))

			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))
				edge := results[0]

				if composition, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc10a: %v", err)
				} else {
					names := []string{}
					for _, node := range composition.AllNodes() {
						name, _ := node.Properties.Get(common.Name.String()).String()
						names = append(names, name)
					}
					require.Equal(t, 8, len(composition.AllNodes()))
					require.Contains(t, names, "Group1")
					require.Contains(t, names, "Domain1")
					require.Contains(t, names, "DC1")
					require.Contains(t, names, "User1")
					require.Contains(t, names, "CertTemplate1")
					require.Contains(t, names, "EnterpriseCA1")
					require.Contains(t, names, "NTAuthStore1")
					require.Contains(t, names, "RootCA1")
				}
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10aHarnessVictim.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC10aHarnessVictim.Group1))
				require.True(t, results.Contains(harness.ESC10aHarnessVictim.Group2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10aHarnessDC1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC10aHarnessDC1.Group0))
				require.True(t, results.Contains(harness.ESC10aHarnessDC1.Group1))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10aHarnessDC2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10a")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10a)
			})); err != nil {
				t.Fatalf("error fetching esc10a edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC10aHarnessDC2.Group0))
			}
			return nil
		})
	})
}

func TestADCSESC13(t *testing.T) { //***
	t.Skip("4 Disabling test to allow engineers to continue submitting PRs and not have significant errors BED-4747")
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC13Harness1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC13")
		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain
			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC13(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC13.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}
		}

		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC13)
			})); err != nil {
				t.Fatalf("error fetching esc13 edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC13Harness1.Group1))
				require.True(t, results.Contains(harness.ESC13Harness1.Group2))

			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC13)
			})); err != nil {
				t.Fatalf("error fetching esc13 edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC13Harness1.Group6))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC13Harness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC13")
		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain
			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC13(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC13.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}
		}

		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC13)
			})); err != nil {
				t.Fatalf("error fetching esc13 edges in integration test; %v", err)
			} else {
				require.Equal(t, 7, len(results))

				require.True(t, results.Contains(harness.ESC13Harness2.Group1))
				require.True(t, results.Contains(harness.ESC13Harness2.Computer1))
				require.True(t, results.Contains(harness.ESC13Harness2.User1))
				require.True(t, results.Contains(harness.ESC13Harness2.Group2))
				require.True(t, results.Contains(harness.ESC13Harness2.Computer2))
				require.True(t, results.Contains(harness.ESC13Harness2.Group3))
				require.True(t, results.Contains(harness.ESC13Harness2.Computer3))

			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC13)
			})); err != nil {
				t.Fatalf("error fetching esc13 edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC13Harness2.Group4))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC13HarnessECA.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC13")
		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain
			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC13(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC13.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}
		}

		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC13)
			})); err != nil {
				t.Fatalf("error fetching esc13 edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC13HarnessECA.Group1))

			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC13)
			})); err != nil {
				t.Fatalf("error fetching esc13 edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC13HarnessECA.Group11))
			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC13)
			})); err != nil {
				t.Fatalf("error fetching esc13 edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))
				edge := results[0]

				if edgeComp, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc13: %v", err)
				} else {
					nodes := edgeComp.AllNodes().Slice()
					assert.Contains(t, nodes, harness.ESC13HarnessECA.Group1)
					assert.Contains(t, nodes, harness.ESC13HarnessECA.Domain1)
					assert.Contains(t, nodes, harness.ESC13HarnessECA.NTAuthStore1)
					assert.Contains(t, nodes, harness.ESC13HarnessECA.RootCA1)
					assert.Contains(t, nodes, harness.ESC13HarnessECA.EnterpriseCA1)
					assert.Contains(t, nodes, harness.ESC13HarnessECA.CertTemplate1)
					assert.Contains(t, nodes, harness.ESC13HarnessECA.Group11)
				}
			}

			return nil
		})
	})
}

func TestADCSESC10b(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10bPrincipalHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				require.Equal(t, 6, len(results))

				require.True(t, results.Contains(harness.ESC10bPrincipalHarness.Group1))
				require.True(t, results.Contains(harness.ESC10bPrincipalHarness.Group2))
				require.True(t, results.Contains(harness.ESC10bPrincipalHarness.Group3))
				require.True(t, results.Contains(harness.ESC10bPrincipalHarness.Group4))
				require.True(t, results.Contains(harness.ESC10bPrincipalHarness.Group5))
			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10bHarness1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC10bHarness1.Group1))
				require.True(t, results.Contains(harness.ESC10bHarness1.Group2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10bHarness2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC10bHarness2.Computer5))
				require.True(t, results.Contains(harness.ESC10bHarness2.User5))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10bHarnessECA.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC10bHarnessECA.Group1))

			}
			return nil
		})

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				assert.Equal(t, 1, len(results))
				edge := results[0]

				if composition, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc10b: %v", err)
				} else {
					names := []string{}
					for _, node := range composition.AllNodes() {
						name, _ := node.Properties.Get(common.Name.String()).String()
						names = append(names, name)
					}
					require.Equal(t, 8, len(composition.AllNodes()))
					require.Contains(t, names, "Group1")
					require.Contains(t, names, "Domain1")
					require.Contains(t, names, "ComputerDC1")
					require.Contains(t, names, "Computer1")
					require.Contains(t, names, "CertTemplate1")
					require.Contains(t, names, "EnterpriseCA1")
					require.Contains(t, names, "NTAuthStore1")
					require.Contains(t, names, "RootCA1")
				}
			}

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10bHarnessVictim.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC10bHarnessVictim.Group1))
				require.True(t, results.Contains(harness.ESC10bHarnessVictim.Group2))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10bHarnessDC1.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				require.Equal(t, 2, len(results))

				require.True(t, results.Contains(harness.ESC10bHarnessDC1.Group0))
				require.True(t, results.Contains(harness.ESC10bHarnessDC1.Group1))

			}
			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC10bHarnessDC2.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC10b")

		groupExpansions, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				innerEnterpriseCA := enterpriseCA

				if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
						}
						return nil
					})
				}
			}
		}
		err = operation.Done()
		require.Nil(t, err)

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC10b)
			})); err != nil {
				t.Fatalf("error fetching esc10b edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.ESC10bHarnessDC2.Group0))
			}
			return nil
		})
	})
}

func TestExtendedByPolicyBinding(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ExtendedByPolicyHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ExtendedByPolicy")

		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)

		if err := ad2.PostExtendedByPolicyBinding(operation, certTemplates); err != nil {
			t.Fatalf("failed post processing for %s: %v", ad.ExtendedByPolicy.String(), err)
		}

		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.ExtendedByPolicyHarness.CertTemplate1.ID, harness.ExtendedByPolicyHarness.IssuancePolicy0.ID, ad.ExtendedByPolicy); err != nil {
				t.Fatalf("error fetching ExtendedByPolicy edge (1) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.ExtendedByPolicyHarness.CertTemplate1.ID, harness.ExtendedByPolicyHarness.IssuancePolicy1.ID, ad.ExtendedByPolicy); err != nil {
				t.Fatalf("error fetching ExtendedByPolicy edge (2) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.ExtendedByPolicyHarness.CertTemplate2.ID, harness.ExtendedByPolicyHarness.IssuancePolicy0.ID, ad.ExtendedByPolicy); err != nil {
				t.Fatalf("error fetching ExtendedByPolicy edge (3) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.ExtendedByPolicyHarness.CertTemplate2.ID, harness.ExtendedByPolicyHarness.IssuancePolicy2.ID, ad.ExtendedByPolicy); err != nil {
				t.Fatalf("error fetching ExtendedByPolicy edge (4) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			if edge, err := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.ExtendedByPolicyHarness.CertTemplate3.ID, harness.ExtendedByPolicyHarness.IssuancePolicy3.ID, ad.ExtendedByPolicy); err != nil {
				t.Fatalf("error fetching ExtendedByPolicy edge (5) in integration test; %v", err)
			} else {
				require.NotNil(t, edge)
			}

			// CertificatePolicy doesn't match CertTemplateOID
			edge, _ := analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.ExtendedByPolicyHarness.CertTemplate2.ID, harness.ExtendedByPolicyHarness.IssuancePolicy1.ID, ad.ExtendedByPolicy)
			require.Nil(t, edge, "ExtendedByPolicy edge exists between IssuancePolicy1 and CertTemplate2 where it shouldn't")

			// Different domains, no edge
			edge, _ = analysis.FetchEdgeByStartAndEnd(testContext.Context(), db, harness.ExtendedByPolicyHarness.CertTemplate4.ID, harness.ExtendedByPolicyHarness.IssuancePolicy4.ID, ad.ExtendedByPolicy)
			require.Nil(t, edge, "ExtendedByPolicy edge bridges domains where it shouldn't")

			return nil
		})
	})
}
