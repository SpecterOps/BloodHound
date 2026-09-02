// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package post

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

func MigrationForDCAPostProcessedEdges(ctx context.Context, db graph.Database, migratedRelationships graph.Kinds) error {
	for _, kind := range migratedRelationships {
		var relationshipIDs []graph.ID

		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			fetchedRelationshipIDs, err := ops.FetchRelationshipIDs(tx.Relationships().Filterf(func() graph.Criteria {
				// Only remove existing post-processed edges if they contain a `lastseen` property and don't involve meta nodes
				return query.And(
					query.Not(query.Kind(query.Start(), graphschema.Meta, graphschema.MetaDetail)),
					query.Kind(query.Relationship(), kind),
					query.Exists(query.RelationshipProperty(common.LastSeen.String())),
					query.Not(query.Kind(query.End(), graphschema.Meta, graphschema.MetaDetail)),
				)
			}))

			relationshipIDs = fetchedRelationshipIDs
			return err
		}); err != nil {
			return err
		}

		// Only run deletion which outputs measurements to the log if there are relationships to delete
		if len(relationshipIDs) > 0 {
			measuref := measure.ContextMeasure(
				ctx,
				slog.LevelInfo,
				fmt.Sprintf("Deleted %d %s relationships for DCA Post-Processing migration", len(relationshipIDs), kind.String()),
				attr.Namespace("analysis"),
				attr.Function("MigrationForDCAPostProcessedEdges"),
				attr.Scope("process"),
			)

			if err := db.BatchOperation(ctx, func(batch graph.Batch) error {
				for _, relationshipID := range relationshipIDs {
					if err := batch.DeleteRelationship(relationshipID); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return err
			}

			// Output the measurement if there is no deletion error
			measuref()
		}
	}

	return nil
}
