//go:build integration

package migration_test

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/migration"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	bhversion "github.com/specterops/bloodhound/cmd/api/src/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type gooseTestContext struct {
	ctx      context.Context
	gormDB   *gorm.DB
	migrator *migration.Migrator
}

func setupGooseTestContext(t *testing.T) gooseTestContext {
	t.Helper()
	connConf := pgtestdb.Custom(t, gooseGetPostgresConfig(t), pgtestdb.NoopMigrator{})
	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	migrator, err := migration.NewMigrator(gormDB)
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDatabase, closeErr := gormDB.DB()
		require.NoError(t, closeErr)
		require.NoError(t, sqlDatabase.Close())
		dbPool.Close()
	})

	return gooseTestContext{
		ctx:      context.Background(),
		gormDB:   gormDB,
		migrator: migrator,
	}
}

func gooseGetPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	config, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(config.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

// discoverGooseMigrationFiles reads the embedded migrations directory and
// returns the *.sql filenames (excluding the legacy subdirectory) sorted
// lexicographically, which matches goose's ordering.
func discoverGooseMigrationFiles(t *testing.T) []string {
	t.Helper()

	entries, err := fs.ReadDir(migration.FossMigrations, "migrations")
	require.NoError(t, err)

	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		filenames = append(filenames, entry.Name())
	}
	sort.Strings(filenames)
	return filenames
}

// discoverGooseVersions parses the leading numeric prefix of each migration
// filename and returns them sorted ascending. Deriving this at runtime avoids
// hard-coded version lists that must be hand-updated for every new migration.
func discoverGooseVersions(t *testing.T) []int64 {
	t.Helper()

	filenames := discoverGooseMigrationFiles(t)
	versions := make([]int64, 0, len(filenames))
	for _, filename := range filenames {
		parts := strings.SplitN(filename, "_", 2)
		parsed, err := strconv.ParseInt(parts[0], 10, 64)
		require.NoErrorf(t, err, "migration file %s must start with a numeric version prefix", filename)
		versions = append(versions, parsed)
	}
	sort.Slice(versions, func(a, b int) bool {
		return versions[a] < versions[b]
	})
	return versions
}

// expectedDescriptionForFile mirrors the derivation performed by
// Migrator.populateMigrationDescription: drop the ".sql" suffix, split on the
// first underscore, and turn remaining underscores into spaces.
func expectedDescriptionForFile(filename string) string {
	trimmed := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(trimmed, "_", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.ReplaceAll(parts[1], "_", " ")
}

func seedLegacyThroughVersion(t *testing.T, migrator *migration.Migrator, targetRawVersion string) {
	t.Helper()

	require.NoError(t, migrator.CreateMigrationSchema())

	manifest, err := manifestThroughVersion(migrator, targetRawVersion)
	require.NoError(t, err)

	require.NotEmpty(t, manifest.VersionTable)
	require.NoError(t, migrator.ExecuteMigrations(manifest))
}

func manifestThroughVersion(migrator *migration.Migrator, targetRawVersion string) (migration.Manifest, error) {
	var (
		filteredManifest = migration.NewManifest()
		targetVersion    bhversion.Version
		fullManifest     migration.Manifest
		err              error
	)

	targetVersion, err = bhversion.Parse(targetRawVersion)
	if err != nil {
		return migration.Manifest{}, err
	}

	fullManifest, err = migrator.GenerateManifest()
	if err != nil {
		return migration.Manifest{}, err
	}

	for _, versionString := range fullManifest.VersionTable {
		currentVersion, parseErr := bhversion.Parse(versionString)
		if parseErr != nil {
			return migration.Manifest{}, parseErr
		}

		if currentVersion.GreaterThan(targetVersion) {
			break
		}

		filteredManifest.VersionTable = append(filteredManifest.VersionTable, versionString)
		for _, migrationEntry := range fullManifest.Migrations[versionString] {
			filteredManifest.AddMigration(migrationEntry)
		}
	}

	return filteredManifest, nil
}

// assertGooseVersionsMatch asserts that the applied goose versions match the
// expected set. The WHERE clause filters out goose's internal version_id = 0
// sentinel row so the test is not coupled to goose's bootstrap behavior.
func assertGooseVersionsMatch(t *testing.T, db *gorm.DB, expectedVersions []int64) {
	t.Helper()

	var appliedVersions []int64
	err := db.Raw(`
		SELECT version_id
		FROM goose_db_version
		WHERE is_applied = true AND version_id > 0
		ORDER BY version_id
	`).Scan(&appliedVersions).Error
	require.NoError(t, err)
	assert.Equal(t, expectedVersions, appliedVersions)
}

func assertTableState(t *testing.T, db *gorm.DB, tableName string, expected bool) {
	t.Helper()

	var tableExists bool
	err := db.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = current_schema() AND table_name = ?
		)
	`, tableName).Scan(&tableExists).Error
	require.NoError(t, err)
	assert.Equalf(t, expected, tableExists, "table %s existence mismatch", tableName)
}

func assertColumnExists(t *testing.T, db *gorm.DB, tableName string, columnName string) {
	t.Helper()

	var columnExists bool
	err := db.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = ?
			  AND column_name = ?
		)
	`, tableName, columnName).Scan(&columnExists).Error
	require.NoError(t, err)
	assert.Truef(t, columnExists, "column %s.%s should exist", tableName, columnName)
}

func assertConstraintExists(t *testing.T, db *gorm.DB, constraintName string) {
	t.Helper()

	var constraintExists bool
	err := db.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM pg_constraint
			WHERE conname = ?
		)
	`, constraintName).Scan(&constraintExists).Error
	require.NoError(t, err)
	assert.Truef(t, constraintExists, "constraint %s should exist", constraintName)
}

func assertDescription(t *testing.T, db *gorm.DB, versionID int64, expectedDescription string) {
	t.Helper()

	var description string
	err := db.Raw(`
		SELECT COALESCE(description, '')
		FROM goose_db_version
		WHERE version_id = ?
	`, versionID).Scan(&description).Error
	require.NoError(t, err)
	assert.Equalf(t, expectedDescription, description, "description mismatch for version %d", versionID)
}

func assertAuthTokensCreatedByMigration(t *testing.T, db *gorm.DB) {
	t.Helper()

	assertColumnExists(t, db, "auth_tokens", "created_by")
	assertConstraintExists(t, db, "fk_auth_tokens_created_by")
}

// assertAllGooseDescriptionsPopulated walks every goose migration file and
// asserts goose_db_version.description matches the filename-derived value.
// A single regression in populateMigrationDescription would otherwise slip
// through the single-row spot checks the original tests relied on.
func assertAllGooseDescriptionsPopulated(t *testing.T, db *gorm.DB) {
	t.Helper()

	filenames := discoverGooseMigrationFiles(t)
	for _, filename := range filenames {
		parts := strings.SplitN(filename, "_", 2)
		parsedVersion, err := strconv.ParseInt(parts[0], 10, 64)
		require.NoError(t, err)
		assertDescription(t, db, parsedVersion, expectedDescriptionForFile(filename))
	}
}

// assertLegacyStepwiseEffects verifies side effects of v8.7.0..v9.0.0 legacy
// catch-up migrations. These assertions prove the legacy stepwise path ran
// before goose rather than being silently skipped.
func assertLegacyStepwiseEffects(t *testing.T, db *gorm.DB) {
	t.Helper()

	// v8.9.0 creates schema_findings and drops schema_relationship_findings
	assertTableState(t, db, "schema_findings", true)
	assertTableState(t, db, "schema_relationship_findings", false)
	// v8.9.0 adds auth_tokens.expires_at
	assertColumnExists(t, db, "auth_tokens", "expires_at")
	// v9.0.0 adds users.support_account
	assertColumnExists(t, db, "users", "support_account")
}

type gooseTestCase struct {
	name                     string
	seedThroughLegacyVersion string
	runTimes                 int
	expectBootstrapVersion2  bool
	expectStepwiseEffects    bool
}

// TestMigrator_ExecuteGooseMigrations exercises the four deployment shapes
// that ExecuteGooseMigrations must support plus an idempotency variant of
// the legacy path:
//
//  1. NewInstall                    - no legacy table, goose runs all migrations.
//  2. NewInstall_Idempotent         - same as (1), but executed twice.
//  3. LegacyAlreadyAtHead           - legacy table seeded through v9.0.0, no stepwise
//     catch-up required, goose bootstraps and applies remaining migrations.
//  4. LegacyBehindStepwise          - legacy table seeded through v8.6.0, stepwise
//     catches up through v9.0.0, goose bootstraps and applies remaining migrations.
//  5. LegacyBehindStepwise_Idempotent - same as (4), executed twice; the second run
//     sees no legacy table and must be a successful no-op.

// TestMigration_ExecuteGooseMigrations :
// Once we fully remove legacy migrations we will have to alter this file to only support integration testing for
// goose migrations moving forward.

func TestMigrator_RunMigrations(t *testing.T) {
	expectedVersions := discoverGooseVersions(t)
	db, _, migrator, err := integration.SetupTestMigrator(t)
	require.Nil(t, err)

	assert.Nil(t, migrator.ExecuteGooseMigrations(context.Background()))

	assertTableState(t, db, "migrations", false)
	assertTableState(t, db, "goose_db_version", true)
	assertGooseVersionsMatch(t, db, expectedVersions)
}
func TestMigrator_ExecuteGooseMigrations(t *testing.T) {

	testCases := []gooseTestCase{
		{
			name:                    "NewInstall",
			runTimes:                1,
			expectBootstrapVersion2: false,
		},
		{
			name:                    "NewInstall_Idempotent",
			runTimes:                2,
			expectBootstrapVersion2: false,
		},
		{
			name:                     "LegacyBehindStepwise",
			seedThroughLegacyVersion: "v8.6.0",
			runTimes:                 1,
			expectBootstrapVersion2:  true,
			expectStepwiseEffects:    true,
		},
		{
			name:                     "LegacyAlreadyAtHead",
			seedThroughLegacyVersion: "v9.0.0",
			runTimes:                 1,
			expectBootstrapVersion2:  true,
			expectStepwiseEffects:    true,
		},
		{
			name:                     "LegacyBehindStepwise_Idempotent",
			seedThroughLegacyVersion: "v8.6.0",
			runTimes:                 2,
			expectBootstrapVersion2:  true,
			expectStepwiseEffects:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testContext := setupGooseTestContext(t)

			if testCase.seedThroughLegacyVersion != "" {
				seedLegacyThroughVersion(t, testContext.migrator, testCase.seedThroughLegacyVersion)
				assertTableState(t, testContext.gormDB, "migrations", true)
			}

			for runIndex := 0; runIndex < testCase.runTimes; runIndex++ {
				require.NoErrorf(t,
					testContext.migrator.ExecuteGooseMigrations(testContext.ctx),
					"ExecuteGooseMigrations run %d", runIndex+1,
				)
			}

			// The legacy `migrations` table must be gone once goose succeeds,
			// regardless of whether it ever existed.
			assertTableState(t, testContext.gormDB, "migrations", false)
			assertTableState(t, testContext.gormDB, "goose_db_version", true)

			expectedVersions := discoverGooseVersions(t)
			if testCase.expectBootstrapVersion2 {
				// bootstrapGoose inserts version_id = 2 alongside 1 to reserve the
				// BHE baseline slot. In a BHCE-only test this row has no backing
				// file but must still appear in goose_db_version.
				expectedVersions = append([]int64{2}, expectedVersions...)
				sort.Slice(expectedVersions, func(a, b int) bool {
					return expectedVersions[a] < expectedVersions[b]
				})
			}
			assertGooseVersionsMatch(t, testContext.gormDB, expectedVersions)

			assertAuthTokensCreatedByMigration(t, testContext.gormDB)
			assertAllGooseDescriptionsPopulated(t, testContext.gormDB)

			if testCase.expectBootstrapVersion2 {
				// Version 2 is bootstrapped without a corresponding goose source
				// in the BHCE FS, so its description is expected to stay empty.
				assertDescription(t, testContext.gormDB, 2, "")
			}

			if testCase.expectStepwiseEffects {
				assertLegacyStepwiseEffects(t, testContext.gormDB)
			}
		})
	}
}
