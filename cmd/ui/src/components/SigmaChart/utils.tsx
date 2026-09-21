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

import AbstractGraph, { Attributes } from 'graphology-types';

export const getIsHighlightedItemInGraph = (
    graph: AbstractGraph<Attributes, Attributes, Attributes>,
    highlightedItem: string | null
) => {
    if (!highlightedItem) return;
    return graph.hasNode(highlightedItem) || graph.hasEdge(highlightedItem);
};

// Compute which nodes and edges should remain fully visible when an item is selected.
// Nodes: two independent directional BFS passes from the selected node —
//   outbound (follows edges pointing away) and inbound (follows edges pointing toward).
//   This highlights the whole directed path in both directions without mixing traversal directions.
// Edges: all edges directly connected to above nodes.

export const getFullPathHighlightedEntities = (
    graph: AbstractGraph<Attributes, Attributes, Attributes>,
    highlightedItem: string | null
) => {
    const highlightedNodeIds = new Set<string>();
    const highlightedEdgeIds = new Set<string>();

    if (highlightedItem) {
        if (graph.hasNode(highlightedItem)) {
            highlightedNodeIds.add(highlightedItem);

            // Outbound BFS: follow edges pointing away (source → target).
            const outboundQueue: string[] = [highlightedItem];
            while (outboundQueue.length > 0) {
                const current = outboundQueue.shift()!;
                graph.outEdges(current).forEach((edge) => {
                    highlightedEdgeIds.add(edge);
                    const neighbor = graph.target(edge);
                    if (!highlightedNodeIds.has(neighbor)) {
                        highlightedNodeIds.add(neighbor);
                        outboundQueue.push(neighbor);
                    }
                });
            }

            // Inbound BFS: follow edges pointing toward (target → source).
            // Uses its own visited set so outbound discoveries don't cause early stops.
            const inboundVisited = new Set<string>([highlightedItem]);
            const inboundQueue: string[] = [highlightedItem];
            while (inboundQueue.length > 0) {
                const current = inboundQueue.shift()!;
                graph.inEdges(current).forEach((edge) => {
                    highlightedEdgeIds.add(edge);
                    const neighbor = graph.source(edge);
                    if (!inboundVisited.has(neighbor)) {
                        inboundVisited.add(neighbor);
                        highlightedNodeIds.add(neighbor);
                        inboundQueue.push(neighbor);
                    }
                });
            }
        } else if (graph.hasEdge(highlightedItem)) {
            highlightedEdgeIds.add(highlightedItem);
            graph.extremities(highlightedItem).forEach((directNodes) => highlightedNodeIds.add(directNodes));
        }
    }

    return { highlightedNodeIds, highlightedEdgeIds };
};
