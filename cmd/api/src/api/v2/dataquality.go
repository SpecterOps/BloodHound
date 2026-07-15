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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/dawgs/graph"
)

const (
	ErrNoAccess                        string = "user does not have permission to access this environment"
	ErrNoEnvironmentId                 string = "environment_id is required"
	ErrNoTenantId                      string = "no tenant id specified in url"
	ErrNoPlatformId                    string = "no platform id specified in url"
	ErrUnknownUser                     string = "unknown user"
	FmtErrInvalidIntegerQueryParameter string = "query parameter must be an integer: %s"
	FmtErrInvalidPlatformId            string = "invalid platform id specified in url: %v"
)

func (s Resources) GetDatabaseCompleteness(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasureWithThreshold(request.Context(), slog.LevelDebug, "Get Current Database Completeness")()

	result := make(map[string]float64)

	if err := s.Graph.ReadTransaction(request.Context(), func(tx graph.Transaction) error {
		if userSessionCompleteness, err := ad.FetchUserSessionCompleteness(tx); err != nil {
			return err
		} else {
			result["LocalGroupCompleteness"] = userSessionCompleteness
		}

		if localGroupCompleteness, err := ad.FetchLocalGroupCompleteness(tx); err != nil {
			return err
		} else {
			result["SessionCompleteness"] = localGroupCompleteness
		}

		return nil
	}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting quality stat: %v", err), request), response)
	} else {
		api.WriteBasicResponse(request.Context(), result, http.StatusOK, response)
	}
}

func (s *Resources) GetADDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		adDataQualityStats       model.ADDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	if order, _, err := parseOrder(queryParams, adDataQualityStats); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if id, hasDomainID := mux.Vars(request)[api.URIPathVariableDomainID]; !hasDomainID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorNoDomainId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetADDataQualityStats(request.Context(), id, start, end, order, limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetAzureDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	if order, _, err := parseOrder(queryParams, azureDataQualityStats); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if id, hasTenantID := mux.Vars(request)[api.URIPathVariableTenantID]; !hasTenantID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoTenantId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetAzureDataQualityStats(request.Context(), id, start, end, order, limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetPlatformAggregateStats(response http.ResponseWriter, request *http.Request) {
	var (
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	// TODO: This is currently using only the Azure stat type, but should check against the appropriate aggregate type for the chosen platform
	// It's safe for now, but should be refactored
	if order, _, err := parseOrder(queryParams, azureDataQualityStats); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if id, hasPlatformID := mux.Vars(request)[api.URIPathVariablePlatformID]; !hasPlatformID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoPlatformId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else {
		var (
			stats any
			count int
		)

		switch id {
		case "ad":
			stats, count, err = s.DB.GetADDataQualityAggregations(request.Context(), start, end, order, limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		case "azure":
			stats, count, err = s.DB.GetAzureDataQualityAggregations(request.Context(), start, end, order, limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(FmtErrInvalidPlatformId, id), request), response)
			return
		}

		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		ctx                      = request.Context()
		dataQualityStats         model.DataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	if environmentId := queryParams.Get(api.QueryParameterEnvironmentId); environmentId == "" {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, ErrNoEnvironmentId, request), response)
	} else if user, found := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx); !found {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, ErrUnknownUser, request), response)
	} else if _, sort, err := parseOrder(queryParams, dataQualityStats); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else {
		if ShouldFilterForETAC(s.DogTags, user) {
			hasAccess, err := CheckUserAccessToEnvironments(ctx, s.DB, user, environmentId)
			if err != nil {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
				return
			} else if !hasAccess {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusForbidden, ErrNoAccess, request), response)
				return
			}
		}
		filters := model.Filters{
			"environment_id": []model.Filter{{
				Value:       environmentId,
				Operator:    model.Equals,
				SetOperator: model.FilterAnd,
			}},
		}
		filters["created_at"] = dataQualityCreatedAtFilters(start, end)
		if stats, count, err := s.DB.GetDataQualityStats(ctx, filters, sort, skip, limit); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithTimeWindowAndPagination(ctx, stats, start, end, limit, skip, count, http.StatusOK, response)
		}
	}
}

// GetDataQualityAggregations returns data quality metric counts aggregated per environment kind for an OpenGraph extension.
// Requires a schema_environment_kind_id, with optional schema_extension_id, created_at range, sorting, and pagination.
func (s *Resources) GetDataQualityAggregations(response http.ResponseWriter, request *http.Request) {
	var (
		ctx         = request.Context()
		queryParams = request.URL.Query()

		schemaEnvironmentKindIDParam = "schema_environment_kind_id"
		schemaExtensionIDParam       = "schema_extension_id"

		err         error
		sortItems   model.Sort
		skip, limit int
		start, end  time.Time

		defaultEnd, defaultStart = DefaultTimeRange()
	)

	// schema_environment_kind_id is required
	kindID := queryParams.Get(schemaEnvironmentKindIDParam)
	if kindID == "" {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest,
			fmt.Sprintf(api.FmtErrorResponseDetailsMissingRequiredQueryParameter, schemaEnvironmentKindIDParam), request),
			response)
		return
	}
	if _, err := strconv.ParseInt(kindID, 10, 32); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest,
			fmt.Sprintf(FmtErrInvalidIntegerQueryParameter, schemaEnvironmentKindIDParam), request),
			response)
		return
	}

	// parse sort_by, skip, limit, start, end
	if sortItems, err = api.ParseSortParameters(model.DataQualityAggregations{}, queryParams); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}
	if skip, err = ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}
	if limit, err = ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}
	if start, err = ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
		return
	}
	if end, err = ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
		return
	}

	// when filtering on schema_extension_id, verify the extension exists or return a 404
	extensionID := queryParams.Get(schemaExtensionIDParam)
	if extensionID != "" {
		id, err := strconv.ParseInt(extensionID, 10, 32)
		if err != nil {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest,
				fmt.Sprintf(FmtErrInvalidIntegerQueryParameter, schemaExtensionIDParam), request),
				response)
			return
		}
		if _, err := s.DB.GetGraphSchemaExtensionById(ctx, int32(id)); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		}
	}

	filters := model.Filters{
		schemaEnvironmentKindIDParam: []model.Filter{
			{Value: kindID, Operator: model.Equals, SetOperator: model.FilterAnd, IsStringData: false},
		},
	}
	if extensionID != "" {
		filters[schemaExtensionIDParam] = []model.Filter{
			{Value: extensionID, Operator: model.Equals, SetOperator: model.FilterAnd, IsStringData: false},
		}
	}

	// created_at is filtered by the start/end params, so set the time window here
	filters["created_at"] = dataQualityCreatedAtFilters(start, end)

	aggs, count, err := s.DB.GetDataQualityAggregations(ctx, filters, sortItems, skip, limit)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}
	api.WriteResponseWrapperWithTimeWindowAndPagination(ctx, aggs, start, end, limit, skip, count, http.StatusOK, response)
}

// parseOrder is a helper function which parses any sort_by query params into both the legacy sort string format and the model.Sort format. Returns an error if the columns is not sortable, or if an empty sort param is provided.
func parseOrder(queryParams url.Values, sortable api.Sortable) (string, model.Sort, error) {
	order := []string{}
	sort := model.Sort{}
	for _, column := range queryParams[api.QueryParameterSortBy] {
		if column == "" {
			return "", sort, errors.New(api.ErrorResponseEmptySortParameter)
		}

		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !sortable.IsSortable(column) {
			return "", sort, errors.New(api.ErrorResponseDetailsNotSortable)
		}

		if descending {
			order = append(order, column+" desc")
			sort = append(sort, model.SortItem{Column: column, Direction: model.DescendingSortDirection})

		} else {
			order = append(order, column)
			sort = append(sort, model.SortItem{Column: column, Direction: model.AscendingSortDirection})
		}

	}
	return strings.Join(order, ", "), sort, nil
}

// dataQualityCreatedAtFilters builds the created_at time window filter, inclusive on start and end bounds
func dataQualityCreatedAtFilters(start, end time.Time) []model.Filter {
	return []model.Filter{
		{
			Value:        start.Format(time.RFC3339Nano),
			Operator:     model.GreaterThanOrEquals,
			SetOperator:  model.FilterAnd,
			IsStringData: false,
		},
		{
			Value:        end.Format(time.RFC3339Nano),
			Operator:     model.LessThanOrEquals,
			SetOperator:  model.FilterAnd,
			IsStringData: false,
		},
	}
}
