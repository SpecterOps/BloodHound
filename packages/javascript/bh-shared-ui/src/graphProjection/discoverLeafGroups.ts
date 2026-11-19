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

import type { GroupDefinition, SourceGraph } from './types';

export function discoverLeafGroups(source: SourceGraph): readonly GroupDefinition[] {
    const buckets = new Map<string, string[]>();
    for (const node of Object.values(source.nodes)) {
        if (!node.kind || ['HIDDEN', 'Meta', 'MetaDetail'].includes(node.kind)) {
            continue;
        }
        const incidentEdgeIds = source.incidentEdges[node.id] ?? [];
        if (incidentEdgeIds.length !== 1) {
            continue;
        }
        const edge = source.edges[incidentEdgeIds[0]];
        const neighbor = edge.source === node.id ? edge.target : edge.source;
        if (!edge.kind || neighbor === node.id || !Object.hasOwn(source.nodes, neighbor)) {
            continue;
        }
        const direction = edge.source === node.id ? 'out' : 'in';
        const key = JSON.stringify([node.kind, edge.kind, direction, neighbor]);
        const members = buckets.get(key) ?? [];
        members.push(node.id);
        buckets.set(key, members);
    }
    return [...buckets.values()]
        .filter((members) => members.length >= 2)
        .map((members) => {
            const memberNodeIds = [...members].sort();
            return {
                kind: 'nodes',
                id: JSON.stringify(['equivalent-leaves', memberNodeIds]),
                ruleId: 'equivalent-leaves',
                memberNodeIds,
            };
        });
}
