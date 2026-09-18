// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package migration_test

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/cmd/api/src/database/migration"
)

func TestMergedFS_Open(t *testing.T) {
	var (
		firstContent  = "first set of migration data"
		secondContent = "second set of migration data"
	)
	t.Run("opens file from first filesystem", func(t *testing.T) {
		var (
			firstFS  = fstest.MapFS{"001_first.sql": &fstest.MapFile{Data: []byte(firstContent)}}
			secondFS = fstest.MapFS{"002_second.sql": &fstest.MapFile{Data: []byte(secondContent)}}
			merged   = migration.MergedFS(firstFS, secondFS)
		)

		file, err := merged.Open("001_first.sql")
		require.NoError(t, err)
		defer file.Close()

		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "first set of migration data", string(content))
	})

	t.Run("opens file from second filesystem", func(t *testing.T) {
		var (
			firstFS  = fstest.MapFS{"001_first.sql": &fstest.MapFile{Data: []byte(firstContent)}}
			secondFS = fstest.MapFS{"002_second.sql": &fstest.MapFile{Data: []byte(secondContent)}}
			merged   = migration.MergedFS(firstFS, secondFS)
		)

		file, err := merged.Open("002_second.sql")
		require.NoError(t, err)
		defer file.Close()

		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "second set of migration data", string(content))
	})

	t.Run("first filesystem takes priority on duplicate names", func(t *testing.T) {
		var (
			firstFS  = fstest.MapFS{"0000001_init.sql": &fstest.MapFile{Data: []byte(firstContent)}}
			secondFS = fstest.MapFS{"0000001_init.sql": &fstest.MapFile{Data: []byte(secondContent)}}
			merged   = migration.MergedFS(firstFS, secondFS)
		)

		file, err := merged.Open("0000001_init.sql")
		require.NoError(t, err)
		defer file.Close()

		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "first set of migration data", string(content))
	})

	t.Run("returns path error when file not found in any filesystem", func(t *testing.T) {
		var (
			firstFS  = fstest.MapFS{"0000001_init.sql": &fstest.MapFile{Data: []byte(firstContent)}}
			secondFS = fstest.MapFS{"0000002_init.sql": &fstest.MapFile{Data: []byte(secondContent)}}
			merged   = migration.MergedFS(firstFS, secondFS)
		)

		_, err := merged.Open("nonexistent.sql")
		require.Error(t, err)

		var pathErr *fs.PathError
		require.ErrorAs(t, err, &pathErr)
		assert.Equal(t, "open", pathErr.Op)
		assert.Equal(t, "nonexistent.sql", pathErr.Path)
		assert.ErrorIs(t, pathErr.Err, fs.ErrNotExist)
	})

	t.Run("returns path error with no filesystems", func(t *testing.T) {
		merged := migration.MergedFS()

		_, err := merged.Open("00000005_temp.sql")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func TestMergedFS_ReadDir(t *testing.T) {
	var (
		firstContent  = "first set of migration data"
		secondContent = "second set of migration data"
		thirdContent  = "third set of migration data"
		fourthContent = "fourth set of migration data"
	)
	t.Run("returns sorted entries from multiple filesystems", func(t *testing.T) {
		var (
			firstFS = fstest.MapFS{
				"00003_insertTempData.sql": &fstest.MapFile{Data: []byte(thirdContent)},
				"00001_init.sql":           &fstest.MapFile{Data: []byte(firstContent)},
			}
			secondFS = fstest.MapFS{
				"00002_createTempTable.sql": &fstest.MapFile{Data: []byte(secondContent)},
				"00004_createUserTable.sql": &fstest.MapFile{Data: []byte(fourthContent)},
			}
			merged = migration.MergedFS(firstFS, secondFS)
		)

		readDirFS, ok := merged.(fs.ReadDirFS)
		require.True(t, ok)

		entries, err := readDirFS.ReadDir(".")
		require.NoError(t, err)
		require.Len(t, entries, 4)
		assert.Equal(t, "00001_init.sql", entries[0].Name())
		assert.Equal(t, "00002_createTempTable.sql", entries[1].Name())
		assert.Equal(t, "00003_insertTempData.sql", entries[2].Name())
		assert.Equal(t, "00004_createUserTable.sql", entries[3].Name())
	})

	t.Run("deduplicates entries with the same name using last filesystem wins", func(t *testing.T) {
		var (
			firstFS  = fstest.MapFS{"00001_migration.sql": &fstest.MapFile{Data: []byte(firstContent)}}
			secondFS = fstest.MapFS{
				"00001_migration.sql": &fstest.MapFile{Data: []byte(secondContent)},
				"00002_migration.sql": &fstest.MapFile{Data: []byte(thirdContent)},
			}
			merged = migration.MergedFS(firstFS, secondFS)
		)

		readDirFS, ok := merged.(fs.ReadDirFS)
		require.True(t, ok)

		file1, err := merged.Open("00001_migration.sql")
		require.NoError(t, err)
		content1, err := io.ReadAll(file1)
		require.NoError(t, err)
		file2, err := merged.Open("00002_migration.sql")
		require.NoError(t, err)
		content2, err := io.ReadAll(file2)
		require.NoError(t, err)

		entries, err := readDirFS.ReadDir(".")
		require.NoError(t, err)
		require.Len(t, entries, 2)
		assert.Equal(t, "00001_migration.sql", entries[0].Name())
		assert.Equal(t, "first set of migration data", string(content1))
		assert.Equal(t, "00002_migration.sql", entries[1].Name())
		assert.Equal(t, "third set of migration data", string(content2))
	})

	t.Run("excludes directories", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"001_migration.sql": &fstest.MapFile{Data: []byte("migration")},
			"subdir/nested.sql": &fstest.MapFile{Data: []byte("nested")},
		}
		merged := migration.MergedFS(fileSystem)

		readDirFS, ok := merged.(fs.ReadDirFS)
		require.True(t, ok)

		entries, err := readDirFS.ReadDir(".")
		require.NoError(t, err)

		for _, entry := range entries {
			assert.False(t, entry.IsDir(), "directory entries should be excluded")
		}
	})

	t.Run("returns error for non-root path", func(t *testing.T) {
		var (
			fileSystem = fstest.MapFS{"001_migration.sql": &fstest.MapFile{Data: []byte("data")}}
			merged     = migration.MergedFS(fileSystem)
		)

		readDirFS, ok := merged.(fs.ReadDirFS)
		require.True(t, ok)

		_, err := readDirFS.ReadDir("subdir")
		require.Error(t, err)

		var pathErr *fs.PathError
		require.ErrorAs(t, err, &pathErr)
		assert.Equal(t, "readdir", pathErr.Op)
		assert.Equal(t, "subdir", pathErr.Path)
		assert.ErrorIs(t, pathErr.Err, fs.ErrNotExist)
	})

	t.Run("returns empty slice with no filesystems", func(t *testing.T) {
		merged := migration.MergedFS()

		readDirFS, ok := merged.(fs.ReadDirFS)
		require.True(t, ok)

		entries, err := readDirFS.ReadDir(".")
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	// multiple subdirectories in migrations directory
	t.Run("reads from root and sub directories containing migration files", func(t *testing.T) {
		var (
			firstFS = fstest.MapFS{
				"migrations/v9/00003_insertTempData.sql": &fstest.MapFile{Data: []byte(thirdContent)},
				"00001_init.sql":                         &fstest.MapFile{Data: []byte(firstContent)},
			}
			secondFS = fstest.MapFS{
				"00002_createTempTable.sql":               &fstest.MapFile{Data: []byte(secondContent)},
				"migrations/v9/00004_createUserTable.sql": &fstest.MapFile{Data: []byte(fourthContent)},
			}
			merged = migration.MergedFS(firstFS, secondFS)
		)

		readDirFS, ok := merged.(fs.ReadDirFS)
		require.True(t, ok)

		entries, err := readDirFS.ReadDir(".")
		require.NoError(t, err)
		require.Len(t, entries, 4)
		assert.Equal(t, "00001_init.sql", entries[0].Name())
		assert.Equal(t, "00002_createTempTable.sql", entries[1].Name())
		assert.Equal(t, "00003_insertTempData.sql", entries[2].Name())
		assert.Equal(t, "00004_createUserTable.sql", entries[3].Name())

	})
}
