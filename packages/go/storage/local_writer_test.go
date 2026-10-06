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
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/packages/go/storage"
	"github.com/stretchr/testify/require"
)

func TestLocalPendingFileCommit(t *testing.T) {
	t.Parallel()

	type testData struct {
		name            string
		content         string
		existingContent string
		options         storage.WriteOptions
	}

	tests := []testData{
		{
			name:    "publishes a new file",
			content: "content",
		},
		{
			name: "publishes an empty file",
		},
		{
			name:    "conditionally publishes a new file",
			content: "content",
			options: storage.WriteOptions{FailIfExists: true},
		},
		{
			name:            "replaces an existing file",
			content:         "new content",
			existingContent: "old content",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			rootPath, localStore := newTestLocalStore(t)
			if testCase.existingContent != "" {
				writeTestFile(t, rootPath, "nested/file.bin", testCase.existingContent)
			}

			pendingFile, err := localStore.BeginWrite(context.Background(), "nested/file.bin", testCase.options)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, pendingFile.Abort()) })
			bytesWritten, err := pendingFile.Write([]byte(testCase.content))
			require.NoError(t, err)
			require.Equal(t, len(testCase.content), bytesWritten)

			if testCase.existingContent != "" {
				require.Equal(t, testCase.existingContent, readTestFile(t, rootPath, "nested/file.bin"))
			} else {
				_, err = os.Stat(filepath.Join(rootPath, "nested", "file.bin"))
				require.ErrorIs(t, err, fs.ErrNotExist)
			}

			// Act
			committedName, err := pendingFile.Commit()

			// Assert
			require.NoError(t, err)
			require.Equal(t, "nested/file.bin", committedName)
			require.Equal(t, testCase.content, readTestFile(t, rootPath, committedName))

			repeatedName, err := pendingFile.Commit()
			require.NoError(t, err)
			require.Equal(t, committedName, repeatedName)
			require.NoError(t, pendingFile.Abort())
			require.Equal(t, testCase.content, readTestFile(t, rootPath, committedName))
			_, err = pendingFile.Write([]byte("late"))
			require.ErrorIs(t, err, io.ErrClosedPipe)
			requireNoTempFiles(t, rootPath)
		})
	}
}

func TestLocalPendingFileAbortDiscardsUnpublishedData(t *testing.T) {
	t.Parallel()

	// Arrange
	rootPath, localStore := newTestLocalStore(t)
	pendingFile, err := localStore.BeginWrite(context.Background(), "aborted", storage.WriteOptions{})
	require.NoError(t, err)
	_, err = pendingFile.Write([]byte("discard"))
	require.NoError(t, err)

	// Act
	err = pendingFile.Abort()

	// Assert
	require.NoError(t, err)
	require.NoError(t, pendingFile.Abort())
	_, err = os.Stat(filepath.Join(rootPath, "aborted"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = pendingFile.Write([]byte("late"))
	require.ErrorIs(t, err, io.ErrClosedPipe)
	_, err = pendingFile.Commit()
	require.ErrorIs(t, err, io.ErrClosedPipe)
	requireNoTempFiles(t, rootPath)
}

func TestLocalPendingFileConditionalCommit(t *testing.T) {
	t.Parallel()

	// Arrange
	rootPath, localStore := newTestLocalStore(t)
	firstFile, err := localStore.BeginWrite(context.Background(), "same", storage.WriteOptions{FailIfExists: true})
	require.NoError(t, err)
	secondFile, err := localStore.BeginWrite(context.Background(), "same", storage.WriteOptions{FailIfExists: true})
	require.NoError(t, err)
	_, err = firstFile.Write([]byte("first"))
	require.NoError(t, err)
	_, err = secondFile.Write([]byte("second"))
	require.NoError(t, err)

	// Act
	_, err = firstFile.Commit()
	require.NoError(t, err)
	_, secondCommitErr := secondFile.Commit()

	// Assert
	require.ErrorIs(t, secondCommitErr, fs.ErrExist)
	require.NoError(t, secondFile.Abort())
	require.Equal(t, "first", readTestFile(t, rootPath, "same"))
	requireNoTempFiles(t, rootPath)
}

func TestLocalPendingFileCanceledWriteBecomesStickyFailure(t *testing.T) {
	t.Parallel()

	// Arrange
	canceledContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	rootPath, localStore := newTestLocalStore(t)
	pendingFile, err := localStore.BeginWrite(canceledContext, "cancelled", storage.WriteOptions{})
	require.NoError(t, err)
	cancel()

	// Act
	_, firstErr := pendingFile.Write([]byte("data"))
	_, repeatedErr := pendingFile.Write([]byte("more"))
	_, commitErr := pendingFile.Commit()

	// Assert
	require.ErrorIs(t, firstErr, context.Canceled)
	require.ErrorIs(t, firstErr, storage.ErrWriteFailed)
	require.ErrorIs(t, repeatedErr, firstErr)
	require.ErrorIs(t, commitErr, firstErr)
	require.NoError(t, pendingFile.Abort())
	requireNoTempFiles(t, rootPath)
}

func TestLocalPendingFileRejectsPathTraversal(t *testing.T) {
	t.Parallel()

	// Arrange
	rootPath, localStore := newTestLocalStore(t)
	escapedPath := filepath.Join(filepath.Dir(rootPath), "escaped")

	// Act
	pendingFile, err := localStore.BeginWrite(context.Background(), "../escaped", storage.WriteOptions{})

	// Assert
	require.Error(t, err)
	require.Nil(t, pendingFile)
	_, err = os.Stat(escapedPath)
	require.ErrorIs(t, err, os.ErrNotExist)
	requireNoTempFiles(t, rootPath)
}
