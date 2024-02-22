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
	"strconv"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/src/model"

	"github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/api"
)

const (
	edgeParameterEdgeType   = "edge_type"
	edgeParameterSourceNode = "source_node"
	edgeParameterTargetNode = "target_node"
)

func (s *Resources) GetEdgeComposition(response http.ResponseWriter, request *http.Request) {
	var (
		params = request.URL.Query()
	)

	if edgeType, hasParameter := params[edgeParameterEdgeType]; !hasParameter {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected %s parameter to be set.", edgeParameterEdgeType), request), response)
	} else if sourceNode, hasParameter := params[edgeParameterSourceNode]; !hasParameter {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected %s parameter to be set.", edgeParameterSourceNode), request), response)
	} else if targetNode, hasParameter := params[edgeParameterTargetNode]; !hasParameter {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected %s parameter to be set.", edgeParameterTargetNode), request), response)
	} else if len(edgeType) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected only one %s.", edgeParameterEdgeType), request), response)
	} else if len(sourceNode) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected only one %s.", edgeParameterSourceNode), request), response)
	} else if len(targetNode) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected only one %s.", edgeParameterTargetNode), request), response)
	} else if kind, err := analysis.ParseKind(edgeType[0]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid edge requested: %s", edgeType[0]), request), response)
	} else if startID, err := strconv.ParseInt(sourceNode[0], 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid value for startID: %s", sourceNode[0]), request), response)
	} else if endID, err := strconv.ParseInt(targetNode[0], 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid value for endID: %s", targetNode[0]), request), response)
	} else if edge, err := analysis.FetchEdgeByStartAndEnd(request.Context(), s.Graph, graph.ID(startID), graph.ID(endID), kind); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Could not find edge matching criteria: %v", err), request), response)
	} else if pathSet, err := ad.GetEdgeCompositionPath(request.Context(), s.Graph, edge); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting composition for edge: %v", err), request), response)
	} else {
		unifiedGraph := model.NewUnifiedGraph()
		unifiedGraph.AddPathSet(pathSet, true)
		api.WriteBasicResponse(request.Context(), unifiedGraph, http.StatusOK, response)
	}
}
