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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/src/utils/validation"
)

type AssetGroupTagCounts struct {
	Selectors int   `json:"selectors"`
	Members   int64 `json:"members"`
}

type AssetGroupTagView struct {
	model.AssetGroupTag
	Counts *AssetGroupTagCounts `json:"counts,omitempty"`
}

type GetAssetGroupTagsResponse struct {
	Tags []AssetGroupTagView `json:"tags"`
}

func (s Resources) GetAssetGroupTags(response http.ResponseWriter, request *http.Request) {
	var rCtx = request.Context()

	if paramIncludeCounts, err := api.ParseOptionalBool(request.URL.Query().Get(api.QueryParameterIncludeCounts), false); err != nil {
		api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specifed for include counts", request), response)
	} else if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(model.AssetGroupTag{}, name); err != nil {
				api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
					queryFilters[name][i].IsStringData = model.AssetGroupTag{}.IsStringColumn(filter.Name)
				}
			}
		}

		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if tags, err := s.DB.GetAssetGroupTags(rCtx, sqlFilter); err != nil && !errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
		} else {
			var (
				resp = GetAssetGroupTagsResponse{
					Tags: make([]AssetGroupTagView, 0, len(tags)),
				}
				selectorCounts map[int]int
			)

			if paramIncludeCounts {
				ids := make([]int, 0, len(tags))
				for i := range tags {
					ids = append(ids, tags[i].ID)
				}
				if selectorCounts, err = s.DB.GetAssetGroupTagSelectorCounts(rCtx, ids); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				}
			}

			for _, tag := range tags {
				tview := AssetGroupTagView{AssetGroupTag: tag}
				if paramIncludeCounts {
					if n, err := s.GraphQuery.CountNodesByKind(rCtx, tag.ToKind()); err != nil {
						api.HandleDatabaseError(request, response, err)
						return
					} else {
						tview.Counts = &AssetGroupTagCounts{
							Selectors: selectorCounts[tag.ID],
							Members:   n,
						}
					}
				}
				resp.Tags = append(resp.Tags, tview)
			}
			api.WriteBasicResponse(rCtx, resp, http.StatusOK, response)
		}
	}
}

// Checks that the selector seeds are valid.
func validateSelectorSeeds(graph queries.Graph, seeds []model.SelectorSeed) error {
	if len(seeds) <= 0 {
		return fmt.Errorf("seeds are required")
	}
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
			if _, err := graph.PrepareCypherQuery(seed.Value, queries.DefaultQueryFitnessLowerBoundSelector); err != nil {
				return fmt.Errorf("cypher is invalid: %v", err)
			}
		}
	}
	return nil
}

func (s *Resources) CreateAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	var (
		sel = model.AssetGroupTagSelector{
			AutoCertify: null.BoolFrom(false), // default if unset
		}
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Create")()

	if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
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
	} else if selector, err := s.DB.CreateAssetGroupTagSelector(request.Context(), assetTagId, actor, sel.Name, sel.Description, false, true, sel.AutoCertify, sel.Seeds); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// Request analysis if scheduled analysis isn't enabled
		if config, err := appcfg.GetScheduledAnalysisParameter(request.Context(), s.DB); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		} else if !config.Enabled {
			if err := s.DB.RequestAnalysis(request.Context(), actor.ID.String()); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		}
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectorId, err := strconv.Atoi(rawSelectorID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if selector, err := s.DB.GetAssetGroupTagSelectorBySelectorId(request.Context(), selectorId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selector.AssetGroupTagId != assetTagId {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "selector is not part of asset group tag", request), response)
	} else if err := json.NewDecoder(request.Body).Decode(&selUpdateReq); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
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
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "this selector cannot be disabled", request), response)
				return
			}
		}

		// we can update AutoCertify on a default selector
		if selUpdateReq.AutoCertify.Valid {
			selector.AutoCertify = selUpdateReq.AutoCertify
		}

		if selector.IsDefault && (selUpdateReq.Name != "" || selUpdateReq.Description != "" || len(selUpdateReq.Seeds) > 0) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "default selectors only support modifying auto_certify and disabled_at", request), response)
			return
		}

		if selUpdateReq.Name != "" {
			selector.Name = selUpdateReq.Name
		}

		if selUpdateReq.Description != "" {
			selector.Description = selUpdateReq.Description
		}

		// if seeds are not included, call the DB update with them set to nil
		var seedsTemp []model.SelectorSeed
		if len(selUpdateReq.Seeds) > 0 {
			if err := validateSelectorSeeds(s.GraphQuery, selUpdateReq.Seeds); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			}
			selector.Seeds = selUpdateReq.Seeds
		} else {
			// the DB update function will skip updating the seeds in this case
			seedsTemp = selector.Seeds
			selector.Seeds = nil
		}

		if selector, err := s.DB.UpdateAssetGroupTagSelector(request.Context(), actor, selector); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			if seedsTemp != nil {
				// seeds were unchanged, set them back to what is stored in the db for the response
				selector.Seeds = seedsTemp
			}
			// Request analysis if scheduled analysis isn't enabled
			if config, err := appcfg.GetScheduledAnalysisParameter(request.Context(), s.DB); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else if !config.Enabled {
				if err := s.DB.RequestAnalysis(request.Context(), actor.ID.String()); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				}
			}
			api.WriteBasicResponse(request.Context(), selector, http.StatusOK, response)
		}
	}
}

type listSelectorsResponse struct {
	Selectors model.AssetGroupTagSelectors `json:"selectors"`
}

func (s *Resources) DeleteAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	var (
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
		rawSelectorID = mux.Vars(request)[api.URIPathVariableAssetGroupTagSelectorID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Delete")()

	if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectorId, err := strconv.Atoi(rawSelectorID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if selector, err := s.DB.GetAssetGroupTagSelectorBySelectorId(request.Context(), selectorId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selector.AssetGroupTagId != assetTagId {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "selector is not part of asset group tag", request), response)
	} else if selector.IsDefault {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "cannot delete a default selector", request), response)
	} else if err := s.DB.DeleteAssetGroupTagSelector(request.Context(), actor, selector); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// Request analysis if scheduled analysis isn't enabled
		if config, err := appcfg.GetScheduledAnalysisParameter(request.Context(), s.DB); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		} else if !config.Enabled {
			if err := s.DB.RequestAnalysis(request.Context(), actor.ID.String()); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		}
		api.WriteBasicResponse(request.Context(), selector, http.StatusNoContent, response)
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
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
		} else if selectorSqlFilter, err := selectorQueryFilter.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if selectorSeedSqlFilter, err := selectorSeedsQueryFilter.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetGroupTagID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if selectors, err := s.DB.GetAssetGroupTagSelectorsByTagId(request.Context(), assetGroupTagID, selectorSqlFilter, selectorSeedSqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), listSelectorsResponse{Selectors: selectors}, http.StatusOK, response)
		}
	}
}

type getAssetGroupTagResponse struct {
	Tag model.AssetGroupTag `json:"tag"`
}

func (s *Resources) GetAssetGroupTag(response http.ResponseWriter, request *http.Request) {
	if tagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), tagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), getAssetGroupTagResponse{Tag: assetGroupTag}, http.StatusOK, response)
	}
}

type ListNodeSelectorsResponse struct {
	Member member `json:"member"`
}

type member struct {
	assetGroupMemberResponse
	Selectors model.AssetGroupTagSelectors `json:"selectors"`
}

type assetGroupMemberResponse struct {
	NodeId      graph.ID       `json:"id"`
	ObjectID    string         `json:"object_id"`
	PrimaryKind string         `json:"primary_kind"`
	Name        string         `json:"name"`
	Properties  map[string]any `json:"properties,omitempty"`

	Source model.AssetGroupSelectorNodeSource `json:"source,omitempty"`
}

func (s *Resources) GetAssetGroupTagMemberInfo(response http.ResponseWriter, request *http.Request) {
	var (
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
		memberStr     = mux.Vars(request)[api.URIPathVariableAssetGroupTagMemberID]
	)

	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selectors By Member Id")()

	if assetGroupTagID, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if memberID, err := strconv.Atoi(memberStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetGroupTagID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectors, err := s.DB.GetSelectorsByMemberId(request.Context(), memberID, assetGroupTagID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if len(selectors) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if entity, err := queries.Graph.FetchNodeByGraphId(s.GraphQuery, request.Context(), graph.ID(memberID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), ListNodeSelectorsResponse{Member: member{nodeToAssetGroupMember(entity, true), selectors}}, http.StatusOK, response)
	}
}

type GetAssetGroupTagMemberCountsResponse struct {
	TotalCount int            `json:"total_count"`
	Counts     map[string]int `json:"counts"`
}

func (s *Resources) GetAssetGroupTagMemberCountsByKind(response http.ResponseWriter, request *http.Request) {
	if tagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if tag, err := s.DB.GetAssetGroupTag(request.Context(), tagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if primaryNodeKindsCounts, err := s.GraphQuery.GetPrimaryNodeKindCounts(request.Context(), tag.ToKind()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		data := GetAssetGroupTagMemberCountsResponse{
			Counts: primaryNodeKindsCounts,
		}

		for _, count := range primaryNodeKindsCounts {
			data.TotalCount += count
		}

		api.WriteBasicResponse(request.Context(), data, http.StatusOK, response)
	}
}

// Used to minimize the response shape to just the necessary member display fields
func nodeToAssetGroupMember(node *graph.Node, includeProperties bool) assetGroupMemberResponse {
	var (
		objectID, _ = node.Properties.GetOrDefault(common.ObjectID.String(), "NO OBJECT ID").String()
		name, _     = node.Properties.GetWithFallback(common.Name.String(), "NO NAME", common.DisplayName.String(), common.ObjectID.String()).String()
	)

	member := assetGroupMemberResponse{
		NodeId:      node.ID,
		ObjectID:    objectID,
		PrimaryKind: analysis.GetNodeKindDisplayLabel(node),
		Name:        name,
	}

	if includeProperties {
		member.Properties = node.Properties.Map
	}

	return member
}
