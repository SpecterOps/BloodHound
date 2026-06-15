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
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/dawgs/graph"
)

const (
	ErrNoTenantId                                string = "no tenant id specified in url"
	ErrNoPlatformId                              string = "no platform id specified in url"
	ErrInvalidPlatformId                         string = "invalid platform id specified in url: %v"
	dataQualityQueryParameterEndDate             string = "end_date"
	dataQualityQueryParameterExtensionID         string = "extension_id"
	dataQualityQueryParameterSchemaEnvironmentID string = "schema_environment_id"
	dataQualityQueryParameterStartDate           string = "start_date"
	dataQualityStatsDefaultPaginationLimit       int    = 1000
	dataQualityStatsDefaultPaginationOffset      int    = 0
	dataQualityStatsDefaultSortableErrorText     string = api.ErrorResponseDetailsNotSortable
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
		order                    []string
		adDataQualityStats       model.ADDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	for _, column := range queryParams[api.QueryParameterSortBy] {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !adDataQualityStats.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	if id, hasDomainID := mux.Vars(request)[api.URIPathVariableDomainID]; !hasDomainID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorNoDomainId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetADDataQualityStats(request.Context(), id, start, end, strings.Join(order, ", "), limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetAzureDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		order                    []string
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	for _, column := range queryParams[api.QueryParameterSortBy] {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !azureDataQualityStats.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	if id, hasTenantID := mux.Vars(request)[api.URIPathVariableTenantID]; !hasTenantID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoTenantId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetAzureDataQualityStats(request.Context(), id, start, end, strings.Join(order, ", "), limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetOpenGraphDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		defaultEnd, defaultStart = DefaultTimeRange()
		dataQualityStats         model.DataQualityStats
		environmentID            null.String
		extensionID              null.Int32
		queryParams              = request.URL.Query()
		schemaEnvironmentID      null.Int32
		sqlSort                  []string
	)

	environmentID = null.NewString(queryParams.Get(api.QueryParameterEnvironmentId), queryParams.Get(api.QueryParameterEnvironmentId) != "")

	// schema_environment_id scopes OpenGraph stats to the selected schema environment kind.
	if sort, err := api.ParseSortParameters(dataQualityStats, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, dataQualityStatsDefaultSortableErrorText, request), response)
	} else if sqlSort, err = api.BuildSQLSort(sort, model.SortItem{}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, dataQualityStatsDefaultSortableErrorText, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, dataQualityQueryParameterStartDate, defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams[dataQualityQueryParameterStartDate]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, dataQualityQueryParameterEndDate, defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams[dataQualityQueryParameterEndDate]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, dataQualityStatsDefaultPaginationLimit); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, dataQualityStatsDefaultPaginationOffset); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if err := extensionID.UnmarshalText([]byte(queryParams.Get(dataQualityQueryParameterExtensionID))); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, dataQualityQueryParameterExtensionID, err), response)
	} else if err := schemaEnvironmentID.UnmarshalText([]byte(queryParams.Get(dataQualityQueryParameterSchemaEnvironmentID))); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, dataQualityQueryParameterSchemaEnvironmentID, err), response)
	} else if stats, count, err := s.DB.GetOpenGraphDataQualityStats(request.Context(), environmentID, extensionID, schemaEnvironmentID, start, end, strings.Join(sqlSort, ", "), limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetOpenGraphDataQualityAggregations(response http.ResponseWriter, request *http.Request) {
	var (
		dataQualityAggregations  model.DataQualityAggregations
		defaultEnd, defaultStart = DefaultTimeRange()
		extensionID              null.Int32
		queryParams              = request.URL.Query()
		schemaEnvironmentID      null.Int32
		sqlSort                  []string
	)

	// Aggregations are per schema environment kind, not just per extension.
	if sort, err := api.ParseSortParameters(dataQualityAggregations, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, dataQualityStatsDefaultSortableErrorText, request), response)
	} else if sqlSort, err = api.BuildSQLSort(sort, model.SortItem{}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, dataQualityStatsDefaultSortableErrorText, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, dataQualityQueryParameterStartDate, defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams[dataQualityQueryParameterStartDate]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, dataQualityQueryParameterEndDate, defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams[dataQualityQueryParameterEndDate]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, dataQualityStatsDefaultPaginationLimit); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, dataQualityStatsDefaultPaginationOffset); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if err := extensionID.UnmarshalText([]byte(queryParams.Get(dataQualityQueryParameterExtensionID))); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, dataQualityQueryParameterExtensionID, err), response)
	} else if err := schemaEnvironmentID.UnmarshalText([]byte(queryParams.Get(dataQualityQueryParameterSchemaEnvironmentID))); err != nil {
		api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, dataQualityQueryParameterSchemaEnvironmentID, err), response)
	} else if aggregations, count, err := s.DB.GetOpenGraphDataQualityAggregations(request.Context(), extensionID, schemaEnvironmentID, start, end, strings.Join(sqlSort, ", "), limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), aggregations, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetPlatformAggregateStats(response http.ResponseWriter, request *http.Request) {
	var (
		order                    []string
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	// TODO: This is currently using only the Azure stat type, but should check against the appropriate aggregate type for the chosen platform
	// It's safe for now, but should be refactored
	for _, column := range queryParams[api.QueryParameterSortBy] {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !azureDataQualityStats.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	if id, hasPlatformID := mux.Vars(request)[api.URIPathVariablePlatformID]; !hasPlatformID {
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
			stats, count, err = s.DB.GetADDataQualityAggregations(request.Context(), start, end, strings.Join(order, ", "), limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		case "azure":
			stats, count, err = s.DB.GetAzureDataQualityAggregations(request.Context(), start, end, strings.Join(order, ", "), limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrInvalidPlatformId, id), request), response)
			return
		}

		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}
