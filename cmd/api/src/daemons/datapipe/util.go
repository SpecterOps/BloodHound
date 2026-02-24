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

package datapipe

import (
	"context"
	"iter"
	"log/slog"
	"math/rand"
	"time"

	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

func RandomDurationBetween(min, max time.Duration) time.Duration {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	durationRange := max - min
	return min + time.Duration(r.Int63())%durationRange
}

/*
BatchUpdateNodes batch updates nodes by objectid.
Nodes without an objectid are individually updated by ID. As such, when using this function,
most if not all nodes should have an objectid to avoid unnecessary database operations.
Operation is all-or-nothing.
*/
func BatchUpdateNodes(ctx context.Context, graphDB graph.Database, nodes iter.Seq[*graph.Node]) error {
	if err := graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {

		missingObjectIDs := make([]*graph.Node, 0, 10)

		// Batch update the nodes that have objectIDs
		if err := graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
			for node := range nodes {
				if _, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
					missingObjectIDs = append(missingObjectIDs, node)
					continue
				}

				if err := batch.UpdateNodeBy(graph.NodeUpdate{
					Node:               node,
					IdentityProperties: []string{common.ObjectID.String()},
				}); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			return err
		}

		if len(missingObjectIDs) > 0 {
			slog.Info("Individually updating nodes without objectids", slog.Int("length", len(missingObjectIDs)))
		}

		// Individually updating nodes without objectIDs
		for _, node := range missingObjectIDs {
			if err := tx.UpdateNode(node); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
