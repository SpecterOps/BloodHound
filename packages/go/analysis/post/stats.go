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
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/specterops/bloodhound/packages/go/bhlog/level"
	"github.com/specterops/dawgs/graph"
)

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
	if level.GlobalAccepts(slog.LevelDebug) {
		return
	}

	slog.Debug("Relationships deleted before post-processing:")

	for _, relationship := range atomicStatsSortedKeys(s.RelationshipsDeleted) {
		if numDeleted := int(*s.RelationshipsDeleted[relationship]); numDeleted > 0 {
			slog.Debug(fmt.Sprintf("    %s %d", relationship.String(), numDeleted))
		}
	}

	slog.Debug("Relationships created after post-processing:")

	for _, relationship := range atomicStatsSortedKeys(s.RelationshipsCreated) {
		if numCreated := int(*s.RelationshipsCreated[relationship]); numCreated > 0 {
			slog.Debug(fmt.Sprintf("    %s %d", relationship.String(), numCreated))
		}
	}
}
