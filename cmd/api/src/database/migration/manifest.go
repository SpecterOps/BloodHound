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

type Manifest struct {
	VersionTable []string
	Migrations   map[string][]Migration
}

func NewManifest() Manifest {
	return Manifest{
		Migrations: make(map[string][]Migration),
	}
}

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

func (s *Migrator) GenerateManifest() (Manifest, error) {
	return s.GenerateManifestAfterVersion(version.Version{
		Major: -1,
		Minor: -1,
		Patch: -1,
	})
}

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
