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

import { MultiDirectedGraph } from 'graphology';
import { buildTestGraph } from 'src/mocks/factories/exploreGraphHighlighting';
import { getFullPathHighlightedEntities, getIsHighlightedItemInGraph } from './utils';

const createEntityArray = (entityList: Set<string>) => [...entityList];

const testPathNodes = ['53069', '51155', '52350'];

const testPathEdges = ['rel_939292', 'rel_931961'];

const testSelectedNode = '52350';

const testSelectedEdge = 'rel_931961';

const testNodesWhenEdgeSelected = ['51155', '52350'];

describe('SigmaChart Utils', () => {
    let graph: MultiDirectedGraph;

    beforeEach(() => {
        graph = buildTestGraph();
    });

    describe('getIsHighlightedItemInGraph', () => {
        it('returns true when the highlighted item is a node in the graph', () => {
            expect(getIsHighlightedItemInGraph(graph, testSelectedNode)).toBe(true);
        });

        it('returns true when the highlighted item is an edge in the graph', () => {
            expect(getIsHighlightedItemInGraph(graph, testSelectedEdge)).toBe(true);
        });

        it('returns undefined when highlightedItem is null', () => {
            expect(getIsHighlightedItemInGraph(graph, null)).toBeUndefined();
        });

        it('returns false when the highlighted item is not in the graph', () => {
            expect(getIsHighlightedItemInGraph(graph, 'node-unknown')).toBe(false);
        });
    });

    describe('getFullPathHighlightedEntities', () => {
        describe('when a node is selected', () => {
            it('returns the full inbound and outbound path when selecting a middle node', () => {
                const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(
                    graph,
                    testSelectedNode
                );

                const nodeIds = createEntityArray(highlightedNodeIds);
                const edgeIds = createEntityArray(highlightedEdgeIds);

                expect(nodeIds).toEqual(expect.arrayContaining(testPathNodes));
                expect(nodeIds).toHaveLength(testPathNodes.length);
                expect(edgeIds).toEqual(expect.arrayContaining(testPathEdges));
                expect(edgeIds).toHaveLength(testPathEdges.length);
            });
        });

        describe('when an edge is selected', () => {
            it('returns the edge and both of its endpoint nodes', () => {
                const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(
                    graph,
                    testSelectedEdge
                );

                const nodeIds = createEntityArray(highlightedNodeIds);
                const edgeIds = createEntityArray(highlightedEdgeIds);

                expect(nodeIds).toEqual(expect.arrayContaining(testNodesWhenEdgeSelected));
                expect(nodeIds).toHaveLength(2);
                expect(edgeIds).toEqual([testSelectedEdge]);
                expect(edgeIds).toHaveLength(1);
            });
        });

        it('returns empty sets when highlightedItem is null', () => {
            const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(graph, null);

            expect(highlightedNodeIds.size).toBe(0);
            expect(highlightedEdgeIds.size).toBe(0);
        });
    });
});
