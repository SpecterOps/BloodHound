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
	"slices"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/src/utils/validation"
)

const (
	ErrInvalidAssetGroupTagId         = "invalid asset group tag id specified in url"
	ErrInvalidAssetGroupTagSelectorId = "invalid asset group tag selector id specified in url"
	ErrInvalidSelectorType            = "invalid selector type"
)

// Checks that the selector seeds are valid.
func validateSelectorSeeds(graph queries.Graph, seeds []model.SelectorSeed) error {
	if len(seeds) > 0 {
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

func (s *Resources) UpdateAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	var (
		selUpdateReq  model.AssetGroupTagSelector
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
		rawSelectorID = mux.Vars(request)[api.URIPathVariableAssetGroupTagSelectorID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Update")()

	if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, ErrInvalidAssetGroupTagId, request), response)
	} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectorId, err := strconv.Atoi(rawSelectorID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, ErrInvalidAssetGroupTagSelectorId, request), response)
	} else if selector, err := s.DB.GetAssetGroupTagSelectorBySelectorId(request.Context(), selectorId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := json.NewDecoder(request.Body).Decode(&selUpdateReq); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if err := validateSelectorSeeds(s.GraphQuery, selUpdateReq.Seeds); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else {
		// we can update DisabledAt on a default selector
		if selUpdateReq.DisabledAt.Valid {
			if selector.AllowDisable {
				selector.DisabledAt = selUpdateReq.DisabledAt
				if selector.DisabledAt.Time.IsZero() {
					// clear DisabledBy if DisabledAt is set to zero
					selector.DisabledBy = null.String{}
				} else {
					selector.DisabledBy = null.StringFrom(actor.ID.String())
				}
			} else {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "this selector cannot be disabled", request), response)
				return
			}
		}

		// we can update AutoCertify on a default selector
		if selUpdateReq.AutoCertify.Valid {
			selector.AutoCertify = selUpdateReq.AutoCertify
		}

		if selUpdateReq.Name != "" {
			if !selector.IsDefault {
				selector.Name = selUpdateReq.Name
			} else {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "cannot update name on a default selector", request), response)
				return
			}
		}

		if selUpdateReq.Description != "" {
			if !selector.IsDefault {
				selector.Description = selUpdateReq.Description
			} else {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "cannot update description on a default selector", request), response)
				return
			}
		}

		// if seeds are not included, call the DB update with them set to nil
		var seedsTemp []model.SelectorSeed
		if len(selUpdateReq.Seeds) > 0 {
			if !selector.IsDefault {
				selector.Seeds = selUpdateReq.Seeds
			} else {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "cannot update seeds on a default selector", request), response)
				return
			}
		} else {
			// the DB update function will skip updating the seeds in this case
			seedsTemp = selector.Seeds
			selector.Seeds = nil
		}

		if selector, err := s.DB.UpdateAssetGroupTagSelector(request.Context(), selector); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			if seedsTemp != nil {
				// seeds were unchanged, set them back to what is stored in the db for the response
				selector.Seeds = seedsTemp
			}
			api.WriteBasicResponse(request.Context(), selector, http.StatusOK, response)
		}
	}
}

func (s *Resources) GetAssetGroupTagSelectors(response http.ResponseWriter, request *http.Request) {
	var (
		assetTagIdStr            = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
		selectorQueryFilter      = make(model.QueryParameterFilterMap)
		selectorSeedsQueryFilter = make(model.QueryParameterFilterMap)
		selectorSeed             = model.SelectorSeed{}
		assetGroupTagSelector    = model.AssetGroupTagSelector{}
	)

	if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		// The below is a workaround to split the query filters by the two tables to be used in the subsequent db calls
		for name, filters := range queryFilters {
			// get valid selector predicates and valid selector seed predicates.
			validSelectorPredicates, selectorFilterErr := api.GetValidFilterPredicatesAsStrings(assetGroupTagSelector, name)
			validSelectorSeedPredicates, seedFilterErr := api.GetValidFilterPredicatesAsStrings(selectorSeed, name)
			// return an error if both attempts fail, as either one could be used to build separate queries.
			if selectorFilterErr != nil && seedFilterErr != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			for _, filter := range filters {
				if !slices.Contains(validSelectorPredicates, string(filter.Operator)) && !slices.Contains(validSelectorSeedPredicates, string(filter.Operator)) {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
					return
				}
				if slices.Contains(validSelectorPredicates, string(filter.Operator)) {
					selectorQueryFilter.AddFilter(filter)
					selectorQueryFilter[name][len(selectorQueryFilter[name])-1].IsStringData = assetGroupTagSelector.IsStringColumn(filter.Name)
				} else if slices.Contains(validSelectorSeedPredicates, string(filter.Operator)) {
					selectorSeedsQueryFilter.AddFilter(filter)
					// There are no string columns on asset group selector seeds table
				}
			}
		}

		defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Label Get Selector")()

		if assetGroupTagID, err := strconv.Atoi(assetTagIdStr); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, ErrInvalidAssetGroupTagId, request), response)
		} else if selectorSqlFilter, err := selectorQueryFilter.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if selectorSeedSqlFilter, err := selectorSeedsQueryFilter.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetGroupTagID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if selectors, err := s.DB.GetAssetGroupTagSelectorsByTagId(request.Context(), assetGroupTagID, selectorSqlFilter, selectorSeedSqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), model.ListSelectorsResponse{Selectors: selectors}, http.StatusOK, response)
		}
	}
}

type getAssetGroupTagResponse struct {
	Tag model.AssetGroupTag `json:"tag"`
}

func (s *Resources) GetAssetGroupTag(response http.ResponseWriter, request *http.Request) {
	if tagId, err := strconv.ParseInt(mux.Vars(request)[api.URIPathVariableAssetGroupTagID], 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), int(tagId)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), getAssetGroupTagResponse{Tag: assetGroupTag}, http.StatusOK, response)
	}
}
