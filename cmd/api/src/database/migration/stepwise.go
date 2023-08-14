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
	"path/filepath"
	"sort"
	"strings"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/version"
	"gorm.io/gorm"
)

const (
	migrationSQLFilenameSuffix = ".sql"
	migrationDirname           = "migrations"
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
func (s *Migrator) MigrationFilenames(migrationDir string) ([]string, error) {
	var migrationFilenames []string

	if dirEntries, err := s.migrations.ReadDir(migrationDir); err != nil {
		return nil, err
	} else {
		for _, entry := range dirEntries {
			if !entry.IsDir() {
				migrationFilenames = append(migrationFilenames, filepath.Join(migrationDir, entry.Name()))
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

		if migrationContent, err := s.migrations.ReadFile(migration.Filename); err != nil {
			return err
		} else {
			// this used to scan the sql in line by line, changed to fix a multiline issue.
			// @JHop - running line by line was to give us pinpoint details on which statement failed in the collection
			// of statements running them all together muddies error negotiation and introspection
			err := s.db.Transaction(func(tx *gorm.DB) error {
				if result := tx.Exec(string(migrationContent)); result.Error != nil {
					return result.Error
				}

				migrationEntry := model.NewMigration(migration.Version)
				if result := tx.Create(&migrationEntry); result.Error != nil {
					return result.Error
				}

				return nil
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Migrator) HasMigrationTable() (bool, error) {
	const tableCheckSQL = `select exists(select * from information_schema.tables where table_schema = current_schema() and table_name = 'migrations');`

	var hasTable bool
	return hasTable, s.db.Raw(tableCheckSQL).Scan(&hasTable).Error
}

func (s *Migrator) CreateMigrationTable() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Migrator().AutoMigrate(
			// Migration model
			&model.Migration{},
		)
	})
}

func (s *Migrator) executeStepwiseMigrations(models []any) error {
	if hasTable, err := s.HasMigrationTable(); err != nil {
		return fmt.Errorf("failed to check if migraiton table exists: %w", err)
	} else if !hasTable {
		if err := s.CreateMigrationTable(); err != nil {
			return fmt.Errorf("failed to create migraiton table: %w", err)
		}
	}

	if err := s.gormAutoMigrate(models); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	if migrationFilenames, err := s.MigrationFilenames(migrationDirname); err != nil {
		return err
	} else if manifest, err := NewManifest(migrationFilenames); err != nil {
		return err
	} else if lastMigration, err := s.LatestMigration(); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		currentVersion := version.GetVersion()
		log.Infof("This is a new database. Creating a migration entry for version %s", currentVersion)

		return s.db.Transaction(func(tx *gorm.DB) error {
			migrationEntry := model.NewMigration(currentVersion)
			return tx.Create(&migrationEntry).Error
		})
	} else {
		return s.ExecuteMigrations(manifest.After(lastMigration.Version()))
	}
}
