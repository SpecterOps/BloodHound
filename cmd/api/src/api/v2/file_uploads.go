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
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/fileupload"
	"github.com/specterops/bloodhound/src/services/ingest"
)

const FileUploadJobIdPathParameterName = "file_upload_job_id"

func (s Resources) ListFileUploadJobs(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams    = request.URL.Query()
		sortByColumns  = queryParams[api.QueryParameterSortBy]
		order          []string
		fileUploadJobs model.FileUploadJobs
	)

	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !fileUploadJobs.IsSortable(column) {
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
			if valid := slices.Contains(fileUploadJobs.GetFilterableColumns(), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := fileUploadJobs.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilters[name][i].IsStringData = fileUploadJobs.IsString(filter.Name)
				}
			}
		}

		// ignoring the error here as this would've failed at ParseQueryParameterFilters before getting here
		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterSkip, err), response)
		} else if limit, err := ParseLimitQueryParameter(queryParams, 100); err != nil {
			api.WriteErrorResponse(request.Context(), ErrBadQueryParameter(request, model.PaginationQueryParameterLimit, err), response)
		} else if fileUploadJobs, count, err := fileupload.GetAllFileUploadJobs(s.DB, skip, limit, strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithPagination(request.Context(), fileUploadJobs, limit, skip, count, http.StatusOK, response)
		}
	}
}

func (s Resources) StartFileUploadJob(response http.ResponseWriter, request *http.Request) {
	defer log.Measure(log.LevelDebug, "Starting new file upload job")()
	var reqCtx = ctx.Get(request.Context())

	if user, valid := auth.GetUserFromAuthCtx(reqCtx.AuthCtx); !valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid, request), response)
	} else if fileUploadJob, err := fileupload.StartFileUploadJob(s.DB, user); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), fileUploadJob, http.StatusCreated, response)
	}
}

func (s Resources) ProcessFileUpload(response http.ResponseWriter, request *http.Request) {
	var (
		requestId             = ctx.FromRequest(request).RequestID
		fileUploadJobIdString = mux.Vars(request)[FileUploadJobIdPathParameterName]
	)

	if fileUploadJobID, err := strconv.Atoi(fileUploadJobIdString); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if fileUploadJob, err := fileupload.GetFileUploadJobByID(s.DB, int64(fileUploadJobID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if fileName, err := fileupload.SaveIngestFile(s.Config.TempDirectory(), request.Body); errors.Is(err, fileupload.ErrInvalidJSON) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error saving ingest file: %v", err), request), response)
	} else if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error saving ingest file: %v", err), request), response)
	} else if _, err = ingest.CreateIngestTask(s.DB, fileName, requestId, int64(fileUploadJobID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err = fileupload.TouchFileUploadJobLastIngest(s.DB, fileUploadJob); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func (s Resources) EndFileUploadJob(response http.ResponseWriter, request *http.Request) {
	defer log.Measure(log.LevelDebug, "Finished file upload job")()

	var fileUploadJobIdString = mux.Vars(request)[FileUploadJobIdPathParameterName]

	if fileUploadJobID, err := strconv.Atoi(fileUploadJobIdString); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if fileUploadJob, err := fileupload.GetFileUploadJobByID(s.DB, int64(fileUploadJobID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if fileUploadJob.Status != model.JobStatusRunning {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "job must be in running status to end", request), response)
	} else if err := fileupload.EndFileUploadJob(s.DB, fileUploadJob); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}
