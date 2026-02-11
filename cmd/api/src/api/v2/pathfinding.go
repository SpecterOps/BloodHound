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
	"slices"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/bloodhoundgraph"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

var errInvalidParameterCombination = errors.New("invalid parameter combination; no valid edges to traverse")

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

func writeShortestPathsResult(paths graph.PathSet, shouldFilterETAC bool, user model.User, response http.ResponseWriter, request *http.Request) {
	if paths.Len() == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "Path not found", request), response)
	} else {
		graphResponse := model.NewUnifiedGraph()

		for _, n := range paths.AllNodes() {
			// ETAC filtering requires pulling the node's properties
			graphResponse.Nodes[n.ID.String()] = model.FromDAWGSNode(n, true)
		}

		edges := slicesext.FlatMap(paths, func(path graph.Path) []model.UnifiedEdge {
			return slicesext.Map(path.Edges, model.FromDAWGSRelationship(false))
		})

		graphResponse.Edges = slicesext.UniqueBy(edges, func(edge model.UnifiedEdge) string {
			return edge.Source + edge.Kind + edge.Target
		})

		if shouldFilterETAC {
			if filteredGraph, err := filterETACGraph(graphResponse, user); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "error filtering graph for ETAC", request), response)
				return
			} else {
				graphResponse = filteredGraph
			}
		}

		// In order to filter nodes for ETAC, we need to grab the node's properties from DAWGs
		// This particular endpoint should not respond with properties, so we can simply clear them after pulling them
		newNodes := make(map[string]model.UnifiedNode)
		for key, node := range graphResponse.Nodes {
			node.Properties = make(map[string]any)
			newNodes[key] = node
		}

		graphResponse.Nodes = newNodes

		api.WriteBasicResponse(request.Context(), graphResponse, http.StatusOK, response)

	}
}

func parseRelationshipKindsParam(acceptableKinds graph.Kinds, relationshipKindsParam string) (graph.Kinds, string, error) {
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
			parameterKinds = make(graph.Kinds, 0)
		)

		for _, kindStr := range relationshipKindsSlice {
			kind := graph.StringKind(kindStr)
			if acceptableKinds.ContainsOneOf(kind) {
				parameterKinds = append(parameterKinds, kind)
			} else if !kindIsValidBuiltIn(kind) {
				return nil, "", fmt.Errorf("invalid query parameter 'relationship_kinds': acceptable relationship kinds are: %v", acceptableKinds.Strings())
			} // silently ignore kinds that are valid built-in kinds but not in the list of acceptable kinds
		}

		return parameterKinds, op, nil
	}

	// Default to all acceptable kinds, inclusive
	return acceptableKinds, "in", nil
}

// kindIsValidBuiltIn determines if a kind exists in the built-in graph
func kindIsValidBuiltIn(kind graph.Kind) bool {
	return graph.Kinds(ad.Relationships()).Concatenate(azure.Relationships()).ContainsOneOf(kind)
}

func createRelationshipKindFilterCriteria(relationshipKindsParam string, onlyIncludeTraversableKinds bool, validKinds graph.Kinds) (graph.Criteria, error) {
	var edgeKinds graph.Kinds
	if onlyIncludeTraversableKinds && relationshipKindsParam == "" {
		edgeKinds = validKinds
	} else if filterKinds, filterOperation, err := parseRelationshipKindsParam(validKinds, relationshipKindsParam); err != nil {
		return nil, err
	} else if filterOperation == "in" {
		edgeKinds = filterKinds
	} else {
		edgeKinds = validKinds.Exclude(filterKinds)
	}

	if len(edgeKinds) == 0 {
		return nil, errInvalidParameterCombination
	}

	return query.KindIn(query.Relationship(), edgeKinds...), nil

}

func (s Resources) GetShortestPath(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams            = request.URL.Query()
		startNode              = queryParams.Get(params.StartNode.String())
		endNode                = queryParams.Get(params.EndNode.String())
		relationshipKindsParam = queryParams.Get(params.RelationshipKinds.String())
		requestContext         = request.Context()
		paths                  graph.PathSet
		apiError               *api.ErrorWrapper
		validBuiltInKinds      = graph.Kinds(ad.Relationships()).Concatenate(azure.Relationships())
	)

	onlyIncludeTraversableKinds, err := api.ParseOptionalBool(request.URL.Query().Get(api.QueryParameterIncludeOnlyTraversableKinds), false)
	if err != nil {
		slog.ErrorContext(requestContext, "Error parsing optional boolean parameter", attr.Error(err))
	}
	if startNode == "" {
		api.WriteErrorResponse(requestContext, api.BuildErrorResponse(http.StatusBadRequest, "Missing query parameter: start_node", request), response)
		return
	} else if endNode == "" {
		api.WriteErrorResponse(requestContext, api.BuildErrorResponse(http.StatusBadRequest, "Missing query parameter: end_node", request), response)
		return
	} else if ogExtensionManagementFeatureFlag, err := s.DB.GetFlagByKey(requestContext, appcfg.FeatureOpenGraphExtensionManagement); err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	} else {
		if onlyIncludeTraversableKinds {
			validBuiltInKinds = graph.Kinds(ad.PathfindingRelationshipsMatchFrontend()).Concatenate(azure.PathfindingRelationships())
		}
		if ogExtensionManagementFeatureFlag.Enabled {
			if paths, apiError = s.getAllShortestPathsWithOpenGraph(requestContext, relationshipKindsParam, startNode, endNode, onlyIncludeTraversableKinds, validBuiltInKinds, request); apiError != nil {
				api.WriteErrorResponse(requestContext, apiError, response)
				return
			}
		} else {
			if paths, apiError = s.getAllShortestPaths(requestContext, relationshipKindsParam, startNode, endNode, onlyIncludeTraversableKinds, validBuiltInKinds, request); apiError != nil {
				api.WriteErrorResponse(requestContext, apiError, response)
				return
			}
		}
		user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx)
		if !isUser {
			slog.Error("Unable to get user from auth context")
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
		} else {
			writeShortestPathsResult(paths, ShouldFilterForETAC(s.DogTags, user), user, response, request)
		}
	}
}

func (s Resources) getAllShortestPaths(ctx context.Context, relationshipKindsParam, startNode, endNode string, onlyTraversable bool, validKinds graph.Kinds, request *http.Request) (graph.PathSet, *api.ErrorWrapper) {
	if kindFilter, err := createRelationshipKindFilterCriteria(relationshipKindsParam, onlyTraversable, validKinds); err != nil {
		return nil, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request)
	} else if paths, err := s.GraphQuery.GetAllShortestPaths(ctx, startNode, endNode, kindFilter); err != nil {
		return nil, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request)
	} else {
		return paths, nil
	}
}

func (s Resources) getAllShortestPathsWithOpenGraph(ctx context.Context, relationshipKindsParam, startNode, endNode string, onlyIncludeTraversableKinds bool, validKinds graph.Kinds, request *http.Request) (graph.PathSet, *api.ErrorWrapper) {
	relationshipKindFilters := model.Filters{}
	if onlyIncludeTraversableKinds {
		relationshipKindFilters["is_traversable"] = append(relationshipKindFilters["is_traversable"], model.Filter{Operator: model.Equals, Value: "true"})
	}
	if openGraphRelationships, _, err := s.DB.GetGraphSchemaRelationshipKinds(ctx, relationshipKindFilters, model.Sort{}, 0, 0); err != nil {
		return nil, api.BuildErrorResponse(http.StatusInternalServerError, api.FormatDatabaseError(err).Error(), request)
	} else {
		openGraphRelationshipKinds := make(graph.Kinds, 0, len(openGraphRelationships))
		for _, relationship := range openGraphRelationships {
			openGraphRelationshipKinds = append(openGraphRelationshipKinds, graph.StringKind(relationship.Name))
		}
		validKinds = validKinds.Concatenate(openGraphRelationshipKinds)
		if kindFilter, err := createRelationshipKindFilterCriteria(relationshipKindsParam, onlyIncludeTraversableKinds, validKinds); err != nil {
			return nil, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request)
		} else if paths, err := s.GraphQuery.GetAllShortestPathsWithOpenGraph(ctx, startNode, endNode, kindFilter); err != nil {
			return nil, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request)

		} else {
			return paths, nil
		}
	}

}

const (
	searchParameterQuery = "query"
	searchParameterType  = "type"
)

func (s *Resources) GetSearchResult(response http.ResponseWriter, request *http.Request) {
	var (
		params          = request.URL.Query()
		customNodeKinds []model.CustomNodeKind
		filteredGraph   map[string]any
	)
	user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
		return
	}

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
			if customNodeKinds, err = s.DB.GetCustomNodeKinds(request.Context()); err != nil {
				slog.Error("Unable to fetch custom nodes from database; will fall back to defaults")
			}
			bhGraph := bloodhoundgraph.NodeSetToBloodHoundGraph(nodes, openGraphSearchFeatureFlag.Enabled, createCustomNodeKindMap(customNodeKinds))

			// ETAC DogTags filtering
			if ShouldFilterForETAC(s.DogTags, user) {
				accessList := ExtractEnvironmentIDsFromUser(&user)
				filteredGraph, err = filterSearchResultMap(bhGraph, accessList)
				if err != nil {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "error filtering search results", request), response)
					return
				}
			} else {
				filteredGraph = bhGraph
			}

			api.WriteBasicResponse(request.Context(), filteredGraph, http.StatusOK, response)
		}
	}
}

// filterSearchResultMap applies ETAC(Environment-based Access Control) filtering to pathfinding.
// Nodes that the user doesn't have access to are marked as hidden.
// The function checks each node's environment (domain sid/tenant id) against the user's access list.
func filterSearchResultMap(graphMap map[string]any, accessList []string) (map[string]any, error) {
	environmentKeys := []string{"domainsid", "tenantid"}
	filteredNodes := make(map[string]any, len(graphMap))

	for id, nodeInterface := range graphMap {
		// type assert to BloodHoundGraphNode struct
		node, ok := nodeInterface.(bloodhoundgraph.BloodHoundGraphNode)
		if !ok {
			// if type assertion fails, keep the node as is
			filteredNodes[id] = nodeInterface
			continue
		}

		hasAccess := false

		// check if the user has access to a node's environment
		if node.BloodHoundGraphItem != nil && node.Data != nil {
			for _, key := range environmentKeys {
				if val, ok := node.Data[key].(string); ok && slices.Contains(accessList, val) {
					hasAccess = true
					break
				}
			}
		}

		if hasAccess {
			// user has access, keep node as is
			filteredNodes[id] = nodeInterface
		} else {
			// user does not have access. create hidden placeholder node
			sourceKind := "Unknown"
			if node.BloodHoundGraphItem != nil && node.Data != nil {
				if kinds, ok := node.Data["kinds"].([]string); ok && len(kinds) > 0 {
					sourceKind = kinds[0]
				}
			}
			// extract the node source kind to display in the hidden label
			filteredNodes[id] = bloodhoundgraph.BloodHoundGraphNode{
				BloodHoundGraphItem: &bloodhoundgraph.BloodHoundGraphItem{
					Data: map[string]any{
						"hidden": true,
					},
				},
				Label: &bloodhoundgraph.BloodHoundGraphNodeLabel{
					Text: fmt.Sprintf("** Hidden %s Object **", sourceKind),
				},
				Shape: "ellipse",
				Size:  20,
			}
		}
	}

	return filteredNodes, nil
}

func createCustomNodeKindMap(customNodeKinds []model.CustomNodeKind) model.CustomNodeKindMap {
	customNodeKindMap := make(model.CustomNodeKindMap)
	for _, kind := range customNodeKinds {
		customNodeKindMap[kind.KindName] = kind.Config
	}
	return customNodeKindMap
}
