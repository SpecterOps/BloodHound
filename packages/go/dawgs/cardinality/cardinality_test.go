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

package cardinality_test

import (
	"testing"

	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestDuplexToGraphIDs(t *testing.T) {
	uint32IDs := []uint32{1, 2, 3, 4, 5}
	duplex := cardinality.NewBitmap32()
	duplex.Add(uint32IDs...)

	ids := cardinality.DuplexToGraphIDs(duplex)

	for _, uint32ID := range uint32IDs {
		found := false

		for _, id := range ids {
			if id.Uint32() == uint32ID {
				found = true
				break
			}
		}

		require.True(t, found)
	}
}

func TestNodeSetToDuplex(t *testing.T) {
	nodes := graph.NodeSet{
		1: &graph.Node{
			ID: 1,
		},
		2: &graph.Node{
			ID: 2,
		},
	}

	duplex := cardinality.NodeSetToDuplex(nodes)

	require.True(t, duplex.Contains(1))
	require.True(t, duplex.Contains(2))
}
