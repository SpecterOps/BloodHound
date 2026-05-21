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

package graphify_test

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/storage"
	storagemocks "github.com/specterops/bloodhound/cmd/api/src/services/storage/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type ingestStorageTrackingReadCloser struct {
	reader io.Reader
	closed bool
}

func (s *ingestStorageTrackingReadCloser) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *ingestStorageTrackingReadCloser) Close() error {
	s.closed = true
	return nil
}

type ingestStorageErrorReadCloser struct {
	err    error
	closed bool
}

func (s *ingestStorageErrorReadCloser) Read([]byte) (int, error) {
	return 0, s.err
}

func (s *ingestStorageErrorReadCloser) Close() error {
	s.closed = true
	return nil
}

type ingestStorageZipEntry struct {
	name    string
	content string
	isDir   bool
}

func buildIngestStorageZip(t *testing.T, entries ...ingestStorageZipEntry) []byte {
	t.Helper()

	var archive bytes.Buffer
	archiveWriter := zip.NewWriter(&archive)

	for _, entry := range entries {
		if entry.isDir {
			_, err := archiveWriter.Create(entry.name)
			require.NoError(t, err)
			continue
		}

		entryWriter, err := archiveWriter.Create(entry.name)
		require.NoError(t, err)

		_, err = entryWriter.Write([]byte(entry.content))
		require.NoError(t, err)
	}

	require.NoError(t, archiveWriter.Close())
	return archive.Bytes()
}

func openIngestStorageZipFile(t *testing.T, archiveBytes []byte, fileName string) *zip.File {
	t.Helper()

	archiveReader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	require.NoError(t, err)

	for _, archiveFile := range archiveReader.File {
		if archiveFile.Name == fileName {
			return archiveFile
		}
	}

	t.Fatalf("archive file %q not found", fileName)
	return nil
}

func requireScratchDirectoryEmpty(t *testing.T, scratchDirectory string) {
	t.Helper()

	entries, err := os.ReadDir(scratchDirectory)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func requireReadFile(t *testing.T, filePath string) []byte {
	t.Helper()

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	return data
}

func TestSpoolToScratch(t *testing.T) {
	t.Parallel()

	var (
		errGet  = errors.New("get failed")
		errRead = errors.New("read failed")
	)

	type expected struct {
		errIs       error
		errContains string
		content     string
		closed      bool
	}

	type testData struct {
		name       string
		storedName string
		setupMock  func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) io.ReadCloser
		expected   expected
		verify     func(t *testing.T, scratchDirectory string, scratchPath string)
	}

	tests := []testData{
		{
			name:       "copies stored file to scratch",
			storedName: "stored.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) io.ReadCloser {
				readCloser := &ingestStorageTrackingReadCloser{reader: strings.NewReader("stored content")}
				mockFileService.EXPECT().
					GetFile(ctx, "stored.json").
					Return(readCloser, storage.FileInfo{}, nil)

				return readCloser
			},
			expected: expected{
				content: "stored content",
				closed:  true,
			},
		},
		{
			name:       "get error returns wrapped error",
			storedName: "missing.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) io.ReadCloser {
				mockFileService.EXPECT().
					GetFile(ctx, "missing.json").
					Return(nil, storage.FileInfo{}, errGet)

				return nil
			},
			expected: expected{
				errIs:       errGet,
				errContains: `open stored ingest file "missing.json"`,
			},
			verify: func(t *testing.T, scratchDirectory string, scratchPath string) {
				require.Empty(t, scratchPath)
				requireScratchDirectoryEmpty(t, scratchDirectory)
			},
		},
		{
			name:       "copy error removes scratch file",
			storedName: "stored.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) io.ReadCloser {
				readCloser := &ingestStorageErrorReadCloser{err: errRead}
				mockFileService.EXPECT().
					GetFile(ctx, "stored.json").
					Return(readCloser, storage.FileInfo{}, nil)

				return readCloser
			},
			expected: expected{
				errIs:       errRead,
				errContains: `copy stored ingest file "stored.json" to scratch`,
				closed:      true,
			},
			verify: func(t *testing.T, scratchDirectory string, scratchPath string) {
				require.Empty(t, scratchPath)
				requireScratchDirectoryEmpty(t, scratchDirectory)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx              = context.Background()
				scratchDirectory = t.TempDir()
				mockFileService  = storagemocks.NewMockFileService(gomock.NewController(t))
				readCloser       = testCase.setupMock(t, ctx, mockFileService)
			)

			// Act
			scratchPath, err := graphify.SpoolToScratch(ctx, scratchDirectory, mockFileService, testCase.storedName)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Contains(t, err.Error(), testCase.expected.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expected.content, string(requireReadFile(t, scratchPath)))
			}

			if readCloser != nil {
				switch typedReadCloser := readCloser.(type) {
				case *ingestStorageTrackingReadCloser:
					require.Equal(t, testCase.expected.closed, typedReadCloser.closed)
				case *ingestStorageErrorReadCloser:
					require.Equal(t, testCase.expected.closed, typedReadCloser.closed)
				}
			}

			if testCase.verify != nil {
				testCase.verify(t, scratchDirectory, scratchPath)
			}
		})
	}
}

func TestOpenScratchReadSeeker(t *testing.T) {
	t.Parallel()

	var errGet = errors.New("get failed")

	type expected struct {
		errIs   error
		content string
	}

	type testData struct {
		name       string
		storedName string
		setupMock  func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService)
		expected   expected
	}

	tests := []testData{
		{
			name:       "opens copied scratch file",
			storedName: "stored.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					GetFile(ctx, "stored.json").
					Return(io.NopCloser(strings.NewReader("stored content")), storage.FileInfo{}, nil)
			},
			expected: expected{
				content: "stored content",
			},
		},
		{
			name:       "spool error returns error",
			storedName: "missing.json",
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					GetFile(ctx, "missing.json").
					Return(nil, storage.FileInfo{}, errGet)
			},
			expected: expected{
				errIs: errGet,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx              = context.Background()
				scratchDirectory = t.TempDir()
				mockFileService  = storagemocks.NewMockFileService(gomock.NewController(t))
			)

			testCase.setupMock(t, ctx, mockFileService)

			// Act
			scratchFile, scratchPath, err := graphify.OpenScratchReadSeeker(ctx, scratchDirectory, mockFileService, testCase.storedName)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Nil(t, scratchFile)
				require.Empty(t, scratchPath)
				return
			}

			require.NoError(t, err)
			defer os.Remove(scratchPath)
			defer scratchFile.Close()

			content, err := io.ReadAll(scratchFile)
			require.NoError(t, err)
			require.Equal(t, testCase.expected.content, string(content))
		})
	}
}

func TestWriteArchiveFileToStorage(t *testing.T) {
	t.Parallel()

	var errWrite = errors.New("write failed")

	type expected struct {
		errIs       error
		errContains string
		path        string
		content     string
	}

	type testData struct {
		name     string
		content  string
		writeErr error
		expected expected
	}

	tests := []testData{
		{
			name:    "writes normalized archive file content to storage",
			content: string([]byte{0xEF, 0xBB, 0xBF}) + "stored content",
			expected: expected{
				path:    "prefix/tmp-file",
				content: "stored content",
			},
		},
		{
			name:     "write error returns wrapped error",
			content:  "stored content",
			writeErr: errWrite,
			expected: expected{
				errIs:       errWrite,
				errContains: `write archive file "file.json" to storage`,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx             = context.Background()
				archiveBytes    = buildIngestStorageZip(t, ingestStorageZipEntry{name: "file.json", content: testCase.content})
				archiveFile     = openIngestStorageZipFile(t, archiveBytes, "file.json")
				mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
			)

			mockFileService.EXPECT().
				WriteTempFile(ctx, "prefix", gomock.Any(), storage.WriteOptions{}).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
					content, err := io.ReadAll(reader)
					require.NoError(t, err)
					if testCase.expected.content != "" {
						require.Equal(t, testCase.expected.content, string(content))
					}

					return "prefix/tmp-file", testCase.writeErr
				})

			// Act
			extractedPath, err := graphify.WriteArchiveFileToStorage(ctx, mockFileService, archiveFile, "prefix")

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Contains(t, err.Error(), testCase.expected.errContains)
				require.Empty(t, extractedPath)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expected.path, extractedPath)
		})
	}
}

func TestExtractIngestFiles(t *testing.T) {
	t.Parallel()

	var (
		errGet   = errors.New("get failed")
		errWrite = errors.New("write failed")
	)

	type expected struct {
		errIs       error
		errContains string
		fileData    []graphify.IngestFileData
	}

	type testData struct {
		name             string
		fileType         model.FileType
		providedFileName string
		setupMock        func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService)
		expected         expected
		verify           func(t *testing.T, scratchDirectory string)
	}

	tests := []testData{
		{
			name:             "json file returns stored path without storage calls",
			fileType:         model.FileTypeJson,
			providedFileName: "provided.json",
			expected: expected{
				fileData: []graphify.IngestFileData{
					{
						Name: "provided.json",
						Path: "stored.json",
					},
				},
			},
		},
		{
			name:     "zip file extracts files and deletes archive",
			fileType: model.FileTypeZip,
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				archiveBytes := buildIngestStorageZip(t,
					ingestStorageZipEntry{name: "nested/", isDir: true},
					ingestStorageZipEntry{name: "nested/one.json", content: "one"},
					ingestStorageZipEntry{name: "two.json", content: "two"},
				)

				mockFileService.EXPECT().
					GetFile(ctx, "stored.json").
					Return(io.NopCloser(bytes.NewReader(archiveBytes)), storage.FileInfo{}, nil)
				gomock.InOrder(
					mockFileService.EXPECT().
						WriteTempFile(ctx, "prefix", gomock.Any(), storage.WriteOptions{}).
						DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
							content, err := io.ReadAll(reader)
							require.NoError(t, err)
							require.Equal(t, "one", string(content))

							return "prefix/tmp-one", nil
						}),
					mockFileService.EXPECT().
						WriteTempFile(ctx, "prefix", gomock.Any(), storage.WriteOptions{}).
						DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
							content, err := io.ReadAll(reader)
							require.NoError(t, err)
							require.Equal(t, "two", string(content))

							return "prefix/tmp-two", nil
						}),
				)
				mockFileService.EXPECT().
					DeleteFile(ctx, "stored.json").
					Return(nil)
			},
			expected: expected{
				fileData: []graphify.IngestFileData{
					{
						Name:       "nested/one.json",
						ParentFile: "provided.zip",
						Path:       "prefix/tmp-one",
					},
					{
						Name:       "two.json",
						ParentFile: "provided.zip",
						Path:       "prefix/tmp-two",
					},
				},
			},
			verify: func(t *testing.T, scratchDirectory string) {
				requireScratchDirectoryEmpty(t, scratchDirectory)
			},
		},
		{
			name:     "spool error returns file data with error",
			fileType: model.FileTypeZip,
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					GetFile(ctx, "stored.json").
					Return(nil, storage.FileInfo{}, errGet)
			},
			expected: expected{
				errIs: errGet,
				fileData: []graphify.IngestFileData{
					{
						Name:   "provided.zip",
						Path:   "stored.json",
						Errors: []string{`Error spooling archive to scratch: open stored ingest file "stored.json": get failed`},
					},
				},
			},
			verify: func(t *testing.T, scratchDirectory string) {
				requireScratchDirectoryEmpty(t, scratchDirectory)
			},
		},
		{
			name:     "bad zip returns file data with error",
			fileType: model.FileTypeZip,
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				mockFileService.EXPECT().
					GetFile(ctx, "stored.json").
					Return(io.NopCloser(strings.NewReader("not a zip")), storage.FileInfo{}, nil)
			},
			expected: expected{
				errIs: zip.ErrFormat,
				fileData: []graphify.IngestFileData{
					{
						Name:   "provided.zip",
						Path:   "stored.json",
						Errors: []string{"Error opening archive: zip: not a valid zip file"},
					},
				},
			},
			verify: func(t *testing.T, scratchDirectory string) {
				requireScratchDirectoryEmpty(t, scratchDirectory)
			},
		},
		{
			name:     "archive file write error records file error and returns aggregate error",
			fileType: model.FileTypeZip,
			setupMock: func(t *testing.T, ctx context.Context, mockFileService *storagemocks.MockFileService) {
				archiveBytes := buildIngestStorageZip(t,
					ingestStorageZipEntry{name: "one.json", content: "one"},
					ingestStorageZipEntry{name: "two.json", content: "two"},
				)

				mockFileService.EXPECT().
					GetFile(ctx, "stored.json").
					Return(io.NopCloser(bytes.NewReader(archiveBytes)), storage.FileInfo{}, nil)
				gomock.InOrder(
					mockFileService.EXPECT().
						WriteTempFile(ctx, "prefix", gomock.Any(), storage.WriteOptions{}).
						DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
							_, err := io.ReadAll(reader)
							require.NoError(t, err)

							return "", errWrite
						}),
					mockFileService.EXPECT().
						WriteTempFile(ctx, "prefix", gomock.Any(), storage.WriteOptions{}).
						DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
							content, err := io.ReadAll(reader)
							require.NoError(t, err)
							require.Equal(t, "two", string(content))

							return "prefix/tmp-two", nil
						}),
				)
				mockFileService.EXPECT().
					DeleteFile(ctx, "stored.json").
					Return(nil)
			},
			expected: expected{
				errIs:       errWrite,
				errContains: `error extracting file one.json in archive stored.json: write archive file "one.json" to storage: write failed`,
				fileData: []graphify.IngestFileData{
					{
						Name:       "one.json",
						ParentFile: "provided.zip",
						Errors:     []string{`write archive file "one.json" to storage: write failed`},
					},
					{
						Name:       "two.json",
						ParentFile: "provided.zip",
						Path:       "prefix/tmp-two",
					},
				},
			},
			verify: func(t *testing.T, scratchDirectory string) {
				requireScratchDirectoryEmpty(t, scratchDirectory)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx              = context.Background()
				scratchDirectory = t.TempDir()
				mockFileService  = storagemocks.NewMockFileService(gomock.NewController(t))
			)

			if testCase.setupMock != nil {
				testCase.setupMock(t, ctx, mockFileService)
			}
			providedFileName := testCase.providedFileName
			if providedFileName == "" {
				providedFileName = "provided.zip"
			}

			// Act
			fileData, err := graphify.ExtractIngestFiles(ctx, scratchDirectory, mockFileService, "stored.json", providedFileName, testCase.fileType, "prefix")

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				if testCase.expected.errContains != "" {
					require.Contains(t, err.Error(), testCase.expected.errContains)
				}
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.expected.fileData, fileData)

			if testCase.verify != nil {
				testCase.verify(t, scratchDirectory)
			}
		})
	}
}
