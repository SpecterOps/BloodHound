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
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/version"
)

type Migration struct {
	Version version.Version
	Execute func(db graph.Database) error
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
	if err := s.executeStepwiseMigrations(); err != nil {
		return err
	}

	return nil
}

func (s *GraphMigrator) createMigrationData() error {
	return s.db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		if _, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"Major": 0,
			"Minor": 0,
			"Patch": 0,
		}), common.MigrationData); err != nil {
			return fmt.Errorf("could not create migration data: %w", err)
		}
		return nil
	})
}

func (s *GraphMigrator) updateMigrationData(target version.Version) error {
	var node *graph.Node

	if err := s.db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
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

		return s.db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			if err := tx.UpdateNode(node); err != nil {
				return fmt.Errorf("could not update migration data node: %w", err)
			}
			return nil
		})
	}
}

func (s *GraphMigrator) getMigrationData() (version.Version, error) {
	var (
		node             *graph.Node
		currentMigration version.Version
	)

	if err := s.db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		var err error

		if node, err = tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), common.MigrationData)
		}).First(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return currentMigration, err
	} else if currentMigration.Major, err = node.Properties.Get("Major").Int(); err != nil {
		return currentMigration, fmt.Errorf("unable to get Major property from node: %w", err)
	} else if currentMigration.Minor, err = node.Properties.Get("Minor").Int(); err != nil {
		return currentMigration, fmt.Errorf("unable to get Major property from node: %w", err)
	} else if currentMigration.Patch, err = node.Properties.Get("Patch").Int(); err != nil {
		return currentMigration, fmt.Errorf("unable to get Major property from node: %w", err)
	} else {
		return currentMigration, nil
	}
}

func (s *GraphMigrator) executeMigrations(target version.Version) error {
	mostRecentMigration := target

	for _, migration := range Manifest {
		if migration.Version.GreaterThan(mostRecentMigration) {
			log.Infof("GraphDB Version %s is greater than %s", migration.Version, mostRecentMigration)

			if err := migration.Execute(s.db); err != nil {
				return fmt.Errorf("migration version %s failed: %w", migration.Version.String(), err)
			}

			mostRecentMigration = migration.Version
		}
	}

	if mostRecentMigration.GreaterThan(target) {
		return s.updateMigrationData(mostRecentMigration)
	}

	return nil
}

func (s *GraphMigrator) executeStepwiseMigrations() error {
	if err := s.db.AssertSchema(context.Background(), graphschema.DefaultGraphSchema()); err != nil {
		return fmt.Errorf("error asserting current schema: %w", err)
	}

	if currentMigration, err := s.getMigrationData(); err != nil {
		if graph.IsErrNotFound(err) {
			if err := s.createMigrationData(); err != nil {
				return fmt.Errorf("could not create graph db migration data: %w", err)
			}

			currentVersion := version.GetVersion()

			log.Infof("This is a new graph database. Creating a migration entry for GraphDB version %s", currentVersion)
			return s.updateMigrationData(currentVersion)
		} else {
			return fmt.Errorf("unable to get graph db migration data: %w", err)
		}
	} else {
		return s.executeMigrations(currentMigration)
	}
}
