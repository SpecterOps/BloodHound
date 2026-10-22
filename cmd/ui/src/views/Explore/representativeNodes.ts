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

import { findRepresentativeNodeSets, formatCompactCount, pluralNodeType, RepresentativeNodeSet } from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { GraphData } from 'js-client-library';
import { Glyph, GlyphLocation } from 'src/rendering/programs/node.glyphs';

export const representativeSets = (data: GraphData) =>
    findRepresentativeNodeSets(
        Object.fromEntries(Object.entries(data.nodes).map(([id, node]) => [id, node.kind])),
        data.edges
    );

/** Keep every distinct glyph; the renderer allocates enough slots for the largest union. */
export function representativeGlyphs(graph: MultiDirectedGraph, members: string[]): Glyph[] {
    const unique = new Map<string, Glyph>();
    for (const member of members) {
        for (const glyph of (graph.getNodeAttribute(member, 'glyphs') ?? []) as Glyph[]) {
            unique.set(JSON.stringify([glyph.image, glyph.backgroundColor, glyph.color]), glyph);
        }
    }
    const occupied = new Map<GlyphLocation, number>();
    return [...unique.values()].map((glyph) => {
        const offset = occupied.get(glyph.location) ?? 0;
        occupied.set(glyph.location, offset + 1);
        return { ...glyph, distance: 1 + offset * 1.4 };
    });
}

export function countGlyph(count: number, background: string, foreground: string): Glyph {
    const text = formatCompactCount(count);
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="192" height="192" viewBox="0 0 192 192"><text x="96" y="100" text-anchor="middle" dominant-baseline="middle" fill="${foreground}" font-family="Arial,sans-serif" font-weight="700" font-size="${Math.min(100, 270 / text.length)}">${text}</text></svg>`;
    return {
        location: GlyphLocation.TOP_LEFT,
        image: `data:image/svg+xml,${encodeURIComponent(svg)}`,
        backgroundColor: background,
        color: foreground,
    };
}

/** Mutate only the display graph, retaining original entities, coordinates and edge IDs. */
export function applyRepresentativeNodes(
    graph: MultiDirectedGraph,
    sets: Record<string, RepresentativeNodeSet>,
    expanded: ReadonlySet<string>,
    selectedItem: string | null
) {
    // Retain synthetic IDs too: Sigma hover/label caches can reference them between frames.
    graph.forEachNode((id, attributes) => graph.setNodeAttribute(id, 'hidden', !!attributes.representative));
    graph.forEachEdge((id, attributes) => graph.setEdgeAttribute(id, 'hidden', !!attributes.representative));
    const selectedNodes =
        selectedItem && graph.hasEdge(selectedItem) ? graph.extremities(selectedItem) : [selectedItem];
    for (const [id, set] of Object.entries(sets)) {
        if (expanded.has(id) || set.members.some((member) => selectedNodes.includes(member))) continue;
        const first = graph.getNodeAttributes(set.members[0]);
        const glyphs = representativeGlyphs(graph, set.members);
        glyphs.push(countGlyph(set.members.length, first.borderColor, first.backgroundColor));
        graph.mergeNode(id, {
            ...first,
            representative: true,
            hidden: false,
            type: 'glyphs',
            glyphs,
            label: `${set.members.length.toLocaleString('en-US')} ${pluralNodeType(first.kind)}`,
            x: set.members.reduce((sum, member) => sum + graph.getNodeAttribute(member, 'x'), 0) / set.members.length,
            y: set.members.reduce((sum, member) => sum + graph.getNodeAttribute(member, 'y'), 0) / set.members.length,
        });
        const edge = graph.edges(set.members[0])[0];
        const [source, target] = graph.extremities(edge);
        graph.mergeEdgeWithKey(
            `representative-edge:${id}`,
            source === set.members[0] ? id : source,
            target === set.members[0] ? id : target,
            { ...graph.getEdgeAttributes(edge), hidden: false, type: 'arrow', representative: true }
        );
        for (const member of set.members) {
            graph.setNodeAttribute(member, 'hidden', true);
            graph.forEachEdge(member, (edgeId) => graph.setEdgeAttribute(edgeId, 'hidden', true));
        }
    }
}

/** Fix normalization to the ordinary node coordinates while visibility changes. */
export function graphBounds(graph: MultiDirectedGraph): { x: [number, number]; y: [number, number] } | null {
    return graph.order
        ? graph.reduceNodes<{ x: [number, number]; y: [number, number] }>(
              (box, _id, node) => ({
                  x: [Math.min(box.x[0], node.x), Math.max(box.x[1], node.x)],
                  y: [Math.min(box.y[0], node.y), Math.max(box.y[1], node.y)],
              }),
              { x: [Infinity, -Infinity], y: [Infinity, -Infinity] }
          )
        : null;
}
