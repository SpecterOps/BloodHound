// Copyright 2026 Specter Ops, Inc.
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

package v2

import (
	"math"
	"testing"

	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginateNodes(t *testing.T) {
	var nodes = graph.NodeSet{
		2: {ID: 2},
		4: {ID: 4},
		1: {ID: 1},
		3: {ID: 3},
	}

	testCases := []struct {
		Name          string
		Nodes         graph.NodeSet
		Skip          int
		Limit         int
		ExpectedNodes []*graph.Node
		ExpectedError string
	}{
		{
			Name:          "normal page returns nodes in descending ID order",
			Nodes:         nodes,
			Skip:          1,
			Limit:         2,
			ExpectedNodes: []*graph.Node{nodes[3], nodes[2]},
		},
		{
			Name:          "limit exceeding remaining nodes returns the remainder",
			Nodes:         nodes,
			Skip:          2,
			Limit:         3,
			ExpectedNodes: []*graph.Node{nodes[2], nodes[1]},
		},
		{
			Name:  "skip equal to node count returns an empty page",
			Nodes: nodes,
			Skip:  4,
			Limit: 2,
		},
		{
			Name:  "empty input with zero skip returns an empty page",
			Nodes: graph.NodeSet{},
			Limit: 2,
		},
		{
			Name:  "zero limit returns an empty page",
			Nodes: nodes,
			Skip:  1,
		},
		{
			Name:          "negative skip returns an invalid skip error",
			Nodes:         nodes,
			Skip:          -1,
			Limit:         2,
			ExpectedError: "invalid skip: -1",
		},
		{
			Name:          "skip beyond node count returns an invalid skip error",
			Nodes:         nodes,
			Skip:          5,
			Limit:         2,
			ExpectedError: "invalid skip: 5",
		},
		{
			Name:          "negative limit returns an invalid limit error",
			Nodes:         nodes,
			Limit:         -1,
			ExpectedError: "invalid limit: -1",
		},
		{
			Name:          "maximum limit safely returns the remainder",
			Nodes:         nodes,
			Skip:          1,
			Limit:         math.MaxInt,
			ExpectedNodes: []*graph.Node{nodes[3], nodes[2], nodes[1]},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var actualNodes, err = paginateNodes(testCase.Nodes, testCase.Skip, testCase.Limit)

			if testCase.ExpectedError != "" {
				require.EqualError(t, err, testCase.ExpectedError)
				assert.Nil(t, actualNodes)
				return
			}

			require.NoError(t, err)
			require.Len(t, actualNodes, len(testCase.ExpectedNodes))
			for index, expectedNode := range testCase.ExpectedNodes {
				assert.Equal(t, expectedNode, actualNodes[index])
			}
		})
	}
}
