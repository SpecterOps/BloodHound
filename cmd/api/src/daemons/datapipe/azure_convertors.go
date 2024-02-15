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

package datapipe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bloodhoundad/azurehound/v2/enums"
	"github.com/bloodhoundad/azurehound/v2/models"
	azureModels "github.com/bloodhoundad/azurehound/v2/models/azure"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/log"
)

const (
	SerialError                   = "error deserializing %s: %v"
	ExtractError                  = "failed to extract owner id/type from directory object: %v"
	PrincipalTypeServicePrincipal = "ServicePrincipal"
	PrincipalTypeUser             = "User"
)

func getKindConverter(kind enums.Kind) func(json.RawMessage, *ConvertedAzureData) {
	switch kind {
	case enums.KindAZApp:
		return convertAzureApp
	case enums.KindAZAppOwner:
		return convertAzureAppOwner
	case enums.KindAZAppRoleAssignment:
		return convertAzureAppRoleAssignment
	case enums.KindAZDevice:
		return convertAzureDevice
	case enums.KindAZDeviceOwner:
		return convertAzureDeviceOwner
	case enums.KindAZFunctionApp:
		return convertAzureFunctionApp
	case enums.KindAZFunctionAppRoleAssignment:
		return convertAzureFunctionAppRoleAssignment
	case enums.KindAZGroup:
		return convertAzureGroup
	case enums.KindAZGroupMember:
		return convertAzureGroupMember
	case enums.KindAZGroupOwner:
		return convertAzureGroupOwner
	case enums.KindAZKeyVault:
		return convertAzureKeyVault
	case enums.KindAZKeyVaultAccessPolicy:
		return convertAzureKeyVaultAccessPolicy
	case enums.KindAZKeyVaultOwner:
		return convertAzureKeyVaultOwner
	case enums.KindAZKeyVaultUserAccessAdmin:
		return convertAzureKeyVaultUserAccessAdmin
	case enums.KindAZKeyVaultContributor:
		return convertAzureKeyVaultContributor
	case enums.KindAZKeyVaultKVContributor:
		return convertAzureKeyVaultKVContributor
	case enums.KindAZManagementGroup:
		return convertAzureManagementGroup
	case enums.KindAZManagementGroupOwner:
		return convertAzureManagementGroupOwner
	case enums.KindAZManagementGroupUserAccessAdmin:
		return convertAzureManagementGroupUserAccessAdmin
	case enums.KindAZManagementGroupDescendant:
		return convertAzureManagementGroupDescendant
	case enums.KindAZResourceGroup:
		return convertAzureResourceGroup
	case enums.KindAZResourceGroupOwner:
		return convertAzureResourceGroupOwner
	case enums.KindAZResourceGroupUserAccessAdmin:
		return convertAzureResourceGroupUserAccessAdmin
	case enums.KindAZRole:
		return convertAzureRole
	case enums.KindAZRoleAssignment:
		return convertAzureRoleAssignment
	case enums.KindAZServicePrincipal:
		return convertAzureServicePrincipal
	case enums.KindAZServicePrincipalOwner:
		return convertAzureServicePrincipalOwner
	case enums.KindAZSubscription:
		return convertAzureSubscription
	case enums.KindAZSubscriptionOwner:
		return convertAzureSubscriptionOwner
	case enums.KindAZSubscriptionUserAccessAdmin:
		return convertAzureSubscriptionUserAccessAdmin
	case enums.KindAZTenant:
		return convertAzureTenant
	case enums.KindAZUser:
		return convertAzureUser
	case enums.KindAZVM:
		return convertAzureVirtualMachine
	case enums.KindAZVMAdminLogin:
		return convertAzureVirtualMachineAdminLogin
	case enums.KindAZVMAvereContributor:
		return convertAzureVirtualMachineAvereContributor
	case enums.KindAZVMContributor:
		return convertAzureVirtualMachineContributor
	case enums.KindAZVMOwner:
		return convertAzureVirtualMachineOwner
	case enums.KindAZVMUserAccessAdmin:
		return convertAzureVirtualMachineUserAccessAdmin
	case enums.KindAZVMVMContributor:
		return convertAzureVirtualMachineVMContributor
	case enums.KindAZManagedCluster:
		return convertAzureManagedCluster
	case enums.KindAZManagedClusterRoleAssignment:
		return convertAzureManagedClusterRoleAssignment
	case enums.KindAZVMScaleSet:
		return convertAzureVMScaleSet
	case enums.KindAZVMScaleSetRoleAssignment:
		return convertAzureVMScaleSetRoleAssignment
	case enums.KindAZContainerRegistry:
		return convertAzureContainerRegistry
	case enums.KindAZContainerRegistryRoleAssignment:
		return convertAzureContainerRegistryRoleAssignment
	case enums.KindAZWebApp:
		return convertAzureWebApp
	case enums.KindAZWebAppRoleAssignment:
		return convertAzureWebAppRoleAssignment
	case enums.KindAZLogicApp:
		return convertAzureLogicApp
	case enums.KindAZLogicAppRoleAssignment:
		return convertAzureLogicAppRoleAssignment
	case enums.KindAZAutomationAccount:
		return convertAzureAutomationAccount
	case enums.KindAZAutomationAccountRoleAssignment:
		return convertAzureAutomationAccountRoleAssignment
	default:
		// TODO: we should probably have a hook or something to log the unknown type
		return func(rm json.RawMessage, cd *ConvertedAzureData) {}
	}
}

func convertAzureApp(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.App
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf("error deserializing azure application: %v", err)
	} else {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertAZAppToNode(data))
		converted.RelProps = append(converted.RelProps, ein.ConvertAZAppRelationships(data)...)
	}
}

func convertAzureVMScaleSet(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VMScaleSet
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine scale set", err)
	} else {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertAZVMScaleSetToNode(data))
		converted.RelProps = append(converted.RelProps, ein.ConvertAZVMScaleSetRelationships(data)...)
	}
}

func convertAzureVMScaleSetRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AzureRoleAssignments

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine scale set role assignments", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureVMScaleSetRoleAssignment(data)...)
	}
}

func convertAzureAppOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.AppOwners
	)

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "app owner", err)
	} else {
		for _, raw := range data.Owners {
			var (
				owner azureModels.DirectoryObject
			)
			if err := json.Unmarshal(raw.Owner, &owner); err != nil {
				log.Errorf(SerialError, "app owner", err)
			} else if ownerType, err := ein.ExtractTypeFromDirectoryObject(owner); err != nil {
				log.Errorf(ExtractError, err)
			} else {
				converted.RelProps = append(converted.RelProps, ein.ConvertAzureOwnerToRel(owner, ownerType, azure.App, data.AppId))
			}
		}
	}
}

func convertAzureAppRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AppRoleAssignment

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "app role assignment", err)
	} else if data.AppId == azure.MSGraphAppUniversalID && data.PrincipalType == PrincipalTypeServicePrincipal {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertAzureAppRoleAssignmentToNodes(data)...)
		if rel := ein.ConvertAzureAppRoleAssignmentToRel(data); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}
	}
}

func convertAzureDevice(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.Device
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure device", err)
	} else {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertAZDeviceToNode(data))
		converted.RelProps = append(converted.RelProps, ein.ConvertAZDeviceRelationships(data)...)
	}
}

func convertAzureDeviceOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.DeviceOwners
	)
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "device owners", err)
	} else {
		for _, raw := range data.Owners {
			var (
				owner azureModels.DirectoryObject
			)
			if err := json.Unmarshal(raw.Owner, &owner); err != nil {
				log.Errorf(SerialError, "device owner", err)
			} else if ownerType, err := ein.ExtractTypeFromDirectoryObject(owner); err != nil {
				log.Errorf(ExtractError, err)
			} else {
				converted.RelProps = append(converted.RelProps, ein.ConvertAzureOwnerToRel(owner, ownerType, azure.Device, data.DeviceId))
			}
		}
	}
}

func convertAzureFunctionApp(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.FunctionApp
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure function app", err)
	} else {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertAzureFunctionAppToNode(data))
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureFunctionAppToRels(data)...)
	}
}

func convertAzureFunctionAppRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AzureRoleAssignments

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure function app role assignments", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureFunctionAppRoleAssignmentToRels(data)...)
	}
}

func convertAzureGroup(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.Group
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure group", err)
	} else {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertAzureGroupToNode(data))
		if onPremNode := ein.ConvertAzureGroupToOnPremisesNode(data); onPremNode.IsValid() {
			converted.OnPremNodes = append(converted.OnPremNodes, onPremNode)
		}
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureGroupToRel(data))
	}
}

func convertAzureGroupMember(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.GroupMembers
	)

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure group members", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureGroupMembersToRels(data)...)
	}
}

func convertAzureGroupOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.GroupOwners
	)
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure group owners", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureGroupOwnerToRels(data)...)
	}
}

func convertAzureKeyVault(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.KeyVault
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure keyvault", err)
	} else {
		node, rel := ein.ConvertAzureKeyVault(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertAzureKeyVaultAccessPolicy(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.KeyVaultAccessPolicy
	)

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure key vault access policy", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureKeyVaultAccessPolicy(data)...)
	}
}

func convertAzureKeyVaultContributor(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.KeyVaultContributors
	)

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure keyvault contributor", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureKeyVaultContributor(data)...)
	}
}

func convertAzureKeyVaultKVContributor(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.KeyVaultKVContributors
	)

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure keyvault kvcontributor", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureKeyVaultKVContributor(data)...)
	}
}

func convertAzureKeyVaultOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.KeyVaultOwners
	)

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure keyvault owner", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureKeyVaultOwnerToRels(data)...)
	}
}

func convertAzureKeyVaultUserAccessAdmin(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.KeyVaultUserAccessAdmins
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure keyvault user access admin", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureKeyVaultUserAccessAdminToRels(data)...)
	}
}

func convertAzureManagementGroupDescendant(raw json.RawMessage, converted *ConvertedAzureData) {
	var data azureModels.DescendantInfo
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure management group descendant list", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureManagementGroupDescendantToRel(data))
	}
}

func convertAzureManagementGroupOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ManagementGroupOwners
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure management group owner", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureManagementGroupOwnerToRels(data)...)
	}
}

func convertAzureManagementGroupUserAccessAdmin(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ManagementGroupUserAccessAdmins
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure management group user access admin", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureManagementGroupUserAccessAdminToRels(data)...)
	}
}

func convertAzureManagementGroup(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ManagementGroup
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure management group", err)
	} else {
		node, rel := ein.ConvertAzureManagementGroup(data)
		converted.RelProps = append(converted.RelProps, rel)
		converted.NodeProps = append(converted.NodeProps, node)
	}
}

func convertAzureResourceGroup(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ResourceGroup
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure resource group", err)
	} else {
		node, rel := ein.ConvertAzureResourceGroup(data)
		converted.RelProps = append(converted.RelProps, rel)
		converted.NodeProps = append(converted.NodeProps, node)
	}
}

func convertAzureResourceGroupOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ResourceGroupOwners
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure keyvault", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureResourceGroupOwnerToRels(data)...)
	}
}

func convertAzureResourceGroupUserAccessAdmin(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ResourceGroupUserAccessAdmins
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure resource group user access admin", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureResourceGroupUserAccessAdminToRels(data)...)
	}
}

func convertAzureRole(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.Role
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure role", err)
	} else {
		node, rel := ein.ConvertAzureRole(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertAzureRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.RoleAssignments
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure role assignment", err)
	} else {
		for _, raw := range data.RoleAssignments {
			var (
				roleObjectId = fmt.Sprintf("%s@%s", strings.ToUpper(raw.RoleDefinitionId), strings.ToUpper(data.TenantId))
			)

			converted.RelProps = append(converted.RelProps, ein.ConvertAzureRoleAssignmentToRels(raw, data, roleObjectId)...)
		}
	}
}

func convertAzureServicePrincipal(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ServicePrincipal
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure service principal owner", err)
	} else {
		nodes, rels := ein.ConvertAzureServicePrincipal(data)
		converted.NodeProps = append(converted.NodeProps, nodes...)
		converted.RelProps = append(converted.RelProps, rels...)
	}
}

func convertAzureServicePrincipalOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var (
		data models.ServicePrincipalOwners
	)
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure service principal owners", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureServicePrincipalOwnerToRels(data)...)
	}
}

func convertAzureSubscription(raw json.RawMessage, converted *ConvertedAzureData) {
	var data azureModels.Subscription
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure subscription", err)
	} else {
		node, rel := ein.ConvertAzureSubscription(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertAzureSubscriptionOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.SubscriptionOwners
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure subscription owner", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureSubscriptionOwnerToRels(data)...)
	}
}

func convertAzureSubscriptionUserAccessAdmin(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.SubscriptionUserAccessAdmins
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure subscription user access admin", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureSubscriptionUserAccessAdminToRels(data)...)
	}
}

func convertAzureTenant(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.Tenant
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure tenant", err)
	} else {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertAzureTenantToNode(data))
	}
}

func convertAzureUser(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.User
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure user", err)
	} else {
		node, onPremNode, rel := ein.ConvertAzureUser(data)
		converted.NodeProps = append(converted.NodeProps, node)
		if onPremNode.IsValid() {
			converted.OnPremNodes = append(converted.OnPremNodes, onPremNode)
		}
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertAzureVirtualMachine(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VirtualMachine
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine", err)
	} else {
		node, rels := ein.ConvertAzureVirtualMachine(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, rels...)
	}
}

func convertAzureVirtualMachineAdminLogin(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VirtualMachineAdminLogins
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine admin login", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureVirtualMachineAdminLoginToRels(data)...)
	}
}

func convertAzureVirtualMachineAvereContributor(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VirtualMachineAvereContributors
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine avere contributor", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureVirtualMachineAvereContributorToRels(data)...)
	}
}

func convertAzureVirtualMachineContributor(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VirtualMachineContributors
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine contributor", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureVirtualMachineContributorToRels(data)...)
	}
}

func convertAzureVirtualMachineVMContributor(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VirtualMachineVMContributors
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine contributor", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureVirtualMachineVMContributorToRels(data)...)
	}
}

func convertAzureVirtualMachineOwner(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VirtualMachineOwners
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine owner", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureVirtualMachineOwnerToRels(data)...)
	}
}

func convertAzureVirtualMachineUserAccessAdmin(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.VirtualMachineUserAccessAdmins
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure virtual machine user access admin", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureVirtualMachineUserAccessAdminToRels(data)...)
	}
}

func convertAzureManagedCluster(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ManagedCluster
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure managed cluster", err)
	} else {
		NodeResourceGroupID := fmt.Sprintf("/subscriptions/%s/resourcegroups/%s", data.SubscriptionId, data.Properties.NodeResourceGroup)

		node, rels := ein.ConvertAzureManagedCluster(data, NodeResourceGroupID)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, rels...)
	}
}

func convertAzureManagedClusterRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AzureRoleAssignments

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure managed cluster role assignments", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureManagedClusterRoleAssignmentToRels(data)...)
	}
}

func convertAzureContainerRegistry(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.ContainerRegistry
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure container registry", err)
	} else {
		node, rels := ein.ConvertAzureContainerRegistry(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, rels...)
	}
}

func convertAzureWebApp(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.WebApp
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure web app", err)
	} else {
		node, relationships := ein.ConvertAzureWebApp(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, relationships...)
	}
}

func convertAzureContainerRegistryRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AzureRoleAssignments

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure container registry role assignments", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureContainerRegistryRoleAssignment(data)...)
	}
}

func convertAzureWebAppRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AzureRoleAssignments

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure web app role assignments", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureWebAppRoleAssignment(data)...)
	}
}

func convertAzureLogicApp(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.LogicApp
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure logic app", err)
	} else {
		node, relationships := ein.ConvertAzureLogicApp(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, relationships...)
	}
}

func convertAzureLogicAppRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AzureRoleAssignments

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure logic app role assignments", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureLogicAppRoleAssignment(data)...)
	}
}

func convertAzureAutomationAccount(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AutomationAccount
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure automation account", err)
	} else {
		node, relationships := ein.ConvertAzureAutomationAccount(data)
		converted.NodeProps = append(converted.NodeProps, node)
		converted.RelProps = append(converted.RelProps, relationships...)
	}
}

func convertAzureAutomationAccountRoleAssignment(raw json.RawMessage, converted *ConvertedAzureData) {
	var data models.AzureRoleAssignments

	if err := json.Unmarshal(raw, &data); err != nil {
		log.Errorf(SerialError, "azure automation account role assignments", err)
	} else {
		converted.RelProps = append(converted.RelProps, ein.ConvertAzureAutomationAccountRoleAssignment(data)...)
	}
}
