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
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
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

	if user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "no associated user found with request", request), response)
		return
	} else if ShouldFilterForETAC(s.DogTags, user) {
		etacAllowedList = ExtractEnvironmentIDsFromUser(&user)
	}

	if searchQuery == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid search parameter", request), response)
	} else if skip, limit, _, err := utils.GetPageParamsForGraphQuery(context.Background(), queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid query parameter: %v", err), request), response)
	} else if openGraphSearchFeatureFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureOpenGraphSearch); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if primaryDisplayKinds, err := s.DB.GetPrimaryDisplayKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if searchableNodeKinds, err := getSearchableNodeKinds(openGraphSearchFeatureFlag.Enabled, primaryDisplayKinds, graph.StringsToKinds(nodeTypes)); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid type parameter", request), response)
	} else if nodes, err := s.GraphQuery.SearchNodesByNameOrObjectId(ctx, searchableNodeKinds, searchQuery, skip, limit); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Graph error: %v", err), request), response)
	} else {
		result := filterAndFormatSearchResults(nodes, etacAllowedList, primaryDisplayKinds)

		api.WriteBasicResponse(request.Context(), result, http.StatusOK, response)
	}
}

func filterAndFormatSearchResults(nodes []*graph.Node, etacAllowedList []string, primaryDisplayKinds graphschema.PrimaryDisplayKinds) []model.SearchResult {
	var results []model.SearchResult

	for _, node := range nodes {
		if !nodeGatedByETAC(etacAllowedList, node) {
			results = append(results, graphNodeToSearchResult(node, primaryDisplayKinds))
		}
	}

	return results
}

func graphNodeToSearchResult(node *graph.Node, primaryDisplayKinds graphschema.PrimaryDisplayKinds) model.SearchResult {
	var (
		name, _              = node.Properties.GetWithFallback(common.Name.String(), graphschema.DefaultMissingName, common.DisplayName.String(), common.ObjectID.String()).String()
		objectID, _          = node.Properties.GetOrDefault(common.ObjectID.String(), graphschema.DefaultMissingObjectId).String()
		distinguishedName, _ = node.Properties.GetOrDefault(ad.DistinguishedName.String(), "").String()
		systemTags, _        = node.Properties.GetOrDefault(common.SystemTags.String(), "").String()
		kindLabel            = graphschema.GetNodeKindDisplayLabel(primaryDisplayKinds, node)
	)

	return model.SearchResult{
		ObjectID:          objectID,
		Type:              kindLabel,
		Name:              name,
		DistinguishedName: distinguishedName,
		SystemTags:        systemTags,
	}
}

// getSearchableNodeKinds returns the kinds that should be searched based on the OpenGraphSearch feature flag and the primary display kinds.
func getSearchableNodeKinds(openGraphSearchEnabled bool, primaryDisplayKinds graphschema.PrimaryDisplayKinds, typeParams graph.Kinds) (graph.Kinds, error) {
	var (
		searchableKinds                graph.Kinds
		validKinds                     graphschema.PrimaryDisplayKinds
		emptyParams                    = len(typeParams) == 0
		invalidParamError              = fmt.Errorf("no primary display kinds found for search types: %v", typeParams)
		kindsShouldNotBeConstrained    = emptyParams && openGraphSearchEnabled
		kindsConstrainedToDefaultKinds = emptyParams && !openGraphSearchEnabled
	)

	if kindsShouldNotBeConstrained {
		return nil, nil
	} else if kindsConstrainedToDefaultKinds {
		return graph.Kinds{ad.Entity, azure.Entity}, nil
	} else {

		if openGraphSearchEnabled {
			// only assign validKinds if OpenGraphSearch is enabled
			// otherwise we pass nil to PrimaryDisplayKind to emulate the old behavior
			validKinds = primaryDisplayKinds
		}

		for _, kind := range typeParams {
			kind := graphschema.PrimaryDisplayKind(validKinds, graph.Kinds{kind})

			if !kind.Is(graphschema.UnknownKind) {
				searchableKinds = searchableKinds.Add(kind)
			}
		}

		if len(searchableKinds) == 0 {
			return nil, invalidParamError
		}

		return searchableKinds, nil
	}
}

func (s *Resources) ListAvailableEnvironments(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	sortItems, err := api.ParseGraphSortParameters(model.EnvironmentSelectors{}, request.URL.Query())
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
		return
	}

	filterResult, err := BuildEnvironmentFilter(ctx, s.DB, s.OpenGraphSchemaService, request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidQueryParameters):
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			return
		default:
			api.HandleDatabaseError(request, response, err)
			return
		}
	}

	// Fetch and filter domain nodes
	nodes, err := s.GraphQuery.GetFilteredAndSortedNodes(sortItems, filterResult.FilterCriteria)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	// Build response with domain type display names
	responseData := BuildEnvironmentSelectors(nodes, filterResult.KindToSchemaEnvironment)

	api.WriteBasicResponse(ctx, responseData, http.StatusOK, response)
}

func BuildEnvironmentSelectors(nodes []*graph.Node, kindToSchemaEnvironment model.EnvironmentKindsToEnvironment) model.EnvironmentSelectors {
	envs := make(model.EnvironmentSelectors, 0, len(nodes))

	for _, node := range nodes {
		name, _ := node.Properties.GetOrDefault(common.Name.String(), graphschema.DefaultMissingName).String()
		objectID, _ := node.Properties.GetOrDefault(common.ObjectID.String(), graphschema.DefaultMissingObjectId).String()

		collected := resolveCollected(node)
		envProperties := resolveEnvProperties(node, kindToSchemaEnvironment)

		envs = append(envs, model.EnvironmentSelector{
			Name:      name,
			ObjectID:  objectID,
			Collected: collected,
			EnvironmentProperties: model.EnvironmentProperties{
				Type:            envProperties.Type,
				KindId:          envProperties.KindId,
				KindDisplayName: envProperties.KindDisplayName,
			},
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

func resolveEnvProperties(node *graph.Node, kindToSchemaEnvironment model.EnvironmentKindsToEnvironment) model.EnvironmentProperties {
	envProperties := model.EnvironmentProperties{}
	// TODO: Remove hardcoded built-in types once they are saved in DB and not CUE
	if node.Kinds.ContainsOneOf(azure.Tenant) {
		envProperties.Type = "azure"
	} else if node.Kinds.ContainsOneOf(ad.Domain) {
		envProperties.Type = "active-directory"
	} else {
		// For custom extensions, use the display name from the schema extension
		// Note: Nodes should only have one environment kind. In the edge case where there are multiple, we take the first.
		for _, kind := range node.Kinds {
			if schemaEnvironment, ok := kindToSchemaEnvironment[kind.String()]; ok {
				envProperties.Type = schemaEnvironment.SchemaExtensionDisplayName
				envProperties.KindDisplayName = &schemaEnvironment.EnvironmentKindName
				envProperties.KindId = &schemaEnvironment.EnvironmentKindId
			}
		}
	}
	return envProperties
}

func resolveExtensionID(node *graph.Node, kindToSchemaEnvironment model.EnvironmentKindsToEnvironment) *int32 {
	// TODO: Remove hardcoded built-in types once they are saved in DB and not CUE
	isBuiltinEnvironment := node.Kinds.ContainsOneOf(azure.Tenant, ad.Domain)
	if isBuiltinEnvironment {
		return nil
	}

	// Note: Nodes should only have one environment kind. In the edge case where there are multiple, we take the first.
	for _, kind := range node.Kinds {
		if schemaEnvironment, ok := kindToSchemaEnvironment[kind.String()]; ok {
			return &schemaEnvironment.SchemaExtensionId
		}
	}

	return nil
}

// EnvironmentFilterResult contains the filter criteria environment data for mapping environments
type EnvironmentFilterResult struct {
	FilterCriteria          graph.Criteria
	KindToSchemaEnvironment model.EnvironmentKindsToEnvironment
}

// ErrInvalidQueryParameters is an error that is used to wrap other errors when the query parameters are invalid
var ErrInvalidQueryParameters = fmt.Errorf("invalid query parameters")

// BuildEnvironmentFilter constructs the graph filter criteria based on environments and feature flags.
func BuildEnvironmentFilter(ctx context.Context, db database.Database, openGraphSchemaService OpenGraphSchemaService, request *http.Request) (EnvironmentFilterResult, error) {
	var result EnvironmentFilterResult

	// Check OpenGraph findings feature flag
	openGraphFlag, err := db.GetFlagByKey(ctx, appcfg.FeatureOpenGraphFindings)
	if err != nil {
		return result, err
		// Fetch schema environments and extension display names
	} else if environmentKinds, envKindToExtension, err := openGraphSchemaService.GetEnvironmentKindsAndSchemaEnvironmentData(ctx, !openGraphFlag.Enabled); err != nil {
		return result, err
	} else {
		// Build base filter criteria
		filterCriteria, err := model.EnvironmentSelectors{}.GetFilterCriteria(request, environmentKinds)
		if err != nil {
			return result, fmt.Errorf("%w: %w", ErrInvalidQueryParameters, err)
		}

		result.FilterCriteria = filterCriteria
		result.KindToSchemaEnvironment = envKindToExtension
		return result, nil
	}
}
