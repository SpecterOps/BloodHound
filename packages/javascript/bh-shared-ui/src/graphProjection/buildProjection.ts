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

import { effectivelyExpandedGroups, projectionId, semanticKey } from './provenance';
import type {
    GroupDefinition,
    ProjectedEdge,
    ProjectedGraph,
    ProjectedNode,
    ProjectionState,
    SourceGraph,
    SourceItemRef,
} from './types';

/** Materialize validated groups. All mutation is confined to newly allocated output. */
export function buildProjection(
    source: SourceGraph,
    groups: readonly GroupDefinition[],
    state: ProjectionState,
    revealed: readonly SourceItemRef[]
): ProjectedGraph {
    const nodes: Record<string, ProjectedNode> = Object.create(null);
    const edges: Record<string, ProjectedEdge> = Object.create(null);
    const sourceNodeToVisible: Record<string, string> = Object.create(null);
    const sourceEdgeToVisible: Record<string, string[]> = Object.create(null);
    const sourceNodeGroups: Record<string, string[]> = Object.create(null);
    const sourceEdgeGroups: Record<string, string[]> = Object.create(null);
    const projectedToSource: Record<string, SourceItemRef[]> = Object.create(null);
    const expanded = effectivelyExpandedGroups(source, groups, state.expandedGroupIds, revealed);

    // Choose visible node IDs before building relationship groups and their endpoints.
    for (const group of groups) {
        if (group.kind === 'nodes') {
            for (const memberNodeId of group.memberNodeIds) {
                (sourceNodeGroups[memberNodeId] ??= []).push(group.id);
            }
            if (expanded.has(group.id)) {
                continue;
            }
            const id = projectionId('group', group.id);
            nodes[id] = { id, kind: 'node-group', groupId: group.id, sourceNodeIds: group.memberNodeIds };
            for (const memberNodeId of group.memberNodeIds) {
                sourceNodeToVisible[memberNodeId] = id;
            }
        } else {
            for (const memberEdgeId of group.memberEdgeIds) {
                (sourceEdgeGroups[memberEdgeId] ??= []).push(group.id);
            }
        }
    }

    for (const member of Object.keys(source.nodes)) {
        if (sourceNodeToVisible[member]) {
            continue;
        }
        const id = projectionId('node', member);
        nodes[id] = { id, kind: 'source-node', sourceNodeIds: [member] };
        sourceNodeToVisible[member] = id;
    }

    const suppressedEdges = new Set<string>();

    // Collapsed relationship groups replace their member edges with two connectors.
    for (const group of groups) {
        if (group.kind !== 'relationships' || expanded.has(group.id)) {
            continue;
        }
        const visibleSourceId = sourceNodeToVisible[group.sourceNodeId];
        const visibleTargetId = sourceNodeToVisible[group.targetNodeId];
        if (!visibleSourceId || !visibleTargetId || visibleSourceId === visibleTargetId) {
            continue;
        }
        const id = projectionId('group', group.id);
        nodes[id] = { id, kind: 'relationship-group', groupId: group.id, sourceEdgeIds: group.memberEdgeIds };
        const incoming = projectionId('connector-in', group.id);
        const outgoing = projectionId('connector-out', group.id);
        edges[incoming] = {
            id: incoming,
            kind: 'connector',
            source: visibleSourceId,
            target: id,
            groupId: group.id,
            sourceEdgeIds: group.memberEdgeIds,
        };
        edges[outgoing] = {
            id: outgoing,
            kind: 'connector',
            source: id,
            target: visibleTargetId,
            groupId: group.id,
            sourceEdgeIds: group.memberEdgeIds,
        };
        for (const memberEdgeId of group.memberEdgeIds) {
            suppressedEdges.add(memberEdgeId);
            sourceEdgeToVisible[memberEdgeId] = [id, incoming, outgoing];
        }
    }

    const redirected = new Map<string, { id: string; members: string[] }>();
    const endpointIndexes = new Map(Object.keys(nodes).map((id, index) => [id, index]));

    // Redirected edges merge only when semantic metadata and projected endpoints match.
    // Numeric endpoint indexes keep the deduplication key compact when group IDs are large.
    for (const edge of Object.values(source.edges)) {
        if (suppressedEdges.has(edge.id)) {
            continue;
        }
        const visibleSourceId = sourceNodeToVisible[edge.source];
        const visibleTargetId = sourceNodeToVisible[edge.target];
        if (!visibleSourceId || !visibleTargetId) {
            continue;
        }
        const touchesNodeGroup =
            nodes[visibleSourceId].kind === 'node-group' || nodes[visibleTargetId].kind === 'node-group';
        const semanticMetadata = {
            ...edge.edge,
            id: undefined,
            exploreGraphId: undefined,
            source: undefined,
            target: undefined,
        };
        const redirectKey = touchesNodeGroup
            ? semanticKey([
                  endpointIndexes.get(visibleSourceId),
                  endpointIndexes.get(visibleTargetId),
                  edge.kind,
                  semanticMetadata,
              ])
            : '';
        const existing = touchesNodeGroup ? redirected.get(redirectKey) : undefined;
        if (existing) {
            existing.members.push(edge.id);
            sourceEdgeToVisible[edge.id] = [existing.id];
        } else {
            const id = projectionId('edge', edge.id);
            const members = [edge.id];
            edges[id] = {
                id,
                kind: 'source-edge',
                source: visibleSourceId,
                target: visibleTargetId,
                sourceEdgeIds: members,
            };
            sourceEdgeToVisible[edge.id] = [id];
            if (touchesNodeGroup) {
                redirected.set(redirectKey, { id, members });
            }
        }
    }

    for (const node of Object.values(nodes)) {
        projectedToSource[node.id] =
            node.kind === 'relationship-group'
                ? node.sourceEdgeIds.map((id) => ({ kind: 'edge', id }))
                : node.sourceNodeIds.map((id) => ({ kind: 'node', id }));
    }
    for (const edge of Object.values(edges)) {
        projectedToSource[edge.id] = edge.sourceEdgeIds.map((id) => ({ kind: 'edge', id }));
    }
    return {
        nodes,
        edges,
        index: { sourceNodeToVisible, sourceEdgeToVisible, sourceNodeGroups, sourceEdgeGroups, projectedToSource },
    };
}
