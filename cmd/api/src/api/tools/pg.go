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
	"log/slog"
	"net/http"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/config"
)

type MigratorState string

const (
	StateIdle      MigratorState = "idle"
	StateMigrating MigratorState = "migrating"
	StateCanceling MigratorState = "canceling"
)

func migrateTypes(ctx context.Context, neoDB, pgDB graph.Database) error {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelInfo, "Migrating kinds from Neo4j to PostgreSQL")()

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

	_, err := pgDB.(*pg.Driver).KindMapper().AssertKinds(ctx, append(neoNodeKinds, neoEdgeKinds...))
	return err
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

func migrateNodesToNeo4j(ctx context.Context, neoDB, pgDB graph.Database) (map[graph.ID]graph.ID, error) {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelInfo, "Migrating nodes from PostgreSQL to Neo4j")()

	var (
		nodeBuffer     []*graph.Node
		nodeIDMappings = map[graph.ID]graph.ID{}
	)

	if err := pgDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for next := range cursor.Chan() {
				nodeBuffer = append(nodeBuffer, next)

				if len(nodeBuffer) > 2000 {
					if err := neoDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
						for _, nextNode := range nodeBuffer {
							if newNode, err := tx.CreateNode(nextNode.Properties, nextNode.Kinds...); err != nil {
								return err
							} else {
								nodeIDMappings[nextNode.ID] = newNode.ID
							}
						}

						return nil
					}); err != nil {
						return err
					}

					nodeBuffer = nodeBuffer[:0]
				}
			}

			if cursor.Error() == nil && len(nodeBuffer) > 0 {
				if err := neoDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
					for _, nextNode := range nodeBuffer {
						if newNode, err := tx.CreateNode(nextNode.Properties, nextNode.Kinds...); err != nil {
							return err
						} else {
							nodeIDMappings[nextNode.ID] = newNode.ID
						}
					}

					return nil
				}); err != nil {
					return err
				}
			}

			return cursor.Error()
		})
	}); err != nil {
		return nil, err
	}

	return nodeIDMappings, nil
}

func migrateNodesToPG(ctx context.Context, neoDB, pgDB graph.Database) (map[graph.ID]graph.ID, error) {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelInfo, "Migrating nodes from Neo4j to PostgreSQL")()

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

func migrateEdges(ctx context.Context, sourceDB, destinationDB graph.Database, nodeIDMappings map[graph.ID]graph.ID) error {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelInfo, "Migrating edges from Neo4j to PostgreSQL")()

	return sourceDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Relationships().Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			if err := destinationDB.BatchOperation(ctx, func(tx graph.Batch) error {
				for nextSourceEdge := range cursor.Chan() {
					var (
						dstStartID = nodeIDMappings[nextSourceEdge.StartID]
						dstEndID   = nodeIDMappings[nextSourceEdge.EndID]
					)

					if err := convertNeo4jProperties(nextSourceEdge.Properties); err != nil {
						return err
					}

					if err := tx.CreateRelationship(&graph.Relationship{
						StartID:    dstStartID,
						EndID:      dstEndID,
						Kind:       nextSourceEdge.Kind,
						Properties: nextSourceEdge.Properties,
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
	ServerCtx           context.Context
	migrationCancelFunc func()
	State               MigratorState
	lock                *sync.Mutex
	Cfg                 config.Configuration
}

func NewPGMigrator(serverCtx context.Context, cfg config.Configuration, graphSchema graph.Schema, graphDBSwitch *graph.DatabaseSwitch) *PGMigrator {
	return &PGMigrator{
		graphSchema:   graphSchema,
		graphDBSwitch: graphDBSwitch,
		ServerCtx:     serverCtx,
		State:         StateIdle,
		lock:          &sync.Mutex{},
		Cfg:           cfg,
	}
}

func (s *PGMigrator) advanceState(next MigratorState, validTransitions ...MigratorState) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	isValid := false

	for _, validTransition := range validTransitions {
		if s.State == validTransition {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("migrator state is %s but expected one of: %v", s.State, validTransitions)
	}

	s.State = next
	return nil
}

func (s *PGMigrator) SwitchPostgreSQL(response http.ResponseWriter, request *http.Request) {
	if pgDB, err := s.OpenPostgresGraphConnection(); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed connecting to PostgreSQL: %w", err),
		}, http.StatusInternalServerError, response)
	} else if err := pgDB.AssertSchema(request.Context(), s.graphSchema); err != nil {
		slog.ErrorContext(request.Context(), fmt.Sprintf("Unable to assert graph schema in PostgreSQL: %v", err))
	} else if err := SetGraphDriver(request.Context(), s.Cfg, pg.DriverName); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed updating graph database driver preferences: %w", err),
		}, http.StatusInternalServerError, response)
	} else {
		s.graphDBSwitch.Switch(pgDB)
		response.WriteHeader(http.StatusOK)

		slog.InfoContext(request.Context(), "Updated default graph driver to PostgreSQL")
	}
}

func (s *PGMigrator) SwitchNeo4j(response http.ResponseWriter, request *http.Request) {
	if neo4jDB, err := s.OpenNeo4jGraphConnection(); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed connecting to Neo4j: %w", err),
		}, http.StatusInternalServerError, response)
	} else if err := SetGraphDriver(request.Context(), s.Cfg, neo4j.DriverName); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": fmt.Errorf("failed updating graph database driver preferences: %w", err),
		}, http.StatusInternalServerError, response)
	} else {
		s.graphDBSwitch.Switch(neo4jDB)
		response.WriteHeader(http.StatusOK)

		slog.InfoContext(request.Context(), "Updated default graph driver to Neo4j")
	}
}

func (s *PGMigrator) StartMigrationToNeo() error {
	if err := s.advanceState(StateMigrating, StateIdle); err != nil {
		return fmt.Errorf("database migration state error: %w", err)
	} else if neo4jDB, err := s.OpenNeo4jGraphConnection(); err != nil {
		return fmt.Errorf("failed connecting to Neo4j: %w", err)
	} else if pgDB, err := s.OpenPostgresGraphConnection(); err != nil {
		return fmt.Errorf("failed connecting to PostgreSQL: %w", err)
	} else {
		slog.Info("Dispatching live migration from Neo4j to PostgreSQL")

		migrationCtx, migrationCancelFunc := context.WithCancel(s.ServerCtx)
		s.migrationCancelFunc = migrationCancelFunc

		go func(ctx context.Context) {
			defer migrationCancelFunc()

			slog.InfoContext(ctx, "Starting live migration from Neo4j to PostgreSQL")

			if err := neo4jDB.AssertSchema(ctx, s.graphSchema); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Unable to assert graph schema in PostgreSQL: %v", err))
			} else if nodeIDMappings, err := migrateNodesToNeo4j(ctx, neo4jDB, pgDB); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed importing nodes into PostgreSQL: %v", err))
			} else if err := migrateEdges(ctx, pgDB, neo4jDB, nodeIDMappings); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed importing edges into PostgreSQL: %v", err))
			} else {
				slog.InfoContext(ctx, "Migration to PostgreSQL completed successfully")
			}

			if err := s.advanceState(StateIdle, StateMigrating, StateCanceling); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Database migration state management error: %v", err))
			}
		}(migrationCtx)
	}

	return nil
}

func (s *PGMigrator) StartMigrationToPG() error {
	if err := s.advanceState(StateMigrating, StateIdle); err != nil {
		return fmt.Errorf("database migration state error: %w", err)
	} else if neo4jDB, err := s.OpenNeo4jGraphConnection(); err != nil {
		return fmt.Errorf("failed connecting to Neo4j: %w", err)
	} else if pgDB, err := s.OpenPostgresGraphConnection(); err != nil {
		return fmt.Errorf("failed connecting to PostgreSQL: %w", err)
	} else {
		slog.Info("Dispatching live migration from Neo4j to PostgreSQL")

		migrationCtx, migrationCancelFunc := context.WithCancel(s.ServerCtx)
		s.migrationCancelFunc = migrationCancelFunc

		go func(ctx context.Context) {
			defer migrationCancelFunc()

			slog.InfoContext(ctx, "Starting live migration from Neo4j to PostgreSQL")

			if err := pgDB.AssertSchema(ctx, s.graphSchema); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Unable to assert graph schema in PostgreSQL: %v", err))
			} else if err := migrateTypes(ctx, neo4jDB, pgDB); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Unable to migrate Neo4j kinds to PostgreSQL: %v", err))
			} else if nodeIDMappings, err := migrateNodesToPG(ctx, neo4jDB, pgDB); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed importing nodes into PostgreSQL: %v", err))
			} else if err := migrateEdges(ctx, neo4jDB, pgDB, nodeIDMappings); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed importing edges into PostgreSQL: %v", err))
			} else {
				slog.InfoContext(ctx, "Migration to PostgreSQL completed successfully")
			}

			if err := s.advanceState(StateIdle, StateMigrating, StateCanceling); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Database migration state management error: %v", err))
			}
		}(migrationCtx)
	}

	return nil
}

func (s *PGMigrator) MigrationStartPGToNeo(response http.ResponseWriter, request *http.Request) {
	if err := s.StartMigrationToNeo(); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": err.Error(),
		}, http.StatusInternalServerError, response)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func (s *PGMigrator) MigrationStartNeoToPG(response http.ResponseWriter, request *http.Request) {
	if err := s.StartMigrationToPG(); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": err.Error(),
		}, http.StatusInternalServerError, response)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func (s *PGMigrator) CancelMigration() error {
	if err := s.advanceState(StateCanceling, StateMigrating); err != nil {
		return err
	}

	s.migrationCancelFunc()

	return nil
}

func (s *PGMigrator) MigrationCancel(response http.ResponseWriter, request *http.Request) {
	if err := s.CancelMigration(); err != nil {
		api.WriteJSONResponse(request.Context(), map[string]any{
			"error": err.Error(),
		}, http.StatusInternalServerError, response)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func (s *PGMigrator) MigrationStatus(response http.ResponseWriter, request *http.Request) {
	api.WriteJSONResponse(request.Context(), map[string]any{
		"state": s.State,
	}, http.StatusOK, response)
}

func (s *PGMigrator) OpenPostgresGraphConnection() (graph.Database, error) {
	return dawgs.Open(s.ServerCtx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: size.Gibibyte,
		DriverCfg:             s.Cfg.Database.PostgreSQLConnectionString(),
	})
}

func (s *PGMigrator) OpenNeo4jGraphConnection() (graph.Database, error) {
	return dawgs.Open(s.ServerCtx, neo4j.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: size.Gibibyte,
		DriverCfg:             s.Cfg.Neo4J.Neo4jConnectionString(),
	})
}
