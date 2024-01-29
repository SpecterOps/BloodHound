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

package ein

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/bloodhoundad/azurehound/v2/constants"
	"github.com/bloodhoundad/azurehound/v2/enums"
	"github.com/bloodhoundad/azurehound/v2/models"
	azure2 "github.com/bloodhoundad/azurehound/v2/models/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

const (
	ISO8601               string = "2006-01-02T15:04:05Z"
	KeyVaultPermissionGet string = "Get"
)

var resourceGroupLevel = regexp.MustCompile(`^[\\w\\d\\-\\/]*/resourceGroups/[0-9a-zA-Z]+$`)

func ConvertAZAppToNode(app models.App) IngestibleNode {
	return IngestibleNode{
		PropertyMap: map[string]any{
			common.Name.String():           strings.ToUpper(fmt.Sprintf("%s@%s", app.DisplayName, app.PublisherDomain)),
			common.Description.String():    app.Description,
			common.DisplayName.String():    app.DisplayName,
			common.LastSeen.String():       time.Now().UTC(),
			common.WhenCreated.String():    ParseISO8601(app.CreatedDateTime),
			azure.AppID.String():           app.AppId,
			azure.PublisherDomain.String(): app.PublisherDomain,
			azure.SignInAudience.String():  app.SignInAudience,
			azure.TenantID.String():        strings.ToUpper(app.TenantId),
		},
		ObjectID: strings.ToUpper(app.AppId),
		Label:    azure.App,
	}
}

func ConvertAZAppRelationships(app models.App) []IngestibleRelationship {
	return []IngestibleRelationship{{
		Source:     strings.ToUpper(app.TenantId),
		SourceType: azure.Tenant,
		TargetType: azure.App,
		Target:     strings.ToUpper(app.AppId),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	}}
}

func ConvertAZDeviceToNode(device models.Device) IngestibleNode {
	return IngestibleNode{
		PropertyMap: map[string]any{
			common.Name.String():                  strings.ToUpper(fmt.Sprintf("%s@%s", device.DisplayName, device.TenantName)),
			common.DisplayName.String():           device.DisplayName,
			common.OperatingSystem.String():       device.OperatingSystem,
			azure.DeviceID.String():               device.DeviceId,
			azure.OperatingSystemVersion.String(): device.OperatingSystemVersion,
			azure.TrustType.String():              device.TrustType,
			azure.TenantID.String():               strings.ToUpper(device.TenantId),
		},
		ObjectID: strings.ToUpper(device.Id),
		Label:    azure.Device,
	}
}

func ConvertAZDeviceRelationships(device models.Device) []IngestibleRelationship {
	return []IngestibleRelationship{{
		Source:     strings.ToUpper(device.TenantId),
		SourceType: azure.Tenant,
		TargetType: azure.Device,
		Target:     strings.ToUpper(device.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	}}
}

func ConvertAZVMScaleSetToNode(scaleSet models.VMScaleSet) IngestibleNode {
	return IngestibleNode{
		ObjectID: strings.ToUpper(scaleSet.Id),
		PropertyMap: map[string]any{
			common.Name.String():    strings.ToUpper(scaleSet.Name),
			azure.TenantID.String(): strings.ToUpper(scaleSet.TenantId),
		},
		Label: azure.VMScaleSet,
	}
}

func ConvertAZVMScaleSetRelationships(scaleSet models.VMScaleSet) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(scaleSet.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.VMScaleSet,
		Target:     strings.ToUpper(scaleSet.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	// Enumerate System Assigned Identities
	if scaleSet.Identity.PrincipalId != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(scaleSet.Id),
			SourceType: azure.VMScaleSet,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(scaleSet.Identity.PrincipalId),
			RelProps:   map[string]any{},
			RelType:    azure.ManagedIdentity,
		})
	}

	// Enumerate User Assigned Identities
	for _, identity := range scaleSet.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(scaleSet.Id),
				SourceType: azure.VMScaleSet,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(identity.PrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.ManagedIdentity,
			})
		}
	}

	return relationships
}

func ConvertAzureVMScaleSetRoleAssignment(data models.AzureRoleAssignments) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.RoleAssignments {
		if strings.EqualFold(raw.Assignee.Properties.Scope, raw.ObjectId) {
			if slices.Contains([]string{
				constants.OwnerRoleID,
				constants.UserAccessAdminRoleID,
				constants.ContributorRoleID,
				constants.VirtualMachineContributorRoleID,
			}, strings.ToLower(raw.RoleDefinitionId)) {
				relationships = append(relationships, IngestibleRelationship{
					Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
					SourceType: azure.Entity,
					TargetType: azure.VMScaleSet,
					Target:     strings.ToUpper(data.ObjectId),
					RelProps:   map[string]any{},
					RelType:    KindFromRoleId(raw.RoleDefinitionId),
				})
			}
		}
	}

	return relationships
}

func ConvertAzureOwnerToRel(directoryObject azure2.DirectoryObject, ownerType graph.Kind, targetType graph.Kind, targetId string) IngestibleRelationship {
	return IngestibleRelationship{
		Source:     strings.ToUpper(directoryObject.Id),
		SourceType: ownerType,
		TargetType: targetType,
		Target:     strings.ToUpper(targetId),
		RelProps:   map[string]any{},
		RelType:    azure.Owns,
	}
}

func ConvertAzureAppRoleAssignmentToNodes(data models.AppRoleAssignment) []IngestibleNode {
	nodes := make([]IngestibleNode, 0)

	nodes = append(nodes, IngestibleNode{
		PropertyMap: map[string]any{
			common.DisplayName.String(): strings.ToUpper(data.PrincipalDisplayName),
			azure.TenantID.String():     strings.ToUpper(data.TenantId),
		},
		ObjectID: strings.ToUpper(data.PrincipalId.String()),
		Label:    azure.ServicePrincipal,
	})

	nodes = append(nodes, IngestibleNode{
		PropertyMap: map[string]any{
			common.DisplayName.String(): strings.ToUpper(data.ResourceDisplayName),
			azure.TenantID.String():     strings.ToUpper(data.TenantId),
		},
		ObjectID: strings.ToUpper(data.ResourceId),
		Label:    azure.ServicePrincipal,
	})

	return nodes
}

func ConvertAzureAppRoleAssignmentToRel(data models.AppRoleAssignment) IngestibleRelationship {
	if appRoleKind, hasAppRoleKind := azure.RelationshipKindByAppRoleID[strings.ToLower(data.AppRoleId.String())]; hasAppRoleKind {
		return IngestibleRelationship{
			Source:     strings.ToUpper(data.PrincipalId.String()),
			SourceType: azure.ServicePrincipal,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(data.ResourceId),
			RelProps:   map[string]any{},
			RelType:    appRoleKind,
		}
	}

	return IngestibleRelationship{
		Source:     "",
		SourceType: nil,
		TargetType: nil,
		Target:     "",
		RelProps:   nil,
		RelType:    nil,
	}
}

func ConvertAzureFunctionAppToNode(data models.FunctionApp) IngestibleNode {
	return IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():    strings.ToUpper(data.Name),
			azure.TenantID.String(): strings.ToUpper(data.TenantId),
		},
		Label: azure.FunctionApp,
	}
}

func ConvertAzureFunctionAppToRels(data models.FunctionApp) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(data.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.FunctionApp,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(data.Id),
			SourceType: azure.FunctionApp,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(data.Identity.PrincipalId),
			RelProps:   map[string]any{},
			RelType:    azure.ManagedIdentity,
		})
	}

	// Enumerate User Assigned Identities
	for _, identity := range data.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(data.Id),
				SourceType: azure.FunctionApp,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(identity.PrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.ManagedIdentity,
			})
		}
	}

	return relationships
}

func ConvertAzureFunctionAppRoleAssignmentToRels(data models.AzureRoleAssignments) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.RoleAssignments {
		if strings.EqualFold(raw.Assignee.Properties.Scope, raw.ObjectId) {
			if slices.Contains([]string{
				constants.OwnerRoleID,
				constants.UserAccessAdminRoleID,
				constants.ContributorRoleID,
				constants.WebsiteContributorRoleID,
			}, strings.ToLower(raw.RoleDefinitionId)) {
				relationships = append(relationships, IngestibleRelationship{
					Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
					SourceType: azure.Entity,
					TargetType: azure.FunctionApp,
					Target:     strings.ToUpper(data.ObjectId),
					RelProps:   map[string]any{},
					RelType:    KindFromRoleId(raw.RoleDefinitionId),
				})
			}
		}
	}

	return relationships
}

func ConvertAzureGroupToNode(data models.Group) IngestibleNode {
	return IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():              strings.ToUpper(fmt.Sprintf("%s@%s", data.DisplayName, data.TenantName)),
			common.WhenCreated.String():       ParseISO8601(data.CreatedDateTime),
			common.Description.String():       data.Description,
			common.DisplayName.String():       data.DisplayName,
			azure.IsAssignableToRole.String(): data.IsAssignableToRole,
			azure.OnPremID.String():           data.OnPremisesSecurityIdentifier,
			azure.OnPremSyncEnabled.String():  data.OnPremisesSyncEnabled,
			azure.SecurityEnabled.String():    data.SecurityEnabled,
			azure.SecurityIdentifier.String(): data.SecurityIdentifier,
			azure.TenantID.String():           strings.ToUpper(data.TenantId),
		},
		Label: azure.Group,
	}
}

func ConvertAzureGroupToOnPremisesNode(data models.Group) IngestibleNode {
	if data.OnPremisesSecurityIdentifier != "" {
		return IngestibleNode{
			ObjectID:    strings.ToUpper(data.OnPremisesSecurityIdentifier),
			PropertyMap: map[string]any{},
			Label:       ad.Group,
		}
	}

	return IngestibleNode{
		ObjectID:    "",
		PropertyMap: nil,
		Label:       nil,
	}
}

func ConvertAzureGroupToRel(data models.Group) IngestibleRelationship {
	return IngestibleRelationship{
		Source:     strings.ToUpper(data.TenantId),
		SourceType: azure.Tenant,
		TargetType: azure.Group,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	}
}

func ConvertAzureGroupMembersToRels(data models.GroupMembers) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Members {
		var (
			member azure2.DirectoryObject
		)
		if err := json.Unmarshal(raw.Member, &member); err != nil {
			log.Errorf(SerialError, "azure group member", err)
		} else if memberType, err := ExtractTypeFromDirectoryObject(member); err != nil {
			log.Errorf(ExtractError, err)
		} else {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(member.Id),
				SourceType: memberType,
				TargetType: azure.Group,
				Target:     strings.ToUpper(data.GroupId),
				RelProps:   map[string]any{},
				RelType:    azure.MemberOf,
			})
		}
	}

	return relationships
}

func ConvertAzureGroupOwnerToRels(data models.GroupOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Owners {
		var (
			owner azure2.DirectoryObject
		)
		if err := json.Unmarshal(raw.Owner, &owner); err != nil {
			log.Errorf(SerialError, "azure group owner", err)
		} else if ownerType, err := ExtractTypeFromDirectoryObject(owner); err != nil {
			log.Errorf(ExtractError, err)
		} else {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(owner.Id),
				SourceType: ownerType,
				TargetType: azure.Group,
				Target:     strings.ToUpper(data.GroupId),
				RelProps:   map[string]any{},
				RelType:    azure.Owns,
			})
		}
	}

	return relationships
}

func ConvertAzureKeyVault(data models.KeyVault) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.Name.String():                   strings.ToUpper(data.Name),
				azure.EnableRBACAuthorization.String(): data.Properties.EnableRbacAuthorization,
				azure.TenantID.String():                strings.ToUpper(data.TenantId),
			},
			Label: azure.KeyVault,
		},
		IngestibleRelationship{
			Source:     strings.ToUpper(data.ResourceGroup),
			SourceType: azure.ResourceGroup,
			TargetType: azure.KeyVault,
			Target:     strings.ToUpper(data.Id),
			RelProps:   map[string]any{},
			RelType:    azure.Contains,
		}
}

func ConvertAzureKeyVaultAccessPolicy(data models.KeyVaultAccessPolicy) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, relType := range getKeyVaultPermissions(data) {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(data.ObjectId),
			SourceType: azure.Entity,
			TargetType: azure.KeyVault,
			Target:     strings.ToUpper(data.KeyVaultId),
			RelProps:   map[string]any{},
			RelType:    relType,
		})
	}

	return relationships
}

func ConvertAzureKeyVaultContributor(data models.KeyVaultContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Contributors {
		if data.KeyVaultId == raw.Contributor.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.Contributor.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.KeyVault,
				Target:     strings.ToUpper(data.KeyVaultId),
				RelProps:   map[string]any{},
				RelType:    azure.Contributor,
			})
		}
	}

	return relationships
}

func ConvertAzureKeyVaultKVContributor(data models.KeyVaultKVContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.KVContributors {
		if data.KeyVaultId == raw.KVContributor.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.KVContributor.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.KeyVault,
				Target:     strings.ToUpper(data.KeyVaultId),
				RelProps:   map[string]any{},
				RelType:    azure.KeyVaultContributor,
			})
		}
	}

	return relationships
}

func ConvertAzureKeyVaultOwnerToRels(data models.KeyVaultOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Owners {
		if data.KeyVaultId == raw.Owner.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.Owner.Properties.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.KeyVault,
				Target:     strings.ToUpper(data.KeyVaultId),
				RelProps:   map[string]any{},
				RelType:    azure.Owner,
			})
		}
	}

	return relationships
}

func ConvertAzureKeyVaultUserAccessAdminToRels(data models.KeyVaultUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if data.KeyVaultId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.KeyVault,
				Target:     strings.ToUpper(data.KeyVaultId),
				RelProps:   map[string]any{},
				RelType:    azure.UserAccessAdministrator,
			})
		}
	}

	return relationships
}

func ConvertAzureManagementGroupDescendantToRel(data azure2.DescendantInfo) IngestibleRelationship {
	return IngestibleRelationship{
		Source:     strings.ToUpper(data.Properties.Parent.Id),
		SourceType: azure.ManagementGroup,
		TargetType: azure.Entity,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	}
}

func ConvertAzureManagementGroupOwnerToRels(data models.ManagementGroupOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		if data.ManagementGroupId == raw.Owner.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.Owner.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.ManagementGroup,
				Target:     strings.ToUpper(data.ManagementGroupId),
				RelProps:   map[string]any{},
				RelType:    azure.Owner,
			})
		}
	}

	return relationships
}

func ConvertAzureManagementGroupUserAccessAdminToRels(data models.ManagementGroupUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if data.ManagementGroupId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.UserAccessAdmin.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.ManagementGroup,
				Target:     strings.ToUpper(data.ManagementGroupId),
				RelProps:   map[string]any{},
				RelType:    azure.UserAccessAdministrator,
			})
		}
	}
	return relationships
}

func ConvertAzureManagementGroup(data models.ManagementGroup) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.Name.String():    strings.ToUpper(fmt.Sprintf("%s@%s", data.Properties.DisplayName, data.TenantName)),
				azure.TenantID.String(): strings.ToUpper(data.TenantId),
			},
			Label: azure.ManagementGroup,
		}, IngestibleRelationship{
			Source:     strings.ToUpper(data.TenantId),
			SourceType: azure.Tenant,
			TargetType: azure.ManagementGroup,
			Target:     strings.ToUpper(data.Id),
			RelProps:   map[string]any{},
			RelType:    azure.Contains,
		}
}

func ConvertAzureResourceGroup(data models.ResourceGroup) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.Name.String():    strings.ToUpper(data.Name),
				azure.TenantID.String(): strings.ToUpper(data.TenantId),
			},
			Label: azure.ResourceGroup,
		}, IngestibleRelationship{
			Source:     strings.ToUpper(data.SubscriptionId),
			SourceType: azure.Subscription,
			TargetType: azure.ResourceGroup,
			Target:     strings.ToUpper(data.Id),
			RelProps:   map[string]any{},
			RelType:    azure.Contains,
		}
}

func ConvertAzureResourceGroupOwnerToRels(data models.ResourceGroupOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		if data.ResourceGroupId == raw.Owner.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.Owner.Properties.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.ResourceGroup,
				Target:     strings.ToUpper(data.ResourceGroupId),
				RelProps:   map[string]any{},
				RelType:    azure.Owner,
			})
		}
	}

	return relationships
}

func ConvertAzureResourceGroupUserAccessAdminToRels(data models.ResourceGroupUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if data.ResourceGroupId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.ResourceGroup,
				Target:     strings.ToUpper(data.ResourceGroupId),
				RelProps:   map[string]any{},
				RelType:    azure.UserAccessAdministrator,
			})
		}
	}

	return relationships
}

func ConvertAzureRole(data models.Role) (IngestibleNode, IngestibleRelationship) {
	roleObjectId := fmt.Sprintf("%s@%s", strings.ToUpper(data.Id), strings.ToUpper(data.TenantId))
	return IngestibleNode{
			ObjectID: roleObjectId,
			PropertyMap: map[string]any{
				common.Name.String():          strings.ToUpper(fmt.Sprintf("%s@%s", data.DisplayName, data.TenantName)),
				common.Description.String():   data.Description,
				common.DisplayName.String():   data.DisplayName,
				common.Enabled.String():       data.IsEnabled,
				azure.IsBuiltIn.String():      data.IsBuiltIn,
				azure.RoleTemplateID.String(): data.TemplateId,
				azure.TenantID.String():       strings.ToUpper(data.TenantId),
			},
			Label: azure.Role,
		}, IngestibleRelationship{
			Source:     strings.ToUpper(data.TenantId),
			SourceType: azure.Tenant,
			TargetType: azure.Role,
			Target:     roleObjectId,
			RelProps:   map[string]any{},
			RelType:    azure.Contains,
		}
}

func ConvertAzureRoleAssignmentToRels(roleAssignment azure2.UnifiedRoleAssignment, data models.RoleAssignments, roleObjectId string) []IngestibleRelationship {
	var (
		scope         string
		relationships = make([]IngestibleRelationship, 0)
	)

	if roleAssignment.DirectoryScopeId == "/" {
		scope = strings.ToUpper(data.TenantId)
	} else {
		scope = strings.ToUpper(roleAssignment.DirectoryScopeId[1:])
	}

	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(roleAssignment.PrincipalId),
		SourceType: azure.Entity,
		TargetType: azure.Role,
		Target:     roleObjectId,
		RelProps: map[string]any{
			azure.Scope.String(): scope,
		},
		RelType: azure.HasRole,
	})

	if CanAddSecret(roleAssignment.RoleDefinitionId) && roleAssignment.DirectoryScopeId != "/" {
		if relType, err := GetAddSecretRoleKind(roleAssignment.RoleDefinitionId); err != nil {
			log.Errorf("Error processing role assignment for role %s: %v", roleObjectId, err)
		} else {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(roleAssignment.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.Entity,
				Target:     scope,
				RelProps:   map[string]any{},
				RelType:    relType,
			})
		}
	}

	return relationships
}

func ConvertAzureServicePrincipal(data models.ServicePrincipal) ([]IngestibleNode, []IngestibleRelationship) {
	nodes := make([]IngestibleNode, 0)
	relationships := make([]IngestibleRelationship, 0)

	nodes = append(nodes, IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():                  strings.ToUpper(fmt.Sprintf("%s@%s", data.DisplayName, data.TenantName)),
			common.Enabled.String():               data.AccountEnabled,
			common.DisplayName.String():           data.DisplayName,
			common.Description.String():           data.Description,
			azure.AppOwnerOrganizationID.String(): data.AppOwnerOrganizationId,
			azure.AppDescription.String():         data.AppDescription,
			azure.AppDisplayName.String():         data.AppDisplayName,
			azure.ServicePrincipalType.String():   data.ServicePrincipalType,
			azure.TenantID.String():               strings.ToUpper(data.TenantId),
		},
		Label: azure.ServicePrincipal,
	})

	nodes = append(nodes, IngestibleNode{
		ObjectID: strings.ToUpper(data.AppId),
		PropertyMap: map[string]any{
			common.DisplayName.String(): data.AppDisplayName,
			azure.TenantID.String():     strings.ToUpper(data.AppOwnerOrganizationId),
		},
		Label: azure.App,
	})

	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(data.AppId),
		SourceType: azure.App,
		TargetType: azure.ServicePrincipal,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.RunsAs,
	})

	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(data.TenantId),
		SourceType: azure.Tenant,
		TargetType: azure.ServicePrincipal,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	return nodes, relationships
}

func ConvertAzureLogicApp(logicApp models.LogicApp) (IngestibleNode, []IngestibleRelationship) {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(logicApp.Id),
		PropertyMap: map[string]any{
			common.Name.String():    strings.ToUpper(logicApp.Name),
			azure.TenantID.String(): strings.ToUpper(logicApp.TenantId),
		},
		Label: azure.LogicApp,
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(logicApp.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.LogicApp,
		Target:     strings.ToUpper(logicApp.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	// Enumerate System Assigned Identities
	if logicApp.Identity.PrincipalId != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(logicApp.Id),
			SourceType: azure.LogicApp,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(logicApp.Identity.PrincipalId),
			RelProps:   map[string]any{},
			RelType:    azure.ManagedIdentity,
		})
	}

	// Enumerate User Assigned Identities
	for _, identity := range logicApp.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(logicApp.Id),
				SourceType: azure.LogicApp,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(identity.PrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.ManagedIdentity,
			})
		}
	}

	return node, relationships
}

func ConvertAzureLogicAppRoleAssignment(roleAssignment models.AzureRoleAssignments) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range roleAssignment.RoleAssignments {
		if strings.EqualFold(raw.Assignee.Properties.Scope, raw.ObjectId) {
			if slices.Contains([]string{
				constants.OwnerRoleID,
				constants.UserAccessAdminRoleID,
				constants.ContributorRoleID,
				constants.LogicAppContributorRoleID,
			}, strings.ToLower(raw.RoleDefinitionId)) {
				relationships = append(relationships, IngestibleRelationship{
					Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
					SourceType: azure.Entity,
					TargetType: azure.LogicApp,
					Target:     strings.ToUpper(roleAssignment.ObjectId),
					RelProps:   map[string]any{},
					RelType:    KindFromRoleId(raw.RoleDefinitionId),
				})
			}
		}
	}

	return relationships
}

func ConvertAzureServicePrincipalOwnerToRels(data models.ServicePrincipalOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		var (
			owner azure2.DirectoryObject
		)

		if err := json.Unmarshal(raw.Owner, &owner); err != nil {
			log.Errorf(SerialError, "azure service principal owner", err)
		} else if ownerType, err := ExtractTypeFromDirectoryObject(owner); err != nil {
			log.Errorf(ExtractError, err)
		} else {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(owner.Id),
				SourceType: ownerType,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(data.ServicePrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.Owns,
			})
		}
	}
	return relationships
}

func ConvertAzureSubscription(data azure2.Subscription) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.DisplayName.String(): data.DisplayName,
				common.ObjectID.String():    data.SubscriptionId,
				common.Name.String():        strings.ToUpper(data.DisplayName),
				azure.TenantID.String():     strings.ToUpper(data.TenantId),
			},
			Label: azure.Subscription,
		},
		IngestibleRelationship{
			Source:     strings.ToUpper(data.TenantId),
			SourceType: azure.Tenant,
			TargetType: azure.Subscription,
			Target:     strings.ToUpper(data.Id),
			RelProps:   map[string]any{},
			RelType:    azure.Contains,
		}
}

func ConvertAzureSubscriptionOwnerToRels(data models.SubscriptionOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Owners {
		if data.SubscriptionId == raw.Owner.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.Owner.Properties.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.Subscription,
				Target:     strings.ToUpper(data.SubscriptionId),
				RelProps:   map[string]any{},
				RelType:    azure.Owner,
			})
		}
	}

	return relationships
}

func ConvertAzureSubscriptionUserAccessAdminToRels(data models.SubscriptionUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.UserAccessAdmins {
		if data.SubscriptionId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.Subscription,
				Target:     strings.ToUpper(data.SubscriptionId),
				RelProps:   map[string]any{},
				RelType:    azure.UserAccessAdministrator,
			})
		}
	}

	return relationships
}

func ConvertAzureTenantToNode(data models.Tenant) IngestibleNode {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.TenantId),
		PropertyMap: map[string]any{
			common.DisplayName.String(): data.DisplayName,
			common.ObjectID.String():    data.Id,
			common.Name.String():        strings.ToUpper(data.DisplayName),
			azure.TenantID.String():     strings.ToUpper(data.TenantId),
		},
		Label: azure.Tenant,
	}

	if data.Collected {
		node.PropertyMap["collected"] = true
	}

	return node
}

// ConvertAzureUser returns the basic node, the on prem node and then the ingestible contains relationship
func ConvertAzureUser(data models.User) (IngestibleNode, IngestibleNode, IngestibleRelationship) {
	onPremNode := IngestibleNode{}
	if data.OnPremisesSecurityIdentifier != "" {
		onPremNode = IngestibleNode{
			ObjectID:    strings.ToUpper(data.OnPremisesSecurityIdentifier),
			PropertyMap: map[string]any{},
			Label:       ad.User,
		}
	}

	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.Name.String():             strings.ToUpper(data.UserPrincipalName),
				common.Enabled.String():          data.AccountEnabled,
				common.WhenCreated.String():      ParseISO8601(data.CreatedDateTime),
				common.DisplayName.String():      data.DisplayName,
				common.Title.String():            data.JobTitle,
				common.PasswordLastSet.String():  ParseISO8601(data.LastPasswordChangeDateTime),
				common.Email.String():            data.Mail,
				azure.OnPremID.String():          data.OnPremisesSecurityIdentifier,
				azure.OnPremSyncEnabled.String(): data.OnPremisesSyncEnabled,
				azure.UserPrincipalName.String(): data.UserPrincipalName,
				azure.UserType.String():          data.UserType,
				azure.TenantID.String():          strings.ToUpper(data.TenantId),
			},
			Label: azure.User,
		}, onPremNode, IngestibleRelationship{
			Source:     strings.ToUpper(data.TenantId),
			SourceType: azure.Tenant,
			TargetType: azure.User,
			Target:     strings.ToUpper(data.Id),
			RelProps:   map[string]any{},
			RelType:    azure.Contains,
		}
}

func ConvertAzureVirtualMachine(data models.VirtualMachine) (IngestibleNode, []IngestibleRelationship) {
	relationships := make([]IngestibleRelationship, 0)
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():            strings.ToUpper(data.Name),
			common.ObjectID.String():        data.Properties.VMId,
			common.OperatingSystem.String(): data.Properties.StorageProfile.OSDisk.OSType,
			azure.TenantID.String():         strings.ToUpper(data.TenantId),
		},
		Label: azure.VM,
	}

	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(data.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.VM,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(data.Id),
			SourceType: azure.VM,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(data.Identity.PrincipalId),
			RelProps:   map[string]any{},
			RelType:    azure.ManagedIdentity,
		})
	}

	// Enumerate User Assigned Identities
	for _, identity := range data.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(data.Id),
				SourceType: azure.VM,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(identity.PrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.ManagedIdentity,
			})
		}
	}

	return node, relationships
}

func ConvertAzureVirtualMachineAdminLoginToRels(data models.VirtualMachineAdminLogins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.AdminLogins {
		if ResourceWithinScope(data.VirtualMachineId, raw.AdminLogin.Properties.Scope) {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.AdminLogin.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.VM,
				Target:     strings.ToUpper(data.VirtualMachineId),
				RelProps:   map[string]any{},
				RelType:    azure.VMAdminLogin,
			})
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineAvereContributorToRels(data models.VirtualMachineAvereContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.AvereContributors {
		if data.VirtualMachineId == raw.AvereContributor.Properties.Scope {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.AvereContributor.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.VM,
				Target:     strings.ToUpper(data.VirtualMachineId),
				RelProps:   map[string]any{},
				RelType:    azure.AvereContributor,
			})
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineContributorToRels(data models.VirtualMachineContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Contributors {
		if ResourceWithinScope(data.VirtualMachineId, raw.Contributor.Properties.Scope) {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.Contributor.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.VM,
				Target:     strings.ToUpper(data.VirtualMachineId),
				RelProps:   map[string]any{},
				RelType:    azure.Contributor,
			})
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineVMContributorToRels(data models.VirtualMachineVMContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.VMContributors {
		if ResourceWithinScope(data.VirtualMachineId, raw.VMContributor.Properties.Scope) {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.VMContributor.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.VM,
				Target:     strings.ToUpper(data.VirtualMachineId),
				RelProps:   map[string]any{},
				RelType:    azure.VMContributor,
			})
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineOwnerToRels(data models.VirtualMachineOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		if ResourceWithinScope(data.VirtualMachineId, raw.Owner.Properties.Scope) {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.Owner.GetPrincipalId()),
				SourceType: azure.Entity,
				TargetType: azure.VM,
				Target:     strings.ToUpper(data.VirtualMachineId),
				RelProps:   map[string]any{},
				RelType:    azure.Owner,
			})
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineUserAccessAdminToRels(data models.VirtualMachineUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if ResourceWithinScope(data.VirtualMachineId, raw.UserAccessAdmin.Properties.Scope) {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
				SourceType: azure.Entity,
				TargetType: azure.VM,
				Target:     strings.ToUpper(data.VirtualMachineId),
				RelProps:   map[string]any{},
				RelType:    azure.UserAccessAdministrator,
			})
		}
	}
	return relationships
}

func ConvertAzureManagedCluster(data models.ManagedCluster, nodeResourceGroupID string) (IngestibleNode, []IngestibleRelationship) {
	relationships := make([]IngestibleRelationship, 0)
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():               strings.ToUpper(data.Name),
			azure.TenantID.String():            strings.ToUpper(data.TenantId),
			azure.NodeResourceGroupID.String(): strings.ToUpper(nodeResourceGroupID),
		},
		Label: azure.ManagedCluster,
	}

	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(data.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.ManagedCluster,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(data.Id),
		SourceType: azure.ManagedCluster,
		TargetType: azure.ResourceGroup,
		Target:     strings.ToUpper(nodeResourceGroupID),
		RelProps:   map[string]any{},
		RelType:    azure.NodeResourceGroup,
	})

	return node, relationships
}

func ConvertAzureManagedClusterRoleAssignmentToRels(data models.AzureRoleAssignments) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.RoleAssignments {
		if strings.EqualFold(raw.Assignee.Properties.Scope, raw.ObjectId) {
			if slices.Contains([]string{
				azure.OwnerRole,
				azure.UserAccessAdminRole,
				azure.ContributorRole,
				azure.AKSContributorRole,
			}, strings.ToLower(raw.RoleDefinitionId)) {
				relationships = append(relationships, IngestibleRelationship{
					Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
					SourceType: azure.Entity,
					TargetType: azure.ManagedCluster,
					Target:     strings.ToUpper(data.ObjectId),
					RelProps:   map[string]any{},
					RelType:    KindFromRoleId(raw.RoleDefinitionId),
				})
			}
		}
	}
	return relationships
}

func ConvertAzureContainerRegistry(data models.ContainerRegistry) (IngestibleNode, []IngestibleRelationship) {
	relationships := make([]IngestibleRelationship, 0)
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():    strings.ToUpper(data.Name),
			azure.TenantID.String(): strings.ToUpper(data.TenantId),
		},
		Label: azure.ContainerRegistry,
	}

	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(data.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.ContainerRegistry,
		Target:     strings.ToUpper(data.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(data.Id),
			SourceType: azure.ContainerRegistry,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(data.Identity.PrincipalId),
			RelProps:   map[string]any{},
			RelType:    azure.ManagedIdentity,
		})
	}

	// Enumerate User Assigned Identities
	for _, identity := range data.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(data.Id),
				SourceType: azure.ContainerRegistry,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(identity.PrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.ManagedIdentity,
			})
		}
	}

	return node, relationships
}

func ConvertAzureWebApp(webApp models.WebApp) (IngestibleNode, []IngestibleRelationship) {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(webApp.Id),
		PropertyMap: map[string]any{
			common.Name.String():    strings.ToUpper(webApp.Name),
			azure.TenantID.String(): strings.ToUpper(webApp.TenantId),
		},
		Label: azure.WebApp,
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(webApp.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.WebApp,
		Target:     strings.ToUpper(webApp.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	// Enumerate System Assigned Identities
	if webApp.Identity.PrincipalId != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(webApp.Id),
			SourceType: azure.WebApp,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(webApp.Identity.PrincipalId),
			RelProps:   map[string]any{},
			RelType:    azure.ManagedIdentity,
		})
	}

	// Enumerate User Assigned Identities
	for _, identity := range webApp.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(webApp.Id),
				SourceType: azure.WebApp,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(identity.PrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.ManagedIdentity,
			})
		}
	}

	return node, relationships
}

func ConvertAzureAutomationAccountRoleAssignment(roleAssignments models.AzureRoleAssignments) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range roleAssignments.RoleAssignments {
		if strings.EqualFold(raw.Assignee.Properties.Scope, raw.ObjectId) {
			if slices.Contains([]string{
				constants.OwnerRoleID,
				constants.UserAccessAdminRoleID,
				constants.ContributorRoleID,
				constants.AutomationContributorRoleID,
			}, strings.ToLower(raw.RoleDefinitionId)) {
				relationships = append(relationships, IngestibleRelationship{
					Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
					SourceType: azure.Entity,
					TargetType: azure.AutomationAccount,
					Target:     strings.ToUpper(roleAssignments.ObjectId),
					RelProps:   map[string]any{},
					RelType:    KindFromRoleId(raw.RoleDefinitionId),
				})
			}
		}
	}

	return relationships
}

func ConvertAzureContainerRegistryRoleAssignment(roleAssignment models.AzureRoleAssignments) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range roleAssignment.RoleAssignments {
		if strings.EqualFold(raw.Assignee.Properties.Scope, raw.ObjectId) {
			if slices.Contains([]string{
				constants.OwnerRoleID,
				constants.UserAccessAdminRoleID,
				constants.ContributorRoleID,
			}, strings.ToLower(raw.RoleDefinitionId)) {
				relationships = append(relationships, IngestibleRelationship{
					Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
					SourceType: azure.Entity,
					TargetType: azure.ContainerRegistry,
					Target:     strings.ToUpper(roleAssignment.ObjectId),
					RelProps:   map[string]any{},
					RelType:    KindFromRoleId(raw.RoleDefinitionId),
				})
			}
		}
	}

	return relationships
}

func ConvertAzureWebAppRoleAssignment(roleAssignment models.AzureRoleAssignments) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range roleAssignment.RoleAssignments {
		if strings.EqualFold(raw.Assignee.Properties.Scope, raw.ObjectId) {
			if slices.Contains([]string{
				constants.OwnerRoleID,
				constants.UserAccessAdminRoleID,
				constants.ContributorRoleID,
				constants.WebsiteContributorRoleID,
			}, strings.ToLower(raw.RoleDefinitionId)) {
				relationships = append(relationships, IngestibleRelationship{
					Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
					SourceType: azure.Entity,
					TargetType: azure.WebApp,
					Target:     strings.ToUpper(roleAssignment.ObjectId),
					RelProps:   map[string]any{},
					RelType:    KindFromRoleId(raw.RoleDefinitionId),
				})
			}
		}
	}

	return relationships
}

func ConvertAzureAutomationAccount(account models.AutomationAccount) (IngestibleNode, []IngestibleRelationship) {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(account.Id),
		PropertyMap: map[string]any{
			common.Name.String():    strings.ToUpper(account.Name),
			azure.TenantID.String(): strings.ToUpper(account.TenantId),
		},
		Label: azure.AutomationAccount,
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, IngestibleRelationship{
		Source:     strings.ToUpper(account.ResourceGroupId),
		SourceType: azure.ResourceGroup,
		TargetType: azure.AutomationAccount,
		Target:     strings.ToUpper(account.Id),
		RelProps:   map[string]any{},
		RelType:    azure.Contains,
	})

	// Enumerate System Assigned Identities
	if account.Identity.PrincipalId != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     strings.ToUpper(account.Id),
			SourceType: azure.AutomationAccount,
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(account.Identity.PrincipalId),
			RelProps:   map[string]any{},
			RelType:    azure.ManagedIdentity,
		})
	}

	// Enumerate User Assigned Identities
	for _, identity := range account.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, IngestibleRelationship{
				Source:     strings.ToUpper(account.Id),
				SourceType: azure.AutomationAccount,
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(identity.PrincipalId),
				RelProps:   map[string]any{},
				RelType:    azure.ManagedIdentity,
			})
		}
	}

	return node, relationships
}

func CanAddSecret(roleDefinitionId string) bool {
	return roleDefinitionId == azure.ApplicationAdministratorRole || roleDefinitionId == azure.CloudApplicationAdministratorRole
}

func GetAddSecretRoleKind(roleDefinitionId string) (graph.Kind, error) {
	switch roleDefinitionId {
	case azure.ApplicationAdministratorRole:
		return azure.AppAdmin, nil
	case azure.CloudApplicationAdministratorRole:
		return azure.CloudAppAdmin, nil
	default:
		// TODO: This should be an error case
		return graph.StringKind(""), fmt.Errorf("invalid get secret role id: %v", roleDefinitionId)
	}
}

func ParseISO8601(datetime string) time.Time {
	if isoTime, err := time.Parse(ISO8601, datetime); err != nil {
		return time.Time{}
	} else {
		return isoTime
	}
}

func KindFromRoleId(roleId string) graph.Kind {
	switch roleId {
	case constants.OwnerRoleID:
		return azure.Owner
	case constants.UserAccessAdminRoleID:
		return azure.UserAccessAdministrator
	case constants.ContributorRoleID:
		return azure.Contributor
	case constants.WebsiteContributorRoleID:
		return azure.WebsiteContributor
	case constants.AutomationContributorRoleID:
		return azure.AutomationContributor
	case constants.LogicAppContributorRoleID:
		return azure.LogicAppContributor
	case constants.VirtualMachineContributorRoleID:
		return azure.VMContributor
	case azure.AKSContributorRole:
		return azure.AKSContributor
	default:
		return graph.StringKind("")
	}
}

func ExtractTypeFromDirectoryObject(directoryObject azure2.DirectoryObject) (objectType graph.Kind, err error) {
	switch directoryObject.Type {
	case enums.EntityGroup:
		return azure.Group, nil
	case enums.EntityUser:
		return azure.User, nil
	case enums.EntityServicePrincipal:
		return azure.ServicePrincipal, nil
	case enums.EntityDevice:
		return azure.Device, nil
	default:
		return nil, errors.New(fmt.Sprintf("invalid type returned from directory object: %s", directoryObject.Type))
	}
}

func getKeyVaultPermissions(data models.KeyVaultAccessPolicy) []graph.Kind {
	var (
		relationships []graph.Kind
	)

	for _, key := range data.Permissions.Certificates {
		if key == KeyVaultPermissionGet {
			relationships = append(relationships, azure.GetCertificates)
			break
		}
	}
	for _, key := range data.Permissions.Keys {
		if key == KeyVaultPermissionGet {
			relationships = append(relationships, azure.GetKeys)
			break
		}
	}
	for _, key := range data.Permissions.Secrets {
		if key == KeyVaultPermissionGet {
			relationships = append(relationships, azure.GetSecrets)
			break
		}
	}
	return relationships
}

func ResourceWithinScope(resource, scope string) bool {
	if strings.EqualFold(resource, scope) {
		return true
	}

	if resourceGroupLevel.MatchString(scope) && strings.HasPrefix(strings.ToLower(resource), strings.ToLower(scope)) {
		return true
	}
	return false
}
