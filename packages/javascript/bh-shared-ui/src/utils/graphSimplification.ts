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

import startCase from 'lodash/startCase';

export interface RepresentativeNodeSet {
    members: string[];
    target: string;
}

const compact = new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 1 });
export const formatCompactCount = (count: number): string => compact.format(count).toLowerCase();

export function pluralNodeType(kind: string): string {
    const noun = startCase(kind.replace(/^AZ(?=[A-Z])/, '')).toLowerCase();
    const irregular: Record<string, string> = { person: 'people', child: 'children' };
    if (irregular[noun]) return irregular[noun];
    if (/[^aeiou]y$/.test(noun)) return noun.slice(0, -1) + 'ies';
    return noun + (/(s|x|z|ch|sh)$/.test(noun) ? 'es' : 's');
}

/** Equivalent leaves must have exactly one incident connection in the loaded result. */
export function findRepresentativeNodeSets(
    nodeTypes: Record<string, string>,
    edges: ReadonlyArray<{ source: string; target: string; kind?: string }>
): Record<string, RepresentativeNodeSet> {
    const incident = new Map<string, { count: number; edge: (typeof edges)[number] }>();
    for (const edge of edges) {
        for (const id of new Set([edge.source, edge.target])) {
            incident.set(id, { count: (incident.get(id)?.count ?? 0) + 1, edge });
        }
    }
    const buckets = new Map<string, RepresentativeNodeSet>();
    for (const [id, type] of Object.entries(nodeTypes)) {
        if (!type || ['HIDDEN', 'Meta', 'MetaDetail'].includes(type)) continue;
        const connection = incident.get(id);
        if (connection?.count !== 1) continue;
        const { edge } = connection;
        const target = edge.source === id ? edge.target : edge.source;
        if (target === id || !Object.hasOwn(nodeTypes, target) || !edge.kind) continue;
        const key = JSON.stringify([type, edge.kind, edge.source === id ? 'out' : 'in', target]);
        const bucket = buckets.get(key) ?? { members: [], target };
        bucket.members.push(id);
        buckets.set(key, bucket);
    }
    return Object.fromEntries(
        [...buckets].flatMap(([key, set]) => {
            const id = `representative:${key}`;
            return set.members.length < 2 || Object.hasOwn(nodeTypes, id)
                ? []
                : [[id, { ...set, members: set.members.sort() }]];
        })
    );
}
