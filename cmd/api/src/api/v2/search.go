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
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
)

func (s Resources) SearchHandler(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams = request.URL.Query()
		searchQuery = queryParams.Get("q")
		nodeTypes   = queryParams["type"]
		ctx         = request.Context()
	)

	if searchQuery == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid search parameter", request), response)
	} else if skip, limit, _, err := utils.GetPageParamsForGraphQuery(context.Background(), queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid query parameter: %v", err), request), response)
	} else if nodeKinds, err := analysis.ParseKinds(nodeTypes...); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid type parameter", request), response)
	} else if result, err := s.GraphQuery.SearchNodesByName(ctx, nodeKinds, searchQuery, skip, limit); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Graph error: %v", err), request), response)
	} else {
		api.WriteBasicResponse(request.Context(), result, http.StatusOK, response)
	}
}

func (s *Resources) GetAvailableDomains(response http.ResponseWriter, request *http.Request) {
	var domains model.DomainSelectors

	if orderCriteria, err := domains.GetOrderCriteria(request.URL.Query()); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else if filterCriteria, err := domains.GetFilterCriteria(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if nodes, err := s.GraphQuery.GetFilteredAndSortedNodes(orderCriteria, filterCriteria); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsInternalServerError, err), request), response)
	} else {
		api.WriteBasicResponse(request.Context(), setNodeProperties(nodes), http.StatusOK, response)
	}
}

func setNodeProperties(nodes graph.NodeSet) model.DomainSelectors {
	domains := model.DomainSelectors{}
	for _, node := range nodes {
		var (
			name, _      = node.Properties.GetOrDefault(common.Name.String(), "NO NAME").String()
			objectID, _  = node.Properties.GetOrDefault(common.ObjectID.String(), "NO OBJECT ID").String()
			collected, _ = node.Properties.GetOrDefault(common.Collected.String(), false).Bool()
			domainType   = "active-directory"
		)

		if node.Kinds.ContainsOneOf(azure.Tenant) {
			domainType = "azure"
		}

		domains = append(domains, model.DomainSelector{
			Type:      domainType,
			Name:      name,
			ObjectID:  objectID,
			Collected: collected,
		})
	}

	return domains
}
