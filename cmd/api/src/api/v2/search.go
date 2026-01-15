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

// TODO: you can extract the environment fetching out into a utilty and call from both apps
func (s *Resources) GetAvailableDomains(response http.ResponseWriter, request *http.Request) {
	var (
		domainSelectors = model.DomainSelectors{}
		ctx             = request.Context()
	)

	sortItems, err := api.ParseGraphSortParameters(domainSelectors, request.URL.Query())
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
		return
	}

	// Fetch schema environments to get environment kinds and their display names
	environments, err := s.DB.GetSchemaEnvironments(ctx)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		}
		return
	}

	// Build environment kind filter and display name mapping
	environmentKinds := make([]graph.Kind, len(environments))
	kindToDisplayName := make(map[string]string, len(environments))
	for i, env := range environments {
		environmentKinds[i] = graph.StringKind(env.EnvironmentKindName)
		kindToDisplayName[env.EnvironmentKindName] = env.SchemaExtensionDisplayName
	}

	flag, err := s.DB.GetFlagByKey(ctx, appcfg.FeatureOpenGraphFindings)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	builtinEnvironmentKinds := []graph.Kind{ad.Domain, azure.Tenant}

	if flag.Enabled {
		builtinEnvironmentKinds = append(builtinEnvironmentKinds, environmentKinds...)
	}

	// Build base filter criteria
	filterCriteria, err := model.DomainSelectors{}.GetFilterCriteria(request, builtinEnvironmentKinds)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	// Fetch and filter domain nodes
	nodes, err := s.GraphQuery.GetFilteredAndSortedNodes(sortItems, filterCriteria)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	// Build response with domain type display names
	responseData := BuildDomainSelectors(nodes, kindToDisplayName)

	api.WriteBasicResponse(ctx, responseData, http.StatusOK, response)
}

func BuildDomainSelectors(nodes []*graph.Node, kindToDisplayName map[string]string) model.DomainSelectors {
	domains := make(model.DomainSelectors, 0, len(nodes))

	for _, node := range nodes {
		name, _ := node.Properties.GetOrDefault(common.Name.String(), graphschema.DefaultMissingName).String()
		objectID, _ := node.Properties.GetOrDefault(common.ObjectID.String(), graphschema.DefaultMissingObjectId).String()
		collected, _ := node.Properties.GetOrDefault(common.Collected.String(), false).Bool()

		domainType := resolveDomainType(node, kindToDisplayName)
		domains = append(domains, model.DomainSelector{
			Type:      domainType,
			Name:      name,
			ObjectID:  objectID,
			Collected: collected,
		})
	}

	return domains
}

func resolveDomainType(node *graph.Node, kindToDisplayName map[string]string) string {
	// TODO: Remove hardcoded built-in types once they are saved in DB and not CUE
	if node.Kinds.ContainsOneOf(azure.Tenant) {
		return "azure"
	}
	if node.Kinds.ContainsOneOf(ad.Domain) {
		return "active-directory"
	}

	// For custom extensions, use the display name from the schema extension
	// Note: Nodes should only have one environment kind. In the edge case where there are multiple, we take the first.
	for _, kind := range node.Kinds {
		if displayName, ok := kindToDisplayName[kind.String()]; ok {
			return displayName
		}
	}

	return ""
}
