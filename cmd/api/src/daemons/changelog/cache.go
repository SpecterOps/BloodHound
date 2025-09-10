// Copyright 2025 Specter Ops, Inc.
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
package changelog

import (
	"fmt"
	"sync"
)

// cache is an in-memory deduplication layer for changelog entries.
//
// It maps a change's identity hash to its most recent content hash,
// allowing the changelog to decide whether an incoming change
// is new/modified (submit) or unchanged (skip).
//
// It also maintains simple hit/miss counters for observability, which can be reset between ingest batches.
type cache struct {
	data  map[uint64]uint64
	mu    sync.Mutex
	stats cacheStats
}

func newCache(size int) *cache {
	if size == 0 {
		size = 1_000_000
	}

	return &cache{
		data:  make(map[uint64]uint64, size),
		mu:    sync.Mutex{},
		stats: cacheStats{},
	}
}

type cacheStats struct {
	Hits   uint64 // unchanged
	Misses uint64 // new or modified
}

// shouldSubmit compares the proposed change against the cached snapshot.
// It returns true if the change is new or modified and should be submitted
// downstream. If the change is identical to the cached version, it returns false.
func (s *cache) shouldSubmit(change Change) (bool, error) {
	idHash := change.IdentityKey()
	contentHash, err := change.Hash()

	if err != nil {
		return false, fmt.Errorf("hash proposed change: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// try to diff against the storedHash snapshot
	if storedHash, ok := s.data[idHash]; ok {
		if storedHash == contentHash {
			s.stats.Hits++
			return false, nil // unchanged
		}
	}

	// new or modified -> update cache to the new snapshot
	s.data[idHash] = contentHash
	s.stats.Misses++
	return true, nil
}

func (s *cache) resetStats() cacheStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.stats
	s.stats = cacheStats{} // zero it out
	return old
}
