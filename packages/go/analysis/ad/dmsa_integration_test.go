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

	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestDMSAPostBadSuccessor_UnpatchedDomain(t *testing.T) {
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	domainSID := RandomDomainSID()
	domain := NewActiveDirectoryDomain(t, &suite, "Domain", domainSID, false, true)
	domainController := NewActiveDirectoryComputer(t, &suite, "Domain Controller", domainSID)
	creator := NewActiveDirectoryUser(t, &suite, "Creator", domainSID)

	domainController.Properties.Set(ad.WindowsBuildNumber.String(), 26100)
	domainController.Properties.Set(ad.WindowsBuildRevision.String(), 4945)
	domainController.Properties.Set(ad.WindowsBuildCollected.String(), true)
	UpdateNode(t, &suite, domainController)
	NewRelationship(t, &suite, domainController, domain, ad.DCFor)
	NewRelationship(t, &suite, creator, domain, ad.CreateDMSA)

	_, err := adAnalysis.PostBadSuccessor(suite.Context, suite.GraphDB)
	require.NoError(t, err)

	edges := fetchDMSABadSuccessorEdges(t, &suite)
	require.Len(t, edges, 1)
	require.Equal(t, creator.ID, edges[0].StartID)
	require.Equal(t, domain.ID, edges[0].EndID)
}

func TestDMSAPostBadSuccessor_PatchedDomain(t *testing.T) {
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	domainSID := RandomDomainSID()
	domain := NewActiveDirectoryDomain(t, &suite, "Domain", domainSID, false, true)
	domainController := NewActiveDirectoryComputer(t, &suite, "Domain Controller", domainSID)
	creator := NewActiveDirectoryUser(t, &suite, "Creator", domainSID)
	genericAllVictim := NewActiveDirectoryUser(t, &suite, "GenericAll Victim", domainSID)
	genericWriteVictim := NewActiveDirectoryComputer(t, &suite, "GenericWrite Victim", domainSID)
	attributeVictim := NewActiveDirectoryUser(t, &suite, "Attribute Victim", domainSID)
	incompleteVictim := NewActiveDirectoryUser(t, &suite, "Incomplete Victim", domainSID)
	nonCreator := NewActiveDirectoryUser(t, &suite, "Non-Creator", domainSID)

	domainController.Properties.Set(ad.WindowsBuildNumber.String(), 26100)
	domainController.Properties.Set(ad.WindowsBuildRevision.String(), 4946)
	domainController.Properties.Set(ad.WindowsBuildCollected.String(), true)
	UpdateNode(t, &suite, domainController)
	NewRelationship(t, &suite, domainController, domain, ad.DCFor)
	NewRelationship(t, &suite, creator, domain, ad.CreateDMSA)
	NewRelationship(t, &suite, creator, genericAllVictim, ad.GenericAll)
	NewRelationship(t, &suite, creator, genericWriteVictim, ad.GenericWrite)
	NewRelationship(t, &suite, creator, attributeVictim, ad.WriteSupersededManagedAccountLink)
	NewRelationship(t, &suite, creator, attributeVictim, ad.WriteSupersededServiceAccountState)
	NewRelationship(t, &suite, creator, incompleteVictim, ad.WriteSupersededManagedAccountLink)
	NewRelationship(t, &suite, nonCreator, incompleteVictim, ad.WriteSupersededManagedAccountLink)
	NewRelationship(t, &suite, nonCreator, incompleteVictim, ad.WriteSupersededServiceAccountState)

	_, err := adAnalysis.PostBadSuccessor(suite.Context, suite.GraphDB)
	require.NoError(t, err)

	edges := fetchDMSABadSuccessorEdges(t, &suite)
	require.Len(t, edges, 3)
	require.ElementsMatch(t,
		[]graph.ID{genericAllVictim.ID, genericWriteVictim.ID, attributeVictim.ID},
		[]graph.ID{edges[0].EndID, edges[1].EndID, edges[2].EndID},
	)
	for _, edge := range edges {
		require.Equal(t, creator.ID, edge.StartID)
		require.NotEqual(t, domain.ID, edge.EndID)
		require.NotEqual(t, incompleteVictim.ID, edge.EndID)
	}
}

func TestDMSAPostBadSuccessor_UnknownPatchLevelUsesTargetedModel(t *testing.T) {
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	domainSID := RandomDomainSID()
	domain := NewActiveDirectoryDomain(t, &suite, "Domain", domainSID, false, true)
	domainController := NewActiveDirectoryComputer(t, &suite, "Domain Controller", domainSID)
	creator := NewActiveDirectoryUser(t, &suite, "Creator", domainSID)
	victim := NewActiveDirectoryUser(t, &suite, "Victim", domainSID)

	domainController.Properties.Set(ad.WindowsBuildNumber.String(), 26100)
	domainController.Properties.Set(ad.WindowsBuildRevision.String(), 4945)
	domainController.Properties.Set(ad.WindowsBuildCollected.String(), false)
	UpdateNode(t, &suite, domainController)
	NewRelationship(t, &suite, domainController, domain, ad.DCFor)
	NewRelationship(t, &suite, creator, domain, ad.CreateDMSA)
	NewRelationship(t, &suite, creator, victim, ad.WriteSupersededManagedAccountLink)
	NewRelationship(t, &suite, creator, victim, ad.WriteSupersededServiceAccountState)

	_, err := adAnalysis.PostBadSuccessor(suite.Context, suite.GraphDB)
	require.NoError(t, err)

	edges := fetchDMSABadSuccessorEdges(t, &suite)
	require.Len(t, edges, 1)
	require.Equal(t, creator.ID, edges[0].StartID)
	require.Equal(t, victim.ID, edges[0].EndID)
}

func fetchDMSABadSuccessorEdges(t *testing.T, suite *IntegrationTestSuite) []*graph.Relationship {
	t.Helper()

	var edges []*graph.Relationship
	require.NoError(t, suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		var err error
		edges, err = ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.BadSuccessor)
		}))
		return err
	}))

	return edges
}
