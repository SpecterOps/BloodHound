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
	"context"
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/graphschema/azure"
)

func ListEntityDescendentPaths(ctx context.Context, db graph.Database, relatedEntityType RelatedEntityType, objectID string) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			var targetKind graph.Kind

			switch relatedEntityType {
			case RelatedEntityTypeDescendentUsers:
				targetKind = azure.User
			case RelatedEntityTypeDescendentGroups:
				targetKind = azure.Group
			case RelatedEntityTypeDescendentManagementGroups:
				targetKind = azure.ManagementGroup
			case RelatedEntityTypeDescendentSubscriptions:
				targetKind = azure.Subscription
			case RelatedEntityTypeDescendentResourceGroups:
				targetKind = azure.ResourceGroup
			case RelatedEntityTypeDescendentVirtualMachines:
				targetKind = azure.VM
			case RelatedEntityTypeDescendentManagedClusters:
				targetKind = azure.ManagedCluster
			case RelatedEntityTypeDescendentWebApps:
				targetKind = azure.WebApp
			case RelatedEntityTypeDescendentLogicApps:
				targetKind = azure.LogicApp
			case RelatedEntityTypeDescendentAutomationAccounts:
				targetKind = azure.AutomationAccount
			case RelatedEntityTypeDescendentKeyVaults:
				targetKind = azure.KeyVault
			case RelatedEntityTypeDescendentApplications:
				targetKind = azure.App
			case RelatedEntityTypeDescendentVMScaleSets:
				targetKind = azure.VMScaleSet
			case RelatedEntityTypeDescendentServicePrincipals:
				targetKind = azure.ServicePrincipal
			case RelatedEntityTypeDescendentDevices:
				targetKind = azure.Device
			case RelatedEntityTypeDescendentContainerRegistries:
				targetKind = azure.ContainerRegistry
			case RelatedEntityTypeDescendentFunctionApps:
				targetKind = azure.FunctionApp
			default:
				return ErrInvalidRelatedEntityType
			}

			paths, err = FetchEntityDescendentPaths(tx, node, targetKind)
			return err
		}
	})
}

func ListEntityDescendents(ctx context.Context, db graph.Database, relatedEntityType RelatedEntityType, objectID string, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			var targetKind graph.Kind

			switch relatedEntityType {
			case RelatedEntityTypeDescendentUsers:
				targetKind = azure.User
			case RelatedEntityTypeDescendentGroups:
				targetKind = azure.Group
			case RelatedEntityTypeDescendentManagementGroups:
				targetKind = azure.ManagementGroup
			case RelatedEntityTypeDescendentSubscriptions:
				targetKind = azure.Subscription
			case RelatedEntityTypeDescendentResourceGroups:
				targetKind = azure.ResourceGroup
			case RelatedEntityTypeDescendentVirtualMachines:
				targetKind = azure.VM
			case RelatedEntityTypeDescendentManagedClusters:
				targetKind = azure.ManagedCluster
			case RelatedEntityTypeDescendentWebApps:
				targetKind = azure.WebApp
			case RelatedEntityTypeDescendentLogicApps:
				targetKind = azure.LogicApp
			case RelatedEntityTypeDescendentAutomationAccounts:
				targetKind = azure.AutomationAccount
			case RelatedEntityTypeDescendentKeyVaults:
				targetKind = azure.KeyVault
			case RelatedEntityTypeDescendentApplications:
				targetKind = azure.App
			case RelatedEntityTypeDescendentVMScaleSets:
				targetKind = azure.VMScaleSet
			case RelatedEntityTypeDescendentServicePrincipals:
				targetKind = azure.ServicePrincipal
			case RelatedEntityTypeDescendentDevices:
				targetKind = azure.Device
			case RelatedEntityTypeDescendentContainerRegistries:
				targetKind = azure.ContainerRegistry
			case RelatedEntityTypeDescendentFunctionApps:
				targetKind = azure.FunctionApp
			default:
				return ErrInvalidRelatedEntityType
			}

			nodes, err = FetchEntityDescendents(tx, node, skip, limit, targetKind)
			return err
		}
	})
}

func ListKeyVaultReaderPaths(ctx context.Context, db graph.Database, relatedEntityType RelatedEntityType, objectID string) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			switch relatedEntityType {
			case RelatedEntityTypeVaultAllReaders:
				paths, err = FetchKeyVaultReaderPaths(tx, node)

			case RelatedEntityTypeVaultCertReaders:
				paths, err = ops.TraversePaths(tx, ops.TraversalPlan{
					Root:        node,
					Direction:   graph.DirectionInbound,
					BranchQuery: FilterCertificateReaders,
				})

			case RelatedEntityTypeVaultKeyReaders:
				paths, err = ops.TraversePaths(tx, ops.TraversalPlan{
					Root:        node,
					Direction:   graph.DirectionInbound,
					BranchQuery: FilterKeyReaders,
				})

			case RelatedEntityTypeVaultSecretReaders:
				paths, err = ops.TraversePaths(tx, ops.TraversalPlan{
					Root:        node,
					Direction:   graph.DirectionInbound,
					BranchQuery: FilterSecretReaders,
				})

			default:
				panic(fmt.Sprintf("invalid reader type: %s", relatedEntityType))
			}

			return err
		}
	})
}

func ListKeyVaultReaders(ctx context.Context, db graph.Database, readerType RelatedEntityType, objectID string, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			switch readerType {
			case RelatedEntityTypeVaultAllReaders:
				nodes, err = FetchKeyVaultReaders(tx, node, skip, limit)

			case RelatedEntityTypeVaultCertReaders:
				nodes, err = ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
					Root:        node,
					Direction:   graph.DirectionInbound,
					BranchQuery: FilterCertificateReaders,
					Skip:        skip,
					Limit:       limit,
				})

			case RelatedEntityTypeVaultKeyReaders:
				nodes, err = ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
					Root:        node,
					Direction:   graph.DirectionInbound,
					BranchQuery: FilterKeyReaders,
					Skip:        skip,
					Limit:       limit,
				})

			case RelatedEntityTypeVaultSecretReaders:
				nodes, err = ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
					Root:        node,
					Direction:   graph.DirectionInbound,
					BranchQuery: FilterSecretReaders,
					Skip:        skip,
					Limit:       limit,
				})

			default:
				panic(fmt.Sprintf("invalid reader type: %s", readerType))
			}

			return err
		}
	})
}

func ListEntityRolePaths(ctx context.Context, db graph.Database, objectID string) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			paths, err = FetchEntityRolePaths(tx, node)
			return err
		}
	})
}

func ListEntityRoles(ctx context.Context, db graph.Database, objectID string, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			nodes, err = FetchEntityRoles(tx, node, skip, limit)
			return err
		}
	})
}

func ListEntityExecutionPrivilegePaths(ctx context.Context, db graph.Database, objectID string, direction graph.Direction) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else if direction == graph.DirectionOutbound {
			paths, err = FetchOutboundEntityExecutionPrivilegePaths(tx, node, direction)
			return err
		} else {
			paths, err = FetchInboundEntityExecutionPrivilegePaths(tx, node, direction)
			return err
		}
	})
}

func ListEntityExecutionPrivileges(ctx context.Context, db graph.Database, objectID string, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else if direction == graph.DirectionOutbound {
			nodes, err = FetchOutboundEntityExecutionPrivileges(tx, node, direction, skip, limit)
			return err
		} else {
			nodes, err = FetchInboundEntityExecutionPrivileges(tx, node, direction, skip, limit)
			return err
		}
	})
}

func ListEntityAbusableAppRoleAssignmentsPaths(ctx context.Context, db graph.Database, objectID string, direction graph.Direction) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			paths, err = FetchAbusableAppRoleAssignmentsPaths(tx, node, direction)
			return err
		}
	})
}

func ListEntityAbusableAppRoleAssignments(ctx context.Context, db graph.Database, objectID string, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			nodes, err = FetchAbusableAppRoleAssignments(tx, node, direction, skip, limit)
			return err
		}
	})
}

func ListEntityObjectControlPaths(ctx context.Context, db graph.Database, objectID string, direction graph.Direction) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else if direction == graph.DirectionOutbound {
			paths, err = FetchOutboundEntityObjectControlPaths(tx, node, direction)
			return err
		} else {
			paths, err = FetchInboundEntityObjectControlPaths(tx, node, direction)
			return err
		}
	})
}

func ListEntityObjectControl(ctx context.Context, db graph.Database, objectID string, direction graph.Direction, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else if direction == graph.DirectionOutbound {
			nodes, err = FetchOutboundEntityObjectControl(tx, node, direction, skip, limit)
			return err
		} else {
			nodes, err = FetchInboundEntityObjectControllers(tx, node, direction, skip, limit)
			return err
		}
	})
}

func ListEntityGroupMembershipPaths(ctx context.Context, db graph.Database, objectID string) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			paths, err = FetchEntityGroupMembershipPaths(tx, node)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func ListEntityGroupMembership(ctx context.Context, db graph.Database, objectID string, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			nodes, err = FetchEntityGroupMembership(tx, node, skip, limit)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func ListEntityGroupMemberPaths(ctx context.Context, db graph.Database, objectID string) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			paths, err = FetchGroupMemberPaths(tx, node)
			return err
		}
	})
}

func ListEntityGroupMembers(ctx context.Context, db graph.Database, objectID string, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			nodes, err = FetchGroupMembers(tx, node, skip, limit)
			return err
		}
	})
}

func ListEntityActiveAssignmentPaths(ctx context.Context, db graph.Database, objectID string) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			paths, err = FetchEntityActiveAssignmentPaths(tx, node)
			return err
		}
	})
}

func ListEntityActiveAssignments(ctx context.Context, db graph.Database, objectID string, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			nodes, err = FetchEntityActiveAssignments(tx, node, skip, limit)
			return err
		}
	})
}

func ListEntityPIMAssignmentPaths(ctx context.Context, db graph.Database, objectID string) (graph.PathSet, error) {
	var paths graph.PathSet

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			paths, err = FetchEntityPIMAssignmentPaths(tx, node)
			return err
		}
	})
}

func ListEntityPIMAssignments(ctx context.Context, db graph.Database, objectID string, skip, limit int) (graph.NodeSet, error) {
	var nodes graph.NodeSet

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			nodes, err = FetchEntityPIMAssignments(tx, node, skip, limit)
			return err
		}
	})
}
