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

	"github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	adSchema "github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/api"
)

const (
	edgeParameterEdgeType   = "edge_type"
	edgeParameterSourceNode = "source_node"
	edgeParameterTargetNode = "target_node"
)

var validEdgeTypes = []string{"adcsesc1"}

func (s *Resources) GetEdgeDetails(response http.ResponseWriter, request *http.Request) {
	var (
		params  = request.URL.Query()
		pathSet graph.PathSet
	)

	if edgeType, hasParameter := params[edgeParameterEdgeType]; !hasParameter {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected %s parameter to be set.", edgeParameterEdgeType), request), response)
	} else if len(edgeType) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Expected only one %s.", edgeParameterEdgeType), request), response)
	} else {
		if startID, err := strconv.ParseInt(params[edgeParameterSourceNode][0], 10, 32); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalided %s", edgeParameterSourceNode), request), response)
		} else if endID, err := strconv.ParseInt(params[edgeParameterTargetNode][0], 10, 32); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalided %s", edgeParameterSourceNode), request), response)
		} else if err := s.Graph.ReadTransaction(request.Context(), func(tx graph.Transaction) error {
			if fetchedPathSet, err := ad.GetEdgeDetailPath(tx,
				graph.Relationship{
					StartID: graph.ID(startID),
					EndID:   graph.ID(endID),
					Kind:    adSchema.ADCSESC1,
				},
			); err != nil {
				return err
			} else {
				pathSet = fetchedPathSet
			}

			return nil
		}); err != nil {
			if graph.IsErrNotFound(err) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "not found", request), response)
			} else {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error: %v", err), request), response)
			}
		} else {
			api.WriteBasicResponse(request.Context(), pathSet, http.StatusOK, response)
		}
	}
}
