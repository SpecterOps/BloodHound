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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func DeleteCollectedGraphData(ctx context.Context, graphDB graph.Database) error {
	operation := ops.StartNewOperation[graph.ID](ops.OperationContext{
		Parent:     ctx,
		DB:         graphDB,
		NumReaders: 1,
		NumWriters: 1,
	})

	operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan graph.ID) error {
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

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
		return tx.Relationships().FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			channels.PipeAll(ctx, cursor.Chan(), outC)
			return cursor.Error()
		})
	})

	if err := operation.Done(); err != nil {
		return fmt.Errorf("error deleting graph relationships: %w", err)
	}

	operation = ops.StartNewOperation[graph.ID](ops.OperationContext{
		Parent:     ctx,
		DB:         graphDB,
		NumReaders: 1,
		NumWriters: 1,
	})

	operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan graph.ID) error {
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

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
		return tx.Nodes().FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			channels.PipeAll(ctx, cursor.Chan(), outC)
			return cursor.Error()
		})
	})

	if err := operation.Done(); err != nil {
		return fmt.Errorf("error deleting graph nodes: %w", err)
	}

	return nil
}

// step 1. delete nodes. pg trigger func drop_node_edges() will drop edges automatically
// step 2. if source kinds requested for deletion, remove them from source_kinds pg table
func DeleteCollectedGraphData2(ctx context.Context, graphDB graph.Database, deleteRequest model.AnalysisRequest, sourceKinds graph.Kinds) error {
	operation := ops.StartNewOperation[graph.ID](ops.OperationContext{
		Parent:     ctx,
		DB:         graphDB,
		NumReaders: 1,
		NumWriters: 1,
	})

	// pretend for now that i have read the source kinds from the delete request.
	// TODO: parse this out of the deleteRequest. gorm nonsense
	// deleteSourceKinds := []graph.Kind{graph.StringKind("hellobase")}
	deleteSourceKinds := make(graph.Kinds, len(deleteRequest.DeleteSourceKinds))
	for i, s := range deleteRequest.DeleteSourceKinds {
		deleteSourceKinds[i] = graph.StringKind(s)
	}

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
		var nodeQuery graph.NodeQuery

		if deleteRequest.DeleteAllGraph {
			// no filter. fetch em all
			nodeQuery = tx.Nodes()
		} else if deleteRequest.DeleteAllOpenGraph {
			// fetch all nodes which don't have a source kind
			nodeQuery = tx.Nodes().Filter(
				query.Not(query.KindIn(query.Node(), sourceKinds...)),
			)
		} else {
			nodeQuery = tx.Nodes().Filter(
				query.KindIn(query.Node(), deleteSourceKinds...),
			)
		}

		return nodeQuery.FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			channels.PipeAll(ctx, cursor.Chan(), outC)
			return cursor.Error()
		})
	})

	operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan graph.ID) error {
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

	if err := operation.Done(); err != nil {
		return fmt.Errorf("error deleting graph nodes: %w", err)
	}

	return nil
}
