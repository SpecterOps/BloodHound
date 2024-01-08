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

package tools

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/config"
	"net/http"
	"sync"
)

type MigratorState string

const (
	stateIdle      MigratorState = "idle"
	stateMigrating MigratorState = "migrating"
	stateCanceling MigratorState = "canceling"
)

func migrateTypes(ctx context.Context, neoDB, pgDB graph.Database) error {
	defer log.LogAndMeasure(log.LevelInfo, "Migrating kinds from Neo4j to PostgreSQL")()

	var (
		neoNodeKinds graph.Kinds
		neoEdgeKinds graph.Kinds
	)

	if err := neoDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			nextKindStr string
			result      = tx.Raw("call db.labels();", nil)
		)

		for result.Next() {
			if err := result.Scan(&nextKindStr); err != nil {
				return err
			}

			neoNodeKinds = append(neoNodeKinds, graph.StringKind(nextKindStr))
		}

		if err := result.Error(); err != nil {
			return err
		}

		result = tx.Raw("call db.relationshipTypes();", nil)

		for result.Next() {
			if err := result.Scan(&nextKindStr); err != nil {
				return err
			}

			neoEdgeKinds = append(neoEdgeKinds, graph.StringKind(nextKindStr))
		}

		return nil
	}); err != nil {
		return err
	}

	return pgDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		_, err := pgDB.(*pg.Driver).KindMapper().AssertKinds(tx, append(neoNodeKinds, neoEdgeKinds...))
		return err
	})
}

func convertNeo4jProperties(properties *graph.Properties) error {
	for key, propertyValue := range properties.Map {
		switch typedPropertyValue := propertyValue.(type) {
		case dbtype.Date:
			properties.Map[key] = typedPropertyValue.Time()

		case dbtype.Duration:
			return fmt.Errorf("unsupported conversion")

		case dbtype.Time:
			properties.Map[key] = typedPropertyValue.Time()

		case dbtype.LocalTime:
			properties.Map[key] = typedPropertyValue.Time()

		case dbtype.LocalDateTime:
			properties.Map[key] = typedPropertyValue.Time()
		}
	}

	return nil
}

func migrateNodes(ctx context.Context, neoDB, pgDB graph.Database) (map[graph.ID]graph.ID, error) {
	defer log.LogAndMeasure(log.LevelInfo, "Migrating nodes from Neo4j to PostgreSQL")()

	var (
		// Start at 2 and assume that the first node of the graph is the graph schema migration information
		nextNodeID     = graph.ID(2)
		nodeIDMappings = map[graph.ID]graph.ID{}
	)

	if err := neoDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			if err := pgDB.BatchOperation(ctx, func(tx graph.Batch) error {
				for next := range cursor.Chan() {
					if err := convertNeo4jProperties(next.Properties); err != nil {
						return err
					}

					if err := tx.CreateNode(graph.NewNode(nextNodeID, next.Properties, next.Kinds...)); err != nil {
						return err
					} else {
						nodeIDMappings[next.ID] = nextNodeID
						nextNodeID++
					}
				}

				return nil
			}); err != nil {
				return err
			}

			return cursor.Error()
		})
	}); err != nil {
		return nil, err
	}

	return nodeIDMappings, pgDB.Run(ctx, fmt.Sprintf(`alter sequence node_id_seq restart with %d`, nextNodeID), nil)
}

func migrateEdges(ctx context.Context, neoDB, pgDB graph.Database, nodeIDMappings map[graph.ID]graph.ID) error {
	defer log.LogAndMeasure(log.LevelInfo, "Migrating edges from Neo4j to PostgreSQL")()

	return neoDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Relationships().Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			if err := pgDB.BatchOperation(ctx, func(tx graph.Batch) error {
				for next := range cursor.Chan() {
					var (
						pgStartID = nodeIDMappings[next.StartID]
						pgEndID   = nodeIDMappings[next.EndID]
					)

					if err := convertNeo4jProperties(next.Properties); err != nil {
						return err
					}

					if err := tx.CreateRelationship(&graph.Relationship{
						StartID:    pgStartID,
						EndID:      pgEndID,
						Kind:       next.Kind,
						Properties: next.Properties,
					}); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return err
			}

			return cursor.Error()
		})
	})
}

type PGMigrator struct {
	graphSchema         graph.Schema
	graphDBSwitch       *graph.DatabaseSwitch
	serverCtx           context.Context
	migrationCancelFunc func()
	state               MigratorState
	lock                *sync.Mutex
	cfg                 config.Configuration
}

func NewPGMigrator(serverCtx context.Context, cfg config.Configuration, graphSchema graph.Schema, graphDBSwitch *graph.DatabaseSwitch) *PGMigrator {
	return &PGMigrator{
		graphSchema:   graphSchema,
		graphDBSwitch: graphDBSwitch,
		serverCtx:     serverCtx,
		state:         stateIdle,
		lock:          &sync.Mutex{},
		cfg:           cfg,
	}
}

func (s *PGMigrator) advanceState(next MigratorState, validTransitions ...MigratorState) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	isValid := false

	for _, validTransition := range validTransitions {
		if s.state == validTransition {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("migrator state is %s but expected one of: %v", s.state, validTransitions)
	}

	s.state = next
	return nil
}

func (s *PGMigrator) SwitchPostgreSQL(response http.ResponseWriter, request *http.Request) {
	if pgDB, err := dawgs.Open(s.serverCtx, pg.DriverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            s.cfg.Database.PostgreSQLConnectionString(),
	}); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed connecting to PostgreSQL: %w", err),
		}, http.StatusInternalServerError, response)
	} else if err := SetGraphDriver(request.Context(), s.cfg, pg.DriverName); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed updating graph database driver preferences: %w", err),
		}, http.StatusInternalServerError, response)
	} else {
		s.graphDBSwitch.Switch(pgDB)
		response.WriteHeader(http.StatusOK)

		log.Infof("Updated default graph driver to PostgreSQL")
	}
}

func (s *PGMigrator) SwitchNeo4j(response http.ResponseWriter, request *http.Request) {
	if neo4jDB, err := dawgs.Open(s.serverCtx, neo4j.DriverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            s.cfg.Neo4J.Neo4jConnectionString(),
	}); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed connecting to Neo4j: %w", err),
		}, http.StatusInternalServerError, response)
	} else if err := SetGraphDriver(request.Context(), s.cfg, neo4j.DriverName); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed updating graph database driver preferences: %w", err),
		}, http.StatusInternalServerError, response)
	} else {
		s.graphDBSwitch.Switch(neo4jDB)
		response.WriteHeader(http.StatusOK)

		log.Infof("Updated default graph driver to Neo4j")
	}
}

func (s *PGMigrator) startMigration() error {
	if err := s.advanceState(stateMigrating, stateIdle); err != nil {
		return fmt.Errorf("database migration state error: %w", err)
	} else if neo4jDB, err := dawgs.Open(s.serverCtx, neo4j.DriverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            s.cfg.Neo4J.Neo4jConnectionString(),
	}); err != nil {
		return fmt.Errorf("failed connecting to Neo4j: %w", err)
	} else if pgDB, err := dawgs.Open(s.serverCtx, pg.DriverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            s.cfg.Database.PostgreSQLConnectionString(),
	}); err != nil {
		return fmt.Errorf("failed connecting to PostgreSQL: %w", err)
	} else {
		log.Infof("Dispatching live migration from Neo4j to PostgreSQL")

		migrationCtx, migrationCancelFunc := context.WithCancel(s.serverCtx)
		s.migrationCancelFunc = migrationCancelFunc

		go func(ctx context.Context) {
			defer migrationCancelFunc()

			log.Infof("Starting live migration from Neo4j to PostgreSQL")

			if err := pgDB.AssertSchema(ctx, s.graphSchema); err != nil {
				log.Errorf("Unable to assert graph schema in PostgreSQL: %v", err)
			} else if err := migrateTypes(ctx, neo4jDB, pgDB); err != nil {
				log.Errorf("Unable to migrate Neo4j kinds to PostgreSQL: %v", err)
			} else if nodeIDMappings, err := migrateNodes(ctx, neo4jDB, pgDB); err != nil {
				log.Errorf("Failed importing nodes into PostgreSQL: %v", err)
			} else if err := migrateEdges(ctx, neo4jDB, pgDB, nodeIDMappings); err != nil {
				log.Errorf("Failed importing edges into PostgreSQL: %v", err)
			} else {
				log.Infof("Migration to PostgreSQL completed successfully")
			}

			if err := s.advanceState(stateIdle, stateMigrating, stateCanceling); err != nil {
				log.Errorf("Database migration state management error: %v", err)
			}
		}(migrationCtx)
	}

	return nil
}

func (s *PGMigrator) MigrationStart(response http.ResponseWriter, request *http.Request) {
	if err := s.startMigration(); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": err.Error(),
		}, http.StatusInternalServerError, response)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func (s *PGMigrator) cancelMigration() error {
	if err := s.advanceState(stateCanceling, stateMigrating); err != nil {
		return err
	}

	s.migrationCancelFunc()

	return nil
}

func (s *PGMigrator) MigrationCancel(response http.ResponseWriter, request *http.Request) {
	if err := s.cancelMigration(); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": err.Error(),
		}, http.StatusInternalServerError, response)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func (s *PGMigrator) MigrationStatus(response http.ResponseWriter, request *http.Request) {
	api.WriteJSONResponse(request.Context(), map[string]any{
		"state": s.state,
	}, http.StatusOK, response)
}
