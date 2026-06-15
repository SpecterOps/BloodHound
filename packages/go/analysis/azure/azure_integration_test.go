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

package azure_test

import (
	"context"
	"slices"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	graphAzure "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAZAddOwner(t *testing.T) {
	t.Parallel()

	//#region Setup for test
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		tenantID                = integration.RandomObjectID(t)
		AZTenant                = NewAzureTenant(t, &suite, tenantID)
		AZApp                   = NewAzureApplication(t, &suite, "AZApp", integration.RandomObjectID(t), tenantID)
		AZServicePrincipal      = NewAzureServicePrincipal(t, &suite, "AZServicePrincipal", integration.RandomObjectID(t), tenantID)
		HybridIdentityAdminRole = NewAzureRole(t, &suite, "HybridIdentityAdminRole", integration.RandomObjectID(t), graphAzure.HybridIdentityAdministratorRole, tenantID)
		PartnerTier1SupportRole = NewAzureRole(t, &suite, "PartnerTier1SupportRole", integration.RandomObjectID(t), graphAzure.PartnerTier1SupportRole, tenantID)
		PartnerTier2SupportRole = NewAzureRole(t, &suite, "PartnerTier2SupportRole", integration.RandomObjectID(t), graphAzure.PartnerTier2SupportRole, tenantID)
		DirSyncAccountsRole     = NewAzureRole(t, &suite, "DirSyncAccountsRole", integration.RandomObjectID(t), graphAzure.DirectorySynchronizationAccountsRole, tenantID)
		// Role that should not generate AZAddOwner Edge
		ConditionalAccessAdministratorRole = NewAzureRole(t, &suite, "ConditionalAccessAdministratorRole", integration.RandomObjectID(t), graphAzure.ConditionalAccessAdministratorRole, tenantID)
	)

	NewRelationship(t, &suite, AZTenant, AZApp, graphAzure.Contains)
	NewRelationship(t, &suite, AZTenant, AZServicePrincipal, graphAzure.Contains)
	NewRelationship(t, &suite, AZTenant, HybridIdentityAdminRole, graphAzure.Contains)
	NewRelationship(t, &suite, AZTenant, PartnerTier1SupportRole, graphAzure.Contains)
	NewRelationship(t, &suite, AZTenant, PartnerTier2SupportRole, graphAzure.Contains)
	NewRelationship(t, &suite, AZTenant, DirSyncAccountsRole, graphAzure.Contains)
	NewRelationship(t, &suite, AZTenant, ConditionalAccessAdministratorRole, graphAzure.Contains)
	//#endregion

	//#region Running AZAddOwner edge post processing
	postProcessingStats, err := azure.CreateAZAddOwnerEdge(context.Background(), suite.GraphDB)
	//#endregion

	//#region Verifying assertions
	require.NoError(t, err)
	require.NotNil(t, postProcessingStats.RelationshipsCreated[graphAzure.AddOwner])
	assert.Equal(t, 8, int(*postProcessingStats.RelationshipsCreated[graphAzure.AddOwner]))

	// Validate that the AZAddOwner edges were created
	err = suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		addOwnerEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), graphAzure.AddOwner)
		}))
		require.NoError(t, err)
		assert.Len(t, addOwnerEdges, 8)

		type edgeKey struct {
			start graph.ID
			end   graph.ID
		}
		expected := make(map[edgeKey]struct{}, 8)
		expected[edgeKey{start: HybridIdentityAdminRole.ID, end: AZApp.ID}] = struct{}{}
		expected[edgeKey{start: HybridIdentityAdminRole.ID, end: AZServicePrincipal.ID}] = struct{}{}
		expected[edgeKey{start: PartnerTier1SupportRole.ID, end: AZApp.ID}] = struct{}{}
		expected[edgeKey{start: PartnerTier1SupportRole.ID, end: AZServicePrincipal.ID}] = struct{}{}
		expected[edgeKey{start: PartnerTier2SupportRole.ID, end: AZApp.ID}] = struct{}{}
		expected[edgeKey{start: PartnerTier2SupportRole.ID, end: AZServicePrincipal.ID}] = struct{}{}
		expected[edgeKey{start: DirSyncAccountsRole.ID, end: AZApp.ID}] = struct{}{}
		expected[edgeKey{start: DirSyncAccountsRole.ID, end: AZServicePrincipal.ID}] = struct{}{}

		for _, edge := range addOwnerEdges {
			key := edgeKey{start: edge.StartID, end: edge.EndID}
			_, present := expected[key]
			assert.True(t, present)
			delete(expected, key)
		}

		assert.Empty(t, expected, "not all expected edges were found")

		return nil
	})

	require.NoError(t, err)
	//#endregion
}

func TestFetchEntityByObjectID(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		node, err := azure.FetchEntityByObjectID(tx, testContext.NodeObjectID(harness.AZBaseHarness.Application))

		require.Nil(t, err)
		assert.Equal(t, harness.AZBaseHarness.Application.ID, node.ID)
	})
}

func TestEntityRoles(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		roles, err := azure.FetchEntityRoles(tx, harness.AZBaseHarness.User, 0, 0)

		require.Nil(t, err)
		assert.ElementsMatch(t, harness.AZBaseHarness.Nodes.Get(graphAzure.Role).IDs(), roles.ContainingNodeKinds(graphAzure.Role).IDs())
	})
}

func TestTraverseNodePaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		// Preform a full traversal of all outbound paths from the user node
		if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:        harness.AZBaseHarness.User,
			Direction:   graph.DirectionOutbound,
			BranchQuery: nil,
		}); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, harness.AZBaseHarness.NumPaths, paths.Len())

			harnessNodes := harness.AZBaseHarness.Nodes.AllNodes().IDs()
			// we have 4 extra nodes are from the AZGroupMembership harness
			assert.Equal(t, len(harnessNodes), len(paths.AllNodes().IDs()))
		}

		// Preform a traversal of only the outbound HasRole paths from the user node
		if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      harness.AZBaseHarness.User,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), graphAzure.HasRole)
			},
		}); err != nil {
			t.Fatal(err)
		} else {
			numRoles := harness.AZBaseHarness.Nodes.Count(graphAzure.Role)
			assert.Equal(t, harness.AZBaseHarness.Nodes.Get(graphAzure.Role).Len(), paths.Len())

			// Add one to the number of roles since the user is included in the result set
			require.EqualValues(t, numRoles+1, paths.AllNodes().Len())
		}
	})
}

func TestAzureEntityRoles(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		if roles, err := azure.FetchEntityRoles(tx, harness.AZBaseHarness.User, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.ElementsMatch(t, harness.AZBaseHarness.Nodes.Get(graphAzure.Role).IDs(), roles.IDs())
		}
	})
}

func TestEntityEligibleRoles(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEligibleAndApproverRoleHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		directRoles, err := azure.FetchEntityEligibleRoles(tx, harness.AZEligibleAndApproverRoleHarness.UserDirectEligible, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, directRoles.Len())
		assert.True(t, directRoles.Contains(harness.AZEligibleAndApproverRoleHarness.RoleDirect))

		groupRoles, err := azure.FetchEntityEligibleRoles(tx, harness.AZEligibleAndApproverRoleHarness.UserGroupEligible, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, groupRoles.Len())
		assert.True(t, groupRoles.Contains(harness.AZEligibleAndApproverRoleHarness.RoleViaGroup))

		emptyRoles, err := azure.FetchEntityEligibleRoles(tx, harness.AZEligibleAndApproverRoleHarness.UserNoEligibility, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 0, emptyRoles.Len())
	})
}

func TestEntityApproverRoles(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEligibleAndApproverRoleHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		directRoles, err := azure.FetchEntityApproverRoles(tx, harness.AZEligibleAndApproverRoleHarness.UserDirectApprover, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, directRoles.Len())
		assert.True(t, directRoles.Contains(harness.AZEligibleAndApproverRoleHarness.RoleDirectApprover))

		groupRoles, err := azure.FetchEntityApproverRoles(tx, harness.AZEligibleAndApproverRoleHarness.UserGroupApprover, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, groupRoles.Len())
		assert.True(t, groupRoles.Contains(harness.AZEligibleAndApproverRoleHarness.RoleGroupApprover))

		emptyRoles, err := azure.FetchEntityApproverRoles(tx, harness.AZEligibleAndApproverRoleHarness.UserNoEligibility, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 0, emptyRoles.Len())
	})
}

func TestAzureEntityGroupMembership(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		if groupPaths, err := azure.FetchEntityGroupMembershipPaths(tx, harness.AZBaseHarness.User); err != nil {
			t.Fatal(err)
		} else {
			assert.ElementsMatch(t, harness.AZBaseHarness.UserFirstDegreeGroups.IDs(), groupPaths.AllNodes().ContainingNodeKinds(graphAzure.Group).IDs())
		}
	})
}

func TestAZMGApplicationReadWriteAll(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGApplicationReadWriteAllHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph))
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.Application))
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipalB))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.Application, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipalB, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azure.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{graphAzure.AZMGAddOwner, graphAzure.AZMGAddSecret}, []graph.Kind{edge.Kind})
				}
			}
		}

	})
}

func TestAZMGAppRoleManagementReadWriteAll(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGAppRoleManagementReadWriteAllHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.Tenant))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.Tenant, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azure.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Equal(t, graphAzure.AZMGGrantAppRoles, edge.Kind)
				}
			}
		}

	})
}

func TestAZMGDirectoryReadWriteAll(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGDirectoryReadWriteAllHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGDirectoryReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGDirectoryReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGDirectoryReadWriteAllHarness.Group))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGDirectoryReadWriteAllHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azure.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{graphAzure.AZMGAddOwner, graphAzure.AZMGAddMember}, []graph.Kind{edge.Kind})
				}
			}
		}

	})
}

func TestAZMGGroupReadWriteAll(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGGroupReadWriteAllHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupReadWriteAllHarness.Group))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupReadWriteAllHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azure.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGGroupReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{graphAzure.AZMGAddOwner, graphAzure.AZMGAddMember}, []graph.Kind{edge.Kind})
				}
			}
		}

	})
}

func TestAZMGGroupMemberReadWriteAll(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGGroupMemberReadWriteAllHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupMemberReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupMemberReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupMemberReadWriteAllHarness.Group))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupMemberReadWriteAllHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azure.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{graphAzure.AZMGAddOwner, graphAzure.AZMGAddMember}, []graph.Kind{edge.Kind})
				}
			}
		}

	})
}

func TestAZMGRoleManagementReadWriteDirectory(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGRoleManagementReadWriteDirectoryHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.Group))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.Application))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipalB))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.Role))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.MicrosoftGraph))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azure.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{graphAzure.AZMGAddOwner, graphAzure.AZMGAddSecret, graphAzure.AZMGGrantRole}, []graph.Kind{edge.Kind})
				}
			}
		}

	})
}

func TestAZMGServicePrincipalEndpointReadWriteAll(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGServicePrincipalEndpointReadWriteAllHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azure.FetchAbusableAppRoleAssignments(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipalB))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azure.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.MicrosoftGraph))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azure.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Equal(t, graphAzure.AZMGAddOwner, edge.Kind)
				}
			}
		}

	})
}

/**********************
 * Entity Panel tests *
 **********************/

func TestEntityDetails(t *testing.T) {
	var (
		testContext = integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
		dbInst      = integration.SetupDB(t)
	)
	primaryDisplayKinds, err := dbInst.GetPrimaryDisplayKinds(testContext.Context())
	require.NoError(t, err)

	t.Run("ApplicationEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			appObjectID, err := harness.AZEntityPanelHarness.Application.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", appObjectID)

			app, err := azure.ApplicationEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, appObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.Application.Properties.Get(common.ObjectID.String()).Any(), app.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, app.InboundObjectControl)

			app, err = azure.ApplicationEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, appObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, app.InboundObjectControl)
		})
	})

	t.Run("DeviceEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			deviceObjectID, err := harness.AZEntityPanelHarness.Device.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", deviceObjectID)

			device, err := azure.DeviceEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, deviceObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.Device.Properties.Get(common.ObjectID.String()).Any(), device.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, device.InboundObjectControl)

			device, err = azure.DeviceEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, deviceObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, device.InboundObjectControl)
		})
	})

	t.Run("GroupEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			groupObjectID, err := harness.AZEntityPanelHarness.Group.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", groupObjectID)

			group, err := azure.GroupEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, groupObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.Group.Properties.Get(common.ObjectID.String()).Any(), group.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, group.InboundObjectControl)

			group, err = azure.GroupEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, groupObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, group.InboundObjectControl)
		})
	})

	t.Run("ManagementGroupEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			groupObjectID, err := harness.AZEntityPanelHarness.ManagementGroup.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", groupObjectID)

			group, err := azure.ManagementGroupEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, groupObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.ManagementGroup.Properties.Get(common.ObjectID.String()).Any(), group.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, group.InboundObjectControl)

			group, err = azure.ManagementGroupEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, groupObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, group.InboundObjectControl)
		})
	})

	t.Run("ResourceGroupEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			groupObjectID, err := harness.AZEntityPanelHarness.ResourceGroup.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", groupObjectID)

			group, err := azure.ResourceGroupEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, groupObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.ResourceGroup.Properties.Get(common.ObjectID.String()).Any(), group.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, group.InboundObjectControl)

			group, err = azure.ResourceGroupEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, groupObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, group.InboundObjectControl)
		})
	})

	t.Run("KeyVaultEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			keyVaultObjectID, err := harness.AZEntityPanelHarness.KeyVault.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", keyVaultObjectID)

			keyVault, err := azure.KeyVaultEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, keyVaultObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.KeyVault.Properties.Get(common.ObjectID.String()).Any(), keyVault.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, keyVault.InboundObjectControl)

			keyVault, err = azure.KeyVaultEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, keyVaultObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, keyVault.InboundObjectControl)
		})
	})

	t.Run("RoleEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			roleObjectID, err := harness.AZEntityPanelHarness.Role.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", roleObjectID)

			role, err := azure.RoleEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, roleObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.Role.Properties.Get(common.ObjectID.String()).Any(), role.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, role.ActiveAssignments)

			role, err = azure.RoleEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, roleObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, role.ActiveAssignments)
		})
	})
	t.Run("RoleAddSecret", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZAddSecretHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			postProcessingStats, err := azure.AppRoleAssignments(context.Background(), testContext.Graph.Database)
			assert.Nil(t, err)
			assert.NotNil(t, postProcessingStats.RelationshipsCreated[graphAzure.AddSecret])
			assert.Equal(t, 4, int(*postProcessingStats.RelationshipsCreated[graphAzure.AddSecret]))

			// Validate that the AZAddSecret edges were created
			addSecretEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), graphAzure.AddSecret)
			}))
			assert.Nil(t, err)
			assert.Len(t, addSecretEdges, 4)

			for _, edge := range addSecretEdges {
				assert.Equal(t, graphAzure.AddSecret, edge.Kind)
				assert.True(t, slices.Contains([]graph.ID{harness.AZAddSecretHarness.AppAdminRole.ID, harness.AZAddSecretHarness.CloudAppAdminRole.ID}, edge.StartID))
				assert.True(t, slices.Contains([]graph.ID{harness.AZAddSecretHarness.AZApp.ID, harness.AZAddSecretHarness.AZServicePrincipal.ID}, edge.EndID))
			}
		})
	})
	t.Run("ServicePrincipalEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			servicePrincipalObjectID, err := harness.AZEntityPanelHarness.ServicePrincipal.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", servicePrincipalObjectID)

			servicePrincipal, err := azure.ServicePrincipalEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, servicePrincipalObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.ServicePrincipal.Properties.Get(common.ObjectID.String()).Any(), servicePrincipal.Properties[common.ObjectID.String()])
			assert.Equal(t, harness.AZEntityPanelHarness.Application.Properties.Get(common.ObjectID.String()).Any(), servicePrincipal.Properties[graphAzure.AppID.String()])
			assert.Equal(t, 0, servicePrincipal.InboundObjectControl)

			servicePrincipal, err = azure.ServicePrincipalEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, servicePrincipalObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, servicePrincipal.InboundObjectControl)
		})
	})
	t.Run("SubscriptionEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			subscriptionObjectID, err := harness.AZEntityPanelHarness.Subscription.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", subscriptionObjectID)

			subscription, err := azure.SubscriptionEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, subscriptionObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.Subscription.Properties.Get(common.ObjectID.String()).Any(), subscription.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, subscription.InboundObjectControl)

			subscription, err = azure.SubscriptionEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, subscriptionObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, subscription.InboundObjectControl)
		})
	})
	t.Run("TenantEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			tenantObjectID, err := harness.AZEntityPanelHarness.Tenant.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", tenantObjectID)

			tenant, err := azure.TenantEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, tenantObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.Tenant.Properties.Get(common.ObjectID.String()).Any(), tenant.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, tenant.InboundObjectControl)

			tenant, err = azure.TenantEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, tenantObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, tenant.InboundObjectControl)
		})
	})
	t.Run("UserEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			userObjectID, err := harness.AZEntityPanelHarness.User.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", userObjectID)

			user, err := azure.UserEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, userObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.User.Properties.Get(common.ObjectID.String()).Any(), user.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, user.OutboundObjectControl)

			user, err = azure.UserEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, userObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, user.OutboundObjectControl)
		})
	})
	t.Run("VMEntityDetails", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZEntityPanelHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {

			vmObjectID, err := harness.AZEntityPanelHarness.VM.Properties.Get(common.ObjectID.String()).String()
			require.Nil(t, err)
			assert.NotEqual(t, "", vmObjectID)

			vm, err := azure.VMEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, vmObjectID, false)

			require.Nil(t, err)
			assert.Equal(t, harness.AZEntityPanelHarness.VM.Properties.Get(common.ObjectID.String()).Any(), vm.Properties[common.ObjectID.String()])
			assert.Equal(t, 0, vm.InboundObjectControl)

			vm, err = azure.VMEntityDetails(testContext.Context(), testContext.Graph.Database, primaryDisplayKinds, vmObjectID, true)

			require.Nil(t, err)
			assert.NotEqual(t, 0, vm.InboundObjectControl)
		})
	})
	t.Run("FetchInboundEntityObjectControlPaths", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZInboundControlHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {
			paths, err := azure.FetchInboundEntityObjectControlPaths(tx, harness.AZInboundControlHarness.ControlledAZUser)
			require.Nil(t, err)
			nodes := paths.AllNodes().IDs()
			require.Equal(t, 8, len(nodes))
			require.NotContains(t, nodes, harness.AZInboundControlHarness.AZAppA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.ControlledAZUser.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupB.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalB.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserB.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZTenant.ID)
		})
	})
	t.Run("FetchInboundEntityObjectControllers", func(t *testing.T) {
		testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.AZInboundControlHarness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, tx graph.Transaction) {
			control, err := azure.FetchInboundEntityObjectControllers(tx, harness.AZInboundControlHarness.ControlledAZUser, 0, 0)
			require.Nil(t, err)
			nodes := control.IDs()
			require.Equal(t, 7, len(nodes))
			require.NotContains(t, nodes, harness.AZInboundControlHarness.ControlledAZUser.ID)
			require.NotContains(t, nodes, harness.AZInboundControlHarness.AZAppA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupB.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalB.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserA.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserB.ID)
			require.Contains(t, nodes, harness.AZInboundControlHarness.AZTenant.ID)
			control, err = azure.FetchInboundEntityObjectControllers(tx, harness.AZInboundControlHarness.ControlledAZUser, 0, 1)
			require.Nil(t, err)
			require.Equal(t, 1, control.Len())
		})
	})
}
