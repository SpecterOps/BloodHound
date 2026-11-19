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

import { act, renderHook } from '@testing-library/react';
import type { GraphNode, GraphResponse } from 'js-client-library';
import { discoverLeafGroups } from './discoverLeafGroups';
import { normalizeSourceGraph } from './normalizeSourceGraph';
import type { SourceItemRef } from './types';
import { useGraphProjection } from './useGraphProjection';

const makeNode = (kind: string): GraphNode => ({
    kind,
    kinds: [kind],
    label: kind,
    objectId: kind,
    lastSeen: '',
    isTierZero: false,
    isOwnedObject: false,
});

const response = {
    data: {
        nodes: { a: makeNode('User'), b: makeNode('User'), t: makeNode('Group') },
        edges: [
            { id: 1, source: 'a', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
            { id: 2, source: 'b', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
        ],
    },
} as GraphResponse;
const rules = [{ id: 'leaves', discover: discoverLeafGroups }];

it('retains explicit choices across selection and refetch but resets for a new query', () => {
    const { result, rerender } = renderHook(
        ({
            key,
            source,
            selected,
        }: {
            key: string;
            source: ReturnType<typeof normalizeSourceGraph>;
            selected?: SourceItemRef;
        }) => useGraphProjection({ source, queryKey: key, rules, selectedSourceItem: selected }),
        {
            initialProps: {
                key: 'one',
                source: normalizeSourceGraph(response),
                selected: undefined as SourceItemRef | undefined,
            },
        }
    );
    expect(result.current.canExpand).toBe(true);
    act(() => result.current.expandAll());
    expect(result.current.canExpand).toBe(false);
    rerender({ key: 'one', source: normalizeSourceGraph(response), selected: undefined });
    expect(result.current.canExpand).toBe(false);
    rerender({ key: 'two', source: normalizeSourceGraph(response), selected: undefined });
    expect(result.current.canExpand).toBe(true);
    rerender({ key: 'two', source: normalizeSourceGraph(response), selected: { kind: 'node', id: 'a' } });
    expect(result.current.canExpand).toBe(false);
    rerender({ key: 'two', source: normalizeSourceGraph(response), selected: undefined });
    expect(result.current.canExpand).toBe(true);
});

it('falls back to original topology for overlaps without registration order deciding a winner', () => {
    const diagnostic = vi.spyOn(console, 'error').mockImplementation(() => {});
    const conflicting = [
        ...rules,
        {
            id: 'other',
            discover: (source: ReturnType<typeof normalizeSourceGraph>) =>
                discoverLeafGroups(source).map((group) => ({ ...group, id: 'other' })),
        },
    ];
    const { result } = renderHook(() =>
        useGraphProjection({ source: normalizeSourceGraph(response), queryKey: 'one', rules: conflicting })
    );
    expect(Object.values(result.current.graph!.nodes).every((node) => node.kind === 'source-node')).toBe(true);
    expect(result.current.canExpand).toBe(false);
    expect(diagnostic).toHaveBeenCalledTimes(1);
    diagnostic.mockRestore();
});
