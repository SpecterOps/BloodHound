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
	"maps"
	"net/http"
	"slices"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/dawgs/util"
)

var (
	errUnauthorizedGraphMutation = errors.New("unauthorized graph mutation")
)

type CypherQueryPayload struct {
	Query             string `json:"query"`
	IncludeProperties bool   `json:"include_properties,omitempty"`
}

// Helper function to handle error conditions in CypherQuery.
func handleCypherDBErrors(response http.ResponseWriter, request *http.Request, err error) {
	if errors.Is(err, errUnauthorizedGraphMutation) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "Permission denied: User may not modify the graph.", request), response)
	} else if util.IsNeoTimeoutError(err) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "transaction timed out, reduce query complexity or try again later", request), response)
	} else {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	}
}

// Helper function to handle processing of property keys.
func processCypherProperties(graphResponse model.UnifiedGraph) model.UnifiedGraphWPropertyKeys {
	eKeys := map[string]struct{}{}
	nKeys := map[string]struct{}{}
	for _, node := range graphResponse.Nodes {
		for key := range node.Properties {
			nKeys[key] = struct{}{}
		}
	}
	for _, edge := range graphResponse.Edges {
		for key := range edge.Properties {
			eKeys[key] = struct{}{}
		}
	}
	eSlice := slices.Sorted(maps.Keys(eKeys))
	nSlice := slices.Sorted(maps.Keys(nKeys))
	return model.UnifiedGraphWPropertyKeys{
		NodeKeys: nSlice,
		EdgeKeys: eSlice,
		Edges:    graphResponse.Edges,
		Nodes:    graphResponse.Nodes,
		Literals: graphResponse.Literals,
	}
}

func (s Resources) CypherQuery(response http.ResponseWriter, request *http.Request) {
	var (
		payload       CypherQueryPayload
		preparedQuery queries.PreparedQuery
		graphResponse model.UnifiedGraph
		err           error
	)

	user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
		return
	}

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed.", request), response)
		return
	}

	if preparedQuery, err = s.GraphQuery.PrepareCypherQuery(payload.Query, queries.DefaultQueryFitnessLowerBoundExplore); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	if preparedQuery.HasMutation {
		// defaulting include properties to true so ETAC filtering logic has access to node properties
		graphResponse, err = s.cypherMutation(request, preparedQuery, true)
	} else {
		// defaulting include properties to true so ETAC filtering logic has access to node properties
		graphResponse, err = s.GraphQuery.RawCypherQuery(request.Context(), preparedQuery, true)
	}

	if err != nil {
		handleCypherDBErrors(response, request, err)
		return
	}

	// Etac DogTags
	if ShouldFilterForETAC(s.DogTags, user) {
		filteredResponse, err := filterETACGraph(graphResponse, user)
		if err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "error filtering graph for ETAC", request), response)
			return
		}
		graphResponse = filteredResponse
	}

	if !preparedQuery.HasMutation && len(graphResponse.Nodes)+len(graphResponse.Edges)+len(graphResponse.Literals) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "resource not found", request), response)
		return
	}

	if !payload.IncludeProperties {
		// removing node properties from the response
		for id, node := range graphResponse.Nodes {
			node.Properties = nil
			graphResponse.Nodes[id] = node
		}
		// removing edge properties from the response
		for i, edge := range graphResponse.Edges {
			edge.Properties = nil
			graphResponse.Edges[i] = edge
		}

		api.WriteBasicResponse(request.Context(), graphResponse, http.StatusOK, response)
		return
	} else {
		api.WriteBasicResponse(request.Context(), processCypherProperties(graphResponse), http.StatusOK, response)
	}
}

func (s Resources) cypherMutation(request *http.Request, preparedQuery queries.PreparedQuery, includeProperties bool) (model.UnifiedGraph, error) {
	var (
		auditLogEntry model.AuditEntry
		graphResponse model.UnifiedGraph
		err           error
	)

	if !s.Authorizer.AllowsPermission(ctx.FromRequest(request).AuthCtx, auth.Permissions().GraphDBMutate) {
		s.Authorizer.AuditLogUnauthorizedAccess(request)
		return model.UnifiedGraph{}, errUnauthorizedGraphMutation
	}

	// All mutation attempts must be audit logged even when failed
	if auditLogEntry, err = model.NewAuditEntry(model.AuditLogActionMutateGraph, model.AuditLogStatusIntent, model.AuditData{"query": preparedQuery.StrippedQuery}); err != nil {
		return model.UnifiedGraph{}, err
	}

	// create an intent audit log
	if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
		return model.UnifiedGraph{}, err
	}

	if graphResponse, err = s.GraphQuery.RawCypherQuery(request.Context(), preparedQuery, includeProperties); err != nil {
		auditLogEntry.Status = model.AuditLogStatusFailure
	} else {
		auditLogEntry.Status = model.AuditLogStatusSuccess
	}

	if err := s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
		// We want to keep err scoped because having info on the mutation graph response trumps this error
		slog.ErrorContext(request.Context(), fmt.Sprintf("failure to create mutation audit log %s", err.Error()))
	}

	return graphResponse, err

}
