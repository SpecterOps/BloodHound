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

package graphcache

import "github.com/specterops/bloodhound/dawgs/graph"

func MaterializeIDPath(idPath graph.IDPath, cache Cache, tx graph.Transaction) (graph.Path, error) {
	var (
		path      = graph.AllocatePath(len(idPath.Edges))
		nodeIndex = make(map[graph.ID]int, len(idPath.Nodes))
		edgeIndex = make(map[graph.ID]int, len(idPath.Edges))
	)

	for idx, node := range idPath.Nodes {
		nodeIndex[node] = idx
	}

	for idx, edge := range idPath.Edges {
		edgeIndex[edge] = idx
	}

	if nodes, err := FetchNodesByID(tx, cache, idPath.Nodes); err != nil {
		return path, err
	} else {
		for _, node := range nodes {
			path.Nodes[nodeIndex[node.ID]] = node
		}
	}

	if edges, err := FetchRelationshipsByID(tx, cache, idPath.Edges); err != nil {
		return path, err
	} else {
		for _, edge := range edges {
			path.Edges[edgeIndex[edge.ID]] = edge
		}
	}

	return path, nil
}
