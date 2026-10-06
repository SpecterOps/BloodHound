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

package storage_test

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/storage"
	"github.com/specterops/bloodhound/packages/go/storage/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type trackingReadCloser struct {
	reader    io.Reader
	closeErr  error
	closed    bool
	closeCall int
}

func (s *trackingReadCloser) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *trackingReadCloser) Close() error {
	s.closed = true
	s.closeCall++
	return s.closeErr
}

type readErrorSource struct {
	err error
}

type testPendingFile struct {
	commitName string
	commitErr  error
	aborted    bool
}

func (s *testPendingFile) Write(data []byte) (int, error) {
	return len(data), nil
}

func (s *testPendingFile) Commit() (string, error) {
	return s.commitName, s.commitErr
}

func (s *testPendingFile) Abort() error {
	s.aborted = true
	return nil
}

func (s readErrorSource) Read([]byte) (int, error) {
	return 0, s.err
}

func TestNewFileService(t *testing.T) {
	t.Parallel()

	// Arrange
	mockStorage := mocks.NewMockStorage(gomock.NewController(t))

	// Act
	fileService := storage.NewFileService(mockStorage)

	// Assert
	require.Same(t, mockStorage, fileService.Storage)
}

func TestStorageFileServiceBeginFile(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	options := storage.WriteOptions{
		ContentType: "application/json",
		SizeHint:    10,
	}
	pendingFile := &testPendingFile{commitName: "named/file"}
	mockStorage := mocks.NewMockStorage(gomock.NewController(t))
	mockStorage.EXPECT().BeginWrite(ctx, "named/file", options).Return(pendingFile, nil)

	// Act
	actual, err := storage.NewFileService(mockStorage).BeginFile(ctx, "named/file", options)

	// Assert
	require.NoError(t, err)
	require.Same(t, pendingFile, actual)
}

func TestStorageFileServiceBeginTempFileCommitsUniqueNames(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	options := storage.WriteOptions{
		ContentType: "application/zip",
		Metadata:    map[string]string{"source": "upload"},
		SizeHint:    99,
	}
	mockStorage := mocks.NewMockStorage(gomock.NewController(t))
	var names []string
	mockStorage.EXPECT().BeginWrite(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(actualContext context.Context, name string, actualOptions storage.WriteOptions) (storage.PendingFile, error) {
		require.Equal(t, ctx, actualContext)
		require.True(t, strings.HasPrefix(name, "ingest/tmp-"))
		names = append(names, name)
		require.Equal(t, options.ContentType, actualOptions.ContentType)
		require.Equal(t, options.Metadata, actualOptions.Metadata)
		require.Equal(t, options.SizeHint, actualOptions.SizeHint)
		require.True(t, actualOptions.FailIfExists)
		return &testPendingFile{commitName: name}, nil
	}).Times(2)

	pendingFile, err := storage.NewFileService(mockStorage).BeginTempFile(ctx, "ingest/", options)
	require.NoError(t, err)

	// Act
	committedName, err := pendingFile.Commit()

	// Assert
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(committedName, "ingest/tmp-"))

	repeatedName, err := pendingFile.Commit()
	require.NoError(t, err)
	require.Equal(t, committedName, repeatedName)
	_, err = pendingFile.Write([]byte("data"))
	require.ErrorIs(t, err, io.ErrClosedPipe)

	secondFile, err := storage.NewFileService(mockStorage).BeginTempFile(ctx, "ingest/", options)
	require.NoError(t, err)
	secondName, err := secondFile.Commit()
	require.NoError(t, err)
	require.Equal(t, names[0], committedName)
	require.Equal(t, names[1], secondName)
	require.NotEqual(t, committedName, secondName)
}

func TestStorageFileServiceBeginTempFileCanAbort(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	pendingFile := &testPendingFile{}
	mockStorage := mocks.NewMockStorage(gomock.NewController(t))
	mockStorage.EXPECT().BeginWrite(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, name string, options storage.WriteOptions) (storage.PendingFile, error) {
		require.True(t, strings.HasPrefix(name, "ingest/tmp-"))
		require.True(t, options.FailIfExists)
		return pendingFile, nil
	})

	// Act
	actual, err := storage.NewFileService(mockStorage).BeginTempFile(ctx, "ingest/", storage.WriteOptions{})
	require.NoError(t, err)
	_, err = actual.Write([]byte("data"))
	require.NoError(t, err)
	err = actual.Abort()

	// Assert
	require.NoError(t, err)
	require.True(t, pendingFile.aborted)
}

func TestStorageFileServiceBeginTempFileCleansOwnedNameAfterUncertainCommit(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	commitErr := errors.New("completion response lost")
	mockStorage := mocks.NewMockStorage(gomock.NewController(t))
	var tempName string
	mockStorage.EXPECT().BeginWrite(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, name string, _ storage.WriteOptions) (storage.PendingFile, error) {
		tempName = name
		return &testPendingFile{commitName: name, commitErr: commitErr}, nil
	})
	mockStorage.EXPECT().Delete(gomock.Any(), gomock.Any()).DoAndReturn(func(cleanupContext context.Context, name string) error {
		require.Equal(t, tempName, name)
		require.NoError(t, cleanupContext.Err())
		_, hasDeadline := cleanupContext.Deadline()
		require.True(t, hasDeadline)
		return nil
	})

	pendingFile, err := storage.NewFileService(mockStorage).BeginTempFile(ctx, "owned/", storage.WriteOptions{})
	require.NoError(t, err)
	cancel()

	// Act
	_, err = pendingFile.Commit()

	// Assert
	require.ErrorIs(t, err, commitErr)
	require.NotErrorIs(t, err, fs.ErrExist)
}

func TestStorageFileService_GetFile(t *testing.T) {
	t.Parallel()

	var (
		errGet     = errors.New("get failed")
		fileInfo   = storage.FileInfo{Path: "file.json", Size: 4, ContentType: "application/json"}
		readCloser = &trackingReadCloser{reader: strings.NewReader("data")}
	)

	type expected struct {
		errIs      error
		fileInfo   storage.FileInfo
		readCloser io.ReadCloser
	}

	type testData struct {
		name     string
		getErr   error
		expected expected
	}

	tests := []testData{
		{
			name: "gets file from storage",
			expected: expected{
				fileInfo:   fileInfo,
				readCloser: readCloser,
			},
		},
		{
			name:   "get error returns error",
			getErr: errGet,
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
				ctx         = context.Background()
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
			)

			mockStorage.EXPECT().
				Get(ctx, "file.json").
				Return(testCase.expected.readCloser, testCase.expected.fileInfo, testCase.getErr)

			// Act
			actualReadCloser, actualFileInfo, err := fileService.GetFile(ctx, "file.json")

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Nil(t, actualReadCloser)
				require.Empty(t, actualFileInfo)
				return
			}

			require.NoError(t, err)
			require.Same(t, testCase.expected.readCloser, actualReadCloser)
			require.Equal(t, testCase.expected.fileInfo, actualFileInfo)
		})
	}
}

func TestStorageFileService_GetPresignedURL(t *testing.T) {
	t.Parallel()

	var (
		ctx         = context.Background()
		mockStorage = mocks.NewMockStorage(gomock.NewController(t))
		fileService = storage.NewFileService(mockStorage)
		ttl         = time.Minute
		expectedURL = "https://storage.example/file.json"
	)

	mockStorage.EXPECT().
		GetPresignedURL(ctx, "file.json", ttl).
		Return(expectedURL, nil)

	actualURL, err := fileService.GetPresignedURL(ctx, "file.json", ttl)

	require.NoError(t, err)
	require.Equal(t, expectedURL, actualURL)
}

func TestStorageFileService_ReadFile(t *testing.T) {
	t.Parallel()

	var (
		errGet  = errors.New("get failed")
		errRead = errors.New("read failed")
	)

	type expected struct {
		errIs   error
		content string
		closed  bool
	}

	type testData struct {
		name      string
		setupMock func(t *testing.T, ctx context.Context, mockStorage *mocks.MockStorage) *trackingReadCloser
		expected  expected
	}

	tests := []testData{
		{
			name: "reads file content and closes reader",
			setupMock: func(t *testing.T, ctx context.Context, mockStorage *mocks.MockStorage) *trackingReadCloser {
				readCloser := &trackingReadCloser{reader: strings.NewReader("content")}
				mockStorage.EXPECT().
					Get(ctx, "file.json").
					Return(readCloser, storage.FileInfo{Path: "file.json"}, nil)

				return readCloser
			},
			expected: expected{
				content: "content",
				closed:  true,
			},
		},
		{
			name: "get error returns error",
			setupMock: func(t *testing.T, ctx context.Context, mockStorage *mocks.MockStorage) *trackingReadCloser {
				mockStorage.EXPECT().
					Get(ctx, "file.json").
					Return(nil, storage.FileInfo{}, errGet)

				return nil
			},
			expected: expected{
				errIs: errGet,
			},
		},
		{
			name: "reader error returns error and closes reader",
			setupMock: func(t *testing.T, ctx context.Context, mockStorage *mocks.MockStorage) *trackingReadCloser {
				readCloser := &trackingReadCloser{reader: readErrorSource{err: errRead}}
				mockStorage.EXPECT().
					Get(ctx, "file.json").
					Return(readCloser, storage.FileInfo{Path: "file.json"}, nil)

				return readCloser
			},
			expected: expected{
				errIs:  errRead,
				closed: true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx         = context.Background()
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
				readCloser  = testCase.setupMock(t, ctx, mockStorage)
			)

			// Act
			content, err := fileService.ReadFile(ctx, "file.json")

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.expected.content, string(content))
			if readCloser != nil {
				require.Equal(t, testCase.expected.closed, readCloser.closed)
			}
		})
	}
}

func TestStorageFileService_WriteFile(t *testing.T) {
	t.Parallel()

	var errPut = errors.New("put failed")

	type expected struct {
		errIs error
	}

	type testData struct {
		name     string
		putErr   error
		expected expected
	}

	tests := []testData{
		{
			name: "writes file content",
		},
		{
			name:   "put error returns error",
			putErr: errPut,
			expected: expected{
				errIs: errPut,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx         = context.Background()
				options     = storage.WriteOptions{ContentType: "application/json"}
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
			)

			mockStorage.EXPECT().
				Put(ctx, "file.json", gomock.Any(), options).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) error {
					content, err := io.ReadAll(reader)
					require.NoError(t, err)
					require.Equal(t, `{"ok":true}`, string(content))

					return testCase.putErr
				})

			// Act
			err := fileService.WriteFile(ctx, "file.json", []byte(`{"ok":true}`), options)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorageFileService_WriteFileFromReader(t *testing.T) {
	t.Parallel()

	var errPut = errors.New("put failed")

	type expected struct {
		errIs error
	}

	type testData struct {
		name     string
		putErr   error
		expected expected
	}

	tests := []testData{
		{
			name: "writes reader",
		},
		{
			name:   "put error returns error",
			putErr: errPut,
			expected: expected{
				errIs: errPut,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx         = context.Background()
				options     = storage.WriteOptions{FailIfExists: true}
				reader      = strings.NewReader("content")
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
			)

			mockStorage.EXPECT().
				Put(ctx, "file.json", reader, options).
				Return(testCase.putErr)

			// Act
			err := fileService.WriteFileFromReader(ctx, "file.json", reader, options)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorageFileService_DeleteFile(t *testing.T) {
	t.Parallel()

	var errDelete = errors.New("delete failed")

	type expected struct {
		errIs error
	}

	type testData struct {
		name      string
		deleteErr error
		expected  expected
	}

	tests := []testData{
		{
			name: "deletes file",
		},
		{
			name:      "delete error returns error",
			deleteErr: errDelete,
			expected: expected{
				errIs: errDelete,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx         = context.Background()
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
			)

			mockStorage.EXPECT().
				Delete(ctx, "file.json").
				Return(testCase.deleteErr)

			// Act
			err := fileService.DeleteFile(ctx, "file.json")

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorageFileService_DeleteFileWithOptions(t *testing.T) {
	t.Parallel()

	var (
		errDelete = errors.New("delete failed")
		errPrune  = errors.New("prune failed")
	)

	type expected struct {
		errIs error
	}

	type testData struct {
		name              string
		options           storage.DeleteOptions
		deleteErr         error
		pruneEmptyParents bool
		pruneErr          error
		expected          expected
	}

	tests := []testData{
		{
			name: "deletes file without pruning when disabled",
		},
		{
			name:              "deletes file and prunes empty parents when enabled",
			options:           storage.DeleteOptions{PruneEmptyParents: true},
			pruneEmptyParents: true,
		},
		{
			name:      "returns delete error without pruning",
			options:   storage.DeleteOptions{PruneEmptyParents: true},
			deleteErr: errDelete,
			expected: expected{
				errIs: errDelete,
			},
		},
		{
			name:              "returns pruning error",
			options:           storage.DeleteOptions{PruneEmptyParents: true},
			pruneEmptyParents: true,
			pruneErr:          errPrune,
			expected: expected{
				errIs: errPrune,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctx         = context.Background()
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
			)

			deleteCall := mockStorage.EXPECT().
				Delete(ctx, "file.json").
				Return(testCase.deleteErr)

			if testCase.pruneEmptyParents {
				pruneCall := mockStorage.EXPECT().
					PruneEmptyParents(ctx, "file.json").
					Return(testCase.pruneErr)
				gomock.InOrder(deleteCall, pruneCall)
			}

			err := fileService.DeleteFileWithOptions(ctx, "file.json", testCase.options)

			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorageFileService_WriteTempFile(t *testing.T) {
	t.Parallel()

	var errPut = errors.New("put failed")

	type expected struct {
		errIs error
	}

	type testData struct {
		name     string
		putErr   error
		expected expected
	}

	tests := []testData{
		{
			name: "writes temp file and returns temp path",
		},
		{
			name:   "put error returns error",
			putErr: errPut,
			expected: expected{
				errIs: errPut,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx         = context.Background()
				options     = storage.WriteOptions{ContentType: "text/plain"}
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
				writtenPath string
			)

			mockStorage.EXPECT().
				Put(ctx, gomock.Any(), gomock.Any(), options).
				DoAndReturn(func(_ context.Context, name string, reader io.Reader, _ storage.WriteOptions) error {
					content, err := io.ReadAll(reader)
					require.NoError(t, err)
					require.Equal(t, "content", string(content))
					require.Regexp(t, `^prefix_tmp-[0-9a-f]{32}$`, name)

					writtenPath = name
					return testCase.putErr
				})

			// Act
			tempPath, err := fileService.WriteTempFile(ctx, "prefix_", strings.NewReader("content"), options)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Empty(t, tempPath)
			} else {
				require.NoError(t, err)
				require.Equal(t, writtenPath, tempPath)
			}
		})
	}
}

func TestStorageFileService_MoveFile(t *testing.T) {
	t.Parallel()

	var errMove = errors.New("move failed")

	type expected struct {
		errIs error
	}

	type testData struct {
		name     string
		moveErr  error
		expected expected
	}

	tests := []testData{
		{
			name: "moves file",
		},
		{
			name:    "move error returns error",
			moveErr: errMove,
			expected: expected{
				errIs: errMove,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx         = context.Background()
				options     = storage.WriteOptions{FailIfExists: true}
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
			)

			mockStorage.EXPECT().
				Move(ctx, "source.json", "destination.json", options).
				Return(testCase.moveErr)

			// Act
			err := fileService.MoveFile(ctx, "source.json", "destination.json", options)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorageFileService_ListFiles(t *testing.T) {
	t.Parallel()

	var (
		errList   = errors.New("list failed")
		fileInfos = []storage.FileInfo{{Path: "file.json"}}
	)

	type expected struct {
		errIs     error
		fileInfos []storage.FileInfo
	}

	type testData struct {
		name     string
		listErr  error
		expected expected
	}

	tests := []testData{
		{
			name: "lists files",
			expected: expected{
				fileInfos: fileInfos,
			},
		},
		{
			name:    "list error returns error",
			listErr: errList,
			expected: expected{
				errIs: errList,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx         = context.Background()
				options     = storage.ListOptions{Recursive: true, Limit: 10}
				mockStorage = mocks.NewMockStorage(gomock.NewController(t))
				fileService = storage.NewFileService(mockStorage)
			)

			mockStorage.EXPECT().
				List(ctx, "prefix", options).
				Return(testCase.expected.fileInfos, testCase.listErr)

			// Act
			actualFileInfos, err := fileService.ListFiles(ctx, "prefix", options)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Nil(t, actualFileInfos)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expected.fileInfos, actualFileInfos)
		})
	}
}

func TestMoveFileBetweenServices(t *testing.T) {
	t.Parallel()

	var (
		errGet    = errors.New("get failed")
		errWrite  = errors.New("write failed")
		errDelete = errors.New("delete failed")
		errClose  = errors.New("close failed")
	)

	type expected struct {
		errIs           error
		additionalErrIs error
		closed          bool
	}

	type testData struct {
		name      string
		setupMock func(
			t *testing.T,
			ctx context.Context,
			sourceService *mocks.MockFileService,
			destinationService *mocks.MockFileService,
			options storage.WriteOptions,
		) *trackingReadCloser
		expected expected
	}

	tests := []testData{
		{
			name: "moves file between services",
			setupMock: func(
				t *testing.T,
				ctx context.Context,
				sourceService *mocks.MockFileService,
				destinationService *mocks.MockFileService,
				options storage.WriteOptions,
			) *trackingReadCloser {
				readCloser := &trackingReadCloser{reader: strings.NewReader("content")}

				gomock.InOrder(
					sourceService.EXPECT().
						GetFile(ctx, "source.json").
						Return(readCloser, storage.FileInfo{Path: "source.json"}, nil),
					destinationService.EXPECT().
						WriteFileFromReader(ctx, "destination.json", readCloser, options).
						DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) error {
							content, err := io.ReadAll(reader)
							require.NoError(t, err)
							require.Equal(t, "content", string(content))

							return nil
						}),
					sourceService.EXPECT().
						DeleteFile(ctx, "source.json").
						DoAndReturn(func(_ context.Context, _ string) error {
							require.True(t, readCloser.closed)
							require.Equal(t, 1, readCloser.closeCall)
							return nil
						}),
				)

				return readCloser
			},
			expected: expected{
				closed: true,
			},
		},
		{
			name: "get error returns error",
			setupMock: func(
				t *testing.T,
				ctx context.Context,
				sourceService *mocks.MockFileService,
				destinationService *mocks.MockFileService,
				options storage.WriteOptions,
			) *trackingReadCloser {
				sourceService.EXPECT().
					GetFile(ctx, "source.json").
					Return(nil, storage.FileInfo{}, errGet)

				return nil
			},
			expected: expected{
				errIs: errGet,
			},
		},
		{
			name: "write error returns error and does not delete source",
			setupMock: func(
				t *testing.T,
				ctx context.Context,
				sourceService *mocks.MockFileService,
				destinationService *mocks.MockFileService,
				options storage.WriteOptions,
			) *trackingReadCloser {
				readCloser := &trackingReadCloser{reader: strings.NewReader("content")}

				gomock.InOrder(
					sourceService.EXPECT().
						GetFile(ctx, "source.json").
						Return(readCloser, storage.FileInfo{Path: "source.json"}, nil),
					destinationService.EXPECT().
						WriteFileFromReader(ctx, "destination.json", readCloser, options).
						Return(errWrite),
				)

				return readCloser
			},
			expected: expected{
				errIs:  errWrite,
				closed: true,
			},
		},
		{
			name: "write and close errors returns both errors and does not delete source",
			setupMock: func(
				t *testing.T,
				ctx context.Context,
				sourceService *mocks.MockFileService,
				destinationService *mocks.MockFileService,
				options storage.WriteOptions,
			) *trackingReadCloser {
				readCloser := &trackingReadCloser{
					reader:   strings.NewReader("content"),
					closeErr: errClose,
				}

				gomock.InOrder(
					sourceService.EXPECT().
						GetFile(ctx, "source.json").
						Return(readCloser, storage.FileInfo{Path: "source.json"}, nil),
					destinationService.EXPECT().
						WriteFileFromReader(ctx, "destination.json", readCloser, options).
						Return(errWrite),
				)

				return readCloser
			},
			expected: expected{
				errIs:           errWrite,
				additionalErrIs: errClose,
				closed:          true,
			},
		},
		{
			name: "delete error returns error",
			setupMock: func(
				t *testing.T,
				ctx context.Context,
				sourceService *mocks.MockFileService,
				destinationService *mocks.MockFileService,
				options storage.WriteOptions,
			) *trackingReadCloser {
				readCloser := &trackingReadCloser{reader: strings.NewReader("content")}

				gomock.InOrder(
					sourceService.EXPECT().
						GetFile(ctx, "source.json").
						Return(readCloser, storage.FileInfo{Path: "source.json"}, nil),
					destinationService.EXPECT().
						WriteFileFromReader(ctx, "destination.json", readCloser, options).
						Return(nil),
					sourceService.EXPECT().
						DeleteFile(ctx, "source.json").
						Return(errDelete),
				)

				return readCloser
			},
			expected: expected{
				errIs:  errDelete,
				closed: true,
			},
		},
		{
			name: "close error returns error and does not delete source",
			setupMock: func(
				t *testing.T,
				ctx context.Context,
				sourceService *mocks.MockFileService,
				destinationService *mocks.MockFileService,
				options storage.WriteOptions,
			) *trackingReadCloser {
				readCloser := &trackingReadCloser{
					reader:   strings.NewReader("content"),
					closeErr: errClose,
				}

				gomock.InOrder(
					sourceService.EXPECT().
						GetFile(ctx, "source.json").
						Return(readCloser, storage.FileInfo{Path: "source.json"}, nil),
					destinationService.EXPECT().
						WriteFileFromReader(ctx, "destination.json", readCloser, options).
						Return(nil),
				)

				return readCloser
			},
			expected: expected{
				errIs:  errClose,
				closed: true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx                = context.Background()
				options            = storage.WriteOptions{FailIfExists: true}
				controller         = gomock.NewController(t)
				sourceService      = mocks.NewMockFileService(controller)
				destinationService = mocks.NewMockFileService(controller)
				readCloser         = testCase.setupMock(t, ctx, sourceService, destinationService, options)
			)

			// Act
			err := storage.MoveFileBetweenServices(ctx, sourceService, destinationService, "source.json", "destination.json", options)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
			} else {
				require.NoError(t, err)
			}
			if testCase.expected.additionalErrIs != nil {
				require.ErrorIs(t, err, testCase.expected.additionalErrIs)
			}

			if readCloser != nil {
				require.Equal(t, testCase.expected.closed, readCloser.closed)
			}
		})
	}
}
