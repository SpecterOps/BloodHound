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
package delta

import (
	"context"
	"encoding/binary"
	"log/slog"
	"slices"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/specterops/bloodhound/packages/go/trace"
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

// EdgeKey retrieves or allocates an encoder from the pool and computes the edge key.
func (s *KeyEncoderPool) EdgeKey(start, end uint64, kind graph.Kind) uint64 {
	var (
		raw             = s.encoders.Get()
		encoder, typeOK = raw.(*KeyEncoder)
	)

	if !typeOK {
		encoder = NewKeyEncoder()
	}

	key := encoder.EdgeKey(start, end, kind)

	if typeOK {
		s.encoders.Put(raw)
	}

	return key
}

// trackedEntity represents a single entity being tracked within a Tracker.
type trackedEntity struct {
	ID  uint64
	Key uint64
}

// Tracker tracks edges and nodes as hashed keys that can be looked and checked. The Tracker also
// maintains a list of seen entities and provides methods to detect deletions.
type Tracker struct {
	entities     []trackedEntity
	seenKeys     map[uint64]struct{}
	seenKeysLock sync.RWMutex
	encoderPool  *KeyEncoderPool
}

// Seen returns the number of unique keys currently tracked and seen by either HasNode or HasEdge.
func (s *Tracker) Seen() int {
	s.seenKeysLock.RLock()
	defer s.seenKeysLock.RUnlock()

	return len(s.seenKeys)
}

// Deleted returns a slice of IDs for edges that were not seen during the operation.
func (s *Tracker) Deleted() []uint64 {
	s.seenKeysLock.RLock()
	defer s.seenKeysLock.RUnlock()

	deletedEdges := make([]uint64, 0, len(s.entities)-len(s.seenKeys))

	for _, edge := range s.entities {
		if _, seen := s.seenKeys[edge.Key]; !seen {
			deletedEdges = append(deletedEdges, edge.ID)
		}
	}

	return deletedEdges
}

// HasEdge checks whether a specific edge exists in the tracker. If found, it marks the key as seen.
func (s *Tracker) HasEdge(start, end uint64, edgeKind graph.Kind) bool {
	var (
		edgeKey  = s.encoderPool.EdgeKey(start, end, edgeKind)
		_, found = slices.BinarySearchFunc(s.entities, edgeKey, func(e trackedEntity, t uint64) int {
			if e.Key < t {
				return -1
			}

			if e.Key > t {
				return 1
			}

			return 0
		})
	)

	if found {
		s.seenKeysLock.Lock()
		s.seenKeys[edgeKey] = struct{}{}
		s.seenKeysLock.Unlock()
	}

	return found
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

// TrackEdge adds an edge to the builder for later tracking.
func (s *EdgeTrackerBuilder) TrackEdge(edge, start, end uint64, kind graph.Kind) {
	s.edges = append(s.edges, trackedEntity{
		ID:  edge,
		Key: s.encoderPool.EdgeKey(start, end, kind),
	})
}

// Build constructs a Tracker from the accumulated edges.
func (s *EdgeTrackerBuilder) Build() *Tracker {
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

	return &Tracker{
		entities:    s.edges,
		encoderPool: s.encoderPool,

		// Assume that the tracker will have a high hit ratio. This may be better exposed as a function parameter
		// but for now this seems like a safe bet.
		seenKeys: make(map[uint64]struct{}, len(s.edges)),
	}
}

// FetchTracker retrieves all relevant edges that match one of the edge kinds given from the database. It uses
// this data to then build a Tracker.
func FetchTracker(ctx context.Context, db graph.Database, edgeKinds graph.Kinds) (*Tracker, error) {
	var (
		tracef     = trace.Function(ctx, "FetchTracker")
		builder    = NewTrackerBuilder()
		numResults = uint64(0)
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Relationships().Filter(query.And(
			query.Not(query.KindIn(query.Start(), graph.StringKind("Meta"), graph.StringKind("MetaDetail"))),
			query.KindIn(query.Relationship(), edgeKinds...),
			query.Not(query.KindIn(query.End(), graph.StringKind("Meta"), graph.StringKind("MetaDetail"))),
		)).Query(
			func(results graph.Result) error {
				var (
					edgeID   graph.ID
					startID  graph.ID
					edgeKind graph.Kind
					endID    graph.ID
				)

				for results.Next() {
					if err := results.Scan(&edgeID, &startID, &edgeKind, &endID); err != nil {
						return err
					}

					builder.TrackEdge(edgeID.Uint64(), startID.Uint64(), endID.Uint64(), edgeKind)
					numResults += 1
				}

				results.Close()
				return results.Error()
			},
			query.Returning(
				query.RelationshipID(),
				query.StartID(),
				query.KindsOf(query.Relationship()),
				query.EndID(),
			),
		)
	}); err != nil {
		return nil, err
	}

	tracef(slog.Uint64("num_edges_tracked", numResults))
	return builder.Build(), nil
}
