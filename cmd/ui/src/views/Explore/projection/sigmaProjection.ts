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

import {
    createRendererIndex,
    formatCompactCount,
    semanticKey,
    type GroupDefinition,
    type ProjectedGraph,
    type ProjectedPresentation,
    type SemanticGlyph,
    type SourceGraph,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import type { Glyph } from 'src/rendering/programs/node.glyphs';
import { GlyphLocation } from 'src/rendering/programs/node.glyphs';

const getGlyphId = (glyph: Glyph) => semanticKey([glyph.image, glyph.location]);

function countGlyph(count: number, background: string, foreground: string): Glyph {
    const text = formatCompactCount(count);
    const fontSize = Math.min(100, 270 / text.length);
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="192" height="192" viewBox="0 0 192 192"><text x="96" y="100" text-anchor="middle" dominant-baseline="middle" fill="${foreground}" font-family="Arial,sans-serif" font-weight="700" font-size="${fontSize}">${text}</text></svg>`;

    return {
        location: GlyphLocation.TOP_LEFT,
        image: `data:image/svg+xml,${encodeURIComponent(svg)}`,
        backgroundColor: background,
        color: foreground,
    };
}

export function sigmaMemberGlyphs(styledGraph: MultiDirectedGraph): Record<string, SemanticGlyph[]> {
    return Object.fromEntries(
        styledGraph.mapNodes((id, attributes) => [
            id,
            ((attributes.glyphs ?? []) as Glyph[]).map((glyph) => ({
                id: getGlyphId(glyph),
                position:
                    glyph.location === GlyphLocation.TOP_RIGHT
                        ? 'zone'
                        : glyph.location === GlyphLocation.BOTTOM_RIGHT
                          ? 'owned'
                          : 'other',
            })),
        ])
    );
}

/**
 * Apply shared projection membership to Sigma's renderer graph while preserving renderer-owned coordinates.
 * Source graph membership and projected IDs come from the shared projection builder.
 */
export function updateSigmaProjection(
    renderer: MultiDirectedGraph,
    styledGraph: MultiDirectedGraph,
    source: SourceGraph,
    projected: ProjectedGraph,
    presentation: Readonly<Record<string, ProjectedPresentation>>,
    groups: readonly GroupDefinition[]
) {
    const rendererIndex = createRendererIndex(source, projected, true);
    const availableGroups = new Set(groups.map((group) => group.id));

    // Retain valid source and group nodes in renderer storage so their positions survive collapse/expand.
    // Remove stale entries, then hide retained entries until this projection reveals them.
    renderer.forEachNode((id, attributes) => {
        if (
            (!attributes.groupId && !Object.hasOwn(source.nodes, id)) ||
            (attributes.groupId && !availableGroups.has(attributes.groupId))
        ) {
            renderer.dropNode(id);
        } else {
            renderer.setNodeAttribute(id, 'hidden', true);
        }
    });

    renderer.forEachEdge((id) => renderer.setEdgeAttribute(id, 'hidden', true));

    // Refresh colors, glyphs, and labels from current theme options without replacing cached coordinates.
    styledGraph.forEachNode((id, attributes) => {
        const previous = renderer.hasNode(id) ? renderer.getNodeAttributes(id) : undefined;
        renderer.mergeNode(id, {
            ...attributes,
            x: previous?.x ?? attributes.x ?? 0,
            y: previous?.y ?? attributes.y ?? 0,
            hidden: true,
            sourceNodeIds: [id],
        });
    });

    for (const node of Object.values(projected.nodes)) {
        const id = rendererIndex.projectedToRenderer[node.id];
        if (node.kind === 'source-node') {
            renderer.setNodeAttribute(id, 'hidden', false);
            continue;
        }
        if (node.kind !== 'node-group') {
            // BHE owns relationship-group rendering.
            continue;
        }

        const firstMember = styledGraph.getNodeAttributes(node.sourceNodeIds[0]);
        const uniqueGlyphs = new Map<string, Glyph>();
        for (const member of node.sourceNodeIds) {
            for (const glyph of (styledGraph.getNodeAttribute(member, 'glyphs') ?? []) as Glyph[]) {
                uniqueGlyphs.set(getGlyphId(glyph), glyph);
            }
        }

        // Fan out glyphs that share a corner so each one remains visible.
        const glyphCountByLocation = new Map<GlyphLocation, number>();
        const glyphs: Glyph[] = presentation[node.id].glyphs.flatMap((descriptor) => {
            const glyph = uniqueGlyphs.get(descriptor.id);
            if (!glyph) {
                return [];
            }

            const offset = glyphCountByLocation.get(glyph.location) ?? 0;
            glyphCountByLocation.set(glyph.location, offset + 1);
            return { ...glyph, distance: 1 + offset * 1.4 };
        });

        glyphs.push(countGlyph(presentation[node.id].count!, firstMember.borderColor, firstMember.backgroundColor));
        const previous = renderer.hasNode(id) ? renderer.getNodeAttributes(id) : undefined;
        renderer.mergeNode(id, {
            ...firstMember,
            hidden: false,
            representative: true,
            projectionKind: node.kind,
            groupId: node.groupId,
            sourceNodeIds: node.sourceNodeIds,
            type: 'glyphs',
            glyphs,
            label: presentation[node.id].caption,
            x:
                previous?.x ??
                node.sourceNodeIds.reduce((sum, member) => sum + renderer.getNodeAttribute(member, 'x'), 0) /
                    node.sourceNodeIds.length,
            y:
                previous?.y ??
                node.sourceNodeIds.reduce((sum, member) => sum + renderer.getNodeAttribute(member, 'y'), 0) /
                    node.sourceNodeIds.length,
        });
    }

    // Keep original edges hidden in storage so the renderer cache stays stable across projections.
    styledGraph.forEachEdge((id, attributes, sourceId, targetId) => {
        if (renderer.hasEdge(id) && (renderer.source(id) !== sourceId || renderer.target(id) !== targetId)) {
            renderer.dropEdge(id);
        }

        renderer.mergeEdgeWithKey(id, sourceId, targetId, { ...attributes, hidden: true, sourceEdgeIds: [id] });
    });

    for (const edge of Object.values(projected.edges)) {
        const id = rendererIndex.projectedToRenderer[edge.id];
        const sourceId = rendererIndex.projectedToRenderer[edge.source];
        const targetId = rendererIndex.projectedToRenderer[edge.target];
        const sourceEdgeId = edge.sourceEdgeIds[0];
        const attributes = styledGraph.hasEdge(sourceEdgeId) ? styledGraph.getEdgeAttributes(sourceEdgeId) : {};
        if (renderer.hasEdge(id) && (renderer.source(id) !== sourceId || renderer.target(id) !== targetId)) {
            renderer.dropEdge(id);
        }

        // Reset inherited curve styling for redirected edges; visible parallel curves are grouped below.
        const hasRedirectedSourceEdge = edge.sourceEdgeIds.some(
            (sourceEdgeId) =>
                source.edges[sourceEdgeId].source !== sourceId || source.edges[sourceEdgeId].target !== targetId
        );
        renderer.mergeEdgeWithKey(id, sourceId, targetId, {
            ...attributes,
            ...(hasRedirectedSourceEdge ? { type: 'arrow', groupSize: 1, groupPosition: 0 } : {}),
            hidden: false,
            sourceEdgeIds: edge.sourceEdgeIds,
            representative: hasRedirectedSourceEdge,
            groupId: edge.groupId,
        });
    }

    // Recompute curve metadata after projection so redirected parallel relationships remain distinct.
    const parallel = new Map<string, string[]>();
    renderer.forEachEdge((id, attributes, sourceId, targetId) => {
        if (attributes.hidden) {
            return;
        }

        const key = JSON.stringify([sourceId, targetId].sort());
        const members = parallel.get(key) ?? [];
        members.push(id);
        parallel.set(key, members);
    });

    for (const members of parallel.values()) {
        const first = members[0];
        const [sourceId, targetId] = [renderer.source(first), renderer.target(first)].sort();
        members.sort().forEach((id, position) =>
            renderer.mergeEdgeAttributes(id, {
                type: sourceId === targetId ? 'self' : members.length > 1 ? 'curved' : 'arrow',
                groupSize: members.length,
                groupPosition: position,
                direction: renderer.source(id) === sourceId ? 1 : -1,
            })
        );
    }

    renderer.setAttribute('projectionRevision', (renderer.getAttribute('projectionRevision') ?? 0) + 1);
    renderer.setAttribute('sourceGraph', styledGraph);
    renderer.setAttribute('projectionIndex', rendererIndex);

    // Sigma retains its program instances; publish the glyph capacity for their next buffer allocation.
    renderer.setAttribute(
        'glyphCapacity',
        renderer.reduceNodes((capacity, _id, attributes) => Math.max(capacity, attributes.glyphs?.length ?? 0), 4)
    );
    return rendererIndex;
}

/** Bounds can exclude hidden renderer cache objects for explicit fit operations. */
export function graphBounds(
    graph: MultiDirectedGraph,
    visibleOnly = false
): { x: [number, number]; y: [number, number] } | null {
    let bounds: { x: [number, number]; y: [number, number] } | null = null;
    graph.forEachNode((_id, node) => {
        if (visibleOnly && node.hidden) {
            return;
        }

        if (!bounds) {
            bounds = { x: [node.x, node.x], y: [node.y, node.y] };
        } else {
            bounds.x = [Math.min(bounds.x[0], node.x), Math.max(bounds.x[1], node.x)];
            bounds.y = [Math.min(bounds.y[0], node.y), Math.max(bounds.y[1], node.y)];
        }
    });
    return bounds;
}
