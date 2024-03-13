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

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/src/api"
)

func TestWriteErrorResponse_InvalidFormat(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteErrorResponse(context.Background(), fmt.Errorf("foo"), response)
	require.Equal(t, response.Code, http.StatusInternalServerError)
	require.Contains(t, response.Body.String(), "internal error")
}

func TestWriteErrorResponse_V1(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteErrorResponse(context.Background(), &api.ErrorResponse{
		HTTPStatus: http.StatusTeapot,
		Error:      json.RawMessage(`{"foo":"bar"}`),
	}, response)
	require.Equal(t, response.Code, http.StatusTeapot)
	require.Contains(t, response.Body.String(), "foo")
}

func TestWriteErrorResponse_V2(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteErrorResponse(context.Background(), &api.ErrorWrapper{
		HTTPStatus: http.StatusTeapot,
		Errors: []api.ErrorDetails{{
			Context: "bar",
			Message: "baz",
		}},
	}, response)
	require.Equal(t, response.Code, http.StatusTeapot)
	require.Contains(t, response.Body.String(), "baz")
}

func TestWriteBasicResponse(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteBasicResponse(context.Background(), json.RawMessage(`{"foo":"bar"}`), http.StatusOK, response)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "foo")
}

func TestWriteResponseWrapperWithPagination(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteResponseWrapperWithPagination(context.Background(), json.RawMessage(`{"foo":"bar"}`), 5, 10, 100, http.StatusOK, response)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "foo")
}

func TestWriteTimeWindowedResponse(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteTimeWindowedResponse(context.Background(), json.RawMessage(`{"foo":"bar"}`), time.Now().Add(-1*time.Second), time.Now(), http.StatusOK, response)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "foo")
}

func TestWriteResponseWrapperWithTimeWindowAndPagination(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteResponseWrapperWithTimeWindowAndPagination(context.Background(), json.RawMessage(`{"foo":"bar"}`), time.Now().Add(-5*time.Second), time.Now(), 5, 10, 100, http.StatusOK, response)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "foo")
	require.Contains(t, response.Body.String(), "skip")
	require.Contains(t, response.Body.String(), "limit")
	require.Contains(t, response.Body.String(), "start")
	require.Contains(t, response.Body.String(), "end")

	response = httptest.NewRecorder()
	api.WriteResponseWrapperWithTimeWindowAndPagination(context.Background(), json.RawMessage(`{"foo":"bar"}`), time.Time{}, time.Time{}, 5, 10, 100, http.StatusOK, response)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "foo")
	require.Contains(t, response.Body.String(), "skip")
	require.Contains(t, response.Body.String(), "limit")
	require.NotContains(t, response.Body.String(), "start")
	require.NotContains(t, response.Body.String(), "end")
}

func TestWriteBinaryResponse(t *testing.T) {
	response := httptest.NewRecorder()
	api.WriteBinaryResponse(context.Background(), []byte(`{"foo":"bar"}`), "filename", http.StatusOK, response)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "foo")
	require.Contains(t, response.Header().Values(headers.ContentType.String()), mediatypes.ApplicationOctetStream.String())
}
