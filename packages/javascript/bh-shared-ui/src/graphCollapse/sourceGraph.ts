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

import type {
    FlatGraphResponse,
    GraphEdge,
    GraphNode,
    GraphResponse,
    StyledGraphEdge,
    StyledGraphNode,
} from 'js-client-library';
import { isGraphResponse } from '../hooks/useExploreGraph/queries/utils';
import { REL_ID_PREFIX } from '../utils/parseItemId';

export type CollapseGraphNode = Readonly<{
    kind: string;
    kinds: readonly string[];
    systemTags: readonly string[];
    isTierZero: boolean;
    isOwnedObject: boolean;
    original: Readonly<GraphNode | StyledGraphNode>;
}>;

export type CollapseGraphEdge = Readonly<{
    source: string;
    target: string;
    kind: string;
    relationshipId: string | undefined;
    /** The flat response key, when the source used the flat format. */
    sourceKey?: string;
    original: Readonly<GraphEdge | StyledGraphEdge>;
}>;

export type CollapseGraph = Readonly<{
    nodes: ReadonlyMap<string, CollapseGraphNode>;
    edges: ReadonlyMap<string, CollapseGraphEdge>;
    incidentEdges: ReadonlyMap<string, readonly string[]>;
}>;

/** The source facts available to every policy rule. */
export type CollapseTopology = Readonly<{
    nodes: ReadonlyMap<string, Omit<CollapseGraphNode, 'original'>>;
    edges: ReadonlyMap<string, Omit<CollapseGraphEdge, 'original'>>;
    incidentEdges: CollapseGraph['incidentEdges'];
}>;

function readSystemTags(raw: unknown): string[] {
    if (typeof raw === 'string') return raw.split(/[\s,]+/).filter(Boolean);
    return Array.isArray(raw) ? raw.filter((tag): tag is string => typeof tag === 'string') : [];
}

function readRelationshipId(raw: unknown): string | undefined {
    if (typeof raw === 'string' && raw.length > 0) return raw;
    if (typeof raw === 'number' && Number.isSafeInteger(raw)) return String(raw);
    return undefined;
}

/** Derive graph facts for the rules while retaining original records for renderer adapters. */
export function createCollapseGraph(response: GraphResponse | FlatGraphResponse): CollapseGraph {
    const nodes = new Map<string, CollapseGraphNode>();
    const edges = new Map<string, CollapseGraphEdge>();

    if (isGraphResponse(response)) {
        for (const [id, original] of Object.entries(response.data.nodes)) {
            nodes.set(id, {
                kind: original.kind,
                kinds: original.kinds?.slice() ?? [],
                systemTags: readSystemTags(original.properties?.system_tags),
                isTierZero: original.isTierZero,
                isOwnedObject: original.isOwnedObject,
                original,
            });
        }

        const anonymousOccurrences = new Map<string, number>();
        for (const original of response.data.edges) {
            const relationshipId = readRelationshipId(original.id);
            let id: string;
            if (relationshipId === undefined) {
                // Redacted relationships can share endpoints and carry no backend identity.
                const signature = JSON.stringify([original.source, original.target, original.kind, original.label]);
                const occurrence = anonymousOccurrences.get(signature) ?? 0;
                anonymousOccurrences.set(signature, occurrence + 1);
                id = `anonymous-edge:${signature}:${occurrence}`;
            } else {
                id = REL_ID_PREFIX + relationshipId;
            }
            if (edges.has(id)) continue;

            edges.set(id, {
                source: original.source,
                target: original.target,
                kind: original.kind,
                relationshipId,
                original,
            });
        }
    } else {
        for (const [sourceKey, original] of Object.entries(response)) {
            if ('id1' in original) {
                // The canonical flat key carries the exact ID even when the embedded number was rounded.
                const keyId = sourceKey.startsWith(REL_ID_PREFIX) ? sourceKey.slice(REL_ID_PREFIX.length) : '';
                const relationshipId = /^\d+$/.test(keyId) ? keyId : readRelationshipId(original.id);
                const id =
                    relationshipId === undefined ? `anonymous-edge:${sourceKey}` : REL_ID_PREFIX + relationshipId;
                if (edges.has(id)) continue;

                edges.set(id, {
                    source: original.id1,
                    target: original.id2,
                    kind: original.label?.text ?? '',
                    relationshipId,
                    sourceKey,
                    original,
                });
            } else {
                const systemTags = readSystemTags(original.data?.system_tags);
                nodes.set(sourceKey, {
                    kind: original.data?.nodetype ?? '',
                    kinds: original.data?.kinds?.slice() ?? [],
                    systemTags,
                    isTierZero: original.data?.isTierZero ?? systemTags.includes('admin_tier_0'),
                    isOwnedObject: original.data?.isOwnedObject ?? systemTags.includes('owned'),
                    original,
                });
            }
        }
    }

    const incidentEdges = new Map<string, string[]>();
    for (const [id, edge] of edges) {
        const endpoints = edge.source === edge.target ? [edge.source] : [edge.source, edge.target];
        for (const endpoint of endpoints) {
            const incident = incidentEdges.get(endpoint);
            if (incident) {
                incident.push(id);
            } else {
                incidentEdges.set(endpoint, [id]);
            }
        }
    }

    return { nodes, edges, incidentEdges };
}
