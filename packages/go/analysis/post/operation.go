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

package post

import (
	"context"
	"time"

	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
)

// MaximumDatabaseParallelWorkers is the upper bound on the number of concurrent
// reader goroutines used when constructing a new operation context. A single
// writer is always used to serialise graph mutations.
const (
	MaximumDatabaseParallelWorkers = 6
)

// NewPropertiesWithLastSeen returns a new graph.Properties instance pre-populated
// with the current UTC time set as the lastseen property. This is used as the
// base set of properties applied to every relationship created during post-processing.
func NewPropertiesWithLastSeen() *graph.Properties {
	newProperties := graph.NewProperties()
	newProperties.Set(common.LastSeen.String(), time.Now().UTC())

	return newProperties
}

// StatTrackedOperation couples an ongoing graph operation with an
// AtomicPostProcessingStats collector so that callers can observe how many
// relationships were created or deleted over the lifetime of the operation.
// The type parameter T is the job type consumed by the operation's writer.
type StatTrackedOperation[T any] struct {
	// Stats accumulates relationship create/delete counts in a thread-safe manner.
	Stats AtomicPostProcessingStats
	// Operation is the underlying concurrent graph operation that manages reader
	// and writer goroutines.
	Operation *ops.Operation[T]
}

// NewPostRelationshipOperation creates and starts a StatTrackedOperation that processes
// EnsureRelationshipJob values. A single writer goroutine is registered that batches
// relationship creation against the provided database. For each job:
//   - If the job carries additional RelProperties they are merged (via clone) on top
//     of the shared LastSeen properties so that base properties are not mutated.
//   - The relationship is inserted by node ID using the job's Kind.
//   - The Stats counter for that Kind is incremented atomically.
//
// Callers must submit jobs through the operation's reader pipeline and then call
// Done to flush all pending writes and obtain any batch error.
func NewPostRelationshipOperation(ctx context.Context, db graph.Database, operationName string) StatTrackedOperation[EnsureRelationshipJob] {
	operation := StatTrackedOperation[EnsureRelationshipJob]{}
	operation.NewOperation(ctx, db)
	operation.Operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan EnsureRelationshipJob) error {
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

// NewOperation initialises the Stats collector and starts a new underlying graph
// operation configured with MaximumDatabaseParallelWorkers readers and a single
// writer. It is called automatically by NewPostRelationshipOperation but can also
// be used directly when constructing a StatTrackedOperation for custom job types.
func (s *StatTrackedOperation[T]) NewOperation(ctx context.Context, db graph.Database) {
	s.Stats = NewAtomicPostProcessingStats()
	s.Operation = ops.StartNewOperation[T](ops.OperationContext{
		Parent:     ctx,
		DB:         db,
		NumReaders: MaximumDatabaseParallelWorkers,
		NumWriters: 1,
	})
}

// Done signals that no more jobs will be submitted, waits for all in-flight
// reader and writer goroutines to finish, and returns the first error encountered
// during the operation, if any.
func (s *StatTrackedOperation[T]) Done() error {
	return s.Operation.Done()
}
