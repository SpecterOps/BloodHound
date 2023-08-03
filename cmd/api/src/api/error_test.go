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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/utils/test"
	"github.com/stretchr/testify/require"
)

func TestIsErrorResponse(t *testing.T) {

	var cases = []struct {
		Input    int
		Expected bool
	}{
		{http.StatusOK, false},
		{http.StatusInternalServerError, true},
		{http.StatusContinue, true},
	}

	for _, tc := range cases {
		resp := http.Response{
			StatusCode: tc.Input,
		}
		actual := api.IsErrorResponse(&resp)

		if actual != tc.Expected {
			t.Errorf("For Status: %v, got %v, expected %v", tc.Input, actual, tc.Expected)
		}
	}
}

func TestError(t *testing.T) {
	details := "Test error"
	errStruct := api.ErrorWrapper{
		HTTPStatus: http.StatusInternalServerError,
		Timestamp:  time.Now(),
		RequestID:  "abcd",
		Errors: []api.ErrorDetails{
			{
				Context: "test",
				Message: details,
			},
		},
	}

	response := errStruct.Error()
	require.Contains(t, response, details)
}

func TestBuildErrorResponseInternalServerError(t *testing.T) {
	details := "Test error"

	response := api.BuildErrorResponse(http.StatusInternalServerError, details, test.Request(t).WithContext(&ctx.Context{
		RequestID: "12345",
	}).Request())
	require.Equal(t, http.StatusInternalServerError, response.HTTPStatus)

	errorJSON, err := json.Marshal(response)
	require.Nil(t, err)

	errorStruct := api.ErrorWrapper{}
	json.Unmarshal(errorJSON, &errorStruct)

	require.Equal(t, http.StatusInternalServerError, errorStruct.HTTPStatus)
	require.Equal(t, len(errorStruct.Errors), 1)
	require.Equal(t, details, errorStruct.Errors[0].Message)
}

func TestBuildErrorResponseBadRequest(t *testing.T) {
	details := "Test error"
	response := api.BuildErrorResponse(http.StatusBadRequest, details, test.Request(t).WithContext(&ctx.Context{
		RequestID: "12345",
	}).Request())
	require.Equal(t, http.StatusBadRequest, response.HTTPStatus)

	errorJSON, err := json.Marshal(response)
	require.Nil(t, err)

	errorStruct := api.ErrorWrapper{}
	json.Unmarshal(errorJSON, &errorStruct)

	require.Equal(t, http.StatusBadRequest, errorStruct.HTTPStatus)
	require.Equal(t, len(errorStruct.Errors), 1)
	require.Equal(t, details, errorStruct.Errors[0].Message)
}

func TestBuildErrorResponseUnauthorized(t *testing.T) {
	details := "Test error"
	response := api.BuildErrorResponse(http.StatusUnauthorized, details, test.Request(t).WithContext(&ctx.Context{
		RequestID: "12345",
	}).Request())
	require.Equal(t, http.StatusUnauthorized, response.HTTPStatus)

	errorJSON, err := json.Marshal(response)
	require.Nil(t, err)

	errorStruct := api.ErrorWrapper{}
	json.Unmarshal(errorJSON, &errorStruct)

	require.Equal(t, http.StatusUnauthorized, errorStruct.HTTPStatus)
	require.Equal(t, len(errorStruct.Errors), 1)
	require.Equal(t, details, errorStruct.Errors[0].Message)
}

func TestBuildErrorResponseConflict(t *testing.T) {
	details := "Test error"
	response := api.BuildErrorResponse(http.StatusConflict, details, test.Request(t).WithContext(&ctx.Context{
		RequestID: "12345",
	}).Request())
	require.Equal(t, http.StatusConflict, response.HTTPStatus)

	errorJSON, err := json.Marshal(response)
	require.Nil(t, err)

	errorStruct := api.ErrorWrapper{}
	json.Unmarshal(errorJSON, &errorStruct)

	require.Equal(t, http.StatusConflict, errorStruct.HTTPStatus)
	require.Equal(t, len(errorStruct.Errors), 1)
	require.Equal(t, details, errorStruct.Errors[0].Message)
}

func TestBuildErrorResponseDefault(t *testing.T) {
	details := "Test error"
	response := api.BuildErrorResponse(http.StatusTeapot, details, test.Request(t).WithContext(&ctx.Context{
		RequestID: "12345",
	}).Request())
	require.Equal(t, http.StatusTeapot, response.HTTPStatus)

	errorJSON, err := json.Marshal(response)
	require.Nil(t, err)

	errorStruct := api.ErrorWrapper{}
	json.Unmarshal(errorJSON, &errorStruct)

	require.Equal(t, http.StatusTeapot, errorStruct.HTTPStatus)
	require.Equal(t, len(errorStruct.Errors), 1)
	require.Equal(t, details, errorStruct.Errors[0].Message)
}

func TestHandleDatabaseErrorNotFound(t *testing.T) {
	bhCtx := ctx.Context{
		RequestID: "requestID",
		AuthCtx:   auth.Context{},
	}

	request, err := http.NewRequest("GET", "www.foo.bar", strings.NewReader("Hello world"))
	require.Nil(t, err)

	request = request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhCtx.WithRequestID("requestID")))

	response := httptest.NewRecorder()
	err = database.ErrNotFound

	api.HandleDatabaseError(request, response, err)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsResourceNotFound)
}

func TestHandleDatabaseError(t *testing.T) {
	bhCtx := ctx.Context{
		RequestID: "requestID",
		AuthCtx:   auth.Context{},
	}

	request, err := http.NewRequest("GET", "www.foo.bar", strings.NewReader("Hello world"))
	require.Nil(t, err)

	request = request.WithContext(context.WithValue(context.Background(), ctx.ValueKey, bhCtx.WithRequestID("requestID")))

	response := httptest.NewRecorder()
	err = errors.New("custom error")

	api.HandleDatabaseError(request, response, err)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
}

func TestFormatDatabaseErrorNotFound(t *testing.T) {
	err := api.FormatDatabaseError(database.ErrNotFound)
	require.Equal(t, err.Error(), api.ErrorResponseDetailsResourceNotFound)
}

func TestFormatDatabaseError(t *testing.T) {
	err := api.FormatDatabaseError(errors.New("custom error"))
	require.Equal(t, err.Error(), api.ErrorResponseDetailsInternalServerError)
}
