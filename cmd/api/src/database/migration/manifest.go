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
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/specterops/bloodhound/src/version"
)

// Manifest is a collection of available migrations. VersionTable is used to order and store
// all version of migrations contained in the manifest. Migrations is the actual
// underlying map of [version:[]migration]
type Manifest struct {
	VersionTable []string
	Migrations   map[string][]Migration
}

// NewManifest creates a new Manifest and initializes the migrations map
func NewManifest() Manifest {
	return Manifest{
		Migrations: make(map[string][]Migration),
	}
}

// AddMigration will add migrations to the correct location in the migrations
// map, with care to make sure the map is initialized, as well as the
// version location within the map.
func (s *Manifest) AddMigration(migration Migration) {
	if s.Migrations == nil {
		s.Migrations = make(map[string][]Migration)
	}
	if migrations, ok := s.Migrations[migration.Version.String()]; ok {
		s.Migrations[migration.Version.String()] = append(migrations, migration)
	} else {
		s.Migrations[migration.Version.String()] = []Migration{migration}
	}
}

// GenerateManifest is a wrapper around GenerateManifestAfterVersion, using
// -1.-1.-1 as the version. This ensures that a full manifest of all
// available migrations is generated. This is most useful for new installations.
func (s *Migrator) GenerateManifest() (Manifest, error) {
	return s.GenerateManifestAfterVersion(version.Version{
		Major: -1,
		Minor: -1,
		Patch: -1,
	})
}

// GenerateManifestAfterVersion takes a version.Version argument and uses
// that to generate a manifest from the available migration Source's.
// It will loop through sources and build Migration's from the available
// migration files. Files labeled `schema` are considered the initial
// migration and get versioned as v0.0.0. All valid migrations that
// are versioned after the version given will be added to the manifest
// for migration. The final step is the build and sort the VersionTable
// that will be used for applying the migration in order by ExecuteMigrations.
func (s *Migrator) GenerateManifestAfterVersion(lastVersion version.Version) (Manifest, error) {
	const migrationSQLFilenameSuffix = ".sql"
	var manifest = NewManifest()

	// loop through sources
	for _, source := range s.Sources {
		if dirEntries, err := fs.ReadDir(source.FileSystem, source.Directory); err != nil {
			return manifest, err
		} else {
			// loop through file system entries
			for _, entry := range dirEntries {
				if !entry.IsDir() {
					filename := filepath.Join(source.Directory, entry.Name())
					basename := filepath.Base(filename)

					// create an entry, which may or may not be added (depending on filename semantics)
					var (
						validMigration = false
						migration      = Migration{
							Version:  version.Version{},
							Filename: filename,
							Source:   source.FileSystem,
						}
					)
					if basename == "schema"+migrationSQLFilenameSuffix {
						// will mark the base schema file as a valid v0.0.0 base migration
						validMigration = true
					} else if strings.HasPrefix(basename, version.Prefix) && strings.HasSuffix(basename, migrationSQLFilenameSuffix) {
						rawVersion := strings.TrimSuffix(basename, migrationSQLFilenameSuffix)
						if migrationVersion, err := version.Parse(rawVersion); err != nil {
							return manifest, err
						} else {
							// will mark the file as a valid versioned migration
							migration.Version = migrationVersion
							validMigration = true
						}
					}
					if validMigration && migration.Version.GreaterThan(lastVersion) {
						manifest.AddMigration(migration)
					}
				}
			}
		}
	}

	// sort the versions, so we can create the version table
	var versions = make([]string, 0, len(manifest.Migrations))
	for ver := range manifest.Migrations {
		versions = append(versions, ver)
	}
	sort.Slice(versions, func(a, b int) bool {
		return manifest.Migrations[versions[a]][0].Version.LessThan(manifest.Migrations[versions[b]][0].Version)
	})
	manifest.VersionTable = versions

	return manifest, nil
}
