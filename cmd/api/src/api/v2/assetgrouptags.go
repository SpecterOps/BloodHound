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
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

const (
	assetGroupPreviewSelectorDefaultLimit = 200
	assetGroupTagsSearchLimit             = 20

	includeProperties = true
	excludeProperties = false
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

type patchAssetGroupTagSelectorRequest struct {
	model.AssetGroupTagSelector
	Description *string `json:"description"`
	Disabled    *bool   `json:"disabled"`
}

func (s Resources) GetAssetGroupTags(response http.ResponseWriter, request *http.Request) {
	var rCtx = request.Context()

	if paramIncludeCounts, err := api.ParseOptionalBool(request.URL.Query().Get(api.QueryParameterIncludeCounts), false); err != nil {
		api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specified for include counts", request), response)
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
		selUpdateReq  patchAssetGroupTagSelectorRequest
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
		if selUpdateReq.Disabled != nil {
			if *selUpdateReq.Disabled {
				if selector.AllowDisable {
					selector.DisabledAt = null.TimeFrom(time.Now())
					selector.DisabledBy = null.StringFrom(actor.ID.String())
				} else {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "this selector cannot be disabled", request), response)
					return
				}
			} else {
				selector.DisabledAt = null.Time{}
				selector.DisabledBy = null.String{}
			}
		}

		// we can update AutoCertify on a default selector
		if selUpdateReq.AutoCertify.Valid {
			selector.AutoCertify = selUpdateReq.AutoCertify
		}

		if selector.IsDefault && (selUpdateReq.Name != "" || selUpdateReq.Description != nil || len(selUpdateReq.Seeds) > 0) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "default selectors only support modifying auto_certify and disabled_at", request), response)
			return
		}

		if selUpdateReq.Name != "" {
			selector.Name = selUpdateReq.Name
		}

		if selUpdateReq.Description != nil {
			selector.Description = *selUpdateReq.Description
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

		if selector, err := s.DB.UpdateAssetGroupTagSelector(request.Context(), actor.ID.String(), actor.EmailAddress.ValueOrZero(), selector); err != nil {
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
		response.WriteHeader(http.StatusNoContent)
	}
}

type AssetGroupTagSelectorCounts struct {
	Members int64 `json:"members"`
}

type AssetGroupTagSelectorView struct {
	model.AssetGroupTagSelector
	Counts *AssetGroupTagSelectorCounts `json:"counts,omitempty"`
}

type GetAssetGroupTagSelectorResponse struct {
	Selectors []AssetGroupTagSelectorView `json:"selectors"`
}

type GetSelectorResponse struct {
	Selector model.AssetGroupTagSelector `json:"selector"`
}

func (s *Resources) GetAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Get Selector")()

	if assetTagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectorId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagSelectorID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if selector, err := s.DB.GetAssetGroupTagSelectorBySelectorId(request.Context(), selectorId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selector.AssetGroupTagId != assetTagId {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "selector is not part of asset group tag", request), response)
	} else {
		if createdByUser, err := s.DB.GetUser(request.Context(), uuid.FromStringOrNil(selector.CreatedBy)); err == nil {
			selector.CreatedBy = createdByUser.EmailAddress.ValueOrZero()
		}

		if updatedByUser, err := s.DB.GetUser(request.Context(), uuid.FromStringOrNil(selector.UpdatedBy)); err == nil {
			selector.UpdatedBy = updatedByUser.EmailAddress.ValueOrZero()
		}

		api.WriteBasicResponse(request.Context(), GetSelectorResponse{Selector: selector}, http.StatusOK, response)
	}
}

func (s *Resources) GetAssetGroupTagSelectors(response http.ResponseWriter, request *http.Request) {
	var (
		assetTagIdStr            = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
		selectorQueryFilter      = make(model.QueryParameterFilterMap)
		selectorSeedsQueryFilter = make(model.QueryParameterFilterMap)
		selectorSeed             = model.SelectorSeed{}
		assetGroupTagSelector    = model.AssetGroupTagSelector{}
		queryParams              = request.URL.Query()
	)

	if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, 99999); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if paramIncludeCounts, err := api.ParseOptionalBool(queryParams.Get(api.QueryParameterIncludeCounts), false); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specifed for include counts", request), response)
	} else if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
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

		defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Get Selectors")()

		if assetGroupTagID, err := strconv.Atoi(assetTagIdStr); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
		} else if selectorSqlFilter, err := selectorQueryFilter.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if selectorSeedSqlFilter, err := selectorSeedsQueryFilter.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), assetGroupTagID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if selectors, count, err := s.DB.GetAssetGroupTagSelectorsByTagId(request.Context(), assetGroupTagID, selectorSqlFilter, selectorSeedSqlFilter, skip, limit); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			var (
				resp = GetAssetGroupTagSelectorResponse{
					Selectors: make([]AssetGroupTagSelectorView, 0, len(selectors)),
				}
			)

			for _, selector := range selectors {
				selectorView := AssetGroupTagSelectorView{AssetGroupTagSelector: selector}
				if paramIncludeCounts {
					memberCount := int64(0)
					// if the selector is not disabled
					if selector.DisabledAt.Time.IsZero() {
						// get all the nodes which are selected
						if selectorNodes, err := s.DB.GetSelectorNodesBySelectorIds(request.Context(), selector.ID); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							nodeIds := make([]graph.ID, 0, len(selectorNodes))
							for _, node := range selectorNodes {
								nodeIds = append(nodeIds, node.NodeId)
							}

							// only count nodes that are actually tagged
							if count, err := s.GraphQuery.CountFilteredNodes(request.Context(), query.And(
								query.KindIn(query.Node(), assetGroupTag.ToKind()),
								query.InIDs(query.NodeID(), nodeIds...),
							)); err != nil {
								api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting member count: %v", err), request), response)
							} else {
								memberCount = count
							}
						}
					}
					selectorView.Counts = &AssetGroupTagSelectorCounts{
						Members: memberCount,
					}
				}
				resp.Selectors = append(resp.Selectors, selectorView)
			}

			api.WriteResponseWrapperWithPagination(request.Context(), resp, limit, skip, count, http.StatusOK, response)
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
		if createdByUser, err := s.DB.GetUser(request.Context(), uuid.FromStringOrNil(assetGroupTag.CreatedBy)); err == nil {
			assetGroupTag.CreatedBy = createdByUser.EmailAddress.ValueOrZero()
		}

		if updatedByUser, err := s.DB.GetUser(request.Context(), uuid.FromStringOrNil(assetGroupTag.UpdatedBy)); err == nil {
			assetGroupTag.UpdatedBy = updatedByUser.EmailAddress.ValueOrZero()
		}

		api.WriteBasicResponse(request.Context(), getAssetGroupTagResponse{Tag: assetGroupTag}, http.StatusOK, response)
	}
}

type GetAssetGroupTagMemberCountsResponse struct {
	TotalCount int            `json:"total_count"`
	Counts     map[string]int `json:"counts"`
}

func (s *Resources) GetAssetGroupTagMemberCountsByKind(response http.ResponseWriter, request *http.Request) {
	environmentIds := request.URL.Query()[api.QueryParameterEnvironments]

	if tagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if tag, err := s.DB.GetAssetGroupTag(request.Context(), tagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		filters := []graph.Criteria{}
		if len(environmentIds) > 0 {
			filters = append(filters, query.Or(
				query.In(query.NodeProperty(ad.DomainSID.String()), environmentIds),
				query.In(query.NodeProperty(azure.TenantID.String()), environmentIds),
			))
		}

		primaryNodeKindsCounts, err := s.GraphQuery.GetPrimaryNodeKindCounts(request.Context(), tag.ToKind(), filters...)
		if err != nil {
			api.HandleDatabaseError(request, response, err)
		}

		data := GetAssetGroupTagMemberCountsResponse{
			Counts: primaryNodeKindsCounts,
		}
		for _, count := range primaryNodeKindsCounts {
			data.TotalCount += count
		}

		api.WriteBasicResponse(request.Context(), data, http.StatusOK, response)
	}
}

type AssetGroupMember struct {
	NodeId        graph.ID       `json:"id"`
	ObjectID      string         `json:"object_id"`
	EnvironmentID string         `json:"environment_id"`
	PrimaryKind   string         `json:"primary_kind"`
	Name          string         `json:"name"`
	Properties    map[string]any `json:"properties,omitempty"`

	Source model.AssetGroupSelectorNodeSource `json:"source,omitempty"`
}

// Used to minimize the response shape to just the necessary member display fields
func nodeToAssetGroupMember(node *graph.Node, includeProperties bool) AssetGroupMember {
	var (
		objectID, _ = node.Properties.GetOrDefault(common.ObjectID.String(), "NO OBJECT ID").String()
		name, _     = node.Properties.GetWithFallback(common.Name.String(), "NO NAME", common.DisplayName.String(), common.ObjectID.String()).String()
		envID, _    = node.Properties.GetWithFallback(ad.DomainSID.String(), "", azure.TenantID.String()).String()
	)

	member := AssetGroupMember{
		NodeId:        node.ID,
		ObjectID:      objectID,
		EnvironmentID: envID,
		PrimaryKind:   analysis.GetNodeKindDisplayLabel(node),
		Name:          name,
	}

	if includeProperties {
		member.Properties = node.Properties.Map
	}

	return member
}

func (s AssetGroupMember) IsSortable(criteria string) bool {
	switch criteria {
	case "id", "objectid", "name":
		return true
	default:
		return false
	}
}

type MemberInfoResponse struct {
	Member memberInfo `json:"member"`
}

type memberInfo struct {
	AssetGroupMember
	Selectors model.AssetGroupTagSelectors `json:"selectors"`
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
	} else if node, err := queries.Graph.FetchNodeByGraphId(s.GraphQuery, request.Context(), graph.ID(memberID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), MemberInfoResponse{Member: memberInfo{nodeToAssetGroupMember(node, includeProperties), selectors}}, http.StatusOK, response)
	}
}

type GetAssetGroupMembersResponse struct {
	Members []AssetGroupMember `json:"members"`
}

func (s *Resources) GetAssetGroupMembersByTag(response http.ResponseWriter, request *http.Request) {
	var (
		members        = []AssetGroupMember{}
		queryParams    = request.URL.Query()
		environmentIds = queryParams[api.QueryParameterEnvironments]
	)

	if tagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), tagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if sort, err := api.ParseGraphSortParameters(AssetGroupMember{}, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsColumnNotFilterable, request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, 10); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else {
		if len(sort) == 0 {
			sort = query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}
		}
		filters := []graph.Criteria{
			query.KindIn(query.Node(), assetGroupTag.ToKind()),
		}
		if len(environmentIds) > 0 {
			filters = append(filters, query.Or(
				query.In(query.NodeProperty(ad.DomainSID.String()), environmentIds),
				query.In(query.NodeProperty(azure.TenantID.String()), environmentIds),
			))
		}

		if nodes, err := s.GraphQuery.GetFilteredAndSortedNodesPaginated(sort, query.And(filters...), skip, limit); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting members: %v", err), request), response)
		} else if count, err := s.GraphQuery.CountFilteredNodes(request.Context(), query.And(filters...)); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting member count: %v", err), request), response)
		} else {
			for _, node := range nodes {
				members = append(members, nodeToAssetGroupMember(node, excludeProperties))
			}
			api.WriteResponseWrapperWithPagination(request.Context(), GetAssetGroupMembersResponse{Members: members}, limit, skip, int(count), http.StatusOK, response)
		}
	}
}

func (s *Resources) GetAssetGroupMembersBySelector(response http.ResponseWriter, request *http.Request) {
	var (
		members     = []AssetGroupMember{}
		queryParams = request.URL.Query()
	)

	if assetTagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectorId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagSelectorID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if selector, err := s.DB.GetAssetGroupTagSelectorBySelectorId(request.Context(), selectorId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if sort, err := api.ParseGraphSortParameters(AssetGroupMember{}, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsColumnNotFilterable, request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, 10); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if selector.AssetGroupTagId != assetTagId {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "selector is not part of asset group tag", request), response)
	} else if selector.DisabledAt.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, "selector is disabled", request), response)
	} else {
		if len(sort) == 0 {
			sort = query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}
		}

		if selectorNodes, err := s.DB.GetSelectorNodesBySelectorIds(request.Context(), selectorId); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			var (
				nodeIds        = make([]graph.ID, 0, len(selectorNodes))
				sourceByNodeId = make(map[graph.ID]model.AssetGroupSelectorNodeSource, len(selectorNodes))
			)

			for _, node := range selectorNodes {
				nodeIds = append(nodeIds, node.NodeId)
				sourceByNodeId[node.NodeId] = node.Source
			}

			filter := query.And(
				query.KindIn(query.Node(), assetGroupTag.ToKind()),
				query.InIDs(query.NodeID(), nodeIds...),
			)

			if nodes, err := s.GraphQuery.GetFilteredAndSortedNodesPaginated(sort, filter, skip, limit); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting members: %v", err), request), response)
			} else if count, err := s.GraphQuery.CountFilteredNodes(request.Context(), filter); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting member count: %v", err), request), response)
			} else {
				for _, node := range nodes {
					member := nodeToAssetGroupMember(node, false)
					member.Source = sourceByNodeId[node.ID]
					members = append(members, member)
				}

				api.WriteResponseWrapperWithPagination(request.Context(), GetAssetGroupMembersResponse{Members: members}, limit, skip, int(count), http.StatusOK, response)
			}
		}
	}
}

type PreviewSelectorBody struct {
	Seeds model.SelectorSeeds `json:"seeds" validate:"required"`
}

func (s *Resources) PreviewSelectors(response http.ResponseWriter, request *http.Request) {
	var (
		seeds   PreviewSelectorBody
		members = []AssetGroupMember{}
	)

	if limit, err := ParseLimitQueryParameter(request.URL.Query(), assetGroupPreviewSelectorDefaultLimit); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if err := json.NewDecoder(request.Body).Decode(&seeds); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if errs := validation.Validate(seeds); len(errs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if _, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if err := validateSelectorSeeds(s.GraphQuery, seeds.Seeds); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else {
		nodes := datapipe.FetchNodesFromSeeds(request.Context(), s.Graph, seeds.Seeds, model.AssetGroupExpansionMethodAll, limit)
		for _, node := range nodes {
			if node.Node != nil {
				members = append(members, nodeToAssetGroupMember(node.Node, excludeProperties))
			}
		}

		api.WriteBasicResponse(request.Context(), GetAssetGroupMembersResponse{Members: members}, http.StatusOK, response)
	}
}

type SearchAssetGroupTagsResponse struct {
	Tags      model.AssetGroupTags         `json:"tags"`
	Selectors model.AssetGroupTagSelectors `json:"selectors"`
	Members   []AssetGroupMember           `json:"members"`
}

type AssetGroupTagSearchRequest struct {
	Query   string                  `json:"query"`
	TagType model.AssetGroupTagType `json:"tag_type"`
}

func validateAssetGroupTagType(maybeType model.AssetGroupTagType) bool {
	switch maybeType {
	case model.AssetGroupTagTypeTier, model.AssetGroupTagTypeLabel, model.AssetGroupTagTypeOwned:
		return true
	default:
		return false
	}
}

func (s *Resources) SearchAssetGroupTags(response http.ResponseWriter, request *http.Request) {
	var (
		reqBody     = AssetGroupTagSearchRequest{}
		members     = []AssetGroupMember{}
		matchedTags = model.AssetGroupTags{}
		selectors   model.AssetGroupTagSelectors
	)

	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if !validateAssetGroupTagType(reqBody.TagType) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseAssetGroupTagInvalid, request), response)
	} else if len(reqBody.Query) < 3 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsQueryTooShort, request), response)
	} else if tags, err := s.DB.GetAssetGroupTags(request.Context(), model.SQLFilter{}); err != nil && !errors.Is(err, database.ErrNotFound) {
		api.HandleDatabaseError(request, response, err)
	} else {
		var (
			kinds  graph.Kinds
			tagIds []int
		)

		for _, t := range tags {
			// owned tag is a label despite distinct designation
			if reqBody.TagType == t.Type || (reqBody.TagType == model.AssetGroupTagTypeLabel && t.Type == model.AssetGroupTagTypeOwned) {

				// filter the below node and selector query to tag type
				kinds = kinds.Add(t.ToKind())
				tagIds = append(tagIds, t.ID)
				if strings.Contains(strings.ToLower(t.Name), strings.ToLower(reqBody.Query)) && len(matchedTags) < assetGroupTagsSearchLimit {
					matchedTags = append(matchedTags, t)
				}
			}
		}
		var (
			nodeFilter = query.And(
				query.Or(
					query.CaseInsensitiveStringContains(query.NodeProperty(common.Name.String()), reqBody.Query),
					query.CaseInsensitiveStringContains(query.NodeProperty(common.ObjectID.String()), reqBody.Query),
				),
				query.KindIn(query.Node(), kinds...),
			)
			selectorFilter = model.SQLFilter{SQLString: "name ILIKE ? AND asset_group_tag_id IN ?", Params: []any{"%" + reqBody.Query + "%", tagIds}}
		)

		if selectors, err = s.DB.GetAssetGroupTagSelectors(request.Context(), selectorFilter, assetGroupTagsSearchLimit); err != nil && !errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
			return
		} else if nodes, err := s.GraphQuery.GetFilteredAndSortedNodesPaginated(query.SortItems{{SortCriteria: query.NodeProperty(common.Name.String()), Direction: query.SortDirectionAscending}}, nodeFilter, 0, assetGroupTagsSearchLimit); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting members: %v", err), request), response)
			return
		} else {
			for _, node := range nodes {
				members = append(members, nodeToAssetGroupMember(node, excludeProperties))
			}
		}

		api.WriteBasicResponse(request.Context(), SearchAssetGroupTagsResponse{Tags: matchedTags, Selectors: selectors, Members: members}, http.StatusOK, response)
	}
}

type AssetGroupHistoryResp struct {
	Records []model.AssetGroupHistory `json:"records"`
}

func (s *Resources) GetAssetGroupTagHistory(response http.ResponseWriter, request *http.Request) {
	var (
		rCtx        = request.Context()
		queryParams = request.URL.Query()
	)

	defer measure.ContextMeasure(rCtx, slog.LevelDebug, "Asset Group Tag Get History Records")()

	if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(rCtx, ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, 100); err != nil {
		api.WriteErrorResponse(rCtx, ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if sort, err := api.ParseSortParameters(model.AssetGroupHistory{}, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(model.AssetGroupHistory{}, name); err != nil {
				api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
					queryFilters[name][i].IsStringData = model.AssetGroupHistory{}.IsStringColumn(filter.Name)
				}
			}
		}

		if len(sort) == 0 {
			sort = model.Sort{{Column: "created_at", Direction: model.DescendingSortDirection}}
		}

		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if historyRecs, count, err := s.DB.GetAssetGroupHistoryRecords(rCtx, sqlFilter, sort, skip, limit); err != nil && !errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithPagination(rCtx, AssetGroupHistoryResp{Records: historyRecs}, limit, skip, count, http.StatusOK, response)
		}
	}
}

type AssetGroupMemberWithCertification struct {
	AssetGroupMember
	FirstSeen     time.Time                     `json:"first_seen"`
	CertifiedBy   string                        `json:"certified_by"`
	Certification model.AssetGroupCertification `json:"certification"`
}

func (AssetGroupMemberWithCertification) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"environments": {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"asset_group_tag_id":        {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"type":           {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		// TODO -- is query a valid filter? 
		"query":     {},
		"certified": {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"certified_by":     {model.Equals, model.NotEquals, model.ApproximatelyEquals},
	}
}

type FilteringConditional func(t AssetGroupMemberWithCertification) bool

func (s AssetGroupMemberWithCertification) CreateFilteringConditionals(filterMap QueryParameterFilterMap) ([]FilteringConditional, error) {
	conditionals := []FilteringConditional{}
	for column, filters := range filterMap {
		for _, filter := range filters {
			if conditional, err := s.BuildServerSideFilteringConditional(column, filter.Operator, filter.Value); err != nil {
				return conditionals, err
			} else {
				conditionals = append(conditionals, conditional)
			}
		}
	}
	return conditionals, nil
}

func (s AssetGroupMemberWithCertification) MeetsFilterRequirements(conditionals []FilteringConditional) bool {
	for _, conditional := range conditionals {
		if !conditional(s) {
			return false
		}
	}
	return true
}

// TODO do this for server-side processing conditionals AGSNs
// TODO there may be a better way to do this
// Maybe don't do this, might just need to make something bespoke
func (s AssetGroupMemberWithCertification) BuildServerSideFilteringConditional(column string, operator model.FilterOperator, value string) (FilteringConditional, error) {
	switch column {
	case "certified":
		if operator == model.Equals {
			return func(t AssetGroupMemberWithCertification) bool { return t.Certified == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.Certified != value }, nil
		} else {
			return nil, errors.New(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "certified_by":
		if operator == model.Equals {
			return func(t AssetGroupMemberWithCertification) bool { return t.CertifiedBy == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.CertifiedBy != value }, nil
		} else {
			return nil, errors.New(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "start":
		if operator == model.Equals {
			return func(t AssetGroupMemberWithCertification) bool { return t.FirstSeen.String() == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind != value }, nil
		} else if operator == model.GreaterThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind > value }, nil
		} else if operator == model.GreaterThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind >= value }, nil
		} else if operator == model.LessThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind < value }, nil
		} else if operator == model.LessThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind <= value }, nil
		} else {
			return nil, errors.New(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "end":
		if operator == Equals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind == value }, nil
		} else if operator == model.NotEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind != value }, nil
		} else if operator == GreaterThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind > value }, nil
		} else if operator == GreaterThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind >= value }, nil
		} else if operator == LessThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind < value }, nil
		} else if operator == LessThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind <= value }, nil
		} else {
			return nil, errors.New(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	default:
		return nil, errors.New(ErrorResponseDetailsColumnNotFilterable)
	}
}

func (s AssetGroupMemberWithCertification) CreateGraphDbFilterCriteria(filterMap QueryParameterFilterMap) ([]graph.Criteria, error) {
	criteria := []graph.Criteria{}
	for column, filters := range filterMap {
		for _, filter := range filters {
			if conditional, err := s.BuildGraphDbFilterCriteria(column, filter.Operator, filter.Value); err != nil {
				return criteria, err
			} else {
				criteria = append(criteria, conditional)
			}
		}
	}
	return criteria, nil
}

// move this
func (s AssetGroupMemberWithCertification) BuildGraphDbFilterCriteria(column string, operator FilterOperator, value string) (graph.Criteria, error) {
	switch column {
	case "certified":
		if operator == Equals {
			return func(t AssetGroupMemberWithCertification) bool { return t.Certified == value }, nil
		} else if operator == NotEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.Certified != value }, nil
		} else {
			return nil, errors.New(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "first_seen_start":
		if operator == Equals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind == value }, nil
		} else if operator == NotEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind != value }, nil
		} else if operator == GreaterThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind > value }, nil
		} else if operator == GreaterThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind >= value }, nil
		} else if operator == LessThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind < value }, nil
		} else if operator == LessThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind <= value }, nil
		} else {
			return nil, errors.New(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	case "first_seen_end":
		if operator == Equals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind == value }, nil
		} else if operator == NotEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind != value }, nil
		} else if operator == GreaterThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind > value }, nil
		} else if operator == GreaterThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind >= value }, nil
		} else if operator == LessThan {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind < value }, nil
		} else if operator == LessThanOrEquals {
			return func(t AssetGroupMemberWithCertification) bool { return t.PrimaryKind <= value }, nil
		} else {
			return nil, errors.New(ErrorResponseDetailsFilterPredicateNotSupported)
		}
	default:
		return nil, errors.New(ErrorResponseDetailsColumnNotFilterable)
	}
}

type GetAssetGroupMembersWithCertificationResponse struct {
	Members []AssetGroupMemberWithCertification `json:"members"`
}

func nodeToAssetGroupMemberWithCertification(node *graph.Node, selectorNode model.AssetGroupSelectorNode, includeProperties bool) AssetGroupMemberWithCertification {
	member := nodeToAssetGroupMember(node, includeProperties)
	// TODO: add selectorNode fields to the member
	return member
}

func (s *Resources) GetAssetGroupTagCertifications(response http.ResponseWriter, request *http.Request) {
	var (
		requestContext         = request.Context()
		defaultSkip            = 0
		defaultLimit           = 100
		defaultStartTime = time.Now()
		defaultEndTime = time.Time()
		assetGroupMemberWithCertification = AssetGroupMemberWithCertification{}
		queryParams            = request.URL.Query()
	)
	// Parse Query Parameters
	if skip, err := ParseSkipQueryParameter(queryParams, defaultSkip); err != nil {
		api.WriteErrorResponse(requestContext, ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, defaultLimit); err != nil {
		api.WriteErrorResponse(requestContext, ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if firstSeenStartTime, err := ParseTimeQueryParameter(queryParams, api.QueryParameterStart, defaultStartTime); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, api.QueryParameterStart, err), response)
	} else if firstSeenEndTime, err := ParseTimeQueryParameter(queryParams, api.QueryParameterEnd, defaultStartTime); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, api.QueryParameterEnd, err), response)
		// TODO -- what filters to ignore here? 
	} else if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request, api.QueryParameterStart, api.QueryParameterEnd, "query"); err != nil {
		api.WriteErrorResponse(requestContext, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else {
		// TODO maybe -- this is repeated in other places, good opportunity to DRY
		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(assetGroupMemberWithCertification, name); err != nil {
				api.WriteErrorResponse(requestContext, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for _, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(requestContext, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
				}
			}

		}

		// SQL Filters for zoneId only
		sqlFilter := model.SQLFilter{
			SQLString: "",
		}

		if assetGroupTagId := request.URL.Query().Get(api.QueryParameterAssetGroupTagId); assetGroupTagId != "" {
			sqlFilter.SQLString = fmt.Sprintf(" AND s.asset_group_tag_id = %s ", assetGroupTagId)
		}

		// DB call: AGT selector node table returns an array of AssetGroupSelectorNode
		if selectorNodes, err := s.DB.GetSelectorNodes(requestContext, sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			// de-dupe the results based on highest certification status / earliest last seen date
			selectorNodeMap := make(map[graph.ID]model.AssetGroupSelectorNode)
			for _, selectorNode := range selectorNodes {
				if mapNode, nodeExists := selectorNodeMap[selectorNode.NodeId]; nodeExists {
					// set the date to be the earliest first seen
					if selectorNode.CreatedAt.Before(mapNode.CreatedAt) {
						mapNode.CreatedAt = selectorNode.CreatedAt
					}
					// set the certification to the highest value of the two
					if selectorNode.Certified > mapNode.Certified {
						mapNode.Certified = selectorNode.Certified
						mapNode.CertifiedBy = selectorNode.CertifiedBy
					}
				} else {
					selectorNodeMap[selectorNode.NodeId] = selectorNode
				}
			}

			// post-processing to get only the nodes that match the certification status, start/end, and certified by filter (if present)
			if _, ok := queryFilters["certification_status"]; !ok {
				queryFilters.AddFilter(model.QueryParameterFilter{
					Name: "certification_status",
					Operator: model.Equals,
					Value: model.AssetGroupCertificationNone,
					IsStringData: true,
				})
			}

			queryFilters.AddFilter(model.QueryParameterFilterMap{
				Name: api.QueryParameterStart, 
				Operator: model.GreaterThanOrEquals,
				Value: firstSeenStartTime,
				IsStringData: false,
			})

			queryFilters.AddFilter(model.QueryParameterFilterMap{
				Name: api.QueryParameterEnd, 
				Operator: model.LessThan,
				Value: firstSeenEndTime,
				IsStringData: false,
			})

			// TODO these filtering conditionals should only be cert status, start/end. and certified by
			conditionals, err := assetGroupMemberWithCertification.CreateFilteringConditionals(queryFilters)

			var nodeIds []graph.ID
			for _, node := range selectorNodeMap {
				if node.MeetsFilterRequirements(conditionals) {
					nodeIds = append(nodeIds, node.NodeId)
				}
			}
			// Call DAWGS to further filter using the nodeIds from the DB
			// TODO -- add fuzzy search (name or objectId), env_id (this will be an OR), and kind filtering

			// this is how you query DAWGS for env ID
			//  query.Or(
			// 	query.Equals(query.NodeProperty(ad.DomainSID.String()), envID),
			// 	query.Equals(query.NodeProperty(azure.TenantID.String()), envID),
			// ))

			//return query.And(
						// query.Kind(query.Node(), kind),
						// query.Or(
						// 	query.StringStartsWith(query.NodeProperty(common.Name.String()), strings.ToUpper(searchValue)),
						// 	query.StringStartsWith(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(searchValue)),
						// ),

			dawgsFilterCriteria, err := assetGroupMemberWithCertification.CreateGraphDBFilterCriteria(queryFilters)

			if nodes, err := s.GraphQuery.GetFilteredAndSortedNodesPaginated(sort, query.And(
				// put these in a variable and then dynamically create filters and spread
				// then you can add filters dynamically
				query.KindIn(query.Node(), assetGroupTag.ToKind()),
				query.InIDs(query.NodeID(), nodeIds...), ...dawgsFilterCriteria), skip, limit); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting members: %v", err), request), response)
			} else if count, err := s.GraphQuery.CountFilteredNodes(request.Context(), dawgsFilterCriteria); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting member count: %v", err), request), response)
			} else {
				// return paginated AssetGroupMemberWithCertification of matches and also a count
				var members []api.AssetGroupMemberWithCertification
				for _, node := range nodes {
					// TODO: nodeToAssetGroupMember does not incude all fields required for return type, create new similar function that takes in the DB response and the DAWGS response and combines them into an  AssetGroupMemberWithCertification
					members = append(members, nodeToAssetGroupMemberWithCertification(node, selectorNodeMap[node.ID], excludeProperties))
				}
				api.WriteResponseWrapperWithPagination(request.Context(), GetAssetGroupMembersWithCertificationResponse{Members: members}, limit, skip, int(count), http.StatusOK, response)
			}

		}

		// fields in array of response matches:
		// 	Object Type (node type) <- nodeToAssetGroupMember helper function
		// 	Object Name (node name) <- nodeToAssetGroupMember helper function
		// 	Environment (DAWGS)
		// 	Zone (asset group tag ID in the AGT table)
		// 	First Seen (AGT selector node table -- server side )
		// 	Certified enum (AGT selector node table)
		// 	Certified by (AGT selector node table) <- probably not going to do this
	}

}
