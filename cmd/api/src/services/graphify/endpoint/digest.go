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
package endpoint

import (
	"encoding/json"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/specterops/bloodhound/packages/go/ein"
)

// CacheEntryDigester generates deterministic hash keys for cache entries based on
// endpoint matching. It manages xxhash.Digest instances using a sync.Pool
// to reduce allocation overhead during high-frequency, concurrent hashing operations.
type CacheEntryDigester struct {
	digesters *sync.Pool
}

// NewCacheEntryDigester creates and returns a new initialized CacheEntryDigester instance.
func NewCacheEntryDigester() *CacheEntryDigester {
	return &CacheEntryDigester{
		digesters: &sync.Pool{
			New: func() any {
				return xxhash.New()
			},
		},
	}
}

// get retrieves an xxhash.Digest from the pool or creates a new one if none are available.
func (s *CacheEntryDigester) get() *xxhash.Digest {
	rawEntry := s.digesters.Get()

	if encoder, typeOK := rawEntry.(*xxhash.Digest); typeOK {
		return encoder
	}

	return xxhash.New()
}

// put returns an xxhash.Digest to the pool for reuse.
func (s *CacheEntryDigester) put(ref *xxhash.Digest) {
	s.digesters.Put(ref)
}

// EncodeAnyValue marshals an arbitrary Go value into a JSON byte slice.
// This is used to ensure consistent string representation of values during hashing.
func EncodeAnyValue(value any) ([]byte, error) {
	return json.Marshal(value)
}

// DigestEndpoint computes a unique 64-bit hash key for a given IngestibleEndpoint.
// The resulting hash is deterministic based on the endpoint's MatchBy strategy:
//   - MatchByProperty: Includes Kind and sorted Matchers (Key + JSON-encoded Value).
//   - MatchByName: Uses "name" prefix and the endpoint Value.
//   - Default (MatchByObjectId): Uses "objectid" prefix and the endpoint Value.
//
// It returns the computed uint64 hash and any error encountered during encoding.
func (s *CacheEntryDigester) DigestEndpoint(endpoint ein.IngestibleEndpoint) (uint64, error) {
	digester := s.get()
	defer s.put(digester)

	digester.Reset()

	switch endpoint.MatchBy {
	case ein.MatchByProperty:
		sortedMatchers := ein.SortedMatchExpressions(endpoint.Matchers)

		if endpoint.Kind != nil {
			digester.WriteString(endpoint.Kind.String())
			digester.WriteString(" ")
		}

		for idx, sortedMatcher := range sortedMatchers {
			if idx > 0 {
				digester.WriteString(" ")
			}

			digester.WriteString(sortedMatcher.Key)
			digester.WriteString(" ")

			if encodedValue, err := EncodeAnyValue(sortedMatcher.Value); err != nil {
				return 0, err
			} else {
				digester.Write(encodedValue)
			}
		}

	case ein.MatchByName:
		digester.WriteString("name")
		digester.WriteString(" ")
		digester.WriteString(endpoint.Value)

	default:
		digester.WriteString("objectid")
		digester.WriteString(" ")
		digester.WriteString(endpoint.Value)
	}

	key := digester.Sum64()
	return key, nil
}
