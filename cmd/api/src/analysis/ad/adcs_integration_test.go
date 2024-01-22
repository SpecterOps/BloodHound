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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/graphschema"

	ad2 "github.com/specterops/bloodhound/analysis/ad"

	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"

	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {

				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC1(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
							} else {
								return nil
							}
						}
					}
				}
				return nil
			})
		}
		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC1)
			})); err != nil {
				t.Fatalf("error fetching esc1 edges in integration test; %v", err)
			} else {
				assert.Equal(t, 7, len(results))

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

			}
			return nil
		})
	})
}

func TestGoldenCert(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ADCSGoldenCertHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - Golden Cert")

		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {

				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if validPaths, err := ad2.FetchEnterpriseCAsCertChainPathToDomain(tx, enterpriseCA, innerDomain); err != nil {
							t.Logf("error fetching paths from enterprise ca %d to domain %d: %v", enterpriseCA.ID, innerDomain.ID, err)
						} else if validPaths.Len() == 0 {
							t.Logf("0 valid paths for eca: %v", enterpriseCA.ID)
							continue
						} else {
							if err := ad2.PostGoldenCert(ctx, tx, outC, innerDomain, enterpriseCA); err != nil {
								t.Logf("failed post processing for %s: %v", ad.GoldenCert.String(), err)
							} else {
								return nil
							}
						}
					}
				}
				return nil
			})
		}

		operation.Done()

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

func TestCanAbuseUPNCertMapping(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.WeakCertBindingAndUPNCertMappingHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - CanAbuseUPNCertMapping")

		if enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA); err != nil {
			t.Logf("failed fetching enterpriseCA nodes: %v", err)
		} else if err := ad2.PostCanAbuseUPNCertMapping(operation, enterpriseCertAuthorities); err != nil {
			t.Logf("failed post processing for %s: %v", ad.CanAbuseUPNCertMapping.String(), err)
		}

		// TODO: We're throwing away the collected errors from the operation and should assert on them
		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CanAbuseUPNCertMapping)
			})); err != nil {
				t.Fatalf("error fetching CanAbuseUPNCertMapping relationships; %v", err)
			} else {
				assert.True(t, len(results) == 2)

				// Positive Cases
				assert.True(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.EnterpriseCA1))
				assert.True(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.EnterpriseCA2))

				// Negative Cases
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer1))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer2))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer3))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer4))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer5))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Domain1))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Domain2))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Domain3))
			}
			return nil
		})
	})
}

func TestCanAbuseWeakCertBinding(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.WeakCertBindingAndUPNCertMappingHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - CanAbuseWeakCertBinding")

		if enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA); err != nil {
			t.Logf("failed fetching enterpriseCA nodes: %v", err)
		} else if err := ad2.PostCanAbuseWeakCertBinding(operation, enterpriseCertAuthorities); err != nil {
			t.Logf("failed post processing for %s: %v", ad.CanAbuseWeakCertBinding.String(), err)
		}

		// TODO: We're throwing away the collected errors from the operation and should assert on them
		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CanAbuseWeakCertBinding)
			})); err != nil {
				t.Fatalf("error fetching CanAbuseWeakCertBinding relationships; %v", err)
			} else {
				assert.True(t, len(results) == 1)

				// Positive Cases
				assert.True(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.EnterpriseCA1))

				// Negative Cases
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.EnterpriseCA2))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer1))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer2))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer3))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer4))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Computer5))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Domain1))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Domain2))
				assert.False(t, results.Contains(harness.WeakCertBindingAndUPNCertMappingHarness.Domain3))
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
		} else if err := ad2.PostIssuedSignedBy(operation, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
			t.Logf("failed post processing for %s: %v", ad.IssuedSignedBy.String(), err)
		}

		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results1, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.IssuedSignedBy),
					query.KindIn(query.Start(), ad.EnterpriseCA),
					query.KindIn(query.End(), ad.EnterpriseCA),
				)
			})); err != nil {
				t.Fatalf("error fetching ECA to ECA IssuedSignedBy relationships; %v", err)
			} else if results2, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.IssuedSignedBy),
					query.KindIn(query.Start(), ad.EnterpriseCA),
					query.KindIn(query.End(), ad.RootCA),
				)
			})); err != nil {
				t.Fatalf("error fetching ECA to RootCA IssuedSignedBy relationships; %v", err)
			} else if results3, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.IssuedSignedBy),
					query.KindIn(query.Start(), ad.RootCA),
					query.KindIn(query.End(), ad.RootCA),
				)
			})); err != nil {
				t.Fatalf("error fetching RootCA to RootCA IssuedSignedBy relationships; %v", err)
			} else {
				assert.True(t, len(results1) == 1)
				assert.True(t, len(results2) == 1)
				assert.True(t, len(results3) == 1)

				// Positive Cases
				assert.True(t, results3.Contains(harness.IssuedSignedByHarness.RootCA2))
				assert.True(t, results2.Contains(harness.IssuedSignedByHarness.EnterpriseCA1))
				assert.True(t, results1.Contains(harness.IssuedSignedByHarness.EnterpriseCA2))

				// Negative Cases
				assert.False(t, results1.Contains(harness.IssuedSignedByHarness.RootCA1))
				assert.False(t, results2.Contains(harness.IssuedSignedByHarness.RootCA1))
				assert.False(t, results3.Contains(harness.IssuedSignedByHarness.RootCA1))

				assert.False(t, results1.Contains(harness.IssuedSignedByHarness.EnterpriseCA3))
				assert.False(t, results2.Contains(harness.IssuedSignedByHarness.EnterpriseCA3))
				assert.False(t, results3.Contains(harness.IssuedSignedByHarness.EnterpriseCA3))
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
		harness.EnrollOnBehalfOfHarnessOne.Setup(testContext)
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
			results, err := ad2.EnrollOnBehalfOfVersionOne(tx, v1Templates, certTemplates)
			require.Nil(t, err)

			require.Len(t, results, 3)

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessOne.CertTemplate11.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessOne.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessOne.CertTemplate13.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessOne.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessOne.CertTemplate12.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessOne.CertTemplate12.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			return nil
		})
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.EnrollOnBehalfOfHarnessTwo.Setup(testContext)
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
			results, err := ad2.EnrollOnBehalfOfVersionTwo(tx, v2Templates, certTemplates)
			require.Nil(t, err)

			require.Len(t, results, 1)
			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessTwo.CertTemplate21.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessTwo.CertTemplate23.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC3(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
							} else {
								return nil
							}
						}
					}
				}
				return nil
			})
		}
		operation.Done()

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC3(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
							} else {
								return nil
							}
						}
					}
				}
				return nil
			})
		}
		operation.Done()

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
			}
			return nil
		})
	})
}
