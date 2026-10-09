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
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

type esc16IntegrationHarness struct {
	attacker     *graph.Node
	domain       *graph.Node
	enterpriseCA *graph.Node
	host         *graph.Node
	authStore    *graph.Node
	certTemplate *graph.Node
}

func setupESC16IntegrationHarness(testContext *integration.GraphTestContext, attackerKind graph.Kind) esc16IntegrationHarness {
	var (
		domainSID = integration.RandomDomainSID()
		harness   = esc16IntegrationHarness{
			domain:       testContext.NewActiveDirectoryDomain("Domain", domainSID, false, true),
			enterpriseCA: testContext.NewActiveDirectoryEnterpriseCAWithThumbprint("EnterpriseCA", domainSID, "enterprise-ca"),
			authStore:    testContext.NewActiveDirectoryNTAuthStore("NTAuthStore", domainSID),
			certTemplate: testContext.NewActiveDirectoryCertTemplate("CertTemplate", domainSID, integration.CertTemplateData{
				AuthenticationEnabled: true,
				SchemaVersion:         1,
				NoSecurityExtension:   false,
			}),
		}
		rootCA = testContext.NewActiveDirectoryRootCAWithThumbprint("RootCA", domainSID, "root-ca")
	)

	if attackerKind == ad.User {
		harness.attacker = testContext.NewActiveDirectoryUser("Attacker", domainSID)
	} else {
		harness.attacker = testContext.NewActiveDirectoryGroup("Attacker", domainSID)
	}
	harness.enterpriseCA.Properties.Set(ad.CertChain.String(), []string{"enterprise-ca", "root-ca"})
	harness.enterpriseCA.Properties.Set(ad.IsUserSpecifiesSanEnabled.String(), true)
	harness.enterpriseCA.Properties.Set(ad.IsUserSpecifiesSanEnabledCollected.String(), true)
	harness.enterpriseCA.Properties.Set(ad.DisabledExtensionsCollected.String(), true)
	harness.enterpriseCA.Properties.Set(ad.DisabledExtensions.String(), []string{"1.3.6.1.4.1.311.25.2"})
	testContext.UpdateNode(harness.enterpriseCA)
	harness.authStore.Properties.Set(ad.CertThumbprints.String(), []string{"enterprise-ca"})
	testContext.UpdateNode(harness.authStore)
	harness.host = addEnabledHostingComputer(testContext, "CAHost", domainSID, harness.enterpriseCA)

	testContext.NewRelationship(rootCA, harness.domain, ad.RootCAFor)
	testContext.NewRelationship(harness.authStore, harness.domain, ad.NTAuthStoreFor)
	testContext.NewRelationship(harness.enterpriseCA, rootCA, ad.IssuedSignedBy)
	testContext.NewRelationship(harness.enterpriseCA, harness.authStore, ad.TrustedForNTAuth)
	testContext.NewRelationship(harness.certTemplate, harness.enterpriseCA, ad.PublishedTo)
	testContext.NewRelationship(harness.attacker, harness.certTemplate, ad.Enroll)
	testContext.NewRelationship(harness.attacker, harness.enterpriseCA, ad.Enroll)
	return harness
}

func fetchESC16IntegrationEdges(t *testing.T, database graph.Database) []*graph.Relationship {
	t.Helper()

	var (
		relationships []*graph.Relationship
	)

	require.NoError(t, database.ReadTransaction(t.Context(), func(transaction graph.Transaction) error {
		var err error
		relationships, err = ops.FetchRelationships(transaction.Relationships().Filter(query.Kind(query.Relationship(), ad.ADCSESC16)))
		return err
	}))
	return relationships
}

func TestADCSESC16DeltaLifecycle(t *testing.T) {
	var (
		testCases = []struct {
			name     string
			target   func(esc16IntegrationHarness) *graph.Node
			property string
			value    any
			remove   bool
		}{
			{name: "security extension restored", target: func(harness esc16IntegrationHarness) *graph.Node { return harness.enterpriseCA }, property: ad.DisabledExtensions.String(), value: []string{}},
			{name: "disabled extensions collection disabled", target: func(harness esc16IntegrationHarness) *graph.Node { return harness.enterpriseCA }, property: ad.DisabledExtensionsCollected.String(), value: false},
			{name: "disabled extensions collection missing", target: func(harness esc16IntegrationHarness) *graph.Node { return harness.enterpriseCA }, property: ad.DisabledExtensionsCollected.String(), remove: true},
			{name: "SAN collection disabled", target: func(harness esc16IntegrationHarness) *graph.Node { return harness.enterpriseCA }, property: ad.IsUserSpecifiesSanEnabledCollected.String(), value: false},
			{name: "SAN collection missing", target: func(harness esc16IntegrationHarness) *graph.Node { return harness.enterpriseCA }, property: ad.IsUserSpecifiesSanEnabledCollected.String(), remove: true},
			{name: "CA host disabled", target: func(harness esc16IntegrationHarness) *graph.Node { return harness.host }, property: common.Enabled.String(), value: false},
			{name: "NTAuth trust removed", target: func(harness esc16IntegrationHarness) *graph.Node { return harness.authStore }, property: ad.CertThumbprints.String(), value: []string{}},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				testContext = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
				harness     esc16IntegrationHarness
			)
			testContext.DatabaseTestWithSetup(func(_ *integration.HarnessDetails) error {
				harness = setupESC16IntegrationHarness(testContext, ad.Group)
				return nil
			}, func(_ integration.HarnessDetails, database graph.Database) {
				// Exercise production transit-edge cleanup as well as the ADCS sink.
				_, err := adAnalysis.Post(t.Context(), database, false, false)
				require.NoError(t, err)
				originalEdges := fetchESC16IntegrationEdges(t, database)
				require.Len(t, originalEdges, 1)
				originalEdge := originalEdges[0]
				require.Equal(t, harness.attacker.ID, originalEdge.StartID)
				require.Equal(t, harness.domain.ID, originalEdge.EndID)
				firstSeen, err := originalEdge.Properties.Get(common.FirstSeen.String()).Time()
				require.NoError(t, err)
				require.False(t, firstSeen.IsZero())
				require.NotContains(t, originalEdge.Properties.Map, common.LastSeen.String())

				_, err = adAnalysis.Post(t.Context(), database, false, false)
				require.NoError(t, err)
				repeatedEdges := fetchESC16IntegrationEdges(t, database)
				require.Len(t, repeatedEdges, 1)
				require.Equal(t, originalEdge.ID, repeatedEdges[0].ID)
				require.Equal(t, originalEdge.Properties.Map, repeatedEdges[0].Properties.Map)

				require.NoError(t, database.WriteTransaction(t.Context(), func(transaction graph.Transaction) error {
					node, err := ops.FetchNode(transaction, testCase.target(harness).ID)
					if err != nil {
						return err
					}
					if testCase.remove {
						node.Properties.Delete(testCase.property)
					} else {
						node.Properties.Set(testCase.property, testCase.value)
					}
					return transaction.UpdateNode(node)
				}))
				_, err = adAnalysis.Post(t.Context(), database, false, false)
				require.NoError(t, err)
				require.Empty(t, fetchESC16IntegrationEdges(t, database))
				composition, err := adAnalysis.GetADCSESC16EdgeComposition(t.Context(), database, originalEdge)
				require.NoError(t, err)
				require.Empty(t, composition)
			})
		})
	}
}

func TestADCSESC16CompositionHostForest(t *testing.T) {
	var (
		testContext = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		harness     esc16IntegrationHarness
	)
	testContext.DatabaseTestWithSetup(func(_ *integration.HarnessDetails) error {
		harness = setupESC16IntegrationHarness(testContext, ad.Group)
		otherDomainSID := integration.RandomDomainSID()
		testContext.NewActiveDirectoryDomain("Other forest", otherDomainSID, false, true)
		harness.host.Properties.Set(ad.DomainSID.String(), otherDomainSID)
		testContext.UpdateNode(harness.host)
		return nil
	}, func(_ integration.HarnessDetails, database graph.Database) {
		_, err := adAnalysis.Post(t.Context(), database, false, false)
		require.NoError(t, err)
		require.Empty(t, fetchESC16IntegrationEdges(t, database))
		composition, err := adAnalysis.GetADCSESC16EdgeComposition(t.Context(), database, graph.NewRelationship(0, harness.attacker.ID, harness.domain.ID, graph.NewProperties(), ad.ADCSESC16))
		require.NoError(t, err)
		require.Empty(t, composition)
	})
}

func TestADCSESC16RetainsOtherCAContributor(t *testing.T) {
	var (
		testContext = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		harness     esc16IntegrationHarness
	)
	testContext.DatabaseTestWithSetup(func(_ *integration.HarnessDetails) error {
		harness = setupESC16IntegrationHarness(testContext, ad.Group)
		return nil
	}, func(_ integration.HarnessDetails, database graph.Database) {
		var (
			otherCA    *graph.Node
			legacyEdge *graph.Relationship
		)

		require.NoError(t, database.WriteTransaction(t.Context(), func(transaction graph.Transaction) error {
			var (
				properties = graph.NewProperties()
				err        error
			)
			properties.Set(common.LastSeen.String(), time.Now().UTC())
			legacyEdge, err = transaction.CreateRelationshipByIDs(harness.attacker.ID, harness.domain.ID, ad.ADCSESC16, properties)
			return err
		}))

		_, err := adAnalysis.Post(t.Context(), database, false, false)
		require.NoError(t, err)
		originalEdges := fetchESC16IntegrationEdges(t, database)
		require.Len(t, originalEdges, 1)
		require.NotEqual(t, legacyEdge.ID, originalEdges[0].ID)
		require.Contains(t, originalEdges[0].Properties.Map, common.FirstSeen.String())
		require.NotContains(t, originalEdges[0].Properties.Map, common.LastSeen.String())

		testContext.SetupHarness(func(_ *integration.HarnessDetails) error {
			domainSID, err := harness.domain.Properties.Get(ad.DomainSID.String()).String()
			if err != nil {
				return err
			}
			otherCA = testContext.NewActiveDirectoryEnterpriseCAWithThumbprint("Other CA", domainSID, "other-ca")
			otherCA.Properties.Set(ad.CertChain.String(), []string{"other-ca", "root-ca"})
			otherCA.Properties.Set(ad.IsUserSpecifiesSanEnabled.String(), true)
			otherCA.Properties.Set(ad.IsUserSpecifiesSanEnabledCollected.String(), true)
			otherCA.Properties.Set(ad.DisabledExtensionsCollected.String(), true)
			otherCA.Properties.Set(ad.DisabledExtensions.String(), []string{"1.3.6.1.4.1.311.25.2"})
			testContext.UpdateNode(otherCA)
			harness.authStore.Properties.Set(ad.CertThumbprints.String(), []string{"enterprise-ca", "other-ca"})
			testContext.UpdateNode(harness.authStore)
			testContext.NewRelationship(harness.host, otherCA, ad.HostsCAService)
			testContext.NewRelationship(harness.certTemplate, otherCA, ad.PublishedTo)
			testContext.NewRelationship(harness.attacker, otherCA, ad.Enroll)
			return nil
		})
		_, err = adAnalysis.Post(t.Context(), database, false, false)
		require.NoError(t, err)
		require.Len(t, fetchESC16IntegrationEdges(t, database), 1)

		harness.enterpriseCA.Properties.Set(ad.DisabledExtensions.String(), []string{})
		testContext.UpdateNode(harness.enterpriseCA)
		_, err = adAnalysis.Post(t.Context(), database, false, false)
		require.NoError(t, err)
		remainingEdges := fetchESC16IntegrationEdges(t, database)
		require.Len(t, remainingEdges, 1)
		require.Equal(t, originalEdges[0].ID, remainingEdges[0].ID)
		require.Equal(t, originalEdges[0].Properties.Map, remainingEdges[0].Properties.Map)
		composition, err := adAnalysis.GetADCSESC16EdgeComposition(t.Context(), database, remainingEdges[0])
		require.NoError(t, err)
		require.True(t, composition.AllNodes().Contains(otherCA))
		require.False(t, composition.AllNodes().Contains(harness.enterpriseCA))
		requireCompositionContainsEdge(t, composition, ad.HostsCAService)
	})
}

func TestADCSESC16ManagedServiceAccountComposition(t *testing.T) {
	var (
		testCases = []struct {
			name              string
			accountProperty   string
			expectedEdgeCount int
		}{
			{name: "ordinary user", expectedEdgeCount: 0},
			{name: "group managed service account", accountProperty: ad.GMSA.String(), expectedEdgeCount: 1},
			{name: "standalone managed service account", accountProperty: ad.MSA.String(), expectedEdgeCount: 1},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				testContext = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
				harness     esc16IntegrationHarness
			)
			testContext.DatabaseTestWithSetup(func(_ *integration.HarnessDetails) error {
				harness = setupESC16IntegrationHarness(testContext, ad.User)
				harness.certTemplate.Properties.Set(ad.SubjectAltRequireDNS.String(), true)
				testContext.UpdateNode(harness.certTemplate)
				if testCase.accountProperty != "" {
					harness.attacker.Properties.Set(testCase.accountProperty, true)
					testContext.UpdateNode(harness.attacker)
				}
				return nil
			}, func(_ integration.HarnessDetails, database graph.Database) {
				_, err := adAnalysis.Post(t.Context(), database, false, false)
				require.NoError(t, err)
				require.Len(t, fetchESC16IntegrationEdges(t, database), testCase.expectedEdgeCount)
				composition, err := adAnalysis.GetADCSESC16EdgeComposition(t.Context(), database, graph.NewRelationship(0, harness.attacker.ID, harness.domain.ID, graph.NewProperties(), ad.ADCSESC16))
				require.NoError(t, err)
				if testCase.expectedEdgeCount == 0 {
					require.Empty(t, composition)
				} else {
					requireCompositionContainsEdge(t, composition, ad.HostsCAService)
					require.True(t, composition.AllNodes().Contains(harness.host))
					require.True(t, composition.AllNodes().Contains(harness.attacker))
				}
			})
		})
	}
}
