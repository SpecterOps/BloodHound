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
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/bloodhoundad/azurehound/v2/constants"
	"github.com/bloodhoundad/azurehound/v2/enums"
	"github.com/bloodhoundad/azurehound/v2/models"
	azure2 "github.com/bloodhoundad/azurehound/v2/models/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
)

const (
	ISO8601               string = "2006-01-02T15:04:05Z"
	KeyVaultPermissionGet string = "Get"
)

var (
	resourceGroupLevel = regexp.MustCompile(`^[\\w\\d\\-\\/]*/resourceGroups/[0-9a-zA-Z]+$`)
	ErrInvalidType     = errors.New("invalid type returned from directory object")
)

func ConvertAZAppToNode(app models.App, ingestTime time.Time) IngestibleNode {
	return IngestibleNode{
		PropertyMap: map[string]any{
			common.Name.String():           strings.ToUpper(fmt.Sprintf("%s@%s", app.DisplayName, app.PublisherDomain)),
			common.Description.String():    app.Description,
			common.DisplayName.String():    app.DisplayName,
			common.LastCollected.String():  ingestTime,
			common.LastSeen.String():       ingestTime,
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
	return []IngestibleRelationship{
		NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(app.TenantId),
				SourceType: azure.Tenant,
			},
			IngestibleTarget{
				TargetType: azure.App,
				Target:     strings.ToUpper(app.AppId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		),
	}
}

func ConvertAZDeviceToNode(device models.Device, ingestTime time.Time) IngestibleNode {
	return IngestibleNode{
		PropertyMap: map[string]any{
			common.Name.String():                  strings.ToUpper(fmt.Sprintf("%s@%s", device.DisplayName, device.TenantName)),
			common.DisplayName.String():           device.DisplayName,
			common.OperatingSystem.String():       device.OperatingSystem,
			azure.DeviceID.String():               device.DeviceId,
			azure.OperatingSystemVersion.String(): device.OperatingSystemVersion,
			azure.TrustType.String():              device.TrustType,
			azure.TenantID.String():               strings.ToUpper(device.TenantId),
			common.LastCollected.String():         ingestTime,
		},
		ObjectID: strings.ToUpper(device.Id),
		Label:    azure.Device,
	}
}

func ConvertAZDeviceRelationships(device models.Device) []IngestibleRelationship {
	return []IngestibleRelationship{
		NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(device.TenantId),
				SourceType: azure.Tenant,
			},
			IngestibleTarget{
				TargetType: azure.Device,
				Target:     strings.ToUpper(device.Id),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		),
	}
}

func ConvertAZVMScaleSetToNode(scaleSet models.VMScaleSet, ingestTime time.Time) IngestibleNode {
	return IngestibleNode{
		ObjectID: strings.ToUpper(scaleSet.Id),
		PropertyMap: map[string]any{
			common.Name.String():          strings.ToUpper(scaleSet.Name),
			azure.TenantID.String():       strings.ToUpper(scaleSet.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Label: azure.VMScaleSet,
	}
}

func ConvertAZVMScaleSetRelationships(scaleSet models.VMScaleSet) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(scaleSet.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.VMScaleSet,
			Target:     strings.ToUpper(scaleSet.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if scaleSet.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(scaleSet.Id),
				SourceType: azure.VMScaleSet,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(scaleSet.Identity.PrincipalId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.ManagedIdentity,
			},
		))
	}

	// Enumerate User Assigned Identities
	for _, identity := range scaleSet.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(scaleSet.Id),
					SourceType: azure.VMScaleSet,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(identity.PrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.ManagedIdentity,
				},
			))
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
				relationships = append(relationships, NewIngestibleRelationship(
					IngestibleSource{
						Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
						SourceType: azure.Entity,
					},
					IngestibleTarget{
						TargetType: azure.VMScaleSet,
						Target:     strings.ToUpper(data.ObjectId),
					},
					IngestibleRel{
						RelProps: map[string]any{},
						RelType:  KindFromRoleId(raw.RoleDefinitionId),
					},
				))
			}
		}
	}

	return relationships
}

func ConvertAzureOwnerToRel(directoryObject azure2.DirectoryObject, ownerType graph.Kind, targetType graph.Kind, targetId string) IngestibleRelationship {
	return NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(directoryObject.Id),
			SourceType: ownerType,
		},
		IngestibleTarget{
			TargetType: targetType,
			Target:     strings.ToUpper(targetId),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Owns,
		},
	)
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
		return NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.PrincipalId.String()),
				SourceType: azure.ServicePrincipal,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(data.ResourceId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  appRoleKind,
			},
		)
	}

	return NewIngestibleRelationship(
		IngestibleSource{
			Source:     "",
			SourceType: nil,
		},
		IngestibleTarget{
			TargetType: nil,
			Target:     "",
		},
		IngestibleRel{
			RelProps: nil,
			RelType:  nil,
		},
	)
}

func ConvertAzureFunctionAppToNode(data models.FunctionApp, ingestTime time.Time) IngestibleNode {
	return IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():          strings.ToUpper(data.Name),
			azure.TenantID.String():       strings.ToUpper(data.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Label: azure.FunctionApp,
	}
}

func ConvertAzureFunctionAppToRels(data models.FunctionApp) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.FunctionApp,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.Id),
				SourceType: azure.FunctionApp,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(data.Identity.PrincipalId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.ManagedIdentity,
			},
		))
	}

	// Enumerate User Assigned Identities
	for _, identity := range data.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(data.Id),
					SourceType: azure.FunctionApp,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(identity.PrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.ManagedIdentity,
				},
			))
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
				relationships = append(relationships, NewIngestibleRelationship(
					IngestibleSource{
						Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
						SourceType: azure.Entity,
					},
					IngestibleTarget{
						TargetType: azure.FunctionApp,
						Target:     strings.ToUpper(data.ObjectId),
					},
					IngestibleRel{
						RelProps: map[string]any{},
						RelType:  KindFromRoleId(raw.RoleDefinitionId),
					},
				))
			}
		}
	}

	return relationships
}

func ConvertAzureGroupToNode(data models.Group, ingestTime time.Time) IngestibleNode {
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
			common.LastCollected.String():     ingestTime,
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
	return NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.TenantId),
			SourceType: azure.Tenant,
		},
		IngestibleTarget{
			TargetType: azure.Group,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	)
}

func ConvertAzureGroupMembersToRels(data models.GroupMembers) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Members {
		var (
			member azure2.DirectoryObject
		)
		if err := json.Unmarshal(raw.Member, &member); err != nil {
			slog.Error(fmt.Sprintf(SerialError, "azure group member", err))
		} else if memberType, err := ExtractTypeFromDirectoryObject(member); errors.Is(err, ErrInvalidType) {
			slog.Warn(fmt.Sprintf(ExtractError, err))
		} else if err != nil {
			slog.Error(fmt.Sprintf(ExtractError, err))
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(member.Id),
					SourceType: memberType,
				},
				IngestibleTarget{
					TargetType: azure.Group,
					Target:     strings.ToUpper(data.GroupId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.MemberOf,
				},
			))
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
			slog.Error(fmt.Sprintf(SerialError, "azure group owner", err))
		} else if ownerType, err := ExtractTypeFromDirectoryObject(owner); errors.Is(err, ErrInvalidType) {
			slog.Warn(fmt.Sprintf(ExtractError, err))
		} else if err != nil {
			slog.Error(fmt.Sprintf(ExtractError, err))
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(owner.Id),
					SourceType: ownerType,
				},
				IngestibleTarget{
					TargetType: azure.Group,
					Target:     strings.ToUpper(data.GroupId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Owns,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureKeyVault(data models.KeyVault, ingestTime time.Time) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.Name.String():                   strings.ToUpper(data.Name),
				azure.EnableRBACAuthorization.String(): data.Properties.EnableRbacAuthorization,
				azure.TenantID.String():                strings.ToUpper(data.TenantId),
				common.LastCollected.String():          ingestTime,
			},
			Label: azure.KeyVault,
		},
		NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.ResourceGroup),
				SourceType: azure.ResourceGroup,
			},
			IngestibleTarget{
				TargetType: azure.KeyVault,
				Target:     strings.ToUpper(data.Id),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
}

func ConvertAzureKeyVaultAccessPolicy(data models.KeyVaultAccessPolicy) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, relType := range getKeyVaultPermissions(data) {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.ObjectId),
				SourceType: azure.Entity,
			},
			IngestibleTarget{
				TargetType: azure.KeyVault,
				Target:     strings.ToUpper(data.KeyVaultId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  relType,
			},
		))
	}

	return relationships
}

func ConvertAzureKeyVaultContributor(data models.KeyVaultContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Contributors {
		if data.KeyVaultId == raw.Contributor.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.Contributor.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.KeyVault,
					Target:     strings.ToUpper(data.KeyVaultId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Contributor,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureKeyVaultKVContributor(data models.KeyVaultKVContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.KVContributors {
		if data.KeyVaultId == raw.KVContributor.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.KVContributor.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.KeyVault,
					Target:     strings.ToUpper(data.KeyVaultId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.KeyVaultContributor,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureKeyVaultOwnerToRels(data models.KeyVaultOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Owners {
		if data.KeyVaultId == raw.Owner.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.Owner.Properties.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.KeyVault,
					Target:     strings.ToUpper(data.KeyVaultId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Owner,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureKeyVaultUserAccessAdminToRels(data models.KeyVaultUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if data.KeyVaultId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.KeyVault,
					Target:     strings.ToUpper(data.KeyVaultId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.UserAccessAdministrator,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureManagementGroupDescendantToRel(data azure2.DescendantInfo) IngestibleRelationship {
	return NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.Properties.Parent.Id),
			SourceType: azure.ManagementGroup,
		},
		IngestibleTarget{
			TargetType: azure.Entity,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	)
}

func ConvertAzureManagementGroupOwnerToRels(data models.ManagementGroupOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		if data.ManagementGroupId == raw.Owner.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.Owner.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.ManagementGroup,
					Target:     strings.ToUpper(data.ManagementGroupId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Owner,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureManagementGroupUserAccessAdminToRels(data models.ManagementGroupUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if data.ManagementGroupId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.UserAccessAdmin.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.ManagementGroup,
					Target:     strings.ToUpper(data.ManagementGroupId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.UserAccessAdministrator,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureManagementGroup(data models.ManagementGroup, ingestTime time.Time) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.Name.String():          strings.ToUpper(fmt.Sprintf("%s@%s", data.Properties.DisplayName, data.TenantName)),
				azure.TenantID.String():       strings.ToUpper(data.TenantId),
				common.LastCollected.String(): ingestTime,
			},
			Label: azure.ManagementGroup,
		}, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.TenantId),
				SourceType: azure.Tenant,
			},
			IngestibleTarget{
				TargetType: azure.ManagementGroup,
				Target:     strings.ToUpper(data.Id),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
}

func ConvertAzureResourceGroup(data models.ResourceGroup, ingestTime time.Time) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.Name.String():          strings.ToUpper(data.Name),
				azure.TenantID.String():       strings.ToUpper(data.TenantId),
				common.LastCollected.String(): ingestTime,
			},
			Label: azure.ResourceGroup,
		}, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.SubscriptionId),
				SourceType: azure.Subscription,
			},
			IngestibleTarget{
				TargetType: azure.ResourceGroup,
				Target:     strings.ToUpper(data.Id),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
}

func ConvertAzureResourceGroupOwnerToRels(data models.ResourceGroupOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		if data.ResourceGroupId == raw.Owner.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.Owner.Properties.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.ResourceGroup,
					Target:     strings.ToUpper(data.ResourceGroupId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Owner,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureResourceGroupUserAccessAdminToRels(data models.ResourceGroupUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if data.ResourceGroupId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.ResourceGroup,
					Target:     strings.ToUpper(data.ResourceGroupId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.UserAccessAdministrator,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureRole(data models.Role, ingestTime time.Time) (IngestibleNode, IngestibleRelationship) {
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
				common.LastCollected.String(): ingestTime,
			},
			Label: azure.Role,
		}, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.TenantId),
				SourceType: azure.Tenant,
			},
			IngestibleTarget{
				TargetType: azure.Role,
				Target:     roleObjectId,
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
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

	if CanAddSecret(roleAssignment.RoleDefinitionId) && roleAssignment.DirectoryScopeId != "/" {
		if relType, err := GetAddSecretRoleKind(roleAssignment.RoleDefinitionId); err != nil {
			slog.Error(fmt.Sprintf("Error processing role assignment for role %s: %v", roleObjectId, err))
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(roleAssignment.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.Entity,
					Target:     scope,
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  relType,
				},
			))
		}
	} else {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(roleAssignment.PrincipalId),
				SourceType: azure.Entity,
			},
			IngestibleTarget{
				TargetType: azure.Role,
				Target:     roleObjectId,
			},
			IngestibleRel{
				RelProps: map[string]any{
					azure.Scope.String(): scope,
				},
				RelType: azure.HasRole,
			},
		))
	}

	return relationships
}

func ConvertAzureServicePrincipal(data models.ServicePrincipal, ingestTime time.Time) ([]IngestibleNode, []IngestibleRelationship) {
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
			azure.LoginURL.String():               data.LoginUrl,
			azure.ServicePrincipalType.String():   data.ServicePrincipalType,
			azure.TenantID.String():               strings.ToUpper(data.TenantId),
			common.LastCollected.String():         ingestTime,
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

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.AppId),
			SourceType: azure.App,
		},
		IngestibleTarget{
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.RunsAs,
		},
	))

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.TenantId),
			SourceType: azure.Tenant,
		},
		IngestibleTarget{
			TargetType: azure.ServicePrincipal,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	return nodes, relationships
}

func ConvertAzureLogicApp(logicApp models.LogicApp, ingestTime time.Time) (IngestibleNode, []IngestibleRelationship) {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(logicApp.Id),
		PropertyMap: map[string]any{
			common.Name.String():          strings.ToUpper(logicApp.Name),
			azure.TenantID.String():       strings.ToUpper(logicApp.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Label: azure.LogicApp,
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(logicApp.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.LogicApp,
			Target:     strings.ToUpper(logicApp.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if logicApp.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(logicApp.Id),
				SourceType: azure.LogicApp,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(logicApp.Identity.PrincipalId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.ManagedIdentity,
			},
		))
	}

	// Enumerate User Assigned Identities
	for _, identity := range logicApp.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(logicApp.Id),
					SourceType: azure.LogicApp,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(identity.PrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.ManagedIdentity,
				},
			))
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
				relationships = append(relationships, NewIngestibleRelationship(
					IngestibleSource{
						Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
						SourceType: azure.Entity,
					},
					IngestibleTarget{
						TargetType: azure.LogicApp,
						Target:     strings.ToUpper(roleAssignment.ObjectId),
					},
					IngestibleRel{
						RelProps: map[string]any{},
						RelType:  KindFromRoleId(raw.RoleDefinitionId),
					},
				))
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
			slog.Error(fmt.Sprintf(SerialError, "azure service principal owner", err))
		} else if ownerType, err := ExtractTypeFromDirectoryObject(owner); errors.Is(err, ErrInvalidType) {
			slog.Warn(fmt.Sprintf(ExtractError, err))
		} else if err != nil {
			slog.Error(fmt.Sprintf(ExtractError, err))
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(owner.Id),
					SourceType: ownerType,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(data.ServicePrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Owns,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureSubscription(data azure2.Subscription, ingestTime time.Time) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: strings.ToUpper(data.Id),
			PropertyMap: map[string]any{
				common.DisplayName.String():   data.DisplayName,
				common.ObjectID.String():      data.SubscriptionId,
				common.Name.String():          strings.ToUpper(data.DisplayName),
				azure.TenantID.String():       strings.ToUpper(data.TenantId),
				common.LastCollected.String(): ingestTime,
			},
			Label: azure.Subscription,
		},
		NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.TenantId),
				SourceType: azure.Tenant,
			},
			IngestibleTarget{
				TargetType: azure.Subscription,
				Target:     strings.ToUpper(data.Id),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
}

func ConvertAzureSubscriptionOwnerToRels(data models.SubscriptionOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.Owners {
		if data.SubscriptionId == raw.Owner.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.Owner.Properties.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.Subscription,
					Target:     strings.ToUpper(data.SubscriptionId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Owner,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureSubscriptionUserAccessAdminToRels(data models.SubscriptionUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.UserAccessAdmins {
		if data.SubscriptionId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.Subscription,
					Target:     strings.ToUpper(data.SubscriptionId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.UserAccessAdministrator,
				},
			))
		}
	}

	return relationships
}

func ConvertAzureTenantToNode(data models.Tenant, ingestTime time.Time) IngestibleNode {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.TenantId),
		PropertyMap: map[string]any{
			common.DisplayName.String():   data.DisplayName,
			common.ObjectID.String():      data.Id,
			common.Name.String():          strings.ToUpper(data.DisplayName),
			azure.TenantID.String():       strings.ToUpper(data.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Label: azure.Tenant,
	}

	if data.Collected {
		node.PropertyMap["collected"] = true
	}

	return node
}

// ConvertAzureUser returns the basic node, the on prem node and then the ingestible contains relationship
func ConvertAzureUser(data models.User, ingestTime time.Time) (IngestibleNode, IngestibleNode, IngestibleRelationship) {
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
				common.LastCollected.String():    ingestTime,
			},
			Label: azure.User,
		}, onPremNode, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.TenantId),
				SourceType: azure.Tenant,
			},
			IngestibleTarget{
				TargetType: azure.User,
				Target:     strings.ToUpper(data.Id),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
}

func ConvertAzureVirtualMachine(data models.VirtualMachine, ingestTime time.Time) (IngestibleNode, []IngestibleRelationship) {
	relationships := make([]IngestibleRelationship, 0)
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():            strings.ToUpper(data.Name),
			common.ObjectID.String():        data.Properties.VMId,
			common.OperatingSystem.String(): data.Properties.StorageProfile.OSDisk.OSType,
			azure.TenantID.String():         strings.ToUpper(data.TenantId),
			common.LastCollected.String():   ingestTime,
		},
		Label: azure.VM,
	}

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.VM,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.Id),
				SourceType: azure.VM,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(data.Identity.PrincipalId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.ManagedIdentity,
			},
		))
	}

	// Enumerate User Assigned Identities
	for _, identity := range data.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(data.Id),
					SourceType: azure.VM,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(identity.PrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.ManagedIdentity,
				},
			))
		}
	}

	return node, relationships
}

func ConvertAzureVirtualMachineAdminLoginToRels(data models.VirtualMachineAdminLogins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.AdminLogins {
		if ResourceWithinScope(data.VirtualMachineId, raw.AdminLogin.Properties.Scope) {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.AdminLogin.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.VM,
					Target:     strings.ToUpper(data.VirtualMachineId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.VMAdminLogin,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineAvereContributorToRels(data models.VirtualMachineAvereContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.AvereContributors {
		if data.VirtualMachineId == raw.AvereContributor.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.AvereContributor.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.VM,
					Target:     strings.ToUpper(data.VirtualMachineId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.AvereContributor,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineContributorToRels(data models.VirtualMachineContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Contributors {
		if ResourceWithinScope(data.VirtualMachineId, raw.Contributor.Properties.Scope) {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.Contributor.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.VM,
					Target:     strings.ToUpper(data.VirtualMachineId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Contributor,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineVMContributorToRels(data models.VirtualMachineVMContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.VMContributors {
		if ResourceWithinScope(data.VirtualMachineId, raw.VMContributor.Properties.Scope) {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.VMContributor.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.VM,
					Target:     strings.ToUpper(data.VirtualMachineId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.VMContributor,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineOwnerToRels(data models.VirtualMachineOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		if ResourceWithinScope(data.VirtualMachineId, raw.Owner.Properties.Scope) {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.Owner.GetPrincipalId()),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.VM,
					Target:     strings.ToUpper(data.VirtualMachineId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.Owner,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureVirtualMachineUserAccessAdminToRels(data models.VirtualMachineUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if ResourceWithinScope(data.VirtualMachineId, raw.UserAccessAdmin.Properties.Scope) {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					SourceType: azure.Entity,
				},
				IngestibleTarget{
					TargetType: azure.VM,
					Target:     strings.ToUpper(data.VirtualMachineId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.UserAccessAdministrator,
				},
			))
		}
	}
	return relationships
}

func ConvertAzureManagedCluster(data models.ManagedCluster, nodeResourceGroupID string, ingestTime time.Time) (IngestibleNode, []IngestibleRelationship) {
	relationships := make([]IngestibleRelationship, 0)
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():               strings.ToUpper(data.Name),
			azure.TenantID.String():            strings.ToUpper(data.TenantId),
			azure.NodeResourceGroupID.String(): strings.ToUpper(nodeResourceGroupID),
			common.LastCollected.String():      ingestTime,
		},
		Label: azure.ManagedCluster,
	}

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.ManagedCluster,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.Id),
			SourceType: azure.ManagedCluster,
		},
		IngestibleTarget{
			TargetType: azure.ResourceGroup,
			Target:     strings.ToUpper(nodeResourceGroupID),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.NodeResourceGroup,
		},
	))

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
				relationships = append(relationships, NewIngestibleRelationship(
					IngestibleSource{
						Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
						SourceType: azure.Entity,
					},
					IngestibleTarget{
						TargetType: azure.ManagedCluster,
						Target:     strings.ToUpper(data.ObjectId),
					},
					IngestibleRel{
						RelProps: map[string]any{},
						RelType:  KindFromRoleId(raw.RoleDefinitionId),
					},
				))
			}
		}
	}
	return relationships
}

func ConvertAzureContainerRegistry(data models.ContainerRegistry, ingestTime time.Time) (IngestibleNode, []IngestibleRelationship) {
	relationships := make([]IngestibleRelationship, 0)
	node := IngestibleNode{
		ObjectID: strings.ToUpper(data.Id),
		PropertyMap: map[string]any{
			common.Name.String():          strings.ToUpper(data.Name),
			azure.TenantID.String():       strings.ToUpper(data.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Label: azure.ContainerRegistry,
	}

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(data.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.ContainerRegistry,
			Target:     strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(data.Id),
				SourceType: azure.ContainerRegistry,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(data.Identity.PrincipalId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.ManagedIdentity,
			},
		))
	}

	// Enumerate User Assigned Identities
	for _, identity := range data.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(data.Id),
					SourceType: azure.ContainerRegistry,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(identity.PrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.ManagedIdentity,
				},
			))
		}
	}

	return node, relationships
}

func ConvertAzureWebApp(webApp models.WebApp, ingestTime time.Time) (IngestibleNode, []IngestibleRelationship) {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(webApp.Id),
		PropertyMap: map[string]any{
			common.Name.String():          strings.ToUpper(webApp.Name),
			azure.TenantID.String():       strings.ToUpper(webApp.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Label: azure.WebApp,
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(webApp.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.WebApp,
			Target:     strings.ToUpper(webApp.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if webApp.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(webApp.Id),
				SourceType: azure.WebApp,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(webApp.Identity.PrincipalId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.ManagedIdentity,
			},
		))
	}

	// Enumerate User Assigned Identities
	for _, identity := range webApp.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(webApp.Id),
					SourceType: azure.WebApp,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(identity.PrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.ManagedIdentity,
				},
			))
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
				relationships = append(relationships, NewIngestibleRelationship(
					IngestibleSource{
						Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
						SourceType: azure.Entity,
					},
					IngestibleTarget{
						TargetType: azure.AutomationAccount,
						Target:     strings.ToUpper(roleAssignments.ObjectId),
					},
					IngestibleRel{
						RelProps: map[string]any{},
						RelType:  KindFromRoleId(raw.RoleDefinitionId),
					},
				))
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
				relationships = append(relationships, NewIngestibleRelationship(
					IngestibleSource{
						Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
						SourceType: azure.Entity,
					},
					IngestibleTarget{
						TargetType: azure.ContainerRegistry,
						Target:     strings.ToUpper(roleAssignment.ObjectId),
					},
					IngestibleRel{
						RelProps: map[string]any{},
						RelType:  KindFromRoleId(raw.RoleDefinitionId),
					},
				))
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
				relationships = append(relationships, NewIngestibleRelationship(
					IngestibleSource{
						Source:     strings.ToUpper(raw.Assignee.GetPrincipalId()),
						SourceType: azure.Entity,
					},
					IngestibleTarget{
						TargetType: azure.WebApp,
						Target:     strings.ToUpper(roleAssignment.ObjectId),
					},
					IngestibleRel{
						RelProps: map[string]any{},
						RelType:  KindFromRoleId(raw.RoleDefinitionId),
					},
				))
			}
		}
	}

	return relationships
}

func ConvertAzureAutomationAccount(account models.AutomationAccount, ingestTime time.Time) (IngestibleNode, []IngestibleRelationship) {
	node := IngestibleNode{
		ObjectID: strings.ToUpper(account.Id),
		PropertyMap: map[string]any{
			common.Name.String():          strings.ToUpper(account.Name),
			azure.TenantID.String():       strings.ToUpper(account.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Label: azure.AutomationAccount,
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     strings.ToUpper(account.ResourceGroupId),
			SourceType: azure.ResourceGroup,
		},
		IngestibleTarget{
			TargetType: azure.AutomationAccount,
			Target:     strings.ToUpper(account.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if account.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     strings.ToUpper(account.Id),
				SourceType: azure.AutomationAccount,
			},
			IngestibleTarget{
				TargetType: azure.ServicePrincipal,
				Target:     strings.ToUpper(account.Identity.PrincipalId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.ManagedIdentity,
			},
		))
	}

	// Enumerate User Assigned Identities
	for _, identity := range account.Identity.UserAssignedIdentities {
		if identity.ClientId != "" {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     strings.ToUpper(account.Id),
					SourceType: azure.AutomationAccount,
				},
				IngestibleTarget{
					TargetType: azure.ServicePrincipal,
					Target:     strings.ToUpper(identity.PrincipalId),
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  azure.ManagedIdentity,
				},
			))
		}
	}

	return node, relationships
}

// ConvertAzureRoleManagementPolicyAssignment will create, or update the properties if it exists, an AZRole of a tenant
// with the supplied RoleManagementPolicyAssignment properties
// If EndUserAssignmentGroupApprovers contains GUIDs: an edge will be created from each group to the created AZRole
// If EndUserAssignmentUsersApprovers contains GUIDs: an edge will be created from each user to the created AZRole
// If both lists are empty: an edge will be created from the tenant's PrivilegedRoleAdministratorRole to the created AZRole
func ConvertAzureRoleManagementPolicyAssignment(policyAssignment models.RoleManagementPolicyAssignment) (IngestibleNode, []IngestibleRelationship) {
	var (
		rels             = make([]IngestibleRelationship, 0)
		combinedObjectId = strings.ToUpper(strings.Join([]string{policyAssignment.RoleDefinitionId, policyAssignment.TenantId}, "@"))
	)

	// Format the incoming user and group ids to uppercase string before creating our nodes
	for i, _ := range policyAssignment.EndUserAssignmentGroupApprovers {
		policyAssignment.EndUserAssignmentGroupApprovers[i] = strings.ToUpper(policyAssignment.EndUserAssignmentGroupApprovers[i])
	}

	for i, _ := range policyAssignment.EndUserAssignmentUserApprovers {
		policyAssignment.EndUserAssignmentUserApprovers[i] = strings.ToUpper(policyAssignment.EndUserAssignmentUserApprovers[i])
	}

	// We will want to create or update any existing AZRole node that matches the combinedObjectId
	// If the node exists, we want to add the new properties to the node
	targetAZRole := IngestibleNode{
		ObjectID: combinedObjectId,
		Label:    azure.Role,
		PropertyMap: map[string]any{
			azure.RoleDefinitionId.String():                                  strings.ToUpper(policyAssignment.RoleDefinitionId),
			azure.TenantID.String():                                          strings.ToUpper(policyAssignment.TenantId),
			azure.EndUserAssignmentRequiresApproval.String():                 policyAssignment.EndUserAssignmentRequiresApproval,
			azure.EndUserAssignmentRequiresCAPAuthenticationContext.String(): policyAssignment.EndUserAssignmentRequiresCAPAuthenticationContext,
			azure.EndUserAssignmentUserApprovers.String():                    policyAssignment.EndUserAssignmentUserApprovers,
			azure.EndUserAssignmentGroupApprovers.String():                   policyAssignment.EndUserAssignmentGroupApprovers,
			azure.EndUserAssignmentRequiresMFA.String():                      policyAssignment.EndUserAssignmentRequiresMFA,
			azure.EndUserAssignmentRequiresJustification.String():            policyAssignment.EndUserAssignmentRequiresJustification,
			azure.EndUserAssignmentRequiresTicketInformation.String():        policyAssignment.EndUserAssignmentRequiresTicketInformation,
		},
	}

	if !policyAssignment.EndUserAssignmentRequiresApproval {
		// We cannot create the edge or nodes if the assignment does not require approval
		return targetAZRole, rels
	}

	if len(policyAssignment.EndUserAssignmentUserApprovers) > 0 {
		// Create an edge for each user with that allow approvals to the target role
		for _, approver := range policyAssignment.EndUserAssignmentUserApprovers {
			rels = append(rels, NewIngestibleRelationship(IngestibleSource{
				Source:     strings.ToUpper(approver),
				SourceType: azure.User,
			}, IngestibleTarget{
				Target:     targetAZRole.ObjectID,
				TargetType: targetAZRole.Label,
			}, IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.AZRoleApprover,
			}))
		}
	}

	if len(policyAssignment.EndUserAssignmentGroupApprovers) > 0 {
		// Create an edge for each group with that allow approvals to the target role
		for _, approver := range policyAssignment.EndUserAssignmentGroupApprovers {
			rels = append(rels, NewIngestibleRelationship(IngestibleSource{
				Source:     strings.ToUpper(approver),
				SourceType: azure.Group,
			}, IngestibleTarget{
				Target:     targetAZRole.ObjectID,
				TargetType: targetAZRole.Label,
			}, IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.AZRoleApprover,
			}))
		}
	}

	if len(policyAssignment.EndUserAssignmentUserApprovers) == 0 && len(policyAssignment.EndUserAssignmentGroupApprovers) == 0 {
		// No users or groups were attached to the policy, we will create the edge from the tenant's PrivilegedRoleAdministratorRole Role node to the target role
		combinedObjectId := strings.ToUpper(strings.Join([]string{azure.PrivilegedRoleAdministratorRole, policyAssignment.TenantId}, "@"))

		rels = append(rels, NewIngestibleRelationship(IngestibleSource{
			Source:     strings.ToUpper(combinedObjectId),
			SourceType: azure.Role,
		}, IngestibleTarget{
			Target:     targetAZRole.ObjectID,
			TargetType: targetAZRole.Label,
		}, IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.AZRoleApprover,
		}))
	}

	return targetAZRole, rels
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
		return nil, fmt.Errorf("%w: %s", ErrInvalidType, directoryObject.Type)
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
