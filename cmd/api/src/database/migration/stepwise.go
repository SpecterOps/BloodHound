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
	"io/fs"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/version"
	"gorm.io/gorm"
)

// LatestMigration retrieves the last entry in the migration table to determine
// what the last successful migration version was.
func (s *Migrator) LatestMigration() (model.Migration, error) {
	var migration model.Migration
	return migration, s.DB.Order("id desc").First(&migration).Error
}

// ExecuteMigrations takes a Manifest and runs all migrations contained in it, by version.
// It starts by ranging over the Manifest.VersionTable, which is an ordered list of all versions
// covered in the Manifest. Each version entry in the Manifest map may have multiple migrations
// from different sources. To ensure consistency, all migrations for each version are run in
// a transaction, as well as logging the successful migration entry in the `migrations` table.
func (s *Migrator) ExecuteMigrations(manifest Manifest) error {
	// range over the manifest map[version: string]migrations: []Migration
	for _, versionString := range manifest.VersionTable {

		// version integrity check
		thisVersion, err := version.Parse(versionString)
		if err != nil {
			return fmt.Errorf("invalid version `%s` detected: %w", thisVersion, err)
		}

		// execute the migration(s) for this version in a transaction
		log.Infof("Executing SQL migrations for %s", versionString)
		if err := s.DB.Transaction(func(tx *gorm.DB) error {

			for _, migration := range manifest.Migrations[versionString] {
				// version validation
				if !thisVersion.Equals(migration.Version) {
					return fmt.Errorf("migration version mismatch: expected %s, got %s", thisVersion.String(), migration.Version.String())
				}

				// read migration file content and execute
				if migrationContent, err := fs.ReadFile(migration.Source, migration.Filename); err != nil {
					return err
				} else if result := tx.Exec(string(migrationContent)); result.Error != nil {
					return result.Error
				}
			}

			// create entry for this migration version in the migration table
			migrationEntry := model.NewMigration(thisVersion)
			if result := tx.Create(&migrationEntry); result.Error != nil {
				return result.Error
			} else {
				return nil
			}
		}); err != nil {
			return fmt.Errorf("failed to execute migrations for %s: %w", versionString, err)
		}

	}

	return nil
}

// HasMigrationTable is a utility for checking if migration schema is initialized. We assume that
// if the `migrations` table exists, the schema must be initialized, and vice versa.
func (s *Migrator) HasMigrationTable() (bool, error) {
	const tableCheckSQL = `select exists(select * from information_schema.tables where table_schema = current_schema() and table_name = 'migrations');`

	var hasTable bool
	return hasTable, s.DB.Raw(tableCheckSQL).Scan(&hasTable).Error
}

// CreateMigrationSchema creates all the necessary SQL schema for tracking migration status.
func (s *Migrator) CreateMigrationSchema() error {
	const (
		createMigrationTableSql = `CREATE TABLE IF NOT EXISTS migrations (
    id integer NOT NULL,
    updated_at timestamp with time zone,
    major integer,
    minor integer,
    patch integer
);`
		createMigrationIdSequenceSql = `CREATE SEQUENCE IF NOT EXISTS migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE migrations_id_seq OWNED BY migrations.id;
ALTER TABLE ONLY migrations ALTER COLUMN id SET DEFAULT nextval('migrations_id_seq'::regclass);
ALTER TABLE ONLY migrations ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);`
	)

	log.Infof("Creating migration schema...")
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		if result := tx.Exec(createMigrationTableSql); result.Error != nil {
			return fmt.Errorf("failed to creation migration table: %w", result.Error)
		} else if result = tx.Exec(createMigrationIdSequenceSql); result.Error != nil {
			return fmt.Errorf("failt to create migration id sequence: %w", result.Error)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failure during migration schema transaction: %w", err)
	}
	return nil
}

func (s *Migrator) RequiresMigration() (bool, error) {
	// check if migration table exists to determine type of manifest to generate
	if hasTable, err := s.HasMigrationTable(); err != nil {
		return false, fmt.Errorf("failed to check if migration table exists: %w", err)
	} else if !hasTable {
		// no migration table, assume this is new installation and requires migration
		return true, nil
	}

	if lastMigration, err := s.LatestMigration(); err != nil {
		return false, fmt.Errorf("could not get latest migration: %w", err)
	} else if manifest, err := s.GenerateManifestAfterVersion(lastMigration.Version()); err != nil {
		return false, fmt.Errorf("failed to generate migration manifest from previous version: %w", err)
	} else {
		return len(manifest.VersionTable) > 0, nil
	}
}

// executeStepwiseMigrations will run all necessary migrations for a deployment.
// It begins by checking if migration schema exists. If it does not, we assume the
// deployment is a new installation, otherwise we assume it may have migration updates.
//
// If the deployment is new, we install migration schema to track migrations
// and then build a manifest of all existing migrations
//
// If the deployment already existed, we query for the last successful migration run
// and then build a manifest starting after the last successful version
//
// Once schema is verified and a manifest is created, we run ExecuteMigrations.
func (s *Migrator) executeStepwiseMigrations() error {
	var (
		manifest      Manifest
		lastMigration model.Migration
	)

	// check if migration table exists to determine type of manifest to generate
	if hasTable, err := s.HasMigrationTable(); err != nil {
		return fmt.Errorf("failed to check if migration table exists: %w", err)
	} else if !hasTable {
		// no migration table, assume this is new installation
		log.Infof("This is a new SQL database. Initializing schema...")
		//initialize migration schema and generate full manifest
		if err = s.CreateMigrationSchema(); err != nil {
			return fmt.Errorf("failed to create migration schema: %w", err)
		} else if manifest, err = s.GenerateManifest(); err != nil {
			return fmt.Errorf("failed to generate migration manifest for new install: %w", err)
		}
	} else {
		// migration table exists, assume we might be upgrading and generate manifest from last version migrated
		if lastMigration, err = s.LatestMigration(); err != nil {
			return fmt.Errorf("could not get latest migration: %w", err)
		} else if manifest, err = s.GenerateManifestAfterVersion(lastMigration.Version()); err != nil {
			return fmt.Errorf("failed to generate migration manifest from previous version: %w", err)
		}
	}

	// run migrations using the manifest we generated
	if len(manifest.VersionTable) == 0 {
		log.Infof("No new SQL migrations to run")
		return nil
	} else if err := s.ExecuteMigrations(manifest); err != nil {
		return fmt.Errorf("could not execute migrations: %w", err)
	} else {
		return nil
	}
}
