// Copyright 2025 Specter Ops, Inc.
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

package ein_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bloodhoundad/azurehound/v2/models"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAzureRoleEligibilityScheduleInstanceToRel(t *testing.T) {
	testData := models.RoleEligibilityScheduleInstance{
		Id:               "lAPpYvVpN0KRkAEhdxReELKn6QMIlSROgkgWZy9fE3c-1-e",
		RoleDefinitionId: "62e90394-69f5-4237-9190-012177145e10",
		PrincipalId:      "03e9a7b2-9508-4e24-8248-16672f5f1377",
		DirectoryScopeId: "/",
		StartDateTime:    "2024-01-04T01:22:36.867Z",
		TenantId:         "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
	}

	expectedRels := ein.ConvertAzureRoleEligibilityScheduleInstanceToRel(testData)
	require.Len(t, expectedRels, 1)
	expectedRel := expectedRels[0]
	require.Equal(t, expectedRel.Target.Value, strings.ToUpper(fmt.Sprintf("%s@%s", testData.RoleDefinitionId, testData.TenantId)))
	require.Equal(t, expectedRel.RelType, azure.AZRoleEligible)
	require.Equal(t, expectedRel.Source.Value, strings.ToUpper(testData.PrincipalId))

	testData = models.RoleEligibilityScheduleInstance{
		Id:               "lAPpYvVpN0KRkAEhdxReELKn6QMIlSROgkgWZy9fE3c-1-e",
		RoleDefinitionId: "62e90394-69f5-4237-9190-012177145e10",
		PrincipalId:      "03e9a7b2-9508-4e24-8248-16672f5f1377",
		DirectoryScopeId: "abc123",
		StartDateTime:    "2024-01-04T01:22:36.867Z",
		TenantId:         "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
	}

	expectedRels = ein.ConvertAzureRoleEligibilityScheduleInstanceToRel(testData)
	require.Len(t, expectedRels, 0)
}

func Test_ConvertAzureRoleManagementPolicyAssignment(t *testing.T) {
	model := models.RoleManagementPolicyAssignment{
		Id:                                "id-1234",
		RoleDefinitionId:                  "role-1234",
		EndUserAssignmentRequiresApproval: true,
		EndUserAssignmentRequiresCAPAuthenticationContext: false,
		EndUserAssignmentUserApprovers: []string{
			"user-approver-1",
			"user-approver-2",
		},
		EndUserAssignmentGroupApprovers: []string{
			"group-approver-1",
			"group-approver-2",
		},
		EndUserAssignmentRequiresMFA:               false,
		EndUserAssignmentRequiresJustification:     false,
		EndUserAssignmentRequiresTicketInformation: false,
		TenantId: "tenant-1234",
	}

	t.Run("Create AZRole node and no relationships", func(t *testing.T) {
		model.EndUserAssignmentRequiresApproval = false

		node, rels := ein.ConvertAzureRoleManagementPolicyAssignment(model)

		// Assert created node properties
		assert.Equal(t, "ROLE-1234@TENANT-1234", node.ObjectID)
		assert.Equal(t, "AZRole", node.Labels[0].String())
		require.Len(t, node.PropertyMap[azure.EndUserAssignmentGroupApprovers.String()], 2)
		assert.Equal(t, []string{"GROUP-APPROVER-1", "GROUP-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentGroupApprovers.String()])
		assert.Equal(t, "TENANT-1234", node.PropertyMap[azure.TenantID.String()])
		assert.Equal(t, false, node.PropertyMap[azure.EndUserAssignmentRequiresApproval.String()])
		assert.Equal(t, []string{"USER-APPROVER-1", "USER-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()])
		require.Len(t, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()], 2)
		assert.Equal(t, []string{"USER-APPROVER-1", "USER-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()])

		require.Len(t, rels, 0)
	})

	t.Run("Create relationships and multiple ndoes for each user and group approver", func(t *testing.T) {
		model.EndUserAssignmentRequiresApproval = true

		node, rels := ein.ConvertAzureRoleManagementPolicyAssignment(model)

		// Assert created node properties
		assert.Equal(t, "ROLE-1234@TENANT-1234", node.ObjectID)
		assert.Equal(t, "AZRole", node.Labels[0].String())
		require.Len(t, node.PropertyMap[azure.EndUserAssignmentGroupApprovers.String()], 2)
		assert.Equal(t, []string{"GROUP-APPROVER-1", "GROUP-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentGroupApprovers.String()])
		assert.Equal(t, "TENANT-1234", node.PropertyMap[azure.TenantID.String()])
		assert.Equal(t, true, node.PropertyMap[azure.EndUserAssignmentRequiresApproval.String()])
		assert.Equal(t, []string{"USER-APPROVER-1", "USER-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()])
		require.Len(t, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()], 2)
		assert.Equal(t, []string{"USER-APPROVER-1", "USER-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()])

		// Assert created relationships
		require.Len(t, rels, 4)

		assert.Equal(t, "USER-APPROVER-1", rels[0].Source.Value)
		assert.Equal(t, azure.User, rels[0].Source.Kind)
		assert.Equal(t, azure.Role, rels[0].Target.Kind)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[0].Target.Value)
		assert.Equal(t, azure.AZRoleApprover, rels[0].RelType)

		assert.Equal(t, "USER-APPROVER-2", rels[1].Source.Value)
		assert.Equal(t, azure.User, rels[1].Source.Kind)
		assert.Equal(t, azure.Role, rels[1].Target.Kind)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[1].Target.Value)
		assert.Equal(t, azure.AZRoleApprover, rels[1].RelType)

		assert.Equal(t, "GROUP-APPROVER-1", rels[2].Source.Value)
		assert.Equal(t, azure.Group, rels[2].Source.Kind)
		assert.Equal(t, azure.Role, rels[2].Target.Kind)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[2].Target.Value)
		assert.Equal(t, azure.AZRoleApprover, rels[2].RelType)

		assert.Equal(t, "GROUP-APPROVER-2", rels[3].Source.Value)
		assert.Equal(t, azure.Group, rels[3].Source.Kind)
		assert.Equal(t, azure.Role, rels[3].Target.Kind)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[3].Target.Value)
		assert.Equal(t, azure.AZRoleApprover, rels[3].RelType)
	})
}
