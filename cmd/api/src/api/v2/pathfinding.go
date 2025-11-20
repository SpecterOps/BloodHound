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
	"fmt"
	"net/http"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

// deprecated - use GetShortestPath instead
func (s Resources) GetPathfindingResult(response http.ResponseWriter, request *http.Request) {
	var (
		params            = request.URL.Query()
		startNodeObjectID = params.Get("start_node")
		endNodeObjectID   = params.Get("end_node")
	)

	if startNodeObjectID == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Missing query parameter: start_node", request), response)
	} else if endNodeObjectID == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Missing query parameter: end_node", request), response)
	} else if paths, err := s.GraphQuery.GetAllShortestPaths(request.Context(), startNodeObjectID, endNodeObjectID, nil); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %v", err), request), response)
	} else {
		api.WriteBasicResponse(request.Context(), bloodhoundgraph.PathSetToBloodHoundGraph(paths), http.StatusOK, response)
	}
}

func writeShortestPathsResult(paths graph.PathSet, response http.ResponseWriter, request *http.Request) {
	if paths.Len() == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "Path not found", request), response)
	} else {
		graphResponse := model.NewUnifiedGraph()

		for _, n := range paths.AllNodes() {
			graphResponse.Nodes[n.ID.String()] = model.FromDAWGSNode(n, false)
		}

		edges := slicesext.FlatMap(paths, func(path graph.Path) []model.UnifiedEdge {
			return slicesext.Map(path.Edges, model.FromDAWGSRelationship(false))
		})

		graphResponse.Edges = slicesext.UniqueBy(edges, func(edge model.UnifiedEdge) string {
			return edge.Source + edge.Kind + edge.Target
		})

		api.WriteBasicResponse(request.Context(), graphResponse, http.StatusOK, response)
	}
}

func parseRelationshipKindsParam(validKinds graph.Kinds, relationshipKindsParam string) (graph.Kinds, string, error) {
	if relationshipKindsParam != "" && !params.RelationshipKinds.Regexp().MatchString(relationshipKindsParam) {
		return nil, "", fmt.Errorf("invalid query parameter 'relationship_kinds': acceptable values should match the format: in|nin:Kind1,Kind2")
	}

	// To get a slice of all requested kinds as strings,
	// 1. fetch the relationship_kinds query param as a substring after the separator ":"
	// 2. then remove all whitespace
	// 3. and then split them using commas as delimiters
	separatorIndex := strings.Index(relationshipKindsParam, ":")

	// If the separator ":" does not exist then it will return -1, thereby skipping the filtering altogether
	if relationshipKindsSlice := strings.Split(strings.ReplaceAll(relationshipKindsParam[separatorIndex+1:], " ", ""), ","); separatorIndex > 0 && len(relationshipKindsSlice) > 0 {
		var (
			op             = relationshipKindsParam[0:separatorIndex]
			parameterKinds = make(graph.Kinds, len(relationshipKindsSlice))
		)

		for idx, kindStr := range relationshipKindsSlice {
			kind := graph.StringKind(kindStr)

			if !validKinds.ContainsOneOf(kind) {
				return nil, "", fmt.Errorf("invalid query parameter 'relationship_kinds': acceptable relationship kinds are: %v", validKinds.Strings())
			}

			parameterKinds[idx] = kind
		}

		return parameterKinds, op, nil
	}

	// Default to all valid kinds, inclusive
	return validKinds, "in", nil
}

func parseRelationshipKindsParamFilter(relationshipKindsParam string) (graph.Criteria, error) {
	validKinds := graph.Kinds(ad.Relationships()).Concatenate(azure.Relationships())

	if filterKinds, filterOperation, err := parseRelationshipKindsParam(validKinds, relationshipKindsParam); err != nil {
		return nil, err
	} else if filterOperation == "in" {
		return query.KindIn(query.Relationship(), filterKinds...), nil
	} else {
		return query.KindIn(query.Relationship(), validKinds.Exclude(filterKinds)...), nil
	}
}

func (s Resources) GetShortestPath(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams            = request.URL.Query()
		startNode              = queryParams.Get(params.StartNode.String())
		endNode                = queryParams.Get(params.EndNode.String())
		relationshipKindsParam = queryParams.Get(params.RelationshipKinds.String())
	)

	if startNode == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Missing query parameter: start_node", request), response)
	} else if endNode == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Missing query parameter: end_node", request), response)
	} else if kindFilter, err := parseRelationshipKindsParamFilter(relationshipKindsParam); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if paths, err := s.GraphQuery.GetAllShortestPaths(request.Context(), startNode, endNode, kindFilter); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		writeShortestPathsResult(paths, response, request)
	}
}

const (
	searchParameterQuery = "query"
	searchParameterType  = "type"
)

func (s *Resources) GetSearchResult(response http.ResponseWriter, request *http.Request) {
	var params = request.URL.Query()

	if searchValues, hasParameter := params[searchParameterQuery]; !hasParameter {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Expected search parameter to be set.", request), response)
	} else if len(searchValues) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Expected only one search value.", request), response)
	} else {
		var (
			searchValue = searchValues[0]
			searchType  = queries.SearchTypeFuzzy
		)

		if searchTypeParameters, hasParameter := params[searchParameterType]; hasParameter {
			if len(searchTypeParameters) > 1 {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Expected only one search type.", request), response)
				return
			} else {
				searchType = strings.ToLower(searchTypeParameters[0])
			}
		}
		if openGraphSearchFeatureFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureOpenGraphSearch); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if nodes, err := s.GraphQuery.SearchByNameOrObjectID(request.Context(), openGraphSearchFeatureFlag.Enabled, searchValue, searchType); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error getting search results: %v", err), request), response)
		} else {
			api.WriteBasicResponse(request.Context(), bloodhoundgraph.NodeSetToBloodHoundGraph(nodes, openGraphSearchFeatureFlag.Enabled), http.StatusOK, response)
		}
	}
}
