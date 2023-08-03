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

package graph_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/dawgs/util/test"
)

var (
	groupKind      = graph.StringKind("group")
	domainKind     = graph.StringKind("domain")
	userKind       = graph.StringKind("user")
	computerKind   = graph.StringKind("computer")
	permissionKind = graph.StringKind("permission")
	membershipKind = graph.StringKind("member")
)

func TestPathSegment_Root(t *testing.T) {
	var (
		domainNode   = test.Node(domainKind)
		groupNode    = test.Node(groupKind)
		userNode     = test.Node(userKind)
		computerNode = test.Node(computerKind)

		expectedStartNodes         = []*graph.Node{domainNode, groupNode, userNode}
		expectedStartNodesReversed = []*graph.Node{userNode, groupNode, domainNode}

		domainSegment   = graph.NewRootPathSegment(domainNode)
		groupSegment    = domainSegment.Descend(groupNode, test.Edge(groupNode, domainNode, permissionKind))
		userSegment     = groupSegment.Descend(userNode, test.Edge(userNode, groupNode, membershipKind))
		computerSegment = userSegment.Descend(computerNode, test.Edge(computerNode, userNode, permissionKind))
		inst            = computerSegment.Path()
	)

	require.Equal(t, domainNode, inst.Root())
	require.Equal(t, computerNode, inst.Terminal())

	walkIdx := 0
	inst.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
		require.Equal(t, expectedStartNodes[walkIdx].ID, start.ID)
		walkIdx++

		return true
	})

	walkIdx = 0
	inst.WalkReverse(func(start, end *graph.Node, relationship *graph.Relationship) bool {
		require.Equal(t, expectedStartNodesReversed[walkIdx].ID, start.ID)
		walkIdx++

		return true
	})

	seen := false
	inst.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
		if seen {
			t.Fatal("Expected to be called only once.")
		} else {
			seen = true
		}

		return false
	})

	seen = false
	inst.WalkReverse(func(start, end *graph.Node, relationship *graph.Relationship) bool {
		if seen {
			t.Fatal("Expected to be called only once.")
		} else {
			seen = true
		}

		return false
	})
}

func TestPathSegment_SizeOf(t *testing.T) {
	var (
		domainNode   = test.Node(domainKind)
		groupNode    = test.Node(groupKind)
		userNode     = test.Node(userKind)
		computerNode = test.Node(computerKind)

		domainSegment = graph.NewRootPathSegment(domainNode)
		originalSize  = int64(domainSegment.SizeOf())
		treeSize      = originalSize
	)

	require.Equal(t, treeSize, int64(domainSegment.SizeOf()))

	// Group segment
	groupSegment := domainSegment.Descend(groupNode, test.Edge(groupNode, domainNode, permissionKind))

	// Add the size of the path edge but also ensure that the capacity increase for storing branches is also tracked
	groupSegmentSize := int64(groupSegment.SizeOf())

	treeSize += groupSegmentSize
	treeSize += int64(size.Of(domainSegment.Branches))

	require.Equal(t, treeSize, int64(domainSegment.SizeOf()))

	// User segment
	userSegment := groupSegment.Descend(userNode, test.Edge(userNode, groupNode, membershipKind))

	// Add the size of the path edge but also ensure that the capacity increase for storing branches is also tracked
	userSegmentSize := int64(userSegment.SizeOf())

	treeSize += userSegmentSize
	treeSize += int64(size.Of(groupSegment.Branches))

	require.Equal(t, treeSize, int64(domainSegment.SizeOf()))

	// Computer segment
	computerSegment := userSegment.Descend(computerNode, test.Edge(computerNode, userNode, permissionKind))

	// Add the size of the path edge but also ensure that the capacity increase for storing branches is also tracked
	computerSegmentSize := int64(computerSegment.SizeOf())

	treeSize += computerSegmentSize
	treeSize += int64(size.Of(userSegment.Branches))

	require.Equal(t, treeSize, int64(domainSegment.SizeOf()))

	// Test detaching from the path tree
	computerSegment.Detach()
	treeSize -= computerSegmentSize

	require.Equal(t, treeSize, int64(domainSegment.SizeOf()))

	userSegment.Detach()

	treeSize -= userSegmentSize
	treeSize -= int64(size.Of(userSegment.Branches))

	require.Equal(t, treeSize, int64(domainSegment.SizeOf()))

	groupSegment.Detach()

	treeSize -= groupSegmentSize
	treeSize -= int64(size.Of(groupSegment.Branches))

	require.Equal(t, treeSize, int64(domainSegment.SizeOf()))

	// The original size should have one slice allocation
	originalSize += int64(size.Of(domainSegment.Branches))
	require.Equal(t, treeSize, originalSize)
}

func TestIDSegment_SizeOf(t *testing.T) {
	var (
		rootSegment        = graph.NewRootIDSegment(0)
		sizeOfEmptySegment = rootSegment.SizeOf()
	)

	// Create a descending group from the root
	groupSegment := rootSegment.Descend(1, 2)
	groupSegmentSize := groupSegment.SizeOf()

	// One descent means one allocation plus the 8 byte pointer to it
	require.Equal(t, int(sizeOfEmptySegment*2)+8, int(rootSegment.SizeOf()))

	// All single-segment sizes without branches should have the same size as a new root segment
	require.Equal(t, sizeOfEmptySegment, groupSegmentSize)

	// Emulate the two emulated membership edges
	userSegment := groupSegment.Descend(3, 4)
	userSegmentSize := userSegment.SizeOf()

	require.Equal(t, int(sizeOfEmptySegment*3)+cap(rootSegment.Branches)*8+cap(groupSegment.Branches)*8, int(rootSegment.SizeOf()))
	require.Equal(t, int(sizeOfEmptySegment*2)+cap(groupSegment.Branches)*8, int(groupSegment.SizeOf()))
	require.Equal(t, sizeOfEmptySegment, userSegmentSize)

	computerSegment := groupSegment.Descend(5, 6)
	computerSegmentSize := computerSegment.SizeOf()

	require.Equal(t, int(sizeOfEmptySegment*4)+cap(rootSegment.Branches)*8+cap(groupSegment.Branches)*8, int(rootSegment.SizeOf()))
	require.Equal(t, int(sizeOfEmptySegment*3)+cap(groupSegment.Branches)*8, int(groupSegment.SizeOf()))
	require.Equal(t, sizeOfEmptySegment, computerSegmentSize)

	// Test detaching nodes
	computerSegment.Detach()

	require.Equal(t, int(sizeOfEmptySegment*3)+cap(rootSegment.Branches)*8+cap(groupSegment.Branches)*8, int(rootSegment.SizeOf()))
	require.Equal(t, int(sizeOfEmptySegment*2)+cap(groupSegment.Branches)*8, int(groupSegment.SizeOf()))

	userSegment.Detach()

	require.Equal(t, int(sizeOfEmptySegment*2)+cap(rootSegment.Branches)*8+cap(groupSegment.Branches)*8, int(rootSegment.SizeOf()))
	require.Equal(t, int(sizeOfEmptySegment)+cap(groupSegment.Branches)*8, int(groupSegment.SizeOf()))

	groupSegment.Detach()

	require.Equal(t, int(sizeOfEmptySegment)+cap(rootSegment.Branches)*8, int(rootSegment.SizeOf()))
}
