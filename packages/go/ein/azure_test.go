package ein

import (
	"testing"

	"github.com/bloodhoundad/azurehound/v2/models"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

		node, rels := ConvertAzureRoleManagementPolicyAssignment(model)

		// Assert created node properties
		assert.Equal(t, "ROLE-1234@TENANT-1234", node.ObjectID)
		assert.Equal(t, "AZRole", node.Label.String())
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

		node, rels := ConvertAzureRoleManagementPolicyAssignment(model)

		// Assert created node properties
		assert.Equal(t, "ROLE-1234@TENANT-1234", node.ObjectID)
		assert.Equal(t, "AZRole", node.Label.String())
		require.Len(t, node.PropertyMap[azure.EndUserAssignmentGroupApprovers.String()], 2)
		assert.Equal(t, []string{"GROUP-APPROVER-1", "GROUP-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentGroupApprovers.String()])
		assert.Equal(t, "TENANT-1234", node.PropertyMap[azure.TenantID.String()])
		assert.Equal(t, true, node.PropertyMap[azure.EndUserAssignmentRequiresApproval.String()])
		assert.Equal(t, []string{"USER-APPROVER-1", "USER-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()])
		require.Len(t, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()], 2)
		assert.Equal(t, []string{"USER-APPROVER-1", "USER-APPROVER-2"}, node.PropertyMap[azure.EndUserAssignmentUserApprovers.String()])

		// Assert created relationships
		require.Len(t, rels, 4)

		assert.Equal(t, "USER-APPROVER-1", rels[0].Source)
		assert.Equal(t, azure.User, rels[0].SourceType)
		assert.Equal(t, azure.Role, rels[0].TargetType)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[0].Target)
		assert.Equal(t, azure.AZRoleApproval, rels[0].RelType)

		assert.Equal(t, "USER-APPROVER-2", rels[1].Source)
		assert.Equal(t, azure.User, rels[1].SourceType)
		assert.Equal(t, azure.Role, rels[1].TargetType)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[1].Target)
		assert.Equal(t, azure.AZRoleApproval, rels[1].RelType)

		assert.Equal(t, "GROUP-APPROVER-1", rels[2].Source)
		assert.Equal(t, azure.Group, rels[2].SourceType)
		assert.Equal(t, azure.Role, rels[2].TargetType)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[2].Target)
		assert.Equal(t, azure.AZRoleApproval, rels[2].RelType)

		assert.Equal(t, "GROUP-APPROVER-2", rels[3].Source)
		assert.Equal(t, azure.Group, rels[3].SourceType)
		assert.Equal(t, azure.Role, rels[3].TargetType)
		assert.Equal(t, "ROLE-1234@TENANT-1234", rels[3].Target)
		assert.Equal(t, azure.AZRoleApproval, rels[3].RelType)
	})
}
