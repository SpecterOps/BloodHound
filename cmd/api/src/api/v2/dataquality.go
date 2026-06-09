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
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/dawgs/graph"
)

const (
	ErrNoTenantId        string = "no tenant id specified in url"
	ErrNoPlatformId      string = "no platform id specified in url"
	ErrInvalidPlatformId string = "invalid platform id specified in url: %v"

	dataQualityQueryParameterEnvironmentKind string = "environment_kind"
	dataQualityQueryParameterNodeKind        string = "node_kind"
	dataQualityQueryParameterRunID           string = "run_id"
	dataQualityQueryParameterSourceKind      string = "source_kind"

	dataQualityErrorLatestWithRunID string = "latest cannot be combined with run_id"
)

type dataQualitySortable interface {
	IsSortable(column string) bool
}

func buildDataQualityOrder(sortByValues []string, sortable dataQualitySortable) ([]string, bool) {
	var order []string

	for _, column := range sortByValues {
		var descending bool

		if strings.HasPrefix(column, "-") {
			descending = true
			column = strings.TrimPrefix(column, "-")
		}

		if column == "" || !sortable.IsSortable(column) {
			return order, false
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	return order, true
}

func getDataQualityLatestQueryParameter(queryParams url.Values) (bool, error) {
	if values, wantsLatest := queryParams["latest"]; !wantsLatest {
		return false, nil
	} else if len(values) == 0 || values[0] == "" {
		return true, nil
	} else if latest, err := strconv.ParseBool(values[0]); err != nil {
		return false, fmt.Errorf(api.ErrorResponseDetailsLatestMalformed)
	} else {
		return latest, nil
	}
}

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

func (s *Resources) GetDataQualitySourceObjectCounts(response http.ResponseWriter, request *http.Request) {
	var (
		objectCounts             model.DataQualitySourceObjectCounts
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
		filters                  = model.DataQualitySourceObjectCountFilters{
			SourceKind: queryParams.Get(dataQualityQueryParameterSourceKind),
			NodeKind:   queryParams.Get(dataQualityQueryParameterNodeKind),
			RunID:      queryParams.Get(dataQualityQueryParameterRunID),
		}
	)

	if order, validOrder := buildDataQualityOrder(queryParams[api.QueryParameterSortBy], objectCounts); !validOrder {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else if latest, err := getDataQualityLatestQueryParameter(queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if latest && filters.RunID != "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, dataQualityErrorLatestWithRunID, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else {
		filters.Latest = latest

		if counts, count, err := s.DB.GetDataQualitySourceObjectCounts(request.Context(), start, end, filters, strings.Join(order, ", "), limit, skip); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), counts, start, end, limit, skip, count, http.StatusOK, response)
		}
	}
}

func (s *Resources) GetDataQualitySourceObjectCountSummaries(response http.ResponseWriter, request *http.Request) {
	var (
		objectCountSummaries     model.DataQualitySourceObjectCountSummaries
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
		filters                  = model.DataQualitySourceObjectCountFilters{
			SourceKind: queryParams.Get(dataQualityQueryParameterSourceKind),
			NodeKind:   queryParams.Get(dataQualityQueryParameterNodeKind),
			RunID:      queryParams.Get(dataQualityQueryParameterRunID),
		}
	)

	if order, validOrder := buildDataQualityOrder(queryParams[api.QueryParameterSortBy], objectCountSummaries); !validOrder {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else if latest, err := getDataQualityLatestQueryParameter(queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if latest && filters.RunID != "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, dataQualityErrorLatestWithRunID, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else {
		filters.Latest = latest

		if summaries, count, err := s.DB.GetDataQualitySourceObjectCountSummaries(request.Context(), start, end, filters, strings.Join(order, ", "), limit, skip); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), summaries, start, end, limit, skip, count, http.StatusOK, response)
		}
	}
}

func (s *Resources) GetDataQualityEnvironmentObjectCounts(response http.ResponseWriter, request *http.Request) {
	var (
		objectCounts             model.DataQualityEnvironmentObjectCounts
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
		filters                  = model.DataQualityEnvironmentObjectCountFilters{
			SourceKind:      queryParams.Get(dataQualityQueryParameterSourceKind),
			EnvironmentKind: queryParams.Get(dataQualityQueryParameterEnvironmentKind),
			EnvironmentID:   queryParams.Get(api.QueryParameterEnvironmentId),
			NodeKind:        queryParams.Get(dataQualityQueryParameterNodeKind),
			RunID:           queryParams.Get(dataQualityQueryParameterRunID),
		}
	)

	if order, validOrder := buildDataQualityOrder(queryParams[api.QueryParameterSortBy], objectCounts); !validOrder {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else if latest, err := getDataQualityLatestQueryParameter(queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if latest && filters.RunID != "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, dataQualityErrorLatestWithRunID, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else {
		filters.Latest = latest

		if counts, count, err := s.DB.GetDataQualityEnvironmentObjectCounts(request.Context(), start, end, filters, strings.Join(order, ", "), limit, skip); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), counts, start, end, limit, skip, count, http.StatusOK, response)
		}
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
