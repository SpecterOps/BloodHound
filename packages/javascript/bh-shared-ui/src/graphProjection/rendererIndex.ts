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

import type { ProjectedGraph, SourceGraph, SourceItemRef } from './types';

/** Map projection IDs to renderer IDs without reserving a user-controlled prefix. */
export function createRendererIndex(source: SourceGraph, graph: ProjectedGraph, separateRedirectedEdges = false) {
    const sourceIds = new Set([...Object.keys(source.nodes), ...Object.keys(source.edges)]);
    const rendererIds = new Set<string>();
    const projectedToRenderer: Record<string, string> = Object.create(null);
    const rendererToProjected: Record<string, string> = Object.create(null);
    const sourceItems: Record<string, SourceItemRef> = Object.create(null);
    const groupIds: Record<string, string> = Object.create(null);

    const allocateRendererId = (projectedId: string, preferredId: string, isSynthetic: boolean) => {
        let rendererId = preferredId;
        while (rendererIds.has(rendererId) || (isSynthetic && sourceIds.has(rendererId))) {
            rendererId = `projection:${rendererId}`;
        }
        rendererIds.add(rendererId);
        projectedToRenderer[projectedId] = rendererId;
        rendererToProjected[rendererId] = projectedId;
        return rendererId;
    };

    for (const node of Object.values(graph.nodes)) {
        const rendererId = allocateRendererId(
            node.id,
            node.kind === 'source-node'
                ? node.sourceNodeIds[0]
                : `${node.kind === 'node-group' ? 'representative' : 'relationship-aggregate'}:${node.groupId}`,
            node.kind !== 'source-node'
        );

        if (node.kind === 'source-node') {
            sourceItems[rendererId] = { kind: 'node', id: node.sourceNodeIds[0] };
        } else {
            groupIds[rendererId] = node.groupId;
        }
    }

    for (const edge of Object.values(graph.edges)) {
        const touchesProjectedGroup =
            separateRedirectedEdges &&
            (graph.nodes[edge.source].kind !== 'source-node' || graph.nodes[edge.target].kind !== 'source-node');
        const isSynthetic = edge.kind !== 'source-edge' || touchesProjectedGroup;
        // Use projection IDs so redirected edges don't reuse cached entries for their original source IDs.
        const preferredId = isSynthetic ? `projection-edge:${edge.id}` : edge.sourceEdgeIds[0];
        const rendererId = allocateRendererId(edge.id, preferredId, isSynthetic);

        if (edge.groupId) {
            groupIds[rendererId] = edge.groupId;
        } else if (edge.sourceEdgeIds.length === 1) {
            sourceItems[rendererId] = { kind: 'edge', id: edge.sourceEdgeIds[0] };
        }
    }
    return { projectedToRenderer, rendererToProjected, sourceItems, groupIds };
}
