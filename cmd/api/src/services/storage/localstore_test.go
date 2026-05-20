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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/stretchr/testify/require"
)

// errorReader is a type to simulate a failed reader
type errorReader struct {
	err error
}

func (s errorReader) Read([]byte) (int, error) {
	return 0, s.err
}

// partialErrorReader creates the error case that some bytes were written and then failed.
type partialErrorReader struct {
	err  error
	done bool
}

func (s *partialErrorReader) Read(p []byte) (int, error) {
	if s.done {
		return 0, s.err
	}

	s.done = true
	return copy(p, "partial"), nil
}

// requireNoTempFiles is used to walk down directories and ensure no temp files exist.
func requireNoTempFiles(t *testing.T, rootPath string) {
	t.Helper()

	err := filepath.WalkDir(rootPath, func(filePath string, entry fs.DirEntry, walkErr error) error {
		require.NoError(t, walkErr)

		if entry.IsDir() {
			return nil
		}

		require.False(t, strings.HasPrefix(entry.Name(), ".tmp-"), "unexpected temp file: %s", filePath)
		return nil
	})
	require.NoError(t, err)
}

func newTestLocalStore(t *testing.T) (string, *storage.LocalStore) {
	t.Helper()

	rootPath := t.TempDir()
	localStore, err := storage.NewLocalStore(rootPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, localStore.Close())
	})

	return rootPath, localStore
}

func writeTestFile(t *testing.T, rootPath string, name string, data string) {
	t.Helper()

	filePath := filepath.Join(rootPath, filepath.FromSlash(name))
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o750))
	require.NoError(t, os.WriteFile(filePath, []byte(data), 0o640))
}

func readTestFile(t *testing.T, rootPath string, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(name)))
	require.NoError(t, err)
	return string(data)
}

func requireReadFile(t *testing.T, filePath string) []byte {
	t.Helper()

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	return data
}

func fileInfoPaths(fileInfos []storage.FileInfo) []string {
	paths := make([]string, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		paths = append(paths, fileInfo.Path)
	}
	return paths
}

func TestNewLocalStore(t *testing.T) {
	t.Parallel()

	type testData struct {
		name      string
		setupRoot func(t *testing.T) string
		wantErr   bool
	}

	tests := []testData{
		{
			name: "opens existing directory",
			setupRoot: func(t *testing.T) string {
				return t.TempDir()
			},
		},
		{
			name: "missing root returns error",
			setupRoot: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing")
			},
			wantErr: true,
		},
		{
			name: "file root returns error",
			setupRoot: func(t *testing.T) string {
				filePath := filepath.Join(t.TempDir(), "file")
				require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o640))
				return filePath
			},
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath := testCase.setupRoot(t)

			// Act
			localStore, err := storage.NewLocalStore(rootPath)

			// Assert
			if testCase.wantErr {
				require.Error(t, err)
				require.Nil(t, localStore)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, localStore)
			require.NoError(t, localStore.Close())
		})
	}
}

func TestLocalStore_Close(t *testing.T) {
	t.Parallel()

	// Arrange
	rootPath := t.TempDir()
	localStore, err := storage.NewLocalStore(rootPath)
	require.NoError(t, err)

	// Act / Assert
	require.NoError(t, localStore.Close())
	require.NoError(t, localStore.Close())
}

func TestLocalStore_Put(t *testing.T) {
	t.Parallel()

	var errRead = errors.New("read failed")

	type expected struct {
		errIs       error
		errContains string
		fileContent map[string]string
		missing     []string
	}

	type testData struct {
		name         string
		setup        func(t *testing.T, rootPath string)
		buildContext func(t *testing.T) context.Context
		buildReader  func(t *testing.T) io.Reader
		fileName     string
		options      storage.WriteOptions
		expected     expected
		verify       func(t *testing.T, rootPath string)
	}

	tests := []testData{
		{
			name:     "writes new file",
			fileName: "file.json",
			buildReader: func(t *testing.T) io.Reader {
				return strings.NewReader(`{"ok":true}`)
			},
			expected: expected{
				fileContent: map[string]string{"file.json": `{"ok":true}`},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "writes new nested file",
			fileName: "nested/file.json",
			buildReader: func(t *testing.T) io.Reader {
				return strings.NewReader(`{"ok":true}`)
			},
			expected: expected{
				fileContent: map[string]string{"nested/file.json": `{"ok":true}`},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "empty file name returns error",
			fileName: "",
			buildReader: func(t *testing.T) io.Reader {
				return strings.NewReader("data")
			},
			expected: expected{
				errContains: "empty path",
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "overwrites existing file by default",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", "old")
			},
			buildReader: func(t *testing.T) io.Reader {
				return strings.NewReader("new")
			},
			expected: expected{
				fileContent: map[string]string{"nested/file.json": "new"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "fail if exists preserves existing file",
			fileName: "nested/file.json",
			options:  storage.WriteOptions{FailIfExists: true},
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", "old")
			},
			buildReader: func(t *testing.T) io.Reader {
				return strings.NewReader("new")
			},
			expected: expected{
				errIs:       fs.ErrExist,
				fileContent: map[string]string{"nested/file.json": "old"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "canceled context does not save file",
			fileName: "file.json",
			buildReader: func(t *testing.T) io.Reader {
				return strings.NewReader("canceled")
			},
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				errIs:   context.Canceled,
				missing: []string{"file.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "failed reader returns error and removes temp file",
			fileName: "file.json",
			buildReader: func(t *testing.T) io.Reader {
				return errorReader{err: errRead}
			},
			expected: expected{
				errIs:   errRead,
				missing: []string{"file.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "partial reader error does not publish partial file",
			fileName: "file.json",
			buildReader: func(t *testing.T) io.Reader {
				return &partialErrorReader{err: errRead}
			},
			expected: expected{
				errIs:   errRead,
				missing: []string{"file.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			var reader io.Reader = strings.NewReader("")
			if testCase.buildReader != nil {
				reader = testCase.buildReader(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			err := localStore.Put(ctx, testCase.fileName, reader, testCase.options)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
			default:
				require.NoError(t, err)
			}

			for fileName, expectedContent := range testCase.expected.fileContent {
				actualContent, err := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(fileName)))
				require.NoError(t, err)
				require.Equal(t, expectedContent, string(actualContent))
			}

			for _, missingFile := range testCase.expected.missing {
				_, err := os.Stat(filepath.Join(rootPath, filepath.FromSlash(missingFile)))
				require.ErrorIs(t, err, os.ErrNotExist)
			}

			if testCase.verify != nil {
				testCase.verify(t, rootPath)
			}
		})
	}
}

func TestLocalStore_Get(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs       error
		errContains string
		content     string
	}

	type testData struct {
		name         string
		setup        func(t *testing.T, rootPath string)
		buildContext func(t *testing.T) context.Context
		fileName     string
		expected     expected
		verify       func(t *testing.T, fileInfo storage.FileInfo)
	}

	tests := []testData{
		{
			name:     "gets file content and metadata",
			fileName: "file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "file.json", `{"ok":true}`)
			},
			expected: expected{
				content: `{"ok":true}`,
			},
			verify: func(t *testing.T, fileInfo storage.FileInfo) {
				require.Equal(t, "file.json", fileInfo.Path)
				require.Equal(t, int64(len(`{"ok":true}`)), fileInfo.Size)
				require.Equal(t, "application/json", fileInfo.ContentType)
				require.False(t, fileInfo.IsDir)
				require.False(t, fileInfo.LastModified.IsZero())
			},
		},
		{
			name:     "gets nested file content and metadata",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				content: `{"ok":true}`,
			},
			verify: func(t *testing.T, fileInfo storage.FileInfo) {
				require.Equal(t, "nested/file.json", fileInfo.Path)
				require.Equal(t, int64(len(`{"ok":true}`)), fileInfo.Size)
				require.Equal(t, "application/json", fileInfo.ContentType)
				require.False(t, fileInfo.IsDir)
				require.False(t, fileInfo.LastModified.IsZero())
			},
		},
		{
			name:     "missing file returns error",
			fileName: "file.json",
			expected: expected{
				errIs: os.ErrNotExist,
			},
		},
		{
			name:     "get directory returns error",
			fileName: "nested",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				errIs: storage.ErrIsDirectory,
			},
		},
		{
			name:     "canceled context returns canceled",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				errIs: context.Canceled,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			readCloser, fileInfo, err := localStore.Get(ctx, testCase.fileName)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Nil(t, readCloser)
				return
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
				require.Nil(t, readCloser)
				return
			default:
				require.NoError(t, err)
				require.NotNil(t, readCloser)
			}

			content, err := io.ReadAll(readCloser)
			require.NoError(t, err)
			require.NoError(t, readCloser.Close())

			require.Equal(t, testCase.expected.content, string(content))
			if testCase.verify != nil {
				testCase.verify(t, fileInfo)
			}
		})
	}
}

func TestLocalStore_Stat(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs       error
		errContains string
		fileInfo    storage.FileInfo
	}

	type testData struct {
		name         string
		setup        func(t *testing.T, rootPath string)
		buildContext func(t *testing.T) context.Context
		fileName     string
		expected     expected
	}

	tests := []testData{
		{
			name:     "gets file metadata",
			fileName: "file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "file.json", `{"ok":true}`)
			},
			expected: expected{
				fileInfo: storage.FileInfo{
					Path:        "file.json",
					Size:        int64(len(`{"ok":true}`)),
					ContentType: "application/json",
					IsDir:       false,
				},
			},
		},
		{
			name:     "gets nested file metadata",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				fileInfo: storage.FileInfo{
					Path:        "nested/file.json",
					Size:        int64(len(`{"ok":true}`)),
					ContentType: "application/json",
					IsDir:       false,
				},
			},
		},
		{
			name:     "missing file returns error",
			fileName: "file.json",
			expected: expected{
				errIs: os.ErrNotExist,
			},
		},
		{
			name:     "get directory returns error",
			fileName: "nested",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				errIs: storage.ErrIsDirectory,
			},
		},
		{
			name:     "canceled context returns canceled",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				errIs: context.Canceled,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			fileInfo, err := localStore.Stat(ctx, testCase.fileName)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
				return
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
				return
			default:
				require.NoError(t, err)
			}

			require.Equal(t, testCase.expected.fileInfo.Path, fileInfo.Path)
			require.Equal(t, testCase.expected.fileInfo.Size, fileInfo.Size)
			require.Equal(t, testCase.expected.fileInfo.ContentType, fileInfo.ContentType)
			require.False(t, fileInfo.IsDir)
			require.False(t, fileInfo.LastModified.IsZero())
		})
	}
}

func TestLocalStore_Exists(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs       error
		errContains string
		exists      bool
	}

	type testData struct {
		name         string
		setup        func(t *testing.T, rootPath string)
		buildContext func(t *testing.T) context.Context
		fileName     string
		expected     expected
	}

	tests := []testData{
		{
			name:     "exists file returns exists",
			fileName: "file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "file.json", `{"ok":true}`)
			},
			expected: expected{
				exists: true,
			},
		},
		{
			name:     "nested file returns exists",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				exists: true,
			},
		},
		{
			name:     "missing file returns false",
			fileName: "file.json",
			expected: expected{
				exists: false,
			},
		},
		{
			name:     "exists on directory returns false",
			fileName: "nested",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				exists: false,
			},
		},
		{
			name:     "canceled context returns canceled",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				errIs: context.Canceled,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			exists, err := localStore.Exists(ctx, testCase.fileName)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
				return
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
				return
			default:
				require.NoError(t, err)
			}

			require.Equal(t, testCase.expected.exists, exists)
		})
	}
}

func TestLocalStore_Delete(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs        error
		errContains  string
		checkRemoved bool
		shouldRemove bool
	}

	type testData struct {
		name         string
		setup        func(t *testing.T, rootPath string)
		buildContext func(t *testing.T) context.Context
		fileName     string
		expected     expected
		verify       func(t *testing.T, rootPath string)
	}

	tests := []testData{
		{
			name:     "delete file removes file",
			fileName: "file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "file.json", `{"ok":true}`)
			},
			expected: expected{
				checkRemoved: true,
				shouldRemove: true,
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "delete nested file removes file",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				checkRemoved: true,
				shouldRemove: true,
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:     "delete nested file does not removes directory with items in it",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
				writeTestFile(t, rootPath, "nested/file2.json", `{"ok":true}`)
			},
			expected: expected{
				checkRemoved: true,
				shouldRemove: true,
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
				stat, err := os.Stat(filepath.Join(rootPath, "nested"))
				require.Nil(t, err)
				require.True(t, stat.IsDir())
			},
		},
		{
			name:     "missing file does not return error",
			fileName: "file.json",
		},
		{
			name:     "delete directory returns error on directory with items",
			fileName: "nested",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				errIs: storage.ErrIsDirectory,
			},
		},
		{
			name:     "delete directory returns error on empty directory",
			fileName: "nested",
			setup: func(t *testing.T, rootPath string) {
				require.NoError(t, os.MkdirAll(filepath.Join(rootPath, filepath.FromSlash("nested")), 0o750))
			},
			expected: expected{
				errIs: storage.ErrIsDirectory,
			},
		},
		{
			name:     "canceled context returns canceled",
			fileName: "nested/file.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				checkRemoved: true,
				shouldRemove: false,
				errIs:        context.Canceled,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			err := localStore.Delete(ctx, testCase.fileName)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
			default:
				require.NoError(t, err)
			}

			if testCase.expected.checkRemoved {
				_, err := os.Stat(filepath.Join(rootPath, filepath.FromSlash(testCase.fileName)))
				if testCase.expected.shouldRemove {
					require.NotNil(t, err)
					require.ErrorIs(t, err, os.ErrNotExist)
				} else {
					require.Nil(t, err)
				}
			}

			if testCase.verify != nil {
				testCase.verify(t, rootPath)
			}
		})
	}
}

func TestLocalStore_List(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs       error
		errContains string
		paths       []string
	}

	type testData struct {
		name         string
		setup        func(t *testing.T, rootPath string)
		buildContext func(t *testing.T) context.Context
		listName     string
		options      storage.ListOptions
		expected     expected
		verify       func(t *testing.T, fileInfos []storage.FileInfo)
	}

	setupFiles := func(t *testing.T, rootPath string) {
		t.Helper()

		writeTestFile(t, rootPath, "file.json", `{"ok":true}`)
		writeTestFile(t, rootPath, "nested/file.json", `{"nested":true}`)
		writeTestFile(t, rootPath, "nested/file.txt", "nested text")
		writeTestFile(t, rootPath, "nested/deeper/file.json", `{"deeper":true}`)
	}

	tests := []testData{
		{
			name:     "empty path lists root files only",
			listName: "",
			setup:    setupFiles,
			expected: expected{
				paths: []string{"file.json"},
			},
			verify: func(t *testing.T, fileInfos []storage.FileInfo) {
				require.Len(t, fileInfos, 1)
				require.Equal(t, "application/json", fileInfos[0].ContentType)
				require.Equal(t, int64(len(`{"ok":true}`)), fileInfos[0].Size)
				require.False(t, fileInfos[0].IsDir)
				require.False(t, fileInfos[0].LastModified.IsZero())
			},
		},
		{
			name:     "slash lists root files only",
			listName: "/",
			setup:    setupFiles,
			expected: expected{
				paths: []string{"file.json"},
			},
		},
		{
			name:     "lists subdirectory files only",
			listName: "nested",
			setup:    setupFiles,
			expected: expected{
				paths: []string{
					"nested/file.json",
					"nested/file.txt",
				},
			},
		},
		{
			name:     "recursively lists root files",
			listName: "",
			options:  storage.ListOptions{Recursive: true},
			setup:    setupFiles,
			expected: expected{
				paths: []string{
					"file.json",
					"nested/deeper/file.json",
					"nested/file.json",
					"nested/file.txt",
				},
			},
		},
		{
			name:     "recursively lists subdirectory files",
			listName: "nested",
			options:  storage.ListOptions{Recursive: true},
			setup:    setupFiles,
			expected: expected{
				paths: []string{
					"nested/deeper/file.json",
					"nested/file.json",
					"nested/file.txt",
				},
			},
		},
		{
			name:     "limit stops list after requested number of files",
			listName: "",
			options:  storage.ListOptions{Recursive: true, Limit: 2},
			setup:    setupFiles,
			verify: func(t *testing.T, fileInfos []storage.FileInfo) {
				require.Len(t, fileInfos, 2)
				require.Subset(t, []string{
					"file.json",
					"nested/deeper/file.json",
					"nested/file.json",
					"nested/file.txt",
				}, fileInfoPaths(fileInfos))
			},
		},
		{
			name:     "limit stops non-recursive list after requested number of files",
			listName: "nested",
			options:  storage.ListOptions{Limit: 1},
			setup:    setupFiles,
			verify: func(t *testing.T, fileInfos []storage.FileInfo) {
				require.Len(t, fileInfos, 1)
				require.Subset(t, []string{
					"nested/file.json",
					"nested/file.txt",
				}, fileInfoPaths(fileInfos))
			},
		},
		{
			name:     "missing path returns empty list",
			listName: "missing",
			setup:    setupFiles,
			expected: expected{
				paths: []string{},
			},
		},
		{
			name:     "missing recursive path returns empty list",
			listName: "missing",
			options:  storage.ListOptions{Recursive: true},
			setup:    setupFiles,
			expected: expected{
				paths: []string{},
			},
		},
		{
			name:     "canceled context returns canceled",
			listName: "",
			setup:    setupFiles,
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				errIs: context.Canceled,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			fileInfos, err := localStore.List(ctx, testCase.listName, testCase.options)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Nil(t, fileInfos)
				return
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
				require.Nil(t, fileInfos)
				return
			default:
				require.NoError(t, err)
			}

			if testCase.expected.paths != nil {
				require.ElementsMatch(t, testCase.expected.paths, fileInfoPaths(fileInfos))
			}

			for _, fileInfo := range fileInfos {
				require.False(t, fileInfo.IsDir)
				require.False(t, fileInfo.LastModified.IsZero())
			}

			if testCase.verify != nil {
				testCase.verify(t, fileInfos)
			}
		})
	}
}

func TestLocalStore_Copy(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs       error
		errContains string
		fileContent map[string]string
		missing     []string
	}

	type testData struct {
		name            string
		setup           func(t *testing.T, rootPath string)
		buildContext    func(t *testing.T) context.Context
		sourceName      string
		destinationName string
		options         storage.WriteOptions
		expected        expected
		verify          func(t *testing.T, rootPath string)
	}

	tests := []testData{
		{
			name:            "copies file to new destination",
			sourceName:      "source.json",
			destinationName: "destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", `{"ok":true}`)
			},
			expected: expected{
				fileContent: map[string]string{
					"source.json":      `{"ok":true}`,
					"destination.json": `{"ok":true}`,
				},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "copies file to nested destination",
			sourceName:      "source.json",
			destinationName: "nested/destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", `{"ok":true}`)
			},
			expected: expected{
				fileContent: map[string]string{
					"source.json":             `{"ok":true}`,
					"nested/destination.json": `{"ok":true}`,
				},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "copy overwrites destination by default",
			sourceName:      "source.json",
			destinationName: "destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", "new")
				writeTestFile(t, rootPath, "destination.json", "old")
			},
			expected: expected{
				fileContent: map[string]string{
					"source.json":      "new",
					"destination.json": "new",
				},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "fail if exists preserves existing destination",
			sourceName:      "source.json",
			destinationName: "destination.json",
			options:         storage.WriteOptions{FailIfExists: true},
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", "new")
				writeTestFile(t, rootPath, "destination.json", "old")
			},
			expected: expected{
				errIs: fs.ErrExist,
				fileContent: map[string]string{
					"source.json":      "new",
					"destination.json": "old",
				},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "missing source returns not exist",
			sourceName:      "missing.json",
			destinationName: "destination.json",
			expected: expected{
				errIs:   os.ErrNotExist,
				missing: []string{"destination.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "directory source returns directory error",
			sourceName:      "nested",
			destinationName: "destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				errIs:   storage.ErrIsDirectory,
				missing: []string{"destination.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "canceled context preserves source and does not copy",
			sourceName:      "source.json",
			destinationName: "destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", `{"ok":true}`)
			},
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				errIs: context.Canceled,
				fileContent: map[string]string{
					"source.json": `{"ok":true}`,
				},
				missing: []string{"destination.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			err := localStore.Copy(ctx, testCase.sourceName, testCase.destinationName, testCase.options)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
			default:
				require.NoError(t, err)
			}

			for fileName, expectedContent := range testCase.expected.fileContent {
				require.Equal(t, expectedContent, readTestFile(t, rootPath, fileName))
			}

			for _, missingFile := range testCase.expected.missing {
				_, err := os.Stat(filepath.Join(rootPath, filepath.FromSlash(missingFile)))
				require.ErrorIs(t, err, os.ErrNotExist)
			}

			if testCase.verify != nil {
				testCase.verify(t, rootPath)
			}
		})
	}
}

func TestLocalStore_Move(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs       error
		errContains string
		fileContent map[string]string
		missing     []string
	}

	type testData struct {
		name            string
		setup           func(t *testing.T, rootPath string)
		buildContext    func(t *testing.T) context.Context
		sourceName      string
		destinationName string
		options         storage.WriteOptions
		expected        expected
		verify          func(t *testing.T, rootPath string)
	}

	tests := []testData{
		{
			name:            "moves file to new destination",
			sourceName:      "source.json",
			destinationName: "destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", `{"ok":true}`)
			},
			expected: expected{
				fileContent: map[string]string{
					"destination.json": `{"ok":true}`,
				},
				missing: []string{"source.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "moves file to nested destination",
			sourceName:      "source.json",
			destinationName: "nested/destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", `{"ok":true}`)
			},
			expected: expected{
				fileContent: map[string]string{
					"nested/destination.json": `{"ok":true}`,
				},
				missing: []string{"source.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "move overwrites destination by default",
			sourceName:      "source.json",
			destinationName: "destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", "new")
				writeTestFile(t, rootPath, "destination.json", "old")
			},
			expected: expected{
				fileContent: map[string]string{
					"destination.json": "new",
				},
				missing: []string{"source.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "fail if exists preserves source and destination",
			sourceName:      "source.json",
			destinationName: "destination.json",
			options:         storage.WriteOptions{FailIfExists: true},
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", "new")
				writeTestFile(t, rootPath, "destination.json", "old")
			},
			expected: expected{
				errIs: fs.ErrExist,
				fileContent: map[string]string{
					"source.json":      "new",
					"destination.json": "old",
				},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "missing source returns not exist",
			sourceName:      "missing.json",
			destinationName: "destination.json",
			expected: expected{
				errIs:   os.ErrNotExist,
				missing: []string{"destination.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "directory source returns directory error",
			sourceName:      "nested",
			destinationName: "destination",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "nested/file.json", `{"ok":true}`)
			},
			expected: expected{
				errIs: storage.ErrIsDirectory,
				fileContent: map[string]string{
					"nested/file.json": `{"ok":true}`,
				},
				missing: []string{"destination"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
		{
			name:            "canceled context preserves source and does not move",
			sourceName:      "source.json",
			destinationName: "destination.json",
			setup: func(t *testing.T, rootPath string) {
				writeTestFile(t, rootPath, "source.json", `{"ok":true}`)
			},
			buildContext: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: expected{
				errIs: context.Canceled,
				fileContent: map[string]string{
					"source.json": `{"ok":true}`,
				},
				missing: []string{"destination.json"},
			},
			verify: func(t *testing.T, rootPath string) {
				requireNoTempFiles(t, rootPath)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)

			ctx := context.Background()
			if testCase.buildContext != nil {
				ctx = testCase.buildContext(t)
			}

			if testCase.setup != nil {
				testCase.setup(t, rootPath)
			}

			// Act
			err := localStore.Move(ctx, testCase.sourceName, testCase.destinationName, testCase.options)

			// Assert
			switch {
			case testCase.expected.errIs != nil:
				require.ErrorIs(t, err, testCase.expected.errIs)
			case testCase.expected.errContains != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
			default:
				require.NoError(t, err)
			}

			for fileName, expectedContent := range testCase.expected.fileContent {
				require.Equal(t, expectedContent, readTestFile(t, rootPath, fileName))
			}

			for _, missingFile := range testCase.expected.missing {
				_, err := os.Stat(filepath.Join(rootPath, filepath.FromSlash(missingFile)))
				require.ErrorIs(t, err, os.ErrNotExist)
			}

			if testCase.verify != nil {
				testCase.verify(t, rootPath)
			}
		})
	}
}

func TestLocalStore_PathSafety(t *testing.T) {
	t.Parallel()

	type testData struct {
		name      string
		setup     func(t *testing.T, rootPath string, outsidePath string) string
		operation func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error
	}

	tests := []testData{
		{
			name: "put rejects parent traversal destination",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Put(ctx, unsafeName, strings.NewReader("changed"), storage.WriteOptions{})
			},
		},
		{
			name: "put rejects absolute destination",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Put(ctx, unsafeName, strings.NewReader("changed"), storage.WriteOptions{})
			},
		},
		{
			name: "get rejects absolute source",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				readCloser, _, err := localStore.Get(ctx, unsafeName)
				if readCloser != nil {
					require.NoError(t, readCloser.Close())
				}
				return err
			},
		},
		{
			name: "stat rejects absolute source",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				_, err := localStore.Stat(ctx, unsafeName)
				return err
			},
		},
		{
			name: "delete rejects absolute target",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Delete(ctx, unsafeName)
			},
		},
		{
			name: "get rejects parent traversal source",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				readCloser, _, err := localStore.Get(ctx, unsafeName)
				if readCloser != nil {
					require.NoError(t, readCloser.Close())
				}
				return err
			},
		},
		{
			name: "stat rejects parent traversal source",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				_, err := localStore.Stat(ctx, unsafeName)
				return err
			},
		},
		{
			name: "delete rejects parent traversal target",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Delete(ctx, unsafeName)
			},
		},
		{
			name: "list rejects parent traversal path",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				_, err := localStore.List(ctx, unsafeName, storage.ListOptions{})
				return err
			},
		},
		{
			name: "copy rejects parent traversal source",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Copy(ctx, unsafeName, "destination.json", storage.WriteOptions{})
			},
		},
		{
			name: "copy rejects parent traversal destination",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Copy(ctx, "source.json", unsafeName, storage.WriteOptions{})
			},
		},
		{
			name: "copy rejects absolute source",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Copy(ctx, unsafeName, "destination.json", storage.WriteOptions{})
			},
		},
		{
			name: "copy rejects absolute destination",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Copy(ctx, "source.json", unsafeName, storage.WriteOptions{})
			},
		},
		{
			name: "move rejects parent traversal source",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Move(ctx, unsafeName, "destination.json", storage.WriteOptions{})
			},
		},
		{
			name: "move rejects parent traversal destination",
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Move(ctx, "source.json", unsafeName, storage.WriteOptions{})
			},
		},
		{
			name: "move rejects absolute source",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Move(ctx, unsafeName, "destination.json", storage.WriteOptions{})
			},
		},
		{
			name: "move rejects absolute destination",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				return outsidePath
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Move(ctx, "source.json", unsafeName, storage.WriteOptions{})
			},
		},
		{
			name: "put rejects symlink escape path",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				linkPath := filepath.Join(rootPath, "outside")
				if err := os.Symlink(filepath.Dir(outsidePath), linkPath); err != nil {
					t.Skipf("unable to create symlink: %v", err)
				}
				return "outside/outside.txt"
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				return localStore.Put(ctx, unsafeName, strings.NewReader("changed"), storage.WriteOptions{})
			},
		},
		{
			name: "get rejects symlink escape path",
			setup: func(t *testing.T, rootPath string, outsidePath string) string {
				linkPath := filepath.Join(rootPath, "outside")
				if err := os.Symlink(filepath.Dir(outsidePath), linkPath); err != nil {
					t.Skipf("unable to create symlink: %v", err)
				}
				return "outside/outside.txt"
			},
			operation: func(ctx context.Context, localStore *storage.LocalStore, unsafeName string) error {
				readCloser, _, err := localStore.Get(ctx, unsafeName)
				if readCloser != nil {
					require.NoError(t, readCloser.Close())
				}
				return err
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				ctx                  = context.Background()
				rootPath, localStore = newTestLocalStore(t)
				outsideDir           = t.TempDir()
				outsidePath          = filepath.Join(outsideDir, "outside.txt")
			)

			require.NoError(t, os.WriteFile(outsidePath, []byte("outside"), 0o640))
			writeTestFile(t, rootPath, "source.json", "source")

			unsafeName := ""
			if testCase.setup != nil {
				unsafeName = testCase.setup(t, rootPath, outsidePath)
			} else {
				relativePath, err := filepath.Rel(rootPath, outsidePath)
				require.NoError(t, err)
				unsafeName = filepath.ToSlash(relativePath)
			}

			// Act
			err := testCase.operation(ctx, localStore, unsafeName)

			// Assert
			require.Error(t, err)
			require.Equal(t, "outside", string(requireReadFile(t, outsidePath)))
			require.Equal(t, "source", readTestFile(t, rootPath, "source.json"))
			requireNoTempFiles(t, rootPath)
		})
	}
}
