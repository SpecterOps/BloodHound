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

package azure_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	azschema "github.com/specterops/bloodhound/graphschema/azure"
)

var (
	user  = graph.NewNode(0, graph.NewProperties(), azschema.User)
	user2 = graph.NewNode(1, graph.NewProperties(), azschema.User)
	group = graph.NewNode(2, graph.NewProperties(), azschema.Group)
	app   = graph.NewNode(3, graph.NewProperties(), azschema.App)
)

// setupRoleAssignments is used to create a testable RoleAssignments struct. It is used in all RoleAssignments tests
// and may require adjusting tests if modified
func setupRoleAssignments() azure.RoleAssignments {
	return azure.RoleAssignments{
		// user2 has no roles! this is intentional
		Nodes: graph.NewNodeSet(user, user2, group, app).KindSet(),
		AssignmentMap: map[graph.ID]map[string]struct{}{
			user.ID:  {azschema.CompanyAdministratorRole: struct{}{}},
			group.ID: {azschema.ReportsReaderRole: struct{}{}, azschema.HelpdeskAdministratorRole: struct{}{}},
			app.ID:   {azschema.PartnerTier1SupportRole: struct{}{}},
		},
	}
}

func TestRoleAssignments_NodeHasRole(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.True(t, assignments.NodeHasRole(user.ID, azschema.CompanyAdministratorRole))
	assert.False(t, assignments.NodeHasRole(user.ID, azschema.HelpdeskAdministratorRole))
	assert.True(t, assignments.NodeHasRole(group.ID, azschema.ReportsReaderRole))
	assert.True(t, assignments.NodeHasRole(group.ID, azschema.HelpdeskAdministratorRole))
	assert.False(t, assignments.NodeHasRole(group.ID, azschema.PartnerTier1SupportRole))
}

func TestRoleAssignments_UsersWithoutRoles(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.NotEqual(t, user, assignments.UsersWithoutRoles().Get(user.ID))
	assert.Equal(t, user2, assignments.UsersWithoutRoles().Get(user2.ID))
}

func TestRoleAssignments_NodesWithRole(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.Equal(t, user, assignments.NodesWithRole(azschema.ReportsReaderRole, azschema.CompanyAdministratorRole).Get(azschema.User).Get(user.ID))
	assert.Equal(t, group, assignments.NodesWithRole(azschema.ReportsReaderRole, azschema.CompanyAdministratorRole).Get(azschema.Group).Get(group.ID))
	assert.Equal(t, group, assignments.NodesWithRole(azschema.ReportsReaderRole, azschema.HelpdeskAdministratorRole).Get(azschema.Group).Get(group.ID))
	assert.Equal(t, graph.EmptyNodeSet().Get(0), assignments.NodesWithRole(azschema.ReportsReaderRole).Get(azschema.User).Get(user.ID))
}

func TestRoleAssignments_NodesWithRolesExclusive(t *testing.T) {
	assignments := setupRoleAssignments()
	assert.Equal(t, user, assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole, azschema.CompanyAdministratorRole).Get(azschema.User).Get(user.ID))
	assert.Equal(t, graph.EmptyNodeSet().Get(0), assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole, azschema.CompanyAdministratorRole).Get(azschema.Group).Get(group.ID))
	assert.Equal(t, group, assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole, azschema.HelpdeskAdministratorRole).Get(azschema.Group).Get(group.ID))
	assert.Equal(t, graph.EmptyNodeSet().Get(0), assignments.NodesWithRolesExclusive(azschema.ReportsReaderRole).Get(azschema.User).Get(user.ID))
}
