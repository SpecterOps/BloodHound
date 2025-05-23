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
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	ingestModel "github.com/specterops/bloodhound/src/model/ingest"

	"github.com/specterops/bloodhound/src/services/ingest"
	"github.com/specterops/bloodhound/src/services/job"
)

const FileUploadJobIdPathParameterName = "file_upload_job_id"

func (s Resources) ListIngestJobs(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams    = request.URL.Query()
		sortByColumns  = queryParams[api.QueryParameterSortBy]
		order          []string
		fileUploadJobs model.IngestJobs
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
		} else if ingestJobs, count, err := job.GetAllIngestJobs(request.Context(), s.DB, skip, limit, strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithPagination(request.Context(), ingestJobs, limit, skip, count, http.StatusOK, response)
		}
	}
}

func (s Resources) StartIngestJob(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Starting new ingest job")()
	reqCtx := ctx.Get(request.Context())

	if user, valid := auth.GetUserFromAuthCtx(reqCtx.AuthCtx); !valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid, request), response)
	} else if ingestJob, err := job.StartIngestJob(request.Context(), s.DB, user); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), ingestJob, http.StatusCreated, response)
	}
}

func (s Resources) ProcessIngestTask(response http.ResponseWriter, request *http.Request) {
	var (
		requestId   = ctx.FromRequest(request).RequestID
		jobIdString = mux.Vars(request)[FileUploadJobIdPathParameterName]
		validator   = ingest.NewIngestValidator(s.IngestSchema)
	)

	if request.Body != nil {
		defer request.Body.Close()
	}

	if !IsValidContentTypeForUpload(request.Header) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Content type must be application/json or application/zip", request), response)
	} else if jobID, err := strconv.Atoi(jobIdString); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if ingestJob, err := job.GetIngestJobByID(request.Context(), s.DB, int64(jobID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if ingestTaskParams, err := ingest.SaveIngestFile(s.Config.TempDirectory(), request, validator); errors.Is(err, ingest.ErrInvalidJSON) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error saving ingest file: %v", err), request), response)
	} else if report, ok := err.(ingest.ValidationReport); ok {
		var (
			msgs       = report.BuildAPIError()
			errDetails = []api.ErrorDetails{}
		)

		for _, msg := range msgs {
			errDetails = append(errDetails, api.ErrorDetails{Message: msg})
		}

		e := &api.ErrorWrapper{
			HTTPStatus: http.StatusBadRequest,
			Timestamp:  time.Now(),
			RequestID:  ctx.FromRequest(request).RequestID,
			Errors:     errDetails,
		}

		api.WriteErrorResponse(request.Context(), e, response)
	} else if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error saving ingest file: %v", err), request), response)
	} else if _, err = ingest.CreateIngestTask(request.Context(), s.DB, ingest.IngestTaskParams{Filename: ingestTaskParams.Filename, FileType: ingestTaskParams.FileType, RequestID: requestId, JobID: int64(jobID)}); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err = job.TouchIngestJobLastIngest(request.Context(), s.DB, ingestJob); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func (s Resources) EndIngestJob(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Finished ingest job")()

	jobIdString := mux.Vars(request)[FileUploadJobIdPathParameterName]

	if jobID, err := strconv.Atoi(jobIdString); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if ingestJob, err := job.GetIngestJobByID(request.Context(), s.DB, int64(jobID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if ingestJob.Status != model.JobStatusRunning {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "job must be in running status to end", request), response)
	} else if err := job.EndIngestJob(request.Context(), s.DB, ingestJob); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func (s Resources) ListAcceptedFileUploadTypes(response http.ResponseWriter, request *http.Request) {
	api.WriteBasicResponse(request.Context(), ingestModel.AllowedFileUploadTypes, http.StatusOK, response)
}

func IsValidContentTypeForUpload(header http.Header) bool {
	rawValue := header.Get(headers.ContentType.String())
	if rawValue == "" {
		return false
	} else if parsed, _, err := mime.ParseMediaType(rawValue); err != nil {
		return false
	} else {
		return slices.Contains(ingestModel.AllowedFileUploadTypes, parsed)
	}
}
