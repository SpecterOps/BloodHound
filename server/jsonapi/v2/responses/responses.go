// Copyright 2026 Specter Ops, Inc.
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

package responses

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

const errorMessageInternalServerError = "an internal error has occurred that is preventing the service from servicing this request"

// BasicResponse is the envelope returned for non-paginated, non-time-windowed responses.
type BasicResponse struct {
	Data json.RawMessage `json:"data"`
}

// ErrorDetails describes a single error returned inside an ErrorWrapper.
type ErrorDetails struct {
	Context string `json:"context"`
	Message string `json:"message"`
}

// ErrorWrapper is the standard error envelope returned by API endpoints.
type ErrorWrapper struct {
	HTTPStatus int            `json:"http_status"`
	Timestamp  time.Time      `json:"timestamp"`
	RequestID  string         `json:"request_id"`
	Errors     []ErrorDetails `json:"errors"`
}

// WriteBasic marshals inputData as the data field of a BasicResponse and writes it with the
// supplied status code.
func WriteBasic(requestCtx context.Context, inputData any, statusCode int, response http.ResponseWriter) {
	var (
		rawData json.RawMessage
		err     error
	)

	rawData, err = json.Marshal(inputData)
	if err != nil {
		slog.ErrorContext(requestCtx, "Failed marshaling data for basic response", attr.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(requestCtx, BasicResponse{Data: rawData}, statusCode, response)
}

// WriteError writes a structured ErrorWrapper to the response with the supplied status code
// and message.
func WriteError(request *http.Request, statusCode int, message string, response http.ResponseWriter) {
	var errorWrapper = ErrorWrapper{
		HTTPStatus: statusCode,
		Timestamp:  time.Now(),
		RequestID:  bhctx.FromRequest(request).RequestID,
		Errors: []ErrorDetails{
			{Message: message},
		},
	}

	slog.WarnContext(
		request.Context(),
		"Writing API error",
		slog.Int("http_status", statusCode),
		slog.Any("errors", errorWrapper.Errors),
	)
	writeJSON(request.Context(), errorWrapper, statusCode, response)
}

// WriteInternalServerError writes a generic 500 error response and logs the underlying cause.
// Use this when a service returns an error that the handler cannot map to a more specific
// failure mode.
func WriteInternalServerError(request *http.Request, cause error, response http.ResponseWriter) {
	slog.ErrorContext(request.Context(), "Unexpected service error", attr.Error(cause))
	WriteError(request, http.StatusInternalServerError, errorMessageInternalServerError, response)
}

func writeJSON(requestCtx context.Context, message any, statusCode int, response http.ResponseWriter) {
	var (
		content []byte
		err     error
	)

	content, err = json.Marshal(message)
	if err != nil {
		slog.ErrorContext(requestCtx, "Failed to marshal value into JSON", attr.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response.WriteHeader(statusCode)
	if _, writeErr := response.Write(content); writeErr != nil {
		slog.ErrorContext(requestCtx, "Failed to write JSON response body", attr.Error(writeErr))
	}
}
