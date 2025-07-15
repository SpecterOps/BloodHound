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

package impact_test

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/analysis/impact"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

var (
	aKind       = graph.StringKind("A")
	edgeKind    = graph.StringKind("EDGE")
	nextID      = graph.ID(0)
)

func resetNextID() {
	nextID = 0
}

func getNextID() graph.ID {
	id := nextID
	nextID++

	return id
}

func descend(trunk *graph.PathSegment, nextNode *graph.Node) *graph.PathSegment {
	return trunk.Descend(nextNode, rel(nextNode, trunk.Node))
}

func rel(start, end *graph.Node) *graph.Relationship {
	return graph.NewRelationship(getNextID(), start.ID, end.ID, nil, edgeKind)
}

func node(nodeKinds ...graph.Kind) *graph.Node {
	return graph.NewNode(getNextID(), nil, nodeKinds...)
}

func requireImpact(t *testing.T, agg impact.PathAggregator, nodeID uint64, containedNodes ...uint64) {
	nodeImpact := agg.Cardinality(nodeID).(cardinality.Duplex[uint64])

	if int(nodeImpact.Cardinality()) != len(containedNodes) {
		t.Fatalf("Expected node %d to contain %d impacting nodes but saw %d: %v", int(nodeID), len(containedNodes), int(nodeImpact.Cardinality()), nodeImpact.Slice())
	}

	for _, containedNode := range containedNodes {
		require.Truef(t, nodeImpact.Contains(containedNode), "Expected node %d to contain node %d. Impact for node 0: %v", int(nodeID), int(containedNode), nodeImpact.Slice())
	}
}

func TestAggregator_Impact(t *testing.T) {
	resetNextID()

	var (
		node0  = node(aKind)
		node1  = node(aKind)
		node2  = node(aKind)
		node3  = node(aKind)
		node4  = node(aKind)
		node5  = node(aKind)
		node6  = node(aKind)
		node7  = node(aKind)
		node8  = node(aKind)
		node9  = node(aKind)
		node10 = node(aKind)
		node11 = node(aKind)

		rootSegment = graph.NewRootPathSegment(node0)

		node1Segment      = descend(rootSegment, node1)
		node3Segment      = descend(node1Segment, node3)
		node5Segment      = descend(node3Segment, node5)
		node8Segment      = descend(node5Segment, node8)
		node8to10Shortcut = descend(node8Segment, node10)

		node6Segment     = descend(node3Segment, node6)
		node6to7Shortcut = descend(node6Segment, node7)

		node11Segment     = descend(rootSegment, node11)
		node11to4Shortcut = descend(node11Segment, node4)

		node2Segment = descend(rootSegment, node2)
		node4Segment = descend(node2Segment, node4)
		node7Segment = descend(node4Segment, node7)
		node9Segment = descend(node7Segment, node9)

		node2to3Shortcut = descend(node2Segment, node3)
		node7to3Shortcut = descend(node7Segment, node3)

		// Node 10 is Terminal for the node9 and node11 segments
		node9to10Terminal  = descend(node9Segment, node10)
		node11to10Terminal = descend(node11Segment, node10)

		// Make sure to use an exact cardinality container (bitset in this case)
		agg = impact.NewAggregator(func() cardinality.Provider[uint64] {
			return cardinality.NewBitmap64()
		})
	)

	agg.AddPath(node9to10Terminal)
	agg.AddPath(node11to10Terminal)

	agg.AddShortcut(node2to3Shortcut)
	agg.AddShortcut(node11to4Shortcut)
	agg.AddShortcut(node6to7Shortcut)
	agg.AddShortcut(node7to3Shortcut)
	agg.AddShortcut(node8to10Shortcut)

	// Validate node 2 impact values and resolutions
	requireImpact(t, agg, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// Validate node 1 impact values and resolutions
	requireImpact(t, agg, 1, 3, 5, 6, 7, 8, 9, 10)

	// Validate node 11 impact values and resolutions
	requireImpact(t, agg, 11, 3, 4, 5, 6, 7, 8, 9, 10)

	// Validate cached resolutions are correct for node 2
	requireImpact(t, agg, 2, 3, 4, 5, 6, 7, 8, 9, 10)
}
