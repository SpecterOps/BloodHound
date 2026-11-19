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

import type { GraphNode, GraphResponse } from 'js-client-library';
import { discoverLeafGroups } from './discoverLeafGroups';
import { normalizeSourceGraph } from './normalizeSourceGraph';
import { validateGroups } from './validateGroups';

const makeNode = (kind: string): GraphNode => ({
    kind,
    kinds: [kind],
    label: kind,
    objectId: kind,
    lastSeen: '',
    isTierZero: false,
    isOwnedObject: false,
});

const sourceFixture = () =>
    normalizeSourceGraph({
        data: {
            nodes: Object.fromEntries(
                ['a', 'b', 'target'].map((id) => [id, { kind: id === 'target' ? 'Group' : 'User' }])
            ),
            edges: ['a', 'b'].map((source, id) => ({ id, source, target: 'target', kind: 'MemberOf' })),
        },
    } as GraphResponse);

it('finds equivalent leaves using original endpoints and deterministic membership', () => {
    const source = sourceFixture();
    expect(discoverLeafGroups(source)).toMatchObject([
        { kind: 'nodes', ruleId: 'equivalent-leaves', memberNodeIds: ['a', 'b'] },
    ]);
    expect(
        discoverLeafGroups({ ...source, nodes: Object.fromEntries(Object.entries(source.nodes).reverse()) })
    ).toEqual(discoverLeafGroups(source));
    expect(
        discoverLeafGroups({
            ...source,
            edges: { ...source.edges, extra: { ...source.edges.rel_0, id: 'extra' } },
            incidentEdges: { ...source.incidentEdges, a: ['rel_0', 'extra'] },
        })
    ).toEqual([]);
});

it('rejects overlapping membership in either registration order', () => {
    const source = sourceFixture();
    const group = discoverLeafGroups(source)[0];
    const other = { ...group, id: 'other' };
    expect(validateGroups(source, [group, other]).valid).toBe(false);
    expect(validateGroups(source, [other, group]).valid).toBe(false);
    expect(validateGroups(source, [group]).valid).toBe(true);
});

it('rejects nonexistent members and inconsistent relationship endpoints', () => {
    const source = sourceFixture();
    expect(
        validateGroups(source, [{ kind: 'nodes', id: 'g', ruleId: 'test', memberNodeIds: ['a', 'missing'] }]).valid
    ).toBe(false);
    expect(
        validateGroups(source, [
            {
                kind: 'relationships',
                id: 'g',
                ruleId: 'test',
                sourceNodeId: 'a',
                targetNodeId: 'target',
                memberEdgeIds: ['rel_0', 'rel_1'],
            },
        ]).valid
    ).toBe(false);
});

it('does not let a source ID resembling the old synthetic prefix suppress a group', () => {
    const source = sourceFixture();
    const collision = 'representative:["User","MemberOf","out","target"]';
    const withCollision = {
        ...source,
        nodes: { ...source.nodes, [collision]: { ...source.nodes.target, id: collision } },
    };
    expect(discoverLeafGroups(withCollision)).toHaveLength(1);
});

it('requires matching type, relationship, direction and complete leaf neighborhood', () => {
    const nodeTypes = { a: 'User', b: 'User', target: 'Group' };
    const edge = (source: string, target = 'target', kind = 'MemberOf') => ({ source, target, kind });
    const graph = (types: Record<string, string>, edges: ReturnType<typeof edge>[]) =>
        normalizeSourceGraph({
            data: {
                nodes: Object.fromEntries(Object.entries(types).map(([id, kind]) => [id, makeNode(kind)])),
                edges: edges.map((value, id) => ({ ...value, id, label: value.kind, lastSeen: '' })),
            },
        } as GraphResponse);
    for (const edges of [
        [edge('a'), edge('target', 'b')],
        [edge('a'), edge('b', 'target', 'Owns')],
        [edge('a'), edge('b'), edge('a', 'target', 'Owns')],
        [edge('a'), edge('b'), edge('a')],
        [edge('a', 'a'), edge('b', 'b')],
        [edge('a', 'missing'), edge('b', 'missing')],
        [edge('a', 'target', ''), edge('b', 'target', '')],
    ])
        expect(discoverLeafGroups(graph(nodeTypes, edges))).toEqual([]);
    expect(discoverLeafGroups(graph({ ...nodeTypes, b: 'Computer' }, [edge('a'), edge('b')]))).toEqual([]);
    for (const kind of ['HIDDEN', 'Meta', 'MetaDetail', ''])
        expect(discoverLeafGroups(graph({ ...nodeTypes, a: kind, b: kind }, [edge('a'), edge('b')]))).toEqual([]);
});
