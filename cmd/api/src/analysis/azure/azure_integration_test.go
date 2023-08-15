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
	schema "github.com/specterops/bloodhound/graphschema"
	"sort"
	"testing"

	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	azureanalysis "github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/src/test/integration"
)

func SortIDs(ids []graph.ID) []graph.ID {
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Int64() > ids[j].Int64()
	})

	return ids
}

func TestFetchEntityByObjectID(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		node, err := azureanalysis.FetchEntityByObjectID(tx, testContext.NodeObjectID(harness.AZBaseHarness.Application))

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
		roles, err := azureanalysis.FetchEntityRoles(tx, harness.AZBaseHarness.User, 0, 0)

		require.Nil(t, err)
		assert.ElementsMatch(t, harness.AZBaseHarness.Nodes.Get(azure.Role).IDs(), roles.ContainingNodeKinds(azure.Role).IDs())
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
				return query.Kind(query.Relationship(), azure.HasRole)
			},
		}); err != nil {
			t.Fatal(err)
		} else {
			numRoles := harness.AZBaseHarness.Nodes.Count(azure.Role)
			assert.Equal(t, harness.AZBaseHarness.Nodes.Get(azure.Role).Len(), paths.Len())

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
		if roles, err := azureanalysis.FetchEntityRoles(tx, harness.AZBaseHarness.User, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.ElementsMatch(t, harness.AZBaseHarness.Nodes.Get(azure.Role).IDs(), roles.IDs())
		}
	})
}

func TestAzureEntityGroupMembership(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZBaseHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		if groupPaths, err := azureanalysis.FetchEntityGroupMembershipPaths(tx, harness.AZBaseHarness.User); err != nil {
			t.Fatal(err)
		} else {
			assert.ElementsMatch(t, harness.AZBaseHarness.UserFirstDegreeGroups.IDs(), groupPaths.AllNodes().ContainingNodeKinds(azure.Group).IDs())
		}
	})
}

func TestAZMGApplicationReadWriteAll(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZMGApplicationReadWriteAllHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		if outboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph))
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.Application))
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipalB))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.Application, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipalB, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azureanalysis.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGApplicationReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{azure.AZMGAddOwner, azure.AZMGAddSecret}, []graph.Kind{edge.Kind})
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

		if outboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.Tenant))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.Tenant, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azureanalysis.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGAppRoleManagementReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Equal(t, azure.AZMGGrantAppRoles, edge.Kind)
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

		if outboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGDirectoryReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGDirectoryReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGDirectoryReadWriteAllHarness.Group))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGDirectoryReadWriteAllHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azureanalysis.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGDirectoryReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{azure.AZMGAddOwner, azure.AZMGAddMember}, []graph.Kind{edge.Kind})
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

		if outboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupReadWriteAllHarness.Group))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupReadWriteAllHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azureanalysis.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGGroupReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{azure.AZMGAddOwner, azure.AZMGAddMember}, []graph.Kind{edge.Kind})
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

		if outboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupMemberReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGGroupMemberReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupMemberReadWriteAllHarness.Group))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGGroupMemberReadWriteAllHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azureanalysis.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGGroupMemberReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{azure.AZMGAddOwner, azure.AZMGAddMember}, []graph.Kind{edge.Kind})
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

		if outboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.Group))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.Application))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipalB))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.Role))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.MicrosoftGraph))
		}

		if inboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.Group, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azureanalysis.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGRoleManagementReadWriteDirectoryHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Subset(t, []graph.Kind{azure.AZMGAddOwner, azure.AZMGAddSecret, azure.AZMGGrantRole}, []graph.Kind{edge.Kind})
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

		if outboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAbusableAppRoleAssignments.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.MicrosoftGraph))
		}

		if inboundAbusableAppRoleAssignments, err := azureanalysis.FetchAbusableAppRoleAssignments(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.MicrosoftGraph, graph.DirectionInbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, inboundAbusableAppRoleAssignments.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipalB))
		}

		if outboundAppRoleAssignmentTransitNodes, err := azureanalysis.FetchAppRoleAssignmentsTransitList(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, outboundAppRoleAssignmentTransitNodes.Contains(harness.AZMGServicePrincipalEndpointReadWriteAllHarness.MicrosoftGraph))
		}

		if outboundAppRoleAssignmentTransitPaths, err := azureanalysis.FetchAppRoleAssignmentsTransitPaths(tx, harness.AZMGServicePrincipalEndpointReadWriteAllHarness.ServicePrincipal, graph.DirectionOutbound); err != nil {
			t.Fatal(err)
		} else {
			for _, path := range outboundAppRoleAssignmentTransitPaths.Paths() {
				for _, edge := range path.Edges {
					assert.Equal(t, azure.AZMGAddOwner, edge.Kind)
				}
			}
		}

	})
}

/**********************
 * Entity Panel tests *
 **********************/

func TestApplicationEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		appObjectID, err := harness.AZEntityPanelHarness.Application.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", appObjectID)

		app, err := azureanalysis.ApplicationEntityDetails(context.Background(), testContext.Graph.Database, appObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.Application.Properties.Get(common.ObjectID.String()).Any(), app.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, app.InboundObjectControl)

		app, err = azureanalysis.ApplicationEntityDetails(context.Background(), testContext.Graph.Database, appObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, app.InboundObjectControl)
	})
}

func TestDeviceEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		deviceObjectID, err := harness.AZEntityPanelHarness.Device.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", deviceObjectID)

		device, err := azureanalysis.DeviceEntityDetails(context.Background(), testContext.Graph.Database, deviceObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.Device.Properties.Get(common.ObjectID.String()).Any(), device.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, device.InboundObjectControl)

		device, err = azureanalysis.DeviceEntityDetails(context.Background(), testContext.Graph.Database, deviceObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, device.InboundObjectControl)
	})
}

func TestGroupEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		groupObjectID, err := harness.AZEntityPanelHarness.Group.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", groupObjectID)

		group, err := azureanalysis.GroupEntityDetails(testContext.Graph.Database, groupObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.Group.Properties.Get(common.ObjectID.String()).Any(), group.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, group.InboundObjectControl)

		group, err = azureanalysis.GroupEntityDetails(testContext.Graph.Database, groupObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, group.InboundObjectControl)
	})
}

func TestManagementGroupEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		groupObjectID, err := harness.AZEntityPanelHarness.ManagementGroup.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", groupObjectID)

		group, err := azureanalysis.ManagementGroupEntityDetails(context.Background(), testContext.Graph.Database, groupObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.ManagementGroup.Properties.Get(common.ObjectID.String()).Any(), group.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, group.InboundObjectControl)

		group, err = azureanalysis.ManagementGroupEntityDetails(context.Background(), testContext.Graph.Database, groupObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, group.InboundObjectControl)
	})
}

func TestResourceGroupEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		groupObjectID, err := harness.AZEntityPanelHarness.ResourceGroup.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", groupObjectID)

		group, err := azureanalysis.ResourceGroupEntityDetails(context.Background(), testContext.Graph.Database, groupObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.ResourceGroup.Properties.Get(common.ObjectID.String()).Any(), group.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, group.InboundObjectControl)

		group, err = azureanalysis.ResourceGroupEntityDetails(context.Background(), testContext.Graph.Database, groupObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, group.InboundObjectControl)
	})
}

func TestKeyVaultEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		keyVaultObjectID, err := harness.AZEntityPanelHarness.KeyVault.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", keyVaultObjectID)

		keyVault, err := azureanalysis.KeyVaultEntityDetails(context.Background(), testContext.Graph.Database, keyVaultObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.KeyVault.Properties.Get(common.ObjectID.String()).Any(), keyVault.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, keyVault.InboundObjectControl)

		keyVault, err = azureanalysis.KeyVaultEntityDetails(context.Background(), testContext.Graph.Database, keyVaultObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, keyVault.InboundObjectControl)
	})
}

func TestRoleEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		roleObjectID, err := harness.AZEntityPanelHarness.Role.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", roleObjectID)

		role, err := azureanalysis.RoleEntityDetails(context.Background(), testContext.Graph.Database, roleObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.Role.Properties.Get(common.ObjectID.String()).Any(), role.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, role.ActiveAssignments)

		role, err = azureanalysis.RoleEntityDetails(context.Background(), testContext.Graph.Database, roleObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, role.ActiveAssignments)
	})
}

func TestServicePrincipalEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		servicePrincipalObjectID, err := harness.AZEntityPanelHarness.ServicePrincipal.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", servicePrincipalObjectID)

		servicePrincipal, err := azureanalysis.ServicePrincipalEntityDetails(context.Background(), testContext.Graph.Database, servicePrincipalObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.ServicePrincipal.Properties.Get(common.ObjectID.String()).Any(), servicePrincipal.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, servicePrincipal.InboundObjectControl)

		servicePrincipal, err = azureanalysis.ServicePrincipalEntityDetails(context.Background(), testContext.Graph.Database, servicePrincipalObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, servicePrincipal.InboundObjectControl)
	})
}

func TestSubscriptionEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		subscriptionObjectID, err := harness.AZEntityPanelHarness.Subscription.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", subscriptionObjectID)

		subscription, err := azureanalysis.SubscriptionEntityDetails(context.Background(), testContext.Graph.Database, subscriptionObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.Subscription.Properties.Get(common.ObjectID.String()).Any(), subscription.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, subscription.InboundObjectControl)

		subscription, err = azureanalysis.SubscriptionEntityDetails(context.Background(), testContext.Graph.Database, subscriptionObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, subscription.InboundObjectControl)
	})
}

func TestTenantEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		tenantObjectID, err := harness.AZEntityPanelHarness.Tenant.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", tenantObjectID)

		tenant, err := azureanalysis.TenantEntityDetails(testContext.Graph.Database, tenantObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.Tenant.Properties.Get(common.ObjectID.String()).Any(), tenant.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, tenant.InboundObjectControl)

		tenant, err = azureanalysis.TenantEntityDetails(testContext.Graph.Database, tenantObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, tenant.InboundObjectControl)
	})
}

func TestUserEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		userObjectID, err := harness.AZEntityPanelHarness.User.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", userObjectID)

		user, err := azureanalysis.UserEntityDetails(testContext.Graph.Database, userObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.User.Properties.Get(common.ObjectID.String()).Any(), user.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, user.OutboundObjectControl)

		user, err = azureanalysis.UserEntityDetails(testContext.Graph.Database, userObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, user.OutboundObjectControl)
	})
}

func TestVMEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZEntityPanelHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {

		vmObjectID, err := harness.AZEntityPanelHarness.VM.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)
		assert.NotEqual(t, "", vmObjectID)

		vm, err := azureanalysis.VMEntityDetails(context.Background(), testContext.Graph.Database, vmObjectID, false)

		require.Nil(t, err)
		assert.Equal(t, harness.AZEntityPanelHarness.VM.Properties.Get(common.ObjectID.String()).Any(), vm.Properties[common.ObjectID.String()])
		assert.Equal(t, 0, vm.InboundObjectControl)

		vm, err = azureanalysis.VMEntityDetails(context.Background(), testContext.Graph.Database, vmObjectID, true)

		require.Nil(t, err)
		assert.NotEqual(t, 0, vm.InboundObjectControl)
	})
}

func TestFetchInboundEntityObjectControlPaths(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZInboundControlHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		paths, err := azureanalysis.FetchInboundEntityObjectControlPaths(tx, harness.AZInboundControlHarness.ControlledAZUser, graph.DirectionInbound)
		require.Nil(t, err)
		nodes := paths.AllNodes().IDs()
		require.Equal(t, 7, len(nodes))
		require.NotContains(t, nodes, harness.AZInboundControlHarness.AZAppA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.ControlledAZUser.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupB.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalB.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserB.ID)
	})
}

func TestFetchInboundEntityObjectControllers(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.ReadTransactionTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.AZInboundControlHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, tx graph.Transaction) {
		control, err := azureanalysis.FetchInboundEntityObjectControllers(tx, harness.AZInboundControlHarness.ControlledAZUser, graph.DirectionInbound, 0, 0)
		require.Nil(t, err)
		nodes := control.IDs()
		require.Equal(t, 6, len(nodes))
		require.NotContains(t, nodes, harness.AZInboundControlHarness.ControlledAZUser.ID)
		require.NotContains(t, nodes, harness.AZInboundControlHarness.AZAppA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZGroupB.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZServicePrincipalB.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserA.ID)
		require.Contains(t, nodes, harness.AZInboundControlHarness.AZUserB.ID)
		control, err = azureanalysis.FetchInboundEntityObjectControllers(tx, harness.AZInboundControlHarness.ControlledAZUser, graph.DirectionInbound, 0, 1)
		require.Nil(t, err)
		require.Equal(t, 1, control.Len())
	})
}
