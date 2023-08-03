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
	"sync"
	"sync/atomic"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

type StatTrackedOperation[T any] struct {
	Stats     AtomicPostProcessingStats
	Operation *ops.Operation[T]
}

func NewPostRelationshipOperation(ctx context.Context, db graph.Database, operationName string) StatTrackedOperation[CreatePostRelationshipJob] {
	operation := StatTrackedOperation[CreatePostRelationshipJob]{}
	operation.NewOperation(ctx, db)
	operation.Operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan CreatePostRelationshipJob) error {
		defer log.Measure(log.LevelInfo, operationName)()

		var (
			relProp = NewPropertiesWithLastSeen()
		)

		for nextJob := range inC {
			if err := batch.CreateRelationshipByIDs(nextJob.FromID, nextJob.ToID, nextJob.Kind, relProp); err != nil {
				return err
			}

			operation.Stats.AddRelationshipsCreated(nextJob.Kind, 1)
		}

		return nil
	})
	return operation
}

func (s *StatTrackedOperation[T]) NewOperation(ctx context.Context, db graph.Database) {
	s.Stats = NewAtomicPostProcessingStats()
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

type AtomicPostProcessingStats struct {
	RelationshipsCreated map[graph.Kind]*int32
	RelationshipsDeleted map[graph.Kind]*int32
	mutex                *sync.Mutex
}

func NewAtomicPostProcessingStats() AtomicPostProcessingStats {
	return AtomicPostProcessingStats{
		RelationshipsCreated: make(map[graph.Kind]*int32),
		RelationshipsDeleted: make(map[graph.Kind]*int32),
		mutex:                &sync.Mutex{},
	}
}

func (s *AtomicPostProcessingStats) AddRelationshipsCreated(kind graph.Kind, numCreated int32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if val, ok := s.RelationshipsCreated[kind]; !ok {
		s.RelationshipsCreated[kind] = &numCreated
	} else {
		atomic.AddInt32(val, numCreated)
	}
}

func (s *AtomicPostProcessingStats) AddRelationshipsDeleted(kind graph.Kind, numDeleted int32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if val, ok := s.RelationshipsDeleted[kind]; !ok {
		s.RelationshipsDeleted[kind] = &numDeleted
	} else {
		atomic.AddInt32(val, numDeleted)
	}
}

func (s *AtomicPostProcessingStats) Merge(other *AtomicPostProcessingStats) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for key, value := range other.RelationshipsCreated {
		if val, ok := s.RelationshipsCreated[key]; !ok {
			s.RelationshipsCreated[key] = value
		} else {
			atomic.AddInt32(val, *value)
		}
	}

	for key, value := range other.RelationshipsDeleted {
		if val, ok := s.RelationshipsDeleted[key]; !ok {
			s.RelationshipsDeleted[key] = value
		} else {
			atomic.AddInt32(val, *value)
		}
	}
}

func (s *AtomicPostProcessingStats) LogStats() {
	// Only output stats during debug runs
	if log.GlobalLevel() > log.LevelDebug {
		return
	}

	log.Debugf("Relationships deleted before post-processing:")

	for _, relationship := range atomicStatsSortedKeys(s.RelationshipsDeleted) {
		if numDeleted := int(*s.RelationshipsDeleted[relationship]); numDeleted > 0 {
			log.Debugf("    %s %d", relationship.String(), numDeleted)
		}
	}

	log.Debugf("Relationships created after post-processing:")

	for _, relationship := range atomicStatsSortedKeys(s.RelationshipsCreated) {
		if numCreated := int(*s.RelationshipsCreated[relationship]); numCreated > 0 {
			log.Debugf("    %s %d", relationship.String(), numCreated)
		}
	}
}

func NewPropertiesWithLastSeen() *graph.Properties {
	newProperties := graph.NewProperties()
	newProperties.Set(common.LastSeen.String(), time.Now().UTC())

	return newProperties
}
