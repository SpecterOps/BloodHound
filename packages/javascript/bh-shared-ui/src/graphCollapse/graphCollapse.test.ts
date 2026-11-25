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

import type { FlatGraphResponse, GraphEdge, GraphNode, GraphResponse, StyledGraphEdge } from 'js-client-library';
import { describe, expect, expectTypeOf, it } from 'vitest';
import {
    collapseSimilarLeaves,
    createCollapseGraph,
    planGraphCollapse,
    projectGraphCollapse,
    type CollapseRule,
} from './index';

const node = (kind: string, extra: Partial<GraphNode> = {}): GraphNode => ({
    kind,
    kinds: [kind],
    label: kind,
    objectId: kind,
    lastSeen: '',
    isTierZero: false,
    isOwnedObject: false,
    ...extra,
});

const edge = (id: number, source: string, target: string, kind = 'MemberOf'): GraphEdge => ({
    id,
    source,
    target,
    kind,
    label: kind,
    lastSeen: '',
});

const response = (nodes: GraphResponse['data']['nodes'], edges: GraphEdge[]): GraphResponse => ({
    data: { nodes, edges },
});

describe('createCollapseGraph', () => {
    it('retains structured records and counts repeated relationship IDs once', () => {
        const a = node('User', { properties: { description: 'original' } });
        const first = edge(1, 'a', 'b');
        const second = edge(2, 'a', 'b');
        const graph = createCollapseGraph(response({ a, b: node('Group') }, [first, { ...first }, second]));
        expect(graph.nodes.get('a')?.original).toBe(a);
        expect(graph.edges.get('rel_1')?.original).toBe(first);
        expect([...graph.edges.keys()]).toEqual(['rel_1', 'rel_2']);
        expect(graph.incidentEdges.get('a')).toEqual(['rel_1', 'rel_2']);
    });

    it('preserves distinct redacted relationships and their keys across response ordering', () => {
        const first = { ...edge(1, 'a', 'b', 'HIDDEN'), id: '' } as unknown as GraphEdge;
        const second = { ...edge(2, 'b', 'a', 'HIDDEN'), id: '' } as unknown as GraphEdge;
        const nodes = { a: node('User'), b: node('HIDDEN') };
        const original = createCollapseGraph(response(nodes, [first, second, { ...first }]));
        const reordered = createCollapseGraph(response(nodes, [second, { ...first }, first]));
        expect(original.edges.size).toBe(3);
        expect(reordered.edges).toEqual(original.edges);
    });

    it('retains flat response keys and exposes comparable kinds and endpoints', () => {
        const flat: FlatGraphResponse = {
            a: {
                color: '',
                border: { color: '' },
                data: { nodetype: 'User', kinds: ['User'], system_tags: 'not_owned admin_tier_0' },
                fontIcon: { text: '' },
                label: { text: 'Alice', backgroundColor: '', center: true, fontSize: 10 },
                size: 1,
            },
            legacy_link: {
                id: 7,
                id1: 'a',
                id2: 'b',
                color: '',
                data: { kind: 'custom-property-value', custom: true },
                label: { text: 'MemberOf' },
            },
        };
        const graph = createCollapseGraph(flat);
        expect(graph.nodes.get('a')).toMatchObject({ kind: 'User', isTierZero: true, isOwnedObject: false });
        expect(graph.nodes.get('a')?.original).toBe(flat.a);
        expect(graph.edges.get('rel_7')).toMatchObject({
            source: 'a',
            target: 'b',
            kind: 'MemberOf',
            sourceKey: 'legacy_link',
        });
        expect(graph.edges.get('rel_7')?.original).toBe(flat.legacy_link);
    });

    it('uses exact flat relationship keys when the embedded number loses precision', () => {
        const first: StyledGraphEdge = {
            id: 9_007_199_254_740_992,
            id1: 'a',
            id2: 'b',
            color: '',
            data: {},
            label: { text: 'MemberOf' },
        };
        const second = { ...first };
        const graph = createCollapseGraph({ rel_9007199254740992: first, rel_9007199254740993: second });
        expect(graph.edges.get('rel_9007199254740992')?.original).toBe(first);
        expect(graph.edges.get('rel_9007199254740993')?.original).toBe(second);
    });
});

describe('Explore collapse', () => {
    const leaves = response(
        { a: node('User'), b: node('User'), c: node('User'), d: node('User'), group: node('Group') },
        [edge(1, 'a', 'group'), edge(2, 'b', 'group'), edge(3, 'group', 'c'), edge(4, 'group', 'd')]
    );

    it('exposes topology facts to rules while keeping the raw record out of the rule contract', () => {
        type RuleGraph = Parameters<CollapseRule['propose']>[0];
        type RuleNode = NonNullable<ReturnType<RuleGraph['nodes']['get']>>;
        type RuleEdge = NonNullable<ReturnType<RuleGraph['edges']['get']>>;
        expectTypeOf<RuleNode>().not.toHaveProperty('original');
        expectTypeOf<RuleEdge>().not.toHaveProperty('original');
    });

    it('finds separate leaf groups by direction and looks members up by their original IDs', () => {
        const plan = planGraphCollapse(createCollapseGraph(leaves), [collapseSimilarLeaves]);
        const incoming = plan.groupByNodeId.get('a');
        const outgoing = plan.groupByNodeId.get('c');
        expect(incoming).toBeDefined();
        expect(outgoing).toBeDefined();
        expect(incoming).not.toBe(outgoing);
        expect(plan.groupByEdgeId.get('rel_2')).toBe(incoming);
        expect(plan.groupByNodeId.get('d')).toBe(outgoing);
        expect([...plan.groups.values()]).toEqual([
            expect.objectContaining({ kind: 'leaves', anchorId: 'group', direction: 'toward-anchor' }),
            expect.objectContaining({ kind: 'leaves', anchorId: 'group', direction: 'away-from-anchor' }),
        ]);
    });

    it('requires matching kinds and tags and counts all original incident edges', () => {
        const graph = createCollapseGraph(
            response(
                {
                    a: node('User'),
                    b: node('User'),
                    owned: node('User', { isOwnedObject: true }),
                    tierZero: node('User', { isTierZero: true }),
                    computer: node('Computer'),
                    differentEdge: node('User'),
                    multipleEdges: node('User'),
                    loop: node('User'),
                    hidden: node('HIDDEN'),
                    anotherHidden: node('HIDDEN'),
                    group: node('Group'),
                },
                [
                    edge(1, 'a', 'group'),
                    edge(2, 'b', 'group'),
                    edge(3, 'owned', 'group'),
                    edge(4, 'tierZero', 'group'),
                    edge(5, 'computer', 'group'),
                    edge(6, 'differentEdge', 'group', 'AdminTo'),
                    edge(7, 'multipleEdges', 'group'),
                    edge(8, 'multipleEdges', 'group', 'AdminTo'),
                    edge(9, 'loop', 'loop'),
                    edge(10, 'loop', 'group'),
                    edge(11, 'hidden', 'group'),
                    edge(12, 'anotherHidden', 'group'),
                ]
            )
        );
        const plan = planGraphCollapse(graph, [collapseSimilarLeaves]);
        expect([...plan.groupByNodeId.keys()].sort()).toEqual(['a', 'b']);
    });

    it('restores individual groups and all original records without rebuilding the plan', () => {
        const graph = createCollapseGraph(leaves);
        const plan = planGraphCollapse(graph, [collapseSimilarLeaves]);
        const collapsed = projectGraphCollapse(plan);
        expect([...collapsed.nodes.keys()]).toEqual(['group']);
        expect(collapsed.edges.size).toBe(0);
        expect(collapsed.collapsedGroups.size).toBe(2);

        const incoming = plan.groupByEdgeId.get('rel_1')!;
        const partial = projectGraphCollapse(plan, { expandedGroupIds: new Set([incoming]) });
        expect([...partial.nodes.keys()]).toEqual(['a', 'b', 'group']);
        expect([...partial.edges.keys()]).toEqual(['rel_1', 'rel_2']);
        expect(partial.collapsedGroups.has(incoming)).toBe(false);
        expect(partial.collapsedGroups.size).toBe(1);

        const expanded = projectGraphCollapse(plan, { mode: 'expanded' });
        expect(expanded.nodes).toBe(graph.nodes);
        expect(expanded.edges).toBe(graph.edges);
        expect(expanded.collapsedGroups.size).toBe(0);
    });

    it('keeps stable group IDs when the server returns items in a different order', () => {
        const reversed = response(
            Object.fromEntries(Object.entries(leaves.data.nodes).reverse()),
            [...leaves.data.edges].reverse()
        );
        const first = planGraphCollapse(createCollapseGraph(leaves), [collapseSimilarLeaves]);
        const second = planGraphCollapse(createCollapseGraph(reversed), [collapseSimilarLeaves]);
        expect(second.groupByNodeId).toEqual(first.groupByNodeId);
        expect(second.groupByEdgeId).toEqual(first.groupByEdgeId);
    });

    it('rejects an overlapping proposal as a whole so an independent later proposal can win', () => {
        const graph = createCollapseGraph(
            response({ a: node('User'), b: node('User'), c: node('User'), d: node('User'), group: node('Group') }, [
                edge(1, 'a', 'group'),
                edge(2, 'b', 'group'),
                edge(3, 'c', 'group'),
                edge(4, 'd', 'group'),
            ])
        );
        const proposal = (key: string, first: [string, string], second: [string, string]): CollapseRule => ({
            id: key,
            propose: () => [
                {
                    kind: 'leaves',
                    key: [],
                    anchorId: 'group',
                    direction: 'toward-anchor',
                    members: [first, second].map(([nodeId, edgeId]) => ({ nodeId, edgeId })),
                },
            ],
        });
        const plan = planGraphCollapse(graph, [
            proposal('first', ['a', 'rel_1'], ['b', 'rel_2']),
            proposal('overlap', ['b', 'rel_2'], ['c', 'rel_3']),
            proposal('last', ['c', 'rel_3'], ['d', 'rel_4']),
        ]);
        expect([...plan.groups.values()].map((group) => group.ruleId)).toEqual(['first', 'last']);
        expect(plan.groupByEdgeId.get('rel_3')).toBe(plan.groupByNodeId.get('d'));
    });
});
