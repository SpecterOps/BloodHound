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

import type { FlatGraphResponse, GraphResponse } from 'js-client-library';
import { normalizeSourceGraph } from './normalizeSourceGraph';

const node = {
    label: 'Alice',
    kind: 'User',
    kinds: ['User', 'owned'],
    objectId: 'oid',
    lastSeen: 'today',
    isTierZero: true,
    isOwnedObject: true,
    properties: { custom: 'retained' },
};

it('preserves structured metadata and graph identities without changing the response', () => {
    const response = {
        data: {
            nodes: { 'representative:test': node, target: node },
            edges: [
                {
                    id: 7,
                    exploreGraphId: 'original-key',
                    source: 'representative:test',
                    target: 'target',
                    kind: 'Owns',
                    label: 'Owns',
                    lastSeen: 'today',
                    impactPercent: 25,
                    data: { custom: 9 },
                },
                {
                    id: 7,
                    exploreGraphId: 'other-key',
                    source: 'target',
                    target: 'representative:test',
                    kind: 'MemberOf',
                    label: 'MemberOf',
                    lastSeen: 'today',
                },
            ],
        },
    } as GraphResponse;
    const before = structuredClone(response);
    Object.freeze(response.data.nodes);
    const source = normalizeSourceGraph(response);
    expect(source.nodes['representative:test'].node).toEqual(node);
    expect(Object.keys(source.edges)).toEqual(['original-key', 'other-key']);
    expect(source.edges['original-key'].edge.impactPercent).toBe(25);
    expect(source.incidentEdges['target']).toEqual(['original-key', 'other-key']);
    expect(response).toEqual(before);
});

it('preserves the full legacy payload and maps semantic fields', () => {
    const response = {
        a: {
            label: { text: 'Alice' },
            data: {
                nodetype: 'User',
                kinds: ['User'],
                objectid: 'oid',
                system_tags: 'owned admin_tier_0',
                lastseen: 'today',
                custom: 42,
            },
            color: 'red',
        },
        b: { label: { text: 'B' }, data: { nodetype: 'Group' } },
        legacy_edge: { id: 4, id1: 'a', id2: 'b', label: { text: 'Owns' }, data: { kind: 'Owns', custom: true } },
    } as unknown as FlatGraphResponse;
    const source = normalizeSourceGraph(response);
    expect(source.nodes.a.original).toEqual(response.a);
    expect(source.nodes.a.node).toMatchObject({
        kind: 'User',
        isOwnedObject: true,
        isTierZero: true,
        lastSeen: 'today',
        properties: { custom: 42 },
    });
    expect(source.edges.legacy_edge).toMatchObject({ source: 'a', target: 'b', kind: 'Owns', edge: { id: 4 } });
});

it('rejects duplicate source graph edge IDs instead of losing relationships', () => {
    const edge = { id: 7, source: 'a', target: 'b', kind: 'Owns', label: '', lastSeen: '' };
    expect(() => normalizeSourceGraph({ data: { nodes: {}, edges: [edge, edge] } } as GraphResponse)).toThrow(
        /duplicate/i
    );
});
