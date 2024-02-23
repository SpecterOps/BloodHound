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
		entityType   = ad.Entity
		countQueries = map[string]any{
			"controllables": adAnalysis.FetchOutboundADEntityControl,
		}
	)
	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetComputerEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.Computer
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

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetContainerEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.Container
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetDomainEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.Domain
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

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetGPOEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.GPO
		countQueries = map[string]any{
			"ous":         adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter),
			"computers":   adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectComputersCandidateFilter),
			"users":       adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter),
			"controllers": adAnalysis.FetchInboundADEntityControllers,
			"tierzero":    adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter),
		}
	)
	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetAIACAEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.AIACA
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetRootCAEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.RootCA
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetEnterpriseCAEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.EnterpriseCA
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)
	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetNTAuthStoreEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.NTAuthStore
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)
	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetCertTemplateEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.CertTemplate
		countQueries = map[string]any{
			"controllers": adAnalysis.FetchInboundADEntityControllers,
		}
	)

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetOUEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.OU
		countQueries = map[string]any{
			"gpos":      adAnalysis.FetchEntityLinkedGPOList,
			"users":     adAnalysis.CreateOUContainedListDelegate(ad.User),
			"groups":    adAnalysis.CreateOUContainedListDelegate(ad.Group),
			"computers": adAnalysis.CreateOUContainedListDelegate(ad.Computer),
		}
	)

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetUserEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.User
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

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}

func (s *Resources) GetGroupEntityInfo(response http.ResponseWriter, request *http.Request) {
	var (
		entityType   = ad.Group
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

	s.handleAdEntityInfoQuery(response, request, entityType, countQueries)
}
