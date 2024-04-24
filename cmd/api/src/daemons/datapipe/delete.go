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
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/util/channels"
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
