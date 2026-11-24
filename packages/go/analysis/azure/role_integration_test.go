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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleTraversalRoleAssignableGroupPropertyTypes(t *testing.T) {
	var (
		suite    = setupIntegrationTestSuite(t)
		tenantID = randomObjectID(t)
		tenant   = NewAzureTenant(t, &suite, tenantID)
		role     = NewAzureRole(t, &suite, "Global Administrator", randomObjectID(t), graphazure.CompanyAdministratorRole, tenantID)
		cases    = []struct {
			name               string
			isAssignableToRole any
			nestedGroup        bool
			wantInheritedRole  bool
		}{
			{name: "boolean true", isAssignableToRole: true, wantInheritedRole: true},
			{name: "string true", isAssignableToRole: "true", wantInheritedRole: true},
			{name: "boolean false", isAssignableToRole: false},
			{name: "string false", isAssignableToRole: "false"},
			{name: "missing property"},
			{name: "invalid string", isAssignableToRole: "not-a-boolean"},
			{name: "numeric value", isAssignableToRole: 1},
			{name: "nested boolean true groups", isAssignableToRole: true, nestedGroup: true},
			{name: "nested string true groups", isAssignableToRole: "true", nestedGroup: true},
		}
	)
	defer teardownIntegrationTestSuite(t, &suite)

	NewRelationship(t, &suite, tenant, role, graphazure.Contains)

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				groupProperties = graph.PropertyMap{
					common.Name:         testCase.name,
					common.ObjectID:     randomObjectID(t),
					graphazure.TenantID: tenantID,
				}
				user = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
					common.Name:         "Group member",
					common.ObjectID:     randomObjectID(t),
					graphazure.TenantID: tenantID,
				}), graphazure.Entity, graphazure.User)
			)
			if testCase.isAssignableToRole != nil {
				groupProperties[graphazure.IsAssignableToRole] = testCase.isAssignableToRole
			}

			var group = NewNode(t, &suite, graph.AsProperties(groupProperties), graphazure.Entity, graphazure.Group)
			NewRelationship(t, &suite, tenant, group, graphazure.Contains)
			NewRelationship(t, &suite, tenant, user, graphazure.Contains)

			// The same group has active and eligible assignments. Each lookup below
			// traverses only its own assignment kind, exercising both directions.
			NewRelationship(t, &suite, group, role, graphazure.HasRole)
			NewRelationship(t, &suite, group, role, graphazure.AZRoleEligible)
			if testCase.nestedGroup {
				var nestedGroup = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
					common.Name:                   "Nested group",
					common.ObjectID:               randomObjectID(t),
					graphazure.TenantID:           tenantID,
					graphazure.IsAssignableToRole: testCase.isAssignableToRole,
				}), graphazure.Entity, graphazure.Group)
				NewRelationship(t, &suite, tenant, nestedGroup, graphazure.Contains)
				NewRelationship(t, &suite, user, nestedGroup, graphazure.MemberOf)
				NewRelationship(t, &suite, nestedGroup, group, graphazure.MemberOf)
			} else {
				NewRelationship(t, &suite, user, group, graphazure.MemberOf)
			}

			err := suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
				// RoleMembers walks inbound from the role toward its active assignees.
				members, err := azure.RoleMembers(tx, tenant, graphazure.CompanyAdministratorRole)
				if err != nil {
					return err
				}
				assert.Equal(t, testCase.wantInheritedRole, members.Contains(user), "active role inherited through group")

				// Eligible-role lookup walks outbound from the user. String-valued
				// true must work here too so role membership displays remain correct.
				eligibleRoles, err := azure.FetchEntityEligibleRoles(tx, user, 0, 0)
				if err != nil {
					return err
				}
				assert.Equal(t, testCase.wantInheritedRole, eligibleRoles.Contains(role), "eligible role inherited through group")
				return nil
			})
			require.NoError(t, err)
		})
	}
}

func TestFetchTenantRoleAssignmentsRoleMembers(t *testing.T) {
	var (
		suite               = setupIntegrationTestSuite(t)
		tenantID            = randomObjectID(t)
		otherTenantID       = randomObjectID(t)
		tenant              = NewAzureTenant(t, &suite, tenantID)
		otherTenant         = NewAzureTenant(t, &suite, otherTenantID)
		globalAdministrator = NewAzureRole(t, &suite, "Global Administrator", randomObjectID(t), graphazure.CompanyAdministratorRole, tenantID)
		userAdministrator   = NewAzureRole(t, &suite, "User Administrator", randomObjectID(t), graphazure.UserAccountAdministratorRole, tenantID)
		otherTenantRole     = NewAzureRole(t, &suite, "Global Administrator", randomObjectID(t), graphazure.CompanyAdministratorRole, otherTenantID)
		newTenantEntity     = func(name string, tenantObjectID string, properties graph.PropertyMap, kinds ...graph.Kind) *graph.Node {
			nodeProperties := graph.PropertyMap{
				common.Name:         name,
				common.ObjectID:     randomObjectID(t),
				graphazure.TenantID: tenantObjectID,
			}
			for key, value := range properties {
				nodeProperties[key] = value
			}

			return NewNode(t, &suite, graph.AsProperties(nodeProperties), append(graph.Kinds{graphazure.Entity}, kinds...)...)
		}
		directUser          = newTenantEntity("Direct user", tenantID, nil, graphazure.User)
		servicePrincipal    = newTenantEntity("Service principal", tenantID, nil, graphazure.ServicePrincipal)
		roleAssignableGroup = newTenantEntity("Role assignable group", tenantID, graph.PropertyMap{graphazure.IsAssignableToRole: true}, graphazure.Group)
		groupMemberUser     = newTenantEntity("Group member user", tenantID, nil, graphazure.User)
		userWithoutRoles    = newTenantEntity("User without roles", tenantID, nil, graphazure.User)
		otherTenantUser     = newTenantEntity("Other tenant user", otherTenantID, nil, graphazure.User)
	)
	defer teardownIntegrationTestSuite(t, &suite)

	for _, containedNode := range []*graph.Node{globalAdministrator, userAdministrator, directUser, servicePrincipal, roleAssignableGroup, groupMemberUser, userWithoutRoles} {
		NewRelationship(t, &suite, tenant, containedNode, graphazure.Contains)
	}

	NewRelationship(t, &suite, otherTenant, otherTenantRole, graphazure.Contains)
	NewRelationship(t, &suite, otherTenant, otherTenantUser, graphazure.Contains)

	NewRelationship(t, &suite, directUser, globalAdministrator, graphazure.HasRole)
	NewRelationship(t, &suite, servicePrincipal, globalAdministrator, graphazure.HasRole)
	NewRelationship(t, &suite, roleAssignableGroup, userAdministrator, graphazure.HasRole)
	NewRelationship(t, &suite, groupMemberUser, roleAssignableGroup, graphazure.MemberOf)
	NewRelationship(t, &suite, otherTenantUser, otherTenantRole, graphazure.HasRole)

	roleAssignments, err := azure.FetchTenantRoleAssignments(suite.Context, suite.GraphDB, tenant)
	require.NoError(t, err)

	globalAdministratorMembers := roleAssignments.RoleMap[graphazure.CompanyAdministratorRole]
	require.NotNil(t, globalAdministratorMembers)
	assert.True(t, globalAdministratorMembers.Contains(directUser.ID.Uint64()), "direct role assignment")
	assert.True(t, globalAdministratorMembers.Contains(servicePrincipal.ID.Uint64()), "service principal role assignment")
	assert.False(t, globalAdministratorMembers.Contains(globalAdministrator.ID.Uint64()), "role node must not be its own member")
	assert.False(t, globalAdministratorMembers.Contains(otherTenantUser.ID.Uint64()), "members of the same role template in another tenant")

	userAdministratorMembers := roleAssignments.RoleMap[graphazure.UserAccountAdministratorRole]
	require.NotNil(t, userAdministratorMembers)
	assert.True(t, userAdministratorMembers.Contains(roleAssignableGroup.ID.Uint64()), "group role assignment")
	assert.True(t, userAdministratorMembers.Contains(groupMemberUser.ID.Uint64()), "role inherited through role assignable group")
	assert.False(t, userAdministratorMembers.Contains(userAdministrator.ID.Uint64()), "role node must not be its own member")

	assert.False(t, roleAssignments.UsersWithAnyRole().Contains(userWithoutRoles.ID.Uint64()), "user without roles")
	assert.True(t, roleAssignments.UsersWithoutRoles().Contains(userWithoutRoles.ID.Uint64()), "user without roles")

	// The role map must match the per-template RoleMembers lookup that it replaces.
	require.NoError(t, suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		for _, roleTemplateID := range []string{graphazure.CompanyAdministratorRole, graphazure.UserAccountAdministratorRole} {
			expectedMembers, err := azure.RoleMembers(tx, tenant, roleTemplateID)
			if err != nil {
				return err
			}

			assert.ElementsMatch(t, expectedMembers.IDBitmap().Slice(), roleAssignments.RoleMap[roleTemplateID].Slice(), roleTemplateID)
		}

		return nil
	}))
}

func TestFetchTenantRoleAssignmentsRoleAssignableGroupMembership(t *testing.T) {
	type groupCase struct {
		name               string
		isAssignableToRole any
		wantMember         bool
		group              *graph.Node
		member             *graph.Node
	}

	var (
		suite    = setupIntegrationTestSuite(t)
		tenantID = randomObjectID(t)
		tenant   = NewAzureTenant(t, &suite, tenantID)
		cases    = []*groupCase{
			{name: "boolean true", isAssignableToRole: true, wantMember: true},
			{name: "string true", isAssignableToRole: "true", wantMember: true},
			{name: "boolean false", isAssignableToRole: false},
			{name: "string false", isAssignableToRole: "false"},
			{name: "missing property"},
			{name: "numeric value", isAssignableToRole: 1},
			{name: "uppercase string true", isAssignableToRole: "TRUE"},
			{name: "invalid string", isAssignableToRole: "not-a-boolean"},
		}
	)
	defer teardownIntegrationTestSuite(t, &suite)

	for _, testCase := range cases {
		groupProperties := graph.PropertyMap{
			common.Name:         testCase.name,
			common.ObjectID:     randomObjectID(t),
			graphazure.TenantID: tenantID,
		}
		if testCase.isAssignableToRole != nil {
			groupProperties[graphazure.IsAssignableToRole] = testCase.isAssignableToRole
		}

		testCase.group = NewNode(t, &suite, graph.AsProperties(groupProperties), graphazure.Entity, graphazure.Group)
		testCase.member = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:         testCase.name + " member",
			common.ObjectID:     randomObjectID(t),
			graphazure.TenantID: tenantID,
		}), graphazure.Entity, graphazure.User)

		NewRelationship(t, &suite, tenant, testCase.group, graphazure.Contains)
		NewRelationship(t, &suite, tenant, testCase.member, graphazure.Contains)
		NewRelationship(t, &suite, testCase.member, testCase.group, graphazure.MemberOf)
	}

	roleAssignments, err := azure.FetchTenantRoleAssignments(suite.Context, suite.GraphDB, tenant)
	require.NoError(t, err)

	roleAssignableGroupMembers := roleAssignments.UsersWithRoleAssignableGroupMembership()

	require.NoError(t, suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		for _, testCase := range cases {
			// Groups that are not role assignable are skipped before querying. The result must match querying every
			// group, which is what FetchTenantRoleAssignments previously did.
			queriedMembers, err := azure.FetchRoleAssignableGroupMembersUsers(tx, testCase.group, 0, 0)
			if err != nil {
				return err
			}

			assert.Equal(t, testCase.wantMember, queriedMembers.Contains(testCase.member), "%s: queried membership", testCase.name)
			assert.Equal(t, testCase.wantMember, roleAssignableGroupMembers.Contains(testCase.member.ID.Uint64()), "%s: tenant role assignments", testCase.name)
		}

		return nil
	}))
}
