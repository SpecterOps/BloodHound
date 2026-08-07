// Copyright 2026 Specter Ops, Inc.
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
	"regexp"
	"strings"

	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/dawgs/graph"
)

const (
	defaultGraphExpansionLimit = 500
	maxGraphExpansionLimit     = 1000

	graphExpansionDirectionInbound  = "inbound"
	graphExpansionDirectionOutbound = "outbound"
	graphExpansionBuiltInKinds      = "ALL_ATTACK_PATHS"
)

var graphExpansionRelationshipKindPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type GraphExpansionPayload struct {
	NodeID            *int64 `json:"node_id"`
	Direction         string `json:"direction"`
	Limit             int    `json:"limit,omitempty"`
	IncludeProperties bool   `json:"include_properties,omitempty"`
}

type GraphExpansionResponse struct {
	NodeKeys  []string                     `json:"node_keys,omitempty"`
	EdgeKeys  []string                     `json:"edge_keys,omitempty"`
	Nodes     map[string]model.UnifiedNode `json:"nodes"`
	Edges     []model.UnifiedEdge          `json:"edges"`
	Literals  graph.Literals               `json:"literals"`
	Limit     int                          `json:"limit"`
	Truncated bool                         `json:"truncated"`
}

func graphExpansionLimit(requestedLimit int) (int, error) {
	if requestedLimit == 0 {
		return defaultGraphExpansionLimit, nil
	}

	if requestedLimit < 0 {
		return 0, fmt.Errorf("limit must be greater than 0")
	}

	if requestedLimit > maxGraphExpansionLimit {
		return 0, fmt.Errorf("limit must be less than or equal to %d", maxGraphExpansionLimit)
	}

	return requestedLimit, nil
}

func buildGraphExpansionQuery(nodeID int64, direction string, relationshipKinds []string, limit int) (string, error) {
	filteredRelationshipKinds := make([]string, 0, len(relationshipKinds)+1)
	seenRelationshipKinds := map[string]struct{}{}

	for _, relationshipKind := range append([]string{graphExpansionBuiltInKinds}, relationshipKinds...) {
		if !graphExpansionRelationshipKindPattern.MatchString(relationshipKind) {
			return "", fmt.Errorf("invalid relationship kind: %s", relationshipKind)
		}

		if _, seen := seenRelationshipKinds[relationshipKind]; seen {
			continue
		}

		seenRelationshipKinds[relationshipKind] = struct{}{}
		filteredRelationshipKinds = append(filteredRelationshipKinds, relationshipKind)
	}

	relationshipMatch := ""
	relationshipFilter := strings.Join(filteredRelationshipKinds, "|")

	switch direction {
	case graphExpansionDirectionOutbound:
		relationshipMatch = fmt.Sprintf("MATCH (source)-[r:%s]->(target)", relationshipFilter)
	case graphExpansionDirectionInbound:
		relationshipMatch = fmt.Sprintf("MATCH (target)-[r:%s]->(source)", relationshipFilter)
	default:
		return "", fmt.Errorf("direction must be either inbound or outbound")
	}

	return fmt.Sprintf(`MATCH (source)
WHERE ID(source) = %d
%s
RETURN source, r, target
LIMIT %d`, nodeID, relationshipMatch, limit), nil
}

func pruneGraphExpansionResponse(graphResponse model.UnifiedGraph, limit int) (model.UnifiedGraph, bool) {
	truncated := len(graphResponse.Edges) > limit
	if truncated {
		graphResponse.Edges = graphResponse.Edges[:limit]
	}

	retainedNodeIDs := map[string]struct{}{}
	for _, edge := range graphResponse.Edges {
		retainedNodeIDs[edge.Source] = struct{}{}
		retainedNodeIDs[edge.Target] = struct{}{}
	}

	retainedNodes := make(map[string]model.UnifiedNode, len(retainedNodeIDs))
	for nodeID, node := range graphResponse.Nodes {
		if _, retain := retainedNodeIDs[nodeID]; retain {
			retainedNodes[nodeID] = node
		}
	}

	graphResponse.Nodes = retainedNodes

	return graphResponse, truncated
}

func clearGraphExpansionProperties(graphResponse model.UnifiedGraph) model.UnifiedGraph {
	for id, node := range graphResponse.Nodes {
		node.Properties = nil
		graphResponse.Nodes[id] = node
	}

	for i, edge := range graphResponse.Edges {
		edge.Properties = nil
		graphResponse.Edges[i] = edge
	}

	return graphResponse
}

func graphExpansionResponse(graphResponse model.UnifiedGraph, limit int, truncated bool, includeProperties bool) GraphExpansionResponse {
	if includeProperties {
		graphWithKeys := processCypherProperties(graphResponse)
		return GraphExpansionResponse{
			NodeKeys:  graphWithKeys.NodeKeys,
			EdgeKeys:  graphWithKeys.EdgeKeys,
			Nodes:     graphWithKeys.Nodes,
			Edges:     graphWithKeys.Edges,
			Literals:  graphWithKeys.Literals,
			Limit:     limit,
			Truncated: truncated,
		}
	}

	graphResponse = clearGraphExpansionProperties(graphResponse)
	return GraphExpansionResponse{
		Nodes:     graphResponse.Nodes,
		Edges:     graphResponse.Edges,
		Literals:  graphResponse.Literals,
		Limit:     limit,
		Truncated: truncated,
	}
}

func (s Resources) graphExpansionRelationshipKinds(request *http.Request) ([]string, *api.ErrorWrapper) {
	openGraphExtensionManagementFeatureFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureOpenGraphExtensionManagement)
	if err != nil {
		return nil, api.BuildErrorResponse(http.StatusInternalServerError, api.FormatDatabaseError(err).Error(), request)
	}

	if !openGraphExtensionManagementFeatureFlag.Enabled {
		return nil, nil
	}

	relationshipKindFilters := model.Filters{
		"is_traversable": []model.Filter{{Operator: model.Equals, Value: "true"}},
	}
	openGraphRelationships, _, err := s.DB.GetGraphSchemaRelationshipKinds(request.Context(), relationshipKindFilters, model.Sort{}, 0, 0)
	if err != nil {
		return nil, api.BuildErrorResponse(http.StatusInternalServerError, api.FormatDatabaseError(err).Error(), request)
	}

	relationshipKinds := make([]string, 0, len(openGraphRelationships))
	for _, relationship := range openGraphRelationships {
		relationshipKinds = append(relationshipKinds, relationship.Name)
	}

	return relationshipKinds, nil
}

func (s Resources) ExpandGraph(response http.ResponseWriter, request *http.Request) {
	var payload GraphExpansionPayload

	user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
		return
	}

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed.", request), response)
		return
	}

	if payload.NodeID == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "node_id is required", request), response)
		return
	}

	if *payload.NodeID < 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "node_id must be greater than or equal to 0", request), response)
		return
	}

	limit, err := graphExpansionLimit(payload.Limit)
	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	relationshipKinds, apiError := s.graphExpansionRelationshipKinds(request)
	if apiError != nil {
		api.WriteErrorResponse(request.Context(), apiError, response)
		return
	}

	query, err := buildGraphExpansionQuery(*payload.NodeID, payload.Direction, relationshipKinds, limit+1)
	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	preparedQuery, err := s.GraphQuery.PrepareCypherQuery(query, queries.DefaultQueryFitnessLowerBoundExplore)
	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	primaryDisplayKinds, err := s.DB.GetPrimaryDisplayKinds(request.Context())
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	graphResponse, err := s.GraphQuery.RawCypherQuery(request.Context(), primaryDisplayKinds, preparedQuery, true)
	if err != nil {
		handleCypherDBErrors(response, request, err)
		return
	}

	wasTruncated := len(graphResponse.Edges) > limit

	if ShouldFilterForETAC(s.DogTags, user) {
		filteredResponse, err := filterETACGraph(graphResponse, user)
		if err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "error filtering graph for ETAC", request), response)
			return
		}
		graphResponse = filteredResponse
	}

	prunedGraph, truncatedAfterFiltering := pruneGraphExpansionResponse(graphResponse, limit)

	api.WriteBasicResponse(request.Context(), graphExpansionResponse(prunedGraph, limit, wasTruncated || truncatedAfterFiltering, payload.IncludeProperties), http.StatusOK, response)
}
