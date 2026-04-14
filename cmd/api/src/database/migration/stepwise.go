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
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// HasMigrationTable is a utility for checking if migration schema is initialized. We assume that
// if the `migrations` table exists, the schema must be initialized, and vice versa.
func (s *Migrator) HasMigrationTable() (bool, error) {
	const tableCheckSQL = `select exists(select * from information_schema.tables where table_schema = current_schema() and table_name = 'migrations');`

	var hasTable bool
	return hasTable, s.DB.Raw(tableCheckSQL).Scan(&hasTable).Error
}

func (s *Migrator) ExecuteExtensionDataPopulation() error {
	const migrationSQLFilenameSuffix = ".sql"

	// loop through extensions data
	for _, source := range s.ExtensionsData {
		dirEntries, err := fs.ReadDir(source.FileSystem, source.Directory)
		if err != nil {
			return err
		}

		// loop through file system entries
		for _, entry := range dirEntries {
			if entry.IsDir() {
				continue
			}

			filename := filepath.Join(source.Directory, entry.Name())
			basename := filepath.Base(filename)

			if !strings.HasSuffix(basename, migrationSQLFilenameSuffix) {
				continue
			}

			slog.Info("Executing extension data population", slog.String("file", basename))
			if err := s.DB.Transaction(func(tx *gorm.DB) error {
				// read migration file content and execute
				if migrationContent, err := fs.ReadFile(source.FileSystem, filename); err != nil {
					return err
				} else if result := tx.Exec(string(migrationContent)); result.Error != nil {
					return result.Error
				}

				return nil
			}); err != nil {
				return fmt.Errorf("failed to execute extension data population for %s: %w", basename, err)
			}
		}
	}

	return nil
}

// ExecuteMigrations runs all necessary migrations for a deployment using goose.
// It begins by checking if a legacy migrations table exists to determine the deployment type.
//
// If the deployment is existing (has legacy migrations table), we bootstrap goose by
// seeding the goose_db_version table with the baseline migration marked as applied,
// then drop the legacy migrations table.
//
// If the deployment is new, goose will run the baseline migration to create the schema.
//
// Once bootstrap is complete (if needed), we run goose to apply any pending migrations.
func (s *Migrator) ExecuteMigrations() error {
	// Check for legacy migrations table (old customers)
	hasLegacyTable, err := s.HasMigrationTable()
	if err != nil {
		return fmt.Errorf("failed to check if migration table exists: %w", err)
	}
	if hasLegacyTable {
		// Seed goose_db_version so baseline is skipped
		if err := s.bootstrapGoose(); err != nil {
			return err
		}
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		s.SqlDB,
		s.GooseFS,
		goose.WithAllowOutofOrder(true),
	)
	if err != nil {
		return fmt.Errorf("failed to create goose provider: %w", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("failed to execute up migrations: %w", err)
	}

	// Only drop legacy table AFTER goose succeeds
	// This ensures we can retry bootstrap on failure
	if hasLegacyTable {
		if err := s.dropLegacyMigrationTable(); err != nil {
			return err
		}
	}
	slog.Info(
		"Successfully ran goose migrations",
		slog.String("fn", "ExecuteNewMigrations"),
		slog.String("db version", string(goose.DialectPostgres)),
	)
	return nil
}

func (s *Migrator) bootstrapGoose() error {
	// Use goose's actual table schema and mark baseline migrations as applied
	// Version 1 = BHCE baseline (00000000000001_init.sql)
	// Version 2 = BHE baseline (00000000000002_init.sql)
	result := s.DB.Exec(`
        CREATE TABLE IF NOT EXISTS goose_db_version (
            id SERIAL PRIMARY KEY,
            version_id BIGINT NOT NULL UNIQUE,
            is_applied BOOLEAN NOT NULL,
            tstamp TIMESTAMP DEFAULT now()
        );

        INSERT INTO goose_db_version (version_id, is_applied)
        VALUES (1, true), (2, true)
        ON CONFLICT (version_id) DO NOTHING;
    `)
	return result.Error
}

func (s *Migrator) dropLegacyMigrationTable() error {
	slog.Info("Dropping legacy migrations table")
	_, err := s.SqlDB.Exec(`DROP TABLE IF EXISTS migrations;`)
	return err
}
