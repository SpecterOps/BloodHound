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
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	bhCtx "github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

func (s Resources) SearchHandler(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams     = request.URL.Query()
		searchQuery     = queryParams.Get("q")
		nodeTypes       = queryParams["type"]
		ctx             = request.Context()
		etacAllowedList []string
	)

	if user, isUser := auth.GetUserFromAuthCtx(bhCtx.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "no associated user found with request", request), response)
		return
	} else {
		// ETAC feature flag
		if etacFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureETAC); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		} else if etacFlag.Enabled && !user.AllEnvironments {
			etacAllowedList = ExtractEnvironmentIDsFromUser(&user)
		}
	}

	if searchQuery == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid search parameter", request), response)
	} else if skip, limit, _, err := utils.GetPageParamsForGraphQuery(context.Background(), queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid query parameter: %v", err), request), response)
	} else if openGraphSearchFeatureFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureOpenGraphSearch); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if nodeKinds, err := getNodeKinds(openGraphSearchFeatureFlag.Enabled, nodeTypes...); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid type parameter", request), response)
	} else if result, err := s.GraphQuery.SearchNodesByNameOrObjectId(ctx, nodeKinds, searchQuery, openGraphSearchFeatureFlag.Enabled, skip, limit, etacAllowedList); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Graph error: %v", err), request), response)
	} else {
		api.WriteBasicResponse(request.Context(), result, http.StatusOK, response)
	}
}

// getNodeKinds preserves legacy parseKinds behavior when the OpenGraphSearch feature flag is disabled.
func getNodeKinds(openGraphSearchEnabled bool, nodeTypes ...string) (graph.Kinds, error) {
	if !openGraphSearchEnabled && len(nodeTypes) == 0 {
		return analysis.ParseKinds(ad.Entity.String(), azure.Entity.String())
	} else {
		return analysis.ParseKinds(nodeTypes...)
	}
}

func (s *Resources) GetAvailableDomains(response http.ResponseWriter, request *http.Request) {
	var (
		domains = model.DomainSelectors{}
		ctx     = request.Context()
	)

	sortItems, err := api.ParseGraphSortParameters(domains, request.URL.Query())
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
		return
	}

	filterCriteria, err := domains.GetFilterCriteriaWithEnvironments(ctx, request, s.DB)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		}
		return
	}

	nodes, err := s.GraphQuery.GetFilteredAndSortedNodes(sortItems, filterCriteria)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	api.WriteBasicResponse(ctx, setNodeProperties(nodes), http.StatusOK, response)
}

func setNodeProperties(nodes []*graph.Node) model.DomainSelectors {
	domains := model.DomainSelectors{}
	for _, node := range nodes {
		var (
			name, _      = node.Properties.GetOrDefault(common.Name.String(), graphschema.DefaultMissingName).String()
			objectID, _  = node.Properties.GetOrDefault(common.ObjectID.String(), graphschema.DefaultMissingObjectId).String()
			collected, _ = node.Properties.GetOrDefault(common.Collected.String(), false).Bool()
			domainType   = ""
		)

		if node.Kinds.ContainsOneOf(azure.Tenant) {
			domainType = "azure"
		} else if node.Kinds.ContainsOneOf(ad.Domain) {
			domainType = "active-directory"
		} else {
			domainType = "opengraph"
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
