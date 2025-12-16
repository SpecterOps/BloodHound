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
	"time"

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
	return model.UnifiedGraphWPropertyKeys{NodeKeys: nSlice, EdgeKeys: eSlice, Edges: graphResponse.Edges, Nodes: graphResponse.Nodes}
}

func (s Resources) CypherQuery(response http.ResponseWriter, request *http.Request) {
	var (
		payload       CypherQueryPayload
		preparedQuery queries.PreparedQuery
		graphResponse model.UnifiedGraph
		err           error
	)

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
	filteredResponse, err := s.filterETACGraph(graphResponse, request)
	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "error", request), response)
		return
	}
	if !preparedQuery.HasMutation && len(filteredResponse.Nodes)+len(filteredResponse.Edges) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "resource not found", request), response)
		return
	}

	if !payload.IncludeProperties {
		// removing node properties from the response
		for id, node := range filteredResponse.Nodes {
			node.Properties = nil
			filteredResponse.Nodes[id] = node
		}

		for i, edge := range filteredResponse.Edges {
			edge.Properties = nil
			filteredResponse.Edges[i] = edge
		}

		api.WriteBasicResponse(request.Context(), filteredResponse, http.StatusOK, response)
		return
	}

	api.WriteBasicResponse(request.Context(), processCypherProperties(filteredResponse), http.StatusOK, response)

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

func (s Resources) filterETACGraph(graphResponse model.UnifiedGraph, request *http.Request) (model.UnifiedGraph, error) {
	accessList, shouldFilter, err := ShouldFilterForETAC(request, s.DB)
	if err != nil {
		slog.Error("Unable to check ETAC filtering")
		return model.UnifiedGraph{}, err
	}

	if !shouldFilter {
		return graphResponse, nil
	}

	filteredResponse := model.UnifiedGraph{}
	filteredNodes := make(map[string]model.UnifiedNode)

	environmentKeys := []string{"domainsid", "tenantid"}

	for id, node := range graphResponse.Nodes {
		include := false
		for _, key := range environmentKeys {
			if val, ok := node.Properties[key]; ok {
				if envStr, ok := val.(string); ok && slices.Contains(accessList, envStr) {
					include = true
					break
				}
			}
		}

		if include {
			filteredNodes[id] = node
		} else {
			var kind string
			if len(node.Kinds) > 0 && node.Kinds[0] != "" {
				kind = node.Kinds[0]
			} else {
				kind = "Unknown"
			}

			label := fmt.Sprintf("** Hidden %s Object **", kind)
			filteredNodes[id] = model.UnifiedNode{
				Label:         label,
				Kind:          "HIDDEN",
				Kinds:         []string{},
				ObjectId:      "HIDDEN",
				IsTierZero:    false,
				IsOwnedObject: false,
				LastSeen:      time.Time{},
				Properties:    nil,
				Hidden:        true,
			}
		}
	}

	filteredResponse.Nodes = filteredNodes
	filteredEdges := make([]model.UnifiedEdge, 0, len(graphResponse.Edges))

	for _, edge := range graphResponse.Edges {
		if filteredNodes[edge.Target].Hidden || filteredNodes[edge.Source].Hidden {
			filteredEdges = append(filteredEdges, model.UnifiedEdge{
				Source:     edge.Source,
				Target:     edge.Target,
				Label:      "** Hidden Edge **",
				Kind:       "HIDDEN",
				LastSeen:   time.Time{},
				Properties: nil,
			})
		} else {
			filteredEdges = append(filteredEdges, edge)
		}
	}
	filteredResponse.Edges = filteredEdges

	return filteredResponse, nil
}
