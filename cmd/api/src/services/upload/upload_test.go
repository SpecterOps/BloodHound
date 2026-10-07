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

package upload_test

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
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
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

	actual := upload.AllowedFileUploadTypes()
	require.Equal(t, expected, actual)

	actual[0] = "mutated"
	require.Equal(t, expected, upload.AllowedFileUploadTypes())
}

func TestAllowedZipFileUploadTypes(t *testing.T) {
	expected := []string{
		"application/zip",
		"application/x-zip-compressed",
		"application/zip-compressed",
	}

	actual := upload.AllowedZipFileUploadTypes()
	require.Equal(t, expected, actual)

	actual[0] = "mutated"
	require.Equal(t, expected, upload.AllowedZipFileUploadTypes())
}

func TestWriteAndValidateZip(t *testing.T) {
	t.Run("valid zip file is ok", func(t *testing.T) {
		var writer bytes.Buffer

		file, err := os.Open("../../test/fixtures/fixtures/goodzip.zip")
		require.NoError(t, err)
		t.Cleanup(func() { _ = file.Close() })

		err = upload.WriteAndValidateZip(file, &writer)
		require.NoError(t, err)
		require.NotEmpty(t, writer.Bytes())
	})

	t.Run("invalid bytes causes error", func(t *testing.T) {
		var writer bytes.Buffer
		badZip := strings.NewReader("123123")

		err := upload.WriteAndValidateZip(badZip, &writer)
		assert.ErrorIs(t, err, upload.ErrInvalidZipFile)
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

			report, err := upload.WriteAndValidateJSON(src, &destination, schema)
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
	src := &errorReader{err: errors.New("read error")}

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	_, err = upload.WriteAndValidateJSON(src, &destination, schema)

	require.Error(t, err)
	assert.ErrorIs(t, err, upload.ErrInvalidJSON)
}

func TestSaveIngestFile(t *testing.T) {
	t.Parallel()

	var (
		jsonPayload        = []byte(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`)
		zipPayload, zipErr = os.ReadFile("../../test/fixtures/fixtures/goodzip.zip")
	)
	require.NoError(t, zipErr)

	tests := []struct {
		name          string
		contentType   string
		input         []byte
		fileType      model.FileType
		expectedData  []byte
		expectedError error
	}{
		{
			name:         "persists normalized JSON",
			contentType:  "application/json",
			input:        append([]byte{0xEF, 0xBB, 0xBF}, jsonPayload...),
			fileType:     model.FileTypeJson,
			expectedData: jsonPayload,
		},
		{
			name:         "persists the complete ZIP",
			contentType:  "application/zip",
			input:        zipPayload,
			fileType:     model.FileTypeZip,
			expectedData: zipPayload,
		},
		{
			name:          "rejects missing payload metadata",
			contentType:   "application/json",
			input:         []byte(`{"data":[{"domain":"example.com"}]}`),
			expectedError: payload.ErrInvalidFileConfiguration,
		},
		{
			name:          "rejects malformed JSON",
			contentType:   "application/json",
			input:         []byte(`{"meta":`),
			expectedError: upload.ErrInvalidJSON,
		},
		{
			name:          "rejects invalid ZIP bytes",
			contentType:   "application/zip",
			input:         []byte("invalid archive"),
			expectedError: upload.ErrInvalidZipFile,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx           = context.Background()
				rootDirectory = t.TempDir()
				request       = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(testCase.input))
			)
			localStore, err := storage.NewLocalStore(rootDirectory)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, localStore.Close()) })
			fileService := storage.NewFileService(localStore)
			request.Header.Set("Content-Type", testCase.contentType)
			schema, err := payload.LoadSchema()
			require.NoError(t, err)

			// Act
			params, report, err := upload.SaveIngestFile(ctx, fileService, request, schema, 1)

			// Assert
			if testCase.expectedError != nil {
				require.ErrorIs(t, err, testCase.expectedError)
				require.Empty(t, params)
				files, readErr := os.ReadDir(rootDirectory)
				require.NoError(t, readErr)
				require.Empty(t, files)
				if testCase.contentType == "application/json" {
					require.NotEmpty(t, report.CriticalErrors)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.fileType, params.FileType)
				require.NotEmpty(t, params.Filename)
				storedData, readErr := fileService.ReadFile(ctx, params.Filename)
				require.NoError(t, readErr)
				require.Equal(t, testCase.expectedData, storedData)
				require.Empty(t, report.CriticalErrors)
				require.Empty(t, report.ValidationErrors)
			}
		})
	}
}

func TestSaveIngestFileStorageFailure(t *testing.T) {
	t.Parallel()

	for _, consumeBody := range []bool{false, true} {
		name := "storage fails before reading"
		if consumeBody {
			name = "storage fails after reading"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				writeErr        = errors.New("storage write failed")
				mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
				request         = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`))
			)
			request.Header.Set("Content-Type", "application/json")
			schema, err := payload.LoadSchema()
			require.NoError(t, err)
			mockFileService.EXPECT().WriteTempFile(gomock.Any(), "file_upload_job1_", gomock.Any(), storage.WriteOptions{}).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
					if consumeBody {
						_, readErr := io.Copy(io.Discard, reader)
						require.NoError(t, readErr)
					}
					return "", writeErr
				})

			// Act
			params, report, err := upload.SaveIngestFile(context.Background(), mockFileService, request, schema, 1)

			// Assert
			require.ErrorIs(t, err, writeErr)
			require.NotErrorIs(t, err, upload.ErrInvalidJSON)
			require.Empty(t, params)
			require.Empty(t, report.CriticalErrors)
			require.Empty(t, report.ValidationErrors)
		})
	}
}

func TestSaveIngestFileCancellationDeletesStoredFile(t *testing.T) {
	t.Parallel()

	// Arrange
	var (
		ctx, cancel     = context.WithCancel(context.Background())
		rootDirectory   = t.TempDir()
		mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
		request         = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`))
	)
	defer cancel()
	localStore, err := storage.NewLocalStore(rootDirectory)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, localStore.Close()) })
	fileService := storage.NewFileService(localStore)
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	mockFileService.EXPECT().WriteTempFile(ctx, "file_upload_job1_", gomock.Any(), storage.WriteOptions{}).
		DoAndReturn(func(writeContext context.Context, prefix string, reader io.Reader, options storage.WriteOptions) (string, error) {
			fileName, writeErr := fileService.WriteTempFile(writeContext, prefix, reader, options)
			require.NoError(t, writeErr)
			cancel()
			return fileName, nil
		})
	mockFileService.EXPECT().DeleteFile(gomock.Any(), gomock.Any()).
		DoAndReturn(func(cleanupContext context.Context, fileName string) error {
			require.NoError(t, cleanupContext.Err())
			return fileService.DeleteFile(cleanupContext, fileName)
		})

	// Act
	params, report, err := upload.SaveIngestFile(ctx, mockFileService, request, schema, 1)

	// Assert
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
	require.Empty(t, report.ValidationErrors)
	files, err := os.ReadDir(rootDirectory)
	require.NoError(t, err)
	require.Empty(t, files)
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
	rootDirectory := t.TempDir()
	localStore, err := storage.NewLocalStore(rootDirectory)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, localStore.Close()) })
	fileService := storage.NewFileService(localStore)
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
	result := make(chan saveIngestResult, 1)
	go func() {
		params, report, saveErr := upload.SaveIngestFile(ctx, fileService, request, schema, 1)
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
		files, readErr := os.ReadDir(rootDirectory)
		require.NoError(t, readErr)
		require.Empty(t, files)
	case <-time.After(2 * time.Second):
		t.Fatal("SaveIngestFile remained blocked after request cancellation")
	}
}

func TestSaveIngestFileStorageFailureUnblocksBody(t *testing.T) {
	t.Parallel()

	// Arrange
	var (
		input           = []byte(`{"meta":{"type":"domains","version":4,"count":1},"data":[{"domain":"example.com"}]}`)
		writeErr        = errors.New("storage failed")
		mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
		body            = &cancellationUploadBody{
			input:    bytes.NewReader(input),
			blocked:  make(chan struct{}),
			released: make(chan struct{}),
		}
		request = httptest.NewRequest(http.MethodPost, "/", nil)
	)
	t.Cleanup(func() { _ = body.Close() })
	request.Body = body
	request.Header.Set("Content-Type", "application/json")
	schema, err := payload.LoadSchema()
	require.NoError(t, err)
	mockFileService.EXPECT().WriteTempFile(gomock.Any(), "file_upload_job1_", gomock.Any(), storage.WriteOptions{}).
		DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
			_, readErr := io.ReadFull(reader, make([]byte, len(input)))
			require.NoError(t, readErr)
			select {
			case <-body.blocked:
			case <-time.After(2 * time.Second):
				t.Fatal("request body did not block")
			}
			return "", writeErr
		})
	result := make(chan saveIngestResult, 1)
	go func() {
		params, report, saveErr := upload.SaveIngestFile(context.Background(), mockFileService, request, schema, 1)
		result <- saveIngestResult{params: params, report: report, err: saveErr}
	}()

	// Assert
	select {
	case saveResult := <-result:
		require.ErrorIs(t, saveResult.err, writeErr)
		require.NotErrorIs(t, saveResult.err, upload.ErrInvalidJSON)
		require.Empty(t, saveResult.params)
		require.Empty(t, saveResult.report.CriticalErrors)
		require.Empty(t, saveResult.report.ValidationErrors)
	case <-time.After(2 * time.Second):
		t.Fatal("SaveIngestFile remained blocked after storage failed")
	}
}

type saveIngestResult struct {
	params upload.IngestTaskParams
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

	// Act
	params, report, err := upload.SaveIngestFile(ctx, mockFileService, request, schema, 1)

	// Assert
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
	require.Empty(t, report.ValidationErrors)
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
	params, report, err := upload.SaveIngestFile(context.Background(), mockFileService, request, schema, 1)

	// Assert
	require.EqualError(t, err, "invalid content type for ingest file")
	require.Empty(t, params)
	require.Empty(t, report.CriticalErrors)
	require.Empty(t, report.ValidationErrors)
}

type errorReader struct {
	err error
}

func (s *errorReader) Read([]byte) (int, error) {
	return 0, s.err
}
