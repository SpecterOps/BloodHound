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

import { GlyphKind } from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { GraphEdges, GraphNodes } from 'js-client-library';
import { GlyphLocation } from 'src/rendering/programs/node.glyphs';
import { EdgeDirection, EdgeParams, NodeParams } from 'src/utils';
import { GLYPHS, NODE_ICON, UNKNOWN_ICON } from './svgIcons';

export const initGraphNodes = (graph: MultiDirectedGraph, nodes: GraphNodes, nodeSize: number) => {
    Object.keys(nodes).forEach((key: string) => {
        const node = nodes[key];
        // Set default node parameters
        const nodeParams: Partial<NodeParams> = {
            color: '#FFFFFF',
            type: 'combined',
            label: node.label,
            forceLabel: true,
        };

        const icon = NODE_ICON[node.kind] || UNKNOWN_ICON;
        nodeParams.color = icon.color;
        nodeParams.image = icon.url || '';

        // Tier zero nodes should be marked with a gem glyph
        if (node.isTierZero) {
            const glyph = GLYPHS[GlyphKind.TIER_ZERO];
            nodeParams.type = 'glyphs';
            nodeParams.glyphs = [
                {
                    location: GlyphLocation.TOP_RIGHT,
                    image: glyph.url || '',
                    backgroundColor: glyph.color,
                },
            ];
        }

        graph.addNode(key, {
            size: nodeSize,
            borderColor: '#000000',
            ...nodeParams,
        });
    });
};

export const initGraphEdges = (graph: MultiDirectedGraph, edges: GraphEdges) => {
    // Group edges with the same start and end nodes into arrays. Should be grouped regardless of direction
    const groupedEdges = edges.reduce<Record<string, GraphEdges>>((groups, edge) => {
        const identifiers = [edge.source, edge.target].sort();
        const id = `${identifiers[0]}_${identifiers[1]}`;

        if (!groups[id]) {
            groups[id] = [];
        }
        groups[id].push(edge);

        return groups;
    }, {});

    // Loop through our group arrays
    for (const group in groupedEdges) {
        const groupSize = groupedEdges[group].length;

        for (const [i, edge] of groupedEdges[group].entries()) {
            const key = `${edge.source}_${edge.kind}_${edge.target}`;

            // Set default values for single edges
            const edgeParams: Partial<EdgeParams> = {
                size: 3,
                type: 'arrow',
                label: edge.label,
                color: '#000000C0',
                groupPosition: 0,
                groupSize: 1,
                exploreGraphId: edge.exploreGraphId || key,
                forceLabel: true,
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

            graph.addEdgeWithKey(key, edge.source, edge.target, edgeParams);
        }
    }
};
