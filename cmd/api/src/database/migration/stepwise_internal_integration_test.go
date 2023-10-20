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
	assert.Len(t, filenames, 4)
	assert.Equal(t, filenames[1], "test_migrations/v0.0.1.sql")
	assert.Equal(t, filenames[2], "test_migrations/v0.1.0.sql")
	assert.Equal(t, filenames[3], "test_migrations/v1.0.0.sql")
}

func TestMigrator_ExecuteMigrations(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	migrator := setupMigrator(db)

	filenames, err := migrator.MigrationFilenames()
	assert.Nil(t, err)

	manifest, err := NewManifest(filenames)
	assert.Nil(t, err)

	var tableExists bool
	err = migrator.ExecuteMigrations([]Migration{manifest.migrations[0]})
	assert.Nil(t, err)
	db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migration_test')").Scan(&tableExists)
	assert.True(t, tableExists)

	var row = struct {
		Id  int
		Foo string
	}{}
	err = migrator.ExecuteMigrations([]Migration{manifest.migrations[1]})
	assert.Nil(t, err)
	db.Raw("SELECT * FROM migration_test").Scan(&row)
	assert.NotEmpty(t, row)
	assert.Equal(t, 1, row.Id)
	assert.Equal(t, "bar", row.Foo)

	err = migrator.ExecuteMigrations([]Migration{manifest.migrations[2]})
	assert.Nil(t, err)
	db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migration_test')").Scan(&tableExists)
	assert.False(t, tableExists)
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

		require.Nil(t, err)

		err = wipeDB(db)
		require.Nil(t, err)

		err = testMigrator.executeStepwiseMigrations()
		assert.Nil(t, err)

		err = db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migrations')").Scan(&migrationsTableExists).Error
		require.Nil(t, err)
		assert.True(t, migrationsTableExists)

		err = db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migration_test')").Scan(&testTableExists).Error
		require.Nil(t, err)
		assert.False(t, testTableExists, "migration test table unexpectedly exists after stepwise migration completed successfully")
	})

	t.Run("Production Schema Migration", func(t *testing.T) {
		var (
			migrationsTableExists bool
		)

		err = wipeDB(db)
		require.Nil(t, err)

		err = prodMigrator.executeStepwiseMigrations()
		require.Nil(t, err)

		err = db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migrations')").Scan(&migrationsTableExists).Error
		require.Nil(t, err)
		assert.True(t, migrationsTableExists)

		ver, err := testMigrator.LatestMigration()
		assert.Nil(t, err)
		assert.Equal(t, version.GetVersion(), ver.Version())
	})
}
