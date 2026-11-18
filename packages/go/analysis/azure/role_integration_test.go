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
