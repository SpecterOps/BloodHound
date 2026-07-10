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
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

const (
	ISO8601               string = "2006-01-02T15:04:05Z"
	KeyVaultPermissionGet string = "Get"
	AzureSerialError      string = "Error deserializing Azure data"
	AzureExtractError     string = "Failed to extract id/type from Azure directory object"
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
		ObjectID: app.AppId,
		Labels:   []graph.Kind{azure.App},
	}
}

func ConvertAZAppRelationships(app models.App) []IngestibleRelationship {
	return []IngestibleRelationship{
		NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: app.TenantId,
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.App,
				Value: app.AppId,
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
		ObjectID: device.Id,
		Labels:   []graph.Kind{azure.Device},
	}
}

func ConvertAZDeviceRelationships(device models.Device) []IngestibleRelationship {
	return []IngestibleRelationship{
		NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: device.TenantId,
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.Device,
				Value: device.Id,
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
		ObjectID: scaleSet.Id,
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
			Value: scaleSet.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.VMScaleSet,
			Value: scaleSet.Id,
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
				Value: scaleSet.Id,
				Kind:  azure.VMScaleSet,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: scaleSet.Identity.PrincipalId,
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
					Value: scaleSet.Id,
					Kind:  azure.VMScaleSet,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: identity.PrincipalId,
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
						Value: raw.Assignee.GetPrincipalId(),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.VMScaleSet,
						Value: data.ObjectId,
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
			Value: directoryObject.Id,
			Kind:  ownerType,
		},
		IngestibleEndpoint{
			Kind:  targetType,
			Value: targetId,
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Owns,
		},
	)
}

func ConvertAppFederatedIdentityCredential(federatedIdentityCredential models.FICData, appID, tenantName, tenantId string) (IngestibleNode, IngestibleRelationship) {
	node := IngestibleNode{
		ObjectID: federatedIdentityCredential.ID,
		PropertyMap: map[string]any{
			common.Description.String(): federatedIdentityCredential.Description,
			common.Name.String():        strings.ToUpper(fmt.Sprintf("%s@%s", federatedIdentityCredential.Name, tenantName)),
			azure.Issuer.String():       federatedIdentityCredential.Issuer,
			azure.Audiences.String():    federatedIdentityCredential.Audiences,
			azure.Subject.String():      federatedIdentityCredential.Subject,
			azure.TenantID.String():     strings.ToUpper(tenantId),
		},
		Labels: []graph.Kind{azure.FederatedIdentityCredential, azure.Entity},
	}

	rel := NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: federatedIdentityCredential.ID,
			Kind:  azure.FederatedIdentityCredential,
		},
		IngestibleEndpoint{
			Kind:  azure.App,
			Value: appID,
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.AZAuthenticatesTo,
		},
	)

	return node, rel
}

func ConvertAzureAppRoleAssignmentToNodes(data models.AppRoleAssignment) []IngestibleNode {
	nodes := make([]IngestibleNode, 0)

	nodes = append(nodes, IngestibleNode{
		PropertyMap: map[string]any{
			common.DisplayName.String(): strings.ToUpper(data.PrincipalDisplayName),
			azure.TenantID.String():     strings.ToUpper(data.TenantId),
		},
		ObjectID: data.PrincipalId.String(),
		Labels:   []graph.Kind{azure.ServicePrincipal},
	})

	nodes = append(nodes, IngestibleNode{
		PropertyMap: map[string]any{
			common.DisplayName.String(): strings.ToUpper(data.ResourceDisplayName),
			azure.TenantID.String():     strings.ToUpper(data.TenantId),
		},
		ObjectID: data.ResourceId,
		Labels:   []graph.Kind{azure.ServicePrincipal},
	})

	return nodes
}

func ConvertAzureAppRoleAssignmentToRel(data models.AppRoleAssignment) IngestibleRelationship {
	if appRoleKind, hasAppRoleKind := azure.RelationshipKindByAppRoleID[strings.ToLower(data.AppRoleId.String())]; hasAppRoleKind {
		return NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: data.PrincipalId.String(),
				Kind:  azure.ServicePrincipal,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: data.ResourceId,
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
		ObjectID: data.Id,
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
			Value: data.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.FunctionApp,
			Value: data.Id,
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
				Value: data.Id,
				Kind:  azure.FunctionApp,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: data.Identity.PrincipalId,
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
					Value: data.Id,
					Kind:  azure.FunctionApp,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: identity.PrincipalId,
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
						Value: raw.Assignee.GetPrincipalId(),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.FunctionApp,
						Value: data.ObjectId,
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
		ObjectID: data.Id,
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

func ConvertAzureGroupToRel(data models.Group) IngestibleRelationship {
	return NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: data.TenantId,
			Kind:  azure.Tenant,
		},
		IngestibleEndpoint{
			Kind:  azure.Group,
			Value: data.Id,
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
			slog.Error(
				AzureSerialError,
				slog.String("type", "group member"),
				attr.Error(err),
			)
		} else if memberType, err := ExtractTypeFromDirectoryObject(member); errors.Is(err, ErrInvalidType) {
			slog.Warn(
				AzureExtractError,
				attr.Error(err),
			)
		} else if err != nil {
			slog.Error(
				AzureExtractError,
				attr.Error(err),
			)
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: member.Id,
					Kind:  memberType,
				},
				IngestibleEndpoint{
					Kind:  azure.Group,
					Value: data.GroupId,
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
			slog.Error(
				AzureSerialError,
				slog.String("type", "group owner"),
				attr.Error(err),
			)
		} else if ownerType, err := ExtractTypeFromDirectoryObject(owner); errors.Is(err, ErrInvalidType) {
			slog.Warn(
				AzureExtractError,
				attr.Error(err),
			)
		} else if err != nil {
			slog.Error(
				AzureExtractError,
				attr.Error(err),
			)
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: owner.Id,
					Kind:  ownerType,
				},
				IngestibleEndpoint{
					Kind:  azure.Group,
					Value: data.GroupId,
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
			ObjectID: data.Id,
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
				Value: data.ResourceGroup,
				Kind:  azure.ResourceGroup,
			},
			IngestibleEndpoint{
				Kind:  azure.KeyVault,
				Value: data.Id,
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
				Value: data.ObjectId,
				Kind:  azure.Entity,
			},
			IngestibleEndpoint{
				Kind:  azure.KeyVault,
				Value: data.KeyVaultId,
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
					Value: raw.Contributor.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: data.KeyVaultId,
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
					Value: raw.KVContributor.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: data.KeyVaultId,
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
					Value: raw.Owner.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: data.KeyVaultId,
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
					Value: raw.UserAccessAdmin.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.KeyVault,
					Value: data.KeyVaultId,
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
			Value: data.Properties.Parent.Id,
			Kind:  azure.ManagementGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.Entity,
			Value: data.Id,
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
					Value: raw.Owner.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ManagementGroup,
					Value: data.ManagementGroupId,
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

func ConvertAzureManagementGroupContributorToRels(data models.ManagementGroupContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Contributors {
		if data.ManagementGroupId == raw.Contributor.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: raw.Contributor.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ManagementGroup,
					Value: data.ManagementGroupId,
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

func ConvertAzureManagementGroupUserAccessAdminToRels(data models.ManagementGroupUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.UserAccessAdmins {
		if data.ManagementGroupId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: raw.UserAccessAdmin.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ManagementGroup,
					Value: data.ManagementGroupId,
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
			ObjectID: data.Id,
			PropertyMap: map[string]any{
				azure.TenantID.String():       strings.ToUpper(data.TenantId),
				common.LastCollected.String(): ingestTime,
				common.DisplayName.String():   strings.ToUpper(data.Properties.DisplayName),
			},
			Labels: []graph.Kind{azure.ManagementGroup},
		}, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: data.TenantId,
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.ManagementGroup,
				Value: data.Id,
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
}

func ConvertAzureResourceGroup(data models.ResourceGroup, ingestTime time.Time) (IngestibleNode, IngestibleRelationship) {
	return IngestibleNode{
			ObjectID: data.Id,
			PropertyMap: map[string]any{
				common.Name.String():          strings.ToUpper(data.Name),
				azure.TenantID.String():       strings.ToUpper(data.TenantId),
				common.LastCollected.String(): ingestTime,
			},
			Labels: []graph.Kind{azure.ResourceGroup},
		}, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: data.SubscriptionId,
				Kind:  azure.Subscription,
			},
			IngestibleEndpoint{
				Kind:  azure.ResourceGroup,
				Value: data.Id,
			},
			IngestibleRel{
				RelProps: map[string]any{},
				RelType:  azure.Contains,
			},
		)
}

func ConvertAzureResourceGroupContributorToRels(data models.ResourceGroupContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Contributors {
		if data.ResourceGroupId == raw.Contributor.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: raw.Contributor.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ResourceGroup,
					Value: data.ResourceGroupId,
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

func ConvertAzureResourceGroupOwnerToRels(data models.ResourceGroupOwners) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Owners {
		if data.ResourceGroupId == raw.Owner.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: raw.Owner.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ResourceGroup,
					Value: data.ResourceGroupId,
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
					Value: raw.UserAccessAdmin.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.ResourceGroup,
					Value: data.ResourceGroupId,
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
	roleObjectId := fmt.Sprintf("%s@%s", data.Id, data.TenantId)
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
				Value: data.TenantId,
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
		scope = data.TenantId
	} else {
		scope = roleAssignment.DirectoryScopeId[1:]
	}

	if CanAddSecret(roleAssignment.RoleDefinitionId) && roleAssignment.DirectoryScopeId != "/" {
		if relType, err := GetAddSecretRoleKind(roleAssignment.RoleDefinitionId); err != nil {
			slog.Error(
				"Error processing role assignment for role",
				slog.String("role_object_id", roleObjectId),
				attr.Error(err),
			)
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: roleAssignment.PrincipalId,
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
				Value: roleAssignment.PrincipalId,
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
		ObjectID: data.Id,
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
		ObjectID: data.AppId,
		PropertyMap: map[string]any{
			common.DisplayName.String(): data.AppDisplayName,
			azure.TenantID.String():     strings.ToUpper(data.AppOwnerOrganizationId),
		},
		Labels: []graph.Kind{azure.App},
	})

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: data.AppId,
			Kind:  azure.App,
		},
		IngestibleEndpoint{
			Kind:  azure.ServicePrincipal,
			Value: data.Id,
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.RunsAs,
		},
	))

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: data.TenantId,
			Kind:  azure.Tenant,
		},
		IngestibleEndpoint{
			Kind:  azure.ServicePrincipal,
			Value: data.Id,
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
		ObjectID: logicApp.Id,
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
			Value: logicApp.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.LogicApp,
			Value: logicApp.Id,
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
				Value: logicApp.Id,
				Kind:  azure.LogicApp,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: logicApp.Identity.PrincipalId,
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
					Value: logicApp.Id,
					Kind:  azure.LogicApp,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: identity.PrincipalId,
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
						Value: raw.Assignee.GetPrincipalId(),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.LogicApp,
						Value: roleAssignment.ObjectId,
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
			slog.Error(
				AzureSerialError,
				slog.String("type", "service principal owner"),
				attr.Error(err),
			)
		} else if ownerType, err := ExtractTypeFromDirectoryObject(owner); errors.Is(err, ErrInvalidType) {
			slog.Warn(
				AzureExtractError,
				attr.Error(err),
			)
		} else if err != nil {
			slog.Error(
				AzureExtractError,
				attr.Error(err),
			)
		} else {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: owner.Id,
					Kind:  ownerType,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: data.ServicePrincipalId,
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
			ObjectID: data.Id,
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
				Value: data.TenantId,
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.Subscription,
				Value: data.Id,
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
					Value: raw.Owner.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.Subscription,
					Value: data.SubscriptionId,
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

func ConvertAzureSubscriptionContributorToRels(data models.SubscriptionContributors) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, raw := range data.Contributors {
		if data.SubscriptionId == raw.Contributor.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: raw.Contributor.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.Subscription,
					Value: data.SubscriptionId,
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

func ConvertAzureSubscriptionUserAccessAdminToRels(data models.SubscriptionUserAccessAdmins) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, raw := range data.UserAccessAdmins {
		if data.SubscriptionId == raw.UserAccessAdmin.Properties.Scope {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: raw.UserAccessAdmin.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.Subscription,
					Value: data.SubscriptionId,
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
		ObjectID: data.TenantId,
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

// ConvertAzureUser returns the basic node and the ingestible contains relationship.
func ConvertAzureUser(data models.User, ingestTime time.Time) (IngestibleNode, IngestibleRelationship) {
	properties := map[string]any{
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
	}

	if data.SignInActivity.LastSuccessfulSignInDateTime != "" {
		properties[azure.LastSuccessfulSignInDateTime.String()] = data.SignInActivity.LastSuccessfulSignInDateTime
	}

	return IngestibleNode{
			ObjectID: data.Id,
			PropertyMap: properties,
			Labels:      []graph.Kind{azure.User},
		}, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: data.TenantId,
				Kind:  azure.Tenant,
			},
			IngestibleEndpoint{
				Kind:  azure.User,
				Value: data.Id,
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
		ObjectID: data.Id,
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
			Value: data.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.VM,
			Value: data.Id,
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
				Value: data.Id,
				Kind:  azure.VM,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: data.Identity.PrincipalId,
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
					Value: data.Id,
					Kind:  azure.VM,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: identity.PrincipalId,
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
					Value: raw.AdminLogin.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: data.VirtualMachineId,
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
					Value: raw.AvereContributor.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: data.VirtualMachineId,
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
					Value: raw.Contributor.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: data.VirtualMachineId,
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
					Value: raw.VMContributor.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: data.VirtualMachineId,
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
					Value: raw.Owner.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: data.VirtualMachineId,
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
					Value: raw.UserAccessAdmin.Properties.PrincipalId,
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Kind:  azure.VM,
					Value: data.VirtualMachineId,
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
		ObjectID: data.Id,
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
			Value: data.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.ManagedCluster,
			Value: data.Id,
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	))

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: data.Id,
			Kind:  azure.ManagedCluster,
		},
		IngestibleEndpoint{
			Kind:  azure.ResourceGroup,
			Value: nodeResourceGroupID,
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
						Value: raw.Assignee.GetPrincipalId(),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.ManagedCluster,
						Value: data.ObjectId,
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
		ObjectID: data.Id,
		PropertyMap: map[string]any{
			common.Name.String():          strings.ToUpper(data.Name),
			azure.TenantID.String():       strings.ToUpper(data.TenantId),
			common.LastCollected.String(): ingestTime,
		},
		Labels: []graph.Kind{azure.ContainerRegistry},
	}

	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: data.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.ContainerRegistry,
			Value: data.Id,
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
				Value: data.Id,
				Kind:  azure.ContainerRegistry,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: data.Identity.PrincipalId,
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
					Value: data.Id,
					Kind:  azure.ContainerRegistry,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: identity.PrincipalId,
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
		ObjectID: webApp.Id,
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
			Value: webApp.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.WebApp,
			Value: webApp.Id,
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
				Value: webApp.Id,
				Kind:  azure.WebApp,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: webApp.Identity.PrincipalId,
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
					Value: webApp.Id,
					Kind:  azure.WebApp,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: identity.PrincipalId,
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
						Value: raw.Assignee.GetPrincipalId(),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.AutomationAccount,
						Value: roleAssignments.ObjectId,
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
						Value: raw.Assignee.GetPrincipalId(),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.ContainerRegistry,
						Value: roleAssignment.ObjectId,
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
						Value: raw.Assignee.GetPrincipalId(),
						Kind:  azure.Entity,
					},
					IngestibleEndpoint{
						Kind:  azure.WebApp,
						Value: roleAssignment.ObjectId,
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
		ObjectID: account.Id,
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
			Value: account.ResourceGroupId,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Kind:  azure.AutomationAccount,
			Value: account.Id,
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
				Value: account.Id,
				Kind:  azure.AutomationAccount,
			},
			IngestibleEndpoint{
				Kind:  azure.ServicePrincipal,
				Value: account.Identity.PrincipalId,
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
					Value: account.Id,
					Kind:  azure.AutomationAccount,
				},
				IngestibleEndpoint{
					Kind:  azure.ServicePrincipal,
					Value: identity.PrincipalId,
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
	id := fmt.Sprintf("%s@%s", instance.RoleDefinitionId, instance.TenantId)

	relationships := make([]IngestibleRelationship, 0)
	//If the scope is not the directory, we are going to skip creating the edges for now until later work is done
	//This isn't necessarily the best spot for this, but it works and it makes testing simple, since none of our azure convertors are exported
	if instance.DirectoryScopeId != "/" {
		return relationships
	}
	relationships = append(relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: instance.PrincipalId,
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
//
// When preserveObjectIdentifierCasing is false, EndUserAssignment user/group approver identifiers are
// emitted uppercased (legacy behavior). When true, raw casing is preserved. In either case, incoming
// approver slices are never mutated in place; a defensive copy is produced.
func ConvertAzureRoleManagementPolicyAssignment(policyAssignment models.RoleManagementPolicyAssignment, preserveObjectIdentifierCasing bool) (IngestibleNode, []IngestibleRelationship) {
	var (
		rels             = make([]IngestibleRelationship, 0)
		combinedObjectId = fmt.Sprintf("%s@%s", policyAssignment.RoleDefinitionId, policyAssignment.TenantId)
		userApprovers    = normalizeApproverIdentifiers(policyAssignment.EndUserAssignmentUserApprovers, preserveObjectIdentifierCasing)
		groupApprovers   = normalizeApproverIdentifiers(policyAssignment.EndUserAssignmentGroupApprovers, preserveObjectIdentifierCasing)
	)

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
			azure.EndUserAssignmentUserApprovers.String():                    userApprovers,
			azure.EndUserAssignmentGroupApprovers.String():                   groupApprovers,
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
	if len(userApprovers) > 0 {
		// Create an AZRoleApprover edge from each user that allow approvals to the target azure role
		for _, approver := range userApprovers {
			rels = append(rels, NewIngestibleRelationship(IngestibleEndpoint{
				Value: approver,
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

	if len(groupApprovers) > 0 {
		// Create an AZRoleApprover edge from each group that allow approvals to the target azure role
		for _, approver := range groupApprovers {
			rels = append(rels, NewIngestibleRelationship(IngestibleEndpoint{
				Value: approver,
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

	if len(userApprovers) == 0 && len(groupApprovers) == 0 {
		// No users or groups were attached to the policy, we will create the edge from the tenant's PrivilegedRoleAdministratorRole Role node to the target role
		combinedObjectId := fmt.Sprintf("%s@%s", azure.PrivilegedRoleAdministratorRole, policyAssignment.TenantId)

		rels = append(rels, NewIngestibleRelationship(IngestibleEndpoint{
			Value: combinedObjectId,
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

// normalizeApproverIdentifiers returns a copy of the provided approver identifier slice, uppercased
// unless preserveObjectIdentifierCasing is true. It never returns the caller-owned slice so that
// downstream property/endpoint assignments cannot mutate ingest input.
func normalizeApproverIdentifiers(identifiers []string, preserveObjectIdentifierCasing bool) []string {
	if len(identifiers) == 0 {
		return nil
	}
	normalized := make([]string, len(identifiers))
	if preserveObjectIdentifierCasing {
		copy(normalized, identifiers)
		return normalized
	}
	for i, identifier := range identifiers {
		normalized[i] = strings.ToUpper(identifier)
	}
	return normalized
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
