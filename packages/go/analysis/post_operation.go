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
	"time"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/trace"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
)

func NewPropertiesWithLastSeen() *graph.Properties {
	newProperties := graph.NewProperties()
	newProperties.Set(common.LastSeen.String(), time.Now().UTC())

	return newProperties
}

type StatTrackedOperation[T any] struct {
	Stats     post.AtomicPostProcessingStats
	Operation *ops.Operation[T]
}

func NewPostRelationshipOperation(ctx context.Context, db graph.Database, operationName string) StatTrackedOperation[post.EnsureRelationshipJob] {
	operation := StatTrackedOperation[post.EnsureRelationshipJob]{}
	operation.NewOperation(ctx, db)
	operation.Operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan post.EnsureRelationshipJob) error {
		defer trace.Function(ctx, "PostRelationshipOperation")()

		relProp := NewPropertiesWithLastSeen()

		for nextJob := range inC {
			relProps := relProp

			if len(nextJob.RelProperties) > 0 {
				relProps = relProp.Clone()

				for key, val := range nextJob.RelProperties {
					relProps.Set(key, val)
				}
			}

			if err := batch.CreateRelationshipByIDs(nextJob.FromID, nextJob.ToID, nextJob.Kind, relProps); err != nil {
				return err
			}

			operation.Stats.AddRelationshipsCreated(nextJob.Kind, 1)
		}

		return nil
	})

	return operation
}

func (s *StatTrackedOperation[T]) NewOperation(ctx context.Context, db graph.Database) {
	s.Stats = post.NewAtomicPostProcessingStats()
	s.Operation = ops.StartNewOperation[T](ops.OperationContext{
		Parent:     ctx,
		DB:         db,
		NumReaders: MaximumDatabaseParallelWorkers,
		NumWriters: 1,
	})
}

func (s *StatTrackedOperation[T]) Done() error {
	return s.Operation.Done()
}
