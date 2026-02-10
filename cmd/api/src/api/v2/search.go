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
		// ETAC DogTags
		if ShouldFilterForETAC(s.DogTags, user) {
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

func (s *Resources) ListAvailableEnvironments(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	sortItems, err := api.ParseGraphSortParameters(model.EnvironmentSelectors{}, request.URL.Query())
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
		return
	}

	filterResult, err := BuildEnvironmentFilter(ctx, s.DB, request)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	// Fetch and filter domain nodes
	nodes, err := s.GraphQuery.GetFilteredAndSortedNodes(sortItems, filterResult.FilterCriteria)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	// Build response with domain type display names
	responseData := BuildEnvironmentSelectors(nodes, filterResult.KindToDisplayName)

	api.WriteBasicResponse(ctx, responseData, http.StatusOK, response)
}

func BuildEnvironmentSelectors(nodes []*graph.Node, kindToDisplayName map[string]string) model.EnvironmentSelectors {
	envs := make(model.EnvironmentSelectors, 0, len(nodes))

	for _, node := range nodes {
		name, _ := node.Properties.GetOrDefault(common.Name.String(), graphschema.DefaultMissingName).String()
		objectID, _ := node.Properties.GetOrDefault(common.ObjectID.String(), graphschema.DefaultMissingObjectId).String()

		envType := resolveEnvType(node, kindToDisplayName)
		collected := resolveCollected(node)

		envs = append(envs, model.EnvironmentSelector{
			Type:      envType,
			Name:      name,
			ObjectID:  objectID,
			Collected: collected,
		})
	}

	return envs
}

func resolveCollected(node *graph.Node) bool {
	// If the collected property doesn't exist, default to false
	if !node.Properties.Exists(common.Collected.String()) {
		return false
	}

	collected, _ := node.Properties.Get(common.Collected.String()).Bool()

	// Built-in environments (AD/Azure) respect the collected property
	isBuiltinEnvironment := node.Kinds.ContainsOneOf(azure.Tenant, ad.Domain)
	if isBuiltinEnvironment {
		return collected
	}

	// OpenGraph extensions always default to true (collected)
	return true
}

func resolveEnvType(node *graph.Node, kindToDisplayName map[string]string) string {
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

// EnvironmentFilterResult contains the filter criteria and display name mapping for environments
type EnvironmentFilterResult struct {
	FilterCriteria    graph.Criteria
	KindToDisplayName map[string]string
}

// BuildEnvironmentFilter constructs the graph filter criteria based on environments and feature flags.
func BuildEnvironmentFilter(ctx context.Context, db database.Database, request *http.Request) (EnvironmentFilterResult, error) {
	var result EnvironmentFilterResult

	// Fetch schema environments
	environments, err := db.GetEnvironments(ctx)
	if err != nil {
		return result, err
	}

	// Build environment kind mappings
	environmentKinds := make([]graph.Kind, len(environments))
	kindToDisplayName := make(map[string]string, len(environments))
	for i, env := range environments {
		environmentKinds[i] = graph.StringKind(env.EnvironmentKindName)
		kindToDisplayName[env.EnvironmentKindName] = env.SchemaExtensionDisplayName
	}

	// Check OpenGraph findings feature flag
	openGraphFlag, err := db.GetFlagByKey(ctx, appcfg.FeatureOpenGraphFindings)
	if err != nil {
		return result, err
	}

	builtinEnvironmentKinds := []graph.Kind{ad.Domain, azure.Tenant}

	if openGraphFlag.Enabled {
		builtinEnvironmentKinds = append(builtinEnvironmentKinds, environmentKinds...)
	}

	// Build base filter criteria
	filterCriteria, err := model.EnvironmentSelectors{}.GetFilterCriteria(request, builtinEnvironmentKinds)
	if err != nil {
		return result, err
	}

	result.FilterCriteria = filterCriteria
	result.KindToDisplayName = kindToDisplayName
	return result, nil
}
