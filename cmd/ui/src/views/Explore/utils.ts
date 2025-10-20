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

import { Theme } from '@mui/material';
import { GLYPHS, GetIconInfo, GlyphKind, IconDictionary, getGlyphFromKinds } from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { random } from 'graphology-layout';
import forceAtlas2 from 'graphology-layout-forceatlas2';
import { GraphData, GraphEdge, GraphEdges, GraphNodes } from 'js-client-library';
import { RankDirection, layoutDagre } from 'src/hooks/useLayoutDagre/useLayoutDagre';
import { EdgeParams, NodeParams, ThemedOptions } from 'src/utils';

export const standardLayout = (graph: MultiDirectedGraph) => {
    forceAtlas2.assign(graph, {
        iterations: 128,
        settings: {
            scalingRatio: 1000,
            barnesHutOptimize: true,
        },
    });
};

export const sequentialLayout = (graph: MultiDirectedGraph) => {
    const { assign: assignDagre } = layoutDagre(
        {
            graph: {
                rankdir: RankDirection.LEFT_RIGHT,
                ranksep: 500,
            },
        },
        graph
    );

    assignDagre();
};

type GraphOptions = {
    theme: Theme;
    darkMode: boolean;
    customIcons: IconDictionary;
    hideNodes: boolean;
    tagGlyphMap: Record<string, string>;
    themedOptions?: ThemedOptions;
};

export const initGraph = (items: GraphData, options: GraphOptions) => {
    const graph = new MultiDirectedGraph();

    const { nodes, edges } = items;
    const { theme, darkMode } = options;

    const themedOptions = {
        labels: {
            labelColor: theme.palette.color.primary,
            backgroundColor: theme.palette.neutral.secondary,
            highlightedBackground: theme.palette.color.links,
            highlightedText: darkMode ? theme.palette.common.black : theme.palette.common.white,
        },
        nodeBorderColor: theme.palette.color.primary,
        glyph: {
            colors: {
                backgroundColor: theme.palette.color.primary,
                color: theme.palette.neutral.primary, //border
            },
            tierZeroGlyph: !darkMode ? GLYPHS[GlyphKind.TIER_ZERO_DARK] : GLYPHS[GlyphKind.TIER_ZERO],
            ownedObjectGlyph: darkMode ? GLYPHS[GlyphKind.OWNED_OBJECT_DARK] : GLYPHS[GlyphKind.OWNED_OBJECT],
        },
    };

    initGraphNodes(graph, nodes, { ...options, themedOptions });
    initGraphEdges(graph, edges, themedOptions);

    random.assign(graph, { scale: 1000 });

    // RUN DEFAULT LAYOUT
    sequentialLayout(graph);

    return graph;
};

const initGraphNodes = (
    graph: MultiDirectedGraph,
    nodes: GraphNodes,
    options: GraphOptions & { themedOptions: ThemedOptions }
) => {
    const { themedOptions, customIcons, hideNodes, tagGlyphMap } = options;

    Object.keys(nodes).forEach((key: string) => {
        const node = nodes[key];
        // Set default node parameters
        const nodeParams: Partial<NodeParams> = {
            type: 'combined',
            label: node.label,
            forceLabel: true,
            hidden: hideNodes,
            ...themedOptions.labels,
        };

        const iconInfo = GetIconInfo(node.kind, customIcons);
        nodeParams.color = iconInfo.color;
        nodeParams.iconColor = '#000';
        nodeParams.image = iconInfo.url || '';
        nodeParams.glyphs = [];
        nodeParams.glyphColor = themedOptions.glyph.colors.color;
        nodeParams.glyphBackgroundColor = themedOptions.glyph.colors.backgroundColor;

        const glyphImage = getGlyphFromKinds(node.kinds, tagGlyphMap);
        if (glyphImage) {
            nodeParams.type = 'glyph';
            nodeParams.glyph = glyphImage;
        }

        // Tier zero nodes should be marked with a gem glyph
        if (node.isTierZero) {
            nodeParams.type = 'glyph';
            nodeParams.glyph = themedOptions.glyph.tierZeroGlyph.url || '';
        }

        if (node.isOwnedObject) {
            nodeParams.type = 'glyph';
            nodeParams.glyph = themedOptions.glyph.ownedObjectGlyph.url || '';
        }

        graph.addNode(key, {
            size: 25,
            borderSize: 27,
            borderColor: themedOptions.nodeBorderColor,
            ...nodeParams,
        });
    });
};

const initGraphEdges = (graph: MultiDirectedGraph, edges: GraphEdges, themedOptions: ThemedOptions) => {
    // Group edges with the same start and end nodes into arrays. Should be grouped regardless of direction
    const lookupSet = new Set<string>();
    const groupedEdges = edges.reduce<Record<string, GraphEdges>>((groups, edge) => {
        const groupKey = getEdgeGroupKey(edge);
        const edgeKey = getEdgeKey(edge);

        if (!groups[groupKey]) {
            groups[groupKey] = [];
        }

        // Dedupe edges before they are added to groups so that curved edges get counted and rendered correctly
        if (!lookupSet.has(edgeKey)) {
            lookupSet.add(edgeKey);
            groups[groupKey].push(edge);
        }

        return groups;
    }, {});

    // Loop through our group arrays
    for (const group in groupedEdges) {
        const groupSize = groupedEdges[group].length;

        for (const [i, edge] of groupedEdges[group].entries()) {
            const key = getEdgeKey(edge);

            // Set default values for single edges
            const edgeParams: Partial<EdgeParams> = {
                size: 3,
                type: 'arrow',
                label: edge.label,
                groupPosition: 0,
                groupSize: 1,
                exploreGraphId: edge.exploreGraphId || key,
                forceLabel: true,
                ...themedOptions.labels,
            };

            // Groups with odd-numbered totals should have a straight edge first, then curve the rest
            const edgeShouldBeCurved = groupSize > 1;
            const isSelfEdge = edge.source === edge.target;

            // Handle edge groups that have a mix of directions that edges travel between source and target.
            // We can use the value of the enum to indicate which direction the curve should bend
            // const groupStart = group.split('_')[0];
            // const edgeDirection = groupStart === edge.source ? EdgeDirection.FORWARDS : EdgeDirection.BACKWARDS;

            if (edgeShouldBeCurved) {
                edgeParams.type = 'curved';
                edgeParams.curvature = (i + 1) / 4;

                // edgeParams.groupPosition = i;
                // edgeParams.groupSize = groupSize;
                // edgeParams.direction = edgeDirection;
            }
            if (isSelfEdge) {
                edgeParams.type = 'curved';
                // edgeParams.curvature = 0.1;
                // edgeParams.groupPosition = i;
                // edgeParams.groupSize = groupSize;
            }

            graph.addEdgeWithKey(key, edge.source, edge.target, edgeParams);
        }
    }
};

const getEdgeKey = (edge: GraphEdge) => {
    return `${edge.source}_${edge.kind}_${edge.target}`;
};

const getEdgeGroupKey = (edge: GraphEdge) => {
    const identifiers = [edge.source, edge.target].sort();
    return `${identifiers[0]}_${identifiers[1]}`;
};
