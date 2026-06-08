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

package tools

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/storage"
	storagemocks "github.com/specterops/bloodhound/packages/go/storage/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type retainedIngestParameterService struct {
	retain bool
	setErr error
}

func (s retainedIngestParameterService) GetAllConfigurationParameters(ctx context.Context) (appcfg.Parameters, error) {
	return nil, nil
}

func (s retainedIngestParameterService) GetConfigurationParameter(ctx context.Context, parameterKey appcfg.ParameterKey) (appcfg.Parameter, error) {
	value, err := types.NewJSONBObject(appcfg.RetainIngestedFilesParameter{
		Enabled: s.retain,
	})
	if err != nil {
		return appcfg.Parameter{}, err
	}

	return appcfg.Parameter{
		Key:   parameterKey,
		Value: value,
	}, nil
}

func (s retainedIngestParameterService) SetConfigurationParameter(ctx context.Context, configurationParameter appcfg.Parameter) error {
	return s.setErr
}

type retainedIngestErrorReadCloser struct {
	err error
}

func (s retainedIngestErrorReadCloser) Read([]byte) (int, error) {
	return 0, s.err
}

func (s retainedIngestErrorReadCloser) Close() error {
	return nil
}

func readRetainedTar(t *testing.T, body []byte) map[string]string {
	t.Helper()

	gzipReader, err := gzip.NewReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer gzipReader.Close()

	var (
		files     = map[string]string{}
		tarReader = tar.NewReader(gzipReader)
	)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		content, err := io.ReadAll(tarReader)
		require.NoError(t, err)
		files[header.Name] = string(content)
	}

	return files
}

func TestRetainedArchiveName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "keeps nested logical path",
			filePath: "nested/file.json",
			expected: "nested/file.json",
		},
		{
			name:     "trims leading slash",
			filePath: "/nested/file.json",
			expected: "nested/file.json",
		},
		{
			name:     "keeps only base name for parent traversal",
			filePath: "../outside.json",
			expected: "outside.json",
		},
		{
			name:     "keeps only base name for absolute parent traversal",
			filePath: "/../outside.json",
			expected: "outside.json",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, retainedArchiveName(testCase.filePath))
		})
	}
}

func TestIngestControl_writeRetainedFileToTar(t *testing.T) {
	t.Parallel()

	var (
		errGet  = errors.New("get failed")
		errRead = errors.New("read failed")
	)

	type expected struct {
		errIs       error
		errContains string
		files       map[string]string
	}

	type testData struct {
		name      string
		filePath  string
		setupMock func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService)
		expected  expected
	}

	tests := []testData{
		{
			name:     "writes retained file to tar",
			filePath: "nested/file.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					GetFile(ctx, "nested/file.json").
					Return(
						io.NopCloser(bytes.NewReader([]byte("file content"))),
						storage.FileInfo{
							Path:         "nested/file.json",
							Size:         int64(len("file content")),
							LastModified: time.Now().UTC(),
						},
						nil,
					)
			},
			expected: expected{
				files: map[string]string{
					"nested/file.json": "file content",
				},
			},
		},
		{
			name:     "get error returns wrapped error",
			filePath: "missing.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					GetFile(ctx, "missing.json").
					Return(nil, storage.FileInfo{}, errGet)
			},
			expected: expected{
				errIs:       errGet,
				errContains: "opening retained file",
			},
		},
		{
			name:     "reader error returns wrapped error",
			filePath: "file.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					GetFile(ctx, "file.json").
					Return(
						retainedIngestErrorReadCloser{err: errRead},
						storage.FileInfo{
							Path: "file.json",
							Size: int64(len("file content")),
						},
						nil,
					)
			},
			expected: expected{
				errIs:       errRead,
				errContains: "copying retained file",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx             = context.Background()
				archiveBuffer   = &bytes.Buffer{}
				tarWriter       = tar.NewWriter(archiveBuffer)
				mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
				ingestControl   = NewIngestControlTool(retainedIngestParameterService{retain: true}, mockFileService)
			)

			testCase.setupMock(t, ctx, mockFileService)

			// Act
			err := ingestControl.writeRetainedFileToTar(ctx, tarWriter, testCase.filePath)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Contains(t, err.Error(), testCase.expected.errContains)
				return
			}

			require.NoError(t, err)
			require.NoError(t, tarWriter.Close())

			tarReader := tar.NewReader(bytes.NewReader(archiveBuffer.Bytes()))
			for expectedName, expectedContent := range testCase.expected.files {
				header, err := tarReader.Next()
				require.NoError(t, err)
				require.Equal(t, expectedName, header.Name)

				content, err := io.ReadAll(tarReader)
				require.NoError(t, err)
				require.Equal(t, expectedContent, string(content))
			}
		})
	}
}

func TestIngestControl_FetchRetainedIngestFiles(t *testing.T) {
	t.Parallel()

	var (
		errList = errors.New("list failed")
		errGet  = errors.New("get failed")
	)

	type expected struct {
		statusCode int
		files      map[string]string
	}

	type testData struct {
		name      string
		retain    bool
		setupMock func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService)
		expected  expected
	}

	tests := []testData{
		{
			name:   "retention disabled returns not found",
			retain: false,
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:   "list error returns internal server error",
			retain: true,
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					ListFiles(ctx, "", storage.ListOptions{Recursive: true}).
					Return(nil, errList)
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name:   "writes retained files to gzip tar and skips directories",
			retain: true,
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					ListFiles(ctx, "", storage.ListOptions{Recursive: true}).
					Return([]storage.FileInfo{
						{Path: "nested", IsDir: true},
						{Path: "nested/one.json"},
						{Path: "../outside.json"},
					}, nil)
				mockFileService.EXPECT().
					GetFile(ctx, "nested/one.json").
					Return(
						io.NopCloser(bytes.NewReader([]byte("one"))),
						storage.FileInfo{
							Path:         "nested/one.json",
							Size:         int64(len("one")),
							LastModified: time.Now().UTC(),
						},
						nil,
					)
				mockFileService.EXPECT().
					GetFile(ctx, "../outside.json").
					Return(
						io.NopCloser(bytes.NewReader([]byte("outside"))),
						storage.FileInfo{
							Path:         "../outside.json",
							Size:         int64(len("outside")),
							LastModified: time.Now().UTC(),
						},
						nil,
					)
			},
			expected: expected{
				statusCode: http.StatusOK,
				files: map[string]string{
					"nested/one.json": "one",
					"outside.json":    "outside",
				},
			},
		},
		{
			name:   "get error ends archive without failing response",
			retain: true,
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					ListFiles(ctx, "", storage.ListOptions{Recursive: true}).
					Return([]storage.FileInfo{{Path: "missing.json"}}, nil)
				mockFileService.EXPECT().
					GetFile(ctx, "missing.json").
					Return(nil, storage.FileInfo{}, errGet)
			},
			expected: expected{
				statusCode: http.StatusOK,
				files:      map[string]string{},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				request         = httptest.NewRequest(http.MethodGet, "/api/v2/tools/ingest/retained", nil)
				response        = httptest.NewRecorder()
				mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
				ingestControl   = NewIngestControlTool(retainedIngestParameterService{retain: testCase.retain}, mockFileService)
			)

			if testCase.setupMock != nil {
				testCase.setupMock(t, request.Context(), mockFileService)
			}

			// Act
			ingestControl.FetchRetainedIngestFiles(response, request)

			// Assert
			require.Equal(t, testCase.expected.statusCode, response.Code)
			if testCase.expected.statusCode != http.StatusOK {
				return
			}

			require.Equal(t, "application/gzip", response.Header().Get(headers.ContentType.String()))
			require.Equal(t, "gzip", response.Header().Get(headers.ContentEncoding.String()))
			require.Contains(t, response.Header().Get(headers.ContentDisposition.String()), `attachment; filename="retained-ingest-`)
			require.Equal(t, testCase.expected.files, readRetainedTar(t, response.Body.Bytes()))
		})
	}
}

func TestIngestControl_DisableIngestFileRetention(t *testing.T) {
	t.Parallel()

	t.Run("deletes retained files and skips directories", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			request         = httptest.NewRequest(http.MethodPost, "/api/v2/tools/ingest/retained/disable", nil)
			response        = httptest.NewRecorder()
			done            = make(chan struct{})
			mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
			ingestControl   = NewIngestControlTool(retainedIngestParameterService{}, mockFileService)
		)

		mockFileService.EXPECT().
			ListFiles(gomock.Any(), "", storage.ListOptions{Recursive: true}).
			Return([]storage.FileInfo{
				{Path: "nested", IsDir: true},
				{Path: "one.json"},
				{Path: "two.json"},
			}, nil)
		mockFileService.EXPECT().
			DeleteFile(gomock.Any(), "one.json").
			Return(nil)
		mockFileService.EXPECT().
			DeleteFile(gomock.Any(), "two.json").
			DoAndReturn(func(context.Context, string) error {
				close(done)
				return nil
			})

		// Act
		ingestControl.DisableIngestFileRetention(response, request)

		// Assert
		require.Equal(t, http.StatusAccepted, response.Code)
		require.Eventually(t, func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		}, time.Second, 10*time.Millisecond)
		require.Eventually(t, func() bool {
			if ingestControl.retainedFileLock.TryLock() {
				ingestControl.retainedFileLock.Unlock()
				return true
			}
			return false
		}, time.Second, 10*time.Millisecond)
	})

	t.Run("parameter error releases lock", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			errSet          = errors.New("set failed")
			request         = httptest.NewRequest(http.MethodPost, "/api/v2/tools/ingest/retained/disable", nil)
			response        = httptest.NewRecorder()
			mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
			ingestControl   = NewIngestControlTool(retainedIngestParameterService{setErr: errSet}, mockFileService)
		)

		// Act
		ingestControl.DisableIngestFileRetention(response, request)

		// Assert
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.True(t, ingestControl.retainedFileLock.TryLock())
		ingestControl.retainedFileLock.Unlock()
	})

	t.Run("list error releases lock", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			errList         = errors.New("list failed")
			request         = httptest.NewRequest(http.MethodPost, "/api/v2/tools/ingest/retained/disable", nil)
			response        = httptest.NewRecorder()
			done            = make(chan struct{})
			mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
			ingestControl   = NewIngestControlTool(retainedIngestParameterService{}, mockFileService)
		)

		mockFileService.EXPECT().
			ListFiles(gomock.Any(), "", storage.ListOptions{Recursive: true}).
			DoAndReturn(func(context.Context, string, storage.ListOptions) ([]storage.FileInfo, error) {
				close(done)
				return nil, errList
			})

		// Act
		ingestControl.DisableIngestFileRetention(response, request)

		// Assert
		require.Equal(t, http.StatusAccepted, response.Code)
		require.Eventually(t, func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		}, time.Second, 10*time.Millisecond)
		require.Eventually(t, func() bool {
			if ingestControl.retainedFileLock.TryLock() {
				ingestControl.retainedFileLock.Unlock()
				return true
			}
			return false
		}, time.Second, 10*time.Millisecond)
	})
}

func TestIngestControl_LockConflict(t *testing.T) {
	t.Parallel()

	t.Run("fetch returns conflict while write lock is held", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			request         = httptest.NewRequest(http.MethodGet, "/api/v2/tools/ingest/retained", nil)
			response        = httptest.NewRecorder()
			mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
			ingestControl   = NewIngestControlTool(retainedIngestParameterService{retain: true}, mockFileService)
		)

		ingestControl.retainedFileLock.Lock()
		defer ingestControl.retainedFileLock.Unlock()

		// Act
		ingestControl.FetchRetainedIngestFiles(response, request)

		// Assert
		require.Equal(t, http.StatusConflict, response.Code)
	})

	t.Run("disable returns conflict while lock is held", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var (
			request         = httptest.NewRequest(http.MethodPost, "/api/v2/tools/ingest/retained/disable", nil)
			response        = httptest.NewRecorder()
			mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
			ingestControl   = NewIngestControlTool(retainedIngestParameterService{}, mockFileService)
		)

		ingestControl.retainedFileLock.Lock()
		defer ingestControl.retainedFileLock.Unlock()

		// Act
		ingestControl.DisableIngestFileRetention(response, request)

		// Assert
		require.Equal(t, http.StatusConflict, response.Code)
	})
}

var _ appcfg.ParameterService = retainedIngestParameterService{}
var _ io.ReadCloser = retainedIngestErrorReadCloser{}
