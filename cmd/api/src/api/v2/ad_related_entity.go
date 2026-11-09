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
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
)

// handleAdRelatedEntityQuery is for retrieving entities related to a specific entity.
// This endpoint returns a polymorphic response dependent upon which return type is requested
// by the `type` parameter, which can be `list` or `graph`.
// Path delegates are for graphing and list delegates are for listing. Endpoints
// without a certain delegate do not support that delegate feature. Related results
// are authorized before formatting, counting, and pagination.
func (s *Resources) handleAdRelatedEntityQuery(response http.ResponseWriter, request *http.Request, queryName string, pathDelegate any, listDelegate any) {
	user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	}

	if params, err := queries.BuildEntityQueryParams(request, queryName, pathDelegate, listDelegate); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.FmtErrorResponseDetailsBadQueryParameters, err), request), response)
	} else if skip, limit, _, err := utils.GetPageParamsForGraphQuery(request.Context(), request.URL.Query()); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.FmtErrorResponseDetailsBadQueryParameters, fmt.Errorf("error getting paging parameters: %w", err)), request), response)
	} else if hasAccess, err := CheckUserHasAccessToNodeById(request.Context(), s.GraphQuery, s.DogTags, user, params.ObjectID, ad.Entity); err != nil {
		slog.ErrorContext(request.Context(), "Error checking if user has access to node for ETAC", attr.Error(err))
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if !hasAccess {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
	} else if entityPanelCachingFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureEntityPanelCaching); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if primaryDisplayKinds, err := s.DB.GetPrimaryDisplayKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		if results, err := s.GraphQuery.GetADEntityQueryResult(request.Context(), params, entityPanelCachingFlag.Enabled); err != nil {
			if errors.Is(err, queries.ErrGraphUnsupported) || errors.Is(err, queries.ErrUnsupportedDataType) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.FmtErrorResponseDetailsBadQueryParameters, err), request), response)
			} else if errors.Is(err, ops.ErrGraphQueryMemoryLimit) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "calculating the request results exceeded memory limitations due to the volume of objects involved", request), response)
			} else {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "an unknown error occurred during the request", request), response)
			}
		} else if params.RequestedType == model.DataTypeGraph {
			var (
				paths, isPathSet = results.(graph.PathSet)
				graphResponse    map[string]any
			)

			if !isPathSet {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
				return
			}

			if ShouldFilterForETAC(s.DogTags, user) {
				graphResponse = pathSetToBloodHoundGraphETAC(primaryDisplayKinds, paths, ExtractEnvironmentIDsFromUser(&user))
			} else {
				graphResponse = bloodhoundgraph.PathSetToBloodHoundGraph(primaryDisplayKinds, paths)
			}

			api.WriteJSONResponse(request.Context(), graphResponse, http.StatusOK, response)
		} else {
			nodes, isNodeSet := results.(graph.NodeSet)
			if !isNodeSet {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
				return
			}

			if ShouldFilterForETAC(s.DogTags, user) {
				nodes = filterNodeSetByETAC(nodes, ExtractEnvironmentIDsFromUser(&user))
			}

			if paginatedNodes, err := paginateNodes(nodes, skip, limit); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			} else {
				api.WriteResponseWrapperWithPagination(request.Context(), adNodesToListEntries(primaryDisplayKinds, paginatedNodes), limit, skip, nodes.Len(), http.StatusOK, response)
			}
		}
	}
}

// adNodesToListEntries formats authorized, paginated nodes for AD list responses.
func adNodesToListEntries(primaryDisplayKinds graphschema.PrimaryDisplayKinds, nodes []*graph.Node) []model.PagedNodeListEntry {
	var renderedNodes = make([]model.PagedNodeListEntry, 0, len(nodes))

	for _, node := range nodes {
		var (
			nodeEntry      model.PagedNodeListEntry
			nodeProperties = node.Properties
		)

		if objectID, err := nodeProperties.Get(common.ObjectID.String()).String(); err != nil {
			if errors.Is(err, graph.ErrPropertyNotFound) {
				slog.Warn("Node missing objectid", slog.Int("node_id", int(node.ID)), attr.Error(err))
			} else {
				slog.Error("Error getting node objectid", slog.Int("node_id", int(node.ID)), attr.Error(err))
			}
		} else {
			nodeEntry.ObjectID = objectID
		}

		if name, err := nodeProperties.Get(common.Name.String()).String(); err != nil {
			if errors.Is(err, graph.ErrPropertyNotFound) {
				slog.Warn("Node missing name", slog.Int("node_id", int(node.ID)), attr.Error(err))
			} else {
				slog.Error("Error getting node name", slog.Int("node_id", int(node.ID)), attr.Error(err))
			}
		} else {
			nodeEntry.Name = name
		}

		nodeEntry.Label = graphschema.GetNodeKindDisplayLabel(primaryDisplayKinds, node)
		nodeEntry.Kinds = node.Kinds.Strings()
		renderedNodes = append(renderedNodes, nodeEntry)
	}

	return renderedNodes
}

func (s *Resources) ListADUserSessions(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADUserSessions", adAnalysis.FetchUserSessionPaths, adAnalysis.FetchUserSessions)
}

func (s *Resources) ListADUserSQLAdminRights(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADUserSQLAdminRights", adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionOutbound), adAnalysis.CreateSQLAdminListDelegate(graph.DirectionOutbound))
}

func (s *Resources) ListADGroupSessions(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGroupSessions", adAnalysis.FetchGroupSessionPaths, adAnalysis.FetchGroupSessions)
}

func (s *Resources) ListADComputerSessions(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADComputerSessions", adAnalysis.FetchComputerSessionPaths, adAnalysis.FetchComputerSessions)
}

func (s *Resources) ListADComputerAdmins(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADComputerAdmins", adAnalysis.CreateInboundLocalGroupPathDelegate(ad.AdminTo), adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo))
}

func (s *Resources) ListADComputerPSRemoteUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADComputerPSRemoteUsers", adAnalysis.CreateInboundLocalGroupPathDelegate(ad.CanPSRemote), adAnalysis.CreateInboundLocalGroupListDelegate(ad.CanPSRemote))
}

func (s *Resources) ListADComputerRDPUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADComputerRDPUsers", adAnalysis.CreateInboundLocalGroupPathDelegate(ad.CanRDP), adAnalysis.CreateInboundLocalGroupListDelegate(ad.CanRDP))
}

func (s *Resources) ListADComputerDCOMUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADComputerDCOMUsers", adAnalysis.CreateInboundLocalGroupPathDelegate(ad.ExecuteDCOM), adAnalysis.CreateInboundLocalGroupListDelegate(ad.ExecuteDCOM))
}

func (s *Resources) ListADGroupMembership(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGroupMembership", adAnalysis.FetchEntityGroupMembershipPaths, adAnalysis.FetchEntityGroupMembership)
}

func (s *Resources) ListADGroupMembers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGroupMembers", adAnalysis.FetchGroupMemberPaths, adAnalysis.FetchGroupMembers)
}

func (s *Resources) ListADComputerSQLAdmins(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADComputerSQLAdmins", adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionInbound), adAnalysis.CreateSQLAdminListDelegate(graph.DirectionInbound))
}

func (s *Resources) ListADComputerConstrainedDelegationUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADComputerConstrainedDelegationUsers", adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionInbound), adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionInbound))
}

func (s *Resources) ListADEntityConstrainedDelegationRights(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADEntityConstrainedDelegationRights", adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionOutbound), adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionOutbound))
}

func (s *Resources) ListADEntityAdminRights(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADEntityAdminRights", adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo), adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo))
}

func (s *Resources) ListADEntityRDPRights(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADEntityRDPRights", adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.CanRDP), adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanRDP))
}

func (s *Resources) ListADEntityPSRemoteRights(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADEntityPSRemoteRights", adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.CanPSRemote), adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanPSRemote))
}

func (s *Resources) ListADEntityDCOMRights(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADEntityDCOMRights", adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.ExecuteDCOM), adAnalysis.CreateOutboundLocalGroupListDelegate(ad.ExecuteDCOM))
}

func (s *Resources) ListADEntityControllers(response http.ResponseWriter, request *http.Request) {

	s.handleAdRelatedEntityQuery(response, request, "ListADEntityControllers", adAnalysis.FetchInboundADEntityControllerPaths, adAnalysis.FetchInboundADEntityControllers)
}

func (s *Resources) ListADEntityControllables(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADEntityControllables", adAnalysis.FetchOutboundADEntityControlPaths, adAnalysis.FetchOutboundADEntityControl)
}

func (s *Resources) ListADEntityLinkedGPOs(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADEntityLinkedGPOs", adAnalysis.FetchEnforcedGPOsPaths, adAnalysis.FetchEnforcedGPOs)
}

func (s *Resources) ListADDomainContainedUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainContainedUsers", nil, adAnalysis.CreateDomainContainedEntityListDelegate(ad.User))
}

func (s *Resources) ListADDomainContainedComputers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainContainedComputers", nil, adAnalysis.CreateDomainContainedEntityListDelegate(ad.Computer))
}

func (s *Resources) ListADDomainContainedGroups(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainContainedGroups", nil, adAnalysis.CreateDomainContainedEntityListDelegate(ad.Group))
}

func (s *Resources) ListADDomainContainedOUs(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainContainedOUs", nil, adAnalysis.CreateDomainContainedEntityListDelegate(ad.OU))
}

func (s *Resources) ListADDomainContainedGPOs(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainContainedGPOs", nil, adAnalysis.CreateDomainContainedEntityListDelegate(ad.GPO))
}

func (s *Resources) ListADDomainForeignGroups(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainForeignGroups", adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.Group), adAnalysis.CreateForeignEntityMembershipListDelegate(ad.Group))
}

func (s *Resources) ListADDomainForeignUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainForeignUsers", adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.User), adAnalysis.CreateForeignEntityMembershipListDelegate(ad.User))
}

func (s *Resources) ListADDomainForeignAdmins(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainForeignAdmins", adAnalysis.FetchForeignAdminPaths, adAnalysis.FetchForeignAdmins)
}

func (s *Resources) ListADDomainForeignGPOControllers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainForeignGPOControllers", adAnalysis.FetchForeignGPOControllerPaths, adAnalysis.FetchForeignGPOControllers)
}

func (s *Resources) ListADDomainOutboundTrusts(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainOutboundTrusts", adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionOutbound), adAnalysis.CreateDomainTrustListDelegate(graph.DirectionOutbound))
}

func (s *Resources) ListADDomainInboundTrusts(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainInboundTrusts", adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionInbound), adAnalysis.CreateDomainTrustListDelegate(graph.DirectionInbound))
}

func (s *Resources) ListADDomainDCSyncers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADDomainDCSyncers", adAnalysis.FetchDCSyncerPaths, adAnalysis.FetchDCSyncers)
}

func (s *Resources) ListADOUContainedUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADOUContainedUsers", adAnalysis.CreateContainedPathDelegate(ad.User), adAnalysis.CreateContainedListDelegate(ad.User))
}

func (s *Resources) ListADOUContainedGroups(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADOUContainedGroups", adAnalysis.CreateContainedPathDelegate(ad.Group), adAnalysis.CreateContainedListDelegate(ad.Group))
}

func (s *Resources) ListADOUContainedComputers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADOUContainedComputers", adAnalysis.CreateContainedPathDelegate(ad.Computer), adAnalysis.CreateContainedListDelegate(ad.Computer))
}

func (s *Resources) ListADGPOAffectedContainers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGPOAffectedContainers", adAnalysis.FetchGPOAffectedContainerPaths, adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter))
}

func (s *Resources) ListADGPOAffectedUsers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGPOAffectedUsers", adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.User), adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter))
}

func (s *Resources) ListADGPOAffectedSites(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGPOAffectedSites", adAnalysis.FetchGPOAffectedSitePaths, adAnalysis.FetchGPOAffectedSites)
}

func (s *Resources) ListADGPOAffectedComputers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGPOAffectedComputers", adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.Computer), adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectComputersCandidateFilter))
}

func (s *Resources) ListADGPOAffectedTierZero(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADGPOAffectedTierZero", adAnalysis.FetchGPOAffectedTierZeroPathDelegate, adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter))
}

func (s *Resources) ListADIssuancePolicyLinkedCertTemplates(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADIssuancePolicyLinkedCertTemplates", adAnalysis.FetchPolicyLinkedCertTemplatePaths, adAnalysis.FetchPolicyLinkedCertTemplates)
}

func (s *Resources) ListRootCAPKIHierarchy(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListRootCAPKIHierarchy", adAnalysis.CreateRootCAPKIHierarchyPathDelegate, adAnalysis.CreateRootCAPKIHierarchyListDelegate)
}

func (s *Resources) ListCAPKIHierarchy(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListCAPKIHierarchy", adAnalysis.CreateCAPKIHierarchyPathDelegate, adAnalysis.CreateCAPKIHierarchyListDelegate)
}

func (s *Resources) ListPublishedTemplates(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListPublishedTemplates", adAnalysis.CreatePublishedTemplatesPathDelegate, adAnalysis.CreatePublishedTemplatesListDelegate)
}

func (s *Resources) ListPublishedToCAs(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListPublishedToCAs", adAnalysis.CreatePublishedToCAsPathDelegate, adAnalysis.CreatePublishedToCAsListDelegate)
}

func (s *Resources) ListTrustedCAs(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListTrustedCAs", adAnalysis.CreateTrustedCAsPathDelegate, adAnalysis.CreateTrustedCAsListDelegate)
}

func (s *Resources) ListADCSEscalations(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADCSEscalations", adAnalysis.CreateADCSEscalationsPathDelegate, adAnalysis.CreateADCSEscalationsListDelegate)
}

func (s *Resources) ListADSiteLinkedServers(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADSiteLinkedServers", adAnalysis.CreateContainedPathDelegate(ad.SiteServer), adAnalysis.CreateContainedListDelegate(ad.SiteServer))
}

func (s *Resources) ListADSiteLinkedGPOs(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADSiteLinkedGPOs", adAnalysis.FetchEntityLinkedGPOPaths, adAnalysis.FetchEntityLinkedGPOList)
}

func (s *Resources) ListADSiteLinkedSubnets(response http.ResponseWriter, request *http.Request) {
	s.handleAdRelatedEntityQuery(response, request, "ListADSiteLinkedSubnets", adAnalysis.CreateContainedPathDelegate(ad.SiteSubnet), adAnalysis.CreateContainedListDelegate(ad.SiteSubnet))
}
