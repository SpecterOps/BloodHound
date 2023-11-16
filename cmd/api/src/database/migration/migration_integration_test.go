// Copyright 2023 Specter Ops, Inc.
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

//go:build serial_integration
// +build serial_integration

package migration_test

import (
	"embed"
	"testing"

	"github.com/specterops/bloodhound/src/database/migration"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_migrations/system1
var testMigrationSystem1 embed.FS

//go:embed test_migrations/system2
var testMigrationSystem2 embed.FS

type FooTest struct {
	Id  int64
	Foo string
}

func (FooTest) TableName() string {
	return "foo_test"
}

type BarTest struct {
	Id  int64
	Bar string
}

func (BarTest) TableName() string {
	return "bar_test"
}

// TestMigrator_LatestMigration tests that the Migrator can retrieve
// the last migration entry in the `migration` table.
func TestMigrator_LatestMigration(t *testing.T) {
	_, migrator, err := integration.SetupTestMigrator(migration.Source{FileSystem: testMigrationSystem1, Directory: "test_migrations/system1"})
	require.Nil(t, err)

	require.Nil(t, migrator.CreateMigrationSchema())

	manifest, err := migrator.GenerateManifest()
	require.Nil(t, err)

	require.Nil(t, migrator.ExecuteMigrations(manifest))

	ver, err := migrator.LatestMigration()
	require.Nil(t, err)
	assert.Equal(t, 2, ver.Version().Major)
	assert.Equal(t, 0, ver.Version().Minor)
	assert.Equal(t, 0, ver.Version().Patch)
}

// TestMigrator_ExecuteMigrations tests the integrity of the stepwise
// migration process. It generates a full manifest and then splits it
// into chunks to test various scenarios the Migrator should handle
// when dealing with multiple migration Source's.
// Each chunk is similar to cases a user may encounter:
// Case 1 simulates a new installation, running schema and some initial migrations.
// Case 2 simulates independent migrations with independent versioning.
// Case 3 simulates Source order-dependent migration operations within the same version.
// Case 4 cleans up using independent migrations from the same version.
func TestMigrator_ExecuteMigrations(t *testing.T) {
	var (
		tableExists bool
		fooTests    []FooTest
		barTests    []BarTest
	)
	db, migrator, err := integration.SetupTestMigrator(
		migration.Source{FileSystem: testMigrationSystem1, Directory: "test_migrations/system1"},
		migration.Source{FileSystem: testMigrationSystem2, Directory: "test_migrations/system2"},
	)
	require.Nil(t, err)

	require.Nil(t, migrator.CreateMigrationSchema())

	fullManifest, err := migrator.GenerateManifest()
	require.Nil(t, err)

	// split manifest into chunks to test different cases
	testManifests := []migration.Manifest{
		{VersionTable: []string{"v0.0.0", "v0.1.0"}},
		{VersionTable: []string{"v0.1.1", "v0.2.0"}},
		{VersionTable: []string{"v1.0.0"}},
		{VersionTable: []string{"v2.0.0"}},
	}
	// construct 4 test manifests by hand
	for idx, manifest := range testManifests {
		for _, ver := range manifest.VersionTable {
			for _, migr := range fullManifest.Migrations[ver] {
				testManifests[idx].AddMigration(migr)
			}
		}
	}

	/**
	 * CASE 1 - SETUP: both systems create schema (v0.0.0) and insert some data (v0.1.0)
	 */
	t.Run("Setup", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations(testManifests[0]))

		// check tables
		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='foo_test')").Scan(&tableExists)
		assert.True(t, tableExists)

		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='bar_test')").Scan(&tableExists)
		assert.True(t, tableExists)

		// check foos
		require.Nil(t, db.Order("id ASC").Find(&fooTests).Error)

		assert.NotEmpty(t, fooTests[0])
		assert.Equal(t, int64(1), fooTests[0].Id)
		assert.Equal(t, "foo", fooTests[0].Foo)

		assert.NotEmpty(t, fooTests[1])
		assert.Equal(t, int64(2), fooTests[1].Id)
		assert.Equal(t, "foobar", fooTests[1].Foo)

		assert.NotEmpty(t, fooTests[2])
		assert.Equal(t, int64(3), fooTests[2].Id)
		assert.Equal(t, "foobaz", fooTests[2].Foo)

		// check bars
		require.Nil(t, db.Order("id ASC").Find(&barTests).Error)

		assert.NotEmpty(t, barTests[0])
		assert.Equal(t, int64(1), barTests[0].Id)
		assert.Equal(t, "bar", barTests[0].Bar)

		assert.NotEmpty(t, barTests[1])
		assert.Equal(t, int64(2), barTests[1].Id)
		assert.Equal(t, "barfoo", barTests[1].Bar)

		assert.NotEmpty(t, barTests[2])
		assert.Equal(t, int64(3), barTests[2].Id)
		assert.Equal(t, "barbaz", barTests[2].Bar)
	})

	/**
	 * CASE 2 - INDEPENDENCE: system1 renames a column (v0.1.1), system2 empties its table (v0.2.0)
	 */
	t.Run("Independence", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations(testManifests[1]))

		// check column rename
		db.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='foo_test' and column_name='baz')").Scan(&tableExists)
		assert.True(t, tableExists)

		// check for empty bars table
		require.Nil(t, db.Order("id ASC").Find(&barTests).Error)
		assert.Empty(t, barTests)
	})

	/**
	 * CASE 3 - ORDERING: system1 will first rename its table (v1.0.0), system2 will copy data from renamed table (v1.0.0)
	 */
	t.Run("Ordering", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations(testManifests[2]))

		// check table rename
		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='foo_test')").Scan(&tableExists)
		assert.False(t, tableExists)

		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='baz_test')").Scan(&tableExists)
		assert.True(t, tableExists)

		// check copied data
		require.Nil(t, db.Order("id ASC").Find(&barTests).Error)

		assert.NotEmpty(t, barTests[0])
		assert.Equal(t, int64(4), barTests[0].Id)
		assert.Equal(t, "foo", barTests[0].Bar)

		assert.NotEmpty(t, barTests[1])
		assert.Equal(t, int64(5), barTests[1].Id)
		assert.Equal(t, "foobar", barTests[1].Bar)

		assert.NotEmpty(t, barTests[2])
		assert.Equal(t, int64(6), barTests[2].Id)
		assert.Equal(t, "foobaz", barTests[2].Bar)
	})

	/**
	 * CASE 4 - CLEANUP: both systems drop their tables (v2.0.0)
	 */
	t.Run("Cleanup", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations(testManifests[3]))

		// check that tables are cleaned up
		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='baz_test')").Scan(&tableExists)
		assert.False(t, tableExists)

		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='bar_test')").Scan(&tableExists)
		assert.False(t, tableExists)
	})
}

// TestMigrator_HasMigrationTable makes sure the Migrator can properly
// detect the `migrations` table.
func TestMigrator_HasMigrationTable(t *testing.T) {
	_, migrator, err := integration.SetupTestMigrator()
	require.Nil(t, err)

	require.Nil(t, migrator.CreateMigrationSchema())

	hasTable, err := migrator.HasMigrationTable()
	assert.Nil(t, err)
	assert.True(t, hasTable)
}

// TestMigrator_CreateMigrationSchema ensures that the Migrator can
// successfully generate all schema needed to track migrations.
func TestMigrator_CreateMigrationSchema(t *testing.T) {
	const (
		checkTableExistsSql    = "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migrations')"
		checkSequenceExistsSql = "SELECT EXISTS(SELECT 0 FROM pg_class WHERE relname='migrations_id_seq')"
		checkIdSequencing      = "INSERT INTO migrations (major, minor, patch) VALUES (0, 1, 0), (0, 2, 0)"
		checkPrimaryKeySql     = `SELECT COUNT(0)
	FROM information_schema.table_constraints tc
	INNER JOIN information_schema.constraint_column_usage cu
	ON cu.CONSTRAINT_NAME=tc.CONSTRAINT_NAME
	WHERE
	tc.CONSTRAINT_TYPE='PRIMARY KEY'
	AND tc.TABLE_NAME='migrations'
	AND cu.COLUMN_NAME='id'`
	)

	var (
		valueExists bool
		count       int64
	)

	db, migrator, err := integration.SetupTestMigrator()
	require.Nil(t, err)

	assert.Nil(t, migrator.CreateMigrationSchema())

	t.Run("Check table exists", func(t *testing.T) {
		assert.Nil(t, db.Raw(checkTableExistsSql).Scan(&valueExists).Error)
		assert.True(t, valueExists)
	})

	t.Run("Check sequence exists", func(t *testing.T) {
		assert.Nil(t, db.Raw(checkSequenceExistsSql).Scan(&valueExists).Error)
		assert.True(t, valueExists)
	})

	t.Run("Check sequence hooked up", func(t *testing.T) {
		assert.Nil(t, db.Exec(checkIdSequencing).Error)
	})

	t.Run("Check primary key constraint", func(t *testing.T) {
		assert.Nil(t, db.Raw(checkPrimaryKeySql).Scan(&count).Error)
		assert.Equal(t, int64(1), count)
	})
}

// TestMigrator_Migrate tests the integrity of FossMigrations.
func TestMigrator_Migrate(t *testing.T) {
	_, migrator, err := integration.SetupTestMigrator(migration.Source{FileSystem: migration.FossMigrations, Directory: "migrations"})
	require.Nil(t, err)

	manifest, err := migrator.GenerateManifest()
	require.Nil(t, err)

	assert.Nil(t, migrator.Migrate())

	lastVersionInManifest := manifest.VersionTable[len(manifest.VersionTable)-1]
	latestMigration, err := migrator.LatestMigration()
	assert.Nil(t, err)
	assert.Equal(t, lastVersionInManifest, latestMigration.Version().String())
}
