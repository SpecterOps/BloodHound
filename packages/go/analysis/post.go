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

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/level"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
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
	if !level.GlobalAccepts(slog.LevelDebug) {
		return
	}

	slog.Debug("Relationships deleted before post-processing:")

	for _, relationship := range statsSortedKeys(s.RelationshipsDeleted) {
		if numDeleted := s.RelationshipsDeleted[relationship]; numDeleted > 0 {
			slog.Debug(
				"Deleted relationship",
				slog.String("relationship", relationship.String()),
				slog.Int("num_deleted", numDeleted),
			)
		}
	}

	slog.Debug("Relationships created after post-processing:")

	for _, relationship := range statsSortedKeys(s.RelationshipsCreated) {
		if numCreated := s.RelationshipsCreated[relationship]; numCreated > 0 {
			slog.Debug(
				"Created relationship",
				slog.String("relationship", relationship.String()),
				slog.Int("num_created", numCreated),
			)
		}
	}
}

type DeleteRelationshipJob struct {
	Kind graph.Kind
	ID   graph.ID
}

func DeleteTransitEdges(ctx context.Context, db graph.Database, baseKinds graph.Kinds, targetRelationships graph.Kinds) (*post.AtomicPostProcessingStats, error) {
	var (
		relationshipIDs []graph.ID
		stats           = post.NewAtomicPostProcessingStats()
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
