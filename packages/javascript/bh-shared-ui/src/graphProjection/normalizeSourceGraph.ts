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

import type { FlatGraphResponse, GraphEdge, GraphResponse, StyledGraphEdge } from 'js-client-library';
import type { SourceEdge, SourceGraph, SourceNode } from './types';

export const endpointKey = (source: string, target: string): string => JSON.stringify([source, target]);

/** Normalize transport data without discarding the payload used by legacy consumers. */
export function normalizeSourceGraph(response: GraphResponse | FlatGraphResponse): SourceGraph {
    const nodes: Record<string, SourceNode> = Object.create(null);
    const edges: Record<string, SourceEdge> = Object.create(null);
    const incidentEdges: Record<string, string[]> = Object.create(null);
    const endpointEdges: Record<string, string[]> = Object.create(null);
    const addEdge = (edgeId: string, edge: GraphEdge, original: SourceEdge['original']) => {
        if (Object.hasOwn(edges, edgeId)) {
            throw new Error(`Duplicate source graph edge ID: ${edgeId}`);
        }
        edges[edgeId] = { id: edgeId, source: edge.source, target: edge.target, kind: edge.kind, edge, original };
        // A self-loop is incident to its node once, though it has two endpoints.
        for (const endpoint of new Set([edge.source, edge.target])) {
            (incidentEdges[endpoint] ??= []).push(edgeId);
        }
        (endpointEdges[endpointKey(edge.source, edge.target)] ??= []).push(edgeId);
    };
    const structured = response as GraphResponse;
    if (structured.data?.nodes && Array.isArray(structured.data.edges)) {
        for (const [id, node] of Object.entries(structured.data.nodes)) {
            nodes[id] = { id, kind: node.kind, node, original: node };
        }
        for (const edge of structured.data.edges) {
            addEdge(edge.exploreGraphId ?? `rel_${edge.id}`, edge, edge);
        }
    } else {
        for (const [id, original] of Object.entries(response as FlatGraphResponse)) {
            if ('id1' in original && 'id2' in original) {
                const item = original as StyledGraphEdge;
                addEdge(
                    id,
                    {
                        id: item.id,
                        source: item.id1,
                        target: item.id2,
                        kind: item.data?.kind ?? item.label?.text ?? '',
                        label: item.label?.text ?? '',
                        lastSeen: item.data?.lastSeen ?? item.data?.lastseen ?? '',
                        data: item.data,
                        exploreGraphId: id,
                    },
                    original
                );
            } else {
                const data = original.data ?? {};
                const tags = String(data.system_tags ?? '');
                const node = {
                    label: original.label?.text ?? '',
                    kind: data.nodetype ?? '',
                    kinds: data.kinds ?? [],
                    objectId: data.objectid ?? data.objectId ?? '',
                    lastSeen: data.lastSeen ?? data.lastseen ?? '',
                    isTierZero: data.isTierZero ?? tags.includes('admin_tier_0'),
                    isOwnedObject: data.isOwnedObject ?? tags.includes('owned'),
                    properties: { ...data },
                };
                nodes[id] = { id, kind: node.kind, node, original };
            }
        }
    }
    return { nodes, edges, incidentEdges, endpointEdges };
}
