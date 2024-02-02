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
	"slices"
	"strings"
	"time"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
)

// AuditLogsResponse holds the data returned to an Audit logs request
type AuditLogsResponse struct {
	Logs model.AuditLogs `json:"logs"`
}

// ListAuditLogs retrieves audit logs
func (s Resources) ListAuditLogs(response http.ResponseWriter, request *http.Request) {
	var (
		order         []string
		auditLogs     model.AuditLogs
		sortByColumns = request.URL.Query()[api.QueryParameterSortBy]
	)

	const (
		logsBeforeQueryParam = "before"
		logsAfterQueryParam  = "after"
		offsetQueryParam     = "offset"
		limitQueryParam      = "limit"
	)

	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !auditLogs.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}

	}

	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(auditLogs.GetFilterableColumns(), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := auditLogs.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilters[name][i].IsStringData = auditLogs.IsString(filter.Name)
				}
			}
		}

		queryParams := request.URL.Query()

		// ignoring the error here as this would've failed at ParseQueryParameterFilters before getting here
		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if offset, err := ParseIntQueryParameter(queryParams, offsetQueryParam, 0); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, offsetQueryParam, err), response)
		} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, limitQueryParam, err), response)
		} else if getLogsBefore, err := ParseTimeQueryParameter(queryParams, logsBeforeQueryParam, time.Now()); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, logsBeforeQueryParam, err), response)
		} else if getLogsAfter, err := ParseTimeQueryParameter(queryParams, logsAfterQueryParam, getLogsBefore.Add(-time.Hour*24*365)); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, logsAfterQueryParam, err), response)
		} else if logs, count, err := s.DB.ListAuditLogs(getLogsBefore, getLogsAfter, offset, limit, strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithPagination(request.Context(), AuditLogsResponse{Logs: logs}, limit, offset, count, http.StatusOK, response)
		}
	}
}
