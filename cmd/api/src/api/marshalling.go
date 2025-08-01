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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"

	"github.com/specterops/bloodhound/cmd/api/src/api/stream"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
)

const (
	// DefaultAPIPayloadReadLimitBytes sets the maximum API body size to 10MB
	DefaultAPIPayloadReadLimitBytes = 10 * 1024 * 1024
)

var (
	ErrContentTypeJson = errors.New("content type must be application/json")
	ErrNoRequestBody   = errors.New("request body is empty")
)

// These are the standardized API V2 response structures

// ToJSONRawMessage takes any value and produces a typed json.RawMessage result from it along with any encoding errors
// encountered
func ToJSONRawMessage(value any) (json.RawMessage, error) {
	return json.Marshal(value)
}

// BasicResponse is used for endpoints that return non-paginated, non-time-windowed data
type BasicResponse struct {
	Data json.RawMessage `json:"data"`
}

// TimeWindowedResponse is used for endpoints that return non-paginated data with start/end times (time window)
type TimeWindowedResponse struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
	Data  any        `json:"data"`
}

// ResponseWrapper is used for endpoints that return paginated data with or without start/end times (time window)
type ResponseWrapper struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
	Count int        `json:"count"`
	Limit int        `json:"limit"`
	Skip  int        `json:"skip"`
	Data  any        `json:"data"`
}

func WriteErrorResponse(ctx context.Context, untypedError any, response http.ResponseWriter) {
	if typedError, ok := untypedError.(*ErrorWrapper); !ok {
		slog.WarnContext(ctx, fmt.Sprintf("Failure Writing API Error. Status: %v. Message: %v", http.StatusInternalServerError, "Invalid error format returned"))
		WriteJSONResponse(ctx, "An internal error has occurred that is preventing the service from servicing this request.", http.StatusInternalServerError, response)
	} else {
		slog.WarnContext(ctx, fmt.Sprintf("Writing API Error. Status: %v. Message: %v", typedError.HTTPStatus, typedError.Errors))
		WriteJSONResponse(ctx, typedError, typedError.HTTPStatus, response)
	}
}

// context should not be handled as part of the ResponseWriter functions, these should preferably be moved
// somewhere in the middleware (but that would be a larger refactor, hence leaving it out of BED-2257)

func WriteBasicResponse(ctx context.Context, inputData any, statusCode int, response http.ResponseWriter) {
	if data, err := ToJSONRawMessage(inputData); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed marshaling data for basic response: %v", err))
		response.WriteHeader(http.StatusInternalServerError)
	} else {
		WriteJSONResponse(ctx, BasicResponse{
			Data: data,
		}, statusCode, response)
	}
}

func WriteResponseWrapperWithPagination(ctx context.Context, data any, limit int, skip, count, statusCode int, response http.ResponseWriter) {
	wrapper := ResponseWrapper{}
	wrapper.Data = data
	wrapper.Skip = skip

	if limit > 0 {
		wrapper.Limit = limit
	}

	if count >= 0 {
		wrapper.Count = count
	}

	WriteJSONResponse(ctx, wrapper, statusCode, response)
}

func WriteTimeWindowedResponse(ctx context.Context, data any, start, end time.Time, statusCode int, response http.ResponseWriter) {
	timeWindowedResponse := TimeWindowedResponse{}
	timeWindowedResponse.Data = data

	if !start.IsZero() {
		timeWindowedResponse.Start = &start
	}

	if !end.IsZero() {
		timeWindowedResponse.End = &end
	}

	WriteJSONResponse(ctx, timeWindowedResponse, statusCode, response)
}

func WriteResponseWrapperWithTimeWindowAndPagination(ctx context.Context, data any, start, end time.Time, limit int, skip, count, statusCode int, response http.ResponseWriter) {
	wrapper := ResponseWrapper{}
	wrapper.Data = data
	wrapper.Skip = skip

	if limit > 0 {
		wrapper.Limit = limit
	}

	if count >= 0 {
		wrapper.Count = count
	}

	if !start.IsZero() {
		wrapper.Start = &start
	}

	if !end.IsZero() {
		wrapper.End = &end
	}

	WriteJSONResponse(ctx, wrapper, statusCode, response)
}

func WriteJSONResponse(ctx context.Context, message any, statusCode int, response http.ResponseWriter) {
	response.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	if content, err := json.Marshal(message); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed to marshal value into JSON for request: %v: for message: %+v", err, message))
		response.WriteHeader(http.StatusInternalServerError)
	} else {
		response.WriteHeader(statusCode)
		if written, err := response.Write(content); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Writing API Error. Failed to write JSON response with %d bytes written and error: %v", written, err))
		}
	}
}

func WriteCSVResponse(ctx context.Context, message model.CSVWriter, statusCode int, response http.ResponseWriter) {
	response.Header().Set(headers.ContentType.String(), mediatypes.TextCsv.String())

	if err := message.WriteCSV(response); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, fmt.Sprintf("Writing API Error. Failed to write CSV for request: %v", err))
		return
	}

	response.WriteHeader(statusCode)
}

func WriteBinaryResponse(ctx context.Context, data []byte, filename string, statusCode int, response http.ResponseWriter) {
	response.Header().Set(headers.ContentType.String(), mediatypes.ApplicationOctetStream.String())
	response.Header().Set(headers.ContentDisposition.String(), fmt.Sprintf(utils.ContentDispositionAttachmentTemplate, filename))
	response.WriteHeader(statusCode)

	if written, err := response.Write(data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Writing API Error. Failed to write binary response with %d bytes written and error: %v", written, err))
	}
}

func ReadJsonResponsePayload(value any, response *http.Response) error {
	if !utils.HeaderMatches(response.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()) {
		return ErrContentTypeJson
	}

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("could not decode json response payload into value: %w", err)
	} else {
		return nil
	}
}

func ReadAPIV2ResponsePayload(value any, response *http.Response) error {
	if !utils.HeaderMatches(response.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()) {
		return ErrContentTypeJson
	}

	var wrapper BasicResponse

	if content, err := io.ReadAll(response.Body); err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	} else if err := json.Unmarshal(content, &wrapper); err != nil {
		return fmt.Errorf("failed to unmarshal body into basic response: %w: body was %s", err, string(content))
	} else if err := json.Unmarshal(wrapper.Data, &value); err != nil {
		return fmt.Errorf("failed to unmarshal basic response into value: %w: body was %s", err, string(content))
	} else {
		return nil
	}
}

func ReadAPIV2ResponseWrapperPayload(value any, response *http.Response) error {
	if !utils.HeaderMatches(response.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()) {
		return ErrContentTypeJson
	}

	if content, err := io.ReadAll(response.Body); err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	} else if err := json.Unmarshal(content, &value); err != nil {
		return fmt.Errorf("failed to unmarshal body into value: %w: body was %s", err, string(content))
	} else {
		return nil
	}
}

func ReadAPIV2ErrorResponsePayload(value *ErrorWrapper, response *http.Response) error {
	if !utils.HeaderMatches(response.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()) {
		return ErrContentTypeJson
	}

	if content, err := io.ReadAll(response.Body); err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	} else if err := json.Unmarshal(content, &value); err != nil {
		return fmt.Errorf("failed to unmarshal body into value: %w: body was %s", err, string(content))
	} else {
		return nil
	}
}

func ReadJSONRequestPayloadLimited(value any, request *http.Request) error {
	if !utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()) {
		return ErrContentTypeJson
	}

	if request.Body == nil {
		return ErrNoRequestBody
	}

	var (
		limitedReader = stream.NewLimitedReader(DefaultAPIPayloadReadLimitBytes, request.Body)
		decoder       = json.NewDecoder(limitedReader)
	)

	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("could not decode limited payload request into value: %w", err)
	} else {
		return nil
	}
}
