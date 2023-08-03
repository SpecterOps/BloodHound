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

	"github.com/specterops/bloodhound/src/api"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
)

var gpoQueries = map[string]any{
	"ous":         adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOContainerCandidateFilter),
	"computers":   adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectComputersCandidateFilter),
	"users":       adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectUsersCandidateFilter),
	"controllers": adAnalysis.FetchInboundADEntityControllers,
	"tierzero":    adAnalysis.CreateGPOAffectedIntermediariesListDelegate(adAnalysis.SelectGPOTierZeroCandidateFilter),
}

func (s *Resources) GetGPOEntityInfo(response http.ResponseWriter, request *http.Request) {
	if hydrateCounts, err := api.ParseOptionalBool(request.URL.Query().Get(api.QueryParameterHydrateCounts), true); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if objectId, err := GetEntityObjectIDFromRequestPath(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("error reading objectid: %v", err), request), response)
	} else if node, err := s.GraphQuery.GetEntityByObjectId(request.Context(), objectId, ad.GPO); err != nil {
		if graph.IsErrNotFound(err) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "node not found", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error getting node: %v", err), request), response)
		}
	} else if hydrateCounts {
		results := s.GraphQuery.GetEntityCountResults(request.Context(), node, gpoQueries)
		api.WriteBasicResponse(request.Context(), results, http.StatusOK, response)
	} else {
		results := map[string]any{"props": node.Properties.Map}
		api.WriteBasicResponse(request.Context(), results, http.StatusOK, response)
	}
}
