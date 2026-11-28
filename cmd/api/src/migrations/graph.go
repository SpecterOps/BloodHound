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

package migrations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/version"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

type Migration struct {
	Version version.Version
	Execute func(ctx context.Context, db graph.Database) error
}

type MigrationsByVersion = map[version.Version][]Migration

const (
	migrationDataMajorProperty      = "Major"
	migrationDataMinorProperty      = "Minor"
	migrationDataPatchProperty      = "Patch"
	migrationDataPrereleaseProperty = "Prerelease"
)

// NewMigrationDataProperties returns the graph properties that record the given version on the MigrationData node.
func NewMigrationDataProperties(migrationVersion version.Version) *graph.Properties {
	return graph.AsProperties(map[string]any{
		migrationDataMajorProperty:      migrationVersion.Major,
		migrationDataMinorProperty:      migrationVersion.Minor,
		migrationDataPatchProperty:      migrationVersion.Patch,
		migrationDataPrereleaseProperty: migrationVersion.Prerelease,
	})
}

// VersionRequiresMigration returns true if the running version is newer than the recorded graph migration version, or
// if both share a major, minor and patch version but differ in prerelease. Prerelease builds (for example release
// candidates) share a version with their eventual release, so a prerelease change must still trigger a graph migration
// to re-assert the schema.
func VersionRequiresMigration(runningVersion, recordedVersion version.Version) bool {
	return runningVersion.GreaterThan(recordedVersion) ||
		(runningVersion.Equals(recordedVersion) && runningVersion.Prerelease != recordedVersion.Prerelease)
}

func UpdateMigrationData(ctx context.Context, db graph.Database, target version.Version) error {
	var node *graph.Node

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error

		if node, err = tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), common.MigrationData)
		}).First(); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("could not query exisiting migration data: %w", err)
	} else {
		node.Properties.Set(migrationDataMajorProperty, target.Major)
		node.Properties.Set(migrationDataMinorProperty, target.Minor)
		node.Properties.Set(migrationDataPatchProperty, target.Patch)
		node.Properties.Set(migrationDataPrereleaseProperty, target.Prerelease)

		return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
			if err := tx.UpdateNode(node); err != nil {
				return fmt.Errorf("could not update migration data node: %w", err)
			}

			return nil
		})
	}
}

var ErrNoMigrationData = errors.New("no migration data")

// GetMigrationData fetches the database migration version for the given graph. This function logs failures but does
// not return the raw error condition to the caller, instead the sentinel ErrNoMigrationData is returned. This is done
// to avoid situations where a version check prevents an otherwise uninitialized database from reaching schema
// assertion.
func GetMigrationData(ctx context.Context, db graph.Database) (version.Version, error) {
	var (
		node             *graph.Node
		currentMigration = version.GetVersion()
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error

		node, err = tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), common.MigrationData)
		}).First()

		return err
	}); err != nil {
		slog.WarnContext(
			ctx,
			"Unable to fetch migration data from graph",
			attr.Error(err),
		)
		return currentMigration, ErrNoMigrationData
	} else if major, err := node.Properties.Get(migrationDataMajorProperty).Int(); err != nil {
		slog.WarnContext(
			ctx,
			"Unable to get Major property from migration data node",
			attr.Error(err),
		)
		return currentMigration, ErrNoMigrationData
	} else if minor, err := node.Properties.Get(migrationDataMinorProperty).Int(); err != nil {
		slog.WarnContext(
			ctx,
			"Unable to get Minor property from migration data node",
			attr.Error(err),
		)
		return currentMigration, ErrNoMigrationData
	} else if patch, err := node.Properties.Get(migrationDataPatchProperty).Int(); err != nil {
		slog.WarnContext(
			ctx,
			"Unable to get Patch property from migration data node",
			attr.Error(err),
		)
		return currentMigration, ErrNoMigrationData
	} else {
		currentMigration.Major = major
		currentMigration.Minor = minor
		currentMigration.Patch = patch

		// Migration data written before prerelease tracking has no Prerelease property; treat it as a final release
		if prerelease, err := node.Properties.Get(migrationDataPrereleaseProperty).String(); err != nil {
			currentMigration.Prerelease = ""
		} else {
			currentMigration.Prerelease = prerelease
		}
	}

	return currentMigration, nil
}

type GraphMigration interface {
	Migrate(ctx context.Context) error
	GetMigrationsByVersion() MigrationsByVersion
	ExecuteStepwiseMigrations(ctx context.Context, migrationsByVersion MigrationsByVersion) error

	executeMigrations(ctx context.Context, originalVersion version.Version, migrationsByVersion MigrationsByVersion) error
}

// schemalessNodeKindBackfillData is the narrow interface consumed by Version_930_Migration to read
// display metadata from Postgres and write backfilled custom_node_kinds entries. It is intended to
// be satisfied by database.Database
type schemalessNodeKindBackfillData interface {
	GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error)
	GetGraphSchemaNodeKinds(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaNodeKinds, int, error)
	CreateCustomNodeKinds(ctx context.Context, customNodeKinds model.CustomNodeKinds) (model.CustomNodeKinds, error)
}

type GraphMigrator struct {
	db                             graph.Database
	schemalessNodeKindBackfillData schemalessNodeKindBackfillData
}

// Option configures optional dependencies on a GraphMigrator. Most migrations
// only need a graph.Database; a small number need access to the relational
// database (e.g. to read source_kinds). Use functional options to inject those
// dependencies without forcing every caller to acknowledge them.
type Option func(*GraphMigrator)

// WithSchemalessKindBackfill wires a SchemalessNodeKindBackfillData provider into the migrator.
// Migrations that backfill custom_node_kinds for orphaned schemaless ingest kinds will
// short-circuit when this option is not supplied.
func WithSchemalessKindBackfill(nodeKindData schemalessNodeKindBackfillData) Option {
	return func(m *GraphMigrator) { m.schemalessNodeKindBackfillData = nodeKindData }
}

func NewGraphMigrator(db graph.Database, opts ...Option) *GraphMigrator {
	migrator := &GraphMigrator{db: db}
	for _, opt := range opts {
		opt(migrator)
	}
	return migrator
}

func (s *GraphMigrator) Migrate(ctx context.Context) error {
	// Assert the schema first
	if err := s.db.AssertSchema(ctx, graphschema.DefaultGraphSchema()); err != nil {
		return err
	}

	// Perform stepwise migrations
	if err := s.ExecuteStepwiseMigrations(ctx, s.GetMigrationsByVersion()); err != nil {
		return err
	}

	return nil
}

func CreateMigrationData(ctx context.Context, db graph.Database, currentVersion version.Version) error {
	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if _, err := tx.CreateNode(NewMigrationDataProperties(currentVersion), common.MigrationData); err != nil {
			return fmt.Errorf("could not create migration data: %w", err)
		}
		return nil
	})
}

func (s *GraphMigrator) executeMigrations(ctx context.Context, originalVersion version.Version, migrationsByVersion MigrationsByVersion) error {
	var (
		mostRecentVersion = originalVersion
		orderedVersions   []version.Version
	)

	// Order the manifest versions to ensure ascending stepwise migrations
	for v := range migrationsByVersion {
		orderedVersions = append(orderedVersions, v)
	}
	sort.Slice(orderedVersions, func(i, j int) bool {
		return orderedVersions[i].LessThan(orderedVersions[j])
	})

	for _, nextVersion := range orderedVersions {
		if nextVersion.GreaterThan(mostRecentVersion) {
			versionMigrations := migrationsByVersion[nextVersion]
			for _, nextMigration := range versionMigrations {
				slog.InfoContext(
					ctx,
					"Graph migration version is greater than current version",
					slog.String("migration_version", nextMigration.Version.String()),
					slog.String("most_recent_version", mostRecentVersion.String()),
				)

				if err := nextMigration.Execute(ctx, s.db); err != nil {
					return fmt.Errorf("migration version %s failed: %w", nextMigration.Version.String(), err)
				}
			}

			slog.InfoContext(
				ctx,
				"Graph migration executed successfully",
				slog.String("next_version", nextVersion.String()),
			)
			mostRecentVersion = nextVersion
		}
	}

	// Record the running release version, including its prerelease, when it is at or beyond the most recent migration
	if releaseVersion := version.GetVersion(); !releaseVersion.LessThan(mostRecentVersion) {
		mostRecentVersion = releaseVersion
	}

	if VersionRequiresMigration(mostRecentVersion, originalVersion) {
		return UpdateMigrationData(ctx, s.db, mostRecentVersion)
	}

	return nil
}

func (s *GraphMigrator) GetMigrationsByVersion() MigrationsByVersion {
	migrationsByVersion := make(MigrationsByVersion)
	for _, migration := range GetManifest(s.schemalessNodeKindBackfillData) {
		migrationsByVersion[migration.Version] = append(migrationsByVersion[migration.Version], migration)
	}
	return migrationsByVersion
}

func (s *GraphMigrator) ExecuteStepwiseMigrations(ctx context.Context, migrationsByVersion MigrationsByVersion) error {
	if currentMigration, err := GetMigrationData(ctx, s.db); err != nil {
		if errors.Is(err, ErrNoMigrationData) {
			currentVersion := version.GetVersion()

			slog.InfoContext(
				ctx,
				"This is a new graph database. Creating a migration entry for GraphDB version",
				slog.String("current_version", currentVersion.String()),
			)
			return CreateMigrationData(ctx, s.db, currentMigration)
		} else {
			return fmt.Errorf("unable to get graph db migration data: %w", err)
		}
	} else {
		return s.executeMigrations(ctx, currentMigration, migrationsByVersion)
	}
}
