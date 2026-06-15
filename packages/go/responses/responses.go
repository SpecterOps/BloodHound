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
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

const errorMessageInternalServerError = "an internal error has occurred that is preventing the service from servicing this request"
const failedToMarshalMessage = "Failed to marshal response. Try again later."

// JSONViewer is implemented by view types that can marshal themselves to the
// JSON byte slice used as the `data` payload of a BasicResponse.
type JSONViewer interface {
	JSONView() ([]byte, error)
}

// CSVViewer is implemented by view types that can serialize themselves as CSV.
// It mirrors JSONViewer: the view type implements the serialization method and
// the matching WriteCSV helper sets headers and the response status.
type CSVViewer interface {
	WriteCSV(writer io.Writer) error
}

// BasicResponse is the envelope returned for non-paginated, non-time-windowed responses.
type BasicResponse struct {
	Data json.RawMessage `json:"data"`
}

// PaginatedResponse is the envelope returned for paginated list responses.
type PaginatedResponse struct {
	Count int             `json:"count"`
	Limit int             `json:"limit"`
	Skip  int             `json:"skip"`
	Data  json.RawMessage `json:"data"`
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

// WriteBasic marshals data as the data field of a BasicResponse and writes it with the
// supplied status code.
func WriteBasic(ctx context.Context, data JSONViewer, statusCode int, response http.ResponseWriter) {
	rawData, err := data.JSONView()
	if err != nil {
		slog.ErrorContext(ctx, "Failed marshaling data for basic response", attr.Error(err))
		WriteError(ctx, http.StatusInternalServerError, failedToMarshalMessage, response)
		return
	}

	writeJSON(ctx, BasicResponse{Data: rawData}, statusCode, response)
}

// WritePaginated marshals data as the data field of a PaginatedResponse along with the
// supplied pagination metadata and writes it with the supplied status code. A non-positive
// limit and a negative count are omitted (left as their zero value) to match the legacy
// pagination envelope semantics.
func WritePaginated(ctx context.Context, data JSONViewer, limit, skip, count, statusCode int, response http.ResponseWriter) {
	rawData, err := data.JSONView()
	if err != nil {
		slog.ErrorContext(ctx, "Failed marshaling data for paginated response", attr.Error(err))
		WriteError(ctx, http.StatusInternalServerError, failedToMarshalMessage, response)
		return
	}

	var wrapper = PaginatedResponse{
		Data: rawData,
		Skip: skip,
	}

	if limit > 0 {
		wrapper.Limit = limit
	}

	if count >= 0 {
		wrapper.Count = count
	}

	writeJSON(ctx, wrapper, statusCode, response)
}

// WriteCSV writes data as a text/csv response with the supplied status code. The caller is
// responsible for setting any Content-Disposition header before invoking this helper.
func WriteCSV(ctx context.Context, data CSVViewer, statusCode int, response http.ResponseWriter) {
	response.Header().Set(headers.ContentType.String(), mediatypes.TextCsv.String())
	response.WriteHeader(statusCode)

	if err := data.WriteCSV(response); err != nil {
		slog.ErrorContext(ctx, "Failed to write CSV response body", attr.Error(err))
		return
	}
}

// WriteError writes a structured ErrorWrapper to the response with the supplied status code
// and message.
func WriteError(ctx context.Context, statusCode int, message string, response http.ResponseWriter) {
	var errorWrapper = ErrorWrapper{
		HTTPStatus: statusCode,
		Timestamp:  time.Now(),
		RequestID:  bhctx.Get(ctx).RequestID,
		Errors: []ErrorDetails{
			{Message: message},
		},
	}

	slog.WarnContext(
		ctx,
		"Writing API error",
		slog.Int("http_status", statusCode),
		slog.Any("errors", errorWrapper.Errors),
	)
	writeJSON(ctx, errorWrapper, statusCode, response)
}

// WriteInternalServerError writes a generic 500 error response and logs the underlying cause.
// Use this when a service returns an error that the handler cannot map to a more specific
// failure mode.
func WriteInternalServerError(ctx context.Context, cause error, response http.ResponseWriter) {
	slog.ErrorContext(ctx, "Unexpected service error", attr.Error(cause))
	WriteError(ctx, http.StatusInternalServerError, errorMessageInternalServerError, response)
}

func writeJSON[T BasicResponse | ErrorWrapper | PaginatedResponse](requestCtx context.Context, message T, statusCode int, response http.ResponseWriter) {
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
