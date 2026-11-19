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

import { formatCompactCount, pluralNodeType } from '../utils/graphSimplification';
import type { ProjectedGraph, SourceGraph } from './types';

export type SemanticGlyph = Readonly<{ id: string; position: 'zone' | 'owned' | 'other' }>;

export type ProjectedPresentation = Readonly<{
    caption: string;
    nodeKind?: string;
    count?: number;
    badge?: string;
    glyphs: readonly SemanticGlyph[];
    highlighted: boolean;
}>;

export function deriveProjectionPresentation(
    source: SourceGraph,
    graph: ProjectedGraph,
    memberGlyphs: Readonly<Record<string, readonly SemanticGlyph[]>>,
    highlightedItems: ReadonlySet<string>
): Readonly<Record<string, ProjectedPresentation>> {
    const result: Record<string, ProjectedPresentation> = Object.create(null);
    for (const node of Object.values(graph.nodes)) {
        const members = node.kind === 'relationship-group' ? [] : node.sourceNodeIds;
        const first = source.nodes[members[0]];
        let count: number | undefined;
        if (node.kind === 'relationship-group') {
            count = node.sourceEdgeIds.length;
        } else if (node.kind === 'node-group') {
            count = members.length;
        }

        const glyphs = new Map<string, SemanticGlyph>();
        for (const member of members) {
            for (const glyph of memberGlyphs[member] ?? []) {
                glyphs.set(glyph.id, glyph);
            }
        }

        let caption: string;
        if (node.kind === 'relationship-group') {
            caption = `${node.sourceEdgeIds.length.toLocaleString('en-US')} relationships`;
        } else if (count === undefined) {
            caption = first?.node.label ?? '';
        } else {
            caption = `${count.toLocaleString('en-US')} ${pluralNodeType(first?.kind ?? '')}`;
        }

        result[node.id] = {
            caption,
            nodeKind: first?.kind,
            count,
            badge: count === undefined ? undefined : formatCompactCount(count),
            glyphs: [...glyphs.values()],
            highlighted: graph.index.projectedToSource[node.id].some((item) => highlightedItems.has(item.id)),
        };
    }

    for (const edge of Object.values(graph.edges)) {
        let caption = '';
        if (edge.kind !== 'connector') {
            const sourceEdge = source.edges[edge.sourceEdgeIds[0]];
            caption = sourceEdge.edge.label ?? sourceEdge.kind;
        }
        result[edge.id] = {
            caption,
            glyphs: [],
            highlighted: edge.sourceEdgeIds.some((id) => highlightedItems.has(id)),
        };
    }
    return result;
}
