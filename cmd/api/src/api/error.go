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

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
)

const (
	ErrorResponseCodeBadRequest          = "BadRequest"
	ErrorResponseCodeInternalServerError = "InternalServerError"
	ErrorResponseConflict                = "Conflict"
	ErrorResponseCodeNotAuthorized       = "NotAuthorized"
	ErrorResponseForbidden               = "Forbidden"

	ErrorInvalidClientVersion                       = "invalid version detected in user agent: %v"
	ErrorResponseClientCompleteJobInvalidStatus     = "invalid status"
	ErrorResponseDataCollectionFlagNotProvided      = "at least one data collection flag must be provided"
	ErrorResponseDetailsAuthenticationInvalid       = "authentication is invalid"
	ErrorResponseDetailsBadQueryParameterFilters    = "there are errors in the query parameter filters specified"
	ErrorResponseDetailsColumnNotFilterable         = "the specified column cannot be filtered"
	ErrorResponseDetailsFilterPredicateNotSupported = "the specified filter predicate is not supported for this column"
	ErrorResponseDetailsForbidden                   = "Forbidden"
	ErrorResponseDetailsFromMalformed               = "from parameter should be formatted as RFC3339 i.e 2021-04-21T07:20:50.52Z"
	ErrorResponseDetailsIDMalformed                 = "id is malformed."
	ErrorResponseDetailsInternalServerError         = "an internal error has occurred that is preventing the service from servicing this request"
	ErrorResponseDetailsInvalidCombination          = "the combination of inputs is not allowed"
	ErrorResponseDetailsLatestMalformed             = "latest parameter has unexpected value"
	ErrorResponseDetailsNotSortable                 = "column format does not support sorting"
	ErrorResponseEmptySortParameter                 = "empty sort_by parameter supplied"
	ErrorResponseDetailsOTPInvalid                  = "one time password is invalid"
	ErrorResponseDetailsResourceNotFound            = "resource not found"
	ErrorResponseDetailsToBeforeFrom                = "to time cannot be before from time"
	ErrorResponseDetailsToMalformed                 = "to parameter should be formatted as RFC3339 i.e 2021-04-21T07:20:50.52Z"
	ErrorResponseMultipleCollectionScopesProvided   = "may only scope collection by exactly one of OU, Domain, or All Trusted Domains"
	ErrorResponsePayloadUnmarshalError              = "error unmarshalling JSON payload"
	ErrorResponseUserSelfDisable                    = "user attempted to disable themselves"

	FmtErrorResponseDetailsBadQueryParameters = "there are errors in the query parameters: %v"
)

func IsErrorResponse(response *http.Response) bool {
	return response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices
}

// ErrorResponse is the V1 response
type ErrorResponse struct {
	HTTPStatus int
	Error      any
}

// ErrorWrapper is the V2 response
type ErrorWrapper struct {
	HTTPStatus int            `json:"http_status"`
	Timestamp  time.Time      `json:"timestamp"`
	RequestID  string         `json:"request_id"`
	Errors     []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Context string `json:"context"`
	Message string `json:"message"`
}

// Error implements the built-in Error() function for the errors package
func (s ErrorWrapper) Error() string {
	errorMessages := make([]string, 0)
	for _, errorDetails := range s.Errors {
		errorMessages = append(errorMessages, errorDetails.Message)
	}

	return fmt.Sprintf("Code: %d - errors: %s", s.HTTPStatus, strings.Join(errorMessages, "; "))
}

// BuildErrorResponse returns an ErrorWrapper struct built with the provided data
func BuildErrorResponse(httpStatus int, message string, request *http.Request) *ErrorWrapper {
	return &ErrorWrapper{
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
		RequestID:  ctx.FromRequest(request).RequestID,
		Errors: []ErrorDetails{
			{
				Message: message,
			},
		},
	}
}

// HandleDatabaseError writes an error (not found or other) depending on the database error encountered
// Alternate: FormatDatabaseError()
func HandleDatabaseError(request *http.Request, response http.ResponseWriter, err error) {
	if errors.Is(err, database.ErrNotFound) {
		WriteErrorResponse(request.Context(), BuildErrorResponse(http.StatusNotFound, ErrorResponseDetailsResourceNotFound, request), response)
	} else {
		log.Errorf("Unexpected database error: %v", err)
		WriteErrorResponse(request.Context(), BuildErrorResponse(http.StatusInternalServerError, ErrorResponseDetailsInternalServerError, request), response)
	}
}

// FormatDatabaseError logs and returns an error (not found or other) depending on the database error encountered
// Alternate: HandleDatabaseError()
func FormatDatabaseError(err error) error {
	if errors.Is(err, database.ErrNotFound) {
		return errors.New(ErrorResponseDetailsResourceNotFound)
	} else {
		log.Errorf("Unexpected database error: %v", err)
		return errors.New(ErrorResponseDetailsInternalServerError)
	}
}
