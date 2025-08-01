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
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	ingestModel "github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/headers"

	"github.com/specterops/bloodhound/cmd/api/src/services/job"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
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
		validator   = upload.NewIngestValidator(s.IngestSchema)
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
	} else if ingestJob.Status != model.JobStatusRunning {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "job must be in running status to attach files", request), response)
	} else if ingestTaskParams, err := upload.SaveIngestFile(s.Config.TempDirectory(), request, validator); errors.Is(err, upload.ErrInvalidJSON) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error saving ingest file: %v", err), request), response)
	} else if report, ok := err.(upload.ValidationReport); ok {
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
	} else if _, err = upload.CreateIngestTask(request.Context(), s.DB, upload.IngestTaskParams{Filename: ingestTaskParams.Filename, ProvidedFileName: "", FileType: ingestTaskParams.FileType, RequestID: requestId, JobID: int64(jobID)}); err != nil {
		if removeErr := os.Remove(ingestTaskParams.Filename); removeErr != nil {
			slog.WarnContext(request.Context(), fmt.Sprintf("Failed to clean up file after task creation error: %v", removeErr))
		}
		api.HandleDatabaseError(request, response, err)
	} else if err = job.TouchIngestJobLastIngest(request.Context(), s.DB, ingestJob); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

type ProcessMultipartIngestTaskResponse struct {
	TotalParts  int                              `json:"total_parts"`
	FailedParts int                              `json:"failed_parts"`
	PartsData   map[string]MultipartPartResponse `json:"parts_data"`
}

type MultipartPartResponse struct {
	PartName string
	FileName string
	Errors   []string
}

func (s Resources) ProcessMultipartIngestTask(response http.ResponseWriter, request *http.Request) {
	var (
		requestId   = ctx.FromRequest(request).RequestID
		jobIdString = mux.Vars(request)[FileUploadJobIdPathParameterName]
		validator   = upload.NewIngestValidator(s.IngestSchema)
	)

	if !strings.HasPrefix(request.Header.Get("Content-Type"), "multipart/form-data") {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Content-Type must be multipart/form-data", request), response)
	} else if reader, err := request.MultipartReader(); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "failed to open multipart form data", request), response)
		return
	} else if jobID, err := strconv.Atoi(jobIdString); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if ingestJob, err := job.GetIngestJobByID(request.Context(), s.DB, int64(jobID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if ingestJob.Status != model.JobStatusRunning {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "job must be in running status to attach files", request), response)
	} else if results, err := s.processMultipart(request.Context(), validator, reader); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("failed to process multipart data: %v", err), request), response)
	} else {
		var (
			total         = len(results)
			failed        = 0
			partsResponse = make(map[string]MultipartPartResponse)
		)

		for _, result := range results {
			fileResponse := MultipartPartResponse{
				PartName: result.PartName,
				FileName: result.ProvidedFileName,
				Errors:   result.Errors,
			}

			if len(result.Errors) > 0 {
				failed += 1
			} else if _, err = upload.CreateIngestTask(request.Context(), s.DB, upload.IngestTaskParams{Filename: result.GeneratedFileName, ProvidedFileName: result.ProvidedFileName, FileType: result.FileType, RequestID: requestId, JobID: int64(jobID)}); err != nil {
				if removeErr := os.Remove(result.GeneratedFileName); removeErr != nil {
					slog.WarnContext(request.Context(), fmt.Sprintf("Failed to clean up file after task creation error: %v", removeErr))
				}
				fileResponse.Errors = append(fileResponse.Errors, fmt.Sprintf("Error creating ingest task: %v", err))
			} else if err = job.TouchIngestJobLastIngest(request.Context(), s.DB, ingestJob); err != nil {
				fileResponse.Errors = append(fileResponse.Errors, fmt.Sprintf("Error updating ingest job: %v", err))
			}

			partsResponse[result.PartName] = fileResponse
		}

		api.WriteBasicResponse(request.Context(), ProcessMultipartIngestTaskResponse{TotalParts: total, FailedParts: failed, PartsData: partsResponse}, http.StatusOK, response)
	}
}

type multipartResult struct {
	PartName          string
	ProvidedFileName  string
	GeneratedFileName string
	FileType          model.FileType
	Errors            []string
}

func (s Resources) processMultipart(ctx context.Context, validator upload.IngestValidator, reader *multipart.Reader) ([]multipartResult, error) {
	var (
		results         = make([]multipartResult, 0)
		processingError error
	)

	for {
		if part, err := reader.NextPart(); err == io.EOF {
			break
		} else if err != nil {
			processingError = fmt.Errorf("failed to read multipart part: %w", err)
			break
		} else if part.FormName() == "" {
			processingError = fmt.Errorf("all form parts must specify a name")
			break
		} else if part.FileName() == "" {
			results = append(results, multipartResult{
				PartName: part.FormName(),
				Errors:   []string{fmt.Sprintf("Must be a file")},
			})
		} else if !IsValidContentTypeForUploadMultipart(part.Header) {
			results = append(results, multipartResult{
				PartName:         part.FormName(),
				ProvidedFileName: part.FileName(),
				Errors:           []string{fmt.Sprintf("Content type must be application/json or application/zip")},
			})
		} else if ingestTaskParams, err := upload.SaveMultipartIngestFile(s.Config.TempDirectory(), part, validator); errors.Is(err, upload.ErrInvalidJSON) {
			results = append(results, multipartResult{
				PartName:         part.FormName(),
				ProvidedFileName: part.FileName(),
				Errors:           []string{fmt.Sprintf("Error saving ingest file: %v", err)},
			})
		} else if report, ok := err.(upload.ValidationReport); ok {
			msgs := report.BuildAPIError()

			results = append(results, multipartResult{
				PartName:         part.FormName(),
				ProvidedFileName: part.FileName(),
				Errors:           msgs,
			})
		} else if err != nil {
			results = append(results, multipartResult{
				PartName:         part.FormName(),
				ProvidedFileName: part.FileName(),
				Errors:           []string{fmt.Sprintf("Error saving ingest file: %v", err)},
			})
		} else {
			results = append(results, multipartResult{
				PartName:          part.FormName(),
				ProvidedFileName:  part.FileName(),
				GeneratedFileName: ingestTaskParams.Filename,
				FileType:          ingestTaskParams.FileType,
				Errors:            []string{},
			})
		}
	}

	if processingError != nil {
		for _, result := range results {
			if result.GeneratedFileName != "" {
				if err := os.Remove(result.GeneratedFileName); err != nil {
					slog.WarnContext(ctx, fmt.Sprintf("Error removing ingest file %s: %v", result.GeneratedFileName, err))
				}
			}
		}

		return []multipartResult{}, processingError
	}

	return results, nil
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

func IsValidContentTypeForUploadMultipart(header textproto.MIMEHeader) bool {
	rawValue := header.Get(headers.ContentType.String())
	if rawValue == "" {
		return false
	} else if parsed, _, err := mime.ParseMediaType(rawValue); err != nil {
		return false
	} else {
		return slices.Contains(ingestModel.AllowedFileUploadTypes, parsed)
	}
}
