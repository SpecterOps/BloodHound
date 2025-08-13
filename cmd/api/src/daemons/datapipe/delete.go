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
	slog.Info("DeleteCollectedGraphData",
		slog.Bool("delete all data", deleteRequest.DeleteAllGraph),
		slog.Bool("delete sourceless data", deleteRequest.DeleteSourcelessGraph),
		slog.String("delete source kinds", strings.Join(deleteRequest.DeleteSourceKinds, ",")))

	operation := ops.StartNewOperation[graph.ID](ops.OperationContext{
		Parent:     ctx,
		DB:         graphDB,
		NumReaders: 1,
		NumWriters: 1,
	})

	deleteSourceKinds := make(graph.Kinds, len(deleteRequest.DeleteSourceKinds))
	for i, s := range deleteRequest.DeleteSourceKinds {
		deleteSourceKinds[i] = graph.StringKind(s)
	}

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- graph.ID) error {
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
