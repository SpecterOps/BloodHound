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

package datapipe

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe/mocks"
	"github.com/specterops/bloodhound/packages/go/storage"
	storagemocks "github.com/specterops/bloodhound/packages/go/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type cleanupInternalFileInfo struct {
	name    string
	modTime time.Time
}

func (s cleanupInternalFileInfo) Name() string       { return s.name }
func (s cleanupInternalFileInfo) Size() int64        { return 0 }
func (s cleanupInternalFileInfo) Mode() fs.FileMode  { return 0 }
func (s cleanupInternalFileInfo) ModTime() time.Time { return s.modTime }
func (s cleanupInternalFileInfo) IsDir() bool        { return false }
func (s cleanupInternalFileInfo) Sys() any           { return nil }

type cleanupInternalDirEntry struct {
	name    string
	isDir   bool
	mode    fs.FileMode
	info    fs.FileInfo
	infoErr error
}

func (s cleanupInternalDirEntry) Name() string {
	return s.name
}

func (s cleanupInternalDirEntry) IsDir() bool {
	return s.isDir
}

func (s cleanupInternalDirEntry) Type() fs.FileMode {
	return s.mode
}

func (s cleanupInternalDirEntry) Info() (fs.FileInfo, error) {
	return s.info, s.infoErr
}

func TestNormalizeStoragePrefixes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{
		"retained",
		"work/tmp",
		"nested/prefix",
	}, normalizeStoragePrefixes([]string{
		" retained/ ",
		"/work/tmp",
		"../escape",
		"",
		".",
		"nested//prefix/",
	}))
}

func TestOrphanFileSweeper_IsExcludedStoragePath(t *testing.T) {
	t.Parallel()

	sweeper := NewOrphanFileSweeper(nil, t.TempDir(), t.TempDir(), "retained", "/nested/prefix/")

	testCases := []struct {
		name        string
		logicalPath string
		expected    bool
	}{
		{
			name:        "exact prefix",
			logicalPath: "retained",
			expected:    true,
		},
		{
			name:        "file under prefix",
			logicalPath: "retained/file.json",
			expected:    true,
		},
		{
			name:        "leading slash",
			logicalPath: "/nested/prefix/file.json",
			expected:    true,
		},
		{
			name:        "similar prefix",
			logicalPath: "retained-other/file.json",
			expected:    false,
		},
		{
			name:        "traversal path",
			logicalPath: "../retained/file.json",
			expected:    false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, sweeper.isExcludedStoragePath(testCase.logicalPath))
		})
	}
}

func TestOrphanFileSweeper_AddExpectedStoragePath(t *testing.T) {
	t.Parallel()

	tempDirectory := t.TempDir()
	sweeper := NewOrphanFileSweeper(nil, tempDirectory, t.TempDir())

	expectedFiles := map[string]struct{}{}
	sweeper.addExpectedStoragePath(expectedFiles, "nested//file.json")
	sweeper.addExpectedStoragePath(expectedFiles, filepath.Join(tempDirectory, "absolute.json"))
	sweeper.addExpectedStoragePath(expectedFiles, filepath.Join(filepath.Dir(tempDirectory), "outside.json"))
	sweeper.addExpectedStoragePath(expectedFiles, "../escape.json")
	sweeper.addExpectedStoragePath(expectedFiles, ".")
	sweeper.addExpectedStoragePath(expectedFiles, " ")

	require.Len(t, expectedFiles, 2)
	assert.Contains(t, expectedFiles, "nested/file.json")
	assert.Contains(t, expectedFiles, "absolute.json")
}

func TestOrphanFileSweeper_AddExpectedLocalPath(t *testing.T) {
	t.Parallel()

	tempDirectory := t.TempDir()
	absoluteExpectedPath := filepath.Join(tempDirectory, "absolute.json")
	sweeper := NewOrphanFileSweeper(nil, tempDirectory, t.TempDir())

	expectedLocalFiles := map[string]struct{}{}
	sweeper.addExpectedLocalPath(expectedLocalFiles, "nested/file.json")
	sweeper.addExpectedLocalPath(expectedLocalFiles, absoluteExpectedPath)
	sweeper.addExpectedLocalPath(expectedLocalFiles, "../escape.json")
	sweeper.addExpectedLocalPath(expectedLocalFiles, ".")
	sweeper.addExpectedLocalPath(expectedLocalFiles, " ")

	require.Len(t, expectedLocalFiles, 2)
	assert.Contains(t, expectedLocalFiles, filepath.Join(tempDirectory, "nested", "file.json"))
	assert.Contains(t, expectedLocalFiles, absoluteExpectedPath)
}

func TestOrphanFileSweeper_ClearStoredIngestFiles(t *testing.T) {
	t.Parallel()

	var (
		mockCtrl        = gomock.NewController(t)
		mockFileService = storagemocks.NewMockFileService(mockCtrl)
		tempDirectory   = t.TempDir()
		sweeper         = NewOrphanFileSweeper(nil, tempDirectory, t.TempDir(), "retained", "/excluded/")
		oldTime         = time.Now().Add(-25 * time.Hour)
		youngTime       = time.Now()
	)

	absoluteExpectedPath := filepath.Join(tempDirectory, "absolute-expected.json")

	mockFileService.EXPECT().
		ListFiles(gomock.Any(), "", storage.ListOptions{Recursive: true}).
		Return([]storage.FileInfo{
			{Path: "expected.json", LastModified: oldTime},
			{Path: "absolute-expected.json", LastModified: oldTime},
			{Path: "retained/file.json", LastModified: oldTime},
			{Path: "/excluded/file.json", LastModified: oldTime},
			{Path: "young.json", LastModified: youngTime},
			{Path: "directory", LastModified: oldTime, IsDir: true},
			{Path: "old.json", LastModified: oldTime},
		}, nil)
	mockFileService.EXPECT().DeleteFile(gomock.Any(), "old.json").Return(nil)

	sweeper.clearStoredIngestFiles(context.Background(), mockFileService, []string{"expected.json", absoluteExpectedPath})
}

func TestOrphanFileSweeper_ClearStoredIngestFilesNormalizesListedPaths(t *testing.T) {
	t.Parallel()

	var (
		mockCtrl        = gomock.NewController(t)
		mockFileService = storagemocks.NewMockFileService(mockCtrl)
		tempDirectory   = t.TempDir()
		sweeper         = NewOrphanFileSweeper(nil, tempDirectory, t.TempDir(), "/excluded/")
		oldTime         = time.Now().Add(-25 * time.Hour)
	)

	mockFileService.EXPECT().
		ListFiles(gomock.Any(), "", storage.ListOptions{Recursive: true}).
		Return([]storage.FileInfo{
			{Path: "/expected.json", LastModified: oldTime},
			{Path: "/old.json", LastModified: oldTime},
		}, nil)
	mockFileService.EXPECT().DeleteFile(gomock.Any(), "old.json").Return(nil)

	sweeper.clearStoredIngestFiles(context.Background(), mockFileService, []string{"expected.json"})
}

func TestOrphanFileSweeper_ClearLegacyLocalIngestFiles(t *testing.T) {
	t.Parallel()

	var (
		mockCtrl      = gomock.NewController(t)
		mockFileOps   = mocks.NewMockFileOperations(mockCtrl)
		tempDirectory = t.TempDir()
		sweeper       = NewOrphanFileSweeper(mockFileOps, tempDirectory, t.TempDir())
		oldTime       = time.Now().Add(-25 * time.Hour)
		youngTime     = time.Now()
	)

	mockFileOps.EXPECT().ReadDir(tempDirectory).Return([]os.DirEntry{
		cleanupInternalDirEntry{
			name: "expected.json",
		},
		cleanupInternalDirEntry{
			name: "young.json",
			info: cleanupInternalFileInfo{
				name:    "young.json",
				modTime: youngTime,
			},
		},
		cleanupInternalDirEntry{
			name:  "nested",
			isDir: true,
		},
		cleanupInternalDirEntry{
			name: "old.json",
			info: cleanupInternalFileInfo{
				name:    "old.json",
				modTime: oldTime,
			},
		},
	}, nil)
	mockFileOps.EXPECT().RemoveAll(filepath.Join(tempDirectory, "nested")).Return(nil)
	mockFileOps.EXPECT().RemoveAll(filepath.Join(tempDirectory, "old.json")).Return(nil)

	sweeper.clearLegacyLocalIngestFiles(context.Background(), []string{"expected.json", "../old.json"})
}

func TestOrphanFileSweeper_ClearLegacyLocalIngestFilesRemovesOrphanDirectories(t *testing.T) {
	t.Parallel()

	var (
		mockCtrl      = gomock.NewController(t)
		mockFileOps   = mocks.NewMockFileOperations(mockCtrl)
		tempDirectory = t.TempDir()
		sweeper       = NewOrphanFileSweeper(mockFileOps, tempDirectory, t.TempDir())
	)

	mockFileOps.EXPECT().ReadDir(tempDirectory).Return([]os.DirEntry{
		cleanupInternalDirEntry{
			name:  "orphan",
			isDir: true,
		},
		cleanupInternalDirEntry{
			name:  "expected-parent",
			isDir: true,
		},
	}, nil)
	mockFileOps.EXPECT().RemoveAll(filepath.Join(tempDirectory, "orphan")).Return(nil)

	sweeper.clearLegacyLocalIngestFiles(context.Background(), []string{"expected-parent/file.json"})
}

func TestClearLocalIngestScratch(t *testing.T) {
	t.Parallel()

	t.Run("removes old files and keeps young files and directories", func(t *testing.T) {
		t.Parallel()

		var (
			scratchDirectory = t.TempDir()
			oldFilePath      = filepath.Join(scratchDirectory, "old.tmp")
			youngFilePath    = filepath.Join(scratchDirectory, "young.tmp")
			nestedDirectory  = filepath.Join(scratchDirectory, "nested")
			oldTime          = time.Now().Add(-2 * time.Hour)
		)

		require.NoError(t, os.WriteFile(oldFilePath, []byte("old"), 0o600))
		require.NoError(t, os.Chtimes(oldFilePath, oldTime, oldTime))
		require.NoError(t, os.WriteFile(youngFilePath, []byte("young"), 0o600))
		require.NoError(t, os.Mkdir(nestedDirectory, 0o700))

		ClearLocalIngestScratch(context.Background(), scratchDirectory, time.Hour)

		requirePathNotExists(t, oldFilePath)
		requirePathExists(t, youngFilePath)
		requirePathExists(t, nestedDirectory)
	})

	t.Run("exits without deleting files when context is canceled", func(t *testing.T) {
		t.Parallel()

		var (
			scratchDirectory = t.TempDir()
			oldFilePath      = filepath.Join(scratchDirectory, "old.tmp")
			oldTime          = time.Now().Add(-2 * time.Hour)
		)

		require.NoError(t, os.WriteFile(oldFilePath, []byte("old"), 0o600))
		require.NoError(t, os.Chtimes(oldFilePath, oldTime, oldTime))

		ctx, done := context.WithCancel(context.Background())
		done()

		ClearLocalIngestScratch(ctx, scratchDirectory, time.Hour)

		requirePathExists(t, oldFilePath)
	})
}

func requirePathExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	require.NoError(t, err)
}

func requirePathNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}
