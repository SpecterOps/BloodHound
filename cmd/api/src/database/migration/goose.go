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

// ExecuteGooseMigrations runs all necessary migrations for a deployment using goose.
// It handles three scenarios:
//
//  1. New installation: No legacy migrations table exists, goose runs the baseline
//     migration to create the schema.
//
//  2. Existing installation (up to date): Legacy migrations table exists and is current.
//     We bootstrap goose by marking baseline migrations as applied, then run goose
//     for any new migrations.
//
//  3. Existing installation (behind): Legacy migrations table exists but is behind.
//     We first run v8 stepwise migrations to catch up, then bootstrap goose and
//     run any new goose migrations.
//
// The legacy migrations table is only dropped after all migrations succeed.
func (s *Migrator) ExecuteGooseMigrations(ctx context.Context) error {
	// Check for legacy migrations table
	hasLegacyTable, err := s.HasMigrationTable()
	if err != nil {
		return fmt.Errorf("failed to check if migration table exists: %w", err)
	}

	if hasLegacyTable {
		// Existing database - check if behind and run v8 stepwise migrations if needed
		// This logic can be removed once v11 is released
		if err := s.runLegacyMigrationsIfNeeded(); err != nil {
			return fmt.Errorf("failed to run legacy migrations: %w", err)
		}

		// Bootstrap goose by marking baseline migrations as applied
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

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("failed to execute up migrations: %w", err)
	}

	if err := s.populateMigrationDescription(provider); err != nil {
		slog.Warn("Failed to populate description column", slog.Any("error", err))
	}

	// Only drop legacy table AFTER goose succeeds
	if hasLegacyTable {
		if err := s.dropLegacyMigrationTable(); err != nil {
			return err
		}
	}

	slog.Info("Successfully ran goose migrations")
	return nil
}

// runLegacyMigrationsIfNeeded checks if the database is behind on v8 stepwise migrations
// and runs them if needed. This can be removed once v11 is released.
func (s *Migrator) runLegacyMigrationsIfNeeded() error {
	lastMigration, err := s.LatestMigration()
	if err != nil {
		return fmt.Errorf("could not get latest migration: %w", err)
	}

	manifest, err := s.GenerateManifestAfterVersion(lastMigration.Version())
	if err != nil {
		return fmt.Errorf("failed to generate migration manifest: %w", err)
	}

	if len(manifest.VersionTable) == 0 {
		slog.Info("Legacy migrations are up to date")
		return nil
	}

	slog.Info("Running legacy v8 migrations to catch up",
		slog.String("from_version", lastMigration.Version().String()),
		slog.Int("pending_versions", len(manifest.VersionTable)))

	if err := s.ExecuteMigrations(manifest); err != nil {
		return fmt.Errorf("failed to execute legacy migrations: %w", err)
	}

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
            tstamp TIMESTAMP DEFAULT now(),
            description TEXT
        );

        -- Add description column if it doesn't exist (for existing tables)
        ALTER TABLE goose_db_version ADD COLUMN IF NOT EXISTS description TEXT;

        INSERT INTO goose_db_version (version_id, is_applied)
        VALUES (0, true), (1, true), (2, true)
        ON CONFLICT (version_id) DO NOTHING;
    `)
	return result.Error
}

func (s *Migrator) dropLegacyMigrationTable() error {
	slog.Info("Dropping legacy migrations table")
	_, err := s.SqlDB.Exec(`DROP TABLE IF EXISTS migrations;`)
	return err
}

func (s *Migrator) populateMigrationDescription(provider *goose.Provider) error {
	// use a transaction so failures don't congest the connection pool
	tx, err := s.SqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	if _, err = tx.Exec(`
        ALTER TABLE goose_db_version
        ADD COLUMN IF NOT EXISTS description TEXT
    `); err != nil {
		return err
	}
	sources := provider.ListSources()

	for _, source := range sources {
		filename := filepath.Base(source.Path)
		parts := strings.SplitN(strings.TrimSuffix(filename, ".sql"), "_", 2)

		var description string
		if len(parts) > 1 {
			description = strings.ReplaceAll(parts[1], "_", " ")
		}
		if _, err = tx.Exec(`
            UPDATE goose_db_version
            SET description = $1
            WHERE version_id = $2 AND (description IS NULL OR description = '')
        `, description, source.Version); err != nil {
			return err
		}
	}

	return tx.Commit()
}
