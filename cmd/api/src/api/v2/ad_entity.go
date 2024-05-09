// Copyright 2024 Specter Ops, Inc.
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
	"fmt"
	"net/http"

	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/api"
)

type DomainPatchRequest struct {
	Collected *bool `json:"collected"`
}

func (s *Resources) PatchDomain(response http.ResponseWriter, request *http.Request) {
	var domainPatchReq DomainPatchRequest
	if err := api.ReadJSONRequestPayloadLimited(&domainPatchReq, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if domainPatchReq.Collected == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "no domain fields sent for patching", request), response)
	} else if objectId, err := GetEntityObjectIDFromRequestPath(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("error reading objectid: %v", err), request), response)
	} else if node, err := s.GraphQuery.GetEntityByObjectId(request.Context(), objectId, ad.Domain); err != nil {
		if graph.IsErrNotFound(err) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "node not found", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error getting node: %v", err), request), response)
		}
	} else {
		node.Properties.Set(common.Collected.String(), *domainPatchReq.Collected)

		nodeUpdate := graph.NodeUpdate{
			Node:               node,
			IdentityKind:       ad.Domain,
			IdentityProperties: []string{common.ObjectID.String()},
		}

		if err := s.GraphQuery.BatchNodeUpdate(request.Context(), nodeUpdate); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error updating node: %v", err), request), response)
		} else {
			api.WriteBasicResponse(request.Context(), map[string]bool{"collected": *domainPatchReq.Collected}, http.StatusOK, response)
		}
	}
}

func (s *Resources) handleAdEntityInfoQuery(response http.ResponseWriter, request *http.Request, entityType graph.Kind, countQueries map[string]any) {
	if hydrateCounts, err := api.ParseOptionalBool(request.URL.Query().Get(api.QueryParameterHydrateCounts), true); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if objectId, err := GetEntityObjectIDFromRequestPath(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("error reading objectid: %v", err), request), response)
	} else if node, err := s.GraphQuery.GetEntityByObjectId(request.Context(), objectId, entityType); err != nil {
		if graph.IsErrNotFound(err) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "node not found", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error getting node: %v", err), request), response)
		}
	} else if hydrateCounts {
		results := s.GraphQuery.GetEntityCountResults(request.Context(), node, countQueries)
		api.WriteBasicResponse(request.Context(), results, http.StatusOK, response)
	} else {
		results := map[string]any{"props": node.Properties.Map}
		api.WriteBasicResponse(request.Context(), results, http.StatusOK, response)
	}
}

func (s *Resources) GetBaseEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllables": adAnalysis.FetchOutboundADEntityControl,
		}
	)
	s.handleAdEntityInfoQuery(response, request, ad.Entity, countQueries)
}

func (s *Resources) GetComputerEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"sessions":         adAnalysis.FetchComputerSessions,
			"adminUsers":       adAnalysis.CreateInboundLocalGroupListDelegate(ad.AdminTo),
			"rdpUsers":         adAnalysis.CreateInboundLocalGroupListDelegate(ad.CanRDP),
			"dcomUsers":        adAnalysis.CreateInboundLocalGroupListDelegate(ad.ExecuteDCOM),
			"psRemoteUsers":    adAnalysis.CreateInboundLocalGroupListDelegate(ad.CanPSRemote),
			"sqlAdminUsers":    adAnalysis.CreateSQLAdminListDelegate(graph.DirectionInbound),
			"constrainedUsers": adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionInbound),
			"groupMembership":  adAnalysis.FetchEntityGroupMembership,
			"adminRights":      adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo),
			"rdpRights":        adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanRDP),
			"dcomRights":       adAnalysis.CreateOutboundLocalGroupListDelegate(ad.ExecuteDCOM),
			"psRemoteRights":   adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanPSRemote),
			"constrainedPrivs": adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionOutbound),
			"controllables":    adAnalysis.FetchOutboundADEntityControl,
			"controllers":      adAnalysis.FetchInboundADEntityControllers,
			"gpos":             adAnalysis.FetchEnforcedGPOs,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.Computer, countQueries)
}

func (s *Resources) GetContainerEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.Container, countQueries)
}

func (s *Resources) GetDomainEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"users":                 adAnalysis.CreateDomainContainedEntityListDelegate(ad.User),
			"groups":                adAnalysis.CreateDomainContainedEntityListDelegate(ad.Group),
			"computers":             adAnalysis.CreateDomainContainedEntityListDelegate(ad.Computer),
			"ous":                   adAnalysis.CreateDomainContainedEntityListDelegate(ad.OU),
			"gpos":                  adAnalysis.CreateDomainContainedEntityListDelegate(ad.GPO),
			"foreignUsers":          adAnalysis.CreateForeignEntityMembershipListDelegate(ad.User),
			"foreignGroups":         adAnalysis.CreateForeignEntityMembershipListDelegate(ad.Group),
			"foreignGPOControllers": adAnalysis.FetchForeignGPOControllers,
			"foreignAdmins":         adAnalysis.FetchForeignAdmins,
			"inboundTrusts":         adAnalysis.CreateDomainTrustListDelegate(graph.DirectionInbound),
			"outboundTrusts":        adAnalysis.CreateDomainTrustListDelegate(graph.DirectionOutbound),
			"controllers":           adAnalysis.FetchInboundADEntityControllers,
			"linkedgpos":            adAnalysis.FetchEntityLinkedGPOList,
			"dcsyncers":             adAnalysis.FetchDCSyncers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.Domain, countQueries)
}

func (s *Resources) GetGPOEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"ous":         adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter),
			"computers":   adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectComputersCandidateFilter),
			"users":       adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter),
			"controllers": adAnalysis.FetchInboundADEntityControllers,
			"tierzero":    adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter),
		}
	)
	s.handleAdEntityInfoQuery(response, request, ad.GPO, countQueries)
}

func (s *Resources) GetAIACAEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.AIACA, countQueries)
}

func (s *Resources) GetRootCAEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.RootCA, countQueries)
}

func (s *Resources) GetEnterpriseCAEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)
	s.handleAdEntityInfoQuery(response, request, ad.EnterpriseCA, countQueries)
}

func (s *Resources) GetNTAuthStoreEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)
	s.handleAdEntityInfoQuery(response, request, ad.NTAuthStore, countQueries)
}

func (s *Resources) GetCertTemplateEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.CertTemplate, countQueries)
}

func (s *Resources) GetOUEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"gpos":      adAnalysis.FetchEntityLinkedGPOList,
			"users":     adAnalysis.CreateOUContainedListDelegate(ad.User),
			"groups":    adAnalysis.CreateOUContainedListDelegate(ad.Group),
			"computers": adAnalysis.CreateOUContainedListDelegate(ad.Computer),
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.OU, countQueries)
}

func (s *Resources) GetUserEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"sessions":              adAnalysis.FetchUserSessions,
			"groupMembership":       adAnalysis.FetchEntityGroupMembership,
			"adminRights":           adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo),
			"rdpRights":             adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanRDP),
			"dcomRights":            adAnalysis.CreateOutboundLocalGroupListDelegate(ad.ExecuteDCOM),
			"psRemoteRights":        adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanPSRemote),
			"sqlAdmin":              adAnalysis.CreateSQLAdminListDelegate(graph.DirectionOutbound),
			"constrainedDelegation": adAnalysis.CreateConstrainedDelegationListDelegate(graph.DirectionOutbound),
			"controllables":         adAnalysis.FetchOutboundADEntityControl,
			"controllers":           adAnalysis.FetchInboundADEntityControllers,
			"gpos":                  adAnalysis.FetchEnforcedGPOs,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.User, countQueries)
}

func (s *Resources) GetGroupEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"sessions":       adAnalysis.FetchGroupSessions,
			"members":        adAnalysis.FetchGroupMembers,
			"membership":     adAnalysis.FetchEntityGroupMembership,
			"adminRights":    adAnalysis.CreateOutboundLocalGroupListDelegate(ad.AdminTo),
			"rdpRights":      adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanRDP),
			"dcomRights":     adAnalysis.CreateOutboundLocalGroupListDelegate(ad.ExecuteDCOM),
			"psRemoteRights": adAnalysis.CreateOutboundLocalGroupListDelegate(ad.CanPSRemote),
			"controllables":  adAnalysis.FetchOutboundADEntityControl,
			"controllers":    adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.Group, countQueries)
}

func (s *Resources) GetIssuancePolicyEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		countQueries = map[string]any{
			"controllers":     adAnalysis.FetchInboundADEntityControllers,
			"linkedTemplates": adAnalysis.FetchPolicyLinkedCertTemplates,
		}
	)

	s.handleAdEntityInfoQuery(response, request, ad.IssuancePolicy, countQueries)
}
