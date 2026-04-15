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
	"fmt"
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func DeleteCollectedGraphData(ctx context.Context, graphDB graph.Database, deleteRequest model.AnalysisRequest, sourceKinds graph.Kinds) error {
	slog.InfoContext(
		ctx,
		"DeleteCollectedGraphData",
		slog.Bool("delete_all_data", deleteRequest.DeleteAllGraph),
		slog.Bool("delete_sourceless_data", deleteRequest.DeleteSourcelessGraph),
		slog.String("delete_source_kinds", strings.Join(deleteRequest.DeleteSourceKinds, ",")),
		slog.String("delete_relationships", strings.Join(deleteRequest.DeleteRelationships, ",")),
	)

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
