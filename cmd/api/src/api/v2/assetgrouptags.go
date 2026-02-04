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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

const (
	assetGroupPreviewSelectorDefaultLimit = 200
	AssetGroupTagDefaultLimit             = 50
	assetGroupTagQueryLimitMin            = 3

	includeProperties = true
	excludeProperties = false
)

type AssetGroupTagView struct {
	model.AssetGroupTag
	Counts *model.AssetGroupTagCounts `json:"counts,omitempty"`
}

type GetAssetGroupTagsResponse struct {
	Tags []AssetGroupTagView `json:"tags"`
}

type assetGroupTagSelectorRequest struct {
	model.AssetGroupTagSelector
	AutoCertify *model.SelectorAutoCertifyMethod `json:"auto_certify"`
	Description *string                          `json:"description"`
	Disabled    *bool                            `json:"disabled"`
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
				assetGroupTagCountsMap = make(model.AssetGroupTagCountsMap, len(tags))
			)

			if paramIncludeCounts {
				ids := make([]int, 0, len(tags))
				for i := range tags {
					ids = append(ids, tags[i].ID)
				}
				if assetGroupTagCountsMap, err = s.DB.GetAssetGroupTagSelectorCounts(rCtx, ids); err != nil {
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
						counts := assetGroupTagCountsMap[tag.ID]
						counts.Members = n
						tview.Counts = &counts
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

func validateAutoCertifyInput(assetGroupTag model.AssetGroupTag, autoCertify *model.SelectorAutoCertifyMethod) error {
	if autoCertify == nil {
		return nil
	}

	if assetGroupTag.Type != model.AssetGroupTagTypeTier && *autoCertify != 0 {
		return fmt.Errorf(api.ErrorResponseAssetGroupAutoCertifyOnlyAvailableForPrivilegeZones)
	}

	switch *autoCertify {
	case model.SelectorAutoCertifyMethodDisabled, model.SelectorAutoCertifyMethodAllMembers, model.SelectorAutoCertifyMethodSeedsOnly:
		return nil
	default:
		return fmt.Errorf(api.ErrorResponseAssetGroupAutoCertifyInvalid)
	}
}

func (s *Resources) CreateAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	var (
		createSelectorRequest assetGroupTagSelectorRequest
		assetTagIdStr         = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Create")()

	if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := json.NewDecoder(request.Body).Decode(&createSelectorRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if errs := validation.Validate(createSelectorRequest.AssetGroupTagSelector); len(errs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if err := validateSelectorSeeds(s.GraphQuery, createSelectorRequest.Seeds); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else {
		// defaults for optional pointer field request values
		autoCertify := model.SelectorAutoCertifyMethodDisabled
		if createSelectorRequest.AutoCertify != nil {
			if err := validateAutoCertifyInput(assetGroupTag, createSelectorRequest.AutoCertify); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			}
			autoCertify = *createSelectorRequest.AutoCertify
		}

		description := ""
		if createSelectorRequest.Description != nil {
			description = *createSelectorRequest.Description
		}

		if selector, err := s.DB.CreateAssetGroupTagSelector(request.Context(), assetTagId, actor, createSelectorRequest.Name, description, false, true, autoCertify, createSelectorRequest.Seeds); errors.Is(err, database.ErrDuplicateAGTagSelectorName) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseAssetGroupTagSelectorDuplicateName, request), response)
		} else if err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if config, err := appcfg.GetScheduledAnalysisParameter(request.Context(), s.DB); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			if !config.Enabled {
				if err := s.DB.RequestAnalysis(request.Context(), actor.ID.String()); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				}
			}
			api.WriteBasicResponse(request.Context(), selector, http.StatusCreated, response)
		}
	}
}

func (s *Resources) UpdateAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	var (
		selUpdateReq  assetGroupTagSelectorRequest
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
		rawSelectorID = mux.Vars(request)[api.URIPathVariableAssetGroupTagSelectorID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Update")()

	if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
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

		// we can update AutoCertify on a default selector (as long as the selector is not tied to label)
		if selUpdateReq.AutoCertify != nil {
			if err := validateAutoCertifyInput(assetGroupTag, selUpdateReq.AutoCertify); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			}
			selector.AutoCertify = *selUpdateReq.AutoCertify
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
		queryParams = request.URL.Query()

		assetTagIdStr  = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
		environmentIds = queryParams[api.QueryParameterEnvironments]

		selectorQueryFilter      = make(model.QueryParameterFilterMap)
		selectorSeedsQueryFilter = make(model.QueryParameterFilterMap)
		selectorSeed             = model.SelectorSeed{}
		assetGroupTagSelector    = model.AssetGroupTagSelector{}
	)

	if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, AssetGroupTagDefaultLimit); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if paramIncludeCounts, err := api.ParseOptionalBool(queryParams.Get(api.QueryParameterIncludeCounts), false); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specified for include counts", request), response)
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
		} else if sort, err := api.ParseSortParameters(model.AssetGroupTagSelector{}, queryParams); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
		} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), assetGroupTagID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if selectors, count, err := s.DB.GetAssetGroupTagSelectorsByTagIdFilteredAndPaginated(request.Context(), assetGroupTagID, selectorSqlFilter, selectorSeedSqlFilter, sort, skip, limit); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			var (
				resp = GetAssetGroupTagSelectorResponse{
					Selectors: make([]AssetGroupTagSelectorView, 0, len(selectors)),
				}
				filter = model.SQLFilter{}
			)

			if assetGroupTag.RequireCertify.ValueOrZero() {
				filter.SQLString = " AND certified > ?"
				filter.Params = append(filter.Params, model.AssetGroupCertificationRevoked)
			}
			for _, selector := range selectors {
				selectorView := AssetGroupTagSelectorView{AssetGroupTagSelector: selector}
				if paramIncludeCounts {
					memberCount := int64(0)
					// if the selector is not disabled
					if selector.DisabledAt.Time.IsZero() {
						// get all the nodes which are selected
						if selectorNodes, _, err := s.DB.GetSelectorNodesBySelectorIdsFilteredAndPaginated(request.Context(), filter, model.Sort{}, 0, 0, selector.ID); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							nodeIds := make([]graph.ID, 0, len(selectorNodes))
							for _, node := range selectorNodes {
								nodeIds = append(nodeIds, node.NodeId)
							}

							// only count nodes that are actually tagged
							filters := []graph.Criteria{
								query.KindIn(query.Node(), assetGroupTag.ToKind()),
								query.InIDs(query.NodeID(), nodeIds...),
							}
							if len(environmentIds) > 0 {
								filters = append(filters, query.Or(
									query.In(query.NodeProperty(ad.DomainSID.String()), environmentIds),
									query.In(query.NodeProperty(azure.TenantID.String()), environmentIds),
								))
							}

							if count, err := s.GraphQuery.CountFilteredNodes(request.Context(), query.And(filters...)); err != nil {
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

type assetGroupTagUpdateRequest struct {
	Name            *string     `json:"name"`
	Description     *string     `json:"description"`
	Position        null.Int32  `json:"position"`
	RequireCertify  null.Bool   `json:"require_certify"`
	AnalysisEnabled null.Bool   `json:"analysis_enabled"`
	Glyph           null.String `json:"glyph"`
}

func HasValidTagName(assetGroupTagName string) bool {
	validNameRegex := regexp.MustCompile("^[a-zA-Z0-9 _]+$")
	return validNameRegex.MatchString(assetGroupTagName)
}

func CheckTagGlyph(glyph null.String, tagType model.AssetGroupTagType, tagPosition null.Int32) error {
	if !glyph.Valid {
		return nil
	}

	if tagType != model.AssetGroupTagTypeTier {
		return fmt.Errorf("only zones support custom glyphs")
	}

	if tagPosition.ValueOrZero() == 1 {
		return fmt.Errorf("tier zero glyph cannot be modified")
	}

	// don't allow the glyph to be set to something that resembles tier zero or owned.
	if glyph.String == model.TierZeroGlyph || glyph.String == model.OwnedGlyph {
		return fmt.Errorf("glyphs similar to tier zero or owned not allowed")
	}

	return nil
}

func (s *Resources) UpdateAssetGroupTag(response http.ResponseWriter, request *http.Request) {
	var (
		tagUpdates    assetGroupTagUpdateRequest
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Update")()

	if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.ErrorContext(request.Context(), "Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if tag, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := json.NewDecoder(request.Body).Decode(&tagUpdates); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else {
		fieldMatched := false
		tagUpdated := false
		kindRefreshNeeded := false
		analysisNeeded := false

		if tagUpdates.Name != nil {
			fieldMatched = true
			if *tagUpdates.Name == "" {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "name can not be empty", request), response)
				return
			} else if (tag.Type == model.AssetGroupTagTypeTier && tag.Position.ValueOrZero() == 1) || tag.Type == model.AssetGroupTagTypeOwned {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "renaming default tags is forbidden currently", request), response)
				return
			} else if !HasValidTagName(*tagUpdates.Name) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseAssetGroupTagInvalidTagName, request), response)
				return
			}
			if tag.Name != *tagUpdates.Name {
				tagUpdated = true
				kindRefreshNeeded = true
				tag.Name = *tagUpdates.Name
			}
		}

		if tagUpdates.Description != nil {
			fieldMatched = true
			if tag.Description != *tagUpdates.Description {
				tagUpdated = true
				tag.Description = *tagUpdates.Description
			}
		}

		if tagUpdates.RequireCertify.Valid {
			if tag.Type != model.AssetGroupTagTypeTier {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "only zones support modifying require_certify", request), response)
				return
			}
			fieldMatched = true
			if !tag.RequireCertify.Equal(tagUpdates.RequireCertify) {
				analysisNeeded = true
				tagUpdated = true
				tag.RequireCertify = tagUpdates.RequireCertify
			}
		}
		// Ensure require certify is only toggle-able on BHE
		s.DB.SanitizeUpdateAssetGroupTagRequireCertify(&tag)

		if tagUpdates.Position.Valid {
			if tag.Type != model.AssetGroupTagTypeTier {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "only zones support modifying position", request), response)
				return
			} else if tag.Position.ValueOrZero() == 1 {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "tier zero position cannot be modified", request), response)
				return
			}
			fieldMatched = true
			if !tag.Position.Equal(tagUpdates.Position) {
				analysisNeeded = true
				tagUpdated = true
				tag.Position = tagUpdates.Position
			}
		}

		if tagUpdates.AnalysisEnabled.Valid {
			if tag.Type != model.AssetGroupTagTypeTier {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "only zones support modifying analysis_enabled", request), response)
				return
			} else if tag.Position.ValueOrZero() == 1 {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "tier zero analysis_enabled cannot be modified", request), response)
				return
			} else if !s.DogTags.GetFlagAsBool(dogtags.PZ_MULTI_TIER_ANALYSIS) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "multi-tier analysis is not enabled for privilege zones", request), response)
				return
			}

			fieldMatched = true
			if !tag.AnalysisEnabled.Equal(tagUpdates.AnalysisEnabled) {
				tagUpdated = true
				analysisNeeded = true
				tag.AnalysisEnabled = tagUpdates.AnalysisEnabled
			}
		}

		if tagUpdates.Glyph.Valid {
			// Ensure no empty string glyphs by setting it to null
			if tagUpdates.Glyph.ValueOrZero() == "" {
				tagUpdates.Glyph = null.String{}
			}
			if err := CheckTagGlyph(tagUpdates.Glyph, tag.Type, tag.Position); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			}
			fieldMatched = true
			if !tag.Glyph.Equal(tagUpdates.Glyph) {
				tagUpdated = true
				tag.Glyph = tagUpdates.Glyph
			}
		}

		if !fieldMatched {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "no valid fields specified", request), response)
		} else if !tagUpdated {
			// return tag as-is since it's unmodified
			api.WriteBasicResponse(request.Context(), tag, http.StatusOK, response)
		} else if updatedTag, err := s.DB.UpdateAssetGroupTag(request.Context(), actor, tag); errors.Is(err, database.ErrDuplicateAGName) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseAssetGroupTagDuplicateKindName, request), response)
		} else if errors.Is(err, database.ErrPositionOutOfRange) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseAssetGroupTagPositionOutOfRange, request), response)
		} else if errors.Is(err, database.ErrDuplicateGlyph) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseAssetGroupTagDuplicateGlyph, request), response)
		} else if err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			// Request analysis if scheduled analysis isn't enabled
			if analysisNeeded {
				if config, err := appcfg.GetScheduledAnalysisParameter(request.Context(), s.DB); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				} else if !config.Enabled {
					if err := s.DB.RequestAnalysis(request.Context(), actor.ID.String()); err != nil {
						api.HandleDatabaseError(request, response, err)
						return
					}
				}
			}

			if kindRefreshNeeded {
				// Because the graph pg driver relies on in-memory kind maps, it's required to refresh the map in order to remove the recently deleted kind
				if err := s.Graph.RefreshKinds(request.Context()); err != nil {
					slog.WarnContext(request.Context(), "AGT: refreshing schemaManager in-memory kind maps failed", attr.Error(err))
				}
			}
			api.WriteBasicResponse(request.Context(), updatedTag, http.StatusOK, response)
		}
	}
}

type GetAssetGroupMemberCountsResponse struct {
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
			return
		}

		data := GetAssetGroupMemberCountsResponse{
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

	Source          model.AssetGroupSelectorNodeSource `json:"source,omitempty"`
	AssetGroupTagId int                                `json:"asset_group_tag_id,omitempty"`
}

// Used to minimize the response shape to just the necessary member display fields
func nodeToAssetGroupMember(node *graph.Node, includeProperties bool) AssetGroupMember {
	primaryKind, displayName, objectId, envId := model.GetAssetGroupMemberProperties(node)

	member := AssetGroupMember{
		NodeId:        node.ID,
		ObjectID:      objectId,
		EnvironmentID: envId,
		PrimaryKind:   primaryKind,
		Name:          displayName,
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

func (s AssetGroupMember) ValidFilters() map[string][]model.FilterOperator {
	return map[string][]model.FilterOperator{
		"name":         {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"object_id":    {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"primary_kind": {model.Equals, model.NotEquals},
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
		groupMember := nodeToAssetGroupMember(node, includeProperties)
		groupMember.AssetGroupTagId = assetGroupTagID
		api.WriteBasicResponse(request.Context(), MemberInfoResponse{Member: memberInfo{groupMember, selectors}}, http.StatusOK, response)
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
		queryFilterMap = make(model.QueryParameterFilterMap)
	)

	if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if tagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), tagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if sort, err := api.ParseGraphSortParameters(AssetGroupMember{}, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, AssetGroupTagDefaultLimit); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else {
		if len(sort) == 0 {
			sort = query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}
		}
		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(AssetGroupMember{}, name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for _, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilterMap.AddFilter(filter)
				}
			}
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

		// Bespoke build the filter for the graph db due to complexity
		if graphDbFilters, err := buildAssetGroupMembersByTagGraphDbFilters(request.Context(), s.DB, queryFilterMap); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		} else if len(graphDbFilters) > 0 {
			filters = append(filters, graphDbFilters...)
		}

		if nodes, err := s.GraphQuery.GetFilteredAndSortedNodesPaginated(sort, query.And(filters...), skip, limit); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting members: %v", err), request), response)
		} else if count, err := s.GraphQuery.CountFilteredNodes(request.Context(), query.And(filters...)); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting member count: %v", err), request), response)
		} else {
			for _, node := range nodes {
				groupMember := nodeToAssetGroupMember(node, excludeProperties)
				groupMember.AssetGroupTagId = assetGroupTag.ID
				members = append(members, groupMember)
			}

			api.WriteResponseWrapperWithPagination(request.Context(), GetAssetGroupMembersResponse{Members: members}, limit, skip, int(count), http.StatusOK, response)
		}
	}
}

func buildAssetGroupMembersByTagGraphDbFilters(ctx context.Context, db database.Database, queryFilterMap model.QueryParameterFilterMap) ([]graph.Criteria, error) {
	var (
		filters          []graph.Criteria
		sourceKindsMap   = make(map[string]bool)
		primaryNodeKinds graph.Kinds
	)

	for _, queryFilters := range queryFilterMap {
		var innerFilters []graph.Criteria

		for _, queryFilter := range queryFilters {
			var (
				ref                       graph.Criteria
				isSourcePrimaryKindFilter bool
			)

			switch queryFilter.Name {
			case "object_id":
				ref = query.NodeProperty(common.ObjectID.String())
				if queryFilter.Operator == model.ApproximatelyEquals {
					ref = query.StringContains(ref, queryFilter.Value)
				} else {
					ref = query.Equals(ref, queryFilter.Value)
				}
			case "name":
				ref = query.NodeProperty(common.Name.String())
				if queryFilter.Operator == model.ApproximatelyEquals {
					ref = query.StringContains(ref, queryFilter.Value)
				} else {
					ref = query.Equals(ref, queryFilter.Value)
				}
			case "primary_kind":
				if len(sourceKindsMap) == 0 {
					if sourceKinds, err := db.GetSourceKinds(ctx); err != nil {
						return filters, err
					} else {
						for _, kind := range sourceKinds {
							sourceKindsMap[kind.Name.String()] = true
						}
					}
				}

				// In the case that the primary_kind filter(s) is of a source kind type such as Base, AZBase, etc.
				if sourceKindsMap[queryFilter.Value] {
					isSourcePrimaryKindFilter = true
					// This ensures unnecessary db calls so that we can obtain all the graph schema node kinds only once so that primaryKindGraphFilters can be efficiently populated and filtered
					if len(primaryNodeKinds) == 0 {
						if graphSchemaNodeKinds, _, err := db.GetGraphSchemaNodeKinds(ctx, model.Filters{}, model.Sort{}, 0, 0); err != nil {
							return filters, err
						} else {
							for _, schemaNodeKind := range graphSchemaNodeKinds {
								if !sourceKindsMap[schemaNodeKind.Name] {
									primaryNodeKinds = append(primaryNodeKinds, schemaNodeKind.ToKind())
								}
							}
						}
					}

					var sourceKindRef graph.Criteria
					switch queryFilter.Operator {
					case model.Equals:
						sourceKindRef = query.KindIn(query.Node(), graph.StringKind(queryFilter.Value))
					// if the neq: operator is used, we want to filter to the other source kinds, still skipping tag kinds
					case model.NotEquals:
						var filteredSourceKinds []string

						for sourceKind := range sourceKindsMap {
							if queryFilter.Value != sourceKind {
								filteredSourceKinds = append(filteredSourceKinds, sourceKind)
							}
						}

						sourceKindRef = query.KindIn(query.Node(), graph.StringsToKinds(filteredSourceKinds)...)
					}

					ref = query.And(
						sourceKindRef,
						query.Not(query.KindIn(query.Node(), primaryNodeKinds...)),
					)
				} else {
					// All other node kinds and not source kinds
					ref = query.KindIn(query.Node(), graph.StringKind(queryFilter.Value))
				}
			}

			if !isSourcePrimaryKindFilter && queryFilter.Operator == model.NotEquals {
				ref = query.Not(ref)
			}

			innerFilters = append(innerFilters, ref)
		}

		filters = append(filters, query.Or(innerFilters...))
	}

	return filters, nil
}

func (s *Resources) GetAssetGroupMembersBySelector(response http.ResponseWriter, request *http.Request) {
	var (
		members               = []AssetGroupMember{}
		sqlFilter             = model.SQLFilter{}
		queryParams           = request.URL.Query()
		environmentIds        = queryParams[api.QueryParameterEnvironments]
		translatedQueryFilter = make(model.QueryParameterFilterMap)
	)
	if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if assetTagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if assetGroupTag, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectorId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagSelectorID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if selector, err := s.DB.GetAssetGroupTagSelectorBySelectorId(request.Context(), selectorId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if sort, err := api.ParseGraphSortParameters(AssetGroupMember{}, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, AssetGroupTagDefaultLimit); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if selector.AssetGroupTagId != assetTagId {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "selector is not part of asset group tag", request), response)
	} else if selector.DisabledAt.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, "selector is disabled", request), response)
	} else {
		if len(sort) == 0 {
			sort = query.SortItems{{SortCriteria: query.NodeID(), Direction: query.SortDirectionAscending}}
		}

		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(model.AssetGroupSelectorNode{}, name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					filter.SetOperator = model.FilterOr

					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					// some of the API filter names do not match the DB column names - so we have to do a translation here
					originalName := filter.Name
					switch filter.Name {
					case "name":
						filter.Name = "node_name"
					case "object_id":
						filter.Name = "node_object_id"
					case "primary_kind":
						filter.Name = "node_primary_kind"
					}
					translatedQueryFilter.AddFilter(filter)
					translatedQueryFilter[filter.Name][i].IsStringData = model.AssetGroupSelectorNode{}.IsStringColumn(originalName)
				}
			}
		}

		if sqlFilter, err = translatedQueryFilter.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		}

		if sqlFilter.SQLString != "" {
			sqlFilter.SQLString = "AND " + sqlFilter.SQLString
		}

		if len(environmentIds) > 0 {
			sqlFilter.SQLString += " AND node_environment_id in ?"
			sqlFilter.Params = append(sqlFilter.Params, environmentIds)
		}

		if assetGroupTag.RequireCertify.ValueOrZero() {
			sqlFilter.SQLString += " AND certified > ?"
			sqlFilter.Params = append(sqlFilter.Params, model.AssetGroupCertificationRevoked)
		}

		// In order to get an accurate count, this needs to grab the entire selector node record space
		if selectorNodes, _, err := s.DB.GetSelectorNodesBySelectorIdsFilteredAndPaginated(request.Context(), sqlFilter, model.Sort{}, 0, 0, selectorId); err != nil {
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

			filters := []graph.Criteria{
				query.KindIn(query.Node(), assetGroupTag.ToKind()),
				query.InIDs(query.NodeID(), nodeIds...),
			}

			if nodes, err := s.GraphQuery.GetFilteredAndSortedNodesPaginated(sort, query.And(filters...), skip, limit); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting members: %v", err), request), response)
			} else if count, err := s.GraphQuery.CountFilteredNodes(request.Context(), query.And(filters...)); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting member count: %v", err), request), response)
			} else {
				for _, node := range nodes {
					member := nodeToAssetGroupMember(node, false)
					member.AssetGroupTagId = assetGroupTag.ID
					member.Source = sourceByNodeId[node.ID]
					members = append(members, member)
				}

				api.WriteResponseWrapperWithPagination(request.Context(), GetAssetGroupMembersResponse{Members: members}, limit, skip, int(count), http.StatusOK, response)
			}
		}
	}
}

func validateAssetGroupExpansionMethodWithFallback(maybeMethod *model.AssetGroupExpansionMethod) (model.AssetGroupExpansionMethod, error) {
	if maybeMethod == nil {
		return model.AssetGroupExpansionMethodAll, nil
	}
	switch *maybeMethod {
	case model.AssetGroupExpansionMethodNone, model.AssetGroupExpansionMethodAll, model.AssetGroupExpansionMethodChildren, model.AssetGroupExpansionMethodParents:
		return *maybeMethod, nil
	default:
		return 0, fmt.Errorf("invalid expansion method")
	}
}

type PreviewSelectorBody struct {
	Seeds     model.SelectorSeeds              `json:"seeds" validate:"required"`
	Expansion *model.AssetGroupExpansionMethod `json:"expansion"`
}

func (s *Resources) PreviewSelectors(response http.ResponseWriter, request *http.Request) {
	var (
		body    PreviewSelectorBody
		members = []AssetGroupMember{}
	)

	if limit, err := ParseLimitQueryParameter(request.URL.Query(), assetGroupPreviewSelectorDefaultLimit); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if errs := validation.Validate(body); len(errs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if _, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if err := validateSelectorSeeds(s.GraphQuery, body.Seeds); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if expansion, err := validateAssetGroupExpansionMethodWithFallback(body.Expansion); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if nodes, errs := datapipe.FetchNodesFromSeeds(request.Context(), appcfg.GetAGTParameters(request.Context(), s.DB), s.Graph, body.Seeds, expansion, limit); len(errs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		for _, node := range nodes {
			if node.Node != nil {
				member := nodeToAssetGroupMember(node.Node, excludeProperties)
				member.Source = node.Source

				members = append(members, member)
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
	Query           string                  `json:"query"`
	TagType         model.AssetGroupTagType `json:"tag_type"`
	AssetGroupTagId *int                    `json:"asset_group_tag_id,omitempty"`
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
		sqlFilter   model.SQLFilter
	)

	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if !validateAssetGroupTagType(reqBody.TagType) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseAssetGroupTagInvalid, request), response)
	} else if len(reqBody.Query) < assetGroupTagQueryLimitMin {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsQueryTooShort, request), response)
	} else {
		var (
			kinds  graph.Kinds
			tagIds []int
			tags   model.AssetGroupTags
		)

		if reqBody.AssetGroupTagId != nil {
			sqlFilter.SQLString = "id = ?"
			sqlFilter.Params = append(sqlFilter.Params, *reqBody.AssetGroupTagId)
		}

		if tags, err = s.DB.GetAssetGroupTags(request.Context(), sqlFilter); err != nil && !errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
			return
		}

		tagIdByKind := make(map[graph.Kind]int)

		for _, t := range tags {
			// owned tag is a label despite distinct designation
			if reqBody.TagType == t.Type || (reqBody.TagType == model.AssetGroupTagTypeLabel && t.Type == model.AssetGroupTagTypeOwned) {

				// filter the below node and selector query to tag type
				kinds = kinds.Add(t.ToKind())
				tagIds = append(tagIds, t.ID)
				tagIdByKind[t.ToKind()] = t.ID
				if strings.Contains(strings.ToLower(t.Name), strings.ToLower(reqBody.Query)) && len(matchedTags) < AssetGroupTagDefaultLimit {
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

		if selectors, err = s.DB.GetAssetGroupTagSelectors(request.Context(), selectorFilter, AssetGroupTagDefaultLimit); err != nil && !errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
			return
		} else if nodes, err := s.GraphQuery.GetFilteredAndSortedNodesPaginated(query.SortItems{{SortCriteria: query.NodeProperty(common.Name.String()), Direction: query.SortDirectionAscending}}, nodeFilter, 0, AssetGroupTagDefaultLimit); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting members: %v", err), request), response)
			return
		} else {
			for _, node := range nodes {
				groupMember := nodeToAssetGroupMember(node, excludeProperties)
				for _, kind := range node.Kinds {
					// Find the first valid kind for this search type and attribute it to this member
					if tagId, ok := tagIdByKind[kind]; ok {
						groupMember.AssetGroupTagId = tagId
						break
					}
				}
				members = append(members, groupMember)
			}
		}

		api.WriteBasicResponse(request.Context(), SearchAssetGroupTagsResponse{Tags: matchedTags, Selectors: selectors, Members: members}, http.StatusOK, response)
	}
}

type AssetGroupHistoryResp struct {
	Records []model.AssetGroupHistory `json:"records"`
}

func (s *Resources) assetGroupTagHistoryImplementation(response http.ResponseWriter, request *http.Request, query string) {
	var (
		rCtx        = request.Context()
		queryParams = request.URL.Query()
		sort        model.Sort
		sqlFilter   model.SQLFilter
	)

	if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(rCtx, ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
	} else if limit, err := ParseOptionalLimitQueryParameter(queryParams, AssetGroupTagDefaultLimit); err != nil {
		api.WriteErrorResponse(rCtx, ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
	} else if sort, err = api.ParseSortParameters(model.AssetGroupHistory{}, queryParams); err != nil {
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

		if sqlFilter, err = queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(rCtx, api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		}

		if query != "" {
			var (
				queryableColumns  = []string{"actor", "email", "action", "target"}
				querySQL          = fmt.Sprintf("(%s ILIKE ANY(?))", strings.Join(queryableColumns, " ILIKE ANY(?) OR "))
				fuzzyQueryPattern = "%" + query + "%"
				fuzzyQueryParams  = pq.StringArray{fuzzyQueryPattern, strings.ReplaceAll(fuzzyQueryPattern, " ", "")}
			)

			if sqlFilter.SQLString != "" {
				querySQL = " AND " + querySQL
			}

			sqlFilter.SQLString += querySQL
			for range len(queryableColumns) {
				sqlFilter.Params = append(sqlFilter.Params, fuzzyQueryParams)
			}
		}

		if historyRecs, count, err := s.DB.GetAssetGroupHistoryRecords(rCtx, sqlFilter, sort, skip, limit); err != nil && !errors.Is(err, database.ErrNotFound) {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithPagination(rCtx, AssetGroupHistoryResp{Records: historyRecs}, limit, skip, count, http.StatusOK, response)
		}
	}
}

type SearchAssetGroupTagHistoryRequest struct {
	Query string `json:"query"`
}

func (s *Resources) SearchAssetGroupTagHistory(response http.ResponseWriter, request *http.Request) {
	reqBody := SearchAssetGroupTagHistoryRequest{}

	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Search History Records")()

	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if len(reqBody.Query) < assetGroupTagQueryLimitMin {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsQueryTooShort, request), response)
	} else {
		s.assetGroupTagHistoryImplementation(response, request, reqBody.Query)
	}
}

func (s *Resources) GetAssetGroupTagHistory(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Get History Records")()

	s.assetGroupTagHistoryImplementation(response, request, "")
}

func (s *Resources) GetAssetGroupSelectorMemberCountsByKind(response http.ResponseWriter, request *http.Request) {
	var (
		environmentIds = request.URL.Query()[api.QueryParameterEnvironments]
		sqlFilter      = model.SQLFilter{}
	)

	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Get Selector Counts by Kind")()

	if assetTagId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if tag, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selectorId, err := strconv.Atoi(mux.Vars(request)[api.URIPathVariableAssetGroupTagSelectorID]); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if selector, err := s.DB.GetAssetGroupTagSelectorBySelectorId(request.Context(), selectorId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if selector.AssetGroupTagId != assetTagId {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "selector is not part of asset group tag", request), response)
	} else if selector.DisabledAt.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, "selector is disabled", request), response)
	} else {
		if len(environmentIds) > 0 {
			sqlFilter.SQLString += " AND node_environment_id in ?"
			sqlFilter.Params = append(sqlFilter.Params, environmentIds)
		}

		if tag.RequireCertify.ValueOrZero() {
			sqlFilter.SQLString += " AND certified > ?"
			sqlFilter.Params = append(sqlFilter.Params, model.AssetGroupCertificationRevoked)
		}

		if selectorNodes, _, err := s.DB.GetSelectorNodesBySelectorIdsFilteredAndPaginated(request.Context(), sqlFilter, model.Sort{}, 0, 0, selector.ID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			nodeIds := make([]graph.ID, 0, len(selectorNodes))

			for _, node := range selectorNodes {
				nodeIds = append(nodeIds, node.NodeId)
			}

			if primaryNodeKindsCounts, err := s.GraphQuery.GetPrimaryNodeKindCounts(request.Context(), tag.ToKind(),
				query.InIDs(query.NodeID(), nodeIds...)); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				data := GetAssetGroupMemberCountsResponse{
					Counts: primaryNodeKindsCounts,
				}
				for _, count := range primaryNodeKindsCounts {
					data.TotalCount += count
				}

				api.WriteBasicResponse(request.Context(), data, http.StatusOK, response)
			}
		}
	}
}
