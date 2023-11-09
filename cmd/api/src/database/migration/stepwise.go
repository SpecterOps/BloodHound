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

package migration

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/version"
	"gorm.io/gorm"
)

const (
	migrationSQLFilenameSuffix = ".sql"
)

type Migration struct {
	Version  version.Version
	Filename string
}

type Manifest struct {
	migrations []Migration
}

func NewManifest(filenames []string) (Manifest, error) {
	var migrations []Migration

	for _, filename := range filenames {
		basename := filepath.Base(filename)

		if strings.HasPrefix(basename, version.Prefix) && strings.HasSuffix(basename, migrationSQLFilenameSuffix) {
			rawVersion := strings.TrimSuffix(basename, migrationSQLFilenameSuffix)

			if migrationVersion, err := version.Parse(rawVersion); err != nil {
				return Manifest{}, err
			} else {
				migrations = append(migrations, Migration{
					Version:  migrationVersion,
					Filename: filename,
				})
			}
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version.LessThan(migrations[j].Version)
	})

	return Manifest{
		migrations: migrations,
	}, nil
}

func (s Manifest) All() []Migration {
	return s.migrations
}

func (s Manifest) After(target version.Version) []Migration {
	for idx, migration := range s.migrations {
		if migration.Version.GreaterThan(target) {
			log.Infof("Version %s is greater than %s", migration.Version, target)
			return s.migrations[idx:]
		}
	}

	return nil
}

// Additional migrator functions to support stepwise migrations
func (s *Migrator) MigrationFilenames() ([]string, error) {
	var migrationFilenames []string

	if dirEntries, err := s.migrations.ReadDir(s.migrationDir); err != nil {
		return nil, err
	} else {
		for _, entry := range dirEntries {
			if !entry.IsDir() {
				migrationFilenames = append(migrationFilenames, filepath.Join(s.migrationDir, entry.Name()))
			}
		}
	}

	return migrationFilenames, nil
}

func (s *Migrator) LatestMigration() (model.Migration, error) {
	var migration model.Migration
	return migration, s.db.Order("id desc").First(&migration).Error
}

func (s *Migrator) ExecuteMigrations(migrations []Migration) error {
	for _, migration := range migrations {
		log.Infof("Executing migration: %s", migration.Version)

		s.executeSQLFile(migration.Filename, migration.Version)
	}

	return nil
}

func (s *Migrator) executeSQLFile(filename string, ver version.Version) error {
	if migrationContent, err := s.migrations.ReadFile(filename); err != nil {
		return err
	} else if err := s.db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Exec(string(migrationContent)); result.Error != nil {
			return result.Error
		}

		migrationEntry := model.NewMigration(ver)
		if result := tx.Create(&migrationEntry); result.Error != nil {
			return result.Error
		}

		return nil
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Migrator) HasMigrationTable() (bool, error) {
	const tableCheckSQL = `select exists(select * from information_schema.tables where table_schema = current_schema() and table_name = 'migrations');`

	var hasTable bool
	return hasTable, s.db.Raw(tableCheckSQL).Scan(&hasTable).Error
}

func (s *Migrator) executeStepwiseMigrations() error {
	if hasTable, err := s.HasMigrationTable(); err != nil {
		return fmt.Errorf("failed to check if migration table exists: %w", err)
	} else if !hasTable {
		log.Infof("This is a new database. Initializing schema...")
		if err := s.executeSQLFile(path.Join(s.migrationDir, "schema.sql"), version.Version{}); err != nil {
			return fmt.Errorf("failed to create initial schema: %w", err)
		}
	}

	currentVersionMigration := model.NewMigration(version.GetVersion())

	if migrationFilenames, err := s.MigrationFilenames(); err != nil {
		return err
	} else if manifest, err := NewManifest(migrationFilenames); err != nil {
		return err
	} else if lastMigration, err := s.LatestMigration(); err != nil {
		return fmt.Errorf("could not get latest migration: %w", err)
	} else if err := s.ExecuteMigrations(manifest.After(lastMigration.Version())); err != nil {
		return fmt.Errorf("could not execute migrations: %w", err)
	} else if lastMigration != currentVersionMigration {
		return s.db.Create(&currentVersionMigration).Error
	} else {
		return nil
	}
}
