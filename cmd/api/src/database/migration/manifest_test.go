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

// TestManifest_AddMigration tests that migrations are added to the correct location
// within the migration map and in the correct order.
func TestManifest_AddMigration(t *testing.T) {
	var (
		migration1 = migration.Migration{Filename: "/migration/1", Version: version.Version{Major: 0, Minor: 1, Patch: 0}}
		migration2 = migration.Migration{Filename: "/migration/2", Version: version.Version{Major: 1, Minor: 0, Patch: 0}}
		migration3 = migration.Migration{Filename: "/migration/3", Version: version.Version{Major: 1, Minor: 0, Patch: 0}}
		manifest   = migration.Manifest{}
	)

	t.Run("No migrations", func(t *testing.T) {
		assert.Empty(t, manifest.Migrations)
	})

	t.Run("One v0.1.0 migration", func(t *testing.T) {
		manifest.AddMigration(migration1)
		assert.Equal(t, 1, len(manifest.Migrations["v0.1.0"]))
		assert.Equal(t, "/migration/1", manifest.Migrations["v0.1.0"][0].Filename)
	})

	t.Run("One v1.0.0 migration", func(t *testing.T) {
		manifest.AddMigration(migration2)
		assert.Equal(t, 1, len(manifest.Migrations["v1.0.0"]))
		assert.Equal(t, "/migration/2", manifest.Migrations["v1.0.0"][0].Filename)
	})

	t.Run("Two migrations in order", func(t *testing.T) {
		manifest.AddMigration(migration3)
		assert.Equal(t, 2, len(manifest.Migrations["v1.0.0"]))
		assert.Equal(t, "/migration/3", manifest.Migrations["v1.0.0"][1].Filename)
	})
}

// TestMigrator_GenerateManifest tests the full integrity of a Manifest generated
// from 3 different Source's:
// 1) Correct number of total migrations detected.
// 2) Correct number of versions detected.
// 3) The VersionTable has all versions, in order.
// 4) All versions have the correct number of migrations.
// 5) Schema files generate v0.0.0 migrations.
// 6) Migrations are ordered within each version, in the order the Source was defined.
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

	t.Run("Basic manifest integrity", func(t *testing.T) {
		// test for correct number of total migrations
		totalMigrations = getTotalMigrations(manifest)
		assert.Equal(t, 14, totalMigrations)

		// test for correct number of versions
		assert.Equal(t, len(correctOrder), len(manifest.Migrations))

		// test that versions are ordered correctly in the version table
		for i := range correctOrder {
			assert.Equal(t, correctOrder[i], manifest.VersionTable[i])
		}
	})

	t.Run("Number migrations per version", func(t *testing.T) {
		// test that all versions have correct number of migrations
		assert.Equal(t, 2, len(manifest.Migrations["v0.0.0"]))
		assert.Equal(t, 2, len(manifest.Migrations["v0.0.1"]))
		assert.Equal(t, 1, len(manifest.Migrations["v0.0.10"]))
		assert.Equal(t, 1, len(manifest.Migrations["v0.1.0"]))
		assert.Equal(t, 3, len(manifest.Migrations["v0.1.1"]))
		assert.Equal(t, 2, len(manifest.Migrations["v0.10.0"]))
		assert.Equal(t, 3, len(manifest.Migrations["v1.0.0"]))
	})

	t.Run("Schema is v0.0.0", func(t *testing.T) {
		// make sure schema is v0.0.0
		assert.Equal(t, "test_manifest/system1/schema.sql", manifest.Migrations["v0.0.0"][0].Filename)
		assert.Equal(t, "test_manifest/system2/schema.sql", manifest.Migrations["v0.0.0"][1].Filename)
	})

	t.Run("Source order", func(t *testing.T) {
		// spot check migration ordering within versions
		assert.Equal(t, "test_manifest/system1/v0.1.1.sql", manifest.Migrations["v0.1.1"][0].Filename)
		assert.Equal(t, "test_manifest/system2/v0.1.1.sql", manifest.Migrations["v0.1.1"][1].Filename)
		assert.Equal(t, "test_manifest/system3/v0.1.1.sql", manifest.Migrations["v0.1.1"][2].Filename)

		assert.Equal(t, "test_manifest/system1/v1.0.0.sql", manifest.Migrations["v1.0.0"][0].Filename)
		assert.Equal(t, "test_manifest/system2/v1.0.0.sql", manifest.Migrations["v1.0.0"][1].Filename)
		assert.Equal(t, "test_manifest/system3/v1.0.0.sql", manifest.Migrations["v1.0.0"][2].Filename)
	})
}

// TestMigrator_GenerateManifestAfterVersion builds upon the GenerateManifest test
// by starting from different version points and spot checking migration counts.
func TestMigrator_GenerateManifestAfterVersion(t *testing.T) {
	var (
		totalMigrations int
		migrator        = migration.Migrator{Sources: []migration.Source{
			{FileSystem: testManifestSystem1, Directory: "test_manifest/system1"},
			{FileSystem: testManifestSystem2, Directory: "test_manifest/system2"},
			{FileSystem: testManifestSystem3, Directory: "test_manifest/system3"},
		}}
	)

	t.Run("After schema", func(t *testing.T) {
		// should contain all but schema (v0.0.0) migrations
		manifest, err := migrator.GenerateManifestAfterVersion(version.Version{Major: 0, Minor: 0, Patch: 0})
		assert.Nil(t, err)
		totalMigrations = getTotalMigrations(manifest)
		assert.Equal(t, 12, totalMigrations)
	})

	t.Run("Latest only", func(t *testing.T) {
		// should only have v1.0.0 migrations
		manifest, err := migrator.GenerateManifestAfterVersion(version.Version{Major: 0, Minor: 10, Patch: 0})
		assert.Nil(t, err)
		totalMigrations = getTotalMigrations(manifest)
		assert.Equal(t, 3, totalMigrations)
	})
}

// getTotalMigrations is a helper function that counts all migrations
// within a manifest
func getTotalMigrations(manifest migration.Manifest) int {
	var total int
	for _, migrations := range manifest.Migrations {
		total += len(migrations)
	}
	return total
}
