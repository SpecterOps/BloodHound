// Copyright 2025 Specter Ops, Inc.
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
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/src/utils/validation"
	"gorm.io/gorm/utils"
)

const (
	ErrInvalidAssetGroupTagId = "invalid asset group tag id specified in url"
	ErrNoAssetGroupTagId      = "no asset group tag id specified in url"
)

// Checks that the selector seeds are valid.
func validateSelectorSeeds(graph queries.Graph, seeds []model.SelectorSeed) error {
	// all seeds must be of the same type
	seedType := seeds[0].Type

	if seedType != model.SelectorTypeObjectId && seedType != model.SelectorTypeCypher {
		return fmt.Errorf("invalid seed type %v", seedType)
	}

	for _, seed := range seeds {
		if seed.Type != seedType {
			return fmt.Errorf("all seeds must be of the same type")
		}
		if seed.Type == model.SelectorTypeCypher {
			if _, err := graph.PrepareCypherQuery(seed.Value, queries.QueryComplexityLimitSelector); err != nil {
				return fmt.Errorf("cypher is invalid: %v", err)
			}
		}
	}
	return nil
}

func (s *Resources) CreateAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	var (
		sel           model.AssetGroupTagSelector
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Create")()

	if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, ErrInvalidAssetGroupTagId, request), response)
	} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := json.NewDecoder(request.Body).Decode(&sel); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if errs := validation.Validate(sel); len(errs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if err := validateSelectorSeeds(s.GraphQuery, sel.Seeds); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if selector, err := s.DB.CreateAssetGroupTagSelector(request.Context(), assetTagId, actor.ID.String(), sel.Name, sel.Description, false, true, sel.AutoCertify, sel.Seeds); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), selector, http.StatusCreated, response)
	}
}

func (s *Resources) GetAssetGroupTagSelectors(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams           = request.URL.Query()
		rawSelectorTypeString = queryParams[api.QueryParameterAssetGroupSelectorType]
		assetGroupSelector    = model.AssetGroupTagSelector{}
	)

	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Label Get Selector")()

	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := assetGroupSelector.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					if !utils.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
					queryFilters[name][i].IsStringData = assetGroupSelector.IsString(filter.Name)
				}
			}
		}

		if rawAssetGroupTagID, hasAssetGroupTagID := mux.Vars(request)[api.URIPathVariableAssetGroupTagID]; !hasAssetGroupTagID {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoAssetGroupTagId, request), response)
		} else if assetGroupTagID, err := strconv.Atoi(rawAssetGroupTagID); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrInvalidAssetGroupTagId, request), response)
		} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetGroupTagID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if selectors, err := s.DB.GetAssetGroupTagSelectorsByTagId(request.Context(), assetGroupTagID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if len(rawSelectorTypeString) != 0 {
			for _, selType := range rawSelectorTypeString {
				if selectorTypeInt, err := strconv.Atoi(selType); err != nil {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrInvalidAssetGroupTagId, request), response)
				} else if selectorTypeInt != 0 {
					var typeSelectors []model.AssetGroupTagSelector
					for _, selector := range selectors {
						for _, seed := range selector.Seeds {
							if seed.Type == model.SelectorTypeObjectId {
								typeSelectors = append(typeSelectors, selector)
							} else {
								typeSelectors = append(typeSelectors, selector)
							}
						}
					}
					api.WriteBasicResponse(request.Context(), model.ListSelectorsResponse{Selectors: typeSelectors}, http.StatusOK, response)
				}
			}

		} else {
			api.WriteBasicResponse(request.Context(), model.ListSelectorsResponse{Selectors: selectors}, http.StatusOK, response)
		}
	}
}
