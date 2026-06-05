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

		// Only validate SQL migration files in the active BHE migrations directory
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

// TestGooseMigrationFilenameUseCurrentVersionPrefix is utilized to ensure that all migration files are named according
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
