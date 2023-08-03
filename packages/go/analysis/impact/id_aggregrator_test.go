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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func TestAggregator_Cardinality(t *testing.T) {
	resetNextID()

	var (
		node0  = getNextID()
		node1  = getNextID()
		node2  = getNextID()
		node3  = getNextID()
		node4  = getNextID()
		node5  = getNextID()
		node6  = getNextID()
		node7  = getNextID()
		node8  = getNextID()
		node9  = getNextID()
		node10 = getNextID()
		node11 = getNextID()

		rootSegment = graph.NewRootIDSegment(node0)

		node1Segment      = idDescend(rootSegment, node1)
		node3Segment      = idDescend(node1Segment, node3)
		node5Segment      = idDescend(node3Segment, node5)
		node8Segment      = idDescend(node5Segment, node8)
		node8to10Shortcut = idDescend(node8Segment, node10)

		node6Segment     = idDescend(node3Segment, node6)
		node6to7Shortcut = idDescend(node6Segment, node7)

		node11Segment     = idDescend(rootSegment, node11)
		node11to4Shortcut = idDescend(node11Segment, node4)

		node2Segment = idDescend(rootSegment, node2)
		node4Segment = idDescend(node2Segment, node4)
		node7Segment = idDescend(node4Segment, node7)
		node9Segment = idDescend(node7Segment, node9)

		node2to3Shortcut = idDescend(node2Segment, node3)
		node7to3Shortcut = idDescend(node7Segment, node3)

		// Node 10 is Terminal for the node9 and node11 segments
		node9to10Terminal  = idDescend(node9Segment, node10)
		node11to10Terminal = idDescend(node11Segment, node10)

		// Make sure to use an exact cardinality container (bitset in this case)
		agg = impact.NewIDA(func() cardinality.Provider[uint32] {
			return cardinality.NewBitmap32()
		})
	)

	agg.AddPath(node9to10Terminal)
	agg.AddPath(node11to10Terminal)

	agg.AddShortcut(node2to3Shortcut)
	agg.AddShortcut(node11to4Shortcut)
	agg.AddShortcut(node6to7Shortcut)
	agg.AddShortcut(node7to3Shortcut)
	agg.AddShortcut(node8to10Shortcut)

	nodeImpact := agg.Cardinality(2).(cardinality.Duplex[uint32])

	assert.Equal(t, 4, int(agg.Resolved().Cardinality()))

	assert.True(t, agg.Resolved().Contains(2))
	assert.True(t, agg.Resolved().Contains(3))
	assert.True(t, agg.Resolved().Contains(7))
	assert.True(t, agg.Resolved().Contains(10))

	require.Equal(t, 8, int(nodeImpact.Cardinality()))

	require.True(t, nodeImpact.Contains(4))
	require.True(t, nodeImpact.Contains(7))
	require.True(t, nodeImpact.Contains(9))
	require.True(t, nodeImpact.Contains(10))
	require.True(t, nodeImpact.Contains(3))
	require.True(t, nodeImpact.Contains(5))
	require.True(t, nodeImpact.Contains(6))
	require.True(t, nodeImpact.Contains(8))

	nodeImpact = agg.Cardinality(1).(cardinality.Duplex[uint32])

	require.Equal(t, 5, int(agg.Resolved().Cardinality()))

	require.True(t, agg.Resolved().Contains(1))
	require.True(t, agg.Resolved().Contains(2))
	require.True(t, agg.Resolved().Contains(3))
	require.True(t, agg.Resolved().Contains(7))
	require.True(t, agg.Resolved().Contains(10))

	require.Equal(t, 7, int(nodeImpact.Cardinality()))

	require.True(t, nodeImpact.Contains(3))
	require.True(t, nodeImpact.Contains(5))
	require.True(t, nodeImpact.Contains(6))
	require.True(t, nodeImpact.Contains(7))
	require.True(t, nodeImpact.Contains(8))
	require.True(t, nodeImpact.Contains(9))
	require.True(t, nodeImpact.Contains(10))

	nodeImpact = agg.Cardinality(11).(cardinality.Duplex[uint32])

	require.Equal(t, 7, int(agg.Resolved().Cardinality()))

	require.True(t, agg.Resolved().Contains(1))
	require.True(t, agg.Resolved().Contains(2))
	require.True(t, agg.Resolved().Contains(3))
	require.True(t, agg.Resolved().Contains(4))
	require.True(t, agg.Resolved().Contains(7))
	require.True(t, agg.Resolved().Contains(10))
	require.True(t, agg.Resolved().Contains(11))

	require.Equal(t, 8, int(nodeImpact.Cardinality()))

	require.True(t, nodeImpact.Contains(3))
	require.True(t, nodeImpact.Contains(4))
	require.True(t, nodeImpact.Contains(5))
	require.True(t, nodeImpact.Contains(6))
	require.True(t, nodeImpact.Contains(7))
	require.True(t, nodeImpact.Contains(8))
	require.True(t, nodeImpact.Contains(9))
	require.True(t, nodeImpact.Contains(10))

	// Validate cached resolutions are correct
	nodeImpact = agg.Cardinality(2).(cardinality.Duplex[uint32])

	require.Equal(t, 8, int(nodeImpact.Cardinality()))

	require.True(t, nodeImpact.Contains(4))
	require.True(t, nodeImpact.Contains(7))
	require.True(t, nodeImpact.Contains(9))
	require.True(t, nodeImpact.Contains(10))
	require.True(t, nodeImpact.Contains(3))
	require.True(t, nodeImpact.Contains(5))
	require.True(t, nodeImpact.Contains(6))
	require.True(t, nodeImpact.Contains(8))
}
