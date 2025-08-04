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
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
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
		Labels:   []graph.Kind{azure.App},
	}
}

func ConvertAZAppRelationships(app models.App) []IngestibleRelationship {
	return []IngestibleRelationship{
		NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(app.TenantId),
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.App,
				Value: strings.ToUpper(app.AppId),
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
		Labels:   []graph.Kind{azure.Device},
	}
}

func ConvertAZDeviceRelationships(device models.Device) []IngestibleRelationship {
	return []IngestibleRelationship{
		NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(device.TenantId),
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.Device,
				Value: strings.ToUpper(device.Id),
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
		Labels: []graph.Kind{azure.VMScaleSet},
	}
}

func ConvertAZVMScaleSetRelationships(scaleSet models.VMScaleSet) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(scaleSet.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.VMScaleSet,
			Value: strings.ToUpper(scaleSet.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if scaleSet.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(scaleSet.Id),
				Kind:  azure.VMScaleSet,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(scaleSet.Identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(scaleSet.Id),
					Kind:  azure.VMScaleSet,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(identity.PrincipalId),
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
					IngestibleEndpoint{
						Value: strings.ToUpper(raw.Assignee.GetPrincipalId()),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.VMScaleSet,
						Value: strings.ToUpper(data.ObjectId),
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
		IngestibleEndpoint{
			Value: strings.ToUpper(directoryObject.Id),
			Kind:  ownerType,
		},
		IngestibleEndpoint{
			Kind:  targetType,
			Value: strings.ToUpper(targetId),
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
		Labels:   []graph.Kind{azure.ServicePrincipal},
	})

	nodes = append(nodes, IngestibleNode{
		PropertyMap: map[string]any{
			common.DisplayName.String(): strings.ToUpper(data.ResourceDisplayName),
			azure.TenantID.String():     strings.ToUpper(data.TenantId),
		},
		ObjectID: strings.ToUpper(data.ResourceId),
		Labels:   []graph.Kind{azure.ServicePrincipal},
	})

	return nodes
}

func ConvertAzureAppRoleAssignmentToRel(data models.AppRoleAssignment) IngestibleRelationship {
	if appRoleKind, hasAppRoleKind := azure.RelationshipKindByAppRoleID[strings.ToLower(data.AppRoleId.String())]; hasAppRoleKind {
		return NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.PrincipalId.String()),
				Kind:  azure.ServicePrincipal,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(data.ResourceId),
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  appRoleKind,
			},
		)
	}

	return NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: "",
			Kind:  nil,
		},
		IngestibleEndpoint{
			Kind:  nil,
			Value: "",
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
		Labels: []graph.Kind{azure.FunctionApp},
	}
}

func ConvertAzureFunctionAppToRels(data models.FunctionApp) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.FunctionApp,
			Value: strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.Id),
				Kind:  azure.FunctionApp,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(data.Identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(data.Id),
					Kind:  azure.FunctionApp,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(identity.PrincipalId),
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
					IngestibleEndpoint{
						Value: strings.ToUpper(raw.Assignee.GetPrincipalId()),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.FunctionApp,
						Value: strings.ToUpper(data.ObjectId),
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
		Labels: []graph.Kind{azure.Group},
	}
}

func ConvertAzureGroupToOnPremisesNode(data models.Group) IngestibleNode {
	if data.OnPremisesSecurityIdentifier != "" {
		return IngestibleNode{
			ObjectID:    strings.ToUpper(data.OnPremisesSecurityIdentifier),
			PropertyMap: map[string]any{},
			Labels:      []graph.Kind{ad.Group},
		}
	}

	return IngestibleNode{
		ObjectID:    "",
		PropertyMap: nil,
		Labels:      nil,
	}
}

func ConvertAzureGroupToRel(data models.Group) IngestibleRelationship {
	return NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.TenantId),
			Kind:  azure.Tenant,
		},
		IngestibleEndpoint{
			Kind:  azure.Group,
			Value: strings.ToUpper(data.Id),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(member.Id),
					Kind:  memberType,
				},
				IngestibleEndpoint{
					Kind:  azure.Group,
					Value: strings.ToUpper(data.GroupId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(owner.Id),
					Kind:  ownerType,
				},
				IngestibleEndpoint{
					Kind:  azure.Group,
					Value: strings.ToUpper(data.GroupId),
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
			Labels: []graph.Kind{azure.KeyVault},
		},
		NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.ResourceGroup),
				Kind:  azure.ResourceGroup,
			},
			IngestibleEndpoint{
				Kind:  azure.KeyVault,
				Value: strings.ToUpper(data.Id),
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
			IngestibleEndpoint{
				Value: strings.ToUpper(data.ObjectId),
				Kind:  azure.Entity,
			},
			IngestibleEndpoint{
				Kind:  azure.KeyVault,
				Value: strings.ToUpper(data.KeyVaultId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.Contributor.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: strings.ToUpper(data.KeyVaultId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.KVContributor.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: strings.ToUpper(data.KeyVaultId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.Owner.Properties.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: strings.ToUpper(data.KeyVaultId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: strings.ToUpper(data.KeyVaultId),
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
		IngestibleEndpoint{
			Value: strings.ToUpper(data.Properties.Parent.Id),
			Kind:  azure.ManagementGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.Entity,
			Value: strings.ToUpper(data.Id),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.Owner.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ManagementGroup,
					Value: strings.ToUpper(data.ManagementGroupId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.UserAccessAdmin.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ManagementGroup,
					Value: strings.ToUpper(data.ManagementGroupId),
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
				azure.TenantID.String():       strings.ToUpper(data.TenantId),
				common.LastCollected.String(): ingestTime,
				common.DisplayName.String():   strings.ToUpper(data.Properties.DisplayName),
			},
			Labels: []graph.Kind{azure.ManagementGroup},
		}, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.TenantId),
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.ManagementGroup,
				Value: strings.ToUpper(data.Id),
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
			Labels: []graph.Kind{azure.ResourceGroup},
		}, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.SubscriptionId),
				Kind:  azure.Subscription,
			},
			IngestibleEndpoint{
				Kind:  azure.ResourceGroup,
				Value: strings.ToUpper(data.Id),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.Owner.Properties.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ResourceGroup,
					Value: strings.ToUpper(data.ResourceGroupId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ResourceGroup,
					Value: strings.ToUpper(data.ResourceGroupId),
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
			Labels: []graph.Kind{azure.Role},
		}, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.TenantId),
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.Role,
				Value: roleObjectId,
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
				IngestibleEndpoint{
					Value: strings.ToUpper(roleAssignment.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.Entity,
					Value: scope,
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  relType,
				},
			))
		}
	} else {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(roleAssignment.PrincipalId),
				Kind:  azure.Entity,
			},
			IngestibleEndpoint{
				Kind:  azure.Role,
				Value: roleObjectId,
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
		Labels: []graph.Kind{azure.ServicePrincipal},
	})

	nodes = append(nodes, IngestibleNode{
		ObjectID: strings.ToUpper(data.AppId),
		PropertyMap: map[string]any{
			common.DisplayName.String(): data.AppDisplayName,
			azure.TenantID.String():     strings.ToUpper(data.AppOwnerOrganizationId),
		},
		Labels: []graph.Kind{azure.App},
	})

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.AppId),
			Kind:  azure.App,
		},
		IngestibleEndpoint{
			Kind:  azure.ServicePrincipal,
			Value: strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.RunsAs,
		},
	))

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.TenantId),
			Kind:  azure.Tenant,
		},
		IngestibleEndpoint{
			Kind:  azure.ServicePrincipal,
			Value: strings.ToUpper(data.Id),
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
		Labels: []graph.Kind{azure.LogicApp},
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(logicApp.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.LogicApp,
			Value: strings.ToUpper(logicApp.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if logicApp.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(logicApp.Id),
				Kind:  azure.LogicApp,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(logicApp.Identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(logicApp.Id),
					Kind:  azure.LogicApp,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(identity.PrincipalId),
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
					IngestibleEndpoint{
						Value: strings.ToUpper(raw.Assignee.GetPrincipalId()),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.LogicApp,
						Value: strings.ToUpper(roleAssignment.ObjectId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(owner.Id),
					Kind:  ownerType,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(data.ServicePrincipalId),
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
			Labels: []graph.Kind{azure.Subscription},
		},
		NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.TenantId),
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.Subscription,
				Value: strings.ToUpper(data.Id),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.Owner.Properties.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.Subscription,
					Value: strings.ToUpper(data.SubscriptionId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.Subscription,
					Value: strings.ToUpper(data.SubscriptionId),
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
		Labels: []graph.Kind{azure.Tenant},
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
			Labels:      []graph.Kind{ad.User},
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
			Labels: []graph.Kind{azure.User},
		}, onPremNode, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.TenantId),
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.User,
				Value: strings.ToUpper(data.Id),
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
		Labels: []graph.Kind{azure.VM},
	}

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.VM,
			Value: strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.Id),
				Kind:  azure.VM,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(data.Identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(data.Id),
					Kind:  azure.VM,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.AdminLogin.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: strings.ToUpper(data.VirtualMachineId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.AvereContributor.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: strings.ToUpper(data.VirtualMachineId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.Contributor.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: strings.ToUpper(data.VirtualMachineId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.VMContributor.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: strings.ToUpper(data.VirtualMachineId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.Owner.GetPrincipalId()),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: strings.ToUpper(data.VirtualMachineId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(raw.UserAccessAdmin.Properties.PrincipalId),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: strings.ToUpper(data.VirtualMachineId),
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
		Labels: []graph.Kind{azure.ManagedCluster},
	}

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.ManagedCluster,
			Value: strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.Id),
			Kind:  azure.ManagedCluster,
		},
		IngestibleEndpoint{
			Kind:  azure.ResourceGroup,
			Value: strings.ToUpper(nodeResourceGroupID),
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
					IngestibleEndpoint{
						Value: strings.ToUpper(raw.Assignee.GetPrincipalId()),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.ManagedCluster,
						Value: strings.ToUpper(data.ObjectId),
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
		Labels: []graph.Kind{azure.ContainerRegistry},
	}

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(data.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.ContainerRegistry,
			Value: strings.ToUpper(data.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if data.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(data.Id),
				Kind:  azure.ContainerRegistry,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(data.Identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(data.Id),
					Kind:  azure.ContainerRegistry,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(identity.PrincipalId),
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
		Labels: []graph.Kind{azure.WebApp},
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(webApp.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.WebApp,
			Value: strings.ToUpper(webApp.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if webApp.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(webApp.Id),
				Kind:  azure.WebApp,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(webApp.Identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(webApp.Id),
					Kind:  azure.WebApp,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(identity.PrincipalId),
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
					IngestibleEndpoint{
						Value: strings.ToUpper(raw.Assignee.GetPrincipalId()),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.AutomationAccount,
						Value: strings.ToUpper(roleAssignments.ObjectId),
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
					IngestibleEndpoint{
						Value: strings.ToUpper(raw.Assignee.GetPrincipalId()),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.ContainerRegistry,
						Value: strings.ToUpper(roleAssignment.ObjectId),
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
					IngestibleEndpoint{
						Value: strings.ToUpper(raw.Assignee.GetPrincipalId()),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.WebApp,
						Value: strings.ToUpper(roleAssignment.ObjectId),
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
		Labels: []graph.Kind{azure.AutomationAccount},
	}

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(account.ResourceGroupId),
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.AutomationAccount,
			Value: strings.ToUpper(account.Id),
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	// Enumerate System Assigned Identities
	if account.Identity.PrincipalId != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: strings.ToUpper(account.Id),
				Kind:  azure.AutomationAccount,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: strings.ToUpper(account.Identity.PrincipalId),
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
				IngestibleEndpoint{
					Value: strings.ToUpper(account.Id),
					Kind:  azure.AutomationAccount,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: strings.ToUpper(identity.PrincipalId),
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

func ConvertAzureRoleEligibilityScheduleInstanceToRel(instance models.RoleEligibilityScheduleInstance) []IngestibleRelationship {
	id := strings.ToUpper(fmt.Sprintf("%s@%s", instance.RoleDefinitionId, instance.TenantId))

	relationships := make([]IngestibleRelationship, 0)
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: strings.ToUpper(instance.PrincipalId),
			Kind:  azure.Entity,
		},
		IngestibleEndpoint{
			Value: id,
			Kind:  azure.Role,
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.AZRoleEligible,
		},
	))

	return relationships
}

// ConvertAzureRoleManagementPolicyAssignment will create, or update the properties if it exists, an AZRole of a tenant
// with the supplied RoleManagementPolicyAssignment properties
// If EndUserAssignmentGroupApprovers contains GUIDs: an edge will be created from each group to the created AZRole
// If EndUserAssignmentUsersApprovers contains GUIDs: an edge will be created from each user to the created AZRole
// If both lists are empty: an edge will be created from the tenant's PrivilegedRoleAdministratorRole to the created AZRole
func ConvertAzureRoleManagementPolicyAssignment(policyAssignment models.RoleManagementPolicyAssignment) (IngestibleNode, []IngestibleRelationship) {
	var (
		rels             = make([]IngestibleRelationship, 0)
		combinedObjectId = strings.ToUpper(fmt.Sprintf("%s@%s", policyAssignment.RoleDefinitionId, policyAssignment.TenantId))
	)

	// Format the incoming user and group ids to uppercase string before creating our nodes
	for i := range policyAssignment.EndUserAssignmentGroupApprovers {
		policyAssignment.EndUserAssignmentGroupApprovers[i] = strings.ToUpper(policyAssignment.EndUserAssignmentGroupApprovers[i])
	}

	for i := range policyAssignment.EndUserAssignmentUserApprovers {
		policyAssignment.EndUserAssignmentUserApprovers[i] = strings.ToUpper(policyAssignment.EndUserAssignmentUserApprovers[i])
	}

	// We will want to create or update any existing AZRole node that matches the combinedObjectId
	// If the node exists, we want to add the new properties to the node
	targetAZRole := IngestibleNode{
		ObjectID: combinedObjectId,
		Labels:   []graph.Kind{azure.Role},
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
		// We cannot create the edges if the assignment does not require approval
		return targetAZRole, rels
	}

	// TODO: Verify the edge creation here. The logic looks identical to the post processing for this edge and we could remove the edge creation here
	if len(policyAssignment.EndUserAssignmentUserApprovers) > 0 {
		// Create an AZRoleApprover edge from each user that allow approvals to the target azure role
		for _, approver := range policyAssignment.EndUserAssignmentUserApprovers {
			rels = append(rels, NewIngestibleRelationship(IngestibleEndpoint{
				Value: strings.ToUpper(approver),
				Kind:  azure.User,
			}, IngestibleEndpoint{
				Value: targetAZRole.ObjectID,
				Kind:  targetAZRole.Labels[0],
			}, IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.AZRoleApprover,
			}))
		}
	}

	if len(policyAssignment.EndUserAssignmentGroupApprovers) > 0 {
		// Create an AZRoleApprover edge from each group that allow approvals to the target azure role
		for _, approver := range policyAssignment.EndUserAssignmentGroupApprovers {
			rels = append(rels, NewIngestibleRelationship(IngestibleEndpoint{
				Value: strings.ToUpper(approver),
				Kind:  azure.Group,
			}, IngestibleEndpoint{
				Value: targetAZRole.ObjectID,
				Kind:  targetAZRole.Labels[0],
			}, IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.AZRoleApprover,
			}))
		}
	}

	if len(policyAssignment.EndUserAssignmentUserApprovers) == 0 && len(policyAssignment.EndUserAssignmentGroupApprovers) == 0 {
		// No users or groups were attached to the policy, we will create the edge from the tenant's PrivilegedRoleAdministratorRole Role node to the target role
		combinedObjectId := strings.ToUpper(fmt.Sprintf("%s@%s", azure.PrivilegedRoleAdministratorRole, policyAssignment.TenantId))

		rels = append(rels, NewIngestibleRelationship(IngestibleEndpoint{
			Value: strings.ToUpper(combinedObjectId),
			Kind:  azure.Role,
		}, IngestibleEndpoint{
			Value: targetAZRole.ObjectID,
			Kind:  targetAZRole.Labels[0],
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

/*

Example of an OAuth2PermissionGrant object in AzureHound format:

{"kind":"AZOAuth2PermissionGrant",
"data":{"clientId":"<ClientID>",
"consentType":"AllPrincipals",
"id":"<ID>",
"resourceId":"<ResourceID>",
"scope":"AppCatalog.Read.All AppCatalog.Submit Channel.ReadBasic.All EduAssignments.ReadBasic EduRoster.ReadBasic Files.Read.All Files.ReadWrite.All Group.Read.All People.Read People.Read.All Presence.Read.All TeamsAppInstallation.ReadWriteSelfForTeam User.Read User.ReadBasic.All Tasks.ReadWrite Group-Conversation.ReadWrite.All Team.ReadBasic.All Channel.Create TeamsAppInstallation.ReadWriteForTeam Sites.Read.All PrinterShare.ReadBasic.All PrintJob.Create PrintJob.ReadBasic FileStorageContainer.Selected Calendars.Read Files.Read GroupMember.Read.All InformationProtectionPolicy.Read ChatMember.Read TeamsTab.Create",
"tenantId":"<TenantID>",
"tenantName":"<Name>"}},

*/

// Since this implementation is not yet part of the current AzureHound model, here's the definition
// of the OAuth2PermissionGrant struct as per the Azure Graph API documentation.
// This can be replaced with the actual model definition once it is available in AzureHound.
type DirectoryObject struct {
	Id   string `json:"id"`
	Type string `json:"@odata.type,omitempty"`
}
type OAuth2PermissionGrant struct {
	TenantId   string `json:"tenantId"`
	TenantName string `json:"tenantName"`
	ClientId   string `json:"clientId,omitempty"`
	DirectoryObject
	ConsentType string `json:"consentType,omitempty"`
	Id          string `json:"id,omitempty"`
	PrincipalId string `json:"principalId,omitempty"`
	ResourceId  string `json:"resourceId,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

func GetPermissionConstant(scope string) graph.Kind {
	switch scope {
	case "APIConnectors.Read.All":
		return azure.APIConnectorsReadAll
	case "APIConnectors.ReadWrite.All":
		return azure.APIConnectorsReadWriteAll
	case "AccessReview.Read.All":
		return azure.AccessReviewReadAll
	case "AccessReview.ReadWrite.All":
		return azure.AccessReviewReadWriteAll
	case "AccessReview.ReadWrite.Membership":
		return azure.AccessReviewReadWriteMembership
	case "Acronym.Read.All":
		return azure.AcronymReadAll
	case "AdministrativeUnit.Read.All":
		return azure.AdministrativeUnitReadAll
	case "AdministrativeUnit.ReadWrite.All":
		return azure.AdministrativeUnitReadWriteAll
	case "AgentApplication.Create":
		return azure.AgentApplicationCreate
	case "AgentIdentity.Create":
		return azure.AgentIdentityCreate
	case "Agreement.Read.All":
		return azure.AgreementReadAll
	case "Agreement.ReadWrite.All":
		return azure.AgreementReadWriteAll
	case "AgreementAcceptance.Read":
		return azure.AgreementAcceptanceRead
	case "AgreementAcceptance.Read.All":
		return azure.AgreementAcceptanceReadAll
	case "AiEnterpriseInteraction.Read":
		return azure.AiEnterpriseInteractionRead
	case "AiEnterpriseInteraction.Read.All":
		return azure.AiEnterpriseInteractionReadAll
	case "AllSites.Read":
		return azure.AllSitesRead
	case "Analytics.Read":
		return azure.AnalyticsRead
	case "AppCatalog.Read.All":
		return azure.AppCatalogReadAll
	case "AppCatalog.ReadWrite.All":
		return azure.AppCatalogReadWriteAll
	case "AppCatalog.Submit":
		return azure.AppCatalogSubmit
	case "AppCertTrustConfiguration.Read.All":
		return azure.AppCertTrustConfigurationReadAll
	case "AppCertTrustConfiguration.ReadWrite.All":
		return azure.AppCertTrustConfigurationReadWriteAll
	case "AppRoleAssignment.ReadWrite.All":
		return azure.AppRoleAssignmentReadWriteAll
	case "Application.Read.All":
		return azure.ApplicationReadAll
	case "Application.ReadWrite.All":
		return azure.ApplicationReadWriteAll
	case "Application.ReadWrite.OwnedBy":
		return azure.ApplicationReadWriteOwnedBy
	case "ApprovalSolution.Read":
		return azure.ApprovalSolutionRead
	case "ApprovalSolution.Read.All":
		return azure.ApprovalSolutionReadAll
	case "ApprovalSolution.ReadWrite":
		return azure.ApprovalSolutionReadWrite
	case "ApprovalSolution.ReadWrite.All":
		return azure.ApprovalSolutionReadWriteAll
	case "ApprovalSolutionResponse.ReadWrite":
		return azure.ApprovalSolutionResponseReadWrite
	case "AttackSimulation.Read.All":
		return azure.AttackSimulationReadAll
	case "AttackSimulation.ReadWrite.All":
		return azure.AttackSimulationReadWriteAll
	case "AuditActivity.Read":
		return azure.AuditActivityRead
	case "AuditActivity.Write":
		return azure.AuditActivityWrite
	case "AuditLog.Read.All":
		return azure.AuditLogReadAll
	case "AuditLogsQuery.Read.All":
		return azure.AuditLogsQueryReadAll
	case "AuthenticationContext.Read.All":
		return azure.AuthenticationContextReadAll
	case "AuthenticationContext.ReadWrite.All":
		return azure.AuthenticationContextReadWriteAll
	case "BillingConfiguration.ReadWrite.All":
		return azure.BillingConfigurationReadWriteAll
	case "BitlockerKey.Read.All":
		return azure.BitlockerKeyReadAll
	case "BitlockerKey.ReadBasic.All":
		return azure.BitlockerKeyReadBasicAll
	case "Bookings.Manage.All":
		return azure.BookingsManageAll
	case "Bookings.Read.All":
		return azure.BookingsReadAll
	case "Bookings.ReadWrite.All":
		return azure.BookingsReadWriteAll
	case "BookingsAppointment.ReadWrite.All":
		return azure.BookingsAppointmentReadWriteAll
	case "Bookmark.Read.All":
		return azure.BookmarkReadAll
	case "BrowserSiteLists.Read.All":
		return azure.BrowserSiteListsReadAll
	case "BrowserSiteLists.ReadWrite.All":
		return azure.BrowserSiteListsReadWriteAll
	case "BusinessScenarioConfig.Read.All":
		return azure.BusinessScenarioConfigReadAll
	case "BusinessScenarioConfig.Read.OwnedBy":
		return azure.BusinessScenarioConfigReadOwnedBy
	case "BusinessScenarioConfig.ReadWrite.All":
		return azure.BusinessScenarioConfigReadWriteAll
	case "BusinessScenarioConfig.ReadWrite.OwnedBy":
		return azure.BusinessScenarioConfigReadWriteOwnedBy
	case "BusinessScenarioData.Read.OwnedBy":
		return azure.BusinessScenarioDataReadOwnedBy
	case "BusinessScenarioData.ReadWrite.OwnedBy":
		return azure.BusinessScenarioDataReadWriteOwnedBy
	case "Calendars.Read":
		return azure.CalendarsRead
	case "Calendars.Read.Shared":
		return azure.CalendarsReadShared
	case "Calendars.ReadBasic":
		return azure.CalendarsReadBasic
	case "Calendars.ReadBasic.All":
		return azure.CalendarsReadBasicAll
	case "Calendars.ReadWrite":
		return azure.CalendarsReadWrite
	case "Calendars.ReadWrite.Shared":
		return azure.CalendarsReadWriteShared
	case "CallAiInsights.Read.All":
		return azure.CallAiInsightsReadAll
	case "CallDelegation.Read":
		return azure.CallDelegationRead
	case "CallDelegation.Read.All":
		return azure.CallDelegationReadAll
	case "CallDelegation.ReadWrite":
		return azure.CallDelegationReadWrite
	case "CallDelegation.ReadWrite.All":
		return azure.CallDelegationReadWriteAll
	case "CallEvents.Read":
		return azure.CallEventsRead
	case "CallEvents.Read.All":
		return azure.CallEventsReadAll
	case "CallRecords.Read.All":
		return azure.CallRecordsReadAll
	case "Calls.AccessMedia.All":
		return azure.CallsAccessMediaAll
	case "Calls.Initiate.All":
		return azure.CallsInitiateAll
	case "Calls.InitiateGroupCall.All":
		return azure.CallsInitiateGroupCallAll
	case "Calls.JoinGroupCall.All":
		return azure.CallsJoinGroupCallAll
	case "Calls.JoinGroupCallAsGuest.All":
		return azure.CallsJoinGroupCallAsGuestAll
	case "ChangeManagement.Read.All":
		return azure.ChangeManagementReadAll
	case "Channel.Create":
		return azure.ChannelCreate
	case "Channel.Delete.All":
		return azure.ChannelDeleteAll
	case "Channel.ReadBasic.All":
		return azure.ChannelReadBasicAll
	case "ChannelMember.Read.All":
		return azure.ChannelMemberReadAll
	case "ChannelMember.ReadWrite.All":
		return azure.ChannelMemberReadWriteAll
	case "ChannelMessage.Edit":
		return azure.ChannelMessageEdit
	case "ChannelMessage.Read.All":
		return azure.ChannelMessageReadAll
	case "ChannelMessage.ReadWrite":
		return azure.ChannelMessageReadWrite
	case "ChannelMessage.Send":
		return azure.ChannelMessageSend
	case "ChannelMessage.UpdatePolicyViolation.All":
		return azure.ChannelMessageUpdatePolicyViolationAll
	case "ChannelSettings.Read.All":
		return azure.ChannelSettingsReadAll
	case "ChannelSettings.ReadWrite.All":
		return azure.ChannelSettingsReadWriteAll
	case "Chat.Create":
		return azure.ChatCreate
	case "Chat.ManageDeletion.All":
		return azure.ChatManageDeletionAll
	case "Chat.Read":
		return azure.ChatRead
	case "Chat.Read.All":
		return azure.ChatReadAll
	case "Chat.Read.WhereInstalled":
		return azure.ChatReadWhereInstalled
	case "Chat.ReadBasic":
		return azure.ChatReadBasic
	case "Chat.ReadBasic.All":
		return azure.ChatReadBasicAll
	case "Chat.ReadBasic.WhereInstalled":
		return azure.ChatReadBasicWhereInstalled
	case "Chat.ReadWrite":
		return azure.ChatReadWrite
	case "Chat.ReadWrite.All":
		return azure.ChatReadWriteAll
	case "Chat.ReadWrite.WhereInstalled":
		return azure.ChatReadWriteWhereInstalled
	case "Chat.UpdatePolicyViolation.All":
		return azure.ChatUpdatePolicyViolationAll
	case "ChatMember.Read":
		return azure.ChatMemberRead
	case "ChatMember.Read.All":
		return azure.ChatMemberReadAll
	case "ChatMember.Read.WhereInstalled":
		return azure.ChatMemberReadWhereInstalled
	case "ChatMember.ReadWrite":
		return azure.ChatMemberReadWrite
	case "ChatMember.ReadWrite.All":
		return azure.ChatMemberReadWriteAll
	case "ChatMember.ReadWrite.WhereInstalled":
		return azure.ChatMemberReadWriteWhereInstalled
	case "ChatMessage.Read":
		return azure.ChatMessageRead
	case "ChatMessage.Read.All":
		return azure.ChatMessageReadAll
	case "ChatMessage.Send":
		return azure.ChatMessageSend
	case "CloudPC.Read.All":
		return azure.CloudPCReadAll
	case "CloudPC.ReadWrite.All":
		return azure.CloudPCReadWriteAll
	case "Community.Read.All":
		return azure.CommunityReadAll
	case "Community.ReadWrite.All":
		return azure.CommunityReadWriteAll
	case "ConfigurationMonitoring.Read.All":
		return azure.ConfigurationMonitoringReadAll
	case "ConfigurationMonitoring.ReadWrite.All":
		return azure.ConfigurationMonitoringReadWriteAll
	case "ConsentRequest.Create":
		return azure.ConsentRequestCreate
	case "ConsentRequest.Read":
		return azure.ConsentRequestRead
	case "ConsentRequest.Read.All":
		return azure.ConsentRequestReadAll
	case "ConsentRequest.ReadApprove.All":
		return azure.ConsentRequestReadApproveAll
	case "ConsentRequest.ReadWrite.All":
		return azure.ConsentRequestReadWriteAll
	case "Contacts.Read":
		return azure.ContactsRead
	case "Contacts.Read.Shared":
		return azure.ContactsReadShared
	case "Contacts.ReadWrite":
		return azure.ContactsReadWrite
	case "Contacts.ReadWrite.Shared":
		return azure.ContactsReadWriteShared
	case "Content.Process.All":
		return azure.ContentProcessAll
	case "Content.Process.User":
		return azure.ContentProcessUser
	case "ContentActivity.Read":
		return azure.ContentActivityRead
	case "ContentActivity.Write":
		return azure.ContentActivityWrite
	case "CrossTenantInformation.ReadBasic.All":
		return azure.CrossTenantInformationReadBasicAll
	case "CrossTenantUserProfileSharing.Read":
		return azure.CrossTenantUserProfileSharingRead
	case "CrossTenantUserProfileSharing.Read.All":
		return azure.CrossTenantUserProfileSharingReadAll
	case "CrossTenantUserProfileSharing.ReadWrite":
		return azure.CrossTenantUserProfileSharingReadWrite
	case "CrossTenantUserProfileSharing.ReadWrite.All":
		return azure.CrossTenantUserProfileSharingReadWriteAll
	case "CustomAuthenticationExtension.Read.All":
		return azure.CustomAuthenticationExtensionReadAll
	case "CustomAuthenticationExtension.ReadWrite.All":
		return azure.CustomAuthenticationExtensionReadWriteAll
	case "CustomAuthenticationExtension.Receive.Payload":
		return azure.CustomAuthenticationExtensionReceivePayload
	case "CustomDetection.Read.All":
		return azure.CustomDetectionReadAll
	case "CustomDetection.ReadWrite.All":
		return azure.CustomDetectionReadWriteAll
	case "CustomSecAttributeAssignment.Read.All":
		return azure.CustomSecAttributeAssignmentReadAll
	case "CustomSecAttributeAssignment.ReadWrite.All":
		return azure.CustomSecAttributeAssignmentReadWriteAll
	case "CustomSecAttributeAuditLogs.Read.All":
		return azure.CustomSecAttributeAuditLogsReadAll
	case "CustomSecAttributeDefinition.Read.All":
		return azure.CustomSecAttributeDefinitionReadAll
	case "CustomSecAttributeDefinition.ReadWrite.All":
		return azure.CustomSecAttributeDefinitionReadWriteAll
	case "CustomSecAttributeProvisioning.Read.All":
		return azure.CustomSecAttributeProvisioningReadAll
	case "CustomSecAttributeProvisioning.ReadWrite.All":
		return azure.CustomSecAttributeProvisioningReadWriteAll
	case "CustomTags.Read.All":
		return azure.CustomTagsReadAll
	case "CustomTags.ReadWrite.All":
		return azure.CustomTagsReadWriteAll
	case "Dataset.Read.All":
		return azure.DatasetReadAll
	case "DelegatedAdminRelationship.Read.All":
		return azure.DelegatedAdminRelationshipReadAll
	case "DelegatedAdminRelationship.ReadWrite.All":
		return azure.DelegatedAdminRelationshipReadWriteAll
	case "DelegatedPermissionGrant.Read.All":
		return azure.DelegatedPermissionGrantReadAll
	case "DelegatedPermissionGrant.ReadWrite.All":
		return azure.DelegatedPermissionGrantReadWriteAll
	case "Device.Command":
		return azure.DeviceCommand
	case "Device.CreateFromOwnedTemplate":
		return azure.DeviceCreateFromOwnedTemplate
	case "Device.Read":
		return azure.DeviceRead
	case "Device.Read.All":
		return azure.DeviceReadAll
	case "Device.ReadWrite.All":
		return azure.DeviceReadWriteAll
	case "DeviceLocalCredential.Read.All":
		return azure.DeviceLocalCredentialReadAll
	case "DeviceLocalCredential.ReadBasic.All":
		return azure.DeviceLocalCredentialReadBasicAll
	case "DeviceManagementApps.Read.All":
		return azure.DeviceManagementAppsReadAll
	case "DeviceManagementApps.ReadWrite.All":
		return azure.DeviceManagementAppsReadWriteAll
	case "DeviceManagementCloudCA.Read.All":
		return azure.DeviceManagementCloudCAReadAll
	case "DeviceManagementCloudCA.ReadWrite.All":
		return azure.DeviceManagementCloudCAReadWriteAll
	case "DeviceManagementConfiguration.Read.All":
		return azure.DeviceManagementConfigurationReadAll
	case "DeviceManagementConfiguration.ReadWrite.All":
		return azure.DeviceManagementConfigurationReadWriteAll
	case "DeviceManagementManagedDevices.PrivilegedOperations.All":
		return azure.DeviceManagementManagedDevicesPrivilegedOperationsAll
	case "DeviceManagementManagedDevices.Read.All":
		return azure.DeviceManagementManagedDevicesReadAll
	case "DeviceManagementManagedDevices.ReadWrite.All":
		return azure.DeviceManagementManagedDevicesReadWriteAll
	case "DeviceManagementRBAC.Read.All":
		return azure.DeviceManagementRBACReadAll
	case "DeviceManagementRBAC.ReadWrite.All":
		return azure.DeviceManagementRBACReadWriteAll
	case "DeviceManagementScripts.Read.All":
		return azure.DeviceManagementScriptsReadAll
	case "DeviceManagementScripts.ReadWrite.All":
		return azure.DeviceManagementScriptsReadWriteAll
	case "DeviceManagementServiceConfig.Read.All":
		return azure.DeviceManagementServiceConfigReadAll
	case "DeviceManagementServiceConfig.ReadWrite.All":
		return azure.DeviceManagementServiceConfigReadWriteAll
	case "DeviceTemplate.Create":
		return azure.DeviceTemplateCreate
	case "DeviceTemplate.Read.All":
		return azure.DeviceTemplateReadAll
	case "DeviceTemplate.ReadWrite.All":
		return azure.DeviceTemplateReadWriteAll
	case "Directory.AccessAsUser.All":
		return azure.DirectoryAccessAsUserAll
	case "Directory.Read.All":
		return azure.DirectoryReadAll
	case "Directory.ReadWrite.All":
		return azure.DirectoryReadWriteAll
	case "DirectoryRecommendations.Read.All":
		return azure.DirectoryRecommendationsReadAll
	case "DirectoryRecommendations.ReadWrite.All":
		return azure.DirectoryRecommendationsReadWriteAll
	case "Domain.Read.All":
		return azure.DomainReadAll
	case "Domain.ReadWrite.All":
		return azure.DomainReadWriteAll
	case "EAS.AccessAsUser.All":
		return azure.EASAccessAsUserAll
	case "EWS.AccessAsUser.All":
		return azure.EWSAccessAsUserAll
	case "EduAdministration.Read":
		return azure.EduAdministrationRead
	case "EduAdministration.Read.All":
		return azure.EduAdministrationReadAll
	case "EduAdministration.ReadWrite":
		return azure.EduAdministrationReadWrite
	case "EduAdministration.ReadWrite.All":
		return azure.EduAdministrationReadWriteAll
	case "EduAssignments.Read":
		return azure.EduAssignmentsRead
	case "EduAssignments.Read.All":
		return azure.EduAssignmentsReadAll
	case "EduAssignments.ReadBasic":
		return azure.EduAssignmentsReadBasic
	case "EduAssignments.ReadBasic.All":
		return azure.EduAssignmentsReadBasicAll
	case "EduAssignments.ReadWrite":
		return azure.EduAssignmentsReadWrite
	case "EduAssignments.ReadWrite.All":
		return azure.EduAssignmentsReadWriteAll
	case "EduAssignments.ReadWriteBasic":
		return azure.EduAssignmentsReadWriteBasic
	case "EduAssignments.ReadWriteBasic.All":
		return azure.EduAssignmentsReadWriteBasicAll
	case "EduCurricula.Read":
		return azure.EduCurriculaRead
	case "EduCurricula.Read.All":
		return azure.EduCurriculaReadAll
	case "EduCurricula.ReadWrite":
		return azure.EduCurriculaReadWrite
	case "EduCurricula.ReadWrite.All":
		return azure.EduCurriculaReadWriteAll
	case "EduRoster.Read":
		return azure.EduRosterRead
	case "EduRoster.Read.All":
		return azure.EduRosterReadAll
	case "EduRoster.ReadBasic":
		return azure.EduRosterReadBasic
	case "EduRoster.ReadBasic.All":
		return azure.EduRosterReadBasicAll
	case "EduRoster.ReadWrite":
		return azure.EduRosterReadWrite
	case "EduRoster.ReadWrite.All":
		return azure.EduRosterReadWriteAll
	case "EngagementConversation.Migration.All":
		return azure.EngagementConversationMigrationAll
	case "EngagementConversation.ReadWrite.All":
		return azure.EngagementConversationReadWriteAll
	case "EngagementMeetingConversation.Read.All":
		return azure.EngagementMeetingConversationReadAll
	case "EngagementRole.Read":
		return azure.EngagementRoleRead
	case "EngagementRole.Read.All":
		return azure.EngagementRoleReadAll
	case "EngagementRole.ReadWrite.All":
		return azure.EngagementRoleReadWriteAll
	case "EntitlementManagement.Read.All":
		return azure.EntitlementManagementReadAll
	case "EntitlementManagement.ReadWrite.All":
		return azure.EntitlementManagementReadWriteAll
	case "EventListener.Read.All":
		return azure.EventListenerReadAll
	case "EventListener.ReadWrite.All":
		return azure.EventListenerReadWriteAll
	case "ExternalConnection.Read.All":
		return azure.ExternalConnectionReadAll
	case "ExternalConnection.ReadWrite.All":
		return azure.ExternalConnectionReadWriteAll
	case "ExternalConnection.ReadWrite.OwnedBy":
		return azure.ExternalConnectionReadWriteOwnedBy
	case "ExternalItem.Read.All":
		return azure.ExternalItemReadAll
	case "ExternalItem.ReadWrite.All":
		return azure.ExternalItemReadWriteAll
	case "ExternalItem.ReadWrite.OwnedBy":
		return azure.ExternalItemReadWriteOwnedBy
	case "ExternalUserProfile.Read.All":
		return azure.ExternalUserProfileReadAll
	case "ExternalUserProfile.ReadWrite.All":
		return azure.ExternalUserProfileReadWriteAll
	case "Family.Read":
		return azure.FamilyRead
	case "FileIngestion.Ingest":
		return azure.FileIngestionIngest
	case "FileIngestionHybridOnboarding.Manage":
		return azure.FileIngestionHybridOnboardingManage
	case "FileStorageContainer.Manage.All":
		return azure.FileStorageContainerManageAll
	case "FileStorageContainer.Selected":
		return azure.FileStorageContainerSelected
	case "FileStorageContainerTypeReg.Selected":
		return azure.FileStorageContainerTypeRegSelected
	case "Files.Read":
		return azure.FilesRead
	case "Files.Read.All":
		return azure.FilesReadAll
	case "Files.Read.Selected":
		return azure.FilesReadSelected
	case "Files.ReadWrite":
		return azure.FilesReadWrite
	case "Files.ReadWrite.All":
		return azure.FilesReadWriteAll
	case "Files.ReadWrite.AppFolder":
		return azure.FilesReadWriteAppFolder
	case "Files.ReadWrite.Selected":
		return azure.FilesReadWriteSelected
	case "Files.SelectedOperations.Selected":
		return azure.FilesSelectedOperationsSelected
	case "Financials.ReadWrite.All":
		return azure.FinancialsReadWriteAll
	case "Forms.ReadWrite":
		return azure.FormsReadWrite
	case "Group.Create":
		return azure.GroupCreate
	case "Group.Read.All":
		return azure.GroupReadAll
	case "Group.ReadWrite.All":
		return azure.GroupReadWriteAll
	case "Group-Conversation.ReadWrite.All":
		return azure.GroupConversationReadWriteAll
	case "GroupMember.Read.All":
		return azure.GroupMemberReadAll
	case "GroupMember.ReadWrite.All":
		return azure.GroupMemberReadWriteAll
	case "GroupSettings.Read.All":
		return azure.GroupSettingsReadAll
	case "GroupSettings.ReadWrite.All":
		return azure.GroupSettingsReadWriteAll
	case "HealthMonitoringAlert.Read.All":
		return azure.HealthMonitoringAlertReadAll
	case "HealthMonitoringAlert.ReadWrite.All":
		return azure.HealthMonitoringAlertReadWriteAll
	case "HealthMonitoringAlertConfig.Read.All":
		return azure.HealthMonitoringAlertConfigReadAll
	case "HealthMonitoringAlertConfig.ReadWrite.All":
		return azure.HealthMonitoringAlertConfigReadWriteAll
	case "IMAP.AccessAsUser.All":
		return azure.IMAPAccessAsUserAll
	case "IdentityProvider.Read.All":
		return azure.IdentityProviderReadAll
	case "IdentityProvider.ReadWrite.All":
		return azure.IdentityProviderReadWriteAll
	case "IdentityRiskEvent.Read.All":
		return azure.IdentityRiskEventReadAll
	case "IdentityRiskEvent.ReadWrite.All":
		return azure.IdentityRiskEventReadWriteAll
	case "IdentityRiskyServicePrincipal.Read.All":
		return azure.IdentityRiskyServicePrincipalReadAll
	case "IdentityRiskyServicePrincipal.ReadWrite.All":
		return azure.IdentityRiskyServicePrincipalReadWriteAll
	case "IdentityRiskyUser.Read.All":
		return azure.IdentityRiskyUserReadAll
	case "IdentityRiskyUser.ReadWrite.All":
		return azure.IdentityRiskyUserReadWriteAll
	case "IdentityUserFlow.Read.All":
		return azure.IdentityUserFlowReadAll
	case "IdentityUserFlow.ReadWrite.All":
		return azure.IdentityUserFlowReadWriteAll
	case "IndustryData.ReadBasic.All":
		return azure.IndustryDataReadBasicAll
	case "InformationProtectionConfig.Read":
		return azure.InformationProtectionConfigRead
	case "InformationProtectionConfig.Read.All":
		return azure.InformationProtectionConfigReadAll
	case "InformationProtectionContent.Sign.All":
		return azure.InformationProtectionContentSignAll
	case "InformationProtectionContent.Write.All":
		return azure.InformationProtectionContentWriteAll
	case "InformationProtectionPolicy.Read":
		return azure.InformationProtectionPolicyRead
	case "InformationProtectionPolicy.Read.All":
		return azure.InformationProtectionPolicyReadAll
	case "LearningAssignedCourse.Read":
		return azure.LearningAssignedCourseRead
	case "LearningAssignedCourse.Read.All":
		return azure.LearningAssignedCourseReadAll
	case "LearningAssignedCourse.ReadWrite.All":
		return azure.LearningAssignedCourseReadWriteAll
	case "LearningContent.Read.All":
		return azure.LearningContentReadAll
	case "LearningContent.ReadWrite.All":
		return azure.LearningContentReadWriteAll
	case "LearningProvider.Read":
		return azure.LearningProviderRead
	case "LearningProvider.ReadWrite":
		return azure.LearningProviderReadWrite
	case "LearningSelfInitiatedCourse.Read":
		return azure.LearningSelfInitiatedCourseRead
	case "LearningSelfInitiatedCourse.Read.All":
		return azure.LearningSelfInitiatedCourseReadAll
	case "LearningSelfInitiatedCourse.ReadWrite.All":
		return azure.LearningSelfInitiatedCourseReadWriteAll
	case "LicenseAssignment.Read.All":
		return azure.LicenseAssignmentReadAll
	case "LicenseAssignment.ReadWrite.All":
		return azure.LicenseAssignmentReadWriteAll
	case "LifecycleWorkflows.Read.All":
		return azure.LifecycleWorkflowsReadAll
	case "LifecycleWorkflows.ReadWrite.All":
		return azure.LifecycleWorkflowsReadWriteAll
	case "ListItems.SelectedOperations.Selected":
		return azure.ListItemsSelectedOperationsSelected
	case "Lists.SelectedOperations.Selected":
		return azure.ListsSelectedOperationsSelected
	case "Mail.Read":
		return azure.MailRead
	case "Mail.Read.Shared":
		return azure.MailReadShared
	case "Mail.ReadBasic":
		return azure.MailReadBasic
	case "Mail.ReadBasic.All":
		return azure.MailReadBasicAll
	case "Mail.ReadBasic.Shared":
		return azure.MailReadBasicShared
	case "Mail.ReadWrite":
		return azure.MailReadWrite
	case "Mail.ReadWrite.Shared":
		return azure.MailReadWriteShared
	case "Mail.Send":
		return azure.MailSend
	case "Mail.Send.Shared":
		return azure.MailSendShared
	case "MailboxFolder.Read":
		return azure.MailboxFolderRead
	case "MailboxFolder.Read.All":
		return azure.MailboxFolderReadAll
	case "MailboxFolder.ReadWrite":
		return azure.MailboxFolderReadWrite
	case "MailboxFolder.ReadWrite.All":
		return azure.MailboxFolderReadWriteAll
	case "MailboxItem.ImportExport":
		return azure.MailboxItemImportExport
	case "MailboxItem.ImportExport.All":
		return azure.MailboxItemImportExportAll
	case "MailboxItem.Read":
		return azure.MailboxItemRead
	case "MailboxItem.Read.All":
		return azure.MailboxItemReadAll
	case "MailboxSettings.Read":
		return azure.MailboxSettingsRead
	case "MailboxSettings.ReadWrite":
		return azure.MailboxSettingsReadWrite
	case "ManagedTenants.Read.All":
		return azure.ManagedTenantsReadAll
	case "ManagedTenants.ReadWrite.All":
		return azure.ManagedTenantsReadWriteAll
	case "Member.Read.Hidden":
		return azure.MemberReadHidden
	case "MLModel.Execute.All":
		return azure.MLModelExecuteAll
	case "MultiTenantOrganization.Read.All":
		return azure.MultiTenantOrganizationReadAll
	case "MultiTenantOrganization.ReadBasic.All":
		return azure.MultiTenantOrganizationReadBasicAll
	case "MultiTenantOrganization.ReadWrite.All":
		return azure.MultiTenantOrganizationReadWriteAll
	case "MutualTlsOauthConfiguration.Read.All":
		return azure.MutualTlsOauthConfigurationReadAll
	case "MutualTlsOauthConfiguration.ReadWrite.All":
		return azure.MutualTlsOauthConfigurationReadWriteAll
	case "MyFiles.Read":
		return azure.MyFilesRead
	case "NetworkAccess.Read.All":
		return azure.NetworkAccessReadAll
	case "NetworkAccess.ReadWrite.All":
		return azure.NetworkAccessReadWriteAll
	case "NetworkAccessBranch.Read.All":
		return azure.NetworkAccessBranchReadAll
	case "NetworkAccessBranch.ReadWrite.All":
		return azure.NetworkAccessBranchReadWriteAll
	case "NetworkAccessPolicy.Read.All":
		return azure.NetworkAccessPolicyReadAll
	case "NetworkAccessPolicy.ReadWrite.All":
		return azure.NetworkAccessPolicyReadWriteAll
	case "Notes.Create":
		return azure.NotesCreate
	case "Notes.Read":
		return azure.NotesRead
	case "Notes.Read.All":
		return azure.NotesReadAll
	case "Notes.ReadWrite":
		return azure.NotesReadWrite
	case "Notes.ReadWrite.All":
		return azure.NotesReadWriteAll
	case "Notes.ReadWrite.CreatedByApp":
		return azure.NotesReadWriteCreatedByApp
	case "Notifications.ReadWrite.CreatedByApp":
		return azure.NotificationsReadWriteCreatedByApp
	case "OnPremDirectorySynchronization.Read.All":
		return azure.OnPremDirectorySynchronizationReadAll
	case "OnPremDirectorySynchronization.ReadWrite.All":
		return azure.OnPremDirectorySynchronizationReadWriteAll
	case "OnPremisesPublishingProfiles.ReadWrite.All":
		return azure.OnPremisesPublishingProfilesReadWriteAll
	case "OnlineMeetingAiInsight.Read.All":
		return azure.OnlineMeetingAiInsightReadAll
	case "OnlineMeetingAiInsight.Read.Chat":
		return azure.OnlineMeetingAiInsightReadChat
	case "OnlineMeetingArtifact.Read.All":
		return azure.OnlineMeetingArtifactReadAll
	case "OnlineMeetingRecording.Read.All":
		return azure.OnlineMeetingRecordingReadAll
	case "OnlineMeetingTranscript.Read.All":
		return azure.OnlineMeetingTranscriptReadAll
	case "OnlineMeetings.Read":
		return azure.OnlineMeetingsRead
	case "OnlineMeetings.Read.All":
		return azure.OnlineMeetingsReadAll
	case "OnlineMeetings.ReadWrite":
		return azure.OnlineMeetingsReadWrite
	case "OnlineMeetings.ReadWrite.All":
		return azure.OnlineMeetingsReadWriteAll
	case "OrgContact.Read.All":
		return azure.OrgContactReadAll
	case "Organization.Read.All":
		return azure.OrganizationReadAll
	case "Organization.ReadWrite.All":
		return azure.OrganizationReadWriteAll
	case "OrganizationalBranding.Read.All":
		return azure.OrganizationalBrandingReadAll
	case "OrganizationalBranding.ReadWrite.All":
		return azure.OrganizationalBrandingReadWriteAll
	case "POP.AccessAsUser.All":
		return azure.POPAccessAsUserAll
	case "PartnerBilling.Read.All":
		return azure.PartnerBillingReadAll
	case "PartnerSecurity.Read.All":
		return azure.PartnerSecurityReadAll
	case "PartnerSecurity.ReadWrite.All":
		return azure.PartnerSecurityReadWriteAll
	case "PendingExternalUserProfile.Read.All":
		return azure.PendingExternalUserProfileReadAll
	case "PendingExternalUserProfile.ReadWrite.All":
		return azure.PendingExternalUserProfileReadWriteAll
	case "People.Read":
		return azure.PeopleRead
	case "People.Read.All":
		return azure.PeopleReadAll
	case "PeopleSettings.Read.All":
		return azure.PeopleSettingsReadAll
	case "PeopleSettings.ReadWrite.All":
		return azure.PeopleSettingsReadWriteAll
	case "Place.Read.All":
		return azure.PlaceReadAll
	case "Place.ReadWrite.All":
		return azure.PlaceReadWriteAll
	case "PlaceDevice.Read.All":
		return azure.PlaceDeviceReadAll
	case "PlaceDevice.ReadWrite.All":
		return azure.PlaceDeviceReadWriteAll
	case "PlaceDeviceTelemetry.ReadWrite.All":
		return azure.PlaceDeviceTelemetryReadWriteAll
	case "Policy.Read.All":
		return azure.PolicyReadAll
	case "Policy.Read.AuthenticationMethod":
		return azure.PolicyReadAuthenticationMethod
	case "Policy.Read.ConditionalAccess":
		return azure.PolicyReadConditionalAccess
	case "Policy.Read.DeviceConfiguration":
		return azure.PolicyReadDeviceConfiguration
	case "Policy.Read.IdentityProtection":
		return azure.PolicyReadIdentityProtection
	case "Policy.Read.PermissionGrant":
		return azure.PolicyReadPermissionGrant
	case "Policy.ReadWrite.AccessReview":
		return azure.PolicyReadWriteAccessReview
	case "Policy.ReadWrite.ApplicationConfiguration":
		return azure.PolicyReadWriteApplicationConfiguration
	case "Policy.ReadWrite.AuthenticationFlows":
		return azure.PolicyReadWriteAuthenticationFlows
	case "Policy.ReadWrite.AuthenticationMethod":
		return azure.PolicyReadWriteAuthenticationMethod
	case "Policy.ReadWrite.Authorization":
		return azure.PolicyReadWriteAuthorization
	case "Policy.ReadWrite.ConditionalAccess":
		return azure.PolicyReadWriteConditionalAccess
	case "Policy.ReadWrite.ConsentRequest":
		return azure.PolicyReadWriteConsentRequest
	case "Policy.ReadWrite.CrossTenantAccess":
		return azure.PolicyReadWriteCrossTenantAccess
	case "Policy.ReadWrite.CrossTenantCapability":
		return azure.PolicyReadWriteCrossTenantCapability
	case "Policy.ReadWrite.DeviceConfiguration":
		return azure.PolicyReadWriteDeviceConfiguration
	case "Policy.ReadWrite.ExternalIdentities":
		return azure.PolicyReadWriteExternalIdentities
	case "Policy.ReadWrite.FeatureRollout":
		return azure.PolicyReadWriteFeatureRollout
	case "Policy.ReadWrite.FedTokenValidation":
		return azure.PolicyReadWriteFedTokenValidation
	case "Policy.ReadWrite.IdentityProtection":
		return azure.PolicyReadWriteIdentityProtection
	case "Policy.ReadWrite.MobilityManagement":
		return azure.PolicyReadWriteMobilityManagement
	case "Policy.ReadWrite.PermissionGrant":
		return azure.PolicyReadWritePermissionGrant
	case "Policy.ReadWrite.SecurityDefaults":
		return azure.PolicyReadWriteSecurityDefaults
	case "Policy.ReadWrite.TrustFramework":
		return azure.PolicyReadWriteTrustFramework
	case "Presence.Read":
		return azure.PresenceRead
	case "Presence.Read.All":
		return azure.PresenceReadAll
	case "Presence.ReadWrite":
		return azure.PresenceReadWrite
	case "Presence.ReadWrite.All":
		return azure.PresenceReadWriteAll
	case "PrintConnector.Read.All":
		return azure.PrintConnectorReadAll
	case "PrintConnector.ReadWrite.All":
		return azure.PrintConnectorReadWriteAll
	case "PrintJob.Create":
		return azure.PrintJobCreate
	case "PrintJob.Manage.All":
		return azure.PrintJobManageAll
	case "PrintJob.Read":
		return azure.PrintJobRead
	case "PrintJob.Read.All":
		return azure.PrintJobReadAll
	case "PrintJob.ReadBasic":
		return azure.PrintJobReadBasic
	case "PrintJob.ReadBasic.All":
		return azure.PrintJobReadBasicAll
	case "PrintJob.ReadWrite":
		return azure.PrintJobReadWrite
	case "PrintJob.ReadWrite.All":
		return azure.PrintJobReadWriteAll
	case "PrintJob.ReadWriteBasic":
		return azure.PrintJobReadWriteBasic
	case "PrintJob.ReadWriteBasic.All":
		return azure.PrintJobReadWriteBasicAll
	case "PrintSettings.Read.All":
		return azure.PrintSettingsReadAll
	case "PrintSettings.ReadWrite.All":
		return azure.PrintSettingsReadWriteAll
	case "PrintTaskDefinition.ReadWrite.All":
		return azure.PrintTaskDefinitionReadWriteAll
	case "Printer.Create":
		return azure.PrinterCreate
	case "Printer.FullControl.All":
		return azure.PrinterFullControlAll
	case "Printer.Read.All":
		return azure.PrinterReadAll
	case "Printer.ReadWrite.All":
		return azure.PrinterReadWriteAll
	case "PrinterShare.Read.All":
		return azure.PrinterShareReadAll
	case "PrinterShare.ReadBasic.All":
		return azure.PrinterShareReadBasicAll
	case "PrinterShare.ReadWrite.All":
		return azure.PrinterShareReadWriteAll
	case "PrivilegedAccess.Read.AzureAD":
		return azure.PrivilegedAccessReadAzureAD
	case "PrivilegedAccess.Read.AzureADGroup":
		return azure.PrivilegedAccessReadAzureADGroup
	case "PrivilegedAccess.Read.AzureResources":
		return azure.PrivilegedAccessReadAzureResources
	case "PrivilegedAccess.ReadWrite.AzureAD":
		return azure.PrivilegedAccessReadWriteAzureAD
	case "PrivilegedAccess.ReadWrite.AzureADGroup":
		return azure.PrivilegedAccessReadWriteAzureADGroup
	case "PrivilegedAccess.ReadWrite.AzureResources":
		return azure.PrivilegedAccessReadWriteAzureResources
	case "PrivilegedAssignmentSchedule.Read.AzureADGroup":
		return azure.PrivilegedAssignmentScheduleReadAzureADGroup
	case "PrivilegedAssignmentSchedule.ReadWrite.AzureADGroup":
		return azure.PrivilegedAssignmentScheduleReadWriteAzureADGroup
	case "PrivilegedAssignmentSchedule.Remove.AzureADGroup":
		return azure.PrivilegedAssignmentScheduleRemoveAzureADGroup
	case "PrivilegedEligibilitySchedule.Read.AzureADGroup":
		return azure.PrivilegedEligibilityScheduleReadAzureADGroup
	case "PrivilegedEligibilitySchedule.ReadWrite.AzureADGroup":
		return azure.PrivilegedEligibilityScheduleReadWriteAzureADGroup
	case "PrivilegedEligibilitySchedule.Remove.AzureADGroup":
		return azure.PrivilegedEligibilityScheduleRemoveAzureADGroup
	case "ProfilePhoto.Read.All":
		return azure.ProfilePhotoReadAll
	case "ProfilePhoto.ReadWrite.All":
		return azure.ProfilePhotoReadWriteAll
	case "ProgramControl.Read.All":
		return azure.ProgramControlReadAll
	case "ProgramControl.ReadWrite.All":
		return azure.ProgramControlReadWriteAll
	case "ProtectionScopes.Compute.All":
		return azure.ProtectionScopesComputeAll
	case "ProtectionScopes.Compute.User":
		return azure.ProtectionScopesComputeUser
	case "ProvisioningLog.Read.All":
		return azure.ProvisioningLogReadAll
	case "PublicKeyInfrastructure.Read.All":
		return azure.PublicKeyInfrastructureReadAll
	case "PublicKeyInfrastructure.ReadWrite.All":
		return azure.PublicKeyInfrastructureReadWriteAll
	case "QnA.Read.All":
		return azure.QnAReadAll
	case "RecordsManagement.Read.All":
		return azure.RecordsManagementReadAll
	case "RecordsManagement.ReadWrite.All":
		return azure.RecordsManagementReadWriteAll
	case "ReportSettings.Read.All":
		return azure.ReportSettingsReadAll
	case "ReportSettings.ReadWrite.All":
		return azure.ReportSettingsReadWriteAll
	case "Reports.Read.All":
		return azure.ReportsReadAll
	case "Report.Read.All":
		return azure.ReportReadAll
	case "ResourceSpecificPermissionGrant.ReadForChat":
		return azure.ResourceSpecificPermissionGrantReadForChat
	case "ResourceSpecificPermissionGrant.ReadForChat.All":
		return azure.ResourceSpecificPermissionGrantReadForChatAll
	case "ResourceSpecificPermissionGrant.ReadForTeam":
		return azure.ResourceSpecificPermissionGrantReadForTeam
	case "ResourceSpecificPermissionGrant.ReadForTeam.All":
		return azure.ResourceSpecificPermissionGrantReadForTeamAll
	case "ResourceSpecificPermissionGrant.ReadForUser":
		return azure.ResourceSpecificPermissionGrantReadForUser
	case "ResourceSpecificPermissionGrant.ReadForUser.All":
		return azure.ResourceSpecificPermissionGrantReadForUserAll
	case "RiskPreventionProviders.Read.All":
		return azure.RiskPreventionProvidersReadAll
	case "RiskPreventionProviders.ReadWrite.All":
		return azure.RiskPreventionProvidersReadWriteAll
	case "RoleAssignmentSchedule.Read.Directory":
		return azure.RoleAssignmentScheduleReadDirectory
	case "RoleAssignmentSchedule.ReadWrite.Directory":
		return azure.RoleAssignmentScheduleReadWriteDirectory
	case "RoleAssignmentSchedule.Remove.Directory":
		return azure.RoleAssignmentScheduleRemoveDirectory
	case "RoleEligibilitySchedule.Read.Directory":
		return azure.RoleEligibilityScheduleReadDirectory
	case "RoleEligibilitySchedule.ReadWrite.Directory":
		return azure.RoleEligibilityScheduleReadWriteDirectory
	case "RoleEligibilitySchedule.Remove.Directory":
		return azure.RoleEligibilityScheduleRemoveDirectory
	case "RoleManagement.Read.All":
		return azure.RoleManagementReadAll
	case "RoleManagement.Read.CloudPC":
		return azure.RoleManagementReadCloudPC
	case "RoleManagement.Read.Defender":
		return azure.RoleManagementReadDefender
	case "RoleManagement.Read.Directory":
		return azure.RoleManagementReadDirectory
	case "RoleManagement.Read.Exchange":
		return azure.RoleManagementReadExchange
	case "RoleManagement.ReadWrite.CloudPC":
		return azure.RoleManagementReadWriteCloudPC
	case "RoleManagement.ReadWrite.Defender":
		return azure.RoleManagementReadWriteDefender
	case "RoleManagement.ReadWrite.Directory":
		return azure.RoleManagementReadWriteDirectory
	case "RoleManagement.ReadWrite.Exchange":
		return azure.RoleManagementReadWriteExchange
	case "RoleManagementAlert.Read.Directory":
		return azure.RoleManagementAlertReadDirectory
	case "RoleManagementAlert.ReadWrite.Directory":
		return azure.RoleManagementAlertReadWriteDirectory
	case "RoleManagementPolicy.Read.AzureADGroup":
		return azure.RoleManagementPolicyReadAzureADGroup
	case "RoleManagementPolicy.Read.Directory":
		return azure.RoleManagementPolicyReadDirectory
	case "RoleManagementPolicy.ReadWrite.AzureADGroup":
		return azure.RoleManagementPolicyReadWriteAzureADGroup
	case "RoleManagementPolicy.ReadWrite.Directory":
		return azure.RoleManagementPolicyReadWriteDirectory
	case "SMTP.Send":
		return azure.SMTPSend
	case "Schedule.Read.All":
		return azure.ScheduleReadAll
	case "Schedule.ReadWrite.All":
		return azure.ScheduleReadWriteAll
	case "SchedulePermissions.ReadWrite.All":
		return azure.SchedulePermissionsReadWriteAll
	case "SearchConfiguration.Read.All":
		return azure.SearchConfigurationReadAll
	case "SearchConfiguration.ReadWrite.All":
		return azure.SearchConfigurationReadWriteAll
	case "SecurityActions.Read.All":
		return azure.SecurityActionsReadAll
	case "SecurityActions.ReadWrite.All":
		return azure.SecurityActionsReadWriteAll
	case "SecurityAlert.Read.All":
		return azure.SecurityAlertReadAll
	case "SecurityAlert.ReadWrite.All":
		return azure.SecurityAlertReadWriteAll
	case "SecurityAnalyzedMessage.Read.All":
		return azure.SecurityAnalyzedMessageReadAll
	case "SecurityAnalyzedMessage.ReadWrite.All":
		return azure.SecurityAnalyzedMessageReadWriteAll
	case "SecurityCopilotWorkspaces.Read.All":
		return azure.SecurityCopilotWorkspacesReadAll
	case "SecurityCopilotWorkspaces.ReadWrite.All":
		return azure.SecurityCopilotWorkspacesReadWriteAll
	case "SecurityEvents.Read.All":
		return azure.SecurityEventsReadAll
	case "SecurityEvents.ReadWrite.All":
		return azure.SecurityEventsReadWriteAll
	case "SecurityIdentitiesAccount.Read.All":
		return azure.SecurityIdentitiesAccountReadAll
	case "SecurityIdentitiesActions.ReadWrite.All":
		return azure.SecurityIdentitiesActionsReadWriteAll
	case "SecurityIdentitiesHealth.Read.All":
		return azure.SecurityIdentitiesHealthReadAll
	case "SecurityIdentitiesHealth.ReadWrite.All":
		return azure.SecurityIdentitiesHealthReadWriteAll
	case "SecurityIdentitiesSensors.Read.All":
		return azure.SecurityIdentitiesSensorsReadAll
	case "SecurityIdentitiesSensors.ReadWrite.All":
		return azure.SecurityIdentitiesSensorsReadWriteAll
	case "SecurityIdentitiesUserActions.Read.All":
		return azure.SecurityIdentitiesUserActionsReadAll
	case "SecurityIdentitiesUserActions.ReadWrite.All":
		return azure.SecurityIdentitiesUserActionsReadWriteAll
	case "SecurityIncident.Read.All":
		return azure.SecurityIncidentReadAll
	case "SecurityIncident.ReadWrite.All":
		return azure.SecurityIncidentReadWriteAll
	case "SensitivityLabel.Evaluate":
		return azure.SensitivityLabelEvaluate
	case "SensitivityLabel.Evaluate.All":
		return azure.SensitivityLabelEvaluateAll
	case "SensitivityLabel.Read":
		return azure.SensitivityLabelRead
	case "SensitivityLabels.Read.All":
		return azure.SensitivityLabelsReadAll
	case "ServiceHealth.Read.All":
		return azure.ServiceHealthReadAll
	case "ServiceMessage.Read.All":
		return azure.ServiceMessageReadAll
	case "ServiceMessageViewpoint.Write":
		return azure.ServiceMessageViewpointWrite
	case "ServicePrincipalEndpoint.Read.All":
		return azure.ServicePrincipalEndpointReadAll
	case "ServicePrincipalEndpoint.ReadWrite.All":
		return azure.ServicePrincipalEndpointReadWriteAll
	case "SharePointTenantSettings.Read.All":
		return azure.SharePointTenantSettingsReadAll
	case "SharePointTenantSettings.ReadWrite.All":
		return azure.SharePointTenantSettingsReadWriteAll
	case "ShortNotes.Read":
		return azure.ShortNotesRead
	case "ShortNotes.Read.All":
		return azure.ShortNotesReadAll
	case "ShortNotes.ReadWrite":
		return azure.ShortNotesReadWrite
	case "ShortNotes.ReadWrite.All":
		return azure.ShortNotesReadWriteAll
	case "SignInIdentifier.Read.All":
		return azure.SignInIdentifierReadAll
	case "SignInIdentifier.ReadWrite.All":
		return azure.SignInIdentifierReadWriteAll
	case "Sites.Archive.All":
		return azure.SitesArchiveAll
	case "Sites.FullControl.All":
		return azure.SitesFullControlAll
	case "Sites.Manage.All":
		return azure.SitesManageAll
	case "Sites.Read.All":
		return azure.SitesReadAll
	case "Sites.ReadWrite.All":
		return azure.SitesReadWriteAll
	case "Sites.Selected":
		return azure.SitesSelected
	case "SpiffeTrustDomain.Read.All":
		return azure.SpiffeTrustDomainReadAll
	case "SpiffeTrustDomain.ReadWrite.All":
		return azure.SpiffeTrustDomainReadWriteAll
	case "Storyline.ReadWrite.All":
		return azure.StorylineReadWriteAll
	case "SubjectRightsRequest.Read.All":
		return azure.SubjectRightsRequestReadAll
	case "SubjectRightsRequest.ReadWrite.All":
		return azure.SubjectRightsRequestReadWriteAll
	case "Subscription.Read.All":
		return azure.SubscriptionReadAll
	case "Synchronization.Read.All":
		return azure.SynchronizationReadAll
	case "Synchronization.ReadWrite.All":
		return azure.SynchronizationReadWriteAll
	case "Tasks.Read":
		return azure.TasksRead
	case "Tasks.Read.All":
		return azure.TasksReadAll
	case "Tasks.Read.Shared":
		return azure.TasksReadShared
	case "Tasks.ReadWrite":
		return azure.TasksReadWrite
	case "Tasks.ReadWrite.All":
		return azure.TasksReadWriteAll
	case "Tasks.ReadWrite.Shared":
		return azure.TasksReadWriteShared
	case "Team.Create":
		return azure.TeamCreate
	case "Team.ReadBasic.All":
		return azure.TeamReadBasicAll
	case "TeamMember.Read.All":
		return azure.TeamMemberReadAll
	case "TeamMember.ReadWrite.All":
		return azure.TeamMemberReadWriteAll
	case "TeamMember.ReadWriteNonOwnerRole.All":
		return azure.TeamMemberReadWriteNonOwnerRoleAll
	case "TeamSettings.Read.All":
		return azure.TeamSettingsReadAll
	case "TeamSettings.ReadWrite.All":
		return azure.TeamSettingsReadWriteAll
	case "TeamTemplates.Read":
		return azure.TeamTemplatesRead
	case "TeamTemplates.Read.All":
		return azure.TeamTemplatesReadAll
	case "TeamsActivity.Read":
		return azure.TeamsActivityRead
	case "TeamsActivity.Read.All":
		return azure.TeamsActivityReadAll
	case "TeamsActivity.Send":
		return azure.TeamsActivitySend
	case "TeamsAppInstallation.ManageSelectedForChat":
		return azure.TeamsAppInstallationManageSelectedForChat
	case "TeamsAppInstallation.ManageSelectedForChat.All":
		return azure.TeamsAppInstallationManageSelectedForChatAll
	case "TeamsAppInstallation.ManageSelectedForTeam":
		return azure.TeamsAppInstallationManageSelectedForTeam
	case "TeamsAppInstallation.ManageSelectedForTeam.All":
		return azure.TeamsAppInstallationManageSelectedForTeamAll
	case "TeamsAppInstallation.ManageSelectedForUser":
		return azure.TeamsAppInstallationManageSelectedForUser
	case "TeamsAppInstallation.ManageSelectedForUser.All":
		return azure.TeamsAppInstallationManageSelectedForUserAll
	case "TeamsAppInstallation.Read.All":
		return azure.TeamsAppInstallationReadAll
	case "TeamsAppInstallation.ReadForChat":
		return azure.TeamsAppInstallationReadForChat
	case "TeamsAppInstallation.ReadForChat.All":
		return azure.TeamsAppInstallationReadForChatAll
	case "TeamsAppInstallation.ReadForTeam":
		return azure.TeamsAppInstallationReadForTeam
	case "TeamsAppInstallation.ReadForTeam.All":
		return azure.TeamsAppInstallationReadForTeamAll
	case "TeamsAppInstallation.ReadForUser":
		return azure.TeamsAppInstallationReadForUser
	case "TeamsAppInstallation.ReadForUser.All":
		return azure.TeamsAppInstallationReadForUserAll
	case "TeamsAppInstallation.ReadSelectedForChat":
		return azure.TeamsAppInstallationReadSelectedForChat
	case "TeamsAppInstallation.ReadSelectedForChat.All":
		return azure.TeamsAppInstallationReadSelectedForChatAll
	case "TeamsAppInstallation.ReadSelectedForTeam":
		return azure.TeamsAppInstallationReadSelectedForTeam
	case "TeamsAppInstallation.ReadSelectedForTeam.All":
		return azure.TeamsAppInstallationReadSelectedForTeamAll
	case "TeamsAppInstallation.ReadSelectedForUser":
		return azure.TeamsAppInstallationReadSelectedForUser
	case "TeamsAppInstallation.ReadSelectedForUser.All":
		return azure.TeamsAppInstallationReadSelectedForUserAll
	case "TeamsAppInstallation.ReadWriteAndConsentForChat":
		return azure.TeamsAppInstallationReadWriteAndConsentForChat
	case "TeamsAppInstallation.ReadWriteAndConsentForChat.All":
		return azure.TeamsAppInstallationReadWriteAndConsentForChatAll
	case "TeamsAppInstallation.ReadWriteAndConsentForTeam":
		return azure.TeamsAppInstallationReadWriteAndConsentForTeam
	case "TeamsAppInstallation.ReadWriteAndConsentForTeam.All":
		return azure.TeamsAppInstallationReadWriteAndConsentForTeamAll
	case "TeamsAppInstallation.ReadWriteAndConsentForUser":
		return azure.TeamsAppInstallationReadWriteAndConsentForUser
	case "TeamsAppInstallation.ReadWriteAndConsentForUser.All":
		return azure.TeamsAppInstallationReadWriteAndConsentForUserAll
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForChat":
		return azure.TeamsAppInstallationReadWriteAndConsentSelfForChat
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForChat.All":
		return azure.TeamsAppInstallationReadWriteAndConsentSelfForChatAll
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForTeam":
		return azure.TeamsAppInstallationReadWriteAndConsentSelfForTeam
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForTeam.All":
		return azure.TeamsAppInstallationReadWriteAndConsentSelfForTeamAll
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForUser":
		return azure.TeamsAppInstallationReadWriteAndConsentSelfForUser
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForUser.All":
		return azure.TeamsAppInstallationReadWriteAndConsentSelfForUserAll
	case "TeamsAppInstallation.ReadWriteForChat":
		return azure.TeamsAppInstallationReadWriteForChat
	case "TeamsAppInstallation.ReadWriteForChat.All":
		return azure.TeamsAppInstallationReadWriteForChatAll
	case "TeamsAppInstallation.ReadWriteForTeam":
		return azure.TeamsAppInstallationReadWriteForTeam
	case "TeamsAppInstallation.ReadWriteForTeam.All":
		return azure.TeamsAppInstallationReadWriteForTeamAll
	case "TeamsAppInstallation.ReadWriteForUser":
		return azure.TeamsAppInstallationReadWriteForUser
	case "TeamsAppInstallation.ReadWriteForUser.All":
		return azure.TeamsAppInstallationReadWriteForUserAll
	case "TeamsAppInstallation.ReadWriteSelectedForChat":
		return azure.TeamsAppInstallationReadWriteSelectedForChat
	case "TeamsAppInstallation.ReadWriteSelectedForChat.All":
		return azure.TeamsAppInstallationReadWriteSelectedForChatAll
	case "TeamsAppInstallation.ReadWriteSelectedForTeam":
		return azure.TeamsAppInstallationReadWriteSelectedForTeam
	case "TeamsAppInstallation.ReadWriteSelectedForTeam.All":
		return azure.TeamsAppInstallationReadWriteSelectedForTeamAll
	case "TeamsAppInstallation.ReadWriteSelectedForUser":
		return azure.TeamsAppInstallationReadWriteSelectedForUser
	case "TeamsAppInstallation.ReadWriteSelectedForUser.All":
		return azure.TeamsAppInstallationReadWriteSelectedForUserAll
	case "TeamsAppInstallation.ReadWriteSelfForChat":
		return azure.TeamsAppInstallationReadWriteSelfForChat
	case "TeamsAppInstallation.ReadWriteSelfForChat.All":
		return azure.TeamsAppInstallationReadWriteSelfForChatAll
	case "TeamsAppInstallation.ReadWriteSelfForTeam":
		return azure.TeamsAppInstallationReadWriteSelfForTeam
	case "TeamsAppInstallation.ReadWriteSelfForTeam.All":
		return azure.TeamsAppInstallationReadWriteSelfForTeamAll
	case "TeamsAppInstallation.ReadWriteSelfForUser":
		return azure.TeamsAppInstallationReadWriteSelfForUser
	case "TeamsAppInstallation.ReadWriteSelfForUser.All":
		return azure.TeamsAppInstallationReadWriteSelfForUserAll
	case "TeamsPolicyUserAssign.ReadWrite.All":
		return azure.TeamsPolicyUserAssignReadWriteAll
	case "TeamsResourceAccount.Read.All":
		return azure.TeamsResourceAccountReadAll
	case "TeamsTab.Create":
		return azure.TeamsTabCreate
	case "TeamsTab.Read.All":
		return azure.TeamsTabReadAll
	case "TeamsTab.ReadWrite.All":
		return azure.TeamsTabReadWriteAll
	case "TeamsTab.ReadWriteForChat":
		return azure.TeamsTabReadWriteForChat
	case "TeamsTab.ReadWriteForChat.All":
		return azure.TeamsTabReadWriteForChatAll
	case "TeamsTab.ReadWriteForTeam":
		return azure.TeamsTabReadWriteForTeam
	case "TeamsTab.ReadWriteForTeam.All":
		return azure.TeamsTabReadWriteForTeamAll
	case "TeamsTab.ReadWriteForUser":
		return azure.TeamsTabReadWriteForUser
	case "TeamsTab.ReadWriteForUser.All":
		return azure.TeamsTabReadWriteForUserAll
	case "TeamsTab.ReadWriteSelfForChat":
		return azure.TeamsTabReadWriteSelfForChat
	case "TeamsTab.ReadWriteSelfForChat.All":
		return azure.TeamsTabReadWriteSelfForChatAll
	case "TeamsTab.ReadWriteSelfForTeam":
		return azure.TeamsTabReadWriteSelfForTeam
	case "TeamsTab.ReadWriteSelfForTeam.All":
		return azure.TeamsTabReadWriteSelfForTeamAll
	case "TeamsTab.ReadWriteSelfForUser":
		return azure.TeamsTabReadWriteSelfForUser
	case "TeamsTab.ReadWriteSelfForUser.All":
		return azure.TeamsTabReadWriteSelfForUserAll
	case "TeamsTelephoneNumber.Read.All":
		return azure.TeamsTelephoneNumberReadAll
	case "TeamsTelephoneNumber.ReadWrite.All":
		return azure.TeamsTelephoneNumberReadWriteAll
	case "TeamsUserConfiguration.Read.All":
		return azure.TeamsUserConfigurationReadAll
	case "Teamwork.Migrate.All":
		return azure.TeamworkMigrateAll
	case "Teamwork.Read.All":
		return azure.TeamworkReadAll
	case "TeamworkAppSettings.Read.All":
		return azure.TeamworkAppSettingsReadAll
	case "TeamworkAppSettings.ReadWrite.All":
		return azure.TeamworkAppSettingsReadWriteAll
	case "TeamworkDevice.Read.All":
		return azure.TeamworkDeviceReadAll
	case "TeamworkDevice.ReadWrite.All":
		return azure.TeamworkDeviceReadWriteAll
	case "TeamworkTag.Read":
		return azure.TeamworkTagRead
	case "TeamworkTag.Read.All":
		return azure.TeamworkTagReadAll
	case "TeamworkTag.ReadWrite":
		return azure.TeamworkTagReadWrite
	case "TeamworkTag.ReadWrite.All":
		return azure.TeamworkTagReadWriteAll
	case "TeamworkUserInteraction.Read.All":
		return azure.TeamworkUserInteractionReadAll
	case "TermStore.Read.All":
		return azure.TermStoreReadAll
	case "TermStore.ReadWrite.All":
		return azure.TermStoreReadWriteAll
	case "ThreatAssessment.Read.All":
		return azure.ThreatAssessmentReadAll
	case "ThreatAssessment.ReadWrite.All":
		return azure.ThreatAssessmentReadWriteAll
	case "ThreatHunting.Read.All":
		return azure.ThreatHuntingReadAll
	case "ThreatIndicators.Read.All":
		return azure.ThreatIndicatorsReadAll
	case "ThreatIndicators.ReadWrite.OwnedBy":
		return azure.ThreatIndicatorsReadWriteOwnedBy
	case "ThreatIntelligence.Read.All":
		return azure.ThreatIntelligenceReadAll
	case "ThreatSubmission.Read":
		return azure.ThreatSubmissionRead
	case "ThreatSubmission.Read.All":
		return azure.ThreatSubmissionReadAll
	case "ThreatSubmission.ReadWrite":
		return azure.ThreatSubmissionReadWrite
	case "ThreatSubmission.ReadWrite.All":
		return azure.ThreatSubmissionReadWriteAll
	case "ThreatSubmissionPolicy.ReadWrite.All":
		return azure.ThreatSubmissionPolicyReadWriteAll
	case "Topic.Read.All":
		return azure.TopicReadAll
	case "TrustFrameworkKeySet.Read.All":
		return azure.TrustFrameworkKeySetReadAll
	case "TrustFrameworkKeySet.ReadWrite.All":
		return azure.TrustFrameworkKeySetReadWriteAll
	case "UnifiedGroupMember.Read.AsGuest":
		return azure.UnifiedGroupMemberReadAsGuest
	case "User.DeleteRestore.All":
		return azure.UserDeleteRestoreAll
	case "User.EnableDisableAccount.All":
		return azure.UserEnableDisableAccountAll
	case "User.Export.All":
		return azure.UserExportAll
	case "User.Invite.All":
		return azure.UserInviteAll
	case "User.ManageIdentities.All":
		return azure.UserManageIdentitiesAll
	case "User.Read":
		return azure.UserRead
	case "User.Read.All":
		return azure.UserReadAll
	case "User.ReadBasic.All":
		return azure.UserReadBasicAll
	case "User.ReadWrite":
		return azure.UserReadWrite
	case "User.ReadWrite.All":
		return azure.UserReadWriteAll
	case "User.ReadWrite.CrossCloud":
		return azure.UserReadWriteCrossCloud
	case "User.RevokeSessions.All":
		return azure.UserRevokeSessionsAll
	case "UserActivity.ReadWrite.CreatedByApp":
		return azure.UserActivityReadWriteCreatedByApp
	case "UserAuthenticationMethod.Read":
		return azure.UserAuthenticationMethodRead
	case "UserAuthenticationMethod.Read.All":
		return azure.UserAuthenticationMethodReadAll
	case "UserAuthenticationMethod.ReadWrite":
		return azure.UserAuthenticationMethodReadWrite
	case "UserAuthenticationMethod.ReadWrite.All":
		return azure.UserAuthenticationMethodReadWriteAll
	case "UserCloudClipboard.Read":
		return azure.UserCloudClipboardRead
	case "UserNotification.ReadWrite.CreatedByApp":
		return azure.UserNotificationReadWriteCreatedByApp
	case "UserState.ReadWrite.All":
		return azure.UserStateReadWriteAll
	case "UserShiftPreferences.Read.All":
		return azure.UserShiftPreferencesReadAll
	case "UserShiftPreferences.ReadWrite.All":
		return azure.UserShiftPreferencesReadWriteAll
	case "UserTeamwork.Read":
		return azure.UserTeamworkRead
	case "UserTeamwork.Read.All":
		return azure.UserTeamworkReadAll
	case "UserTimelineActivity.Write.CreatedByApp":
		return azure.UserTimelineActivityWriteCreatedByApp
	case "UserWindowsSettings.Read.All":
		return azure.UserWindowsSettingsReadAll
	case "UserWindowsSettings.ReadWrite.All":
		return azure.UserWindowsSettingsReadWriteAll
	case "VirtualAppointment.Read":
		return azure.VirtualAppointmentRead
	case "VirtualAppointment.Read.All":
		return azure.VirtualAppointmentReadAll
	case "VirtualAppointment.ReadWrite":
		return azure.VirtualAppointmentReadWrite
	case "VirtualAppointment.ReadWrite.All":
		return azure.VirtualAppointmentReadWriteAll
	case "VirtualAppointmentNotification.Send":
		return azure.VirtualAppointmentNotificationSend
	case "VirtualEvent.Read":
		return azure.VirtualEventRead
	case "VirtualEvent.Read.All":
		return azure.VirtualEventReadAll
	case "VirtualEvent.ReadWrite":
		return azure.VirtualEventReadWrite
	case "WindowsUpdates.ReadWrite.All":
		return azure.WindowsUpdatesReadWriteAll
	case "WorkforceIntegration.Read.All":
		return azure.WorkforceIntegrationReadAll
	case "WorkforceIntegration.ReadWrite.All":
		return azure.WorkforceIntegrationReadWriteAll
	default:
		return graph.StringKind(scope)
	}
}

func ConvertAzureOAuth2PermissionGrantToRels(data OAuth2PermissionGrant) []IngestibleRelationship {
	var relationships []IngestibleRelationship

	// Seperate the Scope into individual permissions
	scopes := strings.Split(data.Scope, " ")
	scopeSet := make(map[string]struct{})
	for _, scope := range scopes {
		scopeSet[strings.TrimSpace(scope)] = struct{}{}
	}
	// If the ConsentType is "Principal", we create a relationship from the Principal to the Tenant
	// If the ConsentType is "AllPrincipals", we create a relationship from the Tenant to the Tenant

	if data.ConsentType == "Principal" && data.TenantId != "" {

		// Create a relationship for each scope
		for scope := range scopeSet {
			if scope == "" || scope == "null" || scope == "openid" || scope == "profile" || scope == "email" || scope == "offline_access" || scope == " " {
				continue
			}
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: strings.ToUpper(data.PrincipalId),
					Kind:  azure.User,
				},
				IngestibleEndpoint{
					Value: strings.ToUpper(data.ClientId),
					Kind:  azure.Tenant,
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  GetPermissionConstant(scope),
				},
			))
		}
	} else if data.ConsentType == "AllPrincipals" && data.TenantId != "" {
		for scope := range scopeSet {
			if scope == "" || scope == "null" || scope == "openid" || scope == "profile" || scope == "email" || scope == "offline_access" || scope == " " {
				continue
			}
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: strings.ToUpper(data.TenantId),
					Kind:  azure.Tenant,
				},
				IngestibleEndpoint{
					Value: strings.ToUpper(data.ClientId),
					Kind:  azure.ServicePrincipal,
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  GetPermissionConstant(scope),
				},
			))
		}

	}
	return relationships
}
