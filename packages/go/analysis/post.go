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
	"sort"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
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
	if log.GlobalLevel() > log.LevelDebug {
		return
	}

	log.Debugf("Relationships deleted before post-processing:")

	for _, relationship := range statsSortedKeys(s.RelationshipsDeleted) {
		if numDeleted := s.RelationshipsDeleted[relationship]; numDeleted > 0 {
			log.Debugf("    %s %d", relationship.String(), numDeleted)
		}
	}

	log.Debugf("Relationships created after post-processing:")

	for _, relationship := range statsSortedKeys(s.RelationshipsCreated) {
		if numDeleted := s.RelationshipsCreated[relationship]; numDeleted > 0 {
			log.Debugf("    %s %d", relationship.String(), s.RelationshipsCreated[relationship])
		}
	}
}

type CreatePostRelationshipJob struct {
	FromID graph.ID
	ToID   graph.ID
	Kind   graph.Kind
}

type DeleteRelationshipJob struct {
	Kind graph.Kind
	ID   graph.ID
}

func DeleteTransitEdges(ctx context.Context, db graph.Database, fromKind, toKind graph.Kind, targetRelationships ...graph.Kind) (*AtomicPostProcessingStats, error) {
	defer log.Measure(log.LevelInfo, "Finished deleting transit edges")()

	var (
		relationshipIDs []graph.ID
		stats           = NewAtomicPostProcessingStats()
	)

	for _, kind := range targetRelationships {
		closureKindCopy := kind

		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			fetchedRelationshipIDs, err := ops.FetchRelationshipIDs(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Start(), fromKind),
					query.Kind(query.Relationship(), closureKindCopy),
					query.Kind(query.End(), toKind),
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
	defer log.Measure(log.LevelInfo, "Finished deleting orphaned nodes")()

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
