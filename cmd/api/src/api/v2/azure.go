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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	bhCtx "github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	azure_schema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
)

const (
	relatedEntityReturnTypeList               = "list"
	relatedEntityReturnTypeGraph              = "graph"
	entityTypePathParameterName               = "entity_type"
	objectIDQueryParameterName                = "object_id"
	relatedEntityTypeQueryParameterName       = "related_entity_type"
	relatedEntityReturnTypeQueryParameterName = "type"

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

var (
	errBadRelatedEntityReturnType = errors.New("invalid return type requested for related entities")
	errParameterRequired          = errors.New("missing required parameter")
	ErrParameterSkip              = errors.New("invalid skip parameter")
	ErrParameterRelatedEntityType = errors.New("invalid related entity type")
)

func graphRelatedEntityType(request *http.Request, db database.Database, graphDb graph.Database, entityType, objectID string) (any, int, *api.ErrorWrapper) {
	ctx := request.Context()

	customNodeKinds, err := db.GetCustomNodeKindsMap(ctx)
	if err != nil {
		slog.Error("Unable to fetch custom nodes from database; will fall back to defaults")
	}
	validPrimaryKinds, err := db.GetDisplayNodeGraphKinds(request.Context())
	if err != nil {
		return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching valid primary kinds: %v", err), request)
	}

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
		if descendents, err := azure.ListEntityDescendentPaths(ctx, graphDb, relatedEntityType, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, descendents), descendents.Len(), nil
		}

	case azure.RelatedEntityTypeActiveAssignments:
		if assignments, err := azure.ListEntityActiveAssignmentPaths(ctx, graphDb, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, assignments), assignments.Len(), nil
		}

	case azure.RelatedEntityTypePIMAssignments:
		if assignments, err := azure.ListEntityPIMAssignmentPaths(ctx, graphDb, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, assignments), assignments.Len(), nil
		}
	case azure.RelatedEntityTypeRoleApprovers:
		if approvers, err := azure.ListRoleApproverPaths(ctx, graphDb, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, approvers), approvers.Len(), nil
		}
	case azure.RelatedEntityTypeVaultKeyReaders, azure.RelatedEntityTypeVaultSecretReaders, azure.RelatedEntityTypeVaultCertReaders, azure.RelatedEntityTypeVaultAllReaders:
		if groupMembers, err := azure.ListKeyVaultReaderPaths(ctx, graphDb, relatedEntityType, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, groupMembers), groupMembers.Len(), nil
		}

	case azure.RelatedEntityTypeGroupMembers:
		if groupMembers, err := azure.ListEntityGroupMemberPaths(ctx, graphDb, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, groupMembers), groupMembers.Len(), nil
		}

	case azure.RelatedEntityTypeGroupMembership:
		if groupMembership, err := azure.ListEntityGroupMembershipPaths(ctx, graphDb, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, groupMembership), groupMembership.Len(), nil
		}

	case azure.RelatedEntityTypeRoles:
		if userRoles, err := azure.ListEntityRolePaths(ctx, graphDb, objectID); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, userRoles), userRoles.Len(), nil
		}

	case azure.RelatedEntityTypeOutboundExecutionPrivileges:
		if executionPrivileges, err := azure.ListEntityExecutionPrivilegePaths(ctx, graphDb, objectID, graph.DirectionOutbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, executionPrivileges), executionPrivileges.Len(), nil
		}

	case azure.RelatedEntityTypeInboundExecutionPrivileges:
		if executionPrivileges, err := azure.ListEntityExecutionPrivilegePaths(ctx, graphDb, objectID, graph.DirectionInbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, executionPrivileges), executionPrivileges.Len(), nil
		}

	case azure.RelatedEntityTypeOutboundAbusableAppRoleAssignments:
		if objectControl, err := azure.ListEntityAbusableAppRoleAssignmentsPaths(ctx, graphDb, objectID, graph.DirectionOutbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, objectControl), objectControl.Len(), nil
		}

	case azure.RelatedEntityTypeInboundAbusableAppRoleAssignments:
		if objectControl, err := azure.ListEntityAbusableAppRoleAssignmentsPaths(ctx, graphDb, objectID, graph.DirectionInbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, objectControl), objectControl.Len(), nil
		}

	case azure.RelatedEntityTypeOutboundControl:
		if objectControl, err := azure.ListEntityObjectControlPaths(ctx, graphDb, objectID, graph.DirectionOutbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, objectControl), objectControl.Len(), nil
		}

	case azure.RelatedEntityTypeInboundControl:
		if objectControl, err := azure.ListEntityObjectControlPaths(ctx, graphDb, objectID, graph.DirectionInbound); err != nil {
			return nil, 0, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error fetching related entity type %s: %v", entityType, err), request)
		} else {
			return bloodhoundgraph.PathSetToBloodHoundGraph(validPrimaryKinds, customNodeKinds, objectControl), objectControl.Len(), nil
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

func listRelatedEntityType(ctx context.Context, db graph.Database, validPrimaryKinds graphschema.ValidPrimaryKinds, entityType, objectID string, skip, limit int) ([]azure.Node, int, error) {
	var (
		nodeSet graph.NodeSet
		err     error
	)
	// NOTE: All skip/limit passed to lower level queries is currently hardcoded to 0 so we can get the full count of the dataset for skip/limit tracking
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
	case azure.RelatedEntityTypeRoleApprovers:
		if nodeSet, err = azure.ListRoleApprovers(ctx, db, objectID, 0, 0); err != nil {
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
		return nil, 0, ErrParameterRelatedEntityType
	}

	nodeCount := nodeSet.Len()

	if skip > nodeCount {
		return nil, 0, ErrParameterSkip
	}

	if skip+limit > nodeCount {
		limit = nodeSet.Len() - skip
	}

	s := nodeSetToOrderedSlice(nodeSet)[skip : skip+limit]

	return azure.FromGraphNodes(validPrimaryKinds, s), nodeCount, nil
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
		if data, _, apiErr := graphRelatedEntityType(request, s.DB, s.Graph, relatedEntityType, objectID); apiErr != nil {
			api.WriteErrorResponse(ctx, apiErr, response)
		} else {
			api.WriteJSONResponse(ctx, data, http.StatusOK, response)
		}
	} else if validPrimaryKinds, err := s.DB.GetDisplayNodeGraphKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		if nodes, count, err := listRelatedEntityType(ctx, s.Graph, validPrimaryKinds, relatedEntityType, objectID, skip, limit); err != nil {
			if errors.Is(err, ErrParameterSkip) {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, skip), request), response)
			} else if errors.Is(err, ErrParameterRelatedEntityType) {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusNotFound, fmt.Sprintf("no matching related entity list type for %s", relatedEntityType), request), response)
			} else if errors.Is(err, ops.ErrGraphQueryMemoryLimit) {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "calculating the request results exceeded memory limitations due to the volume of objects involved", request), response)
			} else {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "an unknown error occurred during the request", request), response)
			}
		} else {
			api.WriteResponseWrapperWithPagination(ctx, nodes, limit, skip, count, http.StatusOK, response)
		}
	}
}

func GetAZEntityInformation(ctx context.Context, db database.Database, graphDb graph.Database, entityType, objectID string, hydrateCounts bool) (any, error) {
	validPrimaryKinds, err := db.GetDisplayNodeGraphKinds(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching valid primary kinds: %v", err)
	}

	switch entityType {
	case entityTypeBase:
		return azure.BaseEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeUsers:
		return azure.UserEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeGroups:
		return azure.GroupEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeTenants:
		return azure.TenantEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeManagementGroups:
		return azure.ManagementGroupEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeSubscriptions:
		return azure.SubscriptionEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeResourceGroups:
		return azure.ResourceGroupEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeVMs:
		return azure.VMEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeManagedClusters:
		return azure.ManagedClusterEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeContainerRegistries:
		return azure.ContainerRegistryEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeWebApps:
		return azure.WebAppEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeLogicApps:
		return azure.LogicAppEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeAutomationAccounts:
		return azure.AutomationAccountEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeKeyVaults:
		return azure.KeyVaultEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeDevices:
		return azure.DeviceEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeApplications:
		return azure.ApplicationEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeVMScaleSets:
		return azure.VMScaleSetEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeServicePrincipals:
		return azure.ServicePrincipalEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeRoles:
		return azure.RoleEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
	case entityTypeFunctionApps:
		return azure.FunctionAppEntityDetails(ctx, graphDb, validPrimaryKinds, objectID, hydrateCounts)
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

	user, isUser := auth.GetUserFromAuthCtx(bhCtx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	}

	if objectID := queryVars.Get(objectIDQueryParameterName); objectID == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("query parameter %s is required", objectIDQueryParameterName), request), response)
	} else if azKind, err := azEntityParamToKind(entityType); err != nil {
		slog.WarnContext(request.Context(), "Could not determine AZ type from entityType request var",
			slog.String("entity_type", entityType),
			attr.Error(err))
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if hasAccess, err := CheckUserHasAccessToNodeById(request.Context(), s.DB, s.GraphQuery, s.DogTags, user, objectID, azKind); err != nil {
		slog.ErrorContext(request.Context(), "Error checking if user has access to node for ETAC", attr.Error(err))
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if !hasAccess {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
	} else if relatedEntityTypeStr := queryVars.Get(relatedEntityTypeQueryParameterName); relatedEntityTypeStr != "" {
		s.GetAZRelatedEntities(request.Context(), response, request, objectID)
	} else if includeCounts, err := api.ParseOptionalBool(queryVars.Get(api.QueryParameterIncludeCounts), true); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if entityInformation, err := GetAZEntityInformation(request.Context(), s.DB, s.Graph, entityType, objectID, includeCounts); err != nil {
		if graph.IsErrNotFound(err) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "not found", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err), request), response)
		}
	} else {
		api.WriteBasicResponse(request.Context(), entityInformation, http.StatusOK, response)
	}
}

// azEntityParamToKind takes a string which is parsed from a user's request params and converts it to a known Azure kind
// For example: `az-base` becomes the Kind `AZBase`
func azEntityParamToKind(entityType string) (graph.Kind, error) {
	switch entityType {
	case entityTypeBase:
		return azure_schema.Entity, nil
	case entityTypeUsers:
		return azure_schema.User, nil

	case entityTypeGroups:
		return azure_schema.Group, nil

	case entityTypeTenants:
		return azure_schema.Tenant, nil

	case entityTypeManagementGroups:
		return azure_schema.ManagementGroup, nil

	case entityTypeSubscriptions:
		return azure_schema.Subscription, nil

	case entityTypeResourceGroups:
		return azure_schema.ResourceGroup, nil

	case entityTypeVMs:
		return azure_schema.VM, nil

	case entityTypeManagedClusters:
		return azure_schema.ManagedCluster, nil

	case entityTypeContainerRegistries:
		return azure_schema.ContainerRegistry, nil

	case entityTypeWebApps:
		return azure_schema.WebApp, nil

	case entityTypeLogicApps:
		return azure_schema.LogicApp, nil

	case entityTypeAutomationAccounts:
		return azure_schema.AutomationAccount, nil

	case entityTypeKeyVaults:
		return azure_schema.KeyVault, nil

	case entityTypeDevices:
		return azure_schema.Device, nil

	case entityTypeApplications:
		return azure_schema.App, nil

	case entityTypeVMScaleSets:
		return azure_schema.VMScaleSet, nil

	case entityTypeServicePrincipals:
		return azure_schema.ServicePrincipal, nil

	case entityTypeRoles:
		return azure_schema.Role, nil

	case entityTypeFunctionApps:
		return azure_schema.FunctionApp, nil

	default:
		return nil, fmt.Errorf("unknown azure entity %s", entityType)
	}
}
