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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
)

const (
	ErrorResponseCodeBadRequest          = "BadRequest"
	ErrorResponseCodeInternalServerError = "InternalServerError"
	ErrorResponseConflict                = "Conflict"
	ErrorResponseCodeNotAuthorized       = "NotAuthorized"
	ErrorResponseForbidden               = "Forbidden"

	ErrorInvalidClientVersion                                        = "invalid version detected in user agent: %v"
	ErrorResponseClientCompleteJobInvalidStatus                      = "invalid status"
	ErrorResponseDataCollectionFlagNotProvided                       = "at least one data collection flag must be provided"
	ErrorResponseDetailsAuthenticationInvalid                        = "authentication is invalid"
	ErrorResponseDetailsBadQueryParameterFilters                     = "there are errors in the query parameter filters specified"
	ErrorResponseDetailsColumnNotFilterable                          = "the specified column cannot be filtered"
	ErrorResponseDetailsFilterPredicateNotSupported                  = "the specified filter predicate is not supported for this column"
	ErrorResponseDetailsForbidden                                    = "Forbidden"
	ErrorResponseDetailsFromMalformed                                = "from parameter should be formatted as RFC3339 i.e 2021-04-21T07:20:50.52Z"
	ErrorResponseDetailsIDMalformed                                  = "id is malformed"
	ErrorResponseDetailsInternalServerError                          = "an internal error has occurred that is preventing the service from servicing this request"
	ErrorResponseDetailsInvalidCombination                           = "the combination of inputs is not allowed"
	ErrorResponseDetailsLatestMalformed                              = "latest parameter has unexpected value"
	ErrorResponseDetailsNotSortable                                  = "column format does not support sorting"
	ErrorResponseEmptySortParameter                                  = "empty sort_by parameter supplied"
	ErrorResponseDetailsOTPInvalid                                   = "one time password is invalid"
	ErrorResponseDetailsResourceNotFound                             = "resource not found"
	ErrorResponseDetailsToBeforeFrom                                 = "to time cannot be before from time"
	ErrorResponseDetailsTimeRangeInvalid                             = "time range provided is invalid"
	ErrorResponseDetailsToMalformed                                  = "to parameter should be formatted as RFC3339 i.e 2021-04-21T07:20:50.52Z"
	ErrorResponseMultipleCollectionScopesProvided                    = "may only scope collection by exactly one of OU, Domain, or All Trusted Domains"
	ErrorResponsePayloadUnmarshalError                               = "error unmarshalling JSON payload"
	ErrorResponseRequestTimeout                                      = "request timed out"
	ErrorResponseUserSelfDisable                                     = "user attempted to disable themselves"
	ErrorResponseUserSelfRoleChange                                  = "user attempted to change own role"
	ErrorResponseUserSelfSSOProviderChange                           = "user attempted to change own SSO Provider"
	ErrorResponseUserSSOProviderRoleProvisionChange                  = "user attempted to change a role for a SSO Provider with role provision enabled"
	ErrorResponseAGTagWhiteSpace                                     = "asset group tags must not contain whitespace"
	ErrorResponseAGNameTagEmpty                                      = "asset group name or tag must not be empty"
	ErrorResponseAGDuplicateName                                     = "asset group name must be unique"
	ErrorResponseAGDuplicateTag                                      = "asset group tag must be unique"
	ErrorResponseSSOProviderDuplicateName                            = "sso provider name must be unique"
	ErrorResponseUserDuplicatePrincipal                              = "principal name must be unique"
	ErrorResponseUserDuplicateEmail                                  = "email must be unique"
	ErrorResponseDetailsUniqueViolation                              = "unique constraint was violated"
	ErrorResponseDetailsNotImplemented                               = "All good things to those who wait. Not implemented."
	ErrorResponseUnknownUser                                         = "unknown user"
	ErrorResponseAssetGroupTagExceededNameLimit                      = "asset group tag name is limited to 250 characters"
	ErrorResponseAssetGroupTagDuplicateKindName                      = "asset group tag name must be unique"
	ErrorResponseAssetGroupTagSelectorDuplicateName                  = "asset group tag selector name must be unique"
	ErrorResponseAssetGroupTagInvalid                                = "valid tag_type is required"
	ErrorResponseAssetGroupTagExceededTagLimit                       = "tag limit has been exceeded"
	ErrorResponseAssetGroupTagInvalidFields                          = "position and require_certify are only allowed for tiers"
	ErrorResponseAssetGroupTagPositionOutOfRange                     = "provided tier position is out of range"
	ErrorResponseDetailsQueryTooShort                                = "search query must be at least 3 characters long"
	ErrorResponseAssetGroupCertTypeInvalid                           = "valid certification action is required"
	ErrorResponseInvalidTagGlyph                                     = "the glyph specified is invalid"
	ErrorResponseAssetGroupTagDuplicateGlyph                         = "asset group tag glyph must be unique"
	ErrorResponseAssetGroupMemberIDsRequired                         = "asset group member IDs are required"
	ErrorResponseAssetGroupAutoCertifyInvalid                        = "auto_certify must be an input value of 0 to 2"
	ErrorResponseAssetGroupAutoCertifyOnlyAvailableForPrivilegeZones = "auto_certify is only available for asset group tags of tag_type = 1 (zones)"
	ErrorResponseAGTCannotUpdateAutoCertifiedNodes                   = "cannot change certification status for auto-certified members"
	ErrorResponseETACBadRequest                                      = "cannot specify environments when all_environments is true"
	ErrorResponseETACInvalidRoles                                    = "administrators and power users may not have an ETAC list applied to them"
	ErrorResponseAssetGroupTagInvalidTagName                         = "asset group tag name must contain only alphanumeric characters, spaces, and underscores"

	FmtErrorResponseDetailsBadQueryParameters            = "there are errors in the query parameters: %v"
	FmtErrorResponseDetailsMissingRequiredQueryParameter = "missing required query parameter: %v"
)

const (
	ErrorParseParams    = "unable to parse request parameters"
	ErrorDecodeParams   = "unable to decode request parameters"
	ErrorNoDomainId     = "no domain id specified in url"
	ErrorInvalidRFC3339 = "invalid RFC-3339 datetime format: %v"
)

func IsErrorResponse(response *http.Response) bool {
	return response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices
}

// ErrorWrapper is the standard API response structure
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
	} else if errors.Is(err, context.DeadlineExceeded) {
		WriteErrorResponse(request.Context(), BuildErrorResponse(http.StatusInternalServerError, ErrorResponseRequestTimeout, request), response)
	} else {
		slog.Error(fmt.Sprintf("Unexpected database error: %v", err))
		WriteErrorResponse(request.Context(), BuildErrorResponse(http.StatusInternalServerError, ErrorResponseDetailsInternalServerError, request), response)
	}
}

// FormatDatabaseError logs and returns an error (not found or other) depending on the database error encountered
// Alternate: HandleDatabaseError()
func FormatDatabaseError(err error) error {
	if errors.Is(err, database.ErrNotFound) {
		return errors.New(ErrorResponseDetailsResourceNotFound)
	} else {
		slog.Error(fmt.Sprintf("Unexpected database error: %v", err))
		return errors.New(ErrorResponseDetailsInternalServerError)
	}
}
