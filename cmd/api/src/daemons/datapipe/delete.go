// Copyright 2024 Specter Ops, Inc.
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

package datapipe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

// graphWiper is implemented by graph database drivers (currently PostgreSQL) that support a fast, bulk truncate-based
// wipe of all graph data. The retain delegate runs within the same transaction as the wipe so survivor nodes can be
// recreated atomically.
type graphWiper interface {
	WipeGraph(ctx context.Context, retain graph.TransactionDelegate) error
}

func DeleteCollectedGraphData(ctx context.Context, graphDB graph.Database, deleteRequest model.AnalysisRequest, sourceKinds graph.Kinds) error {
	slog.InfoContext(
		ctx,
		"DeleteCollectedGraphData",
		slog.Bool("delete_all_data", deleteRequest.DeleteAllGraph),
		slog.Bool("delete_sourceless_data", deleteRequest.DeleteSourcelessGraph),
		slog.String("delete_source_kinds", strings.Join(deleteRequest.DeleteSourceKinds, ",")),
		slog.String("delete_relationships", strings.Join(deleteRequest.DeleteRelationships, ",")),
	)

	// On backends that support a bulk wipe (PostgreSQL), a full graph deletion truncates every node and edge partition
	// in a single transaction rather than streaming and deleting each node row-by-row. This also clears all edges, so
	// any requested relationship deletions are subsumed by the wipe. Backends without this capability (Neo4j) fall
	// through to the row-by-row path below.
	if deleteRequest.DeleteAllGraph {
		if wiper, isWiper := graph.AsDriver[graphWiper](graphDB); isWiper {
			return wipeAllGraphData(ctx, graphDB, wiper)
		}
	}

	if deleteRequest.DeleteAllGraph || deleteRequest.DeleteSourcelessGraph || len(deleteRequest.DeleteSourceKinds) > 0 {
		nodeOperation := ops.StartNewOperation[graph.ID](ops.OperationContext{
			Parent:     ctx,
			DB:         graphDB,
			NumReaders: 1,
			NumWriters: 1,
		})

		deleteSourceKinds := make(graph.Kinds, len(deleteRequest.DeleteSourceKinds))
		for i, sourceKind := range deleteRequest.DeleteSourceKinds {
			deleteSourceKinds[i] = graph.StringKind(sourceKind)
		}

		nodeOperation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
			var (
				nodeQuery graph.NodeQuery
				filters   []graph.Criteria
			)

			// Always exclude MigrationData
			migrationFilter := query.Not(query.Kind(query.Node(), common.MigrationData))

			if !deleteRequest.DeleteAllGraph {
				if deleteRequest.DeleteSourcelessGraph {
					filters = append(filters,
						query.Not(query.KindIn(query.Node(), sourceKinds...)),
					)
				}

				if len(deleteSourceKinds) > 0 {
					filters = append(filters,
						query.KindIn(query.Node(), deleteSourceKinds...),
					)
				}
			}

			if len(filters) > 0 {
				nodeQuery = tx.Nodes().Filter(
					query.And(
						migrationFilter,
						query.Or(filters...),
					),
				)
			} else {
				nodeQuery = tx.Nodes().Filter(migrationFilter)
			}

			return nodeQuery.FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
				channels.PipeAll(ctx, cursor.Chan(), outC)
				return cursor.Error()
			})
		})

		nodeOperation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan graph.ID) error {
			for {
				if nextID, hasNextID := channels.Receive(ctx, inC); hasNextID {
					if err := batch.DeleteNode(nextID); err != nil {
						return err
					}
				} else {
					break
				}
			}

			return nil
		})

		if err := nodeOperation.Done(); err != nil {
			return fmt.Errorf("error deleting graph nodes: %w", err)
		}
	}

	if len(deleteRequest.DeleteRelationships) > 0 {
		edgeOperation := ops.StartNewOperation[graph.ID](ops.OperationContext{
			Parent:     ctx,
			DB:         graphDB,
			NumReaders: 1,
			NumWriters: 1,
		})

		deleteRelationshipKinds := make(graph.Kinds, len(deleteRequest.DeleteRelationships))
		for i, relationshipKind := range deleteRequest.DeleteRelationships {
			deleteRelationshipKinds[i] = graph.StringKind(relationshipKind)
		}

		edgeOperation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
			edgeQuery := tx.Relationships().Filter(query.KindIn(query.Relationship(), deleteRelationshipKinds...))

			return edgeQuery.FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
				channels.PipeAll(ctx, cursor.Chan(), outC)
				return cursor.Error()
			})
		})

		edgeOperation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan graph.ID) error {
			for {
				if nextID, hasNextID := channels.Receive(ctx, inC); hasNextID {
					if err := batch.DeleteRelationship(nextID); err != nil {
						return err
					}
				} else {
					break
				}
			}

			return nil
		})

		if err := edgeOperation.Done(); err != nil {
			return fmt.Errorf("error deleting graph edges: %w", err)
		}
	}

	return nil
}

// wipeAllGraphData performs a full graph wipe via the backend's bulk truncate path. The MigrationData node records the
// graph schema version and must survive the wipe, otherwise the migration system would re-initialize the database on the
// next startup. Its version is read before the wipe and the node is recreated within the same transaction as the
// truncate, so the wipe and recreate are atomic.
func wipeAllGraphData(ctx context.Context, graphDB graph.Database, wiper graphWiper) error {
	migrationData, err := migrations.GetMigrationData(ctx, graphDB)
	if err != nil && !errors.Is(err, migrations.ErrNoMigrationData) {
		return fmt.Errorf("reading migration data prior to graph wipe: %w", err)
	} else if errors.Is(err, migrations.ErrNoMigrationData) {
		// GetMigrationData returns the running binary version when no readable MigrationData node exists, which is the
		// correct value to recreate after the wipe.
		slog.WarnContext(
			ctx,
			"No readable MigrationData node found prior to graph wipe; recreating with current binary version",
			slog.String("version", migrationData.String()),
		)
	}

	return wiper.WipeGraph(ctx, func(tx graph.Transaction) error {
		if _, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"Major": migrationData.Major,
			"Minor": migrationData.Minor,
			"Patch": migrationData.Patch,
		}), common.MigrationData); err != nil {
			return fmt.Errorf("recreating migration data after graph wipe: %w", err)
		}

		return nil
	})
}
