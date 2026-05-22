import { MultiDirectedGraph } from 'graphology';
import { getFullPathHighlightedEntities, getIsHighlightedItemInGraph } from './utils';

// Linear directed graph with a split inbound on node-c:
//   node-a → node-b → node-c ← node-d
const buildTestGraph = () => {
    const graph = new MultiDirectedGraph();
    graph.addNode('node-a');
    graph.addNode('node-b');
    graph.addNode('node-c');
    graph.addNode('node-d');
    graph.addDirectedEdgeWithKey('edge-ab', 'node-a', 'node-b');
    graph.addDirectedEdgeWithKey('edge-bc', 'node-b', 'node-c');
    graph.addDirectedEdgeWithKey('edge-dc', 'node-d', 'node-c');
    return graph;
};

describe('SigmaChart Utils', () => {
    let graph: MultiDirectedGraph;

    beforeEach(() => {
        graph = buildTestGraph();
    });

    describe('getIsHighlightedItemInGraph', () => {
        it('returns true when the highlighted item is a node in the graph', () => {
            expect(getIsHighlightedItemInGraph(graph, 'node-a')).toBe(true);
        });

        it('returns true when the highlighted item is an edge in the graph', () => {
            expect(getIsHighlightedItemInGraph(graph, 'edge-ab')).toBe(true);
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
                const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(graph, 'node-b');

                expect([...highlightedNodeIds]).toEqual(expect.arrayContaining(['node-a', 'node-b', 'node-c']));
                expect([...highlightedEdgeIds]).toEqual(expect.arrayContaining(['edge-ab', 'edge-bc']));
            });

            it('returns only the outbound path when selecting the start node', () => {
                const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(graph, 'node-a');

                expect([...highlightedNodeIds]).toEqual(expect.arrayContaining(['node-a', 'node-b', 'node-c']));
                expect([...highlightedEdgeIds]).toEqual(expect.arrayContaining(['edge-ab', 'edge-bc']));
            });

            it('returns all inbound paths when selecting the end node with multiple inbound edges', () => {
                const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(graph, 'node-c');

                expect([...highlightedNodeIds]).toEqual(
                    expect.arrayContaining(['node-a', 'node-b', 'node-c', 'node-d'])
                );
                expect([...highlightedEdgeIds]).toEqual(expect.arrayContaining(['edge-ab', 'edge-bc', 'edge-dc']));
            });

            it('returns only the selected node when it has no connections', () => {
                const isolatedGraph = new MultiDirectedGraph();
                isolatedGraph.addNode('node-isolated');
                const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(
                    isolatedGraph,
                    'node-isolated'
                );

                expect([...highlightedNodeIds]).toEqual(['node-isolated']);
                expect(highlightedEdgeIds.size).toBe(0);
            });
        });

        describe('when an edge is selected', () => {
            it('returns the edge and both of its endpoint nodes', () => {
                const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(graph, 'edge-ab');

                expect([...highlightedNodeIds]).toEqual(expect.arrayContaining(['node-a', 'node-b']));
                expect([...highlightedEdgeIds]).toEqual(['edge-ab']);
            });
        });

        it('returns empty sets when highlightedItem is null', () => {
            const { highlightedNodeIds, highlightedEdgeIds } = getFullPathHighlightedEntities(graph, null);

            expect(highlightedNodeIds.size).toBe(0);
            expect(highlightedEdgeIds.size).toBe(0);
        });
    });
});
