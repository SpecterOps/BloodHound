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

package v2

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	azure2 "github.com/specterops/bloodhound/src/analysis/azure"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/errors"
)

const (
	relatedEntityReturnTypeList               = "list"
	relatedEntityReturnTypeGraph              = "graph"
	entityTypePathParameterName               = "entity_type"
	objectIDQueryParameterName                = "object_id"
	relatedEntityTypeQueryParameterName       = "related_entity_type"
	relatedEntityReturnTypeQueryParameterName = "type"

	errBadRelatedEntityReturnType = errors.Error("invalid return type requested for related entities")
	errParameterRequired          = errors.Error("missing required parameter")
	errParameterSkip              = errors.Error("invalid skip parameter")
	errParameterRelatedEntityType = errors.Error("invalid related entity type")

	entityTypeBase                = "az-base"
	entityTypeUsers               = "users"
	entityTypeGroups              = "groups"
	entityTypeTenants             = "tenants"
	entityTypeManagementGroups    = "management-groups"
	entityTypeSubscriptions       = "subscriptions"
	entityTypeResourceGroups      = "resource-groups"
	entityTypeVMs                 = "vms"
	entityTypeManagedClusters     = "managed-clusters"
	entityTypeContainerRegistries = "container-registries"
	entityTypeWebApps             = "web-apps"
	entityTypeLogicApps           = "logic-apps"
	entityTypeAutomationAccounts  = "automation-accounts"
	entityTypeKeyVaults           = "key-vaults"
	entityTypeDevices             = "devices"
	entityTypeApplications        = "applications"
	entityTypeVMScaleSets         = "vm-scale-sets"
	entityTypeServicePrincipals   = "service-principals"
	entityTypeRoles               = "roles"
	entityTypeFunctionApps        = "function-apps"
)

func graphRelatedEntityType(ctx context.Context, db graph.Database, entityType, objectID string, request *http.Request) (any, int, *api.ErrorWrapper) {
	switch relatedEntityType := azure.RelatedEntityType(entityType); relatedEntityType {
	case azure.RelatedEntityTypeDescendentUsers, azure.RelatedEntityTypeDescendentGroups,
		azure.RelatedEntityTypeDescendentManagementGroups, azure.RelatedEntityTypeDescendentSubscriptions,
		azure.RelatedEntityTypeDescendentResourceGroups, azure.RelatedEntityTypeDescendentVirtualMachines,
		azure.RelatedEntityTypeDescendentKeyVaults, azure.RelatedEntityTypeDescendentApplications,
		azure.RelatedEntityTypeDescendentServicePrincipals, azure.RelatedEntityTypeDescendentDevices,
		azure.RelatedEntityTypeDescendentManagedClusters,
		azure.RelatedEntityTypeDescendentVMScaleSets,
		azure.RelatedEntityTypeDescendentContainerRegistries,
		azure.RelatedEntityTypeDescendentWebApps,
		azure.RelatedEntityTypeDescendentAutomationAccounts,
		azure.RelatedEntityTypeDescendentLogicApps, azure.RelatedEntityTypeDescendentFunctionApps:
		if descendents, err := azure.ListEntityDescendentPaths(ctx, db, relatedEntityType, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(descendents), descendents.Len(), nil
		}

	case azure.RelatedEntityTypeActiveAssignments:
		if assignments, err := azure.ListEntityActiveAssignmentPaths(ctx, db, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(assignments), assignments.Len(), nil
		}

	case azure.RelatedEntityTypePIMAssignments:
		if assignments, err := azure.ListEntityPIMAssignmentPaths(ctx, db, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(assignments), assignments.Len(), nil
		}

	case azure.RelatedEntityTypeVaultKeyReaders, azure.RelatedEntityTypeVaultSecretReaders, azure.RelatedEntityTypeVaultCertReaders, azure.RelatedEntityTypeVaultAllReaders:
		if groupMembers, err := azure.ListKeyVaultReaderPaths(ctx, db, relatedEntityType, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(groupMembers), groupMembers.Len(), nil
		}

	case azure.RelatedEntityTypeGroupMembers:
		if groupMembers, err := azure.ListEntityGroupMemberPaths(ctx, db, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(groupMembers), groupMembers.Len(), nil
		}

	case azure.RelatedEntityTypeGroupMembership:
		if groupMembership, err := azure.ListEntityGroupMembershipPaths(ctx, db, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(groupMembership), groupMembership.Len(), nil
		}

	case azure.RelatedEntityTypeRoles:
		if userRoles, err := azure.ListEntityRolePaths(ctx, db, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(userRoles), userRoles.Len(), nil
		}

	case azure.RelatedEntityTypeOutboundExecutionPrivileges:
		if executionPrivileges, err := azure.ListEntityExecutionPrivilegePaths(ctx, db, objectID, graph.DirectionOutbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(executionPrivileges), executionPrivileges.Len(), nil
		}

	case azure.RelatedEntityTypeInboundExecutionPrivileges:
		if executionPrivileges, err := azure.ListEntityExecutionPrivilegePaths(ctx, db, objectID, graph.DirectionInbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(executionPrivileges), executionPrivileges.Len(), nil
		}

	case azure.RelatedEntityTypeOutboundAbusableAppRoleAssignments:
		if objectControl, err := azure.ListEntityAbusableAppRoleAssignmentsPaths(ctx, db, objectID, graph.DirectionOutbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(objectControl), objectControl.Len(), nil
		}

	case azure.RelatedEntityTypeInboundAbusableAppRoleAssignments:
		if objectControl, err := azure.ListEntityAbusableAppRoleAssignmentsPaths(ctx, db, objectID, graph.DirectionInbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(objectControl), objectControl.Len(), nil
		}

	case azure.RelatedEntityTypeOutboundControl:
		if objectControl, err := azure.ListEntityObjectControlPaths(ctx, db, objectID, graph.DirectionOutbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(objectControl), objectControl.Len(), nil
		}

	case azure.RelatedEntityTypeInboundControl:
		if objectControl, err := azure.ListEntityObjectControlPaths(ctx, db, objectID, graph.DirectionInbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(objectControl), objectControl.Len(), nil
		}

	default:
		return nil, 0, api.BuildErrorResponse(http.StatusNotFound, fmt.Sprintf("no matching related entity list type for %s", entityType), request)
	}
}

func nodeSetToOrderedSlice(nodeSet graph.NodeSet) []*graph.Node {
	nodes := nodeSet.Slice()

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID > nodes[j].ID
	})

	return nodes
}

func listRelatedEntityType(ctx context.Context, db graph.Database, entityType, objectID string, skip, limit int) ([]azure.Node, int, error) {
	var (
		nodeSet graph.NodeSet
		err     error
	)
	//NOTE: All skip/limit passed to lower level queries is currently hardcoded to 0 so we can get the full count of the dataset for skip/limit tracking
	switch relatedEntityType := azure.RelatedEntityType(entityType); relatedEntityType {
	case azure.RelatedEntityTypeDescendentUsers, azure.RelatedEntityTypeDescendentGroups,
		azure.RelatedEntityTypeDescendentManagementGroups, azure.RelatedEntityTypeDescendentSubscriptions,
		azure.RelatedEntityTypeDescendentResourceGroups, azure.RelatedEntityTypeDescendentVirtualMachines,
		azure.RelatedEntityTypeDescendentKeyVaults, azure.RelatedEntityTypeDescendentApplications,
		azure.RelatedEntityTypeDescendentServicePrincipals, azure.RelatedEntityTypeDescendentDevices,
		azure.RelatedEntityTypeDescendentManagedClusters,
		azure.RelatedEntityTypeDescendentVMScaleSets,
		azure.RelatedEntityTypeDescendentContainerRegistries,
		azure.RelatedEntityTypeDescendentWebApps,
		azure.RelatedEntityTypeDescendentLogicApps, azure.RelatedEntityTypeDescendentFunctionApps,
		azure.RelatedEntityTypeDescendentAutomationAccounts:
		if nodeSet, err = azure.ListEntityDescendents(ctx, db, relatedEntityType, objectID, 0, 0); err != nil {
			return nil, 0, err
		}
	case azure.RelatedEntityTypeActiveAssignments:
		if nodeSet, err = azure.ListEntityActiveAssignments(ctx, db, objectID, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypePIMAssignments:
		if nodeSet, err = azure.ListEntityPIMAssignments(ctx, db, objectID, 0, 0); err != nil {
			return nil, 0, err
		}
	case azure.RelatedEntityTypeVaultKeyReaders, azure.RelatedEntityTypeVaultSecretReaders, azure.RelatedEntityTypeVaultCertReaders, azure.RelatedEntityTypeVaultAllReaders:
		if nodeSet, err = azure.ListKeyVaultReaders(ctx, db, relatedEntityType, objectID, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeGroupMembers:
		if nodeSet, err = azure.ListEntityGroupMembers(ctx, db, objectID, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeGroupMembership:
		if nodeSet, err = azure.ListEntityGroupMembership(ctx, db, objectID, 0, 0); err != nil {
			return nil, 0, err
		}
	case azure.RelatedEntityTypeRoles:
		if nodeSet, err = azure.ListEntityRoles(ctx, db, objectID, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeOutboundExecutionPrivileges:
		if nodeSet, err = azure.ListEntityExecutionPrivileges(ctx, db, objectID, graph.DirectionOutbound, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeInboundExecutionPrivileges:
		if nodeSet, err = azure.ListEntityExecutionPrivileges(ctx, db, objectID, graph.DirectionInbound, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeOutboundAbusableAppRoleAssignments:
		if nodeSet, err = azure.ListEntityAbusableAppRoleAssignments(ctx, db, objectID, graph.DirectionOutbound, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeInboundAbusableAppRoleAssignments:
		if nodeSet, err = azure.ListEntityAbusableAppRoleAssignments(ctx, db, objectID, graph.DirectionInbound, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeOutboundControl:
		if nodeSet, err = azure.ListEntityObjectControl(ctx, db, objectID, graph.DirectionOutbound, 0, 0); err != nil {
			return nil, 0, err
		}

	case azure.RelatedEntityTypeInboundControl:
		if nodeSet, err = azure.ListEntityObjectControl(ctx, db, objectID, graph.DirectionInbound, 0, 0); err != nil {
			return nil, 0, err
		}

	default:
		return nil, 0, errParameterRelatedEntityType
	}

	nodeCount := nodeSet.Len()

	if skip > nodeCount {
		return nil, 0, errParameterSkip
	}

	if skip+limit > nodeCount {
		limit = nodeSet.Len() - skip
	}

	s := nodeSetToOrderedSlice(nodeSet)[skip : skip+limit]

	return azure.FromGraphNodes(s), nodeCount, nil
}

func (s *Resources) GetAZRelatedEntities(ctx context.Context, response http.ResponseWriter, request *http.Request, objectID string) {
	var (
		queryParams = request.URL.Query()
		returnType  = queryParams.Get(relatedEntityReturnTypeQueryParameterName)
	)

	// If return type isn't set default to list
	if returnType == "" {
		returnType = relatedEntityReturnTypeList
	}

	if relatedEntityType := queryParams.Get(relatedEntityTypeQueryParameterName); relatedEntityType == "" {
		api.WriteErrorResponse(ctx, ErrBadQueryParameter(request, relatedEntityTypeQueryParameterName, errParameterRequired), response)
	} else if returnType != relatedEntityReturnTypeGraph && returnType != relatedEntityReturnTypeList {
		api.WriteErrorResponse(ctx, ErrBadQueryParameter(request, "type", errBadRelatedEntityReturnType), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(ctx, ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 100); err != nil {
		api.WriteErrorResponse(ctx, ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if returnType == relatedEntityReturnTypeGraph {
		if data, _, apiErr := graphRelatedEntityType(ctx, s.Graph, relatedEntityType, objectID, request); apiErr != nil {
			api.WriteErrorResponse(ctx, apiErr, response)
		} else {
			api.WriteJSONResponse(ctx, data, http.StatusOK, response)
		}
	} else {
		if nodes, count, err := listRelatedEntityType(ctx, s.Graph, relatedEntityType, objectID, skip, limit); err != nil {
			if errors.Is(err, errParameterSkip) {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, skip), request), response)
			} else if errors.Is(err, errParameterRelatedEntityType) {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusNotFound, fmt.Sprintf("no matching related entity list type for %s", relatedEntityType), request), response)
			} else if errors.Is(err, ops.ErrTraversalMemoryLimit) {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "calculating the request results exceeded memory limitations due to the volume of objects involved", request), response)
			} else {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "an unknown error occurred during the request", request), response)
			}
		} else {
			api.WriteResponseWrapperWithPagination(ctx, nodes, limit, skip, count, http.StatusOK, response)
		}
	}
}

func GetAZEntityInformation(ctx context.Context, db graph.Database, entityType, objectID string, hydrateCounts bool) (any, error) {
	switch entityType {
	case entityTypeBase:
		return azure2.BaseEntityDetails(db, objectID, hydrateCounts)
	case entityTypeUsers:
		return azure.UserEntityDetails(db, objectID, hydrateCounts)

	case entityTypeGroups:
		return azure.GroupEntityDetails(db, objectID, hydrateCounts)

	case entityTypeTenants:
		return azure.TenantEntityDetails(db, objectID, hydrateCounts)

	case entityTypeManagementGroups:
		return azure.ManagementGroupEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeSubscriptions:
		return azure.SubscriptionEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeResourceGroups:
		return azure.ResourceGroupEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeVMs:
		return azure.VMEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeManagedClusters:
		return azure.ManagedClusterEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeContainerRegistries:
		return azure.ContainerRegistryEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeWebApps:
		return azure.WebAppEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeLogicApps:
		return azure.LogicAppEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeAutomationAccounts:
		return azure.AutomationAccountEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeKeyVaults:
		return azure.KeyVaultEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeDevices:
		return azure.DeviceEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeApplications:
		return azure.ApplicationEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeVMScaleSets:
		return azure.VMScaleSetEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeServicePrincipals:
		return azure.ServicePrincipalEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeRoles:
		return azure.RoleEntityDetails(ctx, db, objectID, hydrateCounts)

	case entityTypeFunctionApps:
		return azure.FunctionAppEntityDetails(ctx, db, objectID, hydrateCounts)

	default:
		return nil, fmt.Errorf("unknown azure entity %s", entityType)
	}
}

func (s *Resources) GetAZEntity(response http.ResponseWriter, request *http.Request) {
	var (
		requestVars = mux.Vars(request)
		queryVars   = request.URL.Query()
		entityType  = requestVars[entityTypePathParameterName]
	)

	if objectID := queryVars.Get(objectIDQueryParameterName); objectID == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("query parameter %s is required", objectIDQueryParameterName), request), response)
	} else if relatedEntityTypeStr := queryVars.Get(relatedEntityTypeQueryParameterName); relatedEntityTypeStr != "" {
		s.GetAZRelatedEntities(request.Context(), response, request, objectID)
	} else if hydrateCounts, err := api.ParseOptionalBool(queryVars.Get(api.QueryParameterHydrateCounts), true); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if entityInformation, err := GetAZEntityInformation(request.Context(), s.Graph, entityType, objectID, hydrateCounts); err != nil {
		if graph.IsErrNotFound(err) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "not found", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err), request), response)
		}
	} else {
		api.WriteBasicResponse(request.Context(), entityInformation, http.StatusOK, response)
	}
}
