// Copyright 2026 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addEnabledHostingComputer creates an enabled computer in the given domain,
// links it to the CA via HostsCAService, and returns the computer for tests that
// need to model one host serving multiple EnterpriseCA nodes.
func addEnabledHostingComputer(testCtx *integration.GraphTestContext, name, domainSID string, enterpriseCA *graph.Node) *graph.Node {
	computer := testCtx.NewActiveDirectoryComputer(name, domainSID)
	computer.Properties.Set(common.Enabled.String(), true)
	testCtx.UpdateNode(computer)
	testCtx.NewRelationship(computer, enterpriseCA, ad.HostsCAService)
	return computer
}

// linkEnterpriseCAToDomain adds the per-domain edges that make a domain "chain
// valid" (RootCAFor ∩ TrustedForNTAuth) for the CA. The per-CA EnterpriseCAFor
// edge is created once by the caller; each domain gets its own NTAuthStore so the
// TrustedForNTAuth edges don't collide.
func linkEnterpriseCAToDomain(testCtx *integration.GraphTestContext, enterpriseCA, rootCA *graph.Node, domain *graph.Node, domainSID string) {
	ntAuthStore := testCtx.NewActiveDirectoryNTAuthStore("NTAuthStore-"+domainSID, domainSID)

	// RootCAFor path: domain <-RootCAFor- rootCA <-EnterpriseCAFor- enterpriseCA
	testCtx.NewRelationship(rootCA, domain, ad.RootCAFor)

	// TrustedForNTAuth path: domain <-NTAuthStoreFor- ntAuthStore <-TrustedForNTAuth- enterpriseCA
	testCtx.NewRelationship(ntAuthStore, domain, ad.NTAuthStoreFor)
	testCtx.NewRelationship(enterpriseCA, ntAuthStore, ad.TrustedForNTAuth)
}

func TestADCSPKIHierarchyRequiresRootedCAChain(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainSID = integration.RandomDomainSID()

		validEnterpriseCAID               graph.ID
		directRootCAForEnterpriseCAID     graph.ID
		invalidIntermediateEnterpriseCAID graph.ID
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domain := testContext.NewActiveDirectoryDomain("Domain", domainSID, false, true)
			rootCA := testContext.NewActiveDirectoryRootCA("RootCA", domainSID)
			ntAuthStore := testContext.NewActiveDirectoryNTAuthStore("NTAuthStore", domainSID)
			intermediateEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("IntermediateEnterpriseCA", domainSID)
			intermediateAIACA := testContext.NewActiveDirectoryAIACA("IntermediateAIACA", domainSID, "intermediate", []string{"intermediate"})
			validEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("ValidEnterpriseCA", domainSID)
			directRootCAForEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("DirectRootCAForEnterpriseCA", domainSID)
			invalidIntermediateEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("InvalidIntermediateEnterpriseCA", domainSID)
			invalidIntermediate := testContext.NewActiveDirectoryUser("InvalidIntermediate", domainSID)

			testContext.NewRelationship(rootCA, domain, ad.RootCAFor)
			testContext.NewRelationship(ntAuthStore, domain, ad.NTAuthStoreFor)

			testContext.NewRelationship(validEnterpriseCA, intermediateAIACA, ad.IssuedSignedBy)
			testContext.NewRelationship(intermediateAIACA, intermediateEnterpriseCA, ad.EnterpriseCAFor)
			testContext.NewRelationship(intermediateEnterpriseCA, rootCA, ad.IssuedSignedBy)
			testContext.NewRelationship(validEnterpriseCA, ntAuthStore, ad.TrustedForNTAuth)

			testContext.NewRelationship(directRootCAForEnterpriseCA, domain, ad.RootCAFor)
			testContext.NewRelationship(directRootCAForEnterpriseCA, ntAuthStore, ad.TrustedForNTAuth)

			testContext.NewRelationship(invalidIntermediateEnterpriseCA, invalidIntermediate, ad.IssuedSignedBy)
			testContext.NewRelationship(invalidIntermediate, rootCA, ad.EnterpriseCAFor)
			testContext.NewRelationship(invalidIntermediateEnterpriseCA, ntAuthStore, ad.TrustedForNTAuth)

			addEnabledHostingComputer(testContext, "ValidHost", domainSID, validEnterpriseCA)
			addEnabledHostingComputer(testContext, "DirectRootCAForHost", domainSID, directRootCAForEnterpriseCA)
			addEnabledHostingComputer(testContext, "InvalidIntermediateHost", domainSID, invalidIntermediateEnterpriseCA)

			validEnterpriseCAID = validEnterpriseCA.ID
			directRootCAForEnterpriseCAID = directRootCAForEnterpriseCA.ID
			invalidIntermediateEnterpriseCAID = invalidIntermediateEnterpriseCA.ID
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			chainedDomains := cache.GetECAHostedChainedDomains()
			require.Contains(t, chainedDomains, validEnterpriseCAID.Uint64())
			assert.NotContains(t, chainedDomains, directRootCAForEnterpriseCAID.Uint64())
			assert.NotContains(t, chainedDomains, invalidIntermediateEnterpriseCAID.Uint64())
		},
	)
}

func TestADCSNTAuthStoreChainRequiresExactPattern(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainSID = integration.RandomDomainSID()

		domain                             *graph.Node
		validEnterpriseCA                  *graph.Node
		directEnterpriseCA                 *graph.Node
		invalidIntermediateEnterpriseCA    *graph.Node
		additionalIntermediateEnterpriseCA *graph.Node
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domain = testContext.NewActiveDirectoryDomain("Domain", domainSID, false, true)
			rootCA := testContext.NewActiveDirectoryRootCA("RootCA", domainSID)
			validNTAuthStore := testContext.NewActiveDirectoryNTAuthStore("ValidNTAuthStore", domainSID)
			invalidIntermediate := testContext.NewActiveDirectoryUser("InvalidIntermediate", domainSID)
			firstNTAuthStore := testContext.NewActiveDirectoryNTAuthStore("FirstNTAuthStore", domainSID)
			secondNTAuthStore := testContext.NewActiveDirectoryNTAuthStore("SecondNTAuthStore", domainSID)

			validEnterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("ValidEnterpriseCA", domainSID)
			directEnterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("DirectEnterpriseCA", domainSID)
			invalidIntermediateEnterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("InvalidIntermediateEnterpriseCA", domainSID)
			additionalIntermediateEnterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("AdditionalIntermediateEnterpriseCA", domainSID)

			testContext.NewRelationship(rootCA, domain, ad.RootCAFor)
			for _, enterpriseCA := range []*graph.Node{
				validEnterpriseCA,
				directEnterpriseCA,
				invalidIntermediateEnterpriseCA,
				additionalIntermediateEnterpriseCA,
			} {
				testContext.NewRelationship(enterpriseCA, rootCA, ad.IssuedSignedBy)
			}
			addEnabledHostingComputer(testContext, "ValidHost", domainSID, validEnterpriseCA)
			addEnabledHostingComputer(testContext, "DirectHost", domainSID, directEnterpriseCA)
			addEnabledHostingComputer(testContext, "InvalidIntermediateHost", domainSID, invalidIntermediateEnterpriseCA)
			addEnabledHostingComputer(testContext, "AdditionalIntermediateHost", domainSID, additionalIntermediateEnterpriseCA)

			testContext.NewRelationship(validEnterpriseCA, validNTAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(validNTAuthStore, domain, ad.NTAuthStoreFor)

			testContext.NewRelationship(directEnterpriseCA, domain, ad.NTAuthStoreFor)

			testContext.NewRelationship(invalidIntermediateEnterpriseCA, invalidIntermediate, ad.TrustedForNTAuth)
			testContext.NewRelationship(invalidIntermediate, domain, ad.NTAuthStoreFor)

			testContext.NewRelationship(additionalIntermediateEnterpriseCA, firstNTAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(firstNTAuthStore, secondNTAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(secondNTAuthStore, domain, ad.NTAuthStoreFor)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			chainedDomains := cache.GetECAHostedChainedDomains()
			require.Contains(t, chainedDomains, validEnterpriseCA.ID.Uint64())
			assert.NotContains(t, chainedDomains, directEnterpriseCA.ID.Uint64())
			assert.NotContains(t, chainedDomains, invalidIntermediateEnterpriseCA.ID.Uint64())
			assert.NotContains(t, chainedDomains, additionalIntermediateEnterpriseCA.ID.Uint64())

			require.NoError(t, db.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
				paths, err := adAnalysis.FetchEnterpriseCAsTrustedForAuthPathToDomain(tx, validEnterpriseCA, domain)
				require.NoError(t, err)
				require.Len(t, paths, 1)

				for _, enterpriseCA := range []*graph.Node{
					directEnterpriseCA,
					invalidIntermediateEnterpriseCA,
					additionalIntermediateEnterpriseCA,
				} {
					paths, err := adAnalysis.FetchEnterpriseCAsTrustedForAuthPathToDomain(tx, enterpriseCA, domain)
					require.NoError(t, err)
					assert.Empty(t, paths)
				}
				return nil
			}))
		},
	)
}

func TestADCSESC3SelectsPublishersForTargetDomain(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainASID = integration.RandomDomainSID()
		domainBSID = integration.RandomDomainSID()

		domainA              *graph.Node
		domainB              *graph.Node
		domainAOnlyPrincipal *graph.Node
		bothDomainsPrincipal *graph.Node
		domainAPublisher     *graph.Node
		domainBPublisher     *graph.Node
		targetEnterpriseCA   *graph.Node
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domainA = testContext.NewActiveDirectoryDomain("DomainA", domainASID, false, true)
			domainB = testContext.NewActiveDirectoryDomain("DomainB", domainBSID, false, true)
			domainARootCA := testContext.NewActiveDirectoryRootCA("DomainARootCA", domainASID)
			domainBRootCA := testContext.NewActiveDirectoryRootCA("DomainBRootCA", domainBSID)
			domainANTAuthStore := testContext.NewActiveDirectoryNTAuthStore("DomainANTAuthStore", domainASID)
			domainBNTAuthStore := testContext.NewActiveDirectoryNTAuthStore("DomainBNTAuthStore", domainBSID)
			domainAPublisher = testContext.NewActiveDirectoryEnterpriseCA("DomainAPublisher", domainASID)
			domainBPublisher = testContext.NewActiveDirectoryEnterpriseCA("DomainBPublisher", domainBSID)
			targetEnterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("TargetEnterpriseCA", domainASID)

			enrollmentAgentTemplate := testContext.NewActiveDirectoryCertTemplate("EnrollmentAgentTemplate", domainASID, integration.CertTemplateData{
				RequiresManagerApproval: false,
				SchemaVersion:           2,
				AuthorizedSignatures:    0,
				EffectiveEKUs:           []string{adAnalysis.EkuCertRequestAgent},
				ApplicationPolicies:     []string{},
			})
			targetTemplate := testContext.NewActiveDirectoryCertTemplate("TargetTemplate", domainASID, integration.CertTemplateData{
				RequiresManagerApproval: false,
				AuthenticationEnabled:   true,
				SchemaVersion:           2,
				AuthorizedSignatures:    1,
				EffectiveEKUs:           []string{"1.3.6.1.5.5.7.3.2"},
				ApplicationPolicies:     []string{adAnalysis.EkuCertRequestAgent},
			})

			domainAOnlyPrincipal = testContext.NewActiveDirectoryGroup("DomainAOnlyPrincipal", domainASID)
			bothDomainsPrincipal = testContext.NewActiveDirectoryGroup("BothDomainsPrincipal", domainASID)

			testContext.NewRelationship(domainARootCA, domainA, ad.RootCAFor)
			testContext.NewRelationship(domainBRootCA, domainB, ad.RootCAFor)
			testContext.NewRelationship(domainANTAuthStore, domainA, ad.NTAuthStoreFor)
			testContext.NewRelationship(domainBNTAuthStore, domainB, ad.NTAuthStoreFor)

			// The target CA is valid for both domains. Each CT1 publisher is valid
			// for exactly one domain through both RootCA and NTAuth paths.
			testContext.NewRelationship(targetEnterpriseCA, domainARootCA, ad.IssuedSignedBy)
			testContext.NewRelationship(targetEnterpriseCA, domainBRootCA, ad.IssuedSignedBy)
			testContext.NewRelationship(targetEnterpriseCA, domainANTAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(targetEnterpriseCA, domainBNTAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(domainAPublisher, domainARootCA, ad.IssuedSignedBy)
			testContext.NewRelationship(domainAPublisher, domainANTAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(domainBPublisher, domainBRootCA, ad.IssuedSignedBy)
			testContext.NewRelationship(domainBPublisher, domainBNTAuthStore, ad.TrustedForNTAuth)

			testContext.NewRelationship(enrollmentAgentTemplate, domainAPublisher, ad.PublishedTo)
			testContext.NewRelationship(enrollmentAgentTemplate, domainBPublisher, ad.PublishedTo)
			testContext.NewRelationship(targetTemplate, targetEnterpriseCA, ad.PublishedTo)

			for _, enrollmentTarget := range []*graph.Node{
				enrollmentAgentTemplate,
				targetTemplate,
				domainAPublisher,
				targetEnterpriseCA,
			} {
				testContext.NewRelationship(domainAOnlyPrincipal, enrollmentTarget, ad.Enroll)
				testContext.NewRelationship(bothDomainsPrincipal, enrollmentTarget, ad.Enroll)
			}
			testContext.NewRelationship(bothDomainsPrincipal, domainBPublisher, ad.Enroll)

			addEnabledHostingComputer(testContext, "DomainAPublisherHost", domainASID, domainAPublisher)
			addEnabledHostingComputer(testContext, "DomainBPublisherHost", domainBSID, domainBPublisher)
			addEnabledHostingComputer(testContext, "TargetEnterpriseCAHost", domainASID, targetEnterpriseCA)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			localGroupData, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			enrollOnBehalfOfOperation := post.NewPostRelationshipOperation(t.Context(), db, "ADCS ESC3 publisher provenance - EnrollOnBehalfOf")
			require.NoError(t, adAnalysis.PostEnrollOnBehalfOf(cache, enrollOnBehalfOfOperation))
			require.NoError(t, enrollOnBehalfOfOperation.Done())

			edgeOperation := post.NewPostRelationshipOperation(t.Context(), db, "ADCS ESC3 publisher provenance")
			for _, certificateChains := range cache.GetECAHostedChainedDomains() {
				certificateChains := certificateChains
				require.NoError(t, edgeOperation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
					return adAnalysis.PostADCSESC3(ctx, tx, outC, localGroupData, certificateChains, cache)
				}))
			}
			require.NoError(t, edgeOperation.Done())

			require.NoError(t, db.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
				fetchEdge := func(principal, domain *graph.Node) (*graph.Relationship, error) {
					return tx.Relationships().Filterf(func() graph.Criteria {
						return query.And(
							query.Kind(query.Relationship(), ad.ADCSESC3),
							query.Equals(query.StartID(), principal.ID),
							query.Equals(query.EndID(), domain.ID),
						)
					}).First()
				}

				domainAEdge, err := fetchEdge(domainAOnlyPrincipal, domainA)
				require.NoError(t, err)
				_, err = fetchEdge(domainAOnlyPrincipal, domainB)
				require.True(t, graph.IsErrNotFound(err), "a CT1 publisher for domain A must not create an edge to domain B")

				domainAComposition, err := adAnalysis.GetADCSESC3EdgeComposition(t.Context(), db, domainAEdge)
				require.NoError(t, err)
				require.NotEmpty(t, domainAComposition)
				require.True(t, domainAComposition.AllNodes().Contains(domainAPublisher))
				require.False(t, domainAComposition.AllNodes().Contains(domainBPublisher))

				invalidDomainBComposition, err := adAnalysis.GetADCSESC3EdgeComposition(t.Context(), db, graph.NewRelationship(
					0,
					domainAOnlyPrincipal.ID,
					domainB.ID,
					graph.NewProperties(),
					ad.ADCSESC3,
				))
				require.NoError(t, err)
				require.Empty(t, invalidDomainBComposition)

				for _, testCase := range []struct {
					domain            *graph.Node
					expectedPublisher *graph.Node
					excludedPublisher *graph.Node
				}{
					{domain: domainA, expectedPublisher: domainAPublisher, excludedPublisher: domainBPublisher},
					{domain: domainB, expectedPublisher: domainBPublisher, excludedPublisher: domainAPublisher},
				} {
					edge, err := fetchEdge(bothDomainsPrincipal, testCase.domain)
					require.NoError(t, err)

					composition, err := adAnalysis.GetADCSESC3EdgeComposition(t.Context(), db, edge)
					require.NoError(t, err)
					require.NotEmpty(t, composition)
					require.True(t, composition.AllNodes().Contains(testCase.expectedPublisher))
					require.False(t, composition.AllNodes().Contains(testCase.excludedPublisher))
				}
				return nil
			}))
		},
	)
}

func TestADCSESC1CompositionScopesHostsToExactEnterpriseCA(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainSID                = integration.RandomDomainSID()
		foreignDomainSID         = integration.RandomDomainSID()
		unknownDomainSID         = integration.RandomDomainSID()
		principal                *graph.Node
		domain                   *graph.Node
		validEnterpriseCA        *graph.Node
		unhostedCA               *graph.Node
		disabledHostedCA         *graph.Node
		crossForestCA            *graph.Node
		unknownHostDomainCA      *graph.Node
		validHost                *graph.Node
		disabledValidCAHost      *graph.Node
		crossForestValidCAHost   *graph.Node
		unknownDomainValidCAHost *graph.Node
		disabledHost             *graph.Node
		crossForestHost          *graph.Node
		unknownDomainHost        *graph.Node
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domain = testContext.NewActiveDirectoryDomain("Domain", domainSID, false, true)
			testContext.NewActiveDirectoryDomain("ForeignDomain", foreignDomainSID, false, true)
			rootCA := testContext.NewActiveDirectoryRootCA("RootCA", domainSID)
			ntAuthStore := testContext.NewActiveDirectoryNTAuthStore("NTAuthStore", domainSID)
			certificateTemplate := testContext.NewActiveDirectoryCertTemplate("CertificateTemplate", domainSID, integration.CertTemplateData{
				AuthenticationEnabled:   true,
				AuthorizedSignatures:    0,
				EnrolleeSuppliesSubject: true,
				RequiresManagerApproval: false,
				SchemaVersion:           1,
			})
			principal = testContext.NewActiveDirectoryGroup("Principal", domainSID)

			validEnterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("ValidEnterpriseCA", domainSID)
			unhostedCA = testContext.NewActiveDirectoryEnterpriseCA("UnhostedCA", domainSID)
			disabledHostedCA = testContext.NewActiveDirectoryEnterpriseCA("DisabledHostedCA", domainSID)
			crossForestCA = testContext.NewActiveDirectoryEnterpriseCA("CrossForestCA", domainSID)
			unknownHostDomainCA = testContext.NewActiveDirectoryEnterpriseCA("UnknownHostDomainCA", domainSID)

			testContext.NewRelationship(rootCA, domain, ad.RootCAFor)
			testContext.NewRelationship(ntAuthStore, domain, ad.NTAuthStoreFor)
			testContext.NewRelationship(principal, certificateTemplate, ad.Enroll)

			for _, enterpriseCA := range []*graph.Node{
				validEnterpriseCA,
				unhostedCA,
				disabledHostedCA,
				crossForestCA,
				unknownHostDomainCA,
			} {
				testContext.NewRelationship(enterpriseCA, rootCA, ad.IssuedSignedBy)
				testContext.NewRelationship(enterpriseCA, ntAuthStore, ad.TrustedForNTAuth)
				testContext.NewRelationship(certificateTemplate, enterpriseCA, ad.PublishedTo)
				testContext.NewRelationship(principal, enterpriseCA, ad.Enroll)
			}

			validHost = addEnabledHostingComputer(testContext, "ValidHost", domainSID, validEnterpriseCA)
			disabledValidCAHost = testContext.NewActiveDirectoryComputer("DisabledValidCAHost", domainSID)
			disabledValidCAHost.Properties.Set(common.Enabled.String(), false)
			testContext.UpdateNode(disabledValidCAHost)
			testContext.NewRelationship(disabledValidCAHost, validEnterpriseCA, ad.HostsCAService)
			crossForestValidCAHost = addEnabledHostingComputer(testContext, "CrossForestValidCAHost", foreignDomainSID, validEnterpriseCA)
			unknownDomainValidCAHost = addEnabledHostingComputer(testContext, "UnknownDomainValidCAHost", unknownDomainSID, validEnterpriseCA)

			disabledHost = testContext.NewActiveDirectoryComputer("DisabledHost", domainSID)
			disabledHost.Properties.Set(common.Enabled.String(), false)
			testContext.UpdateNode(disabledHost)
			testContext.NewRelationship(disabledHost, disabledHostedCA, ad.HostsCAService)
			crossForestHost = addEnabledHostingComputer(testContext, "CrossForestHost", foreignDomainSID, crossForestCA)
			unknownDomainHost = addEnabledHostingComputer(testContext, "UnknownDomainHost", unknownDomainSID, unknownHostDomainCA)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			localGroupData, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			chainedDomains := cache.GetECAHostedChainedDomains()
			require.Contains(t, chainedDomains, validEnterpriseCA.ID.Uint64())
			assert.NotContains(t, chainedDomains, unhostedCA.ID.Uint64())
			assert.NotContains(t, chainedDomains, disabledHostedCA.ID.Uint64())
			assert.NotContains(t, chainedDomains, crossForestCA.ID.Uint64())
			assert.NotContains(t, chainedDomains, unknownHostDomainCA.ID.Uint64())

			edgeOperation := post.NewPostRelationshipOperation(t.Context(), db, "ADCS ESC1 exact host CA scoping")
			for _, certificateChains := range chainedDomains {
				certificateChains := certificateChains
				require.NoError(t, edgeOperation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
					return adAnalysis.PostADCSESC1(ctx, tx, outC, localGroupData, certificateChains, cache)
				}))
			}
			require.NoError(t, edgeOperation.Done())

			var edge *graph.Relationship
			require.NoError(t, db.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
				edge, err = tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC1),
						query.Equals(query.StartID(), principal.ID),
						query.Equals(query.EndID(), domain.ID),
					)
				}).First()
				return err
			}))

			composition, err := adAnalysis.GetADCSESC1EdgeComposition(t.Context(), db, edge)
			require.NoError(t, err)
			require.NotEmpty(t, composition)
			require.True(t, composition.AllNodes().Contains(validEnterpriseCA))
			for _, invalidNode := range []*graph.Node{
				unhostedCA,
				disabledHostedCA,
				crossForestCA,
				unknownHostDomainCA,
				disabledValidCAHost,
				crossForestValidCAHost,
				unknownDomainValidCAHost,
				disabledHost,
				crossForestHost,
				unknownDomainHost,
			} {
				require.False(t, composition.AllNodes().Contains(invalidNode))
			}

			var validHostPathFound bool
			for _, path := range composition.Paths() {
				path.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
					if relationship.Kind.Is(ad.HostsCAService) {
						require.Equal(t, validHost.ID, start.ID)
						require.Equal(t, validEnterpriseCA.ID, end.ID)
						validHostPathFound = true
					}
					return true
				})
			}
			require.True(t, validHostPathFound)
		},
	)
}

func TestEnrollOnBehalfOfRequiresSharedHostedCAChains(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainSID        = integration.RandomDomainSID()
		otherDomainSID   = integration.RandomDomainSID()
		validAgent       *graph.Node
		noNTAuthAgent    *graph.Node
		noHostAgent      *graph.Node
		otherDomainAgent *graph.Node
		target           *graph.Node
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domain := testContext.NewActiveDirectoryDomain("Domain", domainSID, false, true)
			otherDomain := testContext.NewActiveDirectoryDomain("OtherDomain", otherDomainSID, false, true)
			rootCA := testContext.NewActiveDirectoryRootCA("RootCA", domainSID)
			otherRootCA := testContext.NewActiveDirectoryRootCA("OtherRootCA", otherDomainSID)
			ntAuthStore := testContext.NewActiveDirectoryNTAuthStore("NTAuthStore", domainSID)
			otherNTAuthStore := testContext.NewActiveDirectoryNTAuthStore("OtherNTAuthStore", otherDomainSID)
			validEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("ValidEnterpriseCA", domainSID)
			noNTAuthEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("NoNTAuthEnterpriseCA", domainSID)
			noHostEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("NoHostEnterpriseCA", domainSID)
			otherDomainEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("OtherDomainEnterpriseCA", otherDomainSID)

			testContext.NewRelationship(rootCA, domain, ad.RootCAFor)
			testContext.NewRelationship(otherRootCA, otherDomain, ad.RootCAFor)
			testContext.NewRelationship(ntAuthStore, domain, ad.NTAuthStoreFor)
			testContext.NewRelationship(otherNTAuthStore, otherDomain, ad.NTAuthStoreFor)

			for _, enterpriseCA := range []*graph.Node{validEnterpriseCA, noNTAuthEnterpriseCA, noHostEnterpriseCA} {
				testContext.NewRelationship(enterpriseCA, rootCA, ad.EnterpriseCAFor)
			}
			testContext.NewRelationship(otherDomainEnterpriseCA, otherRootCA, ad.EnterpriseCAFor)
			testContext.NewRelationship(validEnterpriseCA, ntAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(noHostEnterpriseCA, ntAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(otherDomainEnterpriseCA, otherNTAuthStore, ad.TrustedForNTAuth)
			addEnabledHostingComputer(testContext, "ValidHost", domainSID, validEnterpriseCA)
			addEnabledHostingComputer(testContext, "NoNTAuthHost", domainSID, noNTAuthEnterpriseCA)
			addEnabledHostingComputer(testContext, "OtherDomainHost", otherDomainSID, otherDomainEnterpriseCA)

			newAgentTemplate := func(name string, templateDomainSID string) *graph.Node {
				return testContext.NewActiveDirectoryCertTemplate(name, templateDomainSID, integration.CertTemplateData{
					RequiresManagerApproval: false,
					SchemaVersion:           2,
					AuthorizedSignatures:    0,
					EffectiveEKUs:           []string{adAnalysis.EkuCertRequestAgent},
				})
			}

			validAgent = newAgentTemplate("ValidAgent", domainSID)
			noNTAuthAgent = newAgentTemplate("NoNTAuthAgent", domainSID)
			noHostAgent = newAgentTemplate("NoHostAgent", domainSID)
			otherDomainAgent = newAgentTemplate("OtherDomainAgent", otherDomainSID)
			target = testContext.NewActiveDirectoryCertTemplate("Target", domainSID, integration.CertTemplateData{
				RequiresManagerApproval: false,
				SchemaVersion:           2,
				AuthorizedSignatures:    1,
				ApplicationPolicies:     []string{adAnalysis.EkuCertRequestAgent},
			})

			testContext.NewRelationship(validAgent, validEnterpriseCA, ad.PublishedTo)
			testContext.NewRelationship(target, validEnterpriseCA, ad.PublishedTo)
			testContext.NewRelationship(noNTAuthAgent, noNTAuthEnterpriseCA, ad.PublishedTo)
			testContext.NewRelationship(noHostAgent, noHostEnterpriseCA, ad.PublishedTo)
			testContext.NewRelationship(otherDomainAgent, otherDomainEnterpriseCA, ad.PublishedTo)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			operation := post.NewPostRelationshipOperation(t.Context(), db, "EnrollOnBehalfOf shared hosted CA chains")
			require.NoError(t, adAnalysis.PostEnrollOnBehalfOf(cache, operation))
			require.NoError(t, operation.Done())

			require.NoError(t, db.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
				for _, agentTemplate := range []*graph.Node{validAgent, noNTAuthAgent, noHostAgent, otherDomainAgent} {
					edge, edgeErr := tx.Relationships().Filterf(func() graph.Criteria {
						return query.And(
							query.Kind(query.Relationship(), ad.EnrollOnBehalfOf),
							query.Equals(query.StartID(), agentTemplate.ID),
							query.Equals(query.EndID(), target.ID),
						)
					}).First()

					if agentTemplate.ID == validAgent.ID {
						require.NoError(t, edgeErr)
						require.Equal(t, ad.EnrollOnBehalfOf, edge.Kind)
					} else {
						require.True(t, graph.IsErrNotFound(edgeErr))
					}
				}

				return nil
			}))
		},
	)
}

func TestGoldenCertCreationAndCompositionUseOnlyQualifyingHosts(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainSID         = integration.RandomDomainSID()
		foreignDomainSID  = integration.RandomDomainSID()
		unknownDomainSID  = integration.RandomDomainSID()
		domain            *graph.Node
		enterpriseCA      *graph.Node
		validHost         *graph.Node
		disabledHost      *graph.Node
		crossForestHost   *graph.Node
		unknownDomainHost *graph.Node
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domain = testContext.NewActiveDirectoryDomain("Domain", domainSID, false, true)
			testContext.NewActiveDirectoryDomain("ForeignDomain", foreignDomainSID, false, true)
			rootCA := testContext.NewActiveDirectoryRootCA("RootCA", domainSID)
			enterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("EnterpriseCA", domainSID)

			testContext.NewRelationship(enterpriseCA, rootCA, ad.IssuedSignedBy)
			linkEnterpriseCAToDomain(testContext, enterpriseCA, rootCA, domain, domainSID)

			validHost = addEnabledHostingComputer(testContext, "ValidHost", domainSID, enterpriseCA)
			disabledHost = testContext.NewActiveDirectoryComputer("DisabledHost", domainSID)
			disabledHost.Properties.Set(common.Enabled.String(), false)
			testContext.UpdateNode(disabledHost)
			testContext.NewRelationship(disabledHost, enterpriseCA, ad.HostsCAService)
			crossForestHost = addEnabledHostingComputer(testContext, "CrossForestHost", foreignDomainSID, enterpriseCA)
			unknownDomainHost = addEnabledHostingComputer(testContext, "UnknownDomainHost", unknownDomainSID, enterpriseCA)

			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			chainedDomains := cache.GetECAHostedChainedDomains()
			require.Contains(t, chainedDomains, enterpriseCA.ID.Uint64())

			edgeOperation := post.NewPostRelationshipOperation(t.Context(), db, "GoldenCert exact host scoping")
			require.NoError(t, edgeOperation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				return adAnalysis.PostGoldenCert(ctx, tx, outC, chainedDomains[enterpriseCA.ID.Uint64()])
			}))
			require.NoError(t, edgeOperation.Done())

			var goldenCertEdge *graph.Relationship
			require.NoError(t, db.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
				edges, err := ops.FetchRelationships(tx.Relationships().Filter(query.And(
					query.Kind(query.Relationship(), ad.GoldenCert),
					query.Equals(query.EndID(), domain.ID),
				)))
				if err != nil {
					return err
				}

				require.Len(t, edges, 1)
				goldenCertEdge = edges[0]
				assert.Equal(t, validHost.ID, goldenCertEdge.StartID)
				return nil
			}))

			composition, err := adAnalysis.GetEdgeCompositionPath(t.Context(), db, goldenCertEdge)
			require.NoError(t, err)
			require.NotEmpty(t, composition)
			require.True(t, composition.AllNodes().Contains(validHost))
			for _, invalidHost := range []*graph.Node{disabledHost, crossForestHost, unknownDomainHost} {
				require.False(t, composition.AllNodes().Contains(invalidHost))
			}

			var qualifyingHostPathFound bool
			for _, path := range composition.Paths() {
				path.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
					if relationship.Kind.Is(ad.HostsCAService) {
						require.Equal(t, validHost.ID, start.ID)
						require.Equal(t, enterpriseCA.ID, end.ID)
						qualifyingHostPathFound = true
					}
					return true
				})
			}
			require.True(t, qualifyingHostPathFound)
		},
	)
}

func TestADCSHostEligibilityFallsBackWhenEnterpriseCAForestIsUnresolved(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainSID             = integration.RandomDomainSID()
		unresolvedCADomainSID = integration.RandomDomainSID()
		principal             *graph.Node
		domain                *graph.Node
		enterpriseCA          *graph.Node
		host                  *graph.Node
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domain = testContext.NewActiveDirectoryDomain("Domain", domainSID, false, true)
			rootCA := testContext.NewActiveDirectoryRootCA("RootCA", domainSID)
			ntAuthStore := testContext.NewActiveDirectoryNTAuthStore("NTAuthStore", domainSID)
			certificateTemplate := testContext.NewActiveDirectoryCertTemplate("CertificateTemplate", domainSID, integration.CertTemplateData{
				AuthenticationEnabled:   true,
				AuthorizedSignatures:    0,
				EnrolleeSuppliesSubject: true,
				RequiresManagerApproval: false,
				SchemaVersion:           1,
			})
			principal = testContext.NewActiveDirectoryGroup("Principal", domainSID)
			enterpriseCA = testContext.NewActiveDirectoryEnterpriseCA("EnterpriseCA", unresolvedCADomainSID)

			testContext.NewRelationship(rootCA, domain, ad.RootCAFor)
			testContext.NewRelationship(ntAuthStore, domain, ad.NTAuthStoreFor)
			testContext.NewRelationship(enterpriseCA, rootCA, ad.IssuedSignedBy)
			testContext.NewRelationship(enterpriseCA, ntAuthStore, ad.TrustedForNTAuth)
			testContext.NewRelationship(certificateTemplate, enterpriseCA, ad.PublishedTo)
			testContext.NewRelationship(principal, certificateTemplate, ad.Enroll)
			testContext.NewRelationship(principal, enterpriseCA, ad.Enroll)
			host = addEnabledHostingComputer(testContext, "Host", domainSID, enterpriseCA)
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			localGroupData, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			chainedDomains := cache.GetECAHostedChainedDomains()
			require.Contains(t, chainedDomains, enterpriseCA.ID.Uint64())

			edgeOperation := post.NewPostRelationshipOperation(t.Context(), db, "ADCS unresolved CA forest fallback")
			require.NoError(t, edgeOperation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				return adAnalysis.PostADCSESC1(ctx, tx, outC, localGroupData, chainedDomains[enterpriseCA.ID.Uint64()], cache)
			}))
			require.NoError(t, edgeOperation.Done())

			var edge *graph.Relationship
			require.NoError(t, db.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
				edge, err = tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Relationship(), ad.ADCSESC1),
						query.Equals(query.StartID(), principal.ID),
						query.Equals(query.EndID(), domain.ID),
					)
				}).First()
				return err
			}))

			composition, err := adAnalysis.GetADCSESC1EdgeComposition(t.Context(), db, edge)
			require.NoError(t, err)
			require.NotEmpty(t, composition)
			require.True(t, composition.AllNodes().Contains(enterpriseCA))
			require.True(t, composition.AllNodes().Contains(host))
		},
	)
}

// TestADCSForestScoping_UsesHostForestECAAndKeepsCrossForestDomain models
// shared ADCS across two forests: one computer hosts the CA service for an
// EnterpriseCA in its own forest and for a copied EnterpriseCA in another forest.
// Only the host-forest EnterpriseCA should be retained, and it should still chain
// to every domain reached by RootCAFor and TrustedForNTAuth.
func TestADCSForestScoping_UsesHostForestECAAndKeepsCrossForestDomain(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainASID = integration.RandomDomainSID()
		domainBSID = integration.RandomDomainSID()

		hostForestEnterpriseCAID graph.ID
		copiedEnterpriseCAID     graph.ID
		domainAID                graph.ID
		domainBID                graph.ID
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domainA := testContext.NewActiveDirectoryDomain("ForestA-Domain", domainASID, false, true)
			domainB := testContext.NewActiveDirectoryDomain("ForestB-Domain", domainBSID, false, true)

			// The real CA lives in forest A; a copied EnterpriseCA object exists in
			// forest B and points at the same hosting computer.
			hostForestEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("SharedECA", domainASID)
			copiedEnterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("SharedECA-Copy", domainBSID)
			rootCA := testContext.NewActiveDirectoryRootCA("SharedRootCA", domainASID)

			testContext.NewRelationship(hostForestEnterpriseCA, rootCA, ad.EnterpriseCAFor)

			// Valid cert chain to forest A (the CA's own forest)...
			linkEnterpriseCAToDomain(testContext, hostForestEnterpriseCA, rootCA, domainA, domainASID)
			// ...and a cross-forest chain into forest B (shared ADCS).
			linkEnterpriseCAToDomain(testContext, hostForestEnterpriseCA, rootCA, domainB, domainBSID)

			host := addEnabledHostingComputer(testContext, "HostA", domainASID, hostForestEnterpriseCA)
			testContext.NewRelationship(host, copiedEnterpriseCA, ad.HostsCAService)

			hostForestEnterpriseCAID = hostForestEnterpriseCA.ID
			copiedEnterpriseCAID = copiedEnterpriseCA.ID
			domainAID = domainA.ID
			domainBID = domainB.ID
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			chainedDomains := cache.GetECAHostedChainedDomains()

			require.Contains(t, chainedDomains, hostForestEnterpriseCAID.Uint64(), "CA with an in-forest host should be retained")
			assert.NotContains(t, chainedDomains, copiedEnterpriseCAID.Uint64(), "copied CA with only a cross-forest host should be skipped")
			chains := chainedDomains[hostForestEnterpriseCAID.Uint64()]
			assert.True(t, chains.Domains.Contains(domainAID.Uint64()), "in-forest domain should survive")
			assert.True(t, chains.Domains.Contains(domainBID.Uint64()), "cross-forest chained domain should survive")
		},
	)
}

// TestADCSForestScoping_DropsCAWithOnlyCrossForestHost models a CA whose only
// HostsCAService computer was matched across a forest boundary. With no hosting
// computer in the CA's own forest, the CA should be dropped entirely.
func TestADCSForestScoping_DropsCAWithOnlyCrossForestHost(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	var (
		domainASID = integration.RandomDomainSID()
		domainBSID = integration.RandomDomainSID()

		enterpriseCAID graph.ID
	)

	testContext.DatabaseTestWithSetup(
		func(harness *integration.HarnessDetails) error {
			domainA := testContext.NewActiveDirectoryDomain("ForestA-Domain", domainASID, false, true)
			domainB := testContext.NewActiveDirectoryDomain("ForestB-Domain", domainBSID, false, true)

			enterpriseCA := testContext.NewActiveDirectoryEnterpriseCA("SharedECA", domainASID)
			rootCA := testContext.NewActiveDirectoryRootCA("SharedRootCA", domainASID)

			// The CA chains up to its root once; the per-domain edges are added below.
			testContext.NewRelationship(enterpriseCA, rootCA, ad.EnterpriseCAFor)

			linkEnterpriseCAToDomain(testContext, enterpriseCA, rootCA, domainA, domainASID)
			linkEnterpriseCAToDomain(testContext, enterpriseCA, rootCA, domainB, domainBSID)

			// Only hosting computer lives in forest B (cross-forest from the CA).
			addEnabledHostingComputer(testContext, "HostB", domainBSID, enterpriseCA)

			enterpriseCAID = enterpriseCA.ID
			return nil
		},
		func(harness integration.HarnessDetails, db graph.Database) {
			_, cache, err := FetchADCSPrereqs(db)
			require.NoError(t, err)

			chainedDomains := cache.GetECAHostedChainedDomains()

			assert.NotContains(t, chainedDomains, enterpriseCAID.Uint64(), "CA with no in-forest hosting computer should be skipped")
		},
	)
}
