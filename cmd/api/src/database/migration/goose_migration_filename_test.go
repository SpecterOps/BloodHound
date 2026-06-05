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
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database/migration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Version numbers as of now are hard coded and will need to be manually updated upon next version release (v10)
const (
	gooseMigrationDirectory      = "migrations"
	currentGooseMigrationVersion = "v9"
	fossInitBaseline             = "00000000000001_init.sql"
)

var gooseMigrationFilenamePattern = regexp.MustCompile("^[0-9]{14}_" + currentGooseMigrationVersion + "_[a-z0-9]+(_[a-z0-9]+)*[.]sql$")

func invalidGooseMigrationFilenames(t *testing.T, migrationFileSystem fs.FS, allowedInitMigrationFilename string) []string {
	var (
		directoryEntries          []fs.DirEntry
		err                       error
		invalidMigrationFilenames []string
	)

	t.Helper()

	directoryEntries, err = fs.ReadDir(migrationFileSystem, gooseMigrationDirectory)
	require.NoError(t, err)

	for _, directoryEntry := range directoryEntries {
		migrationFilename := directoryEntry.Name()

		// Only validate SQL migration files in the active FOSS migrations directory
		if directoryEntry.IsDir() || !strings.HasSuffix(migrationFilename, ".sql") {
			continue
		}

		// The FOSS baseline migration should be skipped
		if migrationFilename == allowedInitMigrationFilename {
			continue
		}

		if !gooseMigrationFilenamePattern.MatchString(migrationFilename) {
			invalidMigrationFilenames = append(invalidMigrationFilenames, migrationFilename)
		}
	}

	sort.Strings(invalidMigrationFilenames)
	return invalidMigrationFilenames
}

// TestGooseMigrationFilenamesUseCurrentVersionPrefix is utilized to ensure that all migration files are named according
// to the set naming convention.
func TestGooseMigrationFilenamesUseCurrentVersionPrefix(t *testing.T) {
	invalidMigrationFilenames := invalidGooseMigrationFilenames(t, migration.FossMigrations, fossInitBaseline)
	assert.Emptyf(
		t,
		invalidMigrationFilenames,
		"BHCE: Goose migration filenames must match: timestamp_versionNumber_your_file_name.sql. Invalid files: %s",
		strings.Join(invalidMigrationFilenames, ", "),
	)
}
