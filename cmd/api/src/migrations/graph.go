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

	"github.com/specterops/bloodhound/cmd/api/src/version"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

type Migration struct {
	Version version.Version
	Execute func(ctx context.Context, db graph.Database) error
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
		node.Properties.Set("Major", target.Major)
		node.Properties.Set("Minor", target.Minor)
		node.Properties.Set("Patch", target.Patch)

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
		slog.WarnContext(ctx, fmt.Sprintf("Unable to fetch migration data from graph: %v", err))
		return currentMigration, ErrNoMigrationData
	} else if major, err := node.Properties.Get("Major").Int(); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Unable to get Major property from migration data node: %v", err))
		return currentMigration, ErrNoMigrationData
	} else if minor, err := node.Properties.Get("Minor").Int(); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("unable to get Minor property from migration data node: %v", err))
		return currentMigration, ErrNoMigrationData
	} else if patch, err := node.Properties.Get("Patch").Int(); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("unable to get Patch property from migration data node: %v", err))
		return currentMigration, ErrNoMigrationData
	} else {
		currentMigration.Major = major
		currentMigration.Minor = minor
		currentMigration.Patch = patch
	}

	return currentMigration, nil
}

type GraphMigrator struct {
	db graph.Database
}

func NewGraphMigrator(db graph.Database) *GraphMigrator {
	return &GraphMigrator{db: db}
}

func (s *GraphMigrator) Migrate(ctx context.Context, schema graph.Schema) error {
	// Assert the schema first
	if err := s.db.AssertSchema(ctx, schema); err != nil {
		return err
	}

	// Perform stepwise migrations
	if err := s.executeStepwiseMigrations(ctx); err != nil {
		return err
	}

	return nil
}

func CreateMigrationData(ctx context.Context, db graph.Database, currentVersion version.Version) error {
	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if _, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"Major": currentVersion.Major,
			"Minor": currentVersion.Minor,
			"Patch": currentVersion.Patch,
		}), common.MigrationData); err != nil {
			return fmt.Errorf("could not create migration data: %w", err)
		}
		return nil
	})
}

func (s *GraphMigrator) executeMigrations(ctx context.Context, originalVersion version.Version) error {
	mostRecentVersion := originalVersion

	for _, nextMigration := range Manifest {
		if nextMigration.Version.GreaterThan(mostRecentVersion) {
			slog.InfoContext(ctx, fmt.Sprintf("Graph migration version %s is greater than current version %s", nextMigration.Version, mostRecentVersion))

			if err := nextMigration.Execute(ctx, s.db); err != nil {
				return fmt.Errorf("migration version %s failed: %w", nextMigration.Version.String(), err)
			}

			slog.InfoContext(ctx, fmt.Sprintf("Graph migration version %s executed successfully", nextMigration.Version))
			mostRecentVersion = nextMigration.Version
		}
	}

	if releaseVersion := version.GetVersion(); releaseVersion.GreaterThan(mostRecentVersion) {
		mostRecentVersion = releaseVersion
	}

	if mostRecentVersion.GreaterThan(originalVersion) {
		return UpdateMigrationData(ctx, s.db, mostRecentVersion)
	}

	return nil
}

func (s *GraphMigrator) executeStepwiseMigrations(ctx context.Context) error {
	if currentMigration, err := GetMigrationData(ctx, s.db); err != nil {
		if errors.Is(err, ErrNoMigrationData) {
			currentVersion := version.GetVersion()

			slog.InfoContext(ctx, fmt.Sprintf("This is a new graph database. Creating a migration entry for GraphDB version %s", currentVersion))
			return CreateMigrationData(ctx, s.db, currentMigration)
		} else {
			return fmt.Errorf("unable to get graph db migration data: %w", err)
		}
	} else {
		return s.executeMigrations(ctx, currentMigration)
	}
}
