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

import type { GroupDefinition, SourceGraph, SourceItemRef } from './types';

export const projectionId = (kind: string, id: string): string => JSON.stringify([kind, id]);

export function effectivelyExpandedGroups(
    source: SourceGraph,
    groups: readonly GroupDefinition[],
    expanded: ReadonlySet<string>,
    revealed: readonly SourceItemRef[]
): ReadonlySet<string> {
    const revealedNodeIds = new Set(revealed.filter((item) => item.kind === 'node').map((item) => item.id));
    const revealedEdgeIds = new Set(revealed.filter((item) => item.kind === 'edge').map((item) => item.id));
    for (const edgeId of revealedEdgeIds) {
        const edge = source.edges[edgeId];
        if (edge) {
            revealedNodeIds.add(edge.source);
            revealedNodeIds.add(edge.target);
        }
    }
    return new Set([
        ...expanded,
        ...groups
            .filter((group) => {
                if (group.kind === 'nodes') {
                    return group.memberNodeIds.some((nodeId) => revealedNodeIds.has(nodeId));
                }
                return group.memberEdgeIds.some((edgeId) => revealedEdgeIds.has(edgeId));
            })
            .map((group) => group.id),
    ]);
}

/** Canonicalize semantic metadata only; renderer styles never determine topology. */
export function semanticKey(value: unknown): string {
    return JSON.stringify(value, (_key, item) => {
        if (item && typeof item === 'object' && !Array.isArray(item)) {
            return Object.fromEntries(
                Object.keys(item)
                    .sort()
                    .map((key) => [key, item[key]])
            );
        }
        return item;
    });
}
