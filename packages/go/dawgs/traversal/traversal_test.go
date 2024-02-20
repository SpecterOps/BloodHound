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

package traversal

import (
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

var (
	kindA = graph.StringKind("a")
	kindB = graph.StringKind("b")
	kindR = graph.StringKind("r")

	node0 = graph.NewNode(0, nil, kindA)
	node1 = graph.NewNode(1, nil, kindB)
	node2 = graph.NewNode(2, nil, kindB)
	node3 = graph.NewNode(3, nil, kindB)

	root = graph.NewRootPathSegment(node0)

	// node1Segment: (node0) <-[kindR]- (node1)
	node1Segment = root.Descend(node1, graph.NewRelationship(100, 0, 1, nil, kindR))

	// node2Segment: (node0) <-[kindR]- (node2)
	node2Segment = root.Descend(node2, graph.NewRelationship(101, 0, 2, nil, kindR))

	// node1Node3Segment: (node0) <-[kindR]- (node1) <-[kindR]- (node3)
	node1Node3Segment = node1Segment.Descend(node3, graph.NewRelationship(102, 1, 3, nil, kindR))

	// node2Node3Segment: (node0) <-[kindR]- (node2) <-[kindR]- (node3)
	node2Node3Segment = node2Segment.Descend(node3, graph.NewRelationship(103, 2, 3, nil, kindR))

	// cycleSegment: (node0) <-[kindR]- (node1) <-[kindR]- (node0)
	cycleSegment = node1Segment.Descend(node0, graph.NewRelationship(104, 1, 0, nil, kindR))
)

func TestAcyclicSegmentVisitor(t *testing.T) {
	visitor := AcyclicNodeFilter(func(next *graph.PathSegment) bool {
		return true
	})

	// Disallow cycles
	require.False(t, visitor(cycleSegment))
}

func TestUniquePathSegmentVisitor(t *testing.T) {
	visitor := UniquePathSegmentFilter(func(next *graph.PathSegment) bool {
		return true
	})

	// Visiting the segment for the first time should pass
	require.True(t, visitor(node1Node3Segment))

	// Allow traversal to the same node via different paths
	require.True(t, visitor(node2Node3Segment))

	// Disallow retraversal of the same path
	require.False(t, visitor(node2Node3Segment))

	// Disallow cycles
	require.False(t, visitor(cycleSegment))
}

func TestFilteredSkipLimit(t *testing.T) {
	var nodes []*graph.Node

	visitor := FilteredSkipLimit(
		func(next *graph.PathSegment) (bool, bool) {
			return next.Node.ID == 3, next.Node.ID == 3
		},
		func(next *graph.PathSegment) {
			nodes = append(nodes, next.Node)
		},
		1,
		1)

	// Skip and descend
	require.True(t, visitor(node1Node3Segment))

	// Reject descent of node that doesn't match
	require.False(t, visitor(node1Segment))

	// Collect and descend
	require.True(t, visitor(node2Node3Segment))

	// At limit, reject descent
	require.False(t, visitor(node1Node3Segment))

	// Validate that we've collected exactly one node
	require.Equal(t, 1, len(nodes))
	require.Equal(t, nodes[0].ID, graph.ID(3))
}
