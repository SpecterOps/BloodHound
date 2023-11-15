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

package migration_test

import (
	"embed"
	"testing"

	"github.com/specterops/bloodhound/src/database/migration"
	"github.com/specterops/bloodhound/src/version"
	"github.com/stretchr/testify/assert"
)

//go:embed test_manifest/system1
var testManifestSystem1 embed.FS

//go:embed test_manifest/system2
var testManifestSystem2 embed.FS

//go:embed test_manifest/system3
var testManifestSystem3 embed.FS

func TestNewManifest(t *testing.T) {
	var manifest = migration.NewManifest()
	assert.NotNil(t, manifest.Migrations)
}

func TestManifest_AddMigration(t *testing.T) {
	var (
		migration1 = migration.Migration{Filename: "/migration/1", Version: version.Version{Major: 0, Minor: 1, Patch: 0}}
		migration2 = migration.Migration{Filename: "/migration/2", Version: version.Version{Major: 1, Minor: 0, Patch: 0}}
		migration3 = migration.Migration{Filename: "/migration/3", Version: version.Version{Major: 1, Minor: 0, Patch: 0}}
		manifest   = migration.Manifest{}
	)

	// no migrations
	assert.Empty(t, manifest.Migrations)

	// 1 migration under v0.1.0
	manifest.AddMigration(migration1)
	assert.Equal(t, 1, len(manifest.Migrations["v0.1.0"]))
	assert.Equal(t, "/migration/1", manifest.Migrations["v0.1.0"][0].Filename)

	// 1 migration under v1.0.0
	manifest.AddMigration(migration2)
	assert.Equal(t, 1, len(manifest.Migrations["v1.0.0"]))
	assert.Equal(t, "/migration/2", manifest.Migrations["v1.0.0"][0].Filename)

	// 2 migrations under v1.0.0, in order
	manifest.AddMigration(migration3)
	assert.Equal(t, 2, len(manifest.Migrations["v1.0.0"]))
	assert.Equal(t, "/migration/3", manifest.Migrations["v1.0.0"][1].Filename)
}

func TestMigrator_GenerateManifest(t *testing.T) {
	var (
		totalMigrations int
		migrator        = migration.Migrator{Sources: []migration.Source{
			{FileSystem: testManifestSystem1, Directory: "test_manifest/system1"},
			{FileSystem: testManifestSystem2, Directory: "test_manifest/system2"},
			{FileSystem: testManifestSystem3, Directory: "test_manifest/system3"},
		}}
		correctOrder = []string{"v0.0.0", "v0.0.1", "v0.0.10", "v0.1.0", "v0.1.1", "v0.10.0", "v1.0.0"}
	)

	manifest, err := migrator.GenerateManifest()
	assert.Nil(t, err)
	totalMigrations = getTotalMigrations(manifest)
	assert.Equal(t, 14, totalMigrations)

	// test we have the correct number of versions
	assert.Equal(t, len(correctOrder), len(manifest.Migrations))

	// test that versions are ordered correctly in the version table
	for i := range correctOrder {
		assert.Equal(t, correctOrder[i], manifest.VersionTable[i])
	}

	// test that all versions have correct number of migrations
	assert.Equal(t, 2, len(manifest.Migrations["v0.0.0"]))
	assert.Equal(t, 2, len(manifest.Migrations["v0.0.1"]))
	assert.Equal(t, 1, len(manifest.Migrations["v0.0.10"]))
	assert.Equal(t, 1, len(manifest.Migrations["v0.1.0"]))
	assert.Equal(t, 3, len(manifest.Migrations["v0.1.1"]))
	assert.Equal(t, 2, len(manifest.Migrations["v0.10.0"]))
	assert.Equal(t, 3, len(manifest.Migrations["v1.0.0"]))

	// make sure schema is v0.0.0
	assert.Equal(t, "test_manifest/system1/schema.sql", manifest.Migrations["v0.0.0"][0].Filename)
	assert.Equal(t, "test_manifest/system2/schema.sql", manifest.Migrations["v0.0.0"][1].Filename)

	// spot check system ordering within versions
	assert.Equal(t, "test_manifest/system1/v0.1.1.sql", manifest.Migrations["v0.1.1"][0].Filename)
	assert.Equal(t, "test_manifest/system2/v0.1.1.sql", manifest.Migrations["v0.1.1"][1].Filename)
	assert.Equal(t, "test_manifest/system3/v0.1.1.sql", manifest.Migrations["v0.1.1"][2].Filename)

	assert.Equal(t, "test_manifest/system1/v1.0.0.sql", manifest.Migrations["v1.0.0"][0].Filename)
	assert.Equal(t, "test_manifest/system2/v1.0.0.sql", manifest.Migrations["v1.0.0"][1].Filename)
	assert.Equal(t, "test_manifest/system3/v1.0.0.sql", manifest.Migrations["v1.0.0"][2].Filename)
}

func TestMigrator_GenerateManifestAfterVersion(t *testing.T) {
	var (
		totalMigrations int
		migrator        = migration.Migrator{Sources: []migration.Source{
			{FileSystem: testManifestSystem1, Directory: "test_manifest/system1"},
			{FileSystem: testManifestSystem2, Directory: "test_manifest/system2"},
			{FileSystem: testManifestSystem3, Directory: "test_manifest/system3"},
		}}
	)

	// should contain all but v0.0.0 migrations
	manifest, err := migrator.GenerateManifestAfterVersion(version.Version{Major: 0, Minor: 0, Patch: 0})
	assert.Nil(t, err)
	totalMigrations = getTotalMigrations(manifest)
	assert.Equal(t, 12, totalMigrations)

	// should only have v1.0.0 migrations
	manifest, err = migrator.GenerateManifestAfterVersion(version.Version{Major: 0, Minor: 10, Patch: 0})
	assert.Nil(t, err)
	totalMigrations = getTotalMigrations(manifest)
	assert.Equal(t, 3, totalMigrations)
}

func getTotalMigrations(manifest migration.Manifest) int {
	var total int
	for _, migrations := range manifest.Migrations {
		total += len(migrations)
	}
	return total
}
