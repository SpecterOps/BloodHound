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
	"net/http"

	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/queries"
)

// handleAdRelatedEntityQuery is for retrieving entities related to a specific entity.
// This endpoint returns a polymorphic response dependent upon which return type is requested
// by the `type` parameter, which can be `list`, `count`, or `graph`.
// Path delegates are for graphing, list delegates are for listing and counting. Endpoints
// without a certain delegate do not support that delegate feature.
func (s *Resources) handleAdRelatedEntityQuery(response http.ResponseWriter, request *http.Request, queryName string, pathDelegate any, listDelegate any) {
	if params, err := queries.BuildEntityQueryParams(request, queryName, pathDelegate, listDelegate); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.FmtErrorResponseDetailsBadQueryParameters, err), request), response)
	} else if entityPanelCachingFlag, err := s.DB.GetFlagByKey(appcfg.FeatureEntityPanelCaching); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if results, err := s.GraphQuery.GetADEntityQueryResult(request.Context(), params, entityPanelCachingFlag.Enabled); err != nil {
		if errors.Is(err, queries.ErrGraphUnsupported) || errors.Is(err, queries.ErrUnsupportedDataType) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.FmtErrorResponseDetailsBadQueryParameters, err), request), response)
		} else if errors.Is(err, ops.ErrTraversalMemoryLimit) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "calculating the request results exceeded memory limitations due to the volume of objects involved", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "an unknown error occurred during the request", request), response)
		}
	} else {
		api.WriteJSONResponse(request.Context(), results, http.StatusOK, response)
	}
}

func (s *Resources) ListADUserSessions(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADUserSessions"
		pathDelegate = adAnalysis.FetchUserSessionPaths
		listDelegate = adAnalysis.FetchUserSessions
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADUserSQLAdminRights(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADUserSQLAdminRights"
		pathDelegate = adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionOutbound)
		listDelegate = adAnalysis.CreateSQLAdminListDelegate(graph.DirectionOutbound)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADGroupSessions(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADGroupSessions"
		pathDelegate = adAnalysis.FetchGroupSessionPaths
		listDelegate = adAnalysis.FetchGroupSessions
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADComputerSessions(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADComputerSessions"
		pathDelegate = adAnalysis.FetchComputerSessionPaths
		listDelegate = adAnalysis.FetchComputerSessions
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADComputerAdmins(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADComputerAdmins"
		pathDelegate = adAnalysis.CreateInboundLocalGroupPathDelegate(ad.AdminTo)
		listDelegate = adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADComputerPSRemoteUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADComputerPSRemoteUsers"
		pathDelegate = adAnalysis.CreateInboundLocalGroupPathDelegate(ad.CanPSRemote)
		listDelegate = adAnalysis.CreateInboundLocalGroupListDelegate(ad.CanPSRemote)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADComputerRDPUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADComputerRDPUsers"
		pathDelegate = adAnalysis.CreateInboundLocalGroupPathDelegate(ad.CanRDP)
		listDelegate = adAnalysis.CreateInboundLocalGroupListDelegate(ad.CanRDP)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADComputerDCOMUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADComputerDCOMUsers"
		pathDelegate = adAnalysis.CreateInboundLocalGroupPathDelegate(ad.ExecuteDCOM)
		listDelegate = adAnalysis.CreateInboundLocalGroupListDelegate(ad.ExecuteDCOM)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADGroupMembership(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADGroupMembership"
		pathDelegate = adAnalysis.FetchEntityGroupMembershipPaths
		listDelegate = adAnalysis.FetchEntityGroupMembership
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADGroupMembers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADGroupMembers"
		pathDelegate = adAnalysis.FetchGroupMemberPaths
		listDelegate = adAnalysis.FetchGroupMembers
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADComputerSQLAdmins(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADComputerSQLAdmins"
		pathDelegate = adAnalysis.CreateSQLAdminPathDelegate(graph.DirectionInbound)
		listDelegate = adAnalysis.CreateSQLAdminListDelegate(graph.DirectionInbound)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADComputerConstrainedDelegationUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADComputerConstrainedDelegationUsers"
		pathDelegate = adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionInbound)
		listDelegate = adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionInbound)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityConstrainedDelegationRights(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityConstrainedDelegationRights"
		pathDelegate = adAnalysis.CreateConstrainedDelegationPathDelegate(graph.DirectionOutbound)
		listDelegate = adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionOutbound)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityAdminRights(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityAdminRights"
		pathDelegate = adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.AdminTo)
		listDelegate = adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityRDPRights(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityRDPRights"
		pathDelegate = adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.CanRDP)
		listDelegate = adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanRDP)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityPSRemoteRights(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityPSRemoteRights"
		pathDelegate = adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.CanPSRemote)
		listDelegate = adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanPSRemote)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityDCOMRights(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityDCOMRights"
		pathDelegate = adAnalysis.CreateOutboundLocalGroupPathDelegate(ad.ExecuteDCOM)
		listDelegate = adAnalysis.CreateOutboundLocalGroupListDelegate(ad.ExecuteDCOM)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityControllers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityControllers"
		pathDelegate = adAnalysis.FetchInboundADEntityControllerPaths
		listDelegate = adAnalysis.FetchInboundADEntityControllers
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityControllables(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityControllables"
		pathDelegate = adAnalysis.FetchOutboundADEntityControlPaths
		listDelegate = adAnalysis.FetchOutboundADEntityControl
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADEntityLinkedGPOs(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADEntityLinkedGPOs"
		pathDelegate = adAnalysis.FetchEntityLinkedGPOPaths
		listDelegate = adAnalysis.FetchEntityLinkedGPOList
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainContainedUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName        = "ListADDomainContainedUsers"
		pathDelegate any = nil
		listDelegate     = adAnalysis.CreateDomainContainedEntityListDelegate(ad.User)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainContainedComputers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName        = "ListADDomainContainedComputers"
		pathDelegate any = nil
		listDelegate     = adAnalysis.CreateDomainContainedEntityListDelegate(ad.Computer)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainContainedGroups(response http.ResponseWriter, request *http.Request) {
	var (
		queryName        = "ListADDomainContainedGroups"
		pathDelegate any = nil
		listDelegate     = adAnalysis.CreateDomainContainedEntityListDelegate(ad.Group)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainContainedOUs(response http.ResponseWriter, request *http.Request) {
	var (
		queryName        = "ListADDomainContainedOUs"
		pathDelegate any = nil
		listDelegate     = adAnalysis.CreateDomainContainedEntityListDelegate(ad.OU)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainContainedGPOs(response http.ResponseWriter, request *http.Request) {
	var (
		queryName        = "ListADDomainContainedGPOs"
		pathDelegate any = nil
		listDelegate     = adAnalysis.CreateDomainContainedEntityListDelegate(ad.GPO)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainForeignGroups(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADDomainForeignGroups"
		pathDelegate = adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.Group)
		listDelegate = adAnalysis.CreateForeignEntityMembershipListDelegate(ad.Group)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainForeignUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADDomainForeignUsers"
		pathDelegate = adAnalysis.CreateForeignEntityMembershipPathDelegate(ad.User)
		listDelegate = adAnalysis.CreateForeignEntityMembershipListDelegate(ad.User)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainForeignAdmins(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADDomainForeignAdmins"
		pathDelegate = adAnalysis.FetchForeignAdminPaths
		listDelegate = adAnalysis.FetchForeignAdmins
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainForeignGPOControllers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADDomainForeignGPOControllers"
		pathDelegate = adAnalysis.FetchForeignGPOControllerPaths
		listDelegate = adAnalysis.FetchForeignGPOControllers
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainOutboundTrusts(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADDomainOutboundTrusts"
		pathDelegate = adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionOutbound)
		listDelegate = adAnalysis.CreateDomainTrustListDelegate(graph.DirectionOutbound)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainInboundTrusts(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADDomainInboundTrusts"
		pathDelegate = adAnalysis.CreateDomainTrustPathDelegate(graph.DirectionInbound)
		listDelegate = adAnalysis.CreateDomainTrustListDelegate(graph.DirectionInbound)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADDomainDCSyncers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADDomainDCSyncers"
		pathDelegate = adAnalysis.FetchDCSyncerPaths
		listDelegate = adAnalysis.FetchDCSyncers
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADOUContainedUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADOUContainedUsers"
		pathDelegate = adAnalysis.CreateOUContainedPathDelegate(ad.User)
		listDelegate = adAnalysis.CreateOUContainedListDelegate(ad.User)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADOUContainedGroups(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADOUContainedGroups"
		pathDelegate = adAnalysis.CreateOUContainedPathDelegate(ad.Group)
		listDelegate = adAnalysis.CreateOUContainedListDelegate(ad.Group)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADOUContainedComputers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADOUContainedComputers"
		pathDelegate = adAnalysis.CreateOUContainedPathDelegate(ad.Computer)
		listDelegate = adAnalysis.CreateOUContainedListDelegate(ad.Computer)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADGPOAffectedContainers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADGPOAffectedContainers"
		pathDelegate = adAnalysis.FetchGPOAffectedContainerPaths
		listDelegate = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADGPOAffectedUsers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADGPOAffectedUsers"
		pathDelegate = adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.User)
		listDelegate = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADGPOAffectedComputers(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADGPOAffectedComputers"
		pathDelegate = adAnalysis.CreateGPOAffectedIntermediariesPathDelegate(ad.Computer)
		listDelegate = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectComputersCandidateFilter)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}

func (s *Resources) ListADGPOAffectedTierZero(response http.ResponseWriter, request *http.Request) {
	var (
		queryName    = "ListADGPOAffectedTierZero"
		pathDelegate = adAnalysis.FetchGPOAffectedTierZeroPathDelegate
		listDelegate = adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter)
	)
	s.handleAdRelatedEntityQuery(response, request, queryName, pathDelegate, listDelegate)
}
