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
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/src/test/integration/utils"
)

//go:embed test_migrations
var testMigrations embed.FS

const migrationDir = "test_migrations"

func setupDB() (*gorm.DB, error) {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		return nil, err
	} else {
		return gorm.Open(postgres.Open(cfg.Database.PostgreSQLConnectionString()))
	}
}

func setupMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		migrations: testMigrations,
		db:         db,
	}
}

func TestMigrator_MigrationFilenames(t *testing.T) {
	var migrator = setupMigrator(nil)

	filenames, err := migrator.MigrationFilenames(migrationDir)
	assert.Nil(t, err)
	assert.Len(t, filenames, 3)
	assert.Equal(t, "test_migrations/v0.0.1.sql", filenames[0])
	assert.Equal(t, "test_migrations/v0.1.0.sql", filenames[1])
	assert.Equal(t, "test_migrations/v1.0.0.sql", filenames[2])
}

func TestMigrator_ExecuteMigrations(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	migrator := setupMigrator(db)

	filenames, err := migrator.MigrationFilenames(migrationDir)
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
	assert.Equal(t, row.Id, 1)
	assert.Equal(t, row.Foo, "bar")

	err = migrator.ExecuteMigrations([]Migration{manifest.migrations[2]})
	assert.Nil(t, err)
	db.Raw("SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='migration_test')").Scan(&tableExists)
	assert.False(t, tableExists)
}
