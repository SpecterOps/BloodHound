// Copyright 2024 Specter Ops, Inc.
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

package upload

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/chow/payload"
	"github.com/specterops/bloodhound/packages/go/storage"
	storagemocks "github.com/specterops/bloodhound/packages/go/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAllowedFileUploadTypes(t *testing.T) {
	expected := []string{
		"application/json",
		"application/zip",
		"application/x-zip-compressed",
		"application/zip-compressed",
	}

	actual := AllowedFileUploadTypes()
	require.Equal(t, expected, actual)

	actual[0] = "mutated"
	require.Equal(t, expected, AllowedFileUploadTypes())
}

func TestAllowedZipFileUploadTypes(t *testing.T) {
	expected := []string{
		"application/zip",
		"application/x-zip-compressed",
		"application/zip-compressed",
	}

	actual := AllowedZipFileUploadTypes()
	require.Equal(t, expected, actual)

	actual[0] = "mutated"
	require.Equal(t, expected, AllowedZipFileUploadTypes())
}

func TestWriteAndValidateZip(t *testing.T) {
	t.Run("valid zip file is ok", func(t *testing.T) {
		var writer bytes.Buffer

		file, err := os.Open("../../test/fixtures/fixtures/goodzip.zip")
		require.NoError(t, err)
		t.Cleanup(func() { _ = file.Close() })

		err = WriteAndValidateZip(file, &writer)
		require.NoError(t, err)
		require.NotEmpty(t, writer.Bytes())
	})

	t.Run("invalid bytes causes error", func(t *testing.T) {
		var writer bytes.Buffer
		badZip := strings.NewReader("123123")

		err := WriteAndValidateZip(badZip, &writer)
		assert.ErrorIs(t, err, ErrInvalidZipFile)
	})
}

func TestWriteAndValidateJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectedOutput []byte
		expectedError  error
	}{
		{
			name:           "UTF-8 without BOM",
			input:          []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`),
		},
		{
			name:           "UTF-8 with BOM",
			input:          append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`)...),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`),
		},
		{
			name:           "UTF-16BE with BOM",
			input:          []byte{0xFE, 0xFF, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x6D, 0x00, 0x65, 0x00, 0x74, 0x00, 0x61, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x74, 0x00, 0x79, 0x00, 0x70, 0x00, 0x65, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x22, 0x00, 0x64, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x61, 0x00, 0x69, 0x00, 0x6E, 0x00, 0x73, 0x00, 0x22, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x76, 0x00, 0x65, 0x00, 0x72, 0x00, 0x73, 0x00, 0x69, 0x00, 0x6F, 0x00, 0x6E, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x34, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x63, 0x00, 0x6F, 0x00, 0x75, 0x00, 0x6E, 0x00, 0x74, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x31, 0x00, 0x7D, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x64, 0x00, 0x61, 0x00, 0x74, 0x00, 0x61, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x5B, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x64, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x61, 0x00, 0x69, 0x00, 0x6E, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x22, 0x00, 0x65, 0x00, 0x78, 0x00, 0x61, 0x00, 0x6D, 0x00, 0x70, 0x00, 0x6C, 0x00, 0x65, 0x00, 0x2E, 0x00, 0x63, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x22, 0x00, 0x7D, 0x00, 0x5D, 0x00, 0x7D},
			expectedOutput: []byte{0x7b, 0x22, 0x6d, 0x65, 0x74, 0x61, 0x22, 0x3a, 0x20, 0x7b, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x22, 0x2c, 0x20, 0x22, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x34, 0x2c, 0x20, 0x22, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x3a, 0x20, 0x31, 0x7d, 0x2c, 0x20, 0x22, 0x64, 0x61, 0x74, 0x61, 0x22, 0x3a, 0x20, 0x5b, 0x7b, 0x22, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0x3a, 0x20, 0x22, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x22, 0x7d, 0x5d, 0x7d},
		},
		{
			name:           "Missing meta tag",
			input:          []byte(`{"data": [{"domain": "example.com"}]}`),
			expectedOutput: []byte(`{"data": [{"domain": "example.com"}]}`),
			expectedError:  payload.ErrInvalidFileConfiguration,
		},
		{
			name:           "Missing data tag",
			input:          []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}}`),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}}`),
			expectedError:  payload.ErrInvalidFileConfiguration,
		},
	}

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var destination bytes.Buffer
			src := bytes.NewReader(testCase.input)

			report, err := WriteAndValidateJSON(src, &destination, schema)
			if testCase.expectedError != nil {
				require.ErrorIs(t, err, testCase.expectedError)
				require.NotEmpty(t, report.CriticalErrors)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedOutput, destination.Bytes())
		})
	}
}

func TestWriteAndValidateJSON_NormalizationError(t *testing.T) {
	var destination bytes.Buffer
	src := &ErrorReader{err: errors.New("read error")}

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	_, err = WriteAndValidateJSON(src, &destination, schema)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidJSON)
}

type memoryPendingFile struct {
	data        bytes.Buffer
	writeErr    error
	commitName  string
	commitErr   error
	commitCalls int
	abortCalls  int
}

func (s *memoryPendingFile) Write(data []byte) (int, error) {
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	return s.data.Write(data)
}

func (s *memoryPendingFile) Commit() (string, error) {
	s.commitCalls++
	return s.commitName, s.commitErr
}

func (s *memoryPendingFile) Abort() error {
	s.abortCalls++
	return nil
}

func TestSaveIngestFilePersistsNormalizedJSON(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	jsonPayload := []byte(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`)
	input := append([]byte{0xEF, 0xBB, 0xBF}, jsonPayload...)
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(input))
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	pendingFile := &memoryPendingFile{commitName: "tmp-json"}
	mockFileService.EXPECT().BeginTempFile(ctx, ingestFileTempPrefix(1), storage.WriteOptions{SizeHint: int64(len(input))}).Return(pendingFile, nil)

	// Act
	params, report, err := SaveIngestFile(ctx, mockFileService, request, schema, 1)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "tmp-json", params.Filename)
	require.Equal(t, model.FileTypeJson, params.FileType)
	require.Equal(t, jsonPayload, pendingFile.data.Bytes())
	require.Empty(t, report.CriticalErrors)
	require.Equal(t, 1, pendingFile.commitCalls)
	require.Equal(t, 1, pendingFile.abortCalls)
}

func TestSaveIngestFilePersistsCompleteZIP(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", "application/zip")
	zipBytes, err := os.ReadFile("../../test/fixtures/fixtures/goodzip.zip")
	require.NoError(t, err)
	request.Body = io.NopCloser(bytes.NewReader(zipBytes))
	request.ContentLength = int64(len(zipBytes))
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	pendingFile := &memoryPendingFile{commitName: "tmp-zip"}
	mockFileService.EXPECT().BeginTempFile(ctx, ingestFileTempPrefix(1), storage.WriteOptions{SizeHint: int64(len(zipBytes))}).Return(pendingFile, nil)

	// Act
	params, report, err := SaveIngestFile(ctx, mockFileService, request, schema, 1)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "tmp-zip", params.Filename)
	require.Equal(t, model.FileTypeZip, params.FileType)
	require.Equal(t, zipBytes, pendingFile.data.Bytes())
	require.Empty(t, report.CriticalErrors)
	require.Equal(t, 1, pendingFile.commitCalls)
	require.Equal(t, 1, pendingFile.abortCalls)
}

func TestSaveIngestFileValidationFailureDoesNotCommit(t *testing.T) {
	t.Parallel()

	// Arrange
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"data":[{"domain":"example.com"}]}`))
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	pendingFile := &memoryPendingFile{commitName: "tmp-invalid"}
	mockFileService.EXPECT().BeginTempFile(gomock.Any(), ingestFileTempPrefix(1), gomock.Any()).Return(pendingFile, nil)

	// Act
	params, report, err := SaveIngestFile(context.Background(), mockFileService, request, schema, 1)

	// Assert
	require.ErrorIs(t, err, ErrInvalidJSON)
	require.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)
	require.Empty(t, params)
	require.NotEmpty(t, report.CriticalErrors)
	require.Equal(t, 0, pendingFile.commitCalls)
	require.Equal(t, 1, pendingFile.abortCalls)
}

func TestSaveIngestFileWriteFailureIsStorageError(t *testing.T) {
	t.Parallel()

	// Arrange
	writeCause := errors.New("disk write failed")
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`))
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	pendingFile := &memoryPendingFile{writeErr: errors.Join(storage.ErrWriteFailed, writeCause)}
	mockFileService.EXPECT().BeginTempFile(gomock.Any(), ingestFileTempPrefix(1), gomock.Any()).Return(pendingFile, nil)

	// Act
	params, report, err := SaveIngestFile(context.Background(), mockFileService, request, schema, 1)

	// Assert
	require.ErrorIs(t, err, storage.ErrWriteFailed)
	require.ErrorIs(t, err, writeCause)
	require.NotErrorIs(t, err, ErrInvalidJSON)
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
	require.Empty(t, report.ValidationErrors)
	require.Equal(t, 0, pendingFile.commitCalls)
	require.Equal(t, 1, pendingFile.abortCalls)
}

func TestSaveIngestFileBeginFailure(t *testing.T) {
	t.Parallel()

	// Arrange
	beginErr := errors.New("begin failed")
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`))
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	mockFileService.EXPECT().BeginTempFile(gomock.Any(), ingestFileTempPrefix(1), gomock.Any()).Return(nil, beginErr)

	// Act
	params, report, err := SaveIngestFile(context.Background(), mockFileService, request, schema, 1)

	// Assert
	require.ErrorIs(t, err, beginErr)
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
}

func TestSaveIngestFileCommitFailureAbortsPendingFile(t *testing.T) {
	t.Parallel()

	// Arrange
	commitErr := errors.New("commit failed")
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`))
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	pendingFile := &memoryPendingFile{commitErr: commitErr}
	mockFileService.EXPECT().BeginTempFile(gomock.Any(), ingestFileTempPrefix(1), gomock.Any()).Return(pendingFile, nil)

	// Act
	params, report, err := SaveIngestFile(context.Background(), mockFileService, request, schema, 1)

	// Assert
	require.ErrorIs(t, err, commitErr)
	require.NotErrorIs(t, err, ErrInvalidJSON)
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
	require.Equal(t, 1, pendingFile.commitCalls)
	require.Equal(t, 1, pendingFile.abortCalls)
}

type cancellationUploadBody struct {
	input       *bytes.Reader
	blocked     chan struct{}
	released    chan struct{}
	releaseOnce sync.Once
	blockedOnce sync.Once
}

func (s *cancellationUploadBody) Read(buffer []byte) (int, error) {
	count, err := s.input.Read(buffer)
	if err != io.EOF {
		return count, err
	}
	s.blockedOnce.Do(func() { close(s.blocked) })
	<-s.released
	return count, err
}

func (s *cancellationUploadBody) Close() error {
	s.releaseOnce.Do(func() { close(s.released) })
	return nil
}

func TestSaveIngestFileCancellationUnblocksBody(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	body := &cancellationUploadBody{
		input:    bytes.NewReader([]byte(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`)),
		blocked:  make(chan struct{}),
		released: make(chan struct{}),
	}
	t.Cleanup(func() { _ = body.Close() })
	request := httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx)
	request.Body = body
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	pendingFile := &memoryPendingFile{}
	mockFileService.EXPECT().BeginTempFile(ctx, ingestFileTempPrefix(1), gomock.Any()).Return(pendingFile, nil)
	result := make(chan saveIngestResult, 1)
	go func() {
		params, report, saveErr := SaveIngestFile(ctx, mockFileService, request, schema, 1)
		result <- saveIngestResult{
			params: params,
			report: report,
			err:    saveErr,
		}
	}()

	// Wait for validation to reach the blocked body read.
	select {
	case <-body.blocked:
	case <-time.After(2 * time.Second):
		t.Fatal("request body did not block")
	}

	// Act
	cancel()

	// Assert
	select {
	case saveResult := <-result:
		require.ErrorIs(t, saveResult.err, context.Canceled)
		require.Empty(t, saveResult.params)
		require.Empty(t, saveResult.report.CriticalErrors)
		require.Empty(t, saveResult.report.ValidationErrors)
		require.Equal(t, 0, pendingFile.commitCalls)
		require.Equal(t, 1, pendingFile.abortCalls)
	case <-time.After(2 * time.Second):
		t.Fatal("SaveIngestFile remained blocked after request cancellation")
	}
}

type saveIngestResult struct {
	params IngestTaskParams
	report payload.ValidationReport
	err    error
}

func TestSaveIngestFileCancellationWithNilBody(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	request := httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx)
	request.Body = nil
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	pendingFile := &memoryPendingFile{}
	mockFileService.EXPECT().BeginTempFile(ctx, ingestFileTempPrefix(1), gomock.Any()).Return(pendingFile, nil)

	// Act
	params, report, err := SaveIngestFile(ctx, mockFileService, request, schema, 1)

	// Assert
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
	require.Empty(t, report.ValidationErrors)
	require.Equal(t, 0, pendingFile.commitCalls)
	require.Equal(t, 1, pendingFile.abortCalls)
}

func TestSaveIngestFileRejectsUnsupportedContentTypeWithoutWriting(t *testing.T) {
	t.Parallel()

	// Arrange
	mockFileService := storagemocks.NewMockFileService(gomock.NewController(t))
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("content"))
	request.Header.Set("Content-Type", "text/plain")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	// Act
	params, report, err := SaveIngestFile(context.Background(), mockFileService, request, schema, 1)

	// Assert
	require.EqualError(t, err, "invalid content type for ingest file")
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
	require.Empty(t, report.ValidationErrors)
}

// ErrorReader is a mock reader that always returns an error.
type ErrorReader struct {
	err error
}

func (s *ErrorReader) Read([]byte) (int, error) {
	return 0, s.err
}
