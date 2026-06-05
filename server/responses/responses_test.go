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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/server/responses"
	"github.com/specterops/bloodhound/server/responses/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubCSVViewer is a minimal responses.CSVViewer used to exercise WriteCSV.
type stubCSVViewer struct {
	payload []byte
	err     error
}

func (s stubCSVViewer) WriteCSV(writer io.Writer) error {
	if s.err != nil {
		return s.err
	}
	_, err := writer.Write(s.payload)
	return err
}

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

func TestWritePaginated(t *testing.T) {
	var (
		ctx       = newTestContext()
		validData = []byte(`[{"id":1}]`)
	)

	t.Run("writes paginated envelope with metadata on success", func(t *testing.T) {
		var (
			viewerMock = mocks.NewMockJSONViewer(t)
			recorder   = httptest.NewRecorder()
		)

		viewerMock.EXPECT().JSONView().Return(validData, nil)

		responses.WritePaginated(ctx, viewerMock, 10, 20, 42, http.StatusOK, recorder)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var envelope responses.PaginatedResponse
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		assert.Equal(t, 10, envelope.Limit)
		assert.Equal(t, 20, envelope.Skip)
		assert.Equal(t, 42, envelope.Count)
		assert.JSONEq(t, string(validData), string(envelope.Data))
	})

	t.Run("omits non-positive limit and negative count", func(t *testing.T) {
		var (
			viewerMock = mocks.NewMockJSONViewer(t)
			recorder   = httptest.NewRecorder()
		)

		viewerMock.EXPECT().JSONView().Return(validData, nil)

		responses.WritePaginated(ctx, viewerMock, -1, 5, -1, http.StatusOK, recorder)

		var envelope responses.PaginatedResponse
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		assert.Equal(t, 0, envelope.Limit)
		assert.Equal(t, 0, envelope.Count)
		assert.Equal(t, 5, envelope.Skip)
	})

	t.Run("writes 500 error when JSONView fails", func(t *testing.T) {
		var (
			viewerMock = mocks.NewMockJSONViewer(t)
			recorder   = httptest.NewRecorder()
		)

		viewerMock.EXPECT().JSONView().Return(nil, errors.New("marshal failed"))

		responses.WritePaginated(ctx, viewerMock, 10, 0, 1, http.StatusOK, recorder)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Failed to marshal response")
	})
}

func TestWriteCSV(t *testing.T) {
	var ctx = newTestContext()

	t.Run("writes CSV body with text/csv content type", func(t *testing.T) {
		var recorder = httptest.NewRecorder()

		responses.WriteCSV(ctx, stubCSVViewer{payload: []byte("a,b\n1,2\n")}, http.StatusOK, recorder)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "text/csv", recorder.Header().Get("Content-Type"))
		assert.Equal(t, "a,b\n1,2\n", recorder.Body.String())
	})

	t.Run("writes 500 when serialization fails", func(t *testing.T) {
		var recorder = httptest.NewRecorder()

		responses.WriteCSV(ctx, stubCSVViewer{err: errors.New("csv boom")}, http.StatusOK, recorder)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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
