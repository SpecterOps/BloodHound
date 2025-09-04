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
	//If the scope is not the directory, we are going to skip creating the edges for now until later work is done
	//This isn't necessarily the best spot for this, but it works and it makes testing simple, since none of our azure convertors are exported
	if instance.DirectoryScopeId != "/" {
		return relationships
	}
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
		return azure.AZMGAPIConnectorsReadAll
	case "APIConnectors.ReadWrite.All":
		return azure.AZMGAPIConnectorsReadWriteAll
	case "AccessReview.Read.All":
		return azure.AZMGAccessReviewReadAll
	case "AccessReview.ReadWrite.All":
		return azure.AZMGAccessReviewReadWriteAll
	case "AccessReview.ReadWrite.Membership":
		return azure.AZMGAccessReviewReadWriteMembership
	case "Acronym.Read.All":
		return azure.AZMGAccessReviewReadWriteMembership
	case "AdministrativeUnit.Read.All":
		return azure.AZMGAdministrativeUnitReadAll
	case "AdministrativeUnit.ReadWrite.All":
		return azure.AZMGAdministrativeUnitReadWriteAll
	case "AgentApplication.Create":
		return azure.AZMGAgentApplicationCreate
	case "AgentIdentity.Create":
		return azure.AZMGAgentIdentityCreate
	case "Agreement.Read.All":
		return azure.AZMGAgreementReadAll
	case "Agreement.ReadWrite.All":
		return azure.AZMGAgreementReadWriteAll
	case "AgreementAcceptance.Read":
		return azure.AZMGAgreementAcceptanceRead
	case "AgreementAcceptance.Read.All":
		return azure.AZMGAgreementAcceptanceReadAll
	case "AiEnterpriseInteraction.Read":
		return azure.AZMGAiEnterpriseInteractionRead
	case "AiEnterpriseInteraction.Read.All":
		return azure.AZMGAiEnterpriseInteractionReadAll
	case "AllSites.Read":
		return azure.AZMGAllSitesRead
	case "Analytics.Read":
		return azure.AZMGAnalyticsRead
	case "AppCatalog.Read.All":
		return azure.AZMGAppCatalogReadAll
	case "AppCatalog.ReadWrite.All":
		return azure.AZMGAppCatalogReadWriteAll
	case "AppCatalog.Submit":
		return azure.AZMGAppCatalogSubmit
	case "AppCertTrustConfiguration.Read.All":
		return azure.AZMGAppCertTrustConfigurationReadAll
	case "AppCertTrustConfiguration.ReadWrite.All":
		return azure.AZMGAppCertTrustConfigurationReadWriteAll
	case "AppRoleAssignment.ReadWrite.All":
		return azure.AZMGAppRoleAssignmentReadWriteAll
	case "Application.Read.All":
		return azure.AZMGApplicationReadAll
	case "Application.ReadWrite.All":
		return azure.AZMGApplicationReadWriteAll
	case "Application.ReadWrite.OwnedBy":
		return azure.AZMGApplicationReadWriteOwnedBy
	case "ApprovalSolution.Read":
		return azure.AZMGApprovalSolutionRead
	case "ApprovalSolution.Read.All":
		return azure.AZMGApprovalSolutionReadAll
	case "ApprovalSolution.ReadWrite":
		return azure.AZMGApprovalSolutionReadWrite
	case "ApprovalSolution.ReadWrite.All":
		return azure.AZMGApprovalSolutionReadWriteAll
	case "ApprovalSolutionResponse.ReadWrite":
		return azure.AZMGApprovalSolutionResponseReadWrite
	case "AttackSimulation.Read.All":
		return azure.AZMGAttackSimulationReadAll
	case "AttackSimulation.ReadWrite.All":
		return azure.AZMGAttackSimulationReadWriteAll
	case "AuditActivity.Read":
		return azure.AZMGAuditActivityRead
	case "AuditActivity.Write":
		return azure.AZMGAuditActivityWrite
	case "AuditLog.Read.All":
		return azure.AZMGAuditLogReadAll
	case "AuditLogsQuery.Read.All":
		return azure.AZMGAuditLogsQueryReadAll
	case "AuthenticationContext.Read.All":
		return azure.AZMGAuthenticationContextReadAll
	case "AuthenticationContext.ReadWrite.All":
		return azure.AZMGAuthenticationContextReadWriteAll
	case "BillingConfiguration.ReadWrite.All":
		return azure.AZMGBillingConfigurationReadWriteAll
	case "BitlockerKey.Read.All":
		return azure.AZMGBitlockerKeyReadAll
	case "BitlockerKey.ReadBasic.All":
		return azure.AZMGBitlockerKeyReadBasicAll
	case "Bookings.Manage.All":
		return azure.AZMGBookingsManageAll
	case "Bookings.Read.All":
		return azure.AZMGBookingsReadAll
	case "Bookings.ReadWrite.All":
		return azure.AZMGBookingsReadWriteAll
	case "BookingsAppointment.ReadWrite.All":
		return azure.AZMGBookingsAppointmentReadWriteAll
	case "Bookmark.Read.All":
		return azure.AZMGBookmarkReadAll
	case "BrowserSiteLists.Read.All":
		return azure.AZMGBrowserSiteListsReadAll
	case "BrowserSiteLists.ReadWrite.All":
		return azure.AZMGBrowserSiteListsReadWriteAll
	case "BusinessScenarioConfig.Read.All":
		return azure.AZMGBusinessScenarioConfigReadAll
	case "BusinessScenarioConfig.Read.OwnedBy":
		return azure.AZMGBusinessScenarioConfigReadOwnedBy
	case "BusinessScenarioConfig.ReadWrite.All":
		return azure.AZMGBusinessScenarioConfigReadWriteAll
	case "BusinessScenarioConfig.ReadWrite.OwnedBy":
		return azure.AZMGBusinessScenarioConfigReadWriteOwnedBy
	case "BusinessScenarioData.Read.OwnedBy":
		return azure.AZMGBusinessScenarioDataReadOwnedBy
	case "BusinessScenarioData.ReadWrite.OwnedBy":
		return azure.AZMGBusinessScenarioDataReadWriteOwnedBy
	case "Calendars.Read":
		return azure.AZMGCalendarsRead
	case "Calendars.Read.Shared":
		return azure.AZMGCalendarsReadShared
	case "Calendars.ReadBasic":
		return azure.AZMGCalendarsReadBasic
	case "Calendars.ReadBasic.All":
		return azure.AZMGCalendarsReadBasicAll
	case "Calendars.ReadWrite":
		return azure.AZMGCalendarsReadWrite
	case "Calendars.ReadWrite.Shared":
		return azure.AZMGCalendarsReadWriteShared
	case "CallAiInsights.Read.All":
		return azure.AZMGCallAiInsightsReadAll
	case "CallDelegation.Read":
		return azure.AZMGCallDelegationRead
	case "CallDelegation.Read.All":
		return azure.AZMGCallDelegationReadAll
	case "CallDelegation.ReadWrite":
		return azure.AZMGCallDelegationReadWrite
	case "CallDelegation.ReadWrite.All":
		return azure.AZMGCallDelegationReadWriteAll
	case "CallEvents.Read":
		return azure.AZMGCallEventsRead
	case "CallEvents.Read.All":
		return azure.AZMGCallEventsReadAll
	case "CallRecords.Read.All":
		return azure.AZMGCallRecordsReadAll
	case "Calls.AccessMedia.All":
		return azure.AZMGCallsAccessMediaAll
	case "Calls.Initiate.All":
		return azure.AZMGCallsInitiateAll
	case "Calls.InitiateGroupCall.All":
		return azure.AZMGCallsInitiateGroupCallAll
	case "Calls.JoinGroupCall.All":
		return azure.AZMGCallsJoinGroupCallAll
	case "Calls.JoinGroupCallAsGuest.All":
		return azure.AZMGCallsJoinGroupCallAsGuestAll
	case "ChangeManagement.Read.All":
		return azure.AZMGChangeManagementReadAll
	case "Channel.Create":
		return azure.AZMGChannelCreate
	case "Channel.Delete.All":
		return azure.AZMGChannelDeleteAll
	case "Channel.ReadBasic.All":
		return azure.AZMGChannelReadBasicAll
	case "ChannelMember.Read.All":
		return azure.AZMGChannelMemberReadAll
	case "ChannelMember.ReadWrite.All":
		return azure.AZMGChannelMemberReadWriteAll
	case "ChannelMessage.Edit":
		return azure.AZMGChannelMessageEdit
	case "ChannelMessage.Read.All":
		return azure.AZMGChannelMessageReadAll
	case "ChannelMessage.ReadWrite":
		return azure.AZMGChannelMessageReadWrite
	case "ChannelMessage.Send":
		return azure.AZMGChannelMessageSend
	case "ChannelMessage.UpdatePolicyViolation.All":
		return azure.AZMGChannelMessageUpdatePolicyViolationAll
	case "ChannelSettings.Read.All":
		return azure.AZMGChannelSettingsReadAll
	case "ChannelSettings.ReadWrite.All":
		return azure.AZMGChannelSettingsReadWriteAll
	case "Chat.Create":
		return azure.AZMGChatCreate
	case "Chat.ManageDeletion.All":
		return azure.AZMGChatManageDeletionAll
	case "Chat.Read":
		return azure.AZMGChatRead
	case "Chat.Read.All":
		return azure.AZMGChatReadAll
	case "Chat.Read.WhereInstalled":
		return azure.AZMGChatReadWhereInstalled
	case "Chat.ReadBasic":
		return azure.AZMGChatReadBasic
	case "Chat.ReadBasic.All":
		return azure.AZMGChatReadBasicAll
	case "Chat.ReadBasic.WhereInstalled":
		return azure.AZMGChatReadBasicWhereInstalled
	case "Chat.ReadWrite":
		return azure.AZMGChatReadWrite
	case "Chat.ReadWrite.All":
		return azure.AZMGChatReadWriteAll
	case "Chat.ReadWrite.WhereInstalled":
		return azure.AZMGChatReadWriteWhereInstalled
	case "Chat.UpdatePolicyViolation.All":
		return azure.AZMGChatUpdatePolicyViolationAll
	case "ChatMember.Read":
		return azure.AZMGChatMemberRead
	case "ChatMember.Read.All":
		return azure.AZMGChatMemberReadAll
	case "ChatMember.Read.WhereInstalled":
		return azure.AZMGChatMemberReadWhereInstalled
	case "ChatMember.ReadWrite":
		return azure.AZMGChatMemberReadWrite
	case "ChatMember.ReadWrite.All":
		return azure.AZMGChatMemberReadWriteAll
	case "ChatMember.ReadWrite.WhereInstalled":
		return azure.AZMGChatMemberReadWriteWhereInstalled
	case "ChatMessage.Read":
		return azure.AZMGChatMessageRead
	case "ChatMessage.Read.All":
		return azure.AZMGChatMessageReadAll
	case "ChatMessage.Send":
		return azure.AZMGChatMessageSend
	case "CloudPC.Read.All":
		return azure.AZMGCloudPCReadAll
	case "CloudPC.ReadWrite.All":
		return azure.AZMGCloudPCReadWriteAll
	case "Community.Read.All":
		return azure.AZMGCommunityReadAll
	case "Community.ReadWrite.All":
		return azure.AZMGCommunityReadWriteAll
	case "ConfigurationMonitoring.Read.All":
		return azure.AZMGConfigurationMonitoringReadAll
	case "ConfigurationMonitoring.ReadWrite.All":
		return azure.AZMGConfigurationMonitoringReadWriteAll
	case "ConsentRequest.Create":
		return azure.AZMGConsentRequestCreate
	case "ConsentRequest.Read":
		return azure.AZMGConsentRequestRead
	case "ConsentRequest.Read.All":
		return azure.AZMGConsentRequestReadAll
	case "ConsentRequest.ReadApprove.All":
		return azure.AZMGConsentRequestReadApproveAll
	case "ConsentRequest.ReadWrite.All":
		return azure.AZMGConsentRequestReadWriteAll
	case "Contacts.Read":
		return azure.AZMGContactsRead
	case "Contacts.Read.Shared":
		return azure.AZMGContactsReadShared
	case "Contacts.ReadWrite":
		return azure.AZMGContactsReadWrite
	case "Contacts.ReadWrite.Shared":
		return azure.AZMGContactsReadWriteShared
	case "Content.Process.All":
		return azure.AZMGContentProcessAll
	case "Content.Process.User":
		return azure.AZMGContentProcessUser
	case "ContentActivity.Read":
		return azure.AZMGContentActivityRead
	case "ContentActivity.Write":
		return azure.AZMGContentActivityWrite
	case "CrossTenantInformation.ReadBasic.All":
		return azure.AZMGCrossTenantInformationReadBasicAll
	case "CrossTenantUserProfileSharing.Read":
		return azure.AZMGCrossTenantUserProfileSharingRead
	case "CrossTenantUserProfileSharing.Read.All":
		return azure.AZMGCrossTenantUserProfileSharingReadAll
	case "CrossTenantUserProfileSharing.ReadWrite":
		return azure.AZMGCrossTenantUserProfileSharingReadWrite
	case "CrossTenantUserProfileSharing.ReadWrite.All":
		return azure.AZMGCrossTenantUserProfileSharingReadWriteAll
	case "CustomAuthenticationExtension.Read.All":
		return azure.AZMGCustomAuthenticationExtensionReadAll
	case "CustomAuthenticationExtension.ReadWrite.All":
		return azure.AZMGCustomAuthenticationExtensionReadWriteAll
	case "CustomAuthenticationExtension.Receive.Payload":
		return azure.AZMGCustomAuthenticationExtensionReceivePayload
	case "CustomDetection.Read.All":
		return azure.AZMGCustomDetectionReadAll
	case "CustomDetection.ReadWrite.All":
		return azure.AZMGCustomDetectionReadWriteAll
	case "CustomSecAttributeAssignment.Read.All":
		return azure.AZMGCustomSecAttributeAssignmentReadAll
	case "CustomSecAttributeAssignment.ReadWrite.All":
		return azure.AZMGCustomSecAttributeAssignmentReadWriteAll
	case "CustomSecAttributeAuditLogs.Read.All":
		return azure.AZMGCustomSecAttributeAuditLogsReadAll
	case "CustomSecAttributeDefinition.Read.All":
		return azure.AZMGCustomSecAttributeDefinitionReadAll
	case "CustomSecAttributeDefinition.ReadWrite.All":
		return azure.AZMGCustomSecAttributeDefinitionReadWriteAll
	case "CustomSecAttributeProvisioning.Read.All":
		return azure.AZMGCustomSecAttributeProvisioningReadAll
	case "CustomSecAttributeProvisioning.ReadWrite.All":
		return azure.AZMGCustomSecAttributeProvisioningReadWriteAll
	case "CustomTags.Read.All":
		return azure.AZMGCustomTagsReadAll
	case "CustomTags.ReadWrite.All":
		return azure.AZMGCustomTagsReadWriteAll
	case "Dataset.Read.All":
		return azure.AZMGDatasetReadAll
	case "DelegatedAdminRelationship.Read.All":
		return azure.AZMGDelegatedAdminRelationshipReadAll
	case "DelegatedAdminRelationship.ReadWrite.All":
		return azure.AZMGDelegatedAdminRelationshipReadWriteAll
	case "DelegatedPermissionGrant.Read.All":
		return azure.AZMGDelegatedPermissionGrantReadAll
	case "DelegatedPermissionGrant.ReadWrite.All":
		return azure.AZMGDelegatedPermissionGrantReadWriteAll
	case "Device.Command":
		return azure.AZMGDeviceCommand
	case "Device.CreateFromOwnedTemplate":
		return azure.AZMGDeviceCreateFromOwnedTemplate
	case "Device.Read":
		return azure.AZMGDeviceRead
	case "Device.Read.All":
		return azure.AZMGDeviceReadAll
	case "Device.ReadWrite.All":
		return azure.AZMGDeviceReadWriteAll
	case "DeviceLocalCredential.Read.All":
		return azure.AZMGDeviceLocalCredentialReadAll
	case "DeviceLocalCredential.ReadBasic.All":
		return azure.AZMGDeviceLocalCredentialReadBasicAll
	case "DeviceManagementApps.Read.All":
		return azure.AZMGDeviceManagementAppsReadAll
	case "DeviceManagementApps.ReadWrite.All":
		return azure.AZMGDeviceManagementAppsReadWriteAll
	case "DeviceManagementCloudCA.Read.All":
		return azure.AZMGDeviceManagementCloudCAReadAll
	case "DeviceManagementCloudCA.ReadWrite.All":
		return azure.AZMGDeviceManagementCloudCAReadWriteAll
	case "DeviceManagementConfiguration.Read.All":
		return azure.AZMGDeviceManagementConfigurationReadAll
	case "DeviceManagementConfiguration.ReadWrite.All":
		return azure.AZMGDeviceManagementConfigurationReadWriteAll
	case "DeviceManagementManagedDevices.PrivilegedOperations.All":
		return azure.AZMGDeviceManagementManagedDevicesPrivilegedOperationsAll
	case "DeviceManagementManagedDevices.Read.All":
		return azure.AZMGDeviceManagementManagedDevicesReadAll
	case "DeviceManagementManagedDevices.ReadWrite.All":
		return azure.AZMGDeviceManagementManagedDevicesReadWriteAll
	case "DeviceManagementRBAC.Read.All":
		return azure.AZMGDeviceManagementRBACReadAll
	case "DeviceManagementRBAC.ReadWrite.All":
		return azure.AZMGDeviceManagementRBACReadWriteAll
	case "DeviceManagementScripts.Read.All":
		return azure.AZMGDeviceManagementScriptsReadAll
	case "DeviceManagementScripts.ReadWrite.All":
		return azure.AZMGDeviceManagementScriptsReadWriteAll
	case "DeviceManagementServiceConfig.Read.All":
		return azure.AZMGDeviceManagementServiceConfigReadAll
	case "DeviceManagementServiceConfig.ReadWrite.All":
		return azure.AZMGDeviceManagementServiceConfigReadWriteAll
	case "DeviceTemplate.Create":
		return azure.AZMGDeviceTemplateCreate
	case "DeviceTemplate.Read.All":
		return azure.AZMGDeviceTemplateReadAll
	case "DeviceTemplate.ReadWrite.All":
		return azure.AZMGDeviceTemplateReadWriteAll
	case "Directory.AccessAsUser.All":
		return azure.AZMGDirectoryAccessAsUserAll
	case "Directory.Read.All":
		return azure.AZMGDirectoryReadAll
	case "Directory.ReadWrite.All":
		return azure.AZMGDirectoryReadWriteAll
	case "DirectoryRecommendations.Read.All":
		return azure.AZMGDirectoryRecommendationsReadAll
	case "DirectoryRecommendations.ReadWrite.All":
		return azure.AZMGDirectoryRecommendationsReadWriteAll
	case "Domain.Read.All":
		return azure.AZMGDomainReadAll
	case "Domain.ReadWrite.All":
		return azure.AZMGDomainReadWriteAll
	case "EAS.AccessAsUser.All":
		return azure.AZMGEASAccessAsUserAll
	case "EWS.AccessAsUser.All":
		return azure.AZMGEWSAccessAsUserAll
	case "EduAdministration.Read":
		return azure.AZMGEduAdministrationRead
	case "EduAdministration.Read.All":
		return azure.AZMGEduAdministrationReadAll
	case "EduAdministration.ReadWrite":
		return azure.AZMGEduAdministrationReadWrite
	case "EduAdministration.ReadWrite.All":
		return azure.AZMGEduAdministrationReadWriteAll
	case "EduAssignments.Read":
		return azure.AZMGEduAssignmentsRead
	case "EduAssignments.Read.All":
		return azure.AZMGEduAssignmentsReadAll
	case "EduAssignments.ReadBasic":
		return azure.AZMGEduAssignmentsReadBasic
	case "EduAssignments.ReadBasic.All":
		return azure.AZMGEduAssignmentsReadBasicAll
	case "EduAssignments.ReadWrite":
		return azure.AZMGEduAssignmentsReadWrite
	case "EduAssignments.ReadWrite.All":
		return azure.AZMGEduAssignmentsReadWriteAll
	case "EduAssignments.ReadWriteBasic":
		return azure.AZMGEduAssignmentsReadWriteBasic
	case "EduAssignments.ReadWriteBasic.All":
		return azure.AZMGEduAssignmentsReadWriteBasicAll
	case "EduCurricula.Read":
		return azure.AZMGEduCurriculaRead
	case "EduCurricula.Read.All":
		return azure.AZMGEduCurriculaReadAll
	case "EduCurricula.ReadWrite":
		return azure.AZMGEduCurriculaReadWrite
	case "EduCurricula.ReadWrite.All":
		return azure.AZMGEduCurriculaReadWriteAll
	case "EduRoster.Read":
		return azure.AZMGEduRosterRead
	case "EduRoster.Read.All":
		return azure.AZMGEduRosterReadAll
	case "EduRoster.ReadBasic":
		return azure.AZMGEduRosterReadBasic
	case "EduRoster.ReadBasic.All":
		return azure.AZMGEduRosterReadBasicAll
	case "EduRoster.ReadWrite":
		return azure.AZMGEduRosterReadWrite
	case "EduRoster.ReadWrite.All":
		return azure.AZMGEduRosterReadWriteAll
	case "EngagementConversation.Migration.All":
		return azure.AZMGEngagementConversationMigrationAll
	case "EngagementConversation.ReadWrite.All":
		return azure.AZMGEngagementConversationReadWriteAll
	case "EngagementMeetingConversation.Read.All":
		return azure.AZMGEngagementMeetingConversationReadAll
	case "EngagementRole.Read":
		return azure.AZMGEngagementRoleRead
	case "EngagementRole.Read.All":
		return azure.AZMGEngagementRoleReadAll
	case "EngagementRole.ReadWrite.All":
		return azure.AZMGEngagementRoleReadWriteAll
	case "EntitlementManagement.Read.All":
		return azure.AZMGEntitlementManagementReadAll
	case "EntitlementManagement.ReadWrite.All":
		return azure.AZMGEntitlementManagementReadWriteAll
	case "EventListener.Read.All":
		return azure.AZMGEventListenerReadAll
	case "EventListener.ReadWrite.All":
		return azure.AZMGEventListenerReadWriteAll
	case "ExternalConnection.Read.All":
		return azure.AZMGExternalConnectionReadAll
	case "ExternalConnection.ReadWrite.All":
		return azure.AZMGExternalConnectionReadWriteAll
	case "ExternalConnection.ReadWrite.OwnedBy":
		return azure.AZMGExternalConnectionReadWriteOwnedBy
	case "ExternalItem.Read.All":
		return azure.AZMGExternalItemReadAll
	case "ExternalItem.ReadWrite.All":
		return azure.AZMGExternalItemReadWriteAll
	case "ExternalItem.ReadWrite.OwnedBy":
		return azure.AZMGExternalItemReadWriteOwnedBy
	case "ExternalUserProfile.Read.All":
		return azure.AZMGExternalUserProfileReadAll
	case "ExternalUserProfile.ReadWrite.All":
		return azure.AZMGExternalUserProfileReadWriteAll
	case "Family.Read":
		return azure.AZMGFamilyRead
	case "FileIngestion.Ingest":
		return azure.AZMGFileIngestionIngest
	case "FileIngestionHybridOnboarding.Manage":
		return azure.AZMGFileIngestionHybridOnboardingManage
	case "FileStorageContainer.Manage.All":
		return azure.AZMGFileStorageContainerManageAll
	case "FileStorageContainer.Selected":
		return azure.AZMGFileStorageContainerSelected
	case "FileStorageContainerTypeReg.Selected":
		return azure.AZMGFileStorageContainerTypeRegSelected
	case "Files.Read":
		return azure.AZMGFilesRead
	case "Files.Read.All":
		return azure.AZMGFilesReadAll
	case "Files.Read.Selected":
		return azure.AZMGFilesReadSelected
	case "Files.ReadWrite":
		return azure.AZMGFilesReadWrite
	case "Files.ReadWrite.All":
		return azure.AZMGFilesReadWriteAll
	case "Files.ReadWrite.AppFolder":
		return azure.AZMGFilesReadWriteAppFolder
	case "Files.ReadWrite.Selected":
		return azure.AZMGFilesReadWriteSelected
	case "Files.SelectedOperations.Selected":
		return azure.AZMGFilesSelectedOperationsSelected
	case "Financials.ReadWrite.All":
		return azure.AZMGFinancialsReadWriteAll
	case "Forms.ReadWrite":
		return azure.AZMGFormsReadWrite
	case "Group.Create":
		return azure.AZMGGroupCreate
	case "Group.Read.All":
		return azure.AZMGGroupReadAll
	case "Group.ReadWrite.All":
		return azure.AZMGGroupReadWriteAll
	case "Group-Conversation.ReadWrite.All":
		return azure.AZMGGroupConversationReadWriteAll
	case "GroupMember.Read.All":
		return azure.AZMGGroupMemberReadAll
	case "GroupMember.ReadWrite.All":
		return azure.AZMGGroupMemberReadWriteAll
	case "GroupSettings.Read.All":
		return azure.AZMGGroupSettingsReadAll
	case "GroupSettings.ReadWrite.All":
		return azure.AZMGGroupSettingsReadWriteAll
	case "HealthMonitoringAlert.Read.All":
		return azure.AZMGHealthMonitoringAlertReadAll
	case "HealthMonitoringAlert.ReadWrite.All":
		return azure.AZMGHealthMonitoringAlertReadWriteAll
	case "HealthMonitoringAlertConfig.Read.All":
		return azure.AZMGHealthMonitoringAlertConfigReadAll
	case "HealthMonitoringAlertConfig.ReadWrite.All":
		return azure.AZMGHealthMonitoringAlertConfigReadWriteAll
	case "IMAP.AccessAsUser.All":
		return azure.AZMGIMAPAccessAsUserAll
	case "IdentityProvider.Read.All":
		return azure.AZMGIdentityProviderReadAll
	case "IdentityProvider.ReadWrite.All":
		return azure.AZMGIdentityProviderReadWriteAll
	case "IdentityRiskEvent.Read.All":
		return azure.AZMGIdentityRiskEventReadAll
	case "IdentityRiskEvent.ReadWrite.All":
		return azure.AZMGIdentityRiskEventReadWriteAll
	case "IdentityRiskyServicePrincipal.Read.All":
		return azure.AZMGIdentityRiskyServicePrincipalReadAll
	case "IdentityRiskyServicePrincipal.ReadWrite.All":
		return azure.AZMGIdentityRiskyServicePrincipalReadWriteAll
	case "IdentityRiskyUser.Read.All":
		return azure.AZMGIdentityRiskyUserReadAll
	case "IdentityRiskyUser.ReadWrite.All":
		return azure.AZMGIdentityRiskyUserReadWriteAll
	case "IdentityUserFlow.Read.All":
		return azure.AZMGIdentityUserFlowReadAll
	case "IdentityUserFlow.ReadWrite.All":
		return azure.AZMGIdentityUserFlowReadWriteAll
	case "IndustryData.ReadBasic.All":
		return azure.AZMGIndustryDataReadBasicAll
	case "InformationProtectionConfig.Read":
		return azure.AZMGInformationProtectionConfigRead
	case "InformationProtectionConfig.Read.All":
		return azure.AZMGInformationProtectionConfigReadAll
	case "InformationProtectionContent.Sign.All":
		return azure.AZMGInformationProtectionContentSignAll
	case "InformationProtectionContent.Write.All":
		return azure.AZMGInformationProtectionContentWriteAll
	case "InformationProtectionPolicy.Read":
		return azure.AZMGInformationProtectionPolicyRead
	case "InformationProtectionPolicy.Read.All":
		return azure.AZMGInformationProtectionPolicyReadAll
	case "LearningAssignedCourse.Read":
		return azure.AZMGLearningAssignedCourseRead
	case "LearningAssignedCourse.Read.All":
		return azure.AZMGLearningAssignedCourseReadAll
	case "LearningAssignedCourse.ReadWrite.All":
		return azure.AZMGLearningAssignedCourseReadWriteAll
	case "LearningContent.Read.All":
		return azure.AZMGLearningContentReadAll
	case "LearningContent.ReadWrite.All":
		return azure.AZMGLearningContentReadWriteAll
	case "LearningProvider.Read":
		return azure.AZMGLearningProviderRead
	case "LearningProvider.ReadWrite":
		return azure.AZMGLearningProviderReadWrite
	case "LearningSelfInitiatedCourse.Read":
		return azure.AZMGLearningSelfInitiatedCourseRead
	case "LearningSelfInitiatedCourse.Read.All":
		return azure.AZMGLearningSelfInitiatedCourseReadAll
	case "LearningSelfInitiatedCourse.ReadWrite.All":
		return azure.AZMGLearningSelfInitiatedCourseReadWriteAll
	case "LicenseAssignment.Read.All":
		return azure.AZMGLicenseAssignmentReadAll
	case "LicenseAssignment.ReadWrite.All":
		return azure.AZMGLicenseAssignmentReadWriteAll
	case "LifecycleWorkflows.Read.All":
		return azure.AZMGLifecycleWorkflowsReadAll
	case "LifecycleWorkflows.ReadWrite.All":
		return azure.AZMGLifecycleWorkflowsReadWriteAll
	case "ListItems.SelectedOperations.Selected":
		return azure.AZMGListItemsSelectedOperationsSelected
	case "Lists.SelectedOperations.Selected":
		return azure.AZMGListsSelectedOperationsSelected
	case "Mail.Read":
		return azure.AZMGMailRead
	case "Mail.Read.Shared":
		return azure.AZMGMailReadShared
	case "Mail.ReadBasic":
		return azure.AZMGMailReadBasic
	case "Mail.ReadBasic.All":
		return azure.AZMGMailReadBasicAll
	case "Mail.ReadBasic.Shared":
		return azure.AZMGMailReadBasicShared
	case "Mail.ReadWrite":
		return azure.AZMGMailReadWrite
	case "Mail.ReadWrite.Shared":
		return azure.AZMGMailReadWriteShared
	case "Mail.Send":
		return azure.AZMGMailSend
	case "Mail.Send.Shared":
		return azure.AZMGMailSendShared
	case "MailboxFolder.Read":
		return azure.AZMGMailboxFolderRead
	case "MailboxFolder.Read.All":
		return azure.AZMGMailboxFolderReadAll
	case "MailboxFolder.ReadWrite":
		return azure.AZMGMailboxFolderReadWrite
	case "MailboxFolder.ReadWrite.All":
		return azure.AZMGMailboxFolderReadWriteAll
	case "MailboxItem.ImportExport":
		return azure.AZMGMailboxItemImportExport
	case "MailboxItem.ImportExport.All":
		return azure.AZMGMailboxItemImportExportAll
	case "MailboxItem.Read":
		return azure.AZMGMailboxItemRead
	case "MailboxItem.Read.All":
		return azure.AZMGMailboxItemReadAll
	case "MailboxSettings.Read":
		return azure.AZMGMailboxSettingsRead
	case "MailboxSettings.ReadWrite":
		return azure.AZMGMailboxSettingsReadWrite
	case "ManagedTenants.Read.All":
		return azure.AZMGManagedTenantsReadAll
	case "ManagedTenants.ReadWrite.All":
		return azure.AZMGManagedTenantsReadWriteAll
	case "Member.Read.Hidden":
		return azure.AZMGMemberReadHidden
	case "MLModel.Execute.All":
		return azure.AZMGMLModelExecuteAll
	case "MultiTenantOrganization.Read.All":
		return azure.AZMGMultiTenantOrganizationReadAll
	case "MultiTenantOrganization.ReadBasic.All":
		return azure.AZMGMultiTenantOrganizationReadBasicAll
	case "MultiTenantOrganization.ReadWrite.All":
		return azure.AZMGMultiTenantOrganizationReadWriteAll
	case "MutualTlsOauthConfiguration.Read.All":
		return azure.AZMGMutualTlsOauthConfigurationReadAll
	case "MutualTlsOauthConfiguration.ReadWrite.All":
		return azure.AZMGMutualTlsOauthConfigurationReadWriteAll
	case "MyFiles.Read":
		return azure.AZMGMyFilesRead
	case "NetworkAccess.Read.All":
		return azure.AZMGNetworkAccessReadAll
	case "NetworkAccess.ReadWrite.All":
		return azure.AZMGNetworkAccessReadWriteAll
	case "NetworkAccessBranch.Read.All":
		return azure.AZMGNetworkAccessBranchReadAll
	case "NetworkAccessBranch.ReadWrite.All":
		return azure.AZMGNetworkAccessBranchReadWriteAll
	case "NetworkAccessPolicy.Read.All":
		return azure.AZMGNetworkAccessPolicyReadAll
	case "NetworkAccessPolicy.ReadWrite.All":
		return azure.AZMGNetworkAccessPolicyReadWriteAll
	case "Notes.Create":
		return azure.AZMGNotesCreate
	case "Notes.Read":
		return azure.AZMGNotesRead
	case "Notes.Read.All":
		return azure.AZMGNotesReadAll
	case "Notes.ReadWrite":
		return azure.AZMGNotesReadWrite
	case "Notes.ReadWrite.All":
		return azure.AZMGNotesReadWriteAll
	case "Notes.ReadWrite.CreatedByApp":
		return azure.AZMGNotesReadWriteCreatedByApp
	case "Notifications.ReadWrite.CreatedByApp":
		return azure.AZMGNotificationsReadWriteCreatedByApp
	case "OnPremDirectorySynchronization.Read.All":
		return azure.AZMGOnPremDirectorySynchronizationReadAll
	case "OnPremDirectorySynchronization.ReadWrite.All":
		return azure.AZMGOnPremDirectorySynchronizationReadWriteAll
	case "OnPremisesPublishingProfiles.ReadWrite.All":
		return azure.AZMGOnPremisesPublishingProfilesReadWriteAll
	case "OnlineMeetingAiInsight.Read.All":
		return azure.AZMGOnlineMeetingAiInsightReadAll
	case "OnlineMeetingAiInsight.Read.Chat":
		return azure.AZMGOnlineMeetingAiInsightReadChat
	case "OnlineMeetingArtifact.Read.All":
		return azure.AZMGOnlineMeetingArtifactReadAll
	case "OnlineMeetingRecording.Read.All":
		return azure.AZMGOnlineMeetingRecordingReadAll
	case "OnlineMeetingTranscript.Read.All":
		return azure.AZMGOnlineMeetingTranscriptReadAll
	case "OnlineMeetings.Read":
		return azure.AZMGOnlineMeetingsRead
	case "OnlineMeetings.Read.All":
		return azure.AZMGOnlineMeetingsReadAll
	case "OnlineMeetings.ReadWrite":
		return azure.AZMGOnlineMeetingsReadWrite
	case "OnlineMeetings.ReadWrite.All":
		return azure.AZMGOnlineMeetingsReadWriteAll
	case "OrgContact.Read.All":
		return azure.AZMGOrgContactReadAll
	case "Organization.Read.All":
		return azure.AZMGOrganizationReadAll
	case "Organization.ReadWrite.All":
		return azure.AZMGOrganizationReadWriteAll
	case "OrganizationalBranding.Read.All":
		return azure.AZMGOrganizationalBrandingReadAll
	case "OrganizationalBranding.ReadWrite.All":
		return azure.AZMGOrganizationalBrandingReadWriteAll
	case "POP.AccessAsUser.All":
		return azure.AZMGPOPAccessAsUserAll
	case "PartnerBilling.Read.All":
		return azure.AZMGPartnerBillingReadAll
	case "PartnerSecurity.Read.All":
		return azure.AZMGPartnerSecurityReadAll
	case "PartnerSecurity.ReadWrite.All":
		return azure.AZMGPartnerSecurityReadWriteAll
	case "PendingExternalUserProfile.Read.All":
		return azure.AZMGPendingExternalUserProfileReadAll
	case "PendingExternalUserProfile.ReadWrite.All":
		return azure.AZMGPendingExternalUserProfileReadWriteAll
	case "People.Read":
		return azure.AZMGPeopleRead
	case "People.Read.All":
		return azure.AZMGPeopleReadAll
	case "PeopleSettings.Read.All":
		return azure.AZMGPeopleSettingsReadAll
	case "PeopleSettings.ReadWrite.All":
		return azure.AZMGPeopleSettingsReadWriteAll
	case "Place.Read.All":
		return azure.AZMGPlaceReadAll
	case "Place.ReadWrite.All":
		return azure.AZMGPlaceReadWriteAll
	case "PlaceDevice.Read.All":
		return azure.AZMGPlaceDeviceReadAll
	case "PlaceDevice.ReadWrite.All":
		return azure.AZMGPlaceDeviceReadWriteAll
	case "PlaceDeviceTelemetry.ReadWrite.All":
		return azure.AZMGPlaceDeviceTelemetryReadWriteAll
	case "Policy.Read.All":
		return azure.AZMGPolicyReadAll
	case "Policy.Read.AuthenticationMethod":
		return azure.AZMGPolicyReadAuthenticationMethod
	case "Policy.Read.ConditionalAccess":
		return azure.AZMGPolicyReadConditionalAccess
	case "Policy.Read.DeviceConfiguration":
		return azure.AZMGPolicyReadDeviceConfiguration
	case "Policy.Read.IdentityProtection":
		return azure.AZMGPolicyReadIdentityProtection
	case "Policy.Read.PermissionGrant":
		return azure.AZMGPolicyReadPermissionGrant
	case "Policy.ReadWrite.AccessReview":
		return azure.AZMGPolicyReadWriteAccessReview
	case "Policy.ReadWrite.ApplicationConfiguration":
		return azure.AZMGPolicyReadWriteApplicationConfiguration
	case "Policy.ReadWrite.AuthenticationFlows":
		return azure.AZMGPolicyReadWriteAuthenticationFlows
	case "Policy.ReadWrite.AuthenticationMethod":
		return azure.AZMGPolicyReadWriteAuthenticationMethod
	case "Policy.ReadWrite.Authorization":
		return azure.AZMGPolicyReadWriteAuthorization
	case "Policy.ReadWrite.ConditionalAccess":
		return azure.AZMGPolicyReadWriteConditionalAccess
	case "Policy.ReadWrite.ConsentRequest":
		return azure.AZMGPolicyReadWriteConsentRequest
	case "Policy.ReadWrite.CrossTenantAccess":
		return azure.AZMGPolicyReadWriteCrossTenantAccess
	case "Policy.ReadWrite.CrossTenantCapability":
		return azure.AZMGPolicyReadWriteCrossTenantCapability
	case "Policy.ReadWrite.DeviceConfiguration":
		return azure.AZMGPolicyReadWriteDeviceConfiguration
	case "Policy.ReadWrite.ExternalIdentities":
		return azure.AZMGPolicyReadWriteExternalIdentities
	case "Policy.ReadWrite.FeatureRollout":
		return azure.AZMGPolicyReadWriteFeatureRollout
	case "Policy.ReadWrite.FedTokenValidation":
		return azure.AZMGPolicyReadWriteFedTokenValidation
	case "Policy.ReadWrite.IdentityProtection":
		return azure.AZMGPolicyReadWriteIdentityProtection
	case "Policy.ReadWrite.MobilityManagement":
		return azure.AZMGPolicyReadWriteMobilityManagement
	case "Policy.ReadWrite.PermissionGrant":
		return azure.AZMGPolicyReadWritePermissionGrant
	case "Policy.ReadWrite.SecurityDefaults":
		return azure.AZMGPolicyReadWriteSecurityDefaults
	case "Policy.ReadWrite.TrustFramework":
		return azure.AZMGPolicyReadWriteTrustFramework
	case "Presence.Read":
		return azure.AZMGPresenceRead
	case "Presence.Read.All":
		return azure.AZMGPresenceReadAll
	case "Presence.ReadWrite":
		return azure.AZMGPresenceReadWrite
	case "Presence.ReadWrite.All":
		return azure.AZMGPresenceReadWriteAll
	case "PrintConnector.Read.All":
		return azure.AZMGPrintConnectorReadAll
	case "PrintConnector.ReadWrite.All":
		return azure.AZMGPrintConnectorReadWriteAll
	case "PrintJob.Create":
		return azure.AZMGPrintJobCreate
	case "PrintJob.Manage.All":
		return azure.AZMGPrintJobManageAll
	case "PrintJob.Read":
		return azure.AZMGPrintJobRead
	case "PrintJob.Read.All":
		return azure.AZMGPrintJobReadAll
	case "PrintJob.ReadBasic":
		return azure.AZMGPrintJobReadBasic
	case "PrintJob.ReadBasic.All":
		return azure.AZMGPrintJobReadBasicAll
	case "PrintJob.ReadWrite":
		return azure.AZMGPrintJobReadWrite
	case "PrintJob.ReadWrite.All":
		return azure.AZMGPrintJobReadWriteAll
	case "PrintJob.ReadWriteBasic":
		return azure.AZMGPrintJobReadWriteBasic
	case "PrintJob.ReadWriteBasic.All":
		return azure.AZMGPrintJobReadWriteBasicAll
	case "PrintSettings.Read.All":
		return azure.AZMGPrintSettingsReadAll
	case "PrintSettings.ReadWrite.All":
		return azure.AZMGPrintSettingsReadWriteAll
	case "PrintTaskDefinition.ReadWrite.All":
		return azure.AZMGPrintTaskDefinitionReadWriteAll
	case "Printer.Create":
		return azure.AZMGPrinterCreate
	case "Printer.FullControl.All":
		return azure.AZMGPrinterFullControlAll
	case "Printer.Read.All":
		return azure.AZMGPrinterReadAll
	case "Printer.ReadWrite.All":
		return azure.AZMGPrinterReadWriteAll
	case "PrinterShare.Read.All":
		return azure.AZMGPrinterShareReadAll
	case "PrinterShare.ReadBasic.All":
		return azure.AZMGPrinterShareReadBasicAll
	case "PrinterShare.ReadWrite.All":
		return azure.AZMGPrinterShareReadWriteAll
	case "PrivilegedAccess.Read.AzureAD":
		return azure.AZMGPrivilegedAccessReadAzureAD
	case "PrivilegedAccess.Read.AzureADGroup":
		return azure.AZMGPrivilegedAccessReadAzureADGroup
	case "PrivilegedAccess.Read.AzureResources":
		return azure.AZMGPrivilegedAccessReadAzureResources
	case "PrivilegedAccess.ReadWrite.AzureAD":
		return azure.AZMGPrivilegedAccessReadWriteAzureAD
	case "PrivilegedAccess.ReadWrite.AzureADGroup":
		return azure.AZMGPrivilegedAccessReadWriteAzureADGroup
	case "PrivilegedAccess.ReadWrite.AzureResources":
		return azure.AZMGPrivilegedAccessReadWriteAzureResources
	case "PrivilegedAssignmentSchedule.Read.AzureADGroup":
		return azure.AZMGPrivilegedAssignmentScheduleReadAzureADGroup
	case "PrivilegedAssignmentSchedule.ReadWrite.AzureADGroup":
		return azure.AZMGPrivilegedAssignmentScheduleReadWriteAzureADGroup
	case "PrivilegedAssignmentSchedule.Remove.AzureADGroup":
		return azure.AZMGPrivilegedAssignmentScheduleRemoveAzureADGroup
	case "PrivilegedEligibilitySchedule.Read.AzureADGroup":
		return azure.AZMGPrivilegedEligibilityScheduleReadAzureADGroup
	case "PrivilegedEligibilitySchedule.ReadWrite.AzureADGroup":
		return azure.AZMGPrivilegedEligibilityScheduleReadWriteAzureADGroup
	case "PrivilegedEligibilitySchedule.Remove.AzureADGroup":
		return azure.AZMGPrivilegedEligibilityScheduleRemoveAzureADGroup
	case "ProfilePhoto.Read.All":
		return azure.AZMGProfilePhotoReadAll
	case "ProfilePhoto.ReadWrite.All":
		return azure.AZMGProfilePhotoReadWriteAll
	case "ProgramControl.Read.All":
		return azure.AZMGProgramControlReadAll
	case "ProgramControl.ReadWrite.All":
		return azure.AZMGProgramControlReadWriteAll
	case "ProtectionScopes.Compute.All":
		return azure.AZMGProtectionScopesComputeAll
	case "ProtectionScopes.Compute.User":
		return azure.AZMGProtectionScopesComputeUser
	case "ProvisioningLog.Read.All":
		return azure.AZMGProvisioningLogReadAll
	case "PublicKeyInfrastructure.Read.All":
		return azure.AZMGPublicKeyInfrastructureReadAll
	case "PublicKeyInfrastructure.ReadWrite.All":
		return azure.AZMGPublicKeyInfrastructureReadWriteAll
	case "QnA.Read.All":
		return azure.AZMGQnAReadAll
	case "RecordsManagement.Read.All":
		return azure.AZMGRecordsManagementReadAll
	case "RecordsManagement.ReadWrite.All":
		return azure.AZMGRecordsManagementReadWriteAll
	case "ReportSettings.Read.All":
		return azure.AZMGReportSettingsReadAll
	case "ReportSettings.ReadWrite.All":
		return azure.AZMGReportSettingsReadWriteAll
	case "Reports.Read.All":
		return azure.AZMGReportsReadAll
	case "Report.Read.All":
		return azure.AZMGReportReadAll
	case "ResourceSpecificPermissionGrant.ReadForChat":
		return azure.AZMGResourceSpecificPermissionGrantReadForChat
	case "ResourceSpecificPermissionGrant.ReadForChat.All":
		return azure.AZMGResourceSpecificPermissionGrantReadForChatAll
	case "ResourceSpecificPermissionGrant.ReadForTeam":
		return azure.AZMGResourceSpecificPermissionGrantReadForTeam
	case "ResourceSpecificPermissionGrant.ReadForTeam.All":
		return azure.AZMGResourceSpecificPermissionGrantReadForTeamAll
	case "ResourceSpecificPermissionGrant.ReadForUser":
		return azure.AZMGResourceSpecificPermissionGrantReadForUser
	case "ResourceSpecificPermissionGrant.ReadForUser.All":
		return azure.AZMGResourceSpecificPermissionGrantReadForUserAll
	case "RiskPreventionProviders.Read.All":
		return azure.AZMGRiskPreventionProvidersReadAll
	case "RiskPreventionProviders.ReadWrite.All":
		return azure.AZMGRiskPreventionProvidersReadWriteAll
	case "RoleAssignmentSchedule.Read.Directory":
		return azure.AZMGRoleAssignmentScheduleReadDirectory
	case "RoleAssignmentSchedule.ReadWrite.Directory":
		return azure.AZMGRoleAssignmentScheduleReadWriteDirectory
	case "RoleAssignmentSchedule.Remove.Directory":
		return azure.AZMGRoleAssignmentScheduleRemoveDirectory
	case "RoleEligibilitySchedule.Read.Directory":
		return azure.AZMGRoleEligibilityScheduleReadDirectory
	case "RoleEligibilitySchedule.ReadWrite.Directory":
		return azure.AZMGRoleEligibilityScheduleReadWriteDirectory
	case "RoleEligibilitySchedule.Remove.Directory":
		return azure.AZMGRoleEligibilityScheduleRemoveDirectory
	case "RoleManagement.Read.All":
		return azure.AZMGRoleManagementReadAll
	case "RoleManagement.Read.CloudPC":
		return azure.AZMGRoleManagementReadCloudPC
	case "RoleManagement.Read.Defender":
		return azure.AZMGRoleManagementReadDefender
	case "RoleManagement.Read.Directory":
		return azure.AZMGRoleManagementReadDirectory
	case "RoleManagement.Read.Exchange":
		return azure.AZMGRoleManagementReadExchange
	case "RoleManagement.ReadWrite.CloudPC":
		return azure.AZMGRoleManagementReadWriteCloudPC
	case "RoleManagement.ReadWrite.Defender":
		return azure.AZMGRoleManagementReadWriteDefender
	case "RoleManagement.ReadWrite.Directory":
		return azure.AZMGRoleManagementReadWriteDirectory
	case "RoleManagement.ReadWrite.Exchange":
		return azure.AZMGRoleManagementReadWriteExchange
	case "RoleManagementAlert.Read.Directory":
		return azure.AZMGRoleManagementAlertReadDirectory
	case "RoleManagementAlert.ReadWrite.Directory":
		return azure.AZMGRoleManagementAlertReadWriteDirectory
	case "RoleManagementPolicy.Read.AzureADGroup":
		return azure.AZMGRoleManagementPolicyReadAzureADGroup
	case "RoleManagementPolicy.Read.Directory":
		return azure.AZMGRoleManagementPolicyReadDirectory
	case "RoleManagementPolicy.ReadWrite.AzureADGroup":
		return azure.AZMGRoleManagementPolicyReadWriteAzureADGroup
	case "RoleManagementPolicy.ReadWrite.Directory":
		return azure.AZMGRoleManagementPolicyReadWriteDirectory
	case "SMTP.Send":
		return azure.AZMGSMTPSend
	case "Schedule.Read.All":
		return azure.AZMGScheduleReadAll
	case "Schedule.ReadWrite.All":
		return azure.AZMGScheduleReadWriteAll
	case "SchedulePermissions.ReadWrite.All":
		return azure.AZMGSchedulePermissionsReadWriteAll
	case "SearchConfiguration.Read.All":
		return azure.AZMGSearchConfigurationReadAll
	case "SearchConfiguration.ReadWrite.All":
		return azure.AZMGSearchConfigurationReadWriteAll
	case "SecurityActions.Read.All":
		return azure.AZMGSecurityActionsReadAll
	case "SecurityActions.ReadWrite.All":
		return azure.AZMGSecurityActionsReadWriteAll
	case "SecurityAlert.Read.All":
		return azure.AZMGSecurityAlertReadAll
	case "SecurityAlert.ReadWrite.All":
		return azure.AZMGSecurityAlertReadWriteAll
	case "SecurityAnalyzedMessage.Read.All":
		return azure.AZMGSecurityAnalyzedMessageReadAll
	case "SecurityAnalyzedMessage.ReadWrite.All":
		return azure.AZMGSecurityAnalyzedMessageReadWriteAll
	case "SecurityCopilotWorkspaces.Read.All":
		return azure.AZMGSecurityCopilotWorkspacesReadAll
	case "SecurityCopilotWorkspaces.ReadWrite.All":
		return azure.AZMGSecurityCopilotWorkspacesReadWriteAll
	case "SecurityEvents.Read.All":
		return azure.AZMGSecurityEventsReadAll
	case "SecurityEvents.ReadWrite.All":
		return azure.AZMGSecurityEventsReadWriteAll
	case "SecurityIdentitiesAccount.Read.All":
		return azure.AZMGSecurityIdentitiesAccountReadAll
	case "SecurityIdentitiesActions.ReadWrite.All":
		return azure.AZMGSecurityIdentitiesActionsReadWriteAll
	case "SecurityIdentitiesHealth.Read.All":
		return azure.AZMGSecurityIdentitiesHealthReadAll
	case "SecurityIdentitiesHealth.ReadWrite.All":
		return azure.AZMGSecurityIdentitiesHealthReadWriteAll
	case "SecurityIdentitiesSensors.Read.All":
		return azure.AZMGSecurityIdentitiesSensorsReadAll
	case "SecurityIdentitiesSensors.ReadWrite.All":
		return azure.AZMGSecurityIdentitiesSensorsReadWriteAll
	case "SecurityIdentitiesUserActions.Read.All":
		return azure.AZMGSecurityIdentitiesUserActionsReadAll
	case "SecurityIdentitiesUserActions.ReadWrite.All":
		return azure.AZMGSecurityIdentitiesUserActionsReadWriteAll
	case "SecurityIncident.Read.All":
		return azure.AZMGSecurityIncidentReadAll
	case "SecurityIncident.ReadWrite.All":
		return azure.AZMGSecurityIncidentReadWriteAll
	case "SensitivityLabel.Evaluate":
		return azure.AZMGSensitivityLabelEvaluate
	case "SensitivityLabel.Evaluate.All":
		return azure.AZMGSensitivityLabelEvaluateAll
	case "SensitivityLabel.Read":
		return azure.AZMGSensitivityLabelRead
	case "SensitivityLabels.Read.All":
		return azure.AZMGSensitivityLabelsReadAll
	case "ServiceHealth.Read.All":
		return azure.AZMGServiceHealthReadAll
	case "ServiceMessage.Read.All":
		return azure.AZMGServiceMessageReadAll
	case "ServiceMessageViewpoint.Write":
		return azure.AZMGServiceMessageViewpointWrite
	case "ServicePrincipalEndpoint.Read.All":
		return azure.AZMGServicePrincipalEndpointReadAll
	case "ServicePrincipalEndpoint.ReadWrite.All":
		return azure.AZMGServicePrincipalEndpointReadWriteAll
	case "SharePointTenantSettings.Read.All":
		return azure.AZMGSharePointTenantSettingsReadAll
	case "SharePointTenantSettings.ReadWrite.All":
		return azure.AZMGSharePointTenantSettingsReadWriteAll
	case "ShortNotes.Read":
		return azure.AZMGShortNotesRead
	case "ShortNotes.Read.All":
		return azure.AZMGShortNotesReadAll
	case "ShortNotes.ReadWrite":
		return azure.AZMGShortNotesReadWrite
	case "ShortNotes.ReadWrite.All":
		return azure.AZMGShortNotesReadWriteAll
	case "SignInIdentifier.Read.All":
		return azure.AZMGSignInIdentifierReadAll
	case "SignInIdentifier.ReadWrite.All":
		return azure.AZMGSignInIdentifierReadWriteAll
	case "Sites.Archive.All":
		return azure.AZMGSitesArchiveAll
	case "Sites.FullControl.All":
		return azure.AZMGSitesFullControlAll
	case "Sites.Manage.All":
		return azure.AZMGSitesManageAll
	case "Sites.Read.All":
		return azure.AZMGSitesReadAll
	case "Sites.ReadWrite.All":
		return azure.AZMGSitesReadWriteAll
	case "Sites.Selected":
		return azure.AZMGSitesSelected
	case "SpiffeTrustDomain.Read.All":
		return azure.AZMGSpiffeTrustDomainReadAll
	case "SpiffeTrustDomain.ReadWrite.All":
		return azure.AZMGSpiffeTrustDomainReadWriteAll
	case "Storyline.ReadWrite.All":
		return azure.AZMGStorylineReadWriteAll
	case "SubjectRightsRequest.Read.All":
		return azure.AZMGSubjectRightsRequestReadAll
	case "SubjectRightsRequest.ReadWrite.All":
		return azure.AZMGSubjectRightsRequestReadWriteAll
	case "Subscription.Read.All":
		return azure.AZMGSubscriptionReadAll
	case "Synchronization.Read.All":
		return azure.AZMGSynchronizationReadAll
	case "Synchronization.ReadWrite.All":
		return azure.AZMGSynchronizationReadWriteAll
	case "Tasks.Read":
		return azure.AZMGTasksRead
	case "Tasks.Read.All":
		return azure.AZMGTasksReadAll
	case "Tasks.Read.Shared":
		return azure.AZMGTasksReadShared
	case "Tasks.ReadWrite":
		return azure.AZMGTasksReadWrite
	case "Tasks.ReadWrite.All":
		return azure.AZMGTasksReadWriteAll
	case "Tasks.ReadWrite.Shared":
		return azure.AZMGTasksReadWriteShared
	case "Team.Create":
		return azure.AZMGTeamCreate
	case "Team.ReadBasic.All":
		return azure.AZMGTeamReadBasicAll
	case "TeamMember.Read.All":
		return azure.AZMGTeamMemberReadAll
	case "TeamMember.ReadWrite.All":
		return azure.AZMGTeamMemberReadWriteAll
	case "TeamMember.ReadWriteNonOwnerRole.All":
		return azure.AZMGTeamMemberReadWriteNonOwnerRoleAll
	case "TeamSettings.Read.All":
		return azure.AZMGTeamSettingsReadAll
	case "TeamSettings.ReadWrite.All":
		return azure.AZMGTeamSettingsReadWriteAll
	case "TeamTemplates.Read":
		return azure.AZMGTeamTemplatesRead
	case "TeamTemplates.Read.All":
		return azure.AZMGTeamTemplatesReadAll
	case "TeamsActivity.Read":
		return azure.AZMGTeamsActivityRead
	case "TeamsActivity.Read.All":
		return azure.AZMGTeamsActivityReadAll
	case "TeamsActivity.Send":
		return azure.AZMGTeamsActivitySend
	case "TeamsAppInstallation.ManageSelectedForChat":
		return azure.AZMGTeamsAppInstallationManageSelectedForChat
	case "TeamsAppInstallation.ManageSelectedForChat.All":
		return azure.AZMGTeamsAppInstallationManageSelectedForChatAll
	case "TeamsAppInstallation.ManageSelectedForTeam":
		return azure.AZMGTeamsAppInstallationManageSelectedForTeam
	case "TeamsAppInstallation.ManageSelectedForTeam.All":
		return azure.AZMGTeamsAppInstallationManageSelectedForTeamAll
	case "TeamsAppInstallation.ManageSelectedForUser":
		return azure.AZMGTeamsAppInstallationManageSelectedForUser
	case "TeamsAppInstallation.ManageSelectedForUser.All":
		return azure.AZMGTeamsAppInstallationManageSelectedForUserAll
	case "TeamsAppInstallation.Read.All":
		return azure.AZMGTeamsAppInstallationReadAll
	case "TeamsAppInstallation.ReadForChat":
		return azure.AZMGTeamsAppInstallationReadForChat
	case "TeamsAppInstallation.ReadForChat.All":
		return azure.AZMGTeamsAppInstallationReadForChatAll
	case "TeamsAppInstallation.ReadForTeam":
		return azure.AZMGTeamsAppInstallationReadForTeam
	case "TeamsAppInstallation.ReadForTeam.All":
		return azure.AZMGTeamsAppInstallationReadForTeamAll
	case "TeamsAppInstallation.ReadForUser":
		return azure.AZMGTeamsAppInstallationReadForUser
	case "TeamsAppInstallation.ReadForUser.All":
		return azure.AZMGTeamsAppInstallationReadForUserAll
	case "TeamsAppInstallation.ReadSelectedForChat":
		return azure.AZMGTeamsAppInstallationReadSelectedForChat
	case "TeamsAppInstallation.ReadSelectedForChat.All":
		return azure.AZMGTeamsAppInstallationReadSelectedForChatAll
	case "TeamsAppInstallation.ReadSelectedForTeam":
		return azure.AZMGTeamsAppInstallationReadSelectedForTeam
	case "TeamsAppInstallation.ReadSelectedForTeam.All":
		return azure.AZMGTeamsAppInstallationReadSelectedForTeamAll
	case "TeamsAppInstallation.ReadSelectedForUser":
		return azure.AZMGTeamsAppInstallationReadSelectedForUser
	case "TeamsAppInstallation.ReadSelectedForUser.All":
		return azure.AZMGTeamsAppInstallationReadSelectedForUserAll
	case "TeamsAppInstallation.ReadWriteAndConsentForChat":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentForChat
	case "TeamsAppInstallation.ReadWriteAndConsentForChat.All":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentForChatAll
	case "TeamsAppInstallation.ReadWriteAndConsentForTeam":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentForTeam
	case "TeamsAppInstallation.ReadWriteAndConsentForTeam.All":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentForTeamAll
	case "TeamsAppInstallation.ReadWriteAndConsentForUser":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentForUser
	case "TeamsAppInstallation.ReadWriteAndConsentForUser.All":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentForUserAll
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForChat":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentSelfForChat
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForChat.All":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentSelfForChatAll
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForTeam":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentSelfForTeam
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForTeam.All":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentSelfForTeamAll
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForUser":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentSelfForUser
	case "TeamsAppInstallation.ReadWriteAndConsentSelfForUser.All":
		return azure.AZMGTeamsAppInstallationReadWriteAndConsentSelfForUserAll
	case "TeamsAppInstallation.ReadWriteForChat":
		return azure.AZMGTeamsAppInstallationReadWriteForChat
	case "TeamsAppInstallation.ReadWriteForChat.All":
		return azure.AZMGTeamsAppInstallationReadWriteForChatAll
	case "TeamsAppInstallation.ReadWriteForTeam":
		return azure.AZMGTeamsAppInstallationReadWriteForTeam
	case "TeamsAppInstallation.ReadWriteForTeam.All":
		return azure.AZMGTeamsAppInstallationReadWriteForTeamAll
	case "TeamsAppInstallation.ReadWriteForUser":
		return azure.AZMGTeamsAppInstallationReadWriteForUser
	case "TeamsAppInstallation.ReadWriteForUser.All":
		return azure.AZMGTeamsAppInstallationReadWriteForUserAll
	case "TeamsAppInstallation.ReadWriteSelectedForChat":
		return azure.AZMGTeamsAppInstallationReadWriteSelectedForChat
	case "TeamsAppInstallation.ReadWriteSelectedForChat.All":
		return azure.AZMGTeamsAppInstallationReadWriteSelectedForChatAll
	case "TeamsAppInstallation.ReadWriteSelectedForTeam":
		return azure.AZMGTeamsAppInstallationReadWriteSelectedForTeam
	case "TeamsAppInstallation.ReadWriteSelectedForTeam.All":
		return azure.AZMGTeamsAppInstallationReadWriteSelectedForTeamAll
	case "TeamsAppInstallation.ReadWriteSelectedForUser":
		return azure.AZMGTeamsAppInstallationReadWriteSelectedForUser
	case "TeamsAppInstallation.ReadWriteSelectedForUser.All":
		return azure.AZMGTeamsAppInstallationReadWriteSelectedForUserAll
	case "TeamsAppInstallation.ReadWriteSelfForChat":
		return azure.AZMGTeamsAppInstallationReadWriteSelfForChat
	case "TeamsAppInstallation.ReadWriteSelfForChat.All":
		return azure.AZMGTeamsAppInstallationReadWriteSelfForChatAll
	case "TeamsAppInstallation.ReadWriteSelfForTeam":
		return azure.AZMGTeamsAppInstallationReadWriteSelfForTeam
	case "TeamsAppInstallation.ReadWriteSelfForTeam.All":
		return azure.AZMGTeamsAppInstallationReadWriteSelfForTeamAll
	case "TeamsAppInstallation.ReadWriteSelfForUser":
		return azure.AZMGTeamsAppInstallationReadWriteSelfForUser
	case "TeamsAppInstallation.ReadWriteSelfForUser.All":
		return azure.AZMGTeamsAppInstallationReadWriteSelfForUserAll
	case "TeamsPolicyUserAssign.ReadWrite.All":
		return azure.AZMGTeamsPolicyUserAssignReadWriteAll
	case "TeamsResourceAccount.Read.All":
		return azure.AZMGTeamsResourceAccountReadAll
	case "TeamsTab.Create":
		return azure.AZMGTeamsTabCreate
	case "TeamsTab.Read.All":
		return azure.AZMGTeamsTabReadAll
	case "TeamsTab.ReadWrite.All":
		return azure.AZMGTeamsTabReadWriteAll
	case "TeamsTab.ReadWriteForChat":
		return azure.AZMGTeamsTabReadWriteForChat
	case "TeamsTab.ReadWriteForChat.All":
		return azure.AZMGTeamsTabReadWriteForChatAll
	case "TeamsTab.ReadWriteForTeam":
		return azure.AZMGTeamsTabReadWriteForTeam
	case "TeamsTab.ReadWriteForTeam.All":
		return azure.AZMGTeamsTabReadWriteForTeamAll
	case "TeamsTab.ReadWriteForUser":
		return azure.AZMGTeamsTabReadWriteForUser
	case "TeamsTab.ReadWriteForUser.All":
		return azure.AZMGTeamsTabReadWriteForUserAll
	case "TeamsTab.ReadWriteSelfForChat":
		return azure.AZMGTeamsTabReadWriteSelfForChat
	case "TeamsTab.ReadWriteSelfForChat.All":
		return azure.AZMGTeamsTabReadWriteSelfForChatAll
	case "TeamsTab.ReadWriteSelfForTeam":
		return azure.AZMGTeamsTabReadWriteSelfForTeam
	case "TeamsTab.ReadWriteSelfForTeam.All":
		return azure.AZMGTeamsTabReadWriteSelfForTeamAll
	case "TeamsTab.ReadWriteSelfForUser":
		return azure.AZMGTeamsTabReadWriteSelfForUser
	case "TeamsTab.ReadWriteSelfForUser.All":
		return azure.AZMGTeamsTabReadWriteSelfForUserAll
	case "TeamsTelephoneNumber.Read.All":
		return azure.AZMGTeamsTelephoneNumberReadAll
	case "TeamsTelephoneNumber.ReadWrite.All":
		return azure.AZMGTeamsTelephoneNumberReadWriteAll
	case "TeamsUserConfiguration.Read.All":
		return azure.AZMGTeamsUserConfigurationReadAll
	case "Teamwork.Migrate.All":
		return azure.AZMGTeamworkMigrateAll
	case "Teamwork.Read.All":
		return azure.AZMGTeamworkReadAll
	case "TeamworkAppSettings.Read.All":
		return azure.AZMGTeamworkAppSettingsReadAll
	case "TeamworkAppSettings.ReadWrite.All":
		return azure.AZMGTeamworkAppSettingsReadWriteAll
	case "TeamworkDevice.Read.All":
		return azure.AZMGTeamworkDeviceReadAll
	case "TeamworkDevice.ReadWrite.All":
		return azure.AZMGTeamworkDeviceReadWriteAll
	case "TeamworkTag.Read":
		return azure.AZMGTeamworkTagRead
	case "TeamworkTag.Read.All":
		return azure.AZMGTeamworkTagReadAll
	case "TeamworkTag.ReadWrite":
		return azure.AZMGTeamworkTagReadWrite
	case "TeamworkTag.ReadWrite.All":
		return azure.AZMGTeamworkTagReadWriteAll
	case "TeamworkUserInteraction.Read.All":
		return azure.AZMGTeamworkUserInteractionReadAll
	case "TermStore.Read.All":
		return azure.AZMGTermStoreReadAll
	case "TermStore.ReadWrite.All":
		return azure.AZMGTermStoreReadWriteAll
	case "ThreatAssessment.Read.All":
		return azure.AZMGThreatAssessmentReadAll
	case "ThreatAssessment.ReadWrite.All":
		return azure.AZMGThreatAssessmentReadWriteAll
	case "ThreatHunting.Read.All":
		return azure.AZMGThreatHuntingReadAll
	case "ThreatIndicators.Read.All":
		return azure.AZMGThreatIndicatorsReadAll
	case "ThreatIndicators.ReadWrite.OwnedBy":
		return azure.AZMGThreatIndicatorsReadWriteOwnedBy
	case "ThreatIntelligence.Read.All":
		return azure.AZMGThreatIntelligenceReadAll
	case "ThreatSubmission.Read":
		return azure.AZMGThreatSubmissionRead
	case "ThreatSubmission.Read.All":
		return azure.AZMGThreatSubmissionReadAll
	case "ThreatSubmission.ReadWrite":
		return azure.AZMGThreatSubmissionReadWrite
	case "ThreatSubmission.ReadWrite.All":
		return azure.AZMGThreatSubmissionReadWriteAll
	case "ThreatSubmissionPolicy.ReadWrite.All":
		return azure.AZMGThreatSubmissionPolicyReadWriteAll
	case "Topic.Read.All":
		return azure.AZMGTopicReadAll
	case "TrustFrameworkKeySet.Read.All":
		return azure.AZMGTrustFrameworkKeySetReadAll
	case "TrustFrameworkKeySet.ReadWrite.All":
		return azure.AZMGTrustFrameworkKeySetReadWriteAll
	case "UnifiedGroupMember.Read.AsGuest":
		return azure.AZMGUnifiedGroupMemberReadAsGuest
	case "User.DeleteRestore.All":
		return azure.AZMGUserDeleteRestoreAll
	case "User.EnableDisableAccount.All":
		return azure.AZMGUserEnableDisableAccountAll
	case "User.Export.All":
		return azure.AZMGUserExportAll
	case "User.Invite.All":
		return azure.AZMGUserInviteAll
	case "User.ManageIdentities.All":
		return azure.AZMGUserManageIdentitiesAll
	case "User.Read":
		return azure.AZMGUserRead
	case "User.Read.All":
		return azure.AZMGUserReadAll
	case "User.ReadBasic.All":
		return azure.AZMGUserReadBasicAll
	case "User.ReadWrite":
		return azure.AZMGUserReadWrite
	case "User.ReadWrite.All":
		return azure.AZMGUserReadWriteAll
	case "User.ReadWrite.CrossCloud":
		return azure.AZMGUserReadWriteCrossCloud
	case "User.RevokeSessions.All":
		return azure.AZMGUserRevokeSessionsAll
	case "UserActivity.ReadWrite.CreatedByApp":
		return azure.AZMGUserActivityReadWriteCreatedByApp
	case "UserAuthenticationMethod.Read":
		return azure.AZMGUserAuthenticationMethodRead
	case "UserAuthenticationMethod.Read.All":
		return azure.AZMGUserAuthenticationMethodReadAll
	case "UserAuthenticationMethod.ReadWrite":
		return azure.AZMGUserAuthenticationMethodReadWrite
	case "UserAuthenticationMethod.ReadWrite.All":
		return azure.AZMGUserAuthenticationMethodReadWriteAll
	case "UserCloudClipboard.Read":
		return azure.AZMGUserCloudClipboardRead
	case "UserNotification.ReadWrite.CreatedByApp":
		return azure.AZMGUserNotificationReadWriteCreatedByApp
	case "UserState.ReadWrite.All":
		return azure.AZMGUserStateReadWriteAll
	case "UserShiftPreferences.Read.All":
		return azure.AZMGUserShiftPreferencesReadAll
	case "UserShiftPreferences.ReadWrite.All":
		return azure.AZMGUserShiftPreferencesReadWriteAll
	case "UserTeamwork.Read":
		return azure.AZMGUserTeamworkRead
	case "UserTeamwork.Read.All":
		return azure.AZMGUserTeamworkReadAll
	case "UserTimelineActivity.Write.CreatedByApp":
		return azure.AZMGUserTimelineActivityWriteCreatedByApp
	case "UserWindowsSettings.Read.All":
		return azure.AZMGUserWindowsSettingsReadAll
	case "UserWindowsSettings.ReadWrite.All":
		return azure.AZMGUserWindowsSettingsReadWriteAll
	case "VirtualAppointment.Read":
		return azure.AZMGVirtualAppointmentRead
	case "VirtualAppointment.Read.All":
		return azure.AZMGVirtualAppointmentReadAll
	case "VirtualAppointment.ReadWrite":
		return azure.AZMGVirtualAppointmentReadWrite
	case "VirtualAppointment.ReadWrite.All":
		return azure.AZMGVirtualAppointmentReadWriteAll
	case "VirtualAppointmentNotification.Send":
		return azure.AZMGVirtualAppointmentNotificationSend
	case "VirtualEvent.Read":
		return azure.AZMGVirtualEventRead
	case "VirtualEvent.Read.All":
		return azure.AZMGVirtualEventReadAll
	case "VirtualEvent.ReadWrite":
		return azure.AZMGVirtualEventReadWrite
	case "WindowsUpdates.ReadWrite.All":
		return azure.AZMGWindowsUpdatesReadWriteAll
	case "WorkforceIntegration.Read.All":
		return azure.AZMGWorkforceIntegrationReadAll
	case "WorkforceIntegration.ReadWrite.All":
		return azure.AZMGWorkforceIntegrationReadWriteAll
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
        s := strings.TrimSpace(scope)
        if s != "" && s != "null" && s != "openid" && s != "profile" && s != "email" && s != "offline_access" {
            scopeSet[s] = struct{}{}
        }
    }

	
	// If the ConsentType is "Principal", we create a relationship from the Principal to the Tenant
	// If the ConsentType is "AllPrincipals", we create a relationship from the Tenant to the Tenant
	
	// Modellierung:
    // - consentType == "Principal": User (principalId) --> ServicePrincipal (clientId)  [RelType = Scope]
    // - consentType == "AllPrincipals": Tenant (tenantId) --> ServicePrincipal (clientId) [RelType = Scope]
    //   (resourceId wird als RelProp mitgegeben, damit das Ziel-API/SP ersichtlich ist)

    if data.ConsentType == "Principal" && data.PrincipalId != "" && data.ClientId != "" {

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
					RelProps: map[string]any{
						"consentType": data.ConsentType,
						"resourceId":  strings.ToUpper(data.ResourceId),
					},
					RelType:  GetPermissionConstant(scope),
				},
			))
		}
	} else if data.ConsentType == "AllPrincipals" && data.TenantId != "" && data.ClientId != "" {
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
					RelProps: map[string]any{
                        "consentType": data.ConsentType,
                        "resourceId":  strings.ToUpper(data.ResourceId),
                    },
					RelType:  GetPermissionConstant(scope),
				},
			))
		}

	}
	return relationships
}
