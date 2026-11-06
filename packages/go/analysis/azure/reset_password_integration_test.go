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
	"slices"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	graphAzure "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRoleAssignmentsResetPasswordTargetRoles(t *testing.T) {
	var (
		suite               = setupIntegrationTestSuite(t)
		tenantID            = integration.RandomObjectID(t)
		tenant              = NewAzureTenant(t, &suite, tenantID)
		userAdministrator   = NewAzureRole(t, &suite, "User Administrator", integration.RandomObjectID(t), graphAzure.UserAccountAdministratorRole, tenantID)
		globalAdministrator = NewAzureRole(t, &suite, "Global Administrator", integration.RandomObjectID(t), graphAzure.CompanyAdministratorRole, tenantID)
		reportsReader       = NewAzureRole(t, &suite, "Reports Reader", integration.RandomObjectID(t), graphAzure.ReportsReaderRole, tenantID)
		roleAssignableGroup = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:                   "Role Assignable Group",
			common.ObjectID:               integration.RandomObjectID(t),
			graphAzure.TenantID:           tenantID,
			graphAzure.IsAssignableToRole: true,
		}), graphAzure.Entity, graphAzure.Group)
		roles = map[string]*graph.Node{
			graphAzure.CompanyAdministratorRole: globalAdministrator,
			graphAzure.ReportsReaderRole:        reportsReader,
		}
		cases = []struct {
			name                      string
			activeRoles               []string
			eligibleRoles             []string
			roleAssignableGroupMember bool
			userAdministratorCanReset bool
		}{
			{name: "no target roles", userAdministratorCanReset: true},
			{name: "eligible Global Administrator", eligibleRoles: []string{graphAzure.CompanyAdministratorRole}},
			{name: "active Global Administrator", activeRoles: []string{graphAzure.CompanyAdministratorRole}},
			{name: "eligible allowed role", eligibleRoles: []string{graphAzure.ReportsReaderRole}, userAdministratorCanReset: true},
			{name: "active allowed and eligible protected", activeRoles: []string{graphAzure.ReportsReaderRole}, eligibleRoles: []string{graphAzure.CompanyAdministratorRole}},
			{name: "active protected and eligible allowed", activeRoles: []string{graphAzure.CompanyAdministratorRole}, eligibleRoles: []string{graphAzure.ReportsReaderRole}},
			{name: "active and eligible protected role", activeRoles: []string{graphAzure.CompanyAdministratorRole}, eligibleRoles: []string{graphAzure.CompanyAdministratorRole}},
			{name: "active and eligible allowed role", activeRoles: []string{graphAzure.ReportsReaderRole}, eligibleRoles: []string{graphAzure.ReportsReaderRole}, userAdministratorCanReset: true},
			{name: "role assignable group member", roleAssignableGroupMember: true},
			{name: "allowed role and role assignable group member", eligibleRoles: []string{graphAzure.ReportsReaderRole}, roleAssignableGroupMember: true},
		}
		targets = make(map[string]*graph.Node, len(cases))
	)
	defer teardownIntegrationTestSuite(t, &suite)

	NewRelationship(t, &suite, tenant, userAdministrator, graphAzure.Contains)
	NewRelationship(t, &suite, tenant, globalAdministrator, graphAzure.Contains)
	NewRelationship(t, &suite, tenant, reportsReader, graphAzure.Contains)
	NewRelationship(t, &suite, tenant, roleAssignableGroup, graphAzure.Contains)

	for _, testCase := range cases {
		target := NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:         testCase.name,
			common.ObjectID:     integration.RandomObjectID(t),
			graphAzure.TenantID: tenantID,
		}), graphAzure.Entity, graphAzure.User)
		targets[testCase.name] = target
		NewRelationship(t, &suite, tenant, target, graphAzure.Contains)

		for _, roleTemplateID := range testCase.activeRoles {
			NewRelationship(t, &suite, target, roles[roleTemplateID], graphAzure.HasRole)
		}
		for _, roleTemplateID := range testCase.eligibleRoles {
			NewRelationship(t, &suite, target, roles[roleTemplateID], graphAzure.AZRoleEligible)
		}
		if testCase.roleAssignableGroupMember {
			NewRelationship(t, &suite, target, roleAssignableGroup, graphAzure.MemberOf)
		}
	}

	// Verify that analysis removes an invalid relationship generated by an earlier run.
	NewRelationship(t, &suite, userAdministrator, targets["eligible Global Administrator"], graphAzure.ResetPassword)

	_, err := azure.UserRoleAssignments(suite.Context, suite.GraphDB)
	require.NoError(t, err)

	var (
		resetEdgeCounts       map[graph.ID]map[graph.ID]int
		globalAdminEdgeCounts map[graph.ID]int
	)
	err = suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		resetEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), graphAzure.ResetPassword)
		}))
		if err != nil {
			return err
		}

		resetEdgeCounts = map[graph.ID]map[graph.ID]int{}
		for _, edge := range resetEdges {
			if resetEdgeCounts[edge.StartID] == nil {
				resetEdgeCounts[edge.StartID] = map[graph.ID]int{}
			}
			resetEdgeCounts[edge.StartID][edge.EndID]++
		}

		globalAdminEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), graphAzure.GlobalAdmin)
		}))
		if err != nil {
			return err
		}
		globalAdminEdgeCounts = map[graph.ID]int{}
		for _, edge := range globalAdminEdges {
			if edge.EndID == tenant.ID {
				globalAdminEdgeCounts[edge.StartID]++
			}
		}
		return nil
	})
	require.NoError(t, err)

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				expectedUserAdministratorEdges int
				expectedGlobalAdminEdges       int
			)
			if testCase.userAdministratorCanReset {
				expectedUserAdministratorEdges = 1
			}
			assert.Equal(t, expectedUserAdministratorEdges, resetEdgeCounts[userAdministrator.ID][targets[testCase.name].ID], "User Administrator reset edge")
			assert.Equal(t, 1, resetEdgeCounts[globalAdministrator.ID][targets[testCase.name].ID], "Global Administrator reset edge")
			if slices.Contains(testCase.activeRoles, graphAzure.CompanyAdministratorRole) {
				expectedGlobalAdminEdges = 1
			}
			assert.Equal(t, expectedGlobalAdminEdges, globalAdminEdgeCounts[targets[testCase.name].ID], "Global Administrator privilege edge")
		})
	}
}

func TestFetchRoleAssignableGroupMembersUsersPropertyTypes(t *testing.T) {
	var (
		suite    = setupIntegrationTestSuite(t)
		tenantID = integration.RandomObjectID(t)
		cases    = []struct {
			name               string
			isAssignableToRole any
			wantMember         bool
		}{
			{name: "boolean true", isAssignableToRole: true, wantMember: true},
			{name: "string true", isAssignableToRole: "true", wantMember: true},
			{name: "boolean false", isAssignableToRole: false},
		}
	)
	defer teardownIntegrationTestSuite(t, &suite)

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				group = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
					common.Name:                   testCase.name,
					common.ObjectID:               integration.RandomObjectID(t),
					graphAzure.TenantID:           tenantID,
					graphAzure.IsAssignableToRole: testCase.isAssignableToRole,
				}), graphAzure.Entity, graphAzure.Group)
				user = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
					common.Name:         "Group member",
					common.ObjectID:     integration.RandomObjectID(t),
					graphAzure.TenantID: tenantID,
				}), graphAzure.Entity, graphAzure.User)
			)
			NewRelationship(t, &suite, user, group, graphAzure.MemberOf)

			err := suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
				members, err := azure.FetchRoleAssignableGroupMembersUsers(tx, group, 0, 0)
				if err != nil {
					return err
				}
				assert.Equal(t, testCase.wantMember, members.Contains(user))
				return nil
			})
			require.NoError(t, err)
		})
	}
}
