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
	"sync"
	"testing"

	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

var (
	kindA = graph.StringKind("A")
	kindB = graph.StringKind("B")
	kindC = graph.StringKind("C")
)

func TestKeyEncoder_Key_Deterministic(t *testing.T) {
	enc := NewKeyEncoder()

	start, end := uint64(12345), uint64(67890)

	key1 := enc.EdgeKey(start, end, kindA)
	key2 := enc.EdgeKey(start, end, kindA)

	if key1 != key2 {
		t.Fatalf("KeyEncoder.Key should be deterministic, got %d and %d", key1, key2)
	}
}

func TestKeyEncoder_Key_VaryingInputs(t *testing.T) {
	enc := NewKeyEncoder()
	start, end := uint64(1), uint64(2)

	keys := make(map[uint64]struct{})
	keys[enc.EdgeKey(start, end, kindA)] = struct{}{}
	keys[enc.EdgeKey(start+10, end, kindA)] = struct{}{}
	keys[enc.EdgeKey(start, end, kindB)] = struct{}{}
	keys[enc.EdgeKey(end, start, kindA)] = struct{}{}

	if len(keys) != 4 {
		t.Fatalf("Expected 4 distinct keys for different inputs, got %d {%v+}", len(keys), keys)
	}

	_, exists := keys[enc.EdgeKey(start, end, kindA)]
	require.True(t, exists)

	_, exists = keys[enc.EdgeKey(start+10, end, kindA)]
	require.True(t, exists)

	_, exists = keys[enc.EdgeKey(start, end, kindB)]
	require.True(t, exists)

	_, exists = keys[enc.EdgeKey(end, start, kindA)]
	require.True(t, exists)
}

// Helper to compute a key without using the internal encoder (for expected ordering checks)
func computeKey(start, end uint64, k graph.Kind) uint64 {
	enc := NewKeyEncoder()
	return enc.EdgeKey(start, end, k)
}

func TestTrackerBuilder_Build_SortsKeysAndIDs(t *testing.T) {
	builder := NewTrackerBuilder()

	// Insert edges in unsorted order.
	builder.TrackEdge(100, 5, 10, kindA) // edgeID 100
	builder.TrackEdge(101, 2, 8, kindB)  // edgeID 101
	builder.TrackEdge(102, 7, 3, kindC)  // edgeID 102

	sub := builder.Build()

	// Verify that edgeKeys are sorted ascending.
	for i := 1; i < len(sub.entities); i++ {
		if sub.entities[i-1].Key > sub.entities[i].Key {
			t.Fatalf("edgeKeys not sorted: %v", sub.entities)
		}
	}

	// Verify that edgeIDs have been reordered to match the sorted keys.
	// Compute the keys manually to know the expected order.
	keys := []uint64{
		computeKey(5, 10, kindA), // edge 100
		computeKey(2, 8, kindB),  // edge 101
		computeKey(7, 3, kindC),  // edge 102
	}

	// Sort the keys to get the expected order.
	sortedIdx := make([]int, len(keys))
	for i := range sortedIdx {
		sortedIdx[i] = i
	}

	// Simple bubble sort on indices based on keys (just for test readability)
	for i := range len(keys) {
		for j := i + 1; j < len(keys); j++ {
			if keys[sortedIdx[i]] > keys[sortedIdx[j]] {
				sortedIdx[i], sortedIdx[j] = sortedIdx[j], sortedIdx[i]
			}
		}
	}
	expectedOrder := []uint64{
		[]uint64{100, 101, 102}[sortedIdx[0]],
		[]uint64{100, 101, 102}[sortedIdx[1]],
		[]uint64{100, 101, 102}[sortedIdx[2]],
	}

	if len(sub.entities) != len(expectedOrder) {
		t.Fatalf("unexpected number of edges")
	}

	for i, edge := range sub.entities {
		if edge.ID != expectedOrder[i] {
			t.Fatalf("edge ID at index %d expected %d, got %d", i, expectedOrder[i], edge.ID)
		}
	}
}

func TestTracker_HasEdge_And_DeletedEdges(t *testing.T) {
	builder := NewTrackerBuilder()

	// Edge set we will actually add to the subgraph.
	builder.TrackEdge(10, 1, 2, kindA) // edgeID 10
	builder.TrackEdge(20, 3, 4, kindB) // edgeID 20
	builder.TrackEdge(30, 5, 6, kindC) // edgeID 30

	sub := builder.Build()

	// Query a subset of edges – deliberately omit the second one.
	if !sub.HasEdge(1, 2, kindA) {
		t.Fatalf("expected edge (1,2,kindA) to be present")
	}
	if sub.HasEdge(3, 5, kindB) {
		t.Fatalf("expected edge (3,5,kindB) to be *absent* from HasEdge calls")
	}
	if !sub.HasEdge(5, 6, kindC) {
		t.Fatalf("expected edge (5,6,kindC) to be present")
	}

	// DeletedEdges should now contain only the ID of the edge we never queried.
	deleted := sub.Deleted()
	if len(deleted) != 1 {
		t.Fatalf("expected exactly one deleted edge, got %d", len(deleted))
	}
	if deleted[0] != 20 {
		t.Fatalf("expected deleted edge ID to be 20, got %d", deleted[0])
	}
}

func TestTracker_HasEdge_DuplicateCalls(t *testing.T) {
	builder := NewTrackerBuilder()
	builder.TrackEdge(55, 11, 22, kindA)
	sub := builder.Build()

	// Call HasEdge many times – it should stay true and not affect DeletedEdges.
	for i := range 5 {
		if !sub.HasEdge(11, 22, kindA) {
			t.Fatalf("edge should be found on iteration %d", i)
		}
	}

	if got := sub.Deleted(); len(got) != 0 {
		t.Fatalf("expected no deleted edges after repeated HasEdge calls, got %v", got)
	}
}

func TestTracker_ConcurrentHasEdge(t *testing.T) {
	builder := NewTrackerBuilder()
	// Build a larger set of edges (10 edges)
	for i := range 10 {
		builder.TrackEdge(uint64(100+i), uint64(i), uint64(i+100), kindA)
	}
	sub := builder.Build()

	var wg sync.WaitGroup
	query := func(startIdx, endIdx int) {
		defer wg.Done()
		for i := startIdx; i < endIdx; i++ {
			if ok := sub.HasEdge(uint64(i), uint64(i+100), kindA); !ok {
				t.Errorf("expected edge %d to exist", i)
			}
		}
	}

	wg.Add(2)
	go query(0, 5)  // first half
	go query(5, 10) // second half
	wg.Wait()

	if del := sub.Deleted(); len(del) != 0 {
		t.Fatalf("expected no deleted edges after concurrent queries, got %v", del)
	}
}
