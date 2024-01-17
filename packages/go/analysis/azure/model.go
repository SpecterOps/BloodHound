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

package azure

import (
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
)

// RelatedEntityType is a type for differentiating which related entities a user wants to query for. Technically all
// queries are available on each endpoint. We may want to register routes all separately however that requires a great
// deal of additional boilerplate registration work.
type RelatedEntityType string

const (
	RelatedEntityTypeActiveAssignments                  RelatedEntityType = "active-assignments"
	RelatedEntityTypePIMAssignments                     RelatedEntityType = "pim-assignments"
	RelatedEntityTypeGroupMembership                    RelatedEntityType = "group-membership"
	RelatedEntityTypeVaultAllReaders                    RelatedEntityType = "all-readers"
	RelatedEntityTypeVaultCertReaders                   RelatedEntityType = "certificate-readers"
	RelatedEntityTypeVaultSecretReaders                 RelatedEntityType = "secret-readers"
	RelatedEntityTypeVaultKeyReaders                    RelatedEntityType = "key-readers"
	RelatedEntityTypeGroupMembers                       RelatedEntityType = "group-members"
	RelatedEntityTypeRoles                              RelatedEntityType = "roles"
	RelatedEntityTypeFunctionApps                       RelatedEntityType = "function-apps"
	RelatedEntityTypeInboundExecutionPrivileges         RelatedEntityType = "inbound-execution-privileges"
	RelatedEntityTypeOutboundExecutionPrivileges        RelatedEntityType = "outbound-execution-privileges"
	RelatedEntityTypeOutboundAbusableAppRoleAssignments RelatedEntityType = "outbound-abusable-app-role-assignments"
	RelatedEntityTypeInboundAbusableAppRoleAssignments  RelatedEntityType = "inbound-abusable-app-role-assignments"
	RelatedEntityTypeOutboundControl                    RelatedEntityType = "outbound-control"
	RelatedEntityTypeInboundControl                     RelatedEntityType = "inbound-control"
	RelatedEntityTypeDescendentUsers                    RelatedEntityType = "descendent-users"
	RelatedEntityTypeDescendentGroups                   RelatedEntityType = "descendent-groups"
	RelatedEntityTypeDescendentManagementGroups         RelatedEntityType = "descendent-management-groups"
	RelatedEntityTypeDescendentSubscriptions            RelatedEntityType = "descendent-subscriptions"
	RelatedEntityTypeDescendentResourceGroups           RelatedEntityType = "descendent-resource-groups"
	RelatedEntityTypeDescendentVirtualMachines          RelatedEntityType = "descendent-virtual-machines"
	RelatedEntityTypeDescendentManagedClusters          RelatedEntityType = "descendent-managed-clusters"
	RelatedEntityTypeDescendentWebApps                  RelatedEntityType = "descendent-web-apps"
	RelatedEntityTypeDescendentLogicApps                RelatedEntityType = "descendent-logic-apps"
	RelatedEntityTypeDescendentAutomationAccounts       RelatedEntityType = "descendent-automation-accounts"
	RelatedEntityTypeDescendentKeyVaults                RelatedEntityType = "descendent-key-vaults"
	RelatedEntityTypeDescendentApplications             RelatedEntityType = "descendent-applications"
	RelatedEntityTypeDescendentServicePrincipals        RelatedEntityType = "descendent-service-principals"
	RelatedEntityTypeDescendentDevices                  RelatedEntityType = "descendent-devices"
	RelatedEntityTypeDescendentVMScaleSets              RelatedEntityType = "descendent-vm-scale-sets"
	RelatedEntityTypeDescendentContainerRegistries      RelatedEntityType = "descendent-container-registries"
	RelatedEntityTypeDescendentFunctionApps             RelatedEntityType = "descendent-function-apps"
)

// FromGraphNodes takes a slice of *graph.Node and converts them to serializable node structs.
func FromGraphNodes(nodes []*graph.Node) []Node {
	renderedNodes := make([]Node, len(nodes))

	for idx, node := range nodes {
		renderedNodes[idx] = Node{
			Kind:       analysis.GetNodeKindDisplayLabel(node),
			Properties: node.Properties.Map,
		}
	}

	return renderedNodes
}

// FromGraphNode takes a *graph.Node and converts it to serializable node struct.
func FromGraphNode(node *graph.Node) Node {
	return Node{
		Kind:       analysis.GetNodeKindDisplayLabel(node),
		Properties: node.Properties.Map,
	}
}

// Node is a serializable version of *graph.Node
type Node struct {
	Kind       string         `json:"kind"`
	Properties map[string]any `json:"props"`
}

type BaseDetails struct {
	Node

	OutboundObjectControl int `json:"outbound_object_control"`
}

type UserDetails struct {
	Node

	GroupMembership       int `json:"group_membership"`
	Roles                 int `json:"roles"`
	ExecutionPrivileges   int `json:"execution_privileges"`
	OutboundObjectControl int `json:"outbound_object_control"`
	InboundObjectControl  int `json:"inbound_object_control"`
}

type GroupDetails struct {
	Node

	Roles                 int `json:"roles"`
	GroupMembers          int `json:"group_members"`
	GroupMembership       int `json:"group_membership"`
	OutboundObjectControl int `json:"outbound_object_control"`
	InboundObjectControl  int `json:"inbound_object_control"`
}

type TenantDetails struct {
	Node

	Descendents          Descendents `json:"descendents"`
	InboundObjectControl int         `json:"inbound_object_control"`
}

type ManagementGroupDetails struct {
	Node

	Descendents          Descendents `json:"descendents"`
	InboundObjectControl int         `json:"inbound_object_control"`
}

type SubscriptionDetails struct {
	Node

	Descendents          Descendents `json:"descendents"`
	InboundObjectControl int         `json:"inbound_object_control"`
}

type Descendents struct {
	DescendentCounts map[string]int `json:"descendent_counts"`
}

type ResourceGroupDetails struct {
	Node

	Descendents          Descendents `json:"descendents"`
	InboundObjectControl int         `json:"inbound_object_control"`
}

type VMDetails struct {
	Node

	InboundExecutionPrivileges int `json:"inboundExecutionPrivileges"`
	InboundObjectControl       int `json:"inbound_object_control"`
}

type ManagedClusterDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}

type ContainerRegistryDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}

type WebAppDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}

type AutomationAccountDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}

type FunctionAppDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}

type KeyVaultReaderCounts struct {
	KeyReaders         int `json:"KeyReaders"`
	CertificateReaders int `json:"CertificateReaders"`
	SecretReaders      int `json:"SecretReaders"`
	AllReaders         int `json:"AllReaders"`
}

type KeyVaultDetails struct {
	Node

	Readers              KeyVaultReaderCounts `json:"Readers"`
	InboundObjectControl int                  `json:"inbound_object_control"`
}

type DeviceDetails struct {
	Node

	InboundExecutionPrivileges int `json:"inboundExecutionPrivileges"`
	InboundObjectControl       int `json:"inbound_object_control"`
}

type ApplicationDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}

type VMScaleSetDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}

type ServicePrincipalDetails struct {
	Node

	Roles                              int `json:"roles"`
	InboundObjectControl               int `json:"inbound_object_control"`
	OutboundObjectControl              int `json:"outbound_object_control"`
	InboundAbusableAppRoleAssignments  int `json:"inbound-abusable-app-role-assignments"`
	OutboundAbusableAppRoleAssignments int `json:"outbound-abusable-app-role-assignments"`
}

type RoleDetails struct {
	Node

	ActiveAssignments int `json:"active_assignments"`
	PIMAssignments    int `json:"pim_assignments"`
}

type LogicAppDetails struct {
	Node

	InboundObjectControl int `json:"inbound_object_control"`
}
