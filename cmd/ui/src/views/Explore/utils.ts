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

import { faGem, faSkull } from '@fortawesome/free-solid-svg-icons';
import {
    EdgeTraversability,
    GLYPH_SCALE,
    GetIconInfo,
    IconDictionary,
    REL_ID_PREFIX,
    TagGlyphs,
    Theme,
    getGlyphFromKinds,
    getModifiedSvgUrlFromIcon,
    getNodeSource,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { random } from 'graphology-layout';
import forceAtlas2 from 'graphology-layout-forceatlas2';
import { GraphNode } from 'js-client-library';
import { Glyph, GlyphLocation } from 'src/rendering/programs/node.glyphs';
import { RankDirection, setDagreLayout } from 'src/rendering/utils/dagre';
import { EdgeDirection, EdgeParams, NodeParams, ThemedOptions } from 'src/utils';
import type { SigmaExploreEdge, SigmaExploreItems } from './graphCollapse/projectSigmaCollapse';

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
    const { assign: assignDagre } = setDagreLayout(
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
    tagGlyphs: TagGlyphs;
    pzFeatureFlagEnabled: boolean | undefined;
    themedOptions?: ThemedOptions;
    edgeTraversability?: EdgeTraversability;
};

export const initGraph = (items: SigmaExploreItems, options: GraphOptions) => {
    const graph = new MultiDirectedGraph();

    const { nodes, edges, summaryNodes } = items;
    const { theme } = options;
    const themedOptions = {
        labels: {
            labelColor: theme.contrast,
            backgroundColor: theme.neutral.primary,
            highlightedBackground: theme.link,
            highlightedText: theme.neutral.primary,
        },
        nodeBorderColor: theme.contrast,
        glyph: {
            colors: {
                backgroundColor: theme.contrast,
                color: theme.neutral.primary, //border
            },
        },
    };

    initGraphNodes(graph, nodes, { ...options, themedOptions }, summaryNodes);
    initGraphEdges(graph, edges, { ...options, themedOptions });

    random.assign(graph, { scale: 1000 });

    // RUN DEFAULT LAYOUT
    sequentialLayout(graph);

    return graph;
};

const GLYPHS = {
    TIER_ZERO: getModifiedSvgUrlFromIcon(faGem, {
        styles: { color: '#FFFFFF', scale: GLYPH_SCALE },
    }),
    TIER_ZERO_DARK: getModifiedSvgUrlFromIcon(faGem, {
        styles: { color: '#000000', scale: GLYPH_SCALE },
    }),
    OWNED_OBJECT: getModifiedSvgUrlFromIcon(faSkull, {
        styles: { color: '#FFFFFF', scale: GLYPH_SCALE },
    }),
    OWNED_OBJECT_DARK: getModifiedSvgUrlFromIcon(faSkull, {
        styles: { color: '#000000', scale: GLYPH_SCALE },
    }),
};

export const getNodeGlyphs = (
    node: GraphNode,
    options: GraphOptions & { themedOptions: ThemedOptions },
    standardGlyphs: typeof GLYPHS
) => {
    const { themedOptions, pzFeatureFlagEnabled, tagGlyphs } = options;
    const glyphs: Glyph[] = [];

    if (pzFeatureFlagEnabled) {
        if (node.kinds.includes(tagGlyphs.owned)) {
            glyphs.push({
                location: GlyphLocation.BOTTOM_RIGHT,
                image: tagGlyphs.ownedGlyph,
                ...themedOptions.glyph.colors,
            });
        }

        const glyphImage = getGlyphFromKinds(node.kinds, tagGlyphs);
        if (glyphImage) {
            glyphs.push({
                location: GlyphLocation.TOP_RIGHT,
                image: glyphImage,
                ...themedOptions.glyph.colors,
            });
        }
    } else {
        if (node.isOwnedObject) {
            glyphs.push({
                location: GlyphLocation.BOTTOM_RIGHT,
                image: options.darkMode ? standardGlyphs.OWNED_OBJECT_DARK : standardGlyphs.OWNED_OBJECT,
                ...themedOptions.glyph.colors,
            });
        }

        if (node.isTierZero) {
            glyphs.push({
                location: GlyphLocation.TOP_RIGHT,
                image: options.darkMode ? standardGlyphs.TIER_ZERO_DARK : standardGlyphs.TIER_ZERO,
                ...themedOptions.glyph.colors,
            });
        }
    }

    return glyphs;
};

const getCountGlyphImage = (count: number, darkMode: boolean) => {
    const text = String(count);
    const fontSize = Math.min(40, 84 / text.length);
    const color = darkMode ? 'black' : 'white';
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64"><text x="32" y="32" text-anchor="middle" dominant-baseline="central" font-family="sans-serif" font-size="${fontSize}" font-weight="bold" fill="${color}">${text}</text></svg>`;
    return `data:image/svg+xml,${encodeURIComponent(svg)}`;
};

const initGraphNodes = (
    graph: MultiDirectedGraph,
    nodes: SigmaExploreItems['nodes'],
    options: GraphOptions & { themedOptions: ThemedOptions },
    summaryNodes: SigmaExploreItems['summaryNodes']
) => {
    const { themedOptions, customIcons, hideNodes } = options;

    Object.keys(nodes).forEach((key: string) => {
        const node = nodes[key];
        // Set default node parameters
        const nodeParams: Partial<NodeParams> = {
            type: 'combined',
            label: node.label,
            source: getNodeSource(node.kinds),
            kind: node.kind,
            forceLabel: true,
            hidden: hideNodes,
            ...themedOptions.labels,
        };

        const iconInfo = GetIconInfo(node.kind, customIcons);
        nodeParams.color = iconInfo.color;
        nodeParams.image = iconInfo.url || '';

        const glyphs = getNodeGlyphs(node, options, GLYPHS);
        const summary = summaryNodes?.get(key);
        if (summary) {
            glyphs.push({
                location: GlyphLocation.TOP_LEFT,
                image: getCountGlyphImage(summary.count, options.darkMode),
                ...themedOptions.glyph.colors,
            });
        }
        if (glyphs.length > 0) {
            nodeParams.glyphs = glyphs;
            nodeParams.type = 'glyphs';
        }

        graph.addNode(key, {
            size: 25,
            borderColor: themedOptions.nodeBorderColor,
            ...nodeParams,
            ...(summary && { collapseGroupId: summary.groupId, collapseCount: summary.count }),
        });
    });
};

const initGraphEdges = (
    graph: MultiDirectedGraph,
    edges: SigmaExploreEdge[],
    options: GraphOptions & { themedOptions: ThemedOptions }
) => {
    const { themedOptions, edgeTraversability } = options;

    // Group edges with the same start and end nodes into arrays. Should be grouped regardless of direction
    const lookupSet = new Set<string>();
    const groupedEdges = edges.reduce<Record<string, { edge: SigmaExploreEdge; id: string }[]>>(
        (groups, edge, index) => {
            const groupKey = getEdgeGroupKey(edge);
            const id = getSigmaEdgeId(edge, index);

            if (!groups[groupKey]) {
                groups[groupKey] = [];
            }

            // Dedupe edges before they are added to groups so that curved edges get counted and rendered correctly
            if (!lookupSet.has(id)) {
                lookupSet.add(id);
                groups[groupKey].push({ edge, id });
            }

            return groups;
        },
        {}
    );

    // Loop through our group arrays
    for (const group in groupedEdges) {
        const groupSize = groupedEdges[group].length;

        for (const [i, { edge, id }] of groupedEdges[group].entries()) {
            const key = getEdgeKey(edge);

            // Set default values for single edges
            const edgeParams: Partial<EdgeParams> = {
                size: 4,
                color: options.darkMode ? '#6c6c6c' : '#55595C',
                type: 'arrow',
                label: edge.label,
                kind: edge.kind,
                dashed: edgeTraversability ? !edgeTraversability.get(edge.kind) : false,
                groupPosition: 0,
                groupSize: 1,
                exploreGraphId: key,
                forceLabel: true,
                ...themedOptions.labels,
            };

            // Groups with odd-numbered totals should have a straight edge first, then curve the rest
            const edgeShouldBeCurved = groupSize > 1;
            const isSelfEdge = edge.source === edge.target;

            // Handle edge groups that have a mix of directions that edges travel between source and target.
            // We can use the value of the enum to indicate which direction the curve should bend
            const groupStart = group.split('_')[0];
            const edgeDirection = groupStart === edge.source ? EdgeDirection.FORWARDS : EdgeDirection.BACKWARDS;

            if (edgeShouldBeCurved) {
                edgeParams.type = 'curved';
                edgeParams.groupPosition = i;
                edgeParams.groupSize = groupSize;
                edgeParams.direction = edgeDirection;
            }
            if (isSelfEdge) {
                edgeParams.type = 'self';
                edgeParams.groupPosition = i;
                edgeParams.groupSize = groupSize;
            }

            graph.addEdgeWithKey(id, edge.source, edge.target, {
                ...edgeParams,
                ...('summaryId' in edge && { collapseGroupId: edge.collapseGroupId }),
            });
        }
    }
};

function getSigmaEdgeId(edge: SigmaExploreEdge, index: number): string {
    if ('summaryId' in edge) return edge.summaryId;
    if (
        edge.exploreGraphId?.startsWith(REL_ID_PREFIX) &&
        /^\d+$/.test(edge.exploreGraphId.slice(REL_ID_PREFIX.length))
    ) {
        return edge.exploreGraphId;
    }
    const id = String(edge.id ?? '');
    return id === '' ? `anonymous-edge:${index}` : REL_ID_PREFIX + id;
}

const getEdgeKey = (edge: SigmaExploreEdge) => {
    return `${edge.source}_${edge.kind}_${edge.target}`;
};

const getEdgeGroupKey = (edge: SigmaExploreEdge) => {
    const identifiers = [edge.source, edge.target].sort();
    return `${identifiers[0]}_${identifiers[1]}`;
};
