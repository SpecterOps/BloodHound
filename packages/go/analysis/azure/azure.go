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
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/azure"
)

const (
	ErrNoNonEntityKindFound     = errors.Error("unable to find a non-entity kind")
	ErrInvalidRelatedEntityType = errors.Error("invalid related entity type")
)

func GetDescendentKinds(kind graph.Kind) []graph.Kind {
	switch kind {
	case azure.Tenant:
		return []graph.Kind{
			azure.User,
			azure.Group,
			azure.ManagementGroup,
			azure.Subscription,
			azure.ResourceGroup,
			azure.VM,
			azure.ManagedCluster,
			azure.VMScaleSet,
			azure.ContainerRegistry,
			azure.WebApp,
			azure.LogicApp,
			azure.AutomationAccount,
			azure.KeyVault,
			azure.App,
			azure.ServicePrincipal,
			azure.Device,
			azure.FunctionApp,
		}

	case azure.ManagementGroup:
		return []graph.Kind{
			azure.ManagementGroup,
			azure.Subscription,
			azure.ResourceGroup,
			azure.VM,
			azure.ManagedCluster,
			azure.VMScaleSet,
			azure.ContainerRegistry,
			azure.WebApp,
			azure.LogicApp,
			azure.AutomationAccount,
			azure.KeyVault,
			azure.FunctionApp,
		}

	case azure.ResourceGroup:
		return []graph.Kind{
			azure.VM,
			azure.ManagedCluster,
			azure.VMScaleSet,
			azure.ContainerRegistry,
			azure.WebApp,
			azure.LogicApp,
			azure.AutomationAccount,
			azure.KeyVault,
			azure.FunctionApp,
		}

	case azure.Subscription:
		return []graph.Kind{
			azure.ResourceGroup,
			azure.VM,
			azure.ManagedCluster,
			azure.VMScaleSet,
			azure.ContainerRegistry,
			azure.WebApp,
			azure.LogicApp,
			azure.AutomationAccount,
			azure.KeyVault,
			azure.FunctionApp,
		}
	}

	return nil
}

func AzureNonDescentKinds() graph.Kinds {
	return []graph.Kind{
		azure.MemberOf,
		azure.HasRole,
		azure.RunsAs,
	}
}

func AzureIgnoredKinds() graph.Kinds {
	return []graph.Kind{
		azure.ScopedTo,
		azure.Contains,
		azure.GlobalAdmin,
		azure.PrivilegedRoleAdmin,
		azure.PrivilegedAuthAdmin,
	}
}
