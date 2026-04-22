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
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/version"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"gorm.io/gorm"
)

//go:embed migrations
var FossMigrations embed.FS

//go:embed extensions
var ExtensionMigrations embed.FS

// Source is meant to be a file system source that contains SQL migration files.
type Source struct {
	FileSystem fs.FS
	Directory  string
}

// Migrator is the main SQL migration tool for BloodHound.
type Migrator struct {
	Sources        []Source // Deprecated: Sources supports legacy v8 stepwise migrations.  Can be removed after v10 is released
	ExtensionsData []Source
	DB             *gorm.DB
	SqlDB          *sql.DB
	GooseFS        fs.FS
}

// Migration contains information about a specific migration such as the file location, it's Source, and Version. Can be removed after v10 release
type Migration struct {
	Filename string
	Source   fs.FS
	Version  version.Version
}

// NewMigrator returns a new Migrator with the FossMigrations Source predefined.
func NewMigrator(db *gorm.DB) (*Migrator, error) {
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to connect to database: %v", attr.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	fossMigrationsSubFS, err := fs.Sub(FossMigrations, "migrations")
	if err != nil {
		slog.Error("Failed to open foss migrations directory: %v", attr.Error(err))
		return nil, fmt.Errorf("failed to open foss migrations directory: %v", err)
	}
	return &Migrator{
		// Deprecated: Sources supports legacy v8 stepwise migrations. Can be removed after v10 is released.
		Sources: []Source{
			{FileSystem: FossMigrations, Directory: "migrations/v8"},
		},
		ExtensionsData: []Source{
			{FileSystem: ExtensionMigrations, Directory: "extensions"},
		},
		GooseFS: MergedFS(fossMigrationsSubFS),
		DB:      db,
		SqlDB:   sqlDB,
	}, nil
}
