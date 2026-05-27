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

package responses_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/server/jsonapiv2/responses"
	"github.com/specterops/bloodhound/server/jsonapiv2/responses/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext() context.Context {
	var bhCtx = &bhctx.Context{
		RequestID: "test-request-123",
	}
	return bhctx.Set(context.Background(), bhCtx)
}

func TestWriteBasic(t *testing.T) {
	var (
		ctx        = newTestContext()
		validData  = []byte(`{"key":"value"}`)
		marshalErr = errors.New("marshal failed")
	)

	t.Run("writes JSON envelope with data on success", func(t *testing.T) {
		var (
			viewerMock = mocks.NewMockJSONViewer(t)
			recorder   = httptest.NewRecorder()
		)

		viewerMock.EXPECT().JSONView().Return(validData, nil)

		responses.WriteBasic(ctx, viewerMock, http.StatusOK, recorder)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var envelope responses.BasicResponse
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		assert.JSONEq(t, string(validData), string(envelope.Data))
	})

	t.Run("writes 500 error when JSONView fails", func(t *testing.T) {
		var (
			viewerMock = mocks.NewMockJSONViewer(t)
			recorder   = httptest.NewRecorder()
		)

		viewerMock.EXPECT().JSONView().Return(nil, marshalErr)

		responses.WriteBasic(ctx, viewerMock, http.StatusOK, recorder)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Failed to marshal response")
	})
}

func TestWriteError(t *testing.T) {
	var (
		ctx      = newTestContext()
		recorder = httptest.NewRecorder()
	)

	responses.WriteError(ctx, http.StatusBadRequest, "invalid input", recorder)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var errorWrapper responses.ErrorWrapper
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &errorWrapper))

	assert.Equal(t, http.StatusBadRequest, errorWrapper.HTTPStatus)
	assert.Equal(t, "test-request-123", errorWrapper.RequestID)
	assert.NotZero(t, errorWrapper.Timestamp)
	require.Len(t, errorWrapper.Errors, 1)
	assert.Equal(t, "invalid input", errorWrapper.Errors[0].Message)
}

func TestWriteNoContent(t *testing.T) {
	var recorder = httptest.NewRecorder()

	responses.WriteNoContent(recorder)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Empty(t, recorder.Body.Bytes())
}

func TestWriteInternalServerError(t *testing.T) {
	var (
		ctx      = newTestContext()
		recorder = httptest.NewRecorder()
		cause    = errors.New("database connection failed")
	)

	responses.WriteInternalServerError(ctx, cause, recorder)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var errorWrapper responses.ErrorWrapper
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &errorWrapper))

	assert.Equal(t, http.StatusInternalServerError, errorWrapper.HTTPStatus)
	assert.Equal(t, "test-request-123", errorWrapper.RequestID)
	require.Len(t, errorWrapper.Errors, 1)
	assert.Contains(t, errorWrapper.Errors[0].Message, "internal error")
}
