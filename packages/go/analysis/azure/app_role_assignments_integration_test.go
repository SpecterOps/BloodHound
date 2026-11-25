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

package azure_test

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	graphazure "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAppRoleAssignmentsSourceServicePrincipals verifies the MS Graph permission edges produced by App Role Assignment
// post-processing. Source service principals are resolved for every tenant service principal at once, so this covers
// sources that hold a permission against several tenant service principals and sources whose permission is granted
// against a service principal in a different tenant.
func TestAppRoleAssignmentsSourceServicePrincipals(t *testing.T) {
	type relationshipKey struct {
		from graph.ID
		to   graph.ID
		kind graph.Kind
	}

	var (
		suite                    = setupIntegrationTestSuite(t)
		tenantID                 = randomObjectID(t)
		otherTenantID            = randomObjectID(t)
		tenant                   = NewAzureTenant(t, &suite, tenantID)
		otherTenant              = NewAzureTenant(t, &suite, otherTenantID)
		microsoftGraph           = NewAzureServicePrincipal(t, &suite, "Microsoft Graph", randomObjectID(t), tenantID)
		resourceServicePrincipal = NewAzureServicePrincipal(t, &suite, "Resource service principal", randomObjectID(t), tenantID)
		application              = NewAzureApplication(t, &suite, "Application", randomObjectID(t), tenantID)
		otherTenantResource      = NewAzureServicePrincipal(t, &suite, "Other tenant resource", randomObjectID(t), otherTenantID)
		newGroup                 = func(name string, tenantObjectID string, isAssignableToRole bool) *graph.Node {
			return NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
				common.Name:                   name,
				common.ObjectID:               randomObjectID(t),
				graphazure.TenantID:           tenantObjectID,
				graphazure.IsAssignableToRole: isAssignableToRole,
			}), graphazure.Entity, graphazure.Group)
		}
		group                     = newGroup("Group", tenantID, false)
		roleAssignableGroup       = newGroup("Role assignable group", tenantID, true)
		otherTenantGroup          = newGroup("Other tenant group", otherTenantID, false)
		newSourceServicePrincipal = func(name string) *graph.Node {
			// Source service principals are intentionally not contained by either tenant so that they are not also
			// targets of the edges being verified.
			return NewAzureServicePrincipal(t, &suite, name, randomObjectID(t), randomObjectID(t))
		}
		applicationReadWriteAllSource       = newSourceServicePrincipal("Application.ReadWrite.All source")
		appRoleAssignmentReadWriteAllSource = newSourceServicePrincipal("AppRoleAssignment.ReadWrite.All source")
		directoryReadWriteAllSource         = newSourceServicePrincipal("Directory.ReadWrite.All source")
		groupReadWriteAllSource             = newSourceServicePrincipal("Group.ReadWrite.All source")
		groupMemberReadWriteAllSource       = newSourceServicePrincipal("GroupMember.ReadWrite.All source")
		endpointReadWriteAllSource          = newSourceServicePrincipal("ServicePrincipalEndpoint.ReadWrite.All source")
		otherTenantSource                   = newSourceServicePrincipal("Other tenant Directory.ReadWrite.All source")
	)
	defer teardownIntegrationTestSuite(t, &suite)

	for _, containedNode := range []*graph.Node{microsoftGraph, resourceServicePrincipal, application, group, roleAssignableGroup} {
		NewRelationship(t, &suite, tenant, containedNode, graphazure.Contains)
	}

	for _, containedNode := range []*graph.Node{otherTenantResource, otherTenantGroup} {
		NewRelationship(t, &suite, otherTenant, containedNode, graphazure.Contains)
	}

	// Granted against two tenant service principals; the source must only be resolved once.
	NewRelationship(t, &suite, applicationReadWriteAllSource, microsoftGraph, graphazure.ApplicationReadWriteAll)
	NewRelationship(t, &suite, applicationReadWriteAllSource, resourceServicePrincipal, graphazure.ApplicationReadWriteAll)

	NewRelationship(t, &suite, appRoleAssignmentReadWriteAllSource, microsoftGraph, graphazure.AppRoleAssignmentReadWriteAll)
	NewRelationship(t, &suite, directoryReadWriteAllSource, resourceServicePrincipal, graphazure.DirectoryReadWriteAll)
	NewRelationship(t, &suite, groupReadWriteAllSource, microsoftGraph, graphazure.GroupReadWriteAll)
	NewRelationship(t, &suite, groupMemberReadWriteAllSource, microsoftGraph, graphazure.GroupMemberReadWriteAll)
	NewRelationship(t, &suite, endpointReadWriteAllSource, microsoftGraph, graphazure.ServicePrincipalEndpointReadWriteAll)

	// Granted only against the other tenant's service principal; must only produce edges within the other tenant.
	NewRelationship(t, &suite, otherTenantSource, otherTenantResource, graphazure.DirectoryReadWriteAll)

	_, err := azure.AppRoleAssignments(suite.Context, suite.GraphDB)
	require.NoError(t, err)

	var (
		actual   []relationshipKey
		expected = []relationshipKey{
			// Application.ReadWrite.All: every tenant service principal and application
			{from: applicationReadWriteAllSource.ID, to: microsoftGraph.ID, kind: graphazure.AZMGAddSecret},
			{from: applicationReadWriteAllSource.ID, to: microsoftGraph.ID, kind: graphazure.AZMGAddOwner},
			{from: applicationReadWriteAllSource.ID, to: resourceServicePrincipal.ID, kind: graphazure.AZMGAddSecret},
			{from: applicationReadWriteAllSource.ID, to: resourceServicePrincipal.ID, kind: graphazure.AZMGAddOwner},
			{from: applicationReadWriteAllSource.ID, to: application.ID, kind: graphazure.AZMGAddSecret},
			{from: applicationReadWriteAllSource.ID, to: application.ID, kind: graphazure.AZMGAddOwner},

			// AppRoleAssignment.ReadWrite.All: the tenant
			{from: appRoleAssignmentReadWriteAllSource.ID, to: tenant.ID, kind: graphazure.AZMGGrantAppRoles},

			// Directory.ReadWrite.All and Group.ReadWrite.All: non-role-assignable tenant groups
			{from: directoryReadWriteAllSource.ID, to: group.ID, kind: graphazure.AZMGAddMember},
			{from: directoryReadWriteAllSource.ID, to: group.ID, kind: graphazure.AZMGAddOwner},
			{from: groupReadWriteAllSource.ID, to: group.ID, kind: graphazure.AZMGAddMember},
			{from: groupReadWriteAllSource.ID, to: group.ID, kind: graphazure.AZMGAddOwner},

			// GroupMember.ReadWrite.All: non-role-assignable tenant groups
			{from: groupMemberReadWriteAllSource.ID, to: group.ID, kind: graphazure.AZMGAddMember},

			// ServicePrincipalEndpoint.ReadWrite.All: every tenant service principal
			{from: endpointReadWriteAllSource.ID, to: microsoftGraph.ID, kind: graphazure.AZMGAddOwner},
			{from: endpointReadWriteAllSource.ID, to: resourceServicePrincipal.ID, kind: graphazure.AZMGAddOwner},

			// Other tenant Directory.ReadWrite.All: only the other tenant's groups
			{from: otherTenantSource.ID, to: otherTenantGroup.ID, kind: graphazure.AZMGAddMember},
			{from: otherTenantSource.ID, to: otherTenantGroup.ID, kind: graphazure.AZMGAddOwner},
		}
	)

	require.NoError(t, suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		relationships, err := ops.FetchRelationships(tx.Relationships().Filter(query.KindIn(
			query.Relationship(), graphazure.AZMGGrantRole, graphazure.AZMGGrantAppRoles,
			graphazure.AZMGAddMember, graphazure.AZMGAddSecret, graphazure.AZMGAddOwner,
		)))
		if err != nil {
			return err
		}

		for _, relationship := range relationships {
			actual = append(actual, relationshipKey{from: relationship.StartID, to: relationship.EndID, kind: relationship.Kind})
		}

		return nil
	}))

	assert.ElementsMatch(t, expected, actual)
}
