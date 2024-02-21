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
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/graphschema"

	ad2 "github.com/specterops/bloodhound/analysis/ad"

	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"

	"testing"

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

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
		for _, template := range certTemplates {
			if version, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				continue
			} else if version == 1 {
				v1Templates = append(v1Templates, template)
			} else if version >= 2 {
				continue
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
		v2Templates := make([]*graph.Node, 0)

		for _, template := range certTemplates {
			if version, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				continue
			} else if version == 1 {
				continue
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
		require.Nil(t, err)

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
								t.Logf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
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
		require.Nil(t, err)

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
								t.Logf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
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

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC3Harness3.Setup(testContext)
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
		require.Nil(t, err)

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
								t.Logf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
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

func TestADCSESC9a(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9aPrincipalHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9a")

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9a(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
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
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.DC1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.NTAuthStore1)
					assert.Contains(t, nodes, harness.ESC9aHarnessECA.RootCA1)
				}
			}

			return nil
		})
	})

}

func TestADCSESC9b(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ESC9bPrincipalHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC9b")

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
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

		groupExpansions, _, _, domains, cache, err := FetchADCSPrereqs(db)
		require.Nil(t, err)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
							if err := ad2.PostADCSESC9b(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
								t.Logf("failed post processing for %s: %v", ad.ADCSESC9b.String(), err)
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

				if edgeComp, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc9: %v", err)
				} else {
					nodes := edgeComp.AllNodes().Slice()
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.Group1)
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.Domain1)
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.Computer1)
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.CertTemplate1)
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.EnterpriseCA1)
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.DC1)
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.NTAuthStore1)
					assert.Contains(t, nodes, harness.ESC9bHarnessECA.RootCA1)
				}
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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}
		}
		operation.Done()

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}

		}
		operation.Done()

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						}

						return nil
					})
				}
			}

		}
		operation.Done()

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
				require.Equal(t, 8, len(composition.AllNodes()))
				require.Contains(t, names, "Group1")
				require.Contains(t, names, "Group0")
				require.Contains(t, names, "CertTemplate1")
				require.Contains(t, names, "EnterpriseCA")
				require.Contains(t, names, "RootCA")
				require.Contains(t, names, "NTAuthStore")
				require.Contains(t, names, "DC")
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
				require.Equal(t, 8, len(composition.AllNodes()))
				require.Contains(t, names, "Group2")
				require.Contains(t, names, "Group0")
				require.Contains(t, names, "CertTemplate2")
				require.Contains(t, names, "EnterpriseCA")
				require.Contains(t, names, "RootCA")
				require.Contains(t, names, "NTAuthStore")
				require.Contains(t, names, "DC")
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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}

		}
		operation.Done()

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
				require.Equal(t, 14, len(results))
				require.NotContains(t, names, "User2")
				require.NotContains(t, names, "User3")
				require.NotContains(t, names, "User5")
				require.NotContains(t, names, "User7")
				require.NotContains(t, names, "Computer6")
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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}

		}
		operation.Done()

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}

		}
		operation.Done()

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}

		}
		operation.Done()

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

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)
		require.Nil(t, err)

		cache := ad2.NewADCSCache()
		cache.BuildCache(context.Background(), db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC6b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC6b.String(), err)
						} else {
							return nil
						}

						return nil
					})
				}
			}

		}
		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.ADCSESC6b)
			})); err != nil {
				t.Fatalf("error fetching esc6a edges in integration test; %v", err)
			} else {
				require.True(t, len(results) == 12)

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.User1))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Computer1))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group1))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Computer2))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group2))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Computer3))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group3))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.User4))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Computer4))
				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group4))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group5))

				require.True(t, results.Contains(harness.ESC6bTemplate2Harness.Group6))

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
		cache.BuildCache(context.Background(), db, eca, certTemplates)
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

				if edgeComp, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc10a: %v", err)
				} else {
					nodes := edgeComp.AllNodes().Slice()
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.Group1)
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.User1)
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.Domain1)
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.NTAuthStore1)
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.RootCA1)
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.DC1)
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.EnterpriseCA1)
					assert.Contains(t, nodes, harness.ESC10aHarnessECA.CertTemplate1)
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
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
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
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
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
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
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
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
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

				if edgeComp, err := ad2.GetEdgeCompositionPath(context.Background(), db, edge); err != nil {
					t.Fatalf("error getting edge composition for esc10b: %v", err)
				} else {
					nodes := edgeComp.AllNodes().Slice()
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.Group1)
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.Computer1)
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.Domain1)
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.NTAuthStore1)
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.RootCA1)
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.ComputerDC1)
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.EnterpriseCA1)
					assert.Contains(t, nodes, harness.ESC10bHarnessECA.CertTemplate1)
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
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := ad2.PostADCSESC10b(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							t.Logf("failed post processing for %s: %v", ad.ADCSESC10b.String(), err)
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
}
