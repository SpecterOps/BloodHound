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

package analysis

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/level"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func statsSortedKeys(value map[graph.Kind]int) []graph.Kind {
	kinds := make([]graph.Kind, 0, len(value))

	for key := range value {
		kinds = append(kinds, key)
	}

	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].String() > kinds[j].String()
	})

	return kinds
}

func atomicStatsSortedKeys(value map[graph.Kind]*int32) []graph.Kind {
	kinds := make([]graph.Kind, 0, len(value))

	for key := range value {
		kinds = append(kinds, key)
	}

	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].String() > kinds[j].String()
	})

	return kinds
}

type PostProcessingStats struct {
	RelationshipsCreated map[graph.Kind]int
	RelationshipsDeleted map[graph.Kind]int
}

func NewPostProcessingStats() PostProcessingStats {
	return PostProcessingStats{
		RelationshipsCreated: make(map[graph.Kind]int),
		RelationshipsDeleted: make(map[graph.Kind]int),
	}
}

func (s PostProcessingStats) AddRelationshipsCreated(kind graph.Kind, numCreated int) {
	s.RelationshipsCreated[kind] += numCreated
}

func (s PostProcessingStats) AddRelationshipsDeleted(kind graph.Kind, numCreated int) {
	s.RelationshipsDeleted[kind] += numCreated
}

func (s PostProcessingStats) Merge(other PostProcessingStats) {
	for key, value := range other.RelationshipsCreated {
		s.RelationshipsCreated[key] += value
	}

	for key, value := range other.RelationshipsDeleted {
		s.RelationshipsDeleted[key] += value
	}
}

func (s PostProcessingStats) LogStats() {
	// Only output stats during debug runs
	if level.GlobalAccepts(slog.LevelDebug) {
		return
	}

	slog.Debug("Relationships deleted before post-processing:")

	for _, relationship := range statsSortedKeys(s.RelationshipsDeleted) {
		if numDeleted := s.RelationshipsDeleted[relationship]; numDeleted > 0 {
			slog.Debug(fmt.Sprintf("    %s %d", relationship.String(), numDeleted))
		}
	}

	slog.Debug("Relationships created after post-processing:")

	for _, relationship := range statsSortedKeys(s.RelationshipsCreated) {
		if numDeleted := s.RelationshipsCreated[relationship]; numDeleted > 0 {
			slog.Debug(fmt.Sprintf("    %s %d", relationship.String(), s.RelationshipsCreated[relationship]))
		}
	}
}

//These were created for the new composition method. It was scrapped for the current initiative, but will be useful later
//type CompositionInfo struct {
//	CompositionID int64
//	EdgeIDs       []graph.ID
//	NodeIDs       []graph.ID
//}
//
//func (s CompositionInfo) HasComposition() bool {
//	return len(s.EdgeIDs) > 0 || len(s.NodeIDs) > 0
//}

//
//func (s CompositionInfo) GetCompositionEdges() model.EdgeCompositionEdges {
//	edges := make(model.EdgeCompositionEdges, len(s.EdgeIDs))
//	for i, edgeID := range s.EdgeIDs {
//		edges[i] = model.EdgeCompositionEdge{
//			PostProcessedEdgeID: s.CompositionID,
//			CompositionEdgeID:   edgeID.Int64(),
//		}
//	}
//
//	return edges
//}

//func (s CompositionInfo) GetCompositionNodes() model.EdgeCompositionNodes {
//	edges := make(model.EdgeCompositionNodes, len(s.EdgeIDs))
//	for i, nodeID := range s.NodeIDs {
//		edges[i] = model.EdgeCompositionNode{
//			PostProcessedEdgeID: s.CompositionID,
//			CompositionNodeID:   nodeID.Int64(),
//		}
//	}
//
//	return edges
//}

type CreatePostRelationshipJob struct {
	FromID        graph.ID
	ToID          graph.ID
	Kind          graph.Kind
	RelProperties map[string]any
	//CompositionInfo CompositionInfo
}

type DeleteRelationshipJob struct {
	Kind graph.Kind
	ID   graph.ID
}

func DeleteTransitEdges(ctx context.Context, db graph.Database, baseKinds graph.Kinds, targetRelationships graph.Kinds) (*AtomicPostProcessingStats, error) {
	var (
		relationshipIDs []graph.ID
		stats           = NewAtomicPostProcessingStats()
		operationName   = fmt.Sprintf("Delete %v post-processed relationships", strings.Join(targetRelationships.Strings(), ", "))
	)

	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		operationName,
		attr.Namespace("analysis"),
		attr.Function("DeleteTransitEdges"),
		attr.Scope("process"),
	)()

	for _, kind := range targetRelationships {
		closureKindCopy := kind

		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			fetchedRelationshipIDs, err := ops.FetchRelationshipIDs(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.KindIn(query.Start(), baseKinds...),
					query.Kind(query.Relationship(), closureKindCopy),
					query.KindIn(query.End(), baseKinds...),
				)
			}))

			stats.AddRelationshipsDeleted(closureKindCopy, int32(len(fetchedRelationshipIDs)))
			relationshipIDs = append(relationshipIDs, fetchedRelationshipIDs...)

			return err
		}); err != nil {
			return nil, err
		}
	}

	return &stats, db.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, relationshipID := range relationshipIDs {
			if err := batch.DeleteRelationship(relationshipID); err != nil {
				return err
			}
		}

		return nil
	})
}

func NodesWithoutRelationshipsFilter() graph.Criteria {
	return query.And(
		// Nodes without relationships
		query.Not(query.HasRelationships(query.Node())),

		// And that are not migration related
		query.Not(query.Kind(query.Node(), common.MigrationData)),
	)
}

func ClearOrphanedNodes(ctx context.Context, db graph.Database) error {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Finished deleting orphaned nodes",
		attr.Namespace("analysis"),
		attr.Function("ClearOrphanedNodes"),
		attr.Scope("process"),
	)()

	var operation = ops.StartNewOperation[graph.ID](ops.OperationContext{
		Parent:     ctx,
		DB:         db,
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
		return tx.Nodes().Filterf(NodesWithoutRelationshipsFilter).
			FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
				channels.PipeAll(ctx, cursor.Chan(), outC)
				return cursor.Error()
			})
	})

	return operation.Done()
}
