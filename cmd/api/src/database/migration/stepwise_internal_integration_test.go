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

// TODO: This file is unconventionally named because it was written as both an internal test and an integration test. Proper
// refactoring of these tests is needed, but is non-trivial.

package migration

import (
	"embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/specterops/bloodhound/src/version"
)

//go:embed test_migrations
var testMigrations embed.FS

const migrationDir = "test_migrations"

type MigrationTest struct {
	Id   int64
	Name string
	Foo  string
}

func (MigrationTest) TableName() string {
	return "migration_test"
}

func setupDB() (*gorm.DB, error) {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		return nil, fmt.Errorf("failed to load integration test config: %w", err)
	} else {
		return gorm.Open(postgres.Open(cfg.Database.PostgreSQLConnectionString()))
	}
}

func wipeDB(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		sql := `
				do $$ declare
					r record;
				begin
					for r in (select tablename from pg_tables where schemaname = 'public') loop
						execute 'drop table if exists ' || quote_ident(r.tablename) || ' cascade';
					end loop;
				end $$;
			`
		return tx.Exec(sql).Error
	})
}

func setupMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		migrations:   testMigrations,
		db:           db,
		migrationDir: migrationDir,
	}
}

func TestMigrator_MigrationFilenames(t *testing.T) {
	var migrator = setupMigrator(nil)

	filenames, err := migrator.MigrationFilenames()
	assert.Nil(t, err)
	assert.Len(t, filenames, 5)
	assert.Equal(t, "test_migrations/schema.sql", filenames[0])
	assert.Equal(t, "test_migrations/v0.0.1.sql", filenames[1])
	assert.Equal(t, "test_migrations/v0.1.0.sql", filenames[2])
	assert.Equal(t, "test_migrations/v0.1.1.sql", filenames[3])
	assert.Equal(t, "test_migrations/v1.0.0.sql", filenames[4])
}

func TestMigrator_ExecuteMigrations(t *testing.T) {
	var (
		entries     []MigrationTest
		tableExists bool
	)

	db, err := setupDB()
	require.Nil(t, err)
	migrator := setupMigrator(db)

	filenames, err := migrator.MigrationFilenames()
	require.Nil(t, err)

	manifest, err := NewManifest(filenames)
	require.Nil(t, err)

	t.Run("Create Table Migration", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations([]Migration{manifest.migrations[0]}))
		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migration_test')").Scan(&tableExists)
		assert.True(t, tableExists)
	})

	t.Run("Add Data To Table Migration", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations([]Migration{manifest.migrations[1]}))
		require.Nil(t, db.Order("id ASC").Find(&entries).Error)

		assert.NotEmpty(t, entries[0])
		assert.Equal(t, int64(1), entries[0].Id)
		assert.Equal(t, "foo", entries[0].Name)
		assert.Equal(t, "foo", entries[0].Foo)

		assert.NotEmpty(t, entries[1])
		assert.Equal(t, int64(2), entries[1].Id)
		assert.Equal(t, "foo", entries[1].Name)
		assert.Equal(t, "bar", entries[1].Foo)

		assert.NotEmpty(t, entries[2])
		assert.Equal(t, int64(3), entries[2].Id)
		assert.Equal(t, "foo", entries[2].Name)
		assert.Equal(t, "bar", entries[2].Foo)

		assert.NotEmpty(t, entries[3])
		assert.Equal(t, int64(4), entries[3].Id)
		assert.Equal(t, "bar", entries[3].Name)
		assert.Equal(t, "baz", entries[3].Foo)
	})

	t.Run("Deduplicate Data Migration", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations([]Migration{manifest.migrations[2]}))
		result := db.Order("id ASC").Find(&entries)
		require.Nil(t, result.Error)

		assert.Equal(t, int64(3), result.RowsAffected)

		assert.NotEmpty(t, entries[0])
		assert.Equal(t, int64(1), entries[0].Id)
		assert.Equal(t, "foo_2", entries[0].Name)
		assert.Equal(t, "foo", entries[0].Foo)

		assert.NotEmpty(t, entries[1])
		assert.Equal(t, int64(3), entries[1].Id)
		assert.Equal(t, "foo", entries[1].Name)
		assert.Equal(t, "bar", entries[1].Foo)

		assert.NotEmpty(t, entries[2])
		assert.Equal(t, int64(4), entries[2].Id)
		assert.Equal(t, "bar", entries[2].Name)
		assert.Equal(t, "baz", entries[2].Foo)
	})

	t.Run("Drop Table Migration", func(t *testing.T) {
		require.Nil(t, migrator.ExecuteMigrations([]Migration{manifest.migrations[3]}))
		db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migration_test')").Scan(&tableExists)
		assert.False(t, tableExists)
	})
}

func TestMigrator_Migrate(t *testing.T) {
	db, err := setupDB()
	require.Nil(t, err)

	testMigrator := setupMigrator(db)
	prodMigrator := NewMigrator(db)

	t.Run("Full Initial Migration", func(t *testing.T) {
		var (
			migrationsTableExists bool
			testTableExists       bool
		)

		require.Nil(t, wipeDB(db))
		assert.Nil(t, testMigrator.executeStepwiseMigrations())

		require.Nil(t, db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migrations')").Scan(&migrationsTableExists).Error)
		assert.True(t, migrationsTableExists)

		require.Nil(t, db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migration_test')").Scan(&testTableExists).Error)
		assert.False(t, testTableExists, "migration test table unexpectedly exists after stepwise migration completed successfully")
	})

	t.Run("Production Schema Migration", func(t *testing.T) {
		var (
			migrationsTableExists bool
		)

		require.Nil(t, wipeDB(db))

		require.Nil(t, prodMigrator.executeStepwiseMigrations())

		require.Nil(t, db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migrations')").Scan(&migrationsTableExists).Error)
		assert.True(t, migrationsTableExists)

		ver, err := testMigrator.LatestMigration()
		require.Nil(t, err)
		assert.Equal(t, version.GetVersion(), ver.Version())
	})
}
