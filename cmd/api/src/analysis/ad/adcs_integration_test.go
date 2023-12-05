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
	testContext := integration.NewGraphTestContext(t)

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) {
		harness.ADCSESC1Harness.Setup(testContext)
	}, func(harness integration.HarnessDetails, db graph.Database) error {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "ADCS Post Process Test - ESC1")

		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		enterpriseCertAuthorities, err := ad2.FetchNodesByKind(context.Background(), db, ad.EnterpriseCA)
		require.Nil(t, err)
		certTemplates, err := ad2.FetchNodesByKind(context.Background(), db, ad.CertTemplate)
		require.Nil(t, err)
		domains, err := ad2.FetchNodesByKind(context.Background(), db, ad.Domain)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {

				if enterpriseCAs, err := ad2.FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, enterpriseCA := range enterpriseCAs {
						t.Logf("here is one of the looped over ecas: %v", enterpriseCA.ID)
						if validPaths, err := ad2.FetchEnterpriseCAsCertChainPathToDomain(tx, enterpriseCA, innerDomain); err != nil {
							t.Logf("error fetching paths from enterprise ca %d to domain %d: %v", enterpriseCA.ID, innerDomain.ID, err)
						} else if validPaths.Len() == 0 {
							t.Logf("0 valid paths for eca: %v", enterpriseCA.Properties.Get("name"))
							continue
						} else {
							if err := ad2.PostADCSESC1(ctx, tx, outC, db, groupExpansions, enterpriseCertAuthorities, certTemplates, enterpriseCA, innerDomain); err != nil {
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
		return nil
	})

}

func TestGoldenCert(t *testing.T) {
	testContext := integration.NewGraphTestContext(t)

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) {
		harness.ADCSGoldenCertHarness.Setup(testContext)
	}, func(harness integration.HarnessDetails, db graph.Database) error {
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
		return nil
	})

}

func TestTrustedForNTAuth(t *testing.T) {
	testContext := integration.NewGraphTestContext(t)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) {
			harness.TrustedForNTAuthHarness.Setup(testContext)
		},
		func(harness integration.HarnessDetails, db graph.Database) error {
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

			return nil
		})
}

func TestEnrollOnBehalfOf(t *testing.T) {
	testContext := integration.NewGraphTestContext(t)
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) {
		harness.EnrollOnBehalfOfHarnessOne.Setup(testContext)
	}, func(harness integration.HarnessDetails, db graph.Database) error {
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

			require.Len(t, results, 2)

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

			return nil
		})

		return nil
	})

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) {
		harness.EnrollOnBehalfOfHarnessTwo.Setup(testContext)
	}, func(harness integration.HarnessDetails, db graph.Database) error {
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

			results, err = ad2.EnrollOnBehalfOfSelfControl(tx, v1Templates)
			require.Nil(t, err)

			require.Len(t, results, 1)
			require.Contains(t, results, analysis.CreatePostRelationshipJob{
				FromID: harness.EnrollOnBehalfOfHarnessTwo.CertTemplate25.ID,
				ToID:   harness.EnrollOnBehalfOfHarnessTwo.CertTemplate25.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})

			return nil
		})

		return nil
	})
}
