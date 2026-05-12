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

package ad_test

import (
	"context"
	"testing"

	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostADCSESC6a_ManagedServiceAccounts verifies that gMSA and sMSA enrollees
// (both ingested as User-kind nodes) are preserved as ADCSESC6a edge starts when
// the cert template requires DNS in the SubjectAltName, while regular User
// enrollees are filtered out by filterUserDNSResults.
func TestPostADCSESC6a_ManagedServiceAccounts(t *testing.T) {
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
		enterpriseCA = NewActiveDirectoryEnterpriseCA(t, &suite, "EnterpriseCA", domainSID)
		ntAuthStore  = NewActiveDirectoryNTAuthStore(t, &suite, "NTAuthStore", domainSID)
		rootCA       = NewActiveDirectoryRootCA(t, &suite, "RootCA", domainSID)
		regularUser  = NewActiveDirectoryUser(t, &suite, "RegularUser", domainSID)
		gmsaUser     = NewActiveDirectoryUser(t, &suite, "GMSAUser", domainSID)
		smsaUser     = NewActiveDirectoryUser(t, &suite, "SMSAUser", domainSID)
		computer     = NewActiveDirectoryComputer(t, &suite, "Computer", domainSID)
	)

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

	operation := post.NewPostRelationshipOperation(suite.Context, suite.GraphDB, "ADCS Post Process Test - ESC6a MSA")

	localGroupData, enterpriseCertAuthorities, _, domains, cache, err := FetchADCSPrereqs(suite.GraphDB)
	require.NoError(t, err)

	for _, ca := range enterpriseCertAuthorities {
		innerEnterpriseCA := ca
		targetDomains := &graph.NodeSet{}
		for _, candidateDomain := range domains {
			if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, candidateDomain) {
				targetDomains.Add(candidateDomain)
			}
		}

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			if err := adAnalysis.PostADCSESC6a(ctx, tx, outC, localGroupData, innerEnterpriseCA, targetDomains, cache); err != nil {
				t.Logf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
				return err
			}
			return nil
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
		return nil
	})
	require.NoError(t, err)
}
