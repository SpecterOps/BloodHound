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
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils"
)

const (
	ErrorNoTenantId        string = "no tenant id specified in url"
	ErrorNoPlatformId      string = "no platform id specified in url"
	ErrorInvalidPlatformId string = "invalid platform id specified in url: %v"
)

func (s Resources) GetDatabaseCompleteness(response http.ResponseWriter, request *http.Request) {
	defer log.Measure(log.LevelDebug, "Get Current Database Completeness")()

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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error getting quality stat: %v", err), request), response)
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorNoDomainId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetADDataQualityStats(id, start, end, strings.Join(order, ", "), limit, skip); err != nil {
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorNoTenantId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetAzureDataQualityStats(id, start, end, strings.Join(order, ", "), limit, skip); err != nil {
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorNoPlatformId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRFC3339, queryParams["end"]), request), response)
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
			stats, count, err = s.DB.GetADDataQualityAggregations(start, end, strings.Join(order, ", "), limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		case "azure":
			stats, count, err = s.DB.GetAzureDataQualityAggregations(start, end, strings.Join(order, ", "), limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidPlatformId, id), request), response)
			return
		}

		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}
