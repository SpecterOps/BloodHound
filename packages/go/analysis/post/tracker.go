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
	"encoding/binary"
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

// KeyEncoder encodes node and edge identifiers into hash keys using xxhash. It is used
// to generate unique keys for tracking entities in a Tracker.
type KeyEncoder struct {
	digester *xxhash.Digest
	buffer   [8]byte
}

// NewKeyEncoder creates a new key encoder instance with default settings.
func NewKeyEncoder() *KeyEncoder {
	return &KeyEncoder{
		digester: xxhash.New(),
	}
}

// NodeKey computes the hash key for a given node ID and list of kinds.
func (s *KeyEncoder) NodeKey(node uint64, kinds graph.Kinds) uint64 {
	s.digester.Reset()

	// Node identifier and sorted kinds make up a node key
	binary.LittleEndian.PutUint64(s.buffer[:], node)
	s.digester.Write(s.buffer[:])

	kindStrs := kinds.Strings()
	slices.Sort(kindStrs)

	for _, kindStr := range kindStrs {
		s.digester.WriteString(kindStr)
	}

	// Sum the digest
	return s.digester.Sum64()
}

// EdgeKey computes the hash key for an edge defined by start/end IDs and kind.
func (s *KeyEncoder) EdgeKey(start, end uint64, kind graph.Kind) uint64 {
	s.digester.Reset()

	// Start and end identifiers and the edge's kind make up an edge key
	binary.LittleEndian.PutUint64(s.buffer[:], start)
	s.digester.Write(s.buffer[:])

	binary.LittleEndian.PutUint64(s.buffer[:], end)
	s.digester.Write(s.buffer[:])

	// Edge type
	s.digester.WriteString(kind.String())

	// Sum the digest
	return s.digester.Sum64()
}

// ignoredFingerprintKeys contains property keys that should be excluded from fingerprint
// comparisons as they would result in unnecessary edge churn. While `lastseen` should never be
// present on these edges, `firstseen` won't be known by the calling function.
var ignoredFingerprintKeys = map[string]struct{}{
	"firstseen": {},
	"lastseen":  {},
}

// PropertiesFingerprint computes a deterministic hash of a property map. Keys are sorted
// lexicographically before hashing so that insertion order does not affect the result.
// Properties listed in ignoredFingerprintKeys are excluded from the hash.
func (s *KeyEncoder) PropertiesFingerprint(properties map[string]any) uint64 {
	if len(properties) == 0 {
		return 0
	}

	s.digester.Reset()

	sortedKeys := make([]string, 0, len(properties))
	for key := range properties {
		if _, ignored := ignoredFingerprintKeys[key]; !ignored {
			sortedKeys = append(sortedKeys, key)
		}
	}

	if len(sortedKeys) == 0 {
		return 0
	}

	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		s.digester.WriteString(key)
		fmt.Fprint(s.digester, properties[key])
	}

	return s.digester.Sum64()
}

// KeyEncoderPool manages a pool of KeyEncoder instances for efficient reuse.
type KeyEncoderPool struct {
	encoders *sync.Pool
}

// NewEdgeEncoderPool creates a new pool for edge key encoding.
func NewEdgeEncoderPool() *KeyEncoderPool {
	return &KeyEncoderPool{
		encoders: &sync.Pool{
			New: func() any {
				return NewKeyEncoder()
			},
		},
	}
}

// getEncoder retrieves a KeyEncoder from the pool, allocating a new one if the pool returns
// an unexpected type.
func (s *KeyEncoderPool) getEncoder() (*KeyEncoder, bool) {
	var (
		raw             = s.encoders.Get()
		encoder, typeOK = raw.(*KeyEncoder)
	)

	if !typeOK {
		encoder = NewKeyEncoder()
	}

	return encoder, typeOK
}

// putEncoder returns a KeyEncoder to the pool if it was successfully retrieved from the pool.
func (s *KeyEncoderPool) putEncoder(encoder *KeyEncoder, fromPool bool) {
	if fromPool {
		s.encoders.Put(encoder)
	}
}

// EdgeKey retrieves or allocates an encoder from the pool and computes the edge key.
func (s *KeyEncoderPool) EdgeKey(start, end uint64, kind graph.Kind) uint64 {
	encoder, fromPool := s.getEncoder()
	defer s.putEncoder(encoder, fromPool)

	return encoder.EdgeKey(start, end, kind)
}

// PropertiesFingerprint retrieves or allocates an encoder from the pool and computes the
// deterministic hash of the given property map.
func (s *KeyEncoderPool) PropertiesFingerprint(properties map[string]any) uint64 {
	encoder, fromPool := s.getEncoder()
	defer s.putEncoder(encoder, fromPool)

	return encoder.PropertiesFingerprint(properties)
}

// trackedEntity represents a single entity being tracked within a Tracker.
type trackedEntity struct {
	ID        uint64
	Key       uint64
	PropsHash uint64
}

// Tracker tracks edges and nodes as hashed keys that can be looked and checked. The Tracker also
// maintains a list of seen entities and provides methods to detect deletions.
type Tracker struct {
	entities      []trackedEntity
	entitiesByKey map[uint64]trackedEntity
	uniqueKeys    int
	seenKeys      map[uint64]struct{}
	seenKeysLock  sync.RWMutex
	encoderPool   *KeyEncoderPool
}

// Seen returns the number of unique keys currently tracked and seen by either HasNode or HasEdge.
func (s *Tracker) Seen() int {
	s.seenKeysLock.RLock()
	defer s.seenKeysLock.RUnlock()

	return len(s.seenKeys)
}

// AllSeen returns true when every unique tracked key was visited during reconciliation.
func (s *Tracker) AllSeen() bool {
	s.seenKeysLock.RLock()
	defer s.seenKeysLock.RUnlock()

	return len(s.seenKeys) == s.uniqueKeys
}

// Deleted returns a slice of IDs for edges that were not seen during the operation.
func (s *Tracker) Deleted() []uint64 {
	s.seenKeysLock.RLock()
	defer s.seenKeysLock.RUnlock()

	if len(s.seenKeys) == s.uniqueKeys {
		return nil
	}

	deletedEdges := make([]uint64, 0, len(s.entities)-len(s.seenKeys))

	for _, edge := range s.entities {
		if _, seen := s.seenKeys[edge.Key]; !seen {
			deletedEdges = append(deletedEdges, edge.ID)
		}
	}

	return deletedEdges
}

// HasEdge checks whether a specific edge with matching properties exists in the tracker. An edge
// is considered a match only if the structural key (start, end, kind) matches and the property
// fingerprint is identical. If found, it marks the key as seen.
func (s *Tracker) HasEdge(start, end uint64, edgeKind graph.Kind, properties map[string]any) bool {
	var (
		edgeKey             = s.encoderPool.EdgeKey(start, end, edgeKind)
		propsHash           = s.encoderPool.PropertiesFingerprint(properties)
		edge, keyWasTracked = s.entitiesByKey[edgeKey]
	)

	if keyWasTracked && edge.PropsHash == propsHash {
		s.seenKeysLock.Lock()
		s.seenKeys[edgeKey] = struct{}{}
		s.seenKeysLock.Unlock()

		return true
	}

	return false
}

// EdgeTrackerBuilder builds a Tracker from a sequence of tracked edges.
type EdgeTrackerBuilder struct {
	edges       []trackedEntity
	encoderPool *KeyEncoderPool
}

// NewTrackerBuilder creates a new builder for constructing a Tracker.
func NewTrackerBuilder() *EdgeTrackerBuilder {
	return &EdgeTrackerBuilder{
		encoderPool: NewEdgeEncoderPool(),
	}
}

// TrackEdge adds an edge with its properties to the builder for later tracking.
func (s *EdgeTrackerBuilder) TrackEdge(edge, start, end uint64, kind graph.Kind, properties map[string]any) {
	s.edges = append(s.edges, trackedEntity{
		ID:        edge,
		Key:       s.encoderPool.EdgeKey(start, end, kind),
		PropsHash: s.encoderPool.PropertiesFingerprint(properties),
	})
}

// Build constructs a Tracker from the accumulated edges.
func (s *EdgeTrackerBuilder) Build() *Tracker {
	var (
		entitiesByKey = make(map[uint64]trackedEntity, len(s.edges))
		uniqueKeys    int
	)

	// Sort the edges before building the tracker
	slices.SortStableFunc(s.edges, func(a, b trackedEntity) int {
		if a.Key < b.Key {
			return -1
		}

		if a.Key > b.Key {
			return 1
		}

		return 0
	})

	for _, edge := range s.edges {
		if _, keyAlreadyTracked := entitiesByKey[edge.Key]; !keyAlreadyTracked {
			entitiesByKey[edge.Key] = edge
			uniqueKeys += 1
		}
	}

	return &Tracker{
		entities:      s.edges,
		entitiesByKey: entitiesByKey,
		uniqueKeys:    uniqueKeys,
		encoderPool:   s.encoderPool,

		// Assume that the tracker will have a high hit ratio. This may be better exposed as a function parameter
		// but for now this seems like a safe bet.
		seenKeys: make(map[uint64]struct{}, uniqueKeys),
	}
}

// FetchTracker retrieves all relevant edges that match one of the edge kinds given from the database. It uses
// this data to then build a Tracker. Edge properties are included in the tracker so that property changes
// are detected during delta filtering.
func FetchTracker(ctx context.Context, db graph.Database, edgeKinds graph.Kinds) (*Tracker, error) {
	builder := NewTrackerBuilder()

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Relationships().Filter(query.And(
			query.Not(query.KindIn(query.Start(), graphschema.Meta, graphschema.MetaDetail)),
			query.KindIn(query.Relationship(), edgeKinds...),
			query.Not(query.KindIn(query.End(), graphschema.Meta, graphschema.MetaDetail)),
		)).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for relationship := range cursor.Chan() {
				builder.TrackEdge(relationship.ID.Uint64(), relationship.StartID.Uint64(), relationship.EndID.Uint64(), relationship.Kind, relationship.Properties.Map)
			}

			return cursor.Error()
		})
	}); err != nil {
		return nil, err
	}

	return builder.Build(), nil
}
