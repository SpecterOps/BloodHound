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

	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManagedServiceAccountDNSCompositions verifies that gMSA and sMSA enrollees
// are preserved in ADCS compositions when the cert template requires DNS.
func TestManagedServiceAccountDNSCompositions(t *testing.T) {
	t.Parallel()

	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		domainSID    = RandomDomainSID()
		domain       = NewActiveDirectoryDomain(t, &suite, "ESC6aMSA", domainSID, false, true)
		certTemplate = NewActiveDirectoryCertTemplate(t, &suite, "CertTemplate", domainSID, CertTemplateProperties{
			AuthenticationEnabled: true,
			NoSecurityExtension:   true,
			SchemaVersion:         1,
			SubjectAltRequireDNS:  true,
		})
		enterpriseCA     = NewActiveDirectoryEnterpriseCA(t, &suite, "EnterpriseCA", domainSID)
		ntAuthStore      = NewActiveDirectoryNTAuthStore(t, &suite, "NTAuthStore", domainSID)
		rootCA           = NewActiveDirectoryRootCA(t, &suite, "RootCA", domainSID)
		regularUser      = NewActiveDirectoryUser(t, &suite, "RegularUser", domainSID)
		gmsaUser         = NewActiveDirectoryUser(t, &suite, "GMSAUser", domainSID)
		smsaUser         = NewActiveDirectoryUser(t, &suite, "SMSAUser", domainSID)
		computer         = NewActiveDirectoryComputer(t, &suite, "Computer", domainSID)
		ecaHost          = NewActiveDirectoryComputer(t, &suite, "ECA Host", domainSID)
		domainController = NewActiveDirectoryComputer(t, &suite, "Domain Controller", domainSID)
		attacker         = NewActiveDirectoryUser(t, &suite, "Attacker", domainSID)

		managedServiceAccountEdges []*graph.Relationship
		attackerEdges              = map[graph.Kind]*graph.Relationship{}
	)

	ecaHost.Properties.Set(common.Enabled.String(), true)
	UpdateNode(t, &suite, ecaHost)
	domainController.Properties.Set(ad.CertificateMappingMethodsRaw.String(), "4")
	domainController.Properties.Set(ad.StrongCertificateBindingEnforcementRaw.String(), "1")
	UpdateNode(t, &suite, domainController)
	certTemplate.Properties.Set(ad.SubjectAltRequireUPN.String(), true)
	certTemplate.Properties.Set(ad.EffectiveEKUs.String(), []string{})
	UpdateNode(t, &suite, certTemplate)
	NewRelationship(t, &suite, ecaHost, enterpriseCA, ad.HostsCAService)
	NewRelationship(t, &suite, domainController, domain, ad.DCFor)

	gmsaUser.Properties.Set(ad.GMSA.String(), true)
	UpdateNode(t, &suite, gmsaUser)
	smsaUser.Properties.Set(ad.MSA.String(), true)
	UpdateNode(t, &suite, smsaUser)
	enterpriseCA.Properties.Set(ad.IsUserSpecifiesSanEnabled.String(), true)
	enterpriseCA.Properties.Set(ad.IsUserSpecifiesSanEnabledCollected.String(), true)
	UpdateNode(t, &suite, enterpriseCA)

	NewRelationship(t, &suite, regularUser, certTemplate, ad.Enroll)
	NewRelationship(t, &suite, gmsaUser, certTemplate, ad.Enroll)
	NewRelationship(t, &suite, smsaUser, certTemplate, ad.Enroll)
	NewRelationship(t, &suite, computer, certTemplate, ad.Enroll)
	NewRelationship(t, &suite, regularUser, enterpriseCA, ad.Enroll)
	NewRelationship(t, &suite, gmsaUser, enterpriseCA, ad.Enroll)
	NewRelationship(t, &suite, smsaUser, enterpriseCA, ad.Enroll)
	NewRelationship(t, &suite, computer, enterpriseCA, ad.Enroll)
	NewRelationship(t, &suite, certTemplate, enterpriseCA, ad.PublishedTo)
	NewRelationship(t, &suite, enterpriseCA, ntAuthStore, ad.TrustedForNTAuth)
	NewRelationship(t, &suite, enterpriseCA, rootCA, ad.IssuedSignedBy)
	NewRelationship(t, &suite, ntAuthStore, domain, ad.NTAuthStoreFor)
	NewRelationship(t, &suite, rootCA, domain, ad.RootCAFor)
	NewRelationship(t, &suite, attacker, gmsaUser, ad.GenericAll)
	NewRelationship(t, &suite, attacker, smsaUser, ad.GenericAll)

	operation := post.NewPostRelationshipOperation(suite.Context, suite.GraphDB, "ADCS managed service account composition test")

	localGroupData, cache, err := FetchADCSPrereqs(suite.GraphDB)
	require.NoError(t, err)

	for _, certChains := range cache.GetECAHostedChainedDomains() {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			if err := adAnalysis.PostADCSESC6a(ctx, tx, outC, localGroupData, certChains, cache); err != nil {
				return err
			} else if err := adAnalysis.PostADCSESC9a(ctx, tx, outC, localGroupData, certChains, cache); err != nil {
				return err
			} else {
				return adAnalysis.PostADCSESC10a(ctx, tx, outC, localGroupData, certChains, cache)
			}
		})
	}

	require.NoError(t, operation.Done())

	err = suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.ADCSESC6a)
		}))
		require.NoError(t, err)

		// gMSA, sMSA, and Computer enrollers should produce ESC6a edges even when the
		// cert template requires DNS in the SubjectAltName. The plain User enroller
		// should be filtered out by filterUserDNSResults.
		assert.Equal(t, 3, len(results))
		assert.True(t, results.Contains(gmsaUser),
			"gMSA enroller should retain its ESC6a edge when SubjectAltRequireDNS is true")
		assert.True(t, results.Contains(smsaUser),
			"sMSA enroller should retain its ESC6a edge when SubjectAltRequireDNS is true")
		assert.True(t, results.Contains(computer),
			"Computer enroller should retain its ESC6a edge when SubjectAltRequireDNS is true")
		assert.False(t, results.Contains(regularUser),
			"plain User enroller should be filtered out when SubjectAltRequireDNS is true")

		managedServiceAccountEdges, err = ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Relationship(), ad.ADCSESC6a),
				query.InIDs(query.StartID(), gmsaUser.ID, smsaUser.ID),
			)
		}))
		require.NoError(t, err)

		for _, edgeKind := range []graph.Kind{ad.ADCSESC9a, ad.ADCSESC10a} {
			attackerEdge, err := tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), edgeKind),
					query.Equals(query.StartID(), attacker.ID),
					query.Equals(query.EndID(), domain.ID),
				)
			}).First()
			require.NoError(t, err)
			attackerEdges[edgeKind] = attackerEdge
		}

		return nil
	})
	require.NoError(t, err)
	require.Len(t, managedServiceAccountEdges, 2)

	for _, edge := range managedServiceAccountEdges {
		composition, err := adAnalysis.GetADCSESC6EdgeComposition(suite.Context, suite.GraphDB, edge)
		require.NoError(t, err)
		require.NotEmpty(t, composition)
		require.True(t, composition.AllNodes().ContainsID(edge.StartID))
	}

	for edgeKind, edge := range attackerEdges {
		composition, err := adAnalysis.GetEdgeCompositionPath(suite.Context, suite.GraphDB, edge)
		require.NoError(t, err, edgeKind.String())
		require.NotEmpty(t, composition, edgeKind.String())
		require.True(t, composition.AllNodes().Contains(gmsaUser), edgeKind.String())
		require.True(t, composition.AllNodes().Contains(smsaUser), edgeKind.String())
	}

}

func TestPostADCSESC13_ManagedServiceAccountDNSRequirements(t *testing.T) {
	t.Parallel()

	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		domainSID           = RandomDomainSID()
		domain              = NewActiveDirectoryDomain(t, &suite, "ESC13 managed service accounts", domainSID, false, true)
		dnsFreeCertTemplate = NewActiveDirectoryCertTemplate(t, &suite, "DNS-free cert template", domainSID, CertTemplateProperties{
			AuthenticationEnabled: true,
			SchemaVersion:         1,
			EffectiveEKUs:         []string{},
			ApplicationPolicies:   []string{},
		})
		dnsCertTemplate = NewActiveDirectoryCertTemplate(t, &suite, "DNS cert template", domainSID, CertTemplateProperties{
			AuthenticationEnabled: true,
			SchemaVersion:         1,
			SubjectAltRequireDNS:  true,
			EffectiveEKUs:         []string{},
			ApplicationPolicies:   []string{},
		})
		enterpriseCA = NewActiveDirectoryEnterpriseCA(t, &suite, "EnterpriseCA", domainSID)
		ntAuthStore  = NewActiveDirectoryNTAuthStore(t, &suite, "NTAuthStore", domainSID)
		rootCA       = NewActiveDirectoryRootCA(t, &suite, "RootCA", domainSID)
		regularUser  = NewActiveDirectoryUser(t, &suite, "RegularUser", domainSID)
		dnsOnlyUser  = NewActiveDirectoryUser(t, &suite, "DNSOnlyUser", domainSID)
		gmsaUser     = NewActiveDirectoryUser(t, &suite, "GMSAUser", domainSID)
		smsaUser     = NewActiveDirectoryUser(t, &suite, "SMSAUser", domainSID)
		ecaHost      = NewActiveDirectoryComputer(t, &suite, "ECA Host", domainSID)
		targetGroup  = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:     "TargetGroup",
			common.ObjectID: newRandomObjectID(t),
			ad.DomainSID:    domainSID,
		}), ad.Entity, ad.Group)
		issuancePolicy = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:     "IssuancePolicy",
			common.ObjectID: newRandomObjectID(t),
			ad.DomainSID:    domainSID,
		}), ad.Entity, ad.IssuancePolicy)

		esc13EdgesByStartID = map[graph.ID]*graph.Relationship{}
	)

	ecaHost.Properties.Set(common.Enabled.String(), true)
	UpdateNode(t, &suite, ecaHost)
	gmsaUser.Properties.Set(ad.GMSA.String(), true)
	UpdateNode(t, &suite, gmsaUser)
	smsaUser.Properties.Set(ad.MSA.String(), true)
	UpdateNode(t, &suite, smsaUser)
	for _, certTemplate := range []*graph.Node{dnsFreeCertTemplate, dnsCertTemplate} {
		certTemplate.Properties.Set(ad.SubjectAltRequireDomainDNS.String(), false)
		UpdateNode(t, &suite, certTemplate)
	}

	NewRelationship(t, &suite, ecaHost, enterpriseCA, ad.HostsCAService)
	NewRelationship(t, &suite, enterpriseCA, ntAuthStore, ad.TrustedForNTAuth)
	NewRelationship(t, &suite, enterpriseCA, rootCA, ad.IssuedSignedBy)
	NewRelationship(t, &suite, ntAuthStore, domain, ad.NTAuthStoreFor)
	NewRelationship(t, &suite, rootCA, domain, ad.RootCAFor)
	NewRelationship(t, &suite, domain, targetGroup, ad.Contains)
	NewRelationship(t, &suite, issuancePolicy, targetGroup, ad.OIDGroupLink)

	for _, certTemplate := range []*graph.Node{dnsFreeCertTemplate, dnsCertTemplate} {
		NewRelationship(t, &suite, certTemplate, enterpriseCA, ad.PublishedTo)
		NewRelationship(t, &suite, certTemplate, issuancePolicy, ad.ExtendedByPolicy)
	}

	for _, principal := range []*graph.Node{regularUser, dnsOnlyUser, gmsaUser, smsaUser} {
		NewRelationship(t, &suite, principal, enterpriseCA, ad.Enroll)
	}
	NewRelationship(t, &suite, regularUser, dnsFreeCertTemplate, ad.Enroll)
	for _, principal := range []*graph.Node{regularUser, dnsOnlyUser, gmsaUser, smsaUser} {
		NewRelationship(t, &suite, principal, dnsCertTemplate, ad.Enroll)
	}

	operation := post.NewPostRelationshipOperation(suite.Context, suite.GraphDB, "ADCS Post Process Test - ESC13 managed service account DNS requirements")

	localGroupData, cache, err := FetchADCSPrereqs(suite.GraphDB)
	require.NoError(t, err)

	for _, certChains := range cache.GetECAHostedChainedDomains() {
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			return adAnalysis.PostADCSESC13(ctx, tx, outC, localGroupData, certChains, cache)
		})
	}

	require.NoError(t, operation.Done())

	err = suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		edges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.ADCSESC13)
		}))
		if err != nil {
			return err
		}

		for _, edge := range edges {
			esc13EdgesByStartID[edge.StartID] = edge
		}

		return nil
	})
	require.NoError(t, err)
	require.Len(t, esc13EdgesByStartID, 3)
	require.Contains(t, esc13EdgesByStartID, regularUser.ID)
	require.Contains(t, esc13EdgesByStartID, gmsaUser.ID)
	require.Contains(t, esc13EdgesByStartID, smsaUser.ID)
	require.NotContains(t, esc13EdgesByStartID, dnsOnlyUser.ID)

	regularUserComposition, err := adAnalysis.GetADCSESC13EdgeComposition(suite.Context, suite.GraphDB, esc13EdgesByStartID[regularUser.ID])
	require.NoError(t, err)
	require.NotEmpty(t, regularUserComposition)
	assert.True(t, regularUserComposition.AllNodes().Contains(dnsFreeCertTemplate))
	assert.False(t, regularUserComposition.AllNodes().Contains(dnsCertTemplate))

	for _, managedServiceAccount := range []*graph.Node{gmsaUser, smsaUser} {
		composition, err := adAnalysis.GetADCSESC13EdgeComposition(suite.Context, suite.GraphDB, esc13EdgesByStartID[managedServiceAccount.ID])
		require.NoError(t, err)
		require.NotEmpty(t, composition)
		assert.True(t, composition.AllNodes().Contains(dnsCertTemplate))
		assert.True(t, composition.AllNodes().Contains(managedServiceAccount))
	}

	legacyDNSOnlyEdge := NewRelationship(t, &suite, dnsOnlyUser, targetGroup, ad.ADCSESC13)
	legacyComposition, err := adAnalysis.GetADCSESC13EdgeComposition(suite.Context, suite.GraphDB, legacyDNSOnlyEdge)
	require.NoError(t, err)
	assert.Empty(t, legacyComposition)
}
