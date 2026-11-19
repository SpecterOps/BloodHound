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

import type { GraphResponse } from 'js-client-library';
import { buildProjection } from './buildProjection';
import { discoverLeafGroups } from './discoverLeafGroups';
import { normalizeSourceGraph } from './normalizeSourceGraph';
import type { GroupDefinition } from './types';

const fixture = () =>
    normalizeSourceGraph({
        data: {
            nodes: Object.fromEntries(
                ['a', 'b', 'c', 'd'].map((id) => [id, { kind: id === 'c' || id === 'd' ? 'Computer' : 'User' }])
            ),
            edges: [
                { id: 1, source: 'a', target: 'b', kind: 'Owns' },
                { id: 2, source: 'a', target: 'b', kind: 'AdminTo' },
                { id: 3, source: 'b', target: 'a', kind: 'Owns' },
                { id: 4, source: 'b', target: 'a', kind: 'GenericAll' },
                { id: 5, source: 'c', target: 'a', kind: 'MemberOf' },
                { id: 6, source: 'd', target: 'a', kind: 'MemberOf' },
            ],
        },
    } as GraphResponse);
const relationshipGroups: GroupDefinition[] = [
    {
        kind: 'relationships',
        id: 'out',
        ruleId: 'relationships',
        sourceNodeId: 'a',
        targetNodeId: 'b',
        memberEdgeIds: ['rel_1', 'rel_2'],
    },
    {
        kind: 'relationships',
        id: 'in',
        ruleId: 'relationships',
        sourceNodeId: 'b',
        targetNodeId: 'a',
        memberEdgeIds: ['rel_3', 'rel_4'],
    },
];

it('composes node groups and both directions with complete provenance', () => {
    const source = fixture();
    const groups = [...discoverLeafGroups(source), ...relationshipGroups];
    const graph = buildProjection(source, groups, { expandedGroupIds: new Set() }, []);
    expect(Object.values(graph.nodes).filter((node) => node.kind === 'node-group')).toHaveLength(1);
    expect(Object.values(graph.nodes).filter((node) => node.kind === 'relationship-group')).toHaveLength(2);
    expect(graph.index.sourceEdgeToVisible.rel_1).toHaveLength(3);
    expect(graph.index.sourceNodeToVisible.c).toBe(graph.index.sourceNodeToVisible.d);
    for (const edge of Object.values(graph.edges)) {
        expect(graph.nodes[edge.source]).toBeDefined();
        expect(graph.nodes[edge.target]).toBeDefined();
    }
    const expanded = buildProjection(source, groups, { expandedGroupIds: new Set(['out']) }, []);
    expect(Object.values(expanded.nodes).filter((node) => node.kind === 'relationship-group')).toMatchObject([
        { groupId: 'in' },
    ]);
    expect(Object.values(expanded.nodes).filter((node) => node.kind === 'node-group')).toHaveLength(1);
    const all = buildProjection(source, groups, { expandedGroupIds: new Set(groups.map((group) => group.id)) }, []);
    expect(Object.keys(all.nodes)).toHaveLength(4);
    expect(Object.keys(all.edges)).toHaveLength(6);
});

it('reveals selected edge endpoints and relationships without changing explicit expansion', () => {
    const source = fixture();
    const graph = buildProjection(source, relationshipGroups, { expandedGroupIds: new Set() }, [
        { kind: 'edge', id: 'rel_1' },
    ]);
    expect(graph.index.sourceEdgeToVisible.rel_1).toHaveLength(1);
    expect(graph.index.sourceEdgeToVisible.rel_3).toHaveLength(3);
});

it('merges equivalent boundary edges but preserves differing semantic metadata', () => {
    const source = fixture();
    const groups = discoverLeafGroups(source);
    const collapsed = buildProjection(source, groups, { expandedGroupIds: new Set() }, []);
    expect(collapsed.index.sourceEdgeToVisible.rel_5).toEqual(collapsed.index.sourceEdgeToVisible.rel_6);
    const changed = {
        ...source,
        edges: {
            ...source.edges,
            rel_6: { ...source.edges.rel_6, edge: { ...source.edges.rel_6.edge, impactPercent: 20 } },
        },
    };
    const different = buildProjection(changed, groups, { expandedGroupIds: new Set() }, []);
    expect(different.index.sourceEdgeToVisible.rel_5).not.toEqual(different.index.sourceEdgeToVisible.rel_6);
});
